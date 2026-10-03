import { test, expect } from '@playwright/test'
import path from 'node:path'

test('real CSV workflow, exact boundary, and invalid rows', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByRole('heading', { name: 'Every holding. In perspective.' })).toBeVisible()
  await page.locator('#holdings').setInputFiles(path.resolve('public/samples/valid.csv'))
  await page.locator('#limit').fill('50')
  const rpc = page.waitForRequest(r => r.url().endsWith('/CheckPortfolio'))
  await page.getByRole('button', { name: 'Check portfolio' }).click()
  expect((await rpc).headers()['content-type']).toContain('application/grpc-web')
  await expect(page.getByTestId('summary')).toContainText('ZAR 100.00')
  await expect(page.getByRole('row').filter({ hasText: 'Alpha' })).toContainText('Above limit')
  await expect(page.getByRole('row').filter({ hasText: 'Beta' })).toContainText('Within limit')
  await page.locator('#limit').fill('60')
  await expect(page.getByRole('status')).toContainText('outdated')
  await page.getByRole('button', { name: 'Check portfolio' }).click()
  await expect(page.getByTestId('breach-count')).toHaveText('0')
  await page.locator('#holdings').setInputFiles(path.resolve('public/samples/invalid.csv'))
  await page.getByRole('button', { name: 'Check portfolio' }).click()
  await expect(page.getByRole('alert')).toContainText('Row 3')
  await expect(page.getByRole('alert')).toContainText('duplicate_id')
  await expect(page.getByTestId('summary')).toHaveCount(0)
})

test('mobile layout contains the page and shows a usable review', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await page.goto('/')
  await page.locator('#holdings').setInputFiles(path.resolve('public/samples/valid.csv'))
  await page.getByRole('button', { name: 'Check portfolio' }).click()
  await expect(page.getByTestId('summary')).toBeVisible()
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
})
