# Semantic properties and model conformance

The optional [Bombadil coordinator](../docs/BOMBADIL_TESTING.md) complements these
semantic checks. `GOO_CROSS_APP_QUINT=1 npm run test:cross-app` from the repository
root validates a real report → source readiness → goo display → exploration ledger
through qobs/quint-go. `go test -run TestQuintCrossAppProtocolRejectsMutations`
checks a labelled protocol fixture and twelve rejecting controls without a running
lab. The live-ledger test explicitly skips unless `GOO_CROSS_APP_LEDGER` is supplied.
`npm run test:generated-campaign` now uses `connect.Generate` to produce a seeded
nine-step schedule and `connect.Driver` to check the actual browser handler effects
against its ITF poststates. It requires the running lab and is opt-in; model/projection
fixtures run with `go test -run TestQuintGenerateBrowserCampaign` without browsers.
The live observation test skips unless `GOO_CAMPAIGN_OBSERVATIONS` is supplied.

`npm run test:fault-campaign` exercises ordinary Python OTel retry, rejection and
stalled-work paths. Every inspection is checked against `model/fault_observation.qnt`.
`go test -run 'Test(QuintFaultEvidenceContract|HegelFaultEvidenceProjectionStateSwarm)'`
runs the protocol/negative fixtures and 50 Hegel state/swarm sequences (20 steps)
without a live lab. The Hegel test uses one persistent Node process to execute the
actual test-oracle JavaScript projection. It checks ordering/duplicate invariance,
withdrawal of certainty on conflicts and absence semantics, not production durability.
The live fault-ledger test skips unless `GOO_FAULT_LEDGER` is supplied.

This separate Go module keeps Hegel's native loader/development dependencies out
of the runtime module. It uses hegel-go **0.9.8** (embedded libhegel 0.43.4), Quint
**0.32.0**, and the checked-out sibling `observability/quintgo` tools. No sibling
source is modified. Go 1.27+ is needed for the sibling tool module.

From the root, install JavaScript tools with `npm ci`, then:

```sh
go test ./...
python scripts/check_portable.py
python scripts/check_lifecycle.py
npm run model:check
npm run model:test
cd verification
go test -v ./...
```

In the cloud workspace, initialize the Go PATH/GOPATH/GOCACHE as documented in
`docs/PORTABLE_COMPILER.md`. Also set `XDG_CACHE_HOME=/workspace/.cache` so Hegel
can materialize its checksum-verified embedded library in a writable directory.
Keep TLS and Go module checksum verification enabled. A browser binary is not
needed for these checks.

## Hegel

- Exact deadline inclusion and duplicate delivery invariance: 300 requested cases.
- Invisible later input cannot change historical visible evidence/facts: 200
  requested cases, including a delivery exactly at the exclusive cutoff.
- The real preparer under state/swarm sequences of accept, terminal, duplicate,
  basis advance and snapshot: 150 requested cases, up to 30 steps each, with
  projection and archived-receipt invariants checked after every enabled step.

Hegel automatically chooses per-case rule subsets for swarm testing. This is a
sequential lifecycle state machine; no unsupported concurrency claim is made.
`PBT_SCALE=10` increases case counts. Shrunk examples persist outside the checkout
under the OS temporary directory's `goo-hegel` path.

## Quint and quint-go

The small [model](model/lifecycle.qnt) covers one logical subject, an on-time
terminal delivered at custody time 95, strict basis bounds 80/95/96/100,
idempotent duplicate deliveries, and up to three retained evaluation snapshots.
It intentionally does not model arbitrary schemas, clock skew, distributed
publication, conflicting payloads or the complete preparation implementation.

`connect.Generate` produces eight seeded MBT traces of twenty steps. A
`connect.Driver` dispatches every action to the real `lifecycle.Prepare` harness
and compares model state, observations and archived snapshot counts after each
step. This is implementation replay, not merely running the model alone.

