<template>
  <AppLayout>
  <div class="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
    <!-- Header -->
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('modelPricing.title') }}</h1>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('modelPricing.description') }}</p>
    </div>

    <!-- Tab switcher -->
    <div class="mb-6 border-b border-gray-200 dark:border-dark-700">
      <nav class="-mb-px flex space-x-6" aria-label="Model tabs">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          @click="activeTab = tab.id"
          :class="[
            'whitespace-nowrap py-3 px-1 border-b-2 font-medium text-sm transition-colors',
            activeTab === tab.id
              ? 'border-primary-500 text-primary-600 dark:text-primary-400'
              : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'
          ]"
        >
          {{ tab.label }}
        </button>
      </nav>
    </div>

    <!-- Tab content -->
    <div v-for="tab in tabs" :key="tab.id" v-show="activeTab === tab.id" class="space-y-8">

      <!-- Section 1: 优惠计算公式 -->
      <div class="rounded-xl border border-gray-200 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-800">
        <h2 class="mb-5 text-base font-semibold text-gray-900 dark:text-white">优惠计算公式</h2>

        <!-- 合并表格：类型 / 模型倍率 / 简易公式 -->
        <div class="overflow-hidden rounded-lg border border-gray-200 dark:border-dark-600">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-600">
            <thead class="bg-gray-50 dark:bg-dark-700">
              <tr>
                <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">
                  {{ t('modelPricing.colType') }}
                </th>
                <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">
                  {{ t('modelPricing.colRate') }}
                </th>
                <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">
                  {{ t('modelPricing.colFormula') }}
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 bg-white dark:divide-dark-700 dark:bg-dark-800">
              <tr v-for="row in tab.rows" :key="row.type">
                <td class="px-4 py-3 text-sm font-mono text-gray-800 dark:text-gray-200">{{ row.type }}</td>
                <td class="px-4 py-3 text-sm font-medium text-primary-600 dark:text-primary-400">{{ row.rate }}</td>
                <td class="px-4 py-3 text-sm text-gray-700 dark:text-gray-300">{{ row.formula }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- 补充说明 -->
        <p class="mt-3 text-xs text-gray-400 dark:text-gray-500">
          {{ t('modelPricing.fullFormulaText') }}{{ t('modelPricing.fullFormulaNote') }}
        </p>
      </div>

      <!-- Section 2: 官方价格 -->
      <div class="rounded-xl border border-gray-200 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-800">
        <h2 class="mb-4 text-base font-semibold text-gray-900 dark:text-white">{{ t('modelPricing.officialPrice') }}</h2>
        <div class="mb-4 flex items-start gap-2 rounded-lg border-l-4 border-primary-400 bg-primary-50 px-4 py-3 dark:border-primary-600 dark:bg-primary-900/20">
          <svg class="mt-0.5 h-4 w-4 flex-shrink-0 text-primary-500 dark:text-primary-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z" />
          </svg>
          <p class="text-sm text-primary-700 dark:text-primary-300">{{ t('modelPricing.officialPriceHint') }}</p>
        </div>
        <a
          v-if="tab.officialUrl"
          :href="tab.officialUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="inline-flex items-center gap-2 rounded-lg border border-primary-300 bg-primary-50 px-4 py-2 text-sm font-medium text-primary-700 transition-colors hover:bg-primary-100 dark:border-primary-700 dark:bg-primary-900/20 dark:text-primary-400 dark:hover:bg-primary-900/40"
        >
          {{ t('modelPricing.officialPriceLink') }}
          <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13.5 6H5.25A2.25 2.25 0 003 8.25v10.5A2.25 2.25 0 005.25 21h10.5A2.25 2.25 0 0018 18.75V10.5m-10.5 6L21 3m0 0h-5.25M21 3v5.25" />
          </svg>
        </a>
        <p v-else class="text-sm italic text-gray-400 dark:text-gray-500">暂未配置官网链接</p>
      </div>

    </div>
  </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import AppLayout from '@/components/layout/AppLayout.vue'

const { t } = useI18n()
const appStore = useAppStore()

const activeTab = ref('claudeCode')

const claudeUrl = computed(() => appStore.cachedPublicSettings?.claude_official_url || '')
const codexUrl = computed(() => appStore.cachedPublicSettings?.codex_official_url || '')
const geminiUrl = computed(() => appStore.cachedPublicSettings?.gemini_official_url || '')

const tabs = computed(() => [
  {
    id: 'claudeCode',
    label: t('modelPricing.tabs.claudeCode'),
    officialUrl: claudeUrl.value,
    rows: [
      { type: 'claude code-low-price', formula: '1 人民币 = 1 美元用量', rate: '1x' },
      { type: 'claude code',           formula: '2 人民币 = 1 美元用量', rate: '2x' },
      { type: 'claude code VIP',       formula: '3 人民币 = 1 美元用量', rate: '3x' }
    ]
  },
  {
    id: 'codex',
    label: t('modelPricing.tabs.codex'),
    officialUrl: codexUrl.value,
    rows: [
      { type: 'Codex', formula: '1 人民币 = 1 美元用量', rate: '1x' }
    ]
  },
  {
    id: 'gemini',
    label: t('modelPricing.tabs.gemini'),
    officialUrl: geminiUrl.value,
    rows: [
      { type: 'Gemini', formula: '4 人民币 = 1 美元用量', rate: '4x' }
    ]
  }
])
</script>
