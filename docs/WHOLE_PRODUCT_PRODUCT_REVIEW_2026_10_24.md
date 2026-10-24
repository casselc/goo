# Whole-product product review — 2026-10-24

Status: source-backed product review of baseline `86e35a5`, prepared for the revised roadmap. Recommendations are inputs to the single active implementation queue, not a second backlog. This review changes no application behavior. It examines the current owners, README and earlier reviews; it is not new customer research or a fresh browser accessibility audit.

## Product judgment

Goo's useful distinction is a continuous path from live telemetry to a scoped explanation that another person can inspect and reuse. The initial stream, population tools, ordinary OTel discovery, explicit fixed reads and retained evidence already support this path. The next product improvement should make its last steps dependable: show whether a selected observation survived a later evidence read, and keep a person's unfinished explanation recoverable. Another dashboard, picker or permanent workbench would contribute less.

Keep the entrance simple: show relevant recent logs/spans, make narrowing and inspection immediate, and reveal question authoring and preservation when the user acts. Keep services, native measurements and declared operations as scoped routes into that same investigation flow. Preserve the important distinction between an observation, a later read, a captured result, a reusable definition and a saved explanation; these distinctions need concise presentation rather than removal.

## What is already delivered

The following are existing capabilities, not requests for fresh implementations:

| Product task | Current evidence | Remaining boundary |
| --- | --- | --- |
| See current activity and inspect a row | `ui/tail.js`, `docs/ACTIVITY_ANALYSIS.md` | Live observation handles are not durable event identities. Exact live-selection publication has a separate custody prerequisite. |
| Analyze a fixed population, narrow, widen and compare | `docs/D2_ACCEPTANCE.md`, `docs/POPULATION_WORKSPACE.md` | General nested predicate editing and arbitrary numeric attribute semantics are not implied. |
| Compare visually with useful error/latency ordering | `docs/D1_ACCEPTANCE.md`, `docs/COMPARISON_SERIES.md` | Different units, independent clocks and unsupported native combinations remain explicit. |
| Customize supplied record fields and grouping | `docs/RECORD_COLUMNS.md`, `ui/grouping-control.js` | This is bounded supplied-page discovery, not a global schema or arbitrary transform engine. |
| Discover ordinary OTel services and measurements | `ui/service-browser.js`, `ui/measurements.js`, `docs/E1_ACCEPTANCE.md` | Service suggestions already include HTTP, database and messaging. Matching observed attributes is not verified instrumentation completeness. |
| Use service overviews and supporting records | `ui/service-browser.js`, `docs/CORE_EXPERIENCE_REFINEMENTS.md` | Supporting records and overview aggregates can have different read bases. Existing service/instance navigation should not be rescheduled as absent. |
| Browse investigations and evidence without entering an editor | `ui/saved-workspace.js`, `docs/SAVED_WORKSPACE.md` | Cross-store global search and durable ordinary drafts are not delivered by this composition. |
| Save questions, notes and exact evidence references | `ui/documents.js`, `docs/INVESTIGATION_DOCUMENTS.md` | Ordinary unsaved drafts remain page-local. Native retry recovery is a different, already implemented workflow. |
| Discover reusable definitions without opening every editor | `docs/DEFINITION_LIBRARY.md` | Broader blank-authoring or dashboard-layout products require a concrete task. |

Earlier reviews correctly diagnosed fragmented controls, libraries and ownership. Several of their proposed remedies have since shipped. In particular, do not revive the Oct22 review as a request to rebuild Activity analysis, Saved task composition, native discovery or population operations.

## Ranked findings and bounded acceptance

### 1. Make trace findings disclose anchor continuity

**Priority: next product cut after the in-progress measure-authoring consistency work.**

A user selects an interesting span, expands its trace and preserves a finding. The current finding workflow performs two later explicit captures and records that anchor retention and trace completeness remain unverified (`ui/trace-journey.js`, especially `findingCards` and the finding capture owner). This is honest but leaves the central user question unresolved: does the preserved material still contain the observation that motivated the investigation?

Add a bounded continuity diagnostic for the selected anchor against each captured record page. Report **present**, **not found in the retained page**, or **unavailable**, with the exact matching basis and bounded-page caveat. A reported trace/span identifier is not a globally unique event occurrence. A matching identifier must not establish byte equality, unique occurrence, whole-trace completeness or absence from the underlying source.

Acceptance:

- The staged finding preserves the original anchor reference and independently names both later capture results and their read bases.
- Positive, absent-from-page, unavailable and ambiguous/duplicate cases cannot become an unqualified “same evidence” claim.
- No extra telemetry read is hidden behind displaying the diagnostic; capture remains explicit.
- Existing older findings remain readable without invented continuity data.
- Pure matching fixtures cover missing/empty identifiers, duplicate rows and changed attributes; a real journey captures and reopens the finding. Changed asynchronous staging gets meaningful state/trace checks.

Dependency: use actual stored capture payloads and existing source/capture owners. Do not wait for XTDB, a durable live cursor or a new evidence store. Do not conflate this diagnostic with exact live selection custody.

### 2. Recover ordinary investigation drafts locally

**Priority: following continuity.**

An explanation may contain multiple questions, evidence links and edited prose. Today `ui/documents.js` tells users that the unsaved draft is retained in this page; reload discards it. Native publication recovery already exists, so this gap is specific to ordinary narrative work rather than a request to merge recovery stores.

