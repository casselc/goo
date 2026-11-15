# Paired Activity views acceptance — 2026-11-15

Independent UX/accessibility and source review accepted the paired view workflow
at desktop, 390px and 320px. The implementation reuses Views, the definition
library, record columns, Tail owners and the existing paired transport model.
It adds a saved-definition composition, not a third transport signal.

Review found and confirmed fixes for:

- Interactive recall now focuses Start after closing Views or removing Discard;
  initial page loading does not steal focus. Same-link reopening remains usable.
- Go, JavaScript and OpenAPI reject incompatible layout branches, null layouts,
  wrongly typed signals and unknown/case-mismatched configuration fields.
- Each physical query projection owns a detached filter; edits cannot mutate its
  sibling or stored definition.
- Invalid Unicode-byte-length titles refuse without an uncaught exception or POST.
- A real Bombadil run caught restored layouts resetting on Start. Shared Tail
  preparation now caches the validated detached layout per signal. An actual
  stage/start/prepare regression and subsequent browser/actor runs confirm it.

Final real-source acceptance includes zero-read recall, fresh explicit starts,
independent source refusal/retry, and held fixed-window Return. Bombadil finishes
its own save/reopen/start/inspection path with four tails, one metadata save and
no forbidden source POSTs. Native replay checks the real trace against the
existing model and rejects ten deliberately corrupted controls. Hegel checks
actual store identity, typed policies, lane layouts and mutation isolation.

Saved definitions exclude observation banks, Find, windows, cursors, health and
paired baselines. They do not claim completeness or a shared source-health verdict.
See [executed checks](../../VALIDATION.md) and
[storyboard](../journeys/paired-activity-views.md).
