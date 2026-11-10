# Whole-product frontend and component review after 396

> Root acceptance closure: 396 actual main/actor, property and required observed
> model checks are accepted; see [validation](../../VALIDATION.md). The review
> findings below retain their original read-only context. The adopted queue is
> the [396 roadmap](../ROADMAP_2026_11_10_CHECKPOINT_396.md).

Date: November 10, 2026. Read-only source review; no new browser, performance or
heavy test runs were performed for this document. The 396 baseline reproduced two
Save defects; the fixed implementation's final runtime acceptance was **pending
when this review was written**. This is input to the planner, not an adopted queue.

## Assessment

Goo has a useful architecture for its intended general-purpose investigation loop:
live logs/spans, a scoped fixed question, compatible visual choices, and reusable
configuration. Keep developing that loop through finite data/display profiles.
There is no evidence here for a framework rewrite, a Plotly runtime migration, or
a universal navigation/session manager.

The important boundary is not HTML versus an SPA. It is passive supplied-data
presentation versus application-owned questions, evidence, acquisition and
publication. Native elements, ordinary HTML controls and ES modules already allow
both embedding styles. `scope-view.mjs`, `view-model.mjs` and `snapshots.mjs` keep
exact values, coverage and source authority separate from plotting coordinates.
The 395 grouped profile extends this model instead of introducing an executor.

394 and 395 also improve progression: an accepted distribution can stage its
configuration in the existing workspace, then Run explicitly. Keep one editor and
one save/reuse path. Do not return to record inspection as a prerequisite for
customization or add another dashboard editor to expose the same declarations.

## Lessons grounded in the current implementation

- The 396 baseline found valid visible edits lost after trail Back/Forward, while
  nested population Back was already correct. The latter explicitly renews
  authored controls; `resumeHeld()` can reuse the same adapter. A standalone
  `suspendActions()` test cannot establish the real return contract.
- The baseline also found a 258-byte Find draft invalid in the browser while Save
  remained enabled and stored the preceding query. Field validity must reach the
  existing owner/save gate; browser `setCustomValidity()` alone is insufficient.
- The pending 396 fix separates `localPresentationCurrent()` from source-stage
  epochs in `population-workspace.mjs`. Source staging retains its generation and
  action-context guards. This distinction should remain local to the actual owner.
- `nativeControlReady()` is a small DOM-only predicate for native event handlers.
  Explicit property/API updates remain usable while detached. Do not turn this
  helper into a source capability or automatically wrap programmatic setters.
- Actual 395 frames exposed default figure margins consuming narrow width and
  inherited uppercase styling changing literal `api` to `API`. DOM text assertions
  do not establish the human-visible identity or usable geometry.

## Ranked opportunities

### 1. Make population summaries authorable through the existing grammar

**User benefit:** keep the few relevant count, latency or status facts beside the
records/timeline, then apply that presentation to another compatible population.
This is a breadth opportunity, not another summary-widget implementation.

`view-composition.mjs` admits aggregate, records, timeline, comparison and facet
roles. It has no summary role. Meanwhile `summaryDataset` and
`goo-measure-summary` already exist; `distributions.mjs` renders population
summaries outside the authored composition. Thus a useful class of existing
presentation is not yet expressed by the same authoring/reuse mechanism.

A bounded pilot could add one optional summary child to an ordinary fixed record
composition, selecting only fields already supplied by its summary snapshot.
Retain the records child and current query owner. Keep count zero, missing support,
unknown values, method/unit compatibility and complete-only requirements explicit.
No new aggregate SQL, evaluator, template engine or native-family adoption is
required by the frontend design. Backend/source feasibility still needs review.

Acceptance should compare selected summary fields with independently checked
source summaries, preserve the destination's complete question, and prove local
field/order changes and configuration operations perform no telemetry read. Check
an actual compact log and span workflow, including empty and unavailable states.

**Cost/risk:** medium; declaration/schema/store/UI adoption together. The risk is
claiming a sibling summary read shares a transaction or completeness guarantee
with records. Do not hide that boundary to simplify the widget.

### 2. Align the passive library entry and gallery with the current data grammar

**Caller benefit:** an embedding developer can discover and load the components
used by current Goo views without importing application owners or following dated
pilot notes through several independent entrypoints.

