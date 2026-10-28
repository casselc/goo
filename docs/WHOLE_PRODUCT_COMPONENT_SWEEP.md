# Whole-product component sweep

Reviewed 2026-10-20 against the implemented product at `068389f`. This is a source/design audit with inspection of existing recorded screenshots, not a new live-browser validation campaign. Recommendations below are **proposed**. The existing component contracts and their actual checks remain in [UI_COMPONENTS.md](UI_COMPONENTS.md) and [VALIDATION.md](../VALIDATION.md).

Implementation update: the first [evidence-summary adoption](EVIDENCE_SUMMARIES.md) now covers native capture and linked investigation surfaces. Remaining priorities below stay proposed.

## Main finding

Goo now has enough plot primitives to express several useful investigations. The next broad improvement should make **scope, evidence, selection and continuation** consistent around those plots. Extracting another generic card would preserve the current accumulation of paragraphs and controls. A compact evidence-summary primitive, adopted in multiple real surfaces, would instead make it easier to answer: **What am I looking at, what population does it describe, and what will this action do?**

Keep the product's common journey visible: choose a service or question → inspect a result → select an interval, group or record → inspect context → retain evidence and a finding. Services, Measurements, record exploration and saved investigations are entry points into that journey, not separate visual products. Native metrics and captured span measures still require different semantic adapters.

The most valuable next implementable component is **`goo-evidence-summary`**, first adopted in a captured measurement and a linked investigation evidence entry. It would remove repeated scope/clock paragraphs without changing source reads, question identity, comparison policy or custody. A small-multiples composition is the next complex display candidate, after existing shared plots have been adopted where their contracts already fit.

## What is implemented and should be reused

| Existing primitive or composition | Current responsibility | Further reuse opportunity |
| --- | --- | --- |
| `goo-table`, `goo-result-state`, `goo-facet-list` | Native table viewport; explicit result status; local filtering of supplied facet rows. | Migrate remaining raw table/status variants before building a richer datagrid. Keep paging, ranking, population totals and source reads with adapters. |
| `goo-field-value`, `goo-attribute-list`, `GooAttributes` | Typed safe values and scope-grouped attribute inspection, including local find and explicit copy/filter/breakdown actions in the adapter. | Exact stream identities, receipt/definition fields and selected-record details; avoid new JSON blocks for ordinary identity fields. |
| `goo-latency`, `goo-stat-list`, `goo-metric-comparison` | Compact supplied span statistics and paired metrics on a common axis. | Telemetry-package preview still renders its own latency sentence; retain exact tables for unsupported signed/scientific/rational representations. |
| `goo-chart-frame`, `goo-time-bars`, `goo-value-timeline` | Shared chart metadata; supplied count intervals; exact-time scalar observations/intervals. | Grouped record time series and remaining operation charts have local plot chrome. Do not force every quantity into the count-bar contract. |
| `goo-trace-waterfall`, `goo-bucket-heatmap` | Supplied trace layout/selection and exact bucket-count grids. | Compose them with consistent selection/detail and evidence summaries. No need for a second waterfall or heatmap implementation per entry point. |
| `GooUI.actions/disclosure`, `GooQuestionScope`, `GooServiceIdentity` | Native secondary actions; Go-checked transformation explanation; explicit qualified service identity. | Reuse these helpers. Service identity and question transformation remain application semantics, not field-name guessing inside a component. |
| `GooTail`, `GooLiveComparison`, `GooServiceDashboard` | Application compositions owning subscriptions, read budgets, lifecycle and semantics. | Preserve as adapters. Reuse display subcomponents within them; do not turn them into subscribing custom elements. |

The older [component audit](UI_COMPONENT_AUDIT.md) describes the initial table/facet/value extraction and future streaming requirements. Several of those proposals are now implemented. The current application already has an SSE log-tail adapter and rolling record/counter comparisons; it is inaccurate to describe all streaming delivery as unimplemented. Display lifecycle still must never start a subscription.

## Priority and adoption order

