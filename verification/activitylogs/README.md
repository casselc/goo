# Activity log-context verification

Run serially with the project's Go/Node/Quint tools:

```sh
cd verification
go test ./activitylogs -count=1 -v
GOO_ACTIVITY_LOGS_REPORT=/absolute/path/to/browser/report.json go test ./activitylogs -count=1 -v
```

The 64-case Hegel property invokes the actual JavaScript model and checks its
full-window severity-text population, six bucket counts, scale, reduced log-rate
rational and selected UTC bounds against independent Go big.Int arithmetic.
Empty, uppercase and lowercase severity text stay distinct; zero buckets and
zero populations retain no-data state. Large counts exceed JavaScript's safe
integer range. Original OR predicates and typed environment policies survive the
projection. The actual generated questions pass Go validation/compilation. Each seed also
executes 60 generated commands against the actual shared ReadGate, compared with
an independent state-machine oracle for scope, generation, pending/frozen state
and admission. This exercises duplicate starts, stale completions and
pause/resume/scope-change interleavings.
The bridge reuses actual shared DTO admission with poisoned count, window,
severity, duplicate-group and bucket controls; it never acquires a source.

A strict Quint protocol checks the optional owner's one-request/three-SQL read,
explicit completion, freeze-before-selection, local explanation/return and
retirement of a still-pending read. Hidden reads,
inspection fan-out and late adoption must fail. This is an intended protocol
model, not execution of the UI controller. Missing Quint fails, rather than skips.
The independent controller/browser checks remain responsible for actual timers,
reattached callbacks and stale responses.

The optional observed gate reads `goo.activityLogs/1` reports containing successful
aggregate `reads`, quiescent per-gesture `facts` (show, select, return and optional
refresh/explain), `selected.parent`/`selected.selection` and `errors`. It sends
actual DTOs through the real model, checks the selected frozen window and original
selector, and replays actual read deltas through Quint. Perturbed selection bounds,
parent clocks and local SQL counts must be rejected. Absence of the report skips
only this explicitly optional gate; it is not proof of a real-source pass.
Stream polling and an explicit **Run explanation** belong to separate owners and
are not silently reclassified as optional aggregate reads.
