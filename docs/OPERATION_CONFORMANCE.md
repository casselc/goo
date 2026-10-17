# Observed operation conformance

Implemented: source-manifest conformance and a completed-invocation overview in
Go. This page covers the developer evidence journey. A separate
[live operation journey](LIVE_OPERATIONS.md) now connects a woven sample to
ClickHouse and the investigation UI; automatic live inventory scanning remains absent.

1. A developer adds `//goo:operation` declarations to the checkout package.
2. `goo weave` validates declarations and runs the actual woven example with an
   OTel SDK exporter. Success, error and panic paths preserve application behavior.
3. The test normalizes exported spans, retaining scope name/version/schema,
   explicitly selected resource identity, trace/span IDs, operation/role/outcome
   and exact duration nanoseconds. It checks the overview against actual calls.
4. The developer inspects the same report with the ordinary Go CLI:

```sh
QUINTGO_BACKEND=typescript \
GOO_ANNOTATION_ARTIFACTS="$PWD/artifacts/annotations" \
go run ./cmd/goo weave -module verification/weaving -dir example \
  -test ./scenario -manifest artifacts/annotation-manifest.json

go run ./cmd/goo observe-operations -dir verification/weaving/example \
  -spans artifacts/annotations/operations.json \
  -resource woven-checkout -basis actual-sdk-export
```

Expected: checkout.reserve has two completed invocations, one error;
checkout.panic has one completed invocation and one panic. Counts describe
function invocations, not unique requests, retries or business subjects.
Durations are sorted exact uint64 nanosecond samples, not fleet percentile
estimates. Consumers must preserve integer precision when reading JSON.

One resource is selected explicitly. Every supplied span must match that resource
and the exact instrumentation scope/version/schema contract. Undeclared operation
names, changed roles, invalid identities, unknown outcomes and conflicting copies
fail the report. Identical trace/span copies count once. Declared operations with
no evidence remain `unobserved`, never zero-error healthy. Output order and duration
sample order are deterministic. Input is the full supplied evidence array; there
is no presentation limit before totals.

The source revision is recomputed from the selected source directory. The basis
is a caller-supplied provenance label, not a frozen custody snapshot. Resource
identity is supplied by the source adapter; this first SDK journey uses an explicit
test resource label. This is a normalized operation-span API, not an OTLP decoder.
It does not accept metric points or infer invocation totals from cumulative
counters. Missing scope schemas are rejected rather than silently assumed.

Sampling, omitted spans, instrumentation coverage and completeness remain unknown.
No active-invocation count, rate, error ratio, histogram merge, SLO, durable business
outcome or backend aggregate lowering is claimed. Counter resets/temporality and
histogram compatibility need explicit contracts before native metric operators.
The separate live integration now supplies a woven sample, collector/ClickHouse
adapter and operation discovery/overview panel; see LIVE_OPERATIONS.md.
