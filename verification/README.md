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

Time/duration partitions: `TestHegelTimeDurationPartition` executes the actual JS
transform, compiles both sides with Go and checks disjointness/coverage using Go's
reference interpreter at exact timestamp/duration edges and UInt64 extremes.
`TestQuintRegionPartitionLedger` replays journey 134 through the existing
population-drilldown model and rejects changed read counts, bounds, scope and
non-disjoint rest predicates. `verification/bombadil/region-partition.ts` explicitly
loads the parent, then requires source-free staging and return. See
[the contract](../docs/REGION_PARTITIONS.md).


The [bounded D2 audit](../docs/D2_ACCEPTANCE.md) reuses the population properties,
immutable view state/swarm and observed drilldown traces for columns, expansion,
broadening and partition operations. Journey 134 now additionally checks exact UTC
selector display and question-trail opener focus. Its Bombadil actor requires focus
return after source-free region staging; no new execution model is introduced.


Tail membership prerequisite: `TestHegelTailMembershipStateSwarm` executes the real
unwired Go bank across issuance, time advancement and resolution, with TTL checks
after every action. `TestQuintObservedTailMembership` replays real issue/resolve/
expire/refuse observations through `model/tail_membership.qnt`; four forged traces
must fail. This model does not claim durable publication. Existing XTDB interruption,
concurrent-request and expired fixed-page recovery tests separately establish
reusable retry mechanics. See [the contract/status](../docs/TAIL_SELECTION_CUSTODY.md).

Tail custody follow-up: `TestTailManifestProperties` checks immutable serialization
and rehashed invalid provenance using Hegel. With `GOO_XTDB_TEST_URL`,
`TestQuintObservedTailPublication` replays actual publish/recover/conflict/absent
outcomes against `model/tail_publication.yaml`, including four rejecting controls.
This does not cover the still-unwired SSE, HTTP or shared catalog readers.

`TestQuintObservedTailReviews` uses real optional XTDB reads after publication,
correction, withdrawal and original-request retry. `model/tail_reviews.yaml` checks
head/active counts, current catalog visibility and immutable original manifest/
pinned review, with four rejecting controls. `TestTailManifestProperties` also
checks the typed origin view cannot masquerade as a captured result page.

`TestQuintTailContentLedger` consumes journey 138's real browser ledger through
`GOO_TAIL_CONTENT_LEDGER`. It reuses the annotation lifecycle model with five
rejecting controls for source reads, publication counts, pinned revision, recorded
custody time and exact manifest. `verification/bombadil/tail-evidence.ts` exercises
local provenance/attribute disclosures and search without source reads or writes.
The existing evidence-library actor also exercises the new origin selector.

`TestHegelLatencyHistogramPartitions` checks generated contiguous bucket ranges, including reverse intent, missing/empty parents and exact upper boundaries through the Go compiler/matcher. `TestQuintLatencyHistogramLedger` reads `GOO_LATENCY_HISTOGRAM_LEDGER` from journey 153 and uses the existing drilldown model with five rejecting controls; duration-only scope differs from the earlier time/latency rectangle.

`TestHegelActivityLatencyOwnership` generates 60 ownership sequences of 60 operations and checks immutable query plans through Go. `TestQuintActivityLatencyLedger` reads `GOO_ACTIVITY_LATENCY_LEDGER` from journey 155, checks rolling read/staging scope, and adapts actual fixed selected/rest navigation into the existing drilldown model with five rejecting controls.

`TestHegelActivityBaselineAccounting` generates 40 exact large-count comparisons through Go and the JS result guard, with wrong-basis/incomplete-group controls. `TestQuintActivityBaselineLedger` consumes `GOO_ACTIVITY_BASELINE_LEDGER` from journey 156, checks immutable receipt/full-population counts and replays source-read versus choose/clear actions through `model/activity_baseline.yaml`; five mutations reject hidden reads and baseline drift.

`TestHegelSavedActivityBaselineIdentity` generates 50 pin/replace/clear round trips and checks old identities. `TestQuintSavedActivityBaselineLedger` consumes `GOO_SAVED_ACTIVITY_BASELINE_LEDGER` from 158 and checks configuration identity plus source/receipt/comparison-free reopen and source-free retry; six controls reject reads, reference drift, lost clear and implicit start.

`TestHegelNumericHistogramExactPartition` and
`TestHegelNumericHistogramRangePartition` generate canonical signed/unsigned
values, nonvalid parse states, grouping dimensions and selected/rest ranges.
Independent standard-library parsing and arbitrary-precision boundaries check
coverage, exact bin counts, full partition, immutable scope and safe sentinels.
`TestNumericHistogramDomainEdgesAndRefusal` adds every derived boundary ±1 and
invalid geometry/index controls. Run these in the verification module:

```sh
go test -run 'Test(HegelNumericHistogram|NumericHistogramDomain)' -count=1
```

`TestQuintNumericHistogramLiveLedger` reads an actual journey 174 ledger through
`GOO_NUMERIC_HISTOGRAM_LEDGER`, checks source fixture membership, complete bins,
exact selected/rest and comparison query ranges, capture/YAML reuse, retained
reopen and no-read gestures, then replays observed phases with quint-go. Eighteen
mutations must be rejected. The Bombadil `numeric-histograms.ts` actor explicitly
runs the real fixture once and then explores keyboard bins, exact details,
configuration edits/repair and departure without additional source reads.

`TestHegelNumericComparisonExactRationalsAndCoverage` uses an independent
standard-library canonical integer parser and big.Int/big.Rat summary oracle for
100 generated signed/unsigned cohort comparisons. Deterministic companion tests
cover integer extrema, non-dyadic means/coverage deltas, absent groups, unavailable
and changed-read states, binding mismatches, detachment and legacy count requests.

`TestQuintNumericComparisonLiveLedger` consumes journey 175's actual OTel/source
ledger through `GOO_NUMERIC_COMPARISON_LEDGER`. Its independent oracle checks
record identity, exact totals/groups, rational arithmetic and coverage, previous
windows, selected/rest partitions, logs/uint64, saved/reusable evidence and
source-free staging/reopen. Twenty-two corruption controls must change the input
and be rejected through qobs/Quint replay. The `numeric-comparisons.ts` Bombadil
actor explores supplied ranking and coverage, stages a fixed question, explicitly
reads once, pins numeric evidence and edits/repairs an invalid binding; its terminal
coverage record proves those actions occurred without additional source reads.

`TestHegelDistributionComparisonExactSharesAndSidePlans` and companion distribution tests
use independently parsed integers, derived exact bin bounds and big.Int/Rat
arithmetic to check all-record denominators, zero/unknown distinctions, matched
side scopes, aggregate-only versus categorical modes and legacy hashes.

`TestQuintDistributionComparisonLiveLedger` consumes journey 176 through
`GOO_DISTRIBUTION_COMPARISON_LEDGER`; it independently checks the actual OTel
records, full histograms, exact shares/coverage, union-cropped display, chosen-side
plans/reads and immutable general/v5/v4 receipts. Twenty-five changed corruption
controls must reject. `TestQuintDistributionComparisonLateLedger` reads
`GOO_DISTRIBUTION_COMPARISON_LATE_LEDGER` and replays the actual ownership sequence
through `model/distribution_lifecycle.yaml`; eleven controls reject hidden reads,
scope substitution, stale plans and lost return behavior. Open stages a population
child; its Run performs the source read. The finite Bombadil actor records all ten
phases and a terminal coverage receipt proving two initial side reads, one explicit
child read, one pure planner and a retained second selection.

`TestQuintPopulationExplanationLiveLedger` consumes journey 177's actual ordinary
OTel fixtures through `GOO_POPULATION_EXPLANATION_LEDGER`. An independent stdlib
and big.Int/big.Rat oracle computes full population/group counts, missing versus
present-empty values, exact share changes and deterministic ranked references.
It checks original scopes, count-only planning, partial high-cardinality fields,
source/read versus custody clocks, representation/capture hashes, pure nullable
side filters, and promotion of the exact held observation. The live fixture makes
eight delayed selected records visible before Save: the saved observation retains
160 selected records while explicit reuse sees 168. A lost acknowledgment and
retry must return the same capture, with no source read. Twenty-one changed
corruptions must reject; mathematical/ranking corruptions are rehashed so the
independent oracle, rather than only checksum detection, must reject them.

`TestQuintPopulationExplanationOwnershipLedger` uses
`GOO_POPULATION_EXPLANATION_OWNERSHIP_LEDGER` to check real completed responses
held across cancellation/navigation, retained return, a fresh explanation and
explicit staged-side execution. Eleven changed controls reject stale ownership,
extra reads/plans, shifted scope, wrong nullable membership and reused tokens.

Run these from `verification` after journey 177 produces the compact files under
`artifacts/population-explanations/<environment>/`:

```sh
GOO_POPULATION_EXPLANATION_LEDGER=/absolute/path/to/ledger.json \
  go test -run '^TestQuintPopulationExplanationLiveLedger$' -count=1 -v .
GOO_POPULATION_EXPLANATION_OWNERSHIP_LEDGER=/absolute/path/to/ownership.json \
  go test -run '^TestQuintPopulationExplanationOwnershipLedger$' -count=1 -v .
```

The `population_explanation` and `population_explanation_ownership` Quint models
replay the finite observed action order and logical API effects: source operations,
pure planning and observed-result promotions. The main trace has eight source
operations, two promotions and five plans; ownership has four source operations
and four plans. These are browser API effects, not counts of internal SQL reads.
Arithmetic and fixture membership are checked by the independent Go oracle before
qobs/quintgo replay. The models do not prove the full evaluator, all concurrent
interleavings, collector completeness or production storage durability. Native
Hegel/backend tests separately exercise generated arithmetic, field failure,
changed reads, byte bounds and snapshot expiry/retry behavior.

Recorded local validation on 2026-10-24:

| Actual journey ledger | Native result | Retained log |
| --- | --- | --- |
| `population-explanation-3107cb81d7/ledger.json` (final compact presentation) | Baseline accepted; all 21 controls rejected, 104.93s | [Main validation](../artifacts/population-explanations/population-explanation-3107cb81d7/native-validation.txt) |
| `population-explanation-da32ba5df3/ownership.json` | Baseline accepted; all 11 controls rejected, 41.46s | [Ownership validation](../artifacts/population-explanations/population-explanation-da32ba5df3/native-validation.txt) |

These environment-specific ledgers/logs are generated local artifacts. Re-run
journey 177 to create new files when reproducing outside this workspace. The
main verifier also passed the prior `population-explanation-bb8dfc9ade` ledger
(87.55s); the final compact presentation retained the same arithmetic and API
effect contract. An initial native assumption that unknown observation tokens
returned HTTP 410 was corrected to the established HTTP 409 promotion contract;
the initial diagnostic log remains beside that earlier ledger.

Activity explanations (journey 178) use the existing explanation engine over
fixed Activity scopes. `TestHegelActivityExplanationScopeRoundTrip` generates
nested nullable string and typed integer predicates for logs/spans, including
integers above binary64 precision. It checks Go JSON round trips, preservation
of original saved scopes, adjacent equal-width previous windows and bounded brush
windows through the existing count planner and parameterized SQL boundary. This
does not duplicate the JavaScript scope helper; its unit tests live with that
helper.

