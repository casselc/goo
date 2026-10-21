# Refine the core investigation experience

Surveyed 2026-10-21 against the current workbench and official documentation sources.
This is a product recommendation, not a claim of implemented functionality or
competitive parity. It complements the [component sweep](WHOLE_PRODUCT_COMPONENT_SWEEP.md).

Bounded first slices are now implemented in [core experience refinements](CORE_EXPERIENCE_REFINEMENTS.md). This survey preserves the pre-implementation research and priorities; that workflow document records current behavior and limits.

Goo's strongest direction is **a question that stays understandable as you move
from service symptoms to individual records to a preserved explanation**. The
next additions should shorten that journey. The application already has most of
the necessary chart types; another collection of dashboards would add less value
than connecting the existing views.

## Research basis and limits

Official documentation verified through the vendors' public source repositories:

- **Datadog Service Page:** resource-level request/latency/error breakdowns,
  sortable resource lists, resource side panels, dependencies and deployment
  context. [Published page](https://docs.datadoghq.com/tracing/services/service_page/)
  · [source read](https://github.com/DataDog/documentation/blob/master/hugo/content/en/tracing/services/service_page.md).
- **Datadog Trace View:** flame graph, grouped span list, waterfall, trace map,
  span details and execution-time breakdowns. The useful pattern is several
  coordinated views of the same trace, rather than separate destinations.
  [Published page](https://docs.datadoghq.com/tracing/trace_explorer/trace_view/)
  · [source read](https://github.com/DataDog/documentation/blob/master/hugo/content/en/tracing/trace_explorer/trace_view.md).
- **Grafana exemplars:** explicit links from a metric observation to a
  representative trace; comparing fast and slow traces as a diagnostic workflow.
  [Published page](https://grafana.com/docs/grafana/latest/fundamentals/exemplars/)
  · [source read](https://github.com/grafana/grafana/blob/main/docs/sources/fundamentals/exemplars/index.md).

Published vendor sites returned proxy `403 Forbidden`; the above statements were
verified from successfully retrieved official Markdown, not screenshots or a
running vendor account. These source links follow moving branches.

**Honeycomb BubbleUp is a relevant comparison to revisit**, especially selecting
an unusual population and examining distinguishing dimensions:
[official documentation entry point](https://docs.honeycomb.io/investigate/analyze/explore-data/bubbleup/).
Its site was blocked and attempted public source paths returned 404. Its current
behavior, algorithms, limits and packaging were **not verified**. The cohort
recommendation below is an independent recommendation based on goo's implemented
comparison semantics; it does not depend on reproducing Honeycomb's algorithm.
Grafana correlations and other vendor capabilities not successfully read are not
presented as verified findings. No hands-on competitive usability test was run.

Local review covered README, implementation/known-issues plans, product essence,
core primitives, question workbench, service dashboards, comparison breakdowns,
deployment context, selected-item inspectors and the display-component contract,
with spot checks of the UI's cohort comparison implementation.

## Ranked shortlist

| Rank | Addition | Existing foundation | Actual remaining gap / first useful scope |
| --- | --- | --- | --- |
| 1 | **A resumable investigation trail** | Typed question steps; immutable captures; document revision branches; explicit follow-ups; local selected-item inspectors | A small application-owned trail joining service → question → trace/log → comparison. Back restores fixed scope, applied question, selected evidence and scroll without rerunning. Keep unsaved draft state distinct from the applied result. Start in one browser session; durable resume is a separate storage contract. |
| 2 | **A compact service diagnosis table** | Automatic service dashboards, RED/USE recipes, namespace-aware identities, declared operations, exact reported-edge population questions | One ranked route/operation table with aligned count, recorded-error fraction and latency columns, plus a deliberate instance pivot. Clicking a row focuses a question under the same bounds. Retain current service overview; replace repeated cards where the table communicates better. Dependency rows use reported relationships, not guessed topology. |
| 3 | **Select a visual anomaly, then investigate its population** | Span-duration buckets, grouped time charts, heatmaps, typed filters, one-field cohort comparisons and stored breakdown recapture | An accessible selection contract that turns a supported time/duration selection into a previewed typed question, then an explicit run. Continue into an existing comparison and dimension breakdown, rather than adding another analyzer. Start with span populations; native histogram cells cannot identify contributing traces without additional evidence. |
| 4 | **A trace breakdown beside the waterfall** | Cross-service trace expansion, selected-span inspector, shared waterfall and exact linked-log scopes | A compact service/operation span-count and duration breakdown coordinated with waterfall selection, plus search/error highlighting if absent from the active presenter. Begin with accurately named inclusive duration statistics. Add exclusive elapsed-time measures only after Go defines overlap/clipping/parentage rules; do not label these CPU time or critical path. |
| 5 | **One aligned context lane across relevant views** | Explicitly loaded deployment receipts, event/custody clocks, a marker lane, exact note references, independent telemetry reads | Reuse the same compact context lane beside service/question/comparison timelines, preserving each panel's actual bounds. Selection opens an existing receipt inspector. Keep markers authored context; never silently amend a capture or infer a faulty deployment. First reuse the existing receipt model; CI ingestion/corrections are later work. |
| 6 | **A visible path from missing input to useful instrumentation** | Typed native validity states, source receipts, numeric-measure coverage, package previews, optional annotation handoff, generic OTel exploration | A compact per-output capability/coverage explanation: observed, no matching records, unsupported, needs binding, or unknown. An unavailable route/USE view should offer an actionable binding or package task, while other supported service views continue working. Convention-profile resolution and machine-readable capability negotiation remain real model work. |

### Why these are coherent

The trail makes investigation navigable. The service table locates a suspicious
operation. A visual selection isolates a population. A trace breakdown explains
an individual example. Context and capability states keep that interpretation
honest. Each feature can reuse the same question, evidence summary and inspector
instead of creating another permanent top-level destination.

Do not rebuild already-shipped features as gaps: goo already compares captured
group counts, shares, mean durations and recorded status; supports fixed/live
comparisons within documented limits; has service dashboards, deployment notes,
attribute facets, a waterfall, native timelines and heatmaps. The missing work is
mostly navigation, coordinated selection and semantic coverage, not basic charts.

## Three increments to implement first

1. **Return to the question.** Add the application-owned trail across one service
   question, trace selection and comparison. Use stable result/capture identities;
   returning to retained evidence reads no source. When evidence is no longer
   retained, show that state and offer explicit reopening/rerun. Preserve the
   existing independent trace-to-log read behavior and label it. This builds on,
   rather than expands the responsibilities of, the selected-item inspector.
2. **Find the operation.** Ship the compact service table for one supported OTel
   profile, initially HTTP server routes with generic span-name fallback clearly
   labeled. Reuse Go aggregates, shared scales and passive tables. Offer dependency
   and instance pivots through checked question transforms. Avoid a configuration
   builder: discovery/manifest capabilities determine the available defaults.
3. **Follow a spike.** Support a bounded time/duration selection in the span view,
   with keyboard-equivalent controls, exact boundary preview and explicit run.
   Reuse existing comparison/breakdown capture controls. Open representative
   matching records through a labeled query; do not promise unbiased samples or
   claim a metric exemplar where only temporal proximity exists.

Trace breakdown and context-lane reuse can follow these slices. A genuine metric
exemplar path requires ingesting, retaining and validating exemplar trace/span
references first; rendering a clickable marker alone does not establish that
relation. Unknown/coverage presentation belongs in each increment rather than a
separate quality-score project.

## Semantic and interaction constraints

- **RED is a declared population and policy.** Server/client spans, log severity,
  span status, protocol response and business outcomes are different. A table
  must name which it counts. A span count over a window is not automatically an
  unbiased request rate; sampling/collection coverage can remain unknown.
- **USE needs the relevant resource and denominator.** Database client latency
  does not establish database saturation. Missing host/pool capacity is a reason
  to withhold utilization, not display a healthy zero.
- **Relations have types.** Parent edges, async span links, matching trace IDs,
  explicit exemplars and service/time context must not collapse into one generic
  “related” relation. No inferred causal dependency graph.
- **Comparisons retain their basis.** Independently read windows can overlap,
  disagree or contain late arrivals. Group-share changes are descriptive leads;
  they are not statistical attribution or root-cause findings.
- **Keep time meanings separate.** Event occurrence, source-read time, capture
  custody time and authored document revision time answer different questions.
  A newer read can find earlier events; a later document does not change an old
  capture. Show UTC, retain exact timestamps in details and avoid claiming an
  as-of historical database unless its source contract supports it.
- **OTel is the default vocabulary, not a requirement for goo-specific code.**
  Resolve service namespace/name/instance, resource versus point attributes,
  span kind, units and convention revisions explicitly. Missing schema metadata
  permits generic exploration; convention-specific output may need a binding.
- **Keep the display library passive.** Reuse table/facet/summary/timeline/heatmap/
  waterfall primitives; application adapters own selection, transforms, source
  reads and history. Introduce a new component only through real repeated use.

## Evidence of success

Keep the acceptance scope small and executable:

- **Playwright storyboard:** service overview → operation → suspicious interval →
  trace/log → comparison → capture → finding; return twice without changed scope,
  lost draft, collapsed selection or an unexpected source read. Include screenshots
  at a compact viewport and UTC tooltip/accessibility checks.
- **Bombadil:** actions select, clear, change chart view, open/close inspector,
  navigate back, edit a draft and reopen frozen evidence. Invariants retain
  question/evidence identity, preserve unknowns and prevent stale results from
  becoming attached to a new draft. Assert read behavior according to each
  action's actual contract; existing trace-linked logs are an independent read.
- **Hegel state/swarm tests:** generated selection intervals preserve half-open
  bounds and predicate semantics, cannot widen scope unexpectedly, distinguish
  missing/empty fields and reject incompatible native units or identities.
- **Quint/quint-go MBT:** extend existing navigation/execution models only for the
  new transitions; validate stale completion, cancellation, return-to-evidence
  and explicit rerun traces. Model a retained result separately from a definition
  and immutable capture. Do not add a second competing workflow model.

The product test is whether an operator can answer **“what changed, where should I
look, and what evidence supports that?”** with fewer screens and less rereading.

## Leave out for now

No general dashboard builder, plugin ecosystem, multiplayer whiteboard, AI
root-cause agent, automatic causal attribution, full alert/incident-management
suite, profiler/RUM/security expansion, or decorative topology map. No new query
language in the browser. No native fleet live aggregation before its membership,
coverage and alignment contracts exist. Production sharing also requires the
separate authorization, isolation, retention and durable-storage work already
identified in the project plans.

Those are substantial products in their own right. They should not delay making
goo's existing question-to-evidence journey compact, reversible and trustworthy.
