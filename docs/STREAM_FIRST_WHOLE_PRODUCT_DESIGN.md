# Whole-product design after the stream-first decision

Prepared 2026-10-22. **Source-backed design recommendations, not implemented
capabilities or a new browser usability study.** The user's stream-first direction
changes the product's entry model, not only the order of panels in Explore. This
review considers Explore, Services, Operations, Definitions, Saved investigations,
Examples, comparisons and annotations together. It complements the
[entry design](STREAM_FIRST_ENTRY_DESIGN.md),
[architecture review](STREAM_FIRST_ENTRY_ARCHITECTURE.md) and
[single implementation queue](IMPLEMENTATION_PLAN.md).

At this review's source snapshot, `3fbefea` delivered independent live/saved
operation presentation and compact applied population composition. The new
`ui/stream-entry.js` and compact tail integration are work in progress in the
shared tree; this document does not declare that S1 has shipped. Use
[VALIDATION.md](../VALIDATION.md) for actual release status. Earlier reviews remain
historical observations, not claims that their now-delivered items are missing.

## One workflow spine, several useful entrances

**Observe → narrow → inspect → investigate → explain/preserve → reuse.**

Live Activity is the natural first Observe surface. A service, a saved finding, a
native instrument or a declared operation can enter the same spine at another
point. They should bring a scoped intent into common investigation interactions,
not bring an entire nested workspace with a second scope bar, query button,
selection history, capture form and save library.

Preserve the useful distinction between live monitoring, fixed-window analysis,
stored evidence and reusable definitions. Simplification means fewer competing
controls and clearer transitions between these modes; it does not mean every
signal, record bank and artifact becomes one mutable generic object.

| User is trying to… | Primary surface and visible content | Natural next action |
| --- | --- | --- |
| See what is happening now | Explore / Activity: stream, applied scope, compact stats and live/paused state. | Narrow with row/field pills; inspect a retained row. |
| Understand a service | Services: compact inventory then a focused service overview. | Open scoped Activity, a fixed record question or a supported native measurement. |
| Understand declared behavior | Service's declared operations or compatible Operations route: contracts versus observations. | Inspect operation evidence, contextual spans or a typed outcome/coverage view. |
| Isolate an interesting population | Fixed investigation canvas: one applied scope, charts/facets/records and selection tray. | Compare, expand context, preserve supported evidence or record a finding. |
| Return to prior work | Saved investigations: a typed library and one selected investigation/evidence viewer. | Inspect frozen evidence, continue explicitly or compare a new read. |
| Make understanding reusable | Definitions: compact revision-aware library and a selected editor. | Review, preview explicitly, use, export or create an instrumentation handoff. |
| Learn the system | Examples/help: guided entry into the same real surfaces. | Follow a real question; return with context instead of entering another product. |

The current main destinations can remain while these compositions are adopted.
Consider moving **Operations** from the primary navigation into service/definition
context only after parity exists for cross-service declared packages, direct
links, sample workflows and expert users. Keep `#operations` as a compatibility
route. Do not hide native metrics behind Go instrumentation, and do not turn
Examples into a production navigation dependency.

## Whole-product control hierarchy

Use one small context region per active owner:

1. **Where/what:** applied source, environment, service/namespace or selected
   instrument identity, plus visible staged scope changes.
2. **When/mode:** Live or Paused; a fixed UTC window; or Saved evidence with its
   source/capture clocks. These are mutually understandable states, not competing
   date inputs scattered above and below the result.
3. **Primary continuation:** Pause/Resume for Activity; Run/Apply for a staged fixed
   question; Inspect/Continue for stored evidence; Validate/Use for a definition.
4. **Contextual tools:** filters/fields, selected-item actions, comparison,
   preserve/add finding and advanced details appear when relevant.

A whole-product scope bar is a **presentation contract**, not a global mutable
scope store. Opening a service or saved receipt must not overwrite an unrelated
live draft. Applied and staged scopes remain distinct. Editing a new-query
lookback never changes an already applied fixed interval or the monitor's actual
60-second polling horizon.

The current global `#minutes` control in `ui/index.html` is a new-query lookback,
while `ui/stream-entry.js` has its own live environment/signal and `ui/tail.js`
uses a rolling 60 seconds. Do not show the lookback as if it governs live rows.
Contextualize it under fixed analysis, or label the different windows directly.
Likewise, the existing outer Captured records/Measurements tabs and inner stream
signal picker should not become three persistent layers of choices. Activity
owns Logs/Spans; native measurements remain an explicit analysis entrance with
only their supported controls, rather than a fake live mode for every family.

## Code-backed consolidation priorities

