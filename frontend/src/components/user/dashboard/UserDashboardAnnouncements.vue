<template>
  <div class="card">
    <!-- 头部：标题 + 未读计数徽章 -->
    <div class="flex items-center justify-between border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
        {{ t('announcements.title') }}
      </h2>
      <span v-if="unreadCount > 0" class="badge badge-primary">
        {{ unreadCount }} {{ t('announcements.unread') }}
      </span>
    </div>

    <!-- 内容区域 -->
    <div class="p-6">
      <!-- 加载状态 -->
      <div v-if="loading" class="flex items-center justify-center py-12">
        <LoadingSpinner size="lg" />
      </div>

      <!-- 空状态 -->
      <div v-else-if="announcements.length === 0" class="py-8">
        <EmptyState
          :title="t('announcements.empty')"
          :description="t('announcements.emptyDescription')"
        />
      </div>

      <!-- 公告列表 -->
      <div v-else class="space-y-3">
        <div
          v-for="item in announcements"
          :key="item.id"
          @click="openDetail(item)"
          class="announcement-item group relative flex cursor-pointer items-center justify-between rounded-xl p-4 transition-all"
          :class="item.read_at ? 'bg-gray-50 hover:bg-gray-100 dark:bg-dark-800/50 dark:hover:bg-dark-800' : 'bg-blue-50/30 hover:bg-blue-50/50 dark:bg-blue-900/5 dark:hover:bg-blue-900/10'"
        >
          <!-- 未读指示器（左侧蓝条） -->
          <div
            v-if="!item.read_at"
            class="absolute left-0 top-0 h-full w-1 rounded-l-xl bg-gradient-to-b from-blue-500 to-indigo-600"
          />

          <!-- 内容 -->
          <div class="flex flex-1 items-center gap-4">
            <!-- 状态图标 -->
            <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl transition-all" :class="!item.read_at ? 'bg-gradient-to-br from-blue-500 to-indigo-600 text-white shadow-lg shadow-blue-500/30' : 'bg-gray-100 text-gray-400 dark:bg-dark-700 dark:text-gray-600'">
              <Icon name="bell" size="md" />
            </div>

            <!-- 标题和时间 -->
            <div class="min-w-0 flex-1">
              <h3 class="truncate text-sm font-medium text-gray-900 dark:text-white">
                {{ item.title }}
              </h3>
              <div class="mt-1 flex items-center gap-2">
                <time class="text-xs text-gray-500 dark:text-gray-400">
                  {{ formatRelativeTime(item.created_at) }}
                </time>
                <span
                  v-if="!item.read_at"
                  class="inline-flex items-center gap-1 rounded-md bg-blue-100 px-1.5 py-0.5 text-xs font-medium text-blue-700 dark:bg-blue-900/40 dark:text-blue-300"
                >
                  <span class="relative flex h-1.5 w-1.5">
                    <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-blue-500 opacity-75"></span>
                    <span class="relative inline-flex h-1.5 w-1.5 rounded-full bg-blue-600"></span>
                  </span>
                  NEW
                </span>
              </div>
            </div>
          </div>

          <!-- 箭头 -->
          <div class="flex-shrink-0">
            <svg
              class="h-5 w-5 text-gray-400 transition-transform group-hover:translate-x-1 dark:text-gray-600"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
            </svg>
          </div>
        </div>
      </div>
    </div>

    <!-- 公告详情弹窗 -->
    <Teleport to="body">
      <Transition name="modal-fade">
        <div
          v-if="detailModalOpen && selectedAnnouncement"
          class="fixed inset-0 z-[110] flex items-start justify-center overflow-y-auto bg-gradient-to-br from-black/70 via-black/60 to-black/70 p-4 pt-[6vh] backdrop-blur-md"
          @click="closeDetail"
        >
          <div
            class="w-full max-w-[780px] overflow-hidden rounded-3xl bg-white shadow-2xl ring-1 ring-black/5 dark:bg-dark-800 dark:ring-white/10"
            @click.stop
          >
            <!-- Header with Decorative Elements -->
            <div class="relative overflow-hidden border-b border-gray-100 bg-gradient-to-br from-blue-50/80 via-indigo-50/50 to-purple-50/30 px-8 py-6 dark:border-dark-700 dark:from-blue-900/20 dark:via-indigo-900/10 dark:to-purple-900/5">
              <!-- Decorative background elements -->
              <div class="absolute right-0 top-0 h-full w-64 bg-gradient-to-l from-indigo-100/30 to-transparent dark:from-indigo-900/20"></div>
              <div class="absolute -right-8 -top-8 h-32 w-32 rounded-full bg-gradient-to-br from-blue-400/20 to-indigo-500/20 blur-3xl"></div>
              <div class="absolute -left-4 -bottom-4 h-24 w-24 rounded-full bg-gradient-to-tr from-purple-400/20 to-pink-500/20 blur-2xl"></div>

              <div class="relative z-10 flex items-start justify-between gap-4">
                <div class="flex-1 min-w-0">
                  <!-- Icon and Category -->
                  <div class="mb-3 flex items-center gap-2">
                    <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-blue-500 to-indigo-600 text-white shadow-lg shadow-blue-500/30">
                      <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </div>
                    <div class="flex items-center gap-2">
                      <span class="rounded-lg bg-blue-100 px-2.5 py-1 text-xs font-medium text-blue-700 dark:bg-blue-900/40 dark:text-blue-300">
                        {{ t('announcements.title') }}
                      </span>
                      <span
                        v-if="!selectedAnnouncement.read_at"
                        class="inline-flex items-center gap-1.5 rounded-lg bg-gradient-to-r from-blue-500 to-indigo-600 px-2.5 py-1 text-xs font-medium text-white shadow-lg shadow-blue-500/30"
                      >
                        <span class="relative flex h-2 w-2">
                          <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-white opacity-75"></span>
                          <span class="relative inline-flex h-2 w-2 rounded-full bg-white"></span>
                        </span>
                        {{ t('announcements.unread') }}
                      </span>
                    </div>
                  </div>

                  <!-- Title -->
                  <h2 class="mb-3 text-2xl font-bold leading-tight text-gray-900 dark:text-white">
                    {{ selectedAnnouncement.title }}
                  </h2>

                  <!-- Meta Info -->
                  <div class="flex items-center gap-4 text-sm text-gray-600 dark:text-gray-400">
                    <div class="flex items-center gap-1.5">
                      <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                      <time>{{ formatRelativeWithDateTime(selectedAnnouncement.created_at) }}</time>
                    </div>
                    <div class="flex items-center gap-1.5">
                      <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                        <path stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                      </svg>
                      <span>{{ selectedAnnouncement.read_at ? t('announcements.read') : t('announcements.unread') }}</span>
                    </div>
                  </div>
                </div>

                <!-- Close button -->
                <button
                  @click="closeDetail"
                  class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl bg-white/50 text-gray-500 backdrop-blur-sm transition-all hover:bg-white hover:text-gray-700 hover:shadow-lg dark:bg-dark-700/50 dark:text-gray-400 dark:hover:bg-dark-700 dark:hover:text-gray-300"
                  :aria-label="t('common.close')"
                >
                  <Icon name="x" size="md" />
                </button>
              </div>
            </div>

            <!-- Body with Enhanced Markdown -->
            <div class="max-h-[60vh] overflow-y-auto bg-white px-8 py-8 dark:bg-dark-800">
              <!-- Content with decorative border -->
              <div class="relative">
                <!-- Decorative left border -->
                <div class="absolute left-0 top-0 bottom-0 w-1 rounded-full bg-gradient-to-b from-blue-500 via-indigo-500 to-purple-500"></div>

                <div class="pl-6">
                  <div
                    class="markdown-body prose prose-sm max-w-none dark:prose-invert"
                    v-html="renderMarkdown(selectedAnnouncement.content)"
                  ></div>
                </div>
              </div>
            </div>

            <!-- Footer with Actions -->
            <div class="border-t border-gray-100 bg-gray-50/50 px-8 py-5 dark:border-dark-700 dark:bg-dark-900/30">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <span>{{ selectedAnnouncement.read_at ? t('announcements.readStatus') : t('announcements.markReadHint') }}</span>
                </div>
                <div class="flex items-center gap-3">
                  <button
                    @click="closeDetail"
                    class="rounded-xl border border-gray-300 bg-white px-5 py-2.5 text-sm font-medium text-gray-700 shadow-sm transition-all hover:bg-gray-50 hover:shadow dark:border-dark-600 dark:bg-dark-700 dark:text-gray-300 dark:hover:bg-dark-600"
                  >
                    {{ t('common.close') }}
                  </button>
                  <button
                    v-if="!selectedAnnouncement.read_at"
                    @click="markAsReadAndClose(selectedAnnouncement.id)"
                    class="rounded-xl bg-gradient-to-r from-blue-600 to-indigo-600 px-5 py-2.5 text-sm font-medium text-white shadow-lg shadow-blue-500/30 transition-all hover:shadow-xl hover:scale-105"
                  >
                    <span class="flex items-center gap-2">
                      <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                      </svg>
                      {{ t('announcements.markRead') }}
                    </span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { announcementsAPI } from '@/api'
