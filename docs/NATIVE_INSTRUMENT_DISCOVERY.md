# Native instrument discovery and convention candidates

Implemented E1 discovery and typed-plan continuation in the existing Measurements surface,
including the existing service-to-discovery handoff. No manifest, goo annotations
or Go-specific instrumentation is required. The Go discovery adapter still owns
bounded source reads, supported bindings and exact package generation. It now also
supplies a versioned metadata capability plan and observed per-stream presence.
The browser presents that plan; it does not classify conventions independently.
This is metadata planning, not verification of metric values or time coverage.

## User flow

From a service, choose its metric-discovery continuation; environment, service,
optional namespace and fixed incident window are staged. Alternatively open
Measurements → Discover metrics from an application and choose a service and
environment for a fresh two-minute inventory. Press Discover explicitly.

The compact table shows instrument name, observed kind/unit/count and whether a
package can be previewed. Up to 25 identities appear per local page. More/Previous
appear only when needed; paging and Metadata and support disclosures read no
sources and publish nothing. The shared `goo-table` contains horizontal overflow
on narrow screens. Unsupported shapes in these inventoried families remain visible. Summary and
exponential-histogram tables are outside this inventory.

Build measurement package sends the existing exact definition to the existing Go
preview validator. It does not execute or save a measurement. Preview canonicalizes
the editor to YAML. Run measurement performs a separate, explicit source read;
Save package revision and Capture evidence remain separate actions. A fixed service
handoff retains its window; a manually fresh discovery keeps the existing separate
package-lookback/preview-window behavior.

## Pinned HTTP hint

