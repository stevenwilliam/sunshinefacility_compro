import type { CSSProperties } from 'react'
import { Link } from 'react-router-dom'
import { ArrowRight, WhatsappLogo } from '@/lib/phosphor'
import { lineVars, useSite } from '@/lib/site'
import { usePageMeta } from '@/lib/meta'
import { Icon } from '@/lib/icons'
import { Breadcrumb, cssUrl } from '@/components/bits'
import { waHref } from '@/components/Chrome'

export default function Services() {
  const { b, t, L, section, param } = useSite()
  usePageMeta(t('seo.services.title') + t('seo.suffix'), t('seo.services.description'))
  const hero = section('services.hero')
  const total = String(b.lines.length).padStart(2, '0')
  return (
    <>
      <section className="on-dark bg-charcoal text-white">
        <div className="wrap" style={{ paddingTop: 'clamp(40px,6vw,88px)', paddingBottom: 'clamp(40px,6vw,88px)' }}>
          <div className="mb-7"><Breadcrumb items={[{ label: t('nav.home'), to: '/' }, { label: t('services.kicker') }]} /></div>
          <p className="eyebrow">{t('services.kicker')}</p>
          <h1 className="t-page max-w-[900px]">{L(hero?.title)}</h1>
          <p className="t-lead mb-0 mt-6 max-w-[640px] text-ondark-body">{L(hero?.body)}</p>
        </div>
      </section>

      <section aria-label={t('services.kicker')} className="wrap" style={{ paddingTop: 'clamp(24px,4vw,56px)', paddingBottom: 'clamp(56px,7vw,104px)' }}>
        {b.lines.map((l) => (
          <article key={l.key} id={l.key} style={{ ...lineVars(l), gap: '28px clamp(32px,5vw,72px)', paddingTop: 'clamp(32px,4vw,56px)', paddingBottom: 'clamp(32px,4vw,56px)' }}
            className="flex flex-wrap items-stretch border-b border-line-2">
            <Link to={`/services/${l.key}`} aria-label={L(l.title)} className="poster min-h-[380px] min-w-0 max-w-[520px] flex-[1_1_320px] hover:text-[var(--on-ink)]">
              <div className="poster-head px-6 pb-12 pt-[22px]">
                <div aria-hidden="true" className="mark h-[60px] w-[60px]" style={{ backgroundImage: cssUrl(l.mark) }} />
                <span className="poster-num">{l.n}</span>
              </div>
              <p className="relative z-[1] mx-6 mb-0 mt-1 max-w-[60%] font-[550] leading-[1.05] tracking-[-0.02em]" style={{ fontSize: 'clamp(26px,2.6vw,34px)' }}>{L(l.title)}</p>
              <div aria-hidden="true" className="poster-photo -bottom-[6%] -right-[6%] w-[68%]"
                style={{ backgroundImage: cssUrl(l.photo), WebkitMaskImage: cssUrl(l.mark), maskImage: cssUrl(l.mark) }} />
            </Link>
            <div className="flex min-w-0 flex-[1.3_1_380px] flex-col justify-center gap-[18px]">
              <p className="m-0 text-[13px] font-semibold tracking-[0.06em] text-[var(--label)] tnum">{t('services.count', { n: l.n, m: total })}</p>
              <h2 className="t-h2s">{L(l.title)}</h2>
              <p className="m-0 max-w-[560px] text-[17px] leading-[1.55] text-ink-2" style={{ textWrap: 'pretty' } as CSSProperties}>{L(l.description)}</p>
              <ul className="m-0 mt-1.5 grid list-none gap-x-6 p-0" style={{ gridTemplateColumns: 'repeat(auto-fill,minmax(220px,1fr))' }}>
                {l.serviceTags.map((g, i) => (
                  <li key={i} className="flex items-center gap-[10px] border-t border-line-1 py-[11px] text-[15px]">
                    <span className="sq" style={{ '--sq': l.base } as CSSProperties} />{L(g)}
                  </li>
                ))}
              </ul>
              <Link to={`/services/${l.key}`} className="mt-2 flex items-center gap-[10px] self-start border-b-2 border-[var(--base)] py-[14px] text-[16px] font-[550] hover:text-red">
                {t('services.details')} <ArrowRight aria-hidden="true" />
              </Link>
            </div>
          </article>
        ))}
      </section>

      <section className="border-y border-line-2 bg-cotton">
        <div className="wrap" style={{ paddingTop: 'clamp(48px,6vw,88px)', paddingBottom: 'clamp(48px,6vw,88px)' }}>
          <h2 className="t-h2s mb-9 max-w-[620px]">{L(section('services.benefits')?.title)}</h2>
          <div className="grid gap-8" style={{ gridTemplateColumns: 'repeat(auto-fit,minmax(240px,1fr))' }}>
            {(b.lists.services_benefits || []).map((o, i) => (
              <div key={i} className="flex flex-col gap-[10px] border-t-2 border-charcoal pt-5">
                <Icon name={o.icon} className="text-[30px] text-red" />
                <h3 className="m-0 mt-1.5 text-[21px] font-[550] tracking-[-0.01em]">{L(o.title)}</h3>
                <p className="m-0 text-[15px] leading-[1.55] text-ink-2">{L(o.body)}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="wrap flex flex-wrap items-end justify-between gap-x-12 gap-y-6" style={{ paddingTop: 'clamp(56px,7vw,104px)', paddingBottom: 'clamp(56px,7vw,104px)' }}>
        <div className="max-w-[640px]">
          <h2 className="t-cta">{L(section('services.cta')?.title)}</h2>
          <p className="mb-0 mt-[18px] text-[17px] leading-[1.55] text-ink-2">{L(section('services.cta')?.body)}</p>
        </div>
        <div className="flex flex-wrap gap-3">
          <Link to="/contact" className="btn btn-red">{t('cta.quote')} <ArrowRight aria-hidden="true" /></Link>
          <a href={waHref(param('contact.whatsapp_digits'))} target="_blank" rel="noopener noreferrer" className="btn btn-line-dark">
            <WhatsappLogo weight="fill" aria-hidden="true" className="text-wa" />{t('common.whatsapp')}
          </a>
        </div>
      </section>
    </>
  )
}
