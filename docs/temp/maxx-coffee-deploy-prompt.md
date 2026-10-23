Deploy this project (www.maxx-coffee.com) to my production server, and leave a handbook so the next update is one copy-paste.

## The server: shared, so be careful
- Host `119.8.189.246`, Ubuntu 26.04, admin user `ven` (sudo, SSH key only; root and password SSH logins are off; fail2ban is on).
- **It is SHARED with other live projects.** Sunshine Facilities Group (`sfg`) runs there and others may too. Read `/srv/README.md` first, it lists them, and add this project to it when you are done.
- Before choosing anything, check what is already taken and pick values that do not clash:
  `sudo ss -ltnp`, `ls /etc/nginx/sites-enabled /etc/systemd/system/*.service`, `sudo docker ps -a`, `sudo -u postgres psql -lqt`, `sudo -u postgres psql -c '\du'`, `ls /opt /srv /etc | head -100`.
- Everything you create gets the project's own prefix (for example `maxx`): system user, `/opt/maxx`, `/etc/maxx/maxx.env`, database and role, systemd unit, nginx site, Docker container, backup folder, and a free loopback port.
- **Never** edit, restart, reload-with-a-broken-config or remove anything that belongs to another project. Always run `sudo nginx -t` before `systemctl reload nginx`. Do not change ufw, sshd, Cloudflare realip, `00-default` or certbot hooks; they already work for every site. Use them as they are.
- Do not upgrade shared system packages (Go, Node, PostgreSQL, Docker) without asking me. If you need a newer toolchain, pin it per project or build on this machine.

## Infrastructure map: what is already deployed (surveyed 2026-10-23)
Server `prod-web-sfg`, 119.8.189.246. Ubuntu 26.04.1 LTS, 3.4 GiB RAM (about 2.6 GiB available) + 2 GiB swap, 79 GB disk (10 % used).
Installed: nginx 1.28.3, PostgreSQL 18.6 (native), Docker 29.1.3, certbot. **No Go, no Node.js on the server** (on purpose: artifacts are built elsewhere and copied over).

**Convention for every project** (from `/srv/README.md`, so follow it exactly):

| Item | Pattern | sfg (live) | maxx (yours) |
|---|---|---|---|
| System user / group | `<p>`, nologin shell | `sfg` | `maxx` |
| App directory | `/opt/<p>/app` (binary in `bin/`, assets beside it) | `/opt/sfg/app/bin/sfg` | `/opt/maxx/app/...` |
| Config | `/etc/<p>/<p>.env`, `root:<p>` `0640` | `/etc/sfg/sfg.env` | `/etc/maxx/maxx.env` |
| systemd unit | `/etc/systemd/system/<p>.service` (+ `<p>-*.timer` for jobs); `User=<p>`, `WorkingDirectory=/opt/<p>/app`, `EnvironmentFile=/etc/<p>/<p>.env` | `sfg.service`, `sfg-certificate-reminders.timer`, `sfg-purge-applications.timer` | `maxx.service` |
| App port | loopback only; **next free above 8100** | `127.0.0.1:8098` | `127.0.0.1:8101` (check it is free first) |
| Database | own role + own DB on the shared PostgreSQL `127.0.0.1:5432` | DB `sfg`, role `sfg_app` | DB `maxx`, role `maxx_app` |
| Object storage (if needed) | own Docker container, loopback ports, data in `/srv/<p>-minio` | `sfg-minio` on 9012 (console 9013) | e.g. `maxx-minio`, 9022/9023 |
| nginx site | one file `/etc/nginx/sites-available/<p>`, symlinked into `sites-enabled` | `sfg` | `maxx` |
| nginx names | upstream `<p>_app`; `limit_req_zone` names `<p>_*` (global namespace!) | `sfg_app`, `sfg_auth`, `sfg_forms` | `maxx_app`, `maxx_*` |
| TLS cert | `certbot certonly --webroot -w /var/www/letsencrypt`; deploy hook `/etc/letsencrypt/renewal-hooks/deploy/reload-nginx` already reloads nginx | `/etc/letsencrypt/live/sunshinefacility.sunshinefood.co.id/` | `/etc/letsencrypt/live/www.maxx-coffee.com/` |
| Backups | `/etc/cron.d/<p>-backup`, `pg_dump -Fc` to `/var/backups/<p>`, 30 days | 01:15 UTC | pick another minute, e.g. 01:45 UTC |
| Registry | add one row to `/srv/README.md` | listed | add it |

