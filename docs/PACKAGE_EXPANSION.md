# Explicit HTTP package binding and expansion

The Go package layer now expands the recognized `otel.http.server.spans` profile
into goo's existing typed measurement model, an aggregate view descriptor, an
investigation step, and ClickHouse/Databricks aggregate plans. It accepts an
explicit YAML binding and a versioned inventory. Other profile expansions remain
unsupported and fail explicitly.

This is the first executable bind/expand slice of the broader
[semantic-package design](SEMANTIC_PACKAGES.md). It is independent of the sample
reservation UI and its Python view branches.

## Author-owned choices

See [the example binding](../examples/packages/http.yaml). The author supplies:

- The exact profile and resource scope from the inventory.
- Prepared-table names for supported backend targets; missing target bindings
  produce an `unbound` diagnostic and no plan for that target.
- Failure-status ranges, with no implicit default. The example chooses 500–599;
  an application can explicitly include selected 4xx responses instead.
- `missingStatus: unknown` and `missingRoute: unknown-bucket`. Neither missing
  telemetry nor an absent route is silently discarded from the population.

Ranges must be nonoverlapping final response statuses within 200–599. A missing,
malformed/out-of-range or informational status stays unknown in this subset.
A missing method also makes the classification unknown. Known statuses outside
explicit failure ranges count as observed success; this does not define the
business operation's eventual outcome. Transport errors without response status
remain unknown; a later policy can explicitly support that evidence.

## Generated artifacts

`goo expand` emits:

- A normalized `measurement.Definition`: scoped trace/span identity, route/method
  dimensions, requests/successful/failures/unknown counts and `success_ratio`.
- A population predicate selecting the bound scope. Ratios use the full selected
  request population and become unknown if any outcome is unknown.
- A `route-overview` aggregate-table descriptor and a recipe step referring to
  that same view/measurement, with no separate browser formulas.
- ClickHouse aggregate SQL and Databricks metric-view YAML/DDL/query artifacts
  through the existing compiler. These require prepared sources; no SQL is run
  or deployed by expansion.
- Explicit diagnostics for raw-exporter integration, unsupported duration and
  relation operators, unbound targets, and unconfigured objectives/alerts.

The expansion has content, profile and selected-inventory revision digests, an
inventory-basis label, the complete binding and selected resource values.
Equivalent ordering of failure ranges and duplicate recognized inventory
signatures do not change the expansion. Policy or generated-content changes do.
These hashes identify content; they are not signatures or proof of completeness.

No target, objective window, traffic rate, health threshold or notification rule
is inferred. The ratio describes the supplied observed requests, not guaranteed
full traffic. Input selection, sampling, coverage and temporal completeness
remain explicit assumptions.

## Normalized request preparation

`goo prepare-http` expands the same binding and prepares an already-selected
array of normalized **server** requests. Each request carries resource values,
opaque trace/span keys, a route template, method, and nullable status. This is
not an OTLP decoder or a query over the current exporter tables.

Preparation verifies the selected resource values, requires server kind and
nonempty scoped identity, classifies with the bound policy, and rejects duplicate
subjects. Missing route uses the reserved `(unknown route)` bucket; a literal
input using that reserved label is rejected to avoid merging real and unknown
routes. Missing methods similarly retain an unknown outcome. The caller/source
adapter must supply actual route templates, never derive a route dimension from
raw URLs or query strings.

The artifact carries the expansion revision, evaluation ID, input digest and
facts, with evidence labelled `unverified-input-selection-and-completeness`.
The input digest retains array order; it is an input-manifest identity, not an
order-independent evidence snapshot. Preparation rejects an expansion whose
content no longer matches its revision. The public Go structs and hashes are
not an authenticated publication/storage protocol.

The underlying source adapter must still establish the selection window,
sampling, OTel schema mapping, transport identity validity and evidence basis.
There is no custody-time or historical correction store in this HTTP slice.
The existing lifecycle kernel remains the path for an explicit business-event
contract; it is not implicitly applied to HTTP response status.

## Executable storyboard

```sh
go run ./cmd/goo expand \
  -binding examples/packages/http.yaml \
  -inventory examples/packages/inventory.json

go run ./cmd/goo prepare-http \
  -binding examples/packages/http.yaml \
  -inventory examples/packages/inventory.json \
  -requests examples/packages/requests.json -evaluation http-1

python scripts/check_packages.py
```

Use the configured Python virtual environment and Go toolchain as documented in
PROJECT_DIRECTION.md. The checker builds the actual CLI and executes the entire
journey, including extracting the shared definition/facts and running `evaluate`.

| Scene | Fixture evidence | Expected result |
| --- | --- | --- |
| Bind a compatible server contract | Versioned span metadata for checkout instance a | Explicit policy and prepared-table bindings expand |
| Inspect `/checkout` POST | 200, 503, missing response status | 3 requests; 1 success; 1 failure; 1 unknown; ratio unknown |
| Retain the missing-route bucket | GET with no route and status 404 under the explicit 5xx policy | 1 observed success in a distinct unknown-route bucket |
| Execute the shared definition | Actual Go CLI and generated SQL in embedded ClickHouse | Both match the independent expected groups |
| Review the longer-term target | Generated Databricks metric-view YAML | Structural validation only; execution/deployment not tested |
| Inspect unsupported capabilities | Duration/waterfall and raw-source requirements in the plan | Explicit diagnostics rather than invented implementations |

A [live HTTP source adapter](LIVE_HTTP_PACKAGE.md) now connects the aggregate
view descriptor to the reference UI and verifies it in a real browser journey.
The normalized fixture journey remains a separate deterministic check. Hegel generates response statuses and policy boundaries,
checking classification, unknown propagation and policy-dependent identity.

## Next layer

Connect a real OTel source adapter to this normalized input contract, including
fixed selection/basis and resource enforcement. Then expose the expansion and
capability diagnostics through the investigation UI, rendering the shared view
descriptor. Duration distributions and trace lookup require additional typed IR
operators and target support; histogram/counter packages should wait for their
unit, temporality and reset semantics rather than reuse count operators blindly.
