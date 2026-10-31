# Group count comparison design review — 291

2026-10-31. Proposed interaction contract following the delivered
[captured-count comparison](../CAPTURED_COUNT_COMPARISON.md). This review inspected
source, the roadmap and the actual290 prepared-pair frame. It ran no browser,
property/model test or user study. No group comparison UI is delivered by this
document.

## Recommended bounded outcome

A responder who sees a total count change can switch the same prepared comparison
to **Groups** and see which group counts changed. Reuse the exact baseline,
current held operand, windows and preparation lifecycle. The switch is local
presentation; it must not fetch stored evidence, call comparison again, run a
query or create a capture. The title should be **Count change by group**, avoiding
causal language such as “root cause” or statistical attribution.

Keep the existing capture title/link prominent. Retain the two operand windows,
read clocks and scope/quality disclosure once. Groups do not need a second
baseline selector, preparation button, source status panel or repeated evidence
paragraph. The existing unequal-window warning remains: these are raw recorded
counts, not normalized traffic rates.

The current290 frame already occupies substantial vertical space for one measure.
An additional always-expanded chart below it would bury inspection records. Use
one native **View as** picker with **Summary** and **Groups**, replacing the
presentation within the same comparison child. A bounded optional declaration
field such as `display.view: "summary" | "groups"` can preserve that choice;
its final spelling belongs to the reviewed grammar contract. Default remains
Summary. Opening a saved definition remains staged, independent of the view choice.

## Groups presentation

Use one compact aligned table with inline paired count bars. Exact values and
visual comparison belong in the same row, so a separate Bars/Table switch adds
little value to this first slice.

```text
Baseline: <chosen captured evidence>
View as [Groups]     Find groups [          ]     Order [Change]

Group                  Baseline       Current held       Change
api                    ▰ 120          ▰▰ 180              +60
worker                 ▰  30          ▰   12              −18
(queue attribute missing)   2               2                0

Showing 1–50 of 86 groups           [Previous] [Next]
```

The numbers and sketch are illustrative, not a screenshot. The actual grouping
field labels the first column. Both count columns use one zero-based domain over
the **full aligned union**, not a separate maximum per row, side or visible page.
Finding, sorting or paging rows does not rescale that domain. Bar geometry may be
approximate; counts and changes remain exact strings, with exact values accessible
without requiring a hover. Do not silently round very large values to the same
apparently exact number.

Reuse the compact `goo-table` viewport and existing passive count-pair geometry,
number formatting and field-value labels where they fit. A generic passive grouped
comparison element is reasonable if it consumes the canonical supplied profile;
it must not own capture selection, compatibility decisions, fetching or retention.
Do not route this through the existing application-owned comparison-breakdown
adapter, which contains recapture and navigation actions outside this slice.

## Identity and missingness

- A category whose grouping field was absent is a real group, not an unavailable
  count. Give it a dedicated **Missing value** label. Empty string has its own
  **Empty string** label. Literal strings are quoted/escaped so the string
  `"Missing value"` cannot visually impersonate the absent-value group.
- A group present in only one validated complete captured grouping has count
  **0** on the other side. Keep its exact known difference and optionally expose
  **Not observed in baseline/current captured groups** in a short tooltip or
  disclosure. Do not label that zero as unknown or as an absent attribute.
- Unknown/unavailable count is rendered **Unavailable** with no value bar. The
  reviewed first profile withholds every count on a diagnostic side (null values
  with an explicit reason); it may retain the other coherent side's exact text.
  A diagnostic comparison suppresses all paired bar geometry and must not show
  a numeric difference, ranked increase or
  regression color that implies it is usable. Preserve the canonical reason.
- A known-zero operand may produce an available exact difference even if the
  envelope is `no_data`. Reuse290's concise known-zero explanation rather than
  suppressing useful rows or inventing a new health state.
- No percentage, normalized share, rate, unique-invocation claim or significance
  badge is needed. Do not color increases as bad and decreases as good.

## Local exploration controls

Default order is descending **absolute supplied count change**, with a stable
category-identity tie-breaker. This surfaces both increases and decreases without
making a health judgment. The other useful order is **Group name**. Use the
canonical supplied change value for ranking; the renderer must not derive a new
change from the two count columns. Unavailable changes sort after available ones.
Keep unavailable rows visibly diagnostic rather than silently dropping them.

Find is a local literal label search, not a population filter. Show **No matching
groups** distinctly from **No groups observed** and **Comparison unavailable**.
A bounded union can reach2000 rows, so do not place all rows into an unbounded
page. Use local row paging or an existing bounded viewport (for example50 rows),
with **showing N of M groups**. Its page count/search count never becomes the
population total or changes the prepared pair. No server request occurs.

Defer group-click drilldowns, chart swipes and “investigate this side” until an
explicit action-ownership contract exists. A passive row is not a query action.
The first slice should establish a trustworthy descriptive breakdown before
adding source-navigation gestures.

## Accessibility and retained state

Use native labelled select/search/buttons and a semantic table with column and
row headers. Expose exact Baseline, Current held and Change text for each row;
bar tracks are decorative and can be aria-hidden. Do not require color, pointer
precision or tooltip-only values to distinguish sides or signs. Sign is visible
in the change column; scope metadata identifies which operands are compared.

