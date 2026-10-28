# Library API usability review

Date: 2026-10-28. **Read-only review and acceptance criteria; proposals are not
implemented APIs.** This covers the whole project’s integration surfaces, not
only the recent host controls. No builds, browser campaigns or broad test suites
were run for this review. Updated after the user clarified that this is prerelease:
current API shapes and compatibility floors are negotiable migration facts. A
second clarification makes evidence/custody machinery and development receipts
optional unless a concrete user or development feature needs them.

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


## Prerelease cutover policy

The user's clarification supersedes compatibility-first wording in the original
review: **clean breaking changes and higher platform minimums are acceptable.**
Prefer one coherent supported path over permanent wrappers, duplicate exports or
parallel implementations maintained only to protect prerelease call sites.

Evaluate a proposed cut by its resulting common application example, actual
callers/builds that must migrate, semantic and stored-data consequences, and the
reliability/performance/maintenance benefit. Go 1.27 unification, ESM/browser
baseline changes, package privatization, checked hashing replacement and a
coordinated Runtime/weaving cutover are all negotiable. No constructor, generic
parameter, compatibility alias or old module floor is sacred.

The invariants that remain are telemetry truth and understandable effects:
qualified OTel identity, exact values, unknown/no-data distinctions and meaningful
grain/interval semantics. Evidence/custody protocols are not independent product
requirements. If a retained feature promises saved results or historical meaning,
honor that promise or explicitly change/reset it; do not display an old result as
though it had newly computed semantics. Obsolete prerelease development artifacts
can be declared unsupported or reseeded rather than given a migration subsystem.
This review performs and authorizes no deletion.

The public common path should still be short and concrete. Advanced or internal
features are welcome when they improve it; migration work is a reason to plan the
cutover, not an automatic veto. Revised specialist API sketches need an explicit
usability check before implementation; the precise new binding/constructor/import
shape is not chosen by this policy update.


## Evidence proportional to the feature

The user does not consider evidence/custody a core UX requirement by itself.
Simplification can remove per-value Evidence chrome, lineage/proof views,
duplicated ledgers, frozen receipts and redundant acceptance work when they do
not help users or future feature development. This includes reducing development
artifact retention and declaring old prerelease artifacts obsolete/reseedable;
no actual data deletion is performed or authorized by this document.

Retain only the data and guarantees justified by a concrete feature:

- A saved question needs its reusable scope/definition, not necessarily raw result
  custody, a full source plan and several receipts.
- A saved capture or comparison, if kept, needs enough question/result context to
  reopen and explain what the user saved. It does not automatically need a permanent
  lineage graph, immutable publication protocol or a per-row proof inspector.
- Notes and annotations need useful scoped references and understandable behavior
  when referenced data is unavailable. They do not independently require every
  historical content hash to remain supported.
- A future reproducibility, correction, investigation-sharing or training feature
  may justify stronger provenance. State the actual use and needed fields first;
  a speculative future possibility is not a reason to retain all machinery now.

Normal correctness continues: unknown is not zero, exact OTel identity and numeric
semantics remain truthful, errors remain actionable and tests target real behavior.
A plain “Details” action may be more useful than an Evidence workflow. Decide the
user task first and reduce implementation and documentation to support it.

## Concrete revised cutover decisions

| Proposed cutover | Design/usability decision | Acceptance |
| --- | --- | --- |
| Unify all Go modules on 1.27 | **Accept.** A single supported toolchain can simplify development, builds and modern internal testing. | Update module declarations, CI, containers and examples together; inspect changed runtime defaults deliberately. No public generic API is required. |
| Rename `measurement` to `subjectfacts` | **Accept.** It distinguishes the prepared subject-fact compiler from native `measurements` more clearly. | Move real imports/docs/examples in one cut; remove the old package rather than require an alias. Keep the semantic source/grain distinction. |
| Replace `Hash(any)` with checked concrete domain revision functions | **Accept.** Explicit errors and domain names improve reliability and caller understanding. | Migrate all actual callers; no mandatory old `Hash` alias. If bytes/identity change, preserve only supported saved-feature needs or declare a prerelease reset. |
| Coordinate Runtime and default weaving binding | **Accept the coordinated direction.** Prefer one understandable runtime setup used by both direct calls and woven hooks over a manual-only feature plus permanent globals. | Define the concrete default binding/setup and concurrency semantics; update aspects, examples and tests together. Runtime/provider ownership and constructor errors stay explicit. No ordinary-app panic/recover boilerplate or mandatory generic configuration. |
| Native ESM display-to-app dependency graph | **Accept.** `<script type="module" src="/display.js"></script>` is a plain embedding recipe; direct feature imports provide a smaller path. | Display imports remain free of application reads/subscriptions. Migrate Node tests, gallery, serving and real consumers together; remove CommonJS/globals after the pilot unless a named consumer still needs them. SSR attributes and detached-property use remain practical capabilities. |
| Current/previous stable Chrome, Firefox and Safari target | **Accept as the proposed support policy, not a verified claim.** | Define what “previous” means and record actual versions/platforms at release. Run representative component/interaction conformance before claiming support; current Chromium runs alone do not establish the other engines. Optional unsupported targets need not force fallback code. |

