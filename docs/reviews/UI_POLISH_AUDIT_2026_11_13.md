# Post-415 whole-surface UX audit

Initial independent review and final bounded acceptance, 2026-11-13. The
follow-up sections accept the updated surfaces at 1440/390/320px. This is a
visual/source audit; root-owned journeys supply fresh interaction checks.

## Scope and evidence

Reviewed the existing operator-workbench and result-presentation design notes,
shared filter/comparison sources, and actual journey frames:

- `multi-value-facets-selected-1440.png` and `-320.png` (421/422).
- `service-versions-error-fraction-320.png` (415/419).
- `captured-rates-1440.png` (418).
- `exception-groups-1440.png` and `-390.png` (416 and its dedicated polish).
- `contextual-actions-inspector-390.png` (420).

All are in `docs/journeys/screenshots`. Frames are captured at different scroll
positions; this review does not infer absolute page-entry positions from them.
417/423 are primarily admission/ownership changes, covered here through their
presented result states rather than a separate feature screen.

The shared baseline is substantially improved: paired measures have consistent
roles and scales, record actions are discoverable, and exception tables remain
structured at narrow widths. The biggest remaining issue is presentation of
supporting context competing with the data it qualifies.

## Prioritized changes

1. **Make counted facets data-first.** In `ui/filter-editor.mjs`, a 320px frame
   repeats the exact UTC window before the load controls and again in a large
   result-state paragraph. Together with the technical source label, this
   consumes roughly 300–400px before the first checkbox. Show one short scope cue
   and a result line such as “261 records · 5 values · coverage unknown”. Move
   the full source/precedence/window identity into the existing Read details.
   Keep the preview scope available before an explicit read; retain active
   failures visibly. Avoid presenting a successful uncertain result as a large
   alert-like indented block.

2. **Unify local-search and selection toolbars.** The facet search remains about
   130px wide at 1440px and clips its own placeholder. Give it a flexible desktop
   width and full mobile width through `ui/filter-editor.css`, reusing the local
   search pattern used by exception groups. Put “2 selected”, “Stage values”
   and a quieter Clear control in one wrapping action region. Keep distinct
   accessible names and original-count descriptions. Checkbox row hit areas and
   right-aligned tabular counts are already good; preserve them. Do not replace
   native controls or merge missing, empty, and literal string values.

3. **Use one comparison-context line.** The mobile Population workspace repeats
   “Independent source reads; matching selectors are not identity joins” and
   “Independent populations · upstream coverage unknown” above the actual
   comparison. In `ui/population-workspace.mjs` and the shared comparison host,
   retain one operator-readable line: “Compared independently · coverage
   unknown”. Put the precise identity-join explanation in Scope and read facts.
   Preserve the title, baseline/selected roles, exact operands and explicit
   diagnostic/no-data states. The paired bars themselves should remain intact.

4. **Name measure disclosures by the task.** `ui/measure-summary.mjs` exposes
   “derived” three times in a single comparison. “Calculation” (or “Exact value”
   where appropriate) better explains what opens. The details must still expose
   the supplied derivation state and ratio support; uncertainty must not become
   hidden behind a friendly label. This is a shared display-component change,
   useful to both recorded error fractions and captured rates.

5. **Strengthen exception count/share legibility without enlarging cards.**
   Exception groups have a useful desktop table and mobile reflow, but the
   primary count and share are tiny beside the extensive signature metadata;
   the native meters are visually faint. Give the count modest numeric emphasis
   and a consistent compact share rail within the existing cell. Keep the
   signature cue secondary and full signature in its native disclosure. Retain
   a single UTC date when first/last share a day. Avoid adding another card or
   a new interpretation of the captured share. The owning exception styles in
   `ui/product.css` should eventually reuse the shared numeric/rail tokens.

## Already acceptable / avoid churn

- Inspector controls and visible Inspect/More row actions have clear targets;
  the selected frame shows strong keyboard focus. No redesign of action
  ownership is justified by these frames.
