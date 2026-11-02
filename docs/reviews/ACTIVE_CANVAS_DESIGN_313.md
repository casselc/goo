# One active investigation canvas — design review 313

Status: bounded implementation source and all three final actual frames accepted.
Read with [the live 312 whole-product findings](WHOLE_PRODUCT_DESIGN_312.md).

## User outcome and finite scope

After opening a population or contributed view, the responder sees that active
canvas's scope, data and actions. They should not edit a dormant original question
or decide which of two Run buttons applies. This cut consolidates visible control
ownership; it does not merge the underlying query owners or rebuild the interface.

Hide the dormant root Signal/Environment/Service editor, root Run question,
query-building/package controls, original result chrome and disabled Explore
linked population entry while the population canvas owns the visit. Preserve
those nodes/state for deliberate return. The active population retains its
current scope summary, primary Run, Preserve, View settings and useful data. Do
not add controls above the initial Activity stream.

## Return, focus and held context

Provide one obvious primary **Back to parent population** for a nested child or
the appropriate local return for a root population. A service or paused-Activity
origin may keep one secondary contextual return, named for that destination,
without presenting every historical Back at equal prominence. The target and
held scope should be understandable without knowing the editor hierarchy.

Back is navigation, not Run, Start or a query reset. Restore held parent data,
record/facet presentation, pending selection workflows and scroll/focus through
the existing owner. Preserve the dormant root's exact draft/applied state; hiding
it must not destroy controls, replace its query with the child or broaden its
environment/time. Do not focus a hidden root Run or detached opener. The visible
originating action or owner fallback should receive focus.

Use state/DOM structure rather than brittle text selectors to hide the dormant
surface. It should work for ordinary fixed, recent, saved and service-origin
population visits at 1280px and 390px. Restore root controls on the deliberate
existing exit, keeping current generation/source guarantees. Navigation must not
resume paused Activity or replay a Keep/capture intent.

## Literal titles and actual context

Authored title text remains literal. Do not guess whether a title is generated
from its wording, regex or similarity to a preset. If explicit declaration/default
metadata exists, a computed default may follow that contract. Otherwise retain
the title and show a compact **actual qualified grouping**/scope subtitle beside
it. This lets HTTP server routes remain a user's chosen title without suggesting
a service-grouped destination is secretly route-grouped.

The qualified grouping must come from the checked current question, not the
source template name or rendered row labels. Preserve field namespace and
missing/empty meaning. Avoid a new rename control or another title API in this
cut; one honest local context line is sufficient.

## Count comparison availability

When a selected cohort cannot use the count-only comparison because the current
question includes span status or another unsupported projection, show the exact
capability reason near the existing selection/actions. Do not rely solely on a
disabled button's title, which keyboard/touch users may never discover. Explain
what is unsupported without changing measures automatically or creating a new
Convert button. Existing scope/measure editing remains the explicit route.

Keep supported Count comparison, shared scales, exact operand roles and UTC read
clocks unchanged. This is clarity at the current action, not another comparison
profile or editor.

## Acceptance

Repeat service discovery → Customize → Run at desktop and narrow widths. Confirm
there is one active query editing/execution surface, useful facet/chart/records
are reached with less preceding chrome and the current qualified grouping is
visible even when an inherited authored title says something else.

Open a retained cohort child and trace; return through keyboard and pointer.
Verify the correct visible opener receives focus, held parent/scroll remains and
no source read/resume/publication occurs on Back. Return deliberately to the
original question and verify its controls/query state are restored intact. Include
saved, recent and paused-Activity origins rather than checking CSS in one demo.

Select a group with span-status projection and verify a visible precise reason
for disabled count comparison; switch measures explicitly through existing
controls and show the supported path. Preserve the literal authored title across
Save/export/reopen. Update actual compact desktop/narrow frames and existing
storyboards; source/lifecycle checks separately establish state and effect
correctness. No broad new review matrix or extra permanent evidence bundle is
required.

## Agreed implementation boundary

Frontend will suppress the exact dormant root intro/scope/workbench/status/plan/
Preserve nodes using the existing population visibility callback, preserving their
DOM and prior visibility. The disabled linked-population entry is hidden for all
active population visits, not only recent mode. Outer service/Activity/trail
returns keep their existing ownership.

For a root population, **Edit original question** is the preferred deliberate
local exit label: it describes the action rather than introducing another vague
Back to single question. Nested children retain Back to parent population. Exit
restores the retained root editor and visible focus/scroll anchor without a read
or new history entry. “Original” means that retained current question, never a
foreign source template's specimen query.

