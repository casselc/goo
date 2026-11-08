# Product roadmap — checkpoint 381

Adopted after 381 acceptance, reconciled from the
[architecture review](reviews/WHOLE_PRODUCT_ARCHITECTURE_381.md) and
[product/design review](reviews/WHOLE_PRODUCT_DESIGN_381.md). This is the sole active
finite sequence; validation of completed work is recorded in [VALIDATION](../VALIDATION.md). The preceding
[November 8 roadmap](ROADMAP_2026_11_08.md) retains the 376–381 delivery history.

Goo's central loop remains **see current telemetry → isolate a population → follow
supporting evidence → preserve or reuse intent and observations**. After the
predominantly native-focused 376–380 sequence and source-boundary consolidation,
return to the general cross-service investigation loop. Keep authored intent,
fresh results, held parents and immutable evidence distinct.

## Chosen finite sequence

382 and 383 are delivered; the active task is **384**.

### 382 — breadth: preserve and reuse a trace-search question — delivered

Implemented through [saved trace-search questions](SAVED_TRACE_SEARCH_QUESTIONS.md),
with actual checks in [validation](../VALIDATION.md).

User task: “Find traces containing these conditions across services, then keep
that question for the next incident.” Trace search already exists; the missing
capability is preserving its checked intent through the established question
workflow. Its current panel owns one to four condition drafts and clears them on
invalidation; it has no saved-question continuation.

Extend the existing investigation **question card / typed step** boundary with an
explicit trace-search payload. Reuse the current condition editor, existing
investigation revision/catalog ownership and shared library presentation where
needed. Do not introduce another top-level recipe registry, parallel authoring
format, arbitrary JSON query blob, or pretend a cross-span condition set is one
ordinary record predicate. This first cut keeps the original fixed scope. Rolling
windows or scope rebinding must later be explicit operations, not silent refresh.

Acceptance:

- Save a nontrivial cross-service condition set from the existing search flow;
  reopening an exact saved investigation revision stages that whole request.
  Saving intent and reopening perform no telemetry read; Search remains explicit.
- Go/JavaScript validation and serialized/link step dispatch admit exactly one
  typed payload. Original environment policy, fixed bounds, condition order,
  qualified fields, numeric strings above 2^53 and existing search limit survive
  round trips. Mixed arms, foreign scope and unsupported fields fail closed.
- Existing record/measurement question cards retain their behavior. There is no
  new executor or inferred parent/child join. One span may satisfy several
  conditions; a reported trace ID is not a unique invocation or complete trace.
- Use actual ordinary OTel data with several services, keyboard/focus and stale
  callback checks, source-read counters, meaningful generated properties and an
  executable storyboard. The contextual save/reopen gesture stays compact; it
  does not add a large authoring control block before result data.

Source anchors: `ui/trace-search.mjs`, `exploration/trace_search.go`,
`workbench/steps.go`, `workbench/documents.go`, `ui/documents.mjs`,
`ui/workspace.mjs`. Exact transport names and route additions are implementation
choices to review within this milestone, not claims of existing API support.

### 383 — depth: capture search evidence and follow its qualified witnesses — delivered

User task: “Keep what this search actually found, then follow one trace without
losing the search evidence that justified it.” Extend the existing immutable
capture question/result union with a checked trace-search arm. A search result
must not masquerade as an ordinary span inspection page.

Preserve the search request, source read clock, basis, state, full matching-trace
count, bounded items, `hasMore`, qualified witnesses, plan and warnings. Use the
existing server-owned capture workflow: if capture rereads the fixed question,
make that explicit and preserve its new independent read basis. Do not silently
claim the capture is the earlier displayed search or a transactional snapshot.

Acceptance:

- An explicit capture stores actual checked search evidence and links to its
  question card. Immutable reopen works with the telemetry source unavailable
  and retains the original count/list/witness limitations and exact receipt.
- Independent raw records establish condition/trace membership, including missing
  versus empty qualified values, a shared-span match and distinct-service matches.
  Bounded lists never redefine totals or hide `hasMore` as completeness.
- Following a reported trace stages the current exact trace question and requires
  an explicit fresh read. Back restores the same held search DOM, selection,
  citation and focus with no search rerun or capture-store reread. Required cancel
  control traffic remains distinct from source/data reads.
- Actual refusal, pending cancellation and a late successful child response do
  not replace the held parent. Capture bytes and read/custody clocks remain
  authoritative; newly read traces can disagree or expire without rewriting old
  witnesses or implying causality/complete trace capture.
