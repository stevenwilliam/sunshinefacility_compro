#!/usr/bin/env bash
# One-shot on production (D43): canonical host www.sunshinefacility.com → sunshinefacility.com.
# Run on 119.8.189.246 as:  sudo bash ~/sfg-release/prod-apex-canonical.sh
# Expects in ~ven/sfg-release/: sfg (new binary, migration 0012), nginx-sfg-production.conf.
# Certificates are unchanged: the sunshinefacility.com Origin Certificate covers the apex and www.
# Every file it replaces is backed up first; nginx is restored if `nginx -t` fails.
set -euo pipefail

REL=/home/ven/sfg-release
STAMP=$(date +%F-%H%M)
say() { printf '\n== %s\n' "$*"; }

[[ $EUID -eq 0 ]] || { echo "run with sudo"; exit 1; }

say "preflight"
for f in "$REL/sfg" "$REL/nginx-sfg-production.conf"; do [[ -s $f ]] || { echo "missing $f"; exit 1; }; done
grep -q 'server_name sunshinefacility.com;' "$REL/nginx-sfg-production.conf" || { echo "$REL/nginx-sfg-production.conf is not the apex version"; exit 1; }
echo ok

say "backups (suffix .bak-$STAMP)"
cp -a /opt/sfg/app/bin/sfg "/opt/sfg/app/bin/sfg.bak-$STAMP"
cp -a /etc/sfg/sfg.env "/etc/sfg/sfg.env.bak-$STAMP"
cp -a /etc/nginx/sites-available/sfg "/etc/nginx/sites-available/sfg.bak-$STAMP"
echo ok

say "binary + migrations"
install -o root -g root -m 755 "$REL/sfg" /opt/sfg/app/bin/sfg
sudo -u sfg sh -c 'set -a; . /etc/sfg/sfg.env; set +a; /opt/sfg/app/bin/sfg migrate up'

say "SFG_PUBLIC_BASE_URL"
sed -i 's#^SFG_PUBLIC_BASE_URL=.*#SFG_PUBLIC_BASE_URL=https://sunshinefacility.com#' /etc/sfg/sfg.env
grep '^SFG_PUBLIC_BASE_URL=' /etc/sfg/sfg.env

say "restart sfg"
systemctl restart sfg
for i in $(seq 1 15); do curl -fsS http://127.0.0.1:8098/readyz >/dev/null 2>&1 && break; sleep 1; done
curl -fsS http://127.0.0.1:8098/readyz; echo

say "nginx"
install -o root -g root -m 644 "$REL/nginx-sfg-production.conf" /etc/nginx/sites-available/sfg
if ! nginx -t; then
  echo "nginx -t failed: restoring the previous file"
  cp -a "/etc/nginx/sites-available/sfg.bak-$STAMP" /etc/nginx/sites-available/sfg
  nginx -t && exit 1
fi
systemctl reload nginx

say "verify at the origin"
sleep 2   # reload is asynchronous: let the new workers take over
check() { curl -sk -o /dev/null -w "%{http_code} %{redirect_url}" --resolve "$1:443:127.0.0.1" "https://$1/careers?x=1" || printf 'curl failed (%s)' "$?"; echo "  ← $1"; }
for h in sunshinefacility.com www.sunshinefacility.com www.sunshinefacility.co.id sunshinefacility.co.id \
         www.sunshinefacility.id sunshinefacility.id sunshinefacility.sunshinefood.co.id; do check "$h"; done
curl -sk --resolve sunshinefacility.com:443:127.0.0.1 https://sunshinefacility.com/robots.txt | grep Sitemap || true
curl -sk -o /dev/null -w "%{http_code}  ← www.maxx-coffee.com (other site, must be 200)\n" --resolve www.maxx-coffee.com:443:127.0.0.1 https://www.maxx-coffee.com/ || echo "curl failed ← www.maxx-coffee.com"

cat <<EOF

Done. Expect 200 for sunshinefacility.com and 301 → https://sunshinefacility.com/careers?x=1 for the rest.
Undo:
  cp -a /opt/sfg/app/bin/sfg.bak-$STAMP /opt/sfg/app/bin/sfg
  cp -a /etc/sfg/sfg.env.bak-$STAMP /etc/sfg/sfg.env && systemctl restart sfg
  cp -a /etc/nginx/sites-available/sfg.bak-$STAMP /etc/nginx/sites-available/sfg && nginx -t && systemctl reload nginx
  (migration 0012 only changes the footer text; CMS → Site text → footer.domain if needed)
EOF
