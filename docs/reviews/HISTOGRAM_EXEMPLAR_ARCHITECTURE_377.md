# Reported histogram exemplar references — architecture 377

Status: delivered source contract and bounded source review accepted. Actual
runtime and model checks are recorded in [validation](../../VALIDATION.md).

## Existing source and mathematical boundary

Before 377, `measurements/source.go:Compile` selected explicit histogram count,
buckets, bounds, flags and temporal identity. `metricRow` and `metrics.HistogramPoint`
carry no exemplars. `prepareResult` hashes the complete retained raw rows into the
measurement result basis, while the histogram kernel computes its own basis from
mathematical histogram inputs. Exemplar metadata must remain outside those kernel
inputs and the reusable display declaration/bin snapshot.

The 374 pinned Collector v0.123.0 `convertExemplars` source retains parallel arrays
of time, stored value, trace/span IDs and filtered attributes. It converts OTLP int
values to float64 and absent value type to 0.0; original type/presence and large
integer precision can be lost. Official OTLP v1.11.0 comments define time as the
recorded measurement time and trace/span IDs as optional. Therefore value precision
loss blocks an exact OTLP-value claim, but need not block a valid **reported trace
reference** whose IDs/time/export association are independently retained.

A cumulative source export covers its accumulation epoch. Its exemplars do not
automatically belong to the derived difference interval ending at that export.
Sample time, source export start/end and parent question bounds are three separate
facts. No sampled reference proves contribution, representative latency, causality,
complete trace retention or ingestion coverage.

## Required source gate

Pin the explicit-histogram exporter mapping and actual deployed table columns,
then send genuine points proving nanosecond time, exact lowercase trace/span IDs,
parallel array association and full source stream/export identity. Include integer
precision/type collisions, missing/zero IDs, delta/cumulative epochs, repeated
export end timestamps and malformed/omitted metadata. Prove optional value losses
rather than inferring original types from a stored numeric spelling.

Root selected the prerelease explicit ClickStack adapter contract: standard
nested Exemplars columns are required, like existing ExplicitBounds fields. The
pinned deployed source has them; no speculative old-schema capability probe or
extra per-Run source read is added. Query only the first 32 array items and their
original total, retaining the existing 8 MiB raw source/body budget. A missing
required source column is a source failure. Old/offline Evaluate rows without
metadata are unavailable, never a fabricated known-empty exemplar list.

## Bounded passive metadata arm

The separate optional `Result.histogramExemplars` envelope carries:

```
{version, valueEncoding: "source-float64", originalValueType: "unknown",
 exports: [{sourceRow, stream, startNS, endNS, temporality,
   state: "available" | "unavailable", total, omitted,
   items: [{index, timeNS, traceId, spanId, storedValue}],
   diagnostics: []}], limited, omittedExports}
```

Unknown total/omitted counts use empty strings; known empty arrays have state
available and total "0". Each sourceRow identifies
one retained raw export occurrence, not an evaluated interval index or global
identity. Match the selected evaluated stream/end to exactly one source export;
verify start/temporality/row identity. Multiple matches are ambiguous and withhold
reference actions rather than choosing the first or merging lists.

The implemented limit is 32 items per export and a 64 KiB whole metadata
envelope, with focused boundary tests. Missing entries under a metadata budget mean
unavailable/not retained, never known empty. Decode optional arrays separately
from metricRow's typed mathematical fields. Mismatched arrays, invalid timestamps
or metadata limits produce unavailable diagnostics without corrupting a valid
histogram result or inventing references. A trace action requires an exact nonzero 32-hex TraceId. Missing/all-zero
SpanId remains optional and cannot become a focus/filter; malformed span focus
metadata must not be presented as valid. Invalid/unreported TraceIds remain
inspectable as unavailable link data, never inferred from all-zero IDs.

Uniform value/type/attribute provenance belongs at the envelope, not repeated
prose on every row. Stored float64 values are labeled as such; original OTLP type
is unknown. Filtered attributes are retained exporter strings and never become
record predicates. Adding retained raw source columns may legitimately change the
whole Result basis; equal mathematical points must retain kernel basis/counts.
Measurement Definition/revision and saved histogram display identity stay intact.

## Exact trace staging and existing owner reuse

