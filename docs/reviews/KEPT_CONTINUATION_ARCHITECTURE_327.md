# Continue from a kept observation — architecture 327

Status: final source review accepted; generated and native checks recorded in
[VALIDATION](../../VALIDATION.md).

A kept observation can continue into a trace investigation without changing its
frozen receipt. Open trace as view stages the existing exact-trace query and
record-only waterfall declaration; Run explicitly acquires fresh telemetry.
Reported identity does not prove a complete trace or causal relationship.

## Scope and ownership

KeptSaved.traceView clones and validates the receipt and chosen member. Tail keeps
use that occurrence's own fromMs/toMs; kept result-page and saved-capture records
use the original fixed query window. The checked environment policy is retained.
The staged spans query uses only the reported TraceId, page zero and ordinary
records; old predicates, grouping and measure directives are explicitly omitted.
Missing, all-zero or malformed TraceId and invalid member positions refuse.
No kept rows or source guarantees become current trace evidence.

KeptSaved retains a detached payload and a private show/member lease. Fresh
inspector keys renew commands on repeated inspection of the same member; retained
Back preserves the existing inspector rather than rebuilding it. Captures owns
receipt identity and the explicit retained restore. InvestigationTrail owns the
actual destination entry, with an irreversibly releasable current-visit lease;
query/hash equality cannot revive an old return after another visit.

The app stages through openQuestion(view.queries[0],[],{stage:true,view}), suspends
the destination before returning and releases its retained-return mode. Stale
origin/visit controls refuse and normal trail navigation is restored. Hidden
trail Back/Forward handlers recheck visibility and disabled state. The explicit
Activity enterFixed transition clears unrelated ancestry without resuming an old
question, preventing a fabricated paused-Activity parent.

## Reads, cancellation and retained return

Staging performs no telemetry read or store write. The measured fresh destination
mount makes exactly two existing metadata GETs: recipe definitions and retention
configuration. This is not a zero-HTTP navigation claim. A completed retained Back restores the same kept inspector, opener and
scroll with zero IO. Captures suppresses only this return's normal catalog load,
and replaceState avoids the ordinary #kept receipt loader.

Pending-read Back may send the existing exact-query cancellation control POST,
without a replacement source or store reread. It does not claim that all server work or
unrelated application requests cease. Existing source/return owners remain
separate; no universal navigation owner, new executor or custody copy was added.

## Validation

TestHegelKeptTraceContinuation327 passes thirty generated cases in 10.219 seconds,
zero skips. Actual Go-validated/digested tail, result-page and saved-capture
receipts each plan two members: 180 resulting views are admitted by Go view/query
compilation. Coverage includes five environment representations, namespace
presence, omitted log signal, distinct tail windows, trace-only composition,
invalid IDs/positions, foreign-policy rows and detached receipt/plan mutation.

See the final [validation record](../../VALIDATION.md) and
[native journey](../journeys/327-kept-trace-continuation.md)
for actual stage/Run/Back request budgets, cancellation, repeated inspector
commands, stale leases, focus and Activity ancestry checks. Pure generated tests
are not substituted for those browser results.
