# Proposed operator cycle after CEL consolidation

Status: **adopted after 436.6 acceptance**, 2026-11-14. Subsequently,
[427 exact recall](../EXACT_SAVED_VIEW_RECALL.md) was delivered and validated.
428–429 remain planned. The proposal below records the original sequencing basis.
436.6 completed actual Services/native record-context and held-return acceptance.
The parent maintains executed results in [validation](../../VALIDATION.md).

The next work should help an operator **find a useful investigation, preserve both
sides of a comparison, and return without reconstructing state**. Resume the
existing finite 427–429 contracts with the adjustments below. Do not start another
CEL-platform sequence now that the original six-slice sequence reaches acceptance.
These outcomes apply to ordinary OTel investigations, not just authored CEL or a
sample business domain.

## Why this order still fits

The [430 checkpoint](../ROADMAP_2026_11_14_CHECKPOINT_430.md) deferred preservation
until unfamiliar telemetry could be reached reliably. Its 430–435 slices now
provide observed scope entry, partial-family readiness, metrics-only bootstrap,
sparse-window recovery, stable source refusals, source Body search and numeric
severity. The [vendor gap review](PRODUCT_GAP_ANALYSIS_2026_11_12.md) warned against
rebuilding charts, dashboards and explanation features that already exist. The
[ad-hoc review](ADHOC_OTLP_REVIEW_2026_11_14.md) favored first useful observations;
it did not require a new telemetry receiver or query engine.

CEL 436.1–436.5 strengthens filter admission, exact query identity, recoverable
editor state and source-preserving definitions. It makes an exact saved-view
reference more useful, but does not add that reference to Recent / Pinned or make
selected-question capture preserve a complete comparison. The retained
[427–429 contracts](../ROADMAP_2026_11_13_CHECKPOINT_426.md) remain genuine gaps.
Architecture and UX reviewers independently accepted this proposed finite order
and its owner boundaries; their approval of the proposal is not implementation
acceptance. Their current whole-product findings are recorded in
[architecture](CEL_WHOLE_PRODUCT_436_6_ARCH.md) and
[UX](CEL_WHOLE_PRODUCT_436_6_UX.md); proposed work below is not their execution evidence.

## Next three reviewable outcomes

| Cadence / existing milestone | Operator outcome | Bounded implementation and owner | Acceptance |
| --- | --- | --- | --- |
| **Breadth — 427: exact saved-view recall at the main entry** | “I saved this useful investigation. Find it from Recent / Pinned and stage it intact.” | Add an exact immutable ExplorationView reference to the existing local journal/picker. Resolve through the existing catalog/reader; retain the complete definition, composition, comparison operands and any checked authoring. Do not copy source text, rows, receipts or private request state into browser history. | Exact-ID recall of a composed comparison and a CEL-authored definition. Pin/filter/remove remain local; Stage may fetch definition metadata but reads no telemetry. Fixed bounds and typed environment policy survive. Missing/corrupt/incompatible or old-library references refuse while retaining the current investigation. Both valid-unsaved and invalid local source use the existing owned keep/discard conflict. Pending metadata and route/scope ABA have rejecting controls. Keep returns focus to the retained editor/opener; explicit adoption focuses the new scope heading, never removed controls. Cover keyboard and narrow layout. |
| **Depth — 428: preserve both comparison questions as paired evidence** | “Keep evidence for both sides, reopen it, and share the pair.” | Adopt existing general comparison capture from the fixed population Preserve menu. Keep selected-question capture explicit. Form opening is local; confirmation independently rereads the two fixed questions and publishes the existing child/pair artifacts. | Preview exact operand order, predicates, environment and clocks; one explicit confirmation dispatch. Reopen pair and each child without telemetry reads. Cover unequal/no-data/diagnostic sides, source-limit refusal, second-read/publication failure, withdrawn completion, fresh-browser reopen and actual retained return. A saved successful first child remains useful after partial failure; retry is explicit. Counts must come from complete source reads, not the loaded page. |
| **Consolidate — 429: one compact scope and preservation interaction across those paths** | “The same actions mean the same thing, and I can see what is held, staged or saved.” | Use the exercised 427/428 paths to extract only duplicated pure scope/provenance projections and matching action guards. Reuse shared native controls and the existing preservation shell. Fold redundant fixed service/message/severity entry chrome and duplicated selector/prose where the supplied scope facts match; retain quick-search access and distinguish literal severity text from numeric classification. Keep source acquisition, custody and native measurement owners separate. | Desktop/320/390 journeys through Activity → fixed → exact recall → pair capture → held return. One primary execution action; no query on local disclosure, staging, checking or Return. Deferred success/error, native disabled state, scope ABA and stale save acknowledgement remain guarded. Measure first-record vertical position and visible action count in the same applied/read state before/after at 320/390/1440, keeping edit state separate. Format native rate summaries/axes compactly and record-time constraints as UTC with exact detail, without changing stored values, query bounds, native units or interval semantics. Compare requests and measured work before claiming speedups; refresh frames and actors. |

