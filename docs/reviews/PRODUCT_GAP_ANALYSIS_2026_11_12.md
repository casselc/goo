# Product gap analysis — November 12, 2026

Scope: quick feature/workflow comparison with Datadog, Honeycomb, Grafana and
HyperDX/ClickStack, reconciled against Goo at `9309e44f`. Recommendations below
are proposals, not implemented capabilities or a second execution queue.

## Assessment

Goo already has live logs/spans, typed filters/facets, latency brushing,
selected-versus-baseline explanations, trace waterfalls and timing breakdowns,
service/operation dashboards, host/Kubernetes views, native measurements,
exemplars, reusable scoped views, captures and investigation notes. Rebuilding
those categories is not a gap-closing strategy. The stronger opportunity is to
help an operator find the important change quickly, then reuse the existing
investigation loop. Another chart library or query engine is not a prerequisite.

The current finite roadmap instead starts with review-task discovery and adjacent
labeled occurrences. Both are valid, but they should no longer be the default
next priorities for a general observability product. Recommend an operational
breadth/depth/consolidation cadence first; retain labeling work as a deferred,
explicitly scoped track. The active roadmap records this recommendation without
silently renumbering or claiming delivery of 412–414.

## Prioritized additions

| Priority | User task / reference pattern | Goo today and actual remaining gap | First useful slice |
| --- | --- | --- | --- |
| Next: breadth | **Which service needs attention?** Datadog's service/resource summaries make operational symptoms scannable. | The unified catalog has separate signal counts and event age; individual services already have operation diagnosis and supported profiles. A bounded cross-service symptom ranking is missing. | Add an optional service triage table for one fixed scope: captured server-span volume, recorded Error fraction, supported latency and previous-window changes. Sort/filter, then open the existing service or exact population. Use bounded grouped queries, not dashboard fan-out per service. Retain log-only/metric-only services with unavailable cells; do not invent a universal health score or mix span kinds into request rates. |
| Next: depth | **Is this log burst unusual, and what makes it up?** HyperDX emphasizes live search, event deltas and low-friction investigation. | Activity already has optional refreshed span latency, baseline comparison and explanations. Logs do not have the equivalent compact full-population volume/severity context at the stream entry. | Reuse count timelines for a compact log-volume strip with severity breakdown, explicit cadence and time brush into the existing explanation. Keep full-query counts separate from retained-buffer counts and severity number separate from reported text. Pause on interaction; preserve held return. This extends the existing owners, not a second stream engine. |
| Next: consolidation | **Get back to my useful scope without reconstructing it.** Grafana's query-history pattern and reusable investigation documents are relevant. | Saved questions, Activity configurations and up to eight retained visits exist; durable personal recent-query/resume across reloads does not. | A compact Recent/Pinned picker using existing typed questions and views. Record configuration, not telemetry bodies; distinguish staging a question, explicit Run and reopening a capture. Bound retention and offer Clear. Personal identity/isolation must be designed before this is shared between users. Use this pass to unify pickers, rather than adding another top-level workspace. |
| Following: depth | **Did the new version change latency or recorded errors?** Datadog connects deployments, resources and traces. | Deployment receipts/context lanes and arbitrary population comparisons already exist. A guided, repeatable version-regression workflow is missing. | Discover reported `service.version` values in the current service scope, choose two explicitly and compare the same operation under a common window or explicit before/after windows. Reuse existing comparison AST, rates, captures and held Back. Show unequal exposure/counts and missing versions. A receipt or nearby deployment is context, not proof that a release caused a regression. |
| Following: breadth | **Are these recurring errors the same problem?** Datadog documents errors aggregated into issues; HyperDX offers log search and event-change workflows. | Goo triages recorded exception/status/severity evidence and has versioned log-shape patterns. It does not have durable exception issue identities, observed history or triage lifecycle. | First add opt-in deterministic exception fingerprints from explicit exception type/stack data, preserving original records and fingerprint version. Show count, affected services/versions and first/last *observed within the searched history*. Keep unclassified errors visible. Later add acknowledged/resolved state with explicit reopen rules; do not call existing log patterns issue fingerprints. |
| Strategic, gated | **Tell me when this saved condition persists.** Honeycomb trigger/SLO definitions and HyperDX alerts make proactive use explicit. | Goo has measurement/SLI foundations and preview paths, not a durable production monitoring/notification service. | Start with a saved-query threshold evaluation and dry-run history: durable scheduler, idempotent evaluation identity, pending/firing/recovering/unknown state and explicit no-data policy. Add deduplicated webhook delivery, routing/silence and delivery status after that. True SLOs/burn rates need a validated request/event denominator and coverage/window policy; captured-span Error fractions do not automatically qualify. |

