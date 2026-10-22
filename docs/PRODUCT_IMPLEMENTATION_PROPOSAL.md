# Product implementation proposal

Prepared 2026-10-22; integrated after the architecture/design reviews. **The user
has authorized working through the bounded active queue below. This is not a claim
that queued features are implemented, or approval of every long-term research idea.** This reconciles the latest whole-product
usability review with earlier product, interaction, component and semantic plans.
It recommends a small delivery queue rather than implementing every historical
“next” paragraph.

**Current recommendation:** follow the user’s revised stream-first direction:
fresh Explore opens on compact live observations, with scope and inspection
shaping the question naturally. Preserve the completed evidence-continuity and
live/saved separation work, then continue reusable compositions and compact libraries.
Follow with useful general observability improvements—native-instrument browsing,
comparison leads and explicit capabilities—through those same workflows.

The [architecture](ARCHITECTURE_REFACTORING_REVIEW.md) and [design](DESIGN_REFACTORING_REVIEW.md) findings are incorporated in the single active queue below. Their original recommendation status is historical; they do not create parallel refactoring backlogs.

Current acceptance reconciliation: [S2/A3 audit](S2_A3_ACCEPTANCE.md). Bounded S2/A3 is delivered, including explicit/lazy fixed authoring. Bounded [B1 library consolidation](DEFINITION_LIBRARY.md) is delivered across Definitions, fixed reusable questions, document revisions and retained catalog metadata, with retention owner separation. B2 protocol/qualified-field consolidation is next. Comparison presentation remains D1, rather than an ad hoc next step.

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
imply that the subsequent queued milestones have shipped.

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
| Shared tags/notes and exact displayed-record freezing, [annotation design](PRODUCT_ANNOTATIONS_DESIGN.md) | Shared typed metadata and optional saved/live-page retention are delivered at `290e101`; native catalog/citations/journal are delivered and release-validated in the current round. | Other signal fragments, task-specific reviewed labels, historical cutoff queries, dataset/export policy and cross-store search. Do not rebuild the shared assertion store. |
| Native bitemporal store evaluation, [storage spike](ANNOTATION_STORAGE_SPIKE.md) | Optional XTDB adapter is implemented and tested locally; normal metadata remains file-backed. | Measured operational/conformance work, multi-node visibility and explicit historical bases. No evidence yet justifies migrating all metadata, adding Minigraf in parallel, or building a ClickHouse external-source connector. |
| Databricks translation, [compatibility research](DATABRICKS_METRIC_VIEWS.md) | [Offline prepared-fact compiler](PORTABLE_COMPILER.md) emits bounded metric-view artifacts; this is not raw-OTel or live-warehouse parity. | Extend a typed supported subset, preparation/provenance and physical data conformance. Local artifact checks cannot establish Databricks execution. |

`IMPLEMENTATION_PLAN.md` and `KNOWN_ISSUES.md` mix historical milestones with
current limits. In particular, their broad old claims about missing live tail,
general saved contexts and arbitrary numeric measures must be qualified by the
implemented subsets above. Historical Python compatibility issues M01–M07 are
not a description of the newer Go kernels. Preserve the history, but stop using
it as an undifferentiated active queue.

## Active implementation queue

The user has authorized incorporating the product, architecture and design reviews
and working through the resulting tasks. This is the single active queue; the
reviews provide rationale, not additional independent backlogs. Complete each
bounded cut and its evidence before taking the next. Later capability-dependent
research remains deferred, rather than becoming an automatic commitment.

