# goo — Grammar of Observability

[Population exploration and saved views](docs/POPULATION_WORKSPACE.md) bring aligned charts and matching span/log pages to arbitrary cross-service selectors. Portable filters are retained; independent companion selectors are explicit. Saved views retain fixed queries, layout and columns, and reopen without source reads.

[Linked service timelines](docs/LINKED_TIMELINE.md) align span activity, errors, latency, logs and native streams. Preview a UTC range without reads, then explicitly apply it to every overview widget or open the selected span/log question.

**Turn telemetry into shared definitions of system behavior, and investigate what
happened with evidence you can inspect and preserve.**

Ask a question → inspect its evidence → capture and explain it → reuse the
definition. Start from instrumented code or unfamiliar OpenTelemetry data.
[The product essence](docs/PRODUCT_ESSENCE.md) describes this common loop and the
criteria for keeping the workbench coherent. The [core primitives review](docs/CORE_PRIMITIVES.md)
sets the next consolidation priorities; [shared scope and checked transforms](docs/QUESTION_SCOPE.md)
now connect recipe and follow-up planning.
[One Services catalog](docs/UNIFIED_SERVICE_CATALOG.md) now connects the existing
span, log and native metric discovery paths through one searchable entry.
[Automatic service overviews](docs/SERVICE_DASHBOARDS.md) compose span/log activity,
duration/status breakdowns and supported native instruments. A distributed view
shows resolved parent relationships and keeps uncertain boundaries explicit.
Each widget opens a fixed question or an explicit evidence capture.
[Service browsing](docs/SERVICE_BROWSING.md) connects observed signals, operations,
OTel attribute suggestions and native instruments through fixed-window questions.
[All-signal discovery](docs/METRIC_SERVICE_DISCOVERY.md) includes metric-only
services, counting data-point records without combining metric values.
The first [question-workbench cleanup](docs/QUESTION_WORKBENCH.md) provides one Run
path, explicit draft/applied scope, one-gesture drilldowns and fixed-window metric
discovery. The [design review](docs/ENTRY_POINTS_DESIGN_SPIKE.md) records the broader
proposed interaction and navigation work.

This repository currently runs as a local lab. Reservation, checkout and report
services exercise general-purpose contracts.

