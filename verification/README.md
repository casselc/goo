# Semantic properties and model conformance

Journey 105 adds observed source-free parent returns and cancellation of a delayed real fixed response to the entry Quint protocol. Its ledger counts both tail and fixed requests; `fixedRead` permits an explicit child read while the parent remains paused, and local returns must keep that total unchanged. The shared Bombadil filter actor also opens a fixed child and returns to the same paused selection.

Journey 104 and `verification/bombadil/live-filters.ts` cover live filter staging and explicit fixed-query continuation. `TestHegelLiveFilterStateSwarm` checks applied-selector isolation, exact namespace/value identity, deduplication and immutable snapshots against the actual JavaScript filter model.

Live entry uses `TestHegelEntryIntentStateSwarm` against the actual JavaScript intent gate. Journeys 102/103 feed observed request counts and lifecycle transitions into `stream_entry.qnt` through quint-go (`GOO_STREAM_ENTRY_LEDGER`); five rejecting controls catch hidden reads, duplicate startup, missing resume gaps and overflow. `GOO_ENTRY_BOMBADIL=1` enables the bounded stream-entry browser actor. See [Live Activity](../docs/LIVE_ACTIVITY.md).

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
typed questions and service-discovery/capture custody with nine rejecting controls, including a duplicate publication on retry.
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

The Bombadil actor explores compare/discover/select/clear without source capture and checks that each supplied preview has observed coverage and two UTC bounds. Journey 55 checks exact nanoseconds in timestamp tooltips and unchanged read counts during selection/identity inspection. `TestHegelNativePresentationProperties` additionally varies supported-window rational values and verifies unavailable coverage never presents an observed value.
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

Live span summaries extend the same observed state model: journey71 records
count, latency, recorded status, shared domains and source basis. Set
`GOO_LIVE_SPAN_MEASURES_LEDGER` and run `go test -run TestQuintLiveSpanMeasuresLedger`
to replay with ten rejecting mutations. `TestHegelLiveSpanMeasurePopulationAndReversal`
generates populations larger than the inspection bank, preserves Unknown status,
and checks immutability and reversal of descriptive changes. The source Go test
proves the three measures share one aggregate statement and survive a source-free
freeze. Run the actor with `GOO_LIVE_COMPARISON_BOMBADIL=1` and journey71.

Journey 72 writes `artifacts/live-counter/<run>/ledger.json`. Set
`GOO_LIVE_COUNTER_LEDGER` and run `TestQuintLiveCounterLedger` for observed replay
and seven rejecting mutations. `TestHegelCounterWindowRateAndCoverageState`
generates exact widths, increments and boundary crossings. Run the native UI actor:

```sh
GOO_LIVE_COUNTER_BOMBADIL=1 bash scripts/run_journeys.sh journeys/live-counter.spec.js
```

`verification/bombadil/live-counter.ts` adds rational common-axis geometry and
exact-window ordering to the live preview/start/freeze/link state invariants.

Journey73 writes `artifacts/exact-stream/<run>/ledger.json`. Set
`GOO_EXACT_STREAM_LEDGER` and run `TestQuintExactStreamLedger` to validate the
observed inventory/focus/run/save/live/freeze/clear projection and nine rejecting
mutations using `model/exact_stream.qnt`. `TestHegelExactStreamFocusClearState`
generates 100 twelve-action focus/clear sequences over 2–8 independent streams,
checking immutable inventory and refusal of off-stream or extra-empty dimensions.
Run the UI actor with `GOO_EXACT_STREAM_BOMBADIL=1` and journey73.

Journey74 writes `artifacts/native-navigation/<run>/ledger.json`. Set
`GOO_NATIVE_NAVIGATION_LEDGER` and run `TestQuintNativeNavigationLedger` for the
observed eight-phase navigation/focus projection and nine rejecting mutations.
The adapter compares actual DOM membership, open state, question, basis and read
count to retained evidence. `model/native_navigation.qnt` is deliberately a UI
custody projection, not a replacement metric evaluator. Run the actor with
`GOO_NATIVE_NAVIGATION_BOMBADIL=1` and journey74. Frontend generated sequences
check local paging/search immutability; existing Hegel exact-stream state tests
continue to cover the underlying Go semantic contract.

