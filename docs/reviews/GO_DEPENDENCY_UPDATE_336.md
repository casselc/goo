# Go retention dependency update — 336

## Implemented scope

On 2026-11-05, official Go download metadata still reports **Go 1.27.3** as
latest stable. No compiler, language directive or Docker pin changes are needed.
The module proxy reports pgx **v5.12.0**, x/text **v0.43.0** and x/crypto
**v0.58.0**. Root and verification now select the first two explicit versions.
Their actual `go list -m all` graphs contain **no x/crypto**: upstream pgx moved
SCRAM derivation to the standard `crypto/pbkdf2`, so tidy removes that dependency
instead of upgrading an unused package. pgx requires Go 1.25 and x/text Go 1.26;
the existing Go 1.27 baseline accommodates both.

Only root and verification module/sum files change. Tidy also corrects existing
verification direct/indirect classifications for OTel and YAML without changing
their versions. The checksum closure includes pgx's puddle v2.2.3/testify v1.11.1
and selected x/sync v0.24.0; these are upstream graph/test dependencies, not new
product pool or concurrency machinery. Other four module graphs are untouched.

## Upstream changes relevant to acceptance

Reviewed the checksum-verified pgx v5.12.0 `CHANGELOG.md` and SCRAM source.
The latest release fixes narrow simple-protocol `SendBatch` SQL injection and
required-auth/channel-binding bypass conditions. Goo uses Simple Protocol but
has no `SendBatch` use; no exploitability conclusion is drawn from version names.
The intervening v5.9.2 fixes dollar-quoted placeholder confusion in simple queries;
Goo's fixed retention SQL does not use dollar-quoted literals.

The new timestamp text codecs preserve instants but can change returned time
locations, and reject invalid/out-of-range dates more strictly. Native XTDB tests
must validate actual timestamp/read-basis behavior. Added `Rows.Conn` and
`BatchResults.Tx` methods require custom implementers to update; Goo does not
implement those driver interfaces. Pool cancellation changes do not justify
introducing a pool. Goo retains its existing Simple Protocol, await-token,
context and `errors.Is(err, pgx.ErrNoRows)` behavior without compatibility shims.

## Validation and security assessment

- Explicit `go get` then `go mod tidy` succeeded sequentially in root and
  verification through the managed proxy; actual selected graphs inspected.
- `git diff --check` passes. No builds/tests or driver connections were run by
  this reviewer before coordinator release; native/root/semantic acceptance is
  recorded in [VALIDATION.md](../../VALIDATION.md).
- Optional replaced-root consumer check
  `go -C verification/weaving list -mod=readonly ./...` passes for both local
  packages with empty stderr, without editing its module/sum files or compiling
  woven executables. Its readonly selected graph already resolves pgx v5.12.0,
  text v0.43.0 and crypto v0.58.0 (the latter remains for its other dependencies).
  This checks readonly package resolution, not woven tests.
- Resumed the two previously incomplete example checks sequentially under Go
  1.27.3 with `-p=1`/`GOMAXPROCS=2`: operation-app passes **1 test**, zero
  failures/skips (package 0.010s); report-app passes **2 tests**, zero
  failures/skips (package 0.013s). These are ordinary module checks, not woven
  executable/image validation. Neither example module graph was edited.
- Installed **govulncheck v1.9.0** explicitly to the workspace tool directory,
  sequentially with `-p=1`/`GOMAXPROCS=2`. Binary build information confirms
  `golang.org/x/vuln v1.9.0`, built with Go 1.27.3; the tool is outside product
  module graphs.
- Actual command `govulncheck -json -db https://vuln.go.dev ./...` inspected the
  root source graph at symbol level, then exited **1** because fetching
  `https://vuln.go.dev/index/modules.json.gz` returned **Forbidden**. The earlier
  official DB-index probe also received managed-proxy `403 Forbidden`.
  JSON output identifies scanner v1.9.0, Go 1.27.3 and the selected SBOM, but no
  database snapshot or vulnerability verdict. The assessment is **blocked**, not
  clean. TLS/proxy remained enabled, with no mirror, bypass or configuration
  change attempted by this reviewer.

Sources: [Go releases](https://go.dev/dl/?mode=json),
[pgx release](https://github.com/jackc/pgx/releases/tag/v5.12.0),
[module proxy](https://proxy.golang.org/github.com/jackc/pgx/v5/@v/v5.12.0.mod).
The GitHub API endpoint was denied by managed policy; the official proxy archive
provided the release notes and source without bypassing that policy.

Coordinator validation and build metadata are in [VALIDATION.md](../../VALIDATION.md).
The narrow network draft is confirmed saved and awaits user publication; current
runtime policy still excludes `vuln.go.dev`. Compact reviewed results are retained
in ignored `artifacts/go336/summary.json`; raw scanner/test bulk was pruned.
