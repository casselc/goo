# Cloud baseline validation — 2026-10-16

This is the original onboarding record. Subsequent live browser journeys,
lifecycle preparation and Hegel/Quint evidence are recorded in
[VALIDATION.md](VALIDATION.md); they supersede the browser-unverified status below.

The archive was verified against its bundled SHA256SUMS and imported at the root
of the initially empty goo checkout. Files remain uncommitted. Application logic
and dependency declarations were preserved. PROJECT_DIRECTION.md records the
user's Go/Rust and YAML+CEL/KCL preferences and sibling-repository references.
The original VALIDATION.md remains historical handoff evidence.

Verified in this cloud instance:

- Python 3.12 virtual environment at /workspace/.venvs/goo; requirements-dev.txt installed.
- All 12 unit/API/embedded chDB tests pass, including a repeated installation run.
- Docker-free real HTTP/gRPC/consumer/protobuf OTLP flow passes, including a repeat
  with the Compose web/traffic services stopped to avoid the shared port 8081.
- Docker Compose configuration validates; default ClickHouse 25.3, Redis 7.4,
  collector 0.123.0, and five application roles start.
- Full scripts/smoke.py passes: successful mature completion plus nonempty logs,
  traces, latency, and metric query results through actual collector ingestion.
- HTTP health and UI HTML return successfully. Browser rendering is unverified.
- Actual ClickHouse schema captured at /workspace/cloud-setup/goo-schema.tsv.
- Go 1.27.2 downloaded with its official SHA256 verified; the related
  observability/otel-chdb/query `go test -p 4 ./...` suite passes. This does not
  establish external-service integration readiness or a full observability setup.

Cloud setup corrections:

- Docker configuration uses /workspace/.docker-goo because the default home
  configuration location is read-only.
- Docker builds cannot resolve their package/proxy destination. Host pip downloads
  verified wheels through normal HTTPS; an external generated Docker build context
  installs those wheels offline. The original Dockerfile remains unchanged.
- Imported collector config was mode 0600; made it readable for the image's
  non-root collector. No configuration contents changed.
- Standard Go proxy redirects to storage.googleapis.com were denied. A successful
  GOPROXY=direct download for three affected dependencies preserved checksum checks.
- An initial smoke request preceded collector readiness; a subsequent smoke passed.
  A flow-test repeat run while Compose owned port 8081 failed; stopping only the
  owned web/traffic services and repeating restored the passing result.

Reusable install_script and start_skill were saved to the environment draft.
Saving does not publish the snapshot or prove restoration in a new task.

Not performed: browser rendering/scenario walkthrough, optional HyperDX, a Rust
installation, Jolt workflow installation, full sibling-repository integration,
or the language/measurement implementation backlog. Existing measurement issues
in docs/KNOWN_ISSUES.md remain; successful baseline tests do not resolve them.


## Native operation error ratios

The current checkout image includes explicit SDK zero initialization for all three
outcomes on first operation use. Rebuild the woven app with
`scripts/build_operation_app.sh` after runtime changes; rebuilding only `goo` does
not update the app's instrumentation. The query image must include `ui/ratios.js`
and the package image must include the current checkout YAML and Go executable.
No dependency, service or port was added. The counter source cap is now 3,000
points per contract, with existing precision/body guards retained.

After snapshot restoration this session found ClickHouse/Redis stopped and the
collector/consumer restarting. Existing services were restored with Compose,
without volume resets. Five pre-existing stale pending entries were retained by
the journey wrapper. See VALIDATION.md for actual suite results and
[ratio semantics](docs/OPERATION_ERROR_RATIOS.md) for supported behavior.


## Declared fleet coverage and repository milestones

The default checkout package now declares `woven-demo` and expected `woven-peer`.
Only woven-demo is deployed by the default lab; missing-peer coverage is deliberate.
The counter endpoint concurrently queries at most four explicitly bound instances,
keeps selected-instance results when a peer fails, and returns `fleet` coverage.
Refresh `ui/fleet.js`, checkout YAML and the Go executable when rebuilding the
retained query/packages images. No new service, port or dependency was added.
See [fleet semantics](docs/FLEET_COVERAGE.md) for fixed-window membership and
exact-alignment limits. The complete verification module passed; the targeted live
fleet journey generated three frames.

The initially unborn goo checkout now has its first committed milestone on `work`.
Older onboarding notes about leaving all imported files untracked are superseded.
Commit validated milestones and report their hashes as required by AGENTS.md.


## Two-instance live comparison milestone

The default stack now has twelve services. `operation-peer` runs the same woven
binary with GOO_INSTANCE_ID=woven-peer and loopback8092 mapped to internal8090.
The original instance defaults to woven-demo. Metric collection remains every2s;
peer exports can be paused/resumed through its clearly labelled lab controls.
The browser restart journey manages only the owned local Compose peer and restores
metric export. Existing queue entries/volumes are retained.

When rebuilding the offline image, preserve the woven operation-app binary and tag
`grammar-observability-operation-peer` as well as operation-app/query/packages.
Stage current ui/fleet.js, ui/index.html, checkout YAML and sample source before
building. `docker compose up -d --no-build` includes the peer. No new dependency
or credential is needed. Exact totals remain strict; comparison tolerance does
not shift timestamps or invent a common population. See FLEET_COMPARISONS.md.

## Saved investigation storage

The packages service now mounts the dedicated Compose `investigations` named
volume at `/app/artifacts/investigations`; the Go service owns bounded capture,
immutable publication and read-time verification. Default lab service count stays
12. Journey 19 restarts only the local packages service and verifies receipt
persistence and reopening without source reads. The collector receives a labelled
late test span for fixed-window refresh validation; existing queues/volumes remain.
Host-staged builds include the updated `goo` binary, Python proxy, `ui/index.html`,
`ui/operations.js`, `ui/ratios.js`, `ui/fleet.js` and `ui/investigations.js`.
Only query/packages are recreated for this change. The root and both woven-sample Dockerfiles now
copy the new Go `investigations` package because each compiles the root CLI; their networked multistage builds remain
unverified here, while the offline staged-image build succeeds.

