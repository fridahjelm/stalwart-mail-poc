#!/usr/bin/env bash
# Configures a freshly started Stalwart: logging, domains, TLS, listeners,
# the lab resolver, accounts and the Keycloak OIDC directory.
# Safe to re-run: every step is skipped when the object already exists.
set -euo pipefail
cd "$(dirname "$0")/.."

SW() { ./scripts/sw.sh "$@"; }
# shellcheck disable=SC1091
set -a; . ./.env; set +a

# pick_id <Object> <field> <value> -> id of the first object whose <field>
# equals <value>, or contains it as a key when the field is a set.
pick_id() {
	SW query "$1" --json --fields "id,$2" 2>/dev/null | python3 -c '
import sys, json
field, want = sys.argv[1], sys.argv[2]
for line in sys.stdin:
    line = line.strip()
    if not line.startswith("{"):
        continue
    obj = json.loads(line)
    val = obj.get(field)
    if val == want or (isinstance(val, dict) and want in val):
        print(obj["id"])
        break
' "$2" "$3"
}

echo "==> console logging"
# The stock tracer writes to /var/log/stalwart, which does not exist in the
# image, so every log line is dropped. Replace it with a console tracer, which
# is what `docker logs` reads.
SW query Tracer --json --fields id,path 2>/dev/null | python3 -c '
import sys, json
file_tracers, console = [], False
for line in sys.stdin:
    line = line.strip()
    if not line.startswith("{"):
        continue
    obj = json.loads(line)
    (file_tracers.append(obj["id"]) if "path" in obj else None)
    console = console or "path" not in obj
print(",".join(file_tracers))
print("yes" if console else "no")
' > /tmp/tracers.$$
FILE_TRACERS=$(sed -n 1p /tmp/tracers.$$)
HAS_CONSOLE=$(sed -n 2p /tmp/tracers.$$)
rm -f /tmp/tracers.$$
[ -z "$FILE_TRACERS" ] || SW delete Tracer --ids "$FILE_TRACERS"
[ "$HAS_CONSOLE" = yes ] || SW create Tracer/Stdout --field enable=true --field level=info --field ansi=false

echo "==> domains"
for d in lab.internal other.lab.internal; do
	[ -n "$(pick_id Domain name "$d")" ] || SW create Domain --field name="$d"
done
DOM_PRIMARY=$(pick_id Domain name lab.internal)
DOM_OTHER=$(pick_id Domain name other.lab.internal)

echo "==> TLS certificate"
if [ -z "$(pick_id Certificate subjectAlternativeNames mail.lab.internal)" ]; then
	SW create Certificate \
		--field 'certificate={"@type":"File","filePath":"/certs/lab.pem"}' \
		--field 'privateKey={"@type":"File","filePath":"/certs/lab-key.pem"}'
fi
CERT_ID=$(pick_id Certificate subjectAlternativeNames mail.lab.internal)

echo "==> system settings"
SW update SystemSettings \
	--field defaultHostname=mail.lab.internal \
	--field defaultDomainId="$DOM_PRIMARY" \
	--field defaultCertificateId="$CERT_ID"

echo "==> listeners"
mk_listener() { # <name> <protocol> <port> <useTls> <implicitTls>
	[ -n "$(pick_id NetworkListener name "$1")" ] && return 0
	SW create NetworkListener \
		--field name="$1" --field protocol="$2" \
		--field "bind={\"[::]:$3\":true}" \
		--field useTls="$4" --field implicitTls="$5"
}
mk_listener smtp        smtp        25   true  false
mk_listener submission  smtp        587  true  false
mk_listener submissions smtp        465  true  true
mk_listener imap        imap        143  true  false
mk_listener imaps       imap        993  true  true
mk_listener sieve       manageSieve 4190 true  false
mk_listener https       http        443  true  true

# Stalwart's own CORS switch has no origin list: it is all origins or none.
# The allowlist is in caddy/Caddyfile instead, which is also what the browser
# talks to, so this stays off and the two never emit a conflicting header.
echo "==> CORS stays with Caddy, not Stalwart"
SW update Http --field usePermissiveCors=false

# Lets every signed-in user list all accounts (JMAP Principal/query, WebDAV
# principal search), which is what Bulwark's address lookup uses. Off by
# default; read at startup only.
echo "==> directory queries for address lookup"
SW update Sharing --field allowDirectoryQueries=true

