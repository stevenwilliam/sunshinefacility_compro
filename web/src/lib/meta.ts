import { useEffect } from 'react'

// Per-route <title> and description for client-side navigation. The server
// already wrote them into the first HTML (for bots and link previews); this
// keeps them right as the visitor moves around and switches language.
export function usePageMeta(title: string, description?: string) {
  useEffect(() => {
    if (title) document.title = title
    if (description !== undefined) {
      let m = document.querySelector('meta[name="description"]')
      if (!m) {
        m = document.createElement('meta')
        m.setAttribute('name', 'description')
        document.head.appendChild(m)
      }
      m.setAttribute('content', description)
    }
  }, [title, description])
}
