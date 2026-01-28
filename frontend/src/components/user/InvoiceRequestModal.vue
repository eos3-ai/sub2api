<template>
  <Modal
    :show="show"
    :title="t('invoice.createTitle')"
    size="xl"
    closeOnClickOutside
    @close="handleClose"
  >
    <div class="space-y-6">
      <div v-if="loading" class="flex items-center justify-center py-10">
        <LoadingSpinner />
      </div>

      <div
        v-else-if="unavailable"
        class="rounded-xl border border-gray-200 bg-gray-50 p-4 text-gray-900 dark:border-dark-700 dark:bg-dark-800/50 dark:text-white"
      >
        <p class="text-sm font-medium">{{ t('invoice.unavailableTitle') }}</p>
        <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
          {{ t('invoice.unavailableDesc') }}
        </p>
      </div>

      <div v-else class="grid gap-6 lg:grid-cols-2">
        <!-- Eligible orders -->
        <section class="space-y-3">
          <div class="flex items-center justify-between gap-3">
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ t('invoice.eligibleOrders') }}
            </h3>
            <p class="text-xs text-gray-500 dark:text-dark-400">
              {{ t('invoice.mergeLimitHint', { n: mergeLimit }) }}
            </p>
          </div>

          <div v-if="loadingEligible" class="flex items-center justify-center py-8">
            <LoadingSpinner />
          </div>

          <div v-else-if="eligibleOrders.length === 0" class="py-10 text-center text-sm text-gray-500 dark:text-dark-400">
            {{ t('invoice.noEligibleOrders') }}
          </div>

          <div v-else class="overflow-hidden rounded-2xl border border-gray-200 dark:border-dark-700">
            <div class="table-wrapper overflow-x-auto">
              <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
                <thead class="bg-gray-50 dark:bg-dark-800/60">
                  <tr>
                    <th class="w-10 px-4 py-4 text-left text-xs font-semibold text-gray-600 dark:text-dark-300">
                      {{ t('invoice.select') }}
                    </th>
                    <th class="px-4 py-4 text-left text-xs font-semibold text-gray-600 dark:text-dark-300">
                      {{ t('payment.orderNo') }}
                    </th>
                    <th class="px-4 py-4 text-left text-xs font-semibold text-gray-600 dark:text-dark-300">
                      {{ t('payment.payAmountCny') }}
                    </th>
                    <th class="px-4 py-4 text-left text-xs font-semibold text-gray-600 dark:text-dark-300">
                      {{ t('payment.creditsAmount') }}
                    </th>
                    <th class="px-4 py-4 text-left text-xs font-semibold text-gray-600 dark:text-dark-300">
                      {{ t('payment.status') }}
                    </th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-100 bg-white dark:divide-dark-800 dark:bg-dark-800">
                  <tr v-for="o in eligibleOrders" :key="o.order_no">
                    <td class="px-4 py-4">
                      <input
                        type="checkbox"
                        class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500 dark:border-dark-600 dark:bg-dark-900"
                        :checked="isSelected(o.order_no)"
                        :disabled="isSelectionDisabled(o.order_no)"
                        @change="toggleOrder(o.order_no)"
                      />
                    </td>
                    <td class="px-4 py-4 text-sm text-gray-900 dark:text-white">
                      <span class="font-mono text-xs">{{ o.order_no }}</span>
                    </td>
                    <td class="px-4 py-4 text-sm text-gray-700 dark:text-dark-300">¥{{ o.amount_cny.toFixed(2) }}</td>
                    <td class="px-4 py-4 text-sm text-gray-700 dark:text-dark-300">${{ o.total_usd.toFixed(2) }}</td>
                    <td class="px-4 py-4 text-sm text-gray-700 dark:text-dark-300">
                      {{ paymentStatusLabel(o.status) }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <Pagination
              v-if="eligibleTotal > eligiblePageSize"
              :total="eligibleTotal"
              :page="eligiblePage"
              :page-size="eligiblePageSize"
              :show-page-size-selector="false"
              @update:page="eligiblePage = $event"
            />
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div class="rounded-2xl bg-gray-50 p-4 dark:bg-dark-900/30">
              <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('invoice.totalAmountCny') }}</p>
              <p class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">¥{{ selectedAmountCNY.toFixed(2) }}</p>
            </div>
            <div class="rounded-2xl bg-gray-50 p-4 dark:bg-dark-900/30">
              <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('invoice.totalCreditsUsd') }}</p>
              <p class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">${{ selectedTotalUSD.toFixed(2) }}</p>
            </div>
          </div>
        </section>

        <!-- Invoice form -->
        <section class="space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="input-label">{{ t('invoice.invoiceType') }}</label>
              <select v-model="form.invoiceType" class="input" disabled>
                <option value="normal">{{ t('invoice.invoiceTypeNormal') }}</option>
              </select>
            </div>

            <div>
              <label class="input-label">{{ t('invoice.buyerType') }}</label>
              <select v-model="form.buyerType" class="input" :disabled="form.invoiceType === 'special'">
                <option value="personal">{{ t('invoice.buyerTypePersonal') }}</option>
                <option value="company">{{ t('invoice.buyerTypeCompany') }}</option>
              </select>
              <p v-if="form.invoiceType === 'special'" class="input-hint">{{ t('invoice.specialCompanyOnlyHint') }}</p>
            </div>
          </div>

          <div>
            <label class="input-label">
              {{ t('invoice.invoiceTitle') }}
              <span class="text-rose-600 dark:text-rose-400">*</span>
            </label>
            <input
              v-model="form.invoiceTitle"
              class="input"
              required
              :placeholder="t('invoice.invoiceTitlePlaceholder')"
            />
          </div>

          <div>
            <label class="input-label">
              {{ t('invoice.taxNo') }}
              <span v-if="form.buyerType === 'company'" class="text-rose-600 dark:text-rose-400">*</span>
            </label>
            <input v-model="form.taxNo" class="input" :placeholder="t('invoice.taxNoPlaceholder')" />
          </div>

          <div v-if="form.invoiceType === 'special'" class="grid grid-cols-2 gap-4">
            <div class="col-span-2">
              <label class="input-label">{{ t('invoice.buyerAddress') }}</label>
              <input v-model="form.buyerAddress" class="input" :placeholder="t('invoice.buyerAddressPlaceholder')" />
            </div>
            <div>
              <label class="input-label">{{ t('invoice.buyerPhone') }}</label>
              <input v-model="form.buyerPhone" class="input" :placeholder="t('invoice.buyerPhonePlaceholder')" />
            </div>
            <div>
              <label class="input-label">{{ t('invoice.buyerBankName') }}</label>
              <input v-model="form.buyerBankName" class="input" :placeholder="t('invoice.buyerBankNamePlaceholder')" />
            </div>
            <div class="col-span-2">
              <label class="input-label">{{ t('invoice.buyerBankAccount') }}</label>
              <input v-model="form.buyerBankAccount" class="input" :placeholder="t('invoice.buyerBankAccountPlaceholder')" />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="col-span-2">
              <label class="input-label">
                {{ t('invoice.receiverEmail') }}
                <span class="text-rose-600 dark:text-rose-400">*</span>
              </label>
              <input
                v-model="form.receiverEmail"
                class="input"
                type="email"
                required
                :placeholder="t('invoice.receiverEmailPlaceholder')"
              />
            </div>
            <div class="col-span-2">
              <label class="input-label">{{ t('invoice.receiverPhone') }}</label>
              <input v-model="form.receiverPhone" class="input" :placeholder="t('invoice.receiverPhonePlaceholder')" />
            </div>
          </div>

          <div>
            <label class="input-label">{{ t('invoice.invoiceItemName') }}</label>
            <input v-model="form.invoiceItemName" class="input" :placeholder="t('invoice.invoiceItemNamePlaceholder')" />
            <p class="input-hint">{{ t('invoice.invoiceItemNameHint') }}</p>
          </div>

          <div>
            <label class="input-label">{{ t('invoice.remark') }}</label>
            <textarea v-model="form.remark" rows="3" class="input" :placeholder="t('invoice.remarkPlaceholder')" />
          </div>
        </section>
      </div>
    </div>

    <template #footer>
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <p class="text-sm text-gray-500 dark:text-dark-400">
          {{ t('invoice.selectedCountHint', { n: selectedOrderNos.length, max: mergeLimit }) }}
        </p>

        <div class="flex flex-col gap-3 sm:flex-row sm:items-center">
          <button type="button" class="btn btn-secondary" @click="handleClose">
            {{ t('common.cancel') }}
          </button>
          <button type="button" class="btn btn-primary" :disabled="submitting || !canSubmit" @click="submit">
            {{ submitting ? t('common.loading') : t('invoice.submit') }}
          </button>
        </div>
      </div>
    </template>
  </Modal>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Modal from '@/components/common/Modal.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Pagination from '@/components/common/Pagination.vue'
