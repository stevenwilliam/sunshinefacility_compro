/* The hero ball scene — a port of makeHero() in design/SFG Home v2.dc.html.
 *
 * Geometry, materials, timings and easings are the design's own numbers,
 * kept as they are so the motion matches the approved prototype. Read the
 * README section "Hero ball" next to this file.
 *
 * Performance rules (README → Hero ball → Performance):
 *   - three.js is imported only when this module is (after first paint);
 *   - the pixel ratio is capped at 1.75;
 *   - the loop stops when the hero is off-screen or the tab is hidden;
 *   - reduced motion: no auto-spin, instant transitions.
 */
import type * as THREEType from 'three'

type T = typeof THREEType

export interface BallItem {
  key: string
  n: string
  base: string
  ink: string
  onInk: string
  photoUrl: string
  markUrl: string
}

export interface BallRefs {
  wrap: HTMLDivElement
  host: HTMLDivElement
  hud: () => HTMLDivElement | null
  tag: () => HTMLElement | null
  label: (k: number) => HTMLElement | null
}

export interface BallApi {
  split: (b: boolean) => void
  focus: (k: number) => void
  destroy: () => void
}

export function createBall(
  THREE: T,
  items: BallItem[],
  refs: BallRefs,
  getSide: () => boolean,
  on: { split: (b: boolean) => void; focus: (k: number) => void },
): BallApi {
  let dead = false, raf = 0, running = false, visible = true
  const mq = (q: string) => !!(window.matchMedia && window.matchMedia(q).matches)
  const reduce = mq('(prefers-reduced-motion: reduce)'), coarse = mq('(pointer: coarse)')
  const { host, wrap } = refs

  const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true, powerPreference: 'high-performance' })
  renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 1.75))
  renderer.outputColorSpace = THREE.SRGBColorSpace
  const dom = renderer.domElement
  dom.style.cssText = 'display:block;width:100%;height:100%;touch-action:pan-y;cursor:grab;outline:none'
  host.appendChild(dom)
  const scene = new THREE.Scene()
  const camera = new THREE.PerspectiveCamera(35, 1, 0.1, 200)
  scene.add(new THREE.HemisphereLight(0xf2f2ec, 0x161813, 1.0))
  const key = new THREE.DirectionalLight(0xffffff, 1.7); key.position.set(6, 8, 10); scene.add(key)
  const rim = new THREE.DirectionalLight(0xfd572a, 1.5); rim.position.set(-8, -2, -8); scene.add(rim)
  const group = new THREE.Group(); scene.add(group)

  const LAT = Math.PI * 0.84, RB = 3, D = 0.22, RING = 5.4, N7 = items.length
  const WP = 2 * Math.PI * RB / N7, HP = 4.8, KL = 2.3, KH = 0.29 * KL, NX = 8, NT = 80
  let seed = 11; const rnd = () => { seed = (seed * 16807) % 2147483647; return seed / 2147483647 }
  const TAB: number[][] = [[0.34, 0], [0.40, 0.02], [0.425, 0.05]]
  for (let a = 235; a >= -55; a -= 14.5) { const r = a * Math.PI / 180; TAB.push([0.5 + 0.12 * Math.cos(r), 0.17 + 0.12 * Math.sin(r)]) }
  TAB.push([0.575, 0.05], [0.60, 0.02], [0.66, 0])
  const resample = (pts: number[][], n: number) => {
    const L = [0]; for (let i = 1; i < pts.length; i++) L.push(L[i - 1] + Math.hypot(pts[i][0] - pts[i - 1][0], pts[i][1] - pts[i - 1][1]))
    const tot = L[L.length - 1], out: number[][] = []; let j = 0
    for (let a = 0; a <= n; a++) { const d = tot * a / n; while (j < pts.length - 2 && L[j + 1] < d) j++; const f = (d - L[j]) / Math.max(1e-6, L[j + 1] - L[j]); out.push([pts[j][0] + (pts[j + 1][0] - pts[j][0]) * f, pts[j][1] + (pts[j + 1][1] - pts[j][1]) * f - HP / 2]) }
    return out
  }
  const seamPts = () => {
    const sg = [rnd() < .5 ? 1 : -1, rnd() < .5 ? 1 : -1], cs = [HP * (.3 + (rnd() - .5) * .06), HP * (.7 + (rnd() - .5) * .06)]
    const pts: number[][] = [[0, 0]]
    for (let q = 0; q < 2; q++) TAB.forEach(p => pts.push([sg[q] * p[1] * KL, cs[q] + (p[0] - .5) * KL]))
    pts.push([0, HP]); return resample(pts, NT)
  }
  const seams: number[][][] = []; for (let k = 0; k < N7; k++) seams.push(seamPts())
  const buildGeo = (Lc: number[][], Rc: number[][]) => {
    const Pa: number[] = [], I: number[] = []
    const idx = (x: number, y: number, z: number) => { Pa.push(x, y, z); return Pa.length / 3 - 1 }
    const z1 = D / 2, z0 = -D / 2
    const grid = (z: number) => { const g: number[][] = []; for (let a = 0; a <= NT; a++) { const row: number[] = []; for (let b = 0; b <= NX; b++) { const f = b / NX; row.push(idx(Lc[a][0] + (Rc[a][0] - Lc[a][0]) * f, Lc[a][1] + (Rc[a][1] - Lc[a][1]) * f, z)) } g.push(row) } return g }
    const F = grid(z1)
    for (let a = 0; a < NT; a++) for (let b = 0; b < NX; b++) I.push(F[a][b], F[a][b + 1], F[a + 1][b + 1], F[a][b], F[a + 1][b + 1], F[a + 1][b])
    const fC = I.length
    const B = grid(z0)
    for (let a = 0; a < NT; a++) for (let b = 0; b < NX; b++) I.push(B[a][b], B[a + 1][b + 1], B[a][b + 1], B[a][b], B[a + 1][b], B[a + 1][b + 1])
    const bC = I.length - fC
    const wall = (pts: number[][]) => { const t = pts.map(p => idx(p[0], p[1], z1)), u = pts.map(p => idx(p[0], p[1], z0)); for (let i = 0; i < pts.length - 1; i++) I.push(t[i], u[i], u[i + 1], t[i], u[i + 1], t[i + 1]) }
    const bot: number[][] = [], top: number[][] = []
    for (let b = 0; b <= NX; b++) { const f = b / NX; bot.push([Lc[0][0] + (Rc[0][0] - Lc[0][0]) * f, Lc[0][1]]); top.push([Rc[NT][0] + (Lc[NT][0] - Rc[NT][0]) * f, Rc[NT][1]]) }
    wall(bot); wall(Rc); wall(top); wall(Lc.slice().reverse())
    const geo = new THREE.BufferGeometry()
    geo.setAttribute('position', new THREE.Float32BufferAttribute(Pa, 3))
    geo.setIndex(I)
    geo.addGroup(0, fC, 0); geo.addGroup(fC, bC, 1); geo.addGroup(fC + bC, I.length - fC - bC, 2)
    geo.computeVertexNormals()
    return geo
  }

  const sideMat = new THREE.MeshStandardMaterial({ color: 0x2a2c26, roughness: .45, metalness: .25, emissive: 0xea2915, emissiveIntensity: .08 })
  const backMat = new THREE.MeshStandardMaterial({ color: 0x2a2c26, roughness: .7, side: THREE.DoubleSide })
  const textures: THREEType.Texture[] = []
  const loadImg = (src: string) => new Promise<HTMLImageElement>((res, rej) => { const im = new Image(); im.onload = () => res(im); im.onerror = rej; im.src = src })
  const ballMats: THREEType.MeshStandardMaterial[] = []
  const photoMats = items.map(it => {
    const m = new THREE.MeshStandardMaterial({ color: 0xffffff, roughness: .6, metalness: .04, side: THREE.DoubleSide })
    const bmM = new THREE.MeshStandardMaterial({ color: 0xffffff, roughness: .5, metalness: .05, side: THREE.DoubleSide })
    ballMats.push(bmM)
    const TH = 1024, TW = Math.round(TH * (WP + 2 * KH) / HP)
    const c = document.createElement('canvas'); c.width = TW; c.height = TH; const g = c.getContext('2d')!
    g.fillStyle = it.base; g.fillRect(0, 0, TW, TH)
    const tex = new THREE.CanvasTexture(c); tex.colorSpace = THREE.SRGBColorSpace; tex.anisotropy = 4; m.map = tex; textures.push(tex)
    const c2 = document.createElement('canvas'); c2.width = TW; c2.height = TH; const g2 = c2.getContext('2d')!
    g2.fillStyle = it.base; g2.fillRect(0, 0, TW, TH)
    const tex2 = new THREE.CanvasTexture(c2); tex2.colorSpace = THREE.SRGBColorSpace; tex2.anisotropy = 4; bmM.map = tex2; textures.push(tex2)
    Promise.all([loadImg(it.photoUrl), loadImg(it.markUrl)]).then(([ph, mk]) => {
      if (dead) return
      // An SVG without intrinsic size reports 0 — treat it as square.
      const mw = mk.naturalWidth || mk.width || 1, mh0 = mk.naturalHeight || mk.height || mw
      const ix = TW * KH / (WP + 2 * KH), iw = TW - 2 * ix
      g.fillStyle = '#FFFFFF'; g.beginPath(); g.moveTo(0, 0); g.lineTo(TW, 0); g.lineTo(TW, 190); g.lineTo(0, 260); g.closePath(); g.fill()
      const ms = 150, mh = ms * mh0 / mw; g.drawImage(mk, ix + 44, 44, ms, mh)
      const PS = iw * .86, ph2 = PS * mh0 / mw
      const o = document.createElement('canvas'); o.width = Math.round(PS); o.height = Math.round(ph2); const og = o.getContext('2d')!
      og.drawImage(mk, 0, 0, o.width, o.height); og.globalCompositeOperation = 'source-in'
      const sc = Math.max(o.width / ph.width, o.height / ph.height); og.drawImage(ph, (o.width - ph.width * sc) / 2, (o.height - ph.height * sc) / 2, ph.width * sc, ph.height * sc)
      g.drawImage(o, ix + (iw - PS) / 2, TH - ph2 - 70)
      tex.needsUpdate = true
      const s2 = Math.max(TW / ph.width, TH / ph.height)
      g2.drawImage(ph, (TW - ph.width * s2) / 2, (TH - ph.height * s2) / 2, ph.width * s2, ph.height * s2)
      g2.globalAlpha = .3; g2.fillStyle = it.base; g2.fillRect(0, 0, TW, TH); g2.globalAlpha = 1
      const gr2 = g2.createLinearGradient(0, TH * .55, 0, TH); gr2.addColorStop(0, 'rgba(0,0,0,0)'); gr2.addColorStop(1, it.ink); g2.fillStyle = gr2; g2.globalAlpha = .7; g2.fillRect(0, TH * .55, TW, TH * .45); g2.globalAlpha = 1
      tex2.needsUpdate = true
      kick()
    }).catch(() => { /* the colour field stays; the hero still works */ })
    return m
  })

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const pieces: any[] = [], sideMats: THREEType.Material[] = []
  const mat4 = new THREE.Matrix4()
  const panelTh = (k: number) => ((k + .5) / N7) * Math.PI * 2
  for (let k = 0; k < N7; k++) {
    const th = panelTh(k)
    const Nk = new THREE.Vector3(Math.sin(th), 0, Math.cos(th)), Rk = new THREE.Vector3(Math.cos(th), 0, -Math.sin(th)), Uk = new THREE.Vector3(0, 1, 0)
    const q0 = new THREE.Quaternion().setFromRotationMatrix(mat4.makeBasis(Rk, Uk, Nk))
    const Lc = seams[(k + N7 - 1) % N7].map(p => [-WP / 2 + p[0], p[1]])
    const Rc = seams[k].map(p => [WP / 2 + p[0], p[1]])
    const geo = buildGeo(Lc, Rc)
    const pos = geo.attributes.position, n = pos.count, uvA = new Float32Array(n * 2), mp = new Float32Array(n * 3)
    for (let v = 0; v < n; v++) {
      const x = pos.getX(v), y = pos.getY(v), z = pos.getZ(v)
      uvA[v * 2] = (x + WP / 2 + KH) / (WP + 2 * KH); uvA[v * 2 + 1] = y / HP + .5
      const dl = (x / WP) * (Math.PI * 2 / N7), lat = (y / HP) * LAT, r = RB + z
      const px = r * Math.cos(lat) * Math.sin(dl), py = r * Math.sin(lat), pz = r * Math.cos(lat) * Math.cos(dl) - RB
      mp[v * 3] = px; mp[v * 3 + 1] = py; mp[v * 3 + 2] = pz
    }
    geo.setAttribute('uv', new THREE.BufferAttribute(uvA, 2))
    const tmp = new THREE.BufferGeometry(); tmp.setAttribute('position', new THREE.BufferAttribute(mp, 3)); tmp.setIndex(geo.index); tmp.computeVertexNormals()
    geo.morphAttributes.position = [tmp.attributes.position as THREEType.BufferAttribute]; geo.morphAttributes.normal = [tmp.attributes.normal as THREEType.BufferAttribute]
    const sm = sideMat.clone(); sm.color.set(items[k].ink); sideMats.push(sm)
    const bk = backMat.clone(); bk.color.set(items[k].ink).multiplyScalar(.5); sideMats.push(bk)
    const mesh = new THREE.Mesh(geo, [ballMats[k], bk, sm])
    mesh.updateMorphTargets(); mesh.morphTargetInfluences![0] = 1
    const pA = Nk.clone().multiplyScalar(RB), pB = Nk.clone().multiplyScalar(RING)
    mesh.userData = { k, pA, nA: Nk, qA: q0, pB, qB: q0, nB: Nk, side: sm, photo: photoMats[k], q: 0, flash: 0, on: false, locked: false, delay: k / Math.max(1, N7 - 1) * .4, ph: th * 2 }
    const ud = mesh.userData; ud.ballM = ballMats[k]; ud.posterM = photoMats[k]
    ud.pX = Nk.clone().multiplyScalar(RING * 1.9).addScaledVector(Uk, (rnd() - .5) * 4).addScaledVector(Rk, (rnd() - .5) * 2.5)
    const tAx = new THREE.Vector3(rnd() - .5, rnd() - .5, rnd() - .5).normalize()
    ud.qT = q0.clone().multiply(new THREE.Quaternion().setFromAxisAngle(tAx, (k % 2 ? 1 : -1) * 2.4))
    mesh.position.copy(pA); mesh.quaternion.copy(q0)
    group.add(mesh); pieces.push(mesh)
  }
  const capA = (Math.PI / 2 - LAT / 2) + 0.03
  const capT = new THREE.Mesh(new THREE.SphereGeometry(RB + .03, 32, 4, 0, Math.PI * 2, 0, capA), sideMat)
  const capB = new THREE.Mesh(new THREE.SphereGeometry(RB + .03, 32, 4, 0, Math.PI * 2, Math.PI - capA, capA), sideMat)
  group.add(capT, capB)
  const core = new THREE.Mesh(new THREE.IcosahedronGeometry(1, 3), new THREE.MeshBasicMaterial({ color: 0xfd572a }))
  const glow = new THREE.Mesh(new THREE.IcosahedronGeometry(1, 2), new THREE.MeshBasicMaterial({ color: 0xea2915, transparent: true, opacity: .2, depthWrite: false }))
  const coreLight = new THREE.PointLight(0xfd572a, 0, 40, 1.4)
  group.add(core, glow, coreLight)
  const rp = new Float32Array(N7 * 6)
  for (let k = 0; k < N7; k++) { const th = panelTh(k); rp.set([0, 0, 0, Math.sin(th) * (RING - .3), 0, Math.cos(th) * (RING - .3)], k * 6) }
  const rayGeo = new THREE.BufferGeometry(); rayGeo.setAttribute('position', new THREE.BufferAttribute(rp, 3))
  const rays = new THREE.LineSegments(rayGeo, new THREE.LineBasicMaterial({ color: 0xea2915, transparent: true, opacity: 0 }))
  group.add(rays)
  const arcMat = new THREE.MeshBasicMaterial({ color: 0xffffff, transparent: true, opacity: 0 })
  const arcMat2 = new THREE.MeshBasicMaterial({ color: 0xea2915, transparent: true, opacity: 0 })
  const torA = new THREE.Mesh(new THREE.TorusGeometry(1.55, .045, 8, 72), arcMat)
  const torB = new THREE.Mesh(new THREE.TorusGeometry(2.05, .02, 6, 96), arcMat2)
  group.add(torA, torB)
  const gc = document.createElement('canvas'); gc.width = gc.height = 128
  const gx = gc.getContext('2d')!, gr = gx.createRadialGradient(64, 64, 0, 64, 64, 64)
  gr.addColorStop(0, 'rgba(255,180,150,1)'); gr.addColorStop(.4, 'rgba(234,41,21,.42)'); gr.addColorStop(1, 'rgba(234,41,21,0)')
  gx.fillStyle = gr; gx.fillRect(0, 0, 128, 128)
  const glowTex = new THREE.CanvasTexture(gc); glowTex.colorSpace = THREE.SRGBColorSpace
  const halo = new THREE.Sprite(new THREE.SpriteMaterial({ map: glowTex, blending: THREE.AdditiveBlending, depthWrite: false, transparent: true, opacity: .4 }))
  scene.add(halo)
  const shock = new THREE.Mesh(new THREE.RingGeometry(.93, 1, 96), new THREE.MeshBasicMaterial({ color: 0xfd572a, transparent: true, opacity: 0, blending: THREE.AdditiveBlending, depthWrite: false, side: THREE.DoubleSide }))
  shock.visible = false; scene.add(shock)

  const st = { s: 0, target: 0, focus: -1, hover: 0, hv: 0, cx: 0, cy: 0, rpx: 0, shockT: 9, cur: '', fa: new Array(N7).fill(0) as number[], yaw: 0.5, pitch: 0.22, tYaw: 0.5, tPitch: 0.22, vYaw: 0 }
  let W = 1, Hh = 1, last = performance.now(), clock = 0
  let drag: { x: number; y: number; m: number } | null = null
  const kick = () => { if (!running && visible && !document.hidden && !dead) { running = true; last = performance.now(); raf = requestAnimationFrame(loop) } }
  const split = (b: boolean) => { if (b && !st.target) st.shockT = -.12; st.target = b ? 1 : 0; if (!b) { st.focus = -1; on.focus(-1) } on.split(!!b); kick() }
  const focus = (k: number) => {
    if (!st.target) { st.target = 1; st.shockT = -.12; on.split(true) }
    st.focus = k; on.focus(k)
    if (k >= 0) { let ty = -panelTh(k); while (ty - st.tYaw > Math.PI) ty -= Math.PI * 2; while (ty - st.tYaw < -Math.PI) ty += Math.PI * 2; st.tYaw = ty; st.tPitch = 0.06; st.vYaw = 0 }
    kick()
  }
  const ray = new THREE.Raycaster(), ndc = new THREE.Vector2()
  const click = (ev: PointerEvent) => {
    const r = dom.getBoundingClientRect()
    if (!st.target) { if (Math.hypot(ev.clientX - r.left - st.cx, ev.clientY - r.top - st.cy) < st.rpx * 1.15) split(true); return }
    ndc.set((ev.clientX - r.left) / r.width * 2 - 1, -((ev.clientY - r.top) / r.height) * 2 + 1)
    ray.setFromCamera(ndc, camera)
    const hit = ray.intersectObjects(pieces, false)[0]
    focus(hit ? hit.object.userData.k : -1)
  }
  const onDown = (ev: PointerEvent) => { drag = { x: ev.clientX, y: ev.clientY, m: 0 }; dom.style.cursor = 'grabbing'; try { dom.setPointerCapture(ev.pointerId) } catch { /* ignore */ } kick() }
  const onMove = (ev: PointerEvent) => {
    if (!drag) { const r = dom.getBoundingClientRect(); st.hover = !st.target && Math.hypot(ev.clientX - r.left - st.cx, ev.clientY - r.top - st.cy) < st.rpx * 1.15 ? 1 : 0; kick(); return }
    const dx = ev.clientX - drag.x, dy = ev.clientY - drag.y; drag.x = ev.clientX; drag.y = ev.clientY; drag.m += Math.abs(dx) + Math.abs(dy)
    st.tYaw += dx * .0065; st.vYaw = dx * .0065
    if (!coarse) st.tPitch = Math.max(-.55, Math.min(.6, st.tPitch + dy * .004))
  }
  const onUp = (ev: PointerEvent) => { if (!drag) return; const m = drag.m; drag = null; dom.style.cursor = 'grab'; if (m < 6) click(ev) }
  const onCancel = () => { drag = null; st.hover = 0; dom.style.cursor = 'grab' }
  dom.addEventListener('pointerdown', onDown); dom.addEventListener('pointermove', onMove); dom.addEventListener('pointerup', onUp); dom.addEventListener('pointercancel', onCancel); dom.addEventListener('pointerleave', onCancel)
  const resize = () => { W = Math.max(1, wrap.clientWidth); Hh = Math.max(1, wrap.clientHeight); renderer.setSize(W, Hh, false); camera.aspect = W / Hh; camera.updateProjectionMatrix(); kick() }
  const ro = new ResizeObserver(resize); ro.observe(wrap)
  const iob = new IntersectionObserver(es => { visible = es[0].isIntersecting; if (visible) kick() }, { threshold: 0.01 }); iob.observe(wrap)
  const onVis = () => { if (!document.hidden) kick() }; document.addEventListener('visibilitychange', onVis)

  const clamp = (v: number) => v < 0 ? 0 : v > 1 ? 1 : v
  const io3 = (p: number) => p < .5 ? 4 * p * p * p : 1 - Math.pow(-2 * p + 2, 3) / 2
  const ax = new THREE.Vector3(), lp = new THREE.Vector3(), cam = new THREE.Vector3(), hp = new THREE.Vector3()
  const outC = (p: number) => 1 - Math.pow(1 - p, 3)
  const outB = (p: number) => { const c1 = .9, c3 = c1 + 1; return 1 + c3 * Math.pow(p - 1, 3) + c1 * Math.pow(p - 1, 2) }
  const PH = HP, DMAX = .4, tf = 2 * Math.tan(17.5 * Math.PI / 180)
  // With reduced motion and nothing changing, the loop renders once and
  // stops instead of spinning at 60 fps for nobody.
  let idleFrames = 0

  const loop = (now: number) => {
    if (dead) return
    if (!visible || document.hidden) { running = false; return }
    const dt = Math.min(.05, (now - last) / 1000); last = now; clock += dt
    const dir = st.target > st.s ? 1 : st.target < st.s ? -1 : 0
    st.s = clamp(st.s + dir * (reduce ? 1 : dt / 2.1))
    const s = st.s, se = io3(s)
    if (!drag) { st.tYaw += st.vYaw; st.vYaw *= .9; if (st.focus < 0 && !reduce) st.tYaw += dt * (s > .5 ? .05 : .16) }
    st.yaw += (st.tYaw - st.yaw) * Math.min(1, dt * 5); st.pitch += (st.tPitch - st.pitch) * Math.min(1, dt * 5)
    group.rotation.set(st.pitch, st.yaw, 0)
    for (let k = 0; k < N7; k++) st.fa[k] += ((st.focus === k ? 1 : 0) - st.fa[k]) * Math.min(1, dt * 4)
    const asp = W / Hh, sd = getSide(), narrow = !sd && asp < .9
    st.hv += (st.hover - st.hv) * Math.min(1, dt * 8)
    const dBall = (2 * RB / (sd ? .56 : .6)) / (tf * Math.min(1, asp))
    const ringW = narrow ? (WP + 2 * KH) * 1.25 : 2 * RING + WP + 2 * KH
    const fr = sd ? .56 : .92
    const dRing = narrow ? ringW / .92 / (tf * asp) + RING : Math.max(ringW / fr / (tf * asp), (PH + 1.4) / .7 / tf) + RING * .55
    camera.position.set(0, .3 * se, dBall + (dRing - dBall) * se); camera.lookAt(0, 0, 0)
    const cxF = sd ? .72 - .01 * se : .5, cyF = sd ? .5 : .46
    camera.setViewOffset(W, Hh, -(cxF - .5) * W, (.5 - cyF) * Hh, W, Hh)
    let fmax = 0; for (let k = 0; k < N7; k++) fmax = Math.max(fmax, st.fa[k])
    const sideGlow = .12 + .2 * (.5 + .5 * Math.sin(clock * 2.4)) + .35 * st.hv
    for (let n = 0; n < pieces.length; n++) {
      const m = pieces[n], u = m.userData
      const q = clamp(s * (1 + DMAX) - u.delay); u.q = q
      if (q > 0 && !u.on) { u.on = true; u.flash = 1 } else if (q <= 0) u.on = false
      if (q >= 1 && !u.locked) { u.locked = true; u.flash = 1.5 } else if (q < 1) u.locked = false
      u.flash *= Math.exp(-dt * 4.5)
      if (q <= 0) {
        const w = reduce ? 0 : .06 * Math.max(0, Math.sin(clock * 1.4 + u.ph)) + .12 * st.hv
        m.position.copy(u.pA).addScaledVector(u.nA, w); m.quaternion.copy(u.qA); m.morphTargetInfluences[0] = 1
      } else if (q >= 1) {
        m.position.copy(u.pB).addScaledVector(u.nB, RING * .55 * io3(st.fa[u.k])); m.quaternion.copy(u.qB); m.morphTargetInfluences[0] = 0
      } else {
        const pU = clamp(q / .12), pO = clamp((q - .12) / .38), pI = clamp((q - .5) / .5)
        if (pI <= 0) {
          const eo = outC(pO)
          hp.copy(u.pA).addScaledVector(u.nA, .5 * io3(pU))
          m.position.lerpVectors(hp, u.pX, eo)
          m.quaternion.slerpQuaternions(u.qA, u.qT, eo)
        } else {
          m.position.lerpVectors(u.pX, u.pB, outB(pI))
          m.quaternion.slerpQuaternions(u.qT, u.qB, io3(Math.min(1, pI * 1.15)))
        }
        m.morphTargetInfluences[0] = 1 - io3(clamp((q - .12) / .45))
      }
      const f = u.flash, pb = io3(q)
      u.side.emissiveIntensity = (1 - se) * sideGlow + se * .06 + 1.8 * f
      m.scale.setScalar(1 + .3 * io3(st.fa[u.k]))
      const L = (.86 + .14 * pb + .4 * f + .1 * st.hv * (1 - s)) * (1 - .6 * fmax * (1 - st.fa[u.k]))
      u.photo.color.setRGB(L, L, L); u.ballM.color.setRGB(L, L, L)
      const wm = q >= .985 ? u.posterM : u.ballM; if (m.material[0] !== wm) m.material[0] = wm
    }
    const cs = 1 - clamp(s * 5)
    capT.visible = capB.visible = cs > .01; capT.scale.setScalar(Math.max(.01, cs)); capB.scale.setScalar(Math.max(.01, cs))
    sideMat.emissiveIntensity = (1 - se) * sideGlow
    const cr = .35 + .55 * se + .03 * Math.sin(clock * 3)
    core.scale.setScalar(cr); glow.scale.setScalar(cr * (1.55 + .15 * Math.sin(clock * 2.2)))
    coreLight.intensity = 50 * se
    arcMat.opacity = .95 * se; arcMat2.opacity = .7 * se; torA.visible = torB.visible = se > .02
    torA.scale.setScalar(cr); torB.scale.setScalar(cr)
    torA.rotation.set(clock * 1.3, clock * .4, 0); torB.rotation.set(1.1, -clock * .9, clock * .3)
    ;(rays.material as THREEType.LineBasicMaterial).opacity = .45 * clamp((s - .6) / .4)
    const hb = RB * 2.9 * (1 + .04 * Math.sin(clock * 2.4)) * (1 + .06 * st.hv)
    halo.scale.setScalar(hb + (cr * 6.5 - hb) * se)
    halo.material.opacity = (1 - se) * (.3 + .1 * Math.sin(clock * 2.4) + .3 * st.hv) + se * .6
    if (st.shockT < 1.2) {
      st.shockT += dt
      const p = clamp(st.shockT / 1.0)
      shock.visible = st.shockT > 0 && p < 1
      shock.quaternion.copy(camera.quaternion)
      shock.scale.setScalar(RB * (1.02 + 3.6 * (1 - Math.pow(1 - p, 3))))
      ;(shock.material as THREEType.MeshBasicMaterial).opacity = .85 * Math.pow(1 - p, 1.6)
    } else shock.visible = false
    group.updateMatrixWorld()
    const dist = camera.position.length()
    lp.set(0, 0, 0).project(camera)
    st.cx = (lp.x * .5 + .5) * W; st.cy = (-lp.y * .5 + .5) * Hh; st.rpx = RB * Hh / (dist * tf)
    const rH = (RB * 1.3 + (cr * 2.6 - RB * 1.3) * se) * Hh / (dist * tf)
    const hud = refs.hud()
    if (hud) { hud.style.transform = 'translate(' + (st.cx - 100) + 'px,' + (st.cy - 100) + 'px) scale(' + (rH / 100) + ')'; hud.style.opacity = String((.75 + .25 * st.hv) * (1 - se)) }
    const tag = refs.tag()
    if (tag) { const tv = 1 - clamp(s * 4); tag.style.transform = 'translate(' + st.cx + 'px,' + (st.cy + st.rpx * 1.26 + 18) + 'px) translateX(-50%)'; tag.style.opacity = String(tv); tag.style.pointerEvents = tv > .5 ? 'auto' : 'none' }
    const want = drag ? 'grabbing' : (st.hover ? 'pointer' : 'grab')
    if (want !== st.cur) { dom.style.cursor = want; st.cur = want }
    for (let k = 0; k < N7; k++) {
      const el = refs.label(k); if (!el) continue
      const lo = clamp((pieces[k].userData.q - .9) / .1) * clamp((s - .5) / .2)
      if (lo <= 0) { el.style.opacity = '0'; continue }
      const th = panelTh(k), rr = RING + RING * .55 * io3(st.fa[k])
      lp.set(Math.sin(th) * rr, PH / 2 + .5, Math.cos(th) * rr).applyMatrix4(group.matrixWorld)
      ax.set(Math.sin(th), 0, Math.cos(th)).applyQuaternion(group.quaternion)
      cam.copy(camera.position).sub(lp).normalize()
      const facing = ax.dot(cam)
      lp.project(camera)
      el.style.transform = 'translate(' + ((lp.x * .5 + .5) * W) + 'px,' + ((-lp.y * .5 + .5) * Hh) + 'px) translate(-50%,-100%)'
      el.style.opacity = facing > .2 ? String(lo * (1 - fmax)) : '0'
    }
    renderer.render(scene, camera)
    if (reduce && !drag && st.s === st.target && Math.abs(st.tYaw - st.yaw) < 1e-3 && Math.abs(st.vYaw) < 1e-4 && st.shockT >= 1.2) {
      if (++idleFrames > 30) { running = false; idleFrames = 0; return }
    } else idleFrames = 0
    raf = requestAnimationFrame(loop)
  }
  resize()

  return {
    split, focus,
    destroy: () => {
      dead = true; cancelAnimationFrame(raf)
      ro.disconnect(); iob.disconnect(); document.removeEventListener('visibilitychange', onVis)
      dom.removeEventListener('pointerdown', onDown); dom.removeEventListener('pointermove', onMove); dom.removeEventListener('pointerup', onUp); dom.removeEventListener('pointercancel', onCancel); dom.removeEventListener('pointerleave', onCancel)
      pieces.forEach(p => p.geometry.dispose()); capT.geometry.dispose(); capB.geometry.dispose(); core.geometry.dispose(); glow.geometry.dispose(); rayGeo.dispose()
      photoMats.forEach(m => m.dispose()); ballMats.forEach(m => m.dispose()); textures.forEach(t => t.dispose()); sideMat.dispose(); backMat.dispose(); sideMats.forEach(m => m.dispose())
      torA.geometry.dispose(); torB.geometry.dispose(); arcMat.dispose(); arcMat2.dispose(); glowTex.dispose(); halo.material.dispose(); shock.geometry.dispose(); (shock.material as THREEType.Material).dispose()
      renderer.dispose(); dom.remove()
    },
  }
}