Journey75 writes `artifacts/service-native/<run>/ledger.json`. Set
`GOO_SERVICE_NATIVE_LEDGER` and run `TestQuintServiceNativeLedger` to replay its
actual automatic/value/browse/preview/evaluate/capture/reopen/gauge projection,
including nine rejecting mutations. It verifies latest-value versus window state,
namespace scope, no-read browsing and immutable exact-stream capture custody.
`model/service_native.qnt` describes UI/evidence coherence, not a new metric
kernel. Existing Hegel dashboard-planning properties cover unchanged scope/order
semantics. Run the native actor with `GOO_SERVICE_NATIVE_BOMBADIL=1`.

Native service selection: `TestHegelDashboardSelectionSequences` generates 100
twelve-step selection sequences across reordered inventories, explicit empty
choices and over-budget attempts, checking scope, determinism and detached
metadata. `TestQuintNativeSelectionLedger` consumes journey76's observed real
ledger (`GOO_NATIVE_SELECTION_LEDGER`) and replays defaults/edit/apply/capture/
reopen/clear/restore/supersede against `model/native_selection.{qnt,yaml}`. Nine
negative controls corrupt scope, membership, edit read counts, exact values,
selected/captured definitions, immutable receipt, clear and superseded output.
The companion Bombadil actor requires discovery/open/draft/apply/clear actions.

Native timelines: `TestQuintNativeTimelineLedger` consumes journey77's actual OTLP
gauge/counter/histogram observations (`GOO_NATIVE_TIMELINE_LEDGER`). Replay phases
cover typed visual inputs, local read-free inspection and ordinary capture/reopen.
Nine controls corrupt signed/unknown values, interval rates/states, uncovered time,
histogram P95 bounds, read counts, captured unit and immutable receipt identity.
Existing counter/histogram/measurement Hegel state/swarm properties complement
exact visual-geometry unit tests. Journey78 checks passive server/SPA lifecycle
and controlled keyboard selection independently of source semantics.

Native histogram heatmap display checks reuse the real retained-evidence paths in
journeys 77 and 54. Their existing observed Quint ledgers continue to check exact
stream identity, unknown/zero handling, compatible-window semantics and source-read
boundaries; the display adapter itself does not introduce a new evaluator or
custody transition. Journey 77 additionally runs `goo.timelineHeatmap` in its
saved-capture Bombadil actor, switching layouts and selecting zero/unknown cells.
Pure grid/adapter cases live in `tests/bucket-heatmap.test.js`; journeys 79/80 cover
SSR, detached/pre-definition properties, controlled selection, invalid data and
mobile layout. See [native heatmap contract](../docs/NATIVE_HEATMAPS.md).

## Explicit native histogram time buckets

`TestHegelHistogramTimeAlignedPopulation` and `TestHegelHistogramTimeStateSwarm` use an independent whole-export oracle, including local gaps, repairs, zeros, layout/epoch diagnostics and crossing widths. `TestQuintHistogramTimeLive` consumes `GOO_HISTOGRAM_TIME_LEDGER` from journey81 and rejects nine corruptions spanning derived counts, unavailable values, width, display values, read counts and capture/reopen identity. Model: `model/histogram_time.qnt/.yaml`; optional browser actor: `verification/bombadil/histogram-time.ts`.

## Evidence-summary observed trace

`TestQuintEvidenceSummaryLive` consumes `GOO_EVIDENCE_SUMMARY_LEDGER` from journey82, validating capture state/counts, separately labelled source/storage clocks, local read counts, linked metadata and immutable reopening. Eight corruptions must fail. Model: `model/evidence_summary.qnt/.yaml`; browser actor: `verification/bombadil/evidence-summary.ts`. No new measurement evaluator or semantic property was introduced by this display extraction.

Service-summary adoption extends journey59’s observed ledger with per-widget display snapshots and source result metadata. `TestQuintServiceDashboardLive` checks fixed mode, result state, source-read/basis/count correspondence and absence of invented custody; eleven rejecting mutations include frozen-mode, read-clock and custody corruption. Journeys59/75 Bombadil actors exercise summary disclosures and assert fixed mode/read-clock contracts. Pure display tests cover exact UTC formatting, negative/zero epochs and invalid bounds; the standalone browser journey checks keyboard focus and exact tooltips.

