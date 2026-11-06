# Retained-parent ownership consolidation — 342

Status: implemented source review accepted. The inventory below records the 341
(`bedcc12`) starting point; 342 repairs and review follow it. Actual native/model
results are in VALIDATION.md and the executable storyboard.

A population child pauses its retained parent and increments the parent generation.
Back must restore usable definition controls without renewing old source requests
or measurement intents. The retained query, drafts and read receipts are the
content; callbacks, request serials and connected controls are execution owners.
They have different lifetimes. This cut consolidates existing owners in
`ui/population-workspace.mjs`; it adds no universal session layer, query engine,
public grammar kind, source acquisition or compatibility shim.

## 341 starting inventory

| Owner | Authority and departure | 341 return treatment |
| --- | --- | --- |
| Selection commands | Immutable generation, draft, selection and question keys; active/visible/connected guard | Rebuild only for selection/facet continuation |
| Fixed UTC editor | Same query and baseline identity, visibility and pending guards | 341 renews mutable window-generation lease |
| Customize records | Ordinary same-question local configuration | 341 renews mutable compose-generation lease |
| Facet/Add breakdown editors | Generation and exact question object; checked draft strings and error state | 341 remounts with prior disclosure and stored draft |
| Template/span measures/count editor | Generation and question; count also held-result identity | Explicit existing remounts |
| Records view kind | Generation, same question and connected selector | 341 remounts even with record mount present |
| Record page and value plots | Exact page/request serial, result identity and generation | Dispose old controller/plot and remount over held page |
| Aggregate/timeline adapters | Snapshot-bound action contexts and generation; suspend on departure | Remount authored/timeline views only for selection/facet continuation |
| Group-side supporting-record action | Snapshot identity, comparison reference, generation and visibility | Explicit retirement/rebinding |
| Temporary selected/rest promotion | Generation, selectedComparison identity and connected button | No explicit return renewal found in 341 |
| Saved display/template inspection | Generation, revision, library generation, serial, full question and baseline key | 341 retires pending serial on child departure; delayed regression needed |
| Artifact import/export and attachment | Dynamic visit context plus request/input revisions | Existing pending retirement/context checks; preserve operation boundary |
| Live stream/comparison | Independent live generation and explicit paused continuity | Remains paused; Back must not subscribe/restart |

## Concrete findings and bounded response

1. `renderSelectedComparison` captures a generation for **Include comparison in
   view**, but `closeChild` does not renew or remount that button. A supporting
   record/value child can return to an otherwise valid temporary rest comparison
   with a visible promotion command bound to its old generation. Add a regression
   around this concrete path and renew its local definition owner explicitly.
2. The 341 UTC and Customize records repairs mutate a lease captured by the old
   callback. That makes the current connected control usable, but also restores
   authority to a saved predeparture callback sharing the lease. If old callbacks
   must remain retired, rebind a fresh immutable owner/handler on renewal rather
   than changing authority inside the old closure. A test must distinguish a
   newly dispatched current click from invocation of the saved old handler.
3. Return behavior is partly selected by opener flags: selection commands,
   authored aggregates and timelines renew for selection/facet continuation, while
   other child classes share the same generation departure. Inventory the actual
   supported actions after record/value/customization returns rather than assume
   these flags imply every owner is current. Source actions may intentionally stay
   paused; local settings must not silently no-op because of the opener class.
4. Count preparation has a separate controller/serial. It sets `countPending`
   then renders while `busy` is false. Filter-editor grouping/value/open-edited
   controls and `openChild` gate `busy` but not `countPending`, so an ordinary
   captured-baseline preparation can admit departure. The primary controller abort
   does not abort this count controller. Its late completion fails generation
   ownership but can leave the retained Reading status or pending controls stale
   until settlement. Retire pending comparison acquisition explicitly on departure,
   without deleting an already-ready held pair. Page acquisition needs equivalent
   schedule inspection where departure is admitted; do not copy unrelated
   route-level suspension behavior wholesale.

