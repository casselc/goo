# Whole-product architecture and Go review — 399

> Acceptance closure: 399 is validated separately in [the validation record](../../VALIDATION.md).
> [The adopted checkpoint](../ROADMAP_2026_11_10_CHECKPOINT_399.md) reconciles the
> recommendations below. The original independent review scope remains unchanged.

November 10, 2026. Independent read-only review after delivered 397 quantitative
customization and 398 grouped mean duration. The bounded 399 helper extraction
was underway during this review; its acceptance belongs in
[validation](../../VALIDATION.md), not this document. No browser campaign,
benchmark, build, dependency upgrade or product-source edit was performed here.
Recommendations below are inputs to one reconciled roadmap, not parallel queues.

## Product and layers

The product center remains **current ordinary telemetry → isolate a population →
inspect related evidence → preserve or reuse its question and presentation**.
Manifests add meaning but are not a prerequisite. Live Activity, typed filtering,
trace search/waterfalls, comparisons, service/resource dashboards, native
measurements, captured review and export are delivered capabilities. They should
not be relisted as missing merely because a vendor has a differently named page.

The present boundaries remain useful:

| Layer | Responsibility |
| --- | --- |
| Discovery and typed identity | Qualified fields, signal capabilities, exact missing/empty/literal identity and environment policy. |
| `exploration`, `measurements`, `metrics` | Checked questions, source-specific plans/budgets, pure arithmetic and explicit unknown/reset/support semantics. |
| `scopedview` and source adapters | Exact supplied values, units, support, finite geometry and captured-query completeness. |
| Passive components | Local display, coordinates and typed selection; no source/store ownership. |
| Application owners | Draft versus accepted question, current capabilities, source requests and retained navigation. |
| `workbench` | Immutable configurations/evidence and exact-ID loading; Apply carries presentation without donor observation authority. |

The grammar is the typed contract above these layers. It is neither a second
query optimizer nor permission to collapse native metrics, raw-span durations and
record-page semantics into one executor. A new display should first reuse a
supplied dataset, then add a finite role only when existing composition cannot
express the user task.

## What the recent cadence resolved

397 exposed existing quantitative Bar/Table declarations through the direct
Customize continuation. It preserved the selected measure and exact numeric
binding rather than introducing more arithmetic. Its decoder/API changes closed
ordinary display null/Unicode mismatches. Those are delivered outcomes, not
remaining work.

398 established a reusable mean-only grouped arm with exact duration sum/count
support, captured zero versus unavailable empty duration, common full-population
scales and exact clipped intervals. Source and destination means are independently
checked against physical raw spans. Percentiles remain deliberately outside that
arm; neither a group scalar percentile nor a displayed mean supplies a general
mergeable distribution. See [the contract](../GROUPED_MEAN_DURATION_REUSE.md).

399 should close only the demonstrated Distributions/grouped-renderer import
cycle. Pure formatting and coordinate helpers can move downward; cohort
snapshots, query identity, source witnesses and navigation must stay with their
owners. No speedup is assumed. Do not expand a successful extraction into a global
renderer registry, source client or lifecycle manager.

## Ranked next inputs

### Breadth: preserve a compact population summary in a composed view

The concrete gap is authoring and reuse, not execution. `exploration/source.go`
already assigns `SummaryDataset`; `summaryViewDataset` in
[view_comparison.go](../../exploration/view_comparison.go) derives it from the
whole-population aggregate with explicit measure support. Ordinary
[Distributions](../../ui/distributions.mjs) already passes it to
[`goo-measure-summary`](../../ui/measure-summary.mjs). However,
[`ViewCompositionChild`](../../workbench/view_composition.go) and
[`view-composition.mjs`](../../ui/view-composition.mjs) admit only aggregate,
records, timeline, comparison and facet roles, while the population workspace has
no summary-dataset role consumption.

A user can therefore inspect a useful exact summary but cannot express that same
compact summary as a first-class optional child of a reusable investigation view.
This is a worthwhile general-purpose continuation of the grammar rather than a
new specialized dashboard.

Start with one optional typed summary child over the parent's existing summary
aspect. Let it select only fields supplied by the checked query, reusing the
existing component. Require explicit available/unknown/empty state, actual units
and support; do not invent a fleet percentile, add reductions over displayed rows,
or import captured data into stored configuration. Count and numeric/duration
summary fields may be available without a grouping, so admission should not
accidentally require an unrelated categorical breakdown.

Gate before coding: decide the finite valid role combinations, default visibility
and whether the existing maximum-child bound changes. Preserve one parent query
and source request. Run, Save/reopen, Inspect/Apply, hidden roles and held return
must preserve complete source/destination questions. A row page or grouped mean
must never become the summary denominator.

### Depth: carry nearby gauge-observation context into retained surfaces

The ordinary Measurements gauge-to-records path already exists and has the right
meaning: an exact selected reported timestamp with a clipped nearby window is a
time association, not causality. See
[the existing contract](../GAUGE_OBSERVATION_CONTEXT.md).

