# Independent accessibility review — 2026-11-12

Reviewed commit `837bce4d94e25feae077f87240b512f3935caebb`. This review was performed independently, without consulting the authors’ or other reviewers’ conclusions. Scope: operator journeys and shared UI accessibility. No product source was changed.

## Findings

### A11Y-1 — Medium: Activity service disclosures discard their visible label

**Location:** `ui/tail.mjs:128`.

The service disclosure visibly contains a namespace/service string, such as `platform / catalog-db` (see `docs/journeys/screenshots/operator-activity-390.png`), but its summary receives `aria-label="Service actions"`. That explicit name replaces its visible text in accessible-name computation. Every service disclosure consequently has the same name, and the name does not contain the visible service label.

**Reproduce:** Open Explore Activity with loaded records; inspect a service disclosure summary’s text and `aria-label`. Compare two records belonging to different services. Both controls are named “Service actions” despite showing distinct service identities.

**Impact:** Speech-input users cannot address the control by its visible service name; users navigating disclosure controls by name lose service identity. This is a source-confirmed label-in-name issue (WCAG 2.2 SC 2.5.3, A); actual speech or screen-reader behavior was not tested.

**Recommendation:** Keep the visible label in the accessible name, e.g. `platform / catalog-db — service actions`, or retain the summary text as its name and provide supplementary instructions separately.

### A11Y-2 — Medium: Resource tabs lack programmatic tab/panel relationships

**Locations:** `ui/index.html:12`, `ui/resource-tabs.mjs:4`, `ui/app.mjs:161`.

Services, Hosts and Kubernetes expose `role="tab"` and selected states, but have no `aria-controls`. Their content containers have neither `role="tabpanel"` nor `aria-labelledby` linking back to the tabs. The mounting helper updates visibility, selection and roving focus only; it does not fill these relationships. The adjacent Explore and Definitions tab sets already provide these relationships.

**Reproduce:** Open Services & resources, then select Hosts and Kubernetes. Inspect each tab and its corresponding `serviceEntityPanel`, `hostEntityPanel`, or `kubernetesEntityPanel`. Observe that the visual switch has no associated tabpanel semantics or naming relationship.

**Impact:** Assistive technology receives a partial tab interface and cannot programmatically identify the named panel associated with the selected resource tab. Relevant WCAG 2.2 criteria: SC 1.3.1 (A), SC 4.1.2 (A). Source establishes the missing relationships; no screen-reader-specific failure is claimed.

**Recommendation:** Set each tab’s `aria-controls` to its panel ID and give each panel `role="tabpanel"` and `aria-labelledby` referencing the tab. Keep the existing arrow/Home/End and roving-tabindex behavior.

### A11Y-3 — Medium: Navigation label and enabled search placeholder fall below text contrast minimum

**Locations:** `ui/product.css:139` (navigation label), `ui/product.css:1760` and `ui/tail.mjs:102` (Find loaded records field).

Focused live computed-style measurements in `artifacts/independent-ui-review/narrow.json` establish two normal-size text failures in the default dark theme:

| Text | Foreground / background | Size / weight | Contrast |
| --- | --- | --- | --- |
| INVESTIGATE navigation group label | `rgb(98,114,139)` / `rgb(13,17,24)` | 10px / 650 | 3.871:1 |
| Find loaded records placeholder in enabled empty search | `rgb(117,117,117)` / `rgb(21,26,35)` | 12px / 500 | 3.786:1 |

**Reproduce:** On desktop, inspect the visible sidebar group heading; in populated Activity, leave Find loaded records empty and inspect its rendered placeholder color and opaque input background. Compute sRGB relative luminance and contrast for these pairs.

**Impact:** Both are below the 4.5:1 requirement for normal-size text (WCAG 2.2 SC 1.4.3, AA). Neither qualifies as large text or disabled/inactive text. Placeholder text is included in this criterion; the field’s programmatic name does not remove the visual contrast requirement. The navigation grouping text is information, rather than a logo or incidental decoration.

**Recommendation:** Increase label and placeholder luminance against their actual backgrounds. Explicitly style `::placeholder` to avoid relying on the browser’s insufficient default gray. The footer’s measured 4.775:1 passes; this is a focused finding, not a claim that all muted text fails.

