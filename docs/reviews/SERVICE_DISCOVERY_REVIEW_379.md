# Service discovery review — 379

Independent source review found two concrete problems and rereviewed their fixes.
Runtime acceptance is recorded separately in [validation](../../VALIDATION.md).

The shared optional-table decoder must not collapse duplicate JSON keys: a probe
containing both `result: 1` and `result: 0` is malformed, not table absence. It now
consumes exactly one result field with a canonical numeric 0/1, the object end,
and EOF. Both duplicate orders, extra fields and trailing data are negative tests.
The helper shares only availability syntax, leaving source transport and budgets
in their owners. It does not preempt the planned bounded-decoder consolidation.

Catalog row actions must retain the scope that produced them. An old callback
could previously open an old service name under a newly selected scope. Open
callbacks now bind generation, rendered revision and held evidence, and require
a connected, contained, visible, non-inert origin. Search redraws and scope edits
retire them. The same guard covers declared-only and distributed-view entry.

The compact Scope disclosure preserves the default recent flow. Fixed UTC bounds
reject non-Z offsets and rolled dates; all, missing and reported-empty namespaces
remain distinct. Scope edits cancel pending catalog/service owners; global
lookback edits leave an explicitly fixed request alone. Discover remains explicit.

The four-family union preserves exported-point grain, one global service/signal
cap and span-only dependency evidence. Observed exponential instruments receive
no executable definition. Missing optional source, successfully empty matching
data and failed source operation remain separate outcomes. No remaining source
review blocker was found after the two fixes.

The first actual desktop frame also exposed an inherited catalog layout problem:
long action text squeezed headers and inline namespace card styling overlapped
the service name. The catalog now uses the existing compact/stacked `goo-table`,
scoped identity stacking, native column scopes/caption and a short visible action
with the full accessible name. The refreshed desktop frame keeps every column
readable; this is a display fix within the same scoped entry, not another widget.
