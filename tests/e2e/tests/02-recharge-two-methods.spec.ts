import { test } from '@playwright/test'
import { loadEnvConfig } from '../src/config'
import { createRechargeOrder, expectFirstRechargePromotionVisible, register, uniqueEmail } from '../src/flows'

test('两种方式充值：支付宝/微信（含活动赠送展示）', async ({ page }) => {
  const env = loadEnvConfig()
  const email = uniqueEmail(env.registration.emailDomain)

  await register(page, { email, password: env.registration.password })

  await page.goto('/payment')
  await expectFirstRechargePromotionVisible(page)

  await createRechargeOrder(page, { method: 'alipay', allowNoQRCode: env.features.allowNoQRCode })
  await createRechargeOrder(page, { method: 'wechat', allowNoQRCode: env.features.allowNoQRCode })
})