Separately, `qobs` records a real Go lifecycle sequence. `validate.NewChecker`
checks that observed trace against the model using [binding.yaml](model/binding.yaml).
A deliberately impossible observed good count must fail conformance; it is an
expected rejection and a test of the checker, not a known application defect.

The tests retain generated ITF traces, native observed steps and a validation
summary in `artifacts/verification/` (ignored generated output). Override the
location with `GOO_VERIFICATION_OUT`, or the Quint binary with `QUINT_BIN`.

`npm run model:test` additionally explores 100 simulated model traces at 30 steps
against the `safety` invariant. This is bounded simulation, not exhaustive model
checking. Expand the model and test boundaries alongside implementation changes;
do not interpret conformance to this small abstraction as proof of all behavior.

## Semantic package properties

`go test -run TestHegelPrebuilt -v .` generates missing/incorrect inventory
contracts across the ten profiles. It checks scope isolation, conservative
compatibility, ordering and duplicate invariance. This is contract-shape
verification, not proof of source completeness or a package-expansion model.

## Optional compiler-assisted model observations

`weaving/` is a separate module for Orchestrion/quintgo aspects at the actual
HTTP preparation seam. It has its own tool dependency graph and does not alter
the root runtime module. See ../docs/INSTRUMENTATION.md for commands, model
scope and toolchain compatibility notes.

The native suite also includes `TestHegelOperationAccountingStateSwarm`: 100
requested sequences of up to 30 steps exercise the actual OTel active-operation
accounting across start, error/panic completion and repeated finalization.
The optional woven SDK conformance scenario is documented in GOO_ANNOTATIONS.md.

Operation selection checks now execute the actual Node/browser kernel:
`go test -run TestHegelOperationHeatmapSelection -v .` generates 100 cases with
exact-ns boundaries and UInt64 duration edges.
`go test -run TestQuintObservedOperationNavigation -v .` validates a four-action
fixture trace with quintgo and rejects incorrect cell membership. See
[the scope and live browser journey](../docs/OPERATION_DRILLDOWN.md).

[Subject navigation](../docs/SUBJECT_NAVIGATION.md) adds an actual SDK scope check,
`TestHegelSubjectScopeIsolation` (100 generated cases), and a fifth subject action
in the observed Quint navigation fixture with cross-tenant rejection.

`TestHegelOperationPackagePlanning` adds 150 generated cases for deterministic
nonmutating YAML-derived plans, typed resource bindings and view-sensitive
revisions. See [operation packages](../docs/OPERATION_PACKAGES.md).

Native counters add `TestHegelCounterStateSwarm` (100 sequences, up to 30 actions)
and `TestQuintObservedCounterEpochs` (five actual evaluator observations plus
restart-rate rejection). [Semantics and live journey](../docs/NATIVE_COUNTERS.md).

Native histograms: `go test -run 'TestHegelHistogram|TestQuintObservedHistogram' -v .`
runs 100 state/swarm sequences, 150 generated merge cases and observed-trace
validation with reset/incompatible-merge negative controls. Model and bindings
are `model/histogram.qnt` and `model/histogram.yaml`; generated evidence is under
`../artifacts/`. See [native histogram limits](../docs/NATIVE_HISTOGRAMS.md).

Error ratios: `go test -run 'TestHegelErrorRatio|TestQuintObservedErrorRatio|TestHegelOperationAccounting' -v .` validates 100 ratio state/swarm sequences, explicit runtime outcome baselines, six observed Quint steps and four negative controls. Bindings/model are `model/ratio.yaml` and `model/ratio.qnt`; see [semantics](../docs/OPERATION_ERROR_RATIOS.md).

Fleet coverage: `go test -run 'TestHegelFleet|TestQuintObservedFleet' -v .` runs 100 state/swarm sequences and six observed Quint steps, with four negative controls. See `model/fleet.qnt`, `model/fleet.yaml` and [fleet limits](../docs/FLEET_COVERAGE.md).