## Nested return and title source clarification

Reviewed all actual population mount sites. Recursive population children already
have an owning Back to parent population wrapper. The only separate explicit
`depth:1` mount is live-comparison interval inspection, whose wrapper provides
Back to live comparison. Its inner default exit would otherwise invalidate the
child without invoking that wrapper. Therefore hide the inner heading exit at
`depth > 0`, retaining the heading and owner wrapper Back. Root depth zero keeps
Edit original question. This removes duplicate/orphan exits without a new
navigation API or assumptions about service/saved roots.

For title rendering, keep explicit authored display text unchanged. An absent or
empty title may use the current checked measure plus grouping as the computed
default. Always show a compact qualified grouping/measure context under the
heading; namespace-qualified field text makes a retained title interpretable.
Do not infer title origin or rewrite old authored strings. The shipped descriptor
can omit a formerly computed default title without changing its human catalog
view title or inventing new metadata.

## Initial source acceptance

The primary implementation uses the existing visibility callback, exact dormant
root node set, CSS hiding and restored inert state. A revision-checked queued
focus transition avoids focusing a hidden or superseded visit. It preserves the
root editor DOM and uses the existing navigation/source owners. No new query
state or history layer is introduced.

The title implementation derives a default only when explicit text is empty and
always computes qualified grouping/measure context from the supplied checked
snapshot. Explicit strings remain literal; the shipped HTTP descriptor omits
only its old computed aggregate title. This matches the agreed title boundary.
Primary source UX has no blocker. Final native and actual-frame acceptance is
recorded below.

## Initial actual-frame assessment and remaining density

The actual desktop frame confirms the dormant root question editor and disabled
linked entry no longer precede the active canvas. The actual 390px nested trace
frame confirms no inner Edit original question competes with its owning Back to
parent population; the wrapper retains that navigation action. Qualified group /
measure context is visible with the declared display. The primary ownership cut
is accepted in these frames; final retained-root/native results are recorded below.

This is not a claim that density is solved. In the desktop frame a two-value
method facet, its actions, Find/context and nested cards occupy roughly 430px,
leaving the chart/records well below it. At 390px, contextual Activity/service/
trail navigation remains above the local child and consumes substantial height.
Those controls retain distinct existing owners by this cut's deliberate scope.
The page avoids horizontal overflow, but reaching data still requires scrolling.

Retain a finite rendering candidate for the later roadmap review: consolidate
passive facet row/action spacing and repeated context/card chrome using shared
components, preserving accessible targets, keyboard/scroll behavior, explicit
Read values and scope/support facts. This does not change the agreed next 314
source-demand priority or create a competing immediate queue. Do not expand 313
with unreviewed navigation-owner changes merely to improve a screenshot crop.

The earlier “one secondary contextual return” wording is the target experience,
not a delivered acceptance claim. The agreed implemented 313 boundary preserves
all existing outer Activity/service/trail returns. The mixed-origin narrow frame
still shows more than one contextual return, so that aspiration remains unmet
and is retained with density/context coalescing for follow-on review. No active
ancestry is hidden without a separately checked owner/navigation contract.


## Final 313 acceptance

All three final frames are accepted for the bounded ownership cut. The
[desktop canvas](../journeys/screenshots/313-desktop-active-canvas.png) now starts
at its heading and shows Edit original question, current fixed scope and only the
population Run/Preserve/settings controls before the declared facet. The dormant
root editor is absent. The [narrow trace child](../journeys/screenshots/313-narrow-trace-canvas.png)
has one local owning Back to parent population, without an inner Edit original
question, and retains its existing outer contextual ancestry. The
[restored root](../journeys/screenshots/313-restored-root-question.png) shows the
original query editor and its checked service/environment/window again after the
deliberate exit. This is restored query configuration, not an automatic read.

The verification owner reports the final native journey green in 46.7 seconds
(50.1 seconds including runner), with retained query/hash/history, visible focus,
scroll restoration within 16px, zero navigation IO, narrow no-overflow and actual
Close/history trace-read cancellation. Existing source and native assertions
cover the title/context and visible count-capability boundaries separately from
these broad canvas crops. No new source/UI edits were requested after this review.

No blocker remains for 313. Residual facet density and multiple outer contextual
returns remain explicitly unresolved as described above; acceptance must not be
read as whole-product density/navigation completion. The next 314 source-demand
work remains the active priority, with these rendering observations retained for
its subsequent roadmap review.