Comparison presentation checks now assert horizontal track alignment/equal widths
and separately labelled frozen-side source/storage clocks in journey52. The
existing stored, fresh, saved and complete-native-window observed Quint ledgers
remain authoritative for compatibility, immutable receipts and read budgets.
The stored-comparison Bombadil actor opens both scope summaries and checks frozen
mode, clock labels and identical bar domains. No evaluator/model transition was
added for this passive presentation change.

The selected-item shell uses application-owned keys and DOM content. Journey84
checks retained nodes, disclosure/body scroll, no autofocus, close-to-opener and
clear/mobile behavior in a synthetic standalone fixture. Journey35 checks real
record/trace inspection and distinguishes source-free record inspection/close
from the trace adapter's existing linked-log read on selection. Journeys77/29
exercise the native and typed-inspection Bombadil actors, now checking close/
reopen focus and retained selection. Their existing Go observed Quint ledgers and
rejecting controls remain the semantic oracle; no new metric evaluator or Go
planner state was introduced.

## Core experience refinements

See [the implemented workflow](../docs/CORE_EXPERIENCE_REFINEMENTS.md) for retained
application visits, typed population selection, compact diagnosis, trace
breakdowns and independently loaded context. These compose passive display
elements; application adapters own navigation, read lifecycle and scope. Journey
85 exercises real OTLP data and refreshed screenshots; the core-refinements
Bombadil actor covers selection and returning to retained evidence. Hegel
selection properties and the question-session state swarm complement observed
quint-go validation against the existing question-links model.

Trace findings reuse `model/linked_evidence.qnt` rather than introducing a second
custody state machine. `TestHegelTracePageExamples` checks exact page membership,
permutation invariance, deduplication and incoherent-read withholding over 100
cases. `TestHegelLinkedEvidenceStateSwarm` exercises immutable revision history.
Run the live workflow and composer actor with
`GOO_TRACE_FINDING_BOMBADIL=1 bash scripts/run_journeys.sh journeys/trace-finding.spec.js`.
Its ledger drives `TestQuintTraceFindingLedger`; five rejecting controls alter
scope, trace identity, linked evidence and reads during stored reopening.

`TestHegelLinkedTimelineScope` generates 100 bounded service-window transforms
against the actual JavaScript adapter, preserving namespace and native identity.
`GOO_LINKED_TIMELINE_BOMBADIL=1 bash scripts/run_journeys.sh journeys/linked-timeline.spec.js`
emits real OTLP spans, logs and metrics, verifies preview/Apply/reset and partial
log failure, and replays the observed ledger through `question_links.qnt` with
quint-go. Rejecting controls cover foreign environment, altered bounds, preview
reads, reset-read drift and lost predicates. The passive component conformance
journey needs no source and verifies keyboard and detached-property behavior.

Population workspace checks: `TestHegelPopulationPortableSelectors` generates
100 missing/empty portable-selector plans; `TestHegelPopulationViewStateSwarm`
checks 60 × 15 immutable view edit/save sequences. Journey 90 emits real OTel
spans/logs and saves only configuration. Its `TestQuintPopulationWorkspaceLedger`
replays source reads versus source-free preview/save/reopen with quint-go and
rejects scope, selector, presentation-validation and hidden-read mutations. Run
`GOO_POPULATION_BOMBADIL=1 bash scripts/run_journeys.sh journeys/population-workspace.spec.js`
for the saved-view run/layout actor as well as partial-failure/cancellation checks.

Population journey 91 emits `GOO_BREAKDOWN_LEDGER` for
`TestQuintPopulationBreakdownLedger`: observed reads versus source-free save/open
transitions are checked through quint-go/question_links. Five controls reject
hidden save/open reads, a foreign baseline, a missing baseline replay read and an
invalid ranking. The population Hegel state/swarm rules now include breakdown and
baseline edits, with immutable published histories checked after each step.
`verification/bombadil/population-breakdown.ts` exercises explicit Run, local
ranking and previous-window comparisons, retaining overflow and runtime-error
invariants.

Population live journey 92 emits `GOO_POPULATION_LIVE_LEDGER` for
`TestQuintPopulationLiveLedger`. It observes explicit read/subscription initiations
and source-free saved-view reopening, then checks tail reference identity,
100-row retention and visible reconnect gaps. Five negative controls reject hidden
opening reads, unsupported signals, unbounded retention, erased gaps and changed
baseline references. The existing tail model and generated properties remain
shared with the standalone log tail. Hegel saved-view state/swarm rules now toggle
live settings too. `bombadil/population-live.ts` starts/pauses the selected tail
with bounded-row, overflow and runtime-error invariants.

