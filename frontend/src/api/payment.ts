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
  description?: string
  enabled?: boolean
}

export type PaymentOrderStatus = 'pending' | 'paid' | 'expired' | 'cancelled' | 'failed'

export interface PaymentOrder {
  id: number
  order_no: string
  order_type: string
  provider: PaymentChannel
  remark?: string
  amount_cny: number
  amount_usd: number
  total_usd: number
  status: PaymentOrderStatus
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
  createPaymentOrder,
  getMyPaymentOrders,
  getPaymentOrder,
  cancelPaymentOrder
}

export default paymentAPI
