# Reconciled simplicity proposal

Status: follow-on proposal, 2026-11-17. The walkthroughs/reviews and generic
journey dependency cut are delivered. Product boundaries below are now accepted
user decisions; their implementation and the remaining refactors are pending. This reconciles the component reports
with the independent [Go](holistic-go.md), [architecture](holistic-architecture.md)
and [product](holistic-product.md) reviews, followed by the deliberately dissenting
[adversarial essentials review](adversarial-essentials.md).

## Accepted product decisions

Confirmed through the user’s interactive choices on 2026-11-17. These supersede
conflicting review recommendations; they are decisions, not claims of delivered
implementation.

| Area | Accepted direction | Preserve before retiring the parallel path |
| --- | --- | --- |
| Product scope | More aggressive pruning, including actual capability removals rather than menus alone. Default conceptual spine: **Explore, Services, Saved**. | Ordinary OTel exploration, reliable preservation, reusable team views and the shared grammar. |
| Operations | **Fold in, then remove.** Preserve useful conformance and subject-event workflows in Services/shared views; then retire the parallel Operations interface and storage path. | Declared-versus-observed meaning, reported outcomes/status distinctions, exact subject relationships, relevant source/custody facts and useful saved outcomes. Standard Go/OTel instrumentation stays. |
| Prepared HTTP evaluator | **Unify, then remove.** Express useful outputs through shared definitions and service views; remove its special panel and endpoint afterward. | Useful aggregate semantics and authoring/inspection flows. This is a separate decision from Operations. |
| Business lifecycle / warehouse output | **Explicit extension.** Retain the capability while separating implementation and test obligations from ordinary metrics and first use. | Lifecycle correctness and warehouse-export goals. Subject-event navigation does not require making the lifecycle evaluator a core prerequisite. |
| Training support | **Labels and export only.** Keep annotations, stable targets, corrections and useful export; retire dedicated task/adjudication machinery after those essentials are preserved. | Label meaning, target identity, correction history and usable export. Retiring workflow code does not authorize deleting stored decisions or annotations. |
| XTDB | **Isolate as extension.** Preserve native bitemporal research while removing its implementation/dependency/test obligations from the ordinary core path. | Honest as-of/correction semantics and explicit transaction outcomes in that extension; core Save/Keep/Capture remains independent. XTDB is already gated, so another hidden control is not sufficient isolation. |
| Authoring | **Saved + contextual editing.** Remove the separate Definitions workspace after contextual Edit/Build and Saved provide the useful authoring paths. | Creating, sharing, reopening and editing definitions/team views, with extensibility and the shared grammar intact. |

The choice of module, package, executable and registration boundaries still needs
an implementation design. Use explicit ordinary composition, not a new plugin
framework or a proliferation of feature flags. Core must be understandable,
buildable and testable without learning the lifecycle or XTDB extension internals.

Before retiring an interface/route/store, map its callers and stored artifacts,
prove the replacement journey, and record export/retention handling. No stored
user data is authorized for deletion. A prerelease compatibility floor does not
require keeping obsolete APIs or formats once their useful outcomes are moved.

The broader [conformance integration proposal](../../CONFORMANCE_INTEGRATION.md)
is **set aside** following the user’s clarification. Do not implement its new
span producer/capture family as a prerequisite for simplification.

The delivered first visible move places the existing Operations interface under Services
and removes the parallel workspace/navigation entry. Existing readers, renderer
and operation capture store stay intact. Subsequent backend consolidation must
show a concrete reduction in independent parts before adding a replacement.
Reusable authoring is also delivered under **Saved → Definitions**, with contextual
Build and annotated-package entry points selecting its record editor. The separate
Definitions workspace/navigation is removed. Existing definition APIs and stored
revisions remain intact. Prepared HTTP remains available under application contracts;
its evaluator retirement, extension isolation and other accepted cuts remain separate work.

## Decision

Keep Goo's core design: one Go server, typed questions, explicit source/storage
owners, pure semantic kernels and supplied-data display components. Simplify the
implementation outline and contributor workflow before adding another broad
feature. Do not replace it with a generic executor, repository, event bus or state
framework.