- Service error comparisons distinguish support counts and show a readable
  percentage-point change, with aligned baseline/selected values at 320px.
- Exception groups use a real table, deliberate mobile hierarchy and explicit
  Stage action. Preserve the bounded list and local search.

## Acceptance for the implementation pass

Regenerate actual 1440/390/320px facet and comparison frames and representative
exception/inspector frames. Confirm there is no clipped local-search placeholder,
no new horizontal overflow, readable values before duplicated provenance, and
unchanged keyboard focus/hit targets. Verify the existing exact scope and read
facts remain available, unknown coverage remains visible, and all explicit read,
staging, empty/missing and no-data behavior remains unchanged. Do not introduce
telemetry reads on disclosure, search, selection or component lifecycle.

## Independent follow-up acceptance — comparisons and exceptions

Reviewed the refreshed `captured-rates-{1440,390,320}.png` and
`exception-groups-{1440,390,320}.png`, plus the production diff, without authoring
those component changes. The captured-rate frame is identified by fixture
`15e7beac96`; the exception frames by `b4fd89e96f`.

**Accepted for these surfaces.** Paired rates retain aligned roles, readable
values, count/exposure support and visible unknown upstream coverage. Native
**Calculation** summaries describe the action more clearly than “derived”; their
source accessible names distinguish measure and operand, while the derivation
state remains inside the disclosure. The long identity-join caveat is now in
native scope details. Neither mobile frame shows clipped controls or horizontal
content loss. The actual keyboard focus outline remains conspicuous.

Exception desktop columns and mobile grouped rows retain a consistent signature,
count/share, UTC window and explicit Stage action hierarchy. Search fills its
available row; the source change gives repeated native signature summaries
contextual names without repeating those long names visually. Numeric/share
styling is retained rather than enlarged in this pass; it is small but legible
and does not justify another component variation.

This is visual/source acceptance, not a separately executed keyboard or
screen-reader test. Root-owned journeys supply interaction/source-budget checks.
At this intermediate review the final facet and service Error-comparison frames
were not yet refreshed. The final follow-up below supersedes that limitation.

### Counted-facet follow-up

The final 19:08 UTC `multi-value-facets-selected-{1440,390,320}.png` frames
(fixture `d133796db8`) are **accepted**. The successful result now leads with
count/value/coverage facts, a collapsed exact-read disclosure and a full-width
search. At narrow widths, selection status occupies its own row, with Stage and
Clear sharing the next row; Clear no longer becomes an orphaned third row.
Missing key, present empty string and the literal `"Missing key"` remain visually
distinct, counts stay aligned, and native focus outlines are clear. Desktop keeps
the same components with a single selection-action row. The final Activity observation
search refinement is accepted below after its refreshed frames.


### Final Activity and service-comparison acceptance

**Accepted**, following inspection of the final 19:12 UTC
`activity-facet-entry-ready-{1440,390,320}.png` and corresponding counted frames
(fixture `da811e4bb7`), plus the 19:09 UTC
`service-versions-error-fraction-{1440,390,320}.png` frames
(fixture `8aa9bb11a2`). The observed-field search now uses the available width at
all three sizes; its short placeholder and the compact incomplete-suggestions
note are readable. Exact observation provenance is behind a clearly named native
disclosure. The latest completed window and its distinction from the inspected
record remain visible, preventing compactness from hiding a change of scope.
Counted results preserve the previously accepted data-first hierarchy, including
zero-selection disabled actions and clear keyboard focus.

Service Error fractions retain aligned baseline/selected values, numerator and
denominator support, unset/unknown status counts and a percentage-point change.
Calculation controls and the shorter independence status match captured-rate
comparisons. No clipping or unintended horizontal reflow is evident in the final
frames. No UX blocker remains for this bounded polish pass. This acceptance does
not claim an exhaustive product redesign or independent screen-reader execution;
it covers the changed shared surfaces and the actual frames listed above.
