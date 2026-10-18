# From operation investigations to a general observability workbench

Goo currently provides declared operation packages, bounded span inspection,
raw duration and native histogram heatmaps, exact counter/error semantics,
explicit fleet coverage and bounded member comparison. Saved immutable captures
now provide local handoff/reopen/refresh. General captured span/log filtering,
record grouping and independent trace/log expansion now extend beyond packages.
This is still an investigation lab,
not a replacement for HyperDX or Honeycomb.

## Gaps and recommended order

| Capability | Current Goo boundary | Next useful behavior |
| --- | --- | --- |
| General exploration | Typed span/log filters, field discovery, one/two-field count groups, distributions, exact total/group span-duration summaries and bounded count-over-time charts | Typed arbitrary numeric measures, grouped time-series comparison and general metrics; shared AST and backend capability reports. |
| Full distributed traces | Cross-service exact-trace paging now exists within original environment/time bounds; no complete-trace guarantee or window widening | Fetch a selected trace across services with bounded pagination, explicit missing-parent/partial status; preserve async links. Keep original question scope separate from expanded context. |
| Correlated logs | Exact trace/span scopes, literal message/service/severity controls and bounded same-service surrounding logs now exist; no live tail or broader search language | Exact trace/span IDs first; declared scoped business identity next; time context separately labelled. Search/tail/pagination and log-to-trace navigation. |
| Comparative debugging | Pinned distributions, cross-window exact duration/recorded-error changes and immutable comparison pairs; no significance tests | Selected vs baseline populations, attribute distributions, sample sizes, missing fields, sampling/coverage caveats and ranked differences. BubbleUp-like differences are leads, not causal attribution. |
| Reusable workspaces | Local immutable captures, stable stored pages/catalog, reusable package questions, ordered notes with preserved revision branches and authored deployment markers | Investigation recipes, personal history, verified annotations, authorized team sharing and export. Separate live board definitions from frozen evidence snapshots. |
| APM and resource discovery | Observed service/span-name inventory, span/log signal presence and event age, reported parent-service links with unresolved/ambiguous diagnostics; declared fleet coverage remains separate | Request/root grain, DB/cache/queue conventions, declared dependencies and actual collector ingestion-health inputs. Event age cannot establish pipeline health. |
| SLOs and alerting | Legacy preview; separate typed lifecycle pilots | Full-population SLI execution; scheduler, durable evaluations, multiwindow burn rates, unknown/no-data policies, dedup/silence/routing, notifications and audit. |
| Metrics | Strict supported counter/histogram contracts | Gauges, up/down sums, exemplars, wider native metric exploration, dynamic membership and explicit common-window policies. Never average scalar percentiles. |
| Browser and exception workflows | Executable documentation journeys only | Optional RUM/web vitals, errors/stack grouping and session replay. Needs collection, privacy/redaction and storage work; browser tests are not product RUM. |
| Production operation | Loopback shared lab, local named volume, bounded source reads and per-process general-query admission/cancellation | Authentication, RBAC/tenancy, scoped source credentials, ingestion lag/drop/sampling health, retention/quotas, distributed/per-tenant query budgets, audit and backups. |
| Portable authoring | UI-authored manifest-free telemetry packages and local source-bound operation YAML; separate aggregate compiler pilot | Versioned resources, shared interactive/IaC AST, plan/apply, migration and backend capability diagnostics. Live operation queries do not yet lower to Databricks metric views. |

The first general exploration and trace/log expansion slice is implemented; see
[its design and limits](EXPLORATION_DESIGN.md). Fixed-window distribution comparison and the [cohesive workspace](PRODUCT_WORKSPACE.md)
are implemented locally. General captures and stable stored paging now exist. Reusable questions now use saved package revisions; general span measures, count-over-time charts, discovery and local query controls are described in [core basics](CORE_WORKBENCH.md). Next: broader typed measures/native metric exploration and investigation recipes. Ordered local [notes and deployment context](WORKBENCH_CONTEXT.md) now preserve exact evidence references and revision branches. Explicit [error/latency comparison pairs](ERROR_COMPARISONS.md) are now saved locally.
Production authorization and query budgets must precede shared deployment;
durable SLO/notification workflows are their own milestone. RUM/replay can remain
an integration with the existing observability platform until justified.

Goo's intended contribution is one inspectable measurement definition shared
between code-owned configuration and exploration, with explicit evidence and
portable semantics. ClickStack can continue to provide storage and HyperDX can
provide broader existing exploration. Deep links/integration are worth evaluating
before independently rebuilding every UI capability. Such integration is proposed,
not implemented or validated here.

## Research basis and limits

Reviewed on 2026-10-18:

- [HyperDX current README](https://github.com/hyperdxio/hyperdx#readme): integrated
  logs/metrics/traces/replay, schema-agnostic ClickHouse, property/full-text search,
  event deltas, alerts, high-cardinality dashboards and live tail.
- Honeycomb's published Terraform resource docs for
  [immutable queries](https://github.com/honeycombio/terraform-provider-honeycombio/blob/main/docs/resources/query.md),
  [triggers](https://github.com/honeycombio/terraform-provider-honeycombio/blob/main/docs/resources/trigger.md),
  [SLOs](https://github.com/honeycombio/terraform-provider-honeycombio/blob/main/docs/resources/slo.md)
  and [derived columns](https://github.com/honeycombio/terraform-provider-honeycombio/blob/main/docs/data-sources/derived_column.md).
  The query examples include percentile/heatmap calculations; triggers include
  schedules/recipients; SLOs specify SLIs, targets and evaluation periods.

These are a feature-direction comparison, not an exhaustive edition/pricing or
parity audit. BubbleUp-style cohort comparison is a design reference based on the
known product workflow; its current UI was not exercised. Honeycomb's full hosted
documentation was not fetched; a GitHub API attempt was proxy-denied, so public
raw repository documents were used. Optional local HyperDX remains unvalidated.

Ad hoc [telemetry packages](TELEMETRY_PACKAGES.md) now support existing telemetry
without application manifests, fresh reuse and optional instrumentation templates.
Typed measures, explicit source promotion and richer investigation documents remain
separate next increments; saving a definition does not save its preview.

See [collaborative investigation research](COLLABORATIVE_INVESTIGATIONS_RESEARCH.md)
for verified Honeycomb/Grafana/Datadog patterns and an ordered plan beyond the
[current capture foundation](GENERAL_CAPTURES.md).
