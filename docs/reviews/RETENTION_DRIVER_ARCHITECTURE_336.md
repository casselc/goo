# Retention driver architecture review — 336

Status: dependency/test/workflow source accepted. The coordinator executed native
XTDB, full Go and Hegel checks; see [validation](../../VALIDATION.md). GitHub CI
execution and a clean vulnerability verdict are not claimed.

## Boundaries to preserve

`retention.Store.connect` parses the server-owned URL, selects pgx Simple Protocol
explicitly and opens a context-bound connection. The optional await token is
bounded to 16 KiB and sent on that connection with `SET AWAIT_TOKEN`; each reader
owns its basis. New driver defaults must not replace this protocol choice or turn
per-request read basis into shared mutable connection state.

Publication and correction use transactions, immutable manifest/intent hashes and
request-ID assertions. A retry first reads its committed receipt; only
`errors.Is(err, pgx.ErrNoRows)` establishes absence. Store unavailability or
corruption must not be treated as permission to create new evidence. On a commit
error the owner attempts receipt recovery, retaining the existing idempotency
contract. Concurrent identical requests must converge on one revision, while a
changed intent under the same key must refuse.

The manifest JSON is stored as text to preserve exact numeric tokens, raw row
bytes and missing/empty distinctions. Indexed member columns are projections,
not substitutes for that immutable manifest. Driver modernization does not change
manifest versions, evidence origin, citation/read-basis semantics or public APIs.
No pool, alternate evaluator or generic persistence layer is justified by this cut.

## Focused additional coverage

`TestXTDBDriverStringParameterRoundTrip336` exercises real Simple Protocol
parameters containing apostrophes, backslashes, newline, Unicode and HTML-like
text. It checks Publish/Load, exact stored `ManifestJSON`, annotation note and
indexed service/member payload against the original values, plus decimal-string
nanoseconds and a JSON integer above both 2^53 and signed int64 range. This fills
a driver-encoding gap in otherwise strong native coverage; it is not a version
assertion or a mock of pgx encoding.

## Existing regression set

Run heavy checks sequentially under the pinned compiler. Set
`GOO_XTDB_TEST_URL=postgres://xtdb@127.0.0.1:15432/xtdb?sslmode=disable` for native
checks and inspect individual results; absent configuration makes them skip.

```sh
go test -p 1 ./retention ./cmd/goo -run '^TestXTDB' -count=1 -v
go test -p 1 ./retention ./cmd/goo -run '^(TestRetention|TestEvidenceOriginReadHTTPGuards)' -count=1
go test -C verification -p 1 -run '^TestHegelRetainedSelectionIdentity$' -count=1 .
go test -p 1 ./...
```

The native set includes:

- retained publication/load, system-time history, annotation branches and withdrawal;
- interrupted publication rollback/retry and concurrent same-request recovery;
- concurrent correction/retry, changed-intent refusal and new-connection await basis;
- expired page-selection recovery, catalog filtering and pinned document citations;
- tail publication/recovery, origin-aware shared readers and immutable old reviews;
- HTTP-injected citation verification, independent valid/invalid read tokens,
  offline document reopen and committed tail recovery without the ephemeral bank;
- the additional rich-parameter/exact-token driver round trip above.

The Hegel property checks selected occurrence identity, canonical numeric/time
bytes, duplicate-index refusal and mutation rejection. It is a pure semantic
property; native tests establish actual driver/XTDB behavior. Neither substitutes
for a real XTDB run with the environment variable set.

## Data and review limits

Existing tests append request/test-specific evidence and never require dropping,
truncating or resetting the retained store. Some fixtures deliberately reuse an
identical immutable manifest; request keys and catalog environments isolate runs.
The new case derives unique capture/request IDs and performs no destructive
cleanup. Preserve volumes and genuine saved evidence when starting the old service.

This source audit has not executed tests, changed driver code or checked reachable
vulnerabilities. Review the selected module graph, source/sum diff and release
requirements separately. A vulnerability tool failure is unavailable assessment,
not a clean security result. pgx upgrade validation does not establish a newer
XTDB server release or production retention guarantees.

## Selected graph review

Reviewed the root/verification module and checksum diffs: pgx is v5.12.0 and
x/text v0.43.0 in both graphs, while x/crypto disappears because upstream now
uses standard `crypto/pbkdf2`. No application pooling/concurrency or protocol
configuration changes accompany the upgrade. Verification's OTel/YAML direct
classification changes retain their versions. The Go 1.27 baseline satisfies
upstream's reported Go 1.25/1.26 requirements. No architecture blocker is found
in this bounded graph change; native execution/security results remain pending.

The rich annotation note uses only admitted multiline text: the current meaning
validator explicitly allows newline/tab/carriage return in notes. Service and
row strings are ordinary retained JSON; no validators were relaxed. Test code
was formatted and `git diff --check` passes; no test was executed by this reviewer.

## CI gate review

Reviewed actual `retention-driver.yml`: accepted `.go-version`, existing digest-pinned
XTDB Compose service (`up -d --no-deps --wait --wait-timeout 180 xtdb`), an explicit
local test URL and the full `^TestXTDB` set in retention and cmd/goo. Require a
finite job timeout, native failures/skip refusal, failure logs and an always-run
service cleanup on the disposable runner; never start/reset unrelated services.
A pinned govulncheck v1.9.0 root scan should fail on findings or unavailable
assessment, without continue-on-error. Its database denial locally remains
unavailable assessment until an actual permitted scan succeeds. A workflow can
make its job blocking; branch-protection required-status settings are separate
and must not be claimed configured by the YAML. Coordinator reports native Go
and Hegel passes; this reviewer did not execute those commands independently.
The workflow itself has not run locally; GitHub execution remains unobserved.
