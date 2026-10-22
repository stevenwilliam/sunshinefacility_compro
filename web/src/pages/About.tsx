import type { CSSProperties } from 'react'
import { Link } from 'react-router-dom'
import { ArrowRight, User } from '@/lib/phosphor'
import { lineVars, useSite } from '@/lib/site'
import { usePageMeta } from '@/lib/meta'
import { Icon } from '@/lib/icons'
import { Breadcrumb, cssUrl, Paras, PH, Picture, Stats } from '@/components/bits'
import { CertificateTiles } from '@/components/Certificates'

export default function About() {
  const { b, t, L, section } = useSite()
  usePageMeta(t('seo.about.title') + t('seo.suffix'), t('seo.about.description'))
  const hero = section('about.hero')
  const poster = section('about.poster')
  const story = section('about.story')
  const vm = section('about.vision_mission')
  const vision = section('about.vision')
  const eco = section('about.ecosystem')
  const food = b.lines.find((l) => l.key === 'food')
  const mission = b.lists.about_mission || []
  const vmDraft = !!(vm?.placeholder || vision?.placeholder || mission.some((m) => m.placeholder))
  const mgmtPH = b.management.some((m) => m.placeholder)
  const pad = { paddingTop: 'clamp(56px,7vw,104px)', paddingBottom: 'clamp(56px,7vw,104px)' }
  return (
    <>
      <section className="wrap flex flex-wrap items-center" style={{ gap: '32px clamp(32px,6vw,88px)', paddingTop: 'clamp(32px,5vw,72px)', paddingBottom: 'clamp(32px,5vw,72px)' }}>
        <div className="min-w-0 flex-[1.3_1_420px]">
          <div className="mb-7"><Breadcrumb items={[{ label: t('nav.home'), to: '/' }, { label: t('about.kicker') }]} /></div>
          <p className="eyebrow">{t('about.kicker')}</p>
          <h1 className="t-page">{L(hero?.title)}</h1>
          <p className="t-lead mb-0 mt-6 max-w-[560px] text-ink-2">{L(hero?.body)}</p>
        </div>
        {poster && food && (
          <div className="poster relative min-w-0 max-w-[460px] flex-[1_1_320px]" style={{ ...lineVars({ ...food, ink: food.ink, onInk: '#FFFFFF' }), aspectRatio: '4/5' }}>
            <div className="bg-white px-6 pb-[52px] pt-[22px]" style={{ clipPath: 'polygon(0 0,100% 0,100% 55%,0 100%)' }}>
              <div aria-hidden="true" className="mark h-[60px] w-[60px]" style={{ backgroundImage: cssUrl(food.mark) }} />
            </div>
            <p className="relative z-[1] mx-6 mb-0 mt-1 font-semibold leading-[.95] tracking-[-0.035em]" style={{ fontSize: 'clamp(40px,4.4vw,60px)' }}>{L(poster.title)}</p>
            <p className="relative z-[1] mx-6 mb-0 mt-[10px] max-w-[55%] text-[16px] leading-[1.45]">{L(poster.body)}</p>
            <div aria-hidden="true" className="poster-photo -bottom-[6%] -right-[6%] w-[72%]"
              style={{ backgroundImage: cssUrl(poster.image), WebkitMaskImage: cssUrl(food.mark), maskImage: cssUrl(food.mark) }} />
          </div>
        )}
      </section>

      <section className="border-y border-line-2">
        <div className="wrap" style={{ paddingTop: 'clamp(28px,4vw,48px)', paddingBottom: 'clamp(28px,4vw,48px)' }}>
          <Stats items={b.stats.about || []} />
        </div>
      </section>

      <section aria-labelledby="story-h" className="wrap flex flex-wrap" style={{ ...pad, gap: '32px clamp(32px,6vw,96px)' }}>
        <div className="max-w-[420px] flex-[1_1_300px]">
          <p className="t-kicker">{t('about.story.kicker')}</p>
          <h2 id="story-h" className="t-h2">{L(story?.title)}</h2>
        </div>
        <div className="flex min-w-0 max-w-[680px] flex-[1.5_1_420px] flex-col gap-[18px] text-[18px] leading-[1.6] text-ink-2">
          <Paras text={L(story?.body)} className="m-0" />
        </div>
      </section>

      <section aria-labelledby="jr-h" className="on-dark bg-charcoal text-white">
        <div className="wrap" style={pad}>
          <p className="t-kicker">{t('about.journey.kicker')}</p>
          <h2 id="jr-h" className="t-h2 mb-12 max-w-[720px]">{L(section('about.journey')?.title)}</h2>
          <ol className="m-0 grid list-none gap-x-6 gap-y-8 p-0" style={{ gridTemplateColumns: 'repeat(auto-fit,minmax(210px,1fr))' }}>
            {(b.lists.about_journey || []).map((j, i) => (
              <li key={i} className="flex flex-col gap-[10px] pt-[18px]" style={{ borderTop: `3px solid ${j.color || '#EA2915'}` }}>
                <span className="font-[550] leading-none tracking-[-0.03em] tnum" style={{ fontSize: 'clamp(28px,2.6vw,36px)' }}>{L(j.meta)}</span>
                <span className="text-[13px] font-semibold uppercase tracking-[0.06em] text-ondark-muted">{L(j.title)}</span>
                <p className="m-0 text-[15px] leading-[1.55] text-ondark-body">{L(j.body)}</p>
              </li>
            ))}
          </ol>
        </div>
      </section>

      <section aria-labelledby="vm-h" className="wrap" style={pad}>
        {vm && (
          <>
            <div className="mb-9 flex flex-wrap items-center gap-x-5 gap-y-3">
              <h2 id="vm-h" className="t-h2">{L(vm.title)}</h2>
              <PH show={vmDraft} label={t('about.vm.draft')} />
            </div>
            <div className="grid gap-3" style={{ gridTemplateColumns: 'repeat(auto-fit,minmax(min(100%,320px),1fr))' }}>
              <div className="flex flex-col gap-[14px] rounded border border-line-2 p-7">
                <p className="t-caps">{t('about.vm.vision')}</p>
                <p className="m-0 font-medium leading-[1.25] tracking-[-0.015em]" style={{ fontSize: 'clamp(22px,2.2vw,28px)' }}>{L(vision?.body)}</p>
              </div>
              <div className="flex flex-col gap-[14px] rounded border border-line-2 p-7">
                <p className="t-caps">{t('about.vm.mission')}</p>
                <ol className="m-0 flex list-none flex-col p-0">
                  {mission.map((m, i) => (
                    <li key={i} className="grid grid-cols-[32px_1fr] gap-2 border-t border-line-1 py-3 text-[16px] leading-[1.5]">
                      <span className="font-semibold text-red tnum">{String(i + 1).padStart(2, '0')}</span>{L(m.body)}
                    </li>
                  ))}
                </ol>
              </div>
            </div>
          </>
        )}
        <h3 className="mb-6 mt-14 font-[550] tracking-[-0.02em]" style={{ fontSize: 'clamp(24px,2.6vw,32px)' }}>{L(section('about.values')?.title)}</h3>
        <div className="grid gap-8" style={{ gridTemplateColumns: 'repeat(auto-fit,minmax(240px,1fr))' }}>
          {(b.lists.about_values || []).map((v, i) => (
            <div key={i} className="flex flex-col gap-[10px] border-t-2 border-charcoal pt-[18px]">
              <Icon name={v.icon} className="text-[30px] text-red" />
              <h4 className="m-0 mt-1.5 text-[21px] font-[550] tracking-[-0.01em]">{L(v.title)}</h4>
              <p className="m-0 text-[15px] leading-[1.55] text-ink-2">{L(v.body)}</p>
            </div>
          ))}
        </div>
      </section>

      <section aria-labelledby="eco-h" className="border-y border-line-2 bg-cotton">
        <div className="wrap" style={pad}>
          <div className="mb-9 flex flex-wrap items-end justify-between gap-x-12 gap-y-4">
            <h2 id="eco-h" className="t-h2 max-w-[640px]">{L(eco?.title)}</h2>
            <p className="m-0 max-w-[380px] text-[16px] leading-[1.55] text-ink-2">{L(eco?.body)}</p>
          </div>
          <ul className="m-0 grid list-none gap-3 p-0" style={{ gridTemplateColumns: 'repeat(auto-fill,minmax(min(100%,260px),1fr))' }}>
            {(b.lists.about_ecosystem || []).map((e, i) => (
              <li key={i} className="flex flex-col gap-2 rounded border border-line-2 bg-white p-[22px]">
                <h3 className="m-0 text-[19px] font-[550]">{L(e.title)}</h3>
                <p className="m-0 text-[15px] leading-[1.5] text-ink-2">{L(e.body)}</p>
              </li>
            ))}
          </ul>
        </div>
      </section>

      {b.certificates.length > 0 && (
        <section aria-labelledby="cert-h" className="wrap" style={pad}>
          <h2 id="cert-h" className="t-h2 mb-8">{L(section('about.certificates')?.title)}</h2>
          <CertificateTiles />
        </section>
      )}

      {b.management.length > 0 && (
        <section aria-labelledby="mg-h" className="wrap" style={{ paddingBottom: 'clamp(56px,7vw,104px)', paddingTop: b.certificates.length ? 0 : 'clamp(56px,7vw,104px)' }}>
          <div className="mb-8 flex flex-wrap items-center gap-x-5 gap-y-3">
            <h2 id="mg-h" className="t-h2">{L(section('about.management')?.title)}</h2>
            <PH show={mgmtPH} label={t('about.mgmt.placeholder')} />
          </div>
          <ul className="m-0 grid list-none gap-4 p-0" style={{ gridTemplateColumns: 'repeat(auto-fill,minmax(min(100%,230px),1fr))' }}>
            {b.management.map((m, i) => (
              <li key={i} className="flex flex-col gap-3">
                {m.photo ? (
                  <Picture img={m.photo} alt={m.name} className="block w-full rounded object-cover" style={{ aspectRatio: '4/5' } as CSSProperties} />
                ) : (
                  <div className="grid place-items-center rounded border border-dashed border-line-4 bg-cotton text-ink-3" style={{ aspectRatio: '4/5' }}>
                    <div className="flex flex-col items-center gap-2 text-[14px]"><User aria-hidden="true" className="text-[40px]" />{t('about.mgmt.photo')}</div>
                  </div>
                )}
                <div>
                  <p className="m-0 text-[18px] font-[550]">{m.name}</p>
                  <p className="m-0 mt-1 text-[15px] text-ink-2">{L(m.role)}</p>
                </div>
              </li>
            ))}
          </ul>
        </section>
      )}

      <section className="on-dark bg-charcoal text-white">
        <div className="wrap flex flex-wrap items-end justify-between gap-x-12 gap-y-6" style={{ paddingTop: 'clamp(48px,6vw,88px)', paddingBottom: 'clamp(48px,6vw,88px)' }}>
          <h2 className="t-cta max-w-[620px]">{L(section('about.cta')?.title)}</h2>
          <div className="flex flex-wrap gap-3">
            <Link to="/contact" className="btn btn-red">{t('cta.quote')} <ArrowRight aria-hidden="true" /></Link>
            <Link to="/careers" className="btn btn-line-light">{t('about.cta.careers')}</Link>
          </div>
        </div>
      </section>
    </>
  )
}
