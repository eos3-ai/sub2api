<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="card p-6">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div class="space-y-1">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('admin.paymentOrders.title') }}
            </h2>
            <p class="text-sm text-gray-500 dark:text-dark-400">
              {{ t('admin.paymentOrders.description') }}
            </p>
          </div>

          <div class="flex flex-col gap-3 sm:flex-row sm:items-end">
            <div class="min-w-[180px]">
              <label class="mb-1 block text-xs font-semibold text-gray-600 dark:text-dark-300">
                {{ t('admin.paymentOrders.method') }}
              </label>
              <Select
                v-model="filters.method"
                :options="methodOptions"
                :placeholder="t('common.all')"
              />
            </div>

            <div class="min-w-[220px]">
              <label class="mb-1 block text-xs font-semibold text-gray-600 dark:text-dark-300">
                {{ t('admin.paymentOrders.user') }}
              </label>
              <input
                v-model.trim="filters.user"
                type="text"
                class="w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm outline-none transition focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-dark-700 dark:bg-dark-900 dark:text-white"
                :placeholder="t('admin.paymentOrders.userPlaceholder')"
                @keydown.enter.prevent="applyFilters"
              />
            </div>

            <div class="flex gap-2">
              <button class="btn btn-secondary" :disabled="loading" @click="applyFilters">
                {{ t('common.filter') }}
              </button>
              <button class="btn btn-secondary" :disabled="loading" @click="resetFilters">
                {{ t('common.reset') }}
              </button>
              <button class="btn btn-primary" :disabled="exporting" @click="exportRecords">
                {{ exporting ? t('common.loading') : t('admin.paymentOrders.export') }}
              </button>
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
              <span class="font-mono text-xs text-gray-900 dark:text-white">{{ row.order_no }}</span>
            </template>

            <template #cell-order_type="{ row }">
              <span class="text-sm text-gray-700 dark:text-dark-300">{{ orderTypeLabel(row.order_type) }}</span>
            </template>

            <template #cell-provider="{ row }">
              <span class="text-sm text-gray-900 dark:text-white">{{ providerLabel(row.provider) }}</span>
            </template>

            <template #cell-total_usd="{ row }">
              <span class="text-sm font-semibold text-gray-900 dark:text-white">${{ row.total_usd.toFixed(2) }}</span>
            </template>

            <template #cell-amount_cny="{ row }">
              <span class="text-sm text-gray-900 dark:text-white">¥{{ row.amount_cny.toFixed(2) }}</span>
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
import type { AdminPaymentOrder, AdminPaymentMethod, AdminPaymentProvider } from '@/api/admin/paymentOrders'

const { t } = useI18n()

const loading = ref(false)
const exporting = ref(false)

const filters = reactive<{ method: AdminPaymentMethod | ''; user: string }>({
  method: '',
  user: ''
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

const columns = computed<Column[]>(() => [
  { key: 'order_no', label: t('payment.orderNo') },
  { key: 'user_id', label: t('admin.paymentOrders.userId') },
  { key: 'order_type', label: t('payment.orderType') },
  { key: 'provider', label: t('payment.channel') },
  { key: 'total_usd', label: t('payment.creditsAmount') },
  { key: 'amount_cny', label: t('payment.payAmountCny') },
  { key: 'status', label: t('payment.status') },
  { key: 'created_at', label: t('common.createdAt') }
])

function providerLabel(provider: AdminPaymentProvider): string {
  if (provider === 'zpay') return t('payment.alipay')
  if (provider === 'stripe') return t('payment.wechat')
  if (provider === 'admin') return t('payment.adminRecharge')
  return provider
}

function orderTypeLabel(orderType?: string): string {
  const value = String(orderType || '').toLowerCase()
  if (value === 'admin_recharge') return t('payment.orderTypeAdmin')
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
      user: filters.user || ''
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
      user: filters.user || ''
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
