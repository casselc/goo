# Whole-product review task — 243

Review checkpoint, not a runtime fix. See the [six-reviewer reconciliation](../docs/reviews/INDEPENDENT_REVIEW_SUMMARY_243.md)
and [revised roadmap](../docs/ROADMAP_2026_10_29.md).

An engineer notices a request timeout in Activity, narrows to its service, switches
from logs to spans without losing the predicate, and inspects a record. Opening
the trace reveals a database-client child and a linked log across services; the
interface explicitly states that the original filters are not applied to this
supporting context. Closing the sheet and returning restores the held Activity
rows without reading telemetry again.

The engineer stages analysis of that observation window, explicitly runs it, and
captures the fixed question. This is a fresh read, not an exact bookmark of the
previous live observation. After leaving, title search finds the capture and
reopen preserves its receipt without another telemetry request. The capture saves
the scoped fixed result, not every supporting trace or log visited along the way.

![Mobile supporting trace](../docs/journeys/screenshots/243-supporting-trace-mobile.png)

At 390×844 the full-screen trace sheet covers the global Return control. Normal
Close trace → Return succeeds. At that checkpoint, closing focused `exploreRun`, and reverse
Tab from Close reaches outside the sheet. Parent-aware return/focus and keyboard
ownership need refinement; this is not a claim that all return is blocked.

![Reopened fixed evidence](../docs/journeys/screenshots/243-reopened-task-desktop.png)

The saved-evidence masthead repeats counts and custody/scope explanations before
the record table. The review proposes consolidation, preserving exact receipts.

![Discovered generic service](../docs/journeys/screenshots/243-generic-service.png)

The same owned fixture supports service discovery and an HTTP overview, followed
by Definitions and Operations walkthroughs. These are smoke observations, not
complete coverage of those surfaces or evidence of service health.

## Execution and limits

The checkpoint243 reproduction harness has been removed. Current ordinary
trace/parent interaction coverage is the [260 journey](260-trace-parent-return-storyboard.md);
capacity and exact recovery use the [246 journey](246-native-catalog-capacity-storyboard.md).
The screenshots and observations below record the original review, not a maintained
workaround or compatibility contract.

Final run passed in 17.0s (19.7s including runner setup). Core request totals were
three tail requests, two supporting context requests, one fixed analysis request
and one capture request. Inspect, return and saved reopen added none. These count
browser requests, not internal SQL queries. No product source changes, runtime
deployment, existing evidence deletion or volume resets occurred. This probe is
not a full accessibility audit or a required browser CI tier.
