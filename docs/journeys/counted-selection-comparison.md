# Compare a counted selection with its rest

Executable: [counted-selection-comparison.spec.js](../../journeys/counted-selection-comparison.spec.js).
Status: implemented with actual desktop/narrow frames accepted in a bounded
source/frame review. Executed browser, actor and model outcomes belong in
[validation](../../VALIDATION.md).
See [the feature contract](../COUNTED_SELECTION_COMPARISONS.md).

An operator has isolated an Activity scope, inspected a field and explicitly
loaded the full-scope counted values. They want to compare a small population
with everything else in that same scope and completed window.

| Step | Interaction | Observable contract |
| --- | --- | --- |
| Select | Choose missing and present-empty values. | Literal lookalikes stay distinct; hidden search selections remain selected. No source read. |
| Compare | Choose Compare with rest. | One fixed comparison is staged; neither loaded counts nor retained rows become its operands. |
| Review | Inspect selected/rest scope and UTC window. | Both preserve the parent predicates, qualified field, environment policy and fixed bounds. Rest is relative to this parent. |
| Run | Explicitly run the staged comparison. | Both source populations are acquired independently; counts, clocks and coverage remain honest. |
| Inspect | Read the paired count and exact facts. | Known zero is numeric zero; unavailable/diagnostic data is not inferred healthy. |
| Return | Choose Return to paused activity. | Held Activity and the actual comparison opener return without rereading. |
| Refuse stale actions | Replay withdrawn or changed-scope handlers. | No new source acquisition or obsolete comparison staging. |

The real OTLP/ClickHouse fixture has 261 matching records and a retained Activity
buffer of 100 observations. Missing and present-empty values select two records;
the rest contains 259, including a literal `"Missing key"` value. The journey
checks each question against raw source records, preserving parent predicates,
qualified grouping, effective environment policy and exact UTC window. The count
comparison is separately acquired; loaded facet counts are never used as operands.

## Actual documentation frames

- [Selected values at 1440px](screenshots/counted-selection-controls-1440.png)
- [Selected values at 390px](screenshots/counted-selection-controls-390.png)
- [Selected values at 320px](screenshots/counted-selection-controls-320.png)
- [Completed comparison at 1440px](screenshots/counted-selection-comparison-1440.png)
- [Completed comparison at 390px](screenshots/counted-selection-comparison-390.png)
- [Completed comparison at 320px](screenshots/counted-selection-comparison-320.png)

The main report is `artifacts/counted-selection425/report.json`, version
`goo.countedSelection425/1`. The continuous ledger distinguishes local selection,
staging, hidden/stale action replay and retained return from explicit Load and
Run. Run acquires both populations independently. The optional 35-second Bombadil
actor runs after the main ledger and independently observes staging and retained
return; its result is not inferred from a clicked button or these frames. Native
verification checks the observed ledger and generated categorical partitions.

The shared controls retain native checkbox labels, exact read disclosures,
search-hidden selection status and a separate Compare action. The completed
comparison uses the existing view and Scope and read facts. Its ordered title
identifies rest on the left and selected values on the right; generic fixed-window
role labels continue describing the operational reads. Final executed outcomes
and review limits are recorded in validation and the
[integration review](../reviews/COUNTED_SELECTION_REVIEW_425.md).
