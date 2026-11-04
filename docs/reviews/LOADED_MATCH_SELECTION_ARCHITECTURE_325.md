# Atomic selection of loaded Find matches — architecture 325

Status: source review accepted; generated atomic/state checks pass.

This depth round completes the existing local Find→select→Keep task. Selecting
matches changes only the local exact-occurrence selection. Keep remains the sole
publication action; no telemetry reread, synthetic membership proof, new backend
route or selection store is needed.

## Existing custody and atomic union

TailBuffer bounds loaded rows independently of TailSelection. The latter already
holds detached full occurrences and original server membership references, so a
selected occurrence can be hidden or evicted without losing its Keep intent.
Batch selection must preserve that distinction and every old selected clone.

Add a small pure preflight/atomic commit operation to existing TailSelection.
Construct a detached candidate union before assigning any class state. Start with
all existing selected occurrences, add matching loaded candidates by occurrence
key and retain insertion order. Never replace an old selected occurrence with a
new representation of the same key. Identical duplicates are idempotent; a same-key
conflict in record, original proof, connection or other occurrence metadata refuses.
All union members must have valid unexpired original preservation membership and
one exact selector/connection. Check the final distinct-key limit of twenty.
Validate old selected proofs too: an expired hidden selection does not become
valid by adding fresh rows.

Any invalid proof, expiration, mixed connection/scope, conflict, over-capacity or
clone failure must leave items, connection and selector unchanged. Never truncate,
deselect hidden/evicted members, refresh references or partially select a valid
prefix. A UI availability check may reuse the same pure planner, but click must
recheck before one commit. Avoid a second eligibility implementation with drift.

The selected UI cut is contextual to nonempty valid Find, avoiding a permanent
all-loaded action alongside ordinary individual Select. Zero matching candidates
have no command, and the pure batch operation refuses an empty candidate list.
Report the matching candidate count and union outcome without implying query
totals or exhaustive data. This is a display/interaction choice, not a new custody
or backend capability.

## Existing host owns the command

Enable the operation only for a paused current tail. Bind its command to current
owner/generation, selector/connection, matching occurrence keys, raw Find and chosen
field layout revision, and selection generation. Recompute current candidates and
preflight at invocation. A stale visible or reinserted button must not select a
new batch merely because its text happens to match.

Block while Keep is pending/retrying/committed or annotation preservation is
pending; the original frozen publication request must not be modified. Batch
selection clears no old Keep/annotation request by surprise. An existing selection
must belong to the same connection/scope and union unchanged. Keep/expiry/hidden
selection UI remains driven by existing TailSelection and current generation.

Selecting matches performs no pause/resume/reconnect, query edit, source request,
Keep write or annotation write. Pausing remains explicit. Preserve Find text,
loaded rows, keyed DOM, row inspection, scroll and filter state. Focus can stay on
the batch command or move to existing selection controls only as an intentional
local action; do not revive hidden row action authority.

## Verification

Generated state/negative checks should compare the atomic union with an independent
reference across hidden/evicted selections, duplicate keys, append/Find changes,
expiry and exact cap boundaries. Snapshot every selection field before a refused
operation and assert byte-equivalent state afterward. Verify original selected
clones/refs and input ownership, unchanged Keep intent after failed operations,
valid no-label Keep intent after successful union, and duplicate content conflicts.
DOM/native checks cover paused-only readiness, pending write protection, stale
command leases after Find/layout/batch/scope/selection changes, exact count copy
and source-free selection followed by one explicit existing Keep request. No new
model framework or backend mutation is justified.

## Selected API and generated lane

The implementation seam is `TailSelection.planMany(items,selector,now)` returning
an entirely detached candidate Map/selector plus added/total counts, and
`addMany` with the same arguments. Commit replans from original current inputs and
only then replaces class state; old selected occurrence objects and selector
identity remain unchanged on a successful union. Mutating a returned preflight
must neither alter class state nor grant commit authority.

The generated lane in `verification/loaded_match_selection_property_test.go` uses
30 bounded command sequences with an independent union oracle. It checks duplicate
idempotence/conflict, exact cap boundaries, full class identity on refusal, preserved
Keep references, old hidden/evicted members, fresh candidates against expired old
proofs, mixed selector/connection, malformed references and clone failures. It also
checks detached preflight/input ownership and exact added/total results. No product
source, backend fixture, browser or new model framework belongs to this lane.
Execution results are recorded in the final source review below.


## Final source review

The landed planner validates the applied selector/predicate, proof/key/connection/
window shape and expiration for existing and candidate occurrences, then enforces
twenty distinct members. Cached connection ownership is explicitly checked for a
nonempty selection; a null owner cannot be silently repaired. Duplicate identity
includes original record, readWindow, membership, connection, observedAt, limited
and late metadata. Derived display byte size is intentionally excluded. Commit
replans before replacing class state and preserves existing selected objects and
selector identity. Individual addition now uses the same admission path; removal
retains its explicit existing behavior.

The UI offers the action only for nonempty valid Find. Its paused-only command
is tied to current raw Find/layout, loaded/matching keys, selection/source
generations, selector/connection, phase and readiness. It rechecks at click and
then invokes the same planner/atomic commit. Pending/retry/committed Keep and
annotation preservation block the operation; an already-selected union offers no
extra mutation. Successful selection only updates existing local selection state
and counters. It calls neither pause/source nor Keep/publication.

`TestHegelLoadedMatchSelectionAtomic325` and the retained
`TestHegelHeldFindState324` regression pass together in 6.056 seconds: 30 generated
cases each, zero skips. The old fixture changes only zero-based synthetic keys to
the server's one-based occurrence ordinals; existing Find/state assertions remain.
The new lane checks exact added/total counts, original Keep references, fully
detached preflight/input values, original object identity on success, unchanged
Map/selector/member identities on every refusal, cap boundaries, conflicts and
clone errors, expired hidden selections, cached null ownership and invalid signal
types. Browser command/Keep/source budgets remain separate in VALIDATION.

A nonblocking next consolidation measurement is repeated availability preflight:
refreshBatch currently deep-clones the union before signature reuse on each expiry
refresh. Selected occurrences can outlive the current buffer. Measure actual
readiness work before adding any cache; click-time independent revalidation must
remain authoritative. This is not a blocker or a new ownership framework.
