# Ad hoc telemetry grammar packages

Status: implemented locally. An application does not need a goo manifest or Go
instrumentation to use this path. A package describes a reusable selector and
captured-record count over existing OTel spans or logs. It uses the same Go query
AST, validation and ClickHouse adapter as Explore.

## Author through investigation

1. In **Explore**, select an environment, filter any retained fields and optionally
   choose a breakdown. Run the question, then **Create package**. The action is
   disabled while results are unavailable or superseded: it uses the captured
   question, not unexecuted form inputs or the current row page as a denominator.
2. In **Definitions → Telemetry packages**, name the package, choose a default
   lookback and **Build package**. The original absolute window remains preview
   context. Its timestamps, row offset, row limit and data basis are not saved
   in the reusable definition.
3. Inspect/edit YAML, **Validate definition**, then **Preview captured data**.
   Validation produces the parameterized source plan without reading telemetry.
   Preview separately runs the fixed question, showing exact count/group results,
   source state and an independent evidence basis. Empty remains `no_data`;
   disagreement remains `changed_between_reads`; query failure is an error.
4. **Save package revision** publishes immutable configuration in the local named
   volume. The same content is idempotent; edits create another revision. The old
   revision remains readable. The library supports up to 200 listed revisions;
   it fails explicitly beyond that rather than hiding entries. Revision lookup
   remains available. There is no mutable “latest” alias or automatic deployment.
5. **Use in Explore**, or the library's **Explore** button, uses the saved selector
   with a new live window from its default lookback. The source question/result
   basis stays separate from the package revision. Additional filters AND with
   the retained selector, including imported nested predicates. Two imported
   breakdowns remain intact. Clear filters or change signal/environment detaches
   the selector; editing the live query does not edit stored configuration.
6. **Export YAML** and **Import YAML** make definitions portable and reviewable in
   source control. Imports use the same strict schema/compiler. Filters and literal
   values are included in exports; raw telemetry rows are not.

Example (ordinary Python/Java/Node/etc. OTel telemetry):

```yaml
apiVersion: goo/telemetry-package/v1alpha1
name: search-refreshes
title: Search refreshes
source:
  adapter: clickhouse.otel/v1
  signal: spans
  environment: production
where:
  field: {kind: column, key: service}
  op: eq
  value: search-indexer
groupBy:
  - {kind: attribute, key: region}
measure: captured-record-count
lookbackMinutes: 15
views: [records, distribution, trace-context]
```

`signal` also supports logs. Logical fields are the existing `column`, `resource`
and `attribute` bindings. Nested AND/OR, operators, bounds and missing/empty semantics
are inherited from [Explore](EXPLORATION_DESIGN.md). YAML uses the same camelCase
field names as JSON. Unknown fields, multiple documents, aliases/merges, invalid
source/measure/view combinations and excessive depth/size are rejected. Packages
are at most 32 KiB; JSON authoring requests are capped at 64 KiB. Names are bounded
ASCII identifiers; titles/descriptions support bounded UTF-8. Lookback is 1–120
minutes. Zero/two breakdowns support records; distribution requires exactly one.
Views select preferred entry points in the shared Explorer, not restrictions on
later exploratory changes.

## Optional instrumentation handoff

Any OTel-producing language can use a package directly without changing its source.
The handoff lists logical fields referenced by the definition. Missing keys remain
meaningful; listing a field does not establish that every selected record has it.
Deployment environment and any selector values must continue to match the source
binding. The package does not infer a complete instrumentation contract from samples.

For Go, explicitly choose an operation name and role. The UI generates and can
download a source template using the actual `//goo:operation` syntax, a named
`context.Context` parameter and named error return. Tests pass that generated code
through the real annotation scanner. This is a reviewable starting point, not an
automatic source patch, file/function locator, aggregate annotation or equivalent
replacement for the observed instrumentation. Required attributes still need
explicit implementation. Weaving/adopting it can add new spans and duplicate
instrumentation; review bindings and coverage before switching selectors.

No Go template is required for non-Go services. No annotations are inserted or
source files changed by the UI. Operation manifests/YAML remain a separate,
stronger source-checked contract for the existing weave path.

## API and persistence

| Endpoint | Behavior |
| --- | --- |
| `POST /api/telemetry-packages/draft` | Validate captured question + name/title/lookback; derive a reusable definition and fixed preview plan. |
| `POST /api/telemetry-packages/plan` | Accept YAML **or** definition; optional absolute preview bounds and explicit annotation target; return canonical YAML, definition revision, source plan, handoff and caveats. With no bounds, freeze a new window. No source execution. |
| `POST /api/telemetry-packages` | Save validated YAML as an immutable definition; no telemetry query. |
| `GET /api/telemetry-packages` | List bounded local saved definitions/revisions. |
| `GET /api/telemetry-packages/{revision}` | Verify and load a content-addressed definition. |

The Python proxy forwards to matching Go paths without `/api`. Definitions live
under `<investigations-dir>/telemetry-packages`, retained by the existing named
volume. Publication uses synchronized temporary files and atomic non-overwriting
links; loads validate both schema and digest. The revision hashes the normalized
JSON definition, not YAML formatting, captured evidence, preview window or an
instrumentation template. No revision claims source authenticity or complete data.

Saved operation investigations still archive evidence; these packages store
configuration. Browser baselines remain volatile. These are separate concepts.
The lab Go service still shares startup configuration with bundled operation/HTTP
packages; a standalone general-source server deployment is not introduced here.
The target application's source and manifest are not consulted by this authoring
or execution path.

## Boundaries and next increments

This bounded package supports **captured-record-count**, rows, one-field distribution
and exact trace context; counts include retries/exporter duplicates. It does not
infer business subjects, outcomes, SLIs/SLOs, durable events, native metrics or
coverage from observed spans. Source adapter support is the pinned OTel exporter
schema, not arbitrary SQL or arbitrary ClickHouse tables.

Next: share typed measure/source/grain contracts with the aggregate compiler;
add declared duration/error measures and field/coverage diagnostics; introduce
reviewed operation-package promotion and source mappings rather than guessing
functions. CEL/KCL calculations, warehouse lowering of this package kind, backend
resource deployment, named registries, collaboration and production authorization
remain future work. Existing Databricks lowering is a separate aggregate artifact
pilot and is not claimed for these raw-record packages.

See [journeys 24–25](journeys/README.md#24-build-a-package-over-existing-python-telemetry)
and [architecture diagram 12](ARCHITECTURE_DIAGRAMS.md#12-ad-hoc-telemetry-packages-and-optional-instrumentation).

## Reusable core measures

Packages now retain optional `summary: span-duration` and `bucketMs` settings from
Explore. The same package revisions appear in Explore's Reusable questions picker,
where running resolves a fresh default-lookback window. Summary/chart semantics
and supported widths are in [core basics](CORE_WORKBENCH.md). Definition validation
rejects log-duration summaries and default lookbacks exceeding 240 chart intervals.
Saving a package still stores configuration; new general captures separately retain
the resulting measures and chart points.