Recent-entry journey 93 records `GOO_RECENT_LEDGER` for
`TestQuintRecentEntryLedger`, replaying the existing question-links model through
quint-go. Five mutations reject hidden preview/reopen reads, a foreign companion
environment, erased selection filters and an implicit tail. Run with
`GOO_RECENT_BOMBADIL=1` to execute `bombadil/recent-entry.ts`: explicit recent run
and local log-page switching, with containment/runtime/no-tail invariants.
Existing Hegel population-selector and immutable saved-view state/swarm checks
exercise the unchanged Go planner/storage used by this composition.

Range-drilldown journey 94 writes `GOO_RANGE_LEDGER` for
`TestQuintRangeDrilldownLedger`, using `model/population_drilldown.{qnt,yaml}` through
quint-go. It records bounded child depth and source-read initiations; seven
mutations reject hidden reads, widened bounds, altered baselines, lost retained
rows, extra subscriptions and an invalid return depth. `TestHegelBrushPopulationScope`
generates 80 real JavaScript brush-to-query projections and checks the resulting
queries through Go compilation, including missing/empty predicates and parent
nonmutation. The existing 60 × 15 saved-view state/swarm campaign checks reused
immutable configuration behavior. `GOO_RANGE_BOMBADIL=1` executes explicit
recent-run, stage and return actions with depth/containment/no-live invariants.

Categorical population intent is checked by `TestHegelFacetSelection` (60 cases
against Go compilation/reference matching). `TestQuintFacetLedger` consumes
`GOO_FACET_LEDGER` from journey 96 and reuses `population_drilldown.yaml`, rejecting
hidden preview/return reads, changed companion environment, lost predicate and
incorrect return depth. Optional `GOO_FACET_BOMBADIL=1` runs
`verification/bombadil/facet-selection.ts` over real ingested telemetry.

Record-to-population verification: `TestHegelSharedRecordFields` checks 60
generated shared/mismatched/missing/non-string cases against Go query compilation
and reference matching. `TestQuintRecordPopulationLedger` consumes
`GOO_RECORD_LEDGER` from journey 97, reusing the retained-population model and
rejecting eight controls involving hidden reads, source/predicate/depth changes,
provenance source changes, duplicate observations and out-of-window event time.
Optional `GOO_RECORD_BOMBADIL=1` checks the per-page selection bound, viewport,
no-live-tail and completion/exception properties.

### Cross-product annotations and optional XTDB retention

`TestHegelMeaningStateSwarm` interleaves roots, corrections of old/new revisions
and withdrawals against an independent parent-edge/head oracle. Root Go tests
also exercise actual concurrent file publication. `TestHegelRetainedSelectionIdentity`
generates equal payload occurrences, selection permutations, large numeric/time
tokens and integrity failures without collapsing occurrence identity.

`model/annotations.qnt` and the quint-go `TestQuintAnnotationLedger` replay observed
metadata-only browser actions from journeys 98–99. Negative controls inject a
source read, an implicit publication and incoherent provenance and must fail.
`verification/bombadil/annotations.ts` explores draft/cancel/history on both exact
file annotations and native retained-evidence pages. Set `GOO_ANNOTATION_BOMBADIL=1`
for the corresponding Playwright journeys.

Real native tests are opt-in, separate from the pure default Go suite:

```sh
GOO_XTDB_TEST_URL='postgres://xtdb@127.0.0.1:15432/xtdb?sslmode=disable' \
  go test ./retention -count=1 -v # from repository root, not verification/
```

They cover native system-time cutoff queries, exact record payloads, independent
review branches, a staged transaction interrupted before commit, idempotent retry
and concurrent request recovery. The optional Compose overlay and prototype limits
are in [the annotation contract](../docs/WORKBENCH_ANNOTATIONS.md).

## Exact live-page retention

