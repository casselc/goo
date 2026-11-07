# Large-facet interaction decision — architecture 369

Status: source, performance and integration accepted. A clean genuine baseline
justified bounded local presentation; no source paging or virtualizer was added.

## Evidence and measurement validity

343 removed duplicate filtering and shortened only visible long labels. 346
removed repeated identical owner publications. Those changes remained useful,
but the current disjoint 2,000-value pair still built 32,025 DOM elements.
`CategoricalFacet.render` constructed every row; `GooFacetList.refresh` scanned
and toggled hidden state on all rows for each Find operation.

369 used actual OTLP-backed 1,000 ordinary/disjoint 2,000 paired values at
1280×1000 and 390×844, with missing, empty, long Unicode and baseline-only values.
Its 384-byte labels fit the comparison body budget; this does not establish that
all individually admitted maximum-length labels fit the combined 1 MiB request.
Served asset hashes bind measurements to the tested production source.

An exploratory run inherited Playwright tracing and was stopped. Large DOM
snapshots plausibly confounded its timing; those partial results are excluded.
Accepted runs explicitly disable tracing and report that fact. Capture-phase DOM
event delivery to the second animation frame measures a paint opportunity, not
GPU completion. Render synchronous CPU and driver/source wall time are separate.
No Playwright actionability/scroll delay is described as component CPU.

Budgets were adopted before clean results: search p95 ≤200 ms, maximum ≤500 ms,
selection ≤200 ms and desirable mount/Return ≤1,000 ms; repeated ≥2,000 ms mounts
would be material. These are session decision goals, not cross-device SLAs.

The clean baseline failed all four cases. Paired desktop/narrow Find p95 was
1,087.8/896.6 ms, selection 461.6/489.5 ms and mount 1,112.8/1,196.1 ms. Broad-result
clear searches dominated. Ordinary search also breached budgets. Caret position
3 became 22 after Stage/Return. These findings, rather than the traced samples
or DOM count alone, justified the selected remedy.

## Implemented contract

`ui/categorical-facet.mjs` retains the complete admitted ordinary/paired model,
projection, exact counts and full-population scale. It searches full literal
model text and mounts at most 100 matches. Existing loaded-match totals remain
full-set totals; a separate shown-range status explains the presentation page.
Previous/Next are local controls, never source cursors.

Find redraws only the row body, retaining the actual input node. It resets the
page; same-read publication and Back retain page, Find and exact caret, including
blur capture. A different read resets local presentation state. Moving pages
resets the local row viewport; disabled final-page controls transfer focus to an
enabled neighbor. The shared input/row semantics remain framework-neutral.

Every row and pager callback captures its row-generation token as well as the
component epoch. Detached/reinserted old actions cannot stage off-page values.
Pager actions additionally require connected, visible, non-inert controls and a
current owner. Current local Find remains available during data-disabled source
work, while inactive-owner callbacks refuse. The change uses one scoped pager
CSS rule and existing shared visibility; no global manager or cache authority.

Null, empty and literal identities remain exact. Pair maxima use the entire
admitted union rather than the current page. Unknown side counts do not become
zero, and receipt/read clocks remain independent of DOM rows. Local interactions
compare all API counters, including comparison and GET/store traffic. Independent
Include versus missing-or-not-equal Exclude clauses are checked over both full
parent operands, followed by exact parent scope/receipt restoration on Back.

## Final results and limits

The first complete final run passed all four cases in 46.3 seconds (runner 48.6).
Paired desktop/narrow search p95 was 111.6/65.2 ms, selection 84.6/163.9 ms, mount
144.6/106.2 ms and Return 124.5/93 ms. A repeat passed in 45.5 seconds (runner 48.3):
search p95 65.4/64 ms, selection 135.4/107.9 ms, mount 176.9/132.2 ms and Return
166.9/102 ms. Report this variation rather than one universal speed multiplier.

The mounted pair contains 100 rows and 1,629 elements. Both final runs retained
full identity/scale across every page, exact child predicates, caret 3 and zero
local API work. Source acquisition itself is not made faster by this change.

Root reports UI 1,269 checks passing in 18.032 seconds, the main Go suite passing
(cmd 7.742 seconds; others cached), Hegel atomic publication 40 cases in 4.02
seconds, two 343-supplied/346-native regressions in 45.1 seconds and the 343 live
journey in approximately 1.3 minutes. Those gates supplement the focused paging,
visibility and stale-action tests; no new bespoke lifecycle model was needed.

The existing source/category/byte limits remain. Paging does not establish
upstream completeness, add source pagination, change query authority or reduce
admitted model size. Larger inputs still require their own bounded admission
contract. After this finite consolidation, the whole-product checkpoint selects
the next user outcome rather than extending a permanent facet optimization queue.
