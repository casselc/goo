# Cross-stack library refactor proposal — 2026-10-28

**Proposal, revised for the user's prerelease and evidence priorities.** Current
API names, Go/browser versions and stored artifact formats are negotiable. Prefer
a clean first-party cutover over permanent compatibility wrappers when that gives
a simpler library. Evidence retention, custody hashes, replay receipts and historical
byte compatibility are not architectural goals unless a concrete feature needs them.

Specialists inspected the full project at `af107b7`, including all six Go modules,
frontend application/display libraries, Python forwarding, instrumentation/weaving,
authoring and verification. This is broad structural review, not a line-by-line
correctness audit. No production code, dependencies, versions, runtime settings,
tests or data were changed. No new regression campaigns or benchmarks were run.
The first review was recorded in `40d03aa`; this revision incorporates the user's
subsequent clarifications rather than carrying its compatibility-first policy forward.

## Specialist reports and confirmed API directions

- [Go libraries/runtime](reviews/GO_LIBRARY_REVIEW_2026_10_28.md): package/module
  coverage, structural proposals and verified recent Go facilities.
- [Frontend libraries](reviews/FRONTEND_LIBRARY_REVIEW_2026_10_28.md): lifecycle,
  component/library boundaries, ESM cutover and modern platform assessment.
- [Design/usability](reviews/LIBRARY_API_USABILITY_REVIEW_2026_10_28.md): explicit
  dispositions on the proposed breaking changes and evidence simplification.

The common path should stay ordinary OTel, understandable Go functions and supplied
HTML components. Internal generics and modern platform features can do substantial
work without exposing a generic algebra to callers. Public generics or iterators
are also allowed when a normal usage example shows a real benefit; none is needed
merely to consolidate existing query types.

## Implementation status

