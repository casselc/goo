# Current capabilities and gaps

Goo's main loop is **see current telemetry → isolate a population → follow its
evidence → preserve or reuse the question**. Ordinary OpenTelemetry data is a
useful starting point; a Goo manifest is optional.

This is the current task-oriented capability index, reconciled through milestone
389. The [active roadmap](ROADMAP_2026_11_09_CHECKPOINT_387.md) owns execution order;
[validation](../VALIDATION.md) records checks actually run. Historical proposals
below are research inputs, not additional backlogs or claims of vendor parity.

## What users can do

| User task | Delivered path | Boundary to keep visible |
| --- | --- | --- |
| See what is happening now | [Live Activity](LIVE_ACTIVITY.md), scoped log/span streams, inspection and [contextual investigation actions](ACTIVITY_NAVIGATION.md) | A bounded live buffer is not the full population; disconnects, gaps and pauses remain explicit. |
| Narrow and inspect ordinary telemetry | [Typed filters](FILTER_EDITOR.md), [counted field values](FILTER_VALUES.md), structured attributes and [qualified field presentation](FIELD_PRESENTATION.md) | Missing and present-empty values differ. A displayed page does not define total counts. |
| Compare populations | [Population workspace](POPULATION_WORKSPACE.md), distributions, baseline/selection comparisons and [fresh-window comparisons](FRESH_COMPARISON.md) | Independent reads can disagree. Ranked differences are leads, not statistical significance or cause. |
| Follow traces across services | [Cross-service trace search](TRACE_SEARCH.md), qualified witnesses, [waterfalls and trace findings](TRACE_FINDINGS.md) | Conditions are independent within a reported trace ID. Witnesses do not prove ordering, a unique invocation or a complete trace. |
| Reuse a cross-service search | [Saved trace-search questions](SAVED_TRACE_SEARCH_QUESTIONS.md) and [captured search evidence](CAPTURED_TRACE_SEARCH.md) | Saving intent reads no telemetry. Capture explicitly rereads; reopening evidence is source-free. A fresh child trace keeps its original citation and held Back path. |
| Triage recorded errors and logs | [Error triage](ERROR_TRIAGE.md), [log patterns](LOG_PATTERNS.md), pattern timelines and [time comparisons](LOG_PATTERN_TIME_COMPARISONS.md) | Recorded status/exception evidence is not inferred application failure. Pattern populations retain unclassified coverage. |
| Inspect discovered services and resources | [Service dashboards](SERVICE_DASHBOARDS.md), [host observations](HOST_RESOURCES.md) and [Kubernetes resources](KUBERNETES_RESOURCES.md) | Reported associations do not prove topology, ownership or health. Only supported OTel profiles become executable widgets. |
| Investigate outgoing web dependencies | [Outgoing HTTP observations](OUTGOING_HTTP.md), destination/context groups and ordinary saved questions/captures | Full Client-span record counts are not attempts or unique requests. Addresses do not establish callee identity or remote health; response status remains separate from recorded span Error. |
| Work with native measurements | [Instrument discovery](NATIVE_INSTRUMENT_DISCOVERY.md), [measurement packages](MEASUREMENT_PACKAGES.md), [counters](NATIVE_COUNTERS.md) and [explicit histograms](NATIVE_HISTOGRAMS.md) and [reusable gauge displays](NATIVE_GAUGE_VIEW_REUSE.md) and [counter displays](NATIVE_COUNTER_VIEW_REUSE.md) | Units, temporality, resets and coverage remain typed. Unsupported instruments are visible metadata, not silently executable measurements. |
| Follow a gauge observation into records | [Logs/spans around an observation](GAUGE_OBSERVATION_CONTEXT.md) from ordinary Measurements, explicit Run and held Back | A reported timestamp anchors clipped ±60s time association; no causal join, metric-label filter inference or new coverage claim. Saved/service adoption is deferred. |
| Preserve and hand off an investigation | [Typed question steps](INVESTIGATION_STEPS.md), [immutable captures](GENERAL_CAPTURES.md), [six-family admission](CAPTURE_FAMILY_ADMISSION.md), notes and [linked evidence histories](LINKED_QUESTION_EVIDENCE.md) | Authored intent, source-read basis and custody receipt are separate. Refresh creates a successor; it does not rewrite prior evidence. |
| Keep observations and annotate examples | [Kept observations](KEPT_OBSERVATIONS.md), [tags/notes](WORKBENCH_ANNOTATIONS.md) and [captured-record review](CAPTURE_RECORD_REVIEW.md) | Review targets one immutable occurrence. Labels are local judgments with explicit vocabularies, not ground truth or authenticated reviewer attribution. |
| Author reusable scoped views | [View templates](AUTHORABLE_VIEW_TEMPLATES.md), [saved display declarations](SAVED_DISPLAY_DECLARATIONS.md), record tables/lists, timelines and declared facets | Compatibility is checked against the supplied data profile. Configuration is not captured data or permission to execute arbitrary code. |
| Extend displays and automate workflows | [Framework-neutral components](UI_COMPONENTS.md), checked data/kind contracts and the [Go HTTP/OpenAPI contract](API.md) | Display components consume supplied data. Application owners retain reads, navigation, scope and publication authority. |

