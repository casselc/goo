# Whole-product design checkpoint — 378

Status: bounded 378 source and actual-frame UX accepted. The next sequence is a
proposal, not adoption. This review inspected original PNGs and did not run tests.
Only this document is owned by the design review.

## Product essence and completed foundations

Goo helps an investigator see current telemetry, isolate a meaningful population,
follow supporting signals, then preserve or reuse the result. The product should
start with observable data and a compact scope, with authoring and technical
explanations disclosed when needed. Logs/spans, service/resource presets, native
measurements, trace inspection, captured evidence, reusable views and record review
already support that loop; they are not missing categories to reinvent.

The recent sequence strengthens three distinctions:

- 376 makes observed unsupported exponential instruments visible in explicit
  metric discovery, without making them executable.
- 377 follows reported histogram exemplar references within an exact original
  scope. It preserves unknown original scalar type, separate export/interval
  identity and paused held return. A reported reference is not causal proof or a
  complete trace.
- 378 brings native failure/recovery into the existing source-refusal vocabulary
  and passive display components. Recovery should be an understandable explicit
  action on the same question, rather than a new lookback or another editor.

The 377 narrow-frame experience is a reusable acceptance lesson: a correct table
can still be unusable if a fixed inspector fills with prefatory explanations.
Show the primary value/record, uncertainty that changes interpretation, and its
action first. Put full source/scope detail afterward. Verify actual data and
interaction rectangles inside the real scroll region, not just the page bounds.
Do not turn that lesson into a new generic inspector framework without consumers.

## 378 candidate review

The candidate reuses `source-refusal.mjs`, `source-refusal-view.mjs` and the existing
Measurements Run owner. The concise reason plus collapsed Technical details is a
cohesive fit. Known source failures preserve typed safe explanations; cancellation
has its own meaning, local validation errors remain local, and no healthy/zero
result is fabricated. Capture must remain disabled until successful acquisition.

Retry captures the failed fixed question and native display declaration. Editing,
previewing, applying another display, cancellation and leaving the owner retire
it. Its callback checks the current failure, generation, plan/display identity,
visibility, inertness and connection before reading. Retry should not refresh the
clock or trigger planning first. Those checks deserve actual mounted tests and a
real refused-then-successful source journey; source inspection alone cannot prove
late-response retirement.

The initial candidate left both enabled **Run measurement** and **Retry
measurement** adjacent after the same failure. That finding was sent to root and
is closed in the candidate: typed failure shows Retry as the one primary recovery
action; retiring it restores Run, and a pending retry exposes disabled Run plus
Cancel. Root reports focused mounted tests passing; actual frame acceptance
remains separate. No production code was changed by this review.

A later source re-read also confirms the metadata-status ownership guard: a late
package-list failure cannot replace an active read, retry or result message.
Package metadata keeps its own catalog status. The native refusal/recovery source
candidate has no further design blocker before the real journey/frame gate.

Keyboard recovery should keep a meaningful location: if Retry had focus, focus
new result content after success or the replacement current Retry after another
failure. It must not steal focus after unrelated navigation or an edited owner.
The actual narrow frame should show the reason and recovery action together,
without expanded technical prose displacing them.

## Actual 378 frame acceptance

Viewed the original `378-native-refusal-390.png` and `378-native-recovered.png`.
The 390px failure shows one Retry measurement action, disabled Capture, the safe
source-unavailable reason and Technical details, followed by the preserved fixed
scope. The inspected detail contains no raw source address, SQL or credentials.
Run is absent during this recovery state, so there is no duplicate primary choice.
The UTC input enclosure and exact native-bound summary remain visibly qualified.

The recovered desktop frame shows actual source records and the numeric window's
own unknown/review-needed state. Recovery does not turn incomplete interval
coverage into healthy data. The source title/read clock, independent streams,
inspection controls and observed timeline remain available. These task-positioned
frames support this recovery flow, not whole-product entry density or universal
accessibility certification.

Root separately reports native journey and positive actor DOM completion. Actor
JPEGs after its inspected phase had a stale compositor image; they are not counted
as positive paint evidence. The main Playwright originals reviewed here are the
current visual evidence, and recorded membership/budgets/focus/retirement remain
separate runtime facts in VALIDATION.md. No further production polish is requested.

