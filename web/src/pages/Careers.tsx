import { useEffect, useMemo, useRef, useState, type CSSProperties, type FormEvent } from 'react'
import { ArrowDown, ArrowRight, CaretDown, CaretUp, Check, CheckCircle, Clock, LinkedinLogo, MagnifyingGlass, MapPin, PaperPlaneTilt } from '@/lib/phosphor'
import { lineVars, useSite } from '@/lib/site'
import { usePageMeta } from '@/lib/meta'
import { Icon } from '@/lib/icons'
import { Breadcrumb, cssUrl, PH } from '@/components/bits'
import { EMAIL_RE, fieldErrors, formError, HONEYPOT_STYLE, postForm } from '@/lib/forms'

const OPEN = 'open'

export default function Careers() {
  const { b, t, L, section } = useSite()
  usePageMeta(t('seo.careers.title') + t('seo.suffix'), t('seo.careers.description'))
  const hero = section('careers.hero')
  const work = b.lines.find((l) => l.key === 'workforce')
  const [position, setPosition] = useState('')
  const applyRef = useRef<HTMLElement>(null)

  const choose = (id: string) => {
    setPosition(id)
    const el = applyRef.current
    if (el) window.scrollTo({ top: el.getBoundingClientRect().top + window.scrollY - 80, behavior: 'smooth' })
  }

  return (
    <>
      <section className="wrap" style={{ paddingTop: 'clamp(24px,4vw,48px)' }}>
        <div className="mb-5"><Breadcrumb items={[{ label: t('nav.home'), to: '/' }, { label: t('careers.kicker') }]} /></div>
        <div className="poster" style={{ ...(work ? lineVars(work) : {}), minHeight: 'clamp(460px,48vw,560px)' }}>
          <div className="poster-head h45" style={{ padding: 'clamp(20px,3vw,32px) clamp(20px,3vw,40px) clamp(56px,7vw,96px)' }}>
            <div aria-hidden="true" className="mark" style={{ width: 'clamp(64px,7vw,96px)', aspectRatio: '1/1', backgroundImage: cssUrl(work?.mark) }} />
          </div>
          <div className="relative z-[1] flex max-w-[620px] flex-col gap-[18px]" style={{ padding: '0 clamp(20px,3vw,40px) clamp(28px,4vw,48px)' }}>
            <h1 className="t-page" style={{ lineHeight: 0.98 }}>{L(hero?.title)}</h1>
            <p className="t-lead m-0 max-w-[480px]">{L(hero?.body)}</p>
            <a href="#openings" className="btn btn-white self-start">{t('careers.see')} <ArrowDown aria-hidden="true" /></a>
          </div>
          <div aria-hidden="true" className="poster-photo -bottom-[6%] -right-[4%]"
            style={{ width: 'min(52%,540px)', backgroundImage: cssUrl(hero?.image), WebkitMaskImage: cssUrl(work?.mark), maskImage: cssUrl(work?.mark) }} />
        </div>
      </section>

      <section aria-labelledby="why-h" className="wrap" style={{ paddingTop: 'clamp(56px,7vw,96px)' }}>
        <h2 id="why-h" className="t-h2s mb-8 max-w-[640px]">{L(section('careers.reasons')?.title)}</h2>
        <div className="grid gap-x-6 gap-y-8" style={{ gridTemplateColumns: 'repeat(auto-fit,minmax(230px,1fr))' }}>
          {(b.lists.careers_reasons || []).map((w, i) => (
            <div key={i} className="flex flex-col gap-[10px] border-t-2 border-charcoal pt-[18px]">
              <Icon name={w.icon} className="text-[30px]" style={{ color: work?.ink || '#4A7023' }} />
              <h3 className="m-0 mt-1.5 text-[20px] font-[550] tracking-[-0.01em]">{L(w.title)}</h3>
              <p className="m-0 text-[15px] leading-[1.55] text-ink-2">{L(w.body)}</p>
            </div>
          ))}
        </div>
      </section>

      <Openings onApply={choose} />

      <section id="apply" ref={applyRef} aria-labelledby="ap-h" className="wrap scroll-mt-[88px]" style={{ paddingTop: 'clamp(56px,7vw,96px)', paddingBottom: 'clamp(56px,7vw,96px)' }}>
        <div className="on-dark flex flex-wrap rounded bg-charcoal text-white" style={{ padding: 'clamp(24px,4vw,48px)', gap: '32px clamp(32px,5vw,72px)' }}>
          <div className="min-w-0 max-w-[420px] flex-[1_1_280px]">
            <h2 id="ap-h" className="t-h2s">{L(section('careers.apply')?.title)}</h2>
            <p className="mb-0 mt-4 text-[16px] leading-[1.55] text-ondark-body">{L(section('careers.apply')?.body)}</p>
          </div>
          <ApplyForm position={position} setPosition={setPosition} />
        </div>
      </section>
    </>
  )
}

