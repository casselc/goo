# Whole-project review: Go and system simplicity

Reviewed 2026-11-17 at `9119a967`, after the seven component walkthroughs and
reviews were complete. This is a source-based specialist review, not an
implementation, load test or production reliability certification.

## Judgment

Keep the small Go server and typed, family-owned boundaries. They are not the
main obstacle to understanding Goo. The largest costs are concentrated semantic
cross-products, browser owners with compressed overlapping lifetimes, and a
verification/documentation map larger than the product's ordinary execution map.
A new generic executor, repository framework or public type-state API would make
those costs worse.

An afternoon should let a contributor trace one flow, identify its state and
side effects, and find its rejecting tests. It cannot mean learning every native
profile and historical format. The walkthroughs provide that first path; several
source owners still need the changes below before whole-file reading meets it.

## Verified findings and sequence

### 1. Make executed phases visible before changing abstractions

**Priority: first structural work.** In
[`measurements/source.go`](../../../measurements/source.go), `Source.Run` detaches
and compiles a question; `Evaluate` calls `prepareResult`, which detaches and
compiles again. `Evaluate` then calculates several definition-only profile flags
inside the row loop and mixes original-byte presence admission, decoding and
kernel dispatch. These mechanisms are real, but their coordination is hard to see.
[`exploration/source.go`](../../../exploration/source.go) similarly combines
page-only rewriting, population/row/supporting reads, consistency checks and
result projection in `fetchProfile`.

Extract ordinary named phases, not a dynamic pipeline: admit/freeze; compile;
acquire; decode/profile-check; evaluate; establish result facts. Hoist truly
loop-invariant profile decisions. Keep original raw bytes where null/presence
checks require them; canonical JSON is not a substitute. A private admitted
question/plan value may avoid repeated work within one call, while exported entry
points still admit their own inputs. Do not expose new generic machinery to users.

**Acceptance:** unchanged plans/parameter order, bases, uncertainty states, row
bounds, request order/counts and custody behavior. Mutation/rejection fixtures
must pass before removing any validation. Benchmark repeated compilation and
canonicalization with `testing.B.Loop` before claiming allocation or latency gains.
No current bottleneck is established by this inspection.

### 2. Finish checked identity and envelope boundaries

**Priority: small reliability work alongside phase extraction.**
[`measurements.Hash(any)`](../../../measurements/model.go) discards `json.Marshal`
errors. An unsupported value can therefore receive a digest of empty bytes. The
function is active in discovery, lifecycle, capability/replay and result-basis
paths; it is not dead simply because its comment says legacy. Existing
[`Definition.Revision`](../../../measurements/revision.go) and
[`Question.Key`](../../../measurements/question_key.go) show the clearer API.

Replace remaining families with small typed checked identity functions. Keep
question, authored definition, observed basis and custody identities distinct.
No invalid-current-caller exploit was demonstrated; the proven issue is the
exported permissive contract and invisible encoding failure. Prerelease status
allows retiring that contract instead of maintaining an unnecessary wrapper.

Also reconcile bounded stored-JSON mechanics where appropriate.
[`investigations.Store.Load`](../../../investigations/store.go) already bounds
bytes, rejects trailing garbage through `json.Unmarshal`, validates the question
and verifies a digest; it does **not** reject unknown fields as
[`workbench.Store.read`](../../../workbench/store.go) does. Do not describe it as
an unvalidated store. Choose a deliberate envelope policy, test unknown fields,
then reuse a small decoder if that removes duplication. Do not erase each
family's identity and publication rules.

**Acceptance:** invalid input produces no identity; all retained identity-family
fixtures remain intentional. Any version/encoding change is explicit, not an
incidental consequence of a faster serializer. Corruption/unknown/trailing/size
fixtures describe one policy per stored type.

### 3. Unify ordinary field metadata, not every policy