## Proposed next finite sequence

These candidates are source grounded, not adoption. Close 378 acceptance first.
The architecture review proposes the same sequence; root should reconcile both.

### 379 — breadth: reachable all-signal service discovery

Task: “My application emitted telemetry. Help me find the service before I need
to know which measurement view can evaluate it.”

`exploration/service_discovery.go` still enumerates only gauge, sum and explicit
histogram metric tables. 376's explicit metric discovery can inspect exponential
metadata, but a service emitting only that family remains absent from the broader
Services entry. Close that reachability gap using the existing service inventory,
reported signal/count presentation and fixed-scope handoff. No chart or native
executor is required. Counts remain exported points, not histogram observations,
traffic volume or health.

Preserve typed environment, exact namespace and window. Optional source absence
must be visible rather than becoming either a failure of every family or a false
complete empty inventory. A discovered unsupported source should remain inspectable
without suggesting Run is available. The real journey must begin in Services,
find the actual exponential-only service, and reach its observed unsupported
instrument without a synthetic navigation shortcut.

### 380 — depth: follow a reported exemplar from immutable Saved evidence

Task: “I saved this investigation. Let me follow the reference I was looking at,
while retaining the historical source that justified it.”

377 deliberately enables live Measurements actions only when the complete source
question matches its request. `ui/saved-native-record-context.mjs` and
`ui/native-held-record-context.mjs` own immutable interval-context admission;
merely passing a live source adapter or enabling their existing callback would
create a false working link. Extend that typed owner for the reported-reference
relation, starting with one immutable Saved measurement, rather than bypassing it.

Opening the saved report and staging the trace should not read telemetry. Explicit
Run is a fresh scoped read; Back retains the saved report, original read clock and
source occurrence. Preserve the exact namespace and nanosecond window, source-row/
ordinal association and unambiguous export identity. A newer source read must not
masquerade as the saved witness. Service-owner adoption can follow only if this
same admission fits its ownership contract; do not promise every surface at once.

### 381 — consolidate: common bounded JSONEachRow decoding

Task: “The same bad source response should have the same meaning in records and
native measurements, without changing the question or hiding partial data.”

`exploration.Source.read` and `measurements.Source.read` now both buffer a bounded
body, iterate `bytes.Lines`, decode object rows, apply row caps and classify source
failures. Their limits differ, and their blank-line behavior differs. A small
shared decoder above the deliberately mechanics-only `internal/sourcehttp` can
consolidate that repeated protocol work with explicit caller limits and one
reviewed row policy. Keep family query compilation, evidence, numeric evaluation,
context deadlines and lifecycle custody separate.

The user-facing acceptance is consistent concise refusal/recovery across ordinary
records and Measurements, with unchanged valid results and no partial population
accepted. This is not a global recovery manager, a transport framework rewrite or
a reason to mix persistence failures into malformed telemetry. Use current Go
stdlib facilities where they reduce duplication and retain simple caller APIs.

## Reconcile older proposals

Do not schedule native display Browse/Inspect/Apply again: 373 delivered that.
Do not propose a record-review UI based on the old readiness-gate banner:
`CAPTURE_RECORD_REVIEW_GATE.md` explicitly points to the delivered
`CAPTURE_RECORD_REVIEW.md` feature and actual `ui/capture-record-review.mjs`.
Service/resource defaults and ordinary correlation/tails are implemented; refine
their actual reachability and semantics instead of adding a vendor feature list.

Keep exponential execution deferred until its source and mathematical gates
close. Do not infer absent producer fields, original attribute/scalar types,
complete traces, metric/span causality or a fleet percentile from unavailable
facts. Messaging broker health remains a source-contract research candidate, not
an automatic dashboard promise. XTDB, a new query engine and another authoring
syntax are not prerequisites for the concrete entry/follow-through gaps above.

This sequence balances a broader entry, deeper evidence reuse and shared protocol
cleanup. Reassess after those outcomes or an earlier source-model blocker; do not
allow repeated local guard or screenshot adjustments to become the roadmap.

## Root adoption

After 378 runtime, API and recorded-model acceptance, root adopted the proposed
379/380/381 sequence in the [November 8 roadmap](../ROADMAP_2026_11_08.md). Those
three tasks remain planned; this review does not claim their implementation.
