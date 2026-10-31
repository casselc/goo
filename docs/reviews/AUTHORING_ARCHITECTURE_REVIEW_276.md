# Authoring architecture checkpoint — 276

2026-10-31. Read-only whole-product review of the implemented 275/276 authoring
path and the [composition roadmap](../ROADMAP_2026_10_30.md). No hosted competitor
instance, new backend, or additional validation run is claimed by this review.

## Current architecture and limits

The sound boundary is already present: a checked query owns the population;
source admission owns an immutable dataset and runtime guarantees; a display
owns presentation intent; the application owns reads and execution-bound actions.
`workbench.ExplorationView` stores queries and per-signal `ViewDeclaration`s without
observations. `ui/population-workspace.mjs` uses the existing configuration catalog,
explicit Run, source admission, independent draft revisions, and H1 bindings.
`ui/scope-view.mjs` stays passive. Preserve this separation when adding reuse.

276 adds numeric mean/sum/min/max and duration mean/min/max/p50/p95/p99 to the
ordinary one-group bar/table profile. `exploration.ViewMeasureFields` is the shared
Go metadata source used by canonical projections and saved declaration validation.
`ui/view-declaration.mjs` mirrors the bounded measure profile; non-count population
share is refused. Runtime missing/support states and signed-value table fallback
remain responsibilities of supplied data and the renderer. A valid declaration
cannot assert that any values, completeness, or service behavior were observed.

This is a useful authored display, not a complete composition language. The
following remain absent from this path: independently contributed kinds; reusable
parameterized display templates; authored native/time-series/histogram profiles;
whole-view layout; WYSIWYG editing; and a shared live subscription grammar.

Existing reuse models are adjacent but different:

- `prebuilt/catalog.go` describes telemetry compatibility; `prebuilt/expand.go`
  expands one bounded HTTP prepared-table binding into measurements/views/recipes.
  Its `View` shape is not the implemented `workbench.ViewDeclaration` contract.
- `annotations/manifest.go` describes source operations/events, signals and
  requirements; its manifest does not contain authored displays or guarantee
  that a declared operation emitted telemetry.
- `exploration/dashboard.go` plans typed service widgets and profiles without
  reads. `ui/service-dashboard.mjs` executes record/measurement widgets and owns
  their actions and specialized presenters. Its record widgets are a practical
  entry into saved authoring; native widgets need their own admitted profile.
- `workbench/activity_views.go` stores a rolling selector and record columns,
  without a display declaration. It still explicitly refuses typed environment
  selectors. It must not be treated as equivalent to fixed exploration views.

## Four actionable priorities

### 1. Define the supported authoring profile once as descriptive capabilities

**Gap:** measure keys and profile conditions appear in Go metadata/validation,
`ui/view-declaration.mjs`, and the reviewed OpenAPI annotations in
`scripts/build_openapi.py`. The renderer also has broader capability rules than
this first saved profile. Adding a kind by updating only one list is a footgun.

**Next architecture increment:** create a bounded authoring-profile descriptor
for the current record-group profile: versioned kind choices, field family/key,
unit/support metadata, local display options, and requirements on the checked
query. Reuse or generate metadata from canonical field descriptors. It must
explicitly distinguish *query-declared eligibility* from *observed rendering
eligibility*; the former never grants runtime completeness or an H1 action.
Keep Go semantic validation authoritative and retain cross-language parity tests.
Do not introduce a second query planner or an arbitrary browser expression engine.

**Acceptance:** the same checked numeric/duration query produces consistent
editor choices and refusal reasons in source validation, the UI and API docs;
unknown and unsupported profiles remain explicit. A partial or copied dataset
cannot become complete through profile metadata. This work should ship with the
next user-facing reuse path below rather than become a standalone registry rewrite.

### 2. Make an existing service context open the same authoring flow

**Recommended next product slice:** add a contextual **Customize view** action to
one supported record-based service diagnosis/widget. Lower its typed query and
current grouping/measure into the existing fixed exploration workspace and
`ViewDeclaration`; if a grouping is required, stage a checked choice before Run.
Reuse the existing editor, Save and library. Preserve the fixed environment,
namespace predicates, window and explicit numeric binding. Opening the draft
performs no telemetry read and does not copy widget data or its runtime witness.

Code seams are `DashboardWidget.Query` / `PlanServiceDashboard` in
`exploration/dashboard.go`, widget actions in `ui/service-dashboard.mjs`, and
workspace `stage` / declaration validation in `ui/population-workspace.mjs`.
Measurement-backed widgets must show unsupported authoring rather than being
converted into record questions. Do not migrate every dashboard presenter first.

**Acceptance:** a real discovered OTel service → customize a compatible grouped
span count or duration view → edit → save → reopen → explicit Run journey,
with unchanged typed scope and read counters. This connects the authoring grammar
to normal investigation work instead of creating another authoring destination.

### 3. Add one bounded record-view template expansion, then connect packages

After that journey, support a versioned, declarative record template that expands
explicit typed bindings into the same checked query plus declaration. Start with
one shipped service/HTTP grouping and duration template. Bind environment/window,
service/namespace and qualified grouping fields through existing types; numeric
unit/parse policy must be explicit. Template expansion stages a draft without
executing, publishing, or certifying observed compatibility.

Use `prebuilt` compatibility proposals as recommendations, not proof. A manifest
may reference an exact template/profile revision; do not embed a second renderer
syntax in `annotations.Manifest` or continue producing a competing generic `View`
wire for this new path. A team-authored template can follow once the shipped one
round-trips through the public configuration contract. Keep generated definition
identity separate from template revision and telemetry capture identity.

**Acceptance:** identical template revision and typed bindings produce identical
validated configuration; an absent field/binding gives a bounded reason; no
string substitution builds SQL, JavaScript, or query predicates. Reuse the same
catalog and editor after expansion. No CEL/KCL runtime or WYSIWYG layout is needed
for this slice.

### 4. Extract the editor state machine only when adding the second owner

`ui/population-workspace.mjs` currently combines query drafts, declaration source,
Find drafts, current preview, save acknowledgement and execution ownership. The
275/276 guards are useful; duplicating these handlers in a service editor would
create inconsistent saved-state behavior.

When implementing priority 2, keep one application controller for declaration
edits and source/preview validity, plus a separate owner for applied queries and
source execution. Expose bounded intents for edit/apply/reset/save; share the
existing editor rather than inventing a framework or rewriting the workspace.
Property changes must not overwrite invalid source, a late201 must not mark a
newer draft saved, and partial/unavailable data must not be repaired by editing
presentation. Keep these as state/property/browser checks.

Activity integration should follow as a distinct rolling-scope profile with one
subscription owner. First resolve its explicit selector-persistence gap; do not
reuse a fixed-window declaration by silently manufacturing or widening scope.

## Roadmap decision

Advance through the service-context reuse slice and one shipped record template,
with capability metadata and editor extraction done only as necessary to support
those outcomes. Then review before contributed templates or whole-screen layout.
Do not insert Plotly, another presenter-conversion campaign, a universal query
model, or an external bitemporal store as prerequisites. The existing passive
components, checked queries and immutable configuration catalog are sufficient
for the next useful authoring step.