`TestQuintActivityExplanationLiveLedger` is opt-in with
`GOO_ACTIVITY_EXPLANATION_LEDGER`. The independent source-fact oracle checks the
actual held-row chart as a sample, fixed complete-window and brushed queries,
original saved capture scopes, exact counts/shares/rank, immutable capture hashes,
late-arrival versus observed-result promotion, fresh reuse and side-filter
membership. A finite Quint replay checks the observed action order and browser
API effect counts. These are browser request counts, not internal SQL reads,
and neither replay nor complete aggregate results prove transactional snapshots
or complete telemetry. Actual source-ledger validation is required; absent an
environment variable the test skips.

`TestQuintActivityExplanationOwnershipLedger` is opt-in with
`GOO_ACTIVITY_EXPLANATION_OWNERSHIP_LEDGER`. Its observed held-response journey
covers resume, navigation and dirty-scope cancellation, saved-scope GET ownership,
fresh Run and a late side-filter plan. The independent oracle checks four actual
explanation results and the typed filter. The finite model checks the seventeen
actions and API effects; captured DOM snapshots must show no stale field rows or
staged link, and an interrupted saved-scope loader must become usable again. It
is a replay of these observed schedules, not an exhaustive concurrency proof.

These Activity tests require the read-only copied full stored original capture
as `scope-capture.json` beside the ledger. A public capture page contains only
100 retained records and cannot reconstruct the original hash. The oracle hashes
the full stored representation, verifies every retained row against emitted
facts and compares the public prefix structurally. Raw JSON record formatting may
change across browser transport; structural comparison uses `json.Number`, while
custody hashing retains the actual complete stored representation.

`TestActivityExplanationSavedBaselineLedger`, enabled by
`GOO_ACTIVITY_EXPLANATION_SAVED_BASELINE_LEDGER`, checks the supplemental actual
API observation after one late record arrives in the saved baseline's original
window. It verifies the exact original baseline query, the fresh explanation's
new baseline count, unchanged selection count, independent exact field/rank
results, and byte-identical original capture files before/after. This API-only
record has no invented browser lifecycle replay.

From `verification/`, run the relevant entrypoint against its retained artifact
(use the configured Go toolchain and writable Go/Hegel caches):

```sh
GOO_ACTIVITY_EXPLANATION_LEDGER=../artifacts/activity-explanations/activity-explanation-a79a034c19/ledger.json \
  go test -run '^TestQuintActivityExplanationLiveLedger$' -v .
GOO_ACTIVITY_EXPLANATION_LEDGER=../artifacts/activity-explanations/activity-explanation-8bbd9d4587/ledger.json \
  go test -run '^TestQuintActivityExplanationLiveLedger$' -v .
GOO_ACTIVITY_EXPLANATION_OWNERSHIP_LEDGER=../artifacts/activity-explanations/activity-explanation-4b586590f2/ownership.json \
  go test -run '^TestQuintActivityExplanationOwnershipLedger$' -v .
GOO_ACTIVITY_EXPLANATION_SAVED_BASELINE_LEDGER=../artifacts/activity-explanations/activity-explanation-a79a034c19/saved-baseline-late-arrival.json \
  go test -run '^TestActivityExplanationSavedBaselineLedger$' -v .
```

Actual journey 178 acceptance:

| Evidence | Native result | Retained stdout |
| --- | --- | --- |
| [Spans main ledger](../artifacts/activity-explanations/activity-explanation-a79a034c19/ledger.json) | Baseline accepted; all 25 changed controls rejected; 200.18s | [Main log](../artifacts/activity-explanations/activity-explanation-a79a034c19/native-validation.txt) |
| [Logs main ledger](../artifacts/activity-explanations/activity-explanation-8bbd9d4587/ledger.json) | Baseline accepted; all 25 changed controls rejected; 189.83s | [Main log](../artifacts/activity-explanations/activity-explanation-8bbd9d4587/native-validation.txt) |
| [Held ownership ledger](../artifacts/activity-explanations/activity-explanation-4b586590f2/ownership.json) | Baseline accepted; all 16 changed controls rejected; 79.54s | [Ownership log](../artifacts/activity-explanations/activity-explanation-4b586590f2/native-validation.txt) |
| [Saved baseline late arrival](../artifacts/activity-explanations/activity-explanation-a79a034c19/saved-baseline-late-arrival.json) | Baseline accepted; all 8 changed controls rejected; 4.61s | [Supplemental log](../artifacts/activity-explanations/activity-explanation-a79a034c19/native-saved-baseline-validation.txt) |

The generated ledgers and copied original source captures are local artifacts.
Both Quint models typecheck; the generated scope property passes 70 examples.
Main ledgers contain nineteen actions, five explicit source operations, two
promotions, one pure side plan and one stream. Ownership contains seventeen
actions, four explicit source operations, one pure plan and three explicit stream
starts. The supplemental baseline changes from 240 to 241 after an actual late
arrival; the selection remains 248 and the original stored capture bytes stay
identical. No capture page or held sample supplies the fresh explanation totals.

Initial native diagnostics exposed test representation assumptions: a public page
cannot rehash a full retained capture, raw JSON row formatting can change through
browser/held-response transport, and an omitted tail signal means logs. The
completed checks require full stored original captures and exact hashes, compare
public record values structurally with exact numeric decoding, and normalize the
existing logs default. Initial diagnostics are retained beside the ledgers,
including the spans `native-initial-rejection-transcript.txt` and ownership
`native-rawmessage-diagnostic.txt`. Successful stdout retains some raw-format
comparison debug labels; these are not the structural comparison invariants.
Current test diagnostics report only the actual structural/hash guards.

Error triage (journey 179) adds
`TestHegelErrorTriageProjectionAndRules`. Seventy generated examples each cover
twelve ordinary span-event shapes, independent nullable type projection,
repetition/reversal invariance and mismatched-array refusal by the combined
type/state projection. Names-only exception presence remains independently
knowable when attribute alignment is unavailable. All seven supported
span/log rule modes cross the JavaScript template → Go typed query boundary and
are checked against an independent raw-fact oracle, including parent-scope
exclusion, case-sensitive names/text and exact uint64 severity boundaries. A
handled exception event is evidence; the property does not label the span failed
unless a selected rule actually says so. The source adapter's real ClickHouse
parity checks remain separate. Live journey/model acceptance is not claimed by
this generated test.

`TestQuintErrorTriageLiveLedger` is opt-in with `GOO_ERROR_TRIAGE_LEDGER` and
requires an actual `scope-capture.json` copied read-only beside that ledger.
Its independent ordinary-OTLP oracle checks error-rule membership, record-based
counts, nullable service/type groups, every time bucket, event names/attributes/
nanosecond times, trace IDs and the supplied exception inspector details. It
checks exact parent/inspection/within-service/previous scopes, count-comparison
shares and ordering, comparison basis, full stored capture SHA256 and retained
memberships, the public capture prefix, reusable YAML intent and evidence-note
references. Main observed lifecycle replay has an explicit three-read trace
branch for spans; logs do not invent a corresponding trace action.

`TestQuintErrorTriageOwnershipLedger` is opt-in with
`GOO_ERROR_TRIAGE_OWNERSHIP_LEDGER`. It checks stale open attempts after a dirty
scope, applying a new scope, hiding the owner and resuming Activity: before/after
URLs, hidden panels, source-free counters and explicit stream starts. No actual
fixture acceptance is implied when these opt-in tests skip.

For these checks the independent semantic oracle rejects corrupted arithmetic,
source scopes and capture contents before running Quint. Quint replays the valid
source ledger and lifecycle mutations that preserve those source facts. This
avoids repeatedly invoking the model checker on already-invalid arithmetic and
keeps the two forms of evidence distinct. The models cover observed finite action
orders and logical browser API effects, not all schedules, internal SQL read
counts, telemetry completeness or failure/incident causation. Supplemental state
and alternate-rule API reads are recorded separately from browser action counters.

Actual journey 179 acceptance:

| Evidence | Native result | Retained stdout |
| --- | --- | --- |
| [Spans main ledger](../artifacts/error-triage/error-triage-d2ace37ca6/ledger.json) | Baseline accepted; all 33 changed controls rejected; 23.51s | [Spans log](../artifacts/error-triage/error-triage-d2ace37ca6/native-validation.txt) |
| [Logs main ledger](../artifacts/error-triage/error-triage-c0de66f5dc/ledger.json) | Baseline accepted; all 33 changed controls rejected; 20.09s | [Logs log](../artifacts/error-triage/error-triage-c0de66f5dc/native-validation.txt) |
| [Stale owner ledger](../artifacts/error-triage/error-triage-5c76f5e2eb/ownership.json) | Baseline accepted; all 14 changed controls rejected; 10.85s | [Ownership log](../artifacts/error-triage/error-triage-5c76f5e2eb/native-validation.txt) |

Both Quint models typecheck. The final generated projection/rule property passes
70 examples in 4.95s; opt-in entrypoints skip honestly without their artifact
environment variables. Main spans replay contains twenty-two actions and ten
source operations, including three explicit trace reads. Logs contains twenty-one
actions and seven source operations. Both contain one stream and two explicit
publications (reusable question and evidence notes). Ownership replays eight
actions with zero telemetry reads/publications and three explicit stream starts.
The main full-population counts are 180 spans and 260 logs, versus a 100-record
page. Duplicate exception events count as one captured span. The fresh expanded
trace has one captured span and no captured linked logs in this fixture; this is
bounded evidence, not a complete-trace or telemetry-completeness guarantee.

The independent native oracle initially rejected the logs ledger's supplemental
rule reads. The harness indexed a normalized, flattened OR expression: the read
labeled `status` actually selected only severity numbers 17–24 (160), while the
read labeled `type` selected the literal severity `ERROR` (20). The main UI
observation, source capture and comparison already passed. The harness reran only
those supplemental actual API requests using the public rule helpers; the
independent oracle remained unchanged. Correct full-rule counts are 200 for
status evidence and 80 for structured type/message/stack evidence. They overlap,
so the default combined population is 260, not their sum.

The [rejected original ledger](../artifacts/error-triage/error-triage-c0de66f5dc/ledger-rejected-rule-indexing.json),
[initial rejection](../artifacts/error-triage/error-triage-c0de66f5dc/native-validation-rejected-rule-indexing.txt)
and [actual API correction metadata](../artifacts/error-triage/error-triage-c0de66f5dc/rule-read-correction.json)
are retained. Correction metadata records the original/new ledger hashes and
actual query requests; all other main/lifecycle fields remained identical. No
application semantics or native rule/source/hash guard changed to accept it.

From `verification/`, with the configured Go toolchain and writable Go/Hegel
caches:

```sh
GOO_ERROR_TRIAGE_LEDGER=../artifacts/error-triage/error-triage-d2ace37ca6/ledger.json \
  go test -run '^TestQuintErrorTriageLiveLedger$' -v .
GOO_ERROR_TRIAGE_LEDGER=../artifacts/error-triage/error-triage-c0de66f5dc/ledger.json \
  go test -run '^TestQuintErrorTriageLiveLedger$' -v .
GOO_ERROR_TRIAGE_OWNERSHIP_LEDGER=../artifacts/error-triage/error-triage-5c76f5e2eb/ownership.json \
  go test -run '^TestQuintErrorTriageOwnershipLedger$' -v .
```

Decimal numeric verification uses an independent ASCII grammar and `math/big`
coefficient oracle in `decimal_numeric_test.go`. `decimal64` declares scale 1–9
and a signed int64 coefficient. Exporter values and typed thresholds must contain
exactly that many fractional digits with a mandatory dot; plus signs, leading
integer zeros, negative zero, exponents, whitespace and extra precision (including
trailing zeros) are invalid. Summaries, histogram geometry and numeric deltas stay
coefficient integers/rationals on the wire. Human display values divide by the
binding's power of ten without floating conversion. Integer bindings omit scale.

