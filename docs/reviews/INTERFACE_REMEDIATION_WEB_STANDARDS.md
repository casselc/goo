# Web standards follow-up — interface remediation

Reviewed candidate: working-tree changes following `b2932d4035d4da7a82517a3318987bcc5042cc76`, built as `/tmp/goo-review-fixed` (SHA-256 `00ad462fee45c9067010993ae20862ac39cf1c5c4467b0eeee20fd342676cc67`). This is an independent follow-up to [the original review](INDEPENDENT_WEB_STANDARDS_2026_11_12.md); the reviewer did not implement the fixes. Browser execution was centralized with the coordinator to keep memory-heavy work serial.

## Decision

The original hidden-caption overflow finding is accepted as addressed in the inspected source and recorded browser scenarios. The final real-telemetry journey passed in 58.8 seconds. I inspected its final `report.json`: 42 recorded facts and zero application JavaScript errors, including all caption compositions and the More keyboard/read/return checks.

At a 320-pixel viewport, `artifacts/review-remediation/report.json` records these measurements independently for populated fixed Records, reopened Saved evidence, and actual service record observations:

| Composition | Document width | Caption width × height | Position / padding |
| --- | --- | --- | --- |
| Fixed records | 320px | 1 × 1px | absolute / 0px |
| Saved records | 320px | 1 × 1px | absolute / 0px |
| Service observations | 320px | 1 × 1px | absolute / 0px |

The shared caption rule now includes a compact-table selector whose specificity exceeds the conflicting mobile width/padding rules. The original general telemetry-caption selector is retained alongside it, preserving the shared noncompact policy. Native caption nodes remain present and clipped; no root overflow suppression masks the defect. I inspected the actual fixed and populated service-record 320-pixel frames; record content stays inside the viewport, consistent with the measured geometry.

## Additional regression found and corrected during follow-up

The new native More disclosure initially supplied its child action button as the trace/surrounding return-focus origin. The shared disclosure closes after activation, so that button becomes hidden. I raised this before acceptance. The author changed dispatch to admit the actual clicked button while retaining the persistent More summary as the row action's return origin. Inspector footer actions retain their own visible opener.

The recorded browser probe opens More with Enter, reaches Expand trace with Tab, and closes with Escape back to the summary without issuing API POSTs or SQL. Explicit Expand trace then performs its separate source reads; closing the real trace dialog returns focus to More. The inspected report records SQL `3 → 10` and API POSTs `1 → 3` for the explicit read, rather than treating disclosure as source authority. This confirms the corrected trace path; the corresponding surrounding-log path shares the closure but was not independently browser-exercised in this follow-up.

## Other source checks and limits

- More uses ordinary `details`/`summary` and buttons, without an unsupported ARIA menu role or custom composite-menu keyboard promise.
- Services, Hosts and Kubernetes now have reciprocal tab/panel IDs, panel roles and panel labels, using the existing tab-selection and roving-focus controller.
- The final Pause adjustment ignores the clipped one-pixel mobile header layout box when calculating a visible obstruction. It retains native header semantics and changes viewport alignment only; it does not hide headers from assistive technology or add source operations. The final journey retains 42 facts and zero application JavaScript errors.
- No new source operation or subscription was introduced into a Web Component lifecycle.

No further standards blocker was found in this bounded review. This is Chromium source/render evidence, not cross-engine, screen-reader or whole-product certification. The original report remains the historical assessment of the earlier commit.
