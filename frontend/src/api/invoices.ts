import { apiClient } from './client'
import type { BasePaginationResponse } from '@/types'

export type InvoiceType = 'normal' | 'special'
export type InvoiceBuyerType = 'personal' | 'company'
export type InvoiceStatus = 'submitted' | 'approved' | 'rejected' | 'issued' | 'cancelled'

export interface InvoiceEligibleOrder {
  id: number
  order_no: string
  out_trade_no?: string
  user_id: number
  user_email?: string
  amount_cny: number
  total_usd: number
  status: string
  paid_at?: string
  created_at: string
  updated_at: string
}

export interface InvoiceRequest {
  id: number
  invoice_request_no: string
  user_id: number
  user_email?: string
  status: InvoiceStatus
  invoice_type: InvoiceType
  buyer_type: InvoiceBuyerType
  invoice_title: string
  tax_no: string
  buyer_address: string
  buyer_phone: string
  buyer_bank_name: string
  buyer_bank_account: string
  receiver_email: string
  receiver_phone: string
  invoice_item_name: string
  remark: string
  amount_cny_total: number
  total_usd_total: number
  reviewed_by?: number
  reviewed_at?: string
  reject_reason?: string
  issued_by?: number
  issued_at?: string
  invoice_number?: string
  invoice_date?: string
  invoice_pdf_url?: string
  provider?: 'manual'
  created_at: string
  updated_at: string
}

export interface InvoiceOrderItem {
  id: number
  payment_order_id: number
  order_no: string
  amount_cny: number
  total_usd: number
  active: boolean
  paid_at?: string
  created_at: string
}

export interface InvoiceProfile {
  id: number
  user_id: number
  invoice_type: InvoiceType
  buyer_type: InvoiceBuyerType
  invoice_title: string
  tax_no: string
  buyer_address: string
  buyer_phone: string
  buyer_bank_name: string
  buyer_bank_account: string
  receiver_email: string
  receiver_phone: string
  invoice_item_name: string
  remark: string
  created_at: string
  updated_at: string
}

export interface InvoiceDetail {
  invoice: InvoiceRequest
  items: InvoiceOrderItem[]
}

export interface CreateInvoiceRequestPayload {
  order_nos: string[]
  invoice_type: InvoiceType
  buyer_type: InvoiceBuyerType
  invoice_title: string
  tax_no?: string
  buyer_address?: string
  buyer_phone?: string
  buyer_bank_name?: string
  buyer_bank_account?: string
  receiver_email: string
  receiver_phone?: string
  invoice_item_name?: string
  remark?: string
}

export const invoiceAPI = {
  getEligibleOrders(params?: { page?: number; page_size?: number; from?: string; to?: string }) {
    return apiClient.get<BasePaginationResponse<InvoiceEligibleOrder>>('/invoices/eligible-orders', { params })
  },
  getProfile() {
    return apiClient.get<InvoiceProfile>('/invoices/profile')
  },
  updateProfile(payload: Omit<InvoiceProfile, 'id' | 'user_id' | 'created_at' | 'updated_at'>) {
    return apiClient.put<InvoiceProfile>('/invoices/profile', payload)
  },
  createRequest(payload: CreateInvoiceRequestPayload) {
    return apiClient.post<InvoiceRequest>('/invoices', payload)
  },
  getMyRequests(params?: { page?: number; page_size?: number }) {
    return apiClient.get<BasePaginationResponse<InvoiceRequest>>('/invoices', { params })
  },
  getRequest(id: number) {
    return apiClient.get<InvoiceDetail>(`/invoices/${id}`)
  },
  cancelRequest(id: number) {
    return apiClient.post<InvoiceRequest>(`/invoices/${id}/cancel`)
  }
}

export default invoiceAPI
