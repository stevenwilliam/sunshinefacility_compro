import type { CSSProperties } from 'react'
import { Link } from 'react-router-dom'
import { ArrowRight } from '@/lib/phosphor'
import { useSite } from '@/lib/site'
import { usePageMeta } from '@/lib/meta'
import { Icon } from '@/lib/icons'
import { Breadcrumb, cssUrl } from '@/components/bits'

export default function Industries() {
  const { b, t, L, section } = useSite()
  usePageMeta(t('seo.industries.title') + t('seo.suffix'), t('seo.industries.description'))
  const hero = section('industries.hero')
  const cta = section('industries.cta')
  return (
    <>
      <section className="on-dark bg-charcoal text-white">
        <div className="wrap flex flex-wrap items-end" style={{ gap: '32px clamp(32px,6vw,96px)', paddingTop: 'clamp(40px,6vw,88px)', paddingBottom: 'clamp(40px,6vw,88px)' }}>
          <div className="min-w-0 flex-[1.4_1_420px]">
            <div className="mb-7"><Breadcrumb items={[{ label: t('nav.home'), to: '/' }, { label: t('industries.kicker') }]} /></div>
            <p className="eyebrow">{t('industries.kicker')}</p>
            <h1 className="t-page">{L(hero?.title)}</h1>
          </div>
          <p className="t-lead m-0 min-w-0 max-w-[460px] flex-[1_1_320px] text-ondark-body">{L(hero?.body)}</p>
        </div>
      </section>

      <section aria-label={t('industries.kicker')} className="wrap" style={{ paddingTop: 'clamp(24px,4vw,48px)', paddingBottom: 'clamp(56px,7vw,104px)' }}>
        {b.industries.map((ind) => (
          <article key={ind.key} id={ind.key} className="flex scroll-mt-[88px] flex-wrap border-b border-line-2"
            style={{ gap: '24px clamp(32px,5vw,72px)', paddingTop: 'clamp(28px,4vw,48px)', paddingBottom: 'clamp(28px,4vw,48px)' }}>
            <div className="flex min-w-0 flex-[1.2_1_360px] flex-col gap-[14px]">
              <div className="flex items-center gap-[14px]">
                <span className="grid h-[52px] w-[52px] flex-none place-items-center rounded bg-cotton text-[28px]"><Icon name={ind.icon} /></span>
                <h2 className="m-0 font-[550] leading-[1.05] tracking-[-0.025em]" style={{ fontSize: 'clamp(26px,2.8vw,36px)' }}>{L(ind.title)}</h2>
              </div>
              <p className="m-0 max-w-[560px] text-[17px] leading-[1.55] text-ink-2">{L(ind.body)}</p>
              <p className="t-caps mt-1.5">{t('industries.lines')}</p>
              <div className="flex flex-wrap gap-2">
                {ind.lines.map((k) => {
                  const l = b.lines.find((x) => x.key === k)
                  if (!l) return null
                  return (
                    <Link key={k} to={`/services/${k}`} className="flex min-h-[40px] items-center gap-2 rounded-full border border-line-3 px-[14px] text-[14px] font-medium hover:border-charcoal">
                      <span className="sq" style={{ '--sq': l.base } as CSSProperties} />{L(l.label)}
                    </Link>
                  )
                })}
              </div>
            </div>
            <div className="flex min-w-0 flex-[1_1_300px] flex-col justify-center gap-3">
              {ind.clients.length > 0 ? (
                <>
                  <p className="t-caps">{t('industries.clients')}</p>
                  <ul className="m-0 grid list-none gap-2 p-0" style={{ gridTemplateColumns: 'repeat(auto-fill,minmax(120px,1fr))' }}>
                    {ind.clients.map((ci) => {
                      const c = b.clients[ci]
                      return (
                        <li key={ci} className="box-border h-[72px] rounded border border-line-1 px-[14px] py-3">
                          <div role="img" aria-label={c.name} className="h-full w-full bg-contain bg-center bg-no-repeat" style={{ backgroundImage: cssUrl(c.logo) }} />
                        </li>
                      )
                    })}
                  </ul>
                </>
              ) : (
                <Link to="/contact" className="flex items-center gap-[10px] rounded border border-dashed border-line-4 p-5 text-[15px] text-ink-2 hover:border-charcoal">
                  {t('industries.no_clients')} <ArrowRight aria-hidden="true" />
                </Link>
              )}
            </div>
          </article>
        ))}
      </section>

      <section className="border-t border-line-2 bg-cotton">
        <div className="wrap flex flex-wrap items-end justify-between gap-x-12 gap-y-6" style={{ paddingTop: 'clamp(56px,7vw,104px)', paddingBottom: 'clamp(56px,7vw,104px)' }}>
          <div className="max-w-[640px]">
            <h2 className="t-cta">{L(cta?.title)}</h2>
            <p className="mb-0 mt-[18px] text-[17px] leading-[1.55] text-ink-2">{L(cta?.body)}</p>
          </div>
          <Link to="/contact" className="btn btn-red">{t('cta.quote')} <ArrowRight aria-hidden="true" /></Link>
        </div>
      </section>
    </>
  )
}
