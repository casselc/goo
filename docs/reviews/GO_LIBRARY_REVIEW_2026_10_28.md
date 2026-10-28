# Go libraries and runtime structure review — 2026-10-28

Proposal only; no implementation, dependency or toolchain changes. Reviewed checkout `af107b7`. This review used source inspection and the installed compiler's source/specification, not new execution, load, browser or semantic acceptance. Existing checks in `VALIDATION.md` are historical evidence, not checks performed here. No broad suites or builds were run.

## Recommendation

Keep the concrete public question/result families. Concentrate changes inside the server composition, source policies, numeric preparation and instrumentation lifetime. The project already has useful semantic boundaries; another universal evaluator, generic query builder or store hierarchy would add migration cost without a demonstrated user benefit.

Read in context: repository `AGENTS.md`; the historical warning at `START_HERE.md:3`; current `docs/ARCHITECTURE.md`, `docs/KNOWN_ISSUES.md`, `VALIDATION.md`; `docs/HTTP_BOUNDARIES.md`; the October 22 refactoring/design reviews and October 25 architecture/platform reviews. Several earlier recommendations are now implemented. In particular, `cmd/goo/serve_http.go:28` has a generic startup profile, `cmd/goo/source_dependencies.go:12` snapshots source configuration, and `internal/httpjson`, `internal/sourcehttp`, `internal/immutablefile` already share narrow mechanics. Do not propose these as missing foundations. C1 generic startup acceptance is recorded at `VALIDATION.md:10245` and `:10302`, despite the older opening entries in that file.

## Coverage and current boundaries

This is broad structural coverage, not a line-by-line correctness audit of every file. Source-backed examples below cover every Go module and all runtime package families; test files were sampled as contract evidence.

| Area | Inspected examples | Assessment / next boundary |
| --- | --- | --- |
| Root runtime module | `go.mod:1`, `cmd/goo/serve_http.go:19`, `source_dependencies.go:12`, `query_budget.go:33`, HTTP registration families | Minimal root dependency set; startup composition improved. Route effect/policy membership remains manually synchronized. |
| Prepared aggregate language | `measurement/model.go:14`, `compile.go:13`, `evaluate.go` | Small, deliberate subject-fact language; ClickHouse and Databricks output are distinct. Keep prepared-source requirements, no arbitrary expression DAG. |
| Lifecycle preparation/corrections | `lifecycle/prepare.go:17`, `model.go`, `evaluate.go`, `correction.go` | Good comparable scope/subject keys and frozen-time semantics; repeated canonical serialization in sorting is an optimization candidate. |
| Numeric/native kernels | `metrics/counter.go:16`, `histogram.go:210`, `gauge.go:83`, `stock.go:53`, `state.go:53`, `counter_decimal.go:18`, fleet/ratio/window/time packages | Correctly separate stocks, counters, categorical state and histograms. Share preparation mechanics, not their outcome algebra. |
| Native source/planning | `measurements/model.go:45`, `source.go:26`, `discovery.go:46`, `instrument_capabilities.go:150`; host/pod/container/pool/runtime/exact-stream planners and replay | Concrete bindings make identity and profile assumptions inspectable. Package now combines several concerns; extract files/seams incrementally. |
| Record exploration/planning | `exploration/compiler.go:23`, `source.go:65`, `numeric_source.go:34`, `numeric_decimal.go:21`, `trace_search.go:89`, dashboard/population/comparison/explanation/tail families | Typed SQL lowering and honest population/read limits are strong. One large public result holds optional family outputs; preserve wire compatibility while improving internal projections. |
| Operation annotations / prebuilt | `annotations/manifest.go:57`, `package.go:73`, `clickhouse.go:72`, counters/histograms/ratios/fleet/subjects; `prebuilt/catalog.go`, `expand.go`, `clickhouse_http.go:45` | Explicit optional source declaration and semconv versions. Avoid silently merging distinct profiles or turning source manifests into verified emission. |
| Identity, authored context, custody | `identity/namespace.go`; `investigations/store.go:54`; `workbench/store.go:36`, `documents.go:105`, steps/meaning/views; exploration capture/package stores | Nil/missing/empty and family digest differences matter. Existing publication primitive is the appropriate level of reuse. |
| Optional retained evidence | `retention/store.go:124`, catalog/evidence/tail families, `catalog_page.go:15` | Fresh pgx connections isolate await-token sessions. Preserve ambiguous-commit recovery; pooling is not a default cleanup. |
| Instrumentation module | `instrumentation/go.mod:3`, `runtime.go:22`, `example_logs.go:22` | Global instrument/outcome initialization has process lifetime; telemetry provider ownership and cardinality deserve explicit contracts. |
| Weaving and examples | `cmd/goo/weave.go:17`, `annotations/manifest.go:73`, `verification/weaving/go.mod:3`, weaving integration fixtures; both example `main.go` files | Opt-in tooling stays out of root dependencies. Example shutdown should teach bounded export flushing; manifest emission and successful weave are different outcomes. |
| Verification and cross-stack | `verification/go.mod:3`, `verification/README.md:34`, numeric/identity/source/storage/HTTP property families; `ui/queries.js:3`, `server.py`, API/OpenAPI | Strong negative controls and observed ledgers. Add focused seam fixtures rather than multiplying whole-stack campaigns for mechanical changes. |

