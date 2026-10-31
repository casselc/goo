# Architecture checkpoint — 280

2026-10-31. Read-only review of the working tree through whole-composition reuse.
This is an architectural recommendation, not additional test evidence or a claim
that the current round has shipped; executed checks remain in [VALIDATION](../../VALIDATION.md).

## What is now real

The [276 architecture](AUTHORING_ARCHITECTURE_REVIEW_276.md) and
[product](AUTHORING_PRODUCT_REVIEW_276.md) reviews correctly prioritized reuse over
another renderer migration. Their main concrete handoffs are implemented:

- 277 applies an aggregate display to a compatible current fixed population.
- 278 expands a shipped HTTP route template into checked query/display intent and
  opens the existing editor. `workbench/view_templates.go` checks mandatory service,
  Server-kind and current-method predicates; `workbench/templates/http-server-routes.json`
  owns the bounded default. The template does not reconstruct a wider source query.
- 279 makes record-table and aggregate children explicit in `ViewComposition`.
  Child options cannot coexist with flat presentation options. One application
  owner supplies an immutable bounded record page separately from the aggregate.
- 280 reuses the whole composition: pure Go/JS compatibility retains destination
  query and paging, while the application guards pending reads, drafts and actions.

The useful layers are intact: `exploration.Compile` owns source-query semantics;
`exploration/view_measures.go` supplies exact typed measures; `ui/source-snapshots.mjs`
owns admission; `ui/scope-actions.mjs` owns retained execution bindings; passive
presentation stays in shared components. `ui/record-page.mjs` preserves rich rows
and paging without claiming aggregate completeness. Private record-page ownership
is not authentication, and an exhaustive group read is not upstream completeness.

This is a functioning composition grammar for one fixed record-query family,
not a universal UI language. The generic plotting dataset still has scalar cells;
rich record pages are an explicit separate supplied-input profile. Go/JS declaration
validation and OpenAPI describe a narrower authored profile than all existing
renderers. Query execution remains independent from both template and display syntax.

## Next one user-facing capability: reusable time overview with matching records

Let an operator save/reuse a fixed, bucketed count timeline beside its record page,
select an exact interval, and explicitly Run the narrower population. This advances
"what is happening, and when did this population change?" across ordinary OTel
services/logs without introducing a second editor or a chart vendor.

The foundations already exist: `exploration/source.go` emits `TimeDataset`;
`timeViewDataset` in `exploration/view_dataset.go` checks exact bucket origin,
clipped final bounds and interval coverage. `ui/view-schema.mjs` registers `time-bar`,
`ui/scoped-breakdown.mjs` renders temporal snapshots, and `ui/scope-actions.mjs`
already binds the `intervals` slot. `ui/exploration.mjs` uses that staging path.
The authored composition path instead hardcodes `data.viewDataset`/`groups` in
`population-workspace.mjs:renderAuthored`, while `ViewDeclaration.Validate` only
admits ordinary grouped bar/table intent. The missing capability is saving and
reusing this existing temporal family, not implementing a new time-series engine.

Keep the first slice to record-count time buckets plus records. Require a checked
bucket width and pinned temporal kind/options; preserve exact UTC boundaries and
unknown/empty/partial states. Do not relabel counts as rates, add population-share
semantics without a supported denominator, or imply trace/ingestion completeness.
Selecting an interval stages a checked question through the current owner; no
component fetches. Paging retains the timeline's basis, and source/visual edits,
save/reopen and whole-composition reuse retain their existing explicit-read rules.

## Refactors justified only by that delivery

1. **Resolve the declared aggregate aspect in one small place.** Map a checked
   profile to its existing admitted slot (`groups` or `intervals`), eligible kind,
   measure metadata and action binding. Use that resolver for rendering, editor
   choices and reuse compatibility. Keep Go authoritative and parity tests; the
   browser metadata validator is not a second SQL compiler. This is the concrete
   opportunity to reduce the repeated Go/JS/OpenAPI profile lists identified in276.
   Do not rewrite the whole kind registry or add arbitrary expressions first.
2. **Reuse execution/interaction ownership instead of duplicating controllers.**
   The population owner already coordinates drafts, catalog inspection, paging
   and saved acknowledgements. Add temporal staging through the existing H1
   resolver. Extract a small owner helper only where the new branch would otherwise
   duplicate those guards; no service-specific or timeline-specific editor.

Acceptance should be a real scoped OTel log/span timeline + one-row record-page
journey: exact interval selection, zero local reads, explicit narrower Run,
unchanged timeline during paging, truthful unavailable coverage, immutable
save/reopen and compatible current-scope reuse. Negative controls must reject
wrong width/profile, stale interval ownership and invented completeness.

## Reconciled deferrals

Keep the remaining roadmap families explicit: Activity rolling views still have
an environment-selector persistence gate (`workbench/activity_views.go`); fixed
composition must not be presented as a subscription grammar. Existing waterfalls
(`ui/waterfall.mjs`) remain useful passive trace displays but need their own trace
input/relationship profile before authoring. Native histograms retain measurement
scope and temporality. Team template upload, manifest-contributed kinds, grids and
WYSIWYG follow proven profile delivery rather than blocking it. The distinct
prepared-table `prebuilt` measurement/policy API should not be advertised as the new
view-template wire. Production auth/isolation, retention and durable alerts/SLOs
remain separate product workstreams, not problems solved by composition.