| Priority | Deliverable | Concrete duplication or friction | First two adoption surfaces | Form |
| --- | --- | --- | --- | --- |
| 1 | Evidence/scope summary | `captures.show`, `documents.questionEvidence/evidenceSummary`, `exploration.renderSession`, `service-browser.open`, and comparison adapters repeatedly construct scope, clocks, basis and coverage paragraphs. | Frozen measurement capture; linked capture in investigation notes. Then service widget and fixed/live comparison sides. | Passive `goo-evidence-summary` with adapter-supplied display facts. |
| 2 | Complete existing table/value/chart adoption | `inventory.mountEvidence` uses a raw `.scroll` table; `telemetry-packages` preview uses raw grouped tables and a latency sentence; `distributions.renderComparison` builds another cohort table; `grouped-time-series` builds local `.time-bars`. | Saved inventory and package preview for tables/stats; grouped count series and a Services chart for plot composition. | Existing components and shared styles, not new public elements. |
| 3 | Selected-item inspector composition | `exploration` trace/surrounding inspectors, `measurements.renderTimeline`, grouped chart status and operation span inspection each create selection summaries, close/return controls and detail disclosures. | Native interval inspector; trace/log record inspector. | Shared child-preserving layout/helper first. Introduce a custom element only if the two adapters prove a stable shell contract. |
| 4 | Shared-scale small multiples | `grouped-time-series` combines filtering, shared geometry, labels and per-cell inspection locally; automatic service cards have a different chart composition. | Count split series first; later compatible scalar series with an explicit supplied domain. | Passive composition over existing plots, not a metric evaluator. |
| 5 | Revision-aware catalog rows | `telemetry-packages.refresh`, `exploration.renderQuestions`, measurement libraries, `documents.renderLibrary`, capture cards and example cards repeat labels/search/actions. Several show immutable revisions as many similarly named cards. | Record definition library and measurement definition library. | Shared native row/card CSS and renderer; adapter-owned family grouping/history. A custom element is optional. |
| 6 | Comparison composition | `evidence-comparison`, `distributions.renderComparison`, `fresh-comparison` and `live-comparison` expose similar two-sided scopes and paired results through different layouts. | Stored record comparison and native captured-window comparison. | Shared presentation shell using evidence summaries, existing paired plots and caller-owned actions. Keep compatibility and freeze/capture protocols separate. |
| 7 | Qualified field/filter display | `exploration.markFilters/options`, numeric binding controls, attribute actions and package selectors label fields differently. | Applied Explore chips and definition selector summary. | Small native filter-token/field-label helper first; no universal predicate editor. |

Adopt a new abstraction in two real workflows before broadening its API. Priorities 1 and 2 should reduce product complexity immediately; priorities 3–7 are bounded follow-ups, not a commitment to create seven new elements.

## Proposed evidence-summary contract

The adapter supplies a compact display snapshot. It has already interpreted the typed result, qualified namespace, definition identity and clock roles. A possible bounded shape is:

```js
summary.data = {
  mode: 'frozen',
  state: 'partial',
  facts: [
    {label: 'Population', value: 'Exact native histogram stream'},
    {label: 'Scope', value: 'production · checkout · namespace missing'},
    {label: 'Window', value: '[2026-10-20T10:00Z, 2026-10-20T10:05Z)'},
  ],
  notice: 'Coverage is incomplete; unknown intervals remain unknown.',
  details: [
    {label: 'Source read', value: '2026-10-20T10:05:02Z'},
    {label: 'Stored', value: '2026-10-20T10:05:03Z'},
    {label: 'Definition revision', value: '…'},
    {label: 'Evidence basis', value: '…'},
  ],
};
```

This is an illustrative presentation API, not a new persisted receipt schema. Set limits before implementation, for example at most eight visible facts and 32 exact detail rows, with bounded strings. Preserve exact ID/timestamp strings in a native disclosure; adapter-selected visible facts can be concise. Avoid requiring every caller to supply a service or unit when its population has neither.

Required behavior:

- Support JSON attributes in server-rendered HTML and detached/pre-definition SPA `.data` assignments, using the existing safe-text and detached-getter conventions.
- Keep evidence mode, result state and connection/follow state separate. A frozen capture can contain unknown intervals; a connected tail can have a continuity gap; an observed span result does not imply service health.
- Put the decisive caveat next to the result. Incompatible units/layouts, changed-between-reads and partial coverage cannot become details-only metadata. Repeated general caveats can move into one exact-basis disclosure per owning result.
- Keep source event time/window, source-read time, evaluation time, knowledge cutoff and stored custody time separately labeled when supplied. Never infer a clock role from the first timestamp found in an object. Nanosecond bounds remain exact; display rounding does not replace them.
- Namespace missing, present-empty, a specific namespace and wildcard remain distinct. Full resource/scope/version/schema identity remains inspectable. Prefer the existing typed attribute renderer for large identities.
- Keep native links/buttons in the embedding adapter. Disclosure is the only initial standardized interaction; no subscription, refresh, retry, capture, clipboard policy or navigation belongs inside this component.
- Fail closed on invalid input with an explicit unavailable display. A local search yielding zero rows must not overwrite the original result's evidence summary or counts.

