# Proposed operator cycle after 462

Status: **proposal only**, 2026-11-16. This does not adopt a new execution order or
claim 462 acceptance; the implementation owner records that separately. Proposed
numbers below are reserved only if this plan is adopted.

## Recommendation

Return to general telemetry reach after the held-chronology cycle. Use one
breadth slice, one distinct metric-coverage depth slice, then consolidate the
capability/presentation paths they exercise. Do not extend chronology with another
format, toolbar or mini-framework simply because it was the last edited surface.

The basis is the [active roadmap](../ROADMAP_2026_11_14_CHECKPOINT_429.md),
[current gaps](../KNOWN_ISSUES.md), [vendor gap review](PRODUCT_GAP_ANALYSIS_2026_11_12.md),
[ad-hoc OTLP review](ADHOC_OTLP_REVIEW_2026_11_14.md), and the accepted
[461 UX](RELATED_HELD_461_UX.md) and [accessibility](RELATED_HELD_461_ACCESSIBILITY.md)
findings. This is repository reconciliation, not fresh vendor research or a
hands-on competitive benchmark.

## Next three outcomes

| Proposed slice | Operator outcome | Scope and reuse | Acceptance |
| --- | --- | --- | --- |
| **463 — breadth: explicit All-environments span entry** | “I have span data but do not know its environment names; show me recent spans, then let me narrow.” | Extend the existing typed All selector from its deliberately log-only boundary into ordinary span Activity and fixed span questions. Reuse winning-environment facts and explicit narrowing from 437, existing span adapter, checked predicates, bounded tails and ordinary question carriers. Keep trace search/expansion, native evaluation and specialized profiles narrow until their own contracts support wide scope. This is a source/query capability extension, not a blank-input wildcard or source authorization. Do not silently enable All for paired Activity. | Real spans-only/no-manifest fixtures with named, empty, missing, legacy and conflicting environment labels; explicit All draft then Start/Run. Blank, a literal `all` label and failed discovery never broaden a question. Exact typed scope survives ordinary saved-view/capture/CEL carriers that admit it. Per-row reported environment and narrow staging remain available; unsupported context actions explain their refusal without replacing the current work. No source read on staging, disclosure or Return. Guard cancellation, stale completion, rows/bytes/interval budgets and denied sources. Keyboard/320/390/1440 journeys and a positive Bombadil campaign. Backend admission is authoritative; a UI capability flag cannot bypass it. |
| **464 — depth: inspect reported exponential-histogram exports** | “The SDK emits exponential histograms. Let me inspect what actually arrived even if it cannot yet be evaluated.” | Add a bounded raw-point inspection profile from the existing discovered `exponential-histogram` instrument. Reuse raw observation acquisition/held paging, exact clocks, diagnostics and shared table/inspector grammar. Preserve source scale, signed offsets, positive/negative count vectors, zero count, flags, temporality and occurrence multiplicity. Start with source facts/table, not a merged distribution, percentiles or a new chart family. Keep evaluation ineligible under the existing 374 gate. | First reconcile the current exporter with [374 readiness](../EXPONENTIAL_HISTOGRAM_READINESS.md): zero threshold and optional-statistic presence are lost. Their absence must remain unknown, never invented as zero or “reported.” A real source oracle must establish the fields the raw profile can expose. Use canonical UInt64 counts, exact vector/sum checks and explicit invalid/absent states; bound arrays/bytes and refuse wrong instrument/window identity. Missing/denied optional table stays distinct from successful empty. Held Find, paging, inspection and Return issue no new queries. UI must say raw inspection supported while evaluation remains unsupported. If the source cannot supply trustworthy raw identity/count vectors, stop at a documented source blocker rather than weaken the contract. No finite-width geometry or exact bucket interval is claimed when the required bounds/zero-threshold facts are unavailable. |
| **465 — consolidate: consistent capability and uncertainty entry** | “Across Activity, Services and Measurements I can tell what can be read, displayed or evaluated, and the same action means the same thing.” | Reconcile the existing source-family readiness and instrument support projections exercised by 463/464. Use one checked presentation of raw-read, evaluation and supported scope capabilities; keep acquisition and backend profile admission in their current owners. Reuse native disclosure/action components and data certainty, rather than add another registry or workspace. Finish only demonstrated duplication and compactness issues. | Walk spans-only, exponential-only, mixed, absent-table and refused-source fixtures through the three entry surfaces. Raw-read support must not imply evaluation eligibility; unknown source facts must not become a healthy/zero result. Wide-span support must not leak into narrow-only profiles. Stable action names, visible refusal/recovery, retained local state and no read on local mode switches/disclosure/Return. Measure request counts and repeated projection/copy work before claiming performance gains. Independent UX/a11y review, forced colors and refreshed storyboards/actors; property/model rejecting cases target the newly extended capability boundaries. |

