# 195 — Cumulative stock observations

Delivered integration, source `6e88d57`. Generated ordinary OTLP fixture telemetry
passes through the actual collector and ClickHouse; positive source/UI responses
are not mocked. This slice adds generic cumulative nonmonotonic sum observations,
not an automatic pool dashboard or a stock-to-record interval action.

## Executable flow and actual evidence

```sh
npx playwright test journeys/cumulative-stocks.spec.js --workers=1
GOO_STOCK_LEDGER=artifacts/cumulative-stocks/cumulative-stock-c553c6dc4d/ledger.json \
  npx playwright test journeys/cumulative-stocks-bombadil.spec.js --workers=1
```

The main first pass takes **9.5s** (13.0s invocation). Its retained ledger is
`artifacts/cumulative-stocks/cumulative-stock-c553c6dc4d/ledger.json`, with raw
OTLP, original point facts, discovery, fixed question, result, fresh capture,
reopened result, events and actual browser requests/responses. The executable
fixture/helper is [195-stock-fixture.js](195-stock-fixture.js).

1. Discovery observes cumulative stock support and a separate unsupported delta
   nonmonotonic sibling. Build previews a package without reading its values.
2. Explicit Run reads eleven raw stock reports in seven exact streams, preserving
   pool/state/instance identity. Decreases 10→7→3 are observations; the last changes
   accumulation epoch. Another stream reports 2.5→−1.25. Missing and conflicting
   reports are unknown; old reports are stale relative to the fixed window end.
3. Finding a stream, opening its selected observation and resizing are local.
   The inspector shows UTC observation and accumulation epoch, and explicitly
   offers no interval record-context action.
4. Capture performs a fresh full measurement read. Reload and API reopen are
   source-free and keep the same fixed evaluation and stock samples.

| Action | Browser logical source operations |
| --- | ---: |
| Stage fixed question | 0 |
| Discover | 1 |
| Preview package | 1 |
| Run stocks | 2 |
| Select observation / desktop and mobile framing | 2 |
| Fresh capture | 3 |
| Frozen reopen | 3 |

Setup ingestion/readiness probes are separate from these browser effects.
Logical operations are not underlying ClickHouse statement counts. The actor
uses the accepted facts and fixed question with no new ingestion: **30.3s**
(31.6s invocation), **six phases / 509 snapshots**, one explicit measurement
read, terminal proof true. Receipt: `bombadil/acceptance-summary.json` adjacent
to the main ledger; logs are `main-first.log` and `actor-first.log` in
`artifacts/cumulative-stocks/`.

A separate source-free dashboard-plan receipt `dashboard-plan.json` accepts the
actually discovered stock as a generic native widget. Its initial supplemental
request added a namespace absent from the discovered definition and correctly
received 400; the corrected request uses the exact discovery scope and returns
200. This is a harness scope correction, not an application defect. The checked-in
main now includes that same pure request for future runs; the accepted browser
main was not repeated for this API supplement.

## Frames and custody

[Desktop selected signed observation](../docs/journeys/screenshots/195a-cumulative-stock-observations.png)
· [Mobile UTC and epoch](../docs/journeys/screenshots/195b-mobile-stock-epoch.png).

Both frames use the accepted main fixture and real DOM/styles. They show a
component, not a whole-product density improvement. The mobile retained-details
table pans internally; the page itself remains within the viewport.

Capture ID: `81babd515806c6fff7c47f57e5a4a081673326e6f18b7c929ce027ea10de94fa`.
The full stored representation is **12,964 bytes**, SHA256
`599056e688d0ed0627049978d9d89d8f09191359ad4b13f0faee191330cbd37b`.
Actual `capture-before.json` and `capture-after.json` sidecars are byte-identical.
The public capture is a measurement receipt; no new capture format is introduced.

Deployment receipt: `artifacts/cumulative-stocks/deployment.json`; frozen binary
SHA256 `56a5b6df5aa9c65926d3966d650797138d9bd08f59c0537d649c107d5d87b37e`.
Actual asset HTTP status/MIME/SHA and server hash are recorded there. Only the
existing query/packages containers were restarted; telemetry data was retained.

Independent actual-ledger validation first passes **0.037s** (body 0.02s),
rejecting six guaranteed changed controls. It independently checks all eleven
physical reports/seven streams, exact values/epochs/identity, unknown/stale states
and full stored capture hash/reopen. Log:
`artifacts/cumulative-stocks/native-validation.txt`. No new state model or source
reads are introduced by that offline check.

## Interpretation limits

`pool` and `state` are generic supplied fixture labels. Signed/fractional values
under the chosen metric name test stock transport and presentation; they do not
establish pinned semantic-convention compliance or valid physical pool counts.
No capacity, saturation, fleet total, rates, interpolation, contributor identity
or pool health is inferred. Epoch changes do not become counter resets.
Unknown/stale states are not zero or healthy. Freshness remains evaluated at the
fixed window end when evidence is reopened. Existing native-view verbosity is
not claimed solved.
