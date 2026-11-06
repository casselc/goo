# Compare category populations across fixed windows — 338

Executable: [338-category-pair.spec.js](../../journeys/338-category-pair.spec.js).
Actor: [category-pair.ts](../../verification/bombadil/category-pair.ts).

This journey emits ordinary OTLP logs and server spans with standard service,
environment, HTTP method and route attributes. It discovers those environments
through Activity, stages a previous-window comparison and adds the existing
categorical facet. Definitions are saved through real UI controls; no view or
telemetry response is seeded.

1. Run the original pair: 107 selected records and 13 previous records. The
   inspection page holds only 100 selected rows. Explicit **Read values** acquires
   both original full scopes independently, preserving each window and predicate.
   Selected counts are GET 75, POST 27, missing 3 and present-empty 2. Previous
   counts are GET 5, POST 4, DELETE 2, missing 1 and present-empty 1.
   [Category populations](screenshots/338-category-populations.png).
2. The union has five literal categories. DELETE shows selected **0** and previous
   **2**: zero means absent from the validated selected grouping, while upstream
   telemetry coverage remains unknown. Missing keys and present-empty strings
   retain distinct rows. [Narrow view](screenshots/338-category-pair-narrow.png).
3. Include the baseline-only DELETE category. Parent Run is disabled; invoking
   its retained old handler neither reads nor clears the pending choice. **Filter
   both windows** stages a child without reads. Explicit Run observes 0 selected
   and 2 previous records. Back restores both original counted-value observations
   and the supplied pair without source reads.
   [Baseline-only selection](screenshots/338-baseline-only-selection.png).
4. Clear the local choice, Exclude GET, stage and Run: 32 selected / 8 previous.
   Read both category scopes explicitly, Include POST, stage and Run: 27 / 4.
   Whole original predicates, checked policies and exact windows remain intact.
5. Follow the previous-window API service group. The records-only child stages
   the exact previous operand plus the service predicate; one page read returns
   three actual POST records in that window. Back restores pair, Find and focus.
6. Save and export YAML. Both query definitions and facet survive, while counts,
   read receipts and pending choices are not serialized as saved observations.
   Reopen stages configuration without telemetry; explicit Run reads 27 / 4 anew.
   [Reopened pair](screenshots/338-reopened-category-pair.png).
7. A supplemental logs check plants an HTTP failure for only the baseline
   counted-values read. The real selected read still supplies 27; the opposite
   cell remains unknown, not zero. **Retry baseline values** issues exactly one
   read and retains the selected observation and receipts unchanged.
   [Independent failure](screenshots/338-partial-category-read.png).
8. The same supplemental check retains both value observations and local Find
   through a staged child and Back; a reattached retired old Exclude button stays
   inert. Then hold a genuine baseline response after its backend read, hide/show
   the facet locally and verify healthy selected values survive while the retired
   pending baseline becomes unloaded. Read remains enabled; releasing the old
   response cannot publish. Explicit Read reacquires both sides. These supplemental
   checks make seven source requests in total; all intervening local gestures
   make none. Repeat the ordinary scope/count/drill checks for spans.

The observed 21-step logs ledger ends at eleven population/page reads, four
counted-value lookups, five comparisons, one final Save and one YAML encode:
fifteen source HTTP requests. Activity adoption, original Save, fixture-readiness
polling, the supplemental fault/retry and the separate spans task are outside
that relative ledger. Counts are HTTP effects, not physical SQL statement counts.
Each counted-values endpoint itself records two independent count/group reads;
matching partitions do not establish a transactional or upstream-complete snapshot.

The bounded actor opens the genuine UI-saved original logs definition. It reads
both counts and both category partitions, independently verifies the baseline-only
DELETE row, stages its inclusion without reads and explicitly observes 0 / 2.
Admission requires four verified transitions, six source requests including two
category lookups, two comparisons, zero writes, immutable retained parent facts
and no horizontal overflow.

