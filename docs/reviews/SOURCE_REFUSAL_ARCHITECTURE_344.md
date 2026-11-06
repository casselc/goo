# Ordinary source refusal — 344

Status: source-grounded closed contract after 343 (`36cd951`); implementation and
runtime acceptance are separate. This cut explains failed ordinary acquisitions
without inventing missing observations or weakening query/read ownership.

## Current boundary and selected scope

The record HTTP family already validates before execution through shared
`queryhttp.ReadHandler`. Admission/decoding failures are 400; runtime read errors
currently become 502 with an untyped `error` string. Ordinary `/api/explore`,
`/api/record-aggregates` and `/api/explore/counted-values` share that handler.
The first two execute different projections; the third supplies independent
count/group observations. Keep their query/result protocols unchanged.

`exploration.Source.read` owns transport, upstream HTTP status, the local 8 MiB
response cap, JSON-line object decoding and 1,001-row safety cap. Fetch and aggregate
adapters additionally validate count/measure/group payloads and project-specific
budgets. Route deadlines, the 60-second shared admission ceiling, transport timeout
and user cancellation are distinct owners. The browser's Queries helper already
checks abort before acquisition and after JSON consumption, and sends best-effort
server cancellation without adopting a late completed result.

Implement a single typed failure family at these existing boundaries, not message
classification in every widget or a replacement source executor. Ordinary coverage
must include the three routes above; shared-mechanism adoption elsewhere must not
promise specialized widget recovery without its own checks.

## Closed refusal semantics

A runtime read failure retains HTTP 502 and supplies the agreed closed envelope:
`error` is concise safe display text; `refusal.kind` is one of the following;
`refusal.detail` is bounded sanitized technical context. Exact DTO details belong
to the implementation schema and must agree between backend, shared browser
parser and OpenAPI; no legacy aliases are needed.

| Kind | Authority | Meaning |
| --- | --- | --- |
| `cancelled` | `errors.Is(err, context.Canceled)` or the checked owning context cause | The acquisition was cancelled; no completed result is supplied |
| `timeout` | Deadline identity or a checked transport timeout interface | The acquisition did not complete within its time budget |
| `budget` | Explicit typed local budget refusal | The result/acquisition exceeded a known budget; no partial result is promoted |
| `malformed` | Explicit typed source decoding/protocol failure after acquisition | The returned payload cannot support the requested observation |
| `unavailable` | Transport/upstream availability failure, or safe unclassified runtime failure | No usable observation is supplied; the cause is not guessed |

Classify wrapped errors with `errors.Is`/`errors.AsType` (available in Go 1.27.3)
and preserve unwrap identity. Check cancellation/deadline authority before a
wrapper's general category. Context cause is useful only from the request that
actually owns the failing acquisition; do not classify an unrelated completed
error using a later cancelled global context. Keep transport timeout and local
route deadline evidence honest without exposing arbitrary URLs/credentials/query
parameters. Fixed typed details and bounded safe numeric metadata are preferable
to forwarding `err.Error()` from HTTP transport.

Upstream non-200 currently discards the response body. HTTP 500/503 alone does not
prove a ClickHouse memory, result or execution budget refusal. Such responses
remain unavailable unless a separately checked machine-readable source code is
introduced. Never search raw exception text for budget/timeout words. JSON syntax
errors and invalid aggregate/count shapes are marked malformed at their actual
decoder boundary, not inferred from message wording later. Compiler/caller
validation is not malformed source data.

A refusal is not `no_data`, count zero, healthy, failed application traffic,
complete telemetry or a retryability guarantee. In particular malformed output
can be corrected upstream without changing the question; budget conditions can
also change. All reacquisition remains an explicit user command. If a retry hint
is exposed, it is guidance rather than a prohibition or proof of permanence;
prefer deriving concise revision guidance from the category over a second policy
state machine.

## Independent admission and refusal boundaries

Keep 400 invalid query/request-body failures and 429 admission capacity distinct
from a source runtime refusal. `/api/compare` is a pure supplied-summary evaluator;
its 1 MiB request-body cap is not a source budget. Do not raise body/query limits,
retag that failure as source acquisition, automatically shorten the window or
silently drop grouping values to make a retry succeed. Changed-between-reads
success diagnostics remain successful diagnostic payloads, not transport failure.

