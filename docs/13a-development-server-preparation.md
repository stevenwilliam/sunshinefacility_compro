# 13a — Development server preparation

## Part A — the server, once

`claudedev` already has these (shared by every project): Go, Node 20, PostgreSQL 18 (native), Docker, nginx, ufw, mailpit (`ruuma-mailpit`, 127.0.0.1:1025/8025). Nothing in Part A was changed for sfg.

## Part B — onboard sfg (what was done on 2026-10-22, repeatable)

```bash
# 1. Database
DBPW=$(openssl rand -hex 24)
sudo -u postgres psql -v ON_ERROR_STOP=1 -c "CREATE ROLE sfg_app LOGIN PASSWORD '$DBPW';" \
  -c "CREATE DATABASE sfg OWNER sfg_app;" -c "CREATE DATABASE sfg_test OWNER sfg_app;"

# 2. Object store
MINIOPW=$(openssl rand -hex 24)
docker run -d --name sfg-minio --restart unless-stopped \
  -p 127.0.0.1:9012:9000 -p 127.0.0.1:9013:9001 \
  -e MINIO_ROOT_USER=sfg -e MINIO_ROOT_PASSWORD=$MINIOPW \
  -v sfg-minio-data:/data minio/minio:latest server /data --console-address :9001

# 3. Configuration (copy /home/dev/projects/sfg/.env.example and fill in)
sudo mkdir -p /etc/sfg
sudo vi /etc/sfg/sfg.env
sudo chown root:dev /etc/sfg/sfg.env && sudo chmod 640 /etc/sfg/sfg.env

# 4. Build, migrate, seed
cd /home/dev/projects/sfg
make web build migrate-up seed-assets seed-dev

# 5. Service and timers
sudo cp /home/dev/projects/sfg/deploy/sfg.service /home/dev/projects/sfg/deploy/sfg-*.service /home/dev/projects/sfg/deploy/sfg-*.timer /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now sfg.service sfg-certificate-reminders.timer sfg-purge-applications.timer

# 6. nginx on :8097
sudo cp /home/dev/projects/sfg/deploy/nginx-sfg.conf /etc/nginx/sites-available/sfg
sudo ln -sf /etc/nginx/sites-available/sfg /etc/nginx/sites-enabled/sfg
sudo nginx -t && sudo systemctl reload nginx

# 7. Firewall — BOTH networks Steven arrives from (99 §9)
sudo ufw allow from 192.168.88.0/24 to any port 8097 proto tcp comment 'sfg dev'
sudo ufw allow from 172.16.0.0/24 to any port 8097 proto tcp comment 'sfg dev (VMware host adapter)'

# 8. Browser tooling for visual checks (optional)
cd /home/dev/projects/sfg/web && npx playwright install chromium && npx playwright install --no-shell chromium
sudo env PATH=$PATH npx playwright install-deps chromium
sudo apt-get install -y xvfb
```

**Verify from another machine** (99 §9 — `curl` on the box does not cross
the firewall): open `http://<server-ip>:8097/` from the laptop. If the tab
spins: `sudo grep -a 'DPT=8097' /var/log/ufw.log`.
