# Passive count time bars

`goo-time-bars` is a framework-neutral snapshot display, adopted by the ordinary
Explore time chart and automatic service widgets. It shares count geometry,
axis/scale labels, zero/unknown treatment and accessible interval labels. It does
not replace grouped numeric/duration charts or native metric evaluators.

Load `time-bars.css` after product styles and `time-bars.mjs` before the adopting
app adapters. The standalone element has no dependency on the app or `GooUI`.
Use a JSON `data` attribute or detached `.data` property; properties assigned
before registration upgrade, and reconnects rerender retained snapshot data.
Malformed attributes, unsupported/over-budget point sets and invalid domains
render an unavailable explanation, never silently clip an over-domain count.

```html
<goo-time-bars data='{"title":"Captured activity","unit":"records per interval","ceiling":"20","fromLabel":"12:00 UTC","toLabel":"12:02 UTC","points":[{"key":"first","value":"0","label":"12:00–12:01 UTC: 0 captured records"},{"key":"second","value":"20","label":"12:01–12:02 UTC: 20 captured records"}]}'></goo-time-bars>
```

The caller supplies title, unit, end tick labels, one exact nonnegative integer
ceiling and up to 256 points. Each point has an opaque unique `key`, exact string
`value`, accessible `label` and explicit `state`: observed, no_data, unknown or
unavailable. No-data requires a supplied zero; unknown/unavailable cells do not
claim a numeric count even if a caller retains one privately. Labels and values
are never executable markup. The element computes display geometry with BigInt;
integer strings never pass through floating-point measurement arithmetic.

True zero has zero filled height. A tiny positive value may receive a two-pixel
minimum for visibility; its exact count remains in the accessible label and
`data-count`. Unknown/unavailable uses a hatched full track, not a positive bar
or observed zero. The visible supplied domain stays unchanged when an adapter
removes locally hidden points. It is not recalculated from the visible subset.
A zero ceiling is valid only when all known points are zero.

The default has no buttons or selection. `interactive:true` supplies native
buttons that emit `goo-time-select` with `{key}`; the application owns what that
request means and whether any source read occurs. Selection is controlled via
`selected-key` / `.selectedKey`; clicks do not change it automatically. A point
may supply `href` for an ordinary native anchor. Only HTTP(S)-resolving targets
are accepted, including fragment and relative URLs; script/data targets are
withheld. Links are caller-authored and the component constructs no question.

Explore uses the opaque selection key to display the already-captured interval's
exact bounds/count/duration/numeric details. It makes no new source read. Service
widgets supply existing authored fixed-question links, preserving normal native
link behavior and the selected environment/window. Native interval lengths,
full totals and evidence bases remain the adapter's responsibility. These count
bars do not turn unequal intervals into rates, infer source completeness, or
establish service health. Grouped small multiples remain on their existing
common-scale renderer pending a separate extraction.

The host is block-level, has no page grid and accepts `--time-bars-height` (default
100px). `density="compact"` reduces caption/axis text; Services sets its own 72px
height. Shared `.time-bars` and `.time-bar` classes preserve current journey and
Bombadil selectors. Data snapshot updates replace inner DOM and reset child focus; this is bounded
snapshot rendering, not a high-frequency keyed tailer. Controlled selection
updates only native aria-pressed/current attributes, retaining existing bars and
keyboard focus.

`tests/time-bars.test.js` checks exact scales, zero/absence, domain withholding,
point-key integrity and safe link targets. Journey 68 in
`journeys/time-bars-components.spec.js` checks standalone SSR/SPA upgrade,
detached updates, controlled native-keyboard selection, no source reads,
reconnects, zero/positive/unknown geometry, safe labels/links and mobile display.
Its screenshots are synthetic conformance fixtures. Existing real core-basics
and service-dashboard journeys cover the app adoption and its source contracts.

## Exploratory actor coverage

`verification/bombadil/time-bars-contract.ts` supplies one shared DOM extractor
for exact count labels, supplied state/domain geometry, true zero versus positive
minimum height, unknown hatching and controlled native selection. The Services,
core-basics and grouped-measures actors import that contract.

The Services actor requires an actual chart inspection, retains a focused native
interval link during selected-key updates and checks all interval links stay
inside the widget's fixed environment/window. The core-basics actor compares the
component's supplied points/domain to stored capture time-series values and
requires keyboard focus/control inspection of an actual bar before completion.
The grouped-measures actor applies these checks and offers bar inspection when
a capture contains a chart; journey 34's existing fixture has no time bucket, so
it is not counted as positive time-bar coverage. Run journey 33/core-basics with
`GOO_CORE_BOMBADIL=1` and journey 59/service-dashboards with
`GOO_DASHBOARD_BOMBADIL=1` for the actual chart campaigns. The round's validation
record must distinguish implemented actor coverage from campaigns actually run.

Standalone journey documents explicitly use `text/html` at a synthetic same-origin
fixture URL while loading the actual served component assets. Navigating to a
JSON endpoint and then calling `setContent` retains a JSON document type and
cannot establish HTML custom-element upgrade behavior. The fixture intercept is
synthetic conformance setup, not a replacement for real application telemetry.