## Ordered proposals

### 1. Make route effects and admission policy one reviewed registration concern

**Priority: high; internal API only.** `cmd/goo/query_budget.go:34` contains a long POST path expression. `docs/HTTP_BOUNDARIES.md:18` identifies telemetry-reading GET operations; `:50` explicitly records that GET execution, operation investigations, tail and retention do not all use that wrapper. This is a confirmed policy split, not proof of an exploit or a newly reproduced overload defect. `/compare` is pure supplied-summary computation (`cmd/goo/exploration_http.go:21`); `/tail` owns independent SSE slots/deadlines (`cmd/goo/tail_http.go:15`). They must not inherit ordinary source admission merely because they are POSTs.

Introduce a small **internal** registration value with method/pattern, effect, admission class and deadline owner; retain each handler's body limit/status mapping. Use that inventory for a conformance test against registered routes and documented proxy/OpenAPI coverage. Start by reproducing existing membership exactly; broaden GET source admission only as an explicit behavior decision. Keep SSE separate. Extract cohesive handler groups to `internal/httpapi` after dependencies are explicit; `cmd/goo` should eventually assemble CLI/server configuration, not own all application workflows. Do not export a generic endpoint framework.

Benefit: adding a source route no longer depends on remembering another path list. Risk: accidental changes to timeout, cancel-ID or pure-route semantics. Migration: freeze current policy table first, move one family, compare direct/proxy status and source-call counts, test admission release on failure/cancel and cancellation-before-admission tombstones. No wire/API change is required.

### 2. Define cancellation through response consumption and bounded shutdown

**Priority: high; coordination with frontend.** Go source requests already use caller contexts and close bodies (`internal/sourcehttp/read.go:15`, `exploration/source.go:87`, `measurements/source.go:162`). The browser helper removes its explicit server-cancel listener when `fetch` returns headers (`ui/queries.js:8`), while callers consume the body afterward. Browser abort still affects its fetch body; the uncertain part is explicit server-cancel delivery in the header/body interval and when the proxy disconnects. Do not claim a reproduced server leak from this code alone.

Agree one fixed-query contract: ownership lasts through body consumption/error; disconnect or explicit cancellation propagates to the active Go handler and downstream request; late response publication is rejected by the existing frontend request owner. Keep query IDs ephemeral, unrelated to authentication or saved-question identity. Use an explicit completion hook or a shared decode helper in the frontend after its compatibility review, not a new universal lifecycle abstraction.

`serveHTTP` currently calls `ListenAndServe` directly (`cmd/goo/serve_http.go:18`); example apps terminate through `log.Fatal(...ListenAndServe())` (`examples/report-app/main.go:65`, operation app final statement), without a shutdown/export-flush path. Add a normal `run(ctx)` owner with bounded `Server.Shutdown`, then bounded tracer/meter shutdown in examples. Treat `http.ErrServerClosed` as normal shutdown. Use `signal.NotifyContext` at the executable boundary; do not make libraries install signal handlers.

