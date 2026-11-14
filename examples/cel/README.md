# CEL predicate planning

From the repository root with the built `goo` command:

```sh
goo plan-cel -query examples/cel/logs.json -expression-file examples/cel/errors.cel
```

The fixed parent is production logs from 2026-11-14 12:00–12:05 UTC, restricted to
`api` OR `worker`. The CEL extension adds numeric Error (17–20) and the literal
message `timeout` under AND. The command produces a normal Goo question and
parameterized SQL without contacting a source.

The [language guide](../../docs/CEL_PREDICATE_PILOT.md) describes supported typed
fields, explicit unsigned literals, missing-value semantics and limits. The
[executable storyboard](../../docs/journeys/cel-authoring.md) records checks.

To add reported resource and record attributes with the same planning command:

```sh
goo plan-cel -query examples/cel/logs.json -expression-file examples/cel/attributes.cel
```

`eq`, `ne` and `contains` match only reported keys. A missing HTTP method does
not match `ne("http.request.method", "OPTIONS")`; present-empty does. Use
`!("http.request.method" in record.attributes)` to select missing methods.
See [attribute conditions](../../docs/CEL_ATTRIBUTE_CONDITIONS.md).
