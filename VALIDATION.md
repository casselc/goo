# Validation record

Current additions: see the [cloud baseline](ENVIRONMENT_VALIDATION.md) for the
successful default Compose run, real cross-app Bombadil journeys and generated
Quint MBT success/fault campaigns in this environment. The historical hosted
environment limitations below describe the original handoff session.

## Contextual investigation follow-ups (2026-10-18)

Evidence cards now propose source-free questions from verified captures, with
explicit preserved and changed scope. Go compiles generated record/measurement
questions; exact trace/span IDs remain distinct from service/time resource
context. Discovery retains the original fixed window through package preview.
Authored evidence gaps cite the originating capture and remain unsaved drafts.
See [the contract](docs/INVESTIGATION_FOLLOWUPS.md) and storyboard 42.

Passed: full root Go suite, exploration/HTTP race tests, root vet, 32 frontend
checks, Python compilation and Hegel state/swarm (100 requested sequences, 25
steps). Hegel exposed shared predicate trees between returned actions; each
query now owns an independent deep copy, with a unit regression and passing
property rerun. Stored capture immutability is checked separately.

The focused real-stack journey passed in 1.4 minutes: reproduced delayed sample
work, compared duration/status over the same matching population, followed exact
trace/log IDs, explicitly discovered instruments in the original window and
saved a finding with an origin-cited instrumentation question. It does not infer
resource causality. Its quint-go ledger passed and rejected seven altered ledgers;
the new Bombadil source-free preview campaign passed. The initial attempt failed
only a textarea assertion using text content instead of value; corrected assertions
passed. Initial artifacts remain in `artifacts/followups-initial-failure/`.

Evidence: `artifacts/followups-{go,race,vet,node,hegel}.log`,
`artifacts/followups-journey-final.log`, and
`artifacts/followups/7cc4cfd1/`. Final image `77be6b447ade`: all six affected browser journeys passed in 5.5
minutes, with no skips: coherence 40, context notes 36, follow-ups 42,
investigation workspace 41, measurements 38 and metric discovery 39. All five
enabled Bombadil campaigns and associated live Quint/negative controls passed.
The expanded mobile preview passed containment and source-free checks. Affected
screenshots were regenerated. Final ledger: `artifacts/followups/aec68658/`;
regression log: `artifacts/followups-regression.log`. The wrapper restored owned
traffic and preserved existing telemetry volumes and five stale pending entries.

## Question-first investigation workspace (2026-10-18)

Investigations now retain a question, current finding, open questions and ordered
cross-signal evidence. An active title/question context follows navigation.
Explore's capture dialog can preserve only the capture, append to the current
investigation draft, or publish a child of an explicitly selected saved revision.
Unrelated dirty drafts remain intact; capture and attachment publication are
separate, and attachment failure leaves the capture available. Cards show stored
receipt scope and native definition identity through a receipt-only API, without
returning record banks or querying telemetry.

Passed: full root Go suite; affected workbench/HTTP race suites; root vet; all
32 frontend checks; expanded Hegel context state/swarm (60 requested sequences,
15 steps) covering complete brief mutation, fork and preserved history. Explicit
storage tests preserve the pre-brief version-1 hash/JSON shape, normalize omitted
question-only card lists and reject oversized summaries. The receipt HTTP test
runs with its telemetry source closed and verifies that no source read occurs.

Initial storyboard 41 passed (1.2 minutes) with real checkout/report spans, report
logs and an independent ordinary OTel counter. Its dedicated quint-go ledger
passed and rejected six mutated ledgers. Final deployed image `d13fcbdbce4a`:
all four browser journeys passed, with no skips, including captures 32, context/
notes 36, investigation workspace 41 and measurements 38. All four enabled
Bombadil campaigns and the associated live Quint/negative-control checks passed.
The final new journey additionally creates a saved branch while retaining an
unrelated dirty question draft, with no telemetry read. Fresh metric definition
reuse adds a capture/child revision while the prior capture and handoff reopen
unchanged. Desktop/mobile containment and screenshots cover the connected flow.

Artifacts: `artifacts/investigation-{core-go,core-node,race,vet,hegel}.log`,
`artifacts/investigation-workspace-journey.log`,
`artifacts/investigation-final-journeys.log`, and
`artifacts/investigation-workspace/3d3506c8/`. The diagrams, contract, planning docs,
README and affected storyboards were updated. No telemetry volumes were reset;
the five pre-existing stale pending entries were retained and traffic restored.

## Product essence and connected investigation flow (2026-10-18)

[Product brief](docs/PRODUCT_ESSENCE.md): definitions describe meaning, captures
preserve fixed evidence, and investigations explain it. The README now leads with
this loop. Definitions exposes the same measurement revisions as Explore, with
local search; saved record packages also have local search. Service inventory
carries explicit service/environment into metric discovery without automatically
reading telemetry. Measurement stages, canonical YAML, named time scopes and
consistent cards make authoring/execution boundaries visible. Discovery supports
explicit cancellation and suppresses publication after navigation or mode changes.

Passed: 32 frontend checks and all five affected browser journeys across the
initial run and focused rerun: coherence 40, inventory 31, measurements 38, metric
discovery 39 and attribute/facet polish 29. The initial run had three passes and
two harness failures: an inventory selector assumed one button per row, and
Bombadil refused an existing output path. The selector now uses the intended
Explore action's accessible name; discovery campaigns use unique output paths.
The corrected inventory/discovery rerun passed both cases (1.7 minutes).
Original failure artifacts are retained under `artifacts/coherence-initial-failures`.

The new journey passed its delayed **actual source response** abandonment check,
source-free definition preview, service handoff, canonical authoring and mobile
containment. Measurement, inventory and discovery Quint ledgers/negative controls
and all three enabled Bombadil campaigns passed. Discovery exploration includes
abandon/retry and scope-change invariants. Twenty storyboard images were generated
or refreshed, including three new frames for journey 40; desktop/mobile frames
were inspected. No measurement kernel or source query semantics changed.

Evidence: `artifacts/coherence-node.log`, `artifacts/coherence-journeys.log`,
`artifacts/coherence-final-journeys.log`, and `artifacts/metric-discovery/`.
The query/packages image is `11a6f58b5951`. Traffic was restored and the five
pre-existing stale pending entries and telemetry volumes were retained.

## Ordinary OTel metric discovery (2026-10-18)

Bounded service/environment discovery now proposes editable typed packages over
ordinary gauge/sum/explicit-histogram observations. Optional scope version/schema
remain exact identity values, including empty values. No invented identity or
implicit fleet aggregation. Stream summaries expose sorted dimensions.

Passed: root Go suite; race checks for measurements/metrics/HTTP; root vet;
32 frontend checks; Hegel optional-identity properties and existing measurement
state/swarm checks. The independent Python OTel SDK producer has no goo manifest
and omits scope version/schema. The initial live discovery/package/save/run/
capture/reopen storyboard passed (21.1 s). The expanded run passed both affected
journeys (3.1 minutes): measurement storyboard 38 and discovery storyboard 39,
including notes, package saving, capture reopening and native dimension labels.
Bombadil recorded 440 measurement and 248 discovery observations with zero
violations. The shared measurement Quint ledger passed with seven rejected
mutations. Nine affected storyboard images were refreshed.
The final discovery-only rerun passed (22.5 s), including its dedicated Quint
ledger with three rejected mutations: wrong service, missing proposed definition
and tampered evidence basis. Traffic was restored and stale pending data retained.

Artifacts: `artifacts/metric-discovery-{go,race,vet,node,hegel}.log`,
`artifacts/metric-discovery-final-journeys.log`,
`artifacts/metric-discovery-model-journey.log`, and
`artifacts/metric-discovery/ledger.json`.

## Typed measurement packages and standardized examples (2026-10-18)

[Contract](docs/MEASUREMENT_PACKAGES.md): twelve YAML packages bind actual queue
and application telemetry to Go gauges, reset-aware counters, native histograms
and scoped lifecycle preparation. Five data-only ordered workflows preserve a
fixed window and preview without source execution. General captures, immutable
package revisions and ordered notes are shared across all families. Both Go
examples now emit real trace-correlated OTLP logs; Python lifecycle records carry
exact business nanoseconds and stable per-emission event IDs.

Passed native checks: root Go suite, root race suite/vet, instrumentation module,
19 Python tests, 32 frontend checks, portable compiler/lifecycle/package CLI parity
storyboards and Compose configuration. The full verification module passed
uncached (240.18 s), including Hegel state/swarm and existing Quint MBT/trace
checks. Live-ledger tests in that module intentionally skip without supplied
artifacts; the browser journey supplies its actual measurement ledger.
Additional race coverage checks atomic custody publication by 16 concurrent
readers. Binding tests reject a consistently recomputed receipt that falsely
claims complete coverage, not merely an inconsistent receipt.

