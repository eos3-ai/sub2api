/**
 * Admin Invoices (Invoice Requests Management)
 */

import { apiClient } from '../client'
import type { PaginatedResponse } from '@/types'

export type AdminInvoiceStatus = 'submitted' | 'approved' | 'rejected' | 'issued' | 'cancelled'
export type AdminInvoiceType = 'normal' | 'special'
export type AdminInvoiceBuyerType = 'personal' | 'company'

export interface AdminInvoiceRequest {
  id: number
  invoice_request_no: string
  user_id: number
  user_email?: string
  status: AdminInvoiceStatus

  invoice_type: AdminInvoiceType
  buyer_type: AdminInvoiceBuyerType
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

export interface AdminInvoiceOrderItem {
  id: number
  payment_order_id: number
  order_no: string
  amount_cny: number
  total_usd: number
  active: boolean
  paid_at?: string
  created_at: string
}

export async function list(
  page: number,
  pageSize: number,
  filters?: {
    status?: AdminInvoiceStatus | ''
    user_email?: string
    from?: string
    to?: string
  }
): Promise<PaginatedResponse<AdminInvoiceRequest>> {
  const { data } = await apiClient.get<PaginatedResponse<AdminInvoiceRequest>>('/admin/invoices', {
    params: {
      page,
      page_size: pageSize,
      status: filters?.status || undefined,
      user_email: filters?.user_email || undefined,
      from: filters?.from || undefined,
      to: filters?.to || undefined
    }
  })
  return data
}

export async function getByID(id: number): Promise<{
  invoice: AdminInvoiceRequest
  items: AdminInvoiceOrderItem[]
}> {
  const { data } = await apiClient.get<{
    invoice: AdminInvoiceRequest
    items: AdminInvoiceOrderItem[]
  }>(`/admin/invoices/${id}`)
  return data
}

export async function approve(id: number): Promise<AdminInvoiceRequest> {
  const { data } = await apiClient.post<AdminInvoiceRequest>(`/admin/invoices/${id}/approve`)
  return data
}

export async function reject(id: number, rejectReason: string): Promise<AdminInvoiceRequest> {
  const { data } = await apiClient.post<AdminInvoiceRequest>(`/admin/invoices/${id}/reject`, {
    reject_reason: rejectReason
  })
  return data
}

export async function issue(
  id: number,
  payload: {
    invoice_number: string
    invoice_date?: string
    invoice_pdf_url: string
  }
): Promise<AdminInvoiceRequest> {
  const { data } = await apiClient.post<AdminInvoiceRequest>(`/admin/invoices/${id}/issue`, payload)
  return data
}

export async function exportRecords(filters?: {
  status?: AdminInvoiceStatus | ''
  user_email?: string
  from?: string
  to?: string
}): Promise<Blob> {
  const { data } = await apiClient.get('/admin/invoices/export', {
    params: {
      status: filters?.status || undefined,
      user_email: filters?.user_email || undefined,
      from: filters?.from || undefined,
      to: filters?.to || undefined
    },
    responseType: 'blob'
  })
  return data as Blob
}

export const invoicesAPI = {
  list,
  getByID,
  approve,
  reject,
  issue,
  exportRecords
}

export default invoicesAPI