Bounded comparisons: `go test -run 'TestHegelComparison|TestQuintObservedBounded' -v .` runs 100 state/swarm sequences and six observed Quint steps with three negative controls. See `model/comparison.qnt`, `model/comparison.yaml` and [comparison semantics](../docs/FLEET_COMPARISONS.md).

Saved investigations: `go test -v -run "Test(HegelInvestigation|QuintObservedInvestigation)" ./...` exercises immutable store state/swarm sequences and observed Quint transitions with rejected mutated-history/question controls. Model: `model/investigations.qnt`.

Exploration: `go test -v -run "Test(HegelExploration|QuintObservedExploration)" ./...` explores AST/filter/presence/page state and validates observed search→page→trace→logs against `model/exploration.qnt`, with negative population/scope/identity controls.

Cohort comparison adds a Hegel state/swarm machine for changing populations,
source disagreement and scope drift, plus Quint observed pin/selection/empty/changed/
scope-change validation with three negative controls. `cohorts_test.go` exercises
the actual Go comparison kernel; it does not claim verification of the browser or
source authenticity.

Telemetry-package verification adds Hegel state/swarm edits, save/reopen and window
changes over real pure/compiler/store functions, including adversarial string
round trips and immutable prior revisions. Quint observes draft/save/edit/reopen/
apply and rejects both a mutated stored definition and time/page leakage into
reusable semantics. Generated Go templates also pass the real annotation scanner
in root HTTP tests. These checks do not certify source coverage or source equivalence.

Discovery verification: `go test -run 'TestHegel(DiscoveryContextStateSwarm|DurationBucketPartition)$' -count=1 .`
checks clipped same-service context, page-independent counts and a complete disjoint
span-duration partition. With the real lab, run `GOO_DISCOVERY_QUINT=1 npx playwright test journeys/discovery.spec.js`
from the repository root. The browser ledger feeds `TestQuintDiscoveryLiveLedger`
through `model/discovery.yaml`; Go references validate the actual captured questions,
and negative controls alter context bounds, trace environment and saved grouping.
The model validates this bounded protocol, not ingestion completeness or causality.

`npm run test:discovery-campaign` from the root additionally runs the dedicated
Bombadil discovery actor for 60 seconds over the same isolated fixture. A required
prefix reaches trace/context paging, quick search, duration groups and package
preview/reuse before random actions. The driver requires full duration counts
beyond the record-page limit and passes the actor's own captured questions through
the discovery model, with negative controls and artifacts beside the browser trace.

Observed inventory: `go test -run TestHegelInventoryStateSwarm -v .` exercises
observe/reset accounting. Journey 31 always invokes `TestQuintInventoryLiveLedger`
with actual browser questions and three rejecting controls. See
[service inventory](../docs/SERVICE_INVENTORY.md); the optional Bombadil campaign
checks read-only UI actions and source-derived invariants over the same fixture.

General capture continuity: `go test -run TestHegelCaptureStateSwarm -v .` explores
60 sequences of up to 15 capture/refresh/source-change steps and verifies old
receipts and all retained pages. Journey 32 always invokes
`TestQuintCaptureLiveLedger` with actual receipt/page/successor/reopen data and four
negative controls. `npm run test:captures-campaign` adds read-only Bombadil paging
and inspection with data-aware invariants. See [the contract](../docs/GENERAL_CAPTURES.md).

Core workbench checks: `go test . -run TestHegelDurationSummaryStateSwarm` runs
100 state/swarm sequences up to 30 steps (append/reset/reverse, uint64 extremes,
exact sums, permutation and record-page independence). Journey 33 invokes
`TestQuintCoreBasicsLiveLedger` over actual measure/save/reuse/discover/cancel
observations. Five rejecting controls corrupt p95, archived sum, ambiguous parent
classification, cancellation observation and reused environment. This is observed-trace conformance;
it does not establish source completeness or all browser/database scheduling.

