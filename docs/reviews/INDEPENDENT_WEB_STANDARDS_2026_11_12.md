# Independent web standards review — 2026-11-12

Reviewed commit: `837bce4d94e25feae077f87240b512f3935caebb`.

Reviewer scope: HTML/CSS, native interaction semantics, Web Components, progressive enhancement, and DOM lifecycle. This review was performed independently of implementation authors and the other reviewers. No implementation changes were made.

## Result

One low-severity responsive CSS defect was identified after the initial pass. This is a source, captured-render, and targeted-browser-evidence review; it is not a claim of cross-browser or assistive-technology certification.

## Finding: low severity — hidden caption escapes the mobile table viewport

Location: `ui/record-presentation.css:66`, conflicting with `ui/record-presentation.css:45`.

Reproduction: open a populated fixed captured-span question at a 320-pixel viewport using the compact record table. Measure `document.documentElement.scrollWidth` and the table caption's bounding rectangle. The coordinating review's recorded `artifacts/independent-ui-review/overflow.json` shows document width 332 against viewport width 320; the clipped caption starts at x=12 and ends at x=332 with width 320. The table headers extend to approximately x=344 but remain inside the table's own scrolling wrapper; those are a distinct, locally contained extent.

Cause confirmed by source inspection: the new hidden-caption selector assigns `position:absolute` and `inline-size:1px`, but the existing mobile selector `goo-table[record-layout=compact]>table>caption` has higher specificity and assigns `width:auto`. In the horizontal writing mode these declarations compete for the same size axis, so the latter defeats the intended one-pixel width. `goo-table` supplies `overflow:auto` without establishing a positioned containing block (`ui/components.css:40`), allowing this absolute caption's box to escape the local table viewport. Clipping hides its paint but does not reliably remove its scrollable extent.

User impact: the mobile page acquires unnecessary horizontal scrolling from content intended to be visually hidden. The measured excess is 12 pixels in this reproduction. This is a concrete layout defect; the measurement alone does not establish a WCAG reflow failure or visible data loss.

Isolation control: `overflow.json` records the caption's `offsetParent` as `BODY`. A temporary browser-only `inline-size:1px!important` on that caption reduced its measured width from 320 to 1 pixel and the document scroll width from 332 to exactly 320. No application source was changed. This isolates the hidden caption as the cause of the page-level overflow in this reproduction.

Suggested correction: reconcile the mobile visible-caption rule with the new visually hidden caption policy so hidden captions keep one-pixel geometry at every breakpoint. Check saved and service compositions too, whose application caption selectors also assign width and padding. Preserve the native caption for assistive technology.

## Evidence inspected

- Repository `AGENTS.md`; commit diff for changed UI modules and styles; surrounding shared-component implementations and action eligibility helpers.
- Actual committed screenshots: `operator-activity-1440.png`, `operator-activity-inspect-390.png`, `operator-fixed-390.png`, `operator-fixed-inspect-1440.png`, `operator-services-390.png`, and `operator-saved-1440.png` under `docs/journeys/screenshots`.
- Existing `journeys/operator-polish.spec.js`, including the hidden-workspace docking check, IME Escape negative control, select Escape check, and close-to-opener assertion. These assertions were inspected, not executed by this reviewer.

## Source observations

- Field action disclosure retains native `details`/`summary` and ordinary buttons. It does not claim ARIA menu semantics that would require implementing the composite-menu keyboard model. The compact summary retains an accessible name (`ui/attributes.mjs:113`).
- Field callbacks now admit the actual clicked native control before dispatch, including disabled-fieldset, inert, closed-disclosure, and hidden-ancestor checks. Dialog return origins use the persistent disclosure summary (`ui/attributes.mjs:11`, `ui/dom-visibility.mjs`).
- Inspector Escape handling avoids composing input, nonempty search input, selects, editable content, and relevant composite editors; field action disclosure Escape prevents propagation before inspector dismissal (`ui/components.mjs:8`, `ui/components.mjs:17`).
- The desktop inspector remains a nonmodal region. Its body has an independently scrollable, shrinkable flex track, and page padding excludes inspectors beneath hidden ancestors (`ui/product.css:1777`). The inspected screenshot shows header and footer outside that scrolling body.
- Record tables retain native table, header, and caption nodes. Responsive presentation changes their layout while retaining source structure. Visually hidden captions use clipping instead of `display:none` (`ui/record-presentation.css:39`, `ui/record-presentation.css:66`).
- No changed custom-element constructor, registration, or lifecycle introduces subscriptions, source reads, or global event listeners. Existing latency/stat setters support disconnected property assignment, and existing table observers disconnect on removal (`ui/components.mjs`).
- Reduced-motion handling covers shared component descendants and the changed product button transitions (`ui/components.css:275`, `ui/product.css:1800`). No new animation was introduced by the reviewed patch.

## Requested live probe results

The coordinating reviewer executed the requested browser probes serially against the reviewed application. I inspected the recorded results in `artifacts/independent-ui-review/live.json` and the actual `fixed-forced-colors.png` frame. No application JavaScript errors were recorded.

- Enter opened the native Field actions disclosure. Four successive Tab steps reached Include, Exclude, Find traces containing…, and Plot values; all recorded focused controls were visible and nonoccluded.
- Opening Plot values moved focus into the native numeric-type select (`#plotValuesnumericType`). Cancel restored focus to the persistent Field actions summary, visible and nonoccluded.
- Tab from the final inspector action reached the next record's timestamp in the left table, visible and nonoccluded. The nonmodal dock did not trap keyboard focus.
- At 1440×500, the dock occupied y=56–484, its header y=67–99, its independently scrollable body y=127.8–435, and its footer y=441–473. The body had 1915 pixels of scroll content within approximately 307 pixels of available height. The footer action was focusable without occlusion, and the following Tab reached a visible record timestamp.
- With forced colors and reduced motion emulated, both media queries matched; computed body text became black on white. The inspected frame retained visible table boundaries and controls. This supports the recorded table frame only; chart appearance was not audited.

These observations resolve the requested menu/dialog handoff and short-viewport docking uncertainties without producing a confirmed finding.

## Remaining verification limits

Browser work was centralized with the coordinating reviewer to avoid concurrent memory-heavy jobs. Live evidence above covers the recorded browser and scenarios. No screen-reader exercise or cross-engine native-search Escape comparison was performed. Forced-colors chart behavior and all application surfaces beyond the inspected frames remain outside this review's demonstrated coverage.

No compatibility-floor objection, speculative polyfill request, or unsupported failure is counted as a finding.
