# 202 · Exact pod memory observations

Generated ordinary OTLP fixtures, not a running kubeletstats receiver, kubelet
connection, Kubernetes API client, RBAC integration, or proof of source
instrumentation authenticity. No positive source or UI response was mocked.
Reported pod UIDs/full resource maps remain observations rather than physical-pod
identity, ownership, capacity or health.

Functional source `454088a` is deployed with binary SHA256
`f3e5c83a342f5b737361b8cd3eea37779dcf5c44aca1d2c6cb19e1e50a88cd9c`.
[Deployment receipt](../artifacts/pod-memory/deployment.json) records twelve
actual served asset/spec HTTP/MIME/SHA checks, server/schema/source-guide copies
and metadata readiness. Only existing packages/query containers were restarted;
traffic and data volumes stayed intact.

## Actual journey and pinned identities

1. Find reported Kubernetes resources explicitly. Trace-null schema rows cannot
   start native lookup; missing/empty UID remains metadata-only. Select a
   service-less full resource with a **known empty resource-schema string**.
2. **Observe metrics** explicitly reads exact-resource metadata, pure-plans it,
   and performs at most two default native reads. The primary lookup contains ten
   metric observations, including descriptor controls. Defaults are precisely:

   - `k8s.pod.memory.usage`, gauge `By`, latest **32,768** (earlier 65,536).
   - `k8s.pod.memory.working_set`, gauge `By`, **16,384**.

   Receiver scope is
   `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver`,
   version **`0.137.0`**, scope schema **empty**. A `v` prefix is not inferred.
   The question pins resource/schema, scope/version/schema and metric attributes;
   no service or instance identity is manufactured. Same UID in a different
   cluster map or resource-schema variant does not broaden this exact read.
3. Inspect the existing native measurement owner and return with the same panel,
   local filters and opener focus, without a telemetry read. Explicit native Run
   reads the exact usage question. Individual capture performs a fresh native
   read and publishes the ordinary immutable measurement capture; no composite
   pod capture/store is introduced. Gauges have no interval-to-record action.
4. Return to the held panel source-free, then explicitly inspect eight additional
   resource rows. Each uses one metadata plus one native read. Partial usage is
   useful without an invented working-set zero. The separate semconv profile
   pins scope schema 1.37, raw scope `ordinary.pod.memory` and version `fixture-1`:
   usage **12,345 By** is selected; its reported working-set name is not aliased.
5. Real stale/conflict/missing cases remain stale/unknown; no healthy zero is
   inferred. Negative **−1**, fractional **1.5**, and unsafe **2^53** byte gauges
   remain raw observed evidence in the native result. The strict pod presenter
   labels them invalid and paints no numeric timeline value. This is presentation
   refusal, not source filtering or changed generic gauge semantics.
6. Reopen the actual individual capture with a guard that would abort unexpected
   API POSTs. It makes no attempted POST and uses actual store GETs; full stored
   bytes remain identical. This proves a source-free frozen path, not an actual
   ClickHouse outage or global telemetry drain.

## Actual frames

| Frame | Evidence |
| --- | --- |
| [202a desktop panel](../docs/journeys/screenshots/202a-pod-memory.png) | Accepted primary receiver pair, independent labeled byte axes and UTC observations |
| [202b mobile panel](../docs/journeys/screenshots/202b-mobile-pod-memory.png) | Same actual read; complete controls, values and plots contained |
| [202c workspace context](../docs/journeys/screenshots/202c-pod-memory-in-context.png) | Actual selected resource panel in the Kubernetes inventory |

Both-width visible contents pass X/Y containment and 90px plot checks, with raw
value and exact nanosecond timestamp tooltips. Actual desktop/mobile PNGs were
inspected and approved; no separate display replay was required. Independent
usage and working-set observations are not summed, ratioed or treated as an
atomic snapshot. This panel does not claim to solve wider catalog density.

## Actual receipts and costs

- Main **14.8s**, five-case owner **11.6s**, focused existing host owner **7.8s**
  and Kubernetes inventory owner **4.1s** all passed in the first **41.5s**
  four-spec invocation:
  [main](../artifacts/pod-memory/pod-memory-5979da1993/ledger.json),
  [pod owner](../artifacts/pod-memory/pod-memory-e424b30f5f/ownership.json),
  [host owner](../artifacts/hosts/host-observed-63b524952b/ownership.json),
  [inventory owner](../artifacts/kubernetes/pod-observed-cb64e50395/ownership.json),
  [log](../artifacts/pod-memory/main-owners-first.log).
- Main source journal: inventory **1**; explicit metadata plus two defaults **4**;
  staging/Back stay **4**; native Run **5**; fresh individual capture **6**;
  eight separately requested boundary rows add two each, ending **22 source
  operations / 11 pure plans**. These are application operations, not physical
  SQL statement/scan counts. The initial view keeps its two-default budget.
- Owner holds actual metadata/native responses after `route.fetch`, then withdraws
  authority by filter, Close, row switch, tab or scope change. A newly selected
  partial row can finish its own real read; releasing the old reply cannot hijack
  it or launch follow-on work. Cancel controls are separate from telemetry reads.
- Independent observed verification **0.545s** (main 0.41s / owner 0.12s),
  **fourteen changed controls rejected**:
  [native log](../artifacts/pod-memory/native-validation.txt). The initial observer
  DTO expected timeline decimal strings as float64; that observer type was
  corrected, with diagnostics retained and all raw/profile/mask/clock/source and
  capture checks kept strict. No app, fixture or browser repeat was needed.
- Corrected bounded actor **32.4s / 33.9s invocation**, nine phases and **449
  snapshots**, terminal **5 source operations / 2 pure plans**:
  [summary](../artifacts/pod-memory/pod-memory-5c0d43f61d/bombadil/acceptance-summary.json).
  It proves exact receiver values, local disclosure, explicit native Run, retained
  panel/search and actual opener focus on Back. It uses its own ordinary fixture.
- Individual capture ID
  `ea43a2aa8c6693e8f3e226ff43b628c51354d1330c36d6ffd337e8c7f6f5c454`,
  **6,062 bytes**, SHA256 before/after
  `3f09e5b7e315f322ffe65b41314feb63fbdc12aa23f5896cf474b89946b9070a`.
  Complete actual stored files are adjacent to the main ledger.

```sh
npx playwright test journeys/202-pod-memory.spec.js journeys/202-pod-memory-ownership.spec.js journeys/199-host-ownership.spec.js journeys/201-kubernetes-ownership.spec.js --workers=1
GOO_POD_MEMORY_BOMBADIL=1 npx playwright test journeys/202-pod-memory-bombadil.spec.js --workers=1
```

## Diagnostic history

The first actor reached phase 5 with four reads and one pure plan, then timed out:
its broad descendant `button` selector clicked a nested timeline mark rather than
Inspect measurement. The corrected selector names the widget's direct-child
button. Exact source/pure/terminal assertions stayed unchanged; only that bounded
actor was rerun. The failed trace/screenshots and Playwright diagnostics remain
under `artifacts/pod-memory/first-actor-diagnostics` and its original actor fixture.
No application source, accepted main or owner assertion changed for this fix.

Final screenshot hashes are retained in
[the frame manifest](../artifacts/pod-memory/frame-manifest.json). Normal
diagnostic retention preserves compact ledgers, full capture copies and initial
actor diagnostics. Exact task build/deploy helpers are removed; traffic and data
volumes remain running and intact.
