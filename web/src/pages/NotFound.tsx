import { Link } from 'react-router-dom'
import { ArrowRight } from '@/lib/phosphor'
import { useSite } from '@/lib/site'
import { usePageMeta } from '@/lib/meta'

export default function NotFound() {
  const { t } = useSite()
  usePageMeta(t('notfound.title') + t('seo.suffix'), t('notfound.body'))
  return (
    <section className="on-dark bg-charcoal text-white">
      <div className="wrap" style={{ paddingTop: 'clamp(56px,8vw,120px)', paddingBottom: 'clamp(56px,8vw,120px)' }}>
        <p className="eyebrow">404</p>
        <h1 className="t-page max-w-[820px]">{t('notfound.title')}</h1>
        <p className="t-lead mb-0 mt-6 max-w-[600px] text-ondark-body">{t('notfound.body')}</p>
        <Link to="/" className="btn btn-red mt-8">{t('notfound.home')} <ArrowRight aria-hidden="true" /></Link>
      </div>
    </section>
  )
}
