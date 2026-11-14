# Current capabilities and gaps

Goo's main loop is **see current telemetry → isolate a population → follow its
evidence → preserve or reuse the question**. Ordinary OpenTelemetry data is a
useful starting point; a Goo manifest is optional.

This is the current task-oriented capability index, reconciled through milestone
433a, including [operator workbench polish](OPERATOR_WORKBENCH_POLISH.md) and
[fleet service triage](SERVICE_TRIAGE.md).
The [active roadmap](ROADMAP_2026_11_14_CHECKPOINT_430.md) owns execution order;
[validation](../VALIDATION.md) records checks actually run. Historical proposals
below are research inputs, not additional backlogs or claims of vendor parity.

[Query-only exception groups](EXCEPTION_GROUPS.md) identify exact reported type/stack
signatures across logs or spans, with explicit unclassified records and ordinary drill-down.

[Shared record-intent admission](RECORD_INTENT_ADMISSION.md) preserves exact queries
across staging, links and recall; refused links retain held evidence.

[Captured-rate comparisons](CAPTURED_RATE_COMPARISONS.md) normalize captured counts
by each operand's own fixed exposure, including unequal windows and known zero.

[Investigation preservation](INVESTIGATION_CONSOLIDATION.md) retains complete saved-view
configuration, stages reopen without telemetry reads, and restores held Activity.
Portable views can be imported before the first query.

[Unknown-data entry](UNKNOWN_DATA_ENTRY.md) places existing bounded environment
discovery beside Environment, with explicit preview/stage/start and keyboard-safe
Pause/Resume. [Signal availability](SOURCE_CAPABILITIES.md) now separates schema/access from
bounded presence and supports partial-family catalogs. [Metric scope discovery](METRIC_SCOPE_DISCOVERY.md)
adds metrics-only bootstrap; [quiet Activity recovery](ACTIVITY_SPARSE_RECOVERY.md)
stages wider fixed searches. [Instrument preflight](INSTRUMENT_PREFLIGHT.md) shares
Build/Preview readiness across Services and Measurements. Stable typed live-source
refusal/retry remains next; see the [433 review](reviews/CONSOLIDATION_REVIEW_433.md).

## What users can do

