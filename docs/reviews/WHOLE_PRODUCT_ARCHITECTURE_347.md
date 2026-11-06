# Whole-product architecture checkpoint — 347

Proposal only, after the 332–346 sequence. 346's final native acceptance is still
owned by its current gate; this review starts no implementation and does not
replace the active roadmap. It is a bounded source/document checkpoint, not a new
competitive-feature survey.

## Product essence and what has changed

Goo's core remains **see current telemetry → isolate a population → follow its
supporting signals → preserve/reuse the explanation**. Ordinary OTel records must
be useful without manifests; manifests, reusable grammar and native metrics add
stronger definitions when available. Held occurrences, independent query results,
immutable receipts and reusable configuration retain different authority.

The completed sequence materially strengthened the middle of this flow:
Activity previous-window comparisons; whole-operand supporting records; saved-view
and retained-parent ownership; shared facets with independent category populations;
typed-environment trace search; exact ordinary mean-duration comparison; source
refusal and contextual recovery. 343/346 measured real large lists before changing
local work. This is enough facet/control work for now. The next cuts should expose
existing analytical value and close ordinary service/authoring scope gaps.

Keep one Go runtime/source adapter, typed query AST and passive framework-neutral
components. Go 1.27.3 and modern standard-library error handling already fit the
project; another runtime/optimizer/store is not a prerequisite for the next user
journeys. Prerelease status favors direct correction over compatibility aliases.

## Proposed next three cuts

### A. Breadth: explain differences from an ordinary fixed pair

**User task:** traffic or mean duration changed between windows; which reported
populations differ, and can I inspect one side without rebuilding my investigation?

**Source gap:** `population-workspace.mjs` presents fresh count/mean comparisons
but does not mount Explanation. `ui/distributions.mjs` and
`activity-explanation.mjs` already use `Explanation.mount`, and
`exploration/explanation.go` supplies checked multi-field population explanations,
per-field unknown/refusal state and side-filter planning. This is an entry gap,
not a missing scoring/query engine.

Use the existing explanation owner from the held pair's two complete fixed query
definitions and actual operand roles. Opening/choosing fields stages only; explicit
Run rereads both scopes using the existing bounded field-count/group budgets.
Retain typed environment policy, independent windows and original predicates.
The explanation is new evidence, not a computation from the loaded page or proof
that the held pair's old clocks advanced. Show its independent observations beside
the held comparison. A mean-duration pair can initiate this workflow, but the
existing explanation ranks count/share differences: it does **not** attribute
latency, establish causality or calculate significance.

Acceptance: ordinary logs/spans, unequal populations and a baseline-only value;
field selection without IO; explicit full-scope reads; partial field/refusal
semantics; exact side follow-up preserving operand label/window/policy; retained
Back and cancellation/late-owner refusal. Reuse the existing explanation engine,
passive comparison primitives and native/model oracles. No automatic field/causal
inference or new anomaly engine is selected.

### B. Depth: service-local operation inventory and capture

**User task:** open a discovered service and identify its operations even in a
large environment, then preserve exactly that service's operation observation.

**Source gap:** `service-browser.mjs` requests `/api/inventory` with activeScope,
then filters `inventory.rows` by service in the browser. `InventoryQuestion` has
window/environment/namespace but no service selector. `CompileInventory` groups
all ServiceName/SpanName pairs with LIMIT 1001, and source admission refuses more
than 1,000. Thus unrelated services can prevent a small selected service's
operation inventory from being available. Its current capture intentionally
captures the whole environment/namespace window, not the displayed service.

Add a bounded exact service selection to the existing inventory contract and
compile it before grouping. Preserve typed current/effective environment and
missing/empty namespace semantics. Do not add a second inventory endpoint,
client-side top-K, guessed health, operation deduplication or a cardinality index.
Service-local count/errors/duration buckets describe captured span records, not
unique requests. Global inventory remains an explicitly distinct scope.

Acceptance: target service below the budget while unrelated services exceed the
global limit; same-named services in distinct namespaces; effective/current-only
policy conflict; exact service filter and fixed bounds in SQL parameters/query
echo; no global denominator relabeled as service-local. Capture and reopen preserve
the exact service-scoped question and new read receipt. Local operation Find is
source-free; operation follow-up adds only the checked name condition. Update the
capture action's wording and evidence, rather than silently changing its scope.