The rule set is [OpenTelemetry semantic conventions v1.30.0 HTTP metrics](https://github.com/open-telemetry/semantic-conventions/blob/v1.30.0/model/http/metrics.yaml):
`http.server.request.duration`, histogram, unit `s`. A matching name is an HTTP
**candidate**, not a conformance verdict. A wrong kind/unit is explained while
retaining any valid generic package binding; no conversion or schema migration
occurs. A matching schema URL is evidence of metadata only. Missing or other schema
URLs remain unverified rather than being silently upgraded.

The hints distinguish:

| Supplied information | Meaning |
| --- | --- |
| Name/kind/unit/schema and observed dimension keys | Bounded observed metadata; not declarations or complete coverage. |
| Go `supported` plus a supplied definition | Eligible for package preview under the existing kernels. |
| HTTP method/status key observed | A key appeared somewhere in this inventory, not on every record or export. |
| Successful Preview | Definition and source plan validated without telemetry execution. |
| Explicit Run | Existing kernels evaluate actual samples, exact streams, boundaries, reset/temporality and histogram coverage. Metadata hints remain labelled not evaluated; they do not acquire a verified badge. |

Missing status dimensions withhold an HTTP error-rate suggestion. Even when a status
key is observed, no error policy, denominator or error-rate view is invented.
Latency remains native histogram bucket ranges per compatible exact stream; request
rates and fleet percentiles are not inferred. The catalog summarizes supplied convention candidates; a generic-only catalog
does not imply missing HTTP or messaging instrumentation.
Declared manifests continue to have their existing separate service inventory role.

## Ownership

```mermaid
flowchart LR
  S[Service scope or explicit Measurements scope] -->|Discover explicitly| G[Existing Go metric discovery]
  G --> O[Observed identities and exact package candidates]
  O --> H[Go typed capability plan and stream presence]
  H --> T[Shared table and local metadata disclosure]
  T -->|Build package| P[Existing Go Preview and YAML editor]
  P -->|Explicit Run| R[Existing native measurement kernels]
  R --> V[Native stream results and coverage]
  P -->|Explicit Save| D[Package store]
```

`ui/instrument-capabilities.js` is a pure presenter for the Go capability plan.
It performs no reads, publication, convention classification, binding rewrites or
verification; unsupported/missing plan versions require another discovery. `ui/metric-discovery.js` owns only catalog presentation
and local paging, with Build passed back to the existing measurement owner.
Discovery cancellation, generation invalidation and namespace/scope handling remain
in `ui/measurements.js`. No component lifecycle starts a request.

## Executable storyboard

Journey 130 emits language-neutral OTLP for a manifest-free application: an HTTP
histogram in seconds, the same name in milliseconds, a custom job counter and a
non-monotonic sum. It stages service scope without reading, discovers all four,
explores metadata/refusals, previews the exact seconds binding without execution,
then runs once in the original fixed window. It checks no writes, no page overflow,
actual Go results through the existing Quint discovery replay and rejecting controls,
and disclosure-only Bombadil exploration. A supplied-catalog browser case checks
local paging, exact Build intent and zero API requests. Hegel exercises the actual
Go planner across kinds/units and exact-stream permutations without mutating supplied metadata.

![Observed instrument capabilities](journeys/screenshots/130a-observed-instrument-capabilities.png)

![Selected HTTP metadata hints](journeys/screenshots/130c-http-metadata-hints.png)

![Contained catalog on mobile](journeys/screenshots/130b-instrument-catalog-mobile.png)

## Acceptance and limits

The [E1 acceptance audit](E1_ACCEPTANCE.md) accepts bounded native discovery,
HTTP/messaging metadata profiles and observed stream presence. It distinguishes
preview eligibility from the existing kernels’ interval/value validation. Verified error/denominator policies and typed
relations require concrete supporting data; they are not offered by these profiles. Existing
metric support limits (including unsupported non-monotonic sums and exponential
histograms) are not widened by presentation hints. No automatic dashboard or
persistent convention declaration is created by this cut.

## Typed server plan and observed stream presence

`POST /metric-discovery` now adds `capabilityPlan` to each instrument. Its version is
`goo/instrument-capability-plan/v1alpha1`; it retains `verification: not-run`.
The Go planner owns the pinned HTTP/messaging profiles, generic fallback, shape/eligibility
and diagnostic text. The browser accepts this version without recomputing rules.
Old clients may ignore the additive field; new clients disable Build when no
supported plan is supplied. Request semantics and generated definition IDs remain
unchanged. [OpenAPI](API.md) contains the reflected response types.

The same bounded source query now retains resource/attribute identity and resource
schema URL. Exact streams remain separate across resource, attributes, scope,
version, schema, name and unit; their IDs are scoped to their parent instrument
kind/temporality. Missing and empty instance IDs remain distinct diagnostics.
The HTTP profile watches method/status keys and reports present-nonempty, empty
and missing stream counts. Generic instruments retain stream identity inspection
without inheriting HTTP assumptions. A union of keys no longer hides a missing key
on another stream. Duplicate observations increase record counts, not stream counts.

Only the first 25 sorted streams per instrument are displayed, with an explicit
omitted count; presence totals use all streams in the bounded read. Legacy source
fixtures without resource/attribute metadata report unavailable presence rather
than fabricated counts. The original 10,000-row/8 MiB source budgets and overflow
refusal remain. No extra source query is made. Presence is not interval continuity,
valid attribute values, error classification, conformance or fleet comparability.

Journey 131 ingests three HTTP streams with differing present/empty/missing keys
and one missing instance ID. It checks all three states, exact identity, duplicate
observation counts, disclosure-only browsing and mobile containment. Its observed
Quint trace compares the actual server plan against independent ClickHouse identity
rows and rejects seven mutations (counts, identities, eligibility, effects and extra
reads). Go unit tests and Hegel cover permutation stability, exact bindings, aliasing,
missing/empty identity and the 25-stream display cap without changing totals.

![Observed stream presence](journeys/screenshots/131a-observed-stream-presence.png)

![Stream presence on mobile](journeys/screenshots/131b-stream-presence-mobile.png)

## Messaging processing candidate

The Go planner also recognizes `messaging.process.duration` against the
[experimental v1.30.0 messaging metrics definition](https://github.com/open-telemetry/semantic-conventions/blob/v1.30.0/model/messaging/metrics.yaml).
The pinned shape is a histogram in seconds. A millisecond instrument retains its
eligible generic binding, without conversion or a claim to match the convention.
The profile leaves verification `not-run`, including when a schema URL matches.

The bounded presence matrix watches messaging system, operation type/name,
destination name/template and consumer group. System is required by the pinned
convention; processing operations must have type `process`. The matrix reports
presence, not value conformance. Destination and consumer identity requirements
are conditional: absent labels do not independently prove invalid telemetry.
Arbitrary additional labels remain in the exact stream inspector, without
expanding the matrix. Observed processing durations cannot establish unique
message deliveries, queue lag, success, producer latency or trace relationships.

```mermaid
flowchart LR
  OTLP[Manifest-free OTLP processing histogram] --> CH[Existing ClickStack metrics source]
  CH --> Discover[Bounded discovery and exact stream identity]
  Discover --> Plan[Go experimental messaging candidate and presence]
  Plan --> UI[Passive catalog and local stream inspection]
  UI --> Preview[Explicit package preview]
  Preview --> Run[Explicit Run: existing duration kernel validation]
```

Journey 132 (`journeys/messaging-profile.spec.js`) emits ordinary OTLP with three
streams and four records, including empty and missing messaging labels. The user
scopes discovery to the consumer, opens its candidate metadata and inspects exact
stream presence on desktop and mobile. Inspection performs no evaluation or
publication. An independent ClickHouse identity read supplies the observed Quint
replay oracle, with seven deliberately invalid controls; Bombadil explores the
same manifest-free consumer through the shared discovery actions/invariants.

![Messaging stream presence](journeys/screenshots/132a-messaging-stream-presence.png)
![Messaging presence on mobile](journeys/screenshots/132b-messaging-presence-mobile.png)


Journey 133 (`journeys/instrument-acceptance.spec.js`) completes explicit messaging
Preview/Run with fully covered and incomplete windows, then refuses unidentified
streams and fractional custom counters. Preview and metadata inspection do not
evaluate telemetry; every catalog stays `not-run` after execution. Refusals leave
no result to capture. The [acceptance map](E1_ACCEPTANCE.md) records these boundaries
and the inventory's half-open export window versus Run's interval-boundary reads.
