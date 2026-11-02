# Whole-product architecture checkpoint — 312

Status: source review with one executed pure reproduction; browser usability and
full task observations are owned by the design/verification reviews. No production
source was changed. Scope: Activity → service → contributed view → facet/cohort →
trace → Preserve/reopen, incorporating milestones 304–311.

## Overall assessment

The architecture has useful boundaries worth retaining: one Go server, strict
portable artifacts and explicit catalog effects; shared record/aggregate snapshots;
passive display elements; source reads outside components; typed scope policy;
independent read clocks; and explicit distinction between held observations and
saved questions. Real HTTP customization now exercises a contributed interactive
view rather than a separate demo widget.

The principal weakness is composition of those individually validated capabilities.
Several prohibited combinations are surfaced only after a user follows an offered
action. The UI has accumulated lifecycle restoration and readiness logic in one
large workspace adapter. Prioritize complete task transitions and consolidating
that adapter around proven seams, not replacing the grammar/query engine or adding
more isolated controls.

## Ranked findings

### 1. Confirmed task failure: contributed facet view cannot include its cohort

`ui/selected-count-comparison.mjs:8` permits ordinary authored count compositions
with a facet. `ui/population-workspace.mjs:183` offers **Include comparison in view**
for every remainder continuation. That handler adds the comparison child, while
`ui/view-composition.mjs:27` and Go composition validation prohibit a facet beside
either comparison kind. An ordinary count view with a facet therefore allows a useful temporary
contrast but offers an impossible persistence action. The shipped HTTP widget may
carry `spanStatus: true`, which 307 excludes earlier; this finding does not claim
every untouched HTTP default reaches the failing action.

Executed pure reproduction constructed an ordinary route-count composition with
method facet and called the same `setComparison(createQuery(...))` transition.
It refused with `Facets do not support comparison compositions yet.` This is a
confirmed transition defect, not a claim that the prohibition itself is wrong.
Resolve the user task with an explicit **Include comparison without facet**
transition explaining that it removes the control, not applied filters, and drops
any incompatible template binding. A pure atomic conversion must preserve selected
and baseline predicates, record presentation and aggregate. Prefer this bounded
conversion over expanding coexistence semantics; do not silently drop the facet or
merely suppress the error.

### 2. Confirmed retained-resource growth on repeated facet child returns

`ui/population-workspace.mjs:465` adds each facet lookup to `valueLookups`.
`ui/population-workspace.mjs:253` remounts the facet on a facet continuation's Back,
without disposing/removing the previous lookup. `pauseValueLookups` invalidates
requests but does not remove entries. Full render (`:282`) and invalidate (`:661`)
eventually clear the set, so this is retained growth during repeated child/Back
cycles, not an installation-wide permanent leak. Detached lookup closures remain
reachable until that reset. Give the facet adapter an explicit disposal/replacement
slot and test repeated Back stability; do not introduce a global resource manager.

### 3. Opportunity: source demand exceeds visible comparison needs

`exploration/source.go:176`, `:216`, `:255`, `:274` and `:318` perform sequential
count, rows, groups, field discovery and optional timeline reads. A full grouped
comparison reads selected and baseline through this full profile even though
comparison admission consumes count/groups (plus currently supported measures).
Raw pages correctly use two reads; explicit facet values add exactly two existing
count/group reads (`exploration/counted_values.go:95`, `:113`). A route view with
both timeline and fixed baseline can therefore require up to ten source statements
before explicit facet demand. These are source-path counts, not measured timings.

Keep read clocks independent and avoid unsafe matching-query deduplication across
held/current/baseline policies. First record request/SQL budgets for the actual
incident flow, then use an existing aggregate-only profile where it can return the
same admitted comparison operands. Do not parallelize every read or add caching
before identifying which result roles are actually required.

### 4. Opportunity: workspace lifecycle is the next useful refactoring seam

`ui/population-workspace.mjs` is about 177 KiB in 665 physical lines at review time.
Its `markDraft` (`:134`), render completion (`:306`), control sync (`:427`), Back
restoration (`:253`) and execute (`:532`) manually coordinate many refresh callbacks,
source/filter errors, record workflows and comparison controllers. Recent native
findings repeatedly involved disconnected readiness, stale controls and focus after
adapter replacement. This is evidence for extracting existing app-owned adapters,
not a general state framework rewrite.