- Actual browser/actor checks and recorded-fact model replay validate the contract,
  with independent negative controls. Reuse the retained-parent pattern where it
  fits; do not create a global navigation manager.

Delivered contract and actual journey: [captured search evidence](CAPTURED_TRACE_SEARCH.md).
Generated cross-language admission exposed and closed a disguised-family hole;
retain that corpus when simplifying family dispatch. Independent condition scopes
must remain separate, and history restores held evidence rather than rereading it.

Source anchors: `exploration.TraceSearchResult`,
`exploration/capture_store.go`, `ui/captures.mjs`, `ui/trace-search.mjs` and the
existing retained fixed-question return in `ui/app.mjs`.

### 384 — consolidate: make fixed-result data the primary content — next

The real 380 desktop frame shows a two-span result; the
[383 fresh-trace frame](journeys/screenshots/383-search-trace-followthrough.png)
confirms the same problem with four spans. Secondary controls,
repeated count/summary content, trace examples and deployment context put its
actual rows below a 1000 px viewport. Treat this as a firm shared-layout task,
not indefinite polish or another local spacing patch.

Consolidate ordinary fixed record results and their Saved/Activity retained
continuations around one compact applied-scope/count presentation, one action
row, then the actual record page or explicitly chosen primary visual. Exact
scope/read facts, examples, deployment context and selection instructions remain
available through contextual disclosures. Preserve origin/Back authority and the
separation between dirty drafts and applied evidence.

Acceptance:

- Matched real desktop and narrow frames show applied scope plus at least the
  first actual record or selected primary visual without navigating past
  secondary chrome. Check geometry inside the real scroll container, not only
  the page viewport; retain native headings and accessible labels.
- Keyboard/focus, no-read kind/disclosure changes, child Back, dirty/invalid editor
  visibility and stale asynchronous publication remain correct. Source queries,
  total counts, independent clocks and immutable origins do not change for layout.
- Use existing tables/lists/waterfalls, scope/evidence summaries and disclosures.
  Do not add a new dashboard manager, universal presenter or query executor.
- Reconcile only typed-step dispatch/round-trip rules that 382–383 actually
  duplicate or expose. Required typed admission belongs in those milestones;
  a broad dispatch framework is not a prerequisite for this layout work.
- Curate one current capability/gap index and link historical proposals as dated
  research. Retire contradicted active claims rather than appending another
  competing backlog. Documentation truth is part of this consolidation.

Source anchors: `ui/exploration.mjs`, `ui/record-results.mjs`,
`ui/resource-return.mjs`, the existing evidence/record components and 380's actual
frames. Verify exact file ownership before editing; this is a bounded shared
layout cut, not permission to redesign every surface.

## Explicit next-cycle candidates, not parallel prerequisites

- **Review-task discovery → review one kept observation.** Review labels and
  historical export already exist. Their concrete gaps are a bounded contextual
  task chooser (no pasted exact task ID) and a disjoint verified kept-occurrence
  target (current review resolver admits Explore capture/index only). These form
  a coherent later breadth/depth pair; do not fake an Explore capture or source
  reread to label a kept item. Reassess after the chosen sequence rather than
  starting a second annotation track concurrently.
- **Log-pattern display grammar.** Log patterns, counts/shares, time comparisons
  and saved scopes already exist. The missing grammar admission is the checked
  `body_pattern_v1` display profile, preserving canonical identity, population
  denominator and unclassified coverage. It is not a missing clustering feature
  or prerequisite for trace-search preservation.
- Native gauge/counter display reuse remains a legitimate later facet-depth item.
  Service exemplar adoption is separate; 380's Saved capability does not grant it.

General exponential evaluation retains its actual source/mathematical gate.
Broad broker/database-server health, alert delivery, multi-user tenancy and
presence, contributed executable kinds/WYSIWYG editing, warehouse lowering and
new storage engines remain independent wider initiatives. None is a prerequisite
for these concrete primary-loop gaps.

## Delivery discipline

Finish and review each acceptance cut before the next. Keep real storyboards and
screenshots current; use Hegel where semantic input generation helps and existing
Quint models with measured facts where ownership transitions fit. Do not invent a
success trace or equate process exit with visible completion. Keep evidence useful
and compact, and distinguish source/data reads from metadata/cancel control.

This prerelease has no compatibility floor requiring legacy arm aliases or
migration layers. Use current Go/stdlib features where they simplify internals;
keep public usage direct. After 384, perform another whole-product checkpoint
rather than automatically promoting every deferred candidate.