| Priority | Existing source / friction | Proposed consolidation or removal | Acceptance before removing the old path |
| --- | --- | --- | --- |
| First | `stream-entry.js` wraps the legacy exploration in a disclosure; `index.html` still supplies global lookback and outer modes; `tail.js` owns transport and its former full control panel. | Let the root monitor own one scope/live toolbar and a compact record viewport. Keep advanced questions as an explicit continuation. Remove duplicate scope paragraphs/transport controls from the adopted compact composition; preserve normal full-tail callers. | Real rows visible initially on desktop/mobile; one auto subscription only at true fresh entry; saved routes inert; no contradictory lookback; sticky pause and no anchor eviction while inspecting. No second full workspace merely stacked underneath. |
| First | `exploration.js` renders record rows, attributes, selection, context and capture controls; `tail.js` renders another table and `stream-entry.js` creates an inspector. | Extract supplied-record **content** and bounded continuation slots, then adopt them in fixed pages and live rows. Keep row-bank selection and read-window provenance in each owner. Use the same inspector interaction model. | Both real consumers preserve UTC, severity/status/duration, missing values, exact IDs, focus and compact mobile rows. Rendering/disclosure does not fetch; tail keys never masquerade as page tokens. Delete duplicated content formatting only after both use the extraction. |
| First | `service-browser.js:open` composes service discovery, an aggregate span read, native discovery, operation inventory, a dashboard and three supporting sections. It has its own return bar and follow links. | Service detail becomes one scoped overview with Activity / Analyze continuations and progressive supporting inventory. Reuse the scope/evidence header and selected-detail return treatment. Keep the existing typed dashboard plan; do not add a second live service page. | One service/namespace identity survives every continuation; unavailable native/discovery reads remain explicit; supporting catalogs do not precede the useful overview; returning preserves service detail and does not rerun it. |
| First | `index.html` Operations and `operations.js:overview` now have separate live and saved owners. Their typed rates, ratios, histograms and fleet views are valuable but operation-specific. | Preserve the new ownership separation. Present declared behavior from service/definition context, with operation selection opening the shared inspector/fixed investigation. Contextualize sample links and binding details. | Invocation counts are not business outcomes; counters and span context remain separate; old routes, package selection and saved captures work. Do not fold ratio/coverage semantics into generic stream chips. |
| First | `documents.js:questionEvidence/evidenceSummary`, `captures.js:show`, retained citations and saved operations construct related scope/status text differently. | Extend `GooEvidenceScope` with bounded record/capture transforms and use one compact typed evidence header. Saved contains one selected viewer; browse sections and raw-ID recovery are secondary. | Pinned old reviews, exact native bounds, read/capture clocks, unavailable stores and changed-read caveats remain visible. Opening an artifact never executes the root monitor. Keep dirty notes and local selection intact during navigation. |
| Next | `telemetry-packages.js:refresh`, measurement catalogs, Explore reusable questions, documents, capture and retained catalogs use different full cards and search/action wording. Telemetry package cards eagerly mount meaning panels. | Share compact revision-aware rows and selected-detail disclosure. Definitions owns authoring; contextual pickers offer references, not duplicate full editors. Use semantic grouping only when a real family identity exists. | Same-title unrelated items remain distinct; parallel heads remain accessible; pagination/search are clear; review details mount for selected entries; selecting is not execution. Keep old artifact IDs and recovery by ID. |
| Next | `distributions.js`, `evidence-comparison.js`, `fresh-comparison.js` and `live-comparison.js` each assemble two-side controls and result presentation. | One comparison **display** shell: named baseline/current sides, scope/compatibility, shared-domain visuals, exact table fallback and a contextual finding action. Offer Compare from a population, pinned baseline or selected saved evidence. | Stored comparisons remain source-free; fresh comparisons name the new read; rolling comparisons name their clock/window; incompatible units/support stay refused. Do not add a permanent Compare destination or merge execution owners. |
| Next | `meaning.js` exposes similar tags/notes on many targets; retained reviews and pending-save recovery add different custody behaviors. | Share the compact authored-meaning entry and selected-target summary. Put service context on the service, definition notes on the definition, findings on evidence/investigations. Surface pending count and recovery contextually in Saved. | Tags/notes do not become copied telemetry attributes, inferred classifier labels or silent cross-target edits. Revision parentage, applicability time and exact retained targets remain explicit. Native-only retry promises remain native-only. |
| Next | `example-questions.js`, `guided-measurements.js`, legacy examples in `index.html` and package-specific sample links provide overlapping entrances. | Examples become a small set of real-user journeys that enter Activity/service/analysis with visible scoped intent. Keep the reservation lab clearly marked as compatibility; move implementation-oriented sample launch links to example/developer context. | Generic OTel and non-Go journeys remain first-class; example opening never creates business traffic or publishes evidence without its named action; existing executable documentation and compatibility links survive. |

## Shared primitives to adopt, not another component framework

