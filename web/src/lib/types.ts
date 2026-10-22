// The public bundle, as the Go server builds it (internal/app/public.go).
// Every bilingual value is a [en, id] pair.

export type L2 = [string, string]
export type Lang = 'en' | 'id'

export interface Img {
  src: string
  srcset?: string
  w?: number
  h?: number
  alt: L2
  svg?: boolean
}

export interface Section {
  title: L2
  body: L2
  image?: Img
  published: boolean
  placeholder: boolean
}

export interface ParamValue {
  v: L2
  placeholder?: boolean
}

export interface OptionV {
  key: string
  label: L2
}

export interface Line {
  id: string
  key: string
  n: string
  title: L2
  short: L2
  label: L2
  summary: L2
  description: L2
  lead: L2
  base: string
  ink: string
  shade: string
  onInk: string
  labelColor: string
  mark?: Img
  photo?: Img
  map?: Img
  showMap: boolean
  showCerts: boolean
  cardTags: L2[]
  serviceTags: L2[]
  heroPoints: L2[]
  facts: { value: L2; label: L2 }[]
  groups: { title: L2; body: L2; items: [string[], string[]] }[]
  industries: string[]
}

export interface Industry {
  key: string
  icon: string
  title: L2
  body: L2
  lines: string[]
  clients: number[]
}

export interface Client {
  name: string
  logo?: Img
  industry: string
}

export interface Certificate {
  name: string
  issuer: string
  validUntil: string
  description: L2
  image?: Img
  full: string
}

export interface ListItem {
  icon?: string
  title: L2
  body: L2
  meta: L2
  color?: string
  placeholder?: boolean
}

export interface Opening {
  id: string
  title: L2
  line: string
  location: string
  type: string
  closingDate?: string
  summary: L2
  requirements: [string[], string[]]
  placeholder?: boolean
}

export interface Testimonial {
  quote: L2
  name: string
  role: string
  company: string
  placeholder?: boolean
}

export interface Person {
  name: string
  role: L2
  photo?: Img
  placeholder?: boolean
}

export interface Rules {
  limits: { name: number; company: number; email: number; phone: number; location: number; message: number; areaMax: number }
  cvMaxMb: number
  minFillSeconds: number
}

export interface Bundle {
  version: string
  today: string
  baseUrl: string
  text: Record<string, L2>
  textPlaceholders: string[]
  sections: Record<string, Section>
  params: Record<string, ParamValue>
  buildingTypes: OptionV[]
  locations: OptionV[]
  jobTypes: OptionV[]
  logos: Record<string, Img>
  lines: Line[]
  industries: Industry[]
  clients: Client[]
  certificates: Certificate[]
  lists: Record<string, ListItem[]>
  stats: Record<string, { value: L2; label: L2 }[]>
  openings: Opening[]
  testimonials: Testimonial[]
  testimonialPlaceholders: Testimonial[]
  management: Person[]
  rules: Rules
  ga4: string
  ogImage: string
}
