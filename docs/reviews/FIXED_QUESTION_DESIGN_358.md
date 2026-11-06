# Evidence-first fixed questions — design review 358

Status: source, matched actual baseline/final frames and bounded interaction
acceptance complete. This is the selected consolidation after 356/357, not a new
navigation system or a broad visual redesign.

## User task and existing source

A user opens a package, shared question, service follow-up or reported edge, runs
it and reads its answer. The initial fixed question needs a compact inspectable
scope and intentional Run; after acquisition, actual evidence should precede an
expanded authoring workbench. Editing must remain available without becoming the
default wall before every result.

`ui/exploration.mjs` already has a question session, root-editor ownership,
Activity analysis Edit affordance, applied scope status and existing specialized
renderers. Before 358, general fixed entries exposed service shortcuts, filters,
grouping and measure controls above results. Consolidate the read/Edit presentation using
those owners; do not pretend this is rolling Activity or set a mode with new clock/
subscription effects merely to borrow its CSS.

## Read and Edit contract

- **Staged:** show the complete fixed scope in compact form, Run and Edit. Say it
  has not been read; an empty evidence region is not No data. Scope disclosure is
  local and does not run metadata or telemetry queries.
- **Applied:** keep current scope/uncertainty and the actual measured result first.
  Expanded authoring controls remain behind the existing edit interaction. Do not
  confuse a reusable package binding or source receipt with the result itself.
- **Editing:** reveal the same current controls, predicates, grouping and measures;
  preserve drafts, selection/caret/disclosure where existing owners support them.
  Editing and hiding controls cause no source read, discard or silent Apply.
- **Dirty/invalid:** retain visible unapplied/invalid state outside any collapsed
  editor. Invalid controls cannot disappear after validation failure or an automatic
  collapse. Reopening Edit restores their literal text; Run retains normal admission.
- **Running/failed/cancelled:** keep clear acquisition state, Cancel/recovery and
  original fixed bounds. Failure is not zero/No data. Do not hide the only usable
  recovery control inside an unrelated settings disclosure.

Presentation state is separate from question acquisition state and scope identity.
Run executes the staged fixed question; Use fresh window remains an intentional
clock change. Package lookback does not advance when switching Read/Edit. A new
entry still obeys existing dirty-destination acceptance and old-owner retirement.
No compatibility aliases or second universal state machine are needed.

## Focus, reuse and specialized data

Use one clear Edit/Hide edit action with expanded state/controlled region semantics.
Opening Edit moves focus only as an intentional user action to current controls;
hiding returns to the current visible edit opener only if the editor still owns
focus. Publication cannot steal focus from another task. Back restores the current
read/Edit context and held observation according to the existing owner contract.

Keep count, duration/status, numeric/histogram, grouped-time and reported-edge
profiles on their current renderers and exact data models. “Read first” must not
flatten them into a count table, hide unknown support or drop the edge definition.
Preserve/Save/reuse controls remain discoverable without recreating a sidebar.

356's saved catalog repeats row titles inside Inspect buttons and exposes nearby
authoring controls. Address that only where naturally sharing this read/Edit flow:
keep Open original question distinct from Inspect/Apply to the current destination,
retain compatibility/scope facts and metadata-only effects. No new catalogue or
independent layout project is selected.

## Measured native acceptance

Before edits, use genuine saved/shared fixed entry at actual scroll zero in
1280×1000 and 390×844 documents. Cover a duration/status/grouped-time profile and
representative service/reported-edge continuation. Record viewport/document
position of the **first substantive result**: a real chart, measure table/summary
or record item. Exclude scope/status count strings, navigation, loading text and
package binding from that metric. Count visible authoring controls separately.

Repeat the same entry and fixture after the change; do not compare manually
scrolled task crops or remove profile functionality to improve the number. Report
actual position gain and remaining below-fold content honestly, not a universal
compactness claim. Include staged no-read, applied evidence, keyboard Edit, a dirty/
invalid draft and return. Verify unchanged whole question/policy/window, source
budgets, independent receipts and zero IO for presentation/disclosure.

Mounted/model checks cover old hidden/detached callbacks, failed validation,
superseded/departed requests and repeated read/Edit/Back. Actual baseline/final
frames establish appearance; observed ledgers establish execution and ownership.
Finish this bounded cut before the planned whole-product checkpoint.

## Actual baseline and frozen source review

Viewed all four genuine old-layout baseline frames at original detail. Both package
and numeric entries show the applied scope followed by extensive authoring controls;
no substantive chart is visible in the initial desktop or narrow viewport. Native
scroll-zero measurements place the first package chart at y=1602.375/2543.719 and
numeric chart at y=1771.328/2718.938 (desktop/narrow). Visible in-viewport authoring
controls are 10/3, while ten controls remain CSS-rendered in each profile; closed
Details are excluded. These are baseline facts, not final improvement claims.

The frozen helper keeps presentation state local and delegates acquisition to the
existing Session. Untouched staged/applied questions can read compactly; actual
dirty drafts, local validation errors and failed/cancelled states force controls
open and disable Hide. It adopts no rolling/recent clock mode. Hidden/inert editor
events and stale edit-button leases are refused. Successful applied generations
collapse authoring; intentional Edit reveals/focuses current controls, with no
source effect. Source integration moves the existing workbench after result
regions and keeps specialized renderers intact. Source review accepts the bounded
consolidation; final frames and matched measurement are still pending.

## Final matched visual and interaction acceptance

Viewed all four final frames at original detail after the four baseline frames.
Run question, Edit question and Preserve are clear; applied typed scope, UTC bounds,
selector and package-binding meaning remain readable. Expanded service/filter/
measure authoring is absent from the read view, while the existing result actions
and specialized duration/grouped-time/numeric renderers retain their own meaning.
Narrow controls and exact scope wrap within the viewport. Final appearance passes
for the bounded read/Edit consolidation, not for a claim of universal compactness.

Matched actual target-plot positions are package y=1171.281/1653.531 and numeric
y=1340.234/1828.750 (desktop/narrow), compared with the baseline above. The gains are
431.094 px desktop and 890.188 px narrow for both profiles. CSS-rendered and
in-viewport authoring controls are now zero, versus ten rendered and 10/3 visible
before. **All target plots remain below the viewport fold.** These scroll-zero
frames show reduced authoring prominence, not an above-fold chart. Status/scope
counts and trace-example buttons are not substituted for the target plot measure.

Canonical native checks separately retain the complete question, eight record IDs,
profiles/bounds, receipts and saved bytes, plus fourteen interaction facts and the
actual hidden-Clear guard. Dirty/invalid states remain accessible, Edit causes no
read, and Source Cancel/Back keep their ownership contract. Actor/required-model
results are recorded by the validation owner; their existence does not change the
visual limit above. Bounded source, actual-frame and interaction acceptance passes.
Finish the planned whole-product checkpoint before selecting more layout work.