Use a child-preserving native content region if adapters need rich identity/action children. Do not replace caller-owned child nodes merely because scope text updates. If structured detail rows rerender, document their disclosure/focus behavior instead of implying high-frequency keyed reconciliation.

## Proposed complex compositions

### Selected-item inspector

Use one visual sequence: **selected item identity → immediate value/state → primary continuation → attributes/exact evidence**. The body is supplied native content, allowing a trace span, log record, histogram interval and lifecycle subject to remain genuinely different.

A shared shell can standardize heading level, labeled region, close/return affordance, compact action placement and mobile containment. Selection itself stays controlled by the parent adapter. A chart or table emits only an opaque selection key; the adapter resolves it against retained evidence and supplies inspector content. Changing selection must clear prior fields before showing an unknown/missing item. A disabled other-layout cell must never select a bucket by index.

Focus/return policy remains explicit in the application: preserve the triggering node while a local detail updates; on closing an overlay return to the still-connected opener or a documented fallback. A passive inline inspector should not autofocus on every live update. Loading a trace's additional pages or surrounding logs is a separately labeled source action, not expansion of the retained capture.

Do not implement a universal record schema or a generic side panel that receives raw query objects and decides which telemetry to fetch. First share native layout and close/heading helpers between two existing inspectors; the root workflow can then determine whether a public element earns its complexity.

### Timeline + distribution grid + selection inspector

Native histogram inspection already composes an exact-time timeline, a compatible-layout bucket grid and an interval detail panel in `measurements.renderTimeline`. Preserve that composition as the shared native result renderer used by live fixed reads, saved captures and retained service-card details. A new chart family does not need a new service-only or capture-only inspector.

The stable application contract is a retained series plus a controlled selection: `{intervalKey, bucketKey?}`. Timeline requests identify the interval; grid requests add an optional bucket key. Keys resolve against the current retained snapshot, not row position in a filtered table. The adapter owns layout changes, unknown-state clearing, exact bounds, view mode and whether selected evidence lies outside a plot's display budget. The passive plots own geometry and focus-preserving selection state only. A local view switch must not silently rewrite the question or imply a source read.

Use `goo-chart-frame` once for the owning visual result and keep the shared evidence summary outside repeated plot chrome. Show the primary visual first, with alternate distributions and exact tables in restrained disclosures or explicit local view controls. Selection details should remain in one predictable place. Preserve a count plot's exact-time gaps and distinguish a retained-interval grid from any future fixed-time aggregate; changing a label cannot establish an aggregation contract.

The root implementation is separately working on explicit Go-backed histogram time-bucket aggregation and a local raw/fixed-time view switch. That work was **in progress at this audit snapshot**, not implemented or validated by this report. Subsequent integration and executable checks are documented separately in [HISTOGRAM_TIME_BUCKETS.md](HISTOGRAM_TIME_BUCKETS.md) and the validation record. Its result shape must come from Go; no component should split an interval or infer coverage while drawing cells.

Acceptance should exercise selection across timeline/grid/view controls, unknown data after a known selection, changed bucket layouts, selection outside the displayed subset, keyboard focus and source-read counts. Stable selections can survive a display-only view change only when the adapter can establish the same retained identity; otherwise clear them visibly rather than matching bucket/column indices.

### Investigation case/workspace

The useful larger composition is a **question with its evidence and finding**, rather than a generic whiteboard card. `documents.render`, `questionEvidence`, `evidenceSummary`, `fresh-comparison` and saved comparison cards already assemble its pieces: question scope, reopen/run-and-capture controls, linked capture history, comparison and narrative. Today these may expand into several independent sections with repeated IDs, timestamps and scope paragraphs.

A proposed case layout presents one question heading and scope summary, a compact evidence-history list, one selected evidence/comparison region, then the finding and open questions. Reuse the existing native action disclosure, evidence-summary primitive, comparison displays and selected-item layout. Keep a clear route back to the parent investigation and preserve unsaved note text when switching inspected evidence. A readable linear document should remain the primary representation; a future spatial/multiplayer view can reference the same immutable case/evidence identities rather than creating another persistence model.

Initially implement this as an application renderer in `documents.js`, not a subscribing `goo-investigation-workspace` custom element. The adapter must continue owning draft/context tokens, narrative revision branching, card ordering, linkage, receipt reads and capture/run semantics. A passive shared case-section shell can standardize headings and regions after it proves useful for both linked-question evidence and standalone saved-comparison entries. Do not hide these policies inside a generic object inspector or global component store.

