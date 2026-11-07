# Retained span-page comparison — design 361

Status: accepted bounded functionality after frozen source, native/model/actor and
actual-frame review. Provenance layout debt is recorded below; no compact-layout
or above-fold result claim is made.

## Task and operand grain

A user pins one captured span page, opens another captured page and deliberately
compares the two supplied observations without rereading telemetry. Reuse the
existing `trace-page-comparison` kernel and passive comparison display. The
recommended operand is the whole current stored page, not a silently selected
trace or the full retained capture. Root has closed this whole-page grain. Mixed traces
remain mixed; no new trace-subset chooser is needed.

Keep the current Table/Span waterfall selector and exact local Inspect from 360.
A compact reference tray belongs beside those existing result controls, not in a
new workspace/sidebar. Suggested actions are **Pin stored page as baseline**,
**Compare with baseline**, and **Clear baseline**. Once pinned, label replacement
**Replace baseline with this page** so a second activation cannot silently change
meaning. Compare is explicit even after either operand changes. Opening a capture,
paging, changing representation and selecting a local occurrence must not compare
or read telemetry automatically.

## Independent evidence, not a joined trace

The baseline description must remain visible while navigating to the selection.
Each operand owns its capture identity, receipt/custody time, environment policy,
frozen question UTC bounds, page offset, supplied rows, retained rows and full
source count. Use compact summaries with existing bounded disclosure for full
identity/limits. Never use one side's clock or policy to caption both. This local
reference is session state only, not a new saved reference resource or relationship
between captures. The owner is one SavedRecords instance across capture/page navigation and
suspension, with explicit replace/clear; no persistence beyond the instance.

The common elapsed scale aligns each supplied page to its own earliest valid
supplied timestamp. It does not align business events or prove common wall time,
a complete trace, causality, unique requests or CPU/self time. Show millisecond
values primarily and preserve exact source authority in existing details.

Reuse kernel match states: repeated labels are ambiguous, mixed/unreported traces
and invalid/duplicate identities can prevent row matching, and one-sided groups
remain one-sided supplied evidence. An absent opposite group is not zero work.
Known zero duration remains distinct from unavailable timing. Inclusive and
outside-supplied-children totals describe supplied rows only, with existing unknown
counts and partial-page caveats. Do not invent a percentage or latency attribution.

## Interaction and ownership acceptance

Pin/replace/clear/compare are source-free, keyboard reachable and explicitly named.
Disabled controls explain prerequisite through the adjacent concise tray state;
no modal or extra chooser is needed. Current-page representation changes preserve
baseline identity and local layout. A comparison must name both fixed operands;
subsequent paging/replacement retires old compare and inspection callbacks rather
than relabelling an old result as current. Baseline data must be detached from
mutable display rows. Back restores the current parent and sensible current
opener focus without stealing focus from another control.

Local inspection must retain the selected side and original captured occurrence,
including duplicate IDs. If the existing comparison display has no local row
inspection, do not add a new generic inspector: retain Table/Waterfall's existing
inspection path. An explicit Open baseline/selection page can reuse existing
stored-page navigation before exact local Inspect using root-verified actual
`{id, cursor, offset}` held at Pin/Compare, not a reconstructed question or URL.
Expand trace remains a separate **fresh read** action and cannot mutate either
stored operand or certify missing relationships.

## Actual validation needed

Use genuine separately captured pages with different policy/clocks/offsets and
limits. Include mixed traces, duplicate IDs, missing/ambiguous timing, zero and a
one-sided group. With telemetry unavailable, pin baseline, navigate capture/page,
compare, switch representation, inspect exact occurrences, replace/clear and Back;
count store GETs separately from zero telemetry reads. Verify stale callbacks are
inert and source/capture bytes unchanged. Actual 1280/390 frames must show readable
operand labels, common scale, explicit one-sided/unknown states and actions within
the viewport; task-scrolled screenshots do not establish entry density.


## Closed applied-pair contract

Root closed an important distinction: the applied comparison is a frozen pair,
independent of the current capture/page candidate. Opening either operand through
existing `storeSideOpen` performs stored metadata/page GETs and changes the local
capture rows; it does not silently replace or recompute the pair. Both fixed
provenance headers remain explicit. A compact **Return to comparison** may reveal
and focus the retained comparison without IO after local Inspect. Keep this a
return gesture, not another navigation surface or implicit Compare.

