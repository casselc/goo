# Responder flow design review — 293

2026-10-31. Source and documentation-frame review through292. Inspected
`ui/population-workspace.mjs`, `ui/preservation.mjs`, Activity entry ownership,
the current roadmap and289 product checkpoint; visually inspected actual286,
288 and292 journey frames. An older260 mobile trace-inspector frame informs only
historical density observations. This is not a live walkthrough, user study or
performance measurement. Current executed checks belong in [VALIDATION](../../VALIDATION.md).

## Decision

The responder loop is connected: start with Activity, narrow a population, compare
an explicit baseline, inspect a chosen side, then preserve and reopen intent,
presentation or evidence. The next useful improvement is **make Preserve feel
like finishing an investigation instead of opening another editing workspace**.
Put the saved-view Finding attachment behind a native **Add to investigation**
disclosure, retain the direct Save exploration view action, and protect the form's
ownership. This is a bounded usability and correctness improvement, not another
component framework or chart capability.

The actual [288 expanded Preserve](../journeys/screenshots/288-explicit-signal.png)
spends roughly290 vertical pixels on the empty Finding form, attachment command
and explanation. That form appears even when the user only wants a question link
or reusable view. The form's prominence does not reflect its secondary role.
Architecture review also found that an old attachment handler can read a newer
saved ID alongside its old textarea. Correct ownership and lower visual intrusion
should ship together.

## Flow assessment

| Responder intent | Current useful behavior | Design implication |
| --- | --- | --- |
| See what is happening | Stream-first Activity scopes ordinary OTel logs and spans, with explicit stream ownership. | Keep event payload ahead of authoring; opening a saved definition must not start a stream. This review did not newly validate the live entry. |
| Narrow something interesting | Checked scope, filters, record inspection and interval/group intent connect to staged fixed analysis. | Keep chosen signal, environment and window visible. Preserve typed missing/empty and compound predicates; do not replace them with a friendly but inexact label. |
| Compare a baseline | Explicit independent side reads and aligned count groups are available through290–292. | Side identity and population support matter more than repeated provenance prose. Matching selectors do not establish a joined snapshot. |
| Inspect the chosen side |292 stages the exact baseline/current group without reading, retains its parent and requires Run. | The staged child is honest but repeats the same read warning in its banner, status and footer. One concise status can carry this later. |
| Return to comparison | Back renews the intended group/focus actions while retaining the parent observation. | Preserve selective renewal. Older deliberately retired plot handlers must not become active merely because their DOM returns. |
| Preserve an explanation | Question, view and evidence have deliberately different read effects; two-signal questions require an explicit signal. | Keep these three meanings distinct. Make optional investigation attachment secondary, rather than merging everything into Save. |
| Reopen and reuse | Saved definitions open without implicit telemetry acquisition; captures keep distinct observation identity. | Keep the exact saved revision and pending-draft guards. A successful click alone is not evidence that the intended attachment was published. |

The [286 focused log frame](../journeys/screenshots/286-focused-log-composition.png)
shows a secondary density concern: Inspect, Surrounding logs and missing-trace
messages stack within every record row. It does not justify a second row-action
system in this cut. Logs and spans share the responder flow but retain signal-
specific inspection; neither should acquire trace completeness guarantees from
shared presentation.

## Bounded implementation contract

Use a native, initially closed disclosure labeled **Add to investigation** around
the existing Finding field, attachment action and its reason/status. Keep title,
Save exploration view and its exact saved-view link immediately available. Opening
or collapsing the disclosure performs no save, attachment or telemetry read.
Do not automatically open it when a saved ID appears.

Keep the draft finding and disclosure state through a legitimate redraw and held
parent return. Retired or disconnected controls must fail closed: they cannot
attach the latest saved ID using an old textarea, clear a newer draft, or publish
a document change. Retain exact ID/revision, pending-source and pending-document
guards; an editable note is not authorization to silently save or replace a view.
Use the existing application owner, not a second source or navigation lifetime.

Acceptance should cover the actual UI: share/save without a large empty form;
expand, type, collapse and reopen; save/re-render and return from a child with the
finding intact; attach the current exact definition; refuse dirty/pending states;
and invoke a retained old control after replacement without mutation. Verify
keyboard disclosure behavior and compact viewport layout. Preserve the explicit
Question/View/Evidence distinctions and existing independent read assertions.
Current frames/storyboards and Bombadil actions/invariants should reflect the
new disclosure rather than searching for a permanently visible textarea.

## Small follow-ups, not prerequisites

The [292 staged baseline child](../journeys/screenshots/292-baseline-group-staged.png)
repeats its read effect at least three times. Suggested compact banner:
**Baseline group “shared” · staged question, not stored rows.** One adjacent status
can say **Run reads independently; parent paused.** Keep the explicit Run selected
population command and exact scope disclosure. Remove redundant footer reminders
only if the same facts remain visible; do not hide source policy inside a tooltip.
This can follow the attachment fix without changing read or Back lifetimes.

README currently places eight milestone capability callouts before its strong
purpose paragraph. Lead with “Start with what is happening now,” the general OTel
purpose and the narrow→inspect→compare→preserve loop. Consolidate capabilities
into stable task-oriented links; validation history belongs in its existing log.
The purpose should make sense without knowing the sample application's domain.

After this bounded cut, continue the roadmap from observed responder friction.
Compact per-row secondary actions and repeated evidence/provenance prose remain
candidates when their owners are next touched. Do not turn this review into a
whole-UI rewrite, a generic saved-object abstraction, blanket Back reactivation,
or another isolated chart/measure milestone.

## Delivered frame review

The final [293 default Preserve](../journeys/screenshots/293-preserve-default.png)
now keeps title, Save exploration view and the exact saved-view link visible while
showing only the closed **Add to investigation** summary. This materially improves
the hierarchy over288: optional finding authoring no longer dominates the first
Preserve interaction. Question, Evidence and View definition remain distinct.
The implementation uses native `details`/`summary`; its source keeps publication
behind the attachment action and checks mounted/visible ownership. Native semantics
support keyboard disclosure; this visual/source review does not substitute for the
implementation owner's executed keyboard and stale-handler checks.

README now leads with the general-purpose OTel investigation loop and consolidated
workflow links. No design blocker found in that final frame or these source changes.
The `293-review-child-*` frames inspected during this review still showed the old
always-visible Finding form; they are preliminary whole-flow observations, not
proof of the delivered attachment layout. Subsequent final journey frames and
executed results remain the verifier's responsibility.
