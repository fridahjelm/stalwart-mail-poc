#!/usr/bin/env bash
# One app password per room-* account, for the booking service. Kept apart
# from 35-app-passwords.sh so that minting these does not replace the app
# passwords people have configured in their mail clients. Re-running replaces
# the ones this script made.
set -euo pipefail
cd "$(dirname "$0")/.."

# shellcheck disable=SC1091
set -a; . ./.env; set +a

API=http://127.0.0.1:8080
LABEL='Booking service'
OUT=booking-passwords.txt

jmap() { # jmap <json body> -> response body
	curl -s -u "admin:${STALWART_ADMIN_PASSWORD}" -H 'Content-Type: application/json' \
		-d "$1" "${API}/jmap/"
}

: > "$OUT"
chmod 600 "$OUT"
./scripts/sw.sh query Account --json --fields id,name,emailAddress,credentials 2>/dev/null \
	| python3 -c '
import sys, json
label = sys.argv[1]
for line in sys.stdin:
    line = line.strip()
    if not line.startswith("{"):
        continue
    o = json.loads(line)
    creds = o.get("credentials")
    if creds is None or not o.get("name", "").startswith("room-"):
        continue
    stale = [c["credentialId"] for c in creds.values()
             if c.get("@type") == "AppPassword" and c.get("description") == label]
    print(o["id"], o["emailAddress"], ",".join(stale))
' "$LABEL" | while read -r id addr stale; do
	if [ -n "${stale:-}" ]; then
		destroy=$(python3 -c 'import json,sys; print(json.dumps(sys.argv[1].split(",")))' "$stale")
		jmap "{\"using\":[\"urn:ietf:params:jmap:core\",\"urn:stalwart:jmap\"],\"methodCalls\":[[\"x:AppPassword/set\",{\"accountId\":\"${id}\",\"destroy\":${destroy}},\"c0\"]]}" >/dev/null
	fi
	secret=$(jmap "{\"using\":[\"urn:ietf:params:jmap:core\",\"urn:stalwart:jmap\"],\"methodCalls\":[[\"x:AppPassword/set\",{\"accountId\":\"${id}\",\"create\":{\"c\":{\"@type\":\"AppPassword\",\"description\":\"${LABEL}\"}}},\"c0\"]]}" \
		| python3 -c 'import sys,json; d=json.load(sys.stdin); print(d["methodResponses"][0][1].get("created",{}).get("c",{}).get("secret",""))')
	if [ -n "$secret" ]; then
		printf '%s\t%s\n' "$addr" "$secret" >> "$OUT"
		echo "$addr"
	else
		echo "failed for ${addr}" >&2
		exit 1
	fi
done

echo
echo "written to $(pwd)/${OUT}; restart the booking service to pick them up"