**Priority: before broader span scope.**
[`exploration/model.go`](../../../exploration/model.go) owns `columns`/`Field.Type`,
[`compiler.go`](../../../exploration/compiler.go) owns physical lowering, and
[`querycel/program.go`](../../../querycel/program.go) separately declares the
operator-visible fixed CEL fields. These overlap enough to drift.

Use a small checked description for ordinary field name, scalar type, signal and
physical mapping, with explicit derived-field handlers. Preserve the narrower
CEL allowlist and helper-overload checks. Do not automatically expose every SQL
field to CEL or send physical SQL syntax into the browser. The planned span-scope
carrier matrix should name owner/carrier eligibility separately from field type.

**Acceptance:** every admitted ordinary field lowers with the expected signal and
type; every exposed CEL declaration belongs to its approved subset. Existing
missing/present-empty, exact-uint and typed-parameter cases remain rejecting or
accepted as designed. A metadata table is sufficient; no schema DSL or generator
platform is justified by this field count.

### 4. Make lifetime and resource policy a small readable map

**Priority: reliability during refactoring.**
[`queryhttp.ReadHandler`](../../../internal/queryhttp/handlers.go) already shares
wire mechanics with inferred Go type arguments while families own effects.
[`newQueryBudget`](../../../cmd/goo/query_budget.go) names those families.
[`httpjson.Decode`](../../../internal/httpjson/json.go) joins a running
`context.AfterFunc` before clearing a transport deadline; that nontrivial code
prevents a real ordering hazard and should not be shortened casually.

Name the nested budgets at their existing owners: ordinary query admission,
separate stream admission, per-read client/ClickHouse bounds, route deadline,
SSE frame deadline and shutdown grace. A longer route deadline may cover multiple
bounded physical reads; it is not automatically a mismatch with a shorter client
request timeout. Share constants only when they mean the same thing.

Use the same map for browser query, visit and publication owners. Their epochs
and abort controllers are not duplicate versions of one universal lifetime.
Cancelling a source read must not become rollback or cancellation of an already
committing publication. The 461 hidden-row restoration bug is a useful lesson:
state belongs to its owner, not to a currently visible DOM button.

**Acceptance:** one page identifies who can read/write, who retires the work and
which bound applies. Parent deadlines never expand; shutdown drains then cancels;
stale/hidden actions remain withdrawn; uncertain writes remain uncertain. Preserve
family route/OpenAPI policy parity. No new global registry or event bus.

### 5. Improve the code contributors must actually read

**Priority: before another broad ownership rewrite.** Go files are mostly ordinary,
formatted functions. Browser owners are much harder to scan: the Activity and
preservation reviews found large closures and source lines thousands of characters
long in `population-workspace`, `captures`, `documents` and Tail.

Apply one pinned development formatter to authored JavaScript, then extract one
cohesive owner at a time by lifetime: execution/result lease, declaration editing,
and durable publication. Keep application composition as direct typed/plain
callbacks; passive components keep supplied-data contracts. Do not spread the
same mutable closure across files and call that decomposition. Similarly, keep
pure kernels separate from source adapters even if merging them reduces imports.

**Acceptance:** a contributor can find state, entry, retirement and side-effect
boundaries without expanding compressed statements or searching every event.
Existing source-count, hidden-action, Return, focus and uncertain-save tests remain
unchanged. Formatting and behavior changes should be separately reviewable.

### 6. Measure catalog/adapter costs, then simplify the expensive path

**Priority: bounded performance investigation, not speculative infrastructure.**
[`workbench.Meanings`](../../../workbench/meaning.go) builds history/head state
before filtering; `ActivityViewsForEnvironment` scans stored definitions to find
page matches. These are bounded file-based catalogs, not demonstrated leaks.
If realistic histories make their read cost material, follow the existing derived
index pattern and retain artifacts/direct IDs as authority and repair paths.

