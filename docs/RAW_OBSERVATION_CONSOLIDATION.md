# Compact raw-observation inspection

Gauges, sums and explicit histograms share a compact read context, Find/paging
controls and inspector facts. Exact timestamps remain in native `time` elements
and tooltips; visible clocks use readable UTC milliseconds. The fixed window
keeps its half-open boundaries.

The loaded-point announcement remains available to assistive technology without
repeating the visible page count. Generic source/provenance caveats live in one
**Read scope and limits** disclosure. Recording uncertainty, no-recorded values,
invalid payload and unavailable statistic presence stay explicit in the point
presentation. They are not collapsed into a generic successful-read state.

## Shared presentation

The display-only `goo-stat-list` supports serializable facts, including SSR
`items` attributes and detached SPA property assignment:

```js
UI.stats([
  {label: 'Source read UTC', value: '2026-11-15T12:00:00.123456789Z', format: 'utc'},
  {label: 'Event', value: '2000000000', format: 'unix-ns'},
  {label: 'Window', value: '[1000000000, 61000000000)', format: 'unix-ns-range'}
], {density: 'compact'})
```

Typed formats preserve the exact instant or both range boundaries. Invalid
formatted values show Unavailable. Plain numeric/text facts keep their existing
behavior; facts accept no DOM or HTML payload, own no source read and emit no
acquisition event. Source admission remains in the adapter.

The raw application uses one closed profile map for endpoint, response version,
label and instrument fields. Unknown/exponential kinds refuse rather than
falling through to gauges. Numeric, temporal, bucket and whole-response validators
remain family-specific. Go instrument decoding shares exact-key admission;
case-variant keys cannot override the declared tuple. Empty metadata, zero
temporality and false monotonicity keep their meanings.

## Measured result

Compared embedded 447 with 448 using the same actual OTLP journeys, held
Find/Inspect state, viewport sizes and zero source work during capture. Table
values are pixels from panel top to the first rendered point, rounded to the
nearest pixel. These are layout measurements, not latency benchmarks.

| Family | Width | Before | After |
| --- | ---: | ---: | ---: |
| Gauge | 1440 | 298 | 253 |
| Gauge | 390 | 438 | 298 |
| Gauge | 320 | 463 | 317 |
| Sum | 1440 | 298 | 253 |
| Sum | 390 | 477 | 320 |
| Sum | 320 | 521 | 356 |
| Histogram | 1440 | 298 | 253 |
| Histogram | 390 | 438 | 298 |
| Histogram | 320 | 455 | 333 |

The first histogram bucket appears 175/322/474 pixels sooner relative to the
inspector at 1440/390/320. No measured viewport overflows horizontally.

Each cold point read still makes one product read request and three SQL calls
(schema/probe/data). Discovery has its own family-dependent work. Find, local
pages, Inspect and held Return add zero requests and zero SQL. Retirement of an
owned pending read is counted separately from acquiring data. No readiness cache,
polling change or measurement-eligibility shortcut was added.

![Compact histogram detail](journeys/screenshots/448-histogram-390.png)

![Compact sum detail](journeys/screenshots/448-sum-320.png)

## Executable coverage

The existing gauge, sum and histogram journeys capture the measurements through
[`raw-observation-measure.js`](../journeys/support/raw-observation-measure.js).
`GOO_RAW448_LABEL=baseline` keeps prior-binary frames in artifacts;
`GOO_RAW448_LABEL=consolidated-final GOO_RAW448_SCREENSHOTS=1` refreshes the nine
448 documentation frames. Read clocks are inspected through `time.dateTime`,
with the old text fallback used only by this comparison harness.

Real-source refusal, budget limits, duplicate occurrences, retired successful
responses, native model controls and Bombadil remain acceptance gates. The sum
journey now also meters real SQL for its local actions. Source facts and custody
owners were not merged by this presentation consolidation.
