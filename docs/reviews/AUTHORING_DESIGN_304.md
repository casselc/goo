# Whole-flow authoring and usability review — 304

Status: implemented bounded transition accepted in source and actual browser review. This review follows the current
[implementation plan](../IMPLEMENTATION_PLAN.md) and
[grammar proposal](../VIEW_GRAMMAR_PROPOSAL_260.md). It examines current application
source, executable native journey contracts and actual captured application
frames. It does not claim a newly executed browser session or a live competitor
comparison.

## Product essence and whole-flow assessment

Goo helps a responder start with telemetry, isolate something meaningful, choose
how to see it, compare it and turn the useful result into reusable configuration.
The interface should allow the investigation to grow from what is already on
screen. Users should not need to know the composition grammar before choosing
the next useful representation.

The current journey is substantially connected:

| User intent | Current path | Assessment |
| --- | --- | --- |
| What is happening? | Activity logs/spans, scope and pause | Keeps the data-first entry; do not insert an editor ahead of it. |
| Isolate the relevant observations | Fixed observed window, facets/filters and linked population | Existing explicit query owner preserves scope; several transitions remain visible, but another entry point is not needed. |
| Follow an event across services | Expand trace → Open trace as view | 303 now provides a direct staged records Waterfall with explicit Run and parent return. |
| Choose presentation/fields | Customize and View as Table/List/Waterfall, qualified fields | 302 removes invented grouping and makes raw records useful in their own right. |
| Add a visual explanation | Existing reusable aggregate/display authoring | **Broken from the records-only destination:** its editor suppresses aggregate authoring. |
| Compare | Existing grouped displays, fixed comparisons and explicit captured-count reference | Available once the relevant composition/profile exists; comparison is not a reason to invent aggregates from a page. |
| Save and reuse | Preserve → Save exploration view; catalog Inspect display → Apply composition | Existing configuration reuse keeps destination scope. Category templates are a separate supported route, not a universal service/team authoring mechanism. |

The native 302/303 storyboards and actual frames demonstrate raw customization,
qualified fields, paging, Keep, save/reopen and reuse. The 286 actual composition
frame demonstrates useful scoped Bars beside records; the 300 actual authoring
frame demonstrates separate question/evidence/view definition preservation and a
bounded reusable template. These are useful building blocks. Their existence
does not mean a responder starting from the new raw view can reach all of them.

## Primary gap: the useful starting point is an authoring dead end

A responder opens an exact cross-service trace as a Waterfall or customizes an
ungrouped log population. They then ask, for example, “How are these observations
distributed across services?” Changing the record kind only rearranges supplied
records. The current visual controls cannot add the full-population breakdown.

This is confirmed by `mountDisplayEditors` in `ui/population-workspace.mjs`, which
returns immediately for a records-only composition. `mountCompositionEditor`
offers Compose aggregate and records only when there is no composition and a
current authored display already exists. The record-only branch exposes record
kind/source but no transition to an aggregate composition. Source editing or
applying an already-authored catalog composition is possible, but neither is a
natural first-use authoring path.

The gap cuts across discovery, presentation, comparison and service/team reuse.
It matters more than adding another chart type, moving a disclosure or renaming
Preserve. Another navigation destination would add fragmentation without fixing
it. The bounded next outcome should connect these existing capabilities.

## Recommended implementation: add a breakdown to this view

Use one **Add breakdown** action inside the existing **View settings** for a
supported records-only view. It reveals a compact qualified grouping-field choice
and an explicit staging action. Reuse the existing field/grouping semantics and
ordinary count declaration, initially Bars with the existing compatible Table
switch. Do not introduce a new wizard, chart gallery, global builder or duplicate
query editor.

The operation deliberately changes the view from records-only to the existing
aggregate-and-records composition. It preserves Table/List/Waterfall, qualified
record columns, checked environment, fixed bounds and predicates. It explicitly
chooses/replaces the grouping required by that breakdown and resets delivery to
page zero. It must not silently preserve an incompatible multi-field grouping or
rewrite the original filters. A staged preview/status should identify the chosen
field and make Run the next action.

Adding the declaration acquires no data. The existing explicit Run obtains the
supported full-population aggregate plus bounded records through the existing
profile. Until then the old page is not an aggregate preview or comparison
witness. Prefer the existing coherent staged-state behavior over presenting a
mixture that looks freshly acquired. Do not derive full counts from the supplied
Waterfall/List page or add another data owner.

After Run, use the existing aggregate kind/scale/refinement controls and existing
comparison authoring where the query is compatible. Save and current-scope Apply
remain the existing flows. This cut does not promise general numeric/native
comparisons, arbitrary template parameters, organization permissions or a package
marketplace. A service/team can contribute a useful saved definition using the
current catalog and bounded templates; expanding contribution/distribution should
follow a demonstrated gap in that complete task.

## Acceptance as one meaningful journey

Start with a real cross-service incident in Activity. Isolate its fixed scope,
expand a trace and open the staged records view. Run and choose useful record
fields/Waterfall, then use Add breakdown to select service (or another qualified
field that genuinely separates the incident population). Confirm staging reads
nothing; Run acquires the full aggregate and bounded page. The visual must make
comparison between distinct groups possible without reducing the original scope.

