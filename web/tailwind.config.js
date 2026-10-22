/** @type {import('tailwindcss').Config} */
// Brand tokens from the SFG brand guide (README → Design tokens). The measured
// contrast of every text pairing is in .claude/skills/impeccable/design.md.
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    // Breakpoints are the design's own: burger below 1080, header CTA hidden
    // below 640, hero side-by-side from 960, "mobile" layouts below 760.
    screens: { sm: '640px', md: '760px', hero: '960px', nav: '1080px', xl: '1280px' },
    extend: {
      fontFamily: { sans: ['Nohemi', 'system-ui', 'sans-serif'] },
      colors: {
        charcoal: { DEFAULT: '#1F211C', deep: '#161813', 800: '#2A2C26', 700: '#33352E', 600: '#3A3C35', 500: '#44463E', 400: '#4E5047' },
        cotton: '#F4F4F1',
        line: { 1: '#EFEFEA', 2: '#E6E6E1', 3: '#D6D6CF', 4: '#C4C4BC' },
        ink: { 2: '#45473F', 3: '#5A5C54' },
        ondark: { body: '#D6D7CE', muted: '#C9CABF', bright: '#EDEDE6', dim: '#A8A99F' },
        red: { DEFAULT: '#C6241C', hover: '#A3181A', signal: '#EA2915' },
        tangerine: '#FD572A',
        disabled: '#8A8B83',
        wa: { DEFAULT: '#1E9E55', hover: '#178A49', card: '#1E7A44', cardHover: '#17663A' },
        success: { DEFAULT: '#4A7023', bg: '#E3F0E5', fg: '#1F6B34' },
      },
      maxWidth: { page: '1280px' },
      boxShadow: {
        card: '0 10px 24px -10px rgba(31,33,28,.35)',
        modal: '0 30px 60px -20px rgba(0,0,0,.6)',
      },
    },
  },
  plugins: [],
}
