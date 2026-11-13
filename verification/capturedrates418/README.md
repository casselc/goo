# Captured-rate comparison verification

This isolated package checks the supplied-data boundary; it opens no telemetry
source. Thirty-two deterministic Hegel seeds generate independent log/span
cohorts with counts beyond JavaScript's exact integer range, unequal exposure,
known zero, independent clocks and inconsistent source reads. `math/big.Rat`
computes an independent count × 1000 / window-ms oracle and exact signed changes.
The actual Go comparison is admitted by the real JavaScript validator and measure
model through `bridge.mjs`, with corrupt count/exposure/profile/clock/changes
rejected. Inputs and source-attestation state must remain unchanged.

The native Quint model separates explicit source-API acquisitions from pure
comparison POSTs and local display/configuration gestures. These counters are
API acquisitions, **not SQL query counts**. Rejected mutations prove Stage,
Prepare and local disclosure cannot acquire source telemetry; Prepare may read
stored metadata and execute the existing pure comparison operation.

Root runs all heavy checks serially, from `verification/`:

```
GOO_CAPTURED_RATE_REPORT=/workspace/goo/artifacts/captured-rates418/report.json \
  go test ./capturedrates418 -count=1 -v
```

Put the current Go and local `quint`/Node tools on PATH. Native Quint is mandatory
and fails if absent. The actual-ledger test skips only when no report was supplied
for a local arithmetic-only run; milestone acceptance requires the completed
real-source report and no skips. Browser generation belongs to
`journeys/captured-rate-comparison.spec.js`; raw artifacts remain ignored.
