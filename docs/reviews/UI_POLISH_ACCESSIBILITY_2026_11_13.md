# UI polish accessibility and web standards review — 2026-11-13

Status: accepted within the bounded source and rendered-frame review described below; comparison, exception, service-version, and final counted-facet follow-ups are complete.

## Scope and method

Reviewed shared counted facets, captured rate and recorded Error comparisons, service-version controls, exception groups, and contextual inspector actions. Sources include `ui/filter-editor.mjs`, `ui/filter-editor.css`, `ui/measure-summary.mjs`, `ui/service-version-comparison.mjs`, and `ui/exception-groups.mjs`. Existing 390 px frames reviewed: `multi-value-facets-selected-390.png`, `service-versions-error-fraction-390.png`, `contextual-actions-inspector-390.png`, and `exception-groups-390.png`.

No blocking accessibility defect was established in this bounded review. The following improvements belong in the current polish rather than a new interaction system.

## Findings and priorities

1. **Make calculation disclosures distinguishable.** `measure-summary.mjs` uses the cell state alone as the summary, producing three controls named “derived” in a single comparison. Table headers supply reading context, but a screen reader's controls list cannot identify the relevant measure and side. Use a short useful visible label such as “Calculation”, retaining state as a fact, and provide a contextual accessible name containing measure and side (baseline, selection, or change). Preserve native `details`/`summary`. This is the primary accessibility improvement.

2. **Give counted-value search the available width.** The 390 px facet frame shows a narrow search field that truncates “Search loaded values” despite space across the enclosing panel. An explicit shared full-width search rule improves zoom/reflow and visible discoverability. Preserve its accessible search name and native input behavior.

3. **Reduce scanning burden without removing scope facts.** Counted facets expose environment IDs, precedence rules, timestamps, and coverage prose before values; Error comparison shows “matching selectors are not identity joins” before results. Keep a concise visible scope/quality summary, moving technical derivation and identity explanations into existing native read-detail disclosures. Retain an accessible connection from the loading control to the scope it reads. Coverage unknown, selected-window boundaries, independent reads, and staged-versus-applied state must remain discoverable.

4. **Contextualize repeated exception controls where practical.** Each exception card has “Signature details”; the Stage button already describes its signature via `aria-describedby`. Give the disclosure a contextual name using the same row signature. This is a low-cost navigation improvement, not a blocker: the controls currently have surrounding table/card context.

## Existing strengths to retain

- Counted selections use labelled native checkboxes with counts in `aria-describedby`, a live selection count, explicit Stage/Clear buttons, and explicit hidden-by-search selection state. Missing key and the literal string “Missing key” are visually distinct.
- Service version and measure choices use native selects inside visible labels; Stage is disabled for unavailable or identical versions.
- Comparison values have native table row/column headers and explicit associations. Decorative bars are hidden from assistive technology; exact values and availability remain text.
- Exception groups use a real table, scoped headers, native disclosures, text counts/shares, and a native labelled meter. Scrollable stack/context regions are keyboard reachable and labelled.
- The inspector frame shows a clear visible focus ring on the actual Expand trace control. Shared product styles cover buttons, links, form controls, and summaries with `:focus-visible`.
- Existing action guards and opener restoration avoid substituting synthetic ARIA widget behavior for native controls.

## Follow-up acceptance

After implementation, review narrow and desktop frames for unclipped controls, full-width facet search, clear value/side association, short visible summaries, and retained uncertainty/scope information. Check source for contextual native summary names and ensure no new positive tabindex, custom keyboard widget, or redundant fetch behavior. Root-owned targeted browser tests should retain checkbox Space activation, search-hidden selections, staged scope transitions, native disclosure keyboard activation, and actual-opener focus return.

This review does not claim screen-reader certification, contrast measurement, cross-engine testing, or keyboard execution of the new changes. Screenshot colors alone do not establish a contrast failure. No heavy tests were performed by this reviewer. The reviewer subsequently implemented the assigned comparison/disclosure improvements; follow-up below explicitly limits acceptance to source and rendered-frame review.


## Post-change comparison and exception acceptance

Reviewed regenerated captured-rate frames at 320, 390, and 1440 px (18:59 candidate), and exception-group frames at the same widths (19:00 candidate). Calculation disclosures fit, remain aligned with the operand or change they explain, and retain a clear focus ring. Exact exposure operands, shared scale, coverage uncertainty, and zero change remain visible. Exception cards/table retain signature cues, counts/share, first/last timestamps, qualified context counts, and Stage actions without a new clipped control. The scrollable exception collection intentionally continues beyond the viewport; this is not a missing-card finding.

Source follow-up confirms native Calculation summaries expose measure and side in their accessible names, with visible labels included verbatim. The subsequent support refinement preserves `Calculation · valid/total valid` in the collapsed summary when support exists, rather than hiding that support count. Non-derived availability states remain visible, while derived state and exact facts remain inside the disclosure. Exception Signature details names include type/reason and short fingerprint. The exact independent-read caveat remains inside native population scope details.

Accepted for these bounded source/visual checks; no additional blocker established. This does not substitute for the root-owned keyboard journey assertions, screen-reader testing, or cross-engine accessibility verification. Counted facets await the final refreshed candidate.

### Recorded Error fraction follow-up

Reviewed regenerated `service-versions-error-fraction-{320,390,1440}.png` (19:03 candidate). Accepted within the same source/frame limits: baseline/selected labels remain legible, exact numerator/denominator and Unset/Unknown counts remain visible, the percentage-point change is distinct, and the focused Calculation disclosure is clearly outlined. No clipped value or missing quality fact was established. Final facet-source review also confirms native Observation details and scope preview disclosures, full-width search rules that respect `[hidden]`, retained visible incomplete counts, and explicit Stage/Clear labels; rendered facet acceptance was pending at this intermediate checkpoint; final acceptance follows below.


## Final facet and service acceptance

Reviewed final Activity ready/count frames refreshed at 19:12 after the search max-width correction, multi-value selected/filtered frames refreshed after 19:08, and service-version picker frame refreshed after 19:08. Samples include Activity ready at 320 px, counted Activity at 390 px, selected values at 320 px, filtered values at 390 px, and the service picker at 320 px.

Accepted within this review's stated limits. Observed and counted search fields occupy the available panel width without clipping. Observation incompleteness and counted coverage remain visible above concise native detail disclosures. Exact scope/window provenance remains available in the disclosure body; the inspected-record window difference remains visible in the summary. Stage and Clear fit the mobile action row, with the selection count on its own row. Clear retains the contextual accessible name “Clear selected values,” including its visible label. Hidden-by-search selections remain explicitly counted; literal “Missing key”, absent key, and present empty string remain distinct. Checkbox counts and focus outlines remain visible. Service version choices retain visible native labels, distinct version counts, explicit Stage, and visible coverage/read facts.

Source review confirms hidden search inputs remain hidden, native disclosure/control semantics are retained, and no new positive tabindex or custom keyboard widget was introduced by this polish. Root reports that native-keyboard/exact-provenance journey assertions passed through staging and the frontend unit suite passed; those executions were not independently rerun by this reviewer. There are no outstanding findings from this bounded review.
