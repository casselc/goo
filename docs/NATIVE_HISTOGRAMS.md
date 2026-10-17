# Native duration histograms

Goo evaluates native OTel duration histograms separately from span durations.
The checkout YAML package enables `histogram-heatmap`; the report package does
not. The UI shows per-stream bucket populations over actual exported intervals.
See [architecture diagram 5](ARCHITECTURE_DIAGRAMS.md#5-native-histogram-intervals-and-bounded-investigation)
and [journey 14](journeys/README.md#14-native-duration-histograms).

## Evaluation contract

`metrics/histogram.go` supports cumulative and delta points. Stream identity
includes the complete resource, instrumentation scope/version/schema, metric,
unit and attributes. Bounds are finite, nonnegative and strictly increasing;
there must be one more bucket than explicit bounds. Counts are exact canonical
uint64 decimal strings and must sum to Count without overflow. The ClickHouse
adapter reads UInt64 counts as strings, including values above 2^53.

Cumulative points establish a baseline before subtracting buckets. A changed
start epoch, missing-value flag, bucket/count decrease, layout change, excessive
gap or clipped window boundary produces unknown rather than fabricated data.
A same-epoch decrease/layout change taints subsequent deltas until a new epoch.
Late points can repair a gap and advance the evidence basis without mutating
an earlier result. Conflicting duplicates and overlapping delta intervals reject.
No extrapolation, proportional time allocation or subtraction of min/max occurs.

P95 is a **bucket range**, computed from exact rank counts, never an interpolated
scalar. Explicit buckets have exclusive lower and inclusive upper bounds; the
first/last tails remain unbounded. A zero-population interval has no percentile.
The selected cell shows its population and the P95 range for the entire interval.
No scalar percentiles are averaged.

## Package and live source

`config/operations/checkout.yaml` declares metric `goo.operation.duration`, unit
`s`, cumulative temporality, explicit bounds, allowed outcomes and `maxGap: 10s`.
The adapter checks these against exported data and rejects incompatible layouts;
it does not silently rebin. Resource, scope/version/schema, operation attributes,
flags and timestamp validity are checked before evaluation. Queries are typed,
readonly and bounded to 1,000 source points per contract and 8 MiB. An exceeded
limit is an error, never a partial successful distribution.

`GET /api/histograms?package=checkout&fromMs=…&toMs=…` returns separate metric
bases tied to the package revision. The live checkout app exports cumulative
histograms every two seconds through standard OTLP. Delta semantics are covered
by fixtures rather than a second live SDK configuration.

## Explicit merging

`MergeAlignedHistograms` and `goo merge-histograms -input members.json` accept
explicit stream/interval members. All identities must match except
`service.instance.id`; units, bounds and interval endpoints must agree exactly.
Unknown members, incompatible layouts or unaligned intervals yield an unknown
merge. Duplicate members and uint64 overflow reject. Addition is order-independent.
The merged P95 is recomputed from merged buckets.

This API does not discover fleet membership, certify missing-instance coverage,
rebin durations or align mismatched collection times. The live UI stays per stream.
`goo evaluate-histograms -input request.json` runs the same evaluator offline.
Runnable fixtures:

```sh
go run ./cmd/goo evaluate-histograms -input examples/metrics/histograms.json
go run ./cmd/goo merge-histograms -input examples/metrics/histogram-members.json
```

The first returns four observations; the merge returns eight. JSON shapes are
defined by HistogramRequest and HistogramMember in the Go kernel.

## Investigation and verification

The heatmap labels unknown cells `?`, observed empty buckets `0`, and leaves
uncovered edges unfilled. Bucket selection explains exact population and bounds.
“Inspect histogram context” selects captured spans by operation/outcome and time;
it does **not** assert bucket membership, exact contributors or exemplar identity.
Metric and span bases remain separate.

Root Go race tests exercise exact counts, invalid layouts, resets, flags, gaps,
late repair, boundaries and merge compatibility. Hegel runs 100 state/swarm
sequences and 150 generated merge cases. Quint validates observed evaluator
restart/gap/repair/merge steps and rejects invented reset and incompatible-merge
populations. These bounded checks are not a proof over all possible histories.
Journey 14 uses actual app calls, collected histograms and real browser clicks;
its generated screenshots are linked in the storyboard.
