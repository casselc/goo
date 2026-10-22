# Whole-product architecture after stream-first entry

Prepared 2026-10-22. **Source review and implementation recommendations, not a
completed refactor or a new runtime campaign.** This extends the
[stream-entry architecture](STREAM_FIRST_ENTRY_ARCHITECTURE.md),
[interaction design](STREAM_FIRST_ENTRY_DESIGN.md) and earlier
[architecture review](ARCHITECTURE_REFACTORING_REVIEW.md). It feeds the
[single active queue](IMPLEMENTATION_PLAN.md); it is not another parallel backlog.

The review covers `3fbefea` plus the visible, uncommitted S1 work in
`ui/stream-entry.js` and `ui/tail.js`. That work is in flight and has not been
independently validated here. In particular, a compact monitor wrapped around the
old editor is a migration seam, not the desired final architecture.

## Organizing principle

The common product flow is **live → pause → scope → inspect → investigate →
retain/reuse**. It should be available over generic OTel, service-specific views,
native instruments and declared operations without making users enter a new
workspace for every step. This is a workflow and ownership consolidation, not a
proposal to make all telemetry families share one evaluator or result schema.

Treat live, paused, fixed and frozen as explicit modes of a visit/context. Give
one application owner responsibility for that visit's active requests,
subscriptions, staged/applied scope, retained selection, inspector and return
anchor. A child investigation can have its own owner while its parent remains
paused. Comparison sides retain independent owners/read bases. Do not replace
those owners with one mutable product-wide “current result.”

A small context **reference** can carry mode, typed family, applied scope reference,
read state and optional custody reference. It must point to the family's exact
question/result rather than normalize record pages, metric intervals and
operation contracts into an arbitrary universal object. Native exact nanosecond
bounds, temporality, grain and evaluation clocks remain in their typed contracts.

## Surface-by-surface ownership

| Surface and source | Desired role in the common flow | Concrete consolidation |
| --- | --- | --- |
| Fresh Explore: `stream-entry.js`, `tail.js` | Immediate bounded live observations, then pause/inspect/narrow. | One fresh-entry owner and one active monitor. Do not require the population planner first or let the legacy editor and stream both own execution controls. |
| Fixed/population exploration: `exploration.js`, `population-workspace.js`, `investigation-trail.js` | A scoped continuation of an observation or selected interval; retained parents make branching useful. | Share record presentation/inspection and applied-scope presentation. Keep fixed reads and rolling selectors explicitly different; a stored link remains staged/source-free. |
| Services/distributed services: `inventory.js`, `service-browser.js`, `distributed-services.js`, `service-dashboard.js` | An alternate way to choose context and inspect useful facets/measures. | Feed checked service identities and fixed scopes into the same typed continuations. Adopt the shared record body in one widget before broad replacement. Keep overview planning and independent widget execution out of passive displays. |
| Native instruments: `measurements.js`, `window-discovery.js`, `histograms.js`, `ratios.js`, `fleet.js` | Discover supported instruments, inspect exact streams/windows and compare meaningful measures. | Reuse shell, controls, inspector, status and library rows. Retain native-specific renderers/coverage/units; no browser-buffer percentile, cross-stream rate or fleet averaging shortcut. |
| Application Operations: `operations.js`, operation orchestration in `index.html` | Declared application-contract evidence and contextual navigation, with optional Go weaving. | Keep the delivered live/saved ownership split and `GooOperations.overview`. Move orchestration into a named app adapter as touched; do not reintroduce editor reparenting or make a manifest prerequisite for Explore. |
| Saved investigations: `captures.js`, `documents.js`, `investigations.js`, `retention.js` | Preserved context, exact evidence and narrative, with explicit continuation into live/fixed work. | A selected-item presentation with typed artifact adapters. Saved records should use a supplied-data record presenter, not a complete read-only query editor. Preserve old deep links and separate store authority. |
| Definitions/packages: `telemetry-packages.js`, measurement catalogs, exploration question library | Reusable typed intent, bindings and authored configuration arising from useful observations. | One compact revision-aware catalog-row composition; selected detail/authoring only when needed. Definitions are not telemetry snapshots. Keep code-managed and interactive publication paths aligned with the same typed plans. |
| Examples and instrumentation/test tools | Demonstrate the same real workflow and help authors supply ordinary OTel. | Use examples as clients of shared flows, not as owners of extra product controls. Keep specialized lab compatibility until parity/removal is explicit. Annotations/weaving remain optional for non-Go telemetry. |

