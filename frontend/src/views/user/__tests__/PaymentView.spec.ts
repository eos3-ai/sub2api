import { mount, VueWrapper } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises } from '@vue/test-utils'
import PaymentView from '../PaymentView.vue'

const routerReplace = vi.hoisted(() => vi.fn())

const appStoreState = vi.hoisted(() => ({
  cachedPublicSettings: { payment_enabled: true } as { payment_enabled?: boolean } | undefined
}))
const fetchPublicSettings = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())
const showInfo = vi.hoisted(() => vi.fn())
const showSuccess = vi.hoisted(() => vi.fn())
const showWarning = vi.hoisted(() => vi.fn())

const getPaymentPlans = vi.hoisted(() => vi.fn())
const getSubscriptionPlans = vi.hoisted(() => vi.fn())
const getMyPaymentOrders = vi.hoisted(() => vi.fn())
const createPaymentOrder = vi.hoisted(() => vi.fn())
const getPaymentOrder = vi.hoisted(() => vi.fn())
const qrToDataURL = vi.hoisted(() => vi.fn())

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({
      replace: routerReplace
    })
  }
})

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    cachedPublicSettings: appStoreState.cachedPublicSettings,
    fetchPublicSettings,
    showError,
    showInfo,
    showSuccess,
    showWarning
  })
}))

vi.mock('@/api/payment', () => ({
  paymentAPI: {
    getPaymentPlans,
    getSubscriptionPlans,
    getMyPaymentOrders,
    createPaymentOrder,
    getPaymentOrder
  }
}))

vi.mock('qrcode', () => ({
  default: {
    toDataURL: qrToDataURL
  }
}))

type PaymentChannelOption = {
  provider: 'zpay' | 'stripe'
  method: 'alipay' | 'wechat'
  channel: 'zpay_alipay' | 'zpay_wechat' | 'stripe_alipay' | 'stripe_wechat'
}

type PaymentPlanFixture = {
  id: string
  name: string
  amount_usd: number
  pay_usd: number
  credits_usd: number
  exchange_rate: number
  discount_rate: number
  available_channels?: Array<'alipay' | 'wechat'>
  available_channel_options?: PaymentChannelOption[]
}

function planFixture(overrides: Partial<PaymentPlanFixture> = {}): PaymentPlanFixture {
  return {
    id: 'pkg_0',
    name: 'Starter',
    amount_usd: 10,
    pay_usd: 10,
    credits_usd: 10,
    exchange_rate: 7.2,
    discount_rate: 1,
    available_channels: [],
    available_channel_options: [],
    ...overrides
  }
}

function orderFixture(overrides: Record<string, unknown> = {}) {
  return {
    id: 1,
    order_no: 'pay_1',
    order_type: 'online_recharge',
    provider: 'stripe',
    channel: 'wechat',
    amount_cny: 72,
    amount_usd: 10,
    total_usd: 10,
    status: 'pending',
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    ...overrides
  }
}

async function mountPaymentView(plans: PaymentPlanFixture[] = [planFixture()]) {
  getPaymentPlans.mockResolvedValue(plans)
  getSubscriptionPlans.mockResolvedValue([])
  getMyPaymentOrders.mockResolvedValue({ items: [], total: 0, page: 1, page_size: 20 })

  const wrapper = mount(PaymentView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        FirstRechargePromotion: { template: '<div />' },
        LoadingSpinner: { template: '<div />' },
        Modal: {
          props: ['show', 'title', 'size', 'closeOnClickOutside'],
          template: '<div v-if="show"><slot /><slot name="footer" /></div>'
        },
        Icon: { template: '<span />' },
        InvoiceRequestModal: { template: '<div />' }
      }
    }
  })

  await flushPromises()
  await flushPromises()
  return wrapper
}

function findButtonByText(wrapper: VueWrapper, text: string) {
  const button = wrapper.findAll('button').find((item) => item.text().includes(text))
  if (!button) {
    throw new Error(`Button containing "${text}" was not found`)
  }
  return button
}

