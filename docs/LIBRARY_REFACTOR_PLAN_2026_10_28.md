# Cross-stack library refactor proposal — 2026-10-28

**Review and proposal only.** Specialists inspected the full project at `af107b7`,
including all six Go modules, frontend application/display libraries, Python
forwarding, instrumentation/weaving, authoring and verification. This is broad
structural coverage, not a line-by-line correctness audit. No production code,
public API, dependency, toolchain, runtime settings or deployment changed. No new
regression suites, browser campaigns or performance benchmarks were run for this
review. Source references and past acceptance are evidence of current structure,
not a new passing execution result.

The recommendation is to improve the internals while keeping ordinary OTel,
concrete Go questions/results, validated YAML, and supplied-data HTML components
the simple public paths. Preserve the existing semantic layers; share mechanics
where they are demonstrably the same. An advanced facility is useful when it
removes a real ownership or performance problem, not when it merely reduces lines
or exposes more type parameters.

## Specialist reports and usability decisions

- [Go libraries and runtime](reviews/GO_LIBRARY_REVIEW_2026_10_28.md): module/package
  coverage, eight proposals, internal generics and recent stdlib/runtime assessment.
- [Frontend libraries](reviews/FRONTEND_LIBRARY_REVIEW_2026_10_28.md): whole-app
  coverage, eight proposals, browser/platform gates and extraction/conformance plan.
- [API design/usability](reviews/LIBRARY_API_USABILITY_REVIEW_2026_10_28.md): public
  API dispositions, integration examples and confirmation of both platform matrices.

These are inputs to **one sequence**, not three independent framework projects.
Earlier reviews were reconciled against current code: generic startup, narrow HTTP
and publication primitives, request owners, facet composition, record/evidence
adapters, definition rows and host-control consolidation already exist. Do not
rebuild them because an older review described them as future work. Product C2 and
other existing roadmap work remain separate; this proposal does not turn every
library idea into a prerequisite for the next product feature.

## Recommended sequence

| Cut | Concrete outcome and pilot | Consumer impact and acceptance |
| --- | --- | --- |
| **R0 — Contract and compatibility ledger** | Inventory route effects/admission/deadline owners; map normative Go/browser identity, bounds and error contracts to existing fixtures. Record root Go 1.24 versus nested Go 1.27 requirements and the current Chromium-only browser evidence. | No API change. Preserve canonical IDs, nil/missing/empty distinctions, exact strings, unknown states, source-free staging/reopen and per-family read budgets. Extend existing fixtures rather than create a parallel schema system. |
| **R1 — Request lifetime and orderly shutdown** | First add a held-headers/delayed-body cancellation regression across browser, Python and Go, with downstream cancellation/admission release observation. Then pilot an internal JSON helper in two read-only owners if the contract warrants it. Separately add bounded server/example shutdown and exporter flushing at executable boundaries. | Keep existing raw Response support and SSE separate. A small `{ok, status, headers, data}` result can preserve caller distinctions; do not return a consumed Response or attach source cancellation to every pure/save/publication request. No claim of an existing leaked server query is established by the review. |
| **R2 — Route policy and source boundaries** | Introduce a small internal method/pattern/effect/admission/deadline inventory that initially reproduces current membership. Move one cohesive handler family out of `cmd/goo` only after its dependencies are explicit. Corpus-test equivalent ClickHouse escaping/settings fragments before sharing them. | Concrete DTOs and wire behavior stay unchanged. Pure POSTs, executing GETs, SSE, optional retained-store commits and capture publication have different policies. Changing admission membership or error/budget ordering is an explicit behavior change, not mechanical cleanup. |
| **R3 — Frontend ownership and input leaves** | Relocate proven `RequestOwner`/`PendingScope` mechanics from feature modules with compatibility aliases for identified callers. Pilot private property-upgrade/bounded-JSON helpers in two compatible snapshot elements. | Keep visibility, selected identity, open-panel and active-route publication guards in app adapters. Preserve snapshot versus child-preserving element contracts, pre-definition properties, SSR attributes, focus and zero lifecycle reads. Do not introduce a universal owner or public custom-element superclass. |
| **R4 — Checked revisions and explicit copies** | Add checked domain-specific hashing helpers internally, then named clones for stable owned structures where useful. Freeze current IDs and mutation-isolation behavior first. | Existing `Hash` and stored records remain compatible. Generic `Hash[T]` does not prove JSON encodability; `maps.Clone` is shallow and does not preserve deliberate nil-to-empty normalization by itself. A public checked revision helper is additive only if an actual consumer needs it. |
| **R5 — Prepared numeric keys, measured** | Benchmark parsing/canonical serialization in sort/group preparation, then cache the existing parsed times and canonical ordering bytes in private prepared records. | No public generic evaluator. Preserve duplicate/conflict/freshness/epoch behavior, tie ordering and replay bytes. Compare 1/100/10,000 points, multiple streams, shuffled inputs and adversarial ties with allocations. There is no measured speedup yet. |
| **R6 — Explicit instrumentation runtime, additive** | A proposed `NewRuntime(Config) (*Runtime, error)` can own instrument handles and bounded initialization state for direct integrations/tests. Keep package helpers and annotations working. | Design/usability accepts with revisions below. Provider setup/shutdown belongs to callers; errors must surface. The first scope is manual consumers/tests because existing woven aspects still call package globals. No public generic machinery or mandatory panic/recover boilerplate. |
| **R7 — Delivery, styles and remaining adoption** | Validate existing asset/dependency lists and display-only gallery isolation. Consolidate touched owner CSS and remaining label/evidence duplication using already shipped adapters. Measure startup and supported large-view interactions before optional loading/yield/containment changes. | No mandatory bundler, framework, ESM-only entrypoint, new global store or broad CSS rewrite. Preserve standalone/nested component contracts and native keyboard/touch behavior; update affected real frames only when implementing a change. |

