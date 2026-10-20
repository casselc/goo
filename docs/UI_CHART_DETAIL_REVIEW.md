# Chart and detail component review

Source review, 2026-10-20. This is a proposed extraction order, not a claim that
new charts exist. The shared latency/stat/comparison, table, facet and typed-value
elements are implemented. Live log tail work and the waterfall extraction are
being developed separately. This review made no application or runtime changes
and did not run new browser checks.

Implementation follow-up: passive time bars, trace waterfall and
[chart frame](UI_CHART_FRAME.md) are now implemented. The frame is adopted by
[rolling record comparison](LIVE_COMPARISON.md); broader frame adoption and the
remaining proposed primitives below are still planned.

## Recommended order

| Priority | Proposed primitive | First adoption surfaces | Why it earns a component |
| --- | --- | --- | --- |
| 1 | `goo-chart-frame` and `goo-time-bars` | Explore time chart (`distributions.js`), service widgets (`service-dashboard.js`), then grouped small multiples (`grouped-time-series.js`) | Three renderers repeat the same time axis, bar geometry, interval labels and selection explanation. These are immediately comparable uses with different app-owned actions. |
| 2 | `goo-detail-panel` | Trace/log context (`exploration.js`), selected interval/coverage explanations (`histograms.js`, `ratios.js`, `fleet.js`) | Selection should reveal one predictable place for a readable summary, typed fields, exact evidence and native actions. Today selected content is built into separate sections or paragraphs with varying close/focus behavior. |
| 3 | `goo-heatmap` display matrix | Native duration heatmap (`histograms.js`) and raw-span heatmap (`visualizations.js`) | Share cells, axes, legend, keyboard selection and contained overflow while keeping the two measurement adapters separate. |
| 4 | `goo-observation-plot` | Native gauge dots (`measurements.js`), then the queue observation view (`visualizations.js`) | A consistent sample plot can show gaps, freshness and exact sample labels. Preserve independent samples as the default; do not automatically connect them. |
| 5 | `goo-interval-status-strip` | Counter/reset/coverage and ratio/fleet intervals | A compact aligned strip could make observed, no-data and unsupported intervals visible before the exact table, without turning coverage into a health score. Validate the first four adoptions before adding this abstraction. |

Keep the current comparison matrix and aligned bucket comparison module. They
already solve paired values with shared scales; another comparison-card component
would duplicate them. Continue migrating exact tables to `goo-table` rather than
building a chart-specific grid framework.

## First slice: frame and time bars

Start with one measure, one explicit half-open window and one supplied series.
Use a small display contract: title/unit/grain; exact window; already-computed
axis ceiling; supplied interval bounds, value/display label and explicit state;
optional controlled selected key. Preserve decimal/integer source labels, using
bounded normalized geometry separately from exact values.

The frame supplies a figure/caption, unit and time axis, legend, concise state,
and space for caller-authored explanation/actions. It does not generate a
question or calculate count/rate/percentile/error policy. A native caller button
or link may select an interval or open a question; that intent remains in the
app adapter. Keep local hover/keyboard inspection separate from a source read.

Adopt in Explore and a service widget before extracting grouped cards. Grouped
small multiples must retain one common scale across all supplied series,
including series hidden by local search. App layout owns the card grid. Avoid
putting sorting, source-backed paging, recipe navigation or chart composition
inside the display element.

A specific current issue to check during extraction: `.time-bar::after` has
`min-height:2px`, while only `.time-bar.no-data` suppresses it. Grouped charts set
that class, but ordinary Explore count bars do not. A zero-valued interval can
therefore look like a small positive bar. Render a zero baseline marker or no
filled height, with exact zero in the accessible label; reserve hatching or a
separate symbol for unknown/unavailable. This is a source-review finding, not a
browser-verified regression in this review.

## Heatmaps: share geometry, preserve different questions

The raw-span heatmap uses half-open duration bands over captured spans and can
request the actual matching span population. Native explicit histograms use
SDK bucket boundaries with inclusive upper bounds and interval distributions.
Their span drilldown is time/operation context, not exact bucket contributors.
These must remain separate adapters even if both use the same cell component.