The generated checks cover 90 coverage/partition examples, 70 histogram and
human-threshold range-partition examples, 70 comparison/distribution and strict
scale/legacy-wire examples, and 70 stateful examples with 30 scheduled actions
(2,100 total). The state swarm edits scale and source values and freezes prior
observations; frozen raw facts retain their original binding/scale. Thirty fixed
literal cases cover all nine scales, coefficient extrema/overflow, values above
binary64 precision, signed fractions, negative zero and malformed precision.
Expected parsing/formatting does not call production helpers. Differently scaled
comparison cohorts are each individually valid before compatibility is refused.

`TestQuintDecimalNumericLiveLedger` is opt-in with
`GOO_DECIMAL_NUMERIC_LEDGER`; it requires the actual ledger, full read-only copied
`scope-capture.json` and `capture-custody.json` beside it. The oracle checks unique
raw page identities, exact attributes/resources/time/trace IDs, complete counts,
coverage, grouped/time summaries and all 128 coefficient bins against emitted
OTLP facts. It checks human typed thresholds, selected/rest membership including
excluded parse states, exact comparison means/coverage, human rational DOM titles,
scale 1/9 and integer/resource regressions, saved YAML and fresh reuse. The full
400-record source capture is hashed separately from its 100-record public page;
raw-file SHA256/byte count/receipt/copy clocks are checked alongside the typed
capture identity and immutable reopened evidence.

The generic `numeric_authoring` model replays this observed twenty-one-action
workflow: source-free authoring and invalid-input repair, four coverage filters,
parent/range reads, pure brush/Back/planner actions, explicit comparison reads,
capture/reopen and package reuse. Capture's action contains the separately named
parent reread and capture reread. Independent source/arithmetic corruption is
rejected before Quint; lifecycle mutations preserving source facts are replayed.
These checks do not prove all interaction schedules, evaluator correctness for
all data, internal SQL read counts, original OTLP scalar types or telemetry
completeness. Separate source reads do not establish an atomic snapshot.

Actual journey 180 acceptance:

| Evidence | Native result | Retained stdout |
| --- | --- | --- |
| [Final spans ledger](../artifacts/decimal-numerics/decimal-numerics-6cf7ac4c11/ledger.json) | Baseline accepted; all 38 changed controls rejected; 36.68s | [Spans log](../artifacts/decimal-numerics/decimal-numerics-6cf7ac4c11/native-validation.txt) |
| [Final logs ledger](../artifacts/decimal-numerics/decimal-numerics-630ae6e1cc/ledger.json) | Baseline accepted; all 38 changed controls rejected; 33.20s | [Logs log](../artifacts/decimal-numerics/decimal-numerics-630ae6e1cc/native-validation.txt) |
| [Generated properties and fixed boundaries](../artifacts/decimal-native-properties.txt) | 300 generated examples, including 2,100 scheduled swarm steps; 30 fixed rows; 3.733s | Same linked log |

The generic authoring model typechecks. Both actual traces contain twenty-one
actions, fourteen explicit source operations, two pure selection plans, one
comparison protocol call and one package publication. Each source population has
400 records: 160 valid, 10 missing, 10 present-empty and 220 invalid at scale 2.
The coefficient interval `[128,512)` is authored as human predicates `[1.28,5.12)`
and selects 30 records. Its mean is the exact coefficient rational `983/3`; its
human mean is `983/300 USD`. The UI's unreduced human fraction is accepted only
when it is rationally identical, with an exact matching title. Domain sentinel
geometry remains a coefficient bound and is never emitted as an invalid human
predicate. Scale 1/9, resource-field and legacy integer source observations are
also checked against their original raw values.

Each final ledger has its real full stored source capture and raw custody sidecar:
[spans custody](../artifacts/decimal-numerics/decimal-numerics-6cf7ac4c11/capture-custody.json)
and [logs custody](../artifacts/decimal-numerics/decimal-numerics-630ae6e1cc/capture-custody.json).
The independent guard verifies all 400 retained records and their hashes, not
merely the public 100-record page. Rejection controls target duplicate IDs, raw
attribute/resource mutation, scale metadata, coefficients, bins, human thresholds,
selected/rest membership, rational displays, capture identity/custody, package
reuse, source-free staging/repair/brush/reopen and protocol/publication effects.

An [initial diagnostic](../artifacts/decimal-numerics/decimal-numerics-93cf173cf5/native-validation-rejected-request-effect.txt)
on the earlier accepted logs source fixture passed every source/scope, capture,
comparison and package guard but rejected the request ledger's explicit
`comparison` effect. The verifier had not yet recognized that observer category.
It now requires that entry to be exactly `/api/compare`, match the typed observed
partition request and occur exactly once. Source/read assertions were retained;
no application or arithmetic semantics changed. Final acceptance uses the two
final artifacts above.

From `verification/`, with the configured toolchain and writable Go/Hegel caches:

```sh
go test -run '^(TestHegelDecimal64|TestDecimal64FixedScaleBoundaryContract)' -v .
GOO_DECIMAL_NUMERIC_LEDGER=../artifacts/decimal-numerics/decimal-numerics-6cf7ac4c11/ledger.json \
  go test -run '^TestQuintDecimalNumericLiveLedger$' -v .
GOO_DECIMAL_NUMERIC_LEDGER=../artifacts/decimal-numerics/decimal-numerics-630ae6e1cc/ledger.json \
  go test -run '^TestQuintDecimalNumericLiveLedger$' -v .
```

### Contextual numeric plot scope (181)

`TestHegelContextualNumericPlotScope` invokes the actual client-side
`numeric-plot.plotQuestion` helper through Node for 70 generated cases, then
checks the resulting query through the Go wire/compiler boundary. The independent
checks preserve fixed windows, environment/signal, nested nullable predicates,
exact integer thresholds above 2^53, qualified attribute/resource namespaces and
reported-edge identities with explicit namespace presence. They require the
numeric declaration and powers-of-two histogram while clearing previous grouping,
time buckets, duration/status summaries and pagination offsets. Mutating the
returned predicate and numeric field must leave both original parent and supplied
binding unchanged. This checks authoring and custody, rather than rerunning the
decimal arithmetic campaigns above.

The generated campaign passed in 4.74s; stdout is retained in
[contextual-numeric-plot-properties.txt](../artifacts/contextual-numeric-plot-properties.txt).
Generated requests are not production source-read or browser-lifecycle evidence;
the separate actual journey acceptance is recorded below.

```sh
cd verification
PATH=/workspace/.toolchains/go/bin:$PATH GOPATH=/workspace/.go \
GOCACHE=/workspace/.cache/go XDG_CACHE_HOME=/workspace/.cache GOTOOLCHAIN=local \
go test -run '^TestHegelContextualNumericPlotScope$' -v .
```

`TestQuintContextualNumericPlotLiveLedger` consumes the actual journey ledger,
ordinary raw OTel fixture facts, full stored capture bytes and custody sidecar.
It independently applies the originating SSE batch's fixed half-open window and
original predicate to the fixture, then checks every completed plot's count,
coverage/coefficient summary, all 128 histogram bins, unique inspection-page IDs,
raw strings/timestamps/resources, qualified namespaces and selected/rest partition.
Spans expose both `Attributes` and `SpanAttributes`; both must exactly match the
fixture. Logs expose `Attributes` in their public source DTO. Neither sample
rows nor the fixture corpus size define the all-record measurement population.
The actual original windows contained **387 spans and 382 logs**, from 400 emitted
records per signal, with 27 records in each selected coefficient range.

Final native acceptance passed against those exact ledgers: spans baseline plus
34 guaranteed-changed negative controls in **37.494s**, logs plus 34 in
**43.569s**, ownership plus 14 in **8.277s**, and the two-signal population
supplement plus 14 in **21.532s**. All **96** controls changed the serialized
ledger and were rejected. Controls cover scope/window and namespace drift,
stale presentation measures, duplicate/raw-corrupted inspection rows,
summary/scale/bin corruption, missing excluded rest, altered comparison means,
full stored-capture identity/raw/hash/custody corruption, wrong source endpoints,
and hidden source reads during source-free interactions. Ownership and population
controls also reject stale dialog evidence, changed companion queries and extra
reads when staging or returning.

Each Activity/fixed/discovery/saved origin must preserve its exact parent membership
and fixed window while removing prior presentation measures. The saved origin
remains linked to the original frozen capture question; the newly declared
resource plot runs separately. Full stored capture payloads retain every matching
record, independently validate their typed receipt SHA256, and match the public
inspection-page prefix. Raw file SHA256, byte count, receipt identity and copy/read
clocks are checked against read-only custody sidecars.

The main finite Quint model replays 18 actual actions: opening, declaring,
cancelling, brushing, staging and reopening remain source-free; explicit executions
account for nine source requests, two pure planners, one comparison and one stream
subscription. The capture action contains the separately observed parent rerun and
capture reread. The ownership model replays four actual actions with no fixed
source reads, including navigation, dirty scope and resume invalidating retained
confirmation buttons. `TestQuintContextualNumericPlotPopulationLedger` separately
checks the two-signal population owner: its eight actions preserve both original
queries and companion UI identity while staging a single-signal child from a span
record or observed log resource field. Its four reads are two parent queries and
one per explicitly run child.

These are finite observed action/effect replays, not proofs of all browser schedules,
telemetry completeness, original OTLP scalar types or the evaluator implementation.
Raw-fact arithmetic, membership and hashes are checked independently before replay;
no production evaluator predicts an expected result. The existing decimal parser
property campaigns remain the arithmetic campaign rather than being repeated for
this authoring feature.

Initial control diagnostics are retained. A physical span-attribute mutation first
exposed a verifier blind spot: its generic attribute map was checked while the
redundant physical map was ignored. The oracle was strengthened to check both.
The subsequent logs baseline correctly revealed that the logs public DTO has no
`LogAttributes` field; the guard and mutations now target the actual exposed maps.
Neither correction changed application code, source semantics or fixture results.

- [Spans ledger](../artifacts/contextual-plots/contextual-plot-073c7ceb02/ledger.json),
  [native validation](../artifacts/contextual-plots/contextual-plot-073c7ceb02/native-validation.txt),
  [full capture custody](../artifacts/contextual-plots/contextual-plot-073c7ceb02/capture-custody.json).
- [Logs ledger](../artifacts/contextual-plots/contextual-plot-f339d8d5ac/ledger.json),
  [native validation](../artifacts/contextual-plots/contextual-plot-f339d8d5ac/native-validation.txt),
  [full capture custody](../artifacts/contextual-plots/contextual-plot-f339d8d5ac/capture-custody.json).
- [Ownership ledger](../artifacts/contextual-plots/contextual-plot-334d9637c4/ownership.json)
  and [native validation](../artifacts/contextual-plots/contextual-plot-334d9637c4/native-validation.txt).
- [Two-signal population ledger](../artifacts/contextual-plots/contextual-plot-06a76e3f3d/population.json)
  and [native validation](../artifacts/contextual-plots/contextual-plot-06a76e3f3d/native-validation.txt).
- [Initial span control rejection](../artifacts/contextual-plots/contextual-plot-073c7ceb02/native-validation-rejected-physical-map-control.txt)
  and [logs DTO guard rejection](../artifacts/contextual-plots/contextual-plot-f339d8d5ac/native-validation-rejected-nonexistent-log-physical-map.txt).

