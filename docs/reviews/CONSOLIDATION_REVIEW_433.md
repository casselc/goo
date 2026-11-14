# Discovery and recovery consolidation checkpoint

> Follow-up: 433b is delivered. See [source-recovery acceptance](TAIL_SOURCE_RECOVERY_433.md).
> The review below records the priorities adopted before that implementation.

Independent architecture, product and UX reviews agree on the bounded 433a
consolidation: all-observation service-instance preflight, shared supplied-plan
readiness in Services and Measurements, compact disclosed metadata, and semantic
Activity draft identity. Executed acceptance belongs in [validation](../../VALIDATION.md).

## Reuse without erasing contracts

Keep the discovery controllers separate. Environment discovery can return a
limited 64-group list followed by preview; metric scopes promise complete bounded
canonical groups and refuse overflow. Source capabilities describe schema/access
and bounded presence. Activity owns SSE deduplication, reconnect, continuity and
held observations. A generic controller would obscure these differences.

Share native-control guards, pure identity projections, fixed source profiles and
supplied instrument readiness. Automatic Services overview scheduling applies that
same rule before compacting candidates, so a disabled Preview cannot be bypassed
by an automatic widget read. Original observations and refusal reasons stay in
Native instrument cards. Filtering is browser scheduling policy, not API
authorization or a source absence claim. Numeric validity and interval/reset semantics stay
in the evaluator. Generic resource/scope/instrument/attribute stream evaluation
needs a separate decision for gauges, counters, histograms and epochs; missing
instance identity cannot become an invented producer or fleet aggregate.

## Remaining 433b: stable Activity source refusals

Both product and architecture reviews identify the same remaining P1. The live
handler currently collapses poll errors to `source_unavailable`; the browser
reduces HTTP errors to status numbers and reconnects every two seconds regardless
of cause. Schema and access problems can therefore look transient.

Reuse safe structured source refusals, preserve held rows and original scope, and
distinguish transient reconnect from stable configuration/access/schema refusal.
Terminal failures should offer explicit Retry. Failure must never become an empty
read or trigger automatic widening. Test missing/denied sources, transient loss,
held inspection, retired callbacks and keyboard/mobile Retry. This is the next
bounded milestone; 433 is not complete until it lands.

## Next operator cycle after 433b

| Cadence | Proposed bounded work | Operator value |
| --- | --- | --- |
| Breadth | Explicit local Find → source log-message search | A labeled action stages literal Body search inside the applied scope and a visible fixed window; Run reads beyond the buffer. Local text/ID/field matching keeps its own meaning. |
| Depth | Numeric OTel severity in the existing Activity volume strip | Add complete severity-band counts/time context and exact numeric filters. Keep unspecified/invalid values explicit and SeverityText independent. |
| Consolidate | Log-volume/span-latency selection and held return | Reuse demonstrated matching read-gate and passive chart-intent contracts; refresh narrow/desktop flows and measure before claiming speedups. |

These are proposed follow-ons, not delivered functionality. Ordinary charts,
waterfalls, service dashboards, comparisons, generic filters, annotations and view
templates already exist and are not new missing fundamentals. Monitoring delivery,
multiplayer, arbitrary executable kinds and warehouse lowering remain separate
initiatives. The older 427–429 preservation contracts stay deferred.

Measure repeated source metadata probes under realistic latency before optimizing.
Prefer bounded request-local reuse if justified; durable readiness caching can
retain stale schema or authorization decisions.
