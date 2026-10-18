# Contextual investigation questions

**Investigate further** on an investigation's evidence card proposes another
question from that exact stored capture. It keeps competing explanations visible
without inventing a finding, a business join or a telemetry-completeness claim.

## How it works

1. Expand **Investigate further**. The Go planner reads the verified stored
   capture; it does not query telemetry. Choose among the first 100 retained
   observations when record identity matters. These choices do not define totals.
2. Select a question. Its preview names the relationship, preserved scope,
   replaced predicates/groupings and exact environment/time bounds. Selection
   alone runs no source query.
3. Execute explicitly. Record questions run in Explore; measurement questions
   open a source-free fixed preview; resource questions open instrument discovery
   with the original fixed bounds. The latter two still require an explicit read.
4. Capture useful evidence back into the investigation. Original citations stay
   fixed. Add a finding only after inspecting what the observations support.

## Current actions

| Action | Scope/relationship |
| --- | --- |
| Duration and recorded-status comparison | Same original matching span predicate, environment and fixed window. Groups by service/status class; adds complete duration/status summaries and resets inspection paging. Recorded status is not a business outcome policy. |
| Reported attribute breakdown | Same matching span predicate/window; groups by an observed `goo.operation.outcome`, `http.response.status_code` or `error.type`. Missing/empty groups remain part of the population. A reported value is not an exhaustive outcome contract. |
| Trace across services | Exact valid reported trace ID from the selected retained observation, same environment/window; replaces all original record filters and allows all services. Does not promise complete traces or follow different linked trace IDs automatically. |
| Trace/span logs | Same environment/window; replaces predicates with exact reported trace ID, optionally also span ID. Does not infer absent logs or causal relationships. |
| Resource instrument discovery | Selected record's service, or a measurement definition's declared service, and the original fixed question window. Replaces record/metric selectors with service-level metric metadata discovery. This is contextual evidence, not resource-health or causal attribution. |
| Rerun a fixed measurement | Preserves the entire typed definition, selectors and original question bounds. Fresh source evidence must be captured separately if useful. |

Invalid, zero or absent trace/span IDs do not produce ID-link actions. Inventory
and comparison captures do not infer record identities; open an underlying record
or measurement capture for supported questions. The inspected record list is
bounded and does not assert which attributes exist throughout the source.

## Missing evidence

**Record missing evidence or instrumentation** appends an authored question to the
current investigation brief with its exact origin capture ID. It neither queries
telemetry nor saves automatically. The server's existing 4,000-byte open-question
bound remains; drafts and old revisions retain the existing save/revert semantics.
Absence in a bounded instrument catalog is not proof that instrumentation is absent.

## Source ownership and validation

`GET /api/captures/{id}/followups?row=0` accepts only one canonical retained-row
index (0..99). Callers cannot supply trace IDs, services, filters or SQL. The source
adapter owns physical record fields; every generated record/measurement question
passes its existing Go compiler before being returned. Query trees are copied
independently so editing one returned action cannot change another or the capture.
The endpoint validates the stored receipt and does not return record banks.

Metric discovery normally uses a fresh two-minute window. A contextual follow-up
pins original fixed bounds, displays them, and carries them into package preview.
Editing its service/environment resets discovery to a visibly fresh two-minute
scope; it does not silently keep a different service's fixed context.

The planner is deterministic and does not infer causes or choose the user's next
hypothesis. Automatic cross-signal attribute translation, anomaly attribution,
statistical hypothesis ranking and complete-trace collection remain future work.

[Storyboard 42](journeys/README.md#42--from-a-slow-operation-to-a-supported-finding)
reproduces a real delayed sample operation, tests competing explanations, follows
exact IDs, inspects resource instruments and preserves uncertainty in the handoff.
