<template>
  <AppLayout>
    <div class="space-y-6">
      <UsageStatsCards :stats="usageStats" />
      <!-- Charts Section -->
      <div class="space-y-4">
        <div class="card p-4">
          <div class="flex items-center gap-4">
            <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.dashboard.granularity') }}:</span>
            <div class="w-28">
              <Select v-model="granularity" :options="granularityOptions" @change="loadChartData" />
            </div>
          </div>
        </div>
        <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
          <ModelDistributionChart :model-stats="modelStats" :loading="chartsLoading" />
          <TokenUsageTrend :trend-data="trendData" :loading="chartsLoading" />
        </div>
      </div>
      <UsageFilters v-model="filters" v-model:startDate="startDate" v-model:endDate="endDate" :exporting="exporting" @change="applyFilters" @refresh="refreshData" @reset="resetFilters" @cleanup="openCleanupDialog" @export="exportToCSV" />
      <UsageTable :data="usageLogs" :loading="loading" :columns="visibleColumns" />
      <Pagination v-if="pagination.total > 0" :page="pagination.page" :total="pagination.total" :page-size="pagination.page_size" @update:page="handlePageChange" @update:pageSize="handlePageSizeChange" />
    </div>
  </AppLayout>
  <UsageCleanupDialog
    :show="cleanupDialogVisible"
    :filters="filters"
    :start-date="startDate"
    :end-date="endDate"
    @close="cleanupDialogVisible = false"
  />
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { saveAs } from 'file-saver'
import { useAppStore } from '@/stores/app'; import { adminAPI } from '@/api/admin'; import { adminUsageAPI } from '@/api/admin/usage'
import AppLayout from '@/components/layout/AppLayout.vue'; import Pagination from '@/components/common/Pagination.vue'; import Select from '@/components/common/Select.vue'
import UsageStatsCards from '@/components/admin/usage/UsageStatsCards.vue'; import UsageFilters from '@/components/admin/usage/UsageFilters.vue'
import UsageTable from '@/components/admin/usage/UsageTable.vue'
import UsageCleanupDialog from '@/components/admin/usage/UsageCleanupDialog.vue'
import ModelDistributionChart from '@/components/charts/ModelDistributionChart.vue'; import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'
import type { AdminUsageLog, TrendDataPoint, ModelStat } from '@/types'; import type { AdminUsageStatsResponse, AdminUsageQueryParams } from '@/api/admin/usage'
import type { Column } from '@/components/common/types'

const { t } = useI18n()
const appStore = useAppStore()
const usageStats = ref<AdminUsageStatsResponse | null>(null); const usageLogs = ref<AdminUsageLog[]>([]); const loading = ref(false); const exporting = ref(false)
const trendData = ref<TrendDataPoint[]>([]); const modelStats = ref<ModelStat[]>([]); const chartsLoading = ref(false); const granularity = ref<'day' | 'hour'>('day')
let abortController: AbortController | null = null; let exportAbortController: AbortController | null = null
const cleanupDialogVisible = ref(false)

const granularityOptions = computed(() => [{ value: 'day', label: t('admin.dashboard.day') }, { value: 'hour', label: t('admin.dashboard.hour') }])
const visibleColumns = computed<Column[]>(() => [
  { key: 'user', label: t('admin.usage.user'), sortable: false },
  { key: 'api_key', label: t('usage.apiKeyFilter'), sortable: false },
  { key: 'account', label: t('admin.usage.account'), sortable: false },
  { key: 'model', label: t('usage.model'), sortable: true },
  { key: 'reasoning_effort', label: t('usage.reasoningEffort'), sortable: false },
  { key: 'endpoint', label: t('usage.endpoint'), sortable: false },
  { key: 'group', label: t('admin.usage.group'), sortable: false },
  { key: 'stream', label: t('usage.type'), sortable: false },
  { key: 'tokens', label: t('usage.tokens'), sortable: false },
  { key: 'cost', label: t('usage.cost'), sortable: false },
  { key: 'first_token', label: t('usage.firstToken'), sortable: false },
  { key: 'duration', label: t('usage.duration'), sortable: false },
  { key: 'created_at', label: t('usage.time'), sortable: true },
  { key: 'user_agent', label: t('usage.userAgent'), sortable: false },
  { key: 'ip_address', label: t('admin.usage.ipAddress'), sortable: false }
])
// Use local timezone to avoid UTC timezone issues
const formatLD = (d: Date) => {
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}
const now = new Date(); const weekAgo = new Date(); weekAgo.setDate(weekAgo.getDate() - 6)
const startDate = ref(formatLD(weekAgo)); const endDate = ref(formatLD(now))
const filters = ref<AdminUsageQueryParams>({ user_id: undefined, model: undefined, group_id: undefined, billing_type: null, start_date: startDate.value, end_date: endDate.value })
const pagination = reactive({ page: 1, page_size: 20, total: 0 })