| Shared piece | Display responsibility | Application responsibility |
| --- | --- | --- |
| Scope/mode header | Applied label/pills, one mode/window badge, staged-change indicator, compact contextual action slots. | Resolve scope, validate predicates, choose clocks, execute/subscribe, retain prior owner. |
| Record content/context | UTC time, service, severity/status/duration, summary, selected field pills and honest context affordance. | Own observation bank/key, row membership, selection, source projections and explicit context reads. |
| Selection action strip | Selected count, compact available actions, busy/unavailable state. | Compute valid shared fields, freeze/pause, derive a question or acquire supported exact custody. |
| Inspector | Identity/content/actions, keyboard dismissal, opener and scroll return. | Load context if explicitly requested, pin read provenance, prevent stale completion. |
| Evidence header | Typed supplied facts about source, scope, clocks, coverage and custody. | Parse/resolve the specific receipt and await basis; never invent a universal evidence identity. |
| Catalog row | Name, kind/scope, revision/history/parallel-head cues, action slots. | Family grouping, authorization, pagination, selected revision and publication. |
| Comparison shell | Named sides, compatibility state, common visual domain and exact values. | Stored/fresh/rolling execution, supported semantics, baseline pinning and save/citation. |
| Authored-meaning entry | Compact tag/review badges and selected revision disclosure. | Target identity, applicability, parent heads, retained provenance and retry intent. |

Use existing `goo-table`, timestamps, attributes, facets, `GooUI.inspector`,
`goo-evidence-summary`, `GooEvidenceScope`, charts and comparison geometry. Add
small supplied-data compositions when two concrete consumers agree. Avoid a
single “universal observability card,” a giant option matrix, a second SSE parser,
or new custom elements that fetch, route, calculate population measures or save.
Server HTML, pre-definition property updates and SPA reconnect must remain valid.

## What must remain distinct

- Live buffer observations, fixed result pages and immutable captured records
  have different membership and lifetime. A paused tail row is not yet durable
  evidence, and a later window reread may not contain it.
- A definition, an evidence capture, an investigation assertion and an annotation
  revision are different durable objects even when one catalog browses them.
- Logs, spans and native metrics share display vocabulary, not arithmetic. Span
  duration summaries, cumulative/delta counters, native histograms and resource
  coverage keep their typed Go semantics.
- Reported parent edges, OTel links, business-subject identity and temporal context
  do not collapse into one ancestry chain. Unknown remains unknown.
- Live/fixed/saved execution owners and their independent clocks remain separate.
  The root monitor's intentional auto-connect must not leak into saved reopening.
- Personal local recovery and immutable files do not imply shared-team
  authorization, multi-client merge, complete replay or a training-data platform.

## Migration through the existing queue

1. Finish the already-started root-monitor cut with compact controls, lifecycle
   and real first-viewport evidence. Keep fixed analysis reachable and all old
   links valid. Record supported versus deferred stream custody clearly.
2. Extract/adopt row content, contextual actions and scope/evidence header in the
   stream and fixed pages, then one service surface. This joins S2/A3; it is not a
   separate component initiative. Replace shared content only after parity.
3. Consolidate catalogs and selected viewers under B1, then align comparison and
   annotation presentation under their existing B/D/F cuts. Remove old markup,
   duplicate controls and CSS overrides once their last named consumer migrates.
4. Reassess primary Operations navigation only when service/definition-context
   entry parity exists. Revisit broader nav after actual use, not by renaming all
   destinations in the first stream change.

The [active queue](IMPLEMENTATION_PLAN.md) owns task identifiers and sequencing.
These migration steps describe adoption, not an additional ordered backlog.

## Acceptance across the whole product

- A real user can arrive at live Activity, narrow to a service, pause/inspect,
  enter a fixed question, capture supported evidence, add a finding, and reuse a
  definition without meeting another full editor at every step.
- A service-first user and a saved-investigation-first user can enter that same
  loop with the correct original service/scope or exact frozen evidence. Returning
  preserves their parent; unrelated live drafts and unsaved notes stay unchanged.
- Every active surface has one applied scope, one mode/window treatment and one
  primary continuation. Other actions are contextual and keyboard reachable;
  dense mobile rows and first useful content are measured, not only screenshots
  of isolated widgets.
- Local pill staging, catalog search, disclosure and inspection make no hidden
  source/publication calls. New reads are named, generation-owned and rejected
  after context replacement. Saved routes remain source-inert.
- Playwright uses real OTel for stream-first, service-first and saved-first
  storyboards, including partial/missing signals and non-Go/custom conventions.
  Refresh affected screenshots and architecture/process diagrams after adoption.
- Bombadil actions must traverse the actual transitions, not only wait at one
  screen: pause, narrow, inspect, parent return, select a revision and explicit
  comparison. Invariants protect one owner, visible scope/mode, bounded buffers,
  no stale publication and stable selected content.
- Hegel state/swarm and observed Quint traces remain focused on changed protocols
  and interleavings: subscription generations, paused parent restoration, exact
  membership/custody, revision identity and rejecting controls. Layout CSS needs
  browser evidence rather than mirrored semantic unit tests.

This pass reviewed current source and previous findings. It did not run a fresh
browser campaign, validate the in-flight root monitor, benchmark query budgets or
claim production authorization/custody support.
