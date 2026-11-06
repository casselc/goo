# Fixed-question read/Edit consolidation — architecture contract 358

Consolidation selected by [355 reconciliation](ROADMAP_RECONCILIATION_355.md).
Source and final native acceptance are complete, including the widget-side
mutation correction below. Actual checks and limits are in
[validation](../../VALIDATION.md). The initial contract and review chronology
remain explicit.
No query engine, editor model, workspace or execution owner is introduced.

## Existing authority

`ui/exploration.mjs` owns fixed questions through `QuestionSession.Session`, the
existing raw filter/numeric/group/shortcut controls, environment draft and source
controller. `inputs` retains `session.bounds` until the explicit fresh-window
operation clears them. `withFilters` assembles the actual complete draft, including
numeric/histogram, duration/status, grouped time and reported-edge definitions.
The existing `change` path invalidates old observations and actions when a real
edit occurs. Presentation must reuse those paths rather than synthesize another
question from a compact summary.

Before 358, Activity analysis used `analysisEdit`/`analysis-editing` to reduce
controls while ordinary fixed entries exposed the full workbench. Consolidate that affordance
for all fixed questions. An origin label can still say Activity or package, but
must not choose another read/Edit lifecycle or clock policy. The population canvas
continues to cover/inert the root editors using its existing enter/exit owner.

## State contract

| Execution state | Read/Edit presentation | Read authority |
| --- | --- | --- |
| Initial without a fixed question | Existing initial entry and editing tools remain available | Existing explicit recent/fixed Run rules |
| Untouched staged fixed question | Compact checked scope, visible Run and Edit; no observed result claim | Original complete staged question and fixed bounds |
| Applied | Substantive chosen results lead; optional authoring is collapsed unless intentionally opened | The applied question, independent receipt and existing result renderers |
| Running | Compact scope/loading and visible Cancel; no publishable observation is implied | Existing current request token/controller |
| Dirty or invalid edit | Editor and diagnostic remain visible; no action silently applies/discards or hides the draft | Existing actual raw controls and validation |
| Cancelled or failed | Current failure/cancellation and explicit Run remain visible; validation errors keep editing visible | Existing fixed bounds/draft, no automatic read |

Opening Edit alone changes presentation only: zero source/metadata reads, no
`Session.edit`, no cleared evidence and no changed timestamp. A real control edit
continues to call existing `change`/validation, clearing stale evidence rather than
presenting it as the draft's result. An untouched stage has `phase=draft` and a
`requested` question; an actual edit clears `requested`. Do not treat both as the
same dirty state merely because their phase string matches.

The smallest safe closing policy is to keep dirty/invalid editing open until an
explicit valid Run or existing explicit replacement resolves the draft. If the
chosen UI permits collapsing a valid dirty draft, its unsaved/invalid state and
accessible recovery must stay visible outside the disclosure; it must never look
like the previous applied question. Errors in raw filter syntax, numeric scale,
histogram bounds, grouping or other controls cannot be repaired or reconstructed
from a last valid question merely to collapse the editor. Failed validation must
not trigger evidence-first autocollapse.

A successful explicit read may return to the compact applied presentation. Do not
steal focus while a user is editing or move it into a now-hidden control. Preserve
the ordinary read-only frozen-evidence path and nested population root inertness.

## Clock, profile and lifecycle constraints

Do not call `setRecentMode(true)` as a layout shortcut. That path reparents controls
and its entry resolves a fresh recent question. Fixed read/Edit presentation must
not change `fromMs/toMs`, environment precedence, service/namespace, predicates,
measure profiles, page or any captured/reused baseline. The fresh-window control
remains the sole explicit way to release fixed bounds in this owner.

Keep existing numeric scalar/histogram, span duration/status, grouped-time and
reported-edge renderers. The consolidation cannot coerce specialized profiles
into an ordinary composition or infer values from loaded inspection rows. Compact
scope should disclose the actual applied/requested profile and relationship as
needed; the full editor and existing exact-question details remain available.

Keep DOM controls and raw values instead of remounting them on every Edit toggle.
The existing investigation-trail and population Back owners retain focus, scroll,
editor state, drafts/errors, results and independent clocks. Root editors remain
inert while covered by a population; a hidden editor callback must not bypass
those guards. Reopening a retained view or toggling Edit performs no read and must
not create another Session or acquisition controller.

## Acceptance

