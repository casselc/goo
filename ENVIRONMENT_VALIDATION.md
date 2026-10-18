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
