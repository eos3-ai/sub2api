<template>
  <AuthLayout>
    <div class="space-y-6">
      <div
        class="rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900 shadow-sm dark:border-amber-400/30 dark:bg-amber-400/10 dark:text-amber-100"
      >
        <p class="font-semibold">📢 AITokenCloud 新系统上线了！</p>
        <p class="mt-1 leading-relaxed">
          AITokenCloud 新系统已上线，现已支持 Gemini CLI登录方式已由用户名登录调整为
          邮箱登录（原账户注册邮箱，密码不变）。
        </p>
        <p class="mt-1 leading-relaxed">
          如忘记注册时使用的邮箱，可点击
          <button
            type="button"
            class="font-semibold text-amber-800 underline underline-offset-2 transition-colors hover:text-amber-900 dark:text-amber-200 dark:hover:text-amber-100"
            @click="showAdminQr = true"
          >
            联系管理员
          </button>
          协助找回。目前 Claude Code 暂未支持，恢复后将第一时间通知大家。老系统地址：
          <a
            href="https://archive.tokencloud.ai"
            class="font-medium text-amber-800 underline underline-offset-2 transition-colors hover:text-amber-900 dark:text-amber-200 dark:hover:text-amber-100"
            target="_blank"
            rel="noreferrer"
          >
            archive.tokencloud.ai
          </a>
        </p>
        <p class="mt-1 leading-relaxed">感谢大家的理解与支持～</p>
      </div>
      <!-- Title -->
      <div class="text-center">
        <h2 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ t('auth.welcomeBack') }}
        </h2>
        <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
          {{ t('auth.signInToAccount') }}
        </p>
      </div>

      <!-- LinuxDo Connect OAuth 登录 -->
      <LinuxDoOAuthSection v-if="linuxdoOAuthEnabled" :disabled="isLoading" />

      <!-- Login Form -->
      <form @submit.prevent="handleLogin" class="space-y-5">
        <!-- Email Input -->
        <div>
          <label for="email" class="input-label">
            {{ t('auth.emailLabel') }}
          </label>
          <div class="relative">
            <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5">
              <Icon name="mail" size="md" class="text-gray-400 dark:text-dark-500" />
            </div>
            <input
              id="email"
              v-model="formData.email"
              type="email"
              required
              autofocus
              autocomplete="email"
              :disabled="isLoading"
              class="input pl-11"
              :class="{ 'input-error': errors.email }"
              :placeholder="t('auth.emailPlaceholder')"
            />
          </div>
          <p v-if="errors.email" class="input-error-text">
            {{ errors.email }}
          </p>
        </div>

        <!-- Password Input -->
        <div>
          <label for="password" class="input-label">
            {{ t('auth.passwordLabel') }}
          </label>
          <div class="relative">
            <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5">
              <Icon name="lock" size="md" class="text-gray-400 dark:text-dark-500" />
            </div>
            <input
              id="password"
              v-model="formData.password"
              :type="showPassword ? 'text' : 'password'"
              required
              autocomplete="current-password"
              :disabled="isLoading"
              class="input pl-11 pr-11"
              :class="{ 'input-error': errors.password }"
              :placeholder="t('auth.passwordPlaceholder')"
            />
            <button
              type="button"
              @click="showPassword = !showPassword"
              class="absolute inset-y-0 right-0 flex items-center pr-3.5 text-gray-400 transition-colors hover:text-gray-600 dark:hover:text-dark-300"
            >
              <Icon v-if="showPassword" name="eyeOff" size="md" />
              <Icon v-else name="eye" size="md" />
            </button>
          </div>
          <p v-if="errors.password" class="input-error-text">
            {{ errors.password }}
          </p>
        </div>

        <!-- Turnstile Widget -->
        <div v-if="turnstileEnabled && turnstileSiteKey">
          <TurnstileWidget
            ref="turnstileRef"
            :site-key="turnstileSiteKey"
            @verify="onTurnstileVerify"
            @expire="onTurnstileExpire"
            @error="onTurnstileError"
          />
          <p v-if="errors.turnstile" class="input-error-text mt-2 text-center">
            {{ errors.turnstile }}
          </p>
        </div>

        <!-- Error Message -->
        <transition name="fade">
          <div
            v-if="errorMessage"
            class="rounded-xl border border-red-200 bg-red-50 p-4 dark:border-red-800/50 dark:bg-red-900/20"
          >
            <div class="flex items-start gap-3">
              <div class="flex-shrink-0">
                <Icon name="exclamationCircle" size="md" class="text-red-500" />
              </div>
              <p class="text-sm text-red-700 dark:text-red-400">
                {{ errorMessage }}
              </p>
            </div>
          </div>
        </transition>

        <!-- Submit Button -->
        <button
          type="submit"
          :disabled="isLoading || (turnstileEnabled && !turnstileToken)"
          class="btn btn-primary w-full"
        >
          <svg
            v-if="isLoading"
            class="-ml-1 mr-2 h-4 w-4 animate-spin text-white"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              class="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            ></circle>
            <path
              class="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            ></path>
          </svg>
          <Icon v-else name="login" size="md" class="mr-2" />
          {{ isLoading ? t('auth.signingIn') : t('auth.signIn') }}
        </button>
      </form>
    </div>

    <Modal
      :show="showAdminQr"
      title="联系管理员"
      size="sm"
      closeOnClickOutside
      @close="showAdminQr = false"
    >
      <div class="flex flex-col items-center gap-3">
        <img
          :src="adminQrUrl"
          alt="管理员二维码"
          class="h-56 w-56 rounded-2xl border border-amber-100 bg-white p-2 shadow-sm dark:border-amber-400/20 dark:bg-dark-900"
        />
        <p class="text-xs text-amber-700 dark:text-amber-200">扫码联系管理员</p>
      </div>
    </Modal>

    <!-- Footer -->
    <template #footer>
      <p class="text-gray-500 dark:text-dark-400">
        {{ t('auth.dontHaveAccount') }}
        <router-link
          to="/register"
          class="font-medium text-primary-600 transition-colors hover:text-primary-500 dark:text-primary-400 dark:hover:text-primary-300"
        >
          {{ t('auth.signUp') }}
        </router-link>
      </p>
    </template>
  </AuthLayout>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { AuthLayout } from '@/components/layout'
