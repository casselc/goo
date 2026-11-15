# 454 scoped acceptance

Independent architecture review accepted the finite consolidation: checked
population projection is separate from deliberate reference expansion. No new
controller, policy registry, cache, wire format or source owner was introduced.

- Go `PopulationQuery` and JavaScript `populationQuery` validate before stripping,
  clone complete questions and validate explicit paging/output. Nested numeric
  predicates and both relationship selectors survive; only projection directives
  are removed. JavaScript options cannot silently override signal or filters.
- Go page-only source reads retain the original question echo and original
  derived record columns. Explanation planning retains its two independent
  questions and clocks. Counted-value planning adds its categorical grouping
  only after obtaining the preserved population.
- Reference question construction explicitly carries environment/window plus
  the declared IDs or incoming selector. It does not pretend to preserve the
  original population. Invalid optional span IDs refuse instead of becoming
  trace-only lookups. Tuple/array admission is shared across both link directions.
- The shared case bank includes no relationship, edge, link and both selectors;
  environment policies; numeric predicates versus numeric projections; invalid
  originals and invalid pages. Independent review ran eight JS checks and focused
  Go tests with zero skips and found no essential semantic gap.

Existing UI structure, styling and source authority remain unchanged. Actual
incoming/outgoing journeys recheck keyboard focus, held Back, narrow containment,
paging, cancellation, refusal and stale controls. A separate real counted-value
read uses the production builder and compares results with the physical oracle.
It is not claimed as an additional modeled action in the existing journey model.
See [validation](../../VALIDATION.md) for executed receipts and limits.

The next cycle should provide a held raw-gauge dot plot, then one explicit
histogram-point display, then consolidate those views. Individual dots need no
invented instance identity or connected series. The design must retain duplicate
multiplicity, exact raw detail, recorded/unknown/invalid states and explicit
unplotted/partial-bank counts. Reuse the existing kind/data contracts and keep
local display changes source-free.