Grouped durations: `TestHegelGroupedDurationConservation` checks 100 generated
populations including uint64 extremes; sums/counts conserve across disjoint groups.
`TestQuintGroupedMeasuresLiveLedger` checks real measure/save/page observations,
with rejecting controls for wrong total p95, omitted saved group and altered page
group. Typed comparisons ignore JSON indentation while preserving exact integers
and missing/empty values. This is observed conformance, not proof of completeness.

`error_comparison_test.go` adds 100 Hegel state/swarm sequences of 25 steps for changing counts, source disagreement, environment and fixed windows. Journey 35 writes the real `GOO_ERROR_COMPARISON_LEDGER`; its Quint adapter validates measure/compare/save/reopen and rejects four corrupted ledgers. The saved-only Bombadil actor exercises all comparison measures. See [contract](../docs/ERROR_COMPARISONS.md).

`context_notes_test.go` adds 60 Hegel sequences of 15 state/swarm steps over actual local notes/marker storage, including preserved branches/history and half-open environment/window context. Journey 36 writes `GOO_CONTEXT_LEDGER`; Quint checks observed draft/revision/branch/reopen operations and six rejecting controls. See [contract](../docs/WORKBENCH_CONTEXT.md).

Scoped recipes: `go test -run TestHegelRecipeStateSwarm -count=1 -v .` checks
100 sequences of 25 recipe/signal/scope/predicate steps. The live journey writes a
ledger consumed by `TestQuintRecipesLiveLedger` (set `GOO_RECIPES_LEDGER`); its
preview/run/capture/narrative/reuse conformance checks have six rejecting controls.
`GOO_RECIPES_BOMBADIL=1 bash scripts/run_journeys.sh journeys/recipes.spec.js`
from the repository root includes the preview-only browser guard campaign.

## Shared measurements

`TestHegelMeasurementStateSwarm` runs 120 requested sequences of up to 35 steps
against the actual package adapter and gauge kernel: independent streams, random
observations, permutation invariance, immutable result replay and rejecting scope
tampering. Root tests additionally check freshness/conflicts, bounded input,
Float64 counter precision, lifecycle first-observation custody/corrections and
source-free general capture reopening after a source outage.

`journeys/measurements.spec.js` writes a real four-family ledger. The
`TestQuintMeasurementsLiveLedger` qobs/quint-go checker validates preview → execute
→ capture → reopen → refresh → document → reuse and rejects seven mutations. It
skips without `GOO_MEASUREMENTS_LEDGER`; browser execution supplies the ledger.
This protocol model checks the declared observations, not arbitrary distributed
scheduling. Existing lifecycle MBT covers the preparer separately. Optional
`GOO_MEASUREMENTS_BOMBADIL=1 bash scripts/run_journeys.sh journeys/measurements.spec.js`
checks source-free package preview/edit guards through a real browser.


Metric discovery: `metric_discovery_test.go` uses Hegel to preserve empty optional
version/schema metadata and distinguish versioned streams.
`metric_discovery_live_test.go` validates the real SDK → discovery → definition →
evaluation ledger against `model/metric_discovery.{qnt,yaml}` and rejects mutated
service, definition and evidence-basis ledgers. Storyboard 39 supplies the ledger;
`GOO_METRIC_DISCOVERY_BOMBADIL=1` also exercises discovery/scope-invalidation UI
invariants without evaluating or capturing measurements.

Investigation workspace: storyboard 41 supplies `GOO_INVESTIGATION_LEDGER` for
`TestQuintInvestigationWorkspaceLedger`. The quint-go model validates question-first
creation, exact span/log/metric citations, preserved parent/root lineage, fresh
reuse of the same definition and source-free reopening. Six mutated ledgers must
be rejected. Hegel's existing context state/swarm now mutates and forks the complete
question/finding/open-question brief and checks stored history. The new Bombadil
specification edits/reverts the brief while checking exact evidence references.

