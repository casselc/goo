# Automatic operational dashboards — product initiative

Status: **C2a HTTP/legacy-RPC and C2b database-client/messaging roles delivered**.
Later stages remain proposed. Clarified by the user on 2026-10-25 after the whole-product review.
This is the expanded C2 initiative in the
[current roadmap](ROADMAP_2026_10_25.md), not a claim every family is implemented.

## Desired experience

Connect ordinary OTel data, discover services and resources, then open a useful
default dashboard without authoring queries or supplying a goo manifest. A web
service should suggest an incoming-work RED view; a database dependency should
show operations and latency; a monitored pool, host or Kubernetes resource should
show the appropriate resource signals. The dashboard is an investigation entry:
select a time interval, operation, resource or population and continue into the
existing logs, spans, trace search and comparison tools.

RED means rate, errors and duration. USE means utilization, saturation and errors.
They are useful organizing methods, not a reason to invent missing measurements
or blend incompatible populations. Common service pages can offer Incoming work,
Dependencies, Resources and Logs as contextual views where supported; users
should not need to build a dashboard or navigate a catalog of implementation
families before seeing their own data.

Discovery should automatically select relevant defaults and offer applicable
views. Selecting a discovered entity can load a small bounded default overview
as Services already does. Loading every family, every instrument and every entity
in the background is not required. Local switching can use held data; additional
source execution must remain explicit in the existing owning workflow and its
read policy. Do not turn a metadata-only preview into a hidden population read.

## First HTTP slice — implementation contract

The service planner accepts an optional `profile`: `auto`, `http-server` or
`all-spans`. Omitting it preserves the existing v2 plan, including its mixed-span
defaults. The service UI opts into `auto`; a new service visit starts automatic,
while range and native-instrument changes retain the chosen view. A view switch
uses the existing cancellation/generation owner and performs fresh bounded reads.

An HTTP view is a **candidate**, suggested by the supplied current
`http.request.method` attribute key or a supported HTTP server duration histogram.
A flat attribute inventory cannot prove that the method and Server kind occurred
on the same record. Automatic selection additionally uses a positive matching
server-span observation from the existing held service page, or an eligible
native HTTP duration instrument. Without either, All spans stays the default and
the HTTP candidate remains available to select. A bounded page without a witness
does not establish that HTTP work is absent. Actual span questions require the reported kind `Server`
and a present, nonempty `http.request.method`, within the original service,
namespace, environment and fixed window. Legacy `http.method` alone stays in
All spans; this raw-span profile does not union aliases or certify a convention.
Missing and empty routes remain separate groups. Recorded span Error is distinct
from HTTP response status and business success.

No matching server spans leaves an explicit empty incoming view and an All spans
action; there is no hidden fallback population read. Native HTTP duration results
retain their own exact streams, units, intervals and source basis. Purpose ranking
places relevant supported duration instruments ahead of unrelated defaults, with
the existing two-instrument limit and explicit override. Generic, log and declared
operation widgets retain their own populations rather than becoming HTTP metrics.

C2a.1 is implemented and checked against ordinary OTel data. Actual source,
browser, actor and independent verification evidence is recorded in
[VALIDATION.md](../VALIDATION.md) and [journey 187](../journeys/187-http-server-dashboard-storyboard.md).

## Incoming RPC slice — explicit legacy binding

C2a.2 is delivered. The first RPC view uses a named legacy observation profile,
`goo/rpc-server-legacy-span-observations/v1`: exact reported `Server` kind and
present, nonempty `rpc.system` and `rpc.method` attributes. It does not union
`rpc.system.name` or newer method names into that population. Custom nonempty
values remain reported values rather than verified convention compliance.

Flat key discovery suggests a selectable RPC candidate. Automatic admission also
requires a positive same-row server/system/method observation from the existing
scoped service page, or a compatible legacy native instrument. HTTP keeps its
existing automatic priority; RPC follows when independently admitted, ahead of DB and messaging.
All spans remains the fallback. Manual selection and range/native changes retain the resolved profile.
No absence is inferred from a bounded metadata page, and no fallback read is hidden.

