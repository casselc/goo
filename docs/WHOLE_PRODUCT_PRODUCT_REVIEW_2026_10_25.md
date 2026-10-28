# Whole-product product review — 2026-10-25

Status: source-backed review of `9e07b88`. This is a planning input, not an
implementation authorization or a second roadmap. It examines the whole product,
its prior reviews and recorded acceptance. No application, tests, browser session
or deployment changed for this review. References to shipped behavior mean the
bounded contracts documented and implemented at that baseline, not production
readiness or a new usability study.

## Product judgment

Goo now has enough analytical depth to support a useful investigation. Its next
improvement should help an operator choose the right meaning and move between
existing signals, rather than add another specialized trace tool or chart family.
The organizing job remains: **see what is happening in my system, isolate a
population that differs, inspect supporting evidence, and make the explanation
and recurring question reusable.** Ordinary OTel must work before a manifest.

The largest semantic gap in the default service experience is the difference
between **all observed spans** and an explicitly chosen **server/request grain**.
The largest cross-signal interaction gap is the final step from a selected native
metric observation/interval to a correctly scoped span or log question. These are
broader user outcomes than making the newest trace-search panel durable.

Saving trace searches is worthwhile, but it should not automatically be first.
Current definitions, captures, documents, recovery and annotations already supply
substantial reuse infrastructure. Another isolated saved object would improve one
recent feature while leaving the service's first useful question and a metric-led
incident path weaker. A small reusable trace-search configuration is recommended
below; exact observed-result custody is a separate decision.

## Five user outcomes to optimize

| Outcome | Already shipped | Remaining product gap |
| --- | --- | --- |
| **1. Understand now in the system I own.** Start with recent logs/spans, scope to environment/service, see an interpretable first summary. | Stream-first Activity, compact fields/ancestry, service discovery and automatic multi-signal overviews. | The default span overview is deliberately mixed-kind; it cannot answer a request-grain question merely by relabelling its count. Pipeline availability has no operational evidence source yet. |
| **2. Isolate what differs.** Narrow a range/population, compare another window/version, see distinguishing fields. | Time/range gestures, selected/rest operations, typed integer/decimal filters, distributions, exact numeric comparisons, field-share explanations, log patterns/time comparisons. | Selection and continuation controls still need task-based consistency checks; unsupported or high-cardinality cases need useful next actions, not another competing editor. Statistical anomaly detection is not necessary to use these tools. |
| **3. Follow a symptom across signals.** Move from service/metric/operation to supporting logs and traces without losing context. | Service diagnosis, native discovery, exact stream focus, trace/log expansion, reported links, contextual follow-ups, trace-page breakdown/comparison and multi-condition trace search. | A selected native interval currently exposes identity, buckets and reasons, but not an immediate scoped record continuation. A universal cross-signal correlation or exemplar identity is not delivered. |
| **4. Preserve and communicate what was learned.** Reopen the same evidence, cite it, explain uncertainty and continue later. | Immutable captures, exact citations, investigation revisions, trace-anchor diagnostics, browser-local ordinary draft recovery, optional XTDB retained evidence and annotation history. | Evidence capabilities remain target-specific; not every displayed result can be frozen without rereading. Cross-store global search, reviewed training exports and shared draft concurrency are not present. |
| **5. Turn recurring understanding into shared definitions and better telemetry.** Reuse meaning without rebuilding an investigation. | Record/native packages, canonical YAML, definitions library, ordinary OTel discovery, optional reviewed Go annotation/weave handoff and offline portability work. | New trace-search requests are ephemeral. Production IaC lifecycle, arbitrary expressions and live Databricks execution remain separate contracts. Instrumentation suggestions do not verify runtime coverage. |

Source anchors: [Activity owner](../ui/stream-entry.mjs),
[service planner](../exploration/dashboard.go),
[population owner](../ui/population-workspace.mjs),
[native result owner](../ui/measurements.mjs),
[trace search](../ui/trace-search.mjs),
[documents](../ui/documents.mjs),
[definitions](DEFINITION_LIBRARY.md),
[annotations](WORKBENCH_ANNOTATIONS.md), and
[source annotations](GOO_ANNOTATIONS.md).

