import { useState, type CSSProperties, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { ArrowRight, CheckCircle, Phone, Quotes, SealCheck, WhatsappLogo } from '@/lib/phosphor'
import { lineVars, useSite } from '@/lib/site'
import { usePageMeta } from '@/lib/meta'
import { useMedia } from '@/lib/useMedia'
import { Icon } from '@/lib/icons'
import { cssUrl, PH, Stats } from '@/components/bits'
import { DividerStrip, telHref, waHref } from '@/components/Chrome'
import { RibbonSte, RibbonVen } from '@/components/Ribbon'
import { CertificateCards } from '@/components/Certificates'
import { HeroBall } from '@/hero/HeroBall'

export default function Home() {
  const site = useSite()
  const { b, t, L, section } = site
  usePageMeta(t('seo.home.title'), t('seo.home.description'))
  const side = useMedia('(min-width: 960px)')
  const [focus, setFocus] = useState(-1)
  const hero = section('home.hero')
  return (
    <>
      {/* 1. Hero */}
      <section aria-labelledby="hero-h" className="on-dark relative overflow-hidden text-white"
        style={{
          minHeight: side ? 'clamp(600px, calc(100svh - 72px), 900px)' : 'auto',
          background: `radial-gradient(ellipse 55% 75% at ${side ? '72%' : '50%'} 50%,#33362E 0%,#1F211C 60%,#161813 100%)`,
        }}>
        <div className="pointer-events-none relative z-[1] mx-auto box-border flex max-w-page flex-col justify-center transition-opacity duration-[350ms]"
          style={{
            padding: side ? 'clamp(40px,6vw,88px) clamp(20px,4vw,48px)' : '40px 20px 0',
            minHeight: side ? 'clamp(600px, calc(100svh - 72px), 900px)' : 'auto',
            opacity: side && focus >= 0 ? 0 : 1,
          }}>
          <div className="pointer-events-auto max-w-[540px]">
            <p className="eyebrow mb-5 tracking-[0.02em]">{t('home.hero.eyebrow')}</p>
            <h1 id="hero-h" className="t-hero text-white">
              {L(hero?.title)} <span className="font-normal text-ondark-dim">{t('home.hero.title_accent')}</span>
            </h1>
            <p className="mb-0 mt-6 max-w-[480px] text-ondark-body" style={{ fontSize: 'clamp(17px,1.35vw,19px)', lineHeight: 1.55, textWrap: 'pretty' } as CSSProperties}>
              {L(hero?.body)}
            </p>
            <div className="mt-8 flex flex-wrap gap-3">
              <Link to="/contact" className="btn btn-red rounded-lg">{t('cta.quote')} <ArrowRight aria-hidden="true" /></Link>
              <Link to="/services" className="btn btn-line-light rounded-lg">{t('home.hero.cta_services')}</Link>
            </div>
            <p className="mb-0 mt-7 flex flex-wrap items-center gap-2 text-[14px] text-ondark-muted">
              <SealCheck aria-hidden="true" className="text-[18px] text-tangerine" />{t('home.hero.cert_line')}
            </p>
          </div>
        </div>
        <div style={{ position: side ? 'absolute' : 'relative', inset: 0, height: side ? 'auto' : 460, zIndex: focus >= 0 ? 2 : 0 }}>
          <HeroBall side={side} onFocusChange={setFocus} />
        </div>
      </section>

      {/* 2. Trust bar — "Ste" ribbon behind; white tiles so it shows only in the gaps. */}
      <section aria-label={t('home.trust.aria')} className="relative overflow-hidden border-y border-line-2 bg-white">
        <RibbonSte />
        <div className="wrap relative z-[1]" style={{ paddingTop: 'clamp(28px,4vw,48px)', paddingBottom: 'clamp(28px,4vw,48px)' }}>
          <Stats items={b.stats.home || []} />
          <p className="mb-[18px] text-[13px] font-medium uppercase tracking-[0.06em] text-ink-3" style={{ marginTop: 'clamp(32px,4vw,48px)' }}>{t('home.trust.trusted')}</p>
          <ul className="m-0 grid list-none gap-2 p-0" style={{ gridTemplateColumns: 'repeat(auto-fill,minmax(118px,1fr))' }}>
            {b.clients.map((c) => (
              <li key={c.name} className="box-content grid h-[68px] place-items-center rounded-lg border border-line-1 bg-white px-[14px] py-[10px] opacity-[.72] grayscale transition-[filter,opacity] duration-200 hover:opacity-100 hover:grayscale-0">
                <div role="img" aria-label={c.name} className="h-10 w-full bg-contain bg-center bg-no-repeat" style={{ backgroundImage: cssUrl(c.logo) }} />
              </li>
            ))}
          </ul>
        </div>
      </section>

      {/* 3. Services */}
      <section id="services" aria-labelledby="svc-h" className="wrap" style={{ paddingTop: 'clamp(56px,7vw,112px)', paddingBottom: 'clamp(56px,7vw,112px)' }}>
        <div className="mb-10 flex flex-wrap items-end justify-between gap-x-12 gap-y-4">
          <div className="max-w-[620px]">
            <p className="t-kicker">{t('home.services.kicker')}</p>
            <h2 id="svc-h" className="t-h2">{L(section('home.services')?.title)}</h2>
          </div>
          <p className="m-0 max-w-[380px] text-[16px] leading-[1.55] text-ink-2">{L(section('home.services')?.body)}</p>
        </div>
        <div className="grid gap-3" style={{ gridTemplateColumns: 'repeat(auto-fit,minmax(min(100%,340px),1fr))' }}>
          {b.lines.map((l) => (
            <Link key={l.key} to={`/services/${l.key}`} style={lineVars(l)}
              className="poster min-h-[440px] transition-transform duration-[250ms] hover:-translate-y-[3px] hover:text-[var(--on-ink)]">
              <div className="poster-head px-[22px] pb-11 pt-5">
                <div aria-hidden="true" className="mark h-14 w-14" style={{ backgroundImage: cssUrl(l.mark) }} />
                <span className="poster-num">{l.n}</span>
              </div>
              <div className="relative z-[1] flex max-w-[70%] flex-col gap-3 px-[22px] pt-1">
                <h3 className="m-0 text-[26px] font-[550] leading-[1.08] tracking-[-0.02em]">{L(l.title)}</h3>
                <ul className="m-0 flex list-none flex-col gap-1 p-0 text-[15px] leading-[1.4]">
                  {l.cardTags.map((g, i) => <li key={i}>{L(g)}</li>)}
                </ul>
              </div>
              <div aria-hidden="true" className="poster-photo -bottom-[5%] -right-[5%] w-[66%]"
                style={{ backgroundImage: cssUrl(l.photo), WebkitMaskImage: cssUrl(l.mark), maskImage: cssUrl(l.mark) }} />
              <span className="relative z-[1] mt-auto flex items-center gap-2 px-[22px] pb-[22px] text-[15px] font-[550]">
                {t('home.services.explore')} <ArrowRight aria-hidden="true" />
              </span>
            </Link>
          ))}
          <Link to="/contact" className="on-dark flex min-h-[440px] flex-col justify-between gap-5 rounded bg-charcoal px-6 py-7 text-white hover:bg-charcoal-800 hover:text-white">
            <p className="m-0 text-[28px] font-[550] leading-[1.12] tracking-[-0.02em]">{t('home.services.unsure')}</p>
            <span className="flex items-center gap-2 font-[550] text-white">{t('home.services.unsure_cta')} <ArrowRight aria-hidden="true" className="text-tangerine" /></span>
          </Link>
        </div>
      </section>

      {/* 4. */}
      <DividerStrip />

      {/* 5. Why Sunshine */}
      <WhySunshine />

      {/* 6–7. Certifications and testimonials, "ven" ribbon behind both. */}
      <div className="relative overflow-hidden">
        <RibbonVen />
        <section aria-labelledby="cert-h" className="wrap relative z-[1]" style={{ paddingTop: 'clamp(56px,7vw,112px)' }}>
          <div className="mb-10 flex flex-wrap items-end justify-between gap-x-12 gap-y-4">
            <div className="max-w-[620px]">
              <p className="t-kicker">{t('home.certs.kicker')}</p>
              <h2 id="cert-h" className="t-h2">{L(section('home.certificates')?.title)}</h2>
            </div>
            <p className="m-0 max-w-[380px] text-[16px] leading-[1.55] text-ink-2">{L(section('home.certificates')?.body)}</p>
          </div>
          <CertificateCards />
        </section>
        <Testimonials />
      </div>

      {/* 8. Quotation band */}
      <QuoteBand />
    </>
  )
}

function WhySunshine() {
  const { b, t, L, section } = useSite()
  const before = section('home.wo_before')
  const after = section('home.wo_after')
  const dt = 'text-ink-3'
  return (
    <section aria-labelledby="why-h" className="on-dark bg-charcoal text-white">
      <div className="wrap flex flex-wrap items-start" style={{ gap: 'clamp(40px,6vw,96px)', paddingTop: 'clamp(56px,7vw,112px)', paddingBottom: 'clamp(56px,7vw,112px)' }}>
        <div className="min-w-0 flex-[1_1_420px]">
          <p className="t-kicker">{t('home.why.kicker')}</p>
          <h2 id="why-h" className="t-h2 mb-10 max-w-[560px]">{L(section('home.why')?.title)}</h2>
          <ol className="m-0 flex list-none flex-col p-0">
            {(b.lists.home_why || []).map((w, i) => (
              <li key={i} className="grid grid-cols-[44px_1fr] gap-4 border-t border-charcoal-600 py-[22px]">
                <Icon name={w.icon} className="mt-0.5 text-[28px] text-tangerine" />
                <div>
                  <h3 className="mb-1.5 mt-0 text-[20px] font-medium tracking-[-0.01em]">{L(w.title)} <PH show={!!w.placeholder} small /></h3>
                  <p className="m-0 max-w-[460px] text-[15px] leading-[1.55] text-ondark-muted">{L(w.body)}</p>
                </div>
              </li>
            ))}
          </ol>
        </div>
        <figure className="m-0 min-w-0 max-w-[480px] flex-[1_1_380px]">
          <div className="rounded-[14px] bg-white p-5 text-charcoal" style={{ boxShadow: '0 30px 60px -20px rgba(0,0,0,0.55)', transform: 'rotate(-1.2deg)' }}>
            <div className="flex items-center justify-between gap-3 border-b border-dashed border-line-3 pb-[14px]">
              <div>
                <p className="m-0 text-[12px] tracking-[0.04em] text-ink-3">{t('home.wo.label')}</p>
                <p className="m-0 mt-0.5 text-[17px] font-[550] tnum">{t('home.wo.number')}</p>
              </div>
              <span className="flex items-center gap-1.5 rounded-full bg-[#E5F2E7] px-[10px] py-1.5 text-[13px] font-medium text-success-fg">
                <CheckCircle weight="fill" aria-hidden="true" />{t('home.wo.done')}
              </span>
            </div>
            <dl className="my-[14px] grid grid-cols-[auto_1fr] gap-x-4 gap-y-2 text-[14px]">
              <dt className={dt}>{t('home.wo.site')}</dt><dd className="m-0">{t('home.wo.site_value')}</dd>
              <dt className={dt}>{t('home.wo.task')}</dt><dd className="m-0">{t('home.wo.task_value')}</dd>
              <dt className={dt}>{t('home.wo.crew')}</dt><dd className="m-0">{t('home.wo.crew_value')}</dd>
              <dt className={dt}>{t('home.wo.time')}</dt><dd className="m-0 tnum">{t('home.wo.time_value')}</dd>
            </dl>
            <div className="grid grid-cols-2 gap-2">
              {[[before, 'home.wo.before', 'home.wo.before_time'], [after, 'home.wo.after', 'home.wo.after_time']].map(([s, label, time]) => {
                const sec = s as ReturnType<typeof section>
                return (
                  <div key={label as string} className="relative overflow-hidden rounded-lg bg-line-2" style={{ aspectRatio: '4/3' }}>
                    <div role="img" aria-label={L(sec?.title)} className="h-full w-full bg-cover bg-center bg-no-repeat" style={{ backgroundImage: cssUrl(sec?.image) }} />
                    <span className="absolute left-2 top-2 rounded bg-[rgba(31,33,28,0.86)] px-[7px] py-[3px] text-[11px] text-white">{t(label as string)} · {t(time as string)}</span>
                  </div>
                )
              })}
            </div>
            <div className="mt-4 flex items-end justify-between gap-3">
              <div>
                <p className="m-0 text-[12px] text-ink-3">{t('home.wo.signed')}</p>
                <p className="m-0 mt-1 border-b border-charcoal pb-0.5 pr-[18px] text-[22px] font-light italic tracking-[-0.02em]">{t('home.wo.signer')}</p>
              </div>
              <p className="m-0 text-right text-[12px] text-ink-3 tnum">{t('home.wo.role')}<br />{t('home.wo.signoff_meta')}</p>
            </div>
          </div>
          <figcaption className="mt-[22px] max-w-[400px] text-[14px] leading-[1.5] text-ondark-muted">{t('home.wo.caption')}</figcaption>
        </figure>
      </div>
    </section>
  )
}

function Testimonials() {
  const { b, t, L, section } = useSite()
  // BR-2.4: until a real quote is published, the dashed placeholder cards show.
  const real = b.testimonials.length > 0
  const quotes = real ? b.testimonials : b.testimonialPlaceholders
  if (quotes.length === 0) return null
  const anyPH = quotes.some((q) => q.placeholder)
  return (
    <section aria-labelledby="tst-h" className="wrap relative z-[1]" style={{ paddingTop: 'clamp(56px,7vw,112px)', paddingBottom: 'clamp(56px,7vw,112px)' }}>
      <div className="mb-10 flex flex-wrap items-center gap-x-5 gap-y-3">
        <h2 id="tst-h" className="t-h2">{L(section('home.testimonials')?.title)}</h2>
        <PH show={anyPH} label={t('home.testimonials.placeholder')} />
      </div>
      <div className="grid gap-3" style={{ gridTemplateColumns: 'repeat(auto-fit,minmax(280px,1fr))' }}>
        {quotes.map((q, i) => (
          <figure key={i} className={`m-0 flex flex-col gap-6 rounded-xl bg-white p-7 ${q.placeholder ? 'border border-dashed border-line-4' : 'border border-line-2'}`}>
            <Quotes weight="fill" aria-hidden="true" className="text-[30px] text-tangerine" />
            <blockquote className="m-0 text-[19px] leading-[1.45] tracking-[-0.01em] text-charcoal">{L(q.quote)}</blockquote>
            <figcaption className="mt-auto text-[14px] leading-[1.45] text-ink-3">
              <strong className="block font-[550] text-charcoal">{q.name}</strong>{[q.role, q.company].filter(Boolean).join(', ')}
            </figcaption>
          </figure>
        ))}
      </div>
    </section>
  )
}

/** The short form on Home links on to the full Contact form, carrying what
 * was typed (README → Home §8). Nothing is stored from here. */
function QuoteBand() {
  const { b, t, L, section, param, paramPH } = useSite()
  const nav = useNavigate()
  const [name, setName] = useState('')
  const [company, setCompany] = useState('')
  const [type, setType] = useState('')
  const [svc, setSvc] = useState<string[]>([])
  const sec = section('home.quote')
  const submit = (e: FormEvent) => {
    e.preventDefault()
    nav('/contact', { state: { name, company, buildingType: type, services: svc } })
  }
  const input = 'h-12 rounded-lg border border-line-4 bg-white px-[14px] text-[16px] text-charcoal focus:border-charcoal'
  return (
    <section id="quote" aria-labelledby="quote-h" className="on-dark bg-charcoal text-white">
      <div className="wrap flex flex-wrap" style={{ gap: '48px clamp(40px,6vw,96px)', paddingTop: 'clamp(56px,7vw,104px)', paddingBottom: 'clamp(56px,7vw,104px)' }}>
        <div className="min-w-0 max-w-[480px] flex-[1_1_360px]">
          <h2 id="quote-h" className="m-0 font-[550] leading-none tracking-[-0.035em]" style={{ fontSize: 'clamp(34px,4.4vw,60px)', textWrap: 'balance' } as CSSProperties}>{L(sec?.title)}</h2>
          <p className="mb-0 mt-5 text-[17px] leading-[1.55] text-ondark-muted">{L(sec?.body)}</p>
          <div className="mt-9 flex flex-col gap-4 text-[16px]">
            <a href={waHref(param('contact.whatsapp_digits'))} target="_blank" rel="noopener noreferrer" className="flex items-center gap-3 text-white hover:text-tangerine">
              <span className="grid h-11 w-11 flex-none place-items-center rounded-full bg-wa text-[22px] text-white"><WhatsappLogo weight="fill" aria-hidden="true" /></span>
              <span>{t('wa.chat')} · {param('contact.whatsapp')} <PH show={paramPH('contact.whatsapp')} small /></span>
            </a>
            <a href={telHref(param('contact.phone'))} className="flex items-center gap-3 text-white hover:text-tangerine">
              <span className="grid h-11 w-11 flex-none place-items-center rounded-full border border-charcoal-400 text-[20px]"><Phone aria-hidden="true" /></span>
              <span>{t('home.quote.call')} · {param('contact.phone')} <PH show={paramPH('contact.phone')} small /></span>
            </a>
          </div>
        </div>
        <form onSubmit={submit} className="grid min-w-0 flex-[1.3_1_420px] gap-4 rounded-[14px] bg-white text-charcoal" style={{ padding: 'clamp(20px,3vw,32px)', gridTemplateColumns: 'repeat(auto-fit,minmax(200px,1fr))' }}>
          <label className="field">{t('form.name')}<input className={input} value={name} onChange={(e) => setName(e.target.value)} autoComplete="name" maxLength={b.rules.limits.name} /></label>
          <label className="field">{t('form.company')}<input className={input} value={company} onChange={(e) => setCompany(e.target.value)} autoComplete="organization" maxLength={b.rules.limits.company} /></label>
          <label className="field col-span-full">{t('form.building_type')}
            <select className={`${input} px-3`} value={type} onChange={(e) => setType(e.target.value)}>
              <option value="">{t('form.choose_type')}</option>
              {b.buildingTypes.map((o) => <option key={o.key} value={o.key}>{L(o.label)}</option>)}
            </select>
          </label>
          <fieldset className="col-span-full m-0 border-0 p-0">
            <legend className="mb-2 text-[14px] font-medium">{t('form.services')}</legend>
            <div className="flex flex-wrap gap-2">
              {b.lines.map((l) => {
                const on = svc.includes(l.key)
                return (
                  <label key={l.key} className={`flex cursor-pointer items-center gap-2 rounded-full border px-[14px] py-[9px] text-[14px] hover:border-charcoal ${on ? 'border-charcoal bg-cotton' : 'border-line-4 bg-white'}`}>
                    <input type="checkbox" checked={on} onChange={() => setSvc((s) => (on ? s.filter((x) => x !== l.key) : [...s, l.key]))} className="m-0 h-4 w-4 accent-charcoal" />
                    {L(l.short)}
                  </label>
                )
              })}
            </div>
          </fieldset>
          <div className="col-span-full mt-1 flex flex-wrap items-center gap-3">
            <button type="submit" className="btn btn-red rounded-lg py-[15px]">{t('cta.quote')} <ArrowRight aria-hidden="true" /></button>
            <p className="m-0 flex-[1_1_200px] text-[13px] text-ink-3">{t('home.quote.note')}</p>
          </div>
        </form>
      </div>
    </section>
  )
}