Isolated first-stage COPY contexts for all three Dockerfiles compile the root CLI
successfully with `go build -buildvcs=false ./cmd/goo`. VCS stamping was disabled
for these temporary contexts because the sandbox's ancestor Git metadata caused
an unrelated status failure. Both sample manifests also build from the real tree.
These checks establish local compilation inputs, not a full networked Docker build
or an Orchestrion toolchain rebuild for this milestone.

## General exploration and context expansion

The Go packages service adds POST `/explore`; the Python query service forwards
bounded JSON at `/api/explore` and serves `ui/exploration.js`. No service or port
is added: the default lab remains 12 services. Only query/packages are rebuilt
and recreated; existing telemetry/storage/queues remain. All three Dockerfiles
copy the new Go `exploration` package because their builds compile the root CLI.
The host-staged image includes updated CLI, proxy, index/operation UI and new
exploration UI. The shared CLI supports `plan-exploration -query <JSON file>`.

The live source uses the actual pinned exporter schema. Direct read-only
ClickHouse checks establish exact round trips for backslashes, tabs, newlines,
Unicode and DEL using the adapter's Escaped-format parameter encoding. Journeys
20–22 cover labelled collector fixture populations, real web/gRPC traces/logs and
independent live expansion from retained operation receipts. Full networked
multistage builds and optional HyperDX/Databricks runtimes remain unverified.

## Product workspace and distribution comparison

No new services, dependencies or ports. The query image must include `ui/product.css`,
`ui/workspace.js` and `ui/distributions.js` alongside the existing UI assets. Rebuild
the Go packages binary for `/compare`; the server proxy permits bounded comparison
bodies up to 1 MiB. General exploration baselines live only in browser memory.
The named capture volume continues to store operation receipts, not those baselines.
Host-staged offline image builds and the existing managed Docker daemon were used;
no daemon, credentials, volumes or queues were replaced. See current [checks](VALIDATION.md).

## Ad hoc telemetry packages

The existing Go packages service now serves `/telemetry-packages` authoring/list/
lookup and pure draft/plan routes. Saved definitions live in the existing named
investigations volume's `telemetry-packages` subdirectory, separate from evidence
receipts. Query images include `ui/telemetry-packages.js` and updated UI/proxy assets.
No new dependencies, services, ports or application annotations are required.

Repeated VFS image builds exhausted the 32 GiB workspace during query-service
recreation. Read-only disk diagnostics confirmed full bytes, not inodes. Removing
dangling unused images and unused builder cache recovered roughly 19 GiB free;
running containers and all data volumes were preserved, and recreation succeeded.
No Docker daemon/authentication configuration or storage driver was changed.

## Bombadil browser spike (2026-10-18)

Pinned npm development tool `@antithesishq/bombadil` 0.7.9 and exercised
`verification/bombadil/workspace.ts` against the running UI at localhost:8088
with system Chromium, headless and the container's required `--no-sandbox`.
The initial 20-second run passed. The final 15-second run passed with 160
generated navigation actions, a bounded shell-readiness property, panel/navigation
agreement, and uncaught-exception/unhandled-rejection properties. A temporary
always-false property exited 2 with a reported violation (negative control); the
temporary specification was removed. Runtime traces/screenshots remain ignored
under `artifacts/bombadil/`. No product runtime or screenshots changed.

The cross-app coordinator and optional Quint MBT/trace bridge are documented
proposals, not executed integration claims. No existing Go/Python semantic code
changed, so their suites were not repeated for this browser-only spike.

## Coordinated cross-app journey and optional Quint validation (2026-10-18)

Implemented the next slice of that proposal in `journeys/cross-app.spec.js`.
`GOO_CROSS_APP_QUINT=1 npm run test:cross-app` passed in 35.9 seconds (37.6 seconds
including Playwright startup). A real report-app UI request emitted one captured
span, identified through its actual report subject. The direct read-only ClickHouse
test adapter matched the rendered span IDs/count over the exact UI query window;
the trace inspector rendered that span. A concurrent, independently profiled
Bombadil process completed 270 generated navigation actions with exit code 0.
No responses, telemetry or UI records were mocked; no data was cleared.

The run's accepted/source_ready/displayed/explored ledger passed qobs/quint-go
conformance and the Quint safety invariant. Native steps, report and ITF were
retained beside `ledger.json`, source SQL/results and Bombadil artifacts under
ignored Playwright output. Two real screenshots were generated and the goo frame
visually reviewed in `docs/journeys/screenshots/26*`; architecture diagram 13, the
storyboard, testing design and implementation plan describe the implemented scope.

The labelled protocol fixture and all twelve rejecting controls passed in 23.958
seconds: trace/subject/scope/window/count/membership disagreement, premature display,
unexercised/failed exploration, mixed runs, truncated and empty ledgers. Quint
typechecking, JavaScript syntax and whitespace checks passed. `go vet ./...` in
`verification` initially found a pre-existing unkeyed histogram fixture literal;
the equivalent keyed literal fixes that warning. Vet and the affected Hegel
histogram merge property then passed. Unchanged runtime suites and the full
existing browser suite were not repeated for this test-only increment.

The harness strips ANSI codes when reading Bombadil's progress, removes inherited
Playwright forced-color output before the optional Quint subprocess (quint-go
parses plain verdicts), and launches the pinned native Bombadil binary so SIGINT
cancellation reaches its cleanup handler. Application-side generated exploration,
model-generated browser actions, retries and fault campaigns remain future work.

## Generated browser campaigns and two Bombadil actors (2026-10-18)

Implemented `npm run test:generated-campaign`: quint-go `connect.Generate` creates
a seeded nine-checkpoint schedule; Bombadil submits three real report requests
in fresh profiles while another Bombadil browser explores goo navigation. The
Playwright observer performs three inspections and two generated revisits through
normal UI controls, comparing actual rendered members/counts with independent
read-only ClickHouse facts over each exact frozen question. `connect.Driver`
then validates the observed handler effects against the generated ITF poststates.

The original journey 26 passed after extracting its source/UI checks into the
shared test adapter (37.6 seconds). The generated journey passed with seed 99,
including all five optional qobs/quint-go inspection checks; the two-test run took
2.0 minutes. A final archive-validation run with seed 2026 also passed in about
1.3 minutes: three captured result identities, nine completed checkpoints, five
source/UI comparisons, five passing qobs reports and 526 goo navigation actions.
No application outcomes, telemetry records or UI responses were mocked.

