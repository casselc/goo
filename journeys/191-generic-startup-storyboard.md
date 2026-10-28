# 191 — Start with ordinary OTel and no sample configuration

Current runtime: one compiled Go server with embedded UI/docs, started from an
empty directory. No Python gateway or copied source assets are used. The fixture
still emits Python-tagged OTel spans/logs; this describes the telemetry SDK
attributes, not Goo's implementation language.

The updated executable [generic-startup.spec.js](generic-startup.spec.js) passed in
238 (21.9s). It verifies explicit generic scope before reads, 150 spans including
140 server records plus 140 logs, automatic service discovery, fixed-question
staging, capture and source-free reopening. The separate offline supplement and
actor results recorded below are historical and were not rerun for 238.

## Reproduction

Build once and select the exact binary:

```sh
go build -o /tmp/goo-generic ./cmd/goo
GOO_GENERIC_BINARY=/tmp/goo-generic npx playwright test journeys/generic-startup.spec.js
```

The helper starts one process on loopback18088 and stops it in `finally`. Browser
assets are embedded; the writable capture directory is separate. See the
[new runtime journey](238-single-go-runtime-storyboard.md) for SSE capacity and
opposite-boundary consolidation acceptance.

## Current frames

- [Activity](../docs/journeys/screenshots/191a-generic-python-activity.png)
- [Discovered service](../docs/journeys/screenshots/191b-generic-python-service.png)
- [Reopened mobile evidence](../docs/journeys/screenshots/191c-generic-saved-python-evidence.png)

## Historical implementation and receipt record

The following describes the earlier two-process implementation and its original
acceptance. Process/asset receipts and timings are superseded for current startup;
retained semantic assertions remain useful history.

Status: delivered and live-tested. Source implementation `70e2882`, explicit initial scope correction `c9f0460`, and guided-example gating `1616c2a`. The isolated main, bounded actor, offline API supplement and refreshed lab compatibility smoke all pass.

The user starts the Go telemetry API and Python static/proxy UI from an isolated directory containing only `server.py`, `ui/`, `api/` and a writable `captures/` directory. There is no `grammar.py`, definition, operation binding, inventory, prebuilt or sample file in that runtime asset root. The executable lives outside that directory. Existing lab ClickHouse and OTLP endpoints are explicit test connections; credentials are excluded from receipts.

The generated fixture uses ordinary OTLP spans and logs with Python SDK resource attributes and Python instrumentation scopes. It is generated Python-tagged telemetry, not evidence that a separate Python application was instrumented or executed. There are no mocked positive source or UI responses and no goo manifest.

### Executable flow

1. Open the generic UI. The environment is blank and Activity waits for an explicit Start. Runtime capability/configuration GETs identify optional examples and declared operations as unconfigured; they do not query a population.
2. Ingest 150 spans, including 140 HTTP server records from `ordinary-python / python-orders`, plus 140 correlated logs. Start Activity explicitly and pause the supplied observation window.
3. Discover Services and open the Python service. The automatic HTTP view counts all 140 matching server-span records while its first record page contains 100. The declaration registry stays unconfigured rather than pretending to be a successful empty inventory.
4. Open a widget question without a source read, then Run explicitly. Capture the fixed 140-record population and reopen it without a new telemetry read.
5. Run a separate generic instance against an unavailable ClickHouse endpoint. A valid discovery request fails as a source error, optional families remain explicitly unconfigured, and the saved receipt still reopens with identical stored bytes.

### Reproduction

After the feature source is frozen, compile `cmd/goo` with the repository Go toolchain. Set `GOO_GENERIC_BINARY` to its absolute path:

```sh
GOO_GENERIC_BINARY=/tmp/goo-generic-191 npx playwright test journeys/generic-startup.spec.js
GOO_GENERIC_BINARY=/tmp/goo-generic-191 GOO_GENERIC_BOMBADIL=1 npx playwright test journeys/generic-startup-bombadil.spec.js
GOO_GENERIC_BINARY=/tmp/goo-generic-191 GOO_GENERIC_LEDGER=/absolute/path/to/ledger.json npx playwright test journeys/generic-startup-offline.spec.js
GOO_GENERIC_LAB_SMOKE=1 npx playwright test journeys/generic-startup-lab.spec.js
```

The harness owns ports 18088/18089 and stops both subprocesses in `finally`. Each run retains a sanitized runtime manifest, asset hashes, process logs and the complete capture store before removing its temporary directory. The offline supplement accepts an explicit successful ledger; it does not reuse an ambient capture store.