Focused storyboard 38 passed (1.4 minutes): ordinary reservation scenarios,
Go checkout/report work, independent ClickHouse gauge counts and actual Go logs,
all four measurement families, mature good/retry/missing/rejected/late outcomes,
scoped subject drilldown, fixed question capture/refresh, source-free reopening,
exact notes reference, immutable package save and fresh reuse. Both record and
measurement workflow steps preview without telemetry reads. Quint accepted the
real ledger and rejected seven altered ledgers. Bombadil recorded 462 observations
through initial/preview/invalid/complete stages without invented evidence.

The journey caught an incorrect checkout UI assertion and a real transport bug:
ClickHouse's escaped JSON spellings changed a raw-input evidence hash after a
browser round trip. Source inputs are now canonicalized before hashing, with a
transport regression test. A source metadata mismatch also exposed the Python
meter's missing version/schema; those are now emitted and bound explicitly.
Stored source plans remain reopenable as evidence after compiler changes;
refresh requires an identical current plan.

Broad browser regression: 37 passed, one failed (19.1 minutes), with all
opt-in generated/fault/Bombadil campaigns enabled. The failure was the fault
campaign's six-second actor budget expiring before any browser observation, not
an application or UI invariant failure. The harness now permits browser startup
headroom while preserving its five-second application-response property, and
reports empty traces explicitly instead of a JSON parse error.

Final targeted run on the deployed final image: **2 passed, no skips (4.0
minutes)**. The corrected fault campaign completed all three scenarios and five
inspections/revisits with its model checks (2.5 minutes). Measurement storyboard
38 passed again (1.4 minutes), including valid empty-selector capture
normalization, rejection of unsupported lifecycle scope filters, and a saved
same-name custom package that cannot shadow a builtin workflow's environment.
Thus all 38 browser journeys were verified across the broad run and corrected
focused rerun. All 132 storyboard images (131 PNGs and one JPEG) were refreshed; seven new frames and
architecture flow/sequence diagrams document this slice.

Artifacts: `artifacts/standardization-regression.log`,
`artifacts/standardization-final-journeys.log`,
`artifacts/standardization-race.log`, `artifacts/standardization-verification.log`,
and `artifacts/measurements/360b30f8/`. The first failed fault trace is preserved
under `artifacts/standardization-regression-failure/`; the successful fault run is
`artifacts/bombadil/fault-campaigns/40131596-b9ce-4cb2-9834-9270c7047939/`.
Final scope/normalization/default-resolution checks also have native regressions.
The final measurement Bombadil run recorded 565 observations through all four
required stages; its eight model reports contain one pass and seven intended
conformance rejections.

The broad run used query/packages image
`sha256:7494aa650a6be69c8d06e7d58a1fc88e93149d88b057600bf0beb2da0eea0010`.
The final targeted image, including selector normalization, strict lifecycle scope
validation and builtin workflow resolution, is
`sha256:4df78f6a14a53320181c8f0934f32b73fe1648c22938ab77f215dfb2d6f49b48`.
Runtime frontend hashes match the repository. Telemetry volumes and all five
pre-existing stale Redis pending entries were preserved. The local custody clock
is not warehouse ingestion or a completeness guarantee; missing completion stays
unknown. Databricks YAML was structurally checked, not warehouse-executed.

## Scoped recipes and shared sample investigations (2026-10-18)

[Contract](docs/INVESTIGATION_RECIPES.md): three Go recipe transforms preserve the
applied population and fixed bounds, with source-free preview and explicit
execution. Six ordinary YAML packages move sample investigations onto shared
Explore/package/capture/notes workflows. The specialized lifecycle/resource lab
starts only when opened. Opening a derived recipe question now clears the previous package banner.

Passed native checks: root Go race suite and vet, 19 Python tests and 27 frontend
checks. Native cases cover OR/filter preservation, immutable inputs, parameterized
selectors, wrong-signal rejection, strict HTTP bodies and all six example packages.
Hegel state/swarm passed 100 sequences of 25 recipe/signal/scope/predicate steps
(0.19 s reviewed).

Initial browser harness failures were corrected for the actual save message and
retained disclosure state on same-document navigation. A reserved Quint action
name was renamed. The conformance check then correctly rejected a capture taken
after package reuse changed the time bounds; the journey now verifies the fresh
reuse separately and captures the original fixed question before deriving
a reusable definition from that stored capture. This distinction is part of the documented contract, not a change to
package semantics.

Final browser evidence: the broad 37-journey run completed with 36 passes and
one legacy-sidebar harness timeout (14.8 minutes); that missing disclosure action
was corrected. A final focused run passed both the corrected definition/invalid-
query journey and the strengthened recipe journey (2 passes, 1.2 minutes). The
latter uses ordinary UI filters, captures/cites the fixed question first, then
creates and freshly runs a package from its stored capture. No skips occurred in
the broad run; generated/fault campaigns and cross-app Quint were enabled.

The final recipe ledger passed Quint with six rejecting controls (13.04 s).
Bombadil recorded 467 observations through all four required stages and all three
recipe choices. Wrong-signal rejection, delayed-preview isolation, source-free
notes/preview, literal query preservation and mobile containment were checked.
All 125 storyboard images were refreshed. Five new frames and architecture flow
24 document the slice; new desktop/mobile frames were visually reviewed.
See [environment evidence](ENVIRONMENT_VALIDATION.md) for image and artifact paths.

## Deployment context and ordered notes (2026-10-18)

[Contract](docs/WORKBENCH_CONTEXT.md): authored deployment receipts preserve event
and server custody times; explicit fresh local context accompanies fixed-time
Explore charts without modifying telemetry receipts. Ordered plain-text narrative
cards cite exact general-capture or deployment IDs. Saving creates immutable
parent/root revision branches, preserving older evidence and interpretations.
Dirty/unapplied drafts are retained; no mutable latest head or shared-author claim
is introduced.

Passed: root Go race suite/vet, 19 Python tests and 26 frontend checks. Native
checks cover half-open environment scope, separate custody time, unsafe URLs,
integrity, marker/card/catalog limits, reference validation, pagination and
concurrent branch preservation. Hegel state/swarm passed 60 sequences of 15 steps
(0.38 s reviewed), including Unicode note content, save/fork/order/history,
deployment arrivals and scope/window changes.

The initial browser flow completed its UI assertions but failed verification due
to a test type name colliding with an existing discovery model. Renaming it fixed
the verification build; the captured ledger replay then passed. Journey 36 next
passed with observed evidence/narrative transitions and a seven-stage Bombadil draft prefix
(383 records), including real report-app work, independent ClickHouse trace
discovery, marker boundary/environment exclusions, literal-text safety, ordered
reference editing, immutable revisions/branches, unapplied-draft retention,
source-free capture opening and mobile containment. Quint originally used five
rejecting controls; the strengthened adapter also checks the report's scoped
ready event and rejects a sixth incorrect-subject control (19.97 s replay).
Five new storyboard frames and architecture flow 23 document this slice.

All 36 journeys passed in one broad run (12.6 minutes), with both generated
campaigns and cross-app Quint enabled and no skips. All 120 storyboard images
(119 PNG and one JPEG) were refreshed. Final keyboard/capture-navigation guards
passed three focused journeys (2.0 minutes), including keyboard reorder/remove
focus, cited-evidence focus and delayed-revision isolation. Screenshot inspection
then caught an obsolete loading message after a canceled read; a final status
fix passed journey 36 again on the reviewed image (1.2 minutes), with six rejecting
Quint controls (17.15 s) and 310 Bombadil records through all seven stages.
The final narrow-screen frame was visually inspected. Evidence and image hashes
are in [the environment record](ENVIRONMENT_VALIDATION.md).

## Error and latency comparisons (2026-10-18)

[Contract](docs/ERROR_COMPARISONS.md): exhaustive captured span-status categories,
full-population error-percentage ranking, independently computed mean/p95/p99
changes across fixed windows, retained-baseline group drilldown and explicit
server-reread saved comparison pairs with links to both record captures.
Unset and unrecognized status remain visible; no-data/disagreement suppress
normalized differences. Existing count-only captures/SQL remain compatible.

Passed: root Go race suite/vet; focused capture-refresh race checks; 19 Python
tests; 24 frontend checks; Hegel comparison state/swarm (100 sequences of 25
steps, 0.54 s initial / 0.67 s model regression), plus existing cohort state/swarm
and Quint cohort regression (7.55 s). Native checks cover invalid/missing status
counts, category drift with identical total counts, missing groups, package YAML,
interval validation, source-free saved pair reopening, refresh child lineage and
retaining the predecessor on source outage.

