import { expect, type Page } from '@playwright/test'
import type { E2EAccountConfig } from './config'

export function uniqueEmail(emailDomain: string): string {
  const safeDomain = String(emailDomain || 'example.com')
    .trim()
    .replace(/^@/, '')
  const stamp = Date.now()
  const rand = Math.random().toString(36).slice(2, 8)
  return `e2e_${stamp}_${rand}@${safeDomain}`
}

export async function login(page: Page, account: E2EAccountConfig): Promise<void> {
  await page.goto('/login')

  await page.locator('#email').fill(account.email)
  await page.locator('#password').fill(account.password)

  const submit = page.locator('form button[type="submit"]')
  if (await submit.isDisabled().catch(() => false)) {
    throw new Error('Login submit button is disabled (captcha/turnstile might be enabled).')
  }
  await submit.click()

  await expect(page.locator('text=登录失败')).toHaveCount(0, { timeout: 15_000 })
  await page.waitForURL('**/dashboard', { timeout: 30_000 })
}

export async function register(page: Page, opts: { email: string; password: string; inviter?: string }): Promise<void> {
  const url = opts.inviter ? `/register?inviter=${encodeURIComponent(opts.inviter)}` : '/register'
  await page.goto(url)
  await completeRegisterOnCurrentPage(page, { email: opts.email, password: opts.password, inviter: opts.inviter })
}

export async function completeRegisterOnCurrentPage(
  page: Page,
  opts: { email: string; password: string; inviter?: string }
): Promise<void> {
  const registrationDisabled = page.locator('text=注册功能暂时关闭')
  if (await registrationDisabled.isVisible().catch(() => false)) {
    throw new Error('Registration is disabled in this environment (UI shows: 注册功能暂时关闭).')
  }

  await page.locator('#email').fill(opts.email)
  await page.locator('#password').fill(opts.password)

  if (opts.inviter) {
    const inviteInput = page.locator('#invite-code')
    if (await inviteInput.isVisible().catch(() => false)) {
      const current = (await inviteInput.inputValue().catch(() => '')) || ''
      if (!current.trim()) await inviteInput.fill(opts.inviter)
    }
  }

  const submit = page.locator('form button[type="submit"]')
  if (await submit.isDisabled().catch(() => false)) {
    throw new Error('Register submit button is disabled (captcha/turnstile might be enabled).')
  }
  await submit.click()

  const emailVerify = page.locator('text=验证您的邮箱')
  if (await emailVerify.isVisible().catch(() => false)) {
    throw new Error('Email verification is enabled; E2E registration flow needs mailbox integration.')
  }

  await page.waitForURL('**/dashboard', { timeout: 30_000 })
}

export async function gotoPayment(page: Page): Promise<void> {
  await page.goto('/payment')
  await expect(page.locator('#recharge-plans')).toBeVisible({ timeout: 30_000 })
}

export async function expectFirstRechargePromotionVisible(page: Page): Promise<void> {
  const promoTitle = page.locator('text=新用户首充限时优惠')
  await expect(promoTitle).toBeVisible({ timeout: 30_000 })

  await expect(page.locator('text=还剩')).toBeVisible()
  await expect(page.locator('text=/\\d{2}:\\d{2}:\\d{2}/')).toBeVisible()
}

export async function createRechargeOrder(
  page: Page,
  opts: { method: 'alipay' | 'wechat'; allowNoQRCode: boolean }
): Promise<{ orderNo: string }> {
  await gotoPayment(page)

  const apiNotEnabled = page.locator('text=支付接口未启用')
  if (await apiNotEnabled.isVisible().catch(() => false)) {
    throw new Error('Payment API is not enabled in this environment (UI shows: 支付接口未启用).')
  }

  // Choose a "safe" plan (largest USD) to reduce WeChat minimum-amount flakiness.
  const planButtons = page.locator('#recharge-plans button').filter({
    has: page.locator('p').filter({ hasText: /^\$/ })
  })
  const count = await planButtons.count()
  if (count === 0) throw new Error('No preset recharge plans found.')

  let bestIndex = 0
  let bestUsd = -1
  for (let i = 0; i < count; i += 1) {
    const text = (await planButtons.nth(i).locator('p').first().innerText()).trim()
    const usd = Number.parseFloat(text.replace(/^\$/, ''))
    if (Number.isFinite(usd) && usd > bestUsd) {
      bestUsd = usd
      bestIndex = i
    }
  }
  await planButtons.nth(bestIndex).click()

  const methodLabel = opts.method === 'alipay' ? '支付宝' : '微信'
  await page.locator('#recharge-plans').locator('button', { hasText: methodLabel }).click()

  await page.locator('#recharge-plans').locator('button', { hasText: '立即充值' }).click()

  const payDialog = page.locator('div[role="dialog"]').filter({ hasText: '请完成支付' })
  await expect(payDialog).toBeVisible({ timeout: 30_000 })
  await expect(payDialog.locator('text=订单号')).toBeVisible()

  const orderNoText = (await payDialog.locator('p.font-mono').first().textContent()) || ''
  const orderNo = orderNoText.trim()
  if (!orderNo) throw new Error('Failed to read order number from pay modal.')

  const noQRCode = payDialog.locator('text=暂无法生成支付二维码')
  const qrImage = payDialog.locator('img[alt="qr"]')

  if (opts.allowNoQRCode) {
    await expect(noQRCode.or(qrImage)).toBeVisible({ timeout: 30_000 })
  } else {
    await expect(qrImage).toBeVisible({ timeout: 30_000 })
  }

  await payDialog.locator('button', { hasText: '关闭' }).click()
  await expect(payDialog).toHaveCount(0, { timeout: 30_000 })

  return { orderNo }
}