Reproduce the opt-in live checks individually by setting
`GOO_CONTEXTUAL_NUMERIC_PLOT_LEDGER`,
`GOO_CONTEXTUAL_NUMERIC_PLOT_OWNERSHIP_LEDGER` or
`GOO_CONTEXTUAL_NUMERIC_PLOT_POPULATION_LEDGER` to the respective actual ledger
above, with the same Go/cache environment as the generated command. Run only the
matching `TestQuintContextualNumericPlotLiveLedger`,
`TestQuintContextualNumericPlotOwnershipLedger` or
`TestQuintContextualNumericPlotPopulationLedger` entrypoint. Without a ledger these
tests explicitly skip rather than claiming live acceptance.

### Versioned log body patterns (182)

`TestHegelLogPatternIndependentProjection` uses a manual UTF-8 byte scanner,
independent of the production regexp/encoder. It checks maximal ASCII digit-run
boundaries, lossless lowercase-hex literal segments, literal-marker collisions,
non-ASCII digits, signed/decimal/UUID punctuation, control characters and Unicode,
256-byte limits and explicit empty/oversize/invalid-UTF8 sentinels. Oversize takes
precedence over invalid UTF-8. Sentinel groups remain explicitly unclassified;
sharing `lp1:o` does not imply that long bodies share a text pattern.

`TestHegelLogPatternCanonicalPredicateBoundary` independently validates whole IDs:
lowercase even-length hex chunks without digit bytes, no empty interior chunks,
and a nonempty UTF-8 representative of at most 256 bytes. It covers malformed,
uppercase, digit-literal, invalid-UTF8, overlong and adjacent-marker forms, logs-only
field validation, exact eq/ne complements and rejection of encoded-substring
filters. The default query retains overall bucketed record counts without the
optional grouped-time-series cap. `TestHegelLogPatternRecordStateSwarm` schedules
append/replace/freeze operations and independently checks all-record grouping and
frozen source-body custody after each step.

The generated campaigns passed in **0.270s**: 80 projection examples, 70 canonical
predicate examples and 70 stateful examples with 20 scheduled steps each,
**220 examples and 1,400 state steps** total. Stdout is retained in
[log-pattern-properties.txt](../artifacts/log-pattern-properties.txt).
The [initial label-contract diagnostic](../artifacts/log-pattern-properties-rejected-label-contract.txt)
records a concurrent display-contract update: raw control/Cf/line-separator
characters now render escaped, while identities and source bodies are unchanged.
The independent label expectation follows that explicit display contract.

Malformed physical UTF-8 is a native/source-adapter concern: OTLP JSON cannot
represent it as an ordinary source string. Generated malformed byte cases do not
claim live UI evidence for `lp1:u`. The separate actual journey, immutable full capture and finite observed-lifecycle
acceptance are recorded below.

```sh
cd verification
PATH=/workspace/.toolchains/go/bin:$PATH GOPATH=/workspace/.go \
GOCACHE=/workspace/.cache/go XDG_CACHE_HOME=/workspace/.cache GOTOOLCHAIN=local \
go test -run '^TestHegelLogPattern' -v .
```

`TestQuintLogPatternLiveLedger` independently filters raw fixture records by the
actual originating fixed Activity window and full namespace/team/service predicate.
It verifies complete pattern-group totals, each padded overall time bucket, unique
raw page identities, unchanged body/attribute/resource/timestamp/trace fields,
optional derived pattern projections, decoded labels and exact all-record share
fractions. The actual original window contained **382 logs across 27 groups**;
the selected text shape contained **128**, and its whole-parent rest contained
**254**, including empty and oversize bodies. Their service distributions differ.
Neither the 100-row inspection page nor source examples define those totals.
Fresh three-minute grammar reuse contained 400 matching logs; the original fixed
capture retained its original scope and counts.

Full stored capture bytes are independently hashed both as the typed receipt
representation and as the copied raw file, with receipt ID, byte count and custody
clocks checked. The 222,507-byte source capture holds every one of its 382 matching
records; its public page is checked as the exact retained prefix. The main finite
model replays 16 actual actions and seven logical source requests, one comparison,
one publication and one stream subscription. Sorting, raw inspection, drill
staging, back navigation, reopening and reuse authoring remain source-free.
Those logical HTTP requests are not a count of internal ClickHouse SQL statements.

`TestQuintLogPatternOwnershipLedger` replays three actions against the real hidden
or dirty fixed-question owner: retained links stay inert after invalidation, with
unchanged URL/stage counts and no additional source reads. The hidden owner remains
connected in the recorded observation. `TestQuintLogPatternCardinalityLedger`
independently derives 1,100 distinct valid patterns, checks the ungrouped ready
probe against the exact same scope and the actual grouped refusal HTTP 502, and
requires the absence of count, rows, groups and time-series payloads. Its two-action
model allows one explicit failed read; no partial top list becomes evidence.

Final native acceptance passed with **65 guaranteed-changed negative controls**:
main baseline plus 41 controls in **40.885s**, ownership plus 12 in **11.270s**,
and cardinality plus 12 in **6.712s**. Every control changed serialized evidence
and was rejected. They cover raw-body/identity/group/time corruption, duplicate
rows/groups, rounded or wrong-denominator shares, scope drift, lost unclassified
rest, altered service comparison, capture/hash/custody mutation, stale owners,
extra source effects, and partial high-cardinality payloads.

The finite models cover the observed action sequences and effects; they are not
proofs of every browser schedule, all source telemetry, or the evaluator. Raw-fact
normalization, membership, counts, fractions and hashes are checked independently
before replay. The first native display guard retained the obsolete UI state name
`oversized`; the deployed UI matches the Go contract `oversize`. That diagnostic is
preserved, and the final oracle uses the agreed state without changing application
code, source facts or semantic expectations.

- [Main ledger](../artifacts/log-patterns/log-patterns-7e40a8e9f4/ledger.json),
  [native validation](../artifacts/log-patterns/log-patterns-7e40a8e9f4/native-validation.txt)
  and [full capture custody](../artifacts/log-patterns/log-patterns-7e40a8e9f4/capture-custody.json).
- [Ownership ledger](../artifacts/log-patterns/log-patterns-7e3f5ef9fc/ownership.json)
  and [native validation](../artifacts/log-patterns/log-patterns-7e3f5ef9fc/native-validation.txt).
- [High-cardinality refusal ledger](../artifacts/log-patterns/log-patterns-bccde52ccd/cardinality.json)
  and [native validation](../artifacts/log-patterns/log-patterns-bccde52ccd/native-validation.txt).
- [Initial display-state diagnostic](../artifacts/log-patterns/log-patterns-7e40a8e9f4/native-validation-rejected-ui-state-alias.txt).

Reproduce the native checks individually with `GOO_LOG_PATTERN_LEDGER`,
`GOO_LOG_PATTERN_OWNERSHIP_LEDGER` or `GOO_LOG_PATTERN_CARDINALITY_LEDGER` pointing
to its actual ledger, using the Go/cache environment above and the corresponding
`TestQuintLogPatternLiveLedger`, `TestQuintLogPatternOwnershipLedger` or
`TestQuintLogPatternCardinalityLedger` entrypoint. Without its actual ledger each
entrypoint explicitly skips.

### Log-pattern comparisons across time (183)

`TestHegelLogPatternComparisonCanonicalBoundary` checks the narrow existing count
comparison protocol for 70 generated valid pairs, including known-empty and
changed-read states. Malformed pattern IDs are injected into either side with
positive group counts equal to that side's population count; the rejection must
specifically identify `body_pattern_v1`. This prevents unrelated count/state
errors from masking a missing canonical-ID guard, including when the untouched
opposite side is empty or changed. Legacy service distributions still accept
arbitrary exact string values. Adaptive bucket widths keep generated scopes from
1 ms through 120 minutes valid.

`TestHegelLogPatternComparisonExactClientModel` invokes the actual client module
through Node for 80 generated pairs and all three metric orderings. Independent
standard-library `big.Int`/`big.Rat` arithmetic checks the full identity union,
counts, per-second rates, percentage shares, current-minus-baseline differences
and exact ordering. Cases exceed 2^53, use unequal-duration windows with equal
rates, include an empty side with known zero rates and undefined shares, and
include changed-read accounting mismatches. Unknown evidence preserves reported
counts but cannot infer missing-side zeros, rates, shares, differences or absence.
The property also checks immutable input/output snapshots, exact adjacent previous
windows and the original questions retained by pure planning.

The generated campaigns passed in **17.416s**: **150 examples, 240 metric
projections and 640 malformed response-binding controls**. Those controls reject
mismatched source bases, counts, states, result rows, environments, windows,
canonical identities and distribution fields. Stdout is retained in
[log-pattern-comparison-properties.txt](../artifacts/log-pattern-comparison-properties.txt).
These generated cases are pure API/model evidence. Separate actual source reads,
capture custody and finite observed UI lifecycle acceptance are recorded below.

```sh
cd verification
PATH=/workspace/.toolchains/go/bin:$PATH GOPATH=/workspace/.go \
GOMODCACHE=/tmp/goo-module-cache GOCACHE=/workspace/.cache/go-build \
XDG_CACHE_HOME=/workspace/.cache GOTOOLCHAIN=local GOMAXPROCS=2 \
go test -run '^TestHegelLogPatternComparison' -v .
```

`TestLogPatternComparisonFullTwoThousandUnion` separately checks the actual API
and client model with 1,000 distinct groups on each side and a 2,000-identity union.
The independent exact-rate/order oracle requires every identity rather than a
combined 1,000-group truncation. This deterministic boundary passed in **0.862s**;
stdout is retained in
[log-pattern-comparison-union.txt](../artifacts/log-pattern-comparison-union.txt).
The property campaigns above used the legacy Go-cache profile; the reproduction
profile shown here follows the environment's current module/build-cache paths.

The actual across-time acceptance passed independently against ordinary raw log
fixtures. `TestQuintLogPatternTimeLiveLedger` derives each side's complete pattern
set separately from its exact fixed selector and half-open event-time window,
then verifies the full identity union, raw source pages, group/time totals, exact
counts/rates/shares/deltas, all three metric orderings, DOM exact attributes and
shared visual scales. The common shape has 320 records over the saved 120-second
scope and 160 over the current 60-second scope: both rates are exactly **8/3 logs/s**,
while count bars have a 2:1 ratio. Exact values and titles remain rational strings;
CSS bar widths independently follow the display contract's floor to hundredths
of a percentage point. Shares use a fixed shared 0–100% axis.

The original saved capture deliberately groups by service. Loading its scope
preserves its original selector/window, then replaces presentation with the same
pattern profile for explicit fresh reads. Its original **400** source records and
all 191,983 stored bytes remain unchanged after 12 late baseline-only logs become
visible. The actual fresh baseline contains **412**, with current **400**. Both
source-read clocks follow the actual ready probe, while linked source-capture
clocks establish the comparison capture's later fresh reads. Native checks retain
the complete original capture's typed receipt hash, raw file SHA256, receipt ID,
400-record membership, public-page prefix and before/after byte equality. Fresh
comparison payloads are complete and independently hash-validated; paginated
linked source pages are validated as pages rather than falsely rehashed as full
captures. Reusable configuration retains exact questions rather than observed
counts.