[Question-first investigations](docs/WORKBENCH_CONTEXT.md#question-first-investigation-workspace)
bring the question, finding, open questions and cross-signal evidence together.
[Typed question steps](docs/INVESTIGATION_STEPS.md) preserve record questions and
measurement definitions in the same named investigation. Reopen either without
executing it, then explicitly Run again. [Linked evidence histories](docs/LINKED_QUESTION_EVIDENCE.md)
add Run and capture, show scope differences, and preserve earlier results.
Capture directly into an investigation draft or a new child of a saved revision;
reopen the handoff without telemetry reads, or reuse its definition against fresh
data while retaining the original evidence.

[Contextual follow-ups](docs/INVESTIGATION_FOLLOWUPS.md) connect saved evidence to
another explicit question: duration/status breakdowns, exact trace/log links,
fixed-window resource discovery, or a recorded evidence gap. Each preview names
which population and time scope it keeps or replaces before any source read.

[Typed record measures and split time charts](docs/RECORD_MEASURES.md) compare
captured counts, duration and recorded status across bounded breakdown series.
Local chart controls retain the captured scope and scale; captures and reusable
packages preserve the chart settings and evidence separately.

[Numeric attribute measures](docs/NUMERIC_ATTRIBUTE_PLANNING.md) add explicit integer
bindings, capability diagnostics and exact sums/extrema/means with distinct coverage
counts. Explore, charts, saved evidence and reusable YAML share the binding.

The [investigation workspace](docs/PRODUCT_WORKSPACE.md) brings Explore, Services, Operations,
Saved investigations and Definitions into one interface. Explore includes captured
attribute distributions, pinned baseline comparisons and a trace/log inspector.
[Examples now use ordinary telemetry packages](docs/INVESTIGATION_RECIPES.md) for
reservations, Go checkout and Go reports. Scoped recipes retain the applied
population while exploring duration, recorded errors or log severity. Save the
question as a package, or capture evidence and cite it in ordered notes. The
[measurement library](docs/MEASUREMENT_PACKAGES.md) adds typed gauges, counters,
native histograms and scoped lifecycle outcomes to the same capture/notes flow.
Data-only delay/retry/RED/USE/report workflows preview shared package questions;
the original specialized lab remains available for compatibility.

Explore and trace inspection share searchable, typed attribute rows, exact-value
copying and secondary raw evidence. Scoped filter chips and searchable facets
carry investigations from a captured value into a fixed-window breakdown. See
[the inspection storyboard](docs/journeys/README.md#29-structured-inspection-and-filter-composition).

Log discovery adds literal message/service/severity search and a separately scoped
surrounding-log inspector. Trace discovery opens reported trace IDs from full
matching span groups and shows fixed, full-population span-duration buckets.
Suggested breakdowns feed directly into reusable packages. See
[the discovery storyboard](docs/journeys/README.md#30-from-log-search-to-reusable-span-distributions).

[Services](docs/SERVICE_INVENTORY.md) discovers emitting services and span names
with full captured counts, error statuses and duration buckets. Open a fixed-window
Explore question, then build a reusable package without a source manifest.

[Investigation captures](docs/GENERAL_CAPTURES.md) preserve Explore and inventory
questions with full aggregates and bounded stored records. Reopen without source
queries, page immutable evidence, or refresh the same fixed question into a new
revision. [Collaboration research](docs/COLLABORATIVE_INVESTIGATIONS_RESEARCH.md)
compares boards, notebooks and investigation handoffs across observability tools.

[Ad hoc telemetry packages](docs/TELEMETRY_PACKAGES.md) turn an Explore question
into validated YAML, a live preview and an immutable reusable definition. They work
over existing OTel telemetry from any language without an application manifest;
optional Go annotation templates are reviewable exports.

For the end-to-end design, start with the [architecture and process diagrams](docs/ARCHITECTURE_DIAGRAMS.md):
source validation and weaving → application OTel SDKs → collector/ClickStack capture
→ goo package planning and evidence-based investigations. Implemented local paths
and illustrative production patterns are labelled separately.
The [architecture guide](docs/ARCHITECTURE.md) describes component boundaries;
the [implementation plan](docs/IMPLEMENTATION_PLAN.md) tracks remaining work.


The [Go aggregate compiler pilot](docs/PORTABLE_COMPILER.md) now validates a small
declarative measurement model, evaluates prepared facts, and emits ClickHouse
SQL or Databricks metric-view YAML. See [project direction](PROJECT_DIRECTION.md)
for Go/Rust and YAML+CEL/KCL preferences. The Python lab below remains the
reference prototype; bounded Go HTTP/operation packages now serve its investigation
UI, while the original reservation SLO preview remains separate.

The [lifecycle preparer](docs/LIFECYCLE_PREPARATION.md) now turns normalized
business events into frozen subject facts and evidence receipts. Explore the
[executable user storyboards](docs/journeys/README.md),
[Hegel/Quint verification](verification/README.md), and
[bitemporal design note](docs/BITEMPORALITY.md).

The [Bombadil coordinator](docs/BOMBADIL_TESTING.md) correlates a real sample-app
request with independently captured telemetry and goo's rendered trace while a
second browser explores navigation. Optional Quint validation checks the observed
cross-app checkpoints. An opt-in generated campaign uses Quint MBT schedules to
drive Bombadil app submissions and goo evidence inspections/revisits.
The fault campaign adds Python retry, rejection and stalled-work evidence, with
Hegel state/swarm properties and Quint checks that preserve unknown outcomes.

The reference reservation example emits ClickStack-shaped OTel data: HTTP gateway → gRPC reservation service → Redis Streams → inventory consumer → OTel collector → ClickHouse. Its optional specialized reservation lab under Examples edits a validated JSON query expression and shows its compiled SQL, completion cohorts, gRPC latency, logs, spans and links, and exported metric points.

The [visualization slice](docs/VISUALIZATIONS.md) adds trace waterfalls with span
inspection and async-link navigation, plus a raw-duration heatmap whose cells
open matching spans. The [fourth browser storyboard](docs/journeys/README.md)
executes this path through the actual sample services.
[Guided investigations](docs/INVESTIGATION_WORKFLOWS.md) provide outcome, retry,
RED and USE entry points, with explicit instrumentation gaps.

The [semantic package foundation](docs/SEMANTIC_PACKAGES.md) recognizes versioned
OTel contracts in Go and proposes reusable measures/views/recipes, with explicit
missing requirements. `go run ./cmd/goo catalog` lists the ten current profiles;
`goo discover -inventory ...` checks supplied metadata. These proposals are not
auto-installed dashboards or objectives. The [worker/queue adapter](docs/QUEUE_TELEMETRY.md)
now supplies real lag, pending, worker utilization and observed wait evidence to USE.

The [HTTP package expansion](docs/PACKAGE_EXPANSION.md) now turns an explicit
binding into shared measurement/view/recipe artifacts and both backend plans.
`goo prepare-http` classifies normalized selected server requests under that
policy. The live adapter below also renders these descriptors;
`scripts/check_packages.py` executes the offline CLI parity storyboard.

The [live HTTP package adapter](docs/LIVE_HTTP_PACKAGE.md) now reads real OTel
spans in Go and renders the expansion's shared aggregate descriptor in the UI.
The [instrumentation direction](docs/INSTRUMENTATION.md) covers library discovery
and opt-in Orchestrion/quintgo model observations at goo's semantic seams.

The [goo annotation slice](docs/GOO_ANNOTATIONS.md) adds `//goo:operation`,
`//goo:event`, source-contract manifests and opt-in `goo weave` test builds.
Woven OTel emission and active-operation state/swarm checks are executable;
aggregate policy stays in packages.

## Continue in Codex

Open this folder as the project root. Start with [START_HERE.md](START_HERE.md) and give Codex [CODEX_HANDOFF.md](CODEX_HANDOFF.md). [AGENTS.md](AGENTS.md) contains repository working guidance. No prior chat history is required.

The runnable source is the current Python prototype. The RFCs describe the proposed next implementation; their EDN and richer expression syntax are not accepted by the current API. The handoff bundles the earlier design draft, current limitations, staged tasks, acceptance criteria, and source references.

| Document | Purpose |
| --- | --- |
| [START_HERE.md](START_HERE.md) | Project orientation and reading order |
| [CODEX_HANDOFF.md](CODEX_HANDOFF.md) | Ready-to-use continuation task |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | Current components and target boundaries |
| [Architecture diagrams](docs/ARCHITECTURE_DIAGRAMS.md) | Build, production telemetry flow, investigations and backend portability |
| [Live fleet comparisons](docs/FLEET_COMPARISONS.md) | Two woven instances, actual member windows, bounded skew, loss and restart journey |
| [Fleet coverage](docs/FLEET_COVERAGE.md) | Declared membership, supported intervals, guarded exact totals and missing-member storyboard |
| [Operation error ratios](docs/OPERATION_ERROR_RATIOS.md) | Aligned exact calls/errors, explicit unknown/idle intervals and guided trace investigation |
| [Native histogram semantics](docs/NATIVE_HISTOGRAMS.md) | Exact bucket deltas, aligned explicit merging, heatmap and bucket-bounded P95 |
| [Native counter semantics](docs/NATIVE_COUNTERS.md) | Reset/gap-aware rates, evidence limits and executable storyboard |
| [docs/IMPLEMENTATION_PLAN.md](docs/IMPLEMENTATION_PLAN.md) | Ordered milestones and completion criteria |
| [docs/KNOWN_ISSUES.md](docs/KNOWN_ISSUES.md) | Specific defects, limitations, and validation gaps |
| [docs/rfcs/0001-observability-language.md](docs/rfcs/0001-observability-language.md) | Language and resource model proposal |
| [docs/rfcs/0002-measurement-semantics.md](docs/rfcs/0002-measurement-semantics.md) | Population, time, counting, and evidence rules |
| [docs/design-draft.md](docs/design-draft.md) | Portable copy of the original worked design |
| [docs/SOURCES.md](docs/SOURCES.md) | Reading list and what each source contributes |
| [VALIDATION.md](VALIDATION.md) | Actual checks and unverified environment behavior |

The core stack uses ClickHouse and the upstream OTel ClickHouse exporter with the same baseline table shape used by ClickStack. HyperDX is optional, in the `hyperdx` profile. This is a small exploratory implementation of the design, not a complete grammar interpreter or production SLO service.

## Start

Install Docker Engine/Desktop with Compose v2. Allow roughly 4 GB of memory and a few GB of disk for images. On Linux, ClickHouse may require `sudo sysctl -w vm.max_map_count=262144` if its startup log asks for it. Intel/AMD Linux is the primary target; image support on other architectures has not been tested.

```sh
docker compose up --build -d
docker compose logs -f collector reservations consumer
```

Open **http://localhost:8088**. Traffic starts automatically. Wait 20–40 seconds, then select **Run query**. Services bind host ports only to loopback; credentials in this project are disposable local credentials.

```sh
python3 scripts/smoke.py
```

The smoke script waits for startup, creates a successful reservation, waits for its mature completion, and checks that logs, spans, latency results, and native metrics have arrived through the collector.

## Explore

1. Query **Completion SLO**, grouped by release or scenario. Good and retried requests usually finish on v1; the v2 cohort contains late, missing, and rejected work.
2. Click a reservation ID to inspect its processing spans. The consumer deliberately starts a separate trace linked to the producer span, so business identity works across trace IDs.
3. Change to **Logs** to inspect events for that selected reservation. Clear the selection to return to the full population.
4. Inspect **Latency heatmap**, click a populated cell, select a span, then open its reservation waterfall. **gRPC latency** also shows percentile summaries. Fast synchronous acceptance can coexist with failed asynchronous completion.
5. Inspect **Native metric points**. These are cumulative exported counter points; the UI does not sum them or claim they are rates.
6. Edit the JSON query and run it. Unknown operations, invalid dimensions, and unsupported metric filters are rejected. Expand the compiled SQL panel to inspect parameters.

The application-owned definition is `definitions/reservations.json`. Edit it and rebuild the application images to change the deadline, target, allowed views, or grouping dimensions. The current JSON subset is designed for portable data authoring; EDN parsing and the full package resource model are not yet implemented.

Example query:

```json
{"view":"completion","minutes":30,"groupBy":"region","scenario":"missing"}
```

Create a chosen scenario manually:

```sh
curl -s http://localhost:8081/reserve \
  -H 'Content-Type: application/json' \
  -d '{"scenario":"retry"}'
```

The gRPC service uses a real unary gRPC method `/lab.Reservations/Create` with an explicit JSON codec, avoiding generated protobuf application code in the minimal sample. OTel signals use standard protobuf OTLP/HTTP.

| Scenario | Behavior |
| --- | --- |
| good | Confirmation after about 0.3 seconds |
| retry | One failed processing attempt, then confirmation |
| late | Confirmation after about 14 seconds |
| missing | Accepted but deliberately never completed |
| reject | Terminal rejection |

The demo deadline is 10 seconds with 3 seconds of ingestion allowance. The fixed traffic mix is intentionally unhealthy and has a 95% target to make comparisons visible quickly. It is not a realistic workload or recommended target.

## Optional HyperDX

```sh
docker compose --profile hyperdx up --build -d
```

Open **http://localhost:8080**, create a local account, and use the preconfigured **Lab** ClickHouse connection. If not prepopulated, use host `http://clickhouse:8123`, user `lab`, password `lab-local`. Add log and trace sources for database `default`, tables `otel_logs` and `otel_traces`; accept the detected mappings. Add a metrics source for the `otel_metrics_*` tables if your HyperDX setup asks for one.

HyperDX 2.41.0 is specified explicitly. The optional profile needs MongoDB and extra memory. The upstream collector is deliberately kept independent of HyperDX authentication/OpAMP. These are baseline exporter schemas, without newer HyperDX-specific materialized views, session replay, or PromQL tables.

## Ports and services

| Port | Purpose |
| --- | --- |
| 8088 | Grammar UI and query API |
| 8081 | Sample HTTP endpoint `POST /reserve` |
| 8123 | ClickHouse HTTP, local credentials required |
| 4318 | Collector OTLP/HTTP |
| 8080 / 8000 | Optional HyperDX UI / API |

The gRPC service and Redis stay internal to Compose. All application roles share one small Python image. ClickHouse and Redis use named volumes. No kind cluster is required.

## Query and evidence semantics

`grammar.py` validates a small expression and compiles it to fixed-source ClickHouse SQL with typed value parameters. Record counts collapse by tenant and reservation ID. Duplicate event deliveries do not multiply subject counts. Mature subjects are classified as good, bad, or unknown; recent acceptance remains pending. A missing completion becomes bad only as a demo assumption. Conflicting terminal decisions and negative durations become unknown.

The initial UI shows a bounded lookback anchored on exported event Timestamp values, with recent subjects pending until the deadline plus allowance passes. It is not the fully shifted mature acceptance window from the design document. Completion summaries cover at most the latest 1000 returned subjects; row limits are exposed. Each query captures one evaluation time. The backend is read-only for queries and uses bounded result and execution limits.

**Evidence is always labelled `demo-unverified`.** There is no reconciled acceptance ledger, durable outbox, sampling estimator, or coverage proof. The sample intentionally exposes the limitations of counting telemetry events. Do not promote its success percentage to a production SLO.

The alarm is a single-window **preview computed on demand**. There is no scheduler, multiwindow policy, persisted alarm state, notification delivery, or IaC reconciler in this version. Source mappings are fixed to the pinned exporter’s baseline schema. The UI supports grouping and subject navigation, not automatic statistical cohort discovery or a complete trace waterfall.

Redis Streams are real in Compose, but the sample consumer has no pending-entry reclamation after crashes. Data can remain pending after a restart. Use a volume reset for a clean demo. Terminal telemetry and queue acknowledgement are not transactional. All traces are retained by the sample SDK; there is no collector sampling processor.

## Development checks

```sh
python3 -m venv .venv
. .venv/bin/activate
pip install -r requirements-dev.txt
python -m unittest discover -s tests -v
python tests/local_flow.py
```

The unit suite checks validation, parameter handling, no-data/unknown behavior, and real compiled SQL in embedded ClickHouse (`chdb`, bundled engine 24.8). The local flow test runs real HTTP and gRPC services against a fake Redis TCP server and decodes emitted protobuf OTLP logs, traces, and metrics. It needs free localhost ports 8081 and 50051.

`scripts/smoke.py` is the separate full-stack acceptance check for the real Redis, collector/exporter, and ClickHouse 25.3 containers. Run it after Compose startup. See `VALIDATION.md` for checks performed when this package was created.

## Stop and reset

```sh
docker compose --profile hyperdx down
# Explicitly discard demo data for a clean start:
docker compose --profile hyperdx down -v
```

## References

- [ClickStack Compose deployment](https://clickhouse.com/docs/clickstack/deployment/docker-compose)
- [ClickStack table shapes](https://clickhouse.com/docs/clickstack/ingesting-data/schemas)
- [Pinned upstream ClickHouse exporter configuration](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.123.0/exporter/clickhouseexporter)
- [OTel messaging spans and links](https://opentelemetry.io/docs/specs/semconv/messaging/messaging-spans/)

[Observed operation conformance and executable developer journey](docs/OPERATION_CONFORMANCE.md)
connect source manifests to actual woven SDK spans and invocation overviews.

[Live woven Go operations](docs/LIVE_OPERATIONS.md) now reach the collector,
ClickHouse and the investigation UI, with an executable browser storyboard.

[Operation heatmaps and trace drilldown](docs/OPERATION_DRILLDOWN.md) retain the
selected evidence through error counts, duration cells and trace waterfalls.

[Scoped subject navigation](docs/SUBJECT_NAVIGATION.md) follows declared span
events across related traces while separating tenant and identity namespaces.

[Declarative operation packages](docs/OPERATION_PACKAGES.md) now drive shared CLI/UI
bindings and views for both checkout and a separate report-generator application.

[Architecture and process diagrams](docs/ARCHITECTURE_DIAGRAMS.md) follow source/build
contracts through standard OTel capture into package-based investigation, and show
the separate portable measurement compiler boundary.

[Native counter rates](docs/NATIVE_COUNTERS.md) add reset/gap-aware Go evaluation
and YAML-driven live rate charts, with explicit unknown boundaries.

[Native duration histograms](docs/NATIVE_HISTOGRAMS.md) add exact bucket deltas,
per-stream heatmaps, P95 bucket bounds and explicit aligned merging in Go/CLI.
Journey 14 captures real telemetry and contextual span navigation.

[Operation error ratios](docs/OPERATION_ERROR_RATIOS.md) derive calls/errors from
one aligned counter result and guide contextual span/trace investigation.

[Declared fleet membership](docs/FLEET_COVERAGE.md) now exposes missing-member
coverage explicitly. Exact fleet totals require compatible aligned intervals
from every expected instance; the default lab now runs two instances and separately compares their actual windows.

The second checkout instance runs at loopback **8092**. See
[live fleet comparisons](docs/FLEET_COMPARISONS.md) for explicit tolerance and
metric-export loss/restart controls. Exact totals retain strict boundaries.

Saved operation evidence can now be reopened and refreshed as immutable revisions: [Saved investigations](docs/SAVED_INVESTIGATIONS.md). See the [general observability feature gaps and proposed order](docs/PRODUCT_GAPS.md) for the path toward broader HyperDX/Honeycomb-style exploration.

[Trace findings](docs/TRACE_FINDINGS.md) connect bounded span-page examples to cross-service trace context and two explicit evidence captures, staged with an authored finding before Save revision. Service pages prioritize the overview and operation diagnosis, with supporting inventory and instruments in disclosures.

[General exploration](docs/EXPLORATION_DESIGN.md) now searches captured spans/logs by discovered attributes, groups full matching record populations independently of row pages, expands traces across services and finds exact linked logs. The [JSON example](examples/exploration/errors.json) uses the same compiler through `goo plan-exploration -query examples/exploration/errors.json`.

[Core workbench basics](docs/CORE_WORKBENCH.md): full-population span measures/time charts, reusable questions, service signals/reported dependencies and query controls.

[Duration breakdowns](docs/GROUPED_MEASURES.md) rank full-population latency by service or attribute and retain it with saved investigations.
[Error and latency comparisons](docs/ERROR_COMPARISONS.md) compare fixed windows, follow changed groups into traces/logs and save both populations.
[Deployment context and investigation notes](docs/WORKBENCH_CONTEXT.md) preserve event/custody times, ordered hypotheses and exact references to saved evidence.

Metric discovery in **Explore → Measurements** builds editable packages from
ordinary OTel producers without a goo manifest. See the
[discovery contract](docs/MEASUREMENT_PACKAGES.md#discover-instruments-without-a-goo-manifest)
and [executable journey](journeys/metric-discovery.spec.js).

Explore supports [resumable question links](docs/QUESTION_LINKS.md): share a fixed
question, restore it without a source read, then explicitly Run again. Use captures
to share immutable evidence.

Saved question histories support [capture comparison](docs/CAPTURE_COMPARISON.md):
compare full record summaries and exact native gauge endpoints without rereading
telemetry, with explicit scope, definition, unit and missing-data checks.
[Save a comparison and finding](docs/SAVED_COMPARISONS.md) in an immutable
investigation revision, then share its exact link; reopening preserves the saved
decision without reevaluating it.

[Complete native windows](docs/NATIVE_WINDOW_COMPARISON.md) add exact counter-rate
and histogram bucket-population comparisons, with full coverage, reset/gap and
layout checks. Versioned saved receipts preserve the resulting decision.

Native measurement APIs and shared question steps retain optional exact nanosecond
scopes; see [scope transport](docs/NATIVE_WINDOW_COMPARISON.md#exact-scope-transport).
The [comparison panel discovers supported windows](docs/OBSERVED_WINDOW_DISCOVERY.md)
without a source read and captures explicitly selected sub-scopes.

Comparison sides can become [reusable question cards](docs/REUSABLE_COMPARISON_QUESTIONS.md)
with their exact scope and captured evidence already linked. Save a revision to
share, reopen without a source read, or explicitly rerun the fixed question.

Saved questions can [compare with a fresh window](docs/FRESH_COMPARISON.md): preview
the proposed scope, choose frozen prior evidence and explicitly capture once,
then save the comparison as a finding in the investigation.

## Captured change breakdowns and product direction

[Captured group comparisons](docs/COMPARISON_BREAKDOWNS.md) reuse full stored
aggregates and open focused questions. New dimensions require explicit recapture.
[Whole-product review](docs/WHOLE_PRODUCT_REVIEW.md) guides automatic service
overviews and distributed relationships within Services. The first bounded slice
is implemented; remaining graphical topology and broader dashboard recommendations
are planned. Observed, declared and executed-supported evidence remain distinct.

## Automatic service overviews

[Service dashboards](docs/SERVICE_DASHBOARDS.md) now derive bounded typed widgets
from observed signals, convention keys and native instruments. Matching registry
contracts retain declared revision IDs. Services also offers a local distributed
view of resolved parent relationships and diagnostics, with fixed-scope handoffs
into questions and immutable investigation evidence. Async/declaration topology,
whole-dashboard snapshots remain future work. Declared services without observed
telemetry now remain visible in Services, with independent registry provenance.

Service discovery now separates same-named services by OTel namespace, including
missing and present-empty values. Dashboard cards can explore spans matching
registered operation bindings. See [service identity](docs/SERVICE_IDENTITY.md)
and [journey 60](docs/journeys/README.md#60-namespace-service-identity).

[Reported-edge investigations](docs/REPORTED_EDGES.md) now stage exact child-span
questions from distributed relationships. [Visual comparisons](docs/VISUAL_COMPARISONS.md)
put aligned baseline/current values ahead of detailed provenance and assumptions.

The [display component library](docs/UI_COMPONENTS.md) now provides framework-neutral
Web Components for compact latency, exact label/value lists and grouped metric
comparisons. The standalone `/ui-components` reference works without the app runtime;
Services and Explore use the same display primitives.

The [whole-app component audit](docs/UI_COMPONENT_AUDIT.md) prioritizes passive
table viewports and result states, then facets and typed attribute displays.

Explore and Services share passive native-table viewports and explicit result-state
displays from the [Web Component library](docs/UI_COMPONENTS.md). Caller-owned rows
retain focus during patches; SSE transport and live retention remain planned.

The same library now provides local-search facets and typed attribute values in
Explore and trace/log inspectors. Full-population measures retain their original
scales when searching supplied values; filter actions remain explicit source reads.

Applied log questions now offer a [live log tail](docs/LIVE_LOG_TAIL.md) with
bounded retention, pause/resume, follow-scroll and explicit reconnect gaps.
Reusable [trace waterfalls](docs/UI_WATERFALL.md) and
[count-time bars](docs/UI_TIME_BARS.md) support the same framework-neutral display
contracts; charts and tails retain their separate source semantics.

Saved record questions also support an explicit [rolling record-count comparison](docs/LIVE_COMPARISON.md)
against linked frozen evidence. Both sides share a stable axis; source failures
retain a stale result, and freezing saves the displayed snapshot without another
telemetry read. The passive [chart frame](docs/UI_CHART_FRAME.md) preserves native
caller content and focus across window/state metadata updates.

When a saved span question requests duration and recorded status, live comparison
also shows compact mean/p50/p95/p99 summaries on a shared fixed percentile scale,
and recorded Error/total share on a 0–100% scale. All measures share one window,
source aggregate and evidence basis; freezing retains the complete displayed result.
See [the regression storyboard](docs/journeys/README.md#71-a-service-regression-across-count-latency-and-recorded-error-share).

Native counter questions now support [live exact-stream rate comparisons](docs/LIVE_COMPARISON.md#exact-stream-counter-rates).
The current export window retains the baseline's exact duration; resets, gaps,
changed stream identity and stale exports withhold rates. Freeze retains the
actual displayed historical window. This slice requires one complete baseline stream. [Focus an observed stream](docs/EXACT_STREAM_SELECTION.md)
to save/share/capture an exact source binding from a multi-stream question.
Live gauges, histograms and fleet rates remain planned.

Native results also offer [local stream search and coverage diagnostics](docs/NATIVE_STREAM_NAVIGATION.md).
Find an instance or dimension across retained streams, isolate those needing
review, and inspect readable gap/reset explanations with their exact codes.
Local browsing preserves the measurement's window, totals and evidence basis.

Default [service dashboards](docs/SERVICE_DASHBOARDS.md#native-service-cards-and-exact-stream-questions)
show evaluated native interval values alongside coverage, with exact-stream
question links and source-free browsing of the retained inventory. Counters,
gauges and native histogram ranges remain independent per stream; no manifest
or fleet aggregation is required.

Service overviews now offer [native instrument selection](docs/SERVICE_DASHBOARDS.md#choose-native-instruments-for-this-service-visit):
stage up to two supported observed definitions, then explicitly apply and reread
the same fixed scope. Empty selection and restoring defaults are supported;
selected questions use ordinary investigation and capture custody.

[Native timelines](docs/NATIVE_TIMELINES.md) now show independent gauge samples,
evaluated counter intervals and histogram interval counts in ordinary results
and captures. Keyboard/pointer inspection is read-free; unknowns stay distinct
from zero and histogram P95 remains a native bucket range.

## Native histogram heatmaps

Inspect retained intervals and compatible captured windows with [native bucket heatmaps](docs/NATIVE_HEATMAPS.md). Counts, unknowns and incompatible layouts remain distinct.

## Fixed-time histograms and component reuse

Histogram packages can opt into [Go-backed time buckets](docs/HISTOGRAM_TIME_BUCKETS.md), with local export/fixed-time inspection and replayable captures. The [whole-product component sweep](docs/WHOLE_PRODUCT_COMPONENT_SWEEP.md) prioritizes shared scope/evidence summaries, selected-item inspectors, existing component adoption and larger investigation compositions.

## Shared evidence summaries

Native captures and linked investigation evidence now share a [compact evidence summary](docs/EVIDENCE_SUMMARIES.md), with explicit population, scope, counts, result mode and separate source/storage clocks.

Service widgets and saved native evidence share [compact evidence summaries](docs/EVIDENCE_SUMMARIES.md): fixed scope, independently labelled read clocks, and readable UTC timestamps with exact nanoseconds available in tooltips. Latency summaries use compact inline Mean/p50/P95/P99 values, with the positional rail and exact values in Latency details.

[Comparison views](docs/COMPARISON_PRESENTATION.md) align baseline/current values on shared scales and group exact scope and read clocks into disclosures.

[Selected-item inspectors](docs/SELECTED_ITEM_INSPECTORS.md) share compact trace/log/native detail layout, retained local selection state and close-to-opener focus.

[Core experience refinements](docs/CORE_EXPERIENCE_REFINEMENTS.md) connect retained
investigation visits, compact service diagnosis, chart-population comparisons,
trace highlighting, deployment context and capability-to-binding handoffs. These
are bounded first slices of the [product survey](docs/CORE_EXPERIENCE_PRODUCT_SURVEY.md),
which records verified vendor sources and Honeycomb research access limits. The
workflow document links executable journeys, screenshots and current limits.

The [population workspace](docs/POPULATION_WORKSPACE.md#attribute-breakdowns-and-comparisons)
also offers discovered-attribute breakdowns, full-population count/error/latency
ranking and shared-scale pinned-value or previous-window comparisons. Saved views
retain ranking and baseline questions; reopening stages configuration without
reading telemetry. Run explicitly to refresh each independent population.

The [population live area](docs/POPULATION_WORKSPACE.md#watching-a-population)
can explicitly tail ordinary span/log records, pause for inspection, and compare
full rolling populations against a captured baseline. Saved settings reopen
without starting subscriptions; retention and continuity gaps stay visible.
