# Source refusal and recovery — design review 344

Status: implemented source and actual desktop/narrow frame acceptance complete.
Native execution results are recorded separately in the journey and validation.

A source read can fail while the user's question remains valid. The interface
should say what could not be read and give the next intentional action, preserving
the original scope and any independently known observations. A transport URL is
technical detail, not the primary explanation. This is an ordinary exploration
improvement, not a new incident/error console.

## Existing components and source

`goo-result-state` in `ui/components.mjs` provides a passive polite status region.
Native `details`/`summary` provides keyboard-accessible technical disclosure.
Reuse these patterns and existing buttons; application owners retain IO, retry,
cancellation and navigation authority. The passive component must not fetch.

Current `ui/population-workspace.mjs` catches many failures by concatenating
`error.message` into Baseline unavailable or population status. This can promote a
transport URL into the primary message. A failed companion or baseline is already
independent of successful results; preserve that separation. Classification
belongs in the agreed closed source-error model, not per-widget message matching.

## Concise reason and explicit recovery

Use one short primary reason, for example “Source read timed out”, “Source read
limit reached”, “Source unavailable” or “Source response could not be used”.
Name the failed signal or baseline in the owning section. Do not invent a timeout,
budget or invalid-data cause from an arbitrary message. Unknown failures use the
honest general reason. Avoid promising that retry will succeed.

Keep the bounded, safe source-boundary explanation in a collapsed **Technical
details** disclosure with a literal text body. The accepted wire contract supplies
a canonical kind/reason and at most 256 bytes of safe detail, never a raw
transport cause, URL or stack trace. The primary surface must not contain
a URL or a stack trace. A keyboard user can open the detail; disclosure adds no
source requests and does not move focus unexpectedly. Long tokens wrap within
narrow layouts. The detail is a bounded boundary explanation, not a claim to expose a full
server diagnostic or raw source response.

Place the existing explicit recovery action near the reason. Its label must
match its actual boundary: **Read fixed baseline** retries the baseline;
**Run** rereads the configured questions. A full Run must not be labelled as
failed-side-only retry. Where a safe current-owner failed-side retry already
exists, retain the healthy operand and its original receipt while retrying only
the failed side. Do not introduce automatic loops or increase source budgets.

## Observation and ownership semantics

- Timeout, budget refusal, unavailable source and malformed response are failed
  acquisition/admission states. They are never measured zero, healthy or No data.
- Cancellation on departure or supersession is expected ownership retirement,
  not a new visible source failure. A late aborted/departed result cannot replace
  the current reason, observation, question, focus or retry owner.
- A healthy operand retains its exact values, checked policy and independent
  read clock. The failed side stays unavailable; withhold unsupported change or
  persuasive comparison geometry. Do not silently mix an old failed-side receipt
  with a new successful retry as if both were one atomic read.
- Preserve the whole staged question, typed environment policy, fixed UTC bounds,
  filters and configured measures. Recovery cannot broaden a scope, move the
  window to now or implicitly save/capture evidence.
- Disable duplicate retry while its current request is pending. Restored Back
  controls must be fresh and working; old detached retry handlers stay inert.
- Technical details, rendering, Back and local edits perform zero replacement
  reads. Only the established explicit recovery gesture issues the advertised IO.

## Acceptance journey

Trigger actual source timeout/refusal through the native test's controlled source
boundary while using genuine ordinary Explore/population scopes. Capture concise
reason and collapsed detail at desktop and narrow width. Open detail by keyboard,
verify bounded wrapping and no IO, then perform the explicit current retry and
independently observe success. Assert identical question/policy/bounds before and
after, separate receipt clocks and advertised source counts.

Include a partial pair whose selected side remains known while baseline acquisition
fails; show no invented difference and recover with the correct retry boundary.
Cover cancellation/supersession, malformed-source and budget cases with semantic
and mounted tests, including old-handler/late-response negative controls. Real
frames establish appearance; observed traces establish recovery, scope and IO.
Do not imply that every error surface in the entire product was migrated by this
bounded first cut.


## Frozen source acceptance

`ui/source-refusal.mjs` admits a closed kind/detail DTO, maps it to canonical
reasons, and downgrades malformed refusal envelopes to honest unavailable. It
does not classify arbitrary error strings. `ui/source-refusal-view.mjs` composes
the existing passive status shell, concise reason and native Technical details;
it owns no recovery IO. Existing application actions retain their actual retry
boundaries and owner guards. Source acceptance passes for this bounded path.

Use one live announcer per surface. If the host already owns a status/live region,
the inner shell should be a non-live group rather than announce the same update
twice. This changes neither visible prose nor recovery ownership. Actual frame
acceptance still needs concise URL-free refusal, the matching recovery control,
and a partial side with known selected counts versus unavailable baseline.

## First actual frame acceptance

Reviewed genuine `344-budget-refusal.png`. The owning comparison section says
“Baseline unavailable · Source read exceeded a limit”, followed by the existing
Read fixed baseline action and a keyboard-focused, open Technical details
explanation. The safe explanation describes the bounded read/grouping limit and
suggests narrowing scope; no raw URL or transport trace dominates the surface.
The two original fixed windows remain visible, no fabricated comparison appears,
and the independent selected population surface remains below. First-frame
appearance passes. This viewport does not independently show the entire selected
count/receipt or prove retry success; native semantic/action checks supply those
facts. Partial-pair desktop/narrow frame acceptance remains pending.

## Final actual visual acceptance

Reviewed `344-partial-values-1280.png`, `344-partial-values-390.png` and
`344-recovered-values.png`. Both partial views clearly retain selected GET 6 and
POST 3 while previous-window counts are dashes, not zero. The primary reason names
the unavailable previous-window values, Technical details stays collapsed, and
Retry previous window values names the failed-side boundary. Comparative facet
bars are withheld while the side is unknown. The previously completed 5/9 summary
comparison above remains a separate valid observation, not evidence that the
failed facet read succeeded.

On narrow width the reason wraps, retry occupies its own row, side labels remain
readable and both Include/Exclude actions stay inside the viewport. Recovery
removes the refusal and reveals the exact GET 6/3, POST 3/0 and baseline-only
DELETE 0/2 union with shared-scale bars. The selected service chart and its original
captured time remain visible. Actual desktop/narrow visual acceptance passes;
retry IO, cancellation, late-response rejection and Back preservation are native
execution assertions, not inferred from screenshots.