| Cut | Deliverable and deliberate consolidation | Acceptance before advancing |
| --- | --- | --- |
| 0 — closed | Native retained catalog, exact review citations and browser-local native retry recovery. | Delivered evidence is recorded in [VALIDATION.md](../VALIDATION.md). Preserve native-only recovery, distinct store/search clocks and pinned historical reviews. |
| A1 — delivered applied-result portion; blank entry superseded by S1/S2 | Compact Explore: extract the existing distribution/facet body from enclosing workflow chrome; compose one applied scope/result/action region; disclose mobile facets; remove redundant hidden controls and their CSS overrides after parity. | Real desktop/mobile useful visual and first-record positions; stream-first entry follows S1/S2 below; keyboard/touch, staged/applied scope, explicit partial/unknown state, no reads on local disclosure, delayed response rejection. Refresh affected storyboard frames and bounded Bombadil actions. |
| A2 — delivered | Clean live/saved ownership: leave the live operation editor in Operations; use a read-only operation receipt presenter in Saved with explicit continuation. Extract pure versioned route/link/step helpers where this change touches them. | Old hashes, IDs and receipt reopen still work; navigation/reopen does not execute sources; dirty drafts and retained-parent focus survive return; live and saved owners are not reparented. Go/JS fixtures protect exact bounds, missing/empty and optional defaults where protocol helpers move. |
| A3 | Apply the compact scope/result composition to one service surface; add a record/capture transformation beside existing `GooEvidenceScope` adapters and adopt it in documents/captures. Consolidate only the CSS owned by these compositions. | Two real adopted surfaces share display facts rather than a universal result object; changed-read and coverage caveats stay visible; clocks remain separate; standalone/nested mobile and SSR/pre-definition SPA component checks remain source-free. |
| B1 | Compact revision-aware library rows, initially record and native measurement definitions; then documents/retained catalog where compatible. Mount rich meaning/review detail only for selected entries. Split touched retention code into client/basis, catalog, citation and recovery owners without changing intents. | Same-title unrelated entries, explicit families, parallel heads, optional store unavailable, pagination/local search bounds, stale responses and exact old pinned review remain correct; selecting is not execution. Preserve retry body/key and await-token propagation. Mechanical formatting is a separate change. |
| B2 | Shared pure browser protocol fixtures and qualified-field/presence display helpers; adopt applied chips and definition summaries. Extract a tiny request-owner helper only after comparing two adapters with matching begin/invalidate/current/dispose mechanics. | Golden Go/JS conformance preserves family differences and stored identities; Hegel state/swarm plus observed Quint traces reject obsolete completion and hidden publication; parent restoration and draft guards survive. Do not add a global mutable store. |
| C1 | Freeze identity/response fixtures, then extract immutable-file publication mechanics only. Each store retains validation, digest/canonicalization, lineage and catalog rules. | Byte-identical IDs, large integers, duplicate publication, reopen, corrupt/truncated rejection, parent mismatch and no overwrite. Do not claim power-loss durability without fault evidence. Delete old mechanics after parity. |
| C2 | Inventory endpoint effects/policies and inject explicit startup dependencies; consolidate strict JSON/response helpers with per-route limits. Then extract bounded ClickHouse HTTP transport, one source family at a time. | Preserve method/status/error/body-limit compatibility, typed escaping, cancellation, non-2xx/oversize handling, family basis and limits. GET/comparison/capture execution is accounted for; SSE retains independent lifetime/no-store/disconnect semantics. Embedded and real-source journeys pass. No Python proxy removal in this cut. |
| D1 | Unify two-sided comparison presentation, complete fitting inspector adoption and share count-series geometry where two real consumers agree. Add meaningful error/latency ranking and log-appropriate columns. | Independently named windows/bases; incompatible units/kinds remain refused; exact table fallback, common domains, large values and known-to-unknown updates; explicit source continuation and keyboard return. Stored/reread/live execution owners remain separate. |
| D2 | Bounded population operations: supported column chooser/order, time-duration rectangle with keyboard equivalent, explicit scope widening and exact selection-versus-rest complement. | Typed Go operations preserve parent predicate, range/presence/overlap semantics and full totals independent of paging. Real journey captures/reopens selected evidence; Hegel properties and Quint rejecting controls cover changed execution/selection protocols. No generic predicate-algebra rewrite. |
| E1 | Native instrument discovery within existing entry points; explicit capability planning for one pinned HTTP-server convention plus generic fallback, then one database-pool or messaging case. Add typed relations and coverage only with an actual supporting source. | Ordinary non-Go, manifest-free and custom OTel data work; missing/empty identities, units, temporality, batch and unsupported cases are honest; static declaration/observed data/verified support remain distinct. No inferred trace completeness or fleet percentiles. |
| F1 — prerequisite-gated | Broaden recovery or exact retained targets only for one concrete supported write/fragment after its server idempotency and source-membership contract exists. Inject exact-reference resolution/clocks without making file document loading depend on XTDB. | Lost acknowledgements/reload preserve intent, IDs and distinct clocks; unsupported writes do not inherit native-journal promises; current-head search differs from pinned/history reads. Select a concrete review task before adding formal labels/export. |

