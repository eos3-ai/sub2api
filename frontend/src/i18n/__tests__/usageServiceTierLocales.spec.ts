import { describe, expect, it } from 'vitest'

import en from '../locales/en'
import zh from '../locales/zh'

describe('usage service tier locale keys', () => {
  it('contains zh labels for service tier tooltip', () => {
    expect(zh.usage.serviceTier).toBe('服务档位')
    expect(zh.usage.serviceTierPriority).toBe('Fast')
    expect(zh.usage.serviceTierFlex).toBe('Flex')
    expect(zh.usage.serviceTierStandard).toBe('Standard')
  })

  it('contains en labels for service tier tooltip', () => {
    expect(en.usage.serviceTier).toBe('Service tier')
    expect(en.usage.serviceTierPriority).toBe('Fast')
    expect(en.usage.serviceTierFlex).toBe('Flex')
    expect(en.usage.serviceTierStandard).toBe('Standard')
  })

  it('contains visible admin usage filter labels in zh', () => {
    expect(zh.usage.ws).toBe('WebSocket')
    expect(zh.admin.usage.account).toBe('账号')
    expect(zh.admin.usage.billingMode).toBe('计费模式')
    expect(zh.admin.usage.allBillingModes).toBe('全部计费模式')
    expect(zh.admin.usage.billingModeToken).toBe('按 Token')
    expect(zh.admin.usage.billingModePerRequest).toBe('按次')
    expect(zh.admin.usage.billingModeImage).toBe('按图片')
  })

  it('contains visible admin usage filter labels in en', () => {
    expect(en.usage.ws).toBe('WebSocket')
    expect(en.admin.usage.account).toBe('Account')
    expect(en.admin.usage.billingMode).toBe('Billing Mode')
    expect(en.admin.usage.allBillingModes).toBe('All Billing Modes')
    expect(en.admin.usage.billingModeToken).toBe('Per Token')
    expect(en.admin.usage.billingModePerRequest).toBe('Per Request')
    expect(en.admin.usage.billingModeImage).toBe('Per Image')
  })
})
