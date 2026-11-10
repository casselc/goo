# Roadmap after the 396 checkpoint

Adopted November 10, 2026 after accepted 396 browser, actor, property and observed
model checks. This replaces the completed [393 checkpoint](ROADMAP_2026_11_10_CHECKPOINT_393.md).
[Validation](../VALIDATION.md) owns executed results; the
[review reconciliation](reviews/ROADMAP_RECONCILIATION_396.md) records decisions
across independent design, frontend, architecture and Go reviews.

One finite cadence is active: **397 breadth → 398 depth → 399 consolidation**,
then whole-product review. Deferred candidates are not additional queues.

## Direction and lessons

Keep **current ordinary OTel → isolate a population → inspect/compare evidence →
preserve or reuse the question and presentation**. A manifest is optional. Existing
Live Keep, filters, comparisons, trace waterfalls, service/resource dashboards,
log patterns, captures and reviewed export are delivered capabilities.

394 connected accepted ordinary count distributions directly to existing display
customization. 395 admitted grouped counts through a shared interval dataset with
explicit category identity, coherent zero versus changed-read unknown, and bounded
materialization. 396 fixed two actual held-return defects: valid visible edits
were not always saved, and invalid Find could publish the previous valid value.
Current local definition publication remains distinct from retired acquisition,
staging and source-bound capture/publication leases.

Extend these existing models rather than add another editor or evaluator. Test
composed producer/application wrappers, not only terminal callbacks. Go, JavaScript
and API omission/null/unknown-field behavior must agree at each touched admission.
Use Go 1.27.3 and current standard-library facilities where useful; this prerelease
has no legacy-format or compatibility obligation. Do not introduce caller-facing
type machinery or unmeasured optimization merely to use a language feature.

## 397 — breadth: customize ordinary quantitative distributions

Delivered through the existing flat display path; see [validation](../VALIDATION.md)
and the [quantitative storyboard](../journeys/397-quantitative-customize-storyboard.md).
398 is also delivered; 399 remains before whole-product review.

User task: “Keep this useful numeric or latency breakdown as a reusable view,
without changing its measure or going through record inspection.”

Extend the 394 direct continuation to ordinary numeric/duration Bar/Table
profiles that the existing declaration already supports. The current shortcut
excludes `numeric` and duration-summary queries; that is a progression/admission
gap, not missing mathematical execution. Reuse the same flat saved-view,
`displays[signal]`, population workspace and explicit Run flow.

Before coding, enumerate the exact supported source-query/active-display pairs.
Preserve the selected measure, qualified numeric binding, literal unit, parse
policy and support semantics. Do not silently substitute count, change units,
coerce invalid values or broaden a selector. Keep status ratios, numeric
histograms, grouped-time and other specialized profiles outside this cut.

Acceptance:

- Actual ordinary numeric log/span and span-duration workflows traverse direct
  Customize → staged exact configuration → explicit Run → local Bar/Table →
  save/reopen/reuse. Preserve the full predicate tree, environment policy,
  bounds, grouping, measure and admitted paging, including nondefault limits.
- Independent source oracles check exact aggregates and valid/invalid support.
  Zero, unavailable, missing and present-empty values remain distinct. Shares
  and units cannot borrow authority from a different measure.
- Source-free staging, local changes and held Back do not read telemetry;
  current configuration publication is counted separately. Real wrapper tests
  cover stale, hidden/inert, pending-replacement and repeated-return operations.
- Execute finite shared Go/JS/API vectors through actual consumers: omission,
  null, malformed text, unknown fields and supported combinations, plus rejecting
  mutations. This is feature acceptance, not a universal schema platform.
- Refresh genuine desktop/narrow frames, storyboard, relevant Bombadil actions
  and observed model/property checks. Unsupported inputs have explicit outcomes.

## 398 — depth: reusable grouped mean duration

Delivered through [exact raw-span support](GROUPED_MEAN_DURATION_REUSE.md) and
the existing grouped displays. [Validation](../VALIDATION.md) records actual checks.

User task: “Which route or service group became slower in this interval? Reuse
that duration-over-time view on another compatible span population.”