## Ranked findings

### 1. Make grain explicit before adding more service summaries

The service planner's own warning says its counts are captured records, not
request throughput; generic spans include retries and mixed kinds
([`exploration/dashboard.go`](../exploration/dashboard.go)). The default overview
combines span activity, duration/status summaries, operation groups, logs and up
to two native instruments. This is useful and already implemented. The missing
primitive is **an explicit supported grain/policy selection**, not another table.

A server-span lens would remove client/internal double counting from a common
first investigation while retaining honest terminology. Start with generic
reported Server spans and a bounded HTTP-server profile where the required
conventions are actually observed. Display **observed server-span records/rate**
and **recorded errors**; these are not complete request throughput, unique
invocations or business outcomes. HTTP status and recorded span Error need
separate named policies. A conventional instrument name alone cannot choose an
error denominator or prove OTel conformance.

Reuse the existing pinned HTTP native candidate
`http.server.request.duration` only when its exact metadata and supported kernels
apply. Keep its unit and histogram range semantics. Do not silently combine a
native histogram with span-derived latency or treat absent native input as zero.
RPC, messaging producer/consumer work and logical retry-free operations should
remain later profiles or declared contracts, rather than a universal RED label.

The existing operation diagnosis and HTTP/RPC/database/messaging suggestions are
**delivered** ([core refinements](CORE_EXPERIENCE_REFINEMENTS.md),
[`ui/service-browser.js`](../ui/service-browser.mjs)). They should be reused, not
scheduled again under a new name.

### 2. Close a native-metric-led incident path

Native support is considerably stronger than an old vendor-gap list suggests:
gauges, reset-aware counters, explicit histograms, exact stream focus, coverage
reasons, timelines/heatmaps, supported comparisons and immutable captures exist.
Service-to-discovery already preserves its incident window
([native discovery](NATIVE_INSTRUMENT_DISCOVERY.md),
[native navigation](NATIVE_STREAM_NAVIGATION.md)).

The concrete remaining seam is visible in
[`ui/measurements.js`](../ui/measurements.mjs), `renderTimeline`/`choose`: selecting
an observation/interval displays exact bounds, stream attributes, bucket counts
and reasons. It does not offer a direct span/log continuation from that selected
interval. Existing capture follow-ups in
[`exploration/followups.go`](../exploration/followups.go) are useful foundations;
this is not a request to rebuild their catalog or claim all follow-ups are absent.

Offer a compact **Inspect span context / Inspect log context** continuation for
one exact supported native stream and interval. Carry the environment, service,
known namespace/instance identity and explicit time bounds. Only map fields whose
cross-signal resource meaning is known; point dimensions are not automatically
span attributes. Say **service/instance and time context**, not causal correlation
or an exemplar. A gauge point needs an explicitly chosen surrounding interval;
it must not acquire invented interval coverage. Keep preview source-free and Run
explicit, with return to the retained metric selection.

### 3. Supply evidence about data availability before implying operational trust

An empty Activity/service/native read may mean inactivity, a selection mistake,
unsupported instrumentation, sampling, delayed export or a failed pipeline.
Goo correctly preserves these unknowns. Repeating caveats cannot identify which
case occurred.

The collector configuration currently defines OTLP ingestion and ClickHouse
export pipelines, not a self-metric scrape/export path
([`config/collector.yaml`](../config/collector.yaml)). The service-discovery
warnings explicitly state that event timestamps and signal counts are not ingest
lag, pipeline health or completeness
([`exploration/service_discovery.go`](../exploration/service_discovery.go)).

A bounded collector/exporter diagnostics contract is still relevant from the
Oct24 review. First establish a real source and an observable failure scenario.
Then reuse native instruments to show supported queue/refusal/export-failure
observations with their clocks, units and temporality. Monitoring a failed export
path solely through that same failed path is not sufficient acceptance. A missing
instrument is unavailable, never a healthy green status. This is an explicit
operational-evidence slice, not a topology dashboard or a universal health score.

