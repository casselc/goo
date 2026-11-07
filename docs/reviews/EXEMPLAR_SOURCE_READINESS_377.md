# Explicit-histogram exemplar readiness · 377

Status: pinned-source and actual ingestion/readback gate closed for the bounded
reported-reference subset. Product implementation is a separate authorized step.
The gate itself did not change product behavior. The subsequently authorized
source projection is recorded below; UI/continuation acceptance is tracked by
the milestone validation. The source-readiness probe remains offline by default.

## Gate

The next bounded capability is following an **actually reported trace/span
reference** from an explicitly selected histogram export. It is not proof that
the span contributed to a bucket, caused a latency change, or is fully retained.
A complete histogram does not make its exemplars a complete population.

Before product work, establish the following with a uniquely scoped ordinary
OTLP fixture and independent ClickHouse readback:

1. Valid nonzero trace/span IDs survive unchanged, including their pairing with
   the right exemplar ordinal, timestamp, stored value and filtered attributes.
2. All five nested arrays have equal cardinality, maintain input order and stay
   associated with the correct point/resource/scope identity. Empty arrays remain
   distinguishable from a malformed mismatched projection; do not zip to the
   shortest array or quietly attach another point's reference.
3. Exemplar timestamps retain all nine nanosecond digits. Source value and type
   losses are recorded explicitly, rather than upgrading Float64 storage into an
   exact original OTLP scalar.
4. The known referenced trace/span actually exists in independently queried
   trace rows. This fixture fact does not establish retention for every future
   reported reference. Missing reference IDs and an absent trace are separate
   unavailable outcomes.
5. Missing/all-zero/partial IDs never become invented references. Malformed ID
   requests are exercised separately and their real receiver/readback outcomes
   recorded; acceptance is not inferred from a JSON fixture.

A closed exact-reference subset can be useful despite scalar fidelity loss, if
it is called a stored reported reference and navigation does not depend on
pretending the original numeric type/value is known. Product acquisition must
remain explicit, scoped and owned by the current question; an exemplar timestamp
must not silently expand the checked time window.

## Pinned source

The lab pins Collector Contrib `v0.123.0`, peeled commit
`bef0a5cfb993608bbcce1eb5859c37295ac8eed2`; the tag resolution and running-image
check are documented in [374 source readiness](EXPONENTIAL_HISTOGRAM_SOURCE_374.md).
Read on 2026-11-07 through the configured TLS-verifying HTTPS proxy:

| Inspected source | SHA-256 |
| --- | --- |
| [internal/histogram_metrics.go](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/bef0a5cfb993608bbcce1eb5859c37295ac8eed2/exporter/clickhouseexporter/internal/histogram_metrics.go) | `092401d8fe1e45227f8a0e3b330268be8d192d8d2e0f6f1df8fb9605a9857620` |
| [internal/metrics_model.go](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/bef0a5cfb993608bbcce1eb5859c37295ac8eed2/exporter/clickhouseexporter/internal/metrics_model.go) | `cc3af3a7c1ca91db3a6a01cc5da8638e7acbc15b3196de6d9dac3de140584dd8` |

The histogram DDL declares `Exemplars Nested` with filtered attributes, nanosecond
time, Float64 value, span ID and trace ID. `convertExemplars` iterates the input
slice once and appends all five fields in the same loop. The histogram caller
assigns `(attrs, times, values, traceIDs, spanIDs)` and inserts span IDs into
`Exemplars.SpanId`, then trace IDs into `Exemplars.TraceId`; the similarly named
arrays are not swapped. Source code is not a live-schema or ingestion verdict.

| OTLP field | Pinned insertion/storage | Fidelity limit |
| --- | --- | --- |
| `filteredAttributes` | `AttributesToMap`, nested `Map(String,String)` | Original OTLP value types are stringified |
| `timeUnixNano` | `Timestamp().AsTime()`, `Array(DateTime64(9))` | Live nanosecond retention still needs readback |
| `asDouble` | Float64 | Reported binary64 may be retained; original value-type tag is absent |
| `asInt` | `float64(int64)`, same Float64 array | Integers beyond binary64 precision can round; original type is absent |
| no recognized value | `getValue` logs and returns `0.0` | Invalid OTLP exemplar can collapse to the same stored scalar as a real zero |
| `traceId`, `spanId` | Fixed-size pdata IDs, lower-case `hex.EncodeToString` | Missing/default IDs may become all-zero hex; must check live behavior |
| exemplar ordinal | Parallel append order within one point | No independent cross-point ID or bucket-membership assertion is added |

