<template>
  <AppLayout>
    <div class="space-y-6">
      <section class="card p-4 md:p-5">
        <div class="mb-5 flex flex-wrap items-start justify-between gap-3">
          <div class="flex flex-wrap items-center gap-3">
            <Select
              :model-value="selectedPlatform"
              class="w-full sm:w-48"
              :options="platformFilterOptions"
              @update:model-value="onPlatformFilterChange"
            />
          </div>
          <div class="ml-auto flex items-center gap-3">
            <button
              @click="loadData"
              :disabled="loading"
              class="btn btn-secondary"
              :title="t('common.refresh')"
            >
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>

        <div class="mb-4" />

        <div v-if="errorMessage" class="mb-4 rounded-lg bg-red-50 px-3 py-2 text-xs text-red-600 dark:bg-red-900/20 dark:text-red-400">
          {{ errorMessage }}
        </div>

        <div v-if="loading" class="py-4 text-center text-sm text-gray-500 dark:text-gray-400">
          {{ t('groupMonitor.loading') }}
        </div>
        <div v-else-if="filteredMonitorItems.length === 0" class="rounded-lg bg-gray-50 px-3 py-4 text-sm text-gray-500 dark:bg-dark-800 dark:text-gray-400">
          {{ t('groupMonitor.empty') }}
        </div>

        <div v-else class="space-y-6">
          <section
            v-for="section in monitorSections"
            :key="section.type"
          >
            <h3 class="mb-3 text-sm font-semibold text-gray-900 dark:text-gray-100">
              {{ section.title }}
            </h3>
            <div class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
              <article
                v-for="group in section.items"
                :key="group.group_id"
                class="relative overflow-hidden rounded-2xl border p-4 shadow-sm transition-all duration-200 hover:-translate-y-0.5 hover:shadow-md"
                :class="statusCardClass(group.current_status)"
              >
                <div class="absolute inset-x-0 top-0 h-1" :class="statusAccentClass(group.current_status)" />
                <div class="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <div class="text-sm font-semibold text-gray-900 dark:text-gray-100">
                      {{ group.group_name }}
                    </div>
                    <div class="mt-2">
                      <span class="inline-flex items-center rounded-full border px-2 py-0.5 text-[11px] font-medium" :class="platformBadgeClass(group.platform)">
                        {{ group.platform }}
                      </span>
                    </div>
                  </div>
                  <span class="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-semibold" :class="statusBadgeClass(group.current_status)">
                    <span class="h-2 w-2 rounded-full" :class="statusDotClass(group.current_status)" />
                    {{ statusLabel(group.current_status) }}
                  </span>
                </div>
              </article>
            </div>
          </section>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import Select from '@/components/common/Select.vue'
import {
  userGroupsAPI,
  type PublicGroupMonitorItem,
  type PublicGroupMonitorResponse
} from '@/api/groups'

const { t } = useI18n()

const loading = ref(false)
const errorMessage = ref('')
const monitorData = ref<PublicGroupMonitorResponse | null>(null)
const selectedPlatform = ref('')

const monitorItems = computed<PublicGroupMonitorItem[]>(() => monitorData.value?.items || [])
const platformOptions = computed<string[]>(() => {
  const set = new Set<string>()
  for (const item of monitorItems.value) {
    const platform = String(item.platform || '').trim()
    if (platform) set.add(platform)
  }
  return Array.from(set).sort((a, b) => a.localeCompare(b))
})
const platformFilterOptions = computed(() => [
  { value: '', label: t('groupMonitor.allPlatforms') },
  ...platformOptions.value.map((platform) => ({ value: platform, label: platform }))
])
const filteredMonitorItems = computed<PublicGroupMonitorItem[]>(() => {
  const platform = selectedPlatform.value.trim()
  if (!platform) return monitorItems.value
  return monitorItems.value.filter((item) => item.platform === platform)
})
const publicMonitorItems = computed<PublicGroupMonitorItem[]>(() =>
  filteredMonitorItems.value.filter((item) => resolveGroupType(item) === 'public')
)
const subscriptionMonitorItems = computed<PublicGroupMonitorItem[]>(() =>
  filteredMonitorItems.value.filter((item) => resolveGroupType(item) === 'subscription')
)
const monitorSections = computed(() => {
  const sections: Array<{ type: 'public' | 'subscription'; title: string; items: PublicGroupMonitorItem[] }> = []
  if (publicMonitorItems.value.length > 0) {
    sections.push({
      type: 'public',
      title: t('groupMonitor.publicType'),
      items: publicMonitorItems.value
    })
  }
  if (subscriptionMonitorItems.value.length > 0) {
    sections.push({
      type: 'subscription',
      title: t('groupMonitor.subscriptionType'),
      items: subscriptionMonitorItems.value
    })
  }
  return sections
})