`TestHegelResultSnapshotStateSwarm` generates read/select-old/reject-foreign
sequences and checks immutable page bytes, exact numeric tokens, original offsets
and equal-payload occurrences. Root Go tests use a controlled clock for TTL,
capacity and byte bounds, plus concurrent cache readers. Native opt-in tests cover
correction rollback, concurrent identical retries, changed-intent rejection,
expired-page committed-receipt recovery and await-token reads on new connections.
Journey 100 records the actual original result and canonical retained manifest;
`TestQuintAnnotationLedger` validates token/question/basis/member continuity and
retry transitions, with nine rejecting controls, including a duplicate publication on retry. This small model does not prove
multi-node replica behavior or production isolation. The live-retention Bombadil
campaign observes native POST count while selecting, drafting and clearing.

## Retained evidence catalog, citations and pending recovery

`TestHegelPendingSaveStateSwarm` executes the real JavaScript journal in a persistent
Node worker, generating stage/reload/retry/ack/changed-intent sequences and checking
immutable request bodies and unrelated pending entries. Root tests cover catalog
branch/withdrawal tags, exact service/environment filters, half-open capture time,
cursor invalidation and native pinned-review resolution. Journey 101 writes a real
browser ledger after successful responses are lost and recovered across reloads.
`TestQuintEvidenceLibraryLedger` checks canonical manifest/review/document hashes,
pinned old review versus later correction, current catalog tags, pending counts
and publication/read transitions. Eight negative controls must fail. Its Bombadil
campaign explores readonly citation/catalog/pending surfaces; it does not claim
production multi-tab synchronization or multi-node replica conformance.

Journey 106 compares real log/span presentation between Live and fixed results,
asserts source-free detail/context expansion and exercises SSR, detached property
updates, reconnect, text safety and mobile display in the standalone reference.
`TestHegelRecordPresentationProperties` generates supplied parent graphs for order/
cross-trace invariance, explicit-edge validity, bounded unique paths, non-mutation
and exact time/individual-duration display. [Contract](../docs/RECORD_PRESENTATION.md).

Journey 46 additionally validates the existing supporting service span page as a
source-free local inspection surface: shared table cells/detail/UTC/context, exact
question-link staging, retained selection on return, keyboard close focus and
readable mobile row/time geometry. Its Bombadil actor inspects that page before
following a convention suggestion and checks the shared inspector body. Dashboard
planning and entry/filter/parent properties retain their existing Hegel campaigns;
service/dashboard/entry ledgers retain their existing Quint rejection controls.

Saved record receipt presentation is shared between capture pages and document
references. Journey 32 checks exact count/source/custody facts, page offset, local
disclosure and fixed-question staging without source reads; Run explicitly rereads.
Its Bombadil actor opens scope and checks summary identity/count/custody/page
against the immutable receipt while traversing saved pages. Journeys 41/45/50
cover document and fixed-draft compatibility; existing capture Hegel state/swarm
and Quint capture/investigation protocols remain the validation basis.

Journey 107 exercises the extracted saved owner with real generic OTel records:
no query builder/control nodes, local inspection, explicit all-service trace/log
reads, keyboard close return, surrounding logs, capture switching and rejection of
a held real context completion. `TestQuintSavedOwnerLedger` replays every observed
source read/local transition through the existing read/history model and requires
four counter/read/coherence mutations to fail. It does not model trace completeness.
The capture Bombadil invariant now requires absent editor nodes, rather than hidden
controls. Captured paging/retention/package and numeric/grouped compatibility keep
their existing Hegel/Quint campaigns.

`TestHegelNativePresentationProperties` runs 200 generated supplied-series cases
against the pure native presenter: exact nanosecond latest selection, permutation
invariance for distinct timestamps, no fallback to an older valid interval, zero
versus unavailable, coverage counts and non-mutation. Native comparison properties
and counter/histogram state campaigns protect evaluation separately. Existing browser
Quint ledgers retain their observed protocol limits; no exhaustive native UI model
is implied. See [presentation contract](../docs/NATIVE_PRESENTATION.md).

## Lazy fixed-authoring boundary

Journey 109 checks absent editor DOM on fresh Activity, source-free explicit mounting,
same draft/DOM identity after paused return and inert deep-link staging. Journeys
102/105 assert the lazy mount boundary during real tail/fixed-read transitions and
reuse the existing observed Quint ledger with rejecting controls. The entry Bombadil
campaign now asserts zero eager fixed editors during its pause-only exploration;
the live-filter campaign checks zero before/after fixed authoring and exactly one
while its fixed child is open.
Existing entry-intent and live-filter Hegel state/swarm campaigns protect the
unchanged selector/source protocol; they do not model every DOM lifecycle detail.

