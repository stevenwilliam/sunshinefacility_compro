import type { CSSProperties } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ArrowRight, WhatsappLogo } from '@/lib/phosphor'
import { lineVars, useSite } from '@/lib/site'
import { usePageMeta } from '@/lib/meta'
import { Icon } from '@/lib/icons'
import { Breadcrumb, cssUrl, Picture, Stats } from '@/components/bits'
import { waHref } from '@/components/Chrome'
import { CertificateBadges } from '@/components/Certificates'
import NotFound from './NotFound'

export default function ServiceDetail() {
  const { line } = useParams()
  const { b, t, L, lang, section, param } = useSite()
  const l = b.lines.find((x) => x.key === line)
  usePageMeta(l ? L(l.title) + t('seo.suffix') : '', l ? L(l.lead) : undefined)
  if (!l) return <NotFound />
  const total = String(b.lines.length).padStart(2, '0')
  const industries = l.industries.map((k) => b.industries.find((i) => i.key === k)).filter(Boolean)
  const wa = waHref(param('contact.whatsapp_digits'))
  const map = section('service_detail.map')
  const proof = section('service_detail.proof')
  const cta = section('service_detail.cta')
  const pad = 'clamp(56px,7vw,96px)'
  return (
    <div style={lineVars(l)}>
      <section className="wrap pt-6">
        <Breadcrumb items={[{ label: t('nav.home'), to: '/' }, { label: t('nav.services'), to: '/services' }, { label: L(l.title) }]} />
      </section>

      <section aria-labelledby="svc-h" className="wrap pt-5">
        <div className="poster" style={{ minHeight: 'clamp(520px,60vw,620px)' }}>
          <div className="poster-head h45 gap-4" style={{ padding: 'clamp(20px,3vw,32px) clamp(20px,3vw,40px) clamp(56px,7vw,96px)' }}>
            <div aria-hidden="true" className="mark" style={{ width: 'clamp(64px,7vw,96px)', aspectRatio: '1/1', backgroundImage: cssUrl(l.mark) }} />
            <span className="text-[14px] text-ink-3 tnum">{l.n} / {total}</span>
          </div>
          <div className="relative z-[1] flex max-w-[620px] flex-1 flex-col gap-[18px]" style={{ padding: '0 clamp(20px,3vw,40px) clamp(28px,4vw,48px)' }}>
            <h1 id="svc-h" className="t-page" style={{ lineHeight: 0.98 }}>{L(l.title)}</h1>
            <p className="t-lead m-0 max-w-[520px]">{L(l.lead)}</p>
            <div className="mt-auto flex flex-wrap gap-3 pt-3">
              <Link to={`/contact?service=${l.key}`} className="btn btn-white">{t('cta.quote')} <ArrowRight aria-hidden="true" /></Link>
              <a href={wa} target="_blank" rel="noopener noreferrer" className="btn border-[1.5px] border-[var(--on-ink)] px-[22px] py-[15px] font-medium text-[var(--on-ink)] hover:bg-white/[.12] hover:text-[var(--on-ink)]">
                <WhatsappLogo weight="fill" aria-hidden="true" />{t('common.whatsapp')}
              </a>
            </div>
          </div>
          <div aria-hidden="true" className="poster-photo -bottom-[6%] -right-[4%]"
            style={{ width: 'min(54%,560px)', backgroundImage: cssUrl(l.photo), WebkitMaskImage: cssUrl(l.mark), maskImage: cssUrl(l.mark) }} />
        </div>
      </section>

      <section aria-label={t('detail.other')} className="wrap pt-5">
        <div className="flex flex-wrap items-center gap-2">
          <span className="mr-1 text-[14px] text-ink-3">{t('detail.other')}</span>
          {b.lines.map((x) => {
            const on = x.key === l.key
            return (
              <Link key={x.key} to={`/services/${x.key}`} aria-current={on ? 'page' : undefined}
                className={`flex min-h-[44px] items-center gap-2 rounded-full border px-[14px] text-[14px] font-medium hover:border-charcoal ${on ? 'border-charcoal bg-charcoal text-white' : 'border-line-3 bg-white text-charcoal'}`}>
                <span className="sq" style={{ '--sq': x.base } as CSSProperties} />{L(x.title)}
              </Link>
            )
          })}
        </div>
      </section>

      {l.facts.length > 0 && (
        <section className="wrap" style={{ paddingTop: 'clamp(40px,5vw,64px)' }}>
          <div className="[&_dl]:!grid-cols-[repeat(auto-fit,minmax(170px,1fr))] [&_.t-stat]:!text-[clamp(32px,3.4vw,46px)]">
            <Stats items={l.facts} rule={l.base} />
          </div>
        </section>
      )}

      <section aria-labelledby="inc-h" className="wrap" style={{ paddingTop: pad }}>
        <h2 id="inc-h" className="t-h2s mb-8">{t('detail.included')}</h2>
        <div className="grid gap-3" style={{ gridTemplateColumns: 'repeat(auto-fill,minmax(min(100%,280px),1fr))' }}>
          {l.groups.map((g, i) => {
            const list = g.items[lang === 'id' ? 1 : 0].length ? g.items[lang === 'id' ? 1 : 0] : g.items[0]
            return (
              <div key={i} className="flex flex-col gap-[14px] rounded border border-line-2 bg-white p-6">
                <div className="flex items-center justify-between gap-3">
                  <h3 className="m-0 text-[20px] font-[550] leading-[1.2] tracking-[-0.01em]">{L(g.title)}</h3>
                  <span className="text-[13px] text-ink-3 tnum">{String(i + 1).padStart(2, '0')}</span>
                </div>
                {L(g.body) && <p className="m-0 text-[15px] leading-[1.55] text-ink-2">{L(g.body)}</p>}
                {list.length > 0 && (
                  <ul className="m-0 flex list-none flex-col p-0">
                    {list.map((it, j) => (
                      <li key={j} className="flex items-center gap-[10px] border-t border-line-1 py-[9px] text-[15px]"><span className="sq" style={{ '--sq': l.base } as CSSProperties} />{it}</li>
                    ))}
                  </ul>
                )}
              </div>
            )
          })}
        </div>
      </section>

      {l.showMap && l.map && map && (
        <section className="wrap" style={{ paddingTop: pad }}>
          <h2 className="t-h2s mb-3">{L(map.title)}</h2>
          <p className="mb-7 mt-0 max-w-[620px] text-[16px] leading-[1.55] text-ink-2">{L(map.body)}</p>
          <Picture img={l.map} sizes="(min-width: 1280px) 1184px, 100vw" className="block h-auto w-full rounded border border-line-2 bg-white" />
        </section>
      )}

      {l.showCerts && b.certificates.length > 0 && (
        <section className="wrap" style={{ paddingTop: pad }}>
          <h2 className="t-h2s mb-7">{L(section('service_detail.certificates')?.title)}</h2>
          <CertificateBadges />
        </section>
      )}

      {proof && (
        <section className="wrap" style={{ paddingTop: pad }}>
          <div className="on-dark flex flex-wrap rounded bg-charcoal text-white" style={{ padding: 'clamp(28px,4vw,48px)', gap: '32px clamp(32px,5vw,72px)' }}>
            <div className="min-w-0 flex-[1_1_320px]">
              <h2 className="m-0 font-[550] leading-[1.08] tracking-[-0.025em]" style={{ fontSize: 'clamp(26px,3vw,38px)', textWrap: 'balance' } as CSSProperties}>{L(proof.title)}</h2>
              <p className="mb-0 mt-4 max-w-[460px] text-[16px] leading-[1.55] text-ondark-body">{L(proof.body)}</p>
            </div>
            <ul className="m-0 flex min-w-0 flex-[1_1_320px] list-none flex-col p-0">
              {(b.lists.service_proof || []).map((p, i) => (
                <li key={i} className="grid grid-cols-[32px_1fr] gap-3 border-t border-charcoal-600 py-[14px] text-[16px] leading-[1.45]">
                  <Icon name={p.icon} className="text-[24px] text-tangerine" />{L(p.title)}
                </li>
              ))}
            </ul>
          </div>
        </section>
      )}

      {industries.length > 0 && (
        <section className="wrap" style={{ paddingTop: pad }}>
          <h2 className="mb-5 mt-0 font-[550] leading-[1.1] tracking-[-0.02em]" style={{ fontSize: 'clamp(24px,2.6vw,32px)' }}>{L(section('service_detail.industries')?.title)}</h2>
          <div className="flex flex-wrap gap-2">
            {industries.map((i) => (
              <Link key={i!.key} to={`/industries#${i!.key}`} className="flex min-h-[44px] items-center gap-2 rounded-full border border-line-3 px-4 text-[15px] hover:border-charcoal">
                <Icon name={i!.icon} className="text-[18px]" />{L(i!.title)}
              </Link>
            ))}
          </div>
        </section>
      )}

      <section className="wrap flex flex-wrap items-end justify-between gap-x-12 gap-y-6" style={{ paddingTop: 'clamp(56px,7vw,104px)', paddingBottom: 'clamp(56px,7vw,104px)' }}>
        <div className="max-w-[640px]">
          <h2 className="t-cta">{L(cta?.title)}</h2>
          <p className="mb-0 mt-[18px] text-[17px] leading-[1.55] text-ink-2">{L(cta?.body)}</p>
        </div>
        <Link to={`/contact?service=${l.key}`} className="btn btn-red">{t('cta.quote')} <ArrowRight aria-hidden="true" /></Link>
      </section>
    </div>
  )
}
