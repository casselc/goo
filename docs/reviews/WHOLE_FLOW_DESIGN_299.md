# Whole-flow design checkpoint — 299

Review of checkpoint `0550718`: source, current contracts, executed 294/296/298
journey descriptions and the actual Activity, applied-template and paged-List
frames. This is a design/source review, not a new live-browser accessibility audit
or a claim that the proposal below is implemented.

## Decision

Make an existing **timeline + aggregate + records composition reusable as a
service template**. The user outcome is: “I have a useful view of when the activity
changed, how its populations differ and the underlying records; apply that same
view to another service without rebuilding it.” Use the existing authored-template
catalog, fixed timeline and explicit query owner. This is one bounded extension of
an already usable workflow, not a new dashboard builder or chart family.

Architecture independently recommends the same outcome. Preserve the existing
whole-composition contract: an authored timeline requires an equal, positive source
and destination interval width, including a hidden timeline. A different width or
missing destination temporal aspect is an explicit compatibility refusal. Configure
the destination through the existing interval-width staging controls and Run;
never copy the specimen's time window, service, environment, predicates or rows.

## What is already joined up

- Activity is the working entry: narrow logs/spans, pause or inspect a fixed scope,
  retain independent source/evidence clocks, then continue in an investigation.
  The expanded-filter Activity frame illustrates deliberate editing; it does not
  establish that all those controls appear before the default stream.
- 294 exercises the connected Activity → fixed population → captured comparison
  → actual completed trace → saved view/reopen path. Repeating the same path as a
  new feature would not advance the roadmap.
- 296 already delivers HTTP/DB service defaults, native Customize, authored
  templates and destination-group binding through Save → Inspect → Apply → Run
  → reopen. “Make a DB view reusable in another service” is shipped, not a new
  recommendation. Another catalog or onboarding wizard would duplicate its owner.
- 297/298 supply compact List and Table, qualified fields, readable UTC/durations,
  useful selected-measure ordering and one selection/Keep/Inspect/context owner.
  The actual paged List is compact and has one clear action strip per observation.
  Rebuilding those controls or adding another inspector is unnecessary.
- Fixed compositions already contain count/mean/p95 timelines and aggregate plus
  bounded records. Interval gestures stage one shared UTC window. Paging leaves
  the held aggregate and timeline intact. The authored-template gate currently
  requires removing that useful timeline, which breaks reuse of the whole view.

## Why this gap is next

Timing is central to “what changed?” A table/list answers what the observations
say; grouped comparison answers which population differs; the existing timeline
answers when. Requiring users to dismantle the latter before preserving a reusable
view makes the grammar feel like a collection of exceptions. Extending the existing
template to that already implemented role improves composition and general-purpose
observability without introducing more surface area.

A supplied-span waterfall can be useful later, but an arbitrary bounded service
page may contain unrelated traces or missing parents. It must not look like a
complete cross-service trace. Existing trace expansion owns an independent read;
choose its future authoring contract deliberately rather than presenting it as a
records-kind toggle. It is not required to deliver the temporal-template outcome.

## User flow and acceptance

1. Open an ordinary OTel service view, Customize and add a supported fixed timeline
   using the existing controls. Changing width stages the query; Run is explicit.
   Choose captured count, mean or p95 through existing measure support.
2. Name the composition and enable Reusable template without removing the timeline.
   Preserve existing role order, visibility, title, kind, measure, completeness
   requirement and record Table/List fields. Save uses the existing catalog.
3. In a different compatible service/grouping, Inspect the template. Preview names
   the destination grouping and temporal requirements with human labels, for
   example “Timeline: p95 duration · 10-second intervals.” Show the required and
   current interval when they differ. Do not add a second template editor.
4. If the destination has no intervals or a different width, Apply stays unavailable
   with actionable guidance to the existing interval controls. Opening preview
   performs no source read. Configuring a width does not borrow the source window;
   Run explicitly obtains the destination temporal aspect.
