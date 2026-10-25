# Whole-product planning reconciliation — 2026-10-25

Status: independent planning input at `9e07b88`, not an implementation authorization
or a second execution backlog. The current user request authorizes review and a
revised plan. No application code, tests, browser campaigns or deployments were
changed or executed for this review. Source inspection and prior recorded
acceptance establish the dispositions below; they are not new runtime evidence.

## Recommendation

Make Goo easier to adopt and more precise about the population being observed
before extending the newest trace feature's persistence. The product loop remains
Activity → scope a population → compare/inspect → explain/preserve → reuse. The
next finite round should improve this loop for an operator with ordinary OTel,
rather than add another independent analysis destination.

Use five bounded cuts below, with explicit prerequisites. Treat verification,
source-cost measurement and task-level interaction consistency as delivery gates,
not a new framework project. Pipeline diagnostics need an evidence-source decision
before a product implementation commitment. Exact observed trace-search custody
is not a prerequisite for these cuts.

## Review inputs and source reconciliation

The [product review](WHOLE_PRODUCT_PRODUCT_REVIEW_2026_10_25.md),
[architecture review](WHOLE_PRODUCT_ARCHITECTURE_REVIEW_2026_10_25.md),
[usability review](WHOLE_PRODUCT_USABILITY_REVIEW_2026_10_25.md),
[verification review](WHOLE_PRODUCT_VERIFICATION_REVIEW_2026_10_25.md) and
[platform review](WHOLE_PRODUCT_PLATFORM_REVIEW_2026_10_25.md) are inputs to the
parent's revised roadmap. They are not five independent queues. This review also
read the [Oct24 roadmap](ROADMAP_2026_10_24.md),
[backlog](BACKLOG_2026_10_24.md), [implementation history](IMPLEMENTATION_PLAN.md),
[Oct22 design review](WHOLE_PRODUCT_DESIGN_REVIEW_2026_10_22.md),
[Oct24 architecture review](WHOLE_PRODUCT_ARCHITECTURE_REVIEW_2026_10_24.md),
[earlier usability review](WHOLE_PRODUCT_USABILITY_REVIEW.md) and
[chart/detail review](UI_CHART_DETAIL_REVIEW.md).

Three concrete findings determine the order:

- Generic investigation is manifest-free at query time, but
  [`serve_http.go`](../cmd/goo/serve_http.go) still requires the bound HTTP
  inventory/library inputs and defaults to the checkout operation package.
  [The platform review](WHOLE_PRODUCT_PLATFORM_REVIEW_2026_10_25.md) distinguishes
  this bootstrap gap from the correctness of generic investigation.
- [`dashboard.go`](../exploration/dashboard.go) explicitly warns that all observed
  spans include mixed kinds and retries and are not request throughput. Existing
  HTTP/RPC/database/messaging suggestions change useful dimensions; they do not
  silently establish a request-grain denominator.
- [`filter-values.js`](../ui/filter-values.js) needs complete counted groups but
  posts an ordinary exploration query. [`Source.Fetch`](../exploration/source.go)
  also reads rows and fields. A one-row limit bounds the returned page, not the
  count/group scans or number of source reads. This is a source-backed reason to
  measure and then add a lean projection, not an unmeasured performance claim.

## Enabling gates

| Gate | Concrete deliverable and acceptance | Scheduling boundary |
| --- | --- | --- |
| **G1 — Reproducible fast regression gate** | Extend the existing API CI with root Go, all Node and relevant Python/proxy tests. Execute the new trace-search handler wire tests rather than relying on reflected schemas: the current workflow's fixed regex omits `TestTraceSearchHTTPWireFixtures`. Keep frozen wire fixtures. A clean checkout must run the documented command. | Start before source implementation; this does not require a fresh formal model or browser campaign for each local edit. |
| **G2 — Honest native/integration evidence** | Pin the nested Go/toolchain, sibling quint-go and npm/Quint requirements; report optional jobs as passed/failed/skipped with reason. Maintain a thin fresh-lab spine using existing ordinary-OTel investigation, explicit expansion, capture/offline reopen, saved configuration staging, Activity return and one positive optional-store branch. | Follow G1 in bounded steps. Root `go test ./...` does not test nested modules. Existing local evidence remains historical, not continuous CI. |
| **G3 — Measured source effects and read cost** | Record logical API effects separately from actual source read counts and, where available, ClickHouse rows/bytes/time for the touched source families. Establish a baseline before claiming a performance improvement; compare equivalent populations and outputs. Audit admission/cancellation for the touched executing route, including GET/capture paths, rather than assuming the POST budget applies everywhere. | Baseline before C2/C4 source changes. Adopt one uncovered owner if evidence identifies a gap; do not start a universal transport rewrite or invent scan-cost numbers. |
| **G4 — Consistent operator task** | On each touched entry, keep one primary action and explicit Stage / Run / Capture / Save definition effects. Preserve original scope, paused return, stale-owner refusal and keyboard focus. Review actual desktop/narrow frames; reuse passive components. Rename misleading existing capture actions when touched. | An acceptance gate for C1–C5, not a whole-app aesthetic rewrite. A bounded generated hold/edit/cancel/return owner schedule is preferable to repeating a fixed phase model for every button. |

