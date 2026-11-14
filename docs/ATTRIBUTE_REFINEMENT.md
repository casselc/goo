# Refine by an attribute

Inspect a record and open its field actions, or open **Filter** on a custom
record column. The same actions stage ordinary predicates in Activity and fixed
exploration; population facets retain their selection tray and explicit execution.
No field action below reads telemetry.

| Action | Matching records |
| --- | --- |
| Include | The exact reported string, including a reported empty string |
| Exclude | Everything except that value, including records missing the key |
| Is reported | The key is present, including an empty value |
| Is missing | The key is absent |

Resource and record attributes remain separate namespaces. A dotted OTel key is
one literal key. Actions preserve the current parent scope; **Apply**, **Run** or
opening the staged population remains explicit. **Filter and run** and **Break
down and run** are separate, explicitly executing actions.

Missing custom-column cells offer the two presence actions. Include/Exclude
require a reported string; an unavailable projection, a typed null or a numeric
JavaScript value is not silently interpreted as a missing string. Observed field
suggestions and a categorical facet's **Field filters** offer presence actions
even when the supplied page or counted partition contains no missing values.
Suggestions do not establish source completeness.

## CEL meaning

[CEL attribute helpers](CEL_ATTRIBUTE_CONDITIONS.md) use the same predicate AST.
For `record.attributes` (or `resource.attributes`):

```cel
// Include
record.attributes.eq("http.route", "/health")
// Exclude: the complement includes missing keys
!("http.route" in record.attributes) || record.attributes.ne("http.route", "/health")
// Is reported
"http.route" in record.attributes
// Is missing
!("http.route" in record.attributes)
```

`.ne()` alone deliberately matches only reported unequal values. The condition
editor's **does not equal** operator keeps that meaning. Exclude is a categorical
selection complement, not a rename of `.ne()`.

## Shared implementation

`ui/field-refinement.mjs` supplies pure action metadata, validation, canonical
selection intent and predicate construction. Presence maps to existing
include/exclude-missing selection intents; there is no new wire format or query
executor. Attribute details, custom columns, observed suggestions and categorical
facets consume it. Activity, fixed exploration, categorical selection and
explanation keep their separate scope, revision and acquisition owners.

The native shared action disclosure keeps Copy primary, with keyboard Escape and
focus return. Disabled, inert, hidden, closed or retired controls cannot stage.
Admission and predicate-budget checks happen before mutation; a refused action
cannot retire a checked condition or alter a draft. Existing authored-condition
conflict rules remain in force.

See the [executable storyboard](journeys/cel-attributes.md) and
[validation](../VALIDATION.md). Typed numeric CEL attributes remain a separate
follow-on; this change performs no implicit coercion or numeric parsing.