The relevant native binding is `rpc.server.duration`, histogram, unit `ms`, with
existing typed identity/temporality validation. A same-name seconds instrument
is not silently converted into this binding; newer call-duration names need a
separate explicit profile. Native results keep their own stream populations and
interval coverage. Recorded span Error, gRPC numeric status, HTTP status and
business success are separate facts; this cut does not invent a protocol-wide
failure policy or a unique-call denominator.

Method groups aggregate a reported method value across RPC systems/services
within the selected service and parent population. Drilldown preserves that
parent predicate and adds the exact method value; it does not manufacture a
globally unique operation identity. The existing compact displays,
scoped questions, trace waterfall and retained Back flow are reused.

## Database-client slice — current-key dependency observations

C2b.1 is delivered. The `db-client` profile requires exact reported `Client`
kind and present, nonempty `db.system.name`. Legacy `db.system` alone stays in
All spans; it is not silently unioned into the current-key population. A positive
same-row, same-scope observation from the existing held page, or compatible
native `db.client.operation.duration` histogram in seconds, admits automatic
selection. Existing HTTP and RPC admission keep their priority. Flat field
metadata offers a candidate but does not prove Client/key co-occurrence.

The bounded default breakdown chooses observed `db.operation.name`, then
`db.namespace`, then `db.system.name`. Missing and empty optional labels remain
separate groups; choosing a breakdown does not add a population filter. These
are literal value populations across reported systems and database namespaces,
not globally unique operations or monitored database entities. Existing question
filters/grouping support explicit further scoping without loading every dimension.

No automatic `db.query.summary`, `db.query.text` or `db.statement` grouping is
introduced. Already-reported attributes remain available through existing explicit
record inspection and exploration; the tool does not require raw SQL, sanitize
arbitrary query text or promise that producer-supplied fields contain no secrets.
The page/aggregate source contracts remain unchanged.

Activity, recorded span Error and inclusive client duration describe observed
dependency work. They do not establish unique-query throughput, database-server
health, connection-pool saturation, affected rows or business success. The native
seconds binding retains exact stream identity, temporality and interval coverage;
same-name metrics with another unit remain separate ordinary instruments.

## Messaging slice — role-specific span observations

C2b.2 is delivered; actual acceptance and frames are in
[journey 190](../journeys/190-messaging-dashboard-storyboard.md) and
[VALIDATION.md](../VALIDATION.md).
The `messaging-consumer` and `messaging-producer` profiles require exact reported
`Consumer` or `Producer` kind respectively, with present, nonempty
`messaging.system`. They do not infer roles from span names or operation labels.
A supplied field key offers a candidate; automatic selection additionally needs
a same-row, same-scope positive witness from the existing held service page.
The exact eligible `messaging.process.duration` seconds histogram also admits
the consumer candidate. HTTP, legacy RPC and DB retain priority, followed by
consumer, producer and All spans. This deterministic preference is not severity.
An accepted profile stays selected through range and native changes; a matching
zero population never triggers a hidden fallback read.

The default breakdown chooses observed `messaging.destination.name`, then
`messaging.operation.name`, then `messaging.system`. Optional missing and empty
values remain distinct; grouping does not exclude them. Destination and operation
values describe populations across reported systems and other attributes, not
unique queue or operation identities. Existing explicit filters refine them while
preserving the role, service, resource namespace, environment and fixed window.

Consumer spans may describe receiving or processing; producer and consumer record
counts are neither unique messages nor acknowledgements. Recorded span Error is
not a broker outcome policy. Native process duration has its own exact stream,
interval and operation population; it is not an aggregate of every Consumer
span. No producer native binding is added in this cut. Other reported instruments
remain available through existing explicit selection. Reported asynchronous links
remain producer data: these views do not join messages or infer delivery, backlog,
queue availability, retry uniqueness or exactly-once behavior.

## Dashboard families and useful defaults

