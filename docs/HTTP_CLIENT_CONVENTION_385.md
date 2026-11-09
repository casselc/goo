# Pinned HTTP client convention — 385

Retrieved from the official OpenTelemetry semantic-conventions repository on
2026-11-09. The official latest-release redirect resolves to **v1.45.0**;
`git ls-remote --tags` resolves that tag to immutable commit
`fb30ca17193b551473b1f43c315482918ab6987c`. The observations below use that commit,
not a moving website. Reads used the managed proxy with verified TLS.

## Official sources

- [HTTP spans](https://github.com/open-telemetry/semantic-conventions/blob/fb30ca17193b551473b1f43c315482918ab6987c/docs/http/http-spans.md): client kind, requirement levels, naming, status and retries.
- [HTTP attribute registry](https://github.com/open-telemetry/semantic-conventions/blob/fb30ca17193b551473b1f43c315482918ab6987c/docs/registry/attributes/http.md).
- [Server attribute registry](https://github.com/open-telemetry/semantic-conventions/blob/fb30ca17193b551473b1f43c315482918ab6987c/docs/registry/attributes/server.md): reported destination address/port, including intermediaries.
- [Recording errors](https://github.com/open-telemetry/semantic-conventions/blob/fb30ca17193b551473b1f43c315482918ab6987c/docs/general/recording-errors.md): contextual classification, status and error.type.
- [HTTP metric model](https://github.com/open-telemetry/semantic-conventions/blob/fb30ca17193b551473b1f43c315482918ab6987c/model/http/metrics.yaml) and [metrics](https://github.com/open-telemetry/semantic-conventions/blob/fb30ca17193b551473b1f43c315482918ab6987c/docs/http/http-metrics.md): separate native instrument semantics.

## Span population and dimensions

The client span convention is Stable and requires `SpanKind.CLIENT` (the record
adapter reports `Client`). Instrumentation SHOULD emit one span per send attempt,
but MAY emit a top-level client-operation span when per-attempt hooks are not
available. Goo therefore counts **matching Client-span records**, not attempts,
unique requests or physical network calls. Reported retry spans remain records;
unreported retries cannot be inferred.

| Attribute | Pinned convention | Goo observation policy |
| --- | --- | --- |
| `http.request.method` | Required string; method names are case-sensitive. Unknown methods use literal `_OTHER`; original method is separately reported when needed. | Require a reported nonempty current key. Preserve literal values; no old `http.method` union, uppercasing, trimming or synthesized `_OTHER`. |
| `server.address` | Required string; hostname without reverse DNS lookup, otherwise address/socket. For a proxy, the intended server address should be used when available. | Group qualified reported values. Missing and empty remain distinct. A hostname is not an instrumented callee identity or verified topology. |
| `server.port` | Required int; port from absolute request-target authority or Host/:authority as applicable. | The adapter reports attribute strings. Show literal reported context, retaining missing/empty/noncanonical values; do not infer a default port or silently normalize. |
| `http.response.status_code` | Conditionally required iff a response status was received/sent. | Report separately from recorded span status. Missing status is not success or failure. |
| `error.type` | Conditionally required on failure; bounded predictable identifiers recommended. | Preserve evidence; it does not authorize a new universal error policy. |

The profile is a bounded observation lens, not instrumentation conformance
certification. Requiring method/kind establishes its selected population; missing
other required convention attributes remain visible in that population instead
of being discarded. Flat metadata keys admit an explicit candidate, not proof
that Client records exist. Actual reads retain the exact kind/method predicate.

## Status is contextual

Without further context, the HTTP convention says 1xx/2xx/3xx status should leave
span status unset unless another error occurred (for example while reading the
body). Server 4xx MUST leave status unset; Client 4xx SHOULD be Error. 5xx SHOULD
be Error. Additional application context may classify an expected 404 differently.
Intentional caller cancellation should not be treated as an error when detected.

Goo displays recorded Error/Unset/other span-status evidence independently from
HTTP response values. It does not reconstruct status from a numeric response,
infer remote server health, or claim an Error record proves a failing dependency.

Span naming should use method and a known low-cardinality target when available;
raw URI paths must not be the default target. This cut groups declared destination
and method attributes, not an unbounded `url.full` dimension.

## Native execution and validation boundary

`http.client.request.duration` is a separate histogram instrument with unit seconds
in the pinned model. Generic discovery does not establish an HTTP-client execution
profile. This milestone adds record observations only; native profile adoption is
separate depth with its own source, temporality, unit and applicability checks.

The retained lab's Client spans were attrless manual RPC observations, so they
cannot be relabeled as HTTP. Acceptance uses a small Go workload making actual
HTTP requests through a caller/client/dependency trace and exporting through the
real collector. Deliberate missing/empty attribute variants are explicitly partial
instrumentation test cases, not claims of convention-conforming instrumentation.
