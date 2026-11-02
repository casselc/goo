# Whole-product Go/backend review — 312

Review date: 2026-11-02. Scope: incident exploration, source acquisition, certainty,
query ownership, view/template admission, portable artifacts and Go layering.
This is an independent source review with small executable reproductions, not a
benchmark or a claim that remote CI or every deployment configuration passed.

## Ranked findings and selected disposition

### 1. High: terminal pages can conceal count/read disagreement

Original location: [source.go](../../exploration/source.go), the consistency check
following `HasMore` and row trimming (approximately lines 230–235 before 312).
It detects counts smaller than the observed row minimum, and contradictory
lookahead, but misses the reverse contradiction when a terminal page has fewer
records than the independently acquired count requires.

A custom HTTP transport supplied count `5`, one row and limit `100` to the actual
`Source.FetchRecordPage`. The original result was `state=observed`, `hasMore=false`
with exactly **two SQL POSTs**. This is not a possible coherent terminal page.
The same shared acquisition path serves full exploration. A count/page mismatch
must retain diagnostics and withhold the coherent result state; it must not
silently become a different population or erase the rows.

**Selected for 312:** diagnose both directions of disagreement. Check expected
page availability and lookahead across count, offset and limit; retain valid
empty pages beyond the counted end. Include full-page/no-lookahead contradictions
as well as short and empty pages. The bounded exhaustive regression enumerates
these states through the actual source adapter, not a duplicated implementation
of its conditional. Full exploration also tests that a disagreement withholds
its summary completeness witness. A witness describes supplied captured-query
aspects, never exhaustive upstream telemetry.

### 2. Medium: full/page query metadata aliases caller-owned input

