# Duration timeline review — 287

2026-10-31. Product/design and architecture source review for the bounded saved
count/mean/p95 timeline cut. Implementation and live validation were pending at
initial review. This review runs no heavy checks; executed results belong in
[VALIDATION](../../VALIDATION.md).

## Product and model contract

The useful outcome is seeing when captured span latency changed beside matching
records in a reusable composition. The parent query supplies duration fields only
for spans with `summary: span-duration`. Local measure switching consumes supplied
fields; it cannot silently enable a summary or issue a source request. Keep count
as the default, and offer mean/p95 only where the query supports them.

Keep three distinct checks: query support, a supplied cell's availability/support,
and verified interval coverage for complete-only display. An empty duration cell
can coexist with complete captured-query interval coverage. Unknown coverage,
known empty input, no valid values and a legitimate zero duration are different
states. No scalar percentiles are averaged into a fleet percentile.

Reuse canonical `formatCell`, `exactText` and support metadata. Human units belong
in the visible plot; exact ns/rational values, quantile method and support belong
in accessible values/tooltips. The p95 method remains the source's
`quantilesExactHigh`, not a renderer calculation. Counts remain counts, not rates.

## Rendering and actions

Use the canonical temporal plot/table over the held time dataset. The existing
count-specific BigInt presenter must not become a second duration formatter.
Preserve exact count precision and current zero/unknown count behavior while
adding duration measures.

The canonical temporal plot initially had click selection only. Pointer drag and
Space/Shift+Arrow/Enter range selection should reuse the bounded passive gesture
implementation, emitting supplied endpoint keys. H1 remains responsible for
contiguous known interval membership and deriving exact shared bounds. Choosing a
measure must not change which intervals may be staged.

Reset local selection on presentation changes. Detached old plots must not emit
an actionable range through a still-connected outer host. Validate current track,
keys and connection at gesture commit and accept events only from the current
presenter chart. Existing owner/epoch checks still revoke hidden or retired mounts.

Known zero needs care: the initial canonical plot mapped every non-count zero to
`data-state=no_data`, whose CSS draws a baseline mark. Preserve a visible valid
zero without mislabeling it empty. A null `no_data` duration must never become zero.

Canonical tables expose supplied columns. If selecting a measure does not narrow
the table, make that behavior clear instead of constructing an ad hoc projection
that loses completeness lineage.

## Acceptance before shipment

- Count/mean/p95 save and reopen with exact width and query-support checks.
- Local switching and paging add no source reads and retain held timeline inputs.
- Human duration units, exact values/support/method and a valid zero are checked.
- Known empty, unknown/partial coverage and complete-only refusal remain distinct.
- Pointer and keyboard ranges stage the same checked half-open window; Run is explicit.
- Current hidden/detached/stale controls refuse without changing scope.
- The focused authoring entry reaches this profile with no seeded-only UI dependency.

## Frozen source follow-up

The inspected implementation replaces the count-specific presenter with
`timeline-view.mjs` over canonical `goo-chart`. The range gesture is extracted into
`interval-range.mjs` and shared with the existing count/histogram bars. Commit
checks include connected/current tracks and live supplied keys. Canonical clicks
also reject detached tracks; the presenter accepts events only from its current
chart and resets selection when presentation changes.

The inspected duration-zero state is observed with a known-zero mark; known empty
duration cells remain value-less and shaded. The non-count legend is generic,
so numeric/status timelines are not mislabeled as duration. Table mode hides the
measure selector and states that it shows all supplied fields. Suspended timeline
controls are disabled. No material source-level blocker remains. Browser results,
regressions and screenshots are still the parent validation responsibility.

## Actual-frame review

Reviewed the generated duration plot and exact-table frames. The p95 title and
3 ms / 100 ms values are readable on one shared scale; the valid zero remains
visible, while empty intervals have no duration value. The table exposes UTC
bounds, human duration units, cell state and support rather than presenting missing
duration as zero. Its all-supplied-fields behavior is explicit.

The initial frame revealed a legend mismatch: the teal zero mark was labeled
No data while empty duration intervals were neutral bands. The inspected fix now
uses four non-count keys: measure fill, Known zero, neutral No data and dashed
Unavailable/unreported. Count behavior is unchanged. The final refreshed frame
confirms all four keys after that fix. No further visual blocker was found in
these bounded fixture frames; this is not a high-volume usability or performance
study.

## Canonical input boundary found during live validation

The first live migration attempt reached canonical display admission but supplied
an optional `title` key with the JavaScript value `undefined`. The canonical
constructor correctly refused that value even though the separate structural
validator had accepted the display shape. The caller now omits an absent title,
and a boundary test exercises the actual canonical constructor. This was found
by the live flow, not by the initial source review or pure validator checks.

Unspecified duration plot headings are now derived from the supplied field label
(for example, p95 duration over time); authored titles remain unchanged. This
keeps the plot measure-specific without another formatting/method table.

The final timeline and reopened-duration frames were inspected after the legend
fix. They show the p95 heading, human values, valid-zero mark and four distinct
legend keys. No material source or visual blocker remains in this bounded cut.
The parent reports the final live journey passing in 38.2 s, count regression in
21.2 s and the separate Bombadil campaign with 587 snapshots. This reviewer did
not execute those campaigns; complete results remain in VALIDATION.
