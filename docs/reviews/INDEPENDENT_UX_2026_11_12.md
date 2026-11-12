# Independent UX review — 2026-11-12

Scope: commit `837bce4d`, UX only. Reviewer did not author the implementation. Read `AGENTS.md` and `docs/OPERATOR_WORKBENCH_POLISH.md` as claims; independently inspected the seven current operator Activity, inspection, fixed, Services and Saved screenshots and relevant source. Browser execution is owned by the coordinating reviewer to avoid concurrent browser memory pressure; findings below distinguish screenshot observations from source inference. No compliance certification is implied.

## Findings

### UX-1 — P2: docking inspection removes useful operation-name context

**Observed:** In `docs/journeys/screenshots/operator-fixed-inspect-1440.png`, opening the first span's inspector reduces `/search` to `/sea…` and `/catalog` to `/catal…` in the results. Service names wrap across two lines. Meanwhile each row retains the complete repeated UTC date, checkbox, Inspect and Expand trace buttons. The undocked `operator-fixed-1440.png` displays those operations completely.

**Reproduce:** At a 1440px viewport load the storyboard's 30-span fixed question, then Inspect its first row. Compare operation-name visibility before and after opening detail. Similar-prefix operation names would become indistinguishable (impact inferred from the observed truncation).

**Impact:** The inspector keeps records physically present, but weakens the operator's ability to compare the selected span with neighboring operations—the main purpose of the side-by-side layout.

**Source:** `ui/product.css:1777` fixes a 540px inspector and `ui/product.css:1779` reserves its width plus padding in the workspace. `ui/record-presentation.css:33` forces the operation preview onto one ellipsized line; `ui/record-presentation.css:34` keeps actions unwrapped and `ui/record-presentation.css:35` keeps UTC unwrapped.

**Recommended change:** Give operation text a useful minimum width in the docked state and recover space from repeated date text or secondary row actions. Verify with multiple operations sharing a prefix, not only `/search`, `/jobs` and `/catalog`.

### UX-2 — P3: paused Activity begins with a partial, contextless record

**Observed:** `operator-activity-1440.png` has a sticky header followed by a clipped message tail and action buttons whose timestamp and service are outside the visible area. `operator-activity-390.png` starts its record region with ERROR text and actions; the first visible timestamp/service belongs to the next record. The storyboard explicitly acknowledges the first DOM row can lie above the viewport; this explains the geometry but does not resolve the reading ambiguity.

**Reproduce:** Load enough ordinary long-message logs to fill Activity, follow newest records, then Pause. Inspect the top edge of its scroll viewport at 1440px and 390px.

**Impact:** At the first point an operator starts reading, the visible action group/message has no visible time or service. On mobile, where each record uses several lines and table headers are hidden, associating the following timestamp with the preceding message is particularly easy.

**Source:** `ui/tail.mjs:130` sets `viewport.scrollTop=viewport.scrollHeight` after reconciliation when Follow is checked; the viewport is independently bounded by `ui/product.css:1319` and `ui/product.css:1343`. Mobile row metadata is placed first in each record by `ui/record-presentation.css:53` and `ui/record-presentation.css:54`, so clipping the top of that record removes its identifying context.

**Recommended change:** On explicit pause, align the first visible record below the sticky header to a complete row boundary, or render a clear clipped-row continuation boundary. Preserve the selected record and scroll position during inspection/return.

## Workflow observations and limits

The visible hierarchy is materially clearer in the fixed and Saved frames: scope precedes results, row actions are visible, and the inspector exposes the full selected operation and exact timestamp. Services mobile provides observed counts, fixed UTC bounds and an explicit coverage-unknown statement before the catalog. The preservation menu source distinguishes question configuration from evidence and says capture performs an independent read (`ui/preservation.mjs:5`). Activity source renders filter chips outside its disclosure (`ui/stream-entry.mjs:65`, `:105`), so a hidden-active-filter issue is not asserted.

Saved empty catalog and failed-open messages exist in source (`ui/captures.mjs:84`, `:111`); Activity has explicit no-recent-observations/health-unknown copy (`ui/tail.mjs:132`). Those states were not exercised live by this reviewer. Screenshots establish the populated path, not the full error/empty-state experience or usability on a physical touch device. The coordinating reviewer subsequently ran the requested live probes; measurements are recorded in `artifacts/independent-ui-review/live.json` and assessed below. No unobserved failure is counted as a finding.

## Requested live probe results

The coordinating reviewer measured the paused Activity viewport at y207.5–807.5 with scrollTop 3264. Its first intersecting row spans y204–247, followed by y247–290; the sticky column headers occupy the top of the viewport. This confirms UX-2 is a partial visible record, rather than solely a screenshot crop. Retain P3: the impact is a localized reading ambiguity, not lost data or unavailable actions. Entering a nonmatching Find and clearing it resets scrollTop to 0, with the first row fully visible at y243.5–286. This provides recovery but does not remove the initial paused-state issue. The changed scroll position may interrupt chronological reading; this probe alone does not establish an additional defect or justify a separate finding.

At 390px, the Activity inspector heading spans y406.6–438.6; its body spans y467.4–779.4 (312px usable height, 3886px scroll height), and its action footer y785.4–817.4 within an 844px viewport. This is a long nested-scroll attribute list, but the requested measurements establish that heading, body and footer remain visible. No additional mobile inspector finding is raised from these bounds alone.
