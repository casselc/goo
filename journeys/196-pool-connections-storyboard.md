# 196 — Named database pool observations

Delivered source `2ece59d`, pinned to OpenTelemetry semantic conventions 1.37.0
(database pool metrics remain development-status). Generated ordinary OTLP
fixture telemetry passes through the actual collector and ClickHouse. Positive
source and UI responses are not mocked.

```sh
npx playwright test journeys/pool-connections.spec.js --workers=1
GOO_POOL_BOMBADIL=1 npx playwright test journeys/pool-connections-bombadil.spec.js --workers=1
```

## Actual flow

The accepted main passes **23.9s** (26.7s invocation), ledger
`artifacts/pool-connections/named-pools-7d8247745c/ledger.json`, log
`artifacts/pool-connections/main-second.log`. Its raw fixture and request payloads
remain separate from the production model and DOM observations.

1. Services discovers 120 ordinary HTTP Server spans and four exact native
   instrument identities. The full supporting record page remains 100 rows.
   Incoming HTTP remains the selected record profile. Within the unchanged
   two-native-widget default budget, the HTTP duration widget stays first and
   the named-pool stock widget is second. No pool record profile is added.
2. Thirteen physical stock reports form ten exact streams. Five eligible pool
   identities retain all stream identity except the state dimension; two
   missing-key streams are excluded from pool assembly. `orders` stays separate
   across `worker-a` and `worker-b`. Reported `used`, `idle` and custom `reserved`
   remain separate. Worker B's reported idle zero differs from unavailable.
3. Decreases 10→7→3 and accumulation epoch changes remain observations. The
   archive report is stale; flagged missing and simultaneous conflicting reports
   are unknown. No missing state becomes zero. Local pool search, state filter,
   UTC/epoch disclosure and viewport resizing cause no source reads.
4. Open widget question stages the exact fixed native question with
   `stockDomain: nonnegative-integer`. Explicit Run reads that question; Capture
   makes a fresh full measurement read. Reload and API reopen are source-free,
   retain the domain and preserve the complete stored representation.

| Browser action | Logical source operations | Pure plans |
| --- | ---: | ---: |
| Open service overview | 9 | 1 |
| Local filters and disclosure | 9 | 1 |
| Stage native question | 9 | 3 |
| Explicit native Run | 10 | 3 |
| Fresh capture | 11 | 3 |
| Frozen reopen | 11 | 3 |

These browser operations are not backend SQL statement counts. Setup ingestion,
readiness, and direct API controls are recorded separately. The direct controls
read a signed/fractional generic stock successfully, then receive actual **502**
when explicitly requiring the nonnegative-integer domain over those values.
Malformed request/domain validation is distinct from this observed-source refusal.

Two deliberate caller-supplied negative metadata controls retain actual stream
identities: missing-key-only evidence and a valid witness placed after the first
25 preview entries. Both retain generic stock availability and withhold named-pool
family admission. They are planner boundary checks, not claims that a fabricated
positive source population was observed. Accepted request sizes are **2,281** and
**12,528 bytes**, below the unchanged **16,384-byte** route limit.

## Actor and frames

The bounded actor first passes **30.6s** (33.8s invocation), **eight phases / 478
snapshots**, final source operations **9**, terminal true and no page overflow.
It uses a separate freshly generated ordinary OTLP fixture and verifies service
opening, separate pool instances, custom-state filtering, local reset and UTC/epoch
disclosure with unchanged source effects. Receipt:
`artifacts/pool-connections/named-pools-b419ced9f4/bombadil/acceptance-summary.json`;
log `artifacts/pool-connections/actor-first.log`.

[Desktop pool counts](../docs/journeys/screenshots/196a-named-pool-counts.png)
· [Mobile states and epochs](../docs/journeys/screenshots/196b-mobile-pool-states.png).

Both frames use the accepted main fixture and actual DOM/styles. Geometry checks
keep every pool row within the viewport and reject row/page horizontal overflow.
The frames show the component; inherited whole-product control density is not
claimed solved.

## Capture custody and deployment

Capture ID: `1abfe1fc772a9bd5be1c5c8d0cd3c67f4f2d5de631e0fd996634e5b746e6d105`.
The full stored representation is **16,904 bytes**, SHA256
`e659f1ddbc9da4dad4e441e88b136b920849873e336765b7ae9219b454ae6b0e`.
`capture-before.json` and `capture-after.json` next to the ledger are byte-identical.
No capture format is added or rewritten.

Deployment receipt: `artifacts/pool-connections/deployment.json`, source
`2ece59d`, binary SHA256
`2d896fd13ac38b07391b145505fbeaf96d78749cf7c6cfadb81bca439e7a2a8b`.
Actual asset HTTP status, MIME and hashes are verified. Only existing query and
packages containers restart; no telemetry data or volumes are recreated. The
build binary is removed after installation. Traffic is not paused in this round.

Independent actual-ledger validation first passes **0.362s** (body 0.32s),
rejecting six guaranteed changed controls. It independently checks all thirteen
raw reports, ten streams/five eligible identities, two excluded streams, known
zero versus unknown, latest values/epochs/DOM, domain-bearing question and
capture, source refusal, metadata boundaries and complete stored bytes/hash.
Log: `artifacts/pool-connections/native-validation.txt`. No source calls or new
state model are introduced by that offline check.

## Preserved diagnostic and limits

The first main **11.9s** passes the browser/native/capture checks, then its final
26-entry negative metadata request correctly exceeds the existing JSON body
budget and receives 400. That invocation is not acceptance. Its trace, frame and
context are preserved under `artifacts/pool-connections/first-failure-diagnostics/`.
The harness removes redundant per-key display metadata while keeping exact stream
identities; no application body limit or preview guard changes. A bounded API-only
preflight on the same fixture (no new ingestion) passes **2,281 / 12,075-byte**
controls before the corrected full main. Receipt:
`named-pools-a02da30917/compact-control-preflight.json` under the artifact root.

Counts, state observations, pools, instances and read clocks remain independent.
No totals, utilization, capacity, pool health, contributor identity or inferred
cross-state time/epoch alignment is claimed. Custom nonempty states remain
literal. Missing preview evidence is not proof of absence; generic stock support
remains available. No stock interval-to-record navigation is introduced.
