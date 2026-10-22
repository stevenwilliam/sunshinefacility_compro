// The CMS's API client. The access token lives in memory only (never
// localStorage, where any injected script could read it); the refresh token
// is an httpOnly cookie the browser sends to /api/v1/auth only.

export interface Me {
  id: string
  email: string
  name: string
  role: string
  permissions: string[]
}

export interface ApiErr {
  code: string
  message: string
  fields?: Record<string, string>
}

export class ApiError extends Error {
  constructor(public status: number, public body: ApiErr) {
    super(body.message || 'Request failed')
  }
}

let token = ''
let refreshing: Promise<Me | null> | null = null
let onSignedOut: () => void = () => {}

export function setSignedOutHandler(fn: () => void) {
  onSignedOut = fn
}

export function setToken(t: string) {
  token = t
}

/** Exchange the refresh cookie for a fresh access token. One at a time. */
export function refresh(): Promise<Me | null> {
  if (!refreshing) {
    refreshing = fetch('/api/v1/auth/refresh', { method: 'POST', credentials: 'same-origin' })
      .then(async (r) => {
        if (!r.ok) return null
        const j = await r.json()
        token = j.accessToken
        return j.user as Me
      })
      .catch(() => null)
      .finally(() => {
        refreshing = null
      })
  }
  return refreshing
}

async function raw(method: string, url: string, body?: unknown, retry = true): Promise<Response> {
  const headers: Record<string, string> = {}
  if (token) headers.Authorization = 'Bearer ' + token
  let payload: BodyInit | undefined
  if (body instanceof FormData) payload = body
  else if (body !== undefined) {
    headers['Content-Type'] = 'application/json'
    payload = JSON.stringify(body)
  }
  const r = await fetch(url, { method, headers, body: payload, credentials: 'same-origin' })
  if (r.status === 401 && retry) {
    const me = await refresh()
    if (me) return raw(method, url, body, false)
    onSignedOut()
  }
  return r
}

export async function api<T>(method: string, url: string, body?: unknown): Promise<T> {
  const r = await raw(method, url, body)
  const j = await r.json().catch(() => ({}))
  if (!r.ok) throw new ApiError(r.status, (j as { error?: ApiErr }).error || { code: 'INTERNAL', message: 'Request failed' })
  return j as T
}

export const get = <T,>(url: string) => api<T>('GET', url)
export const post = <T,>(url: string, body?: unknown) => api<T>('POST', url, body)
export const put = <T,>(url: string, body?: unknown) => api<T>('PUT', url, body)
export const del = <T,>(url: string) => api<T>('DELETE', url)

/** Download a pipe-delimited CSV export with the auth header (BR-9.2). */
export async function downloadCSV(url: string) {
  const r = await raw('GET', url + (url.includes('?') ? '&' : '?') + 'format=csv')
  if (!r.ok) throw new ApiError(r.status, { code: 'INTERNAL', message: 'Export failed' })
  const cd = r.headers.get('Content-Disposition') || ''
  const m = /filename="([^"]+)"/.exec(cd)
  const blob = await r.blob()
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = m ? m[1] : 'export.csv'
  document.body.appendChild(a)
  a.click()
  a.remove()
  setTimeout(() => URL.revokeObjectURL(a.href), 1000)
}

export function qs(params: Record<string, string | undefined>) {
  const u = new URLSearchParams()
  for (const [k, v] of Object.entries(params)) if (v) u.set(k, v)
  const s = u.toString()
  return s ? '?' + s : ''
}

export const can = (me: Me | null, p: string) => !!me && me.permissions.includes(p)

/** Date-times in Asia/Jakarta, the operating timezone. */
export function fmtTime(iso?: string | null) {
  if (!iso) return '—'
  const d = new Date(iso)
  if (isNaN(d.getTime())) return iso
  return d.toLocaleString('en-GB', { timeZone: 'Asia/Jakarta', day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' })
}

export function fmtDate(s?: string | null) {
  if (!s) return '—'
  const d = new Date(s.length === 10 ? s + 'T00:00:00' : s)
  if (isNaN(d.getTime())) return s
  return d.toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric', timeZone: s.length === 10 ? undefined : 'Asia/Jakarta' })
}

/** Today in Asia/Jakarta as YYYY-MM-DD (BR-3.1). */
export function todayJakarta() {
  return new Intl.DateTimeFormat('en-CA', { timeZone: 'Asia/Jakarta', year: 'numeric', month: '2-digit', day: '2-digit' }).format(new Date())
}
