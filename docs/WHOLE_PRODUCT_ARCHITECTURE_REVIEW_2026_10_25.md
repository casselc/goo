# Whole-product architecture review — 2026-10-25

Reviewed baseline: `9e07b88`. This is a read-only review of source, contracts, prior reviews and recorded acceptance. No implementation, tests, browser campaign, deployment or new performance measurement was performed. Recommendations are inputs to the revised roadmap, not an activated implementation queue.

## Recommendation

Remove the demo-only startup prerequisite in a bounded parallel enabling cut. Make the existing service workspace explain **which observation population it measures**, then make its aggregate reads cheaper and more useful at ordinary attribute cardinality. Save one reusable investigation definition after that workflow is coherent. Do not start by persisting every recently added transient feature.

Goo already has a substantial general observability core: ordinary OTel record exploration, native metric kernels, service discovery, qualified attribute filtering, numeric distributions, independent population comparisons, log patterns, trace-wide existential search, page interval analysis and immutable evidence. Its next architectural problem is composition and population grain, rather than another evaluator, datastore or top-level product surface.

The strongest spine remains Activity → scoped population/service lens → explicit investigation → evidence and authored context. Services should provide a useful second entrance to that spine. A generic mixed-kind span summary is honest, but does not directly answer a common question such as “which incoming operations are slow?” A saved trace-search definition would improve continuity without fixing that entry problem.

## What is established, and what is not

### Source-backed observations

