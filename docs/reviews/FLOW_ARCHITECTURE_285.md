# Architecture and process checkpoint — 285

> Delivery update: 285 subsequently passed the checks recorded in
> [VALIDATION](../../VALIDATION.md). The review below records its original
> source/screenshot scope and pre-validation status.

2026-10-31. Read-only source review through284 and the285 working tree.
**285 is pending final validation.** This review ran no additional tests and claims
no production performance measurements. Executed checks remain in
[VALIDATION](../../VALIDATION.md). Read alongside the
[product checkpoint](FLOW_CHECKPOINT_285.md) and the
[280 architecture review](ARCHITECTURE_CHECKPOINT_280.md).

## Architectural judgment

The ordinary record investigation path has a useful separation of concerns. Keep
it and deliver the next visual outcome: **show when latency changed beside the
matching records in a saved composition**, after the focused-entry dependency
below makes that composition reachable from recent activity. The source already produces interval
measures. The missing boundary is the authored temporal profile and its
count-specific presenter, rather than a query planner, evaluator or chart engine.
A framework rewrite, universal input registry or another selector migration is not
a prerequisite.

The280–285 sequence closes concrete continuity gaps: local whole-composition reuse;
a held count timeline with exact range selection; separate saved-definition
findings; typed Activity view persistence; record question sharing/recovery; and,
in the pending285 changes, typed recent-population entry. No part of that sequence
makes a fixed composition a rolling subscription or a multi-read atomic receipt.

## Boundaries to retain

| Layer | Current source anchors | Contract to preserve |
| --- | --- | --- |
| Query and measure semantics | `exploration.Compile`, `PlanPopulation`, `ViewMeasureFields`, `timeViewDataset` | Checked environment policy, predicates and bounds own the read. Measures retain their methods, units and support. Planning does not acquire telemetry or drop unsupported predicates. |
| Admission | `ui/source-snapshots.mjs`, `ui/record-page.mjs`, `ui/snapshots.mjs` | Group/time/summary datasets retain their own admitted basis and clock. Rich records remain an explicit bounded-page profile; client JSON cannot create a source completeness witness. |
| Execution authority | `ui/scope-actions.mjs`, population generations and page owners | A supplied key is actionable only under its current owner/epoch. Hidden, detached and retired mounts cannot stage a query or inspect stale records. |
| Presentation | `ui/scope-view.mjs`, `ui/view-values.mjs`, passive elements | Components display supplied inputs and emit intents. Exact values and cells remain separate from approximate plotting coordinates. |
| Preservation | `workbench/documents.go`, saved views, question steps and existing captures | Definitions/questions are reusable intent. Captures have independent custody. Reopening intent stages; opening stored evidence does not query telemetry. |

“Explicit read” has two deliberate forms. Show recent activity is itself a read
request: it resolves one fixed window, obtains a pure plan and executes checked
questions. Opening a saved definition, selected child or question link stages it
until Run. Do not force these distinct entry actions into one misleading label.

## Delivery dependency: focused-signal entry — 286

The product review identified a concrete reachability gap before the duration cut:
recent-population signal tabs change visibility, while authored compositions require
one checked query and an eligible aggregate declaration. Opening a focused signal
must therefore create a staged single-query child, retaining its exact environment,
predicates, fixed bounds and source-supported measure settings, with offset zero.
Keep the multi-signal parent and its held inputs intact. A missing grouping must be
an explicit staged choice before authoring rather than an inferred query rewrite.
Run the child explicitly; switching a visibility tab must not silently discard the
other signal or create a new source read. Reuse the existing child owner, checked
query and declaration paths; no planner/source endpoint is needed.

Deliver that focused entry in286, then the duration timeline in287. This makes the
recommended temporal capability reachable through the ordinary recent flow rather
than relying on a seeded saved composition or a manual configuration edit.

## Priority 1 after that entry: query-supported duration timeline — 287

Extend `TimelineViewDeclaration` and its browser counterpart to a bounded set of
query-supported duration measures. Count plus mean and p95 is an appropriate first
user-facing slice. Require the parent's `summary: span-duration`; if it is absent,
configure it explicitly, stage the root query and require Run. A local measure
switch over already supplied fields must not acquire new inputs. Preserve the
existing single bucket-width authority and strict whole-composition reuse checks.

`timeViewDataset` already appends `ViewMeasureFields` and checked per-cell state,
support and quantile metadata. Go's aggregate declaration validator already uses
that metadata, while `ui/view-declaration.mjs` has the bounded browser capability
list. Reuse that query-supported measure vocabulary rather than adding a third
independent duration-method table. Keep three checks distinct: the query supplies
that measure; the held cell has a value/support; and the requested complete-only
presentation has an interval witness. One does not imply the others.

