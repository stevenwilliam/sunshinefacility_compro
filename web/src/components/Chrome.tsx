import { useEffect, useState, type CSSProperties } from 'react'
import { Link, NavLink, useLocation } from 'react-router-dom'
import { FacebookLogo, InstagramLogo, LinkedinLogo, List, WhatsappLogo, X, YoutubeLogo } from '@/lib/phosphor'
import { lineVars, useSite } from '@/lib/site'
import { cssUrl, PH } from './bits'

const NAV: [string, string][] = [
  ['/', 'nav.home'], ['/services', 'nav.services'], ['/industries', 'nav.industries'],
  ['/about', 'nav.about'], ['/careers', 'nav.careers'], ['/contact', 'nav.contact'],
]

export function waHref(digits: string) {
  return digits ? `https://wa.me/${digits}` : '#'
}

export function telHref(display: string) {
  const d = display.replace(/[^0-9+]/g, '')
  return d ? `tel:${d}` : '#'
}

function LangSwitch() {
  const { lang, setLang, t } = useSite()
  const btn = (l: 'en' | 'id') => (
    <button type="button" onClick={() => setLang(l)} aria-pressed={lang === l}
      className={`lang-btn relative rounded-full border-0 px-[10px] py-[6px] cursor-pointer after:absolute after:-inset-x-0.5 after:-inset-y-1.5 after:content-[''] ${lang === l ? 'bg-charcoal text-white' : 'bg-transparent text-ink-2 hover:text-charcoal'}`}>
      {l.toUpperCase()}
    </button>
  )
  return (
    <div role="group" aria-label={t('header.lang_aria')} className="flex rounded-full border border-line-3 p-[3px] text-[13px] font-medium">
      {btn('en')}
      {btn('id')}
    </div>
  )
}

export function SiteHeader() {
  const { t, b } = useSite()
  const [open, setOpen] = useState(false)
  const loc = useLocation()
  useEffect(() => setOpen(false), [loc.pathname])
  useEffect(() => {
    if (!open) return
    const k = (e: KeyboardEvent) => e.key === 'Escape' && setOpen(false)
    window.addEventListener('keydown', k)
    return () => window.removeEventListener('keydown', k)
  }, [open])
  const logo = b.logos['sfg-logo-primary']
  const isActive = (to: string) => (to === '/' ? loc.pathname === '/' : loc.pathname === to || loc.pathname.startsWith(to + '/'))
  return (
    <header className="sticky top-0 z-50 border-b border-line-2 bg-white/[.96] text-charcoal backdrop-blur-[10px]">
      <a href="#main" className="sr-only-focusable absolute left-4 top-3 z-[60] rounded bg-charcoal px-4 py-2 text-white">{t('common.skip')}</a>
      <div className="wrap flex h-[72px] items-center gap-8">
        <Link to="/" aria-label={t('header.home_aria')} className="flex flex-none items-center">
          <span role="img" aria-label={t('common.brand')} className="block h-[34px] w-[138px] bg-contain bg-left bg-no-repeat" style={{ backgroundImage: cssUrl(logo) }} />
        </Link>
        <nav aria-label={t('nav.main_aria')} className="ml-2 hidden gap-[26px] text-[15px] nav:flex">
          {NAV.map(([to, key]) => {
            const on = isActive(to)
            return (
              <NavLink key={to} to={to} aria-current={on ? 'page' : undefined}
                className={`border-b-2 pb-[22px] pt-6 hover:text-charcoal ${on ? 'border-red font-[550] text-charcoal' : 'border-transparent font-[450] text-ink-2'}`}>
                {t(key)}
              </NavLink>
            )
          })}
        </nav>
        <div className="ml-auto flex items-center gap-[14px]">
          <LangSwitch />
          <Link to="/contact" className="btn btn-red hidden px-[18px] py-3 text-[15px] sm:inline-flex">{t('cta.quote')}</Link>
          <button type="button" onClick={() => setOpen((o) => !o)} aria-label={t('header.menu')} aria-expanded={open} aria-controls="mobile-nav"
            className="grid h-11 w-11 cursor-pointer place-items-center rounded border border-line-3 bg-white text-[22px] text-charcoal nav:hidden">
            {open ? <X aria-hidden="true" /> : <List aria-hidden="true" />}
          </button>
        </div>
      </div>
      {open && (
        <nav id="mobile-nav" aria-label={t('nav.mobile_aria')} className="flex flex-col border-t border-line-2 bg-white pb-5 pt-2 nav:hidden" style={{ paddingLeft: 'clamp(20px,4vw,48px)', paddingRight: 'clamp(20px,4vw,48px)' }}>
          {NAV.map(([to, key]) => {
            const on = isActive(to)
            return (
              <NavLink key={to} to={to} aria-current={on ? 'page' : undefined}
                className={`border-b border-line-1 py-[14px] text-[20px] ${on ? 'font-[550] text-charcoal' : 'font-[450] text-ink-2'}`}>
                {t(key)}
              </NavLink>
            )
          })}
          <Link to="/contact" className="btn btn-red mt-4 p-[15px]">{t('cta.quote')}</Link>
        </nav>
      )}
    </header>
  )
}

