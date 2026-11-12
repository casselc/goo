# Fleet service triage

Executable: [service-triage.spec.js](../../journeys/service-triage.spec.js).
The real-source journey passes, including its 40-second Bombadil campaign.
See [validation](../../VALIDATION.md) and the
[independent review](../reviews/SERVICE_TRIAGE_REVIEW_412.md) for actual checks
and their limits.

A platform operator starts in **Services**, discovers one fixed environment/window,
then chooses **Measure server spans**. The real collector and ClickHouse contain
an ordinary OTLP fixture with current and preceding equal windows. Four qualified
`checkout` identities differ by named, missing and present-empty namespace;
log-only, metric-only and Client-only services remain visible. A previous-only
service and a new service exercise asymmetric window membership. An environment
precedence control is excluded even though its legacy environment matches.

| Step | Operator interaction | Executable check |
| --- | --- | --- |
| Discover | Set exact UTC bounds and discover services. | Original mixed-signal catalog remains the initial view; no summary read is implicit. |
| Measure | Read current and preceding Server-span summaries. | Two aggregate API calls, four SQL statements, qualified groups and no record/field projection. Independent raw ClickHouse rows establish count, status-error numerator, exact duration sum and delta. |
| Rank | Choose mean duration, then search for checkout. | Local sorting/search issue no API POST or SQL, keep namespace identities distinct, and preserve the chosen ranking. |
| Switch | Show Signal counts, then Server summary. | Both presentations reuse retained source results without rereading; Enter activation restores keyboard focus after each local redraw. |
| Inspect | Open the present-empty namespace checkout service. | The existing service overview receives the original typed environment, UTC window and exact empty namespace. After all actual overview widget/API/SQL reads settle, Back restores the triage table/ranking and visible opener without another read. |
| Refuse | Plant an upstream source failure on a deliberate reread. | No rankable partial/zero summary is produced; ordinary signal counts remain available. |
| Retire | Hold an upstream read, edit environment, invoke the retired callback, then release the fault. | Scope invalidation prevents adoption and rejects the stale action; no callback can restart the abandoned read. |

Actual documentation frames, generated only by the running app:

- [Desktop table](screenshots/service-triage-1440.png): compact qualified table and ranking.
- [390px cards](screenshots/service-triage-390.png): three aligned stats, identity and one action.
- [320px cards](screenshots/service-triage-320.png): contained narrow-screen composition.

With `GOO_SERVICE_TRIAGE_BOMBADIL=1`, a separate 40-second actor repeats real
discovery/measurement and local ranking, search and presentation switching. Its
completion invariant requires all nine services and forbids unexpected writes or
local source reads. This is opt-in; do not claim a campaign passed without running it.

The journey writes `artifacts/service-triage/report.json`, including the full
requests/responses, independent raw oracle, observed action/read budgets and
runtime binary provenance. It uses the shared isolated generic runtime and source
fault proxy; successful telemetry responses are never fabricated. Its finally
block stops the isolated runtime and proxy. The lab's telemetry and stores remain.

```sh
GOO_GENERIC_BINARY=/tmp/goo-service-triage npx playwright test journeys/service-triage.spec.js --workers=1
```

Counts describe captured Server-kind spans, not unique requests, complete fleet
membership or verified health. Error fractions use recorded span status. A missing
current group from a complete admitted result has zero captured records, with
unavailable Error fraction/mean; source failure or inconsistent reads do not
become zero. Window deltas are descriptive independent-read comparisons.
