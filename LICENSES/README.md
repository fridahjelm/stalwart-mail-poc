# Licences of the components

The lab runs other projects' software. This folder holds each project's own
licence files, fetched from its repository at the version the lab ran on
13 November 2026, so that the terms can be checked without looking them up.

This repository does not redistribute any of these projects. The containers
are pulled from their publishers' registries, and the booking image is built
locally from `booking/Dockerfile`. Most obligations in these licences apply
when the software is distributed or offered to users, which matters for a
production deployment or for publishing an image built here.

## Containers

| Component | Version | Licence (SPDX) | Files here | Fetched from |
|---|---|---|---|---|
| Stalwart mail server | 1.1 | `AGPL-3.0-only OR LicenseRef-SEL` | [stalwart/](stalwart/) | `stalwartlabs/stalwart` @ `e7cde9b` |
| Stalwart CLI (tooling) | 1.0.14 | `AGPL-3.0-only OR LicenseRef-SEL` | [stalwart-cli/](stalwart-cli/) | `stalwartlabs/cli` @ `c108341` |
| Bulwark webmail | 1.21.2 | `AGPL-3.0-only`, with an app-store permission | [bulwark/](bulwark/) | `bulwarkmail/webmail` @ `933bc18` |
| Keycloak | 26.4.7 | `Apache-2.0` | [keycloak/](keycloak/) | `keycloak/keycloak` @ `26.4.7` |
| PostgreSQL | 17.12 | `PostgreSQL` | [postgresql/](postgresql/) | `postgres/postgres` @ `REL_17_12` |
| Caddy | 2.11.9 | `Apache-2.0` | [caddy/](caddy/) | `caddyserver/caddy` @ `v2.11.9` |
| CoreDNS | 1.12.0 | `Apache-2.0` | [coredns/](coredns/) | `coredns/coredns` @ `v1.12.0` |
| Mailpit (lab only) | 1.29.2 | `MIT` | [mailpit/](mailpit/) | `axllent/mailpit` @ `v1.29.2` |

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
- The other images (Keycloak, PostgreSQL, Caddy, Mailpit) contain operating
  system packages of their own, under their publishers' terms; they are not
  listed here.

## The booking service

`booking/` is this repository's own code, licensed under the EUPL-1.2 like
the rest of the repository; see `booking/LICENSE` and `booking/NOTICE`. What
its build puts into the image:

| Part | Version | Licence (SPDX) | File here | Fetched from |
|---|---|---|---|---|
| Go standard library, compiled into the binary | 1.23.12 | `BSD-3-Clause` | [go/LICENSE](go/LICENSE) | `golang/go` @ `go1.23.12` |
| BusyBox (`busybox`, `busybox-binsh`, `ssl_client`) | 1.36.1 | `GPL-2.0-only` | [alpine/busybox-LICENSE](alpine/busybox-LICENSE) | `mirror/busybox` @ `1_36_1` |
| musl (`musl`, `musl-utils`) | 1.2.5 | `MIT`; `musl-utils` adds `BSD-2-Clause` and `GPL-2.0-or-later` parts | [alpine/musl-COPYRIGHT](alpine/musl-COPYRIGHT) | git.musl-libc.org @ `v1.2.5` |
| OpenSSL (`libcrypto3`, `libssl3`) | 3.3.7 | `Apache-2.0` | [alpine/openssl-LICENSE.txt](alpine/openssl-LICENSE.txt) | `openssl/openssl` @ `openssl-3.3.7` |
| zlib | 1.3.2 | `Zlib` | [alpine/zlib-LICENSE](alpine/zlib-LICENSE) | `madler/zlib` @ `v1.3.2` |
| pax-utils (`scanelf`) | 1.3.7 | `GPL-2.0-only` | [alpine/pax-utils-COPYING](alpine/pax-utils-COPYING) | `gentoo/pax-utils` @ `v1.3.7` |
| apk-tools | 2.14.4 | `GPL-2.0-only` | [alpine/apk-tools-LICENSE](alpine/apk-tools-LICENSE) | gitlab.alpinelinux.org `alpine/apk-tools` @ `v2.14.4` |
| ca-certificates-bundle | 20260413 | `MPL-2.0 AND MIT` | [alpine/ca-certificates-MPL-2.0.txt](alpine/ca-certificates-MPL-2.0.txt) | SPDX licence list @ `v3.29.0` |
| alpine-baselayout, alpine-keys | 3.6.5, 2.4 | `GPL-2.0-only`, `MIT` | (no upstream file) | — |

The base image is Alpine Linux 3.20.10; the package licences are the ones
`apk list --installed` reports in the built image. Alpine's ca-certificates
repository has no licence file: its certificate data is Mozilla's
`certdata.txt` under MPL-2.0, whose canonical text is included instead. The
GPL-2.0 packages oblige whoever distributes the image to offer their source,
which Alpine publishes; building locally, as the lab does, triggers nothing.
The binary is static, so a `scratch` or distroless base would remove every
package in this table except the Go standard library.
