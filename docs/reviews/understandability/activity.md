# Activity/application ownership: understandability review

Reviewed 2026-11-16 at the post-462 implementation. Scope: `app`, `workspace`,
`stream-entry`, Tail, fixed Exploration, population workspace, filter/query
owners, paired Activity and retained navigation. Display primitives and durable
stores have separate component reviews. This is source review plus focused tests,
not new performance measurements or another UX audit.

## Conclusion

The important boundaries already exist: pure question sessions, environment
scope drafts and chronology reading state; source-owning adapters; passive
displays; explicit staged versus executed work. Preserve these. The source is
still too dense and too concentrated in a few owner closures for a newcomer to
read comfortably in an afternoon. A new state framework is not the answer.

The [walkthrough](../../walkthroughs/activity.md) provides an afternoon route
through representative functions. Whole-file reading currently exceeds that
budget: approximately 230 KB in `population-workspace.mjs`, 70 KB in
`exploration.mjs`, 54 KB in `stream-entry.mjs` and 48 KB in `tail.mjs`. These are
not bundle-size or performance defects by themselves; their significance is the
number of independent responsibilities and lifetimes embedded in long statements.

## Prioritized findings

| Priority | Evidence in current code | Recommended bounded change | Acceptance / risk |
| --- | --- | --- | --- |
| High | [`population-workspace.mjs`](../../../ui/population-workspace.mjs): `mount` starts at line 167 and contains source execution plus `mountTemplateAuthor`, artifact import/export, view attachment, summary/facet/display editors and preservation wiring. Many are nested closures over the same large mutable scope. | Split named owner adapters by lifetime: query execution/result lease; declaration/form editing; preservation/publication. Keep the composition root as an explicit set of callbacks. Start with one form owner, not a wholesale rewrite. | Read one concern without loading all others. Existing recall, pending-write, scope-withdrawal and hidden-action tests remain green. Do not merge query cancellation with publication acknowledgment. |
| High | [`tail.mjs`](../../../ui/tail.mjs): `mount` at line 86 owns stream parser/admission, retries/watchdog, record DOM reconciliation, inspection, selection expiry and retention publication. `connect` at line 164 compresses a full transport state machine into a few very long lines. | Move SSE transport/admission and its lifecycle into one named source owner, leaving the view adapter and the separate selection/preservation owner explicit. Format touched functions as ordinary readable statements first. | Preserve reset/batch/gap admission, retained original-window facts, ABA/epoch guards, terminal versus automatic retry, late-body rejection and selection acknowledgment. Test against real stream faults before replacing the owner. |
| High | [`investigation-trail.mjs`](../../../ui/investigation-trail.mjs): `leave`, `restore`, `select` scan every descendant and capture/replace children of every `goo-*` element to retain a visit. This couples the trail to custom-element internals. | Introduce a narrow visit reading-state contract at component owners; retain focus/scroll deliberately rather than restoring arbitrary custom-element child trees. Begin with one retained record-result owner and characterize existing behavior. | Preserve eight-visit limit, detached view behavior, cached inspector/search state and no read on Return. Current code is tested; the review does **not** establish a broken callback or memory leak. The risk is that future element reconnect/render revisions diverge from restored old DOM. |
| Medium | [`app.mjs`](../../../ui/app.mjs), especially lines 139–220, wires scope recall, Activity/fixed visits, resource returns, native record continuation and Saved through callbacks plus document/global events. [`exploration.mjs`](../../../ui/exploration.mjs) combines many event subscriptions with local source generations. | Keep `app` as composition root, extract the retained-record navigation adapter, and publish a small event/owner table: publisher, payload, consumer, lifetime and whether it may read or write. Prefer direct injected callbacks for local one-to-one handoffs already controlled by `app`; do not invent another global bus. | A reader can locate who initiates a read and who invalidates it without searching every event string. No broad event renaming until all publishers/consumers are enumerated. |
| Medium | [`exploration.mjs`](../../../ui/exploration.mjs): `openQuestion`, `run`, `change`, `suspend` and returned `resume` mix root editor adoption, question session, trace/search mode, population mode and result action leases. | Separate fixed question/session execution from feature-panel attachment. Use `Session` as the visible authority; keep specialized capability gates in named adapters. | Preserve staged versus executed semantics, stale result rejection and basis/question matching. Keep trace, reported-edge and native context restrictions explicit. |
| Medium | `Tail.transportIdentity` and `inspectionQuestion` deliberately reject All-environments spans, while scope matching and editor validation occur in other owners. These are policy boundaries, not interchangeable string comparisons. | Before planned span reach, document one owner/carrier admission matrix and reuse a named scope eligibility check where meanings agree. Do not globally relax `requireNarrow` or substitute a generic identity helper for every comparison. | All-log, named, empty and unreported cases remain distinct; specialized and paired profiles refuse before source reads until explicitly admitted. |
| Low | [`workspace.mjs`](../../../ui/workspace.mjs), lines 26–46, creates `{...protocol,...}` then assigns each protocol function to another local solely to re-export it. `stream-entry.plotWindow` wraps statically imported namespace objects in `|| null` (line 48). | Directly re-export protocol symbols and use the imported namespaces directly. Keep routing exports local. | This is proven redundant indirection, not an unused feature. Protocol/link tests should prove the same exported functions and behavior. Runtime savings are immaterial; the benefit is a smaller mental model. |

## Performance and reliability priorities

Measure the following before promising speed improvements:

- Open, suspend and return through eight fixed visits. Count retained DOM nodes,
  active timers, listeners and descendant traversals. Detached views are bounded
  and source suspension already exists; there is no measured leak finding here.
- Reconcile a full Tail bank with a rich attribute set while changing local Find.
  Keep presentation limits separate from source/population totals.
- Measure active versus suspended sources and a pending publication separately.
  Current generation, AbortController, visibility and DOM eligibility checks
  enforce distinct lifetimes. Removing apparently repetitive guards without an
  owner contract risks resurrecting stale actions.

The recent pure reading-state extraction and immutable parent projection are
useful precedents. They reduce a concrete responsibility and preserve public
input isolation; they should not grow into a universal UI state machine.

## Dead/unnecessary work assessment

No whole Activity feature is proven unused by this review. The redundant
workspace re-export layer and impossible namespace fallbacks above are confirmed
small removal candidates. Historical milestone suffixes in test names do not make
tests obsolete: many contain meaningful source-withdrawal and preservation cases.
Before deleting a feature, establish its inbound UI/route/event/API references
and whether it protects a capability refusal or uncertain-data behavior.

## Checks and limits

Independently ran six focused files:
`tail.test.js`, `stream-entry.test.js`, `question-session.test.js`,
`paired-activity-start-445.test.mjs`, `investigation-trail-visit.test.mjs`,
`chronology-reading462.test.mjs` using Node's module mocks. All pass, zero skips.
No production source or backend behavior was changed. Browser performance,
heap/listener counts and cross-surface removal safety remain proposed checks.