The Go baseline cut is implemented: all six modules declare Go 1.27.0;
`.go-version`, CI and Docker builders select the validated Go 1.27.2 compiler.
Root tests/vet and instrumentation tests pass, and all nested modules compile.
CI additionally checks instrumentation and both examples. Native verification
compilation is not a full Hegel/Quint or woven campaign result. See
[compiler setup and runtime defaults](PORTABLE_COMPILER.md) and
[the current validation entry](../VALIDATION.md#unified-go-baseline-2026-10-28).
The native ESM display graph is implemented, including attributes/records, field
pickers, exceptions, tags and aligned numeric distributions. Pure formatters, bins,
pattern decoding and visibility rules are separate from application owners. The
gallery loads only `display.mjs`, with no app namespaces or source requests.
The application-owner cutover is now implemented: 105 former classic/CommonJS
modules use explicit imports and `app.mjs` preserves mount order. The namespace
bridge and unused API assembly objects are removed. Predicate validation and
request/listener ownership are independent internal modules. The retained client
has isolated per-client state and keeps its existing default entrypoints.
See [the embedding contract](UI_COMPONENTS.md#native-module-entrypoints).
The browser response-consumption cancellation cut is also implemented: source
owners use `postJSON`, which retains cancellation through decoding and rejects
late aborted results. Real HTTP regressions cover Chromium, the Python proxy and
Go's already-correct body-generation lifetime. The first endpoint-policy and
server extraction is implemented: `internal/recordhttp` owns record-family limits,
deadlines, effects and typed handlers; `internal/queryhttp` owns shared admission
and cancellation mechanics. Command code composes their configured dependencies.
Native measurement/discovery and host/pod planners now have family-owned policies
and handlers in `internal/nativehttp`. Source and pure-plan mechanics are shared
in `queryhttp`, including their distinct deadline ordering and optional lifecycle
custody effect. Server startup now owns the result/explanation bank explicitly;
tail membership and comparison drafts also remain within their server registrations.
Cross-server/replacement tests distinguish live capabilities from durable captures
and retention recovery. Explanation/capture handlers and POST policies are now
extracted into their own internal packages, with explicit mixed publication
lifetimes and source-only-handler guards. Command admission composes family
lookups. Executable shutdown/export flushing is now implemented: five-second Go HTTP
drain/cancel/close, followed by independently bounded sample-owned SDK shutdown.
Legacy operation source GET/HEAD admission and post-validation source deadlines
are now implemented through `internal/operationhttp` policy declarations. The
source-free declaration bypasses admission; mixed publication/retention and SSE
retain separate owners. This does not complete all broader library cuts.

The first storage-independent numeric path is implemented: native
`measurements.Evaluate(question, rows, at)` rejects lifecycle storage effects;
`ObserveLifecycle` names those explicitly. Supplied callers and stored numeric
replay use the pure API. Source.Run remains the HTTP adapter's declared mixed
dispatch; existing result identity fields are preserved for current saved work.
Broader metadata/hash simplification still needs consumer-led follow-up.

## Preferred cuts

| Cut | Recommendation | Benefit and implementation boundary |
| --- | --- | --- |
| **1 — Simplify the scope** | Map libraries and stored/verification artifacts to current user workflows. Keep useful investigation, comparison, saved-question and annotation capabilities; remove or isolate bookkeeping with no consumer. Replace repetitive acceptance paperwork with concise current validation guidance. | A short keep/simplify/defer/remove list, not another universal registry. Keep targeted correctness, cancellation and failure tests. No data deletion occurs in this review. |
| **2 — One Go baseline** | Prefer Go 1.27 across all six modules, validating with the already used Go 1.27.2 toolchain. Update root/CI/build/setup declarations in one cut. Retain separate dependency graphs for instrumentation, verification and weaving. | One compiler policy and direct access to newer test/concurrency/error facilities. This is a maintenance/testing benefit, not a claimed speedup. Check actual built-main GODEBUG/container defaults and all modules; a named older consumer could justify a different negotiated choice. |
| **3 — Explicit frontend modules** | Prefer native ESM: first a display entry/feature graph, then the application graph. Replace globals/CommonJS dual paths with explicit imports and migrate Node tests, gallery, examples and Python asset serving together. Remove the old graph after the pilot unless a real consumer needs a transition. | Clear dependencies and simpler embedding/maintenance. A plain module script or feature import remains sufficient; framework-neutral SSR HTML and SPA properties remain capabilities, not frozen names. Bundling/lazy loading are separate measured decisions. |
| **4 — Storage-independent evaluation** | **First pure numeric/state path delivered; broader simplification remains.** Separate numeric/native evaluation and source reads from optional snapshots, custody/publication and retention experiments. Make lifecycle correction/storage effects explicit rather than hiding them in a broadly named Evaluate call. Minimize live result metadata and hashes to what actual consumers use. | A library consumer can calculate or query without learning an evidence subsystem. Keep exact numbers/time, units, reset/unknown semantics and honest population identity. Optional saves need only the data required by their workflow; they do not justify generalized cryptographic custody everywhere. |
| **5 — Request and server ownership** | Establish cancellation through headers/body consumption across browser, Python and Go, then use one clear JSON-read helper for appropriate source owners. Consolidate endpoint effect/admission/deadline policy and move a coherent HTTP family out of cmd. Add bounded executable shutdown/export flushing. | Reliability and easier source-route maintenance. This is an identified seam, not a reproduced backend leak. Keep pure computations, source reads, publications/ambiguous commits and SSE lifetimes distinct; cleanly replace obsolete helper signatures when useful. |
| **6 — Clear package and revision APIs** | Prefer renaming prepared aggregate compiler `measurement` to `subjectfacts`, distinguishing it from native/live `measurements`. Migrate imports/docs/tests directly. Delete unused hashes; replace unchecked Hash(any) only where a real definition/reference ID is needed, using checked concrete functions. | Import sites explain which library the user needs. No permanent alias package or unchecked legacy wrapper by default. Persisted development fixtures may be reseeded; a retained saved workflow gets an explicit migration/export decision. |
| **7 — Owned instrumentation runtime** | **Implemented:** `NewRuntime(Config)` eagerly creates instruments with explicit providers and surfaced errors. Direct methods own bounded initialization state; one-time `BindDefault` connects woven hooks. Samples and SDK/Hegel/woven fixtures use the new path. | Manual integrations/tests can isolate providers; generated annotations need no per-operation plumbing. Applications own SDK shutdown. Unbound hooks are no-ops, rebinding is rejected, and overflow skips synthetic zero seeding while retaining actual observations. |
| **8 — Measured/private optimizations** | Parse numeric times and prepare sort/group keys once; use narrow generic/stdlib helpers where they remove real repetition. Relocate existing request-owner mechanics and extract only compatible component-input leaves. Consolidate touched CSS. | Benchmark allocation-sensitive paths before claiming gains. Preserve meaningful semantic differences, not old serialization accidents. Each extraction needs a real consumer and should delete its superseded path. |

Start with a small scope/ownership decision and the Go baseline, then pilot the
display ESM graph and one storage-independent numeric path in parallel. Request
lifetime/route policy can proceed behind focused shared fixtures. Package renaming
and instrumentation deserve coordinated cuts rather than half-migrated facades.
Do not turn this into a prerequisite to every product feature or eight simultaneous
framework projects.

## Public API sketches and usability review

The reviewed Go runtime direction is now implemented (2026-10-28):

```go
rt, err := instrumentation.NewRuntime(instrumentation.Config{
    MeterProvider:  meterProvider,
    TracerProvider: tracerProvider,
})
if err != nil {
    return err
}
```

`BindDefault(rt)` is the one-time process binding for woven hooks. It rejects
rebinding; unbound hooks record nothing. Direct Runtime methods isolate manual
integrations/tests with caller-selected providers. Constructor errors are surfaced;
provider shutdown remains application-owned. Outcome initialization is bounded to
1,024 name/role pairs per runtime without dropping actual observations. See
[the implemented contract](GOO_ANNOTATIONS.md#runtime-ownership-and-startup) for
nil/noop providers, overflow, finalization and test-process semantics. Both samples
and the woven/Hegel/SDK fixtures use the new ownership path. This completes the
owned-runtime cut's implementation; it does not freeze the prerelease API.

The proposed frontend embedding can be as simple as:

```html
<!-- Implemented core/chart entrypoint; mixed legacy widgets remain outside it. -->
<script type="module" src="/display.mjs"></script>
<goo-table density="compact"><table><!-- native content --></table></goo-table>
```

Direct feature imports can serve SPA consumers. Document the CSS dependency.
Current property/event names may change if the replacement is clearer. What matters
is plain embedding, supplied-data displays, native children, keyboard/focus behavior
and no query client or hidden subscription in a chart. The proposed browser policy
is current/previous stable Chromium, Firefox and Safari; it is a target requiring
conformance tests, **not established support from current Chromium-only campaigns**.

No generic public query/evaluator hierarchy, universal store, framework migration
or mandatory custom-element superclass has a demonstrated consumer benefit here.
That is a benefit judgment, not a compatibility prohibition.

## Evidence proportional to a useful feature

Keep enough information to explain what is being viewed and to support actual
saved investigations, comparison baselines, reusable definitions, annotations or a
chosen training-data workflow. Those may need a question, scope, timestamps,
selected rows/values, labels or a small schema version. They do not automatically
need byte-identical historical replay, cryptographic IDs or custody receipts.

Move optional retention/XTDB/bitemporal experiments behind their own boundary unless
a named feature uses them. Remove per-result proof fields and transient hashing
when they have no consumer. Useful metric attributes and why-a-value-is-unknown
inspection remain ordinary observability details; hashes, lineage and proof chrome
need not be primary product concepts. No existing capture/annotation feature is
being removed merely because this review questions its infrastructure.

Old disposable development artifacts can normally be declared obsolete and reseeded.
If actual saved work or a future training dataset needs continuity, decide explicitly
what to migrate/export. Do not silently present an old snapshot as a new read or
change what a number means. No data or artifact has been deleted by this review.

Validation should be proportionate: focused unit/property tests, a few useful
cross-language fixtures, representative real-source/DOM cases, and adversarial
ownership schedules. Keep Bombadil/Hegel/Quint where they catch a distinct failure
class; consolidate repetitive finite scripts and generated receipt bundles when
they do not. Screenshots help changed visual interactions and explanation; every
internal cleanup does not need a new historical evidence package.

## Modern features without consumer cleverness

The Go review verifies generic aliases and benchmark/test-context APIs in 1.24,
stable synctest and WaitGroup.Go in 1.25, generic errors.AsType in 1.26, and newer
nested test facilities in the installed 1.27 toolchain. With a negotiated 1.27
baseline these become available throughout the project; adopt each where it solves
an actual problem. Synctest is for isolated in-memory concurrency, not a replacement
for actual HTTP/browser testing. WaitGroup.Go's panic contract matters for
instrumentation tests. Standard maps/slices operations and private generics can
simplify preparation, but shallow copies and semantic missing/empty distinctions
still require deliberate ownership.

Newer compiler/GC/runtime behavior may already provide benefits. Check actual build
and executable compatibility settings before tuning or attributing a speedup;
changing a go directive can change defaults. No runtime switch, unsafe allocation
strategy or JSON implementation swap is recommended without a measured problem.

Frontend signal-owned listener cleanup and compatible abort composition are useful
internal candidates. Task yielding, rendering containment and optional loading need
a measured interaction/startup problem. Negotiate the browser baseline and then use
its facilities directly; do not maintain fallback machinery solely out of habit.
Exact numeric strings/BigInt and sensible UTC formatting remain useful, while
Temporal or a new visual gesture needs a concrete product requirement.
