# Typed package scope and intentional execution — design review 350

Status: implemented; source and actual-frame review accepted. Execution gates are
recorded separately in [validation](../../VALIDATION.md).

A user discovers useful OTel telemetry, turns that question into a reusable package,
and later applies it intentionally. Typed policy scopes should work throughout the
existing authoring path without being weakened to a scalar environment. This is
parity and simplification of the current editor/compiler, not another grammar,
package catalogue or WYSIWYG replacement.

## Existing path and ambiguity

`ui/exploration.mjs` currently refuses Create package for typed selectors. Package
summaries and both libraries assume scalar `source.environment`.
`ui/telemetry-packages.mjs` already separates build/validation from explicit source
preview and immutable save, but Use in Explore dispatches a package event whose
Explore consumer immediately runs telemetry. Its workflow variant says Run fixed
question, while library Explore uses the same implicit path. The label/effect
contract needs closing before implementation.

## Recommended natural sequence

Create from applied question → edit/build → validate → optional explicit preview →
save/open → **Stage in Explore** → existing explicit **Run**.

Stage in Explore plans/resolves the supported definition and opens the existing
question without reading telemetry. A fixed workflow variant says Stage fixed
question in Explore. Library actions should use the same clear staging meaning,
not an Explore label that unexpectedly executes. If an existing separately labelled
Run package gesture is intentionally retained, it must carry explicit execution
intent; an ambiguous shared event must not decide source effects implicitly.
Prefer one staged handoff and remove the surprise rather than add more controls.

Opening, importing, building, validating, exporting, generating the optional Go
template and saving a revision perform no telemetry reads. Preview remains an
explicit read and says which window it uses. Save stores configuration, not that
preview or a guarantee of instrumentation completeness.

## Typed identity and time meaning

Preserve the complete admitted environment selector through draft, YAML edit,
validation/plan, immutable save/reopen, library search/summary and Explore handoff.
Use the existing environment presenter for current-only, effective precedence,
empty and unreported scopes. Do not display a blank scalar as “all” or replace an
effective policy with a current-key-only name. Namespace/service filters and all
qualified predicates retain their exact meaning; unsupported profiles refuse
explicitly before IO.

The captured question's fixed bounds are preview context. The reusable definition
stores its declared lookback, not an implicit frozen copy of those bounds. An
original-window preview remains labelled fixed; opening a saved relative package
resolves a new fixed question at the declared planning instant. Stage shows those
actual UTC bounds. Run executes that staged fixed question; it does not secretly
move it again unless an existing explicit fresh-window gesture is chosen. Fixed
workflow packages retain their fixed-preview contract rather than pretending to
be live streams. Independent preview/read clocks must not be called save time.

## Optional instrumentation and reuse

Keep the existing reviewable Go annotation template behind its optional disclosure.
It is a handoff, not proof of inferred functions, outcome policies, deployed
instrumentation or complete telemetry. Typed policy support in the query package
does not imply Go annotations encode deployment identity. Non-Go applications can
use the saved package directly over ordinary OTel.

Editing a saved revision creates a new revision. Library selection/open cannot
replace a dirty editor or current Explore task through a late response. Current
controls, cancel/Close and route guards own publication; old detached Use/Stage
handlers cannot read or retarget a newer context. Return to the origin should keep
its held observations and policy, not replay source acquisition.

## Acceptance

Use genuine typed effective/current-only/empty/unreported questions through the
existing Create package path, with one ordinary named/namespace example in native
UI and semantic cases for all policy forms. Build, validate, Save, open and Stage
must have explicit zero-telemetry budgets; Preview and Run have distinct advertised
reads. Inspect YAML, immutable saved definition, plan echo and staged fixed question
for exact policy/predicate equality. Test both original fixed preview and newly
resolved lookback bounds without conflating clocks.

Actual desktop/narrow frames should show policy labels, readable YAML/editor,
clear preview/save/stage actions and the staged Explore Run state. Verify source
refusal, unsupported profiles and late/departed owner negatives without adding
verbose explanatory panels. Do not claim general metric package support beyond
the profiles admitted by the current compiler.

## Frozen candidate source acceptance

The candidate uses one staged handoff across editor, both libraries and examples;
there is no implicit telemetry Run in the package apply consumer. Draft/workflow
entry says Stage fixed question in Explore and retains its original preview
bounds. Saved/imported relative definitions resolve a fresh plan at staging, then
Run executes those held fixed bounds. Build/validation/open/save remain metadata
or pure planning; Preview is the distinct explicit read.

`ui/telemetry-package-intent.mjs` checks the closed source scope union, complete
question/profile equality, selected revision and relative width or original fixed
question as applicable. Existing duration/status/numeric/histogram/grouped-time
profile fields remain admitted; the record-count package measure is not evidence
that these profiles disappeared. Shared environment labels carry typed policy
through editor and libraries instead of showing an empty scalar.

The optional Go template remains restricted to current-key-only named operation
contexts. Typed query support does not claim equivalent generated deployment/Go
operation semantics. The UI makes that handoff limitation distinct from using a
package directly over non-Go OTel. Source review accepts the bounded integration;
actual frame and end-to-end fixed/relative timing acceptance remain pending.

## First actual frame review

Reviewed all four initial 350 frames: editor, freshly staged question, applied
package result and narrow typed package. The editor clearly retains effective
current/legacy policy in YAML/summary, fixed planned-preview UTC bounds, five spans,
recorded status and 14 ms mean. Stage fixed question is distinct from Preview/Save.
Fresh staging shows a separate resolved fixed window with Run question; the applied
frame then shows the same fixed question's evidence.

A real copy defect is visible: the static package-binding note still says “fixed
question staged; Run reads this window” after the dynamic question has become
Applied. Its 64-character revision also dominates the narrow binding paragraph.
Recommend neutral binding copy containing title, short revision and “Changes here
do not edit the saved definition”, retaining the full revision in supplementary
metadata/title. Dynamic question status owns staged/applied/read meaning. This is
a minimal copy correction, not a new state control or layout redesign. Final
visual acceptance was withheld until the corrected actual frames.

## Final actual-frame acceptance

The regenerated fixed-preview editor, fresh staging, applied-result and narrow frames
retain the checked effective policy, literal predicates and UTC bounds. The
phase-neutral origin note shows a short revision and the full value in its title;
it no longer contradicts Applied status. The five-record profile, two recorded
errors and 14ms mean remain readable. These are deliberately scrolled task views,
not a claim about initial Activity-entry density. No further visual blocker remains.
