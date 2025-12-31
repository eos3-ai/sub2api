<template>
  <div
    v-if="visible"
    class="relative overflow-hidden rounded-3xl border border-primary-200/70 bg-gradient-to-r from-primary-100 via-white to-primary-200 p-6 shadow-card dark:border-primary-500/30 dark:from-primary-900/40 dark:via-dark-900 dark:to-primary-900/10"
  >
    <div
      class="pointer-events-none absolute -left-20 -top-20 h-56 w-56 rounded-full bg-primary-400/20 blur-3xl dark:bg-primary-400/10"
    />
    <div
      class="pointer-events-none absolute -bottom-24 -right-24 h-72 w-72 rounded-full bg-cyan-400/20 blur-3xl dark:bg-cyan-400/10"
    />

    <div class="relative flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <div class="flex items-center gap-3">
        <div class="flex h-10 w-10 items-center justify-center rounded-2xl bg-primary-600 text-white shadow-sm">
          <svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path
              d="M13.6 2.2c.4 2.1-.2 3.7-1.1 4.9-.8 1-1.7 1.8-2.5 2.6-1.3 1.3-2.4 2.4-2.4 4.3 0 2.5 2 4.5 4.5 4.5 2.8 0 5-2.2 5-5 0-1.6-.7-2.8-1.5-4-.5-.8-1-1.5-1.3-2.3 1.8 1 4.7 3.1 4.7 7.2 0 4.3-3.5 7.8-7.8 7.8S4.4 18.7 4.4 14.4c0-3.2 1.8-5.1 3.5-6.8 1.3-1.3 2.5-2.5 2.7-4.7.1-1.1 0-2-.2-2.7 1.2.4 2.4 1.1 3.2 2z"
              fill="currentColor"
            />
          </svg>
        </div>

        <div class="space-y-0.5">
          <div class="flex flex-wrap items-center gap-2">
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">
              {{ t('promotion.firstRecharge.title') }}
            </h3>
            <span
              class="inline-flex items-center rounded-full bg-white/60 px-2 py-0.5 text-xs font-semibold text-primary-700 ring-1 ring-primary-200/70 backdrop-blur dark:bg-primary-900/30 dark:text-primary-200 dark:ring-primary-500/30"
            >
              {{ t('promotion.firstRecharge.onlyOnce') }}
            </span>
          </div>
          <p class="text-sm text-gray-600 dark:text-dark-300">
            {{ t('promotion.firstRecharge.subtitle') }}
          </p>
        </div>
      </div>

      <div class="flex gap-2">
        <button class="btn btn-secondary" :disabled="loading" @click="refresh">
          {{ loading ? t('common.loading') : t('common.refresh') }}
        </button>
        <button class="btn btn-primary" @click="onCta">
          <span class="inline-flex items-center gap-2">
            <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" aria-hidden="true">
              <path
                d="M3.75 7.5A2.25 2.25 0 0 1 6 5.25h12A2.25 2.25 0 0 1 20.25 7.5v9A2.25 2.25 0 0 1 18 18.75H6A2.25 2.25 0 0 1 3.75 16.5v-9Z"
                stroke="currentColor"
                stroke-width="1.8"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
              <path
                d="M3.75 9h16.5"
                stroke="currentColor"
                stroke-width="1.8"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
              <path
                d="M7.5 14.25h3"
                stroke="currentColor"
                stroke-width="1.8"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
            {{ t('promotion.firstRecharge.cta') }}
          </span>
        </button>
      </div>
    </div>

    <div class="relative mt-5 grid gap-4 lg:grid-cols-2">
      <div
        class="rounded-2xl border border-white/60 bg-white/70 p-5 shadow-sm backdrop-blur dark:border-primary-500/20 dark:bg-dark-900/40"
      >
        <p class="text-sm font-semibold text-primary-700 dark:text-primary-200">
          {{ t('promotion.firstRecharge.badge') }}
        </p>
        <p class="mt-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white">
          {{ t('promotion.firstRecharge.headline') }}
        </p>
        <ul class="mt-4 space-y-2 text-sm text-gray-700 dark:text-dark-200">
          <li class="flex gap-2">
            <span class="mt-1 h-1.5 w-1.5 flex-none rounded-full bg-primary-600" />
            <span>{{ ruleLine1 }}</span>
          </li>
          <li class="flex gap-2">
            <span class="mt-1 h-1.5 w-1.5 flex-none rounded-full bg-primary-600" />
            <span>{{ ruleLine2 }}</span>
          </li>
          <li class="flex gap-2">
            <span class="mt-1 h-1.5 w-1.5 flex-none rounded-full bg-primary-600" />
            <span>{{ t('promotion.firstRecharge.rule3') }}</span>
          </li>
        </ul>
      </div>

      <div
        class="rounded-2xl border border-white/60 bg-white/70 p-5 shadow-sm backdrop-blur dark:border-primary-500/20 dark:bg-dark-900/40"
      >
        <div class="flex items-start justify-between gap-4">
          <div class="space-y-1">
            <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('promotion.firstRecharge.currentTitle') }}</p>
            <p class="text-4xl font-extrabold tracking-tight text-primary-700 dark:text-primary-200">
              {{ currentPercent }}%
            </p>
            <p class="text-sm text-gray-700 dark:text-dark-200">
              {{ t('promotion.firstRecharge.currentRange', { range: currentRangeText }) }}
            </p>
            <p v-if="nextTierText" class="mt-2 text-xs text-gray-600 dark:text-dark-300">
              {{ nextTierText }}
            </p>
          </div>

          <div
            class="flex min-w-[9.5rem] flex-col items-center justify-center rounded-2xl bg-primary-600 px-4 py-3 text-white shadow-sm"
          >
            <div class="flex items-center gap-2 text-xs font-semibold opacity-95">
              <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                <path
                  d="M12 6v6l4 2"
                  stroke="currentColor"
                  stroke-width="1.8"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
                <path
                  d="M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"
                  stroke="currentColor"
                  stroke-width="1.8"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
              </svg>
              {{ t('promotion.firstRecharge.countdown') }}
            </div>
            <div class="mt-1 text-2xl font-bold tabular-nums">{{ tierCountdown }}</div>
          </div>
        </div>
      </div>
    </div>

    <div class="relative mt-5 grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
      <div
        v-for="(tier, index) in tiers"
        :key="`${tier.hours}-${tier.bonus_percent}-${index}`"
        class="relative overflow-hidden rounded-2xl border p-4 shadow-sm"
        :class="tierCardClass(index)"
      >
        <div class="flex items-center justify-between gap-3">
          <div class="flex items-center gap-2">
            <span class="inline-flex h-9 w-9 items-center justify-center rounded-xl bg-primary-600/10 text-primary-700 dark:bg-primary-500/15 dark:text-primary-200">
              <svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                <path
                  d="M12 6v12m-4-9h8m-8 6h8"
                  stroke="currentColor"
                  stroke-width="1.8"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
                <path
                  d="M7.5 4.5h9A2.25 2.25 0 0 1 18.75 6.75v10.5A2.25 2.25 0 0 1 16.5 19.5h-9A2.25 2.25 0 0 1 5.25 17.25V6.75A2.25 2.25 0 0 1 7.5 4.5Z"
                  stroke="currentColor"
                  stroke-width="1.8"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
              </svg>
            </span>
            <div>
              <p class="text-base font-semibold text-gray-900 dark:text-white">
                {{ t('promotion.firstRecharge.tierLabel', { hours: tier.hours }) }}
              </p>
              <p class="text-sm text-gray-600 dark:text-dark-300">
                {{ t('promotion.firstRecharge.tierBonus', { percent: tier.bonus_percent }) }}
              </p>
            </div>
          </div>

          <span
            v-if="currentTierIndex === index"
            class="inline-flex items-center rounded-full bg-primary-600 px-2 py-0.5 text-xs font-semibold text-white shadow-sm"
          >
            {{ t('promotion.firstRecharge.currentTag') }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter, useRoute } from 'vue-router'