Benefit: clearer request/resource ownership and reliable final telemetry at orderly exit. Risks: shutdown must not wait forever or change SSE lifetime; cancellation after an optional XTDB commit does not imply rollback. Test header-then-delayed-body, early disconnect, already-cancelled admission, proxy forwarding, slot recovery, SSE disconnect, normal shutdown and exporters that refuse to finish. Existing ownership and commit-recovery contracts remain authoritative.

### 3. Finish source-policy factoring without erasing adapter semantics

**Priority: high/medium; internal API only.** The existing `sourcehttp.Post/ReadBounded` deliberately leave URL policy, status ordering and row interpretation to callers. That is good. The next duplication is policy construction, not raw HTTP dispatch. Compare `exploration/source.go:65` (all-control escaping; cancellation-on-close, memory/read limits; status before body) and `measurements/source.go:144` (its own replacer; 10,001 row setting; status after bounded read), plus annotations/prebuilt source adapters. Differences may be intentional; none should be normalized silently.

Create a reviewed adapter-policy matrix, then extract only demonstrated identical escaping/settings fragments. A nonexported ClickHouse escaped-value encoder is worthwhile if corpus tests prove byte-for-byte equivalence for adopted callers. Keep SQL compilation, family response budgets, row validation, partial-result refusal, custody and basis calculation where they are. A concrete small transport config beats an interface with generic row/evidence callbacks.

Benefit: fewer divergent escaping/budget edits. Risk: an apparently equivalent encoding can change exact parameter semantics or error behavior. Tests: controls, backslash, literal tabs/newlines, Unicode, empty strings, endpoint query parameters, non-2xx, oversize/read errors, cancellation and no partial rows. Reuse `*_source_transport_contract_test.go` and verification sourcehttp properties. Performance of shared transport code is not a justification; benchmark before claiming allocation/latency gains.

### 4. Prepare numeric points and canonical ordering keys once

**Priority: medium; internal representation.** Gauge and stock evaluators repeatedly parse `EndNS` inside sort comparators (`metrics/gauge.go:119`, `stock.go:92`). `metrics/counter.go:69` already demonstrates the better pattern: a private parsed point with raw input, parsed times and exact value. Canonical JSON is repeatedly marshaled during sorts in `lifecycle/prepare.go:37`, `metrics/window_discovery.go:101`, `histogram_time.go:89`, and `exploration/dashboard.go:464`. These are concrete repeated operations; their production cost is unmeasured.

Use prepared records such as `struct { raw GaugePoint; end int64; streamKey string }`, and decorate entries with their **existing** canonical JSON key before sorting. Preserve old comparator/tie behavior and exact key bytes. A typed identity struct works for fixed comparable scope keys (as lifecycle already shows); map-bearing stream identity still requires a stable encoding or interned key. Do not concatenate labels with separators and risk collisions.

Share only common grouping/preparation loops across gauge/stock/state where that reduces real duplication. Keep stock epochs, gauge observations, freshness, conflicts and categorical profile validation explicit. No public `Evaluator[T,Policy,Identity,Semantics]` hierarchy.

Benefit hypothesis: less parsing/marshaling, simpler comparators and predictable error ownership. Benchmark 1/100/10,000 points, many streams, duplicates and adversarial ties with allocation reporting. Accept only with matching replay/golden bytes, shuffled-input invariance, conflict/missing/stale cases, and unchanged nil/empty maps. Do not advertise a speedup before measurements.

### 5. Make hashing and cloning contracts explicit before generalizing helpers

**Priority: medium, reliability first.** `measurements.Hash(v any)` ignores `json.Marshal` errors (`measurements/model.go:120`). Unsupported values therefore hash the failed marshal's empty byte slice; this is an actual exported-function behavior, not evidence validated measurement definitions currently collide. Internal unchecked marshaling often operates on constrained structs/maps and has a narrower practical risk. `workbench.digest` already returns an error (`workbench/store.go:36`), whereas investigations additionally normalizes embedded JSON with `UseNumber` (`investigations/store.go:54`). A universal canonicalizer would change persistent IDs.

