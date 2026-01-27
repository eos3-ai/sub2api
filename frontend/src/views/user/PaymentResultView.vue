<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="card p-6">
        <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div class="space-y-1">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('payment.resultTitle') }}
            </h2>
            <p class="text-sm text-gray-500 dark:text-dark-400">
              {{ t('payment.resultDescription') }}
            </p>
          </div>

          <div class="flex gap-2">
            <button class="btn btn-secondary" @click="goPayment">
              {{ t('payment.backToPayment') }}
            </button>
            <button class="btn btn-primary" :disabled="loading" @click="refresh">
              {{ loading ? t('common.loading') : t('common.refresh') }}
            </button>
          </div>
        </div>
      </div>

      <div class="card p-6">
        <div v-if="!orderNo" class="text-sm text-gray-600 dark:text-dark-300">
          {{ t('payment.resultMissingOrder') }}
        </div>

        <div v-else class="space-y-4">
          <div class="grid gap-3 md:grid-cols-2">
            <div class="rounded-xl border border-gray-100 bg-white p-4 dark:border-dark-800 dark:bg-dark-900">
              <div class="text-xs font-semibold text-gray-500 dark:text-dark-400">
                {{ t('payment.orderNo') }}
              </div>
              <div class="mt-1 font-mono text-sm text-gray-900 dark:text-white">
                {{ orderNo }}
              </div>
            </div>

            <div class="rounded-xl border border-gray-100 bg-white p-4 dark:border-dark-800 dark:bg-dark-900">
              <div class="text-xs font-semibold text-gray-500 dark:text-dark-400">
                {{ t('payment.status') }}
              </div>
              <div class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">
                {{ statusLabel(order?.status || initialStatus) }}
              </div>
            </div>
          </div>

          <div v-if="order" class="grid gap-3 md:grid-cols-2">
            <div class="rounded-xl border border-gray-100 bg-white p-4 dark:border-dark-800 dark:bg-dark-900">
              <div class="text-xs font-semibold text-gray-500 dark:text-dark-400">
                {{ t('payment.creditsAmount') }}
              </div>
              <div class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
                ${{ Number(order.total_usd || 0).toFixed(2) }}
              </div>
            </div>

            <div class="rounded-xl border border-gray-100 bg-white p-4 dark:border-dark-800 dark:bg-dark-900">
              <div class="text-xs font-semibold text-gray-500 dark:text-dark-400">
                {{ t('payment.payAmountCny') }}
              </div>
              <div class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
                ¥{{ Number(order.amount_cny || 0).toFixed(2) }}
              </div>
            </div>
          </div>

          <div v-if="errorMessage" class="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700 dark:border-red-900/40 dark:bg-red-950/40 dark:text-red-200">
            {{ errorMessage }}
          </div>

          <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
            <div class="text-sm text-gray-600 dark:text-dark-300">
              <template v-if="order?.status === 'paid'">
                {{ t('payment.resultPaidHint') }}
              </template>
              <template v-else-if="order?.status === 'failed'">
                {{ t('payment.resultFailedHint') }}
              </template>
              <template v-else-if="order?.status === 'expired'">
                {{ t('payment.resultExpiredHint') }}
              </template>
              <template v-else>
                {{ t('payment.resultPendingHint') }}
              </template>
            </div>

            <button class="btn btn-secondary" @click="goPayment">
              {{ t('payment.viewOrders') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { paymentAPI } from '@/api/payment'
import { useAppStore } from '@/stores'
import type { PaymentOrder } from '@/api/payment'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()

const loading = ref(false)
const order = ref<PaymentOrder | null>(null)
const errorMessage = ref('')

const orderNo = computed(() => String(route.query.order || '').trim())
const initialStatus = computed(() => String(route.query.status || '').trim())

function statusLabel(status: string): string {
  const normalized = String(status || '').toLowerCase()
  switch (normalized) {
    case 'pending':
      return t('payment.statusPending')
    case 'paid':
      return t('payment.statusPaid')
    case 'refunded':
      return t('payment.statusRefunded')
    case 'failed':
      return t('payment.statusFailed')
    case 'expired':
      return t('payment.statusExpired')
    case 'cancelled':
    case 'canceled':
      return t('payment.statusCancelled')
    case 'success':
      return t('payment.statusPaid')
    case 'cancel':
      return t('payment.statusCancelled')
    default:
      return status || t('payment.statusPending')
  }
}

function goPayment() {
  router.push('/payment')
}

async function refresh() {
  if (!orderNo.value) return
  loading.value = true
  errorMessage.value = ''
  try {
    order.value = await paymentAPI.getPaymentOrder(orderNo.value)
  } catch (error) {
    const message = (error as { message?: string }).message || t('common.error')
    errorMessage.value = message
    appStore.showError(message)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  if (!orderNo.value) return
  refresh()
})
</script>