## Browser ownership and presentation

One shared parser/presenter consumes the typed envelope and supplies consistent
category text/details. Widgets do not match raw messages, infer a category from
502 alone, or duplicate a public error taxonomy. Unknown/missing refusal metadata
may produce a safe generic failure; it must not be coerced into a known cause.
Browser protocol/echo validation errors remain distinct from a source-declared
refusal even if both need a visible unavailable state.

The existing acquisition owner is authoritative: exact fixed question, generation,
serial, field, side and visit identity must match before publication. Local abort,
departure or Back retires the owner and cannot publish a late refusal. A genuinely
current remotely cancelled read may report stopped/cancelled without implying a
measurement. No automatic Retry, source subscription or delayed source restart is
introduced by rendering the status.

For paired facets, a failed operand stays unknown while the other ready operand
retains its exact counts and receipts; ready partition absence is still zero.
Explicit side Retry uses the same checked frozen question/field. It cannot update
the held comparison's independent receipt or the other operand's observation.
A refused fixed-baseline read cannot replace an earlier held pair with zero or
pretend its source clock advanced. Record-page, whole-population and comparison
status remain their separate owners; staging/local Find/Back still read nothing.

## Acceptance

- Typed Go tests exercise all kinds, wrapped cancellation/deadline/timeout
  precedence, safe unavailable fallback and bounded details; validation remains
  before IO. Successful empty and changed-read payloads retain their old meaning.
- HTTP cases cover the ordinary three routes, distinguish 400/429 from runtime
  502, and prove refusal does not freeze/publish an incomplete result. Cover
  malformed count/measure/group payloads as well as malformed JSON lines.
- Shared browser tests accept the closed DTO, reject unsupported shapes safely,
  preserve exact query/side echo gates and keep local AbortError quiet. Unknown
  category or raw-message text cannot trigger a category-specific assertion.
- A genuine OTel journey with a controlled source fault shows partial paired
  values, typed reason/details, explicit same-question side Retry and restored
  observation. Separately exercise cancellation/departure with late completion:
  no stale status, no retry and no extra Back read. A representative refused
  aggregate keeps prior held evidence separate from failed new acquisition.
- Reuse existing observed lifecycle checks and independent query/receipt oracles;
  prove negative source-count/late-owner/zero-substitution mutations fail. No new
  model is necessary merely to enumerate an error code.

Keep actual source fault provenance, test results, storyboards and screenshots
current. Full arbitrary upstream exception decoding, distributed retry policy,
backoff automation and specialized signal recovery remain outside this cut.

## Frozen implementation source review

The implemented wire envelope is `{error, refusal: {kind, detail}}`; there is no
retry enum. `internal/sourcefailure` retains internal unwrap identity while its
public description uses fixed safe strings (detail cap 256 UTF-8 bytes).
Cancellation/deadline identities and typed transport timeout checks precede the
closed typed kind. Unsupported kinds fall back to unavailable without exposing
the original cause. `errors.AsType` uses the current Go runtime directly.

The ordinary Fetch, RecordAggregates and CountedValues decoder boundaries apply
Protocol wrapping after successful validation/compilation. Typed transport and
known budget failures remain their original category; count/group/measure decoder
failures become malformed. The shared read handler emits the structured 502;
400 validation and the separate admission wrapper remain unchanged. Source
non-200 responses remain unavailable, without raw body/prose interpretation.

The browser protocol module remains pure; a separate display-only renderer uses
the shared result-state component and technical disclosure. Facet models admit
refusal metadata only on error states, never on ready partitions. Existing exact
question/side/serial and signal-abort checks remain the publication authority.
No successful observation or read receipt is added to a refusal. Source review
accepts these boundaries; `git diff --check` passes. Runtime, observed source-fault
and browser acceptance are reported separately in validation, not inferred from
this review. The parser requires HTTP 502 for the typed runtime envelope; refusal-shaped
payloads on 400/409/429 remain admission errors rather than impersonating a
source refusal.
