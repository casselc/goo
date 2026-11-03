# Compact data-first canvas — design review 315

Status: bounded source/layout improvement accepted with measured before/after frames; narrow chart-data visibility remains partial.
Read with [post-source product priorities](POST_SOURCE_PRODUCT_314.md) and
[313 accepted ownership limits](ACTIVE_CANVAS_DESIGN_313.md).

## User outcome

A responder can see and use the contributed facet and reach its chart/records
without first scrolling through sparse controls and repeated context. This is a
shared rendering consolidation, not a new editor, source owner or data-demand
change. Keep the source savings from 314 separate from layout measurements.

Use the same real two-value HTTP method facet and nested trace task at 1280×900
and 390×844 before and after. Previous 313 frames showed roughly 430px for two
facet values and ~400px of narrow outer context. Record actual rectangles before
editing; those are baselines, not desired fixed heights or universal budgets.

## Compact shared facet

Use component-owned styling rather than broad product button/margin overrides:

- One wrapping header contains title and existing Read values / Clear staged facet
  actions. Hide no required execution boundary and add no new action.
- One concise context area retains qualified field, current applied scope,
  matching full-record count and unknown upstream coverage. Wrap honestly on narrow
  screens; do not describe these as self-excluded/global values.
- A compact tools row contains Find loaded values, a live **N of M loaded values**
  result and existing Read details disclosure. Exact independent receipts remain
  available there; they need not occupy the default surface.
- Rows align value, count and Include/Exclude without inheriting large generic
  `.controls` margins. Missing key and Present empty string remain distinct and
  long values can wrap/scroll within the established loaded-value viewport.

Find is local presentation only. Announce zero loaded matches as zero matches,
not no telemetry. Preserve the original matching-population denominator and
loaded-value total separately. No reads, query mutations or publication occur
on mount, Find, layout or readiness update.

## Interaction and target sizes

Keep native button semantics, visible focus, field-qualified Include/Exclude
names and logical Read/Clear focus across loading and rerenders. Preserve loaded
Find while the same basis remains; changed data/scope follows the existing reset
contract. Compactness must not steal focus or hide an active staged intent.

WCAG 2.2's minimum pointer target criterion is 24 CSS pixels, with specified
spacing/exceptions; it is not a blanket 44px requirement. For this dense control,
use at least 32px action height for ordinary fine-pointer layouts and prefer
44px coarse-pointer targets where possible. Verify actual narrow controls and
separation instead of merely shrinking font/padding until a screenshot fits.
Keyboard/touch usability and semantic facts take precedence over a numeric height
goal. Do not force fixed-height clipping of titles, errors or long values.

A normal two-short-value ready facet around ≤220px on desktop is a useful proposed
target. Narrow/coarse-pointer layouts may reasonably be taller (roughly ≤280px
as a starting comparison), with a documented reason. The outcome is materially
less preceding chrome and reachable relevant data, not simultaneous visibility
of every chart and record in 844px.

## Compact context without changing ancestry

Keep each named return bound to its actual owner: paused Activity anchor, service
resource return, trail history and deepest retained parent. The deepest child Back
is primary; outer returns can use a compact secondary context arrangement. Reduce
padding, duplicated headings/lines and card layering before hiding any destination.
Do not create a universal Back or merge history visits with population parents.

Keep checked service/environment/fixed bounds available and visually intelligible.
Do not move source execution into breadcrumbs or relabel a new read as retained
context. Preserve paused parents, current pending workflows, visible focus and
scroll restoration, with no I/O on navigation. Ancestor-hiding cannot bypass
actual visibility checks for an old clickable control; prefer keeping the existing
returns visible in a denser arrangement for this cut.

## Acceptance and finite scope

Compare actual before/after rectangles and frames for the same facet/context task.
Show the full two-value facet, source facts and meaningful chart/record content
so a smaller frame crop cannot masquerade as a smaller component. At 390px
verify no document overflow, long-value readability, target dimensions, keyboard
focus and loaded-match/zero announcement. At desktop verify data moves upward
without an extra collapsed editor hiding the useful controls.

Repeat selection staging, origin-safe Clear, explicit values read, retained child
Back and paused-Activity return; requests and saved/query bytes must stay the
same. Keep existing sampled/current/captured uncertainty visible. No new component
language, profile, registry, response cache or source-demand optimization is
needed for this finite cut. Any long-value/DOM-budget work should follow the
measured legitimate case and remain one shared boundary, not an independent
performance framework.

## Agreed context consolidation boundary

Retain outer Activity and trail actions as compact wrapping secondary context,
without changing labels/owners or hiding an active ancestor. The resource-return
bar keeps its named Back and literal origin title visible. Move its existing
origin environment/full UTC bounds into native **Origin scope** details, clearly
labelled as origin rather than current-child scope. The active canvas still has
its own current environment/window summary. Add actual action-visibility checks
where an old hidden-ancestor resource-return control otherwise remains callable.

The population header may align scope and tools and disclose exact provenance /
read details. Keep staged, loading, error and readiness facts visibly actionable;
no collapsed disclosure may conceal why Run is needed or unavailable. Show one
readable current environment/window context rather than repeat full timestamps
in several surrounding cards. Keep the named return and scope source owners
separate even when their presentation becomes denser.

## Measured before state

The native before run completed in 17.7 seconds (20.4 including runner). Actual
rectangles at the same canvas anchor establish: desktop facet **406.3px** high,
two **75px** rows and chart top **1098px**; narrow facet **448.3px** high and
chart top **1492px** in an 844px viewport. The narrow facet's usable width is
only 272px inside nested wrappers, increasing line wrapping. The earlier ~430px
review estimate is superseded by these task-specific measurements.

