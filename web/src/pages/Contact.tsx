import { useEffect, useMemo, useRef, useState, type CSSProperties, type FormEvent } from 'react'
import { useLocation, useSearchParams } from 'react-router-dom'
import { ArrowRight, ArrowUpRight, CheckCircle, Clock, EnvelopeSimple, MapPin, Phone, WhatsappLogo } from '@/lib/phosphor'
import { useSite } from '@/lib/site'
import { usePageMeta } from '@/lib/meta'
import { Breadcrumb, PH } from '@/components/bits'
import { telHref, waHref } from '@/components/Chrome'
import { EMAIL_RE, fieldErrors, formError, HONEYPOT_STYLE, postJSON } from '@/lib/forms'

const EMPTY = { name: '', company: '', email: '', phone: '', buildingType: '', area: '', location: '', message: '' }
type F = typeof EMPTY

export default function Contact() {
  const { b, t, L, lang, section, param, paramPH } = useSite()
  usePageMeta(t('seo.contact.title') + t('seo.suffix'), t('seo.contact.description'))
  const loc = useLocation()
  const [qs] = useSearchParams()
  const hero = section('contact.hero')
  const lim = b.rules.limits

  // Prefill from the Home quick form (route state) and ?service=<line>.
  const pre = (loc.state || {}) as Partial<F> & { services?: string[] }
  const [f, setF] = useState<F>(() => ({ ...EMPTY, name: pre.name || '', company: pre.company || '', buildingType: pre.buildingType || '' }))
  const [svc, setSvc] = useState<string[]>(() => {
    const s = qs.get('service')
    const fromState = (pre.services || []).filter((k) => b.lines.some((l) => l.key === k))
    if (s && b.lines.some((l) => l.key === s) && !fromState.includes(s)) return [...fromState, s]
    return fromState
  })
  const [hp, setHp] = useState('')
  const [sending, setSending] = useState(false)
  const [sent, setSent] = useState<string | null>(null)
  const [err, setErr] = useState('')
  const [ferr, setFerr] = useState<Record<string, string>>({})
  const started = useRef(Date.now())
  const statusRef = useRef<HTMLDivElement>(null)
  useEffect(() => { if (sent !== null) statusRef.current?.focus() }, [sent])

  const set = (k: keyof F) => (e: { target: { value: string } }) => {
    const v = k === 'area' ? e.target.value.replace(/[^0-9]/g, '') : e.target.value
    setF((s) => ({ ...s, [k]: v }))
    setFerr((x) => ({ ...x, [k]: '' }))
  }

  const missing = useMemo(() => {
    const m: string[] = []
    if (!f.name.trim()) m.push(t('contact.form.field.name'))
    if (!f.company.trim()) m.push(t('contact.form.field.company'))
    if (!EMAIL_RE.test(f.email.trim())) m.push(t('contact.form.field.email'))
    if (!f.phone.trim()) m.push(t('contact.form.field.phone'))
    if (!f.buildingType) m.push(t('contact.form.field.building_type'))
    if (!f.location.trim()) m.push(t('contact.form.field.location'))
    if (svc.length === 0) m.push(t('contact.form.field.services'))
    return m
  }, [f, svc, t])
  const cantSend = missing.length > 0 || sending

  const submit = async (e: FormEvent) => {
    e.preventDefault()
    if (missing.length) return
    setSending(true)
    setErr('')
    try {
      const r = await postJSON<{ firstName: string }>('/api/v1/public/quotation', {
        ...f, services: svc, lang, website: hp, elapsedMs: Date.now() - started.current,
      })
      setSent(r.firstName || f.name.trim().split(/\s+/)[0] || '')
    } catch (x) {
      setErr(formError(x, t))
      setFerr(fieldErrors(x, t, b.rules.cvMaxMb))
    } finally {
      setSending(false)
    }
  }
  const reset = () => { setF(EMPTY); setSvc([]); setSent(null); setErr(''); setFerr({}); started.current = Date.now() }

  const lat = param('contact.map_lat'), lng = param('contact.map_lng')
  const mapSrc = `https://maps.google.com/maps?q=${encodeURIComponent(lat + ',' + lng)}&z=16&hl=${lang}&output=embed`
  const mapLink = `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(lat + ',' + lng)}`

  const field = (k: keyof F, label: string, input: JSX.Element, full = false) => (
    <label className={`field${full ? ' col-span-full' : ''}`}>
      {label}
      {input}
      {ferr[k] && <span className="err" id={`e-${k}`}>{ferr[k]}</span>}
    </label>
  )
  const a = (k: keyof F) => ({ value: f[k], onChange: set(k), className: 'input', 'aria-invalid': !!ferr[k] || undefined, 'aria-describedby': ferr[k] ? `e-${k}` : undefined })

  return (
    <>
      <section className="on-dark bg-charcoal text-white">
        <div className="wrap" style={{ paddingTop: 'clamp(40px,6vw,80px)', paddingBottom: 'clamp(40px,6vw,80px)' }}>
          <div className="mb-7"><Breadcrumb items={[{ label: t('nav.home'), to: '/' }, { label: t('contact.kicker') }]} /></div>
          <h1 className="t-page max-w-[820px]">{L(hero?.title)}</h1>
          <p className="t-lead mb-0 mt-[22px] max-w-[600px] text-ondark-body">{L(hero?.body)}</p>
        </div>
      </section>

      <section className="wrap flex flex-wrap items-start" style={{ gap: '32px clamp(32px,4vw,56px)', paddingTop: 'clamp(32px,5vw,64px)', paddingBottom: 'clamp(56px,7vw,96px)' }}>
        <div className="min-w-0 flex-[1.6_1_460px]">
          {sent !== null ? (
            <div ref={statusRef} tabIndex={-1} role="status" className="flex flex-col items-start gap-3 rounded border border-line-2 outline-none" style={{ padding: 'clamp(24px,4vw,40px)' }}>
              <CheckCircle aria-hidden="true" className="text-[44px] text-success" />
              <h2 className="m-0 text-[28px] font-[550] tracking-[-0.02em]">{t('contact.form.sent_title')}</h2>
              <p className="m-0 max-w-[520px] text-[16px] leading-[1.55] text-ink-2">{t('contact.form.sent_body', { name: sent })}</p>
              <button type="button" onClick={reset} className="btn-quiet mt-2">{t('contact.form.another')}</button>
            </div>
          ) : (
            <form onSubmit={submit} noValidate aria-labelledby="form-h" className="relative grid gap-[18px] rounded border border-line-2"
              style={{ padding: 'clamp(20px,3vw,36px)', gridTemplateColumns: 'repeat(auto-fit,minmax(min(100%,230px),1fr))' }}>
              <div className="col-span-full">
                <h2 id="form-h" className="m-0 font-[550] tracking-[-0.02em]" style={{ fontSize: 'clamp(24px,2.6vw,32px)' }}>{L(section('contact.form')?.title)}</h2>
                <p className="mb-0 mt-2 text-[14px] text-ink-3">{t('contact.required_note')}</p>
              </div>
              {field('name', t('form.name') + ' *', <input {...a('name')} autoComplete="name" maxLength={lim.name} />)}
              {field('company', t('form.company') + ' *', <input {...a('company')} autoComplete="organization" maxLength={lim.company} />)}
              {field('email', t('contact.form.email') + ' *', <input {...a('email')} type="email" autoComplete="email" maxLength={lim.email} />)}
              {field('phone', t('contact.form.phone') + ' *', <input {...a('phone')} type="tel" autoComplete="tel" maxLength={lim.phone} />)}
              {field('buildingType', t('form.building_type') + ' *', (
                <select {...a('buildingType')}>
                  <option value="">{t('form.choose_type')}</option>
                  {b.buildingTypes.map((o) => <option key={o.key} value={o.key}>{L(o.label)}</option>)}
                </select>
              ))}
              {field('area', t('contact.form.area'), (
                <span className="relative block">
                  <input {...a('area')} inputMode="numeric" placeholder={t('contact.form.area_ph')} className="input pr-12" maxLength={9} />
                  <span aria-hidden="true" className="absolute right-[14px] top-1/2 -translate-y-1/2 text-[15px] text-ink-3">m²</span>
                </span>
              ))}
              {field('location', t('contact.form.location') + ' *', <input {...a('location')} placeholder={t('contact.form.location_ph')} maxLength={lim.location} />, true)}
              <fieldset className="col-span-full m-0 border-0 p-0" aria-describedby={ferr.services ? 'e-services' : undefined}>
                <legend className="mb-2 text-[14px] font-medium">{t('form.services')} *</legend>
                <div className="flex flex-wrap gap-2">
                  {b.lines.map((l) => {
                    const on = svc.includes(l.key)
                    return (
                      <label key={l.key} className={`chip${on ? ' on' : ''}`}>
                        <input type="checkbox" checked={on} onChange={() => { setSvc((s) => (on ? s.filter((x) => x !== l.key) : [...s, l.key])); setFerr((x) => ({ ...x, services: '' })) }} />
                        <span className="sq" style={{ '--sq': l.base } as CSSProperties} />{L(l.label)}
                      </label>
                    )
                  })}
                </div>
                {ferr.services && <span className="err mt-1 block" id="e-services">{ferr.services}</span>}
              </fieldset>
              {field('message', t('contact.form.message'), <textarea {...a('message')} rows={5} placeholder={t('contact.form.message_ph')} maxLength={lim.message} />, true)}
              <div style={HONEYPOT_STYLE} aria-hidden="true">
                <label>Website<input tabIndex={-1} autoComplete="off" value={hp} onChange={(e) => setHp(e.target.value)} /></label>
              </div>
              <div className="col-span-full flex flex-wrap items-center gap-3">
                <button type="submit" disabled={cantSend} aria-describedby="send-why" className="btn btn-red">
                  {sending ? t('contact.form.sending') : t('contact.form.send')} <ArrowRight aria-hidden="true" />
                </button>
                <p id="send-why" aria-live="polite" className="m-0 flex-[1_1_220px] text-[13px] text-ink-2">
                  {missing.length ? t('contact.form.missing', { fields: missing.join(', ') }) : t('contact.form.ready')}
                </p>
              </div>
              {err && <p role="alert" className="col-span-full m-0 rounded border border-red px-4 py-3 text-[14px] text-red-hover">{err}</p>}
            </form>
          )}
        </div>

        <aside className="flex min-w-0 flex-[1_1_320px] flex-col gap-3">
          <a href={waHref(param('contact.whatsapp_digits'))} target="_blank" rel="noopener noreferrer"
            className="flex items-center gap-4 rounded bg-wa-card p-[22px] text-white hover:bg-wa-cardHover hover:text-white">
            <span className="grid h-[52px] w-[52px] flex-none place-items-center rounded-full bg-white text-[30px] text-wa-card"><WhatsappLogo weight="fill" aria-hidden="true" /></span>
            <span className="flex flex-col gap-1">
              <span className="text-[19px] font-[550]">{t('wa.chat')}</span>
              <span className="text-[15px]">{param('contact.whatsapp')} · {t('contact.wa_note')} {paramPH('contact.whatsapp') && <span className="ml-1 rounded border border-dashed border-white px-1.5 py-0.5 text-[10px] font-[550] uppercase tracking-[0.05em]">{t('common.placeholder')}</span>}</span>
            </span>
          </a>
          <div className="flex flex-col rounded border border-line-2 p-[22px]">
            <InfoRow icon={<MapPin aria-hidden="true" />} label={t('contact.office')} border>
              <p className="m-0 text-[16px] leading-[1.5]">{param('contact.address')} <PH show={paramPH('contact.address')} small /></p>
            </InfoRow>
            {param('contact.phone').trim() && (
              <InfoRow icon={<Phone aria-hidden="true" />} label={t('contact.phone_label')} border>
                <a href={telHref(param('contact.phone'))} className="text-[16px] hover:text-red">{param('contact.phone')}</a> <PH show={paramPH('contact.phone')} small />
              </InfoRow>
            )}
            <InfoRow icon={<EnvelopeSimple aria-hidden="true" />} label={t('common.email')} border>
              <a href={`mailto:${param('contact.email')}`} className="break-all text-[16px] hover:text-red">{param('contact.email')}</a> <PH show={paramPH('contact.email')} small />
            </InfoRow>
            <InfoRow icon={<Clock aria-hidden="true" />} label={t('contact.hours_label')}>
              <p className="m-0 text-[16px] leading-[1.5]">{param('contact.hours')}</p>
            </InfoRow>
          </div>
          <div className="overflow-hidden rounded border border-line-2">
            <div className="relative">
              <iframe title={t('contact.map_title')} src={mapSrc} loading="lazy" referrerPolicy="no-referrer-when-downgrade"
                className="block h-[280px] w-full border-0" style={{ filter: 'grayscale(0.6)' }} />
              <PH show={paramPH('contact.map_lat') || paramPH('contact.map_lng')} small className="absolute left-2 top-2 bg-white" />
            </div>
            <a href={mapLink} target="_blank" rel="noopener noreferrer" className="flex items-center justify-between gap-2 border-t border-line-2 px-4 py-[14px] text-[15px] font-medium hover:text-red">
              {t('contact.open_map')} <ArrowUpRight aria-hidden="true" />
            </a>
          </div>
        </aside>
      </section>
    </>
  )
}

function InfoRow({ icon, label, children, border }: { icon: JSX.Element; label: string; children: React.ReactNode; border?: boolean }) {
  return (
    <div className={`grid grid-cols-[32px_1fr] gap-3 ${border ? 'border-b border-line-1 py-3' : 'pb-0 pt-3'}`}>
      <span className="text-[22px] text-red">{icon}</span>
      <div>
        <p className="mb-1 mt-0 text-[13px] font-semibold uppercase tracking-[0.06em] text-ink-3">{label}</p>
        {children}
      </div>
    </div>
  )
}
