<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="card p-6">
        <div class="space-y-6">
          <!-- 标题和按钮行 -->
          <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
            <div class="space-y-1">
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t('admin.paymentOrders.title') }}
              </h2>
              <p class="text-sm text-gray-500 dark:text-dark-400">
                {{ t('admin.paymentOrders.description') }}
              </p>
            </div>

            <!-- 按钮组 -->
            <div class="flex flex-wrap gap-3">
              <button class="btn btn-secondary min-w-[100px] px-5 py-2.5" :disabled="loading" @click="applyFilters">
                {{ t('common.filter') }}
              </button>
              <button class="btn btn-secondary min-w-[100px] px-5 py-2.5" :disabled="loading" @click="resetFilters">
                {{ t('common.reset') }}
              </button>
              <button class="btn btn-primary min-w-[120px] px-5 py-2.5" :disabled="exporting" @click="exportRecords">
                {{ exporting ? t('common.loading') : t('admin.paymentOrders.export') }}
              </button>
            </div>
          </div>

          <!-- 筛选条件网格 -->
          <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5">
            <div>
              <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">
                {{ t('admin.paymentOrders.method') }}
              </label>
              <Select
                v-model="filters.method"
                :options="methodOptions"
                :placeholder="t('common.all')"
              />
            </div>

            <div>
              <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">
                {{ t('admin.paymentOrders.userEmail') }}
              </label>
              <input
                v-model.trim="filters.user"
                type="text"
                class="w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm outline-none transition focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-dark-700 dark:bg-dark-900 dark:text-white"
                :placeholder="t('admin.paymentOrders.userPlaceholder')"
                @keydown.enter.prevent="applyFilters"
              />
            </div>

            <div>
              <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">
                {{ t('admin.paymentOrders.status') }}
              </label>
              <Select v-model="filters.status" :options="statusOptions" :placeholder="t('common.all')" />
            </div>

            <div>
              <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">
                {{ t('admin.paymentOrders.from') }}
              </label>
              <input
                v-model="filters.from"
                type="datetime-local"
                class="w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm outline-none transition focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-dark-700 dark:bg-dark-900 dark:text-white"
              />
            </div>

            <div>
              <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-dark-200">
                {{ t('admin.paymentOrders.to') }}
              </label>
              <input
                v-model="filters.to"
                type="datetime-local"
                class="w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm outline-none transition focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-dark-700 dark:bg-dark-900 dark:text-white"
              />
            </div>
          </div>
        </div>
      </div>

      <div class="card p-6">
        <div v-if="loading" class="flex items-center justify-center py-10">
          <LoadingSpinner />
        </div>

        <template v-else>
          <DataTable :columns="columns" :data="items" :loading="loading">
            <template #cell-order_no="{ row }">
              <span class="font-mono text-sm font-medium text-gray-900 dark:text-white">{{ row.order_no }}</span>
            </template>

            <template #cell-order_type="{ row }">
              <span class="text-sm text-gray-700 dark:text-dark-300">{{ orderTypeLabel(row.order_type) }}</span>
            </template>

            <template #cell-provider="{ row }">
              <span class="text-sm text-gray-900 dark:text-white">
                {{ shouldShowChannel(row.order_type) ? channelLabel(row.channel || row.provider) : '-' }}
              </span>
            </template>

            <template #cell-total_usd="{ row }">
              <span class="text-base font-bold text-gray-900 dark:text-white">${{ row.total_usd.toFixed(2) }}</span>
            </template>

            <template #cell-amount_cny="{ row }">
              <span class="text-base font-bold text-gray-900 dark:text-white">
                {{ shouldShowPayAmount(row.order_type) ? `¥${row.amount_cny.toFixed(2)}` : '-' }}
              </span>
            </template>

            <template #cell-status="{ row }">
              <span class="text-sm text-gray-700 dark:text-dark-300">{{ statusLabel(row.status) }}</span>
            </template>

            <template #empty>
              <EmptyState :message="t('common.noData')" />
            </template>
          </DataTable>

          <Pagination
            v-if="pagination.total > 0"
            :page="pagination.page"
            :total="pagination.total"
            :page-size="pagination.page_size"
            @update:page="handlePageChange"
            @update:pageSize="handlePageSizeChange"
          />
        </template>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { saveAs } from 'file-saver'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Select from '@/components/common/Select.vue'
import { adminAPI } from '@/api/admin'
import type { Column } from '@/components/common/types'
import type { AdminPaymentOrder, AdminPaymentMethod } from '@/api/admin/paymentOrders'

const { t } = useI18n()

