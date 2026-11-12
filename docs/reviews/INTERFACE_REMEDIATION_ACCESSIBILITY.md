# Accessibility follow-up acceptance — 2026-11-12

**Disposition: A11Y-1, A11Y-2 and A11Y-3 accepted as resolved.** This is an independent review of the fixes for [the original accessibility findings](INDEPENDENT_ACCESSIBILITY_2026_11_12.md), rather than a whole-product conformance certification. The reviewer did not author the fixes or execute concurrent browser work.

Reviewed the working-tree changes based on `b2932d4035d4da7a82517a3318987bcc5042cc76`, the regression journey, and its final recorded browser evidence. The coordinator ran the browser serially against a freshly built Go application with SHA-256 `00ad462fee45c9067010993ae20862ac39cf1c5c4467b0eeee20fd342676cc67`. Evidence: `artifacts/review-remediation/report.json`; executable journey: [review-remediation.spec.js](../../journeys/review-remediation.spec.js). The final journey passed in 58.8 seconds, recording 42 facts and no browser errors. The reviewer independently inspected the report's relevant values and source changes.

## Acceptance findings

| Original finding | Change and observed result | Disposition |
| --- | --- | --- |
| A11Y-1: service disclosures replace their visible identity | Removed the generic `aria-label`; native summary text supplies the accessible name. The journey asserts exact accessible names for two distinct namespace-qualified services at 1440, 390 and 320 pixels. Recorded examples include `review-platform / platform-customer-catalog-production-europe-west` and its `billing` counterpart. | Resolved |
| A11Y-2: resource tab/panel relationships missing | Each Services, Hosts and Kubernetes tab now references its panel through `aria-controls`; each panel has `role="tabpanel"` and `aria-labelledby` referencing its tab. The journey exercises ArrowRight, ArrowLeft, Home and End, checking focus, selection, panel visibility and reciprocal relationships. Chromium's accessibility tree exposes a named Services tabpanel and its controlling tab. Switching these tabs produced no additional source reads. | Resolved |
| A11Y-3: navigation label and enabled placeholder below 4.5:1 | Both now use the shared muted text token; placeholders have explicit opacity 1. Actual computed contrast is 8.521:1 for the navigation label and 7.860:1 for the enabled search placeholder. | Resolved |

Measured foreground for both text states is `rgb(164,175,191)`. Actual backgrounds are `rgb(13,17,24)` for navigation and `rgb(21,26,35)` for the placeholder. These exceed the 4.5:1 normal-text threshold in the exercised default dark theme.

The dock's compact time presentation retains the full UTC text for assistive technology while marking the short visual text `aria-hidden="true"`. The new More disclosure's accessible name includes its visible word “More”. Source inspection found no targeted accessibility regression in these associated changes.

The final delta excludes the mobile visually hidden, one-pixel THEAD from visible sticky-header alignment calculations. This changes viewport geometry only; it does not hide semantic headers or alter their accessible names. Source review and the final 42-fact journey reveal no new targeted accessibility issue, and the service names, tab relationships and measured contrast remain accepted.

## Limits

Acceptance is limited to these three reported defects and the exercised Chromium states. No physical screen reader, speech-input tool, operating-system IME session, cross-browser matrix, comprehensive chart contrast audit or whole-product WCAG assessment was performed. Geometry and accessible-tree probes supplement source review; they do not substitute for assistive-technology usability testing. The original review remains an accurate historical report of the earlier implementation.
