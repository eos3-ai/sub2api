import { test, expect } from '@playwright/test'
import { loadEnvConfig } from '../src/config'
import { expectFirstRechargePromotionVisible, register, uniqueEmail } from '../src/flows'

test('注册账户后活动页正常展示（新人首充优惠）', async ({ page }) => {
  const env = loadEnvConfig()
  const email = uniqueEmail(env.registration.emailDomain)

  await register(page, { email, password: env.registration.password })

  await page.goto('/payment')
  await expect(page.locator('text=在线充值')).toBeVisible({ timeout: 30_000 })
  await expectFirstRechargePromotionVisible(page)
})
