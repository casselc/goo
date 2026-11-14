# Raw gauge observations

Milestone 438 adds explicit inspection of reported gauge rows from discovered
instruments. This is a raw-observation profile, separate from evaluated
measurements. Discovery metadata alone does not contain the values.

## Read boundary

The operator selects a discovered gauge and chooses **Read observations**. The
request carries the original fixed discovery question: environment selector,
namespace, service and half-open `[fromMs, toMs)` window, plus exact instrument
`kind: gauge`, scope, scope version, scope schema URL, metric name and unit. Empty reported
scope, version, schema and unit remain exact values. No instance identity is
required to inspect rows. Native All-environments scope remains unsupported.

Opening metadata, expanding held rows or changing their local presentation does
not read values. A source-scope edit or new discovery retires the old inspector;
a late response cannot publish into a replacement context. Temporary navigation
cancels pending work while retaining the settled bank and local inspection in
the same mounted workspace; returning does not reread. See
[first-use acquisition](FIRST_USE_ACQUISITION.md). Read observations is
independent of evaluator readiness. Existing Build/Preview readiness and kernel
preflight remain unchanged.

Only the gauge source family is checked and read. The existing metadata schema
probe does not promise value-profile compatibility; unavailable value columns
refuse the explicit read without removing useful discovery metadata. Values and identities are SQL
parameters; physical source projection stays in the source adapter. Existing
10,000-row and 8 MiB transport budgets apply atomically: exceeding either refuses
the result rather than publishing a truncated sample. The serialized result
budget also includes raw-row and typed-display duplication, so a source response
below 8 MiB can still refuse if its full inspection result exceeds 8 MiB. The SQL window bounds the
source read, and each reported observation end timestamp must independently be
a canonical nanosecond integer in that window. Absent, malformed or out-of-window
end timestamps refuse the whole source response. Start zero is retained as raw metadata and displayed as not reported, rather
than a misleading 1970 start date; absent
or malformed starts remain per-row diagnostics without interval inference. Source refusal,
cancellation and malformed responses remain distinct from an empty valid read.

## Raw profile

Every returned row is an occurrence. Repeated identical exports stay repeated;
a read-local row key supports selection and rendering, never producer identity,
deduplication or continuity. The profile preserves resource and point attributes,
resource schema, instrument identity, unit, reported start/observation timestamps
and flags. Nanosecond timestamps travel as strings; the UI shows UTC with exact
reported values available in detail. Source-read time is separate from both
reported timestamps and evaluation time.

The wire carries `rawRow`, `rawValue`, finite-only `value`, `valueState`, typed
`startNS`/`endNS` and flags with field states, and `{field, code}` diagnostics.
Instrument/resource metadata is a nested `stream`; it describes reported labels,
not an established continuous stream. Values preserve their raw text when present. Flags accept exact integral JSON
numbers (including `0.0` or `1e0`) within uint32; unknown bits remain visible
diagnostics. The browser preserves a conservative invalid-flags verdict when
JSON numeric rounding prevents it from reproducing the original token. Invalid
or missing flags never establish sample validity. The actual source projects
scalar values as strings; raw scalar lexical preservation refers to that
projection, not arbitrary numeric JSON tokens. Finite, nonfinite, absent and invalid
states have explicit diagnostics. Missing or no-recorded-value data is unavailable,
never zero. A raw numeric field is not a validated measurement result. Nonfinite
values are not serialized as JSON numbers. Invalid point data stays visible where
it can be represented truthfully; malformed identity or attribute structure
refuses the read because selected scope cannot be established.

The table and local attribute detail reuse supplied-data display components.
There is no new capture carrier or store. These rows establish no fleet aggregate,
stream continuity, rate, reset, health, freshness or source completeness. A future
timeline requires its own reported-identity and sample-validity contract before
connecting any points.

## Acceptance

Exercise real OTLP gauges with and without instance identity, repeated exports,
distinct attribute maps and units, no-recorded-value and nonfinite values. Source
fixtures additionally cover absent/malformed values, budgets, typed scope refusal
and source errors. Browser checks distinguish discovery from explicit value reads,
verify held-row interactions remain local and reject obsolete responses. Executed
checks and the real-app storyboard are recorded in VALIDATION.md and the journey
index; this contract alone is not evidence that those checks passed.
