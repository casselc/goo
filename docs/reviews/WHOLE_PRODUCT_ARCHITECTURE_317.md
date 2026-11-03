# Whole-product architecture checkpoint — 317

Status: independent source/planning review after 316; no product source edits or
new browser measurement. Actual 315/316 journeys and `VALIDATION.md` remain the
execution evidence. Read against the active `ROADMAP_2026_11_03.md`, known issues
and 312–316 contracts, not the older parallel feature queues.

## Recommendation: finish the shared chart surface

317 should put **actual comparable plotted marks** in the first active viewport of
the existing parent/cohort incident flow, using the shared chart header/read-facts
layer. This is the strongest evidenced remaining task failure: 315 measured the
narrow parent chart below the viewport and a visible cohort chart container whose
visible portion was metadata rather than bars. Facet and ancestry compaction were
real improvements; they did not finish visual population comparison.

Make this a task outcome, not a 30-pixel styling target. At desktop and 390×844,
users should identify the measured field, denominator and materially different
population from the marks without scrolling past repeated metadata. Retain readable
touch/keyboard controls, literal authored titles and visible invalid/unknown/partial
states. Exact read clocks/receipts and fuller quality facts remain accessible in one
existing owner-local disclosure. No new chart kind, source demand, query profile,
renderer vendor or execution owner is needed.

## Concrete shared seams

`ui/scoped-breakdown.mjs:26–73` renders qualified context and a toolbar before
`goo-chart`. `ui/scope-view.mjs:221` then constructs a chart frame with another title,
full ISO captured clock/state and a five-item quality strip before geometry.
`ui/chart-frame.css` adds 16px outer padding, title/meta separation and a 12px header
margin. `quality()` (`ui/scope-view.mjs:26`) repeats origin, extent, value, sampling
and freshness. Each fact has valid meaning; their current placement can dominate
the narrow viewport before the user sees a comparison.

Consolidate these specific layers rather than adding local CSS overrides per screen.
The passive chart owns supplied geometry/chrome; the application still owns scope,
read effects and evidence. Keep a compact essential state/denominator/value-method
summary beside the actual title/marks, and move repeated full detail to accessible
read facts. Do not globally suppress quality, silently change percent denominators,
make small values incomparable, or regard a frame's top edge as proof of visible
data. Validate marks and labels in both the parent and nested cohort.

## Broader priorities and structural risks

**Environment coverage is a real next product boundary, not a one-field fix.**
316 correctly supports only current-only named service handoff with a tested
semantically equivalent Services encoding. Effective, empty and unreported scopes
are explicitly refused. Inventory/discovery share an environment-string scope;
service record constructors, dashboard native metrics, annotations and Activity-view
libraries also consume that representation. Broadening only the first API would
mislabel downstream scope. After 317, choose one end-to-end policy outcome with
explicit native/entity limitations instead of opportunistic selector propagation.
The existing refusal is an honest bounded limitation, not evidence of a new scope
corruption bug.

**Public component budgets remain a reliability follow-on.** The facet caps 1,000
items and attribute JSON length, while property input can contain very large
strings/DTOs. Current actual source responses are bounded, so this is a public
extension boundary rather than a demonstrated source incident. Before adding a
second large supplied-control role, define one property/attribute byte/structure
budget, explicit refusal and equivalent tests. Avoid a new scheduling framework
until repeated readiness rendering is measured.

**Ownership consolidation must follow real callers.** The workspace still
coordinates many render/refresh/draft callbacks, but recent fixes restored correct
leases, cancellation and bounded facet lookup lifetime without replacing scope
sessions. Extract typed control/comparison adapters only when changing that seam
for a real task. No universal navigation, snapshot format or automatic evidence
retention system is justified. Named Activity/service/resource/trail/child returns
have different contracts and must stay distinct.

**Source performance work should stay measured.** 314 reduced one supported
fixed-baseline demand from four to two statements while preserving aggregate
meaning; latency did not improve in its measured run. Selected records, stored
captures and held parents have different purposes and cannot be refreshed or
cached interchangeably. Do not restart a broad query-engine migration or optimize
requests that merely have similar windows but different namespace selectors.

## Finite sequence and gates

1. **317 shared chart data-first cut:** capture current parent/cohort frames, then
   prove actual marks, comparable scale and critical state/denominator appear at
   both widths. Presentation/disclosure adds zero source requests; exact values,
   title literals, selection/focus and read facts survive.
2. **One end-to-end environment-policy outcome:** pick a concrete effective-scope
   service task and define exact supported record/native/entity behavior before
   expanding its request contracts. No fallback to a name string or sentinel.
3. **Public supplied-control input budget:** one bounded property/attribute
   admission contract and meaningful over-budget/refusal tests before further
   control expansion, if not required sooner by measured 317 rendering.

Do not mechanically add another service-family facet or reopen the whole feature
catalog between these steps. Design agrees with the chart-first priority, while
broader policy support needs a deliberate task-sized contract. This review found
no newly reproduced scope/source correctness defect requiring emergency changes;
its first priority is grounded in maintained native measurements, not aesthetics
alone.

## Implementation source review

The selected cut changes presentation only: geometry precedes same-owner Chart
settings, the caller's Find/group tools move into those settings, and scope-view
quality keeps extent, value knowledge, upstream sampling and nonfixed freshness
visible. Full origin, exact captured clock and every quality field remain in Read
facts. No snapshot schema, source request, comparison admission or action owner was
changed. Existing attribute caching preserves the same admitted snapshot identity;
Read facts open/focus state is retained only for local rerender of that snapshot.

Review found one concrete relocation dependency: display-source Apply synchronized
Find through the host's first input. Once Find follows the completeness checkbox
inside settings, that selector targets the wrong control and leaves visible Find
stale. Replace that broad selector with explicit Find identity and test source-to-
property synchronization; this is required before final source approval. Other
inspected input selectors target unrelated grouping/comparison owners.

The Find relocation defect is resolved with an explicit `data-display-find` binding
on the caller-owned input and matching source-Apply selector. The completeness
checkbox is no longer an accidental synchronization target. Focused checks pass as
reported by the implementation owner; native source-to-Find/no-read validation
remains with verification. Final architecture source review is accepted with no
remaining blockers. `git diff --check` passed. First-viewport plotted-mark and
keyboard/receipt claims must come from actual after frames and interactions.

## First measured after gate and bounded correction

The first actual after run passed desktop but did **not** meet the narrow parent
mark target: the second plotted mark ended at 916.8px, below the 844px viewport.
Record that failure rather than treating chart-container visibility as success.
The selected follow-up is compact default canvas chrome and successful-read status,
plus a readable shared narrow Run/Fresh row. It must preserve literal authored
navigation/title, full accessible action names, exact scope facts and custom
warnings/errors. Do not reorder authored components merely to satisfy geometry.

The second source cut is accepted: only the generated default canvas heading is
shortened; explicit navigation and authored titles remain literal. Visual Edit/
Run/Fresh labels retain full accessible action names, and the narrow acquisition
row does not reorder components. Only one known successful-read status literal is
compacted to **Applied · independent reads.**, with its full explanatory text in
the title; staged, diagnostic, warning, error and caller-supplied text remain
unchanged. No snapshot quality or source action semantics changed. The failed first
narrow geometry gate remains recorded above; actual rerun decides delivery.
