<template>
  <div class="relative min-h-screen bg-[#f4f1ea] dark:bg-dark-950">
    <!-- Header -->
    <header class="sticky top-0 z-50 border-b border-[#1a1a1a]/10 bg-[#f4f1ea]/95 backdrop-blur-sm dark:border-dark-700 dark:bg-dark-900/95">
      <nav class="mx-auto flex max-w-7xl items-center justify-between px-6 py-4 lg:px-8">
        <!-- Logo -->
        <div class="flex items-center gap-3">
          <div class="h-10 w-10 overflow-hidden rounded-lg">
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <span class="text-xl font-bold text-[#1a1a1a] dark:text-white">
            {{ siteName }}
          </span>
        </div>

        <!-- Nav Actions -->
        <div class="flex items-center gap-4">
          <LocaleSwitcher />

          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="text-sm text-[#4a4a4a] transition-colors hover:text-[#C44A2C] dark:text-dark-400 dark:hover:text-[#C44A2C]"
          >
            {{ t('home.docs') }}
          </a>

          <button
            @click="toggleTheme"
            class="rounded-lg p-2 text-[#4a4a4a] transition-colors hover:bg-[#1a1a1a]/5 dark:text-dark-400 dark:hover:bg-dark-800"
          >
            <svg
              v-if="isDark"
              class="h-5 w-5"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="1.5"
                d="M12 3v2.25m6.364.386l-1.591 1.591M21 12h-2.25m-.386 6.364l-1.591-1.591M12 18.75V21m-4.773-4.227l-1.591 1.591M5.25 12H3m4.227-4.773L5.636 5.636M15.75 12a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0z"
              />
            </svg>
            <svg v-else class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="1.5"
                d="M21.752 15.002A9.718 9.718 0 0118 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 003 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 009.002-5.998z"
              />
            </svg>
          </button>

          <router-link
            v-if="isAuthenticated"
            to="/dashboard"
            class="rounded-full bg-[#C44A2C] px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-[#A33D24]"
          >
            {{ t('home.dashboard') }}
          </router-link>
          <router-link
            v-else
            to="/login"
            class="rounded-full bg-[#C44A2C] px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-[#A33D24]"
          >
            {{ t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <!-- Hero Section -->
    <section class="mx-auto max-w-7xl px-6 py-20 lg:px-8 lg:py-32">
      <div class="mx-auto max-w-4xl text-center">
        <h1
          class="hero-title mb-6 text-5xl font-bold leading-tight text-[#1a1a1a] dark:text-white md:text-6xl lg:text-7xl"
        >
          {{ siteName }}
        </h1>
        <p
          class="hero-subtitle mb-10 text-xl leading-relaxed text-[#4a4a4a] dark:text-dark-300 md:text-2xl"
        >
          {{ siteSubtitle }}
        </p>
        <div class="hero-cta flex flex-col items-center justify-center gap-4 sm:flex-row">
          <router-link
            :to="isAuthenticated ? '/dashboard' : '/login'"
            class="group inline-flex items-center gap-2 rounded-full bg-[#C44A2C] px-8 py-4 text-base font-medium text-white transition-all hover:bg-[#A33D24] hover:shadow-lg"
          >
            {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
            <svg
              class="h-5 w-5 transition-transform group-hover:translate-x-1"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M13.5 4.5L21 12m0 0l-7.5 7.5M21 12H3"
              />
            </svg>
          </router-link>
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex items-center gap-2 rounded-full border border-[#1a1a1a]/20 px-8 py-4 text-base font-medium text-[#1a1a1a] transition-all hover:border-[#C44A2C] hover:text-[#C44A2C] dark:border-dark-600 dark:text-white dark:hover:border-[#C44A2C] dark:hover:text-[#C44A2C]"
          >
            {{ t('home.viewDocs') }}
          </a>
        </div>
      </div>
    </section>

    <!-- Feature Tags -->
    <section class="border-y border-[#1a1a1a]/10 bg-white/50 py-8 dark:border-dark-700 dark:bg-dark-900/50">
      <div class="mx-auto max-w-7xl px-6 lg:px-8">
        <div class="flex flex-wrap items-center justify-center gap-6 md:gap-8">
          <div class="feature-tag flex items-center gap-2.5">
            <svg
              class="h-5 w-5 text-[#C44A2C]"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="1.5"
                d="M7.5 21L3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5"
              />
            </svg>
            <span class="text-sm font-medium text-[#1a1a1a] dark:text-dark-200">
              {{ t('home.tags.subscriptionToApi') }}
            </span>
          </div>
          <div class="feature-tag flex items-center gap-2.5">
            <svg
              class="h-5 w-5 text-[#C44A2C]"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="1.5"
                d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z"
              />
            </svg>
            <span class="text-sm font-medium text-[#1a1a1a] dark:text-dark-200">
              {{ t('home.tags.stickySession') }}
            </span>
          </div>
          <div class="feature-tag flex items-center gap-2.5">
            <svg
              class="h-5 w-5 text-[#C44A2C]"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="1.5"
                d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z"
              />
            </svg>
            <span class="text-sm font-medium text-[#1a1a1a] dark:text-dark-200">
              {{ t('home.tags.realtimeBilling') }}
            </span>
          </div>
        </div>
      </div>
    </section>

    <!-- Features Section -->
    <section class="mx-auto max-w-7xl px-6 py-20 lg:px-8 lg:py-24">
      <div class="mb-16 text-center">
        <h2 class="section-title mb-4 text-3xl font-bold text-[#1a1a1a] dark:text-white md:text-4xl">
          核心功能
        </h2>
        <p class="text-lg text-[#4a4a4a] dark:text-dark-300">
          专业的 AI API 网关解决方案
        </p>
      </div>

      <div class="grid gap-8 md:grid-cols-2 lg:grid-cols-3">
        <!-- Feature 1 -->
        <div class="feature-card group">
          <div class="mb-6">
            <div
              class="inline-flex h-14 w-14 items-center justify-center rounded-xl bg-[#fde9e3] text-[#C44A2C] transition-transform group-hover:scale-110 dark:bg-[#C44A2C]/20 dark:text-[#f08f6f]"
            >
              <svg
                class="h-7 w-7"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="1.5"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M5.25 14.25h13.5m-13.5 0a3 3 0 01-3-3m3 3a3 3 0 100 6h13.5a3 3 0 100-6m-16.5-3a3 3 0 013-3h13.5a3 3 0 013 3m-19.5 0a4.5 4.5 0 01.9-2.7L5.737 5.1a3.375 3.375 0 012.7-1.35h7.126c1.062 0 2.062.5 2.7 1.35l2.587 3.45a4.5 4.5 0 01.9 2.7m0 0a3 3 0 01-3 3m0 3h.008v.008h-.008v-.008zm0-6h.008v.008h-.008v-.008zm-3 6h.008v.008h-.008v-.008zm0-6h.008v.008h-.008v-.008z"
                />
              </svg>
            </div>
          </div>
          <h3 class="mb-3 text-xl font-bold text-[#1a1a1a] dark:text-white">
            {{ t('home.features.unifiedGateway') }}
          </h3>
          <p class="leading-relaxed text-[#4a4a4a] dark:text-dark-400">
            {{ t('home.features.unifiedGatewayDesc') }}
          </p>
        </div>

        <!-- Feature 2 -->
        <div class="feature-card group">
          <div class="mb-6">
            <div
              class="inline-flex h-14 w-14 items-center justify-center rounded-xl bg-[#fde9e3] text-[#C44A2C] transition-transform group-hover:scale-110 dark:bg-[#C44A2C]/20 dark:text-[#f08f6f]"
            >
              <svg
                class="h-7 w-7"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="1.5"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M18 18.72a9.094 9.094 0 003.741-.479 3 3 0 00-4.682-2.72m.94 3.198l.001.031c0 .225-.012.447-.037.666A11.944 11.944 0 0112 21c-2.17 0-4.207-.576-5.963-1.584A6.062 6.062 0 016 18.719m12 0a5.971 5.971 0 00-.941-3.197m0 0A5.995 5.995 0 0012 12.75a5.995 5.995 0 00-5.058 2.772m0 0a3 3 0 00-4.681 2.72 8.986 8.986 0 003.74.477m.94-3.197a5.971 5.971 0 00-.94 3.197M15 6.75a3 3 0 11-6 0 3 3 0 016 0zm6 3a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0zm-13.5 0a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0z"
                />
              </svg>
            </div>
          </div>
          <h3 class="mb-3 text-xl font-bold text-[#1a1a1a] dark:text-white">
            {{ t('home.features.multiAccount') }}
          </h3>
          <p class="leading-relaxed text-[#4a4a4a] dark:text-dark-400">
            {{ t('home.features.multiAccountDesc') }}
          </p>
        </div>

        <!-- Feature 3 -->
        <div class="feature-card group">
          <div class="mb-6">
            <div
              class="inline-flex h-14 w-14 items-center justify-center rounded-xl bg-[#fde9e3] text-[#C44A2C] transition-transform group-hover:scale-110 dark:bg-[#C44A2C]/20 dark:text-[#f08f6f]"
            >
              <svg
                class="h-7 w-7"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="1.5"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z"
                />
              </svg>
            </div>
          </div>
          <h3 class="mb-3 text-xl font-bold text-[#1a1a1a] dark:text-white">
            {{ t('home.features.balanceQuota') }}
          </h3>
          <p class="leading-relaxed text-[#4a4a4a] dark:text-dark-400">
            {{ t('home.features.balanceQuotaDesc') }}
          </p>
        </div>
      </div>
    </section>

    <!-- Providers Section -->
    <section class="border-t border-[#1a1a1a]/10 bg-white/50 py-20 dark:border-dark-700 dark:bg-dark-900/50 lg:py-24">
      <div class="mx-auto max-w-7xl px-6 lg:px-8">
        <div class="mb-12 text-center">
          <h2 class="section-title mb-4 text-3xl font-bold text-[#1a1a1a] dark:text-white md:text-4xl">
            {{ t('home.providers.title') }}
          </h2>
          <p class="text-lg text-[#4a4a4a] dark:text-dark-300">
            {{ t('home.providers.description') }}
          </p>
        </div>

        <div class="flex flex-wrap items-center justify-center gap-4">
          <!-- Claude -->
          <div class="provider-card">
            <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br from-orange-400 to-orange-500">
              <span class="text-sm font-bold text-white">C</span>
            </div>
            <span class="text-sm font-medium text-[#1a1a1a] dark:text-dark-200">
              {{ t('home.providers.claude') }}
            </span>
            <span class="rounded bg-[#fde9e3] px-2 py-0.5 text-xs font-medium text-[#a33d24] dark:bg-[#C44A2C]/20 dark:text-[#f08f6f]">
              {{ t('home.providers.supported') }}
            </span>
          </div>

          <!-- GPT -->
          <div class="provider-card">
            <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br from-green-500 to-green-600">
              <span class="text-sm font-bold text-white">G</span>
            </div>
            <span class="text-sm font-medium text-[#1a1a1a] dark:text-dark-200">GPT</span>
            <span class="rounded bg-[#fde9e3] px-2 py-0.5 text-xs font-medium text-[#a33d24] dark:bg-[#C44A2C]/20 dark:text-[#f08f6f]">
              {{ t('home.providers.supported') }}
            </span>
          </div>

          <!-- Gemini -->
          <div class="provider-card">
            <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-blue-600">
              <span class="text-sm font-bold text-white">G</span>
            </div>
            <span class="text-sm font-medium text-[#1a1a1a] dark:text-dark-200">
              {{ t('home.providers.gemini') }}
            </span>
            <span class="rounded bg-[#fde9e3] px-2 py-0.5 text-xs font-medium text-[#a33d24] dark:bg-[#C44A2C]/20 dark:text-[#f08f6f]">
              {{ t('home.providers.supported') }}
            </span>
          </div>

          <!-- Antigravity -->
          <div class="provider-card">
            <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br from-rose-500 to-pink-600">
              <span class="text-sm font-bold text-white">A</span>
            </div>
            <span class="text-sm font-medium text-[#1a1a1a] dark:text-dark-200">
              {{ t('home.providers.antigravity') }}
            </span>
            <span class="rounded bg-[#fde9e3] px-2 py-0.5 text-xs font-medium text-[#a33d24] dark:bg-[#C44A2C]/20 dark:text-[#f08f6f]">
              {{ t('home.providers.supported') }}
            </span>
          </div>

          <!-- More -->
          <div class="provider-card opacity-60">
            <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br from-gray-400 to-gray-500">
              <span class="text-sm font-bold text-white">+</span>
            </div>
            <span class="text-sm font-medium text-[#1a1a1a] dark:text-dark-200">
              {{ t('home.providers.more') }}
            </span>
            <span class="rounded bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-600 dark:bg-dark-700 dark:text-dark-400">
              {{ t('home.providers.soon') }}
            </span>
          </div>
        </div>
      </div>
    </section>

    <!-- Terminal Demo Section -->
    <section class="mx-auto max-w-7xl px-6 py-20 lg:px-8 lg:py-24">
      <div class="mx-auto max-w-4xl">
        <div class="mb-12 text-center">
          <h2 class="section-title mb-4 text-3xl font-bold text-[#1a1a1a] dark:text-white md:text-4xl">
            快速开始
          </h2>
          <p class="text-lg text-[#4a4a4a] dark:text-dark-300">
            简单的 API 调用,强大的功能
          </p>
        </div>

        <div class="terminal-demo">
          <div class="terminal-window">
            <div class="terminal-header">
              <div class="terminal-buttons">
                <span class="btn-close"></span>
                <span class="btn-minimize"></span>
                <span class="btn-maximize"></span>
              </div>
              <span class="terminal-title">terminal</span>
            </div>
            <div class="terminal-body">
              <div class="code-line line-1">
                <span class="code-prompt">$</span>
                <span class="code-cmd">curl</span>
                <span class="code-flag">-X POST</span>
                <span class="code-url">/v1/messages</span>
              </div>
              <div class="code-line line-2">
                <span class="code-comment"># Routing to upstream...</span>
              </div>
              <div class="code-line line-3">
                <span class="code-success">200 OK</span>
                <span class="code-response">{ "content": "Hello!" }</span>
              </div>
              <div class="code-line line-4">
                <span class="code-prompt">$</span>
                <span class="cursor"></span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Footer -->
    <footer class="border-t border-[#1a1a1a]/10 py-12 dark:border-dark-700">
      <div class="mx-auto max-w-7xl px-6 lg:px-8">
        <div class="flex flex-col items-center justify-between gap-4 sm:flex-row">
          <p class="text-sm text-[#4a4a4a] dark:text-dark-400">
            &copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}
          </p>
          <div class="flex items-center gap-6">
            <a
              v-if="docUrl"
              :href="docUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="text-sm text-[#4a4a4a] transition-colors hover:text-[#C44A2C] dark:text-dark-400 dark:hover:text-[#C44A2C]"
            >
              {{ t('home.docs') }}
            </a>
            <a
              :href="githubUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="text-sm text-[#4a4a4a] transition-colors hover:text-[#C44A2C] dark:text-dark-400 dark:hover:text-[#C44A2C]"
            >
              GitHub
            </a>
          </div>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { getPublicSettings } from '@/api/auth'
import { useAuthStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'

const { t } = useI18n()

const authStore = useAuthStore()

// Site settings
const siteName = ref('Sub2API')
const siteLogo = ref('')
const siteSubtitle = ref('AI API Gateway Platform')
const docUrl = ref('')

// Theme
const isDark = ref(document.documentElement.classList.contains('dark'))

// GitHub URL
const githubUrl = 'https://github.com/Wei-Shaw/sub2api'

// Auth state
const isAuthenticated = computed(() => authStore.isAuthenticated)

// Current year for footer
const currentYear = computed(() => new Date().getFullYear())

// Toggle theme
function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

// Initialize theme
function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(async () => {
  initTheme()
  authStore.checkAuth()

  try {
    const settings = await getPublicSettings()
    siteName.value = settings.site_name || 'Sub2API'
    siteLogo.value = settings.site_logo || ''
    siteSubtitle.value = settings.site_subtitle || 'AI API Gateway Platform'
    docUrl.value = settings.doc_url || ''
  } catch (error) {
    console.error('Failed to load public settings:', error)
  }
})
</script>

<style scoped>
/* Hero Animations */
.hero-title {
  opacity: 0;
  animation: fadeInUp 0.3s ease-out forwards;
  animation-delay: 0.15s;
}

.hero-subtitle {
  opacity: 0;
  animation: fadeInUp 0.3s ease-out forwards;
  animation-delay: 0.3s;
}

.hero-cta {
  opacity: 0;
  animation: fadeInUp 0.3s ease-out forwards;
  animation-delay: 0.45s;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Section Animations */
.section-title {
  opacity: 0;
  animation: fadeInUp 0.3s ease-out forwards;
}

.feature-tag {
  opacity: 0;
  animation: fadeInUp 0.3s ease-out forwards;
}

.feature-tag:nth-child(1) {
  animation-delay: 0.15s;
}
.feature-tag:nth-child(2) {
  animation-delay: 0.3s;
}
.feature-tag:nth-child(3) {
  animation-delay: 0.45s;
}

/* Feature Cards */
.feature-card {
  opacity: 0;
  animation: fadeInUp 0.3s ease-out forwards;
  padding: 2rem;
  border-radius: 1rem;
  background: white;
  border: 1px solid rgba(26, 26, 26, 0.08);
  transition: all 0.3s ease;
}

.dark .feature-card {
  background: rgba(15, 23, 42, 0.5);
  border-color: rgba(51, 65, 85, 0.5);
}

.feature-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 24px rgba(196, 74, 44, 0.12);
}

.feature-card:nth-child(1) {
  animation-delay: 0.15s;
}
.feature-card:nth-child(2) {
  animation-delay: 0.3s;
}
.feature-card:nth-child(3) {
  animation-delay: 0.45s;
}

/* Provider Cards */
.provider-card {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  border-radius: 0.75rem;
  background: white;
  border: 1px solid rgba(26, 26, 26, 0.08);
  opacity: 0;
  animation: fadeInUp 0.3s ease-out forwards;
  transition: all 0.3s ease;
}

.dark .provider-card {
  background: rgba(15, 23, 42, 0.5);
  border-color: rgba(51, 65, 85, 0.5);
}

.provider-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 16px rgba(196, 74, 44, 0.1);
}

.provider-card:nth-child(1) {
  animation-delay: 0.15s;
}
.provider-card:nth-child(2) {
  animation-delay: 0.3s;
}
.provider-card:nth-child(3) {
  animation-delay: 0.45s;
}
.provider-card:nth-child(4) {
  animation-delay: 0.6s;
}
.provider-card:nth-child(5) {
  animation-delay: 0.75s;
}

/* Terminal Demo */
.terminal-demo {
  opacity: 0;
  animation: fadeInUp 0.3s ease-out forwards;
  animation-delay: 0.15s;
}

.terminal-window {
  background: linear-gradient(145deg, #1e293b 0%, #0f172a 100%);
  border-radius: 1rem;
  overflow: hidden;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
}

.terminal-header {
  display: flex;
  align-items: center;
  padding: 1rem 1.25rem;
  background: rgba(30, 41, 59, 0.8);
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.terminal-buttons {
  display: flex;
  gap: 0.5rem;
}

.terminal-buttons span {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.btn-close {
  background: #ef4444;
}
.btn-minimize {
  background: #eab308;
}
.btn-maximize {
  background: #22c55e;
}

.terminal-title {
  flex: 1;
  text-align: center;
  font-size: 0.75rem;
  font-family: 'JetBrains Mono', monospace;
  color: #64748b;
  margin-right: 60px;
}

.terminal-body {
  padding: 1.5rem;
  font-family: 'JetBrains Mono', monospace;
  font-size: 0.875rem;
  line-height: 2;
}

.code-line {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
  opacity: 0;
  animation: lineAppear 0.5s ease forwards;
}

.line-1 {
  animation-delay: 0.5s;
}
.line-2 {
  animation-delay: 1.2s;
}
.line-3 {
  animation-delay: 2s;
}
.line-4 {
  animation-delay: 2.7s;
}

@keyframes lineAppear {
  from {
    opacity: 0;
    transform: translateY(5px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.code-prompt {
  color: #22c55e;
  font-weight: bold;
}
.code-cmd {
  color: #38bdf8;
}
.code-flag {
  color: #a78bfa;
}
.code-url {
  color: #c44a2c;
}
.code-comment {
  color: #64748b;
  font-style: italic;
}
.code-success {
  color: #22c55e;
  background: rgba(34, 197, 94, 0.15);
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 600;
}
.code-response {
  color: #fbbf24;
}

.cursor {
  display: inline-block;
  width: 8px;
  height: 16px;
  background: #22c55e;
  animation: blink 1s step-end infinite;
}

@keyframes blink {
  0%,
  50% {
    opacity: 1;
  }
  51%,
  100% {
    opacity: 0;
  }
}
</style>
