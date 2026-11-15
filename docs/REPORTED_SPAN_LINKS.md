# Investigate reported span links

In a trace inspector, select a span and choose **Look up linked span** to inspect
an OTel link whose target is outside the returned spans. Each action shows its
link position and a short span ID; the accessible name and tooltip contain the
full trace/span pair.

The lookup stages first. Check the displayed environment and UTC window, then
choose **Run target lookup**. The query matches both reported target IDs, retains
the original environment policy and fixed half-open time window, and omits
service, namespace and population conditions. A linked target may belong to
another service. The lookup never widens the environment or time window.

**Back to trace**, or Escape, returns to the held trace and restores focus without
reading it again. Returning during a request aborts it; a late response cannot
replace the held trace. The trace inspector's close control still closes the
whole inspector. After returning, Escape follows the existing inner span-detail
then outer trace-inspector hierarchy.

If exactly one target tuple is already returned, **Follow linked span** selects
that row locally. Duplicate tuples remain separate observations and never select
an arbitrary first match. In the expanded trace inspector, the existing row
selection may separately read associated logs; target lookup itself does not
expand another trace or read its logs.

## Find referencing spans

Select a span in the expanded trace inspector and choose **Find referencing
spans**. This is available even when the selected span has no outgoing links.
**Run reference search** finds span observations reporting a link to that exact
trace/span pair. The environment policy and original fixed window remain;
service, namespace and population conditions are omitted, as disclosed before Run.

Results count **span observations**, not link entries. Repeating a link within
one span does not add another result; independently stored span observations
with identical IDs remain separate results. Both IDs must occur together at the
same link position. A trace ID in one entry and a span ID in another do not match.

**Next** and **Previous** explicitly read another page, with at most 100 rows per
page and offsets through 1,000. The displayed range describes the current page.
The matching count covers the full selected population, independently of the
page limit. Each page has its own read basis and timestamp: records can move
between fresh reads, so pages are not a frozen combined snapshot. Failed paging
retains the requested page for an explicit retry. Back returns to the original
held trace without acquiring source records.

An empty search does not establish that nobody linked to the span: sampling,
ingestion, the environment and the fixed window constrain what is observable.
See the [reference-search storyboard](journeys/span-link-incoming.md).

## Meaning and limits

- A link reports a trace/span reference, not causality, delivery, processing or a
  unique invocation. No matches means none were returned in this environment and
  window; it is not proof of failed delivery.
- Both IDs must be canonical nonzero lowercase hexadecimal values. Missing,
  invalid or misaligned link arrays produce an unavailable reference; Goo does
  not guess a trace-only target.
- The inspector offers the first 32 reported links and shows how many further
  links are not displayed. A lookup returns at most 100 span observations;
  additional matches are disclosed rather than treated as a complete result.
- Target acquisition is available in expanded record trace inspection. Other
  waterfall owners can display links and follow a unique held target, but do not
  acquire targets unless their application adapter supplies that action.
- Recursive traversal and widening the lookup window remain separate work.
  Specialized projections that cannot carry the selector refuse it; a rolling
  tail or trace-search conversion never silently drops the link condition.

## Implementation

`ui/span-link-intent.mjs` owns pure action planning; shared tuple/array admission
lives in `ui/span-link-membership.mjs`, and deliberate reference questions in
`ui/span-reference-question.mjs`. [Population conservation](SELECTOR_CONSERVATION.md)
keeps display projection changes separate from that widening.
The ordinary spans question accepts `linksTo: {traceId, spanId}` for inverse
membership; it can be combined with ordinary `where` conditions. The compiler
binds both IDs and tests correlated array positions without expanding records.
All three link arrays must align. The existing source-profile check verifies
nonnullable link-array types before any matching counts or pages are read.
`ui/span-link-membership.mjs` and the independent Go reference evaluator use the
same declared membership rules.
`ui/visualizations.mjs` presents bounded reference actions. The passive
`goo-trace-waterfall` remains a supplied-data component with no query client.
`ui/span-link-context.mjs` owns the staged panel, request lifetime, response
admission and held return; `ui/record-results.mjs` supplies its source client and
revokes it when the parent changes.

Returned questions must match the requested question and page. Every displayed
row must satisfy the exact environment policy and fixed time bounds, plus the
requested target IDs or incoming-link membership. Unrelated malformed link
entries do not suppress a different valid matching entry.
The target read retains its own basis/read timestamp; the original trace basis
and read timestamp remain separate in **Question and reported link**.

```mermaid
sequenceDiagram
  actor Operator
  participant Trace as Held trace inspector
  participant Lookup as Span-reference owner
  participant Source as Go API / ClickHouse
  Operator->>Trace: Follow link / find referencing spans
  Trace->>Lookup: Admit selected tuple and stage question
  Note over Trace,Lookup: Held origin retained; no source read
  Operator->>Lookup: Run target lookup
  Lookup->>Source: Ordinary spans question, exact target or linksTo + environment/window
  Source-->>Lookup: Target observations and independent read basis
  Operator->>Lookup: Back to trace
  Lookup->>Trace: Restore held DOM and focus
  Note over Trace,Source: No return read
```

See the [executable storyboard](journeys/span-link-target.md) and
[validation record](../VALIDATION.md) for executed checks.
