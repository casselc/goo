# Container CPU source research

Public Collector Contrib `v0.137.0` source inspected 2026-10-27. This document records source semantics and an integration recommendation, not execution of a receiver or access to Kubernetes. All instruments below have Development status in the pinned metadata.

| Role | Exact instrument | Receiver | Type | Unit | Default |
| --- | --- | --- | --- | --- | --- |
| Usage | `container.cpu.usage` | kubeletstats | double gauge | `{cpu}` | enabled |
| Request | `k8s.container.cpu_request` | k8scluster | double gauge | `{cpu}` | enabled |
| Limit | `k8s.container.cpu_limit` | k8scluster | double gauge | `{cpu}` | enabled |
| Cumulative CPU time, separate meaning | `container.cpu.time` | kubeletstats | cumulative monotonic double sum | `s` | enabled |

`addCPUMetrics` records usage only when CPUStats and UsageNanoCores are present. It converts UsageNanoCores to float64 cores by division by 1,000,000,000. The metadata describes usage as averaged over the sample window. The collector timestamp is the scrape observation clock; this gauge does not export explicit sample-window bounds. Goo must describe the observed exported usage rather than invent an instantaneous measurement or an integration interval.

CPU usage is already a gauge in cores: do not apply a counter rate, reset detector or cumulative-time subtraction. The separate CPU time sum converts UsageCoreNanoSeconds to float64 seconds and uses container start-time override. Deriving usage from that sum would require separate cumulative/reset/gap/interval semantics and is outside this slice.

Cluster `RecordSpecMetrics` iterates only present Requests/Limits entries, converts CPU Quantity.MilliValue() to float64 cores by division by 1000, and records the result. Missing entries emit no sample; explicit zero is a reported zero. Neither means unlimited or healthy. The exported double is not the original exact Kubernetes quantity. Fractional cores are expected; integer byte-domain validation would incorrectly reject useful CPU evidence.

The kubelet's optional `k8s.container.cpu_limit_utilization`, `k8s.container.cpu_request_utilization` and `k8s.container.cpu.node.utilization` are double gauges with unit `1`, disabled by default. They are emitted only for positive known denominators. They are separate supplied measurements, not fallback evidence for Goo's explicit usage/request/limit pairing.

## Identity and clock constraints

Reuse the [memory source findings](CONTAINER_MEMORY_SOURCE_RESEARCH.md). Kubelet resources report pod UID/name/namespace and container name; runtime container ID depends on extra metadata. Cluster spec metrics can be emitted without a matching runtime status and hence without usable container ID. Never synthesize an ID, strip its runtime prefix again, or merge restarts by pod/name.

The builders use receiver-specific instrumentation scopes and actual collector BuildInfo.Version; neither sets an instrumentation schema. Cluster generated code sets resource schema from semconv v1.18.0; kubelet generated code does not set a resource schema. Full resource maps also differ. Preserve actual reported schemas/maps and exact stream attributes. Equal runtime IDs alone cannot establish full stream equality or a common observation snapshot.

## Smallest useful contract

Reuse the explicit, bounded memory workflow: supported CPU candidates do not increase automatic defaults; the user chooses exact held resource/schema variants and explicitly loads independent signals. Keep literal container ID, pod UID, container name and cluster/namespace qualifier states; each observation retains quality, timestamp, read clock and evidence basis. Individual captures remain independent and frozen.

Use bounded strict decimal syntax and exact rational arithmetic over the exported source decimal for nonnegative fractional cores. Reject malformed, nonfinite, negative or unsafe-domain values explicitly; do not coerce to integer cores or infer an original nanocore count. State any practical magnitude/precision bounds as Goo policy. Millcores can be a display conversion, retaining the exported decimal as exact evidence. Do not describe the rational as greater precision than the receiver's float64 source.

Reuse strict percentage eligibility: matching full raw resource maps/resource schema/point attributes, same original fixed window, exact latest observation timestamp and freshness policy, fresh known samples and positive denominator. Differing receiver scopes stay explicit, as do independent read bases. Ratios are descriptive exported usage/request or usage/limit, can exceed 100%, and are not CPU saturation, throttling, available capacity, guaranteed allocation or application health. Missing or zero denominator yields unavailable, never infinity, zero usage or unlimited capacity. Ordinary crossreceiver resource/schema/time differences should withhold the ratio while leaving independent values visible.

## Pinned public evidence

SHA256 values were computed over downloaded public response bytes. Tag source is not proof of a deployed collector version.

- [kubeletstatsreceiver/metadata.yaml](https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector-contrib/v0.137.0/receiver/kubeletstatsreceiver/metadata.yaml) — SHA256 `d61170d30282e42a454073d0dcd7715c2cd01646c4adc8a3ae7e773d0a0faede`.
- [k8sclusterreceiver/metadata.yaml](https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector-contrib/v0.137.0/receiver/k8sclusterreceiver/metadata.yaml) — SHA256 `a749eb22a0cade8d0a2f92b2bd79b879dcd039c7202d140423ac8eaecff646f6`.
- [kubeletstatsreceiver/internal/kubelet/cpu.go](https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector-contrib/v0.137.0/receiver/kubeletstatsreceiver/internal/kubelet/cpu.go) — SHA256 `a2e1a392850d088f3e766eeded9a9454f52973c98358aff5ce72eaf1fd4ce8f4`.
- [k8sclusterreceiver/internal/container/containers.go](https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector-contrib/v0.137.0/receiver/k8sclusterreceiver/internal/container/containers.go) — SHA256 `fb802ca574993a8b89dc6ded0ea0562c21e2466387c8a0743d23edca7fbabe16`.
- [kubeletstatsreceiver/internal/kubelet/resource.go](https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector-contrib/v0.137.0/receiver/kubeletstatsreceiver/internal/kubelet/resource.go) — SHA256 `fca0109ddce209d08cbc1fd539baa7dc15c59129b1c8b414e2c5ac441c58feef`.
- [kubeletstatsreceiver/internal/kubelet/accumulator.go](https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector-contrib/v0.137.0/receiver/kubeletstatsreceiver/internal/kubelet/accumulator.go) — SHA256 `cc02c1b524c97b9b21cb9fc1b0b4735fc61e20cb256701917eddd7bcb2b926d5`.
- [kubeletstatsreceiver/internal/metadata/generated_metrics.go](https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector-contrib/v0.137.0/receiver/kubeletstatsreceiver/internal/metadata/generated_metrics.go) — SHA256 `d171a49946550cf41e471d727b73604cdffccfc88d1b3674933b0f008ac067d9`.
- [k8sclusterreceiver/internal/metadata/generated_metrics.go](https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector-contrib/v0.137.0/receiver/k8sclusterreceiver/internal/metadata/generated_metrics.go) — SHA256 `5c6059c5f718a08dcaf7a627ab9b0d709b9425222f052c053068609ab96cc9e9`.
