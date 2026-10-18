# Shared measurement packages

Implemented: `goo/measurement-package/v1alpha1` binds bounded OTel observations
in ClickHouse to the Go gauge, counter, native histogram or lifecycle kernels.
This is distinct from `goo/telemetry-package/v1alpha1`, which describes captured
span/log populations. Both use the same immutable general captures and ordered
investigation notes. The UI exposes **Explore → Measurements**, Definitions and
an Examples measurement library; there is no reservation-specific evaluator in
this path.

## Author, preview, run, preserve

Select an existing package or paste YAML into the package editor. Preview resolves
a fresh lookback or explicit fixed UTC bounds and validates without reading
telemetry. Run explicitly reads that fixed question. Changing the definition or
window invalidates the applied result; cancellation/navigation prevents late
responses from publishing evidence. Save package revision stores canonical YAML
under its content hash. Export YAML provides the same portable definition.

Capture investigation rereads the fixed question on the server, stores its plan,
complete bounded source input, derived result, source read time and evidence basis.
Reopening a capture performs no telemetry or custody read. Refresh creates a child
with the same definition, fixed bounds and compiled plan. Old receipts are retained.
Create measurement package / Continue live reuses the definition with a newly
resolved window. Notes reference the exact capture, not an implicitly refreshed
question. Measurement inputs cap at 10,000 rows / 8 MiB source response; source
limits fail instead of evaluating a partial population. Capture envelopes retain
the existing 24 MiB bound. Lifecycle subjects page by 100; streams page by 25; metric tables show the
latest 200 observations/intervals, with full bounded inputs retained in captures.

```yaml
apiVersion: goo/measurement-package/v1alpha1
name: queue-lag
title: Queue lag by consumer group
kind: gauge
environment: local
lookbackMinutes: 2
maxGapMs: 30000
grain: consumer-group
source:
  services: [lab-consumer]
  scope: goo.worker-queue
  version: 0.1.0
  schemaURL: goo://demo/queue/v1
  metric: lab.queue.depth
  unit: '{message}'
  attributes:
    messaging.destination.name: reservations
    messaging.consumer.group.name: inventory
```

`maxGapMs` is gauge freshness or the native interval gap limit. Lifecycle
classification instead uses its explicit `deadlineMs` and `allowanceMs`.

Source columns and typed parameter SQL belong to the adapter. Definitions cannot
supply SQL. Aliases, merge keys, unknown fields and multiple YAML documents are
rejected. Custom package revisions use the same schema as built-ins.

## Measurement contracts

| Family | Meaning and boundaries |
| --- | --- |
| Gauge | Independent timestamped observations per complete stream identity. Conflicting same-time values and no-recorded-value flags are unknown. Freshness is evaluated at the fixed upper bound, not the current wall clock; old values remain visible as stale. Dots are not interpolated; no rate or sum is inferred. |
| Counter | Existing reset-aware monotonic cumulative/delta kernel. Separate resource instances, start-time resets, missing boundaries and gaps remain explicit. Float64-exported integral counters must be below 2^53; no rounded large-integer claims. |
| Native histogram | Existing per-stream interval reconstruction with explicit bucket bounds/counts and p95 bucket ranges. Bucket inspection retains exact count strings. No percentile averaging or implicit fleet merging. Current kernel supports nonnegative explicit duration bounds; exponential histograms are not supported. |
| Lifecycle | Explicit accepted/terminal event names and field mappings, tenant set, logical owner, outcome domain, deadline and allowance. Mature cohort, pending/unknown and quality receipts use the shared lifecycle kernel. Subject drilldown retains tenant, subject, environment, service set, event names and fixed source horizon. |

Stream identity includes resource schema, full resource map, instrumentation scope,
version/schema, metric name/unit and all point attributes. Native counter/histogram
packages require an explicit scope name; version and schema may be empty and remain exact stream identity values. Missing required metadata is a
validation failure, not invented identity. Gauge grain documents intended resource
meaning; it does not authorize aggregating repeated consumer-group observations
from several workers. Worker utilization is recorded per instance. Lag is
undelivered group work; pending may include stale unreclaimed entries. Neither
proves saturation. Missing CPU, memory and pool measurements remain uninstrumented.

## Lifecycle event and knowledge time

