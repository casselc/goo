# Validation record

## OpenAPI contract, CI conformance and browsable docs (2026-10-22 UTC)

[API contract/limitations](docs/API.md) and the [executable docs journey](docs/journeys/API_DOCS.md)
record 64 operations: 58 Go operations, five Python-only /api operations and Python
/health. The OpenAPI 3.1 artifact has 324 schemas, typed Go request/response
envelopes, explicit dynamic payloads, effects, separate Go/browser byte limits and
schema-linked SSE events. Runtime domain validation is unchanged.

Actual checks:

- The standard OpenAPI validator accepts the spec and its local references. Route
  coverage matches all 58 Go registrations/cancellation operations. All 236 frozen
  HTTP cases are accounted for: 222 documented responses validate, 56 admitted
  request payloads validate, and 14 default mux refusals are checked separately.
  Unknown-field, wrong error type and malformed SSE controls are rejected.
- Actual Go reset/source-unavailable gap frames match `testdata/http/tail.sse` and
  validate against event schemas. Batch examples/negative controls also validate;
  existing loopback tail tests retain actual batch/disconnect behavior.
- All four HTTP fixture groups and the new SSE contract pass under race detection
  (HTTP package 1.503s). `go test ./...` passes, including the schema-generator
  package. No new semantic/state transition required Hegel/Quint/Bombadil changes.
- Python API/docs tests pass (5 tests, 0.528s), including source-free legacy explain,
  explicit query evidence, invalid queries and the actual handler's docs/spec routes.
- The real Chromium documentation journey passes at 1280x900 and 390x844: all
  64 operations load, filtering/disclosure and displayed limits work, no mobile
  horizontal page overflow and no page errors. A real tail-contract screenshot
  is saved in `docs/journeys/screenshots/api-docs-tail.png`.
- Go schema generation plus spec assembly is byte-identical on regeneration.
  CI now refuses generated drift and runs spec/route/fixture/negative-control,
  wire/SSE and Python documentation-route checks. The hosted CI workflow itself
  has not been observed running in this session.
- Whitespace and local Markdown targets pass. Isolated validation dependencies are
  pinned in `requirements-api.txt`, outside application runtime dependencies.

Evidence: `artifacts/openapi-{contract,python,wire-race,browser,root}.log`.
No telemetry was inserted and the main workbench UI/protocol was unchanged. The
docs journey serves this checkout's actual Python handler on an ephemeral port;
the existing lab image was not rebuilt, so its server does not yet serve these new
docs routes. Browser-proxy wire fixtures, broader executed coverage for remaining
API owners and C2 final deployed integration acceptance remain open. Further
mechanical JSON adoption is paused for that API contract work.

## C2 fixed exploration JSON adoption (2026-10-22 UTC)

