# CEL whole-product Go review

Review basis: commit `870e98a0`; source inspection, 2026-11-14. This is an independent design review, not a claim that the proposed integration exists. Read repository guidance, the current heads of VALIDATION/KNOWN_ISSUES/IMPLEMENTATION_PLAN, the CEL pilot and acceptance report, and the domain/transport code cited below. No product code changed.

The delivered pilot has the right architectural direction: checked CEL is an authoring frontend to ordinary record predicates. Keep `Compile(signal, expression)`, detached `Predicate`, whole-parent `Apply`, and the pure versioned OTel library. Do not make CEL a parallel query execution protocol. The next work should consolidate domain operations and then connect one draft owner, rather than introduce another executor or generic language framework.

## Findings in priority order

### 1. High integration priority: predicate admission and refinement belong to the domain

`querycel.Compile` constructs a synthetic `exploration.Query` with a fabricated environment/window/page to call `Validate`, because `exploration.validatePredicate` is private. `Program.Apply` owns JSON deep-copying and parent-AND composition. Existing `exploration.StageGroupSide` separately constructs an AND; follow-up/tail code also uses JSON copying. This is working pilot code, but the next frontend would otherwise repeat the same domain operation and inherit accidental coupling to unrelated query requirements.

Add a small ordinary-domain operation for predicate validation and a detached, validated question refinement, then delegate CEL Apply to it. Preserve the whole OR parent and enforce combined budgets after composition. Add domain-owned typed clone functions for the Query/Predicate graph instead of round-tripping through `Query.UnmarshalJSON` (which performs strict wire admission, reflection and additional JSON normalization). `copyPredicate` currently discards marshal/unmarshal errors; a typed clone also removes that silent-error convention. Use `slices.Clone` for slices and explicit recursion/pointer copies, with all nested numeric/environment/histogram/edge fields covered. Do not replace a small finite model with a reflection-based universal clone framework.

Acceptance: existing CEL mutation isolation/parity tests pass; domain tests cover all pointer/slice-bearing query members and combined depth/node refusal; a non-CEL refinement consumes the same operation. No implicit pagination reset or scope rewrite is added.

### 2. High integration priority: authoring needs typed, positioned diagnostics

`querycel.Compile`, `lowerer.predicate`, and `Program.Apply` currently return `fmt.Errorf` prose. Parser/type-check issues are flattened into strings; lowering has checked expression IDs but emits no source location. `queryhttp.PlanHandler` returns only `{error: ...}`. The browser cannot reliably distinguish an unsupported field, unavailable operation, literal mismatch, or combined parent budget, nor focus the affected text without parsing prose. `exploration.NumericDiagnostic` demonstrates an existing code/path/message convention worth aligning with, rather than inventing a separate error UX.

Introduce one concrete authoring error/diagnostic shape with stable code, phase, message, optional expression span and optional domain path. Preserve `errors.As` inspection and normal `error` returns; avoid custom result monads or caller-configured handlers. Translate CEL's source information while it is available, including lowering failures. Keep actual expression/record contents out of routine telemetry. Share diagnostics between CLI receipt handling and the eventual HTTP planner.

Acceptance: syntax/type/lowering/parent-budget refusals have stable codes; Unicode positions highlight correctly; no successful partial question appears on failure. Browser and CLI do not inspect error prose. No broad conversion of every historical Go error is required.

### 3. Medium: one logical field catalog, separate physical adapter mapping

`exploration.Field.Type` embeds a global type map plus signal exclusions; `querycel.declarations` independently repeats names, types and signal membership; `compiler.field` holds physical ClickHouse mappings. Adding another field currently requires synchronized edits across these structures. `Match` further lists specific uint64 keys for equality. This becomes a product vocabulary drift problem when discovery, visual editing and CEL completions describe the same signal.

Use a closed logical descriptor set for stable identity, scalar type, signal support and presence semantics. Let CEL maintain an explicit authoring exposure policy over these descriptors: derived fields are not automatically CEL variables. Let source adapters retain physical expressions/readiness, and expose a serializable subset for existing UI discovery/completion. Do not put SQL, CEL types or browser widgets into the domain descriptors. Drive reference scalar comparison from admitted field type rather than another key list. Use simple typed constants/structs, not a user-extensible registry.

Acceptance: every exposed CEL field resolves to an admitted logical descriptor and executable adapter field; negative signal cases remain covered; uint64 equality/ranges retain exact boundary behavior. Discovery and completion show units/presence rather than implying every field is an OTel attribute.

### 4. Medium: move the planning use case out of the executable before adding HTTP

`cmd/goo/plan_cel.go:planCEL` currently owns the complete compile/apply/source-plan pipeline and an anonymous receipt. Reimplementing this in a route would duplicate orchestration/version ownership. Extract a small pure package-level planning function and named input/receipt types with one owner; keep CLI file loading/flags in cmd and HTTP decoding/deadline policy in transport. The extracted function should call querycel and exploration, not import cmd or HTTP. Avoid making the reusable CEL compiler itself depend on file IO or server configuration.

