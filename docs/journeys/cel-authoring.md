# Plan an ordinary Goo question from CEL

This is a CLI/library storyboard. No browser interface changes in this pilot, so
it has no new UI screenshots or Bombadil actions.

1. Read the [fixed parent question](../../examples/cel/logs.json). Its service
   selector is `api` OR `worker`; environment, clock and delivery limits are explicit.
2. Author the [CEL predicate](../../examples/cel/errors.cel), using the versioned
   `otel.severityBand` extension and a literal message match.
3. Run `goo plan-cel -query examples/cel/logs.json -expression-file examples/cel/errors.cel`.
   The command typechecks, admits supported overloads, lowers the expression and
   ANDs it with the entire parent. Planning reads no telemetry or store.
4. Inspect the planning receipt. It names the library revision, original text,
   ordinary lowered predicate and full question. The source plan uses typed value
   parameters, including punctuation or SQL-looking message text.
5. Replace the expression with a regex call, signed literal, unsupported field or
   over-budget expression. Planning refuses and emits no partial receipt. A
   supported short-circuit branch cannot hide an unsupported operation.
6. Evaluate well-typed normalized records with the actual CEL reference program
   and ordinary Goo interpreter. Check immutable parent scope and returned-AST
   ownership. A missing referenced field returns an error, not a synthetic value.
7. Compare four admitted expressions over 256 inline records using the real
   ClickHouse compiler/executor. The tests are read-only: they create no source
   table, telemetry record, runtime or investigation store.

Executable acceptance is in [CLI tests](../../cmd/goo/plan_cel_test.go),
[library tests](../../querycel/program_test.go),
[external example](../../querycel/example_test.go) and
[native properties/live parity](../../verification/querycel_test.go).
See [validation](../../VALIDATION.md) for actual results.

The next user-interface step is planned: stage the ordinary output question through
the existing draft editor, with explicit Run and held Return. A text editor,
HTTP compiler route and visual/CEL round trip are not implemented by this pilot.
