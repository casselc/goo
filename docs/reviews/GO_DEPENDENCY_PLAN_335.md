# Go dependency maintenance proposal — 335

Status: read-only assessment, not an implemented dependency update. The 335
saved-view ownership work remains independent of this maintenance proposal.
The driver closure is implemented in [336](GO_DEPENDENCY_UPDATE_336.md); the
security assessment remains pending database access.

## Current graph and relevance

Go 1.27.3 is the active compiler and all three Docker build pins agree. The six
modules keep the Go 1.27.0 language directive; a patch update does not require a
new language directive. The latest-version observations below are the managed
proxy results recorded during [334](GO_RUNTIME_REVIEW_334.md), not a new lookup
or an advisory scan performed for this proposal.

| Dependency | Root / verification declaration | Latest observed in 334 | Product use |
| --- | --- | --- | --- |
| pgx/v5 | 5.7.6 / 5.7.6 | 5.12.0 | Optional XTDB PostgreSQL-wire retention, evidence and tail storage |
| x/crypto | 0.37.0 / 0.37.0 | 0.58.0 | pgx SCRAM authentication imports `pbkdf2` |
| x/text | 0.24.0 / 0.24.0 | 0.43.0 | pgx SCRAM authentication imports `secure/precis` |

The cached pgx 5.7.6 module declares those crypto/text versions, explaining their
presence in both graphs. Goo has no direct crypto/text imports in its own Go
source. Updating example modules does not update the root graph; operation/report
examples already declare crypto 0.58.0 and text 0.43.0, while verification/weaving
has its own graph. Confirm actual selected versions with `go list -m all` in each
affected module during implementation rather than assuming declarations determine
every selected version.

`retention.Store.connect` explicitly selects pgx Simple Protocol for XTDB,
passes the supplied context, and optionally sets an await token. Error handling
also relies on `pgx.ErrNoRows`. These are the useful integration boundaries to
validate after a driver update; a PostgreSQL-only smoke test does not establish
XTDB wire and transaction compatibility.

## Next finite maintenance cut

1. On a responsive executor, inspect the current pgx 5.x release notes and module
   requirements, refresh version metadata through the managed proxy, and choose
   explicit versions. The prerelease has no old Go compatibility floor. Prefer
   the current supported driver release if the integration checks below pass;
   do not step through old minor releases solely for compatibility.
2. Update only the root pgx/crypto/text dependency closure, then reconcile the
   verification graph that replaces the root module. Review both module/sum
   diffs. Do not run unrestricted `go get -u ./...`, edit generated module caches,
   or update all six graphs merely to equalize version strings.
3. Run targeted retention and HTTP tests, then root tests, actual XTDB integration,
   and focused semantic verification sequentially. Explicitly set the existing
   `GOO_XTDB_TEST_URL` for native tests and inspect skip counts; ordinary root
   tests skip those tests when the URL is absent. Use a disposable test identity
   and preserve the current lab's kept data, volumes and stream entries.
4. Run a pinned govulncheck against the actual selected root graph and relevant
   executable paths, recording the database/tool version and reachable call
   findings. Evaluate verification and examples separately as development/runtime
   contexts require. A tool or advisory lookup failure is unavailable assessment,
   not a clean result. Resolve actionable findings before declaring maintenance
   complete.

Acceptance should cover retained manifest/citation round-trips, pinned read basis
and await token, correction/retry/concurrent publication, no-row versus store
unavailability, catalog pagination, tail evidence, and shared readers. Existing
native tests include `TestXTDBRetainedEvidenceIntegration`,
`TestXTDBInterruptedPublicationAndRetry`, `TestXTDBConcurrentRequestRecovery`,
`TestXTDBCorrectionRetryAndReadBasis`, `TestXTDBCatalogAndPinnedDocument`,
`TestXTDBTailPublication` and `TestXTDBTailSharedReaders`; HTTP coverage includes
`TestXTDBHTTPInjectedCitationBasisAndOfflineDocument`. Preserve their real native
preconditions rather than substituting a mocked driver for upgrade acceptance.

The two example-module checks interrupted/not started during 334 remain separate
validation debt. Resume them sequentially before broader example dependency work;
they do not block measuring saved-view metadata restoration in 335. The 334 host
stall is an observed environment incident, not evidence that a dependency version
or the Go runtime is defective. Avoid simultaneous full suites, image builds,
native browsers and cold multi-module compilation.

## Standard and extended library use

The code already uses `maps.Equal`/`maps.Clone`, `slices.Clone`/`SortFunc`,
iterator-based `slices.Sorted(maps.Keys(...))`, `strings.SplitSeq`, `bytes.Lines`,
`errors.Join` for server teardown, and `testing/synctest` for deterministic HTTP
deadline checks. These are concrete modern facilities with useful ownership or
allocation semantics. Keep using them where their contracts fit; no bulk style
conversion is needed.

`sync.WaitGroup.Go` could simplify existing correctly paired Add/Done code, but
the audited fleet read already isolates peer failures and bounds membership.
Changing that code without a concurrency defect or measurable maintenance gain
does not advance this cut. In particular, replacing it with `errgroup` would need
an explicit decision about fail-fast cancellation: peer failure currently must
not erase the selected instance's result. No new pooling, unsafe tricks, generic
public API, or source-read parallelism is justified by version availability alone.

Although modern Go has a standard PBKDF2 API, the SCRAM implementation belongs to
pgx. Do not fork or locally rewrite driver authentication to remove an indirect
dependency. Inspect upstream changes and adopt the reviewed driver release.

## Assessment performed

Read the current six module declarations, the cached pgx 5.7.6 module/authentication
source, root retention connection code and native integration preconditions,
representative modern standard-library uses, and the 334 runtime review. No
build, test, driver connection, version download or module edit was performed.
No installed `govulncheck` executable was found on PATH and no cached x/vuln
module metadata was found at the inspected module-cache location. No new
vulnerability assessment or absence-of-vulnerabilities claim is made.
