# Authored cohort comparison continuity — design review 307

Status: implemented source and refreshed actual frames accepted.
Read with [306 whole-flow continuation](VIEW_ARTIFACT_CLI_DESIGN_306.md) and the
existing population/comparison contracts.

## User outcome

An investigator sees a distinct group in a saved/team breakdown and asks whether
it differs from the parent population or everything else inside that parent.
Using an authored view should not remove the native breakdown's cohort tools.
This cut connects those tools to eligible ordinary count displays without adding
another comparison workspace, source owner or capture requirement.

## One existing selection flow

Use the existing Bars selection or Table Select gesture, then **Stage selected
group** and the existing population selection tray. The first gesture highlights
supplied aggregate data; staging records the typed facet intent without reading.
The investigator then chooses the tray's existing comparison actions. Do not add
a second per-chart row of Compare buttons or force a trip to a comparison lab.
Pointer and keyboard selection must lead to the same tray and availability.

Keep the two meanings explicit:

- **Compare selection to current** compares the selected population with the held whole
  parent. Selected observations are part of that parent; the operands overlap.
- **Compare with rest** compares the selected population with the complement of
  that selection inside the original parent. Other parent predicates remain;
  “rest” never means all telemetry outside the original question.

Use concise child headings such as Selected vs held parent and Selected vs rest
of parent. Put the overlap fact in the operand summary, not in a warning dialog.
A selection tray can contain several category values; do not pretend one clicked
record or the visible page defines a population. Missing, present-empty and
named values remain distinguishable throughout partitioning and labels.

## Retained child and explicit execution

Open the existing retained population/comparison child rather than replacing the
parent's authored composition. The new authored continuation stages configuration
and focuses the existing Run. It reads nothing until Run. Existing native actions
with different established execution timing should retain their documented
behavior; do not introduce a global semantics change solely for this cut.

For selected versus parent, the baseline is the actual held full aggregate of the
parent, with its recorded basis/time. Run obtains the selected population through
the existing owner; it must not silently reacquire the parent and call that held.
For selected versus rest, Run obtains both checked fixed questions independently.
The display must retain independent read/coverage semantics and avoid claiming an
atomic partition observed at one source instant.

Both operations retain the exact checked environment, fixed bounds and original
predicates. Only the typed selection and its checked complement change. The
inspection page, count-only record response or local Find-filtered group display
cannot supply a full-population baseline. Unsupported predicates, exhausted
budgets, pending selection/Keep work or depth limits should explain unavailable
actions through existing status/availability conventions.

Back returns to the held authored parent and the visible originating selection
control/tray, with its composition, record kind, fields and scope intact. It must
not Run, resume live reads or replay a Keep. No additional capture is made when
opening or returning.

## Comparable visuals and interpretation

Reuse the existing group comparison display, with clearly named operands,
counts and a shared scale. The user should see a rare selected group rather than
have it disappear below a threshold. Absolute count and support must remain
visible; percentages must identify the denominator. Selected-versus-parent shares
cannot imply disjoint classes. Missing/no-data/unknown and upstream coverage
remain explicit; zero is not a completeness or health assertion.

The child may use the established native count comparison presentation instead
of cloning every authored child. That is preferable to pretending an unsupported
composition supports comparison. Preserve compatible count kind/order preferences
where the current comparison renderer already admits them. Keep the parent
records presentation and authored declaration untouched for Back.

## Save and reuse

Save configuration is separate from preserving observed operands. In this bounded
cut the comparison pair remains ephemeral for both actions. The ordinary Save
stores the selected view configuration, including its original composition and
record presentation; it does not store the held parent, complementary operand or
comparison pair. Show a concise note in the comparison/preservation flow:
**Save stores the selected view; this comparison is not saved.**

Reopening stages that selected view for explicit Run. Do not imply it reopens the
comparison, silently substitute a fresh parent read, or create a hidden capture.
The existing explicit evidence workflow remains available when durable evidence
is required. A persistable question-based rest comparison can be considered later
through the existing saved form, but is not required to make the current
investigation gesture useful. No comparison-document or custody type is added.

## Acceptance journey

Open an actual saved/team count breakdown with records, select a rare group and
stage it using Bars and keyboard/Table equivalents. Compare selected against the
held parent, verify no read on opening and only the intended selected read on Run.
Return and confirm the same authored parent/fields/records remain held.

