# Discovery and metric-entry understandability review

Reviewed current source at `9119a967`, 2026-11-16. Walkthrough:
[discovery and metric entry](../../walkthroughs/discovery-metrics.md).
This is source review; no new source-backed experiment was run for this chapter.

## Preserve these distinctions

Physical compatibility (`internal/sourcecap`), observed inventory, native profile
eligibility (`measurements` / `ui/instrument-capabilities`) and supplied display
admission solve different problems. Similar words such as supported do not prove
duplicate logic. Do not replace all four with one permissive boolean or registry.

## Proposed simplifications

| Priority | Evidence and problem | Bounded proposal / acceptance |
| --- | --- | --- |
| High | `exploration/dashboard.go` contains `PlanServiceDashboard` and many HTTP/RPC/DB/messaging profile builders in one 912-line file. Learning one profile requires distinguishing general orchestration from profile-specific gates. | Separate named profile planners from orchestration while retaining typed questions and one explicit execution boundary. Existing plan fixtures must keep every admission/refusal and exact question unchanged. File splitting alone is not the outcome. |
| High | `ui/metric-discovery.mjs` builds metadata, eligibility explanations, raw-read/build actions and local paging inside one compact render/draw closure. `ui/measurements.mjs` additionally mixes inventory, context planning and many native displays. | Extract bounded pure instrument row/presentation models and named application actions. Preserve generation/current-owner guards; passive rows must not acquire source authority. Verify raw-only, evaluable, missing-source and refused cases. Coordinate with the presentation/activity reviews rather than adding another abstraction. |
| Medium | Physical source status, instrument verification and display eligibility all reach the UI through different structures and explanatory copy. The separation is justified, but readers must infer which state controls which action. | Document one capability ladder and reuse small checked presentations for those distinct facts. Adopt the planned 465 capability consolidation only after concrete 463/464 use cases show duplication. Do not implement a new universal capability framework speculatively. |
| Medium | `ui/inventory.mjs` hosts service catalog/distributed rendering, read control and service handoff in compact closures. Different counts and continuation lifetimes are easy to conflate during edits. | Name read, render and handoff responsibilities explicitly; preserve count grains, cancellation and retained return. A new developer should trace Discover→open service→Return through three named operations without reading every dashboard profile. |

No production profile, route or source adapter was proven dead. Proposed wider
span scope and raw exponential-histogram inspection remain gated feature work;
unsupported profiles are not legacy code that can simply be deleted.

No measured runtime performance finding is asserted. Inventory bounds, local
paging, cancellation and shared projections already limit work. Benchmark repeated
profile planning/render projection before changing the safety/eligibility gates.
