# Product implementation proposal

Prepared 2026-10-22. **Proposal for review, not an approved backlog or a claim that
these features are implemented.** This reconciles the latest whole-product
usability review with earlier product, interaction, component and semantic plans.
It recommends a small delivery queue rather than implementing every historical
“next” paragraph.

**Recommendation:** finish the current evidence-continuity round, then make the
existing product compact and predictable before broadening it. Start with one
clear Explore entry, a clean saved-work boundary, and a revision-aware library.
Follow with useful general observability improvements—native-instrument browsing,
comparison leads and explicit capabilities—through those same workflows.

Follow-up [architecture](ARCHITECTURE_REFACTORING_REVIEW.md) and [design](DESIGN_REFACTORING_REVIEW.md) reviews identify source-backed refactoring and consolidation opportunities for these milestones.

## Basis, authority and current boundary

This is a documentation/source reconciliation, not another live usability study or
customer research. The newest [usability review](WHOLE_PRODUCT_USABILITY_REVIEW.md)
provides live desktop/mobile observations at the `290e101` baseline. Earlier
reviews retain their historical status. Current contracts, runtime code and
[actual validation](../VALIDATION.md) take precedence over an older proposed
storyboard or feature inventory. The imported [handoff](../START_HERE.md) and
original design examples are background, not new user instructions.

The durable direction comes from [product essence](PRODUCT_ESSENCE.md),
[project direction](../PROJECT_DIRECTION.md), and the user's requests: general
observability over ordinary OTel; Go/Rust semantic implementation; reusable
code-managed and interactively authored definitions; framework-neutral passive
components; explicit evidence and clocks; maintained executable documentation.
A manifestation of that direction must work for non-Go applications without a
Goo manifest. Go-specific weaving remains an optional instrumentation path.

The current round delivers native retained-evidence search, document citations
pinning an exact manifest/review/read basis, and a browser retry journal for native
retention/review saves. Final release validation now passes: real XTDB root checks,
frontend and race checks, Hegel state/swarm, real OTLP journeys 99–101, Quint
rejecting controls, a bounded Bombadil campaign, expanded mobile catalog checks,
and Go/XTDB restart recovery. The [validation record](../VALIDATION.md) contains
the exact image and evidence; milestone 0 is closed for this round. This does not
approve the subsequent proposed milestones.

The journal is browser-local, native-only, bounded to 16 entries and 16 KiB per
request, and does not coordinate cross-tab edits. Ordinary document drafts remain
page-local. Native retained search remains separate from the file-backed global
tag catalog. Server-owned pending work, authentication and ML export are not
part of this delivery.

## Reconciliation: what not to rebuild

