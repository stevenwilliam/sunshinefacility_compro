import { StrictMode, Suspense, lazy, useEffect } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Route, Routes, useLocation } from 'react-router-dom'
import './index.css'
import { loadBundle, SiteProvider } from './lib/site'
import { SiteFooter, SiteHeader } from './components/Chrome'
import { Consent } from './components/Consent'
import Home from './pages/Home'
import Services from './pages/Services'
import ServiceDetail from './pages/ServiceDetail'
import Industries from './pages/Industries'
import About from './pages/About'
import Careers from './pages/Careers'
import Contact from './pages/Contact'
import NotFound from './pages/NotFound'

const Admin = lazy(() => import('./admin/AdminApp'))

function ScrollManager() {
  const { pathname, hash } = useLocation()
  useEffect(() => {
    if (hash) {
      const el = document.getElementById(decodeURIComponent(hash.slice(1)))
      if (el) {
        el.scrollIntoView()
        return
      }
    }
    window.scrollTo(0, 0)
  }, [pathname, hash])
  return null
}

function PublicSite() {
  return (
    <div className="min-h-screen bg-white font-sans text-charcoal">
      <SiteHeader />
      <main id="main" tabIndex={-1} className="outline-none">
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/services" element={<Services />} />
          <Route path="/services/:line" element={<ServiceDetail />} />
          <Route path="/industries" element={<Industries />} />
          <Route path="/about" element={<About />} />
          <Route path="/careers" element={<Careers />} />
          <Route path="/contact" element={<Contact />} />
          <Route path="*" element={<NotFound />} />
        </Routes>
      </main>
      <SiteFooter />
      <Consent />
    </div>
  )
}

const root = createRoot(document.getElementById('root')!)

if (window.location.pathname === '/admin' || window.location.pathname.startsWith('/admin/')) {
  root.render(
    <StrictMode>
      <BrowserRouter>
        <Suspense fallback={null}>
          <Admin />
        </Suspense>
      </BrowserRouter>
    </StrictMode>,
  )
} else {
  loadBundle()
    .then((bundle) => {
      root.render(
        <StrictMode>
          <SiteProvider bundle={bundle}>
            <BrowserRouter>
              <ScrollManager />
              <PublicSite />
            </BrowserRouter>
          </SiteProvider>
        </StrictMode>,
      )
    })
    .catch(() => {
      document.getElementById('root')!.textContent = 'Sunshine Facilities Group — the site could not load. Please refresh.'
    })
}
