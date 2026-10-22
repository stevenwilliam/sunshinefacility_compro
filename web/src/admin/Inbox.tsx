import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { DownloadSimple } from '@/lib/phosphor'
import { ApiError, fmtTime, get, post, qs } from './api'
import { CountLine, Drawer, Empty, ExportButton, ScreenHeader, SearchBox, useDebounced, useToast } from './ui'

interface Quote {
  id: string; createdAt: string; name: string; company: string; email: string; phone: string; buildingType: string
  areaM2: number | null; location: string; services: string[]; message: string; lang: string
}
interface Application {
  id: string; createdAt: string; name: string; phone: string; email: string; position: string; message: string
  cvFilename: string; cvMime: string; cvBytes: number; lang: string
}

function useList<T>(base: string) {
  const [q, setQ] = useState('')
  const dq = useDebounced(q)
  const [rows, setRows] = useState<T[] | null>(null)
  const [all, setAll] = useState(0)
  const url = base + qs({ q: dq })
  useEffect(() => { get<{ rows: T[] }>(url).then((r) => setRows(r.rows)).catch(() => setRows([])) }, [url])
  useEffect(() => { get<{ total: number }>(base + '?limit=1').then((r) => setAll(r.total)).catch(() => {}) }, [base, rows])
  return { q, setQ, rows, url, all, searching: !!dq }
}

function Row({ k, v }: { k: string; v: React.ReactNode }) {
  return (
    <div className="bdr grid grid-cols-[140px_1fr] gap-3 border-b py-2.5 text-[15px]">
      <dt className="faint">{k}</dt><dd className="m-0 whitespace-pre-wrap break-words">{v || '—'}</dd>
    </div>
  )
}

// BR-7.1: the inbox is read-only (decision Q8). Follow-up happens by phone,
// WhatsApp or email outside the CMS.
export function QuotationsScreen() {
  const toast = useToast()
  const nav = useNavigate()
  const { id } = useParams()
  const l = useList<Quote>('/api/v1/admin/quotations')
  const [open, setOpen] = useState<Quote | null>(null)
  useEffect(() => {
    if (!id) return setOpen(null)
    get<Quote>('/api/v1/admin/quotations/' + id).then(setOpen).catch(() => toast('That request no longer exists.', true))
  }, [id, toast])
  return (
    <>
      <ScreenHeader group="Inbox" title="Quotation requests" desc="Requests from the Contact page, newest first. Read-only: reply to the client by phone, WhatsApp or email.">
        <SearchBox value={l.q} onChange={l.setQ} />
        <ExportButton url={l.url} count={l.rows?.length || 0} onError={(m) => toast(m, true)} />
      </ScreenHeader>
      <main className="flex flex-col gap-3" style={{ padding: 'clamp(16px,3vw,28px) clamp(16px,3vw,32px) 48px' }}>
        <CountLine shown={l.rows?.length ?? 0} total={l.searching ? l.all : l.rows?.length ?? 0} reasons={[{ id: 'export-why', text: l.rows?.length === 0 ? 'Nothing to export: no requests match the current search.' : undefined }]} />
        <div className="surface overflow-x-auto">
          <table className="grid-t min-w-[760px]">
            <thead><tr>{['Received', 'Name', 'Company', 'Building', 'Location', 'Services'].map((h) => <th key={h} scope="col">{h}</th>)}</tr></thead>
            <tbody>
              {l.rows?.map((r) => (
                <tr key={r.id} className="click" tabIndex={0} onClick={() => nav('/admin/inbox/quotations/' + r.id)} onKeyDown={(e) => e.key === 'Enter' && nav('/admin/inbox/quotations/' + r.id)}>
                  <td className="whitespace-nowrap tnum">{fmtTime(r.createdAt)}</td><td><span className="ell">{r.name}</span></td><td><span className="ell">{r.company}</span></td>
                  <td><span className="ell">{r.buildingType}</span></td><td><span className="ell">{r.location}</span></td><td><span className="ell">{r.services.join(', ')}</span></td>
                </tr>
              ))}
            </tbody>
          </table>
          {l.rows === null && <p className="muted p-5">Loading…</p>}
          {l.rows?.length === 0 && <Empty text={l.searching ? 'Nothing matches your search.' : 'No quotation requests yet.'} />}
        </div>
      </main>
      {open && (
        <Drawer title={open.company} kicker="Quotation request" onClose={() => nav('/admin/inbox/quotations')}>
          <dl className="m-0">
            <Row k="Received" v={fmtTime(open.createdAt)} />
            <Row k="Name" v={open.name} />
            <Row k="Email" v={<a className="underline" href={`mailto:${open.email}`}>{open.email}</a>} />
            <Row k="Phone" v={<a className="underline" href={`tel:${open.phone.replace(/[^0-9+]/g, '')}`}>{open.phone}</a>} />
            <Row k="Building type" v={open.buildingType} />
            <Row k="Area" v={open.areaM2 ? `${open.areaM2.toLocaleString('en-GB')} m²` : ''} />
            <Row k="Location" v={open.location} />
            <Row k="Services" v={open.services.join(', ')} />
            <Row k="Message" v={open.message} />
            <Row k="Language" v={open.lang === 'id' ? 'Bahasa Indonesia' : 'English'} />
          </dl>
        </Drawer>
      )}
    </>
  )
}