Use a narrow `relation: "reported-exemplar"` branch in the existing
NativeRecordNavigation stage/Return pipeline. It must not call openTrace or fetch
records automatically. A pure Go/JS builder can validate the same checked parent,
stream, export and reported reference, while the source owner separately verifies
actual membership in held metadata before invoking it.

Stage an ordinary spans question with the parent native fixed window, including
its exact nanosecond enclosing predicates when supplied; preserve checked typed
environment and the exact reported stream namespace (present-empty versus missing
included). Add exact reported TraceId. Drop service/instance/metric attribute
filters deliberately so distributed services within that namespace can appear.
SpanId is reported focus metadata, not a whole-trace span filter. Label this a
reported trace within the retained namespace/window, not a complete distributed
trace. Existing `RecordResults.traceQuestion` drops namespace, so it cannot be
reused blindly for this initial exact question.

The relation branch must replace interval-coincidence prose with reported-reference
provenance and show sample time/source export/parent bounds separately. Existing
native DOM, selected export, basis/read clock and focus remain held; explicit Run
alone reads spans. Back suspends/cancels the actual child owner before restoring
the parent. Generation/visibility/creation-time leases reject stale actions and
late adoption. Sharing the question does not preserve durable exemplar custody.
A subsequent broader trace expansion is a separate explicit action with its own
scope disclosure.

## Acceptance and decision

Independent raw ID/time/export association controls and duplicate-export ambiguity
must pass before source-backed reference actions exist. Test Go/JS question parity,
namespace/typed-policy/window membership and unchanged native counts/kernel basis;
then one genuine selected export → reported reference → Stage → explicit Run →
held Back, with source/store counters and pending/late response refusal. Reuse
existing models for actual observed navigation, not a fixed success script.

Proceed only when required reference facts close. Do not defer every navigation
capability solely because optional sample-value fidelity is lost; disclose that
loss and keep the reference claim bounded. Conversely, valid-looking IDs without
verified source occurrence/export association are not sufficient authority.
No exponential executor, generic correlation engine, new router or registry is
required by this contract.

## Bounded pilot cost recommendation

Prefer the agreed first-32-per-export SQL projection plus existing 8 MiB raw-source
refusal and 64 KiB typed-report budget over a new endpoint/window query. For the
initial reference-only pilot, project time, trace/span IDs, stored float64 value
and original count; omit filtered attributes, which are not needed to authorize a
trace reference and may dominate payload size. No additional source read is needed.

The typed-report cap does **not** bound raw query bytes or retained Result.Rows.
Up to 10,000 exports times 32 references can exceed the source budget; adding any
metadata can make a previously near-limit numerical read refuse. Document this
prerelease tradeoff, test byte-budget refusal and measure actual fixture response
bytes. Narrowing the fixed query is preferable to silently dropping metric rows or
publishing partial histogram counts. Per-export and report-level omissions remain
explicit and cannot imply known empty. This recommendation is bounded cost
management, not a promise that all previously maximal-byte questions still fit.

## Frozen candidate source review

The implemented report uses uniform source-float64/original-type-unknown labels,
first-32 SQL arrays with original lengths, separate optional metadata decoding and
a conservative 64 KiB envelope reservation. Omitted exports are counted and do not
become empty reference lists. Kernel inputs and mathematical basis/counts remain
independent; whole retained-row basis may change as documented.

Go replay validates the complete result, and selection checks uniqueness across
all raw source rows, including rows absent from the report budget. JS independently
checks all raw stream/end matches before admitting the report's exact sourceRow,
start/end/temporality and ordinal array association. Both preserve original native
window/environment and exact namespace while querying only the reported TraceId;
span hints remain metadata. Native navigation reuses retained ownership and uses
separate reported-association wording rather than interval-contribution language.

Final source is accepted with no material blocker. The reviewed action-lifecycle
edge is fixed: stale/disabled/canAct/visibility/inert checks run before ordinary
click detection, and the same guard handles auxiliary clicks. Fresh Ctrl/Meta/
middle-click remains normal native sharing; retired gestures prevent default and
stop propagation before navigation. Source-backed trace actions also require the
complete expected source question to match the returned measurement question;
read-only metadata inspection does not acquire that action authority.

Root reports full Go, 1,322 UI checks and Hegel 40 passing before the final bounded
handler regression extension. Focused handler checks pass, and the final full UI/
actual runtime rebuild is tracked separately. No runtime acceptance is inferred
from this source verdict.
