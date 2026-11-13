# Captured-rate comparison visual review — 418

Status: visual acceptance complete for the reviewed equal-rate and zero-baseline
states. No blocking findings remain. Review is independent of the implementing
author; final browser campaign results are reported separately.

Inspected actual 1440, 390 and 320 pixel frames. The shared comparison component
makes the important distinction legible: 120 records over 60 seconds and 240
records over 120 seconds both produce 2 records/s, with equal bars on a common
scale and a zero absolute change. Counts and exposure remain visible directly
under each value. The presentation does not require interpreting prose to
recognize equal rates.

Mobile retains adjacent baseline/selected values and moves change below them;
labels, denominator text and keyboard disclosure focus remain contained. Desktop
uses the same shared component and styling. Upstream coverage remains explicitly
unknown, rather than presenting captured rate as total application throughput.

Source inspection confirms native table headings, explicit cell/header
associations, decorative geometry hidden from assistive technology, and text
values independent of color. Mobile duplicate side labels are hidden from the
accessibility tree; native calculation and scope disclosures remain available.
This is a bounded source/Chromium visual review, not screen-reader or cross-engine
certification.

No blocking findings in the inspected frames. Two follow-up polish opportunities
are nonblocking: contextual calculation-disclosure names would be clearer than
repeated “derived”, and the workspace identity-join explanation could use simpler
operator language with technical detail in its disclosure. Neither calls for a
new component or changes to comparison semantics.

The additional zero-baseline 390 pixel frame correctly shows 0 versus 2 records/s,
an empty baseline bar, visible count/exposure and an absolute change of 2 without
a bogus relative percentage. Initial review found misleading overall “No data”
chrome despite 240 selected records. The regenerated frame resolves this finding:
“Frozen baseline has no records” attributes emptiness to the correct operand,
while the comparison frame says “Observed”. The visible zero and selected rate
remain intact. This corrected state is accepted.
