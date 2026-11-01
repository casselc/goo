# Whole-flow authoring architecture checkpoint — 304

Status: source review and implementation recommendation, not delivered behavior.
Reviewed repository guidance, the 260 grammar proposal, 274 finite adoption
checkpoint, milestones 300–303, and current population/editor ownership. Design
review agrees with the primary missing capability below.

## Product task and current capability

The intended flow is stream → scope/isolate → choose records/chart presentation →
compare → save/reuse a service or team view. Stream and fixed questions already
provide selection/filtering and explicit reads. Records now have Table/List/span
Waterfall over raw scopes; exact trace handoff enters that same saved-view model.
Aggregate compositions provide count/numeric display, timeline, captured-count
comparison and category-bound template reuse. Save/reopen and presentation-only
catalog Apply already have app-owned persistence and acquisition boundaries.

The missing connection is substantive: a user entering through the newly natural
raw records flow cannot visually add the first chart. `mountDisplayEditors`
returns immediately for a records-only composition. `mountCompositionEditor`
exposes aggregate construction only before a composition exists and only when an
existing authored display is available. The raw composition suppresses aggregate,
timeline and comparison controls. Editing grouping does not itself promote the
composition. JSON authoring or a preexisting catalog specimen becomes prerequisite
knowledge for an otherwise ordinary observability task.

## Recommended bounded implementation: Add breakdown

Add one contextual **Add breakdown** action in the existing view editor for an
ordinary records-only log/span view. Let the user explicitly choose one qualified
grouping field with the existing grouping control, then stage the existing count
Bar/Table declaration plus the existing records declaration. Preserve the exact
predicate, environment policy, fixed bounds and record kind/columns/qualified
fields. Reset paging to zero as an explicit question edit. Do not choose an
implicit service grouping or copy query values from a catalog specimen.

Reuse `ViewDeclaration.create`, `Composition.create`, strict composition/query
validation and the current root staging path. If the grouping selection changes,
validate the destination question and resulting composition together before
publishing either. Invalid/cancelled selection leaves the existing raw draft
intact. This is an authoring transition over existing ASTs and acquisition
profiles, not a new composition variant or source API.

Staging performs no read. The new aggregate reports that its aspect is unacquired
until explicit Run obtains the existing full exploration result. Do not reuse raw
count/page as a fabricated aggregate or comparison operand. After Run, the current
count/timeline/comparison/template controls and Save/reuse owners become available
without adding parallel implementations. Preserve ordinary local record changes
and the same source/property representation.

## Consolidation and ownership

Keep one app-owned draft containing question plus composition, with controls as
adapters over that draft. The product currently exposes Customize signal,
Customize records, Compose aggregate and records, Layout, Records kind, Composition
source and per-display source across nested disclosures. These partially duplicate
entry and presentation choices even though their ASTs overlap. The bounded action
should close the missing transition using the same editor, not introduce a third
workbench or another permanent control row above the stream.

The passive components remain renderers. The workspace retains query staging,
source acquisition, comparison preparation, catalog Apply and persistence. Do not
solve this with a generic parameter framework, new query owner, Plotly escape hatch
or fully general WYSIWYG canvas. Further consolidation can follow observed task
friction, but do not make a broad state-container rewrite a prerequisite.

## Acceptance and follow-up

Execute a real service incident journey from raw Activity logs: isolate a service
and anomalous population, Customize/List/qualified field, Add breakdown by an
explicit field, Run, choose existing bar/table, prepare an existing captured-count
baseline, then Save/reopen/Apply to another valid destination scope. Include a span
Waterfall starting point so upgrading retains record geometry/options. Prove the
new gesture stages without fetching and the added aggregate remains unavailable
until Run. Source and properties must round-trip the same supported AST.

This outcome is preferable to another isolated profile/handoff. Milestone 300
explicitly deferred general parameters and numeric/duration/native comparisons;
301 separated global acquisition from local records presentation; 302/303 made raw
entry practical. This checkpoint connects those investments to an end-to-end
usable authoring flow. After delivery, reassess discoverability and reuse with the
whole task before selecting additional widget types or specialized profiles.

## Implementation source review

The pure `Composition.addBreakdown` transition validates the current records-only
composition and explicit qualified field, allows only the existing count Bars/Table
kinds, and requires an explicit boolean replacement confirmation for a different
nonempty or multiple grouping. It clones the question, changes only the declared
grouping and root offset, and preserves the entire records child. The resulting
question and composition are validated together before the application stages them.

The contextual form uses the existing workspace draft/error registries. The first
edit marks a `breakdown` draft, blocking Save and competing Apply operations. Its
own ready check excludes only that error so the user can finish or discard instead
of locking the form on its first edit. Discard removes only that draft and rerenders
without changing query or held records. Successful Add removes the draft then uses
`stageRootQueries`, clearing held result/aspect authority and focusing explicit Run.

The form checks epoch, exact query object, connected current records-only owner,
active parent and existing pending/page/selection/source/live guards. Detached or
reinserted controls cannot mutate a replacement draft. No source request, new API,
query owner or comparison path was introduced. The existing source representation
continues to describe the same composition AST. Architecture source review has no
remaining blocker; executed native/property/model results belong in validation.

## Next outcome after the connected flow

There is one concrete entry inconsistency to close alongside this work: the root
`PopulationWorkspace.current` callback still excludes every `environmentSelector`.
That no longer matches the supported question contract: 302/303 preserve typed
selectors and `exploration/population_scope_test.go` already covers population
planning for current-only/effective named, empty and unreported scopes. Verify the
existing linked entry under that policy and retire obsolete eligibility rather
than inventing another selector-specific owner/profile. Treat this as parity work,
not automatically the next standalone product milestone.

The next substantive recommendation is a trusted code-managed view package pilot.
The 260 proposal promises namespaced/versioned YAML declarations alongside service
manifests, lowered to one checked AST. Today shipped templates are embedded JSON
specimens and authored saved views live behind local store IDs. Interactive Save
and catalog Apply are valuable reuse, but do not yet let a service/team contribute
the same reviewed declaration from a repository.

Bound the pilot to exporting/importing an already supported category-bound view
without new renderers, recipes or parameter types. Use existing typed template
validation/expansion, a namespaced/versioned wrapper and source-free preview of the
same AST. Preserve destination-scope/baseline binding semantics; do not export
source evidence, local store IDs or captured specimen authority as portable inputs.
A real journey should author in the UI, review a portable declaration in a service
repository, then import/apply it to a second service with explicit Run. This would
advance the original contribution goal while keeping the newly connected authoring
workflow and existing catalog as the sole product owners.

## Native findings and final corrections

Native execution found two adapter lifecycle defects that pure transition checks
could not establish. Initial readiness ran before the new form was connected,
leaving controls disabled. The form now refreshes through the existing post-render
completion hook and on disclosure toggle, while retaining the connected/current
owner check. The first field edit already blocked persistence through `settings`,
but its ordinary Save button stayed visibly enabled. The edit now calls existing
`updateSelection` after `markDraft`, synchronizing Save and competing Apply controls
from the same draft errors. Its own Add/Discard exception remains intact.

The obsolete root population-entry selector gate was also removed. Applied results
with valid typed policy now reach the existing planner; the backend's typed scope
validation and all other feature restrictions remain unchanged. This is entry
parity, not a new acquisition mode. Final source review accepts these corrections
with no remaining blocker. `git diff --check` passed. Native task outcomes and
observed replay remain recorded by the verification owner rather than inferred
from these source checks.
