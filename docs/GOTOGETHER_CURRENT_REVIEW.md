# Gotogether: current changes and proposed Goo adoption

Reviewed 2026-11-14 against commit [`771dcce`](https://github.com/chucklehead-dev/gotogether/tree/771dcce13fd24071d3ae2677c567540ad0ec79ff).
This follows the isolated [original pilot](GOTOGETHER_PILOT.md) and
[65093ec follow-up](GOTOGETHER_FOLLOWUP.md). It is a source review and focused
upstream test run, **not a migration of Goo's verification module**. Goo remains
on its existing quintgo and Hegel dependencies.

The subsequent [acquisition shadow pilot](GOTOGETHER_ACQUISITION_PILOT.md) is now
implemented and Quint is upgraded to 0.33.1. Tool versions and unchanged-file
statements below describe this earlier review, before the trial.

## Changes worth using

| Priority | Available now | Concrete fit for Goo |
| --- | --- | --- |
| First | `qdriver/bombadil` and the Playwright request-log adapter | Replace repeated request/completion bookkeeping with one adapter. A sent request is an invocation; only its own response completes it. A DOM `show` observation is separate. |
| First | `qtest.Finishes()`, scenarios, strict toolkit skips | Require the actual journey end, settlement and completed calls instead of accepting a conforming prefix. Name required steps such as read → local inspect → retained return. |
| First pilot | Compiled models and the Go evaluator | Compile a binding/model once, then judge supported runs without Node/Quint at test time. Measure against the current acquisition gate before migrating more checks. |
| Next | `connect.Differential`, `Vary`, required states and `Step.Eval` | Run identical stateful filter/refinement sequences through CEL/reference, ordinary AST and ClickHouse adapters. Vary inputs beyond model guards to test boundary admission. |
| Selectively | `qshrink`, observation mutants, refs and portable replay | Shrink a reproducible failing external journey and verify that dropping/corrupting meaningful recorded fields is caught. Keep compact failure artifacts rather than every successful run. |
| Later | OTLP conversion rules and collection watermarks | Correlate a sample application's domain effects with Goo's observed UI population without requiring every producer to import gotogether. Only assert absence after demonstrated collection progress. |

The previously announced E1–E8 work is now implemented. It is no longer just a
future Bombadil example. The newest merged work adds differential replay and
input variation; required-state coverage and model-definition evaluation also
landed. Those are particularly relevant to Goo's compiler/reference/source
agreement and to workflows that random exploration may never reach.

## Important boundaries

- **HTTP success is not UI adoption.** The adapters complete transport calls;
  Goo must separately record the actual displayed bank, question identity,
  source-read receipt and navigation epoch. A late successful read must not
  manufacture a successful `show` observation.
- `Finishes()` needs a binding with an eventual end condition, a genuine settle
  and completed invocations. Closing a browser or reaching a last click does not
  establish collection drain or a finished investigation.
- Compiled-model artifacts bind the model and binding inputs. Stale artifacts
  refuse. Unsupported native evaluation remains no verdict; running without
  Quint must not turn that into a pass. Keep independent model/evaluator checks
  in the build that produces the artifacts.
- Differential comparison needs a common admitted contract. Visual numeric
  bindings and unsupported CEL expressions intentionally differ; compare their
  shared vocabulary, not every error string or the entirety of both grammars.
  `IgnoreModel` compares implementations with each other and is not model
  conformance. `Step.Eval` is useful for state projections, but independent
  source/semantic oracles should remain independent.
- External shrinking requires fresh environments per candidate. Current
  Bombadil reproduction can diverge on timing; the webjobs shrink example uses
  a Go client to rerun extracted actions, not faithful browser replay. A smaller
  different failure is not evidence of the original bug.
- `CompleteThrough` is a collection watermark, not a source-read timestamp.
  The collector-side watermark writer is still a documented follow-on.
  Converted spans cannot reveal spans that were lost. Correlate cross-app work
  with request/domain IDs and causal order, not host wall-clock order alone.
- `GOTOGETHER_STRICT=1` covers toolkit skips; it does not make every arbitrary
  `testing.T.Skip` or external driver opt-out fail automatically.
- The new Valkey double is useful for supported KV/Lua operations, but has no
  Redis Streams command coverage for the sample application's bus. Its
  simulation profile and `qenv.Store` cannot replace real ClickHouse semantics.
  Grist is a separate application framework, not a reason to rearchitect Goo.

## Recommended first trial

Use [439's acquisition/retained-return journey](journeys/first-use-acquisition.md)
as one shadow consumer, in a separate pinned scratch module. Keep the existing
native gate and independent source assertions authoritative during the trial.

1. Emit the request protocol using the Playwright adapter, with separate actual
   DOM observations. Map the existing Bombadil extraction to `qdriver/bombadil`.
2. Model read, local inspection, departure/return, cancellation and obsolete
   delivery. Require the original bank/clock to survive return and make the
   completed journey an explicit eventual state.
3. Check `Finishes()` and a named scenario. Require truncated logs, unanswered
   calls, missing final display and stale-response adoption to fail or remain
   undecided as appropriate, never pass.
4. Compile the model and compare judgments and elapsed time with the current
   native gate. Refuse a deliberately stale artifact. Retain an independent
   source oracle and cross-check the model toolchain during artifact creation.
5. Promote only if this removes harness code or materially improves diagnosis
   or execution time. Then consider one stateful differential CEL/AST/ClickHouse
   pilot. Do not migrate all journeys or upgrade Hegel as a side effect.

Use one immutable commit for core/CLI and any optional submodules. The reviewed
commit is a **candidate**, not a claim that its full CI is green. No normal Goo
module, installed skills, package lock or runtime tools changed in this review.

## Local checks and tool requirements

- Reviewed full pin: `771dcce13fd24071d3ae2677c567540ad0ec79ff` (latest merged HEAD when fetched).
- New floor: Go **1.27.3** and stable Quint **0.33.1+**. Goo's Go meets it;
  the checked-in Quint **0.32.0** does not. A consumer module would also need
  its Go directive at least 1.27.3 (Goo currently declares 1.27.0). Optional
  `qhegel` now pins Hegel **0.9.30**, versus Goo's **0.9.8**; keep that separate
  from a core/driver pilot. Quint 0.33.1 was installed only under
  `/tmp/goo-gotogether-tools` for this check.
- Focused upstream tests pass with strict toolkit skips and the TypeScript
  backend: request/Playwright protocol **1.691s**, Bombadil adapter **0.006s**,
  differential replay/variation **5.579s**, compiled-model unsupported/stale
  controls **14.619s**. There were no skips. These are upstream tests, not proof
  that Goo has integrated the new APIs or a performance benchmark of Goo.
- The initial differential run correctly refused Quint 0.32.0; the final run
  used the supported isolated version. The Rust evaluator and a live upstream
  Bombadil campaign were not exercised.
- GitHub's check-runs API returned `Forbidden` in this environment. Full CI
  status for the candidate was not independently confirmed.

Source entry points: [request protocol](https://github.com/chucklehead-dev/gotogether/blob/771dcce13fd24071d3ae2677c567540ad0ec79ff/qdriver/qdriver.go),
[Bombadil adapter](https://github.com/chucklehead-dev/gotogether/blob/771dcce13fd24071d3ae2677c567540ad0ec79ff/qdriver/bombadil/bombadil.go),
[webjobs integration and replay limits](https://github.com/chucklehead-dev/gotogether/blob/771dcce13fd24071d3ae2677c567540ad0ec79ff/examples/webjobs/README.md),
[completion checks](https://github.com/chucklehead-dev/gotogether/blob/771dcce13fd24071d3ae2677c567540ad0ec79ff/qtest/expect.go),
[compiled models](https://github.com/chucklehead-dev/gotogether/blob/771dcce13fd24071d3ae2677c567540ad0ec79ff/validate/compiled.go),
[differential testing](https://github.com/chucklehead-dev/gotogether/blob/771dcce13fd24071d3ae2677c567540ad0ec79ff/connect/differential.go),
[required model states](https://github.com/chucklehead-dev/gotogether/blob/771dcce13fd24071d3ae2677c567540ad0ec79ff/connect/require.go),
[model-definition evaluation](https://github.com/chucklehead-dev/gotogether/blob/771dcce13fd24071d3ae2677c567540ad0ec79ff/connect/eval.go),
[external shrinking](https://github.com/chucklehead-dev/gotogether/blob/771dcce13fd24071d3ae2677c567540ad0ec79ff/qshrink/shrink.go),
[OTLP fit](https://github.com/chucklehead-dev/gotogether/blob/771dcce13fd24071d3ae2677c567540ad0ec79ff/docs/fit.md), and
[remaining external-driver gaps](https://github.com/chucklehead-dev/gotogether/blob/771dcce13fd24071d3ae2677c567540ad0ec79ff/docs/ROADMAP.md#left-from-e).
