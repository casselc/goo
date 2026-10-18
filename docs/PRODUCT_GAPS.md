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
| General exploration | Bounded general span/log attribute builder and one/two-field count groups now exist; complete one-field distributions now exist; no broader chart algebra | Discover fields/types/cardinality; typed attribute filters, AND/OR, full-text log search; arbitrary bounded grouping and aggregate charts. Shared Go query AST/compiler and explain, source-specific physical adapters. |
| Full distributed traces | Cross-service exact-trace paging now exists within original environment/time bounds; no complete-trace guarantee or window widening | Fetch a selected trace across services with bounded pagination, explicit missing-parent/partial status; preserve async links. Keep original question scope separate from expanded context. |
| Correlated logs | Exact trace/span log scopes and literal body search now exist; no live tail or broader search language | Exact trace/span IDs first; declared scoped business identity next; time context separately labelled. Search/tail/pagination and log-to-trace navigation. |
| Comparative debugging | Pinned fixed-window cohort distributions plus interval/member context; no significance tests | Selected vs baseline populations, attribute distributions, sample sizes, missing fields, sampling/coverage caveats and ranked differences. BubbleUp-like differences are leads, not causal attribution. |
| Reusable workspaces | Local immutable captures by ID and URL fragment | Named saved queries, boards, annotations, deployment markers, history, team sharing and export. Separate live board definitions from frozen evidence snapshots. |
| APM and resource discovery | Declared membership and bounded OTel profile recognition | Service/operation inventory, dependencies, DB/cache/queue conventions; activation windows and freshness. Do not infer healthy membership from observed-only discovery. |
| SLOs and alerting | Legacy preview; separate typed lifecycle pilots | Full-population SLI execution; scheduler, durable evaluations, multiwindow burn rates, unknown/no-data policies, dedup/silence/routing, notifications and audit. |
| Metrics | Strict supported counter/histogram contracts | Gauges, up/down sums, exemplars, wider native metric exploration, dynamic membership and explicit common-window policies. Never average scalar percentiles. |
| Browser and exception workflows | Executable documentation journeys only | Optional RUM/web vitals, errors/stack grouping and session replay. Needs collection, privacy/redaction and storage work; browser tests are not product RUM. |
| Production operation | Loopback shared lab, bounded queries, local named volume | Authentication, RBAC/tenancy, scoped read credentials, ingestion lag/drop/sampling health, retention/quotas, cancellation/concurrency budgets, audit and backups. |
| Portable authoring | Local YAML packages; separate ClickHouse/Databricks aggregate compiler pilot | Versioned resources, shared interactive/IaC AST, plan/apply, migration and backend capability diagnostics. Live operation queries do not yet lower to Databricks metric views. |

The first general exploration and trace/log expansion slice is implemented; see
[its design and limits](EXPLORATION_DESIGN.md). Fixed-window distribution comparison and the [cohesive workspace](PRODUCT_WORKSPACE.md)
are implemented locally. Next: stable capture/paging, persisted general exploration
questions/baselines and reusable boards.
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
