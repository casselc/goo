# Compact fixed-population controls — 339

Executable: [339-population-chrome.spec.js](../../journeys/339-population-chrome.spec.js).
Actor: [population-chrome.ts](../../verification/bombadil/population-chrome.ts).

This consolidation uses the genuinely UI-saved original logs definition from
[338](338-category-pair.md), backed by ordinary OTLP records still in the lab.
It neither seeds a saved definition nor substitutes returned telemetry. Run 338
first in a fresh lab, or set `GOO_CHROME_VIEW_ID` to its original 107/13 logs view.
A missing fixture fails with an actionable error rather than silently skipping.

```sh
GOO_CHROME_339_BASELINE=1 bash scripts/run_journeys.sh 339-population-chrome.spec.js
GOO_POPULATION_CHROME_BOMBADIL=1 bash scripts/run_journeys.sh 339-population-chrome.spec.js
```

The baseline lane is an explicit pre-change measurement. The final lane remains
runnable without a baseline receipt; numerical before/after claims require that
both receipts identify the same saved definition. Fresh documents are used for
each viewport, since navigating to an unchanged saved-view hash intentionally
retains the already-applied visit.

1. Load the saved definition at 1280×1000 and 390×844. Opening stages configuration
   without telemetry. Run explicitly reads 107 selected and 13 previous records.
   Capture the unmodified natural entry viewport and actual scroll position.
2. Compare the same supplied count-pair component outside application layout and
   inside the nested facet layout. Its selector stays intrinsic, its font and
   height agree, and its independent-population notice remains visible. These
   display-only probes make no source requests.
3. Focus the compact Scope summary and press Enter. The disclosure exposes the
   exact selected/baseline half-open UTC bounds, full checked environment policy
   and original selectors. Close it by keyboard. Both actions issue no read.
   A passive records-layout change preserves the open disclosure for the same
   question; independent comparison and facet receipts stay in their own surfaces.
4. Read both category partitions, Include POST, and choose Filter both windows.
   Staging reads nothing and pending choices disable parent Run. Explicit child
   Run observes 27/4 while preserving both original predicates and windows.
5. Plant one baseline value-read failure. The selected count remains exact 27
   with `data-value=27` and explicit withheld geometry; the failed count remains
   unavailable. Both decorative tracks are hidden. Retry reads only the failed
   side and retains the healthy receipt. Back restores the original held pair,
   category witness and action focus without a source read.
6. A bounded actor independently opens/closes exact Scope facts with zero reads,
   reads both category partitions, stages POST and observes 27/4. It requires four
   verified transitions, six source requests (two category lookups), two derived
   comparisons, zero writes, immutable parent facts and no overflow.

## Baseline observations

The pre-change baseline passes in 10.4 seconds (13.4 runner). Both viewports have
actual scrollY=0. At 1280px, the facet begins at y=575.58 and comparison y=810.11;
the count-pair selector fills its 948px container at height 38. At 390px, the facet
begins at y=669.44 and comparison y=941.16, with a 310×38 selector. Standalone and
nested supplied components both show the same stretched grid-label behavior.
[Before desktop](screenshots/339-before-entry-1280.png) ·
[Before narrow](screenshots/339-before-entry-390.png).

No overflow occurs. The first harness attempt reused an unchanged hash and
correctly retained the applied view; the fresh-document measurement fixes that
harness error before recording these observations.

## Final observations

The final 339 lane passes in 1.1 minutes (1.2 runner), with the explicit baseline
lane intentionally disabled. All four current frames are reviewed separately
from the initial measurement. Both before/after entries have actual scrollY=0 and
the same original selected/baseline query definitions.

| Viewport | Facet y before → after | Comparison y before → after | Selector before → after |
| --- | --- | --- | --- |
| 1280×1000 | 575.58 → 546.80 | 810.11 → 781.33 | 948×38 → 99×37 |
| 390×844 | 669.44 → 637.23 | 941.16 → 908.95 | 310×38 → 99×37 |

The first data surface moves 28.78px earlier on desktop and 32.20px on narrow.
The narrow comparison remains below the fold; this is a bounded consolidation,
not a claim that all evidence fits initially. Supplied standalone and nested
selectors match at 99×37, 12px font, with the passive notice visible at both widths.

[Desktop entry](screenshots/339-entry-1280.png) ·
[Narrow entry](screenshots/339-entry-390.png) ·
[Exact scope disclosure](screenshots/339-exact-scope.png) ·
[Known count with withheld bars](screenshots/339-known-count-withheld-bars.png).

Keyboard Scope disclosure, passive records-layout rerender, common staging and
Back make no source reads; exact UTC/policy facts remain accessible. The partial
selected value remains 27 with withheld geometry and no unavailable styling; the
unknown baseline is separately unavailable. Failed-side Retry reads once and
preserves the healthy receipt. The 35-second actor passes with 383 snapshots,
phases 0–7, scopeVerified=true, four verified transitions, six source requests,
two value lookups, two comparisons, zero writes and no overflow.

Its screenshot stream is not terminal visual proof: seven unique JPEGs occur
among 383 captures, and 371 including the terminal capture repeat an intermediate
child “Reading fixed baseline…” frame despite applied terminal DOM/API checks.
The terminal JPEG was independently viewed; a compact audit/sample is retained.
The four maintained Playwright frames supply visual evidence. Read budgets count
HTTP source operations, not physical SQL statements.


The complete 338 regression passes with its observed model and actor; the affected
321/333 window journeys and updated actors also pass. Their storyboards record
current results separately from historical measurements. Demo traffic is restored,
all 51 pre-existing stale pending stream entries and genuine lab definitions/volumes
remain, and reviewed raw actor/store/report bulk plus the temporary executable are
removed. Compact before/after metrics, effect receipts and audited samples remain.
