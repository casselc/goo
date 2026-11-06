# Contextual source-read recovery — 345

Status: source-grounded implementation contract after 344. This extends the
existing refusal vocabulary into ordinary investigation continuations; it does
not add a trace engine, retry scheduler, result store or new receipt.

## Existing owners and scope

| Surface | Acquisition and scope authority | Existing withdrawal |
| --- | --- | --- |
| Trace-wide search | `/api/trace-search`; original span window/environment, one–four checked conditions, bounded stable trace list | Revision, AbortController, exact parent identity/availability, pending scope owner and close/visibility withdrawal |
| Trace expansion | Injected `/api/explore` fetch; exact trace ID, all services, original checked environment/window, span-duration/status projection and page offset | Context generation, trace generation, kind controller, dialog close/invalidate |
| Exact linked logs | Injected `/api/explore`; same trace ID and optional selected span ID, original checked environment/window, independent log page | Context/log generation and separate log controller; trace replacement aborts logs |
| Surrounding logs | Injected `/api/explore`; Discovery's bounded radius clipped to original window, exact service and namespace, independent log page | Surrounding generation/controller, close and trace/context replacement |

Exact linked logs are reported identity context; surrounding logs are time
proximity. Neither establishes causality, completeness or unique invocations.
Surrounding context intentionally omits the original message/severity/trace/package
and arbitrary attribute filters; retry must not accidentally restore or broaden
those omitted constraints differently. Missing namespace and present-empty remain
separate checked identities. Preserve current paging limits and typed environment
selection, including effective/current-only precedence.

`RecordResults.readContext` delegates `fetchQuery`, owns separate kind controllers
and rejects completion after signal abort. Ordinary active exploration already
parses typed source failures after 344. Saved-record and live-inspection wrappers
still flatten errors in the initial source; parser adoption there is required
before those entry points can claim equivalent typed coverage. Either include
those small wrappers deliberately or document their remaining scope. Do not
claim renderer adoption alone changes injected acquisition behavior.

## Closed recovery contract

Reuse the pure 344 refusal parser and passive reason/details renderer. Runtime
502 refusal is distinct from invalid-input/admission 400/409/429 and browser
protocol/echo mismatch. No widget classifies raw error messages or infers absence
of spans/logs from a failure. Technical detail is safe fixed boundary description,
not raw source exception text or a replacement source clock.

Trace search retains its existing explicit Run gesture as recovery: when the
parent and conditions remain unchanged, it rebuilds the same checked request.
Do not add a second search/retry executor. Editing conditions retires the previous
failure/request; a late response cannot become results for the new conditions.

Trace expansion needs an available explicit recovery gesture because failed
initial acquisition leaves Next/Open-as-view disabled. Show the reason and retry
in the primary inspector region, not only inside collapsed evidence details.
Bind Retry to the exact failed request, including offset, trace ID, checked policy
and original window. The same applies to linked-log and surrounding-log retries;
changing scope/radius/selected span/page creates a different owner rather than
silently retrying the old failure through new controls.

Preserve the existing trace continuation: `traceRows` success subsequently calls
`logs()`. A failed trace read makes one acquisition and no linked-log read;
a successful explicit trace retry may make the trace read and then the existing
linked-log read. Explain that short continuation in the UI. A linked-log Retry
alone must acquire only logs, retaining the successful trace observation. Do not
add an alternate trace mode merely to make the retry appear to perform one read.

Retries are local intent owners, not automatically authorized effects. The handler
must match the current dialog/context, generation, failed-question identity,
connected/visible control and pending state before reading. Retire/replace that
handler at retry start, context/scope change, cancellation, Close, Back and dispose.
A saved old handler cannot revive on same-trace ABA revisit. Failed/aborted reads
supply no observation, no count zero and no new receipt. Clearing old failed rows
is allowed; held unrelated observations must not be relabeled as freshly read.

## Publication and display pitfalls

- Each catch must check the owning generation and local abort before status or
  retry publication. A current remote cancellation may display the typed stopped
  reason; local departure/cancel remains quiet under its retired owner.
- Trace failure disables trace comparison/Open-as-view authority. Linked-log
  failure does not invalidate independently successful trace data or its clock.
- Retry does not reuse a previous successful response as the answer to a new
  request. Response query echo and source admission remain checked before adoption.
- Use one live-status announcement per visible reason; nested shared components
  must not create duplicate announcements. Details disclose without IO. The
  recovery button remains keyboard reachable and receives sensible focus only
  when the current owner still owns focus.
- Repeated Retry clicks, detached controls and late completions must not add
  requests or move focus into an obsolete context. Failure text is not a source
  result and must not be fed into waterfall/list/trace completeness widgets.

## Acceptance and cadence

Use genuine ordinary OTel records and a controlled upstream failure, covering at
least one trace-wide and one log-context recovery. Record actual source requests,
full frozen query definitions, independent clocks/bases and failed/current owner
transitions. Verify the trace retry's established two-read continuation separately
from one-read log retry. Cancel/Close and same-trace ABA with late completions must
leave zero stale status adoption or automatic retry. Include strict typed-protocol,
response-echo and old-handler negative controls.

Reuse existing search/context lifecycle models where their observed projection
fits; a new refusal code alone does not warrant a new model. Mounted owner tests
should exercise all new retry handlers, while the native journey demonstrates
representative real source semantics and actual desktop/narrow primary-status
visibility. Storyboards and screenshots must distinguish failure from successful
empty trace/log data.

After this breadth cut, revisit the already measured facet-publication work only
through operation-level render accounting: the 32,000-node cost and multiple
owner/Back builds are established, but the causal contribution remains to be
measured. That is a bounded next depth/consolidation candidate, not authorization
for a new widget catalogue or virtualizer. Final cadence selection should use
345's observed reliability/usability findings before starting another feature.

## Frozen-source acceptance

Source review accepts the implemented bounded continuation changes. Trace search
uses the shared parser/renderer with its existing Run/revision owner. RecordResults
shows trace failure in the primary inspector summary and creates a new Retry
button bound to the failed context, trace generation, offset and actual connected
button identity. Starting retry advances the owner and retires the prior handler.
Successful trace retry retains the existing span-then-linked-log continuation.

Linked and surrounding controls receive generation-bound handlers, preserve the
failed offset for the unchanged intent and disable Run while their read is
pending. Scope/radius changes replace the intent; local abort and late responses
remain fenced by the original controllers/generations. `openTrace` now detaches
the origin with `structuredClone`, so caller mutation of nested environment
selection cannot change a later retry under the same owner. Saved-record and
live-inspection acquisition wrappers also preserve the typed error; live polling
is outside the change.

No backend wire, source budget or read-receipt protocol changed. Whitespace
validation passes. Runtime/source-fault, semantic regression and actual-frame
results remain separate scheduled acceptance evidence; no new model is selected
for the unchanged source ownership protocol.
