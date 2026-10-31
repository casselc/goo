# Contextual captured-baseline comparison — 295

2026-10-31. Source review of the 294 responder journey and current population,
composition and capture-picker owners. This is a design recommendation, not an
executed browser or accessibility review. Actual results belong in VALIDATION.

## Decision

Expose **Compare with captured baseline** beside the applied aggregate, as a
closed native disclosure containing the existing capture picker. Keep one picker
instance and one composition/query owner. Do not add a second comparison wizard,
new saved object, or automatically captured baseline.

The 294 journey demonstrated that the task works, but requires knowing to open
View settings → Composition. Putting an entry beside the data connects intent to
that task. Merely scrolling into Composition is a defensible smaller fallback,
but still exposes internal editor roles and larger settings to the primary flow.

## Placement and wording

- Place the closed disclosure near the aggregate's heading and local controls,
  before or immediately after the chart. It must not add a full-width toolbar
  above the first records or expose the whole picker by default.
- Use **Compare with captured baseline**. This distinguishes stored evidence from
  previous-window reads and live baselines without requiring a terminology panel.
- Reuse the existing capture-title search, **Load baseline captures**, choice,
  exact-ID fallback and removal behavior. Search/load remains explicit. Opening
  the disclosure performs no metadata fetch, telemetry read, capture or save.
- Selecting a capture configures its exact reference only. Keep the existing
  explicit **Prepare baseline** action and its result/compatibility diagnostics.
  Do not combine selection with preparation or imply a common source transaction.
- If the aggregate is hidden, retain one reachable owner beside the comparison
  child. Staged/records-only configuration may fall back to settings. A settings
  entry may reveal the same owner; it must not mount another picker.

## Unsupported profiles and pending state

Keep the contextual entry understandable when a planner-enriched span query asks
for status. Show a concise reason, such as **Turn off span status to compare record
counts**, with **Edit span measures** revealing the existing scope disclosure.
That action only navigates; the user chooses the measure change and explicitly
Runs it. Never strip status, rewrite captured questions, or substitute a baseline.
For other unsupported profiles, show the checked profile reason and preserve the
specialized editor boundary rather than advertising a conversion that does not exist.

Pending selections, query/display drafts, paging and preparation retain their
existing refusal behavior. Explain the current reason near the contextual owner;
a disabled control with only a tooltip should not be the entire explanation.
Changing the baseline withdraws the old pair, while delayed metadata or retired
controls must not replace a newer choice.

## Acceptance

The browser journey should enter from the applied aggregate without source JSON
or Composition navigation, load/select/prepare a real capture, inspect Summary and
Groups, and save/reopen the exact reference. Opening/closing and changing local
presentation must add no reads or writes. Unsupported status should route to the
existing explicit measure controls. Test both visible aggregate and hidden-
aggregate fallback, pending refusal, stale picker handles and late catalog work.

Capture actual compact closed/open frames and the prepared comparison. Check the
native disclosure from the keyboard, preserve route casing and shared comparison
scales, and distinguish HTTP endpoint counts from physical source-read counts.

## Implemented source review

The current implementation follows the recommended single-editor arrangement.
`mountCountComparisonDisclosure` places one native closed/open owner after the
applied aggregate, falls back beside a visible comparison when the aggregate is
hidden, and uses Composition settings only when no applied display can host it.
It mounts outside the cached chart host, so presentation reuse cannot preserve a
retired picker. No second catalog or comparison controller was introduced.

The unsupported-profile body explains the ordinary-count boundary and exposes
**Edit span measures** where that existing editor is available. This opens the
scope/measure disclosures and focuses the status checkbox; it does not toggle it,
read data or reinterpret a stored capture. The normal picker retains explicit
Load/search, exact reference and preparation semantics. Selecting a baseline
reveals and focuses an already-visible **Prepare baseline** command without
executing it. A deliberately hidden comparison stays hidden with an explicit
instruction to show it before preparation.

Disclosure state, search and ID draft are retained across legitimate redraw/Back;
the current mount and ownership guards retire older controls. The concise passive
summary keeps the default chart surface compact. Source UX review found no blocker.
The pending-state context still supplies the reason when editor controls disable;
this cut need not add a second repeated warning panel.

This review establishes source semantics and placement only. Final keyboard,
compact-layout, catalog/prepare and hidden-aggregate behavior still require the
executing browser journey and actual frames; no browser was run for this review.

## Actual frame review

Inspected the four generated 295 frames after the main browser assertions passed.
The contextual picker sits directly below the applied route breakdown, separated
by the native **Compare with captured baseline** disclosure; it no longer requires
Composition navigation. The open picker fits its container without visible
horizontal overflow. The unsupported-profile frame gives a short reason and a
clear **Edit span measures** action instead of presenting unusable capture choices.

The prepared Groups frame retains original `/checkout` and `/health` casing and
aligns baseline/current counts on the shared 0–2 scale. The reopened-comparison
frame shows the same checked baseline link and held pair, with the sole picker
reachable beside the comparison when the aggregate is hidden. The explicit
**Prepare baseline** control remains visible and distinct from baseline selection.
No visual blocker found in these frames.

The screenshots show opened disclosures and prepared results. Closed-state
compactness follows the native/source arrangement; keyboard activation, focus
movement, source-free opening and stale controls belong to the executing journey's
assertions, not to this visual review. Mobile, screen-reader and load behavior
were not assessed here.

The open picker's native fieldset border is brighter than the adjacent app cards.
Recommend the same scoped palette border, 8px radius, padding and min-inline-size
used by the existing dashboard picker. This is a small consistency correction,
not a global fieldset restyle or a new layout; the executing owner will regenerate
the final frame after applying it.

Final regenerated picker frame reviewed after the scoped fieldset correction.
Its subdued palette border and rounded corners now match the adjacent chart
cards, while the existing controls remain readable and fit the container. No
further visual change or design blocker remains for 295. The final executing run
and 294 regression own the interaction/read assertions, including explicit setting
of a hidden comparison without reading or changing its visibility.
