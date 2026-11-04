# Consistent compact count-pair presentation — design 323

Status: bounded shared presentation source and final before/after frames accepted.
Read with the delivered count cohort, fixed query/capture and previous-window
comparison contracts. Repository guidance and current validation/known limits
were rechecked before this review.

## Product outcome

A responder can read any supported count comparison with the same controls and
facts, without learning separate temporary and saved layouts or repeatedly
opening duplicate context blocks. This is consolidation of existing semantics,
not another comparison family, dashboard, source owner or persistence model.

Current temporary held/rest comparisons render summary and groups together with
one set of read-only context lines. Fixed/captured comparisons use a Summary /
Groups selector, repeat windows before/inside the selected mode and again in Scope
and evidence, and expose some clocks only in titles. Those differences obscure
which facts change with the comparison's meaning.

## Reuse the passive comparison seam

Extend the existing ComparisonVisual presentation with a shared count-pair view
and supplied facts. Hosts supply checked EvidenceScope side models, explicit role
labels, existing summary/group data and extra details. The generic renderer must
not interpret raw capture/aggregate DTOs, classify policies, infer a rest relation
or mint source authority. Do not add a new comparison module/model merely to
avoid editing the shared library.

Both callers use one **Summary / Groups** choice rather than temporary comparisons
showing both long renderings at once. Temporary selection is local presentation;
saved fixed/captured presentation continues through its current app-owned change
handler. Keep existing persisted view/default semantics where useful, without
preserving obsolete markup or creating a compatibility layout. No source read,
comparison execution or durable write occurs merely on local mode change.

## One compact operand-facts surface

Always provide a compact mapping from the common Baseline / Selected columns to
the actual operands, each with count, exact half-open UTC event window and readable
UTC source-read clock. Full timestamp precision and independent receipt/basis
facts can remain in one native **Scope and read facts** disclosure. Do not render
two large evidence cards or repeat the same windows above and below a mode picker.

Use only the host's admitted relation:

- Held parent is the earlier supplied browser observation and overlaps selection.
- Rest of parent is the checked temporary complementary query, independently read.
- Previous window is the admitted adjacent preceding equal-duration window.
- Fixed baseline query is a stored question reread explicitly; it is not frozen data.
- Frozen captured baseline is exact stored evidence, with its custody clock and
  receipt link separate from the original source-read clock.

A literal authored title may remain, but it cannot replace this operand mapping.
Do not label a held parent as frozen stored evidence or a manually authored query
as proven rest. Selected/current means the actual selected observation, not a
claim it is live or the most recently available telemetry.

## Preserve denominator and certainty

Count and change retain exact values/units and a common visual scale. They are
full admitted source-population counts, not current inspection-page lengths or
local Find matches. Group pagination/search keeps the supplied full comparison
and denominator; zero groups/values remain distinct from missing/unknown evidence.

Use one concise visible independent-read/coverage fact rather than several copies.
Partial/no-data/unknown/changed states, aggregate-only limitations that change
interpretation and actual warnings remain visible. Detailed aggregate count/group
receipts, no-row-page-inspection explanation, source basis, query text and custody
can live in the one disclosure. Do not call an unacquired row page checked or
replace an unavailable count with zero for a tidier comparison.

Retain current differing-window/selector qualifications. A count change is not a
rate, significance test or causal explanation; captured windows can have different
durations, and parent/selection can overlap. Consolidating prose must not erase
those facts where they affect the decision.

## Keep action and state ownership in the app

Run, Read fixed baseline, Prepare stored capture, Include comparison and Save
remain in their existing application owners with current leases, pending guards
and explicit effects. No acquisition or publication occurs from the passive
renderer's lifecycle, disclosure or Summary/Groups change. Keep temporary-save
and facet-removal consequences visible at their existing action.

Do not add temporary group drilldown behavior merely because the shared group
renderer already offers an interaction slot. Existing durable side drilldowns
retain their actual question/window/policy and renewed return-focus lease.
Consolidation must not reactivate stale controls or reset local mode when held
parents return.

## Before/after acceptance

Use the same real temporary held/rest, saved fixed previous-window and stored
capture count cases at desktop and narrow width. Record actual layouts before
source edits. Afterward, one mode choice and one operand-facts surface must clearly
explain each case without duplicated full cards/prose. Compare actual counts,
roles, event/read/custody clocks, selected rows and source budgets unchanged.

