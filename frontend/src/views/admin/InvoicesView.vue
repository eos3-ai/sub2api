<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="card p-6">
        <div class="mb-4 flex flex-col gap-3 xl:flex-row xl:items-center xl:justify-between">
          <div class="flex flex-wrap gap-3">
            <select v-model="filters.status" class="input w-36">
              <option value="">全部状态</option>
              <option value="submitted">待审核</option>
              <option value="approved">已通过</option>
              <option value="issued">已开票</option>
              <option value="rejected">已驳回</option>
              <option value="cancelled">已取消</option>
            </select>
            <input v-model.trim="filters.user_email" class="input w-64" placeholder="用户邮箱" @keydown.enter.prevent="applyFilters" />
            <input v-model="filters.from" type="datetime-local" class="input w-56" />
            <input v-model="filters.to" type="datetime-local" class="input w-56" />
          </div>
          <div class="flex gap-2">
            <button class="btn btn-secondary" :disabled="loading" @click="applyFilters">筛选</button>
            <button class="btn btn-secondary" :disabled="loading" @click="resetFilters">重置</button>
            <button class="btn btn-primary" :disabled="exporting" @click="exportCSV">{{ exporting ? '导出中' : '导出 CSV' }}</button>
          </div>
        </div>

        <div v-if="loading" class="flex justify-center py-10">
          <LoadingSpinner />
        </div>

        <div v-else-if="items.length === 0" class="rounded-lg border border-gray-200 p-8 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
          暂无发票申请
        </div>

        <div v-else class="overflow-x-auto rounded-lg border border-gray-200 dark:border-dark-700">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
            <thead class="bg-gray-50 dark:bg-dark-800">
              <tr>
                <th class="px-4 py-3 text-left text-xs font-semibold text-gray-500">申请号</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-gray-500">用户</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-gray-500">抬头</th>
                <th class="px-4 py-3 text-right text-xs font-semibold text-gray-500">金额</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-gray-500">状态</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-gray-500">提交时间</th>
                <th class="px-4 py-3 text-right text-xs font-semibold text-gray-500">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 bg-white dark:divide-dark-800 dark:bg-dark-900">
              <tr v-for="item in items" :key="item.id">
                <td class="px-4 py-3 font-mono text-xs text-gray-900 dark:text-white">{{ item.invoice_request_no }}</td>
                <td class="px-4 py-3 text-sm text-gray-700 dark:text-dark-300">{{ item.user_email || item.user_id }}</td>
                <td class="px-4 py-3 text-sm text-gray-700 dark:text-dark-300">{{ item.invoice_title }}</td>
                <td class="px-4 py-3 text-right text-sm font-semibold text-gray-900 dark:text-white">¥{{ item.amount_cny_total.toFixed(2) }}</td>
                <td class="px-4 py-3"><span :class="['badge', statusClass(item.status)]">{{ statusLabel(item.status) }}</span></td>
                <td class="px-4 py-3 text-sm text-gray-500 dark:text-dark-400">{{ formatDateTime(item.created_at) }}</td>
                <td class="px-4 py-3">
                  <div class="flex flex-wrap justify-end gap-2">
                    <button class="btn btn-secondary btn-sm" @click="openDetail(item.id)">详情</button>
                    <button v-if="item.status === 'submitted'" class="btn btn-primary btn-sm" @click="approve(item.id)">通过</button>
                    <button v-if="item.status === 'submitted'" class="btn btn-danger btn-sm" @click="reject(item.id)">驳回</button>
                    <button v-if="item.status === 'approved'" class="btn btn-primary btn-sm" @click="issue(item.id)">开票</button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <Pagination
          v-if="total > pageSize"
          class="mt-4"
          :page="page"
          :total="total"
          :page-size="pageSize"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </div>
    </div>

    <BaseDialog :show="detailOpen" title="发票申请详情" width="wide" close-on-click-outside @close="detailOpen = false">
      <div v-if="detailLoading" class="flex justify-center py-10">
        <LoadingSpinner />
      </div>
      <div v-else-if="detail" class="space-y-5">
        <div class="grid gap-3 sm:grid-cols-2">
          <InfoItem label="申请号" :value="detail.invoice.invoice_request_no" mono />
          <InfoItem label="用户" :value="detail.invoice.user_email || String(detail.invoice.user_id)" />
          <InfoItem label="状态" :value="statusLabel(detail.invoice.status)" />
          <InfoItem label="开票金额" :value="`¥${detail.invoice.amount_cny_total.toFixed(2)}`" />
          <InfoItem label="抬头" :value="detail.invoice.invoice_title" />
          <InfoItem label="税号" :value="detail.invoice.tax_no || '-'" />
          <InfoItem label="收票邮箱" :value="detail.invoice.receiver_email" />
          <InfoItem label="开票内容" :value="detail.invoice.invoice_item_name || '-'" />
          <InfoItem v-if="detail.invoice.invoice_number" label="发票号码" :value="detail.invoice.invoice_number" />
          <InfoItem v-if="detail.invoice.invoice_pdf_url" label="发票链接" :value="detail.invoice.invoice_pdf_url" />
        </div>
        <p v-if="detail.invoice.reject_reason" class="rounded-lg bg-rose-50 p-3 text-sm text-rose-700 dark:bg-rose-950/30 dark:text-rose-300">
          {{ detail.invoice.reject_reason }}
        </p>
        <div class="overflow-x-auto rounded-lg border border-gray-200 dark:border-dark-700">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
            <thead class="bg-gray-50 dark:bg-dark-800">
              <tr>
                <th class="px-4 py-3 text-left text-xs font-semibold text-gray-500">订单号</th>
                <th class="px-4 py-3 text-right text-xs font-semibold text-gray-500">支付金额</th>
                <th class="px-4 py-3 text-right text-xs font-semibold text-gray-500">余额/套餐额</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-gray-500">支付时间</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 bg-white dark:divide-dark-800 dark:bg-dark-900">
              <tr v-for="order in detail.items" :key="order.id">
                <td class="px-4 py-3 font-mono text-xs text-gray-900 dark:text-white">{{ order.order_no }}</td>
                <td class="px-4 py-3 text-right text-sm text-gray-700 dark:text-dark-300">¥{{ order.amount_cny.toFixed(2) }}</td>
                <td class="px-4 py-3 text-right text-sm text-gray-700 dark:text-dark-300">${{ order.total_usd.toFixed(2) }}</td>
                <td class="px-4 py-3 text-sm text-gray-500 dark:text-dark-400">{{ order.paid_at ? formatDateTime(order.paid_at) : '-' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { defineComponent, h, onMounted, reactive, ref, watch } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Pagination from '@/components/common/Pagination.vue'
import { adminInvoicesAPI } from '@/api/admin/invoices'
import type { InvoiceDetail, InvoiceRequest, InvoiceStatus } from '@/api/invoices'
import { useAppStore } from '@/stores'
import { formatDateTime } from '@/utils/format'

const appStore = useAppStore()
const loading = ref(false)
const exporting = ref(false)
const items = ref<InvoiceRequest[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const detailOpen = ref(false)
const detailLoading = ref(false)
const detail = ref<InvoiceDetail | null>(null)

const filters = reactive({
  status: '' as InvoiceStatus | '',
  user_email: '',
  from: '',
  to: ''
})

const InfoItem = defineComponent({
  props: {
    label: { type: String, required: true },
    value: { type: String, required: true },
    mono: { type: Boolean, default: false }
  },
  setup(props) {
    return () =>
      h('div', { class: 'rounded-lg bg-gray-50 p-3 dark:bg-dark-800' }, [
        h('p', { class: 'text-xs text-gray-500 dark:text-dark-400' }, props.label),
        h('p', { class: ['mt-1 text-sm text-gray-900 dark:text-white', props.mono ? 'font-mono' : ''] }, props.value)
      ])
  }
})

onMounted(load)
watch(page, load)

async function load() {
  loading.value = true
  try {
    const { data } = await adminInvoicesAPI.list({
      page: page.value,
      page_size: pageSize.value,
      status: filters.status || undefined,
      user_email: filters.user_email || undefined,
      from: toRFC3339(filters.from),
      to: toRFC3339(filters.to)
    })
    items.value = data.items || []
    total.value = data.total || 0
  } catch (error) {
    appStore.showError((error as { message?: string })?.message || '加载发票申请失败')
  } finally {
    loading.value = false
  }
}

function applyFilters() {
  page.value = 1
  load()
}

function resetFilters() {
  filters.status = ''
  filters.user_email = ''
  filters.from = ''
  filters.to = ''
  applyFilters()
}

function handlePageChange(nextPage: number) {
  page.value = nextPage
}

function handlePageSizeChange(nextPageSize: number) {
  pageSize.value = nextPageSize
  page.value = 1
  load()
}

async function openDetail(id: number) {
  detailOpen.value = true
  detailLoading.value = true
  try {
    const { data } = await adminInvoicesAPI.getByID(id)
    detail.value = data
  } catch (error) {
    appStore.showError((error as { message?: string })?.message || '加载发票详情失败')
  } finally {
    detailLoading.value = false
  }
}

async function approve(id: number) {
  try {
    await adminInvoicesAPI.approve(id)
    appStore.showSuccess('已通过开票申请')
    await load()
  } catch (error) {
    appStore.showError((error as { message?: string })?.message || '审批失败')
  }
}

async function reject(id: number) {
  const reason = window.prompt('请输入驳回原因')
  if (!reason) return
  try {
    await adminInvoicesAPI.reject(id, reason)
    appStore.showSuccess('已驳回开票申请')
    await load()
  } catch (error) {
    appStore.showError((error as { message?: string })?.message || '驳回失败')
  }
}

async function issue(id: number) {
  const invoice_number = window.prompt('请输入发票号码') || ''
  if (!invoice_number.trim()) return
  const invoice_pdf_url = window.prompt('请输入发票下载链接（可选）') || ''
  const invoice_date = new Date().toISOString().slice(0, 10)
  try {
    await adminInvoicesAPI.issue(id, { invoice_number, invoice_pdf_url, invoice_date })
    appStore.showSuccess('已标记为开票完成')
    await load()
  } catch (error) {
    appStore.showError((error as { message?: string })?.message || '开票失败')
  }
}

async function exportCSV() {
  exporting.value = true
  try {
    const { data } = await adminInvoicesAPI.export({
      status: filters.status || undefined,
      user_email: filters.user_email || undefined,
      from: toRFC3339(filters.from),
      to: toRFC3339(filters.to)
    })
    const url = URL.createObjectURL(data)
    const link = document.createElement('a')
    link.href = url
    link.download = `invoice_requests_${new Date().toISOString().slice(0, 10)}.csv`
    link.click()
    URL.revokeObjectURL(url)
  } catch (error) {
    appStore.showError((error as { message?: string })?.message || '导出失败')
  } finally {
    exporting.value = false
  }
}

function toRFC3339(value: string): string | undefined {
  if (!value) return undefined
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return undefined
  return date.toISOString()
}

function statusLabel(status: InvoiceStatus | string): string {
  const labels: Record<string, string> = {
    submitted: '待审核',
    approved: '已通过',
    rejected: '已驳回',
    issued: '已开票',
    cancelled: '已取消'
  }
  return labels[status] || status
}

function statusClass(status: InvoiceStatus | string): string {
  const classes: Record<string, string> = {
    submitted: 'badge-warning',
    approved: 'badge-primary',
    issued: 'badge-success',
    rejected: 'badge-danger',
    cancelled: 'badge-gray'
  }
  return classes[status] || 'badge-gray'
}
</script>
