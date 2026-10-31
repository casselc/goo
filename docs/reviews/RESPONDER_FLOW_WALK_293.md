# Responder flow walk 293

Executed 2026-10-31 at approximately 19:30 UTC against the current local Go build, Chromium 1280×1000. This is a bounded expert walkthrough, not a user study or performance benchmark. No production files or user saved data changed.

## Finding: defer the saved-view attachment form

Opening **Preserve** in the staged selected-population child immediately displays a large **Finding** textarea and a disabled **Add saved view to investigation** button. No saved definition exists yet, so neither control can complete its purpose. The instruction below says “Save the current view before adding its definition.”

The actual child grows from 334px to 956px when Preserve opens: 622px of additional form and explanation. The attachment form occupies a substantial part of this expansion. After running and saving, the open Preserve state puts the first record at viewport y≈1132, below a 1000px viewport. The pre-save screenshot makes the unnecessary textarea especially clear.

Bounded recommendation: keep the existing Preserve structure and save action, but defer the finding/attachment form until a view has been saved; expose it through an explicit secondary disclosure after save. Preserve the distinction between saved configuration and independently captured evidence. This needs no new navigation or feature menu.

- [Staged child, Preserve closed](../journeys/screenshots/293-review-child-staged.png)
- [Staged child, Preserve open before save](../journeys/screenshots/293-review-child-preserve-before-save.png)
- [Applied child](../journeys/screenshots/293-review-child-applied.png)
- [Preserve after saving](../journeys/screenshots/293-review-child-preserve-saved.png)

## Executed path and observations

A fresh isolated runtime on port 18088 used a temporary capture store. Fresh OTLP spans and logs used service `api`, a unique environment, `run.id`, and `cohort`; no reservation fixture. Baseline had two `shared` spans and the current population three. The comparison parent was prepared through the isolated API, then exercised in the browser.

1. Open Activity; enter the fresh environment; select spans and apply; pause. First observed row at y≈252. Switch to logs (this starts the new read automatically), pause; two fresh logs shown. Both entry frames remain compact.
2. Open the prepared comparison definition, run explicitly, narrow its Groups display with Find `shared`, and choose **Explore baseline group: "shared"**.
3. Confirm the child is staged and inspect Preserve before running. It repeats independent-read qualifications in the outer banner, staged status, and footer. These statements are truthful, but the attachment form is the clearest actionable source of unnecessary bulk.
4. Run selected population. Its table shows the expected two baseline spans, UTC timestamps, human-readable millisecond durations, and trace context actions.
5. Open Expand trace; the trace dialog opens. Its frame was captured during “Reading trace context…” and does **not** establish a settled trace result.
6. Save a titled view from Preserve and reopen it. Reopened configuration is staged with explicit Run; reopening does not initiate a telemetry read.

The successful instrumented pass took 7.9 seconds including local cached build/runtime setup; this is a harness timing, not a product latency result. Browser recorded eight POSTs: two tail reads, parent explore/compare reads, child explore reads, and one saved-view write. No page errors. Review setup writes were confined to the temporary store. Two earlier harness attempts stopped on incorrect assumptions about Start/Apply labels and automatic signal switching; these were harness corrections, not app failures.

## Limits and retained material

This did not execute the natural Activity-to-fixed-question narrowing transition: the comparison definition was seeded to focus the review on Groups→child→Preserve. A separate record-details inspector was not exercised. Trace completion, keyboard/screen-reader operation, mobile layout, and load behavior were not assessed. No Bombadil/model campaign or full regression suite was run for this read-only review.

Compact measurement/request data is in `artifacts/flow-walk-293/walk.json` (local, uncommitted); actual frames above are the review evidence. Temporary capture-store copies and intermediate frames were removed after inspection. No request to retain a long-lived capture archive follows from this review.
