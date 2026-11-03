# Read only what a fixed comparison baseline uses

A saved fixed-query comparison still stores two questions and no observations.
Opening stages them; Run reads them independently. The selected population keeps
its full record read so you can inspect its rows and fields.

For supported string groupings, the fixed baseline now uses the existing aggregate
endpoint: count, groups, optional duration and timeline, with their independent
read receipts. It does not read baseline records or discover their fields. **Scope
and evidence** shows that acquisition and its actual receipts and warnings.
Numeric column groupings (`duration_ns`, `timestamp_ns`, `severity_number`) retain
the full path. The choice happens before the request, with no retry/fallback after
an error. Query predicates, environment policy, bounds, grouping and measures are
preserved.

This is aggregate certainty, not a full-page witness. Independent aggregate reads
can disagree and remain diagnostic. An aggregate-only baseline cannot check a
contradiction between its count and an unacquired record page. No full snapshot,
row membership or upstream coverage guarantee is created. Captured baselines,
held-parent comparisons and temporary selected-versus-rest acquisition are unchanged.
Retiring the owner aborts its request and rejects late publication.

## Measured task

On the exact GUI-authored baseline question from the [314 journey](journeys/314-source-demand.md),
four alternating actual ClickHouse reads retained count 100, one group, duration
and aggregate fingerprint. The baseline changed as follows:

| Measurement | Full | Aggregate |
| --- | ---: | ---: |
| SQL statements | 4 | 2 |
| Consumed source response bytes | 73,680 | 258 |
| Executor DTO bytes | 88,891 | 1,767 |

The two removed reads are rows and field discovery. With a requested timeline,
transport tests cover five statements becoming three without stripping that
projection. These executor measurements are separate from browser HTTP envelopes;
HTTP request count need not fall. Timings and physical scanned rows varied, so no
latency SLA or universal speedup is inferred. There is no cache, new query engine,
SQL rewrite or new API endpoint.

Shared strict aggregate DTO validation serves both service dashboards and this
baseline owner. It rejects snapshot/inspection slots and unrequested receipts;
validating JSON grants no display-snapshot authority.
