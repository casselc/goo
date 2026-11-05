# Typed trace-wide search — design review 340

Status: final source, four actual policy/continuation frames, native return
visibility and actor acceptance pass.

## Purpose and existing interaction

The current Activity and Services entry points already establish typed environment
scope, including effective named, empty and unreported populations. The existing
**Find traces containing…** action should use that scope directly. A user should
not have to enter a different environment name or understand a transport limitation
to continue an ordinary OTel investigation.

Reuse the current trace-search panel, conditions, results and trace viewer. Keep
this a breadth repair to an existing workflow, not a new search surface, saved
search resource or query language. The distinction between a checked source result
and a fresh exact-trace read stays explicit.

The likely Activity/Services entry is the population workspace, not the old root
question editor. Its current span records need the same contextual action and
local source owner. Merely accepting typed requests in the root search builder
would leave that entry path inaccessible. Reuse the existing record action and
trace viewer, with the originating population retained.

## Scope and meaning

- The existing scope paragraph must identify the actual policy and mode. Use the
  shared normalized environment label, including current-only versus current/legacy precedence,
  Empty environment label and Unreported environment. A bare name cannot stand in
  for an effective selector. Fixed UTC bounds remain visible; exact half-open
  semantics and parent predicates must survive the checked request.
- Current-key presence wins under effective policy, including a present empty
  string. Legacy-only values are eligible where that policy permits them;
  unreported means neither key is present. Do not present empty as unreported or
  quietly reinterpret effective named scope as current-only.
- Every condition means at least one matching reported span in the same reported
  trace ID, selected environment policy and fixed window. Different conditions may
  match different services or the same span. Preserve the existing explanation:
  no ordering, parent-child path, completeness or business identity is implied.
- The original applied filters seed condition 1. Additional conditions can describe
  other services; do not accidentally repeat a service filter across every condition
  or require all conditions to match a single span.

## Results and continuation

The full distinct reported-trace count remains separate from the bounded,
lexicographically stable listed traces. A representative witness is not every
matching record, the trace root or a causal path. Keep the existing qualified
field renderer so a missing key, present empty value and literal string remain
distinguishable. Continue displaying witness timestamps as UTC with exact timing
available, rather than raw nanosecond integers as primary text.

**Open trace (fresh read)** must carry the same checked typed selection, fixed
window and reported trace ID into the existing continuation. It must not drop to a
scalar environment, adopt a mismatched echo, broaden to another policy or present
freshly read spans as the frozen search witness. The continuation expands that
reported trace across services within the same policy/window; it must not apply
condition 1’s service restriction to every expanded span. Search and trace receipts
keep their own clocks and coverage limits.

Scope or condition changes clear obsolete matches. Departure, hiding, cancellation
and late completion must keep their established ownership rules. Returning to the
retained parent restores the supplied search result and focus without an automatic
replacement search. An unsupported origin, such as reported-edge scope, should
retain the existing actionable refusal rather than silently broaden its meaning.

## Native and visual acceptance

Use ordinary real OTel spans without Goo manifests or seeded search definitions.
A meaningful journey starts in a typed Activity or service span population,
chooses an observed field's Find traces action, adds a condition from a different
service, explicitly searches and inspects the checked witnesses. Follow a listed
trace through the existing fresh-read path, then return to the retained result.

Repeat the policy proof with effective named, current-only named, present-empty
and unreported contexts. Include a current/legacy conflict and decoys outside the
selected policy, window and trace. The expected trace population must be computed
independently of the product matcher. Cross-service conditions must match one
trace while a same-service-only or same-span-only interpretation fails a negative
control. Check full count versus a deliberately bounded list.

Review actual desktop and narrow frames of the staged scope, matched witnesses
and exact continuation. No horizontal overflow; scope labels and primary actions
remain readable. Opening a local condition/witness disclosure or staging a
continuation must not issue an implicit source search. Preserve cancellation,
retained return and focus, and independently judge completion rather than a click
or zero-exit process alone. Actor JPEGs are not visual proof unless their current
state has been inspected; maintained Playwright viewports own appearance claims.

The existing count/facet facts and independent clocks must not become noisier as
this flow is repaired. No additional always-visible policy editor, dashboard or
wide explanatory toolbar is required.

See the [active roadmap](../ROADMAP_2026_11_05.md),
[prior breadth proposal](POPULATION_CHROME_DESIGN_339.md) and
[architecture review](TYPED_TRACE_SEARCH_ARCHITECTURE_340.md).

## Implementation review points

