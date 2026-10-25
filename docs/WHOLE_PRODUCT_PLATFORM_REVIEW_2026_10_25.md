# Whole-product platform review — 2026-10-25

Reviewed source: `9e07b88`. This is an input to the revised roadmap, not an
additional implementation queue. The review reads source and existing acceptance
contracts; it does not claim a new runtime, browser, load or security test.

## Assessment

Goo's strongest platform boundary is already useful: **typed questions over
ordinary OTel observations, explicit execution, and evidence that retains its
meaning when revisited**. Manifests and source annotations can strengthen a
binding; they must remain optional for investigation. The Go semantic core,
ClickHouse adapters and passive Web Components support that direction without
requiring a new query engine or frontend framework.

The next platform work should make the unit of observation and deployment
contract easier to understand. More persistence alone would not resolve either.
A generic service's span count is not request throughput, an environment filter
is not authorization, and a generated warehouse definition is not an executed
warehouse deployment. The present implementation mostly states these boundaries
correctly. The opportunity is to make useful choices inside them more accessible.

No new source-correctness defect was reproduced in this review. The findings below
separate source-backed capability gaps, intentional local-lab boundaries and
requirements that activate only for a named shared deployment.

## What is already delivered

| Platform concern | Current evidence | Remaining boundary |
| --- | --- | --- |
| General OTel exploration | Activity/fixed questions, native discovery, service dashboards, trace search, waterfalls and local page comparisons work without an application manifest. [README](../README.md), [trace search](TRACE_SEARCH.md). | Deployment bootstrap still loads bundled package contracts; source tables follow a pinned exporter adapter. |
| Typed definitions and optional Go hooks | Reusable record/native YAML packages, immutable revisions and UI-to-Go template handoff exist. `//goo` operation/event hooks and opt-in weaving exist. [Telemetry packages](TELEMETRY_PACKAGES.md), [measurement packages](MEASUREMENT_PACKAGES.md), [annotations](GOO_ANNOTATIONS.md). | These are several explicit semantic families, not one universally portable evaluator. A source manifest does not prove linked-build coverage or runtime emission. |
| API contract | OpenAPI inventories 76 operations, including 70 Go operations; decoding, effects, limits, cancellation and proxy differences are documented and exercised. [API](API.md). | Structural schemas do not enforce all semantic constraints or confer stability, authentication or tenant isolation. |
| Semantic-convention guidance | Qualified field hints, HTTP/messaging capability plans, native metadata/stream identity and explicit legacy spellings exist. [Field hints](OTEL_FIELD_HINTS.md), [capability source](../measurements/instrument_capabilities.go#L164). | Recognized names and schema URLs are evidence to inspect, not proof of conformance. |
| Evidence and annotations | Immutable file-backed artifacts, branch-preserving notes/reviews, exact selected-page retention and optional XTDB support exist. [Publication](IMMUTABLE_PUBLICATION.md), [annotations/retention](WORKBENCH_ANNOTATIONS.md). | Restart-sensitive capabilities, verified identity, multi-user authorization, production backup/deletion and shared-draft concurrency remain distinct concerns. |
| Warehouse portability | The prepared-fact aggregate subset emits ClickHouse plans and Databricks YAML/DDL. [Portable compiler](PORTABLE_COMPILER.md#implemented-semantics). | Databricks execution, physical ingestion and arbitrary interactive-query lowering are not accepted capabilities. |

Do not reopen these delivered foundations as fresh projects merely because an
older proposal calls them future work.

## Source-backed findings

### 1. Manifest-free investigation still has package-coupled server bootstrap

The generic query path does not inspect application source. However,
[`serve_http.go`](../cmd/goo/serve_http.go#L28) requires binding, inventory,
scope-name and scope-version flags before registering any routes, and defaults
an omitted operation-package list to `config/operations/checkout.yaml`.
[Compose](../compose.yaml#L78) supplies those sample contracts. This limitation is
already acknowledged in [telemetry packages](TELEMETRY_PACKAGES.md#api-and-persistence).

This is a concrete onboarding/deployment gap, not evidence that generic queries
have incorrect results. A user with an existing ClickStack-shaped source should
be able to start generic exploration without first supplying irrelevant HTTP or
operation declarations. Add an explicit general-source startup profile; keep
strict package validation when those optional families are enabled. Do not
manufacture a library/schema binding from observed field names.

### 2. Default service views have the right caution but lack a selected work grain

[`dashboard.go`](../exploration/dashboard.go#L105) deliberately creates
`span-record` widgets, including mixed kinds and retries. Its HTTP/RPC/database/
messaging dimension widget changes grouping, not the measured population.
[Service dashboards](SERVICE_DASHBOARDS.md#one-planner-existing-evaluators)
state this correctly.

The missing feature is an explicit population/profile choice. Start with
**All spans** and **Observed HTTP server spans**, carrying their predicate,
record grain, outcome policy and coverage into the existing questions. Presence
of `SpanKind=Server` alone does not establish HTTP traffic, a root request or a
business invocation. A profile must disclose required HTTP attributes, missing
values and its exact current/legacy convention binding.

Do not reimplement `http.server.request.duration` or
`messaging.process.duration`: their native capability plans are delivered in
[`instrument_capabilities.go`](../measurements/instrument_capabilities.go#L164).
They evaluate supported exact streams; they do not already provide a verified
fleet HTTP error rate. Likewise the existing bound HTTP preparer checks IDs,
deduplicates identical observations and refuses conflicting copies in
[`clickhouse_http.go`](../prebuilt/clickhouse_http.go#L73). Reuse that policy when
explicitly offering a request-identity aggregate; do not silently substitute it
for raw-record semantics or generalize its reviewed library binding to all data.

A first useful profile can use current integer span durations, counts and typed
predicates. It needs no float coercion, unit conversion or new metrics engine.

### 3. Query admission is route-specific; proxy cancellation is not universal

The four-slot Go budget covers an explicit POST list, including trace search and
population explanations. Other routes bypass it by design in
[`query_budget.go`](../cmd/goo/query_budget.go#L31). Executing GET operations and
some capture families therefore do not share that admission contract. Tail has
its own stream slots. Per-query ClickHouse read/time/memory caps still exist in
[`source.go`](../exploration/source.go#L65); this is not an unbounded-SQL finding.

The Python JSON proxy forwards a query capability but buffers the upstream
response through `urlopen` in [`server.py`](../server.py#L123). A browser abort
alone is not proof that every proxied backend operation stopped. The separate
cancel endpoint and source cancellation settings matter. Some upstream response
headers, including Retry-After, are not forwarded; [API.md](API.md#two-boundaries-with-existing-differences)
already records this difference.

A bounded follow-up should inventory effects and define cancellation/admission
for each source-executing family, then adopt compatible families individually.
Preserve tail streaming and publication-uncertainty rules. Do not mechanically
put every GET, pure plan, file read and publication under one interchangeable
query owner. These are known coverage boundaries, not a newly reproduced bug.

### 4. Portability should be explained per definition family

Current record queries, native metric definitions, operation/lifecycle policies
and portable prepared-fact aggregates have different operators and source
requirements. YAML syntax shared across them is not cross-target equivalence.
The [Databricks contract](DATABRICKS_METRIC_VIEWS.md#authoring-and-compiler-constraints-to-adopt-now)
already proposes `direct`, `requires-preparation` and `unsupported` target outcomes.
Make a small, source-backed capability report available when validating/exporting
an existing family. Include grain, identity, time/basis, units, outcome policy,
required prepared artifacts and unsupported operators. Avoid a universal IR rewrite
merely to display this report.

The convention references also have different purposes: field hints describe a
curated v1.43-era key catalog, while semantic package profiles pin v1.37. That is
not itself a contradiction. Preserve exact keys and profile versions; add an
explicit adapter/profile compatibility description instead of silently translating
legacy names. The span exporter drops scope schema URL in the bound HTTP path;
[the live HTTP contract](LIVE_HTTP_PACKAGE.md#explicit-source-binding) correctly
calls its version mapping declared rather than observed.

### 5. Pipeline health still needs an actual source of evidence

The collector config has memory limiting, batching, retry policy and a 24-hour
ClickHouse TTL, but no declared pipeline that ingests its own receiver/exporter
health into Goo: [`collector.yaml`](../config/collector.yaml#L6). That config is
not proof that no internal metrics exist; it is evidence that the product has
not defined and accepted their diagnostic source here.

A bounded diagnostic profile should begin with a pinned collector/exporter metric
source and exact instrument/resource identity, then demonstrate an actual
refusal, retry or queue/drop signal. Keep source-query failure, collector health,
record event age, ingestion delay, sampling and completeness separate. An old
record or an empty service view must not become an inferred ingestion incident.

## Ranked next cuts

These ranks inform the planner; none activates work on its own.

1. **Make observation grain a normal service choice.** Deliver one HTTP server
   observation profile over existing query/measurement owners, alongside All
   spans. Acceptance: mixed client/server/internal spans, missing/legacy HTTP
   fields, duplicate observations and unknown outcomes remain explicit; drilldown
   and comparisons retain the exact profile; no automatic fleet percentile or
   request-total claim. This improves the primary investigation question directly.
2. **Offer a standalone general-source runtime profile.** Start ordinary
   exploration, discovery and native inspection without demo HTTP/operation
   contracts. Add a readonly source-compatibility check with useful table/column/
   capability errors. Acceptance: a non-Go ordinary-OTel fixture works with no
   manifest; enabling a malformed optional package still refuses that package
   explicitly; no synthetic schema compliance is inferred.
3. **Close a selected runtime-control gap.** First produce an up-to-date
   effect/admission/cancellation matrix, then adopt one executing family outside
   the current POST budget and specify proxy cancellation behavior. Acceptance:
   capacity and disconnect/cancel interleavings preserve no-publication/fresh-read
   boundaries, route-specific error contracts and tail semantics. Shared-user
   fairness remains gated on a deployment/identity design.
4. **Expose a bounded definition-to-target capability report.** Cover existing
   families and target states honestly, preserving supported YAML/Go semantics.
   Acceptance: an interactive definition exports or explicitly explains the
   preparation/unsupported boundary; no Databricks runtime certification or
   incidental CEL/KCL/DataFusion adoption. Keep the current offline compiler.
5. **Add one evidence-backed pipeline diagnostic when its source is available.**
   Instrument and capture a real local collector/exporter failure case before
   adding a health widget. Acceptance: actual instrument and basis are visible;
   absent, stale, unavailable and known-zero states remain distinct.

Trace-search request/result preservation may be a valuable product continuation,
but its priority should follow a demonstrated revisit/share journey. Saving the
request, freezing the observed aggregate/witnesses and making an independent
fresh capture are different operations. [The current trace-search contract](TRACE_SEARCH.md#continuation-and-custody)
explicitly ships none of those new persistence forms. This review does not make
it the platform prerequisite for all other work or recommend a new database for it.

## Minimum shared-deployment contract

The current supported target is a loopback-bound, trusted local lab. Before a
shared pilot, name the hosting environment, identity provider, user/workspace
population and data ownership model. A reverse proxy by itself is not the whole
contract. The minimum single-replica pilot needs:

- TLS/session authentication and server-enforced permissions for source scopes,
  definitions, captures, annotations, exports and every referenced artifact.
  Business `tenant_id`, OTel environment/namespace fields and unguessable IDs
  cannot substitute for principal authorization. Prevent bypass through direct
  Go ports; use a dedicated least-privilege readonly ClickHouse query identity,
  separate from ingestion/admin credentials. Define trusted OTLP ingress and
  server-owned tenant/source partitioning; SDK-supplied resource claims are not
  proof of ownership. [Current lab account boundary](KNOWN_ISSUES.md).
- Verified authorship and query/cancellation ownership. Today the budget IDs and
  result tokens are process-local capabilities, and notes use
  `local-user-unverified`; see [query budget](../cmd/goo/query_budget.go#L13) and
  [annotation scope](WORKBENCH_ANNOTATIONS.md#cross-product-metadata).
- Per-principal/workspace resource limits and an explicit effect policy for all
  source reads, tail streams, retained evidence and publication. Query-count
  admission is not a storage quota or a sensitive-data export policy.
- A documented data-root/volume layout, retention/deletion/redaction rules,
  disk-pressure behavior and an exercised backup/restore plan. Current atomic
  immutable-file publication and verified reopen are valuable, but do not prove
  crash or disaster durability; [publisher policies](../internal/immutablefile/publish.go#L11)
  intentionally distinguish directory-sync and file-only callers.
- One-replica/restart behavior for exact live-page tokens, tail membership and
  comparison drafts. [Result snapshots](../exploration/result_snapshots.go#L13)
  have bounded memory and expiry. Do not announce horizontal scaling until
  routing/ownership, recovery and capability isolation work across replicas.
- Deployment/version/configuration provenance and operational diagnostics.
  Explicitly pin the OTel exporter schema/adapter and optional XTDB topology;
  credentials, migration, upgrade and restore procedures need real acceptance.

These requirements activate for that named target; they are not a request to
add generic authentication, clustering or a storage migration to the local UI
milestone without a deployment decision.

## Gates for longer-term research

| Direction | Existing useful foundation | Required before implementation/claims expand |
| --- | --- | --- |
| Databricks/lakehouse | Prepared-fact count/ratio compiler and lifecycle receipts | A selected runtime/profile and Unity Catalog permissions; actual generated DDL/query execution; proven prepared grain, evaluation-basis selection and quality propagation; physical timestamp/UInt64/nested OTLP ingestion mapping and parity. Local YAML parsing is not warehouse acceptance. |
| Bitemporal metadata | Immutable revision branches, exact citations and optional XTDB retained evidence | A historical valid-time/system-time query that file revisions cannot reasonably serve, selected topology, transaction/read-basis guarantees, retention and restore acceptance. Retain telemetry evidence separately from evolving assertions; do not migrate all metadata to obtain a temporal query. |
| XTDB external ClickHouse ingestion | Existing bounded direct-write retained-evidence adapter | Stable source identity, snapshot/stream handoff, late inserts, mutation/delete/TTL and replay semantics, plus actual source/XTDB acceptance. The reviewed external-source mechanism is ingestion, not transparent federation. A completed-capture feed is narrower than mirroring all telemetry. |
| Classifier/training export | Tags, notes, retained exact rows and review history | A named consumer/task, vocabulary and review policy, set-versus-member labels, immutable membership, feature projection, sampling provenance and leakage/group split policy. Tags or query predicates do not automatically label every future match. |
| Shared investigations/multiplayer | Immutable ordered documents, scoped notes and exact evidence links | Shared-deployment authority first, expected-version/branch conflict policy, verified authorship and failure recovery. Presence and CRDT text editing do not establish artifact access or resolve publication conflicts. |
| Durable alerts/SLOs | Typed kernels, definitions and lifecycle/evidence semantics | Scheduler/evaluation ownership, lateness/correction policy, durable state, notification delivery/recovery and a named production deployment. On-demand previews are not an alerting service. |

Supporting research: [warehouse](DATABRICKS_METRIC_VIEWS.md),
[bitemporality](BITEMPORALITY.md), [XTDB/Minigraf](ANNOTATION_STORAGE_SPIKE.md),
[training](TRAINING_ANNOTATIONS_DESIGN.md), and
[collaboration](COLLABORATIVE_INVESTIGATIONS_RESEARCH.md).

## Reconcile older plans

- Keep delivered API/OpenAPI, discovery dashboards, exact Activity retention,
  native HTTP/messaging plans, typed integer/decimal operators, annotations,
  trace search and trace comparisons out of the new backlog as fresh builds.
  Any follow-up should name a missing behavior rather than repeat their titles.
- Replace stale broad "C2 active" and source-family inventory prose only after
  checking current handlers/OpenAPI. In particular, the older
  [HTTP boundary table](HTTP_BOUNDARIES.md#endpoint-effects) predates trace search;
  its omission is documentation drift, not a missing runtime route.
- Preserve the prepared-fact compiler's limited portability claim, but do not
  carry old wording about all OTel/lifecycle adapters being future work into
  the revised roadmap: the [shared measurement adapter](MEASUREMENT_PACKAGES.md)
  now executes real bounded OTel/native/lifecycle paths. That does not complete
  warehouse preparation or cross-target execution for every definition family.
- Retain source annotation/weave work as an optional reviewed input. Defer broad
  library integration coverage claims and aggregate annotations; aggregate policy
  stays in packages, and the [manifest](GOO_ANNOTATIONS.md#manifest-and-opt-in-weaving)
  describes source-package declarations rather than the actual linked binary.
- Do not activate broad package resolution, arbitrary expression runtimes,
  schema alias migration, universal storage/result objects, native exponential
  histograms, freeform multiplayer or a warehouse connector solely because this
  review exists. Each needs its own bounded user/source contract and acceptance.