The matrix should accept supplied row/column labels, exact cell labels,
normalized intensity, explicit cell state and opaque keys. It must not fill
missing cells with zero, derive stream coverage, interpolate P95, merge bucket
layouts, or infer that every interval has the same duration. Keep observed zero
visually distinct from unknown. A linear/log color transform must be named in
the legend; local search must preserve its supplied domain. Use color plus text
or pattern, with an accessible exact-value table/description.

`histograms.js` currently normalizes intensity through `Number(BigInt(count))`.
A shared implementation should bound geometry without losing the exact count
label or overflowing for unusually large supplied integer strings. Converting a
count to a display color is not permission to use floating point for measures.

## Detail panels: consolidate without hiding the question

A passive shell should accept caller-authored summary, typed attribute list,
exact-value disclosure and action content. Support an accessible heading label,
optional close button and explicit return-focus target. The application owns
which selection is active and whether the panel is a drawer, inline section or
mobile full-screen inspector. Avoid an element that creates unrelated dialogs,
URLs or subscriptions during upgrade.

Keep signal/grain, fixed versus rolling window, service/namespace and any
incomplete/unknown condition visible. Put exact IDs and raw JSON in disclosures.
A selected trace span, histogram bucket and fleet coverage interval need distinct
labels and compatible actions, even when their container looks the same.
Reusing `goo-field-value`, `goo-attribute-list`, `goo-result-state` and native
`details` is preferable to a second attribute or provenance system.

## Live and fixed comparisons

The application delivers each current snapshot with its window, basis/sequence,
coverage and unit together. A tail ring buffer is a row viewport, never the
population for a latency percentile, error denominator or throughput measure.
A fixed baseline must retain its own bounds and basis while current data moves.
Shared charts should accept an explicit stable axis domain, or visibly mark
rescaling; otherwise a fixed baseline can appear to change when only the axis
changed. Do not connect separate gauge observations across missing intervals,
show counter resets as negative rate, average scalar percentiles, or silently
realign native export intervals.

The frame may display connection/stale/gap information supplied by the adapter.
Connection state is separate from telemetry state. Pure display upgrades,
disclosures, local find and reconnects must not subscribe, retry or read sources.
Transport cancellation, ordering, retention, gap/cursor policy and row identity
remain outside every proposed chart and detail element.

## Acceptance and validation

Each extraction should demonstrate server HTML and SPA updates, assignment before
registration, detached properties, reconnect, invalid input, safe text, keyboard
navigation, accessible exact labels and mobile containment without an app runtime
or API requests. A controlled selection must remain a supplied key; tests should
prove local inspection does not create source reads.

Add composition checks in the real app: shared domains across filtered small
multiples, zero versus unknown, long labels, large exact counts, nested table
headers and chart/detail focus return. Geometry tests should compare pixels or
normalized coordinates as well as strings. For live patches, verify node/focus
retention or document snapshot rerendering explicitly.

Reuse live journeys for core time charts, grouped time series, native windows,
operation heatmaps and fleet/ratio explanations. Regenerate their screenshots
and keep a standalone synthetic reference clearly separate. Extend the existing
Bombadil actions/invariants for selection and stable scales. Keep Hegel/Quint
checks at typed measure/state boundaries; moving drawing code alone does not
justify duplicating the Go evaluator in a test. Include rejecting trace mutations
when new app state transitions or source-read boundaries are introduced.

## Current round adoption update

The count-only `goo-time-bars` extraction is now implemented for ordinary Explore
and automatic Services widgets; its contract and verification are in
[UI_TIME_BARS.md](UI_TIME_BARS.md). Grouped time charts, a general chart frame,
heatmap matrix and detail-panel extraction remain proposed. Controlled bar
selection updates native state without replacing focused nodes; data snapshots
still rerender. This review's no-runtime-validation statement applies to the
review itself; the round's actual executed checks belong in VALIDATION.md.
