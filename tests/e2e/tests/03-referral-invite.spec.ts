import { test, expect } from '@playwright/test'
import { loadEnvConfig } from '../src/config'
import { completeRegisterOnCurrentPage, createRechargeOrder, login, uniqueEmail } from '../src/flows'

async function readReferralTotalInvites(page: import('@playwright/test').Page): Promise<number> {
  await page.goto('/referral')
  await expect(page.locator('text=分享获取额度')).toBeVisible({ timeout: 30_000 })

  // The first "text-4xl font-bold" in the stats grid is total invites.
  const valueText = await page.locator('div.text-4xl.font-bold').first().innerText()
  const n = Number.parseInt(valueText.replace(/[^\d]/g, ''), 10)
  return Number.isFinite(n) ? n : 0
}

test('使用邀请链接注册并发起充值（验证邀请关系建立）', async ({ browser, page }) => {
  const env = loadEnvConfig()

  // Inviter: login and capture invite link.
  await login(page, env.accounts.user)
  await page.goto('/referral')
  await expect(page.locator('text=邀请返利')).toBeVisible({ timeout: 30_000 })

  const totalBefore = await readReferralTotalInvites(page)

  const inviteLinkText = await page
    .locator('span.block.truncate')
    .filter({ hasText: /^https?:\/\// })
    .first()
    .innerText()

  const inviteLink = inviteLinkText.trim()
  expect(inviteLink).toMatch(/^https?:\/\//)

  // Invitee: register via invite link in a fresh context.
  const inviteeContext = await browser.newContext({
    baseURL: env.baseURL,
    ignoreHTTPSErrors: true,
    locale: 'zh-CN',
    timezoneId: 'Asia/Shanghai'
  })
  const inviteePage = await inviteeContext.newPage()

  const inviteeEmail = uniqueEmail(env.registration.emailDomain)
  await inviteePage.goto(inviteLink)
  await completeRegisterOnCurrentPage(inviteePage, { email: inviteeEmail, password: env.registration.password })

  // Invitee: create a recharge order (payment success & referral bonus issuance require real payment/callback).
  await createRechargeOrder(inviteePage, { method: 'alipay', allowNoQRCode: env.features.allowNoQRCode })
  await inviteeContext.close()

  // Inviter: referral stats should reflect the new invitee registration.
  await expect.poll(async () => readReferralTotalInvites(page), { timeout: 30_000 }).toBe(totalBefore + 1)

  if (env.features.requireReferralReward) {
    throw new Error(
      'This environment requires referral reward verification, but paid-recharge automation (payment success callback simulation) is not configured.'
    )
  }
})
