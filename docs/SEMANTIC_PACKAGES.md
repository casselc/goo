# OTel semantic packages for goo

Goo should turn recognizable telemetry into reusable measurement and investigation
building blocks. The same package should work in an interactive investigation,
application-owned configuration, a scheduled evaluation, and a compiler target.
The sample application is one adapter and an integration test for this purpose.

An OTel convention supplies a vocabulary and measurement contract. It does not
supply the application's business objective, completeness guarantees, retention
policy, or an automatically correct causal diagnosis.

## Useful package families

| Family | Reusable defaults | Decisions or evidence still needed |
| --- | --- | --- |
| HTTP server/client | Route/method RED overview, duration distributions, response/error breakdown, slow-request heatmap → waterfall → dependencies | Server and client error policies differ; templated routes and sampling/completeness must be known |
| RPC | Service/method distributions, protocol status breakdown, client/server investigation | Protocol-specific error policy, legacy versus newer names/units, cancellation/deadline meaning |
| Messaging | Destination/group processing views, producer/consumer links, retry investigations | Broker-specific lag/backlog/redelivery adapters; queue wait differs from processing; business identity and terminal outcome are application-defined |
| Database client | Operation/query-summary latency, error breakdown, dependency drilldown | Optional sanitized query summaries; no raw SQL text discovery by default; client duration does not establish server contention |
| Database pool | Connections by state, pending requests, acquisition wait, timeout investigation | Same named pool/instance, actual capacity, snapshot/temporality semantics; open connections do not equal busy connections |
| Language runtime | Go memory/goroutines/GC or JVM heap/GC/thread pressure guides | Runtime version, instance identity, valid memory limit; a growing count is not itself a leak diagnosis |
| Host/container/process | CPU, memory, disk/network USE guides and resource-to-service investigations | Correct resource identity and capacities; host pressure is not proof of which service caused it |
| Cross-cutting | Service/environment/release navigation, semantic attribute discovery, error evidence, exact trace/log correlation and async links | Namespace/tenant scope, multiplicity, cardinality budgets, missing attributes, instrumentation coverage |

These can yield **measures, dimensions, relations, visualizations, investigation
recipes, optional objective templates, and backend mappings**. They should share
one normalized model rather than become unrelated dashboards with their own SQL.
A valid configuration change should affect interactive and managed resources
through the same resolved package revision.

For example, the HTTP package can propose an availability-objective wizard and
its route/error dimensions. The application owner must still select eligibility,
error policy, target, window and evidence assumptions. Goo should not silently
install a 99.9% objective or page someone because it recognized HTTP telemetry.

## Implemented foundation

`prebuilt/` is a Go capability-recognition pilot with ten versioned profiles:
HTTP metrics and spans, legacy RPC metrics, messaging metrics and spans, database
operations and pools, Go runtime, host CPU, and a custom worker/queue adapter.
Profiles are declared in `prebuilt/catalog.yaml`; `goo catalog` exports them.
The UI's **OTel prebuilt library** displays the same generated catalog, including
requirements, proposed outputs and caveats.

```sh
go run ./cmd/goo catalog
go run ./cmd/goo discover -inventory examples/prebuilts/inventory.json
```

The example inventory is explicitly illustrative, not a live scan. Inventory is
versioned JSON with an evidence-basis label, explicit resource scopes, and signal
signatures (kind, name, unit, schema URL and co-observed attribute keys).
`discover` produces a per-profile, per-scope proposal with every requirement's
expected contracts, matched status, missing resource fields and output maturity.

The two statuses are intentionally limited:

- `compatible-shape`: supplied metadata meets this profile's shape requirements.
- `needs-evidence-or-binding`: at least one required signal, dimension, resource
  identity, unit or version is not matched.

The recognition registry does not extract live inventory or install resources.
A first [HTTP binding/expansion](PACKAGE_EXPANSION.md) now adds normalized-input
validation and aggregate compilation through the shared model. Broad profile
compilation and a complete package resolver remain unimplemented.
Most catalog outputs are marked **proposed**. The queue view is implemented in
the reference lab. An inventory declaration is not authenticated evidence and
its basis string is not an immutable lifecycle receipt.

The registry is stricter than a minimal OTel producer where a particular default
view needs optional dimensions. An absent route/status dimension disables that
profile, not the instrument's validity. Future profiles can offer a less detailed
view with explicit unknown buckets. A `span/*` profile may leave the span name
unspecified because library span names vary; kind, version, unit and required
attributes must still match.

## Binding and safety rules

1. Recognize explicit semantic versions and instrument kinds. Names alone are
   insufficient: the pinned RPC duration convention uses milliseconds while
   HTTP/database durations use seconds. Do not guess conversions or silently
   alias renamed conventions. Unversioned sources need an explicit binding.
