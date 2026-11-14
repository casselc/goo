# 436.2 independent architecture and Go acceptance

Accepted for the bounded checked record planning slice. No open correctness
findings remain in the reviewed production code. This is acceptance of pure
planning and its in-process transport, not an app CEL editor or whole-product
436 acceptance.

## Reviewed contracts

- `querycel.PlanRecord` owns Compile → complete-parent refinement → actual
  parameterized ClickHouse compilation. CLI and HTTP delegate to this operation;
  no source, readiness, store or application-clock dependency is accepted.
- Original base validation precedes identity hashing and lowering. Content identity
  preserves typed JSON order/nesting and is explicitly distinct from population
  equivalence or the UI's normalized identity. Result validation remains independent.
  The constructor detaches the supplied base; returned meaning does not alias it.
- Operation, signal, base content identity, slot, epoch, revision, library,
  compiler, target and profile are checked. Revision is a bounded string so exact
  browser correlation does not depend on JavaScript integer precision. Tokens
  confer no execution authority and do not establish current mounted ownership.
  Replacement and stale-response acceptance remain the next owner-state slice.
- Semantic failures expose `*querycel.Diagnostic` through `errors.As`, with
  optional domain path and Unicode code-point location. Zero-width AST anchors
  do not pretend to identify full source extents. Unsupported dead branches refuse.
  The ordinary grammar budget has an owning `ErrPredicateBudget` sentinel and
  explicit lowerer budget errors; classification does not inspect error prose.
  Upstream CEL parser/checker limits retain their syntax/type category because
  that API provides no typed resource category.
- `celhttp.Register` takes no source/store arguments. Strict one-object decoding,
  32KiB transport bound, two-second context, post-computation cancellation check,
  no-store responses and structured semantic refusal preserve the pure boundary.
  Synchronous compilation is bounded by compiler budgets, not asynchronously
  preempted at the HTTP deadline. Malformed transport returns an ordinary error;
  no failure emits a partially admitted question/plan/binding.

## Independently executed checks

With the managed Go 1.27.3 toolchain, existing workspace caches,
`GOTOOLCHAIN=local`, `GOMAXPROCS=2`, `CGO_ENABLED=0`:

```sh
go test ./querycel ./internal/celhttp ./cmd/goo ./exploration \
  -run 'TestCompileDiagnostic|TestRecordPlanning|TestPlanCEL|TestPlanRoute|TestCEL|TestPredicateBudget' -count=1
```

All four packages pass: 0.026s, 0.014s, 0.032s and 0.012s respectively; no skips.
These checks cover typed parse/type/lowering/admission failures, Unicode locations,
combined budgets, binding refusal, detached parent preservation, parameterized
literals, CLI delegation, strict transport refusal and the generic runtime path.
The initial narrower planning/CLI run also passed before final corrections.

Review findings were corrected: transport now bounds computation/publication with
its context, and compiler-owned budget classification no longer depends on error
message wording. The generic runtime test was strengthened to reject/count every
source HTTP attempt, deny source-query budget and require the storage path to
remain absent while both `/cel/plan` and `/api/cel/plan` succeed. Its final strengthened version
was independently rerun and passes (0.026s, no skips).

The [guide](../CEL_CHECKED_PLANNING.md) distinguishes fixed-base content identity,
correlation, static target support and subsequent editor work. No new visual
surface is introduced. This independent report does not claim a real browser
journey, physical assistive-technology check, fresh ClickHouse parity run or
OpenAPI generator execution; those checks belong to their separately recorded
owners.
