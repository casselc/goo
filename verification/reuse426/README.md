# Investigation preservation and reuse (426)

The native Quint model separates explicit telemetry reads from definition-store
writes and reads. A delayed save can reach storage after its control is withdrawn;
it cannot become an adopted receipt or authorize reopening. Only a current,
completed Save produces that receipt. Definition reopen stages configuration;
only explicit Run reads its two comparison populations. Back and stale/hidden
preservation callbacks have no source or store effects. Capture publication is
forbidden throughout this configuration-only journey.

The actual complete browser report supplies seventeen continuous, ordered facts.
Validation requires every fact and checks source/SQL/count/compare/view POST/view
GET/capture counters, so a conforming prefix or missing return cannot pass.
Negative controls add hidden writes, reopen reads, return effects and premature
receipt adoption.

Hegel drives the production counted-value source owner through generated
logs/spans, namespaces and environment policies. Stateful transitions cover ready
custody, inactive callback refusal, source-free return, synchronous withdrawal and
scope replacement during pause, hidden late success/error, generation retirement
and disposal. These checks complement browser assertions over the actual
preservation controls; they do not infer the browser's DOM behavior from counters.

From `verification`, with Go/Node/Quint on PATH:

```sh
GOO_REUSE426_REPORT=../artifacts/investigation-consolidation426/report.json \
  go test -p 1 ./reuse426 -count=1 -v
```

Quint is mandatory. Supply the completed report for acceptance; without its
path only the observed-ledger check skips. Model action identifiers are separate
from UI labels and explicitly mapped from report facts.
