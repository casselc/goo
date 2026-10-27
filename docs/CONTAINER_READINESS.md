# Container identity and reported readiness

Implemented: exact container authority, pinned readiness decoding, existing native
API execution/capture/replay, pod-bearing resource discovery and native UI.
Ordinary OTLP/ClickHouse, browser, ownership, Bombadil and independent raw/custody
acceptance pass; [journey 204b](../journeys/204b-container-readiness-storyboard.md)
records actual costs, frames and limits.

## Receiver evidence and identity

This profile is pinned to Collector Contrib v0.137.0:

- [Metric metadata](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.137.0/receiver/k8sclusterreceiver/metadata.yaml) declares `k8s.container.ready` as a unitless integer gauge.
- [Container emission](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.137.0/receiver/k8sclusterreceiver/internal/container/containers.go) records `ContainerStatus.Ready` and reports `container.id`, `k8s.pod.uid` and `k8s.container.name` on the resource.
- [Pod status preparation](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.137.0/receiver/k8sclusterreceiver/internal/pod/pods.go) omits statuses with empty container IDs. Lack of a readiness observation is not a reported false value.

The receiver strips a runtime prefix before reporting its container ID. Goo does
not strip, synthesize or otherwise normalize any reported ID. Resource binding
`{kind: kubernetes-container, id: <reported container.id>}` uses that actual value.
The exact stream must also contain bounded nonempty pod UID and container name;
pod name is optional context. Full raw resource map/schema, scope/version/schema,
metric/unit and point attributes remain isolated. A restart with a new ID is a
different authority even if the pod and container names stay the same. Repeated
IDs in different resource variants never establish global physical identity or
verified ownership.

## State contract

Use `kind: state`, `stateProfile: k8scluster-container-ready/0.137.0`, and
`grain: per-stream`. This initial container authority admits only this readiness
state profile, not scalar gauges, stocks, counters, histograms or lifecycle data.
It requires an exact stream, empty service selection and no namespace selector.
It fabricates no service/instance labels.

The physical profile requires scope
`github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver`,
version `0.137.0`, empty instrumentation schema, metric `k8s.container.ready` and
empty unit. Semconv count-like instruments, other receiver versions and
unit aliases are not interchangeable with this gauge.

| Code | Supplied category | Observation meaning |
| --- | --- | --- |
| 0 | Not ready | A reported false status, still observed telemetry |
| 1 | Ready | A reported true status, still observed telemetry |
| Other finite value | No decoded category; raw code retained | Unsupported domain code |
| Missing/conflicting sample | No code or label | Unknown observation quality |

Staleness remains separate from code meaning and is evaluated at the fixed
window end. Latest missing/conflicting/unsupported values never fall back to an
older Ready value. Empty input creates no series and overall `no_data`; it does
not invent Not ready. Unsupported values make overall result quality unknown,
while their individual sample preserves observed/stale quality.

This is reported container readiness, not pod readiness, application health or
proof that a readiness probe exists. Kubernetes can report Ready without a
configured probe. No ready-container ratio, fleet total, Boolean aggregation or
phase arithmetic is inferred.

The private categorical kernel shares timestamp grouping, conflicts, freshness
and no-recorded-value handling with pod phase. It uses an explicit profile/code
mapping. The source adapter checks the original bounded decimal before assigning
a supported code: tiny nonzero values that round to zero cannot manufacture
Not ready, and fractions rounding to one cannot manufacture Ready. Ordinary
scalar kernels retain their nonempty-unit requirement.

## Execution and custody

The existing compiler selects the gauge table using typed values, reported
`container.id`, environment, exact resource/schema and full point attributes.
Source rows must explicitly retain known physical identity strings, including
known-empty schemas/unit. Readiness also requires explicitly reported resource/
attribute objects with string values and flags: absent/null fields cannot become
empty maps or zero flags. Existing source bounds refuse partial measurements.
Readiness uses `states`, not a scalar result section. Existing individual capture
and source-free replay verify the original raw rows, question and read clock;
modified category labels fail replay. Generic categorical comparisons remain
unsupported.

```mermaid
flowchart LR
  A[Reported container ID plus exact resource context] --> B[Pinned readiness definition]
  B --> C[Parameterized exact gauge query]
  C --> D[Observation quality and raw code]
  D --> E[Ready or Not ready only for supported codes]
  E --> F[Individual immutable capture]
  F --> G[Source-free replay]
```

## Discovery and interaction

Kubernetes discovery retains full pod-bearing resource variants. Observe metrics
uses the existing exact known-schema metadata route; a readiness candidate needs
reported container ID, pod UID and container name in that resource. The physical
instrument remains a gauge; its executable definition is a categorical state
with container authority and `container-resource-stream` dashboard grain.
No synthetic container or service identity is created from names.

Memory defaults retain their existing profile priority and two-read budget.
On cluster-only variants, one phase and one readiness stream can be defaults;
readiness alone uses one. Candidates remain deterministic and bounded at 16,
with at most 16 instruments and 25 preview streams per instrument. Omission is
not absence. Reads remain independent rather than an atomic state snapshot.

The shared passive categorical component renders supported Ready/Not ready
labels and separate observation quality, UTC times and raw-code history. App
adapters validate pinned profiles, exact authority and stream identity. Explicit
Preview stages native inspection; Run and fresh capture perform source reads.
Retained Back and frozen capture reopening do not reread telemetry. Components
perform no fetching or navigation on their own. No numeric state chart or health
color is introduced.

```mermaid
flowchart LR
  A[Exact pod-bearing resource variant] --> B[Observe metrics metadata read]
  B --> C[Pure bounded plan]
  C --> D[At most two independent default reads]
  D --> E[Compact category plus quality]
  C --> F[Explicit Preview]
  F --> G[Native Run or fresh capture]
  G --> H[Immutable source-free reopen]
  F --> I[Retained Back without read]
```

The [API storyboard](../journeys/204a-container-readiness-storyboard.md) records
the native prerequisite. Discovery/UI live acceptance is tracked separately in
[journey 204b](../journeys/204b-container-readiness-storyboard.md).