A local view/search/order/page change should preserve keyboard focus on its
control and retain the current record-page selection. It must not render the
entire population owner merely to change a passive display. Existing baseline
changes and root Run retire the pair through its current owner. Local controls
must not reactivate old preparation or allow a delayed old pair to return.

On small screens, retain aligned numeric columns in the existing horizontally
scrollable compact table instead of turning each group into a verbose card.
Provide a visible label for the scroll region and test keyboard traversal and
horizontal overflow. Real assistive-technology testing remains a separate gap;
native markup and Playwright checks are not a screen-reader study.

## Acceptance before delivery

Use the same real prepared comparison for Summary→Groups→Summary with zero
additional storage, comparison or telemetry requests. Include a group that
increases, a group that decreases, a group present on only one side, missing
attribute, empty string and a literal label resembling missingness. Assert exact
rows and the global common domain, including after local Find/order/page changes.

Late record arrival and inspection paging must leave both compared group snapshots
held. Save/reopen keeps configuration and stages it without implicit acquisition.
Known-zero, incompatible and changed-between-reads cases remain explicit. Check
no-match versus no-data wording, source-free controls, keyboard focus retention,
record-selection retention and refusal of retired controls. Review actual desktop
and narrow-screen frames alongside executable journeys before claiming delivery.


## Canonical proposal disposition

Reviewed the actual [group-count proposal](../GROUP_COUNT_COMPARISON_PROPOSAL.md).
Its separate `goo.group-count-comparison/1` envelope, shared summary operands,
typed nullable category identity, bounded full union and explicit side accounting
fit this interaction. Coherent absence is a known arithmetic zero; diagnostic
absence remains unknown and diagnostic values/changes are unavailable. This
conservative profile removes the ambiguity of simultaneously showing a positive
value and an unavailable state. No source completeness is inferred from the
supplied accounting. The optional `display.view` vocabulary agrees with the
proposed local Summary/Groups choice.

No design blocker remains for the proposed first implementation. Preserve exact
category text in a keyboard-accessible disclosure when labels are visually clipped,
and keep unit labels in headers where possible rather than repeating a long unit
phrase for every compact count cell. Final source/actual-frame review is still a
delivery gate; this proposal review claims no tests passed.

## Implemented source and actual-frame review

Reviewed the implemented passive renderer, comparison owner and the actual
[aligned groups](../journeys/screenshots/291-aligned-group-counts.png),
[held groups with fresh page](../journeys/screenshots/291-held-groups-fresh-page.png)
and [known-zero narrow frame](../journeys/screenshots/291-known-zero-groups-mobile.png).
Executed outcomes belong to the verifier and [VALIDATION](../../VALIDATION.md);
this reviewer did not run those checks.

The aligned table makes the comparison readable: baseline-only/current-only rows
retain known zero, opposite count changes rank together, and the same common
scale applies to all rows. Missing value, empty string and the literal string
`"Missing value"` remain visually distinct. The fresh-page frame retains the
compared shared group at2→3 while the inspection page reports64, making the held
boundary visible. The known-zero frame keeps available positive changes without
inventing a group on the empty side.

Source confirms local Find/order/page redraw only the passive table body, keeping
the search control and parent record selection outside that redraw. The host
retains the prepared element rather than recreating a new acquisition owner.
Long labels now have a bounded80-code-point preview and a native keyboard-openable
exact-value disclosure. This avoids unbounded4096-byte rows; wrapping in the narrow
frame remains bounded. The horizontally scrollable compact table intentionally
keeps its aligned columns instead of introducing verbose per-group cards.

One material clarity correction is requested before final delivery: Groups hides
the summary's visible unknown-upstream/extent text. The focused Groups frame
needs one compact shared caption identifying these as supplied captured groups
with unknown upstream coverage. Windows and coherent partition accounting alone
do not communicate that limit. Reuse the canonical quality/policy; do not invent
a source completeness claim or duplicate the full provenance header. The parent
and verifier own that small correction and refreshed-frame check.

No other material visual blocker is identified. One optional accessibility
improvement is to expose aligned one-sided zero's existing “not observed in the
supplied side” tooltip as screen-reader text as well; a new visible badge or
control is unnecessary. Repeated `records` units and singular phrasing can be
refined later without changing this model.

## Next bounded user capability

The useful next step is **stage this group on an explicitly chosen comparison
side**. A responder who identifies a changed group should be able to inspect its
baseline or held-current population without manually rebuilding the predicate.
This requires a reviewed side-specific query/action authority contract first:
bind the selected canonical group identity, exact operand selector/window and
held basis; stage a retained child without a source read; keep the comparison
parent returnable; require explicit Run; reject retired or diagnostic actions.
Do not turn a generic table row into a fetch command or infer a single shared
window. This extends the investigation naturally; another measure selector or
unrelated chart type is less valuable at this boundary.

## Final visual disposition

The delivered source adds the compact **Supplied captured groups · upstream
coverage unknown** caption, and the refreshed narrow known-zero frame visibly
retains it above the independently labelled windows. This resolves the only
material clarity correction identified above. The aligned-zero cell also exposes
its “not observed in the supplied side” explanation through an accessible label,
so it is no longer tooltip-only. No new semantic state or control was introduced.

No material design blocker remains. Final browser/model/property outcomes are
recorded by their execution owner; this visual/source review does not claim to
have run them.
