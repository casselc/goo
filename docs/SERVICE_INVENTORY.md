# Observed service inventory

Implemented locally: **Services** discovers emitting services and observed span
names without requiring goo source manifests. It provides a starting question for
ordinary OTel data from any language. Declared operation packages remain under
Operations; observing a span name does not discover a business operation contract.

A capture fixes environment and a start-time window of at most 120 minutes. The
Go ClickHouse adapter executes one read grouped by service and span name. Each
pair contains captured span count, OTel `Error` status count and eight disjoint
span-duration bucket counts using the existing half-open Go partition. There is
no inspection-row limit inside the aggregate. Total captured spans and observed
service count derive from the complete grouped result. Counts retain UInt64 string
precision, including above JavaScript's safe integer range.

The source budget is 1000 service/span-name pairs, 256 MiB scanned/memory, 8 MiB
response and 10 seconds execution. Overflow returns unavailable with no partial
inventory; narrow the time window. Go rejects malformed counts, duplicate pairs,
error counts exceeding the population and bucket partitions that disagree with
the population. Values remain typed query parameters; physical OTel columns live
in the shared adapter. The response includes its question, plan, warnings, capture
receipt and content-derived basis. A single query avoids cross-read accounting
mismatch, but does not guarantee immutable storage, complete ingestion or sampling.

Local service/name search changes only visible rows. Its displayed pair count is
separate from captured totals. Empty results mean no captured spans; no healthy,
failed or absent fleet member is inferred. Duration charts show individual spans,
including retries and duplicates, rather than root latency, request rate, unique
requests or native metric histograms. Error status is not inferred from HTTP status
attributes, log severity or business events. Kind/scope distinctions are not
separate inventory dimensions in this first slice.

**Explore spans** opens an exact service/name AND selector in the captured
environment and time bounds. It starts a separate live read with its own basis;
late arrivals may change counts. Filters are visible and removable. The existing
Create package flow builds strict YAML, previews independently and can save a
reusable definition or export an optional Go annotation template. Package preview
uses its own rolling lookback and does not archive the inventory or its evidence.

Changing environment/lookback clears inventory evidence and invalidates pending
responses; refresh clears the old result before requesting new evidence. Navigating
between destinations preserves an unchanged capture. Inventory and general Explore evidence can now be persisted through
[general captures](GENERAL_CAPTURES.md). Stored inspection cursors are stable within
a receipt. Multi-user tenancy, service dependencies, log-only service discovery
and fleet activation windows remain separate work.

[Journey 31](journeys/README.md#31-from-observed-services-to-a-reusable-question)
uses labelled OTLP fixtures through the real collector/ClickHouse path and an
independent source query. Its source population exceeds the 100-row Explore page.
It checks local-filter denominator stability, fixed-scope drilldown, preview,
mobile card layout, no data and pending-response isolation. Hegel state/swarm
checks incremental/reset accounting. An observed-ledger Quint protocol checks
capture → local search → fixed-scope open → reusable package, with rejecting
negative controls for changed totals, widened scope and changed environment.
The dedicated Bombadil actor exercises discovery/filter/clear/open plus refresh
and scope edits, with accounting, rendered-count, scope and stale-evidence invariants.
