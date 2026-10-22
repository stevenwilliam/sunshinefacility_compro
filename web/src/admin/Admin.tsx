import { useCallback, useEffect, useState, type FormEvent } from 'react'
import { PencilSimple } from '@/lib/phosphor'
import { ApiError, fmtTime, get, post, put, qs } from './api'
import { AddButton, CountLine, Drawer, Empty, ExportButton, ScreenHeader, SearchBox, StatusPill, useDebounced, useToast } from './ui'
import { useSession } from './AdminApp'

const ROLES = [
  ['superadmin', 'Superadmin — everything, including users, parameters and audit'],
  ['editor', 'Editor — all website content and media'],
  ['sales', 'Sales — quotation requests'],
  ['hr', 'HR — job applications and CVs'],
] as const

interface User { id: string; email: string; name: string; role: string; active: boolean; lastLoginAt?: string; createdAt: string }

export function UsersScreen() {
  const { me } = useSession()
  const toast = useToast()
  const [q, setQ] = useState('')
  const dq = useDebounced(q)
  const [rows, setRows] = useState<User[] | null>(null)
  const [all, setAll] = useState(0)
  const [edit, setEdit] = useState<User | 'new' | null>(null)
  const url = '/api/v1/admin/users' + qs({ q: dq })
  const load = useCallback(() => { get<{ rows: User[] }>(url).then((r) => setRows(r.rows)).catch(() => setRows([])) }, [url])
  useEffect(load, [load])
  useEffect(() => { get<{ total: number }>('/api/v1/admin/users').then((r) => setAll(r.total)).catch(() => {}) }, [rows])
  return (
    <>
      <ScreenHeader group="Administration" title="Users" desc="Who can sign in to the CMS, and what each person can do. There is no self-service password reset: a superadmin sets a new one here.">
        <SearchBox value={q} onChange={setQ} />
        <ExportButton url={url} count={rows?.length || 0} onError={(m) => toast(m, true)} />
        <AddButton label="Add user" onClick={() => setEdit('new')} />
      </ScreenHeader>
      <main className="flex flex-col gap-3" style={{ padding: 'clamp(16px,3vw,28px) clamp(16px,3vw,32px) 48px' }}>
        <CountLine shown={rows?.length ?? 0} total={dq ? all : rows?.length ?? 0} reasons={[{ id: 'export-why', text: rows?.length === 0 ? 'Nothing to export: no users match the current search.' : undefined }]} />
        <div className="surface overflow-x-auto">
          <table className="grid-t min-w-[680px]">
            <thead><tr>{['Name', 'Email', 'Role', 'Status', 'Last sign-in', ''].map((h) => <th key={h} scope="col">{h}</th>)}</tr></thead>
            <tbody>
              {rows?.map((u) => (
                <tr key={u.id}>
                  <td>{u.name}{u.id === me.id && <span className="faint"> (you)</span>}</td><td><span className="ell">{u.email}</span></td><td>{u.role}</td>
                  <td>{u.active ? <StatusPill kind="ok" text="Active" /> : <StatusPill kind="draft" text="Deactivated" />}</td>
                  <td className="tnum">{fmtTime(u.lastLoginAt)}</td>
                  <td className="text-right"><button type="button" className="ib" aria-label={'Edit ' + u.name} onClick={() => setEdit(u)}><PencilSimple aria-hidden="true" /></button></td>
                </tr>
              ))}
            </tbody>
          </table>
          {rows?.length === 0 && <Empty />}
        </div>
      </main>
      {edit && <UserEditor u={edit === 'new' ? null : edit} self={me.id} onClose={() => setEdit(null)} onSaved={(m) => { toast(m); setEdit(null); load() }} />}
    </>
  )
}

