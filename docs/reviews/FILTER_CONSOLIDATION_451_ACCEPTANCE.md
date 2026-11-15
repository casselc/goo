# Filter consolidation acceptance — 2026-11-15

Scoped UX, accessibility and web-standards reviews accepted the single/paired
consolidation. Each adapter owns its own PredicateDraft instance; source IO and
Environment/Signal stay outside it. Canonical identity does not bypass validation.

- Exact readable duration summaries retain raw predicates and raw-unit inputs.
  Applied conditions, pending drafts and explicit Apply/Discard remain distinct.
- Predicate Discard preserves Environment/Signal intent, resets CEL work, performs
  no source read, and returns focus to a visible named Filters summary. Invalid
  retained predicates stay visible and refuse incompatible signal acquisition.
- Checked, applied CEL predicates may be saved as ordinary ASTs; unchecked edits
  remain blocked. This does not add rolling-view source provenance.
- Review caught a recalled-view bug: Applied scope labels could be missing or
  left over from the previous view. The same passive accepted-selector presenter
  now initializes labels in manual, paired and recalled paths. Real evidence
  checks the original environment remains visible beside an edited environment.
- Closed metadata rendering is deferred; opening still hydrates held suggestions
  locally. No cache or change to source polling was introduced.

Reviewers inspected sources and current single/paired desktop/narrow frames.
The real journey verifies Enter Apply/Discard summary focus, viewport visibility,
source-free recovery and saved ASTs. Standards review did not independently run
assistive technology or another browser engine. No remaining concrete blocker
was reported within this scope.

The measured paired header/first-row positions and control counts stay unchanged.
The [validation record](../../VALIDATION.md) reports exact DOM/source work and
bounded native/property checks without inferring a latency improvement.
