# 13 — Production deployment handbook

Empty Ubuntu 24.04+ server → www.sunshinefacility.com running. Copy-paste,
top to bottom, as a sudo-capable user. Absolute paths throughout; the editor
is `vi`.

> **As deployed (2026-10-22, D37)** on 119.8.189.246, Ubuntu 26.04, a host
> shared with other projects (`/srv/README.md` lists them). Where it
> differs from the steps below:
> - **Users.** Admin `ven` (sudo, SSH key only). Root SSH and password logins
>   are off (`/etc/ssh/sshd_config.d/00-sfg-hardening.conf`); fail2ban
>   guards sshd; unattended-upgrades is on.
> - **Cloudflare** proxies the domain. ufw admits 80/443 from Cloudflare's
>   ranges only. `/etc/nginx/conf.d/cloudflare-realip.conf` restores the
>   visitor IP. `sites-available/00-default` drops unknown hosts. The
>   certificate comes from `certbot certonly --webroot -w /var/www/letsencrypt`;
>   a deploy hook reloads nginx.
> - **MinIO.** Docker Hub no longer serves `minio/minio`; the dev server's
>   image was copied over (`docker save | ssh … docker load`).
> - **Code.** No clone on the server: build here with `make web build`, then
>   copy `bin/sfg` to `/opt/sfg/app/bin/sfg`, the units to
>   `/etc/systemd/system`, and the seed images to `/opt/sfg/app/assets`.
>   Updating means the same copy, `sfg migrate up`, and a restart.
> - **Mail.** `smtp.office365.com:587` as `noreply@sunshinefood.co.id`
>   (parameter `notify.from` set to match, as Office 365 requires).
> - **PostgreSQL** `timezone = 'UTC'`. Nightly dump at 01:15 UTC to
>   `/var/backups/sfg` (restore tested).

> **Temporary domain.** Until www.sunshinefacility.com is bought, the site
> runs at `sunshinefacility.sunshinefood.co.id`: substitute it for both
> `sunshinefacility.com` names below (one `server_name`, one certbot `-d`,
> no apex redirect). The switch later is `SFG_PUBLIC_BASE_URL`, nginx, a
> certificate, and CMS → Site text → `footer.domain`.

**You need before starting:** the server's public IP; DNS for
`sunshinefacility.com` and `www.sunshinefacility.com` pointing at it; an
SMTP relay (host, port, user, password) allowed to send as
`noreply@sunshinefacility.com`, with SPF and DKIM published; this
repository's deploy access.

## 1. Base system

```bash
sudo apt-get update && sudo apt-get -y upgrade
sudo apt-get install -y nginx certbot python3-certbot-nginx postgresql ufw git make curl ca-certificates
sudo timedatectl set-timezone UTC
sudo ufw allow OpenSSH && sudo ufw allow 'Nginx Full' && sudo ufw --force enable
```

Docker (MinIO only):

```bash
curl -fsSL https://get.docker.com | sudo sh
```

## 2. Service user and directories

```bash
sudo useradd --system --create-home --home-dir /opt/sfg --shell /usr/sbin/nologin sfg
sudo mkdir -p /etc/sfg /var/backups/sfg
sudo chown sfg:sfg /var/backups/sfg
```

## 3. PostgreSQL

```bash
DBPW=$(openssl rand -hex 24); echo "DB password: $DBPW"   # keep it for step 5
sudo -u postgres psql -v ON_ERROR_STOP=1 -c "CREATE ROLE sfg_app LOGIN PASSWORD '$DBPW';" -c "CREATE DATABASE sfg OWNER sfg_app;"
```

## 4. MinIO (private bucket)

```bash
MINIOPW=$(openssl rand -hex 24); echo "MinIO password: $MINIOPW"
sudo docker run -d --name sfg-minio --restart unless-stopped \
  -p 127.0.0.1:9012:9000 -p 127.0.0.1:9013:9001 \
  -e MINIO_ROOT_USER=sfg -e MINIO_ROOT_PASSWORD=$MINIOPW \
  -v /srv/sfg-minio:/data minio/minio:latest server /data --console-address :9001
```

## 5. Get the code

```bash
sudo -u sfg git clone git@github.com:stevenwilliam/sunshinefacility_compro.git /opt/sfg/app
```

## 6. Configuration

```bash
sudo cp /opt/sfg/app/.env.example /etc/sfg/sfg.env
sudo vi /etc/sfg/sfg.env
```

Set: `SFG_ENV=production`, `SFG_PUBLIC_BASE_URL=https://www.sunshinefacility.com`,
`SFG_DATABASE_URL` (password from step 3), `SFG_JWT_SECRET=$(openssl rand -hex 32)`,
`SFG_SETUP_TOKEN=$(openssl rand -hex 16)`, the SMTP relay values,
`SFG_S3_SECRET_KEY` (step 4). **Leave `SFG_DEV_ADMIN_*` unset.**

```bash
sudo chown root:sfg /etc/sfg/sfg.env && sudo chmod 640 /etc/sfg/sfg.env
```

