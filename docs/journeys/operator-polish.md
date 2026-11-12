# Operator workflow polish

This storyboard follows an operator from current observations to fixed evidence.
The services are an edge API, queue worker and catalog database; no application
manifest or reservation-specific knowledge is required.

| Scene | Operator action | Product expectation |
| --- | --- | --- |
| Activity | Start the scoped log stream, then pause. | Time, service, severity and a useful message preview appear before secondary investigation tools. Thirty ordinary observations exercise a real populated table. |
| Inspect | Open a row, read its complete message and exact timestamp, search attributes, then press Escape. | Reading supplied details does not query the source. Escape closes the inspector and returns focus to the originating row action. |
| Fixed records | Analyze the observation window, explicitly Run, then open the persisted fixed span question. | Staging remains separate from reading. A compact record list retains the exact original IDs, UTC times, durations and reported statuses. |
| Services | Discover services in the fixed UTC window. | All three services are available from ordinary OTel telemetry. Inventory presentation does not change the selected window. |
| Saved evidence | Explicitly capture the fixed question; open and reload the capture. | Capture performs its own read. Reopening reads stored evidence, retains the captured scope and performs no source query. |

The runnable storyboard is `journeys/operator-polish.spec.js`. It records actual
1440 px and 390 px browser frames and geometry to `artifacts/operator-polish/`.
The files are named `{before,after}-{activity,fixed,services,saved}-{1440,390}.png`;
`before.json` and `after.json` contain the corresponding measurements and checks.
`{before,after}-fixed-inspect-1440.png` shows the desktop inspector alongside its
record context when that version of the journey is run.
Only completed browser runs produce those artifacts; this document does not
substitute generated mockups for product screenshots.

Run with `GOO_POLISH_PASS=before GOO_GENERIC_BINARY=/tmp/goo-polish-before`,
then `GOO_POLISH_PASS=after GOO_GENERIC_BINARY=/tmp/goo-polish-reviewed`, each
followed by `npx playwright test journeys/operator-polish.spec.js`.
Keep `artifacts/operator-polish/bank.json` between passes. The fixed question,
Records, Services and Saved evidence reuse that same bank; only the before pass
ingests it. Activity has a server-clock rolling 60-second window, so it uses a
separate fresh real OTel bank in each pass. Its source values are not represented
as identical across the two screenshots.

After-pass acceptance requires the first fixed row within 350 px on desktop and
430 px on mobile, desktop rows at most 44 px tall with at least twelve rows in
view, and no horizontal page overflow. Attribute search, closing inspection and
Saved reopening retain their local operation budgets. These are bounded product
checks, not a replacement for query semantics or source ownership tests.

## Actual result

The same fixed bank passed the before/after workflow on November 12, 2026.
At 1440px fixed results move from **500px to 330px** before the first row,
with ordinary rows **123.5px → 42.5px** and **four → fifteen** fully visible.
Saved evidence shows **fourteen** rows instead of three. Mobile fixed records
start at **421px instead of 550px**. Services' first row moves from **773px to
370px** on desktop and **1021px to 482px** on mobile. No page overflows.

Mobile Activity deliberately gives messages full width before their actions;
its long-message rows are taller than the old squeezed preview. This is a
readability improvement, not a claim that every surface displays more rows.
Activity follows the newest observations in its bounded scroll viewport, so
the first DOM row can be above the viewport; that negative coordinate is not
a first-data entry measurement.

| Scene | Actual frame |
| --- | --- |
| Activity | [Desktop](screenshots/operator-activity-1440.png) · [Mobile](screenshots/operator-activity-390.png) |
| Fixed records | [Desktop](screenshots/operator-fixed-1440.png) · [Mobile](screenshots/operator-fixed-390.png) · [Before](screenshots/operator-before-fixed-1440.png) |
| Contextual inspection | [Desktop](screenshots/operator-fixed-inspect-1440.png) · [Mobile](screenshots/operator-activity-inspect-390.png) |
| Services | [Mobile](screenshots/operator-services-390.png) |
| Saved evidence | [Desktop](screenshots/operator-saved-1440.png) |

Exact IDs, timestamps, durations, statuses and service membership are checked
against the supplied bank, with identical fixed records before and after.
Inspect/search/Close, fixed inspection and Saved reload each retain their
API POST and SQL budgets. Capture is an explicit independent source read.
The separate structured-inspection journey checks unsafe/literal text, exact
clipboard values, field-menu Plot/Cancel focus, actual grouping from a collapsed
editor, trace context and mobile keyboard return; its Bombadil campaign must
complete both waterfall selection and field-menu open/Escape phases.