These additions need not add six new destinations. Fleet triage belongs in
Services; log context stays next to Activity; version comparison starts from the
service/context lane; error groups enter through Investigate; Recent/Pinned is a
shared picker; monitor creation begins from an already understood saved question.
Every new visual should consume the checked data model and existing shared
components. Missing, unsupported, stale and uncertain data remain visible states.

## Useful smaller follow-ons

- **Declared external correlations:** safe links from exact trace IDs, service
  scope or an inspected field to an existing HyperDX/Grafana view or runbook.
  Preview the generated scope; allowlist templates/URLs. Do not send captured
  bodies or sensitive arbitrary attributes implicitly. Avoid rebuilding every
  specialized tool before offering a useful continuation.
- **Data/pipeline status:** actual collector/exporter dropped/rejected/lag signals
  and instrument capability diagnostics, where source adapters can supply them.
  Latest event age alone is not ingestion health. This should gate confidence in
  monitoring, rather than becoming another decorative green health indicator.
- **Shared deployment prerequisite:** authentication, authorization/source scoping,
  auditing and resource budgets before team sharing or externally reachable
  monitoring controls. Existing per-process bounds/cancellation are useful but
  are not tenancy or authorization.

## Keep out of the immediate queue

RUM/session replay, continuous profiling, synthetic monitoring, automatic causal
root-cause claims, inferred topology/critical paths, a freeform multiplayer
whiteboard, arbitrary executable view plugins and a broad AI assistant. These
are potentially useful products, but require new collection/security/identity or
semantic contracts and do less to complete Goo's current operator loop. Exponential
histograms and warehouse lowering retain their existing source/semantics gates.

Do not schedule generic brushing, BubbleUp-like explanations, service dashboards,
metric exemplars, deployment markers or trace breakdowns as if absent: bounded
versions already ship. Improve their reach and coverage where the tasks above
need it. Unequal-window rate comparison is also an extension of existing captured
record rates/comparisons, not a new measurement engine.

## Research basis and limits

Hosted vendor documentation and GitHub API requests were blocked by the session
proxy. The following official raw Markdown sources were successfully fetched on
November 12, 2026; links track moving branches. No vendor account, edition/pricing
comparison or hands-on vendor UX benchmark was used.

- **Datadog service page:** service summaries, resource request/latency/error
  graphs, deployments and errors grouped into issues.
  [Official source](https://github.com/DataDog/documentation/blob/master/hugo/content/en/tracing/services/service_page.md).
- **Datadog Notebooks:** investigation documents, version history and collaboration.
  [Official source](https://github.com/DataDog/documentation/blob/master/hugo/content/en/notebooks/_index.md).
- **HyperDX/ClickStack README:** property/full-text search, event deltas, alerts,
  high-cardinality dashboards, live tail and cross-signal exploration.
  [Official source](https://github.com/hyperdxio/hyperdx/blob/main/README.md).
- **Honeycomb trigger and SLO provider documentation:** thresholds, evaluation
  schedules/recipients, SLI targets and evaluation periods.
  [Triggers](https://github.com/honeycombio/terraform-provider-honeycombio/blob/main/docs/resources/trigger.md),
  [SLOs](https://github.com/honeycombio/terraform-provider-honeycombio/blob/main/docs/resources/slo.md).
  These do not verify BubbleUp's current algorithm or a multiplayer whiteboard.
- **Grafana exemplars:** metric-to-trace continuation and fast/slow trace
  investigation. [Official source](https://github.com/grafana/grafana/blob/main/docs/sources/fundamentals/exemplars/index.md).
  Fresh Explore/correlations source paths returned 404; query history and declared
  correlation patterns also draw on the repository's earlier
  [verified research](../COLLABORATIVE_INVESTIGATIONS_RESEARCH.md). They are design
  references, not a claim that today's Grafana UI or packaging was reverified.

Local checks: current capability index and active roadmap; service catalog,
service dashboards, core refinements, trace findings, Activity latency/baseline/
explanations, log patterns, recorded error triage, deployment context, comparison
windows, and corresponding current UI owners. Older gap tables were treated as
historical, not current capability inventories. This docs-only analysis ran no
application tests and makes no new runtime correctness claim.