| Family / entity | Useful default panels | Evidence required and fallback |
| --- | --- | --- |
| Web/API service and incoming RPC work — RED | Traffic over time, recorded errors/status, duration distribution/heatmap, route or operation ranking, slow/error trace examples, relevant deployments and dependency links. | Explicit server/protocol population or supported request metrics; keep observed span-record rates distinct from deduplicated request or native metric counts. RPC is a separate explicit profile. With spans only, still provide a useful span-backed view. |
| Database client/dependency | Operation rate, errors and latency by database/system/operation; slow query summaries, caller services and representative traces. | Standard database spans and/or eligible `db.client.operation.duration`. Use sanitized low-cardinality summaries; raw SQL is not required. Client telemetry describes a dependency, not database-server health. |
| Database pool or monitored database server — USE where supported | Connections by state, utilization against an actual capacity, acquisition wait, pending/timeouts; server CPU, memory, I/O, locks or cache signals when exported. | Pool identity/capacity and compatible typed metrics; actual server/exporter signals for server panels. Operation latency alone is neither pool saturation nor server contention. Partial pool/client views remain useful. |
| Messaging producer/consumer | Send/process rates, processing duration, recorded failures, destinations and consumer groups, retries/redelivery where reported, trace links across asynchronous work. | Explicit operation/role profiles, matching processing metrics and reported attributes. Keep producer, processing attempt, message and acknowledgement grains distinct. |
| Queue/broker — USE where supported | Backlog/lag, oldest-work age or wait, active/capacity, broker resource pressure, rejected/failed operations. | Actual queue/broker instruments and identity. Processing spans do not supply queue depth or broker availability; demo-only `lab.*` instruments are not generic defaults. |
| Host/container/runtime — USE | CPU and memory use, capacity-relative utilization, runnable/pending work or pressure, I/O/network activity and errors, GC/runtime trends where reported. | Resource identity, compatible native gauges/counters and explicit capacity/denominator. CPU use alone is not saturation; memory usage alone is not a utilization percentage. Finite fractional gauges already work. |
| Kubernetes cluster → namespace/workload → pod/container, with node views | Resource inventory and changes, CPU/memory versus requests/limits, readiness/restarts/pending state, node pressure, workload traffic/errors/latency, pod logs and traces. | Reported Kubernetes attributes plus actual kubelet/cluster/exporter metrics and ownership facts as needed. Scope by cluster/namespace/resource identity; show reported associations separately from verified owner relationships. Service inventory is not a cluster inventory. |

A discovered service may have multiple roles. Show separate relevant views rather
than forcing it into one guessed service type. Database and broker endpoints seen
from clients are reported dependencies; monitored server entities require their
own evidence. Infrastructure views must also work for resources with no
`service.name` at all.

## One capability-driven planning contract

Extend the existing service planner and versioned profile/catalog machinery.
Each family should declare:

- Entity identity and scope: environment, logical service namespace or Kubernetes
  namespace as distinct fields, instance/resource IDs and reported relationships.
- Exact source population and grain; required/optional keys and instruments;
  supported convention/version, unit, type, temporality and identity rules.
- Required calculations and their numerical contract: rate/reset behavior,
  distribution aggregation, capacity ratios and resource-state counts.
- Small ranked default panel set, bounded execution/read budget, supplied-data
  presentation and named investigation intents.
- Per-panel state such as supported/observed, not observed, unsupported,
  unavailable or requiring configuration, with an actionable explanation.

A manifest may enrich naming, expected signals or declared operations. It is not
a prerequisite for discovery-driven defaults. Reuse existing passive charts,
tables, field views, scope summaries and inspectors. A general dashboard editor,
new query engine, universal result object or new evidence store is not needed.

Useful panels should appear even when other signals are absent. Put unavailable
prerequisites in compact capability detail, with concrete receiver/exporter/key
requirements; do not surround every useful chart with a wall of warnings or turn
absence into green health. New panels must say what population they measured,
not repeat every semantic caveat above the data.

## Current foundation and actual implementation gaps

Already implemented:

