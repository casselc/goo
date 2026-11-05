# Go runtime and library review — 334

## Verified baseline

On 2026-11-05 the installed `/workspace/.toolchains/go/bin/go version` reports
`go1.27.2 linux/amd64`; its VERSION file records 2026-10-02. Root go.mod declares
Go 1.27.0, .go-version pins 1.27.2, Dockerfile uses golang:1.27.2-alpine, and CI
reads .go-version. Local API history confirms bytes.Lines in Go 1.24
(`api/go1.24.txt:3`), stable testing/synctest in Go 1.25, and container-aware
GOMAXPROCS defaults in Go 1.25 (`doc/godebug.md`). The unified directive already
permits these facilities; there is no older consumer compatibility floor here.

A real request to https://go.dev/dl/?mode=json through the configured managed
proxy returned **go1.27.3** and go1.26.10 as stable releases. Thus the installed
1.27.2 is available but is not the latest stable patch. The initial sandboxed
request could not connect to the proxy; the normal approved escalation succeeded.
The approved patch alignment now pins .go-version and all three Dockerfiles
(root, examples/operation-app and examples/report-app) to 1.27.3.
The official linux/amd64 archive was checked against the manifest size
(70,588,678 bytes) and SHA256
`65efbc80fdc1fa2cd2631cad5b481777a68e8725f34e89f50fc6ac714e379b40`,
extracted into a staging directory, and its actual executable version verified
before atomically replacing the toolchain symlink. The active executable now
reports `go version go1.27.3 linux/amd64`. The previous 1.27.2 installation remains
available for rollback. CI follows .go-version; go.mod retains its Go 1.27.0
language directive. No dependency updates are included this round. A current-pin audit found no other
active 1.27.2 compiler pins: older VALIDATION/ENVIRONMENT_VALIDATION entries are
historical check records, and test_verification_runner.py uses 1.27.2 as mocked
version-check fixtures rather than selecting a compiler.

The six independently versioned modules to validate under the active 1.27.3
compiler are root `.`, `instrumentation`, `examples/operation-app`,
`examples/report-app`, `verification`, and `verification/weaving`. All retain
`go 1.27.0` language directives. Native verification modules retain their separate
dependencies and prerequisites; no heavy validation was started during alignment.

## Implemented bounded improvement

The exploration reader already fully buffers a response through ReadBounded,
which rejects bodies above 8 MiB before parsing. Iterating that existing buffer
with bytes.Lines replaces a redundant bufio.Scanner and its scratch/growth buffer.
It also removes Scanner's independent token-size failure at exactly 8 MiB: one
valid object occupying the entire allowed response now succeeds, as do blank
lines at the inclusive bound. This is a concrete boundary correction, not a
claim of measured latency improvement.

Object-only JSON validation, exact RawMessage bytes, detached per-row copies,
whitespace trimming, 1001-row ceiling, upstream HTTP settings, cancellation,
response closing and all-or-error body-read behavior remain in the same reader.
There is no new API, dependency, concurrency or query orchestration. Tests cover
an exact 8 MiB object, exact-bound whitespace, final unterminated object, existing
1001/1002-row boundaries, byte overflow, invalid JSON and transport failures.

## Extended library inventory and decisions

Read-only proxy metadata returned x/crypto v0.58.0, x/text v0.43.0 and pgx/v5
v5.12.0 as current versions. Root currently selects crypto v0.37.0, text v0.24.0
and pgx v5.7.6; verification also selects the older crypto/text pair. Both example
application modules already declare crypto v0.58.0 and text v0.43.0. These are
separate module graphs: updating an example does not update the server.

Version lag alone is not evidence of an exploitable vulnerability, and version
metadata is not a vulnerability scan. Recommend a separately reviewed root and
verification dependency update with pgx integration and govulncheck, resolving
actual module graphs and relevant advisories before choosing versions. No blind
@latest update was performed and no clean security result is claimed.

Existing queryhttp deadline tests already use synctest for deterministic fake
clock progression; rewriting them adds no value. Do not add errgroup to parallelize
source reads without a measured task: it would change source pressure and current
ordering. No pooling, unsafe, JSON codec substitution, public generics layer or
runtime tuning is justified by this review.

## Validation

- `go test ./exploration -run TestRecordSource -count=1`: passed on Go 1.27.2.
- `go test ./exploration ./internal/sourcehttp -count=1`: passed on Go 1.27.2
  (exploration 16.098 seconds). The ordinary sandbox forbids httptest loopback
  sockets; the successful rerun used approved escalation.
- Live ClickHouse source integration remains opt-in through its existing test
  environment variables; this review does not claim a live-server pass.

### Scheduled nested-module checks on Go 1.27.3

- `instrumentation`: `go test ./... -count=1 -json` passed: 11 test events,
  zero failures/skips, package elapsed 0.029 seconds. Approved socket escalation
  was needed for its local OTLP httptest server; no external collector is required.
- `examples/operation-app`: declared dependency downloads completed through the
  managed proxy, but the check then encountered broad host/IO stalls (even a
  separate two-line tail could not complete). The coordinator requested
  interruption to release pressure; the command exited 130. No completed test
  result is claimed. This is blocked validation, not a product failure or skip.
- `examples/report-app`: not started after the coordinator stopped nested checks
  during the same host/IO incident; blocked, not passed or skipped.

No dependency versions were edited. The example checks remain follow-up work on
a responsive executor; ordinary Go tests also do not validate woven binaries.