# Serves /metrics/prometheus on the HTTP listeners with no authentication,
# which only suits the lab. Read at startup only.
echo "==> Prometheus metrics endpoint"
SW update Metrics --json '{"prometheus":{"@type":"Enabled","authUsername":null,"authSecret":{"@type":"None"}}}'

echo "==> DNS resolver (CoreDNS on the lab network)"
SW update DnsResolver --json '{"@type":"Custom","servers":{"0":{"address":"172.28.0.5","port":53,"protocol":"udp"},"1":{"address":"172.28.0.5","port":53,"protocol":"tcp"}},"maxConcurrentRequests":1,"tcpOnError":true}'

echo "==> relay non-local mail to mailpit, so the lab never reaches real MX hosts"
if [ -z "$(pick_id MtaRoute name sink)" ]; then
	SW create MtaRoute/Relay --field name=sink --field host=sink.lab.internal \
		--field port=1025 --field protocol=smtp --field implicitTls=false \
		--field allowInvalidCerts=true \
		--field description='Lab sink; replaces MX delivery'
fi

# Without this the outbound strategy still falls back to 'mx', and a message
# to a real domain would be delivered over the internet.
SW update MtaOutboundStrategy --json '{"route":{"match":{"0":{"if":"is_local_domain(rcpt_domain)","then":"'"'"'local'"'"'"}},"else":"'"'"'sink'"'"'"}}'

# A room receives every invitation, update and cancellation for it from the
# whole domain, and its iTIP replies all go back to one domain too. The stock
# limit of 25 an hour per sender domain and recipient queues the rest for
# retry, so answers arrive late or after the meeting was already deleted.
# Read at startup only.
echo "==> no sender-to-recipient throttle for room-* accounts"
THROTTLE_ID=$(pick_id MtaInboundThrottle description 'Sender address to recipient throttle')
[ -z "$THROTTLE_ID" ] || SW update MtaInboundThrottle "$THROTTLE_ID" \
	--json '{"match":{"match":{"0":{"if":"starts_with(rcpt, '"'"'room-'"'"') || starts_with(sender, '"'"'room-'"'"')","then":"false"}},"else":"true"}}'

echo "==> accounts"
# No local password: once the OIDC directory is attached Stalwart refuses to
# set one ("Cannot set credentials for accounts in an external directory").
# Keycloak holds the password; 35-app-passwords.sh covers clients that cannot
# use it.
mk_account() { # <local part> <domain id>
	[ -n "$(pick_id Account name "$1")" ] && return 0
	SW create Account/User --field name="$1" --field domainId="$2"
}
mk_account alice "$DOM_PRIMARY"
mk_account bob   "$DOM_PRIMARY"
mk_account carol "$DOM_OTHER"
# Meeting rooms are resource accounts: Stalwart answers their invitations
# itself. Automatic booking declines an overlap with an accepted booking;
# capacity and location are shown to clients but are not booking rules.
# Bob is a delegate, so he sees both room calendars and can answer by hand.
mk_room() { # <local part> <capacity> <location>
	[ -n "$(pick_id Account name "$1")" ] && return 0
	SW create Account/Resource --field name="$1" --field domainId="$DOM_PRIMARY" \
		--field kind=room --field capacity="$2" --field location="$3" \
		--field "description=${1/room-/Room }" \
		--field "delegateIds={\"$(pick_id Account name bob)\":true}"
}
mk_room room-a 8 'Lab, 1st floor'
mk_room room-b 4 'Lab, 2nd floor'

echo "==> Keycloak OIDC directory"
if [ -z "$(pick_id Directory description 'Keycloak lab realm')" ]; then
	SW create Directory/Oidc \
		--field description='Keycloak lab realm' \
		--field issuerUrl='https://kc.lab.internal/realms/lab' \
		--field claimUsername=email \
		--field claimName=name \
		--field claimGroups=groups \
		--field 'acceptedAudiences={"bulwark":true,"imap-client":true,"stalwart":true,"account":true}' \
		--field unscopedTokens=interoperable
fi
DIR_ID=$(pick_id Directory description 'Keycloak lab realm')

echo "==> authentication directory -> ${DIR_ID}"
SW update Authentication --field directoryId="$DIR_ID"

echo
echo "listeners are read at startup only: docker compose restart stalwart"