These decisions approve the public shape and migration direction, not completed
implementation, cross-browser acceptance or a performance improvement. New syntax
and platform minimums are negotiable; ordinary use should still require little
setup and no internal type machinery. The exact Runtime default-binding function
name/behavior remains to be reviewed when its concrete sketch is available.

## Public API decisions

| Proposal | Assessment | Required boundary |
| --- | --- | --- |
| Explicit instrumentation `Runtime` with supplied OTel providers | Accept with the revisions below. Useful for isolated applications/tests and instrument-initialization errors. | A coordinated runtime/aspect/example cutover may replace package helpers; state ownership must actually follow the runtime. |
| Generic public query/expression builder | Do not introduce for cleanup alone. | Concrete Query/Question/Result types and existing validated YAML are a simpler common path. Internal generic machinery may serve these concrete APIs. |
| Checked hash/revision construction | Accept a checked replacement and removal of unchecked callers. | Handle errors explicitly; retain migration/identity guarantees only for concrete supported saved-data features, and allow a documented prerelease reset. |
| Reuse request ownership and private component utilities | Accept incremental adoption. | Source effects and navigation remain in application adapters; independent owners must stay independent. |
| Simplify frontend APIs and module delivery | Accept a clean cutover when the common path improves. | Keep useful SSR and detached-property embedding capabilities; update examples/events/imports together and remove superseded paths. |

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
between per-instance and global ownership after first use. Keep the common setup
short and concrete. A single explicit configuration/binding step is negotiable if it makes runtime ownership clear; avoid mandatory generic
parameters or a long options chain.

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
a Runtime does not route woven calls through it. A manual-only increment must say
so; the preferred clean runtime/weaving cutover should design and test the binding
seam together with aspects and examples. Do not imply woven provider isolation
before it exists.
Context-based binding or one explicit installer is negotiable as part of that
reviewed design, rather than an incidental hidden change. Do not preserve obsolete
package wrappers solely because callers can be migrated; remove them in the same
cut when that leaves a simpler setup.

Acceptance requires two independent runtimes using the same operation name/role
to initialize and emit through their own providers, parallel test isolation,
exactly-once finalization, preserved return/error/panic behavior in the existing
weave path. Custom scope/schema/instrument identity changes need explicit
versioning and matching source bindings rather than silent semantic aliases.
Outcome initialization bookkeeping currently lives in a process-global map; relocating
instruments without relocating that bookkeeping would leave isolation incomplete.
The example-only synchronous log transport remains clearly separate from a
production OTel logging SDK.

## Concrete Go libraries and authoring

`measurement` is the prepared subject-fact aggregate/compiler pilot;
`measurements` binds native/lifecycle kernels to bounded source inputs;
`exploration` represents raw log/span questions. Keep that distinction explicit
in a concise package map with one external-consumer example per path. A package rename, merge or private-package move is acceptable when it removes
real consumer confusion; migrate actual callers and examples in the same cut.
Directory neatness alone is not a sufficient consumer benefit.

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
entry module/package can improve this if it remains display-only. An ESM-only
entry is negotiable: update the standalone recipe, browser baseline and consumers
together, and remove obsolete global/classic wrappers where warranted. Bundling
is not itself an acceptance result.

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
API may be replaced in a coordinated caller migration; compatibility wrappers
are optional and should not survive without a concrete consumer need.

Acceptance should hold an actual response body after headers, withdraw its owner,
then release the body and assert the intended cancellation notification, client
abort, absence of late publication and listener cleanup. Include failed decoding
and HTTP errors. Source-free plans, saved reads and publications must not acquire
source-query cancellation side effects through an overly general helper. Keep
SSE ownership explicit; do not fold every stream into a JSON convenience function.

## Acceptance before calling a cleanup complete