`ui/display.mjs` exposes a closed set of passive feature namespaces, but has no
explicit entry for `scope-view.mjs`, `measure-summary.mjs` or the grouped timeline
renderer. The current grammar is described later in `UI_COMPONENTS.md`, separately
from the primary entrypoint/catalog. Direct feature imports work; this is an
entry/catalog cohesion opportunity, not a claim that the underlying components
are missing.

Review the actual dependency graph, then provide one documented passive entry for
current grammar elements, either by completing the existing facade or naming a
small separate grammar entry. Keep direct imports available. Reuse the existing
gallery and standalone component checks; do not create a package/build ecosystem
or expose application adapters merely to make an import convenient.

Demonstrate server-rendered attributes, pre-upgrade properties and detached updates
for an ordinary ranked view, grouped timeline and summary. Include a plain copied
snapshot with no source witness: it must remain diagnostic and must not acquire
complete-only or continuation authority. The dependency fence must still exclude
fetches, source owners, stores and application bootstrap.

**Cost/risk:** small to medium. Source-authority helpers and passive display APIs
must remain clearly distinguished. Broadening an import graph is not a measured
performance improvement; measure bundle/parse impact before claiming one.

### 3. Own literal-data styling and native-control conformance in shared fixtures

**User/caller benefit:** a new table/chart works at narrow widths and preserves
literal values without each application adapter learning the same CSS and form
rules again.

`product.css` styles all `th` elements as uppercase with letter spacing. Both
`log-patterns.css` and the 395 grouped rules override this for row data. Several
figure renderers separately reset browser margins. Those are concrete signs that
page typography is reaching into reusable data presentation. The 395 fixes are
correctly scoped; the next consolidation can establish the common contract.

A bounded cut should cover literal `th[scope=row]` values and chart figure spacing
inside the existing passive table/chart family, with standalone and nested fixtures.
Preserve intended column-heading styling and application layout. Avoid a global
CSS reset, broad specificity escalation, or cosmetic rewrites of unrelated pages.

In the same conformance fixtures, exercise native details/summary visibility,
disabled fieldsets, inert/hidden ancestors, focus and short visible validation
errors. Use the 396 DOM-only helper where behavior is actually identical; keep each
owner's acceptance/publication decision in its adapter. Check native event
refusals separately from valid detached property updates.

**Cost/risk:** small, bounded to a named component family. Acceptance needs actual
computed text case and viewport geometry as well as keyboard/DOM checks. Extending
one selector or one readiness predicate across every UI module without caller
review would recreate the coupling this work should remove.

## Performance and structural restraint

This review establishes no new performance defect. The grouped profile is bounded
at eight groups, 240 intervals and 1,920 derived cells; it does not justify
virtualization, incremental caches or a rendering framework by itself. If actual
interaction is slow, measure the specific Find/kind/resize operation at supported
limits before changing it. Preserve exact counts and the all-group scale while
measuring; faster filtering that changes the population is not an optimization.

`population-workspace.mjs` is large, but line count is not a refactor acceptance
criterion. Extract a pure profile/validation/presentation helper only where it
removes a demonstrated duplicate. Do not move source and local-definition
lifetimes into one generic owner because their callback shapes look alike.

## Planning handoff

Summary composition is the strongest frontend breadth candidate. Passive-library
entry/conformance and literal-data CSS ownership are bounded consolidation
candidates. They are alternatives for a finite queue, not three parallel
initiatives. Existing ordinary fixed duration comparison remains a separate depth
candidate: `captured-count-comparison.mjs` currently admits Count and fresh
`duration.mean`; any percentile extension needs its own population/method review
rather than frontend relabeling. The planner should reconcile that option with
product/architecture reviews after 396 acceptance.

Sources: [component catalog](../UI_COMPONENTS.md),
[current capability guide](../CAPABILITY_GUIDE.md),
[current queue](../ROADMAP_2026_11_10_CHECKPOINT_393.md),
[component entry](../../ui/display.mjs),
[composition](../../ui/view-composition.mjs),
[population owner](../../ui/population-workspace.mjs),
[timeline adapter](../../ui/timeline-view.mjs),
[shared DOM readiness](../../ui/dom-visibility.mjs),
[shared scope styles](../../ui/scope-view.css), and
[baseline-to-fixed regression journey](../../journeys/396-held-draft.spec.js).
