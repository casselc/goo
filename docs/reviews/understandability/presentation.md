# Presentation and view grammar: understandability review

Reviewed current code on 2026-11-16. Scope: shared Web Components, native plotting,
checked dataset/view/composition contracts, query-bound authoring and CSS
boundaries. Application acquisition and preservation owners are reviewed
separately. [Walkthrough](../../walkthroughs/presentation.md).

The architecture has a useful foundation: passive display elements do not acquire
source authority, exact values remain separate from coordinates, and structural
validation does not fabricate completeness. The principal obstacle to an
“understand it in an afternoon” contribution is the authored code's presentation
and the number of adjacent contracts a contributor must discover.

## Recommended work, in order

### 1. Make authored code readable before introducing more layers — high value

Evidence: `wc -L` reports maximum lines of 1,914 characters in
[components.mjs](../../../ui/components.mjs), 1,509 in
[record-presentation.mjs](../../../ui/record-presentation.mjs), 970 in
[view-composition.mjs](../../../ui/view-composition.mjs) and 669 in
[view-schema.mjs](../../../ui/view-schema.mjs). These are authored files, not build
output. Small line counts therefore misrepresent reading complexity. Inspector
construction, validation, DOM mutation, focus restoration and callbacks can share
one line.

Proposed first patch: format these representative modules into normal statements
and blocks; add short comments where ownership or exactness explains a decision.
Keep public behavior and validation order unchanged. Do not equate splitting a
file with simplifying it. Acceptance: the passive graph and focused semantic,
owner, input-safety and browser tests remain unchanged; review a complete
function without horizontal scrolling or reconstructing compressed control flow.

### 2. Give the contracts one navigable map — high value, partly supplied here

The actual chain is [view-schema](../../../ui/view-schema.mjs) → immutable
[snapshots](../../../ui/snapshots.mjs) → runtime
[view-model](../../../ui/view-model.mjs) →
[scope-view](../../../ui/scope-view.mjs). Query-bound declarations and composition
form a separate chain. `view-model` re-exports the schema and deliberately adds
runtime checks; `createSourceAdapter` and `hasGuarantee` distinguish an admitted
source result from copied JSON.

These are meaningful boundaries, not dead duplicate validators. Publish a small
shape/kind/producer/renderer matrix alongside the walkthrough and distinguish
structural admission, trusted-source adoption and presentation eligibility in
function names/comments. Keep an executable case where copied completeness
metadata cannot enable complete-only rendering. Do not merge the runtime witness
into a boolean supplied by users.

### 3. Extract a cohesive authoring controller, not unrelated callbacks — medium/high

[population-workspace.mjs](../../../ui/population-workspace.mjs) is 899 lines and
contains `mountCompositionEditor`, `mountDisplayEditors`, `mountFacetEditor`,
summary editing, source text parsing, template compatibility, portable export,
save state and acquisition control. The pure declarations already live in
`view-declaration`, `view-composition`, `summary-declaration`, `timeline-declaration`
and related modules. A newcomer looking for “the editor” must still understand
most of the workspace closure.

Propose one application-owned presentation-editor controller receiving explicit
current-question, draft and mutation capabilities. Move the editor's local
lifetime and focus handling as a unit; keep source reads and preservation with
the workspace. This is an ownership refactor, not a generic event bus. Acceptance
must retain stale/hidden/disabled-control withdrawal, save-with-draft refusal,
local-only kind changes and exact held-return behavior. Coordinate with Activity
and preservation reviewers because those closures currently enforce real safety.

### 4. Reduce schema branching through named families — medium

[view-schema.mjs](../../../ui/view-schema.mjs) combines core field/row validation,
quality/attestation checks, exact arithmetic, log patterns, raw gauge/histogram
profiles, grouped intervals, summaries, ratios and renderer eligibility.
`validateSnapshot` accumulates reasons in a long cross-profile traversal.
`aspect` shape mapping is repeated in schema, snapshots and view-model.

Propose a small, explicit core contract plus named family validators and a shared
pure aspect descriptor where meaning is truly identical. Keep one readable
admission sequence and stable rejection semantics. Do not add a dynamically
extensible validator registry: the finite supported profiles are easier to audit.
First document which checks are structural and which depend on the question or
runtime witness; moving the same checks into more files alone adds little value.

### 5. Keep the component portfolio honest — medium product/coherence concern

[display.mjs](../../../ui/display.mjs) is a convenience export of a tested
91-module passive graph. `scope-view` handles checked dataset kinds, while record
pages, span waterfalls, held chronology and specialized metric views also expose
narrower models. [chart-frame.mjs](../../../ui/chart-frame.mjs) is chrome only,
not a base rendering engine. No Plotly runtime or generic WYSIWYG editor is in
`ui`; their design documents explicitly call them proposals.

Add a concise component capability table rather than promising that every current
screen is expressible by the same grammar. Prefer extending an existing
compatible projection/renderer over creating another one-off adapter, but do not
force record occurrence provenance or metric temporality into a generic chart
shape that erases their meaning. Direct feature imports already avoid the broad
bundle when desired; no replacement bundle system is justified by this review.

### 6. Remove confirmed small leftovers; do not invent dead modules — low/routine

[numeric-plot.mjs](../../../ui/numeric-plot.mjs) declares `const global=globalThis`
with no use, and routes literal names through a local `dep(name)` table despite
static imports. The lookup is not an extension API. Similar dependency shims occur
in several UI modules. Remove unused aliases and replace proven literal-only
lookups with direct imported names during the readability pass.

This review does **not** establish that any complete chart, kind, schema family
or authoring module is dead. A filesystem import count is insufficient: the HTML
entry, custom-element registration, saved definitions, source profiles and
executable storyboards can be entry points. Require those checks before deletion.

## Performance and reliability

The 462 frozen parent-ID projection is a good bounded example: one semantic
projection is shared without weakening public defensive-copy behavior. Apply
that approach to a measured repeated transformation, not indiscriminately to
all objects. `scope-view.render` currently rebuilds much of its subtree, but
special cases retain raw-histogram page state and read-facts focus; a broad
“incremental rendering” rewrite must preserve those affordances and show measured
benefit. This review has no timing or memory benchmark that establishes a new
performance regression.

Reusable component styling is in `components.css` and feature CSS, with
application placement and integration rules in the 2,015-line `product.css`.
Consolidate only demonstrated repeated selectors, checking standalone/nested
composition and forced colors. CSS length alone does not prove unused rules.

## Executed checks and limits

Ran seven focused files: view-model, view-composition, composition reuse,
view-artifact, chart-frame, display graph and record-list. All passed, zero
skipped. The review changed documentation only. No browser, screen-reader,
load-performance or dead-code reachability campaign was executed here. Current
real-data and forced-color evidence is recorded in `VALIDATION.md`.
