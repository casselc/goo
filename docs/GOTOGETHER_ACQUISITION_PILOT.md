# Gotogether acquisition shadow pilot

The [proposed trial](GOTOGETHER_CURRENT_REVIEW.md) is implemented over the existing
[acquisition and retained-return journey](journeys/gotogether-acquisition.md).
It runs the existing independent source/native gate first, then a separate
pinned gotogether consumer. Product code and the normal verification module's
quintgo/Hegel dependencies are unchanged.

## What is adopted

- Quint is pinned at **0.33.1** in `package.json` and its lockfile. Existing
  lifecycle model-driven and observed-trace checks pass with the upgraded CLI.
- `verification/gotogether` is a separate Go 1.27.3 module pinned to upstream
  `771dcce13fd24071d3ae2677c567540ad0ec79ff`, with no local upstream replacement.
- The Playwright adapter is vendored unchanged with its MIT license, exact commit
  and byte hash. A small application adapter records actual DOM observations
  separately from the request's own completion.
- The model requires `Finishes()` and a named scenario. Visibility, retained
  receipt/window/question projection, local state and navigation epochs are
  checked. Readable opaque question/detail values use built-in binding `intern`;
  no custom string encoding remains.
- Source and compiled model judgments agree. The compiled positive runs with
  Node and Quint absent from `PATH`. Unsupported evaluation and stale artifacts
  refuse; they do not silently fall back or pass.
- The existing Bombadil actor supplies a separate request-adapter admission
  check. It does not claim to execute the full thirteen-phase scenario.

## Run it

With Go 1.27.3, `npm ci`, and the existing telemetry lab available:

```sh
npm run test:gotogether-controls
GOO_GENERIC_BINARY=/absolute/path/to/goo npm run test:gotogether-shadow
```

The API-contract CI workflow runs the source/model/protocol controls. The live
shadow is opt-in and requires actual OTLP/ClickHouse; its selected Go tests must
pass without skips through the existing strict event-stream checker.

## Result and scope

The final live run uses 240 real OTLP gauge points. Thirteen native phases retain
24 source SQL calls; the actor's nine SQL calls are recorded separately. All
three gates pass without skips, and the runtime and owned database are cleaned.
On one replay of the same 23 protocol records, source evaluation took 3.621s and
compiled evaluation 30.120ms. These checks cover the same shadow projection, not
the native source oracle's larger contract. They do not establish a benchmark.

Pending calls, missing terminal display, stale bank, hidden display and torn-tail
controls check their intended reasons in both evaluators; a stale compiled
artifact refuses. The model does not establish global collection completeness
or arbitrary transport-to-DOM causal ownership. Native source/question/row
assertions remain authoritative. No broad migration or promotion of gotogether
as the only gate has occurred.

The trial adds a shared protocol/completion path and promising compiled replay,
but still needs application-specific DOM projection and terminal drain logic.
Next evaluate that maintenance tradeoff before replacing existing bookkeeping.
A stateful CEL/AST/ClickHouse differential pilot remains a separate proposed step.

[Independent acceptance](reviews/GOTOGETHER_SHADOW_ACCEPTANCE.md),
[upstream feedback](reviews/GOTOGETHER_SHADOW_FEEDBACK.md), and
[executed checks](../VALIDATION.md) record the details.

## Upstream response on #220

The maintainer's [first response](https://github.com/chucklehead-dev/gotogether/issues/220#issuecomment-6884939876)
landed as [PR #221](https://github.com/chucklehead-dev/gotogether/pull/221), merged
at `a1a90fe5653db07cbf6cc467021bda18e4b27d79`. Graph witness replay now applies
backend selection; the new argv regression and end-to-end TypeScript Refutes
test cover our report. All 37 reported PR checks were successful when reviewed.
This records source and upstream CI review. Goo's consumer pin remains
`771dcce`; restoring its optional Refutes claim follows a pin update.

The external-driver reference now explains intern scopes and empty values,
Bombadil's unchanged-return epoch recipe, and the drain/settle bridge. A request-log
settle event (O1) and a compact completion-controls example (O2) remain planned.

The [second response](https://github.com/chucklehead-dev/gotogether/issues/220#issuecomment-6885341673)
proposes the broader contract workflow in [PR #223](https://github.com/chucklehead-dev/gotogether/pull/223),
which is open at this review. CW1 adopts our acquisition document as its first
literate-model consumer: tangle it, map a broken invariant to the document line,
refuse stale output and replay the existing log with the native source gate kept.
Go doc comments link to the authoritative model. Generated alignment, scenario
metadata, causal observations and production monitoring are subsequent roadmap
items; the native Go compiler/evaluator remains the separate Q track.

A useful consumer follow-up is one pinned portable acquisition document/model/log
fixture. Keep that preparation separate from implementing an extractor in Goo.
