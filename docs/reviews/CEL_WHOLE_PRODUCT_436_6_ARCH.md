# 436.6 architecture and Go review

The bounded CEL architecture is accepted at source and focused-test level. Whole-
product completion additionally requires the current real Services/native-context
journeys and request-observed model replay recorded in [validation](../../VALIDATION.md).
This review does not substitute prior captured journeys for those new executions.

## Architectural assessment

The implemented path has one executable meaning:

`exploration` admission/refinement → checked `querycel` planning → ordinary query
→ existing source execution → existing retained result/custody owners.

CEL is a bounded authoring representation, not another executor or query optimizer.
The browser never evaluates filters against its retained page to claim population
membership. Parameterized ClickHouse compilation remains the actual target. No
DataFusion layer, universal field registry or plugin platform is needed to finish
this investigation loop.

- Ordinary predicate admission and typed detached cloning precede refinement.
  The complete original parent remains under AND, including unsupported attribute,
  presence and OR conditions. Paging, environment policy, projections and clocks
  are not silently rewritten by generic refinement. Depth/node budgets also bound
  cyclic input before recursive copying.
- Planning validates the complete base and result. The Go-computed base identity,
  checked overload lowering and condition digest avoid a second browser hashing or
  evaluation implementation. Content identity preserves admitted order/nesting;
  it does not assert Boolean equivalence or complete/simultaneous observation.
- The two planning routes accept no source/store dependencies. Strict body and
  compiler budgets bound work; cancellation withdraws publication rather than
  pretending to interrupt synchronous compilation. Static target support is not
  source readiness. Preserve this distinction in UI copy and observability.
- `AuthoringSlot` owns private invocation/staging leases separately from execution
  generations and serialized provenance. Exact source, binding, complete-question
  conservation and actual native-control availability are checked before handoff.
  Abort is cleanup; it is not the stale-reply proof. A→B→A, hidden, disposed and
  replaced-parent callbacks must remain rejected even when text matches again.
- The existing `ExplorationView`/`ViewArtifact` owner validates preserved source
  through `querycel` without an `exploration` import cycle or caller-supplied
  validation callback. Recompilation compares the entire resulting query. Source,
  typed meaning, parent identity and immutable definition identity stay distinct.
  Formatting or provenance-label changes can produce a new definition without a
  new population. Labels are not authentication or signatures.
- Restoring metadata never restores a private execution lease. Raw source survives
  native textarea line-ending display normalization until actual edits. Valid
  unsaved source and invalid text have explicit recall choices; only an exact
  current acknowledgement can exempt self-recall. Display reuse keeps destination
  question/provenance; Open original intentionally adopts the donor definition.
- The public display graph remains passive. Pure provenance shape/contract code
  is shared without importing authoring/editor controllers. Source acquisition,
  navigation, publication and retained evidence remain application concerns.

The current use of concrete Go structs, `errors.As`/`errors.Is`, `slices.Clone`,
context cancellation and bounded explicit copying is appropriate for Go 1.27.
There is no demonstrated reliability or usability benefit from replacing these
closed types with generic callback or type-machinery APIs. Keep canonical identity
changes explicit; switching serialization packages merely for novelty could change
immutable content addresses. No performance improvement is claimed by this review.

## Services and native measurements: acceptance obligations

These paths add record conditions; they must not reinterpret native kernels.

`NativeRecordContext` already constructs an enclosing millisecond query with
exact `timestamp_ns >= fromNS` and `timestamp_ns <= toNS-1` guards. Its ordinary
parent includes the qualified producer: reported service, exact namespace state
and instance (or the explicit resource-bound path). The stream's metric dimensions
are not copied into record filters. Refinement must retain this complete parent,
not merely the enclosing millisecond window. Shared time/resource context does not
establish metric contribution, trace ancestry, causality or complete capture.

The native parent uses a shallow flat conjunction. Its inability to print as an
exact CEL inverse does not prevent adding a separately owned condition. Do not
flatten it, remove qualifiers or relax the global predicate budget to make a text
conversion appear possible.

