# 381 — read real record and native sources after decoder consolidation

Actual request-only acceptance passes (2.7 seconds). This milestone
changes the shared bounded source decoder, not a user-facing control or layout.
Existing UI screenshots remain applicable; this check does not claim a new visual
journey or fabricate browser paint for an internal refactor.

1. Independently read an existing ordinary service gauge point and a valid span
   from ClickHouse. Both queries are bounded and read-only. Selection requires
   actual captured records; missing source data fails explicitly rather than
   injecting or synthesizing a successful fixture.
2. Build exact typed environment/namespace questions. The native question pins
   complete stream identity and a one-nanosecond observation window. The record
   question pins reported trace/span IDs and exact event time within its original
   fixed window. Separate raw queries establish expected values and identities.
3. Read both questions through the real Go APIs and compare the captured source
   values/timestamps and record IDs to those independent raw observations.
4. At a scoped test-only upstream boundary, refuse each family with HTTP 503, then
   explicitly retry the same immutable question against the real source. No
   successful telemetry or application response is replaced.
5. Repeat with malformed HTTP 200 source data, then recover against real source
   records. Refusals are structured and publish no partial rows. Recovery retains
   the exact question and generated query plan.
6. Verify every application SQL request remains read-only with overflow refusal.
   Keep the actual request/response and upstream-call report. The check performs
   no ingestion, capture, package publication or other custody write.

The executable check is `381-source-decoding.spec.js`. This direct-source check
complements independent unit/property boundaries, including differing row limits,
blank-line/raw-byte policies, byte limits and transport-cause preservation.

## Actual check

The run independently selected a retained `lab.queue.depth` gauge point and a
retained span. Each one-record API result matched independently queried raw
identity, timestamp and value. Across both families, ten actual API responses
covered baseline success, unavailable refusal, exact-question recovery, malformed
source refusal and another exact-question recovery. Four independent source
queries established the oracle. The fault proxy recorded 16 read-only SQL
requests: 12 forwarded unchanged to ClickHouse and four deliberately refused or
malformed at the scoped test boundary. No successful telemetry was substituted.

Report:
`artifacts/source-decoding381/826c4234-69e8-40b8-ba23-9f2cf1b5c849/summary.json`.
The report retains exact questions, responses, source queries and upstream
parameters; there are no screenshots because no UI or browser state changed.