Contextual follow-ups: `TestHegelFollowupStateSwarm` selects observations, checks
same-population/identity/time guarantees and mutates returned plans while asserting
that stored evidence stays immutable. The state test exposed shared predicate
objects between returned actions; each plan now owns an independent copy.
Storyboard 42 supplies `GOO_FOLLOWUPS_LEDGER` for `TestQuintFollowupsLiveLedger`:
preview → duration/status → trace → logs → resources → gap → reopen. Seven mutated
ledgers must be rejected. `verification/bombadil/followups.ts` explores source-free
question previews while retaining exact references and explicit scope explanations.


Scope consolidation: the existing recipe/follow-up Hegel state/swarm checks now
validate shared descriptors, composed transforms and independent metadata, using
OTel/custom attribute names. The live follow-up ledger rederives the complete plan
and rejects nine mutations, including false population grain and relationship
metadata. Recipe replay compares the full shared plan. Bombadil follow-up invariants
require a checked same-population preview with explicit unbound convention metadata.
The existing model protocol is retained; these are stronger observed-value checks,
not a new proof of all telemetry semantics. See docs/QUESTION_SCOPE.md and VALIDATION.md.


Grouped time series: `TestHegelGroupedSeriesStateSwarm` exercises late input,
interval and inspection-page changes while preserving complete count/duration-sum
accounting and explicit no-data cells. Storyboard 43 produces a real collector/
ClickHouse ledger for `TestQuintGroupedSeriesLiveLedger`, with seven rejecting
controls for unit, boundary, local-read, capture-group, reuse-scope and reopen drift.
Its Bombadil actor explores local mean/p95/filter controls while stored evidence
and the captured scale remain fixed. The model observes a bounded plan → execute →
inspect → capture → reuse → reopen protocol; it does not prove all schedules or
telemetry completeness. See RECORD_MEASURES.md and VALIDATION.md for actual results.

## Numeric attribute measures

`go test -run TestHegelNumericBinding -count=1 -v .` checks valid-only arithmetic,
partition conservation and state/swarm appends, declared-type edits, paging and
frozen summaries (100 requested sequences, 25 steps for the state suite).
`GOO_NUMERIC_BOMBADIL=1 bash scripts/run_journeys.sh journeys/numeric-measures.spec.js`
from the repository root executes storyboard 44 with ordinary OTLP spans/logs and
writes a live ledger under `artifacts/numeric-measures/`. It checks signed,
unsigned and resource totals, groups and every time cell against reference values,
then runs the existing six-stage custody Quint model through quint-go. Eleven
mutations must be rejected: units/parsing policy, coverage, exact sum, cell bounds,
local source reads, reuse and reopen custody. The Bombadil actions switch local
numeric measures and search; invariants retain coverage and frozen evidence.
This is bounded controlled integration telemetry, not a completeness certificate.

## Question execution lifecycle

`TestHegelQuestionSessionStateSwarm` generates edits, fresh/fixed scope, starts,
cancels and out-of-order publish/fail completions. It executes the real JavaScript
session and compares every snapshot to an independent Go oracle (100 requested
sequences, 25 steps). `model/question_session.qnt` models generation/phase/publication
ownership; the live `TestQuintQuestionWorkbenchLive` checks observed transitions,
typed questions and service-discovery/capture custody with eight rejecting controls.
Run storyboard 45 with `GOO_QUESTION_BOMBADIL=1` through the standard journey wrapper.
The new Bombadil actor is source-free and bounded; existing discovery campaigns
now target the single Run action and preserve authored OR with AND shortcuts.

### Service browsing

`TestHegelServiceSuggestionScope` invokes the actual JavaScript suggestion planner
with generated fixed bounds and preferred/legacy/absent database attribute keys.
`TestQuintServiceBrowserLive` consumes `GOO_SERVICE_BROWSER_LEDGER`, checks the
three independent read scopes, narrowed HTTP question, local return and native
preview scope, and rejects six scope/selector/read mutations. Storyboard 46 writes
that ledger; `GOO_SERVICE_BROWSER_BOMBADIL=1` adds the 30-second service actor.