Services and Definitions can remain useful navigation destinations. The
consolidation goal is that they supply context or reusable intent to the same
investigation flow, rather than recreate its entire toolbar, inspector and capture
workflow. No source review justifies deleting a useful destination merely to make
the navigation shorter.

## Extract composition, not another enclosing workspace

The strongest existing seam is application presenters above passive elements.
`GooDistributions.mountFacets` and `GooOperations.overview` demonstrate that a
supplied-data body can be shared without nesting a complete workflow.

Use that pattern for these extractions, in this order:

1. **Record body and selected-record inspector.** `exploration.js` currently owns
   its table construction, attributes, selection, trace/log continuation and
   read-only capture mode. `tail.js` constructs a separate stream table and
   attribute disclosure. Extract a compact supplied-record row/list presenter and
   selected-record detail presenter. Accept explicit columns/pills, UTC display
   facts, opaque observation handles and caller-owned actions. Keep rolling
   insertion/eviction, source generation and server custody outside the element.
2. **Applied-scope / result-state / continuation composition.** Extend the existing
   `evidence-scope.js` adapters for records and paused observations rather than
   building yet another family of scope paragraphs. One compact region should say
   what is applied, whether new input is staged, and which result is stale,
   partial, unavailable or frozen. Exact details/caveats can disclose locally;
   decisive gaps and changed-read warnings remain visible.
3. **Interaction controls.** Share action layout and presentation for selection,
   scope edits and comparison sides using existing `GooUI.actions`, inspectors and
   result-state elements. The app owner supplies availability and handlers. A
   control component does not decide whether capture rereads, whether a field
   predicate applies to a companion signal, or whether a frozen result is durable.
4. **Typed detail/comparison bodies.** Adopt existing chart-frame, evidence-summary,
   comparison visual and inspector elements across real consumers. Share the
   two-sided visual structure before trying to unify stored comparison,
   independent reread and rolling comparison execution.

`captures.js` currently mounts `GooExploration` with `readOnly:true` for saved
records. That builds query/editor/context machinery and suppresses parts of it.
After the record presenter exists, replace this with a saved-record adapter that
provides stored pages and named fresh-context actions. The live editor and saved
viewer should share a body; the saved viewer should not become a disabled live
editor. The same rule applies to service widgets and future annotation-review
surfaces.

The new S1 `stream-entry.js` visibly moves the old Explore subtree into
**Questions, charts and saved views**. That is a useful bounded compatibility
seam while initial behavior changes, but it requires a removal gate: once the
shared record body, scope/continuations and explicit fixed-question destination
have parity, stop mounting the complete editor inside every live entry. Do not
add a third workspace wrapper when S2 selection, ancestry or S3 retention arrives.

## One vocabulary, typed effects

The product needs fewer overlapping controls, not fewer distinctions in evidence.
Use the same action words when their effects match:

