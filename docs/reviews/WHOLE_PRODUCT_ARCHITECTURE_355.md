# Whole-product architecture checkpoint — 355

Proposal after 352 breadth → 353 depth → 354 consolidation. At this review,
354 source acceptance is closed but full runtime/native gates are still pending;
finish those before selecting the next implementation sequence. This review makes
no production, API, budget or active-roadmap changes.

## Product and architectural position

The coherent journey remains current telemetry → narrow a population → inspect
supporting signals → preserve or reuse. The record grammar, typed source policy,
fixed questions, composable presentation and independent evidence receipts already
cover much more than the original reservation example. Do not start another
observability engine, replace the planner with an evaluator, or make manifests a
prerequisite for ordinary OTel data.

352 connected typed discovery to wider service records and existing rolling
Activity views without confusing their windows. 353 now applies the same checked
policy to exact parent ownership and child membership, including raw local
admission at source/snapshot/capture boundaries. 354 isolates recovery of failed
widgets while retaining healthy evidence. These retire the corresponding 351
backlog, subject to 354's gates. They do not implement async topology, complete
traces, request-volume counts or aggregate capture without rereading.

## Recommended finite sequence

### Breadth: use authored views naturally from service evidence

A user looking at a useful service widget should be able to apply a team's saved
presentation to that exact population. `ui/view-template-handoff.mjs` already has
`authoredRequest`, `authoredResponse` and `expandAuthored`, with whole-question,
composition, timeline and explicit destination-baseline checks. The authored
library is connected in `ui/population-workspace.mjs`; service dashboard Customize
currently exposes only shipped HTTP-route and DB-operation presets through
`eligible`. This is a missing entry point, not a missing grammar or new registry.

Reuse the existing metadata catalog and checked expansion. Inspection and selection
read no telemetry; opening stages the destination's complete fixed question and
Run is explicit. Preserve typed environment, service/namespace, predicate,
relationship and measure profile; refuse incompatible signal/interval/comparison
requirements visibly. Do not inherit a saved template's source environment,
clock, service or evidence as destination authority. Keep unsupported native
composition explicit. Scope the first cut to existing representable record
widgets, with retired metadata callbacks and Back tested.

### Depth: current RPC semantics as an explicit observed profile

`exploration/dashboard.go` still detects legacy `rpc.system`/`rpc.method`, builds
that predicate and accepts legacy `rpc.server.duration` milliseconds. The UI
explicitly labels that legacy profile. Modern RPC telemetry must not silently
appear compatible through aliases or converted units.

First pin and research the relevant OTel convention revision. Then add only the
supported current profile with exact role/required fields, instrument/unit rules,
namespace/environment and independent record/native scopes. Preserve a truthful
legacy profile if still useful, or remove it deliberately; this prerelease has no
mandatory compatibility floor. Mixed current/legacy records must not be doubled
or merged by accidental fallback. Tests should prove raw membership, candidate vs
observed state, native independence and malformed/absent data. No protocol-success
or unique-request claim follows from recorded span Error.

### Consolidation: fixed questions lead with evidence, not the full editor

The 350 fresh package frames and `ui/exploration.mjs` expose an ordinary workflow
inconsistency: `openQuestion` always calls `setRecentMode(false)` and reveals the
full workbench, while the population canvas already has an evidence-first layout.
Consolidate fixed-question read versus edit presentation across package, saved,
reported-edge and context entries. Show the checked scope and explicit Run before
results; move optional filters/measures/shortcuts behind an intentional editor.
After Run, evidence should lead at fresh scroll zero on desktop and narrow screens.
Preserve controls and drafts rather than rebuilding them or losing unsupported
profile details.

This must remain a presentation change. `setRecentMode` also reparents controls
and its entry uses `recentQuestion` to resolve a fresh clock; it is not a safe
shortcut for a fixed question. Keep the existing question session and
`rootEditors`/population inertness and restoration ownership. Do not replace
specialized profiles with ordinary grammar, introduce a new workspace, advance
fixed bounds or read on editor disclosure. Acceptance should compare query/IO
ledgers, draft/error preservation, cancellation and retained Back as well as fresh
scroll-zero frames across representative fixed entry points.

### Conditional performance follow-up: legitimate large-facet presentation

343/346 fixed duplicate publication and local Find passes, but the admitted
2,000-value paired union still builds approximately 32,000 elements. This is a
remaining product-scale limit, not grounds for another setter optimization or an
assumed virtualizer. Start with a fresh styled desktop/narrow baseline of the
current legitimate single build, local search, keyboard selection and Back.
Separate source time, synchronous build cost, browser layout/paint and total wall
time. Historical repeated-build CPU is not the current single-build latency.

Only if that baseline demonstrates material cost, choose a bounded local
presentation page/window using the existing full immutable counted model. Search
must span all loaded values; exact identities, missing vs empty, baseline-only
values, counts, uncertain sides, staged selections and receipts remain unchanged.
Local paging/search/navigation must issue zero source reads and preserve meaningful
focus. Do not raise admission limits, truncate authority or select a general query
virtualizer in advance. If current cost is acceptable, close the measurement and
choose a newly evidenced consolidation instead of implementing machinery anyway.

## Layering, Go and reliability lessons

The recent work used existing lowering, source adapters and passive components;
keep that layering. The shared environment predicate now serves both sides of
reported-edge selection. Raw identity Unicode is checked before JSON repair.
Do not replace these explicit boundaries with a generic deep-clone or universal
scope manager. A bounded captured child page cannot re-prove parent ownership;
reopening preserves original source custody rather than doing an implicit lookup.

`service-dashboard.mjs` now owns a local two-read scheduler and presentation, while
Go owns profile planning/compiler semantics. There is real duplication between
Go profile predicates/native eligibility and JS profile presentation/positive
witness checks. During current RPC work, table-drive repeated internal Go choices
and separate pure JS profile helpers from acquisition only where it simplifies
that implementation. Avoid a new public descriptor protocol or generating the
same validation oracle on both sides: independent raw-witness admission is useful.

Go 1.27 is already declared and recent typed failures use `errors.AsType`. There
is no source evidence requiring a new generics layer or runtime upgrade project.
Keep cancellation, bounded clients, immutable snapshots and straightforward typed
APIs. The vulnerability-assessment access limitation remains an unresolved check,
not a clean bill of health. Preserve explicit byte/cardinality limits; 343's
comparison-body/label cross-layer budget is not solved by local facet rendering.

## Reconciled and deferred

Retire typed-exact-edge refusal and scalar-only continuation/package backlog.
Keep broad Discovery harness repair distinct from already-passing dedicated
journeys. Defer WYSIWYG/native composition, Databricks lowering, bitemporal storage,
multiplayer and broad platform dashboards until their own concrete journey and
source contracts justify them. Existing current/legacy/empty/unreported record
policies are not equivalent to unrestricted native-metric support. After the next
finite sequence, review the whole product again rather than extending a widget or
ownership micro-backlog indefinitely.
