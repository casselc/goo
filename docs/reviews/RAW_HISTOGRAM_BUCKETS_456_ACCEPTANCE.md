# Single-point histogram bucket display acceptance

Round 456 adds depth to the raw metric inspection workflow, using the same
supplied-data grammar as the preceding gauge Points work.

## Architecture and OTel

Independent review accepted the closed `rawHistogram` observation profile and
common Histogram/Table routing. One selected occurrence retains its exact source
fields, read clock, temporality and diagnostics. Evaluated histogram restrictions
remain intact. Bucket arithmetic does not establish collection completeness;
coverage, upstream sampling and global value certainty remain unknown. No source
basis, record membership, percentile or fleet distribution is invented.

The passive model preserves all 10,001 legal buckets, canonical UInt64 counts,
negative bounds, zero counts, and all-range single buckets. Only known-recorded,
structurally valid vectors get bars. Recording uncertainty remains explicit in
Table. No-recorded, absent and invalid vectors retain their source inspection.

Independent UX/accessibility follow-up accepted final runtime `ce700e32…` and
journey `a3e9415202ab` after inspecting fresh 320/390/1440 frames. No remaining
scoped blocker was reported.

## Interaction and visual corrections

Bars/Table controls belong inside the selected observation, not the whole-bank
selector. Native radios retain focus; local pages move focus away from disabled
end controls. Existing Inspect/Close/Nearby context remains the application
owner’s responsibility. The component neither fetches nor emits record selection.

Review caught inherited nonwrapping headers at narrow widths and insufficient
bar contrast. Scoped wrapping and a solid accent boundary over a translucent fill
correct them without hiding exact text. The actual browser measured an **8.69:1**
bar boundary contrast and **9.27:1** count-text contrast over the fill; header text
fits its own cells at 320 and 390 pixels. Contextual heading levels keep the nested
component within the inspector’s document hierarchy.

A second frame review found the shared inspector’s 24rem nested scroller hid
most of a small histogram. The raw histogram owner now keeps this detail in
document flow, and repeated headings/units were removed. The table caption stays
available to assistive technology without adding visible prose. The browser
checks all four sample buckets are visible even at 320px.

These are scoped checks, not a whole-product accessibility certification.
The existing cross-signal journey `82efa4fad1d4` also passes on that same
runtime, including native Quint replay and 15 rejected negative controls.
See [validation](../../VALIDATION.md) and the
[executed storyboard](../journeys/raw-histogram-buckets.md) for final run results.
