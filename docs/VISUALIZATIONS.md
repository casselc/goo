# Investigation visualizations

The local investigation UI now supports a trace waterfall and a gRPC latency
heatmap over actual OTel exporter spans. Open http://localhost:8088.
These are presentation and raw-span query additions to the reference lab; they
do not migrate the API to Go/Rust or connect the separate Go lifecycle kernel.

## Trace waterfall

Choose **Spans and links**, or open a reservation from Completion SLO.

- Each trace is a separate group, with parent-first nesting within that trace.
- All groups share one time axis, relative to the earliest returned span.
- Exact epoch nanoseconds travel as decimal strings. Geometry subtracts `BigInt`
  timestamps before conversion to floating-point screen coordinates.
- Select a span to inspect status, raw attributes, IDs, and link targets.
  **Follow async link** opens the referenced span when it is in the returned set;
  otherwise the UI explicitly labels the external reference. New sample producer
  spans carry reservation attributes so the real async target is discoverable;
  older telemetry is not backfilled.
- An async span link never creates a parent-child edge. Missing parents and
  malformed cycles are labelled; traversal remains bounded.
- The query returns at most 200 spans. The chart says when that limit is reached
  and always states that completeness is unknown. The displayed extent is not a
  claim about the complete trace duration or critical path.

## Raw-duration heatmap

Choose **Latency heatmap**. Each cell counts `lab-grpc` server spans starting
in a UTC minute and duration band: 0–10, 10–50, 50–100, 100–500 ms,
0.5–1, 1–10, or ≥10 seconds. Bands are half-open, including the first zero-duration
band; the overflow band covers the full UInt64 duration domain.

Click a populated cell to inspect its raw spans. The cell's exact clipped time
range and duration bounds are preserved as typed `spanFilter` parameters. Time
and duration upper endpoints are exclusive. That selection does not acquire a
moving lookback when executed later. It can still gain late-arriving telemetry:
this prototype has no immutable custody basis or historical query snapshot.
Select a span, then **Open reservation waterfall** to widen to its reservation
and follow its asynchronous links. Clear selection removes the cell filter.

Counts aggregate the entire matching span population, not the 200-row drilldown.
The 120-minute maximum yields at most 847 nonempty cells (121 intersecting
minutes × 7 bands), below the service's 1001-row safety cap. `limit` is null for
the heatmap; the inspection query retains its 200-span limit. The exact fixture
checks 1101 calls contribute to the heatmap while only 200 appear in inspection.
These are raw-span durations, not histogram metric points or aggregated p95s.
A blank cell means no observed spans, not proof of no traffic or good health.

## Reuse and boundaries

Inspected related work:

- `observability/otel-chdb/lakeui/src/charts.js`: adapted its dependency-free
  relative-time waterfall approach. The new renderer adds scoped parent nesting,
  exact timestamp transport, selected-span details, and explicit async links.
  Its evidence-completeness presentation informed the permanent unknown label;
  goo cannot copy its settled labels without the corresponding evidence protocol.
- `oscope/src/oscope/visualization/document.clj`: renderer-neutral documents and
  explicit query-backed sources inform separation of pure geometry, query
  selection and DOM rendering. No Jolt/Clojure runtime is imported.

`ui/visualizations.js` contains geometry/rendering; `grammar.py` owns the fixed
exporter-column adapter and validated SQL. The heatmap is a bounded prototype
view, not yet a declarative Go visualization grammar. Follow-on work should
introduce portable visualization/selection descriptors over typed Go result
schemas, with the same UI consuming a Go query service. DataFusion is not needed
for this slice.

Still absent: zoom/drag brushing, trace fetch by global ID, paginated full traces,
critical path calculation, durable selection history, cohort comparisons, and
historical basis selection. Reservation drilldown remains scoped to this demo's
fixed tenant, not a generalized cross-tenant relation.

## Verification

```sh
/workspace/.venvs/goo/bin/python -m unittest discover -s tests -v
node --test tests/visualizations.test.js
npm run test:journeys
```

Embedded ClickHouse checks cover all duration boundaries, zero, UInt64 maximum,
exclusive time endpoints, display-limit independence and typed selection
validation. Node tests cover nanosecond geometry, cross-trace parent isolation,
cycles, missing parents and zero-duration/empty results. The real browser journey
creates a retry through the sample page, selects a heatmap cell, checks its raw
span count, opens the full reservation, and follows a real async link. See the
[storyboard and captured frames](journeys/README.md).

No lifecycle accounting semantics changed in this slice. Existing Hegel/Quint
lifecycle tests remain separate; this renderer does not claim to establish their
properties or evidence completeness.
