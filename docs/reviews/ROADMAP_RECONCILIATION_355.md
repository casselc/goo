# Roadmap reconciliation — 355

Proposal after 352 breadth → 353 depth → 354 consolidation. At this review, 354's
source contract is accepted and its runtime/native gates are pending. Finish those
gates before adopting or implementing this queue. This document changes no
production code, active roadmap, source budget or compatibility requirement.

Read with the [architecture checkpoint](WHOLE_PRODUCT_ARCHITECTURE_355.md),
[design checkpoint](WHOLE_PRODUCT_DESIGN_355.md),
[351 reconciliation](ROADMAP_RECONCILIATION_351.md),
[active roadmap](../ROADMAP_2026_11_05.md) and [known limits](../KNOWN_ISSUES.md).

The product remains **see current telemetry → isolate a population → follow its
supporting signals → preserve/reuse**. The next sequence should make authored
views useful at a natural service entry, deepen an explicitly missing convention
family, then improve reading the resulting fixed question. It should not reopen
completed scope work or become another string of callback and facet tweaks.

## Reconcile completed work and current boundaries

| Work | Keep | Remove from the active queue |
| --- | --- | --- |
| 348–350 | Composition explanation, service-local inventory, typed package profiles, immutable definitions and Stage → explicit Run | Another explanation/inventory/package engine; scalar-only authoring assumptions; Use-means-read behavior |
| 352 | Typed wider-child questions and environment-policy-scoped Activity libraries | Dropped-selector continuations and the native-capability gate around configuration-only reuse |
| 353 | One checked policy for parent ownership and child membership, raw membership admission, exact relationship capture/reopen | “Typed reported edges are unsupported”; inferring parent uniqueness from stored child rows |
| 354, after its own acceptance | Frozen planned widget questions, independent slots/receipts, explicit retry under a bounded owner/queue | Whole-overview rereads as the only recovery path; a new generic retry framework |
| Existing grammar/view foundations | Shared display kinds, ordinary authored templates, portable artifacts, record/aggregate compositions and source-free expansion | Reimplementing those foundations or treating historical proposal 260/checkpoint 274 as a parallel migration queue |

354's lessons apply beyond retry: isolate the planned intent from mutable callers,
keep one current result slot, retire before abort, and distinguish a supported
failed read from an unsupported prerequisite. Successful siblings do not acquire
new clocks when another widget recovers. Those are task-level invariants, not
reasons to expose a universal owner or capability system to users.

## Proposed finite sequence

### 356 — breadth: use an authored template from service evidence

**User task:** “The useful view already exists; apply it to this service's current
population without rebuilding its layout or importing somebody else's scope.”

The foundations exist. `ui/service-dashboard.mjs` has a scoped Customize handoff;
`ui/view-template-handoff.mjs` and `workbench/view_templates.go` provide checked
shipped/authored expansion. The population workspace already lists saved authored
templates and binds them to a destination question. The service entry currently
offers shipped HTTP/database presets, not the existing authored-template catalogue.

Add the smallest contextual bridge using the current view store, catalogue and
expansion endpoint. Start from already-supported ordinary record-widget questions
and existing template profiles; do not treat native stream widgets, reported-edge,
numeric/histogram or grouped-time templates as supported by removing their guards.
Close exact initial eligibility before code. The action may name its selected
widget population, but must not invent a service-wide denominator from that slot.

The current complete widget question supplies service/namespace/environment,
fixed bounds, predicates, measure and destination grouping. The saved template
supplies compatible presentation and declared binding requirements. Its specimen
query, loaded values or baseline do not supply destination authority. Missing
interval widths, explicit destination baselines and incompatible profiles must
remain visible requirements; do not silently drop a child, widen a query or fetch
a baseline on the user's behalf. “Team template” means contributed saved
configuration here, not verified team ownership or multi-user authorization.

**Acceptance:** import/save an authored template in the existing catalogue, enter
through a genuine service widget, inspect compatibility and stage the expanded
view with zero telemetry reads. Use typed current/effective named and a suitable
empty/unreported case with namespace/literal filters. Prove the destination's
whole query survives and the template's different specimen service/window does
not leak. Explicit Run acquires that fixed population; Back restores the service
overview and its independent healthy receipts without rereading them. Refuse wrong
signal/profile/grouping/baseline/interval and changed selected definitions before
adoption. Selection changes, late expansion and retired actions must not publish.
Inspect the actual desktop/narrow chooser and staged scope, not merely an API plan.

**Stop:** a contextual entry into existing declarative reuse. No new catalogue,
template language, contributed JavaScript, kind runtime, whole-view WYSIWYG,
automatic dashboard subscription or permission system. This is the best first
item because it makes the existing grammar useful where users are observing a
service, rather than adding another isolated authoring feature.

### 357 — depth: a pinned current-RPC observation profile

**User task:** inspect current-convention RPC telemetry with a useful default
overview, while understanding exactly which records and native streams it covers.

`exploration/dashboard.go` and the service presentation deliberately implement a
legacy `rpc.system`/`rpc.method` lens and legacy millisecond duration instrument.
`docs/SERVICE_DASHBOARDS.md` explicitly defers current RPC. Do not relabel that
existing interpretation as current or union renamed keys into one population.

First research and pin the applicable published OTel convention. Record its exact
role, required/presence keys, method grouping, native metric names/units and status
meaning before implementation; this review does not establish those facts from
the appearance of `rpc.system.name` in presentation metadata. Then add one named,
versioned current-convention profile through the existing planner, candidate
selection, explicit profile Apply, record widgets and exact native tooling.

