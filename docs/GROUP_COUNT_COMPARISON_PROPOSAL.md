# Aligned group-count comparison — proposal history

> Implemented in291. The [current contract](GROUP_COUNT_COMPARISON.md) and
> [validation](../VALIDATION.md) record delivered behavior. The reviewed proposal
> below is retained as design history.

**Status: proposed, unimplemented.** This contract extends the delivered
[captured-count comparison](CAPTURED_COUNT_COMPARISON.md). It answers “which groups
account for the count difference?” over the already prepared pair. It introduces
no additional telemetry/storage read, baseline reference, receipt or causal claim.
Implementation follows source/product/architecture agreement on this proposal.

## First profile and existing boundary

Keep the exact290 eligibility and root compatibility: one ordinary grouping,
checked signal/environment, identical definition apart from windows and delivery
paging, and no numeric/status/reported-edge/specialized bucket profiles. Reuse
`CompatibleCapturedCountQueries`. Both original cohorts must pass existing count,
unique-group and measure validation before constructing the new payload.

The existing outer `/compare` rows are **not** the new input. They are ranked by
exact rational share difference (only its displayed text is rounded) and are empty for zero or diagnostic populations. Build
the count union directly from the validated original cohort groups. Do not alter
existing share comparison behavior or reinterpret percentage-point differences.

Add optional `groupComparison` to the existing pure comparison response. Unsupported
comparison families omit it; an eligible declared group view must show explicit
unavailability when no compatible profile is returned, not fall back to an older
pair or infer new arithmetic in JavaScript.

## Canonical wire profile

A separate closed canonical envelope, analogous to the histogram-comparison
profile, avoids changing summary `goo.comparison/1` or the generic ranked dataset:

```text
GroupCountComparison {
  version: "goo.group-count-comparison/1"
  baseline: Dataset(summary)
  selection: Dataset(summary)
  field: Field
  state: observed | no_data | diagnostic
  reason?: bounded string
  baselineAccounting: coherent | diagnostic
  selectionAccounting: coherent | diagnostic
  rows: GroupCountRow[0..2000]
}
GroupCountRow {
  key: string
  value: string | null
  baseline: GroupCountCell
  selection: GroupCountCell
  change: ViewChange(count, unit=records)
}
GroupCountCell {
  presence: present | absent | unknown
  state: derived | unavailable
  value: canonical unsigned count string | null
  reason?: bounded string
}
```

`baseline` and `selection` reuse the **same** independently hashed held summary
Datasets as the simultaneous summary comparison. Their IDs, selectors, revisions,
source bases and clocks must agree exactly. They retain `origin: browser-held`,
unknown extent/upstream sampling and no completeness witness. There is no shared
scope, source clock or newly manufactured acquisition guarantee.

`field` equals the single grouping field on both checked selectors. The profile
is count-only even if the compatible parent declares a duration summary. No new
per-group duration decoding or ranked Dataset conversion is necessary.

Each input admits at most1000 groups; the entire distinct union admits at most2000.
Category strings retain the existing4096 UTF-8-byte cohort bound. A nullable group
means the existing null/missing category, distinct from the present empty string
and every literal string. Keys are exactly `"0"` for null and `"1" + value` for
strings, with keys bounded to4097 UTF-8 bytes; frontend validation recomputes this mapping and separately rejects
duplicate values/keys. Never use display labels as identity or coalesce long labels.
Do not assert semantic row ordering: Go byte ordering and browser UTF-16 ordering
can differ. Presentation may sort locally with deterministic tie handling.

Counts are canonical uint64 strings, including zero. Differences are exact signed
integer strings in `[-18446744073709551615,18446744073709551615]`, not int64 or
floating point. Go `big.Int` creates differences; browser `BigInt` validates them.
The existing request byte budget remains unchanged; profile bounds do not widen it.

## Accounting and missingness

“Coherent” means **internally validated supplied partition accounting**. It does
not mean authenticated or exhaustive upstream telemetry. The pure endpoint can
verify uniqueness, positive reported group counts and sum(groups)=captured count;
it cannot prove a caller did not omit source telemetry. Never mint a source
completeness witness from that sum.

