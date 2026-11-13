# Compare captured activity across unequal windows

Executable: [captured-rate-comparison.spec.js](../../journeys/captured-rate-comparison.spec.js).
The journey uses ordinary OTLP logs, the actual Go application, real ClickHouse
results and an independent source-row/count/window oracle. Validation records
actual executed outcomes; the steps below describe the executable contract.

An operator compares a captured baseline with a selected population. Raw counts
are misleading when one window covers twice as long. The comparison uses exact
captured records per second, with each side retaining its own source count,
window exposure and read clock. This does not estimate unsampled request
throughput, unique work, application health or causal change.

| Step | Operator interaction | Observable contract |
| --- | --- | --- |
| Reopen | Open a saved ordinary population view with a captured baseline and rate comparison. | Staging the view reads no telemetry or baseline body. |
| Acquire | Run population view. | One current-population source POST and one pure comparison follow the explicit action; the baseline body is a stored GET. |
| Compare | Read 120 baseline records over 60 seconds against 240 selected records over 120 seconds. | Both exact rates are 2 records/s and their supplied change is zero, despite different raw counts. The selected record page retains only 100 rows. |
| Inspect | Open scope/read facts and exact value details using the keyboard. | Counts, elapsed milliseconds, independent windows/read clocks and exact rational values remain inspectable without a source read or recomputation. |
| Switch measure | Choose Count, Set baseline capture, then Prepare. Repeat for Captured records/s. | The comparison choice is staged explicitly. The complete source question and count aggregate stay unchanged; preparing is stored GET plus pure compare, without telemetry acquisition. Count change is 120 while rate change is zero. |
| Find a change | Choose a captured baseline with 60 records over 60 seconds and Prepare. | The exact rates are 1 and 2 records/s, with a +1 records/s change, independent of retained rows. |
| Inspect known zero | Choose a known empty 60-second baseline and Prepare. | Zero count gives a valid zero captured rate; the selected rate and exact +2 change remain available. |
| Check uncertainty | Supply a deliberate diagnostic cohort to the pure Go comparer. | The controlled diagnostic response exposes an unavailable/null rate change rather than inventing zero. Displaying and restoring that supplied response performs no source query. This is a test control, not a claim that the real source changed. |
| Follow an operand | Open the baseline evidence, then open its captured question. | Staging retains the whole typed environment policy, predicates, grouping, fixed window and page settings without reading telemetry. |
| Read the operand | Run that staged baseline question. | One explicit source POST reads the original baseline window. The known-empty operand remains zero; the window is not replaced with the selected window. |

The independent fixture includes current and legacy environment labels, conflicting
current/legacy labels, a foreign namespace and a record exactly at the excluded
window endpoint. The oracle applies current-key precedence and half-open UTC
windows against actual ClickHouse rows. It reduces `count × 1000 / elapsedMs`
with integer arithmetic and checks selection-minus-baseline as an exact rational.

Actual desktop/phone frames are captured at 1440, 390 and 320 pixels under
`docs/journeys/screenshots/captured-rates-{1440,390,320}.png`. The journey checks
horizontal containment and equal plotting domains/operand-track widths, so equal
rates have directly comparable geometry. Existing shared summary comparisons and
native disclosures provide the presentation; no new investigation workspace is
introduced. A separate `captured-rates-zero-390.png` frame shows the actual
known-empty baseline alongside its nonempty selected population.

## Source/read accounting

The ignored report is `artifacts/captured-rates418/report.json`. Its fourteen facts
separate source acquisition POSTs from pure comparison POSTs, stored baseline GETs
and unexpected publication. The expected final source count is two: the selected
population and the explicitly opened baseline. Six pure compare calls include
the controlled diagnostic case. API acquisition counts are not a claim about SQL
query counts. The native model binds to these observed counters.

## Optional bounded actor

With `GOO_CAPTURED_RATES_BOMBADIL=1`, a fresh 390px actor opens the same saved view,
performs one explicit population read, waits for the full comparison, opens and
closes read facts, and completes without publication. It checks exact rates,
counts/exposure, 100 unique retained source IDs, a single source acquisition and a
single pure comparison. The parent journey independently checks the actor's
source IDs against the raw ClickHouse oracle. Its separate artifact directory
retains Bombadil's trace/screenshots for replay; it does not alter the main
fourteen-fact read ledger.

## Executed acceptance

The live journey plus 35-second actor passed in 1.1 minutes: fourteen main facts,
two explicit source acquisitions, six pure comparisons and no publication. The
separate actor completed across 514 snapshots with one acquisition/comparison.
Four actual frames have independent visual acceptance. Native Quint replays the
observed ledger; generated Hegel arithmetic and rejecting controls pass without
skips (20.291s package run). See [validation](../../VALIDATION.md).
