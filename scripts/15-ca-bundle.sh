#!/usr/bin/env bash
# Stalwart validates Keycloak's certificate against this bundle: the public
# roots shipped in its image plus the lab CA.
set -euo pipefail
cd "$(dirname "$0")/.."

test -f certs/rootCA.pem || { echo "run scripts/10-certs.sh first"; exit 1; }

docker run --rm --entrypoint cat stalwartlabs/stalwart:v1 \
	/etc/ssl/certs/ca-certificates.crt > certs/ca-bundle.pem
cat certs/rootCA.pem >> certs/ca-bundle.pem
chmod 644 certs/ca-bundle.pem
echo "certs/ca-bundle.pem: $(grep -c 'BEGIN CERTIFICATE' certs/ca-bundle.pem) certificates"