5. With compatible width and measure support, Apply keeps the destination's exact
   fixed bounds, selectors and grouping. The timeline, aggregate and records retain
   their existing data/quality semantics. Apply and local display changes read
   nothing; no complete-data witness is serialized or manufactured.
6. Run, page records, stage an interval with pointer/keyboard, save and reopen.
   Paging must keep the held timeline and its basis. Interval staging changes the
   shared root only on explicit application and still requires Run. Reopen stages
   configuration rather than silently querying telemetry.

A real browser journey should use at least two services and different destination
grouping to distinguish template reuse from replaying the specimen. Include native
preview/refusal and recovery, retained Back/current controls, pending record
selection, no-read budgets and a bounded actual screenshot of the temporal view.
Use property/model checks for source/destination interval compatibility, missing
measure support, role preservation, hidden timelines and rejected source authority.
Keep current Table/List and whole-responder journeys as regressions.

## Boundaries and later work

No new expression engine, query source, renderer, template store, live subscription,
comparison-baseline binding, WYSIWYG canvas or production team authorization is
needed. Arbitrary numeric/native timelines and captured-comparison templates remain
separate profiles. Existing unknown coverage, invalid/missing duration cells,
complete-only eligibility, independent read clocks and exact interval bounds stay
visible. Revisit a broader trace-view grammar after this composable temporal reuse
outcome, rather than opening another sequence of guard/copy-only milestones.

## Implementation source review

Reviewed the optional-timeline authored profile and population authoring/preview
integration. No source UX blocker remains. The author retains one Reusable template
and profile control; a timeline adds a short same-width requirement rather than a
second editor. Preview identifies its timeline kind, measure and hidden state and
shows required/current interval widths in human units. Compatibility refuses before
Apply when the destination lacks the required intervals or measure support.

The recovery control reveals and focuses the existing interval-width control.
For a flat destination it reveals and focuses the existing composition action
instead. It must describe revelation accurately, not imply that clicking performs
composition or configures a width. Existing ancestors open to make keyboard focus
visible. This route adds no query, configuration mutation or implicit Run. Source
review requested human timeline measure labels and an accurate flat recovery label;
actual browser behavior and generated frames remain separate validation work.

## Next substantive outcome after temporal reuse

Consider a **reusable captured-count comparison composition with an explicit
destination baseline input**. This would preserve the useful “what changed?” view
while requiring the destination's appropriate capture, rather than copying an
unrelated immutable baseline ID from the specimen. Reuse the existing capture
picker and strict count comparison admission. First scope is ordinary count only;
missing/incompatible capture prevents Apply. Opening the picker/preview does not
read telemetry, and preparation remains explicit.

This is a proposed input-binding contract, not permission to simply remove the
current captured-comparison template refusal. It needs explicit resolution and
preview of the selected destination baseline, unchanged destination population,
independent evidence clocks and no transfer of completeness/source authority.
Existing trace findings already preserve population plus trace evidence, and
service defaults already follow authored reuse; neither should be repackaged as
an unimplemented next feature. A supplied-span waterfall remains another later
shape, with its bounded unrelated-trace limitations clearly distinct from the
independently acquired trace inspector.

## Actual-frame review

Reviewed the final 299 author, incompatible-preview and reopened-timeline frames.
No visual UX blocker was found. Authoring keeps the existing Preserve surface,
one template toggle/profile and short same-width explanation. The refusal preview
names the destination grouping and scope, shows p95 duration in human language,
required 10-second versus current 1-second intervals, a visible Configure
destination intervals action and disabled Apply. It explains that configuration
and Run are explicit. The reopened view visibly retains p95 duration and Complete
intervals only, with UTC bounds, quality context and the existing exact-value path.

The current reopened frame has three positive 9 ms intervals with equal full-height
fills; a suspected missing third bar was not reproduced. Reading pixels at three
interior chart heights confirmed equal-width filled regions for all three bars.
No renderer change was justified by that observation. The application journey's
source/effect accounting, keyboard/focus and actual destination-data checks belong
in VALIDATION.md; screenshot review alone is not a screen-reader audit.
