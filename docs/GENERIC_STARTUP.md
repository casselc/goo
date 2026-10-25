# Start goo without the sample application

The implemented `generic` startup profiles run the Go observation API and Python
UI gateway without loading the lab's prepared HTTP binding, example grammar or
operation packages. Standard OTel telemetry remains the input. This separates
optional example capabilities from ordinary Activity, discovery, typed questions,
service dashboards and stored evidence; it is not a deployment installer or a new
telemetry storage schema.

## Start the two processes

Build `cmd/goo` using the repository's supported Go toolchain, then start the API
with an explicit generic profile and a persistent evidence directory:

```sh
go build -o /tmp/goo ./cmd/goo
CLICKHOUSE_URL=http://127.0.0.1:8123 /tmp/goo serve-http \
  --profile generic --listen 127.0.0.1:8089 \
  --investigations-dir /path/to/goo-evidence
```

Supply the ClickHouse credentials required by that instance through
`CLICKHOUSE_USER` and `CLICKHOUSE_PASSWORD`. The API uses existing source adapters,
query budgets and read-only query semantics. Startup and capability discovery do
not connect to the telemetry source or establish that it is reachable.

From the shipped UI asset directory, start the Python gateway separately:

```sh
GOO_PROFILE=generic GOO_HOST=127.0.0.1 GOO_PORT=8088 \
PACKAGES_URL=http://127.0.0.1:8089 \
CLICKHOUSE_URL=http://127.0.0.1:8123 python3 server.py
```

The generic UI needs `server.py`, `ui/` and `api/` assets. It does not import
`grammar.py` or read `config/`, `definitions/` or the example catalog. Its default
listener is loopback; an explicit `GOO_HOST` changes the bind address. It validates
profile, host, port and explicitly supplied upstream base URLs before listening.
Both upstream URLs must describe the intended installation; changing the startup
profile does not configure authentication, create OTel tables or install a collector.

The profiles are independent: `GOO_PROFILE` configures Python, while `--profile`
configures Go. Set both explicitly. The existing lab defaults remain available;
this change does not turn a missing or broken lab configuration into generic mode.
The Compose lab still supplies its explicit prepared binding and operation packages.

## Capability states are configuration, not telemetry health

| Endpoint | Meaning | Source effect |
| --- | --- | --- |
| Python `GET /api/ui-config` | `goo/ui-config/v1`, profile and example configuration state | None |
| Go `GET /capabilities` (proxied as `/api/capabilities`) | `goo/runtime-capabilities/v1`, profile, `preparedHTTP` and `declaredOperations` states | None |
| Unconfigured prepared/declared-family GET | HTTP 503 with `code: capability_unconfigured` and the unavailable capability | None |
| Ordinary explicit observation question | Existing typed query and source/limit semantics | Actual read on execution |

`configured` means configuration loaded, not that telemetry exists or a service is
healthy. `unconfigured` is different from observed no-data, source failure and
unknown evidence. A capability request failure must not enable example actions or
manufacture empty observations. Generic views remain usable while optional example
areas explain that their capability is not configured.

Without prepared HTTP configuration, `/http-package` is unavailable. Without
operation packages, `/operation-packages`, `/operations`, `/counter-rates` and
`/histograms` return the explicit unavailable-family response. Python's example
package/catalog endpoints likewise refuse in generic mode rather than reading
missing assets. Normal discovery and pure dashboard planning remain registered.

## Choose scope before reading

Generic startup waits for UI configuration and opens Activity with a blank
environment and an explicit **Start stream** action. Opening the page does not
start a local-environment tail. Typing the initial environment changes local
intent; the first explicit start applies it. An unknown or failed UI-configuration
response also remains blank and does not automatically read telemetry. This gate
prevents an installation from silently querying the lab's `local` environment.
The lab profile retains its existing local automatic-start behavior.

## Add optional declarations explicitly

Generic Go startup accepts repeatable `--operation-package /path/package.yaml`
arguments. These load only the supplied registry declarations; no checkout package
is implicitly added. Prepared HTTP requires all four explicit flags together:
`--binding`, `--inventory`, `--scope-name` and `--scope-version`. Supplying a partial
set, an explicitly empty required value, a malformed file or an unreadable path
fails startup. A supplied invalid operation package also fails; it is not silently
ignored. Omission of optional configuration and explicit invalid configuration are
different cases.

Existing investigation and capture stores remain usable independently of the
current registry. A stored operation observation can reopen from its retained
receipt without telemetry reads. Refreshing it requires the matching package
revision: when unavailable, the existing HTTP 409 refusal preserves the original
question and evidence. It must not be relabelled as an unavailable-family GET or
re-executed under an unrelated declaration. Use the same evidence directory across
restarts when retaining this local store is intended.

## Verification boundary

Go startup tests run a child process from an empty working directory, verify
source-free capability/refusal wires and exercise optional configuration. Python
tests run from an asset-only tree with an import guard for the example grammar,
check explicit configuration failures and verify proxy behavior. These tests do
not prove source connectivity or production deployment. [Journey191](../journeys/191-generic-startup-storyboard.md) passed in an isolated
asset-only runtime: main 20s, a 30.9s generated browser actor (six phases, 465
snapshots), a 2s offline supplement and a 6s existing-lab regression. The main
used 150 spans and 140 logs with Python SDK resource attributes; this was ordinary
Python-tagged OTLP fixture data, not a separate instrumented Python application.
The overview measured all 140 matching server-span records with a 100-row page.

The offline process used an actually unavailable ClickHouse endpoint:
`http://127.0.0.1:1`. Valid discovery failed with 502, optional families remained
503, capabilities stayed source-free 200, and the saved receipt reopened with
identical complete stored bytes. A read-only audit matched all 132 runtime asset
hashes and the binary, confirmed the isolated root was removed, and matched the
retained capture store against its manifest and the complete capture copy.

Acceptance required two real startup fixes. The initial run automatically started
a `local` tail; the second still requested the disabled example catalog. Both
failed logs are retained, and the accepted main is the third run. The configuration
and no-read assertions were preserved rather than relaxed. These are bounded
local acceptance results, not production onboarding or source-connectivity proofs.

See [service dashboards](SERVICE_DASHBOARDS.md) for manifest-free record/native
population contracts and [architecture diagrams](ARCHITECTURE_DIAGRAMS.md#generic-startup-and-optional-capabilities)
for the startup boundary.
