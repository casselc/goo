# Selection availability work — architecture 326

Status: frozen source review accepted; generated regressions pass.

## Measured scope

325's atomic selection planner remains authoritative and unchanged. The actual
browser baseline uses 100 loaded rows (1,068,385 bytes) and twenty hidden selected
observations (213,761 bytes), with a refused 21-member union. Three full refresh
ticks took 7.9/24.5/8.2 ms and made 124 structured clones each. Isolated planner
work took 1.3/3.7/1.9 ms and 23 clones. These are local measured observations, not
whole-product latency claims.

The bounded correction snapshots RecordColumns.layout once per applyFind and
batchState, reusing its selected fields through that synchronous projection and
signature. This removes repeated layout cloning inside the 100-row loop. Remove
the unused plan property from returned batch state as well, so button closures
cannot retain a detached record Map that they never use.

Do not introduce a readiness cache. The measured planner share does not justify
new expiry/ABA assumptions or a second noncloning admission path. Full union proof,
conflict, capacity, scope, connection and expiry validation remains on every
readiness check and click, and addMany independently replans before atomic commit.
No publication/source owner, pure API, cached authority or compatibility layer
changes. Reusing one local layout snapshot is not retaining state between calls.

## Review and validation

Inspect the exact source edit for one layout snapshot per function, identical
chosen-field projection/signature, discarded local plan output and unchanged
currentness/refusal/commit/focus behavior. Compare equivalent browser inputs and
report clone reductions separately from variable timings. Existing 324 Find state
and 325 atomic-selection generated properties cover semantic regression; focused
mount tests and source/write counters cover actual readiness behavior. A new model
or property API is unnecessary for this local structural change.

## Next general-purpose outcome

Continue an investigation from a kept observation. `ui/kept-saved.mjs` currently
offers only local Inspect; each stored member already has checked policy, original
window and reported trace identity. An explicit source-free Open trace as view can
reuse RecordResults.traceView and the existing staged question/view owner, then
require Run for fresh telemetry. Tail members use their own observation windows;
kept result-page members use the original fixed query. Keep frozen provenance
separate and never claim retained rows are a current or complete trace.

Captures.mount already receives app-owned openRecordQuestion. Extend that explicit
handoff with a view option/callback, calling openQuestion(view.queries[0],[],
{stage:true,view}); do not dispatch an event whose default path automatically runs.
Guard kept ID/show generation/member position/visibility and preserve invalid-ID
refusal. Any claim of local zero-read Back needs a retained restore seam: current
#kept hash handling otherwise rereads the stored receipt, even without telemetry.
This advances Keep→reopen→continue without another selection microtask or store.


Final frozen source review accepts the exact local diff: applyFind takes one
field-layout snapshot before its loop; batchState takes one layout snapshot for
matching and signature. Both remain synchronous and read-only. PlanMany still runs
for readiness, and click plus addMany retain independent full validation. Removing
the private returned plan eliminates unused detached-map retention without changing
candidate references, action leases or selection contents. Existing generated
regressions pass; no new pure API or semantic property file was added.


The existing TestHegelHeldFindState324 and
TestHegelLoadedMatchSelectionAtomic325 regressions pass together in 34.595 seconds:
30 generated cases each, sixty total, zero skips. Pure matching, loaded/hidden
selection invariants, exact proofs, detached planning, cap/expiry/conflict refusal
and atomic commit remain unchanged. Actual browser after-measurements and source/
publication budgets are recorded separately in VALIDATION.