Provide one bounded browser-local recovery slot or an equivalently explicit small bounded journal for the ordinary draft. Offer restore/discard; keep the base revision and referenced IDs intact. Restoration is local work, never an automatic save, publication, telemetry execution or replay of an uncertain request.

Acceptance:

- Reload can recover edited title/question/finding/open questions, staged cards and an in-progress note edit according to the documented scope.
- Recovery is clearly unsaved and browser-local; storage denial, malformed/old data and quota failure leave the current editor usable.
- Opening another revision cannot silently overwrite either the active dirty draft or the recoverable draft.
- Explicit save, revert and discard have documented cleanup behavior. Failed or uncertain saves do not acquire a fabricated successful receipt.
- Test reload, stale base revision, failed save, discard, storage failure and interleavings with incoming evidence. Restore performs no source read or POST.

Dependency: the document owner retains validation, publication and draft-generation authority. A small passive recovery notice may be shared later; a universal workbench/store is unnecessary. Retained telemetry references need not be copied into the recovery slot.

### 3. Replace historical orientation with an accurate current entrance

**Priority: small documentation cut alongside the first two, not an application redesign.**

The README gives the correct product promise but follows it with a long chronological chain of implementation links. The consolidated proposal acknowledges delivered milestones while later historical tables still say “Next” for work now delivered. The implementation index similarly mixes current intent with many detailed historical slices. This makes it harder to answer “what should I use?” and “what is actually next?” than to find implementation detail.

Acceptance:

- README opens with a short task sequence: start local lab → inspect current activity → narrow/analyze → preserve/reuse; then links to architecture, capabilities and contributing/validation.
- One authoritative finite queue marks delivered, active, blocked and deferred work. Historical reviews retain dates and explicitly defer to that queue.
- Keep capability limits and source/coverage caveats accessible; move chronology rather than deleting evidence.
- Each active product cut names a concrete user outcome and acceptance, rather than a new component name.

Dependency: planning agent owns queue reconciliation; avoid concurrent competing plan edits.

### 4. Make the next general-observability expansion evidence-driven

**Priority: subsequent proposal, not another overnight UI increment.**

The core can already handle ordinary non-Go OTel, service dashboards, HTTP/database/messaging suggestions and native discovery. What remains weaker is explaining data availability and pipeline loss. Existing caveats correctly avoid inferring collector health from an empty result, but repeating “unknown” does not supply operational evidence.

A useful next proposal is a bounded, explicit collector/exporter diagnostic read based on an actual available metric source. Identify supported queue, rejected/dropped and export failure counters by exact instrument identity and metadata; show their observation window and missing support. Do not use record age as proof of collection health, infer sampling completeness, or paint missing metrics green. Inspect the local collector configuration before authorizing implementation: if the source is unavailable, produce the contract and fixture-backed plan rather than a decorative healthy dashboard.

Acceptance for that future cut: one real local pipeline signal source, exact metadata/unit/temporality, independent clock and scope display, known-to-unavailable transitions, and a scenario with an actual failure signal. This can reuse existing native measurement and service navigation; no new Home dashboard is needed.

## Product-wide consistency rules for the revised queue

- Keep inspection, disclosure and draft restoration source-free. Label explicit Run, Capture, Save revision and Resume by their actual effects.
- Prioritize useful compact displays and selected detail over adding global control rows. Keep mobile/keyboard equivalents for chart gestures and return focus.
- Use UTC timestamps with exact raw values available on inspection. Preserve units, typed integer precision, limits and unknown coverage.
- Treat OTel conventions as capability hints with qualified keys, current/legacy distinctions and generic fallback. Never silently rewrite arbitrary telemetry or equate span status with business outcome.
- Reuse passive display components and typed owner intents. The shared grouping/record/detail/comparison components are progress; avoid a single result schema that erases native/record/live differences.
- Definitions describe reusable questions; captures preserve a particular read; investigations preserve explanation and references. Local recovery is none of these publication contracts.
- A screenshot or storyboard is evidence of the exercised scenario, not exhaustive correctness. Every changed journey should refresh the relevant documentation and bounded Bombadil actions/invariants; state/property/model tests should target real identity, generation and effect risks.

## Deliberate deferrals and consolidation

Do not add a generic dashboard builder, multiplayer canvas, classifier suite, new query engine, required bitemporal backend or more sample-only panels to this finite queue. These can be valuable later, but do not solve the next evidence-continuity and unfinished-work problems. Keep XTDB optional; no current finding requires making it the telemetry store. Keep external-platform translation fixture-driven and explicit about supported metric-view subsets; a Databricks instance is not required for offline compiler work, but such work should not displace a more immediate investigation failure without a selected portability task.

Do not remove Operations or native measurements merely to reduce tab count. They represent capabilities not fully replaced by a record stream. Consolidate navigation and presentation when a real continuation exists, preserving old links and exact references. No new monolithic workbench or universal evaluator is justified by this review.

## Recommended finite sequence

1. Complete and review the current measure-authoring field-selection consistency cut.
2. Deliver trace anchor continuity diagnostics; review actual preserved findings and failure cases.
3. Deliver bounded local ordinary draft recovery; review reload and uncertain-save behavior.
4. Reconcile README/current queue and close the round with focused validation and cleanup.

Pure protocol extraction may run as a separate bounded maintenance cut after fixtures establish parity. It is not a replacement for the user outcomes above. Pipeline evidence and broader durable continuity need a follow-up proposal grounded in actual available sources and deployment requirements.
