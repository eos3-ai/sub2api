import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'

import OrderTable from '../OrderTable.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, fallback?: string) => fallback ?? key,
    }),
  }
})

function mockDesktopViewport() {
  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: vi.fn().mockImplementation((query: string) => ({
      matches: true,
      media: query,
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    })),
  })
}

describe('OrderTable', () => {
  beforeEach(() => {
    mockDesktopViewport()
  })

  it('renders legacy order rows with string numeric fields', () => {
    const wrapper = mount(OrderTable, {
      props: {
        loading: false,
        showUser: true,
        orders: [
          {
            id: 22,
            user_id: 1,
            user_email: 'admin@example.com',
            amount: '9.5',
            pay_amount: '10',
            fee_rate: '0',
            payment_type: 'wechat',
            out_trade_no: 'ORD_20260403034204539588239628373',
            status: 'PAID',
            order_type: 'balance',
            created_at: '2026-04-03T03:42:04Z',
            expires_at: '2026-04-03T04:12:04Z',
            refund_amount: 0,
          } as any,
        ],
      },
      global: {
        stubs: {
          Icon: true,
        },
      },
    })

    expect(wrapper.text()).toContain('¥10.00')
    expect(wrapper.text()).toContain('$9.50')
    expect(wrapper.text()).toContain('wechat')
  })
})
