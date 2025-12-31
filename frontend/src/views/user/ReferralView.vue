<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="card overflow-hidden p-0">
        <div class="bg-amber-50/60 p-6 dark:bg-dark-900">
          <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div class="space-y-2">
              <p class="text-sm font-semibold text-amber-700 dark:text-amber-300">
                {{ t('referral.page.tag') }}
              </p>
              <h2 class="text-2xl font-bold text-gray-900 dark:text-white">
                {{ t('referral.page.title') }}
              </h2>
              <p class="max-w-3xl text-sm leading-6 text-gray-600 dark:text-dark-300">
                {{ t('referral.page.description', { reward: formatUSD(info?.reward_usd ?? 0), cny: info?.qualified_recharge_cny ?? 0 }) }}
              </p>
            </div>

            <div class="flex flex-col gap-3 sm:flex-row sm:justify-end">
              <button class="btn btn-secondary" :disabled="loadingInfo || !info?.code" @click="copy(info?.code || '')">
                {{ t('referral.page.copyCode') }}
              </button>
              <button
                class="btn btn-secondary"
                :disabled="loadingInfo || !info?.invite_link"
                @click="copy(info?.invite_link || '')"
              >
                {{ t('referral.page.copyLink') }}
              </button>
            </div>
          </div>

          <div v-if="info && !info.enabled" class="mt-6 rounded-xl border border-gray-200 bg-white p-4 text-gray-900 dark:border-dark-800 dark:bg-dark-900 dark:text-white">
            <p class="text-sm font-medium">{{ t('referral.disabledTitle') }}</p>
            <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('referral.disabledDesc') }}</p>
          </div>

          <div v-else class="mt-6 grid gap-4 lg:grid-cols-2">
            <div class="rounded-2xl border border-gray-100 bg-white p-5 dark:border-dark-800 dark:bg-dark-900">
              <div class="text-xs font-semibold text-gray-500 dark:text-dark-400">
                {{ t('referral.page.inviteLinkLabel') }}
              </div>
              <div class="mt-3 rounded-xl border border-gray-100 bg-gray-50 px-4 py-3 text-sm text-gray-900 dark:border-dark-800 dark:bg-dark-800 dark:text-white">
                <span class="block truncate">{{ info?.invite_link || '-' }}</span>
              </div>
            </div>

            <div class="rounded-2xl border border-gray-100 bg-white p-5 dark:border-dark-800 dark:bg-dark-900">
              <div class="text-xs font-semibold text-gray-500 dark:text-dark-400">
                {{ t('referral.page.inviteCodeLabel') }}
              </div>
              <div class="mt-3 rounded-xl border border-gray-100 bg-gray-50 px-4 py-3 font-mono text-sm font-semibold text-gray-900 dark:border-dark-800 dark:bg-dark-800 dark:text-white">
                {{ info?.code || '-' }}
              </div>
            </div>
          </div>
        </div>

        <div v-if="info?.enabled" class="bg-white p-6 dark:bg-dark-950">
          <div class="grid gap-4 lg:grid-cols-3">
            <div class="rounded-2xl border border-amber-100 bg-amber-50 p-6 dark:border-amber-900/40 dark:bg-amber-900/10">
              <div class="text-sm font-medium text-gray-600 dark:text-dark-300">
                {{ t('referral.page.stats.total') }}
              </div>
              <div class="mt-2 text-4xl font-bold text-gray-900 dark:text-white">
                {{ info?.stats?.total_invites ?? 0 }}
              </div>
            </div>

            <div class="rounded-2xl border border-blue-100 bg-blue-50 p-6 dark:border-blue-900/40 dark:bg-blue-900/10">
              <div class="text-sm font-medium text-gray-600 dark:text-dark-300">
                {{ t('referral.page.stats.qualified') }}
              </div>
              <div class="mt-2 text-4xl font-bold text-blue-600 dark:text-blue-300">
                {{ info?.stats?.qualified_invites ?? 0 }}
              </div>
            </div>

            <div class="rounded-2xl border border-emerald-100 bg-emerald-50 p-6 dark:border-emerald-900/40 dark:bg-emerald-900/10">
              <div class="text-sm font-medium text-gray-600 dark:text-dark-300">
                {{ t('referral.page.stats.rewarded') }}
              </div>
              <div class="mt-2 text-4xl font-bold text-emerald-600 dark:text-emerald-300">
                {{ formatUSD(info?.stats?.rewarded_usd ?? 0) }}
              </div>
            </div>
          </div>

          <div class="mt-6 flex items-center justify-between">
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">
              {{ t('referral.page.recentTitle') }}
            </h3>
            <div class="text-xs text-gray-500 dark:text-dark-400">
              {{ t('referral.page.recentTip', { n: recentLimit }) }}
            </div>
          </div>

          <div v-if="loadingInvitees" class="mt-4 flex items-center justify-center py-10">
            <LoadingSpinner />
          </div>

          <div v-else class="mt-4 space-y-3">
            <div
              v-for="inv in invitees"
              :key="inv.id"
              class="flex flex-col gap-3 rounded-2xl border border-gray-100 bg-white p-5 sm:flex-row sm:items-center sm:justify-between dark:border-dark-800 dark:bg-dark-900"
            >
              <div class="space-y-1">
                <div class="text-sm font-semibold text-gray-900 dark:text-white">
                  {{ inv.invitee_username || '-' }}
                </div>
                <div class="text-xs text-gray-500 dark:text-dark-400">
                  {{ t('referral.page.registeredAt') }}：{{ formatDateTime(inv.created_at) }}
                </div>
              </div>

              <div class="flex items-center gap-2">
                <span
                  class="badge"
                  :class="inv.reward_issued ? 'badge-success' : 'badge-gray'"
                >
                  {{ inv.reward_issued ? t('referral.page.issued') : t('referral.page.notIssued') }}
                </span>
              </div>
            </div>

            <div v-if="invitees.length === 0" class="rounded-2xl border border-gray-100 bg-gray-50 p-6 text-sm text-gray-600 dark:border-dark-800 dark:bg-dark-900 dark:text-dark-300">
              {{ t('referral.page.noInvites') }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { formatDateTime } from '@/utils/format'
import { referralAPI } from '@/api/referral'
import { useAppStore } from '@/stores'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import type { ReferralInfoResponse, ReferralInvite, PaginatedResponse } from '@/types'

const { t } = useI18n()
const appStore = useAppStore()

const loadingInfo = ref(false)
const loadingInvitees = ref(false)
const info = ref<ReferralInfoResponse | null>(null)
const invitees = ref<ReferralInvite[]>([])

const recentLimit = 5

function formatUSD(v: number): string {
  return `$${Number(v || 0).toFixed(2)}`
}

async function copy(text: string) {
  if (!text) return
  try {
    await navigator.clipboard.writeText(text)
    appStore.showSuccess(t('common.copied'))
  } catch {
    appStore.showError(t('common.error'))
  }
}

async function loadInfo() {
  loadingInfo.value = true
  try {
    info.value = await referralAPI.getReferralInfo()
  } finally {
    loadingInfo.value = false
  }
}

async function loadInvitees() {
  if (!info.value?.enabled) {
    invitees.value = []
    return
  }
  loadingInvitees.value = true
  try {
    const data: PaginatedResponse<ReferralInvite> = await referralAPI.listInvitees(1, recentLimit)
    invitees.value = data.items
  } finally {
    loadingInvitees.value = false
  }
}

async function refreshAll() {
  await loadInfo()
  await loadInvitees()
}

onMounted(() => {
  refreshAll()
})
</script>
