# Large facet interaction review — 369

Status: accepted after source review, repeated genuine native measurements and six
original-detail search/paging frames. No source pagination or general virtualizer
is introduced.

## Problem and measured decision

The existing facet searched and mounted every admitted value. A paired 2,000-value
tree contained about 32,025 elements. That node count justified investigation, not
an assumed performance failure. Complete desktop/narrow baselines then demonstrated
repeatable paired clear-search stalls: 620–1,088 ms desktop and 884–897 ms narrow;
selection took 462/490 ms. These crossed the pilot's material interaction budget.

The pilot targets were local input/selection feedback within 200 ms, without
repeatable stalls above 500 ms; usable mount/held return aimed below 1 second.
Source/network time, synchronous work and subsequent-frame timing were recorded
separately. These targets describe the measured environment, not a universal SLA.

## Accepted implementation and semantics

The shared facet searches and projects every admitted literal, then mounts at most
100 local matches. Loaded matches and shown range remain separate. Full counts,
paired maximum/scales, exact identities and independent receipts never derive from
the displayed slice. Missing, present-empty, literal sentinel text and unavailable
opposite-side evidence retain their existing meanings.

Find resets the local page to zero. Same-read return retains query/page and caret;
a changed evidence basis uses the existing reset rule. Search DOM remains stable
during local drawing. Pager focus stays on its current enabled action or the
opposite enabled boundary action. Generation guards retire replaced row/pager
callbacks. Include/Exclude stages an exact filter without reading telemetry.
Existing full-owner focus fallback may use Read values when an old row is absent;
this does not promise preservation of a missing row through arbitrary basis changes.

## Repeated results

The final canonical repeat measured paired desktop/narrow search p95 at 65.4/64 ms,
selection at 135.4/107.9 ms, mount at 176.9/132.2 ms and Return at 166.9/102 ms.
The earlier accepted run measured search at 111.6/65.2 ms and selection at
84.6/163.9 ms. Both runs meet the pilot budget; the variation is retained rather
than selecting the fastest result. Mounting 100 rows uses about 1,629 elements.

All-page traversal, full-set scales, nonterminal caret position 3, exact child
membership and local API reads of zero passed. Root reports the full UI/Go and
40 generated checks, 343 supplied/live and 346 native regressions passed. Native
ledgers establish timings, source budgets and ownership independently of screenshots.

## Actual-frame acceptance and limits

Viewed all six final original-detail frames: ordinary/paired search at 1280 and
390, plus explicit paging at both widths. Unique suffix search visibly finds
1 of 1,000/2,000 loaded values with a distinct Showing 1–1 of 1 matches range.
The complete terminal-selected-0997 suffix and Include/Exclude are readable
together; paired Selected 1/Previous window 0 remain distinguishable on narrow.

Both paging frames separate 2,000 of 2,000 loaded from Showing 1901–2000 of 2000
matches, with Previous enabled and Next disabled. Narrow controls wrap without
clipping their meaning. Exact suffix rows, counts and actions remain readable.
These are task-positioned frames, not proof of compact first-entry layout.

Bounded UX acceptance is final. No universal device/latency guarantee, source
completeness claim or population denominator follows from mounted-row counts.
No further implementation or broader feature queue is required by this review.
