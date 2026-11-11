# 406 · Captured record activity

An operator asks whether captured activity changed, then preserves the useful display without replacing the next investigation's scope. Rates describe captured records, never estimated traffic or a percent of the page.

1. Run a fixed, compound-filtered log question. The regular summary stays count-first. Open Population context → Customize summary; the child remains staged until Run. Select Captured records per second, move it ahead of Count, and see an exact rational backed by the full result count and query window. Four matching records over 3.5 seconds produce 8/7 records/s even though the result page holds two records.
2. Save the display as a reusable ordinary summary template. Back returns to the identical held distribution. Reopening the saved definition does not load evidence. Inspect and apply it to the destination question: its environment policy, predicates, time window, page limit and root position remain intact. Only Run reads five matching destination records over three seconds, producing 5/3 records/s.
3. Open the fixed span timeline. Count remains the default. Select captured record rate locally: bucket counts [2, 0, 1, 1] over widths [1000, 1000, 1000, 500] ms produce rates [2, 0, 1, 2]. The final half-second bucket has the same count as the preceding bucket and twice its rate. Zero is observed zero; an unreported interval would be unknown, not an invented zero.
4. Switch bars/table, save and reopen the rate view, then Run explicitly. Select the clipped final interval and stage it. Run the narrower 500 ms question: one captured record is still 2 records/s. Retired controls cannot replace the current owner's scope or issue reads.
5. Run actual empty and known-zero interval questions. They retain exact exposure and derived zero. Missing facts, malformed/large integer facts and unsupported grouped cases are checked by the contract/property suites rather than invented source-positive fixtures.

The source bank sends real OTel logs and spans to the collector and independently reads their raw ClickHouse rows. It includes foreign environment/namespace and excluded predicate controls, plus a legacy-only environment row that is admitted only by the destination's effective policy. No positive source response is mocked.

Screenshots produced by the executable journey:

- `docs/journeys/screenshots/406-summary-rate.png`: destination log summary.
- `docs/journeys/screenshots/406-timeline-rate.png`: span rate timeline.
- `docs/journeys/screenshots/406-timeline-rate-390.png`: the same timeline at 390 px.

`406-record-rate.spec.js` records the exact queries, raw rows, source responses, snapshots, field order, held-parent identity and per-action read/write/expansion budgets in `artifacts/record-rate406/summary.json`. `406-record-rate-actor.spec.js` drives the same template-to-destination workflow with Bombadil at 390 px using `verification/bombadil/record-rate406.ts`; it permits one expansion and one explicit source read, no publication, and no evidence during staging. Execution results belong in the validation record after the serial checks finish.
