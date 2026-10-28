# HTTP API contract

The current OpenAPI 3.1 contract is [api/openapi.json](../api/openapi.json).
Run the Python UI server and open [API docs](http://localhost:8088/api-docs), or
[download the spec](http://localhost:8088/openapi.json). The docs are read-only,
searchable and self-contained: no CDN, telemetry query or API execution widget.
The running server must include the current routes and assets; changing the
checkout alone does not update an existing process. Rebuild an older image, or
use the supported local code/asset installation and restart procedure.

The spec inventories 78 supported operations: 71 Go operations, six Python-only
`/api` operations and the Python root `/health` probe. Go paths are unprefixed on
port 8089; the browser proxy exposes them under `/api` on port 8088. Python-only
operations override the server list. `/api-docs`, `/openapi.json`, the UI, static
assets and unsupported prefix suffixes are documentation/application resources,
not additional observability API operations. Go's automatic HEAD/mux refusal
behavior is not a separately supported browser endpoint. These counts come from
the checked-in spec's method entries and `x-go-source` ownership, not from the
number of path strings. [Executed acceptance](API_ACCEPTANCE.md) and
[known limits](KNOWN_ISSUES.md) are separate from structural API coverage.

```mermaid
flowchart LR
  Browser[Browser or API client] --> Proxy[Python /api boundary]
  Browser --> Go[Direct Go HTTP boundary]
  Proxy --> Go
  Python[Python-only query / explain / declarations] --> Proxy
  Types[Go JSON wire types] --> Shapes[Reflected wire schemas]
  Policy[Reviewed route effects / limits / SSE contract] --> Spec[OpenAPI 3.1]
  Shapes --> Spec
  Spec --> Docs[Read-only API docs]
  Fixtures[Frozen Go wire fixtures and actual SSE frames] --> Check[CI conformance and rejecting controls]
  Spec --> Check
```

## What the schemas promise

All supported Go request and response envelopes have typed shapes. Nested
`json.RawMessage`, dynamic record rows, arbitrary exporter maps and interface
values remain open where the implementation is open. Python legacy query results
have a typed envelope with adapter-shaped rows; the legacy package keeps its
extensible declaration shape. This is the implemented development API, not a
stability promise or a replacement for semantic validation.

The Go generator reflects JSON field names, embedded structs, pointers, arrays,
maps and omission behavior. Input models allow omitted/null fields where Go's
decoder does, with unknown canonical fields refused. Go's case-insensitive JSON
matching and duplicate-field handling still apply; clients should use the canonical
names in the spec. Response fields without `omitempty` are required. Nanosecond
bounds, counts, revisions and exact telemetry values remain strings where the Go
wire types use strings. Do not convert them through JavaScript `Number`.

Schemas express structural encoding, not every valid question, combination, finite
integer range, lineage relationship, count accounting rule or capability state.
Go remains responsible for validated fields/operators, units, temporality,
resets, complete-population totals, query compilation and evidence identity.
A structurally valid request may return 400. OpenAPI validation does not prove
source-free reopen, custody ordering or absence of hidden reads; behavioral tests
do. No generated handler/client or request-validation middleware was introduced.

## Legacy source GET controls

`/http-package`, `/operations`, `/counter-rates` and `/histograms` now share the
four-slot source budget and optional `X-Goo-Query-ID` cancellation control. Their
implicit HEAD requests use the same limits. Saturation returns 429 with
`Retry-After: 1`; invalid/reused IDs return 400/409. The enclosing deadline is
60 seconds, with a 55-second source deadline after validation. Admission precedes
route validation and runtime capability checks. `/operation-packages` remains a
source-free declaration read outside this admission. Existing success bodies,
query validation and source-error mapping are unchanged. See the
[boundary details](HTTP_BOUNDARIES.md#legacy-operation-source-admission-2026-10-28).

## Two boundaries, with existing differences

Every Go POST operation documents `x-go-body-limit-bytes` and
`x-browser-body-limit-bytes`. These are byte ceilings, not string lengths. The
Python proxy checks positive Content-Length before forwarding. In particular:

| Route | Go | Browser proxy |
| --- | --- | --- |
| `/dashboards/plan` | 256 KiB | 16 KiB |
| `/metric-discovery` | 64 KiB | 16 KiB |
| `/annotations`, `/deployments` | 128 KiB | 16 KiB |
| `/compare` | 1 MiB | 1 MiB |

Direct Go decoders accept JSON without enforcing Content-Type; application/json
is the documented client encoding. Trailing-input errors differ by owner. The
proxy reparses/reserializes JSON, so byte-identical response fixtures describe Go,
not Python's JSON formatting. Direct operation GET validation uses text/plain
`http.Error` bodies containing JSON-looking text; the browser re-encodes parsed
errors as JSON. The proxy does not forward all upstream headers (including query
budget Retry-After). These are current discrepancies, not changes made by this spec.

`x-effects` distinguishes pure plans/supplied-summary comparisons, source reads,
source reads with possible custody/capture publication, saved reads, file
publication, optional XTDB operations and cancellation. GET does not imply no
telemetry execution. `/captures/live` reads fresh evidence or freezes a held draft;
its 200 draft and 201 frozen-page responses are separate. Opening saved evidence
does not imply a telemetry reread. Optional XTDB configuration status is not health;
per-operation read basis remains distinct from the configured connection.

The lab has no authentication layer. `X-Goo-Query-ID` and draft/result tokens are
short-lived local capabilities, not authorization. Do not infer authentication
from ClickHouse credentials held by the server. See the fuller
[HTTP ownership contract](HTTP_BOUNDARIES.md) for budgets and effects.

## Nullable grouping values

Fixed physical columns such as `column:service` are always present in valid source
records. `/compare` now refuses caller-supplied null groups for those fields with
HTTP 400; earlier validation admitted that malformed representation. Attribute and
resource keys remain nullable, and the new derived `column:exception_type` explicitly
allows null when no unique reported event type exists. Its present empty string
remains separate. This is a request-validation tightening; valid legacy source
captures, hashes and unrequested row SQL are unchanged. See [error triage](ERROR_TRIAGE.md).

## Population explanations

Three Go routes are exposed through the browser proxy under `/api`:

| Route | Effect | Request |
| --- | --- | --- |
| `POST /population/explain/plan` | Pure projection/validation | `ExplanationRequest` |
| `POST /population/explain` | Explicit aggregate source read | `ExplanationRequest` |
| `POST /population/explain/filter` | Pure original-side predicate plan | Request, field, required null-or-string value, side and include/exclude mode |

All three have a 64 KiB body ceiling at Go and browser boundaries. The source route
uses the existing controlled query ID/cancellation mechanism and a 55-second total
budget; each source operation has a three-second child budget. Planning and filter
staging do not read telemetry. The versioned request names both original queries
and one to six explicit qualified fields. The result contains independent counts,
field statuses/read bases, exact rational shares/differences and ranked row
references. Structural API validation is not proof that a field is complete or a
population difference is causal.

`POST /captures` also accepts `{title,resultToken}` for an observed explanation.
The server loads its typed held result and derives the saved request; clients do
not submit aggregate data as evidence. This publication performs no telemetry
read. Repeating the same held token/title/store returns the same capture ID.
Expiry/refusal never starts another explanation automatically. The existing
16 KiB capture-body limit remains unchanged because the token avoids resending
two potentially large queries. Ordinary capture creation and explanation promotion
retain different effect contracts; an explanation cannot be refreshed by silently
rerunning its questions through the ordinary capture path.

`CapturePage.explanation` contains the saved result without its ephemeral token.
Its source clocks/bases remain separate from receipt custody time. See
[the explanation contract and UX](POPULATION_EXPLANATIONS.md) for bounds, partial
fields and source-free reuse.

## Tail event contract

POST `/tail` returns `text/event-stream` after validation/admission. Each frame
has a connection-local numeric `id`, an `event` and JSON `data`, followed by a blank
line. `x-sse-events` binds event names to schemas:

| Event | Meaning |
| --- | --- |
| `reset` | Connection ID, rolling window 60,000 ms, selector, unknown continuity, no replay |
| `batch` | Supplied rows, limited flag, fixed poll bounds/observation time and explicit membership state; available batches carry an optional preservation capability |
| `gap` | Source unavailable, identity budget exhausted, or connection ended |

The Go stream has four local slots, a 60-second context, two-second polls and
five-second write deadlines. Python's upstream socket timeout is 45 seconds, not
a guaranteed stream duration. Disconnect releases the upstream slot. Use POST
fetch streaming; browser EventSource cannot submit the selector body. Last-Event-ID
does not provide replay/resume. This polls an event-time window, not an ingestion
cursor or complete capture. A gap after headers remains HTTP 200, and disconnect
can prevent delivery of a final frame. Successful streams set no-store and
X-Accel-Buffering: no; direct early JSON errors omit no-store, whereas the proxy's
JSON errors include it. The JSON response helper must not buffer SSE.

Actual reset/source-unavailable gap bytes are frozen in
[testdata/http/tail.sse](../testdata/http/tail.sse) and executed by
`TestAPITailSSEWireContract`. Batch schema examples and rejecting controls complement
existing real-loopback batch/disconnect tests; they do not replace those tests.

## Updating and checking

From the repository root:

```sh
go run ./scripts/openapi-schema api/go-schemas.json
python3 scripts/build_openapi.py
python3 -m venv /tmp/goo-api-tools
/tmp/goo-api-tools/bin/pip install -r requirements-api.txt
/tmp/goo-api-tools/bin/python scripts/check_openapi.py
go test ./cmd/goo -run 'Test(APITailSSEWireContract|PlanningHTTPWireFixtures|TelemetryHTTPWireFixtures|NativeJSONHTTPWireFixtures|ExplorationJSONHTTPWireFixtures|TraceSearchHTTPWireFixtures)$' -count=1
python3 -m unittest discover -s tests -p test_api_docs.py -v
node scripts/check_api_docs.cjs
```

[CI](../.github/workflows/api-contract.yml) regenerates both checked-in artifacts,
refuses drift, validates OpenAPI/references and every Go route, and checks typed
requests/responses against the existing wire fixtures. Unknown-field, wrong-error-
type and malformed SSE controls must fail. The 236 HTTP response cases comprise
222 documented-operation responses and 14 default mux refusals; the checker also
validates 56 admitted request bodies. The workflow executes those wire tests plus the actual SSE fixture,
and tests the Python docs routes. The browser journey runs locally with the
existing Playwright/Chromium setup and captures a real docs screenshot; it does
not start the full lab or mutate telemetry.

The fast CI tier also installs locked npm dependencies on Node 24.19.0, runs
`go test ./... -count=1` for the root Go module and runs all `tests/*.test.js`
through `npm run test:visualizations`. The explicit wire selection includes the
trace-search HTTP handler fixtures. Source-dependent tests remain opt-in and
report skips without their configured source; nested `verification/` and
`instrumentation/` modules and real browser/store campaigns are separate tiers.
The local checks do not establish that a GitHub-hosted CI run has completed.

Schema roots and inline request envelopes live in
[scripts/openapi-schema/main.go](../scripts/openapi-schema/main.go); reviewed
route policy lives in [scripts/build_openapi.py](../scripts/build_openapi.py).
Route discovery is an inventory aid, not authority to invent an endpoint policy:
new routes must have reviewed effects, request/response roots, limits, supported
statuses and fixture coverage. Extensions describe current behavior and do not
change runtime enforcement. Executed coverage now also includes the capture/live-freeze/document workflow and
browser proxy contracts. The optional local XTDB branch covers retention, correction,
pinned review and source-free document reopen. Broader endpoint/branch coverage and
the remaining C2 acceptance work are still explicit gaps. The spec pauses further mechanical
JSON migration while those contracts become the common API reference.


## Observed workflow and proxy contracts

`TestAPISavedEvidenceObservedContracts` exports actual handler observations when
`GOO_API_OBSERVATIONS` is set. Its base branch has 12 responses: source-backed
capture and live draft, source-free/idempotent freeze, capture/receipt/list reads,
document publication/reopen/list, and optional-store status/refusal. With
`GOO_XTDB_TEST_URL`, seven more responses cover real retention, correction, manifest,
current/pinned review, pinned document creation and offline file reopen. The test
refuses any telemetry read after the source is marked unavailable. CI runs the
base branch; the real XTDB branch requires the lab explicitly.

`tests/test_api_proxy.py` puts real HTTP servers on both sides of the Python
proxy. It checks exact/overflow browser limits using the spec's byte ceilings,
query-string and cancellation-ID forwarding, deliberately unforwarded Retry-After,
JSON re-encoding of text/plain Go errors, early tail headers and idle upstream
closure within two seconds of browser abort. It can export 20 actual proxy JSON
responses using `GOO_PROXY_OBSERVATIONS`. Both observation files are validated by:

```sh
python scripts/check_openapi.py --observations /path/saved.json --proxy-observations /path/proxy.json
```

These tests found and corrected two contract issues: `/documents` takes an outer
`{document, parent}` envelope, and browser cancellation requires a nonempty body
(send `{}`) even though Go ignores it. They also found a runtime bug: non-JSON
upstream HTTP errors previously disconnected clients while handling the error.
GET, POST and pre-stream tail now return a bounded-description JSON 502 instead;
valid JSON errors retain their status. No backend error body is echoed by this
fallback. Existing limit/header differences remain unchanged.

`scripts/check_deployed_api.py` validates the running spec against the checkout
and performs readonly real API checks. The API-docs browser journey can target the
running lab with `GOO_API_DOCS_URL=http://127.0.0.1:8088/api-docs`; otherwise it
starts its own ephemeral server. These are distinct from synthetic loopback data.

## Managed lab rebuild

The Go build now binds source read-only in one build step, so `internal` packages
are included and VFS builders avoid copying a complete toolchain for each package.
The same Go version uses the smaller Alpine build image; the binary stays
CGO-disabled and the runtime remains Python slim. Go caches are disposable build
mounts. For a managed TLS proxy, use an optional trusted combined CA bundle:

```sh
GOO_BUILD_CA_BUNDLE=/etc/ssl/certs/ca-certificates.crt docker compose -f compose.yaml -f compose.xtdb.yaml -f compose.build-ca.yaml build query packages
docker compose -f compose.yaml -f compose.xtdb.yaml up -d --no-deps --no-build query packages
```

The overlay mounts trust only for dependency downloads and keeps certificate
verification enabled; it does not copy certificates into layers or alter runtime
configuration. Omit the XTDB overlay if the optional store is not deployed. Build
and recreate only these two owned services; preserve telemetry volumes, stream
entries, capture ledgers and the previous image for rollback.

[Executed API conformance and deployed acceptance](API_ACCEPTANCE.md) records the
workflow layers, replay commands, fixes, build recovery and remaining coverage.

Record pages now offer a compact [observed-field picker](RECORD_COLUMNS.md) with local
column visibility/order and explicit Include/Exclude intents. Population view
configuration accepts optional ordered `fieldColumns` (up to eight qualified
fields); saved layouts do not change query predicates or capture evidence.

Saved exploration-view breakdown `comparisonMetric` now also accepts `count`.
It remains display configuration alongside share/mean/p95/p99/error-rate; no
query semantics or receipt formats change. Go owns semantic validation.

Live record comparison responses optionally include `recordSeries: {baseline, selection}` when the saved record question requests buckets. These are existing retained `TimePoint` arrays, with question/state/read metadata on the corresponding comparison sides. The field is omitted for native and nonbucketed questions. Freeze returns those exact selection intervals without another source read.

Metric discovery responses now include an optional typed `capabilityPlan` per
instrument, version `goo/instrument-capability-plan/v1alpha1`. It reports metadata
eligibility and bounded observed-stream presence, not value/interval verification.
The request, source-read budget and generated package bindings are unchanged;
response types are reflected in the OpenAPI schema. See
[the contract](NATIVE_INSTRUMENT_DISCOVERY.md#typed-server-plan-and-observed-stream-presence).

Exploration predicates also accept `column:timestamp_ns` as UInt64 event time for
exact [time/duration region partitions](REGION_PARTITIONS.md). It uses the existing
typed predicate wire shape; it is not a new endpoint or an ingestion clock.


Origin-aware retained reads now have two explicit routes:

| Route | Response and effect |
| --- | --- |
| `GET /retained-evidence/catalog` | `EvidenceCatalog`: saved-capture/result-page/tail-selection metadata and named capture/custody clocks. Optional origin and existing exact filters; metadata only. |
| `GET /retained-evidence/{id}/content` | `EvidenceContent`: exactly one capture/tail manifest plus exact `manifestJSON`, and tail recorded custody time. Optional bounded `awaitToken` read lower bound. |

Legacy `GET /retained-evidence` and `GET /retained-evidence/{id}` keep their v1
shapes. Review/correction and exact document validation keep existing meaning wire
shapes. The new response union is reflected and constrained in OpenAPI; actual
handler observations cover both origins and reject complete mixed-origin values.
Content reads distinguish definite absence (404), unavailable/corrupt store (502)
and unconfigured optional store (503). No telemetry read or publication occurs.
The browser catalog/inspector/citation adapter now uses these origin-aware reads.
The subsequent tail-preservation contract below now issues SSE membership
capabilities and registers references-only publication. See
[the origin boundaries](TAIL_SELECTION_CUSTODY.md#public-read-and-presentation-adapters).

Activity preservation uses `POST /retained-evidence/tail` with a typed
`TailIntent` (version, requestID, connection, selector, token/key refs, tags, note).
No client rows, read clocks or windows are accepted. `/tail` reset includes a
random connection ID; batch `membershipState` is available/unavailable/empty, and
only available batches carry token/expiry/basis metadata. Expiry is anchored to
the original read, not browser selection. Existing page retention stays separate.
201 returns the original typed tail receipt on both commit and unchanged retry;
409 rejects changed intent; 410 means definitely uncommitted proof is unavailable;
502 retains retryable storage uncertainty; 503 means optional XTDB is unconfigured.
Committed recovery precedes membership lookup and never rereads telemetry.

## Current native publication conflict boundary

`POST /measurement-packages` validates a duplicate's existing regular-file winner
through a bounded 32 KiB read, package validation and expected revision identity.
An unreadable, invalid, oversized or wrong-revision winner is refused with the
existing 400 convention; it is not overwritten, repaired or reported successful.
Valid duplicate success remains 201, including noncanonical YAML that validates
to the same definition/revision; the existing file is not rewritten. File-only
sync remains unchanged. This is the delivered
M2 correction (`49b31c2`), not a claim of power-loss durability or universal
conflict policy. See [publication contracts](IMMUTABLE_PUBLICATION.md).

## Subsequent Activity-view contract

Reusable [Activity views](ACTIVITY_VIEWS.md) use a typed rolling configuration in
workbench's existing immutable file publisher, separate from fixed exploration
views. Shared revision-library display appears in Saved and Services. The stream
adapter restores configuration through `GooTail.stage`; Start alone opens SSE.
No observations or capabilities cross the configuration boundary. The linked
flow diagram documents publication, staging and later separate retention.

Activity-view browsing adopts `GooDefinitionLibrary` with an explicit `activity`
family and exact revisions, keeping same-title scopes distinct. The existing
record-column picker projects both streamed and staged layouts. Configuration
reads, clipboard sharing and library paging stay in `GooActivityViews`; scope and
draft lifetime stay in `GooStreamEntry`. `GooTail.stage` prepares local display
without a subscription. No display component fetches or starts telemetry.

`POST /activity-views`, `GET /activity-views/{id}` and `GET /activity-views` expose
immutable version-1 rolling selector and column configuration. Publication has a
16 KiB Go/proxy bound and rejects execution/custody fields. List supports exact
optional environment and cursor parameters, returning up to 20 items. These
source-free file routes are independent of fixed `/exploration-views` and XTDB.
Opening configuration never opens a tail connection. See [Activity views](ACTIVITY_VIEWS.md).


## Typed numeric predicate leaf

Record predicates accept the optional `numeric` binding leaf described in
[typed numeric filters](TYPED_NUMERIC_FILTERS.md). It is exclusive with `field`,
`all` and `any`; the existing question-level numeric measure remains independent.
Go validates declared type/unit/parsing and exact decimal-string thresholds before
SQL lowering. State operators partition missing/empty/invalid/valid values;
comparisons, including ne, match valid values only. The feature adds no route and changes no existing omitted-field bytes. Shared
editor/protocol and real telemetry capture/reuse acceptance are documented in
the linked TF2/TF3 contract.

The source-free `POST /numeric-histograms/selection` planner derives exact integer
range and remainder questions from half-open bin indices. Its `/api` proxy has
the same contract. See [integer distributions](NUMERIC_HISTOGRAMS.md).

Numeric measure comparison is an explicit `numeric: true` mode on `POST /compare`.
Both caller-held cohorts must include complete numeric totals and group summaries
with the same declared binding. Exact integer/rational changes are additive under
`measures.numeric`; no source access occurs in this endpoint. Omitting the flag
preserves legacy count-only comparison behavior. The histogram selection planner
also supplies a separate `numericComparison` query pair alongside its unchanged
count-only pair. Numeric stored-evidence decisions use algorithm v4; older frozen
receipt algorithms remain readable without reevaluation. See
[numeric comparison semantics](NUMERIC_COMPARISONS.md).

Aligned distribution comparison is explicit `distribution: true` with numeric
mode and both validated histogram payloads. `POST /numeric-histograms/comparison-selection`
accepts the comparison, side and half-open bin indices, plans without source reads
and permits 1 MiB request bodies. Stored capture comparison accepts the explicit
`distribution=true` query option; publication freezes v5 intent, preserving old
v1–v4 results. See [distribution comparisons](DISTRIBUTION_COMPARISONS.md).

## Fixed-scale decimal numeric bindings

[Decimal64](FIXED_SCALE_DECIMALS.md) extends existing numeric binding objects
with optional integer `scale`. Decimal bindings require scale 1–9; int64/uint64
bindings require zero (normally omitted). Typed thresholds are human literals
with exactly that many fractional digits. There are no new endpoints.

Summary `sum`, `min` and `max`, histogram bounds and comparison differences
remain exact coefficient strings. Decimal summaries include matching `scale`;
comparison bindings and histogram questions carry the interpretation. Mean in
input units is `sum / (valid × 10^scale)`. Scale participates in binding identity,
saved definitions and immutable evidence. No implicit rescaling is performed.
Capability planning reports unsupported scale as diagnostics; execution rejects
invalid bindings. A fractional or string-valued JSON scale is malformed transport.
Older integer questions/results omit scale and retain their wire representation.

## Contextual numeric continuation

[Plot values](CONTEXTUAL_NUMERIC_PLOTS.md) adds no API operation or wire field.
The client stages the existing numeric binding/histogram question with original
membership/window/edge predicates, reset pagination and replaced presentation
measures. Existing explore, selection, comparison, capture and reusable-package
operations retain their contracts. Dialog opening/editing/confirmation performs
no telemetry request; saved-origin Run is a fresh read of the retained scope.

## Versioned log-body pattern field

`column:body_pattern_v1` is log-only and always present, including explicit
unclassified `lp1:e/o/u` buckets. Exact eq/ne predicates require a canonical
profile identity, at most 518 bytes; contains is refused. Display labels are
not predicate values. Existing grouping, time series, comparisons, captures
and YAML carry this field without new operations or wire-shape changes.
Unreferenced row SQL and older captures remain unchanged. See the
[profile contract](LOG_PATTERNS.md) for byte limits and immutable semantics.

Caller-held `/api/compare` pattern groups also require canonical profile IDs,
including on empty or changed-read branches. This validation does not authenticate
caller-supplied source metadata. [Time comparisons](LOG_PATTERN_TIME_COMPARISONS.md)
stage ordinary fixed scopes and reuse existing explore/compare/capture operations.
Counts, per-second rates and shares are display calculations over complete
validated groups, with each side's original window and population denominator.
Saved baseline scope selection does not reuse saved totals; explicit Run rereads
both scopes. Reopening a captured comparison uses stored aggregates and clocks.

## Trace page display projection

[Observed page timing](TRACE_PAGE_TIMING.md) adds no HTTP operation, query flag or
capture field. `exploration.AnalyzeTracePage` defines exact direct-child interval
accounting over public row fields `timestamp_ns`/`duration_ns`; the passive client
projection consumes existing source or stored rows. It does not augment saved
capture bytes or substitute bounded page sums for the independent query aggregate.
Row selection retains the existing linked-log action; disclosure/group display
does not issue telemetry requests.


## Trace search across span conditions

`POST /trace-search` (browser `/api/trace-search`) reads one full scoped source
aggregation for 1–4 span predicates. Its `TraceSearchRequest` declares version
`goo/trace-search/v1`, one environment, a fixed half-open window of at most two
hours and a trace-page limit of 1–100. The request body ceiling is 64 KiB at
both HTTP boundaries. Conditions use the existing typed span predicate tree,
with at most 64 nodes across the request and a 256 KiB compiled-SQL ceiling.
The HTTP boundary compiles and validates before any source effect; oversize
plans return 400 with an instruction to simplify the conditions.

A trace ID qualifies when each condition has at least one matching observed
record in that scope. One record may satisfy several conditions. Conditions
do not require distinct spans, a reported parent edge, matching services or a
complete trace. Canonical nonzero lowercase 32-hex trace IDs define the grouping
identity; the result is a count of observed reported IDs, not requests or
unique invocations. A contextual original record predicate belongs to the
first condition, not a global filter restricting all other conditions.

`TraceSearchResult` returns the exact same-query `matchedTraces` count, bounded
items, a `hasMore` flag and a representative witness for each condition. Witnesses
include full matching-record counts and qualified predicate field values; missing
keys remain null instead of empty strings. Numeric values, timestamps and
counts retain string encoding. Witness selection does not establish a unique
occurrence, root span, causality or telemetry completeness. The source and full
JSON response are bounded to 8 MiB; overflow refuses the entire result.

The route shares source-query admission and optional `X-Goo-Query-ID`
cancellation. It has a 55-second handler deadline under the existing 60-second
query owner. Validation failures return 400; source/response-bound failures
return 502. No result-retention token, capture publication or paired-evidence
format is introduced. Opening a trace from a result is a separate explicit
existing record-context read. See [the semantic contract](TRACE_SEARCH.md).


## Automatic service dashboard families

`POST /dashboards/plan` remains a pure planning operation. It accepts optional
`DashboardRequest.profile`: `auto`, `all-spans` or `http-server`. Omitting it
(including its existing empty/null decoding) preserves the legacy v2 response
without new profile/family fields. Existing request ceilings are unchanged: Go
256 KiB and the browser proxy 16 KiB. No source execution or new route is added.

Opt-in plans add selected `profile`, `families` descriptors and optional widget
`family` identities. A descriptor carries key, title, state, grain, selected
state and optional reason/rule-set identity. HTTP metadata is explicitly a
**candidate**, not proof that the service has HTTP server observations, follows
a particular semantic-convention version or is healthy. `auto` selects HTTP only
with both current span-key metadata and an advisory positive `httpServerObserved`
hint, or a compatible native HTTP instrument. The hint comes from already-held
same-scope observations; false/omitted means no positive hint, not absence. Flat
keys alone leave All spans selected while offering the HTTP candidate. An
explicit unsupported HTTP profile returns 400 instead of silently changing intent.

HTTP span questions retain service/environment/namespace/window and require exact
physical `SpanKind = Server` plus a present nonempty current
`attribute:http.request.method`. Legacy `http.method` is not silently unioned.
HTTP keys observed only on client spans can suggest the candidate but do not
automatically admit it without a positive server hint or compatible native
instrument. An explicitly selected HTTP query can return known empty; the UI
offers an explicit All spans switch without a hidden fallback query. Recorded span status, HTTP status and
business outcome stay distinct. Counts are span observations, not deduplicated
requests. Related native instruments keep separate stream identity, units and
evidence bases. See [automatic dashboards](AUTOMATIC_DASHBOARDS.md).


Incoming RPC uses the same planning route and additive v2 contract. The optional
`rpc-server` profile is distinct from HTTP. For this supported legacy RPC family,
span questions require exact `SpanKind = Server` and present, nonempty
`rpc.system` **and** `rpc.method` on each matching row. Flat keys offer a candidate;
`rpcServerObserved` is an advisory positive hint from already-held same-scope rows,
not an authenticated observation or a convention-compliance assertion. Automatic
selection prefers eligible HTTP, then eligible RPC, then All spans. Explicit RPC
without eligible metadata returns 400. The previously unsupported `rpc-server`
request now reports missing RPC metadata instead of an unsupported-profile error.
All omitted-profile responses and existing successful HTTP plans retain their
wire bytes; an RPC descriptor is added only when RPC metadata is supplied or RPC
is explicitly requested.

Only the supported legacy `rpc.server.duration` native histogram in milliseconds
can independently admit this RPC family. Newer duration/identity conventions are
not silently aliased or mixed with it. Span observations and native metrics retain
separate populations, clocks and units. No request deduplication, RPC status-to-
business-outcome rule, source probe or fallback read is inferred. Profile and hint
additions do not alter route counts, byte limits or publication effects.


Database dependencies use the optional `db-client` dashboard profile and advisory
`dbClientObserved` boolean. Current `db.system.name` metadata offers a candidate;
a positive same-scope Client-row hint admits automatic selection after eligible
HTTP and RPC. A compatible native `db.client.operation.duration` histogram in
seconds can independently admit the family. Explicit DB without a candidate
returns 400. Flat keys alone and false/omitted hints do not prove observations or
absence. A DB descriptor is added only for relevant metadata or explicit choice,
preserving unrelated successful plans and omitted-profile responses.

DB span questions retain service/environment/namespace/window and require exact
`SpanKind = Client` plus a present nonempty current `db.system.name`. Legacy
`db.system` is not an alias. These are observed client dependencies and recorded
span statuses, not database-server health, database completeness or a deduplicated
operation count. Native and span duration remain separate populations with
independent units, intervals and evidence. No new route, source probe, hidden
fallback, annotation requirement or publication is introduced.


Messaging adds separate optional `messaging-consumer` and `messaging-producer`
profiles with advisory `messagingConsumerObserved` and
`messagingProducerObserved` booleans. Current `messaging.system` metadata offers
both candidates; positive same-scope rows admit the corresponding role. Automatic
precedence is HTTP, RPC, DB client, consumer, producer, then All spans. False or
omitted hints mean unknown. Explicit selection needs eligible metadata; no source
probe or fallback read is performed. Unrelated and omitted-profile plans retain
their existing wire bytes.

Consumer span questions require exact `SpanKind = Consumer`; producer questions
require exact `SpanKind = Producer`. Both require present nonempty
`messaging.system`, preserving service/environment/namespace/window. Destination
or operation metadata alone does not admit either role. Optional dimensions use
current destination name, operation name, then system. The profile describes role
span observations, not unique messages, confirmed acknowledgements, delivery
success, queue depth, lag or broker health. Processing metrics cannot establish
those facts or classify producer spans.

Only a compatible native `messaging.process.duration` histogram in seconds can
independently admit the consumer family. Its processing-duration population,
identity, intervals and units remain separate from Consumer span observations;
metric shape does not prove matching operation values or convention compliance.
No new routes, transport budgets or publication effects are introduced.

`GET /capabilities` (browser `GET /api/capabilities`) reports source-free Go
startup metadata: `{version:"goo/runtime-capabilities/v1", profile:"lab"|"generic",
preparedHTTP:{state:"configured"|"unconfigured"},
declaredOperations:{state:"configured"|"unconfigured"}}`. No query parameters are
accepted. Configured means the optional startup inputs were loaded; it does not
establish ClickHouse reachability, ingestion, telemetry coverage, readiness,
authorization or application health. Generic startup requires neither prepared
HTTP binding/inventory nor operation manifests. Explicitly supplied optional
configuration is still validated; the lab profile retains its existing defaults.

Absent prepared HTTP returns JSON 503 from `/http-package`; absent declared
operations returns JSON 503 from `/operation-packages`, `/operations`,
`/counter-rates` and `/histograms`. These responses contain `error`,
`code:"capability_unconfigured"` and `capability:"prepared-http"` or
`"declared-operations"`, before any source read. Existing configured route
responses retain their previous wire behavior. Generic record exploration,
discovery and supplied-data planning do not require those optional capabilities.

Python `GET /api/ui-config` separately returns pure startup configuration:
`{version:"goo/ui-config/v1",profile:"lab"|"generic",
examples:{state:"configured"|"unconfigured"}}`. It does not contact Go or
ClickHouse. `/api/capabilities` transparently forwards Go's capability response;
UI configuration and Go configuration describe their own processes. In generic
mode the optional lab routes `/api/package`, `/api/query`, `/api/explain`,
`/api/example-questions` and `/api/prebuilts` return JSON 503 with the same code
and `capability:"example-grammar"` or `"example-catalog"`. This is intentional
unconfigured functionality, not a claim that the telemetry source failed.

`POST /explore/counted-values` (browser `/api/explore/counted-values`) accepts the
existing strict `RecordQuery` shape, restricted to exactly one string `groupBy`,
zero offset and no measure, histogram, summary, bucket, time-series or span-status
projections. Numeric predicates and reported-edge scope remain supported. Go and
proxy accept at most 16 KiB; Go uses the same 55-second timeout, admission budget
and `X-Goo-Query-ID` cancellation owner as `/explore`.

The `goo/counted-values/v1` response contains the original question, total count,
complete groups (`g0: null|string`, positive string count), capture time and a
content basis, plus separate count/groups read clocks and bases. Exactly two
bounded aggregate reads run: count then groups. No record page, source rows,
retention token or capture is produced. Counts remain canonical decimal strings;
missing and present-empty group values remain distinct. More than 1,000 groups
is refused rather than partially represented. Source failures return 502;
unsupported requests return 400 before source execution.

`observed` and `no_data` describe consistent observed aggregates.
`changed_between_reads` retains diagnostic results when group counts differ from
the separately observed total; it is not a population suitable for actions or a
transactional snapshot. Read times are provenance, not a completeness guarantee.
See [counted values](COUNTED_VALUES.md) for interpretation and UI behavior.

`POST /record-aggregates` (browser `/api/record-aggregates`) projects existing
record questions into aggregate-only service-dashboard results. It uses strict
`RecordQuery` decoding, a 16 KiB request ceiling on both transports, a 55-second
enclosing timeout and the shared query admission/cancellation owner. Requests
require exactly one string group and zero offset. Optional `span-duration`,
recorded span status and fixed time buckets are supported; numeric measures,
numeric histograms and grouped time series are refused. Numeric predicates and
reported-edge scope remain available. Invalid requests return 400 before reading
the source.

The `goo/record-aggregates/v1` response contains question/count/state/groups,
optional duration/status/time-series summaries, capture time, content basis,
warnings and separate read receipts. It reads count and complete groups, plus
exactly one time-series aggregate when requested. It never reads or returns a
record page, discovered fields, query plan, pagination state or retention token.
More than 1,000 groups is refused. Source or decoding failures return 502,
including failures after an earlier aggregate read; partial successful data is
not substituted.

Each summary retains its existing record grain and exact string-count/nanosecond
contract. Separate reads can yield `changed_between_reads`; those results retain
diagnostics but do not support conserved-population actions. Equal aggregates do
not prove a transactional snapshot. Empty time buckets are no observations in
that aggregate read, not evidence of service health. Existing `/explore` and
counted-values responses are unchanged.

Native measurements now accept additive `definition.kind: stock` through the
existing package, `/measurements` and capture routes. Stock binds actual sum-table
observations, requiring cumulative temporality and nonmonotonic input. Delta and
monotonic source sums return 502 rather than being presented as an absolute stock;
invalid package definitions return 400 before source execution. There is no new
route or change to existing transport/cancellation limits.

A stock result adds optional `stocks: {fromNS,toNS,series}`. Each exact stream has
state and samples containing accumulation `startNS`, observation `atNS`, state,
optional numeric value and optional reason. Signed finite fractional tallies are
supported. A decrease is an observation, not a counter reset; a new epoch is
retained without subtraction or interpolation. `observed`, `stale` and `unknown`
remain distinct. Unknown samples omit value; stale samples retain it. This is
neither a gauge-table reinterpretation nor a counter rate, fleet total, verified
physical baseline, pool health or utilization ratio.

Captured stocks reopen through the existing saved-capture API without telemetry
reads and preserve the original source rows, exact identity, epochs and computed
stock result. Existing gauge/counter/histogram/lifecycle result bytes are unchanged
when `stocks` is absent. The actual handler workflow can be exported with
`GOO_STOCK_API_OBSERVATIONS` while running `TestStockHTTPExecuteCaptureAndOfflineReopen`
and checked using `scripts/check_openapi.py --stock-observations <path>`; the
workflow covers execute, capture, offline reopen and delta/monotonic refusals.
See [cumulative stock semantics](CUMULATIVE_STOCKS.md).

`definition.stockDomain: nonnegative-integer` optionally narrows a stock package's
value domain for named-pool connection observations. It is valid only with
`kind: stock`; invalid kind/policy combinations return 400 before source access.
For opted-in, recorded samples the evaluator requires an integer >= 0 and < 2^53;
negative, fractional or unsafe-integer source values return 502. Missing-value
flags remain unknown and do not become observed values. Omitting the policy
retains generic signed finite fractional stock semantics and existing wire bytes.

The policy is carried in the planned/executed/captured question and participates
in package identity. Capture/reopen preserves it and the original native evidence
without rereading telemetry. It is a domain constraint, not proof of pool health,
capacity, utilization, or semantic-convention conformance. The named-pool dashboard
may derive a domain-qualified question from an eligible observation; no route or
transport limit changes. API schemas apply the bounded integer value contract
only when the retained question opts in. Actual execute/capture/offline-reopen,
negative/fractional/unsafe refusals, the accepted maximum safe integer and missing
flags can be exported with `GOO_STOCK_DOMAIN_API_OBSERVATIONS` while running
`TestStockDomainHTTPExecuteCaptureAndOfflineReopen`, then checked with
`scripts/check_openapi.py --stock-domain-observations <path>`.

`POST /pool-signals/plan` (browser `/api/pool-signals/plan`) is pure planning over
`{parent: NativeQuestion, stream: Stream, instruments: Instrument[]}`. Both
transports enforce 64 KiB. Go checks cancellation/deadline before and after the
bounded synchronous planner under a two-second context; this does not preempt
computation in a goroutine. Validation and elapsed/cancelled requests return 400.
It occupies no source-query slot, performs no telemetry/custody read or publication,
and does not treat a supplied query ID as authorization or a source capability.

The parent must be a domain-qualified, schema-pinned connection-count question
and the held stream must satisfy its exact binding. Candidates match the full
shared resource/resource-schema and instrumentation identity, with count stream
attributes minus only connection state. Pool, instance and state must be reported;
no alias or dimension broadening is inferred. Capacity candidates are cumulative
nonmonotonic connection-max stocks with the bounded nonnegative-integer domain;
wait candidates are wait-time histograms in seconds. Options preserve the exact
parent bounds and can run through existing measurement/capture routes.

`goo/pool-signals/v1` returns detached parent/stream, capacity/wait options
`{key,title,question}`, state `candidates|unconfirmed`, omission counts and warnings.
At most 16 instruments and the first 25 preview streams per instrument are
considered, with at most two distinct options per signal. Preview omissions or
no matches do not prove absence. Capacity and wait remain independent evidence
with separate clocks; matching labels do not prove alignment, utilization,
queue health or causality. Actual execute/capture/offline reopen of count, capacity
and wait can be exported with `GOO_POOL_SIGNALS_API_OBSERVATIONS` while running
`TestPoolSignalsHTTPExecuteCaptureOfflineReopen`, and checked with
`scripts/check_openapi.py --pool-signals-observations <path>`.

Pool-signal preview omission counters must be within 0..10,000 per supplied
instrument, including extra preview rows removed by the 25-stream cap. Invalid
or overflowing supplied counters are refused with 400, preserving meaningful
omission disclosures; the bounded total cannot exceed 160,000.

`POST /pool-signals/view` (browser `/api/pool-signals/view`) composes held native
results using `{count: NativeResult, stream: Stream, capacity?: NativeResult,
wait?: NativeResult}`. Both transports enforce 2 MiB. Go checks a two-second
context before and after bounded synchronous composition; cancellation, elapsed
deadlines, malformed input and invalid composition return 400. This pure route
uses no source admission/cancellation token, performs no source or custody I/O,
and publishes no capture or combined receipt.

Every supplied result is replay-validated against its retained raw rows, question,
basis and derived values. That establishes internal evidence consistency, not
source authenticity or authorization. The anchor must exist in the held count
result. Optional signals require their exact pinned identities, the same full
pool/resource/instrumentation identity and identical fixed evaluation bounds.
Wrong identity/window and forged bases or evaluated values are refused instead
of silently repaired or retargeted.

`goo/pool-view/v1` returns separate count/capacity/wait questions and evidence
clocks/bases, reported count states, optional latest capacity/wait observations,
warnings and utilization. Utilization is a rational pair of decimal strings only
when observed bounded integer used/capacity values have the same observation time
and a matching nonzero accumulation epoch, with positive capacity and used <=
capacity. Unavailable or inconsistent results carry a reason and omit the ratio;
raw observations remain visible. A zero start epoch means unreported epoch, even
when both are zero. Matching reported clocks/epochs do not establish a
transactional snapshot, aligned collection, health or causal relation. Existing
count, capacity and wait captures remain independent and unchanged.

Host discovery and host measurements use a distinct resource binding without
manufactured service identity. Native `source.resource:{kind:"host",id:<host.id>}`
requires an exact stream, no service selectors/namespace and per-stream gauge or
stock semantics. The full raw resource map and resource schema remain exact;
missing versus present-empty `service.name` is preserved when actually reported.
`host.name` is display metadata, not an identity fallback. Existing service-bound
measurement validation and saved bytes remain unchanged.

`POST /host-discovery` (browser `/api/host-discovery`) accepts
`{environment,fromMs,toMs}` with a 16 KiB body limit, native source admission/query
cancellation and a 15-second context. One globally bounded UNION read inventories
host-bearing raw metadata across gauge/sum/histogram tables without consulting or
augmenting the exporter ServiceName column. More than 10,000 observations fails
closed; no partial catalog is published. Resource identities are limited to
16 KiB and the complete response to 8 MiB. `goo/host-discovery/v1` returns question,
up to 100 full-map/schema resource entries with opaque keys, records and explicit
host/instrument omissions. Each entry retains at most 16 instruments with 25
preview streams apiece and their omissions. Reported host IDs shared by different
full maps remain separate resource identities, not a proven physical-host count.

Unidentified source rows are not queried, so `missingHostIDRecords` describes only
unqualified rows observed by the bounded decoder, not an inventory-wide missing-ID
count or absence claim. Metadata and coverage descriptors do not establish host
health, value freshness, ingestion completeness or physical relationships.
Host execution currently supports exact gauges and cumulative nonmonotonic stocks;
other families, including CPU seconds counters, remain unsupported metadata.
An instrument's definition binds its first exact preview only, never an aggregate
across the instrument's other streams.

`POST /host-dashboard/plan` (browser `/api/host-dashboard/plan`) accepts
`{question:HostDiscoveryQuestion,host:HostEntry}` under a 64 KiB body limit and
checks a two-second context around bounded synchronous pure planning. It occupies
no source-query slot and performs no source/custody reads or publication. The
`goo/host-dashboard/v1` plan includes at most two default `widgets`, up to 16
explicit `candidates` with exact native questions, state and separate
`candidatesOmitted`, `streamsOmitted`, `instrumentsOmitted` counters. Large retained
metadata that exceeds the request bound is refused rather than silently clipped.

Defaults use the pinned v1.37 CPU-utilization gauge in unit 1 (user mode preferred,
then canonical numeric core order; opaque core labels retain literal order) and
memory-usage cumulative nonmonotonic stock in By (used state preferred, bounded
nonnegative integer domain). A provided memory-utilization gauge in unit 1 can
serve as a fallback; no utilization ratio or capacity is derived. Every logical
CPU, mode, memory state and resource remains a separate stream. A core view is
not a whole-host or fleet aggregate. Remaining candidates run only by explicit
selection through existing measurement/capture APIs.

The actual host-only inventory/plan/gauge/stock/capture/offline-reopen workflow is
exported by `GOO_HOST_API_OBSERVATIONS` with `TestHostHTTPNativeExecuteCaptureOfflineReopen`
and checked using `scripts/check_openapi.py --host-observations <path>`.


Fresh host YAML package authoring canonicalizes an omitted/null services list to
`[]` and omitted/null exact-stream attributes to `{}` before revision hashing.
This prevents YAML publication/reload from changing the package identity. It does
not normalize direct JSON measurement questions, evaluated evidence or existing
captures, and service package serialization remains unchanged. Actual publication,
duplicate/winner verification and fresh-catalog checks cover all nine host
services/attributes empty-shape combinations.

Go runtime observations reuse `/dashboards/plan`, `/measurements`, existing
native selection and capture APIs; no route or record-profile change is added.
Relevant native widgets use `family: go-runtime`. The pinned v1.37 candidates are
`go.memory.used` and `go.memory.gc.goal` cumulative nonmonotonic stocks in By, and
`go.goroutine.count` in `{goroutine}`. Each derived question carries an exact
preview stream, service/environment/namespace identity, a reported nonempty
instance and the bounded nonnegative-integer stock domain. Metadata is not a
semantic-convention compliance certificate.

Default selection prefers one memory component and one GC heap-goal observation
from the same full resource/resource-schema/instrumentation identity when
available, within the existing two-native-read budget and established role/pool
priorities. Goroutines require explicit selection. Components, instances and read
clocks remain independent: memory used is not memory pressure, heap goal is not
capacity, utilization or GC pause/activity, and a goroutine count is not a leak
diagnosis. Recognized runtime metrics without qualifying preview evidence remain
explicit generic options and do not automatically trigger a broad fallback read.

Runtime preview input is bounded to the first 25 streams per instrument.
`runtimeStreamsOmitted,omitempty` reports clipped preview rows plus bounded
supplied discovery omissions, not a proven count of executable choices or source
absence. `nativeOptionsOmitted,omitempty` separately reports supported options
excluded by the global 128-option display cap. Negative or out-of-bound runtime
preview counters are refused. Exact option keys are canonicalized from stream and
policy identity, so input ordering or arbitrary discovery package names do not
multiply equivalent physical/policy choices. Both new fields are absent when
zero, preserving unrelated response bytes. Old non-runtime fixtures remain
unchanged.

`POST /pod-discovery` (browser `/api/pod-discovery`) accepts
`{environment,fromMs,toMs}` with strict 16 KiB decoding, a fixed window of at most
120 minutes, a 30-second enclosing context and the existing source-query
admission/cancellation owner. It requires no Kubernetes credentials or external
cluster access. One read groups pod-bearing raw telemetry resource maps and
resource schemas across logs, spans, gauges, sums and histograms. It never
augments maps from the exporter ServiceName column.

`goo/pod-discovery/v1` returns question, `sourceGroups`, up to 100 `observations`,
`observationsOmitted`, read clock, content basis and warning. The whole request
fails when its 1,001st source-family group is observed, rather than publishing a
partial inventory. Identities are bounded to 64 labels/16 KiB, with an 8 MiB
whole-response limit. A source-family group is not one pod or one record:
`families[].count` is a canonical positive decimal string counting that family's
captured records or metric points; `firstNS`/`lastNS` are exact timestamp strings.
The basis includes canonical grouped observations even when their display
variants are omitted. One query does not establish a transactional or complete
cross-signal snapshot.

Each resource observation retains key, raw resource map, nullable
`resourceSchemaURL`, pod/cluster UID states, contextual `actionable` flag and
family observations. The pinned deployed trace table does not record resource
schema, represented explicitly as null; logs and metric tables record strings,
including known empty. These remain distinct identities. Missing and present-empty
UIDs remain distinct metadata diagnostics. Only a nonempty reported pod UID can
enable contextual record actions; a missing cluster UID remains unverified
context. Equal names or UIDs across different resource/schema variants do not
prove physical identity, ownership or Kubernetes API state.

Record exploration uses existing staged questions and source reads, retaining
present attributes and explicit missing cluster UID/namespace where applicable.
This is contextual matching, not full map/schema equality, every pod's telemetry,
a causal join, native metric authority, health or capacity. No new measurement,
record-query operator, capture format or configuration capability is introduced.
API handler fixtures parse the actual read clock and normalize only that clock
for frozen comparison; content bases remain unchanged. The exact generated SQL
was also verified read-only against the local five-table ClickHouse schema using
an empty environment/window, including the absent trace-schema-column case.

Pod-native metrics reuse measurement execution, capture and package publication.
`POST /pod-metric-discovery` accepts an explicit environment/fixed window, complete
raw `resource` map and required non-null `resourceSchemaURL` string. Known empty
schema is valid; an unknown trace schema is refused before source access. One
bounded UNION over gauge, sum and histogram metadata matches the exact map/schema
without manufacturing service identity. More than 10,000 rows refuses the entire
response; at most 16 instruments and 25 stream previews per instrument are
retained with explicit omissions. Metadata is not pod health or a complete pod
inventory.

`POST /pod-dashboard/plan` accepts `{discovery: <held discovery response>}` and
performs pure bounded planning. The initial Collector kubeletstats v0.137.0
profile requires its exact receiver scope/version and empty scope schema, and
exposes memory usage/working-set `By` gauges. The separate OTel v1.37 profile
exposes usage only. No aliases, ordinal state interpretation, container sums,
capacity ratios or ownership are inferred. Plans contain at most two defaults
and 16 explicit candidates, each a gauge question with resource kind
`kubernetes-pod`, exact reported UID and mandatory exact stream. No service
fallback applies.

Discovery has a strict 64 KiB request limit. The pure planner accepts up to
8 MiB + 1 KiB (8,389,632 bytes), allowing the bounded discovery response plus
its JSON wrapper; semantic instrument/stream limits remain unchanged. Discovery uses the existing shared
source admission/query-ID cancellation owner and a 30-second deadline; planning
uses a two-second context around synchronous bounded computation and does not
reserve source admission or create custody. Execute a returned question through
`/measurements`; use existing capture and package APIs for custody/authoring.
Fresh YAML authoring normalizes empty service/attribute containers for stable
revisions; direct JSON questions and captured raw evidence retain their shapes.

Acceptance covers actual HTTP discovery/planning, two native gauge executions,
capture and exact offline reopening, plus duplicate publication/fresh catalog
reopening for omitted/null/empty services and attributes. Exact generated
resource-map/schema SQL also executes read-only against the local ClickHouse
schema. This syntax check is distinct from receiver ingestion evidence.


Native measurement execution and capture also support [typed pod phases](POD_STATES.md)
via `kind: state` and the pinned `stateProfile`; responses carry `states` with
independent observation quality and domain decoding. No route is added. Generic
capture comparison explicitly declines categorical state arithmetic. This is
also used by exact phase discovery and categorical UI presentation.


The [native container readiness prerequisite](CONTAINER_READINESS.md) adds
`source.resource.kind: kubernetes-container` with the reported container.id and
`stateProfile: k8scluster-container-ready/0.137.0`. Existing measurement/capture
routes return Ready/Not ready categorical samples; code 0 stays observed rather
than becoming missing. Existing `/pod-metric-discovery` and `/pod-dashboard/plan`
now recognize exact container-bearing variants. Readiness widgets use
`container-resource-stream` grain and container authority; phase and memory retain
pod grain. No new endpoint or comparison format is added.

`POST /container-memory/view` is a pure 8 MiB-bounded composition of `parent`
(known-schema exact container resource context) and optional `workingSet`,
`request`, `limit` native gauge results. Every result is replay-validated and
retains its exact source question, read clock and basis. Fixed three-row output
separates byte-domain display state from optional original `observationState`;
valid bytes and certified percentages are exact strings, with rounded percentage
text supplied separately. It performs no source read or custody publication.
Existing pod metadata/plan routes expose pinned container memory role candidates
only; they never add automatic widgets. See [contract](CONTAINER_MEMORY.md).


Milestone 207 adds `POST /container-cpu/view`, a pure 8 MiB-bounded
composition of `parent` and optional `usage`, `request`, `limit` native gauge
results. It retains independent evidence and applies fractional-core domain rules
before strict alignment checks. CPU usage is the pinned reported gauge, not a
derivative of cumulative CPU time. See [contract](CONTAINER_CPU.md) for scope and
accepted evidence; automatic widget and individual capture budgets are unchanged.


`POST /container-resources/history` is a pure 8 MiB-bounded projection of held
`parent` plus optional `memory` and/or `cpu` result groups. It replay-validates
original evidence and certifies every raw timestamp group, retaining original
quality/reason and raw-row indices separately from numeric domain admission.
Stable family/role series retain original questions, clocks, bases and freshness
policy. More than 10,000 total points or 8 MiB output refuses the whole request.
No source admission/read, custody publication, interpolation or percentage
history is added. Existing CPU/memory view response bytes remain unchanged.
See [history contract](CONTAINER_RESOURCE_HISTORY.md).


### Explicit decimal counters (211)

`Definition.counterDomain: nonnegative-decimal` opts a per-stream counter into
exact exported decimal subtraction; omission retains integer behavior. Its
`CounterResult.valueDomain` is explicit, and known intervals add rational
`exactPerSecond` beside exact decimal `delta` and approximate finite `perSecond`.
Unknown intervals carry no arithmetic. Decimal results refuse above 8 MiB, with
1,024-byte deltas and 2,048-byte exact-rate schema bounds. Count/fleet and whole
window comparison contracts remain separate. The first default is pinned host
`system.cpu.time` cumulative monotonic seconds. Nine real mux observations cover
execution, capture/offline reopen and five source/profile refusals; validate with
`check_openapi.py --cpu-time-observations PATH`. See [contract](CPU_TIME_RATES.md).


### Explicit host network profile (212)

No operation or wire envelope is added. Retained host plans may offer pinned
`system.network.io`/`By` cumulative candidates with reported interface and
receive/transmit attributes; they never enter automatic default widgets. Source
reads certify nonnegative whole byte values below 2^53, preserving unknown flags
and rejecting fractional/unsafe tallies or temporal drift. Existing native
capture/reopen and host interval-context contracts apply. Eleven mux observations
include seven source/profile refusals; `check_openapi.py --network-observations
PATH` validates them beside unchanged frozen fixtures. See [contract](HOST_NETWORK_THROUGHPUT.md).

Disk 213 uses existing host discovery/plan and measurement/capture operations.
The pinned system.disk.io Counter/By cumulative profile requires the explicit
nonnegative-decimal domain and certifies whole counts below 2^53. Discovery
retains system.device/read-write candidates; no new endpoint or wire version.

Filesystem 214 uses existing host discovery/plan, measurement and capture endpoints.
Pinned usage/limit stocks opt into nonnegative-integer bytes below 2^53 and require
strict reported cumulative nonmonotonic source metadata. Used-state versus limit
without state is a conservative catalog policy. No new endpoint, version, computed
percentage or stock-point native-record-context contract is added.

Reported filesystem utilization 215 uses existing host-plan/measurement/capture
operations. Its pinned Gauge/1 binding validates original known identity/maps/flags
and exact exported decimal bounds 0..1; malformed or out-of-range source data
refuses the whole read. Missing-value flags stay unknown. Generic wire schemas
and versions remain unchanged; profile admission remains semantic Go validation.
