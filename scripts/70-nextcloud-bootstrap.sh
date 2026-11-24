#!/usr/bin/env bash
# Creates Nextcloud's database in the Postgres the lab already runs, then
# configures the parts that are not environment variables: the lab CA, the
# Collabora editor, Keycloak sign-in, the Mail account and the apps that are
# deliberately switched off. Safe to re-run.
set -euo pipefail
cd "$(dirname "$0")/.."

# shellcheck disable=SC1091
set -a; . ./.env; set +a

OCC() { docker exec -u www-data lab-nextcloud php occ "$@"; }

echo "==> database"
if [ "$(docker exec lab-postgres psql -U keycloak -d postgres -tAc \
	"SELECT 1 FROM pg_roles WHERE rolname='nextcloud'")" != "1" ]; then
	docker exec lab-postgres psql -U keycloak -d postgres -c \
		"CREATE ROLE nextcloud LOGIN PASSWORD '${NEXTCLOUD_DB_PASSWORD}'" >/dev/null
	echo "    role created"
fi
if [ "$(docker exec lab-postgres psql -U keycloak -d postgres -tAc \
	"SELECT 1 FROM pg_database WHERE datname='nextcloud'")" != "1" ]; then
	docker exec lab-postgres psql -U keycloak -d postgres -c \
		"CREATE DATABASE nextcloud OWNER nextcloud" >/dev/null
	echo "    database created"
fi

echo "==> waiting for the installer"
for _ in $(seq 1 60); do
	if docker exec -u www-data lab-nextcloud php occ status 2>/dev/null | grep -q 'installed: true'; then
		break
	fi
	sleep 5
done
OCC status | sed 's/^/    /'

# Two things stand between Nextcloud and Keycloak, and both report the same
# "could not reach the provider" in the browser. The CA has to be imported,
# and the SSRF guard has to be told that a private address is acceptable:
# every lab service is on 172.28.0.0/24, which it refuses by default with
# "violates local access rules".
echo "==> lab CA and the SSRF guard"
OCC security:certificates:import /certs/rootCA.pem >/dev/null 2>&1 || true
OCC config:system:set allow_local_remote_servers --value=true --type=boolean
OCC security:certificates | sed 's/^/    /'

echo "==> Collabora"
OCC app:install richdocuments >/dev/null 2>&1 || OCC app:enable richdocuments >/dev/null
OCC config:app:set richdocuments wopi_url --value 'https://office.lab.internal'
OCC config:app:set richdocuments public_wopi_url --value 'https://office.lab.internal'
OCC config:app:set richdocuments disable_certificate_verification --value 'no'
OCC richdocuments:activate-config >/dev/null 2>&1 || true

echo "==> Stalwart stays authoritative for calendars and contacts"
for app in calendar contacts; do
	OCC app:disable "$app" >/dev/null 2>&1 || true
done

echo "==> Keycloak sign-in"
OCC app:install user_oidc >/dev/null 2>&1 || OCC app:enable user_oidc >/dev/null
if ! OCC user_oidc:provider --output=json 2>/dev/null | grep -q '"identifier":"lab"'; then
	OCC user_oidc:provider lab \
		--clientid=nextcloud \
		--clientsecret="${NEXTCLOUD_OIDC_SECRET}" \
		--discoveryuri='https://kc.lab.internal/realms/lab/.well-known/openid-configuration' \
		--mapping-uid=preferred_username \
		--mapping-email=email \
		--mapping-display-name=name \
		--unique-uid=0 >/dev/null
fi
OCC user_oidc:provider --output=json 2>/dev/null | sed 's/^/    /'

echo "==> Mail against Stalwart"
OCC app:install mail >/dev/null 2>&1 || OCC app:enable mail >/dev/null

echo
echo "done. Nextcloud is at https://cloud.lab.internal"