Exercise Summary/Groups, local Find/order/page, disclosure keyboard/focus, pending /
stale controls, saved mode round-trip and existing durable drilldown/Back. Verify
no local action adds source reads or changes scope. Review actual plotted/group
values and zero/unknown states, not just a compact header screenshot. No new
chart family, response cache, evaluator or universal comparison controller belongs
in this finite consolidation.

## Initial actual before/after findings

The actual before temporary-rest frame renders both the full summary and group
table at once, repeating title/read/quality context. The initial after frame has
one Summary/Groups choice, compact rest **three / selected six** facts with exact
UTC window/read clocks, and the unchanged alpha/beta group comparison. The fixed
previous frame similarly maps **five previous / nine selected** with alpha
**two → six**, beta **zero → three**, gamma **three → zero** and the same 0–6
count scale. Core role/clock/denominator clarity is accepted with no semantic or
accessibility blocker found in those frames.

There are concrete nonblocking residues: Selected · Selected in the temporary
heading, a repeated positive coverage line and shared wrapper/notice inheriting
product card chrome. The initial temporary capture is about 899px versus 997px
before, while the fixed capture grows to about 939px versus 735px. Part of that
fixed growth exposes material aggregate warnings; it must not be hidden merely
to claim universal compactness. These element/crop figures are not an all-viewport
performance benchmark.

Root selected one finite final pass: reset only the shared wrapper/notice card
chrome, remove the duplicate Selected prefix and redundant ready temporary
status, and preserve exact facts, material warnings and target sizes. No broader
spacing loop or new layout/model is requested. Final source/frame acceptance
will use that final generation, with native/actor source budgets reported
separately from visual judgment.


## Final 323 source and actual acceptance

The final source uses the existing passive ComparisonVisual countPair/facts seam,
explicit host-supplied EvidenceScope sides and the existing app-owned action /
mode handlers. The neutral owned wrapper and notice no longer inherit nested
product-card padding. It removes Selected · Selected and suppresses only the
known redundant ready temporary status when the actual pair is current; pending,
error and material certainty remain. No source/admission/persistence controller
or temporary drilldown was added.

All four comparative frames are accepted. The final
[temporary rest](../journeys/screenshots/323-after-temporary-rest.png) shows
Baseline · Rest of parent **three records** versus Selected **six records**, both
exact UTC event windows and separate read clocks, one uncertainty fact, one
Summary/Groups choice, the unchanged alpha/beta group counts and one Scope and
read facts disclosure. The selected-only Save and explicit Include consequences
remain app-owned and visible. Compared with its
[before layout](../journeys/screenshots/323-before-temporary-rest.png), it no
longer renders summary and groups simultaneously or repeats the successful
coverage sentence.

The final [fixed previous pair](../journeys/screenshots/323-after-fixed-previous.png)
shows Previous window **five records** versus Selected window **nine**, their
adjacent exact ranges/read clocks and unchanged alpha **two → six**, beta
**zero → three**, gamma **three → zero** on a shared 0–6 count scale. The
[before frame](../journeys/screenshots/323-before-fixed-previous.png) did not expose
all those totals/limits in the same treatment. The two material warnings stay
visible: captured-record uncertainty and independent aggregate reads without
record/field inspection. They were not hidden to manufacture a shorter result.

Final measured DOM heights are **786.22px vs 997.31px** for temporary rest and
**864px vs 734.5px** for the fixed previous pair. The temporary case shrinks;
the fixed case is taller because the shared treatment exposes explicit facts and
material warnings. This is accepted consistency/clarity, **not a blanket
compactness or first-viewport improvement claim**.

The verification owner reports final main held/rest, local display, fixed pair
and drilldown assertions passed with unchanged source budgets; actor completion
and final timings remain in validation. Static frames do not establish zero-I/O,
lease safety or preserved stored bytes by themselves. No browser/product edits
were performed by this reviewer during execution.

No semantic or accessibility blocker remains. No further layout loop, pixel
polish or added feature is requested. Existing snapshot/aggregate/captured meaning
and distinct Save/read owners remain authoritative after this consolidation.