The code is not uniformly understandable in an afternoon today. The walkthroughs
make one flow traceable; compressed owner closures and cross-cutting lifetimes
still make whole-file understanding unnecessarily expensive. Formatting alone is
necessary but insufficient. File count and physical line count are not acceptance
criteria.

## Essentials decision before broad refactoring

The critic correctly challenges an assumption in the first three reports: active
callers establish reachability, not product necessity. Do not format and extract
every surface before deciding which ones belong in the ordinary product path.

The default conceptual spine is now accepted as **Explore → Services → Saved**,
with reusable authoring and annotation retained. The accepted table above replaces
earlier alternatives for Operations, prepared HTTP, lifecycle, training review,
XTDB and Definitions navigation.

The accepted table above resolves the main product-boundary choices. Task-path,
caller and artifact handling still gate actual removal. The critic’s cut table
records the original losses and counterarguments; accepted choices do not imply
permission to discard user work or unported outcomes.

**First bounded cut delivered:** the generic runner now invokes only selected
Playwright work. It does not manage Docker/Redis, reservation traffic, Python or
automatic diagnostic pruning. `--reservation` / `test:journeys:reservation` explicitly
select the original pause/drain/restore path. Named reservation campaigns and
smoke opt in; fixture-specific prerequisites remain with each journey.

This development dependency cut is separate from the accepted product boundaries
above. Apply those choices before expanding refactoring beyond retained core owners.
Acceptance covers exact argument/exit propagation without lab executables,
reservation cleanup with controlled external commands, and an actual generic
journey with Docker/Python blocked and Redis stopped. No new orchestration
framework was introduced.

After that cut, format/extract one retained core owner at a time. Do not turn the
candidate list below into a requirement to preserve or refactor everything.

## Bounded candidate slices

| Candidate | Outcome and scope | Acceptance / stop condition |
| --- | --- | --- |
| **1. Readable source and effects** | Choose one pinned development formatter. Format authored UI files in separately reviewable changes, starting with a representative owner and then population/capture/document owners. Add a concise publisher/consumer/effect/lifetime map at composition boundaries. | Formatting has no intended semantic change. Existing local/explicit-read, hidden/stale action, Return, focus and uncertain-write checks stay unchanged. A reader locates entry, state, retirement and effects without expanding multi-kilobyte statements. Stop before mixing owner extraction into the formatting diff. |
| **2. Explicit retained-view contract** | Characterize one fixed record visit, then replace `investigation-trail`'s blanket `goo-*` child-node restoration with an owner-provided reading snapshot or documented retained-view suspend/resume contract. | Find, page, selected occurrence, inspector and scroll restore even when selection is not rendered. Eight-visit bound remains. Return issues no source read; old callbacks remain withdrawn. No source token, active request or pending publication authority is serialized into reading state. Extend property/model cases for the changed lifecycle. |
| **3. Cohesive browser owners** | Extract one presentation-editor owner from `population-workspace`, and separate query/result and durable publication lifetimes in capture/document flows. Keep a short composition function with ordinary callbacks. | The same action has one owner and retirement path. Pending writes keep their actual acknowledged/uncertain outcomes. Passive components gain no IO. Do not split one mutable closure across files and call it decomposition. Stop after one demonstrable ownership improvement before repeating the pattern. |
| **4. Visible backend phases and checked identity** | Expose named admit/freeze, compile, acquire, decode/profile-check, evaluate, consistency and projection stages. Replace active `Hash(any)` calls with small checked identity-family helpers; its discarded marshal error is a permissive API problem, not proven dead code. | Preserve valid identity bytes unless intentionally changing a format. Invalid inputs produce no digest. Public entry points still admit their inputs; private admitted plans may remove repeated work within one operation. Exact plans, parameter bytes, source order/counts, raw presence, uncertainty and custody tests remain unchanged. Benchmark before claiming fewer allocations or lower latency. |
| **5. Small contract maps** | Reconcile ordinary field/type/signal/physical metadata with an explicit narrower CEL exposure list. Publish the finite carrier, display-witness and store/receipt maps where owners live. Name the shared conformance/capture carrier used to retire parallel operation storage. | A new field/profile has a short checklist and rejecting parity fixtures. Physical compatibility, semantic eligibility, display shape and runtime provenance remain distinct. Save, Keep, Capture and publication retry guarantees remain distinct. No schema DSL or universal retry/repository abstraction. |
| **6. Truthful contributor checks** | Share an explicit application-module list between local default checks and CI; keep optional native/integration tiers visible. Remove reservation-Redis requirements from generic runtime journeys. Publish a short current check matrix and demote historical ledgers from entry-point authority. | A deliberate nested-module failure fails local and CI checks. Generic journeys run without reservation services; reservation fixtures still restore their owned traffic. Required missing native tools fail rather than look green. Remove duplicate CI execution only after identifying unique assertions and measuring the change. |
| **7. Measured cleanup and product pruning** | Measure retained visits, repeated compile/copy and bounded catalog scans. Apply the accepted extension/retirement choices and give retained extensions a named owner and bounded checks. Retire an offline reference only after moving any unique useful assertions. | Use realistic bounds and before/after request, allocation, file-read or latency observations. Indexes retain authoritative artifacts and repair/direct-recall paths. XTDB setup/pooling changes require instance-backed transaction/await/retry checks. No bulk feature deletion or new infrastructure solely to finish an audit. |

