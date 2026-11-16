# Related held records: independent UX review

Reviewed 2026-11-16: implementation contract, passive chronology and paired owner,
three accepted 461 screenshots at 1440/390/320px, existing focused tests. This is
source and visual review, not a live assistive-technology audit.

## Acceptance findings

**Changes required before UX acceptance:**

1. **Restore the original inspector even when its row is outside restored Find or
   page.** Return currently reopens inspection only if it finds a visible Inspect
   button with the saved selected key (`held-chronology-view.mjs`, return handler).
   Reproduction: inspect A, change chronology Find/page so A is absent, drill into A
   using its still-open inspector, inspect B or close inspection in the related
   scope, then Return. Find/page restore, but the inspector remains B or closed.
   Preserve original inspection independently of rendered-row availability;
   restore A without changing Find/page, and focus Find when A has no visible row.
   Add a browser case covering a hidden original selected row and a changed
   related inspection, not just a selected row on the restored page.

2. **Make the anchored span identity visible.** The heading shows only TraceId;
   `Same reported span` exposes its anchor SpanId only through `title`. Inspecting
   another related record does not change the scope anchor. Operators can therefore
   reasonably interpret the checkbox as narrowing to the newly inspected span.
   Show a compact explicit anchor span reference next to the control, with the
   full ID accessible on demand. If unavailable, show the reason beside the
   disabled control. Keep this stable as related inspection changes. A checkbox
   label such as `Anchor span <short ID>` is clearer than additional prose.

## Accepted parts

- Local drill-in and explicit Return fit the held-bank mental model. Both source
  lanes, their counts and Paused state stay distinct from the local subset.
- Trace coverage remains unknown; counts retain held-bank denominators.
- Repeated observations and ambiguous reported parents are not collapsed into a
  guessed trace or a healthy/complete result.
- Desktop and narrow frames fit without overlays or clipped Return controls.
  Native disclosures keep most explanation off the main reading path.
- The selected record has a clear visual treatment, and the related scope has a
  distinct heading and native control grouping.

## Nonblocking refinements for consolidation

- Related rows are materially taller than ordinary chronology rows because each
  adds a separate context-disclosure line. Consider an inline compact context
  affordance in the shared record-row composition; retain readable touch targets.
- `Back to lanes` and `Return to chronology` describe two valid levels. A small
  shared breadcrumb could eventually make that hierarchy easier to scan, but
  another navigation control is not needed now.
- The upstream `Pause both` button remains visible while both held sources are
  already Paused. Review its state/label with the broader paired-toolbar pass.

Focused chronology model/view/owner tests independently pass (three test files,
zero skips). Existing coverage does not exercise the hidden-original-inspector
case above. Follow-up acceptance is pending fixes and fresh browser evidence.

## Follow-up source acceptance

The revised implementation addresses both required findings:

- Return emits a bounded `goo-chronology-return` reading intent. The application
  resolves the saved occurrence against the original immutable bank, restores its
  inspector even when no matching row is visible, and uses the visible chronology
  Find as a safe focus fallback. It does not change Find or page to reveal A.
- The scope now shows **Anchor span** and the full reported span ID, or **Anchor
  span ID not reported**. The checkbox describes that visible identity through
  `aria-describedby`. Inspecting another related record does not silently change
  the anchor.

New view/owner cases exercise the absent-original-row return path and the visible
anchor association. The focused three-file model/view/owner suite independently
passes again, zero skips. Source acceptance is granted; final responsive visual
acceptance awaits regenerated screenshots and the amended real-OTLP journey.
The nonblocking consolidation suggestions above remain future work.

## Follow-up responsive review

Fresh 390px and 320px frames are accepted. The full Anchor span ID is visible,
the checkbox/anchor/Return group wraps without clipping, and the new Selected
marker distinguishes selection without relying on background color. The scope
and per-lane held counts remain easy to distinguish from individual records.

The fresh desktop frame confirms readable records/inspector and selection marker,
but its scroll position crops the scope header and anchor above the frame. This
is an evidence-framing issue, not a product blocker; capture the related header in
view for final desktop visual acceptance. Final amended browser receipt remains
with the implementation owner.

## Final responsive acceptance

The reframed 1440px screenshot is now accepted as well: Related held records,
TraceId, visible Anchor span, same-span control, Return action, per-lane counts
and Selected marker are all in view and readable. Combined with the accepted
390px and 320px frames, responsive UX acceptance is complete for this increment.
There are no remaining blocking UX findings. The implementation owner records
final browser receipt and test outcomes separately; the consolidation suggestions
remain nonblocking future work.