| Historical recommendation | Current disposition | Actual work still relevant |
| --- | --- | --- |
| Automatic service and distributed dashboards, [whole-product review](WHOLE_PRODUCT_REVIEW.md) | Bounded service overviews, namespace-aware identity, declared-operation cards and reported-edge populations are delivered. | Compact their presentation; instance navigation, typed async relations and a deliberate whole-dashboard capture contract remain. A new Home dashboard is unnecessary. |
| Resumable trail, service diagnosis, chart selection, trace breakdown, context and capability guidance, [competitive survey](CORE_EXPERIENCE_PRODUCT_SURVEY.md) | All six have bounded first slices in [core refinements](CORE_EXPERIENCE_REFINEMENTS.md). | Durable cross-reload resume, broader scope/inspector return, real capability resolution and additional relation/coverage contracts. Do not start six new feature projects. |
| General cross-service timelines and saved population views | Delivered in [population workspace](POPULATION_WORKSPACE.md), following [linked service timelines](LINKED_TIMELINE.md). Some earlier plan paragraphs still list them as next. | Arbitrary native-stream composition, broader view layouts and saved dashboard policy—not another record population workspace. |
| Table, facets, typed attributes, waterfall, chart frame, timelines, heatmaps, evidence summary and selected-item shell, [component audits](WHOLE_PRODUCT_COMPONENT_SWEEP.md) | Existing [display library](UI_COMPONENTS.md) implements these bounded primitives/compositions. Some older audits still call them proposed. | Adoption gaps, shared-scale small multiples, revision-aware rows, comparison/result composition and consistent field labels. Extract only a proven repeated contract. |
| SSE/live exploration, [initial component audit](UI_COMPONENT_AUDIT.md) | [Log SSE tail](LIVE_LOG_TAIL.md), rolling record comparison and exact-stream native counter comparison exist. | Durable ingestion cursors/replay, trace tail, live gauges/histograms and fleet semantics are separate gaps. Do not describe all streaming as absent or move subscriptions into elements. |
| Full stored group comparison and readable geometry, [comparison review](COMPARISON_PRODUCT_REVIEW.md) | [Captured breakdowns](COMPARISON_BREAKDOWNS.md) and [visual comparisons](VISUAL_COMPARISONS.md) are delivered. | Error/latency ranking, log-appropriate columns, numeric comparison support and a common presentation shell. Statistical attribution is not implemented. |
| Numeric-looking attributes are strings / arbitrary measures absent | Declared int64/uint64 bindings, invalid-input coverage and Go/ClickHouse execution exist in [numeric planning](NUMERIC_ATTRIBUTE_PLANNING.md). | Floating/decimal expressions, multiple measures, unit conversion, numeric cohorts and broader backend lowering. Never restore coercion-to-zero behavior. |
| General captures, baseline-pair persistence, narrative and question cards, [collaboration research](COLLABORATIVE_INVESTIGATIONS_RESEARCH.md) | [General captures](GENERAL_CAPTURES.md), [saved comparisons](SAVED_COMPARISONS.md), typed steps and linked capture histories exist. | A cohesive case outline, durable drafts, authenticated editing and a shared activity/audit contract. Earlier “baseline persistence remains future” text is stale for supported comparisons. |
| Shared tags/notes and exact displayed-record freezing, [annotation design](PRODUCT_ANNOTATIONS_DESIGN.md) | Shared typed metadata and optional saved/live-page retention are delivered at `290e101`; native catalog/citations/journal are delivered in the current round, with final release checks outstanding. | Other signal fragments, task-specific reviewed labels, historical cutoff queries, dataset/export policy and cross-store search. Do not rebuild the shared assertion store. |
| Native bitemporal store evaluation, [storage spike](ANNOTATION_STORAGE_SPIKE.md) | Optional XTDB adapter is implemented and tested locally; normal metadata remains file-backed. | Measured operational/conformance work, multi-node visibility and explicit historical bases. No evidence yet justifies migrating all metadata, adding Minigraf in parallel, or building a ClickHouse external-source connector. |
| Databricks translation, [compatibility research](DATABRICKS_METRIC_VIEWS.md) | [Offline prepared-fact compiler](PORTABLE_COMPILER.md) emits bounded metric-view artifacts; this is not raw-OTel or live-warehouse parity. | Extend a typed supported subset, preparation/provenance and physical data conformance. Local artifact checks cannot establish Databricks execution. |

`IMPLEMENTATION_PLAN.md` and `KNOWN_ISSUES.md` mix historical milestones with
current limits. In particular, their broad old claims about missing live tail,
general saved contexts and arbitrary numeric measures must be qualified by the
implemented subsets above. Historical Python compatibility issues M01–M07 are
not a description of the newer Go kernels. Preserve the history, but stop using
it as an undifferentiated active queue.

## Proposed delivery queue

Milestone 0 closes the already-authorized delivery. Milestones 1–6 below are the
recommended sequence for subsequent approval. Each should ship
as bounded changes; the whole list is not one rewrite.

### 0. Close the current evidence-continuity release

**Outcome:** an exact live selection survives uncertainty, can be found again,
and can support an investigation without silently following later interpretations.

