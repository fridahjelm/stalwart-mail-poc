# Licences of the components

The lab runs other projects' software. This folder holds each project's own
licence files, fetched from its repository at the version the lab runs, so
that the terms can be checked without looking them up. Last checked
24 November 2026.

This repository does not redistribute any of these projects; the containers
are pulled from their publishers' registries. Most obligations in these licences apply
when the software is distributed or offered to users, which matters for a
production deployment or for publishing an image built here.

## Containers

| Component | Version | Licence (SPDX) | Files here | Fetched from |
|---|---|---|---|---|
| Stalwart mail server | 1.2.0 | `AGPL-3.0-only OR LicenseRef-SEL` | [stalwart/](stalwart/) | `stalwartlabs/stalwart` @ `b2d3df9` |
| Stalwart CLI (tooling) | 1.0.14 | `AGPL-3.0-only OR LicenseRef-SEL` | [stalwart-cli/](stalwart-cli/) | `stalwartlabs/cli` @ `c108341` |
| Bulwark webmail | 1.21.3 | `AGPL-3.0-only`, with an app-store permission | [bulwark/](bulwark/) | `bulwarkmail/webmail` @ `4ed6c41` |
| Keycloak | 26.4.7 | `Apache-2.0` | [keycloak/](keycloak/) | `keycloak/keycloak` @ `26.4.7` |
| PostgreSQL | 17.12 | `PostgreSQL` | [postgresql/](postgresql/) | `postgres/postgres` @ `REL_17_12` |
| Caddy | 2.11.9 | `Apache-2.0` | [caddy/](caddy/) | `caddyserver/caddy` @ `v2.11.9` |
| CoreDNS | 1.12.0 | `Apache-2.0` | [coredns/](coredns/) | `coredns/coredns` @ `v1.12.0` |
| Mailpit (lab only) | 1.29.2 | `MIT` | [mailpit/](mailpit/) | `axllent/mailpit` @ `v1.29.2` |
| Nextcloud server | 35.0.3 | `AGPL-3.0-only` | [nextcloud/](nextcloud/) | `nextcloud/server` @ `v35.0.3` |
| Collabora Online | 26.04.6.2 | `MPL-2.0` | [collabora/](collabora/) | `CollaboraOnline/online` @ `5f933cc` |
| Valkey | 8.1.12 | `BSD-3-Clause` | [valkey/](valkey/) | `valkey-io/valkey` @ `8.1` |

Notes:

- **Stalwart** and its CLI are dual-licensed: AGPL-3.0, or the proprietary
  Stalwart Enterprise License v2 (SELv2) for the Enterprise features. The lab
  runs without a licence key and uses no Enterprise feature, so the AGPL
  applies. Each source file carries its own notice under the REUSE
  convention; `Apache-2.0.txt`, `MIT.txt` and `Unlicense.txt` cover the files
  that name them.
- **Bulwark** adds a provisional permission for app-store distribution to
  the AGPL (see the top of its `LICENSE`). Its `NOTICE` keeps the MIT notice
  of the project it was forked from and credits bundled third-party material:
  Twemoji artwork (CC-BY-4.0), three fonts (OFL-1.1) and airport time zone
  data (MIT).
- **Collabora Online**'s `COPYING` only points at the MPL-2.0 rather than
  reproducing it, so the canonical text from the SPDX list (v3.29.0) is
  included beside it. The project has no release tag for the version the
  image carries, so the licence was taken from the default branch; the image
  itself is pinned by digest in `docker-compose.yml`.
- **Nextcloud** is AGPL-3.0 for the server. The apps the lab enables
  (`richdocuments`, `user_oidc`, `mail`) are separate AGPL-3.0 projects of
  the same organisation, installed at run time from its app store rather
  than redistributed here.
- The other images (Keycloak, PostgreSQL, Caddy, Mailpit, Valkey) contain
  operating system packages of their own, under their publishers' terms; they
  are not listed here.
