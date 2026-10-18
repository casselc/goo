# Typed record measures and split time charts

Implemented in the Go record adapter. Existing Explore/YAML controls now expand
into a bounded typed measure plan used for aggregate SQL lowering. This is the
first record-measure slice, not a general expression language or a replacement for
native metric/lifecycle kernels.

## Typed operators

`Query.RecordMeasures()` returns the shared semantic scope, dimensions, fixed
bucket settings and typed operators. Source-free package plans expose `measurePlan`;
recipe/follow-up transforms expose their target record plan. The UI's source/OTel
disclosure lists the operators, units and merge requirements.

| Existing question choice | Typed operators | Exact representation / merge rule |
| --- | --- | --- |
| All record questions | Count | uint64 captured records; sum compatible partitions |
| Span duration summary | Sum, min, max | uint128 nanosecond sum; uint64 extrema |
| Span duration summary | Mean | Rational sum/count, not an average of group averages |
| Span duration summary | p50/p95/p99 | uint64 nanoseconds; exact-high order statistic `floor(p*N)` clamped to `N-1`; requires raw values for recomputation |
| Recorded span status | Partitioned count | Aligned Error/Ok/Unset/unknown counts over the same matching span population |

This plan is an adapter over the existing public `summary`, `spanStatus`,
`groupBy` and `bucketMs` fields. Arbitrary caller-authored operator ASTs, casts,
calculated attributes, ratios and arbitrary quantiles are not accepted yet.
Span-duration units come from the known exporter duration field, not an instrument
name. Native OTel histograms/counters retain their separate unit, temporality,
reset and stream-identity contracts. Recorded status is not an HTTP/RPC or business
outcome policy; Unset and unknown are never silently turned into success.

The compiler lowers planner-produced operators into the existing aggregate result
shape. This preserves count/duration/status reads and inspection independently.
Databricks lowering and a public typed authoring schema remain subsequent work.

## Split chart by breakdown

In Explore, choose a time interval and a breakdown, then enable **Split chart by
breakdown** and Apply. YAML/package questions use:

```yaml
summary: span-duration
spanStatus: true
bucketMs: 10000
groupTimeSeries: true
groupBy:
  - kind: attribute
    key: http.route
```

The chart supports one or two explicit dimensions, at most eight observed series,
1,000 observed group/interval cells, and the existing maximum of 240 intervals.
Budgets fail closed: no hidden top-N or partial group population is returned.
Reduce the scope/cardinality or increase the interval when necessary.

`groupedTimeSeries` contains each series' ordered `values` (strings or null) and
dense `points` using the existing time-point shape. A null dimension means missing;
an empty string is a present value. Only groups observed in that read get a series.
Missing intervals inside an observed series become `no_data` with zero captured
count and unavailable duration measures. This does not prove health, absence of
work or instrumentation coverage. Intervals are half-open, anchored to the original
question start; the final interval is clipped.

Small multiples share a captured scale. Switch locally among counts, mean/p95/max
span duration and recorded-error percentage when available. Search hides displayed
series without changing the scale or totals. Selecting a cell exposes its exact
bounds/count/value. These controls do not issue source reads. Counts are records
per interval, not inferred request rates; span latency is not complete-trace/root
latency. Error percentage uses all matching captured spans as denominator.

Each cell's percentile is computed from its raw durations. No scalar percentiles
are averaged to produce another cell or total. The split chart is a separate
aggregate read from the total chart, total measures, group summaries and row page.
The adapter compares additive count, duration-sum and recorded-status states across
these reads and preserves `changed_between_reads` on disagreement. Agreement does
not establish a transaction or matching raw records. Invalid shapes/bounds and
duplicate cells fail rather than becoming ordinary disagreement.

## Capture, reuse and compatibility

Captures freeze all supported series independently of the bounded inspection row
bank. Reopen/paging and local chart controls read stored evidence. Creating a
package retains split-chart settings; fresh reuse resolves new bounds and reads
new evidence while the old capture remains unchanged. Local chart-measure/search
selection resets on reopen; captured measures remain available without rereading.
Refresh retains fixed bounds
and existing source-plan checks.

The flag and new result/plan fields are omitted when disabled. Old questions,
package revisions, YAML and SQL plans keep their previous representation; existing
bucket+breakdown questions still produce the original unsplit chart by default.
Old captures remain readable. New captures include the requested bounded series
and enforce structural/accounting validation without rewriting historical evidence.

[Storyboard 43](journeys/README.md#43--compare-http-route-measures-over-time)
uses ordinary OTLP HTTP-server-shaped telemetry through the real collector and
ClickHouse, including missing/empty route dimensions and more spans than its row
page. This is controlled integration telemetry, not a production traffic sample or
a semantic-convention completeness certificate. It requires no goo manifest.
