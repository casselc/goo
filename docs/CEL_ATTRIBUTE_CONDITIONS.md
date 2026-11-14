# Attribute conditions with CEL

Use **Filters → Expression (CEL)** to add a condition over reported resource or
record attributes. **Check condition** stages the ordinary filter; **Apply** or
**Run** still controls when Goo reads telemetry.

| Expression | Matches |
| --- | --- |
| `resource.attributes.eq("service.namespace", "payments")` | Reported resource key with that exact value |
| `record.attributes.ne("http.route", "/health")` | Reported record key with a different value |
| `record.attributes.contains("exception.message", "timeout")` | Reported record value containing that case-sensitive literal substring |
| `"http.route" in record.attributes` | Key is reported, including an empty value |
| `!("http.route" in record.attributes)` | Key is missing |
| `record.attributes.eq("http.route", "")` | Key is reported with an empty value |

`resource.attributes` selects resource attributes. `record.attributes` selects
log attributes for logs and span attributes for spans. A dotted OpenTelemetry
semantic-convention key is one literal key, not an object path. Goo does not
substitute another convention version or search both maps automatically.

See [attribute refinement](ATTRIBUTE_REFINEMENT.md) for the matching record, column
and facet actions. UI **Exclude** includes missing keys; CEL `.ne()` alone does not.

## Missing is not empty

Attribute `.eq`, `.ne` and `.contains` match only reported keys. All three return
false when the key is missing. In particular, `.ne("key", "value")` is not the
complement of `.eq("key", "value")`: missing records match neither.

To include missing keys deliberately, compose the conditions:

```cel
!("http.route" in record.attributes) || record.attributes.ne("http.route", "/health")
```

General negation of comparisons and bare map indexing are not admitted. Use the
presence forms above and the reviewed comparison helpers. Keys and comparison
values are bounded literal strings; computed keys, map traversal and arbitrary
functions do not gain source-query capability.

These exporter maps contain strings. Comparing `"500"` compares text; it does
not recover an original SDK number. Numeric ordering, parsing and implicit
coercion are not part of this vocabulary. Existing explicitly typed numeric
conditions retain their own grammar and validity rules.

## One condition, one query model

The helpers compile to the same resource/attribute predicates used by Conditions
and the ClickHouse adapter. Presence uses reported key membership; it never
uses an empty string as a stand-in for missing data. Exact supported attribute
predicates can be printed back into this CEL vocabulary without dropping leaves
or rewriting their branch structure.

Checking stays source-free and bounded. Invalid text remains editable with a
located diagnostic. Existing parent scope, fixed or rolling clock, condition
ownership, saved-definition provenance and explicit execution boundaries stay
in place. Unsupported visual predicates are kept as visual predicates; printing
never returns a partial expression.

See [contextual editing](CEL_CONTEXTUAL_EDITOR.md),
[representation ownership](CEL_REPRESENTATION_OWNERSHIP.md) and
[validation](../VALIDATION.md) for executed checks.
