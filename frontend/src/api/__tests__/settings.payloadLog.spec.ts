import { beforeEach, describe, expect, it, vi } from 'vitest'

const { get, put } = vi.hoisted(() => ({
  get: vi.fn(),
  put: vi.fn()
}))

vi.mock('@/api/client', () => ({
  apiClient: { get, put }
}))

import {
  getPayloadLogStatus,
  updatePayloadLogStatus
} from '@/api/admin/settings'

describe('admin payload log settings API', () => {
  beforeEach(() => {
    get.mockReset()
    put.mockReset()
  })

  it('reads the current capture status', async () => {
    const status = {
      enabled: false,
      store_available: true,
      store_error: '',
      dropped: 0
    }
    get.mockResolvedValueOnce({ data: status })

    await expect(getPayloadLogStatus()).resolves.toEqual(status)
    expect(get).toHaveBeenCalledWith('/admin/payload-logs/settings')
  })

  it('updates the runtime capture switch', async () => {
    const status = {
      enabled: true,
      store_available: true,
      store_error: '',
      dropped: 0
    }
    put.mockResolvedValueOnce({ data: status })

    await expect(updatePayloadLogStatus(true)).resolves.toEqual(status)
    expect(put).toHaveBeenCalledWith('/admin/payload-logs/settings', {
      enabled: true
    })
  })
})