function onPlatformFilterChange(value: string | number | boolean | null) {
  selectedPlatform.value = String(value ?? '')
}

function resolveGroupType(item: PublicGroupMonitorItem): 'public' | 'subscription' {
  return item.group_type === 'subscription' ? 'subscription' : 'public'
}

function statusLabel(status: string): string {
  if (status === 'normal') return t('groupMonitor.status.normal')
  if (status === 'abnormal') return t('groupMonitor.status.abnormal')
  return t('groupMonitor.status.unknown')
}

function statusBadgeClass(status: string): string {
  if (status === 'normal') {
    return 'bg-emerald-100 text-emerald-700 ring-1 ring-emerald-200 dark:bg-emerald-900/40 dark:text-emerald-300 dark:ring-emerald-800'
  }
  if (status === 'abnormal') {
    return 'bg-rose-100 text-rose-700 ring-1 ring-rose-200 dark:bg-rose-900/40 dark:text-rose-300 dark:ring-rose-800'
  }
  return 'bg-gray-100 text-gray-700 ring-1 ring-gray-200 dark:bg-dark-700 dark:text-gray-300 dark:ring-dark-600'
}

function statusDotClass(status: string): string {
  if (status === 'normal') return 'bg-emerald-500'
  if (status === 'abnormal') return 'bg-rose-500'
  return 'bg-gray-400'
}

function statusAccentClass(status: string): string {
  if (status === 'normal') return 'bg-gradient-to-r from-emerald-400 via-emerald-500 to-teal-500'
  if (status === 'abnormal') return 'bg-gradient-to-r from-rose-400 via-rose-500 to-orange-500'
  return 'bg-gradient-to-r from-gray-300 via-gray-400 to-gray-500 dark:from-dark-500 dark:via-dark-400 dark:to-dark-300'
}

function statusCardClass(status: string): string {
  if (status === 'normal') {
    return 'border-emerald-200/80 bg-gradient-to-br from-emerald-50 to-white dark:border-emerald-900/50 dark:from-emerald-950/20 dark:to-dark-900'
  }
  if (status === 'abnormal') {
    return 'border-rose-200/80 bg-gradient-to-br from-rose-50 to-white dark:border-rose-900/50 dark:from-rose-950/20 dark:to-dark-900'
  }
  return 'border-gray-200 bg-gradient-to-br from-gray-50 to-white dark:border-dark-700 dark:from-dark-800 dark:to-dark-900'
}

function platformBadgeClass(platform: string): string {
  const value = platform.trim().toLowerCase()
  if (value === 'openai') {
    return 'border-sky-200 bg-sky-50 text-sky-700 dark:border-sky-900/70 dark:bg-sky-900/20 dark:text-sky-300'
  }
  if (value === 'anthropic') {
    return 'border-amber-200 bg-amber-50 text-amber-700 dark:border-amber-900/70 dark:bg-amber-900/20 dark:text-amber-300'
  }
  if (value === 'gemini') {
    return 'border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-900/70 dark:bg-emerald-900/20 dark:text-emerald-300'
  }
  return 'border-gray-200 bg-gray-50 text-gray-700 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-300'
}

async function loadData() {
  loading.value = true
  errorMessage.value = ''

  try {
    monitorData.value = await userGroupsAPI.getPublicGroupMonitor()
  } catch (err: any) {
    errorMessage.value = err?.message || t('groupMonitor.loadFailed')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void loadData()
})
</script>
