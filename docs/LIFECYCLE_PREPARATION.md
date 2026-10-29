# Frozen lifecycle preparation

The Go lifecycle pilot now prepares normalized immutable acceptance, terminal,
and retry events into subject facts. It feeds the aggregate compiler while
leaving the original Python application and UI query semantics unchanged.

## Try it

From the repository root with Go on PATH:

```sh
go run ./cmd/goo prepare -policy examples/lifecycle/policy.yaml \
  -events examples/lifecycle/events.json > /tmp/goo-prepared.json
go run ./cmd/goo evaluate-prepared -definition examples/lifecycle/measurement.yaml \
  -prepared /tmp/goo-prepared.json
python scripts/check_lifecycle.py
```

The sample intentionally declares coverage **unverified**. It produces one known
on-time outcome and one unknown missing outcome, with an unknown success ratio.
Setting coverage to complete is an external assertion requiring an evidence
reference; a configuration string cannot establish complete collection.

## Input and time contract

Input is a JSON array of normalized business events, not arbitrary OTLP records.
A future source adapter must supply immutable event IDs, scoped business identity,
business event timestamps, and a named immutable custody clock. `service_name`
is the logical owner of this lifecycle, not necessarily the emitting consumer's
OTel service name. Acceptance attribution owns the release dimension. A collector
receipt, edge durable custody and warehouse ingestion must not be substituted
for one another without an explicit source contract.

All nanosecond values cross JSON/YAML boundaries as decimal strings and are
parsed as checked signed int64 values. Timestamp zero is valid. Floating-point,
missing, malformed and overflowing timestamps do not become zero. The pilot
assumes event time is not later than custody time; violations impair evidence.
Distributed clock-skew allowances are not implemented.

For frozen evaluation time T, deadline D, allowance L and window W, select
acceptances in `[T-D-L-W, T-D-L)`. A terminal at exactly acceptance+D succeeds;
one nanosecond later is late. Terminal lookup is by scoped subject over the
visible evidence, not restricted to the acceptance interval.

Evidence is visible only when `received_time_ns < evidenceCutoffNs`, matching the
sibling query-basis contract. A delivery exactly at the cutoff remains invisible.
Missing completion becomes bad only with declared complete coverage and a cutoff
strictly beyond acceptance+D+L. Otherwise absence is unknown.

Recent subjects remain available for inspection. `is_mature` describes deadline
maturity, while `in_cohort` describes accounting membership. At exactly the
cohort's exclusive upper boundary, a subject can be mature but outside this
accounting interval. Therefore the paired aggregate definition filters on
`in_cohort`, not merely `is_mature`.

## Identity, ambiguity and evidence quality

Subject identity is `(environment, tenant, logical service, subject ID)`.
Immutable event identity is scoped by environment/tenant/service plus event ID.
Repeated delivery of the same payload counts once, even with a different delivery
timestamp. Consistent repeated acceptance/terminal facts under different event
IDs do not multiply subjects. Retries do not increase eligible counts.

Conflicting acceptance timestamps/releases, conflicting terminal decisions,
negative durations and conflicting reuse of an immutable event ID become unknown.
Ambiguous acceptance does not choose an arbitrary minimum time or release. If any
valid acceptance anchor intersects the cohort, that ambiguous subject remains
in accounting as unknown. Invalid records with no usable anchor remain visible
through receipt diagnostics and quality impairment.

Unknown fields/types, missing IDs/timestamps, unscoped invalid records, orphan
terminals and orphan retries are diagnosed rather than quietly dropped. A record
provably outside explicitly selected scopes is excluded. An invalid record with
unknown scope conservatively impairs all selected scopes.

Every preparation includes scope receipts, invalid/duplicate/orphan diagnostics,
counts, policy/source revisions, custody clock, evidence basis and exact cohort
bounds. Non-complete scope evidence is propagated to facts. A **request-wide
strict guard** also suppresses ratios for impaired empty scopes, which cannot
be represented by an ordinary subject row. This deliberately favors unknown over
claiming a healthy result for a partially observed multi-scope request.