Journey 35 passed initially (54.8 s) and on the polished image (47.0 s). It checks
80 previous/124 current real OTLP spans with only 100 current displayed rows,
2.50% → 16.13% recorded errors (+13.63 pp), +90 ms p95, missing/empty routes,
saved comparison reopening, changed-route-to-trace/log navigation and mobile
containment. The final new-feature campaign passed Quint with four rejecting
controls (7.20 s) and 342 Bombadil records through all six required stages.
All 11 affected regression journeys passed (3.9 minutes). A final trace-action
guard was checked by both discovery and the strengthened comparison journey
(2 passes, 1.3 minutes), including live adapter status-class groups and saved
trace-ID pair behavior. The final comparison campaign recorded 501 Bombadil
records through six stages. Five new frames and affected existing storyboards
were regenerated; see [environment evidence](ENVIRONMENT_VALIDATION.md).

## Grouped duration measures and polish (2026-10-18)

[Contract](docs/GROUPED_MEASURES.md): complete per-group duration aggregates,
normalized semantic results, exact rational ranking, fixed local-search scales,
retained capture pages and reusable package previews. Count-cohort comparison
projects out duration metadata and remains a record-proportion comparison.

Passed: root Go race suite/vet, 19 Python tests, 23 frontend checks and Hegel group
count/sum conservation over 100 generated populations (0.254 s), including uint64
extremes. Native source tests reject missing summaries/duplicate dimensions and
preserve same-count duration disagreement. A live fixture verifies 124 spans versus
100 displayed rows, missing/empty routes, zero duration, p95 ranking, one/two-field
groups, count-cohort projection, package preview and mobile containment.

The initial extended campaign passed five of six journeys, including journey 34
(52.1 s), Quint with three rejecting controls (7.32 s) and 454 Bombadil records
through all nine saved-ranking prefix stages. The discovery regression exposed a
removed accessible trace-button label, now restored. Browser checks also found a
blur-linked control shift, fixed by input-time scope invalidation. Typed model
comparisons preserve exact values without treating JSON indentation as evidence.
The final reviewed-image run passed all 10 selected regression journeys (3.5 minutes), including the original discovery assertion. The final group campaign passed three rejecting Quint controls (6.46 s) and 383 Bombadil records through nine stages. Four new and affected existing frames were refreshed; evidence paths are in [the environment record](ENVIRONMENT_VALIDATION.md).

## Core workbench basics (2026-10-18)

[Contract](docs/CORE_WORKBENCH.md): opt-in full-population span-duration summaries
(count/exact sum/min/max/upper p50/p95/p99), fixed count-over-time intervals with
per-interval durations, immutable capture retention, reusable versioned questions,
span/log service activity and reported parent links. General query admission and
explicit cancellation reach the Go context through the Python proxy. This remains
a local lab; event age is not ingestion lag, and query IDs are not authorization.

Passed: root Go race suite and vet, 19 Python tests, 22 frontend checks, Compose
configuration, Hegel duration state/swarm (100 sequences up to 30 steps). Native
HTTP tests exercise saturated admission, source-free local access, cancellation
propagation and delayed-request tombstones. Source tests reject malformed/missing
summary fields, invalid partitions/bounds and ambiguous dimensions; exact sums
cover values beyond uint64 and JavaScript integer precision.

The first deterministic journey 33 passed in 18.9 s. The strengthened focused
capture/core/inventory run passed all three journeys in 2.3 minutes, including
187 Bombadil records through all six core prefix stages and actual-ledger Quint
validation (13.26 s) with four rejecting controls. Five controls now check plausible-but-wrong p95/sum, ambiguous-link classification, cancellation and reused environment scope. All 33 journeys passed in one comprehensive run (11.9 minutes), with both generated campaigns enabled and no skips. All 106 storyboard images were refreshed. The expanded final campaign passed again with both Bombadil actors (216 saved-measure records plus 181 live-workflow records), all required stages and five strengthened rejecting controls. Artifact paths are in [the environment record](ENVIRONMENT_VALIDATION.md). Architecture flow 20 and four
new screenshot storyboards document this increment.

## General evidence captures and stable saved paging (2026-10-18)

[General captures](docs/GENERAL_CAPTURES.md) now persist manifest-free span/log
questions and service inventory in the existing investigations volume. Reopen and
paging read verified local evidence; refresh creates a successor with the same
normalized fixed question and compiled source plan. Explore captures retain up to
1,000 records, shown in pages of 100, while full supported aggregates retain their
own counts and source states. These bounded records do not claim source completeness.

Passed: root Go race suite and vet; focused race checks after final normalization;
19 Python tests and 21 Node checks. Native HTTP tests reopen with the source closed,
reject client-supplied evidence and cross-revision cursors, and preserve predecessors
on failed refresh. Hegel state/swarm passed 60 sequences of up to 15 steps against
final code (12.939 s), including arrivals, source resets, captures and branching
refreshes. Quint validates actual browser capture/page/refresh/reopen observations,
with four rejecting controls for old-count, page-basis, question and cursor drift.

Journey 32 covers 1,205 matching spans, a 1,000-record retained bank, stable pages
after late arrival, a 1,206-span successor, source-free browser reopening after
process restart, explicit live handoff, inventory, logs, narrow screens and stale
response isolation. Six new storyboard frames accompany diagram 19. Expanded trace
and log context remains a separate fresh read; saved baseline pairs, team access
control and collaborative editing remain future work.

The initial default suite had 29 passes, two opt-in skips and one outdated wording
assertion in journey 22. That assertion now matches the general “saved captures”
disclaimer; its semantic checks remain and focused journeys 19 and 22 passed.
Final capture/Bombadil and both opt-in campaigns passed (three tests, 3.5 minutes); all 32 distinct journeys passed across these runs. The capture actor recorded 367 trace records through all eight stages. All 102 storyboard images were refreshed; detailed evidence paths are in the cloud record.
An initial final-test selector matched no tests, and one retry lacked workspace Go
cache variables; both invocation issues were corrected before final verification.

The agent's [collaboration research](docs/COLLABORATIVE_INVESTIGATIONS_RESEARCH.md)
distinguishes verified Honeycomb Boards and Datadog multiplayer Notebooks from an
unverified Honeycomb whiteboard identification. It proposes ordered investigation
cards before authorized collaboration/presence; no multiplayer feature is claimed.

## Observed service inventory