## Compact definition libraries

`TestHegelDefinitionLibraryProperties` generates 200 supplied catalogs and page
configurations, checking typed exact identities, no title-based collapse, complete
filtered paging, missing matches and non-mutation. Journey 110 checks selected
editor/annotation drafts, exact revision actions, delayed real list response
rejection and failed-refresh retention; existing annotation journey 98 replays
its real Quint ledger and rejecting controls. The definition-library Bombadil
actor inspects, filters and pages without invoking source or publication actions.

Journey 111 holds real reusable-plan responses across selection/close/workspace and
fixed suspension into Activity, then proves that only the subsequent explicit Run
queries a source. The existing core workflow actor now selects the shared revision
row before Run and checks bounded rows/exact selected keys. Journey 33 retains its
real full-population reuse Quint ledger and rejecting controls; these do not claim
an exhaustive model of every inspector state.

Journey 112 exercises real document roots/branches through server cursor pages,
metadata inspection with a pending note, guarded explicit Open, page-scoped search,
failed catalog refresh retention and mobile layout without telemetry reads. The
Hegel catalog generator now includes document as well as record/measurement typed
identities. The context-notes Bombadil actor includes browse/inspect/close actions,
checks at most 25 local rows and exact typed document selection, while retaining its
immutable base/reference/draft invariants. Journey 36 still validates real observed
app/evidence/narrative transitions with Quint/quint-go and rejecting controls; the
catalog DOM is not claimed as a new exhaustive formal model.


## Retained catalog owner consolidation

Journey 101 also checks metadata-only compact row inspection, a real catalog reply
held across a filter change, failed-refresh preservation and exact old review
citations after correction/reload. Its actor deterministically browses, selects and
closes a catalog snapshot before inspecting recovery; invariants bound server-page
rows to 20 and snapshot keys to a load generation plus exact manifest ID. It retains
zero publication, unchanged document revision and exact citation invariants.
`TestQuintEvidenceLibraryLedger` checks the actual observed recovery/citation ledger
and eight rejecting controls; it does not model the full catalog DOM. Existing
Hegel pending-save state/swarm, retained selection identity and library properties
continue to cover unchanged protocols. Direct retention-client tests additionally
check exact body/key replay after lost acknowledgments, optional-store failures and
citation read-bound precedence across the extracted transport boundary.


## Record field presentation — first B2 slice

`testdata/protocol/record-fields.json` is consumed by root Go query/matcher/canonical
wire tests and JS display/identity tests. Ten cases cover missing versus empty,
omitted string defaults, resource/attribute separation, literal sentinel/Unicode
text and exact uint64 bounds. The live-filter Hegel state/swarm campaign now invokes
both display paths while asserting unchanged draft/applied intent and equivalent
empty-default identities. Journey 113 checks source-free fixed/package authoring
and mobile chips. Journey 104 retains its real OTel source, paused return and
observed Quint trace with rejecting controls; its actor checks the qualified live
selector text across staging and return. No request-owner protocol changed.


The B2 namespace continuation adds `service-namespaces.json` Go/JS fixtures for
eight selector representations and six observed resource maps. Go membership and
record predicates agree; JS selection keys/canonical aliases remain stable and
shared labels distinguish literal sentinel names from missing/empty modes.
Journey 60 adds ordinary OTel span/gauge populations for literal `missing` and
`present empty`, retaining the real Quint namespace trace with four rejecting
controls; its actor checks distinct labels when the service-list table is present.
Existing Hegel namespace isolation and entry/live-filter state/swarm campaigns
protect unchanged logical selector and ownership semantics.


## Native/shared-question wire fixtures — B2 continuation

`question-steps.json` freezes 16 accepted/refused wire cases. Root Go checks typed
validation/canonical output, invokes actual JS link round trips and pins native
measurement digests. JS checks comparison equivalence against Go canonical forms,
changed selectors, family separation and exact bounds. The expanded investigation
Hegel property adds exact native strings above 2^53 and default-alias checks.
Journey 114 uses real OTel gauge data and Go-canonical capture metadata; its actor
checks source-free preview/close and permits no POSTs. Existing mixed-step and
fresh/linked journeys validate actual observed ledgers with rejecting controls;
no universal browser validator or complete UI state model is claimed.

