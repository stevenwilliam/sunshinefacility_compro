import { useEffect, useRef, useState } from 'react'
import { MagnifyingGlassPlus, SealCheck, X } from '@/lib/phosphor'
import type { Certificate } from '@/lib/types'
import { fmtMonthYear, useSite } from '@/lib/site'
import { cssUrl } from './bits'

export function useCertMeta() {
  const { t, lang } = useSite()
  return (c: Certificate) =>
    c.validUntil ? t('certs.valid_until', { date: fmtMonthYear(c.validUntil, lang) }) : c.issuer ? t('certs.issued_by', { issuer: c.issuer }) : ''
}

/** The full-size certificate. Closed by the backdrop, the Close button or Esc. */
function Lightbox({ cert, onClose }: { cert: Certificate; onClose: () => void }) {
  const { t } = useSite()
  const meta = useCertMeta()
  const closeRef = useRef<HTMLButtonElement>(null)
  useEffect(() => {
    const prev = document.activeElement as HTMLElement | null
    closeRef.current?.focus()
    const k = (e: KeyboardEvent) => e.key === 'Escape' && onClose()
    window.addEventListener('keydown', k)
    const overflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    return () => {
      window.removeEventListener('keydown', k)
      document.body.style.overflow = overflow
      prev?.focus()
    }
  }, [onClose])
  return (
    <div role="dialog" aria-modal="true" aria-label={cert.name} onClick={onClose}
      className="on-dark fixed inset-0 z-[80] box-border flex items-center justify-center bg-[rgba(22,24,19,0.88)]" style={{ padding: 'clamp(16px,4vw,48px)' }}>
      <figure className="m-0 flex max-h-full flex-col items-center gap-3" onClick={(e) => e.stopPropagation()}>
        <img src={cert.full || cert.image?.src} alt={t('certs.aria', { name: cert.name })}
          className="block max-w-full bg-white object-contain shadow-modal" style={{ height: 'calc(100vh - 140px)', aspectRatio: '1/1.414' }} />
        <figcaption className="flex flex-wrap items-center justify-center gap-4 text-[15px] text-ondark-bright">
          {cert.name} · {meta(cert)}
          <button ref={closeRef} type="button" onClick={onClose}
            className="flex min-h-[44px] cursor-pointer items-center gap-2 rounded-lg border border-charcoal-400 bg-transparent px-4 text-[14px] text-white hover:border-white">
            <X aria-hidden="true" />{t('common.close')}
          </button>
        </figcaption>
      </figure>
    </div>
  )
}

/** Home: clickable cards that open the lightbox. */
export function CertificateCards() {
  const { b, L, t } = useSite()
  const meta = useCertMeta()
  const [open, setOpen] = useState<number>(-1)
  return (
    <>
      <ul className="m-0 grid list-none gap-4 p-0" style={{ gridTemplateColumns: 'repeat(auto-fill,minmax(260px,1fr))' }}>
        {b.certificates.map((c, i) => (
          <li key={c.name} className="flex">
            <button type="button" onClick={() => setOpen(i)} aria-label={t('certs.aria', { name: c.name })}
              className="flex flex-1 cursor-zoom-in flex-col overflow-hidden rounded-xl border border-line-2 bg-white p-0 text-left text-charcoal transition-[border-color,transform] duration-200 hover:-translate-y-0.5 hover:border-charcoal">
              <div className="box-border flex h-[260px] w-full justify-center overflow-hidden bg-cotton px-5 pt-5">
                <div aria-hidden="true" className="w-[78%] self-start shadow-card" style={{ aspectRatio: '1/1.414', background: `#FFFFFF ${cssUrl(c.image)} top center/cover no-repeat` }} />
              </div>
              <div className="flex w-full flex-col gap-1.5 border-t border-line-2 px-5 pb-5 pt-[18px]">
                <div className="flex items-center justify-between gap-3">
                  <h3 className="m-0 text-[19px] font-[550] tracking-[-0.01em]">{c.name}</h3>
                  <MagnifyingGlassPlus aria-hidden="true" className="text-[20px] text-ink-3" />
                </div>
                <p className="m-0 text-[14px] leading-[1.5] text-ink-2">{L(c.description)}</p>
                <p className="m-0 mt-1.5 text-[13px] text-ink-3 tnum">{meta(c)}</p>
              </div>
            </button>
          </li>
        ))}
      </ul>
      {open >= 0 && b.certificates[open] && <Lightbox cert={b.certificates[open]} onClose={() => setOpen(-1)} />}
    </>
  )
}

/** About: the compact cards (220px image area, 72% document). */
export function CertificateTiles() {
  const { b, t } = useSite()
  const meta = useCertMeta()
  return (
    <div className="grid gap-4" style={{ gridTemplateColumns: 'repeat(auto-fill,minmax(min(100%,260px),1fr))' }}>
      {b.certificates.map((c) => (
        <div key={c.name} className="overflow-hidden rounded border border-line-2">
          <div className="box-border flex h-[220px] justify-center overflow-hidden bg-cotton px-5 pt-5">
            <div role="img" aria-label={t('certs.aria', { name: c.name })} className="w-[72%] self-start shadow-card"
              style={{ aspectRatio: '1/1.414', background: `#FFFFFF ${cssUrl(c.image)} top center/cover no-repeat` }} />
          </div>
          <div className="border-t border-line-2 px-5 pb-5 pt-4">
            <p className="m-0 text-[18px] font-[550]">{c.name}</p>
            <p className="m-0 mt-1 text-[14px] text-ink-3">{meta(c)}</p>
          </div>
        </div>
      ))}
    </div>
  )
}

/** Food Services: the badge row. */
export function CertificateBadges() {
  const { b } = useSite()
  return (
    <ul className="m-0 flex list-none flex-wrap gap-3 p-0">
      {b.certificates.map((c) => (
        <li key={c.name} className="flex items-center gap-3 rounded border border-line-2 px-[18px] py-[14px] text-[16px] font-medium">
          <SealCheck aria-hidden="true" className="text-[24px] text-red" />{c.name}
        </li>
      ))}
    </ul>
  )
}