Then compare that group with the rest of its fixed parent. Verify the complement
retains environment/time/predicates and handles missing/empty groups accurately.
Run independently; show actual counts/support with a common visual scale. Save
and reopen the selected view configuration, with a visible statement that the
comparison is not saved and no automatic telemetry read. Exercise stale/hidden controls and pending record work without
expanding the UI. Update real storyboard and compact frames for both operand
meanings and parent return. Static screenshots alone do not establish acquisition
or custody behavior; source and observed lifecycle verification remain separate.

## Initial implementation review: operand orientation

The current child heading correctly distinguishes Selected vs held parent
(overlapping) from Selected vs rest of parent, and the temporary-save note is
visible. The shared components retain generic Baseline / Current held labels.
The heading alone does not specify which operand occupies which column.

Before acceptance, add compact visible application-owned operand context directly
before the shared comparisons: **Baseline · Held parent** (or **Rest of parent**)
and **Current held · Selected**, each with its actual UTC read clock. Keep the
independent-read/unknown-coverage fact concise. This gives the existing column
labels concrete meaning without adding display configuration, another control or
long repeated explanations. Read clocks must not exist only in tooltips when
their difference changes how these operands should be interpreted.

## Source follow-through

The application now adds the requested compact Baseline · Held parent/Rest of
parent and Current held · Selected context with each actual read clock, directly
before the shared comparison display. This resolves operand orientation without
new component properties or controls. The child heading retains overlap meaning,
and the visible note states that saving keeps the selected view, not the pair.
Refreshed actual native frames are accepted below.

## Recommended 308 outcome: preserve the useful comparison recipe

After cohort discovery, the investigator's learned result is often the relation
between selected and rest, not merely the selected query. 307's temporary pair
makes that investigation possible but Save necessarily drops that relation.
Closing this reuse gap is more valuable than adding another measure or CLI flag.

Prefer one bounded saved **question-based selected-versus-rest comparison** in
the existing composition grammar, ordinary Preserve flow and portable artifact.
The current model already supports fixed baseline questions in flat native views,
while a composition comparison child is captured-baseline-only. Reconcile that
split through checked operand references rather than creating another comparison
language, document type or catalog. Keep the records and authored presentation
beside the comparison under the existing query owner.

The saved recipe must explicitly mean independently run fixed selected and
complementary questions, retaining original parent predicates, typed environment,
bounds and category/missing semantics. Reopening stages configuration; Run reads
both. It preserves no historical observation or atomic-partition claim. Export
must remain ID-free and admit through the same canonical model.

Do not treat the held-parent case as the same operation: its baseline is an
actual earlier observation. It remains temporary unless the user uses an existing
explicit evidence-preservation workflow. Saving must never silently convert held
parent into a newly queried parent, strip an operand, or create a hidden capture.
First preserve the rest-query recipe; defer additional comparison families until
this concrete lifecycle works.

Acceptance should reproduce a useful rare-cohort comparison, preserve it, reopen
without a read, Run independently and see both operand meanings again. Export
and import it into a clean catalog, then verify original fixed-question opening
remains distinct from presentation reuse. Ordinary Apply of this fixed pair must
refuse until a checked destination-binding operation exists; it must not copy a
foreign baseline question into a new scope. That advances
the user's compose/capture/reuse intent with one complete outcome rather than a
new profile or an incidental guard milestone.

## Actual 307 frame review

The actual parent frame shows the staged Missing key OR Present empty string
selection and all four groups of the 107-record population. The selected-versus-
held-parent frame correctly shows four selected records against the earlier held
107, explicitly overlapping. The selected-versus-rest frame correctly maps its
three-record rest to Baseline and 104 selected records to Current held; it retains
missing and empty groups separately. Both comparison tables share a 0–100 count
scale, keeping small-group counts visible rather than hiding them.

The final refreshed comparison frames resolve the initial inline-context defect.
The save-only-selected note occupies its own compact line; Baseline · Held parent
or Rest of parent and Current held · Selected appear on two separate lines with
UTC millisecond read clocks. Operand orientation is clear before the generic
shared table headings. Exact clock precision remains available in titles. The
107-versus-four overlap and three-rest-versus-104-selected comparison counts and
common scale remain unchanged. All three actual frames are accepted; no visual
blocker remains. No additional controls or API were needed.

The verification owner reports main native workflow green in 1.6 minutes,
including both Back/focus paths and retained parent query/fields/snapshot. Those
observed assertions establish parent retention separately from the cropped
comparison frames. The approved read-failure/retry secondary remains pending in
this review until its verification report arrives.
