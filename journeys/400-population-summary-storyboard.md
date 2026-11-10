# Keep a few useful population facts

The [actual-source journey](400-population-summary.spec.js) starts with ordinary
OpenTelemetry logs and spans, without a manifest or a categorical grouping.
Execution results belong in [validation](../VALIDATION.md); the frames below are
captured by the app, not mocked.

1. Run a fixed ordinary numeric-log, numeric-span, or span-duration/status
   question. Its nested predicates, environment policy, exact bounds and
   nondefault three-record page remain intact. Persisted raw attributes,
   timestamps, durations and reported statuses establish independent aggregate
   and support oracles.
2. Choose **Customize summary**. The existing population workspace stages the
   exact question plus Summary and Records. Staging reads no telemetry and shows
   no evidence. Only **Run population view** acquires that population's facts.
3. In the folded **Population summary** editor, choose a few supplied facts and
   move one earlier. Numeric sum can precede count and mean; recorded error
   fraction can precede count and mean duration. The visible order matches the
   authored order across field families. The view contains one summary, not an
   additional automatic copy. Local choices perform no source reads.
4. Save the immutable configuration. Back restores the original accepted source
   summary and basis without rereading. Reopening the saved view stages its
   configuration without evidence. Inspect and Apply over a different compatible
   destination retain its full question; explicit Run reads its own values.
5. Read the same compact summary at 390px. Units, exact-value/support disclosures
   and unavailable state remain accessible. Additional raw-source checks distinguish
   no records, no valid numeric values, and a genuine numeric zero.
6. Hidden, inert, superseded and pending-source Customize callbacks cannot stage
   through retired ownership. A separate Bombadil campaign independently selects
   and reorders numeric facts, then returns to the exact held question.

Prepared frame paths: [numeric summary](../docs/journeys/screenshots/400-authored-numeric-summary.png),
[reused duration/status summary](../docs/journeys/screenshots/400-reused-duration-summary.png),
and [390px view](../docs/journeys/screenshots/400-summary-390.png).
Their existence and review are acceptance results, not implied by this storyboard.


Reviewed actual frames: [authored numeric facts](../docs/journeys/screenshots/400-authored-numeric-summary.png),
[reused duration/status facts](../docs/journeys/screenshots/400-reused-duration-summary.png),
and [390px view](../docs/journeys/screenshots/400-summary-390.png).
Measured outcomes and limits are in [validation](../VALIDATION.md).