Eight model traces generated with seed 2026 covered six distinct complete schedules
and passed projection replay. The generator's negative controls reject changed
counts/targets, reordered steps and truncation. Three Node artifact-reader tests
passed, including rejection of absent/duplicate submissions and unstable result
identity. Quint typechecking, verification-module vet, JavaScript syntax, Markdown
local links and whitespace checks passed. Unchanged runtime suites and the full
browser suite were not repeated for this test-only increment.

An initial exploratory run was interrupted after exposing the driver's incorrect
assumption that navigating to the same hash resets an open inspector. The shared
driver now closes that inspector and clears the prior filters through actual UI
controls before a fresh selection; both live validations above exercise revisits.

The generated campaign is explicitly opt-in; its normal-suite skip was verified.
Its archive now lives at `artifacts/bombadil/campaigns/<run-id>/`, outside normal
Playwright cleanup. After the final passing run, another ordinary Playwright
invocation skipped the campaign and the complete archive, all four actors' traces
and all five inspection reports remained present. The source app screenshot is
copied unchanged from Bombadil's own capture. Journey 27, three new documentary
frames, refreshed journey 26 frames, diagram 14 and design/planning/issue docs were
updated. This campaign remains bounded to successful reports and fresh evidence
revisits; retry/fault generation and long-lived independent actors remain future work.

## Generated retry/rejection/stalled evidence campaign (2026-10-18)

Implemented `npm run test:fault-campaign` over the unchanged Python HTTP/gRPC/Redis
sample, using the existing generated nine-checkpoint schedule and fresh Bombadil
app actors alongside concurrent goo navigation. Each inspection checks ordinary
Explore span/log records, full bounded counts and attributes against independent
ClickHouse reads over the exact UI questions. The observer also opens the real
attribute disclosure and checks its terminal outcome/scenario, then closes it so
the documentary frame shows the full captured log sequence.

The first live run with seed 2026 passed in about 1.4 minutes: three accepted IDs,
nine checkpoints, five evidence inspections, five qobs/quint-go fault-model reports
and 503 goo navigation actions. After adding an explicit terminal-presence flag
to distinguish an absent terminal from a present empty outcome, the final seed-7
campaign also passed in about 1.4 minutes, with all five fault checks and 420
navigation actions. Both wrappers restored the owned traffic generator and
retained the five pre-existing stale pending stream entries; no streams/volumes
were cleared and no services were rebuilt.

Captured retry evidence contained two consumer attempts, one error span, retry
and confirmed terminal logs. Business rejection had a rejected terminal and no
error span. Stalled work had a positive stalled log and no terminal, preserving
`unknown-no-terminal`. These are test interpretations of raw displayed records,
not a new product health result, complete-capture claim or business durability
guarantee. The source readiness gate cannot pass on empty capture.

The final Quint protocol checks passed all three positive fixtures and twelve
rejecting controls (healthy label, changed fingerprint, invented terminal value,
wrong terminal-presence flag per case). The actual JavaScript projection passed
50 Hegel state/swarm sequences of up to 20 steps, checking duplicate-log and order
invariance, absence semantics and withdrawal of certainty for conflicting terminal
outcomes. The combined model/property run passed in 41.047 seconds; the Hegel
property portion took 1.64 seconds. Verification-module vet, JavaScript syntax,
local documentation links and whitespace checks passed.

Focused real-browser rechecks over the final captured IDs passed for all three
cases after refining the disclosure/screenshot workflow and checking the displayed
question against the actual request. Those rechecks and updated frames use fresh
source reads and are archived separately as `documentation-recheck.json` in the
same permanent run directory. No UI/backend records were mocked. The default
suite's opt-in skip was verified, and permanent fault archives survived ordinary
Playwright output cleanup. Journey 28, diagram 15, three real screenshots and
design/planning/issue docs were updated. Unchanged runtime suites and the complete
existing journey suite were not repeated for this test-only addition.

## Structured investigation polish (2026-10-18)

Implemented a shared structured record/span/log attribute viewer, scoped local
search, exact copying, expandable long/nested values and secondary raw records.
Explore adds removable typed filter chips, searchable grouped field choices,
operator-aware value controls and searchable captured-value facets. Full-population
counts and missing/empty distinctions remain intact. Trace selection is explicit,
evidence scope has a concise summary plus disclosure, and Close/Escape restores
focus. Mobile rows stack and attribute actions have 40px touch targets. A design
agent reviewed screenshots and caught cross-signal linked-log actions; these are
suppressed and covered by the live journey.

Passed: 16 Node helper checks, 19 Python tests, JavaScript/Python syntax and
Compose config validation. New live journey 29 passed against collector/ClickHouse
records, covering exact clipboard strings, HTML-like text, long values, local
search query counts, ALL/ANY populations, missing controls, duration comparisons,
linked logs, trace selection, focus and mobile overflow. Its documentation fixture
is labelled as such; it is not SDK instrumentation evidence.

A 30-second Bombadil inspection campaign passed without reported violations:
150 trace records, 23 query actions, 19 attribute opens, local searches/scope changes,
17 trace-open actions and 12 closes. The deterministic journey separately verifies
span selection. Evidence is retained under ignored
`artifacts/bombadil/ui-polish-inspection`; this bounded run does not prove liveness,
source completeness or comprehensive accessibility. The initial invocation omitted
this container's required Chromium `--no-sandbox` flag and failed to launch; rerunning
with that flag and explicit Chromium succeeded.

Existing semantic verification also passed in 44 seconds: Hegel cohort state/swarm
and fault-evidence projection state/swarm, Quint observed cohorts and the fault
contract's retry/rejection/missing cases. Negative controls rejected invented
outcomes/terminals and changed evidence. No query AST, count arithmetic or backend
measurement semantics changed in this polish pass.

The complete default Playwright suite passed: 27 journeys in 5.3 minutes, with the
two explicit opt-in generated campaigns skipped. Existing native metric/fleet,
package authoring, saved-receipt and independent trace/log provenance journeys
passed with the shared viewer. Traffic was restored, the five pre-existing stale
pending entries were retained, and the final UI image was rebuilt and recreated.
README, workspace/design/plan docs, architecture diagram 16 and the executable
storyboard were updated together. Markdown target/fence and diff checks passed;
Mermaid rendering and comprehensive accessibility certification are not claimed.

