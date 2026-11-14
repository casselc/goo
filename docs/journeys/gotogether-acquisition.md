# Gotogether acquisition shadow trial

The opt-in shadow runs the existing [439 acquisition journey](first-use-acquisition.md)
with its thirteen phases, real OTLP metric ingestion, transparent source meter,
independent source-cost assertions and native Quint gate. The native gate runs
first. A failure there prevents the shadow from substituting its own verdict.
There are no product UI changes; the existing 439 screenshot frames illustrate
this same flow.

Run with the lab's ClickHouse and OTLP receiver available, a compiled generic
binary and the pinned Go/Quint tools on PATH:

```sh
GOO_GENERIC_BINARY=/absolute/path/to/goo GOO_GOTOGETHER_SHADOW=1 \
  npx playwright test journeys/first-use-acquisition.spec.js --workers=1
```

The dependency-free upstream Playwright adapter is vendored byte for byte at
[the reviewed pin](../../journeys/support/gotogether-driver-PROVENANCE.md).
It records each raw acquisition invocation before forwarding the request and
only completes that invocation on its own response or transport failure. The
late-success adversary holds a real server response at the context route, below
the page recorder, so it also passes through this request protocol.

Each phase separately records actual DOM receipt text, fixed window, local page,
Find value and selected detail. Returning or receiving HTTP 200 cannot itself
manufacture an adoption observation. The navigation epoch changes before the
pending-departure adversary. Final display is observed again even when unchanged.

A `drained` marker is written only after the source meter, browser response-body
observers and adapter body readers have drained and every recorded invocation
has a completion. Closing the file does not establish settlement. Cancellation
may complete with code 0; an unanswered invocation blocks settlement.

Artifacts live in `artifacts/first-use439/<token>/`: `report.json` retains the
native source evidence; `gotogether-requests.jsonl` is the independent transport
and DOM protocol; the two Go test logs retain their separate verdicts. The shadow
uses the isolated `verification/gotogether` module. This document describes the
implemented recorder; executed acceptance and negative-control evidence is
recorded in VALIDATION.md after the complete trial runs.

The same run also reuses the bounded existing `raw-gauge-observations.ts`
Bombadil actor against the owned bank, without another fixture or feature
campaign. Its existing state extract additionally carries actual sent raw
requests and their own completions for `qdriver/bombadil` mapping. Its six actor
phases are a separate adapter check, not a claim to execute the thirteen-phase
retained-return scenario. Actor source costs remain visible as `actorSQL`,
separate from the authoritative thirteen-phase `sql` accounting.

Question identity is observed from the staged service/environment controls and
the displayed instrument read label; it is separate from the fixed-window clock.
Question and inspector text remain readable in the original protocol. The
binding uses gotogether's built-in `intern` identity abstraction for these
arbitrary strings because Quint string literals cannot represent escaped quotes.
Interning preserves equality within the run; it does not validate question
structure or inspector semantics. The shadow bank identity is the DOM source-read
receipt; the native bank/row oracle remains responsible for checking the actual
retained row population.

Executed 2026-11-14 with the frozen readable protocol and upstream interning:
`artifacts/first-use439/6e20fa05c382`. The browser journey passed in 1.3 minutes;
the native gate passed first (10.890 s), then the shadow gate and its controls
(24.490 s), then Bombadil mapping (0.030 s), with zero skipped checks. The native
thirteen phases retained their 24 source SQL calls; the one actor recorded 579
snapshots and 9 separate source SQL calls. Runtime and owned database cleanup
both completed. One cold replay of the same 23 protocol records measured
3.621 s with source evaluation and 30.120 ms with the compiled evaluator; the
compiled positive also passed with Node and Quint absent from PATH. These
single-run figures do not establish a benchmark or justify replacing the native
gate. See VALIDATION.md and the acceptance review for controls and limitations.
