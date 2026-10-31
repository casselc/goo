# Shared record interactions design review — 298

2026-11-01. Source/design review of the shared Table/List interaction work.
Actual browser outcomes are recorded separately in VALIDATION.md; this document
does not claim keyboard or screen-reader testing from source inspection.

## Product decision

Table and List are presentations of the same bounded records child. Both should
support explicit selection, supplied-record inspection, shared-field refinement,
Keep and reported trace/log context through one application owner. A kind change
must not silently discard a selected/submitted set or alter the held aggregate,
record page, scope or evidence basis. Page indexes identify observations only
within that supplied read; they are not durable IDs or proof of unique events.

Do not add a second investigation panel, a List-specific Keep flow, implicit row
selection, card navigation or an app callback invoked by custom-element lifecycle.
Native controls retain their normal keyboard behavior: checkbox selection and
buttons for explicit actions. Opening supplied details is local; expanding a trace
or finding surrounding logs is a fresh contextual read and retains its existing
scope/coverage explanation.

## Component contract

`goo-record-list.recordItems()` exposes a frozen array of current-render host
handles: `{index, element, actionsHost, detailHost, details}`. The index is relative
to the supplied page. Both host slots are empty and hidden by default. The passive
native Details disclosure remains available for standalone consumers. The List
imports no application controller, invokes no callbacks, emits no action requests
and performs no source reads or publication.

An application explicitly mounts its shared owner into these slots and hides the
passive Details disclosure while the shared inspector is available. Thus the app
shows one Inspect affordance, while a standalone List still has one local Details
affordance. Explicit host access renders current supplied inputs even while detached, then
returns only their current slots. Connection preserves those exact slots rather
than replacing controls already mounted during detached app construction. A
page/display replacement recreates slots and retires access to old hosts. Returned DOM handles convey no action authority:
the application must independently enforce owner identity, current page and epoch,
connectivity, pending/submission state and permitted scope before every action.

## Density and accessibility

Keep rows as divider-separated telemetry, not large cards. UTC/service, summary,
qualified fields and optional trace context retain their existing compact order.
Put the shared action strip below the summary/fields; use wrapping, small native
buttons and a properly labelled checkbox. Never make the entire row clickable:
users must be able to select/copy text, use Details, and activate controls without
unintended navigation. A narrow focus-within accent marks the active row without
changing its selected meaning; selection remains the checkbox's explicit state.
The inspector occupies normal document flow and wraps rich content.

The old “Use Table for selection and source actions” explanation should disappear
once parity is implemented. Do not replace it with repeated per-row instructional
prose. Keep the common field picker and outer pager as the primary controls.

## Acceptance checks

Exercise both log/span Lists: select/clear, shared fields, Keep, supplied Inspect,
Expand trace and Surrounding logs where applicable. Confirm actions are absent
from a default standalone component and direct detached property/SSR usage works.
Test selection/submission and pending-page kind guards, stale page/control
reattachment, current Back/revisit controls, keyboard focus restoration, narrow
width wrapping and the held aggregate's independence from a fresh page read.
Refresh actual screenshots/storyboards after executing the complete flow.

## Component implementation checks

The focused passive component suite passes six tests, including an actual render
with an inert DOM harness. It verifies frozen page-relative host handles, empty
hidden slots without click callbacks, detached page/display replacement recreating
current handles, connection retaining explicitly mounted controls, and invalid
replacement clearing all handles. Existing tests retain
UTC/qualified field state, public SSR/property admission and bounded ancestry
projection coverage. This is component validation, not the full app parity verdict.


## Shared owner source review

The extracted application owner mounts the same selection/refinement/Keep/Inspect
and contextual read controls into both presentations. List's passive Details is
hidden while the shared inspector is installed, leaving one clear Inspect action.
Action controls remain outside optional trace-context presentation, so hiding
reported ancestry does not silently remove selection or inspection capability.
Native checkbox names identify the page-local observation; the shared selection
tray holds publication and shared-field controls once per selection rather than
repeating them per item. The Table-only explanatory note is removed. Source
review requested resetting Inspect's pressed state when its inspector closes;
actual focus restoration and narrow viewport appearance remain browser checks.

The browser pass caught a pointer-selection defect in the new focus styling:
adding border/padding on `:focus-within` moved a small checkbox during mousedown,
so a second row could receive focus without toggling. The accent now uses an
inset box shadow and changes no geometry. Pointer and keyboard selection must be
rerun against that corrected stylesheet; visual focus must never shift its target.

## Actual frame and interaction review

Reviewed `298-list-inspect.png` and `298-list-keep.png` generated from the running
app. The scan row stays compact: UTC and qualified service lead, span summary and
duration follow, optional reported context remains collapsed, and the action strip
contains one Inspect plus explicit trace expansion. The opened inspector reuses
the structured attribute viewer with UTC, field scopes and bounded scrolling.
There is no competing passive Details disclosure or extra List-specific inspector.

The selected-record tray appears once above the page. It names two page
observations, keeps shared fields/provenance collapsed, and displays the actual
“Kept 2 observations” outcome with Open saved. It does not label those two rows as
a population total. The executing verifier reports keyboard selection of the first
checkbox and pointer selection of the second now succeed after the no-shift focus
fix; keyboard Inspect/Close focus and pressed-state assertions also pass. No
remaining UX blocker found in this bounded Table/List parity flow.