`TestQuintLogPatternTimeEmptyLedger` verifies a real zero-record preceding gap:
its rates and counts are known zero, shares and share differences are undefined,
and the complete other-side group union remains visible despite the API's empty
`no_data` row array. `TestQuintLogPatternTimeOwnershipLedger` verifies two real
upstream successful reads whose deliveries are respectively aborted or held across
navigation. Failed delivery produces no comparison protocol, comparison section
or absence labels. A held response cannot publish after return; explicit recovery
rereads both scopes. A retained hidden-but-connected side link cannot navigate,
stage another question or read source data.

The finite models replay **20 main, 3 empty and 7 ownership actions**. The main
trace accounts for nine logical source API operations, three pure comparison
calls, one configuration write and six storage-only GETs. Storage GETs include the
saved scope, freshly captured side pages, and reopened comparison/side pages.
Staging, metric changes, side/parent return, scope loading and reopening do not
add source reads. These are observed API/effect sequences, not internal SQL counts
or proofs of every browser schedule. Independent raw-fact arithmetic, scope,
custody and presentation checks run before replay. Generated changed-read cases
exercise unknown semantics; the actual failure journey records transport loss
rather than manufacturing inconsistent source telemetry.

All **93 guaranteed-changed serialized controls** were rejected:

- Main baseline plus 55 controls: **40.999s**.
- Known-empty baseline plus 18 controls: **9.250s**.
- Failed/held ownership baseline plus 20 controls: **9.897s**.

Controls cover unequal-window denominator errors, count/rate confusion, lost
negative signs, missing union/unclassified groups, incorrect absence or
unavailability, exact source scopes, late data beyond an inspection page,
capture-byte/hash/custody mutation, stale read clocks, configuration drift,
independent side scales, clipped/rounded exact labels and hidden source effects.

Initial observer diagnostics are retained. The old custody helper assumed a
literal `docker cp` description; this fixture records the actual read-only query
container capture-store transport instead. The 183 checker validates that actual
description together with all original bytes/hashes/membership/clocks. The first
visual oracle also assumed unquantized CSS widths and a fitted percentage domain;
it now independently calculates the documented hundredth-percent floor and fixed
0–100% share domain. Exact rational model/data/title checks were unchanged. No
application code or raw fixture evidence was modified to pass these checks.

- [Main ledger](../artifacts/log-pattern-time/log-pattern-time-d9bcc1b3fe/ledger.json),
  [native validation](../artifacts/log-pattern-time/log-pattern-time-d9bcc1b3fe/native-validation.txt)
  and [original capture custody](../artifacts/log-pattern-time/log-pattern-time-d9bcc1b3fe/capture-custody.json).
- [Known-empty ledger](../artifacts/log-pattern-time/log-pattern-time-568c93a96c/empty.json)
  and [native validation](../artifacts/log-pattern-time/log-pattern-time-568c93a96c/native-validation.txt).
- [Failed/held ownership ledger](../artifacts/log-pattern-time/log-pattern-time-519626dc3b/ownership.json)
  and [native validation](../artifacts/log-pattern-time/log-pattern-time-519626dc3b/native-validation.txt).
- [Initial observer rejection](../artifacts/log-pattern-time/log-pattern-time-d9bcc1b3fe/native-validation-rejected-custody-frozen-assumptions.txt)
  and [isolated diagnostic](../artifacts/log-pattern-time/log-pattern-time-d9bcc1b3fe/native-validation-diagnostic.txt).

Reproduce one actual checker at a time using the tested module/build-cache profile:

```sh
cd verification
PATH=/workspace/.toolchains/go/bin:$PATH GOPATH=/workspace/.go \
GOMODCACHE=/tmp/goo-module-cache GOCACHE=/workspace/.cache/go-build \
XDG_CACHE_HOME=/workspace/.cache GOTOOLCHAIN=local GOMAXPROCS=2 \
GOO_LOG_PATTERN_TIME_LEDGER=../artifacts/log-pattern-time/log-pattern-time-d9bcc1b3fe/ledger.json \
go test -run '^TestQuintLogPatternTimeLiveLedger$' -v .
```

Use `GOO_LOG_PATTERN_TIME_EMPTY_LEDGER` with
`TestQuintLogPatternTimeEmptyLedger`, or
`GOO_LOG_PATTERN_TIME_OWNERSHIP_LEDGER` with
`TestQuintLogPatternTimeOwnershipLedger`, for the respective receipts above.
Without the appropriate actual ledger each entrypoint explicitly skips.

### Trace-page latency breakdown (184)

`trace_page_breakdown_test.go` and `trace_page_oracle_test.go` check supplied-page interval coverage independently of the normative merging implementation. The endpoint-cell partition oracle is cross-checked against unit-grid occupancy on small integer intervals. Eighty generated pages exercise overlap, zero-length intervals, translation beyond binary64 precision, safe trillion-fold scaling, input immutability and permutation (240 transformed pages plus 240 permutations). Seventy further generated cases compare the whole normative Go/actual-Node DTO and the independent graph oracle, including duplicates, cycles, missing parents, outside intervals, overflow, malformed fields and nullable versus empty group labels. Eight deterministic authority boundaries include links that must never create child edges; a separate group-sum case exceeds uint64 without wrapping. The expanded campaign passed **4.592s**; [actual stdout](../artifacts/trace-page-breakdown-properties.txt).

The actual-ledger native checks independently validate source membership, raw physical timestamps/durations/parent IDs/attribute namespaces, all page-local interval measures, exact reason sets, group indexes and known/unknown counts. Full-query inclusive duration/status aggregates remain separate from page-only coverage. DOM checks bind exact values, shared group maxima, knownness, exact titles, independently floored 0.01% CSS widths, equal physical track widths and unclipped desktop/mobile labels. The physically preserved duplicate IDs, cycles, outside-parent timing and the 126-record query split into pages of 100 and 26 all passed. A child's unavailable residual does not poison an otherwise valid ancestor edge; a parent outside the supplied page does not invalidate the row's own child coverage.

All three actual ledgers passed, with **80 guaranteed-changed rejecting controls**. Each control checks that serialized evidence changed before requiring rejection. Six controls are also replayed through Quint: an unexpected source effect and wrong action order for each lifecycle.

| Check | Actual artifact | Rejected controls | Result |
| --- | --- | ---: | --- |
| `TestQuintTracePageBreakdownLiveLedger` | [main ledger](../artifacts/trace-page-breakdown/trace-breakdown-36054c3400/ledger.json), [stdout](../artifacts/trace-page-breakdown/trace-breakdown-36054c3400/native-validation.txt) | 38 | PASS 7.415s |
| `TestQuintTracePageGraphsLedger` | [graph/page ledger](../artifacts/trace-page-breakdown/trace-breakdown-382ef74648/graphs.json), [stdout](../artifacts/trace-page-breakdown/trace-breakdown-382ef74648/native-validation.txt) | 26 | PASS 6.927s |
| `TestQuintTracePageOwnershipLedger` | [ownership ledger](../artifacts/trace-page-breakdown/trace-breakdown-96fed1dbe0/ownership.json), [stdout](../artifacts/trace-page-breakdown/trace-breakdown-96fed1dbe0/native-validation.txt) | 16 | PASS 5.903s |

The three qobs/Quint model pairs have **7 / 6 / 3 observed actions**. Main effects are search 1, trace expansion plus its existing linked-log read 3, source-free group highlight/clear 3, selected-row linked-log read 4, explicit fresh capture 5, and source-free reopen 5. The six graph cases reach 20 logical source operations, including the second trace page and its existing linked-log read. Ownership's three actions retain the three-read setup while closed/hidden controls remain inert and a current group highlight resumes without a source read. Raw event names map to `observe_`-prefixed model actions. These finite replays check recorded owner/source discipline, not every possible application schedule or internal SQL request count.

The main capture retains all five actual records. Its complete stored representation is **8,651 bytes**, raw SHA256 `f5baae3da9a4d968dde7fcc2c7ab05c7be4e47fd42f981d78c5665edb82dc452`, receipt `de613dc3648092b7bd4a1ad1e7379506a34255df66d75771fe190c7c3e5ad64a`. Before/after bytes are identical. The native check hashes the complete typed representation from the stored bytes, validates raw SHA/length, original question/membership and custody-after-source clocks, then checks the saved/reopened public payload against it semantically.

Initial observer rejections are preserved in [the first diagnostic](../artifacts/trace-page-breakdown/trace-breakdown-36054c3400/native-validation-rejected-initial-observer.txt), [isolated guards](../artifacts/trace-page-breakdown/trace-breakdown-36054c3400/native-validation-diagnostic.txt) and [capture investigation](../artifacts/trace-page-breakdown/trace-breakdown-36054c3400/native-validation-diagnostic-details.txt). The existing waterfall enriches raw rows with a non-authoritative `duration_ms` Number alias. The final oracle validates exactly that derived conversion, removes only that alias for comparison, and still checks every authoritative raw field. Browser JSON roundtripping also removes ClickHouse's raw-row slash escape in `GET \/checkout`; public semantic equality therefore does not reproduce the original raw-message hash. The complete stored typed digest does reproduce the receipt. Neither correction changes application code or fixture evidence, and no full-capture hash guard was dropped. Temporary diagnostic output helpers were removed.

Reproduce serially after browser/actor ownership is released (each live entrypoint skips if its artifact environment variable is absent):

```sh
cd /workspace/goo/verification
export PATH=/workspace/.toolchains/go/bin:$PATH GOPATH=/workspace/.go
export GOMODCACHE=/tmp/goo-module-cache GOCACHE=/workspace/.cache/go-build
export XDG_CACHE_HOME=/workspace/.cache GOTOOLCHAIN=local GOMAXPROCS=2
go test -run '^Test(HegelTracePage|TracePageBreakdownAuthority)' -v .
GOO_TRACE_PAGE_LEDGER=/workspace/goo/artifacts/trace-page-breakdown/trace-breakdown-36054c3400/ledger.json go test -run '^TestQuintTracePageBreakdownLiveLedger$' -v .
GOO_TRACE_PAGE_GRAPHS_LEDGER=/workspace/goo/artifacts/trace-page-breakdown/trace-breakdown-382ef74648/graphs.json go test -run '^TestQuintTracePageGraphsLedger$' -v .
GOO_TRACE_PAGE_OWNERSHIP_LEDGER=/workspace/goo/artifacts/trace-page-breakdown/trace-breakdown-96fed1dbe0/ownership.json go test -run '^TestQuintTracePageOwnershipLedger$' -v .
```

Observed coverage is not CPU/self time, complete exclusive time or a critical path; trace completeness remains unverified. Inclusive group sums may overlap in time. No new endpoint, query, persistence format or source completeness claim is introduced by this verification.

### Supplied trace-page comparisons (185)

`trace_comparison_oracle_test.go` derives both sides from raw facts with the independent endpoint-cell/graph oracle, builds the exact nullable service/operation identity union, and computes signed integer differences and relative starts with `math/big`. It never uses `CompareTracePages` to produce expected values. Repeated labels are never zipped or deduplicated into invocation matches. A unique-label candidate additionally requires reported labels, a single valid reported trace on each page, unambiguous valid row identities, valid row shape and equal nonempty reported span kinds. Missing groups do not become zero; incomplete metric sums withhold their differences independently of other known measures.

