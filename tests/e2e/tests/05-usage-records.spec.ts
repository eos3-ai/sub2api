import { test, expect } from '@playwright/test'
import { loadEnvConfig } from '../src/config'
import { login } from '../src/flows'

test('普通用户：使用记录页可正常访问', async ({ page }) => {
  const env = loadEnvConfig()
  await login(page, env.accounts.user)

  await page.goto('/usage')

  await expect(page.locator('text=使用记录')).toBeVisible({ timeout: 30_000 })
  await expect(page.locator('button', { hasText: '导出 CSV' })).toBeVisible({ timeout: 30_000 })

  // Table should render either rows or empty state.
  const hasEmpty = page.locator('text=暂无数据')
  const table = page.locator('table')
  await expect(hasEmpty.or(table)).toBeVisible({ timeout: 30_000 })
})
