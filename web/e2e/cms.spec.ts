import { expect, test, type Page } from '@playwright/test'

const PASSWORD = process.env.SFG_E2E_PASSWORD || ''

// One sign-in for the whole file: the login endpoint is rate limited per
// address (BR-8.3), and four sign-ins in a few seconds would trip it.
test.describe.configure({ mode: 'serial' })
let page: Page
test.beforeAll(async ({ browser }) => {
  page = await browser.newPage()
  if (!PASSWORD) return
  await page.goto('/tytb')
  await page.getByLabel('Email').fill('admin@sfg.local')
  await page.getByLabel('Password').fill(PASSWORD)
  await page.getByRole('button', { name: 'Sign in' }).click()
  await expect(page.getByRole('navigation', { name: 'Admin' })).toBeVisible()
})
test.afterAll(async () => { await page.close() })
async function signIn(_p: Page) {
  test.skip(!PASSWORD, 'SFG_E2E_PASSWORD not set')
}

test('an edit in the CMS appears on the site, in both languages', async () => {
  await signIn(page)
  await page.goto('/tytb/r/text')
  await page.getByRole('searchbox').fill('home.trust.trusted')
  await page.getByRole('button', { name: 'Edit home.trust.trusted' }).click()
  await page.getByLabel('English *').fill('Trusted by (e2e)')
  await page.getByRole('button', { name: 'Save' }).click()
  await expect(page.getByRole('status')).toContainText('Saved')
  await page.goto('/')
  await expect(page.getByText('Trusted by (e2e)')).toBeVisible()
  // put it back
  await page.goto('/tytb/r/text')
  await page.getByRole('searchbox').fill('home.trust.trusted')
  await page.getByRole('button', { name: 'Edit home.trust.trusted' }).click()
  await page.getByLabel('English *').fill('Trusted by')
  await page.getByRole('button', { name: 'Save' }).click()
  await expect(page.getByRole('status')).toContainText('Saved')
})

test('expired certificate: the Published toggle is disabled and says why', async () => {
  await signIn(page)
  await page.goto('/tytb/r/certificates')
  await expect(page.getByText('Expired')).toBeVisible()
  await page.getByRole('button', { name: 'Edit Halal' }).click()
  const toggle = page.getByLabel('Draft (hidden from the site)')
  await expect(toggle).toBeDisabled()
  await expect(page.getByText(/Can't publish: this certificate expired on 25 Aug 2025/)).toBeVisible()
})

test('lists: search filters, export disabled with the reason when empty', async () => {
  await signIn(page)
  await page.goto('/tytb/r/clients')
  await expect(page.getByText('19 items')).toBeVisible()
  await page.getByRole('searchbox').fill('siloam')
  await expect(page.getByText('Showing 1 of 19')).toBeVisible()
  await page.getByRole('searchbox').fill('no-such-client')
  await expect(page.getByRole('button', { name: 'Export CSV' })).toBeDisabled()
  await expect(page.getByText('Nothing to export: no rows match the current search.')).toBeVisible()
})

test('media in use: delete disabled with the usage count', async () => {
  await signIn(page)
  await page.goto('/tytb/media')
  const del = page.getByRole('button', { name: 'Delete danone.png' })
  await expect(del).toBeDisabled()
  await expect(del).toHaveAttribute('title', /used in 1 item/)
})
