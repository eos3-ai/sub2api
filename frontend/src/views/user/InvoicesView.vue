<template>
  <AppLayout :title="t('invoice.title')" :description="t('invoice.description')">
    <div class="space-y-6">
      <div class="card animate-fade-in-up p-6">
        <div class="mb-4 flex items-center justify-between gap-3">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('invoice.myRequests') }}</h2>
          <div class="flex items-center gap-3">
            <button class="btn btn-secondary" @click="createOpen = true">
              {{ t('invoice.createTitle') }}
            </button>
            <button class="btn btn-secondary" :disabled="loading" @click="load">
              {{ loading ? t('common.loading') : t('common.refresh') }}
            </button>
          </div>
        </div>

        <div v-if="loading" class="flex items-center justify-center py-10">
          <LoadingSpinner />
        </div>

        <template v-else>
          <div
            v-if="unavailable"
            class="rounded-xl border border-gray-200 bg-gray-50 p-4 text-gray-900 dark:border-dark-700 dark:bg-dark-800/50 dark:text-white"
          >
            <p class="text-sm font-medium">{{ t('invoice.unavailableTitle') }}</p>
            <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
              {{ t('invoice.unavailableDesc') }}
            </p>
          </div>

          <div v-else-if="items.length === 0" class="py-10 text-center text-sm text-gray-500 dark:text-dark-400">
            {{ t('invoice.empty') }}
          </div>

          <div v-else class="overflow-hidden rounded-2xl border border-gray-200 dark:border-dark-700">
            <div class="table-wrapper overflow-x-auto">
              <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
                <thead class="bg-gray-50 dark:bg-dark-800/60">
                  <tr>
                    <th class="px-5 py-4 text-left text-xs font-semibold text-gray-600 dark:text-dark-300">
                      {{ t('invoice.requestNo') }}
                    </th>
                    <th class="px-5 py-4 text-left text-xs font-semibold text-gray-600 dark:text-dark-300">
                      {{ t('invoice.invoiceType') }}
                    </th>
                    <th class="px-5 py-4 text-left text-xs font-semibold text-gray-600 dark:text-dark-300">
                      {{ t('invoice.invoiceTitle') }}
                    </th>
                    <th class="px-5 py-4 text-left text-xs font-semibold text-gray-600 dark:text-dark-300">
                      {{ t('invoice.totalAmountCny') }}
                    </th>
                    <th class="px-5 py-4 text-left text-xs font-semibold text-gray-600 dark:text-dark-300">
                      {{ t('common.status') }}
                    </th>
                    <th class="px-5 py-4 text-left text-xs font-semibold text-gray-600 dark:text-dark-300">
                      {{ t('common.createdAt') }}
                    </th>
                    <th class="px-5 py-4 text-right text-xs font-semibold text-gray-600 dark:text-dark-300">
                      {{ t('common.actions') }}
                    </th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-100 bg-white dark:divide-dark-800 dark:bg-dark-800">
                  <tr v-for="r in items" :key="r.id">
                    <td class="px-5 py-4 text-sm text-gray-900 dark:text-white">
                      <span class="font-mono text-xs">{{ r.invoice_request_no }}</span>
                    </td>
                    <td class="px-5 py-4 text-sm text-gray-700 dark:text-dark-300">
                      {{ invoiceTypeLabel(r.invoice_type) }}
                    </td>
                    <td class="px-5 py-4 text-sm text-gray-700 dark:text-dark-300">
                      {{ r.invoice_title }}
                    </td>
                    <td class="px-5 py-4 text-sm text-gray-700 dark:text-dark-300">
                      ¥{{ r.amount_cny_total.toFixed(2) }}
                    </td>
                    <td class="px-5 py-4 text-sm text-gray-700 dark:text-dark-300">
                      {{ invoiceStatusLabel(r.status) }}
                    </td>
                    <td class="px-5 py-4 text-sm text-gray-700 dark:text-dark-300">
                      {{ formatDateTime(r.created_at) }}
                    </td>
                    <td class="px-5 py-4 text-right text-sm">
                      <div class="flex flex-wrap justify-end gap-2">
                        <button class="btn btn-secondary btn-sm" @click="openDetail(r.id)">
                          {{ t('common.details') }}
                        </button>
                        <button
                          v-if="r.status === 'submitted'"
                          class="btn btn-secondary btn-sm"
                          :disabled="cancellingId === r.id"
                          @click="cancel(r.id)"
                        >
                          {{ cancellingId === r.id ? t('common.loading') : t('invoice.cancel') }}
                        </button>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <Pagination
              v-if="total > pageSize"
              :total="total"
              :page="page"
              :page-size="pageSize"
              :show-page-size-selector="false"
              @update:page="page = $event"
            />
          </div>
        </template>
      </div>
    </div>

    <InvoiceRequestModal :show="createOpen" @close="createOpen = false" @created="handleCreated" />

    <Modal
      :show="detailOpen"
      :title="t('invoice.detailsTitle')"
      size="lg"
      closeOnClickOutside
      @close="detailOpen = false"
    >
      <div v-if="detailLoading" class="flex items-center justify-center py-10">
        <LoadingSpinner />
      </div>

      <template v-else-if="detail">
        <div class="space-y-6">
          <div class="rounded-2xl bg-gray-50 p-4 dark:bg-dark-900/30">
            <div class="grid gap-3 sm:grid-cols-2">
              <div>
                <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('invoice.requestNo') }}</p>
                <p class="mt-1 font-mono text-sm font-semibold text-gray-900 dark:text-white">
                  {{ detail.invoice.invoice_request_no }}
                </p>
              </div>
              <div>
                <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('common.status') }}</p>
                <p class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">
                  {{ invoiceStatusLabel(detail.invoice.status) }}
                </p>
              </div>
              <div class="sm:col-span-2">
                <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('invoice.invoiceTitle') }}</p>
                <p class="mt-1 text-sm text-gray-900 dark:text-white">
                  {{ detail.invoice.invoice_title }}
                </p>
              </div>
              <div>
                <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('invoice.totalAmountCny') }}</p>
                <p class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">
                  ¥{{ detail.invoice.amount_cny_total.toFixed(2) }}
                </p>
              </div>
              <div>
                <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('invoice.totalCreditsUsd') }}</p>
                <p class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">
                  ${{ detail.invoice.total_usd_total.toFixed(2) }}
                </p>
              </div>
              <div v-if="detail.invoice.reject_reason" class="sm:col-span-2">
                <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('invoice.rejectReason') }}</p>
                <p class="mt-1 text-sm text-rose-600 dark:text-rose-400">
                  {{ detail.invoice.reject_reason }}
                </p>
              </div>
            </div>
          </div>

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
                      ¥{{ it.amount_cny.toFixed(2) }}
                    </td>
                    <td class="px-5 py-4 text-sm text-gray-700 dark:text-dark-300">
                      ${{ it.total_usd.toFixed(2) }}
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
        <button class="btn btn-secondary" @click="detailOpen = false">
          {{ t('common.close') }}
        </button>
      </template>
    </Modal>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Pagination from '@/components/common/Pagination.vue'
