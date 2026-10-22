import { useEffect, useMemo, useState } from 'react'
import { ArrowDown, ArrowUp, Images, X } from '@/lib/phosphor'
import { ICONS, Icon } from '@/lib/icons'
import { get } from './api'
import { useDebounced } from './ui'

export interface Option { value: string; label: string }
export interface Field {
  key: string; label: string; type: string; required?: boolean; max?: number; minInt?: number; maxInt?: number
  options?: Option[]; group?: string; ref?: string; mediaKind?: string; content?: boolean; list?: boolean
  search?: boolean; immutable?: boolean; help?: string
}
export interface Options { refs: Record<string, Option[]>; params: Record<string, Option[]>; groups: Option[] }
export interface MediaItem { id: string; name: string; kind: string; thumb: string; url: string; usage: number; hash?: string; width: number; height: number; altEn: string; altId: string; slug?: string; mime: string; bytes: number }

/** WCAG contrast, the same arithmetic as internal/domain/contrast. */
export function contrast(a: string, b: string): number | null {
  const lum = (h: string) => {
    const m = /^#?([0-9a-f]{6})$/i.exec(h.trim())
    if (!m) return null
    const n = parseInt(m[1], 16)
    const c = [n >> 16, (n >> 8) & 255, n & 255].map((v) => {
      const x = v / 255
      return x <= 0.04045 ? x / 12.92 : Math.pow((x + 0.055) / 1.055, 2.4)
    })
    return 0.2126 * c[0] + 0.7152 * c[1] + 0.0722 * c[2]
  }
  const la = lum(a), lb = lum(b)
  if (la === null || lb === null) return null
  const [hi, lo] = la > lb ? [la, lb] : [lb, la]
  return (hi + 0.05) / (lo + 0.05)
}