A future display snapshot might supply `{caseKey, title, scopeFacts, evidenceEntries, selectedEvidenceKey, findingState}` with caller-owned narrative controls. Opaque `caseKey`/`evidenceKey` selection requests are reasonable; save/publish/run/capture operations are native adapter actions with their current explicit descriptions. This is a presentation sketch, not a new case schema or server endpoint. Validate focus/scroll retention, exact revision references, unsaved-draft preservation and refusal of stale async responses before claiming a reusable workspace composition.

### Shared-scale small multiples

`grouped-time-series.js` is the clearest remaining local chart composition. It computes display geometry across the full retained series set, then local search hides cards without changing that domain. Preserve those semantics while adopting shared chart frames and plots.

Start with count series, where `goo-time-bars` already fits. Each series receives supplied keys, exact values/states and one caller-supplied common ceiling. A selection request identifies `{seriesKey, pointKey}`; the adapter displays exact cell bounds and measures. Search affects visibility only. A count-series extraction should not add duration evaluation, rate normalization or hidden group truncation.

Duration/rate/signed numeric plots need a declared compatible display domain and appropriate scalar primitive. The existing numeric split chart intentionally plots absolute magnitude while retaining signed labels; do not accidentally turn that into a signed axis through a generic count-bar migration. Different units, histogram layouts and native temporality cannot be made comparable merely by sharing geometry. Native percentile ranges remain ranges.

Proposed acceptance: equal supplied values occupy equal geometry across cards; zero differs from unavailable; local search preserves domain, full group totals and selection; oversized data fails/exposes an explicit display budget; exact cell tables remain accessible. Scope and independent-read discrepancies stay visible outside the plots.

### Revision-aware catalog

A common row can present title, kind/signal, qualified scope, revision metadata, one primary action and a native secondary disclosure. That removes repeated button rows and makes saved definitions, captures and question templates easier to scan.

The display component cannot choose a mutable "latest" revision. The current investigation library explicitly preserves branches and has no mutable latest head; record/measurement definitions also expose immutable revisions. Family grouping requires an explicit adapter/server rule and must retain lineage and exact revision links. Initially, share the row presentation while preserving current ordering and supplied counts. Add grouping only when its family identity and representative-selection rule are designed and tested.

Search on the current catalog page is not global discovery. Show local match/page counts distinctly, and provide the existing explicit next/load action. Empty source catalog and zero local matches need different text. Reuse the `goo-facet-list` child-preserving local-filter pattern where it fits, without pretending catalog cards are measured distributions.

### Comparison composition

All comparison views should lead with a stable **Baseline | Comparison/current** orientation, visible units and compatible scales, then exact values and provenance. Highlight the most informative visual, with related measures consolidated into the existing paired matrix. Keep the native histogram share chart primary; raw-count grids and exact bucket tables are supplementary because differing window duration can affect raw populations.

A shared shell can arrange two evidence summaries, supplied compatibility/refusal text, paired plot children, exact detail children and caller-owned continuation controls. It must not decide that stored, fresh-read and rolling comparisons are interchangeable. Existing baseline pinning, two explicit capture reads, displayed-result freezing, token expiry, rescaling and source-read generations remain in their current adapters/Go kernels.

A shared shell should reduce duplicate warning blocks and empty headings, not introduce a mandatory tall card around every existing plot. Start with static captured comparisons; adopt rolling views only after proving paused/stale/gap display and fixed-baseline identity remain intact.

## Whole-product workflow coherence

| Entry point | Keep the first useful action obvious | Consolidate/defer | Preserve visibly |
| --- | --- | --- | --- |
| Explore | Choose signal/scope, run a question, inspect records/distribution. | Reusable library, numeric binding, advanced filters and raw plans can remain disclosures. Use one applied/draft question summary. | Draft edits versus applied data, fixed bounds, full totals versus inspection page. |
| Services and distributed services | Open a qualified service or exact reported relationship, then a useful default overview. | Show instrument variant identity compactly; retain full scope/version/schema in details. Do not duplicate a metric-only service into a second catalog. | Observed versus declared, missing/empty namespace, independent widget reads, no health inference. |
| Measurements | Choose/build definition → preview scope → run → inspect stream/interval. | Exact identity and raw plan under one evidence section; keep unknown reason guidance near the affected value. | Native temporality/reset/layout, latest-interval versus whole-window result, exact pin and read limits. |
| Trace/log context | Select a record, inspect waterfall and attributes, follow exact links or explicit time context. | Reuse selected-item shell and table/value components. Keep long raw JSON as fallback, not default identity display. | Broader trace scope, additional source reads, missing parents/duplicates, context time proximity versus identity. |
| Comparisons | Understand which populations differ and see the difference on a common scale. | Related percentile measures as one group; counts versus shares labeled distinctly; exact tables secondary. | Compatibility refusal, window width, stable baseline and source-read/custody clocks. |
| Saved investigations | Reopen a question or frozen evidence, see the finding and continue deliberately. | Short evidence-summary entries instead of repeated scope/read/storage paragraphs. Catalog history can be a compact list. | Unsaved drafts, exact immutable references, revision branching, capture reread versus reopen. |
| Definitions, recipes and examples | Choose a useful question, inspect what it will keep/change, preview/run explicitly. | Shared catalog row and action hierarchy; package editor appears after selection. Examples reuse ordinary product paths. | Configuration versus evidence, validation versus telemetry reads, qualified fields and explicit instrumentation gaps. |