### Metric-only service discovery

`TestHegelMetricServiceBounds` generates native signal counts and fixed windows,
checks exact returned scope and rejects the exclusive upper timestamp bound.
`TestQuintMetricServicesLive` consumes `GOO_METRIC_SERVICES_LEDGER` from storyboard
47 and rejects seven count, scope, identity, dependency and local-read mutations.
`GOO_METRIC_SERVICES_BOMBADIL=1` enables the matching discovery/service/native
preview actor. Unit tests also cover malformed rows, duplicates, bounds, parameter
isolation and the global cardinality budget.

### Unified service catalog

`TestHegelServiceCatalogGrain` executes the actual JavaScript planner across 100
requested generated service/signal sets, including UInt64 counts beyond JS numeric
precision. Storyboard 46/48's observed ledger now includes actual catalog rows and
source scope. `TestQuintServiceBrowserLive` checks four independent service reads
and rejects eight scope, identity, custody, implicit-read and mixed-grain mutations.
Updated Bombadil actors use the single catalog entry.

Record-question links: `TestHegelQuestionLinkRoundtrip` generates 100 requested
scope/literal/view cases through the actual JavaScript URL codec and session
staging, including stale-response rejection. `TestQuintQuestionLinksLive` consumes
`GOO_QUESTION_LINKS_LEDGER`, produced by storyboard 49, with three rejecting
mutations for restore/history reads and changed source scope. The model is
`model/question_links.{qnt,yaml}`; the navigation actor is
`bombadil/question-links.ts`. Enable it with `GOO_QUESTION_LINKS_BOMBADIL=1` when
running `journeys/question-links.spec.js` through the normal wrapper.

Mixed investigation steps: `TestHegelInvestigationSteps` runs 100 requested cases
through the actual JavaScript envelope codec and verifies typed Go validation,
window relationships and detached native identity. `TestQuintInvestigationStepsLive`
reads `GOO_STEPS_LEDGER` from storyboard 50 and checks observed save/reopen/run
read counts with seven rejecting controls. Its model is
`model/investigation_steps.{qnt,yaml}`. Enable the saved-card exploration actor
with `GOO_STEPS_BOMBADIL=1`; the actor reopens records/native steps via ordinary
investigation controls and never runs a source query.

The existing `TestHegelContextNotesStateSwarm` also includes typed record/native
card creation and detached question-scope revisions after forks. Its 60 requested
15-step scenarios retain the history/context invariant across saves, ordering and
scope changes.

Linked evidence: `TestHegelLinkedEvidenceDifference` runs 100 requested cases
through actual JS capture projection/difference functions.
`TestHegelLinkedEvidenceStateSwarm` exercises 60 requested 15-step link/save/fork
scenarios against the Go store. `TestQuintLinkedEvidenceLive` consumes
`GOO_LINKED_LEDGER`; its model checks capture reads, failed publication, recovery
and read-free reopening with seven negative controls. Enable the stored-history
Bombadil actor with `GOO_LINKED_BOMBADIL=1` for storyboard 51.

Stored capture comparison: `go test -run TestHegelStoredComparison -v .` checks
100 requested precision/immutability/compatibility cases. Storyboard 52 writes
`GOO_EVIDENCE_PAIR_LEDGER` for `TestQuintStoredComparisonLive` (four rejecting
controls). `GOO_EVIDENCE_PAIR_BOMBADIL=1 bash scripts/run_journeys.sh
journeys/evidence-comparison.spec.js` also explores stored selections for 30 seconds.

Saved comparisons: `go test -run TestHegelSavedComparisonStateSwarm -v .` requests
60 scenarios of 15 save/fork/finding/order steps. Storyboard 53 writes
`GOO_SAVED_PAIR_LEDGER` for `TestQuintSavedComparisonLive`, with six rejecting
controls. Run `GOO_SAVED_PAIR_BOMBADIL=1 bash scripts/run_journeys.sh
journeys/saved-comparison.spec.js` for the real shared-browser/recovery journeys
and 30 seconds of frozen-card exploration.

