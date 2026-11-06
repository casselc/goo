# Independent service-widget recovery — design review 354

Status: implemented source and all four current actual-frame reviews accepted.
Whole-product checkpoint 355 follows this cut's execution gates; this review starts
no additional feature.

A partial service overview should let the user recover one failed read while
continuing to trust the observations that already succeeded. Reuse the existing
widget card, typed refusal/details surface and bounded worker queue. Do not rerun
an entire overview merely to make one card usable.

## Source findings

Before 354, `ui/service-dashboard.mjs` acquired supported widgets with two workers,
collected independent results, then derived overview/profile/timeline state. A
failed card concatenated `error.message` and offered no targeted Retry. Reapplying
the profile reread successful siblings and replaced their clocks. Not-observed
prerequisites were already separated from executable widgets; retain that distinction.

## Card interaction

Show the widget's existing title, one concise typed refusal reason and collapsed
Technical details. Add a current-card **Retry** action with an accessible name such
as Retry span activity or Retry database duration. The visible card title provides
context; do not add a new global recovery toolbar or repeat long scope prose.

Retry explicitly reacquires that widget's existing checked question in its fixed
scope. Keep original service/namespace/environment, profile/binding, chosen native
instrument and UTC bounds. It neither refreshes now, replans another profile nor
increases source limits. Existing query admission remains authoritative.

While queued/reading, disable duplicate Retry, label the card's pending state and
use the existing loading/status semantics; `aria-busy` may describe that owning
region. Preserve the maximum two concurrent supported widget reads, including
retries. Queueing does not imply automatic retries or hidden reacquisition on
render/disclosure/Back.

## Independence and overview truth

Healthy sibling values, DOM presentation/disclosures, questions and exact read
clocks remain untouched. A successful retry has a new independent receipt for
that card only. Failed retry stays unavailable with its current reason; it never
becomes zero or no-data. A successful zero read remains no-data/coverage unknown.
Changed-between-reads stays diagnostic and cannot become healthy merely because
HTTP succeeded.

Recompute aggregate overview/profile/timeline availability from the current result
set after each owned completion, without rereading healthy inputs. Partial means
at least one required attempted widget remains unavailable; an all-recovered view
still says independent reads and unknown coverage/health rather than healthy.
A retained successful native instrument is not merged into a fleet value.

Prerequisites marked not observed, unsupported, or ineligible do not receive Retry:
there is no admitted failed source question to recover. Their existing discovery/
profile meaning remains explicit. A missing prerequisite is not a source outage.

## Focus, accessibility and retirement

If the user activates Retry and focus still belongs to that pending action when
its card is replaced, restore focus to a current meaningful card heading or action.
A temporarily focusable heading is an acceptable fallback. Do not focus a detached
old Retry, jump to page top or steal focus if the user moved to another widget.
Status announcements should use one live region, not nested duplicate announcers.
Keyboard disclosure and current Retry remain visible and unobstructed on narrow
screens.

Profile/range/instrument changes, service departure, Close/Back and newer requests
retire old card work before abort. A late success/refusal or old reattached handler
cannot mutate the current card/result set, update overview status or issue another
read. Identical-question ABA returns still need a current owner. No result token
or cached card becomes a cross-visit source capability.

## Acceptance

Use genuine service overview with one refused record widget and independently
successful sibling/native evidence. Retry only the failed question, verify a new
card receipt and exact unchanged healthy DOM/clocks, then retain/retry another
failure with at most two active requests. Check duplicate Retry, source refusal,
wrong question/identity, scope departure and late/old-handler negatives. Do not
claim every specialized native path is covered by one record-widget fixture.

Actual desktop/narrow frames should show concise refusal plus card Retry, retained
healthy data and the recovered state. Source ledgers establish affected-widget-only
IO, queue limits and immutable sibling receipts; frames establish readable labels,
focus and no overflow. Reuse actual existing automatic overview widgets rather
than add a testing-only dashboard surface.

## Frozen source acceptance

The implemented card owner uses the shared typed refusal and safe disclosure,
current-card Retry, an aria-busy pending state and the existing two-slot queue.
Only source refusals expose Retry; unsupported/not-observed prerequisites do not.
Owned completion replaces that card/result entry, derives overview/profile/timeline
state from the result set and leaves successful sibling cards intact. Profile/
service departure retires queued and in-flight authority. Conditional focus moves
to the current card heading only when the activated retry still owns focus.
Source review accepts this bounded recovery path; it is not a new generic executor.

## Actual desktop/narrow frame acceptance

Viewed the actual `354-record-refusal.png`, `354-native-refusal-390.png` and
`354-native-recovered.png` at original detail. The record refusal card has a short
Source unavailable reason, open safe Technical details and Retry. Healthy log
count 2 and exact native gauge value 42 remain alongside; their provenance is not
replaced with failure. The narrow native refusal keeps Retry and details readable
inside the card while the healthy log's open provenance remains above.

Recovery shows the native exact stream with value 42 and its latest-observation
coverage, retaining the independent/no-fleet-aggregation explanation. The log card
still shows count 2 and open provenance. No transport URL, overflow or misleading
zero/healthy state is visible. All three current frames pass visual acceptance.
These are task-positioned frames; exact unchanged clocks, affected-card-only IO,
queue limits, focus ownership and late-response refusal require the native/model
checks and are not inferred solely from appearance.

Reviewed the fourth actual `354-replacement-service.png` at original detail after
the workspace-departure correction. The replacement service shows two captured
spans, its separate log series and native stream value 7, with independent-read/
unknown-health text and no inherited refusal from the departed service. The task
frame is readable and passes visual acceptance. Correct cancellation and refusal
of late old-service work are established by native ownership checks, not merely
by this replacement appearance. All four current frames are now visually accepted.
