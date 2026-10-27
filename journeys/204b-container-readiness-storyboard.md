# Journey 204b — reported container readiness

Status: source implementation complete; actual OTLP/browser acceptance pending.
The [native prerequisite](204a-container-readiness-storyboard.md) remains the
source/API contract; [readiness design](../docs/CONTAINER_READINESS.md) defines
identity, receiver scope and honest non-goals.

1. Open Kubernetes discovery with an explicit environment and fixed UTC window.
   Retain exact reported resource variants and known resource schemas.
2. Choose Observe metrics on a container-bearing variant. Read metadata once,
   derive the plan without telemetry reads, and execute at most two defaults.
   Memory keeps priority; a cluster-only variant can show phase and readiness
   independently. No implicit pod/container aggregation occurs.
3. See Ready or Not ready with separate observed/stale/unknown quality, UTC time
   and expandable reported history. A reported zero is observed Not ready.
   Missing/conflicting latest data cannot fall back to earlier Ready.
4. Preview an explicit candidate or inspect a default. Staging performs no read;
   explicit native Run does. Back restores the held panel and opener focus
   without a read. Container ID, pod UID, name, full resource/schema and point
   attributes remain fixed.
5. Capture the individual measurement with a fresh explicit read. Reopen the
   immutable capture without source access; compare full stored bytes and hash.
6. Exercise replacement ID, same-ID resource/schema variants, unsupported codes,
   stale, conflict, missing and incorrect physical-profile controls. Withdraw
   held discovery/default/native work to verify no late panel or hidden reads.
7. Run the bounded Bombadil actor over Observe → Inspect → Run → Back, preserving
   artifacts and checking read ownership. Independently replay raw rows and
   capture bytes with rejecting controls; screenshots document actual desktop
   and mobile rendering.

Executable files: `204b-container-readiness.spec.js`, corresponding ownership and
Bombadil specs, and `verification/bombadil/container-readiness.ts`. Generated
ordinary OTLP must flow through the collector into ClickHouse; it is not a real
Kubernetes receiver/API/RBAC or readiness-probe integration test.

Actual results, read costs and frames will be added after acceptance.
