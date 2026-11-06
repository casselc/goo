# Captured span-page waterfall — architecture contract 360

Breadth cut following checkpoint 359. Source and final native/frame acceptance
are complete; the initial reuse contract and source review remain below. See
[validation](../../VALIDATION.md) for actual checks and limits. No comparison 361,
new trace engine, capture format migration or source read is included.

## Existing APIs to reuse

`SavedRecords.mount::showSaved` already calls
`SourceSnapshots.adoptStoredExplore`, which validates a stored page against its
immutable receipt and original question. `RecordPage.fromExploration(data)` then
creates the existing immutable `goo.record-page/1` supplied display model. It
retains raw rich records and nanosecond strings, page offset/limit/continuation,
read clock, basis and unknown upstream coverage. It does not create completeness,
execution or result-token authority. Its ordinary page bound is at most 200 rows;
current stored paging normally delivers 100, independently of up to 1,000 retained
capture inspection records.

The intended application call is the existing `records.table(target, data.rows,
'spans', data.question, data.basis, options)` with
`recordKind:'record-span-waterfall'`, that supplied `recordPage`, existing display
options and a current saved-page predicate. `RecordInteractions.surface` already
mounts `goo-record-span-waterfall`, indexes `recordItems()` by original page
position and binds the normal local record inspector/actions. Do not use the
fresh trace dialog or add a second inspector to achieve this entry.

## Authority and identity

The waterfall is an alternative presentation of **this captured current page**,
including mixed traces. It does not silently acquire other stored pages, remove
other traces or expand all services. Full population count, capture retention count,
current returned count and per-trace supplied fragments have different meanings;
only the supplied page length describes displayed observations.

`RecordSpanWaterfall.waterfallModel` reuses exact timing geometry with a relative
origin per reported trace and one shared duration scale. It preserves duplicate
trace/span-ID occurrences, missing/invalid IDs, missing or ambiguous parents,
cycles, invalid timing and zero duration as reported diagnostics. A partial page
never proves an absent parent did not exist. A supplied extent is not complete
trace latency, exclusive time or CPU. UTC timestamps and raw precision continue
through existing record presentation; do not round absolute nanoseconds through
JavaScript Number before relative subtraction.

Geometry can reorder nodes visually, but application authority remains the
original immutable row occurrence at local page index. `recordItems`/indexed hosts
must map each original position exactly once. The click target is not identified
by SpanId, label or sorted position; duplicate IDs must inspect their different
original values independently. Existing saved selection references use capture ID
and `question.offset + index`, not inferred global trace membership.

## Local inspection and separate fresh context

Clicking a waterfall timing row invokes existing RecordInteractions Inspect with
the exact original record and original page context. It uses the same retained
inspector and attributes as Table; no telemetry acquisition occurs. Passive
Details must not become a fresh trace dialog. Keep Expand trace explicitly
separate and accurately described as a new read in the original fixed bounds;
existing trace expansion may also acquire linked logs, so it cannot serve as the
local Inspect implementation. The stored page's receipt/bytes/read clock stay
unchanged by either display selection or later fresh context.

## Owner and paging lifecycle

Reuse the SavedRecords page/revision owner. Table/Waterfall switches may remount
that page's existing record surface, retiring old mounting handles. Every adapter
callback must also check the captured revision, captured page identity, active
visit and connected/current surface. Paging increments revision immediately,
before its metadata response arrives; old controls cannot act during the pending
page gap. Suspend/dispose, route changes and a later page response must not revive
saved handlers or publish into another capture. No global registry is needed.

Existing capture paging is evidence-store metadata access, not telemetry. Show
which page is being inspected. A page change must adopt the newly checked page,
not append rows into a synthetic trace. Keep logs on their existing display path;
no span waterfall is offered for a log-only page. Metadata or display refusal must
leave the stored receipt intact and must not trigger fresh-source fallback.

## Acceptance

Use genuine captured multi-trace spans with duplicate IDs/different original
attributes, missing parents, zero duration and more than one stored page. Refuse
all telemetry upstreams while reopening, switching Table/Waterfall, inspecting
original occurrences, paging and returning; only explicit Expand may attempt fresh
telemetry. Assert exact question/row order/index/receipt/read-clock/bytes across
switches and local Inspect. Test scalar and typed policies, mixed services,
no-data, invalid/large timing through meaningful supplied-data tests where OTLP
transport cannot represent the case, and stale handlers after page/route changes.
Desktop/narrow frames must visibly distinguish captured-page timing from fresh
context. Reuse existing geometry and data-model property tests plus actual owner
schedules; do not invent complete-trace or comparison model events.

## Candidate source acceptance

The frozen candidate uses `SourceSnapshots.getRecordPage(data)` from the already
admitted saved page, preserving that exact supplied-page identity across kind,
column and retained resume changes. SavedRecords owns Table/Span waterfall and
validated column preferences per capture; a new receipt resets them. The existing
RecordInteractions adapter supplies original-occurrence Inspect and saved
selection references. Current-page, per-render and revision tokens plus disposal
retire old kind/column/row handlers, including during the metadata paging gap.
Kind/column changes are blocked while selection is active or busy.

The root capture owner resumes only the active visible non-busy capture visit.
SavedRecords resumes from the same admitted held page without re-adoption or
source acquisition and cannot resurrect a disposed host. Fresh trace labels are
scoped copy through existing row/example/breakdown actions, with no fresh context
attached to Inspect. Existing local deployment context setup is itself source-free.
The focused tests exercise kind/column retention, exact row occurrence, stale
callbacks, selection/paging, new-capture/log behavior and disposed resume. Source
acceptance has no remaining blocker. Full/native/model/frame validation remains
the parent's separate responsibility.
