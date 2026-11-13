# Product-flow checkpoint after 423

Status: bounded product/usability proposal for the next roadmap cycle. This
review read the current capability index, recent 417/421/422 contracts and
reviews, and application source. The independent architecture reviewer reconciled
the proposed 424 → 425 → 426 order without an objection. It inspected the actual 422 applied mobile and
421 counted-value desktop frames. It ran no browser, source or vendor campaign.
423's performance and execution acceptance belong to its own validation record.

## Product center

Goo's center remains **current telemetry → isolate a population → follow or
compare evidence → preserve or reuse the question**. Services, native
measurements, trace search and saved evidence are useful continuations of that
loop. A Goo manifest is optional; ordinary OTel fields must remain a useful entry.

421 and 422 make the core more usable without adding another destination: an
inspected field reaches full-scope counted values, several values can become one
checked OR predicate, and Apply is explicit. The actual mobile result shows two
matching observations and the retained predicate clearly. The source still
separates buffer observations, counted-query scope and new applied scope.

The next opportunity is helping an operator move outward again after narrowing.
Do not undo 422's deliberate AND-with-parent rule or turn a local selection into
an automatic query. Keep initial data ahead of configuration, and disclose exact
questions/read facts near the actions that need them.

## Existing capabilities to preserve and reuse

| Part of the operator loop | Implemented path | Current adoption gap |
| --- | --- | --- |
| Observe and isolate | Activity log/span stream, typed filters, contextual Facet by, counted values, atomic multi-value selection | Broadening one already-constrained field requires finding and editing its predicate. A facet cannot show values outside that applied constraint. |
| Broaden | Fixed population **Expand time window** and **Broaden by removing a filter**; whole top-level conditions stay atomic | Activity's compact base chip removes the whole saved selector, while its editor can remove individual conditions. The contextual field-to-condition connection is weak. |
| Compare | Fixed population **Compare selection to current**, **Compare with rest**, previous-window and saved-baseline comparisons | The new Activity counted selection does not directly continue into the existing selected/rest comparison; users must first navigate to a fixed population and select again. |
| Reuse and preserve | Activity views, Recent / Pinned configuration, held visits, fixed questions/views, captures and kept observations | These are different promises. A complete branching journey should make the promise and expected reopen behavior obvious without adding another library. |

Evidence: [Activity owner](../../ui/stream-entry.mjs),
[filter editor](../../ui/filter-editor.mjs),
[population owner and transforms](../../ui/population-workspace.mjs),
[preservation meanings](../../ui/preservation.mjs),
[scope recall contract](../RECENT_PINNED_SCOPES.md), and
[current capability index](../CAPABILITY_GUIDE.md).

## Proposed next finite cycle

| Slot | User outcome | Smallest useful delivery |
| --- | --- | --- |
| 424 — breadth | “I narrowed to one value; let me explore alternatives without losing the rest of my scope.” | A contextual **Edit field constraints** / explicit broadening entry that locates the existing condition or stages a safe whole-conjunct removal. Preview the changed scope; Apply/Run remains explicit. |
| 425 — depth | “How does this selected population differ from the rest?” | Continue a bounded Activity counted-value selection into the existing selected/rest fixed comparison, preserving the original fixed parent question and typed categorical complement. Stage locally; Run acquires both populations. |
| 426 — consolidate | “Keep this investigation and return to the right place later.” | One end-to-end observe/facet/broaden/compare/preserve/reopen conformance journey, correcting demonstrated action/label/return inconsistencies through existing shared controls. |

### 424: explicit outward movement

This is an entry adoption over existing editing and safe query transforms, not a
missing filter engine. The fixed population owner already exposes
`removableFilters` / `removeFilterQueries`; it deliberately treats a nested or OR
condition as one whole conjunct. Activity already has manual condition removal.
Do not invent arbitrary AST surgery that silently drops one arm of an OR.

Start with current qualified string fields and a clearly identified removable
condition. Preserve environment policy, other predicates and any fixed bounds.
Keep same-name resource/attribute fields separate. If a field occurs in mixed or
nested conditions, navigate to the existing editor or refuse automatic removal
with a useful reason. Never replace the entire saved selector through a field
shortcut. When several clauses mention a field, show the exact whole condition
being removed; never silently choose the first match or remove a mixed-field
conjunct. Bind its index and canonical content to the original question lease so
a stale index cannot remove a different condition. An explicit view of the
resulting question must precede the read.

A broader counted result must not be smuggled into 422's editor as though it
belonged to the unchanged applied question. Either explicitly stage/apply the
broader parent first, or design a named fixed child with its own source owner and
held return. Acceptance should include same-field constraints, mixed OR refusal,
missing/empty values, stale actions, keyboard/mobile focus, no local reads and an
actual source check that unrelated constraints survive.

### 425: reuse the existing comparison, not another cohort model

`partitionQueries` already derives selected and rest questions from bounded
categorical intent, including missing-key semantics. `SelectedCount.plan` and the
population owner already distinguish an overlapping held parent from the
logical rest. Reuse those definitions and the shared paired display.

Only offer continuation where a valid fixed parent and supported count profile
exist. Complement only the bounded categorical selection, never the arbitrary
parent predicate. Counted results are scope/input facts, not comparison operands.
A profile or predicate-bound refusal should explain why continuation is unavailable
without creating another cohort model. Preserve its UTC window, qualified field and complete predicate, enforce
combined expression bounds and show the selected/rest intent before Run. Each
operand remains an independent read; matching partitions do not establish source
completeness, significance or causality. Keep the original Activity and field
opener available for a source-free return. A secondary contextual action is
preferable to another always-visible toolbar or top-level screen.

### 426: complete the branch and return contract

Do not add new preservation stores or merge configuration with evidence.
[Recent / Pinned](../RECENT_PINNED_SCOPES.md) already distinguishes staged
configuration from a held visit; [Preserve](../../ui/preservation.mjs) distinguishes
questions/views from an explicit capture reread. Saved capture reopening must
remain source-free. Activity view reopening stages a rolling selector, rather
than pretending to restore an old data buffer.

Exercise one coherent user journey across those existing carriers: narrow,
compare, save the question/view, capture if requested, return through held
parents, then reopen both kinds. Assert the exact scope, read budgets and actual
focus origins. Inspect matched mobile/desktop frames and simplify only proven
terminology/action drift. This is finite conformance and product polish, not a
navigation registry or a new investigation wizard.

## Boundaries and next decision

These proposals are not claims that faceting, comparisons, dashboards,
annotations, history or saved evidence are absent. Native histogram coverage,
monitor delivery, tenancy/presence, arbitrary executable kinds, classifier
training and warehouse lowering retain their separate contracts and gates.

423 should first close its measured performance question. The parent can then
reconcile this proposal with its architecture review and choose the next finite
cycle. If the existing field-edit or selected/rest entry already proves natural
in a real end-to-end exercise, consolidate its adoption instead of adding a
second route merely to satisfy this table.
