# Local demo lab

The lab runs Goo, ClickHouse, an OpenTelemetry Collector, Redis and sample services.
Use it to explore the product without connecting your own infrastructure.

## Start

Install Docker with Compose v2. Allow roughly 4 GB of memory and several GB of
free disk. Linux on Intel/AMD is the primary tested target.

```sh
docker compose config --quiet
docker compose up --build -d
python3 scripts/smoke.py
```

Open **http://localhost:8088** and allow 20–40 seconds for sample telemetry.
The smoke script needs Python 3; Goo itself runs as one Go server. Published
ports bind to loopback and use disposable `lab` / `lab-local` credentials.

## Try an investigation

1. In Explore, watch Activity logs or switch to spans. Choose an environment or service.
2. Pause and inspect a record. Use its attributes to narrow the scope, then Apply.
3. Open a trace waterfall or analyze the selected scope. Run the fixed question.
4. Compare a selected population with the rest or a preceding time window.
5. Save a view to reuse its question and layout, or capture observations to retain evidence.

A saved view stores configuration. Reopening stages it; Run reads telemetry.
A capture stores observations and can reopen without querying telemetry. Returning
from a child investigation restores the held parent rather than silently refreshing it.
See the [illustrated investigation journey](journeys/investigation-reuse.md).

## Retention

The generated trace table uses calendar-day expiry (`toDate(Timestamp) + INTERVAL
1 DAY`), despite the exporter's `ttl: 24h` setting. Source records can disappear at
UTC midnight. Captures retain their saved evidence. Inspect the actual ClickHouse
DDL before relying on a retention window.

## Sample workloads

The independent reservation sample emits HTTP/gRPC/messaging telemetry. Explore
it through ordinary Activity, service dashboards and trace views; Goo does not
require its reservation attributes. The former reservation-specific query panel
and its Python API/compiler are retired. Modern YAML examples and Go-backed
operation/measurement workflows remain available under Examples.

Create a chosen scenario manually:

```sh
curl -s http://localhost:8081/reserve \
  -H 'Content-Type: application/json' \
  -d '{"scenario":"retry"}'
```

The gRPC service uses a real unary gRPC method `/lab.Reservations/Create` with an explicit JSON codec, avoiding generated protobuf application code in the minimal sample. OTel signals use standard protobuf OTLP/HTTP.

| Scenario | Behavior |
| --- | --- |
| good | Confirmation after about 0.3 seconds |
| retry | One failed processing attempt, then confirmation |
| late | Confirmation after about 14 seconds |
| missing | Accepted but deliberately never completed |
| reject | Terminal rejection |

The demo deadline is 10 seconds with 3 seconds of ingestion allowance. The fixed traffic mix is intentionally unhealthy and has a 95% target to make comparisons visible quickly. It is not a realistic workload or recommended target.

## Optional HyperDX

```sh
docker compose --profile hyperdx up --build -d
```

Open **http://localhost:8080**, create a local account, and use the preconfigured **Lab** ClickHouse connection. If not prepopulated, use host `http://clickhouse:8123`, user `lab`, password `lab-local`. Add log and trace sources for database `default`, tables `otel_logs` and `otel_traces`; accept the detected mappings. Add a metrics source for the `otel_metrics_*` tables if your HyperDX setup asks for one.

HyperDX 2.41.0 is specified explicitly. The optional profile needs MongoDB and extra memory. The upstream collector is deliberately kept independent of HyperDX authentication/OpAMP. These are baseline exporter schemas, without newer HyperDX-specific materialized views, session replay, or PromQL tables.

## Ports and services

| Port | Purpose |
| --- | --- |
| 8088 | Grammar UI and query API |
| 8081 | Sample HTTP endpoint `POST /reserve` |
| 8123 | ClickHouse HTTP, local credentials required |
| 4318 | Collector OTLP/HTTP |
| 8080 / 8000 | Optional HyperDX UI / API |

The gRPC service and Redis stay internal to Compose. The lab combines Python
sample workload services with the single Go product server and woven Go samples.
ClickHouse and Redis use named volumes; no kind cluster is required.

## Operational limits and stopping

This is a development observability workbench. Telemetry absence is not health;
source/event time, read time and immutable custody time remain distinct. Scalar
percentiles are not averaged into fleet percentiles. Captures are bounded,
independent reads; historical evidence is preserved rather than silently refreshed.
The reservation lab's SLO/alarm is an on-demand preview, without scheduling or
notifications; terminal telemetry and queue acknowledgment are not transactional.

```sh
docker compose --profile hyperdx down
```

Named volumes preserve lab data. Removing volumes is a separate explicit reset.
Deployment-specific authentication, tenant isolation, quotas and retention require
a named target and acceptance checks. Consult the [active roadmap](ROADMAP_2026_11_13_CHECKPOINT_426.md) for
planned work; optional storage research does not imply a deployed data platform.

