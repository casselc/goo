# Captured-rate comparison review — 418

Status: source contract and adapter review accepted. Final browser execution and
independent visual acceptance passed; generated/native results are recorded in validation. This review is
independent of the captured comparison adapters and Go changes; the reviewer
implemented the passive snapshot/comparison display extension.

## Source and comparison contract

The new summary-rate profile is the existing ungrouped record-rate profile plus
the already admitted ordinary single-group captured-count profile. Grouping does
not turn the result into per-category rates: the rate uses the complete captured
population count and its own fixed window. Numeric/status/edge/special grouping
restrictions remain in their respective profiles. Timeline rates remain ungrouped.

Each side retains its own exact count, positive exposure, canonical rational,
window, source clock, dataset/scope identity and quality metadata. A rate equals
`count × 1000 / elapsed milliseconds`. Comparison changes are exact selection
minus baseline; they are not relative count changes, unsampled throughput,
causality or significance. Known zero is a derived zero rate. Diagnostic and
incompatible comparisons withhold changes while retaining inspectable operands.
Imported descriptive data cannot grant source completeness.

Fresh-query declarations continue refusing rate selection and retain their
same/adjacent equal-duration window rules. Captured-rate declarations require an
exact stored baseline and summary display, with no group ordering. Existing
capture compatibility still preserves signal, environment policy, predicates,
grouping and measure directives; only windows and documented delivery paging
may differ. This milestone adds no SQL or background source operation.

## Interface and ownership

The captured comparison editor adds a measure choice but does not automatically
read or publish. Set configures the chosen baseline/measure; Prepare remains
explicit. Retiring a previous choice withdraws old comparison data/actions.
The shared supplied-data comparison renders a common scale, readable rate units,
compact numerator/exposure support and exact details. Count/rate changes remain
distinct. Rates remain opt-in rather than automatically expanding default cards.

Rate mode hides unsupported group controls and group actions. Shared context
retains independent clocks, exact windows and unknown upstream coverage. The
parent removed the unequal-window raw-count warning only for rate mode; it still
applies to raw count comparisons. A minor initial-label refinement was suggested
so the pre-Prepare heading also says records/s when that measure is selected.

## Acceptance limits

Source inspection covered the profile helper, scoped rate validation, declaration
admission, editor staging/retirement and shared comparison composition. It is not
a whole-product accessibility audit or vendor parity assessment. Actual desktop/
mobile frames, real-source source budgets, state replay and generated arithmetic
checks belong in the milestone's validation record before final acceptance.

## Empty-operand chrome remediation

The follow-up zero-baseline frame exposed a blocking presentation error: the
comparison-wide notice and chart said No data even though the selected side
contained records and a valid rate. The shared presentation projection now
scopes the empty-record notice to the actual operand label and uses observed
comparison chrome when the other population is nonempty. Both empty populations
retain no-data chrome; diagnostic comparisons retain unknown chrome and do not
turn uncertain counts into an empty-record assertion. Raw comparison state and
all operand values remain unchanged. Focused tests cover the shared notice DOM,
zero/positive rate pair, both-zero pair and diagnostic projection. Actual rebuilt
frames and execution acceptance remain with the parent.

## Independent zero-state follow-up

A second reviewer of the production changes accepted the pure shared
`comparisonPresentation` helper: it preserves raw operand/comparison states and
exact values, identifies an empty side, retains no-data chrome when both sides
are empty, and leaves diagnostic/incompatible presentation explicit. Both callers
validate supplied comparisons before projection. This follow-up found no blocker;
actual final browser and model results are recorded in validation.