export function DividerStrip() {
  const { b } = useSite()
  return (
    <div aria-hidden="true" className="divider" style={{ '--n': b.lines.length || 5 } as CSSProperties}>
      {b.lines.map((l) => (
        <div key={l.key} style={lineVars(l)}>
          <div style={{ WebkitMaskImage: cssUrl(l.mark), maskImage: cssUrl(l.mark) }} />
        </div>
      ))}
    </div>
  )
}

const SOCIAL = [
  ['social.instagram', 'Instagram', InstagramLogo], ['social.linkedin', 'LinkedIn', LinkedinLogo],
  ['social.facebook', 'Facebook', FacebookLogo], ['social.youtube', 'YouTube', YoutubeLogo],
] as const

export function SiteFooter() {
  const { t, b, param, paramPH } = useSite()
  const phone = param('contact.phone')
  const wa = param('contact.whatsapp')
  const email = param('contact.email')
  const socialPH = SOCIAL.some(([k]) => paramPH(k))
  const link = 'text-ondark-muted hover:text-tangerine'
  return (
    <>
      <DividerStrip />
      <footer className="on-dark bg-charcoal-deep text-[15px] text-ondark-muted">
        <div className="wrap grid gap-10 pb-7 pt-14" style={{ gridTemplateColumns: 'repeat(auto-fit,minmax(200px,1fr))' }}>
          <div className="flex flex-col gap-4">
            <span role="img" aria-label={t('common.brand')} className="block h-9 w-[146px] bg-contain bg-left bg-no-repeat" style={{ backgroundImage: cssUrl(b.logos['sfg-logo-white']) }} />
            <p className="m-0 leading-[1.55]">{param('contact.address')} <PH show={paramPH('contact.address')} small /></p>
            <p className="m-0 text-[13px]">{t('footer.lippo')}</p>
          </div>
          <div>
            <p className="mb-[14px] mt-0 font-medium text-white">{t('footer.contact')}</p>
            <ul className="m-0 flex list-none flex-col gap-[10px] p-0">
              <li><a href={telHref(phone)} className={link}>{phone}</a> <PH show={paramPH('contact.phone')} small /></li>
              <li><a href={waHref(param('contact.whatsapp_digits'))} className={link}>{t('common.whatsapp')} {wa}</a> <PH show={paramPH('contact.whatsapp')} small /></li>
              <li><a href={`mailto:${email}`} className={link}>{email}</a> <PH show={paramPH('contact.email')} small /></li>
            </ul>
          </div>
          <div>
            <p className="mb-[14px] mt-0 font-medium text-white">{t('footer.links')}</p>
            <ul className="m-0 flex list-none flex-col gap-[10px] p-0">
              {NAV.map(([to, key]) => <li key={to}><Link to={to} className={link}>{t(key)}</Link></li>)}
            </ul>
          </div>
          <div>
            <p className="mb-[14px] mt-0 font-medium text-white">{t('footer.follow')}</p>
            <div className="flex flex-wrap items-center gap-2">
              {SOCIAL.map(([k, label, I]) => (
                <a key={k} href={param(k) || '#'} aria-label={label} rel="noopener noreferrer" target="_blank"
                  className="grid h-11 w-11 place-items-center rounded border border-charcoal-500 text-[20px] text-ondark-muted hover:border-tangerine hover:text-tangerine">
                  <I aria-hidden="true" />
                </a>
              ))}
              <PH show={socialPH} small />
            </div>
          </div>
        </div>
        <div className="wrap flex flex-wrap justify-between gap-x-6 gap-y-2 border-t border-charcoal-700 pb-8 pt-5 text-[13px]">
          <span>{t('footer.copyright')}</span>
          <span>{t('footer.domain')}</span>
        </div>
      </footer>
      <a href={waHref(param('contact.whatsapp_digits'))} aria-label={t('wa.chat')} target="_blank" rel="noopener noreferrer"
        className="fixed z-[60] grid h-[60px] w-[60px] place-items-center rounded-full bg-wa text-[32px] text-white hover:bg-wa-hover hover:text-white"
        style={{ right: 'clamp(16px,2vw,28px)', bottom: 'clamp(16px,2vw,28px)', boxShadow: '0 10px 24px -6px rgba(22,24,19,0.45)' }}>
        <WhatsappLogo weight="fill" aria-hidden="true" />
      </a>
    </>
  )
}
