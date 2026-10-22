import { useEffect, useState } from 'react'
import { useLocation } from 'react-router-dom'
import { useSite } from '@/lib/site'

// GA4 (decision Q22, BR-12.1). Nothing loads unless BOTH are true: a
// measurement ID is set in Parameters, and this visitor pressed Accept.
// Consent mode starts denied; the gtag bootstrap is here in the bundle, not
// an inline script, so the CSP needs no 'unsafe-inline'. Admin never loads
// this component.
const KEY = 'sfg-consent'

declare global {
  interface Window {
    dataLayer?: unknown[]
    gtag?: (...args: unknown[]) => void
  }
}

function loadGA(id: string) {
  if (window.gtag) return
  window.dataLayer = window.dataLayer || []
  window.gtag = function gtag() {
    // eslint-disable-next-line prefer-rest-params
    window.dataLayer!.push(arguments)
  }
  window.gtag('consent', 'default', { ad_storage: 'denied', ad_user_data: 'denied', ad_personalization: 'denied', analytics_storage: 'granted' })
  window.gtag('js', new Date())
  window.gtag('config', id, { anonymize_ip: true, send_page_view: false })
  const s = document.createElement('script')
  s.async = true
  s.src = `https://www.googletagmanager.com/gtag/js?id=${encodeURIComponent(id)}`
  document.head.appendChild(s)
}

export function Consent() {
  const { b, t } = useSite()
  const loc = useLocation()
  const id = b.ga4
  const [choice, setChoice] = useState<string | null>(() => {
    try {
      return localStorage.getItem(KEY)
    } catch {
      return null
    }
  })

  useEffect(() => {
    if (id && choice === 'granted') loadGA(id)
  }, [id, choice])

  useEffect(() => {
    if (id && choice === 'granted' && window.gtag) window.gtag('event', 'page_view', { page_path: loc.pathname })
  }, [id, choice, loc.pathname])

  if (!id || choice) return null
  const decide = (v: 'granted' | 'denied') => {
    try {
      localStorage.setItem(KEY, v)
    } catch {
      /* ignore */
    }
    setChoice(v)
  }
  return (
    <div role="dialog" aria-label={t('consent.aria')} className="on-dark fixed bottom-4 left-4 right-24 z-[70] mx-auto flex max-w-[640px] flex-wrap items-center gap-3 rounded bg-charcoal p-4 text-[14px] text-ondark-body shadow-modal">
      <p className="m-0 flex-[1_1_260px] leading-[1.5]">{t('consent.text')}</p>
      <div className="flex gap-2">
        <button type="button" className="btn btn-red px-4 py-2 text-[14px]" onClick={() => decide('granted')}>{t('consent.accept')}</button>
        <button type="button" className="btn btn-line-light px-4 py-2 text-[14px]" onClick={() => decide('denied')}>{t('consent.decline')}</button>
      </div>
    </div>
  )
}