## Remaining boundaries

These are current gaps or gates, not an automatically scheduled feature list.
The active roadmap adopts a finite breadth/depth/consolidation queue after the
387 whole-product checkpoint. Delivered layout acceptance and remaining native
entry work are distinguished below.

| Area | Remaining gap or decision | Current reference |
| --- | --- | --- |
| Native context coverage | Ordinary Measurements gauge nearby-record context is delivered alongside existing counter/histogram interval context. Saved/service gauge continuation, stock and broader resource contexts require their own retained-evidence admission. | [Gauge context contract](GAUGE_OBSERVATION_CONTEXT.md) |
| Further density | [384](DATA_FIRST_FIXED_RESULTS.md) delivers first-item visibility in checked fixed/retained record surfaces. Native identity, diagnostics and controls still need data-first consolidation with real initial-viewport checks. | [390 cut](ROADMAP_2026_11_09_CHECKPOINT_387.md#390--consolidate-data-first-native-entry-and-component-boundaries) |
| Native outgoing HTTP | The explicit Client-span lens is delivered; adopting `http.client.request.duration` as an executable native profile remains separate depth. Generic instrument discovery does not grant that binding. | [Outgoing HTTP contract](OUTGOING_HTTP.md) |
| Training/review workflow | A contextual task chooser and a disjoint verified kept-occurrence review target remain candidates. Existing captured-record review and export are already delivered. | [Review contract](CAPTURE_RECORD_REVIEW.md) and [deferred alternatives](ROADMAP_2026_11_09_CHECKPOINT_387.md#reconciled-deferred-work-and-checkpoint) |
| Display grammar coverage | Gauge, histogram and counter display reuse are delivered. Counter compatibility requires exact unit and matching integer/decimal domain; resource-bound counters, stocks and other grains remain outside that arm. Checked log-pattern display admission remains deferred. | [Counter contract](NATIVE_COUNTER_VIEW_REUSE.md) and [gauge/histogram contract](NATIVE_GAUGE_VIEW_REUSE.md) |
| Exponential histograms | General execution is deferred until the source and mathematical representation preserve the required semantics. Discovery metadata does not grant evaluation support. | [Readiness decision](EXPONENTIAL_HISTOGRAM_READINESS.md) |
| Broader customization | Checked templates and declarative displays exist. Arbitrary contributed executable kinds and a general WYSIWYG editor remain wider initiatives. | [Current authoring](AUTHORABLE_VIEW_TEMPLATES.md), [research](reviews/CUSTOMIZATION_RESEARCH_260.md) |
| Production and external platforms | Multi-user tenancy/presence, alert delivery, broader broker/database-server health, warehouse lowering and new storage engines require separate design and validation. | [Roadmap boundaries](ROADMAP_2026_11_09_CHECKPOINT_387.md), [Databricks research](DATABRICKS_METRIC_VIEWS.md) |

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