Initial 338 delivery validation passed in 4.2 minutes for logs and spans. The
21-step observed Quint replay passes in 29.934 seconds. Its terminal effects are
11 population/page reads, 4 category lookups, 5 comparisons, 1 Save and 1 encode.
The real failed-side retry, retained healthy observation, pending-owner hide/show,
late-response refusal and explicit reacquisition all pass outside that ledger.

The 35-second actor passes with 461 snapshots through phases 0–7, four verified
transitions, six source requests including two category lookups, two comparisons,
zero writes, immutable retained facts and no horizontal overflow. Its 461 JPEGs
have four distinct hashes; 453, including the terminal capture, repeat a child
loading state despite the applied terminal DOM/API snapshot. That independently
viewed capture does not prove terminal appearance. The five reviewed Playwright
viewport frames supply visual evidence. Raw actor output and request bodies were
pruned after review; compact receipts, the model ledger and one audited capture
sample remain. Demo traffic is restored and all 51 pre-existing stale pending
stream entries are preserved.

This cut does not add partial-category paging, rates, upstream completeness,
frozen-capture filtering or a new query engine.

## Consolidation regression — 339

The same complete logs/spans journey passes on 339 in 3.4 minutes (3.5 runner),
with observed Quint replay in 32.383 seconds and the unchanged 21-step terminal
budget: 11 record reads, four value lookups, five comparisons, one Save and one
encode. Counts remain 107/13 → baseline-only 0/2 → 32/8 → 27/4, reopen 27/4 and three
previous-window POST records. Failure/retry, pending-owner retirement and local
zero-read gestures all pass. The maintained five frames are refreshed for the
compact Scope disclosure; exact window/policy facts remain available by opening it.

The current 35-second actor passes 342 snapshots through phases 0–7, four verified
transitions, six source requests including two lookups, two comparisons and zero writes and immutable parent facts. A new
independent capture audit finds two unique JPEGs across 342: 334, including the
terminal image, repeat a configured/staged child despite the applied 0/2 DOM/API
snapshot. It is an intermediate staged image, not a loading image or terminal
visual proof. Current Playwright frames provide the visual record. Reviewed raw
output is pruned; compact current receipts and one audited sample remain.

The 343 regression passes the complete logs/spans journey in 2.8 minutes, with
observed model replay in 15.41 seconds and zero skips. The 21-step ledger retains
11 record reads plus four value lookups, five comparisons, one Save and one YAML
encode. Both signals keep 107/13 → 0/2 → 32/8 → 27/4, reopen at 27/4 and return
three supporting records. All failed-side, pending-owner and local zero-read
supplements pass; the five maintained frames are refreshed.

Its new 35-second campaign exits successfully after 42.028 seconds wall time:
91 snapshots through phases 0–7, four verified transitions, six source requests
including two lookups, two comparisons, no writes, immutable parent facts and no
overflow. Six unique JPEG hashes occur; 79 images repeat the correct settled
terminal state, an applied 0/2 child whose facet values remain explicitly unloaded.
This terminal image agrees with the independent DOM/API verdict. The earlier
339 capture limitation above is historical, rather than a claim about this run.
The bounded model helper now allows 120 seconds for cold compilation and replay;
its validation requirements and source budget are unchanged.


## Source-refusal regression (344)

The unchanged journey passed against the reviewed 344 server in **2.5 minutes**.
The observed Quint ledger passed in **17.99 seconds**, with no skips. The native
logs/spans counts, fixed predicates, partial-side retry, pending-owner retirement
and stale-control rejection remain covered. The current 35.135-second Bombadil
campaign produced 581 DOM snapshots through phase 7: source 6, lookups 2,
comparisons 2, writes 0, verified 4, immutable true and no overflow.

Driver-image validation is separate: all 581 JPEGs repeat the initial loading
frame and disagree with the settled DOM state. This campaign's driver captures
are stale; they are not visual evidence. The reviewed terminal sample and explicit
audit are retained with compact logical receipts. Maintained Playwright frames
were regenerated from the current server. No rerun was required for that known
driver-capture limitation.
