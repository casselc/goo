# Measure one investigation's source demand — 314

This bounded experiment follows real ingested HTTP telemetry through Service
Customize, an explicit ordinary-count projection, method facet selection and
selected-versus-rest comparison. The selected cohort has seven POST spans; its
complement has 100 GET spans. Counts describe the full matching population even
though the parent inspection page holds only 100 of 107 spans.

Promote the rest comparison using **Include comparison in view**, retain its
checked method facet, save its fixed selected/baseline queries, then Run explicitly.
The paired facet identifies Both window values; preserving the control performs no
automatic value or population read. Inspect the timeout trace
and retire one outstanding actual trace read. Local staging and retirement must
not publish late results or issue automatic replacement reads.

[Maintained browser measurement](../../journeys/314-source-demand.spec.js) reuses
the real OTel fixture and standard traffic-pausing wrapper from
[313](313-active-canvas.md). The maintained [337 storyboard](337-paired-facet.md)
covers the retained-facet comparison control.

The measurement records exact HTTP request bodies, response bytes, header/body
elapsed times and checkpoint budgets. An artifact freezes both fixed comparison
queries for a separate backend audit. HTTP actions are not SQL statement counts;
this test does not attribute ClickHouse statements to individual UI actions.

## Historical acquisition-profile experiment

The original 314 before run passed in 22.0s (24.0s runner), with no page errors. It made
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

## Current source-budget contract and replay

The maintained fixed Run budget is one full selected read, one aggregate baseline
read and one source-free comparison. A second Run held the real aggregate
response; navigation aborted it with `net::ERR_ABORTED`, and releasing the late
response produced no second comparison. No page errors occurred.

Run the ordinary self-contained flow with:

```sh
bash scripts/run_journeys.sh journeys/314-source-demand.spec.js
```

The optional `GOO_SOURCE_DEMAND_AFTER=1` mode replays the **current** full flow's
GUI-saved definition and checks retirement. Run the ordinary flow first, then run:

```sh
GOO_SOURCE_DEMAND_AFTER=1 GOO_GENERIC_BINARY=/absolute/path/to/goo bash scripts/run_journeys.sh journeys/314-source-demand.spec.js
```

Replay consumes that run's `http.json`, `fixed-queries.json` and temporary
`capture-store`, selecting the saved view by its exact two operand questions.
It does not need historical `*-before.json` files or manufacture an older full-read
profile. It preserves the facet and Both window label, exact queries, selected
rows, counts and cohort semantics while requiring fresh acquisition clocks. The
baseline must omit rows and `viewDataset`; the source budget remains one full
selected read, one aggregate baseline read and one source-free comparison, with
zero automatic facet value reads. A held second baseline read is aborted on
departure and cannot publish a late comparison. Byte sizes are recorded only as
current observations, with no historical reduction assertion.

These per-run artifacts are temporary; repeat the full phase after cleanup rather
than checking private/raw store data into Git. Earlier measurements above remain
historical. 337 updates the promotion label and retains the facet declaration;
its full-flow rerun passed in 18.8 seconds and current-definition replay passed
in 8.3 seconds. Those are historical 337 measurements, not a new 338 replay.
Set `GOO_GENERIC_BINARY` to reuse the reviewed compiled runtime (otherwise each
lane builds its own binary with `GOO_TEST_GO`, or `go` from PATH).

338 reruns the ordinary flow with the reviewed Go 1.27.3 executable: passes in
33.1 seconds, with the paired **Both window values** label and zero automatic
facet-value reads. The fixed Run still acquires seven selected records, an
aggregate-only baseline count of 100 and one derived comparison; the actual held
trace request aborts with `ERR_ABORTED`. There are no page errors. The optional
exact-definition replay lane is deliberately skipped in this invocation. Reviewed
raw stores/bodies and the temporary executable are removed; compact receipts and
maintained frames remain.
