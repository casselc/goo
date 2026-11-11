# Roadmap after the 405 checkpoint

This cadence is complete. The [408 checkpoint roadmap](ROADMAP_2026_11_11_CHECKPOINT_408.md)
owns the next execution order; this document retains the accepted 406–408 scope.

The [402 cadence](ROADMAP_2026_11_10_CHECKPOINT_402.md) delivered trace-search
continuation, ungrouped record/summary templates and shared ClickHouse parameter
encoding. [Validation](../VALIDATION.md) records actual results; the
[review reconciliation](reviews/ROADMAP_RECONCILIATION_405.md) records the adopted
architecture and design decisions.

The finite cadence is **406 breadth → 407 depth → 408 consolidation**, then
whole-product review. Keep current ordinary OTel → scope a population → inspect
or compare evidence → preserve/reuse the question and presentation. A manifest is
optional. This prerelease has no legacy-format obligation; current Go is 1.27.3.

## 406 — breadth: captured-record activity rate in the shared grammar

Implemented: [captured-record rates](CAPTURED_RECORD_RATES.md) now have a distinct
typed field and exact count/exposure facts in fixed ordinary summaries and
ungrouped timelines. Existing defaults remain unchanged. See the executable
[journey](../journeys/406-record-rate-storyboard.md) and [validation](../VALIDATION.md).
407 is also implemented below; **408** is the next consolidation step.

User task: “How quickly are these log/span records appearing, and when did that
activity change?”

Service dashboards already compute observed records per fixed-window second.
Generalize the supported measure to ordinary fixed-query summaries and ungrouped
count timelines, using existing counts and explicit elapsed time. Offer a measure
choice, not another default card. Admit a finite typed field/profile and its
support facts before implementing; choose supplied dataset versus a checked
transform explicitly. Do not hide new arithmetic inside a renderer.

Use exact count × 1000 / elapsed milliseconds, with each actual clipped interval
width for timelines. Preserve large integers/rationals and units. Rate is distinct
from percentage/ratio metadata, which currently implies a bounded 0–100% domain.
Known zero, empty input, unreported gaps and unknown coverage remain distinct.
Label captured log/span records per second; duplicate observations, upstream
sampling and completeness remain unknown. Do not call it requests, transactions,
SLI success, ingestion health or application throughput.

Initial scope excludes live-buffer rates, grouped rates and rate comparisons.
No new source SQL is required. Saving/reusing presentation must preserve the full
query, source/basis clocks and explicit Run. Test leaf and enclosing Go/JS/API/store
carriers, short final buckets, exact support and unknown/zero states. Use actual
raw count/window oracles and real desktop/narrow gestures; refresh affected actor
and ownership checks rather than creating a separate source owner.

## 407 — depth: author numeric histograms beside records

Implemented: [authored numeric histograms](AUTHORED_NUMERIC_HISTOGRAMS.md) now
preserve Histogram/Table and completeness requirements beside records, with an
exact numeric/bin-bound template profile. Current-source range planning and held
return remain separate from local display changes. See the executable
[journey](../journeys/407-authored-histogram-storyboard.md) and [validation](../VALIDATION.md).
Proceed to **408**, then the planned whole-product review.

User task: “Keep this numeric distribution and its records as a reusable view,
then apply the presentation to another compatible population.”

Existing numeric histogram query execution, supplied histogram datasets, renderer,
range planner and basic question save/capture are delivered. Extend the authored
composition/customization path with one finite histogram/Table-plus-records arm.
Do not introduce another evaluator, percentile aggregation or source client.

Require exact qualified numeric field, type, scale, literal unit and parse policy;
settle compatible bin-specification and support admission explicitly. Preserve the
destination's full predicates, environment, bounds, page limit and histogram query.
Retain existing page-zero authored/template-root admission; continuation pages
do not gain template eligibility.
Never copy donor bins/query options to make it fit. Underflow/overflow, interval
inclusivity, empty bins, invalid/missing values and unknown coverage stay explicit.
Density/normalization belongs only where its existing support proof is admitted.

Inspect/Apply transfers presentation, never histogram observations, completeness,
source leases or range-selection authority. Bind any subsequent selection to the
current accepted destination read. Preserve explicit Run and held Back. Validate
full declaration/composition/store/API carriers and actual author/save/reopen/reuse
with independent raw bin counts/support and stale/pending control refusals.

## 408 — consolidation: complete passive grammar exports and examples

Expose current ScopeView and MeasureSummary through the explicit public display
entry, alongside existing leaf components. Direct imports already work; this is
caller discoverability and embedding coverage. Add a compact reference showing the
same supplied data through supported display kinds and summary, using native HTML
and detached/pre-upgrade SPA properties.

Keep query scope, requests, source subscriptions, evidence state and publication
outside display components. Verify dependency closure and hostile-global imports,
SSR/upgrade/detached behavior, accessible labels and no fetch/store/source effects.
Keep feature-level imports available for a smaller graph. Add no registry,
framework/bundler migration, executable kinds or generic rendering interpreter.

Finish and review each milestone before advancing. Keep heavy jobs serial,
preserve telemetry/stores and regenerate only affected real documentation frames.
Do not append deferred candidates to the execution queue before the next review.