function Openings({ onApply }: { onApply: (id: string) => void }) {
  const { b, t, L, lang, section } = useSite()
  const [q, setQ] = useState('')
  const [fLine, setFLine] = useState('')
  const [fLoc, setFLoc] = useState('')
  const [fType, setFType] = useState('')
  const [open, setOpen] = useState('')
  const locLabel = (k: string) => L(b.locations.find((o) => o.key === k)?.label) || k
  const typeLabel = (k: string) => L(b.jobTypes.find((o) => o.key === k)?.label) || k
  const lineOf = (k: string) => b.lines.find((l) => l.key === k)
  const locs = [...new Set(b.openings.map((o) => o.location))].sort((a, z) => locLabel(a).localeCompare(locLabel(z)))
  const types = [...new Set(b.openings.map((o) => o.type))]
  const lines = b.lines.filter((l) => b.openings.some((o) => o.line === l.key))
  const ql = q.trim().toLowerCase()
  const list = b.openings.filter((o) =>
    (!fLine || o.line === fLine) && (!fLoc || o.location === fLoc) && (!fType || o.type === fType) &&
    (!ql || [L(o.title), L(o.summary), L(lineOf(o.line)?.label), locLabel(o.location)].join(' ').toLowerCase().includes(ql)))
  const noFilters = !q && !fLine && !fLoc && !fType
  const anyPH = b.openings.some((o) => o.placeholder)
  const sel = 'h-[46px] rounded border border-line-4 bg-white px-[10px] text-[16px] text-charcoal'
  return (
    <section id="openings" aria-labelledby="op-h" className="wrap scroll-mt-[88px]" style={{ paddingTop: 'clamp(56px,7vw,96px)' }}>
      <div className="mb-6 flex flex-wrap items-center gap-x-5 gap-y-3">
        <h2 id="op-h" className="t-h2s">{L(section('careers.openings')?.title)}</h2>
        <PH show={anyPH} label={t('careers.sample')} />
      </div>
      <div role="search" className="grid gap-[10px] rounded bg-cotton p-4" style={{ gridTemplateColumns: 'repeat(auto-fit,minmax(min(100%,200px),1fr))' }}>
        <label className="flex flex-col gap-1.5 text-[13px] font-medium">{t('careers.search')}
          <span className="relative block">
            <MagnifyingGlass aria-hidden="true" className="absolute left-3 top-1/2 -translate-y-1/2 text-[18px] text-ink-3" />
            <input type="search" value={q} onChange={(e) => setQ(e.target.value)} placeholder={t('careers.search_ph')}
              className="box-border h-[46px] w-full rounded border border-line-4 bg-white pl-[38px] pr-3 text-[16px] text-charcoal" />
          </span>
        </label>
        <label className="flex flex-col gap-1.5 text-[13px] font-medium">{t('careers.filter_line')}
          <select value={fLine} onChange={(e) => setFLine(e.target.value)} className={sel}>
            <option value="">{t('careers.all')}</option>
            {lines.map((l) => <option key={l.key} value={l.key}>{L(l.label)}</option>)}
          </select>
        </label>
        <label className="flex flex-col gap-1.5 text-[13px] font-medium">{t('careers.filter_location')}
          <select value={fLoc} onChange={(e) => setFLoc(e.target.value)} className={sel}>
            <option value="">{t('careers.all')}</option>
            {locs.map((k) => <option key={k} value={k}>{locLabel(k)}</option>)}
          </select>
        </label>
        <label className="flex flex-col gap-1.5 text-[13px] font-medium">{t('careers.filter_type')}
          <select value={fType} onChange={(e) => setFType(e.target.value)} className={sel}>
            <option value="">{t('careers.all')}</option>
            {types.map((k) => <option key={k} value={k}>{typeLabel(k)}</option>)}
          </select>
        </label>
      </div>
      <div className="mb-2 mt-4 flex flex-wrap items-center justify-between gap-2">
        <p aria-live="polite" className="m-0 text-[14px] text-ink-2">{t('careers.count', { n: list.length, m: b.openings.length })}</p>
        <div className="flex items-center gap-[10px]">
          {noFilters && <span id="clear-why" className="text-[13px] text-ink-3">{t('careers.clear_why')}</span>}
          <button type="button" disabled={noFilters} aria-describedby={noFilters ? 'clear-why' : undefined}
            onClick={() => { setQ(''); setFLine(''); setFLoc(''); setFType('') }}
            className="min-h-[40px] cursor-pointer rounded border border-line-4 bg-white px-[14px] text-[14px] text-charcoal hover:border-charcoal disabled:cursor-not-allowed disabled:opacity-[.45] disabled:hover:border-line-4">
            {t('careers.clear')}
          </button>
        </div>
      </div>
      <ul className="m-0 list-none border-t border-line-2 p-0">
        {list.map((o) => {
          const l = lineOf(o.line)
          const on = open === o.id
          const reqs = o.requirements[lang === 'id' ? 1 : 0].length ? o.requirements[lang === 'id' ? 1 : 0] : o.requirements[0]
          return (
            <li key={o.id} className="border-b border-line-2">
              <button type="button" onClick={() => setOpen(on ? '' : o.id)} aria-expanded={on} aria-controls={`job-${o.id}`}
                className="flex w-full cursor-pointer flex-wrap items-center gap-x-5 gap-y-[10px] border-0 bg-transparent py-5 text-left text-charcoal">
                <span className="flex min-w-0 flex-[1_1_260px] flex-col gap-1.5">
                  <span className="text-[20px] font-[550] tracking-[-0.01em]">{L(o.title)}</span>
                  <span className="flex items-center gap-2 text-[14px] text-ink-2"><span className="sq" style={{ '--sq': l?.base } as CSSProperties} />{L(l?.label)}</span>
                </span>
                <span className="flex flex-wrap gap-x-4 gap-y-2 text-[14px] text-ink-2">
                  <span className="flex items-center gap-1.5"><MapPin aria-hidden="true" />{locLabel(o.location)}</span>
                  <span className="flex items-center gap-1.5"><Clock aria-hidden="true" />{typeLabel(o.type)}</span>
                </span>
                {on ? <CaretUp aria-hidden="true" className="text-[20px]" /> : <CaretDown aria-hidden="true" className="text-[20px]" />}
              </button>
              {on && (
                <div id={`job-${o.id}`} className="flex flex-wrap gap-x-12 gap-y-5 pb-6">
                  <p className="m-0 max-w-[520px] flex-[1_1_300px] text-[16px] leading-[1.55] text-ink-2">{L(o.summary)}</p>
                  <div className="min-w-0 flex-[1_1_280px]">
                    <p className="t-caps mb-2">{t('careers.requirements')}</p>
                    <ul className="m-0 list-none p-0">
                      {reqs.map((r, i) => <li key={i} className="grid grid-cols-[22px_1fr] gap-2 py-1.5 text-[15px] leading-[1.45]"><Check aria-hidden="true" className="mt-0.5 text-success" />{r}</li>)}
                    </ul>
                    <div className="mt-[14px] flex flex-wrap items-center gap-x-5 gap-y-3">
                      <button type="button" onClick={() => onApply(o.id)} className="btn btn-red px-[18px] py-[14px] text-[15px]">
                        {t('careers.apply_role')} <ArrowRight aria-hidden="true" />
                      </button>
                      {o.linkedin && (
                        <a href={o.linkedin} target="_blank" rel="noopener noreferrer" className="flex items-center gap-2 text-[15px] font-medium hover:text-red">
                          <LinkedinLogo aria-hidden="true" className="text-[20px]" />{t('careers.linkedin')}
                        </a>
                      )}
                    </div>
                  </div>
                </div>
              )}
            </li>
          )
        })}
      </ul>
      {list.length === 0 && (
        <div className="mt-3 rounded border border-dashed border-line-4 px-5 py-10 text-center text-[16px] text-ink-2">{t('careers.empty')}</div>
      )}
    </section>
  )
}

