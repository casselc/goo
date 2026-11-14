# 438 raw gauge observations: UX, accessibility and web standards review

Status: final production source and six real responsive runtime frames accepted.
Complete browser/source/model acceptance is owned by the parent integration run.
This review does not claim assistive-technology or cross-engine testing.

## Entry and intent

Keep **Read observations** beside the instrument's existing Build action. A Gauge
with missing instance identity can still be inspected; a disabled evaluator must
not disable the independent raw read. Qualify repeated action names with the
instrument name in accessible text. Unsupported instrument kinds need a concise
visible reason rather than an action that fails after reading.

Metadata disclosure, local Find, Inspect and paging must not acquire data. Reading
is an explicit native button action. Keep the selected instrument, environment
and fixed read bounds visible when observations arrive; do not imply a live tail.
Do not introduce a second scope editor or change evaluator readiness.

## Table and detail

Use shared compact table, UTC timestamp, scalar and attribute components. Lead
with observation time and reported value/unit; keep full attributes and exact
values in local detail. At narrow widths, use the established stacked table
layout with column labels, clear row separation and an easily reached Inspect
action. Long metric names and attribute values wrap without expanding the page.

Show missing instance identity as a reported-context gap, not a synthesized
stream name. Preserve repeated observations as separate inspectable occurrences.
Local row identifiers identify rows only. Do not connect points, group away
duplicates, infer health or advertise completeness, freshness or continuity.

Unavailable and invalid values remain explicit text, distinct from numeric zero.
Reported units remain attached to values; omitted units are clearly unspecified.
Do not convert unknown units. Display UTC dates/times and expose exact source
precision in detail; raw nanoseconds alone are insufficient. Label source-read
time separately from the observation timestamp. Limit messages describe the
held batch, not an exact total unless the source actually supplies that total.

## Interaction and failure

Use native buttons, labels, details and tables with captions and column headers.
Do not make every value a tab stop. Full value/timestamp information must be
available through keyboard-operable detail, not only pointer hover. Find needs
a persistent label and a local-results status. No-match differs from no source
observations, failed read and source-unavailable states.

Pending reads expose busy state and prevent accidental repeated execution.
Cancel/departure must retire publication and focus authority, even if transport
ignores abort. A later result must not steal focus or overwrite a newer owner.
Inspect and Find remain local; any refresh is another explicitly named read.

Errors provide a concise safe reason and retain the previous held batch with
its original read clock. Avoid raw SQL, credentials or stack traces. No broader
scope, different instrument or local-page fallback is implicit on failure.

## Final acceptance required

Review actual production source and 1440/390/320 runtime frames for density,
readable UTC/units, honest missing/invalid states, duplicates, local detail and
Find, source limits/refusal, and native keyboard focus. Exercise raw reads with
and without instance identity while confirming evaluator preflight still
refuses missing identity. Parent-owned integration must prove which actions
execute SQL and that held-row interaction does not.

## Final review

Reviewed final catalog integration and raw application adapter independently.
Read observations remains available for Gauges whose evaluator Build is disabled.
Its accessible name includes the exact instrument; the read uses the retained
discovery question. Shared UTC/scalar/attribute/table/inspector components are
used without treating raw rows as evaluated series.

Review findings were addressed: omitted units are explicit; local pages hold at
most 100 rendered rows and Find searches the complete held batch; explicit Close
and Cancel restore the eligible Read opener. A focused explicit Read follows
the pending observation heading immediately, while response completion never
moves focus. Passive retirement does not restore focus. The focused raw UI test
module executed successfully during this independent review; parent records
the final full-suite counts.

Visually inspected real application table and inspection frames at **1440, 390
and 320 px** in [journey screenshots](../journeys/screenshots/). The final
replacement inspection frames show **Not reported · raw 0** for an OTLP zero
start sentinel, preserving raw detail without inventing a 1970 start. Desktop
rows are compact; mobile rows retain their visible column labels and reachable
Inspect actions. Duplicate occurrences, reported/missing instances and absent
values are distinct. UTC observation time and the source read clock remain
separate. Exact precision wraps at narrow widths without horizontal overflow;
the existing bounded inspector scrolls to full attributes and raw detail.

The initial screenshots showed only the catalog and were not accepted as proof
of the inspector. Final targeted table/detail frames replace those images.
No blocking UX, accessibility source or web-standards findings remain for this
bounded slice. No physical screen-reader or cross-engine result is claimed.
