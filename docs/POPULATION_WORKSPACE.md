# General population exploration and saved views

Start with any applied span or log question in Explore, including cross-service
resource/attribute predicates. **Explore linked population** asks the Go planner
to compose existing record kernels, then reads one or two signal questions. The
workspace shows aligned charts and matching inspection pages together. Clicking a
chart interval or adjusting UTC handles previews a range without a read. Apply
rereads both configured populations in that range; reset restores the workspace's
original bounds. Record paging changes one inspection page without changing its
full-population totals. Layout and column choices are local presentation changes.
The existing record inspector, trace expansion and surrounding-log actions are
reused; there is no new JavaScript measurement evaluator.

The default companion retains the complete predicate tree, including AND/OR,
missing-key and present-empty-string distinctions, where that tree is valid for
the other signal. Signal-specific measures are planned independently. A span-only
or log-only predicate (or reported dependency edge) makes the default companion
unavailable; predicates are never silently dropped to manufacture a broader
population. That reason is visible. **Configure a separate companion population**
lets the user explicitly replace only the companion's selector with an exact
service and/or attribute filter. Both questions must retain the same environment
and half-open window. Each signal's exact selector, source basis and failure state
remain inspectable. Temporal alignment and shared predicates do not establish
trace identity, a causal relationship or complete source coverage.

**Save this exploration view** stores configuration: a bounded title, one or two
fixed queries, a charts-and-records or records-only layout, and visible record
columns. UTC and summary remain required. The Go workbench store publishes a
content-addressed immutable artifact; saving the same configuration returns the
same ID, while changed configuration has a different ID. Saved views contain no
telemetry, captured result basis or inferred health. The catalog is paginated,
and direct `#exploration-view=<id>` links reopen configuration without source
reads. **Run population view** is explicit. **Use fresh population window** stages
a fresh window of the same duration without a read; a saved revision remains
unchanged. Existing telemetry packages continue to provide declarative reusable
measurement definitions; saved exploration views are presentation/configuration
artifacts referencing the same typed query grammar.

Independent signal reads publish together once they settle, retaining each
successful result and showing failures as unavailable. Cancellation, source-scope
edits, workspace suspension and newer requests prevent stale responses from
publishing into another population. Capture evidence through the existing
per-question/capture workflow when results themselves must be retained.

```mermaid
flowchart LR
  Explore[Applied typed question] --> Plan[Go population planner]
  Plan --> Primary[Original selectors + signal measures]
  Plan --> Companion[Identical portable selectors or explicit companion]
  Companion --> Withhold[Withhold incompatible selectors; never drop predicates]
  Primary --> Reads[Existing Go source adapters]
  Companion --> Reads
  Reads --> View[Aligned timeline + independent record pages]
  View --> Range[Range preview: no read]
  Range --> Reads
  View --> Config[Layout + columns + fixed queries]
  Config --> Store[Immutable workbench view]
  Store --> Stage[Reopen configuration only]
  Stage --> Run[Explicit Run / stage fresh bounds]
  Run --> Reads
```

[Journey 90](../journeys/population-workspace.spec.js) emits generic OTel spans and
logs from multiple services, including populations excluded by resource or
attribute predicates. It exercises missing/empty semantics, range selection,
independent companion filters, presentation changes, save/reopen/run and mobile
inspection. Its observed ledger is checked with Quint/quint-go and rejecting
controls. The Bombadil actor opens a saved view, explicitly runs it and changes
layout. Hegel generates portable-selector properties and state/swarm sequences
for immutable saved-view layout, column and range edits.

![Cross-service populations](journeys/screenshots/90a-cross-service-population.png)
![Linked records and chart range](journeys/screenshots/90b-linked-records-and-charts.png)
![Saved presentation configuration](journeys/screenshots/90c-saved-population-view.png)
![Mobile inspection](journeys/screenshots/90d-mobile-population-view.png)