The explicit generated success campaign passed again in 1.4 minutes on the rebuilt
image, with three real report requests, nine generated checkpoints, five evidence
inspections/revisits and live Quint replay. Its ledger/raw reads/reports are
retained under `artifacts/bombadil/campaigns/787eea19-8769-4166-abe4-2545c4067299`.

The explicit fault campaign also passed (1.3 minutes): three real retry/reject/
stalled requests, nine generated checkpoints, five structured rendered-evidence
inspections, live schedule replay and five qobs/quint-go fault-ledger validations.
The goo actor recorded 508 navigation actions. Permanent evidence is in
`artifacts/bombadil/fault-campaigns/166e2d15-e7e3-4ce2-aae9-c0a10d13cd3e`.
The traffic wrapper restored the owned generator without resetting data or queues.
Together the default suite and explicit campaigns passed all 29 distinct journeys
and regenerated the 85 documentation frames, including four new inspection views.

A final focused journey 29 passed on the rebuilt image (9.1 seconds), regenerating
its four frames after documentation framing adjustments. Desktop attribute scopes,
selected trace context and mobile layout were visually inspected. All 12 services
are running after validation.

## Log and trace discovery (2026-10-18)

Implemented generic quick search and readable log messages/severity, a separate
same-service surrounding-log inspector, trace-ID discovery over full matching span
groups, fixed Go-defined duration buckets and field-guided package reuse. The
same typed AST/compiler handles source queries, filters and saved configuration.
Time context omits identity/search/package predicates and clips to the original
environment/window. Rows clear during radius changes; late replies cannot reopen
closed context or replace the original question. Duplicate identical logs cannot
be uniquely identified by anchor highlighting.

Passed: root Go race suite and vet, 19 Python tests, 19 Node helper checks, focused
exploration/CLI Go tests and syntax/Compose checks. Go boundary tests include all
bucket edges and max UInt64; real collector/ClickHouse assertions check all eight
interval populations. Hegel context state/swarm passes 100 sequences of 30 steps,
with 200 partition property samples. These are captured span distributions, not
native histogram or trace-latency aggregation.

The new browser journey passed in 22.3 seconds with its live-ledger Quint checks:
literal message matching, 108 logs versus a 100-row context page, other-service/
environment exclusion, exact-field anchor highlighting, long multiline/HTML-like
text, close/focus isolation, trace discovery and reusable package preview/save/use.
The Go references validate the browser's actual context/trace/package questions;
negative controls reject changed context bounds, trace environment and grouping.
The first test run had an incorrect expected save-status string; the second exposed
qobs' empty-constant import syntax. Correcting the assertion and adding the model's
INITIAL constant produced a passing run. No UI/backend data was mocked.

A 30-second read-only Bombadil inspection run passed: 231 trace records, including
34 signal changes, 40 queries, eight time-context opens, 15 trace opens and two span
selections. It checked one context inspector at a time and explicit attribute counts,
plus browser errors/rejections. Radius stale-read behavior is deterministically
asserted in journey 30; this bounded random run did not select the radius action.
Evidence lives under ignored `artifacts/bombadil/discovery-inspection` and
`artifacts/discovery/`. This does not establish completeness or liveness guarantees.

The updated live ledger from the default journey was independently rechecked with
Quint, plus the existing exploration model and Hegel exploration state/swarm:
all passed together in 12.9 seconds, including rejecting context/package negative
controls. Permanent ledger queries are sufficient to reproduce these checks;
final focused runs also retain conformance summaries beside their ledger.

The complete default browser suite passed: 28 journeys in 4.9 minutes; the two
explicit opt-in campaigns were skipped. Existing operation, native metric/fleet,
immutable receipt, ad hoc authoring and cross-service context regressions passed.
Traffic was restored and the five pre-existing stale pending stream entries/data
volumes were retained. README, query/workspace contracts, known issues, architecture
flow 17, plan and storyboard were synchronized. Markdown targets/fences and diff
checks passed; Mermaid rendering is not claimed.

Visual review identified a narrow-screen grid issue that placed log messages and
second context actions in the label column. Children now align in the value column,
with new geometry assertions. Re-anchoring additionally retains initial search
scope/basis separately from the new anchor's own source question/basis. Focused
journeys cover these refinements after the full-suite run.

Following the request for Bombadil coverage of the new workflows, added a dedicated
actor with a required discovery prefix plus free exploration. The first fixture
campaign passed in 1.4 minutes: all 19 prefix states were reached, with duration
populations beyond the record-page limit (113 versus 100). The actor exercises
query/search, reported-trace opening/paging, duration groups, package draft/
preview/reuse, log search, context paging/radius/re-anchoring. Nine data invariants
cover population sums, interval order, bounded context/anchor provenance, AND
search, preserved definitions, independent previews, invalidated controls, context
exclusivity/selection and parseable evidence, plus readiness/prefix deadlines and
browser errors/rejections. A fresh package draft now clears old preview/plan/handoff.
The generic inspection actor's trace action was also corrected to explicitly choose
Expand trace rather than the first row action, which can now be Surrounding logs.
Earlier trace-labelled action counts describe requested actions, not proof that
every one opened a trace. The new actor records actual questions and reached states.

The final dedicated discovery campaign passed (1.5 minutes) on the rebuilt image.
It retained 737 Bombadil trace records, reached all 19 prefix states and observed
113 grouped span records against a 100-row page. Free exploration included 29 quick
searches, 16 duration breakdowns, 12 trace breakdowns, 38 draft actions, 18 builds,
nine previews, four uses, six edits and three validations. Prefix actions also
covered trace/context paging, radius and unlinked-log re-anchoring. Both the
Playwright ledger and a second ledger built from Bombadil's actual captured
questions passed qobs/quint-go conformance plus all three negative controls.
Evidence: `artifacts/discovery/discovery-0db3caecbc/`, including coverage JSON,
raw trace/screenshots and `bombadil-model/` summaries. The actor performs no saves.

Focused authoring/UI journeys 24, 25 and 29 passed together in 20.5 seconds after
the final runtime changes. The explicit generated success campaign also passed
(1.4 minutes), retaining its usual generated schedule/live replay evidence.