function ApplyForm({ position, setPosition }: { position: string; setPosition: (s: string) => void }) {
  const { b, t, L, lang } = useSite()
  const [f, setF] = useState({ name: '', phone: '', email: '', message: '' })
  const [file, setFile] = useState<File | null>(null)
  const [hp, setHp] = useState('')
  const [sending, setSending] = useState(false)
  const [sent, setSent] = useState(false)
  const [err, setErr] = useState('')
  const [ferr, setFerr] = useState<Record<string, string>>({})
  const started = useRef(Date.now())
  const fileRef = useRef<HTMLInputElement>(null)
  const statusRef = useRef<HTMLDivElement>(null)
  const mb = b.rules.cvMaxMb
  const lim = b.rules.limits
  const tooBig = !!file && file.size > mb * 1024 * 1024
  const badType = !!file && !/\.(pdf|docx?|DOCX?|PDF)$/.test(file.name)
  useEffect(() => { if (sent) statusRef.current?.focus() }, [sent])

  const missing = useMemo(() => {
    const m: string[] = []
    if (!f.name.trim()) m.push(t('careers.form.field.name'))
    if (!f.phone.trim()) m.push(t('careers.form.field.phone'))
    if (!EMAIL_RE.test(f.email.trim())) m.push(t('careers.form.field.email'))
    if (!position) m.push(t('careers.form.field.position'))
    if (!file || tooBig || badType) m.push(t('careers.form.field.cv'))
    return m
  }, [f, position, file, tooBig, badType, t])

  const submit = async (e: FormEvent) => {
    e.preventDefault()
    if (missing.length || !file) return
    setSending(true)
    setErr('')
    const fd = new FormData()
    fd.set('name', f.name); fd.set('phone', f.phone); fd.set('email', f.email); fd.set('message', f.message)
    fd.set('position', position); fd.set('lang', lang); fd.set('website', hp); fd.set('elapsedMs', String(Date.now() - started.current))
    fd.set('cv', file)
    try {
      await postForm('/api/v1/public/application', fd)
      setSent(true)
    } catch (x) {
      setErr(formError(x, t))
      setFerr(fieldErrors(x, t, mb))
    } finally {
      setSending(false)
    }
  }
  const reset = () => { setF({ name: '', phone: '', email: '', message: '' }); setFile(null); setPosition(''); setSent(false); setErr(''); setFerr({}); started.current = Date.now() }

  if (sent) {
    return (
      <div ref={statusRef} tabIndex={-1} role="status" className="flex flex-[1.4_1_380px] flex-col items-start gap-3 rounded bg-white p-8 text-charcoal outline-none">
        <CheckCircle aria-hidden="true" className="text-[40px] text-success" />
        <p className="m-0 text-[22px] font-[550]">{t('careers.form.sent_title')}</p>
        <p className="m-0 text-[16px] leading-[1.55] text-ink-2">{t('careers.form.sent_body')}</p>
        <button type="button" onClick={reset} className="btn-quiet mt-2">{t('careers.form.another')}</button>
      </div>
    )
  }
  const inp = (k: keyof typeof f) => ({
    value: f[k], className: 'input', 'aria-invalid': !!ferr[k] || undefined, 'aria-describedby': ferr[k] ? `ae-${k}` : undefined,
    onChange: (e: { target: { value: string } }) => { setF((s) => ({ ...s, [k]: e.target.value })); setFerr((x) => ({ ...x, [k]: '' })) },
  })
  const fe = (k: string) => (ferr[k] ? <span className="err" id={`ae-${k}`}>{ferr[k]}</span> : null)
  const fileNote = tooBig ? t('err.file_size', { mb }) : badType ? t('err.file_type') : file ? t('careers.form.file_ok', { name: file.name }) : t('careers.form.file_idle', { mb })
  return (
    <form onSubmit={submit} noValidate className="relative grid min-w-0 flex-[1.4_1_380px] gap-4 rounded bg-white text-charcoal"
      style={{ padding: 'clamp(20px,3vw,32px)', gridTemplateColumns: 'repeat(auto-fit,minmax(min(100%,220px),1fr))' }}>
      <label className="field">{t('careers.form.name')} *<input {...inp('name')} autoComplete="name" maxLength={lim.name} />{fe('name')}</label>
      <label className="field">{t('careers.form.phone')} *<input {...inp('phone')} type="tel" autoComplete="tel" maxLength={lim.phone} />{fe('phone')}</label>
      <label className="field">{t('careers.form.email')} *<input {...inp('email')} type="email" autoComplete="email" maxLength={lim.email} />{fe('email')}</label>
      <label className="field">{t('careers.form.position')} *
        <select className="input" value={position} onChange={(e) => { setPosition(e.target.value); setFerr((x) => ({ ...x, position: '' })) }} aria-invalid={!!ferr.position || undefined}>
          <option value="">{t('careers.form.choose')}</option>
          {b.openings.map((o) => <option key={o.id} value={o.id}>{L(o.title)}</option>)}
          <option value={OPEN}>{t('careers.form.open_application')}</option>
        </select>
        {fe('position')}
      </label>
      <label className="field col-span-full">{t('careers.form.cv')} *
        <input ref={fileRef} type="file" accept=".pdf,.doc,.docx,application/pdf,application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document"
          onChange={(e) => { setFile(e.target.files?.[0] || null); setFerr((x) => ({ ...x, cv: '' })) }}
          aria-invalid={tooBig || badType || !!ferr.cv || undefined} aria-describedby="cv-note"
          className="rounded border border-dashed border-line-4 bg-cotton p-[10px] text-[15px]" />
        <span id="cv-note" className="text-[13px] font-normal" style={{ color: tooBig || badType || ferr.cv ? '#A3181A' : '#5A5C54' }}>{ferr.cv || fileNote}</span>
      </label>
      <label className="field col-span-full">{t('careers.form.message')}<textarea {...inp('message')} rows={4} maxLength={lim.message} />{fe('message')}</label>
      <div style={HONEYPOT_STYLE} aria-hidden="true">
        <label>Website<input tabIndex={-1} autoComplete="off" value={hp} onChange={(e) => setHp(e.target.value)} /></label>
      </div>
      <div className="col-span-full flex flex-wrap items-center gap-3">
        <button type="submit" disabled={missing.length > 0 || sending} aria-describedby="apply-why" className="btn btn-red">
          {sending ? t('careers.form.sending') : t('careers.form.send')} <PaperPlaneTilt aria-hidden="true" />
        </button>
        <p id="apply-why" aria-live="polite" className="m-0 flex-[1_1_200px] text-[13px] text-ink-2">
          {missing.length ? t('careers.form.missing', { fields: missing.join(', ') }) : t('careers.form.ready')}
        </p>
      </div>
      {err && <p role="alert" className="col-span-full m-0 rounded border border-red px-4 py-3 text-[14px] text-red-hover">{err}</p>}
    </form>
  )
}
