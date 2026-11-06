# Pinned current RPC convention research — 357

Research retrieved from the official OpenTelemetry semantic-conventions repository
on 2026-11-06. GitHub's published latest-release redirect resolved to
[v1.45.0](https://github.com/open-telemetry/semantic-conventions/releases/tag/v1.45.0).
`git ls-remote --tags` resolved that tag to
`fb30ca17193b551473b1f43c315482918ab6987c`. The research below uses that immutable
commit, not a guess about a moving latest version or the website's current build.
Network reads used the managed proxy and verified TLS; no credential was needed.

## Official sources

- [RPC spans](https://github.com/open-telemetry/semantic-conventions/blob/fb30ca17193b551473b1f43c315482918ab6987c/docs/rpc/rpc-spans.md), including incoming server semantics and attribute requirement levels.
- [RPC metrics](https://github.com/open-telemetry/semantic-conventions/blob/fb30ca17193b551473b1f43c315482918ab6987c/docs/rpc/rpc-metrics.md).
- [RPC metric model](https://github.com/open-telemetry/semantic-conventions/blob/fb30ca17193b551473b1f43c315482918ab6987c/model/rpc/metrics.yaml), instrument/type/unit and required attribute declarations.
- [RPC attribute registry](https://github.com/open-telemetry/semantic-conventions/blob/fb30ca17193b551473b1f43c315482918ab6987c/docs/registry/attributes/rpc.md), including deprecated names.
- [gRPC](https://github.com/open-telemetry/semantic-conventions/blob/fb30ca17193b551473b1f43c315482918ab6987c/docs/rpc/grpc.md) and [recording errors](https://github.com/open-telemetry/semantic-conventions/blob/fb30ca17193b551473b1f43c315482918ab6987c/docs/general/recording-errors.md).

## Span population

The pinned generic RPC span document is Stable. An incoming RPC span represents
the server-side request lifecycle; streaming RPC spans should cover the stream's
full lifetime. Its span kind **MUST be SERVER** (the ClickStack-shaped source
reports `Server`). It is not a per-message or unique-request count.

| Attribute | Pinned requirement and meaning | Consequence for Goo |
| --- | --- | --- |
| `rpc.system.name` | Required string RPC system; known values include `grpc`, with custom values allowed | Use this current key, without union/fallback to `rpc.system`. Goo's observed profile requires a reported nonempty value; this is a bounded product observation policy, not conformance certification. |
| `rpc.method` | Conditionally required **if available**; fully-qualified logical RPC interface method | Group literal values. Missing, present empty and `_OTHER` remain distinct; do not require a method merely to recognize current RPC records. |
| `rpc.method_original` | Conditionally required only when different from `rpc.method` | Do not substitute it into grouping or default to unbounded original methods. |
| `rpc.status_code` | Conditionally required if available; meaning depends on RPC system | Preserve as reported telemetry; do not derive a universal success/failure policy. |
| `error.type` | Conditionally required if operation failed | Distinct from recorded span status; no generic protocol-error synthesis in this cut. |

The method should include the RPC service/interface, for example
`com.example.ExampleService/exampleMethod`. The old `rpc.service` attribute is
Deprecated and its value should now be included in `rpc.method`. Do not concatenate
an old service attribute onto a current method or parse it into another identity.
For unrecognized methods, the convention prescribes `_OTHER`; this is a real
reported category, not missing data. Goo must not synthesize `_OTHER` for absent
attributes. Span naming uses the method when available and not `_OTHER`, otherwise
the system name, but a span name is not a replacement for an absent method field.

The registry marks `rpc.system` Deprecated in favor of `rpc.system.name`. The
stable migration guidance explicitly permits dual emission (`rpc/dup`), so old and
current observations can coexist. This justifies distinct supported observation
profiles, not aliases, automatic deduplication or a compatibility obligation.

## Native duration

`rpc.server.call.duration` is a Stable recommended **Histogram**, unit **`s`**;
the model declares double measurement values. It measures incoming RPC duration.
When emitted alongside a server span, its measured duration SHOULD equal that
span's duration; this does not establish a join between arbitrary captured metric
and span rows. The client instrument is separately named
`rpc.client.call.duration` and is not a server instrument.

The metric's `rpc.system.name` is required; `rpc.method` and `rpc.status_code` are
conditional. The RPC convention does **not** prescribe one export aggregation
temporality. Goo's existing native histogram engine accepts its supported delta
and cumulative forms (OTLP 1/2), retaining resets, intervals, per-stream identity
and coverage. This is a Goo capability bound, not a new convention requirement.
Do not convert legacy `rpc.server.duration` milliseconds into the current series,
accept the wrong unit, blend streams or average scalar percentiles.

Instrument metadata can recognize the exact metric/type/unit and candidate scope.
Its dimension-name list cannot prove nonempty required values on every stream or
establish a positive server-span witness. Describe it as recognized native
evidence, not complete semantic-convention compliance.

## Status limits

Generic RPC explicitly delegates error meanings to each framework. The pinned
gRPC section says all status codes other than `OK` SHOULD be considered errors,
but even there contextual instrumentation may choose differently. Other systems
can have warnings or partial success. The current Goo view should retain
**Recorded span Error**, Unset and Unknown without treating them as RPC success,
gRPC failure classification or business outcome. No new status inference is
required to add the current population profile.

## Retrieval fingerprints

The bounded local scratch copies are not product fixtures. SHA-256 of the two
principal reviewed documents and metric model:

- `rpc-spans.md`: `19ebb3400b8042ffe7f129e0de2399c3c024c58e536750fb59fd595cf0ec3785`
- `rpc-metrics.md`: `9db290d8badf40767d7880462f1931314471fd9a6ca82cc84fb2b522097ce8b2`
- `model/rpc/metrics.yaml`: `c227ea3b0047c9c30423c1d124419b9fa4dae6e10f39f7bca269ad7d0113a1ff`

See the [357 architecture contract](reviews/CURRENT_RPC_ARCHITECTURE_357.md) for
Goo's observation policy, auto-selection and independent evidence boundaries.
The [implemented profile](CURRENT_RPC_PROFILE_357.md) and
[validation](../VALIDATION.md) record subsequent runtime acceptance.
