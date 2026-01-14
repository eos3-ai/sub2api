import { test, expect } from '@playwright/test'
import { loadEnvConfig } from '../src/config'
import { login } from '../src/flows'

test('API 密钥：创建 + 删除', async ({ page }) => {
  const env = loadEnvConfig()
  await login(page, env.accounts.user)

  await page.goto('/keys')
  await expect(page.locator('text=API 密钥')).toBeVisible({ timeout: 30_000 })

  const keyName = `e2e-key-${Date.now()}`

  await page.locator('[data-tour="keys-create-btn"]').click()

  await page.locator('[data-tour="key-form-name"]').fill(keyName)

  // Select group (required).
  await page.locator('[data-tour="key-form-group"] button[aria-label="Select option"]').click()
  await page.locator('div[role="listbox"] div[role="option"]').first().click()

  await page.locator('[data-tour="key-form-submit"]').click()

  const row = page.locator('tr', { hasText: keyName })
  await expect(row).toBeVisible({ timeout: 30_000 })

  // Delete the key we just created.
  await row.locator('button', { hasText: '删除' }).click()
  const confirmDialog = page.locator('div[role="dialog"]').filter({ hasText: '删除密钥' })
  await expect(confirmDialog).toBeVisible({ timeout: 30_000 })
  await confirmDialog.locator('button', { hasText: '删除' }).click()

  await expect(page.locator('tr', { hasText: keyName })).toHaveCount(0, { timeout: 30_000 })
})