Native window comparisons: `go test -run TestHegelNativeWindowComparison -v .`
requests 100 exact-rate, histogram share-conservation/layout-refusal, reversal, immutability
and reset-refusal cases. The source
is `native_window_comparison_test.go` (avoid the OS-specific `_windows_test.go`
suffix). Storyboard 54 writes `GOO_NATIVE_WINDOWS_LEDGER` for
`TestQuintNativeWindowsLive`, with six rejecting controls. Run
`GOO_NATIVE_WINDOWS_BOMBADIL=1 bash scripts/run_journeys.sh
journeys/native-windows.spec.js` for the live OTLP and frozen-reopen workflow.

## Observed window discovery

`TestHegelWindowDiscoverySupport` generates modern nanosecond offsets, gaps/flags
and multiple exact streams, checks every offered scope against the original
kernel points, preserves input immutability and withholds support when one member
has no valid intervals. `TestHegelSavedComparisonStateSwarm` continues to cover
publication and context recovery.

`model/window_discovery.qnt` and its adapter validate observed source-free discovery
and explicit selection, followed by two new captures and source-free frozen
reopening. `TestQuintWindowDiscoveryLive` consumes the journey ledger via
`GOO_WINDOW_DISCOVERY_LEDGER` and rejects six mutations.

```sh
GOO_WINDOW_DISCOVERY_BOMBADIL=1 bash scripts/run_journeys.sh journeys/window-discovery.spec.js
cd verification
go test -run 'TestHegelWindowDiscoverySupport|TestHegelSavedComparisonStateSwarm' -v .
```

The Bombadil actor explores compare/discover/select/clear without source capture.
Its invariants require the original incomplete decision to remain withheld and
the capture control to remain disabled until both scopes are selected.

## Reusable comparison questions

`TestHegelReusableQuestionRevisions` checks exact nanosecond scope and evidence
retention across immutable revisions for 100 requested cases. Existing saved
comparison state/swarm tests continue to cover revision and receipt recovery.
`model/reusable_questions.qnt` models source-free staging, failed publication,
save/reopen/fresh preview and explicit fixed capture/fresh execution. The live
checker consumes `GOO_REUSABLE_QUESTIONS_LEDGER` and rejects six mutations.

```sh
GOO_REUSABLE_QUESTIONS_BOMBADIL=1 bash scripts/run_journeys.sh journeys/reusable-questions.spec.js
cd verification
go test -run 'TestHegelReusableQuestionRevisions|TestHegelSavedComparisonStateSwarm' -v .
```

The Bombadil actor stages baseline questions and reverts unsaved drafts; it never
executes telemetry reads or publishes revisions. It checks the card bound and
that every added question has exactly one linked evidence entry.

## Fresh follow-up comparisons

`TestHegelFreshScopeContract` invokes the browser's pure scope projection from Go
for 100 requested generated nanosecond scopes, then validates the result with the
Go typed-step contract. It checks exact duration, unchanged definition and input
immutability. Existing revision/state/swarm and observed-support properties remain
applicable. `model/fresh_comparison.qnt` replays preview/selection, one capture,
source-free retry, save and reopening. `TestQuintFreshComparisonLive` consumes
`GOO_FRESH_COMPARISON_LEDGER` with six rejecting mutations.

```sh
GOO_FRESH_COMPARISON_BOMBADIL=1 bash scripts/run_journeys.sh journeys/fresh-comparison.spec.js
cd verification
go test -run 'TestHegelFreshScopeContract|TestHegelSavedComparisonStateSwarm' -v .
```

The Bombadil actor previews, selects compatible prior evidence and closes the
preview; it does not capture. Invariants require explicit prior selection and no
implicitly retained fresh evidence.