| User task | Delivered path | Boundary to keep visible |
| --- | --- | --- |
| Recover a quiet Activity stream | [Wider-window search](ACTIVITY_SPARSE_RECOVERY.md) stages a fixed 15-minute question with the applied selector and held return | Empty loaded observations are not a source census; Run explicitly reads the wider event-time window. |
| See what is happening now | [Live Activity](LIVE_ACTIVITY.md), scoped log/span streams, optional [log volume and literal severity-text context](ACTIVITY_LOG_CONTEXT.md), inspection and [contextual investigation actions](ACTIVITY_NAVIGATION.md) | A bounded live buffer is not the full population; disconnects, gaps and pauses remain explicit. |
| Narrow and inspect ordinary telemetry | [Typed filters](FILTER_EDITOR.md), [counted field values](FILTER_VALUES.md), structured attributes and [qualified field presentation](FIELD_PRESENTATION.md) | Missing and present-empty values differ. A displayed page does not define total counts. |
| Inspect captured activity over time | [Exact captured-record rates](CAPTURED_RECORD_RATES.md) in fixed summaries and ungrouped timelines | Count/window rates describe captured records, not unsampled traffic or application throughput. Clipped buckets use their actual duration. |
| Compare populations | [Population workspace](POPULATION_WORKSPACE.md), distributions, baseline/selection comparisons and [fresh-window comparisons](FRESH_COMPARISON.md) | Independent reads can disagree. Ranked differences are leads, not statistical significance or cause. |
| Follow traces across services | [Cross-service trace search](TRACE_SEARCH.md), qualified witnesses, [waterfalls and trace findings](TRACE_FINDINGS.md) | Conditions are independent within a reported trace ID. Witnesses do not prove ordering, a unique invocation or a complete trace. |
| Continue a trace search | [Explicit ID-ordered pages](TRACE_SEARCH_CONTINUATION.md), exact scope-bound positions and current-page capture | Each page is a fresh read; full counts precede position filtering. No stable or complete page union is promised. |
| Reuse a cross-service search | [Saved trace-search questions](SAVED_TRACE_SEARCH_QUESTIONS.md) and [captured search evidence](CAPTURED_TRACE_SEARCH.md) | Saving intent reads no telemetry. Capture explicitly rereads; reopening evidence is source-free. A fresh child trace keeps its original citation and held Back path. |
| Triage recorded errors and logs | [Error triage](ERROR_TRIAGE.md), [log patterns](LOG_PATTERNS.md), pattern timelines and [time comparisons](LOG_PATTERN_TIME_COMPARISONS.md) | Recorded status/exception evidence is not inferred application failure. Pattern populations retain unclassified coverage. |
| Inspect discovered services and resources | [Service dashboards](SERVICE_DASHBOARDS.md), [incoming HTTP responses](INCOMING_HTTP_RESPONSES.md), [previous-window captured counts](SERVICE_WINDOW_COMPARISONS.md), [host observations](HOST_RESOURCES.md) and [Kubernetes resources](KUBERNETES_RESOURCES.md) | Reported associations do not prove topology, ownership or health. Only supported OTel profiles become executable widgets. |
| Find a service to investigate | [Fleet service triage](SERVICE_TRIAGE.md): qualified Server-span count/change, recorded Error fraction and mean duration across adjacent windows | Four aggregate SQL reads, independent windows, complete bounded groups; captured spans are not unsampled requests or inferred health. |
| Investigate outgoing web dependencies | [Outgoing HTTP observations](OUTGOING_HTTP.md), destination/context groups and ordinary saved questions/captures | Full Client-span record counts are not attempts or unique requests. Addresses do not establish callee identity or remote health; response status remains separate from recorded span Error. |
| Work with native measurements | [Instrument discovery](NATIVE_INSTRUMENT_DISCOVERY.md), [measurement packages](MEASUREMENT_PACKAGES.md), [counters](NATIVE_COUNTERS.md) and [explicit histograms](NATIVE_HISTOGRAMS.md) and [reusable gauge displays](NATIVE_GAUGE_VIEW_REUSE.md) and [counter displays](NATIVE_COUNTER_VIEW_REUSE.md) | Units, temporality, resets and coverage remain typed. Unsupported instruments are visible metadata, not silently executable measurements. |
| See native data before configuration | [Data-first native composition](NATIVE_DATA_FIRST.md) across ordinary, service-expanded and saved native views | Compact scope/state and real first data are visible at natural entry; detailed identity/provenance is disclosed. Source, custody and return owners remain separate. |
| Follow a gauge observation into records | [Logs/spans around an observation](GAUGE_OBSERVATION_CONTEXT.md) from ordinary Measurements, Services and Saved, explicit Run and held Back | A reported timestamp anchors clipped ±60s time association; no causal join, metric-label filter inference or new coverage claim. Saved retains immutable custody; service staging binds the accepted widget. |
| Preserve and hand off an investigation | [Typed question steps](INVESTIGATION_STEPS.md), [immutable captures](GENERAL_CAPTURES.md), [six-family admission](CAPTURE_FAMILY_ADMISSION.md), notes and [linked evidence histories](LINKED_QUESTION_EVIDENCE.md) | Authored intent, source-read basis and custody receipt are separate. Refresh creates a successor; it does not rewrite prior evidence. |
| Keep observations and annotate examples | [Kept observations](KEPT_OBSERVATIONS.md), [tags/notes](WORKBENCH_ANNOTATIONS.md) and [captured-record review](CAPTURE_RECORD_REVIEW.md) | Review targets one immutable occurrence. Labels are local judgments with explicit vocabularies, not ground truth or authenticated reviewer attribution. |
| Author reusable scoped views | [View templates](AUTHORABLE_VIEW_TEMPLATES.md), [ungrouped summary templates](UNGROUPED_SUMMARY_TEMPLATES.md), [numeric histogram views](AUTHORED_NUMERIC_HISTOGRAMS.md), [saved display declarations](SAVED_DISPLAY_DECLARATIONS.md), record tables/lists, timelines and declared facets | Compatibility is checked against the supplied data profile. Configuration is not captured data or permission to execute arbitrary code. |
| Extend displays and automate workflows | [Framework-neutral components](UI_COMPONENTS.md), checked data/kind contracts and the [Go HTTP/OpenAPI contract](API.md) | Display components consume supplied data. Application owners retain reads, navigation, scope and publication authority. |