`evaluate-prepared` consumes the complete artifact. Bare `evaluate -facts` remains
an aggregate-only API and cannot recover missing receipts. Both generated query
plans now require a boolean `evidence_impaired` binding in addition to the
evaluation ID. The Databricks YAML is reusable; its generated query applies this
receipt guard outside the measure so empty-source evaluations are handled too.
Direct queries against the published metric view do not enforce that API contract.

## Replay and corrections

Three hashes have different jobs:

- `input_manifest_digest` identifies the whole supplied source input, including
  invisible and out-of-scope records.
- `evidence_digest` identifies the basis-visible supplied evidence, including
  invalid records that affect quality.
- `snapshot_digest` binds the frozen policy/generation to that evidence digest.

Reordering input does not change these hashes. Adding an invisible future delivery
changes the input manifest but not the historical evidence/snapshot identity.
These are deterministic provenance checks, not authentication or completeness
proofs. A reader must retain the input snapshot or facts/receipts to replay it.

A correction uses a new evaluation ID and `supersedesEvaluationId`, preserves T,
cohort, deadline, definition/source revisions, custody clock and scope, and advances
the evidence cutoff/basis. The CLI requires `-previous <prepared-artifact>` when
superseding and invokes `ValidateCorrection`. The pure preparation kernel also
supports independent snapshots; publication callers must explicitly enforce
generation links. Coverage evidence may improve at the newer basis.

No previous artifact is mutated. Both generations may coexist in storage, but an
aggregate request selects exactly one evaluation ID. This pilot does not implement
a durable correction store, event retraction/supersession algebra, notification
history or a correction horizon. Conflicting event payloads are not legitimate
corrections; they remain conflicts. See [the bitemporal spike](BITEMPORALITY.md).

## Validation boundary

Independent fixtures cover half-open cohort edges, one tick before/at/after the
deadline, outcomes outside acceptance bounds, duplicates, retries, missing and
failed outcomes, negative durations, conflicts, invalid records, tenant isolation,
cutoff changes, empty impaired scopes and 1001 raw lifecycles. Go tests additionally
exercise epoch zero, large epoch nanoseconds, int64 edges, replay ordering,
invisible append identity, and invalid correction links.

`scripts/check_lifecycle.py` runs the real CLI, compares prepared identities,
classifications, bounds, diagnostics and counts with independent fixtures, and
executes aggregate SQL in embedded ClickHouse. It checks 26 total/grouped cases
plus coexistence of old/new generations without double counting. It does **not**
execute a SQL implementation of lifecycle preparation. The Hegel and Quint checks
are described in [verification/README.md](../verification/README.md).

The [live measurement adapter](MEASUREMENT_PACKAGES.md) now reads actual OTel
logs, persists a local first-observation custody basis and renders these Go facts
in the shared UI/capture/notes path. It explicitly opts into
`knowledgeAfterEvaluation: true`: a correction can reveal past business events
at a cutoff after T, while future business events remain invalid. The default
policy still requires cutoff ≤ T, and corrections preserve the chosen mode.
Coverage stays unverified, so missing completion is unknown in the live adapter.

Remaining integrations: generated ClickHouse/Spark preparation plans, production
custody/completeness contracts and a coordinated selected-generation publication
protocol. Databricks artifacts remain unexecuted on a warehouse.


## Prepared text boundary — 245

The `subjectfacts` Go evaluator validates execution/routing IDs and selected
string fields before identity, filtering and grouping. Valid foreign evaluation
IDs exclude historical fields; malformed selected rows cannot hide behind a
population filter. Authored literals and enums use the same UTF-8 requirement.
See [the text contract and pipeline](PREPARED_TEXT_INTEGRITY.md). Valid Unicode
and permitted empty values retain exact bytes; no evidence migration is needed.