Save the result as a named service/team view, reopen its configuration with no
telemetry read, and apply its composition to another compatible destination scope.
Run there explicitly and verify that the destination scope remains its own.
Exercise the existing compatible comparison path in the same responder task,
using an actual reference rather than a planted source-JSON definition. If the
profile cannot express that comparison, report the boundary instead of silently
changing the query to manufacture compatibility.

Check keyboard progression through Add breakdown, field, stage and Run; focus
must land on the next visible action. Narrow layouts should wrap these few
controls without bringing the editor above the initial stream. Verify Back returns
the retained parent, and selected/pending Keep or unsaved invalid source does not
get erased by the transition. Update the executable storyboard and actual compact
frames for this end-to-end task rather than adding a collection of isolated demos.

## Follow-on decisions

Keep Save configuration, Keep selected records, Capture and Record a finding
semantically distinct. Their current grouping under Preserve is understandable
when opened, although the discovery vocabulary can be revisited after the new
journey. Likewise catalog Inspect display/Apply has more steps than ideal, but
reuse already works; prioritizing its labels before first-use composition would
optimize a path that new authors cannot yet complete.

After this bounded transition, repeat the task-level assessment. The next choice
should follow a concrete failure to compare, contribute or reuse a meaningful
view. Do not automatically continue with another specialized profile or an
unbounded WYSIWYG/editor rewrite. Maintain one data model, one query owner and one
visible authoring path as the grammar broadens.

## 304 implementation source review

The implemented bounded transition matches this recommendation. Existing View
settings gains one Add breakdown disclosure with Field, View as Bars/Table and a
conditional Replace existing grouping checkbox. Current grouping and the
record/attribute/resource namespaces are stated concisely. Suggestions use service
and current grouping, without claiming exhaustive inventory or adding a lookup.

Editing becomes a tracked draft: the user must Add it or Discard breakdown draft
before saving. Discard retains the current query and held records. Replacing a
nonempty different grouping requires explicit confirmation; keeping the same
single grouping does not. Validation failures stay in the form. Successful Add
retains the records declaration, stages the checked new grouping/page zero and
focuses the existing Run. No aggregate is synthesized from the old supplied page.
No additional editor, acquisition owner or preservation type appears.

Source review has no design blocker. The final native whole-flow journey passes
(reported by the verification owner, 1.9 minutes), including comparison, Save/reuse,
span presentation retention and effective named-environment entry. The obsolete
primary fixed-exploration typed-scope gate was removed; the existing checked
planner remains the owner. No additional control or source profile was required.

## Actual 304 visual review

The actual cohort breakdown frame shows 100 fast and three slow records with a
full-population count of 103, preserving the existing count/share vocabulary and
compact Bars control. The actual comparison frame makes the question legible:
slow grows from three to four while fast remains 100. Baseline/current columns,
change values and shared scale support direct comparison; its independently held
basis remains explicit. The destination reuse frame shows its own three slow
records in the saved Table presentation, rather than borrowing source counts.
No visual blocker was found in these frames.

The refreshed [combined frame](../journeys/screenshots/304-counts-and-records.png)
now visibly includes the 103-record full breakdown and the retained List with
100 supplied records, distinct record/resource zone fields, UTC and native
Inspect/context actions. The refreshed
[reuse frame](../journeys/screenshots/304-reused-team-view.png) likewise includes
its destination's three-record Table breakdown and retained List/qualified worker
values. The [comparison frame](../journeys/screenshots/304-compared-counts.png)
shows the exact baseline/current group change independently. All three actual
frames are accepted; no visual blocker or outstanding screenshot correction
remains. The native journey separately checks Waterfall retention and explicit
group replacement. See the [304 storyboard](../journeys/304-add-breakdown.md) and
actual validation record for execution details.

## Recommended next outcome after the connected authoring cut

Native 303 exposed an older fixed-exploration entry gate that excludes typed
`environmentSelector` questions from Explore linked population, even though the
recent-population planner and trace-to-view path support checked policies. This
was a main-path coherence defect: users should not have to expand a trace to reach
the same supported authoring flow. The obsolete gate is now corrected and native
effective-scope entry passes as part of 304, rather than promoting removal of one
conditional into another large milestone. The existing planner preserves
companion/predicate policy and refusal boundaries.

After that correction, prefer **portable team-owned view declarations** as the
next substantive outcome. A saved catalog entry can already be applied to a new
scope, so do not build another catalog. The missing part of the original grammar
purpose is taking the useful definition out of one development store, checking it
into a service/team repository and loading it into another Goo installation.
The 260 proposal explicitly calls for external YAML over one checked AST; current
saved-display authoring has no corresponding end-to-end portable contribution
workflow.

Keep the first cut to the current supported saved definition/template model:
export a declaration from the existing authoring workflow, review/edit it as an
artifact, and import through the same checked definition admission and catalog.
Give the user a clear choice between reopening a fixed question and applying a
portable presentation/template to a destination scope. Do not carry store IDs,
held rows, source tokens or hidden evidence into another installation. An existing
capture baseline reference must be explicitly rebound or refused, never silently
substituted. A plain noncomparison view is a valid first portable specimen.

Success is one real authored service/team view exported, admitted in a clean
store and applied to a different compatible population through the same UI, with
explicit Run and unchanged destination scope. Add no new kind language, template
parameter system, marketplace, registry or execution authority to achieve that
outcome. This advances the original contribution goal while building directly on
the now-connected records/breakdown/save/reuse path.
