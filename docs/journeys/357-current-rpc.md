# Observe the pinned current RPC population

The current RPC lens follows OpenTelemetry semantic conventions [v1.45.0](https://github.com/open-telemetry/semantic-conventions/tree/v1.45.0), commit `fb30ca17193b551473b1f43c315482918ab6987c`: reported Server kind and nonempty `rpc.system.name`. `rpc.method` is conditional, so missing, empty and `_OTHER` values stay distinct members. The legacy `rpc.system`/method lens remains a separate explicit interpretation. A dual-key record can belong to both; no aliases are unioned.

The genuine OTLP fixture mixes current, legacy, dual-key, client, missing/empty-system and conflicting environment records. Quoted service and namespace literals remain exact. The current-only destination contains five spans (15ms duration sum); effective scope adds one legacy-resource record, giving six spans (26ms). Independent raw ClickHouse IDs establish membership rather than comparing only totals.

1. Open the real saved inventory scope and service overview. Positive same-row evidence selects the current lens ahead of the legacy lens.
2. Select legacy and explicitly select current again. A scoped upstream refusal leaves the current question unavailable; it cannot broaden to All spans. Retry reads only that failed, unchanged question while healthy siblings retain their nodes and clocks.
3. Open the fixed widget question, run the supporting records and compare their IDs to the independent oracle. The Services workspace tab restores the held service with no read; the plain widget link does not add a separate Back control. A retained old Retry callback remains inert.
4. Choose All spans explicitly. A separate fixed interval containing only older-key records yields a known zero under the current lens, without automatic fallback.

Native `rpc.server.call.duration` evidence keeps its seconds, exact stream identity and supported histogram temporality. Legacy milliseconds, wrong units and wrong metric kinds remain separate. Reported span Error is not inferred RPC/protocol failure or deduplicated call throughput.

The canonical combined run passed in **110.168s**. Its required observed model passed in **22.74s**, admitting both actual cases and rejecting six corruptions, with zero skips. The focused Bombadil campaign passed in **35.163s**, with **530 snapshots** through all seven phases. Positive completion independently requires the visible recovered current population, unchanged healthy widgets, inert old callbacks and the expected pending availability of template actions.

The main page observed 70 API requests: 42 telemetry requests and seven overview plans. Each Retry read exactly one failed question; staging and Services return read none. Fixture setup and the separate empty-current public-API probe are outside those page counters. Both current and effective cases acquired genuine successful native delta and cumulative results, each retaining a known four-observation interval in seconds. Their partial window coverage remains explicit; their counts are not span counts.

Initial validation corrected an invalid fixture that mixed delta, cumulative and unknown temporality under one stream identity. Each variant now uses its own real instrumentation scope, with two valid cumulative points. It also corrected the actor’s HTML oracle to retain all data/provenance while treating the template action’s pending disabled/title state explicitly. A later combined model invocation lost its timed wrapper without a verdict; the same actual report passed alone in 13.14s. Removing repeated full native HTML from actor output reduced that final trace to 6.5MB while keeping the full comparisons in the browser; the final combined gate passed at the original budgets.

[Current source refusal](screenshots/357-current-refusal.png) · [Recovered current overview](screenshots/357-current-rpc.png) · [Distinct method values at 390px](screenshots/357-methods-390.png)

The final actor terminal image also shows the settled current overview: five spans, one recorded Error, mean3ms and two independent native lanes. Earlier attempt images are not used as final visual evidence.

No telemetry or successful source responses are fabricated; the fault boundary forwards healthy reads to the OTLP-backed source. The existing lab stores/volumes and 51 pre-existing pending entries were preserved.

With the lab and [verification tools](../../verification/README.md) available,
run the complete journey against a reviewed local build:

```sh
go build -o /tmp/goo-reviewed ./cmd/goo
GOO_GENERIC_BINARY=/tmp/goo-reviewed GOO_CURRENT_RPC_BOMBADIL=1 GOO_VERIFY_CURRENT_RPC=1 npm run test:journeys -- journeys/357-current-rpc.spec.js
```
