# 201 · Reported Kubernetes resource variants

Generated ordinary OTLP fixtures, not a Kubernetes API inventory, running cluster
agent, RBAC integration, physical-pod identity, owner relationship or health
assessment. No positive source or UI response was mocked. Equal pod names and
reported UIDs across different raw resource maps are not deduplicated into entities.

Functional source `bad6727` is deployed with binary SHA256
`2269e3510ec168225733191348b9778fa204b0c0c20791c52f05055a59834c1a`.
[Deployment receipt](../artifacts/kubernetes/deployment.json) retains actual
HTTP/MIME/SHA readiness. The later presentation-only revision is `ff12b8f`.
Only existing packages/query containers were restarted; traffic stayed running.

## Actual main journey

1. Open **Services & resources → Kubernetes** and explicitly **Find pods**.
   One source operation returns **18 full resource variants / 22 family groups /
   24 physical records or metric points**. Family counts remain separate for
   logs, spans, gauge, sum and histogram. Histogram point count is not the number
   of observations inside its buckets.
2. Full resource/schema variants preserve equal names across UIDs, clusters and
   namespaces; missing versus empty UID, namespace, cluster UID and service fields
   remain different. Trace resource schema is null/unavailable from that physical
   source; log and metric literal schema strings, including empty strings, remain
   distinct. Missing/empty pod UID is metadata-only, with no context action.
3. Cluster/namespace facets and search operate on held rows without reading.
   The selected primary has no service fields, cluster UID or namespace. Its
   reported cluster name is an unverified context, not verified cluster identity.
4. **Matching logs/spans** stages ordinary fixed questions with all reported
   attributes plus explicit missing-cluster-UID and missing-namespace predicates.
   Staging reads nothing. Explicit Run finds the exact emitted **4 logs / 3 spans**,
   including records with additional real `service.name=pod-context-app`.
   No missing service is synthesized as `service.name=''`.
5. **Back to Kubernetes** restores the same held inventory, local filter values
   and actual opener focus without a read. Matching attributes is not exact
   map/schema equality, ownership, all telemetry from a pod, or a causal join.

## Final actual frames

| Frame | Provenance |
| --- | --- |
| [201a resource evidence](../docs/journeys/screenshots/201a-pod-resource-evidence.png) | Fresh display fixture; full-width desktop details, raw values and family clocks |
| [201b mobile evidence](../docs/journeys/screenshots/201b-mobile-pod-resource.png) | Same fresh resource/read; contained attributes, schema and UTC timestamps |
| [201c workspace context](../docs/journeys/screenshots/201c-kubernetes-in-context.png) | Actual filtered Kubernetes inventory around the expanded evidence |

Desktop evidence spans the full table width; nested attribute values are no longer
squeezed into one outer-table cell. Visible contents pass X/Y bounds, readable
value-height checks and exact nanosecond tooltip checks. The evidence disclosure
can still be long on mobile; this is not a whole-product density claim.

## Actual acceptance receipts

- Main **12.4s**:
  [ledger](../artifacts/kubernetes/pod-observed-c31414e48a/ledger.json).
  Browser source costs are `1 inventory → 2 after logs Run → 3 after spans Run`.
  Local facets, search, disclosure, staging and retained Back add zero, with no
  planner POSTs. Source operation count is not a benchmark of physical rows scanned.
- Kubernetes owner **5.8s**:
  [ledger](../artifacts/kubernetes/pod-observed-930153d43f/ownership.json).
  Actual discovery bodies held after `route.fetch` cannot publish after an
  environment edit or resource-tab withdrawal. Two explicit discoveries are two
  telemetry operations; cancellation controls are separate.
- Services owner/regression **10.2s / 13.2s invocation**:
  [ledger](../artifacts/kubernetes/runtime-observed-d0f8f65630/services-ownership.json).
  Holds actual supporting-span and native-memory replies, switches to Kubernetes,
  then releases them. No hidden dashboard follow-on work or late publication is
  allowed. Explicit Open recovers; a completed view survives a tab round trip
  unchanged and still supports actual goroutine-only Apply. This uses a separate
  ordinary runtime fixture without captures.
- Focused existing host owner **5.1s**:
  [ledger](../artifacts/hosts/host-observed-3cc1f8c94d/ownership.json).
  The shared resource-owner extraction retains host late-response rejection.
  No full host main/capture or historical suite was repeated.
- Bounded actor **31.2s / 33.2s invocation**, nine phases and 428 snapshots:
  [summary](../artifacts/kubernetes/pod-observed-4252f0d44a/bombadil/acceptance-summary.json).
  A separate ordinary fixture proves local cluster/namespace/search edits,
  source-free context staging, one explicit logs Run with exact emitted members,
  and retained filters/focus. Terminal source cost is **2**: inventory plus Run.
- Independent actual validation **0.251s** (main 0.06s, Kubernetes owner 0.02s,
  Services owner 0.15s), **sixteen changed controls rejected**, no observer failures:
  [native log](../artifacts/kubernetes/native-validation.txt).

```sh
npx playwright test journeys/201-kubernetes.spec.js journeys/201-kubernetes-ownership.spec.js journeys/201-services-ownership.spec.js journeys/199-host-ownership.spec.js --workers=1
GOO_KUBERNETES_BOMBADIL=1 npx playwright test journeys/201-kubernetes-bombadil.spec.js --workers=1
```

## Diagnostic history and separate display replay

The first three cases in the initial invocation passed. The Services helper then
waited for a dashboard attribute before the first service Open, when that element
correctly did not exist. The runner was cleanly interrupted; this was not a combined
four-case pass. The snapshot now atomically returns null for no dashboard, and
only Services was rerun. Full interrupted trace/frame and log are retained under
`artifacts/kubernetes/first-services-diagnostics` and
[the first log](../artifacts/kubernetes/main-owners-first.log). No application
change or weaker late/source assertion was used for that harness correction.

Visual review separately found a real layout issue: the old disclosure occupied
only the first outer-table cell, wrapping desktop values into tiny fragments.
Its family labels also called metric points “records.” Original frames are
retained in `artifacts/kubernetes/first-display-diagnostics`. The reviewed passive
`ff12b8f` fix places evidence in a full-width companion row and labels Logs,
Spans, Gauge points, Sum points and Histogram points. Source/controller semantics
and the Go binary are unchanged.

A fresh ordinary fixture, **pod-observed-d166c357ff**, passed display replay
**9.6s / 12.6s invocation**:
[receipt](../artifacts/kubernetes/pod-observed-d166c357ff/display.json),
[asset deployment](../artifacts/kubernetes/display-deployment.json).
It performs one actual discovery read, no context Run and no custody operation.
Local toggle, staged context and Back preserve focus/filters with zero extra reads.
All final frames above use this separate fixture/window/read clock. Actual desktop
and mobile PNGs were inspected after the geometry checks passed. Accepted main,
Kubernetes owner, Services owner and host owner were not repeated for this fix.

```sh
GOO_KUBERNETES_DISPLAY=1 npx playwright test journeys/201-kubernetes-display.spec.js --workers=1
```

There is no new native Kubernetes metric evaluation, capture format or store in
this slice. Counts describe observed source records/points and bounded resource
variants, not live pod counts, completeness, permissions or health.

Final frame byte hashes are retained in
[the frame manifest](../artifacts/kubernetes/display-frame-manifest.json).
Normal diagnostic retention preserves compact ledgers and initial failures. Exact
task binary/helper files were removed; traffic and data volumes remain intact.
