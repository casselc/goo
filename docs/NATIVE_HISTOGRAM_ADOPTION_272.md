# Native histogram adoption — original proposal272

The bounded implementation is now documented in
[Native interval histograms](SCOPED_NATIVE_HISTOGRAMS.md), with actual checks in
[VALIDATION.md](../VALIDATION.md). The original proposal below is retained as
design context, not a second implementation plan. Milestone271 adopts raw-span duration
partitions. A native OTel histogram counts instrument observations and requires
measurement scope, interval and bucket semantics before shared display adoption.

Start with one evaluated `HistogramInterval`, or one complete `BucketHistograms`
bucket, for one exact instrument stream. Do not begin with fleet aggregation,
a service-wide histogram or a heatmap assembled from incompatible layouts.

## Prerequisite: measurement scope

The current scoped dataset accepts record questions (`exploration.Query`, logs or
spans). Add an explicit checked measurement-scope variant before attaching native
histograms. It must identify the full instrument stream, exact nanosecond interval
and evaluated result basis. A native metric must not masquerade as a span question.
Imported data remains held and cannot mint a runtime completeness witness.

## Proposed descriptor and geometry

Extend the tagged histogram union with `kind: 'otel-explicit'`, observation grain,
original `delta`/`cumulative` temporality, instrument unit and interval total.
Rendered counts are evaluated interval increments in both cases. Explicit buckets
are left-open/right-closed; the first lower endpoint is `-inf`, the final upper
endpoint `+inf`. Preserve finite binary64 source bounds exactly as canonical
rationals, with human formatting separate. Do not reuse the duration half-open
contract or numeric parsing support.

Consume existing evaluator output rather than fetching or re-evaluating in a
component. A private bins witness may attest the complete evaluated exported
interval/layout. It says nothing about upstream ingestion, sampling or coverage of
all fleet members. Missing/reset/layout-change/gap/boundary-crossing/no-recorded-value
results retain diagnostics and withhold known count geometry rather than inventing
zeros. Empty known observation populations remain distinct from unknown intervals.

Quantiles remain ranked bucket ranges, including infinite tails. Do not interpolate
a precise p95, average per-stream percentiles or imply density with equal-width
bars. Selection can inspect the instrument bucket and measurement context; it
cannot claim exact contributing spans or generate a `duration_ns` predicate.

## Existing foundations

- `metrics/histogram.go`: interval DTOs, ranked p95 bucket, `EvaluateHistograms`
  and explicit `MergeAlignedHistograms` compatibility checks.
- `metrics/histogram_time.go`: `BucketHistograms` requires whole increments.
- `metrics/window_comparison.go`: contiguous full-window and layout validation.
- `ui/measurements.mjs`: native timeline and histogram heatmap adapters.
- `ui/histograms.mjs`: existing bucket and percentile-range labels.

Keep272 per-stream. Later merging must reuse explicit compatibility: identical
stream dimensions except instance identity, duplicate-free members, identical
unit/layout/window, known operands and checked overflow. Automatic fleet discovery
and arbitrary time rebucketing are separate work.

## Acceptance

1. Go and JS agree on checked measurement scope and exact native bucket geometry.
2. Property cases cover delta/cumulative increments, reset/layout/gap refusals,
   empty versus unknown intervals, extreme finite bounds and infinite tails.
3. The shared passive histogram/table displays preserve units, clocks and exact
   counts without source I/O or raw-span action authority.
4. A real native-instrument journey demonstrates one interval and read-only saved
   inspection, with model/negative controls, updated diagrams and screenshots.

Renderer and contributed-view editor expansion remains behind this adoption gate.
