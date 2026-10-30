# Product capabilities and view-grammar review — 260

Reviewed 2026-10-30 against Goo `7ef1bb2`, before the concurrent260 implementation.
This is a source/documentation review and design proposal. No Goo browser campaign,
authenticated vendor product trial or customer study was run by this reviewer.
No runtime behavior is implemented by this report.

## Product judgment

**Updated user constraints:** Goo is prerelease with no legacy/API/schema floor.
The continuation should use one canonical data/plot/view grammar and replace old
contracts, callers and tests directly where that simplifies the design. No
compatibility shims, duplicate old/new models or fallback old-format readers are
required. Preserve correct populations, exact evidence and useful workflows; those
are semantic obligations, not reasons to retain an obsolete API. Pinning a view or
kind revision records reproducible meaning, not perpetual old-schema support.

The whole UI must be expressible in that canonical grammar: screens and sections,
charts, tables/lists/waterfalls, inspectors, summaries, search/filter/facet controls,
forms and declared actions. Built-ins and user contributions use the same public
contracts; a chart-only configuration DSL or private internal screen API is not
the target. Every existing chart/widget must retain its semantics. A
full/complete-data-only widget must refuse partial or unknown input until a trusted
data-path guarantee satisfies its exact requirement. Other kinds may display
partial or uncertain data, but value knowledge, coverage, sampling, time and
relationship uncertainty must be modeled separately and carried through rendering
and interaction. These are first-cut product gates, not later polish.

Goo has enough observability primitives to deliver a useful investigation product.
Its principal gap for this request is **composing those primitives around the same
scope**, then letting a person or team save that composition. Another independent
chart, dashboard editor or query language would enlarge the gap. A common display-kind
contract, constrained visual editor and reusable scoped views would close it.

Keep the product loop: **observe now → narrow a population → follow supporting
signals → compare/explain → keep → reuse**. Datadog supplies useful operational
entry-point and investigation-document patterns; Honeycomb supplies query-centric
analysis and reusable boards; HyperDX supplies an approachable event-first entry;
Grafana supplies the strongest reference here for separating data, visualization,
field options and reusable panels. Borrow these working patterns while preserving
Goo's checked populations, units, independent clocks and immutable evidence.

## Research basis and limits

Official web documentation hosts for Datadog, Honeycomb, Grafana and ClickHouse
returned managed-proxy403 responses. Public official GitHub repositories were
reachable. Sources below were read at pinned commits, rather than relying on
unversioned search snippets. The Datadog and Grafana documentation directories
have moved; the verified paths are listed below.

Honeycomb's public docs repository could not be located. Its official Terraform
provider documents query specifications, flexible boards, saved board views,
derived columns, SLOs and triggers. Those establish supported configuration/API
concepts; they do **not** establish the current visual editor, BubbleUp gestures,
multiplayer UX, plan entitlements or exact SaaS behavior. BubbleUp is therefore a
product analogy here, not a freshly verified interface claim. No vendor pricing,
retention entitlement, production reliability or edition parity is asserted.

The loaded cloud skills expose runtime/setup workflows, not a design or product
review skill. This review therefore uses the repository's product essence,
architecture and component contracts directly.

### Verified external sources