Initial source review confirms that the new population-owned entry uses the
existing record actions and trace viewer rather than leaving typed support only
in the older root question editor. Search ownership must retire on definition or
draft changes, child navigation, suspend/hide and disposal. A clean/current guard
alone is insufficient if an obsolete request keeps running in a connected but
covered parent.

The browser scope label should describe normalized selection even for a scalar
current-only origin; the shared `questionLabel` deliberately prints a bare name
for that representation in some older surfaces. The trace-search paragraph can
use the shared normalized label without adding another policy control.

The frozen UI source satisfies this review: normalized policy labels cover scalar
and typed origins; typed requests avoid empty scalar representation placeholders;
the fresh-read callback retains environment fields, exact window and trace ID.
Each witness includes both qualified OTel environment keys and is checked against
the requested policy before admission. The common field-value renderer preserves
missing versus present-empty values.

Population record actions now own a local search context, with retirement through
render, drafts/selections, paging, child navigation, suspend and disposal. It
rechecks currentness after pausing live data. The existing viewer receives the
checked continuation rather than a new search or page-wide source manager. The
implementation agent reports 27 focused checks passing without skips.

These are source/focused-check findings, not claims that the final native journey
or accessibility/viewport checks have passed.

## Actual-frame review and explicit Close return

The reviewed effective-policy frame comes through Activity and the database-client
service view. Its original service, client-kind and database-system filters remain
in condition 1 while the added error condition can match another service. The
result correctly reports 15 full matches and 15 listed under the existing limit
20; it does not invent a new UI limit control.

The narrow empty-scope frame names Empty environment label, retains fixed UTC
scope, two conditions and the two observed matches without horizontal overflow.
The unreported witness frame has a distinct Unreported environment label and the
common missing-value renderer. The checked-trace frame retains that policy and
original fixed bounds, clearly says fresh read/coverage unknown, and shows three
spans across API, database client and worker plus one separately based linked log.
Those spans are broader than the database-only search seed as intended.

These are naturally task-scrolled frames, not pristine entry or complete narrow
above-fold views. Expanded witnesses remain text-heavy existing presentation;
340 adds no extra panel or permanent policy toolbar.

A measured native usability probe found that explicit Close focused the original
record action but left it offscreen in effective and current-only flows. A first
nearest-scroll correction put the target at the viewport top under the sticky
investigation trail; a center hit-test exposed that remaining occlusion. This is
a real new-entry return defect, rather than a reason to redesign search. The
bounded fix captures current opener/revision, retires the search, and restores
scroll/focus only for a still-visible connected current opener, using a centered
return to avoid the sticky trail. Departed/passive
retirement does not steal focus. Source review accepts that change and its two
mounted current/departed cases. The final native checks confirm all four current
openers return inside the viewport and pass an actual center-point hit-test.

## Final observed acceptance

The final native journey passes in 3.1 minutes (3.2 runner), with the observed
Go/Quint replay passing in 31.021 seconds. Effective named, current-only named,
present-empty and unreported scopes independently return 15, 13, 2 and 1 trace
matches. Each checked continuation observes three spans and one exact linked log
under the original policy/window. Independent raw review confirms no browser
errors and all four Trace Back and explicit Search Close targets both in viewport
and unobstructed at their pointer centers.

The Search Close target top positions are 486.078px, 485.609px, 486.016px and
662.016px respectively, replacing offscreen/under-sticky-trail results. No manual
scroll is used to manufacture those return checks. Trace Back likewise restores
the existing Open trace row action without a replacement search.

The 40-second actor reaches all phases 0–8 with 435 snapshots, four verified
transitions, four source requests and policyPreserved=true, without horizontal
overflow. The actor supplies completion/effect evidence, while the four reviewed
Playwright frames supply visual evidence. The prior scalar 186 regression is a
separate gate; it is not relabeled as part of this acceptance.

## Later depth candidate

After 340, ordinary previous-window mean-duration comparison is the recommended
depth candidate: answer whether latency changed as well as traffic volume. Source
review confirms existing fresh previous-window admission remains count-only, while
view-declaration measures and Go duration aggregates already supply the basic
mean/sum/contributing-count data. This is reuse of an existing measurement path,
not a request for another chart catalogue or query engine. It needs a
measure-aware contract, contributing counts and honest no-data/unknown states,
reusing existing duration aggregates. A concrete journey starts with an ordinary
server-span question, chooses mean duration, explicitly compares the adjacent
previous window, sees aligned human time units and each contributing count, and
can inspect either supporting population. Empty or unavailable operands must stay
no-data/unknown rather than zero latency. Preserve exact typed policies/windows
and independent read clocks, and separate volume changes from mean changes.
Do not average scalar percentiles or silently broaden the count-only comparison
kind. This is a future candidate needing its own contract, not 340 scope.
