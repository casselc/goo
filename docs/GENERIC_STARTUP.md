# Start Goo without the sample application

Goo is one Go process serving its embedded browser UI, API documentation and typed
observation API. No Python gateway, Node process, separate asset deployment or
sample manifest is required. The `generic` profile leaves optional lab declarations
unconfigured while ordinary Activity, discovery, dashboards and stored evidence
remain available. See [runtime rationale](SINGLE_GO_RUNTIME.md).

## Start one process

Build with the supported Go toolchain, then start with an explicit profile and
persistent evidence directory:

```sh
go build -o /tmp/goo ./cmd/goo
CLICKHOUSE_URL=http://127.0.0.1:8123 /tmp/goo serve-http \
  --profile generic --listen 127.0.0.1:8088 \
  --investigations-dir /path/to/goo-evidence
```

Open **http://127.0.0.1:8088**. Supply `CLICKHOUSE_USER` and
`CLICKHOUSE_PASSWORD` for that installation. The binary works from an empty
working directory; browser assets, ordinary example definitions and local API docs
are embedded at build time. Generic mode does not load sample declarations or read
an external example grammar. UI configuration and API capabilities share the same
Go `--profile`; the retired Python `GOO_PROFILE`, `GOO_HOST`, `GOO_PORT` and
`PACKAGES_URL` settings no longer configure the product.

The default Go listener is loopback. Choose an explicit `--listen` for another bind
address. Startup validates declarations before listening, without probing telemetry.
Existing ClickStack-shaped source adapters and read-only queries apply; starting
Goo does not install a collector, create source tables or configure authentication.

## One origin and one execution policy

The browser sends `/api/*` requests directly to existing Go handlers. Native paths
such as `/explore` remain available on the same listener for API clients; both
surfaces use the same admission budget, cancellation registry and snapshot/store
owners. There is no HTTP forwarding hop or second body-size policy. JSON receive
has a 15-second transport ceiling and cancellation interrupt; source work and SSE
retain their own handler policies. An error during receive closes that request's
connection rather than draining an unfinished upload.

| Endpoint | Meaning | Source effect |
| --- | --- | --- |
| `GET /api/ui-config` | Profile and example configuration | None |
| `GET /api/capabilities` | Prepared/operation configuration states | None |
| `GET /health` | Process liveness only | None |
| Unconfigured prepared/declared-family GET | HTTP 503 with `capability_unconfigured` | None |
| Explicit observation question | Typed validation and actual read | Telemetry read |
| Retained evidence GET | Load existing immutable evidence | No telemetry read |

`configured` means configuration loaded, not healthy or observed. `/health` no
longer runs the historical Python SELECT 1 readiness probe. Use an explicit scoped
query to establish source access; process liveness cannot establish ingestion
completeness. Missing data, source failure and unconfigured capabilities remain
distinct states.

## Choose scope before reading

Generic Activity starts with a blank environment and an explicit **Start stream**
action. Typing changes intent; starting applies it. Failed/unknown UI configuration
also withholds automatic reads. Lab mode retains automatic local streaming and its
explicit prepared/operation declarations. The standalone component gallery is
`/ui-components`; browsable API docs are `/api-docs` and the spec is `/openapi.json`.

The obsolete reservation-specific JSON query panel and `/api/query`, `/api/explain`,
`/api/package` are removed. Ordinary YAML examples and Go-backed package planning,
native measurements and operation investigation remain supported. Sample
applications may emit any ordinary OTel data independently of Goo's runtime.

## Add declarations explicitly

Generic startup accepts repeatable `--operation-package /path/package.yaml`.
Prepared HTTP requires `--binding`, `--inventory`, `--scope-name` and
`--scope-version` together. Partial, explicitly empty, malformed or unreadable
configuration fails startup; omitted optional configuration stays unconfigured.
Lab mode retains its existing default declaration requirements.

Use the same evidence directory across restarts. Stored observations reopen from
retained receipts without querying telemetry; refresh requires the matching
package revision and otherwise retains its existing 409 refusal. This migration
does not rewrite stored bytes, identifiers or volumes.

## Verification

Go runtime tests cover embedded assets, prefix routing and capability behavior.
Executable generic/runtime journeys start the compiled binary from an empty
working directory and use real OTLP data. Current results and limitations are
recorded in [VALIDATION.md](../VALIDATION.md); earlier journey191 two-process results
are historical, not evidence of this migration.
