<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="card p-6">
        <div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">开票申请</h2>
          <div class="flex gap-2">
            <button class="btn btn-secondary" :disabled="loading" @click="load">刷新</button>
            <button class="btn btn-primary" @click="createOpen = true">申请发票</button>
          </div>
        </div>

        <div v-if="loading" class="flex justify-center py-10">
          <LoadingSpinner />
        </div>

        <div v-else-if="items.length === 0" class="rounded-lg border border-gray-200 p-8 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
          暂无开票申请
        </div>

        <div v-else class="overflow-x-auto rounded-lg border border-gray-200 dark:border-dark-700">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
            <thead class="bg-gray-50 dark:bg-dark-800">
              <tr>
                <th class="px-4 py-3 text-left text-xs font-semibold text-gray-500">申请号</th>
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
                <td class="px-4 py-3 text-sm text-gray-700 dark:text-dark-300">{{ item.invoice_title }}</td>
                <td class="px-4 py-3 text-right text-sm font-semibold text-gray-900 dark:text-white">¥{{ item.amount_cny_total.toFixed(2) }}</td>
                <td class="px-4 py-3">
                  <span :class="['badge', statusClass(item.status)]">{{ statusLabel(item.status) }}</span>
                </td>
                <td class="px-4 py-3 text-sm text-gray-500 dark:text-dark-400">{{ formatDateTime(item.created_at) }}</td>
                <td class="px-4 py-3">
                  <div class="flex justify-end gap-2">
                    <button class="btn btn-secondary btn-sm" @click="openDetail(item.id)">详情</button>
                    <button
                      v-if="item.status === 'submitted'"
                      class="btn btn-secondary btn-sm"
                      :disabled="cancellingId === item.id"
                      @click="cancelRequest(item.id)"
                    >
                      取消
                    </button>
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
          :show-page-size-selector="false"
          @update:page="handlePageChange"
        />
      </div>
    </div>

    <InvoiceRequestModal :show="createOpen" @close="createOpen = false" @created="handleCreated" />

    <BaseDialog :show="detailOpen" title="发票详情" width="wide" close-on-click-outside @close="detailOpen = false">
      <div v-if="detailLoading" class="flex justify-center py-10">
        <LoadingSpinner />
      </div>
      <div v-else-if="detail" class="space-y-5">
        <div class="grid gap-3 sm:grid-cols-2">
          <InfoItem label="申请号" :value="detail.invoice.invoice_request_no" mono />
          <InfoItem label="状态" :value="statusLabel(detail.invoice.status)" />
          <InfoItem label="发票抬头" :value="detail.invoice.invoice_title" />
          <InfoItem label="开票金额" :value="`¥${detail.invoice.amount_cny_total.toFixed(2)}`" />
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
import { defineComponent, h, onMounted, ref, watch } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Pagination from '@/components/common/Pagination.vue'
import InvoiceRequestModal from '@/components/user/InvoiceRequestModal.vue'
import { invoiceAPI, type InvoiceDetail, type InvoiceRequest, type InvoiceStatus } from '@/api/invoices'
import { useAppStore } from '@/stores'
import { formatDateTime } from '@/utils/format'

const appStore = useAppStore()
const loading = ref(false)
const items = ref<InvoiceRequest[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const createOpen = ref(false)
const detailOpen = ref(false)
const detailLoading = ref(false)
const detail = ref<InvoiceDetail | null>(null)
const cancellingId = ref<number | null>(null)

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
    const { data } = await invoiceAPI.getMyRequests({ page: page.value, page_size: pageSize.value })
    items.value = data.items || []
    total.value = data.total || 0
  } catch (error) {
    appStore.showError((error as { message?: string })?.message || '加载发票申请失败')
  } finally {
    loading.value = false
  }
}

function handlePageChange(nextPage: number) {
  page.value = nextPage
}

function handleCreated(invoice: InvoiceRequest) {
  createOpen.value = false
  page.value = 1
  items.value = [invoice, ...items.value]
  total.value += 1
  load()
}

async function openDetail(id: number) {
  detailOpen.value = true
  detailLoading.value = true
  try {
    const { data } = await invoiceAPI.getRequest(id)
    detail.value = data
  } catch (error) {
    appStore.showError((error as { message?: string })?.message || '加载发票详情失败')
  } finally {
    detailLoading.value = false
  }
}

async function cancelRequest(id: number) {
  cancellingId.value = id
  try {
    await invoiceAPI.cancelRequest(id)
    appStore.showSuccess('开票申请已取消')
    await load()
  } catch (error) {
    appStore.showError((error as { message?: string })?.message || '取消开票申请失败')
  } finally {
    cancellingId.value = null
  }
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
