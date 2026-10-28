# Library API usability review

Date: 2026-10-28. **Read-only review and acceptance criteria; proposals are not
implemented APIs.** This covers the whole project’s integration surfaces, not
only the recent host controls. No builds, browser campaigns or broad test suites
were run for this review.

The common path should remain ordinary OTel setup, a concrete question or YAML
definition, and supplied display data. Internal generics, reusable ownership
helpers and newer language features earn their place through measurable
correctness, reliability, performance or maintenance gains. They should not make
an application developer learn a second programming model to use Goo.

## Inspected integration paths

| Consumer | Existing entry | Usability implication |
| --- | --- | --- |
| Ordinary OTel producer | Generic startup, service/resource discovery and ad hoc telemetry packages | No source annotation or manifest is required. A library cleanup must preserve this first-use path. |
| Go application author | `//goo:operation`, `//goo:event`, standard OTel providers, opt-in Orchestrion module | Named context/error results and the restricted annotation dialect need one complete working example, not a new builder DSL. Weaving remains explicitly opt-in. |
| Go semantic consumer | `measurement.Definition`, `measurements.Question`, `exploration.Query`, concrete results and source adapters | Several deliberate semantic layers exist; similar package names can obscure which one a caller needs. |
| Declarative author | Telemetry, measurement and operation YAML packages; portable prepared-fact aggregate pilot | They have different source/grain/deployment requirements and should not be presented as interchangeable versions of one schema. |
| Component integrator | `ui/components.html`, JSON attributes, detached properties, native custom events | Components already support framework-neutral embedding. Consumers should not need application navigation, a query client or the full SPA asset list. |
| HTTP/client author | OpenAPI, direct Go routes and Python `/api` proxy | Structural documentation does not imply stable API versioning or identical limits, effects and error behavior at both boundaries. |

Sources include [README](../../README.md), [generic startup](../GENERIC_STARTUP.md),
[annotation contract](../GOO_ANNOTATIONS.md), the actual
[woven example](../../examples/operation-app/checkout.go),
[runtime](../../instrumentation/runtime.go),
[aspects](../../instrumentation/orchestrion.yml),
[measurement packages](../MEASUREMENT_PACKAGES.md),
[operation packages](../OPERATION_PACKAGES.md),
[portable compiler](../PORTABLE_COMPILER.md),
[components](../UI_COMPONENTS.md), [HTTP contract](../API.md) and
[executed acceptance](../API_ACCEPTANCE.md).

## Public API decisions

| Proposal | Assessment | Required boundary |
| --- | --- | --- |
| Explicit instrumentation `Runtime` with supplied OTel providers | Accept with the revisions below. Useful for isolated applications/tests and instrument-initialization errors. | Preserve existing annotation/package-function entry points; state ownership must actually follow the runtime. |
| Generic public query/expression builder | Do not introduce for cleanup alone. | Concrete Query/Question/Result types and existing validated YAML are a simpler common path. Internal generic machinery may serve these concrete APIs. |
| Checked hash/revision construction | Prefer an internal first increment. | Existing receipt bytes, algorithms and historical reopening must not change as an incidental error-handling cleanup. |
| Reuse request ownership and private component utilities | Accept incremental adoption. | Source effects and navigation remain in application adapters; independent owners must stay independent. |
| Preserve SSR/property/event contracts while reorganizing modules | Accept. | Keep attribute-only HTML, detached property assignment, ordinary custom events and native children working without the SPA runtime. |

### Explicit instrumentation runtime

The proposed plain entry is:

```go
// PROPOSED: not an existing exported constructor.
runtime, err := instrumentation.NewRuntime(instrumentation.Config{
    MeterProvider:  meterProvider,
    TracerProvider: tracerProvider,
})
if err != nil {
    return err
}
```

Prefer `(*Runtime, error)` if construction creates instruments: the existing
runtime discards instrument-construction errors. The runtime must not install or
shut down caller-owned providers/exporters. Document whether nil providers are
rejected or explicitly select the global OTel providers; do not silently alternate
between per-instance and global ownership after first use. Avoid requiring an
options chain, a generic type parameter or a global registration step.