First isolate facet demand/lifecycle and comparison continuation/persistence into
small owners with explicit update/suspend/dispose and immutable input DTOs. Keep
one parent scope generation and existing query executor. Preserve current browser
and observed-model contracts while removing repeated call ordering. Do not move
fetch into custom elements or duplicate scope sessions in each extracted owner.

### 5. Opportunity: validation/schema duplication needs one parity inventory

Query/composition checks exist in Go, browser and OpenAPI; e.g.
`workbench/view_composition.go:186`, `workbench/view_facet.go:55`,
`workbench/view_templates.go:198` compile the query in nested paths, while browser
facet eligibility repeats limits/profile restrictions and the OpenAPI builder
manually spells role combinations. Recent cross-language fixtures caught real
null/default/name drift. Retain strict boundaries; inventory semantic rules and
reuse small Go/browser helpers where their meanings agree. Continue emitting real
server responses into browser/schema checks. Avoid introducing a second validator
language or interpreting successful JSON Schema validation as source authority.

## Typed control model and source reliability

310's counted-values input is consistent with proposal 260's several controlled
model roles. It has independent count/group receipts and app-owned scope binding;
it must not be converted to an invented ranked-snapshot guarantee just to reuse H1.
Standardize this binding when extracting the facet adapter, and require future
controls to use the same scope/session intent envelope. No new control-specific
query executor is justified.

Source transport already bounds result rows, bytes, memory/read budgets and
execution time (`exploration/source.go:69` onward), uses request context and asks
ClickHouse to cancel on disconnect. Portable codecs and CLI enforce bounded strict
admission and no retries after ambiguous publication. Those are strengths. This
review did not benchmark concurrency, heap use or production-scale datasets, and
does not claim those systems are capacity-tested.

## Finite implementation proposal

1. Complete the contributed incident task: explicitly support the facet → remainder
   → durable view transition and make disallowed combinations discoverable before
   an offered action fails. Include Save/reopen/portable semantics and unchanged
   parent/held behavior.
2. Extract the facet application adapter while fixing repeated-Back retention;
   use its explicit lifecycle as the common typed control-binding seam. Preserve
   the passive component and existing counted-values endpoint.
3. Establish and improve incident-flow read budgets: actual HTTP/SQL counts for
   service Customize, cohort Run, trace expansion and facet demand; then remove
   unnecessary baseline row/field acquisition with existing profiles if measured
   checks establish equivalent admitted semantics.

Use those three finite outcomes before another whole-product review or widget
expansion. Validation parity can be strengthened within each changed boundary;
it need not become an open-ended standalone migration. Product design and planner
should reconcile rankings with their observed task failures before final backlog.

## Bounded 312 source resolutions

The correctness bundle addresses the confirmed transition/resource findings without
expanding facet/comparison coexistence. A pure validated conversion removes only
the facet control and adds the existing fixed-query comparison; the UI explicitly
says **Include comparison without facet** and explains that applied filters, chart
and records remain. The category marker is retired visibly. Invalid conversion
leaves the source AST unchanged and staging performs no source read.

All tray actions now share one per-update command token and generation/draft/
selection/question snapshot. Runtime handlers additionally require current tray
membership/visibility and common pending/draft readiness. ABA or reinserted old
controls cannot inherit a new selection. Invalid query derivation disables source
actions while permitting safe Clear/Undo recovery. Facet replacement nulls its
owner, deletes its lookup from the set and disposes it before creating another,
resolving the repeated-Back retention finding.

The backend also freezes complete affected query inputs before compilation/read
and detaches inventory/discovery namespace selectors. The terminal-page consistency
check now marks count greater than terminal observed offset+rows as changed between
reads, while preserving valid empty delivery beyond the counted end. This avoids
source certainty inferred from a short page without matching count. Source review
accepts these bounded corrections; actual SQL/native/model/property results are
tracked in validation rather than inferred here.

The revised next steps are active-workspace consolidation (including dormant root
editors and duplicate readiness/render adapters), followed by measured incident-flow
request/SQL budgets. Keep the earlier findings as historical evidence, and do not
turn the next step into another widget/profile or a generic state rewrite.