Before 463 implementation, architecture should review its exact span owner/carrier
matrix. The existing [All-log contract](../ALL_ENVIRONMENTS_LOG_ENTRY.md) explicitly
refuses spans and specialized continuations; removing one guard globally would be
incorrect. Before 464, source and representation reviews must sign off the raw
profile independently of executable native measurements. Neither gate requires a
new query engine, arbitrary SQL editor or direct OTLP receiver.

## Fourth outcome: stop and review the product

After the three slices, execute a whole-product review of first useful data,
explicit narrowing, comparison, reuse and preservation using ordinary OTLP.
Choose the next cycle from observed friction and measured source coverage. In
particular, decide whether the next major need is operational deployment
foundations or monitoring, rather than automatically continuing native-metric or
chronology work. This checkpoint is a decision outcome, not a fourth feature to
implement speculatively.

## Reconciled and deferred work

- **Already delivered, not new gaps:** service/operation and host/Kubernetes
  dashboards; fleet triage, Activity log-volume context, version comparisons,
  exception groups; Recent/Pinned and exact saved-view recall; paired evidence;
  trace search/waterfalls, reported asynchronous references, explanations and
  general captures; raw gauge/sum/explicit-histogram inspection; independent lane
  filters, saved paired definitions, held chronology and related held records.
  The November 12 priorities and older early-workbench tables are historical
  sequencing, not another current queue.
- **462 remains a bounded consolidation.** Its reading-state/projection changes
  should finish validation before 463 begins. The 461 inspector restoration,
  visible span anchor and non-color selection fixes are accepted; do not plan
  them again. Compact inline row-context disclosure, a clearer two-level return
  hierarchy and the redundant paused-toolbar action are valid small refinements,
  not prerequisites for a broad next feature. Do not add a breadcrumb control
  merely to satisfy a diagram.
- **Still separate depth proposals:** durable chronology presentation, merged
  selection/capture and paired baseline attachments. Definitions are not frozen
  observations; do not serialize runtime/source authority to make them portable.
- **Still semantic/source-gated:** exponential evaluation, general producer
  identity without reported instance IDs, merged/fleet distributions, whole-query
  CEL, arbitrary lowering plugins and warehouse compilation. 464 raw inspection
  does not satisfy those gates or revive an old exponential evaluator plan.
- **Still operationally gated:** durable monitors/notifications, SLO burn rates,
  multi-user sharing, authorization/source scoping, auditing and per-tenant
  budgets. A production monitoring service is a real vendor gap, but it is not a
  small add-on to a local saved query. Actual collector/exporter loss/lag signals
  remain distinct from event age and should inform that future decision.
- **Do not resume by default:** classifier labeling queues, multiplayer whiteboard,
  RUM/session replay, continuous profiling, synthetic monitoring or generic AI.
  They may be useful later but require separate collection, identity or product
  commitments. Shared views, notes and existing investigation primitives remain
  usable without them.

## Documentation hygiene discovered during planning

The active roadmap's opening status still says 460 browser acceptance is pending,
while its later sections and validation record show 460/461 accepted. Reconcile
that summary when adopting the next cycle. Several old gap paragraphs also name
now-delivered chronology/lane functionality as future work. Preserve historical
research with clear dated status; the current execution pointer and capability
index should have one authoritative meaning. Do not turn docs cleanup into an
unbounded prerequisite to 463.
