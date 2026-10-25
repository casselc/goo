# Counted values without record inspection

The filter editor's **Load scoped values** action uses a separate typed aggregate
response. It reads the matching count and complete values for one string field,
without fetching a record page or discovering unrelated fields. The same action
works in fixed exploration, population workspaces and paused Activity scopes.
Service dashboards continue using their existing complete exploration responses.

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

Live acceptance receipts and the executable walkthrough are recorded in
[VALIDATION.md](../VALIDATION.md) as each gate completes.
