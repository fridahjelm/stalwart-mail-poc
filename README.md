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
| `lab-stalwart` | `stalwartlabs/stalwart:v1.2.0` | 172.28.0.30 | mail ports on 127.0.0.1 |
| `lab-caddy` | `caddy:2-alpine` | 172.28.0.40 | `127.0.0.1:443` |
| `lab-bulwark` | `ghcr.io/bulwarkmail/webmail:1.21.3` | 172.28.0.41 | via Caddy |
| `lab-mailpit` | `axllent/mailpit` | 172.28.0.50 | `127.0.0.1:8025` |

Names: `kc.lab.internal` (Keycloak), `mail.lab.internal` (Stalwart),
`webmail.lab.internal` (Bulwark), `sink.lab.internal` (Mailpit).
Mail domains: `lab.internal` and `other.lab.internal`.
Users: `alice@lab.internal`, `bob@lab.internal`, `carol@other.lab.internal`.
Meeting rooms: `room-a@lab.internal` (8 seats, 1st floor) and
`room-b@lab.internal` (4 seats, 2nd floor), both resource accounts.

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
./scripts/40-dns-sync.sh           # publish Stalwart's DKIM/DMARC/SRV records
./scripts/50-verify.sh             # end-to-end check
python3 -I scripts/60-room-booking-test.py # how the rooms answer invitations
```

Every script is safe to re-run. `30-` skips objects that already exist; `35-`
replaces the app passwords it made earlier and leaves others alone.

### Upgrading Stalwart

A server first created by an earlier version does not get the permissions a
new release adds to its default roles, and the web interface is a separate
bundle. Going to 1.2.0 needed both:

```bash
docker compose pull stalwart
# one start with the variable set, then remove it again
STALWART_UPDATE_ROLE_PERMISSIONS=true docker compose up -d stalwart
docker compose up -d --force-recreate stalwart
./scripts/sw.sh create Action/UpdateApps   # web interface to 2.1.1
```

Without the first step, mail rules, junk lists and calendar settings are
missing their permissions; the User, Group and System Administrator roles
should gain `sysMailRule*`, `sysBlockedSender*`, `sysSafeSender*` and
`sysCalendarSettings*`, and the administrator role `sysExchange*` as well. A
lab built from scratch on 1.2.0 has them already.

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

## Meeting rooms

Since Stalwart 1.2.0 a room is a resource account (`Account/Resource`,
`kind: room`) and the server answers its invitations itself. No service of the
lab's own is involved. `scripts/30-stalwart-bootstrap.sh` creates both rooms
with a capacity, a location and Bob as a delegate, which lets him see each room
calendar and answer by hand.

To book, invite the room by address like any attendee; Bulwark's address
lookup finds it. `scripts/60-room-booking-test.py` exercises the behaviour.

The booking policy is the resource's `booking` object, here left at its
`Automatic` defaults:

| Field | Default | Effect |
|---|---|---|
| `allowConflicts` | `false` | an overlap with an accepted booking is declined |
| `allowRecurring` | `true` | recurring requests are answered like any other |
| `allowExternal` | `false` | organizers outside the server are refused |
| `bookingWindow` | unset | no limit on how far ahead a room can be booked |
| `maxDuration` | unset | no limit on how long a booking may be |

Verified in this lab: a free slot is accepted, an overlapping slot is
declined, and the decline carries no reason the organizer can read.

**Capacity is not a booking rule.** It is metadata shown to clients: six
people invited to `room-b`, which seats four, were accepted. Exchange behaves
the same way, so a room that is too small is the organizer's problem in both.
The earlier Go prototype did enforce seats, along with a daily 07:00–20:00
window and a refusal of all-day bookings, and the resource policy has no
equivalent of any of the three. It does have `maxDuration` and
`bookingWindow`, which the prototype lacked, both left unset here.
Per-account working hours arrived in the same release
(`x:CalendarSettings`) and may cover the daily window for a resource as well;
that is untested.

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
- `app-passwords.txt` and the `.env` credentials are lab values in plain text.
- A room's capacity is not a booking rule; see "Meeting rooms".
- The Prometheus endpoint has no authentication; anyone who can reach
  `mail.lab.internal` on 443 can read it.

## Licences

This repository is licensed under the EUPL-1.2 (see `LICENSE` and `NOTICE`).
The exception is `LICENSES/`, which holds the licences of the projects the lab
runs, with an index in `LICENSES/README.md`.

## Teardown

```bash
docker compose down -v
```

That removes the Stalwart data, the Keycloak database and Bulwark's settings.
`certs/` and the generated zone file stay; delete them by hand if you want a
completely clean start.
