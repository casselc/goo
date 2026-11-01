# Supplied span timing — design review 301

2026-11-01. Initial contract/source review for milestone 301. Implementation and
actual browser results are recorded below when available; this initial proposal
does not claim a keyboard, screen-reader or rendered-frame audit.

## User task and entry

From a scoped span population, choose **Span waterfall** beside Table/List to see where
time went in the supplied observations. Inspect or Keep a relevant span using the
same record interaction owner, then explicitly Expand trace when cross-service
context is needed. This is another view of the current page, not another trace
query or automatic row selection. Logs must not offer the span-only kind.

The page's current scope, pagination, read basis and held aggregate/timeline stay
visible and unchanged. View/template save, apply and reopen retain the kind;
selection and publication remain read-local workflows, not saved event identity.

## Timing that humans can compare

A service page may contain many independent short traces spread across minutes.
One page-global wall-clock axis would make their millisecond spans unreadable.
Use each valid trace group's earliest **supplied** span as its relative origin and
ONE common duration scale across all groups. A 5 ms bar must remain visibly shorter
than a 500 ms bar. Do not independently autoscale each group into equal-width
fragments, which would imply equal latency.

State the relative-axis meaning once near the shared scale. Retain absolute UTC
origins and exact duration/timestamp tooltips. An origin is not necessarily the
real trace start, and an observed fragment extent is not complete trace latency or
a critical path. Keep upstream coverage unknown. Parallel/overlapping reported
spans remain visible without inferring causal relationships from time alone.

Use compact group headers: a short trace ID with full accessible/tooltip text and
UTC context, not repeated oversized IDs or card headings. Keep timing rows dense
and divider based; avoid a separate explanatory paragraph in every row.

## Identity, uncertainty and fields

Reported parent IDs can support layout only when the trace and parent identity
are valid and unambiguous in this page. Show missing, duplicate/ambiguous and
cyclic parent cases explicitly. Invalid or unreported trace occurrences remain
isolated observations; grouping them together must not imply a common trace.
Do not silently discard duplicate observations or select one arbitrary duplicate
parent. Async links remain references rather than parent edges.

Original page positions identify observations even when parent-first visual order
differs. Inspect and Keep must act on the visually chosen row's original position.
Retain all eight supported qualified fields and existing missing/empty/null/value
formatting. Rich values may wrap or truncate with their exact accessible value;
resource and span attributes with the same key remain distinct. Invalid geometry
must leave an accessible supplied-record/table alternative instead of losing rows.

## Interaction and accessibility

The timing bar is a native local Inspect affordance when an application adapter
is present. Keep labelled checkbox selection and explicit Expand trace beside it;
never nest buttons, checkboxes or disclosures inside a native bar button. Use the
shared inspector rather than a second waterfall-only detail system. Standalone
consumers retain passive supplied-data disclosure and have no source authority.

Native buttons and checkboxes retain normal keyboard behavior, visible focus and
an accessible row description. Focus styling must not alter row dimensions or
move pointer targets. Avoid making an entire row an implicit navigation surface:
users need to copy text and inspect supplied metadata independently.

Pending selection/publication blocks presentation/field replacement rather than
silently dropping it. Page/display/kind replacement withdraws old hosts, and stale
reattached handles cannot publish, inspect a replacement row or query a source.
The public component's SSR and detached SPA properties remain supported without
fetches, navigation, global application callbacks or lifecycle publication.

## Acceptance

Use actual multi-service OTel spans with multiple traces, out-of-order parent/child
records, 5 ms versus 500 ms timing, duplicate parents, a missing parent and cycles.
Verify local kind changes, supplied inspection and field changes cause no read;
visually reordered Keep retains the exact original page occurrence; explicit Expand
trace is a separate read that can include another service. Preserve held data
through paging and kind through native save/template/reopen.

Check no nested interactive controls, keyboard inspection/selection/close focus,
qualified fields, useful compact/narrow rendering, source-free standalone SSR and
detached updates, and retired handles after kind/page/visit changes. Generated
properties and observed lifecycle/Bombadil checks supplement actual frames; none
is proof of complete upstream telemetry.

## Initial implementation source review

Reviewed the new supplied-page component and shared application surface before the
browser journey. The application uses the actual timing button as its existing
Inspect opener, so closing the inspector can return focus to the originating bar.
Selection and explicit Expand trace mount in adjacent hosts; no interactive
controls are placed inside the bar button. Source-position normalization keeps
visual parent-first order separate from selection and publication membership.

The passive component retains each supplied row, including unavailable timing,
with timestamp, fields and local Details. Compact trace headers disclose an
absolute UTC origin; one shared scale removes idle gaps between traces without
making different-duration fragments look equal. Service visibility affects the
bar label. The context option controls reported ancestry disclosure, while the
trace grouping key remains intrinsic to this display's domain rather than being a
second hidden context column.

