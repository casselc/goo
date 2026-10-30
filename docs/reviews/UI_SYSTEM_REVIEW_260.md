# Shared UI and control review — 260

Reviewed 2026-10-30 against milestone259 plus this round's bounded shared cleanup.
This is a source/contract review with the specific Chromium checks below, not a
claim of a full accessibility audit or external screen-reader certification.

No applicable design skill is installed: the available cloud catalog contains
runtime/setup skills and the filesystem search found only a Hegel skill. The
review therefore uses the existing component contracts and WCAG/WAI practices;
no new skill, framework or design package was installed.

## Product direction

Goo already has a useful display library. The missing layer is a small, consistent
way to choose a compatible presentation of a scoped observation, rather than
more independent widgets and toolbars. Keep the entry stream compact. Let selecting
records, marks or fields reveal the next useful action. Show scope, time and quality
once, close to the result; put the explanation behind a stable disclosure.

Preserve the established boundary: passive Web Components receive data and emit
interaction intents. Application adapters own query scope, source reads, live
subscriptions, navigation, saving and investigation state. A new chart renderer
must not become another query runtime or fetch from its lifecycle.

## Existing pieces to retain

| Family | Existing implementation | Review conclusion |
| --- | --- | --- |
| Evidence and state | `goo-evidence-summary`, `goo-result-state`, `goo-chart-frame` | Reuse these for scope/quality chrome; avoid another card/status vocabulary. |
| Records | `goo-record-summary/context/columns`, structured attributes, exception detail | Keep ordinary OTel records usable without manifests. Share formatting and explicit missing/empty values. |
| Tables | Native table inside `goo-table`, compact record and stacked metadata modes | Keep captions/headings/native controls. Add conformance fixtures for each layout rather than local mobile overrides. |
| Field and grouping controls | `goo-grouping-control`, `goo-record-columns`, facets | Controlled field identity is the right model. Move toward one qualified-field picker and consistent stage/apply wording. |
| Plots | Time bars, value timelines, aligned distributions, heatmap, waterfall and comparisons | Keep specialized semantic models. Their common frame and intent contract should surround a pluggable renderer, not be replaced by raw Plotly configuration. |
| Inspection and menus | `UI.inspector`, `InspectionFocus`, `UI.actions` | Standardize ownership, close/return and keyboard behavior rather than adding per-surface sheets. U0 is handled separately in this round. |

Sources: [public registry](../UI_COMPONENTS.md),
[display entry](../../ui/display.mjs), [shared helpers](../../ui/components.mjs),
[record columns](../../ui/record-columns.mjs),
[chart chrome](../../ui/chart-frame.mjs),
[current chart guidance](../UI_CHART_DETAIL_REVIEW.md).

## Priorities and evidence

