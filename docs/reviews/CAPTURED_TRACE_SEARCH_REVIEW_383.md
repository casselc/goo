# Captured trace-search review — 383

Read-only architecture/usability review of the implementation candidate. Runtime
validation is recorded separately in [VALIDATION](../../VALIDATION.md).

The capture store remains the custody owner. A typed search question/result arm
preserves its own source read, complete count, bounded witness list and historical
plan. Reopening checks stored evidence without telemetry or recompilation; refresh
checks the current plan before issuing a new independent read.

The review identified and coordinated fixes for:

- Duplicate JSON objects, case aliases and null coercion acquiring different
  meanings across live search, saved intent and capture. The shared typed request
  boundary now uses exact keys, last-object semantics and a common byte budget.
- Witnesses satisfying the envelope without satisfying their recorded condition.
  Admission checks qualified fields, missing versus empty values, exact numeric
  predicates, environment policy and column/record metadata agreement.
- Held saved-search actions remaining suspended after a failed refresh. Resumption
  rebinds actions to the current lease; retained custody identity survives a child
  visit while hidden actions cannot publish work.
- A native-only citation label and record-count language appearing on search
  evidence. The search family identifies reported trace IDs and independent reads.

Independent conditions remain separate scope members. Following a whole reported
trace retains environment and window plus trace ID; it does not promote a local
namespace or service predicate to a global trace constraint. The original search
remains in the retained citation.

Generated Go/JavaScript admission additionally found a search payload relabeled as
an ordinary capture reaching the UI record-family fast path. The supplied-data boundary
now rejects a search payload under every other family, including explicit null
foreign arms, before those early returns. Direct regressions supplement generated
rehashed mutants.

No additional manager, registry, compatibility adapter or public type machinery
was warranted. The actual browser/actor journey subsequently proves source-offline reopen,
exact held DOM/focus restoration, and refused, canceled and late child reads.
See validation for the independent recorded-fact replay and negative controls.
