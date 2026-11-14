# 436.4 contextual CEL editor review

Accepted for contextual record condition authoring after source, focused checks
and final regenerated screenshot review. No open correctness findings remain. This is not acceptance of persistence/provenance (436.5),
physical assistive technology or whole-product 436.6 conformance.

## Reviewed behavior

- Activity keeps its stream-first entry. Conditions / Expression (CEL) lives
  inside the existing Filters disclosure; fixed questions reuse the same filter
  editor and their existing visibility owner. There is no new CEL navigation,
  permanent toolbar, primary Run, Save or publication action.
- The shared presentation receives owner preparation and planning functions;
  application adapters perform the source-free context/check transport. Go
  computes validated base content identity. Context target support is static,
  not source readiness or authority to execute.
- Native labelled select and textarea remain persistent. Mode switching preserves
  draft text and caret/undo rather than rebuilding the input. Associated hint and
  diagnostic IDs, `aria-invalid`, explicit-check announcements and focus retention
  are implemented. Typing blocks dependent execution quietly; it does not announce
  a compiler diagnostic per keystroke.
- Conditions are hidden in Expression mode, and natively disabled/inert while a
  text draft owns editing. The visible discard action deliberately keeps the
  current ordinary filters while removing text ownership. Clear condition is a
  distinct operation that restores only the immutable parent.
- Real native availability is checked again after asynchronous preparation and
  planning, including hidden/inert/disabled ancestry before a queued disclosure
  toggle can update the adapter. Private owner tokens reject stale results,
  errors and direct callbacks. Hidden visibility does not silently capture a new
  parent on return.
- External parent changes retain copyable text but lock checking until explicit
  discard/re-adoption. The adapter never guesses which AST clause to remove or
  silently appends the old expression to a new base. Same-owner edits replace
  against their original parent.
- Existing Apply/Run retain source execution. Invalid/pending authoring withdraws
  both the application primary action and native tail Start/Resume paths. Held
  valid observations remain separately inspectable.

## Findings resolved during review

Reviewed corrections include post-await native availability, stale clear/preparation
error handling, external-parent compounding prevention, enabled-but-inert Conditions
controls, `[hidden]` CSS precedence, primary/tail bypasses (including held fixed
Next/Tail callbacks) and duplicated live announcements. Short parent facts replace repeated generic scope prose.

## Independently executed checks

Managed Go 1.27.3:

```sh
go test ./querycel -run '^TestRecordContext' -count=1
go test ./internal/celhttp ./cmd/goo -run 'Test(Context|CEL|PlanRoute)' -count=1
```

These pass (0.006s, 0.020s and 0.032s). The first earlier combined regex selected
no `querycel` tests; the corrected `TestRecordContext` run above is the domain
check. No complete Go-suite, OpenAPI, real ClickHouse or browser result is implied
by these focused reviewer commands.

```sh
node --experimental-test-module-mocks tests/cel-expression-editor.test.mjs
node --experimental-test-module-mocks --test tests/cel-authoring.test.mjs tests/fixed-question-editor-358.test.mjs tests/tail.test.js
```

Eleven component cases pass, zero skips (final diagnostic-presentation reviewer
run 0.044s). The three initial adjacent file runners pass
(0.374s). A final retained-fixed-action regression run plus fixed visibility and
tail file runners passes again (0.558s), including actual extracted Next/Tail
callbacks under invalid/draft/applied states. Component tests use a supplied DOM/native-availability harness and
injected planning responses; they are not live telemetry or accessibility-tree
measurements. Final root validation records the full frontend and browser runs.

## Bounded exploratory actor review

The new Bombadil actor uses real mode/text/check controls, never Apply/Run/save.
It checks native default Conditions, retained invalid text, diagnostic linkage,
primary-action withdrawal, narrow overflow and a retained direct Check callback
after its disclosure closes. Its observational `fetch` wrapper preserves the original receiver, arguments
and returned response, while counting invocation paths/methods including pending
or aborted calls. Terminal acceptance requires actual context and plan requests,
then no extra request after the hidden callback. The actual actor execution is
recorded in parent validation; this source review alone does not prove that
execution happened.

## Visual and accessibility limits

Fresh 1440/390/320 Activity entry and valid-expression frames, plus fixed
390px editing, were visually inspected. The collapsed Activity editor retains
the stream-first view; the shared editor uses coherent native controls and a
compact, vertically resizable textarea. Supplied parent filters remain readable
without horizontal overflow.

The first invalid-expression frame exposed duplicated raw parser text above the
filters and below the textarea. The reviewed correction uses a compact structured
code/line/column summary and optional native Diagnostic details, retaining full
raw compiler text only once. The final 15:10 regenerated invalid-expression
390px frame is accepted: the short outer execution hint, linked local line/column
summary and closed Diagnostic details replace the duplicated parser dump. The
persistent textarea keeps its value and visible invalid border; dependent primary
and tail actions are visibly disabled. The final fixed-expression 390px frame
uses the same compact controls and readable immutable parent facts.

Inspected frames include `4364-live-entry-{1440,390,320}.png`,
`4364-expression-{1440,390,320}.png`, `4364-invalid-expression-390.png` and
`4364-fixed-expression-390.png` under `docs/journeys/screenshots/`. This is visual
inspection of real journey captures, not an independent second browser run. Native semantics, persistent input and association checks reduce
predictable accessibility failures; physical screen-reader/keyboard hardware,
all browser engines and zoom configurations are not independently tested here.


## Nonblocking follow-on

The application-level blocking hint and the precise local check result can both
announce after an explicit refused check; typing is quiet. A physical assistive-
technology pass can assess whether to silence that generic outer hint and retain
only the associated local result. This is not a correctness blocker or a claim
that physical AT acceptance has occurred.