The pure campaign in `trace_comparison_test.go` passed **220 generated examples and eight fixed authority/knownness boundaries in 6.166s**: [actual stdout](../artifacts/trace-comparison-properties.txt). Eighty examples check the raw oracle, side swap and page permutation; seventy compare actual Node/Go whole DTOs against the oracle, including odd-coefficient scaling that exercises signed duration/relative-start values beyond binary64 precision; seventy test simultaneous page translations and exact positive affine scaling. Fixed cases distinguish known zero, missing labels, present empty labels, absent groups, repeated labels, kind mismatch/unreported kind and unknown timing with known inclusive duration. The initial smaller-delta campaign also passed and is retained as `artifacts/trace-comparison-properties-initial-pass.txt`; the final generator deliberately strengthens large-delta coverage. This stateless projection uses generated metamorphisms rather than an artificial evaluator swarm; actual owner state transitions are checked separately below.

The three real ordinary-OTLP ledgers passed on their first native runs, with **80 guaranteed-changed rejecting controls**, including six traces rejected by Quint itself. Each corruption is checked for a serialized change before rejection. No application code or raw fixture was changed to pass these native checks.

| Check | Actual artifact | Rejected controls | Result |
| --- | --- | ---: | --- |
| `TestQuintTraceComparisonLiveLedger` | [main ledger](../artifacts/trace-comparison/trace-comparison-618c4bebb6/ledger.json), [stdout](../artifacts/trace-comparison/trace-comparison-618c4bebb6/native-validation.txt) | 38 | PASS 7.639s |
| `TestQuintTraceComparisonPartialLedger` | [partial/problem ledger](../artifacts/trace-comparison/trace-comparison-ae88eb2a0c/partial.json), [stdout](../artifacts/trace-comparison/trace-comparison-ae88eb2a0c/native-validation.txt) | 22 | PASS 22.960s |
| `TestQuintTraceComparisonOwnershipLedger` | [owner ledger](../artifacts/trace-comparison/trace-comparison-82a9feba5a/ownership.json), [stdout](../artifacts/trace-comparison/trace-comparison-82a9feba5a/native-validation.txt) | 20 | PASS 9.414s |

`trace_comparison_live_test.go` independently validates exact raw timestamp/duration strings, trace/span/parent IDs, reported span kinds, namespace-qualified attribute/resource maps, scope, physical row membership and full-query inclusive duration/status summaries. The main reference and current queries retain their distinct 6-second and 60-second windows. Their query widths do not normalize span-duration deltas or set the display axis: both row placements use the maximum observed page extent, derived from reported intervals. Exact card indexes, matching states, nanosecond titles, shared physical tracks and independently floored 0.0001% placement geometry bind the displayed comparison to those pages. Qualified field differences retain raw typed state/value metadata, distinguish present empty from missing, and keep identical keys in attribute/resource namespaces separate.

The main fixture verifies root inclusive duration 100→130ms and time outside observed direct children 50→75ms; the lookup's signed difference is -10ms, repeated prepare rows remain ambiguous, dispatch kinds do not pair, absent groups do not acquire zero durations, and the marker retains a real zero. The problem fixture withholds the outliving child's parent's residual difference while retaining independently known child measures. The partial query retains full count 126 and its independent aggregate sums while the compared page contains 100 rows, including 99 fanout rows. No aggregate sum is substituted for that supplied page.

The existing reference capture was created by an explicit retained setup request before browser reads; it is not a newly persisted comparison. The complete stored capture is **12,477 bytes**, raw SHA256 `c5d4a6ab3be045df0ed6338f171b3da8fc5127d0c2df5e06094ee5055ce16c0c`, receipt `a3fe17e9079d11802fe0b62847e1a5b86126e52d70e35df5bfc64b6aaec35b0d`. Native checks verify identical before/after bytes, the complete stored typed receipt digest, original question and seven-record membership, custody-after-source chronology and semantic saved/reopened equality. The stored representation remains the hash authority; browser-reserialized raw rows are not substituted for it.

The three qobs/Quint model pairs have **8 / 3 / 8 observed actions**. The main browser watch records six existing Explore source operations: search plus expansion/linked-log reads for each side. Pin, local comparison, pair-attribute inspection and capture reopen add none. The three partial/problem actions reach nine operations. The owner trace begins with six operations, retains the hidden comparison without letting its controls mutate it, requires explicit re-expansion after return (eight), holds a real next-page response (twelve), and rejects its acceptance after a natural environment-draft invalidation. Explicit fresh search plus expansion/linked logs recovers at fifteen. Native guards check the passive DOM states before replay: route-hidden retained controls may remain enabled in the DOM but are inert; pending/stale authority clears the comparison and disables pin/compare. Each model separately rejects a changed source counter and wrong action order.

Reproduce serially after browser/actor ownership is released (live tests skip when their artifact variable is absent):

```sh
cd /workspace/goo/verification
export PATH=/workspace/.toolchains/go/bin:$PATH GOPATH=/workspace/.go
export GOMODCACHE=/tmp/goo-module-cache GOCACHE=/workspace/.cache/go-build
export XDG_CACHE_HOME=/workspace/.cache GOTOOLCHAIN=local GOMAXPROCS=2
go test -run '^Test(HegelTraceComparison|TraceComparisonKnownness)' -v .
GOO_TRACE_COMPARISON_LEDGER=/workspace/goo/artifacts/trace-comparison/trace-comparison-618c4bebb6/ledger.json go test -run '^TestQuintTraceComparisonLiveLedger$' -v .
GOO_TRACE_COMPARISON_PARTIAL_LEDGER=/workspace/goo/artifacts/trace-comparison/trace-comparison-ae88eb2a0c/partial.json go test -run '^TestQuintTraceComparisonPartialLedger$' -v .
GOO_TRACE_COMPARISON_OWNERSHIP_LEDGER=/workspace/goo/artifacts/trace-comparison/trace-comparison-82a9feba5a/ownership.json go test -run '^TestQuintTraceComparisonOwnershipLedger$' -v .
```

The models replay these finite observed owner/source sequences; they do not prove all schedules, internal SQL request counts or telemetry completeness. Label alignment does not establish shared invocation identity, complete traces, CPU/self time or a critical path. Group duration sums may overlap. No endpoint or comparison persistence format is added.

### Trace-wide span-condition search (186)

`trace_search_oracle_test.go` interprets raw logical fields independently, builds an ID set for each condition and intersects those sets within the explicit environment and half-open fixed window. It derives exact matching-record counts and deterministic representative witnesses afterward, including namespace-qualified raw `fieldValues` and missing/null versus present-empty distinctions. It does not use `Match`, `ReferenceTraceSearch` or the SQL compiler as an expected-value oracle. One observed span may satisfy several independent conditions; repeated records count as records, while reported trace IDs are distinct. Parent filters belong only to the first seeded condition; added conditions may match other services.

The pure campaign passed **300 generated examples in 6.115s**: [actual stdout](../artifacts/trace-search-properties.txt). Ninety examples check full-population intersection, condition/record permutation, bounded prefixes and monotonic added matching records. Seventy numeric examples exercise ten operators each (700 operator scenarios) over exact int64/decimal64 strings, including invalid/empty/missing parsing and values beyond binary64 precision. Seventy actual-Node contextual requests validate through Go without parent mutation or filter leakage into added conditions. Seventy typed compiler compositions verify scope, map keys and string values stay in parameters. Fixed cases distinguish cross-trace false joins from allowed same-span matches, duplicate-record counts, and qualified missing-versus-empty witness ties. Stateless set/predicate metamorphisms are used here; actual owner state is checked separately below.

The [initial rejection](../artifacts/trace-search-properties-rejected-first.txt) and [isolated diagnostic](../artifacts/trace-search-diagnostic.txt) are retained: the test generator initially declared unsupported `Parse: canonical-integer`. The existing API's `canonical-decimal` profile also names base-10 integer parsing. Correcting that test declaration preserved raw values and expected arithmetic; no production change or relaxed membership assertion was needed. Temporary diagnostic source was removed.

Both final ordinary-OTLP native baselines passed on their first actual runs, with **80 guaranteed-changed rejecting controls**, including four corrupted traces rejected by Quint itself. Each control checks that serialized evidence changed before requiring rejection; no application or raw fixture changed to pass these gates.

| Check | Actual artifact | Rejected controls | Result |
| --- | --- | ---: | --- |
| `TestQuintTraceSearchLiveLedger` | [main ledger](../artifacts/trace-search/trace-search-fb788e547b/ledger.json), [stdout](../artifacts/trace-search/trace-search-fb788e547b/native-validation.txt) | 62 | PASS 8.396s |
| `TestQuintTraceSearchOwnershipLedger` | [owner ledger](../artifacts/trace-search/trace-search-84a5306a6a/ownership.json), [stdout](../artifacts/trace-search/trace-search-84a5306a6a/native-validation.txt) | 18 | PASS 7.905s |

`trace_search_live_test.go` independently derives membership, exact full totals, sorted bounded ID prefixes, per-condition matching-record counts, deterministic witness metadata and every qualified raw proof value from fixture records. Setup source probes prove **294 records / 20 trace IDs** actually exist: common environment over the broader window has **291 / 19**, and the other environment has **3 / 1**. The requested fixed common scope has **288 / 18** before predicate intersection. This proves the excluded environment/time controls are present; it does not conflate broad setup coverage with the requested question. Native checks validate the full grouped probe populations and retained raw prefixes.

The parent predicate matches 135 records while its inspection page contains 100. Trace search returns 15 matching reported IDs; an independent API request with limit 2 still reports full total 15 and `hasMore`, and a separate impossible second condition produces known `no_data`. The cross-service trace uses different representative spans; the same-span trace legitimately uses one span for both conditions. The repeated trace has 120 matching records for each condition, independent of the parent or trace inspection page. Its explicit continuation retains full trace count 241 across two actual pages of 100 records; the second-condition representative is absent from the first page and present in the second. The native oracle checks raw identities, timestamps, durations, statuses, kinds, maps and the separate full-query aggregates. It also validates the subsequent source-free trace-page comparison with the independent interval/delta oracle.

DOM checks bind displayed IDs, witness condition indexes/counts, exact timestamp/duration titles and typed qualified values to the actual source result. The original parent predicate is required in condition 1 with the selected service/operation seed; condition 2 remains independent. Trace continuations preserve the common environment and fixed bounds but open the selected trace across services, rather than treating representative witnesses as a trace page. No new capture or comparison receipt is created by this flow; the separate legacy custody regression is not presented as a newly saved trace-search result.

The two qobs/Quint models replay **10 / 8 observed actions**. Main browser effects reach 10 source operations: parent Explore, one trace-search Run, three explicit trace expansions with existing linked-log reads, and a later trace page with its linked-log read. Staging, witness disclosure, pinning and comparison add none. Two separately recorded API trace-search probes (bounded/empty) and setup readiness/presence reads are outside that browser counter. The owner model reaches five source operations, proves a closed editor cannot alter the next explicit request, and rejects closed/hidden retained actions. A real completed backend response is held, a visible condition edit supersedes it, and only a fresh Run publishes rows. Query cancellation is recorded as a control request, not another telemetry source read. Both models separately reject an unexpected source count and wrong action order.

Reproduce serially after browser/actor ownership is released (live tests skip without their explicit artifact variable):

```sh
cd /workspace/goo/verification
export PATH=/workspace/.toolchains/go/bin:$PATH GOPATH=/workspace/.go
export GOMODCACHE=/tmp/goo-module-cache GOCACHE=/workspace/.cache/go-build
export XDG_CACHE_HOME=/workspace/.cache GOTOOLCHAIN=local GOMAXPROCS=2
go test -run '^Test(HegelTraceSearch|TraceSearchSameSpan|TraceSearchQualifiedWitness)' -v .
GOO_TRACE_SEARCH_LEDGER=/workspace/goo/artifacts/trace-search/trace-search-fb788e547b/ledger.json go test -run '^TestQuintTraceSearchLiveLedger$' -v .
GOO_TRACE_SEARCH_OWNERSHIP_LEDGER=/workspace/goo/artifacts/trace-search/trace-search-84a5306a6a/ownership.json go test -run '^TestQuintTraceSearchOwnershipLedger$' -v .
```

