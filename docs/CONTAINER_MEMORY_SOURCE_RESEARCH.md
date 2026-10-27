# Container memory source research

Public source inspected 2026-10-27 at Collector Contrib tag `v0.137.0`. This is source research and a proposed integration contract; it is not evidence of running either receiver, Kubernetes access, completeness or shared snapshots.

| Role | Exact metric | Receiver | Physical type | Unit |
| --- | --- | --- | --- | --- |
| Working set | `container.memory.working_set` | kubeletstats | integer gauge | `By` |
| Request | `k8s.container.memory_request` | k8scluster | integer gauge | `By` |
| Limit | `k8s.container.memory_limit` | k8scluster | integer gauge | `By` |

These names are Development metrics, enabled by default in the pinned metadata. There is no alias from `k8s.container.memory.working_set`. The separate kubelet metrics `k8s.container.memory_limit_utilization` and `k8s.container.memory_request_utilization` are disabled-by-default double gauges with unit `1`; they are not substitute raw-byte evidence.

`RecordSpecMetrics` iterates present Kubernetes Requests/Limits map entries and emits `Quantity.Value()`. A missing memory entry produces no sample; an explicitly present zero produces a reported zero. Neither absence nor zero establishes unlimited memory. Spec metrics are recorded before matching container status. A container without a matching runtime status can therefore emit requests/limits without a usable runtime ID. Do not synthesize an ID from pod UID and container name.

The cluster emitter sets pod UID/name, node, namespace, container name, stripped status container ID and image context. The kubelet resource builder sets pod UID/name, namespace and container name; container ID is additional configured metadata and can be unavailable. Its metadata reader strips runtime prefixes when retrieving IDs. Goo must retain the ID actually reported in telemetry verbatim, rather than repeat either receiver's normalization.

Both generated builders set their instrumentation scope name to the exact receiver import path and scope version to the collector's `BuildInfo.Version`. A `0.137.0` profile deliberately requires that reported version; the source tag alone cannot establish a deployment's version. Neither builder sets instrumentation schema. The cluster generated builder sets **resource** schema from semconv `v1.18.0`; the kubelet builder does not set a resource schema. Processors may further alter resource maps/schemas, so source evidence must preserve the actually reported variants.

## Bounded integration contract

Reuse exact resource discovery and explicit native questions. Keep existing automatic default budgets. A scalar container question requires literal nonempty reported container ID, pod UID and container name, exact full stream identity and a supported receiver/metric/unit profile. Missing runtime identity remains unsupported for container authority.

Independent contextual byte observations can be displayed together only after their required container/pod/name identity and cluster UID/name/namespace states agree. Retain every full resource/schema/point-attribute variant and each independent observation/read clock. This contextual alignment is not full stream equality, ownership or a same-snapshot join.

Percentages need a separate stricter gate: identical full resource maps/resource schema/point attributes, original fixed window, exact latest observation timestamp, matching freshness policy, fresh known observations, exact nonnegative safe-integer raw byte values and a positive denominator. Receiver scope/name differences stay explicit. Requests are declared requests, not guaranteed available capacity. Missing, stale, conflicting, zero-denominator or mismatched evidence yields an explicit unavailable reason; no hold-forward denominator, tolerance join, aggregate, name-based restart merge or healthy/unlimited inference. Normal native capture/reopen retains independent evidence unchanged.

## Pinned public evidence

SHA256 values below were computed from the downloaded public response bytes; URLs pin the inspected tag, not a running deployment.

- [kubeletstatsreceiver/metadata.yaml](https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector-contrib/v0.137.0/receiver/kubeletstatsreceiver/metadata.yaml) — SHA256 `d61170d30282e42a454073d0dcd7715c2cd01646c4adc8a3ae7e773d0a0faede`.
- [k8sclusterreceiver/metadata.yaml](https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector-contrib/v0.137.0/receiver/k8sclusterreceiver/metadata.yaml) — SHA256 `a749eb22a0cade8d0a2f92b2bd79b879dcd039c7202d140423ac8eaecff646f6`.
- [k8sclusterreceiver/internal/container/containers.go](https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector-contrib/v0.137.0/receiver/k8sclusterreceiver/internal/container/containers.go) — SHA256 `fb802ca574993a8b89dc6ded0ea0562c21e2466387c8a0743d23edca7fbabe16`.
- [kubeletstatsreceiver/internal/kubelet/resource.go](https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector-contrib/v0.137.0/receiver/kubeletstatsreceiver/internal/kubelet/resource.go) — SHA256 `fca0109ddce209d08cbc1fd539baa7dc15c59129b1c8b414e2c5ac441c58feef`.
- [kubeletstatsreceiver/internal/kubelet/metadata.go](https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector-contrib/v0.137.0/receiver/kubeletstatsreceiver/internal/kubelet/metadata.go) — SHA256 `bb0e359978826f17f6fd72d3fd0b6fd097ec19f28620115b75666f77e9765146`.
- [kubeletstatsreceiver/internal/metadata/generated_metrics.go](https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector-contrib/v0.137.0/receiver/kubeletstatsreceiver/internal/metadata/generated_metrics.go) — SHA256 `d171a49946550cf41e471d727b73604cdffccfc88d1b3674933b0f008ac067d9`.
- [k8sclusterreceiver/internal/metadata/generated_metrics.go](https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector-contrib/v0.137.0/receiver/k8sclusterreceiver/internal/metadata/generated_metrics.go) — SHA256 `5c6059c5f718a08dcaf7a627ab9b0d709b9425222f052c053068609ab96cc9e9`.
- [kubeletstatsreceiver/internal/metadata/generated_status.go](https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector-contrib/v0.137.0/receiver/kubeletstatsreceiver/internal/metadata/generated_status.go) — SHA256 `1210252079f34002407458b325bd286490f8895228de83493dd3ccb28cccd2f2`.
- [k8sclusterreceiver/internal/metadata/generated_status.go](https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector-contrib/v0.137.0/receiver/k8sclusterreceiver/internal/metadata/generated_status.go) — SHA256 `eac467b5c4cefcde64dd24e922ae41dd2b862d6a1f52dd665a497290df33abd5`.
