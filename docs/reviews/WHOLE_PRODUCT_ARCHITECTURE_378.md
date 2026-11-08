# Whole-product architecture checkpoint — 378

Status: proposal for root/planner adoption after 378 acceptance. The current 378
source candidate is reviewed below; this document does not establish runtime,
model or screenshot success and does not adopt another execution queue.

## Product and source boundary

Goo's loop remains live telemetry → scope a population → follow reported evidence
→ preserve or reuse the question/result. Ordinary OTel records and native streams
share that user journey while retaining different mathematical and acquisition
contracts. A passive display, a fixed query, a held result and an immutable receipt
are distinct objects. A reported trace reference is not bucket contribution or
complete trace capture. Keep these distinctions visible without making every
interaction a new planner, owner or recovery framework.

The 376–378 sequence closes useful connections: observed unsupported instruments
are discoverable; selected histogram exports expose bounded reported references;
native Measurements uses the existing source-refusal/recovery contract. The next
sequence should connect these capabilities across actual product entry points,
then consolidate the repeated source boundary rather than add another native
family or reopen deferred mathematics.

## 378 candidate source review

`measurements/source.go` leaves question detachment and Compile validation before
acquisition. HTTP non-200 responses become unavailable before their bodies are
interpreted. Transport/body failures retain causes under the shared typed error;
local payload/row limits become budget refusals. Object-only JSONEachRow admission
uses `bytes.Lines` and detached `bytes.Clone` rows. Numeric Evaluate errors after
acquisition use `sourcefailure.Protocol`. Lifecycle Observe/custody persistence
remains outside that blanket protocol classification, so a filesystem failure is
not falsely described as malformed telemetry. No source-error prose is parsed.

`ui/measurements.mjs` reuses SourceReadError and the shared refusal view. Retry
captures the failed fixed question and display declaration and calls the same Run
path. The current retry node, generation, original plan object, complete question,
display intent and visible/non-inert owner must still match. Edits, preview,
display Apply, cancellation and suspension retire the retry. Invalid configuration
is separate from retryable source failure; cancellation does not acquire a retry.
No new rolling lookback or automatic source read is introduced.

No material source blocker was found in this reviewed candidate. Keep the actual
refusal/retry/cancel/late-response and focus assertions as separate acceptance
requirements. Source reading is not evidence that those runtime gates passed.

## Proposed 379 — breadth: make service discovery reflect observed metric families

The concrete entry gap is `exploration/service_discovery.go`:
`CompileServiceDiscovery` aggregates only gauge, sum and explicit histogram metric
tables. Meanwhile `measurements/discovery.go` can explicitly discover the retained
exponential family. An exponential-only service can therefore be found when its
name is already known but remains absent from broader Services discovery.

Extend that broader discovery through a checked optional-source contract, keeping
existing log/span service entries and reported-edge derivation. Show observed
metric presence and a natural continuation to explicit metric discovery. Counts
remain exported data-point records; they are not histogram observations, traffic,
health or an executable measurement. An unsupported-only service must not acquire
fabricated numeric widgets or dependency edges.

Close the acquisition/compiler boundary first: an absent optional table must be
explicit and must not break supported families; probe/query failures must not turn
into absence. Pin included-family provenance to the actual result/plan so a later
source-capability change does not silently reinterpret held evidence. Preserve the
original environment policy, namespace presence and fixed window and the existing
global signal/group bounds. Reuse 376's narrow availability vocabulary rather than
create a source registry.

Acceptance: actual retained exponential-only service appears in its exact scope,
its explicit inventory agrees with independent raw counts, and known unsupported
status is clear. Ordinary log-only/span-only/supported-metric services still work;
optional absence and source failures remain distinct. No exponential evaluation,
summary-metric support or whole-platform membership claim follows from this cut.

## Proposed 380 — depth: continue reported exemplars from an immutable saved owner

377 deliberately exposes actions only in source-owned live Measurements. The
existing Saved renderer (`ui/saved-native-record-context.mjs`) verifies an immutable
receipt and supplies current-owner leases, but `ui/native-held-record-context.mjs`
only admits an interval relation through NativeRecordContext. Service retained
histogram widgets use that same interval admission. This is an actual relation
contract gap, not an undiscovered feature toggle.

