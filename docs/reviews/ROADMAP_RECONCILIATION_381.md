# Roadmap reconciliation — checkpoint 381

Status: adopted after 381 acceptance. Source reviews establish no runtime or
benchmark result; those are recorded separately in VALIDATION. The active
execution queue is [the checkpoint roadmap](../ROADMAP_2026_11_08_CHECKPOINT_381.md),
not the union of every older review.

## Why this sequence

The [architecture review](WHOLE_PRODUCT_ARCHITECTURE_381.md) identifies a concrete
break in the general investigation loop: trace search has a typed bounded request
and qualified result witnesses, but its condition drafts disappear on invalidate.
`workbench.Step` admits record/measurement intent only, and immutable capture
questions have no search-result arm. Saving/reopening search intent, then capturing
search witnesses, connects existing useful functionality without another engine.

The [design review](WHOLE_PRODUCT_DESIGN_381.md) identifies two separate annotation
reachability gaps and one demonstrated core layout problem. Review task selection
requires an exact ID; Activity Keep cannot enter the Explore-capture-only review
resolver. Those are valid candidates, but immediately scheduling them would create
a second breadth/depth track beside search preservation. They remain an explicit
paired next-cycle candidate, with their honest kept-evidence contract intact.

The inspected 380 frame is stronger evidence for near-term consolidation than a
hypothetical generic dispatch framework: it presents a real two-span result, but
records are below the desktop fold. The chosen sequence therefore takes the
architecture's search breadth and witness depth, then the design's data-first
fixed-result consolidation. Typed dispatch admission is required during 382–383;
only actual repeated validation/round-trip logic merits later deduplication.

The resulting finite queue is **382 saved search intent → 383 captured search
witnesses/held trace follow-through → 384 data-first fixed-result layout and
current capability truth**. It returns to general-purpose investigation after
376–380's largely native work, preserves the requested breadth/depth/consolidate
cadence, and gives the visual problem a committed place rather than “later polish.”

## Source-grounded scope decisions

| Finding | Decision |
| --- | --- |
| `ui/trace-search.mjs` owns 1–4 condition drafts and clears them in `invalidate()`; no save/reopen continuation | 382: extend the existing investigation question-card/typed-step boundary. No separate recipe registry. |
| `exploration.TraceSearchResult` has request/read basis/count/hasMore/qualified witnesses; capture union lacks that result kind | 383: explicit search capture arm and server-owned custody; no ordinary-record-page disguise. |
| Existing retained parent/child navigation already proves zero-data-read Back and cancellation/late-result retirement | Reuse those owners where the checked search relation fits; do not invent a manager. |
| Actual 380 fixed-record frame shows secondary chrome before any record row | 384: common data-first result layout, matched-frame acceptance and unchanged query/custody semantics. |
| `ui/capture-record-review.mjs` exact-ID task load; no task list route | Deferred next-cycle breadth candidate, not missing review UI. |
| `workbench/capture_review.go:evidenceFrom` admits Explore capture/index, not exact kept receipt occurrence | Deferred paired depth candidate; never fabricate an Explore capture or aggregate population. |
| `body_pattern_v1` explicitly excluded from ordinary reusable view declaration | Deferred concrete grammar-profile gap; log patterns/time comparisons themselves are delivered. |

## Retire contradicted historical claims

Current source/contracts and executed validation take precedence over older
proposal prose. Root's immediate current-gap cleanup should remove obvious 379/380
contradictions; 384 should make the broader capability index reliably current.
These are documentation reconciliation targets, not new feature assignments:

- Broader Services exponential-family exclusion is closed by 379. Evaluation
  remains separately deferred by the source/representation gate.
- Saved exemplar continuation is delivered by 380. Service widgets remain
  intentionally interval-only; do not erase that distinction.
- Environment-wide-only operation inventory claims contradict the exact service
  selection/capture contract already implemented by 349/350 and later checks.
- Record review labels/training export are implemented; old annotation design
  introductions must not schedule them again as missing. Task discovery and kept
  target admission are the actual remaining reachability gaps.
- Activity, facets, trace search/waterfalls, log patterns, service/resource defaults
  and reusable ordinary record views are present. A vendor category list is not
  evidence that another implementation is needed.
- Dated entry-point and grammar proposals remain research. Mark their status and
  link current contracts instead of letting them act as parallel execution queues.

## What this reconciliation does not claim

381's bounded decoder still needs its own actual acceptance before this proposal
is adopted. 382–384 are planned; no trace-search Step/API/capture payload is
implemented by these documents. There is no new source engine, universal query
schema, complete causal-trace promise, collaboration/tenancy authority, executable
kind plugin system or general exponential evaluator in the chosen scope.

Capture semantics must be explicit: the existing capture owner may reread the
original fixed search to create evidence with a new read basis. Once captured,
reopen is source-free; fresh expanded traces remain independent and may disagree
or expire. “No data reread on Back” does not forbid necessary metadata/cancel
control traffic. The tests must name and measure the actual contract.
