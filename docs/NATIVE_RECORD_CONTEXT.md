# From a native interval to nearby record evidence

The implemented pure planner `exploration.PlanNativeRecordContext` maps a selected
native stream and exact interval into a typed logs or spans question. The first UI
slice offers **Logs in interval** and **Spans in interval** from counter and
histogram interval inspection. These are contextual observations sharing resource
identity and event time, not proof that particular records caused a metric value.
They are not exemplars, distributed-trace completeness or a metric-to-record join.

## Preserve exact time membership

The mapper receives the original validated measurement question, selected exact
stream, record signal and canonical integer nanosecond bounds `[fromNS,toNS)`.
The interval must be nonempty and inside the native question's frozen evaluation
window. A previously pinned exact stream cannot be replaced. Established service,
environment and instance identity must validate before a question is produced.

Record source windows are expressed in milliseconds, so the mapper encloses the
interval with `floor(fromNS / 1,000,000)` and `ceil(toNS / 1,000,000)`. It also adds
integer `timestamp_ns >= fromNS` and `timestamp_ns <= toNS - 1` predicates. The
coarse source window therefore does not admit adjacent submillisecond records.
The enclosing window must satisfy the existing 120-minute record-query limit;
invalid or oversized choices are refused, not rounded into a wider investigation.
No current-time or arbitrary lookaround fallback is inferred.

The pure API can represent a gauge point only as `[atNS,atNS+1)`. Gauge context
actions are deferred in this UI slice: a point is not silently expanded into a
measurement interval.

## Share only established resource identity

The record question preserves the measurement environment, exact selected
`service.name`, resource `service.namespace` and `service.instance.id`. Missing
namespace is a missing predicate, present empty namespace is an empty-value
predicate, and a named namespace is exact; these do not collapse. Other instances
of the same service are excluded.

Metric dimensions, instrumentation scope/version, resource schema, metric schema
and unrelated resource attributes do not become record predicates. The original
metric stream retains those fields for native identity and validation, but their
presence on a metric does not establish that spans or logs carry them. The mapper
does not copy metric-only operation, method, status or bucket attributes into a
record population. Record pages remain bounded independently of full aggregates.

## Stage, execute and return

Selecting an interval and choosing a record action stages the exact typed question
without querying records. Only explicit **Run** reads the logs or spans source.
The existing record investigation can inspect a row, expand its trace, capture
evidence or refine the question under its normal read and custody rules. A trace
expansion has its own source reads. The existing trace action uses the reported
trace ID and the original **coarse millisecond envelope**, intentionally dropping
the exact nanosecond, instance and namespace predicates. Its context explanation
retains the exact originating record question and evidence basis, and explicitly
discloses the broader population. A trace can therefore contain same-trace records
excluded by the native-context question. Sharing a native interval does not
establish a causal relationship with that trace.

The native parent, selected stream and interval remain retained navigation context.
Returning uses the held parent evidence rather than reevaluating the metric.
Closed, hidden or superseded action owners must not stage or publish a stale
question. Capturing a record question follows existing explicit capture semantics;
it does not freeze a combined native/record transaction or promise simultaneous
reads. Sharing or saving the child record question preserves its exact filters,
but the retained native question, stream, basis and read time are tab-local origin
context, not a durable native citation. A native capture is a separate explicit
read with its own immutable custody.

## Verification boundary

Independent Hegel properties use integer raw facts and `big.Int` arithmetic to
check exact nanosecond membership and enclosing millisecond bounds, including
values above JavaScript's exact integer range. They cover missing/empty/named
namespaces, other instances, both record signals, counter/histogram parents,
input immutability and the strict shared-field set. Finite negative cases cover
noncanonical and reversed/out-of-window bounds, invalid signals, missing identity,
pinned-stream replacement, oversized windows and gauge lookaround refusal.
The first focused campaign passed in 0.037s.

[Journey192](../journeys/192-native-record-context-storyboard.md) accepted an exact
115-record contextual population, then an explicitly broader same-trace population
of four spans and four logs. Return restored the held native evidence and focus
without rereading. Four additional actual scenarios cover empty/missing namespace,
an interval with an unknown native value, and an empty record interval. A held
real record response and hidden/closed actions could not take ownership after
Return; its cancellation POST is a control, not another telemetry read.

The bounded native acceptance passed in 6.878s, reusing the existing
`population_drilldown` model for stage/back/read/history depth and counters. Its
ten replay steps derive from nine observed UI events (trace expansion reads spans
and linked logs separately). Independent raw-fact checks handle exact population
and scope; the finite model does not prove all schedules or interval arithmetic.
Six semantic and two finite-model mutations were rejected, alongside four changed
case counts. The initial native observer incorrectly required a single stream;
the retained inspector's actual Raw record disclosure resolves the selected exact
stream among three namespace streams. The rejected diagnostic is retained without
changing application data. The main browser's initial failure was an obsolete
Expand-button selector; its corrected accepted run changed the harness only.

The 25.8s browser actor covered eight phases and 418 snapshots with two source
operations. A separate native-entry frame receipt used one fresh native read;
selecting its interval was source-free. These are bounded local checks, not
production metric-to-record association or arbitrary schedule proofs.