| Priority | Finding | Change or proposed acceptance |
| --- | --- | --- |
| P0, corrected here | Standalone components rely partly on `product.css` for keyboard focus, whose control selector also omits `textarea`. The standalone gallery loads component styles independently. | Shared scoped focus rules cover native buttons, links, fields, summaries and textareas within component hosts. Verify keyboard navigation with no product CSS and no application imports. |
| P0, corrected here | `.attribute-group-count` used `#6c8299` over the attribute viewer's `#111a24`: calculated normal-text contrast4.42:1, below4.5:1. | Use the existing muted token/fallback. Browser computed colors now meet the normal-text threshold. This does not certify arbitrary embedding themes. |
| P0, corrected here | Record-column action buttons did not set their type; a picker inside a caller-owned form could submit it. | The component's local node helper explicitly makes action buttons `type=button`; grouping's owned Add form retains its intentional submit behavior. Browser test verifies one intent and zero parent submissions. |
| P1 | No common reduced-motion policy existed for the scoped display graph. | CSS transitions/animations in component scopes respect reduced motion. A future JS renderer must independently disable animation/tweening; CSS alone cannot stop Plotly transitions. |
| P1 | Styling remains split between1,500 lines of application CSS and over20 feature stylesheets. Names such as `--panel`/`--surface` and `--warning`/`--warn` have different fallbacks. | Establish one documented role-token map and migrate by family. Do not globally restyle every page in one change. Require standalone and nested composition for each migrated family. |
| P1 | Scope, completeness caveats and action hints repeat at several levels; some field disclosures contain whole explanatory paragraphs before useful choices. | Use one concise status line plus exact scope/quality details. Shorten duplicate copy, but never erase partial/no-data/unknown, selected-versus-population, clocks or measure semantics. |
| P1 | A visualization kind has no common compatibility selector/binding contract across the existing plot families. | Add a registry of predefined compatible kinds, field roles and supported local gestures. Kind change over a supplied result is local; any required additional read becomes an explicit staged request. |
| P1 | A generic renderer can lose exact values, missing states, shared scales and independently observed comparison clocks. | Normalize a semantic display model before the renderer. Keep exact/raw values beside projected numeric coordinates; refuse unsupported precision or incompatible comparisons rather than silently coercing. |
| P2 | Native table layout uses CSS display changes for narrow screens. Chromium visual behavior is checked; Firefox/Safari and assistive-technology behavior are not established. | Include accessible table structure, headings, captions and row actions in cross-browser conformance. The native table remains the fallback even when a chart renderer is unavailable. |
| P2 | Dense plot marks may not provide practical touch/keyboard access even when individual marks have names. Existing families implement gestures separately. | Standardize focus/selection/brush/reset intents and add a keyboard range editor, accessible data view and clear focus return. Do not make every dense mark24px and destroy the visual scale; provide an equivalent usable control. |
| P2 | Controlled element/intent documentation and historical audits are manually maintained and interleaved. The component index still links an older roadmap. | Publish a short current manifest/index with versions, inputs, events and examples; move historical review context behind dated links. Generate the type/renderer support gallery from the same registry. |

The contrast calculation uses the standard sRGB relative-luminance ratio for the
actual specified viewer colors. Normal text uses4.5:1 as the acceptance threshold.
Focus/visibility and target guidance follows WCAG2.2 criteria1.4.3,2.4.7,2.4.11
and2.5.8. Motion handling respects the user preference without claiming that it
alone satisfies every animation criterion.

## Styling and copy rules to standardize

Use a small role-token vocabulary: canvas, surface, raised surface, border, text,
muted text, accent, warning, danger, success, focus ring; then typography, spacing,
radius and control-density tokens. Components consume tokens with usable fallback
values. The application owns layout, not component-specific colors or DOM selectors.

Retain native controls. A common field/control row should standardize label,
help/error linkage, disabled/busy behavior and density without hiding native input
semantics behind a large custom-button framework. Group configuration under one
Options disclosure. Do not render a control merely to explain that another control
is unsupported.

A concise display surface should contain:

1. A useful title or record summary, then data.
2. One compact scope/window/quality line, with exact values in accessible details.
3. One primary action; secondary actions in the shared disclosure.
4. A clear unknown/empty/loading/error state. A sentence belongs here only if it
   changes the interpretation or next action.

Examples of copy policy: “Fields from supplied rows” replaces an introductory
paragraph; the same disclosure still states that the source schema can contain
more fields. “Partial page” needs exact retained-count/coverage details nearby.
Percentiles retain the upper-order statistic rule, display rounding, and separation
from end-to-end latency; they are not relabeled as a histogram.

## Base chart and kind contract

The base chart should sit **inside** the existing frame, with a renderer adapter
behind a stable supplied-data contract. Predefined observability kinds should be
the public API: count/rate over time, value/state timeline, ranked breakdown,
distribution/histogram, latency density, scatter, aligned comparison and trace
waterfall. Existing native components can fulfill some kinds; Plotly can fulfill
others. A caller should not need a Plotly trace/layout object for ordinary use.

Recommended inputs are a result/context identity, semantic rows or series, typed
field roles, unit/temporality, exact values, validity/coverage, supplied domain and
selection state. No adapter should average percentiles, infer a complete trace,
fill absent buckets with zero or silently equate histogram interval bounds.

Expose renderer-independent intents for selecting a mark, staging a range/filter,
opening a referenced record and resetting the local view. Carry opaque stable keys
and exact bounds. Do not emit SQL or mutate the selected scope inside a chart.
A renderer-specific escape hatch belongs in a labeled advanced panel with bounded,
validated options; it must not become the normal package syntax.

