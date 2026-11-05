#!/usr/bin/env bash
# Routine update on production: new binary + pending migrations + restart.
# Build on the dev server (`make web build`), copy bin/sfg to ~ven/sfg-release/sfg, then on 119.8.189.246:
#   sudo bash ~/sfg-release/prod-update.sh
# Touches nothing else (no nginx, no env). The previous binary is kept for rollback.
set -euo pipefail

REL=/home/ven/sfg-release
STAMP=$(date +%F-%H%M)
say() { printf '\n== %s\n' "$*"; }

[[ $EUID -eq 0 ]] || { echo "run with sudo"; exit 1; }
[[ -s $REL/sfg ]] || { echo "missing $REL/sfg"; exit 1; }

say "backups (suffix .bak-$STAMP)"
cp -a /opt/sfg/app/bin/sfg "/opt/sfg/app/bin/sfg.bak-$STAMP"
sudo -u postgres pg_dump -Fc sfg > "/var/backups/sfg/pre-update-$STAMP.dump"
ls -la "/var/backups/sfg/pre-update-$STAMP.dump"

say "binary + migrations"
install -o root -g root -m 755 "$REL/sfg" /opt/sfg/app/bin/sfg
sudo -u sfg sh -c 'set -a; . /etc/sfg/sfg.env; set +a; /opt/sfg/app/bin/sfg migrate up'

say "restart sfg"
systemctl restart sfg
for i in $(seq 1 15); do curl -fsS http://127.0.0.1:8098/readyz >/dev/null 2>&1 && break; sleep 1; done
curl -fsS http://127.0.0.1:8098/readyz; echo
curl -sk -o /dev/null -w "%{http_code}  ← https://sunshinefacility.com/ at the origin\n" --resolve sunshinefacility.com:443:127.0.0.1 https://sunshinefacility.com/ || echo "origin check failed"

cat <<EOF

Done. Undo (migrations are forward-only; the dump above restores data if ever needed):
  cp -a /opt/sfg/app/bin/sfg.bak-$STAMP /opt/sfg/app/bin/sfg && systemctl restart sfg
EOF