R1's cancellation contract and R2's route inventory are the most useful first
reliability cuts. R0 should be a small ledger/conformance addition, not an open-ended
architecture phase. After that, R2 Go boundary work and R3 frontend mechanics can
proceed independently behind the same fixtures. R4/R5 are separable internal pilots;
R6 should be driven by an isolation/error-reporting consumer. R7 is opportunistic
maintenance and measurement, not a platform migration.

## Public API decision: simple defaults, explicit ownership where useful

The only concrete new public API candidate in this review is illustrative:

```go
// PROPOSED; this constructor is not implemented by the review.
runtime, err := instrumentation.NewRuntime(instrumentation.Config{
    MeterProvider:  meterProvider,
    TracerProvider: tracerProvider,
})
if err != nil {
    return err
}
```

Design/usability **accepts this direction with revisions**:

- Runtime owns its instrument/initialization bookkeeping. Applications own SDK
  installation, exporters and shutdown. Nil/default-provider behavior is explicit.
- Constructor failures are returned. Two instances with the same operation names
  must remain isolated; finalizers stay concurrent-safe and exactly once.
- Existing annotations and package functions remain the ordinary path. Current
  weaving calls those globals, so constructing Runtime alone does not isolate a
  woven app. A weaving binding would be a separate reviewed change.
- The low-level `(error, panicked)` finish hook must not force ordinary users to
  implement recover/repanic boilerplate. Examples show the appropriate common path.

No public generic query DSL, `Result[T]` hierarchy, generic numeric policy plugin,
universal store repository or subscribing dashboard component is recommended.
Concrete Query/Question/Result types and existing component property/event APIs
remain the public contracts. Advanced internals may implement them without adding
conceptual work for callers.

## Modern features: use the benefit, respect the floor

The Go specialist verified feature versions against the installed Go 1.27.2
specification/API history/source. The root module declares Go 1.24; instrumentation,
verification, verification/weaving and both examples declare Go 1.27. A newer verification compiler
does not establish root minimum-toolchain acceptance.

Good immediate internal candidates include `testing.B.Loop` for the proposed
benchmarks and test-owned contexts, both within the root floor. Existing private
generic helpers and ordinary typed prepared records are appropriate. Standard
`maps`/`slices` helpers can simplify touched code while preserving shallow/deep-copy,
nil and stable-order policies. Generic aliases are available at the root floor but
need a real compatibility use case; they are not a reason to expose type machinery.

`testing/synctest`, `WaitGroup.Go` and newer generic error inspection can be useful
in already newer nested modules, subject to their documented semantics. They are
not reasons to raise the root minimum or replace real HTTP/browser tests. Runtime
container defaults depend on the built executable's compatibility settings: merely
building a Go 1.24 main with a newer compiler does not establish the same
container-aware GOMAXPROCS defaults as a Go 1.27 main. Inspect actual main/build and
runtime settings before tuning or claiming a benefit. No GC/JSON/runtime switch,
pooling, unsafe allocation strategy or toolchain upgrade is proposed.

On the frontend, the existing custom elements, BigInt, structured clones, native
HTML, abort signals, streams and scoped layout are already strong foundations.
Conditional internal improvements include signal-owned listener cleanup, compatible
abort composition and measured rendering yields/containment or optional loading.
Document browser/version support before requiring additional platform features;
current Chromium campaigns do not prove Firefox/Safari support. Keep exact
integer/decimal/nanosecond strings and existing SSR/property/events. Temporal,
Shadow DOM, a new module loader or a new gesture is not a prerequisite for cleanup.
The usability reviewer confirmed both modern-feature matrices on these terms.

## Verification and stopping rules

Each implementation cut needs a concrete consumer benefit, the narrow relevant
contract checks and removal of its superseded path. Use existing wire/golden
fixtures, rejecting controls, focused Node/Go tests, Hegel owner schedules,
representative real OTel journeys and Bombadil/model traces as appropriate. Do not
repeat every live campaign for a leaf helper or invent a new finite model per
button. New structural work must keep docs/examples/skills, applicable diagrams and
actual UI frames synchronized; this review itself requires no screenshot rerun.

Stop or split a proposed abstraction when its configuration reproduces every
family's semantic branches, changes historical IDs, makes a display fetch data,
loses ownership/unknown distinctions, or requires users to learn a generic algebra
for an existing concrete task. Performance proposals need equal-output benchmarks;
reliability proposals need a failing boundary case or explicit conformance gap.
Ordinary app adoption remains the test of API ergonomics.