function UserEditor({ u, self, onClose, onSaved }: { u: User | null; self: string; onClose: () => void; onSaved: (msg: string) => void }) {
  const [f, setF] = useState({ name: u?.name || '', email: u?.email || '', role: u?.role || 'editor', active: u?.active ?? true, password: '' })
  const [pw, setPw] = useState('')
  const [errs, setErrs] = useState<Record<string, string>>({})
  const [err, setErr] = useState('')
  const isSelf = u?.id === self
  const fail = (x: unknown) => {
    if (x instanceof ApiError) { setErrs(x.body.fields || {}); setErr(x.body.message) } else setErr('Save failed.')
  }
  const save = async () => {
    setErr('')
    try {
      if (u) await put('/api/v1/admin/users/' + u.id, { name: f.name, role: f.role, active: f.active })
      else await post('/api/v1/admin/users', { name: f.name, email: f.email, role: f.role, password: f.password })
      onSaved(u ? 'Saved ' + f.name : 'Created ' + f.email)
    } catch (x) { fail(x) }
  }
  const reset = async () => {
    setErr('')
    try {
      await post(`/api/v1/admin/users/${u!.id}/password`, { password: pw })
      onSaved('Password reset for ' + u!.name + '. Their sessions were ended.')
    } catch (x) { fail(x) }
  }
  const missing = [!f.name.trim() && 'Name', !u && !f.email.trim() && 'Email', !u && !f.password && 'Password'].filter(Boolean)
  return (
    <Drawer title={u ? u.name : 'New user'} kicker="Users" onClose={onClose} footer={<>
      <button type="button" className="b b-red" onClick={save} disabled={missing.length > 0} aria-describedby="u-why">Save</button>
      <button type="button" className="b b-line" onClick={onClose}>Cancel</button>
      <p id="u-why" className="muted m-0 flex-[1_1_100%] text-[13px]">{missing.length ? 'To save, fill in: ' + missing.join(', ') + '.' : 'Role changes take effect on the user’s next request.'}</p>
      {err && <p role="alert" className="ferr m-0 flex-[1_1_100%]">{err}</p>}
    </>}>
      <div className="flex flex-col gap-4">
        <label className="flex flex-col gap-1.5"><span className="lbl">Name *</span><input className="inp" value={f.name} maxLength={120} onChange={(e) => setF({ ...f, name: e.target.value })} aria-invalid={!!errs.name || undefined} />{errs.name && <span className="ferr">{errs.name}</span>}</label>
        <label className="flex flex-col gap-1.5"><span className="lbl">Email *</span><input className="inp" type="email" value={f.email} disabled={!!u} onChange={(e) => setF({ ...f, email: e.target.value })} aria-invalid={!!errs.email || undefined} />
          {u ? <span className="hint">The email is the sign-in name and cannot change. Create a new user instead.</span> : errs.email && <span className="ferr">{errs.email}</span>}</label>
        <fieldset className="m-0 border-0 p-0">
          <legend className="lbl mb-2">Role *</legend>
          {ROLES.map(([r, d]) => (
            <label key={r} className="flex cursor-pointer items-start gap-2 py-1 text-[15px]" style={{ opacity: isSelf && r !== 'superadmin' ? 0.5 : 1 }}>
              <input type="radio" name="role" value={r} checked={f.role === r} disabled={isSelf && r !== 'superadmin'} onChange={() => setF({ ...f, role: r })} className="mt-1 accent-charcoal" />{d}
            </label>
          ))}
          {isSelf && <p className="hint m-0">You cannot remove your own superadmin access. Ask another superadmin.</p>}
          {errs.role && <span className="ferr">{errs.role}</span>}
        </fieldset>
        {u && (
          <label className="flex items-center gap-2 text-[15px]" style={{ opacity: isSelf ? 0.5 : 1 }}>
            <input type="checkbox" checked={f.active} disabled={isSelf} onChange={(e) => setF({ ...f, active: e.target.checked })} className="h-5 w-5 accent-charcoal" />
            Active (can sign in)
          </label>
        )}
        {u && isSelf && <p className="hint m-0">You cannot deactivate yourself.</p>}
        {!u && (
          <label className="flex flex-col gap-1.5"><span className="lbl">Password *</span><input className="inp" type="password" autoComplete="new-password" value={f.password} onChange={(e) => setF({ ...f, password: e.target.value })} aria-invalid={!!errs.password || undefined} />
            {errs.password ? <span className="ferr">{errs.password}</span> : <span className="hint">At least the minimum length in Parameters → Sign-in security. Share it with the person directly.</span>}</label>
        )}
        {u && !isSelf && (
          <div className="bdr mt-2 flex flex-col gap-2 border-t pt-4">
            <span className="lbl">Reset password</span>
            <input className="inp" type="password" autoComplete="new-password" value={pw} onChange={(e) => setPw(e.target.value)} aria-label="New password" aria-invalid={!!errs.password || undefined} />
            {errs.password && <span className="ferr">{errs.password}</span>}
            <button type="button" className="b b-line self-start" disabled={!pw} onClick={reset} aria-describedby="pw-why">Set new password</button>
            <span id="pw-why" className="hint">{pw ? 'Their current sessions end immediately.' : 'Type a new password to enable this.'}</span>
          </div>
        )}
      </div>
    </Drawer>
  )
}