The explicit retry/rejection/stalled fault campaign passed (1.4 minutes), with nine
checkpoints, five evidence inspections and five fault-ledger model validations.
Together with the default suite and explicit success campaign, all 30 distinct
journeys passed in this round. All 91 documentation frames were regenerated,
including the six discovery frames; architecture/process sources and storyboards
were updated with them. The owned traffic generator was restored, and all 12
services are running with existing data and stale pending entries retained.

Final focused journey 30 passed again (20.8 seconds), including the second-draft
four-to-thirteen-span preview isolation assertion, empty-OR quick-search AND
semantics, unlinked-log re-anchoring, mobile geometry and its Quint negative
controls. Its six frames were regenerated once more from the final running image.

## 2026-10-18 — observed service inventory

Added a Services destination over ordinary captured OTel spans, without application
manifests. A single bounded ClickHouse aggregate supplies full service/span-name
counts, Error status counts and fixed duration-bucket counts. Search filters local
rows without altering totals/basis. Drilldown opens a separate exact service/name
Explore read in the original environment/window, then uses the existing reusable
package authoring/preview path. Health, expected membership and unique request/root
latency remain unknown or separately defined.

Root Go race suite/vet passed, followed by focused final-source race checks for
HTTP/source validation. All 19 Python tests and 20 Node checks passed; Compose
configuration is valid. Hegel observe/reset state/swarm passed 100 sequences of
up to 20 steps, including a final-source rerun (2.14 seconds). Actual browser
capture/search/open/package questions passed quint-go conformance and three
negative controls: changed local totals, widened open scope and changed package
environment. Missing/null dimensions, invalid/overflowing counts, contradictory
partitions, duplicate pairs and source overflow/failure remain fail-closed.

The initial dedicated Bombadil campaign passed 609 records and every required
stage against the real collector/ClickHouse fixture. Its driver independently
checked source counts and statuses, 113 matching spans versus 100 displayed,
package preview, local search, no data and stale-read guards. Evidence:
`artifacts/inventory/inventory-a777c2549e/`. Screenshot review caught compressed
mobile columns; the final layout uses readable stacked cards with a width assertion.

The default suite passed 29 journeys in 5.4 minutes; two opt-in campaigns were
skipped in that command. It regenerated the standard documentation frames,
including the four new inventory frames, from the updated UI. The owned generator
was paused/drained/restored; all five pre-existing stale pending entries and
telemetry volumes were retained. Query/packages were then recreated from final
image `6297a2f117e5`; the previous image produced in this round was removed only
after those containers stopped using it, leaving about 3 GiB workspace free.

Final-image opt-in fault, generated-success and inventory campaigns passed together
(3 tests, 3.3 minutes). Together with the default run, all 31 distinct journeys
passed in this round. The final inventory actor recorded 685 steps and all six
required stages; its actual browser ledger passed Quint plus all three negative
controls (6.06 seconds). Evidence: `artifacts/inventory/inventory-b6aa4b353b/`.
All 95 PNG documentation frames and the generated-campaign JPEG frame were regenerated, including the opt-in campaign
frames. Architecture/workspace diagrams, design limits, planning and storyboards
were synchronized; local documentation links and code fences were checked.
The generator was restored and all 12 services are running with existing volumes.

The shared workspace Bombadil actor now includes Services. A final 20-second
navigation smoke passed 174 records and reached all five primary destinations,
checking that selected navigation and the visible panel agree. Its coverage is
retained in `artifacts/inventory/workspace-services-final/`.

## 2026-10-18 — general captures and collaboration research

Go owns immutable capture creation, verification, bounded storage, catalog and
receipt-bound paging. Python remains the HTTP/static proxy. The UI reuses existing
record/distribution/inventory inspectors in a read-only saved view. The final local
query/packages image is `e4d63a3295a5`; only these services were recreated.

Final native validation and browser evidence are described in
[VALIDATION.md](VALIDATION.md). Test wrappers pause and restore only demo traffic;
the five previously stale pending Redis entries and telemetry volumes are retained.
VFS disk pressure was resolved by removing 16 unused untagged goo query images and
pruning build cache, without deleting running/tagged images, containers or volumes.
Build/cleanup logs remain in ignored `artifacts/captures-*.log`.

The final capture journey passed in 1.1 minutes, including the real ledger's
Quint validation (7.05 s), four rejecting controls, and a 35-second read-only
Bombadil exploration: 367 trace records through all eight stages, including the
required complete prefix. Evidence: `artifacts/captures/capture-968a089abf/`.
Focused journeys 19 and 22 passed on the final runtime. The final Hegel rerun
passed in 12.939 s; focused Go race checks and vet passed again.

The generated fault campaign (journey 28) also passed on this runtime in 1.3
minutes, retaining its actual-browser/Quint retry, rejection and missing-terminal
checks. Its three documentation frames were regenerated.

The final serialized capture/fault/generated-success run passed all three tests
in 3.5 minutes; journey 27 passed in 1.1 minutes. Success evidence:
`artifacts/bombadil/campaigns/432b83ac-cc94-4266-ad9a-c97531f1c910/`;
fault evidence: `artifacts/bombadil/fault-campaigns/0e501301-6e99-4bc3-b6eb-d360f0afe089/`.
Together with the default suite and focused corrected assertion, all 32 distinct
journeys passed across these runs. All 101 PNG frames and the generated-campaign
JPEG were refreshed after this round's initial runtime build. The final mobile
saved-investigation frame was visually reviewed; local links and Markdown fences
were checked (no rendered Mermaid validation is claimed). All 21 Node checks
passed again. Demo traffic was restored after each wrapper invocation.

## 2026-10-18 — core workbench basics

Implemented and documented in [core workbench](docs/CORE_WORKBENCH.md): optional
exact span measures/time charts, package-backed reusable questions, observed
span/log service signals/parent relationships, local general-query admission and
explicit cancellation. The final query/packages image is `b8a1e38a13af`; only those
two services were recreated for this increment. Source reads use the verified
ClickHouse 25.3 setting `cancel_http_readonly_queries_on_client_close=1`.

