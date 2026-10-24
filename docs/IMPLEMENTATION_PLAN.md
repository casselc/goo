# Implementation plan

> Current priority/status index: [2026-10-24 roadmap](ROADMAP_2026_10_24.md) and
> [review-gated execution backlog](BACKLOG_2026_10_24.md). The long milestone record
> below is implementation history; conflicting old Next labels are superseded.


Updated 2026-10-23 after the user authorized incorporating the product, architecture
and design reviews and working through the resulting tasks. The
[consolidated product proposal](PRODUCT_IMPLEMENTATION_PROPOSAL.md) holds the
reconciliation, detailed rationale and source links. This active index replaces
competing historical “next” lists; the history below preserves past scope, not
additional current commitments.

## Active implementation queue

The bounded D1 log-comparison cut uses shared signal capabilities for live
selectors and stored breakdown columns; span-only error/latency controls are
omitted for logs. Missing captured span measures retain their unavailable state.
The bounded ranking continuation is now delivered: shared exact count/share/error/
latency ordering, support/coverage display and saved population metric persistence.
The bounded [comparison context composition](COMPARISON_CONTEXT.md) now shares side labels, modes, full counts, UTC windows and evidence disclosure across fixed, stored/fresh/saved and live owners. Automatic fixed-pair results omit duplicate capture selectors; [contextual actions and retained side inspection](COMPARISON_ACTIONS.md) now cover population child inspection, immutable stored-side records/native results and compact live commands. [Retained count-series geometry](COMPARISON_SERIES.md) now shares fixed and explicitly loaded stored/frozen capture charts with UTC/elapsed alignment and staged side inspection. [Rolling bucket transport and pinned selection](LIVE_COMPARISON.md#retained-bucket-series-during-refresh) now reuse the same chart with fixed domains, explicit child reads and exact-series Freeze. The bounded action adoption now also covers fresh previews and stored/saved commands, with receipt-only Reopen under contextual options and finding publication kept explicit. The [D1 acceptance audit](D1_ACCEPTANCE.md) now accepts the bounded comparison milestone and records its deliberate limits.

The bounded [record-field layout cut](RECORD_COLUMNS.md) adds supplied-row field
discovery, local show/hide/order, staged Include/Exclude and saved population
layouts. Reopened saved population records now use the same staged facet and
child-population owner as recent records. Arbitrary transforms, source-wide schema discovery and persisted live
preferences remain outside the [bounded D2 acceptance](D2_ACCEPTANCE.md).

The [2026-10-22 whole-product coherence proposal](WHOLE_PRODUCT_COHERENCE_PROPOSAL_2026_10_22.md)
is a fresh architecture/design review of `f54d07c`. It proposes a contextual
Activity-to-analysis handoff and population workflow ahead of broad comparison
consolidation, plus a simpler Saved entrance. The user accepted continuing with the contextual handoff. Its bounded
[Activity analysis cut](ACTIVITY_ANALYSIS.md) is delivered. Population workflow
is partially implemented through the contextual selection/comparison/restoration
cut below. Bounded [Saved task composition](SAVED_WORKSPACE.md) is also delivered;
broader population controls, case outlines and cross-store search remain follow-up.

The user has authorized incorporating the product, architecture and design reviews
and working through the resulting tasks. This is the single active queue; the
reviews provide rationale, not additional independent backlogs. Complete each
bounded cut and its evidence before taking the next. Later capability-dependent
research remains deferred, rather than becoming an automatic commitment.

| Cut | Deliverable and deliberate consolidation | Acceptance before advancing |
| --- | --- | --- |
| P3 — bounded Saved task composition delivered | Investigation/evidence browsing first, selected evidence and retained notes draft, secondary ID/tools forms and relevant pending-save notice. | [Contract and journey 117](SAVED_WORKSPACE.md). Preserve exact references, source-free browsing, draft guards, optional-store failure and original retry intent. No global search, durable drafts or unified store is implied. |
| P2 — bounded population continuation delivered | Per-signal scope disclosure, staged selection versus supplied-current comparison and explicit reread of starting questions. | [Contract and journey 116](POPULATION_WORKSPACE.md#contextual-selection-comparison-and-restoration). Preserve full aggregates, parent/companion predicates, source-free staging and exact fixed windows. Bounded enclosing time-window widening is delivered; categorical selection-versus-rest is delivered; whole direct-conjunct removal and raw-span region complements are delivered. Nested branch editing and arbitrary numeric attributes remain outside the bounded D2 acceptance. |
| P1 — bounded contextual handoff delivered | Analyze the applied Activity scope in its last completed window, or the selected observation in its original window; stage before Run, disclose editing and return to the paused parent. | [Contract and bounded acceptance](ACTIVITY_ANALYSIS.md). Explicit blank authoring and broader population/Saved consolidation remain follow-up work; do not merge execution owners. |
| 0 — closed | Native retained catalog, exact review citations and browser-local native retry recovery. | Delivered evidence is recorded in [VALIDATION.md](../VALIDATION.md). Preserve native-only recovery, distinct store/search clocks and pinned historical reviews. |
| A1 — delivered applied-result portion; blank entry superseded by S1/S2 | Compact Explore: extract the existing distribution/facet body from enclosing workflow chrome; compose one applied scope/result/action region; disclose mobile facets; remove redundant hidden controls and their CSS overrides after parity. | Real desktop/mobile useful visual and first-record positions; stream-first entry follows S1/S2 below; keyboard/touch, staged/applied scope, explicit partial/unknown state, no reads on local disclosure, delayed response rejection. Refresh affected storyboard frames and bounded Bombadil actions. |
| A2 — delivered | Clean live/saved ownership: leave the live operation editor in Operations; use a read-only operation receipt presenter in Saved with explicit continuation. Extract pure versioned route/link/step helpers where this change touches them. | Old hashes, IDs and receipt reopen still work; navigation/reopen does not execute sources; dirty drafts and retained-parent focus survive return; live and saved owners are not reparented. Go/JS fixtures protect exact bounds, missing/empty and optional defaults where protocol helpers move. |
| A3 — bounded service-record/receipt cut delivered | Apply the compact scope/result composition to one service surface; add a record/capture transformation beside existing `GooEvidenceScope` adapters and adopt it in documents/captures. Consolidate only the CSS owned by these compositions. | Two real adopted surfaces share display facts rather than a universal result object; changed-read and coverage caveats stay visible; clocks remain separate; standalone/nested mobile and SSR/pre-definition SPA component checks remain source-free. |
| B1 — bounded consolidation delivered | Compact revision-aware library rows, initially record and native measurement definitions; then documents/retained catalog where compatible. Mount rich meaning/review detail only for selected entries. Split touched retention code into client/basis, catalog, citation and recovery owners without changing intents. | Same-title unrelated entries, explicit families, parallel heads, optional store unavailable, pagination/local search bounds, stale responses and exact old pinned review remain correct; selecting is not execution. Preserve retry body/key and await-token propagation. Mechanical formatting is a separate change. |
| B2 — bounded consolidation delivered | Shared pure browser protocol fixtures and qualified-field/presence display helpers; adopt applied chips and definition summaries. Extract a tiny request-owner helper only after comparing two adapters with matching begin/invalidate/current/dispose mechanics. | Golden Go/JS conformance preserves family differences and stored identities; Hegel state/swarm plus observed Quint traces reject obsolete completion and hidden publication; parent restoration and draft guards survive. Do not add a global mutable store. |
| C1 — bounded consolidation delivered | Freeze identity/response fixtures, then extract immutable-file publication mechanics only. Each store retains validation, digest/canonicalization, lineage and catalog rules. | Byte-identical IDs, large integers, duplicate publication, reopen, corrupt/truncated rejection, parent mismatch and no overwrite. Do not claim power-loss durability without fault evidence. Delete old mechanics after parity. |
| C2 — startup wiring, pure planning JSON and audited transport/JSON and bounded OpenAPI acceptance delivered; remaining coverage queued | Inventory endpoint effects/policies and inject explicit startup dependencies; consolidate strict JSON/response helpers with per-route limits. Then extract bounded ClickHouse HTTP transport, one source family at a time. | Preserve method/status/error/body-limit compatibility, typed escaping, cancellation, non-2xx/oversize handling, family basis and limits. GET/comparison/capture execution is accounted for; SSE retains independent lifetime/no-store/disconnect semantics. Embedded and real-source journeys pass. No Python proxy removal in this cut. |
| D1 — bounded milestone accepted | Unify two-sided comparison presentation, complete fitting inspector adoption and share count-series geometry where two real consumers agree. Add meaningful error/latency ranking and log-appropriate columns. | Independently named windows/bases; incompatible units/kinds remain refused; exact table fallback, common domains, large values and known-to-unknown updates; explicit source continuation and keyboard return. Stored/reread/live execution owners remain separate. |
| D2 — bounded milestone accepted | Bounded population operations: supported column chooser/order, time-duration rectangle with keyboard equivalent, explicit scope widening and exact selection-versus-rest complement. | Typed Go operations preserve parent predicate, range/presence/overlap semantics and full totals independent of paging. Real journey captures/reopens selected evidence; Hegel properties and Quint rejecting controls cover changed execution/selection protocols. No generic predicate-algebra rewrite. |
| E1 — bounded milestone accepted | Native instrument discovery within existing entry points; explicit capability planning for one pinned HTTP-server convention plus generic fallback, then one database-pool or messaging case. Add typed relations and coverage only with an actual supporting source. | Ordinary non-Go, manifest-free and custom OTel data work; missing/empty identities, units, temporality, batch and unsupported cases are honest; static declaration/observed data/verified support remain distinct. No inferred trace completeness or fleet percentiles. |
| F1 — bounded tail publication/recovery delivered | Broaden recovery or exact retained targets only for one concrete supported write/fragment after its server idempotency and source-membership contract exists. Inject exact-reference resolution/clocks without making file document loading depend on XTDB. | Lost acknowledgements/reload preserve intent, IDs and distinct clocks; unsupported writes do not inherit native-journal promises; current-head search differs from pinned/history reads. Select a concrete review task before adding formal labels/export. |

Current position: bounded S2/A3 is delivered. The [B1 definition-library slices](DEFINITION_LIBRARY.md) adopt compact record/native rows on Definitions, in fixed reusable questions and in the document revision catalog. Compatible retained catalog rows/inspectors and the client/basis, catalog, citation and recovery owner split are also delivered. Bounded B2 is delivered: the [record-field and service namespace presentation/fixture slices](FIELD_PRESENTATION.md) are delivered. [Shared-question/native wire fixtures and comparison default equivalence](QUESTION_PROTOCOL.md) are also delivered. The [acceptance audit](QUESTION_PROTOCOL.md#bounded-b2-acceptance-audit) records why no generic request owner was introduced. [C1 immutable publication](IMMUTABLE_PUBLICATION.md) has begun: package and deployment byte/ID fixtures passed before extraction; reusable-package and workbench publishers now share filesystem mechanics. Exploration and investigation capture fixtures now pin exact values/page response, digest differences and refresh compatibility; their publication mechanics are adopted with distinct duplicate policies. Lifecycle custody and native YAML now have frozen clock/bytes/revision/HTTP contracts and use the explicit file-only sync policy. The [C1 acceptance audit](IMMUTABLE_PUBLICATION.md#custody-and-native-yaml-continuation-fixtures) records bounded consolidation as delivered. [C2 HTTP ownership and transport](HTTP_BOUNDARIES.md) now inject startup ClickHouse/optional XTDB configuration while keeping read basis, source/capture/SSE lifetimes and custody distinct. All six audited Go ClickHouse readers share bounded transport primitives behind family contracts and readonly lab checks. Recipes/dashboard and telemetry/native/fixed exploration JSON share mechanics behind 236 frozen HTTP cases. The user-approved [OpenAPI contract](API.md) now covers 70 operations, reflected wire shapes, effects, limits, SSE events, CI conformance and browsable docs. Further mechanical JSON migration is paused. Executed API coverage now includes saved capture/live freeze/document flows, optional real XTDB retention/correction/pinned reviews and real HTTP proxy contracts. The local query/packages images are rebuilt; readonly API/docs acceptance and selected deployed product journeys are recorded in [API acceptance](API_ACCEPTANCE.md). Remaining endpoint/branch coverage and unresolved proxy-policy differences stay explicit; migrate further JSON mechanics only behind coverage for the touched owner. [Bounded D1 comparison acceptance](D1_ACCEPTANCE.md) is complete. The [E1 acceptance audit](E1_ACCEPTANCE.md) accepts bounded native discovery and HTTP/messaging candidates with explicit evaluation limits. The [D2 acceptance audit](D2_ACCEPTANCE.md) accepts the delivered bounded population operations. The [Activity-selection prerequisite](TAIL_SELECTION_CUSTODY.md) now includes a tested shared membership bank. Its typed manifest, idempotent XTDB publisher and explicit LoadTail resolver are tested. Origin-aware Go catalog/read views and shared correction/pinned-review resolution are tested. Public origin read/catalog and citation presentation are tested. SSE issuance, exact Activity selection/recovery and the [bounded failure acceptance](ACTIVITY_RETENTION_ACCEPTANCE.md) are delivered.

### Stream-first revision of the entry queue

The user's latest direction supersedes the earlier initial-entry “explicit start”
constraint: a true fresh blank Explore entry should begin with live observations,
not a question form. This does not authorize automatic execution of saved views,
shared/deep links, captures or definitions. The peer [product/design](STREAM_FIRST_ENTRY_DESIGN.md)
and [architecture](STREAM_FIRST_ENTRY_ARCHITECTURE.md) reviews define the proposed
interaction and feasible boundaries. S1 and the first S2 interaction cuts are delivered with the limits recorded below; these original
criteria remain the acceptance contract.

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
| S3 — bounded exact Activity retention accepted | Exact retained tail selection only after a server-owned capability proves connection membership, canonical records, per-row read basis/window, expiry and retry identity. Reuse existing retention/citation contracts where their prerequisites genuinely fit. | Current connection-local observation handles are not durable event IDs or a live-page `resultToken`. Do not call the native page-retention API with fabricated membership or substitute a rerun. Test duplicate/late rows, eviction, expiry, changed selector, lost acknowledgement and correction before advertising exact capture. Until then offer explicit reread or existing supported evidence paths. |

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
remembered scope and deployment discovery are not implemented. The first S2 cut adds shared-viewer Include/Exclude, staged chips/Apply and explicit fixed-window staging. Paused-parent return now suspends fixed work and restores selection/scope/scroll without a read. Live/fixed summaries and detail bodies now share passive record presentation, including bounded supplied-parent context. Live, fixed/frozen and held service observations now use the shared record-table shell. Record receipt summaries now share a pure adapter between capture pages and document references; fixed continuation stages before Run. Saved record pages now have a dedicated owner without a hidden query editor; Live/fixed and Saved share record/context rendering while retaining independent state. Shared fixed/saved/context record tables now use one inspector instead of duplicate inline attributes, preserve refinement/selection actions and hide comparison buttons without a usable breakdown or retained baseline. Standalone tails now share cached record details, and native identity views share a lazy inspector/direct selected-point body. Native values/coverage now share presentation across measurement details, service summaries, fixed/live native windows and supported-window previews. Fresh Activity no longer mounts a fixed editor; explicit authoring/continuations mount it lazily and retain drafts on paused return. Other aggregate/native service layouts remain in D1/E1; see the [bounded S2/A3 audit](S2_A3_ACCEPTANCE.md).
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

Bounded S2/A3, B1 and S3/F1 are delivered; S3/F1 acceptance is recorded in
[Activity retention acceptance](ACTIVITY_RETENTION_ACCEPTANCE.md).
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

## Reconciliation of previous top-level next items

- General span/log population timelines and saved population views are implemented
  ([population workspace](POPULATION_WORKSPACE.md)); arbitrary native composition
  and whole-dashboard custody remain separate, not another record workspace.
- Log SSE, rolling record/native-counter comparison, general captures, saved
  comparisons, and declared integer attribute measures have delivered subsets.
  Do not treat historical all-or-nothing statements as the active backlog.
- Existing evidence summary, inspector, facets, tables, attributes, waterfall,
  heatmap and chart-frame elements need adoption/coherent ownership, not initial
  reimplementation. Definition rows and result/comparison composition belong to
  A/B/D above, not separate component projects.
- The retained-evidence catalog, exact review citations and native-only browser
  retry journal are complete. Ordinary page-local document drafts, file/native
  cross-store search, historical cutoffs, formal labels and exports remain limits.
- Service namespace identity, declared-operation cards and reported edges exist.
  Instance navigation, typed async relationships, coverage and capability planning
  are demand-led E work; no causality or completeness inference is implied.
- Shared measurement packages and the generic examples remain the primary path;
  Python reservations remain a labelled compatibility lab. Additional kinds,
  expressions, logging-provider production settings, user workflows and warehouse
  preparation are not all prerequisites for the first compact UI round.

## Earlier plan and milestone history

Use the [architecture/process diagrams](ARCHITECTURE_DIAGRAMS.md) to connect build,
production telemetry and investigation stages. Current implemented slices include
[YAML operation packages](OPERATION_PACKAGES.md), live span/subject investigations,
[native counter interval rates](NATIVE_COUNTERS.md) and
[native duration histograms](NATIVE_HISTOGRAMS.md). The diagrams label
production patterns; they do not certify production rollout. Actual verification
and scoped limitations are in [VALIDATION.md](../VALIDATION.md).

Compatible explicit histogram merging is implemented in Go/CLI. Next native-metric
increments are dynamic fleet membership and automatic alignment policy. Fixed
declared membership/coverage is now implemented (FLEET_COVERAGE.md). Operation error
ratios are now implemented with exact aligned outcome intervals. These build on the metric kernels;
they must retain reset/gap/unknown semantics rather than imply completeness.
Historical milestones below retain their original acceptance criteria.

The prototype exists. This plan turns it into a small language with trustworthy measurements. Milestones are ordered; no full UI rewrite or Kubernetes environment is needed for the first two.

Current partial progress: the [Go aggregate pilot](PORTABLE_COMPILER.md) implements
prepared-fact counts/ratios and local reference/ClickHouse parity. The
[lifecycle pilot](LIFECYCLE_PREPARATION.md) now prepares normalized event fixtures
with strict custody bases and correction-link checks. Real browser storyboards,
Hegel state/swarm properties and Quint MBT/trace validation are documented under
docs/journeys and verification. This does not complete real-exporter adaptation,
UI migration, SQL preparation or Databricks integration. PROJECT_DIRECTION.md
supersedes historical language suggestions: Go/Rust runtime and YAML+CEL/KCL-style
authoring are the direction.

The [visualization slice](VISUALIZATIONS.md), requested as the next exploration
step, adds a bounded trace waterfall, raw-duration heatmap and typed cell
selection in the existing lab, plus question-led retry/outcome/USE/RED recipes. It does not complete the Go UI integration or
Milestone 1; those accounting gaps remain explicit.

The [semantic-package foundation](SEMANTIC_PACKAGES.md) now provides Go contract
recognition and capability proposals, with a concrete worker/queue adapter. An initial HTTP binding/expansion now emits shared measurement/view/recipe
artifacts and prepares normalized requests; broader profiles, raw-source
adaptation and descriptor-driven UI remain next steps. This does not complete profile expansion or live source adaptation.

The bounded live HTTP adapter now exercises real exporter spans through the Go
package evaluator and descriptor-driven UI (LIVE_HTTP_PACKAGE.md). Broader
source profiles and automatic instrumentation discovery remain future work.
Custom Orchestrion/quintgo observation is isolated under verification/weaving.

The [operation/event annotation slice](GOO_ANNOTATIONS.md) now emits source
contracts and supports opt-in woven OTel observations. It does not implement
native operation aggregate compilation or automatic package activation.

## Milestone 0 Establish a real stack baseline

- [ ] Run the existing Python and embedded ClickHouse tests in a clean virtual environment.
- [ ] Validate Compose, pull/build images, and start the default stack on a Docker-capable host.
- [ ] Run scripts/smoke.py against actual collector ingestion.
- [ ] Capture actual table definitions and check timestamp units, SpanKind encoding, metric columns, and map attributes.
- [ ] Verify successful, retry, late, missing, and rejected scenarios through the browser.
- [ ] Start optional HyperDX and confirm log/trace source mappings.
- [ ] Record versions, commands, results, and remaining failures in VALIDATION.md.

Completion: default stack smoke passes, actual schemas match the adapters, and UI behavior has been inspected. Optional HyperDX defects are documented if they do not block the core stack.

## Milestone 1 Correct measurement semantics

- [ ] Add deterministic fixture input and expected-result files with a frozen clock.
- [ ] Separate aggregate measurement queries from paginated inspection.
- [ ] Use mature acceptance bounds and outcome lookups through the appropriate cutoff.
- [ ] Detect invalid identity/timestamps and conflicting acceptance/terminal facts.
- [ ] Carry unknown and no-data states without changing their meanings for charts.
- [ ] Add row-limit independence and time-boundary regressions.
- [ ] Map the current completion UI to the corrected result contract.

Completion: page size cannot change the SLI; all hand-written expectations pass; missing work cannot disappear from mature accounting; no-data/unknown remain distinguishable. No claim of production evidence completeness is required or implied.

## Milestone 2 Implement the first composable grammar

- [ ] Specify minimal normalized AST and result types following RFC 0001.
- [ ] Add pure validation, name resolution, and type checks.
- [ ] Implement source, filter, derive as needed, group, count, and compatible ratio in a reference evaluator.
- [ ] Implement explicit temporal lifecycle classification or its documented lowering.
- [ ] Compile the same expressions to parameterized ClickHouse SQL through source profiles.
- [ ] Add parity tests against fixtures and actual backend checks.
- [ ] Replace mode branches incrementally using a compatibility layer.
- [ ] Add a new named declarative measure without new compiler branching.

Completion: one package-defined success/completion SLI runs identically in reference and SQL evaluators, and is reused by a view and an evaluator test. Explain exposes the normalized expression and inferred type.

## Milestone 3 Make investigations composable

- [ ] Let UI filters, groups, and range selections edit validated AST nodes.
- [ ] Add named cohorts, comparison, and selection history.
- [ ] Define relations with key scope, multiplicity, and evidence strength.
- [x] Add linked log/span inspection and an initial waterfall without assuming a single trace tree (bounded lab implementation).
- [ ] Save/reload a session referencing a package revision and evaluation context.
- [ ] Keep raw attribute discovery and surface unavailable historical dimensions.

Completion: a user follows a bad cohort across async boundaries, branches an investigation, and reloads it without changing the published objective. Approximate or contextual relationships are visibly identified.

## Milestone 4 Publish operational resources

- [ ] Implement package schema, immutable revision identity, and template locking.
- [ ] Add validate/explain/plan/apply commands with stable resource IDs.
- [ ] Implement recording rules, retained counts, and correction handling.
- [ ] Add a scheduler, multiwindow burn evaluation, persisted state, and evaluation receipts.
- [ ] Test restart recovery, unknown evidence, duplicate notification suppression, and historical corrections.
- [ ] Add notification adapters behind externally configured route references.
- [ ] Export only supported HyperDX resources, with capability reports.

Completion: one application package drives an SLO, alarm, dashboard, and investigation entry point with consistent revisions and audited evaluation state.

## Later decisions

Choose Clojure/EDN integration after the small normalized model is demonstrated. Evaluate Plotje and other renderers against actual result types. Add Terraform/CRD adapters after a stable reconciliation API. Consider replacing HyperDX fully only after workflow coverage is measured. Performance work, authorization hardening, multi-region queries, and production evidence reconciliation need separate acceptance criteria.

## Suggested first change sets

1. Baseline validation and observed integration fixes.
2. Deterministic lifecycle fixtures and independent expected outcomes.
3. Full-population counts, mature window semantics, and invalid-evidence handling.
4. AST normalization and reference evaluator.
5. ClickHouse compiler parity and declarative-measure demonstration.

Each change set should leave the lab runnable and update its tests and documentation. Keep speculative feature additions out of the baseline-fix change set.

## Current operation integration increment

Implemented: woven Go sample → collector → ClickHouse → explicit Go conformance
adapter → operation overview and binding explanation. Journey 08 executes real
success/error/panic paths and captures three frames; see LIVE_OPERATIONS.md.
This does not complete native metric operators, subject semantics or general
package activation. Those remain distinct next increments.

Operation investigation now supports duration heatmaps and matching-span/trace
drilldown over one captured evidence response. Journey 09 covers a real failed
checkout and nested stock-check waterfall; see OPERATION_DRILLDOWN.md. Native
metric operators, full cross-service trace fetching and subject navigation remain
separate increments.

Scoped subject event navigation is now implemented over captured operation
evidence. Journey 10 proves related-trace navigation without cross-tenant ID
merging. General relation packages, durable event identity, lifecycle integration
and retry policy remain future work; see SUBJECT_NAVIGATION.md.

Local YAML operation packages now replace sample-specific Go bindings and enable
shared views for checkout and a separate woven report app. Journeys 11/12 cover
real second-app onboarding and stale-response isolation. See OPERATION_PACKAGES.md
for supported syntax and remaining resolver/compiler/deployment boundaries.

Native counter slice: cumulative/delta evaluator, strict operation-rate contracts,
actual OTel metric export, bounded sum adapter and rate/context UI are implemented.
Journey 13 covers live cumulative intervals; delta is fixture-tested. Automatic fleet operators and warehouse metric lowering remain next
increments. Architecture/process diagrams are in ARCHITECTURE_DIAGRAMS.md.

Native histogram slice: exact cumulative/delta bucket evaluation, reset/gap/layout unknowns, explicit aligned merge API/CLI and per-stream heatmap are implemented. Journey 14 documents live cumulative data. Automatic fleet discovery, time rebucketing and warehouse histogram lowering remain future work; see NATIVE_HISTOGRAMS.md.

Operation error ratios now support exhaustive denominator contracts, explicit SDK zero baselines, exact same-basis ratios, unknown/zero-traffic distinctions and guided contextual trace investigation. Journey 15 and OPERATION_ERROR_RATIOS.md describe the live path. Fleet aggregation, burn rates, alerts and change detection remain future work.

Declared fleet slice: 2–4 expected instances, same-bound-source queries, peer-failure isolation, exact aligned population sums and coverage explanations are implemented. Journey16 now uses two observed replicas. Bounded comparisons are implemented separately; exact multi-replica common-window alignment and deployment discovery remain future work.

Live two-instance slice: operation-peer has distinct identity and loopback8092. An explicit bounded-skew policy compares actual valid windows without forming approximate fleet totals. Journey17 covers imbalance, export loss, gap and restart; dynamic membership and common-window fleet aggregation remain future work. See FLEET_COMPARISONS.md.

## Local saved investigation milestone

Implemented: immutable bounded server captures, package-revision/fixed-window refresh, stored trace/filter/subject and ratio/member explanation navigation, separate component bases, local named-volume persistence and revision comparison. See [contract](SAVED_INVESTIGATIONS.md). General exploration, cross-service trace/log expansion and one-field cohort comparison are now implemented in later increments below. Persisted general workspaces and production collaboration remain future; see [gaps](PRODUCT_GAPS.md).

## Bounded general exploration milestone

Implemented: shared UI/API/CLI Go query AST and ClickHouse adapter, typed literal predicates and AND/OR, discovered attribute keys, independent full-record counts/groups, bounded row pages, detected cross-read disagreement, cross-service trace expansion and exact trace/span log scopes. See [design and remaining slices](EXPLORATION_DESIGN.md). One-field distributions/comparison and workspace navigation are implemented below. Next proposed: stable capture/cursor paging and persisted general workspaces; production access control remains a prerequisite for team deployment.

## Cohesive workspace and distribution comparison

Implemented: primary Explore/Operations/Saved investigations/Definitions navigation,
secondary Examples, shared visual tokens and responsive trace inspector. Explore
now has Records/Distribution/Compare tabs, complete captured groups, a browser-held
fixed-window baseline and exact rational Go comparison. Empty/changed populations
remain no_data/unknown. See [contract](PRODUCT_WORKSPACE.md). Next: persisted general
questions and baseline receipts, stable paging and reusable boards; production
authorization remains separate.

General-purpose direction: use service-neutral signal/source onboarding and
capability-driven OTel recipes across independent applications as acceptance criteria.
Keep sample identities and business assumptions in example packages/adapters;
shared views consume declared fields, units, grain and relation contracts.

## Manifest-free ad hoc telemetry packages

Implemented: Explore question → named strict YAML definition → shared Go query plan
→ separate live preview → immutable local definition revision → fresh Explore reuse.
Import/export and optional scanner-validated Go annotation templates coexist with
in-place use of ordinary non-Go telemetry. This is captured-record-count only;
configuration revisions do not archive evidence or infer outcomes. See
[contract and next typed-measure/promotion work](TELEMETRY_PACKAGES.md).

## Coordinated exploratory UI testing

Implemented: bounded Bombadil workspace exploration, a real report-app/ClickHouse/goo
coordinator with rendered-evidence checks and retained ledger, and opt-in quint-go
observed-trace validation with rejecting protocol controls. See
[testing design](BOMBADIL_TESTING.md) and [journey 26](journeys/README.md).
The next opt-in campaign now uses Quint-generated issue/inspect/revisit checkpoints,
application-side Bombadil submissions, simultaneous goo navigation and live ITF
poststate replay. It covers three successful reports and two evidence revisits.
The next fault campaign now exercises existing Python retry, rejection and stalled
paths, checks rendered spans/logs independently, validates each observation with
Quint and property-tests the test interpretation with Hegel state/swarm. Next:
resource/export-loss/late-capture generation and longer-lived independent actors.
The models do not establish browser scheduling or telemetry completeness.

## Cohesive core investigation polish

Implemented: shared scoped/typed attribute inspection across Explore and waterfalls;
local key/value search, exact copy and secondary raw evidence; removable filter
chips, grouped field discovery, valid operator controls and denominator-preserving
facet search; selected-span state, scope disclosure, focus return and mobile rows.
The [workspace contract](PRODUCT_WORKSPACE.md#record-inspection-and-query-composition),
[architecture flow](ARCHITECTURE_DIAGRAMS.md#16-structured-investigation-and-local-refinement)
and [journey 29](journeys/README.md#29-structured-inspection-and-filter-composition)
describe the implementation and tests. Cross-signal linked records do not silently
change the active query. Future work may explicitly open a separate log question;
automatic signal switching and a complete accessibility audit are not implemented.

## Discovery over unfamiliar OTel data

Implemented: literal message/service/severity search and readable log rows;
independently scoped surrounding logs with bounded paging and stale-read guards;
reported-trace discovery from full matching span groups; Go-defined fixed duration
buckets; available-field suggestions leading to reusable package preview/save/use.
Journey 30 exercises the real collector/ClickHouse path, including a source
population larger than the record page. Hegel context state/swarm and bucket
properties complement live-ledger Quint checks and rejecting negative controls.
Observed service/span-name inventory now provides single-read full-population
counts, Error statuses and duration buckets with fixed-window Explore/package
navigation. See [contract](SERVICE_INVENTORY.md) and journey 31.
Next: explicit trace-level/root measures,
personal questions and investigation documents. General captures and stable stored
paging are implemented below. Live tail, statistical
inference and production authorization remain separate work.

## Durable general investigation evidence

Implemented: named span/log/inventory captures, server-only evidence custody,
complete bounded aggregate preservation, 1,000-record inspection retention, receipt-
bound stable paging, source-free reopen, fixed-question/plan successor refresh,
local catalog and copyable receipt links. Shared read-only components distinguish
saved evidence from fresh trace/log context and explicit live-question branches.
[Contract](GENERAL_CAPTURES.md), architecture flow 19 and journey 32 describe the
limits and verification. Ordinary live offsets remain mutable; comparison baseline
pairs and expanded context are not automatically archived.

Reusable questions now use saved telemetry-package revisions (see the core milestone below). Next: an accessible ordered investigation document
that pins exact captures, hypotheses and external links. Then reviewed package
promotion, authorized versioned sharing/conflict handling, annotations/history and
optional presence or spatial canvas. Follow the evidence/definition/document
separation in [the vendor research](COLLABORATIVE_INVESTIGATIONS_RESEARCH.md).
Do not equate a query link, live graph, document revision or cursor with a source
snapshot, nor add a separate expression language for collaboration.

## Core workbench before collaboration

Implemented: opt-in full-population exact span-duration summaries, bounded fixed
record-count time charts (including per-interval durations), archived measures,
versioned reusable questions through telemetry packages, observed span/log service
activity/event age and reported parent-service links. General live query admission
limits and cancellation are implemented locally; these are not production auth or
distributed quotas. See [contract](CORE_WORKBENCH.md), architecture flow 20 and
journey 33. Next: broader typed aggregates/native metrics, actual collector
ingestion-health inputs and explicit identity/authorization design. Ordered cards
and collaboration remain their own milestones.

## Grouped measures and core polish

Implemented: complete per-group duration summaries, adapter normalization, exact
rational ranking and stable local-search scales; measure controls, responsive
summary/table layout and duration package previews. General captures retain these
groups, while older count-only groups remain readable. Source disagreement and
fixed-plan upgrade boundaries remain explicit. See [contract](GROUPED_MEASURES.md),
flow 21 and journey 34. Next: broader typed numeric measures, grouped time-series
comparison and native metric exploration before collaborative documents.


## Error and latency investigation

Implemented: opt-in span-status category accounting for totals/groups/intervals;
recorded-error percentage ranking; exact mean/p95/p99/error changes across fixed
windows; prior-window baseline action; changed-group record/trace/log drilldown;
server-reread immutable comparison pairs linked to ordinary record captures.
See [contract](ERROR_COMPARISONS.md), journey 35 and architecture flow 22.
Acceptance uses full-population denominators, missing/unknown preservation,
source-free pair reopening, real browser/mobile evidence, Hegel state/swarm,
Quint ledger rejection controls and saved-only Bombadil measure exploration.
Deployment markers and ordered notes are implemented in the milestone below.
Broader typed measures/native metrics remain separate. Statistical/causal attribution remains unimplemented.


## Deployment context and ordered investigation notes

Implemented locally: immutable authored deployment receipts with separate event
and custody times; explicit fresh context alongside fixed-time charts; ordered
plain-text narrative cards referencing exact capture/marker IDs; edit/reorder,
immutable revision forests, parent/deep-link navigation and retained dirty drafts.
Concurrent saves preserve branches; no mutable latest head or last-write-wins
update is introduced. [Contract](WORKBENCH_CONTEXT.md), flow 23 and journey 36
separate narrative, deployment context and captured telemetry evidence.
Scoped recipes are implemented below. Explicit capture-successor comparison in
narrative workflows remains next. Authenticated shared editing, CI provenance, marker correction
policies and retention/indexing remain separate work.


## Scoped recipes and sample workflow migration

Implemented: three source-free Go recipe transforms inside Reusable questions,
explicit preview/run and hypothesis-only notes handoff. Six ordinary example
telemetry packages bring reservation, checkout and report spans/logs into the
shared Explore/package/capture/notes flow. The original specialized lab is an
explicit disclosure rather than the default example experience.
[Contract](INVESTIGATION_RECIPES.md), architecture flow 24 and journey 37.
Next: explicit capture-successor comparison within ordered narratives; richer
capability-backed recipes and migration of specialized lifecycle contracts remain
separate.


### Implemented: observed metric package authoring

Bounded service/environment discovery over ordinary OTel gauge/sum/explicit
histogram observations proposes editable measurement definitions. Optional empty
scope version/schema are preserved, and stream summaries expose dimensions.
Storyboard 39 covers a separate Python SDK producer without a goo manifest.
Future work: dimension-value selection, user-selected discovery horizons, service
picker integration, and explicit freshness recommendations.

### Product coherence: definitions, evidence and explanations

The [product essence](PRODUCT_ESSENCE.md) defines the common loop and feature
selection criteria. This pass connects service scope to explicit metric discovery,
places measurement revisions in Definitions, adds local library search, exposes
canonical YAML and distinct time scopes, and cancels abandoned discovery reads
on navigation/mode changes. Storyboard 40 exercises the connected flow.

### Implemented: question-first investigation handoffs

Investigations now retain a question, current finding, open questions and ordered
exact evidence references. Explore capture destinations append to an active draft
or create a child of an explicitly selected saved revision. Stored receipt-only
metadata names evidence scope and metric definition identity without a telemetry
read. Existing dirty drafts and old receipts remain intact. Storyboard 41 follows
two application services, logs and an independent native counter through handoff
and fresh definition reuse. Production shared editing/authorization remains future.

### Implemented: evidence-led next questions

A bounded Go planner derives follow-up questions from verified stored captures.
Evidence cards preview preserved/replaced scope before explicit record execution,
fixed measurement preview or fixed-window resource discovery. Reported IDs and
service/time context have different labels; absent/invalid IDs do not create links.
An evidence-gap action appends an exact-citation open question to the draft.
Storyboard 42 follows competing explanations through a real delayed operation,
trace/log evidence and resource instrumentation without automatic causal claims.

## Service browsing slice

[Service details](SERVICE_BROWSING.md) connect existing inventory/signal discovery,
observed attribute suggestions and native instrument previews. Fixed-window return
is local. [Metric-only service discovery](METRIC_SERVICE_DISCOVERY.md) now includes
native gauge/sum/explicit-histogram emitters. Durable navigation history and a
unified Library remain future work.

## Unified service catalog

[One catalog](UNIFIED_SERVICE_CATALOG.md) replaces the competing live span-inventory
and all-signals entry points. Operation inspection/capture moved inside service
detail; the read-only frozen inventory viewer remains compatible. No backend
measurement kernel or persisted capture schema changed.

Record-question navigation now supports versioned fixed-scope links and source-free
history restoration; see [contract and boundaries](QUESTION_LINKS.md). Validated native
questions now use investigation steps below; service-detail URL state remains
subsequent work.

Implemented: [mixed investigation steps](INVESTIGATION_STEPS.md), using one typed
scope envelope for record and measurement intent, immutable named document
revisions, native sharing, explicit reruns and separate evidence references.

Implemented: [linked question evidence](LINKED_QUESTION_EVIDENCE.md), explicit
Run and capture, typed scope differences and attachment recovery. Scheduled runs,
automatic refresh and statistical comparisons remain later work.

### Stored capture comparison (implemented)

Linked questions compare two stored IDs with full-population record summaries,
exact per-stream native gauge endpoints, explicit compatibility decisions and
source-free native inspection. See [capture comparison](CAPTURE_COMPARISON.md).
Complete-window rate and bucket-population comparisons are implemented below.
Interval realignment, lifecycle deltas and statistical inference remain future
work.

### Shared comparison evidence (implemented)

[Comparison receipts and cards](SAVED_COMPARISONS.md) freeze ordered IDs, the
verified displayed decision/result and an algorithm version. Findings revise in
immutable documents; shared links reopen the snapshot. Partial publication and
context changes support retained-ID recovery. Public hosting, author identity and
access control remain outside the local workspace model.

### Complete native window comparisons (implemented)

[Counter rates and histogram populations](NATIVE_WINDOW_COMPARISON.md) now use
full validated coverage and exact stream matching, preserving reset/gap/layout
refusals and p95 ranges in version 2 receipts. Historical version 1 snapshots
remain readable. Fleet aggregation, lifecycle deltas and automatic rebinning
remain future work.

Exact native scope transport and
[observed-support discovery](OBSERVED_WINDOW_DISCOVERY.md) are implemented. The
picker previews all-stream coverage and exclusions, requires selection on both
sides and explicitly rereads selected exact sub-scopes. Future work may expand
support policies; no interval realignment or member dropping is implied.

[Reusable comparison questions](REUSABLE_COMPARISON_QUESTIONS.md) now close the
comparison-to-investigation loop: either side stages a typed question with its own
capture linked. Existing revision, fixed rerun and fresh-window flows are reused.

[Fresh follow-up comparison](FRESH_COMPARISON.md) now connects saved questions to
new evidence and findings through existing capture/comparison/revision primitives.
Preview fixes a same-duration scope; explicit baseline selection and one new read
lead to a source-free comparison. Lifecycle comparisons remain unsupported.

## Captured change breakdowns and product direction

[Captured group comparisons](COMPARISON_BREAKDOWNS.md) reuse full stored
aggregates and open focused questions. New dimensions require explicit recapture.
[Whole-product review](WHOLE_PRODUCT_REVIEW.md) guides automatic service
overviews and distributed relationships within Services. The first bounded slice
is implemented; remaining graphical topology and broader dashboard recommendations
are planned. Observed, declared and executed-supported evidence remain distinct.

## Automatic service overviews

[Service dashboards](SERVICE_DASHBOARDS.md) now derive bounded typed widgets
from observed signals, convention keys and native instruments. Matching registry
contracts retain declared revision IDs. Services also offers a local distributed
view of resolved parent relationships and diagnostics, with fixed-scope handoffs
into questions and immutable investigation evidence. Async/declaration topology
and whole-dashboard snapshots remain future work. Declared-only entries are implemented.

### Framework-neutral display components

The current polish round starts a reusable custom-element library, with compact
mean/percentile summaries shared by Services and live/frozen Explore, and one
related-metric comparison matrix. Declarative HTML and SPA properties use the same
renderer. Application query and evidence state remain outside the elements.
See [component contracts](UI_COMPONENTS.md); whole-app table/facet/attribute
extraction should follow the design audit rather than introduce a second grid or
query-builder runtime.

### Live delivery and baseline-versus-live comparisons (planned)

Future log/trace tail screens can use application-owned SSE adapters delivering
HTML patches (Datastar/htmx-style) or typed display snapshots. Passive components
retain transport-neutral contracts. Design requirements include stable keys,
bounded retention, focus/scroll-follow, explicit pause/gap/reconnect state and
separate event/observation/custody clocks. Pinned baseline comparisons must advance
an explicitly scoped current question/window/basis coherently; retained tail rows
cannot stand in for full-population measures. See the
[component audit](UI_COMPONENT_AUDIT.md#eventual-live-updates-transport-neutral-display-contracts)
for compatibility, ordering and freeze/capture boundaries. No SSE endpoint or
live-tail/comparison engine is implemented in the component foundation round.

### Passive table and result-state adoption

Implemented shared native-table viewport and explicit result-state elements in
Explore and Services. Preserve caller-owned row/action DOM for future patches.
Next: shared facets and typed attribute displays; streaming adapters remain planned.

### Facets and typed attribute displays

Implemented shared local facet shell, typed values and passive attribute-list
shell, adopted by Explore and trace/log inspection. Next: explicit live transport
and bounded retention adapter design, keeping fixed population measures separate
from tail rows. Field snapshot updates currently reset disclosure state.

### Live logs and reusable trace/time plots

Implemented bounded recent-log SSE observations, explicit gaps, retention and
inspection pause. Reusable waterfalls/count-time plots are adopted in existing
flows. Next: shared plot frames/detail panels and carefully scoped live measure
comparison against a fixed baseline; tail rows remain outside measure totals.

### Rolling baseline comparison slice (implemented 2026-10-20)

Saved record questions can explicitly refresh full-population counts beside a
frozen baseline, preserve their common axis, expose stale source reads, and freeze
the displayed server snapshot without rereading telemetry. Shared passive chart
frame supports SSR/native content and SPA metadata updates. Journey69 covers
actual OTel data, explicit rescale, stale retention, exact snapshot freeze and
old-generation disposal; standalone journey70 covers the component contract.
The next slice below adds live span duration/status; native live metric charts
and SSE measure refresh remain planned.
See LIVE_COMPARISON.md and VALIDATION.md for scope and execution evidence.

### Live span regression measures (implemented 2026-10-20)

Count, requested span latency and recorded error share now refresh/freeze as one
window and evidence basis. Compact shared latency elements accept a supplied
percentile domain; the app preserves it until explicit rescale. Error share has
a fixed 0–100% axis with status coverage disclosure. Journey 71, the live actor,
Hegel properties and observed Quint replay cover the path. Native live metric
comparisons and SSE measure refresh remain future work.

### Native live counter pilot (implemented 2026-10-20)

Single-stream frozen baselines now compare against the newest observed export
window with the same exact nanosecond width. Coverage/reset/identity checks reuse
the native evaluator; freshness is explicit and source-free freeze retains typed
historical evidence. Rational common-domain bars reuse the passive comparison
renderer. Journey 72, native Bombadil invariants, Hegel boundary properties and
observed Quint replay cover the path. Next: explicit exact-stream selection over
multi-stream inventories; live gauges/histograms remain planned.

### Exact native stream selection (implemented 2026-10-20)

Focus a stream in a broad native result to preview an immutable complete
`source.exactStream` binding. Go validates and compiles it before source reads;
sharing, package revisions, comparisons and frozen evidence retain the binding.
Explore all matching streams removes only the exact pin. Journey73 covers actual
lookalike streams; Hegel focus/clear state sequences and observed Quint replay
verify source identity and custody. Live gauges/histograms and logical fleet
selectors remain planned. Next: improve native metric inventory navigation and
compact coverage diagnostics before broadening live native chart semantics.

### Native inventory navigation and coverage guidance (implemented 2026-10-20)

Local identity search, Needs review filtering, explicit empty results and bounded
pagination operate on full retained native inventories. Shared passive stats,
facet shell and attribute disclosures present typed states and readable reason
codes; no new evaluator or source semantics. Live counter gap guidance uses the
same presenter. Journey74 records zero-read navigation and custody for Quint
replay with rejecting mutations; a native navigation Bombadil actor exercises the
controls. Next: default service-native views from observed instrument inventory,
using the same bindings and explicit coverage limits rather than fleet inference.

### Native service-card integration (implemented 2026-10-20)

Default discovered native widgets now show evaluated latest interval values with
separate whole-window/gauge state. Fixed the counter/histogram `ok` versus
`observed` state mismatch; newest unknown intervals still withhold values.
Shared compact tables expose exact-stream question links; lazy retained results
reuse stream navigation/coverage guidance without reads. Existing two-instrument
and two-concurrent-read budgets remain unchanged. Journey75 covers namespace
isolation, interval semantics, histogram ranges, gauge freshness and exact-stream
capture custody with Bombadil/Quint; no new evaluator or planner policy is added.

### Native instrument selection (implemented 2026-10-20)

The pure planner offers validated definition keys and enforces up to two selected
native instruments, with omitted/default versus explicit empty intent. The service
picker stages metadata locally; Apply explicitly rereads the fixed overview under
the existing two-concurrent-read budget. Same-name library/version/schema variants
remain distinct. Later Apply/scope generations invalidate older results. Selection
is visit-local; saved questions/captures retain definitions independently. Journey76,
Hegel generated state sequences and observed Quint/Bombadil checks cover the flow.

### Retained native timelines (implemented 2026-10-20)

A reusable passive scalar timeline composes with the shared chart frame. Native
results/captures show independent gauge dots, evaluated counter interval rates
and histogram interval counts, with explicit unknown bands and read-free
selection. Selected native histogram buckets retain per-interval layouts and
P95 ranges. Exact tables use the shared table shell; unopened stream charts are
lazy. Existing evaluator, question and custody formats are unchanged. Journey77
uses real OTLP plus observed Quint/Bombadil validation; journey78 covers component
server HTML, detached properties, keyboard selection and mobile composition.

Next: richer native histogram distribution comparison/heatmaps over compatible
retained intervals, keeping changed layouts explicit and reusing the same
question/evidence primitives rather than averaging percentile scalars.

## Retained native bucket heatmaps

Added passive exact-count grids, layout selection over the latest 32 retained intervals, synchronized interval inspection, and compatible captured-window counts alongside normalized share comparisons. Unknown and other-layout cells remain distinct. Next: consider time-bucketed distribution views only with an explicit Go aggregation and coverage contract; interval columns do not imply equally spaced time buckets.

## Explicit histogram time buckets

Implemented optional `histogramBucketMs`, whole-interval Go aggregation with complete per-bucket coverage, checked native bucket sums, preserved P95 ranges and explicit unavailable/zero states. Fixed/raw UI views reuse existing passive components and retain source-free capture replay. Journey81, Hegel state/swarm properties and observed Quint/Bombadil validation cover the contract. Next product priority from the [whole-product sweep](WHOLE_PRODUCT_COMPONENT_SWEEP.md): adopt a passive evidence summary in frozen native captures and linked investigation evidence before expanding more chart types.

## Shared native evidence summaries

Implemented the whole-product sweep’s first component in the full captured measurement, linked question evidence, and native capture-card surfaces. Source/storage clocks, exact bounds, namespace identity, independent reads and unknown metadata remain explicit. Service widgets now adopt the same display primitive with fixed mode, domain-specific source clocks, compact facts and inspectable scope. UTC timestamp formatting preserves exact nanoseconds in tooltips. Stored comparison sides now adopt the summary, with aligned scalar pairs and UTC endpoint/window labels. See COMPARISON_PRESENTATION.md. Selected-item inspectors now share the inline shell; bounded in-tab question visits now preserve context; durable cross-session navigation remains future work.

The shared selected-item shell is now adopted in waterfall spans, span/log record inspection and native timeline/bucket selections. See [selected-item inspectors](SELECTED_ITEM_INSPECTORS.md) for lifecycle, retention, focus and source-action boundaries.

## Focused core-experience refinements

The [product survey](CORE_EXPERIENCE_PRODUCT_SURVEY.md) prioritizes a resumable
investigation trail, compact route/operation diagnosis table and accessible
chart-selection continuation into existing typed questions/comparisons. These
now have bounded implementations described below. Continue reusing service
dashboards, comparisons, inspector and display library; do not introduce a
dashboard builder or another analysis destination. The report records official
sources, access limits, semantic boundaries and executable acceptance journeys.

### Survey follow-through

[Core experience refinements](CORE_EXPERIENCE_REFINEMENTS.md) implement bounded
first slices of all six survey items using existing evaluators and display
components. Durable cross-session resume, exclusive/critical-path trace analysis,
general convention resolution and metric exemplar ingestion remain separate
contracts, not implied features.

Population attribute breakdown milestone: discovered per-signal grouping, preserved
predicate intersections, value filter reset, full-population ranking, shared-scale
value/previous-window comparisons and optional saved baseline configuration.
Acceptance includes real OTLP journey 91, source-free save/reopen, explicit baseline
replay, old-view integrity, no-data comparisons, mobile composition, Hegel state
history and observed Quint traces with rejection controls.

Population live milestone: explicit signal-aware SSE tail, bounded inspection,
pausing/navigation/disconnect lifecycle, captured full-population baseline
comparison and source-free saved live settings. Acceptance: real OTLP journey 92,
legacy tail/population regressions, Hegel state history, observed Quint ledger,
Bombadil start/pause, and updated desktop/mobile frames.

The fresh-window preview now also uses the shared comparison action group: one
Capture/retained-retry command position and Preview options for Close, with selector
focus on entry and opener focus on return. This bounded D1 adoption preserves
separate fixed, stored and live execution owners. The bounded adoption is accepted in the D1 audit.

Stored/saved comparison action adoption is delivered alongside the fresh preview:
shared side commands, contextual saved-receipt retry, and grouped finding/receipt
commands. This closes the bounded action pass. The [D1 acceptance audit](D1_ACCEPTANCE.md)
passed; E1 native instrument discovery is the next milestone.

E1's first [native instrument discovery cut](NATIVE_INSTRUMENT_DISCOVERY.md)
adds a compact supplied-metadata table, local paging and pinned HTTP duration
hints with generic fallback. Existing Go discovery/preview/evaluation owners stay
authoritative. Typed convention capability plans, per-stream dimension coverage
and a database-pool or messaging case remain E1 follow-up work.

E1 now also has [typed Go capability plans and observed stream presence](NATIVE_INSTRUMENT_DISCOVERY.md#typed-server-plan-and-observed-stream-presence)
from the existing discovery read. HTTP key/instance presence is distinct from
interval/value validity; generic bindings and definition IDs are preserved.
Temporal/value verification, a pool/messaging case and supporting typed relations
remain queued. Browser convention classification has been removed.

E1's messaging case now recognizes experimental v1.30.0 processing-duration
histograms with conditional per-stream label presence and generic fallback for
shape mismatches. It shares the existing discovery/preview/Run path. No queue-lag,
message-success or trace relation is inferred. Remaining E1 acceptance work is a
bounded audit of the delivered metadata profiles against actual kernel validation;
add typed relations only when a concrete supporting source is available.


The [bounded E1 acceptance audit](E1_ACCEPTANCE.md) is complete. Generated candidate
bindings now have direct evaluator tests, a real messaging observed/unknown/refused
Preview/Run journey, metadata-versus-evidence properties and existing kernel
state/swarm/Quint checks. Discovery does not promise conformance or support for
uninventoried instrument families. Next: D2 selection-versus-rest for the existing
raw-span time/latency rectangle, with parent predicates/window and missing-value
semantics preserved; no generic Boolean editor or native metric complement.


D2 now also stages a disjoint raw-span time/duration region and its remainder in
the full parent window. It uses a typed Go event-timestamp predicate and existing
population comparison/save/capture owners. Arbitrary numeric attributes and nested
Boolean editing remain outside this bounded cut. Next: reconcile D2 acceptance
against the delivered column, rectangle, widening, broadening and partition flows
before extending the predicate surface further.


The [bounded D2 acceptance audit](D2_ACCEPTANCE.md) maps columns, rectangle gestures,
widening, categorical/region partitions and whole-condition broadening to their
actual source, property and navigation checks. Small audit fixes render timestamp
predicates in exact UTC and restore question-trail opener focus. Arbitrary numeric
attributes and Boolean editing stay outside the accepted cut. Next: establish the
S3/F1 membership/idempotency prerequisites for one bounded Activity-row selection;
do not expose exact retention or browser retry recovery before that contract.


The [S3/F1 prerequisite assessment](TAIL_SELECTION_CUSTODY.md) now supplies the
bounded membership bank, typed manifest, idempotent XTDB publisher and explicit
`LoadTail` resolver. Lost acknowledgements, changed intent, absent expired proof,
outage and concurrency are tested against the new backend protocol. Dedicated
tables isolate old receipt bytes and reader assumptions. Origin-aware Go catalog/
read views and shared review/correction/pinned-document resolution are now tested,
including old unprojected receipts, original retry after withdrawal and frozen
legacy cursor bytes. Explicit public content/catalog routes and compact retained/
citation inspectors are now implemented with real OTLP/ClickHouse/XTDB journey 138,
observed browser replay and Bombadil. SSE-issued references and the bounded Activity selection/publication/retry journal
now connect these backend contracts. Journey 139 covers duplicate occurrences and
real commit/response loss/reload retry; per-row read provenance stays separate
from custody. The [bounded Activity retention acceptance](ACTIVITY_RETENTION_ACCEPTANCE.md)
now covers configuration, expiry, journal refusal, actual storage transport
failure, issuance capacity and proof loss after restart, plus keyboard/mobile
recovery. Further store/distributed-identity and classifier work stays deferred.


The user-approved [reusable Activity view cut](ACTIVITY_VIEWS.md) is delivered:
immutable rolling selector/column configurations, source-free deep-link/reload
staging, explicit Start, shared Saved and environment-filtered Services libraries,
clipboard sharing and draft guards. Fixed view hashes/contracts are unchanged.
This closes the bounded persistent Activity layout/scope omission recorded in D2;
arbitrary transforms and predicate-branch editing remain outside that acceptance.

The user-approved [shared filter editor](FILTER_EDITOR.md) extends D2 with bounded
typed AND/OR branch editing. Activity and fixed population/saved configurations
share the supplied AST editor. Activity applies explicitly; fixed edits open a
staged child before Run. Numeric attribute transforms and replacement of the
older single-question builder remain outside this cut.

The single-question adoption is delivered in the [shared editor](FILTER_EDITOR.md):
flat composer/join controls removed, shortcut predicates editable in the full
tree, supplied field suggestions retained, invalid drafts guarded, fixed bounds
retained through filter/environment edits, and shared links staged before Run.
Broader automatic suggestions from Activity/discovery remain follow-up work.

Bounded [observed field/value discovery](FILTER_EDITOR.md#observed-fields-and-values) is delivered across Activity, single questions and fixed/saved views. Existing reads supply scoped metadata and record samples; there is no new suggestion query. Counted value lookup is delivered below; approximate top-K/cardinality remains future work. The bounded semantic-convention enrichment is now delivered in the OTel hints cut below.

[Explicit counted field values](FILTER_VALUES.md) now extend bounded samples using the existing full-scope grouping path. Activity freezes its completed window; single/fixed owners retain applied scope, invalidate stale reads and stage selection. Groupings above 1,000 are refused rather than truncated. Approximate large-cardinality/top-K and persistent lookup captures remain future work.

The approved [latency histogram selection](LATENCY_SELECTION.md) reuses full raw-span duration breakdowns and the existing query/partition owner. Supplied bars emit range intent; exact bounds and child actions stay app-owned. Fixed UTC windows, parent predicates and explicit reads are preserved. The fixed-view component does not subscribe; the subsequent Activity cut below owns optional aggregate reads. Native metric membership remains outside both cuts.

The approved [Activity latency panel](ACTIVITY_LATENCY.md) brings raw-span bucket selection to the entry stream: opt-in manual/10/30-second reads, serialized scope-bound publication, brush pause/staging, Apply and fixed-window remainder comparison. Panel settings are ephemeral; arbitrary bucket layouts, metric membership and automatic default subscriptions remain outside this cut.

The approved [Activity captured baseline](ACTIVITY_BASELINE.md) cut reuses saved raw-span duration captures, existing pure Go comparisons and passive displays. Catalog/link selection, independent windows/clocks, exact count validation, refresh integration and cancellation are delivered. The later saved-reference cut below persists the exact baseline identity. Arbitrary regrouping of saved pages and native histogram membership remain outside these cuts.

The approved [saved Activity baseline reference](SAVED_ACTIVITY_BASELINE.md) cut preserves an exact capture ID in version-1 configuration, with source-free publication checks/reopen, explicit stream/aggregate actions, immutable old IDs, stale-save guards and unavailable-reference retry. Panel/cadence persistence and cross-store deletion transactions remain outside this cut.

The approved Activity-entry saved-view discovery cut embeds the existing library under Views, filtered by applied environment. Draft isolation, canceled/stale catalog responses and cleared cross-environment inspectors retain source-free staging and explicit Start. Search remains loaded-page only; arbitrary selector relevance and automatic startup remain outside the cut. See [storyboard](ACTIVITY_VIEWS.md#activity-entry-storyboard-159).

The approved [Activity navigation consolidation](ACTIVITY_NAVIGATION.md) groups scope analysis and retained fixed authoring under Investigate, alongside the existing Views menu. The duplicated below-stream entrance is hidden during Activity. Source-free staging/return, explicit execution, opener focus and draft identity remain the acceptance boundaries.

The approved [compact mobile Activity cut](COMPACT_MOBILE_RECORDS.md) adds a shared
passive record-layout opt-in, separates legacy responsive rules and retains exact
UTC timestamps, occurrence actions, columns, source-free paused return and desktop
table behavior. Standalone SSR/SPA and nested composition are checked together.

The approved [fixed/saved compact record adoption](COMPACT_MOBILE_RECORDS.md#fixedsaved-adoption-storyboard-162) is delivered through the shared result owner, including explicit linked-log/surrounding context tables. Responsive dates and inline actions preserve full messages, spanning inspectors, selected fields, captured provenance and explicit source reads. Service-held observation tables are now separately adopted through the same presenter, with source-free local inspection and exact fixed-question handoffs verified.

The [record-renderer audit](RECORD_RENDERER_AUDIT_2026_10_24.md) inventories the remaining domain/custody/diagram renderers. Operation UTC/duration and action labels are aligned, while waterfall and retained attributes share the root timestamp adapter. Its operation outcome/status follow-up is delivered below, including compatibility and saved evidence.

The renderer audit’s semantic follow-up is delivered in [independent operation status](OPERATION_STATUS.md): exact optional exporter status, separate declared outcome, truthful legacy absence, immutable explicit refresh, and raw-status-only waterfall error color. Further outcome policy and historical backfill remain outside this cut.

The approved [OTel field discovery cut](OTEL_FIELD_HINTS.md) enriches supplied fields across Activity, fixed editors and cached inspectors using one advisory display catalog. Exact namespace/key identity, independently labeled legacy spellings, explicit reads and source-bound suggestion invalidation are preserved. Full registry/schema migration and inferred query policies remain outside this cut.

The approved [observed-field grouping cut](OBSERVED_GROUPING.md) adds source-free grouping intent inside existing disclosures. Activity analysis, fixed drafts and grouped population children retain predicates, clocks, exact keys and companion scope; explicit Run performs the existing source read. Two-field addition is delivered in the follow-up below; transformed grouping remains outside this action.

The approved [two-field composition cut](COMPOSED_GROUPING.md) adds ordered Add/remove intent and controlled chips across Activity analysis, fixed drafts and population children. Same-scope observations remain ephemeral; unread child edits stay in place; saved reopen retains configuration only. Numeric grouping transforms remain outside the cut.

The approved [grouping consolidation](GROUPING_CONTROL.md) replaces separate fixed/population dropdowns with one passive display element and canonical owner state. Manual, custom-key and observed-field intents stage through the existing owners. Saved contracts and Go execution validation are unchanged.

### Integer distribution and range populations (delivered)

The next approved cut is [exact numeric histograms](NUMERIC_HISTOGRAMS.md), built
on typed attribute predicates and the existing numeric measure binding. H1 adds
same-read powers-of-two bins and checked selected/rest planning. H2 reuses the
numeric form, passive chart and staged population continuation. H3 verifies real
telemetry, saved/reusable configuration, exact boundaries, source-free gestures
and count comparison with property/model and browser exploration checks. Each
step receives independent review before publication. Numeric measure comparisons,
custom bin authoring and inferred OTel types remain separate work.

H1 (`965de9b`) and H2 (`3923f07`) are delivered after independent review. H3
actual telemetry, retained-component regression, model replay and Bombadil
acceptance passed; see VALIDATION.md and the final acceptance commit in git log.

### Numeric measure comparisons (delivered)

Approved follow-up: [numeric comparison contract](NUMERIC_COMPARISONS.md).
NC1 establishes exact arithmetic, strict binding compatibility, complete aggregate
validation and legacy count-only preservation. NC2 extends existing live/stored
comparison owners and passive signed displays. NC3 verifies actual telemetry,
reuse/immutable reopen, source-free interactions and rejection controls with
browser, Hegel and quint-go evidence. Complete and review each gate before
exposing the next; preserve the existing count-only histogram action.

NC1 is delivered in `8c67196`; NC2 live/stored display in `dcb62e4`. NC3 actual OTel,
property/model/Bombadil acceptance are delivered after independent review;
see VALIDATION.md and the final commits in git history.

### Aligned distribution comparisons (delivered)

Approved [distribution comparison contract](DISTRIBUTION_COMPARISONS.md): DC1
establishes exact aligned bins, all-record share denominators, explicit side
planning and legacy intent/version preservation; DC2 reuses passive display/range
components and existing baseline/saved-view owners; DC3 exercises ordinary OTel
shape changes, saved reopen, side drilldowns and stale-response controls through
browser, property/model and Bombadil acceptance. Review and commit each gate.

DC1 `07a6fa9`, DC2 `aa96720` and DC3 actual OTel/property/model/Bombadil and
passive/mobile acceptance are delivered after independent review; see VALIDATION.md.

### Explain a selected population (delivered)

Approved [population explanation contract](POPULATION_EXPLANATIONS.md): PE1 adds
bounded count/group-only comparisons across explicit qualified fields, exact share
ranking and filter plans, plus server-held promotion of the observed result. PE2
uses existing contextual comparison/range and capture surfaces. PE3 verifies full
actual telemetry, partial fields, late arrivals, immutable save/reuse and ownership
through browser/property/model/Bombadil evidence. Keep old capture identities and
numeric/distribution workflows unchanged; do not introduce inferred causal claims.


Population explanation status: **PE1 delivered in `cb4f5af`** after independent
semantic/custody review. **PE2 delivered in `017f32b`** after independent review
and the final compact ranked-table/raw-case/mobile browser check. **PE3 acceptance
is delivered**: final main/ownership browser journeys, native replay with 21+11
rejecting controls, the 60-second/ten-phase Bombadil campaign and legacy 174/176
pass. Final artifact review and push are recorded in VALIDATION.md rather than
inferred from implementation status.

### Activity population explanations (active)

Approved [Activity explanation contract](ACTIVITY_EXPLANATIONS.md): AE1 adds a
compact contextual entrance from a completed Activity window, checked time brushing
and preceding/saved-scope baseline staging using the shared explanation owner.
AE2 validates actual telemetry, immutable observed Save, scope continuity,
source-free return and stale-owner rejection through browser/property/model and
Bombadil checks. Saved baseline scopes are explicitly reread, not presented as
frozen field counts. Complete and independently review both gates before push.