The remaining adoption gap is concrete.
[`native-held-record-context.mjs`](../../ui/native-held-record-context.mjs) admits
only counter/histogram kinds. The retained owner in
[`saved-native-record-context.mjs`](../../ui/saved-native-record-context.mjs)
already verifies receipt/question equality and owns pause/resume leases, but does
not thereby admit a gauge. The service adapter in
[`service-dashboard.mjs`](../../ui/service-dashboard.mjs) binds contextual native
navigation only for histograms. Existing gauge observations on those surfaces
cannot use the same nearby-record continuation.

Extend the existing observation planner and retained owners, not a second time
heuristic. A receipt must retain one compatible result, uniquely identified
stream/sample and original fixed bounds. Preserve service/instance, namespace
presence/value, environment policy and exact nanosecond clipping. A real selected
timestamp can supply context even when its magnitude is unavailable; no timestamp
or scope may be manufactured from freshness text. A retained definition alone is
not captured observation evidence.

The first cut should remain service gauges with the existing ±60-second rule;
resource-bound gauges, stocks and broader stream grains need their own semantic
gates. Do not merge saved-receipt and service-current ownership. Test them through
their actual wrappers, including hidden/pending/replaced owners, exact Run and
source-free Back, while keeping native save/capture/source capabilities separate.

### Consolidation: finish user-visible table context with existing owners

If the design review supplies current viewport evidence, a bounded shared-table
readability pass is a good third step after the two capabilities above. Recent
398 narrow acceptance explicitly relies on internal horizontal overflow; that is
valid behavior, but a scrolled value must retain enough visible group/interval
context to be interpretable. Use existing semantic table markup and the shared
component/styles, retaining exact UTC bounds, qualifiers and keyboard access.
Avoid a new data-grid framework, custom query state or a universal column engine.
Architecture does not require postponing this behind more grammar expansion.

The exact surface and acceptance should be selected from the current design
review, not inferred from this source-only review. No performance or accessibility
improvement is claimed here without actual component/browser checks.

## Reliability requirements across those cuts

The strongest new lesson is **validate the enclosing carrier, not only a leaf**.
398 passed its finite declaration vectors but actual stored-view API observations
still found a count-only summary constraint in the enclosing composition schema.
The fix and 14 enclosing-view cases are delivered. Keep that lesson as acceptance
for future roles: actual Go decoding/validation and immutable-store round trips,
actual JavaScript declaration/application consumers, and API input/output
carriers, with positive combinations and rejecting profile/query/null/unknown-field
mutations. A second JSON-schema interpreter or universal validator platform is
not needed.

The 396 ownership lesson remains active. Local presentation publication after
Back is distinct from source-staging leases; both nested return and outer trail
return need composed tests. A terminal callback test alone does not establish
that visible edits will be saved. New summary children should use the existing
owner predicates rather than add another generation manager.

Continue testing actual physical source membership independently of derived
outputs. Keep generated local/Hegel choices separate from real source reads and
Bombadil DOM observations; model replay should reject deliberately changed
semantic facts and unintended acquisition. Retain the useful current fixtures,
frames and counterexamples without treating every transient report as product
state. Keep heavy jobs serial.

## Go and library conclusions

The project already targets Go 1.27 (`go.mod`), with Go 1.27.3 used in recent
validation and image builds. Existing `bytes.Lines` in
[`internal/sourcejson`](../../internal/sourcejson/read.go), `slices`/`maps` for
explicit small copies, strict concrete JSON decoders and `t.Context` in tests are
appropriate modern facilities. No upgrade-only or generics quota is justified.
Keep caller APIs ordinary concrete questions, declarations and results. Use
advanced type machinery only for a measured usability, reliability or performance
benefit; no such need appeared in this review.

Preserve validate-before-detach at `exploration.Source.fetchProfile` and native
question boundaries. JSON round trips are still used for some typed clones;
[`CreateExplorationView`](../../workbench/exploration_views.go) validates before
its detach, while native views use intentional nested copies. Replacing those
paths requires caller-mutation/invalid-input proof and allocation benchmarks.
`maps.Clone` and `slices.Clone` are shallow; a generic `Clone[T]` wrapper would hide
the important ownership policy. No clone rewrite is recommended in this cadence.

[`internal/sourcehttp`](../../internal/sourcehttp/read.go) and
[`internal/sourcejson`](../../internal/sourcejson/read.go) already share bounded
mechanics without stealing family-specific policy. Keep typed SQL parameters,
readonly settings, row/byte/time limits, framing, cancellation and failure mapping
with their source families. The recent work supplies no evidence for pooling,
six-query concurrent fanout, a universal source engine or a replacement JSON
library. Benchmark a representative path before changing resource pressure or
claiming a speedup.

## Deferrals and limits

Grouped P95/fleet quantiles, arbitrary executable widgets/WYSIWYG, warehouse/new
stores, tenancy/alerts and broad unsupported OTel execution remain separate gates.
They should not become prerequisites or an additional queue. A passive facade or
gallery update can accompany a newly admitted component role if its dependency
closure stays display-only; it is not a reason to broaden 399.

This review adds no production change. The planner should reconcile these inputs
with current design/product evidence after 399 acceptance, adopt one finite
breadth → depth → consolidation cadence, and leave the other candidates explicit.
