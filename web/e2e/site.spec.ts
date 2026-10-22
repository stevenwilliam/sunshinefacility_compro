import { expect, test } from '@playwright/test'

const PAGES = ['/', '/services', '/services/food', '/services/remote', '/industries', '/about', '/careers', '/contact']

test('every page renders with exactly one h1 and no console errors', async ({ page }) => {
  const errors: string[] = []
  page.on('pageerror', (e) => errors.push(e.message))
  page.on('console', (m) => { if (m.type() === 'error' && !/WebGL|THREE/.test(m.text())) errors.push(m.text()) })
  for (const p of PAGES) {
    await page.goto(p)
    await expect(page.locator('h1')).toHaveCount(1)
    await expect(page.locator('footer')).toBeVisible()
  }
  expect(errors).toEqual([])
})

test('360 px: no page scrolls sideways', async ({ page }) => {
  await page.setViewportSize({ width: 360, height: 800 })
  for (const p of PAGES) {
    await page.goto(p)
    const over = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth)
    expect(over, `horizontal overflow on ${p}`).toBeLessThanOrEqual(0)
  }
})

test('language switch applies site-wide, updates <html lang> and persists', async ({ page }) => {
  await page.goto('/')
  await page.getByRole('button', { name: 'ID', exact: true }).click()
  await expect(page.locator('html')).toHaveAttribute('lang', 'id')
  await expect(page.getByRole('link', { name: 'Minta Penawaran' }).first()).toBeVisible()
  await page.getByRole('link', { name: 'Layanan', exact: true }).first().click()
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('Lima lini layanan. Satu mitra di lokasi.')
  await page.reload()
  await expect(page.locator('html')).toHaveAttribute('lang', 'id')
  await page.getByRole('button', { name: 'EN', exact: true }).click()
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('Five business lines. One partner on site.')
})

test('?lang=id link opens in Indonesian', async ({ page }) => {
  await page.goto('/about?lang=id')
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('Lebih dari satu dekade keunggulan operasional.')
})

test('certificate lightbox opens and closes with Esc and the button', async ({ page }) => {
  await page.goto('/')
  await page.getByRole('button', { name: 'ISO 22000:2018 certificate' }).click()
  const dlg = page.getByRole('dialog', { name: 'ISO 22000:2018' })
  await expect(dlg).toBeVisible()
  await page.keyboard.press('Escape')
  await expect(dlg).toBeHidden()
  await page.getByRole('button', { name: 'SLHS certificate' }).click()
  await page.getByRole('button', { name: 'Close' }).click()
  await expect(page.getByRole('dialog')).toBeHidden()
  // The expired halal certificate is not on the site.
  await expect(page.getByText('Halal', { exact: true })).toHaveCount(0)
})

test('placeholders are visibly tagged', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByText('Placeholder · replace with real quotes')).toBeVisible()
  await expect(page.locator('footer .ph-tag').first()).toBeVisible()
  await page.goto('/about')
  await expect(page.getByText('Draft · confirm wording')).toBeVisible()
  await expect(page.getByText('Placeholder · replace with real names and photos')).toBeVisible()
  await page.goto('/careers')
  await expect(page.getByText('Sample openings')).toBeVisible()
})

test('contact form: disabled with the reason, then sends', async ({ page }) => {
  await page.goto('/contact?service=remote')
  const send = page.getByRole('button', { name: 'Send request' })
  await expect(send).toBeDisabled()
  await expect(page.locator('#send-why')).toContainText('To send, add: full name')
  await expect(page.getByLabel('Remote')).toBeChecked() // ?service= preselects
  await page.getByLabel('Full name *').fill('E2E Tester')
  await page.getByLabel('Company *').fill('PT E2E')
  await page.getByLabel('Work email *').fill(`e2e+${Date.now()}@example.com`)
  await page.getByLabel('Phone *').fill('+62 812 0000 0000')
  await page.getByLabel('Building type *').selectOption('office')
  await page.getByLabel('Area size').fill('12a34')
  await expect(page.getByLabel('Area size')).toHaveValue('1234') // digits only
  await page.getByLabel('Location *').fill('Jakarta')
  await expect(page.locator('#send-why')).toHaveText('Ready to send.')
  await page.waitForTimeout(3200) // the minimum fill time (BR-5.4)
  await send.click()
  await expect(page.getByRole('status')).toContainText('Thank you, E2E.')
})

test('home quick form carries its fields to the Contact form', async ({ page }) => {
  await page.goto('/')
  const band = page.locator('#quote')
  await band.getByLabel('Full name').fill('Rina')
  await band.getByLabel('Company').fill('PT Rina')
  await band.getByLabel('Building type').selectOption('hospital')
  await band.getByLabel('Food').check()
  await band.getByRole('button', { name: 'Request a Quotation' }).click()
  await expect(page).toHaveURL(/\/contact$/)
  await expect(page.getByLabel('Full name *')).toHaveValue('Rina')
  await expect(page.getByLabel('Building type *')).toHaveValue('hospital')
  await expect(page.getByLabel('Food Services')).toBeChecked()
})

test('careers: filter, clear (disabled with reason), apply prefills the form', async ({ page }) => {
  await page.goto('/careers')
  await expect(page.getByText('Showing 8 of 8 positions')).toBeVisible()
  await expect(page.getByRole('button', { name: 'Clear filters' })).toBeDisabled()
  await expect(page.getByText('No filters applied')).toBeVisible()
  await page.getByLabel('Location').selectOption({ label: 'Tangerang' })
  await expect(page.getByText('Showing 3 of 8 positions')).toBeVisible()
  await page.getByRole('searchbox').fill('zzzz')
  await expect(page.getByText('No positions match these filters.')).toBeVisible()
  await page.getByRole('button', { name: 'Clear filters' }).click()
  await page.getByRole('button', { name: /Camp Catering Supervisor/ }).click()
  await page.getByRole('button', { name: 'Apply for this role' }).click()
  await expect(page.getByLabel('Position *')).toHaveValue(/.+/)
  const big = Buffer.alloc(6 * 1024 * 1024, 1)
  await page.getByLabel('CV *').setInputFiles({ name: 'big.pdf', mimeType: 'application/pdf', buffer: big })
  await expect(page.locator('#cv-note')).toContainText('over 5 MB')
  await expect(page.getByRole('button', { name: 'Send application' })).toBeDisabled()
})

test('service detail: line switcher, figures, coverage map on Remote only', async ({ page }) => {
  await page.goto('/services/remote')
  await expect(page.getByRole('heading', { name: 'Where we deliver' })).toBeVisible()
  await expect(page.getByText('Retail distribution points', { exact: true })).toBeVisible()
  await page.getByRole('link', { name: 'Food Services' }).first().click()
  await expect(page).toHaveURL(/\/services\/food$/)
  await expect(page.getByRole('heading', { name: 'Where we deliver' })).toHaveCount(0)
  await expect(page.getByRole('heading', { name: 'Certified food safety' })).toBeVisible()
})