**Shared and not yours. Use as-is, do not edit:**
- `/etc/nginx/sites-enabled/00-default`: catch-all. Port 80 serves `/.well-known/acme-challenge/` from `/var/www/letsencrypt` and returns 444 for anything else; port 443 uses `ssl_reject_handshake`. Unknown hosts are dropped.
- `/etc/nginx/conf.d/cloudflare-realip.conf`: restores the visitor IP from `CF-Connecting-IP` for every site.
- ufw (only Cloudflare reaches 80/443), sshd hardening, fail2ban, unattended-upgrades, and the certbot deploy hook.
- Listening now: 22, 80, 443 (public); 127.0.0.1: 5432, 8098, 9012, 9013, plus a few high system ports. Run `ss -ltn` again yourself before you choose.

**The sfg nginx file is the template.** Copy its shape:
```nginx
# maxx — www.maxx-coffee.com → 127.0.0.1:8101
server {
    listen 80; listen [::]:80;
    server_name maxx-coffee.com www.maxx-coffee.com;
    location ^~ /.well-known/acme-challenge/ { root /var/www/letsencrypt; }
    location / { return 301 https://www.maxx-coffee.com$request_uri; }
}
upstream maxx_app { server 127.0.0.1:8101; keepalive 16; }
# limit_req_zone ... zone=maxx_<name>:10m ...   (only maxx_* names)
server {
    listen 443 ssl; listen [::]:443 ssl; http2 on;
    server_name www.maxx-coffee.com maxx-coffee.com;   # apex → www redirect inside, or a separate block
    ssl_certificate     /etc/letsencrypt/live/www.maxx-coffee.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/www.maxx-coffee.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    # do NOT declare ssl_session_cache shared:SSL again with a different size; sfg already defines it
    add_header X-Content-Type-Options nosniff always;
    location / { proxy_pass http://maxx_app; include /etc/nginx/proxy_params;
                 proxy_http_version 1.1; proxy_set_header Connection ""; }
}
```
The certificate does not exist yet, so the first time, enable only the port-80 block (or rely on `00-default` for the ACME challenge), run certbot, and then add the 443 block.