Add checked domain-specific revision helpers internally and migrate error-capable paths first. If an exported checked helper is useful, make it additive (illustrative `DefinitionRevision(Definition) (string,error)`), retaining `Hash` behavior for compatibility until a documented deprecation decision. Do not introduce a public generic `Hash[T]`: a type parameter alone cannot prove JSON encodability. Include malformed `RawMessage`, nonfinite floats and unsupported custom values in targeted boundary tests.

Audit JSON roundtrip cloning (`exploration/population.go:23`, `measurements/host_dashboard.go:59`, pod/container/pool planners) individually. Prefer named clone functions where structure/ownership is stable, but only after preserving deep-copy behavior. `maps.Clone` is shallow; `metrics.cloneStream` (`counter.go:78`) deliberately turns nil maps into empty maps. Replacing it with `maps.Clone` without normalization changes identity/replay bytes. Freeze old IDs and clone-mutation tests before edits. This is primarily contract clarity; reflection removal is not a performance claim.

### 6. Give instrumentation a bounded, testable lifetime without breaking weaving

**Priority: medium; additive public API candidate requiring design/usability review.** `instrumentation/runtime.go:22` has global `sync.Once`, instruments and `initializedOutcomes sync.Map`; public `Operation` accepts arbitrary name/role strings (`:44`). The map retains each name/role for process lifetime, although scanner-validated names are bounded (`annotations/manifest.go:57`). Instrument construction errors are ignored (`runtime.go:31`). Whether provider replacement is safe depends on SDK/global-provider behavior, so verify that behavior in the pinned OTel version before claiming a replacement bug.

Prefer explicit ownership for applications/tests that need it, keeping all current package functions and generated weaving calls working:

```go
// Existing callers and generated code remain valid.
ctx, finish := instrumentation.Operation(ctx, "checkout.reserve", "server")

// Illustrative additive API; names and exact config await design review.
runtime, err := instrumentation.NewRuntime(instrumentation.Config{
    MeterProvider: mp,
    TracerProvider: tp,
})
ctx, finish := runtime.Operation(ctx, "checkout.reserve", "server")
// Low-level compatibility hook; callers must report actual result/panic state.
// Ordinary woven application code should keep annotations, not add this boilerplate.
```

Design review: accept only with revisions. The candidate signature is `NewRuntime(Config) (*Runtime,error)`, with explicit nil/default-provider policy and instrument-creation errors. Scope the first additive runtime to manual integrations and isolated tests: current aspects call the global `Operation`/`Event`, so merely constructing a Runtime does not isolate a woven application. Any woven binding needs its own reviewed migration. Do not require ordinary users to write panic/recover/finalizer boilerplate.

The application owns SDK provider installation, exporters and their shutdown. Runtime owns instrument handles and outcome-initialization state, never shuts down caller providers, and must specify whether registering operations is static/bounded or dynamically limited. `WithSubjectScope` remains context-based and interoperable. A default runtime wrapper preserves existing convenience; weaving can retain package calls until a separate explicit configuration mechanism is justified. Avoid mandatory context-carried runtime injection or generic hooks.

Benefit: isolated tests, clear cardinality/initialization lifetime and observable constructor errors. Risks: more setup, changed zero-series emission, invalid provider assumptions, conflicting default/global configuration. First document actual pinned provider behavior; test two independent runtimes, repeated finish, error/panic outcomes, bounded operation registration, selector panics and late provider setup. Benchmark steady-state allocations only if optimization is a goal. No reflection-free claim: `reflect.TypeOf(err).String()` preserves useful concrete error classification; changing it needs semantic evidence, not novelty.

### 7. Keep evidence publication/recovery semantics while tightening resource ownership

**Priority: medium.** `internal/immutablefile/publish.go:11` already documents link-before-directory-sync ambiguity; its three operations intentionally preserve collision and sync differences. Workbench, investigations, exploration packages/captures and lifecycle custody should keep distinct validation/digest/reload policies. Do not replace them with a generic repository whose defaults invent durability or overwrite collision winners.

