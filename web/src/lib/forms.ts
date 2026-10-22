// Client-side form plumbing. The browser validates for FEEDBACK; the server
// validates again because the browser can be bypassed (CLAUDE.md §7). The
// limits come from the server in the public bundle (bundle.rules), so the two
// sides read one source.

export const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

export interface ApiError {
  code: string
  message: string
  fields?: Record<string, string>
}

export class SubmitError extends Error {
  constructor(public status: number, public body: ApiError) {
    super(body.message)
  }
}

export async function postJSON<T>(url: string, body: unknown): Promise<T> {
  const r = await fetch(url, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) })
  const j = await r.json().catch(() => ({}))
  if (!r.ok) throw new SubmitError(r.status, (j as { error?: ApiError }).error || { code: 'INTERNAL', message: '' })
  return j as T
}

export async function postForm<T>(url: string, fd: FormData): Promise<T> {
  const r = await fetch(url, { method: 'POST', body: fd })
  const j = await r.json().catch(() => ({}))
  if (!r.ok) throw new SubmitError(r.status, (j as { error?: ApiError }).error || { code: 'INTERNAL', message: '' })
  return j as T
}

/** The form-level message for a failed send, in the visitor's language. */
export function formError(err: unknown, t: (k: string) => string): string {
  if (err instanceof SubmitError) {
    if (err.body.code === 'RATE_LIMITED') return t('form.error.rate')
    if (err.body.code === 'VALIDATION') {
      const f = err.body.fields || {}
      if (f._form === 'too_fast') return t('form.error.too_fast')
      if (f._form) return t('form.error.invalid')
      return t('form.error.check')
    }
  }
  return t('form.error.generic')
}

/** Field errors from the server, rendered from the catalogue (err.<code>). */
export function fieldErrors(err: unknown, t: (k: string, v?: Record<string, string | number>) => string, mb: number): Record<string, string> {
  if (!(err instanceof SubmitError) || err.body.code !== 'VALIDATION') return {}
  const out: Record<string, string> = {}
  for (const [k, code] of Object.entries(err.body.fields || {})) {
    if (k.startsWith('_')) continue
    out[k] = t('err.' + code, { mb })
  }
  return out
}

/** The honeypot: a field people never see or reach (BR-5.4). */
export const HONEYPOT_STYLE: React.CSSProperties = { position: 'absolute', left: '-10000px', width: 1, height: 1, overflow: 'hidden' }
