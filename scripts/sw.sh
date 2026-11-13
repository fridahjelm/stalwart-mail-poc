#!/usr/bin/env bash
# stalwart-cli wrapper: scripts/sw.sh query Domain --fields id,name
exec docker compose -f "$(dirname "$0")/../docker-compose.yml" run --rm -T cli "$@"