Keep positive held-page hints advisory and tied to the checked question. A key
appearing somewhere in a bounded page is not a same-row witness or proof of
absence. Auto precedence and mixed current/legacy fixtures need a deliberate rule;
manual profile selection remains explicit. Recorded span Error is not inferred
protocol success/failure, and a method name is not a globally unique operation.
Native units, stream identity, temporality and clocks remain independent; do not
convert or pool older instruments into eligibility.

**Acceptance:** genuine current-only, legacy-only, mixed-key and missing/empty
fixtures distinguish exact queried populations and profile rule sets. Verify
Server/client separation, supported versus incompatible native units, candidate
versus observed state, explicit Apply and manual All spans fallback. An empty or
failed current profile does not silently broaden. Supporting records, 354's
single-widget recovery and retained Back preserve the same full scope. Generated
semantic tests use independent raw membership/witness oracles; native frames show
the selected convention and unknown-data meaning.

**Stop:** this one researched RPC family. No alias-normalization engine, broad
protocol outcome inference, broker/Kubernetes catalogue or request-count merging.
Straightforward internal Go/profile cleanup is welcome if this implementation
demonstrates duplication, but not a prerequisite public descriptor protocol,
cross-language generator or clever user-facing type machinery.

### 358 — consolidate: read applied fixed-question evidence first

**User task:** “I ran a saved or shared question; show the result, and let me edit
the definition when I choose.”

The actual 350 fresh-package result frame shows an Applied scope and count summary,
but its chosen duration/status/chart/record result remains below signal controls,
shortcuts, suggested breakdowns and an expanded filter editor. This is a real
reading-versus-authoring inconsistency, not an assertion that no evidence is
visible. Activity analysis and saved/population canvases already have data-first
presentation and explicit editing affordances.

Consolidate read-versus-edit presentation in the existing fixed-question owner.
Use existing scope/status, result renderers and Edit controls for package, shared,
service/edge and other fixed entries; avoid a package-specific screen. Staged
configuration must still show its checked scope and explicit Run without claiming
an observation. Applied state should put substantive result content before
expanded authoring. Opening Edit reads nothing and retains the complete draft;
collapsing cannot hide invalid changes, silently apply them or discard them.

Keep presentation independent of window policy and acquisition. Do not use a
recent/rolling mode flag merely to borrow CSS if that changes fixed bounds or
execution behavior. Specialized numeric/histogram/grouped-time/reported-edge
renderers remain specialized; this cut does not make them ordinary templates or
move them into a new workspace. Existing query-session, dirty-navigation and
retained-parent contracts remain authoritative.

**Acceptance:** matched fresh-entry, scroll-zero 1280×1000 and 390×844 journeys
before/after for Stage → Run → inspect → Edit → dirty draft → return. Measure first
substantive result position and visible authoring-control count for the actual
duration/status/grouped-time fixture, plus representative simple and specialized
profiles. Do not substitute a scope/count summary or a manually cropped task view
for the selected result. Prove identical full questions, source budgets, fixed
bounds, read receipts, saved bytes and draft/Back behavior. Keyboard reachability,
focus and retired hidden callbacks remain valid; no new read occurs on display or
edit toggles. No universal above-fold or latency improvement is claimed without
the corresponding measurements and actual frames.

**Stop:** evidence-first presentation using current owners/renderers. No new
navigation shell, retention promise, source cache, generic session manager,
automatic Run or wholesale exploration rewrite.

## Selection tradeoff and retained work

The architecture review proposes a measurement-gated large-facet presentation cut;
the design review identifies the current applied-question reading issue. Prefer
the latter for this finite third slot because it blocks an ordinary completed
workflow and has direct viewport evidence. Keep performance work available for a
demonstrated current task, not as a presumed virtualizer implementation.

The roughly 32,000-element paired facet tree remains. In 346, the 1.6–2.1 second
figures were baseline repeated-publication CPU, not final render/paint. Final
single-build paired CPU samples were roughly 428/408ms; wall samples were mixed
and did not establish a uniform improvement or sole DOM cause. A future bounded
presentation task must establish a fresh styled baseline, then preserve all
admitted values, full-population counts, local Find, exact literals, keyboard
focus and source-free Back. It must not silently truncate values, raise source
budgets, or infer total work from setter counts alone.

Also retain without adding them to this sequence: durable investigation paths and
collaboration after their schema/authority contract; broader Kubernetes/USE/broker
defaults; native fractional/grain extensions where a real panel needs them;
specialized live recovery; production tenancy/deployment and retention guarantees.
Existing HTTP/legacy-RPC/DB/messaging/resource/native defaults and portable views
are implemented foundations. Databricks lowering, general CEL/KCL calculations,
DataFusion/XTDB replacement and whole-view WYSIWYG are not prerequisites here.

Finish the vulnerability assessment when official database access is available;
no clean assessment is claimed. Go 1.27.3 and modern standard/extended libraries
are tools for simpler internals, not a reason for another maintenance-only cut.
Preserve independent semantic tests when consolidating implementation tables;
generating both the validator and its supposed oracle from the same descriptor
would not prove convention correctness.

Retire 351's delivered continuation/edge/recovery proposals only after their own
acceptance, including any unresolved historical harness limitations as such.
Review source, genuine entry/return, compact observed/model outcomes and actual
frames after each new cut. Repair concrete blockers and rewrite the remaining
order explicitly; after the third, reconsider the whole product rather than
extend this queue indefinitely with ownership or styling minutiae.
