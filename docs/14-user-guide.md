# 14 — User guide (CMS)

The user guide now lives in **[`user_guide/`](user_guide/)**:

- [`user_guide/user-guide.md`](user_guide/user-guide.md) — the source, with screenshots in `user_guide/img/`
- [`user_guide/user-guide.pdf`](user_guide/user-guide.pdf) — the same guide as a PDF to share

Both are regenerated with `make user-guide` whenever the CMS changes, and
`make check` fails while they are out of date (decision D39). Server
operation (break-glass, jobs, backups) stays in `15-admin-guide.md`.
