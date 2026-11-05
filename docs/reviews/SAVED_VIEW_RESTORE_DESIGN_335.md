# Saved-view restoration design review — 335

Status: final source and actual frames accepted; see the final acceptance below
and [runtime validation](../../VALIDATION.md).

## Purpose and source findings

A saved view is a reusable definition, not stored telemetry. Opening it should
stage its exact questions, typed scope and display settings once; Run is the
separate decision to acquire fresh evidence. A passive browser notification is
not a request to discard the user's current task.

Reviewed the active roadmap, repository guidance, validation/known-issue records,
`ui/population-workspace.mjs`, `ui/saved-workspace.mjs` and the existing retained
comparison/child contract. This is source review and proposed acceptance; no
browser, screenshots or runtime IO measurements were performed by this reviewer.

The population owner currently listens to both `popstate` and `hashchange` and
calls `load(id)` for either matching saved-view hash. Each load clears child and
comparison state, selection and selection history, marks the panel loading and
replaces its contents before fetching metadata. Generation checks help reject an
older response, but do not by themselves prevent a second passive notification
from resetting the task or moving focus when its controls are removed. The saved
workspace wrapper only reveals annotation tools; it is not the owner of this
saved-view acquisition.

Do not infer redundant requests from a campaign-wide total. Measure one real
history traversal and its paired events, metadata GETs, response adoption and
visible state. Keep telemetry requests distinct from definition metadata GETs.

## Interaction contract

Use the existing saved-view links, loading/failure feedback, staged population and
Run control. No new navigation panel, persistent warning, generic state manager
or extra confirmation is needed for this consolidation.

- One visit to one saved-view identity may receive several passive notifications.
  While loading, join the same acquisition; after adoption, leave the current
  panel and local task state intact. Avoid repeated loading flashes.
- Preserve keyboard focus, open disclosures, Find, ordering, paging and applied
  parent evidence when a duplicate notification arrives. Scope edits and draft
  text are local user work, not permission to reload the saved definition.
- Opening a definition remains staged. A metadata response must not start a
  telemetry Run, capture existing observations, or claim frozen evidence.
- A real departure retires the old visit before asynchronous completion can
  adopt results, status, URL or focus. A later genuine revisit to the same saved
  ID is a new visit and may fetch metadata again; this is not a cache promise.
- Back from a supporting-record child remains an in-task return. It restores
  its unchanged supplied parent, local controls and opener focus, rather than
  reopening the saved definition or rereading source data.
- Do not replace typed current/effective environment policy, exact UTC bounds,
  namespace missing/empty distinctions, signal filters or configured operand
  roles with a reconstruction from visible labels.
- Failed loads offer a native **Retry opening view** button only in the failed
  state. This retries the current failed visit’s exact saved ID and performs no
  telemetry read. A duplicate passive event is never automatic retry. Starting
  retry immediately disables/replaces the failed control; departed, replaced or
  disposed buttons remain inert. A new failed attempt may offer a fresh button.
  An explicit keyboard retry may return focus to the current staged Run or fresh
  Retry control; do so only while focus still belongs to that task, never after
  the user has moved elsewhere or departed. Passive events do not move focus.

The repeated UI scope/evidence prose remains a separate density concern. This
work should reduce resets and surprise, not add more explanatory chrome.

## Native storyboard and acceptance

1. Create a valid saved definition through the real UI. Begin from a known
   fixed scope and record its exact questions, composition and display settings.
   Navigate elsewhere, then use browser history to return to the saved route.
2. Record actual `popstate`/`hashchange` notifications for that traversal. Hold
   the first genuine definition GET response and confirm one loading task and
   no telemetry reads while the second notification occurs. Release the response
   and confirm one metadata adoption, exact staged scope and explicit Run.
3. Change an existing local control and focus a meaningful input or action.
   Deliver a same-visit passive notification after settlement. Assert no GET,
   telemetry IO, rerendered control identity, lost text, focus movement or child
   teardown. Synthetic notifications supplement, rather than substitute for,
   the actual browser-history reproduction.
