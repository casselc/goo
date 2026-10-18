# Core investigation basics

Status: implemented in the local lab. These extend the shared Go Explore query;
there is no second measure/query language for boards or saved questions.

## Span measures and record time charts

Explore offers **Span duration summary** and **Time chart** controls. Both are
opt-in. The JSON/YAML query additions are:

```yaml
summary: span-duration   # spans only; omit for count-only/log questions
bucketMs: 60000          # omit for no chart
```

The duration summary reads the full matching captured span population, independent
of record offset/limit. One source read returns count, UInt128 sum of UInt64
nanosecond durations, min/max, and p50/p95/p99. The exact mean remains `sumNs/count`;
the UI rounds milliseconds to three decimal places using integer arithmetic.
A future Decimal(38) warehouse adapter cannot represent every UInt128 sum: it must validate its narrower range or report unsupported values, never convert silently to float. No-data measures are unavailable, not observed zero durations. Values stay decimal
strings across Go/JSON/browser boundaries.

Percentiles use ClickHouse `quantilesExactHigh`: sorted zero-based index
`floor(p*N)`, clamped to `N-1`, with no interpolation. Thus four durations
`[0,1,2,3]` have p50 `2`. The independent Go reference uses the same declared
order-statistic rule. Future warehouse lowering must preserve this percentile policy or report a capability gap; it must not silently substitute approximate/interpolated values. These percentiles require bounded backend memory and may
fail when the query exceeds the source budgets; failure does not become a zero or
partial percentile. These are individual **span** durations, not inferred root
latency, request SLIs, complete traces, or native metric histograms. Scalar
percentiles must not be averaged to form a larger population's percentile.

Time charts accept 1 second, 10 seconds, 1 minute or 5 minutes, with at most
240 intervals. They group the full matching population into half-open intervals
anchored at the resolved question start. The final interval is clipped to its end.
Absent intervals become `no_data` with zero captured count; this is not a claim
about service health or ingestion completeness. Selecting a bar exposes its exact
bounds/count and, when requested, mean/p95. Bucket summaries are computed from that
bucket's raw durations, never averages of other percentile values.

The time chart is a separate aggregate read from total measures, groups and row
pages. Count disagreement preserves `changed_between_reads`; agreement does not
prove a transaction or stable telemetry. Counts/summary share their own read.
The result basis includes the measures/chart. See [general captures](GENERAL_CAPTURES.md)
for immutable retention, stable stored pages and refresh: summary/chart data is
archived with new captures. Old captures lacking these fields remain readable.

## Typed record measures and split charts

The existing controls now expand to a shared typed record-measure plan, used by
the SQL compiler and exposed in source-free planning. Opt-in `groupTimeSeries`
adds bounded group/interval aggregates and shared-scale local charts. Counts,
span-duration exact statistics and recorded-status partitions retain their existing
semantics; arbitrary numeric expressions remain future. See [the contract](RECORD_MEASURES.md).

## Reusable questions

A reusable question is the existing versioned **telemetry package**. **Create
package** from applied Explore results saves its selector, breakdowns, optional
summary/chart and default lookback. **Reusable questions** in Explore lists and
filters those revisions, then runs one in a fresh resolved window. This is distinct
from reopening frozen evidence. Definitions still provides YAML inspection,
preview, import/export and immutable revision editing. No separate mutable “latest”
question store, account ownership or private history is implied.

The package's `measure: captured-record-count` remains its population measure;
`summary: span-duration` adds the duration summary. Changing chart width does not
change the selected population. Definition validation applies the interval budget
to its default lookback; choosing an incompatible combination fails explicitly.

## Service signals, event age and reported dependencies

Under Services, expand **Signal activity and reported dependencies**. Its explicit
run performs two independent aggregate reads in the chosen environment/window:

- Service/signal counts and first/latest event times across span and log tables,
  including log-only services. Event age is relative to the fixed window end.
- Reported cross-service parent links. Child `TraceId`/`ParentSpanId` matches parent
  `TraceId`/`SpanId` within the same environment and window. A unique parent service
  gives an edge; conflicting service ownership gives `ambiguous`; an absent parent
  gives `unresolved`; malformed reported IDs give `invalid`. Same-service parent
  records are reported separately as `internal`.

Counts describe child span records, not deduplicated requests or calls. Parent
resolution outside the window, async links, span-kind request grain, declared fleet
membership, external peer inference and native-metric service discovery are not
implemented. This observed graph does not establish causal completeness. Both
reads fail closed above 1,000 result groups and retain their physical plans and
scope in an explain disclosure. Service rows explicitly open a fresh fixed-window
span/log question. Discovery is not automatically archived by an inventory capture.

Event age **is not ingestion lag**. Collector drop/sampling, queue/backlog, clock
accuracy and exporter freshness remain unknown without additional pipeline signals.
A recent event cannot certify healthy ingestion; no event cannot prove an outage.

## Query execution controls

General Explore, Inventory, Discovery and capture creation share a per-process
limit of **four concurrent requests**, with no admission queue. Excess requests
return 429 with retry guidance; completed/cancelled requests release capacity.
The outer Go context is bounded at 60 seconds, with the existing narrower handler
and per-source-read limits. Existing source read, result-row, byte, memory and
window bounds still apply. This is not a global distributed or per-tenant budget;
legacy operation/native-metric/preview routes are not covered by this admission gate.
Local saved-evidence GETs remain available while live query capacity is saturated.

Explore and Services expose cancellation. The browser sends an ephemeral UUID in
`X-Goo-Query-ID`, forwarded by the Python proxy. `POST /queries/{id}/cancel` cancels
the matching Go context. Short-lived bounded tombstones prevent a delayed request
from starting after its cancellation arrived first; IDs cannot be reused for a
minute. ClickHouse receives `cancel_http_readonly_queries_on_client_close=1` on
source reads. Cancellation is best effort: it may race with completion and does
not revoke published evidence. Source limits remain the fallback bound. New scope
edits also cancel superseded main Explore/Services reads and suppress stale results.
Other independently scoped context reads retain their own bounds/stale guards.

Cancellation IDs are local capabilities, **not authentication**. Shared deployment
still requires an explicit identity/tenant model, authorized source scopes and
saved-object access, redaction, retention/quotas, audit and backups. Keep the lab's
loopback bindings; this milestone does not claim production authorization.

## Executable evidence and remaining basics

[Journey 33](journeys/README.md#33-core-measures-reuse-and-service-discovery)
uses real OTLP ingestion with 205 spans (versus 100 displayed), log-only service
activity and ambiguous/unresolved parent fixtures. Native tests cover malformed
source responses, exact arithmetic, no-data, query admission and actual HTTP
context cancellation. Hegel state/swarm and observed Quint traces with rejecting
controls complement a read-only Bombadil actor over saved measures/charts.

Next basics: typed arbitrary numeric measures and grouped/time-series comparison,
native gauge/general metric exploration, collector ingestion-health inputs and an
explicit identity/authorization design. Investigation documents, annotations and
multiplayer remain separate increments after the useful core workflows.

[Grouped duration measures](GROUPED_MEASURES.md) now extend the same summary to one/two breakdown fields, with exact ranking, retained capture data and package previews. Count-cohort comparison remains separate from duration ranking.

[Recorded span status and comparisons](ERROR_COMPARISONS.md) now add exact status-category aggregates, prior-window mean/p95/p99/error changes and saved two-population receipts. These remain captured-span measures rather than request/fleet SLOs.