Existing grouped execution already supplies duration information. Extend 395's
count-only shared data/declaration arm to **raw-span mean duration only**, using
one qualified grouping and existing fixed intervals, source limits and owners.
No new query engine or renderer framework is needed.

The prerequisite is a finite raw support gate: each admitted cell must retain
exact duration sum and contributing count, with independently verified membership
and support. If current source data cannot attest these fields, revise the cut
before implementing display admission. Empty cells have no duration, never zero
latency. Changed/missing cells retain explicit unavailable state; counts remain
counts. No invented all-zero category, averaged scalar percentile, stacked latency
total, fleet quantile or numeric-transform platform is admitted.

Use explicit duration units and a common duration scale across the full admitted
population. Find/order must not redefine the scale population or denominator.
Preserve category identity, exact intervals and any shorter final interval.
Retain the existing eight-group/1,000-observed-cell and derived-display bounds
unless separate evidence justifies a change.

Acceptance: raw sum/count and exact rational-mean oracles; missing/empty/Unicode
groups; known zero count, unknown/invalid support and changed-read refusals;
Go/JS/API admission vectors; exact destination compatibility and whole-question
preservation; no-read Inspect/Apply/local switches; explicit Run and held return;
actual grouped timeline/Table/eligible Heatmap frames with truthful units and
unknown states. Count and duration displays must remain visibly distinguishable.

P95 and other percentiles are deferred. Their existence in a specialized renderer
does not establish the bounded raw-cell support needed for this shared arm.

## 399 — consolidation: extract pure grouped display helpers

Caller task: “Use grouped supplied-data presentation without importing its
application/source owner.”

There is a concrete import cycle: `distributions.mjs` imports
`grouped-time-series.mjs`, whose legacy renderer imports formatting/coordinate
helpers back from Distributions. Extract only the pure formatting/width/coordinate
helpers needed to remove that cycle. Keep cohort snapshots, source capability,
requests, navigation and publication in their existing owners. No registry,
universal session manager, source client or framework migration belongs here.

Acceptance: dependency checks prove the cycle is removed and the helper module
has no source/store/application effects; before/after exact output and boundary
vectors preserve formatting, null/zero, units, full-population scales and literal
labels; relevant actual distribution/grouped journeys and held-return regressions
remain valid. Touched component fixtures may check SSR, detached/nested SPA,
narrow geometry and inherited CSS without a broad stylesheet redesign.

No performance benefit is assumed. Profile/benchmark a real path before changing
allocation, caching, pooling, parallel reads or source budgets. A pure structural
closure is sufficient when behavior is unchanged.

## Reconciled deferred candidates

| Candidate | Decision |
| --- | --- |
| Optional typed summary composition child | Useful medium-sized grammar opportunity, deferred; summary datasets/components already render. Do not bundle a new role/schema into 397. |
| Saved/service gauge nearby-record context | Valid coverage gap with retained sample/custody admission, deferred; ordinary gauge context already ships. |
| Grouped P95 / ordinary percentile comparison reuse | Deferred raw-cell/independent-population support gates, not part of mean-only 398. |
| Narrow table overflow/accessibility | Evidence-backed refinement; include only touched component conformance in 399, otherwise defer rather than add a fourth milestone. |
| Passive display facade/gallery expansion | Useful discoverability candidate, deferred; removing the actual cycle is the bounded structural task. |
| Ownership-copy/clone audit | Preserve validate-before-clone; require caller-mutation regressions and allocation benchmarks before replacing correct cloning. No generic JSON clone helper. |
| Contextual task discovery / kept-occurrence review | Deferred; captured-record review/vocabularies/export and existing Live Keep are delivered. |
| Exponential histograms, arbitrary executable kinds/WYSIWYG, tenancy, alerts, broad DB/broker health, warehouse/new stores | Separate source/design/platform gates, not prerequisites or parallel queues. |

Keep heavy jobs serial and retain useful current documentation/acceptance artifacts,
preserving stores. Reassess the whole product after 399 or sooner if actual use or
398's feasibility gate invalidates the order. Do not automatically append deferrals.
