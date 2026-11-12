# Independent UX remediation review

Date: 2026-11-12. Reviewer did not author the fixes. Scope: UX-1 and UX-2 from
[the original independent review](INDEPENDENT_UX_2026_11_12.md).

This follow-up independently inspects the current source changes and actual
Chromium frames produced by the coordinator's serial real-OTel journey. Browser
execution remains centralized to avoid resource contention; no separate browser
campaign or physical-device study is implied.

## UX-1 — accepted in the exercised layouts

The new [1440px dock frame](../journeys/screenshots/review-remediation-dock-1440.png)
shows `/api/customer/search-products/expanded-details` and
`/api/customer/search-orders/expanded-details` on neighboring rows, with long
qualified service names. The distinguishing operation text is visible through
wrapping rather than ellipsis. The selected row remains in place and its
checkbox, Inspect and native More disclosure remain visible. Repeated UTC dates
are condensed visually while the timestamp renderer retains its full accessible
UTC text and exact tooltip/datetime.

The [1440×500 frame](../journeys/screenshots/review-remediation-short-dock-1440.png)
shows the inspector heading, Close and action footer within the viewport while
its body scrolls. The [390px frame](../journeys/screenshots/review-remediation-dock-390.png)
retains inline inspection with the selected operation, timestamp and qualified
service above it. Long inspector keys wrap more in the narrower dock, but these
frames establish no new control occlusion or missing value.

Source review confirms secondary row actions use the shared native disclosure;
the current visible disclosure summary is the return opener, while inspector
footer actions retain their own opener. This avoids restoring focus to a button
inside the menu after that menu closes. Final execution evidence is recorded in
the remediation storyboard and validation log.

## UX-2 — accepted after a second mobile correction

The initial follow-up run passed its geometry assertions but its actual 390px and
320px Pause frames still showed orphan actions above the next timestamp. The
mobile header is visually clipped while its child TH layout boxes retain their
height; both the initial alignment routine and assertion treated those
nonpainted TH bounds as the readable edge. Independent screenshot review rejected
that result. The implementation now excludes the clipped header's layout boxes;
the browser probe independently checks its clipping and actual viewport edge.

The corrected [390px Pause frame](../journeys/screenshots/review-remediation-pause-390.png)
and [320px Pause frame](../journeys/screenshots/review-remediation-pause-320.png)
start with the first record's UTC and qualified service, followed by its message
and actions. No orphan action row remains above it. The
[desktop frame](../journeys/screenshots/review-remediation-pause-1440.png)
places the first complete record below the painted sticky header.

The final run records these actual readable-edge/first-row-top pairs:

| Width | Readable edge | First row top |
| --- | --- | --- |
| 1440px | 243px | 243px |
| 390px | 368px | 368.125px |
| 320px | 424px | 423.625px |

The subpixel differences are below 1px; the actual frames retain timestamp and
service context. Both ordinary Follow-newest explicit Pause and a partially
visible newest-row case passed. Automatic inspection does not invoke the new
alignment, and held return retains its captured inspection anchor and original
window without additional SQL/API reads. Closing inspection retains the existing
behavior of revealing/centering its opener; exact preinspection scroll retention
on Close is deliberately not claimed. The new alignment is limited to explicit
Pause, rather than continuously snapping an operator's reading position.

## Evidence and acceptance

Both UX findings are accepted in the exercised paths. The coordinator's final
real-OTel journey passed in 58.8 seconds with 42 recorded facts and zero page
errors. The reviewer inspected its source, actual desktop/mobile/short-height
frames, and `artifacts/review-remediation/report.json`; this is not approval based
solely on a green test result. The review itself caught a false-positive mobile
geometry assertion and required correction before acceptance.

Reviewed Go binary SHA-256:
`00ad462fee45c9067010993ae20862ac39cf1c5c4467b0eeee20fd342676cc67`.
The [executable storyboard](../journeys/review-remediation.md) describes the
covered interactions. This is bounded Chromium acceptance of these findings,
not whole-product usability certification or a physical-device study.