Concrete `runtime.Operation` and `runtime.Event` methods are reasonable low-level
hooks. The current `Operation` finalizer `(error, bool)` is designed for weaving:
the aspect supplies panic detection and rethrows the original panic. It should
not become the ordinary application author’s obligation to write recover/repanic
boilerplate. The annotation example remains the common application path. A small
explicit runtime test can demonstrate isolation without pretending to implement
the weaving protocol:

```go
// PROPOSED test sketch; constructor is not implemented by this review.
ctx, finish := runtime.Operation(context.Background(), "checkout.reserve", "server")
_ = ctx // Pass the returned context to work performed within this operation.
finish(errors.New("test failure"), false)
finish(nil, false) // Must not record a second completion.
```

The actual aspects call **package-level** `Operation` and `Event`. Merely creating
a Runtime does not route woven calls through it. The first increment must either
scope its promise to direct/manual consumers or explicitly design and test a
weaving injection seam. Do not imply woven provider isolation before it exists.
Do not add context-based runtime lookup, per-function registration or a new global
installer incidentally during a mechanical refactor.

Acceptance requires two independent runtimes using the same operation name/role
to initialize and emit through their own providers, parallel test isolation,
exactly-once finalization, preserved return/error/panic behavior in the existing
weave path, and unchanged custom scope/schema/instrument identities. Outcome
initialization bookkeeping currently lives in a process-global map; relocating
instruments without relocating that bookkeeping would leave isolation incomplete.
The example-only synchronous log transport remains clearly separate from a
production OTel logging SDK.

## Concrete Go libraries and authoring

`measurement` is the prepared subject-fact aggregate/compiler pilot;
`measurements` binds native/lifecycle kernels to bounded source inputs;
`exploration` represents raw log/span questions. Keep that distinction explicit
in a concise package map with one external-consumer example per path. Do not
rename or merge public packages merely to make internal directory structure neat.

Important integration pitfalls visible in current code:

- `measurement.Read(io.Reader)` and `measurements.Read(string)` serve different
  schemas. A short “which package do I use?” example prevents more confusion than
  a polymorphic universal reader.
- `measurements.Source.Run(ctx, question, previous)` performs a source read.
  `measurements.Evaluate(question, rows, at, custodyDir, previous)` is not a blanket
  promise of filesystem-free evaluation: lifecycle processing can use custody.
  Document effects and fixed clocks next to the callable API, not only in a
  feature document.
- Exported map/slice-bearing DTOs are convenient, but caller mutation must not
  silently alter a compiled plan, held result, digest or replay basis. Document
  snapshot versus borrowed input ownership; preserve existing defensive copies
  where evidence needs them.
- Missing namespace, empty namespace and unrestricted namespace are different
  selections. The existing `identity.Namespace` API encodes this explicitly.
  Convenience functions must not collapse nil, missing and empty values.
- Exact counts, nanoseconds, revisions and decimal telemetry values must survive
  JSON/YAML and JavaScript round-trips. Avoid public convenience APIs that require
  converting them through float64/Number.
- Definition validation, planning, source execution and publication are separate
  operations. Common examples should demonstrate that planning does not read
  telemetry and that saving a definition does not deploy it.

The simplest authoring examples should start with a manifest-free log/span
package and a single native gauge. Operation manifests belong in a separate
source-aware example: `sourceDirectory`, exact declared operations and subject
identity are stronger requirements. The portable Databricks aggregate artifact
must remain labelled as a bounded prepared-fact compiler target, not a promise
that every raw telemetry package can already translate to a metric view.

## Frontend embedding and interaction contracts

Preserve three explicit component contracts rather than forcing one lifecycle
onto all elements:

1. Snapshot displays accept supplied JSON/properties and may replace their
   rendered children.
2. Passive shells retain caller-owned native table/details/action children.
3. Controlled selections emit intent and reflect a caller-supplied selected key,
   grouping or range; the embedding application owns the resulting action.

