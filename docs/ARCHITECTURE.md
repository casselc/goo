# Architecture

This describes implemented module and lifetime boundaries. The explicitly labeled
target diagram is design context, not another execution queue; current priorities
are in the [reviewed roadmap](ROADMAP_2026_10_24.md). See
[known limits](KNOWN_ISSUES.md) and [executed validation](../VALIDATION.md) for
capability/acceptance boundaries. HTTP paths and direct/proxy differences are
indexed in [the API contract](API.md).

[HTTP boundary inventory](HTTP_BOUNDARIES.md) distinguishes pure planning, source
execution, custody/publication, optional XTDB and SSE lifetimes. Native measurement
and record/operation/capture registrations receive configured source or connection
values from one startup snapshot. Query, SSE and held-draft owners retain distinct
lifetimes. Operation GET registration is separate from startup loading; its source
bindings and investigation lineage remain family-owned. Recipes and
dashboard planning share explicit-policy bounded JSON decoding/writing;
semantic validation, status mapping and SSE lifetimes remain in their owners.
Retention routes and document citation checks share an injected optional XTDB URL;
each operation receives its own read basis, and document file loading remains
independent of XTDB. Native measurements now adopt shared POST/auth/context and
bounded-body primitives, while compilation, escaping, status/row guards, evidence
basis and custody remain in the native adapter. Native, record, operation and
prebuilt HTTP mechanics now share bounded `internal/sourcehttp` primitives behind
their distinct adapter contracts. The [HTTP acceptance inventory](HTTP_BOUNDARIES.md#c2-acceptance-audit--2026-10-22)
records adopted transports and remaining endpoint/branch coverage; no common
semantic evaluator or universal request owner is implied.

[Immutable publication](IMMUTABLE_PUBLICATION.md) shares write/sync/link mechanics
across reusable telemetry packages, workbench artifacts, capture stores, lifecycle
custody and native YAML. The latter two retain explicit file-only sync.
Investigation captures retain exclusive collision errors; other adopters accept
existing names and retain their family-specific integrity behavior. Native YAML
POST now validates the existing regular-file winner with a bounded read, package
validation and expected revision; invalid/unavailable winners are refused without
repair or false success (M2, `49b31c2`). Valid noncanonical duplicate winners
still return 201 without rewriting the existing file. Family stores retain encoding, digests,
lineage, catalogs and distinct reload/sync policies.
The documented inventory keeps digest/custody/native YAML differences explicit.

[Native presentation](NATIVE_PRESENTATION.md) shares supplied latest values and coverage across measurements, services and native comparisons. Go evaluation and application read/custody owners remain separate.

[Record population gestures](RECORD_POPULATIONS.md) reuse the attribute viewer and typed selection owner. Shared-field discovery is local string intersection; query evaluation remains in Go. [Shared annotations](WORKBENCH_ANNOTATIONS.md) attach exact-target revision sidecars; optional native retention freezes selected server-owned observations. The broader reviewed-label design remains proposed.


[Population explanations](POPULATION_EXPLANATIONS.md) extend the same contextual
comparison/range workflow. `ui/explanation.js` owns explicit execution, staged
original-side filters and source-free request reuse; supplied result rendering
does not infer causes or execute queries. `exploration/explanation.go` projects
count-only questions, validates complete groups, preserves per-read bases and
ranks exact descriptive share changes. `cmd/goo/population_explanation_http.go`
keeps pure plans/filter plans separate from the controlled aggregate read.
Observed promotion uses the shared bounded `ResultSnapshots` pool and immutable
capture publication: the server-held result is saved without a second source
read. Durable explanation captures hold aggregate facts and reusable questions,
not raw inspection rows or ephemeral result tokens. Final journey 177 acceptance
is recorded separately in VALIDATION.md.

[Categorical selection](FACET_SELECTION.md) constructs bounded typed predicates in the population adapter, then delegates validation/compilation and measurement to Go. Staging and child return are source-free.


[Range drilldowns](RANGE_DRILLDOWNS.md) retain connected parent DOM/owners while
staging an independently namespaced child population. Display components emit
selection intent; the adapter projects checked fixed bounds without source reads.
Only explicit Run executes existing Go questions. Back disposes the child's
requests/descendants and restores the paused parent without reads. There are at
most three child levels; this session state is separate from immutable evidence
and stored view configuration. The linked sequence diagram documents ownership.

[Recent activity](RECENT_ACTIVITY.md) enters the existing Go population planner
directly from Explore. The adapter resolves fresh fixed bounds, runs at most two
typed exploration requests and composes existing charts/facets/record pages;
those requests retain their independent backend read bases. Local signal switching,
view settings and range preview do not query telemetry. Live owners remain explicit.

The [population workspace](POPULATION_WORKSPACE.md) composes existing record queries through a data-only Go planner, reuses aligned timeline/record display adapters, and stores immutable exploration-view configuration in the workbench. Query kernels, captured evidence and saved presentation configuration remain separate.

[Linked timeline architecture](LINKED_TIMELINE.md) composes existing Go widget results into passive shared-time lanes. `goo-time-range` emits bounded selections; the service adapter preserves identity and native definitions while rereading the overview.

The [trace finding flow](TRACE_FINDINGS.md) keeps exact page-example selection in Go, navigation and two sequential capture requests in application adapters, and immutable custody in existing capture/document stores. Passive display elements perform no source reads. The sequence diagram documents compilation/query boundaries and the capture-to-finding handoff.

See the [architecture and process diagrams](ARCHITECTURE_DIAGRAMS.md) for the
source/build path, illustrative production service topology, OTel/ClickStack
capture, investigation sequence and separate Databricks artifact boundary.
The [implementation plan](IMPLEMENTATION_PLAN.md) distinguishes completed slices
from remaining work; [validation](../VALIDATION.md) records actual checks.

## Current implementation

| Component | Source | Responsibility |
| --- | --- | --- |
| Local stack | compose.yaml | Default services: capture/storage, Python sample roles, two woven Go apps, query and Go packages; optional MongoDB/HyperDX |
| OTel pipeline | config/collector.yaml | Receive OTLP logs/traces/metrics, batch, export to ClickHouse baseline tables |
| Sample roles | sample.py | HTTP gateway, unary gRPC acceptance, Redis publisher, consumer attempts, traffic generator |
| Telemetry setup | telemetry.py | Python SDK providers, explicit always-on demo tracing, OTLP/HTTP exporters |
| Prototype query compiler | grammar.py | Validate the original lab request and select bounded SQL templates |
| Query service | server.py | Serve UI/API docs, Python compatibility query/explain and the JSON/SSE browser proxy on 8088; Go routes remain native on 8089 |
| Browser workspace | ui/workspace.js, ui/product.css, ui/index.html | Navigation/layout facade and visual system; Examples contains the original Python reservation workflow |
| Pure question protocol | ui/question-protocol.js | Browser/Node step/link/capture projection and comparison-only native-default equivalence; compatibility exports retain identical function references |
| Activity and retained analysis | ui/stream-entry.js, ui/investigation-trail.js, ui/exploration.js | Visit-gated live entry, explicit lazy fixed-owner mounting, paused-parent return and independent source generations |
| Ordinary draft recovery | ui/investigation-draft.js, ui/documents.js | Bounded same-tab authored checkpoints and explicit source-free Restore/Discard, separate from immutable publication and native retry journals |
| Telemetry package authoring | exploration/package*.go, cmd/goo/telemetry_packages_http.go, ui/telemetry-packages.js | Manifest-free reusable selectors/counts, strict YAML, immutable definition revisions, preview plans and optional annotation handoff |
| Observed service inventory | exploration/inventory.go, ui/inventory.js | Single-read captured service/span-name counts, Error statuses and duration buckets; fixed-window Explore/package navigation; no membership inference |
| General capture custody | exploration/capture_store.go, cmd/goo/captures_http.go, ui/captures.js | Immutable manifest-free evidence, bounded stored row bank, receipt-bound paging, local catalog and fixed-question/plan successor refresh |
| General exploration | exploration/, ui/exploration.js, ui/distributions.js | Arbitrary retained OTel attributes, population counts, distributions, fixed-window cohort comparison and cross-service trace/log context |
| Prototype package data | definitions/reservations.json | Demo deadline, target, permitted views, grouping choices |
| Annotation build/runtime | annotations/, instrumentation/, scripts/build_*_app.sh | Validate selected source contracts, weave opt-in Go builds, emit standard OTel spans/events/metrics |
| Operation package planner | annotations/package.go, config/operations/ | Shared strict YAML bindings, identity, enabled views, counter/histogram contracts and content revisions |
| Go investigation service | cmd/goo/serve_http.go | Serve configured HTTP/operation plans and bounded ClickHouse-backed observations |
| Physical operation sources | annotations/clickhouse.go, annotations/counters.go, annotations/histograms.go | Normalize and validate exporter rows without presentation limits defining totals |
| Native counter kernel | metrics/counter.go | Exact deltas, temporality/epoch/gap handling, unknown intervals and separate evidence bases |
| Fleet coverage | metrics/fleet.go, annotations/fleet.go, ui/fleet.js | Fixed expected membership, bounded peer queries, exact aligned totals or unknown coverage |
| Error-ratio kernel | metrics/ratio.go, annotations/ratios.go | Exact aligned outcome counts, unknown/idle semantics, declared policy and shared counter basis |
| Native histogram kernel | metrics/histogram.go | Exact bucket deltas, unknown boundaries, explicit aligned merges and P95 bucket ranges |
| Shared aggregate/lifecycle kernels | measurement/, lifecycle/ | Prepared-fact evaluation and compiler targets; normalized lifecycle fixtures and receipts |
| Operation browser components | ui/operations.js, ui/visualizations.js, ui/histograms.js, ui/ratios.js | Package-driven overview, raw heatmaps, scoped subject navigation, waterfalls, native histograms and counter/context views |

The sample has no durable business outbox or complete acceptance ledger. It emits application-defined log attributes and can lose telemetry independently of queue acknowledgement. The consumer deliberately creates a new trace with a producer link. gRPC application payloads use a JSON codec; telemetry uses standard protobuf OTLP.

The original Python query backend uses fixed table names and logical assumptions embedded in SQL. Values are passed through ClickHouse typed parameters. The result summary is calculated from the bounded returned rows. The alarm is a preview calculated when a user queries. None of these constitute a general package compiler or operational SLO evaluator.

The Go operation path now separates [YAML package planning](OPERATION_PACKAGES.md),
source normalization and evaluation. [Native counter rates](NATIVE_COUNTERS.md)
use actual metric points; span links from rate intervals are contextual, with a
separate basis. [Subject navigation](SUBJECT_NAVIGATION.md) uses declared scoped
identity, not inferred causality or durable completion. Neither path repairs the
original Python reservation-preview accounting gaps automatically.

## Target boundaries (partly implemented)

```mermaid
flowchart TD
    A["Application package"] --> N["Normalized typed expressions"]
    U["UI session edits"] --> N
    N --> V["Semantic validation"]
    V --> R["Reference evaluator"]
    V --> C["ClickHouse compiler"]
    C --> D["Source adapters and database"]
    D --> E["Typed results and evidence"]
    R --> E
    E --> P["Views and interactions"]
    V --> S["Scheduled evaluations"]
    S --> I["Alarm state and receipts"]
```

The diagram above is a target decomposition, not a claim that scheduled evaluations,
managed alarm state or general template resolution exist. The current YAML
operation planner and portable aggregate compiler implement bounded subsets.

Normalization resolves names, expands pinned templates, and produces a stable representation. Semantic validation determines subject grain, units, dimensions, time bounds, aggregation state, and evidence assumptions. Backend compilation chooses physical field bindings and query plans. Rendering consumes typed results; changing a geometry should not silently change a measurement.

SLO measurement and evidence inspection are separate requests sharing the same population definition. SLO totals aggregate all eligible subjects; the inspector returns a page of subjects. A trace relation changes the evidence being inspected, and cannot multiply the original SLI denominator accidentally.

## Source layout and design rationale

The original proposal favored a small typed core rather than a frontend rewrite
or premature service split. Current Go packages now implement bounded record
planning/evaluation (`exploration`), native metric kernels (`metrics`), measurement
packages (`measurements`), reusable artifacts (`workbench`) and prepared
aggregate/lifecycle logic. Shared mechanics live in `internal/httpjson`,
`internal/sourcehttp` and `internal/immutablefile`, behind their distinct owners.
General scheduled evaluation remains a separate undeployed design; no new
evaluator service or module rewrite is required by this documentation repair.

## Representation and language

JSON remains the browser/result transport. Strict YAML operation packages and
prepared aggregate definitions are implemented; CEL/KCL remain authoring candidates. The original design uses EDN and Clojure-style expressions because plain values and functional composition fit the project. The next implementation must preserve semantic structure across transport, regardless of syntax. Prefer a small normalized data model over a decision to port the entire app before its semantics are established.

Plotje may serve as a renderer or design influence. Trace waterfalls and related-record browsing need explicit result contracts even if they use a different rendering adapter. Native metric points and span-derived measurements also need distinct typed semantics.

Native duration histograms now have a separate Go kernel (`metrics/histogram.go`), strict YAML contracts, bounded histogram-table adapter and per-stream heatmap. Explicit aligned merging is core/CLI only. See [semantics and evidence](NATIVE_HISTOGRAMS.md) and [diagram 5](ARCHITECTURE_DIAGRAMS.md#5-native-histogram-intervals-and-bounded-investigation).

[Operation error ratios](OPERATION_ERROR_RATIOS.md) now derive from one counter
result with exhaustive outcome membership. Diagram 6 follows explicit SDK zero
baselines through the guided contextual investigation.

[Fleet coverage](FLEET_COVERAGE.md) now separates observed members from valid
aligned intervals. Diagram 7 shows guarded fleet totals and the selected-instance
path. Membership is declarative; discovery and automatic alignment remain planned.

Two live checkout instances now support [bounded-skew comparison](FLEET_COMPARISONS.md) through `metrics/comparison.go`, keeping actual member populations distinct from exact fleet totals. Diagram 8 includes the sample-only export fault gate.

Immutable capture storage lives in `investigations/`; the Go packages HTTP boundary captures trusted source results, retains separate bases/errors and publishes receipts to a local named volume. Browser reopening renders stored evidence and refresh creates a successor with the same fixed question. See [Saved investigations](SAVED_INVESTIGATIONS.md) and architecture diagram 9.

The bounded general exploration AST, reference predicates and ClickHouse compiler/source executor live in `exploration/`; `cmd/goo/exploration_http.go` exposes the Go API and the CLI shares its compiler. The UI retains original query and expanded trace/log bases separately. See [design](EXPLORATION_DESIGN.md) and architecture diagram 10.

Application-specific examples and general-purpose product boundaries are documented
in [the workspace contract](PRODUCT_WORKSPACE.md#general-purpose-boundary).

The shared Explore Go query now includes optional duration summaries/time buckets;
`query_budget.go` bounds general execution/cancellation, while service discovery
keeps event age and reported-parent uncertainty explicit. See
[core workbench](CORE_WORKBENCH.md) and architecture flow 20 for code contracts
and remaining production boundaries.

Built-in recipes are pure transforms in `exploration/recipes.go` followed by the
same source compiler. Sample investigation entry points now consume ordinary
telemetry packages rather than adding example-specific query branches. See
[the contract](INVESTIGATION_RECIPES.md) and architecture flow 24.

## Display component boundary

[UI components](UI_COMPONENTS.md) are native custom elements with declarative
JSON attributes and structured properties, usable from server-rendered HTML or
an SPA. They own semantic DOM, formatting and native disclosure behavior. The
application owns source reads, typed questions, scope changes, navigation and
custody. Shared component styling is separate from page layout; light-DOM adapters
must scope table/viewport rules to their own children. Current adoption covers
Services/Explore latency and stat lists, grouped latency comparisons, native-table
viewports and explicit result states. Table/state shells retain caller-authored
children across upgrade, attribute changes and reconnects; adapters own row
patches and status explanations. Implemented live transport remains application-owned, outside display elements.

Facets retain caller rows and filter them locally; typed field values render
JSON-compatible snapshot values with explicit presence. `attributes.js` remains
the app adapter for record mapping and query actions while exporting the reusable
value/list elements. Components do not interpret OTel namespaces or issue reads.

## Live observation adapter

The Go tail handler validates a supported log/span selector, performs bounded rolling row reads
and emits SSE observations through the Python streaming proxy. Explore owns
connection generations, reconnect gaps, retention and focus/follow behavior.
The passive table and attribute components retain caller rows. No tail buffer
becomes a measurement population. See the [sequence diagram](LIVE_LOG_TAIL.md).
Waterfall and count-time geometry are separate passive components; app adapters
own details, interval actions and source questions.

### Rolling record comparison and snapshot custody

The [live comparison sequence](LIVE_COMPARISON.md) describes the current path:
app-owned polling → frozen baseline question → equal-duration Go source read →
compatibility kernel → passive shared chart frame/paired bars. Freeze publishes
an expiring server-owned snapshot through the capture store without a source
read. The local tail buffer never enters this measurement path. `goo-chart-frame`
updates metadata without replacing caller-owned plots, native controls or focus.

The live span adapter now renders count, duration and status from one aggregate
result. Its single response supplies the current basis/read/window and all
summaries; synchronous rendering replaces those measures together. Paired
`goo-latency` elements accept the same app-supplied `ceiling-ns` while recorded
error share uses a fixed percentage axis. Freeze retains the original aggregate
and basis. The [sequence diagram](LIVE_COMPARISON.md) and journey 71 include this
path; inspection rows/field discovery are still separate source reads.

Population breakdowns reuse the existing distribution/comparison adapter with
per-signal DOM namespaces. Field selection and value intersections produce typed
Go queries; comparison uses the existing Go cohort kernel and shared zero-based
visual renderer. `ExplorationView.breakdowns` optionally stores display choices
and baseline queries, never results. Validation requires compatible signal,
environment and grouping; source reads occur only after explicit Run/compare.
See [population contract and flow](POPULATION_WORKSPACE.md#attribute-breakdowns-and-comparisons).

Population live subscriptions use the existing Go SSE tail transport with an
optional signal selector (omitted means logs for compatibility). The source
adapter compiles span/log predicates uniformly; observation identity remains a
bounded per-connection multiplicity tracker. Independent baseline captures and
rolling comparison reads use existing Go kernels, never the display buffer.
See the [live process diagram](POPULATION_WORKSPACE.md#watching-a-population).

## Authored context and optional retained evidence

[Shared tags and notes](WORKBENCH_ANNOTATIONS.md) use typed exact targets and
immutable parent-linked revisions in `workbench/meaning.go`. UI adapters share
`ui/meaning.js`; its `goo-tag-list` element remains display-only. Scoped service
assertions name an applicability window; exact saved artifacts do not manufacture
one. Active heads derive from parent edges, preserving parallel reviews.

The optional `retention` Go package writes selected **server-owned live-page or
saved-capture** observations and their initial annotation atomically into XTDB
through pgx. `exploration/result_snapshots.go` holds bounded exact result pages
for two minutes; client records never become authoritative evidence. Refresh and
paging create new tokens, while original page indexes remain stable until expiry.
Native member projections support joins while exact JSON payloads remain retained.
Initial and correction request receipts recover uncertain responses without
additional revisions. New connections apply the caller's XTDB await-token lower
bound; the UI retains it across reloads in the same tab. This does not pin a
historical knowledge cutoff. The capture store, ephemeral pages, live ClickStack
source and native evidence store remain distinct boundaries; reopening retained
evidence reads neither live telemetry nor the original capture/page. The
[contract and process diagram](WORKBENCH_ANNOTATIONS.md) identify limits including
cross-store search, cross-client basis sharing and authentication. No ClickHouse
external-source connector is implemented.


## Retained evidence in investigations

`retention/catalog.go` enumerates bounded immutable manifests and explicit current
review heads; metadata filtering never reads ClickStack. Browser ownership is split
into `retention-client.js` (transport/read basis/journal boundary),
`retention-catalog.js` (filters/cursors/snapshot selection), `retention-citations.js`
(pinned reads/attachment), and `retention-recovery.js` (explicit retry/forget).
`retention.js` composes selection authoring and retained observation views.
See the [owner diagram](DEFINITION_LIBRARY.md#retained-catalog-snapshots). Retained document cards pin both
a manifest ID and an exact review revision with an await-token lower bound.
`workbench/documents.go` validates the typed reference and a native resolver checks
it before publication. Opening the file-backed narrative does not require the
native database; resolving its reference body does. The native retry journal in
`ui/pending-saves.js` performs no network work: application adapters persist the
original request before sending and require explicit recovery after reload. Passive
custom elements remain display-only. See the [process diagram and limits](WORKBENCH_ANNOTATIONS.md#find-cite-and-recover-retained-evidence).

The [whole-product usability review](WHOLE_PRODUCT_USABILITY_REVIEW.md) proposes
entry/action/layout consolidation across these existing adapters. It is a review
of the preceding runtime, not a claim those structural changes are implemented.


## Application composition ownership

[Live and saved operations](LIVE_SAVED_OWNERSHIP.md) now have separate destinations
and selection controllers. Workspace navigation never reparents the live editor.
`GooOperations.overview` renders supplied operation evidence for both owners;
saved component IDs are scoped. No receipt hash or Go semantic contract changes.

Population owns its applied scope, total, timeline and record page. Its facet
adapter uses `GooDistributions.mountFacets`, which shares the existing facet and
comparison mechanics without constructing a second count/chart/measure workspace.
Recent timeline controls and mobile facets are source-free disclosures. The full
single-question distribution and service timeline compositions remain available.


## Fresh live entry

The investigation trail now registers handlers without constructing a fixed editor
on fresh Activity or unrelated initial routes. Explicit fixed authoring, record
question/step/view links and supported service/package/trace continuations mount
one. First legacy continuation events are delivered once after mounting; later
ones restore the retained owner. Mounting itself never executes telemetry. Existing
continuations keep their original staged-versus-explicit-read behavior. Suspended
editors keep their draft/DOM identity; they are not rebuilt on paused Activity return.

```mermaid
flowchart LR
  Fresh[Fresh Activity] --> Tail[One bounded tail; no fixed editor]
  Fixed[Explicit authoring / record link / continuation] --> Trail[Trail ensures one editor]
  Trail --> Stage[Stage fixed scope / retain draft]
  Stage --> Run[Explicit read under existing typed owner]
  Stage --> Return[Paused Activity return]
  Return --> Hold[Detach / suspend retained editor]
  Hold --> Reopen[Explicit reopen: same draft and DOM]
  Reopen --> Stage
```

[Live Activity](LIVE_ACTIVITY.md) classifies visit intent before subscribing, reuses the existing bounded tail core and keeps its local inspector separate from fixed execution. Saved/shared entry remains inert. The [whole-product architecture review](STREAM_FIRST_WHOLE_PRODUCT_ARCHITECTURE.md) records the original shared record/context rationale. Delivered adoption is
described below; its broader proposals do not override the current roadmap.

Live Activity now keeps signal-local `ScopeFilters` draft/applied predicates in its
application owner. Shared `GooAttributes` emits typed Include/Exclude intents;
chip edits pause locally, Apply starts a new tail generation, and the inspector's
fixed continuation stages the original row read window in `GooExploration` without
fetching. Only the existing Run action rereads it. The existing trail exposes
suspend/resume of its retained fixed view; Live Activity uses these to cancel hidden
work and restore its paused parent without a read. Shared record bodies are now
implemented with the shared table shell described below.

The [shared record presentation](RECORD_PRESENTATION.md) adapter now removes
parallel row-formatting paths from Live and fixed results. Its custom elements
render supplied data only; application adapters retain reads, scope/basis, selection,
fixed/trace continuations and retention. Supplied-parent context has bounded exact
ID traversal and explicit ambiguity/missing/cycle states, with no wider lookup.

Live, fixed/frozen pages and held service observations share the passive record
table builder. The Services adapter lazily presents its existing supporting span
page, validating the returned question before adoption. It does not bind that page
to independently reread overview selections. See the [ownership diagram](RECORD_PRESENTATION.md#held-service-observations).

The [saved record presenter](SAVED_RECORD_PRESENTATION.md) maps capture pages and
metadata-only document references to shared frozen display facts. The capture owner
supplies the single scope summary; its compatibility viewer retains paging and
context reads. A record's fixed continuation stages a draft; Run owns the fresh read.

Stored record pages now use `GooSavedRecords`, independent of the live query
editor. `GooRecordResults` supplies record/group tables and named context actions
to both owners, with independent generations and retained evidence bases. Capture
switch/leave/document transitions invalidate old context publication. See the
[updated ownership diagram](SAVED_RECORD_PRESENTATION.md).

## Selected definition revision libraries

The first [B1 library slice](DEFINITION_LIBRARY.md) shares loaded-catalog paging,
compact rows and a cached inspector between record/native Definitions, fixed reusable-question
and document revision adapters. Retained catalogs share the row/inspector primitives
with load-specific snapshot keys, retaining independent server filtering and cursors.
List fetch/refresh generations, compiler previews, execution and authoring stay in
typed adapters. Selection preserves exact family/revision identity; it does not
replace the editor draft or infer lineage from a name/title. Record annotation
panels mount only in selected detail; their existing publication/basis owners are
unchanged. Only the row list scrolls; the inspector sits outside that viewport.

The fixed reusable-question owner invalidates pending plans on selection/close,
workspace/mode changes and suspension. After its compiler response, it checks the
owner generation, active visit and exact revision before explicit source dispatch.
Local selection does not mutate the fixed draft or execute a query.


## Record field display boundary — B2 first slice

`field-presentation.js` shares qualified field/operator labels and record predicate
text across fixed chips, the existing live selector summary and record package
handoffs. It consumes supplied predicates without changing them. Shared Go/JS wire
fixtures pin omitted empty defaults, field-family separation and exact decimal
strings; native binding semantics remain independent. See the [contract and flow](FIELD_PRESENTATION.md).
The source exploration owner and retained metadata catalog have different abort,
visit, snapshot and cursor lifetimes; this slice introduces no common request owner.


The B2 service namespace continuation shares `GooServiceIdentity.label` with
annotation summaries, preserving the existing selector/key/name owners. Shared
Go/JS namespace wire fixtures pin wildcard, missing, empty and literal populations
without changing native bindings or server digests. See the [namespace flow](FIELD_PRESENTATION.md#service-namespace-identity--b2-continuation).


## Shared question protocol — B2 continuation

`ui/question-protocol.js` is the pure browser/Node owner for question/step/link
transport, capture-question projection, fixed-scope checks and comparison-only
native default equivalence. N1 was independently approved and delivered in
`9277f5b`. `ui/workspace.js` owns navigation/layout and reexports the same function
references for existing consumers; it is not a second protocol implementation.
Server HTML loads the protocol before workspace. The lazy fixed-owner mount and
existing navigation/read/publication lifetimes remain unchanged.

Evidence comparison resolves `question-session.identity` at call time, retaining
the existing browser pre-session stable fallback. Loading the protocol does not
create a Session, touch DOM, fetch, subscribe, read storage or publish. Node's
identity helper is likewise resolved only when comparison needs it. Recovery,
trace findings and fresh comparison consume the pure module directly; other
callers may retain the compatibility facade.

Go/JS fixtures pin accepted/refused record/native steps, exact bounds, error/wire
behavior and native definition digests across actual serialization. Native reuse
comparison keeps its cloned Go-default projection; record predicate identity is
not used to normalize stored native definitions. Go remains authoritative for
full semantic validation. See [the protocol contract and executed acceptance](QUESTION_PROTOCOL.md).

```mermaid
flowchart LR
  Workspace[Navigation / layout facade] -->|Same compatibility function references| Protocol[Pure question / step / link protocol]
  Owners[Recovery / trace finding / fresh comparison owners] --> Protocol
  Protocol -->|Lazy call-time identity lookup| Identity[Question-session identity helper]
  Protocol --> Wire[Unchanged question transport]
  Source[Application read owners] --> Go[Existing Go semantic validation and execution]
```

Activity exact retention now injects one process-local membership bank into the
trusted SSE poll issuer and references-only publisher at startup. Issuance freezes
original read provenance; publication first resolves committed request identity,
then consults ephemeral proof only for a definitely new intent. The browser holds
at most 20 original refs and journals immutable writes before sending. Shared
origin readers/citations remain downstream of XTDB, with no telemetry reread.
The custody flow and limits are in [Activity selection](TAIL_SELECTION_CUSTODY.md).

## Subsequent Activity, filtering and record composition

Reusable [Activity views](ACTIVITY_VIEWS.md) use a typed rolling configuration in
workbench's existing immutable file publisher, separate from fixed exploration
views. Shared revision-library display appears in Saved and Services. The stream
adapter restores configuration through `GooTail.stage`; Start alone opens SSE.
No observations or capabilities cross the configuration boundary. The linked
flow diagram documents publication, staging and later separate retention.

The [shared predicate editor](FILTER_EDITOR.md) is a supplied-input helper, reused
by Activity and fixed population adapters. It normalizes Go's empty-value wire
shape but does not evaluate, read or publish telemetry. Activity owns draft/apply;
fixed adapters stage child configurations and retain their parent's source basis.

Single-question filtering now uses that same supplied editor. Full restored ASTs
remain editable; shortcuts become explicit draft predicates. The existing session
continues owning generations, fixed bounds and requested/applied evidence, and
the query owner continues owning cancellation and reads. Environment edits retain
the draft and fixed bounds; incompatible signal edits reset predicates.

[Observed suggestion flow](FILTER_EDITOR.md#observed-fields-and-values) reuses successful applied record reads. The source adapter projects bounded qualified fields/values; the editor presents supplied observations and emits draft predicates. Scope ownership, invalidation, clocks and source execution remain in Activity/single/fixed adapters. Saved configuration never persists sampled suggestions.

[Explicit value lookup](FILTER_VALUES.md) adds an application owner over the existing exploration grouping route. The passive editor emits a Load intent; each Activity/single/fixed adapter supplies its applied scope and lifetime. Lookup results retain their own read clock/basis and never replace page evidence, source capabilities or custody.

Latency distribution interaction reuses passive `goo-time-bars` range intent. `latency-histogram.js` validates complete raw-span bucket counts and translates keys into existing `population-selection.js` transforms; the population/trail owners stage and execute queries. No new evaluator, route, source adapter or component subscription is introduced. See [flow and limits](LATENCY_SELECTION.md).

`activity-latency.js` owns optional aggregate refresh/cancellation over the latest completed stream window. It supplies snapshots and interaction callbacks to the existing histogram adapter, stages predicates through `stream-entry.js`, and uses the existing fixed continuation for disjoint comparison. Components never read or subscribe. [Activity flow and clock boundaries](ACTIVITY_LATENCY.md).

`activity-baseline.js` owns immutable capture selection and cancellable pure comparisons over supplied Activity latency groups. Receipt loading and current-result comparison use independent generations; component lifecycles own neither reads nor subscriptions. Existing Go comparison and shared histogram/context displays retain accounting and presentation contracts. [Flow and clock boundaries](ACTIVITY_BASELINE.md).

Activity views now have an optional exact `baselineCapture` reference. Go publication resolves and validates stored duration evidence; GET/catalog validate only configuration integrity. The baseline owner stages the ID separately from loaded data, and resolves it after explicit current-window loading. Legacy omitted-field hashes remain unchanged. [Reference flow](SAVED_ACTIVITY_BASELINE.md).

Activity’s Views menu reuses the configuration-library adapter with explicit applied-environment and visibility inputs. Scope changes cancel publication and clear the passive library’s rows/inspector cache. No telemetry, baseline receipt, source clock or subscription is loaded by browsing. [Flow and storyboard](ACTIVITY_VIEWS.md#activity-entry-storyboard-159).

The [Activity navigation adapter](ACTIVITY_NAVIGATION.md) now exposes fixed-window staging and retained authoring through one Investigate disclosure. It pauses/remembers before mounting the existing fixed owner and hides the old entrance during Activity; parent return and explicit execution contracts are unchanged.

[Compact mobile records](COMPACT_MOBILE_RECORDS.md) are a supplied presentation opt-in shared by SSR and SPA tables. Responsive styles stay with the record component; legacy tail formatting excludes opt-in tables. Existing Activity owners retain selection, query windows, timestamps, navigation and source execution.

The shared record-result owner now selects the [responsive record layout](COMPACT_MOBILE_RECORDS.md#fixedsaved-adoption-storyboard-162) for fixed, saved and explicitly read context tables. Presentation does not change stored receipts, result models or read clocks. Mobile formatting is independent from preview clipping; spanning inspectors and accessible action identities remain with the existing owner.

[Service-held observations](SERVICE_BROWSING.md#held-service-observations) adopt the same passive responsive record table. Service detail still supplies the original supporting page, retains its inspector and stages the fixed-question link. Disclosure, inspection and resizing introduce no source read or new custody semantics.

The [record renderer audit](RECORD_RENDERER_AUDIT_2026_10_24.md) keeps custody, diagram and declared-operation semantics in their owners while sharing identity/time/attribute presentation. Operation tables use bounded passive table viewports; local trace/subject selection still uses the held snapshot. [Independent operation status](OPERATION_STATUS.md) now retains exact optional exporter status alongside declared outcome. Conformance uses outcome; waterfall error color uses reported Error. Legacy captures remain immutable and explicitly mark status as Not retained. The passive waterfall receives separate display facts, without source policy or source reads.

[OTel field hints](OTEL_FIELD_HINTS.md) add an advisory exact-qualified-key lookup to the shared field presenter. Filter and attribute adapters reuse it for names, descriptions and local search. Source adapters, query types, saved identities and execution owners remain separate; recognizing a name never rewrites a key or infers schema compliance.

[Observed grouping](OBSERVED_GROUPING.md) reuses the existing regroup transform across Activity and fixed owners. The field browser emits exact field/opener intent; app adapters freeze or retain scope, stage drafts/children and preserve companion questions. Go remains the execution validator; no source client lives in the editor.

[Two-field composition](COMPOSED_GROUPING.md) keeps exact ordered grouping transforms in the population/query owner. Same-membership grouping drafts can hold scoped observation metadata, while explicit reads, changed scope and saved reopen withdraw it. Unread population drafts update in place; applied parents remain separate. No observation metadata enters saved query schemas.

[One grouping control](GROUPING_CONTROL.md) replaces selector DOM state with the shared editor’s canonical grouping array. The passive custom element receives cloned display DTOs and emits intent; scope owners retain validation, staging, reads and publication. Legacy immediate population dropdowns are removed.


## Typed numeric record predicates

[Numeric filter leaves](TYPED_NUMERIC_FILTERS.md) reuse explicit NumericBinding
metadata while remaining independent of question-level numeric aggregates.
`exploration/` shares canonical parsing and nullable SQL value guards between
reference matching, numeric aggregates and predicates; typed threshold parameters
stay exact. Missing/empty/invalid/valid form an explicit partition, and comparison
inequality is valid-only. Physical map/column choices remain in source adapters.
The existing editor, protocol/recovery guard and field presenter preserve authored
intent; they do not evaluate captured source values. Query/capture/definition and
read-time versus immutable-custody owners retain their existing lifetimes.


## Activity explanations reuse the population engine

[Activity explanation staging](ACTIVITY_EXPLANATIONS.md) is source-approved in
`cdd5a88`, with retained-visit ownership fixed in `a6533a6`. Journey 178 spans/logs,
browser ownership and its 60-second actor pass; the reviewed
[desktop](journeys/screenshots/178a-spans-activity-explanation.png) and
[mobile](journeys/screenshots/178b-spans-mobile-explanation.png) frames show this flow.
AE2 acceptance is complete, including native main/ownership/baseline rejecting
controls and 70 Hegel examples. Journey 155 native/actor checks passed separately
after a combined-wrapper timeout; see [executed validation](../VALIDATION.md).
`stream-entry.js` supplies the last
completed applied 60-second window and selector. `activity-explanation.js` pauses
the stream and optional latency updates, then stages that scope in the existing
explanation owner. Its compact held-occurrence chart guides a fixed time brush;
its displayed occurrences never define full counts or the window boundaries.

The default baseline is the immediately preceding equal-width window. An explicit
capture metadata load can instead supply the original same-signal/environment
query window and selector. This is scope reuse: **Run explanation** independently
rereads both populations through the shared Go count/group engine. Frozen capture
counts are not substituted for that baseline read. The originating capture link
and original source/custody clocks are UI context; the saved explanation retains
its request and newly observed aggregates, without that originating capture ID.

Opening, brushing, choosing fields and paused return do not execute telemetry
queries. Loading a saved scope is an explicit metadata request. Scope edits, Apply,
resume and replacement withdraw obsolete results/tokens; hiding or leaving cancels
pending work. Component lifecycle owns no subscription. **Save observed explanation**
promotes the exact held observation through the existing bounded snapshot bank and
immutable capture store, without rereading either population.


## Recorded error triage over ordinary telemetry

[Error triage](ERROR_TRIAGE.md) reuses applied Activity scope, existing fixed
questions, count/group/time reads and captured investigation work. ET1 is deployed
in `982b9e9`, and ET2 staging/passive details in `0c4ab38`. ET3 acceptance is
complete: real OTel spans/logs, ownership, native/model rejecting controls and
actor checks pass. The passive SSR/SPA cases use supplied synthetic data, not
OTLP. [Actual frames and evidence](journeys/README.md) distinguish these checks;
VALIDATION.md records corrected supplemental-log probes and final review/push.

`exploration/exception_fields.go` projects one span record from named `exception`
events. Event presence is names-only; type/state projection requires aligned
name/attribute arrays. A nullable type exists only for one distinct reported
`exception.type`, including a present empty value. Unreported or multiple types
have no unique type, with a separate reporting-state field. Grouping counts
records once without exploding events or choosing the first type. Logs retain raw
severity text and gain intrinsic numeric severity; explicit rules preserve the
independent reported signals without inferring failures or issue identity.

Derived row aliases appear only when requested, preserving non-opt-in row SQL
and valid legacy captures. Shared nullable-field rules allow missing map keys and
`column:exception_type`; caller-held `/compare` groups with null nonnullable fixed
columns now return 400 instead of accepting malformed logical-source data.

The Activity adapter stages and pauses; the fixed owner executes only explicit
Run. Existing group links stage exact service/type questions, and service-scoped
single-type grouping enters existing comparison controls. Captures, reusable YAML
packages and investigation notes retain those typed scopes and evidence.

The passive `goo-exception-details` Web Component accepts supplied `{record,signal}`
through SSR's JSON `data` attribute or a cloned SPA `.data` property, including
pre-definition assignment. It shows bounded local event/structured-log evidence,
UTC times, messages and disclosed stack traces; lifecycle and disclosure do not
fetch, subscribe or publish. Source policy and query lifetime remain in owners.
Raw nested-event Databricks lowering remains unsupported; prepared-fact export
requires an explicit grain/alignment/nullability contract.

## Fixed-scale decimal interpretation

[Decimal bindings](FIXED_SCALE_DECIMALS.md) extend the existing numeric adapter,
not the query runtime or physical OTel schema. An explicit scale defines a signed
int64 coefficient lattice. Pure validation, reference evaluation and guarded
ClickHouse integer parsing share the literal contract; sums widen to exact
integer arithmetic. No floating point parser or Decimal cast defines coverage.

```mermaid
flowchart LR
  D[Declared field, decimal64, scale and unit] --> Q[Typed record question]
  Q --> V[Pure binding and threshold validation]
  V --> C[Guarded source string to int64 coefficient]
  C --> A[Exact aggregate and coefficient bins]
  A --> W[Coefficient strings and explicit scale]
  W --> P[Passive scale-aware decimal display]
  W --> E[Immutable capture with original question]
  P --> S[Range intent staged by application owner]
  S --> Q
```

Scale is part of measurement identity. Incompatible scales withhold comparison
arithmetic. Legacy integer bindings/results omit scale. Source-read budgets,
record grain, independent-read uncertainty and capture custody are unchanged.

## Contextual numeric declaration ownership

[Plot values](CONTEXTUAL_NUMERIC_PLOTS.md) is an application-owned adapter over
the existing question, numeric binding and histogram contracts. Attribute and
observed-field presenters emit qualified field intent only. The shared dialog
locks that field and collects explicit type/unit/scale without parsing a sample.
Activity, applied fixed results, immutable saved pages and population children
supply their own exact question and lifecycle authority.

The pure transformation clones membership predicates and reported-edge scope,
resets page/presentation measures and adds one validated numeric binding/histogram.
Native question-trail staging retains fixed/saved parents; population children
retain their separate parent and companions; Activity preserves its paused
window, inspector and opener. Run uses the existing source adapter and custody
owners. Hidden/disconnected or obsolete dialog owners cannot publish staging.
No display component, new API, numeric cast policy or evidence store owns this
continuation. The flow diagram is in the contextual contract.

## Log-pattern projection

The [versioned body profile](LOG_PATTERNS.md) is a derived logical log column in
the existing Go source adapter. Reference and ClickHouse lowering map one body
to one collision-safe identity; the stored Body is not rewritten. Length is
checked before UTF-8 validation or tokenization, preserving bounded evaluation
with short-circuiting disabled. Empty/oversize/invalid inputs contribute explicit
unclassified buckets rather than truncated apparent shapes.

Activity stages the applied completed window. Existing fixed queries own full
counts/groups and overall time series; per-pattern children and selected/rest
service comparisons reuse the same query/capture/publication paths. Shared
display adapters decode identities for labels, while typed predicates retain
the exact canonical IDs. No sampled classifier, ingestion processor, ML model
or independent pattern store is introduced. The process diagram is in the guide.

### Pattern comparison across time

[Time comparisons](LOG_PATTERN_TIME_COMPARISONS.md) reuse complete fixed-query
cohorts and the existing population owner. Baseline selection stages questions;
explicit Run owns independent source reads. The comparison boundary validates
canonical pattern identities before returning even an empty or changed state.
A pure display adapter aligns complete groups and computes exact integer ratios
for duration-normalized rates and all-record shares. Stored-evidence adapters
load immutable full group aggregates from storage, never infer them from a record
page, and do not start telemetry reads. Source and evidence ownership remain in
application adapters; no display element subscribes or fetches.

## Supplied-page trace interval accounting

[Trace page timing](TRACE_PAGE_TIMING.md) is a source-free Go projection over
existing public span rows, mirrored in the passive JavaScript display with exact
parity tests. No new endpoint, query flag or persisted capture field is needed.
Reported IDs resolve unique direct-parent edges; ambiguous or incompatible
intervals remain unavailable locally. Union length uses exact integer intervals,
and nullable service/operation groups retain original row indexes and separate
known/unavailable measure coverage. Group sums may overlap across spans.

The component emits controlled selection intents. The waterfall application
adapter validates current ownership and visible local controls, then coordinates
page/aggregate highlights and the existing inspector. Its separate target-level
linked-log listener has the same generation and visibility guard. A closed or
superseded component cannot bypass it through a retained native button. Existing
full-query aggregates remain independent of this bounded-page calculation.

The trace inspector can compare two supplied pages locally. Reference custody is
bounded to its record-results owner; no new API, token or persistent resource is
introduced. Relative clocks are normalized per supplied page, while original
fixed bounds, source clocks and bases remain visible.

```mermaid
flowchart LR
  A[Expand trace: existing fresh source read] --> B[Bounded supplied page]
  B --> C[Pin detached page in owner memory]
  D[Expand another trace: explicit fresh read] --> E[Current bounded page]
  C --> F[Pure Go normative / JS mirrored comparison]
  E --> F
  F --> G[Passive aligned relative-time display]
  G --> H[Local unique-label pair attribute inspection]
```

Repeated labels remain ambiguous instead of being zipped. Group span-duration sums
may overlap, and missing page observations are not zero durations. Comparison
selection is independent of native waterfall selection, which retains its existing
explicit linked-log read behavior.

Contextual trace search preserves the distinction between a span predicate and
independent conditions that may be witnessed by different spans. Its common scope
is only the applied environment and fixed window; the original WHERE is retained
inside condition one, so additional conditions can match other services.

```mermaid
flowchart LR
  A[Inspect applied span or qualified field] --> B[Stage condition 1 with original WHERE]
  B --> C[Add up to 3 independent span conditions]
  C --> D[Explicit Run trace search]
  D --> E[Go validates and executes one grouped source query]
  E --> F[Full distinct reported-ID count and bounded representative witnesses]
  F --> G[Explicit Open trace: existing fresh expansion]
  G --> H[Existing waterfall and local reference comparison]
```

The same span record can witness multiple conditions. No topology or unique
invocation is inferred, and conditions remain fixed to the original window.
Cancellation, scope edits, hidden/disposed ownership and stale generations
withdraw pending results. This contextual adapter introduces no capture format.

## Counted field values as a bounded source projection

The first counted-values adoption adds `Source.FetchCountedValues` and
`POST /explore/counted-values` (UI proxy `/api/explore/counted-values`). This is a
separate `goo/counted-values/v1` result, not a new mode on the full record result.
Existing `Source.Fetch`, record/capture DTOs, saved bytes and their custody paths
remain unchanged.

A valid request has one string grouping field, offset zero and no numeric result,
duration/status summary or time-series projection. It retains the original fixed
environment/window, WHERE (including typed numeric predicates) and reported-edge
scope. The source detaches caller-owned state before I/O and uses the existing
compiler/transport to read only the independent count and complete groups. It
does not read inspection rows or discover fields as hidden prerequisites.

```mermaid
flowchart LR
    I[Explicit field-value request] --> Q[Validate and detach typed question]
    Q --> C[Bounded count read]
    C --> G[Bounded complete group read]
    G --> V[Validate values and count conservation]
    V --> R[Versioned supplied counted-values result]
    R --> E[Passive editor display and literal filter intent]
```

The result retains separate count/group read clocks and raw-read bases, plus a
combined result basis independent of those clocks. Missing values are null and
present empty strings remain distinct. Groups require unique canonical values and
positive canonical uint64 counts; nonnullable dimensions and log-pattern IDs keep
their existing validation rules. More than 1000 groups refuses the entire result,
without partial grouping totals. Count/group disagreement returns explicit
`changed_between_reads` diagnostic groups, not a coherent actionable population.
These sequential reads are not a transactional snapshot.

`ui/filter-values.js` owns explicit execution, cancellation and generation/scope
binding. The filter editor receives supplied values and emits filter intent; it
does not fetch telemetry. Opening the editor or changing local conditions does not
automatically load counted values. Stale, failed, malformed and changed-scope
responses cannot become usable value suggestions. Counts derive from full query
populations, not the visible record page. [Journey193](../journeys/193-counted-values-projection-storyboard.md) accepted
143 matching logs with a 100-row inspection page and four complete value groups,
a separate five-record population, and explicit missing-value selection of one
record. Numeric WHERE retained the exact integer threshold above 2^53. A held real
response could not publish after a scope edit. The bounded native raw-fact check
passed in 0.177s with eight changed controls; it validates this observed scope and
membership, not all internal SQL or arbitrary schedules. The browser journal's
five telemetry API POSTs exclude one cancellation control and the separately
recorded small-population API/readiness probes. The architecture does not claim
universal adoption by every grouped read or a replacement for capture semantics.
