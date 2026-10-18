# Investigation workspace

Implemented locally: goo opens at **Explore**. Four primary destinations share
one navigation rail, visual system and evidence vocabulary:

| Destination | Purpose |
| --- | --- |
| Explore | Search captured spans/logs; inspect records, distributions and cohort differences; expand traces and linked logs in a side inspector. |
| Operations | Run declared application packages; inspect operation, metric, fleet and subject evidence; explicitly save an investigation. |
| Saved investigations | Open immutable operation receipts and compare fixed-question revisions. The evidence badge distinguishes saved receipts from a currently loaded operation query. |
| Definitions | Build/import/preview/save telemetry packages over existing data; inspect application contracts, the HTTP package and proposed capabilities. |

**Examples** is a secondary destination for the original reservation workflows,
JSON editor and preview SLOs. Entering Examples starts its initial query lazily.
Navigation preserves loaded evidence and input state; it does not run telemetry
queries. The common **New query lookback** applies to newly captured questions,
not previously saved receipts or a pinned baseline's fixed time window.

New functionality should extend one of these investigation tasks. Use existing
result tabs and the trace inspector instead of adding permanent page sections.
Keep counts, provenance, unavailable states and source-plan details consistent.
The responsive layout supports small screens; result tabs support arrow/Home/End
keys, and Escape closes the inspector. This is a local lab shell, not team/workspace
administration or a complete accessibility certification.

## Distributions and a pinned baseline

1. Search an explicit environment and time window, discover fields and choose
   one breakdown. Distribution bars use complete captured group counts, never
   the bounded records page. Missing keys and present-empty strings stay distinct.
2. **Pin as baseline** retains the actual question, count, groups and source basis
   in browser memory. Add a filter or click a distribution value to refine (value drilldown requires
   AND; an existing OR question must be changed explicitly).
3. Apply the question. The baseline's exact time bounds remain fixed. Compare
   ranks the union of values by absolute percentage-point difference and displays
   both populations' counts and shares. Populations may overlap. These are
   descriptive investigation leads, not significance tests or causal attribution.
4. Scope changes (signal, environment, lookback or breakdown) clear the baseline.
   Navigation alone preserves it. Reload loses it: general exploration captures
   and baselines are **not** persisted by the saved operation receipt service.

`POST /api/compare` proxies to Go `/compare`. It accepts two bounded client-held
cohort captures. Both must use the same signal, environment, fixed window and
single breakdown field. The Go kernel validates canonical uint64 counts, unique
values, positive groups and exact population accounting. It computes shares with
rational arithmetic and rounds display strings to two decimals, without converting
counts to floating point. Results hash the actual supplied questions/groups.
Caller-supplied source bases are metadata, **not authenticated provenance**.
This endpoint does not requery ClickHouse or certify the supplied data.

An empty population returns `no_data` without normalized rows; detected disagreement
between source reads returns `unknown`. A missing category is zero only within a
complete finite captured distribution. Counts include captured records/retries,
not necessarily unique requests or business outcomes. Separate reads, late data,
sampling and retention remain limits. The current source rejects distributions
over 1,000 groups; comparison rejects partial stable group totals.

See [journey 23](journeys/README.md#23-one-investigation-workspace-from-distribution-to-trace),
[exploration source contract](EXPLORATION_DESIGN.md), and
[remaining product gaps](PRODUCT_GAPS.md).

## General-purpose boundary

The product's unit of investigation is an OTel signal, declared operation or
explicit subject relation, not a reservation. Explore questions require no application
package, reservation ID, tenant field or `goo` annotation: arbitrary retained
resource/record keys are discoverable and filterable. Shared waterfalls use
standard trace/span IDs and accept caller-supplied related actions; the reservation
link is supplied only by the Examples caller. A subject namespace and optional
tenant scope come from a package's identity contract, not an assumed order model.

Checkout and report generation are independent example packages. Journey 11
executes the real report service, discovers its declared views, follows its report
subject and expands its trace in the generic inspector. Journey 23 uses unrelated
search-index refresh telemetry without reservation attributes. New adapters and
recipes should declare source fields, units, population grain and capabilities,
then reuse these views. They should not add application-name branches in the shell,
comparison kernel or general query compiler.

Remaining coupling is explicit: Python completion/queue workflows and their SQL
templates are reservation-specific Examples. The general source adapter still
targets the pinned OTel exporter schema; native gauges, broader metric exploration,
configurable source onboarding and a shared typed measurement/query expression
layer remain unfinished. The existence of a generic shell does not imply those
backend capabilities are complete.

## Ad hoc package authoring

Explore now offers **Create package** after a captured result. The Definitions
workspace owns naming, YAML editing, validation, preview, immutable publication,
import/export and optional instrumentation handoff. The same package can run over
non-Go services without manifests. See [the contract](TELEMETRY_PACKAGES.md).

## Record inspection and query composition

Explore uses the same structured attribute viewer as trace waterfalls throughout
Operations, saved investigations and Examples. Record fields, resource attributes
and span/log attributes have separate sections. Local key/value search and scope
selection do not issue source queries. Values retain their captured type; empty
strings, nulls and missing keys have distinct meanings. Copy preserves complete
values, including whitespace and numeric-looking strings. Long and nested values
expand on demand; Raw record remains a secondary evidence disclosure.

Explore attribute actions add exact filters or select a breakdown. Filter chips
show the current predicates and can be removed individually. The field picker is
searchable and grouped by scope. Missing/exists require no value, and ordered
comparisons are offered only for duration. Changes to the composer require Apply;
a value drilldown immediately refines the last captured question with its fixed
window. OR questions require an explicit change to ALL before value drilldown.

Facet search only hides rendered categories. Shares and totals continue to use
the complete captured population, including present-empty and missing categories.
The trace inspector summarizes scope before its expandable evidence limitations;
selected spans are visible and keyboard Escape/Close returns focus to the opener.
On narrow screens, record cells and attribute rows stack while retaining labels.
See [journey 29](journeys/README.md#29-structured-inspection-and-filter-composition).

## Finding unfamiliar telemetry

Explore's quick search adds visible filters for service, literal log message and
exact severity, or minimum span duration and error status. It then offers relevant
breakdowns from available fields. The attribute composer collapses after quick
search while its filter chips and grouping remain visible. It can be reopened to
add arbitrary retained attributes. Search fields add AND predicates; changing an
input does not silently replace existing chips.

Log rows use severity badges and readable, expandable message text. Surrounding
logs opens its own inspector, with same-service time context clipped to the
original question. Scope/radius, separate evidence and paging are visible. It
never presents nearby logs as trace-linked evidence. Close/Escape returns focus;
late reads cannot reopen a closed inspector or overwrite the original results.

Find matching traces opens reported IDs from complete matching-span groups.
Span latency distribution uses Go-defined fixed duration buckets over that same
population. Selecting either suggestion captures a new breakdown; Create package
then previews, saves and reuses the same filters and grouping. See the
[discovery contract](EXPLORATION_DESIGN.md#log-discovery-time-context-and-span-duration-buckets)
and [storyboard](journeys/README.md#30-from-log-search-to-reusable-span-distributions).

The [dedicated Bombadil discovery campaign](BOMBADIL_TESTING.md#dedicated-discovery-actions-and-invariants)
checks the new exploration/package workflows with a required prefix, data
invariants and optional validation of actual actor questions through Quint.