## Observations and coverage limits

- Inspected actual stored frames: operator fixed results at 1440 and 390 pixels, Activity at 390, Activity inspector at 390, and Services at 390. Fixed results and services visibly stack at narrow widths; screenshots alone do not establish 320 CSS pixel reflow or 400% zoom compliance.
- Main/nav landmarks, document language, named capture dialog, native disclosure controls, named record-selection checkboxes, table captions/column headers, and visible keyboard focus styles are present in source. The capture flow uses `showModal()` and native cancellation. These are positive implementation observations, not end-to-end accessibility passes.
- Shared inspector Escape handling in `ui/components.mjs:17–23` checks composition state and defers to editors. `ui/inspection-focus.mjs:13` adds another Escape handler without composition guards, although it excludes common text inputs; keyboard/IME interaction requires an actual composition-session test before treating this as a defect.
- The shared result-state component supplies polite, atomic status semantics (`ui/components.mjs:60–62`). Activity updates its status text on each SSE batch (`ui/tail.mjs:120,132`). Actual announcement frequency and interruption behavior remain untested; no screen-reader announcement claim is made.
- Focused contrast measurements identify the two failures above. A comprehensive measured contrast audit of every state, chart and control boundary was not performed. Small text alone is not classified as a contrast failure.
- Shared tables preserve native table markup, but narrow layouts change CSS display types (`ui/record-presentation.css:37–54`, `ui/components.css:260–268`). Browser accessibility-tree and cross-browser assistive-technology checks remain necessary; no inferred table-semantics failure is reported.
- No automated scanner or screen reader was run by this reviewer. Browser execution was centralized with the coordinating reviewer to avoid concurrent resource exhaustion. Requested runtime probes cover accessible names, resource tab relationships, 320-pixel overflow/target geometry, unnamed controls, and composing Escape. This report does not certify WCAG conformance.

## Coordinated live validation of this reviewer’s requested probes

The coordinating reviewer ran the requested probes serially against the running app. Evidence: `artifacts/independent-ui-review/live.json`. These observations confirm the source findings without relying on another reviewer’s conclusions.

- **A11Y-1 confirmed:** Visible service summary strings included `lab-grpc` and `lab-consumer`; all inspected service summaries had the accessible-name override `Service actions`.
- **A11Y-2 confirmed:** All three resource tabs had null `aria-controls`; all three resource panels had null role and `aria-labelledby`. The browser accessibility tree exposed tabs but no associated tabpanel. ArrowRight correctly selected and focused `entitiesHosts`, so this is a relationship/semantics defect, not an observed arrow-key failure.
- At a 320 CSS pixel viewport on Hosts, document and body widths were both 320; no visible enabled target was under 24 pixels in the measured set. The visible lookback and host-environment fields were named and had associated labels. This narrow Hosts check does not establish all-state reflow or target-size compliance.
- In the Activity inspector, a focused empty search field receiving a synthetic Escape key event with `isComposing: true` and keyCode 229 stayed visible and retained input focus. This probe did not reproduce the potential composition concern; it was a synthetic event, not a real operating-system IME session.
- At 1440 × 500, the exercised native focus/menu paths had no observed occlusion. Other zoom levels, browser engines and physical assistive technologies remain untested.

The live checks supplement source and screenshot review. They are not a screen-reader test or a conformance certification.

Additional narrow-state evidence: populated Activity measured 320 CSS pixels document width at a 320-pixel viewport. Fixed results measured 332 pixels. A subsequent coordinating-browser isolation check (`artifacts/independent-ui-review/overflow.json`, `isolation`) traced the 12-pixel excess to the visually hidden `.telemetry-table` caption: its computed 320-pixel width began at x=12 with BODY as offset parent. A browser-only `inline-size:1px!important` override reduced document width to 320. This is a measured layout defect; this accessibility review does not automatically classify the small overflow as a WCAG reflow failure. After excluding descendants of closed disclosures, remaining small targets included 18-pixel-high summaries and 16-pixel checkboxes. Adjacent spacing may satisfy the WCAG 2.5.8 spacing exception, so no target-size failure is claimed. No product source changes were made for these probes.