Complete the remaining mobile-expanded-catalog check and final documentation/
health checks. Preserve the now-tested behavior: catalog tags follow current
active revision heads while a saved document continues to cite its old exact
review; a lost initial save or correction response recovers across actual reload
with its original request key. State the bounds of browser storage, source-page
expiry, optional XTDB and file-backed metadata clearly.

Acceptance is the real OTLP journey, native integration, Hegel journal state/swarm,
Quint observed traces with rejecting controls, Bombadil and regenerated frames.
File/document drafts outside the native journal remain a distinct follow-up; do
not advertise recovery for all writes unless those paths are idempotent too.

### 1. Make first arrival and recent activity compact

**Outcome:** “What is happening now?” has an obvious first action and useful
results before the control scaffolding.

- Give initial Explore one primary action. Default to recent activity; retain a
  clearly named advanced fixed-question path, exact shared links and source-free
  staging. Do not restore competing executors under new labels.
- Use one compact applied-scope/result-action strip. Keep active filters,
  draft/applied distinction and decisive unknown/partial/gap state visible; move
  repeated exact provenance into one disclosure per owning result.
- Show a bounded primary activity chart set. Reveal range editing through
  selection; make additional measures and exact native stream plots deliberate.
  Move mobile facets into a disclosure or drawer instead of stacking them before
  every record. Display human-scale axis ticks with exact values retained.
- Make the row/inspector interaction dominant: compact trace continuation,
  controlled selection and consistent keyboard focus. Remove duplicate inline
  attribute bodies only after inspector parity is proven.

**Acceptance:** with pinned representative fixtures, the first useful visual
appears in the initial viewport and the first record targets roughly 1000px or
less on desktop/mobile. This is a layout target, not a universal performance
claim. No count, coverage warning, scope or typed metric semantics is hidden to
meet it. Fixed links still do not execute on open; disclosure and local filtering
perform no source reads. Test keyboard/touch cancellation, delayed completions,
selection pause and retained-parent return. Regenerate recent/service frames.

**Dependency:** none beyond the current release. This is the first recommended
next round, split into Explore compactness and then service/native adoption.

### 2. Separate live work from saved work and unify the actions

**Outcome:** users know whether they are changing a question, reading telemetry,
preserving evidence, citing it or publishing a narrative.

Remove the live `#operations` editor from Saved investigations; `workspace.js`
currently reparents it there. Keep legacy operation receipts reopenable through a
read-only presenter and an explicit Continue in Operations/Explore action. Keep
Operations as a route until ordinary service/definition workflows cover its
capabilities; do not delete supported metrics or contracts merely to reduce tabs.

Within Saved investigations, use views for investigations and evidence plus a
visible nonzero pending-save count. Compose one selected narrative/evidence area
instead of expanding every receipt. Reuse the new retained catalog; a federated
library may adapt file/native stores without migrating their payloads or inventing
a common authoritative latest head. Explain whether a search is local to a page
or indexes the supported catalog, and which clock/type it filters.

