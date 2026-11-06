# Whole-product architecture checkpoint — 359

Proposal after the completed 355 sequence: 356 authored-template service entry,
357 pinned current RPC, and 358 fixed-question read/Edit consolidation. All three
have source and execution acceptance. This review changes no production code,
source budget or active roadmap and runs no tests or network research.

## What is now coherent

Ordinary OTel data can enter through Activity, discovery or a fixed question;
users can isolate it, inspect context, compare/explain populations and preserve or
reuse definitions. 356 makes the existing grammar useful from service widgets
without inheriting a template specimen's scope. 357 adds current RPC with optional
methods and independent native seconds rather than aliases. 358 puts applied
fixed evidence ahead of optional authoring while preserving actual raw drafts,
complete question profiles, fixed bounds and existing owners. Retire those items
from the active queue; do not start another template/RPC/Edit cleanup sequence.

The architecture still benefits from one Go server, typed source lowering,
independent read receipts, immutable retained artifacts and passive display
components. Precision/completeness rules are useful product semantics: full
counts do not come from row pages, incomplete relationships are not health, and
native streams are not span populations. Keep these through future features.

## Proposed finite sequence

### Breadth: inspect retained spans as a source-free waterfall

The ordinary incident question is “What did the evidence I saved show?”
`ui/saved-records.mjs::showSaved` renders retained rows, but TraceJourney examples
still call `records.openTrace`; `ui/record-results.mjs::traceRows` performs a fresh
trace read and then linked logs. Known issues explicitly say a stored-waterfall UI
has not been added. Existing passive waterfall/timing projection can already draw
supplied rows. This is a missing retained-evidence entry, not a new trace engine.

Add a clearly named retained-page trace view from saved span evidence, preserving
the original receipt/question and selected trace identity. Use only checked stored
rows from the declared page/bounded retained selection. Count these displayed
observations explicitly; a generic capture's full population count is not the
trace's count. Missing parents, omitted pages, capture limits and out-of-window
spans remain unknown. Never infer a complete trace, critical path or exclusive
CPU time. Existing fresh Expand remains a separately labelled explicit action.

Acceptance: a genuine capture with multiple traces, missing parent, duplicate IDs
and limited rows; offline/source-refused reopen must still show retained timing
without telemetry reads. Test scalar and typed policy, exact record identities,
zero-duration/large timestamps, partial coverage, keyboard inspection and retained
Back. Receipt-store paging is metadata/evidence access, not telemetry; disclose
which supplied rows the view uses instead of silently aggregating all pages.
Reuse the existing passive component/declared data semantics, with no capture
format migration unless a demonstrated missing field requires it.

### Depth: compare two retained trace pages

Once a saved trace subset can be inspected, the natural next question is “How did
this retained example differ from the other one?” The pure
`exploration/trace_page_comparison.go::CompareTracePages`, matching JS model and
`ui/trace-comparison-view.mjs` already compare supplied observations. The current
`record-results.mjs` reference controls hold ephemeral freshly acquired pages;
they are not a durable stored-page selection entry.

Add explicit baseline/selection from two checked retained trace-page selections,
reusing the comparison model and common elapsed-time scale. Keep each receipt,
original query, page extent, selected trace and truncation separately visible.
Labels align observations, not unique invocations; missing groups are not zero
work. Do not pool counts, compute a fleet percentile, infer causal latency
attribution or acquire fresh trace/log context while opening a stored comparison.

Acceptance: different fixed windows/services/policies, duplicate labels and IDs,
missing parents, unequal page extents, empty/invalid selections and partial pages.
Prove selected capture/page identity and immutability, zero telemetry reads on
reopen/selection/Back, explicit operand roles and readable paired timing/attributes
on desktop/narrow. Reuse existing pure Go/JS parity fixtures plus independent raw
observations. A new comparison engine or capture format is not presumed.

### Consolidate: convention capabilities agree across discovery and defaults

