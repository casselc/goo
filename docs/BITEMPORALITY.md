# Bitemporality in Goo

Quick design spike, 2026-10-16. Recommendation: preserve reproducible evaluations
and append corrected generations before introducing a general bitemporal store.

## What the two times mean

Business event time answers when acceptance/completion happened. Immutable
custody time answers when the selected evidence system durably knew it. Evaluation
time is a separate parameter: it establishes the cohort and maturity boundary.
Reevaluating yesterday's cohort with new evidence must keep that evaluation time
fixed. Advancing the clock would change the question as well as the evidence.

Collector receipt, edge durable custody, and warehouse ingestion are different
clocks. Name the chosen clock and preserve it through export. A cutoff alone does
not prove that all evidence before it has arrived; coverage needs an independently
established completeness reference at the same scope and basis.

## Existing precedents and their limits

The sibling [research note](../../observability/otel-chdb/research/bitemporal.md)
starts from XTDB's bitemporal indexing series and separates append-only telemetry
from corrected catalog claims. Concrete mechanisms are available locally:

- [Query bases](../../observability/otel-chdb/query/internal/basis/basis.go)
  implement per-cluster custody bounds and strictly select `received_at < C`.
  [Verification](../../observability/otel-chdb/query/internal/basis/verify.go)
  rejects ahead-of-completeness and expired bases. The research note contains an
  older inclusive-filter sentence; the implemented contract is exclusive.
- [Alert late-data handling](../../observability/otel-chdb/alerts/README.md)
  retains window bases and reevaluates entire affected windows at newer bases.
  It does not add delta aggregates to old verdicts. Historical corrections never
  resolve an active firing group; notification policy remains separate.
- [Catalog bitemporal events](../../observability/otel-chdb/entities/bitemp/README.md)
  have a Go reference resolver, property tests, replay measurements and a
  [Quint model](../../observability/otel-chdb/model/bitemporalCatalog.qnt).
  Adoption into catalog storage remains proposed. Its authority tiers, trust
  window and announcement fallback solve catalog-specific problems and should
  not become generic business-event correction rules.

## Current Goo pilot

[Lifecycle policy and receipts](../lifecycle/model.go) now name `custodyClock`,
use an exclusive evidence cutoff, and carry optional `supersedesEvaluationId`.
[Preparation](../lifecycle/prepare.go) separates the full input-manifest digest,
visible-evidence digest and policy-bound snapshot digest. Adding excluded future
or out-of-scope deliveries can change the manifest without changing the visible
evidence identity. These hashes identify supplied artifacts; they do not attest
source completeness or preserve the artifacts themselves.

The CLI now requires a supplied predecessor artifact for a correction and checks
its identity, fixed question/revisions/scope/custody clock, and forward evidence
movement with `ValidateCorrection`. It does not establish predecessor existence
in a durable store, coordinate forks/concurrent publication, or provide immutable
receipt storage. Conflicting payloads for one event ID remain impaired evidence;
they are not silently interpreted as legitimate corrections.

## Minimal next design change

Define an evaluation-generation publication contract. A correction gets a new
evaluation ID, references its predecessor, retains the original cohort,
evaluation time, scope and definition revision, and advances the evidence basis.
Keep the original receipt and result available. Distinguish this operation from
reevaluation under a new definition. Validate predecessor compatibility before
publication; choose explicit conflict handling before allowing concurrent writers.
Legitimate source-event corrections will separately need immutable correction IDs
and explicit supersession/retraction semantics.

For [Databricks portability](DATABRICKS_METRIC_VIEWS.md), prepare one fact per
`(evaluation_id, scope, subject)` and select one evaluation generation before
metric-view aggregation. Summing all generations double-counts subjects. Preserve
receipts separately so empty or impaired scopes remain representable. Delta table
time travel can assist physical replay, but does not replace business event time,
custody provenance or the evaluation-generation contract.

## Small verification journey

The first two semantic checks now have lifecycle fixtures and Hegel/Quint
coverage (see [verification](../verification/README.md)); their UI journey is
still proposed. Concurrent publication remains a future model/implementation:

1. **Historical stability:** evaluate at basis B1; deliver a terminal received
   exactly at B1 or later; replay B1 and obtain the same visible evidence and
   subject results. Receipt diagnostics about supplied excluded input may differ.
2. **Correction preserves the question:** publish B2 superseding B1 for the same
   cohort and policy. A newly visible on-time completion can change bad/unknown
   to good, while B1 remains reproducible and selecting B2 counts the subject once.
3. **Concurrent publication safety:** interleave duplicate deliveries, retries and
   competing correction publishers. Require deterministic preparation and prevent
   a stale publisher from replacing the selected generation. This last guarantee
   belongs to a future storage/publication protocol, not the present pure evaluator.

Model the immutable evaluation map and selected-generation pointer separately.
Check history preservation, compatible forward supersession and uniqueness of the
selected contribution. Trace comparisons should assert semantic facts and evidence
identity rather than expecting whole receipts to match after excluded input grows.
