# Counted values without record inspection

The filter editor's **Load scoped values** action uses a separate typed aggregate
response. It reads the matching count and complete values for one string field,
without fetching a record page or discovering unrelated fields. The same action
works in fixed exploration, population workspaces and paused Activity scopes.
[Service widgets](SERVICE_RECORD_AGGREGATES.md) have their own richer projection;
the service browser supporting read and capture paths retain full exploration.

`POST /explore/counted-values` (UI proxy `/api/explore/counted-values`) accepts the
existing exact exploration question. It requires one string `groupBy`, offset zero
and no requested measure or series projections. A numeric **predicate** remains
valid and applies to both statements; this restriction does not discard filters.
Existing validation, source adapters, parameter binding and reported-edge scope
apply. `limit` remains part of question identity but never limits group totals.

The response is `goo/counted-values/v1`: question, state, exact decimal-string
count, nullable `g0`/count groups, capturedAt, basis and independent count/groups
receipts (`readAt`, `basis`). There are no fake inspection rows, fields, page flags,
result token or capture receipt. Per-read bases hash the raw responses. The combined
basis includes question and content while excluding observation clocks. A matching
basis is not an atomic-snapshot or telemetry-completeness guarantee.

```mermaid
sequenceDiagram
  participant Person
  participant Editor as Filter editor owner
  participant API as Counted-values API
  participant CH as ClickHouse
  Person->>Editor: Load scoped values
  Editor->>API: Exact fixed question + query ID
  API->>CH: Count matching records
  CH-->>API: Count
  API->>CH: Complete groups (1001 sentinel)
  CH-->>API: Missing / empty / reported values
  API-->>Editor: Versioned aggregate + two receipts
  Editor->>Editor: Check scope, owner and conservation
  Person->>Editor: Stage a value
  Note over Editor: Local edit; no source read
```

Both statements execute even if count is zero. More than 1,000 groups refuses the
whole lookup, and malformed, duplicate, zero-count or impossible-null groups fail.
Missing map keys and present empty strings are distinct. A count/group sum mismatch
returns `changed_between_reads` with diagnostic observations; the UI refuses to
publish usable values. A valid empty lookup is `no_data`, not source health.

The action retains existing pause, request ownership, exact scope identity,
cancellation and late-result rejection. Opening/disclosing/searching/staging loaded
values performs no additional telemetry read. Scope changes invalidate loaded
values. Source failures clear old values and require another explicit action.
The route shares admission, query IDs and cancellation, strict 16 KiB decoding and
a 55-second enclosing deadline. Existing per-statement source budgets also apply.

`/explore`, its serialization, and stored exploration/capture IDs and bytes remain
unchanged. An aggregate response cannot substitute for a complete exploration
result or be promoted into an old capture. Choosing a counted value stages the
normal question; subsequent exploration and capture use their existing owners.

## Measurement and validation

The [initial read audit](AGGREGATE_READ_AUDIT.md) documented four actual statements:
count, rows, groups and fields. `cmd/goo-read-audit` now accepts
`GOO_AUDIT_MODE=counted-values` (default `explore`) to compare the same fixed query.
It records source work separately from decoded response bytes and elapsed time.
Removing reads does not make the remaining count/group scans cheap, and modest
fixture timings do not establish production latency distributions.

The [executable walkthrough and actual frames](../journeys/193-counted-values-projection-storyboard.md)
cover full-scope values, exact numeric filters, local staging and stale publication.
[VALIDATION.md](../VALIDATION.md) records the separate source, live and actor gates.

## Controlled source comparison

With only the owned lab traffic generator stopped, the final tool ran each fixed
[baseline query](AGGREGATE_READ_AUDIT.md) first through legacy execution and then
through counted-values execution. No fixture was added or changed during this
window. Both cases matched the same 123 span records; the dimensions retained four
namespace groups versus 123 span-name groups. Count and group SQL hashes and
consumed response sizes matched exactly before/after.

| Case | Statements before → after | Serialized result bytes before → after | Summed CH read rows before → after | Summed CH read bytes before → after | Elapsed ms before → after |
| --- | --- | --- | --- | --- | --- |
| small | 4 → 2 | 5819 → 766 | 24708 → 12354 | 4829863 → 1595728 | 140.397 → 48.336 |
| larger | 4 → 2 | 9472 → 4508 | 24708 → 12354 | 4836040 → 1601905 | 83.391 → 47.768 |

Receipts: [comparison manifest](evidence/aggregate-read-audit/comparison.json),
[small legacy](evidence/aggregate-read-audit/small-before.json),
[small projected](evidence/aggregate-read-audit/small-after.json),
[larger legacy](evidence/aggregate-read-audit/larger-before.json), and
[larger projected](evidence/aggregate-read-audit/larger-after.json). Source commit
`eae70ab`; binary provenance and traffic state are recorded in the manifest and
local deployment receipt. The generator's prior running state is restored after
browser acceptance.

Each statement reported 6,177 physically read rows, much more than the 123
matching records. Sums count work across statements, not distinct records. These
are two modest dimensions and one observation per mode, not a cardinality-only
experiment, a large-cardinality benchmark or a latency distribution. Source/header
measurements and serialized executor-result bytes differ from network traffic and
HTTP envelopes. The remaining two aggregates still have normal scan/memory costs.
