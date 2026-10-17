# Declared fleet membership and interval coverage

Goo can now declare expected instance membership and evaluate fleet calls/errors
without treating an absent member as zero. The UI separates source observation
from valid aligned interval coverage. Membership is fixed for the entire selected
window; it is not discovered from telemetry or a deployment inventory.

```yaml
fleet:
  name: checkout-coverage
  ratio: operation-errors
  members: [woven-demo, woven-peer]
views: [operation-overview, error-ratios, fleet-coverage]
```

The referenced ratio supplies the exhaustive calls outcome domain and error
subset. A fleet requires 2–4 distinct explicit instances, including the package's
selected resource instance. All members use the same service/environment and
instrumentation binding; instance is the only overridden source selector. The
sorted membership declaration contributes to the package revision. Optional
`description` is explanatory text, limited to 500 characters.

## Source and semantics

`GET /api/counter-rates` retains selected-instance rates/ratios and adds `fleet`.
At most four instance queries run concurrently; each member retains the existing
readonly typed SQL, 3,000-point/8 MiB guards, counter precision checks and outcome
validation. No broad service scan discovers members. A failed peer source is
unavailable and does not discard valid selected-instance results. The selected
instance's own source error still fails the request.

`metrics.EvaluateFleet` first evaluates outcome ratios independently per member,
then groups complete stream identity with only `service.instance.id` removed.
Other resource attributes, scope/version/schema, metric, unit and operation/role
must match; incompatible identities remain separate unsupported groups.

Every member must have valid outcome counts for the same exact interval. Endpoint
partitioning never prorates wide intervals to fit narrower ones. Missing members,
missing compatible operations, skew, unknown resets/gaps/flags or unsupported
outcome intervals prevent a total. Unknown intervals contain no fleet counts,
fraction or percentage. Per-interval member states explain why; `supported` counts
members valid for that exact interval, not a probability or ingestion guarantee.
The full evaluator accepts at most 10,000 member counter intervals.

When every member supports the interval, calls/errors are summed with exact
arbitrary-precision integers and the percentage is computed from summed counts.
A member with zero completed calls contributes a supported zero. An entirely idle
supported fleet is `no_data`, with no percentage. For example, 2/10 and 10/100
produce **12/110**, not the average of 20% and 10%. Cumulative/delta point semantics
are normalized independently before combining interval populations. No scalar
percentile, percentage, or rate is averaged.

Source states are `observed` (some counter streams seen), `unobserved` (successful
query with no streams), and `unavailable` (query/input unavailable). Observed does
not imply valid window coverage, deployment health, sampling completeness or all
operations observed. A baseline alone cannot establish an interval. Fleet bases
cover member evidence bases, policy, expected membership, window and output;
input/member order does not change the basis.

## Local journey and limits

The default lab deploys only `woven-demo`. Checkout deliberately declares an
expected `woven-peer` to make missing coverage visible using a real empty source
query. This is a coverage demonstration, not a second deployed service. Expand
**Fleet membership and coverage**, inspect the unknown fleet total and choose
**Explain fleet coverage**. Selected-instance error investigation still works.
Changing packages clears the fleet panel and rejects stale responses.

[Journey 16](journeys/README.md#16-declared-fleet-membership-and-missing-coverage)
executes the real sample, collector/source queries and browser interactions.
Aligned multi-member aggregation is verified with Go/Hegel/Quint fixtures, not
claimed as a live two-replica deployment. See [diagram 7](ARCHITECTURE_DIAGRAMS.md#7-fleet-membership-and-coverage-before-totals).

Go tests exercise weighting, exact totals above uint64, zero traffic, missing and
unavailable members, unknown/reset intervals, skew, resource mismatch, membership
validation and stable immutable bases. A race-tested HTTP adapter fixture verifies
peer failure preserves the selected instance. Hegel runs 100 state/swarm sequences
up to 30 steps through calls, one-member restart, gaps, missing member and skew.
Quint validates six observed aligned/missing/skew/reset/idle/recovered steps and
rejects four invented fleet totals. These bounded checks do not prove arbitrary
production histories.

Dynamic membership/activation windows, trusted discovery, automatic alignment,
rebucketing, rolling-deployment coverage and live fleet histogram views remain
future work. Independently timed SDK exports can be unaligned even when every
member is observed. The existing explicit histogram merge API remains separate.
There is no Databricks fleet lowering or production deployment validation.