For each public element, a consumer needs a small load recipe naming its JS/CSS
dependencies, minimal input example, bounds/unavailable behavior, emitted event
shape and update/selection retention behavior. The current component index is a
good foundation, but an integrator should not need to copy the SPA’s long script
inventory or follow many feature-document links to display one chart. A future
entry module/package can improve this if it remains display-only and supports
the documented classic-script path; bundling is not itself an acceptance result.

Plain SSR attribute input and SPA property updates both remain first-class.
Attribute JSON must use normal template escaping; values are text/data, not HTML
or expressions. Preserve property-before-definition upgrade, detached assignment,
reconnection, malformed-input refusal, native keyboard interaction and a touch
fallback for pointer gestures. Test standalone and nested layouts: the recent
container-heading width regression illustrates why horizontal containment alone
does not prove readable text.

Public theme tokens and documented semantic markup should carry customization.
Avoid making consumers depend on private child positions, unadvertised global
helper load order or application-owned CSS selectors. A compact host adapter
must not accidentally change a pod display using the same primitive.

### Request ownership through response consumption

Current `ui/queries.js` attaches a query-cancel listener around `fetch`, then
removes it when the fetch promise returns a Response. Many callers subsequently
await `response.json()`. That is a concrete lifetime seam to investigate before
standardizing more request helpers.

The fetch AbortSignal can still abort body consumption after headers; removing
the additional cancel listener does not, by itself, prove every request is
broken. Conversely, a generation guard that blocks DOM publication does not
prove server work was canceled. Verify the actual transport and handler behavior.
Define whether an operation owns headers only, bounded JSON decoding, an SSE
reader, or publication through completion. A source-query JSON helper may own
fetch plus decode and cleanup internally, while the existing Response-returning
API remains compatible until callers are deliberately migrated.

Acceptance should hold an actual response body after headers, withdraw its owner,
then release the body and assert the intended cancellation notification, client
abort, absence of late publication and listener cleanup. Include failed decoding
and HTTP errors. Source-free plans, saved reads and publications must not acquire
source-query cancellation side effects through an overly general helper. Keep
SSE ownership explicit; do not fold every stream into a JSON convenience function.

## Acceptance before calling a cleanup complete

| Common consumer task | Minimum reviewable evidence |
| --- | --- |
| Instrument one Go operation | Existing annotation example builds/lints/weaves; ordinary OTel setup remains sufficient; optional Runtime example has no hidden registration. |
| Isolate two instrumentation tests | Separate providers receive only their own operations, zero-outcome initialization and drops; idempotent finalizers remain safe. |
| Evaluate/compile a definition | External-package Go example with concrete DTOs, validation errors and fixed time; exact deterministic output and unchanged historical receipts. |
| Use a non-Go OTel application | Manifest-free package/discovery path remains available; no Go hook or source checkout becomes mandatory. |
| Embed one display in SSR or an SPA | Standalone dependency recipe, attribute/pre-upgrade/detached/reconnect checks and zero fetch/subscription/publication. |
| Select a chart or attribute | Typed bounded intent, keyboard/touch alternative, exact scope preservation and no implicit source read. |
| Cancel an active query | Body-lifetime and late-result tests cover actual effects, not only the early fetch promise or DOM visibility. |
| Reopen shared evidence | Exact prior algorithm/revision/custody preserved, no silent recomputation or source read; failed publication retains recovery identity. |
| Claim a performance improvement | Representative before/after allocation, CPU, network/effect count or bundle evidence, with semantic behavior unchanged. |

Keep checks proportional to the touched boundary. Existing frozen API fixtures,
property/state tests, browser journeys and negative-control model replay provide
different evidence; none substitutes for the others. Prefer one bounded cleanup
with a concrete consumer gain over a new facade layer spanning the whole product.

## Specialist sketch disposition

The provisional Go Runtime sketch is **accepted with revisions**, as above.
Concrete public Query/Question/Result types, an internal-first checked hashing
increment and frontend private-helper/RequestOwner reuse are accepted directions.
Existing component SSR/property/event contracts are acceptance constraints.
Further public sketches should be checked against the consumer tasks in this
review before implementation; this document does not authorize or claim a new
stable external API, package release, runtime constructor or transport helper.

