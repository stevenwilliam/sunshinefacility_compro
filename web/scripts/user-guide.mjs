// The CMS user guide: docs/user_guide/user-guide.md (source), img/ (real
// screenshots of the running CMS) and user-guide.pdf (generated).
//
//   node scripts/user-guide.mjs shots   capture img/ from the running dev CMS (needs SFG_GUIDE_PASSWORD)
//   node scripts/user-guide.mjs pdf     render the Markdown to user-guide.pdf and write the stamp
//   node scripts/user-guide.mjs check   fail if the guide is older than the code it describes
//
// The stamp is a hash of the guide and of everything a CMS user sees: the
// admin UI, the resource definitions, roles and parameters. Change any of
// them and `make check` fails until `make user-guide` has been run again —
// which is how the guide stays current (CLAUDE.md §7).
import { chromium } from '@playwright/test'
import { createHash } from 'node:crypto'
import { existsSync, mkdirSync, readdirSync, readFileSync, statSync, writeFileSync } from 'node:fs'
import { join, relative, resolve } from 'node:path'
import { marked } from 'marked'

const root = resolve(import.meta.dirname, '../..')
const dir = join(root, 'docs/user_guide')
const img = join(dir, 'img')
const md = join(dir, 'user-guide.md')
const pdf = join(dir, 'user-guide.pdf')
const stampFile = join(dir, '.stamp')

// What the guide describes. A change to any of these can change what a user sees.
const watched = [
  'docs/user_guide/user-guide.md',
  'docs/user_guide/img',
  'web/src/admin',
  'web/scripts/user-guide.mjs',
  'internal/domain/resource',
  'internal/domain/access',
  'internal/domain/params',
]

function files(p) {
  const abs = join(root, p)
  if (!existsSync(abs)) return []
  if (statSync(abs).isFile()) return [p]
  return readdirSync(abs).sort().flatMap((n) => files(join(p, n)))
}

function stamp() {
  const h = createHash('sha256')
  for (const f of watched.flatMap(files).filter((f) => !f.endsWith('_test.go'))) {
    h.update(f + '\0')
    h.update(readFileSync(join(root, f)))
    h.update('\0')
  }
  return h.digest('hex')
}

