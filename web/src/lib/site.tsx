import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'
import type { Bundle, L2, Lang, Section } from './types'

// The language lives in localStorage['sfg-lang'] and applies site-wide
// without a reload (README → Language). ?lang=id in the URL wins and is
// remembered, so a shared Indonesian link opens in Indonesian (decision Q17).
const KEY = 'sfg-lang'

function initialLang(): Lang {
  try {
    const q = new URLSearchParams(window.location.search).get('lang')
    if (q === 'en' || q === 'id') {
      localStorage.setItem(KEY, q)
      return q
    }
    const s = localStorage.getItem(KEY)
    if (s === 'en' || s === 'id') return s
  } catch {
    /* storage blocked: fall back to English */
  }
  return document.documentElement.lang === 'id' ? 'id' : 'en'
}

interface SiteValue {
  b: Bundle
  lang: Lang
  setLang: (l: Lang) => void
  /** Pick the visitor's language from a [en, id] pair. */
  L: (v: L2 | undefined) => string
  /** A catalogue string, with {name} placeholders filled. */
  t: (key: string, vars?: Record<string, string | number>) => string
  /** Is this catalogue key still a placeholder (BR-2.1)? */
  tph: (key: string) => boolean
  section: (key: string) => Section | undefined
  param: (key: string) => string
  paramPH: (key: string) => boolean
}

const Ctx = createContext<SiteValue | null>(null)

export function SiteProvider({ bundle, children }: { bundle: Bundle; children: ReactNode }) {
  const [lang, setLangState] = useState<Lang>(initialLang)

  useEffect(() => {
    document.documentElement.lang = lang
  }, [lang])

  const setLang = useCallback((l: Lang) => {
    try {
      localStorage.setItem(KEY, l)
    } catch {
      /* ignore */
    }
    setLangState(l)
  }, [])

  const value = useMemo<SiteValue>(() => {
    const i = lang === 'id' ? 1 : 0
    const L = (v: L2 | undefined) => (v ? v[i] || v[0] : '')
    const ph = new Set(bundle.textPlaceholders)
    const t = (key: string, vars?: Record<string, string | number>) => {
      const v = bundle.text[key]
      if (!v) {
        // A missing key must be loud in development and invisible in
        // production (impeccable §2: a missing catalogue key renders as the
        // key). The guard test in test/catalogue_test.go keeps it from
        // shipping at all.
        if (import.meta.env.DEV) console.error(`missing site text: ${key}`)
        return import.meta.env.DEV ? `⟦${key}⟧` : ''
      }
      let s = v[i] || v[0]
      if (vars) for (const [k, x] of Object.entries(vars)) s = s.split(`{${k}}`).join(String(x))
      return s
    }
    return {
      b: bundle,
      lang,
      setLang,
      L,
      t,
      tph: (key) => ph.has(key),
      section: (key) => {
        const s = bundle.sections[key]
        return s && s.published ? s : undefined
      },
      param: (key) => L(bundle.params[key]?.v),
      paramPH: (key) => !!bundle.params[key]?.placeholder,
    }
  }, [bundle, lang, setLang])

  return <Ctx.Provider value={value}>{children}</Ctx.Provider>
}

export function useSite(): SiteValue {
  const v = useContext(Ctx)
  if (!v) throw new Error('useSite outside SiteProvider')
  return v
}

/** The bundle the server embedded in the page, or a fetch in development. */
export async function loadBundle(): Promise<Bundle> {
  const el = document.getElementById('sfg-bundle')
  if (el?.textContent) return JSON.parse(el.textContent) as Bundle
  const r = await fetch('/api/v1/public/bundle')
  if (!r.ok) throw new Error('bundle ' + r.status)
  return (await r.json()) as Bundle
}

/** The line's colours as CSS variables, for classes that use var(--ink). */
export function lineVars(l: { base: string; ink: string; shade: string; onInk: string; labelColor?: string }) {
  return {
    '--base': l.base,
    '--ink': l.ink,
    '--shade': l.shade,
    '--on-ink': l.onInk,
    '--label': l.labelColor || l.ink,
  } as React.CSSProperties
}

export function fmtMonthYear(iso: string, lang: Lang) {
  const d = new Date(iso + 'T00:00:00')
  if (isNaN(d.getTime())) return iso
  return d.toLocaleDateString(lang === 'id' ? 'id-ID' : 'en-GB', { month: 'short', year: 'numeric' })
}
