# 11 — Local development

Everything runs on the shared dev server `claudedev`; there is no laptop setup.

## Ports

| Port | What |
|---|---|
| 8097 | nginx → the site and `/tytb` (LAN + VMware host net) |
| 8098 | `sfg serve`, loopback |
| 9012 / 9013 | MinIO API / console, loopback |
| 1025 / 8025 | mailpit SMTP / UI, loopback (shared) |
| 5173 | `npm run dev` (Vite) if you want hot reload; it proxies `/api` and `/media` to 8098 |

## First time

```bash
cd /home/dev/projects/sfg
make web build migrate-up
make seed-assets          # images into MinIO (≈ 25 s)
make seed-dev             # admin@sfg.local, password from /etc/sfg/sfg.env
sudo systemctl start sfg
```

## Everyday

```bash
make deploy               # rebuild both halves, migrate, restart
make test                 # Go suites (fresh schema per test in sfg_test)
make e2e                  # Playwright against :8097
make check                # the gate before a commit
make icons                # after importing a new icon from '@/lib/phosphor'
make shots                # screenshots vs the approved design (Xvfb)
make job JOB=deliver-mail # run a job now
journalctl -u sfg -f
```

## Notes

- `web/.npmrc` sets `legacy-peer-deps` (npm 10.8/10.9 crashes on an optional peer of the build chain, D31). `npm ci` works with it.
- The Go toolchain is pinned to 1.26.10 in `go.mod`; `go` downloads it on first use.
- Headless Chromium on this VM has no WebGL (VMware GPU). The e2e suite does not need it; the hero ball is checked visually under Xvfb with software GL (`make shots`, or `xvfb-run` + headed Chromium).
- Mail in development lands in mailpit: `ssh -L 8025:127.0.0.1:8025 claudedev` then http://localhost:8025.
