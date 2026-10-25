# 194 — Service summaries without record-page reads

Status: source `4a49f61` is deployed. The live main, late-owner case, bounded actor and paired source audit pass; the independent raw-ledger checks also pass.

A discovered service still opens its ordinary HTTP/log/native views. Record widgets now read `/api/record-aggregates` and receive the full-population count, groups, requested span duration/status and overall time series with independent read receipts. They do not fetch record rows, discovered fields, source plans or result tokens. The already-completed supporting record page remains available for profile admission and inspection. Native measurements keep their own endpoint and meaning.

## Actual user journey

1. Discover `commerce / edge-api`. Its held supporting page still contains 100 actual spans and a same-row HTTP server witness. The automatic HTTP profile counts 137 matching server records; its log widget counts 155 separate log records.
2. Read route diagnosis, duration/status summaries and time bins. Missing and empty routes remain separate. The generated fixture includes 120 varying durations from 1–120ms and 12 custom durations from 5–60ms, making exact percentile boundaries observable. Independent assertions cover every group and time bin, not just a displayed total.
3. Open the log widget's original fixed question without a read. Run explicitly loads 155 logs, 100 displayed rows and discovered fields through the existing full Explore endpoint. Back restores the retained service overview without rereading it.
4. Capture the span widget. Opening the dialog is local; confirmation performs the documented fresh full read. The stored receipt retains 137 inspection records, with 100 on its first public page. Reopening it adds no telemetry read or format migration.
5. Navigate to Services using the existing sidebar. This preserves the held overview without a source read. Capture navigation uses its existing history contract; this journey does not invent a browser-Back return from capture.
6. Open the `/items` diagnosis population locally, Run 120 spans, inspect one reported trace through the existing two fresh span/log reads, then return to the retained service. Matching trace context does not establish complete tracing or causality.
7. Apply a one-millisecond range whose server record is absent from the held 100-row page. Local editing reads nothing; Apply explicitly rereads the overview. The accepted HTTP profile remains selected even when the old supporting page provides no positive witness in that narrow interval. Its full count is 1.
8. In a separate owner case, hold a real all-span aggregate response, leave the service, then release it. It cannot reopen the departed owner. A subsequent explicit service open reads fresh HTTP evidence.

The fixture is generated ordinary OTLP telemetry: 225 spans across roles/services/namespaces, 160 logs (155 in scope and 5 in a different resource namespace) and the existing native instruments. There are no mocked positive source/UI responses and no goo manifest. Full raw facts are retained beside the actual API responses.

## Actual frames

- [194a — Rich service summary and diagnosis](../docs/journeys/screenshots/194a-rich-service-aggregates.png).
- [194b — The same overview on mobile](../docs/journeys/screenshots/194b-mobile-service-aggregates.png).
- [194c — Desktop product context](../docs/journeys/screenshots/194c-service-overview-in-context.png).

All three are unedited frames from the accepted main's same fixture and running app. Shared mean/p95 tracks have equal physical widths, and counts, quantiles, statuses and time bins match the raw population oracle. Existing catalog/header overhead and long mobile dashboard scrolling are inherited; this read-projection slice does not claim to solve whole-product density.

## Accepted receipts

- Main **19.9 seconds**, within the third 51.9-second two-test invocation: `artifacts/service-dashboard-projection/service-projection-b7b0685a04/ledger.json`, log `third.log`. Initial overview has ten logical source operations and one pure plan: three rich record widgets, two unchanged native reads and existing discovery/support reads. The full main has 20 source operations and two pure plans, including drilldown, capture, trace and explicit range Apply. These are browser operation counts, not ClickHouse statement counts.
- Full capture ID `0afbdb9ef581afedbe71ebe4bd5c22725b87ee808196e82f7fb0d83200caee7e`, **138,964 bytes**, SHA-256 `9837ae1af29939e979f715421ac830077674ae05b4ee551bcbd3d7ba011256c2`. `capture-before.json` and `capture-after.json` are exact stored files read from the existing capture store; their hashes match. The rich aggregate response itself has no capture token and is not stored as immutable evidence.
- Late owner **8.0 seconds**, 10.2-second invocation: `service-projection-fbc939b42b/ownership.json`, log `owner-final.log`. It retains the actual held DTO, departed visibility, current request identities, cancellation control and fresh accepted result. Controls are not telemetry rereads.
- Bombadil **31.3 seconds**, 32.8-second invocation: `service-projection-133c4d4bff/bombadil/acceptance-summary.json`, log `actor-first.log`. All nine phases and 462 snapshots prove three rich widgets, a full 100-row admission witness, local diagnosis sorting, source-free widget staging, a full log 155 read and retained Back. Source budget is 11, including one explicit full drilldown; no additional widget reads occur on local interactions.

