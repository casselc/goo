# Product checkpoint after first reusable compositions — 280

2026-10-31. Read-only source review of the 280 working tree. Shipment and executed
checks belong in [VALIDATION.md](../../VALIDATION.md). No vendor trial, new external
survey or user study is claimed. This builds on the pinned research in
[the 260 product review](PRODUCT_CAPABILITIES_REVIEW_260.md) and the
[276 authoring review](AUTHORING_PRODUCT_REVIEW_276.md).

## Product judgment

The first authoring loop is now credible: scope ordinary OTel data, choose a useful
representation, save it, and apply that presentation to another checked population.
The next user outcome should answer **“When did this scoped population change?”**
inside the reusable view. Add one optional time overview with range selection,
sharing the composition's query owner. Do not start another component migration,
chart vendor integration or general layout editor first.

This advances the existing product loop in [Product essence](../PRODUCT_ESSENCE.md):
observe now → narrow → inspect/compare → preserve → reuse. The references in 260
support HyperDX/ClickStack's event-first filtering, Grafana Explore's shared time
context and Honeycomb's reusable query/board concepts. They do not establish current
SaaS UX or edition parity, and do not justify importing their whole feature suites.

## What exists, and where continuity stops

| Surface | Implemented in the reviewed code | Remaining boundary |
| --- | --- | --- |
| Activity | `ui/stream-entry.mjs` starts with logs/spans and shared filters, observed-field grouping, source-window analysis, latency, explanations and error triage. | Preserve this entry. More authoring controls should remain contextual; saved Activity selectors still have an explicit typed-selector gate in `workbench/activity_views.go`. |
| Services | `ui/service-dashboard.mjs` supports HTTP/RPC/database/messaging-shaped defaults. The shipped HTTP route template enters the same staged editor through `ui/view-template-handoff.mjs`. | Only one service template is declarative; native and other profile owners remain explicit. Candidate discovery is not proof of emitted or complete telemetry. |
| Fixed exploration/reuse | `workbench/view_composition.go`, `ui/view-composition.mjs` and `ui/population-workspace.mjs` express a record table and count/numeric/duration aggregate under one query. 280 adds whole-composition reuse alongside aggregate-only reuse. | The stack has exactly two roles. `population-workspace.mjs` deliberately mounts its existing `ServiceTimeline` only when `!composition`; reusable compositions currently lose the visual time overview. |
| Records and traces | `ui/record-page.mjs` admits bounded immutable pages; trace/context inspection and waterfalls already exist through `ui/record-results.mjs`. | Pages are not full populations or generic plotting Datasets. Another waterfall implementation is not the next gap. |
| Saved/investigations | Configuration views are distinct from captures. Documents support questions, notes, captures, retained evidence, comparisons and deployments. | `workbench/documents.go` has no composition/view-reference card; `ui/documents.mjs` and `ui/investigation-draft.mjs` cannot preserve the authored representation as part of a finding. |

## Ordered priorities

### 1. Next: a shared time overview and range selection

Use existing capabilities rather than a new plotting stack: `timeViewDataset` in
`exploration/view_dataset.go` already exposes checked count intervals;
`ui/source-snapshots.mjs` admits `timeDataset`; `ui/view-schema.mjs` registers
`time-bar`; and `ui/scoped-breakdown.mjs` already stages interval intents.

Extend the bounded composition with an optional timeline child, still referencing
its single parent query. Stage an explicit bucket width when the query does not
supply time intervals, then require Run. Do not hide a query rewrite inside Apply
or let the child own a second fetch. Start with captured counts and human UTC bounds;
query-supported duration measures can reuse their existing exact methods.

A pointer range gesture and keyboard-equivalent selection should stage one shared
half-open window for the record page and aggregate. Run refreshes the owner; Reset
returns to the prior window. Keep local hide/order/measure changes read-free.
Missing/unknown intervals must not turn into invented zeros, and interval coverage
must come from the time aspect, never a record-page or group-completeness claim.

Acceptance: real uneven OTel traffic → saved composition → time selection → explicit
Run → changed records and aggregate over exactly that window → reset/save/reopen.
Check clocks, clipped final buckets, retired selections and source-read counters.
This is a concrete investigation improvement, not universal time-series adoption.

### 2. Preserve a composed finding in an investigation

Add a pinned view-definition reference with an optional separately checked evidence
reference and finding note. Reuse existing document/card and capture custody paths.
Opening configuration stages it; opening evidence uses held data. Neither should
silently run telemetry. A page that advanced after an aggregate has a different
clock: do not stitch the two into a falsely atomic frozen result. Definition-only
cards can ship before a general multi-aspect capture bundle. Existing capture and
question cards remain useful and should not be replaced by an autosubscribing widget.

### 3. Close scope portability at actual entry points

When a user saves Activity's current scope or moves it into a reusable view, retain
the full environment policy, empty/unreported values and filters. The fixed-view
capabilities do not remove Activity's explicit persistence limitation. Close that
concrete handoff rather than opening another whole-product selector migration.
Add further service templates when a real investigation needs them, using the same
composition/compatibility path rather than converting every default dashboard first.

## Roadmap cleanup and deferral

- Close the delivered numeric/duration, display reuse, HTTP template and first
  composition cuts. Supersede their stale “next” labels; keep older reviews dated.
- Put the time-overview outcome before broader layouts or more chart kinds.
  Plotly remains conditional on a visualization the existing components cannot serve.
- Defer arbitrary nesting/grids, full WYSIWYG, user scripts/CEL transforms,
  contributed-kind marketplaces and multiplayer until these actual workflows need them.
- Keep the entire-UI grammar as the direction. Explicit record-page and aggregate
  input profiles are honest progress; neither universal Dataset adoption nor a new
  registry framework is a prerequisite for the next user-visible outcome.
- Keep production authorization/source isolation, durable notifications and storage
  policy as separate operational workstreams. A reusable chart is not an alert engine.