export function ApplicationsScreen() {
  const toast = useToast()
  const nav = useNavigate()
  const { id } = useParams()
  const l = useList<Application>('/api/v1/admin/applications')
  const [open, setOpen] = useState<Application | null>(null)
  useEffect(() => {
    if (!id) return setOpen(null)
    get<Application>('/api/v1/admin/applications/' + id).then(setOpen).catch(() => toast('That application no longer exists.', true))
  }, [id, toast])
  const cv = async (a: Application) => {
    try {
      const r = await post<{ url: string }>(`/api/v1/admin/applications/${a.id}/cv`)
      window.open(r.url, '_blank', 'noopener')
    } catch (x) {
      toast(x instanceof ApiError ? x.body.message : 'The CV could not be opened.', true)
    }
  }
  return (
    <>
      <ScreenHeader group="Inbox" title="Job applications" desc="Applications from the Careers page. CVs are personal data: each download is logged, and applications are deleted automatically after the retention period in Parameters → Careers.">
        <SearchBox value={l.q} onChange={l.setQ} />
        <ExportButton url={l.url} count={l.rows?.length || 0} onError={(m) => toast(m, true)} />
      </ScreenHeader>
      <main className="flex flex-col gap-3" style={{ padding: 'clamp(16px,3vw,28px) clamp(16px,3vw,32px) 48px' }}>
        <CountLine shown={l.rows?.length ?? 0} total={l.searching ? l.all : l.rows?.length ?? 0} reasons={[{ id: 'export-why', text: l.rows?.length === 0 ? 'Nothing to export: no applications match the current search.' : undefined }]} />
        <div className="surface overflow-x-auto">
          <table className="grid-t min-w-[700px]">
            <thead><tr>{['Received', 'Name', 'Position', 'Phone', 'Email'].map((h) => <th key={h} scope="col">{h}</th>)}</tr></thead>
            <tbody>
              {l.rows?.map((r) => (
                <tr key={r.id} className="click" tabIndex={0} onClick={() => nav('/admin/inbox/applications/' + r.id)} onKeyDown={(e) => e.key === 'Enter' && nav('/admin/inbox/applications/' + r.id)}>
                  <td className="whitespace-nowrap tnum">{fmtTime(r.createdAt)}</td><td><span className="ell">{r.name}</span></td><td><span className="ell">{r.position}</span></td>
                  <td><span className="ell">{r.phone}</span></td><td><span className="ell">{r.email}</span></td>
                </tr>
              ))}
            </tbody>
          </table>
          {l.rows === null && <p className="muted p-5">Loading…</p>}
          {l.rows?.length === 0 && <Empty text={l.searching ? 'Nothing matches your search.' : 'No applications yet.'} />}
        </div>
      </main>
      {open && (
        <Drawer title={open.name} kicker="Job application" onClose={() => nav('/admin/inbox/applications')}
          footer={<button type="button" className="b b-red" onClick={() => cv(open)}><DownloadSimple aria-hidden="true" />Open CV ({(open.cvBytes / 1024).toFixed(0)} KB)</button>}>
          <dl className="m-0">
            <Row k="Received" v={fmtTime(open.createdAt)} />
            <Row k="Position" v={open.position} />
            <Row k="Phone" v={<a className="underline" href={`tel:${open.phone.replace(/[^0-9+]/g, '')}`}>{open.phone}</a>} />
            <Row k="Email" v={<a className="underline" href={`mailto:${open.email}`}>{open.email}</a>} />
            <Row k="CV file" v={open.cvFilename} />
            <Row k="Message" v={open.message} />
            <Row k="Language" v={open.lang === 'id' ? 'Bahasa Indonesia' : 'English'} />
          </dl>
          <p className="hint mt-4">The CV link works for five minutes and the download is recorded in the audit history.</p>
        </Drawer>
      )}
    </>
  )
}
