<template>
  <div class="min-h-screen bg-[#f4f1ea] dark:bg-dark-950">
    <!-- Background Decoration -->
    <div class="pointer-events-none fixed inset-0 bg-mesh-gradient"></div>

    <!-- Sidebar -->
    <AppSidebar />

    <!-- Main Content Area -->
    <div
      class="relative min-h-screen transition-all duration-300"
      :class="[sidebarCollapsed ? 'lg:ml-[72px]' : 'lg:ml-64']"
    >
      <!-- Header -->
      <AppHeader ref="appHeaderRef" />

      <!-- Main Content -->
      <main class="relative p-4 md:p-6 lg:p-8">
        <!-- Content area subtle gradient overlay -->
        <div class="pointer-events-none absolute inset-0 bg-gradient-to-b from-primary-50/30 via-transparent to-transparent dark:from-primary-950/10"></div>
        <div class="relative animate-fade-in-up">
          <slot />
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import '@/styles/onboarding.css'
import { computed, onMounted, ref } from 'vue'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { useOnboardingTour } from '@/composables/useOnboardingTour'
import { useAnnouncementAutoPopup } from '@/composables/useAnnouncementAutoPopup'
import { useOnboardingStore } from '@/stores/onboarding'
import AppSidebar from './AppSidebar.vue'
import AppHeader from './AppHeader.vue'

const appStore = useAppStore()
const authStore = useAuthStore()
const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)
const isAdmin = computed(() => authStore.user?.role === 'admin')

// AppHeader ref (用于访问公告弹窗)
const appHeaderRef = ref<InstanceType<typeof AppHeader> | null>(null)

const { replayTour } = useOnboardingTour({
  storageKey: isAdmin.value ? 'admin_guide' : 'user_guide',
  autoStart: true
})

const onboardingStore = useOnboardingStore()

// 公告自动弹窗（延迟 1.5 秒，避免与新手引导冲突）
useAnnouncementAutoPopup({
  delay: 1500,
  onShowAnnouncements: () => {
    appHeaderRef.value?.openAnnouncementModal()
  }
})

onMounted(() => {
  onboardingStore.setReplayCallback(replayTour)
})

defineExpose({ replayTour })
</script>
