# Roadmap after the 417 checkpoint

The completed cycle delivers guided service-version comparisons (415), query-only
exception groups (416), and shared record-intent admission (417). The subsequent
interface refinement keeps configuration out of the way of group inspection.
See [validation](../VALIDATION.md), [the admission contract](RECORD_INTENT_ADMISSION.md)
and the [whole-product flow review](reviews/PRODUCT_FLOW_CHECKPOINT_417.md).

The product remains **current telemetry → isolate a population → inspect/compare →
preserve/reuse**. Existing Activity counted facets, brushing, traces, service
inventories, native measurements, declarative views and captured evidence are
foundations to extend. Historical vendor research does not make them new gaps.

## 418 — breadth: ordinary captured-rate comparisons (delivered)

Delivered contract: [captured-rate comparisons](CAPTURED_RATE_COMPARISONS.md),
[storyboard](journeys/captured-rates.md), and [review](reviews/CAPTURED_RATES_REVIEW_418.md).
419 adds recorded Error context; next is the bounded 420 consolidation.

Compare captured activity fairly across unequal fixed windows. Promote existing
`record.rate` count/exposure facts into ordinary supplied-summary comparisons,
starting with captured baselines where unequal windows are already admitted.
Specialized log-pattern count/rate/share comparisons already support unequal
windows; reuse their arithmetic lessons rather than rebuilding that workflow.

Use each side's exact count and window exposure, rational arithmetic, common
visual scale and readable units. Preserve independent clocks and uncertainty.
Known zero differs from unavailable; captured records/s is not unsampled service
throughput. No new SQL, scheduler, query engine or workspace. Keep fresh-query
same/adjacent-window compatibility unchanged in this slice. Verify equivalent
rates across unequal exposure, zero/no-data cases, independent operand identity,
whole-query staging and desktop/mobile comparison geometry.

## 419 — depth: recorded Error context in version comparisons (delivered)

[Contract](SERVICE_VERSION_COMPARISONS.md), [storyboard](journeys/service-versions.md),
and [review](reviews/RECORDED_ERROR_VERSION_REVIEW_419.md).

Extend the existing guided service-version flow with recorded Error counts and
fractions alongside count/mean duration. Reuse typed span status aggregates and
current comparison displays. Show each numerator/denominator; zero population
makes the fraction unavailable and unset status stays visible.

Preserve qualified service, selected versions, operation grouping, profile,
environment policy, fixed window and held return. Recorded span Error is not an
HTTP failure classifier, inferred success, SLO or deployment-causality claim.
Do not add another version dashboard or discovery mechanism. Require source
count conservation, generated exact fractions, stale-action controls and the
existing real-source version journey extended through held Back.

## 420 — consolidate: contextual actions and return focus (delivered)

[Contract](CONTEXTUAL_ACTIONS.md), [storyboard](journeys/contextual-actions.md),
and [independent review](reviews/CONTEXTUAL_ACTIONS_REVIEW_420.md).

Apply existing native disclosure, `UI.actions` and `dom-visibility.mjs` patterns
across two representative owners with demonstrated drift. Recent regression
drivers assumed actions remained inside record rows or returned to a generic
menu; the actual inspector and clicked opener now own those transitions.

Separate readiness to invoke, readiness to complete a pending operation and
retained local interaction. Closed/hidden/stale actions refuse, valid in-flight
completion stays possible, and Back restores the actual opener without rereading.
Use shared journey gestures for native disclosure ancestry, keyboard/mobile
checks and explicit source budgets. This is a finite adoption/conformance pass,
not a global eligibility registry or whole-app rewrite.

## 421 — breadth: contextual Activity facet entry (delivered)

[Contract](ACTIVITY_FACET_ENTRY.md), [storyboard](journeys/activity-facet-entry.md),
and [independent review](reviews/ACTIVITY_FACET_ENTRY_REVIEW_421.md).

After the bounded 420 acceptance gates, expose existing counted-value exploration
from a field the operator is already inspecting. This is an entry-point adoption,
not a new facet engine or another dashboard.

The current path is Activity **Filters → observed field → Load values for this
scope**. `filter-editor.mjs` supplies the field controls; `stream-entry.mjs`
connects `FilterValues.mount` to the applied selector and latest completed read
window. `/api/explore/counted-values` already returns independent count/group
receipts and checked full-scope counts, with a 1,000-group bound. Row service/
namespace shortcuts and inspector Include/Exclude already stage predicates;
**Group by this field** already stages a fixed grouped question. None of those
existing capabilities should be relisted as missing.

