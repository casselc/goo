# Native bitemporal storage for Goo

Research assessment, 2026-10-22. The main annotation store uses append-only
revision files. Subsequent [implementation follow-through](WORKBENCH_ANNOTATIONS.md)
adds an optional XTDB 2.2.0 retained-evidence experiment with native integration
tests; a general migration and a native ClickHouse connector remain proposed.
No comparative backend benchmark is claimed here.

## Recommendation

A native bitemporal store is a good candidate for Goo's authored metadata and
historical entity relationships. Keep telemetry in ClickStack and retained
evidence in its existing immutable artifact boundary. Evaluate XTDB for a shared
server deployment and Minigraf for an embedded Rust deployment. Do not make either
database part of the measurement grammar or browser component contract.

This extends the earlier [bitemporality spike](BITEMPORALITY.md): evolving tags,
findings and training judgments create a stronger need for historical joins than
the original lifecycle evaluator alone did.

## What becomes easier

Two independent coordinates matter:

- **Valid time:** when a claim applies to the observed world, such as a service's
  ownership during an incident or the interval covered by a judgment.
- **System time:** when the metadata database committed that claim or correction.

For example, on Thursday someone corrects an ownership assignment for Monday.
An investigation saved Tuesday must still be able to show Tuesday's knowledge
about Monday. A current retrospective can show the corrected ownership instead.
Neither operation changes Monday's telemetry.

Useful queries across product surfaces include:

1. Find services tagged with an incident category at valid time V, using only
   knowledge committed by K; join their scoped identities to retained findings.
2. Reopen an investigation with the definition, ownership and annotations known
   when it was saved, then explicitly compare with today's interpretation.
3. Assemble a training dataset using a fixed review cutoff, retaining corrections,
   reviewer disagreements and exact evidence revisions for later audit.
4. Traverse evidence → finding → definition → test run, preserving historical
   relationships when any one of those objects is revised.

An immutable definition revision usually needs system history and exact identity;
it does not automatically need a valid-time interval. Explicit service claims do.
Do not manufacture applicability windows for every artifact to satisfy a database.

## Candidate fit

| Concern | Minigraf | XTDB, current 2.x architecture |
| --- | --- | --- |
| Deployment | Embedded single-file Rust library | Server database; cloud architecture uses a coordination log and object storage |
| Query model | Temporal EAV graph and recursive Datalog | SQL and XTQL; SQL:2011 temporal functionality |
| Integration with Goo | Direct Rust integration, or an additional boundary from Go; no Go binding listed in reviewed support table | PostgreSQL wire protocol offers a familiar Go client boundary; actual driver/type behavior still needs testing |
| Likely role | Local workbench, portable investigation graph, small embedded installation | Shared metadata service and historical cross-entity queries |
| Principal cost | Binding, file lifecycle and operational maturity evaluation | Additional database service and its deployment/recovery responsibilities |