const loadLogs = async () => {
  abortController?.abort(); const c = new AbortController(); abortController = c; loading.value = true
  try {
    const res = await adminAPI.usage.list({ page: pagination.page, page_size: pagination.page_size, ...filters.value }, { signal: c.signal })
    if(!c.signal.aborted) { usageLogs.value = res.items; pagination.total = res.total }
  } catch (error: any) { if(error?.name !== 'AbortError') console.error('Failed to load usage logs:', error) } finally { if(abortController === c) loading.value = false }
}
const loadStats = async () => {
  try {
    const params = {
      ...filters.value,
      billing_type: filters.value.billing_type ?? undefined
    }
    const s = await adminAPI.usage.getStats(params)
    usageStats.value = s
  } catch (error) {
    console.error('Failed to load usage stats:', error)
  }
}
const loadChartData = async () => {
  chartsLoading.value = true
  try {
    const params = { start_date: filters.value.start_date || startDate.value, end_date: filters.value.end_date || endDate.value, granularity: granularity.value, user_id: filters.value.user_id, model: filters.value.model, api_key_id: filters.value.api_key_id, account_id: filters.value.account_id, group_id: filters.value.group_id, stream: filters.value.stream, billing_type: filters.value.billing_type }
    const [trendRes, modelRes] = await Promise.all([adminAPI.dashboard.getUsageTrend(params), adminAPI.dashboard.getModelStats({ start_date: params.start_date, end_date: params.end_date, user_id: params.user_id, model: params.model, api_key_id: params.api_key_id, account_id: params.account_id, group_id: params.group_id, stream: params.stream, billing_type: params.billing_type })])
    trendData.value = trendRes.trend || []; modelStats.value = modelRes.models || []
  } catch (error) { console.error('Failed to load chart data:', error) } finally { chartsLoading.value = false }
}
const applyFilters = () => { pagination.page = 1; loadLogs(); loadStats(); loadChartData() }
const refreshData = () => { loadLogs(); loadStats(); loadChartData() }
const resetFilters = () => { startDate.value = formatLD(weekAgo); endDate.value = formatLD(now); filters.value = { start_date: startDate.value, end_date: endDate.value, billing_type: null }; granularity.value = 'day'; applyFilters() }
const handlePageChange = (p: number) => { pagination.page = p; loadLogs() }
const handlePageSizeChange = (s: number) => { pagination.page_size = s; pagination.page = 1; loadLogs() }
const openCleanupDialog = () => { cleanupDialogVisible.value = true }

const exportToCSV = async () => {
  if (exporting.value) return; exporting.value = true
  const c = new AbortController(); exportAbortController = c
  const exportFilters = { ...filters.value }
  try {
    const blob = await adminUsageAPI.exportCsv(exportFilters, { signal: c.signal })
    if (!c.signal.aborted) {
      const from = exportFilters.start_date || startDate.value
      const to = exportFilters.end_date || endDate.value
      saveAs(blob, `usage_${from}_to_${to}.csv`)
      appStore.showSuccess(t('usage.exportSuccess'))
    }
  } catch (error: any) { if (error?.name !== 'AbortError' && error?.code !== 'ERR_CANCELED') { console.error('Failed to export:', error); appStore.showError(t('usage.exportFailed')) } }
  finally { if(exportAbortController === c) { exportAbortController = null; exporting.value = false } }
}

onMounted(() => { loadLogs(); loadStats(); loadChartData() })
onUnmounted(() => { abortController?.abort(); exportAbortController?.abort() })
</script>
