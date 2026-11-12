# Review reconciliation — 411

Read-only architecture/Go and whole-product design review after 409/410, while 411 quantitative composition consolidation awaited final acceptance. This review ran no tests, benchmarks or browser campaigns; actual acceptance belongs in [validation](../../VALIDATION.md). The accepted next finite queue is [412 → 413 → 414](../ROADMAP_2026_11_12_CHECKPOINT_411.md).

## Adopt

- **Contextual task discovery:** the captured-record review interface requires an exact task ID, while its metadata API provides creation and exact-ID loading without a catalogue. Add a bounded chooser inside the existing captured occurrence, using validated immutable task identities and explicit vocabulary selection. This connects an already implemented review/export capability to a natural user entry.
- **Same-capture occurrence navigation:** deepen that workflow with Previous/Next under the existing Saved record owner. Absolute retained occurrence identity must survive duplicate reported IDs and visual ordering. Explicit draft/pending disposition and task continuity are design prerequisites, not incidental button behavior.
- **Compact shared record operations:** actual product frames show stacked secondary commands. Consolidate hierarchy across Activity, detail, capture and review with shared accessible controls and concise labels. This is a bounded interaction cleanup, not another whole-app navigation rewrite.

Activity already offers scoped live telemetry, contextual investigation, reusable views and held return. Services gained useful count comparisons and literal incoming response context in 409/410. The review found no stronger demonstrated gap that warrants another HTTP lens or a new top-level labeling workspace. Keep review contextual to evidence operators have selected.

## Preserve architectural boundaries

The one Go runtime, typed measurement/exploration packages, physical source adapters, source-neutral supplied display contracts and application acquisition/navigation owners remain appropriate. Passive components consume data and emit local intents; they do not acquire telemetry or publish evidence.

411 addresses concrete composition and handoff issues: embedded histogram framing, accessible native overflow and the distinction between control invocation and checked asynchronous completion. An opener may be transiently disabled during its own operation without invalidating a still-current completion. Visibility, disabled ancestry and owner/source epochs remain authoritative. Redrawing held evidence cannot renew original-read authority. These distinctions do not call for a generic view/source manager.

Inspected production code already uses concrete domain types, bounded readers, exact arithmetic and current Go facilities such as `errors.AsType` and `slices.Clone`; the module declares Go 1.27. No evidence in this review warrants a public generic builder, experimental JSON migration, additional query engine, new storage engine or broad runtime rewrite. Use newer standard-library/runtime functionality where a change demonstrates a benefit. Shallow clones do not replace nested ownership validation.

## Defer with an explicit gate

Static inspection still counts **92,281 lines** across top-level `verification/*_test.go`. This is a maintenance/compile-coupling concern, not a measured production bottleneck. The isolated 409/410 semantic packages establish a useful direction. Avoid new dependencies on private numbered milestone helpers.

A bounded extraction of independent record-membership/exact-number/histogram support remains worthwhile, but should displace 414 only if helper coupling materially blocks that work. Otherwise schedule it separately, preserving independent oracles and negative controls, with measured cold/warm compilation and memory before claiming an improvement. Do not migrate the entire historical suite or build a general campaign interpreter inside product polish.

Fixed-summary captured-rate comparisons remain useful for unequal windows. They require exact count/exposure support per operand and explicit independent clocks; they must not imply unsampled application throughput. Equal-width preceding-window count comparisons already answer a common service question, so rates remain deferred for this finite cadence.

Fleet percentiles, exponential histogram execution, inferred health, authenticated reviewer consensus, executable contributed kinds and external warehouse deployment retain their existing separate gates. Earlier review proposals remain research inputs rather than additional automatic backlog items.
