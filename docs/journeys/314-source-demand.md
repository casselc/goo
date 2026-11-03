# Measure one investigation's source demand — 314

This bounded experiment follows real ingested HTTP telemetry through Service
Customize, an explicit ordinary-count projection, method facet selection and
selected-versus-rest comparison. The selected cohort has seven POST spans; its
complement has 100 GET spans. Counts describe the full matching population even
though the parent inspection page holds only 100 of 107 spans.

Promote the rest comparison using Include comparison without facet, save its
fixed selected/baseline queries, then Run explicitly. Inspect the timeout trace
and retire one outstanding actual trace read. Local staging and retirement must
not publish late results or issue automatic replacement reads.

[Maintained browser measurement](../../journeys/314-source-demand.spec.js) reuses
the real OTel fixture and standard traffic-pausing wrapper from
[313](313-active-canvas.md); its existing screenshots describe the unchanged UI.

The measurement records exact HTTP request bodies, response bytes, header/body
elapsed times and checkpoint budgets. An artifact freezes both fixed comparison
queries for a separate backend audit. HTTP actions are not SQL statement counts;
this test does not attribute ClickHouse statements to individual UI actions.
The before run passed in 22.0s (24.0s runner), with no page errors. It made
18 source HTTP requests, including one retired trace read. Entering the service
accounted for eight; Customize and local staging added none. Parent/facet reads,
ephemeral comparison and fixed comparison each added two reads. Inspecting the
trace added two, and the held read canceled on Close added one.

The fixed baseline's full `/api/explore` response carried 100 matching spans in
88,973 bytes and reached headers in 160ms. Selected seven spans used 20,582 bytes
and 106ms; source-free comparison used 17,136 bytes and 58ms. These are one-run
observations, not a latency benchmark. Exact frozen queries feed the separate
backend acquisition-profile audit.

The two discovery bodies differed: the service-specific request included the
namespace selector while the initial inventory request did not. Their matching
environment and time bounds do not establish redundant owner/freshness work.
The after-only run reopened the GUI-saved view in a new isolated runtime and
passed in 6.9s (10.2s runner). Both query bodies and selected record rows were
identical. Counts, state, groups and duration summaries matched; acquisition basis
and captured-at receipts remained independent and fresh. The baseline aggregate
response was 1,768 bytes, about 98% smaller than the full response. Header elapsed
was 366ms, compared with 160ms before: this run demonstrates less response data,
not a latency improvement. The selected inspection still acquired a full result.

The maintained fixed Run budget is one full selected read, one aggregate baseline
read and one source-free comparison. A second Run held the real aggregate
response; navigation aborted it with `net::ERR_ABORTED`, and releasing the late
response produced no second comparison. No page errors occurred.

Run the ordinary self-contained flow with:

```sh
bash scripts/run_journeys.sh journeys/314-source-demand.spec.js
```

The optional `GOO_SOURCE_DEMAND_AFTER=1` mode is a two-phase measurement: it needs
the preceding GUI-authored store and frozen before-response artifacts from this
experiment. It reopens those exact queries rather than substituting a fresh
fixture/window. It is not an independent test requiring checked-in raw evidence. The temporary
store and full-response copies were pruned after this experiment; only the compact
summary and fixed baseline query remain locally.
The regular flow enforces the same full/aggregate/comparison acquisition budget.
The existing 313 images are reused because the workflow's visible controls did
not change.