import { useAppStore } from '@/stores/app'
import { formatRelativeTime, formatRelativeWithDateTime } from '@/utils/format'
import type { UserAnnouncement } from '@/types'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import EmptyState from '@/components/common/EmptyState.vue'

const { t } = useI18n()
const appStore = useAppStore()

// Configure marked
marked.setOptions({
  breaks: true,
  gfm: true,
})

// State
const announcements = ref<UserAnnouncement[]>([])
const detailModalOpen = ref(false)
const selectedAnnouncement = ref<UserAnnouncement | null>(null)
const loading = ref(false)

// Computed
const unreadCount = computed(() =>
  announcements.value.filter((a) => !a.read_at).length
)

// Methods
function renderMarkdown(content: string): string {
  if (!content) return ''
  const html = marked.parse(content) as string
  return DOMPurify.sanitize(html)
}

async function loadAnnouncements() {
  try {
    loading.value = true
    announcements.value = await announcementsAPI.list(false)
  } catch (err: any) {
    console.error('Failed to load announcements:', err)
    appStore.showError(err?.message || t('common.unknownError'))
  } finally {
    loading.value = false
  }
}

function openDetail(announcement: UserAnnouncement) {
  selectedAnnouncement.value = announcement
  detailModalOpen.value = true
  if (!announcement.read_at) {
    markAsRead(announcement.id)
  }
}

