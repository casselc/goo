# Reopen a saved definition without replacing the current task — 335

Executable: [335-saved-view-restore.spec.js](../../journeys/335-saved-view-restore.spec.js).
Optional actor: [saved-view-restore.ts](../../verification/bombadil/saved-view-restore.ts).

Ordinary OTLP logs produce two fixed count questions through Activity's existing
**Compare previous window by service** action. The UI saves that configuration;
no definition, receipt, response or evidence is seeded. Opening it stages the
question, while Run explicitly reads fresh telemetry.

1. Reopen the saved definition through a real browser Back traversal. Native
   `popstate` and `hashchange` refer to one saved-ID visit. Hold genuine metadata
   responses at the browser boundary and count requests independently of source
   reads. The unchanged baseline issues two metadata GETs and zero source reads;
   the corrected owner must issue one.
2. Edit the staged view title without saving, focus the input, and deliver
   passive same-ID history notifications. The existing panel, draft and focus
   remain intact. [Staged definition](screenshots/335-staged-definition.png).
3. Run the saved comparison explicitly: its full populations contain 13 previous
   and 107 selected-window records. Show Groups, find `api333`, then deliver the
   same notifications. Preserve the exact supplied pair, local Find and focus.
4. Stage that group's Previous-window child. Notifications preserve the child and
   focused Run action without acquisition; Back restores the original pair,
   Find and group-action focus. [Retained comparison](screenshots/335-retained-comparison.png).
5. Hold a genuine completed saved-definition response, depart within Explore,
   then release it. It cannot adopt into the departed task. Repeat a same-ID
   A → Explore → A sequence: the old visit's response cannot satisfy the new one.
6. Fail one metadata request at the browser transport boundary. Passive events
   do not silently retry. **Retry opening view** belongs to the visible failed
   visit and reads only the definition; successful retry transfers owned focus
   to Run. If the user moves focus while waiting, completion does not steal it.
   A retained retry button becomes inert after success or departure.
   [Failed definition and retry](screenshots/335-retry-definition.png).
7. At 390px, inspect the staged definition and document its actual viewport
   without manual scrolling. [Narrow restoration](screenshots/335-restored-narrow.png).

The bounded actor independently checks staged drafts, full comparison data,
Groups/Find, staged child and retained Back through passive history events. Its
terminal state requires all four observations, two source reads, one comparison,
zero metadata reloads after initial opening, zero writes and an unchanged pair.
Exit status alone does not establish completion.

The corrected baseline audit passed against the preceding deployed source:
10.3-second test, two metadata GETs/two held responses, native saved-hash event
pair, zero source reads and no page errors. One earlier fixture assertion counted
the preceding Explore departure event too; the corrected assertion filters by the
saved hash.

Final deployed journey passed in 1.0 minute: one metadata GET/one held response
for the same traversal, zero source reads during opening, real 107/13 independent
comparison reads, unchanged draft/Find/pair/child and focus under notifications,
late-response/ABA refusal, loader removal on Explore departure, explicit retry and
conditional focus recovery. Across deliberate later visits/retries the main test
observes ten metadata GETs; source reads remain unchanged at seven after Run.
There are no page errors. The actor ran 30 seconds with 287 admitted snapshots,
all phases 0–8, four independently verified transitions, two source reads, one
comparison, zero metadata reloads/writes, unchanged pair and no overflow.
Desktop and 390px frames are actual viewport captures without manual scrolling;
the failure/retry frame is a component-region capture. The wrapper restored
traffic and preserved all 51 pre-existing stale pending stream entries.

This checks one route visit's ownership, not metadata caching, telemetry
completeness, network replay, frozen evidence or cross-tab synchronization.

After coordinator/design review, compact baseline/main/native/campaign summaries
remain under `artifacts/restore335`; reviewed raw actor traces and current browser
reports are removed. The compact event count is explicitly the entire journey's
ledger, separate from the exact single native traversal. The harness copies the
traversal events at capture time so later passive events cannot append to them.
Genuine saved definitions, lab data and stream entries remain intact.
