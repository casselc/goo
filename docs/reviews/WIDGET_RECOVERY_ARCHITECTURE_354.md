# Independent service-widget recovery — architecture review 354

Source review for the third cut of the 351 breadth/depth/consolidation sequence.
This document closes the intended contract; implementation and native validation
must be reported separately. No new query engine, automatic retry, source budget,
or generic ownership registry is required.

## Existing boundaries

`ui/service-dashboard.mjs` plans once, builds one article per visible widget and
executes supported questions with two workers. Record widgets use the existing
`/api/record-aggregates` profile; native widgets use `/api/measurements`. `matches`
checks the complete returned question, and record aggregate admission checks the
consumed profile. Widget clocks, basis, state and limitations remain independent.
Before 354, initial failures flattened `data.error`; the shared source-refusal
protocol and passive renderer already provided a closed recovery vocabulary.

`ui/service-browser.mjs` owns service generation, overview revision and an abort
controller per overview run. Profile, instrument selection and range Apply replace
the overview, rather than retrying a single slot. The retained parent-abort
listener already forwards cancellation to the latest overview controller after
initial loading resolves. The separate per-run listener is removed in `finally`;
that removal is not itself a cancellation gap.

## Closed recovery contract

- Retry is an explicit action on a failed supported widget. It resubmits that
  widget's frozen complete query/definition and fixed bounds to its existing
  endpoint. It does not replan, refresh discovery, change profile or reread healthy
  siblings. Unsupported/not-observed prerequisites have no Retry action.
- Initial work and retries share the existing maximum of two concurrent widget
  reads. A widget has one current immutable attempt token. Repeated clicks while
  queued/running do not enqueue duplicate attempts. A saved old failure callback
  cannot revive after a replacement attempt, successful recovery or overview ABA.
- Keep one result slot per planned widget. Replace that slot on recovery; never
  append a second result for the same widget. Preserve healthy result objects,
  articles, disclosures, controls, basis and read clocks.
- A reply may publish only while its load owner, attempt, host and browser owner
  remain current and the owner signal is not aborted. Cancellation caused by
  departure is quiet. No late failure may replace a newer success, and no late
  success may attach to another service/profile/window.
- Aggregate overview/profile status and timeline are derived from current slots.
  Loading is not success or zero; `changed_between_reads` continues to withhold
  incompatible rankings/rates. Recovery of a native widget cannot establish the
  record profile as observed. Missing/unsupported evidence remains unknown.
- Typed 502 source refusals reuse the shared checked protocol/renderer. Ordinary
  admission errors remain ordinary errors; do not classify by message substrings.
  An explicit retry may succeed after an external fix without changing the query.

## Concrete lifecycle pitfall

The browser currently clears `pending.dashboard` when `load` finishes its initial
batch. A later widget retry must expose pending work to that owner too. Otherwise
`suspendPending` can decline to invalidate a pending retry when the user leaves;
the existing `current` predicate checks generation/revision/connected detail, not
visibility. Reuse a narrowly scoped pending callback or equivalent existing owner
hook, guarded by the same overview revision. Count queued as well as active work,
and ensure an old owner's finalizer cannot clear a new owner's pending flag.
The retained parent abort forwarding can then retire the retry without a second
controller manager.

## Actions and custody

Retry only acquires the failed widget. Existing “Capture widget evidence” stages a
capture draft which explicitly rereads the fixed question after confirmation; it
does not archive the currently displayed aggregate response verbatim. “Open widget
question,” Add, Customize, native inspection and timeline/range actions retain
separate ownership and must be inert after retirement. Preserve this distinction
rather than adding an unsupported aggregate capture format in this cut.

## Acceptance to verify

Use a genuine service overview with one refused widget and a healthy sibling.
Prove retry sends only the failed exact question, keeps sibling identity/clocks
and updates status/profile/timeline from one current slot. Exercise repeated Retry,
two-worker saturation, second failure then success, old callback invocation,
service/profile/range replacement and late completion after suspension. Include a
native-widget or focused native-adapter schedule without equating native streams
with record totals. Browser evidence should show the primary reason and explicit
Retry in narrow and desktop layouts. Reuse the existing observed owner/model
protocol where sufficient; do not invent model events for unobserved behavior.

After this finite cut, perform whole-product checkpoint 355 before starting a new
implementation sequence.

## Candidate source review

The 354 candidate uses a load-local host owner, composed abort signal, two-slot
queue and keyed result replacement. Browser pending reporting is guarded by service
generation and overview revision; its finalizer retains the latest load-pending
value, including retries that outlive the initial batch. New pending state is set
after aborting the old run, preventing synchronous old-abort callbacks from
clearing the new request's flag. Derived timeline handlers carry their own render
revision and current/visible checks. Healthy widget articles and results are not
remounted by this derived refresh.

Two adjacent delegated controls require final ownership closure before acceptance:
(1) the browser-created native picker must not let a saved retired submit callback
start another overview run; (2) runtime inspection callbacks supplied to a widget
must retain the dashboard's local current-owner fence, even if a caller does not
supply the service browser's additional generation checks. These are bounded
callback guards, not a new ownership system. Full/native gates remain the parent's
responsibility; this review ran no heavy checks.

The final candidate closes both findings: the native picker callback captures
service generation, overview revision and request signal and requires its current
submit control to be visible; delegated runtime/pool inspection callbacks check
the dashboard's current owner and visible opener. Source acceptance is closed.
The parent runs the full UI/native/model gates; this is source-review acceptance,
not a substitute for those results.

## Native integration finding

Actual workspace navigation exposed a missing application-level connection: the
existing `liveInventory.suspend()` was called by resource-tab departure, but not
when leaving the Services workspace for Explore. Unit tests exercising suspend
directly therefore did not prove the real navigation path retired pending retries.
The bounded `ui/app.mjs` fix listens to the existing `goo:workspace` event after
inventory mounting and calls suspend whenever the destination route is not
`services`. `Workspace.navigate` emits that event after updating panel visibility.
The existing inventory/service adapters remain the retirement owners: pending
catalog work is cancelled, pending service/handoff work follows its existing
retirement path, and completed views are retained because their pending checks
make suspension a no-op. This adds no reads or new lifecycle manager. Source
review accepts the integration fix; the parent is rerunning actual native gates.