4. Run explicitly and inspect real data. If using a comparison fixture, retain
   the exact supplied operands and the natural roles from 333/334. Open a group
   child, then return; ensure the same parent values, Find/order and focused
   action remain, with no restore GET or source read caused by the return.
5. Delay a genuine saved-view response, depart and release it. The current route,
   panel, status and focus must remain unchanged. Repeat with A→B→A so that
   identical saved IDs alone cannot renew an earlier visit's authority.
6. Revisit the same saved ID after a real departure. Confirm a fresh metadata GET
   and staged opening, with no inherited old response or implicit telemetry Run.
   After a genuine metadata failure, verify **Retry opening view** explicitly
   retries metadata only, while passive notifications and stale retry handles
   perform no request. Confirm failed/loading/staged feedback and usable keyboard
   focus without duplicating normal-state controls.
7. Capture actual desktop staged, retained-task and returned states. Check a
   narrow viewport without hiding controls or scrolling merely to make the
   original state look more compact. The change is accepted on stable task state
   and clear scope/Run/return behavior, not a prettier crop.

Record the exact request/event ledger and positive terminal assertions. A driver
exit status, aggregate request count, or screenshot alone does not establish the
contract. Use bounded property/state tests for event orderings and stale visit
admission where helpful; real native navigation remains responsible for actual
browser event and focus behavior.

## Review status

Source audit accepts a bounded same-visit restoration fix inside the existing
population owner, conditional on a measured reproduction. The failure-state-only native **Retry opening view** control is approved: the
existing text-only failure leaves a same-hash link inert, so deliberate recovery
is preferable to requiring departure/reload. No normal-state chrome or saved
format changes are required. Final source and actual frame acceptance are recorded
below; runtime assertions remain the native owner’s responsibility.

## Final source and frame acceptance

Reviewed the final `ui/population-workspace.mjs` and `ui/exploration.mjs` changes.
The existing population owner now joins same-ID passive notifications for the
same visit; retirement also inspects the observed queued hashchange URL, so a
subsequent same-ID location does not renew an earlier departure. Response identity
is checked before adoption. The restoration callback uses the narrowly scoped
`preservePopulation` option while synchronizing the surrounding question editor,
avoiding invalidation of its own newly admitted population. This does not create
a separate acquisition owner or change explicit Run semantics.

The failure-only native Retry control is accepted. Its callback requires the
current failed visit and connected, visible control. Explicit keyboard retry
focus moves through the current loading panel to the new Run or Retry control
only while focus still belongs to that panel; a user who focuses elsewhere keeps
that focus. Passive notification handling adds no focus action or normal chrome.

Inspected all four actual generated frames:

- `335-staged-definition.png` includes the desktop shell and a naturally focused
  unsaved view-title input. The draft text, current/legacy environment policy,
  both exact window roles and explicit Run remain visible. Preserve is open
  because the title is being edited; this is not the default closed state.
- `335-retained-comparison.png` shows the naturally returned comparison viewport:
  Previous window 13, Selected window 107, `api333` Find, the 9/70 group counts and
  +61 change on a shared scale. This is the viewport after return focus, not a
  claim that all comparison content appears above the fold on initial entry.
- `335-retry-definition.png` is a component-region capture of the failure text and
  native **Retry opening view** button. It demonstrates the bounded recovery UI,
  not whole-page layout or asynchronous focus behavior by itself.
- `335-restored-narrow.png` is the full 390px-wide viewport. The staged questions,
  typed policy, both window intervals and Run/Use fresh controls fit visibly;
  long visit-breadcrumb text is ellipsized. The detailed scope remains available
  in the population body, so the breadcrumb is not its sole accessible source.

No material source or visual design blocker remains. Existing scope/provenance
prose and comparison warnings remain; this milestone does not establish a broad
density improvement. Native request, event, stale-response and keyboard-focus
assertions supply behavior evidence beyond these frames. The native owner reports
completion of its deterministic paths; this reviewer has not independently run a
browser or reviewed a final actor verdict, so actor success is not claimed here.