| ID | Official source, pinned commit | What the source supports |
|---|---|---|
| D1 | [Datadog Trace Explorer](https://github.com/DataDog/documentation/blob/cbea4a625ec1d3a245c97c0bab78d33384e24524/hugo/content/en/tracing/trace_explorer/_index.md) | Live versus indexed search, tags/facets, trace/root/service-entry populations, trace-pattern and anomaly-analysis entrances. |
| D2 | [Datadog dashboards](https://github.com/DataDog/documentation/blob/cbea4a625ec1d3a245c97c0bab78d33384e24524/hugo/content/en/dashboards/_index.md) | Predefined visualization families, dashboard composition and common graph configuration. |
| D3 | [Datadog notebooks](https://github.com/DataDog/documentation/blob/cbea4a625ec1d3a245c97c0bab78d33384e24524/hugo/content/en/notebooks/_index.md) | Investigation/runbook documents with graphs, templates, shared time controls, variables, versions, comments and documented multiplayer editing. |
| H1 | [Honeycomb query specification](https://github.com/honeycombio/terraform-provider-honeycombio/blob/75b6d4807b181491d3178c93ccd93ac0f0b963b3/docs/data-sources/query_specification.md) | Calculations, filters, breakdowns and calculated fields as a reusable query specification. |
| H2 | [Honeycomb flexible board](https://github.com/honeycombio/terraform-provider-honeycombio/blob/75b6d4807b181491d3178c93ccd93ac0f0b963b3/docs/resources/flexible_board.md), [board view](https://github.com/honeycombio/terraform-provider-honeycombio/blob/75b6d4807b181491d3178c93ccd93ac0f0b963b3/docs/resources/board_view.md) | Named board compositions of queries/text/SLO panels and reusable filtered views. |
| H3 | [Honeycomb derived column](https://github.com/honeycombio/terraform-provider-honeycombio/blob/75b6d4807b181491d3178c93ccd93ac0f0b963b3/docs/resources/derived_column.md), [SLO](https://github.com/honeycombio/terraform-provider-honeycombio/blob/75b6d4807b181491d3178c93ccd93ac0f0b963b3/docs/resources/slo.md), [trigger](https://github.com/honeycombio/terraform-provider-honeycombio/blob/75b6d4807b181491d3178c93ccd93ac0f0b963b3/docs/resources/trigger.md) | Reusable calculations and operational follow-through beyond a plotted query. |
| X1 | [HyperDX README](https://github.com/hyperdxio/hyperdx/blob/0212c7bd34896f79546d0a4630df9fc063165dff/README.md) | Unified event search, live tail, high-cardinality dashboards, correlations and alerting as the product's documented goals/features. |
| X2 | [ClickStack search](https://github.com/ClickHouse/clickhouse-docs/blob/89036285e33b88b424ebc397cce3cb70e9c2b5a9/docs/use-cases/observability/clickstack/search.md), [overview](https://github.com/ClickHouse/clickhouse-docs/blob/89036285e33b88b424ebc397cce3cb70e9c2b5a9/docs/use-cases/observability/clickstack/overview.md), [alerts](https://github.com/ClickHouse/clickhouse-docs/blob/89036285e33b88b424ebc397cce3cb70e9c2b5a9/docs/use-cases/observability/clickstack/alerts.md) | Shared event filtering across search/charts, saved searches, time selection, event deltas and alert configuration. |
| G1 | [Grafana Explore](https://github.com/grafana/grafana/blob/6ef9abd5ae799c3ff65c2da1f4777ef62afc4502/docs/sources/visualizations/explore/_index.md), [logs](https://github.com/grafana/grafana/blob/6ef9abd5ae799c3ff65c2da1f4777ef62afc4502/docs/sources/visualizations/explore/logs-integration.md), [traces](https://github.com/grafana/grafana/blob/6ef9abd5ae799c3ff65c2da1f4777ef62afc4502/docs/sources/visualizations/explore/trace-integration.md) | Ad hoc exploration before dashboard authoring and signal-specific inspection within that flow. |
| G2 | [Grafana panels](https://github.com/grafana/grafana/blob/6ef9abd5ae799c3ff65c2da1f4777ef62afc4502/docs/sources/visualizations/panels-visualizations/_index.md), [transformations](https://github.com/grafana/grafana/blob/6ef9abd5ae799c3ff65c2da1f4777ef62afc4502/docs/sources/visualizations/panels-visualizations/query-transform-data/transform-data/index.md) | Query/result versus visualization, field formatting, units and a distinct result-transform layer. |
| G3 | [Grafana library panels](https://github.com/grafana/grafana/blob/6ef9abd5ae799c3ff65c2da1f4777ef62afc4502/docs/sources/visualizations/dashboards/build-dashboards/manage-library-panels/index.md), [variables](https://github.com/grafana/grafana/blob/6ef9abd5ae799c3ff65c2da1f4777ef62afc4502/docs/sources/visualizations/dashboards/variables/_index.md), [correlations](https://github.com/grafana/grafana/blob/6ef9abd5ae799c3ff65c2da1f4777ef62afc4502/docs/sources/administration/correlations/_index.md) | Reused panels, contextual parameters and declared cross-source links. |

These are documentation/source references, not proof that all offerings expose
all features in the same edition or deployment mode. Grafana's frontend patterns
are reviewed here; this is not a parity claim for operating all Loki, Grafana,
Tempo, Mimir and profiling backends.

## What Goo already has

The [current validation](../../VALIDATION.md) and source owners support the following
assessment. Historical paragraphs in `KNOWN_ISSUES.md` retain obsolete absence
claims; use its current opening section and current implementation when reconciling.

| User task | Goo state at259 | Practical gap or next standardization |
|---|---|---|
| See current logs/spans and isolate a population | **Implemented, bounded:** stream-first Activity, shared filters, facets, columns, readable UTC, explicit Start/Pause and signal continuity. [README](../../README.md), [scope continuity](../../journeys/239-activity-signal-scope-storyboard.md). | Preserve this entry. A view selector should be contextual, not another introductory configuration screen. |
| Inspect a record, trace or supporting signal | **Implemented, bounded:** attribute inspector, reported parent context, waterfalls, linked/surrounding logs, trace witnesses and observed-page breakdowns. [Trace search](../TRACE_SEARCH.md), [trace timing](../TRACE_PAGE_TIMING.md). | Consolidate return/focus behavior. Page support, trace completeness and causal claims remain explicit. |
| Compare and explain differences | **Implemented, bounded:** numeric/distribution comparisons, log patterns, field/value share explanations, explicit baseline/selection scopes. [Population explanations](../POPULATION_EXPLANATIONS.md), [comparisons](../NUMERIC_COMPARISONS.md). | Make side identity, denominators, unsupported values and selection gestures consistent across kinds. More chart types alone will not solve this. |
| Discover ordinary OTel services/resources | **Implemented, partial coverage:** no manifest required; HTTP, legacy RPC, database-client and messaging roles plus bounded host/Kubernetes/native observations. [Service dashboards](../SERVICE_DASHBOARDS.md), [automatic dashboards](../AUTOMATIC_DASHBOARDS.md). | Suggest applicable views from supported signals; expose missing prerequisites. Do not imply complete topology, fleet health or universal semantic-convention coverage. |
| Save exactly what was observed and find it later | **Implemented, bounded:** immutable captures plus Activity/fixed/Saved record Keep, source-free exact Open and cross-browser metadata discovery. [Keep contract](../KEPT_OBSERVATIONS.md). | Unify concise evidence chrome and later optional meaning. A selection is not a complete population, classification or training label. |
| Save a reusable question/definition | **Implemented, partial scope adoption:** versioned Activity views, telemetry packages and native packages. [Question protocol](../QUESTION_PROTOCOL.md), [telemetry packages](../TELEMETRY_PACKAGES.md), [measurement packages](../MEASUREMENT_PACKAGES.md). | Checked environment selectors are not universally supported by reusable views/packages/native/service families. Visual authoring must preserve the checked scope or make its parameterization explicit. |
| Build an investigation/runbook | **Implemented, bounded:** immutable investigation revisions, question/evidence references, ordered notes and local draft recovery. [Draft recovery](../INVESTIGATION_DRAFT_RECOVERY.md), [annotations](../WORKBENCH_ANNOTATIONS.md). | A view-aware composition editor is missing. Shared editing, presence, conflict handling and permissioned team ownership are not implemented. |
| Represent certainty, uncertainty and complete-only eligibility | **Partial specialized behavior:** source limits, gaps, invalid values, exact units, independent clocks and ambiguous trace relationships already exist in individual families. | The canonical model must express every current widget's semantics, including distinct value knowledge and coverage facets. Complete-only requirements need basis-bound data-path attestation and hard refusal, not a warning displayed after rendering. |
| Change presentation without changing the question | **Partial:** specific record/distribution tabs and specialized display components exist. A passive [chart frame](../UI_CHART_FRAME.md) already supplies shared chrome. | No common application-wide input/compatibility/selection contract for a registry of display kinds. Existing package `views` is a small fixed allowlist, not a contributor renderer system. See [package schema](../../exploration/package.go). |
| Author a team view visually | **Missing general capability:** the existing [telemetry package UI](../../ui/telemetry-packages.mjs) builds/validates/previews/saves YAML; service defaults and investigation documents supply partial composition paths. | This is a foundation, not a general WYSIWYG view-template editor, reusable panel library or team extension API. Avoid calling it one. |
| Alert or operate an SLO from the same definition | **Partial primitives; operational loop missing:** typed measures and lifecycle semantics exist. [Measurement limits](../MEASUREMENT_PACKAGES.md#remaining-work), [known limits](../KNOWN_ISSUES.md). | No durable general scheduler, alarm lifecycle, notification routing or complete multiwindow SLO system. A plotted threshold is not an alert service. |
| Work safely as multiple production teams | **Missing production boundary:** local prototype with bounded stores, no production authentication/authorization or shared-draft concurrency. [Architecture](../ARCHITECTURE.md), [known limits](../KNOWN_ISSUES.md). | Required before claiming team-hosted production readiness. Namespace labels alone are not access control. |
| Browser/session replay, profiles, broad infrastructure/cloud suite | **Outside demonstrated core:** bounded runtime/resource signals are not an end-to-end RUM, replay, profiling or cloud-management suite. | Keep integrations possible, but do not put a broad suite ahead of the scoped view-grammar loop. |

## Useful patterns to borrow

| Reference | Borrow for Goo | Avoid importing wholesale |
|---|---|---|
| Datadog D1–D3 | Service-oriented defaults, familiar facets, an investigation document that can mix text and scoped graphs, templates and clearly scoped time controls. | A new top-level destination for each technical family, a widget catalogue as the product entrance, or feature-count parity as the roadmap. |
| Honeycomb H1–H3 | A reusable query specification distinct from its display, heatmap/distribution-led exploration, named filtered views of a board, and reusable calculated meaning. | Treating every rendered row as the measured population; inventing current BubbleUp/multiplayer specifics without verifying the live docs. Do not adopt its expression syntax by imitation. |
| HyperDX/ClickStack X1–X2 | One event-first entrance, approachable field/value narrowing, consistent filtering across search and charts, quick reuse of a search. | Making Lucene, free SQL and Goo's typed predicates three unrelated user-facing truths. The same canonical typed query should back any future textual convenience syntax. |
| Grafana G1–G3 | Data/visualization separation, a small field-options contract, typed context parameters, reusable panels, declared correlations and progressive authoring. | An unbounded transform/plugin toolbox that silently changes grain, exactness or source scope; a second query/runtime owner inside every panel. |

Goo's differentiator should be that **meaning survives a change of surface**:
ordinary OTel exploration, code-managed definitions, displays, evidence, notes and
future warehouse translation share the same explicit population and units. Avoid
claims that one preserved observation is a trained label or that related time
windows prove causation. Those boundaries are strengths when expressed once and
progressively, rather than repeated as paragraphs before every result.

## Priority proposal

### 1. Finish the shared investigation contract

Ship the already planned trace-parent/keyboard fix and shared compact evidence
chrome. Standardize empty/loading/stale/partial/unavailable states, time/scope
labels, exact values/units, selection actions and Close/Back behavior. Keep rows and
charts above explanatory prose. Put uncommon provenance in one consistent details
location while leaving source effect and unsupported state visible at the action.

A kind switch must preserve the applied scope, original window, selection identity
and draft ownership. A local rendering change should perform zero source reads.
When a new kind needs another projection, show **Run** in the existing source
owner instead of making a component fetch implicitly.

### 2. Make whole-UI parity, quality and eligibility the P1 grammar gate

A `ViewDefinition` should describe an entire screen, a reusable section or one
widget using the same public component/control registry. The shared context needs
typed scope sessions (draft/applied and parent/history), dataset roles, local
selection, view parameters, documents/drafts and action outcomes. It need not force
all values into a single numeric dataframe. Charts, tables, lists, waterfalls,
inspectors and controls consume those declared roles; source execution and
privileged effects remain registered host capabilities.

Expose both local and scope-bound search/filter/facet/table variants. Local sort or
loaded-row filtering is not source filtering. A scope-bound control emits a typed
operation to its named scope owner, preserving unrelated clauses and checked field
identity; it does not privately mutate a population or construct an endpoint.
Staged versus immediate/debounced user-commit behavior is declared by that scope's
execution policy, not hidden in a control lifecycle. Several widgets can bind to
one scope/selection and share the same declared data demand instead of launching
duplicate subscriptions.

Facet counts must declare what they count against: the applied scope, scope
excluding that facet's own clause, an explicit baseline or supplied inspection
rows. Preserve support, omission and completeness. Own-clause exclusion needs an
unambiguous control/clause binding when several constraints target the same field;
matching a field name alone must not remove another constraint. Missing and empty
values remain distinct. New facet/value reads are explicit data demands handled
by the shared source owner. Complete-only and uncertainty requirements apply to
controls as well as charts.

Before replacing current display paths, inventory **every existing chart/widget**
and map it to the canonical dataset shape, semantic roles, quality facets, view
composition, renderer and typed interactions. Cover records/attributes/facets,
stats/quantiles, ranked groups, time/grouped/value series, raw/native distributions
and heatmaps, comparisons/explanations, state intervals, trace waterfall/tree/page
breakdowns and service/host/Kubernetes/native resource widgets. The
[chart-platform report](CHART_PLATFORM_DESIGN_260.md) carries the detailed mapping.
A line/bar/table pilot is an implementation slice; it does not discharge this
whole-product expressiveness gate. Replace existing implementations and tests
instead of preserving their API shapes as a compatibility requirement.

Specialized renderers remain useful when their interactions are better than a
generic library. Their semantics must still be fully declared in the same grammar;
an opaque native-renderer branch is not a semantic escape hatch. Business/demo
widgets must become ordinary declarations, not hardcoded exceptions.

Quality is multidimensional: recorded/observed, exactly derived, estimated, bounded,
unknown, invalid, missing and not-applicable values are distinct from result
coverage, collection/sampling, temporal coverage/freshness and relationship basis.
Exact integer encoding does not establish measurement certainty. Unknown is not
zero, observed zero is not no-data, a raw status may be known while its meaning is
unknown, and an exact queried aggregate may still have unknown upstream collection.
Bounds need their method/assumptions; statistical confidence is not interchangeable
with deterministic bounds or rounded plot coordinates. Propagate these facets
through transforms and selections; do not silently interpolate gaps, clamp invalid
values or turn estimated inputs into certain outputs.

**Complete-only widget requirements are hard eligibility rules.** Declare which
aspect must be complete: matching query records, required aggregates/bins/support,
a fixed window under its cutoff/lateness policy, reported graph closure, or an
upstream collection/no-sampling contract. The trusted planner/source path must
attest the required guarantee for the exact scope, dataset revision, plan and basis.
A team template or renderer cannot assert it for itself. A complete aggregate need
not download every raw record; exhaustive captured-query results do not prove
complete real-world telemetry. A complete empty result may be valid while a
percentile remains unavailable.

The kind menu should distinguish available now, requires complete data and
unsupported. A blocked kind may stage an explicit acquisition plan; it does not
silently reread or render partial data with a caveat. Re-evaluate eligibility on
live replacement/append/correction revisions and refuse when required guarantees
are lost. Unknown collection may remain acceptable for a kind requiring complete
query bins, but is ineligible for one requiring complete upstream collection.

A base renderer adapter over Plotly or another chart library can accelerate new
kinds, but the ordinary contract should specify **kind + typed field bindings +
units + context + supported gestures**, not the plotting library's trace/layout
objects. The adapter is an implementation detail. An advanced escape hatch should
be explicit, versioned and unable to grant source access, change authoritative
scope, bypass complete-data eligibility or discard canonical quality facets. No
vendor/native option may introduce opaque semantics outside the grammar. Renderer
choice still needs the parallel architecture/performance/a11y spike.

Suggested gestures: time brush; numeric/category selection; compare selected with
rest/baseline; inspect supporting records; Keep selected; add the applied question
or exact evidence to an investigation. Every gesture emits a typed intent to the
application owner. It does not directly execute arbitrary SQL or publish evidence.

Acceptance: maintain the exhaustive existing-family parity matrix and test every
family's meaningful exact/derived/estimated/bounded/unknown/invalid/no-data states,
coverage/sampling/time/relationship facets and incompatible comparisons. The same
scoped result can switch compatible kinds without reads; selections preserve exact
keys, bounds and quality; keyboard/touch and exact-value alternatives work. Prove
complete-only refusal for partial/unknown input, forged completeness, page/count
coincidence, incomplete windows/graphs and loss of a live revision's attestation.
Show explicit acquisition where possible and refusal where the source cannot
supply the requirement. Both a generic adapter and specialized renderer must use
the same declared semantics; neither may fabricate certainty or completeness.
Whole-screen declarations must additionally cover Activity, service/resource
inspection, trace/detail, Saved discovery/detail, Definitions authoring and an
investigation with evidence/view/text sections. Use the public contracts for all
of them; verify shared scope/selection, local-versus-source filtering, facet count
basis, declared action availability, keyboard paths, stale-response suppression
and one owner per source demand. Navigation chrome may remain host-provided through
an explicit capability, not a hidden internal-only screen interface.

### 3. Make scoped views the first customization object

A saved view should compose a **scope/session reference, typed dataset and
projection/measure bindings, display/control kinds, layout, declared actions and
bounded parameters**. Its revision references existing
query/measurement definitions rather than duplicating their compiler. A preview
uses an explicit result/evidence basis. Saving a view saves reusable intent; Keep
or Capture preserves observations through the existing, separately named actions.

Use one canonical view object for code/YAML contributions and visual authoring.
Replace the current string-only authoring scope representation where necessary;
update its readers, writers, planners and tests together instead of retaining both
representations to satisfy a legacy floor. Team or
service contributions should declare a namespace, supported input shapes, semantic
requirements (including hard complete-data and supported-quality requirements),
defaults and versioned dependencies. Initially, namespace/owner
metadata is organizational context; actual shared production ownership awaits
identity and authorization. Do not quietly convert unreported/effective scopes
into a current-only environment name to make an existing package schema accept it.

Named context parameters should bind typed selectors, dimensions, windows or
supported resources. Avoid string-substituting raw SQL or HTML. Display-only
transforms may project, order, map roles or format supplied compatible values.
Calculations that change values, support or quality belong to the checked Go plan
with declared provenance and facet propagation. Aggregation, units, quantiles,
native reset/temporality and population denominators stay in that semantic layer. A view cannot invent a
fleet percentile from scalar percentiles or infer a request count from visible rows.

### 4. Deliver a constrained visual editor, then team reuse

Start from an actual result or existing service overview: **Customize view** →
choose compatible kind → bind fields → set concise labels/units → arrange a few
blocks → preview → save revision → apply to another compatible scope. Keep a
readable generated YAML/JSON canonical view definition available. Valid supported
properties must round-trip through source and visual editing. Invalid or unsupported
clauses must produce a clear refusal while preserving the current draft; there is
no obligation to keep obsolete schemas readable or maintain a second legacy editor.
Published investigation/embed references should pin their view/kind revision at this
persistence stage. An unsupported pinned revision is reported explicitly rather
than silently adopting a newer definition; any deliberate conversion is an explicit
prerelease data decision, not a permanent compatibility mechanism.

The first editor should cover a small useful composition: a scope-bound search
and facet, service activity, error support, latency distribution and grouped slow
operations, with record/trace inspection and continuation actions. This tests the entire grammar rather than a freeform empty canvas.
Prefer schema-driven controls and a constrained layout template to arbitrary
HTML/CSS/JavaScript. Add richer WYSIWYG layout and user-defined presentation templates
once the canonical data/context contract, renderer boundaries and checked
round-trip behavior are demonstrated. Kindly/Plotje/Portal and oscope research should refine this
proposal; their syntax should not be adopted solely for familiarity.

Acceptance: a team contribution applies to HTTP and messaging contexts where its
requirements fit; absent fields/unsupported units produce explicit incompatibility;
changing its display cannot broaden its population; imported code and editor saves
have the same canonical revision; existing evidence reopens with its pinned view
revision even if the team publishes a later one.

### 5. Reuse investigations before adding a collaboration canvas

Let an investigation embed a scoped view, a reusable question, exact kept evidence
and short explanatory text, with their different lifetimes clear. Named reusable
sections/runbooks are more valuable immediately than multiplayer cursors or a
separate whiteboard. Later shared editing needs explicit draft ownership, revision
conflicts, presence, authorization and recovery; immutable revisions alone are not
multiplayer. Optional annotations/review labels can then attach meaning to exact
observations without making interpretation a prerequisite for Keep.

### Operational follow-through after this loop

Production authentication, workspace/source isolation, permission checks, bounded
retention/deletion and deployment observability are a separate required workstream
before offering the product to multiple teams. After definition/view reuse works,
implement one small durable alert/SLO workflow over the same definition only if
operational monitoring is the next product goal: scheduled evaluation, explicit
no-data/error policy, durable state, notification routing, test and recovery.
Do not market a graph threshold or an unscheduled preview as that workflow.

## Work to defer

Defer a broad widget marketplace, arbitrary renderer JavaScript, unrelated query
languages, a separate dashboard/investigation storage hierarchy, an unlimited
transform DAG, automatic statistical-cause claims, RUM/session replay/profiler
suites, AI-authored production changes and blanket vendor parity. They do not
close the current observe/isolate/keep/reuse gap. Preserve extension points without
funding every possible extension in the first implementation.

## How to evaluate the resulting proposal

Use four real journeys rather than a feature-count score; each includes controls
and navigation through the public grammar, not only chart switching:

1. **Unexpected errors:** start with ordinary unlabelled OTel; isolate an anomalous
   population; explicitly stage/Run any required aggregate/distribution projection,
   then switch compatible records/heatmap/ranked views without changing scope or
   quality guarantees; inspect a trace; Keep exactly observed records; reopen in a
   second browser.
2. **Team-specific meaning:** take the applied scope, choose a contributed kind/view,
   bind a service-specific dimension in the visual editor, save/export its revision,
   reuse it on another compatible service and disclose why an incompatible service
   cannot use it.
3. **Explain a change:** compare baseline/current distributions with explicit support,
   select an interesting region, inspect its records, then add the scoped view and
   exact evidence to an investigation. Reopen after a view-definition change and
   verify pinned evidence/view identities and separate clocks.

4. **Complete-data eligibility:** attempt a complete-only widget or facet action
   on a limited page/aggregate or unknown collection contract, observe its refusal,
   explicitly acquire the required complete result where supported, then verify
   that a later incomplete live revision removes eligibility. Exercise a
   partial-capable kind over the same input and preserve its canonical quality
   facets. Verify that search and own-facet exclusion preserve unrelated clauses
   and never mislabel loaded-row counts as source counts.

Measure time to first meaningful record, number of source reads for local view
switches (target zero), successful keyboard/touch completion, unexplained scope
changes (target zero), successful code/editor round-trips and whether another
person can reuse the saved meaning without reading the implementation. These
outcomes are more relevant than matching a vendor's number of chart types.
