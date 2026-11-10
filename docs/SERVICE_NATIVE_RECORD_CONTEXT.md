# From a service native interval to records

Milestone 368 is implemented and validated with generated ownership checks,
a genuine browser journey, observed model replay and positive actor completion.
See [validation](../VALIDATION.md) and the
[executable storyboard](journeys/368-service-native-context.md).

From a service dashboard, open **Native result · exact streams**, select a
histogram export interval, and choose **Logs in interval** or **Spans in
interval**. The existing record workspace stages the question for explicit
**Run**. **Back to service** returns to the held service card and selection.

For a gauge, [retained observation context](RETAINED_GAUGE_RECORD_CONTEXT.md)
uses a uniquely retained reported timestamp and the requested widget question.
It clips a nearby window; it does not promote the point into an export interval.

```mermaid
sequenceDiagram
  participant User
  participant Service as Held service dashboard
  participant Fixed as Fixed record workspace
  participant Records as Record source
  User->>Service: Select actual histogram export interval
  Service->>Service: Admit held stream, bounds and action owner
  Service->>Fixed: Stage exact contextual question
  Note over Service,Fixed: No source read; service result remains held
  User->>Fixed: Run
  Fixed->>Records: Read fresh contextual logs or spans
  Records-->>Fixed: Records and independent read clock
  User->>Fixed: Back to service
  Fixed->>Fixed: Suspend child visit and reject late publication
  Fixed->>Service: Restore held card, selection, focus and scroll
```

The selected bounds belong to an actual held export interval, rather than the
chart's whole window or an aggregate time bucket. The existing native context
planner preserves exact half-open nanosecond membership, typed environment
policy, reported service, namespace and instance. Metric names, instrumentation
scope and metric dimensions remain native stream metadata; they do not become
record predicates.

This is tab-held service evidence. It has no immutable capture receipt. The
original metric basis and read clock remain distinct from fresh record evidence.
Shared resource identity and time provide context, not exemplars, contributing
requests, causality or an atomic cross-signal snapshot. Saving the child question
preserves ordinary filters, not custody of the originating metric result.

Only a quiescent current service owner may open a child. Creation-time action
leases, actual held interval admission and visible opener checks reject replaced,
hidden and stale callbacks. Return suspends the actual child visit before restoring
the service. A failed child read does not change the held metric evidence; a late
response must not publish after Return. No native or capture-store read belongs
to staging or returning.

This interval cut excludes gauge lookaround; the separate
[gauge contract](RETAINED_GAUGE_RECORD_CONTEXT.md) owns that gesture. Stock
observations, aggregate buckets, fleet percentiles and broad resource joins remain
outside these service interval actions. The interval path reuses existing rendering,
planning and fixed-record navigation; no endpoint or storage format is added.

See [architecture review](reviews/SERVICE_NATIVE_RECORD_CONTEXT_ARCHITECTURE_368.md)
and [design review](reviews/SERVICE_NATIVE_RECORD_CONTEXT_DESIGN_368.md).
