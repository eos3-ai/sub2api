import { apiClient } from '../client'
import type { BasePaginationResponse } from '@/types'
import type { InvoiceDetail, InvoiceRequest, InvoiceStatus } from '@/api/invoices'

export const adminInvoicesAPI = {
  list(params?: {
    page?: number
    page_size?: number
    status?: InvoiceStatus | ''
    user_email?: string
    from?: string
    to?: string
  }) {
    return apiClient.get<BasePaginationResponse<InvoiceRequest>>('/admin/invoices', { params })
  },
  export(params?: { status?: InvoiceStatus | ''; user_email?: string; from?: string; to?: string }) {
    return apiClient.get<Blob>('/admin/invoices/export', { params, responseType: 'blob' })
  },
  getByID(id: number) {
    return apiClient.get<InvoiceDetail>(`/admin/invoices/${id}`)
  },
  approve(id: number) {
    return apiClient.post<InvoiceRequest>(`/admin/invoices/${id}/approve`)
  },
  reject(id: number, rejectReason: string) {
    return apiClient.post<InvoiceRequest>(`/admin/invoices/${id}/reject`, { reject_reason: rejectReason })
  },
  issue(id: number, payload: { invoice_number: string; invoice_date?: string; invoice_pdf_url?: string }) {
    return apiClient.post<InvoiceRequest>(`/admin/invoices/${id}/issue`, payload)
  }
}

export default adminInvoicesAPI
