#!/usr/bin/env bash
# Host preparation: name resolution and CA trust. Needs sudo. Run once.
set -euo pipefail
cd "$(dirname "$0")/.."

NAMES=(kc.lab.internal mail.lab.internal webmail.lab.internal ns.lab.internal sink.lab.internal)

test -f certs/rootCA.pem || { echo "run scripts/10-certs.sh first"; exit 1; }

echo "==> /etc/hosts"
for n in "${NAMES[@]}"; do
	if ! grep -qE "^127\.0\.0\.1[[:space:]]+${n}\$" /etc/hosts; then
		echo "127.0.0.1	${n}" | sudo tee -a /etc/hosts >/dev/null
	fi
done
grep lab.internal /etc/hosts

echo "==> trusting the lab CA in the system keychain"
sudo security add-trusted-cert -d -r trustRoot \
	-k /Library/Keychains/System.keychain certs/rootCA.pem

# Firefox and Thunderbird use NSS, not the macOS keychain, so the system trust
# added above does not reach them. Per-host exceptions clicked through in the
# browser are not enough either: the webmail fetches JMAP from a second host in
# the background, and a missing exception there fails with no prompt.
echo "==> Firefox and Thunderbird trust stores"
if ! command -v certutil >/dev/null; then
	echo "    certutil is missing. Install it with: brew install nss"
	echo "    then re-run this script. Without it, import certs/rootCA.pem by hand"
	echo "    under Certificates > Authorities and tick 'identify websites'."
else
	echo "    close Firefox and Thunderbird first if they are running"
	added=0
	for db in "$HOME/Library/Application Support/Firefox/Profiles/"*/ \
		"$HOME/Library/Thunderbird/Profiles/"*/; do
		[ -f "${db}cert9.db" ] || continue
		certutil -A -d "sql:${db}" -t C,, -n "mail-poc lab CA" -i certs/rootCA.pem
		echo "    added to ${db}"
		added=$((added + 1))
	done
	echo "    ${added} profile(s) updated; restart the application to pick it up"
fi

echo "==> done"
