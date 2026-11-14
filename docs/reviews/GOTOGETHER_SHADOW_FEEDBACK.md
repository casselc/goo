# Upstream feedback: external browser shadow integration

Posted as [gotogether #220](https://github.com/chucklehead-dev/gotogether/issues/220):
**Goo adoption feedback: Refutes backend selection and external-driver ergonomics**.
Earlier attempts were blocked by environment access; the user-requested retry
succeeded through `gh issue create`. The submitted issue body follows.

We trialed gotogether at `771dcce13fd24071d3ae2677c567540ad0ec79ff` as a
shadow consumer of Goo's acquisition/retained-return browser journey. The native
source/SQL and DOM checks remain independent and authoritative. Requests are
invocations, only their own transport outcomes complete them, and actual DOM
observations separately establish the held receipt. A deliberately obsolete 200
completes transport without being adopted. Code 0 completes a transport failure;
no response leaves the call pending.

The final raw-string/interned live trial passed with strict toolkit skips and
zero skips. Evidence: `artifacts/first-use439/6e20fa05c382`. Native admission ran
first and passed in 10.890s; `TestRecordedJourney` passed in 24.490s including
source/compiled negative controls and the compiled positive run with Node/Quint
unavailable. The same 23 protocol records took source cold wall 3.62135399s and
compiled cold wall 30.119915ms. Compilation is outside those two replay timings.
These are single cold executions, not a benchmark or equivalent full native
source-oracle costs. The actual Bombadil adapter check passed in 0.030s over
579 snapshots; its 9 SQL calls were accounted separately from the deterministic
journey's 24 SQL calls. Runtime and database cleanup passed. Original DOM strings
remain in the protocol log; built-in `intern` handles opaque equality identities.

## Reproducible backend selection inconsistency

`QUINT_BACKEND=typescript` applies to `Possible` sampling through
[`backendArgs`](https://github.com/chucklehead-dev/gotogether/blob/771dcce13fd24071d3ae2677c567540ad0ec79ff/validate/model/simulate_possible.go#L267).
Graph counterexample replay instead invokes `runQuint` with
`test --main simulate --match ...` and no backend argument in
[`simulate_graph.go:389`](https://github.com/chucklehead-dev/gotogether/blob/771dcce13fd24071d3ae2677c567540ad0ec79ff/validate/model/simulate_graph.go#L389).
`runQuintWithPulse` forwards argv unchanged. `ModelOptions` exposes constants,
states and depth, with no backend option.

A reproduction using the existing upstream graph fixture is:

```go
func TestRefutesTypeScript(t *testing.T) {
    t.Setenv("QUINT_BACKEND", "typescript")
    qtest.Model(t, "../validate/testdata/graph.yaml").Expect(t,
        qtest.Refutes("small").InAtMost(5))
}
```

Place that alongside `qtest/modelclaims_test.go`, run just this test with Quint
0.33.1 and without an installed/downloaded Rust evaluator. Witness replay runs
default-backend `quint test` and attempts Rust acquisition instead of honoring
TypeScript. Goo's implementation agent observed that download attempt; a
separate reviewer confirmed the call path by source inspection. We removed the
optional Refutes claim rather than reporting it as passed. Sampling plus explicit
TypeScript simulation and source/compiled trace controls remain separate checks.

Could graph witness replay consistently apply backend selection, with an argv
regression test? If Rust is deliberately required, please document that exception
and expose a clear option. This is distinct from the consumer modeling issues
below. The Rust-acquisition failure output was not retained as a standalone artifact; the call path above identifies the missing backend argument.

## Existing intern support resolved the string-boundary compromise

The first source replay correctly refused question JSON and selected-item text
containing double quotes. This is an intentional Quint literal limitation, not
a missing JSON escaping implementation: `binding.go:928-945` and
`itf.UnquotableError` explain that Quint strings have no escapes. `toQuint` and
collection helpers also reject such strings.

We initially tried reversible URI encoding, which preserved equality but added
custom projection code and made counterexamples harder to read. Reviewing the
existing API revealed the natural solution already available:

```yaml
intern:
  question: actor
  detail: actor
```

The model now uses integer identities for these equality-only fields while the
request log retains original DOM strings. That removes the custom encoding and
sidecar workaround. A small external-driver example demonstrating raw opaque
JSON/text plus `intern`, its scope, and the model's initial/empty identity
semantics would make this feature easier to discover. This is a documentation
request, not a claim that gotogether lacks the necessary capability. Our bounded
journey uses an admitted plain search string; arbitrary text should be interned
or otherwise explicitly admitted before expanding that input vocabulary.

## Explicit terminal drain bridge

The Playwright adapter's `close()` removes listeners and drains already-started
answer-body tasks. EOF correctly does not imply settlement. We drain the owned
source meter and browser body observers, close the adapter, verify every owned
request ID has an outcome, then emit a custom `drained` step and map it to
`qtrace.Step.Settled`. This is a small custom bridge, not a product change.

Could documentation or a helper standardize this bridge while requiring the
caller to establish quiescence? Browser journey settlement must remain distinct
from a telemetry collection watermark or proof that no producer records were
lost. Our checks retain explicit eventual end, real final visible DOM observation,
completed invocations and `Finishes()`; closing the page alone is insufficient.

## Bombadil unchanged-return observations

`converter.shown` deduplicates display values per key across the trace. For
show A → leave → return to unchanged A, a consumer requiring a named retained
return may receive only the initial A unless `Show.Fields` also includes a
phase/navigation epoch. Please document that recipe or consider a deliberate
observation-emission policy. This is mapping friction, not a claim that display
deduplication is wrong. Our Bombadil campaign validates adapter admission over
its exploratory actor; it does not claim the deterministic 439 journey finished.

## Consumer responsibilities and limits

Our model initially omitted visibility and question identity and only checked
nonnegative epochs; those were harness/model weaknesses, not upstream defects.
It now projects actual DOM question identity, checks visibility, unique completed
read IDs and monotonic navigation epochs. It deliberately does not infer
read-body-to-DOM adoption causality. The native oracle retains full source rows,
query scope, SQL accounting, cleanup and bounded DOM checks.

We compare source and compiled judgments on identical controls and require the
intended diagnosis: pending invocation, absent eventual end, retained-ownership
violation, hidden final display, or lost tail. A code-0 fixture completes; stale
compiled artifacts refuse. A compact upstream external-driver example covering
these distinctions would help consumers avoid confusing conforming prefixes
with finished journeys. No promotion or gate replacement follows from one cold
timing comparison.
