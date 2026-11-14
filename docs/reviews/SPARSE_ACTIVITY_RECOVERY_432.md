# Sparse Activity recovery — 432b review

[Contract](../ACTIVITY_SPARSE_RECOVERY.md) ·
[Storyboard](../journeys/sparse-data-recovery.md) ·
[Validation](../../VALIDATION.md).

Architecture review selected the existing fixed record-question executor rather
than a second search API. The wider question ends at the latest completed poll,
preserves the whole applied selector, and uses explicit Run. The held parent is
paused, and Return performs no source read.

The review corrected a key interpretation: SSE batches contain deduplicated
observations. A zero-row batch is not a source census. Recovery requires an empty
retained buffer and a completed unlimited read, and describes loaded observations
rather than claiming source absence. Event-time bounds, source read time and
unknown ingestion completeness remain distinct.

Independent standards/accessibility source review accepts native button/details,
no focus changes during polling, hidden/current-owner guards and the pause-time
lease recheck. Follow-up handling restores the paused parent and opener if the
fixed view refuses or throws. Focused tests exercise the actual application callback.

Independent visual review accepts regenerated desktop and 390/320 entry, fixed
staging, result and held-return frames. A discovered stale “Unapplied scope” label
after undoing an edit is corrected; the clean paused view now hides Apply while
keeping Resume explicit. No remaining blocker was reported for this slice.

Limits: Chromium/source review is not assistive-technology certification. Limited,
stale, hidden and failure states receive focused unit checks, separate from the
real OTLP sparse-event journey. The feature does not diagnose skew, retention or
arrival ordering. Existing duplicated staged explanation prose and property-order
sensitivity in Activity draft identity remain consolidation work in
[433](../ROADMAP_2026_11_14_CHECKPOINT_430.md).
