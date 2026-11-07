# Exponential histogram representation readiness — 374

Status: source/protocol research only. **Defer general exponential-histogram
execution through the current explicit-duration representation.** A separate
source/table probe may establish transport, but cannot close the representation
gap below. No production kernel, schema or view was changed.

## Reproducible official protocol pin

The project's weaving/sample modules already pin
`go.opentelemetry.io/proto/otlp v1.11.0`. Its cached Go module origin records tag
`otlp/v1.11.0`, repository `open-telemetry/opentelemetry-proto-go`, commit
`bc625d6e040020737ab65c675c87e03bc841fd60`, module checksum
`h1:5rrYs0Ykyj50sdU/JU0x8etU+LubXWb+gED6TbEdMIk=`. This research read the existing
verified module and origin metadata; no network or executable probe was needed.

The official generated protocol binding preserves its source comments:
[ExponentialHistogram and data point](https://github.com/open-telemetry/opentelemetry-proto-go/blob/bc625d6e040020737ab65c675c87e03bc841fd60/otlp/metrics/v1/metrics.pb.go),
locally `metrics/v1/metrics.pb.go` lines 881–893, 1334–1419 and 1810–1835.
This pins the binding actually available to the project; it does not invent an
upstream specification commit or claim this is the newest published version.

## Exact point semantics to preserve

Let scale be signed integer `s`, base `b = 2^(2^-s)`, and a bucket array entry `j`
have index `i = offset + j`. Positive magnitudes occupy `(b^i,b^(i+1)]`.
Negative values map by absolute value to the same magnitude buckets, so their
original-domain interval is `[-b^(i+1),-b^i)`. That reversal of endpoint closure
matters; negative buckets are not positive intervals with a minus label.

`zeroCount` represents the closed region `[-zeroThreshold,+zeroThreshold]`.
At zero threshold zero, the protocol also includes values not expressible under
the exponential formula and rounded-to-zero values. A receiver cannot reconstruct
those original measurements. A nonzero threshold cannot silently be replaced by
one nominal exponential bucket or distributed proportionally among them.

`count` equals positive bucket counts + negative bucket counts + zero count, all
uint64. Preserve counts as exact integers/decimal strings and reject overflowing
sums. Arrays encode consecutive indexes, including explicit zero-count buckets.
Offset arithmetic needs wider checked integers, and a local array/scale/work
budget must be explicit. The protocol's int32 scale is **not restricted**; SDK
preferred ranges are not protocol validity limits.

Start/end nanosecond timestamps, resource/scope/attribute identity, flags and
aggregation temporality remain part of the observation. Delta and cumulative are
distinct, unspecified temporality cannot be guessed, and no-recorded-value is not
zero. Sum/min/max are optional statistics, not substitutes for missing buckets.
Units remain exact instrument units; no seconds/milliseconds conversion or fleet
aggregation is implied by supporting a representation.

## Why the current representation is insufficient

`metrics/histogram.go` accepts at most 256 finite, nonnegative, increasing binary64
explicit bounds and uses a single left-open/right-closed closure. Its distribution
and p95 range expose those same float64 boundaries. `measurements.HistogramView`
and `ui/native-histogram-view.mjs` project them into explicit-bin scoped datasets;
JS records exact rationals for the **binary64 values supplied**, not arbitrary
mathematical exponential boundaries. `scopedview.Histogram` currently declares
explicit binning and one closure for the whole histogram.

At scale 1, boundary index 1 is `sqrt(2)`, which has no finite binary64 or rational
representation. Converting it with math.Pow/Math.pow would change the boundary,
not preserve the original exact interval. Negative closure and the separate
closed zero region cannot be encoded by the current single-closure nonnegative
shape either. Large index/scale combinations also underflow/overflow binary64.
A rounded chart label must not become an exact selector or p95 endpoint.

A scale-zero, nonnegative, finite-power-of-two subset could avoid some numeric
issues, but would still require a genuine source contract and explicit zero/tail
semantics. Accepting that exceptional subset must not advertise general OTLP
exponential support, and no evidence currently selects it as a useful product cut.

## Temporal/layout evaluation is a separate prerequisite

The existing explicit kernel rejects overlapping delta intervals and represents
cumulative epoch/layout changes, count/bucket decreases and excessive gaps as
unknown/tainted intervals. Those policies cannot be copied blindly:

- Changing compact offset/array extent may only add zero buckets and is not alone
  a reset. Align indexes before subtraction; preserve exact counts.
- Different scales can sometimes be downscaled exactly by merging buckets.
  The index mapping uses floor division (not truncation toward zero for negative
  indexes). Bound work and reject overflow; do not upsample invented populations.
- Changed zero thresholds can require splitting previously aggregated mass.
  Without evidence allowing exact redistribution, withhold the interval.
- Epoch, temporality, duplicate/conflicting point, gap and no-recorded-value rules
  must remain explicit and independent of chart layout. A layout change is not
  permission to infer zero traffic or a fresh cumulative baseline.

An independent fixture should exercise positive/negative indexes and values,
zero regions, mixed offsets/scales, delta/cumulative epochs, count overflow and
unknown/reset states. Exact symbolic or conservatively enclosing boundary
semantics must be decided before a renderer or query selector can consume them.

## Bounded decision and prerequisites to reopen

`measurements/source.go` currently selects `otel_metrics_histogram`; discovery
only enumerates existing gauge/sum/explicit histogram tables. The separate 374
collector/exporter probe must pin actual exponential fields and retention without
assuming transport conversion proves kernel support.

Reopen implementation only with (1) that verified source contract, (2) a bounded
exact-bound/closure representation or explicitly conservative uncertainty model,
(3) interval/layout evaluation and independent arithmetic oracle, and (4) a scoped
view contract that distinguishes exponential from explicit bins. Existing owners,
immutable evidence, typed question scope and acquisition APIs can then be reused.
Do not relax ordinary record/native explicit-bin guards, silently convert to
rounded explicit buckets, add a generic evaluator or call a transport-only probe
an implemented observability capability. A precise defer is the valid result of
this readiness gate under current representation constraints.