## Immutable publication mechanics (C1 first adoption)

`TestHegelImmutablePublicationSequence` generates 100 first-winner/duplicate-write
sequences on real temporary files, including exclusive collision refusal, both
initial sync policies and file-only duplicates. Run with `go test -run
'^TestHegelImmutablePublicationSequence$' -count=1` in this module. Root race tests
and store golden fixtures accompany it; see the [publication contract and remaining
store inventory](../docs/IMMUTABLE_PUBLICATION.md). No browser/Quint transition
protocol changed in this filesystem extraction.

## Bounded planning JSON (C2)

`TestHegelPlanningJSONBoundaries` generates 100 Unicode/escaping, exact typed integer,
whitespace and byte-budget cases against the shared decoder/writer. Run
`go test -run '^TestHegelPlanningJSONBoundaries$' -count=1` in this module. Root HTTP
tests also run 32 before/after wire fixtures from `testdata/http/planning.json`;
see [route policy and adoption boundaries](../docs/HTTP_BOUNDARIES.md#pure-planning-json-adoption).
No new browser execution-state or Quint protocol is introduced by this extraction.

## Bounded source HTTP (C2 native and record adoption)

`TestHegelSourceHTTPByteBounds` generates 100 binary payload/byte-budget/partial-read
failure cases against the bounded-body primitive. Run
`go test -run '^TestHegelSourceHTTPByteBounds$' -count=1` here. Native request/error/
row/scanner contracts and the gated readonly real ClickHouse test live in the root
measurements package; see [ownership and live-test variables](../docs/HTTP_BOUNDARIES.md#native-source-http-mechanics--first-transport-adoption).
No browser execution-state or Quint protocol changed in this transport extraction.

Record transport contracts in `exploration/source_transport_contract_test.go` also
pass before/after shared-mechanics adoption. They preserve record-specific
status-before-read, object validation and 1,001-row behavior. The readonly gated
record integration test roundtrips every ASCII control plus DEL against ClickHouse;
see [record transport ownership](../docs/HTTP_BOUNDARIES.md#record-source-transport-adoption).

Prebuilt transport now also adopts the bounded-read helper behind pre-extraction
contracts in `prebuilt/source_transport_contract_test.go`, preserving raw parameter
values, status-before-body and family evaluation. Its gated readonly ClickHouse
check exercises a deliberately empty real-table scope, not positive live data.
See [prebuilt transport adoption](../docs/HTTP_BOUNDARIES.md#prebuilt-http-transport-adoption).

Operation spans also use the bounded-read primitive behind before/after contracts
in `annotations/source_transport_contract_test.go`. Raw parameters, typed rows,
status-before-body, exact duration strings and conformance stay adapter-owned.
The gated readonly real-table check uses an absent service and fixed window;
see [operation-span adoption](../docs/HTTP_BOUNDARIES.md#operation-span-transport-adoption).

Counter transport adopts the same primitive behind independent before/after tests
in `annotations/counter_transport_contract_test.go`: body reads and immediate close
precede status, and later-contract failures return no partial rates. Existing
counter/fleet semantic tests remain separate. The gated readonly sum-table test
uses an absent service; see [counter adoption](../docs/HTTP_BOUNDARIES.md#counter-transport-adoption).

Histogram transport completes the audited ClickHouse helper adoption behind
`annotations/histogram_transport_contract_test.go`. Independent contracts preserve
read/status/overflow precedence, immediate closure, exact bucket counts above 2^53
and no partial later-contract success. The gated readonly histogram-table test
uses an absent service; see [histogram adoption](../docs/HTTP_BOUNDARIES.md#histogram-transport-adoption).

Telemetry package JSON uses the shared decoder/responder behind 68 exact before/
after wire cases in `testdata/http/telemetry.json`. Hegel planning-JSON properties
apply unchanged; route tests separately prove no files from planning/rejection
and preserve explicit publication. See [telemetry JSON adoption](../docs/HTTP_BOUNDARIES.md#telemetry-package-json-adoption).

Native measurement/registry JSON adopts the same helpers behind 68 frozen cases
in `testdata/http/native-json.json`, retaining its distinct trailing-input message.
Tests count source reads and file effects; existing positive execution/custody and
publication fixtures remain separate. See [native JSON adoption](../docs/HTTP_BOUNDARIES.md#native-measurement-and-registry-json-adoption).

Fixed exploration JSON uses shared mechanics behind 68 frozen cases in
`testdata/http/exploration-json.json`. Tests retain comparison's 1 MiB budget,
source routes' 16 KiB budgets, exact large counts, source-free rejection/comparison
and 15/30/55-second source deadlines. Positive source/snapshot behavior stays in
existing tests; see [fixed exploration adoption](../docs/HTTP_BOUNDARIES.md#fixed-exploration-json-adoption).

## OpenAPI HTTP contract

[API.md](../docs/API.md) documents generation and checks for OpenAPI 3.1, all Go
routes and Python-only API operations. CI validates the 236 frozen HTTP cases,
56 admitted requests and rejecting unknown-field/error/SSE controls; it executes
actual Go wire fixtures and the reset/gap SSE fixture. Go remains authoritative
for semantic validity, source/custody effects and capability lifetimes. The
standalone API docs journey checks the real Python handler without lab traffic.
Existing Hegel/Quint/Bombadil workbench models are unchanged by this docs surface.

Observed API continuation: `TestAPISavedEvidenceObservedContracts` exports 12 actual
base responses or 19 with a real XTDB URL, while refusing source reads after the
live draft is held. Proxy loopback tests export 20 responses and check idle SSE
abort cleanup. `scripts/check_openapi.py --observations ... --proxy-observations ...`
validates both. CI runs base observations; local native-store/deployed coverage is
recorded separately in [API acceptance](../docs/API_ACCEPTANCE.md). Existing deployed
core/capture/retention journeys retain their Quint ledger checks and rejecting
controls. No new workbench protocol/model transition was introduced.

`TestQuintComparisonInspectionLedger` replays journey 127 via
`GOO_COMPARISON_INSPECTION_LEDGER`. The existing drilldown model checks source-free
staging/return and exact read counts; coherence additionally pins each selected
side’s complete question, original UTC window, parent predicate and qualified
missing/empty group. Six rejecting controls cover hidden reads and changed scope.


Retained comparison series: `TestHegelComparisonSeriesGeometry` checks the actual JS BigInt domain and bounded brush properties. Journey 128 supplies `GOO_COMPARISON_SERIES_LEDGER` to `TestQuintComparisonSeriesLedger`, replaying `population_drilldown.qnt` with five rejecting controls. `GOO_COMPARISON_SERIES_BOMBADIL=1` enables the 25-second alignment/selection/stage/Back actor.

Rolling record series extend `TestQuintLiveComparisonLedger` to twelve rejecting controls, checking exact baseline/frozen arrays, pinned selection, source-free staging/return and fixed interval domain. Journey 69 and the existing live Bombadil actor execute the refresh/select/stage/Back/Freeze path; Hegel geometry covers supplied fixed domains and overflow.

`TestHegelInstrumentCapabilityHints` exercises the real Go capability planner
across native kinds, units and exact-stream permutations, requiring immutable
metadata, honest presence totals and no verified-state invention. Journey 130 reuses `TestQuintMetricDiscoveryLedger` with
actual Go discovery/result bindings and three rejecting mutations. Its configured
Bombadil campaign retains scope in `campaign.ts` beside the compact ledger.

Instrument capability Hegel checks now exercise the actual Go planner, not a
parallel JavaScript classifier. `TestQuintInstrumentPresenceLedger` replays real
discovery/disclosure effects against independent identity rows and rejects seven
forgeries, including extra source reads and false key/instance coverage.


E1 acceptance: `TestHegelInstrumentCandidateDoesNotPromoteEvidence` varies actual
histogram observations under the same metadata candidate; missing time/values
stay unknown and invalid identities/counts are refused. Existing counter/histogram
state/swarm and observed Quint checks remain the kernel authorities. Journey 133
reuses `TestQuintMetricDiscoveryLedger` separately for observed and unknown real
results (three rejecting controls each) and configures the existing Bombadil
inspection actor for a consumer with missing instance identity. No new model or
browser-side convention classifier is introduced. See [the audit](../docs/E1_ACCEPTANCE.md).