A kind chooser should show only compatible options, or explain why an option
requires another measure/read. Switching a log table to a count chart over an
already returned count series is local. Requesting a new duration distribution
from raw rows is a new query plan, staged for explicit Run. The visual editor,
code template and preview must round-trip one versioned view IR so WYSIWYG and
source editing do not become separate languages.

## Changes implemented in this review

- Shared scoped focus styling, including textareas, works without `product.css`.
- Component scopes respect `prefers-reduced-motion` for CSS effects.
- Attribute group count uses the shared muted token instead of the failing color.
- Record-column actions cannot submit an embedding application's form.
- Shared latency and record-field help text is shorter, retaining uncertainty and
  interpretation limits.

Files: [components.css](../../ui/components.css),
[components.mjs](../../ui/components.mjs),
[record-columns.mjs](../../ui/record-columns.mjs).
No query, navigation, persistence or trace-owner code was changed by this reviewer.

## Validation

`node --test tests/components.test.js tests/record-columns.test.js
tests/activity-latency.test.js` passes. The browser-only
[shared control check](../../scripts/check_component_controls.cjs) uses real static
component assets and explicitly synthetic supplied controls, not a mocked product
journey. It checks standalone keyboard focus for summary/button/link/input/select/
textarea, reduced motion, computed contrast, embedding-form isolation,390px
nonoverflow and zero application/source imports or requests. Chromium passes.
The parent round owns integration journeys, affected product screenshots and final
whole-project gates. No Firefox/Safari or screen-reader campaign is claimed.

## Clarifications from the final user direction

The user confirmed this is prerelease software with no legacy compatibility floor.
The recommendations above preserve useful behavior and data meaning, not old
component APIs. Consolidate onto the public model directly: replace redundant
properties, aliases and page/control branches, update consumers/tests together,
and remove the replaced implementation. Error recovery for a current unsupported
renderer is still necessary; that is distinct from maintaining an old API.

Complete-data requirements are **hard eligibility**, including for controls and
compositions whose actions depend on exhaustive data. “View as” distinguishes
Available now, Requires complete data, and Unsupported. A blocked kind explains
the missing scope/aspect guarantee and offers an explicit acquisition plan when
possible. It never silently downgrades or draws a partial approximation under the
complete-only kind's name. Eligibility is rechecked on every dataset revision;
withdraw an ineligible current rendering or clearly retain the last eligible
frozen revision as stale. Completeness remains separate from sample support,
empty/no-data, and completeness of real-world telemetry collection.

The grammar covers the **whole UI**, not charts alone: tables/lists/streams,
waterfalls/details, scope-aware search/filter/facet variants, forms, action rows,
layouts and screens use the same public `ViewDefinition` and component contracts.
A user or team can consume and extend these public primitives; built-in screens
must prove parity through representative declarations rather than privileged
hidden screen payloads. Display components consume controlled context roles;
application/host services execute checked operations and own effects.

Three UX/architecture clarifications make that direction implementable:

- **Protected chrome is still public.** Scope/quality summaries should be public
  composable kinds with sealed host-supplied bindings and enforced placement/
  accessibility constraints. Protecting data meaning must not create an
  imperative, internal-only screen language outside the grammar.
- **Shared context does not erase domain types.** Telemetry scope sessions,
  catalog filters/cursors, authoring drafts and documents remain distinct typed
  roles within one context/intent envelope. Saved Find is a catalog operation,
  not a telemetry predicate or implicit source Run. Controls declare local versus
  source-scoped effects and the relevant target owner.
- **Shared selection is revision-bound.** Cross-widget selection references name
  the dataset role/ID, revision and stable key or exact typed range. Bare row/vendor
  point indexes cannot identify a selection across differently ordered widgets.
  Scope changes invalidate derived selections; a retained original snapshot can
  remain an explicitly named parent/history context.

The root [view grammar proposal](../VIEW_GRAMMAR_PROPOSAL_260.md) now reflects the
broader direction. These are proposal clarifications, not newly implemented
component, compiler or editor capabilities. No further code changes were made.