Compare fresh scroll-zero desktop and narrow screenshots of the same genuine
fixed package/query on old and new builds: stage, applied result, Edit, dirty/error,
Run and retained Back. The applied chosen chart/table must precede verbose
shortcuts/filter/measure authoring; this is stronger than merely showing a count
above the existing control wall. Assert complete query equality and read ledgers
around Edit, disclosure, viewport changes and Back. Exercise raw invalid draft
preservation and valid draft changes under the original fixed bounds. Cover
specialized numeric/histogram, grouped-time and reported-edge profiles with
meaningful focused tests and representative native frames; do not replace all
renderers or add shallow tests for CSS class names only.

After this finite cut, perform the planned whole-product checkpoint rather than
extending another control/owner micro-backlog.

## Candidate source review

The candidate adds only a local presentation helper and moves the existing
workbench/signal-environment fields once, after the record panel. Actual controls,
raw values and specialized renderers remain owned by Explore. Untouched staged
and applied questions collapse; dirty/invalid/failed/cancelled states expose their
controls. The Edit handler rotates a local callback lease, changes no Session and
blocks hidden/inert editor events. Direct authored callbacks use a separate
root-editor eligibility predicate; Run/Cancel retain their root execution guard.

Closed Run clones the Session's complete requested/applied question rather than
reconstructing it from hidden controls, preserving specialized fields and fixed
bounds. Existing edited/baseline behavior continues through `inputs`/`withFilters`
when authoring is open. Current group-less questions are handled with optional
`groupBy` access. Population exit, workspace changes and retained resume renew the
presentation lease while preserving the existing root inertness owner.

An initial review concern about frozen views was ruled out by the actual early
`readOnly` return to `SavedRecords.mount` before this helper is reached; no dead
read-only branch or fix is needed. Source acceptance has no remaining blocker.
The parent owns full, generated and native baseline/final gates; source review
does not establish those runtime outcomes or viewport improvement by itself.

## Late direct-handler audit

A deeper source check found a gap in the host-only mutation fence. FilterEditor
user callbacks mutate their private tree before calling Explore's guarded
`onChange`: invoking a saved hidden Clear handler could clear `snapshot()` while
the Session still appeared clean. Parent capture listeners do not run when a
saved handler is invoked directly. This finding supersedes the earlier broad
hidden-authoring acceptance until the widget-side correction is reviewed.

The bounded correction is an optional local `canMutate` admission callback on the
existing filter and numeric control widgets, checked before user-handler tree or
presentation mutations. Explore supplies `rootEditorCurrent`; other established
consumers keep the default behavior. Guard append/clear, nested and leaf mutation,
remove and other user actions, while retaining programmatic `set`/populate APIs
for the application owner to restore controls. Existing grouping intent already
delegates through `canGroup` without changing private grouping. Numeric handlers
must check admission before refresh/change. Arbitrary external writes to DOM
`.value` are outside this callback contract and do not justify a new rollback or
DOM-sanitization framework. Actual widget-handler tests should prove the private
tree, raw draft diagnostics, Session token and read counters remain unchanged.

The bounded widget correction is implemented in `filter-editor.mjs` and
`numeric-measures.mjs`: optional `canMutate` defaults preserve other consumers,
while Explore supplies its editor-current predicate. Filter user callbacks check
before private-tree mutation, render, focus or value loading; numeric user
callbacks check before refresh/change and also reject read-only/disabled inputs.
Programmatic restoration APIs remain available. Syntax and whitespace checks
passed. Source acceptance is closed again; actual hidden Clear and mounted widget
negative controls remain part of the parent's required validation.

## Final guard review and validation status

Read-only review confirms both Explore widget mounts pass `rootEditorCurrent` to
`canMutate`. The new mounted `editor-mutation-358.test.mjs` exercises actual Clear,
append/nested/remove/leaf/group handlers: denied callbacks preserve the private
filter tree, rendered-node identity, feedback, focus and emitted changes. Loader
negatives prove no acquisition; admitted/default mutation and programmatic setters
still work. Numeric negatives verify refresh/change remain blocked for denied,
read-only and disabled handlers. These tests intentionally do not claim to undo
arbitrary external writes to input properties.

Final source acceptance is closed, including the widget-side correction. The
parent reports the post-guard full UI suite (1,220 tests), generated 40-case
properties and wrapper pass. The first actual native pass includes hidden Clear,
raw-invalid Back, cancellation/late completion and whole-question/clock/DOM-return
facts; its final rerun remains pending after a request-vs-Go-echo oracle correction.
No product change was needed for that normalization. Final native/model/frame
acceptance is recorded separately by the parent rather than presumed here.
