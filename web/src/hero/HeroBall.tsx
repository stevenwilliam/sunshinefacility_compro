import { useEffect, useMemo, useRef, useState, type CSSProperties } from 'react'
import { Link } from 'react-router-dom'
import { ArrowLeft, ArrowRight, ArrowsIn, CheckCircle, HandGrabbing, X } from '@/lib/phosphor'
import { useSite } from '@/lib/site'
import type { BallApi, BallItem } from './scene'

// The 3D work starts after the page has painted and the browser is idle:
// the hero text is usable immediately, and on a WebGL or import failure it
// simply stays the only thing there (README → Hero ball → Performance).
function afterFirstPaint(fn: () => void) {
  let cancelled = false
  const go = () => { if (!cancelled) fn() }
  const w = window as Window & { requestIdleCallback?: (c: () => void, o?: { timeout: number }) => number }
  const idle = (cb: () => void) => (w.requestIdleCallback ? w.requestIdleCallback(cb, { timeout: 1500 }) : setTimeout(cb, 200))
  requestAnimationFrame(() => requestAnimationFrame(() => idle(go)))
  return () => { cancelled = true }
}

const pill: CSSProperties = {
  display: 'flex', alignItems: 'center', gap: 8, background: 'rgba(22,24,19,0.78)', border: '1px solid #44463E', color: '#EDEDE6',
  borderRadius: 999, padding: '9px 14px', fontSize: 14, fontWeight: 500, backdropFilter: 'blur(6px)', fontFamily: 'inherit',
}