There is a concrete cross-surface gap after 357: the service planner recognizes
`rpc.server.call.duration` in `exploration/dashboard.go`, while the curated
`measurements/instrument_capabilities.go` profile paths still concentrate on
HTTP/DB/messaging, leaving that instrument generic in metric discovery. A user
should not receive incompatible descriptions of the same admitted instrument.
Another convention binding would otherwise extend that drift.

Use the existing `InstrumentCapabilityPlan` boundary and small internal checked
facts for recognized name/type/unit/pinned profile, sharing them where that avoids
real duplication. Keep Go planning/compilation authoritative and retain independent
UI raw-row positive-witness checks. Do not generate both admission and test oracle
from the same descriptor, publish a new universal convention registry, or remove
custom metric support. Metadata recognition remains a candidate, never proof of
producer conformance or observed values. Acceptance should compare discovery,
service/native option descriptions and actual emitted questions across current/
legacy/mixed/wrong-unit cases without changing old stream identity or budgets.

## Other material opportunities and prerequisites

- **Broker backlog:** messaging spans do not establish lag. The existing
  `config/measurements/queue-lag.yaml` is a demo `lab.queue.depth` binding, not a
  portable broker default. Before a later depth cut, pin one published receiver/
  exporter and close metric/unit/type/temporality and destination/group/partition
  identity. Reuse per-stream native evaluation; do not sum repeated worker reports
  or infer broker health. This prerequisite is why retained comparison is the
  better ready-to-implement depth cut now.
- **Large-facet presentation:** 343/346 removed duplicate work, not the roughly
  32,000-element paired tree. A fresh styled interactive baseline must separate
  source, build, layout/paint and wall time before selecting paging/windowing.
  Search still spans the full loaded model; counts, uncertainty, selection and
  receipts cannot be truncated. This is a conditional performance investigation,
  not an assumed virtualizer or another publication getter change.
- **Reviewed training judgments:** tags/notes and retained evidence exist, but
  `TRAINING_ANNOTATIONS_DESIGN.md` explicitly says reviewed task labels, dataset
  snapshots and export are unimplemented. This remains a relevant user request.
  A later bounded task-vocabulary/review/export contract can reuse immutable
  evidence and append-only metadata; neither an XTDB migration nor a model runtime
  is a prerequisite. Do not silently turn notes or population predicates into
  ground-truth record labels.

## Reconcile architecture and operational backlog

Host/Kubernetes support is substantial: host resources, CPU/memory, network/disk/
filesystem families, pod memory/phase/readiness and container resource context
already have source modules. Do not advertise “add basic K8s/USE dashboards” as if
none exist. Broad host utilization, workload aggregation and broker health still
need explicit denominators/identity/time alignment; one signal is preferable to
an ungrounded dashboard catalog.

Go 1.27.3, modern typed errors and bounded clients are already used. There is no
source evidence for a new user-facing generic type system or executor. The large
`population-workspace.mjs` and mixed presentation/acquisition service adapter are
real maintenance pressure, but file size alone does not choose a refactor. Extract
pure question/profile helpers when the selected cut changes them; retain simple
APIs and independently checked authority. No compatibility floor prevents a
cleaner shape, but prerelease status does not justify deleting useful distinct
legacy data profiles.

358's late FilterEditor finding is a useful lesson: host callbacks cannot veto
private mutation that already happened. Local widgets now admit user mutation
before changing tree/render/focus; programmatic setters remain separate. Carry
that boundary into new interactive controls without launching a global lease or
rollback framework. Keep actual-entry actor/model checks alongside pure
properties: direct suspend unit tests alone do not establish real navigation
wiring.

Preserve documented byte/cardinality limits and the blocked vulnerability-check
status. Do not conflate UI publication with the unresolved cross-layer comparison
body/label byte budget. Retire old unsupported typed-edge/package assumptions and
keep historical failed broad harnesses distinct from accepted dedicated journeys.
Portable whole-UI grammar, WYSIWYG, Databricks lowering, async topology, multiplayer
and a native bitemporal store remain separate proposals with their own concrete
source/user contracts. After the next finite sequence, review the whole product
again rather than extending a maintenance-only queue.
