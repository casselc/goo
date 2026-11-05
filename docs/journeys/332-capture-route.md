# 332 — Saved capture route custody

Executable: `journeys/332-capture-route.spec.js`.

The journey starts an isolated generic native Go runtime with the shared
`191-generic-runtime` helper and sends 107 uniquely scoped real OTLP logs to the
collector. An independent read of ClickHouse establishes ingestion. In the UI,
Build fixed question → Run → Save capture creates the original receipt; no
receipt is seeded or fabricated.

Before the read-only budget, a second UI Save is held after the native server
commits it. Navigate to Activity and back to Investigations, then release its
response. A test-client GET proves publication succeeded while the chosen route,
original title and status remain unchanged, and the completed original Save dialog
closes without an explicit Cancel. This exercises the draft display
lease independently from publication.

1. Leave the exact capture hash for `#investigations`, then use the ordinary
   catalog Open button. Exactly one full capture GET opens the stored evidence.
2. Leave and revisit the exact hash while holding an actual native server
   response. Dispatch both `popstate` and `hashchange` before releasing it. The
   pending route owns one GET and adopts the returned stored evidence once.
   After loading, repeated events preserve the exact inspector and record row DOM
   and its focused Close button without another GET.
3. Hold another actual server response, then leave for `#investigations` within
   the same workspace. Releasing the response must neither adopt its evidence
   nor restore the capture URL. The existing status remains unchanged.
4. Revisit the exact hash. A fresh GET succeeds. Page the stored records with its
   opaque cursor: offset 100 contains the remaining seven records with the same
   frozen basis. At a 390 px viewport, the evidence fits without page overflow.
5. Check the original native store file and raw GET response byte for byte.
   Opening, returning and paging perform zero source reads and zero writes.

The response gates use `route.fetch()` followed by delayed fulfillment of those
same response bytes. They test ownership timing without replacing product data.
The independent source query is outside the browser request budget. The UI Save
is an intentional source reread and custody write before that budget begins.

Validation: the final native journey passed in 26.0 s against the reviewed
compiled Go binary. Both frames were generated from that run; the narrow frame
shows all seven stored rows at offset 100 without horizontal overflow. Native
runtime cleanup is confirmed. The initial browser launch required execution
outside the sandbox for Chromium local sockets. An earlier native run exposed
the completed Save dialog remaining open; the final product fix closes only
that original draft, and the final journey requires closure without Cancel.
The journey writes `artifacts/capture-route332/summary.json`, native startup/cleanup
receipts. Copied capture stores and raw browser/actor bulk were independently
reviewed by root, then removed; compact main/campaign/native/viewport summaries
and maintained frames remain. Screenshots are generated from
the running native UI:

- `screenshots/332-frozen-capture-route.png`
- `screenshots/332-frozen-capture-narrow.png`

The representative 102 entry journey also passes its main UI checks, native Go
ledger verification and 30-second actor. Its old absolute first-row threshold
predated the loaded-record Find shelf. The reviewed current 390 × 844 contract
checks the first record starts in the upper half, both fixture records and their
Inspect/Select actions fit completely, the empty Find shelf with collapsed
Record fields stays at most 112 px tall, scope/Pause/coverage/loaded-count facts
remain visible, and there is no horizontal overflow. Actual measurements were
390.5 px for the first record, 100 px for the shelf and 550.875 px for the second
record's bottom. This preserves a useful viewport requirement alongside the
current supplied-record controls.

Representative shared campaign admission also passed on real native UI flows:
331 matching-record search/open (45.9 s including its 30-second actor; 449
snapshots, original member visible, one receipt read, zero source reads/writes),
and 295 contextual comparison (1.1 min including its 30-second actor; 470
snapshots, completed baseline 2/current 3, one source read, two comparisons and
zero writes). The 295 actor initially failed because it compared compact button
text with the semantic accessible label; its corrected accessible-name resolver
retains the same bounded completion and source budgets. The first 102 geometry
failure and the failed 295 actor receipts were preserved until root review,
then reduced to concise failure facts during approved cleanup. No product
source changed for those harness corrections.