The source reads accepted/terminal logs over the extended horizon needed for
`[T − D − L − W, T − D − L)`, including later terminals up to fixed evaluation T.
Lifecycle scope/version/schema selectors are not implemented and are rejected,
rather than silently ignored; service/environment/event bindings define source
selection. Exact business nanoseconds are strings. A payload hash records its first Goo
observation in persistent local custody, `goo.local-first-observed/v1`. Repeated
reads retain that timestamp; differing payloads with the same event ID remain
conflicts. Concurrent readers retain the atomically published winning observation;
this is a local publication clock, not a global earliest-arrival guarantee. A
clock inversion/correction that cannot advance its basis fails closed. This is
**not** a producer acknowledgement or warehouse ingestion clock.

The adapter opts into `knowledgeAfterEvaluation: true`: a refresh may discover an
old business event after evaluation T, while preserving the original cohort and
business question. A later knowledge cutoff never licenses future business facts.
The default pure lifecycle policy still requires cutoff ≤ evaluation unless this
mode is explicit. Corrections cannot change that mode, revisions or subject scope.
All live source scopes are `unverified`; absence of completion stays unknown. An
observed late or rejected terminal may be bad, but there is no completeness or
production SLO claim. Older sample records without exact mapped identity remain
invalid diagnostic evidence rather than being silently dropped.

## Defaults and workflows

Twelve built-in definitions cover queue lag/pending, worker active/capacity/
utilization, queue-wait histograms, checkout counters/durations, reservation call
counters, and lifecycle outcomes for reservations, checkout and reports.
The [workflow catalog](../config/workflows.json) is data-only, versioned and
strictly validated. Delay, retry, RED, USE and report flows declare requirements,
ordered package steps and the reason for each step. Workflow references resolve
the repository builtin; a saved custom revision with the same name cannot shadow
that default or silently change its environment. Steps open previews over a
shared fixed two-minute window; they do not automatically run telemetry or assert
findings. Choosing the workflow again resolves a new window. Population/resource
steps deliberately disclose scope changes; subject drilldown carries explicit
identity. These descriptors are repository-authored; workflow authoring/publishing
in the UI remains future work.

The Python example emits exact event IDs/business nanoseconds through its OTel
logging SDK. Both woven Go examples now emit real OTLP/HTTP JSON logs with active
trace/span IDs and explicit subject scope. Their small synchronous, 500 ms bounded
example transport reports export failures without changing business outcomes. It
is not a production logging SDK: batching, retries, credentials, backpressure and
provider configuration remain application responsibilities. Existing span events
remain interoperable. Reports illustrate independent accepted/ready event names.

## Remaining work

These packages are bounded local ClickHouse adapters, not an arbitrary expression
language. CEL/KCL authoring, general arithmetic, dynamic fleet membership,
exponential histograms, automatic source discovery and ingestion completeness
contracts remain future. The existing aggregate compiler can emit Databricks
metric-view definitions over prepared facts; these live measurement adapters do
not yet lower raw OTel gauges/lifecycles into Databricks preparation tables.
No Databricks instance is needed for current execution or verification.


## Discover instruments without a goo manifest

In Explore → Measurements, expand **Discover metrics from an application**,
enter its `service.name` and deployment environment, and discover the last two
minutes of observations. The source adapter scans gauges, sums and explicit
histograms together, capped at 10,000 records, 8 MiB response, 256 MiB source
read and a 15-second request deadline. Overflow fails without a partial catalog.
The catalog describes observed identities and dimension **names**, not dimension
value cardinality, instrument completeness or healthy coverage. It uses no goo
annotations, manifest or service-specific attribute conventions.

**Build measurement package** opens a proposed editable definition using exact
observed scope/name/unit/version/schema bindings. The initial two-minute lookback
and 60-second gap are authoring defaults to review. Empty optional version/schema
values match empty values; they never act as wildcards. Separate versions and
schemas remain separate streams. Supported sums are monotonic counters;
histograms currently support explicit nonnegative duration distributions in
seconds/milliseconds. Actual values, resource-instance identity and interval
quality are validated when run. Unsupported metadata remains visible.

Preview does not read telemetry. Run explicitly, save a package revision, capture
the fixed question and cite it in notes through the existing shared controls.
Native stream summaries now show up to six sorted dimensions; full identity is
available in attribute disclosure. Discovery is intentionally service-scoped;
fleet inventory, dimension-value selection and automatic freshness recommendations
remain future work.

`scripts/independent_metrics.py` is an ordinary Python OTel SDK producer with no
goo manifest and no scope version/schema. Storyboard 39 exercises its actual
collector → ClickHouse → discovery → package → capture → notes path.