### C. Consolidation: scope parity for manifest-free grammar authoring

**User task:** take a useful investigation over ordinary telemetry, save it as a
reusable telemetry package and use it again without losing the applied policy.

**Source gap:** `ui/exploration.mjs` explicitly disables Create package when
`question.environmentSelector` is present. `exploration/package.go` has scalar-only
PackageSource.Environment and DraftTelemetryPackage explicitly rejects typed
selectors. Modern Activity/Services workflows intentionally use typed scopes,
so the core manifest-free grammar path is disabled in ordinary new investigations.
A user should not have to weaken an effective/empty/unreported scope to scalar
current-key naming to author a package.

Reuse the shared typed environment selection through draft, YAML/JSON validation,
plan, immutable package publication, reopen and explicit preview/reuse. Preserve
relative lookback configuration versus fixed preview question and source receipt.
Do not copy absolute observation clocks/rows into reusable configuration. This is
scope parity, not wholesale language redesign or warehouse lowering.

Importantly, actual TelemetryPackage already retains Numeric, NumericHistogram,
Summary, SpanStatus, GroupTimeSeries and BucketMS, despite its fixed
`measure: captured-record-count` tag. Do not propose implementing these profiles
again based on old count-only prose. Inventory their actual supported round trips
and correct the docs; broader measure-expression unification needs its own task.

Acceptance: typed named effective/current-only, empty and unreported definitions;
strict policy-preserving round trip through draft/YAML/plan/save/reopen; compile
and preview match the original full selector and profile; no source IO during
validation/build/open; explicit preview/new-window reuse retains separate clocks;
scalar weakening and invalid selector unions refuse before acquisition. Optional
Go annotation handoff remains reviewable and is not instrumentation inferred from
a sample. No legacy alias/migration layer is selected for this prerelease.

## Stale or misaligned backlog to reconcile

- Retire execution proposals for scalar-only trace search, repeated saved-view
  loads, stale parent controls, source failure URL cleanup and contextual retries:
  335/340/342/344/345 delivered their bounded paths. Keep their documented limits
  and regression coverage, not parallel implementation tickets.
- Reword generic "facets cannot compose with comparisons" and "count-only
  comparison" claims where 337–341 superseded them. Frozen/captured mean,
  per-category means and percentile pooling remain deliberately deferred; they
  are not automatic acceptance criteria for every new renderer.
- Old telemetry-package prose is narrower than the actual optional numeric/
  duration/status/profile fields. Audit code first; keep scalar typed-scope parity
  as the demonstrated missing piece. Do not start another parallel grammar format.
- Saved rolling Activity views and fixed reusable questions already have explicit
  staging/lookback semantics. A new generic saved-scope abstraction needs a real
  unsupported journey; neither "make it rolling" nor universal session state is
  currently a sufficient task definition.
- The 1 MiB pure comparison-body budget is independent of 1,000-per-operand group
  admission and facet response bytes. Keep this real cross-layer limitation
  visible; do not treat source-refusal UI or atomic publication as solving it.
  Coordinate a separate bounded admission task when an actual required definition
  cannot fit, rather than raising all limits.
- The roughly 32,000-element paired DOM remains. 346 removes measured duplicate
  publication; that does not authorize a virtualizer or prove all mount/paint
  cost solved. Further list optimization needs another measured user problem.
- Dependency vulnerability assessment remains blocked by configured network access
  in the current record; no clean scan is claimed. Treat it as an operational
  follow-up when the published policy permits the database, not a reason to defer
  independent product work or invent a clean result.
- Whole-view WYSIWYG, general CEL/KCL calculations, DataFusion, warehouse lowering,
  XTDB expansion, multiplayer editing, notifications and broad competitor parity
  remain larger proposals. Current source does not make them prerequisites for
  the three cuts above. Preserve research without presenting it as the next queue.

## Process and verification

Finish 346 before selecting A. For each proposed cut, close one scope contract,
implement one real journey, obtain architecture/design review and then proceed.
Use meaningful Hegel properties and observed model validation where authority or
lifecycle changes; do not create a model merely for display copy. Native source
membership, IO, actual focus and current screenshots remain separate evidence.
Retain compact reproducible material and current documentation, not every raw
campaign artifact. Whole-product reviews should reconcile/supersede stale tasks,
not accumulate append-only completion lists as competing roadmaps.
