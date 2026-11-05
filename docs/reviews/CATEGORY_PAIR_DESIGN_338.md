# Category population comparison design — 338

339 compacts the page-level scope/control presentation and refreshes maintained
frames; this review records the original 338 acceptance. See the
[current presentation contract](../POPULATION_CHROME.md) for that consolidation.

Status: final surface source and five actual viewport frames accepted; native,
model and actor checks pass. This review directly inspects source and frames;
execution results are reported by the native owner and independently reviewed by
the root agent.

## Investigator task

After noticing a count change, identify which routes, tenants or other categorical
values explain the changed population. Compare the full observed populations,
including values present only in the baseline, then restrict both original
questions through the existing **Filter both windows** continuation.

Keep the categorical facet in its current position. Do not introduce a separate
comparison panel, duplicate Summary/Groups selector, filter language or wizard.
The current saved facet declaration already names the qualified field; runtime
operands and observations belong to the application owner.

## Compact presentation

Use one **Read values** action to explicitly acquire the two counted partitions.
The facet header identifies **Selected / Previous window values**, or
**Selected / Fixed baseline values** for a nonadjacent baseline. Reuse the existing
facet title, Find, Clear staged and Read details disclosure.

Each union row contains its literal category label, two aligned count cells with
compact bars, and one Include/Exclude pair. The column/series labels retain the
same operand order and names throughout. Reuse `comparison-visual.cell` and its
shared count scale rather than copying a chart or embedding the whole
`goo-group-count-comparison` table: that table brings a 580px minimum width and
another set of presentation/scope controls.

The bar maximum comes from all admitted union rows, not the filtered or visible
subset. Exact integer labels remain the authority; no scalar float coercion,
percentage-change infinity, rates, health classification or causal claim is
needed. Use a deterministic maximum-reported-count order, with exact literal identity
as tie-breaker; no added order selector is needed for this cut. An incomplete
pair must not be ranked by an invented difference. Keep Find local and count
its matches against the full loaded union, including baseline-only values.

Missing key and Present empty string remain separate rows; a literal string such
as `"Missing key"` is not the missing category. Long labels have a bounded visual
preview and accessible/full literal text. A baseline-only category still offers
Include/Exclude; those actions stage the same checked categorical intent against
both original predicates.

At narrow widths, keep both count cells together, with their short operand labels
available without a tooltip. Category text and the shared actions may use a
second line; do not require horizontal scrolling to compare the two numbers or
place each operand in a separate card. Bars supplement aligned labels and never
encode distinctions solely by color. Read/action controls retain keyboard names,
focus return and the existing coarse-pointer target minimum.

## Certain accounting, uncertain telemetry

The current endpoint refuses more than 1,000 groups rather than returning a
silently capped list. Its ready result requires unique positive groups whose sum
matches the independent count read; `changed_between_reads` is refused by the UI.
An absent category in such a ready partition may display **0**, meaning no matching
observed records in that returned partition. It does not prove complete upstream
telemetry. Retain one compact **Upstream coverage unknown** notice.

Unloaded, loading, failed, refused or mismatched operands display an unavailable
count, not zero. Keep an independently successful side visible if the other fails;
show which operand needs retry in the existing status area. Do not compute a
signed change, comparative bar geometry or difference ranking until both sides
are admitted. A retry must be explicit, preserve the healthy witness and make
its older independent read clock discoverable. Prefer the existing read command
with an accurate retry label over another permanent toolbar.

Read details groups each operand's exact UTC interval, environment/policy and its
count/group read receipts. Neither side inherits the other side's clock. A
zero-record ready operand is valid: its categories are all aligned zeros against
the other's observed rows, without a missing-data alarm.

## Interaction and acceptance

- Opening/rendering, Find, Include/Exclude, Undo/Clear and staging read nothing.
  Parent Run and retained handlers cannot discard pending paired choices.
- One Read values command reads the two original definitions independently;
  baseline-only categories, missing/empty values and asymmetric filters survive.
- Partial failure retains the healthy side and identifies the failed operand;
  explicit retry and late-result retirement cannot restore an obsolete pair.
- Filter both windows stages a child; child Run obtains new observations. Back
  restores the retained facet union, local Find, presentation and action focus
  without rereading. Old detached/reconnected controls remain inert.
- Save/export/reopen retain definitions and the facet, never read values, staged
  choices or acquired observations.
- Review natural desktop and 390px viewports with both ready sides, a baseline-only
  category, a partial failure and a pending common filter. Do not crop away verbose
  controls or use actor captures as visual proof without inspecting them.

