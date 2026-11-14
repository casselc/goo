# Read what a gauge actually reported

This journey starts with ordinary OTLP metrics from a producer that reports no
`service.instance.id`. Discovery can identify the instrument, but a reusable
measurement cannot safely evaluate that producer as an identified stream.

The raw inspection path answers a smaller question: **what observations did the
source return for this instrument and window?** It does not evaluate a gauge or
claim that a set of samples is continuous, complete, fresh, or healthy.

## Executable storyboard

`journeys/raw-gauge-observations.spec.js` sends actual OTLP to the local collector,
independently reads its exported ClickHouse rows, and copies only this run's rows
into an isolated source database. A generic Goo server starts without lab assets.
Successful requests use those actual rows. A source boundary injects one declared
access refusal. The stale-response test holds the exact completed Go HTTP response
after the real source read, then delivers those bytes after navigation. It never
synthesizes successful telemetry. A separate source service expands one real
exported row to 10,001 occurrences to exercise the actual atomic read budget;
that expansion is identified as a source fixture, not 10,001 SDK reports.

1. Discover the reported service. Review metadata without reading values. The
   missing instance identity keeps **Build measurement package** unavailable.
2. Choose **Read observations** for the gauge and its exact unit. A separate,
   bounded source request returns individual reported observations.
3. Compare repeated samples, zero, attributes, reported flags and precise UTC
   timestamps. Duplicate rows remain separate occurrences. An occurrence key is
   a local display key, never a manufactured producer identity.
4. Find a held observation and inspect its attributes. Both gestures operate on
   the already returned bank and issue no source requests or store writes.
5. Read another unit explicitly. The same metric name does not merge units into
   a fleet value, series, or timeline.
6. Encounter a declared source refusal or budget error. The interface reports
   the refusal without publishing a partial successful bank.
7. Leave a pending read, then return to the same scope. A held actual response
   from the retired visit cannot populate the new visit, restore stale focus,
   or make an old callback authoritative.

The final frozen Go binary passed this real browser journey in 56.5 seconds:
16 main steps, 35 main source queries, and 10 actor source queries. The 28-second
Bombadil campaign produced 553 snapshots, reached its observed completion state,
and reported no invariant violation. `TestQuintRawGauge438Observed` then passed
in 14.59 seconds with zero skips, independently checking the actual OTLP/exporter
multiset, raw observations, held receipts, acquisition authority and negative
controls. No page errors occurred. The owned server and database both cleaned up.

The final report is `artifacts/raw-gauge438/4074a73653d9/report.json`; its native
log is adjacent. Artifacts are local validation output; these generated frames
are the checked-in documentation:

- [Raw table, desktop](screenshots/438-raw-table-1440.png)
- [Raw table, narrow screen](screenshots/438-raw-table-320.png)
- [Observation detail, desktop](screenshots/438-held-inspection-1440.png)
- [Observation detail, narrow screen](screenshots/438-held-inspection-390.png)

Ten actual SDK-exported gauge rows cover missing and present instance identity,
repeated zero-valued observations, distinct units and attributes, no-recorded-value
flags, and OTLP nonfinite scalars. The separate 10,001-row source expansion fails
atomically with a typed budget refusal. A structurally valid All-environments
request and a non-gauge request both refuse before source SQL. The stale actual
HTTP 200 has a distinct source-read receipt; delivering it after a fresh read
changes neither the current receipt, held observations, nor keyboard focus.

## Deliberate limits

This is a gauge-only raw observation surface. Unsupported instrument types and
unsupported broad scopes do not acquire data. Raw inspection never relaxes the
measurement evaluator's instance identity or validity checks. Source read time
and each reported sample time remain separate facts. Exporter storage cannot
reconstruct an omitted OTLP scalar when the physical table has already replaced
it with a default; any deliberately corrupted source record is identified as a
fixture fault rather than ordinary successful OTLP.