The first delivered cut separates live/saved operation ownership (A2) and implements the applied-result portion of A1: facet-only composition, compact timeline controls and mobile facet disclosure. The remaining blank-entry plan is superseded by S1/S2 below. Actual checks and image identity are in [VALIDATION.md](../VALIDATION.md).

### Stream-first revision of the entry queue

The user's latest direction supersedes the earlier initial-entry “explicit start”
constraint: a true fresh blank Explore entry should begin with live observations,
not a question form. This does not authorize automatic execution of saved views,
shared/deep links, captures or definitions. The peer [product/design](STREAM_FIRST_ENTRY_DESIGN.md)
and [architecture](STREAM_FIRST_ENTRY_ARCHITECTURE.md) reviews define the proposed
interaction and feasible boundaries. This is revised intent, not a claim that the
stream-first UI has shipped.

Preserve the delivered `3fbefea` baseline: A1's applied-result facet-only/compact
composition and A2's live/saved ownership separation remain implemented. Replace
A1's remaining blank-entry action consolidation with the cuts below, ahead of A3.
A–F remains one queue; S cuts are its revised entry work, not another workbench.
The design review’s capability sketches are not another numbered backlog:
its compact entry maps to this S1, local context/filtering to this S2, exact custody
to this S3, and mixed-signal/population enrichment remains deferred below.

| Cut | Deliverable | Acceptance before advancing |
| --- | --- | --- |
| S1 — implemented with limits | Fresh blank Explore auto-connects one bounded active-signal monitor using an explicit valid configured/previous environment. Offer the already-supported Logs/Spans switch; reuse the existing `/api/tail` owner and passive record/attribute components. Put compact rows first, with a single scope/connection strip, chosen facet pills and honest buffer statistics. Keep question/capture forms below or in contextual actions. | Existing log and span selectors retain their typed source contract. One active connection and no prerequisite population/aggregate read; abort reader, reconnect timer and watchdog on leave/hide/signal change; old generations cannot publish. Missing environment prompts scope selection, never silent all-environment reads. Saved/deep-link routes remain inert. Focus/inspection and intentional pause persist on return; scrolling away pauses transport as well as follow, preserving the buffer/anchor without hiding continuity warnings; resume is explicit and shows the new-window gap. Show empty/loading/stale/limited/gap/disconnected distinctly. Real logs and spans, desktop/mobile, keyboard, bounded Bombadil actions and source-request counts validate the entry; 429/source failure retains last-good rows with unavailable/gap state, not healthy emptiness. |
| S2 — bounded interaction and fixed-editor mount gate delivered | Stream interaction shapes scope: local inspection with minimal qualified attributes, chosen environment/facet pills, staged selector edits and explicit Apply starting a new generation. Show trace/span ancestry only from actually held exact IDs/parent relations, with explicit missing/ambiguous-parent, cycle and depth-limit states; deeper context reads remain named actions. Expose “Investigate this window” as a deliberate fixed-question reread, and keep shared inspector/retained-parent return. | Never mix old/new selectors or treat current buffer as full population. No hidden reads/publications from local inspection, breadcrumb rendering or staging. Paused anchor/focus survives updates and navigation; resume is explicit, starts a new bounded read and shows a continuity gap. Reread wording distinguishes fresh results from held rows. Hegel state/swarm and Quint observed traces/rejecting controls cover owner transitions, scope drift and obsolete publication. |
| S3 — capability-gated | Exact retained tail selection only after a server-owned capability proves connection membership, canonical records, per-row read basis/window, expiry and retry identity. Reuse existing retention/citation contracts where their prerequisites genuinely fit. | Current connection-local observation handles are not durable event IDs or a live-page `resultToken`. Do not call the native page-retention API with fabricated membership or substitute a rerun. Test duplicate/late rows, eviction, expiry, changed selector, lost acknowledgement and correction before advertising exact capture. Until then offer explicit reread or existing supported evidence paths. |

