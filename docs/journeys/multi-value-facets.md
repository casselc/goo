# Select several counted values

Executable: [multi-value-facets.spec.js](../../journeys/multi-value-facets.spec.js).
This storyboard uses ordinary OTLP logs in an isolated Go runtime backed by real
ClickHouse telemetry. The current validation record identifies executed checks.

An operator has isolated an Activity population, inspected a field and loaded its
full-scope values. They want the records where the field is missing **or** present
empty, while retaining the original environment, service and other filters.

| Step | Interaction | Observable contract |
| --- | --- | --- |
| Inspect and facet | Inspect a retained log, then choose Facet by on `attribute:cohort`. | Existing Filters and Load receive focus; no source acquisition. |
| Load | Explicitly load this field's values. | One counted-value API acquisition/two independent SQL reads; 261 matching records versus 100 retained observations. |
| Select | Check Missing key and Present empty string. | Two local choices; no predicate or source changes. Literal `"Missing key"` remains a separate unchecked value. |
| Search | Search canary while the two choices are hidden. | Selection remains; the status reports two selected values hidden by search. |
| Edit selection | Toggle canary, clear the local selection, then reselect the two desired values. | Zero selected disables Stage; search/toggle/clear remain local and preserve source counts. |
| Stage | Choose Stage selected values once. | One OR group for this qualified field is ANDed with the retained scope. Held records remain unchanged and Apply becomes explicit. |
| Reject stale handlers | Replay detached old Stage/checkbox callbacks. | No source read and no duplicate predicate handoff. |
| Apply | Apply scope. | One new Activity acquisition; the first completed batch is paused deterministically. |
| Verify | Inspect the two returned observations. | Independent source IDs prove missing OR empty, excluding the lookalike literal and preserving all original filters. |

Counts continue describing the original loaded scope. The interface does not
invent a selected-population count or update the counts when checkboxes change.
Selection does not save or publish evidence.

The main report is `artifacts/multi-value-facets422/report.json`, version
`goo.multiValueFacets422/1`. Eleven continuous facts separate local interaction,
the explicit counted read (one acquisition/two SQL reads), and Apply (one
acquisition/one SQL read for its first completed batch). A native model validates
that observed ledger. A separate 35-second Bombadil actor runs afterward so its
shared-proxy SQL cannot interrupt the main ledger. It checks hidden selection,
exact whole-scope source facts, independent staged-DOM completion, retained IDs
and no publication.

Generated documentation frames cover selected state at 1440/390/320 pixels,
hidden selection at 390 pixels, and the applied two-record result at 390 pixels.
These are actual app captures; final acceptance and bounded review are recorded
in [validation](../../VALIDATION.md).

## Executed acceptance

The real-source journey and 35-second actor pass in 50.9 seconds. All eleven main
facts are continuous, with zero page errors. The one explicit counted lookup
uses one acquisition/two SQL reads; Apply uses one acquisition/one SQL read and
returns `record-0` and `record-1`, proving missing OR empty while excluding the
literal lookalike. Selection, search, clear, staging, screenshots and stale-handler
checks acquire no telemetry. The original 261-record counts remain unchanged.

The independent actor completes across 656 snapshots, preserving its held 100
records and observing staged OR controls before completion. It performs one
counted lookup and no publication. Five actual documentation frames are current;
the isolated runtime stopped cleanly. Native replay and independent visual
acceptance are recorded in validation.
