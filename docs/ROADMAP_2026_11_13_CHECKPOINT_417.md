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

## 420 — consolidate: contextual actions and return focus

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
