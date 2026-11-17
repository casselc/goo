# Simplicity and understandability review

This review began at `9119a967` on 2026-11-16 and continued on 2026-11-17.
The user requested component-by-component reviews and walkthroughs, followed by
independent whole-system Go, architecture and product reviews. This package changes
documentation only. Recommendations below do not claim implementation.

Start with the [code walkthrough index](../../walkthroughs/README.md) when learning
the application. Read this index when choosing simplification work.

## Component reviews

| Component | Walkthrough | Findings |
| --- | --- | --- |
| Runtime and source access | [Runtime](../../walkthroughs/runtime.md) | [Review](runtime.md) |
| Semantic/query/measurement core | [Semantics](../../walkthroughs/semantics.md) | [Review](semantics.md) |
| Discovery, dashboards and metric entry | [Discovery and metrics](../../walkthroughs/discovery-metrics.md) | [Review](discovery-metrics.md) |
| Live Activity and application state | [Activity](../../walkthroughs/activity.md) | [Review](activity.md) |
| Shared presentation and authoring | [Presentation](../../walkthroughs/presentation.md) | [Review](presentation.md) |
| Preservation and investigations | [Preservation](../../walkthroughs/preservation.md) | [Review](preservation.md) |
| Instrumentation, verification and delivery | [Delivery](../../walkthroughs/delivery.md) | [Review](delivery.md) |

The component reviews were completed before the holistic reviewers were assigned.
Reviewers followed current source and registrations, rather than treating old
roadmaps as proof of current behavior. Targeted checks and their limits are
recorded in each report; overlapping checks must not be summed as independent
coverage.

## Holistic reviews

| Perspective | Report |
| --- | --- |
| Go, runtime/stdlib and API simplicity | [Go specialist](holistic-go.md) |
| Whole-system boundaries and ownership | [Architect](holistic-architecture.md) |
| Product coherence and keep/defer/stop decisions | [Product manager](holistic-product.md) |

All three independently reviewed the completed component package and
representative source. Their recommendations are reconciled into the
[staged implementation proposal](implementation-plan.md). The reconciled next step now includes the critic's challenge: choose the
core/optional boundary before refactoring everything. The first concrete cut is
reservation lifecycle coupling in generic journeys, followed by readable retained
core owners—not a rewrite or new framework.

## Adversarial essentials review

The [adversarial critic](adversarial-essentials.md) challenges the specialist
consensus: active code is not necessarily essential, and formatting all of it
could entrench an oversized product. The critic recommends Explore/Services/Saved
as the default conceptual spine, retains reusable grammar and annotation, and
names optional/removal candidates with losses and gates. It also challenges
per-milestone review/test growth and automatic 464 sequencing. The reconciled
proposal preserves those disagreements rather than implying unanimous approval
of every cut. No code, data or requested capability was removed.

The generic/reservation journey split has since been implemented and validated;
the user has since confirmed the [product boundaries](implementation-plan.md#accepted-product-decisions).
Those choices are accepted; implementation and remaining owner refactors are
pending. Historical review disagreements remain recorded below.

## What is already established

- Dense authored JavaScript makes small physical line counts misleading. A reader
  has to unpack many responsibilities within individual statements and closures.
- Several application owners combine state machines with rendering and unrelated
  operation lifetimes. Naming the owners and transitions is more valuable than
  splitting files mechanically.
- Source schema, semantic/profile admission, supplied-data validation and runtime
  provenance checks protect different boundaries. They are not interchangeable
  validation layers that can all be deleted.
- No complete production subsystem was proven unused by the component reviews.
  A few aliases and indirections are demonstrably redundant; larger removal
  proposals need caller/route and product decisions.
  An active path is not a reason to keep it: utility and maintenance cost still
  need a product decision, as the adversarial review explicitly challenges.
- There are concrete delivery costs worth removing: implicit reservation-lab
  dependencies in generic journeys, inconsistent default nested-module checks,
  duplicated CI work and oversized historical entry documents.
- Runtime performance improvements need measurement. These reports identify
  repeated compilation, catalog scanning and copying candidates, not benchmarked
  regressions or evidence that removing guards is safe.

## Acceptance for follow-on changes

A component should have a clear read-first path and a small set of named owners.
A new reader should be able to trace one request or interaction, identify source
and write effects, explain retirement/Return, and locate focused rejecting tests.

Keep formatting-only changes separate from behavior and ownership changes.
Preserve scope, uncertainty, event/read/custody clocks and explicit publication
outcomes. Confirm suspected dead code through registrations and callers before
removal. Use measured request counts and meaningful property/model checks where
state or authority changes; do not create more diagnostic artifacts by default.

Documentation links are checked locally; the review does not claim a new full
Go/UI suite, live telemetry campaign, deployment audit or performance benchmark.
The last executed product milestones remain in [validation](../../../VALIDATION.md).