| Intent | Presentation contract | Effect owner |
| --- | --- | --- |
| Live / Pause / Resume | Current observation mode and pause reason; resume opens a new recent window with gap semantics. | The active visit's subscription owner. |
| Edit / Apply scope | Input is staged until applied; old rows never appear under new applied pills. | Typed scope/execution owner, including generation invalidation. |
| Inspect | Retained supplied observation/interval, stable focus and anchor. | Local selected-item owner; no hidden source read. |
| Expand trace / surrounding context | Clearly named fresh read, preserving original anchor/read window and independent basis. | Context-read adapter. |
| Open fixed question / Rerun fixed window | Stage a checked exact question, or explicitly execute it. | Family-specific query adapter. |
| Preserve selected observations | Exact supplied membership is copied through a server-owned custody contract. | Typed retention application service; disabled/unavailable until supported. |
| Capture a new fixed read | Source reread may differ from the observed display. | Capture application service, with its own resulting receipt. |
| Save definition / save view | Reusable intent/presentation configuration, not evidence. | Definition/view revision owner. |
| Add to investigation / save revision | Stage references/context, then publish authored immutable narrative. | Document draft/revision owner with dirty/pending guards. |

Do not consolidate all three save-like operations into an ambiguous **Save**
button. Their visual placement can be shared while their explanatory labels,
availability and typed handlers remain different.

A shared field/presence display helper should back pills, columns, attributes,
facets and definitions. Display labels must distinguish resource/record fields,
missing values, present empty strings and opaque literal values. Reuse existing
Go-checked scope transformations in `question-scope.js` and `exploration/scope.go`.
A string label or service-name pill must not become the source of query truth.

## Lifecycle and continuation seams

The current code has independent generation/controller patterns in stream,
population, service, measurement, documents and catalog adapters. `GooQueries.post`
already supports request cancellation; `GooQuestionSession.Session` already models
record execution. Neither is a universal lifecycle for all product effects.

Compare two real consumers before extracting a tiny operation owner with
`begin/invalidate/current/dispose`. It can centralize cancellation/late-completion
mechanics, but must not auto-retry writes, reconnect streams, choose a scope or
publish results. Stream pause/resume, dirty document editing and one-shot widget
reads retain distinct state transitions and test models.

Replace scattered global event payload assumptions incrementally with named typed
continuation builders: record → trace context, service → fixed population,
instrument → exact stream question, selected observation → retention request,
evidence → narrative reference. The routing adapter interprets these intents;
passive components emit them. Preserve `#step`, `#question`, capture, document and
native evidence links, and keep absent optional/default fields compatible.

Navigation must classify intent before starting work. Fresh Explore may auto-start
one monitor; opening a saved artifact or shared fixed question must not. A hidden
retained visit must own no active reader, timer or capacity slot. Multiple DOM
panels or event listeners must not each infer that they are the active entry.
User/inspection pause outranks automatic visibility changes; scroll-away must not
silently evict the row being read. The initial implementation can pause collection
on scroll-away; a separately bounded pending lane is later evidence-driven work.

## Storage, backend and semantic boundaries

Stream-first raises the priority of exact selected-observation custody but does
not justify a new general database layer. The existing tail does not issue result
page tokens, and its public snapshot omits per-row read windows. Introduce/version
the exact membership contract before offering freeze/tag/attach from the stream.
Do not manufacture a `Result` with invented totals/basis to fit the current native
retention API. Reuse native idempotency/recovery only after the new intent is
server-supported. File captures, process-local capabilities, native retention and
narrative revisions remain separate typed authorities.

Current-head catalogs, pinned review revisions, event time, source-read time,
custody time, evaluation time and await-token visibility lower bounds answer
different questions. A compact result shell can disclose them consistently but
must not collapse them into “current” or one timestamp. Loading an existing file
document should continue to work when optional XTDB is unavailable; referenced
native content can be explicitly unavailable.

The previous backend extractions still make sense after UI effect boundaries are
clear:

- Extract immutable-file publication mechanics only after byte-identical hash and
  lineage fixtures. Domain validation/digest/catalog rules remain in each store.
- Extract bounded ClickHouse HTTP transport one source family at a time. Preserve
  typed parameters, limits, cancellation, decoding and family basis construction.
