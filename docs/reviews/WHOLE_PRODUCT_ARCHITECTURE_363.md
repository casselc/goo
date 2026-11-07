# Whole-product architecture checkpoint — 363

Independent proposal after the accepted 360 breadth, 361 depth and 362
consolidation. This review changes no production code, execution order or budgets
and runs no tests, images or new convention research. Read with checkpoint 359,
the active roadmap and current known issues; historical plans are not parallel
implementation queues.

## Close the delivered sequence

360 now exposes stored current span pages through the existing passive waterfall,
with exact occurrence inspection and source-free return. 361 compares two explicit
whole retained pages, preserves independent frozen operands and opaque locations,
and opens either stored side without telemetry. 362 makes pinned current RPC
metadata capability descriptions agree with service recognition while keeping
shape, generic eligibility, evaluated values and conformance separate.

Retire “stored waterfall absent,” “retained-page comparison absent” and “current
RPC discovery is generic” from active work. Preserve their actual limits: current
pages are not complete traces; label alignment is not invocation or namespace-
qualified service identity; matched source metadata is not producer conformance.
No new capture formats, executors, global reference stores or generic registries
were necessary. These are useful architectural precedents.

## Three concrete next outcomes to consider

These are bounded candidates for planner/user selection, not an automatically
approved feature sequence.

### Breadth: follow a saved native interval into supporting records

A user can reopen metric evidence and inspect the original interval but cannot
naturally follow that saved interval into records while retaining a return to its
capture. `ui/measurements.mjs::recordContextPlans` and `appendRecordContext` already
map counter/histogram export intervals to exact-nanosecond contextual record
questions. Live Measurements supplies `onRecordContext`; captured measurement
rendering in `ui/captures.mjs` does not. The service's retained native rendering
also omits that handoff. This is an entry/return gap, not a need for exemplars or
new metric source columns.

Start with the saved capture owner only. Reuse the pure context planner and staged
fixed-question route, carrying an immutable citation to the original native
receipt, stream, interval, basis and read/custody clocks. Extend or adapt the
existing `native-record-navigation` origin/return contract carefully: its current
active/native-host assumptions are for the live Measurements workspace, so simply
passing that live callback from a capture would be incorrect. Return must restore
the same saved capture without rereading the native source.

Keep only already-supported counter/histogram export intervals. Gauges/stocks
have no invented interval and aggregated histogram time buckets are not export
intervals. Shared time/resource identity is contextual, not proof the records
contributed to the metric, and metric attributes/scope are not record filters.
Stage performs zero telemetry reads; explicit Run acquires records. Test exact
nanosecond bounds, typed policy, namespace/instance, stale callbacks, failed record
reads and saved return/custody. Do not add service-native and every resource origin
simultaneously; establish one honest saved-origin contract first.

### Depth: one researched broker backlog signal, if its identity is adequate

Messaging span defaults and processing-duration metrics already exist; a backlog
observation would answer a different operational question. However,
`config/measurements/queue-lag.yaml` is a demo `lab.queue.depth` binding, not a
portable broker convention. There is no source-grounded universal lag metric to
turn on today.

The next depth proposal should first pin one published receiver/exporter and
close name/unit/type/temporality plus destination/group/partition/resource
identity. Only then implement one per-stream candidate using native discovery,
evaluation and explicit selection if that shape fits existing kernels. Refuse
ambiguous or insufficient identity instead of summing repeated reports across
workers. Native backlog is not saturation, consumer failure, unique deliveries or
producer-consumer causality. If research does not support a safe bounded default,
return a concrete proposal rather than broadening a generic registry or building
a broker API connector to meet an assumed deadline.

### Consolidation: compare evidence before reading stacked provenance

361's real frames expose a concrete product defect: two stacked operand fact
panels consume roughly a desktop viewport before timing bars; narrow frames prove
provenance readability, not chart appearance. 358 improved selected plot positions
but still leaves them below the fold. Another unrelated feature must not obscure
this known visual cost.

Use a compact paired scope/evidence header with Baseline/Selected labels and
adjacent comparable facts, then disclose full questions/receipts. Reuse
`EvidenceSummary.formatFact`, shared display primitives and existing exact
metadata; do not create another acquisition owner or hide partial/unknown state,
policy differences, fixed windows or read-versus-custody clocks. Start with the
actual retained comparison, then reuse a demonstrated common primitive where
other paired owners need it; a blanket presenter migration is not required.

