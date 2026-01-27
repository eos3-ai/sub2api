/**
 * Invoice APIs (User-side, MVP)
 *
 * - Eligible orders: GET /invoices/eligible-orders
 * - Create request: POST /invoices
 * - My requests: GET /invoices, GET /invoices/:id, POST /invoices/:id/cancel
 * - Default profile: GET/PUT /invoices/profile
 */

import { apiClient } from './client'
import type { PaginatedResponse } from '@/types'
import type { PaymentOrder } from './payment'

export type InvoiceType = 'normal' | 'special'
export type InvoiceBuyerType = 'personal' | 'company'
export type InvoiceStatus = 'submitted' | 'approved' | 'rejected' | 'issued' | 'cancelled'

export interface InvoiceRequest {
  id: number
  invoice_request_no: string
  user_id: number
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

export async function getEligibleInvoiceOrders(query?: {
  page?: number
  page_size?: number
  from?: string
  to?: string
}): Promise<PaginatedResponse<PaymentOrder>> {
  const { data } = await apiClient.get<PaginatedResponse<PaymentOrder>>('/invoices/eligible-orders', {
    params: query
  })
  return data
}

export async function getInvoiceProfile(): Promise<InvoiceProfile> {
  const { data } = await apiClient.get<InvoiceProfile>('/invoices/profile')
  return data
}

export async function updateInvoiceProfile(payload: {
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
}): Promise<InvoiceProfile> {
  const { data } = await apiClient.put<InvoiceProfile>('/invoices/profile', payload)
  return data
}

export async function createInvoiceRequest(payload: {
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
}): Promise<InvoiceRequest> {
  const { data } = await apiClient.post<InvoiceRequest>('/invoices', payload)
  return data
}

export async function getMyInvoiceRequests(query?: {
  page?: number
  page_size?: number
}): Promise<PaginatedResponse<InvoiceRequest>> {
  const { data } = await apiClient.get<PaginatedResponse<InvoiceRequest>>('/invoices', {
    params: query
  })
  return data
}

export async function getMyInvoiceRequest(id: number): Promise<{
  invoice: InvoiceRequest
  items: InvoiceOrderItem[]
}> {
  const { data } = await apiClient.get<{
    invoice: InvoiceRequest
    items: InvoiceOrderItem[]
  }>(`/invoices/${id}`)
  return data
}

export async function cancelInvoiceRequest(id: number): Promise<InvoiceRequest> {
  const { data } = await apiClient.post<InvoiceRequest>(`/invoices/${id}/cancel`)
  return data
}

export const invoiceAPI = {
  getEligibleInvoiceOrders,
  getInvoiceProfile,
  updateInvoiceProfile,
  createInvoiceRequest,
  getMyInvoiceRequests,
  getMyInvoiceRequest,
  cancelInvoiceRequest
}

export default invoiceAPI

