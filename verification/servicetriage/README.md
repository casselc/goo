# Qualified service-triage verification

This isolated package imports no private milestone helpers and changes no module
requirements. Run serially with the configured project Go/Node/Quint toolchain:

```sh
(cd verification && go test ./servicetriage -count=1 -v)
GOO_SERVICE_TRIAGE_REPORT=/absolute/path/to/real/browser/report.json \
  sh -c 'cd verification && go test ./servicetriage -run TestQuintObservedTriageLedger -count=1 -v'
```

`TestQualifiedTriageProjectionAndExactRanking` uses 64 Hegel draws against the
actual JavaScript model through `bridge.mjs`. An independent Go big.Int oracle
checks qualified missing/empty/named namespace identities, exact adjacent windows,
record-only Server membership, count changes and rational Error/mean differences.
Counts exceed JavaScript's safe integer range. Each rank is independently computed
by cross multiplication; a metric-only service cannot acquire span evidence.
The generated questions pass Go validation/compilation. Poisoned echoes, changed
reads, duplicate groups and mismatched count/status/duration measures are refused.
These aggregate fixtures are generated contract inputs, not additional telemetry
observations or proof of a complete source snapshot.

`TestQuintTriageReadAndRetirementProtocol` uses the sibling observability repo's
quint-go qobs/validate tools with `triage.qnt` and `binding.yaml`. It checks the
bounded protocol (two aggregate requests/four SQL statements for Measure, no new
triage reads for sort/search/return, no adoption after retirement). Negative
controls require rejection of hidden reads, read-before-admission, fan-out and
late adoption. This is a model/toolchain unit check, not execution of the UI
controller; actual stale-response behavior is the controller/browser test's job.
Missing Quint fails this required test rather than looking like a passing skip.

The optional observed gate consumes the real browser report
`goo.serviceTriage/1`: exact current+baseline request/response carriers, accepted
scope and actual gesture counters. It independently validates qualification,
Server/namespace population, typed environment policy, adjacent windows and exact
measure conservation, then replays the four actual Measure/sort/search/return
read budgets through Quint. Foreign windows, altered fields, changed-read results
and added local reads must be rejected. An absent report explicitly skips only
this observed gate; it is not a claim that the real campaign passed.

No gotogether dependency was added. This new owner uses the already adopted
quint-go bridge directly; adopting a second runner here would not strengthen the
underlying observation or arithmetic oracle.
