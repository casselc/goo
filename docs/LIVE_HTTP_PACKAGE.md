# Live HTTP package: OTel → Go → shared view

The investigation UI's **Bound HTTP package · Go** panel now renders the aggregate
view descriptor emitted by `prebuilt.Expand`. It queries the Go `packages` service,
which reads actual OTel exporter spans from ClickHouse, prepares requests and
executes the shared reference evaluator. Browser code displays the descriptor's
dimensions/measures and returned states; it does not classify errors or calculate
ratios independently.

## Explicit source binding

The sample's `goo.http-sample` instrumentation scope/version `0.1.0` emits server
spans with `http.request.method`, `http.route` and `http.response.status_code`.
`POST /reserve` records successful acceptance as 202. A controlled
`GET /demo-failure` endpoint returns and records an actual 503 for the live journey.
Uninstrumented endpoints are not included; this is not a claim of total service
traffic or complete auto-instrumentation coverage.

`config/http-inventory.json` declares the semantic contract and resource scope:
`service.name=lab-web`, `deployment.environment.name=local`. The binding covers
those resources across instances; scoped trace/span identity distinguishes requests.
`config/http-package.yaml` supplies the explicit 5xx policy and aggregate target
bindings. It does not publish a business objective or alert.

The pinned exporter table retains `ScopeName` and `ScopeVersion`, but **does not
retain the OTel scope schema URL**. Compatibility with semantic-conventions 1.37.0
is an explicit library binding, not an observed schema-version claim. The SDK
scope declares that schema; the exporter drops it. The UI reports this limitation.
Other instrumentation libraries need their own reviewed scope/version/schema
mapping and conformance evidence.

## Selection and evidence contract

- Source adapter uses one fixed, half-open span-start window `[fromMs,toMs)`,
  at most 120 minutes. Resource keys/values, scope, version and times use typed
  ClickHouse parameters; the table/columns are fixed in the adapter.
- Server spans only. Request method, route template and nullable status are
  normalized into the existing package preparer. Raw URL is never a dimension.
- Exact decimal timestamps and valid nonzero OTel hex trace/span identities are
  checked. Returned timestamps must remain within the requested bounds.
- Identical source records deduplicate by trace/span identity. Conflicting copies
  fail the request; they do not choose an arbitrary healthy value.
- At 1001 raw rows, the adapter refuses to evaluate the 1000-record scope rather
  than display partial-population counts. The source payload is also bounded to
  8 MiB and query execution to ten seconds. Narrow the window on overflow.
- Missing/malformed response status remains unknown; missing route retains its
  unknown bucket. Sampling and ingestion completeness remain unverified.
- Evaluation identity includes expansion revision, bounds and source content.
  It is not an immutable custody snapshot. Late arrivals can change a rerun.

Generated SQL/Databricks plans still target declared prepared tables. This live
path normalizes/evaluates in Go without writing those tables, so it does not make
the export plans deployment-ready. Generic expansion diagnostics describe source
prerequisites; the live response also includes the actual query, parameters,
source binding, prepared facts and result provenance.

## Run

The Compose stack now includes a Go `packages` service (loopback port 8089).
The Python query service proxies `/api/http-package` to it for same-origin UI
access. The normal Dockerfile builds the Go executable in a separate build stage;
native verification and Orchestrion dependencies are not in the runtime module.

```sh
docker compose up --build -d
curl 'http://127.0.0.1:8088/api/http-package?minutes=5'
npm run test:journeys
```

The managed cloud build uses the existing offline-wheel image workaround plus a
host-built static `goo` binary copied to `/usr/local/bin/goo`. This actual image
was built/run here; the networked multi-stage Docker build was not run here.
When recreating the offline setup, compile the binary before building images,
copy the two new HTTP config files, and tag the image for `packages` as well as
the Python services. Preserve Compose proxy bypasses, including `packages`.

## Executed journey

Create a real reservation, request the controlled failing endpoint, wait for
collector ingestion, then evaluate the package in the browser. Assertions check
both routes, successful/failure counts, exact descriptor-driven columns, visible
coverage warnings, source binding, revision and policy. Frames:
[shared view](journeys/screenshots/07a-live-http-package-view.png) and
[source/policy](journeys/screenshots/07b-http-source-and-policy.png).

Unit tests cover typed source parameters, exclusive time bounds, invalid IDs,
duplicates, conflicting copies, missing outcomes and overflow rejection. Hegel
adds generated duplication/conflict checks. No custody lifecycle or storage
protocol was added; existing lifecycle Quint models do not cover this adapter.
