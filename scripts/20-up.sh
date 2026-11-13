#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

test -f certs/lab.pem      || { echo "run scripts/10-certs.sh first"; exit 1; }
test -f certs/ca-bundle.pem || { echo "run scripts/15-ca-bundle.sh first"; exit 1; }

# CoreDNS needs a zone file on the first start; 40-dns-sync.sh fills in the
# records Stalwart generates once it is running.
[ -f dns/zones/lab.internal.zone ] || cp dns/zones/static.zone dns/zones/lab.internal.zone

# Docker would create a directory in place of a missing bind-mounted file.
# The booking service restarts until 37-room-passwords.sh fills this in.
[ -f booking-passwords.txt ] || : > booking-passwords.txt

docker compose up -d --build

# --resolve rather than /etc/hosts, so this works before scripts/00-prereqs.sh.
wait_for() { # wait_for <name> <url>
	printf '==> waiting for %s' "$1"
	for _ in $(seq 1 60); do
		if curl -fsS --cacert certs/rootCA.pem --resolve "${1}:443:127.0.0.1" "$2" >/dev/null 2>&1; then
			echo " ok"; return 0
		fi
		printf '.'; sleep 3
	done
	echo " timed out"; return 1
}
wait_for kc.lab.internal https://kc.lab.internal/realms/lab/.well-known/openid-configuration

# Stalwart's HTTPS listener has no certificate until the bootstrap runs, so
# this checks the plain HTTP listener.
printf '==> waiting for mail.lab.internal'
for _ in $(seq 1 60); do
	if curl -fsS http://127.0.0.1:8080/healthz/live >/dev/null 2>&1; then
		echo " ok"; break
	fi
	printf '.'; sleep 3
done

docker compose ps --format 'table {{.Name}}\t{{.Status}}'
