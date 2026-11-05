#!/usr/bin/env bash
# One-shot cutover on production (D41): temporary domain → www.sunshinefacility.com.
# Run on 119.8.189.246 as:  sudo bash ~/sfg-release/prod-switch-domain.sh
# Expects in ~ven/sfg-release/: sfg (new binary), nginx-sfg-production.conf.
# Expects the three Cloudflare Origin Certificates in /etc/ssl/sfg/<zone>.pem/.key.
# Every file it replaces is backed up first; nginx is restored if `nginx -t` fails.
set -euo pipefail

REL=/home/ven/sfg-release
STAMP=$(date +%F-%H%M)
ZONES=(sunshinefacility.com sunshinefacility.co.id sunshinefacility.id)
say() { printf '\n== %s\n' "$*"; }

[[ $EUID -eq 0 ]] || { echo "run with sudo"; exit 1; }

say "preflight"
for f in "$REL/sfg" "$REL/nginx-sfg-production.conf"; do [[ -s $f ]] || { echo "missing $f"; exit 1; }; done
for z in "${ZONES[@]}"; do
  for ext in pem key; do [[ -s /etc/ssl/sfg/$z.$ext ]] || { echo "missing /etc/ssl/sfg/$z.$ext (Cloudflare Origin Certificate)"; exit 1; }; done
  # certificate and key must belong together
  [[ $(openssl x509 -noout -pubkey -in /etc/ssl/sfg/$z.pem | sha256sum) == $(openssl pkey -pubout -in /etc/ssl/sfg/$z.key | sha256sum) ]] \
    || { echo "/etc/ssl/sfg/$z.pem and .key do not match"; exit 1; }
  openssl x509 -noout -ext subjectAltName -in /etc/ssl/sfg/$z.pem | grep -q "DNS:\*\.$z" \
    || { echo "/etc/ssl/sfg/$z.pem does not cover *.$z"; exit 1; }
done
chown root:root /etc/ssl/sfg/*; chmod 644 /etc/ssl/sfg/*.pem; chmod 600 /etc/ssl/sfg/*.key
echo ok

say "backups (suffix .bak-$STAMP)"
cp -a /opt/sfg/app/bin/sfg "/opt/sfg/app/bin/sfg.bak-$STAMP"
cp -a /etc/sfg/sfg.env "/etc/sfg/sfg.env.bak-$STAMP"
cp -a /etc/nginx/sites-available/sfg "/etc/nginx/sites-available/sfg.bak-$STAMP"
sudo -u postgres pg_dump -Fc sfg > "/var/backups/sfg/pre-domain-$STAMP.dump"
ls -la "/var/backups/sfg/pre-domain-$STAMP.dump"

say "binary + migrations"
install -o root -g root -m 755 "$REL/sfg" /opt/sfg/app/bin/sfg
sudo -u sfg sh -c 'set -a; . /etc/sfg/sfg.env; set +a; /opt/sfg/app/bin/sfg migrate up'

say "SFG_PUBLIC_BASE_URL"
sed -i 's#^SFG_PUBLIC_BASE_URL=.*#SFG_PUBLIC_BASE_URL=https://www.sunshinefacility.com#' /etc/sfg/sfg.env
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

say "verify at the origin (before DNS)"
check() { curl -sk -o /dev/null -w "%{http_code} %{redirect_url}" --resolve "$1:443:127.0.0.1" "https://$1/careers?x=1"; echo "  ← $1"; }
for h in www.sunshinefacility.com sunshinefacility.com www.sunshinefacility.co.id sunshinefacility.co.id \
         www.sunshinefacility.id sunshinefacility.id sunshinefacility.sunshinefood.co.id; do check "$h"; done
curl -sk --resolve www.sunshinefacility.com:443:127.0.0.1 https://www.sunshinefacility.com/robots.txt | grep Sitemap
for h in $(awk '/server_name/{for(i=2;i<=NF;i++){gsub(";","",$i);print $i}}' /etc/nginx/sites-enabled/maxx | sort -u); do
  curl -sk -o /dev/null -w "%{http_code}  ← $h (other site, must be unchanged)\n" --resolve "$h:443:127.0.0.1" "https://$h/"
done

cat <<EOF

Done. Expect 200 for www.sunshinefacility.com and 301 → https://www.sunshinefacility.com/careers?x=1 for the rest.
Undo:
  cp -a /opt/sfg/app/bin/sfg.bak-$STAMP /opt/sfg/app/bin/sfg
  cp -a /etc/sfg/sfg.env.bak-$STAMP /etc/sfg/sfg.env && systemctl restart sfg
  cp -a /etc/nginx/sites-available/sfg.bak-$STAMP /etc/nginx/sites-available/sfg && nginx -t && systemctl reload nginx
  (migration 0008 only changes the footer text; CMS → Site text → footer.domain if needed)
EOF