## Remaining boundaries

These are current gaps or gates, not an automatically scheduled feature list.
The active roadmap adopts a finite breadth/depth/consolidation queue after the
393 whole-product checkpoint: [direct distribution customization](DISTRIBUTION_CUSTOMIZATION.md)
(394) and [grouped count timeline reuse](GROUPED_COUNT_TIMELINE_REUSE.md) (395)
and [held editing/save consolidation](HELD_PRESENTATION_OWNERSHIP.md) (396)
are delivered. Quantitative customization (397) preserves the selected numeric or
duration measure through that same workflow. [Grouped mean duration](GROUPED_MEAN_DURATION_REUSE.md) (398) preserves exact
sum/count support with unavailable empty latency. [Pure helper consolidation](PURE_DISPLAY_HELPERS.md) (399) removes grouped
renderer back edges without moving source ownership.
[Summary authoring](AUTHORED_POPULATION_SUMMARIES.md) (400) uses existing exact
whole-population facts with no forced grouping. [Retained gauge context](RETAINED_GAUGE_RECORD_CONTEXT.md)
(401) and [shared-table context](TABLE_CONTEXT.md) (402) are delivered. [Trace-search continuation](TRACE_SEARCH_CONTINUATION.md) (403) is delivered.
[Ungrouped summary templates](UNGROUPED_SUMMARY_TEMPLATES.md) (404) are also delivered;
[Shared parameter encoding](CLICKHOUSE_PARAMETER_ENCODING.md) (405) is delivered.
[Captured-record rates](CAPTURED_RECORD_RATES.md) (406) are delivered.
[Authored numeric histograms](AUTHORED_NUMERIC_HISTOGRAMS.md) (407) are delivered.
[Passive exports and embedding examples](PASSIVE_DISPLAY_EMBEDDING.md) (408) are delivered.
[Service-window captured-count comparisons](SERVICE_WINDOW_COMPARISONS.md) (409) are delivered.
[Incoming HTTP response context](INCOMING_HTTP_RESPONSES.md) (410) is delivered.
[Shared quantitative chrome](QUANTITATIVE_VIEW_COMPOSITION.md) (411) consolidates
numeric and selected native histogram surfaces and action readiness.
[Fleet service triage](SERVICE_TRIAGE.md) (412) is delivered.
[Log volume/severity-text context](ACTIVITY_LOG_CONTEXT.md) (413) is delivered.
[Recent/pinned scope recall](RECENT_PINNED_SCOPES.md) (414) consolidates browser configuration and held-visit entry.
[Guided service-version comparisons](SERVICE_VERSION_COMPARISONS.md) (415/419) now reuse count/mean and recorded Error comparisons from an exact fixed service scope.
[Ordinary captured-rate comparisons](CAPTURED_RATE_COMPARISONS.md) (418) are delivered.
Recorded Error context (419) and [contextual action consistency](CONTEXTUAL_ACTIONS.md)
(420) are delivered. [Contextual Activity facet entry](ACTIVITY_FACET_ENTRY.md)
(421) reuses existing counted values. [Multi-value selection](MULTI_VALUE_FACETS.md)
(422) stages atomic OR choices within the retained scope. [Measured render/scope
snapshot consolidation](ACTIVITY_SCOPE_SNAPSHOTS.md) (423) is delivered. Contextual field-constraint
editing (424), selected/rest entry (425), preservation (426), and first-use scope
entry (430) are delivered; see the [active roadmap](ROADMAP_2026_11_14_CHECKPOINT_430.md).
Captured-example task discovery and adjacent labeling remain deferred. Delivered
capabilities and remaining boundaries are distinguished below.

