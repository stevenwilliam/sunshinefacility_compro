import { createContext, useCallback, useContext, useEffect, useState, type FormEvent, type ReactNode } from 'react'
import { Link, NavLink, Navigate, Route, Routes, useLocation, useNavigate } from 'react-router-dom'
import './admin.css'
import { ApiError, can, get, post, refresh, setSignedOutHandler, setToken, type Me } from './api'
import { ToastProvider } from './ui'
import ResourceScreen, { type Schema } from './ResourceScreen'
import MediaScreen from './MediaScreen'
import { ApplicationsScreen, QuotationsScreen } from './Inbox'
import { AccountScreen, AuditScreen, OutboxScreen, UsersScreen } from './Admin'
import { ArrowUpRight, SignOut } from '@/lib/phosphor'

interface Session {
  me: Me
  schema: Schema | null
  placeholders: Record<string, number>
  reloadPH: () => void
  signOut: () => void
}

const Ctx = createContext<Session | null>(null)
export const useSession = () => {
  const s = useContext(Ctx)
  if (!s) throw new Error('no session')
  return s
}

export default function AdminApp() {
  const [me, setMe] = useState<Me | null | undefined>(undefined)
  const [setupNeeded, setSetupNeeded] = useState(false)
  useEffect(() => {
    document.title = 'Admin · Sunshine Facilities Group'
    setSignedOutHandler(() => setMe(null))
    refresh().then(async (m) => {
      if (m) return setMe(m)
      const s = await get<{ setupNeeded: boolean }>('/api/v1/auth/setup').catch(() => ({ setupNeeded: false }))
      setSetupNeeded(s.setupNeeded)
      setMe(null)
    })
  }, [])
  return (
    <div className="cms">
      <ToastProvider>
        {me === undefined ? (
          <p className="muted p-8">Loading…</p>
        ) : me === null ? (
          setupNeeded ? <Setup onDone={setMe} /> : <Login onDone={setMe} />
        ) : (
          <Shell me={me} onSignedOut={() => setMe(null)} />
        )}
      </ToastProvider>
    </div>
  )
}

function AuthCard({ title, children }: { title: string; children: ReactNode }) {
  return (
    <main className="grid min-h-screen place-items-center p-4">
      <div className="surface w-[min(420px,100%)] p-8">
        <div className="mb-6 flex items-center gap-3">
          <span className="block h-[30px] w-[120px] bg-contain bg-left bg-no-repeat" style={{ backgroundImage: 'url("/favicon.svg")', width: 30 }} aria-hidden="true" />
          <span className="text-[13px] font-semibold uppercase tracking-[0.08em] faint">Content admin</span>
        </div>
        <h1 className="mb-6 mt-0 text-[26px] font-[550] tracking-[-0.02em]">{title}</h1>
        {children}
      </div>
    </main>
  )
}

function Login({ onDone }: { onDone: (m: Me) => void }) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)
  const submit = async (e: FormEvent) => {
    e.preventDefault()
    setBusy(true)
    setErr('')
    try {
      const r = await post<{ accessToken: string; user: Me }>('/api/v1/auth/login', { email, password })
      setToken(r.accessToken)
      onDone(r.user)
    } catch (x) {
      setErr(x instanceof ApiError ? x.body.message : 'Sign-in failed. Please try again.')
    } finally {
      setBusy(false)
    }
  }
  const missing = !email.trim() || !password
  return (
    <AuthCard title="Sign in">
      <form onSubmit={submit} className="flex flex-col gap-4" noValidate>
        <label className="flex flex-col gap-1.5"><span className="lbl">Email</span>
          <input className="inp" type="email" autoComplete="username" value={email} onChange={(e) => setEmail(e.target.value)} required /></label>
        <label className="flex flex-col gap-1.5"><span className="lbl">Password</span>
          <input className="inp" type="password" autoComplete="current-password" value={password} onChange={(e) => setPassword(e.target.value)} required /></label>
        {err && <p role="alert" className="ferr m-0">{err}</p>}
        <button type="submit" className="b b-red justify-center" disabled={missing || busy} aria-describedby="login-why">{busy ? 'Signing in…' : 'Sign in'}</button>
        <p id="login-why" className="hint m-0">{missing ? 'Enter your email and password to sign in.' : 'Forgotten your password? Ask a superadmin to reset it.'}</p>
      </form>
    </AuthCard>
  )
}