const loading = ref(false)
const exporting = ref(false)

const filters = reactive<{ method: AdminPaymentMethod | ''; user: string; status: string; from: string; to: string }>({
  method: '',
  user: '',
  status: '',
  from: '',
  to: ''
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const items = ref<AdminPaymentOrder[]>([])

const methodOptions = computed(() => [
  { label: t('common.all'), value: '' },
  { label: t('payment.alipay'), value: 'alipay' },
  { label: t('payment.wechat'), value: 'wechat' }
])

const statusOptions = computed(() => [
  { label: t('common.all'), value: '' },
  { label: t('payment.statusPending'), value: 'pending' },
  { label: t('payment.statusPaid'), value: 'paid' },
  { label: t('payment.statusFailed'), value: 'failed' },
  { label: t('payment.statusExpired'), value: 'expired' },
  { label: t('payment.statusCancelled'), value: 'cancelled' }
])

const columns = computed<Column[]>(() => [
  { key: 'order_no', label: t('payment.orderNo') },
  { key: 'user_email', label: t('admin.paymentOrders.userEmail') },
  { key: 'order_type', label: t('payment.orderType') },
  { key: 'provider', label: t('payment.channel') },
  { key: 'total_usd', label: t('payment.creditsAmount') },
  { key: 'amount_cny', label: t('payment.payAmountCny') },
  { key: 'status', label: t('payment.status') },
  { key: 'created_at', label: t('common.createdAt') }
])

function channelLabel(channel: string): string {
  // 根据实际支付渠道返回标签
  if (channel === 'alipay') return t('payment.alipay')
  if (channel === 'wechat' || channel === 'wxpay') return t('payment.wechat')
  if (channel === 'admin') return t('payment.adminRecharge')
  if (channel === 'activity') return t('payment.activityRecharge')
  return channel
}

function shouldShowChannel(orderType?: string): boolean {
  const value = String(orderType || '').toLowerCase()
  return value === '' || value === 'online_recharge'
}

function shouldShowPayAmount(orderType?: string): boolean {
  return shouldShowChannel(orderType)
}

function orderTypeLabel(orderType?: string): string {
  const value = String(orderType || '').toLowerCase()
  if (value === 'admin_recharge') return t('payment.orderTypeAdmin')
  if (value === 'activity_recharge') return t('payment.orderTypeActivity')
  return t('payment.orderTypeOnline')
}

function statusLabel(status: string): string {
  const normalized = String(status || '').toLowerCase()
  switch (normalized) {
    case 'pending':
      return t('payment.statusPending')
    case 'paid':
      return t('payment.statusPaid')
    case 'failed':
      return t('payment.statusFailed')
    case 'expired':
      return t('payment.statusExpired')
    case 'cancelled':
    case 'canceled':
      return t('payment.statusCancelled')
    default:
      return status
  }
}

async function load() {
  loading.value = true
  try {
    const resp = await adminAPI.paymentOrders.list(pagination.page, pagination.page_size, {
      method: filters.method || '',
      user: filters.user || '',
      status: filters.status || '',
      from: filters.from ? new Date(filters.from).toISOString() : '',
      to: filters.to ? new Date(filters.to).toISOString() : ''
    })
    items.value = resp.items
    pagination.total = resp.total
  } finally {
    loading.value = false
  }
}

function applyFilters() {
  pagination.page = 1
  load()
}

function resetFilters() {
  filters.method = ''
  filters.user = ''
  filters.status = ''
  filters.from = ''
  filters.to = ''
  applyFilters()
}

function handlePageChange(page: number) {
  pagination.page = page
  load()
}

function handlePageSizeChange(pageSize: number) {
  pagination.page_size = pageSize
  pagination.page = 1
  load()
}

async function exportRecords() {
  exporting.value = true
  try {
    const blob = await adminAPI.paymentOrders.exportRecords({
      method: filters.method || '',
      user: filters.user || '',
      status: filters.status || '',
      from: filters.from ? new Date(filters.from).toISOString() : '',
      to: filters.to ? new Date(filters.to).toISOString() : ''
    })
    const now = new Date()
    const stamp = `${now.getFullYear()}${String(now.getMonth() + 1).padStart(2, '0')}${String(now.getDate()).padStart(2, '0')}_${String(
      now.getHours()
    ).padStart(2, '0')}${String(now.getMinutes()).padStart(2, '0')}${String(now.getSeconds()).padStart(2, '0')}`
    saveAs(blob, `recharge_records_${stamp}.csv`)
  } finally {
    exporting.value = false
  }
}

onMounted(() => {
  load()
})
</script>
