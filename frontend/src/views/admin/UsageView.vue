<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Stats Cards -->
      <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
          <!-- Total Requests -->
          <div class="card p-4">
          <div class="flex items-center gap-3">
            <div class="rounded-lg bg-blue-100 p-2 dark:bg-blue-900/30">
              <svg
                class="h-5 w-5 text-blue-600 dark:text-blue-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                />
              </svg>
            </div>
            <div>
              <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                {{ t('usage.totalRequests') }}
              </p>
              <p class="text-xl font-bold text-gray-900 dark:text-white">
                {{ usageStats?.total_requests?.toLocaleString() || '0' }}
              </p>
              <p class="text-xs text-gray-500 dark:text-gray-400">
                {{ t('usage.inSelectedRange') }}
              </p>
            </div>
          </div>
        </div>

        <!-- Total Tokens -->
        <div class="card p-4">
          <div class="flex items-center gap-3">
            <div class="rounded-lg bg-amber-100 p-2 dark:bg-amber-900/30">
              <svg
                class="h-5 w-5 text-amber-600 dark:text-amber-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="m21 7.5-9-5.25L3 7.5m18 0-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9"
                />
              </svg>
            </div>
            <div>
              <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                {{ t('usage.totalTokens') }}
              </p>
              <p class="text-xl font-bold text-gray-900 dark:text-white">
                {{ formatTokens(usageStats?.total_tokens || 0) }}
              </p>
              <p class="text-xs text-gray-500 dark:text-gray-400">
                {{ t('usage.in') }}: {{ formatTokens(usageStats?.total_input_tokens || 0) }} /
                {{ t('usage.out') }}: {{ formatTokens(usageStats?.total_output_tokens || 0) }}
              </p>
            </div>
          </div>
        </div>

        <!-- Total Cost -->
        <div class="card p-4">
          <div class="flex items-center gap-3">
            <div class="rounded-lg bg-green-100 p-2 dark:bg-green-900/30">
              <svg
                class="h-5 w-5 text-green-600 dark:text-green-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            </div>
            <div class="min-w-0 flex-1">
              <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                {{ t('usage.totalCost') }}
              </p>
              <p class="text-xl font-bold text-green-600 dark:text-green-400">
                ${{ (usageStats?.total_actual_cost || 0).toFixed(4) }}
              </p>
              <p class="text-xs text-gray-500 dark:text-gray-400">
                <span class="line-through">${{ (usageStats?.total_cost || 0).toFixed(4) }}</span>
                {{ t('usage.standardCost') }}
              </p>
            </div>
          </div>
        </div>

        <!-- Average Duration -->
        <div class="card p-4">
          <div class="flex items-center gap-3">
            <div class="rounded-lg bg-primary-100 p-2 dark:bg-primary-900/30">
              <svg
                class="h-5 w-5 text-primary-600 dark:text-primary-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            </div>
            <div>
              <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                {{ t('usage.avgDuration') }}
              </p>
              <p class="text-xl font-bold text-gray-900 dark:text-white">
                {{ formatDuration(usageStats?.average_duration_ms || 0) }}
              </p>
              <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('usage.perRequest') }}</p>
            </div>
          </div>
        </div>
        </div>

        <!-- Charts Section -->
        <div class="space-y-4">
        <!-- Chart Controls -->
        <div class="card p-4">
          <div class="flex items-center gap-4">
            <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.dashboard.granularity') }}:</span>
            <div class="w-28">
              <Select v-model="granularity" :options="granularityOptions" @change="loadChartData" />
            </div>
          </div>
        </div>
        <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
          <ModelDistributionChart :model-stats="modelStats" :loading="chartsLoading" />
          <TokenUsageTrend :trend-data="trendData" :loading="chartsLoading" />
        </div>
      </div>

      <!-- Filters Section -->
      <div class="card">
          <div class="px-6 py-4">
          <div class="flex flex-wrap items-end gap-4">
            <!-- User Search -->
            <div class="min-w-[200px]">
              <label class="input-label">{{ t('admin.usage.userFilter') }}</label>
              <div class="relative">
                <input
                  v-model="userSearchKeyword"
                  type="text"
                  class="input pr-8"
                  :placeholder="t('admin.usage.searchUserPlaceholder')"
                  @input="debounceSearchUsers"
                  @focus="showUserDropdown = true"
                />
                <button
                  v-if="selectedUser"
                  @click="clearUserFilter"
                  class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                >
                  <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M6 18L18 6M6 6l12 12"
                    />
                  </svg>
                </button>
                <!-- User Dropdown -->
                <div
                  v-if="showUserDropdown && (userSearchResults.length > 0 || userSearchKeyword)"
                  class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-gray-700 dark:bg-gray-800"
                >
                  <div
                    v-if="userSearchLoading"
                    class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400"
                  >
                    {{ t('common.loading') }}
                  </div>
                  <div
                    v-else-if="userSearchResults.length === 0 && userSearchKeyword"
                    class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400"
                  >
                    {{ t('common.noOptionsFound') }}
                  </div>
                  <button
                    v-for="user in userSearchResults"
                    :key="user.id"
                    @click="selectUser(user)"
                    class="w-full px-4 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-gray-700"
                  >
                    <span class="font-medium text-gray-900 dark:text-white">{{ user.email }}</span>
                    <span class="ml-2 text-gray-500 dark:text-gray-400">#{{ user.id }}</span>
                  </button>
                </div>
              </div>
            </div>

            <!-- API Key Filter -->
            <div class="min-w-[180px]">
              <label class="input-label">{{ t('usage.apiKeyFilter') }}</label>
              <Select
                v-model="filters.api_key_id"
                :options="apiKeyOptions"
                :placeholder="t('usage.allApiKeys')"
                searchable
                @change="applyFilters"
              />
            </div>

            <!-- Model Filter -->
            <div class="min-w-[180px]">
              <label class="input-label">{{ t('usage.model') }}</label>
              <Select
                v-model="filters.model"
                :options="modelOptions"
                :placeholder="t('admin.usage.allModels')"
                searchable
                @change="applyFilters"
              />
            </div>

            <!-- Account Filter -->
            <div class="min-w-[180px]">
              <label class="input-label">{{ t('admin.usage.account') }}</label>
              <Select
                v-model="filters.account_id"
                :options="accountOptions"
                :placeholder="t('admin.usage.allAccounts')"
                @change="applyFilters"
              />
            </div>

            <!-- Stream Type Filter -->
            <div class="min-w-[150px]">
              <label class="input-label">{{ t('usage.type') }}</label>
              <Select
                v-model="filters.stream"
                :options="streamOptions"
                :placeholder="t('admin.usage.allTypes')"
                @change="applyFilters"
              />
            </div>

            <!-- Billing Type Filter -->
            <div class="min-w-[150px]">
              <label class="input-label">{{ t('usage.billingType') }}</label>
              <Select
                v-model="filters.billing_type"
                :options="billingTypeOptions"
                :placeholder="t('admin.usage.allBillingTypes')"
                @change="applyFilters"
              />
            </div>

            <!-- Group Filter -->
            <div class="min-w-[150px]">
              <label class="input-label">{{ t('admin.usage.group') }}</label>
              <Select
                v-model="filters.group_id"
                :options="groupOptions"
                :placeholder="t('admin.usage.allGroups')"
                @change="applyFilters"
              />
            </div>

            <!-- Date Range Filter -->
            <div>
              <label class="input-label">{{ t('usage.timeRange') }}</label>
              <DateRangePicker
                v-model:start-date="startDate"
                v-model:end-date="endDate"
                @change="onDateRangeChange"
              />
            </div>

            <!-- Actions -->
            <div class="ml-auto flex items-center gap-3">
              <button @click="resetFilters" class="btn btn-secondary">
                {{ t('common.reset') }}
              </button>
              <button @click="exportToExcel" :disabled="exporting" class="btn btn-primary">
                {{ t('usage.exportExcel') }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Table Section -->
      <div class="card overflow-hidden">
        <div class="overflow-auto">
          <DataTable :columns="columns" :data="usageLogs" :loading="loading">
          <template #cell-user="{ row }">
            <div class="text-sm">
              <span class="font-medium text-gray-900 dark:text-white">{{
                row.user?.email || '-'
              }}</span>
              <span class="ml-1 text-gray-500 dark:text-gray-400">#{{ row.user_id }}</span>
            </div>
          </template>

          <template #cell-api_key="{ row }">
            <span class="text-sm text-gray-900 dark:text-white">{{
              row.api_key?.name || '-'
            }}</span>
          </template>

          <template #cell-account="{ row }">
            <span class="text-sm text-gray-900 dark:text-white">{{
              row.account?.name || '-'
            }}</span>
          </template>

          <template #cell-model="{ value }">
            <span class="font-medium text-gray-900 dark:text-white">{{ value }}</span>
          </template>

          <template #cell-group="{ row }">
            <span
              v-if="row.group"
              class="inline-flex items-center rounded px-2 py-0.5 text-xs font-medium bg-primary-100 text-primary-800 dark:bg-primary-900 dark:text-primary-200"
            >
              {{ row.group.name }}
            </span>
            <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
          </template>

          <template #cell-stream="{ row }">
            <span
              class="inline-flex items-center rounded px-2 py-0.5 text-xs font-medium"
              :class="
                row.stream
                  ? 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200'
                  : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
              "
            >
              {{ row.stream ? t('usage.stream') : t('usage.sync') }}
            </span>
          </template>

          <template #cell-tokens="{ row }">
            <!-- 图片生成请求 -->
            <div v-if="row.image_count > 0" class="flex items-center gap-1.5">
              <svg
                class="h-4 w-4 text-indigo-500"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
                />
              </svg>
              <span class="font-medium text-gray-900 dark:text-white">{{ row.image_count }}{{ $t('usage.imageUnit') }}</span>
              <span class="text-gray-400">({{ row.image_size || '2K' }})</span>
            </div>
            <!-- Token 请求 -->
            <div v-else class="flex items-center gap-1.5">
              <div class="space-y-1.5 text-sm">
                <!-- Input / Output Tokens -->
                <div class="flex items-center gap-2">
                  <!-- Input -->
                  <div class="inline-flex items-center gap-1">
                    <svg
                      class="h-3.5 w-3.5 text-emerald-500"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M19 14l-7 7m0 0l-7-7m7 7V3"
                      />
                    </svg>
                    <span class="font-medium text-gray-900 dark:text-white">{{
                      row.input_tokens.toLocaleString()
                    }}</span>
                  </div>
                  <!-- Output -->
                  <div class="inline-flex items-center gap-1">
                    <svg
                      class="h-3.5 w-3.5 text-primary-500"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M5 10l7-7m0 0l7 7m-7-7v18"
                      />
                    </svg>
                    <span class="font-medium text-gray-900 dark:text-white">{{
                      row.output_tokens.toLocaleString()
                    }}</span>
                  </div>
                </div>
                <!-- Cache Tokens (Read + Write) -->
                <div
                  v-if="row.cache_read_tokens > 0 || row.cache_creation_tokens > 0"
                  class="flex items-center gap-2"
                >
                  <!-- Cache Read -->
                  <div v-if="row.cache_read_tokens > 0" class="inline-flex items-center gap-1">
                    <svg
                      class="h-3.5 w-3.5 text-sky-500"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4"
                      />
                    </svg>
                    <span class="font-medium text-sky-600 dark:text-sky-400">{{
                      formatCacheTokens(row.cache_read_tokens)
                    }}</span>
                  </div>
                  <!-- Cache Write -->
                  <div v-if="row.cache_creation_tokens > 0" class="inline-flex items-center gap-1">
                    <svg
                      class="h-3.5 w-3.5 text-amber-500"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                      />
                    </svg>
                    <span class="font-medium text-amber-600 dark:text-amber-400">{{
                      formatCacheTokens(row.cache_creation_tokens)
                    }}</span>
                  </div>
                </div>
              </div>
              <!-- Token Detail Tooltip -->
              <div
                class="group relative"
                @mouseenter="showTokenTooltip($event, row)"
                @mouseleave="hideTokenTooltip"
              >
                <div
                  class="flex h-4 w-4 cursor-help items-center justify-center rounded-full bg-gray-100 transition-colors group-hover:bg-blue-100 dark:bg-gray-700 dark:group-hover:bg-blue-900/50"
                >
                  <svg
                    class="h-3 w-3 text-gray-400 group-hover:text-blue-500 dark:text-gray-500 dark:group-hover:text-blue-400"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path
                      fill-rule="evenodd"
                      d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                      clip-rule="evenodd"
                    />
                  </svg>
                </div>
              </div>
            </div>
          </template>

          <template #cell-cost="{ row }">
            <div class="flex items-center gap-1.5 text-sm">
              <span class="font-medium text-green-600 dark:text-green-400">
                ${{ row.actual_cost.toFixed(6) }}
              </span>
              <!-- Cost Detail Tooltip -->
              <div
                class="group relative"
                @mouseenter="showTooltip($event, row)"
                @mouseleave="hideTooltip"
              >
                <div
                  class="flex h-4 w-4 cursor-help items-center justify-center rounded-full bg-gray-100 transition-colors group-hover:bg-blue-100 dark:bg-gray-700 dark:group-hover:bg-blue-900/50"
                >
                  <svg
                    class="h-3 w-3 text-gray-400 group-hover:text-blue-500 dark:text-gray-500 dark:group-hover:text-blue-400"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path
                      fill-rule="evenodd"
                      d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                      clip-rule="evenodd"
                    />
                  </svg>
                </div>
              </div>
            </div>
          </template>

          <template #cell-billing_type="{ row }">
            <span
              class="inline-flex items-center rounded px-2 py-0.5 text-xs font-medium"
              :class="
                row.billing_type === 1
                  ? 'bg-primary-100 text-primary-800 dark:bg-primary-900 dark:text-primary-200'
                  : 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900 dark:text-emerald-200'
              "
            >
              {{ row.billing_type === 1 ? t('usage.subscription') : t('usage.balance') }}
            </span>
          </template>

          <template #cell-first_token="{ row }">
            <span
              v-if="row.first_token_ms != null"
              class="text-sm text-gray-600 dark:text-gray-400"
            >
              {{ formatDuration(row.first_token_ms) }}
            </span>
            <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
          </template>

          <template #cell-duration="{ row }">
            <span class="text-sm text-gray-600 dark:text-gray-400">{{
              formatDuration(row.duration_ms)
            }}</span>
          </template>

          <template #cell-created_at="{ value }">
            <span class="text-sm text-gray-600 dark:text-gray-400">{{
              formatDateTime(value)
            }}</span>
          </template>

          <template #cell-request_id="{ row }">
            <div v-if="row.request_id" class="flex items-center gap-1.5 max-w-[120px]">
              <span
                class="font-mono text-xs text-gray-500 dark:text-gray-400 truncate"
                :title="row.request_id"
              >
                {{ row.request_id }}
              </span>
              <button
                @click="copyRequestId(row.request_id)"
                class="flex-shrink-0 rounded p-0.5 transition-colors hover:bg-gray-100 dark:hover:bg-dark-700"
                :class="
                  copiedRequestId === row.request_id
                    ? 'text-green-500'
                    : 'text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'
                "
                :title="copiedRequestId === row.request_id ? t('keys.copied') : t('keys.copyToClipboard')"
              >
                <svg
                  v-if="copiedRequestId === row.request_id"
                  class="h-3.5 w-3.5"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  stroke-width="2"
                >
                  <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                </svg>
                <svg
                  v-else
                  class="h-3.5 w-3.5"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  stroke-width="2"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
                  />
                </svg>
              </button>
            </div>
            <span v-else class="text-gray-400 dark:text-gray-500">-</span>
          </template>

          <template #empty>
            <EmptyState :message="t('usage.noRecords')" />
          </template>
        </DataTable>
        </div>
      </div>

      <!-- Pagination -->
      <Pagination
        v-if="pagination.total > 0"
        :page="pagination.page"
        :total="pagination.total"
        :page-size="pagination.page_size"
        @update:page="handlePageChange"
        @update:pageSize="handlePageSizeChange"
      />
    </div>
  </AppLayout>
  <UsageExportProgress :show="exportProgress.show" :progress="exportProgress.progress" :current="exportProgress.current" :total="exportProgress.total" :estimated-time="exportProgress.estimatedTime" @cancel="cancelExport" />
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { saveAs } from 'file-saver'
import { useAppStore } from '@/stores/app'; import { adminAPI } from '@/api/admin'; import { adminUsageAPI } from '@/api/admin/usage'
import AppLayout from '@/components/layout/AppLayout.vue'; import Pagination from '@/components/common/Pagination.vue'; import Select from '@/components/common/Select.vue'
import UsageStatsCards from '@/components/admin/usage/UsageStatsCards.vue'; import UsageFilters from '@/components/admin/usage/UsageFilters.vue'
import UsageTable from '@/components/admin/usage/UsageTable.vue'; import UsageExportProgress from '@/components/admin/usage/UsageExportProgress.vue'
import ModelDistributionChart from '@/components/charts/ModelDistributionChart.vue'; import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'
import type { UsageLog, TrendDataPoint, ModelStat } from '@/types'; import type { AdminUsageStatsResponse, AdminUsageQueryParams } from '@/api/admin/usage'

const { t } = useI18n()
const appStore = useAppStore()
const usageStats = ref<AdminUsageStatsResponse | null>(null); const usageLogs = ref<UsageLog[]>([]); const loading = ref(false); const exporting = ref(false)
const trendData = ref<TrendDataPoint[]>([]); const modelStats = ref<ModelStat[]>([]); const chartsLoading = ref(false); const granularity = ref<'day' | 'hour'>('day')
let abortController: AbortController | null = null; let exportAbortController: AbortController | null = null
const exportProgress = reactive({ show: false, progress: 0, current: 0, total: 0, estimatedTime: '' })

const granularityOptions = computed(() => [{ value: 'day', label: t('admin.dashboard.day') }, { value: 'hour', label: t('admin.dashboard.hour') }])
const formatLD = (d: Date) => d.toISOString().split('T')[0]
const now = new Date(); const weekAgo = new Date(Date.now() - 6 * 86400000)
const startDate = ref(formatLD(weekAgo)); const endDate = ref(formatLD(now))
const filters = ref<AdminUsageQueryParams>({ user_id: undefined, model: undefined, group_id: undefined, start_date: startDate.value, end_date: endDate.value })
const pagination = reactive({ page: 1, page_size: 20, total: 0 })

const loadLogs = async () => {
  abortController?.abort(); const c = new AbortController(); abortController = c; loading.value = true
  try {
    const res = await adminAPI.usage.list({ page: pagination.page, page_size: pagination.page_size, ...filters.value }, { signal: c.signal })
    if(!c.signal.aborted) { usageLogs.value = res.items; pagination.total = res.total }
  } catch (error: any) { if(error?.name !== 'AbortError') console.error('Failed to load usage logs:', error) } finally { if(abortController === c) loading.value = false }
}
const loadStats = async () => { try { const s = await adminAPI.usage.getStats(filters.value); usageStats.value = s } catch (error) { console.error('Failed to load usage stats:', error) } }
const loadChartData = async () => {
  chartsLoading.value = true
  try {
    const params = { start_date: filters.value.start_date || startDate.value, end_date: filters.value.end_date || endDate.value, granularity: granularity.value, user_id: filters.value.user_id }
    const [trendRes, modelRes] = await Promise.all([adminAPI.dashboard.getUsageTrend(params), adminAPI.dashboard.getModelStats({ start_date: params.start_date, end_date: params.end_date, user_id: params.user_id })])
    trendData.value = trendRes.trend || []; modelStats.value = modelRes.models || []
  } catch (error) { console.error('Failed to load chart data:', error) } finally { chartsLoading.value = false }
}
const applyFilters = () => { pagination.page = 1; loadLogs(); loadStats(); loadChartData() }
const resetFilters = () => { startDate.value = formatLD(weekAgo); endDate.value = formatLD(now); filters.value = { start_date: startDate.value, end_date: endDate.value }; granularity.value = 'day'; applyFilters() }
const handlePageChange = (p: number) => { pagination.page = p; loadLogs() }
const handlePageSizeChange = (s: number) => { pagination.page_size = s; pagination.page = 1; loadLogs() }
const cancelExport = () => exportAbortController?.abort()

const exportToExcel = async () => {
  if (exporting.value) return; exporting.value = true; exportProgress.show = true
  const c = new AbortController(); exportAbortController = c
  try {
    const all: UsageLog[] = []; let p = 1; let total = pagination.total
    while (true) {
      const res = await adminUsageAPI.list({ page: p, page_size: 100, ...filters.value }, { signal: c.signal })
      if (c.signal.aborted) break; if (p === 1) { total = res.total; exportProgress.total = total }
      if (res.items?.length) all.push(...res.items)
      exportProgress.current = all.length; exportProgress.progress = total > 0 ? Math.min(100, Math.round(all.length/total*100)) : 0
      if (all.length >= total || res.items.length < 100) break; p++
    }
    if(!c.signal.aborted) {
      const XLSX = await import('xlsx')
      const headers = [
        t('usage.time'), t('admin.usage.user'), t('usage.apiKeyFilter'),
        t('admin.usage.account'), t('usage.model'), t('admin.usage.group'),
        t('usage.type'),
        t('admin.usage.inputTokens'), t('admin.usage.outputTokens'),
        t('admin.usage.cacheReadTokens'), t('admin.usage.cacheCreationTokens'),
        t('admin.usage.inputCost'), t('admin.usage.outputCost'),
        t('admin.usage.cacheReadCost'), t('admin.usage.cacheCreationCost'),
        t('usage.rate'), t('usage.original'), t('usage.billed'),
        t('usage.firstToken'), t('usage.duration'),
        t('admin.usage.requestId'), t('usage.userAgent'), t('admin.usage.ipAddress')
      ]
      const rows = all.map(log => [
        log.created_at,
        log.user?.email || '',
        log.api_key?.name || '',
        log.account?.name || '',
        log.model,
        log.group?.name || '',
        log.stream ? t('usage.stream') : t('usage.sync'),
        log.input_tokens,
        log.output_tokens,
        log.cache_read_tokens,
        log.cache_creation_tokens,
        log.input_cost?.toFixed(6) || '0.000000',
        log.output_cost?.toFixed(6) || '0.000000',
        log.cache_read_cost?.toFixed(6) || '0.000000',
        log.cache_creation_cost?.toFixed(6) || '0.000000',
        log.rate_multiplier?.toFixed(2) || '1.00',
        log.total_cost?.toFixed(6) || '0.000000',
        log.actual_cost?.toFixed(6) || '0.000000',
        log.first_token_ms ?? '',
        log.duration_ms,
        log.request_id || '',
        log.user_agent || '',
        log.ip_address || ''
      ])
      const ws = XLSX.utils.aoa_to_sheet([headers, ...rows])
      const wb = XLSX.utils.book_new()
      XLSX.utils.book_append_sheet(wb, ws, 'Usage')
      saveAs(new Blob([XLSX.write(wb, { bookType: 'xlsx', type: 'array' })], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' }), `usage_${filters.value.start_date}_to_${filters.value.end_date}.xlsx`)
      appStore.showSuccess(t('usage.exportSuccess'))
    }
  } catch (error) { console.error('Failed to export:', error); appStore.showError('Export Failed') }
  finally { if(exportAbortController === c) { exportAbortController = null; exporting.value = false; exportProgress.show = false } }
}

onMounted(() => { loadLogs(); loadStats(); loadChartData() })
onUnmounted(() => { abortController?.abort(); exportAbortController?.abort() })
</script>
