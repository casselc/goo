# Contextual actions and retained return

Executable: [contextual-actions.spec.js](../../journeys/contextual-actions.spec.js).
This bounded journey uses an isolated Go binary and actual OTLP spans in
ClickHouse. Executed results belong in [validation](../../VALIDATION.md); the
following storyboard specifies the acceptance flow.

| Step | Operator gesture | Expected behavior |
| --- | --- | --- |
| Open service | Continue from a captured fixed inventory scope to one qualified service. | Service overview retains its own exact window and observed widget evidence. |
| Compare | Open the HTTP route widget’s native Widget actions and use Compare captured records with Enter. | The existing population comparison stages locally. The menu closes after activation. |
| Run | Explicitly acquire the selected and preceding fixed populations. | Full source counts agree with the independent source oracle; two source API calls, six SQL reads and one pure comparison. |
| Back | Return to service. | Original widget DOM/evidence remains; its native action menu reveals the actual clicked Compare control and focuses it without another source read. |
| Customize | Open the current HTTP route widget configuration. | Valid asynchronous configuration preparation may finish while its own initiating button is disabled. No telemetry is read. |
| Read records | Explicitly Run the population. | One source API acquisition and four SQL reads; bounded rows do not define whole-population counts. |
| Inspect | Use Enter on the first retained record. | Local selected-record details open; inert, fieldset-disabled, ARIA-disabled and disabled saved handlers refuse. |
| Expand trace | Use the actual inspector's Expand trace action. | Fresh span and linked-log contexts are acquired independently: two source API acquisitions and seven SQL reads. |
| Close trace | Close the native trace dialog. | Focus returns to the actual inspector action without rereading. |
| Row menu | Expand through the row's More menu, then close. | The intentional persistent More summary regains focus; the menu remains closed. |
| Replay withdrawn action | Invoke saved hidden/retired controls. | No new source reads, comparisons or destination mutation. |

The source ledger distinguishes API acquisitions, pure comparison calls and SQL.
Local gestures conserve all three counters. Native disclosure ancestry is revealed
only during explicit retained return; invocation does not make closed controls
eligible. Hidden/inert/disabled owners are never revived for focus. A supplied
fallback remains available when the original action is no longer eligible.

Frames are generated at 1440, 390 and 320 pixels for service return and record
inspector return. The optional `GOO_ACTIONS420_BOMBADIL=1` campaign independently
repeats the current service comparison and actual-button return at 390 pixels.
It runs after the main ledger to keep its shared-proxy SQL outside main gestures.
No publication, authentication, screen-reader certification, exhaustive schedule
coverage or whole-product action audit is claimed.

## Executed acceptance

The actual journey plus 35-second actor passes in 55.9 seconds. All fifteen main
facts are continuous, with zero page errors. The independent source oracle
confirms 320 selected spans versus one preceding-window boundary-control span.
This HTTP route widget has no interval directive; its explicit comparison costs
six SQL reads. The overview’s separate span timeline widget would also acquire
intervals and is outside this journey’s chosen budget.

The actor completed across 697 snapshots with held evidence intact, actual
Compare-button focus restored, and no publication. It runs after the main facts.
Both inspector and intentional row-menu trace returns pass. The isolated runtime
stopped cleanly. All six service/inspector frames were regenerated from the actual
binary; independent reviews and native replay results are recorded in validation.
