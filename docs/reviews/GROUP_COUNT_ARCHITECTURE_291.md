# Aligned group-count comparison architecture review — 291

2026-10-31. Read-only source review following the authored captured-count journey.
The proposed contract is [group-count comparison](../GROUP_COUNT_COMPARISON_PROPOSAL.md).
This review ran no browser, source performance or property/model checks. Delivery
checks belong in [VALIDATION](../../VALIDATION.md).

## Architectural direction

A bounded optional canonical group comparison alongside the existing summary
comparison fits the current owner. Reuse the prepared capture/current pair,
existing pure `/compare` operation and count compatibility profile. No new source
read, snapshot-token family, evidence type, query-bearing child or group action
lease is needed. A local Summary/Groups presentation choice belongs to the same
comparison declaration and owner; it must not prepare the baseline again.

The proposed canonical envelope should reuse the same independently scoped held
summary snapshots as `viewComparison`, plus closed aligned group rows. This avoids
widening the generic Dataset string profile: existing cohorts allow group values
up to 4096 bytes, whereas generic view values currently allow 1024 characters.
The optional envelope is a new bounded comparison profile, not another source
Dataset shape or an excuse to weaken existing Dataset validation.

## Existing source boundaries that must be retained

`Cohort.validate` checks one grouping field, distinct category values, positive
canonical uint64 group counts, at most 1000 groups and whole-count accounting.
Accounting may disagree only for `changed_between_reads`. That exception cannot
be treated as permission to infer zeros or calculate a descriptive group delta.

The outer `CohortComparison.Rows` is a percentage-point comparison. It is sorted
by share difference and is empty for `no_data` and `unknown` states. The new count
union must derive from the validated original cohorts, never reinterpret that
legacy row list as raw count change or rely on its ordering.

`CompareCohorts` intentionally supports a broader set of callers than the authored
capture profile, including independently filtered populations, numeric modes and
other comparison families. Keep existing validation and outputs unchanged.
Derive the new optional payload only for its explicitly supported compatible
ordinary count profile. Do not make the whole endpoint call the stricter
`CompatibleCapturedCountQueries` gate and thereby reject existing valid requests.

## Arithmetic and absence contract

- Form the complete union of supplied category identities: at most 2000 rows.
  Null/missing category, present empty string and ordinary strings remain distinct.
  Do not use a visible label such as “missing” as an identity sentinel.
- Coherent observed/no_data sides may use zero for a category absent from that
  validated supplied partition. Presence should distinguish observed membership
  from aligned absence. This is arithmetic over supplied internally coherent
  evidence, not a claim that upstream telemetry is exhaustive.
- For a diagnostic side, observed membership may be `present`, and an unseen key
  must be `unknown`, never `absent`. Comparison values are unavailable/null with
  an explicit source-disagreement reason. Withhold all row deltas/bars for the
  diagnostic pair; do not manufacture a useful-looking difference from the other
  side alone.
- Non-diagnostic row changes are exact selection minus baseline. Use Go big.Int
  and JavaScript BigInt for validation/sorting. The signed result may span
  `[-MaxUint64,+MaxUint64]`; it is not restricted to int64. Geometry may approximate
  display scale, but exact strings stay available.
- For a coherent pair, each side row sum must equal its independently supplied
  summary count, and row-change sum must equal the canonical summary count change.
  Whole zero populations permit derived exact changes. No-data is not health.

Pure `/compare` receives browser-held cohorts. Internal count/group accounting
cannot mint source completeness witnesses. Keep source/caller basis, independent
scope/content revision/read clock and unknown upstream sampling on each side.
Any `accounting: coherent` vocabulary means validated supplied partition only;
it must not become a `completeness` capability or claim signed/exhaustive source
provenance. Capture custody stays in the existing application metadata.

## Admission and mutation controls

The browser profile validator should check closed fields, version and profile,
valid compatible side selectors, bounded rows/categories, unique row keys AND
unique category identities, canonical unsigned counts, bounded signed deltas,
presence/state/value/reason consistency, and exact aggregate accounting. A changed
key must not allow two rows for one category. Null, empty and strings require
negative collision controls.

The new source admission path must validate the optional envelope explicitly.
Currently `importComparisonResult` validates known summary/histogram slots while
spreading other response properties. Merely adding a property to the HTTP result
must not accidentally expose unchecked group JSON to the presenter. Check expected
side questions and align side content identities/clocks/counts with the summary
comparison used by this prepared pair. Freeze or detach imported row arrays and
nested values; passive components must not acquire runtime/source authority.

Negative controls should mutate category identity, duplicate category with a new
key, row count, sign/value of a delta, coherent totals, missing required union
entry, diagnostic absence-as-zero, unavailable non-null values, side scope, side
revision/read clock, overflow and unknown fields. Include exact uint64 extremes.
Malformed optional data must be refused, never silently accepted as a valid
alternate display. Unsupported profiles may omit the optional envelope while
retaining their existing summary or specialized comparison behavior.

