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
| Understand instrument preview eligibility | [Instrument preflight](INSTRUMENT_PREFLIGHT.md) |
| Recover a failed live source without losing held rows | [Activity source recovery](ACTIVITY_SOURCE_RECOVERY.md) |
| Search stored log messages from local Find | [Activity message search](ACTIVITY_MESSAGE_SEARCH.md) |
| Inspect OTel severity without trusting producer text | [Numeric severity](ACTIVITY_NUMERIC_SEVERITY.md) |
| Investigate a quiet stream over a wider event-time window | [Activity recovery](ACTIVITY_SPARSE_RECOVERY.md) |
| Find telemetry without knowing environment labels | [Unknown-data entry](UNKNOWN_DATA_ENTRY.md) |
| Follow a complete investigation | [Illustrated preserve/reopen journey](journeys/investigation-reuse.md) |
| Browse worked examples | [Journey index](journeys/README.md) |
| Save questions, views or observations | [Preservation](PRESERVATION.md) |
| Create reusable views | [Authorable view templates](AUTHORABLE_VIEW_TEMPLATES.md) |
| Author a record predicate with CEL | [CEL predicate pilot](CEL_PREDICATE_PILOT.md) |
| Understand the planned visual/expression workflow | [Whole-product CEL integration](CEL_PRODUCT_INTEGRATION.md) |
| Call the API | [API guide](API.md) and [OpenAPI](../api/openapi.json) |

## Understand and extend Goo

- [System overview](SYSTEM_OVERVIEW.md): runtime and data flow.
- [Architecture diagrams](ARCHITECTURE_DIAGRAMS.md): instrumentation, planning and telemetry paths.
- [Measurement packages](MEASUREMENT_PACKAGES.md): code-managed measurement definitions.
- [Shared UI components](UI_COMPONENTS.md): supplied-data display contracts.
- [Contributing](../CONTRIBUTING.md): builds, tests and documentation conventions.

## Project status and development record

[Ad-hoc live telemetry review](reviews/ADHOC_OTLP_REVIEW_2026_11_14.md) covers
manifest-free first use, source compatibility, empty-stream recovery and the
proposed next usability cycle.

- [Current roadmap](ROADMAP_2026_11_14_CHECKPOINT_430.md): adopted next work.
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