| Observation | Evidence | Architectural implication |
| --- | --- | --- |
| Generic record/metric routes are registered only after `serve-http` requires prepared binding, inventory, instrumentation scope/version and loads a default checkout operation package. Python eagerly imports the reservation package through `grammar`. | [`serve_http.go:19–55,74–83`](../cmd/goo/serve_http.go#L19), [`server.py:13`](../server.py#L13), [`grammar.py:5`](../grammar.py#L5) | Ordinary OTel is usable semantically, but the documented general-purpose bootstrap still needs demo-specific configuration. This is a confirmed adoption constraint, not an assertion the legacy lab is broken. |
| Service record questions restrict service, namespace, environment and window, without selecting span kind. The default span widget groups by name and requests span-duration/status summaries. | [`dashboard.go:71–76,113–146`](../exploration/dashboard.go#L71) | Mixed server/client/internal/retry spans can occupy one overview. This is explicitly labeled `span-record`, not a confirmed counting bug. A separate explicit observation lens is feasible with the existing AST. |
| Service native widgets already execute typed measurements. There are at most two selected instruments; defaults follow deterministic instrument sorting, not an HTTP incident-purpose ranking. | [`dashboard.go:150–206`](../exploration/dashboard.go#L150), [`service-dashboard.js:38–46`](../ui/service-dashboard.js#L38) | Reuse the planner and native presenter. Do not build another service metrics engine or call observed histogram records a request count. |
| HTTP and messaging instrument candidates, dimension presence, missing/empty instance diagnostics and exact stream identity already exist. The HTTP capability profile pins v1.30.0 and marks verification `not-run`. | [`instrument_capabilities.go:8–11,74–145,164–224`](../measurements/instrument_capabilities.go#L143), [E1 acceptance](E1_ACCEPTANCE.md) | Semconv-aware service work should adopt these capabilities. Discovery eligibility is not verified interval coverage, conformance or a valid fleet denominator. |
| A separate prebuilt catalog pins v1.37.0 profiles. Its prepared HTTP expansion requires an explicit binding/policy; prepared requests are not arbitrary raw OTLP. | [`prebuilt/catalog.yaml:3–58`](../prebuilt/catalog.yaml#L3), [`expand.go`](../prebuilt/expand.go), [`http.go`](../prebuilt/http.go) | These are different supported contracts, not interchangeable “latest OTel.” Keep rule versions explicit; no silent alias union or common-version upgrade. |
| Counted field values already read full-scope groups through `/api/explore`; they are not merely sampled suggestions. Generic exploration additionally reads rows and discovered fields. | [`filter-values.js:5–24`](../ui/filter-values.js#L5), [`source.go:130–220`](../exploration/source.go#L130), [`source.go`](../exploration/source.go) | Do not rebuild full-source counted values. A narrow aggregate-read projection can remove unnecessary reads from this existing feature and future service lenses. |
| Ordinary grouping refuses more than 1,000 groups. Explanations make a field partial/unavailable rather than rank truncated groups. | [`source.go:212–218`](../exploration/source.go#L212), [`explanation.go:409–465`](../exploration/explanation.go#L409) | Honest refusal is implemented. A high-cardinality summary would be a new explicit contract, not permission to present the current first 1,000 groups as a complete distribution. |
| A normal exploration combines independent count, row, group, discovery and optional series reads. Explanations use up to fourteen reads with three-second child deadlines. Trace search instead uses one aggregate statement for qualifying IDs, total and witnesses. | [`source.go:133–218`](../exploration/source.go#L133), [`explanation.go:362–465`](../exploration/explanation.go#L362), [`trace_search.go:89–127`](../exploration/trace_search.go#L89) | Reuse actual consistency contracts. A common UI shell must not turn separate reads into an asserted snapshot. |
| Current execution defenses are substantial: four admitted controlled requests, bounded cancellation registry, route/child deadlines, source timeout/read/memory limits, and a compiled trace-search SQL limit. Admission still names effectful POST routes explicitly. | [`query_budget.go:28–107`](../cmd/goo/query_budget.go#L28), [`source.go:65–95`](../exploration/source.go#L65), [`trace_search.go:113–115`](../exploration/trace_search.go#L113) | Preserve budgets. Route admission coverage and query cost are related but distinct; a bounded result list does not make a full-source aggregation cheap. |
| Browser proxies retain their own method/path/body/timeout policies. OpenAPI and frozen fixtures document compatibility rather than eliminate these differences. | [`server.py:42–134`](../server.py#L42), [API contract](API.md), [`openapi.json`](../api/openapi.json) | Add policy conformance when touching a route. Removing Python is not a prerequisite for a service lens or definition save. |
| Stream Activity polls a bounded rolling event-time window. It is not an ingestion cursor. Trace page timing/comparison analyze supplied rows; trace search evaluates a different, full-source distinct-ID population. | [`tail.go:14–63`](../exploration/tail.go#L14), [trace timing](TRACE_PAGE_TIMING.md), [trace comparison](TRACE_COMPARISON.md), [trace search](TRACE_SEARCH.md) | Preserve those separate grains. Neither a page miss nor a bounded tail establishes telemetry absence or complete traces. |

### Confirmed defects versus hypotheses

This review confirms documentation contradictions, not a new blocking runtime defect:

- [`KNOWN_ISSUES.md:337`](KNOWN_ISSUES.md#L337) says investigation drafts persist only in the current page, while [`investigation-draft.js`](../ui/investigation-draft.js) and the delivered M4 contract implement bounded explicit same-tab reload recovery.
- [`KNOWN_ISSUES.md:741`](KNOWN_ISSUES.md#L741) says full-source value ranking is unsupported, while [`filter-values.js`](../ui/filter-values.js) explicitly requests complete counted groups. The remaining issue is cardinality/expense, not the absence of counted lookup.
- The older issue-list API inventory says 75 operations; a source inventory of `api/openapi.json` contains 76, including `/trace-search`. Historical acceptance totals should remain historical, while the current inventory should have one authoritative source.
- The roadmap has correctly appended delivered work through trace search, but still contains multiple “Current continuation” headings for already delivered N1/N2. The older implementation proposal and review tables remain historical input, not independent active backlogs.

Potential duplicate execution policies, dense browser owner composition and repeated decimal SQL guards are **design risks**, not proof of a current stale-response defect or slow production deployment. The prior decimal timeout is recorded separately from successful reproductions; this review does not reinterpret it as a proven optimizer regression. Likewise, the recent nullable-witness tie defect is already corrected and reviewed; it is not unfinished work.

## Preserve the semantic layers

The present package boundaries are useful:

1. `exploration/`: record predicates, physical record lowering, record populations, full-source trace search, and supplied-page models.
2. `metrics/` and `measurements/`: exact native stream identity, unit, temporality, reset, interval coverage and supported compositions.
3. `measurement/`: portable aggregates over explicitly prepared, typed subject facts. [`compile.go:10–20,65–84`](../measurement/compile.go#L10) labels the artifact `requires-prepared-source` and does not deploy it.
4. `annotations/` and `instrumentation/`: source declarations/build weaving and runtime OTel instruments. [`manifest.go:40–50,75–83`](../annotations/manifest.go#L40) binds source revisions; [`runtime.go:16–33,36–80`](../instrumentation/runtime.go#L16) emits named standard OTel signals with explicit goo attributes/outcomes. A declaration is not runtime coverage, and goo operation outcomes are not a generic business outcome policy for arbitrary OTel spans.
5. `workbench/`, immutable captures and optional retention: authored context, reusable configuration, observed evidence and exact external citations, each with its own authority.
6. Passive components and pure browser display models: supplied-data display and bounded intent. Application owners retain source effects, navigation, cancellation and publication.

Go/JavaScript mirrors for page geometry and comparison need one normative contract plus shared complete fixtures, rather than an HTTP service for every local calculation. Server-authoritative query predicates, evidence validation and immutable identity must not move into an unreviewed browser convenience layer. The new trace-search field-evidence projection is a good reusable physical-field seam; do not copy map-key conventions into its renderer.

## Definitions, observations and fresh reads

These meanings already differ intentionally:

| Artifact | Authority and lifetime | Required behavior |
| --- | --- | --- |
| Reusable exploration/Activity/package definition | Immutable typed configuration; no captured statistics | Opening stages intent. An explicit Run obtains a new observation. [`ExplorationView`](../workbench/exploration_views.go#L22) already says this directly. |
| Live buffer or retained application visit | Bounded connection/tab-local observations | Pause/Back preserves supplied context; no automatic source replay or implied completeness. |
| Process-local result token | Expiring capability over server-produced bytes, not durable custody | Existing pool: two-minute TTL, 64 entries, 64 MiB total and 2 MiB per result. [`result_snapshots.go`](../exploration/result_snapshots.go#L34) must not silently accept an 8 MiB search result merely because another endpoint can return it. |
| Immutable capture | Validated observed result with original question, source basis/read time and separate custody time | Reopen stored bytes without telemetry reads; refresh produces new evidence. [`capture_store.go`](../exploration/capture_store.go#L135) and explanation promotion preserve that distinction. |
| Saved baseline scope | A saved artifact provides selectors/window for an explicitly fresh comparison | Do not label the resulting reread as the original stored counts. Keep originating reference provenance separate from the new read basis. |
| Authored document/review | Branching authored context and exact references | Restore intent separately from publication; revisions are not a mutable collaborative head. [`documents.go:103–165`](../workbench/documents.go#L103) preserves that boundary. |

Trace-search definitions and observed search results have neither the same size nor the same custody requirements. A definition-only save fits the existing workbench scale. Persisting its exact result would require an explicit typed result contract, byte budget, promotion/idempotency and reopen policy; it must not be smuggled into a generic record capture or browser-submitted “observed” JSON. Existing raw-record captures can already supply page observations for future source-free trace display; that does not require persisting derived timing/comparison arithmetic separately.

## Ranked next cuts

Each cut is finite and independently reviewable. Cut 1 is a small enabling track that can proceed alongside the main functional priority, cut 2. Cut 3 should follow the lens contract or adopt counted facets first. Later cuts need their own approval and concrete design.

### 1. Explore-only startup without demo operation bindings

**User outcome:** point a local Goo instance at compatible ordinary OTel tables and use Activity, Services, record/native exploration and saved artifacts without inventing a reservation/checkout package or prepared HTTP inventory.

**Bounded implementation:** add one explicit startup profile to the existing Go process; retain the legacy profile and its old startup validation. Assemble generic source/store/route owners independently of optional prepared-fact and operation bindings. Do not manufacture empty declarations or a fake prepared expansion. A route that requires an unconfigured family must return an explicit not-configured/unsupported result or remain deliberately absent according to a documented capability policy; it must not report a healthy empty operation catalog as evidence of runtime absence.

The browser entry must work in that same profile. Decouple Python static/proxy startup from the eager `grammar.PACKAGE` reservation-file load; keep the Python example query/compiler available when its example capability is configured. Make unavailable Operations/Examples explain their configuration boundary, while generic data discovery stays usable. Keep `PACKAGES_URL`, ClickHouse source dependencies and store directories explicit. This does not require removing Python, adding authentication, deploying a new service or converting all configuration into a universal schema.

**Gates:** startup from a directory without demo operation/binding/inventory files; generic source/store routes registered; explicit unsupported legacy routes; static assets/API docs and proxy work without the reservation package; an ordinary non-Go OTel fixture can open Activity, scope a service and execute a fixed query. Misconfigured optional packages fail clearly when that capability is requested. Old lab startup, declared operation routes and response fixtures remain compatible. Source-dependent startup probes must not become implicit expensive population queries.

### 2. Explicit service observation lenses

**User outcome:** from a service, choose incoming server operations or a named protocol view, inspect latency/error evidence at that declared grain, then continue to existing logs, trace search and waterfall owners.

**Bounded implementation:** extend the existing dashboard planner and view adapter with an explicit lens. Keep generic all-span activity available. Start with one incoming HTTP server-span lens using an explicit supported kind/key profile; group by reported operation or route while preserving missing/empty values. Call it “HTTP server span observations,” not request throughput, unique requests, roots, transactions or business outcomes. RPC/messaging and client/dependency lenses should be additional explicit profiles, not one blended denominator. The first implementation should activate one lens, not multiply every lens into simultaneous queries.

Reuse existing HTTP/messaging capability diagnostics and exact native measurement packages. A native `http.server.request.duration` candidate can be offered alongside spans when eligible, with its own stream grain, units, interval coverage and basis. Native histogram observations, span statuses and log severities must not share an inferred success/error denominator. No automatic fleet percentile or missing-instance relaxation is included.

**Refactoring tied to the feature:** separate pure lens planning from widget execution in `dashboard.go`; centralize the small profile-to-predicate mapping beside the Go source/field adapters. Adopt existing supplied-data summary and continuation components rather than another landing panel. Preserve namespace and original fixed-window selectors in every continuation.

**Gates:** mixed server/client/internal and retry observations; missing/empty route and attributes; ordinary manifest-free OTel; semconv version mismatch; invalid native coverage; same service name in distinct namespaces; true zero versus missing signal; source-free planning/staging; explicit Run; no added hidden source reads on local lens changes; old dashboard behavior/fixtures remain compatible or get an explicitly versioned change. A real journey must follow service lens → narrowed operation → trace search/waterfall and return to the paused parent.

### 3. Lean aggregate execution for service and counted facets

**User outcome:** loading a service aggregate or counted values does not also fetch an irrelevant inspection page and field inventory.

**Bounded implementation:** introduce a narrow record-aggregate execution entry point used by exactly two callers first: the new service lens and counted-value lookup. Reuse `Query` validation, compiled count/group/series expressions, physical adapters and result validation. Specify selected reads and per-read bases explicitly; leave `Source.Fetch` and old response bytes unchanged behind their compatibility contract. Do not emulate an inspection result with fake empty rows.

A first slice can support count plus one grouped distribution, retaining the current 1,000-group refusal. Add time series only when the service lens needs them. Projection is a source-family application seam, not a universal evaluator shared with native streams or prepared facts. A narrow shared request-owner lifecycle is justified only if those two application adapters still duplicate begin/invalidate/current/dispose; it must not auto-subscribe or retry.

**Bounds and gates:** retain existing source timeout, memory/read limits, cancellation and global admission. Assert the actual issued read set, exact counts/groups, missing/empty and changed-between-reads states; compare typed parameters and numerical/source parity against the existing path. Measure representative elapsed time and compiled SQL size before making performance claims. Record a held-response/edit/navigation test and preserve legacy capture bytes. Decimal histogram alias/subquery optimization remains a separate measured follow-up, not an incidental rewrite during this cut.

### 4. Explicit high-cardinality facet summaries

**User outcome:** an ordinary high-cardinality attribute does not force the user to abandon scoping after a truthful 1,000-group refusal.

**Bounded design first:** provide a dedicated exact top-K categorical summary, with full-population denominator and an explicit remainder count, or choose another bounded semantic contract after comparing costs. It must say which groups are shown and which population was measured. Do not reuse the complete `Cohort` contract for truncated groups, rank a sample as a population, or silently drop the remainder. Exact selected-value drilldown can stage the existing predicate; remainder drilldown needs its own bounded exclusion/membership plan and must refuse if the predicate budget cannot express it.

Build on cut 3 instead of adding another full `Source.Fetch`. Start with one field and a small fixed K, explicit missing/empty groups and one source-supported rank policy; no approximate sketches, general chart algebra or automatic anomaly attribution. If obtaining exact top-K exceeds the existing source budget, return unavailable rather than widening budgets.

**Gates:** more than 1,000 actual values; ties; missing and empty; selected-count/remainder conservation; full denominator versus rendered list; late source changes; excessive values; Unicode keys; preserved parent numeric predicates and namespace; no forged complete comparison from a partial summary. Real ClickHouse evidence is required before claiming this works at useful cardinality. This is a confirmed capability gap, not a claim current bounded grouping is incorrect.

### 5. Save and reuse one trace-search definition

**User outcome:** keep a useful multi-condition investigation question and reopen it staged, without remembering its conditions or mistaking it for saved evidence.

Start with a versioned definition containing the exact typed trace-search request and a bounded title. Reuse the existing immutable workbench mechanics and definition-library presenter; keep a distinct family adapter. Absolute fixed bounds should stay absolute unless the user explicitly chooses a separately modeled fresh-window action. Opening, sharing, restoring and saving configuration perform no telemetry query. Do not automatically pin witnesses or persist local trace comparison pairs in the same cut.

An observed-result save is a later distinct decision: the search response can reach 8 MiB, while the current server-held observation pool admits only 2 MiB per entry. Do not widen shared token limits or accept client-supplied proof statistics as a shortcut. Reuse the explanation promotion/idempotency pattern only after choosing a typed budget and failure policy.

**Gates:** old stores/IDs unchanged; source-free create/reopen; exact predicate/binding/window round trip; title/revision collisions; malformed schema; source unavailable on reopen; explicit Run updates the observation but not the stored definition; canceled/late source responses cannot replace staged intent. This is lower priority than a clear service observation lens, unless user evidence establishes a pressing continuity need.

### Conditional follow-up, outside the five-cut queue: pipeline/source diagnostics

**User outcome:** distinguish an application with no observed records from an unavailable query source or an actual reported collection/export failure.

Start with bounded query-source diagnostic categories and, only with an actual pinned collector instrument source, one collection/export failure view. The current collector configuration wires OTLP pipelines but does not ingest its own diagnostic metrics into those tables. There is no authority to infer exporter drops from record age or absent service metrics. Do not add a global green health indicator.

Reuse source transport and native kernels. Retain a sanitized bounded error category, affected operation and clock; unknown errors stay unknown. Collector counters require a declared scope/version/unit/reset contract and a real induced failure/recovery acceptance case. This cut is conditional on that named source gate; it is not a reason to deploy a new monitoring backend.

## Refactoring opportunities that should travel with those cuts

- **Aggregate projection and validation:** cut 3 can retire repeated unnecessary reads from two actual callers. Keep complete-distribution accounting in Go and state/basis vocabulary explicit. Avoid a new configurable query planner or a `Result` union containing every family.
- **Capability descriptors:** cut 2 can consume the existing instrument plan rather than repeat HTTP name/unit checks in JavaScript. Separate profile identity, source-shape match, kernel eligibility and verified read coverage. Do not merge v1.30 metadata profiles with v1.37 prepared profiles by string replacement.
- **Effect policy:** the next touched route should update one reviewed route/effect inventory and direct/proxy conformance. A later small route-registration descriptor can generate admission/documentation assertions, while explicit family handlers retain status mapping, SSE and body-limit policies. No router/framework migration is needed.
- **Result composition:** the service lens should reuse scope/state/summary/actions from one existing fixed-record composition, then delete the superseded local markup. Avoid a generic owner that silently equates native source results, held page arithmetic and immutable receipts.
- **Readability:** several browser adapters are dense despite low line counts. Format or split only touched application boundaries in separate mechanical changes. Fewer physical lines or more shared helpers do not establish correct lifetime ownership.
- **Catalog mechanics:** file catalogs deliberately refuse scans beyond 10,000 entries, and direct IDs remain openable ([capture catalog](../exploration/capture_store.go#L492), [workbench catalog](../workbench/store.go#L79)). Introduce an index only when a real catalog workload reaches that limit; do not migrate all artifact families for hypothetical scale.

## Reconcile prior unfinished work

Treat these earlier proposals as follows when constructing the new backlog:

| Prior proposal | Current disposition |
| --- | --- |
| Pure question protocol extraction, shared source HTTP mechanics, startup dependency injection, immutable-file mechanics, read-only saved presenter, bounded draft recovery | Delivered; adopt the existing boundaries. Do not reopen as broad cleanup. |
| Numeric attribute predicates, numeric/distribution comparisons, decimal64, contextual plotting, explanations, log patterns, page timing/comparison, trace existential search | Delivered bounded contracts. Extend them only for a named user workflow; preserve old receipt intent and wire omission. |
| Universal result composition or universal async owner | Reframe as two-caller adoption tied to cuts 2–3. No whole-product replacement. |
| Automatic HTTP/RPC/messaging/USE dashboards | Partly delivered discovery/capabilities and service widgets. The next missing piece is an explicit grain/policy lens, not another metadata registry or assumption all planned prebuilt features execute. |
| General full query-session persistence, collaborative whiteboard, training-label export | Deferred pending a concrete recovery/collaboration/consumer contract. Existing authored documents and tags do not imply a collaborative mutable head or training-dataset validity. |
| Trace-search save, stored page comparison replay | Genuine continuity gaps, but definition reuse and observed evidence are separate cuts. No blanket persistence-first dependency. |
| Dynamic native fleet membership, missing-instance execution, exponential histograms, exemplars | Separate semantic work with actual instruments and coverage tests. Never fill missing identity with an empty shared instance. |
| Databricks live execution, XTDB external ClickHouse ingestion, production weaving rollout | External gates remain open. Offline generation, source review and a local example are not substitutes. |

The revised roadmap should provide one finite active queue, and link delivered historical sections rather than accumulate another list of contradictory “Next” headings. Repair the confirmed issue-list contradictions in that documentation closeout; retain recorded acceptance evidence and original artifact IDs.

## Named external gates and limits of this review

- **Databricks execution gate:** a named workspace/catalog, authorized deployment target, prepared fact ingestion and runtime comparison against real Metric Views. Current generated artifacts do not establish Databricks acceptance or raw nested OTel/event translation.
- **XTDB integration gate:** an actual supported XTDB release/configuration and concrete record-custody/external-source acceptance. No source review here proves external ClickHouse ingestion, cross-store transactions or safe production deletion.
- **Native instrumentation gate:** actual exporter shapes, full OTel resource/scope identity, instance policy, temporality/reset/bucket coverage and an agreed instrument set. Broader names in a discovery catalog do not validate the kernel input.
- **Shared deployment gate:** named target and approved authentication, tenant authorization, query-source credentials, quotas, redaction, backup/retention and concurrency ownership. The local loopback/disposable lab account is explicitly not that deployment model ([known issues](KNOWN_ISSUES.md#L213), [`compose.yaml`](../compose.yaml#L20)).
- **Collector diagnostic gate:** pinned self-telemetry instruments and an induced, observed failure/recovery. No current record/tail timestamp proves end-to-end ingest health.

This review has not benchmarked real customer volumes, interviewed users, audited every publisher fault mode, or rerun the recorded formal/browser checks. Generated properties, observed trace/model validation, negative controls and real journeys are valuable acceptance layers; each proves its stated model or finite schedule, not exhaustive runtime interleavings or telemetry completeness.

Do not introduce DataFusion, another frontend framework, CEL/KCL execution, a universal authoring schema, a fleet database, a new bitemporal store, a service split, distributed collaboration or automatic statistical anomaly claims as incidental work. Keep Rust/Go, authoring-language and backend-target choices separate from a measured need. The next useful simplification is a runnable generic entry and a smaller, clearer path through the existing kernels.