2. Match attribute keys within one supplied signal signature. Never union keys
   from unrelated series to manufacture an apparently complete instrument.
3. Match requirements within one declared resource scope. Do not borrow capacity
   from a different service instance. Later value-level binding must prove the
   same pool/destination/group as well; matching attribute *names* is insufficient.
4. Before executing a proposed measure, validate value types, dimensions,
   temporality/reset behavior, bucket boundaries and coverage assumptions.
   Gauge snapshots, cumulative counters and histograms need distinct operators.
5. Use low-cardinality defaults: route templates, service/method, database
   operation and known destinations. Raw URL, SQL text and message IDs should
   not become default metric dimensions. Sensitive attributes need explicit
   handling and presentation policy.
6. Keep exact parent edges/trace links distinct from semantic business identity
   and contextual searches. Matching a destination or timestamp is not an exact
   producer-consumer relation or evidence of causation.
7. Treat no-data, stale, invalid and unknown as separate results. Recognition
   never establishes healthy status, complete traffic, or valid absence evidence.

These rules apply equally to library instrumentation, framework auto-instrumentation
and manually emitted telemetry. A library name/version can suggest a binding,
but does not override the observed contract.

## Package expansion and portability — proposed next layer

A complete package should contain pinned imports, source selectors, normalized
measures, dimension definitions, relation contracts, capability requirements,
view/recipe descriptors, evidence policy defaults and target lowering support.
Application configuration would bind resources and override policy; the resolved
package would carry a content digest and source/schema revisions.

A first [typed bind/expand slice](PACKAGE_EXPANSION.md) is now implemented for HTTP server spans. It emits the shared aggregate model, backend plans and view/recipe descriptors, with normalized-input preparation. Broader package expansion remains proposed:

A complete **bind/expand** stage should: take a discovery
proposal, explicit source and dimension-value bindings, and policy choices; emit
reviewable measurement/view/recipe artifacts with capability diagnostics. That
stage should resolve into goo's shared semantic IR. It should not run hidden
browser computations or emit arbitrary SQL supplied by a template.

ClickHouse and Databricks remain physical targets. Counts/ratios over prepared
facts can reuse the current Go aggregate compiler. Trace distributions, histogram
merging, counter rates, resource gauges and relation lookup need additional typed
operators and explicit target support. Databricks metric-view export is therefore
not claimed for the new catalog. Dynamic investigation selections may remain
query targets rather than metric-view definitions.

The Go lifecycle kernel supplies subject/outcome and evidence-basis semantics
when a package has an explicit business event contract. OTel messaging conventions
alone cannot supply that contract. A reservation outcome is one example of this
broader, application-bound layer.

## Pinned primary references

Profiles were researched against OpenTelemetry semantic-conventions **v1.37.0**,
commit `aec6e9d3e86754683dab7c707655d69d953b2768`, not assumed to describe every current instrumentation library. Consult each
family's stability and migration rules before adding another version:

- [HTTP metric model](https://github.com/open-telemetry/semantic-conventions/blob/aec6e9d3e86754683dab7c707655d69d953b2768/model/http/metrics.yaml): request duration in seconds is stable; several additional instruments are development.
- [RPC metric model](https://github.com/open-telemetry/semantic-conventions/blob/aec6e9d3e86754683dab7c707655d69d953b2768/model/rpc/metrics.yaml): legacy duration in milliseconds, development in this version.
- [Messaging metric model](https://github.com/open-telemetry/semantic-conventions/blob/aec6e9d3e86754683dab7c707655d69d953b2768/model/messaging/metrics.yaml): processing versus client-operation duration, destination/group attributes; development.
- [Database metric model](https://github.com/open-telemetry/semantic-conventions/blob/aec6e9d3e86754683dab7c707655d69d953b2768/model/database/metrics.yaml): stable operation duration; pool metrics remain development.
- [Go runtime metric model](https://github.com/open-telemetry/semantic-conventions/blob/aec6e9d3e86754683dab7c707655d69d953b2768/model/go/metrics.yaml): runtime memory and goroutine instruments, development.
- [System metric model](https://github.com/open-telemetry/semantic-conventions/blob/aec6e9d3e86754683dab7c707655d69d953b2768/model/system/metrics.yaml): resource instruments and CPU state dimensions, development.

The custom queue metrics use `goo://demo/queue/v1`; that is goo's own contract,
not an OTel standard. Standard messaging identity attributes are reused. See
[the queue adapter](QUEUE_TELEMETRY.md) and [browser journeys](journeys/README.md).