import { promotionAPI } from '@/api/promotion'
import type { PromotionStatusResponse } from '@/types'

const props = defineProps<{
  ctaTo?: string
}>()

const { t } = useI18n()
const router = useRouter()
const route = useRoute()

const loading = ref(false)
const status = ref<PromotionStatusResponse | null>(null)
const fetchedAt = ref(Date.now())
const tickNow = ref(Date.now())
let ticker: number | undefined

const tiers = computed(() => status.value?.tiers ?? [])
const currentTierIndex = computed(() => status.value?.current_tier?.tier ?? -1)
const currentPercent = computed(() => Number(status.value?.current_tier?.bonus_percent ?? 0).toFixed(0))

const visible = computed(() => {
  if (!status.value?.enabled) return false
  if (status.value.status !== 'active') return false
  if (!status.value.current_tier) return false
  return (status.value.remaining_seconds ?? 0) > 0
})

const durationHours = computed(() => {
  const v = status.value?.duration_hours
  if (typeof v === 'number' && v > 0) return v
  const last = tiers.value[tiers.value.length - 1]?.hours
  return typeof last === 'number' && last > 0 ? last : 72
})

const ruleLine1 = computed(() => t('promotion.firstRecharge.rule1', { hours: durationHours.value }))
const ruleLine2 = computed(() =>
  t('promotion.firstRecharge.rule2', { n: Math.max(0, tiers.value.length) })
)

