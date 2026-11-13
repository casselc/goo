# Ordinary record-intent admission — 417

This isolated package compares the actual pure JavaScript configuration validator
with Go's strict JSON decoder followed by `exploration.Query.Validate`. It runs
32 Hegel-generated environments/windows over a fixed independent legality corpus
covering both signals, exact UTF-8 field/value boundaries, predicate depth/node
budgets, nested unknown fields, intrinsic/numeric operators and range limits,
grouping/chart constraints, typed environment policy and reported-edge ownership.
Every case has an explicit expected result before either implementation runs;
matching two incorrect validators cannot pass the expected-result oracle.

The bridge also checks caller immutability and source-free operation, and rejects
JavaScript-only malformed/noncanonical configuration: explicit nulls, unpaired
surrogates, hidden symbol members, exotic prototypes, accessors and cycles. Object-only
cases are separate from wire parity because JSON cannot represent descriptors,
prototypes, symbols or cycles. Explicit null arms and case aliases are also in
the common wire corpus, checking the hardened Query decoder. This package
checks semantic configuration admission, not URL-size/editor-affordance budgets
or browser navigation. The owner journey provides those integration checks.

From `verification/`, with the supported Go toolchain and `node` on PATH:

```sh
go test ./recordintent -count=1 -v
```

The check never requires a ClickHouse or Databricks instance, never skips a
missing Node executable, and never reads or modifies telemetry or stored scopes.

## Explicit handoff protocol

`model.qnt` keeps three states (initial, staged, held) and four actions: local
configuration checks, source-free staging, one explicit Run, and refused intent.
Refusal cannot clear held evidence or acquire telemetry. Native Quint negative
controls prove that implicit reads, dropped held state and Run before staging are
rejected. The actual browser ledger supplies DOM phase/held observations and API
POST counts; the validator checks all five named facts and rejects mutated actual
refusal observations.

```sh
GOO_RECORD_INTENT_REPORT=/absolute/path/to/artifacts/record-intent417/report.json \
  go test ./recordintent -count=1 -v
```

Without the report variable only the observed-browser-ledger test skips. The
mandatory protocol test never skips missing Quint; CI/acceptance must provide the
actual completed report and report zero skipped checks.