These finite traces validate observed client source/owner discipline, not every schedule or internal SQL request count. Representative witnesses and distinct reported IDs do not establish unique invocations, distinct spans per condition, parent topology, ordering, causality, complete traces, complete ingestion or absence of late data. No live malformed/all-zero OTLP identity ingestion is claimed; identity exclusion is covered by generated logical records and separate backend checks.

## Incoming HTTP dashboard slice (C2a.1)

`TestHegelHTTPDashboardPopulationAndIsolation` uses 24 generated scope/literal
examples with exhaustive namespace, span-kind and method-presence combinations.
Its small independent predicate interpreter checks each actual HTTP/log widget
population. `TestHTTPDashboardNativePriorityAndCandidateLimits` checks exact
native priority, independent units/grain, mutable metadata immunity and explicit
empty selection; `TestHTTPDashboardFallbackAndUnknowns` covers six hint/key/signal
boundaries and invalid explicit choices. These focused pure checks passed in
0.036s (`artifacts/http-dashboard-properties.txt`). No generated lifecycle model
is needed for this pure planner.

The opt-in `TestHTTPDashboardObservedPopulations` independently derives matching
records from ordinary OTLP facts and actual returned typed questions. It checks
full counts/groups (including missing versus empty routes), duration sums,
recorded span status, page uniqueness and raw attribute/scope/time fidelity.
Against journey187's accepted `http-dashboard-13dc64d238/ledger.json`, all 19
applicable source responses passed in 0.055s; a guaranteed changed count was
rejected for every response. The adjacent `native-validation.txt` retains stdout.
Browser intent/stale-owner effects and native measurement semantics remain covered
by their respective journey/kernel checks; this test does not model schedules,
prove schema conformance or infer requests from span-record counts.

With the toolchain/cache environment above, reproduce after browser ownership is
released:

```sh
go test -run '^Test(HegelHTTPDashboard|HTTPDashboardFallback|HTTPDashboardNative)' -v .
GOO_HTTP_DASHBOARD_LEDGER=/workspace/goo/artifacts/http-server-dashboard/http-dashboard-13dc64d238/ledger.json go test -run '^TestHTTPDashboardObservedPopulations$' -v .
```

The live test skips unless its artifact variable is supplied. Artifact paths are
local acceptance receipts, not required repository inputs or CI fixtures.

## Legacy incoming RPC dashboard slice (C2a.2)

`TestHegelRPCDashboardRecordPopulation` uses 16 generated exhaustive grids across
service/resource namespace, reported kind, missing/empty/custom system and method
values. It reuses the independent predicate interpreter to check every actual
RPC/log record question without calling the production matcher. Six candidate
boundaries distinguish flat keys from positive hints and reject the newer
`rpc.system.name` alias/resource-key substitutions. Deterministic checks retain
HTTP automatic priority, explicit RPC intent and exact native millisecond/scope
identity under discovery permutation and changed record counts. The focused
campaign passed in 0.075s (`artifacts/rpc-dashboard-properties.txt`). The first
invocation lacked writable `XDG_CACHE_HOME`; its infrastructure diagnostic is
retained separately, with no semantic assertion failure.

`TestRPCDashboardObservedPopulations` is opt-in via `GOO_RPC_DASHBOARD_LEDGER` and
reuses the ordinary raw-fact/full-population oracle from HTTP. It checks the
actual returned questions, full groups/counts/status/durations, raw page identity
and scope, and rejects guaranteed changed counts. Actual journey188 `rpc-dashboard-8bcb4af8ed/ledger.json` passed in 0.091s: all
18 applicable responses, including the 120-record cross-system method population
and its actual trace continuation, passed with changed-count rejection. The
adjacent `native-validation.txt` retains stdout. An initial observer diagnostic
identified omitted trace/span identity fields in the shared raw-fact map; adding
those raw fields corrected the oracle without changing fixtures or application
semantics. The rejected log is retained as
`native-validation-missing-oracle-identity.txt`. The shared-oracle refactor also
passed accepted journey187's 19 responses plus the explicit 188 skip/compile gate
in 0.083s (`artifacts/rpc-dashboard-shared-oracle-regression.txt`). No new formal
lifecycle model or fixed campaign quota is introduced.

After browser ownership is released, use the environment above:

```sh
go test -run '^Test(HegelRPCDashboard|RPCDashboardCandidate|RPCDashboardNative)' -v .
GOO_RPC_DASHBOARD_LEDGER=/workspace/goo/artifacts/rpc-server-dashboard/rpc-dashboard-8bcb4af8ed/ledger.json go test -run '^TestRPCDashboardObservedPopulations$' -v .
```

## Current database client dashboard slice (C2b)

`TestHegelDBDashboardPopulation` uses 16 generated exhaustive service/resource
namespace, reported kind and missing/empty/custom current-system grids, with
legacy `db.system` supplied independently as an exclusion control. The shared
independent predicate interpreter checks all actual DB/log widget questions.
Four candidate boundaries distinguish positive hints, flat keys, resource-key
spoofs and legacy aliases. Deterministic checks cover optional dimension priority,
HTTP/RPC precedence, explicit DB choice, native seconds-only eligibility, exact
scope/grain and stable default ordering under descriptor permutation/count
changes. The focused campaign passed in 0.098s
(`artifacts/db-dashboard-properties.txt`); the actual-ledger test compiled and
skipped without its artifact variable.

`TestDBDashboardObservedPopulations` reuses the independent raw-record oracle via
`GOO_DB_DASHBOARD_LEDGER`. Actual journey189
`db-dashboard-00281d8421/ledger.json` passed on its first invocation in 0.145s:
25 full-population responses, including DB 137 versus mixed 189, literal SELECT
120, explicitly filtered 108 and actual trace continuation 1, matched the raw
facts. Each guaranteed changed count was rejected. The adjacent
`native-validation.txt` retains stdout; no observer correction, fixture or
application change was needed. No new model, mandatory generated case quota or
synthetic lifecycle replay is introduced for this pure-planner addition.

After browser ownership is released, use the toolchain/cache environment above:

```sh
go test -run '^Test(HegelDBDashboard|DBDashboardCandidates|DBDashboardNative)' -v .
GOO_DB_DASHBOARD_LEDGER=/workspace/goo/artifacts/db-client-dashboard/db-dashboard-00281d8421/ledger.json go test -run '^TestDBDashboardObservedPopulations$' -v .
```

## Messaging role dashboard slices (C2b.2)

`TestHegelMessagingDashboardRolePopulations` uses 12 generated exhaustive grids
for both roles across service/resource namespace, reported kind and missing/empty/
custom systems. The shared independent predicate interpreter checks every actual
role/log question. Focused deterministic checks cover four hint-role combinations,
resource-key and missing-system refusals, literal dimension priority, all family
precedence, explicit Producer intent, native processing seconds-only eligibility,
exact scope/grain and stable ordering under permutation/record-count changes.
The first campaign passed in 0.145s
(`artifacts/messaging-dashboard-properties.txt`); the actual-ledger entrypoint
compiled and skipped without its artifact variable.

`TestMessagingDashboardObservedPopulations` uses
`GOO_MESSAGING_DASHBOARD_LEDGER` with the shared independent ordinary raw-fact
oracle. Actual journey190
`messaging-dashboard-bc0f633c21/ledger.json` passed on its first invocation in
0.225s: 27 responses matched full raw populations/groups/status/durations and
raw page scope/time/attributes/identities, including consumer 137, producer 120,
mixed 309, destination 120, explicit filtered 108 and trace continuation 1.
A guaranteed changed count was rejected for each response (27 controls). The
adjacent `native-validation.txt` retains stdout. No observer correction or
application/fixture change was needed. This pure-planner extension introduces no
new formal lifecycle model or mandatory evidence quota.

After browser ownership is released, use the toolchain/cache environment above:

```sh
go test -run '^Test(HegelMessagingDashboard|MessagingDashboardCandidate|MessagingDashboardNative)' -v .
GOO_MESSAGING_DASHBOARD_LEDGER=/workspace/goo/artifacts/messaging-dashboard/messaging-dashboard-bc0f633c21/ledger.json go test -run '^TestMessagingDashboardObservedPopulations$' -v .
```


## Generic startup and optional capabilities (C1)

No additional Hegel campaign or lifecycle model was added: startup flags and
capability states form a small finite contract already checked by
`TestGenericStartupWithoutSampleFiles`,
`TestStartupProfilesFailClosedExplicitConfiguration`,
`TestGenericExplicitOperationPackageRemainsAvailable` and
`TestRuntimeCapabilitiesHTTPWireFixtures` in `cmd/goo`, plus
`tests/test_server_startup.py`. These distinguish absent optional configuration,
invalid explicit input, unavailable family responses and telemetry failures.
They do not certify source connectivity or a production deployment.

Actual [journey191](../journeys/191-generic-startup-storyboard.md) uses isolated Go
and Python processes with no sample grammar/configuration/definitions. Accepted
main `artifacts/generic-startup/generic-python-e570bbe8e2/ledger.json` passed in
20s. Its asset-only `runtime.json` and read-only `file-audit.json` match 132 asset
hashes, the frozen binary and the complete capture store, confirm cleanup, and
compare the saved/reopened pages. Complete capture SHA256:
`5a7cccde32af802e66062fec59148f21012dbadadfd9bd27efb12f73bb369844`.
The audit ran no telemetry queries, build or native campaign.

The actor at `generic-python-38477cf34c/bombadil/acceptance-summary.json` passed
30.9s (six phases, 465 snapshots, eight source operations and one plan). The main's
`offline/offline.json` passed 2s against an actually unavailable localhost source:
valid discovery 502, optional-family 503, capability 200 and immutable retained
capture GET. The existing lab regression passed 6s. Artifact durations describe
individual bounded checks, not a universal performance budget.

The first two main attempts found real defects: automatic local tail startup and
a disabled example-catalog request. `artifacts/generic-startup/main-first.log`
and `main-second.log` remain rejected diagnostics. The accepted third run retained
the strict source-free and disabled-example assertions; this was not first-pass
acceptance. The Python-tagged OTLP fixtures are generated ordinary telemetry,
not evidence of executing a separately instrumented Python application.

Reproduce only with the lab source/collector available and the browser slot free:

```sh
# From repository root, using its supported Go toolchain:
go test ./cmd/goo -run 'Test(GenericStartup|StartupProfiles|GenericExplicitOperation|RuntimeCapabilitiesHTTP)'
python3 -m unittest discover -s tests -p test_server_startup.py -v
go build -o /tmp/goo-generic-191 ./cmd/goo
GOO_GENERIC_BINARY=/tmp/goo-generic-191 npx playwright test journeys/generic-startup.spec.js
GOO_GENERIC_BINARY=/tmp/goo-generic-191 GOO_GENERIC_BOMBADIL=1 npx playwright test journeys/generic-startup-bombadil.spec.js
GOO_GENERIC_BINARY=/tmp/goo-generic-191 GOO_GENERIC_LEDGER=/absolute/path/to/ledger.json npx playwright test journeys/generic-startup-offline.spec.js
```