The optional [`retention.Store`](../../../retention/store.go) establishes a pgx
connection and issues schema DDL in publication/correction paths; Tail has related
DDL in [`retention/tail.go`](../../../retention/tail.go). Product should first
decide whether this adapter graduates or remains an experiment. If retained,
explicit schema initialization and measured connection reuse are easier to
understand than per-write setup. A pool is not automatically a win for rare local
operations, and it introduces shutdown/health ownership.

**Acceptance:** scan/index parity against a full-scan oracle, explicit byte/read
budgets, unchanged branches/cursors and direct recall during index failure. For
XTDB, instance-backed transaction/await/retry tests and measured setup costs must
precede a connection strategy change. Core local workflows remain independent.

### 7. Make the maintained test/dependency surface obvious

**Priority: first low-risk contributor improvement.** Root
[`Makefile:test`](../../../Makefile) excludes nested instrumentation and example
modules that CI tests. Provide one truthful default application check entry and
an explicit optional-native tier. Share the module list with CI. Generic journeys
should not require reservation Redis drainage merely because the historical
wrapper does. Keep actual native verdict checks, not silent missing-tool passes.

A current check matrix and the seven walkthroughs should be the entry, with
historical ledgers retained as records. Archive or clearly label superseded
present-tense architecture prose. This reduces unnecessary reading and execution
without discarding meaningful failure controls.

**Acceptance:** an intentional nested-module failure is caught by the local and CI
entry; generic-runtime journeys run without reservation services; required native
checks fail on missing tools; before/after CI duration is measured before deleting
apparently repeated runs.

## Modern Go: use what simplifies the implementation

The module floor is Go 1.27 and `.go-version` validates 1.27.3. The project already
uses method-aware ServeMux routes, inferred generic HTTP handlers, `slices`/`maps`,
`atomic.Pointer`, `context.AfterFunc`, `http.ResponseController`, `errors.Join`,
`testing.T.Context`, `testing/synctest`, and `testing.B.Loop`. This is not a codebase
stuck on an obsolete Go style.

- Prefer concrete structs/functions and inferred generic helpers such as the
  existing route adapters. No public phantom types, reflection registry or
  generic repository is needed for versioned wire/profile admission.
- Use `slices.Sort`/`SortFunc`, `maps.Clone`, or `slices.Sorted(maps.Keys(m))` in
  touched code when clearer. `sort.Strings` is not a correctness problem or a
  performance finding. Preserve stable canonical ordering in plans/identities.
- Shallow `maps.Clone`/`slices.Clone` cannot replace detachment of a nested question.
  Keep a reviewed typed deep copy or the current checked JSON copy until equivalent
  alias isolation is proven. The prerelease floor does not relax ownership.
- Reuse `testing/synctest` for suitable timer/lifetime unit cases and `B.Loop` for
  new benchmarks; retain real HTTP/SSE/browser checks for physical behavior.
  Do not introduce concurrency merely to exercise a newer synchronization API.
- Do not adopt an experimental JSON encoder as a modernization sweep. Exact
  number, null/presence and digest semantics matter here. Likewise, no source
  import of `x/exp` was found in this inspection, but its indirect module entry
  is not proof it can be manually removed from the dependency graph.

## Removal decisions and review limits

No entire Go production package was proven dead. Optional XTDB, operation evidence,
subject-fact warehouse output and declared manifests have active entry points.
Product may choose to remove a prerelease capability; first identify its callers,
unique assertions and stored-carrier consequence. Do not preserve obsolete APIs
solely for compatibility, but do not confuse a purposeful refusal/clock boundary
with legacy scaffolding. Small redundant UI aliases are evidenced by component
reviews; unchecked Hash is active work to replace, not unreachable code to delete.

This specialist independently inspected the cited Go source, module/toolchain
pins, current walk-throughs and component findings. No implementation, new source
campaign, full Go suite or benchmark was run for this documentation-only pass.
Component-executed checks are recorded in their reports; recent system acceptance
remains in `VALIDATION.md`. The proposals above must be validated when implemented.
