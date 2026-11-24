#!/usr/bin/env bash
# The public roots shipped in an image plus the lab CA. Stalwart validates
# Keycloak's certificate against its bundle; PHP in the Nextcloud image
# validates Stalwart's against the other one, through openssl.cafile in
# nextcloud/php/lab-ca.ini. Nextcloud's HTTP client is a separate matter and
# uses occ security:certificates:import instead.
set -euo pipefail
cd "$(dirname "$0")/.."

test -f certs/rootCA.pem || { echo "run scripts/10-certs.sh first"; exit 1; }

bundle() { # bundle <image> <output>
	docker run --rm --entrypoint cat "$1" /etc/ssl/certs/ca-certificates.crt > "$2"
	cat certs/rootCA.pem >> "$2"
	chmod 644 "$2"
	echo "$2: $(grep -c 'BEGIN CERTIFICATE' "$2") certificates"
}

bundle stalwartlabs/stalwart:v1.2.0 certs/ca-bundle.pem
bundle nextcloud:35.0.3-apache certs/ca-bundle-nextcloud.pem