import Modal from '@/components/common/Modal.vue'
import InvoiceRequestModal from '@/components/user/InvoiceRequestModal.vue'
import { invoiceAPI, type InvoiceRequest } from '@/api/invoices'
import { formatDateTime } from '@/utils/format'
import { useAppStore } from '@/stores'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const unavailable = ref(false)

const items = ref<InvoiceRequest[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const createOpen = ref(false)

const cancellingId = ref<number | null>(null)

const detailOpen = ref(false)
const detailLoading = ref(false)
const detail = ref<Awaited<ReturnType<typeof invoiceAPI.getMyInvoiceRequest>> | null>(null)

function isNotFoundError(error: unknown): boolean {
  const status = (error as { status?: number }).status
  return status === 404
}

function invoiceTypeLabel(value: string): string {
  const normalized = String(value || '').toLowerCase()
  if (normalized === 'special') return t('invoice.invoiceTypeSpecial')
  return t('invoice.invoiceTypeNormal')
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

async function load() {
  loading.value = true
  unavailable.value = false
  try {
    const resp = await invoiceAPI.getMyInvoiceRequests({ page: page.value, page_size: pageSize.value })
    items.value = resp.items || []
    total.value = resp.total || 0
  } catch (err) {
    if (isNotFoundError(err)) {
      unavailable.value = true
      items.value = []
      total.value = 0
    } else {
      appStore.showError(String((err as { message?: string })?.message || t('common.error')))
    }
  } finally {
    loading.value = false
  }
}

watch(
  () => page.value,
  async () => {
    await load()
  }
)

async function cancel(id: number) {
  cancellingId.value = id
  try {
    await invoiceAPI.cancelInvoiceRequest(id)
    appStore.showSuccess(t('invoice.cancelSuccess'))
    await load()
  } catch (err) {
    appStore.showError(String((err as { message?: string })?.message || t('common.error')))
  } finally {
    cancellingId.value = null
  }
}

async function openDetail(id: number) {
  detailOpen.value = true
  detailLoading.value = true
  detail.value = null
  try {
    detail.value = await invoiceAPI.getMyInvoiceRequest(id)
  } catch (err) {
    appStore.showError(String((err as { message?: string })?.message || t('common.error')))
    detailOpen.value = false
  } finally {
    detailLoading.value = false
  }
}

function handleCreated() {
  createOpen.value = false
  load()
}

onMounted(() => {
  load()
})
</script>
