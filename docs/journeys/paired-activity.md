# Investigate logs and spans without losing either lane

Executable: `journeys/paired-activity.spec.js`.

An operator receives unfamiliar ordinary OTLP data and wants to watch both
families under one environment and attribute scope.

1. Stage Logs + spans without a source request. Start explicitly; each family
   opens its own ordinary tail. Independently queried ClickHouse IDs establish
   that the two rendered buffers contain the actual exported records.
2. Pause logs, Find a loaded value and inspect one record. Its fixed-question
   action uses the displayed original observation window, rather than assuming
   the newest lane window belongs to that occurrence. Run remains explicit.
3. Return locally. Both buffers, windows, Find, scroll and focus survive with no
   new tail connection.
4. Refuse only a span source read with HTTP403. Logs and their Find stay held;
   Retry reconnects only spans. Hold a real successful source response, cancel
   that span read, then release the actual200: no late data is adopted.
5. Start another environment containing logs but no spans. The empty span lane
   reports unknown coverage independently of the observed logs.
6. Open a saved single view. Resolve the visible discard/keep conflict and verify
   both pair streams pause while the saved configuration stages without a read.
7. Reject All environments and a log Body predicate in paired mode. Keep the
   complete filter editable and make no source request.

The source meter is test-only. It forwards real ClickHouse requests, introduces
one explicit403 refusal and delays one real200. It does not synthesize successful
telemetry or UI results. The product remains one Go server.

## Frames

- [Desktop paired Activity](screenshots/444-paired-activity-1440.png)
- [390 px paired Activity](screenshots/444-paired-activity-390.png)
- [320 px paired Activity](screenshots/444-paired-activity-320.png)

## Executed acceptance

The final journey passes **57.2s**, with **19 observed steps** and **seven main
browser tail connections**. Fourteen steps create no tail connection; local
steps also reject unrelated source/publication POSTs. The source meter records
**nine source attempts** across the main journey and separate actor: three logs,
six spans; eight actual200 responses and one deliberate403 refusal. These are
measured attempts for this run, not a universal polling or completeness claim.

The separate **25-second Bombadil campaign** records **409 snapshots** and reaches
its required seventh phase over actual nonempty lanes. It opens exactly two tail
connections and makes no forbidden POSTs. Family separation, original fixed
scope/window, held rows/Find and focus restoration remain true. Fresh ordinary
OTLP records are emitted before the campaign to avoid testing an expired rolling
window. Actor actions are separate from the main native trace's request counters.

The final native observed-trace gate passes **12.71s**, rejecting five mutated
real traces for source reads on Return, changed retained keys, sibling reads on
inspection, and canceled-late data/generation adoption. The standalone Quint
model runs 50 samples × 30 steps; seven protocol mutations are rejected. The
model's logical ownership epochs are distinct from Tail's private abort counter;
native checks validate actual retained keys, selector, connection, windows and
Find. This is bounded custody validation, not a completeness or causal-order proof.

Receipt: `artifacts/paired-activity444/12b1bd3df522/report.json` and its
`TestQuintPaired444Observed.log`. Owned runtime/source-meter cleanup succeeds.
See [contract](../PAIRED_ACTIVITY.md) and [validation](../../VALIDATION.md).