### 4. Finish recurring-question reuse selectively, without inventing custody

Trace search now has a coherent contextual start, independent per-span conditions,
full distinct reported-ID count and bounded qualified witnesses
([trace search contract](TRACE_SEARCH.md),
[`exploration/trace_search.go`](../exploration/trace_search.go)). Its request is
not an ordinary one-span Query and cannot be shoved into that schema without
changing meaning. Closing/reloading currently loses this configuration.

The bounded next reuse improvement is a versioned trace-search **configuration**
that reopens staged and requires another explicit Run. Use existing definition or
question infrastructure where its typed extension is appropriate; avoid a new
primary destination or unrelated store. Preserve common environment/window and
each independent condition, including the original selector inside condition one.

Saving a request is not preserving the observed result. Trace search has no new
server-held result capability or capture kind at this baseline. Do not publish
browser-supplied witnesses as independently verified evidence, silently reread
under a “Save observed” label, or imply complete trace retention. An actual user
need to cite this exact observed search should select the separate immutable
custody design. Existing waterfall finding capture/anchor diagnostics remain
available, with their independent-read boundaries.

### 5. Product consistency needs whole tasks, not another permanent toolbar

The recent trace-search acceptance found a roughly 1,800-pixel inherited filter
editor before results. Its shipped compact disclosures and accepted-draft guard
fixed that concrete problem; the lesson applies to the product, not a request to
repeat that implementation everywhere.

Keep a small primary decision visible, then reveal authoring and exact provenance.
In particular, the Measurements opening still introduces a Define → Preview →
Investigate package workflow before displaying data
([`ui/measurements.js`](../ui/measurements.mjs)). That expert workflow is valid,
but service-scoped discovery and observed instruments should remain the natural
entry for an unfamiliar system. Do not add another generic YAML wizard just to
make discovery feel less package-centric.

Apply consistency through the two incident tasks above: one clear selected scope,
UTC/human units, one selected detail, appropriate source-free preview and explicit
execution, and a reliable return. Compact exact values and an accessible receipt
are both necessary. Do not remove local unknown/subtotal labels that determine
whether a comparison is meaningful; move repeated general caveats into a single
nearby disclosure.

## Finite next-round recommendation

These are proposed cuts in ranked order, not claims of shipped behavior. The
planner should select a finite round rather than activate every historical idea.

| Rank / cut | Benefit and bounded scope | Acceptance | Prerequisites and non-goals |
| --- | --- | --- | --- |
| **P1 — Explicit server-span service lens** | Answer the common service-entry question at a meaningful grain. Add a generic Server-span lens and one HTTP-server capability profile inside Services, using existing charts/diagnosis. | Mixed server/client/internal/retry fixtures demonstrate the chosen grain; labels say observed records/rate and named error policy; generic overview remains available. Missing conventions fall back or refuse explicitly. Namespace absent/empty/value states remain distinct. Native HTTP duration uses exact supported identity/unit/kernel when available. Full totals remain independent of page limits. One ordinary non-Go OTel journey covers overview → operation → records → trace → return. | Reuse existing service planner/capability contract and independent source owners. No new landing page, complete throughput claim, percentile mixing, messaging RED or inferred logical request identity. |
| **P2 — Selected native interval to record context** | Complete a metric-led investigation with one supported stream/window continuation into spans and logs. | Selected interval/identity preview makes no source call; explicit Run reads exactly the named context. Original native selection survives return. Unknown namespace/instance and unsupported dimensions never become invented exact correlation. Gauge points require explicit interval choice. Test logs-only/metric-only/no-matching-records and stale selection/navigation. | Reuse existing native point inspector, scope planning and record staging. No exemplars, automatic correlation, unit conversion, fleet aggregation or metric-to-span attribute guessing. |
| **P3 — One real pipeline diagnostic source and failure** | Distinguish supported operational pipeline evidence from no telemetry. Begin with source/contract and prove a local failure before UI. | Exact observed metadata/unit/temporality; normal → actual failure → unavailable transitions; explicit reads/clocks; no failure-to-zero fallback. If the evidence source cannot observe the failure independently, stop with an honest validated plan rather than ship a healthy badge. | Requires actual collector self-metric availability and a separate viable observation path. No external paid instance is required in principle, but local capability must be established first. No generic health dashboard or sampling-completeness inference. |
| **P4 — Reusable trace-search configuration** | Make a recurring cross-service condition question portable and reopenable. Extend an existing typed definition/question path, source-free staging and explicit Run. | Canonical request round-trip preserves all independent conditions/window/namespace values; duplicate titles do not replace identity. Reopen/edit/export triggers no telemetry or publication. Old ordinary Query definitions remain unchanged. A rerun can honestly produce a different answer. | Reuse existing catalogs/protocol fixtures; select exact resource shape with architecture review. No exact observed-result capture, training dataset, universal search schema or extra Saved destination. |