Root Go race tests/vet passed after the final semantic changes; 19 Python tests
and 22 Node checks passed. The first real journey passed in 18.9 s; focused
capture/core/inventory checks passed in 2.3 minutes. Focused core evidence:
`artifacts/core-basics/basics-5c2b58cc/`, including 187 Bombadil records, six
required stages and actual-ledger Quint validation plus four negative controls.
The final ledger checker adds a fifth negative control for reused environment
scope. Hegel duration state/swarm passed 100 sequences, up to 30 steps (0.248 s).
Local links and fences passed across 41 documents. The new desktop summary frame
was visually reviewed; final narrow-screen and screenshot freshness checks follow
the comprehensive run. No production identity or ingestion-health claim is made.

The final core journey passed in 1.0 minute, including five rejecting Quint
controls (10.72 s), reused-environment preservation, the actual Python cancellation
endpoint and a 30-second Bombadil run. The actor recorded 195 trace records through
all six prefix stages. Evidence: `artifacts/core-basics/basics-a594217a/`. The final
mobile saved-measures frame was visually reviewed and passed document-width checks.

The comprehensive final runtime run passed **all 33 journeys in 11.9 minutes**,
including both generated success/fault campaigns, actual application/ClickHouse
observations, native counter/histogram/fleet journeys and UI polish. No tests were
skipped in this run. All 105 PNG storyboard frames and the generated-campaign
JPEG (106 images total) were regenerated after the final image build.
Forty-three documentation files passed local-link and fence checks.

Native focused race checks passed again after adding cancellation-before-admission
and summary-versus-time-read disagreement tests. Source disagreement remains
explicit rather than converting an empty chart to zero duration. The final
focused browser rerun adds a companion live workflow Bombadil actor and strengthens
Quint controls to use plausible-but-wrong p95/sum values.

The final expanded journey 33 passed again in 2.0 minutes (2.1-minute wrapper),
including both read-only Bombadil actors and five strengthened Quint controls.
Saved-measure actor: 216 records / 6 prefix stages.
Live workflow actor: 181 records / 10 prefix stages, including
fixed handoff, saved question reuse, service signal/parent-link diagnostics and
cancellation. Final evidence: `artifacts/core-basics/basics-197824e0/`; live workflow
trace/coverage lives in its `workflows/` directory. Neither actor saves resources
or emits application workload. All 12 services were restored/running; the five
pre-existing stale pending entries and telemetry volumes were retained.

## 2026-10-18 — grouped duration measures and polish

Implemented per-group exact duration aggregates, semantic adapter normalization,
integer/rational chart ranking, package previews and readable saved/mobile layouts.
See [contract](docs/GROUPED_MEASURES.md), architecture flow 21 and journey 34.
The reviewed query/packages image is `96cd32499e15`. Only those services were
recreated for the feature; telemetry volumes and the five pre-existing stale Redis
entries were retained. Test wrappers continue to restore owned demo traffic.

Root Go race/vet, 19 Python tests and 23 Node checks passed. Hegel group conservation
passed 100 generated populations (0.254 s). The initial six-journey campaign passed
five, including the new group journey with one/two dimensions, count comparison,
package preview, stored-page replay and mobile width checks. Its Quint trace plus
three negative controls passed in 7.32 s; the saved-only Bombadil actor recorded
454 records through nine required stages. Evidence:
`artifacts/grouped-measures/groups-15b514e3/`. The discovery failure was an omitted
accessible trace label, restored in the reviewed image. The earlier checkbox
failure was a blur-triggered layout shift, corrected with input-time invalidation;
a model formatting mismatch was corrected with typed group comparisons.

The new mobile group frame was visually reviewed. Static local links and Markdown
fences passed across 44 documents; no rendered Mermaid validation is claimed.
Final targeted regression and affected-frame refresh results follow below.

The final reviewed-image regression passed **all 10 selected journeys in 3.5
minutes**: capture/core/discovery, grouped measures, general exploration, exact
trace context, distribution comparison, package authoring/import and UI polish.
The restored accessible trace label passed the original discovery assertion.
Journey 34 passed in 52.9 s, including one/two-field aggregates, package preview,
mobile Records/Distribution containment, three rejecting Quint controls (6.46 s)
and a 30-second saved-only Bombadil run: 383 records through all nine stages.
Final evidence: `artifacts/grouped-measures/groups-e219e14d/`.

Four new storyboard frames and the affected existing Explore/capture/definition/
mobile frames were regenerated from the running app. The final mobile ranking
frame was visually reviewed with readable saved values and explicit count-baseline
labelling. Unaffected native-operation screenshots were retained. All 12 services
were restored; no queues, volumes or stale pending entries were deleted.


## Error and latency comparisons — 2026-10-18

The reviewed query/packages offline image is `96871367e708` (build log:
`artifacts/error-comparison-reviewed-build.log`). The preceding polished build
`e5b8714307c2` passed all 11 affected regression journeys. The Go semantic core and UI assets
were staged into the existing offline Python image context. Only query/packages
were recreated; no service/port/dependency or volume reset was added. The first
build invocation selected the read-only default Docker config and failed before
building; selecting the already configured `/workspace/.docker-goo` resolved it.
Existing proxy/CA configuration was preserved.

Passed native evidence: `artifacts/error-comparison-go.log` (Go race/vet),
`error-comparison-capture-refresh.log` (focused refresh/lineage race checks),
`error-comparison-python.log` (19 tests), `error-comparison-node.log` (24 checks),
`error-comparison-hegel.log` and `error-comparison-model-regression.log`.
Hegel exercises 100 sequences of 25 comparison state/swarm steps; the existing
cohort swarm also passes with environment drift separated from supported windows.
Quint's existing cohort adapter passes with the updated scope contract.

Initial journey 35 passed in 54.8 s; evidence is
`artifacts/error-comparison/changes-26580233/`, including four rejecting Quint
controls (9.12 s) and 384 Bombadil records across six required stages. The polished
image campaign passed journey 35 in 47.0 s; final new-feature evidence is
`artifacts/error-comparison/changes-a7f4e1b8/`, including four rejecting Quint
controls (7.20 s) and 342 Bombadil records across all six stages. Five new screenshot
frames document errors, latency/missing groups, saved pair, mobile comparison and
trace/log continuation. The final mobile frame was visually inspected.

The full affected regression log is `artifacts/error-comparison-regression.log`.
The journey wrapper pauses/drains/restores only owned traffic and retains all
telemetry volumes and the five pre-existing stale Redis pending entries.


