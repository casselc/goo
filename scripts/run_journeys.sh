#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
# Only the owned demo generator is paused. No stream entries or volumes are
# reset; stale pending entries left by prior crashes are retained for diagnosis.
traffic_was_running=false
if docker compose ps --status running --services | rg -qx traffic; then
  traffic_was_running=true
  docker compose stop traffic
fi
restore() {
  local result=$?
  if "$traffic_was_running"; then
    if ! docker compose start traffic; then
      echo 'Could not restore the demo traffic generator.' >&2
      result=1
    fi
  fi
  if ! python3 scripts/retain_diagnostics.py --apply; then
    echo 'Diagnostic retention failed; test exit status is preserved. Run npm run diagnostics:plan.' >&2
  fi
  exit "$result"
}
trap restore EXIT
python3 - <<'PY'
import json, subprocess, time
def redis(*args):
    return json.loads(subprocess.check_output(
        ['docker','compose','exec','-T','redis','redis-cli','--json',*args], text=True))
deadline=time.monotonic()+120
while time.monotonic()<deadline:
    groups=redis('XINFO','GROUPS','reservations')
    groups=[dict(zip(g[::2],g[1::2])) if isinstance(g,list) else g for g in groups]
    group=next((g for g in groups if g['name']=='inventory'),None)
    pending=redis('XPENDING','reservations','inventory','-','+','100')
    # Normal demo work takes at most 14 s. Older pending entries are a known
    # unreclaimed-crash limitation, not active work that this test may delete.
    if group and group.get('lag')==0 and all(item[2]>20000 for item in pending):
        print('Demo stream drained; pre-existing stale pending entries:',len(pending))
        break
    time.sleep(1)
else:
    raise SystemExit('Demo stream did not drain; inspect lag/pending work without resetting it.')
PY
if [[ "${1:-}" == "--smoke" ]]; then
  shift
  python3 scripts/smoke.py "$@"
else
  node_modules/.bin/playwright test "$@"
fi
