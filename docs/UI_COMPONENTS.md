# Known issues and gaps

[Diagnostic retention](DIAGNOSTIC_RETENTION.md) now expires raw browser recordings, with a 512 MiB soft budget, active/pinned protection and explicit Bombadil outcomes. Compact ledgers/results/logs remain. Legacy outcomes use a 72-hour policy; direct Playwright invocations require explicit pruning. ClickHouse Docker console logs now rotate at 10 MB × 3 files; internal system-table and file-log retention remain separate.

Latest diagnostic cleanup recovered space from 509 MB to 6.4 GB by hardlinking
byte-identical raw browser screenshots and removing unused build cache. All
screenshot paths/bytes, traces and application evidence were retained.

Local disk maintenance: internal ClickHouse text/stack-trace/processor-profile
histories grew beyond 5 GB, in addition to a large debug file, and exhausted the
shared development filesystem. The named diagnostic histories were cleared and
the current debug file was reduced to its latest 20 MB; application OTel tables,
query audit/error logs, saved evidence and retry data were preserved. Docker console logging is now bounded at 10 MB × 3 files. Internal system-table
and file-log retention budgets still need a separate configuration change. See
[actual cleanup evidence](../VALIDATION.md).

[C2 HTTP boundaries](HTTP_BOUNDARIES.md) inventories current route effects/limits.
All Go HTTP ClickHouse endpoint/credential/client configuration is now resolved by
one startup snapshot and injected into native, record, tail, capture, operation GET
and investigation owners. Optional XTDB URL configuration is likewise injected
into retention/document citation owners, while request/card read bases remain
per-operation values. Configuration status does not prove database availability;
document files reopen without the optional store. Native and record source HTTP mechanics
now use shared POST/bounded-read primitives; the [acceptance audit](HTTP_BOUNDARIES.md#c2-acceptance-audit--2026-10-22)
identified four additional readers; all are now adopted, including histograms.
Telemetry/native/fixed exploration JSON is adopted behind frozen wire cases.
The [OpenAPI contract](API.md) now inventories 64 operations and documents direct-Go
versus browser-proxy limits, headers and SSE differences. Those discrepancies are
preserved, not repaired by documentation. Further mechanical JSON migration remains paused;
[executed API acceptance](API_ACCEPTANCE.md) now covers proxy and saved/retained flows. Dynamic telemetry row
payloads remain open; structural schemas do not enforce all domain/capability rules.
The local query/packages images now serve the current docs/spec and backend.
Remaining C2 endpoint/branch coverage is open; selected deployed API and browser
acceptance does not prove every route or fault mode. Non-JSON upstream HTTP errors
now return JSON 502 rather than disconnecting; documented limit/header differences
remain. Docker still reports the existing XTDB container unhealthy despite successful
native-store and review/citation checks; healthcheck status is not availability proof.
ClickHouse/XTDB environment mutation after startup is not a reload.
Live comparison drafts remain registration-local and expiring; persisted captures
reopen without a source, while old draft tokens do not survive a new owner.

[Bounded C1 publication consolidation](IMMUTABLE_PUBLICATION.md) is delivered for
all six Go publication paths. Family validation, identity, catalogs and duplicate
policies remain separate. Custody and native YAML retain file-only sync (no directory
sync). Custody verifies business identity but does not detect a valid integer edit
to its received clock. Native YAML POST retains its existing success response for a
corrupt existing winner without repair; registry GET then refuses that file.
Investigation capture loading retains permissive unknown-field behavior while
exploration capture envelopes remain strict. These compatibility boundaries are
pinned, not hardened by the refactor. No power-loss durability is claimed.

## Current annotation retention boundaries

Retained evidence now has tag/environment/literal-service/capture-time browsing
and exact review citations in documents. Catalog scans and browser pending-request
storage are bounded; multi-tab journal coordination and server-owned recovery are
not implemented. The [active consolidation queue](IMPLEMENTATION_PLAN.md) now tracks the reviews. Separate live/saved operation ownership and compact applied population composition are implemented; the first bounded live entry now has a compact scope/action bar and collapsed mobile navigation. Typed live attribute filters and fixed-window staging are implemented. Paused-parent return is tab-local and source-free; reload does not retain it. Live/fixed record and detail presentation now share bounded supplied-parent context; it is not a complete trace. Held service observations now share the record-table shell. Saved record scope/provenance now shares a presenter with document references, and its fixed continuation stages before Run. Saved record pages now have a dedicated owner without a hidden query editor. Fixed/saved/context record tables now consolidate attributes into one inspector and hide empty comparison buttons. Standalone tails now share the record inspector; native stream identity shares a lazy inspector or an existing selected-point body. Native latest values and coverage now share a presenter across measurement details, service summaries, native-window comparisons and supported-window previews. Bounded S2/A3 adoption is delivered; fresh Activity defers fixed-editor mounting until explicit authoring/continuation, retaining drafts on return. Bounded B1 library consolidation and retention owner separation are delivered; retained catalog tags remain explicitly per-load current-head snapshots, distinct from exact pinned review citations; see the [acceptance audit](S2_A3_ACCEPTANCE.md). Other aggregate/native layouts remain D1/E1 work. See [Live Activity](LIVE_ACTIVITY.md) for current limits.

Optional XTDB retention now accepts selected exact live result pages as well as
saved captures, with idempotent initial/correction requests and await-token
propagation on fresh connections. Live pages expire after two minutes and are
process-local/capability-addressed; refresh or pagination clears selection. An
already committed save can be recovered after expiry with its unchanged request.
Native request intents now survive reload in a bounded browser retry journal;
ordinary document/note drafts remain page-local. Document citations carry a pinned
review and read lower bound; bare evidence links retain tab-local basis behavior. No production auth/redaction/deletion,
unified cross-store tag search, historical-cutoff UI, reviewed labels or dataset export
is implied. See [contract](WORKBENCH_ANNOTATIONS.md).


[Bounded page-record selection](RECORD_POPULATIONS.md) and inspector/shared-field population actions are implemented. Shared service/artifact tags and notes plus optional XTDB retention of selected **live-page or saved-capture** records are implemented in the [annotation slice](WORKBENCH_ANNOTATIONS.md). Exact live-page freezing is implemented; richer [whole-product review, training labels and dataset exports](PRODUCT_ANNOTATIONS_DESIGN.md) remain proposed.


Recent activity supports [staged categorical selections](FACET_SELECTION.md). General predicate complements, automatic difference scoring and durable selection history remain future work.


The [recent-entry canvas](RECENT_ACTIVITY.md) now includes
[horizontal time brushing and bounded retained child populations](RANGE_DRILLDOWNS.md).
[Raw-span heatmap rectangle selection](HEATMAP_SELECTION.md) is implemented for single duration-bucket split charts with at most 32 intervals. Cross-page/durable record selection, persistent branches,
arbitrary field columns and automatic anomaly/difference scoring remain proposed.
Saved exploration views retain their existing staged presentation and do not
automatically start live updates.

Population live tails are bounded recent event-time observations, not ingestion cursors or complete traces. Reconnect/resume exposes a continuity gap. Full-population live comparison uses independent reads over the baseline duration; its window may differ from the 60-second inspection tail. Saved settings never auto-subscribe. See [live contract](POPULATION_WORKSPACE.md#watching-a-population).

Service-overview time selections scope that overview and its span/log follow-up links. Supporting discovery inventory retains its original window. General cross-service span/log populations and saved exploration views are now supported in the [population workspace](POPULATION_WORKSPACE.md). Discovered per-signal breakdowns and saved baseline comparisons are supported; numeric cohort comparisons remain unsupported. Arbitrary native-stream composition and saved multi-widget dashboards remain future work. Native crossing intervals are omitted, never split or interpolated. See [contract](LINKED_TIMELINE.md).

Trace examples rank valid linked spans on a bounded record page, not complete traces or the full matching population. Findings link two later independent reads; automatic verification that an original anchor survived those reads remains future work. See [trace finding contract](TRACE_FINDINGS.md).

This list describes the current code at handoff. Items are observations or missing features, not completed fixes.

## Measurement correctness

| ID | Issue | Required direction |
| --- | --- | --- |
| M01 | summarize() derives the SLO preview from at most 1000 returned subjects | Separate full-population aggregation from paginated inspection |
| M02 | Completion lookup filters both starts and terminals using the same Timestamp lookback | Anchor eligibility on acceptance time and fetch related outcomes through the relevant deadlines and evidence cutoff |
| M03 | Pending subjects occupy part of the lookback instead of a fully shifted mature window | Implement explicit mature cohort bounds; expose pending as a separate preview population |
| M04 | toFloat64OrZero turns absent or malformed timestamps into zero | Validate required identities/timestamps and propagate invalid evidence |
| M05 | Duplicate acceptance conflicts are not detected; min/any collapse inconsistent records | Define immutable event IDs, scoped subject keys, and conflict rules |
| M06 | Completeness and clock uncertainty are not verified | Add declared assumptions and observable evidence-health inputs; retain demo-unverified until justified |
| M07 | The completion view can show reason 'no completion' for a pending subject | Make reason and classification coherent |
| M08 | The Python native metric view remains point listings; bounded Go operation counter rates and histogram intervals now exist | Use the bounded metric evaluators where supported; general metrics, automatic fleet discovery and SLOs remain separate |
| M09 | Embedded SQL tests use a minimal fixture schema and an older engine than the Compose image | Check actual exporter schemas and run real container integration tests |

## Language and resources

[Numeric attribute measures](NUMERIC_ATTRIBUTE_PLANNING.md) now support one explicit
int64/uint64 binding across totals, groups, charts, captures and YAML, with separate
coverage counts and capability diagnostics. Multiple measures, numeric cohort
comparisons, floating/decimal parsing and unit conversion remain unsupported.

grammar.py still uses view-specific query branches. Grouping fields and map bindings are hardcoded. definitions/reservations.json is a small demo config and has no complete resource schema or version resolver. The original EDN package and deadline-outcome syntax are illustrative. The [Go aggregate](PORTABLE_COMPILER.md) and [lifecycle](LIFECYCLE_PREPARATION.md)
kernels implement bounded typed models, preparation, reference evaluation and
aggregate backend compilation. [Shared measurement packages](MEASUREMENT_PACKAGES.md)
now connect real exporter data to the Go kernels, general captures and the UI
through the Python proxy. Generated lifecycle SQL preparation, coordinated
production publication, CLI plan/apply and an IaC provider remain future work.
The Python view-specific branches remain a labelled compatibility lab.

No scheduler, durable alarm state, multiwindow SLO evaluation, notification delivery, or historical correction store exists. The UI is an initial query interface, not a complete investigation workbench: an initial bounded waterfall and raw-duration heatmap are implemented (see VISUALIZATIONS.md), but general saved workspaces, unbounded/complete trace capture, general selection algebra, and statistical cohort comparison remain absent.

The new raw-span heatmap aggregates the full matching span population; its cell
inspection remains bounded at 200 spans. Trace completeness is unknown. Guided
USE/RED entry points now include the custom queue/worker adapter and identify
remaining CPU/memory/pool and error-rate gaps; they are not complete resource/service scorecards.
See VISUALIZATIONS.md and INVESTIGATION_WORKFLOWS.md.

The Go semantic-package registry currently performs metadata-shape recognition,
not automatic live scanning or general expansion/compiler lowering for all proposed measures.
A bounded HTTP server-span binding/expansion and normalized request preparer now
exist. A bounded ClickHouse exporter adapter and shared aggregate UI now work
for an explicit HTTP instrumentation binding (LIVE_HTTP_PACKAGE.md). General
auto-discovery, immutable custody and prepared-table materialization remain absent. See SEMANTIC_PACKAGES.md. A supplied matching
inventory cannot establish data completeness, freshness, health or a valid SLO.

## Application and environment

The consumer does not reclaim pending Redis entries after process failure. There is no durable outbox, transactional terminal telemetry, or acceptance reconciliation. A local reset can clear the demo but does not repair these semantics. Trace/span drilldown currently filters by reservation ID alone; tenant is fixed to demo in the sample. Generalized relations must carry the full scoped identity.

Published ports bind to loopback. Query credentials are disposable local credentials, not scoped production identities. The server uses a read-only query setting, but the supplied database account is also the ingestion account and is not a production authorization model.

The original hosted environment denied Docker startup and browser installation.
The current cloud baseline runs the default stack and real Chromium journeys;
see VALIDATION.md and docs/journeys/README.md. Optional HyperDX remains unverified.
A nominally “good” sample request can miss its deadline behind the consumer's
batch barrier under background load. Run `npm run test:smoke` for the controlled
pause/drain/restore fixture; direct `scripts/smoke.py` assumes an otherwise quiet
queue and may fail correctly under load. The initial background-traffic journey missed its nominal deadline while queued;
documentation runs now pause/drain/restore the owned generator without deleting
data. Stale pending stream entries from crashes are still not reclaimed. A
recreated cloud container also inherited proxy defaults that routed local service
traffic incorrectly; Compose now bypasses the proxy for its own service names.

## Annotation tooling boundaries

The first operation/event dialect, source-contract manifest and test-only weaving
wrapper are implemented (GOO_ANNOTATIONS.md). Full build/dependency manifests,
automatic live inventory scanning, generic operation aggregate compilation,
production weaving rollout and instrumentation coverage are not implemented.
The custom runtime lives in a separate OTel module; the existing deployed Go
packages service remains a plain build; the separate operation sample is woven. Event delivery/dropped-event diagnostics do not
establish durable business completion or complete absence evidence.

Normalized completed operation spans can now be checked against source manifests
and summarized per operation/role (OPERATION_CONFORMANCE.md). A separate woven
sample now integrates collector/ClickHouse and the UI (LIVE_OPERATIONS.md);
this does not verify completeness. Native operation contracts now validate reset/temporality and histogram compatibility; see NATIVE_COUNTERS.md and NATIVE_HISTOGRAMS.md.

## Operation investigation scope

Operation heatmaps and trace drilldown now share a captured, validated source
response (OPERATION_DRILLDOWN.md). Trace inspection contains only operation spans
in the selected resource/library/start-time window; it is not a complete
distributed trace fetch. Empty cells do not prove health, and late-data reruns
can change results. Local immutable operation captures now exist (SAVED_INVESTIGATIONS.md); bounded cross-service expansion now exists through the separate general exploration surface (EXPLORATION_DESIGN.md).
General business-subject evaluation remains absent. Native histogram intervals are now supported separately.

Bounded scoped subject navigation now follows declared events across traces in
the captured operation evidence (SUBJECT_NAVIGATION.md). It joins event-bearing
invocations only. General relation-package resolution, immutable event IDs,
business deduplication, retry classification and durable completion remain absent.

## YAML operation package boundaries

Local YAML operation packages now replace the sample-specific Go binding and
select built-in views for checkout and the separate report app (OPERATION_PACKAGES.md).
They require the selected Go source package, load at startup, and use a fixed
ClickHouse adapter/view vocabulary. General registry/version resolution, source-free
contracts, CEL/KCL, deployment/apply and operation Databricks lowering remain absent.

Native operation interval rates now support explicit cumulative/delta temporality,
unknown reset/gap/boundary intervals, late-input revisions and per-stream scope
(NATIVE_COUNTERS.md). The live sample uses cumulative metrics; delta is fixture
tested. Automatic fleet aggregation and metric exemplars remain unsupported. Operation error ratios now use exhaustive aligned same-basis outcome counts (OPERATION_ERROR_RATIOS.md); missing members remain unknown and zero traffic is no data. Explicit aligned histogram merging is available through the Go core/CLI; it does not discover fleet membership. Counter-to-span links are time-based context only.

Fixed declared fleet membership and exact compatible interval totals are now supported (FLEET_COVERAGE.md). Source observation is not completeness. The live lab now deploys both expected instances. Bounded-skew comparison is supported without shifting timestamps or asserting exact totals; automatic discovery, membership activation windows and common-window aggregation remain absent. See FLEET_COMPARISONS.md.

## Saved investigation limits

Local server-captured immutable revisions and fixed-question refresh now exist. A general capture catalog now exists separately. No multi-user authorization, operation-capture catalog, retention/quota, full query-session persistence or raw-metric-input replay exists. See SAVED_INVESTIGATIONS.md and PRODUCT_GAPS.md.

## General exploration boundaries

A bounded Go span/log AST, captured-record counts/groups, discovered keys and paged cross-service trace/exact-log navigation now exist. Reads are not transactional, ordinary live offset pages are not frozen, trace completeness is unknown, numeric-looking attributes remain strings, and general explorations/expanded context are not archived by operation captures. One-field distributions and fixed-window descriptive cohort comparison now exist. General chart algebra, arbitrary-field numeric casts, complete trace accumulation, live tail and production tenancy remain absent. See EXPLORATION_DESIGN.md.

- Explore baselines are volatile browser-held captures, not durable saved operation
  receipts. Comparison checks accounting/scope, not caller-supplied provenance;
  it provides descriptive differences without statistical/causal claims. See
  [workspace contract](PRODUCT_WORKSPACE.md).

- Manifest-free telemetry packages now persist reusable selectors and captured-record
  counts, but do not yet share all typed measure contracts with the aggregate
  compiler. Source promotion is an optional Go template, not automatic function
  inference or patching. General metrics, calculations and warehouse lowering of
  this package kind remain future; see [contract](TELEMETRY_PACKAGES.md).

## Exploratory UI testing boundaries

The opt-in Quint/Bombadil campaigns exercise successful reports and Python retry,
rejection and stalled-work paths with independent source checks. Generated schedule
replay checks observed handler effects; it does not establish deterministic browser
scheduling, business durability or telemetry completeness. Resource/export-loss/
late-capture generation and long-lived independent actors remain unimplemented.
See [testing contract](BOMBADIL_TESTING.md); existing curated journeys cover their
separately documented retry, export-loss and late-evidence scenarios.

Log discovery now includes literal message/service/severity controls, readable
messages and independently scoped surrounding logs. Reported trace-ID groups and
fixed full-population span-duration buckets are implemented. These do not supply
root/trace latency, unique invocation counts, stable live cursors, live tail or complete
capture. Quick search adds visible AND filters rather than silently replacing
existing predicates. Context highlights compare captured fields; identical duplicate
logs have no unique event identity. See the discovery section of EXPLORATION_DESIGN.md.

Observed service/span-name inventory now exists, bounded to 1000 pairs and one
environment/window. It excludes log-only services and does not infer fleet
membership, health, dependencies, request rate or declared business operations.
General inventory/Explore evidence can now be saved as immutable bounded receipts;
expanded live context and unsaved browser baselines remain volatile. Explicit comparison-pair captures now exist (ERROR_COMPARISONS.md). See [contract](SERVICE_INVENTORY.md).

General captures now retain full supported aggregates and up to 1,000 inspection
records with stable stored pages. Ordinary live offsets still change with arrival
and retention. Catalog listing is bounded and uses ID seek order, not a frozen
catalog snapshot; quotas/indexing and team authorization remain future. Refresh
requires the same fixed question and compiled plan; old receipts remain reopenable
when that plan is unavailable. See [GENERAL_CAPTURES.md](GENERAL_CAPTURES.md).

## Core workbench boundaries

[Core basics](CORE_WORKBENCH.md) now provide full-population span-duration summaries,
fixed record-count time charts, reusable package questions, span/log service event
activity and reported parent links. Arbitrary numeric expressions, ingestion lag/drop evidence and production
authorization remain absent. Typed native metric packages now exist; see
[measurement boundaries](MEASUREMENT_PACKAGES.md).
Per-process query limits/cancellation cover general routes, not all legacy routes
or cross-replica/per-tenant quotas. Service discovery is a separate live read and
is not archived by saved inventory. Event age does not establish pipeline health.

[Grouped span-duration measures](GROUPED_MEASURES.md) now support complete per-group summaries/ranking. Arbitrary numeric measures remain absent. Opt-in bounded grouped time series now
exist through the [typed record-measure adapter](RECORD_MEASURES.md). Old summary+group capture plans may require explicit new live evidence rather than refresh; count-only groups are not retroactively upgraded.

[Error/latency comparison boundaries](ERROR_COMPARISONS.md): span status is exporter status, not request/business outcome policy; captured-error percentages include Unset/unknown in the denominator. Different fixed windows and saved server-read pairs now exist. Statistical attribution and custom outcome policies remain future work. Local [deployment context and ordered notes](WORKBENCH_CONTEXT.md) now exist, with explicit authored provenance, immutable branches and exact evidence references.

Investigation documents currently reference general captures and local deployment receipts; legacy operation receipts and live-question cards are not supported. Marker corrections/retractions, automatic CI ingestion, verified authorship and team authorization remain future work. Drafts persist only in the current page until saved. See [contract](WORKBENCH_CONTEXT.md).

## Scoped recipes and examples

Three [built-in recipes](INVESTIGATION_RECIPES.md) preserve an applied population,
with explicit source-free preview and live execution. Six example packages use
the shared product flow. Prerequisites are declared, not automatically verified;
repository-authored ordered workflows and live typed lifecycle packages now exist.
UI workflow authoring and Databricks lifecycle lowering remain absent.
The specialized reservation/Redis lab remains an explicit legacy section.


## Shared measurement boundaries

See [MEASUREMENT_PACKAGES.md](MEASUREMENT_PACKAGES.md) for source limits, required
stream metadata, local custody semantics, display limits and example log transport.
No metric family implies telemetry completeness. Lifecycle absence is unknown;
first Goo observation is not warehouse ingestion. The original compatibility lab
has different preview assumptions and is not retroactively rewritten. Package
revision catalogs have a 256-file bound; persistent custody storage currently has
no quota/garbage collection. Source data and custody reads are not transactional.


Metric discovery is now implemented for a named service/environment over a
bounded two-minute observation window. It is not fleet membership discovery or a
complete instrument inventory. Dimension-value cardinality, exemplar navigation,
exponential histograms, omitted metric units and missing resource-instance IDs
remain unsupported by the package workflow. Empty optional scope version/schema
are accepted as exact identities. See MEASUREMENT_PACKAGES.md.

## Product coherence boundaries

[PRODUCT_ESSENCE.md](PRODUCT_ESSENCE.md) is the organizing direction. Definitions
now exposes searchable record and measurement package revisions, but the two
schemas retain separate editors/catalog sections. Discovery defaults to an explicit
two-minute service window; contextual evidence follow-ups can pin original bounds; measurement lookback/fixed bounds are separate. Full
workspace-session persistence, a unified cross-family authoring schema and
production multi-user collaboration remain future work.

## Investigation workspace boundaries

Question-first briefs and direct capture destinations now exist. A selected saved
revision receives an immutable child; current drafts require an explicit save.
There is still no mutable latest head, automatic branch merge, multi-user presence
or production authorization. Capture and document publication are separate, so
attachment failure retains a saved capture for retry. Drafts are page-local until
saved. Capture-card scope reads validate the stored capture and return only receipt
metadata; local disk integrity checks still read its bounded stored evidence.


## Contextual question boundaries

Evidence-card follow-ups now preserve or explicitly replace scope. Trace/log
relationships use exact reported IDs; resource discovery is service/time context.
The observation chooser covers the first 100 retained records. Inventory and
comparison captures require opening an underlying supported capture. Automatic
causal ranking, cross-signal attribute translation and completeness inference are
not implemented. See INVESTIGATION_FOLLOWUPS.md.


## Shared scope planning boundaries

The usability [entry-point review](ENTRY_POINTS_DESIGN_SPIKE.md) found that
Services → Discover metrics previously dropped the inventory window. The first
[question-workbench slice](QUESTION_WORKBENCH.md) now passes those bounds and
consolidates Explore execution with explicit draft/applied scope. Broader
Investigation/Evidence and Library entry-point restructuring, standalone native
browse-first discovery, navigation history and durable draft recovery remain
proposed. [Service browsing](SERVICE_BROWSING.md) now provides scoped instrument
previews from service detail.

Recipes and follow-ups now share checked scope/transform metadata; persisted
question and capture formats remain unchanged. OTel/custom selectors retain their
namespaces and native bindings retain declared schema identity. Automatic convention
profile resolution, record schema-URL projection, cross-signal attribute mapping,
arbitrary typed numeric expressions and backend capability negotiation remain
future work. Existing record count/duration/status controls now lower through a
bounded typed measure plan; opt-in split charts preserve their semantics.
Comparison scopes retain separate members; automatic comparison transforms are
rejected. See [QUESTION_SCOPE.md](QUESTION_SCOPE.md).

## Service browsing boundaries

The [service detail flow](SERVICE_BROWSING.md) offers observed HTTP/database/messaging
span questions and native instrument previews. The [unified catalog](UNIFIED_SERVICE_CATALOG.md)
starts from all supported signals; [all-signal discovery](METRIC_SERVICE_DISCOVERY.md) now exposes log-only and
metric-only services from gauge/sum/explicit-histogram records. Metric summary and
exponential-histogram families remain outside the adapter. Attribute
discovery is bounded, convention aliases are not merged, and missing observations
do not establish instrumentation coverage or health. Service context reads can
fail independently. Service return is local, not durable browser-history recovery.

## Unified catalog boundaries

[Services](UNIFIED_SERVICE_CATALOG.md) now has one discovery entry and one row per
service. Operation inventory still reads the whole environment/window with the
existing 1000-pair limit, then filters displayed operations by service. Its capture
action explicitly captures that whole environment scope. Service-only operation
inventory compilation, restoring catalog state from a link and catalog captures
remain future work.

Explore record questions now have [fixed-scope share links](QUESTION_LINKS.md) and
source-free browser-history restoration. Unsaved keystrokes, unpreviewed native drafts, service-detail panels and
trace-inspector state are not serialized. Validated native questions now have
typed investigation steps, described below.

[Typed investigation steps](INVESTIGATION_STEPS.md) now persist record questions
and validated measurement definitions with fixed scope and view. Native questions
can be shared and reopened without telemetry reads. Service-detail panels,
inspector state, baselines and unsaved keystrokes remain local. There is no
automatic association with a latest capture or synchronized multiuser editing.

[Linked question evidence](LINKED_QUESTION_EVIDENCE.md) now supports explicit
capture-to-step histories. Capture and document publication are separate; retained
capture IDs support retrying a failed attachment without rereading. Histories are
bounded to 16 references per question, with no scheduled refresh or distributed
transaction. Matching scope is not a shared snapshot or population guarantee.

## Stored capture comparison boundaries

[Capture comparison](CAPTURE_COMPARISON.md) supports source-record counts, stored
span summaries, exact-stream gauge endpoints and [complete native counter and
histogram windows](NATIVE_WINDOW_COMPARISON.md). Full coverage, identity, unit,
temporality and bucket compatibility are required; one unsupported member
withholds all changes. Lifecycle cohort deltas, fleet merging, interval
realignment and histogram rebinning remain unsupported. Different definitions, units or environments withhold all numeric
deltas. Exploratory selection is transient until explicitly saved as a
[comparison card](SAVED_COMPARISONS.md). Frozen receipts and document publication
are separate; retained receipt IDs support explicit recovery. Complete comparison
snapshots are limited to 128 KiB and fail rather than truncate. Sharing requires
access to the same workspace; verified authorship, access control and receipt
garbage collection remain future work.

Native questions and shared steps preserve exact nanosecond scope. The comparison
panel now offers [observed-support discovery](OBSERVED_WINDOW_DISCOVERY.md) for
counter and explicit histogram captures, with explicit selection and rereads.
Asynchronous streams without common boundaries, unsupported members and overflowing
histogram populations withhold support. Discovery does not select stream subsets
or realign intervals; the ordinary time editor still authors millisecond bounds.

Typed comparison sides can now become
[reusable question cards](REUSABLE_COMPARISON_QUESTIONS.md) with their own capture
linked. Staging is local until revision publication; fresh-window previews retain
the definition but do not rewrite the saved fixed scope or relink old evidence.
Inventory/legacy pair captures are not reusable question sides.

[Fresh follow-up comparison](FRESH_COMPARISON.md) uses the saved question's duration
and definition. Native scope transport preserves nanoseconds but does not infer
SDK export phase; incomplete fresh windows still withhold numeric differences.
Lifecycle question cards do not offer this comparison action.

## Captured breakdowns and automatic dashboard scope

[Stored one-field group comparisons](COMPARISON_BREAKDOWNS.md) now reuse complete
aggregates. Numeric/two-field comparisons and arbitrary regrouping of old evidence
remain unsupported; new dimensions require explicit source recapture. Automatic
[service overviews and a distributed relationship table](SERVICE_DASHBOARDS.md)
now execute bounded typed widgets within Services. Instance navigation,
graphical topology, async/declaration edges and complete
dashboard snapshots remain future work.

Comparison breakdowns now offer local count/share ranking and recorded
error/mean/p95/p99 ranking for spans, preserving full counts and missing-last
coverage. Ranking is descriptive; support thresholds and statistical significance
are not implemented. Stored receipt ranking is local, while saved population
views retain their selected comparison metric. Log breakdowns now omit unsupported span-measure columns, and live log
comparison controls use population share without a span-measure selector. Span
measures absent from captured data remain explicitly unavailable.

Namespace-aware service identity and registered operation-span dashboard cards
are implemented; see [service identity](SERVICE_IDENTITY.md). Trace schema URL
remains declaration-only because the current exporter does not store it. These
cards do not replace explicit operation conformance or business-outcome evaluation.

Declared services without observations now remain visible in Services. Their
metadata comes from the current registry, not a frozen deployment inventory.
Discovery failure withholds presence claims; registry failure is shown separately.

[Reported-edge questions](REPORTED_EDGES.md) now select exact resolved child records
within the captured discovery scope. Async links, external peer inference and
end-to-end request latency remain outside that relationship. The first
[visual comparison pass](VISUAL_COMPARISONS.md) uses aligned scales and compact
metrics; some signed/scientific native formats still remain exact tables.

The [display component library](UI_COMPONENTS.md) is incremental. Compact latency,
stat lists and grouped metric comparisons are adopted. Explore and Services
share passive tables and explicit result states.
Shared facets and typed attribute values are now adopted in Explore and
trace/log inspection. Other table migrations remain incremental.
Native histogram P95 remains a bucket range; this is not a new scalar percentile
or a P90 evaluator. Light-DOM app styles must stay scoped to their owning view.

A bounded log SSE tail is now implemented; see [limits](LIVE_LOG_TAIL.md).
Durable replay/ingestion cursors, trace tails and live gauges/histograms
remain planned. Summary property updates currently rerender inner
DOM and reset disclosures; log-tail row patches, bounded retention and explicit gaps now live in the app
adapter. Snapshot value components still reset disclosure state on data updates.

## Rolling comparison scope

Live baseline comparison displays record counts for saved span/log questions
with windows of at most two hours, plus duration and recorded status when the
saved span question requests them. Native counters also support exact-stream live rate comparison; gauges and
histograms retain fixed comparison flow. Count-only captures are not silently enriched or reread as a new definition. Refresh uses bounded polling;
only log tail currently uses SSE. Live snapshots expire after two minutes and
are process-local until explicitly frozen. Production authentication/session
isolation and durable live workspaces remain future work. See LIVE_COMPARISON.md.

- Live native counter comparison requires a single complete baseline stream. Focus this stream now creates an explicit exact source binding. Unselected
  definitions with additional stream membership still withhold comparison. Endpoint discovery and exact-window reads are bounded, separate
  source reads. maxGapMs also bounds live endpoint age. Frozen fixed-window
  evidence retains historical coverage, not a continuing freshness claim.
  Live gauge/histogram and fleet comparisons remain unimplemented.

- Exact stream selection pins all captured resource and point dimensions, including
  service.instance.id. A restart, added resource attribute or new schema produces
  a different stream; Goo does not auto-follow it. Focus is bounded to observed
  native inventories, not a fleet selector. See EXACT_STREAM_SELECTION.md.

- Native stream search/review filters are local over bounded retained evidence;
  they do not discover additional source streams, revise questions, aggregate
  fleet rates or imply health. Local filters are not saved in question links or evidence. Outer stream
  disclosure state survives local filtering within one rendered result; nested value components retain their existing lifecycle limits.
  Unrecognized diagnostics preserve the exact code with neutral guidance.

- Compact native service values are latest evaluated endpoint/interval values,
  not whole-window rates or fleet percentiles. A valid interval may be displayed
  alongside incomplete whole-window coverage. Counter interval rates retain the
  kernel's floating representation; rational full-window comparisons are separate.
  Automatic service native selection remains capped at two supported instruments;
  supported choices can be changed with the overview picker. Selection is local
  to the service visit and resets on reopening. Apply explicitly rereads all
  supported overview questions from retained metadata; it does not rediscover
  instruments or create a transactional snapshot. Unsupported instruments remain
  visible in Native instruments. Saved dashboard configurations remain planned.

- [Native timelines](NATIVE_TIMELINES.md) are fixed retained views capped at the
  latest 200 observations/intervals per stream. Scalar domains are labelled and
  local to each visual snapshot, not aligned fleet or cross-stream comparisons.
  Histogram timelines show interval counts; P95 ranges/native buckets live in
  selection details. Historical gauge samples may be stale under the fixed
  evaluation policy. Live gauges/histograms and timeline-driven window reruns remain planned.
  Fixed questions may opt into [Go-evaluated time buckets](HISTOGRAM_TIME_BUCKETS.md).

## Retained histogram heatmap limits

Native bucket grids display the latest 32 retained intervals and at most 64 buckets per layout. These are interval columns, not equally spaced elapsed-time buckets. Oversized layouts retain their exact tables; unknown intervals have no inferred layout. Shared raw-count scales do not normalize differing window durations.

## Explicit native histogram time-bucket limits

Opt-in `histogramBucketMs` derives at most 200 anchored buckets per stream and 20,000 total output buckets. Exports crossing bucket boundaries are withheld, not split; narrow widths may therefore make many buckets unavailable. Final buckets can be shorter. Display grids retain the latest-32/64-native-bucket limits; no resampling, fleet merge, live histogram tail or normalized-rate heatmap is implied. Export-window coverage and time-bucket coverage are separate.

## Evidence-summary adoption limits

Shared evidence summaries currently cover captured measurements and their native investigation entries/cards. Fixed record/native service widgets also use compact shared summaries. Stored comparison sides now use the shared summary too. Other record and inventory presenters retain their current layouts. Receipt-only metadata lacks a full evidence basis and diagnostics; the summary leaves those unavailable and the existing Open captured evidence action exposes the full result. Display input bounds and timestamp formats fail closed; arbitrary adapter-owned interactive children are unsupported. UTC/tooltips cover summary bounds and native stream/selection timestamps; stored comparison endpoint/window clocks and fresh native previews also use UTC/tooltips. Legacy diagnostic JSON and older chart presenters may still expose raw nanoseconds.

The [shared inline inspector](SELECTED_ITEM_INSPECTORS.md) retains the 32 most recent selected detail bodies per rendered result. It does not retain selection across new source results or provide cross-query navigation history. Original inline row attribute disclosures and outer trace/surrounding-log source workspaces remain separate; only their selected-item presentation is shared.

## Core experience refinement boundaries

The [first survey slices](CORE_EXPERIENCE_REFINEMENTS.md) retain eight Explore
visits in one tab; reload/eviction discards them and browser URL history still
stages shared questions separately. Leaving stops tail subscriptions and invalidates
pending publications. Service diagnosis counts matching mixed-kind captured spans,
not verified requests or service health. Duration selection is for record
populations; native histogram cells have no inferred trace membership. Trace
breakdown sums are inclusive and can overlap. Deployment context reads remain
explicit and independent of telemetry and captured custody. Convention capability
explanations offer existing binding authoring; no general schema resolver is added.


B2's first [record field/display slice](FIELD_PRESENTATION.md) is delivered. Shared
fixtures cover record predicates and omitted empty defaults; they do not yet cover
every service/native/envelope protocol. Source owners retain their independent
cancellation and snapshot rules; no generic request-owner abstraction was added.


B2 now also shares service namespace label formatting between catalogs/dashboard
scopes and annotation summaries. Named labels are quoted to distinguish literal
sentinel text from missing/empty modes. Shared namespace fixtures cover logical
membership and record predicates, not all native/envelope protocols or server
artifact digests; those retain their existing contracts.


B2 now includes [shared-question/native wire fixtures](QUESTION_PROTOCOL.md) and
comparison-only Go-default equivalence for canonical native capture reuse. Go still
owns full definition validation; the fixture set is bounded. The bounded B2 acceptance
audit is complete; store-specific publication consolidation is the next cut.

Observed record field columns are bounded to eight qualified fields from supplied
rows; this is not source-wide schema discovery. Core and added columns reorder
within their own groups. Population layouts can be saved; live preferences remain
page-local, empty fixed pages omit the picker, and arbitrary transforms are still
future work. See [record field contract](RECORD_COLUMNS.md).

Population scope widening currently expands an enclosing aligned UTC window, within the 120-minute query limit. Removing arbitrary predicates and exact selection-versus-rest remain planned; see [time expansion](POPULATION_WORKSPACE.md#expanding-the-time-window).

[Categorical selection-versus-rest](POPULATION_PARTITIONS.md) is now implemented inside a fixed filtered parent. General numeric/range complements remain planned. Partition queries are logically disjoint over a fixed record state; independent reads are not a transactional snapshot.

Population broadening now supports removing whole direct conjuncts in a staged
child. Nested OR/AND groups stay atomic; arbitrary nested branch editing and
noncategorical complements remain outside this cut. See
[the bounded contract](POPULATION_BROADENING.md).

The bounded [comparison context pass](COMPARISON_CONTEXT.md) shares UTC windows,
counts, modes and provenance across fixed/stored/live results and omits duplicate
selectors on automatic fixed-pair results. Bounded shared actions now cover fresh
previews and stored/saved inspection, receipt retry and finding publication.
Count-series adoption is recorded below; the [D1 audit](D1_ACCEPTANCE.md) now
accepts this bounded milestone. Stored first-page inspection, absent series in old
captures and separate native/record semantics remain explicit boundaries. E1 native
instrument discovery is next; no unit conversion or fleet percentile is implied.

[Comparison action/inspection continuation](COMPARISON_ACTIONS.md) now retains
population comparison parents and provides in-place stored-side inspection. The
legacy single-question row action still refines/rereads; full-capture navigation
from receipts remains separate from retained inline Back. Stored inline inspection
is its first retained page, not a full-population membership browser.


[Retained record activity comparisons](COMPARISON_SERIES.md) now share count geometry across fixed and explicitly loaded stored/frozen capture owners. Rolling record responses now optionally include retained series for bucketed questions, with pinned UTC selection and exact Freeze. Reconstruction of old cohort-only comparison captures remains unimplemented. Native metric series keep their existing semantics and presenters.

The first [E1 native discovery catalog](NATIVE_INSTRUMENT_DISCOVERY.md) provides
HTTP duration metadata hints and generic package fallback. It is not complete
convention capability planning: observed dimension keys do not establish per-stream
coverage, error denominators or verified conformance. Non-monotonic sums and
exponential histograms remain unsupported by the existing kernels/inventory.

[E1 typed planning](NATIVE_INSTRUMENT_DISCOVERY.md#typed-server-plan-and-observed-stream-presence)
now replaces browser rules and reports HTTP key presence per exact observed stream,
including missing/empty identities and display truncation with full totals. This
is not verified interval coverage or an error denominator; generic profiles do not
inherit HTTP semantics. Plans remain `not-run` even when matching metadata exists.

The messaging processing candidate uses experimental OTel v1.30.0 semantics.
Its presence matrix does not validate operation values, destination cardinality,
unique deliveries, queue lag, delivery success or end-to-end trace relations.
Other messaging instruments retain generic bindings; this is not a full broker
or consumer dashboard. Non-monotonic pool sums still require kernel support.


The [bounded E1 audit](E1_ACCEPTANCE.md) accepts discovery/metadata planning, not
convention verification. Native Run currently requires nonempty instance IDs and
fails the whole matching input on invalid identity. The ClickStack counter adapter
refuses fractional or >=2^53 values, even if metadata permits preview. Summary and
exponential histograms are not inventoried. Discovery excludes exports exactly at
its end; interval evaluation includes that endpoint and bounded preceding context.
Discovery and Run row counts can differ without changing the fixed question.


[Time/latency region partitions](REGION_PARTITIONS.md) are limited to required
native span UInt64 durations and event timestamps. Both sides use independent
source reads; no atomic membership snapshot is claimed. Numeric attributes,
missing/invalid attribute durations and arbitrary Boolean negation remain unsupported
by this operation. A selection exceeding the existing predicate budget is refused.


The [bounded D2 audit](D2_ACCEPTANCE.md) accepts current population operations.
Nested Boolean branch editing, arbitrary numeric-attribute complements/transforms,
source-wide field discovery remain outside that cut. Reusable Activity layouts
are now separately delivered in [Activity views](ACTIVITY_VIEWS.md). Exact Activity-row retention/recovery still requires server membership,
occurrence identity, expiry and idempotency; connection-local handles are insufficient.


[Activity membership prerequisites](TAIL_SELECTION_CUSTODY.md) include a tested Go
bank shared by the tail SSE issuer and a separate references-only publication route. Process-local proofs expire from original read time
and disappear on restart. The typed backend publisher now supports durable
committed-request recovery without those proofs. Origin-aware backend catalog/read
views, explicit public read/catalog routes, compact retained inspectors and pinned
browser citations are implemented. Activity selection/publication/retry now uses these capabilities. The bank remains
process-local and expiry/capacity can make new preservation unavailable; normal
observation continues. Old unprojected tail requests require bounded receipt
scans (1,000 / 16 MiB); exceeding those bounds refuses rather than omitting roots.
Existing v1 captured/result-page manifests cannot represent mixed-poll tail
provenance and remain unchanged.

The [bounded Activity retention acceptance](ACTIVITY_RETENTION_ACCEPTANCE.md) now
covers client expiry, configuration failures, actual database transport outages,
capacity exhaustion and proof loss on process restart with keyboard/mobile retry.
Browser-clock advancement does not advance server proof expiry; backend/Hegel
checks cover that separate clock. Conflict/refusal messages have unit coverage;
no browser concurrency campaign or distributed bank guarantee is claimed.

[Activity views](ACTIVITY_VIEWS.md) persist rolling selectors and column layouts,
not browser observation buffers, capabilities or source clocks. Start remains
explicit after reopen. Service relevance is an exact environment filter, not
Boolean selector satisfiability or evidence of matching telemetry. Catalog search
is limited to the loaded 20-item page. Mutable aliases, team access control and
automatic startup preferences remain unsupported.

Reusable [Activity views](ACTIVITY_VIEWS.md) use a typed rolling configuration in
workbench's existing immutable file publisher, separate from fixed exploration
views. Shared revision-library display appears in Activity’s Views menu, Saved and Services. The stream
adapter restores configuration through `GooTail.stage`; Start alone opens SSE.
No observations or capabilities cross the configuration boundary. The linked
flow diagram documents publication, staging and later separate retention.

Activity-view browsing adopts `GooDefinitionLibrary` with an explicit `activity`
family and exact revisions, keeping same-title scopes distinct. The existing
record-column picker projects both streamed and staged layouts. Configuration
reads, clipboard sharing and library paging stay in `GooActivityViews`; scope and
draft lifetime stay in `GooStreamEntry`. `GooTail.stage` prepares local display
without a subscription. No display component fetches or starts telemetry.

`GooFilterEditor.mount(host, {where, signal, onChange})` is the shared supplied
predicate editor for Activity and fixed population adapters. `set` restores a
supplied tree; `snapshot` returns detached local input. It emits normalized
predicates or a validation error and never reads or executes. Styles live in
`components.css`; adapters retain execution and custody. See [contract and
journey](FILTER_EDITOR.md).

Single-question filtering now adopts the same helper. `setFields(fields)` supplies
local attribute/resource-key suggestions without changing input, emitting edits
or reading data. The old flat field picker/composer and their page styles are
removed; service/message/duration shortcuts are application-owned AST inputs.

`GooFilterEditor.setSuggestions(data)` accepts detached, scope-labeled observed fields/values supplied by application adapters. `GooAttributes.suggestionData(records,fields)` bounds the physical-record projection, preserving namespaces and missing/empty distinctions. Neither reads telemetry. Suggestion search and selection use the same passive editor across all record-query surfaces.

The shared filter editor accepts `setValueLoader` and supplied `setCountedValues` states. Count lists reuse passive `goo-facet-list` and `goo-result-state`. The application-owned `GooFilterValues` module validates fixed breakdown scope, exact counts, cancellation and obsolete responses; presentation and local search never fetch. See [contract](FILTER_VALUES.md).

`goo-time-bars[range-select]` adds optional supplied-key range gestures for interactive button bars. `goo-time-range` emits `{fromKey,toKey}` in supplied order. Drag uses pointer capture, keyboard uses Space/Shift-arrows/Enter, Escape cancels an unfinished range, and touch drag scrolls normally. Removing the attribute or replacing data clears gesture state; components never fetch. The app-owned [latency histogram adapter](LATENCY_SELECTION.md) maps those keys to typed duration bounds and existing population actions. `GooPopulationSelection.mount` accepts `compact:true` to disclose exact bounds while keeping preview/actions visible.

`goo-time-bars.clearRange()` withdraws painted/unfinished range state without emitting a selection or reading a source; the histogram owner uses it when exact bounds are edited.

The latency histogram adapter accepts app-supplied `onBegin` and `onSelect(bounds,{count,total})` callbacks, plus a compact hint mode. Activity uses these instead of the fixed selection form. Activity's begin callback cancels its pending reads before staging a typed range; source cadence, cancellation, filters and comparison remain in `activity-latency.js`/`stream-entry.js`. See [Activity latency](ACTIVITY_LATENCY.md).

Activity baseline comparison reuses `GooComparisonVisual.histogram` and supplied `goo-comparison-context` with a fixed 0–100% shape scale. App-owned `activity-baseline.js` loads stored metadata, validates native comparison accounting and publishes independent side clocks. Count/context disclosures do not fetch. [Contract](ACTIVITY_BASELINE.md).

`GooComparisonVisual.histogram` now accepts `layout:'rows'` and supplied side labels. It renders a semantic table with aligned bucket labels and paired percentage bars on the same 0–100% scale, using the existing exact-string bar primitive. The original SVG layout stays the default. Journey 157 is a labeled passive fixture with no application/source requests and mobile label/overflow checks.

Saved Activity baseline selection remains app-owned configuration: `GooActivityBaseline.restore` stages an exact reference without fetching, while `reference` exposes the selected ID separately from loaded evidence. `GooActivityLatency.restoreBaseline` resets the panel/cadence without subscribing. Existing supplied comparison displays are unchanged. [Contract](SAVED_ACTIVITY_BASELINE.md).

Activity’s embedded library receives applied environment and visibility from the stream owner. `setScope` cancels stale reads and clears supplied rows plus inspector cache when the environment changes. `GooDefinitionLibrary.clear()` resets supplied presentation without reading. The catalog’s server page is capped at 20, so its redundant internal 25-row pager is hidden; explicit server first/next controls remain.

Activity’s Investigate control uses native disclosure and ordinary action buttons. The stream adapter owns its transition to the existing fixed editor, parent retention and focus restoration; no reusable display element gains navigation or read behavior. Record inspectors and folded definition-library inspectors remain separate regions with separate selection lifetimes. [Navigation contract](ACTIVITY_NAVIGATION.md).

`goo-table[record-layout="compact"]` opts supplied record rows into the shared narrow-screen layout in `record-presentation.css`. `GooRecordPresentation.table({compact:true})` emits it; server-rendered cells use the same semantic `data-record-core`/`data-record-field` markers. Core fields group by role on mobile, chosen fields keep qualified labels, and action buttons remain caller-owned. Legacy application rules explicitly exclude this layout; no component read/subscription is introduced. [Composition and storyboard](COMPACT_MOBILE_RECORDS.md).

`GooRecordPresentation.table({mobileCompact:true})` separates responsive record layout from `compact` preview content. Responsive time markup retains full accessible date text and an aria-hidden short mobile rendition in one exact `time` element. Generic spanning detail cells stay full-width; fixed messages keep disclosure/expansion. The shared result owner adopts this for fixed, saved and explicit context tables while retaining source/provenance and action ownership. Legacy telemetry-card rules exclude the opt-in. [Adoption contract](COMPACT_MOBILE_RECORDS.md#fixedsaved-adoption-storyboard-162).

Service-held span observations also call `GooRecordPresentation.table({mobileCompact:true})`, with short visible Inspect text and the existing accessible name. Service detail retains its cached inspector, original supporting question and source lifetime; the shared primitive owns only responsive presentation. [Service adoption](COMPACT_MOBILE_RECORDS.md#service-observation-adoption-46).

`GooRecordPresentation.attributes(host, record, options)` centralizes exact root event-time rendering while preserving caller options, scoped fields and custom fallback rendering. Record details, retained cards and waterfall inspectors use it. Retained cards keep `raw:false`; original manifest disclosure stays separate. [Renderer inventory and action meanings](RECORD_RENDERER_AUDIT_2026_10_24.md).
