# 436.1 independent Go and architecture review

Accepted for the reviewed domain and passive selection consolidation. No open
correctness findings in the reviewed production diff. This report does not by
itself establish the separate live-browser held-return acceptance.

## Contract review

- `exploration` owns ordinary predicate admission, detached typed cloning and
  complete-parent refinement. CEL compilation uses predicate admission directly;
  `Program.Apply` and the existing non-CEL `StageGroupSide` delegate refinement.
  There is no synthetic validation question or JSON cloning in these paths.
- Admission precedes recursive cloning and bounds cyclic input by the existing
  depth/node limits. Nil is an unrestricted predicate; explicit empty branches
  remain invalid. Default unsigned decimal admission still accepts `01`.
- Refinement preserves the whole parent OR under AND and retains environment,
  fixed window, projection, grouping and paging. `StageGroupSide` retains its own
  offset-reset policy. Combined admission precedes returning an owned question.
- Typed copies detach predicate fields/numeric bindings, grouping slices,
  environment selector, numeric projection, histogram specification and all
  reported-edge namespace pointers. Current nested scalar/value fields need no
  further allocation. No wire canonicalization or retained evidence rewrite is
  introduced.
- `holdSelection` shares only synchronous pre-hold/post-hold admission. Activity
  volume and latency retain their acquisition, read gate, render/scope ownership
  and invocation checks. Native control eligibility and stale/reentrant callbacks
  remain adapter concerns. Latency retains comparison of an already staged dirty
  filter while refusing a change of cleanliness during the invocation.

## Independently executed checks

Managed Go 1.27.3, with the workspace toolchain/caches and CGO disabled:

```sh
go test ./exploration ./querycel -run 'Test(Refine|PredicateAdmission|CloneQuery|StageGroupSide|Apply|Compile)' -count=1
```

Both packages pass (0.025s and 0.009s). This includes nested clone isolation,
whole-parent preservation, combined budgets, cyclic input, nil identity and
explicit empty-branch refusal. A separate complete `querycel` run passes (0.042s).
The independent broad exploration attempt hits sandbox loopback `EPERM` in the
existing `TestAcceptedLargeNativeAndComparisonCaptures`; that attempt is not a
whole-package pass.

```sh
node --experimental-test-module-mocks --test tests/activity-latency-selection.test.mjs tests/activity-log-context.test.mjs tests/activity-latency.test.js tests/activity-log-model.test.mjs
```

All four file runners pass, zero skips (0.98s). Mounted owner tests exercise
post-hold dirty/hidden/disabled/disposed/scope/render-ABA withdrawal and retained
local comparison. Their passive histogram/baseline substitutions are explicit;
these checks do not establish browser layout, real telemetry, physical assistive
technology or cross-engine behavior. An initial Node invocation omitted the
required module-mocks flag and failed setup; the corrected command above passes.

The final bounded browser-driven correction was independently reviewed and the
same four UI file runners rerun successfully (0.516s, zero skips). Latency Compare
now exposes native disabled state when its selected render has retired; calling
its retained handler after inactive/active Return cannot pause, compare or read.
Its scope check binds the acquisition gate's original selector rather than
adopting the provider's invocation-time selector. A foreign environment supplied
before sync refuses without pausing or reading, and sync disables the control.
Current same-render dirty comparison remains valid. No further findings.
