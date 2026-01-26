/**
 * Admin Payment Orders (Recharge Records)
 */

import { apiClient } from '../client'
import type { PaginatedResponse } from '@/types'

export type AdminPaymentProvider = 'zpay' | 'stripe' | 'admin' | 'activity'
export type AdminPaymentOrderType = 'online_recharge' | 'admin_recharge' | 'activity_recharge'

export interface AdminPaymentOrder {
  id: number
  order_no: string
  order_type: string
  user_id: number
  user_email?: string
  provider: AdminPaymentProvider
  channel?: string  // 实际支付渠道（alipay/wechat）
  status: string
  amount_cny: number
  amount_usd: number
  total_usd: number
  created_at: string
  updated_at: string
  paid_at?: string | null
}

export interface AdminPaymentOrdersSummary {
  total_usd: number
  amount_cny: number
}

export async function list(
  page: number,
  pageSize: number,
  filters?: {
    orderType?: AdminPaymentOrderType | ''
    user?: string
    status?: string
    from?: string
    to?: string
  }
): Promise<PaginatedResponse<AdminPaymentOrder>> {
  const { data } = await apiClient.get<PaginatedResponse<AdminPaymentOrder>>('/admin/payment/orders', {
    params: {
      page,
      page_size: pageSize,
      order_type: filters?.orderType || undefined,
      user: filters?.user || undefined,
      status: filters?.status || undefined,
      from: filters?.from || undefined,
      to: filters?.to || undefined
    }
  })
  return data
}

export async function summary(filters?: {
  orderType?: AdminPaymentOrderType | ''
  user?: string
  status?: string
  from?: string
  to?: string
}): Promise<AdminPaymentOrdersSummary> {
  const { data } = await apiClient.get<AdminPaymentOrdersSummary>('/admin/payment/orders/summary', {
    params: {
      order_type: filters?.orderType || undefined,
      user: filters?.user || undefined,
      status: filters?.status || undefined,
      from: filters?.from || undefined,
      to: filters?.to || undefined
    }
  })
  return data
}

export async function exportRecords(filters?: {
  orderType?: AdminPaymentOrderType | ''
  user?: string
  status?: string
  from?: string
  to?: string
}): Promise<Blob> {
  const { data } = await apiClient.get('/admin/payment/orders/export', {
    params: {
      order_type: filters?.orderType || undefined,
      user: filters?.user || undefined,
      status: filters?.status || undefined,
      from: filters?.from || undefined,
      to: filters?.to || undefined
    },
    responseType: 'blob'
  })
  return data as Blob
}

export default {
  list,
  summary,
  exportRecords
}