import { Modal } from '@/components/common'
import Icon from '@/components/icons/Icon.vue'
import TurnstileWidget from '@/components/TurnstileWidget.vue'
import { useAuthStore, useAppStore } from '@/stores'
import { getPublicSettings } from '@/api/auth'

const { t } = useI18n()

// ==================== Router & Stores ====================

const router = useRouter()
const authStore = useAuthStore()
const appStore = useAppStore()

// ==================== State ====================

const isLoading = ref<boolean>(false)
const errorMessage = ref<string>('')
const showPassword = ref<boolean>(false)
const showAdminQr = ref<boolean>(false)
const adminQrUrl = '/admin-qr.png'

// Public settings
const turnstileEnabled = ref<boolean>(false)
const turnstileSiteKey = ref<string>('')
const linuxdoOAuthEnabled = ref<boolean>(false)

// Turnstile
const turnstileRef = ref<InstanceType<typeof TurnstileWidget> | null>(null)
const turnstileToken = ref<string>('')

const formData = reactive({
  email: '',
  password: ''
})

const errors = reactive({
  email: '',
  password: '',
  turnstile: ''
})

// ==================== Lifecycle ====================

onMounted(async () => {
  const expiredFlag = sessionStorage.getItem('auth_expired')
  if (expiredFlag) {
    sessionStorage.removeItem('auth_expired')
    const message = t('auth.reloginRequired')
    errorMessage.value = message
    appStore.showWarning(message)
  }

  try {
    const settings = await getPublicSettings()
    turnstileEnabled.value = settings.turnstile_enabled
    turnstileSiteKey.value = settings.turnstile_site_key || ''
    linuxdoOAuthEnabled.value = settings.linuxdo_oauth_enabled
  } catch (error) {
    console.error('Failed to load public settings:', error)
  }
})

// ==================== Turnstile Handlers ====================

function onTurnstileVerify(token: string): void {
  turnstileToken.value = token
  errors.turnstile = ''
}

function onTurnstileExpire(): void {
  turnstileToken.value = ''
  errors.turnstile = t('auth.turnstileExpired')
}

function onTurnstileError(): void {
  turnstileToken.value = ''
  errors.turnstile = t('auth.turnstileFailed')
}

// ==================== Validation ====================

function validateForm(): boolean {
  // Reset errors
  errors.email = ''
  errors.password = ''
  errors.turnstile = ''

  let isValid = true

  // Email validation
  if (!formData.email.trim()) {
    errors.email = t('auth.emailRequired')
    isValid = false
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
    errors.email = t('auth.invalidEmail')
    isValid = false
  }

  // Password validation
  if (!formData.password) {
    errors.password = t('auth.passwordRequired')
    isValid = false
  } else if (formData.password.length < 6) {
    errors.password = t('auth.passwordMinLength')
    isValid = false
  }

  // Turnstile validation
  if (turnstileEnabled.value && !turnstileToken.value) {
    errors.turnstile = t('auth.completeVerification')
    isValid = false
  }

  return isValid
}

// ==================== Form Handlers ====================

async function handleLogin(): Promise<void> {
  // Clear previous error
  errorMessage.value = ''

  // Validate form
  if (!validateForm()) {
    return
  }

  isLoading.value = true

  try {
    // Call auth store login
    await authStore.login({
      email: formData.email,
      password: formData.password,
      turnstile_token: turnstileEnabled.value ? turnstileToken.value : undefined
    })

    // Show success toast
    appStore.showSuccess(t('auth.loginSuccess'))

    // Redirect based on user role and query parameter
    const queryRedirect = router.currentRoute.value.query.redirect as string
    let redirectTo: string

    if (queryRedirect) {
      // If there's a redirect query, use it
      redirectTo = queryRedirect
    } else {
      // Otherwise, redirect based on user role
      redirectTo = authStore.isAdmin ? '/admin/dashboard' : '/dashboard'
    }

    await router.push(redirectTo)
  } catch (error: unknown) {
    // Reset Turnstile on error
    if (turnstileRef.value) {
      turnstileRef.value.reset()
      turnstileToken.value = ''
    }

    // Handle login error
    const err = error as { message?: string; response?: { data?: { detail?: string } } }

    if (err.response?.data?.detail) {
      errorMessage.value = err.response.data.detail
    } else if (err.message) {
      errorMessage.value = err.message
    } else {
      errorMessage.value = t('auth.loginFailed')
    }

    // Also show error toast
    appStore.showError(errorMessage.value)
  } finally {
    isLoading.value = false
  }
}
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
