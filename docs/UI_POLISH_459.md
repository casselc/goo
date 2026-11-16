# Shared interface polish and optional guidance

The 459 pass improves shared attribute inspection, narrow service catalogs,
keyboard bypass, chart-state scanning and chronology announcements. It then uses
independent design and UX reviews to distinguish persistent result facts from
optional explanations and first-use guidance.

## Implemented

- Attribute Search and Scope share a compact labelled toolbar. Native accessible
  names remain; matching counts stay visible. At narrow widths, keys and Copy/
  actions share a row, with the value below. Expanded menus flow at full width.
- Copy and field-action names include their namespace/key, so identical names in
  resource and record attributes remain distinguishable to assistive technology.
- Service cards keep identity, three signal counts, latest UTC and Open together.
  **Not observed** stays explicit; missing metrics are never turned into zero.
- **Skip to main content** bypasses navigation. Its application enhancement keeps
  Services/Saved routes intact; the native anchor still works without JavaScript.
- Chronology uses one atomic polite announcement for counts, page range and errors,
  rather than two competing live regions.
- Shared chart frames pair persistent state text with a neutral ring/filled marker
  or warning diamond. Shapes remain legible in forced colors. Observed is not a
  green health verdict.
- The Activity staging paragraph is now a closed native **Local observation · not
  saved** disclosure. The state remains visible and the existing explanation is
  available on demand. This is the first agreed extraction, not a shipped tour.

These changes do not alter selectors, source reads, sampling/coverage semantics,
measurement totals, saved record identities or preservation actions.

## Measured layout changes

Real OTLP before/after runs use the same fixed 30-span bank for Records, Services
and Saved. Activity uses a fresh ordinary bank per pass for its rolling window.
The native controls and field actions are exercised before screenshots are saved.

| At 390 CSS pixels | Before | After |
| --- | ---: | ---: |
| Attribute toolbar height | 146px | 52.5px |
| First field from attribute-viewer top | 206px | 112.5px |
| Service card height | 228px | 156px |

The first field is 93.5px closer to the viewer top. Long actual messages can still
require scrolling in the bounded inspector; this result does not claim every
attribute is visible without scrolling. Native coarse-pointer controls/actions
remain at least 44px. All six service cells remain visible at 320px, including the
explicit missing-signal state.

## Native contextual-help pattern

Reuse `UI.disclosure` and the shared `ui-context-help` style for optional conceptual
explanations. The summary must retain any current fact that matters during work.
It is a normal HTML disclosure, with no source action or subscription.

```html
<details class="ui-context-help">
  <summary>Local observation · not saved</summary>
  <p>Stages a fixed question for a separate source read; held rows are not saved evidence.</p>
</details>
```

The [consolidated guidance proposal](CONTEXTUAL_GUIDANCE.md) records the independent
[designer](reviews/PROSE_GUIDANCE_DESIGN_459.md) and
[UX](reviews/PROSE_GUIDANCE_UX_459.md) inventories: exact current scope, clocks,
state, values, limits and failures stay visible; recurring definitions and workflow
lessons move into contextual help or optional replayable mini-guides. Larger help
migration and walkthrough work remains proposed.

See [executed storyboard](journeys/ui-polish.md),
[scoped acceptance](reviews/UI_POLISH_459_ACCEPTANCE.md) and
[validation](../VALIDATION.md).
