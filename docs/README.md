# Goo documentation

Start with the [product introduction and quick start](../README.md).
This index separates using Goo, extending it, and the development record.

## Use Goo

| I want to… | Start here |
| --- | --- |
| Try the sample application | [Local lab](LOCAL_LAB.md) |
| Connect existing ClickHouse telemetry | [Generic startup](GENERIC_STARTUP.md) |
| Understand supported workflows and limits | [Capability guide](CAPABILITY_GUIDE.md) |
| Find which supported signals are available | [Source capabilities](SOURCE_CAPABILITIES.md) |
| Investigate metrics without known service or environment labels | [Metric scope discovery](METRIC_SCOPE_DISCOVERY.md) |
| Return to held metric observations without rereading | [First-use acquisition](FIRST_USE_ACQUISITION.md) |
| Plot held gauge points without another read | [Gauge Points](RAW_GAUGE_OBSERVATIONS.md#points-over-the-held-bank) and [component contract](UI_RAW_GAUGE_POINTS.md) |
| Compare buckets within one exported histogram point | [Histogram Bars / Table](UI_RAW_HISTOGRAM_BUCKETS.md) |
| Inspect exported metric points before evaluation | [Raw gauges](RAW_GAUGE_OBSERVATIONS.md), [raw sums](RAW_SUM_OBSERVATIONS.md) and [explicit histogram points](RAW_HISTOGRAM_OBSERVATIONS.md); [compact inspection](RAW_OBSERVATION_CONSOLIDATION.md) |
| Follow an OTel link or find spans referencing a span | [Reported span links](REPORTED_SPAN_LINKS.md) |
| Find logs or spans around an exported metric point | [Nearby raw-point context](RAW_OBSERVATION_CONTEXT.md) |
| Understand instrument preview eligibility | [Instrument preflight](INSTRUMENT_PREFLIGHT.md) |
| Edit or discard filters without confusing the applied scope | [Consistent Activity drafts](FILTER_CONSOLIDATION.md) |
| Watch error logs beside slow spans | [Independent paired lane filters](PAIRED_ACTIVITY_FILTERS.md) |
| Recover a failed live source without losing held rows | [Activity source recovery](ACTIVITY_SOURCE_RECOVERY.md) |
| Search stored log messages from local Find | [Activity message search](ACTIVITY_MESSAGE_SEARCH.md) |
| Inspect OTel severity without trusting producer text | [Numeric severity](ACTIVITY_NUMERIC_SEVERITY.md) |
| Investigate a quiet stream over a wider event-time window | [Activity recovery](ACTIVITY_SPARSE_RECOVERY.md) |
| Find telemetry without knowing environment labels | [Unknown-data entry](UNKNOWN_DATA_ENTRY.md) |
| Start logs across environments | [Explicit All-environments logs](ALL_ENVIRONMENTS_LOG_ENTRY.md) |
| Follow a complete investigation | [Illustrated preserve/reopen journey](journeys/investigation-reuse.md) |
| Browse worked examples | [Journey index](journeys/README.md) |
| Save questions, views or observations | [Preservation](PRESERVATION.md) |
| Find and stage a saved investigation | [Exact saved-view recall](EXACT_SAVED_VIEW_RECALL.md) |
| Preserve evidence for both sides of a comparison | [Paired comparison capture](PAIRED_COMPARISON_CAPTURE.md) |
| Understand scope and capture controls | [Measured scope/preservation controls](MEASURED_SCOPE_PRESERVATION.md) |
| Create reusable views | [Authorable view templates](AUTHORABLE_VIEW_TEMPLATES.md) |
| Query reported or missing attributes with CEL | [Attribute conditions](CEL_ATTRIBUTE_CONDITIONS.md) |
| Author a record predicate with CEL | [CEL predicate pilot](CEL_PREDICATE_PILOT.md) |
| Understand the planned visual/expression workflow | [Whole-product CEL integration](CEL_PRODUCT_INTEGRATION.md) |
| Call the API | [API guide](API.md) and [OpenAPI](../api/openapi.json) |

## Understand and extend Goo

- [System overview](SYSTEM_OVERVIEW.md): runtime and data flow.
- [Architecture diagrams](ARCHITECTURE_DIAGRAMS.md): instrumentation, planning and telemetry paths.
- [Measurement packages](MEASUREMENT_PACKAGES.md): code-managed measurement definitions.
- [Preserving query populations](SELECTOR_CONSERVATION.md): checked projections and explicit reference expansion.
- [Shared UI components](UI_COMPONENTS.md): supplied-data display contracts.
- [Contributing](../CONTRIBUTING.md): builds, tests and documentation conventions.
- [Gotogether shadow pilot](GOTOGETHER_ACQUISITION_PILOT.md): pinned browser/model checks, completion controls and compiled replay.
- [Paired-filter design matrix](../verification/gotogether/design/paired-filters.md): checked alternatives for independent lane filters under one shared scope; implementation remains planned.
- [Ideal contract workflow](GOTOGETHER_IDEAL_WORKFLOW.md): design exercise spanning authoring, refactoring and operation.

## Project status and development record

[Ad-hoc live telemetry review](reviews/ADHOC_OTLP_REVIEW_2026_11_14.md) covers
manifest-free first use, source compatibility, empty-stream recovery and the
proposed next usability cycle.

- [Current roadmap](ROADMAP_2026_11_14_CHECKPOINT_429.md): adopted next work.
- [Known issues](KNOWN_ISSUES.md): detailed limitations and open gaps.
- [Validation](../VALIDATION.md): checks actually executed, newest first.
- [Implementation plan](IMPLEMENTATION_PLAN.md): milestone history and execution guidance.
- [Detailed architecture](ARCHITECTURE.md): implementation contracts accumulated during development.

Dated roadmaps, reviews, RFCs and proposals record decisions at their writing time.
They can contain planned or superseded behavior; they are not independent feature
promises. Use the capability guide for current behavior and the current roadmap
for priorities. The [historical README](history/README_BEFORE_DOCS_CLEANUP.md) and
[original handoff](../START_HERE.md) are background, not getting-started guides.

## Refinement and held selections

[Shared refinement and aggregate handoff](REFINEMENT_CONSOLIDATION.md) documents
the domain operations, owner-specific interaction guards and remaining CEL work.

[Checked CEL planning](CEL_CHECKED_PLANNING.md) explains the shared CLI/HTTP
contract, owner bindings, Unicode diagnostics and source-free limits.

[CEL representation and ownership](CEL_REPRESENTATION_OWNERSHIP.md) explains
exact conversion coverage and the pure additional-condition owner before UI mounting.

[Contextual CEL editing](CEL_CONTEXTUAL_EDITOR.md): optional Conditions /
Expression inside Activity and fixed-question Filters; Check stages, Apply/Run reads.

[Saved CEL questions](CEL_SAVED_QUESTIONS.md): preserve exact checked source,
reopen its original base, and use existing saved-view YAML import/export.

[Whole-product CEL acceptance](CEL_WHOLE_PRODUCT_ACCEPTANCE.md) connects Activity,
Services/native record context and saved reuse, with the next operator roadmap.

[Refine by an attribute](ATTRIBUTE_REFINEMENT.md): Include, Exclude, reported and missing across record details, columns, facets and CEL.

[Numeric CEL conditions](CEL_NUMERIC_ATTRIBUTES.md): exact typed thresholds and explicit parse states over resource/record attributes.

[Logs and spans together](PAIRED_ACTIVITY.md): independent live lanes, local inspection and read-free retained return.