All 11 affected regression journeys passed in 3.9 minutes, including both opt-in
comparison/group campaigns. The group campaign recorded 517 Bombadil records
through nine stages (`artifacts/grouped-measures/groups-c3914da2/`). Final review
disabled an otherwise inactive trace action in aggregate-only pair views; their
linked child capture remains the path to record/trace inspection. The final
reviewed image passed both discovery (18.9 s) and the strengthened comparison
journey (54.3 s), including actual adapter status-class groups and saved trace-ID
pair behavior. The final comparison ledger/campaign is
`artifacts/error-comparison/changes-bd815a45/`, with four rejecting Quint controls (9.20 s)
and 501 Bombadil records through all six stages. The focused log is
`artifacts/error-comparison-reviewed-journeys.log` (two passes, 1.3 minutes).
A fresh CGO-disabled Go build matched the staged runtime binary byte for byte;
all UI files matched the staged runtime assets. All 12 services were running after
the wrapper restored traffic. Changed-doc relative links and Markdown fences were
checked; rendered Mermaid diagrams were not separately browser-validated.


## Deployment context and ordered investigation notes — 2026-10-18

Added local Go workbench storage/routes and UI modules without new services,
ports or dependencies. All three CLI-building Dockerfiles now include the
`workbench` package. The existing host-staged offline image path preserves CA/proxy
configuration. Initial image `a5c89ff38465` is in `artifacts/context-build.log`;
UI iteration used explicit copies into the owned query container. Reviewed image
`8b0b37f85790` (`artifacts/context-reviewed-build.log`) restaged those assets and is
the basis of the broad regression. Only query/packages were recreated; telemetry
volumes and pre-existing Redis entries were retained.

Native evidence: `artifacts/context-go.log` (root race/vet), `context-python.log`
(19 tests), `context-node.log` (26 checks), `context-hegel-reviewed.log` (60 Hegel
sequences of 15 state/swarm steps, 0.38 s) and `context-correlation-replay.log`
(strengthened six-control Quint ledger replay, 19.97 s). The original verification
attempt collided with an existing `contextMachine` type; renaming the new model
fixed the build. `artifacts/context-journey.log` records that initial verifier
failure after successful UI steps; the captured ledger replay passed afterward.

The first complete new-feature campaign (`artifacts/context-campaign.log`) passed
journey 36 in about one minute. Its actual sample-app/notes ledger is
`artifacts/context-notes/8901ddb9/`, with five rejecting Quint controls (11.71 s)
and 383 Bombadil records through all seven draft-prefix stages. The broad run's
new-feature ledger is `artifacts/context-notes/9df68764/`: journey 36 passed in
51.8 s with five rejecting controls (9.76 s), 361 Bombadil records and all seven
stages. A subsequent replay added explicit scoped report-ready correlation and
the sixth rejecting incorrect-subject control. Markers in this journey are authored
claims; the test generates reports but does not deploy a new service version.

The broad log is `artifacts/context-full-journeys.log`, with generated workload,
fault and cross-app Quint campaigns enabled. Keyboard focus and delayed-navigation
guards found during review are checked separately on the final staged build.


The broad run passed **all 36 journeys in 12.6 minutes, with no skips**. All 120
storyboard images (119 PNG / one JPEG) were regenerated after the reviewed build.
The final keyboard/navigation image `8d8c93e8f37f`
(`artifacts/context-final-build.log`) passed capture continuity, notes and
structured inspection together (`artifacts/context-final-journeys.log`, three
passes in 2.0 minutes). Its notes ledger is
`artifacts/context-notes/eb9730e6/`, with six rejecting Quint controls (13.16 s)
and 406 Bombadil records through all seven stages. Assertions cover keyboard
reorder/removal focus, exact cited-capture heading focus and a genuinely delayed
stored-revision response after workspace navigation.

Final visual inspection found a stale loading message left by the canceled read.
The corrected reviewed image is `0b90e3545d60`
(`artifacts/context-final-status-build.log`). The final journey passed in
1.2 minutes (`artifacts/context-final-status-journey.log`), including the new
status assertion. Its complete ledger/coverage is
`artifacts/context-notes/07a93eea/`: six rejecting Quint controls (17.15 s),
310 Bombadil records and all seven required stages. The final mobile frame was
visually reviewed. Runtime UI assets and Python proxy matched the staged sources;
a fresh Go build was compared against the staged executable. All 12 services were
running after traffic restoration; the five stale pending entries and all volumes
were retained.

Before final staging, unused Docker builder cache was reclaimed without deleting
volumes or running services (`artifacts/context-builder-prune.log`); the separate
dangling-image prune reclaimed zero bytes. Changed-document relative links and
Markdown fences were checked; Mermaid rendering was not separately browser-tested.
Full networked multistage Dockerfile builds were not rerun; the three COPY updates
were reviewed and the actual CLI was built/staged and exercised in the offline lab.


## Shared sample investigations and scoped recipes — 2026-10-18

Baseline: `5c66e1148d157fdc6f04e8c2ce2a8f2d4a05eff7`. Three pure Go recipe
transforms share the existing Explore compiler; six ordinary YAML definitions
move sample investigation entry points into Explore/packages/captures/notes.
The original lifecycle/resource lab is explicitly opened rather than queried on
Examples navigation. A derived question clears the previous package banner.
[Contract](docs/INVESTIGATION_RECIPES.md), architecture flow 24 and storyboard 37.

Verified in this instance:

- Root Go race suite and vet: `artifacts/recipes-go.log`, `recipes-vet.log`.
  Input isolation, full OR/predicate preservation, typed SQL, signal guards,
  strict HTTP and all six package definitions are covered; readable YAML was
  separately validated in `recipes-examples.log`.
- 19 Python checks (`recipes-python.log`), 27 Node checks
  (`recipes-node-final.log`). Hegel state/swarm: 100 sequences of 25 steps,
  reviewed pass 0.19 s (`recipes-verification.log`; the standalone live-ledger
  opt-in is skipped there, then executed by the real browser journey).
- Broad run: `artifacts/recipes-full-journeys.log`, 37 executed / 36 passed /
  one harness failure in 14.8 minutes, no skips. Generated/fault campaigns and
  cross-app Quint enabled. The sidebar-driven legacy test needed the same
  explicit disclosure action as direct-navigation tests. Its old hidden-control
  timeout is retained as evidence, not reported as a passing run.
