# 438 raw gauge observations: architecture review

Status: backend source contract accepted after independent source inspection and
focused Go checks. UI response-admission corrections have been inspected. This review does
not assert browser or real-source execution.

## Decision

Keep a dedicated raw-observation request/result/profile beside native evaluation.
Do not construct GaugeResult, weaken evaluator preflight or invent instance IDs.
Discovery retains identity metadata, not values; consequently inspection requires
an explicit, independently bounded source read. Missing service.instance.id blocks
native evaluation where required but does not prevent seeing reported values.

Validate selected environment, namespace, service and fixed time directly,
without constructing an evaluator Definition. Independent instrument validation
permits exact reported empty scope and unit. Require explicit gauge kind, the
exact instrument tuple and existing narrow native
environment semantics; 437's log-only All selector does not widen gauge execution.

Gauge-only capability checking, typed SQL parameters and the existing atomic
row/byte transport protect source ownership. Never use a UI row limit as a source
total or publish a partial response after a budget failure. Preserve occurrence
multiplicity, attribute-key presence, units and source flags. No raw lookup should
silently filter absent instance labels or deduplicate repeated rows.

Use string timestamps and raw scalar text with explicit finite/nonfinite/absent/
invalid states. Missing and no-recorded-value remain unavailable. Identity/map
structure that cannot prove selected scope is malformed-source refusal; point
value diagnostics should remain inspectable without inventing a numeric zero.
Require canonical reported end timestamps in the half-open request window;
missing, malformed or out-of-window ends refuse the whole source. Start zero
is retained and invalid/missing starts remain diagnostic without interval claims.
Read-local occurrence keys are presentation references, not producer identities.

The UI owns acquisition; reusable table/detail components consume supplied data.
Freeze the original discovery question for the read. Scope replacement retires
held results and invalidates pending response ownership. Disclosure and local
presentation cannot fetch, resubscribe or create persistence. No timeline,
continuity, completeness, freshness or health claim follows from this profile.

## Follow-up review

Inspected `measurements/raw_gauge.go` and native route/policy registration. The
source validates before probing, binds the exact selected scope, requires canonical
in-window end timestamps, preserves duplicate rows and applies acquisition plus
serialized-response budgets atomically. No evaluator or persistence dependency
was introduced. `go test ./measurements -run RawGauge -count=1` passed.

UI follow-up corrected invalid reported start-string admission and removed
frontend-only attribute caps. Reinspection confirms raw-row/typed-field checks
retain point diagnostics; focused Node raw-profile tests pass. The flags
numeric-token parity edge is resolved through bounded exact integral JSON-number
parsing; equivalent forms such as `0.0` and `1e0` no longer disagree after JSON
decoding. JSON parsing can erase a fractional numeric token through rounding;
the browser therefore retains a conservative source-side invalid flags verdict
rather than admitting the rounded bits. Raw scalar lexical preservation is scoped
to the actual ClickHouse `toString(Value)` projection, not arbitrary numeric JSON
tokens. Unknown flag bits remain diagnostic, never sample-validity proof.
Parent-owned journey,
UX/accessibility review and model verification must be reported separately.