Slices 4 and 6 can proceed independently of browser ownership work once their
boundaries and rejecting tests are explicit. Keep each code change small enough
for an independent reviewer to explain its effects. The candidate order expresses
risk and dependency, not a requirement for one enormous serial rewrite.

## Keep, remove, defer

**Keep:** inferred generic HTTP wire helpers; family-owned effect policies; typed
source adapters and pure kernels; exact numeric/null/presence semantics; source
versus publication lifetimes; immutable observations and explicit uncertainty;
negative controls that catch wrong scope/clock/authority.

**Remove when locally verified:** evidenced unused aliases, literal-only shims and
unnecessary namespace/protocol forwarding. These are small cleanup candidates,
not a reason to delete a component wholesale. Delete superseded APIs after moving
their active callers; prerelease status does not require compatibility wrappers.

**Design before executing accepted cuts:** port the selected operation/HTTP
outcomes, isolate lifecycle and XTDB, and retire adjudication while preserving
annotation/export. Map stored-artifact handling before removing code. The offline
Python reference remains a separate retirement candidate once unique assertions
are moved. Product observations/custody differ from disposable test diagnostics;
pruning implementation does not authorize deleting useful saved work.

**Defer:** universal registries/engines, a new state platform, speculative pools or
indexes, public clever type machinery, another chart portfolio and vendor parity
for its own sake. Modern Go is already used extensively; adopt improvements when
they simplify touched code, not as a churn target.

## Relationship to the feature roadmap

Insert the bounded core-dependency cut and readability/owner-contract work before
more feature expansion. **463 ordinary All-environments spans** remains the next
user-facing breadth proof. The critic disputes automatically advancing 464 just
to satisfy cadence: raw exponential-histogram inspection should follow if actual
source coverage makes it the next blocker. **465 capability consolidation** should
follow demonstrated duplication, not anticipate a universal framework. The span
carrier matrix and any raw-profile design review remain required. Do not pause
useful work until every audit recommendation is implemented, and do not commission
another whole-product audit after every small slice.

## Test the afternoon goal

After each structural slice, give an uninvolved reviewer a component walkthrough
and a time-boxed code-reading exercise. Ask them to identify:

1. The entry, authoritative state, side effects and retirement path for one flow.
2. Where invalid/stale/hidden work is refused and which focused tests prove it.
3. How a local Return differs from a source read, and how a durable write differs
   from either.
4. The smallest safe place to add one field, kind, profile or action.

Record friction and adjust the read-first map. Do not claim the time goal from
line counts, passing tests or the existence of these documents alone. Broad
performance and production reliability also remain matters for measured checks,
not guarantees supplied by this source audit.