Adopt the newest [action vocabulary](WHOLE_PRODUCT_USABILITY_REVIEW.md#a-small-action-vocabulary):
Save view/definition preserves meaning; Capture question rereads; Retain selection
copies exact held observations; Add to investigation stages or explicitly creates
a child of a chosen saved revision; Save investigation revision publishes the
narrative; Retry reuses the same intent. Rename legacy **Save investigation** when
it actually saves an operation capture. Move new-query lookback defaults into
relevant live scope controls so frozen evidence never appears retimed.

**Acceptance:** old URLs and artifact hashes still open; Saved investigations
contains no unrelated live editor; all navigation/reopen-only paths make zero
telemetry reads. Preserve dirty drafts, source-free Back and exact citation
identity during route changes. Do not add a second persistence model or a new
permanent route for pending saves.

### 3. Make definitions and native instruments discoverable

**Outcome:** an operator can find a reusable question or an emitting instrument
without first learning authoring schemas.

Use compact revision-aware catalog rows across record questions, native
measurements and contracts. Group only by explicit family identity; identical
titles do not establish identity, authority or lineage. Provide type, environment,
origin and support filters, exact revision access, and a purpose-oriented preview.
Keep differing schemas behind typed adapters. Preserve advanced YAML/source
inspection and package promotion, but put them after the useful operational result.

Lead native entry with Find an instrument / Use a saved measurement, using current
discovery and validation. A metrics-only service and custom producer must remain
useful. Show complete stream identity, unit, temporality and coverage; an observed
name is not a verified convention profile or instrumentation-health verdict.
Use service → instrument → question → evidence → investigation as the same core
journey. A log-only service must not default to an empty spans-only result.

**Acceptance:** exercise hundreds of same-title revisions and disjoint bindings,
units and namespaces; distinguish zero local matches, empty catalog and read
failure. No auto-run, auto-subscribe or silently selected “latest” revision on
open. Native unsupported kinds remain visible with a useful explanation. Reuse
passive rows/tables/facets and adopt each extraction in two real surfaces.

**Dependency:** use the action/scope conventions from 1–2; catalog layout and
native browse-first entry can ship as separate increments.

### 4. Improve population isolation and comparison within the same canvas

**Outcome:** operators find a distinct population and understand what differs
without opening another analyzer.

First add error/latency ranking beside population-share ranking, hide inapplicable
log columns, and share a two-sided comparison composition. Reuse existing complete
stored group aggregates, native compatibility checks and exact side-specific
follow-up questions. Preserve full totals and explicit one-sided/missing/empty
semantics when local search hides rows.

Then add a bounded supported column/field chooser and clearer local column order.
Fetching fields not already retained is a separate explicit read. Extend the
existing selection owner to time × duration rectangle selection with keyboard
forms. Reuse Include/Exclude, retained children and comparisons; do not build a
second selection tray or expression evaluator. Stage explicit inspector widening
(service/time/predicate removal) with an honest change preview and return path.

“Selection versus rest” needs an exact Go complement inside the parent predicate,
including OR/missing-value semantics and overlap policy. A later What differs?
view can rank bounded dimensions by counts/share, coverage and minimum support;
call these descriptive leads. Do not label a large relative change root cause,
statistical significance or an unbiased anomaly.

**Acceptance:** generated transforms preserve parent scope and existing predicate
trees; local column/search/plot changes do not alter totals or issue queries.
Equal values share geometry where units/domains are compatible. Native histogram
cells do not create contributing traces; signed values, percentile ranges and
unequal windows are never forced into misleading comparable bars.

**Dependency:** 1–2; the ranking/log-column polish can ship independently of the
new selection complement and rectangle semantics.

### 5. Add explicit OTel capabilities and missing operational evidence

**Outcome:** defaults explain what they can answer and what evidence is missing,
then offer a useful existing question or instrumentation task.

Advance the [core primitives](CORE_PRIMITIVES.md) in vertical slices, preserving
current wire formats and the specialized record/native/lifecycle kernels:

1. Define machine-readable supported/unsupported/needs-binding outcomes for a
   small pinned HTTP server profile and a generic custom-data fallback. Separate
   static backend support, observed fields and verified coverage. Preserve original
   fields and require declared legacy/current compatibility mappings.
2. Apply the capability contract to one additional real use case, such as database
   pool pressure or messaging processing. Carry resource/scope/schema identity,
   span kind, units and explicit denominator. Client latency is not server USE;
   processing spans are not unique delivered messages or end-to-end outcomes.
3. Type relations independently: parent edges, async links, exact trace/span IDs,
   validated exemplars, scoped business identities and service/time context.
   First deliver explicit trace-window widening and anchor-retention diagnostics;
   then add async/exemplar continuations only when source fields are captured.
4. Add a separate pipeline-evidence source for actual collector/exporter queue,
   drop, sampling and lag facts. Event age or an empty query alone cannot establish
   pipeline health. Display this evidence beside existing unknown coverage.

Also retain typed measure work: multiple/integer/decimal measures, compatible
ratios, numeric cohorts and explicit conversions should use Go operators and
capability refusal, not browser arithmetic. Add a new native kind only for a
concrete discovery/browse/inspection journey with reference/source conformance.

**Acceptance:** pinned stable/legacy/mixed/custom OTel fixtures, missing and
present-empty identity, version/unit/type/temporality conflicts, async/batch cases
and insufficient coverage. Generic exploration survives unrecognized conventions.
Cross-signal pivots disclose exact versus contextual relations and scope changes.

**Dependency:** source/model work may proceed alongside UI 3–4 in isolated
increments; it is not a prerequisite to compact the current truthful results.

### 6. Extend durable continuity where an actual workflow requires it

**Outcome:** strengthen the foundation for longer investigations before adding
multiplayer or classifiers.

Prioritize idempotent publication/durable recovery for remaining document and
citation workflows, then explicit historical read cutoffs and cross-client
visibility. Await tokens are lower bounds, not frozen historical snapshots.
Keep authored applicability, event time, source read, immutable custody and database
system time distinct. Maintain optional XTDB and test backups/rebuild/restart and
reader lag before a general store migration.

Add exact trace fragments, native intervals/cells or tail observations one target
capability at a time. Retain every evaluator prerequisite (for example a counter
predecessor), original source/basis and bounded membership. Unsupported evidence
must not acquire a label through a generic Annotate button. Cross-page event
identity cannot come from payload hashes or reported trace IDs alone.

Only then choose one concrete reviewed-label task: version its vocabulary and
target grain, support abstention/disagreement/withdrawal, and export deterministic
pinned datasets with sampling, leakage and split provenance. Descriptive tags,
experiment intent and model suggestions remain separate from reviewed labels.
No classifier runtime or training platform is required for this increment.

**Dependency:** current release, clear saved-work composition and a demonstrated
user task. A large training-label suite is not the next automatic milestone.

## Parallel tracks and gates

| Track | Useful bounded work | Gate / why separate |
| --- | --- | --- |
| Shared deployment | Authenticated identity, tenant/workspace isolation, source-scoped credentials, rights on documents **and referenced evidence**, query concurrency/budgets, quotas/GC, redaction/deletion policy, audit and backup/restore. | Must precede exposing this lab as a multi-user/shared service. Do not wait for a whiteboard to start this if deployment is requested. Native storage or presence does not supply authorization. |
| Portability and IaC | Extend one typed supported subset from normalized OTel/prepared facts to ClickHouse SQL and Databricks metric-view artifacts; produce direct/requires-preparation/unsupported reports and deterministic plans. Retain YAML authoring and explicit reviewed instrumentation handoff. | Offline planning, fixture/reference parity and artifact checks need no Databricks instance. Actual SQL/ingestion/runtime claims require one. Do not imply row search, custody or lifecycle preparation is automatically a metric-view measure. |
| Semantic operational features | A durable SLO/burn-rate evaluator with mature cohorts, reset/coverage policy, scheduler, state and notification delivery; narrowly scoped native-kind or declared-fleet improvements. | General observability needs these eventually, but they are real execution/storage projects. Prototype RED charts are not a production SLO. Pick one end-to-end contract instead of a menu of scorecards. |
| Developer and test integration | Maintain optional Go weaving/annotations, non-Go OTel examples, verified experiment provenance and correlated app/UI/model journeys. | Instrumentation declarations do not prove runtime conformance; Bombadil outcomes and synthetic experiment intent do not become reviewed labels. No new sample-only feature branch. |

CEL versus KCL should follow the typed expression/cost/portability requirements;
there is no need to revisit the imported EDN proposal. DataFusion could later
execute or optimize a lowered plan, but it does not replace identity, coverage,
window, evidence or review semantics. Evaluate it only for a specific backend
integration or measured query-planning problem—not as a rewrite prerequisite.

## Consolidated remaining backlog

This register groups repeated proposals; the milestone columns prevent parallel
reviews from creating duplicate queues. “Later” is deliberate deferral, not denial
that a feature could be valuable.

| Work | Disposition / first useful boundary |
| --- | --- |
| Compact entry, one primary action, mobile facets, bounded charts, human axis units | **Next: 1.** Latest observed whole-product friction. |
| Saved/live separation, action vocabulary, compact case outline and relevant time editor | **Next: 2.** Preserve all historical IDs and capabilities. |
| Native retained catalog, pinned review citation, native retry journal | **Delivered this round; release gate 0.** Finish final checks; do not reschedule. |
| Cross-store search, broader document/draft recovery, search/index quotas | **Partial: 2/6 and deployment gate.** Native catalog is not a global file/native index or durable resume for every draft. |
| Revision-aware definition library, native browse-first entry, contextual example entry | **Next: 3.** Keep exact revisions and typed bindings. |
| Remaining table/stat/chart adoption, small multiples, field tokens, comparison shell | **Consolidate in 1–4.** Reuse existing elements; do not create seven new generic components. |
| Error/latency ranking, log-appropriate columns, numeric cohort compatibility | **4.** Immediate useful comparison refinements. |
| Supported arbitrary columns, chart rectangles, inspector widening, exact parent complement | **4.** Typed bounded operations; no generic predicate algebra project first. |
| General anomaly/difference scoring or significance | **Later.** Start with descriptive support/coverage-aware leads; scoring needs explicit baseline/sampling/multiple-testing policy. |
| Native arbitrary stream composition, summary/exponential histograms, float/decimal expressions | **Demand-led 3/5.** One usable source/reference/UI path per increment. No fleet percentile averaging. |
| Instance navigation, async relationships, metric exemplars, trace anchor/coverage diagnostics | **5.** Reported edges/instance diagnosis already exist; add actual relation contracts before graphical topology. |
| Collector health/drop/queue/sampling evidence and query cancellation/concurrency budgets | **5 + deployment.** Do not infer collection health from absence. |
| Durable tail cursor/replay, trace tail, live gauges/histograms, fleet live comparison | **Later after 5.** Separate identity, membership, reset, gap and replay contracts; existing log tail remains supported. |
| Explicit membership activation windows, common native windows and fleet policies | **Demand-led semantic track.** Existing fixed declared membership is not automatic fleet discovery. |
| Whole-dashboard persistence and saved arbitrary layouts | **Later.** Define component custody, partial failures and independent bases before a dashboard snapshot label. First use existing saved views/case citations. |
| Historical metadata/citation cutoffs, review relations and exact native/trace/tail fragments | **6.** Separate from await/read-your-write and from generic storage migration. |
| Reviewed task labels, adjudication, dataset JSONL export and campaign ingestion | **6 after a concrete task.** Pinned membership, vocabulary, origin, retention and leakage policy first. |
| Label rules, model suggestions, active learning, inline inference | **Defer.** Versioned derivations and human review boundary required; no automatic ground truth. |
| Deployment-marker correction/retraction and verified CI ingestion | **Demand-led 5/6.** Existing local context lane is delivered; authored context never proves causal release impact. |
| User-authored workflows, plan/apply/deployment and richer IaC authoring | **Portability/developer track.** Reuse typed steps/packages rather than add an independent workflow language. |
| Raw-source preparation/warehouse lowering and physical data transfer | **Portability track.** Offline subset next; warehouse execution remains a later validation gate. |
| Broad metadata migration, native ClickHouse→XTDB connector or second Minigraf backend | **Defer pending evidence.** Current direct selected-record retention addresses the demonstrated workflow. Benchmark/ops/conformance need, not backend novelty, should drive adoption. |
| Presence, shared editing, comments, whiteboard/CRDT canvas | **Defer until deployment gate and demand.** Ordered investigations are primary; spatial composition must reference the same exact objects. Honeycomb multiplayer was not verified by the research. |
| RUM, session replay, profiling/security suites, generic dashboard builder, plugin marketplace, autonomous root-cause agent | **Not in the current core queue.** Prefer safe contextual links/integrations if a user journey needs one. No vendor feature checklist. |
| Specialized reservation/Redis compatibility lab and its M01–M07 issues | **Contain and decide, not generalize.** Keep visibly non-production and testable; repair only needed guarantees or retire the compatibility surface after parity. Do not import its bounded preview arithmetic into Go. |
| Parallel new sample applications / another standalone tail-to-investigation panel | **Remove from the queue.** Extend executable examples through the normal product instead. |

## What to remove, keep or stop doing

Remove duplicate controls and default chrome once an equivalent, tested workflow
exists. Stop placing live operation execution in saved evidence. Stop presenting
raw backend names, receipt formats and long IDs as the everyday task vocabulary;
keep exact values in inspection. Stop documenting already-delivered first slices
as wholly missing. Stop adding independent permanent workbenches for each new
visualization, native kind or research idea.

Keep existing routes as compatibility entries until a migration proves parity.
Keep exact artifacts and old schemas readable. Do not delete historical receipts,
merge semantic package types by title, manufacture a latest head, remove unknown
states, hide unsupported instruments, or replace the typed Go kernels with a
universal UI evaluator. Do not remove advanced capabilities simply because their
current presentation is dense.

## Validation and documentation policy for every milestone

Use the existing real user journeys rather than another standalone demo. Maintain
an unfamiliar-data path and a native-instrument path through service/selection →
inspection/comparison → evidence → investigation → reuse. Keep the sample app,
Goo UI and optional model/experiment expectations correlated where they actually
share declared evidence.

- **Playwright:** real OTLP/source evidence, desktop/mobile frames, keyboard/focus,
  meaningful cold/populated/failed/stale/expired states, no hidden reads on local
  actions, and old artifact/deep-link reopen. Test representative catalog density.
- **Bombadil:** bounded actions plus reachability for the new controls, generation
  and target continuity, no implicit writes/subscriptions, partial-state visibility,
  focus/viewport and recovery. Browser exception checks alone are insufficient.
- **Hegel state/swarm:** actual typed transforms, parent predicates, revisions,
  immutability, retry/correction interleavings, missing/empty identity and
  unit/temporality refusals. Do not property-test CSS or mirror implementation.
- **Quint/quint-go:** extend existing execution/selection/publication models for
  changed protocols, replay actual traces and reject controls for hidden reads,
  stale completion, altered evidence, lost intent or invalid relation/tenant scope.
  Pure layout/wording changes need appropriate browser checks, not a new model.
- **Docs:** each active item links to its owning current contract and last evidence;
  diagrams/storyboards/screenshots change with behavior. Label proposed work,
  historical snapshots, skipped checks and unavailable integrations explicitly.

Create one compact active status index linking to this proposal and the current
contracts; retain the long milestone history as an archive. Use statuses such as
proposed, in progress, implemented-with-limits and deferred. Do not rewrite old
research to imply it was implemented then, and do not copy every old “next” list
into that index.

## Decisions and proposed defaults

No additional decision is required to recommend milestone 1. The existing product
essence and latest review provide a clear compact-first default. This document
itself does not authorize implementation of subsequent milestones.

Three choices materially affect later scope:

1. **Shared deployment:** is the next real target a personal/local workbench or a
   shared team installation? Default to the current local boundary; if shared is
   requested, start its authorization/isolation/retention gate alongside UI cleanup.
2. **First review/training task:** which concrete target and task should be supported
   first—record examples, native intervals or investigation retrieval/evaluation?
   Default to no formal task vocabulary/model runtime until there is a real use case.
3. **Compatibility-lab retirement:** should the reservation-specific Python UI remain
   a teaching/compatibility surface long term? Default to keeping it secondary and
   clearly bounded, rather than spending general-product work repairing every
   legacy preview or deleting a useful runnable example without parity.

Everything else above is sequenced implementation judgment. No need to choose a
whiteboard framework, universal dashboard schema, CEL/KCL runtime, query engine
or new metadata backend before simplifying the existing product.