Acceptance is a fresh-scroll desktop/narrow task: users can see both operand
roles, the meaningful timing comparison and any uncertainty without opening a
JSON block. Exact provenance remains available and unchanged. Measure actual
layout/paint and compare source/store ledgers around disclosure, selection and
Back. No automatic read or state mutation follows from compacting the header.
This is a more concrete consolidation than assuming a virtualizer or refactoring
files merely because they are large.

## Data, grammar and library boundaries

The shared grammar already covers ordinary record aggregates/compositions,
several chart kinds, record lists/waterfalls, typed fixed questions, authored
view reuse and independent native measurements. Do not describe it as a count-only
demo or restart a portable whole-UI grammar migration. Specialized profiles and
native data still have deliberate compatibility boundaries; refuse them visibly
rather than coerce them into the ordinary record template arm.

361 exposed an important authority footgun: generic admitted record pages could
be wrapped in receipt-shaped metadata unless stored admission was associated with
the exact returned wrapper. The new `getStoredRecordPage` association binds the
immutable receipt, admitted page and retained/full-count metadata without freezing
caller inputs. Preserve that distinction in future capture-origin features; JSON
labels, passive component handles and copied wrappers do not grant source or
custody authority. Avoid turning this successful narrow guard into a universal
capability framework.

362's small shared native shape helper is the right kind of library reuse: plain
Go name/kind/unit facts, followed by existing binding/temporality/scope validation.
Do not generate UI raw-witness admission and its oracle from that same helper.
Go 1.27.3, `errors.AsType`, immutable data and straightforward typed APIs are already
available; there is no evidence for a new user-facing generic type system or a
runtime rewrite. Useful pure helper extraction can accompany a chosen feature,
not become a maintenance-only roadmap.

## Verification effort and operational cost

Recent real-entry tests found gaps that pure helpers missed: hidden FilterEditor
callbacks mutated before host refusal; workspace navigation omitted an existing
suspend hook; failed stored-side opening left valid held evidence suspended. Keep
actual owner/route schedules alongside independent semantic oracles.

There is also mounting harness cost: milestone-specific journey, actor, ledger and
fixture files repeat transport counters, OTLP setup and evidence-store plumbing.
The reused `tests/support/saved-records-owner-360.mjs` for 361 is a positive example
of bounded fixture reuse. Factor mechanical transport/fixture utilities when the
next selected test needs them, while keeping raw domain membership, timing and
ownership oracles independent. Use existing models for observed actions; 362
correctly reused discovery→inspect instead of inventing a six-step lifecycle.
Do not add another evidence/runner framework or require every metadata description
change to repeat all historical journeys. Required gates should follow the actual
changed authority and failure modes.

## Keep relevant; defer or retire honestly

Host/Kubernetes CPU/memory/network/disk/filesystem, pod states and container
resource/history support are substantial already. Broader USE/workload totals
need exact denominators and alignment, not more generic dashboard labels. Reviewed
task labels/dataset snapshots remain a relevant earlier user request, but need an
explicit vocabulary/review/export contract; ordinary notes are not training labels,
and neither XTDB migration nor a model runtime is a prerequisite. This is not
already supported by the file-backed tags UI: `workbench/meaning.go` accepts only
service/view/document/package/evidence targets, and `CreateMeaning` refuses evidence
without the optional evidence store. A captured-record task would first need a
server-resolved immutable capture-member target (receipt plus absolute occurrence),
explicit task/vocabulary/review revision semantics and a file-store publication
contract. Do not call this ready-to-implement labeling merely because notes exist.

The roughly 32,000-element paired facet tree remains real. Before selecting
paging/windowing, obtain a fresh styled interactive baseline separating source,
build, layout/paint and wall time; earlier duplicate-build CPU is not today's
single-build latency. Keep comparison-body/label byte limits and the blocked
vulnerability assessment explicit. Avoid scheduling repeated facet getters or
unsupported operational assurances as product progress.

Portable WYSIWYG, Databricks lowering, async topology, exemplars, multiplayer and
bitemporal-store migration remain separate contracts. Choose a small useful next
outcome after reconciliation, and review again when that finite work is complete.