The [frontend specialist review](FRONTEND_LIBRARY_REVIEW_2026_10_28.md) was also
read in full. Its current-source reconciliation and private-first recommendations
are accepted, with these API-ergonomics clarifications:

- **Internal JSON helper: accept, clarify the return contract.** A sketch returning
  only `data` hides the HTTP distinction existing callers inspect. Prefer an
  ordinary small result such as `{ok, status, headers, data}` if those fields are
  required by the two pilot callers. Keep transport/abort/decode failure separate
  from domain-specific HTTP errors; preserve caller messages and required response
  headers. Do not return a body-consumed Response while implying callers can still
  invoke `response.json()`. Keep the current raw-Response function until its named
  consumers migrate; no new public transport API or generic fetch wrapper is
  needed. Source-query cancellation registration must not attach to every pure
  plan, saved read, publication or SSE stream.
- **Owner relocation: accept.** Move proven mechanics, not publication policy.
  A temporary alias for existing `RequestOwner`/`PendingScope` consumers is a
  migration aid, not a new public lifecycle protocol. Preserve the actual owner
  algebra: scope/route withdrawal differs from local metadata disclosure, and a
  family edit must not cancel an unchanged sibling. Do not turn component mount,
  disclosure or attribute update into an implicit subscription.
- **DOM helpers and styling: accept narrowly.** Parameterized private JSON/property
  helpers can remove real repetition, but caller-owned chart/table children,
  retained disclosure/focus, snapshot replacement and each element’s existing byte
  bounds remain explicit policies. Public user code and event details stay the
  same. A dependency check is useful before adding another load-order requirement;
  a mandatory bundler, generic custom-element superclass or whole-product CSS
  rewrite is unnecessary.
- **Reuse of existing work: accept.** The shipped facet composition, record/evidence
  adapters, definition rows and fixed workspace boundary already solve earlier
  extraction requests. Extend only demonstrated residual duplication, using a
  current consumer and the contract-to-test matrix. Do not recreate those APIs or
  require product users to understand their internal adapters.

These decisions require no new user-facing advanced API. Future advanced hooks
remain reasonable when a concrete embedding or performance/reliability need
cannot be met through the simple path, with that benefit and its compatibility
cost reviewed explicitly.

### Addendum: modern platform adoption

The added frontend platform matrix and the Go review’s recent-features matrix
were reviewed. **Their conditional, internal-first adoption is confirmed from a
design/usability perspective.** This confirms the proposed gates, not new browser
compatibility, runtime performance or independently executed feature tests.

For browsers, establish a supported browser/version policy before requiring new
APIs; the current Chromium campaigns do not establish other-engine support.
Optional AbortSignal composition/listener cleanup may simplify ownership, with
tested fallback and preserved deadlines, abort reasons and publication guards.
Yielding, containment or deferred loading needs a measured interaction benefit
and must preserve focus, keyboard/touch access, find-in-page and exact evidence.
Native HTML remains the ordinary interaction model. SSR attributes, detached
properties and existing events remain unchanged; no new gesture, syntax, framework
or module loader becomes mandatory. BigInt/string transport remains authoritative
for exact values; Date formatting and any later Temporal adoption do not redefine
identity, elapsed time or custody.

For Go, root-compatible benchmark/test context APIs and isolated newer-module
concurrency tests are reasonable selective tools. Respect the root module’s
supported version floor and each nested module’s declared floor. New aliases,
generic error helpers, runtime scheduling defaults and compiler/GC features do
not justify a public API redesign or a minimum-version increase by themselves.
Benchmark claims need equal-workload evidence; deployment/runtime claims need the
actual built-main compatibility settings and environment. Cleanup after canceled
test contexts and instrumentation panic tests require their original lifetime and
panic semantics. These additions preserve the plain Runtime/DTO/annotation paths
reviewed above rather than adding a second integration model.
