/**
 * Admin Payment Orders (Recharge Records)
 */

import { apiClient } from '../client'
import type { PaginatedResponse } from '@/types'

export type AdminPaymentProvider = 'zpay' | 'stripe' | 'admin'
export type AdminPaymentMethod = 'alipay' | 'wechat'

export interface AdminPaymentOrder {
  id: number
  order_no: string
  order_type: string
  user_id: number
  provider: AdminPaymentProvider
  status: string
  amount_cny: number
  amount_usd: number
  total_usd: number
  created_at: string
  updated_at: string
  paid_at?: string | null
}

export async function list(
  page: number,
  pageSize: number,
  filters?: {
    method?: AdminPaymentMethod | ''
    user?: string
  }
): Promise<PaginatedResponse<AdminPaymentOrder>> {
  const { data } = await apiClient.get<PaginatedResponse<AdminPaymentOrder>>('/admin/payment/orders', {
    params: {
      page,
      page_size: pageSize,
      method: filters?.method || undefined,
      user: filters?.user || undefined
    }
  })
  return data
}

export async function exportRecords(filters?: { method?: AdminPaymentMethod | ''; user?: string }): Promise<Blob> {
  const { data } = await apiClient.get('/admin/payment/orders/export', {
    params: {
      method: filters?.method || undefined,
      user: filters?.user || undefined
    },
    responseType: 'blob'
  })
  return data as Blob
}

export default {
  list,
  exportRecords
}
