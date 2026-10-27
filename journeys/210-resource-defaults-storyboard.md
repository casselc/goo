# Journey 210 — open Resources and see usage

Status: source and initial live main checks pass; final acceptance in progress.
[Contract](../docs/AUTOMATIC_RESOURCE_DEFAULTS.md).

1. Discover ordinary OTel resource inventory for a fixed environment/window.
2. Open a literal container Resources row. Without source setup or Load clicks,
   see memory and CPU usage histories; the compact overview omits unrequested
   capacity rows. Initial cost: one inventory read, one metadata read and at
   most two usage reads, plus pure plans/compositions/histories.
3. Expand memory capacity with the keyboard. Its request/limit signals load
   while the original usage receipt and CPU history remain held. Collapse and
   reopen without new reads.
4. Inspect a point, stage matching logs and return to the exact held panel with
   focus retained. Staging/return remain source-free.
5. Open a resource with two memory usage streams and one CPU stream. CPU loads;
   memory requires an exact choice. Incomplete discovery similarly stays manual.
6. Hold real default replies: withdraw and release to prove no late history or
   stale capacity reads; open capacity while usage is pending to preserve intent
   without duplicating usage. Generated Hegel schedules exercise actual owner
   code; finite Quint replay checks observed effect budgets with negative traces.
7. Run the updated Bombadil actor through automatic usage, capacity and context
   return. Capture current desktop/mobile frames from the live app.

Prepared OTLP proves collector/ClickHouse/app integration, not deployed receiver
agents or source authenticity. No arbitrary concurrency proof is claimed.
