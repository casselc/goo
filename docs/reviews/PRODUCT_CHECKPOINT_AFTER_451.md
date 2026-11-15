# Product checkpoint after 451

The next cycle should make reported asynchronous span links useful across trace boundaries: **452 target lookup → 453 inverse reference search → 454 consolidation**. This advances general-purpose investigation from an existing, visible dead end. Retain raw metric visualization as the following cycle rather than starting both tracks together.

## What is already present

The waterfall follows an async link when its exact trace/span target is among the returned rows. An outside target is currently plain text ([adapter](../../ui/visualizations.mjs)). Span projection already exports `linked_traces`, `linked_spans`, and `linked_attributes` ([compiler](../../exploration/compiler.go)); new ingestion is not the prerequisite. The [waterfall contract](../UI_WATERFALL.md) explicitly treats links separately from parent edges. The inspected UI has no other consumer that resolves these exported link targets, and existing [trace search](../TRACE_SEARCH.md) searches record conditions rather than inverse link-tuple membership.

Dashboards, waterfalls, database/messaging views, templates, annotations, and captures are delivered capabilities. This cycle should connect the existing investigation surfaces, not recreate those features.

## 452 — Explicit lookup of a reported link target

Replace the outside-returned-span dead end with a staged lookup. Retain the parent environment policy and fixed window; disclose that a target lookup may span other services and namespaces. Require an explicit Run before source acquisition. Return restores the held origin without another source read.

Acceptance requires valid canonical trace/span IDs, exact tuple matching rather than independent array membership, an explicit policy for missing/invalid span IDs, and visible unavailable/ambiguous results. A reported link is not proof that the target was exported, exists in this window, or caused the current operation. A local returned target should continue to use local selection without fetching.

Prerequisites are a bounded admitted target intent, a source response checked against that intent, an application-owned navigation action, and a real cross-trace OTLP journey. Include missing targets, duplicate IDs, environment/window exclusions, cancellation, and keyboard Return. Reuse existing ordinary trace expansion where its scope semantics fit; do not pretend an expanded trace is the exact target lookup itself.

## 453 — Spans referencing this span

Add bounded inverse discovery for a selected exact trace/span tuple: “Which spans report a link to this span?” Validate pair membership at the same link-array index. Return matching span records and complete matching counts within the explicit environment/window, with bounded paging and evidence that supports source admission.

Acceptance needs an independent reference implementation and ClickHouse parity for crossed tuples, duplicate links, duplicate span observations, malformed IDs/arrays, empty results, and bounds. Count matching records according to the declared query grain; do not count duplicate link entries as separate span records. Explain any broader service/namespace search before acquisition. No delivery, processing, uniqueness, completeness-of-ingestion, or causal claim follows from a reported link or its absence.

## 454 — Consolidate reference interactions

After both workflows exist, unify their reference presentation, admission rules, compact scope disclosure, keyboard behavior, and tests. Keep display components passive and navigation/source ownership explicit. Consolidate demonstrated duplication; do not introduce a universal investigation owner or another query DSL. Review desktop/narrow layouts and actual operator journeys before expanding the feature set.

## Following cycle: raw metric visuals

The independent UX review proposes a held raw gauge point/time plot, then bars for one explicit histogram point, then raw-presentation consolidation. Current [gauge](../RAW_GAUGE_OBSERVATIONS.md) and [histogram](../RAW_HISTOGRAM_OBSERVATIONS.md) inspection remain table/detail oriented. This is valuable visual depth after the reference cycle: reuse the existing kind/data grammar, expose gaps and no-recorded values, preserve UInt64 precision, and avoid interpolation, rates, fleet percentiles, or collection-completeness claims. Do not connect these plots to polling or new acquisition merely to draw held data.

## Lessons from 451

451's shared exact duration presentation and predicate-draft owner show the useful consolidation pattern: reusable policy inside the library, simple adapters outside. Canonical identity cannot stand in for validation; null-normalized invalid drafts must remain visibly invalid. Editing and CEL checking stay separate from source acquisition. Shared scope remains protected when adding local predicates.

The 451 validation round includes the full 1,957 UI tests, Go checks, single/paired native verification, and performance checks. Those gates support the delivered changes; they do not prove the proposed link workflows. Each next step still needs its own source-backed journey, negative controls, and independent review.
