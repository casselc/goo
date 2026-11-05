# Whole-product architecture checkpoint — 332

Status: independent source and maintained-evidence review after 331. No product
source edits, new native run, latency measurement or source-capacity test. Read
against the active Next section of ROADMAP_2026_11_03.md, the 312/317 reviews,
331 architecture/design, START_HERE, VALIDATION and known limitations. Harness
completion gates are reviewed separately; this report does not infer campaign
success from process exit.

## Recommendation

Close the remaining **general saved-capture acquisition ownership** gap now.
A pending general capture can outlive departure to the Saved investigations root;
paired browser notifications can also start competing acquisitions of the same
capture. This affects the general preserve/reopen task rather than adding another
kept-record action. Keep the correction inside captures.mjs, with focused actual-
owner tests and one native preserve/reopen/leave pass. Do not create a universal
navigation manager, receipt cache or executor framework.

This is a source-confirmed missing guard, not a newly measured browser incident.
The 331 ledger's nineteen capture-catalog GETs do **not** establish nineteen
duplicate reads. Their request intents, routes and query parameters would need
separate attribution before any catalog optimization claim.

## Actual core ownership today

| Task boundary | Current owner and authority | Boundary to retain |
| --- | --- | --- |
| Activity entry and scope | stream-entry.mjs owns applied selector, completed occurrence windows, staged filters, pause and retained return anchor; tail owns the observed buffer | Apply/Start changes acquisition; a held row or brush is not a population denominator |
| Activity to fixed/service | Analyze stages the completed applied window; ServiceHandoff checks qualified identity; continuation owners retain the paused origin | Staging does not read telemetry; incompatible predicates cannot silently become an exact cross-signal scope |
| Fixed population Run and child | population-workspace.mjs owns generation, query copies, selected/full reads, child parent retention and independent baseline reads | One source owner; passive charts/controls do not fetch; Back retains the held parent |
| Compare | Activity baseline loads immutable capture evidence; Activity explanation explicitly rereads a saved scope; fixed query comparisons execute their baseline independently; shared count presentation is passive | Reusing frozen operands and rereading the same question are distinct actions, clocks and certainty |
| Preserve | preservation.mjs supplies passive meanings; population actions save configuration, share questions or dispatch capture form intent; captures confirmation rereads the chosen question | A saved view is configuration. Capture confirmation is a new read; Keep retains original occurrence references |
| Reopen and continue | captures.mjs acquires/validates stored pages; saved-records.mjs renders frozen results and stages fresh questions; kept route has the 331 visit owner and retained inspector continuation | Stored paging reads custody, not telemetry. Fresh exploration leaves original receipt bytes unchanged |

These distinctions are useful and already exercised in the maintained 312–331
journeys. Typed environment propagation, source adapters, passive presentation,
original-reference selection and explicit comparison admission should remain.
The consolidation opportunity is at asynchronous acquisition admission, not a
replacement of these semantics.

## Concrete remaining gaps

### 1. General saved capture has weaker route ownership than its kept sibling

In captures.mjs, sharedLink (currently line 88) listens independently to hashchange
and popstate. For #capture=ID, it calls open whenever current has another ID or is
unset. open (line 66) increments viewGeneration and requests the receipt without
an AbortSignal. Consequently, two notifications before completion start two GET
attempts; the generation check prevents the earlier completion from rendering,
but does not join demand. For an already loaded capture, each notification calls
show(current), which rerenders saved records and their inspection state.

There is also a stronger correctness case. While open(ID) is pending, depart from
#capture=ID to #investigations. The goo:workspace handler (line 82) takes its
investigations branch and calls library; it does not advance viewGeneration.
sharedLink sees no capture ID and retires only kept ownership. The pending open
therefore still satisfies its generation check. show(page) can reveal the old
capture and replace the hash with #capture=ID after the user has left it. A route
change to another top-level workspace *does* advance viewGeneration; document and
retained-evidence display events also advance it. The missing case is departure
within the same workspace, not all navigation.

This is a direct control-flow result from current source. No new native request
count, inspector-reset observation or latency improvement is claimed here.
The 331 kept owner already handles this distinction correctly: retire before
abort, admit success and failure by visit identity, join pending/loaded events,
allow explicit retry and reacquire after a real departure.

Smallest correction: give general capture opens an owner-local visit covering the
exact capture ID **and stored-page cursor**, abortable read and completion
admission. Passive paired notifications join the current page visit; an already
rendered visit is not rebuilt. Different capture/page, same-workspace route
departure or another evidence owner retires it. Explicit Open/Retry is an explicit
intent, with failure terminal only for passive notifications of that visit.
Preserve the current explicit page/error behavior and distinguish write completion
from read adoption: an already authorized capture/refresh POST must not be treated
as an abortable GET or silently retried.

Do not merge general and kept renderers or introduce cross-visit reuse. Reusing a
small private visit idiom is reasonable if it reduces actual duplication; a new
public abstraction is unnecessary. Prerelease status permits removing obsolete
branches rather than maintaining parallel compatibility callbacks.

