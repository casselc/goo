# Activity render/scope snapshot review — 423

Status: independent source, baseline/candidate performance and public API
contract checks accepted. Broader journey/model execution is recorded separately
in validation. This reviewer
changes no product code and runs no heavy checks.

## Finding to measure

`Tail.mount().snapshot()` copies every retained observation with `structuredClone`
on every call, in addition to copying selector/window metadata. Activity's
`groupable()`, facet current leases and other scope checks call it even though
they do not consume rows. Attribute availability can repeat that work for many
fields in one inspector render. This establishes a copy path, not a measured
user-visible latency claim.

Measure the actual owner path with a full retained buffer and many attributes.
Record retained-row clone counts as deterministic work, and use elapsed time as
context rather than a portable pass/fail threshold. Keep fixture and operation
sequence identical before and after. Do not justify caching by a synthetic metric
alone or change source budgets to produce a faster measurement.

## Baseline measurement acceptance

The reviewer read the actual passed baseline report and its Playwright profiler.
It times synchronous actions in the real Activity owner and observes
`structuredClone` calls; it does not substitute source data or queries. Both
variants retain 100 real records. The normal records total 43,968 JSON bytes;
adding 48 attributes per record produces 364,568 bytes.

| Action | Normal retained-item copies | Wide retained-item copies | Normal / wide observed time |
| --- | ---: | ---: | --- |
| Each first/second/repeated inspection | 5,200 | 24,400 | Median 150 / 824.2 ms across five inspections |
| Contextual facet entry | 2,700 | 12,300 | 69 / 466.2 ms |
| Local selection/search/Clear | 200 per action | 200 per action | Approximately 2.5–4.1 / 5–7.4 ms |
| Stage union | 1,600 | 1,600 | 56 / 81.2 ms |

This confirms material retained-row copying beyond the rows actually needed for
inspection, with amplification as displayed fields increase. It justifies a
small metadata projection. These timings are observations from this executor,
not production percentiles or portable benchmark thresholds. The profiler
counts copies of actual retained-item shapes separately from raw record-value
copies; it does not claim to measure every allocation or source latency. The
identical-query source/SQL guards pass during local actions.

## Minimal boundary if measurement warrants adoption

A fresh detached `scope()` projection can preserve the existing metadata:
`phase`, `staged`, `readWindow`, `connection`, `evicted`, `gap` and `selector`.
It omits `rows`. `snapshot()` can compose that projection with independently
cloned retained rows. There is no cached authority, global registry, new evidence
format or source operation. Mutating either returned object must not mutate Tail.

| Consumer | Data actually needed |
| --- | --- |
| Field/group availability; counted-value and facet current leases | Current phase, selector, completed window and enclosing application owner/epoch |
| Fixed question/group construction, numeric query projection and previous-window intent | Detached selector/window metadata; explicit source execution remains elsewhere |
| Service return selector checks, stop/remember-parent/staged checks | Current metadata and existing return/source generation lease |
| Latency, log-volume context and error triage current providers | Current phase, selector/window and application active/clean state |
| Inspector record context/ancestry | Retained rows as well as metadata |
| Observed-field suggestions in `tailState` | Retained rows as well as metadata |
| Activity explanation held bins and field discovery | Retained rows as well as metadata; keep the full snapshot provider |

Reuse one render snapshot only for facts used to present availability. Every
invocation, Stage and pending completion must still consult fresh current state
and the existing owner/source-generation lease. A detached render projection
must never become permission for a stale click, including an ABA return to the
same selector text. Construction-time availability must not inspect an
unmounted inspector; the 421 first/second-inspection regression remains relevant.

## Final source acceptance

The detached `scopeSnapshot()` implementation preserves all original metadata,
creates a new read-window object and deeply copies the selector on every call.
Its other fields are primitives. `snapshot()` composes that fresh projection
with fresh retained-row copies. There is no memoized permission or shared mutable
selector reference.

The Activity adoption matches the consumer classification above. Inspector
record context/ancestry, observed-field suggestions and the Activity explanation
provider still receive full rows. `tailState` now checks its fresh metadata and
dirty-scope guards before obtaining rows for suggestions, so a staged filter
need not copy a buffer whose suggestions are being withdrawn. Source-only
current providers and invocation guards call fresh `scope()`; owner epochs,
source generations, current predicates and native readiness checks remain in
place. Render-time field availability still avoids unmounted inspector state.

This source review found no blocker. Scope-only paths are bounded to this
adoption; it does not claim that every Activity operation becomes row-copy-free.
For example, an opened Activity explanation legitimately consumes held rows.
Final profile comparisons and actual source/focus/conformance checks are separate
acceptance gates, recorded in validation.

## Candidate and public API acceptance

The actual candidate report passes on the same two 100-record fixtures with the
same 43,968/364,568 JSON-byte sizes. Every first/second/repeated inspection now
copies 200 retained items. All measured non-inspection local operations—including
facet entry, checkbox selection, search, Clear and Stage—copy zero retained
items and preserve their source/SQL counters. These deterministic copy budgets
are accepted.

Observed inspection medians are 61.2 ms normal and 277.1 ms wide, compared with
150/824.2 ms in the baseline. Timings are mixed elsewhere: Stage takes
130.7/117.4 ms in the candidate versus 56/81.2 ms in the baseline, and one normal
selection measurement is also higher. Therefore this review does **not** accept
a claim that every interaction is faster, a global latency percentage, or a
portable wall-clock performance threshold. It accepts removal of the unnecessary
retained-item work on the measured paths.

The separate real-source public API report also passes. Before Start, detached
scope metadata equals full-snapshot metadata and stage/reset/restage make no
source requests. After an explicit Start obtains 12 independently checked source
records, `scope()` makes zero retained-item copies and `snapshot()` makes 12.
Mutating a returned selector, window or full snapshot row cannot change the next
scope/snapshot. Scope omits rows, full snapshots retain them, and the expected
phase/staged/window/connection/gap/eviction fields remain available. There are no
page errors in either accepted report.

The profiled Activity flow keeps the row-consuming explanation panel closed;
zero local retained-item copies are not a claim about every possible open-panel
combination. Actual interaction regression and native replay remain separate
root-owned acceptance checks.

## Acceptance boundaries

Require detached metadata mutation checks, unchanged full-snapshot rows and
metadata, first/second inspection, unfiltered scope, filtered multi-value
selection, hidden/stale native controls and return focus. Preserve the explicit
Load/Apply source and SQL budgets in the actual journeys. Generated properties
and native replay supplement the browser assertions; a copy-count benchmark is
not an accessibility, concurrency or source-completeness proof.

## Post-cycle product checkpoint

Review the complete ordinary-telemetry journey rather than another feature list:
unfiltered stream → inspect a field → full-scope counted OR selection → Apply →
expand/compare → preserve/reopen. Include large attributes, unknown coverage and
missing/empty values in that whole-product pass.

A concrete next design question is explicit field broadening: 422 intentionally
ANDs with the retained predicate, so it cannot discover alternatives excluded by
an already applied field constraint. Such an action must show a new question and
require an explicit read; it must not silently remove predicates. Check existing
paired facets, selected/rest comparisons and scope recall before classifying an
entry/adoption problem as a missing engine or dashboard.
