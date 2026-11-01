# Records-only saved views — design review 302

2026-11-01. Initial product/source review of checkpoint `f98e05e`. The decisions
below are the accepted 302 scope; implementation and actual browser outcomes are
reviewed separately rather than inferred from this proposal.

## User outcome

Let someone save and reuse an ordinary log search or exact-trace span view without
inventing a grouping or retaining an unwanted aggregate. The user's first question
is still what is happening in this scope, followed by inspect, narrow, Keep and
save. Presentation should not require an unrelated measurement setup step.

301 delivered Table/List/Span waterfall over one bounded record page. The current
flat `Records only` layout can save query/columns but does not own the new kind.
Canonical compositions require an aggregate, while Customize routes ungrouped or
multi-group queries through grouping controls. A records-only composition closes
that gap through existing owners, rather than another editor, query or catalog.

## Entry and controls

Use the current single-signal Customize entry. For raw logs or an exact-trace span
scope it should stage a records-only view with the same complete query definition:
environment policy, predicates, fixed bounds, paging and existing query directives.
Do not add a group, weaken the trace predicate or run telemetry on that gesture.
When a multi-signal parent is involved, selecting one signal remains explicit and
the parent stays retained; do not pretend the signals are identity joined.

Show the one existing records kind selector and qualified-field picker. Table and
List remain available for logs/spans; Span waterfall is available only for spans.
Inspect/Keep, field actions, explicit trace/log context and the pager retain the
same application controller. Kind/field changes remain local and pending records,
drafts or reads keep their current guards.

Record-only settings should omit controls that cannot work in this cut: timeline,
comparison, reverse order and category-bound reusable templates. Avoid a forest of
disabled controls plus explanatory prose. Preserve the existing aggregate-based
composition controls unchanged. Existing scope/filter controls can remain behind
their current disclosure; this is not another first-page toolbar.

Preserve remains the owner for naming and saving the definition. Save is
configuration; Capture is an independent evidence read. Opening a saved definition
stages it for explicit Run, and no selected row identity is saved with the layout.

## Reuse and data semantics

The saved catalog must expose ordinary Apply composition even when a view has no
aggregate declaration. The current early empty-declarations return would hide this
valid shape. Preview should name its record kind/fields and make the destination
scope clear. Apply preserves the destination query, page and held data; it does
not import source service, environment, trace ID, filters or time bounds.

Use the existing count-plus-rows acquisition profile only when admitted. Describe
its actual request/statement behavior after measuring it; do not promise that
removing a renderer removed source work. The matching count and bounded supplied
page keep their distinct roles. Unknown continuation, changed-between-reads,
missing fields, invalid timing and unknown upstream coverage stay explicit.

The new form is exactly one visible records child. Existing aggregate+records
forms remain valid; new timeline/comparison-without-aggregate combinations and
generic record templates are outside this cut. Go, JS and OpenAPI must explicitly
retain the aggregate requirement for current destination-group templates rather
than accidentally inheriting the new record-only admission.

## Acceptance

1. Start with a real ungrouped Activity log scope. Customize, choose List and
   qualified fields, Inspect and Keep an occurrence, then Save without adding a
   grouping or aggregate. Check source-free staging/presentation and exact Keep
   question/basis/index semantics.
2. Open another ungrouped log scope. Inspect and Apply the saved composition with
   no source read, retaining the destination's environment, filters and window.
   Save/reopen stages the destination definition; explicit Run reads it.
3. Use an exact-trace spans scope with absent or already-multiple grouping. Choose
   Span waterfall, inspect/Keep/page and save without changing that grouping.
   Existing explicit Expand trace still has its separate acquisition semantics.
4. Verify span-only refusal, pending selection/read guards, stale reattached
   controls and retained return. Run and source reads remain explicit.
5. Retain grouped aggregate, timeline/comparison and category-template regressions.
   Include direct Go/JS/API/schema refusal of records-only category templates.

Review three compact actual frames showing natural entry, saved-layout reuse and
span timing. Check keyboard names/focus and narrow-width overflow through the
browser journey. Do not use a giant full-page screenshot or mock the product data
for the main acceptance story. Actual results belong in VALIDATION.md.

## Implementation and actual-frame review

Reviewed the current Go/JS admission boundary through its app integration and the
three generated actual frames: `302-record-only-logs.png`,
`302-record-only-reuse.png` and `302-record-only-spans.png`. No design blocker was
found. Final executable keyboard, narrow-width and request accounting remains the
validation lane's responsibility; image inspection is not a screen-reader audit.

The source uses the existing Customize entry to stage a records-only composition
for raw unsupported-aggregate scopes, without opening the grouping prerequisite.
The regular View settings path also exposes Customize records. One existing kind
selector and field picker remain; category-template authoring, timeline/comparison
configuration and component reversal are absent in the one-child mode. Save stays
in Preserve, distinct from Capture. The saved-composition branch now remains
reachable when the aggregate declaration map is empty.

The log frame presents List immediately: UTC/service/message, selected qualified
resource/attribute fields, explicit missing/empty formatting, shared Inspect and
Surrounding logs, then the normal pager. Matching count 103 and supplied page 100
remain separately stated. No blank aggregate chart or grouping form occupies the
entry. The destination reuse frame shows the worker service's own three records
and field values, not the source gateway data; standard source-free Apply ownership
and exact target-query checks are enforced by the executable journey.

The spans frame shows a record-only exact-trace waterfall at offset 100. UTC,
relative common timing scale, parent-outside-supplied state, explicit trace
expansion and original-position selection remain readable. Trace groups use the
301 compact divider styling rather than nested cards. The supplied three rows do
not imply complete trace membership or upstream coverage.

The inverse transition from a page-only read into an aggregate composition does
not pretend that an aggregate was acquired. Source review shows `countCurrent`
requires the admitted aggregate snapshot; a newly configured comparison remains
unavailable with explicit Run guidance until that aspect is actually acquired.
The implementation does not infer a full measurement from displayed records.

One small nonblocking copy suggestion was sent to the frontend owner: ordinary
Apply composition preview should name the actual Table/List/Span waterfall kind,
because kind is the principal payload of a record-only saved layout, rather than
only saying Records and listing its columns. No extra control or new explanation
panel is needed.

## Proposed next outcome — 303

Use the new foundation for **Open trace as view** in the existing trace inspector.
The user task is to continue an expanded trace in the editable record canvas, use
Table/List/Waterfall and qualified fields, then save/reuse that configuration.
This is distinct from the shipped Record a finding flow, which independently
captures the population and expanded trace with a note for durable custody.

Reuse the current admitted trace question and existing Population/Investigation
Trail owner. Keep its checked environment policy, exact trace predicate, fixed
bounds, requested summary/status and existing grouping; do not force grouping or
reapply the originating service filters. A deliberate page-zero delivery reset is
reasonable for opening the view, but it must not broaden the selected question.
Default to Span waterfall without selecting an observation by reported ID.

The bounded first handoff can stage configuration without a read and leave Run
explicit through the existing record-page acquisition profile. Do not add a new
trace endpoint, data store, query owner, capture or generic template parameter.
Adopting an already held trace page is optional only if an existing trusted
ownership handoff can do so cleanly; it should not create a new capability-transfer
framework merely to avoid the normal explicit Run.

The action must belong to the current trace/read lease and retire on page/trace
replacement. Closing the dialog for this deliberate handoff must not invoke the
old callback that returns to Activity or resumes a parent visit. Acceptance should
prove cross-service exact trace scope, source-free staging, explicit Run,
Inspect/Keep/page/save, parent return and refusal of reattached old dialog actions.
Existing finding/capture behavior must remain intact.
