# Explicit side-specific group staging design — proposed292

2026-10-31. Read-only interaction review of the delivered291 group comparison,
its actual aligned-group frame, the population child owner and the next roadmap
item. This review ran no browser, model, property test or user study. The controls
and journey below are proposed, not delivered by this document.

## Bounded user outcome

After noticing a changed group, a responder can choose its **baseline** or
**current held** count and stage that side's group question. The parent comparison
stays held and returnable. Staging performs no telemetry, storage or comparison
read; the existing explicit Run action performs a new independent query.

The baseline action explores the captured baseline's question and fixed window.
It must not say that a later query opens the exact captured group records or
recreates the captured count. Late arrivals, retention and ingestion changes can
make the result differ. The original capture remains unchanged.

## Gesture and visual treatment

Keep the existing four-column table and shared bars. In an action-enabled host,
replace only an eligible count's primary text with a subtle native button. The
visible label remains the count, for example `3 records`; the count's own column
supplies side context. Use an always-visible text-action affordance and keyboard
focus, not a click handler on the whole cell or a hover-only indication. Bars
remain display-only and should not become another ambiguous target.

Accessible names should identify both side and group:

- `Explore baseline group: <bounded category preview>`
- `Explore current group: <bounded category preview>`

A tooltip or short common helper can say **Stages a question; Run is required**.
Do not add two large Explore buttons, a new action column or a per-row wizard.
The native count button gives pointer and keyboard users the same direct gesture.
The existing exact-value disclosure remains separate, especially for long values.
Missing, empty-string and literal-string categories retain their distinct labels;
none may become a display-label-derived identity.

The standalone framework-neutral component remains static by default. An explicit
host interaction opt-in may make eligible counts buttons and emit a typed staging
intent. It must not grant source authority, carry an executable child query,
navigate or fetch. The application owner decides whether to honor each intent.

## Eligibility and refusal

Actions are available only for a nondiagnostic coherent pair and a chosen side's
**present, positive, derived** count. An aligned absent zero stays static even
when the other side is positive. Unknown, unavailable and diagnostic cells do not
become links. A known-empty baseline can therefore offer current-side actions
without pretending a baseline group was observed.

The owner must recheck the active pair, exact side, row key/category, side selector
and basis/revision, connected visible mount and current application generation.
Refuse stale/hidden/detached events, pending reads or baseline preparation, dirty
scope/source changes, depth exhaustion and pending selected records. Existing
local Summary/Groups, Find, order and display paging remain source-free and need
not lose their usefulness merely because a scope action is blocked.

Where a host-wide pending operation blocks otherwise eligible buttons, use one
concise status reason rather than repeated row prose. For example, **Clear selected
records before exploring a group**. Preserve the selection; never clear it as a
side effect of clicking a group action. Keep the existing unknown-upstream quality
caption and unequal-window warning intact.

## Retained-child presentation

Use the existing retained population child, with a compact label that explicitly
names **Baseline group** or **Current group** and the typed category. Keep its exact
half-open UTC window and selector inspectable in the existing staged question.
The child's initial state must say that it is staged and that Run reads telemetry
again. For a baseline child, a concise note should distinguish the question from
the stored capture; the parent already retains the baseline evidence link.

The child inherits the chosen query's semantic configuration and appends exactly
one typed group predicate. Null produces a missing predicate, while an empty
string produces equality with empty string. Keep all existing predicates,
including whole OR groups, and reset only delivery offset to zero. Preserve the
chosen side's page limit and requested measures rather than substituting the
latest record inspection page or the other side's window.

Do not inherit the comparison child/reference automatically into the narrowed
child. Its new predicate would differ from the full captured baseline definition.
An ordinary staged record/aggregate question is sufficient; no nested comparison
or automatic baseline preparation is needed for this first drilldown.

The existing **Back to parent population** action remains the return path. It
restores the held comparison, Find, group page/order, current inspection page and
scroll position without rereading. It should focus the newly current counterpart
of the opener where practical, not leave focus on a removed old button.

## Back and action retirement

A retained visual parent is not permission to revive an old source-action lease.
Opening a child must retire the previous group-action binding. Back may establish
a fresh binding and fresh row controls against the same held pair, preserving the
passive component's Find/page/order state. Previously saved element handles or
intents must remain refused after Back. Do not achieve renewed usability by
removing the generation guard from old handlers.

Display configuration commands can retain the existing dynamic-owner treatment;
query-staging intents need their own current action context. Dataset content
revision and application owner generation are different clocks and must not be
conflated. Baseline changes, new Run, navigation, hiding the comparison or leaving
Groups similarly retire the old staging context.

