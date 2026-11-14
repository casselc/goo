# Roadmap after measured operator consolidation

The finite 427–429 cycle covers exact definition recall, paired evidence and
measured scope/preservation controls. Executed results belong in
[validation](../VALIDATION.md), not in this proposal's acceptance criteria.
The [architecture checkpoint](reviews/PRODUCT_CHECKPOINT_AFTER_429_ARCH.md) and
[product/UX checkpoint](reviews/PRODUCT_CHECKPOINT_AFTER_429_UX.md) review the
broader product rather than extend another preservation queue.

Status: **429 delivered and validated; the next cycle is adopted** after the
independent architecture/product cross-review.
The next cycle is **437 breadth → 438 depth → 439 consolidation**.
These outcomes are planned, not implemented. The existing Go runtime, source
adapters, typed questions and supplied-data components remain the foundation.

## 437 — explicit All-environments log entry

An operator who does not know environment labels should be able to explicitly
start a bounded log stream across the configured workspace source. Current named,
empty and unreported scopes are real distinctions; none means All. Blank input or
a failed discovery must never silently broaden the scope.

Design a typed All scope before implementing the entry. Check its meaning through
AST validation, source lowering, links, saved definitions, captures and held return.
Retain fixed bounds, limits and source refusal. Environment selectors are scope,
not a substitute for source authorization. All stays inside the configured
workspace source; this does not introduce multi-user authorization or tenancy.

Start with logs. Show reported environment facts per row and preserve missing
versus empty labels. Narrowing and row-context actions must retain or explicitly
narrow the parent scope; unsupported routes refuse before a read. Do not infer a
replacement query from a display label or treat a named scope as All.

Acceptance: named/empty/unreported data in one source, explicit opt-in, typed
links/recall, source/access refusal, bounded high-cardinality reads, local narrowing
and Return, desktop/320/390 and keyboard. Measure requests and source work. Mixed
log/span streaming needs its own per-family watermarks and failure policy later.

## 438 — inspect raw gauge observations before evaluation

Ad-hoc metric producers can lack service.instance.id. Discovery currently shows
identity labels and counts, while service evaluation correctly refuses known
missing identity. The evaluated data-first view is downstream of that refusal;
it is not a raw-value inspection path.

Add a bounded, explicit, table-first raw gauge read from discovery. Define a
raw-observation profile rather than constructing an evaluated GaugeResult.
Preserve reported attributes, duplicate occurrences, units, timestamps and
diagnostics. Local row keys identify occurrences, not invented producer IDs.
Missing values remain unavailable, not zero. No fleet aggregation, continuity,
rate, reset, health, freshness or completeness claim follows from these rows.

Keep evaluator preflight intact. A timeline is permitted only after its reported
identity and sample-validity contract is established; missing identity does not
authorize connecting points or merging streams. Do not add a new capture carrier
to make this inspector fit. Reuse supplied-data table/detail components.

Acceptance: real OTLP gauges with and without instance identity, repeated rows,
multiple attribute sets/units, malformed or absent values, source limits and
refusal. Opening metadata or changing a local display cannot read values;
Read observations explicitly does; inspecting held rows stays local. Keep
source-read time distinct from observation time.

## 439 — measure first-use acquisition and consolidate matching work

Measure discovery → explicit read → inspect → retained return over the new paths.
Record cold and retained requests, per-family SQL/bytes, cancellation, row work
and high-cardinality behavior before proposing caching or refactoring. Compare
the same question and read state; rolling views need separate clock-aware checks.

Consolidate only duplicated pure projections or probe work with identical source,
scope, completeness and lifetime contracts. A schema probe is not evidence of
fresh data. No durable readiness cache without an invalidation contract; no
universal identity or permission layer. Keep source and native evaluator owners
separate. Refresh executable storyboards, responsive frames and native tests.

## Retained follow-ons

Presence-safe qualified CEL attributes remain valuable. Visual attribute queries
already work, so the immediate priority is reaching observations that currently
have no value path. A later CEL extension must distinguish absent, empty, invalid
and typed values using the existing predicate AST and source adapters; it must
not add another query executor or an unbounded extension registry.

Combined log/span Activity needs explicit independent polling, ordering, loss and
partial-failure semantics. Monitoring delivery, multi-user collaboration,
arbitrary executable kinds and external warehouse/storage integration retain
their separate design and operational gates. Existing dashboards, waterfalls,
templates, annotations and captures are delivered capabilities, not new gaps.
