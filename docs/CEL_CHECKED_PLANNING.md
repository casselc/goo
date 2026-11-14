# Checked record planning

436.2 adds a pure planning operation shared by `goo plan-cel` and the Go server's
`POST /api/cel/plan`. It checks an additional CEL condition against a complete,
immutable ordinary record question. It does not read telemetry, probe readiness,
write storage or execute CEL over the visible record page.

## One operation, two callers

```mermaid
flowchart LR
 CLI[goo plan-cel] --> P[querycel.PlanRecord]
 HTTP[POST /api/cel/plan] --> P
 P --> B[Validate base and exact binding]
 B --> C[Parse and typecheck CEL]
 C --> L[Lower supported predicate]
 L --> R[Refine complete original parent]
 R --> T[Compile parameterized ClickHouse plan]
 T --> O[Detached planning receipt]
```

Go callers use `querycel.NewPlanRequest(base, expression, slot, epoch, revision)`
and then `querycel.PlanRecord(request)`. The constructor detaches the base. The
planner independently validates the supplied base and resulting question; neither
requires an environment, clock or source manufactured by the compiler.

The result includes the ordinary predicate, refined question, parameterized target
plan and exact binding. The existing CLI flags remain `-query` and exactly one of
`-expression` or `-expression-file`; its output now includes that same binding.

## Binding and ownership

Every request declares:

| Field | Meaning |
| --- | --- |
| `operation` | `record.refine`, an additional owned condition |
| `signal` | Same record signal as the supplied base |
| `baseIdentity` | SHA-256 of validated typed base JSON |
| `slot`, `epoch`, `revision` | Nonempty bounded owner correlation tokens |
| `library`, `compiler` | Exact current checked language/compiler versions |
| `target`, `profile` | `clickhouse`, `fixed-record/1` |

Revision is a string, avoiding JavaScript integer rounding. Base identity preserves
predicate ordering and nesting; it does not claim algebraic equivalence or reuse
another owner's normalized identity. A changed base refuses rather than silently
rebasing an expression. Wrong library, operation, signal or target also refuses.

These tokens are correlation, not authorization. The future editor must match the
whole originating request binding before accepting a response, including an error.
Error envelopes do not echo the binding: callers correlate them with their
originating request. HTTP transport
alone cannot know whether a mounted editor has retired. No server session or
slot-replacement reducer is introduced here. Rechecking an edit must use its
original base, not repeatedly refine the last result. That owner is delivered in 436.3; 436.4 now mounts it in existing Filters.

## Diagnostics

`errors.As` exposes `*querycel.Diagnostic` with stable code and phase, a human
message, optional domain path and optional source span. Parse, typechecking,
unsupported operations, input/work budgets, binding and target admission are
separate refusal stages. No failure returns a partially admitted question.

Spans use zero-based Unicode code-point offsets `[start,end)`, with one-based
line and column. They are not byte or browser UTF-16 offsets. Some CEL nodes only
retain an anchor; a zero-width span honestly marks that point. Missing spans mean
no trustworthy expression location is available. The first CEL issue supplies
the structured location; its message may include further compiler issues.
Upstream CEL parser/checker resource refusals retain syntax/type classifications
because that API supplies no typed resource category. Goo-owned input, lowering,
admission and cost limits have explicit budget identity.
Unsupported branches refuse even when a Boolean expression would short-circuit.

The HTTP transport accepts one bounded JSON object with no unknown fields or
trailing value and returns a structured diagnostic when semantic checking fails.
Successful checks express static target support, never live source availability.

## Remaining work

436.3 delivers honest representation coverage and revision-safe slot ownership;
[436.4](CEL_CONTEXTUAL_EDITOR.md) mounts compact contextual editing. Existing
attribute predicates in the base remain intact, but attribute-map CEL expressions
still need a presence contract. Native metrics retain their independent grammar.

See [integration design](CEL_PRODUCT_INTEGRATION.md),
[planning storyboard](journeys/cel-checked-planning.md) and
[validation](../VALIDATION.md).

## Representation and owner integration

[436.3](CEL_REPRESENTATION_OWNERSHIP.md) adds bounded exact predicate printing and
a pure authoring owner that consumes bound planning receipts. It correlates errors
through their originating check tokens and independently validates result meaning.
The owner remains separate from the mounted [editor](CEL_CONTEXTUAL_EDITOR.md)
and the execution controller.