Minigraf's README currently recommends 2.0.9 and discloses a v2.x integrity issue:
writing or retracting multiple values of the same attribute for one entity in a
single call can read back as one value (#371). The documented fix requires the
v3 file format. This matters directly for multi-tag assertions. Model distinct
assertion/member entities and test actual behavior; do not trade away transaction
atomicity merely to work around a storage limitation. Its documented file-backed
fact limit is 4,080 serialized bytes, so a 4,000-byte Goo note plus framing needs
explicit testing or separate blob storage. Rust/Python are Tier 1; C/FFI and
browser bindings are experimental Tier 2. Clustering and replication are outside
its stated scope.

XTDB's own current README describes an immutable, ACID, columnar Arrow engine,
bitemporal valid/system time, SQL/XTQL and PostgreSQL wire support. That is a
different offering from XTDB v1's Datalog API; third-party comparison tables
describing XTDB simply as a Datalog graph database should not drive this decision.
The Go application can remain Go even if an external database uses another runtime.

These are documentation findings, not evidence that either system meets Goo's
latency, durability or deployment requirements.

## Semantics the database cannot decide

Keep each authored assertion revision as its own identity, with its parent,
target, origin and optional applicability. Two people correcting the same parent
must remain two reviewable branches. Overwriting one service row and accepting
the database's most recent version would erase that application-level distinction.
Selecting or resolving branches is an explicit Goo operation.

A database system timestamp is not collector receipt, telemetry custody time,
business event time, reviewer identity or an attestation. A metadata cutoff does
not establish telemetry completeness. Historical annotation queries cannot
reconstruct deleted evidence, turn bounded record pages into complete populations,
or establish unbiased classifier training sets.

Dataset artifacts should pin assertion IDs, evidence IDs, definitions, review
policy and the committed database basis. Valid time plus a wall-clock cutoff alone
is insufficient when commits are asynchronous or reads lag. Return a transaction
token/read basis and require the query to have indexed that commit. An intentional
historical replay must not silently issue a fresh telemetry query.

## Adapter and acceptance spike

### Annotation-triggered evidence retention

The user's follow-up proposes retaining selected ClickHouse records in XTDB when
annotating them. This materially broadens the useful role: XTDB could hold both
authored meaning and the bounded evidence that meaning refers to, while ClickStack
continues to serve large live populations. The bounded direct-write experiment is now implemented for selected saved-capture observations; live-result freezing and continuous ingestion remain proposed.

The current official external-source docs describe **ingestion**, rather than
remote query federation: a secondary database subscribes to an upstream feed,
records XTDB transactions and becomes a read-only mirror. Postgres uses an initial
snapshot followed by transaction streaming; Kafka Connect consumes keyed records.
Annotations therefore belong in a writable database, not in the mirrored database.
Do not assume atomic publication across those database boundaries.

There are three distinct implementation choices:

| Mechanism | Role | Tradeoff |
| --- | --- | --- |
| Goo capture worker writes selected records to XTDB | On-demand retention when a user saves an annotation | Smallest first experiment; no ClickHouse connector required |
| Durable capture feed → Kafka Connect external source | Replayable ingestion of capture envelopes | Reuses an existing source, with asynchronous indexing and feed operations |
| Native ClickHouse external source | Continuous mirror of selected source tables or a capture outbox | Useful only after defining source identity, snapshot/stream handoff, ordering, mutation/delete and replay semantics |

A ClickHouse polling query over event timestamps alone cannot establish a complete
change feed: late inserts, mutations, TTL and retries need explicit handling. For
the on-demand case, a durable feed of **completed capture requests** is a clearer
boundary than trying to mirror every telemetry table. Current Kafka Connect docs
also limit consumption to partition 0 and halt on malformed records; pin and test
the chosen release before designing around it.

Proposed user workflow:

1. Select rows, a trace or a chart population and choose **Annotate and retain
   evidence**. Describe which bounded evidence will be copied. A saved query alone
   remains a reusable question, not a frozen dataset.
2. For exact displayed records, preserve server-retained result bytes or verified
   stable identities. A fresh query must be described as a fresh capture; do not
   pretend it necessarily returns the rows the user originally selected.
3. Persist immutable capture members plus a manifest: source identity, schema and
   adapter revision, query/parameters, event window, available custody basis,
   extraction timestamps, counts, byte limits, truncation/coverage and member hashes.
   Keep source record identity separate from payload hash: identical legitimate
   events may be distinct occurrences, and corrected payloads may share a source ID.
4. Publish the annotation referencing the completed manifest. Small captures could
   use one native XTDB transaction with the annotation. For chunked imports or an
   external-source mirror, stage members and publish a completion marker only after
   verifying the expected member set is queryable; readers ignore incomplete sets.
   Use idempotent request/member IDs and recover interrupted publication.
5. Reopen and review the retained evidence without reading ClickHouse. Later
   expansion is another capture linked to the original; it never mutates the old
   training example or silently upgrades a bounded sample into a full population.

Preserve OTLP scalar types, nested attributes, nanosecond timestamps and exact
missing/empty distinctions in the capture encoding. Test XTDB's actual type
round-trips; do not rely on floating-point JSON numbers for nanosecond precision.
Retention permissions, redaction and deletion need to apply to the copied records
as well as the live source. Database history alone is not an access-control policy.

External sources also refine the clock contract above: their XTDB system time is
not universally the instant Goo observed a commit. Postgres snapshot system time
uses XTDB's internal clock, streaming uses the upstream commit timestamp, and Kafka
uses each record's timestamp. Preserve those meanings along with Goo capture time,
original event/custody clocks and an indexed read basis. Starting a mirror today
does not reconstruct upstream history that the snapshot never contained.

This makes XTDB a promising retained-evidence and annotation backend, rather than
merely a metadata index. First prove one selected-record capture, an interrupted
capture recovery, annotation correction and historical reopening after source
expiry. A full ClickHouse connector can follow if continuous ingestion has a
demonstrated use case.

Keep target/revision validation in Go. A storage adapter should publish immutable
assertions atomically, return a committed basis, enumerate branches, query at an
explicit valid/system basis, and export/rebuild the historical graph. Authored
timestamps supplied by clients must not impersonate database commit time.

Before adoption, execute the same conformance cases against a reference model
and each candidate:

- Backdated corrections preserve results at earlier knowledge cutoffs; interval
  boundaries are half-open and missing/empty namespace identities remain distinct.
- Concurrent corrections, withdrawals, retries and explicit resolutions retain
  every branch and never silently choose a reviewer winner.
- Multi-tag publication is atomic; long UTF-8 notes round-trip without truncation.
- Kill/reopen, transaction failure, backup/restore and export/rebuild preserve
  assertion identities and temporal answers. Test reader lag and read-your-write.
- Historical investigation/dataset reopening reads retained metadata and evidence
  only; an injected telemetry read fails the trace check.

Use Hegel state/swarm campaigns for publication interleavings and Quint traces for
commit visibility, correction and branch resolution. Measure realistic metadata
cardinality, graph depth, latency and recovery cost separately from telemetry load.
Keep the normal annotation store independent of the optional adapter experiment.
Its current validation and remaining limitations are recorded in the
[implemented contract](WORKBENCH_ANNOTATIONS.md).

## Reviewed sources

- [Minigraf README](https://github.com/project-minigraf/minigraf/blob/efe26096f53750f00277dcdaa99312c53b3fde81/README.md),
  including release recommendation, known issues, storage limits and binding tiers.
- [Minigraf philosophy and support policy](https://github.com/project-minigraf/minigraf/blob/efe26096f53750f00277dcdaa99312c53b3fde81/PHILOSOPHY.md),
  including embedded scope, format migration and explicit non-goals.
- [XTDB README](https://github.com/xtdb/xtdb/blob/c6c19623d90c658c15400fb46dc056f3735ed57b/README.adoc),
  including current SQL, temporal and deployment architecture claims.
- [XTDB external-source overview](https://github.com/xtdb/xtdb/blob/c6c19623d90c658c15400fb46dc056f3735ed57b/docs/src/content/docs/ops/external-sources/overview.md),
  [Postgres reference](https://github.com/xtdb/xtdb/blob/c6c19623d90c658c15400fb46dc056f3735ed57b/docs/src/content/docs/ops/external-sources/postgres/reference.md)
  and [Kafka Connect reference](https://github.com/xtdb/xtdb/blob/c6c19623d90c658c15400fb46dc056f3735ed57b/docs/src/content/docs/ops/external-sources/kafka-connect/reference.md).
  Reviewed repository HEAD documentation; a deployed release still needs pinning
  and conformance testing.
- [Existing sibling research](../../observability/otel-chdb/research/bitemporal.md)
  and the catalog resolver references in [Goo's earlier spike](BITEMPORALITY.md).