Start S1 with one selected signal. A coordinated “Both” view, separate global
population totals, per-service concurrent subscriptions, ingestion replay and
automatically derived ancestry are not prerequisites. Existing tail is bounded
rolling event-time polling, not an ingestion cursor: say “live observations,”
name retained/evicted/observed as local counts, and retain source limits/gaps.
Basic severity/error/duration summaries are buffer summaries only, never fleet
percentiles, complete request/error rates or application-health proofs.

Reuse the compact distribution body, applied scope facts and inspectors already
landed. The application owner controls startup, cancellation, source visibility,
filter generations and capture; custom elements remain passive. S1's real journey
must include a fresh entry that connects, a saved entry that does not, environment
retargeting, route/background disposal and no automatic restart of a paused stream.
Update the existing tail/recent-entry contracts, diagram and actual storyboard
frames rather than creating a second independent “firehose” application.

### Whole-product consolidation through the same flow

The subsequent [whole-product design](STREAM_FIRST_WHOLE_PRODUCT_DESIGN.md) and
[architecture](STREAM_FIRST_WHOLE_PRODUCT_ARCHITECTURE.md) reviews extend the
stream-first direction beyond the landing page. Their shared spine is **observe →
narrow → inspect → investigate a fixed scope → explain/preserve → reuse**. Services,
Operations, Definitions and Saved are alternate entrances to that flow, not new
copies of the complete query/editor/capture workspace. These recommendations are
assigned to the existing cuts below; they are not a second backlog.

The first S1 monitor is implemented and validated; see [Live Activity](LIVE_ACTIVITY.md)
and [release evidence](../VALIDATION.md). The configured lab environment is explicit;
remembered scope and deployment discovery are not implemented. The first S2 cut adds shared-viewer Include/Exclude, staged chips/Apply and explicit fixed-window staging. Paused-parent return now suspends fixed work and restores selection/scope/scroll without a read. Live/fixed summaries and detail bodies now share passive record presentation, including bounded supplied-parent context. Live, fixed/frozen and held service observations now use the shared record-table shell. Record receipt summaries now share a pure adapter between capture pages and document references; fixed continuation stages before Run. Saved record pages now have a dedicated owner without a hidden query editor; Live/fixed and Saved share record/context rendering while retaining independent state. Shared fixed/saved/context record tables now use one inspector instead of duplicate inline attributes, preserve refinement/selection actions and hide comparison buttons without a usable breakdown or retained baseline. Standalone tails now share cached record details, and native identity views share a lazy inspector/direct selected-point body. Native values/coverage now share presentation across measurement details, service summaries and fixed/live native windows. Fresh Activity now defers fixed-editor mounting until explicit authoring or a supported continuation. Bounded S2/A3 is delivered; other aggregate/native service layouts remain D1/E1 work.
The compact wrapper is a migration adapter, not the final whole-product architecture. The entry now consolidates environment/signal/Pause in one bar, reveals Apply for drafts, and collapses mobile navigation; first-record viewport bounds are tested.