For optional XTDB, `retention.Store.connect` sets session `AWAIT_TOKEN` (`retention/store.go:124`) and callers close per-operation connections. That is a useful isolation boundary. Consider a small consumer-owned connection function for fault tests, but do not add pooling until connection setup is measured as material and token/session reset semantics are proven. Audit cleanup with cancelled contexts and bounded independent cleanup only where the pgx version actually needs it; `Close(ctx)` behavior must be verified rather than assumed to leak.

For CLI weaving, `cmd/goo/weave.go:39` writes a manifest before running the tool subprocess (`:53`). This is acceptable as a declaration artifact, but output must not imply linked-build verification succeeded. Add context cancellation to subprocess ownership when the CLI gains a context; retain explicit module opt-in and no dependency edits. Keep provenance (scanner/tool/config revisions) inspectable before proposing richer automatic code-generation machinery.

Benefit: clearer failure/retry and tooling contracts. Risks: ambiguous commit acknowledgement and partial publication cannot be turned into automatic retry success. Reuse storage/negative controls for collision, directory-sync error after publication, corrupt winner, cancellation, commit-then-ack-loss recovery and await-token isolation. Preserve direct file-backed document loading without XTDB.

### 8. Modernize private code with the existing Go floor and focused conformance gates

**Priority: medium/low; no version changes.** Actual directives: root `go 1.24.0`; instrumentation, both examples, verification and verification/weaving `go 1.27.0`. Installed compiler reports `go1.27.2 linux/amd64`. These are local verified facts, not predictions. The root must not accidentally acquire a 1.27-only dependency/construct because verification runs a newer compiler. The root's small dependencies and separate native verification/weaving modules are strengths.

Useful existing standard-library facilities include `slices.SortFunc`, `slices.Clone`, `maps.Clone`, and (available by the root floor) `maps.Keys` with `slices.Sorted`. Their signatures were checked in installed `src/slices/{sort,slices,iter}.go` and `src/maps/{maps,iter}.go`; introduction versions were checked against installed `api/go1.21.txt:349,377,401` and `api/go1.23.txt:54,95–97`. Apply them to touched private code when clearer, retaining nil/empty semantics and required stable ordering. Existing private generics such as `measurement.names[V]`, `lifecycle.orderedKeys[K,V]` and `retention.paginateCatalog[T]` are appropriate: they eliminate a bounded mechanical repetition without forcing users to reason about type parameters.

| Proposed feature | Compatibility floor / scope | Compatibility requirement |
| --- | --- | --- |
| Internal route policy, explicit handler dependencies, context propagation | Root Go 1.24; existing `net/http`/`context` facilities | No new runtime dependency or public wire change. |
| Typed prepared points and cached canonical sort keys | Root Go 1.24; ordinary structs/functions | Existing JSON/hash ordering and exact numeric semantics remain frozen. |
| `slices.SortFunc`, `slices.Clone`, `maps.Clone` | Available before root floor (Go 1.21 standard-library packages) | Shallow copies require explicit nested-map ownership; nil normalization preserved. |
| `maps.Keys` → `slices.Sorted` | Go 1.23 iterator additions, within root 1.24 floor | No public iterator-returning API required; deterministic output still explicit. |
| Private generic mechanical helpers | Go 1.18 language facility, within root floor | No generic public query/evaluator hierarchy. |
| Checked revision helper | Root Go 1.24; ordinary concrete function | Additive public helper only if needed; existing Hash and stored bytes preserved. |
| Optional `NewRuntime` | Instrumentation already declares Go 1.27 | No root minimum bump; preserve annotation/global entry points; provider ownership explicit. |
| Hegel/Quint/weaving integration | Separate verification/weaving and verification Go 1.27 modules | No native/tool dependencies added to root runtime module. |

Use ordinary generic helper functions or concrete methods. This review relies on no newly imagined generic method, specialization, const-generic, ownership or higher-kinded-type feature. The installed specification explicitly disallows type parameters on interface methods (`doc/go_spec.html:1297`); no future language proposal is needed for these changes. Modern syntax is not itself a benefit, and no runtime switch, JSON implementation or toolchain bump is recommended.