A small documentation closeout accompanies these cuts: keep one authoritative
finite queue and mark contradictory old gap rows historical. It should not become
a large implementation milestone that displaces a real operator task.

## Reconciliation with earlier proposals

| Earlier theme | Status at this review | Planning consequence |
| --- | --- | --- |
| Oct24 measure guidance, publication conflict, trace-anchor continuity, ordinary draft recovery and orientation | Delivered M1–M5 in the active roadmap/backlog. | Do not re-add them as unfinished. Browser-local recovery is not shared draft collaboration. |
| Oct22 stream-first Analyze handoff, Saved task landing, retained visits and selection/restoration | Bounded adoption delivered in `ACTIVITY_ANALYSIS.md`, `SAVED_WORKSPACE.md`, `POPULATION_WORKSPACE.md` and `CORE_EXPERIENCE_REFINEMENTS.md`. | Preserve those owners/returns. Broader “one workbench” language does not authorize replacing them. |
| Earlier chart/detail and vendor surveys | Numeric distributions, aligned comparisons, population explanations, exception detail, patterns, trace timing/comparison and trace search now shipped. | Stop treating charts/trace detail as the main general-observability gap. Assess the next operator decision instead. |
| Service defaults/OTel discovery | Automatic overviews, operation diagnosis, instance navigation and native candidate plans shipped. | Missing piece is explicit semantic grain and selected-context continuation, not another dashboard catalog. |
| Annotation/training proposal | Shared tags/notes, exact references, optional retained evidence and review history shipped; reviewed label tasks, pinned dataset membership/export and classifier integration remain proposed. | Require a real label vocabulary/consumer and leakage policy before dataset tooling. Do not substitute a metadata-store migration for that product decision. |
| Collector/exporter diagnostic proposal | Still a relevant unmet operational-evidence task; source prerequisite unresolved by Activity event age. | Retain as P3 with a hard source/failure gate. |
| Databricks/CEL/KCL/DataFusion/IaC expansion | Offline prepared-fact translation and typed YAML exist. Live Databricks, richer expressions and production plan/apply remain separate. | No reason to insert a query engine or replace the semantic core for the four proposed cuts. |
| Multiplayer canvas, full alerting, RUM/replay, profiler, general dashboard builder | Not delivered and not necessary for this round. | Keep deferred. Alerting requires durable scheduled evaluation and an explicit unknown/error policy, not just a notification button. |

Some old gap prose is now directly contradictory: `PRODUCT_GAPS.md` still mentions
no live tail in a historical row, and older blocks in `KNOWN_ISSUES.md` describe
saved workspaces or general selection/comparison as absent while later sections
record their implementation. Preserve history, but label it historical or replace
its current-facing index with capability-specific links. README's current entry
is substantially improved; do not restart its original rewrite simply because
older reviews requested it.

## Remove, consolidate, retain and stop

