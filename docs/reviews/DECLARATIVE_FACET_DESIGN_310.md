# Declarative categorical facet — design review 310

Status: implemented source and all three actual native browser frames accepted.
Read with [309 continuation](PAIRED_WINDOW_DESIGN_309.md) and
[existing staged facet semantics](../FACET_SELECTION.md).

## User outcome

A service/team contributes not only useful charts and records, but the field
people naturally use to narrow them. A saved view can place one declared
categorical facet beside its existing data displays. This advances the whole-UI
grammar through one useful control; it does not add a separate search builder,
facet registry or private component query.

The first cut targets one ordinary single-query composition, excluding fixed
comparisons and specialized profiles. The declaration supplies a checked qualified
field and bounded presentation. The application's existing checked question
remains the sole authority for applied predicates, environment and time.

## Minimal authoring and placement

Use the existing View settings composition authoring to add/configure the facet,
with the existing qualified-field vocabulary. Give the rendered child the field's
human label plus namespace where necessary. Do not put its editor before Activity
or introduce another permanent toolbar. The control is useful beside the view's
breakdown and records; source/schema editing remains an advanced escape hatch.

Reuse existing passive facet/count-value display behavior, selection tray and
application actions. A child mount, local Find, visibility toggle, imported view
or restored parent must never acquire values merely because it rendered.

## Explicit values, honest support

The child offers **Read values** through the existing counted-values owner when
no admissible current values are held. Reading is explicit source IO for the
current applied fixed query and declared field. It retains all existing filters;
these are not self-excluded values or a global inventory. A short basis line can
say **Values in current fixed scope**, with count/read time and existing
availability facts. Values beyond the supplied record page require the actual
counted-values result, not an aggregate invented from visible rows.

Pending, failed, unavailable or changed-between-reads input must not become zero
or healthy. Preserve the current counted-values admission rules and finite
budgets. Missing key and Present empty string are separate visible choices.
A local Find narrows the displayed supplied values only; it does not change the
telemetry predicate or claim exhaustive discovery.

## Existing selection and execution boundaries

Include/Exclude gestures stage the existing typed categorical intents in the
application selection tray, without telemetry IO. The tray retains its existing
Apply facet selection (explicit acquisition) and Open selected facets (staged
child, then explicit Run) meanings. Do not introduce a second component-local
Apply, silently auto-run after a click, or rewrite the applied question in private
component state. Pointer and keyboard interactions use the same checked pathway.

Use **Clear staged facet** to remove only this facet's pending intents. An
application-owned origin marker may distinguish them from other pending choices;
it is not another persisted predicate language. Clear does not remove applied
filters, the same-field conditions from another source, the declaration or the
whole selection tray. Existing global Undo/Clear retain their own meanings.

After applying a narrowed scope, the child still declares the same field. Its
values are unavailable or held for their original query until explicitly read
for the new current scope. Do not quietly refill them from a different scope or
reuse a retired source control after return. The application owns request leases,
read cancellation, selection identity and retained-parent focus.

## Save and contribution

Save stores the facet declaration and the checked applied/staged query through
the ordinary composition. Unfinished tray choices remain pending and block Save
under existing rules; they are not serialized as another selected-values property
beside the predicate. Export/import/CLI validation use the same declaration and
admission. Reopening stages configuration; Run obtains the ordinary query, and
Read values remains its own explicit action.

Current-scope composition reuse copies the supported field/control configuration,
not the source's predicates, counts, value selections, source tokens or clocks.
If the field/signal/profile is incompatible, show the existing refusal rather
than silently dropping the facet. No fixed-comparison baseline is rebound by
this first ordinary-view feature.

## Acceptance as a reusable team workflow

Author a real service view with a meaningful facet such as route/cohort, chart
and List/Table. Read values explicitly and show that its full scoped count is
separate from bounded records. Include one value and exclude another; verify
staging reads nothing and the existing tray clearly names the selected clauses.
Use Clear staged facet without erasing unrelated pending or applied filters.
Exercise missing and empty choices and keyboard equivalents.

Apply or open/run the chosen population through existing controls, then save and
export it. Import through a clean catalog/CLI and apply the composition to a
different compatible destination, preserving that scope. Values must not load on
mount or transfer from the source. Read there explicitly and verify the new
counts. Update actual compact frames and executable storyboard; source/lifecycle
checks separately establish read/write effects and stale-control refusal.

## Initial source usability and accessibility review

The source uses the existing compact counted-values styling, bounded 18rem value
list, native buttons/search, qualified field labels, status text and shared visible
focus treatment. Missing and present-empty values have distinct Include/Exclude
accessible names. Its source/pending states and explicit Read values match the
contract; no data request is owned by custom-element lifecycle.

Two concrete corrections were requested before final review. A newly added facet
must default before Records so its useful control is discoverable above a
100-row List; updating an existing facet preserves its authored position, and
serialized YAML order stays authoritative. No new reorder toolbar is needed.
Also Read values / Clear staged facet need logical keyboard-focus continuity
through rerenders: the initial component restores focus only by aria-label while
those buttons lack one. Retain the field-qualified action focus through loading /
completion and staged-clear updates instead of dropping the user to the page.
Include/Exclude already have qualified accessible names.

Both corrections are implemented. New facets are inserted before existing
children; updates preserve their authored position. Read/Clear now have stable
accessible action names and pending logical-focus handling through loading/status
and completion, with fallback to the enabled Read action. These are bounded
keyboard/discoverability fixes; no extra toolbar or editor was added.

## Actual 310 visual acceptance

All three actual frames are accepted. The
[current scope frame](../journeys/screenshots/310-current-scope-values.png) puts
Service cohorts before the retained List, showing 107 matching records and four
supplied categorical values separately from the 100 supplied inspection records.
Read values and Clear staged facet are concise explicit actions; qualified
record/resource zone fields remain visible in the List.

The [staged child](../journeys/screenshots/310-selected-facet-staged.png) shows the
original service/tenant predicates plus Missing OR present-empty selection,
explicit Run selected population and an unloaded/disabled facet until Run. It
does not relabel parent values as newly acquired child values. The
[selected scope frame](../journeys/screenshots/310-current-selected-values.png)
then shows four matching records and two supplied values, missing and empty at
two each, beside the retained List and qualified fields. Counts and upstream
coverage remain clear. The bounded scroll list avoids an unbounded value wall.
No visual blocker remains.

The verification owner reports the main native workflow green in 1.9 minutes,
including origin-safe Clear, zero-IO staging, child/field retention, stale controls
and explicit values for the new scope. Source/display checks separately cover
keyboard focus and invalid/unrepresentable values; static screenshots alone do
not establish their execution. The final representability guard disables values
that cannot become valid predicates rather than implying every displayed string
is actionable.

## Next substantive outcome — useful service default adoption

Prefer extending the existing OTel HTTP service template/Customize path with this
facet, chart and records declaration, rather than adding a second scope-control
language or another profile. A discovered HTTP service can contribute a useful
route facet over its existing exact server-span population and carry that view
through normal Customize, Save/export and catalog reuse. Read values stays explicit
and the child must not claim an exhaustive route inventory before acquisition.

Counted values are a distinct typed role with independent receipts; they should
not be forged into an unrelated ranked snapshot merely to satisfy a universal
model aesthetic. Reuse or simplify their app binding when the actual adoption
needs it. The product outcome is a meaningful built-in/team view using the new
control in an existing service journey, not an adapter rewrite as a milestone.
