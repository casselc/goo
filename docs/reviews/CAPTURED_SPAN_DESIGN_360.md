# Inspect a captured span page — design review 360

Status: accepted after final source, native interaction and original-detail actual-frame
review. No 361 comparison, new capture store or template editor is included.

## User task and existing source

A user reopens preserved span evidence and wants its timing representation without
changing the evidence. `ui/saved-records.mjs` currently renders the stored page as
a table; its Expand trace/examples call a fresh context reader. The existing
`record-span-waterfall` kind already projects supplied span pages, retains original
row indices, separates traces and shows incomplete/ambiguous timing honestly.
Reuse it under the stored-page owner, keeping local inspection separate from new
source context.

## Compact representation control

Use the existing View as convention with **Table** and **Span waterfall** for a
captured spans page. Avoid a plain Trace label: one page may contain several traces
or unreported IDs. Keep a short stored-page/no-source-read meaning alongside
existing frozen evidence, UTC, page offset and retained-limit facts. Do not repeat
another large provenance paragraph or add a permanent sidebar.

The toggle is local and synchronous over exactly the current supplied rows. It
changes neither receipt, original fixed question/policy, row membership, page
bounds nor saved bytes. Native saved paging acquires stored pages, not telemetry;
each new page gets its own current rendering/actions. Unsupported log pages keep
their current table rather than an apparently available invalid waterfall.

## Exact occurrence inspection

Inspect selects the original stored page occurrence, not the visual order or an
assumed unique TraceId/SpanId. Duplicate IDs/identical rows remain distinct original
positions. Parent-first layout, grouping and relative axes must not retarget the
inspector, annotation or other existing action to another record.

Local details show the stored record and frozen page provenance without fresh
trace or linked-log reads. Keep explicit **Expand trace (fresh read)** separate:
that existing action may acquire new cross-service spans and linked logs under its
own original-window scope and receipts. A fresh result is not a replacement for
the saved fragment or proof that its attributes/parent ownership never changed.

## Timing and completeness

Group different valid trace IDs separately. Unreported/invalid IDs do not establish
a shared trace; ambiguous/duplicate parents and cycles remain diagnostic. Each
trace's earliest supplied timed span is a local origin, not its actual start, and
shared scale preserves duration comparability without implying simultaneous events.

Missing/invalid timing stays unavailable. Recorded zero duration is measured zero.
A one-page capture, missing parent or retained-row limit is not a complete trace,
critical path, CPU time or unique invocation count. No UI label should upgrade
inclusive supplied timing to those claims. Existing source-changed-between-reads
warnings and unknown coverage travel with the page.

## Keyboard, ownership and return

Toggling keeps focus on the current representation selector. Inspect and close
return to a current visible semantic opener, including when visual row order differs.
Switching representation/page/capture, leaving Saved or replacing the current visit
retires old handlers and local inspectors. Late fresh expansion cannot overwrite
newer preserved evidence; detached controls cannot read or inspect a new page.
Back/reopen remains source-free for stored representation and inspection.

## Acceptance

Use a genuine multi-service captured span page, reopen with source reads counted
(or the source unavailable), switch Table/Span waterfall, inspect the exact stored
occurrence and return. Include multiple traces, duplicate IDs/row-order contrasts,
missing/ambiguous parents, incomplete pages and zero/invalid timing in focused
checks. Verify unchanged question/policy/frozen UTC/receipt/saved bytes and no
telemetry for toggle/Inspect/Back/reopen. A deliberate fresh Expand trace has its
separately advertised source count and preserves the stored parent on return.

Actual desktop/narrow frames should show compact view choice, frozen page limits,
readable timing/diagnostics and local exact-record detail. Task-positioned captures
do not establish fresh-entry density; native ledgers establish source budgets,
original-row identity and ownership independently of appearance.

## Candidate source review

The candidate reuses the passive record kind under the saved-page owner, with a
local Captured spans view selector and explicit stored-page/no-source-read note.
Inspect accessible names refer to original captured occurrence positions, while
Expand trace and trace examples are explicitly marked fresh read. Layout carries
across Table/Span waterfall for the same capture; capture replacement resets it.
Suspension/page/capture revisions retire stale row owners and navigation handlers.
These semantics match the bounded contract.

One concrete keyboard issue needs a small correction before final acceptance:
`onColumns` calls `refreshRows`, which disposes/remounts `goo-record-columns`.
The passive picker's normal open/search/focus preservation works within one element,
so it cannot preserve a focused Show button removed with the entire old picker.
The new picker closes and focus can fall to the document. Capture local open/search
and whether focus belonged to the old picker, then restore the current replacement
and focus its search (the chosen Show action naturally disappears), with the view
selector as fallback. Never move focus when the user already moved elsewhere or
the owner departed. This is bounded current-owner local presentation, not a new
keyboard framework or source effect. Actual frames remain pending.

The root authorized the bounded picker correction. The saved-page callback now
captures local open/search/focus before remount, restores them only for the current
replacement owner and focuses its visible search or summary fallback only when
focus belonged to the removed picker. Unrelated active focus is left alone.
A focused regression case and minimal passive-port harness cover state restoration,
new search focus, outside-focus preservation and departed-owner refusal with zero
source reads. Serial test execution remains with root; this note is implementation
status, not an unrun test-pass claim. Final native picker/frame acceptance is pending.


## Final actual-frame and interaction acceptance

Accepted the final original-detail frames `360-captured-waterfall.png`,
`360-exact-occurrence-390.png` and `360-explicit-fresh-expansion.png`.
The desktop stored-page view clearly labels Span waterfall, no source read,
100 supplied spans at offset 0 and unknown coverage. UTC, readable millisecond
bars and duplicate-ID diagnostics distinguish the supplied occurrences. The
narrow local inspector keeps occurrence 4 separate from adjacent occurrence 5,
with readable service/time/duration, attribute search and explicit fresh-read
action inside the viewport. The fresh Trace inspector is visibly separate,
labels its original fixed bounds and unknown/truncated evidence, and retains
parent ambiguity/duplicate diagnostics. These task-positioned frames establish
readability, not first-entry density or visibility of the full retained count.

The canonical native/actor/model ledger separately proves both captures' full
source counts (1007 scalar and 1014 typed), 1000 retained rows and 100 current-page
rows; none denotes a complete trace. Both nine-step local journeys ran while
source was unavailable with no telemetry API/SQL reads and three store GETs each.
Exact duplicate-occurrence selection, unchanged custody and keyboard picker
open/query/new-search focus passed. The deliberate fresh expansion used two API
reads: one COUNT and one raw LIMIT 100 read, within the 101-row budget. Root's
full UI, generated/property and required Quint checks also passed. The earlier
picker remount defect is resolved; no remaining bounded UX blocker was found.
