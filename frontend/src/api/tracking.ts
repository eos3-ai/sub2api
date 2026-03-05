/**
 * Tracking API — 百度 OCPC 转化上报（公开接口，无需认证）
 */

import { apiClient } from './client'
import { getBdVid, getBdLandingUrl } from '@/utils/baiduTracking'

/**
 * 上报百度 OCPC 转化事件
 * @param newType 转化类型：1=关键页面浏览
 */
export function reportOcpcEvent(newType: number): Promise<void> {
  const bdVid = getBdVid()
  if (!bdVid) return Promise.resolve()
  return apiClient.post('/api/v1/public/ocpc', {
    bd_vid: bdVid,
    landing_url: getBdLandingUrl(),
    new_type: newType
  })
}
