# Journey 180: declare an exact decimal quantity

A developer receives ordinary spans or logs from a service that reports a charge as a string attribute. There is no Goo manifest. They want to distinguish missing or malformed telemetry from valid amounts, inspect a range, and preserve the question for teammates.

The fixture emits 400 real OTel records per signal, including exact scale-two fractions, negative amounts, zero, integer-coefficient extrema and coefficients beyond JavaScript's safe-integer range. Attribute and resource declarations remain separate. The declared unit is USD; Goo does not infer or convert it.

## 1. Declare precision before reading

Open the fixed question, choose **Fixed-scale decimal**, enter scale **2** and unit **USD**, then author a typed predicate such as `charge.amount >= 1.20`. A missing scale, scale `2.0`, `1.200`, or negative zero `-0.00` blocks Run. Editing and repairing the declaration never reads telemetry.

Running the predicate uses the full fixed population. Switching to a resource attribute preserves the explicit scale and unit rather than treating the two namespaces as interchangeable. Valid, missing, empty and invalid predicates remain separate operations.

## 2. See coverage and select a human range

Run the complete question. It contains 400 records and displays a 100-record inspection page, while source aggregates report **160 valid, 10 missing, 10 empty and 220 invalid** values. Min/max remain exact even when their internal coefficients exceed 2^53.

Use the keyboard to select **[1.28, 5.12) USD**. The chart labels human amounts; its immutable bin keys and bounds remain coefficient metadata. Selection itself performs no request. **Run selected population** explicitly plans and reads the selected 30 records, preserving the original time window. The complementary rest still includes malformed and absent values.

Actual frames: [spans](../docs/journeys/screenshots/180a-spans-decimal-range.png), [logs](../docs/journeys/screenshots/180a-logs-decimal-range.png).

## 3. Compare the range with the rest

Return to the retained parent question and reselect the same range without a source read. **Stage numeric comparison with rest** is a pure planning operation. Run the staged view explicitly to read both populations.

The shared signed display uses human units and exact rational tooltips, with separate coverage. The selected mean is exactly 983/300 USD; its displayed value may round for readability. Finite axis endpoints use decimal text. A service present on only one side has unavailable arithmetic rather than an invented zero.

Actual compact mobile frames: [spans](../docs/journeys/screenshots/180b-spans-decimal-comparison.png), [logs](../docs/journeys/screenshots/180b-logs-decimal-comparison.png). The acceptance test verifies that all four metric rows, both values and complete range captions fit without clipping or sticky-header overlap.

## 4. Preserve evidence and reuse intent

Capture the complete fixed question through the named capture action. It rereads the source; the capture records its own custody and source facts. Reloading that capture reads stored evidence only.

Build and save a reusable YAML question containing `type: decimal64` and `scale: 2`. Reuse runs a fresh window explicitly; it does not mutate the saved capture. The source ledger verifies the same binding, coverage, exact summaries and bins.

## 5. Check boundaries rather than rounding them away

Real API supplements cover scale-one and scale-nine attributes, resource values, unchanged signed-integer behavior, and the final histogram bin. The max+1 coefficient is geometry only and is never emitted as an invalid predicate threshold. Missing/zero/excess scale and incompatible comparison declarations are refused.

## Verification status

- Both real-source main journeys passed: spans 53.2 seconds and logs approximately 1.2 minutes, 2.2 minutes together. Independent BigInt browser oracles verify the full source facts. Final fixtures are `decimal-numerics-6cf7ac4c11` (spans) and `decimal-numerics-630ae6e1cc` (logs) under `artifacts/decimal-numerics/`; each includes full stored capture bytes and `capture-custody.json`.
- The dedicated Bombadil campaign passed for 55 seconds with all retained-range and comparison phases completed. It observed four explicit source reads, two pure planners and one comparison protocol. Its receipt is `artifacts/decimal-numerics/decimal-numerics-bb3c33b889/bombadil/acceptance-summary.json`.
- Passive SSR/predefinition and detached SPA component acceptance passed in 4.9 seconds with zero API requests. It verifies exact scale-nine values, zero-anchored human axes, unchanged all-record shares, local retained range gestures and snapshot cloning. [The synthetic component frame](../docs/journeys/screenshots/180c-passive-scaled-decimal-components.png) is explicitly separate from real telemetry acceptance.
- Independent native/model checks passed for final logs (33.20 seconds) and spans (36.68 seconds), each rejecting 38 changed negative controls: 76 total. The oracle checks raw-value/page uniqueness, all source coverage/summaries/bins, human predicate and display rationals, exact request order and complete capture custody.
- The first spans reuse attempt hit a recorded transient ClickHouse pre-execution timeout. The original failure log and backend diagnostic report remain separate from the successful final mains; original raw response-byte custody was not retained.