Conformance should cross the actual seams: exact integers/timestamps beyond JS safe integer range; nil vs empty resource/schema/namespace; canonical decimal and histogram boundaries; source-free plan/reopen versus explicit execution; Go status/error envelope through Python forwarding; cancellation from browser through body consumption. Retain shared JSON fixtures consumed by Go and existing Node/Python contracts. Use existing request owners in the frontend; this Go proposal does not create new public element APIs. Run root focused tests with the declared minimum toolchain in CI when available, separately from 1.27 verification/weaving; do not claim this review established minimum-toolchain acceptance from a 1.27 compiler.

## Sequence and stop conditions

1. Freeze the current endpoint/effect matrix, canonical IDs and cross-language identity/window fixtures. Reconcile old review findings against delivered C1/shared primitives.
2. Land policy registration and focused cancellation/body-lifetime contracts; add orderly executable shutdown. Preserve SSE and optional-store lifetime distinctions.
3. Consolidate only equivalent source policy fragments using transport fixtures. Move one coherent HTTP handler family after dependencies are explicit.
4. Add checked internal hash helpers and explicit clone contracts, with unchanged persisted bytes. No bulk codec/canonicalizer rewrite.
5. Benchmark prepared numeric/sort keys, then adopt only clear measured or maintainability wins. Keep numeric domains and profile rules separate.
6. Review instrumentation constructor usability and ownership, then add it compatibly only if tests/apps need that control. Keep default weaving working throughout.
7. Apply standard-library cleanup opportunistically. Keep verification/tooling dependencies outside root and run feature-specific existing property/negative suites when implementing each cut.

Stop an extraction when its config starts mirroring every family's semantic branches, when it changes hash bytes without a versioned migration, or when it requires users to learn a new generic algebra for an existing concrete task. Do not pool optional database sessions, parallelize independent telemetry reads, stream new JSON formats, or replace big-number arithmetic solely for modernity. Parallel reads need explicit source budget/consistency analysis; reduced wall time would be a benchmark hypothesis, not an atomic snapshot.

## Existing abstractions to preserve / rejected overengineering

Preserve pure `Compile`/`Evaluate` and explicit-time inputs, qualified `Field` lowering, version-pinned capability plans, exact decimal strings and bounded big-number arithmetic, unknown/no-data/pending states, comparable lifecycle scope keys, source-neutral workbench reference callbacks, immutable publication and server-held snapshot promotion without rereads. Preserve separate record populations, native intervals, page-level trace analysis and prepared subject facts.

Reject a public `Query[TSignal,TMeasure,TGrain,TBackend]` builder, a universal `Result[T]` rewrite, generic numeric policy plugins, an all-store repository, a reflection ban, a wholesale package rename, and replacing `encoding/json` or the Go runtime without measured evidence. The public result DTOs have real compatibility costs; private typed projections and additive constructors achieve most benefits with a much smaller adoption burden.

## Recent Go assessment: explicit adoption gates

This matrix supplements the small-helper matrix above. Versions were checked against the **installed** toolchain's API history/specification/source, not inferred from the calendar. “Available in nested modules” means their existing Go 1.27 declarations permit the API; it does not authorize moving their dependencies or minimum version into root. No benchmark or runtime experiment was performed.

