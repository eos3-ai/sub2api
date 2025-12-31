<template>
  <div
    v-if="visible"
    class="rounded-2xl border border-primary-200 bg-primary-50 p-5 dark:border-primary-500/30 dark:bg-primary-900/10"
  >
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div class="space-y-1">
        <div class="flex items-center gap-2">
          <div class="inline-flex h-8 w-8 items-center justify-center rounded-xl bg-primary-600 text-white">
            %
          </div>
          <h3 class="text-base font-semibold text-gray-900 dark:text-white">
            {{ t('promotion.title') }}
          </h3>
        </div>

        <p class="text-sm text-gray-600 dark:text-dark-300">
          {{ t('promotion.subtitle', { percent: currentPercent }) }}
          <span v-if="remainingText" class="ml-1 text-gray-500 dark:text-dark-400">
            · {{ remainingText }}
          </span>
        </p>
      </div>

      <div class="flex gap-2">
        <button class="btn btn-secondary" :disabled="loading" @click="refresh">
          {{ loading ? t('common.loading') : t('common.refresh') }}
        </button>
        <button class="btn btn-primary" @click="goPayment">
          {{ t('promotion.cta') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { promotionAPI } from '@/api/promotion'
import type { PromotionStatusResponse } from '@/types'

const { t } = useI18n()
const router = useRouter()

const loading = ref(false)
const status = ref<PromotionStatusResponse | null>(null)
const fetchedAt = ref(Date.now())
const tickNow = ref(Date.now())
let ticker: number | undefined

const visible = computed(() => {
  if (!status.value?.enabled) return false
  if (status.value.status !== 'active') return false
  return !!status.value.current_tier && (status.value.remaining_seconds ?? 0) > 0
})

const currentPercent = computed(() => Number(status.value?.current_tier?.bonus_percent ?? 0).toFixed(0))

const remainingText = computed(() => {
  const remainingSeconds = Number(status.value?.remaining_seconds ?? 0)
  if (!remainingSeconds) return ''
  const elapsedSeconds = Math.max(0, Math.floor((tickNow.value - fetchedAt.value) / 1000))
  const left = Math.max(0, remainingSeconds - elapsedSeconds)
  const hours = Math.floor(left / 3600)
  const minutes = Math.floor((left % 3600) / 60)
  if (hours > 0) return t('promotion.remainingH', { h: hours, m: minutes })
  return t('promotion.remainingM', { m: minutes })
})

function goPayment() {
  router.push('/payment')
}

async function refresh() {
  loading.value = true
  try {
    status.value = await promotionAPI.getPromotionStatus()
    fetchedAt.value = Date.now()
    tickNow.value = fetchedAt.value
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  refresh()
  ticker = window.setInterval(() => {
    tickNow.value = Date.now()
  }, 1000)
})

onBeforeUnmount(() => {
  if (ticker) window.clearInterval(ticker)
})
</script>