[Fixed exploration contract and diagram](docs/HTTP_BOUNDARIES.md#fixed-exploration-json-adoption)
record shared JSON adoption for compare/inventory/discovery/Explore. Comparison
retains 1 MiB, source routes 16 KiB, and each keeps its validation/error mapping.
Source deadlines and Explore's result-page freeze remain in their existing owners.

Actual checks:

- All 68 static wire fixtures pass before extraction (0.058s): exact statuses,
  bodies and selected headers across valid comparison, source admission/failure,
  malformed/null/array/unknown/nested-invalid/second/trailing inputs, exact/overflow
  bounds, whitespace, content-type behavior, numeric type refusal and mux refusal.
  Comparison retains counts above 2^53. Transport counters prove source-free
  comparison/rejection and one read for admitted failures; deadline assertions
  preserve inventory/discovery/Explore's 15/30/55-second contexts.
- After extraction, race suites pass for HTTP routes (2.381s) and exploration
  (cached), including the same fixtures and existing positive source, comparison,
  snapshot, dependency-isolation and SSE regression tests.
- Hegel's 100 shared planning-JSON cases pass (0.023s).
- `go test ./...` passes. JSON-only changes require no new live database test;
  existing gated live checks skip without their configured URL variables.
- Whitespace and local Markdown targets pass.

Evidence: `artifacts/c2-exploration-json-{before,race,hegel,root}.log`.
No UI/SSE/Quint protocol changed; no screenshots or running images were regenerated.
Tail request decoding is next, preserving its separate error/stream policies.
Other JSON owners and C2 final deployed-journey acceptance remain open.

## C2 native measurement/registry JSON adoption (2026-10-22 UTC)

[Native JSON contract and diagram](docs/HTTP_BOUNDARIES.md#native-measurement-and-registry-json-adoption)
record shared bounded typed decoding and JSON/no-store writing. Native validation,
64 KiB limits, `one JSON object required`, source deadlines, custody and file-only
YAML publication remain in their existing owners.

Actual checks:

- All 68 static native JSON fixtures pass before extraction (0.050s): exact statuses,
  body bytes and selected headers across planning/publication/discovery/execution,
  malformed/null/array/unknown/nested-invalid/second-value input, exact/overflow
  bounds, whitespace, content-type behavior, numeric type refusal and mux refusal.
  Source-read counters enforce rejection/planning/publication before reads. Valid
  execution uses a deterministic injected 503 backend and retains 502 responses;
  file assertions permit only successful explicit publication. Plan windows are fixed.
- After extraction, race suites pass for HTTP routes (3.170s), native measurements
  (3.248s) and lifecycle (5.929s), including the same wire fixtures and existing
  positive execution, dependency isolation, registry/duplicate/reopen, workflow and
  custody clock/identity contracts.
- Hegel's 100 shared planning-JSON cases pass (0.079s).
- `go test ./...` passes. This JSON-only change does not require a new live database
  check; existing gated live tests skip without their configured URL variables.
- Whitespace and local Markdown targets pass.

Evidence: `artifacts/c2-native-json-{before,race,hegel,root}.log`.
No UI/SSE/Quint protocol changed; no screenshots or running images were regenerated.
Fixed exploration JSON is next; other JSON owners and C2 final deployed-journey
acceptance remain open.

## C2 telemetry package JSON adoption (2026-10-22 UTC)

[Telemetry JSON contract and diagram](docs/HTTP_BOUNDARIES.md#telemetry-package-json-adoption)
record shared 64 KiB typed decoding and JSON/no-store writing. Planning/publication
validation, default planning clock, YAML/handoff, revisions and store effects stay
local. Trailing input retains `require one JSON object` and existing 400 behavior.

Actual checks:

- All 68 static wire fixtures pass before extraction (0.196s), freezing statuses,
  exact bodies and selected headers across draft/numeric/package planning and
  publication, including valid, malformed, unknown/nested-invalid, null/array,
  second/trailing values, exact/overflow bytes, ignored content type and mux refusal.
  Explicit-window success fixtures are deterministic. Each case verifies plans and
  rejected input leave an empty registry; successful publication creates one file.
- After extraction, race suites pass for HTTP routes (2.280s) and exploration
  (5.312s), including the same wire fixtures, existing generated-Go authoring,
  numeric capability, publication identity/duplicate/reopen and other HTTP regressions.
- Hegel's 100 generated shared planning-JSON cases pass (0.025s).
- `go test ./...` passes. No live database check was required for this JSON-only
  change; existing gated live tests skip without their configured URL variables.
- Whitespace and local Markdown targets pass.

Evidence: `artifacts/c2-telemetry-json-{before,race,hegel,root}.log`.
No UI/SSE/Quint protocol changed, no screenshots were regenerated and no running
image was rebuilt. Native measurement/registry JSON is next; other JSON owners and
C2's final deployed-journey acceptance remain open.

## C2 histogram HTTP transport adoption (2026-10-22 UTC)

[Histogram transport contract and diagram](docs/HTTP_BOUNDARIES.md#histogram-transport-adoption)
record shared POST/auth/context and bounded-read adoption only. Read errors still
precede status, status precedes overflow, and every response closes immediately
after reading. Typed buckets, limits, per-contract semantics and basis stay local.

Actual checks:

- New contracts pass before extraction (0.279s): complete SQL/auth/settings,
  baselines/raw complex parameters, empty/positive payloads, exact/overflow 8 MiB
  and 1,000-point boundaries, malformed JSON, error precedence, cancellation/error
  identity, close-before-next-contract and no partial later-contract success.
  Two named histogram contracts exercise multiple reads; cumulative bucket counts
  above 2^53 preserve an exact interval count of one.
- After extraction, race suites pass for annotations (13.359s), HTTP routes (3.840s)
  and investigations (cached), with gated readonly ClickHouse checks enabled.
  Both histogram fixture contracts execute through the real table/query/decoder/
  empty evaluator on an absent service/fixed window. No inserts or positive live
  population claims. Existing unit/layout/population semantic refusals still pass.
- Hegel's 100 shared bounded-read cases pass (0.019s).
- `go test ./...` passes. Gated live checks skip without their URL variable.
- Whitespace and local Markdown targets pass. Source inspection confirms all six
  audited ClickHouse readers now delegate POST/body mechanics to the shared helper.

Evidence: `artifacts/c2-histogram-transport-{before,race,hegel,root}.log`.
No UI/SSE/Quint protocol changed; no screenshots or running images were regenerated.
C2 remains active: pure telemetry planning JSON next, then other JSON owners and
final deployed-journey acceptance.

## C2 counter HTTP transport adoption (2026-10-22 UTC)

[Counter transport contract and diagram](docs/HTTP_BOUNDARIES.md#counter-transport-adoption)
record shared POST/auth/context and bounded-read adoption only. Counter reads and
immediate closure still precede status checking; read errors beat status, which
beats overflow. Baselines, per-contract evaluation, ratios, fleet isolation and
semantic precision/temporality/reset rules remain unchanged.

Actual checks:

- New contracts pass before extraction (0.605s): complete SQL/auth/settings and
  raw parameters, baseline calculation, empty/positive payloads, exact/overflow
  8 MiB and 3,000-row boundaries, malformed JSON, cancellation/error identity,
  error precedence, close-before-next-contract and no partial later-contract success.
- After extraction, race suites pass for annotations (14.065s), HTTP routes (5.112s)
  and investigations (cached), with the gated readonly ClickHouse check enabled.
  Both rate contracts execute against the real sum table with an absent service
  and fixed window. No inserts or positive live-data claims; positive fixtures
  retain the cumulative delta of three and existing fleet tests still pass.
- Hegel's 100 shared bounded-read payload/budget/error cases pass (0.064s).
- `go test ./...` passes. Gated live checks skip without their URL variable.
- Whitespace and local Markdown targets pass.

Evidence: `artifacts/c2-counter-transport-{before,race,hegel,root}.log`.
No UI/SSE/Quint protocol changed, screenshots were not regenerated and no running
image was rebuilt. Histogram transport is next, then route-specific JSON adoption;
C2's final deployed-journey acceptance remains open.

## C2 operation-span HTTP transport adoption (2026-10-22 UTC)

[Operation-span contract and diagram](docs/HTTP_BOUNDARIES.md#operation-span-transport-adoption)
record adoption of shared POST/auth/context and bounded body mechanics. Raw
parameters, status-before-body, closure, typed decoding, row/window limits,
conformance, events, deduplication, exact integer strings and basis remain local.

Actual checks:

- New transport contracts pass before extraction (0.173s), pinning SQL/auth/exact
  settings and complex raw parameters, empty/valid/duplicate payloads, exact/overflow
  8 MiB and 1,000-span boundaries, malformed JSON, unread non-200 bodies, closure,
  cancellation and error identity. Exact duration strings above 2^53 are preserved.
- After extraction, race suites pass for annotations (8.791s), HTTP routes (5.888s)
  and investigations (cached), with the gated readonly real ClickHouse test enabled.
  The new test executes `FetchOperations` against an absent service/fixed window:
  real query/decoder/empty conformance integration, without inserts or positive live
  population claims. Existing fixtures cover positive evidence and semantic refusals.
- Hegel's 100 shared bounded-read payload/budget/error cases pass (0.042s).
- `go test ./...` passes. Gated live checks skip without their URL variable.
- Whitespace and local Markdown targets pass.

Evidence: `artifacts/c2-operation-transport-{before,race,hegel,root}.log`.
No UI/SSE/Quint protocol changed, no screenshots were regenerated and no running
image was rebuilt. C2 remains active: counter transport next, then histogram and
route-specific JSON adoption. Final deployed-journey acceptance remains open.

## C2 prebuilt HTTP transport adoption (2026-10-22 UTC)

[Prebuilt transport contract and diagram](docs/HTTP_BOUNDARIES.md#prebuilt-http-transport-adoption)
record adoption of shared POST/auth/context and bounded reads only. Raw parameter
values, status-before-body ordering, closure, typed decoding, window/row limits,
deduplication, evaluation identity and warnings stay family-owned.

Actual checks:

- New prebuilt transport contracts pass before extraction (0.683s): request bytes,
  exact settings/auth, unchanged complex parameters, empty/valid/duplicate payloads,
  exact/overflow 8 MiB and 1,000-row boundaries, malformed JSON, unread non-200
  bodies, body closure, cancellation and error identity.
- After extraction, race suites pass for prebuilt (7.269s) and HTTP routes (4.585s)
  with the gated readonly ClickHouse test enabled. The new integration test executes
  the actual `FetchHTTP` query and evaluator on a deliberately absent scope/fixed
  window; it inserts nothing and does not prove positive live-population coverage.
- Hegel's 100 generated shared bounded-read cases pass (0.019s).
- Full `go test ./...` passes. Gated live checks skip without their URL variable.

Evidence: `artifacts/c2-prebuilt-transport-{before,race,hegel,root}.log`.
No UI/SSE/Quint protocol changed; no screenshots or running images were regenerated.
C2 remains active: operation spans next, then counters/histograms and remaining
route-specific JSON adoption. Final deployed-journey acceptance remains open.

## C2 remaining-work acceptance audit (2026-10-22 UTC)

[Audit and ordered completion proposal](docs/HTTP_BOUNDARIES.md#c2-acceptance-audit--2026-10-22)
were checked against current production request construction, bounded reads,
route decoders, startup environment reads and existing test names. Four ClickHouse
readers remain unadopted: prebuilt HTTP, operation spans, counters and histograms.
Nine HTTP files retain local request-body decoder mechanics. Existing semantic
checks do not substitute for missing transport/wire contracts.

This documentation-only cut updates stale next-step text and keeps C2 active.
Prebuilt HTTP transport is next; remaining readers and route-specific JSON follow.
Final deployed real-source/browser acceptance remains open. No code, model, UI or
runtime image changed; no test suite or browser journey was rerun for this audit.
Whitespace and local Markdown link targets were checked.

## C2 record source transport adoption (2026-10-22 UTC)

[Record transport contract and diagram](docs/HTTP_BOUNDARIES.md#record-source-transport-adoption)
document adoption of shared POST/auth/context and bounded-body mechanics. Record
escaping, readonly/resource settings, status-before-body ordering, response closing,
JSON-object validation, scanner behavior and 1,001-row guard remain in the adapter.
Counts, evidence basis, capture and tail lifetimes are unchanged.

Actual checks:

- New request/response/error contracts pass before extraction (0.168s), including
  exact/overflow 8 MiB and row limits, invalid objects, preserved uint64 spelling,
  body closure, cancellation, read-error identity and unread non-200 bodies.
  Scanner-size failure retains the existing non-nil empty slice plus scanner error.
- After adoption, race suites pass for exploration (11.808s), HTTP routes (6.219s)
  and measurements (4.951s), with both gated real-source tests enabled against the
  existing local ClickHouse lab. The new readonly record integration test roundtrips
  all ASCII controls, DEL, backslashes, Unicode and punctuation through `numbers(3)`.
  No records are inserted. These tests do not establish ingestion completeness.
- Hegel's 100 generated bounded-read payload/budget/error cases pass (0.023s).
- `go test ./...` passes; live checks skip when their URL variable is absent.

Evidence: `artifacts/c2-record-transport-{before,race,hegel,root}.log`.
No UI/SSE/Quint protocol changed, no screenshots were regenerated and no running
image was rebuilt. Remaining transports/routes need a C2 acceptance audit.

## C2 native source HTTP mechanics and readonly ClickHouse validation (2026-10-22 UTC)

[Native transport contract/diagram](docs/HTTP_BOUNDARIES.md#native-source-http-mechanics--first-transport-adoption)
records first adoption of `internal/sourcehttp.Post` and `ReadBounded`. Native
compilation, parameter escaping/query flags, client fallback, response closure,
status/error mapping, row/scanner guards, evaluation, basis and custody stay in
`measurements.Source`. Other family transports are not adopted yet.

Actual checks:

- New native contracts pass before extraction (0.155s). They pin exact SQL/auth/
  URL flags and escaping without input mutation, empty/raw-line behavior, exact
  10,000-row and 8 MiB boundaries, overflows, single-line scanner failure, 201/503,
  transport errors, cancellation and body closure. Read failure still precedes a
  non-200 status error. All refused reads return nil rows rather than partial data.
  The same cases pass after extraction.
- Affected race suites pass: measurements (3.644s), HTTP routes (2.343s), including
  native discovery, lifecycle/counter/histogram checks and existing source injection,
  planning wire fixtures, budgets, SSE, captures and optional-store unit contracts.
- Hegel's 100 generated payload/budget/read-error cases pass (0.021s): exact
  in-budget bytes, consumption at most limit+1, overflow detection and no partial
  payload on overflow/read failure. Request timeouts and body lifetime remain
  caller-owned; the helper adds no retries or generic status policy.
- The gated readonly real ClickHouse test passes with race detection against the
  existing local lab (0.10s live test; 2.598s package including native contracts).
  Complex literal-backslash/newline/tab/CR/NUL/quote/Unicode/plus/percent parameters
  roundtrip through `numbers(3)`. A native gauge source query/evaluation passes
  `ValidateResult`, and native instrument discovery succeeds. These scope queries
  may legitimately have no data; no telemetry was inserted and no completeness
  claim is made. The fixture's nil client also exercises the existing 15s fallback.
- `go test ./...` passes. Live checks explicitly skip without
  `GOO_CLICKHOUSE_TEST_URL`; configured local test credentials are disposable lab
  values. Whitespace and changed local Markdown targets pass.

The running application image was not rebuilt. No public query/execution-state,
stream, stored schema or UI protocol changed, so no new browser screenshot or Quint
model was required. C2 remains active: freeze record adapter escaping/status/row
contracts next, then adopt only matching primitives and retain its distinct rules.

## C2 optional XTDB startup wiring and real pinned-citation isolation (2026-10-22 UTC)

[Optional retention ownership and diagram](docs/HTTP_BOUNDARIES.md#optional-retention-and-document-citation-startup-wiring)
now describe one injected optional URL snapshot shared by retention routes and
workbench citation checks. `retentionConnection` carries no await token; each
operation constructs a fresh store value with its request/card basis. Status,
30-second contexts, exact references, retry/recovery and file-document loading
retain their existing owners. Both HTTP families stop rereading process config.

Actual checks:

- Existing optional-store/document/population/capture checks pass before injection
  (0.052s). After injection, HTTP tests retain unconfigured 503, configured-invalid
  400 and configured-unavailable recovery 502 behavior, no-store status and strict
  request validation. Independent registrations retain configured/unconfigured
  status after ambient environment changes.
- The local XTDB HTTP health endpoint returned 200. The new gated real-instance
  test passes with race detection (1.91s test / 2.949s package): selected ordinary
  records are published, a review corrected, and a document created with both old
  and current exact review IDs plus the committed lower-bound token. Changing
  ambient URL configuration and the caller's connection value does not redirect
  the registered verifier.
- Concurrent valid/invalid await-token HTTP requests return the old pinned review
  (200) and reject the invalid basis (404), respectively. A fresh unconfigured
  owner reopens the persisted document with exact review/basis fields; new citation
  publication without the optional store remains refused (400). No global mutable
  basis was introduced. During fixture preparation the retained-record adapter
  refused a numeric timestamp; the final fixture uses its existing exporter
  decimal-string timestamp contract, without changing product decoding.
- The complete `go test -race ./cmd/goo` suite passes with the real XTDB test enabled
  (3.501s), alongside source injection, 32 planning fixtures, budgets, SSE and capture
  checks. `go test ./...` also passes; real-instance tests are explicitly gated
  without `GOO_XTDB_TEST_URL`. Other retention-package live campaigns were not rerun
  in this cut. Whitespace and changed local Markdown targets pass.

No database schema, bitemporal semantics, public execution-state or UI protocol
changed. No new Hegel/Quint model or screenshot was needed for configuration wiring;
actual HTTP/XTDB isolation was checked instead. The running application was not
rebuilt/redeployed, and no XTDB restart was performed. C2 remains active; bounded
native source transport mechanics and parity/integration evidence are next.

## C2 operation readers/investigations and unified startup connection (2026-10-22 UTC)

[Operation adoption contract/diagram](docs/HTTP_BOUNDARIES.md#operation-get-and-investigation-adoption)
records the extracted `registerOperationReads` and injected investigation connection.
One `clickHouseConnectionFromEnvironment` snapshot now supplies all Go HTTP
ClickHouse routes, replacing the separate native/record environment builders.
Typed adapters, operation plans, HTTP instrumentation bindings, custody and source
transport remain family-owned. Server startup now loads/validates configuration
and registers route families without embedding their GET implementations.

Actual checks:

- Existing operation investigation trusted-source/optional-error/offline-reopen
  tests pass before migration (0.018s). After moving the unchanged GET handlers,
  and before changing connection wiring, 11 exact validation cases plus the
  declaration response pass (0.012s). Error text, 400 status, text/plain/no-store/
  nosniff headers remain pinned; source failures retain 502 JSON.
- New tests hold two operation/investigation registrations with distinct backends,
  credentials and clients at once. All four source-reading GET endpoints execute
  against the assigned backend, including default HTTP window/package selections.
  Declarations remain source-free; readonly/auth/client wiring is checked. Caller
  connection mutation and ambient trap configuration do not redirect requests.
- Already-cancelled GETs avoid backend reads. Investigations retain revision and
  parent-question 409 checks before reads. After backend closure, fresh handlers
  reopen persisted captures with no plan registry and no telemetry requests.
  Existing optional-rate/histogram errors and failed-refresh receipt checks remain.
- Shared startup tests pin configured/default endpoint, credentials, client and
  typed source adapters against later environment mutation. Inspection finds
  ClickHouse environment reads only in the startup connection constructor among
  HTTP/dependency files. This snapshot is configuration, not a shared query engine.
- `go test -race ./cmd/goo` passes on final cases (2.180s); `go test ./...` passes.
  Existing 32 planning wire fixtures, native/record/capture source injection,
  query-budget, tail and optional-store tests remain included. Whitespace and
  changed local Markdown targets pass.

No public query/execution-state, JSON/SSE policy or UI changed; no new Hegel/Quint
model or screenshot was required for this wiring/registration cut. The running
stack was not rebuilt. GET readers keep incoming contexts/shared client timeout
and remain outside the POST query budget. C2 remains active: optional XTDB
configuration lifetimes are next, followed by bounded transport extraction and its
own parity/integration evidence.

## C2 capture/live-comparison source dependency injection (2026-10-22 UTC)

[Capture dependency ownership and diagram](docs/HTTP_BOUNDARIES.md#capture-and-live-comparison-dependency-ownership)
now document configured source values for saved capture and live comparison.
Startup constructs one record source for exploration/tail and capture/live routes.
Neither capture handler rereads ClickHouse environment. Source adapters, storage
roots, lineage checks, 55-second read contexts, draft budgets and freeze semantics
remain in their existing owners.

Actual checks:

- Existing capture/paging, live comparison exact freeze/admission/replacement and
  native-window source-free checks pass before migration (0.059s). They retain
  assertions after switching registration tests to explicit source values.
- A new loopback test holds two capture routers, backends and storage roots at
  once, with distinct source totals and credentials. It checks capture/live reads,
  readonly/client wiring, isolation from caller source-value mutation and zero
  requests to an ambient trap endpoint. Changed refresh questions retain 409
  before source reads; already-cancelled captures retain 502 without backend reads.
- After both backends close, each owner refuses the other's token (410), freezes
  its own held snapshot without source reads (201), and returns the same receipt
  on repeat freeze. Fresh registrations reopen both original/frozen captures from
  disk while their prior draft tokens return 410. Source counters stay unchanged.
- `go test -race ./cmd/goo` passes (2.044s), including existing native/record/SSE
  injection, 32 planning wire fixtures, query budgets, optional-store and capture
  contracts. `go test ./...` passes. Whitespace and changed local Markdown targets
  pass. The frozen store/identity contracts and capture source adapters are unchanged.

No public execution-state, stream, persistence or UI protocol changed; no new
Hegel/Quint model or screenshot was needed for this source-wiring cut. The running
stack was not rebuilt. Drafts remain expiring registration-local capabilities;
injection does not make them durable or global. C2 remains active; operation GET
and investigation source dependencies are next, followed by optional-store wiring
and bounded source transport adoption with its own evidence.

## C2 exploration and tail startup dependency injection (2026-10-22 UTC)

[HTTP boundary inventory/diagram](docs/HTTP_BOUNDARIES.md#startup-dependencies-native-and-explorationtail-adoption)
now includes record sources. `registerExploration` and `registerTail` receive a
configured `exploration.Source`; `explorationSourceFromEnvironment` resolves the
existing endpoint fallback, credentials and client at startup. Inventory, service
discovery, Explore and SSE no longer reread process environment. Native startup
wiring is unchanged. Bounded query and streaming owners remain separate.

Actual checks:

- Existing exploration/inventory/comparison/tail validation, source-failure,
  framing/disconnect and admission/cancellation tests pass before injection
  (0.121s). They retain their assertions after switching test wiring to explicit
  source values. Tail still has four slots, reset/batch/gap framing, 60s context,
  two-second polling and per-write deadlines; record contexts remain 15s/30s/55s.
- A new loopback test holds two independently configured routers/backends at once.
  It checks inventory/discovery/Explore and actual SSE reset/batch frames, endpoint
  and credential isolation, readonly parameters, client use and stream headers.
  Mutating the caller's source value does not redirect the registered owner.
  An ambient trap receives zero reads; exercising one router leaves the other's
  source counter unchanged. Invalid and already-cancelled record requests avoid
  backend reads, retaining 400/502 responses respectively.
- Startup tests pin localhost fallback, configured credential/client fields and
  snapshot behavior after environment mutation. Configuration mutation is not a
  record/tail reload API. All credential values used by tests are local fixtures.
- `go test -race ./cmd/goo` passes (2.925s), retaining native source injection,
  32 planning wire fixtures, query budgets, optional-store and capture contracts.
  `go test ./...` passes. Whitespace and changed local Markdown targets pass.

No source transport, decoder, public query/stream protocol or UI changed. No new
Hegel/Quint model or screenshot was needed for this dependency wiring cut; existing
stream/query state contracts are exercised by the HTTP suite. The running stack
was not rebuilt. Result-page snapshots remain process-global and tail slot budgets
remain registration-local. C2 is active; capture/live-comparison dependencies are
next, followed by operation readers and remaining optional-store configuration.

## C2 frozen pure planning JSON contracts and shared mechanics (2026-10-22 UTC)

[Pure planning adoption contract/diagram](docs/HTTP_BOUNDARIES.md#pure-planning-json-adoption)
records the bounded recipe/dashboard cut. Both routes now share typed bounded JSON
decoding and JSON/no-store writing in `internal/httpjson`; they retain semantic
validation, status mapping, exact byte limits and trailing-input messages. Other
route families and streaming responders are unchanged.

Actual checks:

- Thirty-two frozen wire cases pass on the old handlers before extraction (0.029s)
  and on the shared implementation. They compare exact status/body and content-type,
  cache, Allow and nosniff headers for valid/empty/null/array/truncated input,
  top-level/nested unknown fields, extra values/malformed tails, whitespace,
  exact/over-limit bodies, request content-type compatibility, empty objects and
  method refusal. Compact fill/padding fields reconstruct large request bodies.
- Recipe's 16 KiB limit and “require one JSON object” message remain distinct from
  dashboard's 256 KiB limit and “expected one JSON object”. First-value oversize and
  trailing-whitespace oversize retain their different error bodies and 400 status.
  Null still reaches domain validation. The helper does not introduce media-type
  enforcement, generic null/duplicate rejection or universal HTTP errors.
- HTTP race tests pass (1.879s), including all 32 fixtures, existing source-free
  recipe plans, dashboard registry declarations and unrelated native/SSE/query
  budget/retention/capture contracts. `go test ./...` passes; the new helper is
  exercised through actual route fixtures rather than a mirrored root unit suite.
- Hegel's 100 generated decoder/writer cases pass (0.033s): Unicode/escaped labels,
  typed int64 values above JS safe range, whitespace, exact/insufficient byte limits,
  second values, unknown fields and exact encoded JSON/no-store responses.
  Whitespace and changed local Markdown targets pass.

No source/persistence dependency or browser execution-state protocol changed; no
new screenshot or Quint model/campaign was needed for this mechanical cut. The
running stack was not rebuilt. Writer encode/write errors remain ignored as in the
original routes. C2 remains active; additional source dependency injection is next,
with further JSON adoption and bounded transport extraction still queued.

## C2 endpoint inventory and native startup dependency injection (2026-10-22 UTC)

[HTTP effects, policies and dependency diagram](docs/HTTP_BOUNDARIES.md) inventories
source reads (including GET executions), pure planning, custody/file publication,
live draft/freeze branches, optional XTDB, decoder limits and SSE/query-budget
lifetimes. Native measurement registration now accepts `measurements.Source` from
startup wiring. Endpoint/credentials/client are resolved once, with the existing
localhost fallback; custody remains rooted under the configured server data path.
Native handlers no longer reread process environment. Other families remain queued.

Actual checks:

- Existing native registry/preview/YAML fixture tests pass before the change
  (0.042s) and after injection in the full HTTP race suite. Exact stored bytes,
  revision, POST response/headers, builtin resolution and source-free plans retain
  their existing fixture contracts.
- Two independently injected loopback backends return distinct native gauge
  values and instrument discovery responses. Tests check credentials, readonly
  parameters and explicit client use. An ambient trap endpoint receives zero
  requests; mutating the caller's source value after registration does not redirect
  the handler. Invalid question validation and cancelled requests avoid backend
  reads; cancellation retains the existing 502 response. Gauge/discovery calls do
  not create lifecycle custody files.
- Startup tests pin localhost fallback, credential/client wiring and snapshot
  behavior after environment changes. Secrets are test-only strings, not read or
  printed from configured runtime credentials. Native environment mutation is not
  a configuration reload API.
- `go test -race ./cmd/goo` passes on final source (1.853s), including existing SSE,
  budget/cancellation, capture, retention and publication tests. `go test ./...`
  passes across all 13 root packages. Whitespace and changed local Markdown targets
  pass. Decoder/response helpers and ClickHouse transport are unchanged.

C2 remains active: freeze two pure planning route decoder/response contracts before
sharing matching mechanics, then inject additional families and extract bounded
source transport with its own validation. No browser/UI/execution-state protocol
changed; no screenshot or new Hegel/Quint campaign was needed for this dependency
wiring cut. The running stack was not rebuilt or redeployed.

## C1 custody/native YAML fixtures and bounded consolidation completion (2026-10-22 UTC)

[Publication contract, diagram and acceptance audit](docs/IMMUTABLE_PUBLICATION.md)
now cover all six Go temporary-write/link publication paths. Lifecycle custody and
native YAML use explicit file-only sync; directory sync was not added. Their
validation, event/definition identity, clocks, winner verification and HTTP policies
remain in their callers. Old write/sync/link copies are removed. Bounded C1 is
delivered; C2 endpoint-policy/startup inventory is next.

Actual checks:

- Custody and native YAML fixtures pass before extraction (measurements 0.017s,
  HTTP routes 0.014s). They pin event identity/bytes and exact decimal business and
  received clocks, plus native canonical YAML/revision, exact POST response JSON,
  201 status and JSON/no-store headers. Fresh registry reopen, invalid YAML/request
  fields/trailing JSON, duplicate/no-overwrite and corrupt existing-winner behavior
  are checked after adoption with the same fixtures.
- Custody ignores caller-supplied received time, preserves its first winner even
  when a later call proposes an earlier clock, and separates changed business
  payloads. Sixteen concurrent distinct observation times all return one winning
  event/clock. Corrupt event/truncated/non-integer-clock files are refused without
  repair. Existing received-time trust boundaries are explicitly pinned: a valid
  integer clock edit is not detected by the business-event hash.
- Race-enabled suites pass for immutablefile (1.245s), measurements (1.775s), HTTP
  routes (2.638s), exploration (5.303s), investigations (1.114s), workbench (1.723s).
  Existing native/lifecycle/capture/parent/duplicate tests remain included.
- Expanded Hegel publication checks pass: 100 generated sequences (0.324s), choosing
  either initial sync policy and checking idempotent, file-only and exclusive
  duplicates while retaining first bytes. All 13 root packages pass `go test ./...`.
  Whitespace and changed local Markdown targets pass.
- Native YAML POST retains its historical 201 submitted-definition response even
  for a corrupt existing winner; it does not repair or verify that file, and GET
  registry then reports 400. This refactor pins that compatibility behavior rather
  than implying hardening. Investigation unknown-field policy is also unchanged.

No UI or execution-state protocol changed; no screenshots/browser/Quint campaign
was regenerated for this filesystem-only cut. The deployed stack was not rebuilt.
No sync-failure injection, power-loss experiment, tamper-evident custody or native
duplicate repair is claimed. Follow-up policy changes require separate contracts.

## C1 capture publication adoption and frozen evidence contracts (2026-10-22 UTC)

[Publication contract and diagram](docs/IMMUTABLE_PUBLICATION.md) now cover both
capture stores. Exploration captures use idempotent publication followed by their
existing verified reload. Investigation captures use `PublishNew`, preserving the
original `*os.LinkError` on a duplicate and skipping directory sync on that error,
as before. Encoding, digests, validation, source execution, lineage, catalog/page
rules and size limits stay in the family stores. Old duplicated writers are removed.

Actual checks:

- New exploration/investigation storage fixtures pass on the old publishers before
  extraction (0.039s/0.033s). After extraction they pin the same stored identities,
  raw large integer `18446744073709551615`, independent source-read/receipt clocks,
  exact exploration page JSON, investigation loaded JSON, reopen and refresh
  compatibility. Exploration duplicate publication refuses corrupt winners without
  overwriting them. Foreign cursors/references, truncation, digest mismatch and
  trailing JSON remain refused. Investigation embedded-key-order normalization
  preserves identity. Its historical permissive unknown-field policy is pinned;
  exploration's envelope remains strict.
- Affected race tests pass: immutable publisher (1.319s), exploration (8.080s),
  investigations (1.240s), workbench (2.384s). Existing native metric, inventory,
  comparison, parent/plan compatibility and concurrent capture checks remain in
  those suites. Exclusive publication additionally pins original link-error
  type/existence classification, no overwrite and temporary cleanup.
- Expanded Hegel publication sequences pass: 100 generated sequences (0.053s),
  preserving the first bytes through idempotent and exclusive duplicate attempts.
  `go test ./...` passes across all 13 root packages, including HTTP route tests.
  Changed local Markdown targets and whitespace checks pass.
- No UI or execution-state protocol changed; existing browser/Quint storyboards
  remain applicable and no screenshots were regenerated. The deployed stack was
  not rebuilt for this backend-only cut. No crash/power-loss durability is claimed.

C1 remains active for lifecycle first-observation custody and native measurement
YAML publication. Both currently omit directory sync and have different winner
verification contracts; freeze their fixtures before deciding compatible adoption.

## Disk cleanup and C1 immutable publication first adoption (2026-10-22 UTC)

[Publication contract, diagram and store inventory](docs/IMMUTABLE_PUBLICATION.md)
records this bounded cut. Reusable telemetry packages and workbench artifacts now
share temporary-write/file-sync/no-replace-link/directory-sync mechanics. Family
validation, JSON encoding, byte limits, IDs, reload integrity, lineage and catalogs
remain in their stores. Other store writers remain queued within C1.

Actual checks:

- Package and deployment stored-byte/ID fixtures passed against the old publishers
  before extraction (`c1-before-extraction.log`, 0.017s/0.016s). The same fixtures
  pass afterward, including reopen, duplicate/no-overwrite, invalid references,
  truncated/foreign-field/digest-mismatched files and preservation of corrupt winners.
- `go test -race ./internal/immutablefile ./exploration ./workbench` passes. The
  helper race test requires exactly one complete winner among 24 distinct 128 KiB
  payloads and no temporary leaks; failure checks cover a file used as a directory
  and a failed link after temporary writing. Existing store concurrency/branch and
  corruption tests remain included. An initial sandbox run could not bind existing
  httptest loopback servers; rerunning with loopback access passed (exploration
  3.692s; helper/workbench also pass).
- `go test ./...` passes across all 13 root packages, including HTTP route tests.
  Hegel's 100 generated first-winner/duplicate sequences pass (0.048s), including
  empty initial payloads and 2–12 attempts. Its initial run used an unwritable home
  cache; the established writable XDG cache resolved that environment issue.
- Whitespace and changed local Markdown targets pass. No UI, request-owner or
  Quint transition protocol changed; no browser screenshots were regenerated for
  this filesystem-only cut. No crash/power-loss durability is claimed.

Cleanup removed 13 inspected dangling obsolete task UI images, then 12 explicitly
selected unshared COPY cache layers (479.3 MB reported; VFS disk recovery roughly
3.6 GB). The running image and a recent rollback image, all containers/volumes,
telemetry, test ledgers and repository data were preserved. Disk usage fell from
93% to 81%; free space was 5.8 GB after checks. No broad system/volume prune ran.
The existing deployed stack was not rebuilt for this cut. A post-cleanup audit
confirms 13 running services, restored traffic, all 14 original pending IDs
(15 total), exact retained manifests/review histories/pinned citations and source
hashes, reachable XTDB and zero OOM kills. No XTDB restart was performed.

Next within C1: freeze capture-specific identity/response/parent/large-number
fixtures before adopting matching publication mechanics; custody and native YAML
contracts remain separate until pinned.

## Shared question/native wire fixtures — bounded B2 completion (2026-10-22 UTC)

Final query/packages image:
`sha256:712f64c2c394716fa558c266096fd3425ecf40e54e16ed1180b1380a7c1c1254`.
[Shared-question protocol and acceptance audit](docs/QUESTION_PROTOCOL.md) records
this cut. Native evidence comparison now recognizes known Go zero/omitempty wire
defaults without rewriting stored definitions, digests or exact nanosecond bounds.
Seven valid fixture pairs falsely differed on the previous implementation; all
seven match with the fix. Changed selectors, units, namespaces, family and exact
windows remain distinct. Browser step guards also reject null/array native sources.

Actual checks:

- All 39 frontend test files pass through `npm run test:visualizations`. Sixteen
  shared protocol fixtures (nine accepted, seven refused) pass actual Go/Node link
  round trips, typed equality, canonical JSON and pinned native definition digests.
  The root fixture test passes (0.182s). JS syntax and whitespace checks pass.
- Expanded Hegel investigation-step properties generate millisecond and exact
  native scopes above JS safe integer range, preserving native bindings and
  distinguishing changed metadata. These pass with the question-link and
  question-session state/swarm campaigns (36.997s).
- Five real browser tests pass together on the final image: fresh comparison 57,
  mixed investigation steps 50, both linked-evidence 51 tests, and new native wire
  journey 114. Ordinary OTel gauge telemetry is captured with exact sub-millisecond
  bounds; omitted metadata drafts match canonical saved evidence without an
  implicit source read. Explicit document save and baseline preview preserve scope.
  Affected desktop/mobile screenshots were regenerated; frame 114a was inspected.
- Fresh comparison, mixed steps and linked evidence retain their actual observed
  Quint/quint-go ledgers and rejecting controls. The three enabled Bombadil
  campaigns pass: fresh comparison (446 records), mixed steps (157) and native
  protocol preview/close (425), with no violating records. The native actor reaches
  stage 2 with matching evidence, exact string bounds, zero POSTs and no overflow.
  These bounded checks do not claim exhaustive browser/Go validator parity.
- An initial run found one stale linked-evidence assertion expecting raw ISO clock
  text. It now verifies readable UTC plus the original precise datetime/title for
  both source-read and custody clocks. The full final run passes after this fix.
- Post-wrapper health verifies 13 running services, restored traffic, all 14 original
  pending IDs preserved (15 total), exact retained/live manifests and review
  histories, pinned citation/read basis, current catalog, XTDB health, final served
  source/server/binary hashes and zero OOM kills. No XTDB restart experiment was run.

Bounded B2 is delivered. Source-execution and retained-catalog adapters have distinct
lifetimes, so no generic request helper was forced. C1 store-specific identity and
publication fixtures are next; family validation/canonicalization stay authoritative.

## Service namespace wire fixtures and shared labels — B2 continuation (2026-10-22 UTC)

Final query/packages image:
`sha256:95239d0d41286e357c9a2064505adf0db7b32923855ad52ebdd1af1c159fc95c`.
Service catalogs/dashboard scopes and annotation target summaries now share the
namespace label formatter. Named values are quoted, distinguishing a literal
`missing` or `present empty` namespace from the corresponding presence state.
The lowercase sentence option preserves literal value case; wildcard descriptions
say “all namespaces”. Existing wildcard annotation rejection, qualified service
names, selection keys, query selectors/native bindings and stored digests retain
their existing contracts. [Namespace contract/diagram](docs/FIELD_PRESENTATION.md#service-namespace-identity--b2-continuation).

Actual checks:

- All 38 frontend test files pass. Eight shared Go/JS fixtures pin wildcard, missing,
  omitted/explicit empty and named selectors against six resource maps. Go checks
  logical matching, derived record predicates and canonical omitted-default JSON;
  JS checks tab-local selection keys, canonical aliases, predicate equivalence and
  common annotation/catalog labels without mutation. Root identity/exploration
  fixture tests pass (0.005s/0.021s). JS syntax, whitespace and changed local
  Markdown links pass.
- Hegel namespace predicate isolation (100 cases) and existing entry/live-filter
  state/swarm campaigns pass together (1.712s). No request-owner protocol changed.
- Three browser journeys pass across serialized runs on the final image: 98 scoped
  service/saved annotations, 113 source-free field/package staging and 60 namespace
  isolation. Expanded journey 60 ingests ordinary OTel spans and gauges for six
  namespace populations, including both literal sentinel names; it checks distinct
  catalog labels, scoped native values, record links and declared bindings.
  Its final complete run passes in 57s. Affected desktop/mobile screenshots were
  regenerated; the catalog frame was inspected.
- Final namespace Quint/quint-go observed ledger passes with four rejecting controls
  (8.934s). Expanded Bombadil actor passes with 350 records, including 347 catalog
  snapshots, with the distinct namespace label invariant valid throughout. Existing
  scope coherence and overflow invariants remain. Annotation journey 98 retains
  its actual observed trace/rejecting controls. No exhaustive label/UI model is claimed.
- Post-wrapper health verifies 13 running services, restored traffic, all 14 original
  pending IDs preserved (15 total), exact retained/live manifests, current reviews,
  pinned citation/read bound, XTDB health, served source/server/binary hashes and
  zero OOM kills. No XTDB restart experiment was performed.

Environment evidence: disk pressure was resolved by removing only verified obsolete
UI image references and ten explicitly selected, unshared task COPY-layer cache IDs.
This recovered approximately 3 GB; no volumes or telemetry records were removed.
The initial namespace run passed browser assertions/model validation but Bombadil
failed creating its browser tab (`Target.createTarget` timeout). The unchanged
integrated rerun passed, including the actor; no checks were weakened.

Local evidence: `artifacts/service-protocol-unit.log`, `service-protocol-go.log`,
`service-protocol-hegel.log`, `service-protocol-journeys.log` (98/113 pass and actor
startup failure), `service-protocol-final.log`, `service-protocol-health.json`,
and `artifacts/service-namespaces/1792693388017/`.

B2 remains active for broader native/envelope protocol fixtures and qualified
identity/presence adoption; no generic request owner was introduced.


## Record field presentation and Go/JS wire fixtures — first B2 slice (2026-10-22 UTC)

Final query/packages image:
`sha256:5a51acd51aed035ccaf5745fc05815e974a27b78bd33869e4de265dfc476e217`.
The pure field presenter now supplies live selector text, fixed filter chip parts
and record-package field handoffs. Fixed chips show Go's omitted empty string
value explicitly and quote non-empty literals, distinguishing `Empty string` from
`"Empty string"`. Missing/exists operators carry no value label. Qualified field
families, Unicode keys and exact decimal strings remain unchanged; presentation
never mutates predicates or question identities. Native binding semantics and
execution ownership remain independent. [Contract/diagram](docs/FIELD_PRESENTATION.md).

Actual checks:

- All 37 frontend test files pass. JS syntax, Python compile, whitespace and changed
  local Markdown links pass.
- Ten shared protocol fixtures pass root Go query validation, reference matcher and
  canonical omitted-default JSON checks (0.013s). JS consumes the same fixtures for
  display and existing question-identity equivalence, including resource/attribute
  separation, missing/present-empty, literal sentinel/Unicode text and uint64 max.
- Hegel live-filter and entry-intent state/swarm campaigns pass together (1.466s),
  each with 80 generated 24-step cases. The live-filter campaign now invokes display
  functions while checking unchanged draft/applied selectors and empty-default
  identity. No new request-owner abstraction was introduced.
- Four final real browser journeys pass together (1.4m): 113 source-free fixed
  staging/real package compilation/field handoff and mobile chips; 104 real OTel live
  filtering, fixed authoring and paused return; 110 Definitions draft/identity
  compatibility; 111 fixed reusable plan ownership. Both initial and final browser
  passes succeeded; the final pass includes the literal-sentinel distinction.
- Journey 104's observed Quint/quint-go ledger and rejecting controls pass (16.734s).
  Its updated Bombadil campaign passes with 371 records/stages 0–5; 368 snapshots
  carry the expected qualified live selector text. Existing no-overflow, bounded
  retention, explicit editor and paused-selection restoration invariants remain.
  This is not an exhaustive model of all field presentation.
- Affected screenshots were regenerated; new mobile frame 113a was inspected for
  clear missing/empty/literal separation and containment.
- Final post-wrapper audit verifies 13 running services, restored traffic, all 14
  pre-existing pending IDs preserved (15 total), exact retained/live manifests,
  current reviews and old pinned citation/read bound, XTDB health, served source/
  server/binary hashes and zero OOM kills. No XTDB restart experiment was performed.

Root disk space remains tight. Only identified obsolete task UI images were removed
for build/deployment headroom; no volumes or telemetry data were removed.
Local evidence: `artifacts/field-presentation-final.log`,
`field-presentation-final-unit.log`, `field-presentation-final-go.log`,
`field-presentation-final-hegel.log`, `field-presentation-health.json`, and
`artifacts/live-filters/filters-2074045a/`.

B2 remains active for broader protocol/qualified identity fixtures and adoption.
The inspected source exploration and retained catalog owners have different
abort/visit/snapshot/cursor lifetimes; a common request owner is not established.


## Retained catalog and owner consolidation — bounded B1 delivered (2026-10-22 UTC)

Final query/packages image:
`sha256:bde7234f2befb6397fe23ce61a994c5253a83b7f6f511a8a9b8a4280045a68f6`.
Retained metadata uses the shared passive compact row and inspector while retaining
its server-owned filters and 20-item cursor pages. Inspection performs no manifest,
review, telemetry or publication request. Tags are explicitly per-load current-head
snapshots; keys include load generation and manifest ID. Filter/refresh changes
retain selected metadata with an earlier-load notice; stale completions are rejected
and unavailable refreshes preserve existing rows. Exact pinned reviews remain
separate. Client/basis, catalog, citation and explicit recovery now have distinct
browser modules; existing transport, citation and recovery function bodies were
moved verbatim. [Owner diagram and contracts](docs/DEFINITION_LIBRARY.md).

Actual checks on final source/image:

- All 36 frontend test files pass. New transport tests protect exact request-body/key
  replay after lost acknowledgment, pending intent on optional-store failure and
  pinned citation read-bound precedence. JS syntax, Python compile, whitespace and
  changed local Markdown links pass.
- Hegel definition catalog properties, retained selection identity and pending-save
  state/swarm campaigns pass together (0.986s). Root `retention` tests pass, including
  branch/head/time/filter/cursor rules; no backend semantics changed.
- Seven browser journeys pass across serialized runs on this image: 98 annotations,
  99 saved retention, 100 exact live retention, 101 recovery/catalog/pinned citations,
  110 Definitions, 111 fixed reusable ownership and 112 document revision inspection.
  The final complete 101 run passes in 1.2m, including real response loss/reload/retry,
  source-free catalog inspection, a real delayed catalog reply overtaken by a filter
  change, injected store unavailability, mobile layout and an unchanged old pinned
  review after correction. Affected screenshots were regenerated and inspected.
- The final actual 101 ledger passes Quint/quint-go with eight rejecting controls
  (21.826s). The expanded Bombadil actor passes with 558 records, stages 0–5,
  at most 20 catalog rows and zero publication calls. Its deterministic sequence
  includes catalog browse/select/close and local recovery inspection. Existing
  observed annotation/native retention ledgers also pass; no exhaustive catalog DOM
  model is claimed.
- Final post-wrapper audit verifies 13 running services, restored traffic, all 14
  pre-existing pending IDs preserved (15 pending total), exact stored native/live
  manifests and review histories, pinned document citation/read bound/current tag
  catalog, optional XTDB health, served source/server/binary hashes and zero OOM
  kills. No new XTDB restart experiment was performed.

Environment/retry evidence: the root filesystem filled during deployment. Removing
only two identified obsolete task-built UI images recovered space; no volumes or
telemetry records were removed. An initial journey stopped at its unchanged 15s
reload limit (successful static requests took about 18s). A subsequent run completed
all browser assertions but its model subprocess hit the existing 60s harness limit;
that same ledger passed isolated replay in 19.530s. The final unchanged integrated
journey then passed, including the actor. Checks were not weakened.

Local evidence: `artifacts/retained-catalog-complete.log` (final 101),
`retained-catalog-final.log` (98 pass/model timeout), `retained-catalog-journeys.log`
(other five passes/reload timeout), `retained-catalog-final-unit.log`,
`retained-catalog-hegel.log`, `retained-catalog-go.log`,
`retained-catalog-ledger-check.log`, `retained-catalog-health.json`, and
`artifacts/evidence-library/library-891caf94/`.

Bounded B1 is delivered. B2 protocol fixtures and qualified-field helpers are next;
comparison/discovery and prerequisite-gated custody work retain their queue order.


## Investigation revision library — B1 continuation (2026-10-22 UTC)

Final query/packages image:
`sha256:7b3be499773ffc22114720dd62dacd53d38645b30e3e5404141941c111bb8924`.
Investigation revisions now share compact typed rows and a selected metadata
inspector. Opening the editor remains explicit and retains pending-note/dirty/busy
guards. Server cursors remain owned by documents; local row controls and search
state their current-server-page boundary. Exact IDs, preserved parent branches and
UTC custody timestamps remain visible. B1 remains active for retained catalogs and
touched retention owners. [Contract and diagram](docs/DEFINITION_LIBRARY.md).

Actual checks:

- Frontend test suite passes (35 file-level results); JavaScript syntax, whitespace
  and changed local Markdown links pass.
- Hegel supplied-catalog properties pass 200 cases including document, record and
  measurement typed identities, bounded paging, literal search and non-mutation
  (0.260s). This is not an exhaustive DOM model.
- Four final real-browser journeys pass together (1.4m): 36 deployment context and
  frozen evidence/narrative revisions; 110 Definitions draft/identity/stale-refresh
  compatibility; 111 fixed reusable plan ownership; 112 new document inspection.
  Journey 112 follows real server cursors, inspects a preserved child while an
  editor note is pending, checks guarded explicit Open, page-scoped filtering,
  failed-refresh retention, mobile containment and zero telemetry/publication reads.
- Updated context Bombadil actor passes with 421 records and all required stages;
  411 observed snapshots have catalog rows and 231 have a selected document.
  Browse/inspect/close actions preserve base revision, exact references and note
  guards; invariants check bounded rows and typed exact selection.
- Journey 36's actual app/evidence/narrative Quint/quint-go ledger passes with its
  four rejecting controls (12.47s). Existing model scope is unchanged.
- Affected journey screenshots were regenerated, including new mobile frame 112a.
  Final post-wrapper health verifies 13 running services, restored traffic, all 14
  preserved pending IDs, exact existing retained manifests/reviews/pinned document
  citation/current catalog, optional XTDB health, served source/server/binary hashes
  and zero cgroup OOM kills. No new XTDB restart experiment was performed.

Local evidence: `artifacts/document-library-final.log`,
`document-library-final-unit.log`, `document-library-hegel.log`,
`document-library-health.json`, and `artifacts/context-notes/1b1cb58e/`.


## Shared fixed reusable-question library — B1 continuation (2026-10-22 UTC)

Final image: `sha256:3864ab2b82b7d572b9a1bce50748059e114d046ffb434dca671a536f128bef23`.
The fixed reusable-question list now uses the same compact exact-revision rows,
loaded-catalog paging/search and selected inspector as Definitions. Selection is
source-free; Run remains a separate explicit fresh-window query. The application
invalidates pending compiler plans on selection/close, workspace/mode changes and
fixed-owner suspension. A completion must match the current generation, active
visit and selected revision before dispatching a source read. Abandoned work leaves
an explicit non-loading status. The presenter's optional selection callback carries
only user intent; refresh/disclosure lifecycle does not execute sources.

B1 stays active for compatible document/retained catalogs and touched retention
owners. This cut removes the older fixed-library card renderer, not typed execution
owners, saved source bounds or immutable identity rules. [Contract/diagram](docs/DEFINITION_LIBRARY.md).

Actual checks:

- All 33 frontend test files pass on final source; JavaScript syntax, whitespace
  and changed local Markdown links pass.
- Existing Hegel catalog paging (200 cases) and reusable revision properties
  (100 cases) pass together (0.389s). They protect supplied data/identity semantics,
  not a complete DOM lifecycle model.
- Final four browser journeys pass together (2.4m): 33 full-population measures,
  saved paging, real definition reuse, discovery and cancellation; 110 Definitions
  draft/identity/stale-refresh compatibility; 111 new reusable ownership checks;
  and 109 explicit/lazy fixed authoring and retained draft identity.
- Journey 111 holds real compiler responses across selection changes, detail close,
  workspace leave and fixed suspension into Activity. None triggers an obsolete
  query. A deliberately mismatched returned revision is refused; a later valid
  explicit Run issues exactly one query in the selected two-minute window. The
  explicit Activity visit is separately counted as one tail read. Mobile containment
  and non-loading restored status are checked; source data is not mocked.
- Core Quint/quint-go replay and rejecting controls pass (11.04s). Three enabled
  Bombadil campaigns pass: saved measures (326 steps), core workflows including
  selected-row reuse (297 steps), and shared Definitions library exploration. Both
  core campaign coverage files reach their required complete stage. These ledgers
  cover their existing accounting/semantic contracts, not every inspector state.
- Updated storyboards/screenshots document the shared fixed list and real core
  workflow. The mobile fixed-library frame is visually reviewed. Served hashes
  match source; all 13 services run after traffic restoration. Existing pending IDs,
  retained manifests/review history and pinned citations survive; OOM kills remain
  zero. No XTDB restart was tested this round.

Initial-image checks passed 110/111/109. The older core fixture first used a
non-unique title to select a revision, then attempted a service action inside a
collapsed support disclosure. It now selects its own definition and opens that
section explicitly. It also enters explicit fixed authoring and explicitly reruns
staged saved evidence, matching the previously delivered S2 contracts. Timeouts,
full-population/identity assertions and required Bombadil completion checks remain
intact. Final core and compatibility runs pass on the same final image.

Evidence: `artifacts/fixed-library-unit.log`, `fixed-library-final-unit.log`,
`fixed-library-hegel.log`, `fixed-library-journeys.log`,
`fixed-library-core-final.log` (older fixture diagnostic), `fixed-library-final.log`,
`fixed-library-health.json` and `artifacts/core-basics/basics-8675071a/`
(real ledger, replay, campaign traces and coverage).

## Compact record/native definition rows — first B1 slice (2026-10-22 UTC)

Final image: `sha256:0d60f896bc4f2a15b893b7f8cd56b368c4f71d6bc6a755dece2086ecb7684f08`.
`GooDefinitionLibrary` now composes compact typed exact-revision rows, local loaded-
catalog search/paging and one selected inspector in both Definitions catalogs.
Record annotation panels mount only in selected detail; the inspector's bounded
cache preserves visited local drafts. Selection does not validate/execute/publish or
replace the package editor draft. Same-title definitions remain separate. Explicit
Edit/Explore/Preview actions retain their existing typed owner contracts. Native
refreshes now reject older responses; failed refreshes retain old rows/detail.

This is a bounded first B1 cut, not completion of the whole milestone. The fixed
reusable-question list and compatible document/retained catalog adoption remain
open, as does subsequent touched retention-owner separation. No backend cursor,
family/lineage inference, global search or new recovery promise is introduced.
[Contract and ownership diagram](docs/DEFINITION_LIBRARY.md).

Actual checks:

- All 33 frontend test files pass on final source. Added pure checks cover typed
  exact IDs, same-title independence, local literal search/page boundaries and
  non-mutation. JavaScript/Python syntax, whitespace and changed Markdown links pass.
- Hegel: 200 generated catalog/page configurations pass (0.475s), including empty
  catalogs, typed keys, complete filtered paging, bounds and non-mutation. This is
  a supplied-data property campaign, not a complete DOM or lineage model.
- Initial image `5e487426`: journeys 98/40/110 pass (1.9m). Screenshot review found
  that the old whole-catalog height limit clipped selected detail. It is removed;
  only the shared row list now scrolls. An empty selected-status bar is hidden.
- Final image: all five journeys pass together (2.8m): 98 annotations/parallel heads/
  history, 40 native preview/service handoff/abandoned reads, 110 exact same-title
  record/native selection and draft retention, 24 real Python OTel package reuse,
  and 25 nested log-package import/refinement. No source manifest is required.
- Journey 110 additionally checks close-to-row keyboard focus, selected detail
  outside a filter, zero source/publication calls from selection, delayed real
  catalog-response rejection, forced failed-refresh retention and mobile inspector
  containment. Real Go APIs supply definitions; only the catalog failure boundary
  is injected. Existing source reads remain explicit in the reuse journeys.
- Final Bombadil campaigns pass: library inspection/filter/paging (264 steps) and
  annotation draft/history (85 steps), with zero violations. The existing real
  annotation Quint/quint-go replay and rejecting controls pass (15.32s); publication,
  parallel heads and pinned revision semantics retain their existing owners.
- Affected storyboards are regenerated; desktop and corrected mobile library
  frames are visually reviewed. Served assets/server/binary hashes match source;
  all 13 services run after traffic restoration. Old pending IDs, retained manifests,
  annotation history and pinned document citations survive; OOM kills remain zero.
  No XTDB restart test was performed in this round.

Evidence: `artifacts/definition-library-unit.log`, `definition-library-final-unit.log`,
`definition-library-hegel.log`, `definition-library-journeys.log`,
`definition-library-final.log`, `definition-library-health.json`,
`artifacts/definition-library/library-2237a710/` and
`artifacts/annotations/annotations-11da5486/` (real ledger/model/Bombadil evidence).

## Explicit/lazy fixed authoring — S2 mount gate (2026-10-22 UTC)

Image: `sha256:31c65cf004b22fac9a8672ebf30cbd346c20b016cc7331642e296140ddd24113`.
Fresh Activity and unrelated initial routes now register trail continuations without
constructing the fixed query editor. Explicit authoring, supported record links and
service/package/trace continuations mount it. Paused return suspends/detaches the
retained editor; reopening restores the same draft/DOM identity. First legacy
continuation events reach the new owner once. Native measurement links retain their
own owner; source reads, immutable evidence clocks and typed kernels are unchanged.
The explicit local live environment no longer depends on a hidden fixed input.

The [acceptance audit](docs/S2_A3_ACCEPTANCE.md), active queue, README, contracts and
architecture diagram now mark bounded S2/A3 delivered. **B1 is next**; D1 comparison
work and S3 capability-gated tail custody retain their existing positions.

Actual checks:

- All 33 frontend test files pass; changed JS syntax, local Markdown links and
  whitespace pass.
- Existing Hegel entry-intent and live-filter state/swarm campaigns pass (80 × 24
  each, 1.31s). These protect selector/source ownership, not every DOM mount detail.
- Six focused browser tests pass (2.2m): new 109 lazy mount/draft identity and
  source-free fixed staging, paused return 105, both question-link 49 cases, and
  Activity 102/103. Real OTLP/ClickHouse-backed surrounding journeys retain existing
  read counts, anchors/focus, cancellation, resume gaps and inert saved links.
- Compatibility journeys 31, 50, 104 and 106 pass on the same image, covering first
  service/package continuations, mixed record/native steps, typed live filtering,
  supplied-parent context and passive SSR/detached-property/reconnect behavior.
  Updated saved-operation journey 22 passes separately (12.4s; 15.6s runner): zero
  source reads on reopen, one editor and exactly two trace/log reads on expansion,
  original bounds/basis and unchanged saved receipt. Eleven distinct browser tests
  pass across these runs. Existing Quint/quint-go observed ledgers and rejecting
  controls pass for the exercised protocols; no universal DOM model is implied.
- Four enabled Bombadil campaigns pass: entry (631 steps, no eager editor), live
  filters (533 steps, zero/one/zero editor lifecycle), inventory (565 steps) and
  mixed investigation steps (166 steps). No violations.
- Affected real storyboards are regenerated, including new explicit-authoring
  frame 109a. Key desktop/mobile Activity, authoring and trace-context frames are
  visually reviewed. Source asset hashes, application images and all 13 services
  match after traffic restoration. Old pending IDs and retained evidence/reviews/
  pinned citations survive; cgroup OOM kills remain zero. No XTDB restart was tested.

The initial compatibility batch had four passes and one stale journey failure:
22 still expected the saved receipt/waterfall in the live Operations panel, despite
the earlier live/Saved ownership split. Its selectors now target the saved owner,
its context wording matches the existing separate-finding action, and stronger
read-count/completed-log-basis assertions protect the lazy continuation. The full
journey rerun passes with its original timeouts and receipt/bounds checks intact.

Evidence: `artifacts/lazy-authoring-unit.log`, `lazy-authoring-hegel.log`,
`lazy-authoring-focused.log`, `lazy-authoring-compatibility.log`,
`lazy-authoring-operation-final.log`, `lazy-authoring-health.json`, and the
corresponding real journey/model/Bombadil ledgers.

## S2/A3 acceptance audit and contextual lookback (2026-10-22 UTC)

Image: `sha256:67f1090a47789aeeda843e997e9530f3c6143302462caf870bc63bb79b22474b`.
The [acceptance audit](docs/S2_A3_ACCEPTANCE.md) maps active S2/A3 criteria to
implementation and existing validation. Bounded A3 service-record/receipt adoption
is delivered. S2's fixed-editor mount removal gate remains open: fresh Activity
still eagerly mounts the complete fixed editor beneath its disclosure. Explicit/lazy
fixed authoring with draft/link/paused-return parity is next, followed by B1.
Native comparison work stays D1. Historical validation entries are not rewritten
as if this audit reran all earlier campaigns.

Navigation now hides the new-read lookback on Saved and labels its purpose on
Services, Operations, Definitions HTTP preview and Examples. It preserves the input
value and fixed/saved bounds; navigation performs no source execution. No new
query/capture protocol or semantic evaluator was introduced.

Checks on this image:

- All 33 frontend test files pass; JavaScript syntax, whitespace and changed local
  Markdown links pass.
- Real telemetry journeys 107, 102 and 103 pass together (2.1m): source-free saved
  inspection/route changes, unchanged default, explicit trace/log context, delayed
  result rejection, one fresh monitor, scroll/background pause and explicit resume.
  Existing Quint/quint-go ledgers and rejecting controls pass, including the
  scroll/background ledger in 51.37s. No new semantic property campaign was added
  for this display-only navigation change; earlier Hegel evidence is cited explicitly
  in the audit rather than claimed as rerun.
- The existing Bombadil workspace actions execute 27 navigations in the 30-second
  campaign, with zero violations and the new Saved-lookback-hidden invariant.
- Final Saved journey rerun passes (28.0s) after correcting only the screenshot
  scroll position. The top-of-page frame is visually reviewed.
- Served source hashes match, all 13 services run after traffic restoration, old
  pending IDs and retained evidence/review citations survive; OOM kills remain zero.
  No XTDB restart test was performed this round.
- Affected Saved and Activity storyboards are regenerated; the added top-of-page
  Saved frame documents the missing misleading lookback. Presenter/flow ownership
  is shown in the audit diagram; the eager editor gate is explicitly unimplemented.

Evidence: `artifacts/scope-context-unit.log`, `scope-context-journeys.log`,
`scope-context-saved-final.log`, `scope-context-bombadil.log`,
`scope-context-bombadil/trace.jsonl`, `scope-context-health.json` and existing journey ledger directories.

## Supported-window native previews (2026-10-22 UTC)

Image: `sha256:43635ed83d33894a1b39957c8e63334ea76fbf484c4bd524a79b07144e3689b6`.
Supported-window previews now compose the shared native presenter for supplied
counter/histogram values, validated export coverage and diagnostics. Counter rates
retain exact rational formatting; histogram P95 remains a bucket range. Unsupported
coverage withholds values. Original/selected/excluded bounds use UTC timestamps
with exact nanosecond tooltips; excluded duration retains its exact nanoseconds in
a tooltip. Discovery/selection/inspection remain source-free; explicit recapture
still performs two independent reads and preserves partial-publication recovery.

Completed checks:

- All 33 frontend test files pass, including exact large rational rates, zero,
  unavailable coverage, histogram units and bucket ranges.
- Hegel native-presentation properties (200 cases, now including supplied-window
  values/non-mutation) and existing window-discovery support properties pass (0.27s).
- Real OTLP counter/histogram journey 55 passes (1.2m), with exact UTC tooltip,
  identity inspection, no-read selection, original-capture immutability, mobile
  containment, frozen sharing and forced second-capture failure recovery checks.
- The 30-second Bombadil support-discovery campaign passes with supplied-coverage
  timestamp invariants. Quint/quint-go replays the real observed ledger and rejects
  six invalid mutations (12.70s). This model checks read/capture boundaries, not every
  rendered UI state.
- All five affected storyboard frames are regenerated; desktop and narrow mobile
  previews reviewed. Docs and the presenter ownership diagram are synchronized.
- Served assets match source, both application services run the expected image,
  all 13 services run after traffic restoration, pre-existing pending IDs and
  retained annotations/document citations survive, and cgroup OOM kills remain zero.
  No optional XTDB restart test was performed this round.

Evidence: `artifacts/window-preview-unit.log`, `window-preview-hegel.log`,
`window-preview-journey.log`, `window-preview-health.json`, and
`artifacts/window-discovery/1792652961819/` (real ledger, Quint replay, Bombadil trace).

## Shared native values and coverage (2026-10-22 UTC)

Final query/packages image:
`sha256:5950e73230cc980cc896aa61fe08d3892e5043f91481dce851e76387daf4e056`.
Initial native checks ran on `86f7e492`; the final application change is the mobile
service-table layout. Shared native presenter/server assets are unchanged between
those images.

`GooNativePresentation` now shares latest native values, series diagnostics,
comparison-window coverage, diagnostic explanations and the existing exact-rate
formatter. Measurement details, service native summaries and fixed/live native
windows compose it. Latest intervals remain distinct from whole-window validity;
unknown coverage with no diagnostics never becomes validated coverage. Counter and
histogram evaluation remains in Go, including reset/gap/temporality semantics.

Service coverage is collapsed in its existing column; narrow service tables use
labelled rows with geometry checks so stream/value/expanded coverage remain together.
Comparison windows show UTC with exact nanoseconds in timestamp tooltips. Evidence
read/custody/evaluation clocks and controller ownership remain separate. Public
helper exports delegate for compatibility; shared styles have their own module.
[Contract and diagram](docs/NATIVE_PRESENTATION.md).

Completed checks:

- All 33 frontend test files pass on final source. JavaScript/Python syntax,
  whitespace and changed local Markdown file links pass.
- Hegel passes: 200 new supplied-series presentation cases, 100 native-window
  comparison cases, and counter/histogram state/swarm campaigns (100 × 30 each).
  Generated cases check exact nanosecond latest ordering, no older-value fallback,
  zero versus unavailable, count coherence and non-mutation (1.72s total).
- Initial real-source batch: six journeys pass (72, 74, 54, 59, 76, 55).
  Corrected timeline journey 77 and service-selection 76 pass on the final image,
  including their Bombadil and existing Quint/quint-go ledgers. Full service journey
  59 also passes on the final image (1.3m; 1.4m with wrapper). Seven distinct real-source
  journeys and all six enabled Bombadil campaigns pass across these rounds.
- Final served asset/server/binary hashes match source, including the new JS/CSS
  modules. All 13 services run after traffic restoration. Pre-existing pending queue
  IDs, retained manifests/reviews, pinned document citations and current catalogs
  survive; cgroup OOM counts remain zero. No package/XTDB restart test was performed
  this round.
- Service and comparison coverage storyboards are regenerated, including new
  desktop/mobile frames. Visual review found the old mobile horizontal table
  separated identity from coverage; the labelled-row layout now keeps them together.
  Geometry and source-read assertions cover opening coverage locally.

The timeline Bombadil actor initially closed the first inspector (stream identity)
instead of the selected native point; its selector now targets the selected-point
owner. A final service run completed its UI/Bombadil checks but hit the browser
helper's 60-second Go subprocess limit. Direct replay of that exact real ledger
passes in 25.36s, including 11 rejecting controls. The full service rerun then passes
with the original helper timeout and all assertions unchanged. Existing model ledgers
validate observed protocol/accounting boundaries, not complete native UI composition.

Evidence: `artifacts/native-presentation-final-unit.log`,
`native-presentation-hegel.log`, `native-presentation-journeys.log`,
`native-presentation-final.log`, `native-presentation-dashboard-replay.log`,
`native-presentation-dashboard-final.log`, `native-presentation-health.json`.

## Standalone tail and native identity inspection (2026-10-22 UTC)

Final query/packages image:
`sha256:0772f866195494ba396be40bbe69ae7e4fece7ace9f05bd55d7a446732fc59dc`.
Native checks first ran on `e6067667`; the final image changes only tail pointer-focus
handling and restores summary-before-actions ordering. Final tail checks run on
`0772f866`; served source hashes verify the common native assets are unchanged.

Standalone log/span tails pause before opening one cached record inspector, using
the common detail body and bounded supplied-parent context. Close/reopen retains
local search/body scroll; explicit resume/reset/scope replacement clears that owner.
Live Activity retains its callback inspector without a duplicate local panel.
Original observation windows and new-window continuity gaps remain distinct.

Native measurement stream lists, service summaries, supported-window previews,
native window comparisons and live counters share a lazy stream-identity inspector.
Existing native point inspectors mount the same typed identity directly. Name/unit,
resource versus attribute namespaces, scope/version and both schema URLs retain exact
values; no record timestamp, fleet aggregation or source read is invented. The old
attribute-disclosure export remains for compatibility, without application call sites.
Other aggregate/native result layouts remain separate work.
[Contract/ownership diagram](docs/RECORD_PRESENTATION.md#standalone-tails-and-native-identity).

Completed checks:

- All 33 frontend test files pass on final source, including native identity
  namespace separation, empty/exact large strings, absent schema and non-mutation.
  JavaScript syntax, changed local Markdown links and whitespace pass.
- Existing Hegel tail overlap/multiplicity state, native-window comparison and
  supported-window properties pass (100 cases each; 0.25s).
- Ten distinct real-source browser journeys pass across the rounds: 54, 55, 59,
  66, 72, 74, 77, 92, 102 and 103. These cover identity inspection without reads,
  native values/gaps/bounds, supported-window recapture/recovery, exact receipts,
  service summaries, tail retention, pause/resume/disconnect, retained search and
  close-to-opener focus. Native mobile and final mobile tail frames are regenerated
  and visually reviewed. Existing Quint/quint-go ledgers and their rejecting controls
  pass; they validate documented observed accounting/state, not trace completeness.
- Native-window, service-dashboard, supported-window and tail Bombadil campaigns
  pass after the verification corrections below. The service actor exercises the
  shared timeline and native identity; the native actor exercises identity before
  broader exploration. Tail observation invariants target the standalone owner.
- The final tail batch passes all four checks together (66, 92, 102, 103), with
  tail and entry Bombadil enabled. The pointer fix preserves keyboard pause and
  explicit resume while preventing a pause notice from moving the pressed button.
- Final served asset/server/binary hashes match source. All 13 services run after
  traffic restoration; pre-existing pending queue IDs, retained manifests/reviews,
  pinned document citations and current catalogs survive. Cgroup OOM counts remain
  zero. No package/XTDB restart was performed this round.

Initial campaigns caught stale assumptions: the tail actor mixed a hidden Live
Activity tail's state with standalone rows; the service actor required superseded
per-widget time bars rather than the shared timeline; the supported-window fixture
expected a v2 receipt. These now target the correct owner/current timeline and v3
receipt. The added mobile tail case then caught a real pause-status reflow between
pointer focus and click. Standalone tails now share Live Activity's pointer focus
guard; keyboard focus still pauses immediately. Corrected checks all pass.

Evidence: `artifacts/tail-native-final-unit.log`, `tail-native-hegel.log`,
`tail-native-journeys.log`, `tail-native-corrected.log`, `tail-native-final-tail.log`,
`tail-native-health.json` and associated journey/Bombadil ledgers.

## Single record inspector and contextual comparison controls (2026-10-22 UTC)

Query/packages image:
`sha256:0baa43164b7c08ade2be1efd05caa34f8f4a797f42f926da66a654544c2f041b`.

Fixed/frozen and linked/surrounding-log tables now use the shared cached inspector
as their single attribute entry point. Duplicate inline disclosures are removed;
closed detail slots occupy no space. Selection precedes inspection and existing
follow-ups. Typed search, exact Copy, raw evidence, owner-permitted refinement and
retention contracts remain. Legacy standalone noncompact tail/native-window
attribute disclosures remain compatibility consumers. Comparison buttons are hidden
without a single-field nonnumeric breakdown or retained baseline; retained baseline
Clear remains reachable while a question is draft. No source API/core semantics change.

Completed checks:

- Frontend suite: all 33 test files pass; JavaScript syntax and changed local
  Markdown links/whitespace pass.
- Hegel entry state/swarm (80 × 24) and capture state/swarm (60 × 15) pass,
  including local source fixtures (47.6s).
- Eight real-source journeys pass together (7.1m): 32, 35, 44, 45, 97, 106,
  107 and 29. These cover stored paging/revisions and actual package restart,
  error/latency comparison, numeric/grouped coverage/package reuse, fixed
  refinement, shared-field selection, Live/fixed presentation, saved context
  switching and keyboard/mobile typed inspection. Existing Quint/quint-go ledgers
  pass, including negative controls where defined.
- Updated capture and inspection Bombadil campaigns pass. Both now exercise the
  single inspector and require absent duplicate disclosures, at most one visible
  inspector and no visible empty detail slot per table. Actual screenshot frames
  are regenerated; saved desktop/mobile frames are visually reviewed.

- Four continuation runs pass together (4.2m): cross-app report/Bombadil with
  Quint enabled (26), generated retry/rejection/missing-terminal workload and
  observed Quint checkpoints (28), investigation follow-ups (42), and strengthened
  saved inspection/cache/focus assertions (107). Helpers now read each rendered
  record through the real inspector rather than the removed inline JSON. Fixed
  campaigns use the explicit Questions route; fresh Explore keeps its live intent.
- Final deployed asset/server/binary hashes match source; all 13 services run after
  traffic restoration. Pre-existing pending queue IDs, retained manifests/reviews,
  pinned document citations and current catalogs survive; cgroup OOM counts are zero.
  Packages restart was exercised by journey 32; XTDB was not restarted this round.

The runs use actual source telemetry, not substituted UI records. Existing Quint
protocols validate their documented observed accounting/state boundaries, not full
trace completeness. Other aggregate/native layouts and legacy standalone-tail
inspection remain separate work.

Evidence: `artifacts/inspector-unit.log`, `inspector-hegel.log`,
`inspector-focused.log`, `inspector-continuations.log`, `inspector-health.json`.

## Dedicated saved record owner and shared context (2026-10-22 UTC)

Query/packages image:
`sha256:2bbc7023bb0946ebea0213d4cb2e14a1af0214b1c3d73b33e50fb17168842f85`.

Saved record pages now mount `GooSavedRecords` with no query editor, measure
composer, question session or tail. Stored paging, distributions, grouped/numeric
values, record selection/retention and package creation retain their existing
contracts. `GooRecordResults` shares record/group tables and explicit trace/log
context with Live/fixed Explore. The old saved paint/state path and Live's unused
saved pager are removed; legacy readOnly callers delegate before editor setup.
Capture switch/leave/document transitions invalidate pending context publication.
Read bases remain independent; generation invalidation adds no upstream abort
promise. [Contract/diagrams](docs/SAVED_RECORD_PRESENTATION.md) and
[actual storyboards](docs/journeys/README.md).

Actual checks:

- Frontend suite: 33 test files pass after extraction and teardown correction.
- Hegel passes: supplied-parent properties (200), entry state/swarm (80 × 24),
  and capture state/swarm (60 × 15), including local source fixtures.
- Twelve real-source journeys pass together on the corrected image (8.3m):
  32, 85, 86, 34, 43, 100, 44, 105, 106, 107, 102 and 103. These cover exact
  stored paging/revisions and process restart, numeric/grouped/chart results,
  package reuse, retention, retained navigation and Live/fixed pause/return.
  Capture, entry and live-retention Bombadil campaigns and existing Quint ledgers
  pass. The capture actor now requires absent editor controls rather than hidden ones.
- Journey 107 passes (28.7s): absent query controls, local inspection, explicit
  wider trace/log reads, focus return, surrounding logs and rejection of a held real
  response after capture switching. Every source read/local transition passes the
  existing Quint/quint-go read/history protocol with four rejecting controls.
  This checks observed accounting/coherence, not trace completeness.
- Deployed asset/server/binary hashes match source; all 13 services run after
  traffic restoration. Pre-existing pending queue IDs, retained manifests, native
  review history and pinned citations survive; cgroup OOM kills remain zero.
  Packages restart is exercised by journey 32; XTDB was not restarted this round.
- JavaScript/Python syntax, changed Markdown local links and whitespace pass.
  Affected storyboards/screenshots are regenerated; new context/mobile frames
  are visually reviewed. The held response is real source data delayed at the
  Playwright transport boundary; UI data is not substituted.

Initial retained-view checks caught teardown resolving context IDs through the
whole document after detachment. The shared context owner now resolves its own DOM
subtree. A date-input fixture rejected redundant fractional zeros; its string is
canonicalized without changing the selected instant. A legacy fixed-query fixture
now enters the explicit question route.

Unused build cache and dangling images verified to descend from the task base
were reclaimed to resolve workspace disk pressure. Running images, tagged bases,
application volumes and queue data were preserved.

Evidence: `artifacts/saved-owner-final-unit.log`, `saved-owner-hegel.log`,
`saved-owner-focused.log`, `saved-owner-release.log`, `saved-owner-health.json`
and ledgers under `artifacts/saved-owner/`.

## Shared saved-record scope and deliberate continuation (2026-10-22 UTC)

Query/packages image:
`sha256:936c6bc57ff5a5e04cc5045495049643d66769853bc24b416260c69858bbc8a3`.

Capture record pages and metadata-only document references now share the pure
record receipt adapter and passive evidence summary. Full source count, fixed UTC
window, source-read time and stored custody remain distinct from retained/page
counts. Count/page disagreement is unknown with an explicit notice. Saved record
continuation stages the exact original question; Run explicitly rereads it. The
capture owner suppresses repeated embedded viewer scope/intro text. Its hidden
read-only editor remains a compatibility consumer for paging, measures, selection,
package creation and trace/log context; full removal remains planned.
[Contract/diagram](docs/SAVED_RECORD_PRESENTATION.md) and
[storyboards](docs/journeys/README.md).

Actual checks:

- Frontend suite: 33 test files pass, including full-page/metadata equivalence,
  large decimal counts, empty selectors, offset/retention versus total, unavailable
  clocks, disagreement states and non-mutation.
- Hegel passes: supplied-parent properties (200 cases), entry state/swarm
  (80 × 24 steps), and capture state/swarm (60 × 15 steps) covering immutable pages
  under arrivals, source resets and refreshes.
- Final-image journey 32 passes (1.4m), including capture Bombadil, four rejecting
  Quint/quint-go controls, actual packages restart, source-free disclosure and fixed
  staging before explicit Run. Full count/source/custody/page facts and mobile
  caption height are checked. Updated screenshots are visually reviewed.
- Journeys 41, 45 and 50 pass on image `421b7864abf4…`, before the final change to
  saved-view caption CSS/class only. They cover document reference summaries,
  fixed drafts, measurement continuation and source-free reopen. Investigation
  and step Bombadil campaigns and existing Quint/quint-go ledgers pass.
- Deployed assets/server/binary hashes match source; all 13 services run after
  traffic restoration. Pre-existing pending queue IDs, retained manifests, native
  review history and exact pinned citations survive; cgroup OOM kills remain zero.
  Packages restart is tested by journey 32; no XTDB restart was performed this round.
- JavaScript syntax, changed Markdown local links and whitespace checks pass.

Earlier runs exposed a Chromium `ERR_NETWORK_CHANGED` after the intentional
packages restart (one reload now handles only that transport error), an obsolete
copy assertion (replaced by frozen-mode/count/custody checks), and a new Bombadil
assertion that referenced an absent explanation count (corrected to use the
receipt count). Screenshot review found narrow mobile caption wrapping despite
an overflow check passing; final geometry validation covers the correction.

The Docker deployment exhausted workspace disk temporarily. Only verified-unused
prior task images and unused build cache were removed; volumes, retained evidence
and queue data were preserved.

Evidence: `artifacts/saved-records-unit.log`, `saved-records-hegel.log`,
`saved-records-journeys.log`, `saved-records-final-journeys.log`,
`saved-records-capture-final.log`, `saved-records-release.log` and
`saved-records-health.json`.

## Held service observations and shared record table (2026-10-22 UTC)

Query/packages image:
`sha256:60aa125903d7f5779e752545a145e22bfa2727a9dfcc27bf167cab7cc3c5c093`.

Live, fixed/frozen records and held service observations now share the passive
native-table/row builder. Owners retain stable keys, insertion, selection, actions,
viewports and all source lifecycle. Services lazily exposes the bounded span page
already read for field discovery, only after returned-question validation. Its
scope/basis stays separate from independently reread overview widgets. Opening,
inspecting and staging the exact fixed question do not read telemetry.
[Contract/diagram](docs/RECORD_PRESENTATION.md#held-service-observations) and
[actual storyboards](docs/journeys/README.md).

Actual checks:

- Frontend suite: 33 test files pass.
- Hegel: supplied-parent properties (200 cases), entry/filter state/swarm
  campaigns (80 × 24 steps each), and dashboard scope planning (100 cases) pass.
  The first invocation used a read-only default cache and failed to initialize
  libhegel; rerunning with the existing writable session cache passes.
- Initial real-source journeys 106 and 46 pass, including the updated service
  Bombadil actor and existing Quint/quint-go ledgers. Screenshot review then caught
  unreadably narrow mobile cells despite the overflow check passing. The final
  candidate uses the existing responsive record-row layout and a bounded Services
  viewport; journey 46 now also asserts mobile row, timestamp and caption heights.
- Ten real-source compatibility journeys pass together (8.1m): 85, 86, 100,
  105, 106, 46, 59, 60, 102 and 103. Service, dashboard, entry and live-retention
  Bombadil campaigns pass; existing Quint/quint-go trace replay and rejecting
  controls pass. This run used image `ad964ca493bc…`; the final image differs only
  by a Services-scoped mobile caption CSS rule and its added geometry assertion.
- Final-image journey 46 passes again (1.1m), including the updated Bombadil
  actor, exact-link staging without a read, retained selection on return, close
  focus, and mobile row/time/caption geometry. Final desktop/mobile frames are
  regenerated and visually reviewed.
- Served asset/server/binary hashes match source. All 13 services run after
  traffic restoration; pre-existing pending queue IDs, exact retained manifests,
  native history and pinned review citations survive. Cgroup OOM kills remain zero.
  No native-store restart was performed; prior restart evidence is carried forward.
- JavaScript syntax, changed Markdown local links and whitespace checks pass.

Other Services aggregate/native table schemas, inline attribute compatibility
consumers and full saved-editor removal remain separate work. No new source/API,
subscription, full-trace inference or exact tail-retention capability is added.

Evidence: `artifacts/service-records-unit.log`, `service-records-hegel.log`,
`service-records-focused.log`, `service-records-release.log`,
`service-records-final-focused.log` and `service-records-health.json`.

## Shared Live/fixed record presentation (2026-10-22 UTC)

Query/packages image:
`sha256:fcbb123a8793f112711c3e0399fc247b23cfbd3c4ee7d0492ed0f4a5e79ec807`.

Live and fixed/frozen record tables now share UTC timestamps, qualified service
names, severity/body and individual span status/duration. Their inspectors compose
the same passive detail body and typed attribute viewer. Reported span parents are
resolved only within supplied same-trace records, with explicit missing, ambiguous,
invalid, cycle and depth-limit states. Logs show correlation IDs without ancestry.
Inspection/disclosure never performs a wider read. [Contract and diagram](docs/RECORD_PRESENTATION.md)
and [actual storyboards](docs/journeys/README.md).

Actual checks:

- Frontend suite: 33 test files pass; syntax and whitespace checks pass.
- Hegel: 200 generated supplied-parent cases pass against the actual JavaScript,
  covering permutation/cross-trace independence, bounded unique paths, reported
  edges and non-mutation. Existing entry/filter state campaigns also pass
  (80 cases × 24 steps each).
- Real Chromium journey 106 passes (24.3s): generic OTel logs and spans have the
  same display projection in Live and fixed results. A foreign trace with the same
  span ID cannot supply a parent. Local inspection/context expansion and fixed
  staging make no request; Run remains an explicit read. Mobile and standalone
  SSR/detached-property/reconnect checks pass, including no component API calls.
  Its observed Quint/quint-go entry/fixed-read ledger and five rejecting controls pass.
- Nine compatibility journeys pass together (4.2m): 85, 86, 104, 100, 66,
  105, 93, 102 and 103. They cover fixed/frozen records, bounded trail navigation,
  typed filters, exact live-page retention through late arrivals/response loss,
  keyboard/scroll/background pauses, explicit resume gaps and retained paused return.
  The filter, entry and live-retention Bombadil campaigns pass. Affected actual
  screenshots are regenerated from this image.
- Served asset/server/binary hashes match source; all 13 services run after the
  traffic generator is restored. Pre-existing pending queue IDs, exact retained
  manifests, native review history and pinned citations survive. Cgroup OOM kills
  remain zero. No native-store restart was performed this round; earlier restart
  evidence is carried forward explicitly.

Wider Services adoption, a common table shell, removal of inline attribute
compatibility consumers and removal of the full saved editor remain open. The
shared display does not merge Live/fixed/frozen source ownership or evidence bases.
The journey uses real local OTel/source reads; component-gallery fixtures are
synthetic and do not stand in for the integration journey.

Evidence: `artifacts/record-presentation-unit.log`, `record-presentation-hegel.log`,
`record-presentation-focused.log`, `record-presentation-compatibility.log`,
`record-presentation-health.json` and ledgers under `artifacts/record-presentation/`.

## Return from fixed work to paused activity (2026-10-22 UTC)

Query/packages image:
`sha256:3c1e708b13a748d673388a897f89df8429470592ae30356b6a452b25acda2ea4`.

The existing header action now returns to a retained paused activity after fixed
or trace work. It preserves scope drafts, rows, selection, page/table/inspector
scroll and selected-row focus. Fixed work is suspended/detached before restoration;
return itself does not fetch or resume. The fixed action uses the shared inspector
action slot, outside its scrolling body. Reload drops this tab-local parent; shared
record presentation and ancestry remain open. [Contract/diagram](docs/LIVE_ACTIVITY.md)
and [storyboards](docs/journeys/README.md).

Actual checks:

- Frontend suite: 32 test files pass; JavaScript syntax/document links/whitespace pass.
- Actual JavaScript Hegel entry/filter state campaigns pass (80 cases × 24 steps
  each). Existing typed selector and fresh-entry ownership contracts are preserved.
- Journey 105 passes with a delayed real source response, pending-query cancellation,
  restored completed-query navigation, mobile layout, unchanged selected identity,
  draft predicates and keyboard focus. It proves preservation of table position,
  page scroll at 444px and nonzero inspector scroll at 50px in the final ledger.
- Its final observed Quint protocol counts both tail and fixed requests: four
  intended tail connections plus two explicit fixed reads; Return and releasing
  the cancelled response leave the combined count unchanged. Five rejecting
  controls still fail as intended. Final focused run: 23.2s.
- Compatibility journeys 85, 86, 104, 102 and 103 pass together (2.7m), including
  bounded question-trail retention/branching, generic service diagnosis, filtering,
  fixed/saved-link intent, explicit resume gaps, scroll/background pause and both
  Bombadil actors. The updated filter actor opens a fixed child and returns to
  the same paused selection. Screenshots are regenerated from the final image.
- Deployed source/binary hashes match (including the shared trail adapter); all
  13 services run; old pending queue IDs and exact native review citations survive;
  OOM kills zero. No native-store restart was performed this round.

Evidence: `artifacts/paused-return-unit.log`, `paused-return-hegel.log`,
`paused-return-compatibility.log`, `paused-return-final-focused.log`,
`paused-return-health.json` and ledgers under `artifacts/paused-return/`.
The held response is fetched from the real source, then deliberately delayed at
Playwright's transport boundary. The compatibility background event/HTTP 429 are
injected inputs; ordinary OTel/source reads use the real lab.

## Live attribute filters and fixed-window staging (2026-10-22 UTC)

Query/packages image:
`sha256:2cea13a255fbc260911b61963a105edcfc02077ee2fc89adfa94a5a221f9f503`.

The first S2 interaction cut reuses the shared attribute viewer for typed Include/
Exclude, removable draft chips and explicit Apply. Drafts pause without changing
held records; Resume refuses a changed scope. The inspector stages a fixed query
in its original read window without fetching; Run explicitly rereads it. Signal
changes clear signal-specific filters. Shared record presentation, ancestry and
retained-parent return remain planned. See [contract/diagram](docs/LIVE_ACTIVITY.md)
and [storyboards](docs/journeys/README.md).

Actual checks:

- Frontend suite: all 32 test files pass, including namespace/empty-value identity,
  draft/applied isolation, immutable snapshots and bounded predicate state.
- Hegel: actual JavaScript entry-gate and filter state campaigns pass, each with
  80 cases and 24 stateful steps (`artifacts/live-filters-hegel.log`).
- Real Chromium journey 104 passes (48.5s), including the dedicated Bombadil
  filter actor and observed Quint/quint-go replay with five rejecting controls.
  Its ledger records six intended tail connections, one explicit fixed query,
  zero unexpected publications/aggregate reads and zero page errors. Exclude
  returns east and empty while excluding absent region; fixed staging preserves
  the empty-string predicate. The filtered mobile first row is at 262px.
- Compatibility journeys 66, 93, 102 and 103 pass on this same image, including
  the existing entry Bombadil actor, keyboard/scroll/background pauses, explicit
  resume gaps, fixed/saved-link intent and desktop/mobile compactness.
- Deployed source/binary hashes match; all 13 services run; old pending queue IDs,
  pinned native review citations and retained evidence survive; OOM kills zero.
  No native-store restart was performed this round.

Evidence: `artifacts/live-filters-unit.log`, `live-filters-hegel.log`,
`live-filters-final-release.log` (four compatibility passes; the new actor's first
URL setup failed), `live-filters-focused.log` (corrected journey 104 passes),
`live-filters-health.json` and ledgers under `artifacts/live-filters/`.
The first candidate exposed a mobile focus-induced toolbar reflow that displaced
Inspect before pointer-up. The corrected compact adapter leaves pointer-button
pause to its click handler; keyboard focus still pauses. Frames are regenerated
from the final image. The compatibility background event and HTTP 429 remain
injected inputs; normal telemetry/source reads use the real local lab.

## Compact live-entry controls (2026-10-22 UTC)

Query/packages image:
`sha256:30b45206ce589c262b5ea202a730d67c902c29db6152f14095de27e439fbb71f`.

Environment, signal and Pause share one bar. Apply appears for scope drafts, which
retain the previous applied scope. The normal mode removes repeated scope/source
banners and explanatory headings; coverage/gap/error states remain visible. Mobile
navigation is collapsed behind Menu. See [contract](docs/LIVE_ACTIVITY.md) and
[updated frames](docs/journeys/README.md).

The real mobile first-record coordinate is **220.5px**, compared with approximately
750px in the previous frame. Journey 102 requires the desktop first row below
300px and mobile below 340px, checks that Menu performs no read, and checks draft
scope presentation. Journey 93 guards against the Menu leaking into fixed mode.

Final Chromium journeys 66, 93, 102 and 103 pass together, including the bounded
Bombadil actor and observed Quint/quint-go trace checks. Normal OTLP/source reads
use the real lab; the background event and capacity failure in 103 are injected.
Frontend tests: all 32 pass. JavaScript syntax and changed-document links pass.
Final source/binary hashes match the running image; 13 services running, prior
pending queue IDs and exact native review citations preserved, OOM kills zero.
Evidence: `artifacts/compact-entry-unit.log`, `compact-entry-health.json`,
`compact-entry-final-release.log` and observed entry ledgers. The initial candidate
exposed a mobile fixed-query layout regression from Menu CSS specificity; corrected
before the final release check. No native-store restart was performed this round.

## First Live Activity entry and whole-product consolidation plan (2026-10-22 UTC)

Final query/packages image:
`sha256:3399ec9c523d0b2fac85402da4af78a887542a44c12dc543aa3d89ae1120b134`.

Fresh entry now opens one bounded log/span monitor with explicit environment
scope, compact UTC rows, source-free paused inspection and explicit gap-bearing
resume. Fixed/saved links remain inert. See [contract and flow](docs/LIVE_ACTIVITY.md).
Design and architecture agents completed whole-product reviews; the planner merged
their removal/parity gates into S2/A3 and the existing queue. This ships S1 with
limits, not the planned shared record/inspector or facet interaction refactor.

Completed checks:

- Frontend suite: 32 test files pass; JavaScript/Python syntax checks pass.
- Hegel actual entry-gate state/swarm campaign passes (80 campaigns × 24 steps).
- Real Chromium journeys 101, 66, 93, 102 and 103 pass together: retained evidence,
  legacy tail, fixed recent activity, fresh stream entry and scroll/reconnect.
  Observed entry ledgers pass Quint/quint-go with five rejecting controls;
  journey 102 also runs the bounded Bombadil actor. Updated desktop/mobile frames
  are linked from the [storyboard index](docs/journeys/README.md).
- A final focused journey 103 rerun also passes empty/missing environment scope
  assertions, with no request on a missing scope. Evidence: `artifacts/stream-entry-final-scope.log`.
- Local SSE cancellation check: eight rapid aborted subscriptions return 200;
  four held streams are admitted and the fifth receives 429. Capacity unchanged.
- Deployed source/binary hashes match; all 13 services running; prior pending queue
  IDs and native review citations preserved; cgroup OOM/kill counters zero.
  Native restart verification is carried forward from the prior retention round;
  no new XTDB restart is claimed.

Evidence: `artifacts/stream-entry-release.log`, `stream-entry-unit.log`,
`stream-entry-hegel.log`, `stream-entry-proxy-cancel.json`,
`stream-entry-health.json` and the observed ledgers under `artifacts/stream-entry/`.
Background lifecycle input and HTTP 429 are deliberately injected in journey 103;
normal telemetry and source reads use the real running lab.

## Stream-first entry review (2026-10-22 UTC)

Product/design and architecture agents reviewed the current tail/source/UI code at
`3fbefea` against the revised stream-first direction. The planning agent integrated
these recommendations into the active S1–S3/A–F queue, preserving delivered work.
This round changes documentation only; no new runtime behavior, screenshots, load
results or browser integration campaign is claimed. Local document links and
whitespace checks pass.

## First product consolidation cut (2026-10-22 UTC)

Final query/packages image:
`sha256:c131da17eb70440bc2d703cbac98f311359750c157f2cba83b33386cd36bf2f3`.

- The planning agent integrated the product, architecture and design reviews into
  one active A–F queue. A2 live/saved operation ownership is delivered. The applied-
  result part of A1 is delivered; blank-entry action consolidation remains open.
- The frontend suite and JavaScript syntax checks pass. Four Hegel campaigns pass:
  operation heatmap selection, facet selection, population view state/swarm and
  brush/population scope. Go kernels and persisted artifact protocols did not change.
- Final-image journeys 101, 19 and 94 pass; journey 93 passes after making its
  rounded 1,250-pixel desktop target inclusive. The final focused 19 rerun also
  preserves an unsaved investigation title/note and live evidence/selection across
  capture publication and navigation. Full receipt reload blocks browser source
  APIs; a real late span still produces an explicit child without changing its parent.
- Journey 93 measures first record positions of 1,250px desktop and 1,775px mobile
  in its fixture (the review baseline measured approximately 1,585/3,053px).
  It checks no overflow, no duplicate facet overview sections, no reads on facet
  disclosures, explicit partial/no-data behavior and rejection of obsolete responses.
  Layout figures describe these viewports/content, not every dataset.
- Observed Quint ledgers pass for retained evidence, recent activity and range
  drilldowns. Bounded Bombadil campaigns pass for range selection, recent activity,
  and workspace live/saved ownership. An additional recent campaign explicitly
  closes/reopens facets and reaches its completion invariant.
- All 13 services run with traffic restored; served UI/server/binary hashes match
  the workspace and query/packages use the same final image. Previously recorded
  Redis pending IDs remain preserved (15 pending); OOM/OOM-kill counters remain zero.
  Existing pinned native review resolution still works after the packages restart.

Evidence: `artifacts/compact-unit.log`, `artifacts/compact-hegel.log`,
`artifacts/compact-browser-final.log` (three passed; initial strict desktop boundary
failed), `artifacts/compact-recent-final.log`, `artifacts/ownership-draft-final.log`,
`artifacts/ownership-bombadil.log`, `artifacts/compact-facets-bombadil.log`,
`artifacts/compact-layout-health.json` and the journey ledger directories.
Affected 19/93/94/101 storyboard frames are refreshed from the running app.

## Retained-evidence library and continuity (2026-10-22 UTC)

Final query/packages image: `sha256:5f0ea7394c034065f8c70fb297edabdacfe049b00e6d512d972e983b81a889a8`.

- Full root Go suite passes with actual XTDB integration enabled. Catalog checks cover exact filters, branch/withdrawal tags, half-open capture time, year-one bounds, pagination drift and immutable revision resolution. Retention/workbench race checks pass.
- All 115 frontend checks pass. Hegel state/swarm tests exercise the actual JavaScript pending-save journal across reload, retry, acknowledgment and changed-intent rejection.
- Journeys 99, 100 and 101 pass. The final-image 101 rerun also verifies expanded mobile catalog and pinned-review layout. Storyboard screenshots are refreshed. Lost committed responses recover the original retention/review IDs after reload, catalogs use current review tags, and documents preserve exact older reviews without source queries.
- The observed journey ledger passes Quint trace validation and eight rejecting controls. The new Bombadil campaign passes on the preceding feature image; the final image adds catalog validation boundary fixes and a bounded mobile date-control layout.
- After restarting Go packages and XTDB, the original pending request recovers the same commit, an unknown expired-source request returns 410, and the pinned review still resolves. All 13 services run with traffic restored, served source/binary hashes match, prior Redis pending IDs remain preserved (15 pending), and OOM-kill counters remain zero. Evidence: `artifacts/evidence-library-restart.json` and `artifacts/evidence-library-health.json`.
- The product usability study, reconciled implementation proposal, and architecture/design refactoring reviews are documentation/source reviews with explicit boundaries; they do not establish new runtime capabilities or approve their proposed milestones.

Evidence: `artifacts/evidence-library-go-all.log`, `artifacts/evidence-library-race.log`, `artifacts/evidence-library-unit.log`, `artifacts/evidence-library-hegel.log`, `artifacts/evidence-library-browser-final.log`, `artifacts/evidence-library-mobile-final.log`, and `artifacts/evidence-library/`.

## Exact live-page retention and retry visibility (2026-10-22 UTC)

Final query/packages image:
`sha256:82f34d996ca95120aebdb56439398286f48ad0ef23271694b8b74bb383e737d8`.
All 13 services run with traffic restored; served UI/server/binary hashes match the
workspace. Previously recorded Redis pending IDs remain preserved (16 pending at
the final check), and OOM/OOM-kill counters remain zero. Evidence:
`artifacts/live-retention-health.json`.

- Full root Go tests pass, including actual XTDB integration with concurrent
  correction retries, rollback, changed-intent rejection, fresh-connection await
  tokens and recovery without an ephemeral source page. Controlled-clock cache
  tests cover exact expiry, entry/byte bounds, immutable input/output and concurrent
  readers. The HTTP failure check proves unavailable recovery is not treated as an
  absent commit. Race checks pass for exploration, retention and workbench; native
  integration is separately opt-in and was enabled for the full root suite.
- All 30 frontend test files pass. Hegel annotation/selection properties and the
  new result-page state/swarm campaign pass: older reads remain unchanged across
  new pages, caller mutations and selections of equal-payload occurrences. Quint
  typechecking, Python compilation, wrapper syntax and changed-doc local links pass.
- Final-image journeys 100 and 97 pass with fresh storyboard frames. Journey 100
  seeds 102 real OTLP spans, selects a page observation, ingests an earlier event
  later, loses committed save/correction responses, recovers the same revisions,
  carries await tokens on reload, and rejects changed request intent. Browser
  telemetry APIs are blocked while reopening; ClickHouse itself remains running.
  Paging/refresh clear membership. Its observed ledger passes quint-go replay and
  nine negative controls, including changed source token/scope/member content,
  hidden reads/publications, broken history and duplicate publication on retry.
  Journey 97 also passes its existing observed-trace model and Bombadil campaign.
- Journeys 98/99 and both new journey-100 Bombadil campaigns pass on the preceding
  image with identical UI and successful-save paths. The final image additionally
  handles unavailable request recovery; journey 100 and the root HTTP check were
  rerun after that change. Bombadil exercises native review and live selection →
  unpublished draft → clear, observing that drafting publishes no native POST.
  The initial journey-100 fixture incorrectly assumed descending row order; it was
  corrected to insert an earlier event after the original read, without changing
  the product's query ordering or weakening the assertions.
- After the browser wrapper restored traffic, both Go and XTDB were restarted.
  The original page cache was gone, but the same committed request recovered the
  original annotation and exact canonical manifest. A new request against the
  unavailable token returned 410; native history stayed at two revisions. Saved
  capture evidence also survived the native restart. Evidence:
  `artifacts/live-retention-restart.json` and the health record above.

Live custody is bounded to two minutes, 64 pages / 64 MiB and 2 MiB per page;
selection publication remains explicit. Native initial/correction retries are
idempotent. Await tokens are read lower bounds on fresh connections, with tab-local
UI persistence; multi-node lag, cross-client basis sharing and exact historical
cutoffs remain unverified/unimplemented. File-backed annotation writes retain their
existing API. Cross-store catalogs, durable unacknowledged drafts, reviewed labels
and dataset export remain future work. See the [contract](docs/WORKBENCH_ANNOTATIONS.md).

Logs: `artifacts/live-retention-{go,unit,race,hegel,browser,browser-final,release,model-final,image-final}.log`.

## Shared annotations and optional XTDB retained evidence (2026-10-22 UTC)

Query/packages image: `sha256:22e51f29cc819afc8258905aa855b51da2aa60ed395cbbf221a86d09862c9744`.
The 12 core services and optional XTDB service run with traffic restored. Served
UI assets, server and Go binary hashes match the workspace. Previously recorded
Redis pending IDs remain preserved (15 pending after traffic resumed); OOM-kill
counters remain zero. Evidence: `artifacts/annotation-retention-health.json`.
Earlier iterative vfs builds exhausted local disk headroom and temporarily caused
collector insert failures. Additional legacy smoke attempts under background
traffic timed out; inspected fresh examples were correctly classified late while
waiting behind a slow consumer batch.
Unused build layers were reclaimed and fresh ingestion recovered. Evidence from
the affected interval remains unverified; no data volumes or stream entries were
reset to conceal the gap. The native database was restarted independently; the
retained manifest's exact serialized identity and two review revisions reopened unchanged afterward.

- The end-to-end sample smoke check passes with the owned traffic generator
  paused/drained/restored. `npm run test:smoke` now makes that controlled fixture
  reusable through the existing journey wrapper, preserving stale pending work
  and the original good/deadline assertions.
- The root Go suite passes, including real XTDB 2.2.0 integration tests using the
  pinned optional image. Cases cover system-time cutoff queries, independent review
  branches/withdrawals, exact record tokens and equal-payload occurrences, atomic
  rollback before commit, response retry, concurrent request recovery and native
  member projections. Race checks pass for the annotation/retention packages;
  native instance tests are separately opt-in. The original Go 1.24 module floor
  is retained; validation uses the installed cloud toolchain.
- All 30 frontend test files pass (111 checks in the final run). Python server
  compilation, Quint typechecking and new contract/research local links pass.
- Hegel annotation state/swarm and retained-selection properties pass alongside
  the existing saved-view campaign. They check immutable history, branch/head
  membership, selection-order independence, equal occurrences, exact large
  numeric/time tokens, detached input and rejected content mutation.
- Real OTLP journeys 98 and 99 pass on the final image with fresh documentation
  frames. Journey 98 crosses scoped services, saved views, narrative revisions and
  packages, then checks exact-tag lookup, concurrent correction and withdrawal.
  Journey 99 selects a saved observation, loses a successful native save response,
  recovers the same receipt, corrects its review and reopens with browser source
  API reads blocked. ClickHouse itself remains running. Both observed ledgers
  pass quint-go replay with negative controls for hidden reads/publication,
  altered scope/member content and broken parent history. Bombadil draft/cancel/
  history campaigns pass on both file and native annotation surfaces.
- Passive SSR/SPA component journeys 63–65 pass, including detached tag properties,
  invalid duplicate tags and no API calls on the standalone page. Namespace
  service journey 60 and record-population journey 97 pass; affected frames were
  regenerated. The final safeguard change reran journeys 98–99; it did not alter
  the service/record measurement paths.

The optional adapter retains selected **saved-capture** records; it does not
freeze earlier live result pages. Shared metadata stays file-backed. Cross-store
catalogs, correction retry keys, cross-node await-token propagation, reviewed
training labels and dataset exports remain future work. See the
[implemented contract](docs/WORKBENCH_ANNOTATIONS.md) and
[storage assessment](docs/ANNOTATION_STORAGE_SPIKE.md).


## Record-derived populations and product annotation design (2026-10-22 UTC)

Query/packages image: `sha256:742ec802dd2670b871f275bd5630c7bfde4a1e2738d9c38964b898dd70067938`.
All 12 services run with traffic restored. Served assets (including attributes),
server and binary hashes match the workspace. Prior Redis pending IDs remain
preserved (15 pending at the check after traffic resumed); OOM/OOM-kill counters
remain zero. Evidence: `artifacts/record-population-health.json`.

- Root Go tests and all 29 frontend test files passed. Generated Hegel shared-field
  cases passed against Go compilation/reference matching, alongside the existing
  immutable-view state/swarm campaign. Field inspection preserves namespaces,
  exact strings and empty values; it withholds absent/non-string/mismatched fields
  and values outside the typed predicate's UTF-8/byte constraints.
- All 13 browser journeys passed on this image: passive components (63–65,
  including both SSR/SPA 63 cases), legacy inspector/facets (29), and population,
  comparison, tail, recent-entry, retained-range and categorical workflows (90–97).
  Affected documentation frames were regenerated. A targeted 97 rerun then focused
  the parent facet view on Distribution for the documentation captures; baseline
  comparison remains explicitly exercised after applying the derived question.
- Real OTLP journey 97 passed inspector Include/Exclude without source reads,
  shared-field intersection of two observations from different services, exact
  session provenance, staged child/run/return, retained row checks/search/baseline,
  two-record population versus four-record baseline, mobile layout, a 20-of-22
  observation bound, reset on refresh and rejection of obsolete read/shared-set
  actions. Selected rows never provide aggregate counts or durable event identity.
- Quint/quint-go accepted the observed read/depth/provenance trace and rejected
  eight controls: hidden preview/return reads, changed companion environment,
  lost predicate, wrong return depth, changed provenance source, duplicate
  observations and out-of-window event time. The retained-population model is reused.
- Bombadil produced 822 trace records with zero violations; all five actions
  executed (recent/select/shared-field-and-open/return/clear), reaching the applied
  parent. Selection is bounded per materialized page; viewport, no-live-tail and
  exception/rejection invariants passed. Evidence:
  `artifacts/record-population-coverage.json` and the final actor trace.
- The separately requested design-agent pass is a whole-product source/design
  review, not implemented annotation functionality or customer research. It covers
  all current workspaces and developer/test/export surfaces, with typed targets,
  revision/origin contracts, diagrams, proposed cross-surface journeys and a first
  shared metadata slice. Shared annotation storage, task labels, exact fixed-record
  freezing and dataset exports remain proposed. No training data was published.

Evidence: `artifacts/record-population-{go,unit,hegel,browser,release,storyboards,image}.log`,
`artifacts/record-population/*/{ledger.json,TestQuintRecordPopulationLedger.log,bombadil/trace.jsonl}`.
See [implemented record workflow](docs/RECORD_POPULATIONS.md),
[whole-product annotation proposal](docs/PRODUCT_ANNOTATIONS_DESIGN.md), and its
[observation-specific design supplement](docs/TRAINING_ANNOTATIONS_DESIGN.md).


## Staged categorical populations (2026-10-21 UTC)

Query/packages final image: `sha256:dc5d7e229dddf7135f10767416279932be32df9878933317291830ef8c5daba7`.
The seven-journey regression batch passed on `ac12bc3f3922…`; the final tray
visibility/mobile-action polish passed targeted journey 96 on this image,
including a viewport assertion while browsing facets. The final documentation
capture checks scroll position explicitly so sticky elements render correctly.
All 12 services run with traffic restored. Served UI/server/binary hashes match
the workspace; prior Redis pending IDs are preserved (15 pending at the check
after traffic resumed), and OOM/OOM-kill counters are zero. Health evidence is in
`artifacts/facet-selection-health.json`.

- Root Go tests and all 29 frontend test files passed. Hegel checked 60 categorical
  intent cases against Go compilation and reference matching, including missing,
  empty and present-value exclusions; the existing immutable-view state/swarm
  campaign also passed. UI construction preserves the original nested predicate
  and independent companion, and rejects over-budget intent before Apply/Open.
- Real OTLP journey 96 passed source-free Include/Exclude, Undo/Clear, retained
  child staging/return, explicit two-record missing-or-empty reads, four-record
  pinned baseline comparison, configuration save/reopen, and three-record
  exclusion of west retaining missing/empty. The final build also passed pending
  save protection and cross-field selection across an explicit breakdown read.
  Desktop/mobile frames were regenerated from this running build.
- Quint/quint-go accepted the observed categorical depth/read ledger and rejected
  five controls: hidden preview read, hidden return read, changed companion
  environment, lost selection predicate and incorrect return depth. It reuses
  the retained-population model.
- The final Bombadil campaign produced 627 records with zero violations. All five
  custom actions executed and completed at the applied parent, depth zero, with
  no viewport overflow or live tail. Exception/rejection invariants passed.
- Regression journeys 90–95 passed alongside 96 in the seven-test release batch,
  covering saved populations, comparisons, live tails, recent entry and retained
  range drilldowns. Their affected documentation frames were regenerated.
- Arbitrary complements, selection-versus-rest and automated difference scoring
  remain proposed. The staged tray/Undo history are temporary intent; saved views
  store typed questions and optional baseline configuration, not frozen evidence.

Evidence: `artifacts/facet-selection-{go,unit,hegel,recheck,release,final,storyboards,image,final-image}.log`,
`artifacts/facet-selection/*/{ledger.json,TestQuintFacetLedger.log,bombadil/trace.jsonl}`,
and `artifacts/facet-selection-coverage.json`. The first draft browser run was
interrupted to correct its navigation: pinning selects Compare, so the journey
must explicitly return to Distribution before choosing facet values.
See the [contract, process diagram and storyboards](docs/FACET_SELECTION.md).


## Time brushing and retained population drilldowns (2026-10-21 UTC)

Deployment: query/packages use image
`sha256:94e51837384f221a1cf0db8d532bb8af386946cff4de78ae615456a18eb6a15c`.
All 12 services run with traffic restored; served assets, server and binary hashes
match the workspace. Prior Redis pending IDs remain preserved (14 pending at
check), and OOM/OOM-kill counters remain zero. Evidence:
`artifacts/range-drilldown-health.json` and `artifacts/range-drilldown-coverage.json`.

- Root Go tests and all 29 frontend test files passed. Hegel generated 80
  JavaScript brush-to-range-query projections checked by Go compilation, preserving
  bounds, source identity, missing/empty predicates and unmodified parent inputs.
  The reused immutable-view state/swarm campaign passed (60 × 15 steps). The new
  Quint depth/read model typechecks.
- Real OTLP journey 94 passed reverse brushing, Escape cancellation, source-free
  child staging, independent selected reads, nested returns, retained baseline,
  actual parent row/inspector nodes, tail observation keys and paused state,
  focus/scroll restoration, late child response rejection, exact keyboard handles
  and native touch selection. Journey 95 passed server HTML, native point/keyboard
  compatibility, detached property updates and large exact values without API calls.
- Quint/quint-go replay accepted the observed depth/read ledger and rejected seven
  controls: hidden preview/return reads, widened bounds, altered baseline, lost
  tail rows, extra subscription and wrong return depth. The first replay exposed
  an empty-parameter generated import; the model now supplies its explicit initial
  read-counter constant, and replay/rejecting controls pass.
- Regression journeys 77, 78, 88, 89, 90, 91, 92 and 93 passed alongside 94/95
  in the ten-journey batch. These cover native metrics and frozen reopening,
  SSR/SPA timelines and UTC controls, linked services, saved populations,
  baseline comparisons, live populations and recent entry. Their affected
  screenshots were regenerated. Later staged-child Run/Back presentation changes
  were followed by targeted 94/95 reruns, with actual frame regeneration.
  The final CSS-polished image passed 94/95 again, including one visible nested
  Back action, a visible primary Run and no misleading child advanced-mode exit.
- The final Bombadil trace contains 415 records with zero violations. All three
  actions (`goo.rangeRecent`, `goo.rangeStage`, `goo.rangeReturn`) executed and
  reached completion at parent depth zero with no live tail or viewport overflow.
- Selection primitives remain passive; query scope, subscriptions, child lifetime
  and capture state remain application-owned. Child staging/return never uses
  tail rows as population inputs or automatically starts live updates. Retention
  is bounded session state, not immutable captured evidence.

Evidence: `artifacts/range-drilldown-{go,unit,hegel,quint-typecheck,browser,model-recheck,recheck,release,final,navigation}.log`,
`artifacts/range-drilldown/*/{ledger.json,TestQuintRangeDrilldownLedger.log}`,
the optional Bombadil traces under the same directories, and the
[contract, sequence diagram and actual storyboards](docs/RANGE_DRILLDOWNS.md).

## Direct recent-activity entry (2026-10-21 UTC)

Deployment: query/packages use image
`sha256:ae4a96113d5c19b256c846a888013c6fa4d48d2c0e0a79da9889851f5557f55f`.
All 12 services run with traffic restored, served asset/server/binary hashes match
the workspace, prior Redis pending entries remain preserved (14 pending at check),
and OOM/OOM-kill counters are zero. Evidence:
`artifacts/recent-entry-health.json` and `artifacts/recent-entry-coverage.json`.

- Root Go tests and 29 frontend test files passed. The fresh-window helper retains
  complete selectors, signal/source identity and detached configuration, resets
  inspection offset, and rejects invalid clock/lookback inputs. Existing Hegel
  portable-selector and saved-view state/swarm campaigns passed for the unchanged
  Go planner and immutable storage.
- Real OTLP journey 93 passed against the packaged image: one explicit plan and
  two typed exploration requests with shared fresh bounds, local signal switching,
  service-facet refinement, source-free range preview/column edits/save/reopen,
  fresh-run selector preservation including an independent companion, log-only
  and no-data environments, partial-source failures and obsolete response rejection.
  The fixture independently checked ingestion through ClickHouse. No tail or other
  telemetry endpoint started implicitly. Desktop/mobile frames were regenerated.
- Quint/quint-go accepted the observed initiation ledger and rejected five
  mutations for hidden preview/reopen reads, changed environment, erased filtering
  and an implicit subscription. Bombadil executed `goo.recentActivity` and
  `goo.recentLogs`: 450 trace records, zero violations, both actions observed and
  completion reached with no live tail or viewport overflow. OOM counters remain
  zero.
- Regression journeys 66, 69, 90, 91 and 92 passed against the same final image:
  standalone tail, live comparison/freeze, saved populations, attribute/baseline
  comparisons and live population settings. Their affected desktop/mobile
  documentation frames were regenerated. Combined with journey 93, six browser
  journeys passed; `artifacts/recent-entry-regression.log` records the five
  existing workflows.
- Initial browser attempts identified a quoted-value locator and retained
  same-document button text; test assumptions were corrected. Screenshot review
  corrected a narrow facet toolbar and moved advanced controls below disclosure.
  The final successful journey uses the packaged source, not a hand-built mockup.

Evidence: `artifacts/recent-entry-{go,unit,hegel,browser,recheck,release}.log`,
`artifacts/recent-entry/recent-f301a19b/{ledger.json,TestQuintRecentEntryLedger.log}`,
`artifacts/recent-entry/recent-f301a19b/bombadil/trace.jsonl`,
`artifacts/recent-entry-coverage.json` and the
[contract, diagram and screenshots](docs/RECENT_ACTIVITY.md).

This implements the direct-entry/layout slice. General brushing, retained child
scope navigation and composable multi-item capture remain subsequent design work.

## Recent-activity product design (2026-10-21 UTC)

Documentation-only [whole-product design pass](docs/LIVE_EXPLORATION_DESIGN.md):
reviewed current adapters/component contracts and existing journey 92 desktop
evidence. Includes proposed layout, gestures, composition boundaries, sequence
diagram and acceptance storyboards. No runtime behavior changed; no new browser,
model or property campaign was run. Proposed behavior is explicitly labelled.

Population-live deployment (2026-10-21 UTC): query/packages use image
`sha256:86105ae4b88e41060744a07d971331f1e648bfdd221af0befbe80a60903d6467`. All 12 services run; traffic is restored,
served UI/server/binary hashes match the workspace, existing Redis pending entries
remain preserved (14 pending at check), and OOM/OOM-kill counters remain zero.
Evidence: `artifacts/population-live-health.json` and
`artifacts/population-live-coverage.json`.

## Explicit population live telemetry (2026-10-21 UTC)

- Root Go suite and focused vet passed. Tail selectors preserve logs as the
  omitted-signal default and also compile span queries; unsupported signals and
  incompatible saved settings are rejected. Optional saved tail settings retain
  old-view encoding and immutable prior configurations.
- The frontend suite passed 29 files. Pure cases cover bounded span/log retention,
  source-free configuration and original-window trace inspection. Hegel generated
  tail/selector properties and 60 × 15 saved-view state/swarm campaigns passed,
  including live-setting mutations and immutable histories.
- Real OTLP journey 92 passed explicit span/log starts, fixed baseline capture,
  full rolling count/latency/status comparison, 123 matching spans versus a
  100-row display buffer, explicit axis rescaling, focused/paused trace inspection,
  retained original read windows after later polls advance, source-free save/open,
  reconnect gaps, signal switching, navigation pause and baseline-selector
  mismatch rejection before a current telemetry read. Mobile record text remains
  in a readable content column with exact UTC/ns details.
- Standalone log-tail 66, live comparison/freeze 69 and population journeys 90/91
  passed across the recorded regressions. Final source and screenshots match the
  deployed image. Earlier assertions were corrected to allow retained hidden DOM
  on same-document reopening; screenshot review led to explicit mobile message
  width checks and column placement fixes.
- Observed Quint/quint-go read-initiation ledgers passed and rejected hidden
  reopening reads, unsupported signals, unbounded retention, erased gaps and
  changed baseline references. Bombadil explicitly started and paused the tail
  without runtime, retention or overflow violations. The initiation ledger does
  not count internal Go polling reads; existing tail semantics remain shared.
- Display primitives remain passive. The population adapter owns subscriptions,
  selector compatibility, metadata preparation and cancellation. Tail rows never
  define population totals; span records do not establish complete traces.

Evidence: `artifacts/population-live-{go,vet,unit,hegel,browser,recheck,final,complete,verified,release}.log`,
`artifacts/population-live/*/{ledger.json,TestQuintPopulationLiveLedger.log}`,
`artifacts/population-live/*/bombadil/trace.jsonl` and the
[live storyboard/diagram/screenshots](docs/POPULATION_WORKSPACE.md#watching-a-population).

Population-breakdown deployment (2026-10-21 UTC): query/packages use image
`sha256:99b583a7cfcdafa726d7f48827d547a93684163c8558025a45ab88c5136eaa04`. All 12 services run; traffic is restored,
served UI/server/binary hashes match the workspace, prior Redis pending entries
remain preserved (14 pending at check), and OOM/OOM-kill counters are zero.
Evidence: `artifacts/population-breakdown-health.json` and
`artifacts/population-breakdown-coverage.json`.

## Population attribute breakdowns and saved comparisons (2026-10-21 UTC)

- Root Go suite and focused vet passed. Saved-view checks cover addressed ranking/
  baseline revisions, unchanged prior views, detached caller configuration and
  rejection of foreign environments/incompatible grouping. Existing version-1
  artifacts retain their encoding when optional breakdown configuration is absent.
- The frontend suite passed 29 files. Pure checks retain nested selectors,
  independent companions, missing versus empty values and canonical optional query
  fields. Hegel passed generated portable selectors and 60 × 15 state/swarm
  sequences expanded with breakdown/baseline edits and immutable history checks.
- Real OTLP journey 91 passes full-population ranking, value-to-value comparisons,
  explicit filter reset, source-free save/open, baseline replay/refresh, missing/
  empty filtering, previous-window no-data and mobile composition. Failed baseline
  reads retain successful signal results and configured baseline queries for retry.
  Keyboard navigation visits only visible tabs; late previous-window responses
  cannot mutate replacement controls. Population journey 90 and retained-navigation
  journeys 85/86 also passed across the recorded runs.
- Observed Quint/quint-go read ledgers passed and rejected hidden save/open reads,
  foreign baselines, missing replay reads and invalid ranking. Bombadil explicitly
  ran a saved view, changed ranking locally and compared a previous window without
  runtime/overflow violations. Earlier runs caught canonical false-field mismatch
  (fixed) and an order-sensitive tie assertion (corrected to test both values).
- Source queries/group aggregates and comparison semantics reuse the existing Go
  implementation. Shared distribution/facet/comparison adapters use workspace and
  signal namespaces; view artifacts contain configuration, never result evidence.
  Numeric cohort comparisons remain unsupported.

Evidence: `artifacts/population-breakdown-{go,vet,unit,pure,hegel,release,final,verified,retry,complete}.log`,
`artifacts/population-breakdown/*/{ledger.json,TestQuintPopulationBreakdownLedger.log}`,
`artifacts/population-breakdown/*/bombadil/trace.jsonl`,
`artifacts/population-breakdown-coverage.json` and
[storyboards/diagrams/screenshots](docs/POPULATION_WORKSPACE.md#attribute-breakdowns-and-comparisons).

Population-workspace deployment (2026-10-21 UTC): query/packages use image
`sha256:5be65187e1c5f7ddd03c4b0bce63c3fa9ab3f1d68b54baa2ac9818fefae78904`. All 12 services run, traffic is restored, and served UI assets, Python server
and Go binary match the workspace. Existing Redis pending entries remain preserved
(14 pending at check); cgroup OOM/OOM-kill counters remain zero. Evidence:
`artifacts/population-final-health.json`.

Current additions: see the [cloud baseline](ENVIRONMENT_VALIDATION.md) for the
successful default Compose run, real cross-app Bombadil journeys and generated
Quint MBT success/fault campaigns in this environment. The historical hosted
environment limitations below describe the original handoff session.

## General population workspace and immutable saved views (2026-10-21 UTC)

- Root Go suite and focused vet passed. Planner checks retain portable predicates,
  withhold incompatible companions without dropping clauses, and reject foreign
  companion scopes. Store/HTTP checks cover immutable content identity, changed
  presentation revisions, mixed scopes, required columns, strict decoding and
  tampered artifacts. The actual Python proxy accepted a 17,482-byte two-query
  planning request, beyond its previous single-query budget, without evaluation.
- The frontend suite passed 29 files. Hegel passed 100 generated portable-selector
  plans and 60 × 15 saved-view state/swarm sequences covering layout, columns,
  range changes, saves and immutable prior configurations.
- Real OTLP journey 90 passed cross-service span/log filtering with excluded
  resource/attribute groups, missing/empty selector preservation, two-signal range
  preview/apply, explicit independent companion selectors, presentation-only
  changes, save/reopen/run, normal builder restoration, source-free navigation,
  mobile layout, failed-log partial results, cancellation/late-response rejection
  and source-free fresh-window staging. Configuration reopening and saving do not
  read telemetry. Existing question-workbench, retained-navigation, trace-finding
  and linked-service-timeline journeys also passed across the recorded runs.
- The observed Quint/quint-go ledger passed and rejected five mutations: foreign
  environment, lost saved predicates, hidden preview/reopen reads and invalid
  columns. Bombadil opened a saved view, explicitly ran its two populations and
  changed layout without property violations. Artifacts retain earlier test
  corrections for canonical empty-value JSON and checkbox replacement; column
  edits now preserve controls and record inspectors.
- Shared timeline labels/notes are supplied by the owning adapter. Entering
  population mode stops an old single-question tail/context; mode-specific source
  requests retain independent cancellation and publication generations. Saved
  configuration uses existing typed queries, not another measurement evaluator.

Evidence: `artifacts/population-{go,vet,unit,hegel,browser,final,recheck,release,verified}.log`,
`artifacts/population-workspace/*/{ledger.json,TestQuintPopulationWorkspaceLedger.log}`,
`artifacts/population-workspace/*/bombadil/trace.jsonl`,
`artifacts/population-proxy-budget.json` and the refreshed
[storyboard and diagrams](docs/POPULATION_WORKSPACE.md).

## Aligned service timelines and shared UTC selection (2026-10-21 UTC)

- The frontend suite passed 28 test files. Pure checks preserve exact counts and
  formatted nanosecond duration values, missing latency, independent input data,
  service/namespace/native selection identity and half-open bounded selections.
- Hegel passed 100 generated range-preservation cases against the actual adapter,
  plus the existing 60-case immutable-history state/swarm check (15 steps each).
- Real OTLP journey 88 passed Apply/reset across four widget reads, source-free
  range preview, identical time domains across five lanes, unchanged record
  predicates and native definitions, filtered span/log counts, partial log failure
  with usable span lanes, and mobile overflow checks. Journey 89 passed native
  keyboard controls, server markup, detached properties and zero API requests.
  Existing service-dashboard and metric-service journeys also passed on this image.
- The observed Quint/quint-go ledger passed and rejected five mutated histories:
  environment drift, changed bounds, a preview read, reset-read drift and lost
  selectors. Bombadil executed all four discovery/range/Apply actions and reached
  a loaded aligned overview with no property violations. Its initial setup read
  the fixture environment after route navigation; the corrected actor retains
  that value before navigating and passes the rerun.
- Duplicate per-widget activity charts are omitted from this composition; fixed
  summaries, exact questions and independently read evidence remain available.
  Native lanes do not resample or merge streams. This milestone scopes the service
  overview, not arbitrary cross-service populations or saved dashboard layouts.

Evidence: `artifacts/linked-timeline-{unit,hegel,browser,final,recheck}.log`,
`artifacts/linked-timeline/*/{ledger.json,TestQuintLinkedTimelineLedger.log}`,
`artifacts/linked-timeline-bombadil-coverage.json` and
[storyboard/diagrams/screenshots](docs/LINKED_TIMELINE.md).


Linked-timeline deployment (2026-10-21 UTC): query/packages use image
`sha256:f39b44ec66cd9cdefb199f923d4bcf9c0c3e503a8dff107b457620ea220de506`. All 12 services run, traffic is restored, and served component/app assets,
server and Go binary match the workspace. Existing pending Redis entries remain
preserved (14 pending at check); cgroup OOM/OOM-kill counters remain zero. Evidence:
`artifacts/linked-timeline-final-health.json`.

## Population → trace → finding (2026-10-21 UTC)

- Root Go suite and focused vet passed. Exact uint64 page selection checks cover
  invalid identity/time/duration, duplicate traces, ambiguous field separators,
  order invariance, changed reads and frozen-page membership. Legacy captures do
  not synthesize new metadata. All 27 frontend test files passed.
- Hegel passed 100 generated page-example cases plus 60 immutable linked-history
  state/swarm cases (15 steps each).
- Real OTLP browser journeys covered compact service navigation, retained history,
  captures, inventory, service dashboards and metric discovery handoffs. An initial
  native-instrument assertion needed to open its new disclosure; the test and
  Bombadil actor were corrected to follow the visible UI. The extended inventory
  actor also exposed a stale first-button selector and pre-namespace question
  construction; it now chooses a named service and uses the shared transform.
  Trace-finding actor artifacts now use a unique per-ledger directory so reruns
  preserve previous evidence without output-path collisions.
- Journey 87 exercised failed second-capture recovery, two independent successful
  captures, no implicit document write, explicit revision saving, rejection of
  staging into a concurrently edited draft, stored reopening without source reads
  and mobile overflow. Its Quint/quint-go observed ledger passed and rejected five
  scope/reference/read mutations. Bombadil executed run, example navigation and
  finding composition with no property violations. Corrected inventory and
  metric-service actors also completed their live journeys; all ten distinct
  targeted browser journeys passed across the recorded runs.

Evidence: `artifacts/trace-journey-{go,unit,vet,browser,regression,final,final-recheck,inventory-recheck}.log`,
`artifacts/trace-examples-hegel.log`, `artifacts/trace-finding/*/ledger.json`,
`artifacts/trace-finding/*/TestQuintTraceFindingLedger.log`,
`artifacts/trace-finding/*/bombadil/trace.jsonl` and the refreshed
[storyboard frames](docs/TRACE_FINDINGS.md).


Final trace-finding deployment (2026-10-21 UTC): query/packages run image
`sha256:0c146e9dcea3b238605d807f21a72357797f00de6e7df2160c20c3625826e14b`. All 12 services run; served UI assets, Go binary and Python server match the
verified workspace. Prior pending Redis entries remain preserved (14 pending at
check); traffic is restored and cgroup OOM/OOM-kill counters remain zero. Evidence:
`artifacts/trace-finding-final-health.json`.

## Live span latency and recorded error share (2026-10-20 UTC)

Saved span questions requesting duration/status now compare count, compact
mean/p50/p95/p99 and recorded Error/total share under one rolling window and
current evidence basis. The existing count aggregate already returns these
measures in one SQL statement. Current refresh responses expose current/baseline
basis metadata without changing historical comparison receipt formats. Freeze
preserves the complete original result without another source read.

- Root Go suite/vet passed; exploration/HTTP checks passed under the race
  detector. A new source check retains full measures over 1,200 records while
  keeping only 1,000 inspection rows, verifies one combined aggregate per read,
  and freezes/reopens summaries after the source is closed.
- The frontend suite passed 66 tests. Shared latency elements accept optional
  server `ceiling-ns` or detached `.ceilingNs` properties. Exact geometry retains
  nanosecond integers; invalid/undersized domains withhold markers. Existing
  automatic scales remain the default for other callers.
- Hegel passed 100 generated large-population comparisons, immutable summaries,
  Unknown status preservation and descriptive-change reversal checks.
- The main campaign passed five browser checks (3.5 minutes): both chart-frame
  contracts, existing record/native fresh comparisons, count-only live comparison
  and real OTel live span regression. Both live actors executed required
  preview/choose/start/freeze/link actions. The span actor additionally checks
  per-measure common domains, exact percentile positions and snapshot basis.
- The span journey observed baseline 1 ms/1 recorded error out of 2, current
  10 ms/1 out of 4, then 100 ms p95/p99 and 4 errors out of 7. Count and latency
  required explicit rescale; recorded error share retained its 0–100% axis.
  Failed refresh preserved all displayed summaries, basis and bounds. After an
  eighth 1,000 ms error arrived, freeze still retained the displayed 7 spans,
  340 ms duration sum and 4 errors. Quint accepted the observed ledger and
  rejected all ten corruptions. The count-only ledger retained six controls.
- A separate real-source check passed empty-current-window capture/freeze:
  no percentile or error-share geometry, no zero-latency/zero-error claim, and
  unavailable comparison state retained after freeze. The supplied-domain HTML
  and pre-upgrade SPA checks also passed. Eight distinct browser tests passed
  across the completed campaigns; source-backed 71 and affected 57/63/69 frames
  were regenerated and desktop/mobile 71 frames inspected.
- The first supplemental wrapper attempt stopped before Playwright: Docker
  briefly reported the owned traffic generator as a zombie. It subsequently
  exited and was restarted; the retry drained without deleting data and restored
  traffic. One added standalone assertion incorrectly assumed the unit fixture's
  p50 rather than the gallery's 40 ms p50; the corrected assertion passed. These
  unsuccessful attempts are not counted as passing checks.

Final runtime check matched image and served source/binary bytes, all 12 services
running, health OK and historical pending IDs retained. The traffic process was
also checked alive (not zombie). Evidence: `artifacts/live-span-journeys.log`,
`live-span-component-empty-retry.log`, `live-span-components-final.log`,
`live-span-hegel.log`, `live-span-final-health.json`, and
`artifacts/live-span-measures/0cb74c185d/` including its actual ledger/actor trace.

Count-only captures retain their definition; native metric comparisons remain
fixed-window. Measure refresh is bounded polling, not SSE. Status coverage and
sampling/completeness remain explicit unknowns. See [the contract](docs/LIVE_COMPARISON.md),
[the component domain API](docs/UI_COMPONENTS.md#supplied-percentile-domain), and
[journey 71](docs/journeys/README.md#71-a-service-regression-across-count-latency-and-recorded-error-share).

## Rolling record baseline and passive chart frame (2026-10-20 UTC)

Saved span/log questions now support explicit rolling full-population count
refresh beside frozen linked evidence. The app owns cancellation, generations,
source-window labels, stale retention and explicit axis rescaling. Freeze
publishes the displayed server snapshot without another telemetry read; linking
it stages an investigation draft and saving creates an immutable child revision.
The parallel component slice supplies `goo-chart-frame` with SSR attributes,
detached SPA metadata, native child/focus retention and no source lifecycle.

- Root Go suite and vet passed. Source/capture and HTTP checks passed under the
  race detector, including source-free exact freeze, repeated-freeze idempotence,
  strict request validation, 16-draft admission, budget replacement and obsolete
  capability rejection. Python proxy syntax and Compose configuration passed.
- The frontend suite passed 64 tests, including unavailable/incompatible frozen
  display state and unstarted-panel pause semantics. Hegel passed 100 generated
  shifted equal-duration windows and incompatible environment cases. Its first
  run generated durations beyond the existing two-hour record-query limit;
  the corrected generator matches that supported domain rather than widening it.
- Four browser checks passed together: both passive frame contracts, the existing
  fixed-window fresh-comparison record/native journey, and the real OTel rolling
  comparison journey. Frame tests use explicit synthetic standalone HTML; the
  product journey reads actual ClickHouse ingestion through standard OTLP.
- The rolling journey observed 2 baseline/4 current records, grew current to 7,
  required explicit rescale, injected a 503, retained the stale result/window,
  then ingested an eighth source record before freezing exactly the displayed 7.
  It reopened the stored result, verified the baseline unchanged, and rejected a
  completed response held across preview disposal. Quint accepted the observed ledger and rejected all six
  corrupted observations.
- Initial source failure is labelled unavailable with no invented retained result.
  Mobile frame bounds are asserted independently of page overflow: screenshot
  inspection exposed nested paired-bar clipping, fixed in the shared component
  stylesheet. Baseline/source provenance is disclosed without repeating it inline.
- A subsequent full rolling run passed explicit investigation linking and saved
  revision preservation. Its 30-second Bombadil actor executed preview, baseline
  choice, start, freeze and link; the test requires all five observed actions.
  The final state mapping also has focused regression coverage: unavailable
  frozen evidence remains unavailable, and pausing an empty panel stays unknown.

Final release: all four selected browser tests passed together in 1.4 minutes
(`artifacts/live-measures-release.log`), including the required Bombadil actions
and live Quint replay with six rejecting mutations. Runtime validation matched
the final image and served source bytes, all 12 services running, health OK and
all five historical pending stream IDs preserved. The wrapper restored traffic.
`artifacts/live-measures-final-health.json` records the final checks.

Live native metrics, duration/status charts and SSE measure refresh remain
planned. Record refresh uses bounded polling and a two-minute process-local
snapshot capability until explicitly frozen. Serialized draft budgets do not
establish a production authorization model or telemetry completeness.
See [the contract and sequence diagram](docs/LIVE_COMPARISON.md),
[journeys 69–70](docs/journeys/README.md#69-a-frozen-baseline-beside-a-rolling-record-population)
and [the passive component contract](docs/UI_CHART_FRAME.md).

## Live logs, passive waterfalls and count-time plots (2026-10-20 UTC)

Explore now tails an applied log selector over a separate rolling event-time
window. The Go source/compiler adapter emits bounded SSE batches through the
Python streaming proxy; the app owns generations, reconnect gaps, buffer
retention, follow-scroll and inspection pause. Counts label local observations,
not the full question population. Two parallel agents extracted passive
`goo-trace-waterfall` and `goo-time-bars`, adopted by existing trace inspection,
Explore and Services, and prepared a chart/detail reuse review.

- Root Go tests and vet passed; the frontend suite passed 60 tests. The tail
  HTTP admission/cancellation checks passed under the race detector.
- Hegel passed 100 generated overlapping-multiset/late-arrival campaigns. An
  initial property assumption incorrectly classified a first observation as late;
  the corrected property compares against a previously observed maximum.
- The release campaign passed nine tests: core basics, discovery, workbench,
  Services, UI inspection and both standalone waterfall/time-bar contracts.
  Core/Services Bombadil campaigns require actual chart interaction and check
  exact count/state geometry and focus preservation. The inspection campaign
  completed deterministic waterfall selection/detail checks before exploration.
- Tail browser assertions passed duplicates, same reported IDs, late data,
  foreign-scope exclusion, 100-row eviction, focus pause, resumed gaps, injected
  transport failure recovery, mobile containment and scope clearing. Its first
  integrated model check found a Quint built-in-name collision; after renaming
  that action, the observed ledger and all six rejecting mutations passed.
- Earlier live checks exposed newest-first batch eviction discarding the newest
  observations; the source now presents each bounded batch in event-time order,
  with a regression check. A chart fixture retained application/json from /health
  and did not upgrade custom elements; its explicit synthetic HTML fixture passed.
  An interrupted image deployment was not counted as validation; subsequent
  deployment checks matched running image IDs and served asset bytes before tests.

- The final tail rerun passed end to end, including the 30-second Bombadil
  start/pause/resume/focus/close sequence with required completed-stage evidence,
  Quint live replay and all six negative controls. Across the corrected runs all
  ten selected browser tests passed. New standalone and real app frames were
  regenerated and inspected; the chart/detail review is linked from the library.
- Runtime verification passed: all 12 services running, health OK, served tail,
  waterfall, time-bar and application hashes plus Go binary match the workspace
  and build context. Existing five stale Redis pending IDs were preserved; one
  current active entry brought the final pending count to six. Traffic restored.
- Durable ingestion cursors/replay, trace tails, arbitrary chart algebra and
  baseline-versus-live measures remain planned. Count plots retain explicit
  zero/unknown distinctions; native interval/histogram semantics remain separate.

## Shared facets and typed attributes (2026-10-20 UTC)

Explore live/frozen distributions now compose a caller-owned `goo-facet-list`.
Local search hides rows without changing supplied exact counts, scales, ranking
or action state. No local matches is distinct from no captured source records.
Attribute inspectors now reuse `goo-field-value` and a passive
`goo-attribute-list`, keeping qualified mapping, search, clipboard feedback and
query actions in the existing app adapter. Values preserve explicit presence,
null/empty/false, exact integer strings, safe text, lazy nested batches and a raw
value escape. Standalone SSR and detached SPA use are documented.

- Frontend suite passed all 54 tests. New pure checks cover exact values and
  absence/type distinctions, JSON rejection, cyclic/unsupported input and
  accessor/inherited serialization hooks without executing hooks.
- The initial live campaign passed eight browser tests: component conformance,
  log/trace discovery, grouped measures, question workbench and UI inspection.
  Discovery/group/workbench observed Quint ledgers and rejecting controls passed;
  30-second grouped and attribute-inspection Bombadil campaigns passed with new
  facet visibility/presence invariants.
- Standalone checks cover SSR/pre-upgrade properties, detached values,
  reconnects, safe text and native facet row/state retention without API reads.
  Live inspection verifies scoped duplicate keys, literal strings, clipboard,
  no-match status, original-window narrowing and trace/log/mobile adoption.
- Source reads, query ASTs and Go semantics were unchanged. A parallel agent
  extracted and reviewed typed fields; review caught an inherited array
  serialization hook and a page/component CSS precedence issue, both corrected.
- Final post-review campaign passed all seven browser tests, including new
  invalid-property and 50-to-75-item lazy-expansion assertions, corrected
  trace-facet layout, updated grouped/inspection Bombadil campaigns and Quint
  discovery/group trace checks. Real storyboards and synthetic component frames
  were regenerated; standalone and live attribute screenshots were inspected.
- Runtime checks passed: all 12 Compose services running; health OK; served
  component/attribute/application hashes and binary match workspace/build context.
  All five pre-existing Redis pending entries remain and traffic was restored.
- Field snapshot updates reset disclosures. Facet row patches require explicit
  `refresh()`; there is no mutation observer or live transport. Broader
  screen-reader validation remains outstanding as documented for the library.

## Passive tables and explicit result states (2026-10-20 UTC)

`goo-table` and `goo-result-state` now provide native caller-owned table viewports
and supplied state displays in Explore and Services. Captions, scoped headers,
keyboard overflow and compact density share one implementation. Source reads,
full totals, paging and action callbacks remain in the app. Native row patches,
attribute changes and reconnects retain existing row/action identity; the component
has no subscription or query behavior. Draft/cancel paths clear loading state.
The design subagent implemented Services adoption and reviewed the shared contract.

- Frontend suite passed all 52 tests. The first live campaign passed all seven
  browser tests (component contracts, grouped measures, workbench, reported edges
  and Services), including observed Quint trace validation/negative controls and
  the 30-second Services Bombadil campaign. New Bombadil invariants check table
  captions/scoped headers and explicit polite result regions.
- Browser contracts check server HTML, pre-upgrade state assignment, caller-owned
  row/action identity through append/reconnect, retained focus, supplied state
  transitions, no component API reads and mobile keyboard horizontal scrolling.
  Services checks local-search misses without new reads and independent partial
  source availability. Workbench checks loading-to-draft state reset.
- Parallel review caught stale loading styling on cancelled/edited questions,
  a lost widget viewport bound and redundant outer scroll containers; corrected.
  Screenshot inspection also corrected mid-word table headings. Shared styles
  preserve readable headers with contained horizontal overflow.
- Final post-review campaign passed all six tests (components, grouped measures,
  workbench and Services), including the updated state/table assertions, Quint
  checks and a second Services Bombadil campaign. Affected real journey frames
  and synthetic reference frames were regenerated and inspected.
- Runtime verification passed: all 12 services running, health OK, query/packages
  image matches, served component/application asset hashes and Go binary match
  the workspace. All five pre-existing Redis pending entries retained; demo
  traffic restored. Build cache/dangling images were pruned without touching volumes.
- No Go semantic core changes were made. Full screen-reader validation of the
  existing mobile table-to-card pattern remains outstanding; the browser checks
  establish markup, keyboard behavior and containment. SSE remains planned.

## Framework-neutral display components and compact exploration (2026-10-20 UTC)

The public library now registers `goo-latency`, `goo-stat-list` and
`goo-metric-comparison`, using ordinary server HTML attributes and detached SPA
properties. Its standalone `/ui-components` reference has synthetic fixtures and
loads no app runtime/query client. Services and live/frozen Explore reuse the
latency/stat primitives; stored comparisons group mean/P95/P99 on one matrix.
Only supplied percentiles are shown; no P90 evaluator was introduced. Query
construction now occupies one compact workbench, and applied measures do not
force advanced settings open. Widget secondary actions use native disclosure.

The [whole-app design audit](docs/UI_COMPONENT_AUDIT.md) prioritizes passive tables,
result states, facets and typed attributes next. It also specifies eventual SSE
and baseline-vs-live boundaries. These additional elements and streaming features
remain proposed. Current summary updates replace inner DOM/reset disclosures;
they are not high-rate focus/scroll-preserving tailers.

- Frontend suite passed 52 tests, including exact large-value formatting,
  zero/absence, missing quantiles and input immutability.
- `TestHegelDashboardPlanning` passed 100 generated cases for the unchanged typed
  query/adaptor contract.
- Standalone browser tests passed server HTML upgrade, pre-definition property
  assignment/reactivity, detached values, malformed JSON, safe text, keyboard
  disclosure and mobile containment with no API requests.
- The corrected live campaign passed all 11 browser tests (component conformance,
  error/latency and stored comparisons, grouped measures, native windows, question
  workbench, full breakdowns, exact edges and service dashboards). Their observed
  Quint checks passed. Thirty-second Bombadil campaigns passed for comparisons,
  exact edges and service dashboards, including the new quantile display invariant.
- The standalone fixture also checks configurable heading levels, reconnects and
  detached current-side comparison updates that leave supplied baseline values
  unchanged. This validates display updates, not SSE delivery or a live evaluator.
- Direct screenshot inspection exposed an ancestor CSS minimum width leaking
  from an outer breakdown table into a nested component. App selectors are now
  scoped to their owning table; a composition check requires the Change column
  to fit on desktop. Grid placement belongs to the app, not the reusable element;
  a bounded heading-level setting preserves the embedding outline.
- After component chrome isolation, the six focused browser checks passed again,
  including standalone/reconnect updates, nested comparison geometry, mobile
  layouts, source-free edge evidence and the service-dashboard Bombadil actor.
  Affected live frames and clearly labeled synthetic component frames were
  regenerated and inspected. All 12 Compose services are running; health and
  served component/app asset hashes match the built workspace. Existing Redis
  pending IDs and volumes were retained and traffic restored. Unused build cache
  and dangling images were pruned after a full-disk build condition.
- The first live campaign passed eight tests and exposed two older fixture
  assumptions. A demo gauge legitimately became unknown while traffic was paused;
  its test now injects an isolated ordinary OTel gauge with a fixed fresh scope.
  A metric-handoff assertion now expects the explicit missing-namespace selector
  introduced by the earlier identity work. Neither fix relaxes source semantics.

## Reported-edge investigation and visual comparisons (2026-10-20 UTC)

Distributed relationships now stage typed exact child-population questions, keeping
parent ownership in the original environment/window/namespace universe. Filters,
packages, captures and investigation steps retain that selector; explicitly wider
child-service and whole-trace actions remain labeled. Real fixtures include 130
children beyond the 100-record page, duplicate/ambiguous parents and namespace-
scoped ownership. See [reported edges](docs/REPORTED_EDGES.md).

Comparison views now use paired zero-based bars, common duration ceilings,
0–100% error/share axes, equal-width breakdown tracks, explicit unavailable
hatching and aligned native histogram bins. Detailed scope/provenance and exact
tables remain in disclosures. See [visual comparisons](docs/VISUAL_COMPARISONS.md).

- Root Go tests and vet passed; the frontend suite passed 51 tests.
- Hegel edge planning passed 100 generated cases, including missing/empty/literal
  namespaces, parameterized service names and detached question custody.
- Live journeys 35, 52 (record and native cases), 54, 58, 59 and 62 passed,
  including real OTLP ingestion, saved evidence, source-free reopening, mobile
  containment and independent source failures. Their Quint observed-trace checks
  passed; edge replay rejected five deliberately corrupted ledgers.
- Thirty-second Bombadil campaigns passed for latency/error comparisons and
  reported edges. Comparison invariants check common scale metadata and bounded
  geometry; the breakdown journey additionally measures equal track pixel widths.
- A final native-layout refinement promotes the actual rate/distribution ahead of
  coverage clocks and retains metric-record counts in the exact-values disclosure.
  Its focused journey passed, checking visual order as well as reset/gap/layout
  withholding; the refreshed native frames were inspected directly.
- Final runtime checks: all 12 Compose services running, health OK, served UI
  hashes and Go binary match the built workspace. Existing Redis pending IDs were
  preserved; traffic was restored after each journey wrapper.
- Initial journey runs exposed fixture/test issues: the UI's two-minute scope
  legitimately included an extra parent, the trace action was attempted while
  Distribution was selected, and native-window tests still expected receipt v2
  after the earlier v3 rollout. The journeys now select the intended one-minute
  window, open Records before expanding a trace, and check the current version.
  No comparison receipt algorithm or historical receipt was rewritten here.

## Declared services alongside observed telemetry (2026-10-20 UTC)

Services now combines independent fixed-window discovery and current operation-
package registry metadata. Declared-only rows retain explicit not-observed states,
namespace uncertainty and all declared bindings without creating counts, event
timestamps or parent edges. Registry failure remains distinct from no declarations;
late registry results are discarded after scope edits. Current declarations do not
establish deployment or emission during a historical telemetry window.

- Frontend suite: 50 tests passed.
- `TestHegelDeclaredCatalogIsolation`: 100 generated cases passed, checking
  namespace presence/value isolation, declaration order, retained bindings and
  unchanged source evidence.
- Journey 61 passed its initial live historical-inventory → service → empty capture
  → source-free recipient reopening path, registry failure and stale-result checks.
  Its Quint replay and 30-second Bombadil campaign passed. The final focused run
  additionally checks explicit namespace retention and unavailable-registry wording.
- That added check exposed a real namespace-inventory persistence defect (HTTP
  502, `invalid saved inventory`): the receipt validator compared selector pointers
  after serialization. It now compares semantic values. Missing/empty/named
  namespace save/reopen/refresh and mismatched-selector regressions pass in
  `go test ./...`; `go vet ./...` also passes. Existing receipt formats are unchanged.
- The first post-fix run passed the live namespace capture and all five Quint
  negative controls, then exposed a Bombadil startup race: repeated start actions
  canceled their own in-flight discovery. The actor now starts once after the
  capture is enabled and waits for discovery; its liveness deadline is unchanged.
- Initial browser campaign: journeys 46, 47, 59, 60 and 61 all passed (4.5 minutes).
  Final focused journey 61 passed (52.1 seconds), including the namespace capture,
  `no_data` custody, five rejecting Quint mutations and the corrected 30-second
  Bombadil campaign. Ledger/actor output: `artifacts/declared-services/1792462359930/`.
- Refreshed affected service screenshots and added storyboard 61 with desktop/mobile
  frames and a registry → discovery → question → capture diagram.
- Final health/source checks: all 12 services running; query/packages image
  `sha256:5c88917f0a65ffd2ae2cab406ffdb88d275d37a1e4f74667de7f9bdad7b62251`; served assets and binary
  match the tested source/build. All five pre-existing stale pending IDs remain;
  traffic is restored (6 total pending entries at the health read).
  Evidence: `artifacts/declared-final-health.json`. Unused build cache/images were
  removed; application volumes and stream entries were preserved.

## Namespace-aware service identity and declared operation questions (2026-10-19, America/Los_Angeles)

[Service identity](docs/SERVICE_IDENTITY.md) now separates missing, present-empty
and literal OTel namespaces across discovery, reported parent ownership, fixed
record questions, native bindings and operation-package readers. Omitted
selectors retain legacy question/definition serialization and SQL. Registered
operation bindings generate ordinary span-record dashboard questions; schema URL
remains declared metadata because actual `DESCRIBE TABLE otel_traces` exposes
scope name/version but no schema URL column.

- `go test ./...` and `go vet ./...`: passed.
- Frontend suite: 49 tests passed.
- Hegel dashboard, namespace predicate, service suggestions and catalog properties:
  four checks passed (100 generated cases each). Inventory state/swarm regression:
  100 sequences of up to 20 steps passed.
- The initial browser campaign passed four journeys; inventory's final Quint
  comparison rejected its old two-filter expectation after namespace was added.
  The expected ordinary question now explicitly includes the namespace predicate.
- The subsequent campaign passed all four affected regression journeys. Journey
  60 completed its product and desktop/mobile assertions, then exposed an empty
  parameter-list import generated by the Quint harness. Its model now uses the
  existing explicit `INITIAL` constant convention.
- Final affected regressions: journeys 31, 46, 47 and 59 passed. Focused journey
  60 then passed, including missing/empty/named namespace dashboards, actual
  ClickHouse cross-namespace parent SQL, exact native discovery, declared-binding
  role exclusion, desktop/mobile containment, Quint observed replay and four
  rejecting mutations. The existing dashboard replay rejected eight mutations.
- Namespace Bombadil actor: 30-second campaign completed without violations.
  Ledger and actor output: `artifacts/service-namespaces/1792461286328/`.
- Refreshed affected screenshots and added journey 60 storyboard/frames, service
  identity diagram, dashboard design and product-plan updates.
- Final health/source checks: all 12 services running, query/packages image
  `sha256:10c97f0e3fda42c9d458c66fbd15b83d93f3bb79e063e3033da9cce08d2d1817`,
  served UI/binary hashes match source, all five pre-existing stale pending IDs
  preserved. Traffic restored; six total pending entries at the health read
  include active demo work. Evidence: `artifacts/namespace-final-health.json`.
  No Databricks instance was used or claimed.

## Automatic service overviews and distributed relationships (2026-10-20)

[Service dashboards](docs/SERVICE_DASHBOARDS.md) now use a pure Go planner over
supplied observed metadata plus matching server-registry declarations. Existing
record/native evaluators execute bounded widgets with two concurrent requests.
The Services distributed table reuses resolved parent evidence without a new
source read and labels child-span drilldown as a wider population. Widget questions,
explicit captures and shared revisions reuse existing investigation contracts.

Local checks passed: full root Go suite and vet (`dashboard-final-go.log`,
`dashboard-final-vet.log`), pure planning/registry HTTP checks, **48 frontend
tests** (`dashboard-final-node.log`), **100 requested Hegel dashboard planning
cases** and **60 saved-comparison state/swarm scenarios of 15 steps**
(`dashboard-hegel-final.log`), verification vet, Python server syntax compilation
and Quint model typecheck. Planning properties check deterministic order,
capability gating, exact scope/service retention and input detachment. Concrete
native tests reject mismatched units/bindings; unsupported temporality/kinds fail.

The first live campaign passed metric-only browsing but exposed mobile activity
bar overflow and a new test fixture without the kernel-required service instance
identity. The chart now uses at most 60 default intervals with shrinking bars and
no bar for zero-count intervals; the supported-native fixture supplies identity.
The older missing-identity fixture still exercises an unavailable widget rather
than weakening kernel validation. Instrument preview selectors now target the
instrument inventory rather than matching a similarly titled automatic widget.
The deployment ran out of disk during recreation; pruning unused Docker build
cache with the workspace Docker configuration restored space, without deleting
volumes or stream entries. Recreation completed with all services running.

The follow-up regression campaign passed inventory/package handoff, metric-only
browsing and full service-browser scope/cancellation checks (**3 passed**); journey
59's product checks and Quint replay passed, but its actor startup failed by
reading the fixture environment after navigation changed the URL hash
(`dashboard-final-journeys.log`). Correcting the actor to retain that environment
also allows it to inspect actual widget questions and native detail disclosures.
A filtered-list focus regression was fixed by focusing a real service-row action
or the search field, rather than a view-switch button. The new capture assertion
now checks the existing linked-revision status rather than a generic save label.

Final focused journey: **1 passed, 1.4 minutes**
(`artifacts/dashboard-final-focused.log`), with the three existing regressions
already passed on the same build. Real OTLP fixtures cover manifest-free spans,
logs and gauges, resolved and unresolved parents, exact widget question handoff,
immutable Run and capture, source-free fresh-browser reopening, partial log-source
failure and rejection of a late widget response after environment change. Registry
planning also preserves declared-only bindings without inventing observed data.
Frames 31, 46–48 were refreshed; new 59a–d show distributed relationships,
automatic widgets, native values/mobile layout and partial availability.

Observed ledger: `artifacts/service-dashboards/1792459008491/ledger.json`.
Quint replay passed with **eight rejecting mutations**, including invented edges,
changed unit/scope/selector, extra reads and an invalid action sequence. The final
**30-second Bombadil campaign** passed with no violations: 21 view switches,
17 service opens, 28 widget-question inspections, 18 native disclosure toggles
and 16 local returns. The actor validates scope, containment, reported edges and
absence of fabricated totals in failed widgets.

Final image:
`sha256:badca4da5a34205ba719c5db679119052ad578ba6cdac4ca060901e847f37bef`
(`dashboard-native-image.log`, `dashboard-native-up.log`). Final health, served
asset hashes and Go binary hash passed, with all **12 services running** and all
five pre-existing stale Redis pending entries preserved. There were six pending
entries at the health observation with normal traffic restored
(`artifacts/dashboard-final-health.json`); no stream entries or volumes were reset.

Limits: no whole-dashboard custody artifact, graphical/async/declared topology,
automatic SLOs, authoritative fleet membership or declared-only catalog route.
Application contracts retain declared revision IDs rather than claiming runtime
conformance. Record-service identity remains environment/service-name scoped;
native evaluations retain full OTel stream identities. Unsupported native source
identity/intervals stay unavailable/unknown. See the updated whole-product review
and known issues for the remaining roadmap.

## Captured attribute breakdowns and whole-product review (2026-10-20)

[Captured breakdowns](docs/COMPARISON_BREAKDOWNS.md) compare complete one-field
span/log group aggregates, preserve missing/empty/one-sided groups, and open
fixed-side focused questions without source reads. A new dimension requires two
explicit recaptures of the original fixed windows. Frozen receipts use algorithm
v3; v1/v2 receipts reopen without reevaluation. Results precede provenance,
with readable duration/error displays and internally scrolling mobile tables.

Local checks passed: full root Go suite and vet (`breakdown-go.log`,
`breakdown-vet.log`), explicit v1/v2 frozen-receipt compatibility
(`breakdown-legacy-go.log`), **45 frontend tests** (`breakdown-node-final.log`),
**100 requested Hegel partition cases** and **60 saved-comparison state/swarm
scenarios of 15 steps** (`breakdown-hegel-final.log`), verification vet, Python
server syntax compilation and Quint typecheck. The first Hegel invocation used
the default read-only home cache; setting the documented workspace XDG cache fixed
that environment error without changing the property.

The live OTLP route journey exercises 1200/1300 matching records and complete
aggregates beyond the 1000-row inspection cap, explicit recapture, source-free
focused questions, frozen sharing and mobile containment. Its first model replay
exposed the adapter's requirement for a parameterized module import; supplying the
initial-state constant fixed the harness. Replay then passed with six rejecting
mutations (implicit read, changed scope, altered group count, evidence-link drift,
wrong baseline, wrong selected group). An initial regression draft was stopped
to update selectors for the intentional provenance disclosure. A later visual
inspection found that the mobile table squeezed headers into vertical text despite
page containment; minimum readable table widths inside scroll areas fix it, and
the live test now requires an actual group value in view. Interrupted drafts are
retained as `breakdown-journeys.log` and `breakdown-mobile-draft-journeys.log`.

Final live campaign: **8 passed, no skips, 2.5 minutes**
(`artifacts/breakdown-final-journeys.log`). It covers the new route/log breakdowns,
existing record and native comparisons, fresh follow-ups, reusable exact questions,
saved findings, publication recovery and fresh-browser frozen reopening. Frames
52, 53, 56 and 57 were regenerated alongside new 58a–c; the corrected mobile frame
shows readable columns in an internal scroll area, with an actual group in view.

Observed ledger: `artifacts/comparison-breakdown/1792456617714/ledger.json`.
Quint replay passed with **six rejecting mutations**. The **30-second Bombadil
campaign** passed, recording 319 provenance-toggle and 275 focused-question
inspection actions without invariant failures. These inspect saved links and
receipt disclosure; the actor does not execute new telemetry reads.

Final image:
`sha256:476e83d5d4b7a7e537a30a5b4e4d55b803db6358496ea55b3e5ebfd95340445d`
(`breakdown-mobile-image.log`, `breakdown-mobile-up.log`). Final source/binary
hashes, HTTP health and all **12 running services** passed, with exactly the five
pre-existing stale pending Redis entries preserved
(`artifacts/breakdown-final-health.json`). Traffic was restored by the wrapper;
no volumes or stream entries were reset.

The [whole-product review](docs/WHOLE_PRODUCT_REVIEW.md) covers navigation,
discovery, service/distributed overviews, manifests, measurements and investigation
flows. Automatic service/distributed dashboards are **designed, not implemented**;
the next sequence is a bounded capability-driven planner, service defaults,
distributed relationships and investigation integration. The narrower
[comparison review](docs/COMPARISON_PRODUCT_REVIEW.md) records this feature's
usability rationale and remaining proposals.

## Fresh follow-up comparison (2026-10-20)

[Saved questions compare with fresh evidence](docs/FRESH_COMPARISON.md) through a
fixed same-duration preview, explicit compatible prior selection and one new
capture. The baseline is not refreshed. Existing stored comparison and finding
publication preserve the original question and evidence links. Native scopes
retain exact duration without inferring an SDK export phase.

Local checks passed: full root Go suite and vet (`fresh-comparison-go.log`,
`fresh-comparison-vet.log`), all **44 frontend tests**
(`fresh-comparison-node.log`), **100 requested cross-language fresh-scope cases**,
**100 immutable-revision cases**, **100 observed-support cases** and **60 saved
comparison state/swarm scenarios of 15 steps** (`fresh-comparison-hegel-final.log`),
verification vet and Quint typecheck. The fresh-scope property invokes the actual
browser projection in Node and validates it through Go's typed-step contract,
checking exact duration, definition retention and input immutability.

Early browser drafts corrected disabled-option and textarea-value assertions
after inspecting the DOM. The native mobile case then exposed a real overflow:
a long selected baseline label expanded its inline containing label. Block labels
and a width-constrained selector fix the layout without hiding content. Prior
scope, source-read time and custody time are now explicit in the preview.

Final live campaign: **12 passed, no skips, 5.4 minutes**
(`artifacts/fresh-comparison-final-journeys.log`). The new journey uses real
standard OTLP HTTP spans and system-memory gauges. It proves preview and baseline
selection cause no telemetry read; capture creates one new evidence request; a
forced comparison GET failure retries without recapture; changed-definition prior
evidence is disabled; original questions, evidence links and baselines remain
unchanged; and findings reopen source-free in a fresh browser. Native mobile
containment passes with the selected long baseline label.

Observed ledger: `artifacts/fresh-comparison/1792455384426/ledger.json`. Quint replay
passed with **six rejecting mutations**. The **30-second Bombadil campaign** passed
with 177 preview, 169 prior-selection and 181 close-preview actions. Frames 57a–d
show scope/clock preview, actual comparison, recipient finding and mobile native
follow-up. Affected frames 38 and 50–56 were regenerated. Existing shared
measurements, fixed reruns, linked history, saved comparisons and native support
discovery all passed in the same campaign.

Final image:
`sha256:f8cd5f6c377bed933fa5c336216ce7ad09999af32b79fbbc982afffb4a221161`
(`fresh-comparison-fixed-image.log`, `fresh-comparison-fixed-up.log`).
Final checks confirm all **12 services running**, HTTP health, served source asset
and Go binary hashes matching the tested build, and all five pre-existing stale
pending Redis IDs preserved (`artifacts/fresh-comparison-final-health.json`). The
wrapper restored the owned traffic generator; no volumes or stream entries were
reset.

## Reusable comparison questions (2026-10-20)

[Comparison sides become reusable questions](docs/REUSABLE_COMPARISON_QUESTIONS.md)
with a detached typed step, unique question key and exactly that capture linked.
Staging and shared reopening do not read telemetry. Existing immutable document
publication and fixed Run and capture append evidence in a child revision. Fresh
window preview now sits beside the time controls and retains the definition while
proposing a new scope, leaving the saved question unchanged.

Passed local checks: full root Go suite and vet (`reusable-questions-go.log`,
`reusable-questions-vet.log`), all **43 frontend tests**
(`reusable-questions-node.log`), **100 requested Hegel exact-scope/revision cases**
and **60 saved-comparison state/swarm scenarios of 15 steps**
(`reusable-questions-hegel.log`), verification vet and Quint model typecheck.
Frontend checks cover detached evidence/definition objects, exact native scope,
record questions and malformed/unsupported receipt rejection.

The first Quint draft used reserved action names; the renamed model typechecks.
The initial live run reached successful fixed rerun, then exposed that fresh-window
preview was hidden inside the package editor. That owned run was interrupted and
its wrapper restored traffic. Moving the existing preview button beside the time
controls makes the intended workflow directly accessible without opening YAML.

Final live campaign: **11 passed, no skips, 4.0 minutes**
(`artifacts/reusable-questions-final-journeys.log`). It covers record/gauge/native
comparisons, mixed question steps, linked evidence, shared measurements, reusable
questions, saved findings and observed-window discovery. Storyboard 56 uses real
standard OTLP counters with **123 ns** offsets. Adding both sides, retrying a
failed publication and reopening a shared exact question perform no telemetry
read; fixed Run and capture appends one evidence ID in a child revision; fresh
preview retains the definition and requires explicit execution for its new scope.
Earlier revisions and captures remain unchanged; mobile containment passed.

Observed ledger: `artifacts/reusable-questions/1792454583766/ledger.json`. Quint
replay passed and all **six deliberate mutations were rejected**. The **30-second
Bombadil campaign** passed with 120 baseline-question actions and 114 draft
revert actions, checking linked evidence and the card bound. Storyboard 56 frames and
affected frames 38 and 50–55 were regenerated from the running app. Architecture,
workflow docs and the implementation plan describe the implemented contract.

Final image:
`sha256:1bc22f7438d40cba68f4468177c3f6d9a5b518a234e1972bdc4284dd888f7f86`
(`reusable-questions-final-image.log`, `reusable-questions-final-up.log`).
Final health checks passed for all **12 running services**, HTTP health, Go binary
and served HTML/JS/CSS hashes matching tested sources, and preservation of all
five pre-existing stale pending Redis IDs
(`artifacts/reusable-questions-final-health.json`). The wrapper restored the owned
traffic generator; no stream entries or volumes were reset.

## Observed native window discovery (2026-10-19)

[Find comparable windows](docs/OBSERVED_WINDOW_DISCOVERY.md) reads validated
retained kernel intervals without a telemetry query. It intersects boundaries
across all retained streams, previews exact support and exclusions, requires
explicit selection on both sides and rereads two new exact questions. New results
are compared through the existing engine; original captures remain unchanged.

Local checks passed: full root Go suite and vet (`window-discovery-go-final.log`,
`window-discovery-vet-final.log`); all **42 frontend tests**
(`window-discovery-node-final.log`); **100 requested Hegel support properties**,
**100 native exact arithmetic/histogram properties** and **60 saved-comparison
state/swarm scenarios of 15 steps** (`window-discovery-hegel-final.log`);
verification vet and Quint model typecheck. Unit coverage includes source-offline
HTTP discovery, strict request validation, reset taint, all-member coverage,
asynchronous boundaries, layout changes, empty histograms, aggregate overflow,
10,000 intervals and explicit 32-window/64-exclusion output bounds.

The first Hegel invocation omitted the configured writable XDG cache and failed
before exercising properties; the configured-cache rerun passed. The first Quint
model used a built-in action name; renaming it fixed typecheck. The initial live
browser workflow and six rejecting Quint controls passed, but the new Bombadil
invariant rejected the pre-result null state. Its guard now allows that initial
state while continuing to reject an implicit comparable result.

Final live regression campaign: **6 passed, no skips, 3.7 minutes**
(`artifacts/window-discovery-final-journeys.log`). Real standard OTLP cumulative
counters and delta explicit histograms carry **123 ns** SDK offsets. Browser
checks prove discovery/selection do not read telemetry, both selections are
required, capture performs two explicit reads, exact shared questions restore
without telemetry reads, original captures remain unchanged and saved receipts
reopen in a fresh browser. A forced second-capture failure retains the first ID
and leaves the original comparison selected. Mobile containment passed.

The new observed ledger is
`artifacts/window-discovery/1792453916512/ledger.json`. Quint replay passed with
all six rejecting mutations; the **30-second Bombadil campaign** passed, recording
1 compare, 141 discovery and 170 selection/clearing actions. Existing record/gauge,
complete-native-window and saved-comparison journeys also passed. Frames 52–55
were regenerated as affected; storyboard 55 includes exclusions, exact selection,
actual recapture/comparison, mobile histogram support and partial recovery.

Final offline-built image:
`sha256:db18045a876d23a228443a9a6420d60c24fd1065988a244771523abf7c89f2e6`
(`window-discovery-final-image.log`, `window-discovery-final-up.log`). Health
checks confirm all **12 services running**, HTTP health, served JS/CSS/HTML and
Go binary hashes matching tested sources, and preservation of all five pre-existing
stale pending Redis IDs (`window-discovery-final-health.json`). The wrapper
restored the owned traffic generator; no volumes or stream entries were reset.
Only unused Docker build cache and dangling images were pruned to recover disk.

## Exact native scope transport (2026-10-19)

Implemented optional native nanosecond question bounds, parameterized Nano source
predicates, exact kernel evaluation, capture replay and shared-step round trips.
Passed full root `go test ./...` and all 41 frontend tests, including exact-bound
plan/evaluation/replay and shared-link/capture round trips with rejecting controls.
Logs: `artifacts/exact-native-scopes-go.log` and `exact-native-scopes-node.log`.
The discovery picker is not implemented in this milestone. No live deployment
or screenshot regeneration was performed; no new interactive control is added.

## Complete native window comparisons (2026-10-19)

[Native window comparison](docs/NATIVE_WINDOW_COMPARISON.md) uses complete validated
coverage for exact OTel streams. Counter rates use exact window increments and
actual durations; histogram comparisons combine compatible bucket populations
and retain p95 ranges. Missing members, resets, gaps, incomplete boundaries,
temporality changes, incompatible layouts and histogram count overflow withhold
changes. One unsupported member withholds all deltas. New receipts use algorithm
v2; v1 snapshots remain readable without reevaluation.

Passed local checks: full root Go suite including counter exact totals beyond
uint64, histogram overflow refusal, whole-population p95s, multi-stream withholding
and v1 receipt compatibility (`artifacts/native-windows-go-final.log`); 40 frontend
checks (`native-windows-node.log`); 100 requested Hegel exact-rate,
share-conservation/layout-refusal, reversal and immutability cases plus 60 saved
comparison state/swarm scenarios of 15 steps (`native-windows-hegel-complete.log`);
verification vet and Quint model typecheck (`native-windows-vet-final.log`,
`native-windows-model.log`). An initial property filename used the Go Windows-only
suffix and was renamed before accepting results. The first actual run then exposed
an incomplete test stream fixture; adding its required instrumentation scope fixed
that fixture. Both named Hegel tests executed and passed afterward.

Focused live journey: **1 passed, no skips, 1.3 minutes**
(`artifacts/native-windows-focused.log`). It ingests real standard OTLP delta and
cumulative sums and explicit histograms, compares complete windows, shares a v2
snapshot in a fresh browser, checks mobile layout and refuses counter decrease,
coverage gaps and bucket mismatch. Live Quint replay passed with six rejecting
mutations; 30 seconds of Bombadil frozen-result exploration passed. Initial ledger:
`artifacts/native-windows/1792448783045/ledger.json`.

Compatibility campaign: **5 passed, no skips, 3.6 minutes**
(`artifacts/native-windows-release.log`), covering record/gauge comparison 52,
native windows 54 and saved sharing/recovery 53. All enabled Bombadil exploration
and live Quint checks passed. Visual review then improved native table scrolling
on mobile and reused the existing captured-attribute viewer for stream identity.
The final affected journey passed again (**1 passed, no skips, 1.1 minutes**,
`native-windows-polish-focused.log`), including attribute disclosure and actual
horizontal scrolling, plus six rejecting trace controls and Bombadil. Its ledger
is `artifacts/native-windows/1792449184577/ledger.json`. Five new focused frames and
affected 52/53 screenshots are updated; all new frames were visually reviewed.

Final health confirms 12 running services, functional query health, restored owned
traffic, all five pre-existing stale Redis pending IDs retained, and served assets
matching source (`native-windows-final-health.json`, `native-windows-served-assets.json`).
Final query image:
`sha256:f234be3fd02517e080c77f1e94c7bb4c5f15b8e3c525c1f052763cf48bbcac29`.
Exact native coverage remains deliberately strict; millisecond question bounds
can reject nanosecond-offset interval edges, as documented in known issues.

## Saved comparison evidence (2026-10-19)

[Saved comparison cards](docs/SAVED_COMPARISONS.md) freeze the server-verified
ordered capture IDs, displayed decision/result and algorithm version. Findings
revise in immutable documents. Shared investigation links read the saved snapshot
without querying telemetry or reevaluating the comparison. Retained receipt IDs
support explicit recovery after context changes; failed document publication
retains the draft and finding.

Passed local checks: full root Go suite including source-free HTTP publication,
forged-decision rejection, missing references, integrity checks and preserved
finding history (`artifacts/saved-pair-go-http.log`); 40 frontend checks
(`saved-pair-node.log`); 60 requested Hegel state/swarm scenarios of 15 steps
(`saved-pair-state.log`); verification vet (`saved-pair-vet-final.log`), Quint
model typecheck (`saved-pair-model.log`) and Python server compilation.

The first live campaign completed all sharing, recovery, mobile and native UI
assertions, and Quint trace validation with six rejecting controls, but its new
Bombadil actor used an overly broad last-child selector that clicked Remove.
Targeting the saved comparison's direct reopen button and adding a card-count
invariant corrects the test actor. That campaign reports one failed/one passed;
its passing ledger replay is
`artifacts/saved-comparison/1792447998043/TestQuintSavedComparisonLive.log`.

Final live campaign: **5 passed, no skips, 3.1 minutes**
(`artifacts/saved-pair-release.log`): notes/context 36, both comparison journeys 52,
and both new sharing journeys 53. All enabled Bombadil campaigns and live Quint
validations passed. The final ledger is
`artifacts/saved-comparison/1792448210686/ledger.json`; its six rejecting controls
cover implicit reads, changed receipt references, forged decisions, lost findings,
false publication transitions and reversed capture identity. Both native gauge
snapshots and frozen unit incompatibility reopen in a fresh browser without engine
requests. Four focused storyboard screenshots were captured and visually reviewed;
affected 36/52 frames were regenerated.

All 12 services are running, query health passed, owned traffic was restored, and
all five pre-existing stale Redis pending IDs were retained
(`artifacts/saved-pair-final-health.json`). Served document/comparison scripts and
styles match source (`saved-pair-served-assets.json`). Final query image:
`sha256:e3dc20bac4df747e4230b10be5747cc6dbbf4d19b19118f669740d7b01823bc7`.

## Stored capture comparison (2026-10-19)

[Stored comparison](docs/CAPTURE_COMPARISON.md) loads immutable capture IDs and
compares full record counts, span duration/error summaries and exact-stream native
gauge endpoints without telemetry reads. Incompatible definitions, environments
or units withhold deltas. Empty/unknown results retain their uncertainty; native
counter/histogram/lifecycle results remain inspectable side by side.

Passed local checks: full root Go suite, 40 frontend checks, 100 requested Hegel
precision/reversal/immutability/compatibility cases, verification vet, Quint
model typecheck and Python server compilation. Logs:
`artifacts/evidence-pair-{go,node,hegel-ready,vet,model}.log`.
The live ledger replay passed with four rejecting controls
(`artifacts/evidence-pair-quint-replay-ready.log`). The first trace-validation
attempt exposed the checker's generated empty-argument import for a model without
constants; declaring the explicit initial read count fixed the model adapter.
The first Hegel attempt used an unwritable default cache; setting the documented
workspace `XDG_CACHE_HOME` fixed the setup failure. The first live browser campaign
found a missing entry for the new script in the Python server's static asset
allowlist; the corrected build serves source-matching assets.

Final live campaign: **5 passed, no skips, 3.4 minutes**
(`artifacts/evidence-pair-release.log`). This covers saved captures 32, both new
comparison journeys 52, and both linked-history journeys 51. Enabled Bombadil
campaigns and live Quint validations passed. The new observed ledger is
`artifacts/evidence-comparison/1792447023081/ledger.json`; its four negative
controls reject hidden reads, incompatible deltas, forged arithmetic and changed
receipts. Final native expansion retains explicit Baseline/Comparison labels.
Storyboards 32/51 were refreshed with the changed UI, and four new 52 screenshots
were captured and visually reviewed (record, incompatible, mobile, native).

Final health checks confirm 12 running services, a functional query health request,
restored owned traffic, all five pre-existing stale Redis pending IDs preserved,
and source-matching served HTML/JS/CSS (`evidence-pair-final-health.json`,
`evidence-pair-served-assets.json`). Final query image:
`sha256:8e031dc91f396851ce4e9865975b048d9da9df1d60cabf10a07842d86116ed03`.
Build disk pressure was resolved by clearing unused build cache and dangling
images; telemetry volumes and pending stream entries were retained.

## Linked question evidence (2026-10-19)

[Linked evidence](docs/LINKED_QUESTION_EVIDENCE.md) gives saved questions an
explicit Run and capture action, immutable capture history, stored result counts
and source read times, and visible question/scope differences. Partial publication
preserves the capture ID for explicit recovery without another source read.

Passed: full root Go suite, 40 frontend checks, verification vet, 100 requested
Hegel cases comparing the actual JavaScript scope/definition logic, and 60
requested state/swarm scenarios of 15 steps preserving immutable revisions.
Logs: `artifacts/linked-go-summary.log`, `linked-node-ready.log`,
`linked-verification-vet.log`, `linked-hegel.log`, `linked-state-final.log`.

Both new browser journeys passed (no skips, 1.1 minutes): record capture history,
refined-question drift, failed-publication recovery, stored reopening, mobile
layout, and native capture with one source execution. The live ledger passed Quint
trace validation with seven rejecting mutations; 30 seconds of Bombadil exploration
passed without implicit source reads (`artifacts/linked-release-focused.log`,
`artifacts/linked-evidence/1792445673965/ledger.json`). The first focused run exposed
mobile overflow from an unwrapped capture ID; using the existing reference-ID
style fixed it, and the final mobile assertion passed. Storyboard 51 and its five
screenshots document the final behavior.

Final compatibility campaign: **4 passed, no skips, 5.2 minutes**, covering
saved captures 32, investigation notes 36, mixed steps 50 and measurements 38,
with their enabled Bombadil campaigns and live trace validations
(`artifacts/linked-regression.log`). Refreshed affected screenshots accompany the
new storyboard. All 12 services are running, owned traffic was restored, all five
pre-existing stale Redis pending IDs were preserved, and the three served UI assets
match the tested source (`linked-final-health.json`, `linked-served-assets.json`).
Final query image: `sha256:495e49a385806d103d9c398b3dda90f1c245b9c0490b2217bfde524478ffac3e`.

## Typed investigation steps (2026-10-19)

[Question steps](docs/INVESTIGATION_STEPS.md) persist record questions and complete
measurement definitions in named immutable investigation revisions. The common
versioned scope/view envelope rejects mismatches between a record/definition and
its declared environment/window. Question intent stays separate from frozen
capture references. Both record and native question URLs stage explicit Run again;
native reopening performs definition planning without telemetry reads. Legacy
record URLs and existing stored document IDs remain compatible.

Passed local checks: full root Go suite (including mixed-card validation and
immutable predecessor checks), 39 frontend checks, verification vet, and 100
requested Hegel cases through the actual JavaScript envelope codec. Logs:
`artifacts/steps-{go,node-final,vet-final,hegel}.log`.

The first focused browser run completed all mixed-question UI assertions, source
reads, capture attachment, second-browser reopening and screenshots, but its new
Quint model used a reserved action name. Renaming the modeled explicit-read action
fixed it. Replaying that real ledger passes, with seven rejecting controls for
scope drift, native unit/identity changes, hidden reopens/recipient reads, false
window equivalence and wrong observed read transitions (`steps-quint.log`). Both
existing question-link checks passed in that initial focused run
(`steps-focused.log`). The model checks observed save → reopen → explicit-read
transitions, preserving source read counts across saving and reopening.

Final live campaign: **11 passed, no skips, 11.7 minutes**
(`artifacts/steps-release.log`). This covers coherence 40, notes/context 36, core 33,
inventory 31, new mixed questions 50, measurements 38, metric-only services 47,
both question-link cases 49, workbench 45 and service browser 46/48. All enabled
Bombadil campaigns and associated live Quint validations passed. The new actor
returns to the active investigation and reopens each question type without running
it. The mixed-step ledger is
`artifacts/investigation-steps/1792444529620/ledger.json`.

The existing Hegel investigation state/swarm model now also creates record/native
question cards and revises detached question scopes after saving/forking. Its
60 requested scenarios of 15 steps passed, preserving all prior revisions and
context semantics (`steps-state-swarm.log`); verification vet passed again afterward.
New frames 50a–c and affected existing frames were regenerated from real telemetry.
The final desktop investigation and mobile native-question frames were visually
reviewed. README, architecture/process flow, contracts, plans and known boundaries
are synchronized.

Deployed image:
`sha256:5bae60c988431e0ec7edcbc9dab734428fe303e16960f1a3270f9d0eb06f4fd2`.
All 12 services run, traffic is restored, all five stale pending entries remain,
and served navigation assets match the checkout (`steps-final-health.json`).
No telemetry volumes were removed.

## Resumable record-question links (2026-10-19)

[Question links](docs/QUESTION_LINKS.md) preserve fixed record scope, selectors,
grouping, measures and Records/Distribution view. Reopening or traversing history
stages Run again and clears live evidence without a source read. Capture receipt
links remain a separate frozen-evidence workflow. Native drafts, service-detail
panels and unsaved keystrokes are outside this first link format.

Passed: full root Go suite, 38 frontend checks, verification vet, and 100 requested
Hegel cases through the actual URL codec and session state. The observed browser
read ledger passes quint-go validation and rejects three mutations (restore read,
foreign fixed window, history read). Logs: `artifacts/navigation-{go,node-final,
vet,hegel-final,quint}.log`.

The first browser run completed its UI assertions and screenshots but failed on a
new Quint model import with no constants. Adding its explicit initial-state constant
fixed that setup. The first Hegel invocation used the read-only default cache;
rerunning with the established XDG cache path passed. Initial logs remain in
`navigation-journey.log` and `navigation-hegel.log`.

Final browser evidence: seven existing journeys passed across
`artifacts/navigation-release.log` (captures, workbench, service browser) and
`artifacts/navigation-final.log` (coherence, core, inventory, metric-only services).
Their seven enabled Bombadil campaigns and live Quint checks passed. The new
link journey initially exposed an actor trying Back before creating in-app history;
its actions now stay within their own entries. A retry hit Bombadil's existing-output
safeguard; runs now retain separate trace directories rather than overwrite evidence.
Those earlier failures remain in the two campaign logs.

The final image adds sender-side draft-window pinning. Both link browser tests
passed on it, including the corrected eighth Bombadil campaign and live Quint check
with all three rejecting mutations (`artifacts/navigation-links-final.log`,
52.5 seconds, no skips). The final frontend rerun again passed all 38 checks
(`navigation-node-shared.log`). Storyboard 49 uses actual populated `lab-web`
telemetry, a second browser, fixed-window history, a duration refinement, malformed
links, view retention and mobile containment. Existing affected frames 31–33,
40 and 45–48 were refreshed. Desktop and mobile frames were visually inspected.

Final deployed image:
`sha256:9d761f84bbe610e2a9cb4fee88fbfbc7be5e7009551ad7b232030071c5e00c97`.
All 12 services run, traffic is restored, and all five stale pending entries remain
(`artifacts/navigation-final-health.json`). Unused Docker build cache was pruned;
telemetry volumes were retained. Served navigation assets match checkout hashes.

## Unified Services catalog (2026-10-19)

[Services](docs/UNIFIED_SERVICE_CATALOG.md) now has one discovery action and local
search, with one row per observed service. Span/log/native point counts retain
separate grains and exact decimal strings; absent observations are not zeros.
Service detail independently reads signals, span attributes, native instruments
and environment-wide operation inventory in the original catalog window. Operation
inspection and capture actions moved inside detail. The inventory capture label
explicitly describes its original whole-environment scope. Saved inventory rendering
is source-free; capture formats, backend kernels and query semantics are unchanged.

Passed: root Go suite, 36 frontend checks, Python compilation and verification vet.
Hegel ran 100 requested catalog-grain cases plus 100 suggestion-scope cases through
the actual JavaScript planners. The catalog property includes UInt64 values beyond
JS number precision and service/signal grouping. Logs:
`artifacts/catalog-{go,node,hegel,vet}.log`.

Final live integration: all seven journeys passed in 11.3 minutes, without
skips (`artifacts/catalog-release.log`): captures 32, coherence 40, core basics 33,
inventory 31, native-only services 47, question workbench 45 and service/catalog
46/48. All seven enabled Bombadil campaigns and associated live Quint checks
passed. The updated service ledger rejects eight mutations, including summed/mixed
signal counts and a foreign catalog window. The browser also deliberately tampers
a real catalog response's window and confirms rejection before publication.
Ledger: `artifacts/service-browser/browse-aafd713e/ledger.json`.

Storyboards 31/32/33/40/45/46/47 were migrated and affected frames regenerated.
New frames 48a–c show one entry, retained filtered return and mobile catalog;
desktop/mobile catalog frames were visually reviewed. The list-return path checks
read-free navigation and search focus when no row matches. Architecture/process
diagrams, README, current contracts, planning and known boundaries are updated.
Final deployed image: `sha256:2ce745c0a8e180bd06aba310c7f086b5cc34880fcc557376cf7963c538b0a469`.
All 12 services are running, traffic is restored, and all five pre-existing stale
Redis pending entries remain (`artifacts/catalog-final-health.json`). No telemetry
volumes were removed.

The first focused run passed the native-only journey, then found an old browser
test trying to reopen a service while its catalog was hidden behind detail. The
journey now returns to the catalog explicitly before opening again. That initial
attempt is retained in `artifacts/catalog-focused.log`.

## Metric-only service discovery (2026-10-19)

[All-signal service discovery](docs/METRIC_SERVICE_DISCOVERY.md) now includes native
gauge, sum and explicit-histogram data-point records. It groups presence by service
without combining metric values, units or histogram observations. A direct Services
entry opens the catalog; metric rows open the original fixed window in service
browser/native measurement flow, never as a record query. Dependency inference
remains trace/parent-span only. Summary/exponential-histogram families remain outside
the current adapter, and no complete fleet or health claim is introduced.

Passed: full root Go suite, discovery race checks, root and verification vet,
35 frontend tests, Python compilation and Compose configuration validation.
Hegel generated 100 requested valid scope/count/bound cases and verified rejection
of the exclusive upper event bound. Unit cases cover malformed/duplicate rows,
zero counts, event bounds, UInt64 decimal counts, parameter isolation and the
1000-pair global cardinality budget. Quint model typechecking passed.
Logs: `artifacts/metric-services-{go,race,root-vet,vet,node,hegel}.log`.

Final integration passed: core basics 33, native metric discovery 39, new
metric-only services 47 and service browsing 46 in 4.8 minutes
(`artifacts/metric-services-release.log`), plus inventory 31 in 1.1 minutes
(`artifacts/metric-services-inventory.log`). All six enabled Bombadil campaigns
and associated live Quint validations passed. New storyboard 47 validates the
actual ledger with seven rejecting mutations: point count, window, service identity,
preview read, local-return read, forged dependency and exclusive event bound.
The histogram with 5000 observations contributes one of four point records;
gauge execution returns its two selected points. Local search remains applied
after an explicit catalog refresh, without altering source totals.
Ledger: `artifacts/metric-services/metric-services-195c85b9/ledger.json`.

New frames 47a–d and affected inventory/core/native/service-browser frames were
regenerated. The new desktop catalog and mobile service detail were visually
reviewed. README, source/process diagram, implementation plan, known boundaries
and executable storyboards are synchronized. Final image:
`sha256:fa69ff90cb9bfb46bac4197878b845c0df03817efca3fedbcb27e197cf6f6f18`.
All 12 services are running, traffic is restored and all five pre-existing stale
Redis pending entries remain (`artifacts/metric-services-final-health.json`).

The first live attempt confirmed cross-family record counts and exact boundary
inclusion using independent ClickHouse SQL, then failed native execution because
the new fixture omitted `service.instance.id`. The fixture now supplies that
required stream identity; the native validator remains intact. The initial attempt
is retained in `artifacts/metric-services-journey.log`. Disk pressure during image
recreation was resolved by removing unused Docker build cache, without removing
telemetry volumes or running images (`metric-services-cache-prune.log`).

## Browse-first service investigations (2026-10-19)

[Service browsing](docs/SERVICE_BROWSING.md) now connects fixed-window service
inventory and span/log discovery with observed-attribute questions and native
instrument previews. Opening a service explicitly reads three independent bounded
sources; unavailable sources remain separate. HTTP/database/messaging suggestions
use exact observed OTel attribute keys (including documented aliases), add an
explicit exists predicate, and enter the shared Explore executor. Native previews
retain instrument identity and fixed bounds; unsupported kernels cannot execute.
Back to service is local. Scope changes abort reads and reject late publication.
Metric-only fleet enumeration and durable navigation history remain unimplemented.

Passed: 35 frontend tests; Python compilation; verification vet; Quint typecheck;
Hegel (100 requested generated cases) invoking the actual JavaScript suggestion
planner with generated windows and preferred/legacy/absent keys. Storyboard 46
passed with real OTLP spans/logs/metrics through ClickHouse, a 30-second Bombadil
campaign, and observed quint-go ledger validation with six rejecting mutations.
The journey checks narrowed HTTP count 4 of 6 spans, preserved service/window,
source-free native preview, read-free return, unsupported nonmonotonic sums,
log-only service browsing and cancellation of a delayed real metadata response.
Evidence: `artifacts/service-{node,hegel,vet,final-journey}.log` and
`artifacts/service-browser/browse-9939f5b3/ledger.json` (release rerun).

Existing core basics 33, inventory 31, metric discovery 39 and question workbench
45 all passed in 4.5 minutes (`artifacts/service-regression.log`), including five
enabled Bombadil campaigns and their live Quint validations. These ran before the
final two presentation fixes for local service-button filtering and unread
operation inventory wording; storyboard 46 checks both on the release image.

The release rerun passed (`artifacts/service-release-journey.log`), including
local service-button filtering without reads and explicit unread operation inventory
for log-only service detail. Its live Quint validation and Bombadil campaign passed;
all six distinct enabled campaigns across this slice and regressions succeeded.
Final image: `sha256:f51dff13fb845ad45b648608bbadeee36af1cea13dab41b6aa1f692e029b50f1`.
All 12 services are running, traffic is restored and all five pre-existing stale
Redis pending entries remain (`artifacts/service-final-health.json`). Only dangling
Docker images were pruned for space; no telemetry volumes were removed.

Frames 46a–d document service detail, fixed-window record exploration, native
measurement preview and mobile detail. Final desktop/mobile service frames were
visually reviewed. The detail replaces the operation table while open, with
question/instrument cards arranged in desktop grids and mobile columns. README,
architecture/process diagram, implementation plan and known boundaries are updated.

Early integration attempts found two fixture/assertion mistakes: monotonic delta
counters are supported (the unsupported fixture now uses a nonmonotonic sum), and
measurement environment is nested in the definition rather than the question root.
Those attempts remain in `service-journey.log` and `service-journey-rerun.log`;
the final successful run is distinct. No backend kernel or persisted schema changed.

## One scoped question workbench (2026-10-19)

The first [entry-point design slice](docs/QUESTION_WORKBENCH.md) is implemented:
shortcuts, advanced filters and measures share one Run question executor. Draft,
requested and applied scopes have generation ownership; editing cancels pending
reads and clears stale evidence actions. Ordinary edits and explicit drilldowns
retain fixed bounds. Fresh-window staging performs no read. Evidence capture and
reusable definitions have separate labels, and Services → native metric discovery
preserves inventory service/environment/bounds without asserting the same population.

Passed: full root Go suite, 34 frontend checks, Python compilation, verification
vet, and Hegel state/swarm (100 requested sequences × 25 steps) replaying the actual
JavaScript session against an independent Go oracle. Quint typechecking and the
new live session/evidence ledger validation passed, including eight rejecting
controls for generation, scope, bounds, local navigation reads and frozen custody.
Storyboard 45 passed through the real OTLP/ClickHouse pipeline, delaying an actual
response to check stale completion rejection; its 30-second Bombadil actor passed.
The structured-attribute desktop/mobile polish journey 29 also passed. See
`artifacts/entry-{go,node,hegel,vet,ledger,focused}.log`; the latest workbench ledger
is `artifacts/question-workbench/entry-6f69f0e8/ledger.json`.

The final regression results cover nine distinct browser journeys: new workbench
45 and attribute polish 29 (`entry-focused.log`), core basics 33, follow-ups 42,
grouped measures 34, grouped time series 43 and recipes 37 (`entry-regression.log`),
plus corrected discovery 30 and numeric measures 44 (`entry-rerun.log`, two passed
in 2.8 minutes). All nine enabled Bombadil campaigns and their associated live
Quint checks passed across those successful journeys. Discovery includes both
its Playwright and Bombadil-observed ledger validations. Existing affected frames
were regenerated. The numeric rerun opens the relocated measure disclosure rather
than waiting on its old direct-child selector.

The final deployed image is
`sha256:628c8735182a85cb1b25556b4a8cfd2f66bc82fd6df8601c7926a4a509e359c3`.
All 12 services are running, traffic is restored, and all five pre-existing stale
Redis pending entries remain (`artifacts/entry-final-health.json`). No telemetry
volumes were removed; only dangling Docker images were pruned for disk space.

Desktop/mobile frames 45a–d document initial entry, applied scope, metric handoff
and frozen evidence. The initial and applied desktop layouts were visually
reviewed; the frozen mobile layout was also visually reviewed. The broader
navigation/Library migration, native browse-first flow and durable navigation/draft
recovery remain proposed, rather than claimed complete.

Early validation found and corrected a Quint action-name collision, Go omitted
empty predicate defaults in returned question identity, service-shortcut loss
when switching signals, and older tests acting on hidden result/disclosure controls.
The first focused discovery attempt also assumed the first log had a trace link;
the mobile check now explicitly selects a linked log. The exploratory discovery
actor also retained a reuse expectation after intentionally changing the breakdown;
its explicit breakdown actions now clear that old intent. These failed attempts are
retained in the local artifact logs.

## Live numeric attribute measures (2026-10-19)

One optional explicit int64/uint64 string-map binding now executes through the Go
ClickHouse adapter, Explore summaries/breakdowns/time charts, immutable captures
and reusable YAML. Missing/empty/invalid values retain separate counts; only valid
values enter exact sum/extrema/rational means. Signed chart rankings use labelled
absolute magnitude with signed values retained. Preflight now reports supported
bindings executable. Multiple numeric bindings, float/decimal parsing, conversion
and numeric cohort comparisons remain unsupported. See [contract and process diagram](docs/NUMERIC_ATTRIBUTE_PLANNING.md).

Passed: full root Go suite; exploration/HTTP race tests; root vet; 29 frontend
checks; Python compilation; Hegel partition conservation and numeric state/swarm
(100 requested cases/sequences, 25 state steps), plus existing package state/swarm.
Native checks cover parameterized keys/units, exact type bounds, impossible summary
states, missing source summaries, same-count sum disagreement, capture rejection
and YAML/typed-plan round trips. Three retained historical source-free plans still
match SQL, canonical YAML and package revisions exactly on the final image.

Controlled integration sends 128 ordinary OTLP spans and 128 logs through the real
collector/exporter. Signed values produce 116 valid, one missing, one empty and ten
invalid records; exact sum is 9223372036854776141. Unsigned and resource bindings,
all groups and every interval are checked against the Go reference, independent of
the 100-row page. A live six-stage custody ledger passed quint-go/Quint validation
and rejected eleven mutations, including coverage, units/parsing, sum, cell bounds,
local reads and reopened/reused evidence. All six final browser journeys passed
without skips in 7.5 minutes: core basics 33, follow-ups 42, grouped measures 34,
grouped time series 43, numeric measures 44 and recipes 37. All seven enabled
Bombadil campaigns and associated live Quint checks passed. Final numeric ledger:
`artifacts/numeric-measures/numeric-deeb5719/ledger.json`. Desktop/mobile numeric
frames and affected existing frames were regenerated; both numeric layouts were
visually reviewed. All 12 services are running, traffic is restored, and inspection
confirmed all five pre-existing stale pending entries remain.

An early browser attempt was interrupted after a time-control blur moved the
editor click target; the journey now waits for the layout change. A subsequent
attempt completed the UI path but launched an unrelated system `go`; the final
campaign uses the pinned Go toolchain. An immediate post-deploy probe also hit a
startup connection reset; the ready-service preflight passed. Visual review led
to wider exact-value tiles and a clear unsupported numeric-comparison message;
enabling a numeric measure clears the earlier count baseline.

Final image: `15d9fcb7d278`. Artifacts: `artifacts/numeric-live-{go,race,vet,node,hegel,compatibility,preflight,ledger,regression}.log` and
`artifacts/numeric-measures/*/ledger.json`. Existing volumes and the five stale
pending entries are retained. Screenshot/storyboard updates are in journey 44
and affected regression frames. No environment configuration changed.

## Numeric attribute preflight and reference semantics (2026-10-19)

Added explicit int64/uint64 attribute/resource bindings, canonical-decimal parsing,
declared units, exact reference sum/extrema/rational-mean states and separate
valid/missing/empty/invalid counts. A source-free HTTP preflight reports structured
capability diagnostics and always marks execution unavailable. Live numeric SQL,
UI controls and package/capture persistence remain future work; existing executable
question/package/capture formats are unchanged. See [contract and process diagram](docs/NUMERIC_ATTRIBUTE_PLANNING.md).

Passed: full root Go suite, exploration/HTTP race tests, root vet, Python compilation
and Hegel partition conservation (100 requested cases). Tests include signed and
unsigned limits, sums above uint64, zero versus unavailable extrema, malformed
syntax, bad unit labels, detached scope and strict HTTP decoding. Reproducible live
proxy checks passed with `python3 scripts/check_numeric_preflight.py`, including
supported declarations, multiple unsupported diagnostics and unknown-field rejection.
No UI behavior changed, so browser journeys/screenshots and Bombadil/Quint live
ledgers were not regenerated for this source-free milestone.

Artifacts: `artifacts/numeric-plan-{go,race,vet,hegel,image,up,live,services,pending}.log`.
Deployed image `bb4ee24d75af`; all 12 default services remain running, including
traffic. Volumes were preserved. Inspection confirmed all five pre-existing stale
pending entries remain; an additional ordinary in-flight entry was present with
traffic active. No telemetry reset or traffic pause was performed.

## Typed record measures and grouped time charts (2026-10-18)

Existing record controls now expand into typed count, duration and recorded-status
operators used for SQL lowering. Input/unit/operator/percentile-profile checks fail
closed; arbitrary attribute casts and general expressions remain unsupported.
Optional `groupTimeSeries` adds bounded group/interval evidence, common-scale local
chart controls, capture/reopen and definition reuse. See [the contract](docs/RECORD_MEASURES.md).

Passed: full root Go suite, exploration/HTTP race tests, root vet, 32 frontend
checks, Python compilation and expanded Hegel state/swarm (100 requested sequences
of 25 steps each for grouped-series, recipes and follow-ups). Native tests cover
missing/empty dimensions, no-data cells, duplicate/invalid bounds, cardinality/cell
budgets, same-count duration disagreement, signal-specific lowering and unsupported
measure bindings. Three source-free plans retained from the prior deployed image
match the final image's SQL, package revisions and canonical YAML exactly.

The initial real-stack storyboard passed in one minute with seven rejecting Quint
controls and its Bombadil campaign (`artifacts/grouped-time-series/series-d6d2af76/`).
It sends 124 ordinary OTLP server-shaped spans through the real collector/exporter,
compares four route-value series with a 100-row inspection page, checks local scale
and no-read guards, and captures/reuses/reopens the evidence. This is controlled
integration telemetry, not production traffic or a convention-coverage certificate.
Screenshot inspection exposed a checkbox style issue, corrected before regression.

The build filled the workspace disk. Regenerable Go compilation cache, dangling
images and unused Docker build cache were removed; telemetry volumes and running
application data were preserved, restoring approximately 14 GiB free before caches
were warmed. An intermediate test-file editing error was caught by compilation and
corrected; the final full native suite passed. No environment configuration changed.

Artifacts: `artifacts/series-{go,race,vet,node,hegel,compatibility}.log`,
`artifacts/series-journey.log`, `artifacts/series-regression.log` and the live
storyboard ledgers. Final image: `bfb62c257374`. All five affected browser journeys passed in 5.9
minutes, without skips: core basics 33, follow-ups 42, grouped measures 34,
grouped time series 43 and recipes 37. All six enabled Bombadil campaigns and
associated live Quint checks passed. The final grouped-series ledger rejected all
seven mutated controls: `artifacts/grouped-time-series/series-68ffb67e/`. Final
ledger replay also passed after strengthening exact integer accounting and total
series validation (`artifacts/series-ledger-final.log`). Affected
desktop/mobile frames were regenerated. Traffic was restored; existing telemetry
volumes and all five stale pending entries were retained.

## Shared question scope and checked transforms (2026-10-18)

Recipes and capture-derived follow-ups now use shared detached Go planning metadata.
Same-population/same-question and exact-ID relationship claims are checked against
actual typed questions; grouping/measure/chart/page changes are computed. Record
scope excludes presentation paging and retains complete typed OTel/custom selectors.
Native measurement bindings retain declared schema/library/instrument identity;
unknown record schema or convention profile remains explicit. Inventory and
comparison adapters retain underlying grain and separate member windows.

Passed: full root Go suite; exploration/HTTP race tests; root vet; 32 frontend
checks; Python compilation; expanded Hegel recipe and follow-up state/swarm
(100 requested sequences each, 25 steps). Checks cover detached ownership,
OTel/custom namespaces, record/package/wire scope roundtrips, native binding
identity, false-claim rejection and unchanged capture serialization. The first
Hegel invocation used the unwritable default cache; rerunning with the existing
workspace XDG cache passed. No environment configuration change was required.

Artifacts: `artifacts/scope-{go,race,vet,node,hegel}.log`. The deployed offline
image is `15330826a024`. Initial image `b501ce099b8c` passed all three affected
browser journeys in 3.7 minutes, with all enabled Bombadil and live Quint checks.
Final review separated metric signal from gauge/counter/histogram measurement kind;
root/race/vet/Hegel checks passed again after that correction. Final image passed both planning journeys again (2.4 minutes), with no skips,
including both Bombadil campaigns and live Quint checks. Follow-up replay rejected
all nine mutated ledgers; recipe replay checked the full shared transform plan.
Final evidence: `artifacts/scope-final-journeys.log`,
`artifacts/followups/47900071/`, `artifacts/recipes/c1bc9301/`. The earlier broader
run is `artifacts/scope-journeys.log`. Desktop/mobile screenshots were regenerated;
the desktop follow-up preview was inspected. Traffic was restored after both
wrappers; existing telemetry volumes and all five stale pending entries remain.
The [contract](docs/QUESTION_SCOPE.md), diagrams, planning docs and existing
storyboards describe this bounded slice. General typed measures, convention-profile
resolution and full backend capability negotiation remain future work.

## Core primitives review (2026-10-18)

Reviewed the implemented question/package/capture, workflow/recipe/follow-up and
investigation models. [The consolidation plan](docs/CORE_PRIMITIVES.md) records
missing shared contracts, migration boundaries and next-slice acceptance criteria.
The plan also explicitly requires OTel convention profiles across all six primitives,
with versioned bindings, optional-field degradation and conformance cases.
Documentation only: changed-file local links and `git diff --check` passed; no
runtime changes, new browser claims or application test reruns in this review.

## Contextual investigation follow-ups (2026-10-18)

Evidence cards now propose source-free questions from verified captures, with
explicit preserved and changed scope. Go compiles generated record/measurement
questions; exact trace/span IDs remain distinct from service/time resource
context. Discovery retains the original fixed window through package preview.
Authored evidence gaps cite the originating capture and remain unsaved drafts.
See [the contract](docs/INVESTIGATION_FOLLOWUPS.md) and storyboard 42.

Passed: full root Go suite, exploration/HTTP race tests, root vet, 32 frontend
checks, Python compilation and Hegel state/swarm (100 requested sequences, 25
steps). Hegel exposed shared predicate trees between returned actions; each
query now owns an independent deep copy, with a unit regression and passing
property rerun. Stored capture immutability is checked separately.

The focused real-stack journey passed in 1.4 minutes: reproduced delayed sample
work, compared duration/status over the same matching population, followed exact
trace/log IDs, explicitly discovered instruments in the original window and
saved a finding with an origin-cited instrumentation question. It does not infer
resource causality. Its quint-go ledger passed and rejected seven altered ledgers;
the new Bombadil source-free preview campaign passed. The initial attempt failed
only a textarea assertion using text content instead of value; corrected assertions
passed. Initial artifacts remain in `artifacts/followups-initial-failure/`.

Evidence: `artifacts/followups-{go,race,vet,node,hegel}.log`,
`artifacts/followups-journey-final.log`, and
`artifacts/followups/7cc4cfd1/`. Final image `77be6b447ade`: all six affected browser journeys passed in 5.5
minutes, with no skips: coherence 40, context notes 36, follow-ups 42,
investigation workspace 41, measurements 38 and metric discovery 39. All five
enabled Bombadil campaigns and associated live Quint/negative controls passed.
The expanded mobile preview passed containment and source-free checks. Affected
screenshots were regenerated. Final ledger: `artifacts/followups/aec68658/`;
regression log: `artifacts/followups-regression.log`. The wrapper restored owned
traffic and preserved existing telemetry volumes and five stale pending entries.

## Question-first investigation workspace (2026-10-18)

Investigations now retain a question, current finding, open questions and ordered
cross-signal evidence. An active title/question context follows navigation.
Explore's capture dialog can preserve only the capture, append to the current
investigation draft, or publish a child of an explicitly selected saved revision.
Unrelated dirty drafts remain intact; capture and attachment publication are
separate, and attachment failure leaves the capture available. Cards show stored
receipt scope and native definition identity through a receipt-only API, without
returning record banks or querying telemetry.

Passed: full root Go suite; affected workbench/HTTP race suites; root vet; all
32 frontend checks; expanded Hegel context state/swarm (60 requested sequences,
15 steps) covering complete brief mutation, fork and preserved history. Explicit
storage tests preserve the pre-brief version-1 hash/JSON shape, normalize omitted
question-only card lists and reject oversized summaries. The receipt HTTP test
runs with its telemetry source closed and verifies that no source read occurs.

Initial storyboard 41 passed (1.2 minutes) with real checkout/report spans, report
logs and an independent ordinary OTel counter. Its dedicated quint-go ledger
passed and rejected six mutated ledgers. Final deployed image `d13fcbdbce4a`:
all four browser journeys passed, with no skips, including captures 32, context/
notes 36, investigation workspace 41 and measurements 38. All four enabled
Bombadil campaigns and the associated live Quint/negative-control checks passed.
The final new journey additionally creates a saved branch while retaining an
unrelated dirty question draft, with no telemetry read. Fresh metric definition
reuse adds a capture/child revision while the prior capture and handoff reopen
unchanged. Desktop/mobile containment and screenshots cover the connected flow.

Artifacts: `artifacts/investigation-{core-go,core-node,race,vet,hegel}.log`,
`artifacts/investigation-workspace-journey.log`,
`artifacts/investigation-final-journeys.log`, and
`artifacts/investigation-workspace/3d3506c8/`. The diagrams, contract, planning docs,
README and affected storyboards were updated. No telemetry volumes were reset;
the five pre-existing stale pending entries were retained and traffic restored.

## Product essence and connected investigation flow (2026-10-18)

[Product brief](docs/PRODUCT_ESSENCE.md): definitions describe meaning, captures
preserve fixed evidence, and investigations explain it. The README now leads with
this loop. Definitions exposes the same measurement revisions as Explore, with
local search; saved record packages also have local search. Service inventory
carries explicit service/environment into metric discovery without automatically
reading telemetry. Measurement stages, canonical YAML, named time scopes and
consistent cards make authoring/execution boundaries visible. Discovery supports
explicit cancellation and suppresses publication after navigation or mode changes.

Passed: 32 frontend checks and all five affected browser journeys across the
initial run and focused rerun: coherence 40, inventory 31, measurements 38, metric
discovery 39 and attribute/facet polish 29. The initial run had three passes and
two harness failures: an inventory selector assumed one button per row, and
Bombadil refused an existing output path. The selector now uses the intended
Explore action's accessible name; discovery campaigns use unique output paths.
The corrected inventory/discovery rerun passed both cases (1.7 minutes).
Original failure artifacts are retained under `artifacts/coherence-initial-failures`.

The new journey passed its delayed **actual source response** abandonment check,
source-free definition preview, service handoff, canonical authoring and mobile
containment. Measurement, inventory and discovery Quint ledgers/negative controls
and all three enabled Bombadil campaigns passed. Discovery exploration includes
abandon/retry and scope-change invariants. Twenty storyboard images were generated
or refreshed, including three new frames for journey 40; desktop/mobile frames
were inspected. No measurement kernel or source query semantics changed.

Evidence: `artifacts/coherence-node.log`, `artifacts/coherence-journeys.log`,
`artifacts/coherence-final-journeys.log`, and `artifacts/metric-discovery/`.
The query/packages image is `11a6f58b5951`. Traffic was restored and the five
pre-existing stale pending entries and telemetry volumes were retained.

## Ordinary OTel metric discovery (2026-10-18)

Bounded service/environment discovery now proposes editable typed packages over
ordinary gauge/sum/explicit-histogram observations. Optional scope version/schema
remain exact identity values, including empty values. No invented identity or
implicit fleet aggregation. Stream summaries expose sorted dimensions.

Passed: root Go suite; race checks for measurements/metrics/HTTP; root vet;
32 frontend checks; Hegel optional-identity properties and existing measurement
state/swarm checks. The independent Python OTel SDK producer has no goo manifest
and omits scope version/schema. The initial live discovery/package/save/run/
capture/reopen storyboard passed (21.1 s). The expanded run passed both affected
journeys (3.1 minutes): measurement storyboard 38 and discovery storyboard 39,
including notes, package saving, capture reopening and native dimension labels.
Bombadil recorded 440 measurement and 248 discovery observations with zero
violations. The shared measurement Quint ledger passed with seven rejected
mutations. Nine affected storyboard images were refreshed.
The final discovery-only rerun passed (22.5 s), including its dedicated Quint
ledger with three rejected mutations: wrong service, missing proposed definition
and tampered evidence basis. Traffic was restored and stale pending data retained.

Artifacts: `artifacts/metric-discovery-{go,race,vet,node,hegel}.log`,
`artifacts/metric-discovery-final-journeys.log`,
`artifacts/metric-discovery-model-journey.log`, and
`artifacts/metric-discovery/ledger.json`.

## Typed measurement packages and standardized examples (2026-10-18)

[Contract](docs/MEASUREMENT_PACKAGES.md): twelve YAML packages bind actual queue
and application telemetry to Go gauges, reset-aware counters, native histograms
and scoped lifecycle preparation. Five data-only ordered workflows preserve a
fixed window and preview without source execution. General captures, immutable
package revisions and ordered notes are shared across all families. Both Go
examples now emit real trace-correlated OTLP logs; Python lifecycle records carry
exact business nanoseconds and stable per-emission event IDs.

Passed native checks: root Go suite, root race suite/vet, instrumentation module,
19 Python tests, 32 frontend checks, portable compiler/lifecycle/package CLI parity
storyboards and Compose configuration. The full verification module passed
uncached (240.18 s), including Hegel state/swarm and existing Quint MBT/trace
checks. Live-ledger tests in that module intentionally skip without supplied
artifacts; the browser journey supplies its actual measurement ledger.
Additional race coverage checks atomic custody publication by 16 concurrent
readers. Binding tests reject a consistently recomputed receipt that falsely
claims complete coverage, not merely an inconsistent receipt.

Focused storyboard 38 passed (1.4 minutes): ordinary reservation scenarios,
Go checkout/report work, independent ClickHouse gauge counts and actual Go logs,
all four measurement families, mature good/retry/missing/rejected/late outcomes,
scoped subject drilldown, fixed question capture/refresh, source-free reopening,
exact notes reference, immutable package save and fresh reuse. Both record and
measurement workflow steps preview without telemetry reads. Quint accepted the
real ledger and rejected seven altered ledgers. Bombadil recorded 462 observations
through initial/preview/invalid/complete stages without invented evidence.

The journey caught an incorrect checkout UI assertion and a real transport bug:
ClickHouse's escaped JSON spellings changed a raw-input evidence hash after a
browser round trip. Source inputs are now canonicalized before hashing, with a
transport regression test. A source metadata mismatch also exposed the Python
meter's missing version/schema; those are now emitted and bound explicitly.
Stored source plans remain reopenable as evidence after compiler changes;
refresh requires an identical current plan.

Broad browser regression: 37 passed, one failed (19.1 minutes), with all
opt-in generated/fault/Bombadil campaigns enabled. The failure was the fault
campaign's six-second actor budget expiring before any browser observation, not
an application or UI invariant failure. The harness now permits browser startup
headroom while preserving its five-second application-response property, and
reports empty traces explicitly instead of a JSON parse error.

Final targeted run on the deployed final image: **2 passed, no skips (4.0
minutes)**. The corrected fault campaign completed all three scenarios and five
inspections/revisits with its model checks (2.5 minutes). Measurement storyboard
38 passed again (1.4 minutes), including valid empty-selector capture
normalization, rejection of unsupported lifecycle scope filters, and a saved
same-name custom package that cannot shadow a builtin workflow's environment.
Thus all 38 browser journeys were verified across the broad run and corrected
focused rerun. All 132 storyboard images (131 PNGs and one JPEG) were refreshed; seven new frames and
architecture flow/sequence diagrams document this slice.

Artifacts: `artifacts/standardization-regression.log`,
`artifacts/standardization-final-journeys.log`,
`artifacts/standardization-race.log`, `artifacts/standardization-verification.log`,
and `artifacts/measurements/360b30f8/`. The first failed fault trace is preserved
under `artifacts/standardization-regression-failure/`; the successful fault run is
`artifacts/bombadil/fault-campaigns/40131596-b9ce-4cb2-9834-9270c7047939/`.
Final scope/normalization/default-resolution checks also have native regressions.
The final measurement Bombadil run recorded 565 observations through all four
required stages; its eight model reports contain one pass and seven intended
conformance rejections.

The broad run used query/packages image
`sha256:7494aa650a6be69c8d06e7d58a1fc88e93149d88b057600bf0beb2da0eea0010`.
The final targeted image, including selector normalization, strict lifecycle scope
validation and builtin workflow resolution, is
`sha256:4df78f6a14a53320181c8f0934f32b73fe1648c22938ab77f215dfb2d6f49b48`.
Runtime frontend hashes match the repository. Telemetry volumes and all five
pre-existing stale Redis pending entries were preserved. The local custody clock
is not warehouse ingestion or a completeness guarantee; missing completion stays
unknown. Databricks YAML was structurally checked, not warehouse-executed.

## Scoped recipes and shared sample investigations (2026-10-18)

[Contract](docs/INVESTIGATION_RECIPES.md): three Go recipe transforms preserve the
applied population and fixed bounds, with source-free preview and explicit
execution. Six ordinary YAML packages move sample investigations onto shared
Explore/package/capture/notes workflows. The specialized lifecycle/resource lab
starts only when opened. Opening a derived recipe question now clears the previous package banner.

Passed native checks: root Go race suite and vet, 19 Python tests and 27 frontend
checks. Native cases cover OR/filter preservation, immutable inputs, parameterized
selectors, wrong-signal rejection, strict HTTP bodies and all six example packages.
Hegel state/swarm passed 100 sequences of 25 recipe/signal/scope/predicate steps
(0.19 s reviewed).

Initial browser harness failures were corrected for the actual save message and
retained disclosure state on same-document navigation. A reserved Quint action
name was renamed. The conformance check then correctly rejected a capture taken
after package reuse changed the time bounds; the journey now verifies the fresh
reuse separately and captures the original fixed question before deriving
a reusable definition from that stored capture. This distinction is part of the documented contract, not a change to
package semantics.

Final browser evidence: the broad 37-journey run completed with 36 passes and
one legacy-sidebar harness timeout (14.8 minutes); that missing disclosure action
was corrected. A final focused run passed both the corrected definition/invalid-
query journey and the strengthened recipe journey (2 passes, 1.2 minutes). The
latter uses ordinary UI filters, captures/cites the fixed question first, then
creates and freshly runs a package from its stored capture. No skips occurred in
the broad run; generated/fault campaigns and cross-app Quint were enabled.

The final recipe ledger passed Quint with six rejecting controls (13.04 s).
Bombadil recorded 467 observations through all four required stages and all three
recipe choices. Wrong-signal rejection, delayed-preview isolation, source-free
notes/preview, literal query preservation and mobile containment were checked.
All 125 storyboard images were refreshed. Five new frames and architecture flow
24 document the slice; new desktop/mobile frames were visually reviewed.
See [environment evidence](ENVIRONMENT_VALIDATION.md) for image and artifact paths.

## Deployment context and ordered notes (2026-10-18)

[Contract](docs/WORKBENCH_CONTEXT.md): authored deployment receipts preserve event
and server custody times; explicit fresh local context accompanies fixed-time
Explore charts without modifying telemetry receipts. Ordered plain-text narrative
cards cite exact general-capture or deployment IDs. Saving creates immutable
parent/root revision branches, preserving older evidence and interpretations.
Dirty/unapplied drafts are retained; no mutable latest head or shared-author claim
is introduced.

Passed: root Go race suite/vet, 19 Python tests and 26 frontend checks. Native
checks cover half-open environment scope, separate custody time, unsafe URLs,
integrity, marker/card/catalog limits, reference validation, pagination and
concurrent branch preservation. Hegel state/swarm passed 60 sequences of 15 steps
(0.38 s reviewed), including Unicode note content, save/fork/order/history,
deployment arrivals and scope/window changes.

The initial browser flow completed its UI assertions but failed verification due
to a test type name colliding with an existing discovery model. Renaming it fixed
the verification build; the captured ledger replay then passed. Journey 36 next
passed with observed evidence/narrative transitions and a seven-stage Bombadil draft prefix
(383 records), including real report-app work, independent ClickHouse trace
discovery, marker boundary/environment exclusions, literal-text safety, ordered
reference editing, immutable revisions/branches, unapplied-draft retention,
source-free capture opening and mobile containment. Quint originally used five
rejecting controls; the strengthened adapter also checks the report's scoped
ready event and rejects a sixth incorrect-subject control (19.97 s replay).
Five new storyboard frames and architecture flow 23 document this slice.

All 36 journeys passed in one broad run (12.6 minutes), with both generated
campaigns and cross-app Quint enabled and no skips. All 120 storyboard images
(119 PNG and one JPEG) were refreshed. Final keyboard/capture-navigation guards
passed three focused journeys (2.0 minutes), including keyboard reorder/remove
focus, cited-evidence focus and delayed-revision isolation. Screenshot inspection
then caught an obsolete loading message after a canceled read; a final status
fix passed journey 36 again on the reviewed image (1.2 minutes), with six rejecting
Quint controls (17.15 s) and 310 Bombadil records through all seven stages.
The final narrow-screen frame was visually inspected. Evidence and image hashes
are in [the environment record](ENVIRONMENT_VALIDATION.md).

## Error and latency comparisons (2026-10-18)

[Contract](docs/ERROR_COMPARISONS.md): exhaustive captured span-status categories,
full-population error-percentage ranking, independently computed mean/p95/p99
changes across fixed windows, retained-baseline group drilldown and explicit
server-reread saved comparison pairs with links to both record captures.
Unset and unrecognized status remain visible; no-data/disagreement suppress
normalized differences. Existing count-only captures/SQL remain compatible.

Passed: root Go race suite/vet; focused capture-refresh race checks; 19 Python
tests; 24 frontend checks; Hegel comparison state/swarm (100 sequences of 25
steps, 0.54 s initial / 0.67 s model regression), plus existing cohort state/swarm
and Quint cohort regression (7.55 s). Native checks cover invalid/missing status
counts, category drift with identical total counts, missing groups, package YAML,
interval validation, source-free saved pair reopening, refresh child lineage and
retaining the predecessor on source outage.

Journey 35 passed initially (54.8 s) and on the polished image (47.0 s). It checks
80 previous/124 current real OTLP spans with only 100 current displayed rows,
2.50% → 16.13% recorded errors (+13.63 pp), +90 ms p95, missing/empty routes,
saved comparison reopening, changed-route-to-trace/log navigation and mobile
containment. The final new-feature campaign passed Quint with four rejecting
controls (7.20 s) and 342 Bombadil records through all six required stages.
All 11 affected regression journeys passed (3.9 minutes). A final trace-action
guard was checked by both discovery and the strengthened comparison journey
(2 passes, 1.3 minutes), including live adapter status-class groups and saved
trace-ID pair behavior. The final comparison campaign recorded 501 Bombadil
records through six stages. Five new frames and affected existing storyboards
were regenerated; see [environment evidence](ENVIRONMENT_VALIDATION.md).

## Grouped duration measures and polish (2026-10-18)

[Contract](docs/GROUPED_MEASURES.md): complete per-group duration aggregates,
normalized semantic results, exact rational ranking, fixed local-search scales,
retained capture pages and reusable package previews. Count-cohort comparison
projects out duration metadata and remains a record-proportion comparison.

Passed: root Go race suite/vet, 19 Python tests, 23 frontend checks and Hegel group
count/sum conservation over 100 generated populations (0.254 s), including uint64
extremes. Native source tests reject missing summaries/duplicate dimensions and
preserve same-count duration disagreement. A live fixture verifies 124 spans versus
100 displayed rows, missing/empty routes, zero duration, p95 ranking, one/two-field
groups, count-cohort projection, package preview and mobile containment.

The initial extended campaign passed five of six journeys, including journey 34
(52.1 s), Quint with three rejecting controls (7.32 s) and 454 Bombadil records
through all nine saved-ranking prefix stages. The discovery regression exposed a
removed accessible trace-button label, now restored. Browser checks also found a
blur-linked control shift, fixed by input-time scope invalidation. Typed model
comparisons preserve exact values without treating JSON indentation as evidence.
The final reviewed-image run passed all 10 selected regression journeys (3.5 minutes), including the original discovery assertion. The final group campaign passed three rejecting Quint controls (6.46 s) and 383 Bombadil records through nine stages. Four new and affected existing frames were refreshed; evidence paths are in [the environment record](ENVIRONMENT_VALIDATION.md).

## Core workbench basics (2026-10-18)

[Contract](docs/CORE_WORKBENCH.md): opt-in full-population span-duration summaries
(count/exact sum/min/max/upper p50/p95/p99), fixed count-over-time intervals with
per-interval durations, immutable capture retention, reusable versioned questions,
span/log service activity and reported parent links. General query admission and
explicit cancellation reach the Go context through the Python proxy. This remains
a local lab; event age is not ingestion lag, and query IDs are not authorization.

Passed: root Go race suite and vet, 19 Python tests, 22 frontend checks, Compose
configuration, Hegel duration state/swarm (100 sequences up to 30 steps). Native
HTTP tests exercise saturated admission, source-free local access, cancellation
propagation and delayed-request tombstones. Source tests reject malformed/missing
summary fields, invalid partitions/bounds and ambiguous dimensions; exact sums
cover values beyond uint64 and JavaScript integer precision.

The first deterministic journey 33 passed in 18.9 s. The strengthened focused
capture/core/inventory run passed all three journeys in 2.3 minutes, including
187 Bombadil records through all six core prefix stages and actual-ledger Quint
validation (13.26 s) with four rejecting controls. Five controls now check plausible-but-wrong p95/sum, ambiguous-link classification, cancellation and reused environment scope. All 33 journeys passed in one comprehensive run (11.9 minutes), with both generated campaigns enabled and no skips. All 106 storyboard images were refreshed. The expanded final campaign passed again with both Bombadil actors (216 saved-measure records plus 181 live-workflow records), all required stages and five strengthened rejecting controls. Artifact paths are in [the environment record](ENVIRONMENT_VALIDATION.md). Architecture flow 20 and four
new screenshot storyboards document this increment.

## General evidence captures and stable saved paging (2026-10-18)

[General captures](docs/GENERAL_CAPTURES.md) now persist manifest-free span/log
questions and service inventory in the existing investigations volume. Reopen and
paging read verified local evidence; refresh creates a successor with the same
normalized fixed question and compiled source plan. Explore captures retain up to
1,000 records, shown in pages of 100, while full supported aggregates retain their
own counts and source states. These bounded records do not claim source completeness.

Passed: root Go race suite and vet; focused race checks after final normalization;
19 Python tests and 21 Node checks. Native HTTP tests reopen with the source closed,
reject client-supplied evidence and cross-revision cursors, and preserve predecessors
on failed refresh. Hegel state/swarm passed 60 sequences of up to 15 steps against
final code (12.939 s), including arrivals, source resets, captures and branching
refreshes. Quint validates actual browser capture/page/refresh/reopen observations,
with four rejecting controls for old-count, page-basis, question and cursor drift.

Journey 32 covers 1,205 matching spans, a 1,000-record retained bank, stable pages
after late arrival, a 1,206-span successor, source-free browser reopening after
process restart, explicit live handoff, inventory, logs, narrow screens and stale
response isolation. Six new storyboard frames accompany diagram 19. Expanded trace
and log context remains a separate fresh read; saved baseline pairs, team access
control and collaborative editing remain future work.

The initial default suite had 29 passes, two opt-in skips and one outdated wording
assertion in journey 22. That assertion now matches the general “saved captures”
disclaimer; its semantic checks remain and focused journeys 19 and 22 passed.
Final capture/Bombadil and both opt-in campaigns passed (three tests, 3.5 minutes); all 32 distinct journeys passed across these runs. The capture actor recorded 367 trace records through all eight stages. All 102 storyboard images were refreshed; detailed evidence paths are in the cloud record.
An initial final-test selector matched no tests, and one retry lacked workspace Go
cache variables; both invocation issues were corrected before final verification.

The agent's [collaboration research](docs/COLLABORATIVE_INVESTIGATIONS_RESEARCH.md)
distinguishes verified Honeycomb Boards and Datadog multiplayer Notebooks from an
unverified Honeycomb whiteboard identification. It proposes ordered investigation
cards before authorized collaboration/presence; no multiplayer feature is claimed.

## Observed service inventory

Implemented: single-read full-population service/span-name counts, OTel Error
statuses and fixed span-duration buckets; local filtering preserves captured
accounting; drilldown opens a separate fixed-window Explore read and package
preview. [Contract](docs/SERVICE_INVENTORY.md), [diagram 18](docs/ARCHITECTURE_DIAGRAMS.md#18-discover-services-and-promote-an-investigation-question)
and [journey 31](docs/journeys/README.md#31-from-observed-services-to-a-reusable-question)
record semantics and executable frames. Unknown fleet membership/health and
span-versus-request grain remain explicit.

Passed: root Go race suite/vet, 19 Python tests, 20 Node checks, Compose config,
Hegel observe/reset state/swarm (100 sequences, up to 20 steps), and actual-browser
Quint capture/search/open/package conformance with three rejecting controls.
Source tests retain counts above JavaScript integer precision and reject invalid
partitions, errors exceeding counts, missing/null dimensions, duplicate groups,
cardinality overflow and unavailable-to-zero conversion. Query values remain
parameters; HTTP rejects unknown fields and multiple objects.

The live 116-span fixture checks independent ClickHouse counts, 113 matching
spans versus 100 displayed rows, package preview, local filter stability, no data
and pending-read invalidation. The dedicated read-only Bombadil campaign passed
609 records across all six prefix stages plus refresh/search/scope edits. Evidence:
`artifacts/inventory/inventory-a777c2549e/`. Narrow-screen screenshot review exposed
compressed table headings; stacked mobile cards and a width assertion corrected
that issue, and the updated journey passed. See the cloud validation record for
the subsequent broader browser run and final-image verification.

## Manifest-free ad hoc telemetry packages

Implemented: captured Explore question → named telemetry package YAML → shared Go
query plan → separate live preview → immutable local configuration revision → fresh
Explore reuse. Import/export, nested predicates/two-field breakdown preservation,
local refinements and optional downloadable Go annotation templates are included.
No target application manifest is loaded. See [contract](docs/TELEMETRY_PACKAGES.md),
[diagram 12](docs/ARCHITECTURE_DIAGRAMS.md#12-ad-hoc-telemetry-packages-and-optional-instrumentation)
and [journeys 24–25](docs/journeys/README.md#24-build-a-package-over-existing-python-telemetry).

Passed: root Go race suite/vet, 19 Python tests, 14 Node checks and Compose config.
New Go fixtures check strict bounded YAML, unknown fields/aliases/multiple documents,
unsupported grains/adapters/roles, time/page separation, stable YAML round trips,
concurrent idempotent publication, immutable old revisions, path validation and
corruption detection. HTTP tests pass generated Go code through the actual source
annotation scanner. No generated template is patched or deployed into an app.

Hegel state/swarm passes 100 sequences of up to 30 edit/save/window steps; Quint
observes draft/save/edit/reopen/apply and rejects mutated-original and nonreusable
negative controls. Focused verification passed (5.1 seconds); the complete nested
verification suite passed (95.5 seconds), with ignored log
`artifacts/verification/telemetry-packages-full-suite.log`.

The first actual Python SDK authoring journey passed (9.4 seconds). After preview
and handoff presentation improvements, journeys 24/25 passed together (14.0 seconds):
ordinary SDK spans without goo attributes/manifests, collector ingestion, exact
preview counts, save/export/import, service-restart persistence, fresh reuse,
local refinements, preserved nested log/two-field selectors, no-data preview and
stale-validation isolation. The Node-style log workload is a labelled OTLP fixture,
not a claim that a Node SDK service ran. The generated Go template is scanner-tested,
not source-equivalence, weaving, deployment or coverage proof.

VFS build history exhausted workspace bytes during recreation; unused dangling
images/build cache were reclaimed, retaining running containers, all telemetry
volumes and existing stale pending entries. Recreation then succeeded. No daemon,
proxy, credentials or storage-driver settings were changed.

This package kind supports captured-record counts and existing views, not arbitrary
measures, inferred SLOs/business semantics, automatic source patches, source authenticity,
production collaboration or Databricks lowering. Saved definitions do not archive
preview evidence. Mermaid source/fences and local links are checked without claiming
rendered Mermaid validation.

All 25 real browser journeys passed together (4.1 minutes), including the new
Python SDK and imported-log paths, native metric/fleet regressions, saved receipts
and baseline comparison. All 74 documentation frames were regenerated. Preview
tables, readable handoff code and mobile authoring were visually inspected. Final
editor-focus and metadata/navigation generation guards are covered by a focused
rerun of the two affected authoring journeys, which passed together (16.7 seconds).
Their six new frames were regenerated again after those refinements. All 12 services
are running, traffic is restored, and the five pre-existing stale pending stream
entries and all data volumes remain intact.

## Cohesive general-purpose workspace and cohort distributions

Implemented locally: Explore/Operations/Saved investigations/Definitions navigation,
secondary reservation Examples, shared responsive styles, result tabs and trace/log
inspector. Complete captured distributions support a browser-held fixed-window
baseline and Go rational percentage-point comparison. The shared waterfall accepts
caller-provided related actions; reservation navigation is supplied only by Examples.
See [workspace and scope contract](docs/PRODUCT_WORKSPACE.md),
[diagram 11](docs/ARCHITECTURE_DIAGRAMS.md#11-workspace-navigation-and-captured-distribution-comparison)
and [journey 23](docs/journeys/README.md#23-one-investigation-workspace-from-distribution-to-trace).

Passed: root Go race tests/vet, 19 Python tests, 14 Node checks, strict comparison
HTTP validation and Compose configuration. Fixtures cover exact uint64 counts,
missing/empty identities, category unions, scope drift, duplicate/partial groups,
empty populations and changed-read unknowns. Hegel cohort state/swarm runs 100
sequences of up to 30 steps. Quint validates observed pin/selection/empty/changed/
scope-change results with three rejecting negative controls. Full nested verification
suite passed (91.9 seconds); ignored log: `artifacts/verification/workspace-full-suite.log`.

The initial targeted seven browser journeys passed (37.6 seconds), covering actual
workspace navigation, old and new query paths, saved/live expansion, and a labelled
12-record collector fixture compared with a four-record filtered population. The
source fixture is documentation telemetry, not a production error attribution.
Mobile screenshot review prompted labelled comparison rows and visible navigation.
The full run then passed 22/23 journeys (4.8 minutes). Journey 07 exposed a
pre-existing readiness assumption: its five-minute API poll could see an older
failure while the UI used the new two-minute default. Its fixture now aligns
the window and waits for a newly captured failure identity. The first focused
rerun also encountered an actual transient source error during baseline acquisition;
readiness now checks HTTP success before accessing facts. Source failures are
not converted into zero-valued evidence. Expanded journey 23 passed again (6.7
seconds), including distribution-value refinement in the baseline's fixed window.
The corrected journey 07 then passed (4.0 seconds; focused command 6.3 seconds).
Together with the full run and focused journey 23, all 23 journeys have passing
results against this round's runtime. All 68 documentation frames were regenerated;
desktop comparison/definitions and mobile comparison/inspector frames were visually
reviewed. The inspector frame uses the actual viewport, avoiding full-page fixed-overlay
screenshot distortion. All 12 services are running and traffic is restored. Existing
five stale pending entries and all volumes were retained.
Cohort observations were strengthened to check expected positive/empty/unknown
results explicitly; focused Hegel/Quint verification passed again (4.7 seconds).
Baselines remain volatile browser state; caller-held comparison evidence is not
source-authenticated. Optional HyperDX, a Databricks instance, production deployment
and a complete accessibility audit were not validated. Mermaid source/fences and
relative documentation links were checked; no rendered-Mermaid validation is claimed.

## Bounded general exploration and expanded trace/log context

Implemented: shared Go query model/reference predicates and parameterized
ClickHouse compiler; UI/API/CLI source planning; discovered attribute keys,
string and UInt64-duration predicates with bounded AND/OR; independent captured
record counts/groups and bounded pages; missing-vs-empty grouping; detected
cross-read disagreement; cross-service trace pages and exact trace/span log scopes.
Operation/saved waterfalls explicitly expand into separate live evidence. See
[design](docs/EXPLORATION_DESIGN.md),
[diagram 10](docs/ARCHITECTURE_DIAGRAMS.md#10-general-exploration-and-independent-expanded-evidence)
and [journeys 20–22](docs/journeys/README.md#20-general-exploration-and-population-safe-paging).

Passed: root Go race tests/vet, 19 Python checks and 12 Node checks. Core/HTTP
fixtures cover type/operator/budget validation, literal parameter safety, exact
UInt64 comparisons, missing values, page-independent count/group SQL, fixed
trace scope, source/group overflow failures, detected read disagreement, CLI
planning and invalid requests rejected before source execution. Hegel state/swarm
passes 100 sequences of up to 30 steps; Quint observed search/page/trace/log steps
pass with three negative population/scope/identity controls. All three Dockerfile first-stage COPY contexts compile the root CLI with
VCS stamping disabled for temporary contexts. Full nested
verification suite passed (89.4 seconds), log in ignored
`artifacts/verification/exploration-full-suite.log`.

Live targeted journeys 20/21 passed together (13.8 seconds); expanded 20–22
passed together (37.6 seconds), including trace paging with missing parents and
saved/live evidence separation. An initial journey read collapsed diagnostics
using rendered text and failed; it now reads the actual stored text content.
Direct read-only ClickHouse checks found that raw HTTP query parameters reinterpret
backslashes and reject literal tabs. The adapter now escapes ClickHouse parameter
text as well as URL encoding. Exact backslash/tab/newline/Unicode/DEL round trips
were verified, and journey 20 now includes literal transport characters in both
keys and values. This is a transport correction, not an SQL interpolation path.

All 22 live browser journeys passed together (4.2 minutes), including the literal
transport regression in the 123-record fixture. The 63 documentation frames were
regenerated. After a UI origin-metadata refinement, focused journeys 20–22 passed again
(22.7 seconds), including the assertion that log→trace navigation carries the log
response question/basis as its origin. Context diagnostics also retain source
state/warnings; actual read disagreement is covered by source fixtures. No Go
query kernels changed after the full browser pass. The six new documentation
frames were regenerated, and the exploration, distributed waterfall/log and
saved/live-context frames were visually inspected. All 12 services are running;
the owned traffic generator is restored and no queues/volumes were cleared.
Mermaid source and relative documentation links were checked without claiming a
local Mermaid render.

General query/cursor persistence, whole-trace completeness/accumulation, live tail,
value distributions, cohort comparison, production authorization and warehouse
lowering remain unimplemented. No production deployment, Databricks instance or
full networked Docker build is claimed. General context is not added to saved
operation receipts.

## Saved fixed-window investigations

Implemented: bounded server-captured immutable operation evidence; local named
volume; package revision/fixed-window successor refresh; reopen without telemetry
queries; stored operation/filter/trace/subject and ratio/member explanation state;
component error preservation and comparison of independent evidence bases.
See [contract](docs/SAVED_INVESTIGATIONS.md),
[architecture diagram 9](docs/ARCHITECTURE_DIAGRAMS.md#9-captured-investigations-and-explicit-refresh)
and [journey 19](docs/journeys/README.md#19-save-reopen-refresh-and-compare).
The capture is an explicit fresh read when saving, not a silent claim to freeze
the prior UI response. It archives normalized results, not all raw metric inputs
or renderer/evaluator binaries. No production collaboration is claimed.

Passed: root Go race tests/vet, 19 Python checks and 11 Node checks. HTTP fixtures
verify trusted source capture, optional metric failures, offline reopening,
package/question conflicts, invalid selections and original survival after a
failed refresh. Store tests cover process-independent reload, exact JSON integer
preservation, concurrent publication, historical retention and corruption rejection.
Hegel state/swarm passes 100 sequences of up to 25 steps; Quint validates five
observed save/reopen/refresh/reject/reopen transitions and rejects two controls
mutating original integrity or the fixed question. Complete nested verification
suite passes (75.3 seconds); log in ignored
`artifacts/verification/investigations-full-suite.log`. An initial model adapter
parse failure with a parameterless module was corrected by declaring its initial
state constant; the final observed checks pass.

Root and both sample Dockerfile first-stage COPY contexts compile the root CLI
with VCS stamping disabled for temporary sandbox contexts; sample manifests pass.
Full networked multistage image builds are still unverified.

Targeted live journey 19 passed twice (21.2 and 24.5 seconds), including actual
packages-service restart, unchanged receipt reopening, real collector ingestion
of a labelled late test span, and successor retention of the fixed question and
trace selection. The second pass also verifies ratio/member explanation state.
All 19 browser journeys passed in two full runs (3.9 minutes each). The final
run follows the heatmap selection-state correction and includes saved cell-bound
restoration, ratio/member explanations, real late arrival, restart persistence
and a delayed real saved response discarded after a newer package selection.
All 57 documentation frames were regenerated; the three new saved-investigation
frames were visually inspected. A final test-only refinement scopes saved ratio
and comparison explanations to checkout (rather than whichever operation appears
first). Focused journey 19 then passed (16.2 seconds), regenerated its three
frames and asserted both selected stream identities. No application code changed
after the final full 19-journey pass. All 12 default services are running, metric
export is enabled, and the owned traffic generator was restored. No queues or
volumes were cleared. Documentation source links resolve; Mermaid is documented
without a claim of local rendering. General observability gaps and sequencing are recorded in
[PRODUCT_GAPS.md](docs/PRODUCT_GAPS.md), with reviewed public research sources and
explicit limits.

## Live peer and bounded-skew comparisons

Implemented: two actual woven checkout instances with distinct resource identity
(default ports 8090/8092); optional explicit bounded-skew policy; actual per-member
window/count preservation; pairwise endpoint spreads; deterministic conservative
matching without reusing intervals in successful rows. Comparison rows have no
combined percentage; exact fleet totals remain strict. The lab-only export gate
can drop metrics while requests and the independent trace exporter remain enabled.
See [semantics](docs/FLEET_COMPARISONS.md),
[diagram 8](docs/ARCHITECTURE_DIAGRAMS.md#8-two-live-instances-and-bounded-comparisons)
and [journey 17](docs/journeys/README.md#17-two-live-instances-imbalance-loss-and-restart).

Passed: root Go race tests/vet, 19 Python checks, ten Node checks; Hegel comparison
state/swarm: 100 sequences of up to 30 steps; six observed Quint outcomes and three
negative controls rejecting invented common populations. Targeted live journey 17
passed through imbalance, successful requests during metric loss, excessive gap,
and controlled peer restart/new epoch. Four screenshots were generated. The first
attempt exposed a real startup-query/control race (a late 30-minute response
replaced the edited two-minute lookback); controls now preserve edits and stale query
responses are generation-guarded. Source limits were retained. Targeted journeys 17/18 passed together (41 seconds); journey 18 guards
that race using a delayed real response. The full nested verification module passed (60 seconds).
All 18 browser journeys passed in one full run (3.2 minutes), refreshing 54
documentation screenshots. New member-window and loss frames were visually
inspected; updated documentation file links resolve. All twelve services are
running and the peer metric-export gate is enabled. Mermaid source is documented
without claiming a local Mermaid render or production deployment validation.
A focused post-run storyboard check scoped same-timestamp rows to checkout
and asserted the displayed reset cause. Gap/restart checks also require timestamps
from the current injected fault, avoiding matches against older retained telemetry.
The final focused journey17 rerun passed (36.7 seconds), including fresh-fault
timestamp guards and operation-scoped screenshot/assertion selection. No
application code changed after the successful full18-journey run.

## Historical declared fleet membership milestone


Implemented: fixed-window YAML membership for 2–4 instances; bounded concurrent
member queries; separate observed/unobserved/unavailable source states; exact
aligned population totals only when every member supports an interval. Missing,
reset, skewed or incompatible members leave totals unknown. Valid selected-instance
rates survive peer source failure. The default lab intentionally declares an
undeployed peer, not a live second replica. See [semantics](docs/FLEET_COVERAGE.md),
[diagram 7](docs/ARCHITECTURE_DIAGRAMS.md#7-fleet-membership-and-coverage-before-totals)
and [journey 16](docs/journeys/README.md#16-declared-fleet-membership-and-missing-coverage).

Passed: root Go race tests/vet, 19 Python checks, nine Node checks; Hegel fleet
state/swarm: 100 sequences up to 30 steps; Quint six observed aligned/missing/
skew/reset/idle/recovered steps and four rejected invented totals. Race-tested HTTP
fixture verifies peer failure preserves selected-instance rates. Targeted live
journey 16 passed with three generated frames. Dynamic membership, automatic
alignment and live multi-replica aggregation remain unvalidated/unimplemented.
The complete nested verification module also passed (70.8 seconds).
All 16 live browser journeys passed in one full run (2.8 minutes), refreshing
49 documentation screenshots. Fleet coverage frames were visually inspected and
updated documentation file links resolve. The default lab retains eleven services;
no second replica or automatic alignment is claimed.

## Operation error ratios and guided contextual investigation

Implemented: strict YAML ratio policy references an exhaustive calls-rate contract;
exact aligned outcome deltas produce calls/errors/fraction and a display percentage.
Missing/skewed/unsupported intervals remain unknown, supported zero calls are no
data, and full resource identity prevents implicit instance/operation merging.
The runtime explicitly observes zero outcome counters on first operation use.
The UI shows interval charts and guides error-context selection into captured
spans and the existing scoped trace waterfall. It does not detect increases or
causes automatically. See [semantics](docs/OPERATION_ERROR_RATIOS.md),
[diagram 6](docs/ARCHITECTURE_DIAGRAMS.md#6-from-aligned-counters-to-an-error-investigation)
and [journey 15](docs/journeys/README.md#15-from-operation-error-ratios-to-traces).

Passed: root Go race tests/vet, 19 Python checks and eight Node checks. Go fixtures
cover exact sums exceeding uint64, unknown/idle/alignment, invalid members/policies,
bases and a mixed idle/traffic window. Hegel ratio state/swarm checks pass: 100 sequences of up to 30 steps.
Runtime outcome-baseline state/swarm checks also pass 100 sequences. Quint validates
six observed traffic/missing/idle/reset/gap/late-repair steps and rejects four
invented ratios. Targeted live journey 15 passed, generating four readable frames.
Its first attempt correctly found no contextual spans before span ingestion;
the journey now waits for both kinds of real evidence before demonstrating traces.
Counter source cap increased to 3,000 raw points per contract for initialized outcome
streams; 8 MiB and Float64 precision guards remain, histogram caps unchanged.
The full nested verification module also passed (47.8 seconds). All 15 browser
journeys passed in one full run (2.0 minutes), refreshing 46 documentation frames.
Updated documentation file links resolve; ratio and waterfall frames were visually
inspected. All eleven default services, including traffic, are running. Mermaid source/link checks
are performed without claiming a local Mermaid render or production validation.

## Native histogram intervals and synchronized documentation

Implemented: exact uint64 cumulative/delta histogram evaluation, explicit
unit/layout/temporality/resource contracts, unknown reset/decrease/gap/flag/
boundary intervals, late repair and explicit aligned merging. Live checkout
histograms now feed a per-stream bucket heatmap with bucket-bounded P95 and
context-only span navigation. Automatic fleet discovery is not implemented.
See [semantics](docs/NATIVE_HISTOGRAMS.md),
[updated diagram 5](docs/ARCHITECTURE_DIAGRAMS.md#5-native-histogram-intervals-and-bounded-investigation)
and [journey 14](docs/journeys/README.md#14-native-duration-histograms).

Passed: root Go race tests/vet; 19 Python checks; seven Node visualization checks;
CLI fixture count4 and explicit merged count8; Hegel histogram state/swarm
100 sequences and merge properties150 cases; Quint observed restart/gap/repair/
merge validation with reset and incompatible-merge negative controls. All 14 live browser journeys passed in one full run (2.0 minutes), refreshing
42 documentation frames, including three cropped histogram frames. Targeted
journey14 also passed separately. File links resolve; Mermaid source was inspected,
not locally rendered. Documentation, diagrams, storyboard and generated frames
are required alongside feature changes in AGENTS.md.

## Native counter semantics, live rates and architecture diagrams

Implemented: pure Go cumulative/delta monotonic counter evaluation with exact
uint64 deltas, explicit stream identity, reset/epoch/decrease handling, maximum
gaps, missing flags, clipped unknown boundaries, duplicate/conflict checks and
canonical late-input revisions. CLI evaluate-counters runs without ClickHouse.
YAML rates and counter-rates views bind supported operation calls/error outcomes.
The checkout app now exports actual OTel native metrics every two seconds through
the collector; a bounded sum-table adapter validates resource/schema/unit/flags/
temporality and rejects Float64 precision risk. Charts show per-stream supported
intervals and separate counter bases. Linked spans are contextual start-time
matches, not exact contributors or exemplars.

Passed: root Go race tests/vet, exact CLI fixture deltas 3/4, 19 Python checks and
six Node checks. New counter/source fixtures cover uint64 precision, zeros,
resets/decreases, missing points, delta overlaps, conflicting copies/temporality,
identity, edges, late repair and rejected exporter precision loss. Hegel passes
100 state/swarm sequences (up to 30 steps). Quint validates five observed evaluator
steps through restart/gap/late repair, and rejects a rate invented across restart.
Artifacts are in artifacts/verification/counter-*.

The first full browser run had 12 passing journeys and one failure: an old broad
summary selector became ambiguous after adding the counter panel. That selector
was scoped to its intended operation explanation label. Affected journeys 08/11
and native-rate journey 13 then all passed in a targeted rerun (17 seconds).
Thus all thirteen journeys are verified across the two runs; this is not a claim
of a subsequent full-suite rerun. Two new counter frames bring the total to 39.
Eleven services run, traffic is restored and four stale pending entries remain.
Environment restoration initially left ClickHouse/Redis stopped; their existing
containers were started without resetting volumes.

The user-requested subagent prepared four Mermaid architecture/process diagrams
covering build/planning, illustrative production OTel/ClickStack topology, captured
evidence investigation and the separate Databricks compiler boundary. Links and
fences were checked; Mermaid syntax was inspected but not locally rendered.
These diagrams are documentation, not production deployment evidence.

Running images use offline-wheel builds plus host-built static binaries. Cumulative
metrics are verified live; delta semantics are fixture-tested. Native histograms,
fleet aggregation, error ratios, metric exemplars, native metric Databricks lowering,
immutable custody and production automation remain outside this slice. Updated
startup instructions are saved as a draft, not published. See
[native counters](docs/NATIVE_COUNTERS.md) and
[architecture diagrams](docs/ARCHITECTURE_DIAGRAMS.md).

## Declarative YAML operation packages and second application

Implemented: strict versioned YAML operation packages with source adapter/resource/
instrumentation bindings, exact source operation/event contracts, full subject
identity, enabled built-in views and explicit unverified evidence assumptions.
`goo plan-operations` and the investigation service share the same planner. Plans
carry content revisions that contribute to captured evidence bases; mutation is
rejected before fetching. Repeatable operation-package flags replace the former
source-directory flag and sample-specific Go binding. The UI selects packages,
honors enabled views, clears previous evidence and discards stale responses.

A separate woven Go report generator now uses its own service/instance, source
manifest, report.generate operation and report namespace. Its YAML omits the
heatmap and reuses the same query/UI code. Both packages run in default Compose;
there is no application-name branch in the operation renderer or source adapter.

Passed: all twelve live Playwright journeys (1.2 minutes), including report
creation → YAML-backed view → report subject and deliberate delayed delivery of
a real checkout response during a switch to reports. No telemetry response data
was mocked. Five new documentation frames bring the total to 37. Eleven services
run, demo traffic was restored and three stale pending entries retained.

Root Go race tests/vet, 19 Python checks and six Node checks pass. CLI tests plan
both checked-in packages. Source/plan tests cover stale roles/versions, weakened
identity, unknown views/fields, duplicate keys, aliases and multi-doc YAML, revision
determinism/nonmutation, resource provenance and changed-plan rejection. Hegel
passes 150 new planner cases for typed resource values, revision sensitivity and
nonmutation. Existing 100-case heatmap and 100-case subject properties plus the
five-action Quint navigation checks/negative controls remain passing.

The report helper actually validates and builds a static woven executable. Running
images use the established offline build with host-built Go binaries; the new
networked report Dockerfile was not executed here. Install/start additions are
saved in the cloud environment draft, not published. No root dependency was added.

This is a local-source, fixed-adapter/view YAML package slice. General package
resolution, source-free manifests, arbitrary expressions/CEL/KCL, IaC apply,
native metric reset/histogram policy and operation Databricks lowering remain
unimplemented. See [package specification and executable onboarding](docs/OPERATION_PACKAGES.md).

## Scoped subject event navigation

Implemented: optional runtime subject namespace/tenant context, declared-event
normalization from actual ClickHouse nested event arrays, and trace → subject
events → related event-bearing invocations across traces in the captured evidence.
The key retains environment/service/namespace/tenant/ID; missing or mismatched
identity stays unlinked with a diagnostic. The sample now generates fresh default
IDs and can explicitly reuse an ID in either demo tenant, fixing its former
constant-subject collision. No subject dimensions were added to metrics.

Passed: all ten real browser journeys (1.2 minutes). The new live journey submits
two invocations with one scoped subject, plus an identical ID under a second
tenant. It confirms four events across two related traces, excludes the other
tenant, retains the original basis and performs no additional operation request.
Three new frames bring the storyboard total to 32. All ten services run; demo
traffic is restored and three pre-existing stale pending entries were retained.

Go race tests/vet, 19 Python checks and six Node checks pass. The actual OTel SDK
scope test passes, as do 100 new Hegel scope-isolation cases, the existing 100
heatmap cases and 200 observation-copy cases. The five-action Quint kernel fixture
passes conformance/safety; both wrong-cell and cross-tenant membership mutations
are rejected. Observed steps and ITF are retained under artifacts/verification/
subject-navigation*. Existing actual woven SDK/HTTP preparation checks and Quint
negative control remain passing. Root runtime dependencies remain unchanged.

Returned event arrays/timestamps are validated against their parent spans. The
relation joins only event-bearing invocations, not failed calls inferred from
absence. Reported terminal outcomes do not establish retries, transaction commit,
durability or complete subject lifecycles. Query scope is still one explicit local
resource/library/span-start window with mutable late-data reruns. Updated startup
instructions are saved as a draft, not published. See [subject storyboard](docs/SUBJECT_NAVIGATION.md).

## Operation duration heatmaps and trace drilldown

Implemented: the Go source response retains deduplicated logical span evidence
with exact string start/duration nanoseconds and validated parent IDs. Operation,
calls, error and panic selections feed a raw-duration heatmap and span inspector.
Cells inspect exact contributors; opening a trace expands to the other operation
spans in that trace within the same resource/window/basis. The existing waterfall
renderer is reused with a separate details target. The Go sample now includes a
nested stock-check operation with bounded simulated success/error work.

Passed: nine real Playwright journeys (about one minute), including live failed
checkout → reserve error selection → duration cell → contributor → parent/child
waterfall. The new journey independently checks membership and contributor count,
changes the lookback control during exploration, and verifies a retained basis,
original bounds and no additional operation request. Three new frames bring the
storyboard total to 29. Existing flows remain passing; traffic was restored and
three stale pending entries preserved.

Go race tests/vet, 19 Python checks and five Node checks pass. Hegel passes 100
new generated cases exercising the actual browser kernel against a Go reference,
plus the existing 200-case operation-copy property. A four-action observed
navigation fixture passes quintgo/Quint conformance and safety; an incorrect cell
membership mutation is rejected. This models the actual selection kernel over a
bounded fixture, not browser DOM behavior or source completeness. Reports are in
artifacts/verification/operation-navigation*.txt. Source tests additionally check
logical evidence precision, parent validation and deduplicated inspector records.

The updated Go binaries and UI run in the established offline-built images.
Trace scope stays limited to observed operation spans in one selected resource,
library and span-start window. Sampling/completeness remain unknown, late-data
refreshes can change results, and no immutable custody or full distributed trace
fetch is claimed. See [the executable storyboard](docs/OPERATION_DRILLDOWN.md).

## Live woven Go operation integration

Implemented: separate Go checkout sample built with actual Orchestrion weaving,
standard OTLP/HTTP trace export through the collector into ClickHouse, explicit
resource/library binding, and a Go operation endpoint behind the existing query
proxy. The investigation panel lists declared operations, evidence status,
completed invocations, errors, panics and duration sample counts, with source,
window, manifest revision, evidence basis, typed SQL and duplicate explanation.
Unobserved declarations display unknown counts.

Passed: all eight real Playwright journeys (53.8 seconds), including actual sample
HTTP 200/503/500 paths and observed compatible operation rows. Three new captured
frames bring the storyboard total to 26. Ten default Compose services run; traffic
was restored and three pre-existing stale pending entries retained. Go race tests
and vet pass; 19 Python checks and three Node geometry tests pass. Hegel passes
200 generated cases checking invocation totals under duplicate copies and reversed
input order. Source-adapter tests cover typed values, exact bounds, resource/kind/
scope rejection, conflicting copies, deduplication, raw overflow and HTTP decoding.
The lint-before-build sample helper was executed successfully. Existing actual
woven SDK conformance and HTTP preparation/Quint replay plus negative control pass.

The running images use offline Python wheels and host-built static binaries.
The new networked sample Dockerfile was not executed. A read-only buildx cache
path was corrected with BUILDX_CONFIG=/workspace/.cache/buildx. A Compose name
conflict during replacement was recovered by recreating only query/packages and
the new sample; ClickHouse/Redis volumes were preserved. Updated reusable install
and start instructions were saved to the environment draft, not published.

Schema compatibility is declared because the exporter drops scope schema URLs.
This sample exports traces; native metric export/rate/reset/histogram semantics
remain separate. Coverage, sampling, completeness, immutable custody and durable
business subject semantics remain unknown. See [live operation storyboard](docs/LIVE_OPERATIONS.md).

## Observed operation conformance

Implemented source-manifest matching against normalized completed operation
spans, an invocation overview and `goo observe-operations`. Root race tests pass.
The actual woven SDK test also passes and checks three completed spans against
the overview (two reserve calls including one error, one panic). It exports
artifacts/annotations/operations.json for the executable CLI journey. Negative
checks reject cross-resource evidence, mismatched scope/version/schema,
undeclared roles, invalid outcomes/IDs and conflicting duplicates; identical
copies count once. Exact duration precision and unobserved declarations are tested.
Existing woven HTTP/Quint replay and negative control still pass.

This is SDK evidence, not collector/ClickHouse/browser integration. Native metric
rates and histogram merging, live operation discovery and business aggregation
remain unimplemented. See [the operation journey](docs/OPERATION_CONFORMANCE.md).

## Goo operation/event annotation slice

Implemented: bounded operation/event directives, deterministic source-contract
manifests, isolated OTel runtime hooks and test-only lint-before-weave CLI.
The normal runtime compiler module has no new OTel dependency. The default
Dockerfile now includes the annotation parser in its Go build stage; no running
service or UI was changed by this opt-in developer slice.

Passed: root race tests/vet; actual Orchestrion weaving of the example with real
OTel SDK export. Checks cover success/error/panic application behavior, three
completed-invocation/duration observations, zero leaked active invocations,
two subject events, three diagnosed event drops, and manifest scope/version/unit
agreement. The existing woven HTTP preparer/Quint test and dropped-fact negative
control remain passing. Telemetry is retained in artifacts/annotations/otel.json;
manifest in artifacts/annotation-manifest.json.

The native Hegel state/swarm check passes with 100 requested sequences of up to
30 steps, checking actual SDK active counts after start, successful/error/panic
completion and repeated finalization. The developer journey uses in-memory OTel
export, not live collector ingestion or a new browser screen. Full build manifests,
automatic inventory activation, operation aggregate lowering and production
weaving deployment remain next steps. See [the executable annotation storyboard](docs/GOO_ANNOTATIONS.md).

## Prior live HTTP source, shared view and opt-in weaving

Passed in the current workspace:

- Go race tests/vet and 19 Python checks, plus the shared package CLI/reference/
  embedded ClickHouse parity storyboard and three Node geometry checks.
- Seven real browser journeys (47.8 seconds), capturing 23 frames. The new
  journey creates actual successful and failing HTTP responses, waits for OTel
  ingestion, renders the Go view descriptor and inspects policy/source/revision.
- Source adapter unit checks for typed parameters, exact half-open bounds,
  valid OTel IDs, deduplication, conflicting copies and partial-scope rejection.
- Hegel: 600 requested cases across discovery, HTTP policy and source dedup.
- Real Orchestrion weaving of `Expansion.PrepareHTTP` into quintgo recording;
  its retained-count trace passes Quint and a deliberate dropped-fact trace is
  rejected. Positive and negative artifacts are kept separately.

The isolated weaving module pins x/tools 0.52.0: the initially selected 0.49.0
could not read Go 1.27 export data. The updated tooling passes; the root runtime
module is unaffected. The initial module download denial was resolved through
the official GitHub source with Go checksum verification. No network policy was
changed to complete the check.

The actual running images use the established offline-wheel build plus a
host-built static Go executable. The new networked multi-stage Dockerfile was
not executed here. Default Compose now includes the Go packages service; all
nine services run and traffic is restored after journeys. Reusable setup draft
instructions were saved, not published.

Schema compatibility is declared by library binding because this exporter drops
scope schema URLs. Sampling/completeness, immutable custody snapshots, generic
instrumentation auto-discovery, prepared-table materialization and Databricks
execution remain unverified/unimplemented. Woven model observations are tested
in the optional module, not enabled in the running service. See
[LIVE_HTTP_PACKAGE.md](docs/LIVE_HTTP_PACKAGE.md) and
[INSTRUMENTATION.md](docs/INSTRUMENTATION.md).

## Prior HTTP binding and expansion slice

Implemented: explicit HTTP server-span binding into the existing shared Go
measurement model, aggregate view/recipe descriptors, both backend plans and
normalized server-request preparation. Content revisions cover policy and emitted
artifacts. No current UI or raw-exporter adapter integration is claimed.

Passed: root Go race tests and vet; package CLI storyboard with two independently
specified route groups matching both reference evaluation and generated embedded
ClickHouse SQL; generated Databricks YAML structural checks. Unit checks reject
unsupported profiles, missing policy/source contracts, resource mismatches,
client spans, duplicate identities and mutated expansion content. Hegel adds
250 requested policy/status cases alongside the existing 200 discovery cases.

Databricks execution/deployment, raw-source SQL preparation, duration operators,
objective/alert configuration and browser rendering of these descriptors remain
unimplemented. Existing Quint lifecycle models were not extended or rerun for
this stateless binding/classification slice. See [the executable CLI storyboard](docs/PACKAGE_EXPANSION.md).

## Prior semantic packages and worker/queue adapter

Current additions on 2026-10-17:

- Ten versioned Go contract profiles and per-resource-scope capability proposals.
  CLI catalog/discover examples run; unknown versions/units and co-observed
  dimension requirements are tested. Generated UI catalog matches the Go source.
- Root Go race tests and vet pass. New Hegel property passes with 200 requested
  generated cases across profiles: absent/mis-versioned/mis-unit evidence cannot
  be borrowed from another scope; inventory ordering and duplication do not
  change proposals. Native verification stays in its separate module.
- Nineteen Python API/SQL/adapter tests pass, including exception-safe bounded
  worker counts, unknown broker lag, clock violations, per-instance gauge
  snapshots and raw-observation wait percentiles. Three Node geometry tests pass.
- Six real Chromium journeys pass (45.1 seconds), capturing 21 documentation
  frames. New coverage observes active work through actual Redis → OTel →
  ClickHouse ingestion, checks all queue/worker metrics and opens the package
  catalog. The final catalog-only update is also checked with the queue/library
  journey. Demo traffic is restored by the wrapper.

The new catalog is metadata recognition/proposal generation, not automatic live
inventory extraction, source/value-level binding, executable expansion of all
profiles, or Databricks deployment. The reference queue view is implemented;
most other profile outputs remain explicitly proposed. Versioned examples are
illustrative declarations, not live-scan evidence. No new scheduler or custody
model was introduced; existing Quint lifecycle models were not extended to claim
queue-adapter or package-expansion verification.

Queue observations distinguish undelivered lag, pending (including stale work),
local active/capacity and first-start waiting time. No stream reset was used;
the current wrapper reports three stale pending entries, up from two before the
consumer update. The sample still cannot reclaim crash/interrupted work.
See [semantic packages](docs/SEMANTIC_PACKAGES.md),
[queue contracts](docs/QUEUE_TELEMETRY.md), and [journeys](docs/journeys/README.md).

## Prior visualization and guided-investigation slice

Passed against the current local stack on 2026-10-17:

- Fifteen Python API/validation/embedded ClickHouse tests. New checks cover
  duration-band boundaries, zero and UInt64-maximum durations, exclusive time
  bounds, invalid selections, and 1101-span heatmap totals versus 200-row inspection.
- Three Node geometry checks cover exact nanosecond offsets, scoped parent
  identities, missing/cyclic parents, and zero-duration/empty spans.
- Five live Chromium journeys pass (41.7 seconds), capturing nineteen storyboard
  frames. Added paths exercise heatmap → exact raw spans → reservation waterfall
  → async producer link, and guided retry/USE/RED navigation with scope checks.
- Inspected actual waterfall and USE screenshots; Compose configuration and
  JavaScript/Python syntax checks pass. Demo traffic is restored after the run.

The first async navigation run correctly exposed an omitted producer span: its
sample instrumentation lacked reservation identity. New producer spans now carry
that identity and related attributes; the passing journey follows the exact
linked span. Older data is not backfilled and external links remain labelled.
The heatmap's minute conversion was also checked against both embedded and live
ClickHouse rather than relying on assumed timestamp function return types.

These visualizations remain in the Python reference lab, not a Go/Rust API
migration. USE utilization/queue measurements and a complete RED error-rate
measure are explicitly absent. Historical custody bases and Go lifecycle receipt
integration remain unimplemented in the UI. Existing Hegel/Quint semantics were
not changed or rerun for this presentation/raw-span slice.
See [visualization contracts](docs/VISUALIZATIONS.md),
[guided investigations](docs/INVESTIGATION_WORKFLOWS.md), and
[captured journeys](docs/journeys/README.md).

## Prior lifecycle, journeys and verification

Passed in the current cloud workspace:

- Go race tests and vet, plus all twelve existing Python tests.
- Thirteen independent lifecycle scenarios and 26 total/grouped aggregate SQL
  comparisons in embedded ClickHouse, including old/new snapshot coexistence.
- Three Hegel properties: generated deadline/dedup cases across epoch ranges,
  historical evidence identity under invisible appends, and state/swarm sequences
  with invariants after every enabled step (300/200/150 requested cases).
- Quint 0.32.0 bounded simulation: 100 traces, maximum 30 steps, seed 42.
- Eight Quint-generated MBT traces replayed through the actual Go preparer:
  160 steps with state comparisons. An eight-step observed Go trace conforms
  and passes `safety`; an impossible good count is correctly rejected.
- Three real Chromium/Playwright journeys pass after workload isolation,
  capturing twelve documentation frames. The sample's good, retry, late, missing
  and rejected scenarios traverse the real services and investigation UI.

Observed and addressed: restored ClickHouse/Redis processes needed startup;
recreated application containers needed local-service proxy bypasses; an initial
live journey under background load missed the deadline; the documentation runner
now pauses/drains/restores the owned generator without deleting stream data.
Pre-existing stale pending messages remain unreclaimed. Screenshot review also
found and fixed stale view controls after expression edits and an inappropriate
metric-point disclaimer on trace/log/latency results.

The normalized-event kernel is tested in Go; generated SQL **preparation** from
raw exporter data is not implemented. The Go semantics are not yet connected to
the live UI. Browser coverage does not resolve the Python accounting defects.
Databricks deployment/query/ingestion and optional HyperDX remain unverified.
See [journey frames](docs/journeys/README.md), [lifecycle contracts](docs/LIFECYCLE_PREPARATION.md),
[verification commands](verification/README.md), and [the bitemporal spike](docs/BITEMPORALITY.md).

## Portable Go aggregate pilot

The follow-on pilot is documented in [PORTABLE_COMPILER.md](docs/PORTABLE_COMPILER.md).
Passed in this workspace:

- `go test -race ./...` and `go vet ./...` using Go 1.27.2.
- `python scripts/check_portable.py`: ten total/group fixture comparisons across
  independent expected results, the Go evaluator, and generated embedded
  ClickHouse SQL; an additional declarative-count/escaped-literal case passes.
- Structural parsing of generated Databricks YAML and CLI output contracts.
- The runnable CLI example returns two eligible subjects and success 0.5.
- All original twelve Python unit/API/embedded ClickHouse tests still pass.

The docs URL was retried and remains blocked by proxy 403; the pinned official
Databricks creation reference was re-read from accessible GitHub content.
Databricks DDL/query execution, Unity Catalog deployment and physical ingestion
remain unrun. Normalized lifecycle preparation is now tested as described above;
real exporter adaptation and SQL preparation remain pending. The Go pilots do
not resolve the Python prototype's existing measurement defects.

## Original handoff record

The current lifecycle and journey validation is recorded separately in
[LIFECYCLE_PREPARATION.md](docs/LIFECYCLE_PREPARATION.md),
[the journey storyboard](docs/journeys/README.md), and
[verification/README.md](verification/README.md). These additions supersede the
original environment's unverified browser/default-stack status, while preserving
the distinction between the Python prototype and new Go semantics.

Created 2026-10-16.

Passed:

- Docker Compose 2.35.1 `config --quiet` against the complete manifest.
- Twelve Python unit/API tests, including all five query modes executed in embedded ClickHouse 24.8 through chdb 3.3.0.
- Application flow test: real HTTP and gRPC processes, fake Redis TCP/Streams, real consumer processing, and a local protobuf OTLP receiver. Verified accepted/retry/terminal log records, gRPC server spans, asynchronous span links, and exported metrics.
- Python compile checks and JavaScript syntax validation.
- OTel SDK/exporter construction with pinned Python dependencies.

Not verified in this workspace:

- Full Compose startup, collector configuration acceptance and ingestion into the pinned ClickHouse 25.3 container, image pulls, optional HyperDX startup, and `scripts/smoke.py` against that stack.
- Browser rendering. The Playwright browser download was unavailable; only JavaScript syntax was checked.

Docker 28.0.4 and Compose 2.35.1 binaries were installed for validation. The hosted workspace denies mount, Unix-socket, and user-namespace operations needed by the Docker daemon; rootless setup is also unavailable. Installing binaries cannot supply those host permissions. Run the provided full-stack smoke test on a Docker-capable machine before treating the container integration as verified.

The first application-flow run exposed inherited trace sampling that discarded demo spans. The SDK now explicitly retains all sample traces. A subsequent integration run verified the emitted spans and links.

## Codex handoff packaging

The handoff adds repository guidance, a continuation task, architecture, two proposed RFCs, an implementation plan, known issues, source references, and a portable snapshot of the original design Page. No application behavior changed during this packaging step.

The twelve unit/API/embedded ClickHouse tests and Compose configuration validation were rerun successfully. Local Markdown links and fenced code blocks were checked. The archive excludes caches, environments, downloaded runtime binaries, and database volumes, and includes SHA256SUMS for its source and documentation files. Full container startup and browser rendering remain unverified as noted above.

## Log and trace discovery

Implemented readable literal log search, independently scoped surrounding logs,
matching reported trace-ID groups, fixed full-population span-duration buckets and
available-field suggestions feeding reusable packages. See the
[discovery contract](docs/EXPLORATION_DESIGN.md#log-discovery-time-context-and-span-duration-buckets),
[journey 30](docs/journeys/README.md#30-from-log-search-to-reusable-span-distributions)
and the current detailed results in [ENVIRONMENT_VALIDATION.md](ENVIRONMENT_VALIDATION.md).
This is a captured-record investigation workflow; it does not establish complete
traces, root latency, live tail or production readiness.

Final inventory milestone: default browser run passed 29 journeys (two opt-in
skips); explicit fault, generated-success and final-image inventory campaigns
passed together (3.3 minutes). All 31 distinct journeys passed across these runs.
The final inventory actor recorded 685 steps with all required stages and rejecting
Quint controls. All 95 PNG documentation frames and the generated-campaign JPEG frame were regenerated; see
[the cloud record](ENVIRONMENT_VALIDATION.md#2026-10-18--observed-service-inventory).

## 2026-10-20 — exact-stream live counter comparison

The native live pilot uses the existing typed evaluator and exact OTel identity.
Focused source tests and race checks cover nanosecond bounds, newest-window gaps,
cumulative resets, changed instances, stale endpoints and source-free freeze.
`go test ./...`, `go vet ./...`, the three affected packages under `-race`, 68
frontend tests and `TestHegelCounterWindowRateAndCoverageState` pass. See the
[counter contract](docs/LIVE_COMPARISON.md#exact-stream-counter-rates) and
[journey 72](docs/journeys/README.md#72-follow-a-native-counter-stream-without-hiding-export-gaps).
Five real-source browser checks and the final-image native rerun (including
Bombadil and observed Quint replay with seven rejecting controls) passed.
All 12 services are running with matching served-source and binary hashes.
Deployed image evidence is recorded in ENVIRONMENT_VALIDATION.md.

## 2026-10-20 — exact native stream selection

Native measurement results now offer read-free Focus this stream previews using
an immutable `source.exactStream` package binding. Compilation applies exact
resource/point map and schema identity before evaluation; off-stream evaluator
input is rejected. Share/save/capture/live/freeze/reopen retain the binding;
Explore all matching streams removes only the exact pin from a new question.

Root Go suite/vet, affected-package race tests, cross-kind replay checks, frontend
suite (22 test files) and 100 generated twelve-action Hegel focus/clear sequences
passed. Journey73 verifies seven real OTel identities, exact-window focus without
a read, package sharing/save/capture, pinned 3/s versus 5/s live comparison,
source-free freeze/reopen and clear. Its observed Quint model rejects nine
corruptions; Bombadil exercises Run/Focus/Run/Clear. Actual image, campaign and
screenshot evidence is in ENVIRONMENT_VALIDATION.md and
[the selection contract](docs/EXACT_STREAM_SELECTION.md).

## 2026-10-20 — retained native stream navigation

Native results and captures now offer full-retained-inventory identity search,
Needs review filtering, 25-stream paging and explicit local empty results. These
controls preserve question, source totals and evidence basis. Coverage guidance
maps existing typed states/reasons to readable explanations while retaining exact
codes. Gauges keep independent sample/freshness semantics; no new rates or
coverage evaluator is implemented in the browser.

The frontend suite passes 22 test files, including ten focused native/stream/live
checks and generated local navigation sequences. Existing Hegel exact-stream
state checks pass. Journey74's actual 28-stream capture validates zero-read local
navigation, unknown coverage, immutable question/basis, disclosure retention,
mobile bounds, exact Focus preview and reopen. Quint accepts its observed trace
and rejects nine corruptions; the navigation Bombadil actor exercises review,
empty/matching search, clear and paging. Journey72's live native gap explanation
also passes. Final image and additional regression results are recorded in
ENVIRONMENT_VALIDATION.md. See [the contract](docs/NATIVE_STREAM_NAVIGATION.md).

Final navigation campaign: all five real-source journeys passed together (4.1
minutes), including both native navigation and exact-stream Bombadil actors.
The navigation actor recorded 443 steps and exact-stream actor 534; actual Quint
ledgers passed with nine rejecting controls each. All 12 services are healthy
with matching served-source/binary hashes and historical pending IDs preserved.
Affected 38/57/72/73/74 frames were regenerated; desktop/mobile diagnostics were
visually inspected. See ENVIRONMENT_VALIDATION.md for the exact final image.

## 2026-10-20 — native service overview integration

Native service cards now present evaluated latest interval/sample values with
separate whole-window coverage, using shared passive tables/stats/attributes.
Counter/histogram `ok` interval states and native unbounded quantile ranges are
handled correctly; shared histogram formatting preserves bounds and empty
observations. Newer unknown values never fall back to older valid intervals.
Rows open exact-stream questions; retained inventories are constructed lazily
and searched without telemetry reads. Planner budgets and Go evaluator semantics
are unchanged.

The frontend suite passes 22 files; focused native-card checks cover actual kernel
states, unknown latest observations, zero, partial-window/valid-interval coexistence,
stream detachment and histogram ranges. Existing Hegel dashboard-planning
properties pass. The existing dashboard journey passed an initial real-source
campaign; final affected service journeys, native actors, observed Quint ledgers,
image health and documentation frames are recorded in ENVIRONMENT_VALIDATION.md.
See [native service cards](docs/SERVICE_DASHBOARDS.md#native-service-cards-and-exact-stream-questions).

Final affected browser campaign: five journeys passed in 3.7 minutes. Native
Bombadil exercised all four required actions over 529 steps; dashboard Bombadil
recorded 228 steps. The observed native ledger passed Quint replay and nine
rejecting controls. Updated desktop/mobile frames were reviewed. Final health
checks passed for all 12 services, matching served assets/binary and preserved
historical pending telemetry; see ENVIRONMENT_VALIDATION.md for image and evidence.

## 2026-10-20 — bounded native service selection

Service dashboards offer supported observed instrument choices from the pure Go
planner. Edits are read-free; explicit Apply rereads the same fixed overview with
a two-native/two-concurrent-read budget. Omitted selection uses defaults; empty
selection performs no native reads. Same-name version/schema variants remain
separate, and older Apply generations cannot publish over newer intent. Choices
are visit-local; ordinary questions/captures retain their exact definitions.

Root Go/vet, affected-package race checks, the 22-file frontend suite and Hegel
planning/selection checks pass. Journey76 documents actual OTLP selection,
read-free edits, budget feedback, namespace isolation, capture/reopen, default
restoration and superseded responses, with observed Quint replay and nine
rejecting controls. Final campaign evidence is recorded in ENVIRONMENT_VALIDATION.md.
See [the selection contract](docs/SERVICE_DASHBOARDS.md#choose-native-instruments-for-this-service-visit).

Final selection campaign: six journeys passed in 4.5 minutes with all three
service Bombadil actors. The new actor recorded 540 steps and reached empty
selection confirmation; its observed Quint ledger passed all nine rejecting
controls. Keyboard focus and regenerated desktop/mobile documentation frames
were reviewed. Final health/source/binary checks passed for all 12 services and
historical pending telemetry remained preserved.

## 2026-10-20 — retained native timelines

Ordinary native results/captures and expanded service results now use a passive
scalar timeline with shared chart-frame/table components. Gauge dots remain
independent, counter rates come from evaluated intervals, and histogram charts
show interval counts with native bucket/P95 details. Unknowns are not zero;
selection is read-free and preserves fixed question/custody. Charts are lazy per
stream and capped at the latest 200 retained observations/intervals.

All 23 frontend files and existing Hegel native measurement state/swarm checks
pass. Journey77 uses actual OTLP evidence with observed Quint replay and nine
rejecting controls; journey78 tests server/SPA lifecycle, keyboard selection and
mobile bounds without API reads. The initial two-journey campaign passed with
Bombadil. Final affected regressions and deployed source/image checks are recorded
in ENVIRONMENT_VALIDATION.md. See [the timeline contract](docs/NATIVE_TIMELINES.md).

Final timeline validation: all seven affected journeys passed together (7.1
minutes), with exact-stream/native-navigation/timeline actors. After the visual
unknown-dot correction, both new journeys passed again (1.4 minutes), including
a computed-style assertion. Latest observed Quint replay passed all nine
rejecting controls; Bombadil exercised all required inspection actions. Final
health and served asset/server/binary checks passed for all 12 services; historical
pending telemetry remained preserved. Documentation frames were refreshed and
reviewed on desktop/mobile.

Final bounded-table confirmation: journeys38/77/78 passed again in 2.1 minutes
with an explicit table-height assertion, the unknown-glyph assertion and native
timeline Bombadil/Quint checks. Final source/image health matches this build;
the original 510 px observation-table scroll bound is preserved.

## 2026-10-20 — compatible native histogram bucket heatmaps

Implemented a passive `goo-bucket-heatmap` display element, a lazy retained-interval
layout adapter and a captured-window count view gated by the Go comparison's
`comparable` result. Unknown intervals have no inferred layout, recorded zero stays
observed and other-layout cells are disabled. Counts remain exact strings; only
color geometry uses a numeric ratio. The retained grid is limited to the latest
32 intervals and 64 buckets per layout; exact tables remain authoritative.

Validation completed:

- `npm run test:visualizations`: all 24 frontend test files passed, including exact
  large counts, zero/unknown/outside states, fail-closed input/budget checks,
  incompatible equal-length layouts, stable count ceilings, original retained keys
  and input immutability. Artifact: `artifacts/native-heatmap-frontend.log`.
- Final image browser campaign: journeys 38, 54, 77, 79 and 80 all passed in 2.6m.
  Real OTLP evidence covers native measurements, layout switching, zero/unknown
  selection, cleared stale bucket tables, source-free saved-capture reuse,
  compatible/incompatible captured-window grids, mobile overflow and exact table
  viewport behavior. Standalone contracts cover SSR attributes, detached and
  pre-definition properties, controlled selection/focus, invalid/empty data and
  exact counts above JavaScript's safe integer range. Artifact:
  `artifacts/native-heatmap-final-journeys.log`.
- `GOO_NATIVE_TIMELINE_BOMBADIL=1`: updated saved-capture actor completed 559 trace
  entries, exercised all five required custom actions including
  `goo.timelineHeatmap`, and passed continuous supplied-cell-state/count,
  disabled-layout, overflow and default exception invariants. Evidence:
  `artifacts/native-timeline/056759d6ff/bombadil`.
- Observed Quint replay passed for native timelines (17.52s, nine rejecting
  controls), native captured windows (11.40s, six rejecting controls) and the
  Measurements regression (22.99s). Actual ledgers/logs:
  `artifacts/native-timeline/056759d6ff`,
  `artifacts/native-windows/1792539097069`, and
  `artifacts/measurements/fee03094`. Existing models validate retained semantic and
  custody/read boundaries; no new browser metric evaluator was introduced.
- Hegel histogram state/swarm (100 cases) and merge (150 cases) checks passed,
  `go test -run 'TestHegelHistogramStateSwarm|TestHegelHistogramMerge' -count=1 .`
  in `verification`. The first invocation used the read-only default user cache;
  setting `XDG_CACHE_HOME=/workspace/.cache` corrected the environment. Artifact:
  `artifacts/native-heatmap-hegel.log`.
- JavaScript syntax checks, Python server compilation and `git diff --check`
  passed. Regenerated journey 38/54/77 frames and the new labelled synthetic
  component frame; reviewed desktop/mobile grid screenshots. Updated README,
  component/design docs, architecture diagram, executable storyboard, Bombadil
  notes, planning and known display limits.

The final deployed query/packages image is
`sha256:463ab7de755a77512de87e9e2211f1f5aeafe04bceb6dc095d4ddf9b51192fae`.
The wrapper restored the traffic generator and retained existing telemetry volumes
and historical pending entries. Final service/asset/binary verification is recorded
in `artifacts/native-heatmap-final-health.json`. Unused build caches and dangling
images were reclaimed without touching volumes. No production Go semantics changed,
so unrelated root/race suites were not repeated for this display milestone.

## 2026-10-21 — explicit histogram time buckets and whole-product component sweep

Implemented opt-in `histogramBucketMs` in native histogram packages and
`metrics.BucketHistograms` over evaluated exports. Half-open buckets anchor at
exact question start, retain a possibly shorter final window, and combine only
whole compatible intervals with complete coverage. Checked native bucket sums
produce exact populations/P95 ranges. Crossings, gaps, unknown/reset diagnostics,
layout changes, invalid counts and overflow withhold the entire affected bucket;
valid neighbors remain usable. Width belongs to the question/revision/basis.
Existing definitions/results omit the new fields and retain their transport shape.

The UI locally switches fixed-time/export views using the existing passive chart,
timeline, grid and inspector. Source reads and capture replay remain in adapters
and Go. Export diagnostics are labelled separately from bucket coverage. Selected
zero/unavailable buckets have explicit guidance and no invented percentile or
stale distribution. Added the executable journey81 storyboard and desktop/mobile
frames, architecture diagram, component/verification notes and documented budgets.

The parallel full-product agent prepared `docs/WHOLE_PRODUCT_COMPONENT_SWEEP.md`
and marked older audit proposals against current implemented components. The
report is a source/design and recorded-screenshot review, not a new user study or
live whole-app validation campaign. It prioritizes a passive evidence summary in
frozen native captures and linked investigation evidence; then selected-item
inspectors, existing table/chart adoption, shared-scale compositions, revision
catalogs and comparison presentation. Proposed components remain clearly distinct
from implemented histogram aggregation.

Completed checks:

- `go test ./...`: all root packages passed, including measurement definition,
  canonical width/budget validation, stored replay and corruption rejection.
  Artifact: `artifacts/histogram-time-root-tests.log`. Initial sandboxed transport
  tests could not bind loopback sockets; the actual full run used local socket
  permission and passed.
- `go test -race ./measurements ./metrics ./exploration` passed; the kernel agent
  also ran `go test -race ./metrics`. Final extra cumulative-differencing/reset
  case passed in `go test ./metrics`. Artifacts: `histogram-time-race.log` and
  `histogram-time-kernel-final.log` under `artifacts`.
- Hegel aligned-population property: 100 cases. Hegel state swarm: 100 cases × 20
  generated steps. The independent whole-export oracle covers arrivals/removal,
  repair, zeros, layout/epoch diagnostics, crossing widths, exact local sums,
  immutable inputs, full identity and deterministic ordering. Artifact:
  `artifacts/histogram-time-hegel.log` (0.776s).
- All 24 frontend test files passed, including the supplied-time-bucket adapter's
  complete identity matching, empty/unknown preservation and immutable input.
  Artifact: `artifacts/histogram-time-frontend.log`.
- Final-image browser journeys81/38/54 all passed in 2.2m. Journey81 injects real
  OTLP, merges two exports to seven observations, retains recorded zero,
  withholds unknown/layout/crossing buckets, preserves the clipped final bucket,
  switches views without reads and reopens exact saved evidence. Journey38
  exercises all measurement families and capture/notes reuse. Journey54 covers
  compatible comparisons and withheld reset/gap/layout cases. Artifact:
  `artifacts/histogram-time-final-journeys.log`. Raw native timeline journey77
  also passed (33.9s) in the first campaign.
- Observed Quint validation: journey81 passed in 15.45s with nine rejecting
  mutations; measurements passed in 13.56s; native windows passed in 11.24s with
  six rejecting mutations. Actual ledgers/logs:
  `artifacts/histogram-time/9b5aaa00fd`,
  `artifacts/measurements/30d839a1`, and
  `artifacts/native-windows/1792541195706`.
- Journey81's saved-capture Bombadil actor completed 651 trace entries and all
  five required custom actions (open, raw, fixed, zero, unknown). Supplied
  bounds/state/value, overflow, exception and completion properties passed.
- During the first regression campaign, the unchanged native-window Quint
  subprocess exceeded its existing 60s timeout during a transient load spike.
  UI assertions had completed. Cgroup `oom`/`oom_kill` counters were 0 and available
  memory was ample; observed load average reached 38.30. No timeouts/invariants
  were relaxed. The same captured ledger passed independently in 9.90s
  (`artifacts/histogram-time-native-windows-replay.log`), and the complete final
  comparison journey then passed normally.
- JavaScript syntax, Go formatting and `git diff --check` passed. Reviewed actual
  desktop/mobile time-bucket screenshots and refreshed affected 38/54/77 frames.

Final query/packages image:
`sha256:de83ea6e9b4b4bf5c7890af3d6d01b7ad9671b0875afbe5ed1d5e120e03003d4`.
The wrapper restored demo traffic, preserved telemetry volumes and historical
pending entries. Final 12-service health and served UI/server/binary hash checks
are recorded in `artifacts/histogram-time-final-health.json`.

Limits are explicit: 200 time buckets per stream, 20,000 total output buckets and
10,000 retained intervals; grid display remains latest 32 columns/64 native buckets.
No splitting, interpolation, fleet aggregation, normalized-rate heatmap or live
histogram subscription was added. Next: implement the sweep's evidence-summary
primitive in two real workflows before broadening its API.

## 2026-10-21 — shared evidence-summary primitive

Implemented passive `goo-evidence-summary` with bounded supplied facts, separate
mode/result state, a visible completeness/independent-read caveat and native
exact-detail disclosure. The application presenter `GooEvidenceScope` interprets
typed measurement receipts; the component owns no reads, actions, subscriptions,
clock inference, navigation or persistence. Adopted in the full measurement capture,
linked native question evidence and a separate native capture card in notes.
Repeated capture scope/count/read paragraphs are suppressed where the summary
owns them; operational refresh/error/attachment status remains adapter-owned.

The presenter preserves all four namespace selector states, exact source counts,
question bounds, declared grain, binding and revision identity. Source read and
stored custody clocks are separately labelled. Missing receipt-only basis/read
metadata remains unavailable; no full-result fetch or custody substitution fills
it in. Full lifecycle results may additionally supply business evaluation and
knowledge-cutoff clocks. Invalid snapshots fail closed and clear prior facts/mode.

Completed validation:

- All 25 frontend test files passed. New pure cases cover bounds/invalid data,
  detached facts, mode/state independence, counts above JS safe integer range,
  exact nanosecond bounds, namespace distinctions and missing source-read clocks.
  Artifact: `artifacts/evidence-summary-frontend.log`.
- Final browser campaign: five tests passed in 2.1m: standalone journey83, real
  OTLP summary journey82, both linked-evidence journey51 cases (record and native),
  and shared Measurements journey38. The first campaign additionally passed
  histogram time-bucket journey81 with unchanged native semantics. Artifact:
  `artifacts/evidence-summary-final-journeys.log`.
- Standalone browser contracts cover pre-definition properties, SSR attributes,
  detached getter isolation, safe text, invalid state clearing, native keyboard
  disclosure, focus/open-state preservation during valid updates, no component
  buttons/links/inputs and mobile overflow.
- Real OTLP capture → linked investigation checks retain exact source/storage
  clocks, source counts, matching-scope guidance and receipt-only limitations.
  Local inspection preserves unsaved findings, performs no source POSTs and
  leaves the saved capture unchanged. Desktop/mobile frames 82a/82b/82c and the
  clearly synthetic frame 83 were generated from the browser and reviewed.
- New observed Quint replay passed in 13.56s with eight rejecting corruptions:
  evidence mode, clock roles/values, local reads, result state/count, linked read
  metadata and immutable reopening. Actual ledger/log:
  `artifacts/evidence-summary/637efe06fe`.
- The saved-capture/investigation Bombadil actor completed 581 trace entries,
  exercised all four required custom actions and passed passive-content,
  supplied frozen mode/state, overflow, exception and completion properties.
- Existing observed Measurements and linked-evidence Quint checks also passed;
  actual ledgers/logs: `artifacts/measurements/b0c8860d` and
  `artifacts/linked-evidence/1792543434585`.
- The first campaign caught a test-driver retry that overwrote navigation IDs
  after the initial hash changed. It was fixed by retaining IDs and making repeat
  navigation idempotent; completion bounds were unchanged. The old native linked
  test expected one source-read paragraph; it now checks the rendered definition
  list's source-read and stored-custody values separately. Both final cases passed.
- JavaScript syntax checks and `git diff --check` passed. No production Go evaluator
  or storage schema changed; backend semantic/race and Hegel suites were not
  repeated for this bounded presentation extraction. The actual capture paths
  continue to validate/replay typed Go results through the observed trace tests.

Updated README, architecture diagram, component/adoption contracts, whole-product
sweep status, executable storyboards, known limits and verification/Bombadil docs.
Final query/packages image:
`sha256:881e6fb0920695b40d565eb075f5496a14d7ff4e10a95e6781cb289975ccd253`.
Final health/served asset/server/binary hashes are recorded in
`artifacts/evidence-summary-final-health.json`; all 12 services run and historical
pending telemetry is preserved. The wrapper restored demo traffic. Unused build
caches/dangling images were reclaimed without touching volumes.

Current adoption is native captured evidence. Record/inventory summaries, service
widgets, comparison sides and selected-item compositions remain explicit follow-up
work. Next: reuse the evidence summary for service-widget evidence, with its own
typed fixed/live display adapter and clocks.

## 2026-10-21 — compact service evidence and readable UTC timestamps

Implemented shared summaries for supported record/native service widgets through
`GooEvidenceScope.serviceWidget`. Widgets lead with observations/charts/actions;
one collapsed scope/evidence disclosure holds their summary, read basis and typed
question/source warnings. Reduced summary/card/latency spacing and stopped short
cards stretching to match neighbors. The overview keeps independent-read,
unverified-coverage and capture-reread guidance visible. No planner, query budget,
source mapping or capture persistence schema changed.

Fixed widgets use fixed mode, with record `capturedAt` and native `readAt` treated
as source-read clocks, never custody. Exact counts and namespace distinctions
survive. Shared summaries show UTC timestamps/ranges and shorter same-day bounds;
raw nanoseconds and RFC3339 fractions remain in focusable tooltips and display
snapshots. Native stream tables, endpoint/interval disclosures and selected
interval details use the same timestamp formatter. Durations/counts retain their
native units; no query bounds are rounded.

Validation:

- All 25 frontend test files passed (`artifacts/service-evidence-unit.log`),
  including domain clocks, zero/large counts, four namespace states, exact bounds,
  negative/zero epochs, nine-digit fractions and invalid display formats.
- Final deployed browser run: **5 passed in 2.9 minutes**, journeys59/75/77/82/83
  (`artifacts/service-evidence-final-journeys.log`). Real OTLP inputs exercise
  partial source failure, stale-response guards, exact-stream continuation,
  immutable capture reopening, no reads during local disclosure, compact card
  sizing, mobile containment, keyboard inspection and exact-value tooltips.
- Journey59 observed Quint replay passed in 19.21s with eleven rejecting controls,
  including invented frozen mode, source clock and custody. Its Bombadil actor
  completed 466 steps without violations and exercised all eight custom actions
  (`artifacts/service-dashboards/1792544287285`).
- Journey75 observed Quint replay passed in 15.92s with its nine rejecting controls.
  Bombadil completed 567 steps without violations, covering all four required
  discovery/open/browse/link actions (`artifacts/service-native/a13c35576c`).
  Capture-summary and native-timeline journeys also replayed their observed Quint
  models and rejecting controls. No new Go semantic evaluator was introduced, so
  the existing Hegel planner/state suites were not rerun for this display change.
- Preliminary run: four browser journeys passed; service Bombadil caught a test
  selector reading a nested native diagnostic instead of widget source metadata.
  Corrected it to the widget provenance path and reran the complete five-journey
  set against the final image. The initial failure is retained in
  `artifacts/service-evidence-journeys.log`.
- Refreshed all affected service, capture/investigation, native-timeline and
  component-reference screenshots from the running app. Updated README,
  architecture/presentation diagram, component contract, storyboards, roadmap,
  whole-product adoption status, known limits and verification/Bombadil guidance.

Final query/packages image:
`sha256:703330e0f8b05c4a0b495e20eece866c380eaae2dfcec649e7e9cbd2e4093d43`.
Health, served asset/server/binary hashes and preserved historical pending IDs are
recorded in `artifacts/service-evidence-final-health.json`. The journey wrapper
restored demo traffic; no volumes or historical stream entries were reset.

Remaining adoption: comparison-side metadata, selected-item inspectors, legacy
chart presenters and diagnostic JSON. Those are not claimed to have migrated to
UTC/tooltips or the shared summary by this milestone.

## 2026-10-21 — inline latency readouts

Replaced the shared latency component’s large stacked labels/values and stretched
percentile columns with compact inline Mean/p50/P95/P99 pairs. Mean uses the same
13px value size as the quantiles. Spacing is smaller and pairs wrap on mobile.
The percentile-position rail and scale now live in the existing keyboard-accessible
Latency details disclosure with exact values and source caveats. Calculations,
units, missing/zero semantics and public component properties are unchanged.

All 25 frontend test files passed (`artifacts/latency-compact-unit.log`). Five
standalone/service browser tests passed in 1.3m
(`artifacts/latency-compact-journeys.log`). The standalone check bounds the
collapsed desktop fixture below 100px and proves the rail is hidden until keyboard
expansion. Real OTLP service/mobile/source-failure workflows passed with refreshed
journey59/63 screenshots. Observed Quint replay rejected all eleven corruptions
in 22.28s; service Bombadil completed 441 steps without violations, including widget
and latency disclosure controls (`artifacts/service-dashboards/1792545035823`).

Updated component/adoption documentation, README and storyboard index. Image:
`sha256:b2bdd6243e2e1d51ce2c2d163ca36f6931bc321aa03ffe30c638f42a802a042e`.
`artifacts/latency-compact-health.json` verifies 12 running services, served asset
hashes and preserved historical pending telemetry. Demo traffic was restored.

## 2026-10-21 — comparison presentation coherence

Aligned shared scalar pairs horizontally, with equal-width zero-based tracks,
explicit side labels and signed changes. Reduced pair/matrix padding and retained
separate count, duration and percentage domains. Captured latency continues to
use one aligned Mean/P95/P99 matrix and one duration scale. Native histogram bins,
exact count grids, incompatibility refusals and all Go calculations are unchanged.

Stored record/native comparison sides now use `GooEvidenceScope.comparisonSide`
and the existing passive summary inside the scope/receipt disclosure. Exact
counts, windows, predicate/binding mode and separate read/custody clocks remain
inspectable. Missing source reads never borrow storage time; unavailable basis
metadata is not fabricated or fetched. Gauge endpoint/native-window times and
fresh native bounds show UTC with exact nanosecond tooltips. Fresh baseline
selection also no longer substitutes storage time for a missing source read.

Validation:

- All 25 frontend files passed (`artifacts/comparison-coherence-unit.log`), including
  frozen-side metadata, exact/zero counts and missing-clock handling. An initial
  pure fixture exposed missing grain metadata; the presenter now leaves it
  explicitly unavailable rather than invalidating the display.
- **11 browser tests passed in 5.0m**
  (`artifacts/comparison-coherence-journeys.log`): standalone SSR/SPA components,
  live/saved error and latency comparisons, stored record/gauge comparisons,
  fresh capture/retry, native reset/gap/layout refusal and saved findings/recipient
  reopening. Mobile containment and affected screenshot frames were refreshed.
- Journey52 checks equal track widths/top alignment, frozen mode, exact independent
  read/storage fields, receipt immutability and no source reads during inspection.
  Stored-comparison observed Quint replay passed in 8.36s with rejecting controls;
  its updated Bombadil actor completed 399 steps with both compare and scope/clock
  inspection actions and no violations
  (`artifacts/evidence-comparison/1792552326230`).
- Native-window observed Quint replay passed in 26.71s with its rejecting controls;
  Bombadil completed 197 steps without violations
  (`artifacts/native-windows/1792552422225`). Fresh, error and saved-comparison
  observed replay also passed; saved replay took 28.79s
  (`artifacts/saved-comparison/1792552512464`). Existing Hegel semantic suites were
  not rerun because no Go evaluator or planner changed.
- Updated README, comparison presentation/diagram, component and evidence
  contracts, storyboards, adoption plan/audit, known limits and Bombadil/verification
  guidance. All frames came from the actual OTLP-fed app.

Final query/packages image:
`sha256:05477c1b103fc4b05507d00d599d5e1510025961a8dbafc8d62de334dc18a368`.
`artifacts/comparison-coherence-health.json` verifies served assets, server/binary
hashes, 12 running services and preserved historical pending telemetry. The wrapper
restored traffic. Unused build cache/dangling images were reclaimed before the
build; no volumes or stream entries were reset.

Remaining work includes general selected-item inspectors, legacy diagnostic JSON,
and deeper live-comparison composition. This milestone does not change capture,
refresh, freeze, SSE or query lifecycle protocols.

### Compact selected-item inspectors (2026-10-21)

Shared caller-owned inspector presentation now covers selected trace spans, log/span
records and native measurement points. It retains 32 recent content nodes and their
scroll/disclosure state within one rendered result; Close returns focus without
changing source selection. Record details appear beside their selected row. UTC
summaries preserve exact timestamp bounds in disclosures/tooltips. Existing inline
attribute disclosures remain available; this is not a cross-query selection store.

`npm run test:visualizations` passed all 25 frontend test files. The broad browser
run passed all 25 legacy investigation journeys, plus three inspector-related
journeys (28/29 overall); the remaining test incorrectly counted the trace adapter's
existing linked-log refresh as an inspector read. Its assertion now distinguishes
span-selection reads from source-free inspector disclosure/close.

The next targeted runs exposed a native focus-return defect and a test-actor
visibility error. Native openers now identify the exact plot mark/heatmap cell;
a browser regression covers programmatic selection while another mark has focus.
The actor now enters Saved investigations and opens the stream disclosure before
asserting focus, rather than interacting with hidden content. A hidden-panel guard
also preserves scroll when reopening a closed inspector. The initial targeted
runs remain recorded as failures, not successful validation.

A parallel, read-only product research pass produced
[CORE_EXPERIENCE_PRODUCT_SURVEY.md](docs/CORE_EXPERIENCE_PRODUCT_SURVEY.md).
It distinguishes recommendations from shipped foundations and verified official
Datadog/Grafana source documentation from inaccessible Honeycomb material.
README and the implementation plan link its prioritized follow-ups; the survey
does not itself change runtime behavior.

Final release validation: **4/4 passed in 3.4 minutes**
(`artifacts/inspector-release-final.log`): journey 35 actual log/span inspection,
84 local shell retention and mobile composition, 77 native timelines with
Bombadil action coverage and observed Quint validation, and 29 typed inspection
with Bombadil. The release image is
`sha256:2cb6bc1c7827984f73d642e36e13d3e7b40f0ffc54e9ff4831db7c4dfafa9a7d`.
Screenshots and executable storyboards are refreshed. No query, capture, SSE or
semantic evaluator protocol changed. Existing trace selection still refreshes
linked logs independently; the inspector adds no source read.

`artifacts/inspector-final-health.json` confirms matching query/packages images,
served source hashes, server/binary hashes, 12 running services and preservation
of all five tracked historical pending IDs (14 pending entries at the check).
The wrapper restored traffic. Only unused Docker caches/dangling images were
reclaimed during builds; no volumes or stream entries were reset. Cgroup OOM and
OOM-kill counters remained zero.

### Core experience refinements (2026-10-21)

Bounded first slices of the six product-survey recommendations now compose existing
Go aggregates and passive UI elements: eight retained Explore visits, exact typed
population selection with an imported retained comparison baseline, compact service
operation diagnosis/instance pivots, inclusive trace breakdown/highlighting,
independent deployment context on service and comparison views, and actionable
convention/instrument capability explanations. See
[the workflow and limits](docs/CORE_EXPERIENCE_REFINEMENTS.md).

Frontend checks pass all 26 test files, including exact half-open selection bounds,
OR-preserving narrowing, missing/empty values and rejected incompatible selections.
Hegel's 80 generated selection cases and 100 question-session state/swarm cases
passed (`artifacts/core-refinements-hegel.log`, 23.072s). The initial Hegel invocation
could not create its default home cache; rerunning with the existing writable
`XDG_CACHE_HOME=/workspace/.cache` succeeded. No credentials or runtime selectors
were replaced.

Initial browser validation found ambiguous selectors after nested disclosures and
the Services recipe were included. The first broader run passed discovery (with
Bombadil/observed Quint), error comparison and service browsing, but failed five
of eight tests: one ambiguous navigation selector, three assertions exposing the
missing outer Explore phase contract, and one old diagnosis-table shape assertion.
These are recorded failures (`artifacts/core-refinements-release.log`), not passing
release evidence. The wrapper now preserves the public phase contract, inactive
visits cannot publish state, and tests assert the new table/links. The discovery
ledger continues to verify exact scope and now expects the explicit service/name
aggregate presentation fields on expanded trace questions.

Lifecycle review also added listener disposal on visit eviction, preserved selected
bounds when importing a wider baseline, retained display DOM during reconnection,
shared diagnosis scales across mean/p95 values, and source-free staged link entry.
The dimension widget replaces repeated latency summaries with its diagnosis table.

A standalone Chromium probe confirmed that detaching a scroll container resets its
80-pixel offset to zero. The trail now saves/restores nested offsets explicitly;
journey 85 includes a long typed attribute inspector and an 80-pixel return check.
An intermediate release attempt was intentionally interrupted while waiting on an
incorrect service-button accessible name; its wrapper restored traffic. The test
now uses the catalog’s actual Open service name and bounded action timeouts.

Final image:
`sha256:49b36d545ef052ae01e3f9f73f9cff886ce6a9c0dce4167e3175b6f7eb01faaa`.
The broad final regression run passed **9/10 in 3.6 minutes**, including both shared
question journeys, workbench stale-completion rejection, service browsing/default
widgets, both stored comparisons (with independent receipt-lane bounds), retained
visit eviction/drafts, and typed inspection with its Bombadil actor. Journey 85's
product checks and observed Quint ledger passed, but the new actor passed action
factories instead of invoking them; this caused its sole failure.

After fixing the actor registration, the focused final rerun passed **2/2 in
55.1 seconds** (`artifacts/core-refinements-final-actor.log`). Its actor executed
Run, Preview, Follow, Back and Forward, satisfied eventual completion and retained
unique active DOM identities without overflow. The actual journey verified an
80-pixel inspector return, retained actual DOM/search, exact narrowing of six
spans to two, preserved narrowed shared-question bounds, source-free return,
trace highlighting, compact shared diagnosis scales, instance staging and explicit
deployment inspection. The observed ledger accepted the real run and rejected
four scope/read mutations (7.85s). The union of the broad and corrected targeted
runs validates all ten distinct journeys; the earlier failing run is not relabelled
as a clean ten-test pass.

Screenshots 85a–e and affected existing journey frames were regenerated from the
OTLP-fed application. README, survey status, implementation/known-issues plans,
component and architecture docs, executable storyboards and verification guidance
are updated. No Go evaluator/compiler, telemetry schema, capture protocol or SSE
protocol changed. In-tab visits remain bounded and non-durable; native exemplars,
exclusive/critical-path analysis and general convention resolution are not added.

`artifacts/core-refinements-final-health.json` verifies matching query/packages
images, served source hashes (including both new adapters), unchanged server/binary
identity against the build context, 12 running services and all five tracked
historical pending IDs preserved (14 pending entries at the final check). Traffic
was restored. Only unused Docker build caches/dangling images were reclaimed;
no data volumes or stream entries were reset. Cgroup OOM and OOM-kill counters
remained zero.