Acceptance must hold both success and failure responses after route departure;
include pending same-ID joining, loaded-node identity, distinct-ID supersession,
A→root→A fresh visits, same-ID different stored-page cursors, failure followed by
explicit retry, ordinary link behavior, and saved inspector focus. Count custody
GET attempts separately from catalog metadata, cancellation controls, source
reads and writes. Preserve/reopen should retain the exact typed question and
original evidence; only the explicit later Run/Refresh may read telemetry.

### 2. Fixed saved-view restore repeats the same ownership pattern

population-workspace.mjs restore (currently line 701) registers both route events
and invokes load(id) on every notification. load (line 582) immediately clears
results, advances generation, aborts the preceding controller and fetches the
saved configuration. A duplicate event therefore restarts acquisition; a passive
notification after Run can discard held results and restore configuration-only
state. Abort/generation prevent stale completion but do not establish one route
intent or protect a loaded visit from passive restoration.

This is source evidence of repeated work and destructive reload semantics under
those notifications, not proof that every normal navigation currently emits both
or a measured source performance regression. A saved-view GET is configuration,
not a telemetry read. Do not present its removal as an SQL saving.

This is a known remaining gap, explicitly excluded from the selected 332 cut.
Apply the same bounded pending/loaded/retired distinction in a later adapter-sized
change. Retire on staging another question, route exit,
and explicit replacement. A passive route event must never reset a user-authored
draft or successful Run. An explicit reopen must remain possible. Do not unify
this with Activity's rolling selector restoration or retained child Back.

### 3. Catalog refresh demand needs intent measurements before consolidation

captures.mjs reloads its general catalog on each investigations workspace event;
workspace.mjs dispatches that event on both hashchange and popstate even within
the same route. This is a concrete possible repeated-demand path. However explicit
capture/refresh completion, catalog filters, pages, another evidence visit and
normal saved navigation can all legitimately refresh metadata. The existing
nineteen-GET total does not distinguish them. No cache, global event suppression
or catalog deletion is justified from that number.

If demand remains material during the selected native task, log each request's
owner, intent, applied search/cursor and triggering transition. Only eliminate a
proven redundant automatic refresh, preserving explicit Find/Refresh and post-write
visibility. This measurement is bounded acceptance work, not an observability
subsystem project.

## Revised finite tick/tock sequence

1. **Tock — finish general evidence acquisition ownership.** Implement finding 1,
   focused deferred-response tests and native preserve/reopen/leave evidence.
   Finding 2 stays outside this cut and remains a specific subsequent
   consolidation candidate. Close the separate positive-campaign-observation audit
   with its own evidence. No new user-facing kept action.
2. **Tick — one general observability task.** Select a concrete ordinary OTel
   question across Activity → fixed population → compare → preserve from a short
   actual task pass. A suitable bounded candidate is a logs/spans population change
   across adjacent windows using the existing query/comparison grammar, with clear
   empty/unknown states and supporting-record continuation. This is a candidate,
   not a promise of a new comparison family: review current supported admission
   before committing scope. A missing task must justify any new capability.
3. **Tock — consolidate only the seam that task changes.** Finish saved-view visit
   ownership if not included above, or remove a measured repeated read/render at
   its actual owner. Carry request/SQL/DTO budgets where relevant; do not refresh
   held baselines, add speculative caches or resurrect the broad 312 workspace
   extraction merely because the file remains large.

After these finite outcomes, reconcile again. General task breadth should drive
future depth; do not alternate indefinitely between another kept action and its
local polish. Native metrics, logs, spans and reusable questions remain the product
surface. Deployment, retention/restore and alerting still require their separate
end-to-end contracts.

## Selected implementation review

332 is limited to general saved captures; the fixed saved-view gap above remains
explicitly outside this milestone. The implementation now owns pending/loaded/
failed acquisition by ID and stored-page cursor, joins passive route events,
checks visit identity and generation on both completion paths, and retires before
abort. Passive notifications preserve a loaded page and inspector. Refresh keeps
a loaded owner for the displayed receipt while its authorized POST is pending;
passive events cannot replace it with a GET. Retirement advances display generation
so a completed refresh cannot restore a departed route. The write itself remains
uncancelled and is not retried. Successful Save adoption and explanation rendering renew the
same general receipt owner, and the asynchronous explicit-open focus continuation
checks the current loaded visit and exact route.

Independent review requested the refresh sentinel/display-admission and explicit-
open focus checks during implementation; both are present in the reviewed source.
The initial source review accepted the GET/Refresh owner; a subsequent root
review identified an additional pending-Save display-admission gap: confirmation
checks draft generation, so a completed authorized publication can still show its
receipt after route departure. The current generated fixture covers Refresh POST
admission, not Save-confirm departure. Publication and explicitly requested
attachment must remain authorized even when display adoption is retired. This
was corrected by the bounded Save display lease described below; the initial
review did not cover that boundary and its earlier evidence is not retroactively
claimed to do so.

