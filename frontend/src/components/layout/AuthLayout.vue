<template>
  <div class="relative flex min-h-screen items-center justify-center overflow-hidden bg-[#f4f1ea] p-4 dark:bg-dark-950">
    <!-- Background -->
    <div
      class="absolute inset-0 bg-mesh-gradient"
    ></div>

    <!-- Decorative Elements -->
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <!-- Gradient Orbs -->
      <div
        class="absolute -right-40 -top-40 h-80 w-80 rounded-full bg-primary-400/20 blur-3xl"
      ></div>
      <div
        class="absolute -bottom-40 -left-40 h-80 w-80 rounded-full bg-primary-500/15 blur-3xl"
      ></div>
      <div
        class="absolute left-1/2 top-1/2 h-96 w-96 -translate-x-1/2 -translate-y-1/2 rounded-full bg-primary-300/10 blur-3xl"
      ></div>

      <!-- Grid Pattern -->
      <div
        class="absolute inset-0 bg-[linear-gradient(rgba(20,184,166,0.03)_1px,transparent_1px),linear-gradient(90deg,rgba(20,184,166,0.03)_1px,transparent_1px)] bg-[size:64px_64px]"
      ></div>
    </div>

    <!-- Content Container -->
    <div class="relative z-10 w-full max-w-md">
      <!-- Logo/Brand -->
      <div class="mb-8 animate-fade-in-up text-center" style="animation-delay: 0.1s; animation-fill-mode: both">
        <!-- Custom Logo or Default Logo -->
        <div
          class="mb-4 inline-flex h-16 w-16 items-center justify-center overflow-hidden rounded-2xl shadow-glow"
        >
          <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
        </div>
        <h1 class="mb-2 text-3xl font-bold text-primary-600 dark:text-primary-400">
          {{ siteName }}
        </h1>
        <p class="text-sm text-gray-500 dark:text-dark-400">
          {{ siteSubtitle }}
        </p>
      </div>

      <!-- Card Container -->
      <div class="card-glass animate-fade-in-up rounded-2xl p-8 shadow-glass" style="animation-delay: 0.2s; animation-fill-mode: both">
        <slot />
      </div>

      <!-- Footer Links -->
      <div class="mt-6 animate-fade-in-up text-center text-sm" style="animation-delay: 0.3s; animation-fill-mode: both">
        <slot name="footer" />
      </div>

      <!-- Copyright -->
      <div class="mt-8 animate-fade-in-up text-center text-xs text-gray-400 dark:text-dark-500" style="animation-delay: 0.4s; animation-fill-mode: both">
        &copy; {{ currentYear }} {{ siteName }}. All rights reserved.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getPublicSettings } from '@/api/auth'
import { sanitizeUrl } from '@/utils/url'

const siteName = ref('Sub2API')
const siteLogo = ref('')
const siteSubtitle = ref('Subscription to API Conversion Platform')

const currentYear = computed(() => new Date().getFullYear())

onMounted(async () => {
  try {
    const settings = await getPublicSettings()
    siteName.value = settings.site_name || 'Sub2API'
    siteLogo.value = sanitizeUrl(settings.site_logo || '', { allowRelative: true })
    siteSubtitle.value = settings.site_subtitle || 'Subscription to API Conversion Platform'
  } catch (error) {
    console.error('Failed to load public settings:', error)
  }
})
</script>

