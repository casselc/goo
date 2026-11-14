# CEL attribute conditions: architecture review

## Contract

The admitted vocabulary is deliberately bounded to the existing string-valued
record grammar. `resource.attributes` refers to resource attributes;
`record.attributes` refers to log or span attributes for the selected signal.
Literal dotted OTel semantic-convention keys remain complete keys, with no
implicit traversal, alias or old/new convention fallback.

Map-member `eq`, `ne` and `contains` lower to ordinary present-only attribute
leaves. All three return false when the key is absent, including `ne`. Standard
literal-key membership lowers to `exists`; negated membership lowers to
`missing`. Empty reported values therefore remain distinct from absent values.
Bare indexing, dynamic keys, arbitrary negation and numeric coercion are outside
this vocabulary. In particular, negating a present-only comparison would include
missing records and cannot silently become the ordinary `ne` operation.

Scoped helpers avoid a guard-dominance proof system or special boolean rewrites.
Each helper has one exact leaf representation; inverse printing must retain the
existing structural check, including branch order and nesting. The checked CEL
program remains a reference evaluator. The ordinary predicate AST and source
adapter continue to own SQL execution; no second query engine is introduced.

## Review checks

Review resolved overload admission, literal-key/value validation, ordinary AST
budgets, bounded reference input and CEL cost estimation. SQL keys and values
must remain typed parameters through the existing source compiler. Reference
missingness must be represented by map absence rather than a fabricated empty
string, while malformed supplied text remains an error.

The library must not infer original SDK numeric or boolean types from exporter
strings. This extension establishes neither source availability nor readiness;
checking remains source-free, with execution authorized by existing explicit
Apply/Run actions and the owning revision-bound editor.

## Acceptance

Implementation source review **accepted**. Resolved overloads are allowlisted;
only the two declared map identifiers are accepted as receivers, and keys and
comparison values must be string literals. The resulting leaf still passes
ordinary predicate validation, including key/value and node/depth budgets.
No physical-column names or source authorization move into CEL. Existing source
compilation supplies `mapContains` and typed parameters for the same leaf.

Reference activation includes only referenced keys, preserves absence as absence,
and validates all referenced supplied text before evaluation, including branches
that would short-circuit. Custom helper cost estimates conservatively bound
string work; expression, field and runtime CEL limits remain in force. Exact
inverse printing recompiles and checks typed predicate structure as before.

Independent `go test ./querycel -count=1` passes on Go 1.27.3 (**0.312s**).
A separate exported-API probe agrees with ordinary `exploration.Match` for absent,
empty and present values across comparison/presence forms, including an explicit
missing-inclusive inequality. It rejects literal-map receivers, dynamic keys or
values, bare indexes, arbitrary comparison negation and unrelated membership.
No architecture blocker remains for this bounded extension. Generated/source
parity, browser behavior, screenshots and final product acceptance remain
separately reported by their owners; this review does not claim those ran.