These remain **427, 428 and 429**, rather than duplicate renumbered milestones.
The active roadmap and implementation introduction adopt this order after
436.6 acceptance. Each milestone should ship with its own review; do not batch the
three into one unreviewable change.

## Reconciliation with recent work

- **427 is not delivered by 436.5.** [Recent / Pinned](../RECENT_PINNED_SCOPES.md)
  still supports Activity selectors, fixed-question configuration and in-memory
  held visits. [`scope-history.mjs`](../../ui/scope-history.mjs) accepts only
  `activity` and `question` intents. Existing saved-view library/import access is
  useful and should be reused, not relabeled as main-entry exact recall.
- **428 is not another capture engine.** [General captures](../GENERAL_CAPTURES.md)
  and `CaptureStore.createComparison` already capture both server-read operands.
  The gap is adopting them from the population Preserve entry. A general pair
  capture (`SavedComparison`) is not the separate versioned
  [saved comparison receipt](../SAVED_COMPARISONS.md). Do not weaken the latter's
  selector compatibility to force selected/rest questions into it.
- **436 source provenance stays definition metadata.** Pair evidence retains its
  immutable typed questions and source receipts. No speculative CEL field is
  added to captures. Applying a display keeps the destination query/provenance;
  Open original intentionally adopts a donor definition. Exact journal recall
  must preserve this distinction and must not restore serialized invocation tokens.
- **429 has a specific polish target.** Current fixed authoring still stacks
  service/message/severity controls, a selector chip, Filters and lower tool
  disclosures. Preserve repeats configuration/no-read explanations. Simplify
  those shared surfaces while keeping environment/window/selection meaning
  visible; do not add another editor or generalized permission coordinator.
- **Semantic row trust is a current acceptance issue, not a new milestone.** The
  436.6 UX review found producer `FATAL` text can make numeric Info records look
  red in common rows. 436.6 corrects the shared presenter and validates actual neutral Info rows.
  Do not plan it again as an undelivered future feature if acceptance fixes it.

## What is deliberately outside this cycle

Generic native stream evaluation without reported `service.instance.id` remains a
real ad-hoc OTel coverage gap, but [433's architecture decision](CONSOLIDATION_REVIEW_433.md)
explicitly separates it from discovery/preflight. Gauge/counter/histogram producer
identity, dimensions, epochs, resets and units need a bounded contract. Do not
remove a service evaluator guard, synthesize instance IDs or call combined streams
a fleet aggregate. Revisit as a separately scoped next-depth decision after this
cycle, with one real non-service metric consumer and independent interval oracles.

Map/presence CEL, whole-predicate conversion, broader lowering/plugin registries,
warehouse compilation, durable monitoring/notifications, team authorization,
multiplayer investigation and classifier training also retain their separate
semantic/operational gates. None is needed to make 427/428 useful. In particular,
this proposal does not revive the older deferred labeling queue by default.

## Review and stopping rule

Before implementing each item, reconcile any 436.6 correctness findings into the
current owner; fix a broken acceptance path rather than scheduling around it.
For 427/428, use real request-observed browser journeys, source totals independent
of retained pages, and native semantic/model checks with targeted rejecting
controls where new authority/custody transitions are introduced. Keep Bombadil
observations positive and nonvacuous. Distinguish source review, screenshots,
physical assistive technology and cross-engine testing rather than treating one
as proof of the others.

After 429, stop and review the product again. The decision should use actual
operator reach, failure recovery, UI density and request/latency measurements;
it should not automatically extend a preservation or language backlog.