// BR-8.2: the first superadmin, once, with the setup token from the server's
// configuration (/etc/sfg/sfg.env → SFG_SETUP_TOKEN).
function Setup({ onDone }: { onDone: (m: Me) => void }) {
  const [f, setF] = useState({ token: '', name: '', email: '', password: '' })
  const [errs, setErrs] = useState<Record<string, string>>({})
  const [err, setErr] = useState('')
  const set = (k: keyof typeof f) => (e: { target: { value: string } }) => setF((s) => ({ ...s, [k]: e.target.value }))
  const submit = async (e: FormEvent) => {
    e.preventDefault()
    setErr('')
    try {
      const r = await post<{ accessToken: string; user: Me }>('/api/v1/auth/setup', f)
      setToken(r.accessToken)
      onDone(r.user)
    } catch (x) {
      if (x instanceof ApiError) {
        setErrs(x.body.fields || {})
        setErr(x.body.message)
      } else setErr('Setup failed.')
    }
  }
  const missing = Object.entries(f).filter(([, v]) => !v.trim()).map(([k]) => k)
  return (
    <AuthCard title="Create the first superadmin">
      <form onSubmit={submit} className="flex flex-col gap-4" noValidate>
        {(['token', 'name', 'email', 'password'] as const).map((k) => (
          <label key={k} className="flex flex-col gap-1.5">
            <span className="lbl">{{ token: 'Setup token', name: 'Your name', email: 'Email', password: 'Password' }[k]} *</span>
            <input className="inp" type={k === 'password' || k === 'token' ? 'password' : k === 'email' ? 'email' : 'text'} value={f[k]} onChange={set(k)} aria-invalid={!!errs[k] || undefined} />
            {errs[k] && <span className="ferr">{errs[k]}</span>}
            {k === 'token' && <span className="hint">From SFG_SETUP_TOKEN in /etc/sfg/sfg.env on the server.</span>}
          </label>
        ))}
        {err && <p role="alert" className="ferr m-0">{err}</p>}
        <button type="submit" className="b b-red justify-center" disabled={missing.length > 0}>Create account</button>
        {missing.length > 0 && <p className="hint m-0">To continue, fill in: {missing.join(', ')}.</p>}
      </form>
    </AuthCard>
  )
}

interface NavItem { to: string; label: string; ph?: string; perm: string[] }

