# Preserve a fixed cohort comparison — design review 308

Status: implemented source and all three final actual browser frames accepted.
Read with [307 cohort design](AUTHORED_COHORT_DESIGN_307.md) and existing portable
view/catalog contracts.

## User outcome

A responder compares an interesting population with the rest of its fixed parent,
then wants to keep that useful pair of questions in the saved/team view. 307 makes
the comparison usable but deliberately saves only the selected view. This cut
makes preserving the query-based comparison an explicit configuration choice.
It does not preserve the historical observations or silently capture evidence.

## One action within the existing comparison

For an eligible temporary selected-versus-rest pair, offer **Include comparison
in view** in the existing comparison panel. Prefer this label to Keep comparison:
Keep already names exact-record custody elsewhere. It adds the checked baseline
question to the existing composition locally, without reading or writing. The
ordinary Preserve → Save exploration view and Export YAML flows then preserve
that configuration. Do not add another save wizard, comparison editor or catalog.

Before inclusion, retain the current note that ordinary Save stores the selected
view only. After inclusion, replace it with a concise fact: **Both fixed questions
are included. Run reads them independently; earlier observations are not saved.**
The action should disappear or become a clear completed state rather than allow
repeated duplicate children. Focus should move to a visible confirmation or the
existing Preserve control. Existing pending/draft/selection constraints apply.

The held-parent path must not offer this action: its baseline is an earlier
observation. A baseline query is not a substitute for that observation. The
existing explicit capture/evidence workflow remains separate, and this feature
must never create a hidden capture or convert held parent into a fresh read.

## Reopen, Run and labels

Opening the saved definition stages both checked fixed questions and their
presentation, without telemetry. Run obtains each operand independently through
the existing owner and shows their actual counts, read clocks and coverage.
Predicates, typed environment, fixed bounds, grouping and page-zero semantics
remain checked. Record display/qualified fields and the selected query remain
the ordinary saved view's own configuration.

Use **Selected vs baseline** or an equivalent generic saved-comparison title unless
the stored model contains and validates the original partition recipe. A manually
authored baseline question does not prove that it is the rest of the selected
population, even if scopes align. Do not preserve a cosmetic Rest of parent title
as if it were a validated relationship. Baseline and Selected columns and concise
UTC context should map operands directly, reusing 307's established orientation.

The original live comparison may legitimately say rest while its temporary
partition witness is valid. After conversion/reopen, any generic query-pair model
must use the honest baseline label. The UI should not imply that two independent
reads establish a simultaneous disjoint partition or telemetry completeness.

## Configuration portability and reuse

Ordinary Save, YAML export/import and CLI validation/import should use the same
comparison declaration and admission. The artifact contains fixed question
intent only, without source tokens, held rows, receipts or store-local capture
IDs. Import creates the ordinary saved catalog copy and does not Run.

Open original question remains useful and honest: it stages the saved pair's
original fixed scopes. **Apply composition** to another population must explicitly
refuse this fixed-pair form until a checked destination-binding transform exists.
Never copy the source baseline into a different current scope, silently omit the
comparison, or call this a freely reusable category template. The refusal should
state that the saved comparison uses fixed questions and offer the existing Open
original question action, rather than introduce another navigation destination.

## Acceptance as a complete task

Use actual authored telemetry to make a selected-versus-rest comparison. Include
it in the view without source/capture/store effects, then save through the existing
Preserve form. Return to the held parent and confirm it remains unchanged.
Reopen the saved definition without a read; verify the staged baseline meaning,
then Run independently and inspect the actual comparable counts/read clocks.

Export the definition, validate/import through the real CLI into a clean store,
and open it there using the same catalog. Verify configuration and record fields
survive while observations are newly acquired only by explicit Run. Attempt
ordinary Apply into another scope and verify a clear refusal without query change
or hidden read. Include held-parent nonavailability and an edited arbitrary
baseline that must never acquire an unproven Rest label.

Review actual compact frames for the inclusion/confirmation state, reopened
comparison with its records and operand clocks, and catalog reuse refusal. Source
and lifecycle checks must separately establish publication/read effects and
completion. Static frames alone cannot establish them.

## Initial implementation and actual frame review

The source implements the local Include comparison in view transition, clears the
temporary observed pair, removes any category-template binding with a visible
message and stages the same records presentation for explicit Run. The saved
form uses a fixed baseline query and generic comparison labels. Held parent does
not offer inclusion. Fixed-pair catalog Apply is refused rather than stripping
the pair or copying its baseline into another population.

The actual inclusion frame clearly states that Run reads both questions
independently and earlier observations are not saved. Preserve remains the
existing collapsed surface. The actual fresh and imported pair frames show the
same baseline groups (101 stable, two missing, two empty) and five selected
canaries on a shared 0–101 scale. Generic Baseline/Current held labels do not
claim a permanently proven complement relationship. The real CLI copy visually
retains that query/presentation meaning.

The final refreshed Groups frames resolve the requested context correction.
Separate visible Baseline · Fixed query and Current held · Selected lines show
UTC millisecond read clocks, with exact original precision retained in titles.
The first run maps baseline read 19:51:41.039Z and selected read 19:51:40.724Z;
the imported run shows its new independent read clocks. Counts remain 105 baseline
(101 stable, two missing, two empty) against five selected canaries, using a
shared 0–101 group scale. No unproven Rest relationship is asserted.

All three final actual frames are accepted; no visual blocker remains. The native
verification owner reports the final frozen-markup workflow green in 1.3 minutes,
including the real saved/CLI import path and refusal controls. Static frames
establish visual interpretation; acquisition/publication effects and completion
remain supported by the separate native/model report. No additional controls or
source changes were requested after this final review.

## Bounded editing limit and next substantive outcome

Disabling root-only Fresh window, grouping/profile and interval edits is an
acceptable admission boundary for a fixed-pair recipe: changing one operand would
otherwise violate its shared checked context. The current explanation directs
advanced authors to edit both questions, rather than silently repairing one.
Selected predicates may still be deliberately edited; generic Baseline labeling
therefore remains necessary.

The next useful outcome should be **rerun this comparison in a new shared
window**, through one existing owner operation that changes both fixed operands
together. A saved question pair tied to an old incident is useful for reproduction,
but a team cannot naturally reuse it to ask whether the same distinction exists
now. Requiring a YAML roundtrip for that ordinary task turns the new saved
comparison into a workflow dead end.

Prefer a bounded shared-window edit/rebase for both questions, with one deliberate
fixed UTC range (or an explicit fresh lookback resolved once), page-zero reset and
staged Run. Preserve each predicate, checked environment, grouping, measures and
record presentation; invalidate both observed operands. Do not independently
resolve now for each side or reintroduce hidden live subscriptions. Keep ordinary
Apply refused until a separate checked scope-binding transform is justified.

This is more valuable than adding another comparison measure or generic operand
editor. It lets a preserved/portable team comparison answer its next real user
question using the same grammar, while keeping held evidence and fresh reads
distinct. Actual task evidence should drive any later service/environment or
category-binding work rather than starting a universal parameter framework.
