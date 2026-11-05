# Qualified surrounding logs — 329

Status: native journey and shared kept-context regression pass after full suites
and package/served-asset verification.

The older fixed-record **Surrounding logs** action now uses the same strict
qualified service question as kept observations. The action explicitly reads
fresh logs. Its own panel retains radius, paging and close/focus ownership.
Trace expansion continues to search the exact trace across services.

`journeys/329-qualified-log-context.spec.js` builds an isolated generic Go runtime
using the 191 helper and sends real OTLP logs to the existing collector. No
capture, definition or custody receipt is seeded. Three unique services exercise
named, present-empty and missing resource namespaces. Each case includes a
same-service foreign namespace and foreign environment distractor. A record-only
fixed question selects the anchor; the surrounding action drops its body filter
and returns only the exact qualified service.

The named case has 103 matching observations, proving independent total/page
counts, the 100-row page and offset-100 remainder. Default ±30 seconds clips to
the original lower bound and excludes an observation 40 seconds later. Changing
to ±1 minute resets offset and clips to both original bounds. Closing restores
the original action focus without a read. A held genuine response checks browser
abort and late-adoption refusal after Close; it does not establish physical SQL
cancellation.

The existing `328-kept-log-context.spec.js` supplies the second entry-point
regression, genuine UI Keeps and broader cross-service trace continuation. Its
30-second Bombadil campaign must reach verified phase 6 with one page, independent
DTO count/rows two, visible indexed rows under the applied owner, zero kept or
catalog reads/writes after instrumentation, and no overflow or violations.
Process exit zero alone is insufficient.

Run after full checks and packaging, with Chromium 154, Bombadil 0.7.9 and healthy
ClickHouse/Redis preflight, through the traffic-preserving wrapper:

```
GOO_KEPT_CONTEXT_BOMBADIL=1 bash scripts/run_journeys.sh journeys/329-qualified-log-context.spec.js journeys/328-kept-log-context.spec.js
```

Maintained frame: `screenshots/329-qualified-surrounding.png`, captured
from the actual running surrounding panel. Compact joined results remain under `artifacts/qualified-context329`.

Validation: combined wrapper passes both tests in 1.1 minutes (direct 329: 12.5
seconds; kept 328 including actor: 48.4 seconds). Main 329 UI makes exactly eleven
`POST /api/explore` attempts: three anchor reads, three context actions, one next
page, three radius changes and one held context read. One matching cancellation
control is `POST {}` with response 202. No custody/catalog writes occur. Named
context is 103 total / 100 first-page / 3 offset-100 rows, widening to 104; empty
and missing each return two, widening to three. Exact body/service/resource
namespace/environment membership, scope labels and independent visible indexed
row counts pass. Page errors are empty.

The unchanged 328 regression proves two genuine UI Keeps, immutable receipts,
named/empty exact context and the broader two-service trace continuation. Its
main source attempts remain discovery two, preview two, tail two and record pages
four; one held page is aborted with one matching cancellation control.

The configured 30-second actor passes in 30.171 seconds with 616 snapshots, phases
0–6 and verified phase 6. Independent raw review finds zero violations/overflow
and valid source/write budgets in every snapshot. Applied ownership contains two
indexed data rows plus two hidden detail rows; independent response count and rows
are both two. Initial stored opening remains separate from actor instrumentation.
The wrapper restored traffic and initially removed no raw files. Root then
independently reviewed the direct membership/cancellation ledger and every actor
snapshot; design accepted all eight maintained frames. After that approval, owned
raw actor/browser/store/runtime bulk was removed. Compact joined results remain
in `artifacts/qualified-context329/joined-summary.json`, alongside the retained
328 main and campaign summaries. No volumes or stream entries were reset.