function Shell({ me, onSignedOut }: { me: Me; onSignedOut: () => void }) {
  const [schema, setSchema] = useState<Schema | null>(null)
  const [ph, setPh] = useState<Record<string, number>>({})
  const loc = useLocation()
  const nav = useNavigate()
  const reloadPH = useCallback(() => {
    if (can(me, 'content.read')) get<Record<string, number>>('/api/v1/admin/placeholders').then(setPh).catch(() => {})
  }, [me])
  useEffect(() => {
    if (can(me, 'content.read') || can(me, 'params.content')) get<Schema>('/api/v1/admin/schema').then(setSchema).catch(() => {})
    reloadPH()
  }, [me, reloadPH])
  const signOut = async () => {
    await post('/api/v1/auth/logout').catch(() => {})
    setToken('')
    onSignedOut()
  }

  const groups: { title: string; items: NavItem[] }[] = []
  if (schema) {
    const byNav: Record<string, NavItem[]> = {}
    for (const d of schema.resources) {
      ;(byNav[d.nav] ||= []).push({ to: `/admin/r/${d.name}`, label: d.label, ph: d.name, perm: ['content.read', 'params.content'] })
    }
    if (byNav.Content) groups.push({ title: 'Content', items: [...byNav.Content, { to: '/admin/media', label: 'Media library', perm: ['content.read'] }] })
    if (byNav.Services) groups.push({ title: 'Services', items: byNav.Services })
    if (byNav.Settings) groups.push({ title: 'Settings', items: byNav.Settings })
  }
  const inbox: NavItem[] = [
    { to: '/admin/inbox/quotations', label: 'Quotation requests', perm: ['quotations.read'] },
    { to: '/admin/inbox/applications', label: 'Job applications', perm: ['applications.read'] },
  ].filter((i) => i.perm.some((p) => can(me, p)))
  if (inbox.length) groups.unshift({ title: 'Inbox', items: inbox })
  const admin: NavItem[] = [
    { to: '/admin/users', label: 'Users', perm: ['users.manage'] },
    { to: '/admin/audit', label: 'Audit history', perm: ['audit.read'] },
    { to: '/admin/outbox', label: 'Outgoing mail', perm: ['audit.read'] },
  ].filter((i) => i.perm.some((p) => can(me, p)))
  if (admin.length) groups.push({ title: 'Administration', items: admin })
  groups.push({ title: 'You', items: [{ to: '/admin/account', label: 'Your account', perm: [] }] })
  const flat = groups.flatMap((g) => g.items)
  const first = flat[0]?.to || '/admin/account'
  const totalPH = Object.values(ph).reduce((a, b) => a + b, 0)

  return (
    <Ctx.Provider value={{ me, schema, placeholders: ph, reloadPH, signOut }}>
      <div className="flex min-h-screen">
        <aside className="side sticky top-0 hidden h-screen w-[248px] flex-none flex-col hero:flex">
          <div className="flex flex-col gap-2.5 border-b border-charcoal-700 px-[22px] pb-[18px] pt-[22px]">
            <span className="text-[17px] font-semibold tracking-[-0.01em] text-white">Sunshine Facilities Group</span>
            <span className="text-[12px] font-semibold uppercase tracking-[0.08em] text-ondark-dim">Content admin</span>
          </div>
          <nav aria-label="Admin" className="flex flex-1 flex-col gap-0.5 overflow-y-auto p-3">
            {groups.map((g) => (
              <div key={g.title}>
                <p className="grp">{g.title}</p>
                {g.items.map((n) => (
                  <NavLink key={n.to} to={n.to} className="nav">
                    <span className="flex-1">{n.label}</span>
                    {n.ph && ph[n.ph] > 0 && <span className="badge-ph" title={`${ph[n.ph]} placeholder(s) to replace`}>{ph[n.ph]}</span>}
                  </NavLink>
                ))}
              </div>
            ))}
          </nav>
          <div className="flex flex-col gap-2 p-3">
            {totalPH > 0 && <p className="m-0 px-1 text-[12px] text-ondark-muted">{totalPH} placeholder{totalPH === 1 ? '' : 's'} still to replace before launch.</p>}
            <Link to="/" target="_blank" className="flex min-h-[44px] items-center justify-between gap-2 rounded border border-charcoal-500 px-[14px] text-[14px] text-ondark-bright hover:border-tangerine hover:text-white">View site <ArrowUpRight aria-hidden="true" /></Link>
            <button type="button" onClick={signOut} className="flex min-h-[44px] cursor-pointer items-center justify-between gap-2 rounded border border-charcoal-500 bg-transparent px-[14px] text-[14px] text-ondark-bright hover:border-tangerine">
              <span className="truncate">Sign out · {me.name}</span><SignOut aria-hidden="true" />
            </button>
          </div>
        </aside>
        <div className="flex min-w-0 flex-1 flex-col">
          <div className="side flex flex-wrap items-center gap-3 px-4 py-3 hero:hidden">
            <span className="text-[15px] font-semibold text-white">SFG admin</span>
            <label className="ml-auto flex items-center gap-2 text-[13px] text-ondark-muted">Section
              <select value={flat.find((n) => loc.pathname.startsWith(n.to))?.to || ''} onChange={(e) => nav(e.target.value)}
                className="h-10 rounded border border-charcoal-500 bg-charcoal-800 px-2.5 text-[15px] text-white">
                {groups.map((g) => (
                  <optgroup key={g.title} label={g.title}>
                    {g.items.map((n) => <option key={n.to} value={n.to}>{n.label}</option>)}
                  </optgroup>
                ))}
              </select>
            </label>
            <button type="button" onClick={signOut} className="ib" aria-label="Sign out"><SignOut aria-hidden="true" /></button>
          </div>
          <Routes>
            <Route path="/admin" element={<Navigate to={first} replace />} />
            <Route path="/admin/r/:name" element={<ResourceScreen />} />
            <Route path="/admin/media" element={<MediaScreen />} />
            <Route path="/admin/inbox/quotations/:id?" element={<QuotationsScreen />} />
            <Route path="/admin/inbox/applications/:id?" element={<ApplicationsScreen />} />
            <Route path="/admin/users" element={<UsersScreen />} />
            <Route path="/admin/audit" element={<AuditScreen />} />
            <Route path="/admin/outbox" element={<OutboxScreen />} />
            <Route path="/admin/account" element={<AccountScreen />} />
            <Route path="*" element={<Navigate to={first} replace />} />
          </Routes>
        </div>
      </div>
    </Ctx.Provider>
  )
}
