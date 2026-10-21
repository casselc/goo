# Implementation plan

Implemented: general span/log population exploration with paired record pages, checked companion selectors, shared time selection, and immutable saved views carrying layout/columns. [Contract and sequence](POPULATION_WORKSPACE.md). Subsequent slices include arbitrary native-stream composition and broader saved dashboard layouts.

Implemented: [aligned service telemetry with linked time selection](LINKED_TIMELINE.md). Next general-purpose slices: arbitrary population/cross-service timelines and saved multi-widget dashboard composition; these are not implemented by this service-level milestone.

Implemented: [population → trace → finding](TRACE_FINDINGS.md), including bounded Go-selected page examples, explicit two-capture staging, partial-failure recovery, compact service disclosures and a generic OTel storyboard. Next: declared trace coverage/anchor-retention diagnostics and validation of late evidence changes; no automatic completeness or causality inference.

## Current example standardization

[Shared measurement packages](MEASUREMENT_PACKAGES.md) now connect gauges,
reset-aware counters, native histograms and live lifecycle preparation to general
captures, ordered notes and reusable YAML definitions. Twelve defaults and five
validated data-only workflows cover reservations, checkout and reports. Both Go
examples emit actual trace-correlated OTLP logs. The specialized Python views
remain a labelled compatibility lab, not the primary implementation.

Next: production logging SDK/provider configuration, user-authored workflows,
additional metric kinds/expressions, explicit fleet policies and durable source
coverage contracts. Raw-source preparation/lowering to Databricks is still future.
Historical milestones below retain their original scope; they are not a claim
that these remaining integrations are complete.

## Current consolidation priority

The [core primitives review](CORE_PRIMITIVES.md) supersedes feature-by-feature
ordering above: first unify resolved scope and checked question transforms across
recipes/follow-ups, then typed measures and capability-aware planning. Preserve
existing wire formats, immutable receipts and specialized metric/lifecycle kernels.
OTel semantic-convention profiles are a first-class input across all six primitives,
with explicit version/unit/identity bindings and generic fallback for custom data.
The first [scope/transform slice](QUESTION_SCOPE.md) now supplies checked shared
planning metadata for recipes/follow-ups, with adapters for typed capture questions.
The first typed record-measure lowering and bounded split-time-chart slice is now
implemented ([contract](RECORD_MEASURES.md)). General numeric expressions,
convention-profile resolution and backend capability negotiation remain subsequent
work. [Numeric attribute measures](NUMERIC_ATTRIBUTE_PLANNING.md) now provide
declared integer bindings, exact reference/ClickHouse execution, coverage states,
Explore controls, capture validation and optional YAML persistence. Existing
definitions retain their previous representation when numeric measures are off.

## Next interaction consolidation

The [entry-point design spike](ENTRY_POINTS_DESIGN_SPIKE.md) reviews the running
app and proposes one scoped question workbench, consistent question/evidence/
investigation vocabulary and contextual one-gesture drilldowns through a shared
executor. The first [question-workbench slice](QUESTION_WORKBENCH.md) is now implemented. It consolidates
Explore's three competing query actions, exposes draft versus applied scope,
clarifies capture/reuse semantics and preserves incident bounds when Services
opens metric discovery. Broader navigation changes remain proposed. Keep current routes and typed kernels during the change;
catalog/navigation migrations follow only after equivalent workflows exist.

## Service overview delivery and next product slices

The [automatic service dashboard slice](SERVICE_DASHBOARDS.md) implements a pure
Go planner, bounded typed widgets, matching declaration revisions and a local
distributed relationship table. Next priorities from the whole-product review:

1. Namespace-aware identity and registered operation-span cards are implemented;
   see [service identity](SERVICE_IDENTITY.md). Declared-only catalog membership
   now retains explicit presence states and independent registry provenance.
2. Add deliberate instance navigation without changing service-level populations.
3. Exact reported-edge populations are implemented; see [reported edges](REPORTED_EDGES.md).
   Add separately typed async links before a graphical topology view.
4. Unify investigation outlines and broader definition entry points; a frozen
   whole-dashboard artifact needs its own explicit custody contract.

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
