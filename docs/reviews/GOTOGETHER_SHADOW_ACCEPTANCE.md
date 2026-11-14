# Acquisition 439 shadow review acceptance

Review basis: upstream gotogether `771dcce13fd24071d3ae2677c567540ad0ec79ff`,
[adoption proposal](../GOTOGETHER_CURRENT_REVIEW.md), and the existing
`verification/raw_acquisition439_observed_test.go` source and lifecycle oracle.
This document records the independent review boundary; an acceptance requirement
is not evidence that a check has executed.

## Required semantic boundary

The request protocol proves which invocation received which transport outcome.
It does not prove that the UI adopted the response. Actual DOM observations must
carry the displayed bank, question, source-read receipt and navigation epoch.
The deliberately obsolete 200 response completes its own invocation while the
before/after displayed bank remains unchanged. An aborted network request has a
recorded code-0 outcome and is complete; a request with no response record is
still pending. A neighboring response or display observation cannot complete it.

A terminal marker must follow the actual final observed display and transport
completion. It establishes that this bounded browser journey ended; it cannot
establish an OTLP collection watermark, ClickHouse ingestion drain, or absence of
unrecorded producer telemetry. `Finishes()` requires a nonempty explicit eventual
journey-end property, genuine terminal settlement, conformance and completed
invocations. Merely appending a marker after the last click is insufficient.

The existing native acquisition admission and independent source checks remain
authoritative. The shadow model may consume observed projections but must not
replace the native oracle with its own success flag. Cleanup, cardinality,
SQL accounting, scope, raw response rows and bounded DOM remain separately
checked by their existing owners.

## Required controls and comparison

| Control | Required result |
| --- | --- |
| Full actual bounded journey | Conforms and finishes, with named read/local/return scenario reached |
| Truncated or torn log | Rejected or undecided; never finishes |
| Removed response | Pending invocation; never finishes |
| Removed final display | Explicit eventual end remains false or unjudged; never finishes |
| Obsolete response adopted into the bank | Model violation |
| Aborted request with recorded code 0 | Completed transport outcome, preserving held bank |
| Wrong request identity / duplicate response | Protocol rejection |
| Log gap, unsupported extraction or lost records | Explicit refusal or loss; never silently repaired |
| Deliberately stale compiled artifact | Refusal |
| Unsupported compiled evaluator operation | No verdict; never a pass |

Controls must be deterministic transformations of a known accepted trace, with
failure asserted explicitly. They must not pass only because an unrelated
parsing error or empty trace prevented evaluation. Toolkit strict mode does not
cover arbitrary `testing.T.Skip`: required evidence must report zero skips.

Compare native and compiled judgments over the same admitted recorded journey
and controls. Report compilation/setup separately from warm replay, use the same
iteration count and inputs, and retain the native source oracle in both cases.
An upstream test duration is not a Goo performance comparison. Artifact creation
must cross-check the model with the independent Quint backend before measuring
the supported Go evaluator. No promotion follows solely from a faster partial
check that omits native admission or journey-completion obligations.

## Protocol and adapter review notes

Correlation IDs in request `corr` become diagnostic refs; they are not checked
semantic joins. Request IDs and request-specific responses supply transport
correlation. Driver timestamps preserve local order, not cross-host causality.
The adapter must not recreate the trace using a new recorder and thereby replace
original time, sequence, call identity or artifact refs.

Upstream Bombadil `Show` deduplicates prior values by key. An unchanged retained
bank can therefore suppress a later return observation unless the mapping also
carries the phase or navigation epoch. A final return observation must be real
and distinguishable even when the bank itself is unchanged. The mapping must
refuse requests that fell out between snapshots and unmapped custom actions.

## Current implementation review

The model has a nonempty eventual `journeyFinished`, unique completed read IDs,
and a monotonic navigation epoch fence. A late read from an older epoch can
complete without changing the held DOM receipt. This deliberately does not infer
DOM adoption from transport bodies. Bank identity is source-read UTC, clock is
the actual fixed-window text, and question is JSON projected from actual staged
service, environment and instrument DOM fields. The native oracle independently
checks the complete question, returned rows and source ledger. The model checks
receipt/content retention; read-ID-to-DOM causality is not established.

Explicit phase and final DOM observations avoid unchanged-bank deduplication.
The driver drains source meter and browser body observers, closes the adapter,
checks unmatched owned request IDs and writes `drained`; the Go bridge maps only
that explicit terminal marker to settlement. EOF alone is insufficient. A lost
tail after a marker remains loss evidence and cannot finish.

The source and compiled negative controls now compare verdict/reason and require
the expected diagnosis: pending calls, eventual end absent, retainedOwnership
violation, or lost tail. A separate protocol fixture records code 0 as a completed
transport outcome. Stale artifact refusal is explicit. Agent-reported checks:
`TestModelPossible` passed (1.885s), and `TestProtocolControls` passed (19.91s).
These results were reported by the implementation owner; this reviewer inspected
the code and did not rerun them. Timing output is explicitly cold wall duration,
not a benchmark. Final live browser evidence is recorded below.

Visibility was subsequently added to the model's show signature; safe ownership
now requires visibility after acquisition/adoption. A hidden-final negative
control was added for both source and compiled judgments. These changes address
the initial hidden-DOM gap; the implementation owner reports strict fixture rerun passed (23.404s, zero skips).

The live trial exposed quoted DOM text that cannot be represented as a Quint
string literal. Upstream intentionally refuses it. Built-in actor-scoped
`intern` now maps raw `question` and `detail` strings to integer identities for
the equality-only model. The raw request log preserves the original strings;
custom URI encoding and its sidecar were removed. A fixture with JSON quotes and
multiline detail passed both evaluators and controls, including the positive
compiled run with PATH empty (owner-reported source cold 2.981s, compiled cold
17.96ms). This is finite schema calibration, not browser evidence or a benchmark.
The final raw/interned live trial passed: `artifacts/first-use439/6e20fa05c382`.
Native admission passed first (10.890s), then `TestRecordedJourney` passed
(24.490s) with strict toolkit skips, zero skips, source/compiled control parity,
and positive compiled execution without Node/Quint. The same 23 protocol records
used source cold wall 3.62135399s and compiled cold wall 30.119915ms; compilation
is outside both replay durations. Those are single cold executions, not a warm
benchmark or equivalent native-source-oracle cost comparison.

Bombadil adapter admission passed (0.030s, 579 actual snapshots). Its 9 SQL calls
were separately attributed from the deterministic acquisition journey's 24 SQL
calls. Runtime and database cleanup passed. These executed results were supplied
by the implementation owner/coordinator; this reviewer independently inspected
the model, binding, driver and control code rather than rerunning the live suite.

The bounded shadow trial is accepted with the documented projection limits. It
adds completion/transport diagnostics and preserves the independent native
source oracle. It does not establish general transport-to-DOM causality,
telemetry collection completeness, arbitrary-input support, a performance
benchmark, or approval to replace an existing gate.
