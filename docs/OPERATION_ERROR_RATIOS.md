# Operation error ratios

The checkout package now shows calls, errors and an error percentage over the
same supported counter intervals. Expand **Operation error ratios**, select
**Why did errors rise?**, choose an error outcome and open a contextual trace.
This is an investigation aid, not an automatic increase detector or causal claim.
See [journey 15](journeys/README.md#15-from-operation-error-ratios-to-traces) for
real browser evidence and [diagram 6](ARCHITECTURE_DIAGRAMS.md#6-from-aligned-counters-to-an-error-investigation)
for the data path.

## Declared policy

```yaml
rates:
  - name: calls
    metric: goo.operation.calls
    unit: '{invocation}'
    temporality: cumulative
    maxGap: 10s
    outcomes: [ok, error, panic]
ratios:
  - name: operation-errors
    totalRate: calls
    errorOutcomes: [error, panic]
views: [operation-overview, error-ratios, trace-waterfall]
```

A ratio references one declared total-rate contract containing the exhaustive
operation outcome domain. Its error outcomes are a distinct, nonempty subset of
`error` and `panic`. Thus policy can explicitly include or exclude panics. `ok`
cannot be an error outcome. Unknown rate references, partial denominator domains,
duplicate names/outcomes and invalid view bindings reject during planning.
The package revision includes this policy. Reports remains trace-only.

`GET /api/counter-rates?package=checkout&fromMs=…&toMs=…` returns `ratios` alongside
existing `rates`. Each ratio derives solely from its referenced rate's evaluated
result. It does not divide independently fetched calls/errors contracts.
`counterBasis` identifies that source result; the ratio basis additionally covers
its policy, canonical full input and output. Metric and span bases remain separate.

## Exact counts and alignment

`metrics.EvaluateErrorRatios` groups only streams identical in complete resource,
scope/version/schema, metric, unit and attributes except outcome. Instances and
operations never merge implicitly. Unit must be `{invocation}` and temporalities
must agree. Every required outcome must have an observed valid interval with
exactly the same endpoints. Endpoint partitions prevent overlapping output;
counts are never prorated to fit a partial interval.

An absent outcome, unaligned interval, reset/decrease, missing flag, excessive gap
or clipped edge stays **unknown**, with no counts or percentage attached. Even
when a missing outcome cannot affect the numerator, it prevents an unsupported
denominator. The underlying counter kernel handles epochs and late repair before
ratio evaluation. A new result/basis can incorporate late evidence without
mutating previous results. Duplicate streams, overlapping intervals and invalid
counts reject rather than producing partial ratios.

Supported intervals sum exact uint64 member deltas using arbitrary-precision
integers, so the total can exceed uint64. Output retains decimal `calls`, `errors`
and the exact unreduced `fraction`. `percent` is a floating-point presentation
value; the UI rounds it to two decimal places and also displays exact counts.
A supported zero denominator returns **no_data / zero-traffic**, calls/errors 0
and no fraction or percentage. Zero traffic is not 0% errors. Here traffic means completed invocations;
it says nothing about in-flight work, unique requests or business demand. The chart draws
separate supported intervals and distinguishes unknown from idle intervals; it
does not interpolate gaps or average percentages across time or instances.

## Explicit zero baselines

On an operation's first use, goo's instrumentation runtime adds zero to its
`ok`, `error` and `panic` counters. These are real SDK observations, not zeros
inferred by a query. The three known outcomes are initialized once per
operation/role; duration histograms are not filled with zero observations.
Never-invoked operations remain unobserved. A cumulative baseline still needs a
later collection before establishing a delta. Missing exported points or outcomes
still produce unknown ratios; initialization does not prove ingestion completeness.

The live sample exports cumulative counters every two seconds. Old rows from
before initialization or a restart can coexist with new rows; unsupported epochs
stay unknown. Delta evaluation is fixture/property tested, not a separate live
SDK deployment. The counter source limit is now 3,000 raw points per contract
(previously 1,000), accommodating nine initialized outcome streams over the default
five-minute window. The 8 MiB body guard, readonly typed SQL and rejection of
Float64 counter values >=2^53 remain. Exceeded guards fail explicitly without
partial totals. Histogram source limits are unchanged.

## Investigation boundaries and verification

The guided step shows the selected interval's exact counts and offers each
configured error outcome. Span selection uses operation, role, outcome and span
start time against the already captured operation response. It may find fewer,
more or no spans because span and metric capture are independent. It does not
identify exact counter contributors or prove cause. Opening a trace uses the
existing scoped waterfall, not a complete distributed-trace fetch. Switching
packages clears ratio evidence and stale responses cannot replace the new scope.

Go tests cover exact counts above uint64 after summation, zero traffic, missing
outcomes, skew, unknown/reset intervals, identity/unit/temporality mismatch,
duplicates, malformed counts and immutable/order-independent bases. Hegel runs
100 state/swarm sequences up to 30 steps over actual counter and ratio evaluators,
including calls, restart, gap and missing-outcome transitions. The instrumentation
state/swarm test checks explicit outcome baselines. Quint validates six observed
traffic/missing/idle/reset/gap/late-repair steps and rejects four invented ratios.
These are bounded checks, not a proof of arbitrary production histories.

Journey 15 executes real successful and failed checkout calls, waits for collected
metrics, checks the shared counter basis, follows error context and opens a trace.
Screenshots are generated from the live UI. Fleet membership/coverage, automatic
alignment/rebucketing, burn rates, alarms, statistical change detection and
Databricks metric lowering remain future work.
