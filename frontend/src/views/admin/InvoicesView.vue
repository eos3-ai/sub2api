<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex w-full flex-col gap-3 md:flex-row md:flex-wrap md:items-center md:justify-between md:gap-4">
          <!-- Left: Filters -->
          <div class="flex min-w-[280px] flex-1 flex-wrap content-start items-center gap-3">
            <div class="w-full sm:w-40">
              <Select v-model="filters.status" :options="statusOptions" :placeholder="t('common.all')" />
            </div>

            <div class="w-full sm:w-64">
              <input
                v-model.trim="filters.userEmail"
                type="text"
                class="input w-full"
                :placeholder="t('admin.invoices.userPlaceholder')"
                @keydown.enter.prevent="applyFilters"
              />
            </div>

            <div class="w-full sm:w-72">
              <div class="[&_.date-picker-trigger]:w-full">
                <DateRangePicker
                  :start-date="filters.startDate"
                  :end-date="filters.endDate"
                  @update:startDate="filters.startDate = $event"
                  @update:endDate="filters.endDate = $event"
                />
              </div>
            </div>
          </div>

          <!-- Right: Actions -->
          <div class="flex items-center justify-end gap-2 md:ml-auto">
            <button class="btn btn-secondary" :disabled="loading" @click="applyFilters">
              {{ t('common.filter') }}
            </button>
            <button class="btn btn-secondary" :disabled="loading" @click="resetFilters">
              {{ t('common.reset') }}
            </button>
            <button class="btn btn-primary" :disabled="exporting || unavailable" @click="exportCSV">
              {{ exporting ? t('common.loading') : t('admin.invoices.export') }}
            </button>
            <button class="btn btn-secondary px-2 md:px-3" :disabled="loading" :title="t('common.refresh')" @click="loadList">
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <div v-if="unavailable" class="rounded-xl border border-gray-200 bg-gray-50 p-4 text-gray-900 dark:border-dark-700 dark:bg-dark-800/50 dark:text-white">
          <p class="text-sm font-medium">{{ t('invoice.unavailableTitle') }}</p>
          <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
            {{ t('invoice.unavailableDesc') }}
          </p>
        </div>

        <DataTable v-else :columns="columns" :data="items" :loading="loading">
          <template #cell-invoice_request_no="{ value }">
            <span class="font-mono text-sm font-medium text-gray-900 dark:text-white">{{ value }}</span>
          </template>

          <template #cell-user_email="{ value }">
            <span class="text-sm text-gray-700 dark:text-dark-300">{{ value || '-' }}</span>
          </template>

          <template #cell-invoice_type="{ value }">
            <span class="text-sm text-gray-700 dark:text-dark-300">{{ invoiceTypeLabel(value) }}</span>
          </template>

          <template #cell-amount_cny_total="{ value }">
            <span class="text-base font-bold text-gray-900 dark:text-white">{{ formatCNY(value) }}</span>
          </template>

          <template #cell-status="{ value }">
            <span :class="['badge', statusBadgeClass(value)]">{{ invoiceStatusLabel(value) }}</span>
          </template>

          <template #cell-created_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">{{ value ? formatDateTime(value) : '-' }}</span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center justify-end gap-2">
              <button class="btn btn-secondary btn-sm" @click="openDetail(row.id)">
                {{ t('common.details') }}
              </button>
            </div>
          </template>

          <template #empty>
            <EmptyState :message="t('common.noData')" />
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <!-- Detail Modal -->
    <Modal
      :show="detailOpen"
      :title="t('admin.invoices.detailTitle')"
      size="xl"
      closeOnClickOutside
      @close="closeDetail"
    >
      <div v-if="detailLoading" class="flex items-center justify-center py-10">
        <LoadingSpinner />
      </div>

      <template v-else-if="detail">
        <div class="space-y-6">
          <!-- Summary -->
          <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            <div class="rounded-xl border border-gray-100 bg-white p-4 dark:border-dark-800 dark:bg-dark-900">
              <div class="text-xs font-semibold text-gray-500 dark:text-dark-400">{{ t('invoice.requestNo') }}</div>
              <div class="mt-1 font-mono text-sm text-gray-900 dark:text-white">
                {{ detail.invoice.invoice_request_no }}
              </div>
            </div>
            <div class="rounded-xl border border-gray-100 bg-white p-4 dark:border-dark-800 dark:bg-dark-900">
              <div class="text-xs font-semibold text-gray-500 dark:text-dark-400">{{ t('common.status') }}</div>
              <div class="mt-1">
                <span :class="['badge', statusBadgeClass(detail.invoice.status)]">
                  {{ invoiceStatusLabel(detail.invoice.status) }}
                </span>
              </div>
            </div>
            <div class="rounded-xl border border-gray-100 bg-white p-4 dark:border-dark-800 dark:bg-dark-900">
              <div class="text-xs font-semibold text-gray-500 dark:text-dark-400">{{ t('admin.invoices.userEmail') }}</div>
              <div class="mt-1 text-sm text-gray-900 dark:text-white">
                {{ detail.invoice.user_email || '-' }}
              </div>
            </div>
          </div>

          <!-- Amounts -->
          <div class="grid gap-3 sm:grid-cols-2">
            <div class="rounded-xl border border-gray-100 bg-white p-4 dark:border-dark-800 dark:bg-dark-900">
              <div class="text-xs font-semibold text-gray-500 dark:text-dark-400">{{ t('invoice.totalAmountCny') }}</div>
              <div class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
                {{ formatCNY(detail.invoice.amount_cny_total) }}
              </div>
            </div>
            <div class="rounded-xl border border-gray-100 bg-white p-4 dark:border-dark-800 dark:bg-dark-900">
              <div class="text-xs font-semibold text-gray-500 dark:text-dark-400">{{ t('invoice.totalCreditsUsd') }}</div>
              <div class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
                {{ formatUSD(detail.invoice.total_usd_total) }}
              </div>
            </div>
          </div>

          <!-- Invoice Info -->
          <div class="rounded-2xl border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900">
            <div class="grid gap-4 sm:grid-cols-2">
              <div>
                <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('invoice.invoiceType') }}</p>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{ invoiceTypeLabel(detail.invoice.invoice_type) }}
                </p>
              </div>
              <div>
                <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('invoice.buyerType') }}</p>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{ buyerTypeLabel(detail.invoice.buyer_type) }}
                </p>
              </div>
              <div class="sm:col-span-2">
                <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('invoice.invoiceTitle') }}</p>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{ detail.invoice.invoice_title }}
                </p>
              </div>
              <div class="sm:col-span-2">
                <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('invoice.taxNo') }}</p>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{ detail.invoice.tax_no || '-' }}
                </p>
              </div>

              <template v-if="detail.invoice.invoice_type === 'special'">
                <div class="sm:col-span-2">
                  <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('invoice.buyerAddress') }}</p>
                  <p class="mt-1 text-sm text-gray-900 dark:text-white">
                    {{ detail.invoice.buyer_address || '-' }}
                  </p>
                </div>
                <div>
                  <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('invoice.buyerPhone') }}</p>
                  <p class="mt-1 text-sm text-gray-900 dark:text-white">
                    {{ detail.invoice.buyer_phone || '-' }}
                  </p>
                </div>
                <div>
                  <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('invoice.buyerBankName') }}</p>
                  <p class="mt-1 text-sm text-gray-900 dark:text-white">
                    {{ detail.invoice.buyer_bank_name || '-' }}
                  </p>
                </div>
                <div class="sm:col-span-2">
                  <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('invoice.buyerBankAccount') }}</p>
                  <p class="mt-1 text-sm text-gray-900 dark:text-white">
                    {{ detail.invoice.buyer_bank_account || '-' }}
                  </p>
                </div>
              </template>

              <div class="sm:col-span-2">
                <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('invoice.receiverEmail') }}</p>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{ detail.invoice.receiver_email }}
                </p>
              </div>
              <div class="sm:col-span-2">
                <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('invoice.receiverPhone') }}</p>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{ detail.invoice.receiver_phone || '-' }}
                </p>
              </div>
              <div class="sm:col-span-2">
                <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('invoice.invoiceItemName') }}</p>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{ detail.invoice.invoice_item_name || '-' }}
                </p>
              </div>
              <div class="sm:col-span-2">
                <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('invoice.remark') }}</p>
                <p class="mt-1 whitespace-pre-wrap text-sm text-gray-900 dark:text-white">
                  {{ detail.invoice.remark || '-' }}
                </p>
              </div>

              <div v-if="detail.invoice.reject_reason" class="sm:col-span-2">
                <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('invoice.rejectReason') }}</p>
                <p class="mt-1 whitespace-pre-wrap text-sm text-rose-600 dark:text-rose-400">
                  {{ detail.invoice.reject_reason }}
                </p>
              </div>

              <div v-if="detail.invoice.status === 'issued'" class="sm:col-span-2">
                <div class="grid gap-4 sm:grid-cols-2">
                  <div>
                    <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.invoices.invoiceNumber') }}</p>
                    <p class="mt-1 text-sm text-gray-900 dark:text-white">
                      {{ detail.invoice.invoice_number || '-' }}
                    </p>
                  </div>
                  <div>
                    <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.invoices.invoiceDate') }}</p>
                    <p class="mt-1 text-sm text-gray-900 dark:text-white">
                      {{ detail.invoice.invoice_date ? formatDateTime(detail.invoice.invoice_date) : '-' }}
                    </p>
                  </div>
                  <div v-if="detail.invoice.invoice_pdf_url" class="sm:col-span-2">
                    <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('invoice.invoicePDF') }}</p>
                    <a
                      class="mt-1 inline-flex break-all text-sm text-primary-600 hover:underline dark:text-primary-400"
                      :href="detail.invoice.invoice_pdf_url"
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      {{ detail.invoice.invoice_pdf_url }}
                    </a>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Order Items -->
          <div class="overflow-hidden rounded-2xl border border-gray-200 dark:border-dark-700">
            <div class="table-wrapper overflow-x-auto">
              <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
                <thead class="bg-gray-50 dark:bg-dark-800/60">
                  <tr>
                    <th class="px-5 py-4 text-left text-xs font-semibold text-gray-600 dark:text-dark-300">
                      {{ t('payment.orderNo') }}
                    </th>
                    <th class="px-5 py-4 text-left text-xs font-semibold text-gray-600 dark:text-dark-300">
                      {{ t('payment.payAmountCny') }}
                    </th>
                    <th class="px-5 py-4 text-left text-xs font-semibold text-gray-600 dark:text-dark-300">
                      {{ t('payment.creditsAmount') }}
                    </th>
                    <th class="px-5 py-4 text-left text-xs font-semibold text-gray-600 dark:text-dark-300">
                      {{ t('invoice.paidAt') }}
                    </th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-100 bg-white dark:divide-dark-800 dark:bg-dark-800">
                  <tr v-for="it in detail.items" :key="it.id">
                    <td class="px-5 py-4 text-sm text-gray-900 dark:text-white">
                      <span class="font-mono text-xs">{{ it.order_no }}</span>
                    </td>
                    <td class="px-5 py-4 text-sm text-gray-700 dark:text-dark-300">
                      ¥{{ Number(it.amount_cny || 0).toFixed(2) }}
                    </td>
                    <td class="px-5 py-4 text-sm text-gray-700 dark:text-dark-300">
                      ${{ Number(it.total_usd || 0).toFixed(2) }}
                    </td>
                    <td class="px-5 py-4 text-sm text-gray-700 dark:text-dark-300">
                      {{ it.paid_at ? formatDateTime(it.paid_at) : '-' }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </template>

      <template #footer>
        <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <button class="btn btn-secondary" @click="closeDetail">
            {{ t('common.close') }}
          </button>

          <div class="flex flex-col gap-3 sm:flex-row sm:items-center">
            <button
              v-if="detail?.invoice.status === 'submitted'"
              class="btn btn-secondary"
              :disabled="approving"
              @click="approveCurrent"
            >
              {{ approving ? t('common.loading') : t('admin.invoices.approve') }}
            </button>
            <button
              v-if="detail?.invoice.status === 'submitted'"
              class="btn btn-secondary"
              @click="openReject"
            >
              {{ t('admin.invoices.reject') }}
            </button>
            <button
              v-if="detail?.invoice.status === 'approved'"
              class="btn btn-primary"
              :disabled="issuing"
              @click="openIssue"
            >
              {{ issuing ? t('common.loading') : t('admin.invoices.issue') }}
            </button>
          </div>
        </div>
      </template>
    </Modal>

    <!-- Reject Modal -->
    <Modal
      :show="rejectOpen"
      :title="t('admin.invoices.rejectTitle')"
      size="md"
      closeOnClickOutside
      @close="rejectOpen = false"
    >
      <div class="space-y-3">
        <div>
          <label class="input-label">{{ t('admin.invoices.rejectReason') }}</label>
          <textarea
            v-model="rejectReason"
            rows="4"
            class="input"
            :placeholder="t('admin.invoices.rejectReasonPlaceholder')"
          />
        </div>
      </div>

      <template #footer>
        <div class="flex items-center justify-end gap-3">
          <button class="btn btn-secondary" :disabled="rejecting" @click="rejectOpen = false">
            {{ t('common.cancel') }}
          </button>
          <button class="btn btn-primary" :disabled="rejecting" @click="submitReject">
            {{ rejecting ? t('common.loading') : t('admin.invoices.reject') }}
          </button>
        </div>
      </template>
    </Modal>

    <!-- Issue Modal -->
    <Modal
      :show="issueOpen"
      :title="t('admin.invoices.issueTitle')"
      size="md"
      closeOnClickOutside
      @close="issueOpen = false"
    >
      <div class="space-y-4">
        <div>
          <label class="input-label">{{ t('admin.invoices.invoiceNumber') }}</label>
          <input
            v-model.trim="issueForm.invoice_number"
            class="input"
            :placeholder="t('admin.invoices.invoiceNumberPlaceholder')"
          />
        </div>
        <div>
          <label class="input-label">{{ t('admin.invoices.invoiceDate') }}</label>
          <input
            v-model.trim="issueForm.invoice_date"
            class="input font-mono"
            :placeholder="t('admin.invoices.invoiceDatePlaceholder')"
          />
          <p class="input-hint">{{ t('admin.invoices.invoiceDateHint') }}</p>
        </div>
        <div>
          <label class="input-label">{{ t('admin.invoices.invoicePdfUrl') }}</label>
          <input
            v-model.trim="issueForm.invoice_pdf_url"
            class="input font-mono"
            :placeholder="t('admin.invoices.invoicePdfUrlPlaceholder')"
          />
        </div>
      </div>

      <template #footer>
        <div class="flex items-center justify-end gap-3">
          <button class="btn btn-secondary" :disabled="issuing" @click="issueOpen = false">
            {{ t('common.cancel') }}
          </button>
          <button class="btn btn-primary" :disabled="issuing" @click="submitIssue">
            {{ issuing ? t('common.loading') : t('admin.invoices.issue') }}
          </button>
        </div>
      </template>
    </Modal>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { saveAs } from 'file-saver'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Select from '@/components/common/Select.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Modal from '@/components/common/Modal.vue'
import Icon from '@/components/icons/Icon.vue'
import { adminAPI } from '@/api/admin'
import type { Column } from '@/components/common/types'
import type { AdminInvoiceRequest, AdminInvoiceStatus } from '@/api/admin/invoices'
import { formatDateTime } from '@/utils/format'
import { useAppStore } from '@/stores'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const exporting = ref(false)
const unavailable = ref(false)

const formatYMD = (d: Date): string => {
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const getDefaultRange = () => {
  const now = new Date()
  const weekAgo = new Date(now)
  weekAgo.setDate(weekAgo.getDate() - 6)
  return { startDate: formatYMD(weekAgo), endDate: formatYMD(now) }
}

const defaultRange = getDefaultRange()

const filters = reactive<{
  status: AdminInvoiceStatus | ''
  userEmail: string
  startDate: string
  endDate: string
}>({
  status: 'submitted',
  userEmail: '',
  startDate: defaultRange.startDate,
  endDate: defaultRange.endDate
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const items = ref<AdminInvoiceRequest[]>([])

const statusOptions = computed(() => [
  { label: t('common.all'), value: '' },
  { label: t('invoice.statusSubmitted'), value: 'submitted' },
  { label: t('invoice.statusApproved'), value: 'approved' },
  { label: t('invoice.statusRejected'), value: 'rejected' },
  { label: t('invoice.statusIssued'), value: 'issued' },
  { label: t('invoice.statusCancelled'), value: 'cancelled' }
])

const columns = computed<Column[]>(() => [
  { key: 'invoice_request_no', label: t('invoice.requestNo') },
  { key: 'user_email', label: t('admin.invoices.userEmail') },
  { key: 'invoice_type', label: t('invoice.invoiceType') },
  { key: 'invoice_title', label: t('invoice.invoiceTitle') },
  { key: 'amount_cny_total', label: t('invoice.totalAmountCny') },
  { key: 'status', label: t('common.status') },
  { key: 'created_at', label: t('common.createdAt') },
  { key: 'actions', label: t('common.actions') }
])

function isNotFoundError(error: unknown): boolean {
  const status = (error as { status?: number }).status
  return status === 404
}

function toRFC3339Start(dateStr: string): string {
  if (!dateStr) return ''
  const d = new Date(`${dateStr}T00:00:00`)
  if (Number.isNaN(d.getTime())) return ''
  return d.toISOString()
}

function toRFC3339End(dateStr: string): string {
  if (!dateStr) return ''
  const d = new Date(`${dateStr}T23:59:59.999`)
  if (Number.isNaN(d.getTime())) return ''
  return d.toISOString()
}

function formatCNY(amount: number | null | undefined): string {
  const n = Number(amount ?? 0)
  if (!Number.isFinite(n)) return '-'
  return `¥${n.toFixed(2)}`
}

function formatUSD(amount: number | null | undefined): string {
  const n = Number(amount ?? 0)
  if (!Number.isFinite(n)) return '-'
  return `$${n.toFixed(2)}`
}

function invoiceTypeLabel(value: string): string {
  const normalized = String(value || '').toLowerCase()
  if (normalized === 'special') return t('invoice.invoiceTypeSpecial')
  return t('invoice.invoiceTypeNormal')
}

function buyerTypeLabel(value: string): string {
  const normalized = String(value || '').toLowerCase()
  if (normalized === 'personal') return t('invoice.buyerTypePersonal')
  return t('invoice.buyerTypeCompany')
}

function invoiceStatusLabel(value: string): string {
  const normalized = String(value || '').toLowerCase()
  switch (normalized) {
    case 'submitted':
      return t('invoice.statusSubmitted')
    case 'approved':
      return t('invoice.statusApproved')
    case 'rejected':
      return t('invoice.statusRejected')
    case 'issued':
      return t('invoice.statusIssued')
    case 'cancelled':
      return t('invoice.statusCancelled')
    default:
      return value
  }
}

function statusBadgeClass(value: string): string {
  const normalized = String(value || '').toLowerCase()
  switch (normalized) {
    case 'submitted':
      return 'badge-warning'
    case 'approved':
      return 'badge-purple'
    case 'issued':
      return 'badge-success'
    case 'rejected':
      return 'badge-danger'
    case 'cancelled':
      return 'badge-gray'
    default:
      return 'badge-gray'
  }
}

type InvoiceListQuery = {
  status: AdminInvoiceStatus | ''
  user_email: string
  from: string
  to: string
}

function buildFilters(): InvoiceListQuery {
  return {
    status: filters.status,
    user_email: filters.userEmail,
    from: toRFC3339Start(filters.startDate),
    to: toRFC3339End(filters.endDate)
  }
}

async function loadList() {
  loading.value = true
  unavailable.value = false
  try {
    const resp = await adminAPI.invoices.list(pagination.page, pagination.page_size, buildFilters())
    items.value = resp.items || []
    pagination.total = resp.total || 0
  } catch (err) {
    if (isNotFoundError(err)) {
      unavailable.value = true
      items.value = []
      pagination.total = 0
      return
    }
    appStore.showError(String((err as { message?: string })?.message || t('common.error')))
  } finally {
    loading.value = false
  }
}

async function exportCSV() {
  exporting.value = true
  try {
    const blob = await adminAPI.invoices.exportRecords(buildFilters())
    const now = new Date()
    const stamp = `${now.getFullYear()}${String(now.getMonth() + 1).padStart(2, '0')}${String(now.getDate()).padStart(2, '0')}_${String(
      now.getHours()
    ).padStart(2, '0')}${String(now.getMinutes()).padStart(2, '0')}${String(now.getSeconds()).padStart(2, '0')}`
    saveAs(blob, `invoice_requests_${stamp}.csv`)
  } catch (err) {
    appStore.showError(String((err as { message?: string })?.message || t('common.error')))
  } finally {
    exporting.value = false
  }
}

function applyFilters() {
  pagination.page = 1
  loadList()
}

function resetFilters() {
  filters.status = 'submitted'
  filters.userEmail = ''
  const { startDate, endDate } = getDefaultRange()
  filters.startDate = startDate
  filters.endDate = endDate
  applyFilters()
}

function handlePageChange(page: number) {
  pagination.page = page
  loadList()
}

function handlePageSizeChange(pageSize: number) {
  pagination.page_size = pageSize
  pagination.page = 1
  loadList()
}

// Detail / actions

const detailOpen = ref(false)
const detailLoading = ref(false)
const detail = ref<Awaited<ReturnType<typeof adminAPI.invoices.getByID>> | null>(null)

const approving = ref(false)

const rejectOpen = ref(false)
const rejecting = ref(false)
const rejectReason = ref('')

const issueOpen = ref(false)
const issuing = ref(false)
const issueForm = reactive({
  invoice_number: '',
  invoice_date: '',
  invoice_pdf_url: ''
})

function closeDetail() {
  detailOpen.value = false
  detail.value = null
}

async function openDetail(id: number) {
  detailOpen.value = true
  detailLoading.value = true
  detail.value = null
  try {
    detail.value = await adminAPI.invoices.getByID(id)
  } catch (err) {
    appStore.showError(String((err as { message?: string })?.message || t('common.error')))
    detailOpen.value = false
  } finally {
    detailLoading.value = false
  }
}

async function approveCurrent() {
  const invoice = detail.value?.invoice
  if (!invoice) return
  if (!confirm(t('admin.invoices.approveConfirm'))) return

  approving.value = true
  try {
    const updated = await adminAPI.invoices.approve(invoice.id)
    if (detail.value) {
      detail.value.invoice = updated
    }
    appStore.showSuccess(t('admin.invoices.approveSuccess'))
    await loadList()
  } catch (err) {
    appStore.showError(String((err as { message?: string })?.message || t('common.error')))
  } finally {
    approving.value = false
  }
}

function openReject() {
  rejectReason.value = ''
  rejectOpen.value = true
}

async function submitReject() {
  const invoice = detail.value?.invoice
  if (!invoice) return
  const reason = rejectReason.value.trim()
  if (!reason) {
    appStore.showWarning(t('admin.invoices.rejectReasonRequired'))
    return
  }

  rejecting.value = true
  try {
    const updated = await adminAPI.invoices.reject(invoice.id, reason)
    if (detail.value) {
      detail.value.invoice = updated
    }
    rejectOpen.value = false
    rejectReason.value = ''
    appStore.showSuccess(t('admin.invoices.rejectSuccess'))
    await loadList()
  } catch (err) {
    appStore.showError(String((err as { message?: string })?.message || t('common.error')))
  } finally {
    rejecting.value = false
  }
}

function openIssue() {
  const invoice = detail.value?.invoice
  if (!invoice) return

  issuing.value = true
  adminAPI.invoices
    .issue(invoice.id, {
      invoice_number: '',
      invoice_pdf_url: ''
    })
    .then(async (updated) => {
      if (detail.value) {
        detail.value.invoice = updated
      }
      appStore.showSuccess(t('admin.invoices.issueSuccess'))
      await loadList()
    })
    .catch((err) => {
      appStore.showError(String((err as { message?: string })?.message || t('common.error')))
    })
    .finally(() => {
      issuing.value = false
    })
}

function isValidInvoiceDate(value: string): boolean {
  const v = value.trim()
  if (v === '') return true
  // Backend expects YYYY-MM-DD, optional.
  return /^\d{4}-\d{2}-\d{2}$/.test(v)
}

async function submitIssue() {
  const invoice = detail.value?.invoice
  if (!invoice) return

  const invoiceNumber = issueForm.invoice_number.trim()
  const invoiceDate = issueForm.invoice_date.trim()
  const pdfURL = issueForm.invoice_pdf_url.trim()

  if (!invoiceNumber) {
    appStore.showWarning(t('admin.invoices.invoiceNumberRequired'))
    return
  }
  if (!pdfURL) {
    appStore.showWarning(t('admin.invoices.invoicePdfUrlRequired'))
    return
  }
  if (!isValidInvoiceDate(invoiceDate)) {
    appStore.showWarning(t('admin.invoices.invoiceDateInvalid'))
    return
  }

  issuing.value = true
  try {
    const updated = await adminAPI.invoices.issue(invoice.id, {
      invoice_number: invoiceNumber,
      invoice_date: invoiceDate || undefined,
      invoice_pdf_url: pdfURL
    })
    if (detail.value) {
      detail.value.invoice = updated
    }
    issueOpen.value = false
    appStore.showSuccess(t('admin.invoices.issueSuccess'))
    await loadList()
  } catch (err) {
    appStore.showError(String((err as { message?: string })?.message || t('common.error')))
  } finally {
    issuing.value = false
  }
}

onMounted(() => {
  loadList()
})
</script>
