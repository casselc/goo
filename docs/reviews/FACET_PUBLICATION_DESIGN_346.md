# Facet publication consolidation — design review 346

Status: source and all four actual desktop/narrow frame reviews accepted. This
is an internal publication repair, not a redesign or a new facet workflow.
Native measurements and execution results are recorded separately.

The user should see the same exact values and continue the same local interaction
with less repeated work. Fewer render calls are useful only when source-derived
state, current action ownership and the visible result remain coherent.

## Source review

`ui/categorical-facet.mjs` separately renders on connected `data` and `interaction`
updates. `ui/population-workspace.mjs` publishes interaction and data during owner
setup and source/state updates. A render rebuilds the subtree, retains local Find
by receipt identity and restores focus by accessible name; Read details open state
is retained for the same read. These paths are candidates for duplicate work,
not proof of the total roughly 32,000-element paired cost.

Measure builds separately for owner setup, source publication, stage and retained
Back before choosing an atomic publication method. Keep mount/update/next-paint
measurements distinct, with actual styled production paths and matching fixtures.
Do not add a global scheduler, virtualizer, paging, source cache authority or
asynchronous owner-adoption interval.

## Observable behavior acceptance

- All admitted values remain fully searchable. Keep the bounded prefix/suffix
  preview, full literal title/accessibility identity, exact missing/empty category
  distinction and independent selected/baseline counts and receipts.
- Same-read publication preserves local Find and current accessible focus. Read
  details open/closed state and intended local scroll remain stable. The user must
  not see a flash of an empty filter or jump to a different action. Arbitrary
  mid-string caret/selection preservation is a separate source finding below; do
  not claim it from the existing terminal-caret native probes.
- A true field/signal/read-basis identity change keeps the existing reset semantics.
  Preserving Find indiscriminately across new observations would be a behavior
  change, not an optimization. A held receipt is evidence, not cache permission.
- Publication adopts supplied data and interaction authority synchronously as one
  current state. No event may observe new values with an old action owner, or old
  values under a new owner. Departed/detached handlers remain inert even for an
  identical-definition ABA return.
- Keyboard Tab/Shift-Tab reaches the same visible controls. Hidden filtered rows
  stay out of focus order; Include/Exclude stages the exact intended full literal.
  Current-owner disables, loading states and partial-side unknown geometry remain
  faithful to the supplied state.
- Source publication cannot steal focus from an unrelated active control. Retained
  Back restores the current semantic opener, not a detached old node. Existing
  focus-visible treatment and native search clearing behavior remain intact.
- Stage, Find, disclosure, local publication and Back add zero source reads or
  writes. Any improved render count must not be purchased by silently dropping a
  meaningful update, receipt, status or current-owner refresh.

## Actual acceptance plan

Review genuine desktop/narrow frames before and after with the same local Find,
long-value identity, exact paired counts and disclosure state. Capture the task
position that actually shows the control under review; no above-fold claims from
cropped frames. Native checks must exercise typing/caret, keyboard Include and a
retained child/Back transition with source budgets and original receipts. Use
mounted/negative checks for intermediate ownership, detached old handlers and
new-receipt reset. Frames establish appearance; measured builds and observed
checks separately establish performance and semantic equivalence.


## Frozen source review and evidence limit

The bounded implementation adds synchronous `publish({data, interaction})`,
validates both inputs before adoption, prepares detached elements before append,
and uses an owner-local last-publication guard. It does not defer adoption through
a microtask or turn held source data into shared cache authority. Pending read
normalization produces a new object, so legitimate unavailable/partial/full source
transitions remain publishable. Source acceptance passes for this internal cut.

The original component restores focus by accessible name and does not capture
arbitrary text selection offsets. The implementation does not change that behavior.
The prior 343/native probes place the caret at End; those probes cannot establish
mid-string selection preservation. Treat that as a bounded future usability finding
requiring its own actual test, not as a new claim or an implicit 346 repair.
Actual baseline/final publication frame and measurement acceptance remains pending.

## First final actual frames

Reviewed `346-final-1280-ordinary-returned.png` and
`346-final-1280-pair-returned.png`. Both retain the focused full-suffix Find,
unique terminal-selected-0997 preview, staged exact literal and current action
controls. Ordinary remains 1 of 1,000 loaded values; the pair remains 1 of 2,000
with Selected 1 / Previous window 0, while the independent whole-population
summary remains 1,000/1,000. Desktop appearance passes. Narrow acceptance awaits
its final frames. Images establish these visible facts, not the measured five-to-
one build reduction or arbitrary caret selection preservation.

## Final actual-frame acceptance

Reviewed `346-final-390-ordinary-returned.png` and
`346-final-390-pair-returned.png` after the desktop frames. Both narrow surfaces
retain the visibly focused terminal-selected-0997 Find and show its distinguishing
suffix with Include/Exclude actions inside the viewport. Ordinary retains
1 of 1,000 loaded values and count 1; the pair retains 1 of 2,000, labelled
Selected 1 / Previous window 0 and the separate 1,000-record selected chart.
The same native details affordance and unknown-coverage language remain. No
material appearance regression or page overflow is visible. Final visual
acceptance passes across all four supplied frames.

The native checks report five-to-one ready builds on Back with exact receipt,
Find and current-action preservation. This is not an overall Back-speed promise:
wall observations vary, including a slower desktop paired sample, despite less
synchronous build work. Terminal-caret End usability is checked; arbitrary
mid-string selection preservation remains outside the demonstrated result.