Independent native gates pass with eight main and four owner controls rejected: combined 1.008 seconds, with a final stricter owner-only check 0.218 seconds. Raw full-population groups, exact-high quantiles, statuses, complete time bins, full admission/drilldown rows, capture bytes and receipt identity are independently checked. The initial native observer bound the held all-span 198 result to the original HTTP 137 question; it was corrected to the actual all-span plan in the same original scope, while the fresh HTTP result remains strict. No application or fixture changed; the observer diagnostic is retained. See the adjacent native-validation logs.

Setup ingestion/readiness and native-instrument discovery probes are separately recorded. Aggregate clocks/bases describe independent source observations; they do not establish a transactional snapshot, sampling completeness, request health or equality with a later fresh capture read.

## Paired source footprints

The [audit comparison](../docs/evidence/service-aggregate-audit/comparison.json) uses the exact archived journey 187/190 widget questions, with their existing telemetry and traffic stopped. Each case matches 137 records and four groups. Count/group/time-series SQL and body sizes are unchanged where retained, and canonical aggregate fingerprints match.

| Fixed widget | Full Explore reads | Rich reads | Full result bytes | Rich result bytes |
| --- | ---: | ---: | ---: | ---: |
| HTTP main | 5 | 3 | 103,936 | 5,456 |
| HTTP route diagnosis | 4 | 2 | 99,954 | 2,546 |
| Messaging main | 5 | 3 | 117,788 | 5,376 |
| Messaging destination diagnosis | 4 | 2 | 113,926 | 2,570 |

This removes record-page and field-discovery work from those widgets. It does not claim the remaining aggregate SQL became cheaper, or generalize single controlled observations into a latency benchmark. Native reads and the supporting admission page are unchanged. The separate legacy-equivalence receipt verifies full `Source.Fetch` output against archived source, normalizing only its observation clock.

## Diagnostics and compatibility

Initial main attempts exposed harness assumptions: browser Back after capture's existing replace-state navigation (**35.3 seconds**) and an unquoted exact locator for the JSON-quoted route label (**33.1 seconds**). Full aggregate and capture assertions had passed; only navigation/locator expectations changed. The accepted main was not repeated after passing.

The owner collector initially awaited bodies from canceled client responses. It now binds cancellation to actual request failure or matching query-ID control, retaining true header status and explicit unavailable-body metadata. A 15-second diagnostic guard identified the remaining pending entries as keepalive cancellation acknowledgments, not data responses. Those controls now record status/request with `control: cancel` and `bodyObserved: false`; no positive body is fabricated. Complete data responses remain strict. Initial 120-second waits and the 28.5-second diagnostic run are retained under the three failure directories; application code and owner guards were not relaxed.

Future 187–190 helpers/specs now expect rich widget transport and no widget rows, while preserving their separate 100-row supporting-page and full drilldown checks. Native verification adds an explicit narrow-response branch without synthesizing rows. Archived receipts and historical frames are untouched. Those live campaigns were **updated, not rerun** in this round; their existing archived 187 native regression was checked separately.

## Reproduction and runtime

```sh
npx playwright test journeys/service-dashboard-projection.spec.js
GOO_SERVICE_PROJECTION_BOMBADIL=1 npx playwright test journeys/service-dashboard-projection-bombadil.spec.js
```

`artifacts/service-dashboard-projection/deployment.json` records actual runtime binary SHA-256 `6af45f128aacd5540cbf25942cbd82acdd37750ef2904891ba5ac66eb167e11c`, server/static/spec hashes and correct MIME. Only packages/query were refreshed; no data volumes were changed. Traffic was stopped for the baseline, paired final audit and browser work, then restored to its prior running state. The temporary binary was removed. Normal diagnostic retention was planned/applied afterward, preserving compact ledgers, capture bytes, current failures and documentation frames.