export function HeroBall({ side, onFocusChange }: { side: boolean; onFocusChange: (k: number) => void }) {
  const { b, t, L } = useSite()
  const wrapRef = useRef<HTMLDivElement>(null)
  const hostRef = useRef<HTMLDivElement>(null)
  const hudRef = useRef<HTMLDivElement>(null)
  const tagRef = useRef<HTMLButtonElement>(null)
  const labelRefs = useRef<(HTMLDivElement | null)[]>([])
  const sideRef = useRef(side)
  sideRef.current = side
  const api = useRef<BallApi | null>(null)
  const [isSplit, setIsSplit] = useState(false)
  const [focus, setFocus] = useState(-1)
  const [status, setStatus] = useState<'loading' | 'ready' | 'error'>('loading')

  const lines = b.lines
  const items: BallItem[] = useMemo(() => lines.map((l) => ({
    key: l.key, n: l.n, base: l.base, ink: l.ink, onInk: l.onInk,
    photoUrl: l.photo?.src || '', markUrl: l.mark?.src || '',
  })), [lines])

  useEffect(() => { onFocusChange(focus) }, [focus, onFocusChange])
  useEffect(() => {
    const k = (e: KeyboardEvent) => { if (e.key === 'Escape') api.current?.focus(-1) }
    window.addEventListener('keydown', k)
    return () => window.removeEventListener('keydown', k)
  }, [])

  useEffect(() => {
    if (items.length === 0) { setStatus('error'); return }
    let alive = true
    const cancel = afterFirstPaint(async () => {
      try {
        const [THREE, mod] = await Promise.all([import('three'), import('./scene')])
        if (!alive || !wrapRef.current || !hostRef.current) return
        api.current = mod.createBall(THREE, items, {
          wrap: wrapRef.current, host: hostRef.current,
          hud: () => hudRef.current, tag: () => tagRef.current, label: (k) => labelRefs.current[k],
        }, () => sideRef.current, { split: setIsSplit, focus: setFocus })
        setStatus('ready')
      } catch {
        if (alive) setStatus('error')
      }
    })
    return () => { alive = false; cancel(); api.current?.destroy(); api.current = null }
  }, [items])

  const fItem = focus >= 0 ? lines[focus] : null
  const region: CSSProperties = side ? { left: '46%', right: 'clamp(20px,4vw,48px)' } : { left: 16, right: 16 }
  const nb = (label: string, icon: JSX.Element, fn: () => void) => (
    <button type="button" aria-label={label} onClick={fn}
      className="grid h-11 w-11 cursor-pointer place-items-center rounded border border-line-3 bg-white text-[18px] text-charcoal hover:border-charcoal">{icon}</button>
  )

  return (
    <div ref={wrapRef} className="absolute inset-0 overflow-hidden">
      <div ref={hostRef} aria-hidden="true" className="absolute inset-0" />
      <div ref={hudRef} aria-hidden="true" data-hud="" className="pointer-events-none absolute left-0 top-0 h-[200px] w-[200px] opacity-0" style={{ willChange: 'transform' }}>
        <div className="absolute inset-2 rounded-full" style={{ border: '1.5px solid rgba(253,87,42,0.7)', boxShadow: '0 0 18px rgba(234,41,21,0.45), inset 0 0 18px rgba(234,41,21,0.25)' }} />
        {!isSplit && <div className="absolute inset-2 rounded-full" style={{ border: '1.5px solid rgba(253,87,42,0.6)', animation: 'sfgPing 2.6s ease-out infinite' }} />}
      </div>
      {status === 'loading' && (
        <p className="absolute m-0 -translate-x-1/2 -translate-y-1/2 text-[15px] text-ondark-muted" style={{ left: side ? '72%' : '50%', top: '46%' }}>{t('home.ball.loading')}</p>
      )}
      {status === 'ready' && (
        <button ref={tagRef} type="button" onClick={() => api.current?.split(true)}
          className="absolute left-0 top-0 flex min-h-[44px] cursor-pointer items-center gap-[10px] whitespace-nowrap rounded-full border border-tangerine px-[18px] text-[15px] font-medium text-white opacity-0"
          style={{ background: 'rgba(22,24,19,0.82)', boxShadow: '0 0 28px rgba(234,41,21,0.35)', backdropFilter: 'blur(6px)' }}>
          <span className="h-2 w-2 rounded-full bg-tangerine" style={{ boxShadow: '0 0 10px #FD572A', animation: 'sfgDot 1.6s ease-in-out infinite' }} />
          {t('home.ball.open')} <span className="text-[13px] text-ondark-muted">{t('home.ball.drag')}</span>
        </button>
      )}
      {lines.map((l, k) => (
        <div key={l.key} ref={(el) => { labelRefs.current[k] = el }} aria-hidden="true"
          className="pointer-events-none absolute left-0 top-0 flex items-center gap-2 whitespace-nowrap pl-2 text-[15px] font-medium text-white opacity-0"
          style={{ textShadow: '0 1px 8px rgba(22,24,19,0.9)', borderLeft: '3px solid ' + l.base }}>
          <span className="rounded-[3px] px-1.5 py-[3px] text-[12px] font-semibold" style={{ background: l.ink, color: l.onInk, textShadow: 'none' }}>{l.n}</span>{L(l.title)}
        </div>
      ))}
      {isSplit && (
        <div className="pointer-events-none absolute top-5 flex flex-wrap justify-between gap-[10px]" style={region}>
          <span style={pill}><HandGrabbing aria-hidden="true" style={{ color: '#FD572A', fontSize: 17 }} />{t('home.ball.hint_split')}</span>
          <button type="button" onClick={() => api.current?.split(false)} style={{ ...pill, cursor: 'pointer', pointerEvents: 'auto', minHeight: 44 }}>
            <ArrowsIn aria-hidden="true" style={{ color: '#FD572A', fontSize: 17 }} />{t('home.ball.reassemble')}
          </button>
        </div>
      )}
      {fItem && (
        <div role="region" aria-label={L(fItem.title)} aria-live="polite" className="absolute z-[3] box-border overflow-y-auto rounded shadow-modal"
          style={{
            ...(side ? { left: 'clamp(20px,4vw,48px)', top: '50%', transform: 'translateY(-50%)', width: 'min(460px, 38vw)', maxHeight: 'calc(100% - 48px)' } : { left: 12, right: 12, bottom: 12, maxHeight: 'calc(100% - 24px)' }),
            background: fItem.ink, color: fItem.onInk,
          }}>
          <div className="flex items-start justify-between gap-3 bg-white" style={{ padding: side ? '20px 22px 40px' : '16px 16px 32px', clipPath: 'polygon(0 0,100% 0,100% 58%,0 100%)' }}>
            {fItem.mark && <img src={fItem.mark.src} alt="" style={{ width: side ? 60 : 48, height: 'auto', display: 'block' }} />}
            <div className="flex gap-1.5">
              {nb(t('home.ball.prev'), <ArrowLeft aria-hidden="true" />, () => api.current?.focus((focus + lines.length - 1) % lines.length))}
              {nb(t('home.ball.next'), <ArrowRight aria-hidden="true" />, () => api.current?.focus((focus + 1) % lines.length))}
              {nb(t('home.ball.back'), <X aria-hidden="true" />, () => api.current?.focus(-1))}
            </div>
          </div>
          <div style={{ padding: side ? '4px 28px 28px' : '0 20px 20px' }}>
            <p className="m-0 text-[13px] font-semibold tracking-[0.04em] tnum">{fItem.n} / {String(lines.length).padStart(2, '0')}</p>
            <h3 className="mb-0 mt-1.5 font-[550] leading-[1.05] tracking-[-0.025em]" style={{ fontSize: side ? 32 : 26 }}>{L(fItem.title)}</h3>
            <p className="mb-0 mt-[14px] text-[16px] leading-[1.55]">{L(fItem.summary)}</p>
            <p className="mb-2 mt-[22px] text-[13px] font-semibold uppercase tracking-[0.06em]">{t('home.ball.included')}</p>
            <ul className="m-0 flex list-none flex-col p-0">
              {fItem.heroPoints.map((p, i) => (
                <li key={i} className="grid grid-cols-[24px_1fr] gap-[10px] py-[10px] text-[15px] leading-[1.45]"
                  style={{ borderTop: '1px solid ' + (fItem.onInk === '#FFFFFF' ? 'rgba(255,255,255,0.28)' : 'rgba(31,33,28,0.28)') }}>
                  <CheckCircle weight="fill" aria-hidden="true" className="mt-px text-[20px]" />{L(p)}
                </li>
              ))}
            </ul>
            <div className="mt-[22px] flex flex-wrap gap-[10px]">
              <a href="#quote" className="flex items-center gap-[10px] rounded bg-white px-[18px] py-[14px] text-[15px] font-[550] text-charcoal hover:bg-cotton">
                {t('home.ball.quote')} <ArrowRight aria-hidden="true" />
              </a>
              <Link to={`/services/${fItem.key}`} className="flex items-center gap-[10px] rounded px-[18px] py-[13px] text-[15px] font-medium"
                style={{ border: '1.5px solid ' + fItem.onInk, color: fItem.onInk }}>{t('home.ball.learn')}</Link>
            </div>
          </div>
        </div>
      )}
      {/* As in the design the chips show while the scene loads (a click then
          does nothing yet). If WebGL is unavailable they are removed and the
          hero text is the whole hero. */}
      {(side || isSplit) && status !== 'error' && (
        <div role="group" aria-label={t('home.ball.aria')} className="absolute flex gap-2" style={{ ...region, bottom: side ? 28 : 16, flexWrap: side ? 'wrap' : 'nowrap', overflowX: side ? 'visible' : 'auto', scrollbarWidth: 'none' }}>
          {lines.map((l, k) => {
            const on = focus === k
            return (
              <button key={l.key} type="button" aria-pressed={on} onClick={() => api.current?.focus(on ? -1 : k)}
                className="flex min-h-[44px] flex-none cursor-pointer items-center gap-2 rounded-full px-[14px] text-[14px] font-medium"
                style={{ border: '1px solid ' + (on ? l.ink : '#44463E'), background: on ? l.ink : 'rgba(22,24,19,0.78)', color: on ? l.onInk : '#EDEDE6', backdropFilter: 'blur(6px)' }}>
                <span className="text-[12px] font-semibold" style={{ color: on ? l.onInk : l.base }}>{l.n}</span>{L(l.title)}
              </button>
            )
          })}
        </div>
      )}
    </div>
  )
}
