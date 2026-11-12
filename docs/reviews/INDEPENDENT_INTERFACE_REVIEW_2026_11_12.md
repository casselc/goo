# Independent interface review — reconciliation

Reviewed product commit: `837bce4d94e25feae077f87240b512f3935caebb`.
Status: review only; the findings below are **open**, not implemented fixes.

Three fresh reviewers assessed UX, accessibility and web standards separately,
without authoring the implementation or seeing one another's conclusions.
Browser work was centralized and serial to avoid resource contention. Reviewers
received measurements for their own requested probes and amended their reports.

| Priority | Finding | Evidence and owner report |
| --- | --- | --- |
| Medium | Docked inspection leaves too little width for distinguishing operation names. Repeated full dates and row actions retain space while `/search` becomes `/sea…`. | [UX-1](INDEPENDENT_UX_2026_11_12.md#ux-1--p2-docking-inspection-removes-useful-operation-name-context) — source and actual desktop frame. |
| Medium | Activity service disclosures replace the visible service identity with the identical accessible name “Service actions.” | [A11Y-1](INDEPENDENT_ACCESSIBILITY_2026_11_12.md#a11y-1--medium-activity-service-disclosures-discard-their-visible-label) — live names and source; label-in-name problem. |
| Medium | Services/Hosts/Kubernetes tabs lack associated, named tabpanels. Arrow navigation works, but programmatic relationships are incomplete. | [A11Y-2](INDEPENDENT_ACCESSIBILITY_2026_11_12.md#a11y-2--medium-resource-tabs-lack-programmatic-tabpanel-relationships) — live DOM/accessibility tree and source. |
| Medium | Navigation section text and the enabled local-search placeholder have insufficient normal-text contrast in the default dark theme: **3.871:1** and **3.786:1**, below **4.5:1**. | [Accessibility report](INDEPENDENT_ACCESSIBILITY_2026_11_12.md) — measured computed foreground/background styles and sRGB calculation. |
| Low | A visually hidden table caption gains auto width at the mobile breakpoint and creates unnecessary page overflow. | [Standards finding](INDEPENDENT_WEB_STANDARDS_2026_11_12.md#finding-low-severity--hidden-caption-escapes-the-mobile-table-viewport) — viewport 320px, document 332px; temporary browser-only 1px caption sizing restores 320px. |
| Low | Paused Activity can start with a partial record whose timestamp/service context is clipped. | [UX-2](INDEPENDENT_UX_2026_11_12.md#ux-2--p3-paused-activity-begins-with-a-partial-contextless-record) — desktop/mobile frames and live viewport geometry. |

No critical/high-severity issue was established in the exercised paths. That is
not a full-product conformance or cross-browser certification.

## Checks that held up

- Native field-disclosure Enter/Tab traversal, Plot dialog initial focus and
  Cancel returning to the visible field-menu summary.
- Nonmodal Tab traversal out of the inspector, including a 1440×500 viewport:
  header, scroll body and footer remain reachable without observed occlusion.
- Synthetic composing Escape in the Activity inspector retained input focus
  and kept inspection open. This was not an operating-system IME session.
- Sampled forced-colors/reduced-motion table rendering retained controls and
  boundaries. Charts and other browser engines were not covered by that check.
- Sampled Activity and Hosts screens fit 320px. The populated fixed screen's
  isolated caption defect demonstrates why 390px testing alone was insufficient.

## Recommended remediation sequence

1. Fix shared accessible names, tab/panel relationships and text/placeholder
   color tokens. Add focused DOM/accessibility-tree and contrast regressions.
2. Reconcile the hidden-caption responsive cascade; test fixed, Saved and
   service compositions at 320px without masking page overflow globally.
3. Rebalance columns/actions while inspection is docked. Use similar-prefix
   operation names and long service names to verify readable comparison.
4. Improve paused stream row boundaries while preserving inspection/return
   anchors. Treat Clear Find's scroll reset as an observation requiring a
   deliberate interaction decision, rather than inventing a separate defect.

These recommendations are a proposed follow-up to this review, not another
implemented milestone or an expansion of backend/query scope.

Evidence is retained under `artifacts/independent-ui-review/`: `live.json`,
`narrow.json`, `overflow.json` and corresponding actual browser frames. The
caption isolation changed only a temporary browser DOM style; application
source, telemetry and stored evidence were not modified. No scanner or physical
screen reader was run, and small targets were not classified as failures without
considering visibility and spacing exceptions.