- **Consolidate continuation and action vocabulary through actual tasks.** Stage /
  Run / Capture / Save definition / Save finding have different effects; keep
  those differences and remove duplicate controls for the same effect. Reuse
  passive record/field/comparison/interval displays; source ownership stays in
  application adapters and Go semantics.
- **Demote raw-ID and expert authoring paths, not their capabilities.** The Saved
  task landing already does this. Keep exact-ID recovery, YAML and receipts in
  disclosures; avoid moving them back above telemetry.
- **Retain Operations while it supplies distinct declared contracts.** It is not
  replaced by a generic Server-span lens. A future navigation demotion requires
  parity for package contracts, cross-service usage and exact old links first.
- **Retain Examples as executable task entries, not another broad product.**
  Instrumentation samples and shared example packages are useful; do not add
  reservation-only analysis panels or duplicate definition libraries.
- **Stop expanding the component catalog without a second real consumer.** New
  passive components have delivered value; adapt existing field/table/chart/
  inspector contracts before adding another custom element or option matrix.
- **Stop defining the next task as the newest feature's next missing button.**
  Trace-search persistence is a legitimate P4, but service/request meaning,
  metric-led context and operational evidence affect more users and signals.

## Boundaries that need a separate decision

This repository remains a loopback local workbench. Authentication, tenant-scoped
source credentials, query quotas, production retention/redaction/deletion,
backup/restore and audit need a named shared-deployment target
([README limits](../README.md#operational-limits-and-stopping),
[`server.py`](../server.py), [known issues](KNOWN_ISSUES.md)). If the next goal is
shared deployment, that gate takes priority over P1–P4; it should not be hidden in
a UI polish milestone. Query cancellation and read-only settings are not an
access-control model.

The fixed-window/two-hour and bounded-input/group limits are also real product
limits. Do not promise large historical fleet analysis, arbitrary floating-point
expressions, full trace accumulation or statistical anomaly detection from the
current lab acceptance. A named scale/history use case should drive their next
architecture work. Optional XTDB remains useful for retained annotation history;
it need not become the default telemetry/query backend.

## Review gate for the next round

For each selected cut, reviewers should follow an ordinary OTel incident from its
entry to a preserved explanation, not merely inspect the added control. Record
what changed in the user's decision, exact scope/read/custody behavior, readable
compact desktop/mobile states, and unsupported/missing evidence. Refresh affected
storyboards/screenshots and Bombadil actions; add Hegel/state and Quint trace
checks for changed semantic or asynchronous contracts. Existing supported native,
manifest-free, declared/weaved and saved-definition paths should be represented
across the round without rerunning every legacy suite after every local edit.

This review recommends **meaningful defaults and connected investigations before
more analytical breadth**. It does not ask for a new framework, universal result
schema, required manifest, required bitemporal store or new top-level workbench.

## User steering: automatic discovered-system dashboard families

The user subsequently clarified the desired first initiative: **automatic USE,
RED and related dashboards for discovered OTel services, Kubernetes, database,
web and messaging systems**. This supersedes the *breadth* of the initial P1
recommendation above. Retain that initial review as history: HTTP/server-span
semantics are a sensible first executable slice, not the whole dashboard vision.
The revised roadmap should make **capability-driven automatic dashboards** its
first product initiative and deliver it in bounded family slices. It should not
turn this request into a manual dashboard builder or another package wizard.

### Common product contract

A discovered entity should open a useful default overview from the telemetry
actually available. Profiles select supported questions and displays; users should
not have to know metric names, write YAML or supply a goo manifest to obtain that
first overview. Manifests may add explicit meaning, expected signals or operation
grain, but cannot fabricate observations.

Reuse the existing service dashboard planner and execution owner
([`exploration/dashboard.go`](../exploration/dashboard.go),
[`ui/service-dashboard.js`](../ui/service-dashboard.mjs)). Extend the capability
plan with family/entity identity, eligible widget questions, exact prerequisites
and unavailable reasons. Keep metadata **candidate**, executable supported binding,
and actual read state distinct. Pin each convention rule set; a matching metric
name without kind/unit/temporality/resource identity is insufficient.

Default composition should be small: the most useful supported activity/latency,
resource-pressure and error/reliability panels, with shared scales and compact
selected detail. Partial data should produce useful partial dashboards. Show a
short explanation for a missing essential question and put the full prerequisite
inventory in a disclosure; do not replace telemetry with a wall of empty tiles.
A failed independent widget stays unavailable without zeroing successful widgets.
No automatic default may imply a common transactional read or complete telemetry.

### Proposed family matrix

All family additions below are **proposed execution profiles**, except the stated
existing foundations. Each must route into current fixed analysis, native stream
inspection, logs/traces and saved definitions rather than become a separate app.

| Family / discovered entity | Useful automatic defaults when supported | Evidence and eligibility requirements | Partial-data behavior and boundary |
| --- | --- | --- | --- |
| **HTTP/web service RED** | Observed server-span activity/rate, duration distribution and named recorded-error/status policy; route/operation breakdown. Prefer a supported native HTTP duration instrument when its semantics are appropriate. | Reported Server role plus actual HTTP convention evidence for HTTP labels; exact service/namespace/environment. Existing `http.server.request.duration` candidate requires supported histogram kind/unit/schema/identity and actual execution. Counts/latency from spans and native histograms remain separate sources. | Generic Server-span profile works without HTTP dimensions; generic mixed-span overview remains available. Missing HTTP status does not become a zero error ratio. Retries/sampling do not become unique or complete request throughput. |
| **RPC service/client RED** | Reported server or client operation rate/duration/status by actual RPC system/method; separate perspectives rather than double-counting both ends. | Explicit role and supported RPC conventions, protocol-specific named error policy, exact source metadata for native instruments if used. | Missing RPC evidence retains generic server/client observations. Span Error is not automatically an RPC status policy. No inferred end-to-end request count. |
| **Database client and pool** | Client operation latency/volume/recorded errors by database system/operation; pool used/idle/limit, pending acquisition and wait duration where supported. | DB client span evidence and explicit role; independently discovered pool instruments with exact kind/unit/dimensions and supported gauge/counter/histogram kernels. Ratios require an actual compatible limit/denominator. | Client spans describe client-side observations, not database-server CPU, locks or total query throughput. Pool metrics can appear without traces and vice versa. Missing pool capacity cannot produce utilization. Database-server health requires its own source/profile. |
| **Messaging producer/consumer and broker** | Separate send/process volume/duration/recorded outcomes; backlog/queue depth/consumer lag and processing capacity when actual instruments support them. | Actual messaging system/destination/operation-role dimensions and per-stream identity; preserve message/byte/time units and temporality. Existing `messaging.process.duration` candidate is a foundation, not proof of consumer success or broker coverage. | Producer, consumer, process attempts and broker queues are different grains. No inferred one-to-one delivery, acknowledgment success, business completion or retry-free denominator. Broker metrics require broker evidence; lag and queue depth are not interchangeable. |
| **Host/runtime USE** | Supported CPU/memory utilization; run-queue/wait/GC or thread/worker saturation signals; supported resource error counters. Show raw usage/capacity when a ratio is not justified. | Exact host/process/runtime identity, resource and library schema, units and actual compatible capacity limits. Finite gauge values are already supported; fractional monotonic counters and derived ratios need separate semantic/kernel support where absent. | CPU busy alone does not prove saturation; memory usage without limit does not imply utilization. Per-host/process streams are not a fleet average. Missing instance identity stays unsupported rather than receiving a fabricated service instance. |
| **Kubernetes cluster/workload/pod/node** | Cluster/workload inventory and readiness/restarts where actually sourced; pod/container CPU/memory vs requests/limits; node capacity/pressure; scoped service logs/traces and resource drilldowns. | Dedicated entity inventory/ownership model with cluster identity, namespace and stable workload/pod/container/node IDs, plus supported cluster/node/container metric sources. Resource attributes on service telemetry supply context but do not establish complete cluster inventory. | A pod metric without cluster ownership must remain a scoped observation, not an invented workload rollup. Missing metrics/limits/owner refs remain partial. Service namespace and Kubernetes namespace are different qualified resource keys. Restart counts require compatible counter semantics; current service discovery is not Kubernetes discovery. |

Examples of metric names above are capability candidates under a pinned profile,
not permission to fuzzy-match aliases or coerce units. The existing HTTP rule set
is pinned, and the messaging candidate is explicitly experimental
([`measurements/instrument_capabilities.go`](../measurements/instrument_capabilities.go)).
New DB/RPC/runtime/Kubernetes rules need the same versioned treatment and ordinary
OTel fixtures, including wrong unit/kind/schema and missing dimension controls.

### Foundations already shipped versus actual new work

- **Already shipped:** all-signal service discovery, service identity including
  missing/empty/value namespace handling, automatic service overview planning,
  bounded concurrent independent widget execution, partial failure behavior,
  observed HTTP/RPC/DB/messaging dimension suggestions, compact operation
  diagnosis, native instrument discovery, generic kernels and HTTP/messaging
  metadata capability candidates. These should not be renamed as new work.
- **Missing execution layer:** automatic family selection at the correct entity
  and grain, complete supported profile questions/default displays, usable
  prerequisite explanations, and contextual routing from those defaults. A
  candidate returned by discovery is not yet an automatically executed family.
- **Missing Kubernetes foundation:** a dedicated cluster/workload/pod/node entity
  inventory, ownership resolution and corresponding metric bindings. The existing
  service catalog cannot safely stand in for it.
- **Specific numeric gates:** native gauges already accept finite Float64
  observations (`measurements/source.go`); claiming all floating-point metrics
  are unsupported would be wrong. Current counter input requires nonnegative
  integral values below 2^53. Fractional CPU-time counters, unsupported
  temporality/bucket types, derived capacity ratios and metric streams lacking
  the current required service-instance identity need explicit work. Record
  attribute decimal support does not solve these native metric contracts.

### Staged delivery recommendation

1. **Shared family capability contract plus HTTP/server first execution.** Extend
   the current planner, execute the first bounded family without authoring, and
   prove generic fallback/partial data. Retain source-free planning, actual read
   clocks and automatic supported widget execution already used by Services.
2. **RPC and DB client/pool, then messaging roles.** Reuse current span conventions
   and native candidates; add only the supported questions for each profile.
   Ship useful span-only and metric-only partial dashboards, not an all-or-nothing
   requirement for every instrument. Broker/server observability is a separate
   profile from application client observations.
3. **Host/runtime USE with narrow numeric prerequisites.** Begin with supported
   finite gauges and integral counters. Add fractional counter or ratio semantics
   only for a named profile/instrument, with kernel/reference/source parity.
   Never make a USE profile silently downgrade unsupported input into zero.
4. **Kubernetes discovery and one vertical pod/workload path.** Establish stable
   entity identity and a real source before broad cluster aggregation. Deliver a
   useful pod/workload overview and scoped record continuation first; expand node
   and cluster rollups only with explicit ownership, coverage and aggregation
   contracts. Unsupported identity is visible rather than patched with aliases.

Acceptance should exercise an actual heterogeneous ordinary-OTel deployment:
discover entities → open automatic defaults → inspect one pressure/error/latency
population → follow scoped native/record context → preserve a finding → return.
Include a sparse deployment, one unavailable source, missing conventions and an
unsupported metric alongside a valid one. A manifest-free service and optional
manifest enrichment must both work. Family changes should update the common
storyboard/component usage and capability evidence; no family requires a new
primary navigation destination or a manual layout editor.

This revised first initiative changes the initial ranking: service family
automation is the umbrella; the native-context continuation remains part of its
useful investigation path, and collector diagnostic evidence remains a separately
gated source-health profile. Reusable trace-search configuration is still useful,
but it should not displace the user-requested automatic discovered-system
experience or become a prerequisite for these dashboards.