const currentRangeText = computed(() => {
  const tier = status.value?.current_tier
  if (!tier) return ''
  const upper = tier.hours
  const lower = tiers.value[tier.tier - 1]?.hours ?? 0
  if (lower <= 0) return `0-${upper}${t('promotion.firstRecharge.hourSuffix')}`
  return `${lower}-${upper}${t('promotion.firstRecharge.hourSuffix')}`
})

function formatDuration(sec: number): string {
  const s = Math.max(0, Math.floor(sec))
  const h = Math.floor(s / 3600)
  const m = Math.floor((s % 3600) / 60)
  const ss = s % 60
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${pad(h)}:${pad(m)}:${pad(ss)}`
}

const tierCountdown = computed(() => {
  const remainingSeconds = Number(status.value?.current_tier_remaining_seconds ?? 0)
  if (!remainingSeconds) return '--:--:--'
  const elapsedSeconds = Math.max(0, Math.floor((tickNow.value - fetchedAt.value) / 1000))
  const left = Math.max(0, remainingSeconds - elapsedSeconds)
  return formatDuration(left)
})

const nextTierText = computed(() => {
  const idx = currentTierIndex.value
  if (idx < 0) return ''
  if (idx >= tiers.value.length - 1) return ''
  const remainingSeconds = Number(status.value?.current_tier_remaining_seconds ?? 0)
  if (!remainingSeconds) return ''
  return t('promotion.firstRecharge.nextTierHint', { time: tierCountdown.value })
})

function tierCardClass(index: number): string {
  if (currentTierIndex.value === index) {
    return 'border-primary-400 bg-primary-50 ring-2 ring-primary-500/20 dark:border-primary-400/60 dark:bg-primary-900/20 dark:ring-primary-400/20'
  }
  return 'border-white/60 bg-white/60 dark:border-primary-500/20 dark:bg-dark-900/30'
}

async function refresh() {
  loading.value = true
  try {
    status.value = await promotionAPI.getPromotionStatus()
    fetchedAt.value = Date.now()
    tickNow.value = fetchedAt.value
  } catch {
    status.value = { enabled: false, status: 'none' }
  } finally {
    loading.value = false
  }
}

function onCta() {
  const to = props.ctaTo || '/payment'
  if (to.startsWith('#')) {
    const el = document.querySelector(to)
    if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' })
    return
  }
  if (route.path === to) return
  router.push(to)
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