import { useAppStore } from '@/stores'
import { invoiceAPI, type InvoiceBuyerType, type InvoiceRequest, type InvoiceType } from '@/api/invoices'
import type { PaymentOrder } from '@/api/payment'

const { t } = useI18n()
const appStore = useAppStore()

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'created', invoice: InvoiceRequest): void
}>()

const mergeLimit = 5

const loading = ref(false)
const submitting = ref(false)
const unavailable = ref(false)

const eligibleOrders = ref<PaymentOrder[]>([])
const eligibleTotal = ref(0)
const eligiblePage = ref(1)
const eligiblePageSize = ref(20)
const loadingEligible = ref(false)

const selectedOrderNos = ref<string[]>([])

const form = reactive({
  invoiceType: 'normal' as InvoiceType,
  buyerType: 'company' as InvoiceBuyerType,
  invoiceTitle: '',
  taxNo: '',
  buyerAddress: '',
  buyerPhone: '',
  buyerBankName: '',
  buyerBankAccount: '',
  receiverEmail: '',
  receiverPhone: '',
  invoiceItemName: '',
  remark: ''
})

watch(
  () => props.show,
  async (value) => {
    if (!value) {
      resetState()
      return
    }
    await bootstrap()
  }
)

watch(
  () => eligiblePage.value,
  async () => {
    if (!props.show || unavailable.value) return
    await loadEligibleOrders()
  }
)

