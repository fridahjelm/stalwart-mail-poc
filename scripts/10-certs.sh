#!/usr/bin/env bash
# Generates a lab CA and one leaf certificate covering every lab name.
# Self-contained: no mkcert, no Homebrew. Trust it with scripts/00-prereqs.sh.
#
#   (no argument)  issue the leaf, reusing the CA if there is one
#   --force        re-issue the leaf even when certs/lab.pem exists
#   --new-ca       start over with a fresh CA
#
# Reusing the CA matters: it is in the macOS keychain and in the Firefox and
# Thunderbird stores, and replacing it means trusting it again everywhere. A
# new name only needs a new leaf.
set -euo pipefail
cd "$(dirname "$0")/.."

mode="${1:-}"
if [ -f certs/lab.pem ] && [ "$mode" != "--force" ] && [ "$mode" != "--new-ca" ]; then
	echo "certs/lab.pem exists; pass --force to re-issue the leaf, or --new-ca to replace the CA too"
	exit 0
fi

if [ ! -f certs/rootCA.pem ] || [ ! -f certs/rootCA-key.pem ] || [ "$mode" = "--new-ca" ]; then
	echo "==> new lab CA"
	openssl req -x509 -newkey rsa:4096 -sha256 -days 825 -nodes \
		-keyout certs/rootCA-key.pem -out certs/rootCA.pem \
		-subj "/CN=mail-poc lab CA/O=mail-poc" \
		-addext "basicConstraints=critical,CA:TRUE,pathlen:0" \
		-addext "keyUsage=critical,keyCertSign,cRLSign"
else
	echo "==> reusing the existing CA ($(openssl x509 -in certs/rootCA.pem -noout -subject))"
fi

cat > certs/leaf.cnf <<'CNF'
[req]
distinguished_name = dn
prompt = no
[dn]
CN = mail.lab.internal
[ext]
basicConstraints = CA:FALSE
keyUsage = critical,digitalSignature,keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt
[alt]
DNS.1 = lab.internal
DNS.2 = mail.lab.internal
DNS.3 = kc.lab.internal
DNS.4 = webmail.lab.internal
DNS.5 = ns.lab.internal
DNS.6 = sink.lab.internal
DNS.7 = other.lab.internal
DNS.8 = cloud.lab.internal
DNS.9 = office.lab.internal
DNS.10 = localhost
IP.1 = 127.0.0.1
CNF

openssl req -newkey rsa:2048 -nodes -keyout certs/lab-key.pem \
	-out certs/lab.csr -config certs/leaf.cnf
openssl x509 -req -in certs/lab.csr -CA certs/rootCA.pem -CAkey certs/rootCA-key.pem \
	-CAcreateserial -days 825 -sha256 \
	-extfile certs/leaf.cnf -extensions ext -out certs/lab-leaf.pem

cat certs/lab-leaf.pem certs/rootCA.pem > certs/lab.pem
rm -f certs/lab.csr certs/lab-leaf.pem

# Keycloak and Caddy run as non-root and must be able to read the key.
chmod 644 certs/lab-key.pem certs/lab.pem certs/rootCA.pem

openssl x509 -in certs/lab.pem -noout -subject -ext subjectAltName
