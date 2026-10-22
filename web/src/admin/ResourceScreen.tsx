import { useCallback, useEffect, useMemo, useState } from 'react'
import { useParams } from 'react-router-dom'
import { PencilSimple, Trash } from '@/lib/phosphor'
import { Icon } from '@/lib/icons'
import { ApiError, del, fmtDate, get, post, put, qs, todayJakarta } from './api'
import { AddButton, Confirm, CountLine, Drawer, Empty, ExportButton, ScreenHeader, SearchBox, StatusPill, useDebounced, useToast } from './ui'
import { contrast, FieldInput, type Field, type MediaItem, type Option, type Options } from './fields'
import { useSession } from './AdminApp'

export interface Def {
  name: string; idColumn: string; label: string; one: string; description: string; nav: string; writePerm: string
  fields: Field[]; filters?: { key: string; label: string; options?: Option[]; ref?: string; group?: string }[]
  nameField: string; canCreate: boolean; createWhy?: string; canDelete: boolean; deleteWhy?: string
}
export interface Schema { resources: Def[]; icons: string[]; paramGroups: { key: string; label: string; system: boolean; open: boolean }[] }
type Row = Record<string, unknown>

export function useLookups() {
  const [opts, setOpts] = useState<Options | null>(null)
  const [media, setMedia] = useState<Record<string, MediaItem>>({})
  const reload = useCallback(() => {
    get<Options>('/api/v1/admin/options').then(setOpts).catch(() => {})
    get<{ rows: MediaItem[] }>('/api/v1/admin/media').then((r) => setMedia(Object.fromEntries(r.rows.map((m) => [m.id, m])))).catch(() => {})
  }, [])
  useEffect(reload, [reload])
  return { opts, media, reload }
}

