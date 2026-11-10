# HTTP-client native duration convention and design review — 392

Reviewed 2026-11-09 before profile execution changes. This is an implementation
contract/proposal, not runtime acceptance. The current delivered record lens is
[Outgoing HTTP](../OUTGOING_HTTP.md); it remains independently acquired evidence.

## Pinned official sources

A fresh official latest-release redirect resolved to **v1.45.0** and
`git ls-remote --tags` resolved that tag to immutable commit
**fb30ca17193b551473b1f43c315482918ab6987c**, matching the 385 review.
Reads used the inherited managed proxy and verified TLS. Runtime networking
skill and startup policy were read; current environment status reported an
enforced restricted policy with GitHub/raw GitHub allowed and VPN unconfigured.
The default sandbox could not reach the proxy; the same requests succeeded in
an approved outer-executor command. No new domains or credentials were needed.

- [HTTP metrics, including client duration](https://github.com/open-telemetry/semantic-conventions/blob/fb30ca17193b551473b1f43c315482918ab6987c/docs/http/http-metrics.md#metric-httpclientrequestduration)
- [Machine-readable HTTP metric model](https://github.com/open-telemetry/semantic-conventions/blob/fb30ca17193b551473b1f43c315482918ab6987c/model/http/metrics.yaml)
- [HTTP spans, including retry/status semantics](https://github.com/open-telemetry/semantic-conventions/blob/fb30ca17193b551473b1f43c315482918ab6987c/docs/http/http-spans.md), previously reviewed in [385](../HTTP_CLIENT_CONVENTION_385.md)

The table below transcribes the generated normative client-duration section,
not experimental body-size/open-connection instruments or HTTP-server attributes.

## Convention and finite source matrix

| Concern | Pinned convention | 392 admission/meaning |
| --- | --- | --- |
| Instrument | Stable `http.client.request.duration`; metric is recommended | Exact current name only. No legacy alias or incoming HTTP substitution. |
| Type/unit | Histogram, UCUM `s` | Explicit Histogram exporter representation, exact `s`. Generic `ms` support does not grant this profile; no conversion. Exponential histogram stays unsupported by the current source adapter/kernel. |
| Boundaries | SHOULD use advisory boundaries `[.005,.01,.025,.05,.075,.1,.25,.5,.75,1,2.5,5,7.5,10]` | Advisory, not a mandatory layout. Retain actual valid bounds/counts and existing layout-change safeguards; never rebin to the recommendation. |
| Temporality | This metric section does not mandate delta or cumulative | Existing delta `1` and cumulative `2` kernels; unknown/unsupported temporality cannot gain profile execution. Preserve start/end/reset/gap and full-interval coverage rules. |
| Schema | Stable current instrument semantics do not certify reported schema conformance | Keep scope/resource schema URLs literally in full identity and source selection, including omitted/other values. Explain matched/omitted/other metadata; no schema migration or conformance claim. A schema label alone cannot rename a legacy instrument. |
| `http.request.method` | Required string; case-sensitive known values, `_OTHER` for unknown to instrumentation | Preserve reported values/missing/empty. Do not uppercase, synthesize `_OTHER`, union `http.method`, or require span fields. |
| `server.address` | Required string; intended reported server/authority context | Literal context, not callee service identity, resolved topology or reverse-DNS enrichment. |
| `server.port` | Required integer | Literal reported context; no default-port inference or silent normalization. |
| `error.type` | Conditionally required when request ended with error | Missing does not establish success; no new universal failure policy. |
| `http.response.status_code` | Conditionally required iff received/sent | Distinct from instrument count, recorded span Error and application outcome; absent does not mean failed/successful. |
| `network.protocol.name` | Conditional when not `http` and protocol version is set | Preserve literal evidence, no protocol inference. |
| `network.protocol.version` | Recommended; negotiated/actual version when known | Optional for profile execution. |
| `url.scheme` | Opt-in | Optional literal context. |
| `url.template` | Opt-in, Development; MUST have low cardinality | Optional literal context; do not derive templates from raw URLs or make raw paths a default grouping. |

Required-versus-optional above describes instrumentation guidance. Discovery is
an observation lens, not a conformance validator: missing required dimensions
must remain visible partial metadata instead of silently dropping histogram
streams. Profile identity/shape and generic source validity determine eligibility;
reported dimensions do not establish numerical validity or completeness. Full
resource, instance, namespace presence, scope/version/schema and attributes stay
in stream identity; service-level metadata preview is not a complete inventory.

## Aggregation and prohibited inference

Reuse the existing per-stream explicit histogram evaluator and supplied-data
Histogram/Table displays. Count represented **instrument observations**, separately
from exported rows and matching Client-span records. Do not average scalar
percentiles into fleet percentiles, interpolate a precise latency from bucket
ranges, repair gaps/resets, merge differing identities/layouts, or call unknown
coverage zero. The advisory boundaries are not evidence of actual bounds.

The convention says the metric value SHOULD equal HTTP client span duration when
reported alongside a client span. This is not a join, existence guarantee,
completeness guarantee or proof that one metric observation is one retained span.
385 permits different retry-span representations; neither signal proves unique
requests, physical attempts, retries, an instrumented remote service or remote
health. Metric status/error dimensions and span Error remain independent evidence.

## Minimal coherent UX

Use the existing Services native instrument card, histogram defaults and ordinary
Measurements continuation. Title it **HTTP client duration** and identify seconds,
observations, exact fixed window, state and coverage with existing compact chrome.
Native metadata may offer the default native action without enabling the
Outgoing HTTP record lens. Metrics-only must work; spans-only keeps the current
record lens; combined data offers both independent choices and read clocks.

Keep actual plots/first data before collapsed full identity/provenance. Expose
unknown/stale/invalid/no-data states and source errors without a fallback to
incoming HTTP or an unrelated record population. Missing dimensions are inspectable
context, not a prominent new wizard. Reuse current explicit Run, contextual Back,
Save/capture and destination reuse contracts; local kind/disclosure changes do not
read telemetry. Any exemplar continuation must retain its existing separate
reported-reference authority or stay outside this cut.

Acceptance must separately prove supported/refused metadata and metrics-only,
spans-only and combined discovery; delta/cumulative/reset/gap/unknown behavior;
raw bucket/count/range oracles; exact source/destination identity; and no-read
held/local operations. This document itself proves none of those runtime claims.