The new browser/model acceptance must establish:

1. Real Services/native entry stages exact record intent without a telemetry read.
   Check/stage is pure; the single explicit Run owns fresh execution.
2. Parent environment policy, namespace/producer and exact interval guards survive
   additional conditions, including boundaries finer than a millisecond.
3. Independent source counts exceed the retained page where claimed. Numeric Info
   with literal FATAL text stays Info in both measurement and human-facing styling.
4. Return restores retained original Services/native data and opener without source
   reads. Capture/citation facts remain separate; saving a record definition does
   not invent a durable native-evidence association.
5. The observed positive trace admits and mutations for hidden reads, dropped
   parent, changed bounds, stale authority, substituted severity, page/source total
   and provenance are rejected. An actor needs positive observation witnesses;
   an all-zero counter alone is not evidence of a source-free action.

The semantic row-styling finding from the UX review belongs in this acceptance
round: producer text cannot promote a numeric Info/unknown record into an error
population visually. It does not require a new predicate or source grammar.

## Independently executed checks

Managed Go 1.27.3, existing module/build caches, CGO disabled:

```sh
go test ./exploration ./querycel ./workbench ./internal/celhttp -run 'Test(Refine|CloneQuery|PredicateAdmission|PrintPredicate|CompileDiagnostic|RecordPlanning|CELAuthoring|RecordContext|PlanRoute|NativeRecordContext)' -count=1
```

All four packages pass, zero skips: **0.014s / 0.065s / 0.137s / 0.073s**.
These exercise admission, preservation and pure planning/transport contracts; they
are not a fresh complete Go, ClickHouse or browser run.

```sh
node --experimental-test-module-mocks --test tests/cel-authoring.test.mjs tests/cel-expression-editor.test.mjs tests/cel-preservation-integration.test.mjs tests/cel-provenance.test.mjs tests/native-record-navigation.test.js tests/display-modules.test.mjs
```

All **six file runners** pass, zero failures/skips, **0.683s**. These include
private lease/provenance behavior, production callback fixtures, retained native
navigation and the passive display graph. Mocked/controller checks do not establish
physical assistive-technology, cross-engine, real telemetry or production capacity.
The parent records fresh runtime/journey/model execution separately.

## Recommended next operator cycle

After 436.6 succeeds, resume the retained **427 → 428 → 429** contracts rather
than extending CEL into another platform backlog:

- **427, breadth:** find and stage an exact saved view from Activity Recent/Pinned.
  Recall its immutable reference through the existing catalog; retain both queries,
  composition and authoring. Do not duplicate source/rows/private state into local
  history or flatten a composed view into a single selected query.
- **428, depth:** adopt the existing whole-pair comparison capture from population
  Preserve. Keep selected-question capture distinct; confirmation rereads both
  fixed operands. Partial child publication, clocks and explicit retry retain the
  existing custody contract. This is not another capture engine.
- **429, consolidate:** use those real paths to extract duplicated pure scope and
  preservation projections, compact redundant fixed shortcuts/prose and unify
  matching guards. Keep native acquisition, evidence custody and return owners
  separate. Measure request/work/latency before claiming speedups.

The [post-CEL proposal](ROADMAP_AFTER_CEL_436_PROPOSAL.md) accurately reconciles
these gaps. Generic instance-free native evaluation, map/presence CEL, warehouse
compilation, monitoring, collaboration and classifier training keep separate
semantic decisions. None is a prerequisite for making exact recall and paired
preservation useful. Review the whole product again after that finite cycle.

## Parent-run completion

The final browser journey now requires the existing-model native gate after
writing the complete report and closing its runtime. Actual Services histogram
and Measurements counter paths pass with 115 source records, 112 selected and
100 displayed, retained native snapshots/focus and no reads on local/return steps.
The native gate rejects ten semantic/authority controls. These executed results,
the preservation regression and full-suite checks are recorded by the parent in
[validation](../../VALIDATION.md); this paragraph does not claim an independent
repeat of those integration runs. The post-CEL 427–429 cycle is adopted there.