Source review requested two small clarity refinements: truncated-value guidance
should use the Inspect action's name in both standalone and hosted modes, and a
known positive sub-microsecond duration must not share a rounded zero label with
an actual reported zero. Exact raw nanoseconds remain available in tooltips and
inspection. Actual compact/narrow rendering and pointer/keyboard outcomes remain
browser checks, not conclusions inferred from this source review.

The component owner implemented these refinements before source freeze. Positive
1–999 ns values now say `<0.001 ms`, including the shared-scale label; actual zero
remains `0.000 ms`. Exact tooltip values are retained. A single-observation group
no longer repeats an identical UTC timestamp in its header and row; the mandatory
row UTC remains visible. Multi-observation headers identify their earliest timed
supplied origin. Truncated fields direct users to Inspect. Fourteen focused
component/geometry tests pass; actual browser/frame review remains pending.

## Initial actual-frame findings

The first real native frames showed correct shared inspection and exact duration
labels, but exposed an embedding issue that source-only review did not establish:
application `section` card styling made every one-span trace group a large rounded
card. Three small observations occupied roughly 700 pixels, undermining the dense
record-scan purpose. Requested a component-scoped reset of group section margin,
padding, border, radius and background, retaining the group divider and existing
inspector chrome. This is a real nested-composition styling correction rather than
screenshot-only concealment.

Short trace labels also need both the beginning and end of the identifier to
separate structured-prefix IDs, while full IDs remain in accessible/tooltip text.
The final inspection frame should show the shared scale and both fast/slow timing
alongside the open inspector, using a bounded real clip rather than relying solely
on numerical geometry assertions. Final compact-frame review follows after those
changes; the initial frames are not a visual acceptance verdict.

## Final actual-frame review

Reviewed all three regenerated actual native frames after the component-scoped
styling correction. No visual UX blocker remains. Reopened and paged observations
now use compact headers and divider-separated timing rows, without nested rounded
cards or card padding. The default ancestry/action controls remain visible rather
than being hidden only for documentation. Short IDs expose both ends and remain
distinguishable; UTC times and human duration labels are readable.

The bounded inspection frame now shows the shared 0–500 ms scale, the 5 ms fast
span, its existing structured inspector and the 500 ms slow span together. Their
100-fold duration difference is visible under the same scale, while the earliest
timed supplied origin and parent-first display remain distinct from complete
trace latency. The inspector is the existing qualified attribute viewer; explicit
Expand trace stays a separate adjacent/footer action. Source row positions, not
visual order, remain the selection identity.

The native journey has also established the geometry ratio, shared Inspect opener
and no nested interactive controls; final keyboard/narrow-viewport and source/effect
accounting are recorded by the executable validation lane. This frame review is
not a screen-reader audit or proof of complete telemetry. No further source or
frame changes are requested by design.

## Proposed next outcome — 302

Recommend **save and customize an ad hoc record view without inventing a grouping
or carrying an unwanted aggregate**. This is a prerequisite adoption gap exposed
by 301, rather than another chart kind or template profile.

The shipped flat `Records only` layout is a visibility choice. Canonical
compositions still require aggregate plus records, aggregate declarations require
one grouping, and Customize sends ungrouped/multi-group questions into grouping
controls. Thus an ordinary log search or an exact trace can already be inspected,
but cannot naturally use/save the new record kinds without satisfying an unrelated
aggregate prerequisite. The trace dialog already has its own independent query and
finding capture; adding another trace query owner is unnecessary.

Bound 302 to a canonical one-records-child composition over the existing admitted
record-page profile, pager, Table/List/Span waterfall, shared interactions and saved
catalog. Keep the user's exact environment, predicates, window and query support;
never add dummy grouping or silently run a query. Acquisition should use the
existing count-plus-rows profile when applicable; do not claim statement savings
without validating the actual path. A complete row page remains distinct from
complete upstream telemetry.

First acceptance should start with ungrouped Activity logs and an exact-trace span
question: choose/customize record presentation, inspect/Keep, save/reopen and Run
without introducing a group or aggregate. Span-only and pending/stale-owner rules
remain. Ordinary saved-composition presentation reuse can follow the existing
source-free owner if its record-only branch is admitted explicitly; it must not
borrow source filters or masquerade as a destination-category template.

Keep category-bound authored templates, timeline/comparison children and general
parameter declarations outside this first record-only cut. A later natural
`Open as view` handoff from the trace inspector can use its existing exact trace
question and this saved-record foundation, rather than forcing a second query
model or expanding the 302 scope into a new inspector framework.
