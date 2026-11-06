# Large counted facets — design review 343

Status: source and actual desktop/narrow frame acceptance complete for the bounded
single-pass filtering and label-preview repair. Measurement results are recorded
separately in the native journey and validation record.

A user reads a scope's available values, types a distinctive fragment and includes
or excludes the intended population. The task is finding a useful value and
keeping its exact meaning, not scrolling through a thousand controls. Preserve
that natural existing search-first interaction without another query surface.

## Existing source

`ui/categorical-facet.mjs` supplies all admitted values to `goo-facet-list`, with
up to 1,000 values per read and a 2,000-value paired union from
`ui/categorical-pair.mjs`. The list's local `refresh` scans full `data-facet-text`
and hides unmatched rows. Rows are created eagerly. Text is visibly abbreviated
after 160 characters, but full literal identity remains in the title, full search
text and accessible Include/Exclude button names. The list already has bounded
vertical scrolling; paired narrow layout retains per-side labels.

These are potential costs, not evidence of a user-visible failure. Do not add a
virtualizer, paging, debounce or framework because a large row count sounds risky.

## Practical measurement budget

Measure a genuine 1,000-value single read and disjoint 1,000/1,000 paired union
at desktop and narrow widths, including long admitted labels. Record synchronous
mount work, next paint, DOM count and local input-to-next-paint for a short typing
sequence, unique match, no match and clearing. Distinguish source latency from
component work and disclose browser/CPU conditions.

Use 100 ms input-to-paint as the responsive interaction goal; repeatable samples
above 200 ms merit investigation. An initial local mount within about 500 ms on
desktop and 1 second in the narrow browser profile is a diagnostic target, not a
promise about all devices. Width alone does not simulate a slow phone. Report
actual distributions and individual worst cases rather than declaring a pass
from a single mean. Avoid fragile millisecond CI assertions; structural tests and
repeatable before/after measurements should establish a demonstrated improvement.

## Interaction and exactness acceptance

- Every admitted value remains searchable using its full literal content,
  including suffixes beyond the visible abbreviation. A unique tail match must
  not disappear because only the first page or first 160 characters was searched.
- Missing key and present empty string stay distinct. Baseline-only values show
  known selected zero when both partitions are ready; an unread side stays unknown.
  Counts, side labels, independent receipts and global comparison scale remain exact.
- The existing loaded-match count describes all matches, not mounted/visible rows.
  Clearing restores all available values and leaves the search control focused.
- Keyboard users can Tab from search to available Include/Exclude actions and
  activate the exact intended value. Hidden filtered rows must not receive focus.
  Ordinary native text editing, caret, selection and search-clear behavior survive.
- Long and Unicode labels wrap or abbreviate without pushing count/action columns
  beyond the viewport. Full identities remain available to assistive technology
  and search. Similar prefixes must not make the chosen action ambiguous.
- Local searching, scrolling and any presentation-only paging cause no source
  reads, writes or scope changes. Staging still requires its existing explicit
  Apply/Open flow; disabled/departed controls retain current-owner guards.

If measurement demonstrates a problem, prefer the smallest shared-component
repair. If bounded presentation paging proves necessary, search the whole admitted
set before slicing, show total matches separately from the page, retain the search
node/caret, make paging keyboard reachable and preserve exact value identity.
Do not add paging controls unless their benefit exceeds their interaction cost.

## Pending real-frame acceptance

Review actual desktop/narrow baseline and, if changed, final frames for search,
unique long-label match, zero matches and keyboard focus. Include a paired
baseline-only result. Frames establish appearance; observed actions/source budgets
and measurements separately establish responsiveness, complete search and exact
staging. No fabricated telemetry or mocked full-app response substitutes for the
native integration journey.

## Long-label source concern awaiting frames

Prefix-only abbreviation can hide the only distinguishing suffix of two values.
The full accessible action name helps a screen-reader user, while a static span's
`title` does not provide a reliable sighted-keyboard inspection gesture. Full-text
search still finds a unique suffix, but the abbreviated result does not visually
confirm that suffix. Inspect this in actual frames before accepting the surface.

If demonstrated, prefer a bounded Unicode-safe prefix-and-suffix preview over
adding paging. Retain the full searchable text, accessible action names and exact
staging identity. Prefix/suffix is still a preview and cannot promise uniqueness
for differences in its omitted middle; do not label it as the full value. Any full
value disclosure should reuse a passive existing pattern and avoid thousands of
new default tab stops. This is a proposed usability repair, not an implemented
change or a performance conclusion.

## Actual baseline frames and bounded decision

Reviewed `343-baseline-single-1280-loaded.png`,
`343-baseline-single-390-loaded.png` and
`343-baseline-single-390-suffix.png`. Desktop rows with the shared long prefix
are visibly indistinguishable. More decisively, narrow full-suffix search for
`terminal-selected-0997` finds exactly one of 1,000 values, but the matching row
still displays only the common prefix and repeated characters. The user cannot
visually confirm the suffix they found before activating Include/Exclude.
Missing and present-empty values are separately labelled as intended.

This demonstrates a bounded identity-preview issue, independently of performance.
Design endorses a Unicode-safe prefix-and-suffix preview within the existing
160-character bound. Keep the terminal distinguishing text visible while retaining
the full literal in search, title, accessible action names and emitted staging
identity. Add no paging or new controls for this repair. At this baseline
checkpoint, final actual-frame acceptance was pending; the baseline narrow loaded capture does not show the
long rows above the fold and is not proof of their detailed layout.

## Final single-side source and initial visual review

The implemented preview uses 104 initial Unicode code points, one ellipsis and
55 terminal code points, bounded to 160. Full title, search text, accessible names
and literal action identity remain unchanged. The shared list gains an atomic
filter call that updates its query and returns match statistics with one refresh;
the existing property/attribute paths remain supported. No paging/control is added.

`343-final-single-1280-loaded.png` now visibly distinguishes
`terminal-selected-0000`, `0001` and `0002` in the loaded rows. This resolves the
actual desktop prefix-only ambiguity. The initial final narrow suffix viewport
ends before the long row's terminal identifier; it shows no layout overflow but
cannot independently prove terminal-suffix readability. A task-scrolled narrow
row frame is requested for that specific visual claim, without another read or a
new product change. Native DOM checks separately establish 160-code-point bound,
full 384-byte title and visible suffix text in the rendered element.

## Final narrow acceptance

The existing `343-final-pair-390-suffix.png` supplies the needed task-scrolled
view without another read or replay. The whole terminal-selected-0997 suffix and
Include/Exclude actions are visible together. Selected 1 and Previous window 0
remain labelled, the search stays visibly focused and the result count is
1 of 2,000 loaded values. The row wraps within the viewport without clipping.
Together with the final desktop frame, actual visual acceptance passes for the
bounded preview repair. No supplemental capture or source action is necessary.

These frames establish the measured case's prefix/suffix readability, not
uniqueness for every possible omitted-middle difference or universal device
performance. The native measured input/paint results and exact action/source
checks remain separate from this visual judgment. Eager mounting remains by
design; no new paging or virtualizer was justified by this review.