| Facility and verified version | Concrete relevance | Decision and gate |
| --- | --- | --- |
| Generic aliases, Go 1.24 (`doc/go_spec.html:2489`, `:8834`) | Could preserve an old exported name while relocating a genuinely generic type. Goo's concrete question families do not need a generic alias facade. | Available at root floor; defer unless a real compatibility move needs one. Do not invent public alias/type machinery just to use the feature. |
| `testing.B.Loop`, Go 1.24 (`api/go1.24.txt:212`, `src/testing/benchmark.go:502`) | Suitable for new prepared-point/canonical-sort-key microbenchmarks in proposal 4. Makes benchmark iteration ownership explicit. | Root-compatible internal test candidate. Keep setup outside the measured body, report allocations, compare equal workloads; API adoption itself supplies no speedup evidence. |
| `testing.T.Context` / `B.Context`, Go 1.24 (`api/go1.24.txt:211,216`) | Test-owned contexts can replace unrelated `context.Background()` in cancellable source/handler tests. Context cancellation is tied to test lifetime, not the production query deadline. | Root-compatible selective adoption. Cleanup that must run after test cancellation needs its own bounded context; do not mechanically pass the cancelled test context to exporter shutdown. |
| Stable `testing/synctest.Test` and `Wait`, Go 1.25 (`api/go1.25.txt:107–108`; package docs `src/testing/synctest/synctest.go:5`) | Useful for isolated cancellation/admission/timer state tests; deterministic fake time can reduce sleep-based flakiness. Package guidance excludes real network/external-process interaction. | Candidate in existing 1.27 verification modules for self-contained in-memory tests. Not a root test import at the supported 1.24 floor; do not raise root merely for this facility or claim it replaces actual HTTP/proxy/browser cancellation tests. |
| `testing/synctest.Sleep`, Go 1.27 (`api/go1.27.txt:288`) | Additional facility available to nested test modules. Existing injected clocks and `synctest.Test`/`Wait` may already express the needed property. | Deferred until a concrete test needs its documented behavior; no new helper layer or minimum bump. |
| `sync.WaitGroup.Go`, Go 1.25 (`api/go1.25.txt:96`; `src/sync/waitgroup.go:223`) | Can reduce launch/Add/Done mistakes in new verification concurrency harnesses. Its function must not panic, which matters for instrumentation panic-path tests. | Nested-module candidate only. Keep existing root Add/Done where root 1.24 is supported; no mechanical conversion of panic-oriented tests and no concurrency redesign. |
| Generic `errors.AsType[E error]`, Go 1.26 (`api/go1.26.txt:129`, `src/errors/wrap.go:167`) | Can make typed error inspection concise in nested verification. It does not improve unstructured string errors, choose HTTP status, or provide a reason to redesign all public errors. | Available in nested 1.27 modules; defer absent a real typed-error consumer. Root uses compatible `errors.As`/`errors.Is`; do not raise root for syntax reduction. |
| Container-aware default `GOMAXPROCS` and periodic adjustment, Go 1.25 (`doc/godebug.md:267–276`, `src/internal/godebugs/table.go:30,72`) | Relevant to the containerized query server's CPU scheduling and load experiments; complements but does not replace explicit query/SSE budgets. | Deployment/runtime assessment, not library refactor. Both settings have `Changed:25, Old:"0"`: a root Go 1.24 executable built by 1.27 does **not** automatically get the same compatibility defaults as a nested Go 1.27 executable. Verify actual built-main defaults and environment before any claim or opt-in. No GODEBUG change proposed here. |
| Newer compiler and GC implementation, locally installed Go 1.27.2 | Installed `src/internal/buildcfg/exp.go:81–88` sets `GreenTeaGC`, `SizeSpecializedMalloc` and `JSONv2` baseline flags true. This proves local build defaults, not release-introduction dates, effective per-build overrides, or workload performance. | Prefer measuring the current compiler/runtime before rewriting allocation-sensitive code. No compiler specialization/public-generic promise, no manual pool/arena/unsafe rewrite, no JSON codec swap and no GC tuning recommendation. Do not infer that every baseline flag changes the public `encoding/json` contract. |

GODEBUG compatibility belongs to the built executable/test main, not whichever library or verification module happens to be read first. Installed `doc/godebug.md:65–92,111–144` documents toolchain defaults amended by the work module/workspace `go` version, explicit `godebug`/`//go:debug` overrides and environment settings; workspace directives can supersede module directives. Runtime implementation optimizations are not uniformly governed by `go.mod`, whereas specific compatibility behaviors have documented version gates. Before future deployment tuning, inspect the actual main package/build configuration (`go list -f '{{.DefaultGODEBUG}}'` in the relevant build context is the documented inspection mechanism), Docker/runtime environment and any explicit `GOMAXPROCS` setting. This review does not assert those effective values for running images and does not change them.

Adopt modern Go where it reduces a real ownership/testing/maintenance problem within an existing module floor. The strongest near-term additions are root-compatible benchmark/test context APIs and nested isolated concurrency tests; the generic alias/error facilities currently offer little product benefit. Automatic compiler/runtime changes are a reason to establish a measured baseline, not a reason to promise a speedup or force a toolchain upgrade.
