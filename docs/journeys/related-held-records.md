# Follow reported context, then return

Executable journey: `journeys/related-held461.spec.js`.

| Scene | Interaction | Expected result |
| --- | --- | --- |
| Ordinary telemetry | Start both lanes with 60 logs and 60 spans. Hold chronology. | Existing source scopes and original read facts remain independent. |
| Inspect | Move to the second page and inspect the child operation. | The selected held occurrence is the return point. |
| Related scope | Choose Related held records. | Two logs and three span observations share the reported trace. Two root observations retain the same reported span ID. |
| Parent context | Read the child’s reported context. | Its parent is ambiguous; the interface does not choose one observation. |
| Same span | Toggle Same reported span. | One log and the child span remain; unchecking restores all five observations. |
| Return | Inspect a different related log, then Return to chronology. | Original page, selected child, inspector, Find, table scroll and route are restored. |
| Existing Find | Find the child operation, drill into its trace, then return. | Related scope bypasses Find locally; returning restores the one-row Find result. |

All interactions after Hold must issue zero SQL and zero API POSTs. Fixtures
are real ordinary OTLP data ingested through the lab collector, without a Goo
manifest. The journey checks responsive overflow at 1440, 390 and 320 pixels.

Run with an explicitly built embedded runtime and the lab source running:

```sh
GOO_GENERIC_BINARY=/absolute/path/to/goo \
 npx playwright test journeys/related-held461.spec.js --workers=1
```

The independent `verification/bombadil/related-held461.ts` driver separately
starts the saved fixture, holds both lanes, inspects the child, toggles its local
span scope and returns. After Hold it rejects API POSTs, bank mutation, navigation
and overflow. Its terminal state means the local exploration returned, not that
a trace was complete or healthy.

## Executed acceptance — 2026-11-16

Final reviewed real-OTLP journey passed in 37.5 seconds with 120 held occurrences.
All three local drill-in/return boundaries issued zero SQL and zero API POSTs. The
independent Bombadil driver reached phase 8 across 388 snapshots, with its held
bank and route unchanged and no local API calls or horizontal overflow. The
complete fixture used six SQL reads, including the independent driver's explicit
source setup. Runtime cleanup succeeded; no browser errors occurred.

Receipt: `artifacts/related-held461/4b47fa3041d3/report.json`.

[Desktop](screenshots/461-related-held-1440.png) ·
[390px](screenshots/461-related-held-390.png) ·
[320px](screenshots/461-related-held-320.png)

The new local subset/return behavior has focused model, generated-bank, owner
and passive-view tests. The existing chronology regression additionally exercises
its native Quint source-authority contract; that model does not yet describe the
related-scope reading state itself.


Before this review follow-up, the existing chronology regression passed on the
461 implementation: 21 steps, 510
Bombadil snapshots and native Quint replay with all nine negative controls
rejected, zero skips. Runtime and isolated database cleanup succeeded.

## Independent polish acceptance

[UX](../reviews/RELATED_HELD_461_UX.md) and
[accessibility/web standards](../reviews/RELATED_HELD_461_ACCESSIBILITY.md)
reviewers accepted the fixes and refreshed responsive frames. The extended
journey also hides the original inspector row with both Find and paging, inspects
a different related occurrence, then restores the original inspector without
changing Find/page. Its fixed anchor remains accessible by description throughout.

[390px forced colors](screenshots/461-related-forced-colors-390.png) shows the
persistent Selected marker and 2px system-color boundary. The runtime records no
browser errors and cleans up successfully. Native disclosure activation and local
source isolation remain covered.

## Consolidation regression (462)

The same storyboard passed after immutable reading-state extraction and shared
parent-context projection: 43.3s, three zero-SQL/API local boundaries, 226
Bombadil snapshots reaching phase 8, bank/route unchanged, no browser errors and
successful cleanup. Receipt: `artifacts/related-held461/10a0702b0295/report.json`.
Refreshed frames retain the independently accepted UI; no new controls were added.

The existing native chronology regression also passed on the consolidated build:
21 steps, 615 Bombadil snapshots and nine rejected Quint negative controls.
Native Hegel order/conservation checks pass over 53 synthetic banks; these are
property fixtures, not OTLP ingestion evidence.
