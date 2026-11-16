# Help without losing your place

Goo's header **Help** opens an inline reading panel. It offers short topics for
what you are viewing and a five-page **First investigation guide**. The panel
starts closed. Reading a guide does not run the application for you.

Executable storyboard: `journeys/contextual-guidance460.spec.js`.
Independent reading driver: `verification/bombadil/guidance460.ts`.

| Scene | Operator action | What remains visible or unchanged |
| --- | --- | --- |
| Live activity | Start ordinary OTLP logs and spans, then open Help. | Both existing streams continue; their connections, scope, loaded rows, Find and table scroll stay unchanged. Help preserves lane width. |
| First investigation | Choose the guide, page forward/backward and Replay. | The page number describes reading position. No app controls are activated and no completed investigation is claimed. |
| Return to work | Close help or Return to work. | Close restores the header Help button; Return to work focuses the workspace outside record viewports, avoiding an accidental stream pause. |
| Retained inspection | Pause explicitly, inspect a log and search its attributes, open and close Help. | The inspection and attribute search remain intact. Reading Help does not reread telemetry. |
| Services | Change route with the guide open, then return to contextual topics. | Reading position survives route changes. Service teaching follows the route; real signal counts and Not observed remain on screen. |
| Fixed question | Run the original fixed span question and open Help. | Question, route and source row identities remain unchanged. An unavailable source stays visibly unavailable while Help is open. |
| Saved evidence | Open a capture and read its help topics. | The stored capture is displayed without rerunning its question. Help distinguishes captured evidence from saved configuration. |
| Narrow and touch screens | Open Help at 320, 390 and 1440 pixels; exercise coarse pointer buttons. | No horizontal page overflow, overlay or hidden controls. Coarse pointer buttons retain 44-pixel targets. |

The real fixture sends 30 logs and 30 spans through the lab OTLP collector. The
runtime forwards successful queries to ClickHouse; only the explicit failure
scene plants an upstream refusal. Per-step receipts distinguish normal polling
on already-active streams from strict zero-SQL Help interactions after Pause.

The independent Bombadil actor has only reading actions: open, start guide,
Next, Previous, Replay, Back to help and close. Its invariants reject API POSTs,
route changes, overflow and claims of completed application work. Its terminal
phase means the reading scenario finished, not that an investigation succeeded.

## Execution

Run against an explicitly built embedded Go runtime with the lab collector and
ClickHouse running:

```sh
GOO_GENERIC_BINARY=/absolute/path/to/goo \
  npx playwright test journeys/contextual-guidance460.spec.js --workers=1
```

Port 18088 is reserved by the isolated journey runtime; run these tests one at a
time.

Executed against the final embedded runtime on 2026-11-16: **passed in 45.8s**.
The 12 local boundaries passed; all five active-stream boundaries preserved both
connections, phases, selectors, nonempty Find, retained keys and table scroll.
Existing polling continued normally. All seven paused/fixed/saved boundaries
issued **zero SQL and zero API POSTs**. The independent Bombadil campaign admitted
339 snapshots and reached phase 8 with no API POSTs, navigation, overflow or
false application-completion claim. Coarse pointer buttons measured 44 pixels.
Runtime cleanup succeeded; no page errors occurred.

Receipt: `artifacts/contextual-guidance460/95cb4640e2b7/report.json`.
Runtime SHA-256:
`134aa8e668f365192bdff70546b575abf95a03d409b1a41e584f317a3cf4b7e5`.

The fixture and real discovery/fixed/capture reads made 27 SQL requests in total,
including one deliberately refused upstream request. These are distinct from the
zero-SQL local Help boundaries.

| Screen | Desktop | Narrow screen |
| --- | --- | --- |
| Activity guide | [1440px](screenshots/460-activity-guide-1440.png) | [390px](screenshots/460-activity-guide-390.png), [320px](screenshots/460-activity-guide-320.png) |
| Services topics | [1440px](screenshots/460-services-help-1440.png) | [390px](screenshots/460-services-help-390.png), [320px](screenshots/460-services-help-320.png) |
| Source refusal with guide | [1440px](screenshots/460-fixed-failure-guide-1440.png) | [390px](screenshots/460-fixed-failure-guide-390.png), [320px](screenshots/460-fixed-failure-guide-320.png) |
| Saved evidence topics | [1440px](screenshots/460-saved-help-1440.png) | [390px](screenshots/460-saved-help-390.png), [320px](screenshots/460-saved-help-320.png) |

The existing held chronology regression also passed against the same runtime
(1.2m): 21 observed steps, 562 independent Bombadil snapshots reaching phase 12,
six source SQL requests, and successful runtime/database cleanup. Native Quint
trace validation passed in 16.07s and rejected nine negative controls, with no
skips. Its receipt is
`artifacts/held-chronology458/0bb6994b7102/report.json`; its existing five
chronology/inspection frames were regenerated.