## Proposed storyboard and acceptance

1. Prepare a count comparison and choose Groups. Identify a shared positive row.
2. Activate the baseline count by keyboard. A baseline-group child is staged;
   no read count changes. Inspect its baseline window and exact appended predicate.
3. Run explicitly. The new records are independent evidence, not the stored
   baseline bank. Return to the unchanged held parent without a reread.
4. Activate the freshly bound current-side count. It stages the current window,
   not the baseline window or latest inspection-page offset. Repeat for a missing
   category and a present empty string, verifying distinct predicates.
5. Find/page the groups and inspect another record page. These actions preserve
   the pair and local presentation; staging still binds the original side inputs.
6. Select records and attempt group staging. Refusal preserves the selection.
   Clear it explicitly, then stage successfully.
7. Exercise an old button and old intent after child/Back, rebaseline, hide/show,
   Summary/Groups change and navigation. They add no query or child and do not
   replace a newer context. Fresh visible controls remain usable after Back.
8. Show a known-empty side and a diagnostic pair. Absent/unknown/diagnostic values
   remain noninteractive; quality text does not imply healthy or exhaustive data.

The implementation contract should settle the exact typed intent and host binding
before code. No modal confirmation, new durable artifact, automatic read or broad
query-binding language is warranted by this interaction review.

## Proposal reconciliation

Reviewed [GROUP_SIDE_STAGING_PROPOSAL](../GROUP_SIDE_STAGING_PROPOSAL.md). The
closed `scope-group-side` tuple, property-only host callback/private lease,
positive-present eligibility and exact query derivation agree with this design.
No new row action column or default-side gesture is needed.

The public tuple remains descriptive: a pure planner may derive the same question
from identical held data before or after Back. Retirement applies to the old
runtime callback/lease that carried that tuple. New current controls may produce
an equal tuple under a fresh binding; this does not revive old handlers or make a
serialized event an authenticated action. Verification should distinguish those
two boundaries.

No design blocker is identified in the reviewed proposal. Implement the ordinary
staged child without inheriting the parent comparison reference, and verify fresh
Back controls plus exact state preservation in the real journey. Source and actual
frame review remain future delivery checks, not outcomes claimed here.

## Implementation and actual-frame review

Reviewed the implemented inline count buttons, property-only interaction binding,
retained-child description and Back focus path, plus the verifier's actual
[staged baseline](../journeys/screenshots/292-baseline-group-staged.png) and
[returned parent](../journeys/screenshots/292-return-held-parent.png) frames.
This review did not execute the journey. The delivery owner reports that baseline
and current explicit Runs can return newer counts than their held operands, while
Back restores focus and refuses an old handle; those are execution findings, not
new tests run by this review.

The native count actions add no extra column or tall row controls. Baseline and
current side remain explicit, the shared scale is retained, and the returned
parent keeps Find/order, held comparison and the independently fresh inspection
page. Source inspection confirms that the child is an ordinary staged question
with an explicit baseline/current label, independent-read note and Back action.
The component remains display-only without a host binding. New bindings create
fresh controls; an old button captures its retired binding rather than borrowing
a newer one.

No material product or interaction blocker is identified. One documentation
capture correction was requested: capture the outer selected-population child,
including its Baseline/Back/independent-read attribution, and close Preserve via
its native disclosure after testing the share link. The first inner-region frame
otherwise hides the key attribution and is dominated by a large expanded finding
form. This is a frame-selection/user-state correction, not a request to hide data
or alter the delivered feature.

## Next checkpoint

The Activity → prepared comparison → groups → side question → explicit Run →
held parent → Preserve path is now connected. The next bounded step should be a
whole-flow product/usability checkpoint, rather than defaulting to another measure
or row action. Review the initial stream-to-analysis transitions, where source
reads become explicit, how held/fresh/staged states remain recognizable, and
whether selective Back renewal is understandable alongside older source controls
that still require Run. Inventory repeated scope/evidence prose and duplicated
controls before changing layout; keep necessary independent-read attribution.

Use that review to choose one concrete simplification or missing primitive. Do not
turn this checkpoint into a broad query-binding framework or a cosmetic backlog
that obscures the completed responder journey.


Delivery follow-up: the final executable journey recaptured the outer staged
child after closing Preserve through its native disclosure. The refreshed frame
shows Baseline, Back and the independent-read notice; delivery inspection confirms
the requested documentation-framing correction. Product data and behavior did not
change for this capture. Actual test outcomes remain in VALIDATION.md.