Implemented: single-read full-population service/span-name counts, OTel Error
statuses and fixed span-duration buckets; local filtering preserves captured
accounting; drilldown opens a separate fixed-window Explore read and package
preview. [Contract](docs/SERVICE_INVENTORY.md), [diagram 18](docs/ARCHITECTURE_DIAGRAMS.md#18-discover-services-and-promote-an-investigation-question)
and [journey 31](docs/journeys/README.md#31-from-observed-services-to-a-reusable-question)
record semantics and executable frames. Unknown fleet membership/health and
span-versus-request grain remain explicit.

Passed: root Go race suite/vet, 19 Python tests, 20 Node checks, Compose config,
Hegel observe/reset state/swarm (100 sequences, up to 20 steps), and actual-browser
Quint capture/search/open/package conformance with three rejecting controls.
Source tests retain counts above JavaScript integer precision and reject invalid
partitions, errors exceeding counts, missing/null dimensions, duplicate groups,
cardinality overflow and unavailable-to-zero conversion. Query values remain
parameters; HTTP rejects unknown fields and multiple objects.

The live 116-span fixture checks independent ClickHouse counts, 113 matching
spans versus 100 displayed rows, package preview, local filter stability, no data
and pending-read invalidation. The dedicated read-only Bombadil campaign passed
609 records across all six prefix stages plus refresh/search/scope edits. Evidence:
`artifacts/inventory/inventory-a777c2549e/`. Narrow-screen screenshot review exposed
compressed table headings; stacked mobile cards and a width assertion corrected
that issue, and the updated journey passed. See the cloud validation record for
the subsequent broader browser run and final-image verification.

## Manifest-free ad hoc telemetry packages

Implemented: captured Explore question → named telemetry package YAML → shared Go
query plan → separate live preview → immutable local configuration revision → fresh
Explore reuse. Import/export, nested predicates/two-field breakdown preservation,
local refinements and optional downloadable Go annotation templates are included.
No target application manifest is loaded. See [contract](docs/TELEMETRY_PACKAGES.md),
[diagram 12](docs/ARCHITECTURE_DIAGRAMS.md#12-ad-hoc-telemetry-packages-and-optional-instrumentation)
and [journeys 24–25](docs/journeys/README.md#24-build-a-package-over-existing-python-telemetry).

Passed: root Go race suite/vet, 19 Python tests, 14 Node checks and Compose config.
New Go fixtures check strict bounded YAML, unknown fields/aliases/multiple documents,
unsupported grains/adapters/roles, time/page separation, stable YAML round trips,
concurrent idempotent publication, immutable old revisions, path validation and
corruption detection. HTTP tests pass generated Go code through the actual source
annotation scanner. No generated template is patched or deployed into an app.

Hegel state/swarm passes 100 sequences of up to 30 edit/save/window steps; Quint
observes draft/save/edit/reopen/apply and rejects mutated-original and nonreusable
negative controls. Focused verification passed (5.1 seconds); the complete nested
verification suite passed (95.5 seconds), with ignored log
`artifacts/verification/telemetry-packages-full-suite.log`.

The first actual Python SDK authoring journey passed (9.4 seconds). After preview
and handoff presentation improvements, journeys 24/25 passed together (14.0 seconds):
ordinary SDK spans without goo attributes/manifests, collector ingestion, exact
preview counts, save/export/import, service-restart persistence, fresh reuse,
local refinements, preserved nested log/two-field selectors, no-data preview and
stale-validation isolation. The Node-style log workload is a labelled OTLP fixture,
not a claim that a Node SDK service ran. The generated Go template is scanner-tested,
not source-equivalence, weaving, deployment or coverage proof.

VFS build history exhausted workspace bytes during recreation; unused dangling
images/build cache were reclaimed, retaining running containers, all telemetry
volumes and existing stale pending entries. Recreation then succeeded. No daemon,
proxy, credentials or storage-driver settings were changed.

This package kind supports captured-record counts and existing views, not arbitrary
measures, inferred SLOs/business semantics, automatic source patches, source authenticity,
production collaboration or Databricks lowering. Saved definitions do not archive
preview evidence. Mermaid source/fences and local links are checked without claiming
rendered Mermaid validation.

All 25 real browser journeys passed together (4.1 minutes), including the new
Python SDK and imported-log paths, native metric/fleet regressions, saved receipts
and baseline comparison. All 74 documentation frames were regenerated. Preview
tables, readable handoff code and mobile authoring were visually inspected. Final
editor-focus and metadata/navigation generation guards are covered by a focused
rerun of the two affected authoring journeys, which passed together (16.7 seconds).
Their six new frames were regenerated again after those refinements. All 12 services
are running, traffic is restored, and the five pre-existing stale pending stream
entries and all data volumes remain intact.

## Cohesive general-purpose workspace and cohort distributions

Implemented locally: Explore/Operations/Saved investigations/Definitions navigation,
secondary reservation Examples, shared responsive styles, result tabs and trace/log
inspector. Complete captured distributions support a browser-held fixed-window
baseline and Go rational percentage-point comparison. The shared waterfall accepts
caller-provided related actions; reservation navigation is supplied only by Examples.
See [workspace and scope contract](docs/PRODUCT_WORKSPACE.md),
[diagram 11](docs/ARCHITECTURE_DIAGRAMS.md#11-workspace-navigation-and-captured-distribution-comparison)
and [journey 23](docs/journeys/README.md#23-one-investigation-workspace-from-distribution-to-trace).

Passed: root Go race tests/vet, 19 Python tests, 14 Node checks, strict comparison
HTTP validation and Compose configuration. Fixtures cover exact uint64 counts,
missing/empty identities, category unions, scope drift, duplicate/partial groups,
empty populations and changed-read unknowns. Hegel cohort state/swarm runs 100
sequences of up to 30 steps. Quint validates observed pin/selection/empty/changed/
scope-change results with three rejecting negative controls. Full nested verification
suite passed (91.9 seconds); ignored log: `artifacts/verification/workspace-full-suite.log`.

The initial targeted seven browser journeys passed (37.6 seconds), covering actual
workspace navigation, old and new query paths, saved/live expansion, and a labelled
12-record collector fixture compared with a four-record filtered population. The
source fixture is documentation telemetry, not a production error attribution.
Mobile screenshot review prompted labelled comparison rows and visible navigation.
The full run then passed 22/23 journeys (4.8 minutes). Journey 07 exposed a
pre-existing readiness assumption: its five-minute API poll could see an older
failure while the UI used the new two-minute default. Its fixture now aligns
the window and waits for a newly captured failure identity. The first focused
rerun also encountered an actual transient source error during baseline acquisition;
readiness now checks HTTP success before accessing facts. Source failures are
not converted into zero-valued evidence. Expanded journey 23 passed again (6.7
seconds), including distribution-value refinement in the baseline's fixed window.
The corrected journey 07 then passed (4.0 seconds; focused command 6.3 seconds).
Together with the full run and focused journey 23, all 23 journeys have passing
results against this round's runtime. All 68 documentation frames were regenerated;
desktop comparison/definitions and mobile comparison/inspector frames were visually
reviewed. The inspector frame uses the actual viewport, avoiding full-page fixed-overlay
screenshot distortion. All 12 services are running and traffic is restored. Existing
five stale pending entries and all volumes were retained.
Cohort observations were strengthened to check expected positive/empty/unknown
results explicitly; focused Hegel/Quint verification passed again (4.7 seconds).
Baselines remain volatile browser state; caller-held comparison evidence is not
source-authenticated. Optional HyperDX, a Databricks instance, production deployment
and a complete accessibility audit were not validated. Mermaid source/fences and
relative documentation links were checked; no rendered-Mermaid validation is claimed.

## Bounded general exploration and expanded trace/log context

Implemented: shared Go query model/reference predicates and parameterized
ClickHouse compiler; UI/API/CLI source planning; discovered attribute keys,
string and UInt64-duration predicates with bounded AND/OR; independent captured
record counts/groups and bounded pages; missing-vs-empty grouping; detected
cross-read disagreement; cross-service trace pages and exact trace/span log scopes.
Operation/saved waterfalls explicitly expand into separate live evidence. See
[design](docs/EXPLORATION_DESIGN.md),
[diagram 10](docs/ARCHITECTURE_DIAGRAMS.md#10-general-exploration-and-independent-expanded-evidence)
and [journeys 20–22](docs/journeys/README.md#20-general-exploration-and-population-safe-paging).

Passed: root Go race tests/vet, 19 Python checks and 12 Node checks. Core/HTTP
fixtures cover type/operator/budget validation, literal parameter safety, exact
UInt64 comparisons, missing values, page-independent count/group SQL, fixed
trace scope, source/group overflow failures, detected read disagreement, CLI
planning and invalid requests rejected before source execution. Hegel state/swarm
passes 100 sequences of up to 30 steps; Quint observed search/page/trace/log steps
pass with three negative population/scope/identity controls. All three Dockerfile first-stage COPY contexts compile the root CLI with
VCS stamping disabled for temporary contexts. Full nested
verification suite passed (89.4 seconds), log in ignored
`artifacts/verification/exploration-full-suite.log`.

Live targeted journeys 20/21 passed together (13.8 seconds); expanded 20–22
passed together (37.6 seconds), including trace paging with missing parents and
saved/live evidence separation. An initial journey read collapsed diagnostics
using rendered text and failed; it now reads the actual stored text content.
Direct read-only ClickHouse checks found that raw HTTP query parameters reinterpret
backslashes and reject literal tabs. The adapter now escapes ClickHouse parameter
text as well as URL encoding. Exact backslash/tab/newline/Unicode/DEL round trips
were verified, and journey 20 now includes literal transport characters in both
keys and values. This is a transport correction, not an SQL interpolation path.

All 22 live browser journeys passed together (4.2 minutes), including the literal
transport regression in the 123-record fixture. The 63 documentation frames were
regenerated. After a UI origin-metadata refinement, focused journeys 20–22 passed again
(22.7 seconds), including the assertion that log→trace navigation carries the log
response question/basis as its origin. Context diagnostics also retain source
state/warnings; actual read disagreement is covered by source fixtures. No Go
query kernels changed after the full browser pass. The six new documentation
frames were regenerated, and the exploration, distributed waterfall/log and
saved/live-context frames were visually inspected. All 12 services are running;
the owned traffic generator is restored and no queues/volumes were cleared.
Mermaid source and relative documentation links were checked without claiming a
local Mermaid render.

General query/cursor persistence, whole-trace completeness/accumulation, live tail,
value distributions, cohort comparison, production authorization and warehouse
lowering remain unimplemented. No production deployment, Databricks instance or
full networked Docker build is claimed. General context is not added to saved
operation receipts.

## Saved fixed-window investigations

Implemented: bounded server-captured immutable operation evidence; local named
volume; package revision/fixed-window successor refresh; reopen without telemetry
queries; stored operation/filter/trace/subject and ratio/member explanation state;
component error preservation and comparison of independent evidence bases.
See [contract](docs/SAVED_INVESTIGATIONS.md),
[architecture diagram 9](docs/ARCHITECTURE_DIAGRAMS.md#9-captured-investigations-and-explicit-refresh)
and [journey 19](docs/journeys/README.md#19-save-reopen-refresh-and-compare).
The capture is an explicit fresh read when saving, not a silent claim to freeze
the prior UI response. It archives normalized results, not all raw metric inputs
or renderer/evaluator binaries. No production collaboration is claimed.

Passed: root Go race tests/vet, 19 Python checks and 11 Node checks. HTTP fixtures
verify trusted source capture, optional metric failures, offline reopening,
package/question conflicts, invalid selections and original survival after a
failed refresh. Store tests cover process-independent reload, exact JSON integer
preservation, concurrent publication, historical retention and corruption rejection.
Hegel state/swarm passes 100 sequences of up to 25 steps; Quint validates five
observed save/reopen/refresh/reject/reopen transitions and rejects two controls
mutating original integrity or the fixed question. Complete nested verification
suite passes (75.3 seconds); log in ignored
`artifacts/verification/investigations-full-suite.log`. An initial model adapter
parse failure with a parameterless module was corrected by declaring its initial
state constant; the final observed checks pass.

Root and both sample Dockerfile first-stage COPY contexts compile the root CLI
with VCS stamping disabled for temporary sandbox contexts; sample manifests pass.
Full networked multistage image builds are still unverified.

Targeted live journey 19 passed twice (21.2 and 24.5 seconds), including actual
packages-service restart, unchanged receipt reopening, real collector ingestion
of a labelled late test span, and successor retention of the fixed question and
trace selection. The second pass also verifies ratio/member explanation state.
All 19 browser journeys passed in two full runs (3.9 minutes each). The final
run follows the heatmap selection-state correction and includes saved cell-bound
restoration, ratio/member explanations, real late arrival, restart persistence
and a delayed real saved response discarded after a newer package selection.
All 57 documentation frames were regenerated; the three new saved-investigation
frames were visually inspected. A final test-only refinement scopes saved ratio
and comparison explanations to checkout (rather than whichever operation appears
first). Focused journey 19 then passed (16.2 seconds), regenerated its three
frames and asserted both selected stream identities. No application code changed
after the final full 19-journey pass. All 12 default services are running, metric
export is enabled, and the owned traffic generator was restored. No queues or
volumes were cleared. Documentation source links resolve; Mermaid is documented
without a claim of local rendering. General observability gaps and sequencing are recorded in
[PRODUCT_GAPS.md](docs/PRODUCT_GAPS.md), with reviewed public research sources and
explicit limits.

## Live peer and bounded-skew comparisons

Implemented: two actual woven checkout instances with distinct resource identity
(default ports 8090/8092); optional explicit bounded-skew policy; actual per-member
window/count preservation; pairwise endpoint spreads; deterministic conservative
matching without reusing intervals in successful rows. Comparison rows have no
combined percentage; exact fleet totals remain strict. The lab-only export gate
can drop metrics while requests and the independent trace exporter remain enabled.
See [semantics](docs/FLEET_COMPARISONS.md),
[diagram 8](docs/ARCHITECTURE_DIAGRAMS.md#8-two-live-instances-and-bounded-comparisons)
and [journey 17](docs/journeys/README.md#17-two-live-instances-imbalance-loss-and-restart).

Passed: root Go race tests/vet, 19 Python checks, ten Node checks; Hegel comparison
state/swarm: 100 sequences of up to 30 steps; six observed Quint outcomes and three
negative controls rejecting invented common populations. Targeted live journey 17
passed through imbalance, successful requests during metric loss, excessive gap,
and controlled peer restart/new epoch. Four screenshots were generated. The first
attempt exposed a real startup-query/control race (a late 30-minute response
replaced the edited two-minute lookback); controls now preserve edits and stale query
responses are generation-guarded. Source limits were retained. Targeted journeys 17/18 passed together (41 seconds); journey 18 guards
that race using a delayed real response. The full nested verification module passed (60 seconds).
All 18 browser journeys passed in one full run (3.2 minutes), refreshing 54
documentation screenshots. New member-window and loss frames were visually
inspected; updated documentation file links resolve. All twelve services are
running and the peer metric-export gate is enabled. Mermaid source is documented
without claiming a local Mermaid render or production deployment validation.
A focused post-run storyboard check scoped same-timestamp rows to checkout
and asserted the displayed reset cause. Gap/restart checks also require timestamps
from the current injected fault, avoiding matches against older retained telemetry.
The final focused journey17 rerun passed (36.7 seconds), including fresh-fault
timestamp guards and operation-scoped screenshot/assertion selection. No
application code changed after the successful full18-journey run.

## Historical declared fleet membership milestone


Implemented: fixed-window YAML membership for 2–4 instances; bounded concurrent
member queries; separate observed/unobserved/unavailable source states; exact
aligned population totals only when every member supports an interval. Missing,
reset, skewed or incompatible members leave totals unknown. Valid selected-instance
rates survive peer source failure. The default lab intentionally declares an
undeployed peer, not a live second replica. See [semantics](docs/FLEET_COVERAGE.md),
[diagram 7](docs/ARCHITECTURE_DIAGRAMS.md#7-fleet-membership-and-coverage-before-totals)
and [journey 16](docs/journeys/README.md#16-declared-fleet-membership-and-missing-coverage).

Passed: root Go race tests/vet, 19 Python checks, nine Node checks; Hegel fleet
state/swarm: 100 sequences up to 30 steps; Quint six observed aligned/missing/
skew/reset/idle/recovered steps and four rejected invented totals. Race-tested HTTP
fixture verifies peer failure preserves selected-instance rates. Targeted live
journey 16 passed with three generated frames. Dynamic membership, automatic
alignment and live multi-replica aggregation remain unvalidated/unimplemented.
The complete nested verification module also passed (70.8 seconds).
All 16 live browser journeys passed in one full run (2.8 minutes), refreshing
49 documentation screenshots. Fleet coverage frames were visually inspected and
updated documentation file links resolve. The default lab retains eleven services;
no second replica or automatic alignment is claimed.

## Operation error ratios and guided contextual investigation

Implemented: strict YAML ratio policy references an exhaustive calls-rate contract;
exact aligned outcome deltas produce calls/errors/fraction and a display percentage.
Missing/skewed/unsupported intervals remain unknown, supported zero calls are no
data, and full resource identity prevents implicit instance/operation merging.
The runtime explicitly observes zero outcome counters on first operation use.
The UI shows interval charts and guides error-context selection into captured
spans and the existing scoped trace waterfall. It does not detect increases or
causes automatically. See [semantics](docs/OPERATION_ERROR_RATIOS.md),
[diagram 6](docs/ARCHITECTURE_DIAGRAMS.md#6-from-aligned-counters-to-an-error-investigation)
and [journey 15](docs/journeys/README.md#15-from-operation-error-ratios-to-traces).

Passed: root Go race tests/vet, 19 Python checks and eight Node checks. Go fixtures
cover exact sums exceeding uint64, unknown/idle/alignment, invalid members/policies,
bases and a mixed idle/traffic window. Hegel ratio state/swarm checks pass: 100 sequences of up to 30 steps.
Runtime outcome-baseline state/swarm checks also pass 100 sequences. Quint validates
six observed traffic/missing/idle/reset/gap/late-repair steps and rejects four
invented ratios. Targeted live journey 15 passed, generating four readable frames.
Its first attempt correctly found no contextual spans before span ingestion;
the journey now waits for both kinds of real evidence before demonstrating traces.
Counter source cap increased to 3,000 raw points per contract for initialized outcome
streams; 8 MiB and Float64 precision guards remain, histogram caps unchanged.
The full nested verification module also passed (47.8 seconds). All 15 browser
journeys passed in one full run (2.0 minutes), refreshing 46 documentation frames.
Updated documentation file links resolve; ratio and waterfall frames were visually
inspected. All eleven default services, including traffic, are running. Mermaid source/link checks
are performed without claiming a local Mermaid render or production validation.

## Native histogram intervals and synchronized documentation

Implemented: exact uint64 cumulative/delta histogram evaluation, explicit
unit/layout/temporality/resource contracts, unknown reset/decrease/gap/flag/
boundary intervals, late repair and explicit aligned merging. Live checkout
histograms now feed a per-stream bucket heatmap with bucket-bounded P95 and
context-only span navigation. Automatic fleet discovery is not implemented.
See [semantics](docs/NATIVE_HISTOGRAMS.md),
[updated diagram 5](docs/ARCHITECTURE_DIAGRAMS.md#5-native-histogram-intervals-and-bounded-investigation)
and [journey 14](docs/journeys/README.md#14-native-duration-histograms).

Passed: root Go race tests/vet; 19 Python checks; seven Node visualization checks;
CLI fixture count4 and explicit merged count8; Hegel histogram state/swarm
100 sequences and merge properties150 cases; Quint observed restart/gap/repair/
merge validation with reset and incompatible-merge negative controls. All 14 live browser journeys passed in one full run (2.0 minutes), refreshing
42 documentation frames, including three cropped histogram frames. Targeted
journey14 also passed separately. File links resolve; Mermaid source was inspected,
not locally rendered. Documentation, diagrams, storyboard and generated frames
are required alongside feature changes in AGENTS.md.

## Native counter semantics, live rates and architecture diagrams

Implemented: pure Go cumulative/delta monotonic counter evaluation with exact
uint64 deltas, explicit stream identity, reset/epoch/decrease handling, maximum
gaps, missing flags, clipped unknown boundaries, duplicate/conflict checks and
canonical late-input revisions. CLI evaluate-counters runs without ClickHouse.
YAML rates and counter-rates views bind supported operation calls/error outcomes.
The checkout app now exports actual OTel native metrics every two seconds through
the collector; a bounded sum-table adapter validates resource/schema/unit/flags/
temporality and rejects Float64 precision risk. Charts show per-stream supported
intervals and separate counter bases. Linked spans are contextual start-time
matches, not exact contributors or exemplars.

Passed: root Go race tests/vet, exact CLI fixture deltas 3/4, 19 Python checks and
six Node checks. New counter/source fixtures cover uint64 precision, zeros,
resets/decreases, missing points, delta overlaps, conflicting copies/temporality,
identity, edges, late repair and rejected exporter precision loss. Hegel passes
100 state/swarm sequences (up to 30 steps). Quint validates five observed evaluator
steps through restart/gap/late repair, and rejects a rate invented across restart.
Artifacts are in artifacts/verification/counter-*.

The first full browser run had 12 passing journeys and one failure: an old broad
summary selector became ambiguous after adding the counter panel. That selector
was scoped to its intended operation explanation label. Affected journeys 08/11
and native-rate journey 13 then all passed in a targeted rerun (17 seconds).
Thus all thirteen journeys are verified across the two runs; this is not a claim
of a subsequent full-suite rerun. Two new counter frames bring the total to 39.
Eleven services run, traffic is restored and four stale pending entries remain.
Environment restoration initially left ClickHouse/Redis stopped; their existing
containers were started without resetting volumes.

The user-requested subagent prepared four Mermaid architecture/process diagrams
covering build/planning, illustrative production OTel/ClickStack topology, captured
evidence investigation and the separate Databricks compiler boundary. Links and
fences were checked; Mermaid syntax was inspected but not locally rendered.
These diagrams are documentation, not production deployment evidence.

Running images use offline-wheel builds plus host-built static binaries. Cumulative
metrics are verified live; delta semantics are fixture-tested. Native histograms,
fleet aggregation, error ratios, metric exemplars, native metric Databricks lowering,
immutable custody and production automation remain outside this slice. Updated
startup instructions are saved as a draft, not published. See
[native counters](docs/NATIVE_COUNTERS.md) and
[architecture diagrams](docs/ARCHITECTURE_DIAGRAMS.md).

## Declarative YAML operation packages and second application

Implemented: strict versioned YAML operation packages with source adapter/resource/
instrumentation bindings, exact source operation/event contracts, full subject
identity, enabled built-in views and explicit unverified evidence assumptions.
`goo plan-operations` and the investigation service share the same planner. Plans
carry content revisions that contribute to captured evidence bases; mutation is
rejected before fetching. Repeatable operation-package flags replace the former
source-directory flag and sample-specific Go binding. The UI selects packages,
honors enabled views, clears previous evidence and discards stale responses.

A separate woven Go report generator now uses its own service/instance, source
manifest, report.generate operation and report namespace. Its YAML omits the
heatmap and reuses the same query/UI code. Both packages run in default Compose;
there is no application-name branch in the operation renderer or source adapter.

Passed: all twelve live Playwright journeys (1.2 minutes), including report
creation → YAML-backed view → report subject and deliberate delayed delivery of
a real checkout response during a switch to reports. No telemetry response data
was mocked. Five new documentation frames bring the total to 37. Eleven services
run, demo traffic was restored and three stale pending entries retained.

Root Go race tests/vet, 19 Python checks and six Node checks pass. CLI tests plan
both checked-in packages. Source/plan tests cover stale roles/versions, weakened
identity, unknown views/fields, duplicate keys, aliases and multi-doc YAML, revision
determinism/nonmutation, resource provenance and changed-plan rejection. Hegel
passes 150 new planner cases for typed resource values, revision sensitivity and
nonmutation. Existing 100-case heatmap and 100-case subject properties plus the
five-action Quint navigation checks/negative controls remain passing.

The report helper actually validates and builds a static woven executable. Running
images use the established offline build with host-built Go binaries; the new
networked report Dockerfile was not executed here. Install/start additions are
saved in the cloud environment draft, not published. No root dependency was added.

This is a local-source, fixed-adapter/view YAML package slice. General package
resolution, source-free manifests, arbitrary expressions/CEL/KCL, IaC apply,
native metric reset/histogram policy and operation Databricks lowering remain
unimplemented. See [package specification and executable onboarding](docs/OPERATION_PACKAGES.md).

## Scoped subject event navigation

Implemented: optional runtime subject namespace/tenant context, declared-event
normalization from actual ClickHouse nested event arrays, and trace → subject
events → related event-bearing invocations across traces in the captured evidence.
The key retains environment/service/namespace/tenant/ID; missing or mismatched
identity stays unlinked with a diagnostic. The sample now generates fresh default
IDs and can explicitly reuse an ID in either demo tenant, fixing its former
constant-subject collision. No subject dimensions were added to metrics.

Passed: all ten real browser journeys (1.2 minutes). The new live journey submits
two invocations with one scoped subject, plus an identical ID under a second
tenant. It confirms four events across two related traces, excludes the other
tenant, retains the original basis and performs no additional operation request.
Three new frames bring the storyboard total to 32. All ten services run; demo
traffic is restored and three pre-existing stale pending entries were retained.

Go race tests/vet, 19 Python checks and six Node checks pass. The actual OTel SDK
scope test passes, as do 100 new Hegel scope-isolation cases, the existing 100
heatmap cases and 200 observation-copy cases. The five-action Quint kernel fixture
passes conformance/safety; both wrong-cell and cross-tenant membership mutations
are rejected. Observed steps and ITF are retained under artifacts/verification/
subject-navigation*. Existing actual woven SDK/HTTP preparation checks and Quint
negative control remain passing. Root runtime dependencies remain unchanged.

Returned event arrays/timestamps are validated against their parent spans. The
relation joins only event-bearing invocations, not failed calls inferred from
absence. Reported terminal outcomes do not establish retries, transaction commit,
durability or complete subject lifecycles. Query scope is still one explicit local
resource/library/span-start window with mutable late-data reruns. Updated startup
instructions are saved as a draft, not published. See [subject storyboard](docs/SUBJECT_NAVIGATION.md).

## Operation duration heatmaps and trace drilldown

Implemented: the Go source response retains deduplicated logical span evidence
with exact string start/duration nanoseconds and validated parent IDs. Operation,
calls, error and panic selections feed a raw-duration heatmap and span inspector.
Cells inspect exact contributors; opening a trace expands to the other operation
spans in that trace within the same resource/window/basis. The existing waterfall
renderer is reused with a separate details target. The Go sample now includes a
nested stock-check operation with bounded simulated success/error work.

Passed: nine real Playwright journeys (about one minute), including live failed
checkout → reserve error selection → duration cell → contributor → parent/child
waterfall. The new journey independently checks membership and contributor count,
changes the lookback control during exploration, and verifies a retained basis,
original bounds and no additional operation request. Three new frames bring the
storyboard total to 29. Existing flows remain passing; traffic was restored and
three stale pending entries preserved.

Go race tests/vet, 19 Python checks and five Node checks pass. Hegel passes 100
new generated cases exercising the actual browser kernel against a Go reference,
plus the existing 200-case operation-copy property. A four-action observed
navigation fixture passes quintgo/Quint conformance and safety; an incorrect cell
membership mutation is rejected. This models the actual selection kernel over a
bounded fixture, not browser DOM behavior or source completeness. Reports are in
artifacts/verification/operation-navigation*.txt. Source tests additionally check
logical evidence precision, parent validation and deduplicated inspector records.

The updated Go binaries and UI run in the established offline-built images.
Trace scope stays limited to observed operation spans in one selected resource,
library and span-start window. Sampling/completeness remain unknown, late-data
refreshes can change results, and no immutable custody or full distributed trace
fetch is claimed. See [the executable storyboard](docs/OPERATION_DRILLDOWN.md).

## Live woven Go operation integration

Implemented: separate Go checkout sample built with actual Orchestrion weaving,
standard OTLP/HTTP trace export through the collector into ClickHouse, explicit
resource/library binding, and a Go operation endpoint behind the existing query
proxy. The investigation panel lists declared operations, evidence status,
completed invocations, errors, panics and duration sample counts, with source,
window, manifest revision, evidence basis, typed SQL and duplicate explanation.
Unobserved declarations display unknown counts.

Passed: all eight real Playwright journeys (53.8 seconds), including actual sample
HTTP 200/503/500 paths and observed compatible operation rows. Three new captured
frames bring the storyboard total to 26. Ten default Compose services run; traffic
was restored and three pre-existing stale pending entries retained. Go race tests
and vet pass; 19 Python checks and three Node geometry tests pass. Hegel passes
200 generated cases checking invocation totals under duplicate copies and reversed
input order. Source-adapter tests cover typed values, exact bounds, resource/kind/
scope rejection, conflicting copies, deduplication, raw overflow and HTTP decoding.
The lint-before-build sample helper was executed successfully. Existing actual
woven SDK conformance and HTTP preparation/Quint replay plus negative control pass.

The running images use offline Python wheels and host-built static binaries.
The new networked sample Dockerfile was not executed. A read-only buildx cache
path was corrected with BUILDX_CONFIG=/workspace/.cache/buildx. A Compose name
conflict during replacement was recovered by recreating only query/packages and
the new sample; ClickHouse/Redis volumes were preserved. Updated reusable install
and start instructions were saved to the environment draft, not published.

Schema compatibility is declared because the exporter drops scope schema URLs.
This sample exports traces; native metric export/rate/reset/histogram semantics
remain separate. Coverage, sampling, completeness, immutable custody and durable
business subject semantics remain unknown. See [live operation storyboard](docs/LIVE_OPERATIONS.md).

## Observed operation conformance

Implemented source-manifest matching against normalized completed operation
spans, an invocation overview and `goo observe-operations`. Root race tests pass.
The actual woven SDK test also passes and checks three completed spans against
the overview (two reserve calls including one error, one panic). It exports
artifacts/annotations/operations.json for the executable CLI journey. Negative
checks reject cross-resource evidence, mismatched scope/version/schema,
undeclared roles, invalid outcomes/IDs and conflicting duplicates; identical
copies count once. Exact duration precision and unobserved declarations are tested.
Existing woven HTTP/Quint replay and negative control still pass.

This is SDK evidence, not collector/ClickHouse/browser integration. Native metric
rates and histogram merging, live operation discovery and business aggregation
remain unimplemented. See [the operation journey](docs/OPERATION_CONFORMANCE.md).

## Goo operation/event annotation slice

Implemented: bounded operation/event directives, deterministic source-contract
manifests, isolated OTel runtime hooks and test-only lint-before-weave CLI.
The normal runtime compiler module has no new OTel dependency. The default
Dockerfile now includes the annotation parser in its Go build stage; no running
service or UI was changed by this opt-in developer slice.

Passed: root race tests/vet; actual Orchestrion weaving of the example with real
OTel SDK export. Checks cover success/error/panic application behavior, three
completed-invocation/duration observations, zero leaked active invocations,
two subject events, three diagnosed event drops, and manifest scope/version/unit
agreement. The existing woven HTTP preparer/Quint test and dropped-fact negative
control remain passing. Telemetry is retained in artifacts/annotations/otel.json;
manifest in artifacts/annotation-manifest.json.

The native Hegel state/swarm check passes with 100 requested sequences of up to
30 steps, checking actual SDK active counts after start, successful/error/panic
completion and repeated finalization. The developer journey uses in-memory OTel
export, not live collector ingestion or a new browser screen. Full build manifests,
automatic inventory activation, operation aggregate lowering and production
weaving deployment remain next steps. See [the executable annotation storyboard](docs/GOO_ANNOTATIONS.md).

## Prior live HTTP source, shared view and opt-in weaving

Passed in the current workspace:

- Go race tests/vet and 19 Python checks, plus the shared package CLI/reference/
  embedded ClickHouse parity storyboard and three Node geometry checks.
- Seven real browser journeys (47.8 seconds), capturing 23 frames. The new
  journey creates actual successful and failing HTTP responses, waits for OTel
  ingestion, renders the Go view descriptor and inspects policy/source/revision.
- Source adapter unit checks for typed parameters, exact half-open bounds,
  valid OTel IDs, deduplication, conflicting copies and partial-scope rejection.
- Hegel: 600 requested cases across discovery, HTTP policy and source dedup.
- Real Orchestrion weaving of `Expansion.PrepareHTTP` into quintgo recording;
  its retained-count trace passes Quint and a deliberate dropped-fact trace is
  rejected. Positive and negative artifacts are kept separately.

The isolated weaving module pins x/tools 0.52.0: the initially selected 0.49.0
could not read Go 1.27 export data. The updated tooling passes; the root runtime
module is unaffected. The initial module download denial was resolved through
the official GitHub source with Go checksum verification. No network policy was
changed to complete the check.

The actual running images use the established offline-wheel build plus a
host-built static Go executable. The new networked multi-stage Dockerfile was
not executed here. Default Compose now includes the Go packages service; all
nine services run and traffic is restored after journeys. Reusable setup draft
instructions were saved, not published.

Schema compatibility is declared by library binding because this exporter drops
scope schema URLs. Sampling/completeness, immutable custody snapshots, generic
instrumentation auto-discovery, prepared-table materialization and Databricks
execution remain unverified/unimplemented. Woven model observations are tested
in the optional module, not enabled in the running service. See
[LIVE_HTTP_PACKAGE.md](docs/LIVE_HTTP_PACKAGE.md) and
[INSTRUMENTATION.md](docs/INSTRUMENTATION.md).

## Prior HTTP binding and expansion slice

Implemented: explicit HTTP server-span binding into the existing shared Go
measurement model, aggregate view/recipe descriptors, both backend plans and
normalized server-request preparation. Content revisions cover policy and emitted
artifacts. No current UI or raw-exporter adapter integration is claimed.

Passed: root Go race tests and vet; package CLI storyboard with two independently
specified route groups matching both reference evaluation and generated embedded
ClickHouse SQL; generated Databricks YAML structural checks. Unit checks reject
unsupported profiles, missing policy/source contracts, resource mismatches,
client spans, duplicate identities and mutated expansion content. Hegel adds
250 requested policy/status cases alongside the existing 200 discovery cases.

Databricks execution/deployment, raw-source SQL preparation, duration operators,
objective/alert configuration and browser rendering of these descriptors remain
unimplemented. Existing Quint lifecycle models were not extended or rerun for
this stateless binding/classification slice. See [the executable CLI storyboard](docs/PACKAGE_EXPANSION.md).

## Prior semantic packages and worker/queue adapter

Current additions on 2026-10-17:

- Ten versioned Go contract profiles and per-resource-scope capability proposals.
  CLI catalog/discover examples run; unknown versions/units and co-observed
  dimension requirements are tested. Generated UI catalog matches the Go source.
- Root Go race tests and vet pass. New Hegel property passes with 200 requested
  generated cases across profiles: absent/mis-versioned/mis-unit evidence cannot
  be borrowed from another scope; inventory ordering and duplication do not
  change proposals. Native verification stays in its separate module.
- Nineteen Python API/SQL/adapter tests pass, including exception-safe bounded
  worker counts, unknown broker lag, clock violations, per-instance gauge
  snapshots and raw-observation wait percentiles. Three Node geometry tests pass.
- Six real Chromium journeys pass (45.1 seconds), capturing 21 documentation
  frames. New coverage observes active work through actual Redis → OTel →
  ClickHouse ingestion, checks all queue/worker metrics and opens the package
  catalog. The final catalog-only update is also checked with the queue/library
  journey. Demo traffic is restored by the wrapper.

The new catalog is metadata recognition/proposal generation, not automatic live
inventory extraction, source/value-level binding, executable expansion of all
profiles, or Databricks deployment. The reference queue view is implemented;
most other profile outputs remain explicitly proposed. Versioned examples are
illustrative declarations, not live-scan evidence. No new scheduler or custody
model was introduced; existing Quint lifecycle models were not extended to claim
queue-adapter or package-expansion verification.

Queue observations distinguish undelivered lag, pending (including stale work),
local active/capacity and first-start waiting time. No stream reset was used;
the current wrapper reports three stale pending entries, up from two before the
consumer update. The sample still cannot reclaim crash/interrupted work.
See [semantic packages](docs/SEMANTIC_PACKAGES.md),
[queue contracts](docs/QUEUE_TELEMETRY.md), and [journeys](docs/journeys/README.md).

## Prior visualization and guided-investigation slice

Passed against the current local stack on 2026-10-17:

- Fifteen Python API/validation/embedded ClickHouse tests. New checks cover
  duration-band boundaries, zero and UInt64-maximum durations, exclusive time
  bounds, invalid selections, and 1101-span heatmap totals versus 200-row inspection.
- Three Node geometry checks cover exact nanosecond offsets, scoped parent
  identities, missing/cyclic parents, and zero-duration/empty spans.
- Five live Chromium journeys pass (41.7 seconds), capturing nineteen storyboard
  frames. Added paths exercise heatmap → exact raw spans → reservation waterfall
  → async producer link, and guided retry/USE/RED navigation with scope checks.
- Inspected actual waterfall and USE screenshots; Compose configuration and
  JavaScript/Python syntax checks pass. Demo traffic is restored after the run.

The first async navigation run correctly exposed an omitted producer span: its
sample instrumentation lacked reservation identity. New producer spans now carry
that identity and related attributes; the passing journey follows the exact
linked span. Older data is not backfilled and external links remain labelled.
The heatmap's minute conversion was also checked against both embedded and live
ClickHouse rather than relying on assumed timestamp function return types.

These visualizations remain in the Python reference lab, not a Go/Rust API
migration. USE utilization/queue measurements and a complete RED error-rate
measure are explicitly absent. Historical custody bases and Go lifecycle receipt
integration remain unimplemented in the UI. Existing Hegel/Quint semantics were
not changed or rerun for this presentation/raw-span slice.
See [visualization contracts](docs/VISUALIZATIONS.md),
[guided investigations](docs/INVESTIGATION_WORKFLOWS.md), and
[captured journeys](docs/journeys/README.md).

## Prior lifecycle, journeys and verification

Passed in the current cloud workspace:

- Go race tests and vet, plus all twelve existing Python tests.
- Thirteen independent lifecycle scenarios and 26 total/grouped aggregate SQL
  comparisons in embedded ClickHouse, including old/new snapshot coexistence.
- Three Hegel properties: generated deadline/dedup cases across epoch ranges,
  historical evidence identity under invisible appends, and state/swarm sequences
  with invariants after every enabled step (300/200/150 requested cases).
- Quint 0.32.0 bounded simulation: 100 traces, maximum 30 steps, seed 42.
- Eight Quint-generated MBT traces replayed through the actual Go preparer:
  160 steps with state comparisons. An eight-step observed Go trace conforms
  and passes `safety`; an impossible good count is correctly rejected.
- Three real Chromium/Playwright journeys pass after workload isolation,
  capturing twelve documentation frames. The sample's good, retry, late, missing
  and rejected scenarios traverse the real services and investigation UI.

Observed and addressed: restored ClickHouse/Redis processes needed startup;
recreated application containers needed local-service proxy bypasses; an initial
live journey under background load missed the deadline; the documentation runner
now pauses/drains/restores the owned generator without deleting stream data.
Pre-existing stale pending messages remain unreclaimed. Screenshot review also
found and fixed stale view controls after expression edits and an inappropriate
metric-point disclaimer on trace/log/latency results.

The normalized-event kernel is tested in Go; generated SQL **preparation** from
raw exporter data is not implemented. The Go semantics are not yet connected to
the live UI. Browser coverage does not resolve the Python accounting defects.
Databricks deployment/query/ingestion and optional HyperDX remain unverified.
See [journey frames](docs/journeys/README.md), [lifecycle contracts](docs/LIFECYCLE_PREPARATION.md),
[verification commands](verification/README.md), and [the bitemporal spike](docs/BITEMPORALITY.md).

## Portable Go aggregate pilot

The follow-on pilot is documented in [PORTABLE_COMPILER.md](docs/PORTABLE_COMPILER.md).
Passed in this workspace:

- `go test -race ./...` and `go vet ./...` using Go 1.27.2.
- `python scripts/check_portable.py`: ten total/group fixture comparisons across
  independent expected results, the Go evaluator, and generated embedded
  ClickHouse SQL; an additional declarative-count/escaped-literal case passes.
- Structural parsing of generated Databricks YAML and CLI output contracts.
- The runnable CLI example returns two eligible subjects and success 0.5.
- All original twelve Python unit/API/embedded ClickHouse tests still pass.

The docs URL was retried and remains blocked by proxy 403; the pinned official
Databricks creation reference was re-read from accessible GitHub content.
Databricks DDL/query execution, Unity Catalog deployment and physical ingestion
remain unrun. Normalized lifecycle preparation is now tested as described above;
real exporter adaptation and SQL preparation remain pending. The Go pilots do
not resolve the Python prototype's existing measurement defects.

## Original handoff record

The current lifecycle and journey validation is recorded separately in
[LIFECYCLE_PREPARATION.md](docs/LIFECYCLE_PREPARATION.md),
[the journey storyboard](docs/journeys/README.md), and
[verification/README.md](verification/README.md). These additions supersede the
original environment's unverified browser/default-stack status, while preserving
the distinction between the Python prototype and new Go semantics.

Created 2026-10-16.

Passed:

- Docker Compose 2.35.1 `config --quiet` against the complete manifest.
- Twelve Python unit/API tests, including all five query modes executed in embedded ClickHouse 24.8 through chdb 3.3.0.
- Application flow test: real HTTP and gRPC processes, fake Redis TCP/Streams, real consumer processing, and a local protobuf OTLP receiver. Verified accepted/retry/terminal log records, gRPC server spans, asynchronous span links, and exported metrics.
- Python compile checks and JavaScript syntax validation.
- OTel SDK/exporter construction with pinned Python dependencies.

Not verified in this workspace:

- Full Compose startup, collector configuration acceptance and ingestion into the pinned ClickHouse 25.3 container, image pulls, optional HyperDX startup, and `scripts/smoke.py` against that stack.
- Browser rendering. The Playwright browser download was unavailable; only JavaScript syntax was checked.

Docker 28.0.4 and Compose 2.35.1 binaries were installed for validation. The hosted workspace denies mount, Unix-socket, and user-namespace operations needed by the Docker daemon; rootless setup is also unavailable. Installing binaries cannot supply those host permissions. Run the provided full-stack smoke test on a Docker-capable machine before treating the container integration as verified.

The first application-flow run exposed inherited trace sampling that discarded demo spans. The SDK now explicitly retains all sample traces. A subsequent integration run verified the emitted spans and links.

## Codex handoff packaging

The handoff adds repository guidance, a continuation task, architecture, two proposed RFCs, an implementation plan, known issues, source references, and a portable snapshot of the original design Page. No application behavior changed during this packaging step.

The twelve unit/API/embedded ClickHouse tests and Compose configuration validation were rerun successfully. Local Markdown links and fenced code blocks were checked. The archive excludes caches, environments, downloaded runtime binaries, and database volumes, and includes SHA256SUMS for its source and documentation files. Full container startup and browser rendering remain unverified as noted above.

## Log and trace discovery

Implemented readable literal log search, independently scoped surrounding logs,
matching reported trace-ID groups, fixed full-population span-duration buckets and
available-field suggestions feeding reusable packages. See the
[discovery contract](docs/EXPLORATION_DESIGN.md#log-discovery-time-context-and-span-duration-buckets),
[journey 30](docs/journeys/README.md#30-from-log-search-to-reusable-span-distributions)
and the current detailed results in [ENVIRONMENT_VALIDATION.md](ENVIRONMENT_VALIDATION.md).
This is a captured-record investigation workflow; it does not establish complete
traces, root latency, live tail or production readiness.

Final inventory milestone: default browser run passed 29 journeys (two opt-in
skips); explicit fault, generated-success and final-image inventory campaigns
passed together (3.3 minutes). All 31 distinct journeys passed across these runs.
The final inventory actor recorded 685 steps with all required stages and rejecting
Quint controls. All 95 PNG documentation frames and the generated-campaign JPEG frame were regenerated; see
[the cloud record](ENVIRONMENT_VALIDATION.md#2026-10-18--observed-service-inventory).