## Native interval to record context (C3)

`TestHegelNativeRecordContextExactMembership` checks 24 generated nanosecond
intervals above 2^53 across counter/histogram parents, both record signals and
missing/empty/named resource namespaces. Independent `big.Int` geometry and raw
fact membership validate millisecond enclosure plus exact half-open membership,
other-instance exclusion, immutable inputs and the strict shared-field whitelist.
`TestNativeRecordContextRefusalsAndPoints` covers canonical bounds, window limits,
invalid signal/identity, pinned-stream replacement and pure gauge point versus
lookaround semantics. The first focused campaign passed in 0.037s
(`artifacts/native-record-context-properties.txt`). No production evaluator is the
expected membership oracle.

Actual journey192 passed bounded native acceptance in 6.878s:

- `TestQuintNativeRecordContextObserved`: 6.85s, accepted
  `artifacts/native-record-context/native-context-a57c48baf4/ledger.json`;
  six semantic and two finite-model controls rejected.
- `TestNativeRecordContextCasesAndOwnership`: 0.02s, accepted
  `native-context-1a693076d7/cases.json` and
  `native-context-1480119d7a/ownership.json`; four changed case counts rejected.

The main checks exact115 raw records, broader trace4/linkedlogs4, integer
boundaries, page identity/scope/time/attributes and retained parent evidence.
Selection resolves the actually displayed Raw record against the original exact
stream set. A rejected initial observer assumed only one stream; all source and
return guards were already true. The first and isolated diagnostic logs are
retained beside final `native-validation.txt`; no fixture/application change was
made to satisfy it.

The existing `population_drilldown` model replays ten steps derived from nine
observed UI events: stage/back/history and five actual source operations, including
two reads in trace expansion. It validates this observed depth/counter protocol;
independent raw-fact checks, not the model, validate interval arithmetic and
population membership. Its two changed counter/order controls reject. The older
fixed-phase `native_navigation` model was not reused. No new model framework or
case quota was introduced.

The four actual scenarios cover histogram empty/missing namespace, a flagged
unknown native interval and an empty record interval. Held-response ownership
checks distinguish two telemetry reads from one cancellation control POST; Return
and stale-response rejection add no source read. The browser main initially had
an obsolete Expand-button selector and passed after a harness-only correction.
The accepted actor (25.8s, eight phases, 418 snapshots, two reads) and separate
3.6s native-entry frame (one fresh read, source-free selection) retain their own
receipts. Child record filters persist when shared/saved; tab-local native origin
is not a durable native citation.

With the toolchain/cache environment above and browser ownership released:

```sh
go test -run '^Test(HegelNativeRecordContext|NativeRecordContextRefusals)' -v .
GOO_NATIVE_RECORD_CONTEXT_LEDGER=/workspace/goo/artifacts/native-record-context/native-context-a57c48baf4/ledger.json \
GOO_NATIVE_RECORD_CONTEXT_CASES=/workspace/goo/artifacts/native-record-context/native-context-1a693076d7/cases.json \
GOO_NATIVE_RECORD_CONTEXT_OWNERSHIP=/workspace/goo/artifacts/native-record-context/native-context-1480119d7a/ownership.json \
go test -run '^Test(QuintNativeRecordContextObserved|NativeRecordContextCasesAndOwnership)$' -v .
```

## Counted-values source projection (C4)

`TestHegelCountedValuesReferenceAndSourceBoundary` uses 24 generated raw datasets
with independent canonical int64 filtering and missing/empty/literal partitions,
for both record and resource dimensions. An in-memory HTTP RoundTripper supplies
those aggregates to the real source decoder, checking exactly two bounded source
requests, detached question identity, count/group conservation, raw-read bases and
read-clock ordering. Deliberate count disagreement remains
`changed_between_reads` with diagnostic groups. Eight malformed response cases
and caller mutation after the first read exercise refusal and detachment.

The focused gate passed in 0.031s (`artifacts/counted-values-properties.txt`).
The first attempt was blocked by socket sandboxing; replacing the fake listener
with injected in-memory transport avoids network permissions. A second diagnostic
corrected the test's dual numeric/untyped field declaration to the existing
Numeric-only predicate DSL. Both diagnostics are retained; no application source
was changed to satisfy the tests. This is source wire/decoder/reference-boundary
coverage, not SQL execution parity. Actual journey193 supplies the real
ClickHouse population evidence described below.

`TestCountedValuesObserved` passed the accepted actual
`artifacts/counted-values-projection/counted-projection-ab82d8e588/ledger.json`
on its first invocation in 0.177s (test body 0.13s). An independent raw-fixture
predicate/numeric oracle derives large143 and small5 complete group partitions,
missing versus empty values, original full143/page100 and explicit missing1 row
membership/attributes/time. It verifies exact parent projection, question binding,
read receipt structure/clocks and the held-response stale-publication receipt.
Eight serialized, guaranteed changed controls reject wrong total/group counts,
missing-to-empty collapse, invalid receipt, wrong binding, a rounded >2^53
threshold, changed-read state and stale publication. Stdout is retained at
`artifacts/counted-values-projection/native-validation.txt`.

The observed page request sequence has five telemetry API POSTs: three full
Explore and two counted-values requests. One query-cancellation POST is a control,
not a read. The small-population direct API request and setup readiness probes are
outside that page journal. This is exact-scope/raw-membership and bounded observed
owner verification, not a claim to observe every internal source SQL statement or
prove all race schedules. The accepted browser main passed 21.7s and actor passed
31.6s (seven phases, 511 snapshots, three source operations with one counted lookup).

No new generic model is added. The old single-filter-editor ledger is a one-step
staging check, not a suitable execution/cancellation lifecycle proof for this
journal. Existing broad capture tests retain legacy wire/hash invariants.

Using the toolchain/cache environment above after browser ownership is released:

```sh
go test -run '^Test(HegelCountedValues|CountedValuesMalformed)' -v .
GOO_COUNTED_VALUES_LEDGER=/workspace/goo/artifacts/counted-values-projection/counted-projection-ab82d8e588/ledger.json go test -run '^TestCountedValuesObserved$' -v .
```

## Service record-aggregate projection (C4 second caller)

`TestHegelRecordAggregatesLegacyProjectionParity` uses 18 generated raw datasets
with null/empty/value groups, durations above 2^53 and all recorded status classes.
Independent integer summaries/order statistics feed the real old/new source
decoders through socket-free in-memory transport. Valid aggregate projections
match while primary reads fall 5→3 and dimension reads 4→2; receipt presence and the
absence of rows/fields/plan/token are checked. Group/series duration disagreement
retains changed-read diagnostics without quantile merging. The focused campaign
passed in 0.088s (`artifacts/record-aggregates-properties.txt`). This is decoder/source
boundary parity, not SQL execution or capture-byte equivalence proof; existing
legacy fixtures and broad capture tests retain their own contract checks.

The shared 187–190 observed verifier now branches explicitly on
`goo/record-aggregates/v1`, validating complete groups/status/duration/order
statistics/time series against raw span facts and separately supplied log facts.
Full supporting/drilldown Explore results retain raw-page guards; no synthetic
rows or weakened legacy DTO acceptance are introduced. The final shared-validator regression passed in 0.103s: accepted journey187
now checks 22 full span responses, including three supporting metadata pages;
the new projection entrypoint compiled in that pre-browser gate. Completed
journey194 acceptance is recorded below.
`artifacts/record-aggregates-legacy-validator.txt` retains stdout. This is an
archived-187 validator regression, not a fresh replay of every 187–190 browser
journey after their transport expectations changed.

Actual 194 native acceptance passed: main and owner together 1.008s (main 0.77s,
owner 0.19s), then the owner-only explicit original-HTTP equality guard passed
0.218s (body 0.16s). The final log is
`artifacts/service-dashboard-projection/native-validation.txt`. Main
`service-projection-b7b0685a04/ledger.json` covers raw 137 spans, 155 independent
logs (ERROR 23/INFO 132), four route groups, exact per-group/global order statistics,
zero-filled time buckets and per-read receipts. Full supporting witness and
explicit drilldown pages keep raw identity/attributes/time guards. The full
137-row stored capture (100-row public page) validates its content-addressed ID,
unchanged 138964 bytes and SHA256
`9837ae1af29939e979f715421ac830077674ae05b4ee551bcbd3d7ba011256c2`.
Eight guaranteed changed main controls reject count, group quantile, series,
scope, receipt, inspection-field leakage, missing full rows and a hidden return read.

Owner `service-projection-fbc939b42b/ownership.json` retains the actual complete
All-spans 198 DTO and its actual pure-plan receipt with original service/scope;
fresh HTTP recovery must equal the original exact question. Four changed owner
controls reject held count/binding, stale visibility and a fresh receipt error.
An initial observer incorrectly bound the held read to the original HTTP 137 plan;
`native-validation-original-profile-diagnostic.txt` preserves that rejected
assumption. Correcting its plan binding changed no fixture or application source.

The accepted browser main 19.9s/owner 8.0s and actor 31.3s (nine phases, 462 snapshots,
11 source API operations) remain separate from native checks. Initial harness
failures involved unsupported capture browser-history navigation and a collector
waiting on canceled bodies. The accepted collector binds partial-body diagnostics
to actual request-failure/cancellation metadata and preserves true header status;
it supplies no positive substitute data. Header 202 cancellation receipts with
`bodyObserved:false` are controls, not telemetry bodies or upstream-completion
proof. Complete held DTOs remain strictly validated. Fixed-question Back is
source-free; capture return is explicit Services navigation, not a fabricated
history edge. No new formal model is required for this projection adoption.

Using the toolchain/cache environment above after browser ownership is released:

```sh
go test -run '^Test(HegelRecordAggregates|RecordAggregatesChanged)' -v .
GOO_SERVICE_AGGREGATES_LEDGER=/workspace/goo/artifacts/service-dashboard-projection/service-projection-b7b0685a04/ledger.json \
GOO_SERVICE_AGGREGATES_OWNERSHIP=/workspace/goo/artifacts/service-dashboard-projection/service-projection-fbc939b42b/ownership.json \
go test -run '^TestServiceRecordAggregates(Observed|Ownership)$' -v .
```

`TestHegelStockObservations` checks signed fractional cumulative stocks and epoch preservation under descending values, reordered arrival and duplicate replay. It is an in-memory prerequisite, not source/dashboard acceptance; see [contract](../docs/CUMULATIVE_STOCKS.md).

## Stock measurement integration (C2c)

The existing `TestHegelStockObservations` remains the kernel property: decreases,
accumulation epochs, arrival permutation and identical replay are already covered.
The integration adds a typed native sum path and `result.stocks`; it does not
change float64 observations into exact-integer values. New source/observed checks
will focus on raw exporter values, epochs, stream identity, source/replay mapping
and full capture bytes. Actual journey195 acceptance is pending; no new lifecycle
model or redundant kernel campaign is claimed.

`TestHegelStockAdapterIdentityAndEpochs` passed on its first invocation in 0.141s
(`artifacts/stock-integration-properties.txt`). Twelve generated signed-fraction
shifts cross six distinct namespace/instance streams. Independent physical input
facts check every timestamp, value and accumulation epoch, including decreases
and epoch transitions; valid serialized replay succeeds and an edited epoch is
refused. This tests public adapter/replay integration without calling the stock
kernel as the expected oracle. The pre-existing kernel property was not rerun.
Actual journey195 source/capture acceptance remains pending.
