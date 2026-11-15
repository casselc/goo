# Independent paired filters acceptance — 2026-11-15

Scoped UX, accessibility and web-standards reviews accepted the shared-component
implementation after the following fixes:

- Applied conditions now remain visible outside closed Log/Span filter disclosures.
  Pending drafts have their own marker, status and Apply/Discard actions.
- Successful Apply captures keyboard focus before disabling its button, then
  reveals the persistent named lane summary. Dirty Resume reveals Apply/Discard.
- The shared editor hides source-loading actions without a registered loader.
  Metadata-only field/constraint navigation still focuses a visible field summary;
  source-backed owners retain their loader actions and existing guards.
- The shared definition library wraps long lane summaries inside their grid cell
  rather than overlap immutable revision labels.

Reviewers inspected source, closed desktop/390/320 frames and open draft frames.
The accessibility review also read real Enter-Apply and recall-focus assertions.
The standards review did not perform an independent screen-reader or cross-engine
run. No remaining concrete blocker was reported within that scope.

Final real OTLP execution checks selective acquisition, protected shared scope,
missing/empty matching, exact saved predicates and paused recall. The actor and
existing paired ownership model pass; fifteen negative controls fail as intended.
[Validation](../../VALIDATION.md) records the measured SQL/request work and limits.

The next consolidation should humanize typed duration thresholds using shared
formatting and retain exact raw values. That is a readability improvement, not
an additional predicate language or a prerequisite for this feature.