Do not solve this by incrementing/decrementing generation back to an old value,
loosening guards, updating every old binding, or rerunning the parent. A full
render also needs deliberate held-adapter handling: same-data reuse can retain a
suspended adapter. Nonauthored Distributions adapters keep baseline aggregates,
baseline-read clocks and last comparison internally; presentation() exposes only
the baseline question. Recreating them from presentation alone would lose held
evidence or require an unauthorized reread. Keep those retained owners paused.
Authored/timeline adapters can acquire fresh local action contexts over their
existing held snapshot, preserving snapshot identity and clocks. Prefer a clear, local list of definition-owner renewal steps
and independent retirement of asynchronous source/metadata work. Preserve existing
specialized owner checks rather than add a second manager.

## Acceptance and executable boundaries

- Exercise ordinary selection Open/Back twice, a comparison supporting-record
  return, a record/value child and Customize. Include a temporary selected/rest
  comparison so local promotion after Back is covered.
- Retain exact query definitions, independent source clocks/bases, held population
  and record data, selection content/history and checked draft strings/errors.
  Invalid UTC/facet drafts survive; Discard works; disclosure and sensible focus
  survive renewal. A remount may replace DOM identity, not semantic content.
- Fresh current handlers work after return. Saved retired handlers, stale
  snapshot intents, late requests and ABA revisit responses cannot publish or
  stage. Explicit Run remains the only renewed source acquisition.
- Hold saved-display GET/template expansion across departure and return, including
  a repeated inspection. The old response neither publishes nor strands a pending
  flag; the next genuine inspection remains usable. Count/page pending schedules
  need equivalent coverage where departure is actually admitted.
- Back and local settings produce zero telemetry reads, comparison evaluations,
  captures or saves. A user-commanded metadata inspection remains separately
  counted; do not count `/api/compare` as telemetry IO or infer clock identity from
  equal counts.
- Reuse `population_drilldown` for observed depth/source counters. Independent
  semantic assertions verify query/draft/receipt preservation and old-callback
  effects; model conformance alone does not establish those facts. Negative
  controls change a Back read, old callback effect or receipt/draft identity.

Generated owner schedules and mounted tests should precede a representative real
OTel native journey and actual frame/focus review. Keep heavyweight execution
serialized. The high-cardinality facet performance measurement remains a later
conditional consolidation task, not a competing implementation queue.

## Implementation source review (before native acceptance)

The implementation now centralizes renewal in the local
`renewRetainedParentDefinitions` operation. Window/Customize event wrappers
capture immutable owner tokens; retirement clears the live token, and renewal
creates a new token and replaces event properties. An old wrapper still compares
its original token and therefore cannot regain authority through a later renewal.
The live definition predicate remains necessary but does not replace that fence.
Pending comparison retirement is conditional, preserving already-ready pairs;
page-request serials advance before abort so late pages cannot adopt.

Existing editors retain raw control state across remount, including exact closed
root disclosure state and checked facet/breakdown diagnostic text. Capture picker
state is captured before aggregate remount and restored afterward. New handlers
and current disabled-state authority remain new-owner responsibilities.

The retained nonauthored Distributions owner cannot simply stay suspended if its
ordinary group actions are expected to work after Back. Its local
`resumeHeldActions` path rebuilds only the distribution action subtree from the
same held result; it does not invoke full render, compare or source acquisition.
A frozen distribution-generation counter fences old include/exclude/trace
handlers even if old DOM is reattached. Private baseline aggregates, baseline-read
clock, last comparison and Compare DOM remain in the existing owner. This is
source review, not a substitute for the native held-baseline/no-HTTP regression.
Specialized pattern/histogram restoration and Compare-specific action callbacks
must not acquire broad renewal guarantees from this ordinary-subtree change.

### Frozen-source acceptance

Final source review accepts the bounded implementation: the Distributions renewal
is restricted to connected ordinary facet compositions (`overview: false`),
retained ready comparison identities survive, and only pending count work is
retired. Old page responses are fenced by their advanced serial. Local editor
state is captured before authored-adapter reconstruction and restored afterward;
closed disclosure state and checked diagnostics are retained explicitly.

The distinction is deliberate: an adapter-level live-definition handler may remain
in its retained owner, while a retired snapshot-bound child handler cannot regain
its old execution token. This cut does not require a universal callback registry
or recreate the whole adapter. Fresh group-action usability, old-handler refusal,
held baseline/clock/Compare preservation and zero-HTTP Back still require the
scheduled native and observed-model acceptance. Source whitespace validation
(`git diff --check`) passes; no runtime check is claimed by this review.