| Existing cut | Whole-product migration and what it replaces | Deletion/parity gate |
| --- | --- | --- |
| S1 → S2 | Treat the compact stream wrapper and legacy-detail disclosure as a migration adapter. Extract a supplied-record row/context presenter and common inspector; use them for tail, fixed records and frozen saved records rather than mounting a full read-only query editor to inspect one record. | Keep fixed queries reachable explicitly. Delete duplicate record/detail bodies, competing action rows and CSS that hides their controls only after timestamps, typed attributes, trace continuations, selection, keyboard/mobile and exact source basis match. Saved record inspection makes no source read. |
| S2 + A3 | One application-level scope/mode/state/body/action composition, first across Activity/fixed exploration and one service-record result. Services/Operations offer explicit scoped Activity or typed fixed/native continuations. A workflow-local investigation context retains applied scope, execution mode, selected item and return anchor. | Live/paused/fixed/frozen modes remain visibly distinct; family AST/results, native semantics and immutable references are not flattened. Navigation consumes typed continuation intents instead of reparenting another owner's DOM. Back restores the paused parent without starting a stream or rerunning a source. No global mutable store prerequisite. Contextualize global `#minutes` to new fixed questions so it cannot imply the live 60-second horizon; avoid stacking outer record/measurement tabs and stream pickers as simultaneous modes. |
| A3 + B2 | Standardize the scope/mode bar, staged/applied chips, selected-evidence actions, busy/disabled reasons and qualified field labels. Extract only proven request-owner mechanics and versioned continuation conversions; keep execution, capture and draft transitions in typed owners. | One relevant primary action per scope, no simultaneous parent/child executor, generation-safe completion and no hidden execution on staging/disclosure. Extend Go/JS fixtures and existing Hegel/Quint models when protocol ownership changes. Replace old helper paths only after equivalent consumers migrate. |
| B1 | Definitions become compact browse/manage rows with a selected contextual editor; Saved has typed library sections and one selected artifact/notes body. Share rows, evidence headers and exact citation presenters through explicit adapters rather than repeating every catalog/editor on each page. | Same-title distinct identities, parallel revision heads, branch histories, disabled optional stores, bounded search scope, unsaved drafts and pinned old reviews survive. Selecting a row never runs it. Remove eager per-revision meaning/editor mounts and duplicate catalogs after selected-entry parity. Do not merge backend heads or migrate stores to obtain common layout. Meaning/tags and pending recovery appear beside the selected typed target, not as copied telemetry attributes or inferred training labels; native-only recovery stays native-only. |
| D1 | Reuse a two-sided comparison composition and selected-item presentation across stored, fresh and rolling comparisons; contextual baseline/capture actions appear at the owning comparison/selection. | Preserve independently named windows/read bases, compatibility refusal and unknown/gap states. Delete repeated comparison chrome and baseline/capture toolbars after action parity; keep distinct execution and publication protocols. |
| C1/C2 | Keep bounded publication/HTTP/transport extraction behind frozen identity/response fixtures. Whole-product control cleanup does not require a platform/service split or backend migration. | Existing cut-specific hash/error/SSE/source tests remain mandatory; these mechanics do not block S1/S2 or justify moving semantic kernels into a universal evaluator. |

Cross-surface acceptance follows real ordinary OTel data: fresh Activity → stage/
apply a facet → inspect without fetching → explicitly open fixed scope → compare/
preserve supported evidence → add an exact citation → reopen from Saved → reuse a
definition. Include alternate Services and Definitions entry points, unsupported
native capabilities, expired references, a failed optional store and old deep links.
Examples enter the same real Activity/fixed/native paths rather than another demo
workbench. Keep Operations in navigation until service/definition-context parity
exists for cross-service packages, expert use, sample workflows and direct links;
then reassess demotion with the old route preserved. A change is complete only when the superseded controls/compositions are removed
or have a named compatibility consumer and removal condition, affected actual
storyboards/diagrams/screenshots are regenerated, and tests distinguish local
inspection from execution and reread from exact custody. Do not hide uncertainty
or remove advanced functionality merely to achieve a cleaner screenshot.

