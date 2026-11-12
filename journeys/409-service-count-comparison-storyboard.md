# 409 · Compare captured service records

An operator sees a discovered service and asks whether its captured activity
changed from the immediately preceding window.

1. Reopen a fixed inventory capture and explicitly browse its live discovery.
   Open the exact namespace/service under its effective environment policy. The
   automatic HTTP Server dashboard contains nine current Server-span records;
   foreign namespace, conflicting environment, Client-kind and other-service
   controls do not enter that widget. A legacy-only resource does enter the
   effective-policy population. A real native gauge widget has no captured-record
   comparison action.
2. In the span widget's actions, choose **Compare captured records**. The existing
   population workspace stages a count-only presentation and adjacent equal-width
   previous window. It reads no telemetry and borrows no widget observations.
   Recorded span-status and duration projections are intentionally omitted; every
   population predicate, grouping, environment policy and page limit remains.
3. **Run** independently reads the selected and previous populations. The full
   counts are nine versus five, with a derived change of four captured records.
   Group comparison shows alpha 2→6, beta 0→3 and gamma 3→0. These are captured
   record counts, not request identity, throughput, health or statistical cause.
4. Switch the supplied comparison to Groups locally, inspect both fixed-window
   labels, and save the immutable comparison definition. **Back to service**
   restores the identical held dashboard without another source read. A callback
   retained from the departed service cannot stage a competing question.
5. Reopen the saved definition without evidence. Its selected and baseline
   questions remain exact, and another explicit **Run** reads the same fixed
   populations independently.

`409-service-count-comparison.spec.js` sends ordinary OTel data, reads the raw
persisted ClickHouse bank, and independently checks membership and group totals
for both windows. `409-service-count-comparison-actor.spec.js` drives the same
service entry and return using Bombadil at 390 px; it observes actual responses,
checks source-free staging/return, and permits no publication.

Actual frames are produced by the executable main journey:

- `docs/journeys/screenshots/409-service-comparison.png`
- `docs/journeys/screenshots/409-service-comparison-390.png`

Preparation of this storyboard is not runtime acceptance. Final execution and
scope/ownership review results belong in `VALIDATION.md`.
