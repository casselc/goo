# Compare a view with its previous window — design 321

Status: bounded implementation source and all three final actual frames accepted.
Read with the active roadmap and existing saved fixed-query comparison / paired
window contracts.

## Product task and why now

A responder has a useful log/span count view and asks **“Has this changed from the
previous period?”** The older Distribution workflow has a previous-window action;
the canonical saved/composed service/team view lacks that natural continuation.
Its durable count pair currently admits only the same window. Closing that gap
is a general observability outcome, not another metric/facet profile or spacing
pass.

This priority was weighed against two real remaining gaps: local Find in the
held Activity buffer and connecting authored cohorts to the existing explanation
engine. Local Find is useful but only searches loaded rows; it does not answer
full-population time change. Explanation reuse remains valuable, but the simpler
same-question prior-period comparison should be usable first. Existing portable
views, interval families and presentation consolidation remain delivered rather
than new parallel initiatives.

## One existing action and staged execution

Offer **Compare previous window** in the existing count-comparison/population
actions context, not a new comparison lab, top-level toolbar or facet editor.
It constructs the same checked question for the immediately preceding equal-width
window and stages the existing comparison composition without reading. Retain the
selected question, record display/qualified fields and supported chart settings;
retire held results rather than relabel them as new operands. Focus the existing
Run action. Explicit Run obtains both independent observations.

The first cut is an ordinary eligible single-query count view. Facet compositions
and unsupported numeric/status/specialized projections may remain refused with a
visible precise reason. Do not strip measures, silently remove a facet or add
another conversion wizard to force eligibility. Existing deliberate Remove facet
configuration keeps applied predicates; no new promotion action is required here.
Captured and held-parent comparison meaning is unchanged.

## Exact relation and honest labels

Let selected be `[from, to)` with positive checked duration `d = to − from`.
Previous is exactly `[from − d, from)`. Preserve signal, checked environment
policy, complete predicates, grouping, profile/units and declared limits. Reset
both delivery offsets to zero. Refuse a pre-epoch, invalid or over-budget range
before any mutation/read/write. No new wall-clock value is sampled during this
construction.

Admission allows the existing same-window cohort pair or this exact adjacent
preceding equal-duration pair. It does not authorize arbitrary distinct windows,
unequal-duration count comparisons or unvalidated “baseline” joins.

Use **Previous window** and **Selected window** as operand labels when the checked
relation proves it. Show both UTC half-open ranges and actual independent read
clocks, with the captured-count unit and denominator. Count change is meaningful
over equal duration, but it is not automatically a normalized rate or traffic
causality. Unknown sampling/completeness, missing groups and changed reads retain
their existing meaning; absence is not healthy.

A saved question pair describes future explicit reads of fixed questions, not
preserved historical observations. Data can arrive late into either window.
Labels must never suggest the previous operand is an immutable capture or the
same held snapshot that preceded the action.

## Rebase, preserve and return

For the adjacent-previous form, shared range edits/resizing and Use fresh must
recompute the previous window from the newly selected range and its new duration.
Never reuse the same-window helper to collapse both operands onto one range.
Resolve fresh now once; stage both questions and retire both observations.
Same-window pairs keep their existing shared-window behavior.

Save/export/import and original-question reopening use the existing checked view
and portable artifact. Opening stages configuration for explicit Run. Ordinary
Apply to another scope remains refused for fixed pairs until a checked rebinding
operation exists. No new comparison document, capture owner or registry appears.

Back restores the retained parent query/presentation, scroll/focus and paused
Activity where applicable, with no read/resume. Pending/retired actions and
pre-epoch failures must not replace another visit or mutate only one operand.

## Actual-task acceptance

Use real log/span data where the selected period has a visibly different count
from its exact preceding period, with one stable typed scope and predicates.
From the ordinary saved view, stage Compare previous window with zero IO, Run,
and show actual labelled counts/change with common visual scale and both exact
ranges/read clocks. Include zero/no-data and missing-group interpretation rather
than a purely perfect traffic fixture.

Resize/shift the selected window and prove adjacency/width recompute; exercise
Use fresh with one clock and pre-epoch refusal without effects. Save/reopen and
portable validation retain the relation. Existing same-window cohort and captured
comparisons must remain semantically unchanged. Show facet/status ineligibility
without dropping configuration. Update the real storyboard and compact frames;
source/lifecycle assertions separately establish query preservation and effect
budgets. No speculative styling or metric-family work belongs in this round.

## Frozen source and first actual-frame review

The source provides Compare previous window in the existing count-comparison
context with a concise source-free stage/Run explanation and checked availability.
The generated baseline copies the selected predicate and checked scope. Generic
authored pairs still use existing cohort compatibility, so a manually authored
adjacent baseline may deliberately have different predicates; its questions must
be inspected rather than assuming the window relation proves population identity.
The shared rebase helper distinguishes same-window from adjacent-previous and
recomputes both bounds atomically, refusing pre-epoch intervals before mutation.
Source UX review has no blocker.

The first actual staged frame clearly names Previous and Selected UTC windows,
shows the effective environment policy and says Run reads both independently.
The first Groups frame shows **five previous / nine selected** records, with
alpha **two → six**, beta **zero → three**, gamma **three → zero** on a shared
0–6 count scale. Both event windows and distinct actual read clocks are visible;
no rate or significance claim is made. These two frames are provisionally
accepted. Final rebase/fresh and completed native/campaign results are pending;
the first harness stop inside a closed ancestor is not a delivered product
failure or a completed task verdict.


## Final actual 321 acceptance

All three refreshed actual frames are accepted. The
[staged pair](../journeys/screenshots/321-staged-previous-window.png) names the
previous **01:24:05.246–01:25:05.246 UTC** and selected
**01:25:05.246–01:26:05.246 UTC** windows, preserving effective current/legacy
policy and the explicit Run boundary. The
[executed Groups view](../journeys/screenshots/321-previous-window-groups.png)
shows five previous versus nine selected records: alpha two → six, beta zero →
three, gamma three → zero, with exact count changes and a shared 0–6 count scale.
Both independent read clocks are visible. No rate, significance, causality or
upstream completeness claim appears.

The [shifted pair](../journeys/screenshots/321-shifted-previous-window.png) moves
both windows forward ten seconds, to **01:24:15.246–01:25:15.246 UTC** previous
and **01:25:15.246–01:26:15.246 UTC** selected. Adjacency and equal width remain,
and the actual fixture retains the nine/five group counts after explicit reread.
This resolves the earlier provisional frame status. The first harness disclosure /
date-format corrections required no product change and are not reported as
product defects.

The native owner reports all final main assertions passed, including staged zero
IO, Save/export/reopen and rebase query preservation. Actor and final execution
verdict/timing remain with validation; they are not assumed from these frames.
No browser work or product edits were performed by this reviewer during that run.

No source UX or visual blocker remains, and no additional feature, control or
pixel-polish change is requested. This is the bounded previous-period count task,
not general arbitrary-window or full metric-comparison support.
