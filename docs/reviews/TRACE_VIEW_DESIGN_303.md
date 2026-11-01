# Trace inspector → editable view: design review (303)

Status: source and all three actual browser frames accepted; final native journey passed. Read with
[the feature contract](../TRACE_AS_VIEW.md),
[architecture review](TRACE_VIEW_ARCHITECTURE_303.md) and
[records-only design review](RECORD_ONLY_VIEWS_DESIGN_302.md).

## User outcome

An expanded trace is a useful starting point for ordinary exploration, not only
an item to read or capture. **Open trace as view** lets the user continue with the
existing editable Waterfall, Table/List, fields, paging, Inspect/Keep and Preserve
controls. It creates no second editor or trace-specific saved-object type.

The single command belongs near the inspector summary and trace paging. Its label
names both its subject and destination. Keep it outside reference comparison and
Record a finding: those actions have different purposes. The current placement
adds one ordinary button and no additional explanatory panel or control forest.
Unavailable hosts and pending reads must not offer a working handoff.

## Continuity and deliberate execution

The destination stages the exact reported trace ID with the original fixed bounds
and checked environment policy, starts at offset zero and defaults to the existing
spans-only records Waterfall. Originating service/search filters do not follow the
trace across services. Inspector grouping/summary options and a selected span do
not become artificial restrictions on the new view.

The handoff transfers configuration only. It borrows no inspector rows, counts,
completeness, selected occurrence or evidence custody. Explicit **Run population
view** acquires fresh count and bounded rows. Focus goes to that existing Run
control, making the staged state actionable without another confirmation surface.
The resulting count must continue to distinguish matching population from the
supplied page; a full trace count is not a completeness guarantee.

For a nested population record, the originating population owner retains its
parent and opens an ordinary child. Back returns through that owner and restores
focus to a visible originating control. Activity/root origins use the existing
exploration/return owner. An intentional handoff closes the inspector without
invoking its normal close callback or resuming Activity behind the staged view.
If the destination refuses, the current inspector remains usable and focused.

Each admitted trace read owns a fresh action node. A previous page, closed trace,
reinserted old button or suspended owner cannot redirect the current workflow.
These lifecycle checks protect a predictable interaction; they do not add more
controls for the user.

## Existing findings remain distinct

Record a finding already captures the originating population and expanded trace
through existing independent custody and attaches an explanation. This new
command is the editable presentation/query continuation. Preserve saves its
configuration; it does not silently capture data, publish a finding or convert a
reported span ID into unique record membership.

## Source review

The implemented action uses the existing inspector context and per-read closure,
prepares the ordinary records composition and dispatches through the originating
owner. A successful handoff suppresses the ordinary modal close callback. Refusal
and exception paths restore the current dialog; stale reads cannot restore it.
Nested dispatch uses the retained-child path and root dispatch uses staged
exploration. The existing focus target is Run. No source UX blocker remains.

## Actual browser frame review

The actual [staged frame](../journeys/screenshots/303-trace-view-staged.png)
leads with the staged status and existing Run control. Preserve and View settings
stay collapsed. It explicitly says configuration only and parent paused; no
borrowed result or extra editor appears.

The actual [last-page frame](../journeys/screenshots/303-trace-view-page.png)
distinguishes 103 matching records from three on this page and labels offset 100,
supplied spans and unknown coverage. UTC timestamps, shared duration scale,
qualified record/resource fields and parent-outside-supplied context remain
legible. Waterfall, fields and native Expand/paging controls retain the compact
302 presentation. Neither frame exposes an aggregate control forest or finding
cards. No visual blocker was found in these two frames.

The actual [inspector frame](../journeys/screenshots/303-trace-inspector.png)
places the single handoff beside paging, directly after the summary/limitations
area and before independent reference actions. Record a finding remains a
separate collapsed section. The command is easy to find without another editor
or explanation panel. No visual blocker remains across all three frames.

Final native keyboard/return checks pass for nested and Activity entry.
Verification caught a return-focus mismatch; the frontend corrected restoration
to the original occurrence's Expand action, and the final native rerun must
confirm it. Source inspection or a static frame cannot establish keyboard return.

## Next substantive outcome: 304 whole-flow authoring checkpoint

303 closes a natural entry into the records-only authoring path. The next step
should assess that path in one realistic cross-service incident, rather than add
another specialized display or handoff mechanically:

1. Start with current telemetry, isolate the relevant population and inspect a trace.
2. Continue as a view, choose useful records/chart presentations and compare the
   observations that explain the incident.
3. Save a reusable service/team view, reopen it in another relevant scope and
   understand which configuration and which evidence were preserved.

Review discoverability, number of competing entry points, scope continuity,
configuration vocabulary and service/team contribution as one task. Consolidate
controls where the task exposes duplication. Select the next implementation
outcome from a demonstrated inability to express or reuse the investigation,
using the current grammar and query owners first. A generic WYSIWYG grid,
another chart vendor, or universal template parameters should not precede that
assessment. This follows the current implementation-plan checkpoint and broader
roadmap while producing a user-visible authoring outcome instead of another
round of isolated guard refinements.

Final native validation passes nested, typed-policy root and direct Activity entry,
including exact focus restoration and paused Activity return without tail resume.
Executed checks are recorded in [VALIDATION.md](../../VALIDATION.md).
