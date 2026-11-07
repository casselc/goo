# Exponential histogram source readiness · 374

Status: pinned source and live ingestion checked. Thirteen uniquely scoped OTLP
points were appended by the probe; existing records, product, collector and schema
were unchanged. Independent verification confirms exact retained fields and the
zero-threshold/optional-presence losses described below. See the
[readiness decision](../EXPONENTIAL_HISTOGRAM_READINESS.md).

## Repository facts

- `compose.yaml` pins `otel/opentelemetry-collector-contrib:0.123.0`.
- `config/collector.yaml` already routes the OTLP HTTP/gRPC metrics pipeline
  through the memory limiter and batch processor to ClickHouse. Its
  `metrics_tables.exponential_histogram.name` is `otel_metrics_exp_histogram`.
- The exporter has `create_schema: true` and synchronous inserts configured.
  Configuration proves intended routing; it does not prove the running image,
  existing table layout, ingestion or losslessness.
- The neighboring observability repository contains
  `otel-chdb/metrics-layout/sql/contrib/otel_metrics_exponential_histogram.sql`.
  It has scale, zero count, signed positive/negative offsets and unsigned bucket
  counts, flags, min/max, temporality and resource/scope identity columns, but
  no zero-threshold column. Its timestamps are `DateTime`, unlike the nanosecond
  timestamps seen in current goo explicit-histogram journeys. This is a local
  research reference, **not evidence of the pinned or live exporter schema**.

## Pinned exporter source

Read on 2026-11-07 through the configured HTTPS proxy with TLS verification.
`git ls-remote` resolves the annotated `v0.123.0` tag object
`e5b6b90d3ba663dcc64adee2e3389d2ffba97581` to commit
`bef0a5cfb993608bbcce1eb5859c37295ac8eed2`.

| Source | SHA-256 of inspected bytes |
| --- | --- |
| [internal/exponential_histogram_metrics.go](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/bef0a5cfb993608bbcce1eb5859c37295ac8eed2/exporter/clickhouseexporter/internal/exponential_histogram_metrics.go) | `1743791dacc7876c3c893be77c938c95504ba4f12ff568643b3a0b17a102c55c` |
| [internal/metrics_model.go](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/bef0a5cfb993608bbcce1eb5859c37295ac8eed2/exporter/clickhouseexporter/internal/metrics_model.go) | `cc3af3a7c1ca91db3a6a01cc5da8638e7acbc15b3196de6d9dac3de140584dd8` |
| [exporter_metrics.go](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/bef0a5cfb993608bbcce1eb5859c37295ac8eed2/exporter/clickhouseexporter/exporter_metrics.go), fetched at resolved tag | `0484ea854efb2e6215059c1ce5ce58f352863a4d0a0a4e7edcc0843a628e497b` |

The first file contains both `createExpHistogramTableSQL` and
`insertExpHistogramTableSQL`, plus the `ExecContext` value mapping. It proves:

- **Zero threshold is lost:** neither the DDL, insert columns nor inserted
  values contains `ZeroThreshold`. Adding a column alone would not restore it.
- **Optional-statistic presence is lost:** sum/min/max are non-nullable `Float64`;
  insertion calls `dp.Sum()`, `dp.Min()` and `dp.Max()` without their presence
  indicators. Rows cannot distinguish absence from a reported getter-default
  value. The real probe must check the concrete absent/zero result.
- Nanosecond interval precision is represented by `DateTime64(9)` for start/end
  timestamps and `AsTime()` values; the older neighboring schema is not the
  pinned schema.
- Scale and offsets are `Int32`, counts and bucket counts are `UInt64`, flags
  are `UInt32`, and aggregation temporality is `Int32`; insertion uses the
  corresponding point fields and the metric's temporality directly.
- Resource attributes/schema URL, scope name/version/attributes/dropped-attribute
  count/schema URL, metric name/description/unit, point attributes and exemplars
  are included. `AttributesToMap` calls `pcommon.Value.AsString()` into
  `Map(..., String)`: original OTLP attribute value types are not retained.
  Resource dropped-attribute count is not an inserted column. This is a source
  retention limitation, not a new goo identity or schema proposal.

These statements describe source code at the pinned commit. `CREATE TABLE IF NOT
EXISTS` does not migrate an older table, and an image tag does not establish the
running image digest. The completed live probe below supplies that separate evidence.

## OTLP data model

