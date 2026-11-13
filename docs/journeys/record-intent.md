# Preserve a record question across entry points

Executable: [record-intent.spec.js](../../journeys/record-intent.spec.js).
This consolidation journey uses the actual served JavaScript modules, browser
history storage, the fixed-question route owner, ordinary OTLP logs and real
ClickHouse results. The property/differential checks in `verification/recordintent`
cover the broader Go/JavaScript admission matrix.

An operator shares or recalls a scoped question and expects the same population.
A malformed link must not turn into a weaker query, erase useful held records or
silently execute telemetry.

| Step | Interaction | Observable check |
| --- | --- | --- |
| Define | Validate ordinary log questions with a 128-byte UTF-8 field, 1024-byte value, 32 predicate nodes, depth four, or a typed numeric threshold. | All use the same semantic admission contract; caller objects remain unchanged. |
| Share | Encode and read each question link. | The exact environment policy, window, predicates, paging and view survive. Links carry configuration, not evidence. |
| Recall | Remember, pin, opt into persistence, and reconstruct the browser journal. | The same four configurations return without telemetry reads. The journey uses a separate storage key and removes it afterward. |
| Refuse | Try extra nested members, mixed predicate arms, oversized UTF-8 fields/values, ill-formed Unicode, excessive nodes/depth, invalid pages, log duration summaries and duplicate grouping. | Admission, link creation, link reading and history all refuse the supplied intent. Neither input objects nor remembered entries change. |
| Stage | Open the valid shared question in the real application. | A fixed question is staged with no rows borrowed and no source request. |
| Execute | Explicitly Run. | The actual source returns 140 matching records and a separate 100-row inspection page. Returned row membership is checked against raw source records. |
| Protect | Open each malformed link while those rows are held. | A visible refusal appears; the same records remain, with no additional source request and no adopted malformed question. |
| Recover | Reopen the valid link. | The exact question stages again without execution. Previous results are withdrawn by this explicit valid staging action. |

The meaningful documentation frame is
`docs/journeys/screenshots/record-intent-refusal-390.png`: a narrow-screen refusal
beside the still-held records. The executable checks horizontal containment.
No new visualization or control family is introduced by this consolidation.

Execution status and actual command outcomes are recorded in
[validation](../../VALIDATION.md). This document describes the executable
contract; it does not claim an unexecuted browser pass.

## Optional exploratory driver

With `GOO_RECORD_INTENT_BOMBADIL=1`, a separate fresh browser runs the same live
bank for a bounded 35-second campaign using
[record-intent417.ts](../../verification/bombadil/record-intent417.ts). It opens a
valid question, explicitly Runs once, checks the 140-record count and 100 unique
source/DOM record identities, then opens a malformed link. The held rows and
applied state must survive with the same single source request. Restaging the
valid question withdraws those rows without another request. The actor checks
narrow containment, no uncaught errors and no publication; completion is checked
from admitted snapshots, not merely successful process exit.

This optional campaign adds a separate `bombadil` result to the ignored report;
it does not alter the main journey's five-fact native read ledger. Actual campaign
outcomes remain recorded in validation after execution.

The accepted run completes in 45.5 seconds including the 35-second actor. The
main journey records five state/read facts and one explicit POST. All 11 invalid
links retain the same 100 row contents after the 140-record result; the primary
question omits offset to check Go zero-default adoption. Frontend, Go, generated
and native-model outcomes are recorded in the validation link above.