The new actual-owner fixture mocks passive display,
strict receipt validation and custody transport, so its claims are about acquisition
ownership, not receipt integrity, browser DOM/focus, real source membership or
physical SQL cancellation. Native evidence must establish those separately.

The initial GET/Refresh focused command passed in **5.601 seconds**, zero skips
before the pending-Save display lease was added:

```sh
PATH=/workspace/.toolchains/go/bin:$PATH GOPATH=/workspace/.go \
GOMODCACHE=/tmp/goo-module-cache XDG_CACHE_HOME=/workspace/.cache \
go test -run '^TestHegel(CaptureRoute332|KeptMatchNavigation331)$' -count=1
```

It runs twenty generated general-capture scenarios plus the twenty existing kept-
owner scenarios. The new independent oracle varies receipt IDs, one to eight
cycles and one to twelve passive notifications, expecting nine exact-page GET
attempts, five adoptions and two authorized refresh POSTs per cycle. It exercises
pending/loaded joins, cursor ownership, distinct-ID late success, fresh same-ID
revisits, terminal passive failure, explicit retry, same-workspace late success/
failure and passive/departed refresh completion. A direct largest fixture run
passes with 72 GETs, 40 adoptions and 16 POSTs across eight cycles. Two temporary
module mutations fail: bypassing visit joining produces `pending duplicate GET`;
removing success admission produces `late distinct-ID success adopted`. Temporary
mutated modules were removed; product source was not altered for these controls.


## Root follow-up: pending Save display authority

The final Save correction uses a private irreversible display lease, independent
of any existing capture visit. It records the submitting hash/workspace and
retires on observed departure (including departure and return to the same hash),
competing receipt opens, other evidence displays, refresh or a new draft.
An unchanged passive route notification keeps it valid. Successful Save temporarily
detaches the lease during its own authorized navigation, then reattaches it to the
new receipt route to guard delayed attachment status. Publication and the
explicitly selected document attachment continue even after display retirement;
they are neither aborted nor retried. A retired completion cannot show its receipt,
restore the hash or publish late status/focus. This remains local application
ownership, without a generic manager.

The independent fixture now includes four Activity-origin Save cases per cycle,
with no active capture visit: departure, departure-and-return ABA, queued
HashChangeEvent.newURL ABA while location already equals the original hash, and
a positive unchanged-route case with duplicate notifications. It holds POST and attachment
completion separately, checks exact requested attachment references, and verifies
that retired displays receive no adoption, route restoration, status or focus
call. This tests display-adapter calls; native browser focus remains separate.
The largest direct run passes with 72 GETs, 48 adoptions, 48 POSTs and 32 attachments
across eight cycles. A third temporary mutation bypassing the Save display guard
is rejected by `departed Save adopted receipt`. Earlier joining and GET-admission
negative controls are rerun and rejected as well.


Final post-Save source approval is green. The focused command above passes again
in **5.421 seconds**, zero skips, with twenty expanded general-capture scenarios
and twenty existing kept-owner scenarios. The new per-cycle oracle is nine GETs,
six adoptions, six authorized POSTs and four exact document attachments. A fourth
mutation removes queued-event newURL retirement and is rejected by `departed Save
adopted receipt`. All temporary modules are removed.

The initial expanded run found a fixture-oracle collision: a generated next-cycle
ID could equal the preceding fixed Save result ID, legitimately retaining that
loaded visit. The fixture now explicitly departs between cycles; its seed-1093
reproduction and largest eight-cycle run pass. This corrected the test's assumed
visit boundary and required no product change. Final checks include replay of the
recorded failing fixture input. Native task/DOM/focus and general full-suite
results remain separate from these source/property claims.

## Native follow-up: original Save dialog completion

The actual native pass found a separate ownership defect after the display lease
cut: a successfully committed Save with retired display authority left its original
modal busy. Successful publication now closes and clears that original dialog when
its draft generation still owns it, independently of receipt display adoption.
An error recovers controls/error text in that still-owned dialog. A newer draft
retains its own modal and cannot be closed or overwritten by either old outcome.
This intentional modal completion/error recovery differs from publishing stale
saved-workspace status or restoring a departed route. Durable attachment remains
independent of both display and modal ownership.

Independent final source review accepts this correction. Every generated Save
case now asserts the original successful modal closes; focused frontend tests
separately cover newer-modal success/failure and original failure recovery. A fifth
temporary mutation restores display-gated modal cleanup and is rejected by
`original completed Save dialog remained open`. The final combined command passes
again in **4.978 seconds**, zero skips, with the same twenty expanded general-
capture and twenty kept-owner scenarios and unchanged read/write/adoption budgets.
This supersedes the earlier source/PBT freeze time; actual native modal/focus
acceptance remains separate.
