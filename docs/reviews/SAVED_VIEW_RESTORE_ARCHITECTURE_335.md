# Saved-view restoration ownership — 335

## Status and scope

Delivered consolidation milestone. The initial source audit below is against
commit `1f14811`; final implementation/Hegel acceptance is recorded at the end.
Actual browser request counts and campaign results are in [VALIDATION.md](../../VALIDATION.md).
No product files, builds or browser sessions were changed by this reviewer.

The task is to open a saved definition once per route visit, keep it staged until
explicit Run, and prevent a departed definition response from replacing the
user's current task. Existing fixed query bounds, typed environment policies,
comparison operand facts, local controls and retained child return must survive
passive history notifications. This is an owner correction, not a new query
planner, cache, renderer or router programme.

## Existing owner and demonstrated source behavior

The saved-view owner is `ui/population-workspace.mjs`, not the ordinary question
link restoration in `ui/exploration.mjs`. Its root mount registers the same
`restore` callback for `hashchange` and `popstate`, and schedules an initial
restore. Each admitted callback calls `load(id)` independently. The load function
immediately clears selection/children, advances `generation`, aborts the current
controller, resets the panel and starts `GET /api/exploration-views/:id`.

This establishes two independent acquisition paths. Only a measured paired-event
history traversal can establish whether both requests actually reach the server.
Even a successfully completed first load remains vulnerable to a later same-visit
notification resetting local configuration, focus or fresh held results.
Therefore an in-flight-only promise join is insufficient if the acceptance task
includes passive events after completion.

`currentRun(g)` already checks generation, active workspace and no child owner.
It is necessary for generic source execution, but insufficient as the sole saved
restore guard. `Workspace.routeFromHash` maps saved-view, ordinary question/step,
and plain Explore hashes to the same Explore workspace. A pending saved-view
response can therefore outlive a change of task while `active()` remains true.
The current restore callback does nothing for a non-saved-view hash; it does not
itself retire that pending restore. A same-ID comparison is also insufficient for
an A→B→A visit if the first response ignores cancellation.

The existing captured-evidence route has a local visit object that joins repeated
same-ID notifications. It is a useful nearby pattern, not a mandate to extract a
universal navigation manager. Saved definitions and immutable evidence still have
different acquisition/adoption semantics.

## Recommended bounded owner change

Keep a root-owner saved-route visit with the requested ID/hash, its own identity
and one metadata acquisition. Repeated same-route notifications join or retain
that visit while pending or completed; they must not call `load` again, abort an
explicit Run, clear a local draft, close a child or move focus.

Retire the visit before aborting when the observed route changes, another task
explicitly stages/replaces the population, the owner suspends/disposes, or a later
intentional visit starts. Check the visit identity and exact route immediately
before both success and failure adoption, in addition to the existing generation
check. Capture the request controller rather than relying on a subsequently
reassigned shared controller. A response that ignores abort must remain unable
to run `onRestore`, replace controls/status, or restore focus after departure.

If queued hash events are included in the task, use their `newURL` to notice an
observed departure even when the current location already returned to the same
hash. Retirement is irreversible; returning to identical bytes starts a new
visit. Avoid a bare saved-ID cache and avoid coalescing all navigation into a
microtask: both can lose the distinction between one paired traversal and a new
intentional visit.

Failure behavior needs an explicit task contract. Passive duplicate events must
not create an automatic retry loop. A later intentional visit, or an existing
explicit retry control, may acquire again. Do not accidentally make a failed
same-route visit permanently unrecoverable. No new retry button is justified if
existing navigation already supports an understandable retry.

No Go runtime feature or advanced generic API is needed for this frontend owner
fix. The 334 review's verified Go 1.27.3 runtime and independently reviewed module
graph maintenance remain separate; this task must not revive blocked heavyweight
example checks in parallel with browser validation.

## Acceptance and negative controls

1. Measure real same-ID history Back/Forward, including delayed first metadata
   response. Record event type, GET start/abort/completion and accepted adoption.
   Report requests per route visit, not aggregate campaign totals.
2. A paired same-visit notification produces one metadata GET. Neither pending
   nor completed repeats reset local controls, focus, held comparison data or a
   child. Opening remains configuration-only with zero telemetry reads.
3. Changing A→B starts one B load. Resolve A last with an abort-insensitive
   response: it must not affect B or its status. Repeat A→plain Explore/question
   and A→another workspace; rejected success and rejected error must stay inert.
4. A→B→A establishes a fresh A visit. The old A response cannot adopt merely
   because its ID/hash matches again. Include queued departure notifications if
   the implementation claims that boundary.
5. Deliberate later re-entry/retry still obtains the definition. Explicit Run
   remains a fresh telemetry read and must not be cancelled by passive paired
   route notifications. Retained Back never reloads the parent definition.
6. Test the actual mounted owner, not just a helper that mirrors the new code.
   Prefer delayed abort-insensitive reads and independent DOM/status/focus facts.
   Run the focused check against baseline source as a negative control where
   practical; the original behavior should fail the duplicate/adoption assertion.
