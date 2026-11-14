# CEL predicate pilot acceptance

Status: accepted. [Guide](../CEL_PREDICATE_PILOT.md),
[CLI storyboard](../journeys/cel-authoring.md).

The independent reviewer inspected the final source and executed nine focused
test functions across the library and CLI. The implementation also has a passing
external Go example. No blockers remain.

## Accepted boundaries

- The shipped OTel helper is an actual publicly registerable, versioned CEL
  library. Typed signal-specific variables and resolved overload identities drive
  lowering into existing Goo predicates; expression text is never source SQL.
- Unsupported operations and subtrees refuse before reference-program creation.
  No short-circuit branch can hide an unsupported function, cast or comprehension.
- Applying the expression deep-copies the question and ANDs the whole existing
  selector, preserving clocks, environment policy, projections and delivery limits.
  Returned objects cannot mutate either compiled representation.
- Nil and zero programs fail safely. Reference inputs require canonical typed
  fixed columns; missing/malformed inputs remain errors. NUL record data retains
  its value while NUL literals follow the existing predicate refusal.
- Source/parser/AST/input/work limits apply before execution, with expanded Goo
  predicate budgets enforced again after parent composition. The final work limit
  is 1,000,000 CEL units. The shipped host binding is pure and constant-work.
- CLI authoring reads are bounded independently of file metadata. Successful
  output is a planning receipt; refusals emit no partial receipt or source request.

## Executed parity and limits

Hegel properties cover typed CEL/Goo equivalence, literal handling, uint64 extremes,
whole-parent membership and mutation isolation. Exhaustive code/band checks and
real read-only ClickHouse execution compare four expressions over 256 inline
records each. The actual CLI output also passes the existing UI question validator
and semantic link encode/decode; this is not a browser text-editor implementation.

No app UI or HTTP wire model changed, so no new screenshots, Bombadil driver or
OpenAPI generation is claimed. The pilot is stateless, so additional Quint state
modeling is unnecessary. Attribute-map presence, a general extension/lowering
registry, app text editing, richer units and Databricks record lowering remain
separate acceptance work. See [validation](../../VALIDATION.md) for actual results.