Start with one Saved native capture. Admit a discriminated reported-exemplar
relation against that exact receipt/result, all raw export rows, source-row
occurrence, stream/end and sample ordinal. Reuse the checked Go/browser planner
and original fixed namespace/time/environment scope. Keep the old capture's read
clock separate from the fresh trace acquisition; absent trace evidence in the new
read is not proof that the captured reference was false.

Do not set `saved:false`, fake a live SourceAdapter, blindly expose callbacks or
weaken the existing interval validator. Keep numeric snapshot authority and
reported-reference authority separate. Source-free Stage, explicit Run, paused or
renewed actions under an explicit owner policy, stale-callback refusal and the
same immutable parent on Back need a closed contract before enabling the UI.
Service-widget continuation is a follow-on only if the same typed admission and
owner semantics fit without another manager; do not bundle every surface at once.

Acceptance: a captured export with actual retained exemplar metadata stages the
same scoped query as its checked live counterpart; trace-only references work,
invalid trace IDs and ambiguous/omitted exports do not act, and a capture without
metadata stays unavailable. Successful capture reopen/Stage/Back reads no source;
only explicit trace Run acquires new spans. Capture bytes remain unchanged.

## Proposed 381 — consolidate: one bounded JSONEachRow decoder, explicit caller limits

`exploration/source.go` and `measurements/source.go` now repeat successful-body
JSONEachRow splitting/object validation, row-copy and row-budget logic around the
same source HTTP mechanics. Their family limits and blank-line policies must not
drift accidentally. This duplication has a concrete cross-product effect: the
same malformed source can otherwise produce different refusal meaning in Activity,
service widgets and Measurements.

Extract a small internal decoder above `internal/sourcehttp`, with an explicit
caller row bound and one documented object/blank-line/no-partial-result contract.
Keep sourcehttp's existing mechanics-only Post/ReadBounded boundary; do not turn
it into a query executor or attach a global retry/owner state machine. Compilation,
URL/parameter policy, source status, numeric/record evaluation, evidence clocks and
lifecycle custody remain with their current families. No advanced public generic
API or format compatibility layer is needed.

Require rejecting tests for malformed/non-object/torn/over-limit inputs, exact
raw numeric tokens and detached rows, plus actual record/native transport parity
and unchanged query parameters/read budgets. Keep independent domain oracles;
a shared parser must not become the common implementation used to prove its own
numeric or population truth. Measure meaningful cost if this changes buffering,
rather than claim performance from a refactor alone.

## Remove obsolete work; keep independent gates independent

- 370/372 typed follow-through and 373 bounded native-view compatibility/discovery
  are delivered. Do not schedule their old scalar-only or exact-ID-only gaps again.
- Log patterns, trace/log tails, retained comparisons, default service/resource
  dashboards, authored record views and captured-record review UI already exist.
  In particular, `ui/capture-record-review.mjs` is implemented; the historical
  review-gate banner is not evidence that the whole review UI remains absent.
- 374's general exponential execution stays deferred for source fidelity and
  representation reasons. Discovering a service or following a reported ID does
  not recover zero thresholds, original scalar type or symbolic bucket geometry.
- Broker-health contracts, broad WYSIWYG, Databricks lowering, durable alert delivery
  and authenticated team/consensus semantics remain separate product decisions.
  They are not implied by adding samples, labels or a bounded source catalog.

Go 1.27.3 is already current here. `errors.AsType`, `bytes.Lines`, and ordinary
maps/slices cloning are useful when they simplify these boundaries. A runtime
upgrade or clever public type machinery is not an independent roadmap item.
Review again after this finite breadth/depth/consolidation sequence, or earlier
if a required source/relation contract cannot close.

## Root adoption

After 378 runtime, API and recorded-model acceptance, root adopted the proposed
379/380/381 sequence in the [November 8 roadmap](../ROADMAP_2026_11_08.md). Those
three tasks remain planned; this review does not claim their implementation.
