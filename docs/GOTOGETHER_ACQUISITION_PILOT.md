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