Pin/Replace baseline, Compare and Clear are deliberate. Replace and Clear retire
the applied pair and its callbacks. Only explicit Compare establishes a new pair.
Current capture selection alone is never evidence that either applied operand
changed. A side-opening action uses verified stored capture ID/cursor/offset held
with that operand; focus returns to the current applicable comparison opener,
without activating stale side links or moving unrelated focus.

## Candidate placement and provenance review

The first candidate placed the comparison after current records and exposed the
independent question/clocks only as JSON. Before source acceptance, move the region
before `recordsPanel` using an ordinary secondary disclosure **Compare stored span
pages**, initially collapsed. Preserve user disclosure state across page/capture
updates. Pin/Compare keep it open; Return to comparison opens it before focusing
and scrolling the retained heading. This keeps the result reachable without a
100-row wall and adds no new workspace.

Each side must show a readable fixed UTC window, source-read time and custody time,
its environment policy and supplied/retained/full-count limits with existing
EvidenceSummary/app facts. Short IDs may have full identity in disclosure. Advanced
JSON remains supplemental; it cannot be the only readable evidence-clock surface.
This is a source acceptance requirement, not an actual-frame claim.

## Final frozen source acceptance

Accepted the final source contract, pending actual native frames. The disclosure
now precedes current records and starts collapsed. Human-readable per-side fixed
UTC window, source-read and custody clocks, page counts, full matching count and
retention facts supplement the independent policy/identity caption. Exact JSON is
secondary. The frozen pair survives current-page navigation; Replace/Clear retire
it, Compare explicitly establishes it, and Return opens the disclosure and focuses
the retained heading without a read. Verified side opening checks the loaded visit,
actual offset and visible current picker; it avoids taking focus from an unrelated
connected control.

The two provenance sections are stacked before the shared comparison display.
That is semantically clear but uses vertical space; source alone cannot establish
paired visual comparability at desktop/narrow widths. It is not a blocker requiring
new layout infrastructure before native review. Actual frames must confirm readable
independent facts, accessible operand actions and the shared comparison's scale and
one-sided/unknown meanings. No source-test or image-pass claim is made here.

## Actual frames and remaining visual evidence

Viewed the three canonical original-detail frames: `361-retained-page-comparison.png`,
`361-retained-page-comparison-390.png` and `361-exact-baseline-occurrence.png`.
Desktop independent Baseline/Selected facts are readable: 20 rows at offset 100
versus 100 at offset 0, full matching/retained counts 120 versus 135, distinct fixed
windows, source-read and custody clocks, and differing policy labels. Narrow facts
wrap within the viewport and the baseline page action remains usable. Exact local
inspection shows the stored gateway record, UTC timestamp, 10 ms duration and
attribute viewer, with fresh expansion explicitly separate.

The native ledger independently passed 18 interaction facts with zero telemetry
API/SQL reads (store GETs counted separately), frozen-pair/return/current-owner
semantics and required observed-model/positive-actor checks. The maintained frames
support provenance and inspection acceptance, but neither comparison frame actually
shows the paired timing bars: desktop reaches only the comparison introduction,
and narrow shows only the baseline card. A further actual task-scrolled chart
frame is needed for shared-scale/one-sided/unknown visual acceptance. This is a
documentation evidence gap, not a production-source blocker or first-entry claim.


## Final timing-frame acceptance and layout debt

Also viewed the genuine settled canonical actor frame
`artifacts/retained-comparison361/actor-terminal.jpeg` at original detail (maintained
copy: `361-retained-page-timing.jpeg`). It shows paired baseline/selected timing
bars and readable gateway 10 to 35 ms, inclusive +25 ms and outside-supplied-children
+17 ms. SQL 4 to 12 ms is distinct; repeated cache labels explicitly refuse row
pairing. Baseline-only old-path and selected-only retry show Not in supplied page
and Unavailable, not zero. This closes actual desktop comparison visual acceptance.
The narrow maintained frame establishes bounded provenance readability only, not a
visible timing chart. Canonical native geometry/interaction checks separately cover
narrow operation; do not label the provenance crop a narrow chart screenshot.

Bounded 361 functionality is accepted. The stacked provenance occupies roughly
900 desktop pixels and substantially more narrow space before the timing chart.
It is truthful and readable, but not a compact comparative layout or polished
results-first experience. Carry compact paired headers with independently disclosed
full facts into the next whole-product checkpoint, rather than expanding this
milestone into another layout project. Keep both operands' clocks/limits accessible
and never consolidate their meaning into one shared receipt.