Captured breakdowns: `TestHegelBreakdownPartitions` checks partition accounting,
order independence and immutable missing/empty identities. Live journey 58 writes
`GOO_BREAKDOWN_LEDGER` for `TestQuintBreakdownLive`, replaying source-read boundaries
and rejecting six corrupted observations. Bombadil spec `comparison-breakdown.ts`
checks frozen group rows, focused question links and receipt disclosure actions.

Service dashboards: `TestHegelDashboardPlanning` checks pure deterministic planning,
capability gating, scope/selector preservation and input detachment. Journey 59
writes `GOO_DASHBOARD_LEDGER` for `TestQuintServiceDashboardLive`, using observed
query contracts, relationship identities and read-intent boundaries with eight
rejecting mutations. `service-dashboard.ts` explores list/distributed switching, service loading,
widget question inspection and native disclosures with scope, no invented-edge,
no fabricated failed-widget totals and containment invariants. Existing service
browser journeys retain stale-scope cancellation and log/metric-only coverage.

Namespace identity: `TestHegelNamespacePredicateIsolation` checks the logical
resource predicate against missing/empty/literal selectors with parameterized
SQL. Journey 60 writes `GOO_NAMESPACE_LEDGER` for `TestQuintNamespaceLive`: reported
cross-namespace edges, staged questions and actual declared-binding results are
replayed with four rejecting mutations. Optional `GOO_NAMESPACE_BOMBADIL=1` uses
`service-namespace.ts` to explore transitions and check namespace-preserving
record/native questions. Root tests reject foreign-namespace native/operation rows.

Declared services: `TestHegelDeclaredCatalogIsolation` generates counts and OTel
namespace presence/value cases while checking independent declaration order,
unchanged evidence and retained multiple bindings. Journey 61 writes
`GOO_DECLARED_LEDGER` for `TestQuintDeclaredServicesLive`, replaying actual absence,
fixed-question exploration and empty capture with six rejecting mutations,
including a changed namespace in the stored question.
`declared-services.ts` explores declaration disclosures and catalog/service
transitions while checking no invented telemetry, graph edges or observed widgets.

Reported edges: `TestHegelReportedEdgePlanning` generates parameterized service
identities and missing/empty/literal namespaces, checking ownership SQL and detached
question custody. Journey 62 writes `GOO_EDGE_LEDGER` for
`TestQuintReportedEdgeLive`: actual duplicate/ambiguous/out-of-window parent data,
full totals beyond the record page, namespace-scoped ownership, narrowed errors
and stored investigation selectors are replayed with six rejecting mutations.
`GOO_EDGE_BOMBADIL=1` explores the staged edge, reruns and error filters while
checking exact parent/child namespace preservation and page containment. The
journey separately checks a 390-pixel mobile viewport.
Comparison actors additionally check bounded visual bar geometry; journeys 35,
stored evidence, native windows and breakdowns exercise the common-scale visuals.

Display components: journey 63 is a standalone conformance fixture with no app
runtime or API reads. It checks server attributes, property assignment before
custom-element definition, detached/reactive updates, invalid JSON, exact large
values, zero/absence, keyboard disclosure and mobile geometry. Application
journeys 35/45/52/54/58/59/62 retain their observed Quint traces and rejecting
mutations. `service-dashboard.ts` now checks rendered quantile values/positions
against the supplied summary and explores native widget/latency disclosures.
These passive-display checks do not establish streaming delivery, gap recovery or
scroll-follow. The separate tail journey and rolling record comparison below
cover their own application adapters. Native live measure comparisons remain planned.

Live baseline comparison: `go test -run TestHegelRecordComparisonWindowCompatibility`
generates equal-duration shifted windows and incompatible environment cases.
Journey 69 writes an actual OTel-backed ledger; set `GOO_LIVE_COMPARISON_LEDGER`
and run `go test -run TestQuintLiveComparisonLedger` to replay it with six
rejecting controls. Run the actual preview/selection/start/freeze/link actor with:

```sh
GOO_LIVE_COMPARISON_BOMBADIL=1 bash scripts/run_journeys.sh journeys/live-comparison.spec.js
```
