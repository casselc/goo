# Contextual guidance 460: independent acceptance

Review scope: the integrated Help owner, passive guidance component, static reading
state, workspace markup and styling. This follows the 459 designer and UX reviews.

## Source and interaction review

Accepted for the implemented scope:

- Help is explicit and initially closed. Its inline placement preserves workspace
  lane widths and does not overlay source controls or require a focus trap.
- The application owner listens to route and Explore-mode context. It does not
  navigate, query, subscribe, change filters, save data or dispatch source actions.
- The display component receives closed, immutable reading state and emits only
  explicit reading intents. Stale, hidden, detached and disabled controls cannot
  invoke those intents.
- Five manual pages teach an investigation; Next advances reading position and
  never claims that an application task succeeded. Context changes retain that
  position. Save definition, Keep and Capture have distinct meanings.
- Help has a native complementary landmark, heading, buttons and closed native
  disclosures. Escape closes Help while focus is within it. Close restores the
  header trigger; Return to work closes Help and focuses the named workspace
  outside live record viewports.
- Paging preserves equivalent button focus, or moves focus to the Help heading
  when the focused control disappears or becomes disabled. Updates do not acquire
  focus from application controls. The header remains outside the scrolling body.
- Coarse-pointer targets have a 44-pixel minimum; visible keyboard focus and
  forced-colors rules are provided. Current failures and custody facts remain in
  their application owners rather than being replaced by tutorial prose.

Validation independently run:

```sh
node --test tests/guidance460.test.mjs tests/guidance-view460.test.mjs tests/contextual-guidance460.test.mjs
```

All three test files passed, with zero skips. These checks cover closed state,
reading transitions, malformed input, stale action withdrawal, focus boundaries,
route context, panel return and disposal. They do not substitute for real-browser
or assistive-technology testing.

## Browser acceptance

Independently inspected the Activity guide screenshots at 1440, 390 and 320
pixels in `docs/journeys/screenshots/460-activity-guide-*.png`. Accepted: controls
wrap without clipping, the Close control remains visible, reading position is
clear, and application scope and paused-state facts remain outside the guidance
panel. The panel does not overlay source controls or narrow desktop lanes.

Also independently inspected Services Help and fixed-query source-failure guide
frames at 1440, 390 and 320 pixels (`460-services-help-*` and
`460-fixed-failure-guide-*`). Accepted: service counts, coverage unknown and Not
observed remain factual source output outside optional topics; source refusal
remains visible in normal document flow alongside its recovery controls. Narrow
layouts wrap controls without an overlay or clipped Help dismissal. The guide
does not turn an unavailable read into a successful investigation.

Independently inspected Saved Help at 1440, 390 and 320 pixels
(`460-saved-help-*`). Accepted: Saved context has its own topic headings, stored
evidence identity/count/window remain outside Help, and record cards and controls
remain readable at narrow widths. Close and Return to work remain unobstructed.
Responsive visual acceptance is complete for the four exercised surfaces.

Final executable journey passes (run `95cb4640e2b7`), with zero browser errors
and successful runtime cleanup. Its independent Bombadil driver records 339
snapshots and reaches phase 8 with no API POSTs, route changes or overflow. The
source and responsive visual reviews above are accepted independently; they are
not a claim of complete WCAG conformance or a screen-reader audit.
