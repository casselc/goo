# Preserve and reopen a selected/rest investigation

Executable: [investigation-reuse.spec.js](../../journeys/investigation-reuse.spec.js).
Status: implemented with actual desktop/narrow frames accepted in the bounded
source/frame review. Actual browser, actor and model outcomes are recorded in
[validation](../../VALIDATION.md).
See the [artifact and ownership contract](../INVESTIGATION_CONSOLIDATION.md).

The operator starts with ordinary OTel records and uses existing Activity field
controls. Missing, present-empty and literal values remain distinct. The selected
and rest questions share the original parent predicates and fixed window.

| Step | Operator interaction | Contract |
| --- | --- | --- |
| Observe and inspect | Pause a scoped Activity and inspect a field. | Held rows and original completed window remain identifiable. |
| Choose values | Load full-scope field values, then choose a compound categorical selection. | Explicit Load reads; search and selection stay local. |
| Compare | Choose Compare with rest, then Run. | Staging reads nothing; Run independently reads both exact fixed questions. |
| Preserve | Save the exploration view with a useful title. | Store publication saves both questions and presentation, not telemetry or a pair receipt. |
| Reopen | Open the saved view. | Metadata loads and stages configuration; no telemetry acquisition. |
| Read again | Run the saved definition explicitly. | New independent observations retain their own clocks and scope facts. |
| Return | Return to paused Activity. | The same held rows, selections, search and actual opener return without a source read. |
| Refuse stale controls | Replay withdrawn save, reopen or scope callbacks. | Retained handlers issue no new publication/read; completed old responses cannot adopt a receipt or steal focus. |

The report is `artifacts/investigation-consolidation426/report.json`, version
`goo.investigationConsolidation426/1`. It separates store/metadata operations from
telemetry acquisitions. It deliberately starts one save, closes Preserve before
completion, and checks that the response cannot adopt its receipt in the closed
context. That already-issued explicit store write may finish; closing a panel is
not an undo of publication. A later explicit save provides the definition used
for reopening. A capture is a different action: the capture form names the selected
query and confirmation makes a new independent read. This executable makes no capture POST and claims no immutable pair capture. It
checks that withdrawn Capture selected query callbacks do not open a capture
form or publish.

## Documentation frames

- Preserve saved configuration: [1440px](screenshots/investigation-reuse-preserve-1440.png), [390px](screenshots/investigation-reuse-preserve-390.png), [320px](screenshots/investigation-reuse-preserve-320.png).
- Reopened definition before Run: [1440px](screenshots/investigation-reuse-staged-1440.png), [390px](screenshots/investigation-reuse-staged-390.png), [320px](screenshots/investigation-reuse-staged-320.png).
- Explicitly reread comparison: [1440px](screenshots/investigation-reuse-comparison-1440.png), [390px](screenshots/investigation-reuse-comparison-390.png), [320px](screenshots/investigation-reuse-comparison-320.png).
- [Returned Activity with hidden selections retained, 390px](screenshots/investigation-reuse-returned-390.png).

These are actual app frames. Native controls, selected-only preservation labels,
whole-view configuration meaning, fixed UTC scope, exact values, unknown coverage
and visible focus fit the desktop/narrow hierarchy. The source/frame review does
not infer successful keyboard, source-budget, actor or model checks from images.
The executable separately asserts those outcomes; validation records execution.


## Metadata-only import regression

The existing [portable-view exchange journey](../../journeys/305-portable-view.spec.js)
also reaches the same Saved exploration views library from the empty fixed-question
builder, before a telemetry read. It previews/imports configuration through native
controls, then exercises that definition with existing owners. The shared library
moves into View settings for an active population and returns to its advanced
query-options host when the population is invalidated.

Actual narrow metadata-entry frames: [390px](screenshots/305-import-preview-390.png)
and [320px](screenshots/305-import-preview-320.png). These frames are accepted in
the final bounded review; root records source/store budgets and native model
execution in validation. This entry is distinct from the next Recent/Pinned
saved-reference adoption and introduces no duplicate library or preservation type.
