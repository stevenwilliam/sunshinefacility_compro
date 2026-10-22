import { useCallback, useEffect, useRef, useState } from 'react'
import { PencilSimple, Trash, UploadSimple } from '@/lib/phosphor'
import { ApiError, api, del, fmtTime, get, put, qs } from './api'
import { Confirm, CountLine, Drawer, Empty, ExportButton, ScreenHeader, SearchBox, useDebounced, useToast } from './ui'
import type { MediaItem } from './fields'
import { useSession } from './AdminApp'

const KINDS = [['photo', 'Photo'], ['logo', 'Logo'], ['certificate', 'Certificate'], ['brand', 'Brand']] as const

export default function MediaScreen() {
  const { me } = useSession()
  const toast = useToast()
  const [q, setQ] = useState('')
  const dq = useDebounced(q)
  const [kind, setKind] = useState('')
  const [rows, setRows] = useState<MediaItem[] | null>(null)
  const [all, setAll] = useState(0)
  const [upKind, setUpKind] = useState('photo')
  const [busy, setBusy] = useState(false)
  const [edit, setEdit] = useState<MediaItem | null>(null)
  const [confirm, setConfirm] = useState<MediaItem | null>(null)
  const fileRef = useRef<HTMLInputElement>(null)
  const writable = me.permissions.includes('media.write')
  const url = '/api/v1/admin/media' + qs({ q: dq, kind })
  const load = useCallback(() => {
    get<{ rows: MediaItem[] }>(url).then((r) => setRows(r.rows)).catch(() => setRows([]))
  }, [url])
  useEffect(load, [load])
  useEffect(() => { get<{ total: number }>('/api/v1/admin/media').then((r) => setAll(r.total)).catch(() => {}) }, [rows])

  const upload = async (f: File) => {
    setBusy(true)
    const fd = new FormData()
    fd.set('file', f)
    fd.set('kind', upKind)
    try {
      await api('POST', '/api/v1/admin/media', fd)
      toast('Uploaded ' + f.name)
      load()
    } catch (x) {
      toast(x instanceof ApiError ? x.body.message : 'Upload failed.', true)
    } finally {
      setBusy(false)
      if (fileRef.current) fileRef.current.value = ''
    }
  }

  return (
    <>
      <ScreenHeader group="Content" title="Media library" desc="Every image used on the site. Uploads are checked, then re-encoded to WebP at 480, 960 and 1600 px. Images in use cannot be deleted.">
        <SearchBox value={q} onChange={setQ} />
        <select aria-label="Type" className="inp inp-auto" value={kind} onChange={(e) => setKind(e.target.value)}>
          <option value="">All types</option>
          {KINDS.map(([k, l]) => <option key={k} value={k}>{l}</option>)}
        </select>
        <ExportButton url={url} count={rows?.length || 0} onError={(m) => toast(m, true)} />
        {writable && (
          <>
            <select aria-label="Upload as" className="inp inp-auto" value={upKind} onChange={(e) => setUpKind(e.target.value)}>
              {KINDS.map(([k, l]) => <option key={k} value={k}>Upload as {l.toLowerCase()}</option>)}
            </select>
            <label className="b b-red relative">
              <input ref={fileRef} type="file" accept="image/png,image/jpeg,image/webp" className="absolute h-px w-px overflow-hidden opacity-0"
                onChange={(e) => e.target.files?.[0] && upload(e.target.files[0])} disabled={busy} />
              <UploadSimple aria-hidden="true" className="text-[18px]" />{busy ? 'Uploading…' : 'Upload'}
            </label>
          </>
        )}
      </ScreenHeader>
      <main className="flex flex-col gap-3" style={{ padding: 'clamp(16px,3vw,28px) clamp(16px,3vw,32px) 48px' }}>
        <CountLine shown={rows?.length ?? 0} total={dq || kind ? all : rows?.length ?? 0}
          reasons={[
            { id: 'export-why', text: rows && rows.length === 0 ? 'Nothing to export: no images match the current search.' : undefined },
            { id: 'up-why', text: writable ? 'PNG, JPEG or WebP. SVG is not accepted (the brand marks arrive with the seed).' : 'Your role can view the library but not upload.' },
          ]} />
        {rows === null && <p className="muted">Loading…</p>}
        {rows?.length === 0 && <div className="surface"><Empty /></div>}
        <ul className="m-0 grid list-none gap-3 p-0" style={{ gridTemplateColumns: 'repeat(auto-fill,minmax(170px,1fr))' }}>
          {rows?.map((m) => {
            const why = !writable ? 'Your role cannot delete media.' : m.usage > 0 ? `Can’t delete: used in ${m.usage} item${m.usage > 1 ? 's' : ''}. Replace it there first.` : ''
            return (
              <li key={m.id} className="surface flex flex-col overflow-hidden">
                <div role="img" aria-label={m.altEn || m.name} className="bdr border-b bg-contain bg-center bg-no-repeat" style={{ aspectRatio: '4/3', backgroundColor: 'var(--a-surface-2)', backgroundImage: m.thumb ? `url("${m.thumb}")` : undefined }} />
                <div className="flex flex-col gap-1 px-3 py-2.5 text-[13px]">
                  <span className="truncate text-[14px] font-medium" title={m.name}>{m.name}</span>
                  <span className="faint">{m.kind} · {m.width}×{m.height} · {m.usage ? `Used ${m.usage}×` : 'Not used'}</span>
                  <div className="mt-1 flex items-center justify-end gap-1.5">
                    {m.usage > 0 && <span className="faint mr-auto text-[12px]">In use</span>}
                    <button type="button" className="ib h-9 w-9" aria-label={'Edit ' + m.name} onClick={() => setEdit(m)}><PencilSimple aria-hidden="true" /></button>
                    <button type="button" className="ib danger h-9 w-9" aria-label={'Delete ' + m.name} disabled={!!why} title={why || 'Delete image'} onClick={() => setConfirm(m)}><Trash aria-hidden="true" /></button>
                  </div>
                </div>
              </li>
            )
          })}
        </ul>
      </main>
      {edit && <MediaEditor m={edit} writable={writable} onClose={() => setEdit(null)} onSaved={() => { setEdit(null); load(); toast('Saved') }} />}
      {confirm && (
        <Confirm title={`Delete “${confirm.name}”?`} onCancel={() => setConfirm(null)} onConfirm={async () => {
          try {
            await del('/api/v1/admin/media/' + confirm.id)
            toast('Deleted ' + confirm.name)
            load()
          } catch (x) {
            toast(x instanceof ApiError ? x.body.message : 'Delete failed.', true)
          }
          setConfirm(null)
        }} />
      )}
    </>
  )
}