- Final focused run: `artifacts/recipes-final-focused.log`, two passes in
  1.2 minutes, including the corrected legacy test and current recipe flow.
  Real report work is independently located in ClickHouse; UI filters narrow it;
  recipe preview/run preserves that population; capture and notes precede deriving
  and freshly running its package. Source-free preview/notes, wrong-signal and
  delayed-result guards, mobile containment and all six sample definitions pass.
- Final ledger: `artifacts/recipes/933a4f7b/ledger.json`. Quint positive replay and
  six rejecting controls pass in 13.04 s. The model is bounded observed conformance,
  not exhaustive browser verification. Bombadil records 467 observations through
  initial/latency/errors/complete, with all three recipe choices; see `coverage.json`
  and `bombadil/trace.jsonl` in that directory.
- All 125 storyboard files (124 PNG, one JPEG) refreshed since the broad run;
  the focused run refreshes the corrected definition frames and latest recipe
  frames. Desktop gallery/preview/notes and mobile recipe frames were inspected.
  Markdown local links/fences checked; Mermaid rendering was not separately run.

Initial harness/model failures for save wording, disclosure state and a reserved
Quint action name were corrected. The model then rejected treating a fresh package
window as the original fixed hypothesis; the current UI journey captures first
and derives the reusable definition from that stored question. Captures may
include later arrivals than live results; the model checks exact question scope,
not equality of two separate-read counts.

Retained query/packages image:
`sha256:007a33fef1e368313bacd99755b6947af689559a8dc53542f6dd074bdd4fb656`.
Offline build log: `artifacts/recipes-reviewed-build.log`; runtime UI/Python/example
inputs match the staged source, and a fresh CGO-disabled Go build matches its
binary byte-for-byte. This is a precommit build with parent-revision/dirty metadata.
The networked multistage Dockerfiles were not rerun. Compose configuration passed;
all 12 services run after the wrapper restores traffic. Existing volumes and five
stale Redis pending entries were retained; no resets were used.

No environment initialization or dependency changes were required. The retained
startup instructions and local Docker/toolchain setup still apply. Declared recipe
prerequisites are not verified coverage; the specialized lifecycle/Redis contracts,
production authorization/storage and general user-authored workflows remain
separate work.

## Shared measurement standardization — 2026-10-18

Current behavior is documented in [MEASUREMENT_PACKAGES.md](docs/MEASUREMENT_PACKAGES.md)
and [VALIDATION.md](VALIDATION.md). New runtime work remains in Go; the Python lab
continues to emit standard OTel data and proxy the bounded APIs. Both woven Go
examples were rebuilt and actual OTLP logs with trace/span IDs were independently
observed in ClickHouse. The Python meter now declares instrumentation version/schema.

Root race/vet, the separate verification module (uncached, 240.18 s), instrumentation,
19 Python tests, 32 frontend checks and compiler/lifecycle/package parity checks
passed. The focused measurement journey passed all four families, scoped lifecycle
outcomes, capture/refresh/reopen, notes and reusable definitions. Its live Quint
ledger passed with seven rejecting controls; Bombadil recorded all four required
preview/edit stages. Full browser and final targeted results are recorded in the
new top section of VALIDATION.md; historical entries below/above retain their
original scope.

The external offline build context remains `/workspace/.goo-image`; the repository
Dockerfiles also include the new `measurements` package. Default telemetry volumes
and five pre-existing stale pending Redis entries were retained. Journey wrappers
pause/drain/restore only the owned traffic generator. New storyboard frames 38a–g
and shared measurement architecture diagrams describe actual implemented paths.

## Metric discovery round (2026-10-18)

The existing cloud toolchains and running lab were reused. The SDK-only Python
producer exports through the real collector; discovery and typed evaluation read
actual ClickHouse tables. Both affected browser journeys passed, including both
Bombadil specifications. Root Go tests, affected race tests, vet, 32 frontend
checks and Hegel optional-identity/state checks passed. Updated verification and
storyboard details are in VALIDATION.md.

The offline image build exhausted disk space. Regenerable Go cache and unused
Docker image/build cache were cleaned, preserving all running application data
and telemetry volumes. Sandbox startup and build/test operations recovered;
free space returned to approximately 14 GiB. The traffic wrapper retained all
five pre-existing stale pending entries. No environment draft change was needed.

## Coherence pass (2026-10-18)

Existing tooling/lab reused; no environment configuration change required.
Five affected browser journeys were verified across the initial run and corrected
focused rerun, including three Bombadil campaigns and the existing live Quint
ledgers/negative controls. All 32 frontend checks passed. The new image
`11a6f58b5951` was built from the existing offline context; telemetry volumes and
five pre-existing stale pending entries were preserved. The owned traffic generator
was restored after both wrappers. Details and diagnosed harness failures are in
VALIDATION.md. The product brief is docs/PRODUCT_ESSENCE.md.

## Question-first investigations (2026-10-18)

The existing cloud toolchains/lab were reused; no setup configuration changed.
The final offline image `d13fcbdbce4a` passed four browser journeys with all four
optional Bombadil campaigns and live Quint checks enabled. Root Go tests,
affected race tests, vet, 32 frontend checks and expanded Hegel state/swarm checks
passed. Storyboard 41's real application/OTel workflow also verifies fresh reuse,
exact handoffs and dirty-draft preservation. All evidence is linked from
VALIDATION.md. Owned traffic was restored; existing telemetry volumes and the
five stale pending entries remain intact.

## Contextual investigation follow-ups (2026-10-18)

Existing cloud tooling and the running lab were reused; no environment setup
configuration changed. Offline image `77be6b447ade` serves the Go follow-up
planner and connected UI. Root Go tests, affected race tests, vet, 32 frontend
checks, Python compilation and Hegel state/swarm passed. The new real-stack
journey passed with seven rejecting Quint controls and its Bombadil campaign;
details and final regression evidence are recorded in VALIDATION.md.

Final browser regression passed all six affected journeys in 5.5 minutes, including
five enabled Bombadil campaigns and live Quint/negative controls. Desktop/mobile
storyboards were regenerated. The wrapper restored traffic and preserved existing
telemetry volumes and all five pre-existing stale pending entries.