Sources: [CI](../.github/workflows/api-contract.yml),
[verification constraints](WHOLE_PRODUCT_VERIFICATION_REVIEW_2026_10_25.md),
[HTTP policies](HTTP_BOUNDARIES.md), [query admission](../cmd/goo/query_budget.go).
No gate claims authentication, production tenancy or power-loss durability.

## Ranked finite cuts

### C1 — Start with an ordinary source, without sample contracts

**User outcome:** connect an existing compatible OTel/ClickStack source and reach
Activity/Services without preparing reservation/checkout declarations.

Add an explicit generic exploration startup profile. Keep optional declared
HTTP/operation/native families strict when enabled; missing packages should not
block unrelated generic routes. Reconcile Go startup and the Python lab's eager
sample-package assumptions. Keep the sample lab as an explicit example profile.

Acceptance: fresh configuration with no sample contracts starts generic reads;
unsupported optional families have a named unavailable capability; explicitly
configured invalid packages still fail clearly. An ordinary non-Go, manifest-free
OTel journey reaches Activity → Services → fixed inspection. No source write,
automatic package inference or silent binding substitution occurs. G1 applies;
the fresh-lab part of G2 should use this profile.

Non-goals: arbitrary exporter-schema compatibility, production installation,
automatic credentials, removing the Python proxy or deleting the examples.

### C2 — Choose a meaningful service population

**User outcome:** answer a common service question without mistaking client,
internal and retry spans for independent incoming requests.

Keep All observed spans. Add one explicit HTTP server-span observation
lens with named predicate, grain, outcome policy and convention eligibility.
Resolve the exact initial HTTP attributes/current-versus-legacy policy before
implementation. Keep recorded span Error separate from HTTP-status policy.
Reuse existing charts, service/operation diagnosis and native capability plans.

Acceptance: mixed-kind, retry, missing/empty namespace and missing-convention
fixtures distinguish the populations. Labels say observed records/rate, not
verified request throughput. Carry the selected lens into operation/record/trace
continuations and back. Full totals remain independent of row limits. Native
HTTP latency is used only through its exact supported identity/unit/kernel;
span-derived percentiles are not merged with native histograms.

Dependencies: G3 baseline; C1 for clean adoption acceptance. Non-goals: deduplicated
business invocations, fleet RED, inferred success, general messaging/RPC policies
or replacing the distinct declared Operations contract.

### C3 — Continue from a native interval into record context

**User outcome:** follow a native-metric symptom to the relevant logs/spans
without rebuilding the environment, service and incident window.

Adopt one supported native stream/selected interval. Offer source-free context
preview and explicit fixed-question Run; retain the selected metric on return.
Elevate observed-instrument discovery over definition authoring in this touched
entry, using the existing discovery and native owners rather than another picker.

Acceptance: preserve exact environment/service and supported resource identity;
do not map arbitrary point dimensions to span attributes. Unknown instance or
namespace remains explicit. A gauge point requires a deliberate context interval.
Cover metric-only, logs-only, no matching records, stale selection and return.
Call this service/instance/time context, not causal correlation or an exemplar.

Dependencies: G4; C2 provides a clear destination population but is not a license
to invent HTTP semantics for arbitrary metrics. Non-goals: exemplar ingestion,
automatic metric/log joins, unit conversion or loosening native stream identity.

### C4 — Read only what a counted-value/service task needs

**User outcome:** narrowing a real population remains responsive without paying
for unused record pages and field discovery.

Introduce a narrow complete count/group projection behind its own explicit
contract, first adopting counted values and one measured service consumer.
Preserve full-source counts, canonical fields, read clocks, source disagreement
and cancellation. Protect existing full exploration/capture response bytes.

Acceptance: independent SQL/source observations demonstrate which reads were
removed and equivalent count/group results on the same fixtures. Retained
staging/selection/capture behavior is unchanged. Test the existing 1,000-group
refusal and useful next actions. A future high-cardinality summary must be a
separate explicit partial-summary type: top-N is not a complete Cohort.

Dependencies: G1/G3 and C2 semantic fixtures. This is an optimization/projection
cut, not authorization for approximate counts, a query engine, a universal
result object, a new generic request owner or an unmeasured latency claim.

### C5 — Reuse a trace-search request, if the recurring task warrants it

**User outcome:** reopen/share the same independent span-condition question
without reconstructing its clauses.

Extend an existing typed configuration/definition path. First decide fixed versus
relative time semantics and how the definition appears in the existing catalog.
Preserve all conditions, qualified identities and versioned limits. Reopening and
editing stage intent only; explicit Run can return a different result.