Original location: [source.go](../../exploration/source.go), `fetchProfile` entry;
[query_environment.go](../../exploration/query_environment.go#L39) only clones the
environment selector. Query value copies still share predicates, nested slices,
grouping and bindings. The reproduction changed `q.Where.Value` **after** fetching;
the returned `result.Question` changed while its already computed SQL and basis
were unchanged. HTTP handlers generally own their decoded input privately, so
this is primarily a Go library ownership/reliability defect, not an HTTP client
escape.

**Selected for 312:** freeze the whole query through the existing JSON wire
representation before compiling/reading, as already done by
[counted_values.go](../../exploration/counted_values.go#L79) and
[record_aggregates.go](../../exploration/record_aggregates.go#L58). Preserve invalid
query refusal before SQL. Test mutation during the synchronous transport callback
and after return, including nested predicates, grouping slices and the typed
environment selector. No new public shadow query type or generic clone framework
is needed.

The direct-return audit also found `InventoryQuestion.Namespace` aliases in
[inventory.go](../../exploration/inventory.go) and
[service_discovery.go](../../exploration/service_discovery.go). Both now use the
existing nil-safe `Namespace.Clone()` before compilation. Tests exercise their
actual empty source responses and confirm one and two SQL POSTs respectively.

Other audited paths: counted values and record aggregates already detach the
full question; trace-search returned requests use a detached copy; pod discovery
and environment adoption question structures are value-only. Tail's query
constructor shares a predicate pointer, but its poll returns rows rather than a
question receipt; a broader constructor rewrite was not selected for this round.

### 3. Low, measurement task: budget acquisition aspects rather than API calls

The source code establishes the statement counts below. It does **not** establish
latency, scanned bytes or an optimizer bottleneck. A statement can contain joins
or unions; statement count is not the number of physical scans. There is no
measured basis here for parallelizing reads, changing indexes or introducing a
new query engine.

| Operation | Source SQL POSTs | Boundaries |
| --- | ---: | --- |
| Ordinary `/explore` | 3 base | count, bounded rows, field discovery |
| Grouped `/explore` | +1 | full group accounting |
| Bucketed `/explore` | +1 | unsplit timeline |
| Split-time `/explore` | +1 more | grouped time series; existing projections still apply |
| `/explore/page` | 2 | matching count plus bounded rows; no groups/fields/time |
| `/explore/counted-values` | 2 | matching count plus complete bounded categorical groups |
| `/record-aggregates` | 2, or 3 with timeline | count, groups, optional time; no rows/field discovery |
| `/environment-scopes` | 1 | combined discovery statement |
| `/environment-scopes/records` | 1 | bounded adoption preview |
| `/inventory` | 1 | service/span-name aggregate |
| `/discovery` | 2 | observed signals and reported-parent edges |
| `/trace-search` | 1 | aggregate reported trace search |
| Tail poll | 1 per poll | bounded rolling event-time records |
| `/compare`, template expansion, artifact codec | 0 | supplied-data/pure configuration work |
| Ordinary view save/open | 0 | local catalog work; optional captured references may read stored files |

References: [source.go](../../exploration/source.go#L136),
[compiler.go](../../exploration/compiler.go#L150),
[counted_values.go](../../exploration/counted_values.go#L73),
[record_aggregates.go](../../exploration/record_aggregates.go#L52),
[service_discovery.go](../../exploration/service_discovery.go#L79).
An ordinary grouped HTTP service Run therefore has four source statements;
subsequent explicit method **Read values** adds two. Template expansion and
mounting the facet add none. These are source-code accounting facts, not a timed
whole-browser campaign measurement.

**Recommendation:** keep native journey HTTP budgets, but add source statement
and acquired-aspect accounting to the highest-value incident journeys before
selecting optimization work. Field discovery currently uses `mapKeys`/`arrayJoin`
over the matching scope ([compiler.go](../../exploration/compiler.go#L184)); measure
its actual cost before splitting or caching it. Preserve separate read clocks
and diagnostic states when optimizing; concurrent SQL is not a snapshot.

## Boundaries worth preserving

- Query HTTP admission has a 60-second request ceiling and cancellation ownership
  ([budget.go](../../internal/queryhttp/budget.go#L97)); source statements carry
  10-second execution, memory/read limits and bounded 8 MiB responses
  ([source.go](../../exploration/source.go#L74)). These are configured limits, not
  measurements of actual resource consumption.
- Narrow pages do not fabricate aggregate/field/timeline snapshots. Full/source
  witnesses are separate from browser-held cohort accounting and annotation
  custody. Independent selectors and typed missing/empty values remain explicit.
- View role and count-operand unions are closed. Fixed-query pairs retain checked
  bounds and reject presentation reuse; captured references stay distinct.
  Literal facet fields survive category-template rebinding without selected
  values or source requests travelling with them.
- Shipped numeric HTTP Customize keeps its exact query and data-only composition;
  ordinary HTTP gets the method facet. Fresh child slices and JSON-detached
  expansion protect later shipped defaults from mutation.
- YAML uses a bounded AST bridge into existing strict JSON/view validators, with
  aliases, duplicate keys, unsafe tags, multiple documents and integer overflow
  refused. CLI import has an explicit destination, no redirects/retries, bounded
  receipts, and verifies the full submitted view and its shared content ID.

## Go structure and modernization

The project already uses Go 1.27, small inferred generic route registration,
`slices.Clone`, `context.AfterFunc`, bounded readers and ordinary typed structs.
There is no evidence that users need advanced generic machinery, new typed state
proof wrappers, or a second evaluator/planner to solve this review's findings.

The useful immediate consistency improvement is query ownership at acquisition
boundaries, implemented with the existing wire representation. Repeated
validation/SQL construction across nested public validators is a maintainability
observation, not a proven performance defect: keep independently callable
validation safe and profile before adding validated-query tokens or caches.
Separate transport, source-family semantics, saved configuration and display
admission remain sensible layers; avoid moving source access into controls or
portable declarations.

## Validation and next work

The initial reproduction ran the actual narrow source adapter against an
in-process custom HTTP transport; it performed two SQL POSTs and reproduced both
findings above. The selected fixes have focused executable adapter tests,
including 972 bounded exhaustive page/count/offset/lookahead combinations, without
requiring a Databricks instance, Docker or a network source. The changed exploration
package passes in 6.550 seconds after repairing positive synthetic fixtures to
honor the actual SQL row/lookahead bounds; their existing aggregate disagreement
controls remain intact. Full-suite authority and the final roadmap
remain with the root agent; no full suites were repeated for the read-only review.

Prioritize the two selected correctness fixes and the whole-product entry-flow
issues from the parallel design/architecture reviews before more shipped widgets.
The 310 blanket facet-template exclusion is superseded by 311's validated ordinary
category-template contract. Fixed-query reuse refusal and facet/comparison
exclusion remain deliberate current boundaries, not accidental missing support.

Older capture-bank isolation and prepared-plan ownership proposals are not solved
by freezing an acquisition query. Revalidate those proposals against current bank
and preparation consumers before adding them to a later implementation slice.
