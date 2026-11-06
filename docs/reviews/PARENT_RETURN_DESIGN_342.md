# Retained parent return — design review 342

Status: implemented source and actual native-frame review accepted. This
reliability consolidation supersedes immediate large-facet performance tuning.

A user opens a narrower population, inspects it and returns to continue the same
investigation. Back should restore the place they left, including reachable unfinished
configuration, while providing working current controls. Correct retained numbers
with enabled-looking controls that silently refuse input are not a usable return.
No new control, explanation panel or navigation mode is needed.

## Source review

`ui/population-workspace.mjs` advances the generation and aborts requests when
opening a child. `closeChild` retains observations, reveals the parent and renews
several controls through individual remount/lease paths. Facet and breakdown
editors retain their drafts in `sourceDrafts`; other definition controls capture
query/generation owners. These definition lifetimes must not be confused with
requests that were retired on departure. Route-level `resumeHeld` is a distinct
path and must not accidentally grant a departed source response authority.

The current facet/breakdown remounts use `previous.open || draft`. When an actual
retained disclosure exists, its explicit closed state should win: a user can
intentionally close a disclosure containing an unfinished draft. A first mount
may reveal a recovered draft, but Back should not override the retained gesture.

## Return acceptance

- Keep both complete fixed questions, checked policies, record selections,
  comparison data and original independent receipt/read clocks. Back is not Run.
- Preserve existing clean-action navigation guards. Unapplied facet/window
  definition edits intentionally block continuation; a blocked click performs no
  IO, and the user explicitly applies or discards the draft before child entry.
  Do not weaken that guard to make a retained-draft scenario reachable.
- Across an actual permitted child/Back trip, retain reachable unsaved view title,
  local Find, metadata search, chosen display/profile, toggles and disclosures.
  Do not silently apply or discard those local states to renew ownership. Raw text
  and caret restoration by remount helpers can be checked separately without
  claiming that invalid definition drafts traversed guarded navigation.
- Preserve open/closed disclosure state and the local scroll context. An unchanged
  control should feel continuous; if replaced, its accessible name and value stay
  stable. Keep focus/caret for local replacement when applicable.
- Return explicit Back focus to a current connected, visible, unobstructed opener
  or its current semantic replacement. A removed opener gets a nearby meaningful
  fallback. Do not focus an old detached copy or let late work steal focus.
- Current controls must accept genuine local editing and submission on the first
  return and a second child/Back cycle. Existing disabled states still reflect
  real pending/invalid work; never enable an action that will silently no-op.
- Back, disclosure, local typing and local display changes perform no replacement
  telemetry reads. Only the established explicit source action starts a read.

## Retired work remains retired

Renewing a current definition control must not revive its departed metadata or
telemetry request. A response completing in the child or after Back cannot replace
a current draft, options, status, observation or focus. Test identical-definition
ABA returns as well as changed definitions: matching text alone is insufficient
ownership. Old detached handler references cannot stage, submit, publish or read.
A new intentional metadata action may start its own current request through the
existing gesture; Back must not replay the old one automatically.

## Native journey and review

Cover ordinary selection-driven child entry, an authored comparison/facet and a
record-driven child. First verify that a dirty facet/window definition blocks navigation with zero IO,
then explicitly discard before child entry. Capture a permitted parent state with
an unsaved title/local Find/metadata search and chosen disclosure state, open a
child, return and continue editing. Repeat the trip and
prove a real local action succeeds after the second Back. Record source budgets
and unchanged questions/clocks separately from appearance. Include delayed
metadata and detached-control negatives in mounted/model checks.

Review actual desktop/narrow frames for reachable retained local state, disclosure state, readable
current control labels and visible focus; use independent Playwright captures.
Frames alone do not prove handler ownership or no extra reads. The observed ledger
and negative controls provide that evidence. Preserve the existing compact flow
and omit extra prose explaining an internal lifecycle repair.

## First actual frame acceptance

Reviewed `342-held-baseline-before.png` and
`342-held-baseline-returned.png`. The held comparison remains 9 baseline records
versus 3 selected records, with the same original windows, independent UTC read
clocks, aligned bars, category ordering and missing-versus-empty rows. Baseline
options and comparison disclosures remain available; the returned capture shows
the existing investigation trail at its restored viewport position rather than a
new result surface. No material visual regression is visible. These frames have
different scroll/crop positions, so they do not by themselves prove pixel-perfect
scroll restoration or focus; native viewport/center-hit checks supply that proof.
The observed first case separately checks identical comparison DOM/raw snapshots,
a fresh working group action and inert reattached old handler references.

The source concerns above describe the 341 starting point. The 342 implementation
now introduces bounded retained-definition control/state helpers and deliberately
renews the current parent controls; implementation/source acceptance is completed below with the authored cases.

## Final source and actual-frame acceptance

The implemented `renewRetainedParentDefinitions` now renews local window/record
composition leases and remounts bounded current definition editors with captured
raw values, choices and disclosure state. The helper deliberately excludes
handlers and disabled state from restoration. Departed work is retired before
child entry; current comparison/facet actions are rebound separately from held
observations. This resolves the 341 starting-point concern without a global state
manager or a new user-facing concept. Existing dirty-definition guards remain.

Reviewed actual `342-parent-local-state-before.png` and
`342-parent-local-state-returned.png`, including the returned original-resolution
frame. The unsaved view title, open Preserve/View settings and scope facts,
closed Facet, Groups mode and local “canary” Find survive. The same 5/9 pair and
2/3 canary counts, exact bounds and independent read clocks remain visible.
The existing record table and categorical bars retain their presentation; no
new explanation or repair control has been added. Final visual acceptance passes.

The full-page captures place the sticky investigation trail at different document
positions; they are not standalone proof of scroll/focus correctness. The final
native run independently checks seven Back transitions with viewport and center-hit
focus, working fresh stage actions with zero reads, and retained comparison/facet
receipts. Its 30 source requests, 27 projected transitions, 57 events and 16 raw
facts are execution evidence recorded by the native owner, not inferred from
images. Mounted/helper checks separately cover invalid raw text/caret restoration;
these screenshots do not claim blocked dirty definitions traversed navigation.
