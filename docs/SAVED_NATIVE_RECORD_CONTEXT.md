# From a saved native interval to records

[Reported histogram exemplars](SAVED_HISTOGRAM_EXEMPLARS.md) extend this existing
Saved owner with a separately checked export/sample relation. Interval gestures
retain the contract below; reported references preserve the original measurement
window and exact namespace rather than substituting a derived interval.

Milestone 364 adds the existing native interval-context gesture to saved
counter/histogram evidence. Implementation, UI/generated, genuine native/recorded-fact model replay and
positive actor checks pass, together with the shared kept-context regression. See [validation](../VALIDATION.md) for execution status.

Select a retained export interval, then choose **Logs in interval** or **Spans in
interval**. The child shows **From saved native interval ·** the capture ID prefix.
It stages the existing fixed record question; only explicit Run reads fresh
records. **Return to saved measurement** restores the held captured result,
selected interval and opener without a metric reread or capture-store GET.

```mermaid
sequenceDiagram
  participant User
  participant Capture as Existing capture owner
  participant Saved as Saved native context owner
  participant Fixed as Existing fixed-question owner
  participant Records as Record source
  User->>Capture: Open immutable measurement capture
  Capture->>Saved: Verified receipt and retained result
  User->>Saved: Select actual stream export interval
  Saved->>Fixed: Stage exact context question and saved citation
  Note over Saved,Fixed: No telemetry read; captured DOM remains held
  User->>Fixed: Explicit Run
  Fixed->>Records: Fresh logs or spans in exact context
  Records-->>Fixed: Independent record evidence and read clock
  User->>Fixed: Return to saved measurement
  Fixed->>Fixed: Suspend child work and retire pending publication
  Fixed->>Capture: Restore held capture, interval, focus and scroll
  Note over Capture,Saved: Same captured DOM/receipt; no metric or store read
```

## Admit the actual captured interval

The saved-capture protocol checks that the measurement question equals its
immutable receipt question before adoption. The bounded
`ui/saved-native-record-context.mjs` owner detaches and freezes its own copy of
the receipt/result, without freezing or mutating caller inputs.

An action must identify a stream actually held in that result and exact export
bounds actually present in its intervals. The original measurement question,
native basis and read clock must also match. The owner recomputes the record
question through the unchanged Go/JavaScript native-context contract and rejects
forged questions, replacement streams and invented bounds.

Existing planner restrictions remain: supported counter/histogram export
intervals only, with established service/resource identity and exact half-open
nanosecond membership. The enclosing millisecond source window retains exact
timestamp predicates, policy, namespace and instance distinctions. Metric
attributes and instrumentation scope do not become record filters. Gauges,
stocks, aggregated histogram time buckets and unsupported resources acquire no
invented interval or lookaround. An unknown native value does not manufacture a
numeric contribution.

## Retain origin, not a joint observation

The origin disclosure keeps the original capture receipt, measurement question,
exact stream, interval, native basis and source-read clock. Custody time, original
metric read time and fresh record read time remain separate. Shared reported
resource/time is contextual, not exemplars, contributing requests, causality,
complete traces or an atomic native/record snapshot.

The captured evidence and its DOM stay held while the existing fixed owner runs
the child. A refused record read is unavailable, not empty or a failed native
measurement. Return suspends current child work, rejects late publication and
restores the same saved route, selected interval, inspection state, focus and
scroll. Capture replacement, refresh or departure retires the old origin.

Saved action callbacks carry creation-time capture/render leases. Suspension
retires them; resume renews only current connected anchors, including lazily
created interval controls, without remounting the saved result. An old callback
does not become current again merely because its DOM node is reused.

The app reuses its retained-record staging helper and `ResourceReturn` presentation
for kept and saved-native origins, with optional citation/notice fields. The live
Measurements return path retains its own ownership. No new HTTP endpoint, source
query, calculation kernel, capture format or global navigation framework is added.

Saving/sharing the child record question preserves its ordinary filters, not this
tab's native citation or a new cross-signal artifact. Further explicit trace
expansion keeps its existing broader-context disclosure and separate fresh reads.

## Acceptance and next work

The genuine journey covers a current-only counter and effective-policy histogram,
submillisecond boundary records, namespace/instance controls, explicit Run,
refused/late child responses and repeated return. It must preserve captured
bytes, independent clocks, source DOM and focused opener, with Stage/Return
reading no telemetry or stored capture. Store access during initial opening is
accounted separately. Generated owner/interval checks, the existing drilldown
model and positive actor completion are separate from source acceptance. See
[architecture](reviews/SAVED_NATIVE_CONTEXT_ARCHITECTURE_364.md),
[design](reviews/SAVED_NATIVE_CONTEXT_DESIGN_364.md) and
[storyboard](journeys/364-saved-native-context.md).

After this cut's acceptance, 365 has a strict captured-record review contract gate;
if it cannot close in a bounded spike, defer it and advance 366's measured compact
paired headers. No subsequent whole-product sequence is selected here.