The justified renderer refactor is small and specific. `count-timeline.mjs`
currently converts counts directly to `BigInt`, labels the unit records and mounts
`goo-time-bars`, while generic `scope-view.mjs:timeBars` already uses canonical
measure formatting, cell support, exact details and duration units. Route authored
count/duration intervals through that canonical temporal display path. Retain the
existing pointer/keyboard range gesture by exposing a bounded range intent from
the shared temporal component, or extracting its passive gesture adapter for both
callers. Do not duplicate duration formatting or calculate percentiles in a new
presenter. A count remains a count; no rate/share semantics appear by extension.

The selected measure must not alter action authority. Interval membership and
contiguity still come from known supplied source intervals and the H1 resolver.
Unknown coverage cannot be selected through a gap; unavailable/no-support measure
cells cannot become invented zeros. Complete intervals do not mean complete
production telemetry, and p95 is an exact empirical captured-span quantile, not an
SLO or a percentile that may be averaged across buckets.

Acceptance should include a real short latency burst, exact clipped UTC buckets,
count/mean/p95 switching with zero local reads, human duration units with exact
method/support details, range Stage and explicit Run, held timeline during record
paging, source-free save/reopen, and reuse refusal when source measure support is
absent. Preserve stale/hidden-owner and no-data/unknown negative controls.

## Priority 2: keep owner cleanup proportional to that feature

`population-workspace.mjs` now owns query staging, execution, paging, composition
editing, catalog reuse, saved acknowledgements and preservation entry actions.
Its current generation, draft revision and owner guards are meaningful; moving
everything into a generic coordinator would obscure them.

Use the existing `stageRootQueries` path when configuring latency support, and the
existing aspect resolver for held time inputs. Extract a small checked-query
configuration helper only if the feature would otherwise duplicate width/summary
staging and reset validation. Do not make presentation adapters execute, publish
or import completeness. Do not merge preview validation with H1 ownership: the
resolver must recheck membership and execution state at the effect boundary even
when the preview used the same pure selection arithmetic.

Preservation-control consolidation recommended by the product review should reuse
existing document context/recovery and immutable objects. It is a UI arrangement
change, not justification for a new universal saved-object store or evidence
bundle. Keep the exact published-view revision guard before attaching a definition.

## Priority 3: measure paging work before a bounded read optimization

`pageSignal` calls the full `/api/explore` path and admits its response, then keeps
only the newest record page while holding the original aggregate and timeline.
That is semantically honest but can repeat aggregate/time SQL and normalization
that the screen does not publish. This is a concrete performance opportunity,
not a measured bottleneck or reason to change held-input semantics.

After the duration cut, measure source statement count, response bytes and page
latency on a representative bounded page. If material, add an explicit record-only
read/profile in the existing source owner and planner rather than silently
replacing held aggregates with fresh ones. Preserve exact policy/predicates,
paging limits, cancellation and independent page clocks. Acceptance must prove
both reduced source work and unchanged aggregate/time identities. Do not infer
aggregate coverage from page extent.

## Process changes justified by recent work

- Validate the whole user path, not merely a codec. A previously gated share or
  recent-population entry can exercise labels, selected-range links, recovery and
  delayed owner publication together. Existing positive and refusal tests should
  change in the same cut as the capability; unsupported families stay explicit.
- Use a bounded verification-only query comparison for known Go omission defaults
  when asserting executed questions. Empty `groupBy` omission is not changed scope.
  Continue asserting typed selector, predicates and fixed bounds explicitly; broad
  normalization must not hide a current-only/effective policy mismatch or rewrite
  stored identity. This avoids repeatedly debugging journey-fixture serialization
  assumptions without weakening source checks.
- Observe completion independently. Count Activity tail connection starts apart
  from internal polling reads; ResourceTiming for a streaming request is complete
  after Pause/close. A click or conforming prefix is not a completed journey.
- Keep serialized heavy Go/browser/build work and bounded disposable artifacts.
  Preserve compact ledgers, current documentation frames and failure diagnostics
  needed for unresolved issues. Repeating a whole campaign after documentation-only
  changes or accumulating successful raw screenshots adds little product value.

## Explicit deferrals

Numeric timeline measures can follow the duration journey using the same cell and
binding vocabulary; signed values require the existing exact-table fallback.
Saved comparison children, native metric temporality/histograms, trace relationship
profiles, rolling composition semantics and team-contributed kinds remain distinct
profiles. Production identity/isolation, quota/retention and durable notification
work remain separate product concerns. None is silently solved by typed record
handoffs or by an authored duration chart.
