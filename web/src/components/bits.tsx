import { Link } from 'react-router-dom'
import type { CSSProperties, ReactNode } from 'react'
import type { Img } from '@/lib/types'
import { useSite } from '@/lib/site'

/** A responsive image from the media library, sized to avoid layout shift. */
export function Picture({ img, sizes, className, style, alt, eager }: {
  img?: Img; sizes?: string; className?: string; style?: CSSProperties; alt?: string; eager?: boolean
}) {
  const { L } = useSite()
  if (!img) return null
  return (
    <img src={img.src} srcSet={img.srcset} sizes={sizes} width={img.w || undefined} height={img.h || undefined}
      alt={alt ?? L(img.alt)} className={className} style={style} loading={eager ? 'eager' : 'lazy'} decoding="async" />
  )
}

/** url("…") for CSS backgrounds and masks. */
export const cssUrl = (img?: Img) => (img ? `url("${img.src}")` : 'none')

/** The visible "Placeholder" tag (BR-2.1). */
export function PH({ show, label, small, className }: { show: boolean; label?: string; small?: boolean; className?: string }) {
  const { t } = useSite()
  if (!show) return null
  return <span className={`ph-tag${small ? ' sm' : ''}${className ? ' ' + className : ''}`}>{label || t('common.placeholder')}</span>
}

export function Breadcrumb({ items }: { items: { label: string; to?: string }[] }) {
  const { t } = useSite()
  return (
    <nav aria-label={t('common.breadcrumb')} className="crumbs">
      {items.map((it, i) => (
        <span key={i} className="contents">
          {i > 0 && <span aria-hidden="true">/</span>}
          {it.to ? <Link to={it.to}>{it.label}</Link> : <span aria-current="page">{it.label}</span>}
        </span>
      ))}
    </nav>
  )
}

export function Stats({ items, rule }: { items: { value: [string, string]; label: [string, string] }[]; rule?: string }) {
  const { L } = useSite()
  return (
    <dl className="m-0 grid gap-x-8 gap-y-6" style={{ gridTemplateColumns: 'repeat(auto-fit,minmax(150px,1fr))' }}>
      {items.map((s, i) => (
        <div key={i} className="stat" style={rule ? ({ '--rule': rule } as CSSProperties) : undefined}>
          <dt>{L(s.label)}</dt>
          <dd className="t-stat">{L(s.value)}</dd>
        </div>
      ))}
    </dl>
  )
}

/** Paragraphs from a body with blank-line breaks. */
export function Paras({ text, className }: { text: string; className?: string }) {
  const parts = text.split(/\n\s*\n/).filter(Boolean)
  return <>{parts.map((p, i) => <p key={i} className={className}>{p}</p>)}</>
}

export function Section({ children, className, style, ...rest }: { children: ReactNode; className?: string; style?: CSSProperties } & Record<string, unknown>) {
  return <section className={className} style={style} {...rest}>{children}</section>
}