The actual narrow-before frame confirms the stack: population status/actions /
Preserve/settings, then nested facet title, qualified scope, actions, count facts,
Read details, Find and rows. Both component spacing and surrounding card/context
chrome therefore need the bounded consolidation; changing only font size or
cropping the screenshot would not address the task. Before assertions retain
named returns and establish zero-I/O local returns and held facet state.

## Combined source review

The frozen facet source matches the contract: component-owned wrapping header,
qualified/applied-scope and full-count/unknown-coverage facts, local Find with a
polite loaded-match status, inline native receipt disclosure, compact rows and
32px controls with coarse-pointer enlargement. Same-read Find, disclosure state
and logical focus are retained. No source/model/transport changes are introduced.

The resource bar retains named Back and literal origin identity with native
Origin scope details; actual visibility/inert checks retire inaccessible actions.
The product canvas keeps staged/readiness facts visible, shows one compact current
UTC range with full exact half-open bounds in its title, removes repeated card
chrome and keeps named Activity/trail/local returns. Literal ancestry titles may
ellipsis with exact title available; they are not inferred or relabelled.
Combined source UX has no blocker. After frames must establish the actual
one-page result, not merely a smaller isolated facet rectangle.

The before ancestry measurement covers trail (~71px), local Back (~30px) and
child descriptor (~55px), about 156px in that measured subset. It does not measure
the earlier ~400px outer-context estimate as one comparable total. Do not claim a
50% reduction of that larger total from incomplete measurements. Preserve full
child pause/independent-read meaning where that reasonably limits compression.

## First after-frame assessment

The first measured after run improves the two-value facet from 406.3px to
197.7px desktop and 448.3px to 216.5px narrow. The desktop chart starts at 667px
and its meaningful bars are visible within the 900px viewport. Field/scope/count /
coverage facts and explicit Read/Include/Exclude remain legible; local loaded
matches and receipts are distinct. The native owner reports keyboard/disclosure /
Find and 44px coarse-pointer facet targets passing.

The same first narrow frame still places the root chart at 948px. Its selected
child chart begins around 833.6px in an 844px viewport: a ten-pixel intersection
is **not** meaningful chart visibility. These frames establish partial improvement,
not completed data-first narrow acceptance. Existing group tools/selector layout
and missed trail-wrapper selectors are legitimate remaining consolidation in this
same finite cut.

The agreed final context refinement keeps Back plus native **Parent paused**
details in the child's owner row; the full original independent-read/retained
parent description remains inside. A literal meaningful navigation title may
replace generic Population workspace, without changing query/history ownership.
Keep current staged/readiness/independent-read status visible. This is preferable
to silently hiding ancestry or adding a universal Back. Coarse-pointer selection
tray actions must retain the same enlarged target treatment as the facet.


## Final 315 actual acceptance and limits

All four final after frames are accepted as a bounded density improvement. The
[desktop](../journeys/screenshots/315-after-desktop-facet.png) retains title,
qualified scope, full count, unknown coverage, explicit read/stage controls and
loaded-match denominator in a **197.7px** facet (before **406.3px**). Its chart
container starts at **701.3px**, and the first actual count/bar is visible in the
900px viewport. The [narrow root](../journeys/screenshots/315-after-narrow-facet.png)
facet is **216.5px** (before **448.3px**); both method rows and their counts/actions
fit in the viewport. Its chart container still starts at **867.2px**, 23.2px below
the 844px viewport, so this is not full narrow chart-first completion.

The [narrow cohort](../journeys/screenshots/315-after-narrow-cohort.png) keeps
Back to parent population plus Parent paused details, a meaningful Selected
population heading, the visible applied/independent-read state, seven-record
count and contributed control. The chart container starts at **703.1px**, yielding
**140.9px** intersection. Visual inspection confirms that intersection contains
heading/read/scope facts, **not the actual plotted count/bar row**. Report it as
chart-container visibility only, not meaningful chart-data visibility. The selected
count and facet context are visible earlier, but the remaining chart chrome is
still a real limitation. No further CSS expansion is requested in this cut.

The [narrow trace](../journeys/screenshots/315-after-narrow-trace.png) retains
named parent return, Parent paused details and literal Exact trace view purpose,
with current scope/readiness, UTC/page/coverage facts and the existing Waterfall.
Its record container begins around **394.5px**. Context compression preserves the
actual owner meanings instead of inventing a universal Back.

The measured trail/child subset improves from roughly **156px to 73px**. This does
not measure or prove halving the former ~400px whole-origin estimate. Literal
origin titles, detailed fixed scope and existing named callbacks remain available.
The final native verification owner reports green in **27.3 seconds** (**31.3**
including runner), with keyboard/native disclosures, loaded Find/zero status,
zero-I/O local updates/returns, held read restoration and actual **44px** coarse
Include and Open-child targets. Final Read details is at least 32px fine / 44px
coarse; no model/source-demand changes were introduced.

No blocker remains for accepting this finite improvement. Keep narrow chart-data
visibility and remaining read/plot chrome explicit in the follow-on review; do not
claim all first-viewport or ancestry-density goals were met, and do not continue
indefinitely shrinking targets or hiding interpretation facts to meet a pixel
threshold. The next roadmap item should follow the agreed core-flow priority,
with these concrete residuals retained rather than reopening a broad polish loop.
