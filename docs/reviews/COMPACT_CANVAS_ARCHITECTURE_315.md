# Compact canvas ownership review — 315

Status: architecture source review accepted; actual before/after browser metrics
and executed checks are tracked in `VALIDATION.md`. Scope is compact presentation of existing context/navigation and
read facts, plus passive facet density. No universal navigation or scope framework.

## Existing owners and concrete DOM seams

- `ResourceReturn.mount` owns a named resource ancestry return, revision and held
  restore token (`ui/resource-return.mjs:7–24`). Its `.resource-return` contains a
  callback-owned Back, title, environment and fixed-window metadata. That window is
  ancestor context, not necessarily the currently selected population.
- `StreamEntry` owns the paused Activity anchor and stream selector. Its
  `rememberParent`/`returnToParent` (`ui/stream-entry.mjs:84–85`) restore actual
  viewport/inspector/focus without invoking Start. Its outer fixed-tools disclosure
  must remain the boundary between paused Activity and fixed investigation.
- `InvestigationTrail` owns retained visit Back/Forward/history, up to eight views
  (`ui/investigation-trail.mjs:8–14`). Returning restores retained DOM and scroll;
  it is not a synonym for parent population Back.
- `PopulationWorkspace` owns current query/composition, retained child Back and
  local Edit original question. The scope-context, fixed window/evidence note,
  independent-read explanation and record-page facts currently repeat context
  before and around records. The child wrapper's Back is the primary local return;
  its ancestry owners remain separate secondary actions.

Keep those callbacks and state where they are. Compact their presentation with
small owner-local context rows or disclosures; do not copy callbacks into a central
universal Back, flatten several histories or infer a service return from query
text. One active canvas can expose compact ancestry without deleting return paths.

## Dangerous hiding and restoration

`ResourceReturn.show` currently checks owner/revision and `bar.hidden` before
calling onBack, but not hidden ancestors. A blanket ancestor CSS rule can therefore
make it visually absent while an old programmatic button click still triggers the
resource return. If compact presentation hides a callback-owned action, its handler
must also require current visible action/owner state. Prefer a reachable secondary
context disclosure to concealing the only resource return. Do not overwrite held
restore tokens or use CSS visibility as authorization.

Trail Back/Forward intentionally change retained visit and hash; resource Back
returns to its originating panel; Activity return preserves a paused stream. Keep
those meanings visible in names. Deepest child Back remains local and source-free;
never expose an ancestor action as the primary Back while its child is active.
Retain prior focus/scroll and root editor drafts through compact transitions.

## Current facts versus ancestor facts

Keep a compact always-visible current applied environment/window and unknown
coverage/independent-read qualifier. Exact source receipts, warning text and bases
can live in one **Read facts** disclosure owned by the current population adapter.
Describe each current signal and current inspection page separately: paging may
have a different clock/basis from a held aggregate. No ancestor resource window or
stored-capture custody can replace those current facts.

The existing per-entry result/recordResult are the input DTOs. A tiny presentation
adapter may render these facts into a current disclosure, but must not fetch,
re-admit snapshots, subscribe, own a second result cache or mutate evidence. Page
changes must refresh the corresponding facts through the current record owner.
Do not merely move a detached page-facts DOM node into a shared host and leave its
renewal callback targeting the old parent. Preserve captured/fixed/held operand
labels and their independent clocks.

## Passive facet density

The facet component may compact its supplied title, field, current-scope qualifier,
Read/Clear actions and Find/loaded-match count. Retain independent receipts in its
existing disclosure. No changed wire model, hidden source read or new selection
state is justified. Preserve explicit Include/Exclude semantics, missing/empty
identity, unsupported predicate reasons, keyboard focus and readable coarse-pointer
targets. Avoid fixed-height clipping that makes supplied values/actions inaccessible.

## Acceptance and review

Record before/after geometry and interaction facts from the same real flow before
editing. Then verify current canvas shows useful data earlier, every named return
still works with its own callback/held state, and local returns make zero requests.
Test hidden/reinserted ancestor controls, deepest-child navigation, history/Activity
pause, resource hold/restore, page/read-facts refresh, facet Find/focus and keyboard
traversal. Existing read budgets and semantics must remain unchanged; reduced text
or height alone is not proof of task improvement.

## Implementation source review

Resource return keeps literal title and named Back outside its Origin scope
metadata disclosure. The callback now requires actual visibility and absence of
inert ancestry before clearing its owner/hold lease. Trail buttons remain visible;
only their literal visit label ellipsizes, with full text in its title. Their
callbacks/history remain unchanged.

Population window text compresses a repeated date while retaining exact
millisecond bounds, UTC and end-excluded semantics; its tooltip keeps full values.
Product CSS compacts gaps/headings/chrome and hides only generic workspace prose,
not current scope, status, records or named returns. Existing `[hidden]` behavior
remains authoritative. Passive facet markup groups header/actions/Find/loaded-match
count and keeps exact independent receipts in a disclosure; state, owner and source
wire semantics remain unchanged. Keyboard/coarse-pointer checks belong in native
validation rather than inferred from smaller padding.

One visual-scope caveat was reported: a retained population inside closed fixed
tools can still match `:has(.population-open)` on the outer Activity shell. Restrict
outer compact-header rules to actual active/open fixed context so reduced generic
prose does not bleed into another owner. This is not evidence of lost query state
or source facts; no critical certainty/action was hidden in the inspected rules.

## Bounded final context refinement

The retained-child wrapper may group its existing Back with a native **Parent
paused** disclosure holding the same full description and independent-read facts.
An optional ephemeral navigation title can identify the current child heading;
it must not rewrite saved view/display titles or introduce persisted navigation
state. This removes repeated generic headings and long always-open parent prose
without losing its facts or callback owner.

Hide only an ancestor wrapper's direct header when a deeper retained child exists.
Keep the existing nested-child Back command refusal, so visual compaction cannot
skip the active child. No universal Back or navigation manager is required. Final
height targets must be reported from actual after frames, not inferred from these
DOM changes or partial measurements.


## Final source acceptance

The final wrapper header preserves Back's existing callback/nested-owner guard and
moves only the previous parent description into native disclosure. Current child
heading uses an ephemeral navigation title; saved titles and data remain untouched.
Ancestor header hiding is direct-child scoped and the deepest header remains
reachable. No source request or history merge was added.

The outer compact chrome selectors now require the fixed-tools disclosure to be
open, resolving the closed-Activity bleed caveat. Trail selectors target its actual
nested DOM rather than assuming a direct exploration child. Critical current scope,
status, independent read facts and named returns remain. Source review has no
remaining architecture blockers; `git diff --check` passed. Final geometry/focus
and zero-read-return claims require the actual after checks in validation.

The final duplicate-predicate removal is accepted: the recent Primary signal WHERE
text now appears once in the existing Applied/Configured population selectors
disclosure with its Primary label. The visible summary marks filtered scope. Exact
predicates, current UTC bounds, status and unknown-coverage/readiness facts remain;
no query or control behavior changed. Removing this duplicate does not imply that
filter scope is absent or that ancestry defines the current population.
