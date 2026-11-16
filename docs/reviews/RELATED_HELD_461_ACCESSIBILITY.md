# Related held records 461 — independent accessibility review

Reviewed 2026-11-16: passive chronology, application inspection owner, shared
reported-context element, native controls and 1440/390/320 documentation frames.
This is a focused accessibility/web-standards review, not a full WCAG certification
or screen-reader audit. No implementation files were changed by the reviewer.

## Acceptance finding

**A1 — fix before closing this polish pass:** selected occurrences use only
`tr[aria-current=true]` background shading. There is no visual text/shape cue or
forced-colors boundary. A real Chromium probe using the current component CSS
produced selected `rgba(255,255,255,.09)` and ordinary
`rgba(255,255,255,0)` under forced colors, with no outline or border on either.
On the default white Canvas both composite to white, erasing the visual selected
state. `aria-current` correctly preserves programmatic state but does not supply
a visible alternative. Add a persistent non-color indication and an explicit
system-color boundary under forced colors. Verify selected/unselected rows in
both light and dark forced palettes without weakening native keyboard focus.

## Recommended refinement

**A2 — anchored span context:** Same reported span stays attached to the original
scope anchor even when the operator inspects another related observation. Its
actual SpanId and unavailable reason currently live only in the label's `title`.
Associate an on-screen anchor label or concise context with the checkbox using
`aria-describedby`; provide an accessible explanation when it is disabled.
The explanation must distinguish the fixed scope anchor from the current
inspector. A full ID can remain in an on-demand disclosure if its relationship
and any unavailable state are clear without hovering.

## Accepted aspects

- Native buttons, checkbox/label, `details`/`summary`, captioned table, semantic
  timestamp elements and grouped source facts avoid custom keyboard emulation.
- Entry focuses and reveals the related heading. Close falls back to that visible
  heading when the original opener was replaced. Return restores selection and
  a visible Inspect control rather than focusing the hidden chronology Find.
- Checkbox updates preserve its focus; the existing atomic polite count/page
  announcement updates without adding competing live regions.
- Each Inspect action has a signal/occurrence accessible name. Current selection
  uses `aria-current`; missing, ambiguous and unreported parent context remains
  explicit and does not infer causal or healthy state.
- Related mode keeps its native return control in normal flow, has no modal/focus
  trap, and does not hide application failures behind teaching text. Native
  disabled/readiness and stale-control guards withdraw invalid controls.
- 390/320 frames show wrapped scope/actions with no clipped checkbox or Return
  action. The related checkbox label and buttons have 44px coarse-pointer sizing.
- Source isolation is preserved: the passive view derives a subset of the admitted
  held bank; it has no fetch/subscription/navigation/preservation path.

## Checks and follow-up

Independently ran:

```sh
node --experimental-test-module-mocks --test \
 tests/held-chronology-view458.test.mjs \
 tests/held-chronology-owner458.test.mjs tests/related-held461.test.mjs
```

All three files pass, zero skips. Standalone Chromium forced-colors probe used
actual `ui/held-chronology-view.css`, with selected and ordinary rows, without
using the shared runtime port. Follow-up acceptance is pending A1 and the agreed
A2 decision, plus new native browser/forced-color evidence from the final build.

## Follow-up acceptance — 2026-11-16

A1 and A2 are resolved in the reviewed source:

- Selected rows now include a persistent **Selected** text marker with a border.
  It is `aria-hidden` because the row already exposes `aria-current`; visual and
  assistive-technology state stay equivalent without a duplicated announcement.
- Forced colors supplies a 2px `Highlight` outline on the current row and
  `CanvasText` text/border over `Canvas` for the marker. Independently reran an
  actual Chromium CSS probe in light and dark forced palettes. Selected outline
  is solid 2px in both; ordinary rows have no outline. Marker contrast is black
  over white and white over black, respectively (21:1 in both palettes).
- The **Anchor span** ID or **Anchor span ID not reported** is visible in the
  scope. The native checkbox references that stable, per-component anchor via
  `aria-describedby`. Inspecting another related occurrence does not silently
  change this filter anchor. No hover is needed to identify or explain it.
- Return emits an explicit local restoration intent. The application resolves
  the original retained occurrence independently of whether its row remains
  visible on the recalled page, and uses the visible chronology Find as the
  focus fallback. This avoids activating a hidden row or making a source read.

Independently reran the same three focused test files after the fixes; all pass,
zero skips. No remaining blocking source, native-semantics or forced-color finding
for this bounded feature. Independently inspected the final refreshed 1440/390/320 normal-color frames
and `461-related-forced-colors-390.png`. The visible Anchor span, selected text
marker and forced-color selected boundary are readable and unobstructed; all
scope controls and Return remain in normal flow and wrap without clipping. The
integration browser additionally verifies the visible marker, solid 2px outline,
checkbox description and restoration when the original row is hidden by Find.
Accessibility follow-up acceptance is complete for this bounded feature. This
does not claim a full screen-reader or WCAG audit.