watch(
  () => form.invoiceType,
  (value) => {
    if (value === 'special') {
      form.buyerType = 'company'
    }
  }
)

const selectedOrders = computed(() => {
  const set = new Set(selectedOrderNos.value)
  return eligibleOrders.value.filter((o) => set.has(o.order_no))
})

const selectedAmountCNY = computed(() => selectedOrders.value.reduce((sum, o) => sum + (o.amount_cny || 0), 0))
const selectedTotalUSD = computed(() => selectedOrders.value.reduce((sum, o) => sum + (o.total_usd || 0), 0))

const canSubmit = computed(() => {
  if (selectedOrderNos.value.length === 0) return false
  if (selectedOrderNos.value.length > mergeLimit) return false
  if (!form.invoiceTitle.trim()) return false
  if (!form.receiverEmail.trim() || !form.receiverEmail.includes('@')) return false
  if (form.buyerType === 'company' && !form.taxNo.trim()) return false
  if (form.invoiceType === 'special') {
    if (!form.buyerAddress.trim()) return false
    if (!form.buyerPhone.trim()) return false
    if (!form.buyerBankName.trim()) return false
    if (!form.buyerBankAccount.trim()) return false
  }
  return true
})

function isNotFoundError(error: unknown): boolean {
  const status = (error as { status?: number }).status
  return status === 404
}

function paymentStatusLabel(status: string): string {
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
    case 'refunded':
      return t('payment.statusRefunded')
    default:
      return status
  }
}

function isSelected(orderNo: string): boolean {
  return selectedOrderNos.value.includes(orderNo)
}

function isSelectionDisabled(orderNo: string): boolean {
  if (isSelected(orderNo)) return false
  return selectedOrderNos.value.length >= mergeLimit
}

function toggleOrder(orderNo: string) {
  const idx = selectedOrderNos.value.indexOf(orderNo)
  if (idx >= 0) {
    selectedOrderNos.value.splice(idx, 1)
    return
  }
  if (selectedOrderNos.value.length >= mergeLimit) {
    appStore.showError(t('invoice.mergeLimitExceeded', { n: mergeLimit }))
    return
  }
  selectedOrderNos.value.push(orderNo)
}

