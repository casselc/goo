# Record list and selected-measure table review — 297

2026-11-01. Source/design review against the 296 checkpoint, roadmap and existing
record presentation/page/column owners. This is a proposed interaction contract;
no browser, accessibility or performance test was executed for this review.

## Decision

Add **List** as a second kind of the existing records child, alongside **Table**.
Use the same supplied immutable page, configured columns and qualified fields,
scope/read facts and pager. Do not create a selected-record role, choose the first
row implicitly, or introduce another investigation/modal owner. Save and template
reuse should carry the record kind with the other presentation choices.

A list is useful for scanning log bodies and spans whose summary/attributes need
more width than a table column. Keep it dense: divider rows rather than large
cards, restrained padding, and no repeated per-record provenance paragraphs.

## List content and interaction

- Put **View as: Table / List** beside the existing record fields controls. Keep
  **Table** as default and reuse the same choice in the records property editor.
  Changing kind is local presentation, not a telemetry read or a change to scope,
  ordering, page offset, count or measurement totals.
- Respect configured columns. The compact first line holds UTC and qualified
  service when enabled; the summary uses existing `goo-record-summary` for log
  severity/body or span name/reported status/duration. Use existing UTC formatting
  and exact timestamp tooltips, not nanosecond numbers as the main display.
- Show configured qualified fields as small label/value pairs. Reuse the existing
  field lookup/formatting; distinguish **Missing**, **Present empty string**, null,
  and literal text. Preserve field namespaces and case. Do not add predicates or
  infer missing values from absent map/column metadata.
- A native **Details** disclosure can show the existing supplied-record detail and
  structured attribute viewer. It must not fetch, infer a complete trace, or make
  parent ancestry claims beyond supplied observations. Keep expansion state local;
  it is not saved selection or evidence.
- Keep the current page facts, matching count, continuation limits and Previous /
  Next controls outside the list, owned by the application. A list of five supplied
  rows is not five matching records or proof of complete upstream telemetry.

The current selection/actions adapter depends on table rows and table column
controllers. The bounded first list can be explicitly display-only: keep one short
note **Use Table for selection and source actions**, with a quick Table choice.
Block switching while selection/submission/page work is pending so changing kind
cannot discard intent. Do not copy Keep/filter/trace logic into another controller.
If full action parity is required now, extract a shared owner first; never claim
parity from identical-looking buttons with independent guards.

## Selected-measure visibility

The current exact-table path uses every snapshot field in schema order, so a chosen
p95 can sit after count and several latency fields outside the initial viewport.
For table rendering, place dimension/category fields first, then the selected
measure (deduplicated), then all remaining supplied fields in their original order.
This makes the user's choice visible without inventing a new All-fields wizard or
hiding source/support information. Keep fractions, unavailable/diagnostic states,
units, tooltips and row selection behavior unchanged.

This is a render-column ordering rule, not a dataset projection. Do not construct
a smaller snapshot, recompute totals, drop guarantee fields or transfer a witness.
For multi-dimension tables retain all declared dimensions before the measure. A
real browser assertion should check the selected cell is in the initial viewport,
not merely that its text exists somewhere in a wide table; wrap/cap long labels
only if needed without losing their exact tooltip text.

## Acceptance

A real log/span page should switch Table→List→Table without a read, preserve exact
UTC/summary/qualified fields and missing/empty distinctions, open supplied Details,
page through the same application controls, and save/reopen the chosen kind in a
view/template. Unsupported/pending actions must leave intent intact. Include a
narrow viewport frame and keyboard operation of the kind choice/Details. Separately,
replay the 296 DB template with p95 Table and prove the selected measure is visible
before horizontal scrolling, with the original values, totals and quality intact.

## Implemented source review

The implementation follows this bounded choice: Records view offers Table/List;
List reuses the admitted page, configured fields and outer pager, while explicitly
providing Use Table for selection/source actions. The kind is a property of the
existing records child, so saving or reusing a template does not add an inspector
role or select a record implicitly. Switching is guarded while work or selection
is pending. The passive list uses compact divider rows, existing UTC/summary/
context content and lazy native Details with the existing structured viewer.

The exact-table helper creates only a new render-order array: dimensions, chosen
measure, remaining supplied fields. It retains field references, rows, support and
quality; it does not shrink or relabel the snapshot. No source UX blocker found.

The shared field picker now uses neutral **Trace context**, replacing **Trace and
inspection actions**. This correctly describes the supplied List content without
promising source actions, matches the table header and avoids a kind-specific label
API. Configured values use a bounded two-line visual preview; exact contents remain
in supplied Details. Actual-frame review should check metadata density (the app and component
both show page facts), large configured-value wrapping and the selected-measure
initial viewport. These are presentation observations to verify, not new source
or completeness requirements. Keyboard/mobile claims await executed tests.


Final source recheck: temporal tables retain both interval anchors before the
chosen value, and histogram tables retain from/to bounds. The ordering rule still
keeps all fields. No further source UX blocker found; actual frames remain pending.

## Actual frame review

Inspected the generated supplied-list, paged-list and selected-measure frames.
The paged List is compact and readable: UTC and qualified service lead each item,
summary/status/duration remain together, and `resource:zone` and `attribute:zone`
show their distinct values. Table-only actions are clearly described with an
immediate Use Table button. The offset100 frame reports 103 matching at page read
and three supplied rows separately, with continuation and the same outer pager;
it does not visually substitute page length for the population count.

Expanded Details uses the existing structured attribute viewer with type/empty
labels and exact values. It is intentionally large when opened, rather than
expanding every row by default. The selected-measure frame places p95 duration
immediately after the operation category, with exact millisecond values and support
text; additional fields remain to the right. The executing journey owns the
viewport bounding-box and unchanged-data assertions.

No product UX blocker found. Two capture issues should be cleaned if the final
journey is rerun anyway: capture the `goo-table` viewport rather than the overflowing
raw table (the latter creates a wide blank image), and hide the unrelated sticky
trail in capture styling so it does not cover the first list row. Repeated app and
component page facts remain a small density follow-up; the current paged frame
still fits all three rows and its pager. These are framing/presentation limits,
not a reason to change source scope, continuation or supplied-data semantics.
