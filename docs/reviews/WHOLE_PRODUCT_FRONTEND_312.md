# Whole-product frontend review — 312

Reviewed 2026-11-02 against the delivered 311 baseline (`6ff16ae`). This is a source review, not a new browser or performance certification. Findings and line citations below describe that pre-implementation baseline; the existing 302–311 journeys remain the behavioral evidence for their delivered slices.

The subsequently selected 312 implementation addresses finding 1, explicit facet-control removal when promoting a temporary rest comparison, and bounded replacement/disposal of facet lookup owners. It does not claim to deliver the remaining title, payload-budget, trace-cancellation or local-search recommendations. Native verification and final delivery status belong in the implementation/validation records.

## Product through-line

Goo's useful core is: **see activity, narrow a population, compare it, inspect an example, preserve the question or evidence**. The UI should keep that question visible while the source read, display, selection draft and return destination change independently.

| Incident-flow step | Existing owner worth retaining | Important boundary |
| --- | --- | --- |
| Activity | [stream-entry.mjs](../../ui/stream-entry.mjs#L49), especially `rememberParent`/`returnToParent` at lines 84–85 | Returning restores paused activity rather than starting a new stream. |
| Service → contributed view | [view-template-handoff.mjs](../../ui/view-template-handoff.mjs#L40) | Customize stages the actual checked widget question. Template category binding must not rebuild scope from labels. |
| View / facets / records | [population-workspace.mjs](../../ui/population-workspace.mjs#L123), [filter-values.mjs](../../ui/filter-values.mjs#L22) | Explicit counted reads are independent of inspection pages; selection drafts do not silently alter the applied population. |
| Cohort comparison | [selected-count-comparison.mjs](../../ui/selected-count-comparison.mjs#L1), existing population Run and comparison preparation | Held parent, independently read rest, and saved fixed baseline are different operands. None establishes a partition or transactional snapshot merely by being displayed. |
| Trace | [record-results.mjs](../../ui/record-results.mjs#L53) | Inspector expansion is independent context; Open trace as view transfers question intent, not an admitted result. |
| Preserve | [population-workspace.mjs](../../ui/population-workspace.mjs#L359), [view-artifact.mjs](../../ui/view-artifact.mjs#L1) | A saved/exported definition is configuration. Capture is a source read. Investigation attachment is an explicit separate action. |

The common display graph is genuinely useful: passive elements and pure declaration/model validators remain separate from application acquisition. Keep that separation. The next work should close connected workflow gaps rather than add another chart family or parallel query owner.

## Ranked findings

### 1. High: selection-tray commands do not share the same currentness and readiness contract

**Source-confirmed inconsistency:** [updateSelection](../../ui/population-workspace.mjs#L246) captures `generation`, `draftRevision`, selection identity, connectedness, pending records and other guards for authored Compare. Sibling Apply, Open, Undo and Clear have weaker handlers. In particular, Apply calls `execute(next)` unconditionally with a question captured when the tray was rendered. Its disabled expression does not include query/display drafts, pending page reads or count preparation.

An ordinary counterexample needs no unusual input: stage a facet value, edit the query or display without applying that draft, then use Apply facet selection. The tray can still offer Apply, reading the old configured selector plus the old selection while another draft is unresolved. A retained old Apply callback also lacks the epoch protection already present on Compare. This is an application action-lifecycle defect, not a passive-component defect.

**Recommendation:** one small, app-local immutable selection-command context and readiness function shared by all tray actions. Both the disabled state and the handler must consult it. Reissue controls after the same existing state changes that already refresh the tray. Preserve the current distinction between Apply's explicit read and Open's staged child; do not introduce a command bus or global store.

**Acceptance:** query/source draft, pending row selection, page read, baseline preparation, inactive visit and child ownership make every conflicting action unavailable; old callbacks cannot read or overwrite a newer draft after scope change/Back/ABA. Current Apply performs its intended reads once; current Open performs zero reads and focuses the staged continuation. Undo/Clear affect only the current pending selection.

### 2. High product clarity: category rebinding can leave a semantically wrong chart title

311 correctly preserves literal facet fields. However, existing category binding can change the aggregate's category while retaining a title that describes the original category. The HTTP preset supplies **HTTP server routes** as both view and aggregate title ([descriptor](../../workbench/templates/http-server-routes.json#L77)). [authoredRequest](../../ui/view-template-handoff.mjs#L40) accepts another eligible destination grouping; [authoredResponse](../../ui/view-template-handoff.mjs#L53) requires the original presentation unchanged. [ScopedBreakdown](../../ui/scoped-breakdown.mjs#L39) then uses that authored title.

Consequently an HTTP template applied to `resource:service.namespace` can still present a chart titled HTTP server routes. The data and facet binding may be correct, but the human's interpretation is wrong. This is specifically about category-dependent presentation text, not a request to rewrite user titles or infer semantic aliases.

**Recommendation:** retain literal authored titles and show an always-visible effective category/measure beside them. Distinguish the transport/view name from the chart's bound category. Consider an explicit category-neutral default for shipped reusable templates. Do not silently rewrite a user title or make the facet follow the category.

**Acceptance:** apply the same HTTP view to route and namespace groupings; each chart visibly identifies its actual qualified category and requested measure, while the method facet, exact destination scope and literal title remain unchanged. Save/export/reopen preserve that interpretation.

### 3. Medium: bounded row counts do not yet provide bounded facet rendering cost

[FilterValues.result](../../ui/filter-values.mjs#L12) bounds group count, but accepts arbitrary string lengths. The passive model bounds 1,000 items and validates partition accounting ([categorical-facet.mjs](../../ui/categorical-facet.mjs#L8)); its **property** input has no overall value-byte budget. The HTML attribute path has an 8 MiB limit ([line 14](../../ui/categorical-facet.mjs#L14)), creating a different bound for the same public model.

Rendering clones the model and rebuilds every row and both action buttons ([lines 19–21](../../ui/categorical-facet.mjs#L19)). Visible labels are shortened, but exact text is repeated in titles, accessible names and closures. A body facet with long legitimate values can therefore be expensive despite a scroll viewport and disabled unrepresentable predicates. Repeated readiness-only updates rebuild this same structure. This is a source-level resource risk; no latency or memory measurement is claimed here.

**Recommendation:** give the supplied facet model a documented total byte budget shared by property and attribute input, and avoid rebuilding unchanged rows for readiness/hasStaged-only updates. Preserve exact values for actionable membership. If the budget is exceeded, show an explicit unavailable/budget state; do not truncate the domain while retaining an exhaustive-total claim. Measure the 1,000-value case before choosing virtualization.

### 4. Medium: trace invalidation fences stale publication but does not cancel its acquisition

[traceRows/logs/surrounding](../../ui/record-results.mjs#L43) call `fetchQuery(q)` without a signal. [invalidate](../../ui/record-results.mjs#L50) advances generations, so late responses are ignored, but the requests continue. The existing [exploration fetch adapter](../../ui/exploration.mjs#L97) already accepts an AbortSignal. Rapid inspection changes or leaving a large trace can therefore leave avoidable work running.

**Recommendation:** add cancellation to the existing trace owner, separately for trace/log/surrounding reads as their current epochs require. Do not add another fetch layer. Closing, switching traces and disposal should abort the relevant requests while preserving the existing modal/return leases.

**Acceptance:** delayed reads are aborted on withdrawal; an old reply cannot repaint, move focus, invoke Activity's old close callback or replace a newer trace. Current independent trace/log paging still works.

### 5. Medium accessibility: local facet Find has no matching-result announcement

The facet keeps search text and button focus across same-read refreshes, and makes unsupported exact predicates unavailable—good improvements. However [its Find callback](../../ui/categorical-facet.mjs#L21) only changes the passive list's query. [GooFacetList.refresh](../../ui/components.mjs#L29) hides rows and returns counts, but no caller announces them. A zero-match Find leaves a blank viewport beneath the unchanged whole-scope total. Keyboard/screen-reader users receive no explicit distinction between “no local match” and “no observed values.”

**Recommendation:** use the existing list projection counts for a concise live status such as “0 of 37 loaded values match.” Keep whole-scope count and local-search count distinct. Add no lookup or source read and do not remove focus from the search input.

## Maintainability and reuse

The largest obstacle is concentrated application orchestration, not a shortage of components. `population-workspace.mjs` currently has 665 physical lines and a longest line of 5,542 characters. It owns acquisition, comparison, draft synchronization, field authoring, catalog import/export, composition rendering and nested return restoration. Several “refresh” functions are called from both draft and record-selection paths. Recent bugs fixed one missing refresh or readiness branch at a time; finding 1 is the same class still present in adjacent commands.

Make structural changes follow a concrete behavior fix:

- Extract a small **selection state / command policy** module as part of finding 1. Pure scope/predicate construction stays separate from DOM and reads. Keep the existing population owner responsible for effects.
- Separate catalog/artifact UI mounting from population rendering after its lifecycle tests are in place. Existing generation/input-revision and download retirement rules must remain attached to the visit.
- Do not consolidate snapshot admission with passive display validation. [SourceSnapshots](../../ui/source-snapshots.mjs#L54) deliberately distinguishes full exploration and record-page authority. Do consolidate repeated *application readiness* facts; those are currently recomputed unevenly.
- Keep the Go compiler/server admission authoritative. JS declaration checks, template eligibility, count-cohort context and artifact omission normalization have legitimate different jobs. Prefer a finite cross-language fixture matrix over another browser query compiler. Improve refusal reasons where a generic “ordinary query” message actually means a page/bucket/profile limit.

The trace handoff's [per-read node/epoch lease](../../ui/record-results.mjs#L53), facet pending-origin identity, and typed environment preservation are good patterns to reuse. Reuse the invariant, not a new generalized capability framework.

## Recommended finite sequence

**First fix — selection command consistency.** Deliver finding 1 with one real Activity/service-contributed selection journey, draft/pending/retired-action negatives, and keyboard Back/fresh-control checks. This is substantive reliability work across the core incident flow and creates a useful extraction seam.

Then choose these three bounded milestones:

1. **Contributed cohort continuity.** Resolve the known facet → cohort/rest → further exploration discontinuity using the existing question/composition owners. Preserve literal facet and record presentation where supported; give explicit, actionable refusal where a saved pair cannot carry a role. Stop when one service → facet → selected/rest → trace → Preserve journey completes without an accidental flat-view downgrade or lost return context.
2. **Bound category presentation.** Deliver finding 2 plus concise category/measure/scope context across chart, facet and comparison surfaces. Validate one reused HTTP view over two destination groupings and one independent baseline pair. Keep data semantics and literal authored text distinct.
3. **Bounded, accessible investigation refresh.** Deliver findings 3–5: shared facet byte budget, unchanged-row readiness refresh, local-match status, and cancellation in the existing trace owner. Use focused stress/keyboard/withdrawal tests, not a new general UI engine.

Review the whole incident journey again after these outcomes. Do not expand the backlog with more widget kinds, template parameter families or aesthetic-only cleanup before these workflow and interpretation gaps are closed.