The cached OTLP Go protocol source is `otlp/v1.11.0`, origin commit
`bc625d6e040020737ab65c675c87e03bc841fd60`, inspected file SHA-256
`54159f4444309a1e82d9208553b770cdd56263603cecbfb1bdfa03b7268e69ed`.
[Exemplar definition](https://github.com/open-telemetry/opentelemetry-proto-go/blob/bc625d6e040020737ab65c675c87e03bc841fd60/otlp/metrics/v1/metrics.pb.go)
requires a recognized value oneof and explicitly permits missing trace/span IDs.
The exporter default for a missing value does not make that exemplar valid OTLP.

## Minimal real probe proposal

Emit two known trace spans under a unique service/environment/namespace, plus an
explicit histogram export containing ordered exemplars with distinct filtered
markers: an exactly representable fraction, an int64 above 2^53, integer/double
zero, no IDs and trace-only IDs. Include an empty-exemplar point and a distinct
qualified stream identity. Keep malformed trace/span ID lengths/hex in separate requests so a rejected
request cannot conceal the valid retention probe. The actual bounded probe used
two malformed-ID requests. Missing-value behavior above is a pinned-source fact,
not an additional live malformed-value test.

Capture actual image identity/version, live `DESCRIBE`/`SHOW CREATE`, original
OTLP bytes, input hashes, full raw point/trace rows, decimal timestamps, stored
Float64 bits, and array cardinalities/order. Query by the unique service and
bounded timestamps. Do not reset tables, alter collector configuration, create a
synthetic row in ClickHouse, or route the test through a goo evaluator.

## Authorized source projection after the gate

`measurements.HistogramExemplarReport` is an optional result field, version
`goo/histogram-exemplar-references/1`, with `valueEncoding: source-float64` and
`originalValueType: unknown`. Each export retains its original zero-based
`sourceRow`, full existing stream identity, export bounds/temporality, explicit
availability, ordered item essentials and total/omission counts. Items contain
`index`, `timeNS`, `traceId`, `spanId` and `storedValue`. Filtered attributes and
Float64 bits stay in readiness evidence, not this bounded product projection.

The source selects only the first **32** exemplar essentials and all original
array lengths. The existing **8 MiB** raw-source refusal remains authoritative;
metric rows are never silently discarded. The typed report has a **64 KiB**
ceiling, preferring later source rows and reporting `limited`/`omittedExports`.
Its absence means metadata was not retained. An available export with reported
zero and empty arrays is known empty; unavailable unknown totals are empty
strings, never an invented zero. Partial/malformed metadata makes only the
metadata unavailable, leaving independently valid numeric evaluation intact.

The exported selector replays the supplied result, scans **all raw rows** for
unique stream/end association including report-budget omissions, and returns
detached metadata. A missing export or item is unavailable, not empty. Reference
IDs and original scalar validity are not invented: an all-zero trace ID gives
no trace link, while a valid reported trace can stage context without a span ID.
A valid span is a focus hint, never a filter that removes peer spans. Canonical
sample time zero is retained as stored zero; an outside-window sample does not
widen the original question. Optional metadata failures never become a numeric
histogram failure or an automatic acquisition.

Focused source/selector tests pass in **0.159s**, including capping, empty versus
unknown, malformed optional fields, retained nonfinite stored values, detached
selection, replay tampering and an older duplicate omitted by the report budget.
This focused result is not a claim of completed browser/continuation acceptance.

## Completed live gate

`scripts/probe_histogram_exemplars_377.py` defaults to offline preparation and was
run once with `--run` against the preserved lab. Its original OTLP bytes, raw
schema/rows/image, request outcomes and separate offline retention verdict are in
`artifacts/histogram-exemplars377/`. No table, collector configuration, volume or
existing application data was reset. Three bounded persistence polls found all
three points and four trace rows.

The running collector reports `otelcol-contrib version 0.123.0`, image ID
`sha256:f0eaa24275f0eac77186cf0c9b1d1c72c9edde884a7c6bf9e5332ca8b5939d97`,
repository digest
`sha256:e39311df1f3d941923c00da79ac7ba6269124a870ee87e3c3ad24d60f8aee4d2`.
Live exemplar columns are `Array(Map(LowCardinality(String),String))`,
`Array(DateTime64(9))`, `Array(Float64)`, and two `Array(String)` ID columns.

All five nested arrays retained cardinalities **6 / 0 / 1** for the primary,
empty-exemplar and qualified-identity-decoy exports. Input order was preserved
even though exemplar timestamps deliberately were not sorted. Resource and scope
schema URLs, resource attributes, scope name/version/attributes/dropped count,
metric identity, point labels/bounds/counts and all nine timestamp digits matched
independent input/readback comparisons. The distinct identity export did not
receive another point's reference. Filtered attributes retained the exact Unicode
strings; integer attribute markers became strings as expected.

Known trace/span IDs matched per ordinal. The missing-ID entry became zero hex
for both IDs; the trace-only entry retained its trace ID and zero span ID. The
known trace contains two services in the exact namespace/environment. Two other
raw rows deliberately reuse the same trace/span IDs in a foreign namespace or
environment; applying those exact scope constraints leaves the two intended rows.
Both malformed-ID requests returned **400**, with no rows observed for their
separate service tokens.

The original int64 **9007199254740993** became stored Float64
**9007199254740992**. Integer zero and double zero both stored bits **0**. The
stored value-type discriminator is absent. These are confirmed losses; the
reference gate does not claim original scalar fidelity or valid original numeric
presence. The subset may expose a stored Float64 and an independently reported
trace reference while retaining those uncertainty limits.

Root independently checked all three exports/seven samples, the two scoped trace
rows, and rejecting association mutants before authorizing a bounded product
implementation. The gate does not authorize automatic acquisition, widened time,
metric-service-only trace filtering, exemplar-to-bucket attribution, full-trace
claims, or reopening exponential-histogram evaluation.
