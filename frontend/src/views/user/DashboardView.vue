<template>
  <AppLayout>
    <div class="space-y-6">
      <div v-if="loading" class="flex items-center justify-center py-12"><LoadingSpinner /></div>
      <template v-else-if="stats">
        <FirstRechargePromotion class="-mt-2" />

        <!-- Row 1: Core Stats -->
        <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
          <!-- Balance -->
          <div v-if="!authStore.isSimpleMode" class="card animate-fade-in-up p-4 stagger-1">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-emerald-100 p-2 dark:bg-emerald-900/30">
                <svg
                  class="h-5 w-5 text-emerald-600 dark:text-emerald-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z"
                  />
                </svg>
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('dashboard.balance') }}
                </p>
                <p class="text-xl font-bold text-emerald-600 dark:text-emerald-400">
                  ${{ formatBalance(user?.balance || 0) }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('common.available') }}</p>
              </div>
            </div>
          </div>

          <!-- API Keys -->
          <div class="card animate-fade-in-up p-4 stagger-2">
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
                    d="M15.75 5.25a3 3 0 013 3m3 0a6 6 0 01-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 1121.75 8.25z"
                  />
                </svg>
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('dashboard.apiKeys') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ stats.total_api_keys }}
                </p>
                <p class="text-xs text-green-600 dark:text-green-400">
                  {{ stats.active_api_keys }} {{ t('common.active') }}
                </p>
              </div>
            </div>
          </div>

          <!-- Today Requests -->
          <div class="card animate-fade-in-up p-4 stagger-3">
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
                    d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z"
                  />
                </svg>
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('dashboard.todayRequests') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ stats.today_requests }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  {{ t('common.total') }}: {{ formatNumber(stats.total_requests) }}
                </p>
              </div>
            </div>
          </div>

          <!-- Today Cost -->
          <div class="card animate-fade-in-up p-4 stagger-4">
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
                    d="M12 6v12m-3-2.818l.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('dashboard.todayCost') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  <span class="text-primary-600 dark:text-primary-400" :title="t('dashboard.actual')"
                    >${{ formatCost(stats.today_actual_cost) }}</span
                  >
                  <span
                    class="text-sm font-normal text-gray-400 dark:text-gray-500"
                    :title="t('dashboard.standard')"
                  >
                    / ${{ formatCost(stats.today_cost) }}</span
                  >
                </p>
                <p class="text-xs">
                  <span class="text-gray-500 dark:text-gray-400">{{ t('common.total') }}: </span>
                  <span class="text-primary-600 dark:text-primary-400" :title="t('dashboard.actual')"
                    >${{ formatCost(stats.total_actual_cost) }}</span
                  >
                  <span class="text-gray-400 dark:text-gray-500" :title="t('dashboard.standard')">
                    / ${{ formatCost(stats.total_cost) }}</span
                  >
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Row 2: Token Stats -->
        <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
          <!-- Today Tokens -->
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
                  {{ t('dashboard.todayTokens') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ formatTokens(stats.today_tokens) }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  {{ t('dashboard.input') }}: {{ formatTokens(stats.today_input_tokens) }} /
                  {{ t('dashboard.output') }}: {{ formatTokens(stats.today_output_tokens) }}
                </p>
              </div>
            </div>
          </div>

          <!-- Total Tokens -->
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
                    d="M20.25 6.375c0 2.278-3.694 4.125-8.25 4.125S3.75 8.653 3.75 6.375m16.5 0c0-2.278-3.694-4.125-8.25-4.125S3.75 4.097 3.75 6.375m16.5 0v11.25c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125V6.375m16.5 0v3.75m-16.5-3.75v3.75m16.5 0v3.75C20.25 16.153 16.556 18 12 18s-8.25-1.847-8.25-4.125v-3.75m16.5 0c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125"
                  />
                </svg>
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('dashboard.totalTokens') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ formatTokens(stats.total_tokens) }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  {{ t('dashboard.input') }}: {{ formatTokens(stats.total_input_tokens) }} /
                  {{ t('dashboard.output') }}: {{ formatTokens(stats.total_output_tokens) }}
                </p>
              </div>
            </div>
          </div>

          <!-- Performance (RPM/TPM) -->
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
                    d="M13 10V3L4 14h7v7l9-11h-7z"
                  />
                </svg>
              </div>
              <div class="flex-1">
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('dashboard.performance') }}
                </p>
                <div class="flex items-baseline gap-2">
                  <p class="text-xl font-bold text-gray-900 dark:text-white">
                    {{ formatTokens(stats.rpm) }}
                  </p>
                  <span class="text-xs text-gray-500 dark:text-gray-400">RPM</span>
                </div>
                <div class="flex items-baseline gap-2">
                  <p class="text-sm font-semibold text-primary-600 dark:text-primary-400">
                    {{ formatTokens(stats.tpm) }}
                  </p>
                  <span class="text-xs text-gray-500 dark:text-gray-400">TPM</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Avg Response Time -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-rose-100 p-2 dark:bg-rose-900/30">
                <svg
                  class="h-5 w-5 text-rose-600 dark:text-rose-400"
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
                  {{ t('dashboard.avgResponse') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ formatDuration(stats.average_duration_ms) }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  {{ t('dashboard.averageTime') }}
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Charts Section -->
        <div class="space-y-6">
          <!-- Date Range Filter -->
          <div class="card p-4">
            <div class="flex flex-wrap items-center gap-4">
              <div class="flex items-center gap-2">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300"
                  >{{ t('dashboard.timeRange') }}:</span
                >
                <DateRangePicker
                  v-model:start-date="startDate"
                  v-model:end-date="endDate"
                  @change="onDateRangeChange"
                />
              </div>
              <div class="ml-auto flex items-center gap-2">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300"
                  >{{ t('dashboard.granularity') }}:</span
                >
                <div class="w-28">
                  <Select
                    v-model="granularity"
                    :options="granularityOptions"
                    @change="loadChartData"
                  />
                </div>
              </div>
            </div>
          </div>

          <!-- Charts Grid -->
          <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
            <!-- Model Distribution Chart -->
            <div class="card relative overflow-hidden p-4">
              <div
                v-if="loadingCharts"
                class="absolute inset-0 z-10 flex items-center justify-center bg-white/50 backdrop-blur-sm dark:bg-dark-800/50"
              >
                <LoadingSpinner size="md" />
              </div>
              <h3 class="mb-4 text-sm font-semibold text-gray-900 dark:text-white">
                {{ t('dashboard.modelDistribution') }}
              </h3>
              <div class="flex items-center gap-6">
                <div class="h-48 w-48">
                  <Doughnut
                    v-if="modelChartData"
                    ref="modelChartRef"
                    :data="modelChartData"
                    :options="doughnutOptions"
                  />
                  <div
                    v-else
                    class="flex h-full items-center justify-center text-sm text-gray-500 dark:text-gray-400"
                  >
                    {{ t('dashboard.noDataAvailable') }}
                  </div>
                </div>
                <div class="max-h-48 flex-1 overflow-y-auto">
                  <table class="w-full text-xs">
                    <thead>
                      <tr class="text-gray-500 dark:text-gray-400">
                        <th class="pb-2 text-left">{{ t('dashboard.model') }}</th>
                        <th class="pb-2 text-right">{{ t('dashboard.requests') }}</th>
                        <th class="pb-2 text-right">{{ t('dashboard.tokens') }}</th>
                        <th class="pb-2 text-right">{{ t('dashboard.actual') }}</th>
                        <th class="pb-2 text-right">{{ t('dashboard.standard') }}</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr
                        v-for="model in modelStats"
                        :key="model.model"
                        class="border-t border-gray-100 dark:border-gray-700"
                      >
                        <td
                          class="max-w-[100px] truncate py-1.5 font-medium text-gray-900 dark:text-white"
                          :title="model.model"
                        >
                          {{ model.model }}
                        </td>
                        <td class="py-1.5 text-right text-gray-600 dark:text-gray-400">
                          {{ formatNumber(model.requests) }}
                        </td>
                        <td class="py-1.5 text-right text-gray-600 dark:text-gray-400">
                          {{ formatTokens(model.total_tokens) }}
                        </td>
                        <td class="py-1.5 text-right text-green-600 dark:text-green-400">
                          ${{ formatCost(model.actual_cost) }}
                        </td>
                        <td class="py-1.5 text-right text-gray-400 dark:text-gray-500">
                          ${{ formatCost(model.cost) }}
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>

            <!-- Token Usage Trend Chart -->
            <div class="card relative overflow-hidden p-4">
              <div
                v-if="loadingCharts"
                class="absolute inset-0 z-10 flex items-center justify-center bg-white/50 backdrop-blur-sm dark:bg-dark-800/50"
              >
                <LoadingSpinner size="md" />
              </div>
              <h3 class="mb-4 text-sm font-semibold text-gray-900 dark:text-white">
                {{ t('dashboard.tokenUsageTrend') }}
              </h3>
              <div class="h-48">
                <Line
                  v-if="trendChartData"
                  ref="trendChartRef"
                  :data="trendChartData"
                  :options="lineOptions"
                />
                <div
                  v-else
                  class="flex h-full items-center justify-center text-sm text-gray-500 dark:text-gray-400"
                >
                  {{ t('dashboard.noDataAvailable') }}
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Main Content Grid -->
        <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
          <!-- Recent Usage - Takes 2 columns -->
          <div class="lg:col-span-2">
            <div class="card">
              <div
                class="flex items-center justify-between border-b border-gray-100 px-6 py-4 dark:border-dark-700"
              >
                <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                  {{ t('dashboard.recentUsage') }}
                </h2>
                <span class="badge badge-gray">{{ t('dashboard.last7Days') }}</span>
              </div>
              <div class="p-6">
                <div v-if="loadingUsage" class="flex items-center justify-center py-12">
                  <LoadingSpinner size="lg" />
                </div>
                <div v-else-if="recentUsage.length === 0" class="py-8">
                  <EmptyState
                    :title="t('dashboard.noUsageRecords')"
                    :description="t('dashboard.startUsingApi')"
                  />
                </div>
                <div v-else class="space-y-3">
                  <div
                    v-for="log in recentUsage"
                    :key="log.id"
                    class="flex items-center justify-between rounded-xl bg-gray-50 p-4 transition-colors hover:bg-gray-100 dark:bg-dark-800/50 dark:hover:bg-dark-800"
                  >
                    <div class="flex items-center gap-4">
                      <div
                        class="flex h-10 w-10 items-center justify-center rounded-xl bg-primary-100 dark:bg-primary-900/30"
                      >
                        <svg
                          class="h-5 w-5 text-primary-600 dark:text-primary-400"
                          fill="none"
                          viewBox="0 0 24 24"
                          stroke="currentColor"
                          stroke-width="1.5"
                        >
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            d="M9.75 3.104v5.714a2.25 2.25 0 01-.659 1.591L5 14.5M9.75 3.104c-.251.023-.501.05-.75.082m.75-.082a24.301 24.301 0 014.5 0m0 0v5.714c0 .597.237 1.17.659 1.591L19.8 15.3M14.25 3.104c.251.023.501.05.75.082M19.8 15.3l-1.57.393A9.065 9.065 0 0112 15a9.065 9.065 0 00-6.23.693L5 14.5m14.8.8l1.402 1.402c1.232 1.232.65 3.318-1.067 3.611A48.309 48.309 0 0112 21c-2.773 0-5.491-.235-8.135-.687-1.718-.293-2.3-2.379-1.067-3.61L5 14.5"
                          />
                        </svg>
                      </div>
                      <div>
                        <p class="text-sm font-medium text-gray-900 dark:text-white">
                          {{ log.model }}
                        </p>
                        <p class="text-xs text-gray-500 dark:text-dark-400">
                          {{ formatDateTime(log.created_at) }}
                        </p>
                      </div>
                    </div>
                    <div class="text-right">
                      <p class="text-sm font-semibold">
                        <span class="text-green-600 dark:text-green-400" :title="t('dashboard.actual')"
                          >${{ formatCost(log.actual_cost) }}</span
                        >
                        <span class="font-normal text-gray-400 dark:text-gray-500" :title="t('dashboard.standard')">
                          / ${{ formatCost(log.total_cost) }}</span
                        >
                      </p>
                      <p class="text-xs text-gray-500 dark:text-dark-400">
                        {{ (log.input_tokens + log.output_tokens).toLocaleString() }} tokens
                      </p>
                    </div>
                  </div>

                  <router-link
                    to="/usage"
                    class="flex items-center justify-center gap-2 py-3 text-sm font-medium text-primary-600 transition-colors hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300"
                  >
                    {{ t('dashboard.viewAllUsage') }}
                    <svg
                      class="h-4 w-4"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      stroke-width="1.5"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M13.5 4.5L21 12m0 0l-7.5 7.5M21 12H3"
                      />
                    </svg>
                  </router-link>
                </div>
              </div>
            </div>
          </div>

          <!-- Quick Actions - Takes 1 column -->
          <div class="lg:col-span-1">
            <div class="card">
              <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
                <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                  {{ t('dashboard.quickActions') }}
                </h2>
              </div>
              <div class="space-y-3 p-4">
                <button
                  @click="navigateTo('/keys')"
                  class="group flex w-full items-center gap-4 rounded-xl bg-gray-50 p-4 text-left transition-all duration-200 hover:bg-gray-100 dark:bg-dark-800/50 dark:hover:bg-dark-800"
                >
                  <div
                    class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-xl bg-primary-100 transition-transform group-hover:scale-105 dark:bg-primary-900/30"
                  >
                    <svg
                      class="h-6 w-6 text-primary-600 dark:text-primary-400"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      stroke-width="1.5"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M15.75 5.25a3 3 0 013 3m3 0a6 6 0 01-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 1121.75 8.25z"
                      />
                    </svg>
                  </div>
                  <div class="min-w-0 flex-1">
                    <p class="text-sm font-medium text-gray-900 dark:text-white">
                      {{ t('dashboard.createApiKey') }}
                    </p>
                    <p class="text-xs text-gray-500 dark:text-dark-400">
                      {{ t('dashboard.generateNewKey') }}
                    </p>
                  </div>
                  <svg
                    class="h-5 w-5 text-gray-400 transition-colors group-hover:text-primary-500 dark:text-dark-500"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    stroke-width="1.5"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M8.25 4.5l7.5 7.5-7.5 7.5"
                    />
                  </svg>
                </button>

                <button
                  @click="navigateTo('/usage')"
                  class="group flex w-full items-center gap-4 rounded-xl bg-gray-50 p-4 text-left transition-all duration-200 hover:bg-gray-100 dark:bg-dark-800/50 dark:hover:bg-dark-800"
                >
                  <div
                    class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-xl bg-emerald-100 transition-transform group-hover:scale-105 dark:bg-emerald-900/30"
                  >
                    <svg
                      class="h-6 w-6 text-emerald-600 dark:text-emerald-400"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      stroke-width="1.5"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z"
                      />
                    </svg>
                  </div>
                  <div class="min-w-0 flex-1">
                    <p class="text-sm font-medium text-gray-900 dark:text-white">
                      {{ t('dashboard.viewUsage') }}
                    </p>
                    <p class="text-xs text-gray-500 dark:text-dark-400">
                      {{ t('dashboard.checkDetailedLogs') }}
                    </p>
                  </div>
                  <svg
                    class="h-5 w-5 text-gray-400 transition-colors group-hover:text-emerald-500 dark:text-dark-500"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    stroke-width="1.5"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M8.25 4.5l7.5 7.5-7.5 7.5"
                    />
                  </svg>
                </button>

                <!-- Redeem code entry is not needed in this deployment -->
                <!--
                <button
                  @click="navigateTo('/redeem')"
                  class="group flex w-full items-center gap-4 rounded-xl bg-gray-50 p-4 text-left transition-all duration-200 hover:bg-gray-100 dark:bg-dark-800/50 dark:hover:bg-dark-800"
                >
                  <div
                    class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-xl bg-amber-100 transition-transform group-hover:scale-105 dark:bg-amber-900/30"
                  >
                    <svg
                      class="h-6 w-6 text-amber-600 dark:text-amber-400"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      stroke-width="1.5"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M21 11.25v8.25a1.5 1.5 0 01-1.5 1.5H5.25a1.5 1.5 0 01-1.5-1.5v-8.25M12 4.875A2.625 2.625 0 109.375 7.5H12m0-2.625V7.5m0-2.625A2.625 2.625 0 1114.625 7.5H12m0 0V21m-8.625-9.75h18c.621 0 1.125-.504 1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125h-18c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125z"
                      />
                    </svg>
                  </div>
                  <div class="min-w-0 flex-1">
                    <p class="text-sm font-medium text-gray-900 dark:text-white">
                      {{ t('dashboard.redeemCode') }}
                    </p>
                    <p class="text-xs text-gray-500 dark:text-dark-400">
                      {{ t('dashboard.addBalanceWithCode') }}
                    </p>
                  </div>
                  <svg
                    class="h-5 w-5 text-gray-400 transition-colors group-hover:text-amber-500 dark:text-dark-500"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    stroke-width="1.5"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M8.25 4.5l7.5 7.5-7.5 7.5"
                    />
                  </svg>
                </button>
                -->

                <button
                  @click="navigateTo('/payment')"
                  class="group flex w-full items-center gap-4 rounded-xl bg-gray-50 p-4 text-left transition-all duration-200 hover:bg-gray-100 dark:bg-dark-800/50 dark:hover:bg-dark-800"
                >
                  <div
                    class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-xl bg-primary-100 transition-transform group-hover:scale-105 dark:bg-primary-900/30"
                  >
                    <svg
                      class="h-6 w-6 text-primary-600 dark:text-primary-400"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      stroke-width="1.5"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M2.25 8.25h19.5M2.25 9h19.5m-16.5 5.25h6m-6 2.25h3m-3.75 3h15a2.25 2.25 0 002.25-2.25V6.75A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25v10.5A2.25 2.25 0 004.5 19.5z"
                      />
                    </svg>
                  </div>
                  <div class="min-w-0 flex-1">
                    <p class="text-sm font-medium text-gray-900 dark:text-white">
                      {{ t('dashboard.recharge') }}
                    </p>
                    <p class="text-xs text-gray-500 dark:text-dark-400">
                      {{ t('dashboard.topUpBalance') }}
                    </p>
                  </div>
                  <svg
                    class="h-5 w-5 text-gray-400 transition-colors group-hover:text-primary-500 dark:text-dark-500"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    stroke-width="1.5"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M8.25 4.5l7.5 7.5-7.5 7.5"
                    />
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'; import { useAuthStore } from '@/stores/auth'; import { usageAPI, type UserDashboardStats as UserStatsType } from '@/api/usage'
import AppLayout from '@/components/layout/AppLayout.vue'; import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import UserDashboardStats from '@/components/user/dashboard/UserDashboardStats.vue'; import UserDashboardCharts from '@/components/user/dashboard/UserDashboardCharts.vue'
import UserDashboardRecentUsage from '@/components/user/dashboard/UserDashboardRecentUsage.vue'; import UserDashboardQuickActions from '@/components/user/dashboard/UserDashboardQuickActions.vue'
import type { UsageLog, TrendDataPoint, ModelStat } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Select from '@/components/common/Select.vue'
import FirstRechargePromotion from '@/components/FirstRechargePromotion.vue'

const authStore = useAuthStore(); const user = computed(() => authStore.user)
const stats = ref<UserStatsType | null>(null); const loading = ref(false); const loadingUsage = ref(false); const loadingCharts = ref(false)
const trendData = ref<TrendDataPoint[]>([]); const modelStats = ref<ModelStat[]>([]); const recentUsage = ref<UsageLog[]>([])

const formatLD = (d: Date) => d.toISOString().split('T')[0]
const startDate = ref(formatLD(new Date(Date.now() - 6 * 86400000))); const endDate = ref(formatLD(new Date())); const granularity = ref('day')

const loadStats = async () => { loading.value = true; try { await authStore.refreshUser(); stats.value = await usageAPI.getDashboardStats() } catch (error) { console.error('Failed to load dashboard stats:', error) } finally { loading.value = false } }
const loadCharts = async () => { loadingCharts.value = true; try { const res = await Promise.all([usageAPI.getDashboardTrend({ start_date: startDate.value, end_date: endDate.value, granularity: granularity.value as any }), usageAPI.getDashboardModels({ start_date: startDate.value, end_date: endDate.value })]); trendData.value = res[0].trend || []; modelStats.value = res[1].models || [] } catch (error) { console.error('Failed to load charts:', error) } finally { loadingCharts.value = false } }
const loadRecent = async () => { loadingUsage.value = true; try { const res = await usageAPI.getByDateRange(startDate.value, endDate.value); recentUsage.value = res.items.slice(0, 5) } catch (error) { console.error('Failed to load recent usage:', error) } finally { loadingUsage.value = false } }

onMounted(() => { loadStats(); loadCharts(); loadRecent() })
</script>

<style scoped>
</style>