## Adapt this project BEFORE deploying
Do not bend the server to fit the project. Change the project to fit the map above, and commit that first:
1. **One `deploy/` folder in the repo** holding `maxx.service` (with the exact paths, user and EnvironmentFile above, `MemoryMax=`, and hardening like sfg's), the nginx file, any `maxx-*.timer`/`.service` job units, and the backup cron line. The repo is the source; the server gets copies.
2. **Configuration only from environment variables** with one prefix (`MAXX_*`). `.env.example` lists all of them with placeholders. Nothing is hard-coded: no host, port, domain, path or secret. Bind address `127.0.0.1:8101`, public base URL `https://www.maxx-coffee.com`, trusted proxy `127.0.0.1`.
3. **A self-contained build artifact**, because the server has no Go or Node. A Go binary with the frontend embedded (like sfg), or a static `dist/` that nginx serves straight from `/opt/maxx/app/dist`. If the project needs a Node runtime on the server (SSR / Next.js), **stop and ask me** before installing anything shared. Options: a Node standalone build inside the app dir, or a container.
4. **A health endpoint** (`/readyz` or similar) that checks the database, for the deploy check.
5. **Migrations as a command** (`maxx migrate up`) that runs before restart, forward-only in production. Seeds/dev admins are refused in production.
6. **Real visitor IP**: read it from the proxy header that nginx sets (`X-Forwarded-For`/`X-Real-IP` from `proxy_params`), trusting only `127.0.0.1`. Secure cookies rely on `X-Forwarded-Proto`.
7. **One `make deploy-prod`** (or a script in `deploy/`) that builds here, rsyncs to `/opt/maxx/app`, runs migrations, restarts `maxx` only, and curls the health endpoint. Write a matching handbook in `docs/`. Afterwards the next update is one command.
8. Run the project's full test suite and a local production build before anything goes to the server.

## Do NOT crash any project already deployed (most important rule)
Other live websites run on this server. If one goes down, it is a production outage for a real company. Treat every command as if it could take them down.
1. **Snapshot before you touch anything.** Record and keep the output: every site on the box (`curl -sI` on each `server_name` in `/etc/nginx/sites-enabled`, plus each app's local health URL), `systemctl list-units --type=service --state=running`, `sudo docker ps`, `free -h`, `df -h`. After the deploy, run the same checks and compare. Anything that differs is your problem to fix before you report done.
2. **Shared services: reload, never restart or reconfigure.**
   - nginx: `sudo nginx -t && sudo systemctl reload nginx`, never `restart`. If `nginx -t` fails, undo your file right away.
   - PostgreSQL, Docker, ufw, sshd, fail2ban, Cloudflare realip, certbot hooks: do not restart them, and do not edit `postgresql.conf`, `pg_hba.conf`, `/etc/docker/daemon.json` or the firewall. If you think one needs a change, stop and ask me.
3. **Back up before editing any file outside your own `maxx` paths**: `sudo cp -a FILE FILE.bak-$(date +%F-%H%M)`. Write in `docs/RUN-WHEN-BACK.md` how to undo each change.
4. **No destructive cleanup commands**: no `docker system/image/volume prune`, `docker rm` or `docker stop` on containers that are not yours, `apt autoremove`/`purge`, `DROP` on anything that is not yours, `rm -rf` outside `/opt/maxx` / `/srv/maxx*`, `kill`/`pkill` by name, or changes to other projects' crontabs or timers.
5. **Resources.** Check free RAM, disk and CPU first. **Do not build on the server.** Builds (npm, go, next build) can run out of memory and kill the other apps. Build on this machine and copy the result. Give the systemd unit a `MemoryMax=` so this app cannot starve the others. Check that the new backup cron does not fill the disk.
6. **Ports and names.** Before you bind a port or name anything, prove it is unused (`ss -ltnp`, `ls`). Never reuse another project's port, database, role, container, bucket, upstream name or `limit_req_zone` name. nginx zone names are global, so prefix them with `maxx_`. A duplicate zone name fails `nginx -t` for everyone.
7. **If anything else breaks, stop.** Undo your last change first, confirm the other sites are back, then tell me what happened. Do not keep going on top of a broken server.

## How the existing production is set up (copy this pattern)
- **Cloudflare** proxies the domains. ufw only lets Cloudflare's IP ranges reach 80/443. `/etc/nginx/conf.d/cloudflare-realip.conf` restores the visitor IP. Cloudflare SSL mode is **Full (strict)**, so the origin needs a real certificate:
  `sudo certbot certonly --webroot -w /var/www/letsencrypt -d www.maxx-coffee.com -d maxx-coffee.com`. Serve `/.well-known/acme-challenge/` from that webroot on port 80. A deploy hook already reloads nginx.
- **App behind nginx**: the app listens on `127.0.0.1:<free port>` only. nginx owns 443 for `www.maxx-coffee.com`, with an apex → www 301 redirect, and edge `limit_req` on login and public forms. Proxy-trust settings name `127.0.0.1` only.
- **No git clone on the server.** Build on this machine, then `scp`/`rsync` the artifact (binary or built bundle), the systemd units and any seed assets to `/opt/maxx/app/`. Updating is the same copy, then run migrations, restart, and check health.
- **systemd** service runs as its own system user (nologin shell), with `EnvironmentFile=/etc/maxx/maxx.env` and hardening (`NoNewPrivileges`, `ProtectSystem=strict`, `PrivateTmp`, …). Use timers, not cron, for app jobs.
- **Config file**: `/etc/maxx/maxx.env`, owned `root:maxx`, mode `0640`. Never in the repo, and never printed.
- **PostgreSQL** is native and shared: create your own role and database only. Server `timezone = 'UTC'` is already set; business dates go in Asia/Jakarta inside the app. Add a nightly `pg_dump -Fc` to `/var/backups/maxx` that keeps 30 days, at a different minute from the existing 01:15 sfg dump. **Test one restore** into a scratch database, then drop that database.
- **Docker** is only for satellites (object storage etc.). Docker Hub no longer serves `minio/minio`. If you need MinIO, reuse the image already on the server (`sudo docker images`) under a new container name, a new port and a new data dir. Do not touch `sfg-minio`.
- **Mail** (if the site sends any): Office 365 `smtp.office365.com:587`, STARTTLS, AUTH LOGIN. Office 365 does not offer AUTH PLAIN, so make sure the mailer falls back to LOGIN. The From address must equal the mailbox you log in as. I will give you the mailbox and password when you ask.
- Health check with `curl -fsS http://127.0.0.1:<port>/<health>` on the server. Check the public URL from outside too (headers: HSTS, CSP; `<title>`, `og:` tags, `robots.txt`, `sitemap.xml`).

## Passwords and secrets: this is where it broke before
Special characters in passwords (`$ ! @ # % & ' " \ / : ? = + space`) silently break shell commands, `.env` files, systemd `EnvironmentFile`, and `postgres://user:pass@host/db` URLs. Rules:
1. **Generate every secret you control as hex only**: `openssl rand -hex 24` (DB password, object-store keys, JWT/session secret, setup tokens). Hex cannot break anything. If an existing secret has special characters, replace it with a hex one.
2. **Secrets I give you that you cannot change** (Office 365, Cloudflare, third-party API keys) may contain anything. So:
   - never paste them inline into a command line. Write them into the env file with `sudo vi` or a quoted heredoc (`<<'EOF'`), so `$` and `!` are not expanded;
   - in the env file put such a value in **single quotes** if it has spaces or `#`, and check that the app's env loader *and* systemd both read it back identically;
   - inside a URL (DB URL, SMTP URL, Redis URL) **percent-encode** it (`python3 -c 'import urllib.parse,sys;print(urllib.parse.quote(sys.argv[1],safe=""))'`), or better, use separate host/user/password variables;
   - prove it works by logging in with it (`psql`, an SMTP test send), not by eyeballing it.
3. **Mask everything you print.** Never echo, `cat`, `grep` or `systemctl show` a secret into the transcript. Show values as `abcd…****`. Show env files with `sed -E 's/(PASS|SECRET|KEY|TOKEN)([A-Z_]*)=.*/\1\2=****/'`. Never put secrets in commit messages, docs, logs or `git` history. `.env.example` holds placeholders only.
4. List every secret that now exists (name and location, **not** the value) in the handbook, so I can save them in my password manager. Tell me which ones went through chat so I can rotate them.

## Access
SSH as `ven@119.8.189.246` with your key. There is no SSH password; key login is the only way in.

**sudo password for `ven`.** It is not in this prompt on purpose. Do not ask me to paste it into chat. I will put it in a local file on this machine that never leaves it and is never committed: `~/.config/maxx-deploy/sudo-pass`, mode `0600`, with the password on one line. Use it **only** by piping the file in, so special characters are never parsed by a shell and the password never shows up on screen or in a command line:
`ssh ven@119.8.189.246 "sudo -S -p '' <command>" < ~/.config/maxx-deploy/sudo-pass`
(Each call reads only the first line. For a command that also needs stdin, copy a script over first and run `sudo -S -p '' bash /tmp/script.sh`.) Never `cat`, `echo` or log that file. Check it is listed in no git repo, and tell me to delete it when the deploy is done. If the file is missing, ask me to create it with `! mkdir -p ~/.config/maxx-deploy && read -rs P && printf '%s\n' "$P" > ~/.config/maxx-deploy/sudo-pass && chmod 600 ~/.config/maxx-deploy/sudo-pass; unset P`. `read -s` hides what I type, and the password is in no shell history.

If `ssh -o BatchMode=yes ven@119.8.189.246 true` fails, stop. Print this machine's public key (`cat ~/.ssh/id_ed25519.pub`, or create one with `ssh-keygen -t ed25519`) and tell me to add it to `/home/ven/.ssh/authorized_keys` on the server. Anything that needs an interactive terminal or a password prompt (sudo password, certbot first run, Cloudflare dashboard, DNS) goes in `docs/RUN-WHEN-BACK.md` with the exact command; do not stall on it.

## Before you start, ask me (one question at a time)
Only what you cannot find out yourself: is the DNS for `maxx-coffee.com` / `www` already on Cloudflare and pointing at `119.8.189.246`; the mail sender and password (if any); whether the site has a CMS/admin and who the first admin is. Everything else, decide sensibly and write the decision down.

## Done means
- `https://www.maxx-coffee.com` serves the site through Cloudflare with a valid origin certificate, apex redirects to www, and the admin (if any) is reachable with its first user created.
- The sfg site and every other site on the box still answer exactly as before. Show me the before and after snapshots side by side (status codes, health checks, running services and containers, free RAM and disk).
- A `docs/…production-deployment-handbook.md` in this repo that says what was actually deployed: paths, port, units, user, backup, update and rollback commands, and secret names (no values). `/srv/README.md` on the server lists this project too.
- Committed and pushed with a conventional commit.
- Finish with a table: public URL, server, app port, service name, admin URL, admin user, and where each password is stored (never the password itself).
