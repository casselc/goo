# Observed native instruments and HTTP duration hints

Implemented first E1 cut: compact discovery in the existing Measurements surface,
including the existing service-to-discovery handoff. No manifest, goo annotations
or Go-specific instrumentation is required. The Go discovery adapter still owns
bounded source reads, supported bindings and exact package generation. This cut
adds informational browser presentation, not a new semantic planner or API.

## User flow

From a service, choose its metric-discovery continuation; environment, service,
optional namespace and fixed incident window are staged. Alternatively open
Measurements → Discover metrics from an application and choose a service and
environment for a fresh two-minute inventory. Press Discover explicitly.

The compact table shows instrument name, observed kind/unit/count and whether a
package can be previewed. Up to 25 identities appear per local page. More/Previous
appear only when needed; paging and Metadata and support disclosures read no
sources and publish nothing. The shared `goo-table` contains horizontal overflow
on narrow screens. Unknown or unsupported instruments remain visible.

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
rates and fleet percentiles are not inferred. An absent HTTP duration name is
reported as not observed in this bounded window, not as missing instrumentation.
Declared manifests continue to have their existing separate service inventory role.

## Ownership

```mermaid
flowchart LR
  S[Service scope or explicit Measurements scope] -->|Discover explicitly| G[Existing Go metric discovery]
  G --> O[Observed identities and exact package candidates]
  O --> H[Pure versioned metadata hints]
  H --> T[Shared table and local metadata disclosure]
  T -->|Build package| P[Existing Go Preview and YAML editor]
  P -->|Explicit Run| R[Existing native measurement kernels]
  R --> V[Native stream results and coverage]
  P -->|Explicit Save| D[Package store]
```

`ui/instrument-capabilities.js` is pure and performs no reads, publication, binding
rewrites or verification. `ui/metric-discovery.js` owns only catalog presentation
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
JS hint function across kinds/units/temporality without mutating supplied metadata.

![Observed instrument capabilities](journeys/screenshots/130a-observed-instrument-capabilities.png)

![Selected HTTP metadata hints](journeys/screenshots/130c-http-metadata-hints.png)

![Contained catalog on mobile](journeys/screenshots/130b-instrument-catalog-mobile.png)

## Remaining E1 scope

This does not complete E1. Typed server-side convention capability plans, per-stream
HTTP dimension coverage, verified error/denominator policies, a database-pool or
messaging case and any supporting typed relations remain follow-up work. Existing
metric support limits (including unsupported non-monotonic sums and exponential
histograms) are not widened by presentation hints. No automatic dashboard or
persistent convention declaration is created by this cut.
