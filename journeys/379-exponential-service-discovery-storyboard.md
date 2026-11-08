# 379 — find an exponential-only service

Implemented and exercised against the reviewed 379 Go binary. This journey uses
the retained real 374 Collector exports without ingesting replacement telemetry.

1. Open Services, expand Scope, choose the exact UTC window and namespace, and
   Discover services. The inventory identifies 13 exported metric points for the
   exponential-only service. Histogram bucket mass is not a service record count.
2. Open the actual service row. Native instruments remain observed telemetry with
   unsupported evaluation; no healthy service or executable definition is inferred.
3. Use Discover metrics, then the explicit Discover instruments control. The exact
   fixed service/environment/namespace scope reaches eight instrument identities.
4. Inspect metadata on a narrow viewport. Unsupported instruments have no Build,
   Preview or Run action. Metadata inspection and local search do not reread data.

Actual frames: [desktop Services inventory](../docs/journeys/screenshots/379-exponential-service.png)
and [390 px metadata inspection](../docs/journeys/screenshots/379-unsupported-service-instruments-390.png).
The final desktop frame uses the shared compact table, separate namespace label
and short Open service action. Scope controls also fit 390 px without source reads
on resizing. Screenshots are captured from the running app; actor DOM completion
alone does not establish that a Bombadil JPEG has current paint.

The executable journey independently reads raw source exports, checks ordinary
source-family counts, rejects capture/package writes and executable exponential
definitions, and runs the positive actual-control Bombadil campaign when enabled.
It also exercises missing and reported-empty namespace controls. Capability fault
checks inject an absent optional table or a failed table probe at the scoped test
proxy. They never replace positive telemetry responses. Optional-source absence
has an explicit diagnostic; present source with no matching data has none; failed
probes refuse the whole read.

Final actual journey passes in **58.8s**, including a **50.268s** Bombadil campaign
with **983** observed states and positive completion. Its exact scope, 13 exported
points/eight identities, unsupported disabled action, disabled Run and no-pan
checks all pass. The actor makes seven POSTs, including two catalog reads and two
instrument reads, and no execution/package/capture writes. Final actor paint was
visually checked against completed metadata; one verified JPEG is retained beside
the compressed trace and DOM report. Repeated JPEGs are removed, so their original
trace artifact paths are unavailable. The first run's final JPEG was stale and is
explicitly excluded from current visual evidence.

Final report: `artifacts/service-discovery379/525516fc-37d0-406f-a55f-7a3a73239d7e/summary.json`.