Acceptance: canonical Go/JS/YAML round-trips, malformed/deep conditions refused,
same-title identities distinct, no telemetry/publication during reopen/edit, old
ordinary Query definitions unchanged, and one actual repeated-investigation task.
Do not make representative witnesses look like captured complete traces.

Dependencies: G1/G4 and a named recurrence/share journey. Do not block C1–C4 on
this. Non-goals: a new Saved destination/store, persisted paired trace-page
comparison, exact observed trace-search result custody, subscriptions or alerts.
If no recurring task is identified, stop after the reviewed design rather than
automatically ship another resource kind.

## Disagreements and stop/go decisions

**Trace persistence first:** rejected as the default ordering. Product,
architecture, usability, platform and verification reviews identify broader
adoption, grain, continuation or cost gaps. Definitions/captures/documents already
provide substantial reuse. A saved request and exact observed-result custody are
different contracts; C5 addresses the former conditionally, not both by default.

**Pipeline diagnostics:** product ranks this highly, but platform/architecture
correctly require a real collector/exporter observation source and failure path.
Authorize a bounded source-feasibility proposal only if chosen in the revised
roadmap: define source, identity/unit/temporality and independent observation path;
demonstrate normal → actual failure → unavailable before UI. If not observable,
record the blocker. Event age and missing records cannot become a healthy badge.
This gate does not silently add a sixth implementation cut.

**High-cardinality facets:** architecture proposes an explicit summary contract.
C4 first removes unnecessary reads while preserving complete-group refusal.
Partial/top-N summaries deserve a later separate decision, not an incidental
change to comparison membership or full-population denominators.

**Whole-product usability cleanup:** retain the usability review's action
consolidation, answer-first Saved/native/Operations composition and observed-first
native entry. Apply these through G4/C3 and one named later surface, not an
unbounded component sweep. Lack of a contextual trace-search button in Activity
or service inspectors is an entry-consistency gap, not absence of trace search.

## Prior work: shipped, superseded and still open

| Earlier request/theme | Disposition at `9e07b88` | Revised planning consequence |
| --- | --- | --- |
| M1–M5 guidance, publication winner, anchor continuity, ordinary recovery, orientation | Delivered in the Oct24 backlog with reviewed commits. | Do not schedule again; same-tab recovery is not shared-draft collaboration. |
| Pure protocol, immutable publication, HTTP/OpenAPI consolidation | Bounded cuts delivered. Current API is 76 operations/70 Go; older counts are historical. | Broaden executed/default gates and touched admission coverage; no repeat mechanical migration. |
| Stream-first handoff, Saved task landing, retained visits, selection/rest | Delivered bounded adoption in Activity/population/Saved contracts. | Refine touched action placement and answer-first composition; do not replace the owners. |
| Numeric filters/measures/histograms/comparisons/decimal scale/contextual plots | Delivered with explicit exact parsing, coverage and source/custody boundaries. | No new numeric engine or picker sweep. Floats/conversion require a new contract. |
| Population explanations, Activity explanations, recorded error/exception triage | Delivered; counts are observed records and independent reads. | Reuse these surfaces. Pipeline evidence and business outcome policies are still distinct gaps. |
| Log patterns/time comparison; trace timing/page comparison/search | Delivered through journey 186 with bounded/page/witness limits. | No critical-path or complete-trace inference; request reuse remains C5, observed custody separate. |
| Automatic service views, operation diagnosis, OTel HTTP/messaging plans | Delivered; not a verified request/fleet denominator. | C2 selects grain; C3 closes a selected-native-context seam. |
| Annotation/training/bitemporal research | Tags, exact references, reviews and optional retained evidence exist. | Reviewed label tasks/dataset export need a consumer, vocabulary, leakage and membership policy; no XTDB migration as a proxy for that decision. |
| Databricks/IaC/CEL/KCL/DataFusion | Typed YAML and offline prepared-fact translation exist. | Live Databricks needs an instance/source acceptance; no general evaluator/planner replacement in this round. |
| Diagnostic disk cleanup | Diagnostic budgets are separate from evidence retention; prior cleanup shipped in bounded forms. | Preserve capture/custody data; scheduler/system-table limits need explicit ownership and measurements. Not a feature prerequisite by default. |
| Shared deployment/auth/tenant quotas/backup/restore | Not established by local-lab evidence. | Name deployment/identity/storage topology and rights across linked evidence before promising shared production use. |
| Multiplayer canvas, alerts, RUM, profiler, generic dashboard builder | Deferred, not accidentally completed by current investigations. | Require a concrete user task and durable scheduling/error/concurrency semantics before activation. |

The parent should make a new dated roadmap/backlog authoritative, link this review
as an input and mark the Oct24 queue historical/complete. Keep earlier acceptance
records intact. Correct current-facing contradictory “missing” claims by linking
the delivered contract, not by pretending the historical review was wrong when
written. The next implementation begins only after the user accepts the revised
finite queue; this planning review itself changes no product behavior.
