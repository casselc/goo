# Facet an Activity field

Executable: [activity-facet-entry.spec.js](../../journeys/activity-facet-entry.spec.js).
The journey uses an isolated Go binary and real OTLP/ClickHouse telemetry. A
separate bounded span smoke checks signal preservation. Actual execution results
belong in [validation](../../VALIDATION.md).

| Step | Operator action | Contract |
| --- | --- | --- |
| Observe | Start a saved Activity scope and pause after two completed windows. | The buffer holds 100 records; the matching source population contains 261. |
| Inspect | Inspect a retained record from the earlier window. | Its original window remains distinguishable from Activity's latest completed window. |
| Choose field | Choose **Facet by…** on `attribute:cohort`. | Existing Filters and the exact field open; the existing explicit Load gains focus. No source read occurs. |
| Read values | Activate **Load values for this scope**. | One API acquisition and two aggregate SQL reads return full-scope counts and exact read receipts. |
| Browse | Search loaded values and close Filters. | Search stays local; close returns focus to the clicked field action. Reopening does not reread. |
| Choose namespace | Choose **Facet by…** on `resource:cohort`. | The distinct qualified field replaces the old context; prior loaded values withdraw. |
| Read resource values | Explicitly Load again. | Query field identity changes; selector, complete predicates and fixed window remain exact. |
| Refine | Stage the present-empty resource value. | Staging is local and retains held records. Apply remains explicit. |
| Refuse stale action | Replay the old entry/Load callbacks after scope withdrawal. | No source read or obsolete result revival occurs. |

The fixture includes a rare value absent from the 100-row buffer, missing and
present-empty values, resource/attribute keys with the same spelling, a foreign
predicate row and a current-environment-wins control. Full counts are checked
against an independent raw ClickHouse oracle. The initial rolling stream runs
before the continuous local/read ledger; the 35-second Bombadil campaign runs
after it so its independent SQL cannot appear between main facts.

The six actual documentation frames show ready-to-load and loaded values at
1440, 390 and 320 pixels. They are generated only by the running journey:
`activity-facet-entry-{ready,counted}-{1440,390,320}.png`.

## Executed acceptance

The real log journey, separate span smoke and 35-second Bombadil campaign pass in
58.7 seconds. All eleven main facts are continuous, with zero page errors. The
main page records one initial stream acquisition followed by exactly two explicit
counted-value acquisitions, each executing two SQL reads. The independent actor
runs afterward and completes across 686 snapshots, checking its full count of
261 against the raw source, 100 retained normalized record IDs, no publication and
local staging with unchanged held records. The span smoke reads all 112 matching
spans. Both isolated runtimes stop cleanly.

The journey exercises repeated inspection before entry, keyboard activation,
current/older window disclosure, resource-versus-attribute identity, hidden/inert/
ARIA and stale handlers, local close-to-opener return, and dirty-scope fallback to
Apply scope. A rare source value is absent from the 100-row buffer but appears in
full-scope counted values. Loaded values withdraw when a new contextual entry is
chosen; reopening does not promise cached results or execute a read.

The initial browser checks exposed disabled-at-construction field actions and a
4px mobile overflow in the Filters flex item. The accepted frames were regenerated
after those shared fixes. Driver corrections set an explicit raw-oracle row budget,
used literal-label matching for quoted values, and read the source adapter's
normalized `Attributes` map for retained identity checks. Native replay and
independent visual acceptance are recorded in validation.

## Current counted-value selection

The shared counted-values surface now uses local checkboxes and an explicit
**Stage selected values** action. This journey selects the one empty-string value
before staging it; the source-free entry, read windows, native return and original
read budgets are unchanged. Multi-value semantics are exercised by the
[422 storyboard](multi-value-facets.md).

The 422 regression run passes both log/span checks and the 35-second actor in
57.5 seconds. Its main eleven-fact ledger remains continuous, and the actor
completes across 683 snapshots with held observations and source-free staging
preserved. The 112-span source check and isolated-runtime cleanup also pass.
These current frames use the shared checkbox/explicit-Stage controls.

The [post-423 shared polish](../UI_POLISH_AFTER_423.md) refreshes these frames.
Exact scope and read facts remain accessible through native disclosures; local
presentation controls do not acquire telemetry.