| Side accounting | Category membership | Cell |
| --- | --- | --- |
| coherent | present in supplied groups | present / derived / positive captured count |
| coherent | absent from supplied groups | absent / derived / `"0"` |
| diagnostic | present in supplied groups | present / unavailable / null + reason |
| diagnostic | not observed in supplied groups | unknown / unavailable / null + reason |

A positive group was actually reported; zero on the opposite side is arithmetic
completion of a coherent supplied partition. It is not another observed row and
not proof that a service never existed. At least one side must have `presence:
present` for every union row. A category row whose value is null is independent of
an operand cell whose presence is absent: the former is a real group identity,
the latter says that identity was not reported in that operand.

A source side marked `changed_between_reads` has diagnostic accounting. For this
first profile all of that side's cells use unavailable/null, even if its raw group
observation exists. The owner still holds original diagnostic groups for existing
inspection; the canonical count profile does not duplicate a second ambiguous
recorded-count channel. If either side is diagnostic, **all changes are unavailable
and all paired bar geometry is suppressed**. Keep the other side's independent
coherent counts, but do not manufacture a comparable pair or absent zero for the
diagnostic side. Unavailable cells require null and an explicit reason.

A legitimate empty side has count0, groups[] and coherent accounting. Its missing
union cells are known arithmetic zeros. The envelope is `no_data`, yet its exact
count differences remain derived. Both sides empty produce rows[] and a known
summary change0, not invented empty-group rows.

For coherent pairs, every row change equals selection−baseline, and the sum of all
row changes equals selection total−baseline total. Filtering or visible pagination
must never change either equality, the supplied union or the summary operands.

## Public validation and presentation

The framework-neutral display receives the canonical envelope. It validates both
summary Datasets and profile eligibility, exact field/selector compatibility,
closed fields, union bounds, category/key identity, cell states/presence, canonical
numeric ranges, coherent side sums, and exact row/whole changes. The host additionally binds each side to the requested cohort and its matching
summary-comparison operand; a completely self-consistent swapped pair cannot be
rejected by a generic display without that external context. It rejects changed
keys, duplicate categories, fake absent zeros, missing known groups that break
accounting, diagnostic derived values and contradictory totals. It performs no
source IO and does not create displayed differences itself.

Use a local **Summary / Groups** choice in the existing comparison child, proposed
as optional `display.view: "summary" | "groups"` with summary as the default. The
same prepared response contains both profiles, so switching makes no request.
Persist that presentation choice in the existing definition, not pair evidence.
Optional `display.order: "change" | "label"` defaults to change.

Groups use one compact table with baseline/current exact counts, signed change and
paired bars on one common count domain over the **full union**. No extra Bars/Table
control is necessary. A local Find/order/page can reduce visible rows but cannot
rescale to the visible subset or imply the rest is absent. Default ordering may use
absolute count change with stable typed-category ties. Label null/missing, empty
string and literal strings distinctly. Keep existing unequal-window/raw-count
warning, independent clocks and explicit baseline title. No rates, health colors,
statistical significance or causal-attribution wording.

Local view/filter/order/page controls must not rebuild the parent record table or
discard a pending record selection. They operate within the passive display or a
small application display owner. Parent record paging retains the prepared pair;
Run or baseline changes retire it as in290. Defer group drilldown gestures until
there is a reviewed side-specific action authority contract.

## Delivery gates

Before implementation, reconcile this contract with the
[design review](reviews/GROUP_COUNT_DESIGN_291.md) and architecture review. Then
ship one complete prepared-summary → groups → held page → saved staged reopen
journey. Real fixtures should include shared, one-sided, null and empty categories,
a known-empty operand and diagnostic refusal. Check visible accounting and no
extra requests when switching presentation. Update actual frames and diagrams.

Go tests cover extremes, nonmutation, duplicate/overflow/accounting refusal,
complete unions and zero/diagnostic states. Hegel compares generated source unions
and canonical frontend validation; corruptions must reject false zeros, swapped
sides, deleted rows and incorrect signed differences. Quint tracks the existing
prepared owner and source-free local exploration; Bombadil must independently
reach rendered groups and staged saved reopen, not just click the toggle. Existing
290 summary and289 paging regressions must retain their source/read boundaries.
