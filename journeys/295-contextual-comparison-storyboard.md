# Compare captured evidence beside the applied chart — 295

Executable: [295-contextual-comparison.spec.js](295-contextual-comparison.spec.js).
The isolated generic Go runtime receives real OTLP spans; no saved view, capture
or comparison is seeded through the API. Setup ingestion/readiness polling is
separate from browser request accounting.

1. Start Activity, narrow to the API service, group observed `http.route`, analyze
   the fixed window and Customize spans. The current root scope remains exact.
2. Open **Compare with captured baseline** beside the aggregate using Enter,
   close it with Space, then open it again. These native gestures are passive. The planner's
   status measure is explicitly unsupported by the ordinary-count profile. Opening
   and **Edit span measures** read nothing and do not change the question. Turn
   status off explicitly, then Run and capture a matching baseline.
   [Explicit profile choice](../docs/journeys/screenshots/295-explicit-profile.png).
3. Read the current population after a delayed arrival. Open the same disclosure,
   search and explicitly Load saved captures. A delayed response for an older
   title search cannot populate the new search. Choosing the exact capture stages
   configuration only; **Prepare baseline** explicitly loads verified stored
   evidence and performs the pure comparison.
   [One contextual picker](../docs/journeys/screenshots/295-contextual-picker.png).
4. Compare Groups: `/checkout` changes from 1 to 2 and `/health` remains 1. Open
   the chosen current-side question without reading, then Back. The prepared pair
   stays held; old picker controls refuse and fresh controls retain the exact
   reference and open disclosure. [Aligned groups](../docs/journeys/screenshots/295-compared-groups.png).
5. Hide the aggregate: the sole picker moves beside the visible comparison. Hide
   both: it falls back to Composition settings. Setting the same reference while
   hidden leaves the comparison hidden and reads nothing. Showing it also reads
   nothing; a separate Prepare reacquires the stored baseline and pure comparison.
   These changes never duplicate the picker.
6. Save the whole comparison definition and reopen. Its staged settings contain
   one picker and the exact baseline reference, with no automatic acquisition.
   Explicit Run reads the current question and prepares the visible comparison.
   [Reopened comparison](../docs/journeys/screenshots/295-reopened-comparison.png).

The independent Bombadil campaign runs the saved definition, removes its configured
comparison, then uses the sole picker to choose the actual capture and explicitly
prepare again. Completion requires observed canonical baseline/current counts of
2 and 3, not a click. Invariants limit source reads and writes and forbid duplicate
pickers. Existing290/291 observed Quint and comparison regressions exercise the
unchanged comparison/profile semantics.

Validation: the final journey and independent 30-second Bombadil campaign pass
(about one minute combined). The final 295/294 regression run passes in 1.1
minutes, including the hidden-reference case on final styling. The browser makes two tail subscriptions, seven
Explore POSTs, one capture, three pure comparisons and one view save; no page or
document POSTs. This counts endpoint operations, not physical SQL reads inside
subscriptions/captures. Fixture readiness polling is excluded. The independent
campaign observes 549 snapshots across seven phases, one Explore request, two
comparisons, zero writes and a canonical 2-versus-3 prepared pair. All four frames
are generated from the running app. Native keyboard disclosure is exercised;
screen readers, mobile layout and production load are not assessed.

Regression 290 passes with observed Quint and a 30-second Bombadil campaign
(489 snapshots across six phases). Regression 291 passes its observed Quint replay.
Regression 294 initially reached normal Activity DOM but its initial cold navigation
missed the 15-second load deadline; the rerun uses 45 seconds for navigation only,
while the same interaction assertions retain 15-second deadlines. No product
behavior was changed to satisfy that navigation harness timing.