7. Preserve one real UI-saved definition in the native journey. Check its exact
   fixed windows, typed environment and composition on reopen, then explicit Run
   and retained child return. Keep storyboards/maintained frames synchronized;
   compact summaries suffice after raw artifacts have been reviewed.

## Delivery review pending

Architecture accepts a local saved-route visit correction under these bounds.
The final diff, focused actual-owner results, measured native event ledger and
staged/explicit-read/retained-return behavior still require delivery review.

## Measured baseline and generated-owner check preparation

The delivery owner independently reviewed the native baseline ledger
`artifacts/restore335-baseline/baseline.json`: one genuine saved-definition hash
history visit emitted `popstate` and `hashchange`, started two definition GETs and
held both responses, with zero telemetry reads and no page errors. The baseline
wrapper passed in 12.3 seconds. This establishes the duplicate acquisition beyond
source inference; it is not a full successful-render baseline.

The separate actual-owner fixture in `tests/support/saved-view-owner.mjs` mounts
the product population owner with deferred abort-insensitive transport and real
saved-view validation. Its `onRestore` observer deliberately throws after counting
admission, before broad UI rendering. Consequently generated schedules prove
acquisition/adoption refusal and post-callback passive retention, not successful
rendered controls, focus, or held record/comparison custody. Those remain native
journey assertions. The Hegel wrapper and generated fixture live in
`verification/saved_view_restore_property_test.go` and
`verification/fixtures/saved-view-restore-335.mjs`.

Each generated round varies paired notification order/repetition, definition IDs,
ordinary versus queued ABA, and late success/error settlement. An independent
oracle requires eight definition GETs, three caller admissions, four rejected
stale settlements, zero telemetry reads and zero writes per round. It also checks
failure passive joining and suspension retirement. The fixture does not inspect
or reproduce the product's visit-token fields. The first lightweight Node run
(two rounds, seven duplicate notifications, seed 17) passed with 16 GETs, six
caller admissions and eight stale refusals. Heavy Go/Hegel validation awaits the
delivery owner's sequential-check release.

## Host synchronization boundary found during implementation

The real exploration host's `onRestore` calls `openQuestion` to synchronize the
parent question editor. That function currently calls `change`, whose ordinary
edit path invalidates the population owner synchronously. Adding an after-callback
lease guard without changing that host contract would reject the expected saved
view itself. The observer-barrier fixture cannot discover this integration effect.

The accepted bounded relayering is an explicit restore-only `preservePopulation`
option through `openQuestion` to `change`: synchronize the parent editor and retire
its old source actions, but skip the redundant population invalidation for this
notification. All ordinary user edits retain default invalidation. Keep the
population's after-callback guard to reject a genuinely reentrant new stage,
invalidation, route departure or disposal. Do not refresh the request generation
or add a general ignore-invalidation flag merely to get the original callback
past ownership checks. Actual native successful restore is required acceptance.

Review also found that cancelling a pending definition on plain `#explore` could
leave the loading panel visible and `busy` true when no next task stages content.
Route departure must retire the request and release its displayed loading state
when that same lease still owns it. Cleanup must not wipe a newer staged task or
an unrelated source Run. Native acceptance should establish that the destination
is usable, not merely that the old response is rejected.

## Generated schedule result

Released sequential validation on verified Go 1.27.3 passed:

```sh
go test -C verification -p 1 -run '^TestHegelSavedViewRestore335$' -count=1 -v .
```

Twenty generated cases complete in 5.74 seconds (package 5.753 seconds), with no
skips. Each round's exact eight GETs, three callback admissions and four stale
refusals are required by the independently computed oracle. This result retains
the observer-barrier limitation above; it does not assert real rendered success
or browser focus. No browser, server build or concurrent heavy check was run by
this reviewer.

The frozen-source review accepts the restore-only parent synchronization and
the distinct route-visit marker/request lease. Explicit Run is not given the
preservation exception; ordinary question edits continue to invalidate population
state. Wrong saved-definition IDs refuse before caller admission. Retry belongs
to the current failed visit and transfers focus only while the retry task still
owns it. A queued initial restore is blocked after disposal. One final pending
load suspension cleanup finding was sent to the delivery owner before full-suite
freeze; final source acceptance awaits that resolution and native success.

## Final frozen-source acceptance

The final suspension adjustment captures whether the pending lease owns the
loading panel, retires it before abort, and uses existing invalidation to hide
that obsolete panel. The retired route marker still joins passive events, so
suspension is not an automatic retry. It does not apply this cleanup to a
completed definition or unrelated running query. The focused owner checks now
assert hidden/cleared loading state after suspension as well as rejected late
adoption.

Final source review accepts both changed UI owners under the bounded contract:
local visit/request ownership, restore-only host synchronization, continued
before/after callback admission checks, strict returned identity, explicit failed
visit retry, and retained completed visits for passive notifications. No generic
manager, saved-definition cache or telemetry executor was introduced. The
post-freeze Hegel rerun passes all twenty generated cases in 5.86 seconds (package
5.870 seconds), zero skips. `git diff --check` passes. Full frontend/Go, rebuilt
served-asset equality and final native rendered-success acceptance belong to the
delivery owner; this review does not claim those results in advance.
