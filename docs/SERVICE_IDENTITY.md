# Service identity and declared-operation questions

Service discovery partitions OTel resources by `service.name` and
`service.namespace` inside the selected environment and fixed event-time window.
A missing namespace, a present empty namespace, and a literal namespace are
separate identities. Neither namespace nor service name proves deployment
membership, ownership, health, or complete topology.

The service catalog, reported parent edges, service overview, attribute questions,
operation inventory and native instrument discovery carry the selected namespace.
A cross-namespace parent link can connect two resources with the same service
name. Duplicate parent ownership across namespaces remains ambiguous. Selecting
one namespace can exclude a parent; this produces an unresolved diagnostic rather
than an inferred edge.

```mermaid
flowchart LR
  OTLP[OTLP resources] --> Catalog[Service name + namespace inventory]
  Catalog --> Scope[Environment + namespace + fixed window]
  Scope --> Records[Ordinary span and log questions]
  Scope --> Native[Exact native metric bindings]
  Scope --> Declared[Registered operation binding]
  Declared --> Question[Instance + scope/version + operation name/role question]
  Records --> Evidence[Explore, stage, capture and reopen]
  Native --> Evidence
  Question --> Evidence
```

The optional selector is `namespace: {mode: value, value: payments}`,
`namespace: {mode: value}` for present-empty, or `namespace: {mode: missing}`.
An omitted selector retains the earlier **all namespaces** semantics. It does
not mean missing. Existing questions, native definitions and operation packages
omit the new field when unselected, preserving their serialized content and SQL.
Record questions express namespace as an ordinary resource predicate; workbench
step scope retains its existing environment and window fields.

Operation packages may bind `resource.namespace` with the same selector. Their
trace, counter and histogram readers constrain and validate returned resources.
Dashboard planning uses server-registered declarations, never caller-supplied
contracts. Up to two complete bindings generate ordinary span questions matching
instance, instrumentation scope name/version and declared operation name/role
pairs. A conflicting declared namespace withholds the question; an unspecified
namespace is visibly narrowed by the selected service identity.

These cards count captured span records. They do not evaluate business outcomes,
conformance, sampling or completeness. The current trace exporter stores scope
name/version but **does not store scope schema URL**. Schema URL therefore stays
declared metadata and is explicitly unverified. No nonexistent physical column
is queried. Bindings above the six-operation automatic-card budget are withheld;
the underlying package remains available through its normal entry point.

Executable journey: [60 — namespace identity](../journeys/service-namespaces.spec.js).


## Declared services and observed services

Services reads the existing operation-package registry independently alongside
telemetry discovery. Registered declarations with no matching observed identity
remain visible as **Declared · Not observed in this window**. Their signal cells
and latest-event cell say not observed; no zero health value or synthetic event
time is created. Source discovery must succeed before this absence statement is
shown. Registry failure leaves observed services usable with a separate warning.

An explicit namespace binding joins only its exact missing/empty/literal identity.
An unspecified namespace annotates observed candidates with **namespace unspecified**;
if there are no candidates it produces an all-namespaces entry (or retains an
already selected namespace). This association does not verify instrumentation
emission or resolve namespace ownership. Multiple package revisions/bindings stay
visible; none is chosen as the authoritative runtime binding.

```mermaid
flowchart LR
  Registry[Current operation-package registry] --> Declared[Declared bindings and revisions]
  Source[Fixed-window discovery] --> Observed[Captured signal counts and reported parents]
  Declared --> Catalog[Service catalog with separate presence states]
  Observed --> Catalog
  Catalog --> Overview[Service overview and declared contract details]
  Overview --> Question[Explicit fixed-scope span or log question]
  Question --> Capture[Immutable record evidence, including no data]
```

The distributed view lists declared services separately; declarations create no
reported parent edges. Opening a declared service reads the same bounded metadata
as any other service. Contract widgets remain **not_observed** when spans are
absent. Explicit **Explore service spans/logs** actions can investigate the fixed
question even when no records were observed. Capturing and sharing that question
uses existing evidence semantics, including explicit empty results. Registry
metadata is current configuration and is not silently frozen into a record receipt.
No deployment roster, availability judgment or complete fleet is inferred.

[Journey 61](../journeys/declared-services.spec.js) exercises this through a real
historical inventory capture, independent registry read, service overview, empty
span capture, source-free reopening, registry failure and late-read cancellation.