- [Service overview planner](../exploration/dashboard.go), automatic service views,
  bounded widget execution, scoped record/native presenters and investigation
  continuations. C2a/C2b.1 supply HTTP/legacy-RPC server-span and current-key
  database-client defaults, with All spans fallback. Omitted-profile plans retain
  mixed-kind span grain.
- [Native HTTP/messaging capability candidates](../measurements/instrument_capabilities.go),
  gauge/counter/explicit-duration-histogram kernels and exact stream identity.
- [Catalog profiles](../prebuilt/catalog.yaml) for HTTP, messaging, host CPU,
  database client/pool and runtime. Several catalog visualization/calculation
  features are explicitly **proposed**; catalog presence is not executed support.
- Generic record filtering, time/distribution displays, explanations, trace
  inspection/comparison/search and saved definitions/evidence.

Work that must be planned explicitly:

1. Extend the delivered HTTP/legacy-RPC/database-client/messaging selection with
   newer convention bindings and the remaining role/resource families.
2. [Host inventory and explicit gauge/stock authority](HOST_RESOURCES.md) are now
   delivered separately from service-name discovery. Kubernetes ownership/entity
   mappings remain planned; do not manufacture service labels or relax original
   service-stream validation globally.
3. Cumulative nonmonotonic stocks now cover pool counts, host memory and the
   [first Go runtime profile](GO_RUNTIME.md). Delta-stock reconstruction remains
   separate; their temporality cannot be
   substituted with a gauge or monotonic counter.
4. Fractional counter semantics where required, such as CPU-time seconds. Native
   gauges already accept finite Float64 values; this is not a blanket missing
   float-metrics engine. Current counters accept integral nonnegative values
   below 2^53 from Float64 exports.
5. Any histogram family/unit beyond current duration distributions, if an actual
   panel needs it; do not silently treat byte distributions as latency histograms.
6. Kubernetes identity/ownership/state and capacity mappings from an actual
   selected source. Requests, limits and current use require compatible resource
   identity, time and units before ratios or workload aggregation.

Keep current and legacy convention profiles explicit. Existing native HTTP rules
and prepared catalog rules pin different versions; discovery is not proof that
all observations conform to the same version.

## Staged delivery and acceptance

| Stage | Concrete result | Acceptance before advancing |
| --- | --- | --- |
| C2a — shared family contract + web/RPC RED | C2a.1: discover a service and load a useful HTTP server-span dashboard with relevant eligible native request metrics. C2a.2: add the incoming RPC profile and method/status/duration panels with its own semantics. | Mixed span kinds, missing/empty keys, namespace collisions, manifest-free data, span/native separate grains and clocks, bounded automatic defaults, record/trace drilldown, desktop/mobile whole task. |
| C2b — messaging + database dependency | Discover standard producer/consumer and database-client telemetry and open role-specific dashboards using existing kernels where supported. | Actual processing/client instruments or spans; operation/role separation, sanitized query summaries, asynchronous trace links, partial-signal usefulness and correct per-family source budgets. Do not claim broker/server/pool health from client spans. |
| C2c — pool/runtime/host resource panels | Implement the typed stock/counter/identity/calculation prerequisites needed by a selected real pool/runtime/host source, then ship useful USE panels. | Actual fractional gauges, relevant sum temporality and counter tests, capacities/units/identity, reset/gap/missing behavior, independent source parity and real resource-to-record investigation. Reuse supported gauges now rather than block them on fractional counters. |
| C2d — Kubernetes operational dashboards | Discover cluster/workload/pod/container/node entities from a chosen OTel receiver/resource source and ship bounded resource/state/workload views. | Real inventory/ownership evidence, absent service identity, cluster and namespace separation, requests/limits/state/pressure semantics, receiver gaps, scoped logs/traces and workload aggregation. Introduce identity/discovery prerequisites as part of this delivery. |

This is the product direction, not an instruction to implement every stage at
once. HTTP is the first vertical slice, not the final scope of the initiative.
Kubernetes discovery design can proceed in parallel once its actual source and
resource-identity contract are selected. Native-to-record context (C3) and lean
aggregate execution (C4) support these dashboards; they should not postpone useful
panels indefinitely or be expanded into unrelated framework projects.

