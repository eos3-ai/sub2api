<template>
  <AppLayout>
    <div class="space-y-6">
      <FirstRechargePromotion cta-to="#recharge-plans" />

      <!-- Plans -->
      <div id="recharge-plans" class="card animate-fade-in-up p-6 stagger-1">
        <div class="mb-6 rounded-2xl bg-primary-600 px-6 py-4 text-white">
          <div class="flex items-center gap-3">
            <div class="flex h-9 w-9 items-center justify-center rounded-xl bg-white/15">
              <span class="text-lg font-bold">$</span>
            </div>
            <div class="text-lg font-semibold">{{ t('payment.onlineRecharge') }}</div>
          </div>
        </div>

        <div v-if="loadingPlans" class="flex items-center justify-center py-10">
          <LoadingSpinner />
        </div>

        <template v-else>
          <div
            v-if="plansUnavailable"
            class="rounded-xl border border-amber-200 bg-amber-50 p-4 text-amber-900 dark:border-amber-900/30 dark:bg-amber-900/10 dark:text-amber-200"
          >
            <p class="text-sm font-medium">{{ t('payment.apiNotEnabledTitle') }}</p>
            <p class="mt-1 text-xs opacity-90">
              {{ t('payment.apiNotEnabledDesc') }}
            </p>
          </div>

          <div v-else-if="plans.length === 0" class="py-10 text-center text-sm text-gray-500 dark:text-dark-400">
            {{ t('payment.noPlans') }}
          </div>

          <div v-else class="space-y-4">
            <div class="space-y-3">
              <div class="text-sm font-semibold text-gray-900 dark:text-white">
                {{ t('payment.quickSelectAmount') }}
              </div>

              <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
              <button
                v-for="plan in plans"
                :key="plan.id"
                type="button"
                class="relative rounded-2xl border p-6 text-left shadow-sm transition hover:shadow-md"
                :class="
                  selectedKind === 'plan' && selectedPlan?.id === plan.id
                    ? 'border-primary-500 bg-primary-50 ring-2 ring-primary-500/20 dark:border-primary-400 dark:bg-primary-900/10 dark:ring-primary-400/20'
                    : 'border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800'
                "
                @click="selectPlan(plan, true)"
              >
                <span
                  v-if="selectedKind === 'plan' && selectedPlan?.id === plan.id"
                  class="absolute right-4 top-4 inline-flex h-7 w-7 items-center justify-center rounded-full bg-primary-600 text-sm font-bold text-white shadow-sm"
                  aria-label="selected"
                >
                  ✓
                </span>

                <p class="text-3xl font-bold text-gray-900 dark:text-white">{{ formatUSDCompact(plan.amount_usd) }}</p>
                <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
                  {{ t('payment.estimatedPay') }} ¥{{ estimatePayCNY(plan.amount_usd, plan.exchange_rate, plan.discount_rate).toFixed(2) }}
                </p>
              </button>

              <button
                type="button"
                class="relative rounded-2xl border p-6 text-left shadow-sm transition hover:shadow-md"
                :class="
                  selectedKind === 'custom'
                    ? 'border-primary-500 bg-primary-50 ring-2 ring-primary-500/20 dark:border-primary-400 dark:bg-primary-900/10 dark:ring-primary-400/20'
                    : 'border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800'
                "
                @click="selectCustom(true)"
              >
                <span
                  v-if="selectedKind === 'custom'"
                  class="absolute right-4 top-4 inline-flex h-7 w-7 items-center justify-center rounded-full bg-primary-600 text-sm font-bold text-white shadow-sm"
                  aria-label="selected"
                >
                  ✓
                </span>

                <p class="text-3xl font-bold text-gray-900 dark:text-white">
                  {{ customAmountUSD > 0 ? formatUSDCompact(customAmountUSD) : t('payment.other') }}
                </p>
                <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
                  {{ t('payment.customAmountSubtitle') }}
                </p>
              </button>
              </div>
            </div>

            <div class="space-y-3 pt-2">
              <div class="text-sm font-semibold text-gray-900 dark:text-white">
                {{ t('payment.choosePayMethod') }}
              </div>
              <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <button
                  type="button"
                  class="flex items-center gap-3 rounded-2xl border px-6 py-4 text-left text-base font-semibold transition"
                  :class="
                    payMethod === 'alipay'
                      ? 'border-primary-500 bg-primary-50 text-gray-900 dark:border-primary-400 dark:bg-primary-900/10 dark:text-white'
                      : 'border-gray-200 bg-white text-gray-900 hover:bg-gray-50 dark:border-dark-700 dark:bg-dark-800 dark:text-white dark:hover:bg-dark-700'
                  "
                  @click="payMethod = 'alipay'"
                >
                  <span class="flex h-10 w-10 items-center justify-center rounded-2xl bg-white shadow-sm dark:bg-dark-900">
                    <svg viewBox="0 0 24 24" class="h-6 w-6 text-primary-600 dark:text-primary-400" aria-hidden="true">
                      <path
                        fill="currentColor"
                        d="M12 2a10 10 0 100 20 10 10 0 000-20zm4.4 6.1l-2.2 7.7h-1.8l-.9-2.7H8.8l-.9 2.7H6.1l2.7-7.7h1.8l.7 2.2h2.4l.7-2.2h1.8zM9.3 11.1h2.4l-1.2-3.5-1.2 3.5z"
                      />
                    </svg>
                  </span>
                  <span>{{ t('payment.alipay') }}</span>
                </button>

                <button
                  type="button"
                  class="flex items-center gap-3 rounded-2xl border px-6 py-4 text-left text-base font-semibold transition"
                  :class="
                    payMethod === 'wechat'
                      ? 'border-primary-500 bg-primary-50 text-gray-900 dark:border-primary-400 dark:bg-primary-900/10 dark:text-white'
                      : 'border-gray-200 bg-white text-gray-900 hover:bg-gray-50 dark:border-dark-700 dark:bg-dark-800 dark:text-white dark:hover:bg-dark-700'
                  "
                  @click="payMethod = 'wechat'"
                >
                  <span class="flex h-10 w-10 items-center justify-center rounded-2xl bg-white shadow-sm dark:bg-dark-900">
                    <svg viewBox="0 0 24 24" class="h-6 w-6 text-emerald-600 dark:text-emerald-400" aria-hidden="true">
                      <path
                        fill="currentColor"
                        d="M8.2 4.8C4.8 4.8 2 7.1 2 10c0 1.7.9 3.2 2.3 4.2L3.6 16c-.1.2 0 .4.2.5.1.1.3.1.4 0l2-1.2c.6.2 1.3.3 2 .3 3.4 0 6.2-2.3 6.2-5.2S11.6 4.8 8.2 4.8zm-2 4.7c-.5 0-.9-.4-.9-.9s.4-.9.9-.9.9.4.9.9-.4.9-.9.9zm4 0c-.5 0-.9-.4-.9-.9s.4-.9.9-.9.9.4.9.9-.4.9-.9.9zM21.9 13.3c0-2.4-2.4-4.3-5.4-4.3-3 0-5.4 1.9-5.4 4.3s2.4 4.3 5.4 4.3c.5 0 1-.1 1.5-.2l1.6.9c.2.1.4.1.6-.1.1-.1.1-.3.1-.4l-.5-1.3c1.1-.8 1.9-1.9 1.9-3.2zm-7.4.3c-.4 0-.8-.3-.8-.8s.3-.8.8-.8.8.3.8.8-.4.8-.8.8zm3.5 0c-.4 0-.8-.3-.8-.8s.3-.8.8-.8.8.3.8.8-.4.8-.8.8z"
                      />
                    </svg>
                  </span>
                  <span>{{ t('payment.wechat') }}</span>
                </button>
              </div>

              <div class="flex flex-col gap-4 pt-4 sm:flex-row sm:items-center sm:justify-between">
                <div class="text-sm text-gray-700 dark:text-dark-200">
                  <span class="font-medium">{{ t('payment.payAmountLabel') }}：</span>
                  <span class="font-semibold">
                    {{ formatUSD2(selectedAmountUSD) }}
                  </span>
                  <span class="text-gray-500 dark:text-dark-400">
                    ({{ t('payment.estimatedPay') }} ¥{{ computedPayCNY.toFixed(2) }})
                  </span>
                </div>

                <button
                  class="inline-flex items-center justify-center rounded-2xl bg-primary-600 px-10 py-4 text-base font-semibold text-white shadow-sm transition hover:bg-primary-700 disabled:cursor-not-allowed disabled:opacity-60"
                  :disabled="creatingOrder || !canPayNow"
                  @click="payNow"
                >
                  {{ creatingOrder ? t('common.loading') : t('payment.rechargeNow') }}
                </button>
              </div>
            </div>
          </div>
        </template>
      </div>

      <!-- Orders -->
      <div class="card animate-fade-in-up p-6 stagger-2">
        <div class="mb-4 flex items-center justify-between">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('payment.myOrders') }}</h2>
          <div class="flex items-center gap-3">
            <button class="btn btn-secondary" :disabled="loadingOrders" @click="openInvoice">
              {{ t('invoice.createTitle') }}
            </button>
            <button class="btn btn-secondary" :disabled="loadingOrders" @click="loadOrders">
              {{ loadingOrders ? t('common.loading') : t('common.refresh') }}
            </button>
          </div>
        </div>

        <div v-if="loadingOrders" class="flex items-center justify-center py-10">
          <LoadingSpinner />
        </div>

        <template v-else>
          <div
            v-if="ordersUnavailable"
            class="rounded-xl border border-gray-200 bg-gray-50 p-4 text-gray-900 dark:border-dark-700 dark:bg-dark-800/50 dark:text-white"
          >
            <p class="text-sm font-medium">{{ t('payment.ordersApiNotEnabledTitle') }}</p>
            <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
              {{ t('payment.ordersApiNotEnabledDesc') }}
            </p>
          </div>

          <div v-else-if="orders.length === 0" class="py-10 text-center text-sm text-gray-500 dark:text-dark-400">
            {{ t('payment.noOrders') }}
          </div>

          <div v-else class="overflow-hidden rounded-2xl border border-gray-200 dark:border-dark-700">
            <div class="table-wrapper overflow-x-auto">
              <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
                <thead class="bg-gray-50 dark:bg-dark-800/60">
                  <tr>
                    <th class="px-5 py-4 text-left text-xs font-semibold text-gray-600 dark:text-dark-300">
                      {{ t('payment.orderNo') }}
                    </th>
                    <th class="px-5 py-4 text-left text-xs font-semibold text-gray-600 dark:text-dark-300">
                      {{ t('payment.orderType') }}
                    </th>
                    <th class="px-5 py-4 text-left text-xs font-semibold text-gray-600 dark:text-dark-300">
                      {{ t('payment.remark') }}
                    </th>
                    <th class="px-5 py-4 text-left text-xs font-semibold text-gray-600 dark:text-dark-300">
                      {{ t('payment.channel') }}
                    </th>
                    <th class="px-5 py-4 text-left text-xs font-semibold text-gray-600 dark:text-dark-300">
                      {{ t('payment.creditsAmount') }}
                    </th>
                    <th class="px-5 py-4 text-left text-xs font-semibold text-gray-600 dark:text-dark-300">
                      {{ t('payment.payAmountCny') }}
                    </th>
                    <th class="px-5 py-4 text-left text-xs font-semibold text-gray-600 dark:text-dark-300">
                      {{ t('payment.status') }}
                    </th>
                    <th class="px-5 py-4 text-right text-xs font-semibold text-gray-600 dark:text-dark-300">
                      {{ t('common.actions') }}
                    </th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-100 bg-white dark:divide-dark-800 dark:bg-dark-800">
                  <tr v-for="o in orders" :key="o.order_no">
                    <td class="px-5 py-4 text-sm text-gray-900 dark:text-white">
                      <span class="font-mono text-xs">{{ o.order_no }}</span>
                    </td>
                    <td class="px-5 py-4 text-sm text-gray-700 dark:text-dark-300">
                      {{ orderTypeLabel(o.order_type) }}
                    </td>
                    <td class="px-5 py-4 text-sm text-gray-700 dark:text-dark-300">
                      {{ o.remark || '-' }}
                    </td>
                    <td class="px-5 py-4 text-sm text-gray-700 dark:text-dark-300">
                      {{ shouldShowChannel(o.order_type) ? channelLabel(o.channel || o.provider) : '-' }}
                    </td>
                    <td class="px-5 py-4 text-sm text-gray-700 dark:text-dark-300">
                      ${{ o.total_usd.toFixed(2) }}
                    </td>
                    <td class="px-5 py-4 text-sm text-gray-700 dark:text-dark-300">
                      {{ shouldShowPayAmount(o.order_type) ? `¥${o.amount_cny.toFixed(2)}` : '-' }}
                    </td>
                    <td class="px-5 py-4 text-sm text-gray-700 dark:text-dark-300">
                      {{ paymentStatusLabel(o.status) }}
                    </td>
                    <td class="px-5 py-4 text-right text-sm">
                      <button class="btn btn-secondary btn-sm" @click="copyOrderNo(o.order_no)">
                        {{ t('payment.copyOrder') }}
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </template>
      </div>
    </div>

    <!-- Checkout Modal -->
    <Modal
      :show="checkoutOpen"
      :title="t('payment.calcTitle')"
      size="lg"
      closeOnClickOutside
      @close="closeCheckout"
    >
      <template v-if="selectedKind">
        <div class="space-y-4">
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('payment.calcSubtitle') }}</p>

          <div class="rounded-3xl bg-gray-50 p-5 dark:bg-dark-900/30">
            <p class="text-sm font-medium text-gray-600 dark:text-dark-300">{{ t('payment.displayAmount') }}</p>

            <div class="mt-1 text-4xl font-bold text-gray-900 dark:text-white">
              {{ formatUSD2(selectedAmountUSD) }}
            </div>

            <div v-if="selectedKind === 'custom'" class="mt-4">
              <div class="flex items-center gap-2">
                <span class="text-lg font-semibold text-gray-700 dark:text-dark-200">$</span>
                <input
                  v-model="customAmountUSDInput"
                  inputmode="decimal"
                  class="w-full rounded-2xl border border-gray-200 bg-white px-4 py-3 text-sm text-gray-900 shadow-sm outline-none transition focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-dark-700 dark:bg-dark-900 dark:text-white"
                  :placeholder="t('payment.customAmountPlaceholder')"
                  @keydown.enter.prevent
                />
              </div>
              <p v-if="customAmountUSD <= 0" class="mt-2 text-xs text-rose-600 dark:text-rose-400">
                {{ t('payment.invalidAmount') }}
              </p>
            </div>

            <div class="mt-5 rounded-2xl bg-white p-4 shadow-sm dark:bg-dark-800">
              <p class="text-sm font-medium text-gray-600 dark:text-dark-300">{{ t('payment.formula') }}</p>
              <p class="mt-2 font-mono text-sm text-gray-900 dark:text-white">
                {{ formulaText }}
              </p>
            </div>

            <div class="mt-5 space-y-3">
              <div class="flex items-center justify-between text-sm text-gray-700 dark:text-dark-200">
                <span>{{ t('payment.exchangeRate') }}</span>
                <span class="font-semibold">{{ exchangeRate.toFixed(2) }}</span>
              </div>
              <div class="flex items-center justify-between text-sm text-gray-700 dark:text-dark-200">
                <span>{{ t('payment.discountRate') }}</span>
                <span class="font-semibold">{{ discountRate.toFixed(4) }}</span>
              </div>
              <div class="pt-2">
                <p class="text-sm font-medium text-gray-700 dark:text-dark-200">{{ t('payment.finalPay') }}</p>
                <div class="mt-1 text-4xl font-bold text-emerald-600 dark:text-emerald-400">
                  ¥{{ computedPayCNY.toFixed(2) }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>

      <template #footer>
        <button
          class="w-full rounded-2xl bg-primary-600 px-6 py-4 text-base font-semibold text-white shadow-sm transition hover:bg-primary-700 disabled:cursor-not-allowed disabled:opacity-60"
          :disabled="selectedKind === 'custom' && customAmountUSD <= 0"
          @click="closeCheckout"
        >
          {{ t('payment.confirmAmount') }}
        </button>
      </template>
    </Modal>

    <!-- Pay Modal -->
    <Modal
      :show="payOpen"
      :title="t('payment.payTitle')"
      size="lg"
      closeOnClickOutside
      @close="closePay"
    >
      <template v-if="payOrder">
        <div class="space-y-5">
          <div class="flex flex-wrap items-center justify-between gap-3 rounded-2xl bg-gray-50 px-4 py-3 dark:bg-dark-900/30">
            <div class="space-y-1">
              <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('payment.orderNo') }}</p>
              <p class="font-mono text-sm font-semibold text-gray-900 dark:text-white">
                {{ payOrder.order_no }}
              </p>
            </div>
            <div class="space-y-1 text-right">
              <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('payment.status') }}</p>
              <p
                class="text-sm font-semibold"
                :class="
                  payOrder.status === 'paid'
                    ? 'text-emerald-600 dark:text-emerald-400'
                    : payOrder.status === 'failed' || payOrder.status === 'expired'
                      ? 'text-rose-600 dark:text-rose-400'
                      : 'text-amber-600 dark:text-amber-400'
                "
              >
                {{ paymentStatusLabel(payOrder.status) }}
              </p>
            </div>
          </div>

          <div class="space-y-3">
            <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('payment.scanToPay') }}</p>
            <div
              class="flex flex-col items-center justify-center gap-4 rounded-3xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-800"
            >
              <div v-if="qrImage" class="flex flex-col items-center gap-3">
                <img :src="qrImage" alt="qr" class="h-56 w-56 rounded-2xl bg-white p-2" />
                <p v-if="polling && payOrder.status === 'pending'" class="text-xs text-gray-500 dark:text-dark-400">
                  {{ t('payment.waitingForPayment') }}
                </p>
              </div>
              <div v-else class="flex flex-col items-center gap-2 py-8 text-center">
                <p class="text-sm text-gray-600 dark:text-dark-300">
                  {{ t('payment.noQRCode') }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </template>

      <template #footer>
        <button
          class="w-full rounded-2xl bg-gray-900 px-6 py-4 text-base font-semibold text-white shadow-sm transition hover:bg-gray-800 dark:bg-dark-700 dark:hover:bg-dark-600"
          @click="closePay"
        >
          {{ t('common.close') }}
        </button>
      </template>
    </Modal>

    <InvoiceRequestModal
      :show="invoiceOpen"
      @close="invoiceOpen = false"
      @created="handleInvoiceCreated"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import FirstRechargePromotion from '@/components/FirstRechargePromotion.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Modal from '@/components/common/Modal.vue'
import InvoiceRequestModal from '@/components/user/InvoiceRequestModal.vue'
import { useAppStore } from '@/stores'
import { paymentAPI, type PaymentOrder, type PaymentPayMethod, type PaymentPlan } from '@/api/payment'
import QRCode from 'qrcode'

const { t } = useI18n()
const appStore = useAppStore()

const loadingPlans = ref(false)
const plansUnavailable = ref(false)
const plans = ref<PaymentPlan[]>([])

const loadingOrders = ref(false)
const ordersUnavailable = ref(false)
const orders = ref<PaymentOrder[]>([])

const selectedKind = ref<'plan' | 'custom' | null>(null)
const selectedPlan = ref<PaymentPlan | null>(null)
const customAmountUSDInput = ref('')
const payMethod = ref<PaymentPayMethod>('alipay')
const creatingOrder = ref(false)
const checkoutOpen = ref(false)

const payOpen = ref(false)
const payOrder = ref<PaymentOrder | null>(null)
const payURL = ref('')
const qrPayload = ref('')
const qrImage = ref('')
const polling = ref(false)
let pollTimer: number | null = null

const invoiceOpen = ref(false)

const exchangeRate = computed(() => {
  const rate = selectedPlan.value?.exchange_rate ?? plans.value[0]?.exchange_rate
  return typeof rate === 'number' && rate > 0 ? rate : 7.2
})

const discountRate = computed(() => {
  const rate = selectedPlan.value?.discount_rate ?? plans.value[0]?.discount_rate
  return typeof rate === 'number' && rate > 0 && rate <= 1 ? rate : 1
})

const customAmountUSD = computed(() => {
  const parsed = Number.parseFloat(customAmountUSDInput.value)
  return Number.isFinite(parsed) ? parsed : 0
})

const selectedAmountUSD = computed(() => {
  if (selectedKind.value === 'custom') return customAmountUSD.value
  return selectedPlan.value?.amount_usd ?? 0
})

const computedPayUSD = computed(() => {
  const usd = selectedAmountUSD.value
  const discount = discountRate.value
  if (!(usd > 0)) return 0
  if (!(discount > 0 && discount <= 1)) return usd
  return usd * discount
})

const computedPayCNY = computed(() => {
  const payUSD = computedPayUSD.value
  return payUSD > 0 ? payUSD * exchangeRate.value : 0
})

const formulaText = computed(() => {
  const usd = selectedAmountUSD.value
  const rate = exchangeRate.value
  const discount = discountRate.value
  const cny = computedPayCNY.value
  if (!(usd > 0) || !(rate > 0)) return '$0.00 × 0.00 × 0.0000 = ¥0.00'
  const usdText = formatUSD2(usd)
  const rateText = rate.toFixed(2)
  const discountText = discount.toFixed(4)
  const cnyText = cny.toFixed(2)
  return `${usdText} × ${rateText} × ${discountText} = ¥${cnyText}`
})

const canPayNow = computed(() => {
  if (!selectedKind.value) return false
  if (selectedKind.value === 'plan' && !selectedPlan.value) return false
  if (selectedKind.value === 'custom' && !(customAmountUSD.value > 0)) return false
  return true
})

function isNotFoundError(error: unknown): boolean {
  const status = (error as { status?: number }).status
  return status === 404
}

function paymentErrorMessage(error: unknown): string {
  const err = error as { status?: number; message?: string }
  const message = String(err?.message || '')
  if (err?.status === 503) return t('payment.serviceUnavailableHint')
  if (err?.status === 500 || message.toLowerCase() === 'internal error') return t('payment.internalErrorHint')
  if (message) return message
  return t('common.error')
}

function formatUSDCompact(amount: number): string {
  if (!Number.isFinite(amount) || amount <= 0) return '$0'
  if (Math.abs(amount - Math.round(amount)) < 1e-9) return `$${Math.round(amount)}`
  return `$${amount.toFixed(2)}`
}

function formatUSD2(amount: number): string {
  if (!Number.isFinite(amount) || amount <= 0) return '$0.00'
  return `$${amount.toFixed(2)}`
}

function estimatePayCNY(usd: number, rate?: number, discount?: number): number {
  const resolvedRate = typeof rate === 'number' && rate > 0 ? rate : exchangeRate.value
  const resolvedDiscount = typeof discount === 'number' && discount > 0 && discount <= 1 ? discount : discountRate.value
  if (!(usd > 0)) return 0
  if (!(resolvedRate > 0)) return 0
  if (!(resolvedDiscount > 0 && resolvedDiscount <= 1)) return usd * resolvedRate
  return usd * resolvedRate * resolvedDiscount
}

async function loadPlans() {
  loadingPlans.value = true
  try {
    plansUnavailable.value = false
    plans.value = await paymentAPI.getPaymentPlans()
  } catch (error) {
    if (isNotFoundError(error)) {
      plansUnavailable.value = true
      plans.value = []
      return
    }
    appStore.showError(paymentErrorMessage(error))
  } finally {
    loadingPlans.value = false
  }
}

function selectPlan(plan: PaymentPlan, openModal: boolean) {
  selectedKind.value = 'plan'
  selectedPlan.value = plan
  if (openModal) checkoutOpen.value = true
}

function selectCustom(openModal: boolean) {
  selectedKind.value = 'custom'
  selectedPlan.value = null
  if (openModal) checkoutOpen.value = true
}

function closeCheckout() {
  checkoutOpen.value = false
}

function closePay() {
  payOpen.value = false
}

function stopPolling() {
  polling.value = false
  if (pollTimer != null) {
    window.clearInterval(pollTimer)
    pollTimer = null
  }
}

function isLikelyImageSrc(value: string): boolean {
  const v = value.trim().toLowerCase()
  if (!v) return false
  if (v.startsWith('data:image/')) return true
  if (v.endsWith('.png') || v.endsWith('.jpg') || v.endsWith('.jpeg') || v.endsWith('.webp') || v.endsWith('.gif'))
    return true
  return false
}

async function refreshPayQR() {
  const payload = qrPayload.value.trim() || payURL.value.trim()
  if (!payload) {
    qrImage.value = ''
    return
  }
  if (isLikelyImageSrc(payload)) {
    qrImage.value = payload
    return
  }
  try {
    qrImage.value = await QRCode.toDataURL(payload, { width: 280, margin: 1 })
  } catch {
    qrImage.value = ''
  }
}

async function pollOrderStatus(orderNo: string) {
  if (!orderNo) return
  stopPolling()
  polling.value = true

  const tick = async () => {
    try {
      const latest = await paymentAPI.getPaymentOrder(orderNo)
      payOrder.value = latest

      if (latest.status === 'paid') {
        stopPolling()
        appStore.showSuccess(t('payment.paymentSuccess'))
        await loadOrders()
        setTimeout(() => {
          payOpen.value = false
        }, 1500)
        return
      }
      if (latest.status === 'failed') {
        stopPolling()
        appStore.showError(t('payment.paymentFailed'))
        await loadOrders()
        return
      }
      if (latest.status === 'expired') {
        stopPolling()
        appStore.showWarning(t('payment.paymentExpired'))
        await loadOrders()
      }
    } catch (error) {
      const status = (error as { status?: number }).status
      if (status === 404) {
        stopPolling()
      }
    }
  }

  await tick()
  pollTimer = window.setInterval(tick, 2500)
}

async function loadOrders() {
  loadingOrders.value = true
  try {
    ordersUnavailable.value = false
    const resp = await paymentAPI.getMyPaymentOrders({ page: 1, page_size: 20 })
    orders.value = resp.items
  } catch (error) {
    if (isNotFoundError(error)) {
      ordersUnavailable.value = true
      orders.value = []
      return
    }
    appStore.showError(paymentErrorMessage(error))
  } finally {
    loadingOrders.value = false
  }
}

async function payNow() {
  if (!selectedKind.value) {
    appStore.showWarning(t('payment.selectAmount'))
    return
  }
  if (selectedKind.value === 'custom' && !(customAmountUSD.value > 0)) {
    appStore.showWarning(t('payment.invalidAmount'))
    return
  }
  if (selectedKind.value === 'plan' && !selectedPlan.value) {
    appStore.showWarning(t('payment.selectAmount'))
    return
  }

  try {
    creatingOrder.value = true
    appStore.showInfo(t('payment.creatingOrder'))
    const resp =
      selectedKind.value === 'custom'
        ? await paymentAPI.createPaymentOrder({
            amount_usd: customAmountUSD.value,
            channel: payMethod.value
          })
        : await paymentAPI.createPaymentOrder({
            plan_id: selectedPlan.value!.id,
            channel: payMethod.value
          })
    appStore.showSuccess(t('payment.orderCreated') + ` (${resp.order.order_no})`)
    await loadOrders()
    closeCheckout()

    payOrder.value = resp.order
    payURL.value = resp.pay_url || ''
    qrPayload.value = resp.qr_url || resp.pay_url || ''
    payOpen.value = true
    await refreshPayQR()
    await pollOrderStatus(resp.order.order_no)
  } catch (error) {
    if (isNotFoundError(error)) {
      appStore.showWarning(t('payment.apiNotEnabledToast'))
      return
    }
    appStore.showError(paymentErrorMessage(error))
  } finally {
    creatingOrder.value = false
  }
}

watch(
  () => payOpen.value,
  async (open) => {
    if (!open) {
      stopPolling()
      return
    }
    await refreshPayQR()
  }
)

watch(
  () => qrPayload.value,
  async () => {
    if (payOpen.value) await refreshPayQR()
  }
)

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

function openInvoice() {
  invoiceOpen.value = true
}

function handleInvoiceCreated() {
  // No need to refresh order list, but keep selection UI in sync if user immediately opens again.
  invoiceOpen.value = false
}

async function copyOrderNo(orderNo: string) {
  try {
    await navigator.clipboard.writeText(orderNo)
    appStore.showSuccess(t('common.copiedToClipboard'))
  } catch {
    appStore.showError(t('common.copyFailed'))
  }
}

onMounted(() => {
  loadPlans()
  loadOrders()
})

onUnmounted(() => {
  stopPolling()
})
</script>
