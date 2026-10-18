# Known issues and gaps

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

grammar.py still uses view-specific query branches. Grouping fields and map bindings are hardcoded. definitions/reservations.json is a small demo config and has no complete resource schema or version resolver. The original EDN package and deadline-outcome syntax are illustrative. Separate [Go aggregate](PORTABLE_COMPILER.md) and [lifecycle](LIFECYCLE_PREPARATION.md) pilots now implement a bounded typed model, normalized-event preparation, reference evaluation, and aggregate backend compilation. They are not integrated into the Python API and do not implement real-exporter source adapters, generated SQL preparation, template expansion, durable immutable publication, CLI plan/apply, or an IaC provider.

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
The initial background-traffic journey missed its nominal deadline while queued;
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
expanded live context and browser baseline pairs remain volatile. See [contract](SERVICE_INVENTORY.md).

General captures now retain full supported aggregates and up to 1,000 inspection
records with stable stored pages. Ordinary live offsets still change with arrival
and retention. Catalog listing is bounded and uses ID seek order, not a frozen
catalog snapshot; quotas/indexing and team authorization remain future. Refresh
requires the same fixed question and compiled plan; old receipts remain reopenable
when that plan is unavailable. See [GENERAL_CAPTURES.md](GENERAL_CAPTURES.md).

## Core workbench boundaries

[Core basics](CORE_WORKBENCH.md) now provide full-population span-duration summaries,
fixed record-count time charts, reusable package questions, span/log service event
activity and reported parent links. Arbitrary numeric measures, general native
metrics, ingestion lag/drop evidence and production authorization remain absent.
Per-process query limits/cancellation cover general routes, not all legacy routes
or cross-replica/per-tenant quotas. Service discovery is a separate live read and
is not archived by saved inventory. Event age does not establish pipeline health.