function MediaEditor({ m, writable, onClose, onSaved }: { m: MediaItem; writable: boolean; onClose: () => void; onSaved: () => void }) {
  const [f, setF] = useState({ name: m.name, altEn: m.altEn, altId: m.altId })
  const [errs, setErrs] = useState<Record<string, string>>({})
  const save = async () => {
    try {
      await put('/api/v1/admin/media/' + m.id, f)
      onSaved()
    } catch (x) {
      if (x instanceof ApiError) setErrs(x.body.fields || { name: x.body.message })
    }
  }
  return (
    <Drawer title={m.name} kicker="Media library" onClose={onClose} footer={<>
      <button type="button" className="b b-red" onClick={save} disabled={!writable || !f.name.trim()} aria-describedby="m-why">Save</button>
      <button type="button" className="b b-line" onClick={onClose}>Cancel</button>
      <p id="m-why" className="muted m-0 flex-[1_1_100%] text-[13px]">{!writable ? 'Your role can view media but not change it.' : !f.name.trim() ? 'To save, fill in: Name.' : 'Alt text describes the image for people who cannot see it.'}</p>
    </>}>
      <div className="flex flex-col gap-4">
        <img src={m.url || m.thumb} alt="" className="max-h-[260px] w-full rounded border object-contain bdr" style={{ background: 'var(--a-surface-2)' }} />
        <p className="faint m-0 text-[13px]">{m.kind} · {m.mime} · {m.width}×{m.height} · {(m.bytes / 1024).toFixed(0)} KB original · added {fmtTime((m as unknown as { createdAt: string }).createdAt)}</p>
        {(['name', 'altEn', 'altId'] as const).map((k) => (
          <label key={k} className="flex flex-col gap-1.5">
            <span className="lbl">{{ name: 'Name *', altEn: 'Alt text (EN)', altId: 'Alt text (ID)' }[k]}</span>
            <input className="inp" value={f[k]} disabled={!writable} maxLength={k === 'name' ? 200 : 300} onChange={(e) => setF((s) => ({ ...s, [k]: e.target.value }))} aria-invalid={!!errs[k] || undefined} />
            {errs[k] && <span className="ferr">{errs[k]}</span>}
          </label>
        ))}
      </div>
    </Drawer>
  )
}
