# Inspect what a sum reported

A service exports ordinary OTLP sums without `service.instance.id`. Its reported
values are useful to inspect even when measurement evaluation cannot establish
an identified producer stream. The operator wants the samples, not an inferred
rate, reset, interval total or fleet aggregate.

## Executable storyboard

`journeys/raw-sum-observations.spec.js` sends **232 real OTLP sum points** through
the collector. The fixture uses the same metric name and unit across delta and
cumulative, monotonic and nonmonotonic sums. It includes missing and reported
instance identities, repeated records, explicit zero starts, not-recorded flags,
a second unit and nonfinite values. The collector's exported rows are read
independently before copying only this run's rows into an owned isolated source.
The main exact instrument has 210 occurrences, with local pages capped at 100.

1. Discover instrument metadata. The missing producer identity keeps measurement
   package building unavailable. Metadata alone does not acquire raw values.
2. Choose **Read observations** for a reported sum tuple. The request retains its
   kind, temporality, monotonicity, scope/version/schema, metric, unit and original
   fixed discovery question. Cumulative nonmonotonic sums keep discovery's
   `stock` kind; the interface does not pretend they are monotonic counters.
3. Inspect each delta/cumulative and monotonic/nonmonotonic population separately.
   Compare returned rows with the exact exported population, preserving duplicate
   occurrences, attributes, units, starts, timestamps, values and flags.
4. Page within the 210-row bank, then Find the not-recorded observations. Their
   numeric value is unavailable rather than a fabricated zero. Open a row using
   the keyboard and inspect its reported temporality and monotonicity.
5. Leave Metrics and return. The held rows, filter, selected detail and read clock
   remain unchanged. These local gestures acquire no source data.
6. Encounter a source projection failure on an explicit reread. The test renames
   `Value` only in its owned isolated sum table, then restores it before cleanup.
   The request refuses without publishing a partial bank; prior held rows remain.
7. Start another read, navigate away and return, then deliver the exact completed
   successful Go response late. Only this declared adversary ignores the fetch
   AbortSignal. A retired completion cannot replace the earlier bank or clock or
   steal focus. Successful telemetry is never synthesized.
8. Read the other unit and nonfinite instrument explicitly. Neither joins the
   first tuple into a series. Missing monotonicity and native All scope refuse
   at the API boundary.

The existing raw-observation Bombadil actor uses the same actions and invariants
with an explicit sum endpoint/tuple fixture. It discovers, reads once, finds and
inspects held observations, then observes completion. This is not a claim that
arbitrary exploration covers all metadata combinations; the deterministic
journey covers those independently.

## Deliberate limits

Reported temporality and monotonicity describe exported points, not readiness,
freshness or mathematical compatibility. Start zero is not an inferred epoch
interval. No aggregation, rate/reset inference, connected timeline or new capture
carrier is introduced. Metadata probes do not establish value-column readiness.
Backend tests separately cover malformed structural identity, exact scope refusal
before source work, unknown temporality and atomic transport/result budgets.

## Runtime frames

- [Desktop sum inspection](screenshots/441-sum-observations-1440.png)
- [390 px sum inspection](screenshots/441-sum-observations-390.png)
- [320 px sum inspection](screenshots/441-sum-observations-320.png)

## Validation

Final real-app journey passes **43.0s** (test body 40.9s) against the final
compiled runtime. All **232 exported OTLP points** are checked against the
independent exporter bank. The four exact same-name/unit metadata tuples return
6, 6, 6 and 210 occurrences, including discovery's cumulative nonmonotonic
`stock` tuple. The other unit and real nonfinite instrument each return two rows.

The **14 recorded steps** preserve reported stream metadata, duplicate
occurrences, flags and original source timestamps. Local paging, Find, keyboard
Inspect, retained return and late-response delivery start no API POSTs. The
210-row bank renders 100 rows per page; its two not-recorded points display
**Unavailable · not recorded**, while source state and raw values remain intact.
The declared owned-table projection failure returns 502 with no partial bank;
the previous receipt and selected detail remain. Two typed HTTP refusals cover
omitted monotonicity and native All scope; backend tests prove their zero-source
boundary separately. An exact completed successful response delivered late
changes neither the held bank nor focus.

The updated existing Bombadil actor completes its **28-second campaign** with
**586 snapshots**, one explicit sum read, two held matching rows and zero
forbidden operations or overflow. Its local Find and inspection issue no further
requests. These actor counts apply to this one configured sum scenario; they do
not claim that every tuple ran under the actor. Runtime and isolated database
cleanup pass, and no browser console errors occur. Three final responsive frames
include the corrected not-recorded label.

One calibration corrected a broad unit-text selector: `ms` also occurs in the
reported scope `ordinary.sums`. Acceptance uses the exact accessible Read label
and retains all tuple/population assertions. Successful telemetry was unchanged.

Run receipt: `artifacts/raw-sum441/70ca0e346bc1/report.json`.
