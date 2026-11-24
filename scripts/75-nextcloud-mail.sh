#!/usr/bin/env bash
# Adds a Nextcloud Mail account against Stalwart for every lab user that
# already exists in Nextcloud. Keycloak provisions those users on their first
# sign-in, so run this after someone has signed in; re-running skips the
# accounts that are already there.
set -euo pipefail
cd "$(dirname "$0")/.."

test -f app-passwords.txt || { echo "run scripts/35-app-passwords.sh first"; exit 1; }

OCC() { docker exec -u www-data lab-nextcloud php occ "$@"; }

existing=$(OCC user:list --output=json 2>/dev/null | python3 -c '
import sys, json
print(" ".join(json.load(sys.stdin).keys()))')
echo "    nextcloud users: ${existing:-none}"

while IFS=$'\t' read -r address secret; do
	[ -n "$address" ] || continue
	user="${address%%@*}"
	case " $existing " in
		*" $user "*) ;;
		*) echo "    ${user}: not in nextcloud yet, skipped"; continue ;;
	esac
	if OCC mail:account:export "$user" 2>/dev/null | grep -q "$address"; then
		echo "    ${user}: mail account already configured"
		continue
	fi
	OCC mail:account:create-imap "$user" "$user" "$address" \
		mail.lab.internal 993 ssl "$address" "$secret" \
		mail.lab.internal 587 tls "$address" "$secret" >/dev/null
	echo "    ${user}: mail account created"
done < app-passwords.txt