interface Audit { id: string; at: string; actorEmail: string; action: string; entity: string; entityId: string; summary: string; before: unknown; after: unknown; ip: string; requestId: string }

export function AuditScreen() {
  const toast = useToast()
  const [q, setQ] = useState('')
  const dq = useDebounced(q)
  const [action, setAction] = useState('')
  const [rows, setRows] = useState<Audit[] | null>(null)
  const [total, setTotal] = useState(0)
  const [open, setOpen] = useState<Audit | null>(null)
  const url = '/api/v1/admin/audit' + qs({ q: dq, 'f.action': action })
  useEffect(() => { get<{ rows: Audit[]; total: number }>(url + (url.includes('?') ? '&' : '?') + 'limit=500').then((r) => { setRows(r.rows); setTotal(r.total) }).catch(() => setRows([])) }, [url])
  return (
    <>
      <ScreenHeader group="Administration" title="Audit history" desc="Every change made in the CMS and by the scheduled jobs: who, when, from where, and what it was before. It cannot be edited or deleted.">
        <SearchBox value={q} onChange={setQ} />
        <select aria-label="Action" className="inp inp-auto" value={action} onChange={(e) => setAction(e.target.value)}>
          <option value="">All actions</option>
          {['create', 'update', 'delete', 'upload', 'cv_download', 'password_change', 'password_reset', 'lockout', 'setup'].map((a) => <option key={a} value={a}>{a}</option>)}
        </select>
        <ExportButton url={url} count={rows?.length || 0} onError={(m) => toast(m, true)} />
      </ScreenHeader>
      <main className="flex flex-col gap-3" style={{ padding: 'clamp(16px,3vw,28px) clamp(16px,3vw,32px) 48px' }}>
        <CountLine shown={rows?.length ?? 0} total={total} reasons={[{ id: 'export-why', text: rows?.length === 0 ? 'Nothing to export: no entries match.' : total > (rows?.length || 0) ? 'Showing the newest 500 on screen; the export holds every matching entry.' : undefined }]} />
        <div className="surface overflow-x-auto">
          <table className="grid-t min-w-[760px]">
            <thead><tr>{['When', 'Who', 'Action', 'What', 'Summary'].map((h) => <th key={h} scope="col">{h}</th>)}</tr></thead>
            <tbody>
              {rows?.map((e) => (
                <tr key={e.id} className="click" tabIndex={0} onClick={() => setOpen(e)} onKeyDown={(k) => k.key === 'Enter' && setOpen(e)}>
                  <td className="whitespace-nowrap tnum">{fmtTime(e.at)}</td><td><span className="ell">{e.actorEmail}</span></td><td>{e.action}</td><td>{e.entity}</td><td><span className="ell">{e.summary}</span></td>
                </tr>
              ))}
            </tbody>
          </table>
          {rows?.length === 0 && <Empty />}
        </div>
      </main>
      {open && (
        <Drawer title={open.summary || open.action} kicker="Audit entry" onClose={() => setOpen(null)}>
          <dl className="m-0 grid grid-cols-[110px_1fr] gap-2 text-[14px]">
            <dt className="faint">When</dt><dd className="m-0">{fmtTime(open.at)}</dd>
            <dt className="faint">Who</dt><dd className="m-0">{open.actorEmail}</dd>
            <dt className="faint">Action</dt><dd className="m-0">{open.action} · {open.entity} {open.entityId}</dd>
            <dt className="faint">From</dt><dd className="m-0">{open.ip || '—'}</dd>
            <dt className="faint">Request</dt><dd className="m-0 break-all">{open.requestId || '—'}</dd>
          </dl>
          <p className="lbl mb-1.5 mt-4">Before</p><pre className="json">{open.before ? JSON.stringify(open.before, null, 2) : '—'}</pre>
          <p className="lbl mb-1.5 mt-4">After</p><pre className="json">{open.after ? JSON.stringify(open.after, null, 2) : '—'}</pre>
        </Drawer>
      )}
    </>
  )
}

interface Mail { id: string; createdAt: string; kind: string; subject: string; recipients: string[]; attempts: number; sentAt: string | null; lastError: string }