C2c now adds [automatic named-pool connection observations](POOL_CONNECTIONS.md) over [native stocks](CUMULATIVE_STOCKS.md). Same-stream pinned metadata derives an explicit count-domain question, with separate per-instance states and frozen replay. The original count-only view does not infer capacity, acquisition wait, saturation or pool health. Related-signal inspection and strict reported-ratio composition are delivered below; runtime/host remains staged.


[Pool capacity and acquisition-wait context](POOL_SIGNALS.md) now stages
independently evaluated exact-stream measurements from a selected named pool.
Default service reads remain bounded; explicit Run and individual capture retain
the original window, identity and separate clocks. Strict reported used / maximum composition is delivered in the compact pool
view below. Broader time/capacity joins, causal attribution and automatic
runtime/host panels remain follow-ons.


C2c's [compact pool view](POOL_VIEW.md) composes held counts and explicitly loaded
capacity/wait with exact per-pool identity and raw replay. A reported used /
maximum fraction requires matching latest observation time and nonzero epoch,
positive maximum and a valid bounded numerator. Missing, stale and incompatible
signals retain separate values/reasons; no states are summed and no health is
inferred. Default reads remain unchanged. The initial host resource identity path is delivered below; runtime/GC and
Kubernetes discovery remain family follow-ons; pool-to-record scope
expansion remains separate. Acceptance is recorded in
[journey 198](../journeys/198-pool-view-storyboard.md).


C2c now includes [host-resource observations](HOST_RESOURCES.md): explicit host
identity without service/instance fabrication, bounded host-bearing inventory,
reported per-core CPU utilization and memory-stock defaults, matching-resource
record context and existing native preservation. [Journey 199](../journeys/199-host-storyboard.md)
records real ordinary OTLP acceptance and source costs. This closes the first
host gauge/stock identity prerequisite; runtime/GC panels, fractional CPU-time
counters, derived capacity calculations, missing-environment discovery and
Kubernetes resource/ownership mappings remain separately scoped follow-ons.


### C2c — first Go runtime observations (200)

Source `a98b9a9` adds [pinned Go runtime observations](GO_RUNTIME.md) to the
existing service dashboard: exact-instance memory-used and GC heap-target stocks,
with goroutines explicitly selectable. The existing two-native-read budget,
role/pool precedence, native inspection and individual capture owners remain.
Runtime names alone cannot admit an automatic broad read. Candidate expansion
is bounded and omissions disclosed; memory categories and process identities
remain separate. Source/property/API and ordinary OTLP/browser, owner, actor and independent
raw/custody gates pass; see [journey 200](../journeys/200-runtime-storyboard.md)
for exact costs, actual frames and validation limits. This closes the first Go runtime stock-profile slice, not GC pauses,
pressure classification, cross-runtime support, capacity calculations or
Kubernetes discovery. Those remain separately scoped work.


### C2d — Kubernetes resource discovery (201)

Source `bad6727` implements the [initial Kubernetes discovery path](KUBERNETES_RESOURCES.md):
full reported resource variants across five OTel source families, local cluster/
namespace filtering and explicit matching-resource record investigation. Missing
cluster UID remains unverified context; names never manufacture pod identity.
Grouped source bounds refuse whole over-budget inventories, and presentation
omissions remain explicit. Shared request/tab primitives preserve host behavior
and withdraw pending service work while retaining completed views. Source,
property, API, physical SQL and ordinary OTLP/browser/ownership/native gates pass;
[journey 201](../journeys/201-kubernetes-storyboard.md) records exact costs and
separate display/actor receipts.

This delivers the discovery prerequisite, not the whole Kubernetes dashboard
family. The next C2d slice must select an actual receiver metric profile and
implement explicit pod-resource native measurement authority before useful
memory/state defaults. CPU/request-limit calculations, ownership facts, readiness,
resource changes and workload aggregation retain separate contracts.