function resetState() {
  loading.value = false
  submitting.value = false
  unavailable.value = false

  eligibleOrders.value = []
  eligibleTotal.value = 0
  eligiblePage.value = 1

  selectedOrderNos.value = []

  form.invoiceType = 'normal'
  form.buyerType = 'company'
  form.invoiceTitle = ''
  form.taxNo = ''
  form.buyerAddress = ''
  form.buyerPhone = ''
  form.buyerBankName = ''
  form.buyerBankAccount = ''
  form.receiverEmail = ''
  form.receiverPhone = ''
  form.invoiceItemName = ''
  form.remark = ''
}

async function bootstrap() {
  unavailable.value = false
  loading.value = true
  eligiblePage.value = 1

  try {
    await Promise.all([loadProfile(), loadEligibleOrders()])
  } catch (err) {
    if (isNotFoundError(err)) {
      unavailable.value = true
    } else {
      appStore.showError(String((err as { message?: string })?.message || t('common.error')))
    }
  } finally {
    loading.value = false
  }
}

async function loadProfile() {
  try {
    const profile = await invoiceAPI.getInvoiceProfile()
    // User-side invoice requests currently only support normal e-invoice.
    form.invoiceType = 'normal'
    form.buyerType = profile.buyer_type
    form.invoiceTitle = profile.invoice_title
    form.taxNo = profile.tax_no
    form.buyerAddress = profile.buyer_address
    form.buyerPhone = profile.buyer_phone
    form.buyerBankName = profile.buyer_bank_name
    form.buyerBankAccount = profile.buyer_bank_account
    form.receiverEmail = profile.receiver_email
    form.receiverPhone = profile.receiver_phone
    form.invoiceItemName = profile.invoice_item_name
    form.remark = profile.remark
  } catch (err) {
    // If profile endpoint is not available, treat as feature disabled.
    if (isNotFoundError(err)) {
      unavailable.value = true
      return
    }
    // Otherwise keep defaults but surface a lightweight warning.
    appStore.showError(String((err as { message?: string })?.message || t('invoice.profileLoadFailed')))
  }
}

async function loadEligibleOrders() {
  loadingEligible.value = true
  try {
    const resp = await invoiceAPI.getEligibleInvoiceOrders({
      page: eligiblePage.value,
      page_size: eligiblePageSize.value
    })
    eligibleOrders.value = resp.items || []
    eligibleTotal.value = resp.total || 0

    // If selected order is not in current page, keep it in selection (submit uses order_no list).
    // But if the order is no longer eligible, backend will reject and frontend will show the message.
  } catch (err) {
    if (isNotFoundError(err)) {
      unavailable.value = true
      eligibleOrders.value = []
      eligibleTotal.value = 0
      return
    }
    appStore.showError(String((err as { message?: string })?.message || t('invoice.eligibleLoadFailed')))
  } finally {
    loadingEligible.value = false
  }
}

async function submit() {
  if (!canSubmit.value) return
  submitting.value = true
  try {
    const created = await invoiceAPI.createInvoiceRequest({
      order_nos: selectedOrderNos.value,
      invoice_type: form.invoiceType,
      buyer_type: form.buyerType,
      invoice_title: form.invoiceTitle.trim(),
      tax_no: form.taxNo.trim() || undefined,
      buyer_address: form.buyerAddress.trim() || undefined,
      buyer_phone: form.buyerPhone.trim() || undefined,
      buyer_bank_name: form.buyerBankName.trim() || undefined,
      buyer_bank_account: form.buyerBankAccount.trim() || undefined,
      receiver_email: form.receiverEmail.trim(),
      receiver_phone: form.receiverPhone.trim() || undefined,
      invoice_item_name: form.invoiceItemName.trim() || undefined,
      remark: form.remark.trim() || undefined
    })

    appStore.showSuccess(t('invoice.submitSuccess'))
    emit('created', created)
    handleClose()
  } catch (err) {
    appStore.showError(String((err as { message?: string })?.message || t('common.error')))
  } finally {
    submitting.value = false
  }
}

function handleClose() {
  emit('close')
}
</script>
