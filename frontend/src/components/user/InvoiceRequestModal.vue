<template>
  <BaseDialog :show="show" title="申请发票" width="extra-wide" close-on-click-outside @close="close">
    <div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_360px]">
      <section class="space-y-3">
        <div class="flex items-center justify-between gap-3">
          <h3 class="text-sm font-semibold text-gray-900 dark:text-white">可开票订单</h3>
          <button class="btn btn-secondary btn-sm" :disabled="loadingOrders" @click="loadEligibleOrders">
            {{ loadingOrders ? '加载中' : '刷新' }}
          </button>
        </div>

        <div v-if="loadingOrders" class="flex justify-center py-10">
          <LoadingSpinner />
        </div>

        <div v-else-if="eligibleOrders.length === 0" class="rounded-lg border border-gray-200 p-6 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
          暂无可开票订单
        </div>

        <div v-else class="overflow-x-auto rounded-lg border border-gray-200 dark:border-dark-700">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
            <thead class="bg-gray-50 dark:bg-dark-800">
              <tr>
                <th class="w-12 px-4 py-3 text-left text-xs font-semibold text-gray-500">选择</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-gray-500">订单号</th>
                <th class="px-4 py-3 text-right text-xs font-semibold text-gray-500">支付金额</th>
                <th class="px-4 py-3 text-right text-xs font-semibold text-gray-500">余额/套餐额</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-gray-500">支付时间</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 bg-white dark:divide-dark-800 dark:bg-dark-900">
              <tr v-for="order in eligibleOrders" :key="order.order_no">
                <td class="px-4 py-3">
                  <input
                    type="checkbox"
                    class="h-4 w-4 rounded border-gray-300 text-primary-600"
                    :checked="selectedOrderNos.includes(order.order_no)"
                    :disabled="!selectedOrderNos.includes(order.order_no) && selectedOrderNos.length >= 5"
                    @change="toggleOrder(order.order_no)"
                  />
                </td>
                <td class="px-4 py-3 font-mono text-xs text-gray-900 dark:text-white">{{ order.order_no }}</td>
                <td class="px-4 py-3 text-right text-sm text-gray-700 dark:text-dark-300">¥{{ order.amount_cny.toFixed(2) }}</td>
                <td class="px-4 py-3 text-right text-sm text-gray-700 dark:text-dark-300">${{ order.total_usd.toFixed(2) }}</td>
                <td class="px-4 py-3 text-sm text-gray-500 dark:text-dark-400">{{ formatDateTime(order.paid_at || order.created_at) }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="grid gap-3 sm:grid-cols-3">
          <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-800">
            <p class="text-xs text-gray-500">已选订单</p>
            <p class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">{{ selectedOrderNos.length }}/5</p>
          </div>
          <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-800">
            <p class="text-xs text-gray-500">开票金额</p>
            <p class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">¥{{ selectedAmountCNY.toFixed(2) }}</p>
          </div>
          <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-800">
            <p class="text-xs text-gray-500">余额/套餐额</p>
            <p class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">${{ selectedTotalUSD.toFixed(2) }}</p>
          </div>
        </div>
      </section>

      <section class="space-y-4">
        <div class="grid grid-cols-2 gap-3">
          <label class="block">
            <span class="input-label">发票类型</span>
            <select v-model="form.invoice_type" class="input">
              <option value="normal">电子普票</option>
              <option value="special">专票</option>
            </select>
          </label>
          <label class="block">
            <span class="input-label">抬头类型</span>
            <select v-model="form.buyer_type" class="input" :disabled="form.invoice_type === 'special'">
              <option value="company">企业</option>
              <option value="personal">个人</option>
            </select>
          </label>
        </div>

        <label class="block">
          <span class="input-label">发票抬头</span>
          <input v-model.trim="form.invoice_title" class="input" />
        </label>
        <label class="block">
          <span class="input-label">税号</span>
          <input v-model.trim="form.tax_no" class="input" />
        </label>

        <div v-if="form.invoice_type === 'special'" class="space-y-3">
          <input v-model.trim="form.buyer_address" class="input" placeholder="注册地址" />
          <input v-model.trim="form.buyer_phone" class="input" placeholder="注册电话" />
          <input v-model.trim="form.buyer_bank_name" class="input" placeholder="开户行" />
          <input v-model.trim="form.buyer_bank_account" class="input" placeholder="银行账号" />
        </div>

        <label class="block">
          <span class="input-label">收票邮箱</span>
          <input v-model.trim="form.receiver_email" class="input" type="email" />
        </label>
        <label class="block">
          <span class="input-label">收票手机号</span>
          <input v-model.trim="form.receiver_phone" class="input" />
        </label>
        <label class="block">
          <span class="input-label">开票内容</span>
          <input v-model.trim="form.invoice_item_name" class="input" />
        </label>
        <label class="block">
          <span class="input-label">备注</span>
          <textarea v-model.trim="form.remark" rows="3" class="input" />
        </label>
      </section>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button class="btn btn-secondary" @click="close">取消</button>
        <button class="btn btn-primary" :disabled="submitting || !canSubmit" @click="submit">
          {{ submitting ? '提交中' : '提交申请' }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { useAppStore } from '@/stores'
import { formatDateTime } from '@/utils/format'
import { invoiceAPI, type InvoiceBuyerType, type InvoiceEligibleOrder, type InvoiceRequest, type InvoiceType } from '@/api/invoices'

const props = defineProps<{ show: boolean }>()
const emit = defineEmits<{
  (e: 'close'): void
  (e: 'created', invoice: InvoiceRequest): void
}>()

const appStore = useAppStore()
const loadingOrders = ref(false)
const submitting = ref(false)
const eligibleOrders = ref<InvoiceEligibleOrder[]>([])
const selectedOrderNos = ref<string[]>([])

const form = reactive({
  invoice_type: 'normal' as InvoiceType,
  buyer_type: 'company' as InvoiceBuyerType,
  invoice_title: '',
  tax_no: '',
  buyer_address: '',
  buyer_phone: '',
  buyer_bank_name: '',
  buyer_bank_account: '',
  receiver_email: '',
  receiver_phone: '',
  invoice_item_name: '',
  remark: ''
})

watch(
  () => props.show,
  async (show) => {
    if (!show) return
    selectedOrderNos.value = []
    await Promise.all([loadEligibleOrders(), loadProfile()])
  }
)

watch(
  () => form.invoice_type,
  (value) => {
    if (value === 'special') form.buyer_type = 'company'
  }
)

const selectedOrders = computed(() => {
  const selected = new Set(selectedOrderNos.value)
  return eligibleOrders.value.filter((order) => selected.has(order.order_no))
})

const selectedAmountCNY = computed(() => selectedOrders.value.reduce((sum, order) => sum + order.amount_cny, 0))
const selectedTotalUSD = computed(() => selectedOrders.value.reduce((sum, order) => sum + order.total_usd, 0))

const canSubmit = computed(() => {
  if (selectedOrderNos.value.length === 0) return false
  if (!form.invoice_title || !form.receiver_email.includes('@')) return false
  if (form.buyer_type === 'company' && !form.tax_no) return false
  if (form.invoice_type === 'special') {
    return Boolean(form.buyer_address && form.buyer_phone && form.buyer_bank_name && form.buyer_bank_account)
  }
  return true
})

async function loadEligibleOrders() {
  loadingOrders.value = true
  try {
    const { data } = await invoiceAPI.getEligibleOrders({ page: 1, page_size: 100 })
    eligibleOrders.value = data.items || []
  } catch (error) {
    eligibleOrders.value = []
    appStore.showError((error as { message?: string })?.message || '加载可开票订单失败')
  } finally {
    loadingOrders.value = false
  }
}

async function loadProfile() {
  try {
    const { data } = await invoiceAPI.getProfile()
    Object.assign(form, {
      invoice_type: data.invoice_type || 'normal',
      buyer_type: data.buyer_type || 'company',
      invoice_title: data.invoice_title || '',
      tax_no: data.tax_no || '',
      buyer_address: data.buyer_address || '',
      buyer_phone: data.buyer_phone || '',
      buyer_bank_name: data.buyer_bank_name || '',
      buyer_bank_account: data.buyer_bank_account || '',
      receiver_email: data.receiver_email || '',
      receiver_phone: data.receiver_phone || '',
      invoice_item_name: data.invoice_item_name || '',
      remark: data.remark || ''
    })
  } catch {
    form.invoice_item_name = form.invoice_item_name || '技术服务费'
  }
}

function toggleOrder(orderNo: string) {
  if (selectedOrderNos.value.includes(orderNo)) {
    selectedOrderNos.value = selectedOrderNos.value.filter((item) => item !== orderNo)
    return
  }
  if (selectedOrderNos.value.length < 5) {
    selectedOrderNos.value.push(orderNo)
  }
}

async function submit() {
  if (!canSubmit.value) return
  submitting.value = true
  try {
    const { data } = await invoiceAPI.createRequest({
      order_nos: selectedOrderNos.value,
      ...form
    })
    appStore.showSuccess('开票申请已提交')
    emit('created', data)
    close()
  } catch (error) {
    appStore.showError((error as { message?: string })?.message || '提交开票申请失败')
  } finally {
    submitting.value = false
  }
}

function close() {
  emit('close')
}
</script>