async function shots() {
  const base = process.env.SFG_GUIDE_BASE || 'http://127.0.0.1:8097'
  const email = process.env.SFG_GUIDE_EMAIL || 'admin@sfg.local'
  const password = process.env.SFG_GUIDE_PASSWORD
  if (!password) throw new Error('SFG_GUIDE_PASSWORD is not set')
  mkdirSync(img, { recursive: true })

  const browser = await chromium.launch()
  const ctx = await browser.newContext({ viewport: { width: 1366, height: 820 }, deviceScaleFactor: 1, colorScheme: 'light' })
  const page = await ctx.newPage()
  const shot = async (name) => {
    await page.waitForTimeout(900)
    await page.screenshot({ path: join(img, name + '.png') })
    console.log('shot', name)
  }
  // Move around by the sidebar, as a person does: a full page load refreshes
  // the session, and the refresh endpoint is rate limited.
  const go = async (label) => {
    await page.getByRole('navigation', { name: 'Admin' }).getByRole('link', { name: new RegExp('^' + label) }).click()
    await page.waitForTimeout(800)
  }
  const drawer = async (label) => { await page.getByRole('button', { name: label, exact: true }).first().click(); await page.waitForTimeout(500) }
  const close = async () => { await page.keyboard.press('Escape'); await page.waitForTimeout(300) }

  await page.goto(base + '/tytb', { waitUntil: 'load' })
  await shot('01-sign-in')
  await page.getByLabel('Email').fill(email)
  await page.getByLabel('Password').fill(password)
  await page.getByRole('button', { name: 'Sign in' }).click()
  await page.getByRole('navigation', { name: 'Admin' }).waitFor()
  await shot('02-after-sign-in')

  await go('Clients')
  await shot('03-list')
  await page.getByRole('searchbox').fill('no-such-client')
  await shot('04-nothing-to-export')

  await go('Site text')
  await page.getByRole('searchbox').fill('home.trust.trusted')
  await drawer('Edit home.trust.trusted')
  await shot('05-editor')
  await close()

  await go('Pages')
  await shot('06-pages')
  await go('Services')
  await shot('07-services')

  await go('Certificates')
  await drawer('Edit Halal')
  await page.getByText(/Can't publish/).scrollIntoViewIfNeeded()
  await shot('08-expired-certificate')
  await close()

  await go('Media library')
  await shot('09-media')
  await go('Parameters')
  await shot('10-parameters')

  await go('Quotation requests')
  await shot('11-quotations')
  await page.locator('tr.click').first().click()
  await shot('11b-quotation-detail')
  await close()
  await go('Job applications')
  await shot('12-applications')
  await go('Users')
  await shot('13-users')
  await go('Audit history')
  await shot('14-audit')
  await go('Outgoing mail')
  await shot('15-outgoing-mail')
  await go('Your account')
  await shot('16-account')

  await browser.close()
}

const css = `
@page { size: A4; margin: 18mm 16mm 20mm; }
* { box-sizing: border-box; }
body { font: 10.5pt/1.5 "Inter", "Segoe UI", "Helvetica Neue", Arial, sans-serif; color: #1d1d1b; margin: 0; }
h1 { font-size: 26pt; line-height: 1.15; margin: 0 0 6pt; letter-spacing: -0.01em; }
h2 { font-size: 16pt; margin: 22pt 0 6pt; padding-top: 6pt; border-top: 2px solid #f2b705; break-after: avoid; }
h3 { font-size: 12pt; margin: 14pt 0 4pt; break-after: avoid; }
h2.new-page { break-before: page; }
p, li { orphans: 3; widows: 3; }
table { border-collapse: collapse; width: 100%; margin: 6pt 0 10pt; font-size: 9.5pt; break-inside: auto; }
th, td { border: 1px solid #d9d6cf; padding: 4pt 6pt; text-align: left; vertical-align: top; }
th { background: #f4f2ed; }
tr { break-inside: avoid; }
code { font: 9pt/1.3 "JetBrains Mono", Consolas, monospace; background: #f4f2ed; padding: 0 3pt; border-radius: 3px; }
img { display: block; max-width: 100%; border: 1px solid #d9d6cf; border-radius: 4px; margin: 6pt 0 2pt; break-inside: avoid; }
p:has(> img) { break-inside: avoid; }
em.caption, p > img + em { display: block; font-size: 8.5pt; color: #6b6860; margin-bottom: 8pt; }
blockquote { margin: 8pt 0; padding: 6pt 10pt; border-left: 3px solid #f2b705; background: #fbf8ef; }
.cover { height: 240mm; display: flex; flex-direction: column; justify-content: center; break-after: page; }
.cover .kicker { font-size: 10pt; text-transform: uppercase; letter-spacing: 0.12em; color: #8a6a00; }
.cover .meta { margin-top: 18pt; color: #6b6860; }
nav.toc ol { padding-left: 16pt; } nav.toc li { margin: 2pt 0; }
`

async function render() {
  const src = readFileSync(md, 'utf8')
  // Front matter-free Markdown: the first H1 and the paragraph under it make the cover.
  const [, title = 'User guide', rest = src] = src.match(/^# (.+)\n([\s\S]*)$/) || []
  const tokens = marked.lexer(rest)
  const toc = tokens.filter((t) => t.type === 'heading' && t.depth === 2).map((t) => t.text)
  // Images are inlined: a page made with setContent may not load file:// URLs.
  let body = marked.parser(tokens).replace(/<img src="([^"]+)"/g, (_, src) =>
    `<img src="data:image/png;base64,${readFileSync(join(dir, src)).toString('base64')}"`)
  let n = 0
  body = body.replace(/<h2>/g, () => (n++ === 0 ? '<h2>' : '<h2 class="new-page">'))
  const updated = new Date().toISOString().slice(0, 10)
  const html = `<!doctype html><html lang="en"><head><meta charset="utf-8"><title>${title}</title>
<style>${css}</style></head><body>
<section class="cover"><div class="kicker">Sunshine Facilities Group · Website CMS</div><h1>${title}</h1>
<div class="meta">Generated ${updated} from docs/user_guide/user-guide.md</div>
<nav class="toc"><h3>Contents</h3><ol>${toc.map((t) => `<li>${t}</li>`).join('')}</ol></nav></section>
${body}</body></html>`

  const browser = await chromium.launch()
  const page = await browser.newPage()
  await page.setContent(html, { waitUntil: 'load' })
  await page.pdf({
    path: pdf, format: 'A4', printBackground: true, preferCSSPageSize: true,
    displayHeaderFooter: true, headerTemplate: '<span></span>',
    footerTemplate: `<div style="font:8px Arial,sans-serif;color:#888;width:100%;padding:0 16mm;display:flex;justify-content:space-between"><span>${title}</span><span><span class="pageNumber"></span> / <span class="totalPages"></span></span></div>`,
  })
  await browser.close()
  writeFileSync(stampFile, stamp() + '\n')
  console.log('wrote', relative(root, pdf))
}

function check() {
  const want = stamp()
  const have = existsSync(stampFile) ? readFileSync(stampFile, 'utf8').trim() : ''
  if (!existsSync(pdf) || have !== want) {
    console.error('user guide is out of date: the CMS or the guide changed since the PDF was built.\n' +
      'Update docs/user_guide/user-guide.md if behaviour changed, then run `make user-guide`.')
    process.exit(1)
  }
  console.log('user guide: current')
}

const cmd = process.argv[2]
if (cmd === 'shots') await shots()
else if (cmd === 'pdf') await render()
else if (cmd === 'check') check()
else { console.error('usage: user-guide.mjs shots|pdf|check'); process.exit(2) }
