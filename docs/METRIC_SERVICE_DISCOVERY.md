# Discover services from native metric data

Implemented extension of [service browsing](SERVICE_BROWSING.md). Services →
**Discover services** discovers emitting services from spans, logs and
native metrics in the selected environment/lookback window. This explicit action
opens the [unified catalog](UNIFIED_SERVICE_CATALOG.md) directly. Open a service
to inspect its operations and instruments; an empty span inventory does not hide
services emitting another signal.

The Go `/api/discovery` adapter groups gauge, sum and explicit-histogram exporter
records by service, then combines those record counts into one `metrics` signal
row per service. Values, units, stream identities and histogram observations are
not combined. For example, two gauge points, one counter point and one histogram
point with `Count=5000` produce **4 metric data-point records**, not 5000 requests
or a sum of the metric values. Unsupported instrument kernels still contribute
presence observations and remain visible in instrument discovery.

All source families use the same parameterized environment and half-open fixed
window `[fromMs,toMs)`. Native metric time is the point's `TimeUnix`, not its start
time or ingestion time. First/latest event times and event age describe captured
points only. Counts remain decimal strings; UI presentation limits never define
totals. A single global limit of 1000 service/signal pairs fails the read rather
than publishing a partial catalog. Malformed/duplicate signal rows and out-of-window
timestamps are rejected by the source adapter.

The catalog offers **Open service** for metric-only emitters. It never sends a
`signal: metrics` question to the record executor. The existing service browser
retains the discovery window while independently reading span/log activity, span
attributes and instrument metadata. No spans/logs observed remains an unknown
coverage state; an operation inventory that was not read is labelled unread.
**Preview measurement** uses the native typed definition, identity, unit and
original bounds without a telemetry read. **Run measurement** executes explicitly. Catalog presence and kernel support do
not establish that observed records have valid stream identity; native execution
continues to require `service.instance.id` and the existing identity/unit/flag checks.
**Back to service** returns locally. Search filters catalog rows locally while
retaining the original totals and evidence.

Reported dependency edges still come only from exact captured trace/parent-span
identity. Metric-only services create no topology edges. Discovery does not infer
service health, declared membership, unique streams, throughput, process utilization
or a complete fleet. Metric summaries and exponential histograms are outside the
current adapter families, as in the existing native instrument browser. Durable
navigation/draft recovery remains future work.

```mermaid
flowchart LR
  OTLP[Ordinary OTel resource service and environment] --> Exporter[ClickHouse exporter]
  Exporter --> Spans[Span and log service groups]
  Exporter --> Points[Gauge / sum / explicit histogram point records]
  Spans --> Scope[Same fixed environment and half-open window]
  Points --> Scope
  Scope --> Catalog[Bounded service/signal catalog]
  Catalog --> Detail[Open service with original bounds]
  Detail --> Instruments[Exact native instrument identities and support]
  Instruments --> Preview[Source-free fixed-window preview]
  Preview --> Run[Explicit typed native measurement]
  Run --> Return[Local return to service detail]
```

[Storyboard 47](journeys/README.md#47--discover-metric-only-services) ingests ordinary
OTLP metric-only and unsupported-only emitters. Independent SQL checks exact
boundary inclusion/exclusion. The real UI checks point counts across three families,
absence of invented dependency edges, local search/return, native preview and gauge
execution, and desktop/mobile layouts. Hegel generates valid windows/point times
and proves the exclusive end is rejected by source validation. quint-go validates
the live journey ledger and rejects seven count, scope, identity, edge and read
mutations. Bombadil explores discovery → service → preview → local return.
See [VALIDATION.md](../VALIDATION.md) for actual runs.

[Automatic service dashboards](SERVICE_DASHBOARDS.md) now evaluate at most two
supported discovered instruments per overview. All other instruments stay in
the discovery inventory; kernels still withhold invalid identity/interval results.