This design builds on [337](PAIRED_FACET_DESIGN_337.md), the
[roadmap](../ROADMAP_2026_11_05.md) and the source-grounded
[architecture contract](CATEGORY_PAIR_ARCHITECTURE_338.md). The final source and real journey satisfy this interaction contract; the
validation record separately owns detailed execution counts and limits.

## Implementation review findings

The paired modifier must preserve the shared facet-list `[hidden]` contract.
A later equally specific `display:grid` rule can make filtered rows visible even
when the DOM says hidden; acceptance must inspect actual visibility, not only a
match counter or hidden property.

A successful zero-record operand is still a ready witness if its sibling fails
or loads. Show that side's zero total and receipts, with the opposite unavailable;
do not gate partial rendering on a nonempty union. Both source findings are corrected: paired hidden rows receive an explicit
scoped display rule, and any ready operand enables partial rendering, including
zero-record witnesses. Focused/native acceptance remains separate.

Final source review accepts the shared count-cell geometry, explicit partial
unknown states, healthy zero totals and receipts, baseline-only actions and
retired action epochs. The mounted tests separately exercise these cases. Actual
Find visibility and narrow comparability still require the native viewport pass.

The paired numeric wrappers now use a named group role, so their operand/category
labels are valid accessible names rather than labels on generic divs. Mounted
checks assert those roles and names. This does not add a visual control.

## Proposed next consolidation checkpoint — 339

Proposed, not implemented or accepted by this review. After the 338 viewport
review, assess one bounded cleanup of fixed-population scope/action chrome.
Environment policy, exact operand windows and read facts currently appear in
several places. Prefer a compact shared summary with an existing disclosure for
full facts, preserving distinct held comparison and newer facet witnesses.

The target is less repeated chrome before the useful population data, with the
same explicit Run, staging, return and source-read boundaries. Reuse presentation
helpers; do not introduce a universal scope manager, combine separate source
models or widen the chart catalog. Measure the natural initial viewport before
and after, including keyboard/disclosure focus and narrow layouts. The 338
screenshots alone may be task-scrolled viewports and cannot establish entry-height
savings. Final task selection follows actual frames and architecture review.

Architecture refinement: consolidate duplicate page-level definition/window/policy
chrome, not the component-specific observed receipts. Embedded count-pair and
facet components must remain self-contained. Existing `populationWindowText` and
formatting helpers are preferable to a new universal scope component. A secondary
candidate is an explicit geometry-disabled option on shared count cells: preserve
a known numeric value while suppressing incomplete-pair bars, rather than passing
null as the value and overriding the visible text. Select this only with a clear
reuse benefit; it is not another 338 capability or acceptance blocker.

## Actual viewport review

The four generated 1280×1000 desktop frames are independently inspected. The
initial union shows all five rows with aligned counts and a shared scale:
GET 75/5, POST 27/4, Missing key 3/1, Present empty string 2/1 and DELETE 0/2.
The pending frame keeps Filter both windows primary and visibly disables Run.
Reopen shows 27/4 held comparison facts while the facet is unloaded. The planted
partial read retains selected 27, an unavailable previous-window cell and an
explicit previous-window Retry; it does not draw misleading comparison bars.

The initial/pending frames reflect natural task scrolling with a clipped Explore
heading, not pristine entry screenshots. Repeated page-level environment/window
facts substantiate the proposed consolidation target, but the wrong-field advice
line was produced by a preceding refused gesture and is not default entry text.
The full-width Groups selector in the reopened frame is another existing compact
control/scoping opportunity: the page's descendant select rule reaches reusable
comparison controls. Preserve component-specific independent receipts when
addressing either issue. The 390×844 frame is also inspected: both count labels
and numbers remain
comparable without horizontal overflow, with the shared actions beside them.
Four full category rows and the DELETE label fit; the fifth category counts are
below the viewport. This is a task-scrolled view, not evidence that all categories
fit at entry. There is no visual blocker; the final native, observed-model and
actor checks also pass.

Final completion is confirmed: the 4.2-minute native journey passes for ordinary
logs and spans, including baseline-only selection, the failed-side retry and
held-read retirement/reacquisition. The observed 21-step model passes. The
35-second actor reaches all eight phases with four verified transitions, six
source requests including two value lookups, two comparisons, zero writes,
immutable retained facts and no overflow across 461 snapshots. These results do
not substitute for the five independently inspected Playwright frames; actor
capture quality remains a separate test-driver question. The proposed 339
consolidation above remains future work.
