# Pod memory observations

Implemented and accepted in [journey 202](../journeys/202-pod-memory-storyboard.md).
This adds explicit pod gauge authority and the first memory defaults to
[Kubernetes discovery](KUBERNETES_RESOURCES.md).
[Compact receipt](evidence/pod-memory-202.json).
Reported phase/readiness requires a separate typed state contract.

## Supported profiles

The first receiver profile follows the published Collector Contrib
[kubeletstats v0.137.0 metadata](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.137.0/receiver/kubeletstatsreceiver/metadata.yaml)
and its generated resource metrics. Its scope is
`github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver`,
version `0.137.0`, with an explicitly empty instrumentation schema URL.

| Profile | Instrument | Shape and meaning |
| --- | --- | --- |
| Collector kubeletstats v0.137.0 | `k8s.pod.memory.usage` | Byte gauge: reported memory usage for one exact resource stream. |
| Collector kubeletstats v0.137.0 | `k8s.pod.memory.working_set` | Byte gauge: reported working set, retained independently from usage. |
| [OTel semconv v1.37](https://github.com/open-telemetry/semantic-conventions/blob/v1.37.0/docs/k8s/k8s-metrics.md) | `k8s.pod.memory.usage` | A separate usage-only candidate with the exact semconv instrumentation schema. It does not supply a working-set alias. |

Receiver and semantic-convention profiles are distinct. Names, schema URLs,
instrumentation scope/version and units are not silently migrated or combined.
A complete receiver pair is preferred for defaults, with at most two independent
reads. No usage/working-set percentage, capacity ratio, container sum, inferred
limit, memory pressure or pod health is calculated.

The existing gauge evaluator retains Float64 observation semantics. The pod
presenter withholds negative, fractional or unsafe byte values from its headline
and plotted copy; original results and raw source rows remain inspectable.
Metadata eligibility does not prove exporter conformance. Stale, missing,
conflicting and invalid latest values do not fall back to an older usable value.

## Identity and effects

Native definitions use `source.resource: {kind: kubernetes-pod, id: ...}` with
mandatory full exact-stream identity. The ID must equal the reported
`k8s.pod.uid`; the complete raw resource map, resource schema, instrumentation
scope/version/schema, metric/unit and point attributes remain pinned. Services
and service-namespace selectors are absent, not synthesized. Missing cluster
UID remains unverified context. Equal UIDs across different resource variants
do not become a verified physical pod or a cross-cluster join.

This authority supports gauges only. Existing service and host entry points
retain their validation and wire behavior; pod stocks, counters, histograms and
lifecycle measurements are not admitted by this slice. The service-native
record-context planner refuses resource authority rather than inventing a
service selector. Existing Kubernetes matching-resource log/span actions retain
their contextual-search contract.

**Observe metrics** is explicit. It discovers metric metadata for the selected
raw resource map, known resource-schema string and fixed window, then plans from
that retained metadata and reads at most two supported defaults. Discovery is a
source operation; planning is pure. The source request is bounded to 64 KiB;
the pure plan accepts the discovery envelope up to 8 MiB + 1 KiB so valid
multi-stream metadata is not trapped by a smaller request limit. Planning has a
two-second pre/post deadline, not an interruptible CPU or scan-cost guarantee.
Discovery retains at most 16 instrument descriptors and 25 stream previews per
descriptor, with explicit omissions. Planning selects its two defaults before
limiting visible candidates to 16; point-attribute streams remain independent. Merely inspecting resource attributes,
editing local facets or staging native inspection does not fetch metric data.

A trace variant's unknown resource schema (`null`) is not converted into an empty
schema. Select a variant with a retained schema value for native discovery;
retained empty strings are valid exact identity values. The legacy native
`metrics.Stream` representation omits a zero/empty resource-schema string. That
omission is interpreted only within the independently validated known-empty
lookup context; an explicit null or unknown lookup schema is not accepted. Absence of returned
metadata is not proof that the pod or an instrument does not exist.

The panel reuses supplied-data native displays and the existing native
Preview/Run, individual capture and retained navigation owners. Source times,
observation times and capture custody times stay separate. Leaving/changing the
selected row, facets, fixed scope or tab withdraws pending panel work. Back from
native inspection restores held Kubernetes filters and focus without rerunning.
There is no composite dashboard capture or new mutable metric store.

Fresh pod YAML authoring canonicalizes absent/null service lists and empty
exact-stream attribute maps before hashing, following the existing explicit host
authoring boundary. Direct JSON questions, legacy service packages and stored
capture bytes are not rewritten.

## State and later work

Collector [k8scluster v0.137.0](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.137.0/receiver/k8sclusterreceiver/metadata.yaml)
reports `k8s.pod.phase` as an empty-unit ordinal gauge: 1 Pending, 2 Running,
3 Succeeded, 4 Failed, 5 Unknown. This is not a boolean health signal. Its
container-ready gauge is container-scoped and differs from current semconv
typing. Neither is an alias for pod readiness, and the current native scalar
kernel's nonempty-unit contract is not relaxed to accommodate them.

The [typed phase contract](POD_STATES.md) now supports exact-source native
execution and individual capture/replay, plus exact-profile discovery and a
compact categorical view. Its separate live acceptance is tracked in that contract.
Container readiness, requests/limits, CPU calculations, ownership and workload
rollups remain separate deliveries. Generated ordinary OTLP acceptance passes
through the actual collector and ClickHouse; it is not a receiver scrape, live
Kubernetes cluster or RBAC configuration test.

## Accepted workflow

The primary receiver resource has no service/instance identity and a known empty
resource-schema string. Usage changes from 65,536 to 32,768 bytes; working set is
16,384 bytes, observed independently. Exact receiver scope/version/empty
instrumentation schema are retained. Separate partial-receiver and semconv-only
resources, stale/conflicting/missing observations and negative/fractional/unsafe
byte values are explicitly read through the UI. Stale values stay visibly labeled
stale; unknown samples do not fall back, and invalid byte counts are masked only
in the display copy.

The default path costs one resource inventory operation, one explicit metric
metadata read and two native reads (four source operations/one pure plan).
Native staging and retained Back are source-free. Explicit native Run and fresh
individual capture advance to six source operations. Eight deliberately selected
boundary panels each add one metadata and one gauge read, ending at 22 source
operations/11 pure plans; this is not a larger automatic default budget.

The individual capture remains 6,062 bytes with its original full stored SHA-256
through public and UI reopening. Browser frozen reopening attempts no telemetry
POSTs, and a separate HTTP handler check proves reopening with its source
unavailable. That is not a claim of a live ClickHouse outage. Five actual held
response cases check panel ownership, and prior host/Kubernetes owner regressions
pass. Independent raw/profile/mask/clock/journal/custody validation rejects
fourteen changed controls. The storyboard links actual desktop/mobile frames and
records the corrected actor and observer diagnostics.
