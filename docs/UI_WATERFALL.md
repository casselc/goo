# Passive trace waterfall

`ui/waterfall.js` defines `goo-trace-waterfall`. Load its independent stylesheet
`waterfall.css`; no application runtime or query client is needed. It is a display
of supplied span records, not a trace retrieval API, completeness evaluator,
critical-path calculator or unique-invocation counter.

```html
<goo-trace-waterfall
  data='{"rows":[{"timestamp_ns":"1700000000000000000","duration_ns":"1000000","TraceId":"a","SpanId":"b","ServiceName":"api","SpanName":"GET /items"}],"truncated":false}'
  selected-index="0">
</goo-trace-waterfall>
<script src="/waterfall.js"></script>
```

`data` / `.data` accepts `{rows, truncated}`. Attribute JSON is bounded to 64 KiB;
properties accept up to 1,000 snapshot rows and detach values on assignment and
readback. Component input requires exact nonnegative nanosecond strings within
UInt64 bounds. Invalid values withhold the plot instead of fabricating zero.
`truncated` is an explicit caller boolean. It does not imply completeness when
false. Empty supplied rows have an explicit empty state.

`selected-index` / `.selectedIndex` identifies the original supplied array index;
it does not use SpanId as physical record identity. It is controlled by the caller.
Clicking a native row button alone does not change selection, open a detail view,
emit a query event or fetch telemetry. An app may delegate native clicks and use
`button.dataset.rowIndex` to resolve its own record. Updating selection adjusts
`aria-pressed` without rebuilding rows or losing focus. Data updates/reconnects
rerender the snapshot and reset DOM/disclosures; this is not a live keyed renderer.
Pre-definition property assignments are upgraded when the element connects.

The layout subtracts exact nanosecond timestamps before converting relative
geometry to Number. Span duration is individual duration; the displayed extent
runs from earliest supplied start to latest supplied end. Neither is a derived
root/request latency. Zero-duration spans retain zero duration even though a thin
visual marker remains visible. Parent indentation follows reported IDs within
supplied trace IDs; missing parents and cycles are labelled. Duplicate IDs remain
separate rows, with explicit ambiguity caution; parent matching retains the
existing bounded renderer behavior. Async links never create parent edges.

## Application adapter

`GooVisuals.waterfall(container, rows, truncated, relatedActions, detailsId)` remains
the compatibility entry point. It mounts the component, listens to native row
clicks and supplies controlled selection. The adapter owns the separate detail
section, `GooAttributes` inspection, callback actions and async-link navigation.
The selected detail duration is formatted from exact `duration_ns`; it no longer
requires callers to also provide a redundant floating-point `duration_ms`.
Existing `.waterfall-row`, `.span-label`, `.span-track`, `.span-bar`, trace/span
metadata and detail IDs remain available to app handlers and tests.

The app loads `waterfall.js` before `visualizations.js`. `waterfall.css` supplies
component-scoped fallback styling. Embedding view styles still own context-panel
layout. Query scope, trace expansion, paging and saved evidence formats are
unchanged. A component upgrade starts no subscription or source read.

## Validation and storyboards

The frontend geometry checks cover exact relative nanoseconds, trace-scoped
parents, cycles, zero duration, input validation and duplicate-row indices.
`journeys/waterfall-components.spec.js` has synthetic conformance checks for SSR,
pre-upgrade properties, detached snapshots, controlled native keyboard selection,
safe text, reconnects, unavailable/empty values and mobile containment without API
reads. It also exercises the compatibility adapter's detail/action ownership.
Frames 67a/67b are synthetic reference screenshots. Real trace storyboards remain
journeys 29, 30 and 45, using ordinary OTLP records and source-backed expansion.
Actual browser campaign results and refreshed frames are recorded in
[VALIDATION.md](../VALIDATION.md). Synthetic conformance fixtures remain separate
from the source-backed trace storyboards.

The read-only `verification/bombadil/inspection.ts` actor now checks controlled
waterfall selection against supplied row indices/IDs, finite 0–100% geometry,
visible tracks and at most one selected row per plot. Once a real seeded trace
question has records, it deterministically opens a trace, selects two rows and
checks the detail heading/IDs before resuming broad inspection. Its snapshot
version is `goo.inspection.v2`; campaign coverage can require
`waterfallStage == "complete"`. No fixture rows are injected by the actor.

Legacy compatibility was checked against `grammar.py`: its trace SELECT already
returns exact `timestamp_ns` and `duration_ns` alongside `duration_ms`. The operation
adapter also supplies `duration_ns`; the old layout required it before extraction.
There is therefore no existing duration-ms-only source contract to reinterpret.
