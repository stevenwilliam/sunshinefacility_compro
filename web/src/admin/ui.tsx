import { createContext, useCallback, useContext, useEffect, useRef, useState, type ReactNode } from 'react'
import { CheckCircle, DownloadSimple, MagnifyingGlass, Plus, Warning } from '@/lib/phosphor'
import { downloadCSV } from './api'

// ---------------------------------------------------------------------------
// Toast
// ---------------------------------------------------------------------------

const ToastCtx = createContext<(msg: string, bad?: boolean) => void>(() => {})

export function ToastProvider({ children }: { children: ReactNode }) {
  const [t, setT] = useState<{ msg: string; bad: boolean } | null>(null)
  const timer = useRef<number>()
  const show = useCallback((msg: string, bad = false) => {
    setT({ msg, bad })
    window.clearTimeout(timer.current)
    timer.current = window.setTimeout(() => setT(null), bad ? 5000 : 2600)
  }, [])
  return (
    <ToastCtx.Provider value={show}>
      {children}
      {t && (
        <div role={t.bad ? 'alert' : 'status'} className="toast">
          {t.bad ? <Warning weight="fill" aria-hidden="true" style={{ color: '#FD572A' }} /> : <CheckCircle weight="fill" aria-hidden="true" style={{ color: '#7BC47F' }} />}
          {t.msg}
        </div>
      )}
    </ToastCtx.Provider>
  )
}

export const useToast = () => useContext(ToastCtx)

// ---------------------------------------------------------------------------
// Debounced search (CLAUDE.md §8: every list has a debounced search box)
// ---------------------------------------------------------------------------

export function useDebounced<T>(v: T, ms = 250): T {
  const [d, setD] = useState(v)
  useEffect(() => {
    const id = window.setTimeout(() => setD(v), ms)
    return () => window.clearTimeout(id)
  }, [v, ms])
  return d
}

export function SearchBox({ value, onChange, label = 'Search' }: { value: string; onChange: (v: string) => void; label?: string }) {
  return (
    <span className="relative block min-w-0 flex-[1_1_220px]">
      <MagnifyingGlass aria-hidden="true" className="faint absolute left-3 top-1/2 -translate-y-1/2 text-[18px]" />
      <input type="search" aria-label={label} placeholder={label} value={value} onChange={(e) => onChange(e.target.value)} className="inp inp-search" />
    </span>
  )
}

// ---------------------------------------------------------------------------
// Screen header and toolbar
// ---------------------------------------------------------------------------

export function ScreenHeader({ group, title, desc, children }: { group: string; title: string; desc?: string; children?: ReactNode }) {
  return (
    <header className="bdr flex flex-wrap items-end justify-between gap-x-6 gap-y-4 border-b" style={{ background: 'var(--a-surface)', padding: 'clamp(16px,3vw,28px) clamp(16px,3vw,32px)' }}>
      <div className="min-w-0 max-w-[560px]">
        <p className="faint mb-1.5 mt-0 text-[13px]">{group}</p>
        <h1 className="m-0 font-[550] tracking-[-0.02em]" style={{ fontSize: 'clamp(26px,3vw,34px)' }}>{title}</h1>
        {desc && <p className="muted mb-0 mt-2 text-[14px] leading-[1.5]">{desc}</p>}
      </div>
      <div className="flex w-full flex-wrap items-center gap-2 md:w-auto">{children}</div>
    </header>
  )
}

/** Export CSV, disabled with its reason when there is nothing to export. */
export function ExportButton({ url, count, onError }: { url: string; count: number; onError?: (m: string) => void }) {
  const [busy, setBusy] = useState(false)
  const why = count === 0 ? 'Nothing to export: no rows match the current search.' : 'Download these rows as a pipe-delimited (|) CSV'
  return (
    <button type="button" className="b b-line" disabled={count === 0 || busy} title={why} aria-describedby={count === 0 ? 'export-why' : undefined}
      onClick={async () => {
        setBusy(true)
        try {
          await downloadCSV(url)
        } catch {
          onError?.('The export failed. Please try again.')
        } finally {
          setBusy(false)
        }
      }}>
      <DownloadSimple aria-hidden="true" className="text-[18px]" />Export CSV
    </button>
  )
}