## Owner and interface constraints

The count pair remains bound to parent generation, baseline capture ID and held
current result identity. Local Summary/Groups switching, sorting, finding values
and showing exact rows must not call storage, telemetry or the pure comparison
endpoint again. Record paging replaces only the record page; group comparison
uses the same held original groups as its summary. Run/rebaseline/navigation and
baseline drafts retain the 290 retirement rules.

Use paired count bars on one shared nonnegative domain and an exact comparison
table. Sort by exact absolute count change with deterministic category tie-breaks;
never rank raw count change by the old share-difference score. Distinguish “only
in supplied baseline/current partition” from a temporal claim such as new/gone.
This first presentation needs no source filter gesture. Side/group drilldown can
follow only after its exact typed scope/ownership operation is designed.

## Proposal disposition

The summary-side snapshots plus closed aligned-row direction is approved in
principle, conditional on the contract spelling out the constraints above.
Implementation and actual-frame review remain pending. This document does not
claim any delivery tests have passed.

### Final proposal review

The completed proposal matches the architectural constraints: summary-side reuse,
closed raw-key aligned rows, coherent supplied accounting distinct from source
completeness, diagnostic unavailable/null cells and exact signed-uint64-range
changes. Approved for bounded implementation after product/design agreement.

Two precision notes accompany implementation: legacy share rows rank exact
rational differences and only round their displayed values; a raw string key has
one prefix byte beyond the maximum category length. A fully and consistently
swapped pair can be valid supplied data, so runtime admission must preserve the
expected request's baseline/selection role identity and match each side to the
simultaneous summary comparison. Generic display validation proves internal
arithmetic, not which source operand the user intended.

## Implementation source review

Go source review finds no blocker. The new payload attaches only after existing
broad comparison admission succeeds; stricter captured-count compatibility gates
only the optional profile. The constructor reuses the actual summary operand
pointers, forms the full distinct typed-key union, copies categories/count values
and computes exact big.Int changes. Diagnostic sides carry null unavailable
counts and unknown absence, and every diagnostic pair delta is withheld. Existing
caller-held summary metadata remains conservative, with no source witness.

The reviewed tests include the 2000-row union, 4096-byte category, signed uint64
extremes, null/empty/literal identities, input nonmutation and a valid broader
comparison whose new group payload is omitted. Test execution belongs to the
implementation/delivery owners; this reviewer did not run those suites.
Browser admission, rendering and actual-frame review remain pending.

### Existing comparison receipt size boundary

A deterministic local wire-size probe using exported `CompareEvidence` and
`ComparisonReceipt` construction found a material integration boundary. With 400
ordinary shared groups, the previous receipt representation is 58,864 bytes;
embedding the new optional group envelope makes it 152,948 bytes, above the
existing 128 KiB workbench artifact limit. With 300 groups it grows from 45,164 to
116,248 bytes. This is a serialization-size observation, not a source latency or
throughput benchmark.

The new authored profile should therefore remain on the pure cohort comparison
response. Keep the existing stored-evidence comparison algorithm/receipt output
free of this extra envelope until a deliberate stored-group-result contract is
reviewed. Do not silently raise all workbench artifact limits or allow an
otherwise publishable existing comparison to fail solely because it acquired a
redundant new presentation payload. Delivery owns the narrow correction and its
regression test; this reviewer made no production edits.

### Final browser source disposition

No remaining architecture blocker is identified in the frozen browser source.
The group profile enforces the full 2000-row union and each side's 1000 reported
membership bound, nullable category identity and byte limits, uint64 counts,
signed exact changes, coherent sums and diagnostic unavailable values.

Application admission explicitly validates the new optional envelope rather than
letting the old response spread pass unchecked data. It binds each side to the
request's selector, source basis, clock and count; checks full requested category
union/membership/counts; and requires identical summary operands and comparison
state/reason in both canonical profiles. Imported groups and nested data are
copied/frozen conservatively without source-completeness capabilities.

The passive element now retains separate invalid-input errors, clears removed
comparison attributes, and cannot resurrect old supplied data after invalid JSON
or another property change. Correcting an invalid display back to its previous
valid value renders recovery. Standalone title input is bounded. Diagnostic
status remains visible for empty unions and no-match filters. Local presentation
updates operate within the pair/component and retain pending record selection;
no new source operation is introduced.

The measured receipt-size regression is addressed at the existing evidence result
projection: its breakdown omits the transient optional authored group envelope.
The regression test preserves a 400-group receipt's publication/reopen below the
existing bound while pure cohort comparison retains the full new profile. No
artifact budget or historical receipt algorithm is widened. Actual execution
results and final frame review remain delivery-owned.
