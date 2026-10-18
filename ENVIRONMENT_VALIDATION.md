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