Bounded S2/A3 is delivered; the next round is B1. S3 waits for its membership capability.
A1–A3 and S1–S3 are bounded cuts, not one unreviewable rewrite. B–F are subsequent
bounded cuts in the same queue. Backend mechanics C1/C2 may run alongside B only
with disjoint ownership and frozen fixtures; they do not block the first compact
release. Reorder later cuts when a real usage result changes their value, and
record the reason here instead of opening another “next” list.

For every cut: update its current contract, affected architecture/process diagram,
real user storyboard/screenshots, README/issue status and actual validation record.
Use Playwright for layout and actual source behavior, bounded Bombadil for reachable
actions/invariants, Hegel for meaningful semantic/state interleavings, and existing
Quint/quint-go replay plus rejecting controls for changed protocols. Cosmetic CSS
needs browser evidence, not mirrored semantic tests. Passive components must not
fetch, subscribe, navigate or publish. Remove superseded paths only after parity;
any temporary adapter needs a named consumer and removal condition.

### Explicit deferrals, not executable backlog

Shared-team authorization/isolation/quotas require a shared deployment target.
Formal review-task vocabularies, adjudication and dataset exports require a concrete
training/evaluation task and exact target/membership policy. Durable multi-client
history, cross-store global search, native metric/fleet live extensions, whole-
dashboard custody and additional numeric families need their own capability and
usage evidence. Keep local defaults meanwhile.

Do not start a framework migration, universal evaluator/result schema, DataFusion
integration, CRDT whiteboard, broad XTDB migration, Minigraf installation, ClickHouse
external connector, Python proxy/lab removal, or CEL/KCL runtime selection merely
to complete this queue. Preserve existing typed kernels, old artifacts, optional
Go weaving, non-Go OTel support and the runnable compatibility lab.

## Product rationale for the integrated queue

The original milestones 1–6 below explain product outcomes behind active cuts
A–F. Their wider possibilities are not separate tasks: follow the bounded cuts
and explicit deferrals above. Milestone 0 is closed.

### 0. Close the current evidence-continuity release

**Outcome:** an exact live selection survives uncertainty, can be found again,
and can support an investigation without silently following later interpretations.

Release checks are complete. Preserve the now-tested behavior: catalog tags follow current
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

- Fresh blank Explore now defaults to one bounded live observation connection
  under S1/S2 above; this supersedes the earlier explicit-start proposal. Retain
  advanced fixed-question access, exact shared links and source-free saved staging.
  A stopped/paused or saved route does not silently start a stream.
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
| Compact stream-first entry, mobile facets, bounded charts, human axis units | **Delivered bounded S1/S2.** Fresh Activity, applied-result compactness and explicit/lazy fixed authoring are implemented; see the acceptance audit. |
| Saved/live separation, action vocabulary, compact case outline and relevant time editor | **Next: 2.** Preserve all historical IDs and capabilities. |
| Native retained catalog, pinned review citation, native retry journal | **Delivered and validated: 0.** Do not reschedule. |
| Cross-store search, broader document/draft recovery, search/index quotas | **Partial: 2/6 and deployment gate.** Native catalog is not a global file/native index or durable resume for every draft. |
| Revision-aware definition library, native browse-first entry, contextual example entry | **Next: B1** for compact revision-aware definitions. Native discovery remains E1; keep exact revisions and typed bindings. |
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

## Decisions and current defaults

The user has authorized progressing through the integrated active queue. Start
S1/S2 with the revised stream-first and existing local deployment defaults; no further
confirmation is needed for those bounded changes. Broader capability-dependent
work remains deferred until its stated evidence or target exists.

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
