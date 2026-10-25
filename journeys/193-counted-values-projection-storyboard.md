# 193 — Count full-scope field values without loading record projections

Status: source `eae70ab` is deployed; the real browser main, bounded actor and controlled before/after read audit pass. The independent raw-ledger check also passes.

The engineer opens an applied fixed question, inspects an observed field, and explicitly loads its counted values. Opening fields, searching supplied values and staging a filter stay local. The dedicated `/api/explore/counted-values` result contains only fixed scope, total count, groups, independent read clocks/bases and aggregate observation; it does not load a record page, discovered fields, latency or time charts.

The generated ordinary OTLP log fixture is synthetic telemetry, not a mocked source response. The large admitted population has 143 records: 140 common values, one rare value, one present empty string and one missing key. A small admitted population has 5 records. Four controls have a below-threshold, invalid, empty or missing numeric value. Both counted queries retain the exact signed integer threshold `9007199254740993`, service, resource namespace and original fixed window.

## Executed user journey

1. Run the fixed log question. Its first 100 records are not the source for full counted values.
2. Open an observed field without a source read; explicitly load values through the new endpoint. Verify 143 total and all four groups, with missing and empty distinct.
3. Search loaded values locally, stage missing, then Run explicitly to read the single selected record. The original numeric predicate remains in place.
4. Execute a separately labeled small-scope counted request and verify 5 records/four groups.
5. Hold a real counted response, edit the environment and release the response. The outdated values cannot publish into the changed owner.
6. Run a bounded actor checking the new endpoint, no legacy limit1 Explore group request, source-free staging and explicit selected Run.

## Reproduction

```sh
npx playwright test journeys/counted-values-projection.spec.js
GOO_COUNTED_PROJECTION_BOMBADIL=1 npx playwright test journeys/counted-values-projection-bombadil.spec.js
```

Existing journeys 151/152 and their actor retain their prior semantic assertions but use the dedicated endpoint. The controlled ClickHouse before/after read audit runs separately, with its own fixed scope/settings and paused lab traffic; browser logical POST counts are not SQL read counts.

## Actual frames and receipts

- [193a — Full-population counted values](../docs/journeys/screenshots/193a-counted-values-full-population.png): 143 matching records, including values absent from the first 100-record page. Missing and present empty are separate rows.
- [193b — Mobile counted values](../docs/journeys/screenshots/193b-mobile-counted-values.png): the same accepted main, fixed scope and supplied counts; viewport changes and local search perform no reads.

Both frames are unedited screenshots of the actual running UI over generated ordinary OTLP data. They do not replace historical 151/152 screenshots.

Main **21.7 seconds**, 25.8-second invocation: `artifacts/counted-values-projection/counted-projection-ab82d8e588/ledger.json`, log `main-third.log`. The browser performs five explicit source operations: initial fixed records, counted values, selected records, reapplying original records, and the held counted lookup. A separate cancellation-control POST is not a telemetry reread. The small-population API probe and setup ingestion/readiness are separately labeled request-context operations.

The narrow response contains only version, question, count/state, groups, aggregate observation basis/clock and independent count/group read receipts. It contains no record page, fields, duration or time-series projection. Both read clocks and bases are retained, and the independent reads do not establish a transactional snapshot or source completeness. No immutable capture is created by this endpoint.

Actor **31.6 seconds**, 32.8-second invocation: `counted-projection-bad4370743/bombadil/acceptance-summary.json`, log `actor-second.log`. All seven phases and **511 snapshots** prove one dedicated lookup, zero legacy limit-one Explore group reads, local search/staging and an explicit selected Run with the original exact numeric predicate. Its total is three source POSTs.

The actual held response is fetched from the live endpoint, then delayed at transport. Editing the environment clears its ownership; releasing that response cannot republish the outdated values. This is a negative transport test, not mocked positive telemetry.

Independent raw-ledger verification: **first pass 0.177 seconds** (0.13-second test body), eight guaranteed changed controls rejected. `artifacts/counted-values-projection/native-validation.txt` checks complete raw 143/5-record memberships and groups, null versus empty, the original 100-row page, selected missing 1, exact numeric threshold, receipt identity, stale ownership and five actual browser reads. Cancellation is not counted as a read. No additional model or large campaign was introduced.

## Controlled source-read comparison

The separate [audit comparison](../docs/evidence/aggregate-read-audit/comparison.json) uses existing journey 192 telemetry, with traffic stopped and unchanged source settings. Its small/larger labels describe **group cardinality**, not the 5/143-record UI fixtures above. Each case contains 123 matching records.

| Supplied groups | Existing Explore reads | Counted-values reads | Existing result bytes | Counted result bytes |
| --- | ---: | ---: | ---: | ---: |
| 4 | 4 | 2 | 5,819 | 766 |
| 123 | 4 | 2 | 9,472 | 4,508 |

Each pair retains identical count/group SQL hashes and body sizes. The dedicated path removes record-page and field-discovery reads; it does not claim to reduce the work of the remaining count/group SQL. The corresponding observed scan rows fall from 24,708 to 12,354 with 6,177 per read in these cases. Four executions are single controlled observations, not a latency benchmark or a guarantee for every data size/cache state. Exact source footprints/settings and response receipts are linked from the comparison directory.

## Diagnostics, compatibility and runtime

Two setup attempts were correctly refused before browser navigation: missing explicit numeric parse policy (**1.7 seconds**) and then missing unit (**1.6 seconds**). The fixture now supplies `parse: canonical-decimal` and `unit: {item}`; pure Go and UI validation passed before the accepted main. `first-failure-diagnostics/`, `second-failure-diagnostics/` and `fixture-validation.log` retain evidence. These are harness corrections, not application defects or relaxed numeric semantics.

The initial actor (**31.6 seconds**, failed liveness) read the hidden Activity editor through an unscoped counted-values selector and waited at phase 2. The corrected actor scopes controls/state to the fixed editor and reads the standard count component's numeric span. Its original phase trace and failure diagnostics remain retained; source budgets and liveness bounds were not increased.

Existing 151/152 journey and actor expectations now point at the dedicated endpoint and retain their semantic assertions, including high-cardinality refusal. They were **updated but not executed in this round**; their historical frames are unchanged. The bounded source/HTTP gates cover the 1,001-group refusal without adding another live campaign.

`artifacts/counted-values-projection/deployment.json` records actual runtime/static hashes, MIME and lab/configured capabilities. Binary SHA-256 is `7c68976da9112b187168aa21c6e569c3479c67ddd68190247c45e81a1040b53d`. Only packages/query were refreshed; no data volumes were changed. The owned traffic generator was stopped for the controlled audit and browser work, then restored to its prior running state. The temporary binary was removed. Normal diagnostic retention was planned/applied afterward; current failures, compact ledgers, actor summaries and documentation frames remain retained.
