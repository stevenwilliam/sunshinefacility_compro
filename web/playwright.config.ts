import { defineConfig } from '@playwright/test'

// End-to-end tests against the running dev site (nginx :8097 → sfg).
// Run: SFG_E2E_PASSWORD=… npx playwright test   (make e2e does this)
export default defineConfig({
  testDir: './e2e',
  timeout: 30_000,
  retries: 0,
  workers: 1,
  reporter: [['list']],
  use: { baseURL: process.env.SFG_E2E_BASE || 'http://127.0.0.1:8097', trace: 'off' },
})
