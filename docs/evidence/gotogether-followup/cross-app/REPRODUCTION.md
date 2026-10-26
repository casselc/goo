# Bounded real cross-app journey

Run from `/workspace/goo` so the existing checkout's module resolves the public
`github.com/casselc/goo/exploration.SavedCapture` type. No production digest helper
is called and no application or module dependencies are changed.

```sh
PATH=/workspace/.toolchains/go/bin:$PATH \
GOMODCACHE=/tmp/goo-module-cache GOCACHE=/workspace/.cache/go-build \
GOTOOLCHAIN=local GOMAXPROCS=2 \
go build -o /tmp/goo-gotogether-followup/cross-app/capture-hash \
  /tmp/goo-gotogether-followup/cross-app/capture-hash.go
node /tmp/goo-gotogether-followup/cross-app/run.js
```

The script uses the installed Playwright package/Chromium, actual report app8091,
Goo8088, direct readonly ClickHouse8123, and read-only local Docker copy of the
published full capture. Preserve Docker config; clear endpoint/context/TLS
selectors and use unix:///var/run/docker.sock (the script does this). It leaves
traffic running. Obtain the exclusive browser slot before reproduction. This
receipt describes one accepted corrected run, not a request to repeat it.

`ledger.json` atomically checkpoints each invocation before the real action and
records completion only after its asserted outcome. Ten dense coordinator events
cover generate, collect, inspect, capture and reopen. The final settled marker is
only the bounded coordinator: no pending owned call or finite POST body. It is
not a global collector drain, proof of telemetry completeness or full-trace claim.

The successful run independently observed one matching span and checked exact
source/API/local-inspector/waterfall membership in the original fixed bounds.
Nine direct ClickHouse HTTP probes comprise seven subject-correlation polls and
two exact-window membership reads. Three Goo Explore POSTs read question/trace/
linked-log context; one Capture POST makes a fresh source read and publishes
custody. Five capture/store GETs and 29 other metadata GETs are separate. Backend
SQL statement counts inside Goo were not measured. Frozen reopening adds no
telemetry POST. See `acceptance-summary.json` for IDs, hashes, exact counts and
actual frames; `trace.zip` retains the accepted browser trace.

The initial failed run is in `first-failure-diagnostics/`. Its JS JSON
parse/stringify removed escaped slashes retained in Go json.RawMessage rows,
changing the content-ID preimage. Typed Go unmarshal/remarshal with ID empty and
standard SHA256 verifies the original exact content ID without using production
captureDigest. The original failure has no capture completion or settled event.
Its frame is excluded from the accepted frame set.

No Go test campaign or Bombadil was run by this journey. Root owns the separate
gotogether adapter/model check; the external Bombadil adapter was not tested.
