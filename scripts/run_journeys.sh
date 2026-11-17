#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
usage() {
  cat <<'HELP'
Usage:
  npm run test:journeys -- journeys/NAME.spec.js [Playwright options]
  npm run test:journeys:reservation -- [specs / Playwright options]
  npm run test:smoke

The default runner only invokes Playwright. Each generic journey owns its runtime
and source prerequisites; provide the spec and any required GOO_GENERIC_BINARY.
--reservation explicitly opts into reservation traffic pause/drain/restore and
post-run diagnostic retention. With no specs that mode runs the full lab suite.
HELP
}
case "${1:-}" in
  --help|-h) usage; exit 0 ;;
  --reservation) shift; exec bash scripts/run_reservation_journeys.sh "$@" ;;
  --smoke) echo 'Smoke uses the reservation lab: run npm run test:smoke.' >&2; exit 2 ;;
  '') usage >&2; exit 2 ;;
esac
# No Docker, Redis, Python, lab traffic, or implicit artifact pruning here.
exec node_modules/.bin/playwright test "$@"