Start with one contextual **Facet by…** affordance in the Activity attribute
inspector. Reveal/select that field in the existing filter/value surface and
focus its explicit **Load values for this scope** action. Field selection and
native disclosure are local; the clearly named load acquires counted values.
Reuse the shared attribute, native action/disclosure and counted-value controls,
including their existing search, scope facts and predicate staging. Avoid a new
panel, parallel loaded-value store, eager query on inspection, or persistent
configuration format for this slice.

Acceptance should follow the stream workflow: inspect an ordinary log/span field,
choose the facet entry, read values once, stage an exact value and return without
rereading. Preserve qualified field namespace, typed environment policy, the
complete applied predicate, missing versus present-empty values and exact count/
group conservation. Counts describe the fixed full scope, never the retained
stream buffer. The source/read window must be visible in UTC: an inspected
buffered record can belong to an older observation window than Activity's latest
completed window, and the entry must not imply those windows are identical.
Refuse dirty/changed/retired scope, closed or stale controls, unsupported fields
and late source responses. Cover keyboard/mobile focus and explicit read budgets
with one actual-source journey and generated scope/identity controls before
expanding to other entry surfaces.

This proposal follows the breadth/depth/consolidation cadence. Choose the next
depth item after this entry improvement is exercised; it does not move deferred
monitor delivery, tenancy, classifier training or warehouse work through their
separate gates.

## 422 — depth: select several counted values (delivered)

[Contract](MULTI_VALUE_FACETS.md), [storyboard](journeys/multi-value-facets.md),
and [independent review](reviews/MULTI_VALUE_FACETS_REVIEW_422.md).

Deepen the counted-facet workflow after 421 with explicit multi-value selection
within one qualified field. Use OR among the selected values and AND with the
retained applied scope; two country choices must never silently become US AND DE.
Reuse the existing counted-values source and shared facet/list controls. No new
query engine, inferred field taxonomy or persisted selection format is needed.

Selection stays local over one immutable counted result. An explicit Stage
selected action builds the checked predicate; Apply scope acquires a new
observation. Keep missing, present-empty and literal lookalikes distinct, enforce
predicate size/depth/value bounds before staging, and label counts as belonging
to the original applied window. Selecting values does not compute a new
population total, refresh counts or grant coverage. Changed/retired read scope
withdraws selection authority. Existing context entry and native return behavior
must remain cohesive across keyboard/mobile interaction.

This follows actual 421 usability and source review: the current individual
counted-value action stages immediately and invalidates the lookup. Merely
allowing two sequential clicks would compose contradictory AND predicates;
selection needs its own explicit local draft and one atomic predicate handoff.
Require actual-source scope conservation, exact generated predicate semantics,
stale-action controls and updated executable storyboards before broadening the
workflow to other data families.

## 423 — next consolidation: Activity render and scope snapshots

Profile field inspection and counted-selection interaction with a full retained
buffer and many attributes. Source review found that render-time availability
can repeatedly call `tail.snapshot()`, which clones retained rows although the
checks need only scope/window metadata. This is a measured-performance hypothesis,
not an established user-visible bottleneck. Establish the baseline before
choosing a refactor, and do not add caching merely to reduce a synthetic metric.

If confirmed, separate small current-scope snapshots from requested row copies
and reuse one render snapshot for availability. Keep fresh invocation/commit
checks and source-generation leases; a frozen render must never grant a stale
click or pending handoff permission. Exercise first/second inspection, unfiltered
scope, filtered multi-selection, and native return in the same shared components.
No new global eligibility registry, persisted format or query engine is needed.
Record performance/copy-count evidence and preserve current source/read budgets.
Use the result to refresh the next breadth/depth roadmap rather than extending
this into a whole-app rewrite.

## Reconciliation and boundaries

Architecture/product and usability reviews converged on this breadth/depth/
consolidation cadence after checking actual implemented capabilities. A direct
Activity “Facet by…” affordance remains a useful future entry refinement, but
full-scope counted values already exist behind the filter editor; do not relist
faceting as missing.

Monitoring delivery, auth/tenancy, arbitrary executable kinds, warehouse lowering
and broader native histogram support retain their documented operational, source
and authorization gates. Go already uses the current 1.27 runtime and standard
library; prefer measured improvements over broad generic rewrites. Keep heavy
checks serial and preserve the telemetry stack, streams, captures and volumes.
