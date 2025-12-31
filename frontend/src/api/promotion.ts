import { apiClient } from './client'
import type { PromotionStatusResponse } from '@/types'

export async function getPromotionStatus(): Promise<PromotionStatusResponse> {
  const { data } = await apiClient.get<PromotionStatusResponse>('/user/promotion/status')
  return data
}

export const promotionAPI = {
  getPromotionStatus
}