The UI should stage the returned ordinary question through the existing question session. Its draft identity, late-response retirement, Apply/Run distinction, held return, save and capture ownership remain authoritative. Text and visual controls edit the same predicate draft. Arbitrary imported AST need not have invented source text: canonical serialization or explicit unsupported conversion is preferable to stale text silently becoming authoritative. Retain expression/library provenance separately when useful; ordinary query identity remains based on ordinary semantics.

Acceptance: CLI and HTTP produce the same receipt for the same fixed parent; planning performs zero source reads/publications; stale completions cannot stage a question; Run uses the existing executor. Round-trip tests include nested OR and severity expansion without repeated parent-AND accumulation.

### 5. Medium before HTTP reuse: state the concurrency and resource contract

`Program` has private immutable compiler state and detached getters, a good basis for safe reuse. It currently creates a CEL environment/program per Compile, has static/runtime cost limits and bounded inputs, and exposes synchronous `Eval`. `PlanHandler` explicitly checks deadlines before/after synchronous work; it cannot preempt compilation. This is not a demonstrated runaway bug in the bounded pilot, but a network authoring service needs a deliberate concurrency policy.

Initially keep compile requests bounded and cap concurrent work at the request owner. Do not add an unbounded expression cache or a goroutine per compile merely to race a deadline. Record aggregate compile latency and refusal codes without expression text/high-cardinality keys. Document whether Program supports concurrent Eval/Apply and prove it with concurrent reuse under `go test -race` before sharing instances. If reference evaluation becomes a request-serving loop, use CEL context-aware evaluation/cancellation with its documented polling contract; the currently tiny supported subset does not justify a clever cancellation API now.

If measurement later proves caching worthwhile, key immutable compiled artifacts by exact expression, signal, language/lowering semantics version and library set. Never cache a refined parent question as the program. CEL library version alone does not describe changes in admitted syntax, field exposure or lowering. Preserve a small explicit compiler semantics identifier when authoring provenance is persisted; do not support a matrix of historical evaluators by default.

Acceptance: concurrent immutable reuse and caller mutation pass race checks; bounded request concurrency and maximum-input refusal are exercised; shutdown/cancellation does not leak workers. Cache remains deferred until measured need.

## Product model boundaries and older proposals

Retain the typed normalization direction in PROJECT_DIRECTION and the pilot's strict missing-input/unsupported-subtree semantics. `measurements.Definition` binds native metrics/lifecycle kernels; `subjectfacts.Definition` compiles prepared facts; `exploration.Query` describes bounded raw records. These are meaningful different domains, not three interchangeable versions of one CEL object. Share low-level identity/type/diagnostic contracts where useful, but do not flatten counter reset, histogram temporality, lifecycle custody or unknown states into boolean record predicates. A raw severity predicate does not define application health.

Retain YAML-plus-CEL as an authoring direction and portable target compilation as explicit contracts. Defer whole-query builders, general derived measures, KCL selection, record-level Databricks lowering, and general extension registries until a concrete second user outcome needs them. Remove any assumption that Go users need fluent query builders, arbitrary `cel.EnvOption` injection into the source compiler, binary plugins, or a legacy compatibility layer. The public OTel CEL library can remain generally registerable while source lowering stays closed and reviewed.

Attribute presence is the next semantic extension after the bridge, not a generic map escape hatch. `Field.Type` admits string-valued resource/attribute fields and `Match` distinguishes absent from empty; numeric bindings explicitly declare interpretation and valid/missing/empty/invalid coverage. CEL map-key errors and source absence cannot be reconciled by defaulting keys to empty strings. Require checked presence-aware forms and explicit typing before exposing them in text/visual controls. Preserve missing/unknown outcomes instead of treating evaluation errors as false health results.

## Small staged acceptance

1. Consolidate predicate validation/refinement/cloning, logical field descriptors and typed authoring diagnostics. Keep all current pilot acceptance; add only tests for the new shared contracts.
2. Extract the shared pure planning use case and connect one source-free app draft bridge through existing HTTP and question ownership. Demonstrate syntax correction, visual/text round trip, stale response retirement, explicit Run, and save/reopen with ordinary questions. Reuse existing display/editor composition; no CEL-specific execution owner.
3. Add presence-aware attributes or a real managed definition consumer, selected by a user journey. Gate on reference/source parity and missing/empty semantics. Expand portability only with target-specific evidence.

This order complements 436's ownership consolidation rather than creating a competing platform program. No existing legacy API needs to be retained merely for compatibility, but semantic custody and query identity must survive intentional model simplification.

## Verification scope

Inspected actual source and existing focused tests. Reran the existing focused checks with the available `/workspace/.toolchains/go/bin/go` toolchain, existing module/build caches, `GOTOOLCHAIN=local`, `GOMAXPROCS=2` and `CGO_ENABLED=0`: `go test ./querycel ./cmd/goo -run '^(TestPlanCEL|TestChecked|TestAllSeverity|TestApply|TestUnsupported|TestBounded|TestLibrary)' -count=1`. Both packages pass (0.075s and 0.014s). The initial default-PATH invocation resolved to an unrelated `/usr/bin/go`; using the existing toolchain resolved that environment issue without an install. The pilot's broader native acceptance remains documented prior evidence, not rerun evidence from this review. No HTTP/UI implementation, browser execution, race check or benchmark is claimed by this document.
