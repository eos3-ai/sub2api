<template>
  <AppLayout>
    <div class="space-y-6">
      <section class="card p-4 md:p-5">
        <div class="mb-4 flex flex-wrap items-start justify-between gap-3">
          <div>
            <h2 class="text-base font-bold text-gray-900 dark:text-white">
              {{ t('groupMonitor.title') }}
            </h2>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              {{ t('groupMonitor.description') }}
            </p>
          </div>
          <button class="btn btn-secondary btn-sm" :disabled="loading" @click="loadData">
            {{ t('groupMonitor.refresh') }}
          </button>
        </div>

        <div v-if="errorMessage" class="mb-4 rounded-lg bg-red-50 px-3 py-2 text-xs text-red-600 dark:bg-red-900/20 dark:text-red-400">
          {{ errorMessage }}
        </div>

        <div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
          <div class="rounded-lg bg-gray-50 px-3 py-2 dark:bg-dark-800">
            <div class="text-[11px] text-gray-500 dark:text-gray-400">{{ t('groupMonitor.summary.publicGroups') }}</div>
            <div class="mt-1 text-sm font-semibold text-gray-900 dark:text-gray-100">{{ summary.publicGroups }}</div>
          </div>
          <div class="rounded-lg bg-emerald-50 px-3 py-2 dark:bg-emerald-900/20">
            <div class="text-[11px] text-emerald-600/80 dark:text-emerald-300">{{ t('groupMonitor.summary.healthyGroups') }}</div>
            <div class="mt-1 text-sm font-semibold text-emerald-700 dark:text-emerald-200">{{ summary.healthyGroups }}</div>
          </div>
          <div class="rounded-lg bg-rose-50 px-3 py-2 dark:bg-rose-900/20">
            <div class="text-[11px] text-rose-600/80 dark:text-rose-300">{{ t('groupMonitor.summary.riskyGroups') }}</div>
            <div class="mt-1 text-sm font-semibold text-rose-700 dark:text-rose-200">{{ summary.riskyGroups }}</div>
          </div>
        </div>
      </section>

      <section class="card p-4 md:p-5">
        <div class="mb-3 text-sm font-semibold text-gray-900 dark:text-gray-100">
          {{ t('groupMonitor.impactedKeys') }}
        </div>
        <div v-if="loading" class="py-4 text-center text-sm text-gray-500 dark:text-gray-400">
          {{ t('groupMonitor.loading') }}
        </div>
        <div v-else-if="impactedKeys.length === 0" class="rounded-lg bg-gray-50 px-3 py-4 text-sm text-gray-500 dark:bg-dark-800 dark:text-gray-400">
          {{ t('groupMonitor.noImpactedKeys') }}
        </div>
        <div v-else class="overflow-x-auto">
          <table class="min-w-full text-sm">
            <thead>
              <tr class="border-b border-gray-200 text-left text-xs text-gray-500 dark:border-dark-700 dark:text-gray-400">
                <th class="px-3 py-2">{{ t('groupMonitor.keyName') }}</th>
                <th class="px-3 py-2">{{ t('groupMonitor.currentGroup') }}</th>
                <th class="px-3 py-2">{{ t('groupMonitor.currentStatus') }}</th>
                <th class="px-3 py-2">{{ t('groupMonitor.switchTo') }}</th>
                <th class="px-3 py-2">{{ t('groupMonitor.action') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="key in impactedKeys" :key="key.id" class="border-b border-gray-100 dark:border-dark-800">
                <td class="px-3 py-2 font-medium text-gray-900 dark:text-gray-100">{{ key.name }}</td>
                <td class="px-3 py-2 text-gray-600 dark:text-gray-300">
                  {{ currentGroupName(key) }}
                </td>
                <td class="px-3 py-2">
                  <span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium" :class="statusBadgeClass(groupStatusForKey(key))">
                    {{ statusLabel(groupStatusForKey(key)) }}
                  </span>
                </td>
                <td class="px-3 py-2">
                  <select
                    v-model.number="switchTargetByKey[key.id]"
                    class="w-full rounded-md border border-gray-300 bg-white px-2 py-1 text-xs text-gray-700 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-200"
                  >
                    <option :value="0">{{ t('groupMonitor.selectTargetGroup') }}</option>
                    <option
                      v-for="candidate in switchCandidates(key)"
                      :key="candidate.group_id"
                      :value="candidate.group_id"
                    >
                      {{ candidate.group_name }} ({{ candidate.platform }})
                    </option>
                  </select>
                </td>
                <td class="px-3 py-2">
                  <button
                    class="btn btn-primary btn-xs"
                    :disabled="switchingKeyId === key.id || !switchTargetByKey[key.id]"
                    @click="switchGroup(key)"
                  >
                    {{ switchingKeyId === key.id ? t('common.processing') : t('groupMonitor.switchNow') }}
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section class="card p-4 md:p-5">
        <div class="mb-4 text-sm font-semibold text-gray-900 dark:text-gray-100">
          {{ t('groupMonitor.groupDetails') }}
        </div>

        <div v-if="loading" class="py-4 text-center text-sm text-gray-500 dark:text-gray-400">
          {{ t('groupMonitor.loading') }}
        </div>
        <div v-else-if="monitorItems.length === 0" class="rounded-lg bg-gray-50 px-3 py-4 text-sm text-gray-500 dark:bg-dark-800 dark:text-gray-400">
          {{ t('groupMonitor.empty') }}
        </div>

        <div v-else class="space-y-4">
          <article
            v-for="group in monitorItems"
            :key="group.group_id"
            class="rounded-xl border border-gray-200 p-4 dark:border-dark-700"
          >
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div>
                <div class="text-sm font-semibold text-gray-900 dark:text-gray-100">
                  {{ group.group_name }}
                </div>
                <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  {{ group.platform }}
                </div>
              </div>
              <span class="inline-flex items-center rounded-full px-2.5 py-1 text-xs font-medium" :class="statusBadgeClass(group.current_status)">
                {{ statusLabel(group.current_status) }}
              </span>
            </div>

            <div class="mt-3 grid grid-cols-1 gap-2 sm:grid-cols-3">
              <div class="rounded-lg bg-gray-50 px-3 py-2 dark:bg-dark-800">
                <div class="text-[11px] text-gray-500 dark:text-gray-400">{{ t('groupMonitor.totalRequests1h') }}</div>
                <div class="mt-1 text-sm font-semibold text-gray-900 dark:text-gray-100">{{ formatNumber(group.total_requests_1h) }}</div>
              </div>
              <div class="rounded-lg bg-emerald-50 px-3 py-2 dark:bg-emerald-900/20">
                <div class="text-[11px] text-emerald-600/80 dark:text-emerald-300">{{ t('groupMonitor.successRequests1h') }}</div>
                <div class="mt-1 text-sm font-semibold text-emerald-700 dark:text-emerald-200">{{ formatNumber(group.success_requests_1h) }}</div>
              </div>
              <div class="rounded-lg bg-rose-50 px-3 py-2 dark:bg-rose-900/20">
                <div class="text-[11px] text-rose-600/80 dark:text-rose-300">{{ t('groupMonitor.failureRequests1h') }}</div>
                <div class="mt-1 text-sm font-semibold text-rose-700 dark:text-rose-200">{{ formatNumber(group.failure_requests_1h) }}</div>
              </div>
            </div>

            <div class="mt-3">
              <div class="mb-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                {{ t('groupMonitor.samples') }}
              </div>
              <div class="overflow-x-auto pb-1">
                <div class="inline-flex flex-nowrap items-center gap-1.5">
                  <div
                    v-for="(sample, idx) in sampleSlots(group.samples)"
                    :key="`${group.group_id}-${idx}`"
                    class="relative"
                  >
                    <div
                      v-if="sample"
                      class="h-4 w-4 rounded-sm border"
                      :class="sampleStatusClass(sample.status)"
                      @mouseenter="onSampleMouseEnter($event, sample)"
                      @mousemove="onSampleMouseMove($event, sample)"
                      @mouseleave="onSampleMouseLeave"
                    />
                    <div
                      v-else
                      class="h-4 w-4 rounded-sm border border-gray-300 bg-gray-200 dark:border-dark-600 dark:bg-dark-700"
                    />
                  </div>
                </div>
              </div>
            </div>
          </article>
        </div>
      </section>

      <Teleport to="body">
        <div
          v-if="hoveredSample"
          class="pointer-events-none fixed z-[9999] w-56 rounded-lg border border-gray-200 bg-white p-2 text-[11px] leading-5 text-gray-700 shadow-2xl dark:border-dark-600 dark:bg-dark-900 dark:text-gray-200"
          :class="sampleTooltipPlacement === 'top' ? '-translate-x-1/2 -translate-y-full' : '-translate-x-1/2'"
          :style="sampleTooltipStyle"
        >
          <div>{{ t('groupMonitor.sample.requestTime') }}: {{ formatSampleTime(hoveredSample.sample.started_at) }}</div>
          <div>{{ t('groupMonitor.sample.status') }}: {{ sampleStatusLabel(hoveredSample.sample.status) }}</div>
          <div>{{ t('groupMonitor.sample.model') }}: {{ formatSampleModel(hoveredSample.sample.model) }}</div>
          <div>{{ t('groupMonitor.sample.latency') }}: {{ formatLatency(hoveredSample.sample.latency_ms) }}</div>
        </div>
      </Teleport>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import type { ApiKey } from '@/types'
import { keysAPI } from '@/api/keys'
import {
  userGroupsAPI,
  type PublicGroupMonitorItem,
  type PublicGroupMonitorResponse,
  type PublicGroupMonitorSample
} from '@/api/groups'
import { useAppStore } from '@/stores'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const errorMessage = ref('')
const monitorData = ref<PublicGroupMonitorResponse | null>(null)
const apiKeys = ref<ApiKey[]>([])
const switchingKeyId = ref<number | null>(null)
const switchTargetByKey = reactive<Record<number, number>>({})

const displaySampleSize = computed(() => monitorData.value?.sample_size || 30)
const monitorItems = computed<PublicGroupMonitorItem[]>(() => monitorData.value?.items || [])
const monitorMap = computed(() => {
  const map = new Map<number, PublicGroupMonitorItem>()
  for (const item of monitorItems.value) {
    map.set(item.group_id, item)
  }
  return map
})

const summary = computed(() => {
  const items = monitorItems.value
  return {
    publicGroups: items.length,
    healthyGroups: items.filter((item) => item.current_status === 'normal').length,
    riskyGroups: items.filter((item) => item.current_status === 'abnormal').length
  }
})

const impactedKeys = computed(() => {
  return apiKeys.value.filter((key) => {
    if (!key.group_id) return false
    const status = groupStatusForKey(key)
    return status !== 'normal'
  })
})

const sampleTooltipHalfWidth = 112
const sampleTooltipViewportPadding = 8
const sampleTooltipEstimatedHeight = 108

interface SampleTooltipState {
  sample: PublicGroupMonitorSample
  x: number
  top: number
  bottom: number
}

const hoveredSample = ref<SampleTooltipState | null>(null)
const sampleTooltipPlacement = computed<'top' | 'bottom'>(() => {
  const state = hoveredSample.value
  if (!state) return 'top'
  const topCandidate = state.top - 8 - sampleTooltipEstimatedHeight
  if (topCandidate >= sampleTooltipViewportPadding) return 'top'
  return 'bottom'
})
const sampleTooltipStyle = computed(() => {
  const state = hoveredSample.value
  if (!state) return {}

  const viewportWidth = typeof window !== 'undefined' ? window.innerWidth : 0
  const viewportHeight = typeof window !== 'undefined' ? window.innerHeight : 0
  const minLeft = sampleTooltipHalfWidth + sampleTooltipViewportPadding
  const maxLeft = Math.max(minLeft, viewportWidth - sampleTooltipHalfWidth - sampleTooltipViewportPadding)
  const left = viewportWidth > 0 ? Math.min(Math.max(state.x, minLeft), maxLeft) : state.x

  if (sampleTooltipPlacement.value === 'bottom') {
    const rawTop = state.bottom + 8
    const maxTop = Math.max(sampleTooltipViewportPadding, viewportHeight - sampleTooltipEstimatedHeight - sampleTooltipViewportPadding)
    const top = viewportHeight > 0 ? Math.min(rawTop, maxTop) : rawTop
    return { left: `${Math.round(left)}px`, top: `${Math.round(top)}px` }
  }

  return { left: `${Math.round(left)}px`, top: `${Math.round(state.top - 8)}px` }
})

function sampleStatusClass(status: string): string {
  if (status === 'success') {
    return 'border-emerald-400 bg-emerald-500/85 dark:border-emerald-500 dark:bg-emerald-500/75'
  }
  return 'border-rose-400 bg-rose-500/85 dark:border-rose-500 dark:bg-rose-500/75'
}

function sampleStatusLabel(status: string): string {
  return status === 'success'
    ? t('groupMonitor.sampleStatus.success')
    : t('groupMonitor.sampleStatus.failed')
}

function statusLabel(status: string): string {
  if (status === 'normal') return t('groupMonitor.status.normal')
  if (status === 'abnormal') return t('groupMonitor.status.abnormal')
  return t('groupMonitor.status.unknown')
}

function statusBadgeClass(status: string): string {
  if (status === 'normal') {
    return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
  }
  if (status === 'abnormal') {
    return 'bg-rose-100 text-rose-700 dark:bg-rose-900/30 dark:text-rose-300'
  }
  return 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-gray-300'
}

function groupStatusForKey(key: ApiKey): string {
  if (!key.group_id) return 'unknown'
  return monitorMap.value.get(key.group_id)?.current_status || 'unknown'
}

function currentGroupName(key: ApiKey): string {
  if (key.group?.name) return key.group.name
  if (!key.group_id) return '-'
  return monitorMap.value.get(key.group_id)?.group_name || `Group #${key.group_id}`
}

function currentGroupPlatform(key: ApiKey): string {
  if (key.group?.platform) return key.group.platform
  if (!key.group_id) return ''
  return monitorMap.value.get(key.group_id)?.platform || ''
}

function switchCandidates(key: ApiKey): PublicGroupMonitorItem[] {
  const platform = currentGroupPlatform(key)
  if (!platform) return []
  return monitorItems.value.filter((group) => {
    return group.platform === platform && group.current_status === 'normal' && group.group_id !== key.group_id
  })
}

async function switchGroup(key: ApiKey) {
  const targetGroupID = switchTargetByKey[key.id]
  if (!targetGroupID || targetGroupID <= 0) {
    appStore.showError(t('groupMonitor.selectTargetGroup'))
    return
  }

  switchingKeyId.value = key.id
  try {
    await keysAPI.update(key.id, { group_id: targetGroupID })
    appStore.showSuccess(t('groupMonitor.switchSuccess'))
    await loadData()
  } catch (error) {
    appStore.showError(t('groupMonitor.switchFailed'))
  } finally {
    switchingKeyId.value = null
  }
}

function sampleSlots(samples?: PublicGroupMonitorSample[] | null): Array<PublicGroupMonitorSample | null> {
  const size = displaySampleSize.value
  const out: Array<PublicGroupMonitorSample | null> = (samples || []).slice(0, size)
  while (out.length < size) out.push(null)
  return out
}

function formatNumber(v: number): string {
  if (!Number.isFinite(v) || v < 0) return '-'
  return Math.round(v).toLocaleString()
}

function formatSampleTime(v?: string | null): string {
  if (!v) return '-'
  const d = new Date(v)
  if (Number.isNaN(d.getTime())) return v
  return d.toLocaleString()
}

function formatSampleModel(v?: string | null): string {
  const model = String(v || '').trim()
  return model || '-'
}

function formatLatency(v: number): string {
  if (!Number.isFinite(v) || v < 0) return '-'
  return `${Math.round(v)} ms`
}

function updateHoveredSamplePosition(event: MouseEvent, sample: PublicGroupMonitorSample) {
  const target = event.currentTarget as HTMLElement | null
  if (!target) return
  const rect = target.getBoundingClientRect()
  hoveredSample.value = {
    sample,
    x: rect.left + rect.width / 2,
    top: rect.top,
    bottom: rect.bottom
  }
}

function onSampleMouseEnter(event: MouseEvent, sample: PublicGroupMonitorSample) {
  updateHoveredSamplePosition(event, sample)
}

function onSampleMouseMove(event: MouseEvent, sample: PublicGroupMonitorSample) {
  updateHoveredSamplePosition(event, sample)
}

function onSampleMouseLeave() {
  hoveredSample.value = null
}

async function loadAllKeys(): Promise<ApiKey[]> {
  const pageSize = 100
  let page = 1
  const all: ApiKey[] = []

  while (true) {
    const resp = await keysAPI.list(page, pageSize)
    all.push(...(resp.items || []))
    if (page >= (resp.pages || 1)) break
    page += 1
  }
  return all
}

async function loadData() {
  loading.value = true
  errorMessage.value = ''
  hoveredSample.value = null

  try {
    const [monitorResp, keys] = await Promise.all([
      userGroupsAPI.getPublicGroupMonitor({ sample_size: 30, bucket_seconds: 15 }),
      loadAllKeys()
    ])
    monitorData.value = monitorResp
    apiKeys.value = keys
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