export function OutboxScreen() {
  const toast = useToast()
  const [q, setQ] = useState('')
  const dq = useDebounced(q)
  const [rows, setRows] = useState<Mail[] | null>(null)
  const url = '/api/v1/admin/outbox' + qs({ q: dq })
  useEffect(() => { get<{ rows: Mail[] }>(url).then((r) => setRows(r.rows)).catch(() => setRows([])) }, [url])
  return (
    <>
      <ScreenHeader group="Administration" title="Outgoing mail" desc="Notification emails, newest first. Failed sends retry with back-off (1 min, 5 min, 30 min, 2 h, 6 h, then daily, up to 10 tries); the last error is shown here.">
        <SearchBox value={q} onChange={setQ} />
        <ExportButton url={url} count={rows?.length || 0} onError={(m) => toast(m, true)} />
      </ScreenHeader>
      <main className="flex flex-col gap-3" style={{ padding: 'clamp(16px,3vw,28px) clamp(16px,3vw,32px) 48px' }}>
        <CountLine shown={rows?.length ?? 0} total={rows?.length ?? 0} reasons={[{ id: 'export-why', text: rows?.length === 0 ? 'Nothing to export.' : undefined }]} />
        <div className="surface overflow-x-auto">
          <table className="grid-t min-w-[760px]">
            <thead><tr>{['Created', 'Kind', 'Subject', 'To', 'Status'].map((h) => <th key={h} scope="col">{h}</th>)}</tr></thead>
            <tbody>
              {rows?.map((m) => (
                <tr key={m.id}>
                  <td className="whitespace-nowrap tnum">{fmtTime(m.createdAt)}</td><td>{m.kind}</td><td><span className="ell">{m.subject}</span></td>
                  <td><span className="ell">{m.recipients.join(', ')}</span></td>
                  <td>{m.sentAt ? <StatusPill kind="ok" text={'Sent ' + fmtTime(m.sentAt)} /> : m.lastError ? <span title={m.lastError}><StatusPill kind="bad" text={`Failed ×${m.attempts}`} /></span> : <StatusPill kind="draft" text="Queued" />}</td>
                </tr>
              ))}
            </tbody>
          </table>
          {rows?.length === 0 && <Empty text="No mail yet." />}
        </div>
      </main>
    </>
  )
}

export function AccountScreen() {
  const { me, signOut } = useSession()
  const toast = useToast()
  const [cur, setCur] = useState('')
  const [pw, setPw] = useState('')
  const [pw2, setPw2] = useState('')
  const [err, setErr] = useState('')
  const submit = async (e: FormEvent) => {
    e.preventDefault()
    setErr('')
    try {
      await post('/api/v1/me/password', { current: cur, password: pw })
      toast('Password changed. Please sign in again.')
      signOut()
    } catch (x) {
      setErr(x instanceof ApiError ? x.body.message : 'Could not change the password.')
    }
  }
  const why = !cur ? 'Enter your current password.' : !pw ? 'Enter a new password.' : pw !== pw2 ? 'The two new passwords do not match.' : ''
  return (
    <>
      <ScreenHeader group="You" title="Your account" desc={`${me.name} · ${me.email} · ${me.role}`} />
      <main style={{ padding: 'clamp(16px,3vw,28px) clamp(16px,3vw,32px) 48px' }}>
        <form onSubmit={submit} className="surface flex max-w-[460px] flex-col gap-4 p-6" noValidate>
          <h2 className="m-0 text-[20px] font-[550]">Change password</h2>
          <label className="flex flex-col gap-1.5"><span className="lbl">Current password</span><input className="inp" type="password" autoComplete="current-password" value={cur} onChange={(e) => setCur(e.target.value)} /></label>
          <label className="flex flex-col gap-1.5"><span className="lbl">New password</span><input className="inp" type="password" autoComplete="new-password" value={pw} onChange={(e) => setPw(e.target.value)} /></label>
          <label className="flex flex-col gap-1.5"><span className="lbl">New password again</span><input className="inp" type="password" autoComplete="new-password" value={pw2} onChange={(e) => setPw2(e.target.value)} /></label>
          {err && <p role="alert" className="ferr m-0">{err}</p>}
          <button type="submit" className="b b-red self-start" disabled={!!why} aria-describedby="pw-why2">Change password</button>
          <p id="pw-why2" className="hint m-0">{why || 'All your sessions end and you sign in again.'}</p>
        </form>
      </main>
    </>
  )
}

