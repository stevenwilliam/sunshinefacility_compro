// The signature ribbons, transcribed from patterns/signature-ribbon.html
// as-is: same paths, gradients, stroke widths, opacities, dash arrays and
// animation classes (the CSS is in index.css under "Signature ribbon").
// Drop one inside a section with position:relative and overflow:hidden.

const VEN =
  'M -80 900 C 120 840 220 660 300 620 C 330 710 360 820 400 860 C 440 820 480 680 520 600 C 540 560 520 550 505 580 C 495 610 560 640 620 680 C 700 720 820 700 830 620 C 838 560 760 550 720 600 C 670 670 690 820 800 860 C 880 885 940 820 980 760 C 1010 710 1020 640 1030 620 C 1040 680 1030 780 1020 860 C 1060 680 1140 600 1200 620 C 1250 640 1240 760 1250 840 C 1260 880 1320 900 1400 850 C 1500 790 1600 760 1700 780'

const STE =
  'M -125 600 C 75 600 275 560 345 500 C 415 440 475 220 575 160 C 675 100 735 160 695 200 C 655 240 515 240 525 320 C 535 400 715 400 695 480 C 675 560 515 560 495 500 C 475 440 575 450 655 470 C 735 490 835 470 865 380 C 895 290 905 200 885 180 C 865 160 850 220 850 300 C 850 380 855 500 915 500 C 975 500 1045 460 1095 420 C 1145 380 1175 320 1135 310 C 1095 300 1045 360 1065 430 C 1085 500 1175 510 1255 480 C 1335 450 1435 420 1675 420'

const STE_BAR = 'M 775 300 C 835 290 925 288 985 298'

function Gradients({ a, b }: { a: string; b: string }) {
  return (
    <defs>
      <linearGradient id={a} x1="0" y1="0" x2="1" y2="0">
        <stop offset="0" stopColor="#EA2915" />
        <stop offset=".35" stopColor="#FD572A" />
        <stop offset=".7" stopColor="#28937A" />
        <stop offset="1" stopColor="#5E9628" />
      </linearGradient>
      <linearGradient id={b} x1="0" y1="0" x2="1" y2="0">
        <stop offset="0" stopColor="#5E9628" />
        <stop offset=".5" stopColor="#C6241C" />
        <stop offset="1" stopColor="#FD572A" />
      </linearGradient>
    </defs>
  )
}

function Bands({ d, a, b }: { d: string; a: string; b: string }) {
  return (
    <>
      <path d={d} className="r1" pathLength={2000} stroke={`url(#${a})`} strokeWidth="56" strokeOpacity=".07" strokeDasharray="520 1480" />
      <path d={d} className="r1" pathLength={2000} stroke={`url(#${a})`} strokeWidth="1.6" strokeOpacity=".32" strokeDasharray="520 1480" />
      <path d={d} className="r2" pathLength={2000} stroke={`url(#${b})`} strokeWidth="26" strokeOpacity=".06" strokeDasharray="360 1640" />
      <path d={d} className="r2" pathLength={2000} stroke={`url(#${b})`} strokeWidth="1.2" strokeOpacity=".26" strokeDasharray="360 1640" />
    </>
  )
}

/** "ven" — lowercase, floating. Behind Certifications and Testimonials. */
export function RibbonVen() {
  return (
    <svg className="sig-ribbon" aria-hidden="true" viewBox="0 0 1600 1400" preserveAspectRatio="xMidYMid slice">
      <Gradients a="venA" b="venB" />
      <g className="float" fill="none" strokeLinecap="round" strokeLinejoin="round">
        <path d={VEN} stroke="url(#venA)" strokeWidth="2" strokeOpacity=".1" />
        <Bands d={VEN} a="venA" b="venB" />
      </g>
    </svg>
  )
}

/** "Ste" — capital S, still, centred, wide. Behind the stats and client logos. */
export function RibbonSte() {
  return (
    <svg className="sig-ribbon" aria-hidden="true" viewBox="0 0 1600 700" preserveAspectRatio="xMidYMid slice">
      <Gradients a="steA" b="steB" />
      <g fill="none" strokeLinecap="round" strokeLinejoin="round">
        <path d={STE} stroke="url(#steA)" strokeWidth="2" strokeOpacity=".1" />
        <path d={STE_BAR} stroke="url(#steA)" strokeWidth="2" strokeOpacity=".1" />
        <Bands d={STE} a="steA" b="steB" />
      </g>
    </svg>
  )
}
