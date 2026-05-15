<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="card p-5">
        <div class="flex flex-wrap gap-2 border-b border-gray-200 pb-4 dark:border-dark-700" role="tablist">
          <button
            v-for="tab in tabs"
            :key="tab.id"
            type="button"
            class="rounded-lg px-4 py-2 text-sm font-medium transition-colors"
            :class="
              activeTab === tab.id
                ? 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300'
                : 'text-gray-600 hover:bg-gray-100 dark:text-dark-300 dark:hover:bg-dark-800'
            "
            role="tab"
            :aria-selected="activeTab === tab.id"
            @click="activeTab = tab.id"
          >
            {{ tab.label }}
          </button>
        </div>

        <div v-if="active" class="mt-5 space-y-5">
          <div class="overflow-x-auto rounded-lg border border-gray-200 dark:border-dark-700">
            <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
              <thead class="bg-gray-50 dark:bg-dark-800">
                <tr>
                  <th class="px-4 py-3 text-left text-xs font-semibold uppercase text-gray-500 dark:text-dark-400">
                    {{ t('modelPricing.colType') }}
                  </th>
                  <th class="px-4 py-3 text-left text-xs font-semibold uppercase text-gray-500 dark:text-dark-400">
                    {{ t('modelPricing.colRate') }}
                  </th>
                  <th class="px-4 py-3 text-left text-xs font-semibold uppercase text-gray-500 dark:text-dark-400">
                    {{ t('modelPricing.colFormula') }}
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 bg-white dark:divide-dark-800 dark:bg-dark-900">
                <tr v-for="row in active.rows" :key="row.type">
                  <td class="px-4 py-3 font-mono text-sm text-gray-900 dark:text-white">{{ row.type }}</td>
                  <td class="px-4 py-3 text-sm font-semibold text-primary-600 dark:text-primary-300">{{ row.rate }}</td>
                  <td class="px-4 py-3 text-sm text-gray-700 dark:text-dark-300">{{ row.formula }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="rounded-lg border border-primary-200 bg-primary-50 p-4 dark:border-primary-800 dark:bg-primary-950/30">
            <p class="text-sm text-primary-800 dark:text-primary-200">
              {{ t('modelPricing.fullFormulaText') }}
              <span class="text-primary-600 dark:text-primary-300">{{ t('modelPricing.fullFormulaNote') }}</span>
            </p>
          </div>

          <div class="flex flex-col gap-3 rounded-lg border border-gray-200 p-4 dark:border-dark-700 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h2 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('modelPricing.officialPrice') }}</h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('modelPricing.officialPriceHint') }}</p>
            </div>
            <a
              :href="active.officialUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="btn btn-secondary inline-flex items-center justify-center gap-2"
            >
              {{ t('modelPricing.officialPriceLink') }}
              <span aria-hidden="true">&nearr;</span>
            </a>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { useAppStore } from '@/stores'

type PricingTabID = 'claudeCode' | 'codex' | 'gemini'

interface PricingRow {
  type: string
  rate: string
  formula: string
}

interface PricingTab {
  id: PricingTabID
  label: string
  officialUrl: string
  rows: PricingRow[]
}

const { t } = useI18n()
const appStore = useAppStore()
const activeTab = ref<PricingTabID>('claudeCode')

const fallbackOfficialUrls: Record<PricingTabID, string> = {
  claudeCode: 'https://docs.anthropic.com/en/docs/about-claude/pricing',
  codex: 'https://openai.com/api/pricing/',
  gemini: 'https://ai.google.dev/gemini-api/docs/pricing'
}

function publicSettingString(key: string): string {
  const settings = appStore.cachedPublicSettings as unknown as Record<string, unknown> | null
  const value = settings?.[key]
  return typeof value === 'string' ? value.trim() : ''
}

const tabs = computed<PricingTab[]>(() => [
  {
    id: 'claudeCode',
    label: t('modelPricing.tabs.claudeCode'),
    officialUrl: publicSettingString('claude_official_url') || fallbackOfficialUrls.claudeCode,
    rows: [
      {
        type: t('modelPricing.rows.claudeLow.type'),
        rate: '1x',
        formula: t('modelPricing.rows.claudeLow.formula')
      },
      {
        type: t('modelPricing.rows.claudeStandard.type'),
        rate: '2x',
        formula: t('modelPricing.rows.claudeStandard.formula')
      },
      {
        type: t('modelPricing.rows.claudeVip.type'),
        rate: '3x',
        formula: t('modelPricing.rows.claudeVip.formula')
      }
    ]
  },
  {
    id: 'codex',
    label: t('modelPricing.tabs.codex'),
    officialUrl: publicSettingString('codex_official_url') || fallbackOfficialUrls.codex,
    rows: [
      {
        type: t('modelPricing.rows.codex.type'),
        rate: '1x',
        formula: t('modelPricing.rows.codex.formula')
      }
    ]
  },
  {
    id: 'gemini',
    label: t('modelPricing.tabs.gemini'),
    officialUrl: publicSettingString('gemini_official_url') || fallbackOfficialUrls.gemini,
    rows: [
      {
        type: t('modelPricing.rows.gemini.type'),
        rate: '4x',
        formula: t('modelPricing.rows.gemini.formula')
      }
    ]
  }
])

const active = computed(() => tabs.value.find((tab) => tab.id === activeTab.value) || tabs.value[0])
</script>
