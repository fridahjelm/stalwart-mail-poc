# Local mail lab: Keycloak + Stalwart + JMAP webmail

A self-contained mail stack in Docker, built to answer one question: can
Keycloak be the account store for a mail server that serves JMAP to a webmail
client and IMAP to Thunderbird?

Short answer: yes for JMAP and for any IMAP client that speaks SASL
`OAUTHBEARER`; everything else authenticates with a Stalwart app password,
which Stalwart verifies itself. See [Authentication paths](#authentication-paths).

## What runs

| Container | Image | Address on the lab network | Reachable from macOS |
|---|---|---|---|
| `lab-dns` | `coredns/coredns:1.12.0` | 172.28.0.5 | `127.0.0.1:15353/udp` |
| `lab-postgres` | `postgres:17-alpine` | 172.28.0.10 | no |
| `lab-keycloak` | `quay.io/keycloak/keycloak:26.4` | 172.28.0.20 | via Caddy |
| `lab-stalwart` | `stalwartlabs/stalwart:v1` | 172.28.0.30 | mail ports on 127.0.0.1 |
| `lab-caddy` | `caddy:2-alpine` | 172.28.0.40 | `127.0.0.1:443` |
| `lab-bulwark` | `ghcr.io/bulwarkmail/webmail` | 172.28.0.41 | via Caddy |
| `lab-mailpit` | `axllent/mailpit` | 172.28.0.50 | `127.0.0.1:8025` |
| `lab-booking` | built from `booking/` | 172.28.0.60 | no |

Names: `kc.lab.internal` (Keycloak), `mail.lab.internal` (Stalwart),
`webmail.lab.internal` (Bulwark), `sink.lab.internal` (Mailpit).
Mail domains: `lab.internal` and `other.lab.internal`.
Users: `alice@lab.internal`, `bob@lab.internal`, `carol@other.lab.internal`.
Meeting rooms: `room-a@lab.internal` (8 seats), `room-b@lab.internal` (4 seats).

## DNS and TLS

The lab uses `.internal`, which ICANN reserves for private use. `.local` is
mDNS and macOS would intercept it.

Two consumers need to resolve the lab names, and they need the same answers:

- **Inside the lab network**, CoreDNS serves `lab.internal` from
  `dns/zones/lab.internal.zone`. That file is the static A records plus the
  zone data Stalwart generates for each domain: DKIM public keys, DMARC, SPF,
  MX, the `_imaps`/`_jmap`/`_submissions` SRV records, MTA-STS and TLSA.
  `scripts/40-dns-sync.sh` regenerates it. DKIM signing and DMARC verification
  therefore work against real records rather than against a stub.
- **On macOS**, the same names are in `/etc/hosts`, all pointing at
  `127.0.0.1`, where Caddy terminates TLS for every HTTPS name.

The HTTPS names resolve to Caddy on *both* sides, so a URL is the same string
in the browser and inside a container. That matters for one thing in
particular: Keycloak's issuer. Stalwart compares the configured `issuerUrl`
with the issuer the provider reports, so `https://kc.lab.internal/realms/lab`
has to be reachable under exactly that name from Stalwart as well as from the
browser. `mail.lab.internal` is the exception — it resolves to Stalwart itself,
because Caddy does not speak SMTP or IMAP, and Stalwart serves JMAP on 443
there directly.

TLS is a lab CA created by `scripts/10-certs.sh`, with one leaf certificate
covering every name. Each component that validates another one's certificate
is given the lab CA rather than having verification turned off:

- Stalwart reaches Keycloak with `SSL_CERT_FILE=/certs/ca-bundle.pem`, the
  public roots from its own image plus the lab CA.
- Bulwark reaches Stalwart with `NODE_EXTRA_CA_CERTS=/certs/rootCA.pem`. Node
  ships its own root store, so without it the setup wizard fails to verify the
  JMAP endpoint with `SELF_SIGNED_CERT_IN_CHAIN`.
- Caddy reaches both with `tls_trusted_ca_certs /certs/rootCA.pem`.
- macOS gets the CA in the system keychain, and Firefox and Thunderbird get it
  in their own NSS stores. `scripts/00-prereqs.sh` does all three.

## Running it

```bash
./scripts/10-certs.sh              # lab CA and leaf certificate
./scripts/15-ca-bundle.sh          # public roots + lab CA, for Stalwart
./scripts/00-prereqs.sh            # /etc/hosts and system trust (sudo)
./scripts/20-up.sh                 # build and start everything
./scripts/30-stalwart-bootstrap.sh # domains, listeners, accounts, OIDC directory
docker compose restart stalwart    # listeners, the outbound route, directory queries, metrics and throttles are read at startup only
./scripts/35-app-passwords.sh      # one app password per account
./scripts/37-room-passwords.sh     # app passwords for the booking service's rooms
./scripts/40-dns-sync.sh           # publish Stalwart's DKIM/DMARC/SRV records
./scripts/50-verify.sh             # end-to-end check
python3 -I scripts/60-booking-test.py # end-to-end check of room booking
```

Every script is safe to re-run. `30-` skips objects that already exist; `35-`
and `37-` replace the app passwords they made earlier and leave others alone.
`37-` is separate so that minting room passwords does not replace the ones
configured in people's mail clients.

Firefox and Thunderbird use NSS rather than the macOS keychain, so
`00-prereqs.sh` installs the CA into every profile it finds. That needs
`certutil`, from `brew install nss`; without it the script says so and the
import has to be done by hand under Certificates > Authorities.

Clicking through the certificate warning per host is not a substitute. The
webmail page fetches JMAP from `mail.lab.internal` in the background, and an
exception for `webmail.lab.internal` does not cover it, so the sign-in fails
at the callback with no prompt and nothing in the server logs.

### Bulwark's first run

Bulwark configures itself through a browser wizard, which has no scripted
equivalent. Get the token and open the page:

```bash
docker logs lab-bulwark 2>&1 | grep -A2 'SETUP REQUIRED'
```

Then open `https://webmail.lab.internal/setup`, paste the token, and set:

- **Server**: JMAP endpoint `https://mail.lab.internal`
- **Auth**: single sign-on, issuer `https://kc.lab.internal/realms/lab`,
  client id `bulwark`, client secret `bulwark-lab-secret`
- the admin password the wizard asks for is Bulwark's own, unrelated to
  Keycloak or Stalwart

The token expires after an hour; restarting the container issues a new one.

## Authentication paths

| Client | Mechanism | Credential |
|---|---|---|
| Bulwark webmail (JMAP) | OIDC authorization code against Keycloak | Keycloak password |
| Any IMAP client with `OAUTHBEARER` | SASL `OAUTHBEARER`/`XOAUTH2` | Keycloak access token |
| Thunderbird, Apple Mail, Outlook | `LOGIN`/`PLAIN` over TLS | Stalwart app password |

Stalwart's `Authentication.directoryId` points at the OIDC directory, so the
internal password of an account is no longer accepted — `LOGIN` with it
returns `NO [CONTACTADMIN] Not configured`. App passwords still work, because
Stalwart verifies those itself. This is the configuration this lab
demonstrates; the accounts also carry an internal password only so that the
behaviour is visible.

Thunderbird's OAuth2 provider table is hardcoded, so it cannot use a
self-hosted issuer without the `raa-org/thunderbird-custom-idp` add-on
(Thunderbird 140+). The `imap-client` public client in the realm exists for
that experiment; the app-password path is what the lab is set up for.

Accounts are **not** looked up in Keycloak. Stalwart creates a local account on
first successful sign-in, and rejects mail addressed to it until then, so
`30-stalwart-bootstrap.sh` pre-creates them.

### Token shape

The realm puts the mappers on each client rather than on a shared client
scope: during a realm import the built-in scopes do not exist yet when the
clients are created, so a `defaultClientScopes` list naming them is silently
dropped. Each client therefore carries its own audience mapper (`aud:
stalwart`) and group mapper.

The directory sets `unscopedTokens: interoperable`, because Keycloak does not
issue the interoperable mail scopes. Such a token grants mail, contacts,
calendars and files — not management — which is why `35-app-passwords.sh`
mints app passwords as the administrator instead of as the user.

## Demonstrating it

- **Webmail**: `https://webmail.lab.internal`, sign in as `alice` / the
  password in `.env`. The redirect goes to Keycloak and back.
- **IMAP**: Thunderbird against `mail.lab.internal`, 993 implicit TLS, 587
  STARTTLS for submission, username `alice@lab.internal`, password from
  `app-passwords.txt`.
- **Delivery**: send `alice@lab.internal` -> `bob@lab.internal`; it is DKIM
  signed with the key published in the lab zone. `50-verify.sh` does exactly
  this and checks for the `DKIM-Signature` header.
- **Account management**: disable alice in Keycloak, and her next sign-in
  fails while her app password keeps working — the clearest demonstration of
  where the boundary between the two credential stores actually falls.
- **Outbound**: anything not addressed to a lab domain goes to Mailpit
  (`http://127.0.0.1:8025`) through an `MtaRoute` relay, so the lab never
  contacts a real MX host. The outbound strategy's default route is only
  re-read at startup, so this holds from the `docker compose restart stalwart`
  onwards, not before it. `50-verify.sh` checks it by sending to a
  non-lab domain and looking for the message in Mailpit.
- **Address lookup**: `Sharing.allowDirectoryQueries` is on, so any signed-in
  user can list every account and group through JMAP `Principal/query` (and
  WebDAV principal search). Bulwark uses this to look up addresses. It takes
  effect after the `docker compose restart stalwart`; before it, the query
  fails with "The administrator has disabled directory queries".
- **Metrics**: `https://mail.lab.internal/metrics/prometheus` serves
  Prometheus text format without authentication, after the
  `docker compose restart stalwart`; before it, the path returns 404.

## Meeting rooms (prototype)

Stalwart has only user and group accounts and no booking policy, so a room is
an ordinary account and `booking/` answers its invitations. It is a Go
service using only the standard library; the image build runs its unit tests.
One worker per room polls the room's calendar over CalDAV every 5 seconds and
sets the room's PARTSTAT, which makes Stalwart send the iTIP reply. Rooms and
rules are in `booking/rooms.json`:

- no overlap with the room's accepted bookings
- at most 4 hours, at most 90 days ahead, within 07:00–20:00 Stockholm time
- organizers on `lab.internal` or `other.lab.internal` only
- no more people invited than the room seats (other rooms do not count)
- recurring and all-day bookings are declined

To book, invite the room by address like any attendee; Bulwark's address
lookup finds it. Requests are handled one at a time per room, oldest first,
so two requests for one slot cannot both be accepted.

What the service works around, all observed on Stalwart 1.1:

- **A decline reason never reaches the organizer.** A `COMMENT` on the room's
  copy is not carried in the reply, so the service also mails the reason
  from the room's address over submission.
- **Answers survive changes.** Stalwart keeps the room's PARTSTAT when the
  organizer moves a meeting. The service marks each answer with the version
  it answered (`X-LAB-BOOKING`: sequence, start, end, number of people) and
  judges an event again when that no longer matches.
- **A declined room hears nothing more.** Stalwart sends no update or cancel
  to an attendee who declined. The service deletes its declined copies; to
  try again after a decline, the organizer has to ask the room again (remove
  and re-add it, or reset its status).
- **Cancelled copies stay.** A cancelled meeting stays in the room's
  calendar with `STATUS:CANCELLED`; the service deletes it.
- **The inbound throttle.** Stalwart's stock rule allows 25 messages an hour
  per sender domain and recipient. Over the limit, some messages are queued
  for retry and arrive late, after the meeting was moved or deleted; a room's
  iTIP reply was dropped instead, and the service's decline mail refused
  with `452`. The bootstrap exempts messages to and from `room-*` accounts
  from that rule; the per-IP rule still applies.

Organizers get two messages for a decline (Stalwart's reply and the reason)
and one for an accept. Stalwart sends its iTIP reply even between accounts on
the same server and has no setting to stop it.

## Two things that bite in the browser

- The webmail is served from `webmail.lab.internal` and the browser calls JMAP
  on `mail.lab.internal`, so that endpoint needs CORS headers. Stalwart's
  `Http.usePermissiveCors` allows every origin or none, so it stays off and
  `caddy/Caddyfile` answers the preflight and sets
  `Access-Control-Allow-Origin: https://webmail.lab.internal` for that one
  origin. Requests carrying any other `Origin` get no such header and the
  browser drops the response. With neither in place, sign-in reaches Keycloak,
  returns to Bulwark's callback and fails there with `TypeError: Failed to
  fetch` in the console, while the server logs nothing, because the failing
  request is the browser's.
- Caddy's HTTP/3 advertisement is turned off. Only TCP 443 is published to the
  host, so a browser that acts on `alt-svc: h3=":443"` tries QUIC against a
  UDP port that is not there.

## Limits

- No inbound mail from the internet; port 25 is bound to `127.0.0.1`.
- TLSA records are published but DANE is not exercised; nothing in the lab
  validates them.
- Keycloak's `start --import-realm` uses `IGNORE_EXISTING`. Changing
  `keycloak/import/lab-realm.json` after the first run has no effect until the
  `postgres-data` volume is removed.
- `app-passwords.txt`, `booking-passwords.txt` and the `.env` credentials are
  lab values in plain text.
- The booking service polls instead of using JMAP push, keeps no state
  outside the room calendars, and has no room list or free/busy view in any
  client.
- The Prometheus endpoint has no authentication; anyone who can reach
  `mail.lab.internal` on 443 can read it.

## Licences

The booking service in `booking/` is licensed under the EUPL-1.2 (see
`booking/LICENSE` and `booking/NOTICE`). The rest of this repository has no
licence yet. The licences of the projects the lab runs, and of what goes into
the booking image, are in `LICENSES/`, with an index in
`LICENSES/README.md`.

## Teardown

```bash
docker compose down -v
```

That removes the Stalwart data, the Keycloak database and Bulwark's settings.
`certs/` and the generated zone file stay; delete them by hand if you want a
completely clean start.
