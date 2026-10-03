import { defineConfig } from '@playwright/test'
export default defineConfig({
  testDir: './e2e', fullyParallel: false, workers: 1,
  use: { baseURL: 'http://127.0.0.1:8080', channel: 'msedge', headless: true, screenshot: 'only-on-failure' },
  reporter: 'list',
})