export function MediaPicker({ kind, current, onPick, onClose }: { kind?: string; current?: string; onPick: (m: MediaItem) => void; onClose: () => void }) {
  const [q, setQ] = useState('')
  const [k, setK] = useState(kind || '')
  const dq = useDebounced(q)
  const [items, setItems] = useState<MediaItem[] | null>(null)
  useEffect(() => {
    get<{ rows: MediaItem[] }>(`/api/v1/admin/media?q=${encodeURIComponent(dq)}&kind=${encodeURIComponent(k)}`).then((r) => setItems(r.rows)).catch(() => setItems([]))
  }, [dq, k])
  useEffect(() => {
    const e = (ev: KeyboardEvent) => ev.key === 'Escape' && onClose()
    window.addEventListener('keydown', e)
    return () => window.removeEventListener('keydown', e)
  }, [onClose])
  return (
    <div className="modal">
      <div role="dialog" aria-modal="true" aria-labelledby="pk-h" className="flex max-h-[86vh] w-[min(900px,100%)] flex-col rounded shadow-modal" style={{ background: 'var(--a-surface)' }}>
        <header className="bdr flex flex-wrap items-center gap-3 border-b px-5 py-4">
          <h2 id="pk-h" className="m-0 flex-[1_1_200px] text-[20px] font-[550]">Choose an image</h2>
          <input type="search" aria-label="Search images" placeholder="Search images" value={q} onChange={(e) => setQ(e.target.value)} className="inp flex-[1_1_200px]" autoFocus />
          <select aria-label="Type" value={k} onChange={(e) => setK(e.target.value)} className="inp inp-auto">
            <option value="">All types</option>
            {['photo', 'logo', 'certificate', 'brand'].map((x) => <option key={x} value={x}>{x[0].toUpperCase() + x.slice(1)}</option>)}
          </select>
          <button type="button" onClick={onClose} aria-label="Close" className="ib"><X aria-hidden="true" /></button>
        </header>
        <ul className="m-0 grid list-none gap-[10px] overflow-y-auto p-5" style={{ gridTemplateColumns: 'repeat(auto-fill,minmax(140px,1fr))' }}>
          {items === null && <li className="muted">Loading…</li>}
          {items?.length === 0 && <li className="muted">No images match. Upload one in the Media library.</li>}
          {items?.filter((m) => m.thumb).map((m) => (
            <li key={m.id}>
              <button type="button" onClick={() => onPick(m)} aria-pressed={m.id === current}
                className="w-full cursor-pointer overflow-hidden rounded p-0 text-left" style={{ border: `2px solid ${m.id === current ? '#C6241C' : 'var(--a-border)'}`, background: 'var(--a-surface)', color: 'var(--a-text)' }}>
                <div aria-hidden="true" className="bg-cotton bg-contain bg-center bg-no-repeat" style={{ aspectRatio: '4/3', backgroundImage: `url("${m.thumb}")` }} />
                <span className="block truncate px-2.5 py-2 text-[13px]">{m.name}</span>
              </button>
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}

interface FieldProps {
  f: Field
  value: unknown
  set: (v: unknown) => void
  error?: string
  opts: Options | null
  media: Record<string, MediaItem>
  icons: string[]
  locked?: boolean
  disabledWhy?: string
  hintExtra?: string
}

export function FieldInput({ f, value, set, error, opts, media, icons, locked, disabledWhy, hintExtra }: FieldProps) {
  const [pick, setPick] = useState(false)
  const id = 'f-' + f.key
  const desc = [error ? id + '-err' : '', f.help || disabledWhy || hintExtra ? id + '-hint' : ''].filter(Boolean).join(' ') || undefined
  const common = { id, 'aria-invalid': !!error || undefined, 'aria-describedby': desc, disabled: locked }
  const str = value == null ? '' : String(value)
  const label = (
    <label htmlFor={id} className="lbl">{f.label}{f.required ? ' *' : ''}{locked && <span className="faint text-[12px] font-normal"> · fixed</span>}</label>
  )
  let input: JSX.Element
  switch (f.type) {
    case 'textarea':
    case 'lines':
      input = <textarea {...common} className="inp" rows={f.type === 'lines' ? 5 : 4} maxLength={f.max} value={str} onChange={(e) => set(e.target.value)} />
      break
    case 'int':
      input = <input {...common} className="inp" type="number" inputMode="numeric" min={f.minInt} max={f.maxInt} value={str} onChange={(e) => set(e.target.value)} />
      break
    case 'date':
      input = <input {...common} className="inp" type="date" value={str} onChange={(e) => set(e.target.value)} />
      break
    case 'bool': {
      const on = !!value
      input = (
        <label className="flex cursor-pointer items-center gap-[10px] text-[15px]" style={{ opacity: disabledWhy ? 0.5 : 1 }}>
          <input {...common} type="checkbox" checked={on} disabled={!!disabledWhy || locked} onChange={(e) => set(e.target.checked)} className="m-0 h-5 w-5 accent-charcoal" />
          {f.key === 'published' ? (on ? 'Published on the site' : 'Draft (hidden from the site)') : f.key === 'is_placeholder' ? (on ? 'Still a placeholder (tagged on the site)' : 'Real content') : on ? 'On' : 'Off'}
        </label>
      )
      break
    }
    case 'select':
      input = (
        <select {...common} className="inp" value={str} onChange={(e) => set(e.target.value)}>
          <option value="">Choose…</option>
          {(f.options || []).map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
        </select>
      )
      break
    case 'param':
      input = (
        <select {...common} className="inp" value={str} onChange={(e) => set(e.target.value)}>
          <option value="">Choose…</option>
          {(opts?.params[f.group || ''] || []).map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
        </select>
      )
      break
    case 'ref':
      input = (
        <select {...common} className="inp" value={str} onChange={(e) => set(e.target.value)}>
          <option value="">{f.required ? 'Choose…' : '— none —'}</option>
          {(opts?.refs[f.ref || ''] || []).map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
        </select>
      )
      break
    case 'refs':
      input = <RefsInput {...common} options={opts?.refs[f.ref || ''] || []} value={Array.isArray(value) ? (value as string[]) : []} set={set} />
      break
    case 'icon':
      input = (
        <div className="flex items-center gap-3">
          <span className="grid h-11 w-11 flex-none place-items-center rounded border text-[24px] bdr"><Icon name={str} /></span>
          <select {...common} className="inp" value={str} onChange={(e) => set(e.target.value)}>
            <option value="">— none —</option>
            {icons.filter((i) => ICONS[i]).map((i) => <option key={i} value={i}>{i}</option>)}
          </select>
        </div>
      )
      break
    case 'colour':
      input = (
        <div className="flex items-center gap-3">
          <span className="swatch h-11 w-11 flex-none rounded" style={{ background: /^#[0-9a-f]{6}$/i.test(str) ? str : 'transparent' }} aria-hidden="true" />
          <input {...common} className="inp" value={str} onChange={(e) => set(e.target.value.toUpperCase())} placeholder="#C6241C" list="brand-colours" maxLength={7} />
          <datalist id="brand-colours">{(opts?.params.brand_colours || []).map((o) => <option key={o.value} value={o.label} />)}</datalist>
        </div>
      )
      break
    case 'media': {
      const m = media[str]
      input = (
        <div className="flex flex-wrap items-center gap-3">
          <div role="img" aria-label={m ? m.name : 'No image'} className="faint grid h-24 w-[132px] place-items-center rounded border bg-contain bg-center bg-no-repeat text-[12px] bdr"
            style={{ backgroundColor: 'var(--a-surface-2)', backgroundImage: m?.thumb ? `url("${m.thumb}")` : undefined }}>{m ? '' : 'No image'}</div>
          <div className="flex flex-col gap-2">
            <button {...common} type="button" className="b b-line min-h-[40px] text-[14px]" onClick={() => setPick(true)}><Images aria-hidden="true" />Choose from media</button>
            {str && !f.required && <button type="button" className="b b-ghost min-h-[36px] px-2.5 text-[13px]" onClick={() => set('')}>Remove image</button>}
            {m && <span className="faint text-[12px]">{m.name}</span>}
          </div>
          {pick && <MediaPicker kind={f.mediaKind} current={str} onClose={() => setPick(false)} onPick={(x) => { set(x.id); setPick(false) }} />}
        </div>
      )
      break
    }
    default:
      input = <input {...common} className="inp" maxLength={f.max} value={str} onChange={(e) => set(e.target.value)} />
  }
  const full = ['textarea', 'lines', 'media', 'bool', 'refs'].includes(f.type)
  const count = (f.type === 'text' || f.type === 'textarea' || f.type === 'lines') && f.max ? `${str.length}/${f.max}` : ''
  return (
    <div className="flex flex-col gap-1.5" style={{ gridColumn: full ? '1 / -1' : 'auto' }}>
      {f.type !== 'bool' ? label : <span className="lbl">{f.label}</span>}
      {input}
      {error && <p id={id + '-err'} className="ferr m-0">{error}</p>}
      {(disabledWhy || hintExtra || f.help) && (
        <p id={id + '-hint'} className="hint m-0" style={disabledWhy ? { color: 'var(--a-danger)' } : undefined}>
          {disabledWhy || hintExtra || (f.help && f.help[0] === f.help[0].toLowerCase() ? f.label + ' ' + f.help : f.help)}
        </p>
      )}
      {count && <span className="faint self-end text-[11px] tnum">{count}</span>}
    </div>
  )
}

function RefsInput({ options, value, set, ...rest }: { options: Option[]; value: string[]; set: (v: unknown) => void; id: string }) {
  const label = useMemo(() => Object.fromEntries(options.map((o) => [o.value, o.label])), [options])
  const left = options.filter((o) => !value.includes(o.value))
  const move = (i: number, d: number) => {
    const v = [...value]
    const j = i + d
    if (j < 0 || j >= v.length) return
    ;[v[i], v[j]] = [v[j], v[i]]
    set(v)
  }
  return (
    <div className="flex flex-col gap-2">
      <ol className="m-0 flex list-none flex-col gap-1.5 p-0">
        {value.length === 0 && <li className="faint text-[14px]">None chosen.</li>}
        {value.map((v, i) => (
          <li key={v} className="flex items-center gap-2 rounded border px-3 py-1.5 text-[14px] bdr">
            <span className="faint w-5 tnum">{i + 1}.</span><span className="flex-1">{label[v] || v}</span>
            <button type="button" className="ib h-8 w-8" aria-label={`Move ${label[v]} up`} disabled={i === 0} onClick={() => move(i, -1)}><ArrowUp aria-hidden="true" /></button>
            <button type="button" className="ib h-8 w-8" aria-label={`Move ${label[v]} down`} disabled={i === value.length - 1} onClick={() => move(i, 1)}><ArrowDown aria-hidden="true" /></button>
            <button type="button" className="ib danger h-8 w-8" aria-label={`Remove ${label[v]}`} onClick={() => set(value.filter((x) => x !== v))}><X aria-hidden="true" /></button>
          </li>
        ))}
      </ol>
      <select {...rest} className="inp" value="" onChange={(e) => e.target.value && set([...value, e.target.value])} disabled={left.length === 0}>
        <option value="">{left.length ? 'Add…' : 'All are chosen'}</option>
        {left.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
      </select>
    </div>
  )
}
