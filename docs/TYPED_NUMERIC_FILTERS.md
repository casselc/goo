# Typed numeric attribute filters

Status: TF1 Go/parity independently approved and delivered; TF2 editor/protocol
implementation follows. Browser controls are not yet accepted in this slice. Existing string predicates and numeric aggregate
bindings retain their current behavior.

## Contract

A leaf declares its own interpretation of an exporter string map:

```yaml
numeric:
  field: {kind: attribute, key: db.response.returned_rows}
  type: uint64
  unit: '{row}'
  parse: canonical-decimal
op: gte
value: '1000'
```

Exactly one of `all`, `any`, `field` or `numeric` identifies a branch/leaf. Numeric
bindings support attribute/resource keys and explicit int64/uint64 types, unit
labels and canonical-decimal parsing. No type/unit is inferred from an OTel name.
The predicate binding is independent of the optional question-level numeric measure.
Adding the omitted optional field does not alter existing JSON/YAML or saved IDs.

| Raw map value | Numeric state |
| --- | --- |
| Absent key | missing |
| Present empty string | empty |
| Canonical decimal within the declared integer type | valid |
| Present nonempty value failing parsing/range | invalid |

Canonical integers have no whitespace, plus sign, leading zeros, negative zero,
fraction or exponent. Limits include exact int64 min/max and uint64 max; strings
above JavaScript's safe integer range remain strings. Units label values without
conversion. Comparisons `eq`, `ne`, `gt`, `gte`, `lt`, `lte` require a canonical
in-range threshold and match **valid values only**, including `ne`. State operators
`valid`, `missing`, `empty`, `invalid` take no threshold. These four states partition
the supplied records; they do not infer original OTLP scalar types or data coverage.

A full complement of a typed comparison includes missing/empty/invalid and the
inverse comparison over valid values. Threshold inversion uses gt/lte or gte/lt,
without increment/decrement at integer extrema. Existing selected/rest gestures
complement their generated categorical intent within the unchanged parent query;
they do not acquire an arbitrary-query NOT operator in this change.

## Layers and gates

1. **TF1 — Go semantics and SQL parity.** Shared canonical parsing and nullable
   SQL guards with existing numeric measures, strict AST/threshold validation,
   reference matching, typed ClickHouse parameters and total boolean results.
   Real ClickHouse parity and Hegel properties/state/swarm gate frontend exposure.
2. **TF2 — Existing editor and continuation.** Compact explicit interpretation
   in the shared condition editor, source-free staging, meaningful typed labels,
   protocol/recovery preservation and unchanged string controls. Time-chart
   selection preserves typed predicates through existing checked scope transforms.
3. **TF3 — End-to-end acceptance.** Real OTLP/ClickHouse boundary-value fixtures,
   exact counts/membership, missing/empty/invalid isolation, saved-link/capture and
   reusable YAML round trips, paused/saved reopening, actual screenshots, existing
   chart gestures, Bombadil and observed Quint rejecting controls. Independent
   review and validated commits complete each gate.

```mermaid
flowchart LR
  Author[Explicit field / integer type / unit / threshold] --> AST[Typed predicate leaf]
  AST --> Go[Go validation and reference matching]
  Go --> SQL[Canonical parse guards and typed SQL params]
  SQL --> CH[ClickHouse record populations]
  AST --> Continue[URL / YAML / saved question / capture]
  Continue --> Go
  UI[Existing editor / time-chart gestures] -->|Staged intent only| AST
```

No floating-point filters, unit conversion, inferred bindings, new evaluation
runtime, production cursor or storage migration is introduced by these gates.
