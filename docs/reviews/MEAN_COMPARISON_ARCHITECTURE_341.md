# Previous-window mean duration — 341

Status: bounded implementation contract; validation results belong in VALIDATION.md.

Ordinary span-duration exploration may stage a fresh previous-window summary
comparison through the existing query-count-comparison declaration. This reuses
record-aggregate acquisition, `/api/compare`, `ViewComparison`, and the passive
summary renderer. It introduces no aggregation endpoint or query engine.

## Admission and scope

`display.value: duration.mean` requires fresh span queries with
`summary: span-duration` on both sides. Only summary display is admitted; omit
ordering. Captured-baseline mean and per-category mean are deferred. Count facets
may still narrow both whole questions; their counts remain counts, not means.
Both complete fixed query definitions retain typed environment selection, filters,
windows and independent acquisition clocks. Staging is pure. Explicit Run owns
reads and cancellation. Supporting-record navigation retains its selected side's
whole question; Back restores the comparison without replacement reads.

## Authority and presentation

Render `duration.mean` from each summary and retain its checked `count` as the
operand contributing count. The field projection is `["duration.mean"]`; it
does not introduce a duration-count field. DurationSummary carries
uint128 `sumNs` and uint64 `count` as canonical decimal strings. The mean is exactly
`sumNs / count` nanoseconds; ViewComparison carries its exact rational difference,
selection minus baseline. Convert to milliseconds with rational integer arithmetic
(divide by 1,000,000) before formatting. Never coerce sums/counts through JavaScript
Number or average group means/percentiles. The older MeasureComparison
`meanDifferenceNs` is rounded to three decimal places; it is not the exact rational
authority. Expose exact values as supplemental detail while the visible comparison
uses readable milliseconds and each side's contributing span count.

An observed zero with positive contributors is a measurement. A zero-count
`no_data` side has no mean; unknown/absent input does not acquire zero geometry.
A changed-between-reads diagnostic suppresses derived difference and comparable
geometry. Completeness and source provenance remain those of each existing
snapshot; requesting a mean does not upgrade a browser-held claim. Compatible
units and aggregation metadata are required. These are span-record durations,
not trace latency, unique request counts, rates, significance or causal effects.

## Verification boundary

Semantic generated tests exercise unequal contributor weights, unsigned values
beyond 2^53, exact rational subtraction, measured zero, empty population and
source-disagreement diagnostics against an independent integer/rational oracle.
Native journeys must inspect real aggregate responses and demonstrate explicit
Run, supporting-record scope and retained return. Reuse the existing comparison
lifecycle model for observed effects when its event projection is unchanged;
a new model merely restating mean arithmetic would add no lifecycle coverage.
No pooled percentile or fleet-percentile assertion is introduced.

### Observed lifecycle projection

Reuse `verification/model/population_drilldown.qnt` and its existing binding:
`stage` records previous-pair preparation or child entry without a source read;
each settled telemetry acquisition records one `read`; retained return records
`back`; a disclosure or save records `history`. `/api/compare` evaluates supplied
summaries and is not a telemetry read. An explicit two-source Run therefore
contributes two read events, not one combined event. Begin the projected ledger
before the first counted source read (`INITIAL: 0`) and retain observed depth and
cumulative counts. Independent coherence checks still verify exact questions,
clocks and rational values; model conformance alone does not prove those facts.
Mutate source count, Back's source count and coherence to prove rejection. This
checks local observed effects, not complete telemetry, provenance or liveness.

The generated OpenAPI conditional admits the mean only for the fresh span arm,
a span-duration baseline and summary/no-order display. The declaration has no
embedded selected question: parent-query span-duration agreement remains the
existing Go semantic validation responsibility, rather than an invented schema
cross-reference or a duplicate selected question.

## Native-discovered return ownership and next consolidation

The mean journey exercised the existing child-return path and exposed a broader
reliability gap: `openChild` advances the parent generation while retaining its
DOM. `closeChild` renewed some controls but left others bound to the retired
generation. The selection tray and fixed UTC editor could consequently remain
visible yet refuse their next legitimate action. Independent source review found
the same pattern in Customize records, Facet settings and Add breakdown; the
record-kind renewal also had an exclusion when the record mount was present.
These are definition-control owners, not permission to replay a retired source
read or renew a checked old measurement intent.

341's bounded repair renews the specific same-question connected local controls
and remounts the relevant existing editors, retaining checked draft text and
open disclosure state. It does not perform a full render, rerun the query or add
a scope manager. Full rendering alone would not establish renewal: held timeline
and aggregate adapters have their own same-data reuse and suspension contracts.
Source review additionally confirmed that pending saved-display/template
inspection could be stranded on child departure: generation refusal of its late
response did not clear the pending flag. The departure now calls the existing
retireReuseRequest operation, clearing that pending state and retiring its serial
without restarting IO. That repair still needs a delayed-response regression
before the path is described as verified.

Select **342: retained-parent ownership consolidation** before another feature
extension. Inventory controls by their existing authority: local definition,
source acquisition, snapshot-bound measurement intent and asynchronous metadata
visit. Cover child entry through selections, comparison supporting records,
record/value plots and Customize; exercise a second Back and an ABA departure
while old callbacks or requests remain pending. Returned same-question controls
must work, while callbacks retained from a retired owner remain inert. Preserve
query/draft/selection content, disclosure and sensible focus, independent read
receipts and clocks; Back/local edits cause no source reads. Use existing mounted
owner tests, generated schedules and observed lifecycle replay, then run a real
representative native journey. This is a bounded cleanup of current owners, not
a universal session layer or framework migration. The previously noted measured
high-cardinality facet/DOM follow-up remains deferred until this reliability cut.