export function AddButton({ label, onClick, disabledWhy }: { label: string; onClick: () => void; disabledWhy?: string }) {
  return (
    <button type="button" className="b b-red" onClick={onClick} disabled={!!disabledWhy} title={disabledWhy || undefined}
      aria-describedby={disabledWhy ? 'add-why' : undefined}>
      <Plus aria-hidden="true" className="text-[18px]" />{label}
    </button>
  )
}

/** "Showing n of m" plus the visible reasons for any disabled toolbar control. */
export function CountLine({ shown, total, reasons }: { shown: number; total: number; reasons?: { id: string; text?: string }[] }) {
  return (
    <div className="muted flex flex-wrap justify-between gap-2 text-[14px]">
      <p aria-live="polite" className="m-0">{shown === total ? `${total} item${total === 1 ? '' : 's'}` : `Showing ${shown} of ${total}`}</p>
      <div className="flex flex-col items-end gap-0.5">
        {(reasons || []).filter((r) => r.text).map((r) => <p key={r.id} id={r.id} className="faint m-0 text-[13px]">{r.text}</p>)}
      </div>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Confirm
// ---------------------------------------------------------------------------

export function Confirm({ title, body = 'This cannot be undone.', action = 'Delete', onConfirm, onCancel }: {
  title: string; body?: string; action?: string; onConfirm: () => void; onCancel: () => void
}) {
  const ref = useRef<HTMLButtonElement>(null)
  useEffect(() => {
    ref.current?.focus()
    const k = (e: KeyboardEvent) => e.key === 'Escape' && onCancel()
    window.addEventListener('keydown', k)
    return () => window.removeEventListener('keydown', k)
  }, [onCancel])
  return (
    <div className="modal">
      <div role="alertdialog" aria-modal="true" aria-labelledby="cf-h" aria-describedby="cf-b" className="flex w-[min(440px,100%)] flex-col gap-3 rounded p-6 shadow-modal" style={{ background: 'var(--a-surface)' }}>
        <h2 id="cf-h" className="m-0 text-[20px] font-[550]">{title}</h2>
        <p id="cf-b" className="muted m-0 text-[15px] leading-[1.5]">{body}</p>
        <div className="mt-2 flex flex-wrap gap-[10px]">
          <button type="button" onClick={onConfirm} className="b border-0 font-[550] text-white" style={{ background: '#A3181A' }}>{action}</button>
          <button ref={ref} type="button" onClick={onCancel} className="b b-line">Cancel</button>
        </div>
      </div>
    </div>
  )
}

export function StatusPill({ kind, text }: { kind: 'ok' | 'draft' | 'bad' | 'ph'; text: string }) {
  return <span className={`pill ${kind}`}>{text}</span>
}

export function Empty({ text = 'Nothing matches your search.' }: { text?: string }) {
  return <p className="muted m-0 px-5 py-10 text-center text-[15px]">{text}</p>
}

/** A side drawer with focus management and Esc to close. */
export function Drawer({ title, kicker, onClose, children, footer }: { title: string; kicker?: string; onClose: () => void; children: ReactNode; footer?: ReactNode }) {
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    const prev = document.activeElement as HTMLElement | null
    ref.current?.querySelector<HTMLElement>('input,select,textarea,button')?.focus()
    const k = (e: KeyboardEvent) => e.key === 'Escape' && !document.querySelector('.modal') && onClose()
    window.addEventListener('keydown', k)
    return () => {
      window.removeEventListener('keydown', k)
      prev?.focus()
    }
  }, [onClose])
  return (
    <>
      <div className="scrim" onClick={onClose} />
      <aside ref={ref} role="dialog" aria-modal="true" aria-labelledby="dr-h" className="drawer">
        <header className="bdr flex items-center justify-between gap-3 border-b px-[22px] py-[18px]">
          <div>
            {kicker && <p className="faint m-0 text-[13px]">{kicker}</p>}
            <h2 id="dr-h" className="m-0 mt-0.5 text-[22px] font-[550] tracking-[-0.015em]">{title}</h2>
          </div>
          <button type="button" onClick={onClose} aria-label="Close" className="ib h-11 w-11 text-[20px]">×</button>
        </header>
        <div className="flex-1 overflow-y-auto p-[22px]">{children}</div>
        {footer && <footer className="bdr flex flex-wrap items-center gap-[10px] border-t px-[22px] py-4">{footer}</footer>}
      </aside>
    </>
  )
}