- Assemble explicit server dependencies and an endpoint-effect inventory.
  `query_budget.go` currently controls selected POST routes; operations GET reads,
  comparison/capture effects and SSE have different paths. Admission should be
  reviewed by actual effect, not by HTTP verb alone. SSE retains its independent
  lifetime and disconnect semantics instead of inheriting ordinary query timeouts.
- Preserve typed kernels: `exploration`, native `metrics`, `measurements`, operation
  `annotations`, `lifecycle` and prepared-fact `measurement` are not interchangeable
  evaluators. Databricks/ClickHouse lowering remains explicit. DataFusion, CEL/KCL,
  a framework migration, proxy removal or a broad XTDB migration is not a
  prerequisite for this product flow.

## Migration and deletion gates in the existing queue

| Existing cut | Bounded adoption | Remove only after parity |
| --- | --- | --- |
| S1 | Fresh-entry subscription owner and compact stream; preserve fixed/saved links and the current editor as explicit compatibility detail. | Remove redundant stream headings/toolbars from the compact composition. Do not delete the legacy editor yet or claim the temporary wrapper is the final layout. |
| S2 / A1–A3 | Shared record body/inspector, applied scope/action region, ambiguity-aware ancestry and one service-record widget adoption. | Remove duplicated tail/exploration record construction; stop mounting a full read-only query editor for saved records; delete only superseded hiding selectors and nested chrome after tests show parity. |
| S3 / F1 | Exact selected-tail membership and supported durable retention, with progressive preserve/tag/narrative actions. | Remove reread substitutes or ambiguous capture labels; preserve explicit new-read capture as a separate valid action. Never remove a durable path merely because a new optional store exists. |
| B1–B2 | Compact typed catalog rows, selected detail/review, protocol/field labels and narrow request-owner adoption. | Remove duplicated catalog card layouts, eager review panels and obsolete serializers after fixtures pass; equal titles and parallel heads remain distinct. |
| C1–C2 | Hash-preserving publication/transport helpers and effect-aware dependency/admission wiring. | Delete copied mechanics one family at a time only after compatibility/source/cancellation tests pass. No service split or universal backend rewrite. |
| D / E | Common comparison presentation and checked scope/isolation/capability continuations arising from the stream/service flow. | Retire duplicate analyzers only after their supported cases exist in shared continuations; keep explicitly unsupported native/OTel cases honest. |

This order folds the earlier architecture findings into the active delivery plan.
It does not require completing all backend cleanup before users receive the new
entry. Each extraction needs two named consumers or a concrete duplication removal,
not merely a generic interface with speculative users.

## Verification and documentation ownership

Test the common transitions and typed boundaries across product surfaces, then
keep example journeys as concrete clients:

- A shared navigation/visit contract test covers fresh auto-start, fixed/saved
  source-free restore, pause/leave/hide/dispose, dirty drafts and retained-parent
  focus/scroll. Do not delete old no-hidden-read checks; make their intent/mode
  explicit rather than exempting all startup requests.
- Presenter fixtures and browser component pages cover supplied/SSR data,
  pre-definition property upgrades, detached/remounted SPA data, keyboard/touch,
  unknown/partial/invalid states and large exact timestamps. New shared display
  components never fetch, subscribe or mutate artifacts.
- Hegel state/swarm tests operate on the real owners/protocol helpers; pure graph
  properties cover parent ambiguity/cycles. Quint observed traces and rejecting
  controls cover cross-surface read/publication ownership, stale completion,
  resume gaps and exact-member preservation, not only one happy storyboard.
- Playwright/Bombadil exercise the stream first, then pause → field/service
  refinement → fixed/trace/native context → retained evidence → narrative/reuse.
  Test service and manifest-free/non-Go data as well as examples. Full-population
  statistics and native semantics remain independent of buffer limits.
- Each extraction updates its owner diagram, affected storyboard/screenshots,
  shared component documentation and active queue status. Historical reviews
  remain dated evidence; the validation record states what actually ran.

This review ran source inspection only. It makes no new browser, source-load,
benchmark, security, custody or production readiness claim.