export default function ResourceScreen() {
  const { name = '' } = useParams()
  const { schema, me, reloadPH } = useSession()
  const toast = useToast()
  const def = schema?.resources.find((d) => d.name === name)
  const { opts, media, reload } = useLookups()
  const [q, setQ] = useState('')
  const dq = useDebounced(q)
  const [filters, setFilters] = useState<Record<string, string>>({})
  const [rows, setRows] = useState<Row[] | null>(null)
  const [total, setTotal] = useState(0)
  const [all, setAll] = useState(0)
  const [edit, setEdit] = useState<{ id: string | null; row: Row } | null>(null)
  const [confirm, setConfirm] = useState<{ id: string; name: string } | null>(null)

  useEffect(() => { setQ(''); setFilters({}); setEdit(null) }, [name])
  const listURL = `/api/v1/admin/r/${name}` + qs({ q: dq, ...Object.fromEntries(Object.entries(filters).map(([k, v]) => ['f.' + k, v])) })
  const load = useCallback(() => {
    if (!def) return
    get<{ rows: Row[]; total: number }>(listURL).then((r) => { setRows(r.rows); setTotal(r.total) }).catch(() => setRows([]))
  }, [def, listURL])
  useEffect(() => { setRows(null); load() }, [load])
  useEffect(() => {
    if (def) get<{ total: number }>(`/api/v1/admin/r/${name}?limit=1`).then((r) => setAll(r.total)).catch(() => {})
  }, [def, name, rows])

  if (!schema) return <p className="muted p-8">Loading…</p>
  if (!def) return <p className="muted p-8">This section does not exist or you do not have access to it.</p>

  const writable = me.permissions.includes(def.writePerm)
  const isParams = def.name === 'parameters'
  const groupOf = (k: string) => schema.paramGroups.find((g) => g.key === k)
  // Parameters: rows can be added only to an open (option) group.
  const addWhy = !writable ? 'Your role can view this section but not change it.'
    : !def.canCreate ? def.createWhy
    : isParams && !(filters.grp && groupOf(filters.grp)?.open) ? 'To add a parameter, first filter to an option group: ' + schema.paramGroups.filter((g) => g.open).map((g) => g.label).join(', ') + '. Other groups have a fixed set of keys.'
    : undefined
  const listFields = def.fields.filter((f) => f.list)
  const label = (r: Row) => String(r[def.nameField] || def.one)
  const deleteWhy = (r: Row) => !writable ? 'Your role cannot delete here.' : !def.canDelete ? def.deleteWhy : isParams && !r._open ? 'This parameter is used by the site and cannot be deleted. You can change its value.' : undefined

  const doDelete = async () => {
    if (!confirm) return
    try {
      await del(`/api/v1/admin/r/${name}/${encodeURIComponent(confirm.id)}`)
      toast('Deleted ' + confirm.name)
      setConfirm(null)
      setEdit(null)
      load()
      reload()
      reloadPH()
    } catch (x) {
      setConfirm(null)
      toast(x instanceof ApiError ? x.body.message : 'Delete failed.', true)
    }
  }

  return (
    <>
      <ScreenHeader group={def.nav} title={def.label} desc={def.description}>
        <SearchBox value={q} onChange={setQ} />
        {(def.filters || []).map((fl) => {
          const options: Option[] = fl.options || (fl.ref ? opts?.refs[fl.ref] || [] : fl.group ? opts?.params[fl.group] || [] : isParams ? schema.paramGroups.map((g) => ({ value: g.key, label: g.label })) : [])
          return (
            <select key={fl.key} aria-label={fl.label} className="inp inp-auto" value={filters[fl.key] || ''} onChange={(e) => setFilters((s) => ({ ...s, [fl.key]: e.target.value }))}>
              <option value="">All {fl.label.toLowerCase()}s</option>
              {options.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
            </select>
          )
        })}
        <ExportButton url={listURL} count={rows?.length || 0} onError={(m) => toast(m, true)} />
        <AddButton label={'Add ' + def.one} disabledWhy={addWhy} onClick={() => {
          const row: Row = {}
          for (const f of def.fields) row[f.key] = f.type === 'bool' ? f.key === 'published' && def.name !== 'certificates' && def.name !== 'testimonials' : f.type === 'refs' ? [] : ''
          if (isParams && filters.grp) row.grp = filters.grp
          for (const fl of def.filters || []) if (filters[fl.key] && fl.key !== 'grp') row[fl.key] = filters[fl.key]
          setEdit({ id: null, row })
        }} />
      </ScreenHeader>
      <main className="flex flex-col gap-3" style={{ padding: 'clamp(16px,3vw,28px) clamp(16px,3vw,32px) 48px' }}>
        <CountLine shown={rows?.length ?? 0} total={dq || Object.values(filters).some(Boolean) ? all : total}
          reasons={[{ id: 'export-why', text: rows && rows.length === 0 ? 'Nothing to export: no rows match the current search.' : undefined }, { id: 'add-why', text: addWhy }]} />
        <div className="surface overflow-x-auto">
          <table className="grid-t min-w-[680px]">
            <thead>
              <tr>
                {listFields.map((f) => <th key={f.key} scope="col">{f.key === 'published' ? 'Status' : f.key === 'sort' ? '#' : f.label}</th>)}
                <th scope="col" className="text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              {rows?.map((r) => {
                const id = String(r.id)
                const dw = deleteWhy(r)
                return (
                  <tr key={id}>
                    {listFields.map((f) => <td key={f.key}><Cell f={f} r={r} def={def} opts={opts} media={media} /></td>)}
                    <td className="whitespace-nowrap text-right">
                      <div className="inline-flex gap-1.5">
                        <button type="button" className="ib" aria-label={'Edit ' + label(r)} onClick={() => setEdit({ id, row: r })}><PencilSimple aria-hidden="true" /></button>
                        <button type="button" className="ib danger" aria-label={'Delete ' + label(r)} disabled={!!dw} title={dw || 'Delete'} onClick={() => setConfirm({ id, name: label(r) })}><Trash aria-hidden="true" /></button>
                      </div>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
          {rows === null && <p className="muted p-5">Loading…</p>}
          {rows?.length === 0 && <Empty />}
        </div>
        {rows && rows.length > 0 && !def.canDelete && <p className="faint m-0 text-[13px]">Delete is unavailable here: {def.deleteWhy}</p>}
      </main>
      {edit && (
        <Editor def={def} schema={schema} initial={edit.row} id={edit.id} writable={writable} opts={opts} media={media}
          onClose={() => setEdit(null)}
          onDelete={edit.id && !deleteWhy(edit.row) ? () => setConfirm({ id: edit.id!, name: label(edit.row) }) : undefined}
          onSaved={(r, created) => { toast((created ? 'Created ' : 'Saved ') + label(r)); setEdit(null); load(); reload(); reloadPH() }} />
      )}
      {confirm && <Confirm title={`Delete “${confirm.name}”?`} onConfirm={doDelete} onCancel={() => setConfirm(null)} />}
    </>
  )
}

function Cell({ f, r, def, opts, media }: { f: Field; r: Row; def: Def; opts: Options | null; media: Record<string, MediaItem> }) {
  const v = r[f.key]
  if (f.key === 'published') {
    if (def.name === 'certificates' && r._status === 'expired') return <StatusPill kind="bad" text="Expired" />
    return v ? <StatusPill kind="ok" text="Published" /> : <StatusPill kind="draft" text="Draft" />
  }
  if (f.key === 'is_placeholder') return v ? <StatusPill kind="ph" text="Placeholder" /> : <span className="faint">—</span>
  switch (f.type) {
    case 'media': {
      const m = media[String(v || '')]
      return <div role="img" aria-label={m?.name || 'No image'} className="thumb" style={{ backgroundImage: m?.thumb ? `url("${m.thumb}")` : undefined }} />
    }
    case 'colour':
      return <span className="inline-flex items-center gap-2 tnum"><span className="swatch" style={{ background: String(v || 'transparent') }} />{String(v || '—')}</span>
    case 'icon':
      return <span className="text-[22px]"><Icon name={String(v || '')} /></span>
    case 'ref':
      return <span className="ell">{opts?.refs[f.ref || '']?.find((o) => o.value === v)?.label || '—'}</span>
    case 'param':
      return <span className="ell">{opts?.params[f.group || '']?.find((o) => o.value === v)?.label || String(v || '—')}</span>
    case 'select':
      return <span className="ell">{f.options?.find((o) => o.value === v)?.label || String(v || '—')}</span>
    case 'date':
      return <span className={`tnum ${def.name === 'certificates' && r._status === 'expired' ? 'text-red-hover' : ''}`} style={def.name === 'certificates' && r._status === 'expired' ? { color: 'var(--a-danger)' } : undefined}>{fmtDate(String(v || ''))}</span>
    case 'bool':
      return <span>{v ? 'Yes' : 'No'}</span>
  }
  if (def.name === 'parameters' && f.key === 'grp') return <span className="ell">{String(r._groupLabel || v)}</span>
  const s = v == null || v === '' ? '—' : String(v)
  return <span className="ell" title={s.length > 40 ? s : undefined} style={{ color: s === '—' ? 'var(--a-faint)' : undefined }}>{s}</span>
}

function Editor({ def, schema, initial, id, writable, opts, media, onClose, onSaved, onDelete }: {
  def: Def; schema: Schema; initial: Row; id: string | null; writable: boolean; opts: Options | null; media: Record<string, MediaItem>
  onClose: () => void; onSaved: (r: Row, created: boolean) => void; onDelete?: () => void
}) {
  const [draft, setDraft] = useState<Row>(() => ({ ...initial }))
  const [errs, setErrs] = useState<Record<string, string>>({})
  const [busy, setBusy] = useState(false)
  const [formErr, setFormErr] = useState('')
  const set = (k: string) => (v: unknown) => { setDraft((d) => ({ ...d, [k]: v })); setErrs((e) => ({ ...e, [k]: '' })) }
  const missing = def.fields.filter((f) => f.required && (draft[f.key] === undefined || draft[f.key] === null || (Array.isArray(draft[f.key]) ? (draft[f.key] as unknown[]).length === 0 : String(draft[f.key]).trim() === ''))).map((f) => f.label)

  // BR-3.2 in the form: an expired certificate's Published toggle is
  // disabled, with the reason and the date.
  const expired = def.name === 'certificates' && !!draft.valid_until && String(draft.valid_until) < todayJakarta()
  // BR-11.1 in the form: live contrast for the colour pairs that carry text.
  const ratios = useMemo(() => {
    if (def.name !== 'services') return {}
    const r1 = contrast(String(draft.color_on_ink || ''), String(draft.color_ink || ''))
    const r2 = contrast(String(draft.color_label || ''), '#FFFFFF')
    return { on: r1, label: r2 }
  }, [def.name, draft.color_on_ink, draft.color_ink, draft.color_label])
  const grp = def.name === 'parameters' ? schema.paramGroups.find((g) => g.key === draft.grp) : undefined

  const save = async () => {
    setBusy(true)
    setFormErr('')
    const body: Row = {}
    for (const f of def.fields) {
      if (id && f.immutable) { body[f.key] = initial[f.key]; continue }
      let v = draft[f.key]
      if (f.key === 'published' && expired) v = false
      body[f.key] = v
    }
    try {
      const r = id ? await put<Row>(`/api/v1/admin/r/${def.name}/${encodeURIComponent(id)}`, body) : await post<Row>(`/api/v1/admin/r/${def.name}`, body)
      onSaved(r, !id)
    } catch (x) {
      if (x instanceof ApiError) {
        setErrs(x.body.fields || {})
        setFormErr(x.body.message)
      } else setFormErr('Save failed. Please try again.')
    } finally {
      setBusy(false)
    }
  }

  const title = id ? String(initial[def.nameField] || 'Edit') : 'New ' + def.one
  const cantSave = !writable || missing.length > 0 || busy
  const why = !writable ? 'Your role can view this section but not change it.' : missing.length ? 'To save, fill in: ' + missing.join(', ') + '.' : id ? 'Changes go live when you save.' : 'Ready to save.'
  return (
    <Drawer title={title.length > 60 ? title.slice(0, 60) + '…' : title} kicker={def.label} onClose={onClose}
      footer={<>
        <button type="button" className="b b-red" onClick={save} disabled={cantSave} aria-describedby="save-why">{busy ? 'Saving…' : 'Save'}</button>
        <button type="button" className="b b-line" onClick={onClose}>Cancel</button>
        {id && onDelete && writable && <button type="button" className="b b-ghost ml-auto" onClick={onDelete}>Delete</button>}
        <p id="save-why" aria-live="polite" className="muted m-0 flex-[1_1_100%] text-[13px]">{why}</p>
        {formErr && <p role="alert" className="ferr m-0 flex-[1_1_100%]">{formErr}</p>}
      </>}>
      <div className="grid gap-4" style={{ gridTemplateColumns: 'repeat(auto-fit,minmax(min(100%,220px),1fr))' }}>
        {def.fields.map((f) => {
          let disabledWhy: string | undefined
          let hintExtra: string | undefined
          if (f.key === 'published' && expired) disabledWhy = `Can't publish: this certificate expired on ${fmtDate(String(draft.valid_until))}. Upload the renewed certificate and update the date.`
          if (f.key === 'color_on_ink' && ratios.on) hintExtra = `${ratios.on.toFixed(2)}:1 against the ink colour — ${ratios.on >= 4.5 ? 'passes' : 'FAILS'} AA (4.5:1).`
          if (f.key === 'color_label' && ratios.label) hintExtra = `${ratios.label.toFixed(2)}:1 on white — ${ratios.label >= 4.5 ? 'passes' : 'FAILS'} AA (4.5:1).`
          if (def.name === 'parameters' && f.key === 'value_en' && initial.description) hintExtra = String(initial.description)
          if (def.name === 'parameters' && f.key === 'grp' && !id) {
            return (
              <div key={f.key} className="flex flex-col gap-1.5">
                <label htmlFor="f-grp" className="lbl">Group *</label>
                <select id="f-grp" className="inp" value={String(draft.grp || '')} onChange={(e) => set('grp')(e.target.value)}>
                  <option value="">Choose…</option>
                  {schema.paramGroups.filter((g) => g.open).map((g) => <option key={g.key} value={g.key}>{g.label}</option>)}
                </select>
                <p className="hint m-0">Only option groups take new rows.</p>
              </div>
            )
          }
          if (def.name === 'parameters' && f.key === 'value_id' && grp && !grp.open && !['contact', 'building_types', 'locations', 'job_types'].includes(grp.key) && initial.key !== 'address' && initial.key !== 'hours') {
            hintExtra = 'Not language-specific: the English value is used for both.'
          }
          return (
            <FieldInput key={f.key} f={f} value={draft[f.key]} set={set(f.key)} error={errs[f.key]} opts={opts} media={media}
              icons={schema.icons} locked={!writable || (!!id && f.immutable)} disabledWhy={disabledWhy} hintExtra={hintExtra} />
          )
        })}
      </div>
    </Drawer>
  )
}
