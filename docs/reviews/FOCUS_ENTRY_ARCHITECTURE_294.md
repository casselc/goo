# Fixed-population customization architecture — 294

Source review, 2026-10-31. Read alongside the
[usability review](RESPONDER_FLOW_294.md). This review does not claim a completed
browser journey; actual execution belongs in VALIDATION and the 294 storyboard.

## Decision

Approve extending the existing **Customize spans/logs** entry to applied fixed
and saved populations. The reviewed production diff removes `!recent` from `focusReason` and
`mountFocusControl`, and synchronizes staged-child guidance with the already
rendered Run button label. It also adds the explicit bounded span-measure
staging described below. Scope derivation, stored definitions and action
lifetime remain under the existing owners.

Previously, Activity could stage its observed fixed question, run it and open a
linked population, but that non-recent parent lacked the Customize action. Using
Show recent activity instead resolves a fresh clock and changes the window. The
new entry connects the existing fixed-query workflow without that substitution.
The label remains beside each signal's Matching spans/logs heading; it is an
explicit signal choice, not a new global authoring command.

## Preserved boundaries

- `focusedQuestion` copies the selected configured root and resets only paging
  offset to zero. Fixed bounds, environment policy, complete predicates, grouping
  and requested measures remain unchanged. The inspected page is not the root.
- Customize opens the same retained child with `recent:false`; it stages
  configuration without a telemetry read. Run remains explicit and independent.
- A supported ordinary one-field grouping uses the existing display/composition
  constructors. Ungrouped or unsupported grouping does not invent a grouping: it
  stages the existing explicit grouping workflow with bounded held suggestions.
  Numeric histogram and grouped-time specializations remain refused.
- Composition parents still do not offer this entry. A saved definition has no
  applied result before Run, so loading it cannot expose an actionable Customize.
- Existing all-query reuse checks remain conservative: pending facets, query or
  display drafts, selected records, page reads, baseline comparison and live
  configuration can block focus. Depth remains bounded. This cut does not loosen
  those checks to make a particular journey pass.
- Opening a child advances the source generation and retires previous controls.
  Back remounts current Customize buttons with a fresh owner/epoch; a retained
  old button stays retired. It does not revive old trace, record or plot actions.
- Dual populations retain both independent parents. Only the explicitly chosen
  signal enters the child; no join, common transaction or completeness guarantee
  follows from aligned windows.

No new endpoint, generalized action registry, duplicate scope helper or test-only
production hook is warranted.

## Explicit span-measure staging

The natural linked planner enriches spans with duration and recorded status.
Captured-count comparison intentionally excludes status queries in both frontend
and Go validation. Its supplied-cohort adapter does not carry the required
top-level status summary. Removing the eligibility check would therefore admit
an inconsistent payload; it is not a harmless presentation change.

Approve the explicit **Measures · spans** editor in the existing scope context.
It is bounded to one ordinary composed span query, excluding numeric, histogram,
reported-edge and grouped-time profiles. Native checkboxes change only the chosen
summary/status directive and reset offset. All predicates, environment policy,
grouping and absolute bounds remain intact. The parent planner, Customize and
stored captures never silently lose their measurement directives.

The handler checks the captured root reference, source generation, connected and
visible disclosure, active parent and composition readiness before mutation.
Pending page/compare reads and live configuration also block it. It validates the
new question and current composition before retiring the held comparison or
calling existing `stageRootQueries`. Removing a duration required by the display,
or enabling status on a configured count comparison, visibly refuses and resets
only that checkbox. The existing query, display and held evidence stay intact.
An unchanged checkbox event is a no-op.

A valid change clears held inputs and stages Run without source IO. A new capture
must be acquired from the resulting query after an explicit Run; existing status
captures remain incompatible and are not rewritten. Back replaces measure
controls with fresh generation-bound nodes, preserving disclosure state while
old handles remain retired. Render-final refresh enables current controls only
after their parent context has connected, including staged children. Source
review found no remaining lifecycle or temporal-dead-zone blocker.

## Acceptance checks

The real unseeded journey should assert the Activity observed window through the
fixed query, linked plan and focused root, including exact predicates and policy.
Customize must add no reads; only the focused Run reads its query. After Back,
parent evidence stays held, old Customize handles refuse and the fresh visible
control can stage the same root again. Trace inspection must reach a completed
result rather than merely opening its dialog. Saving and reopening configuration
must stage it until Run, without automatically retrieving baseline evidence.

The existing focused-question tests already cover policy variants, predicate
cloning, measures, grouping and page-zero derivation. The real browser path is the
important new coverage for this bounded visibility/availability change. Reuse
the existing focused-authoring regression for ungrouped fallback and private
ownership behavior rather than duplicating the implementation in a Go oracle.

The local guidance fix reads `run.textContent` after the button label is chosen,
so a composed child no longer directs users to a nonexistent Run selected
population button. The text is application-owned and inserted with the existing
text-node helper; this adds no action or IO.

Final diff review, including label synchronization and span-measure staging:
no architecture blockers.
Browser results remain separately recorded by the executing journey owner.
