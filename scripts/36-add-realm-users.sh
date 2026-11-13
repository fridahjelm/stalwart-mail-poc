#!/usr/bin/env bash
# Adds any user present in keycloak/import/lab-realm.json but missing from the
# running realm. The import only runs on an empty database, so new users added
# to that file after the first start need this.
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

GROUP_ID=$("${CURL[@]}" -H "Authorization: Bearer ${TOKEN}" "${KC}/admin/realms/lab/groups" \
	| python3 -c 'import sys,json; print(next(g["id"] for g in json.load(sys.stdin) if g["name"]=="mail-users"))')

python3 -c '
import json, sys
realm = json.load(open("keycloak/import/lab-realm.json"))
for u in realm["users"]:
    print(json.dumps(u))
' | while read -r user; do
	name=$(python3 -c 'import json,sys; print(json.loads(sys.argv[1])["username"])' "$user")
	existing=$("${CURL[@]}" -H "Authorization: Bearer ${TOKEN}" \
		"${KC}/admin/realms/lab/users?username=${name}&exact=true" \
		| python3 -c 'import sys,json; print(len(json.load(sys.stdin)))')
	if [ "$existing" != "0" ]; then
		echo "    ${name} already present"
		continue
	fi
	body=$(python3 -c '
import json, sys
u = json.loads(sys.argv[1])
u.pop("groups", None)
print(json.dumps(u))
' "$user")
	"${CURL[@]}" -X POST -H "Authorization: Bearer ${TOKEN}" -H 'Content-Type: application/json' \
		-d "$body" "${KC}/admin/realms/lab/users" >/dev/null
	uid=$("${CURL[@]}" -H "Authorization: Bearer ${TOKEN}" \
		"${KC}/admin/realms/lab/users?username=${name}&exact=true" \
		| python3 -c 'import sys,json; print(json.load(sys.stdin)[0]["id"])')
	"${CURL[@]}" -X PUT -H "Authorization: Bearer ${TOKEN}" \
		"${KC}/admin/realms/lab/users/${uid}/groups/${GROUP_ID}" >/dev/null
	echo "    ${name} created"
done
