# Implementation plan

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
