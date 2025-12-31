import { apiClient } from './client'
import type { ReferralInfoResponse, ReferralInvite, PaginatedResponse } from '@/types'

export async function getReferralInfo(): Promise<ReferralInfoResponse> {
  const { data } = await apiClient.get<ReferralInfoResponse>('/user/referral/info')
  return data
}

export async function listInvitees(page = 1, pageSize = 20): Promise<PaginatedResponse<ReferralInvite>> {
  const { data } = await apiClient.get<PaginatedResponse<ReferralInvite>>('/user/referral/invitees', {
    params: { page, page_size: pageSize }
  })
  return data
}

export const referralAPI = {
  getReferralInfo,
  listInvitees
}
