# Supplied count-pair presentation — architecture 323

Status: final frozen source review accepted; final generated checks pass.

This consolidation follows 321's general-purpose previous-window task and 322's
saved-baseline depth. It removes duplicated count-pair presentation from two real
callers in population workspace, without creating a universal comparison owner.
Measure the temporary selected/rest and saved fixed-pair flows before edits; judge
the result by coherent visible operands, retained local interactions and source
budgets rather than reduced line count alone.

## Concrete duplication and minimal boundary

`renderSelectedComparison` and `renderCountComparison` both create count summary,
optional grouped comparison and operand contexts. The temporary path displays
read clocks without fixed UTC scope, while the fixed path separately assembles
windows, clocks, custody and aggregate receipts. 321 required changes to that
second path independently. Both already consume admitted supplied comparison
summaries and use the same passive measure/group elements.

Extend the existing `ui/comparison-visual.mjs` passive composition surface with a
bounded countPair mount and pure supplied-facts helper. It accepts admitted supplied
comparison and optional groups, literal title/explicit labels and side context
models constructed by the host using the existing
`EvidenceScope.comparisonContextSide`. Optional custody/aggregate receipt details
are constructed by their checked host owner, not interpreted from raw capture or
aggregate DTOs inside the reusable renderer. This serves two real callers while
keeping the framework-neutral shared display portfolio cohesive.

Return or retain summary/group DOM nodes needed by existing host action binding.
The helper does not fetch, publish, import source snapshots, compute arithmetic or
acquire records. It organizes explicit supplied facts; it does not infer quality
or treat a capture ID as authority. Existing EvidenceSummary validates fact shape,
and existing measure/group elements validate their own supplied data.

Derive each visible fixed window/read clock from the corresponding supplied
comparison operand in the host, not from the selected root query. Avoid another
query model just for display. Compact operand role, count, half-open UTC window and
read time remain visible in both callers. Exact source basis, immutable custody
and independent aggregate receipts may live in one local disclosure, while
unknown/partial/diagnostic warnings remain visible. Avoid duplicating full large
evidence-summary cards in front of plotted values; the 317 task baseline still
applies. This extraction adds no new source authority or persisted wire type.

## Ownership and semantics to preserve

- Keep selectedPair/countPair acquisition, serial/controller lifecycle, currentRun,
  fixed baseline read capability and captured/held validation in the workspace.
- Keep display persistence, view/order changes, readiness and group-side action
  currentness in their existing host adapters. Passive elements may emit intent;
  only the application decides whether it is current and persists presentation.
- Preserve cached summary/groups element identities across workspace rerenders;
  otherwise local Find/order/focus and interaction references may be lost. Restore
  facts disclosure open/focus state only for the same supplied pair.
- Retain explicit labels for selected versus held overlapping parent, logical rest,
  fixed baseline, previous window and captured evidence. A fixed query with a
  different predicate is not automatically a complement. Temporary comparison
  still saves only the selected view unless explicitly included.
- Both operand half-open UTC windows remain visible. Exact read clocks and immutable
  custody time stay distinct, including independently acquired aggregate count/group
  receipts. Aggregate-only baselines must never claim inspected rows/fields or
  baseline row-page consistency.
- Preserve actual no-data/diagnostic/unavailable states, upstream coverage unknown,
  absolute count denominators and unequal-duration warnings. No rate or health
  inference follows from a count difference.

## Verification scope

This extraction needs focused renderer/DOM checks and real owner journeys rather
than a new model-based subsystem. Tests should cover both callers, supplied operand
windows (especially previous-window baseline drilldown), all provenance modes,
aggregate receipts, missing groups, diagnostic/no-data states, no telemetry calls,
input nonmutation, literal titles, local view/Find/order/disclosure focus and stale
interaction refusal on replacement/Back. Native comparison marks and UTC labels
should remain visible at the existing desktop/narrow task dimensions.

A new Hegel Go/JS lane is justified only if implementation adds semantic admission
or data projection beyond assembling already-validated supplied facts. Existing
comparison/schema/property tests remain the source of arithmetic and scope truth;
do not duplicate them merely because DOM construction moved to a file.


## Selected facts projection verification

The chosen shared helper projects exact operand count/window/read-clock fields
from an admitted comparison into compact display facts. That projection is a
small meaningful Go/JS boundary, so a targeted generated lane is warranted without
retesting comparison arithmetic. `verification/count_pair_facts_property_test.go`
builds actual Go CompareCohorts DTOs, admits them with the existing browser source
adapter, supplies existing EvidenceScope context models and checks projection,
full-precision read/custody preservation, explicit invalid-context refusal and
input/output ownership. It includes observed, no-data, diagnostic and maximum
uint64 counts under typed policy. Execution remains pending the shared helper.


## Source review findings during extraction

The first implementation discarded EvidenceSummary row format when composing the
shared disclosure, causing unix-ns-range windows to appear as raw epoch numbers.
Reuse of the existing formatter resolves that regression and preserves exact
precision in tooltips. Pure compact-facts projection also needs finite Date bounds
before the renderer calls toISOString; safe integers alone exceed the Date range.
Missing optional read clocks should remain honestly unavailable, not invalidate
otherwise valid counts. A malformed present clock must refuse: the existing generic
snapshot validator checks provenance as text, so it does not itself protect Date
formatting. These are local passive presentation guards, not new query admission.

The initial actual-Go facts property passed 30 cases in 6.902 seconds. Additional
clock/date-range negative cases run against the completed local guards before final
acceptance. No browser, Docker or acquisition mutation is part of this lane.


## Final frozen source acceptance

Both workspace callers now compose the same ComparisonVisual countPair surface.
Temporary mode/order lives only on the existing selectedPair; fixed mode/order
still updates the existing declaration. Cached summary/group/root nodes remain
attached to their existing pair, and host readiness/currentness governs callbacks.
Only a display mode change retires/rebinds fixed group side actions; order changes
preserve the existing interaction owner. Temporary comparisons gain no drilldown
authority. Acquisition, Save/include semantics and all source endpoints are unchanged.

The shared facts model derives count/window/clock from actual supplied operands,
retains exact uint64 strings and checks display-safe bounds. Context models are
explicit host assertions and grant no new source witness. No-data/diagnostic
comparison notices remain visible, aggregate warnings stay outside the disclosure,
and custody/aggregate receipt limitations remain supplied by their original owner.
The formatter/date findings above are resolved in the frozen source. Optional
clock omission becomes a visible unavailable fact; a malformed present clock
refuses safely. Final generated execution follows; native measurements and owner
journeys remain separate evidence in VALIDATION.


Final `TestHegelCountPairSuppliedFacts323` execution passes 30 actual-Go→JS cases
in 17.037 seconds, zero skips. This includes exact max-integer counts and native
clock precision, no-data/diagnostic projection, detached supplied context models,
invalid labels/facts/counts/windows, out-of-Date-range safe integers, malformed
present clocks and honestly unavailable omitted clocks. Earlier pre-completion
negative checks exposed the clock/date guard gaps; no failed partial-source run
is counted as final validation.

Final copy/chrome review accepts the neutral labelled `role=group` root, removal
of the redundant Selected prefix from host-supplied labels and component-owned
notice padding reset. The temporary wrapper suppresses only its exact known
successful status while the matching actual supplied pair is present; pending,
error and material warning text remains. The shared notice still states unknown
coverage. No facts projection, owner, action or source behavior changed, so the
17.037-second generated result remains applicable without an unnecessary rerun.
