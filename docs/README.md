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
| Investigate a quiet stream over a wider event-time window | [Activity recovery](ACTIVITY_SPARSE_RECOVERY.md) |
| Find telemetry without knowing environment labels | [Unknown-data entry](UNKNOWN_DATA_ENTRY.md) |
| Follow a complete investigation | [Illustrated preserve/reopen journey](journeys/investigation-reuse.md) |
| Browse worked examples | [Journey index](journeys/README.md) |
| Save questions, views or observations | [Preservation](PRESERVATION.md) |
| Create reusable views | [Authorable view templates](AUTHORABLE_VIEW_TEMPLATES.md) |
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
