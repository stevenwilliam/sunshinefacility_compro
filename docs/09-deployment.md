# 09 — Deployment (development server) and operations

The development server `claudedev` runs the full stack. Production is
`13-production-deployment-handbook.md`.

## What runs where

| Piece | How | Where |
|---|---|---|
| sfg | systemd `sfg.service` (user `dev`, hardened) | `/home/dev/projects/sfg/bin/sfg serve` on 127.0.0.1:8098 |
| Jobs | `sfg-certificate-reminders.timer` 07:00, `sfg-purge-applications.timer` 02:30 (Asia/Jakarta) | oneshot services |
| nginx | `/etc/nginx/sites-available/sfg` | :8097, allowlist LAN + VMware host net |
| ufw | `8097/tcp` from 192.168.88.0/24 and 172.16.0.0/24 | |
| PostgreSQL 18 | native | `sfg`, `sfg_test`, role `sfg_app` |
| MinIO | Docker `sfg-minio` (`--restart unless-stopped`, volume `sfg-minio-data`) | 127.0.0.1:9012 (API), 9013 (console) |
| Mail | shared mailpit (Docker `ruuma-mailpit`) | 127.0.0.1:1025, UI 127.0.0.1:8025 |
| Config | `/etc/sfg/sfg.env` root:dev 0640 | |

## Deploy a change

```bash
cd /home/dev/projects/sfg
make deploy        # web build → go build → migrate up → restart → /readyz
```

The service refuses to start with pending migrations, and refuses a
migration file edited after it was applied.

## Backups

```bash
# database
sudo -u postgres pg_dump -Fc sfg > /var/backups/sfg/sfg-$(date +%F).dump
# objects (needs the mc client)
docker run --rm --network host -v /var/backups/sfg/minio:/backup minio/mc:latest \
  sh -c 'mc alias set s http://127.0.0.1:9012 sfg "$SECRET" && mc mirror --overwrite s/sfg /backup/sfg'
```

Restore: `pg_restore -c -d sfg <dump>`; mirror the bucket back. Media URLs
carry content hashes, so a restored bucket and database must be from the
same moment, or images whose rows point at missing objects will 404.

## Rollback

Forward-only migrations: roll back the **binary** (`git checkout <tag>`,
`make web build`, restart) — the schema stays. A migration that must be
undone gets a new forward migration.

## Logs

`journalctl -u sfg -f` (JSON). Mail: CMS → Outgoing mail, or mailpit UI.
