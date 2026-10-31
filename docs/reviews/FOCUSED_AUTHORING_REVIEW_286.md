# Focused-signal authoring review — 286

2026-10-31. Product and architecture review of the existing retained-child,
query, display and composition owners. Implementation and live validation were
pending when the initial recommendations below were made. Actual executed checks
belong in [VALIDATION](../../VALIDATION.md).

## Decision

A contextual **Customize** action near the selected signal closes the reachability
problem found in the [flow checkpoint](FLOW_CHECKPOINT_285.md). Record tabs change
visibility; customization intentionally creates a single-query authoring context.
It must preserve the exact chosen root question, with page offset zero, while the
multi-signal parent remains held and paused.

For an ordinary eligible single-field grouping, initialize the existing checked
records-plus-aggregate composition using the current valid display or its count
default. This changes presentation only; it avoids making users discover three
unrelated authoring controls before reaching the feature. Do not transfer source
rows, aggregate inputs, action bindings, capture custody or completeness claims.

For an ungrouped or categorical multi-field query, stage the original question
and expose existing grouping controls. Numeric histograms and grouped-time
profiles remain explicitly unavailable from this ordinary authoring action; their
conversion needs the specialized editor. Adding a service grouping automatically would
violate the promise to retain the chosen question. Grouping remains an explicit
configuration action, and Run remains the explicit acquisition action. Parent
field/value suggestions can assist configuration, but are metadata from the paused
parent, not observations acquired for the child.

## Acceptance and guardrails

- Use the configured root query rather than a currently paged record request;
  preserve environment policy, predicates, bounds, numeric/duration settings and
  other query fields, changing only offset to zero.
- Reuse the retained-child owner and existing display/composition validators.
  Do not add another query, evidence or layout framework.
- Refuse while a page, filter/source draft, facet selection, record Keep workflow,
  baseline or live configuration makes the destination ambiguous. Refuse the
  chosen histogram/grouped-time profile without disabling an ordinary companion.
  Respect the
  existing child-depth limit and check eligibility again at click time.
- Focus, grouping configuration, composition edits, Save and parent return must
  not read telemetry. Only the explicit child Run or existing explicit reads do.
- Returning restores held parent contents and page, without reactivating retired
  record or aggregate actions. Customize should still work again from the current
  root configuration; detached or navigated-away controls must not act.
- Preserve the parent signal choice and independent companion predicate. The
  focused child cannot mutate the parent through shared arrays or objects.
- Keep the entry contextual near signal/record controls and avoid another large
  toolbar above the first rows. The focused child need not repeat Customize.

The next duration-timeline cut remains separate. This milestone makes the existing
single-query grammar reachable; it does not expand its supported measure families.

## Implementation source check

The working implementation reuses `openChild` and stages `recent:false` children,
so the focused context does not repeat the Customize entry. `focusedQuestion`
selects the configured root query and resets only offset. Eligible queries use the
existing checked composition constructor; the fallback reveals grouping and labels
suggestions as held, bounded parent metadata rather than fresh child evidence.

Per-signal focus controls check their current mount, generation, visibility and
entry again before acting. Returning remounts only these configuration controls at
the new generation. It does not rerender or renew old source-action owners. The
pending-draft, selection, page, comparison, live and depth checks reuse existing
owner state. No material source-level blocker was found in this cut.

Final source follow-up confirms both child Back and child close/onExit callbacks
check current owner and connectedness, so a detached older child cannot close a
newer child. Unsupported histogram/grouped-time entry is also refused explicitly.

Live acceptance must still cover grouped authoring, the ungrouped companion path,
parent page/signal retention, repeated focus after return and retired controls.
This review did not run the browser campaign itself.

## Live-found paging defect and follow-up source review

The live journey subsequently found a material rendering defect that the focused
unit checks and initial source review had not caught. During paging, existing
aggregate hosts are temporarily detached while the population is rebuilt. A new
focus-availability refresh could call the existing DOM-reading presentation getter
in that interval. Rendering could then fail after the source response had already
been admitted. Because rendering was inside the page request catch, that exception
could be reported as a page acquisition failure instead of a rendering defect.

The follow-up source check confirms three bounded fixes:

- Focus refresh is deferred through the full render/reattach interval using a
  balanced render-depth `try/finally`, and is skipped after failed rendering.
- Availability reads baseline metadata through a data-only accessor and the
  existing comparison state, rather than querying detached presentation controls.
- The page request catch owns fetch and admission errors only. Successful page
  publication and rendering occur outside it, so a render failure is no longer
  relabeled as unavailable source data.

No new owner framework, source read or schema is introduced. Final browser and
frontend reruns remain necessary; this review verifies the source boundaries and
does not substitute for those checks.

A returned parent is intentionally a paused held view: its old source-action
owners remain retired. Showing a clearer disabled state on any still-visible
retired buttons is a possible feedback improvement, rather than a reason to
reactivate them or read automatically. The new configuration-only Customize
control is remounted separately for repeat use.

## Follow-up execution status

The parent-run frontend log after the rendering fix records 879 passing tests.
The refreshed browser run records journey285 passing in 20.4 seconds and the
focused-authoring journey286 passing in 27.6 seconds. This reviewer read those
results; it did not execute the campaigns. The live-found defect and bounded fixes
above remain part of the review record. No remaining source-level blocker was
identified. Final campaign, deployment and artifact details remain in
[VALIDATION](../../VALIDATION.md).
