/**
 * User Groups API endpoints (non-admin)
 * Handles group-related operations for regular users
 */

import { apiClient } from './client'
import type { Group } from '@/types'

/**
 * Get available groups that the current user can bind to API keys
 * This returns groups based on user's permissions:
 * - Standard groups: public (non-exclusive) or explicitly allowed
 * - Subscription groups: user has active subscription
 * @returns List of available groups
 */
export async function getAvailable(): Promise<Group[]> {
  const { data } = await apiClient.get<Group[]>('/groups/available')
  return data
}

/**
 * Get current user's custom group rate multipliers
 * @returns Map of group_id to custom rate_multiplier
 */
export async function getUserGroupRates(): Promise<Record<number, number>> {
  const { data } = await apiClient.get<Record<number, number> | null>('/groups/rates')
  return data || {}
}

export interface PublicGroupMonitorItem {
  group_id: number
  group_name: string
  platform: string
  current_status: 'normal' | 'abnormal' | 'unknown'
}

export interface PublicGroupMonitorResponse {
  generated_at: string
  public_group_num: number
  items: PublicGroupMonitorItem[]
}

export async function getPublicGroupMonitor(
): Promise<PublicGroupMonitorResponse> {
  const { data } = await apiClient.get<PublicGroupMonitorResponse>('/groups/monitor')
  return data
}

export const userGroupsAPI = {
  getAvailable,
  getUserGroupRates,
  getPublicGroupMonitor
}

export default userGroupsAPI
