# Independent interface review remediation journey

Executable: [review-remediation.spec.js](../../journeys/review-remediation.spec.js).
Scope: the six accepted findings in the [independent review](../reviews/INDEPENDENT_INTERFACE_REVIEW_2026_11_12.md). This is a bounded Chromium regression, not a full accessibility certification.

The journey sends a fixed bank of 48 ordinary OTLP log records and 48 spans through the real collector/exporter, then waits for their real ClickHouse rows. Each Activity screen size gets a fresh bank and selector so its rolling window cannot expire during earlier checks. Services have long names; operation paths share `/api/customer/search-` and differ at `orders` / `products`. The runtime uses the existing isolated generic-runtime helper and a transparent source proxy. No successful telemetry response or browser telemetry data is mocked.

| Frame | User action and objective assertion |
| --- | --- |
| `review-remediation-dock-1440.png` | Run the fixed span question, dock Inspect, and compare both operation paths. DOM Range geometry must place the distinguishing `/api/customer/search-orders` and `/api/customer/search-products` characters inside their clipping boundaries. Retained text or a title alone cannot pass. Full span names remain in the real result and selected inspector. |
| `review-remediation-dock-390.png`, `-320.png` | Inspect long identities at narrow widths. Open a native field menu with Enter, Tab into its actions, close Inspect, and restore opener focus. These gestures issue no API POSTs or SQL. |
| `review-remediation-fixed-320.png`, `saved-320.png`, `services-320.png` | Exercise populated fixed Records, reopened Saved evidence, and the populated service span-observations composition independently. Every visually hidden caption is at most 1×1 CSS pixels; document width is at most viewport width. |
| `review-remediation-short-dock-1440.png` | At 1440×500, focus Close and the final inspector footer action within the viewport. Full UTC datetime/title remain retained. Record More supports native Enter, Tab and Escape, then explicitly expanding a trace performs a separate source read and closing returns focus to its durable summary. |
| Resource tab evidence in `report.json` | Services/Hosts/Kubernetes expose tab-to-panel ID relationships, named tabpanels, selected state and focus movement for ArrowRight, ArrowLeft, Home and End. The actual Chromium accessibility tree contains the named Services tabpanel. |
| `review-remediation-pause-1440.png`, `-390.png`, `-320.png` | Start a fresh real log activity view, explicitly Pause while following newest observations, then resume and Pause eight pixels from the newest edge, and require its first intersecting row to be whole and contain service context. Source SQL/API counts stay unchanged during Pause. Mobile assertions exclude visually clipped headers rather than trusting their child cell boxes. Frames are captured immediately after Pause. Inspection opening preserves the viewport; Close reveals its opener using the existing centering behavior. Analyze/Return preserves the post-opening held viewport and original observation window. |
| Contrast/name evidence in `report.json` | Two distinct Activity service disclosures have exact accessible names matching their namespace-qualified visible identities. Computed normal-theme navigation text and enabled local-search placeholder have contrast of at least 4.5:1 against their actual opaque ancestor background. |

Run serially against an explicitly built binary and the existing collector/ClickHouse lab:

```sh
GOO_GENERIC_BINARY=/tmp/goo-review-fixed npx playwright test journeys/review-remediation.spec.js --workers=1
```

The executable writes `artifacts/review-remediation/report.json`, runtime provenance and source-proxy calls. Documentation frames are captured from the actual running app under `docs/journeys/screenshots/`. The `finally` block preserves partial evidence and stops the isolated runtime on failure. The complete real-browser journey passes; see [validation](../../VALIDATION.md) and the linked independent follow-up reports for the reviewed scope and limits.