The cached OTLP protobuf source (`go.opentelemetry.io/proto/otlp@v1.11.0`,
`metrics/v1/metrics.pb.go`, `ExponentialHistogramDataPoint`) establishes the
following protocol semantics. It is a protocol reference, not the Collector
0.123.0 implementation. Its module-cache origin is commit
`bc625d6e040020737ab65c675c87e03bc841fd60`, tag `otlp/v1.11.0`; the inspected
`metrics/v1/metrics.pb.go` SHA-256 is
`54159f4444309a1e82d9208553b770cdd56263603cecbfb1bdfa03b7268e69ed`.
[Immutable protocol source](https://github.com/open-telemetry/opentelemetry-proto-go/blob/bc625d6e040020737ab65c675c87e03bc841fd60/otlp/metrics/v1/metrics.pb.go).

| OTLP JSON | Meaning to preserve | Pinned ClickHouse mapping; live checks below |
| --- | --- | --- |
| `startTimeUnixNano`, `timeUnixNano` | Exact nanosecond interval bounds | `StartTimeUnix`, `TimeUnix`: `DateTime64(9)` |
| `count` | Unsigned total, equal to positive + negative + zero counts | `Count UInt64` |
| `scale` | Signed resolution; base = 2^(2^-scale) | `Scale Int32` |
| `positive.offset`, `positive.bucketCounts` | Signed first bucket index and contiguous unsigned counts | `PositiveOffset`, `PositiveBucketCounts` |
| `negative.offset`, `negative.bucketCounts` | Same index model for absolute magnitudes of negative observations | `NegativeOffset`, `NegativeBucketCounts` |
| `zeroCount` | Count in the zero region | `ZeroCount` |
| `zeroThreshold` | Zero region is closed `[-threshold, threshold]` | Not inserted: lost |
| `sum`, `min`, `max` | Optional values; absence differs from a reported zero | Non-nullable `Float64`; presence lost |
| `flags` | Includes the no-recorded-value bit; unknown data is not zero | `Flags` |
| `aggregationTemporality` | Delta/cumulative meaning belongs to the metric | `AggregationTemporality` |
| resource/scope schema URLs and attributes, scope name/version, metric name/unit, point attributes | Qualified stream identity | Separate resource/scope identity columns; attributes stringified |

The protocol does not restrict scale to a fixed small range. A bounded first
implementation must label other supported-protocol scales **unsupported**, not
malformed OTLP. JSON unsigned 64-bit values must remain decimal strings in
fixtures and source queries; converting counts through binary64 loses identity.

For positive index `i`, the nominal bucket is `(base^i, base^(i+1)]`.
Negative buckets use absolute magnitudes, so the corresponding observation
interval is `[-base^(i+1), -base^i)`. Negating boundaries also reverses closure.
A nonzero zero threshold can clip the zero-adjacent regions; it must not be
silently replaced by zero.

Positive scales generally produce irrational boundaries. Display approximations
must not become exact rational bounds, persisted identity or selection
predicates. Retaining `(scale, signed index, side)` gives an exact symbolic
bucket identity even when a formatted boundary is approximate. Exact downscaling
uses floor division of signed bucket indexes; truncation toward zero is wrong
for negative indexes.

## Required pinned/live evidence

1. Pinned Collector Contrib **v0.123.0** source inspection is complete above.
   Record the image identity actually used by the lab independently.
2. Capture the live `DESCRIBE TABLE`/`SHOW CREATE TABLE` without changing it.
   Check timestamps, signedness, bucket count types, zero threshold, optional
   sum/min/max and resource/scope schema/attribute columns.
3. In an isolated real OTLP fixture, retain input and independently queried rows.
   Include positive and negative offsets, a negative bucket, zero counts with a
   nonzero zero threshold, absent versus explicit-zero optional statistics,
   no-recorded-value flags, delta/cumulative temporality and distinct resource /
   scope identities. Counts, offsets, timestamps and identity should be compared
   directly, rather than through a goo evaluator.
4. Classify each field as retained, transformed or lost. A missing column or
   absent presence bit is a source limitation, not evidence of a default value.
   In particular, absent zero-threshold storage would prevent asserting exact
   zero-region geometry from these rows alone.

## Completed live verification

The [reproducible probe](../../scripts/probe_exponential_histogram_374.py) sent 13
valid points to the existing collector. All 13 appeared in the live table after
two bounded polls. The actual image/version, DESCRIBE/SHOW CREATE and raw rows are
retained separately from the repository configuration hash. That hash names the
declared bind-mounted file, not an independent applied-process configuration read.

The independent verifier checked exact integer arrays/counts (including a positive
count of 9007199254740993), signed offsets, scales, flags, temporality, reported
identity and all nine timestamp digits. Two actual projection collisions confirm
zero-threshold loss and omitted-versus-zero Sum/Min/Max presence loss. Seven
changed-report controls were rejected. Original observations remain unchanged.
General execution remains deferred under the source and representation contracts;
no exporter upgrade, new manifest or explicit-bin conversion was implemented.
