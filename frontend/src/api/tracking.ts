import { apiClient } from './client'
import { getBdLandingUrl, getBdVid } from '@/utils/baiduTracking'

export async function reportOcpcEvent(newType: number): Promise<void> {
  const bdVid = getBdVid()
  if (!bdVid) return

  await apiClient.post('/public/ocpc', {
    bd_vid: bdVid,
    landing_url: getBdLandingUrl(),
    new_type: newType
  })
}