### Initial diagnostic

`artifacts/generic-startup/main-first.log` and `first-failure-diagnostics/` retain the first failure. Before fixture ingestion the UI automatically read existing `local` logs instead of presenting an explicit Start. The failing run's `generic-python-23e7cfa8c7/runtime.json` records successful subprocess cleanup. The startup/source-free assertion remains strict; this failure is not reported as a passing main.

The second attempt (2.8 seconds) confirmed the explicit idle gate but found a remaining guided-measurements example-catalog GET. `1616c2a` gates both example questions and example workflows while retaining user measurement definitions. Its diagnostic is in `second-failure-diagnostics/`. These were actual startup integration defects, not relaxed test expectations.

### Actual frames

These links now point to the refreshed 238 single-Go run of the same scenario.
The historical receipt values below remain the earlier two-process run, not the
provenance of the updated frames.

- [191a — Explicit ordinary Python Activity](../docs/journeys/screenshots/191a-generic-python-activity.png): paused local observation buffer, not a full-population total.
- [191b — Discovered Python HTTP service](../docs/journeys/screenshots/191b-generic-python-service.png): 140 server-span records, independent log volume and explicit source scope; no declared operation manifest.
- [191c — Mobile frozen evidence](../docs/journeys/screenshots/191c-generic-saved-python-evidence.png): the captured 140-record population, first 100 displayed and immutable paging.

### Accepted receipts and effects

- Main: **20.0 seconds**, 23.8-second invocation; `artifacts/generic-startup/main-third.log`, `generic-python-e570bbe8e2/ledger.json` and `runtime.json`.
- Initial page: seven store/configuration GETs, **zero POST/tail connections and zero example/workflow catalog fetches**. Live environment is blank. The service declaration registry remains explicitly unconfigured.
- Main browser ledger: ten source-owning POST operations (including the explicit fresh capture read), one explicit streaming connection and one pure dashboard plan. These are logical browser operations, not a count of backend SQL statements. Readiness probes for 150 spans and 140 logs, capability/refusal probes and OTLP ingestion are separate request-context setup operations.
- Widget navigation stages without a source read. Explicit Run returns 140 spans; capture makes its documented fresh read. Reopening and resizing frozen evidence make no telemetry POSTs. Every returned source row retains Python SDK language and the qualified namespace.
- Actor: **30.9 seconds**, 32.2-second invocation, six phases and **465 snapshots**; `generic-python-38477cf34c/bombadil/acceptance-summary.json`. It proves service140 → source-free staging → explicit fixed140, with eight source POSTs and one pure plan. This separate fixture intentionally contains spans only; its supported widget fanout differs from the main's spans-plus-logs fixture.
- Offline supplement: **2.0 seconds**, 4.6-second invocation; `generic-python-e570bbe8e2/offline/offline.json`. With the real source endpoint set to unreachable `127.0.0.1:1`, discovery returns 502, optional operations return 503, runtime capabilities remain 200, and the frozen capture GET remains 200 with identical content and full stored bytes.
- Full retained capture: `d204e5d5cac66cb7a6b01bd4265e47aa8599673ee1b256cc04c1cc5b29fe0398`, **145,983 bytes**, SHA-256 `5a7cccde32af802e66062fec59148f21012dbadadfd9bd27efb12f73bb369844`. `scope-capture.json` and the complete `capture-store/` are retained beside the main ledger.
- Invalid startup checks reject an unknown Go profile, partial explicitly supplied binding, explicit missing operation package, malformed Python port and unknown Python profile before listening successfully.
- Existing lab: refreshed only packages/query binary and current assets, preserving explicit operation/binding/inventory arguments and all data volumes. **6.0-second** smoke, 9.2-second invocation; `lab-first.log`, `lab-smoke.json`, `lab-deployment.json`. It confirms lab/configured capabilities, nonempty operation registry, legacy automatic `local` Activity and explicit pause.

Runtime binary SHA-256 is `a8d9b8e6b82cef589f9c205de5d1d4db472f0a13b6f8f299b202d94f05a782a9`; the same binary ran generic and lab profiles. Every isolated run records cleanup, copies its actual capture store and removes its temporary runtime root. The temporary build binary was removed after the lab refresh. The lab runtime now serves the frozen implementation; no ClickHouse/data volume was recreated.

The existing catalog/header density and wider whole-product usability backlog are outside this startup slice. This acceptance demonstrates generated Python-tagged OTel interoperability, not automatic instrumentation installation or deployment of a Python sample application.