| Common consumer task | Appropriate acceptance |
| --- | --- |
| Instrument one Go operation | Existing annotation example builds/lints/weaves; ordinary OTel setup remains sufficient; optional Runtime example has no hidden registration. |
| Isolate two instrumentation tests | Separate providers receive only their own operations, zero-outcome initialization and drops; idempotent finalizers remain safe. |
| Evaluate/compile a definition | External-package Go example with concrete DTOs, validation errors and fixed time; correct deterministic output; a simple documented prerelease reset is acceptable when old artifacts have no supported use. |
| Use a non-Go OTel application | Manifest-free package/discovery path remains available; no Go hook or source checkout becomes mandatory. |
| Embed one display in SSR or an SPA | Standalone dependency recipe, attribute/pre-upgrade/detached/reconnect checks and zero fetch/subscription/publication. |
| Select a chart or attribute | Typed bounded intent, keyboard/touch alternative, exact scope preservation and no implicit source read. |
| Cancel an active query | Body-lifetime and late-result tests cover actual effects, not only the early fetch promise or DOM visibility. |
| Reopen saved work, where retained as a feature | Saved questions/results remain understandable and usable; preserve only the identity/data needed by that feature. Remove obsolete proof/lineage requirements or document a development reset. |
| Claim a performance improvement | Representative before/after allocation, CPU, network/effect count or bundle evidence, with semantic behavior unchanged. |

Keep checks proportional to the touched boundary. Use focused Go/Node checks,
property/state tests, a representative browser journey or model replay when they
resolve the actual risk. Do not require every layer, a new protocol model, a saved
receipt or repeated whole-stack campaigns for each cleanup. Keep enough failure
diagnostics to develop and debug the feature. Prefer one bounded cleanup with a
concrete consumer gain over a facade or proof system spanning the whole product.

## Specialist sketch disposition

The provisional Go Runtime sketch is **accepted with revisions**, as above.
Concrete public Query/Question/Result types, checked hashing replacement and
frontend private-helper/RequestOwner cleanup are accepted directions. Existing
APIs identify migration work, not an obligation to keep aliases indefinitely.
SSR/property embedding, accessible interaction and exact semantic intent remain
acceptance capabilities even when names/imports/events change in a coordinated cut.
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
  from domain-specific HTTP errors. Keep meaningful error distinctions and required
  response headers; changed messages/statuses are explicit migration decisions. Do not return a body-consumed Response while implying callers can still
  invoke `response.json()`. Migrate named consumers and remove the raw-Response
  helper when no actual consumer needs it; do not require a compatibility alias.
  No additional public generic fetch wrapper is needed. Source-query cancellation
  registration must not attach to every pure plan, saved read, publication or SSE stream.
- **Owner relocation: accept.** Move proven mechanics, not publication policy.
  Update identified `RequestOwner`/`PendingScope` imports directly where feasible;
  a temporary alias is optional migration tooling, not a required public protocol.
  Preserve the actual owner algebra: scope/route withdrawal differs from local metadata disclosure, and a
  family edit must not cancel an unchanged sibling. Do not turn component mount,
  disclosure or attribute update into an implicit subscription.
- **DOM helpers and styling: accept narrowly.** Parameterized private JSON/property
  helpers can remove real repetition, but caller-owned chart/table children,
  retained disclosure/focus, snapshot replacement and each element’s existing byte
  bounds remain explicit policies. Public names and event details may change
  with coordinated documentation, example and consumer migration. A dependency
  check is useful before adding another load-order requirement;
  an ESM/package delivery cut is negotiable when it simplifies real integration.
  A generic custom-element superclass or whole-product CSS rewrite still needs
  a concrete benefit.
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
the chosen browser baseline and preserved deadlines, abort reasons and publication
guards. A fallback is required only for environments the project elects to support.
Yielding, containment or deferred loading needs a measured interaction benefit
and must preserve focus, keyboard/touch access, find-in-page and truthful values.
Native HTML remains the ordinary interaction model. SSR and detached-property
embedding remain useful capabilities; names, events, syntax and module delivery may
change with a clear recipe and declared baseline. Do not make ordinary users learn
a new gesture or type system merely to complete the same simple task. BigInt/string transport remains authoritative
for exact values; Date formatting and any later Temporal adoption do not redefine
identity, elapsed time or custody.

For Go, benchmark/test context APIs and isolated newer concurrency tests are
reasonable tools. The current root/nested module floors are migration facts, not
permanent restrictions: unifying on Go 1.27 is acceptable if it simplifies build,
test or runtime ownership. Update modules, CI, examples, containers and documented
requirements together. New aliases, generic error helpers and runtime features
should solve a real problem rather than become mandatory public abstractions.
Benchmark claims need equal-workload evidence; deployment/runtime claims need the
actual built-main compatibility settings and environment. Cleanup after canceled
test contexts and instrumentation panic tests require their original lifetime and
panic semantics. These additions preserve the plain Runtime/DTO/annotation paths
reviewed above rather than adding a second integration model.
