# Trace inspector to records view: architecture review (303)

Status: implementation source review accepted. Native validation is recorded
separately in VALIDATION.md. Read with the 302
records-only architecture review and current validation record. This is an
application navigation feature over existing query and source owners.

## Accepted outcome

A user inspecting a trace can open an editable, saveable records-only span
waterfall. The action stages configuration only. Explicit Run acquires fresh
count plus bounded rows through the existing `/api/explore/page` profile. Opening
must not copy inspector rows, source basis, result token, count or completeness
claims into the destination. Existing trace finding capture remains separate.

Use the existing `RecordResults.traceQuestion` constructor with original fixed
bounds, exact trace ID, spans signal, no selected-span restriction and offset zero.
It preserves typed environment policy and replaces the original population search
predicates deliberately. No grouping, summary or status options are needed for a
raw waterfall. Validate the generated question and ordinary records composition
before changing the source UI. The resulting `goo.composition/1` has one visible
`record-span-waterfall` child, retaining existing UTC/summary and field rules.

## Concrete ownership seam

`ui/record-results.mjs` owns the modal trace inspector and its context/read
generations. `ui/exploration.mjs` mounts it once and supplies tables to nested
population workspaces. The existing population `openChild` already retains and
pauses the parent, stages a child without reading, enforces nesting/pending guards,
and manages Back/focus. Reuse that path for a population-origin trace. For an
Activity/root origin, use the existing exploration staged-view path and its
existing return owner. Do not replace a nested parent with root `.stage()` merely
because the inspector happens to be mounted at the root.

Pass a narrow per-open handoff callback/lease through the existing `openTrace`
options, or an equivalent owner-bound interaction capability. The originating app
must validate its own generation, active child/depth/pending state and opener;
the inspector must validate its own active context and visible control. Avoid a
new global event bus, query executor, source owner or serialized URL protocol.
The payload is a query/presentation intent, never an admitted source snapshot.

## Lifecycle and failure behavior

The action must belong to the current trace context. Retire stale handlers when a
new trace opens, the inspector closes, its owner suspends or a read/context is
replaced. A detached/reinserted previous control must not inherit the new context.
Pending or failed trace expansion must not accidentally expose another trace's
handoff. The feature may require the current trace expansion to succeed; state
that honestly instead of treating its rows as transfer authority.

Closing normally invokes `contextOnClose` (notably Activity return/resume). A
successful handoff must instead invalidate/hide the dialog without firing that
ordinary close callback and resuming the stream behind a staged child. Retain the
existing return owner until Back. Do not focus a hidden modal button when the
child closes: use the visible originating record/trace control, with the existing
app fallback if that node was retired. Prevalidate before hiding, so rejection
leaves the current inspector usable. Recheck the application lease at dispatch;
failed dispatch must not silently strand a closed inspector.

## Acceptance and review checklist

- Activity and nested population origins stage the correct owner without replacing
  held parents, copying snapshots or starting a source read.
- Exact trace selector, environment policy and fixed bounds survive; source page
  offset and selected span are not inherited.
- Explicit Run uses the raw count/rows profile; ordinary Save/reopen and waterfall
  Inspect/Keep retain the 301/302 contracts.
- Back returns to a visible owner with predictable focus; normal close continues
  to use its existing Activity behavior.
- Stale context/opener, reinserted control, owner suspension, nested-depth limit,
  pending Keep/page and failed trace expansion are negative controls.
- Update the executable storyboard and observed lifecycle replay; report actual
  checks separately from this source-level contract.

## Next roadmap checkpoint

After this handoff, prefer a whole-flow grammar-authoring/usability checkpoint to
another isolated profile or safety milestone. Milestones 300–302 deliberately
separate general template parameters and captured numeric/duration/native
comparisons, keep trace acquisition distinct from local presentation, and identify
raw customization/save as an adoption gap rather than another kind. Together the
implemented primitives now support a meaningful end-to-end assessment.

Use one cross-service incident task: scope the stream, isolate a population,
choose/change records and chart presentations, compare relevant observations,
then save and reuse a service/team view. Have design and architecture identify
where users cannot discover or express that workflow, consolidate competing
entry/authoring controls, and select the next substantive missing capability from
observed task failure. Success is a coherent composable authoring experience;
additional parameter/profile work should follow a demonstrated need rather than
continuing the recent sequence mechanically.

## Initial implementation review

The implementation uses the existing per-open inspector options, forwards a
population-record owner callback into its retained-child path, and routes root
origins through staged `openQuestion`. The pure `traceView` constructor creates
an exact-trace offset-zero waterfall without importing inspector data.

Three source findings were sent for resolution before final approval:

- A singleton global action reading mutable current context permits an old action
  node to survive a different trace/read. Use a fresh button/closure per admitted
  trace read, capturing context/read epochs and node identity; retire it on change.
- Callback refusal restores the modal, but callback exceptions must also restore
  the still-current inspector rather than writing an error into a hidden dialog.
- The native staged-view route currently rejects typed `environmentSelector`
  scopes. Either support them through the existing route or state this bounded
  limitation clearly; do not claim support by silently changing environment policy.

## Final source review

All three initial findings are resolved. Each successful trace read replaces the
action with a fresh node and closure capturing context/read epochs, admitted data
identity and current-node identity. Opening, reading and invalidating retire the
previous action. Dispatch refusal and exceptions restore only the still-current
modal and its close callback/focus. The artificial named-environment restriction
was removed; the existing exact-trace constructor preserves typed policy.

The parent population callback rechecks the original record owner, pending work
and depth before opening its existing retained child. Root origins use the
existing staged-view route. Normal inspector close and successful handoff retain
different behavior deliberately: handoff suppresses the ordinary close callback,
while the existing Activity return anchor remains owned by its original flow.
No query executor, source snapshot transfer or additional evidence owner was added.
Architecture source review has no remaining blockers. `git diff --check` passed;
actual native browser/model results remain with the verification owner.

Native Back validation found that record-owner renewal detaches the original
Expand button; the older fallback selected Inspect. The correction captures the
original row occurrence index, signal and action kind before entering the child.
After renewal, Back resolves that occurrence/action within the retained parent
panel. It does not infer the originating row from trace/span identity, search a
global owner or read the source. Source review accepts this correction; the native
focus assertion remains the verification check.
