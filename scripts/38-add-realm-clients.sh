#!/usr/bin/env bash
# Adds any client present in keycloak/import/lab-realm.json but missing from
# the running realm. The import only runs on an empty database, so clients
# added to that file after the first start need this.
set -euo pipefail
cd "$(dirname "$0")/.."

# shellcheck disable=SC1091
set -a; . ./.env; set +a

KC=https://kc.lab.internal
CURL=(curl -fsS --cacert certs/rootCA.pem --resolve "kc.lab.internal:443:127.0.0.1")

TOKEN=$("${CURL[@]}" -d client_id=admin-cli -d username=admin \
	-d "password=${KC_BOOTSTRAP_ADMIN_PASSWORD}" -d grant_type=password \
	"${KC}/realms/master/protocol/openid-connect/token" \
	| python3 -c 'import sys,json; print(json.load(sys.stdin)["access_token"])')

python3 -c '
import json
for c in json.load(open("keycloak/import/lab-realm.json"))["clients"]:
    print(json.dumps(c))
' | while read -r client; do
	id=$(python3 -c 'import json,sys; print(json.loads(sys.argv[1])["clientId"])' "$client")
	found=$("${CURL[@]}" -H "Authorization: Bearer ${TOKEN}" \
		"${KC}/admin/realms/lab/clients?clientId=${id}" \
		| python3 -c 'import sys,json; print(len(json.load(sys.stdin)))')
	if [ "$found" != "0" ]; then
		echo "    ${id} already present"
		continue
	fi
	"${CURL[@]}" -X POST -H "Authorization: Bearer ${TOKEN}" \
		-H 'Content-Type: application/json' -d "$client" \
		"${KC}/admin/realms/lab/clients" >/dev/null
	echo "    ${id} created"
done
