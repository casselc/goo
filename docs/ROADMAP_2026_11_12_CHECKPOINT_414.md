# Roadmap after the 414 checkpoint

The preceding breadth/depth/consolidation cycle delivered fleet service triage
(412), full-query log volume (413), and configuration recall (414). See [scope recall](RECENT_PINNED_SCOPES.md) and
[validation](../VALIDATION.md) for its actual status. This queue follows
independent product and architecture source reviews, reconciled below.

Keep the product loop: **current ordinary telemetry → narrow a population → inspect
or compare → preserve/reuse**. Activity remains the entry point. Do not add another
workspace, chart family or history service to extend this cycle.

## 415 — depth: guided service-version comparison

“Did version B differ from A?” Start from one exact service/namespace identity.
Discover reported resource `service.version` values in a fixed window, explicitly
choose two, and stage existing population comparisons. Preserve the service,
operation and typed environment predicates. Missing and empty versions remain
visible; loaded values are not necessarily exhaustive discovery. Show each side's
captured count, fixed window and supported duration/status facts. Deployment
receipts are context, not an inferred version predicate or causal proof.

Reuse current source adapters, aggregate compiler, population comparisons and
held Back. Discovery within one qualified service avoids adding service,
namespace and version as three simultaneous groups. Explicit Run, bounded group
admission, stale-action refusal, independent source checks and desktop/mobile
journeys are required. No fleet percentile from averaged scalar percentiles.

## 416 — breadth: recurring exception groups

“Which repeated exception should I inspect?” Add query-only deterministic,
versioned fingerprints over explicitly reported exception data. Reuse Errors &
exceptions and ordinary record drill-down. Count occurrences within the searched
window, preserve unclassified data, and make identity inputs inspectable. Report
first/last only within that window; do not claim a lifetime first occurrence.

First design the log/span source contract. Span exception event arrays need an
explicit supported projection; a reported exception type alone cannot become a
stack identity. Bound affected service/version summaries and define behavior when
source data is incomplete. Do not silently coalesce missing/empty values.
Defer persistent issue lifecycle, resolution state and cross-deployment deduplication.

## 417 — consolidate: ordinary record-intent admission

Share one pure strict configuration admission contract across link, staging and
history boundaries. Current question-link, predicate-editor and history checks
have different node/depth/field-size limits and recursive field admission. Align
with the authoritative Go execution contract and distinguish a valid query from
a particular editor's supported affordances. Refuse unsupported intent instead
of weakening it or changing its population.

This is a bounded semantic consolidation, not a registry or lifecycle manager.
Use generated legal/illegal queries, exact UTF-8 boundaries, nested unknown-field
controls and actual stage/link/history round trips. Preserve display/source
separation and current query authority. Do not add compatibility formats.

The product reviewer also recommended comparable population-fact presentation.
Keep that as an acceptance requirement for 415/416; extract a shared supplied
model only where the implementation exposes real duplication. Do not prebuild a
new denominator framework without a concrete carrier and source contract.

## Reconciliation and limits

Independent product and architecture reviews agreed on version comparison and
query-only exception grouping. For the consolidation slot, the architect found
concrete admission drift in `question-protocol.mjs`, `predicate-validation.mjs`
and `scope-history.mjs`; that measurable correctness concern takes precedence
over speculative comparison-model abstraction. Relevant existing code includes
`service-dashboard.mjs`, `service-count-comparison.mjs`, `population-workspace.mjs`,
`error-triage.mjs` and `log-patterns.mjs`.

Go is already 1.27 and uses current standard-library/runtime facilities. There is
no measured justification for a runtime rewrite, broad generic API migration or
new query engine. Keep semantic verification packages isolated. Split historical
verification only when measured compile/memory costs block this queue.

Monitoring/alert delivery, authenticated shared deployment, deeper labeling,
arbitrary executable kinds and broader platforms remain behind their existing
identity, authorization, completeness and operational gates. Finish and review
each milestone; review the whole product again after 417. Heavy checks stay
serial and existing telemetry, streams, captures and volumes are preserved.
