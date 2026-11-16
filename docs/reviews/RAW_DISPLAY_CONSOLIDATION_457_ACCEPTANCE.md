# Raw-display consolidation acceptance

Independent architecture, UX and accessibility review accepted the prepared-bank,
local-state and shared-control changes, followed by fresh narrow-screen frames.

- The prepared bank validates against the original fixed request before freezing
  its cloned data and indexing occurrences. Projections detach their output;
  arbitrary-input convenience APIs still validate the whole supplied bank.
- Histogram paging belongs to the current admitted snapshot. Bars/Table and
  heading changes retain it; new data resets it and retires old callbacks.
- Cached inspection avoids rebuilding discarded content while renewing Nearby
  action ownership. Immediate selected-row updates preserve the original opener.
- A mutable option closure in the new native picker was caught and corrected:
  emitted values now match the radio’s captured label/value after caller mutation.
- Both owners share native fieldset/radio controls. Gauge frames share compact
  chrome; histogram statistics preserve exact values and validity states in three
  compact rows with one producer-presence qualification.

Final 320px frames show all histogram buckets and readable Field/Source value
headers without collision. The gauge view retains the all-held population and
Table Find boundary, duplicate chooser and exact inspection context. No remaining
scoped blocker was reported. This is not a whole-product accessibility claim.

The final gauge/histogram journeys use the same embedded runtime, `d7bf1ea661…`.
They verify local state, native focus and zero additional source reads. Existing
cross-signal journey `35ca67834d77` and native Quint replay also pass, including
15 rejected negative controls. Full
results, property checks and the cross-signal replay are in
[validation](../../VALIDATION.md). The
[consolidation guide](../RAW_OBSERVATION_CONSOLIDATION.md#display-state-and-prepared-banks)
separates synthetic projection measurements from browser/source behavior.