describe('PaymentView legacy channel switches', () => {
  beforeEach(() => {
    vi.useRealTimers()
    appStoreState.cachedPublicSettings = { payment_enabled: true }
    fetchPublicSettings.mockReset().mockResolvedValue(appStoreState.cachedPublicSettings)
    routerReplace.mockReset().mockResolvedValue(undefined)
    showError.mockReset()
    showInfo.mockReset()
    showSuccess.mockReset()
    showWarning.mockReset()
    getPaymentPlans.mockReset()
    getSubscriptionPlans.mockReset()
    getMyPaymentOrders.mockReset()
    createPaymentOrder.mockReset().mockResolvedValue({
      order: orderFixture(),
      pay_url: 'weixin://pay'
    })
    getPaymentOrder.mockReset().mockResolvedValue(orderFixture())
    qrToDataURL.mockReset().mockResolvedValue('data:image/png;base64,qr')
  })

  it('redirects away when the legacy PAYMENT_ENABLED public flag is false', async () => {
    appStoreState.cachedPublicSettings = { payment_enabled: false }

    await mountPaymentView()

    expect(routerReplace).toHaveBeenCalledWith('/dashboard')
    expect(getPaymentPlans).not.toHaveBeenCalled()
    expect(getSubscriptionPlans).not.toHaveBeenCalled()
    expect(getMyPaymentOrders).not.toHaveBeenCalled()
  })

  it('keeps the page visible but disables payment when no provider channel is enabled', async () => {
    const wrapper = await mountPaymentView([
      planFixture({
        available_channel_options: [],
        available_channels: []
      })
    ])

    await findButtonByText(wrapper, '$10').trigger('click')

    expect(wrapper.text()).toContain('payment.noPayMethodEnabled')
    expect(findButtonByText(wrapper, 'payment.rechargeNow').attributes('disabled')).toBeDefined()
  })

  it('renders enabled provider and method combinations with provider labels', async () => {
    const wrapper = await mountPaymentView([
      planFixture({
        available_channel_options: [
          { provider: 'zpay', method: 'alipay', channel: 'zpay_alipay' },
          { provider: 'zpay', method: 'wechat', channel: 'zpay_wechat' },
          { provider: 'stripe', method: 'wechat', channel: 'stripe_wechat' }
        ]
      })
    ])

    expect(wrapper.text()).toContain('payment.alipay\uFF08zpay\uFF09')
    expect(wrapper.text()).toContain('payment.wechat\uFF08zpay\uFF09')
    expect(wrapper.text()).toContain('payment.wechat\uFF08stripe\uFF09')
  })

  it('creates orders with the selected provider and payment method channel', async () => {
    const wrapper = await mountPaymentView([
      planFixture({
        available_channel_options: [
          { provider: 'zpay', method: 'alipay', channel: 'zpay_alipay' },
          { provider: 'stripe', method: 'wechat', channel: 'stripe_wechat' }
        ]
      })
    ])

    await findButtonByText(wrapper, '$10').trigger('click')
    await findButtonByText(wrapper, 'payment.wechat\uFF08stripe\uFF09').trigger('click')
    await findButtonByText(wrapper, 'payment.rechargeNow').trigger('click')
    await flushPromises()

    expect(createPaymentOrder).toHaveBeenCalledWith({
      plan_id: 'pkg_0',
      channel: 'stripe_wechat'
    })
    wrapper.unmount()
  })

  it('falls back to legacy available_channels as zpay payment methods', async () => {
    const wrapper = await mountPaymentView([
      planFixture({
        available_channel_options: undefined,
        available_channels: ['wechat']
      })
    ])

    await findButtonByText(wrapper, '$10').trigger('click')
    await findButtonByText(wrapper, 'payment.rechargeNow').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('payment.wechat\uFF08zpay\uFF09')
    expect(createPaymentOrder).toHaveBeenCalledWith({
      plan_id: 'pkg_0',
      channel: 'zpay_wechat'
    })
    wrapper.unmount()
  })
})
