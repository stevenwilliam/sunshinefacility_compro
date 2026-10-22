// Screenshots of the running site and of the approved design, for a side-by-
// side comparison. Not a test: a tool for looking.
//
//   xvfb-run … node scripts/shots.mjs <outdir> [base=http://127.0.0.1:8097] [routes=all] [widths=1440,360]
import { chromium } from '@playwright/test'
import { mkdirSync } from 'node:fs'
import { resolve } from 'node:path'

const out = process.argv[2] || 'shots'
const base = process.argv[3] || 'http://127.0.0.1:8097'
const only = (process.argv[4] || 'all').split(',')
const widths = (process.argv[5] || '1440,360').split(',').map(Number)
const design = 'file://' + resolve('../docs/design/handoff/SFG Site (standalone).html')
mkdirSync(out, { recursive: true })

const routes = [
  ['home', '/', '#/'],
  ['services', '/services', '#/services'],
  ['detail-food', '/services/food', '#/service/food'],
  ['detail-remote', '/services/remote', '#/service/remote'],
  ['industries', '/industries', '#/industries'],
  ['about', '/about', '#/about'],
  ['careers', '/careers', '#/careers'],
  ['contact', '/contact', '#/contact'],
].filter((r) => only.includes('all') || only.includes(r[0]))

// Headed Chromium under Xvfb with software GL: on this VM headless Chromium
// cannot create a WebGL context (it picks the VMware GPU and fails), so the
// hero ball would never render. Run as:
//   LIBGL_ALWAYS_SOFTWARE=1 xvfb-run -a -s "-screen 0 1600x1000x24" node scripts/shots.mjs …
const browser = await chromium.launch({ headless: false, channel: 'chromium', args: ['--use-gl=angle', '--use-angle=gl', '--ignore-gpu-blocklist'] })
for (const w of widths) {
  for (const which of ['site', 'design']) {
    const ctx = await browser.newContext({ viewport: { width: w, height: w < 760 ? 800 : 900 }, deviceScaleFactor: 1 })
    const page = await ctx.newPage()
    for (const [name, path, hash] of routes) {
      const url = which === 'site' ? base + path : design + hash
      await page.goto(url, { waitUntil: 'load' })
      await page.waitForTimeout(which === 'design' ? 2500 : 1800)
      await page.screenshot({ path: `${out}/${name}-${w}-${which}.png`, fullPage: true })
      console.log('shot', name, w, which)
    }
    await ctx.close()
  }
}
await browser.close()