function closeDetail() {
  detailModalOpen.value = false
  selectedAnnouncement.value = null
}

async function markAsRead(id: number) {
  try {
    await announcementsAPI.markRead(id)
    const announcement = announcements.value.find((a) => a.id === id)
    if (announcement) {
      announcement.read_at = new Date().toISOString()
    }
    if (selectedAnnouncement.value?.id === id) {
      selectedAnnouncement.value.read_at = new Date().toISOString()
    }
  } catch (err: any) {
    appStore.showError(err?.message || t('common.unknownError'))
  }
}

async function markAsReadAndClose(id: number) {
  await markAsRead(id)
  appStore.showSuccess(t('announcements.markedAsRead'))
  closeDetail()
}

function handleEscape(e: KeyboardEvent) {
  if (e.key === 'Escape' && detailModalOpen.value) {
    closeDetail()
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleEscape)
  loadAnnouncements()
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleEscape)
  // Restore body overflow in case component is unmounted while modal is open
  document.body.style.overflow = ''
})

watch(detailModalOpen, (isOpen) => {
  if (isOpen) {
    document.body.style.overflow = 'hidden'
  } else {
    document.body.style.overflow = ''
  }
})
</script>

<style scoped>
/* Modal Animations */
.modal-fade-enter-active {
  transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.modal-fade-leave-active {
  transition: all 0.2s cubic-bezier(0.4, 0, 1, 1);
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}

.modal-fade-enter-from > div {
  transform: scale(0.94) translateY(-12px);
  opacity: 0;
}

.modal-fade-leave-to > div {
  transform: scale(0.96) translateY(-8px);
  opacity: 0;
}

/* Scrollbar Styling */
.overflow-y-auto::-webkit-scrollbar {
  width: 8px;
}

.overflow-y-auto::-webkit-scrollbar-track {
  background: transparent;
}

.overflow-y-auto::-webkit-scrollbar-thumb {
  background: linear-gradient(to bottom, #cbd5e1, #94a3b8);
  border-radius: 4px;
}

.dark .overflow-y-auto::-webkit-scrollbar-thumb {
  background: linear-gradient(to bottom, #4b5563, #374151);
}

.overflow-y-auto::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(to bottom, #94a3b8, #64748b);
}

.dark .overflow-y-auto::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(to bottom, #6b7280, #4b5563);
}
</style>

<style>
/* Enhanced Markdown Styles */
.markdown-body {
  @apply text-[15px] leading-[1.75];
  @apply text-gray-700 dark:text-gray-300;
}

.markdown-body h1 {
  @apply mb-6 mt-8 border-b border-gray-200 pb-3 text-3xl font-bold text-gray-900 dark:border-dark-600 dark:text-white;
}

.markdown-body h2 {
  @apply mb-4 mt-7 border-b border-gray-100 pb-2 text-2xl font-bold text-gray-900 dark:border-dark-700 dark:text-white;
}

.markdown-body h3 {
  @apply mb-3 mt-6 text-xl font-semibold text-gray-900 dark:text-white;
}

.markdown-body h4 {
  @apply mb-2 mt-5 text-lg font-semibold text-gray-900 dark:text-white;
}

.markdown-body p {
  @apply mb-4 leading-relaxed;
}

.markdown-body a {
  @apply font-medium text-blue-600 underline decoration-blue-600/30 decoration-2 underline-offset-2 transition-all hover:decoration-blue-600 dark:text-blue-400 dark:decoration-blue-400/30 dark:hover:decoration-blue-400;
}

.markdown-body ul,
.markdown-body ol {
  @apply mb-4 ml-6 space-y-2;
}

.markdown-body ul {
  @apply list-disc;
}

.markdown-body ol {
  @apply list-decimal;
}

.markdown-body li {
  @apply leading-relaxed;
  @apply pl-2;
}

.markdown-body li::marker {
  @apply text-blue-600 dark:text-blue-400;
}

.markdown-body blockquote {
  @apply relative my-5 border-l-4 border-blue-500 bg-blue-50/50 py-3 pl-5 pr-4 italic text-gray-700 dark:border-blue-400 dark:bg-blue-900/10 dark:text-gray-300;
}

.markdown-body blockquote::before {
  content: '"';
  @apply absolute -left-1 top-0 text-5xl font-serif text-blue-500/20 dark:text-blue-400/20;
}

.markdown-body code {
  @apply rounded-lg bg-gray-100 px-2 py-1 text-[13px] font-mono text-pink-600 dark:bg-dark-700 dark:text-pink-400;
}

.markdown-body pre {
  @apply my-5 overflow-x-auto rounded-xl border border-gray-200 bg-gray-50 p-5 dark:border-dark-600 dark:bg-dark-900/50;
}

.markdown-body pre code {
  @apply bg-transparent p-0 text-[13px] text-gray-800 dark:text-gray-200;
}

.markdown-body hr {
  @apply my-8 border-0 border-t-2 border-gray-200 dark:border-dark-700;
}

.markdown-body table {
  @apply mb-5 w-full overflow-hidden rounded-lg border border-gray-200 dark:border-dark-600;
}

.markdown-body th,
.markdown-body td {
  @apply border-r border-b border-gray-200 px-4 py-3 text-left dark:border-dark-600;
}

.markdown-body th:last-child,
.markdown-body td:last-child {
  @apply border-r-0;
}

.markdown-body tr:last-child td {
  @apply border-b-0;
}

.markdown-body th {
  @apply bg-gradient-to-br from-blue-50 to-indigo-50 font-semibold text-gray-900 dark:from-blue-900/20 dark:to-indigo-900/10 dark:text-white;
}

.markdown-body tbody tr {
  @apply transition-colors hover:bg-gray-50 dark:hover:bg-dark-700/30;
}

.markdown-body img {
  @apply my-5 max-w-full rounded-xl border border-gray-200 shadow-md dark:border-dark-600;
}

.markdown-body strong {
  @apply font-semibold text-gray-900 dark:text-white;
}

.markdown-body em {
  @apply italic text-gray-600 dark:text-gray-400;
}
</style>
