# One checked-condition flow across investigations

436.6 closes the finite CEL integration sequence with real cross-surface checks
and three shared correctness fixes. It adds no evaluator, artifact family or
native metric language. Ordinary OTel records remain the shared investigation
input; explicit source reads and immutable evidence remain separate operations.

## Accepted paths

| Path | Actual boundary exercised |
| --- | --- |
| Activity → fixed question → inspect → Return | Additional conditions retain their original parent; Apply opens a fresh Activity window, fixed Run uses checked bounds, local checks and held Return do not read telemetry |
| Fixed question → Save view → YAML → original recall | Exact source, base and compiler binding survive; formatting has a distinct definition identity; private invocation tokens are not restored |
| Saved display → another destination | Presentation reuse keeps the destination question and provenance; adopting the donor question is explicit |
| Measurements counter → contextual logs → CEL → Back | Exact producer, namespace, effective environment and nanosecond interval remain under the additional condition; native data/selection/focus are retained |
| Services histogram → contextual logs → CEL → Back | The same record editor works inside the existing service return flow; choosing a nondefault instrument remains explicit |

The new Services/native fixture has **115 qualifying source records**, **112 CEL
matches** and a **100-row visible page**. Independent source assertions and the
native gate distinguish these quantities. Record context means records from the
selected producer and interval, not inferred membership in the exported metric.

Eight new responsive context frames and the executable sequence are in the
[storyboard](journeys/cel-context-entry.md). The existing
[preservation journey](journeys/cel-preservation.md) now also exercises a native
same-ID saved link, Keep, and explicit Discard—not just a different revision ID.

## Corrections found during acceptance

**Numeric severity controls common record presentation.** A producer can report
SeverityNumber 9 with SeverityText `FATAL`. The shared record label now reads
`Info · producer: FATAL` in a neutral tone. Numeric Error with producer text INFO
still has Error tone. Missing/invalid numeric values never inherit an alarm color
from free text. Exact recognized span status remains separate. Raw records are
unchanged, and loaded Find remains a literal search over visible canonical bands
and producer text; it is not a substitute for a numeric source filter.

**Explicit same-view recall overrides deduplication.** A normal repeated history
notification remains a no-op. After a dirty same-ID recall prompt, explicit
Discard reloads the original definition and retires pending/old callbacks. Keep
retains the current draft. It cannot silently leave the edited source in place.

**Source cannot disappear from an older held question.** A newly checked condition
may describe a different question from the still-visible held result. Preserving
that older question must not interpret a source mismatch as intentional source
discard. The owning adapter now refuses publication until the checked question is
run or source is deliberately discarded.

These fixes live in the shared record presenter and existing recall/publication
owners. They do not introduce per-entry-point copies of the editor or controls.

## Observed acceptance, not a screenshot-only claim

The browser journey requires the native Go/Quint gate after writing its completed
ledger and closing the owned runtime. That gate reconstructs the selected native
context and CEL plan, checks complete source membership and rendered severity,
and replays observed read/local/return steps through the existing model. Ten
negative controls reject semantic, origin, custody, source/page and authority
mutations. The browser gate is mandatory and rejects skipped native verdicts;
a standalone native invocation without a recorded artifact is an integration skip.

The existing Hegel semantic/preservation properties, real read-only ClickHouse
parity and private-owner Quint replay complement the new browser/model boundary.
[Architecture](reviews/CEL_WHOLE_PRODUCT_436_6_ARCH.md) and
[UX/web standards](reviews/CEL_WHOLE_PRODUCT_436_6_UX.md) independently review source,
focused checks and actual frames. Exact executed commands, timings and limits are
recorded in [validation](../VALIDATION.md). Physical assistive-technology and
cross-engine acceptance remain unclaimed.

## Next operator cycle

Resume the existing **427 → 428 → 429** contracts: exact saved-view recall from
Recent / Pinned, paired evidence from the existing population Preserve surface,
then measured compactness and shared scope/preservation consolidation. The
[reviewed roadmap](reviews/ROADMAP_AFTER_CEL_436_PROPOSAL.md) carries forward exact
source ownership and dirty-recall guards. The older contracts are not retroactively
marked delivered by CEL work. More language vocabulary is not the next priority.