The inspected service-overview frame [76b](journeys/screenshots/76b-selected-service-overview.png) shows why a compact identity/evidence treatment matters: native variants are correctly distinguished but create repeated metadata lines. The current captured-histogram frame [54b](journeys/screenshots/54b-whole-window-histogram-buckets.png) shows successful shared visual scales alongside several provenance/table disclosures that could share one predictable hierarchy. Historical frames such as 53a are useful to understand former layouts, but must not be treated as proof that current comparison visuals are still raw tables.

The full-product entry-point taxonomy should remain a product decision. Do not encode navigation routes or "what should I investigate next?" rules into a display library. Prefer the same question/preview/open/capture vocabulary wherever the underlying action is actually the same; retain explicit qualifiers when it is not.

## Verification and adoption plan

1. Implement the evidence-summary contract and standalone reference fixture. Adopt it in captured native measurement and investigation evidence entries in the same milestone. Keep old receipt data and source-read behavior unchanged.
2. Prove SSR escaping, exact large IDs/timestamps, missing/present-empty namespaces, safe text, detached/pre-definition properties, invalid input, reconnect behavior, heading hierarchy, native disclosure keyboard behavior, mobile containment and supplied-child focus retention. Do not announce every metadata row through nested live regions.
3. Extend real capture/investigation journeys to prove a recipient reopens stored evidence without a telemetry read, clocks keep their labels, and unknown/partial/incompatible states do not inherit older values. Update screenshots for both adoption paths. The standalone synthetic fixture is not integration evidence.
4. Extend Bombadil with actions that open exact-basis disclosures, move between question and evidence, and return from inspectors. Invariants should cover unchanged exact scope/receipt, no source reads for local disclosure/selection, explicit missing/unknown state, mobile overflow and valid focus return.
5. Keep existing Go/Hegel semantic properties and observed Quint custody/read traces around the adapters. Add new model transitions only when application behavior actually changes. Pure passive-component extraction does not justify inventing a second formal metric evaluator.
6. Complete the concrete table/stat/chart adoption debt before extracting the next complex composition. For shared-scale plots, add invariant tests for stable domain under local filtering, retained selection and no percentile arithmetic. For any later live adoption, test keyed row/focus preservation, pause/gap/reconnect and immutable baseline behavior against the real adapter.

No new backend endpoints, semantic evaluator, global store, transport framework, charting framework or component registry is required by this audit. Reuse light DOM, native tables/details/buttons and existing theme tokens. Standardize public `data`/detached-property behavior and document each element's small selection event, but keep their typed selection keys; do not unify unrelated chart events into a universal business-action bus.

## Source map

Primary implementation references for follow-up work: [shared primitives](../ui/components.mjs), [typed attributes](../ui/attributes.mjs), [record exploration and inspectors](../ui/exploration.js), [grouped series](../ui/grouped-time-series.js), [native results and selection](../ui/measurements.js), [Services entry/detail](../ui/service-browser.js), [service widgets](../ui/service-dashboard.js), [saved inventory](../ui/inventory.js), [captures](../ui/captures.js), [investigation cases](../ui/documents.js), [captured comparisons](../ui/evidence-comparison.js), [rolling comparisons](../ui/live-comparison.js), [definition library/preview](../ui/telemetry-packages.js), and [navigation/question links](../ui/workspace.js).

Service-widget follow-through is implemented: one compact population fact, local scope/read/basis disclosure, fixed evidence mode and explicit capture reread semantics. Timestamp display now separates readable UTC text from exact tooltip values in shared summaries and native tables/selections. Stored comparison sides now use the shared summary; scalar pairs align horizontally and retain common scales. The shared inline selected-item shell is adopted for trace spans, record inspection and native selections; cross-query navigation history remains future work.

The shared selected-item shell is now adopted in waterfall spans, span/log record inspection and native timeline/bucket selections. See [selected-item inspectors](SELECTED_ITEM_INSPECTORS.md) for lifecycle, retention, focus and source-action boundaries.