| Area | Remaining gap or decision | Current reference |
| --- | --- | --- |
| Native context coverage | Ordinary Measurements gauge nearby-record context is delivered alongside existing counter/histogram interval context. Services/Saved gauge continuation is delivered in 401. Stock and broader resource contexts require their own retained-evidence admission. | [Gauge context contract](GAUGE_OBSERVATION_CONTEXT.md) |
| Further density | [384](DATA_FIRST_FIXED_RESULTS.md) covers checked fixed/retained records; [390](NATIVE_DATA_FIRST.md) delivers actual first-item visibility in ordinary, service-expanded and saved native entry. Other screens and larger-label/population cases remain evidence-driven refinements, not another automatic redesign. | [Current native contract](NATIVE_DATA_FIRST.md) |
| Native outgoing HTTP | The explicit Client-span lens is delivered; the exact seconds explicit-histogram `http.client.request.duration` discovered profile is delivered in 392. Generic duration execution/discovery does not grant Client-span lens authority. | [Outgoing HTTP contract](OUTGOING_HTTP.md) and [native duration contract](HTTP_CLIENT_NATIVE_DURATION.md) |
| Training/review workflow | Contextual task discovery and adjacent captured-occurrence navigation are deferred behind the adopted operator cadence. A disjoint verified kept-occurrence review target remains deferred. Existing captured-record review and export are delivered. | [Review contract](CAPTURE_RECORD_REVIEW.md) and [next roadmap](ROADMAP_2026_11_12_CHECKPOINT_414.md) |
| Display grammar coverage | Gauge, histogram and counter display reuse are delivered. Counter compatibility requires exact unit and matching integer/decimal domain; resource-bound counters, stocks and other grains remain outside that arm. [Checked count-only log-pattern display admission/reuse](LOG_PATTERN_DISPLAY_REUSE.md) is delivered in 391; pattern execution already exists. | [Counter contract](NATIVE_COUNTER_VIEW_REUSE.md) and [gauge/histogram contract](NATIVE_GAUGE_VIEW_REUSE.md) |
| Exponential histograms | General execution is deferred until the source and mathematical representation preserve the required semantics. Discovery metadata does not grant evaluation support. | [Readiness decision](EXPONENTIAL_HISTOGRAM_READINESS.md) |
| Broader customization | Checked templates and declarative displays exist. Arbitrary contributed executable kinds and a general WYSIWYG editor remain wider initiatives. | [Current authoring](AUTHORABLE_VIEW_TEMPLATES.md), [research](reviews/CUSTOMIZATION_RESEARCH_260.md) |
| Production and external platforms | Multi-user tenancy/presence, alert delivery, broader broker/database-server health, warehouse lowering and new storage engines require separate design and validation. | [Roadmap boundaries](ROADMAP_2026_11_12_CHECKPOINT_414.md), [Databricks research](DATABRICKS_METRIC_VIEWS.md) |

## Research and delivery history

Use feature contracts above for current behavior. Read these documents in their
original context; older “next” labels do not activate work today.

- [Checkpoint 381 product/design](reviews/WHOLE_PRODUCT_DESIGN_381.md) and
  [architecture review](reviews/WHOLE_PRODUCT_ARCHITECTURE_381.md): inputs to the
  adopted November 8, 2026 roadmap, not independent execution queues.
- [Customization research, October 30, 2026](reviews/CUSTOMIZATION_RESEARCH_260.md)
  and [grammar proposal 260](VIEW_GRAMMAR_PROPOSAL_260.md): rationale for scoped
  kinds and team-authored views; delivered authoring is described above.
- [Whole-product usability review, October 22, 2026](WHOLE_PRODUCT_USABILITY_REVIEW.md)
  and [implementation proposal](PRODUCT_IMPLEMENTATION_PROPOSAL.md): dated
  investigation-flow and consolidation recommendations.
- [October 24 roadmap](ROADMAP_2026_10_24.md) and
  [execution backlog](BACKLOG_2026_10_24.md): historical status/acceptance records.
- [Databricks research, October 16, 2026](DATABRICKS_METRIC_VIEWS.md): offline
  aggregate compiler direction, not deployment certification.
- [Bitemporality](BITEMPORALITY.md), [annotation storage spike](ANNOTATION_STORAGE_SPIKE.md)
  and [collaboration research](COLLABORATIVE_INVESTIGATIONS_RESEARCH.md): design
  alternatives, not claims that a new store or multiplayer workspace ships.
