#!/usr/bin/env bash
# End-to-end check of every path the lab is meant to demonstrate.
set -euo pipefail
cd "$(dirname "$0")/.."

echo "dns"
for q in "MX lab.internal" "MX other.lab.internal" "A mail.lab.internal" "A kc.lab.internal"; do
	# shellcheck disable=SC2086
	answer=$(dig +short -p 15353 @127.0.0.1 $q)
	printf '  [%s] %s: %s\n' "$([ -n "$answer" ] && echo ok || echo FAIL)" "$q" "${answer:-no answer}"
done

python3 -I scripts/verify.py
