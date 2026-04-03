/**
 * Payment API endpoints (migration WIP)
 * Provides typed wrappers for payment-related APIs.
 *
 * Note: Backend routes may not be enabled yet; callers should handle 404 gracefully.
 */

import { apiClient } from './client'
import type { PaginatedResponse } from '@/types'

export type PaymentChannel = 'zpay' | 'stripe' | 'admin' | 'activity'
export type PaymentPayMethod = 'alipay' | 'wechat'
export type PaymentCreateChannel = PaymentChannel | PaymentPayMethod

export interface PaymentPlan {
  id: string
  name: string
  amount_usd: number
  pay_usd: number
  credits_usd: number
  exchange_rate: number
  discount_rate: number
  available_channels?: PaymentPayMethod[]
  description?: string
  enabled?: boolean
}

export interface PaymentSubscriptionPlan {
  group_id: number
  group_name: string
  description?: string
  platform: string
  price_usd: number
  validity_days: number
  daily_limit_usd?: number | null
  weekly_limit_usd?: number | null
  monthly_limit_usd?: number | null
  exchange_rate: number
  available_channels?: PaymentPayMethod[]
  has_active_subscription?: boolean
}

export type PaymentOrderStatus = 'pending' | 'paid' | 'expired' | 'cancelled' | 'failed' | 'refunded'

export interface PaymentOrder {
  id: number
  order_no: string
  order_type: string
  provider: PaymentChannel
  channel?: string // 实际支付渠道（alipay/wechat）
  biz_group_id?: number
  biz_validity_days?: number
  remark?: string
  amount_cny: number
  amount_usd: number
  total_usd: number
  status: PaymentOrderStatus
  paid_at?: string
  created_at: string
  updated_at: string
  expire_at?: string
}

export async function getPaymentPlans(): Promise<PaymentPlan[]> {
  const { data } = await apiClient.get<PaymentPlan[]>('/payment/plans')
  return data
}

export async function createPaymentOrder(payload: {
  plan_id?: string
  amount_usd?: number
  subscription_group_id?: number
  channel: PaymentCreateChannel
}): Promise<{
  order: PaymentOrder
  pay_url?: string
  qr_url?: string
}> {
  const { data } = await apiClient.post<{
    order: PaymentOrder
    pay_url?: string
    qr_url?: string
  }>('/payment/orders', payload)
  return data
}

export async function getSubscriptionPlans(): Promise<PaymentSubscriptionPlan[]> {
  const { data } = await apiClient.get<PaymentSubscriptionPlan[]>('/payment/subscription-plans')
  return data
}

export async function getMyPaymentOrders(query?: {
  page?: number
  page_size?: number
  status?: PaymentOrderStatus
}): Promise<PaginatedResponse<PaymentOrder>> {
  const { data } = await apiClient.get<PaginatedResponse<PaymentOrder>>('/payment/orders', {
    params: query
  })
  return data
}

export async function getPaymentOrder(orderNo: string): Promise<PaymentOrder> {
  const { data } = await apiClient.get<PaymentOrder>(`/payment/orders/${orderNo}`)
  return data
}

export async function cancelPaymentOrder(orderNo: string): Promise<{ message: string }> {
  const { data } = await apiClient.post<{ message: string }>(`/payment/orders/${orderNo}/cancel`)
  return data
}

export const paymentAPI = {
  getPaymentPlans,
  getSubscriptionPlans,
  createPaymentOrder,
  getMyPaymentOrders,
  getPaymentOrder,
  cancelPaymentOrder
}

export default paymentAPI
