# From Activity to a reusable investigation view — 294

Executable: [294-responder-flow.spec.js](294-responder-flow.spec.js).

This journey starts at generic Goo's ordinary Activity entry. Its fixture sends
standard OTLP spans to the collector; it does not create saved views, comparisons,
captures or investigation documents through setup APIs. Setup polling checks only
ingestion readiness and is separate from the browser request ledger.

1. Choose the fixture environment and Spans, then Apply scope. Narrow to service
   `api` using the filter editor, apply and pause. Open the observed `http.route`
   field and stage grouping. [Activity](../docs/journeys/screenshots/294-activity.png).
2. Run the grouped fixed question and choose Explore linked population. This reads
   the two signal questions in the same environment, predicate and fixed window;
   the planner also requests duration/status/timeline summaries.
3. Choose Customize spans. It stages a one-query composition without a source
   read, preserving the full planned question at offset zero. Run explicitly.
   Back retains the parent. The old Customize handle refuses; its fresh replacement
   can stage another child, again without a read.
4. Open the applied population's **Measures · spans**. Explicitly turn off Span
   status and errors to choose an ordinary count-comparison profile. This stages
   the same query without that optional measure; it does not read or capture.
   The old checkbox refuses after redraw. Run explicitly, then Preserve an
   independent baseline capture.
   [Measure controls](../docs/journeys/screenshots/294-explicit-measures.png).
5. After later-arriving records enter that same fixed window, Run the current
   population independently. Configure the selected capture through the existing
   contextual capture catalog (295 entry; originally the Composition catalog).
   Baseline has two spans, current has three. Turning status
   back on while this comparison is configured refuses without discarding the
   comparison or reading telemetry.
   The computed category text transform must preserve the literal route case.
   [Compared groups](../docs/journeys/screenshots/294-compared-groups.png).
6. Compare Groups, choose the current `/checkout` population, and Run its exact
   staged query. Expand its trace and wait for the completed two-span context
   across `api` and `worker`, including the worker operation outside the service
   filter. Two actual waterfall rows and bars with the exact fixture span IDs must
   be rendered before continuing. Coverage remains explicitly unknown.
   [Completed trace](../docs/journeys/screenshots/294-completed-trace.png).
7. Save a view definition from Preserve. Reopen the returned saved-view link and
   verify that it is staged without reading telemetry. Run independently again.
   [Reopened view](../docs/journeys/screenshots/294-reopened-view.png).

## Observed issues and validation

The initial natural path exposed two missing connections: Customize was restricted
to Recent activity, and the linked planner's additional `spanStatus` profile is
intentionally ineligible for ordinary captured-count comparisons. The implemented
bridge exposes existing Customize on an applied linked population; explicit measure
staging resolves the second issue without silently changing the profile, fixed
window or baseline. A duration-backed display also refuses removal of its required
summary, preserving the display and current results.

The final real unseeded journey passes in 42.5 seconds. Its browser ledger records two tail subscriptions,
11 explicit explore POSTs (including trace spans and linked logs), one independent
capture, one pure compare and one saved-view publication. It performs no document
publication or record-page request. These are browser endpoint operations, not a
physical SQL-query count inside subscriptions or capture execution. Capture
confirmation reads independently; reopening the saved definition is source-free
until the explicit Run. The comparison retains raw captured counts, not rates.

The existing focused-root Hegel test passes 100 generated cases against the actual
JavaScript helper. The original 286 observed Quint/Bombadil regression and 293
attachment regression pass. No second model was added solely to mirror the same
focused-query derivation. Keyboard and screen-reader operation, mobile layout,
production load and multiple human users were not assessed in this journey.

The remaining workflow friction is discoverability: creating a count comparison
requires knowing to open View settings → Composition, and the optional-measure
restriction was formerly explained only after choosing the capture. A contextual
**Compare** entry near the held population is the next substantive opportunity;
it should stage a checked profile explicitly and reuse the same capture picker,
without adding a second comparison implementation.

Follow-up 295 delivers that contextual entry through the
[existing comparison owner](../docs/CAPTURED_COUNT_COMPARISON.md); the 294 frames
remain the original executed journey.

The comparison-discovery friction recorded above was addressed in [295](295-contextual-comparison-storyboard.md). The current executable uses that sole contextual picker; the original 294 run and its recorded request budget remain historical validation.