## 7. Build

On the server (needs Go 1.26 and Node 20: `sudo apt-get install -y golang nodejs npm`, or build elsewhere and copy `bin/sfg`):

```bash
cd /opt/sfg/app
sudo -u sfg make web build      # produces /opt/sfg/app/bin/sfg with the site embedded
```

## 8. Migrate and seed

```bash
cd /opt/sfg/app
sudo -u sfg sh -c 'set -a; . /etc/sfg/sfg.env; set +a; /opt/sfg/app/bin/sfg migrate up'
sudo -u sfg sh -c 'set -a; . /etc/sfg/sfg.env; set +a; /opt/sfg/app/bin/sfg seed-assets /opt/sfg/app/docs/design/handoff/design/assets'
```

## 9. systemd

```bash
sudo cp /opt/sfg/app/deploy/sfg.service /opt/sfg/app/deploy/sfg-*.service /opt/sfg/app/deploy/sfg-*.timer /etc/systemd/system/
sudo sed -i 's#/home/dev/projects/sfg#/opt/sfg/app#g; s#^User=dev#User=sfg#; s#^Group=dev#Group=sfg#' /etc/systemd/system/sfg*.service
sudo systemctl daemon-reload
sudo systemctl enable --now sfg.service sfg-certificate-reminders.timer sfg-purge-applications.timer
curl -fsS http://127.0.0.1:8098/readyz
```

## 10. nginx + TLS

```bash
sudo vi /etc/nginx/sites-available/sfg
```

```nginx
upstream sfg_app { server 127.0.0.1:8098; keepalive 16; }
limit_req_zone $binary_remote_addr zone=sfg_auth:10m rate=10r/m;
limit_req_zone $binary_remote_addr zone=sfg_forms:10m rate=20r/m;
server {
    listen 80; listen [::]:80;
    server_name sunshinefacility.com www.sunshinefacility.com;
    location / { return 301 https://www.sunshinefacility.com$request_uri; }
}
server {
    listen 443 ssl http2; listen [::]:443 ssl http2;
    server_name www.sunshinefacility.com;
    # certbot fills in ssl_certificate / ssl_certificate_key
    ssl_protocols TLSv1.2 TLSv1.3;
    client_max_body_size 52m;
    location = /api/v1/auth/login { limit_req zone=sfg_auth burst=5 nodelay; proxy_pass http://sfg_app; include /etc/nginx/proxy_params; }
    location ~ ^/api/v1/public/(quotation|application)$ { limit_req zone=sfg_forms burst=10 nodelay; proxy_pass http://sfg_app; include /etc/nginx/proxy_params; }
    location / { proxy_pass http://sfg_app; include /etc/nginx/proxy_params; proxy_http_version 1.1; proxy_set_header Connection ""; }
}
```

```bash
sudo ln -sf /etc/nginx/sites-available/sfg /etc/nginx/sites-enabled/sfg
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t && sudo systemctl reload nginx
sudo certbot --nginx -d www.sunshinefacility.com -d sunshinefacility.com --redirect --agree-tos -m itdept.sfg@gmail.com
```

`/etc/nginx/proxy_params` must pass `X-Forwarded-Proto` (Ubuntu's does), so
the refresh cookie is `Secure`. `SFG_TRUSTED_PROXIES=127.0.0.1`.

## 11. First superadmin

Open `https://www.sunshinefacility.com/admin`. The setup screen asks for the
setup token (`SFG_SETUP_TOKEN` in `/etc/sfg/sfg.env`), name, email and
password. Afterwards blank the token and restart:

```bash
sudo vi /etc/sfg/sfg.env        # SFG_SETUP_TOKEN=
sudo systemctl restart sfg
```

Then CMS → Parameters → Notifications: confirm recipients and sender.

## 12. Verify (from outside the server)

```bash
curl -sI https://www.sunshinefacility.com/ | grep -i 'strict-transport\|content-security'
curl -s https://www.sunshinefacility.com/ | grep -i 'og:\|<title'
curl -s https://www.sunshinefacility.com/robots.txt
```

Send a test quotation; it arrives by mail and in CMS → Quotation requests.

## 13. Backups (nightly)

```bash
sudo vi /etc/cron.d/sfg-backup
```

```
15 1 * * * postgres pg_dump -Fc sfg > /var/backups/sfg/sfg-$(date +\%F).dump && find /var/backups/sfg -name 'sfg-*.dump' -mtime +30 -delete
```

Mirror `/srv/sfg-minio` off the machine with your backup tool. Test a
restore before launch (`pg_restore -c -d sfg <dump>`).

## 14. Updating

```bash
cd /opt/sfg/app && sudo -u sfg git pull && sudo -u sfg make web build
sudo -u sfg sh -c 'set -a; . /etc/sfg/sfg.env; set +a; /opt/sfg/app/bin/sfg migrate up'
sudo systemctl restart sfg && curl -fsS http://127.0.0.1:8098/readyz
```

Rollback = previous commit's binary; migrations are forward-only.
