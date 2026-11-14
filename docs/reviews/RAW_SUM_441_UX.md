# 441 raw sum observations: UX review

## Product contract

A raw sum read answers what an exact discovered instrument reported in a fixed
window. It does not calculate a rate, identify a reset, establish continuity or
claim source completeness. The same inspection path covers monotonic counters
and nonmonotonic sums. Goo's discovered `stock` classification for cumulative
nonmonotonic sums must not be relabeled as a monotonic counter.

Reported temporality and monotonicity belong near the held observation receipt,
with exact enum values available in detail. Delta, cumulative, unspecified and
unrecognized enum values remain distinct; an unknown enum is not a readiness or
health verdict. The interface must not imply that a reported start/end pair is a
validated interval.

## Shared interaction

Reuse the existing supplied-data table, local Find, bounded paging and attribute
inspector. Read observations is the explicit acquisition gesture; local display,
metadata disclosure and navigation back to a settled bank do not read the source.
The fixed window and source-read clock remain visible and separate from each
reported point timestamp.

Zero, missing scalar data and no-recorded-value flags remain distinct. An omitted
instance label is an honest reported fact, not a reason to invent producer
identity. Unit omissions are explicit, while different units and temporality
profiles cannot silently merge into one read.

Pending navigation retires publication authority without discarding a settled
bank. Late completion must not replace its rows, read clock or focus. Explicit
Read may follow its pending heading; completion must not steal later focus.
Close and Cancel restore an eligible opener. Source refusal retains held rows
without silently broadening scope or rereading.

## Acceptance status

Production source review accepts the shared inspector and explicit sum profile.
The discovery catalog now shows temporality and monotonicity in the visible
binding and accessible Read name, so exact variants can be distinguished before
reading. Metadata uses Unspecified (0) or Unrecognized (n), rather than calling
unknown sum temporality not applicable. Gauge metadata remains separate.

The inspector uses neutral Raw sum observations and Sum observation headings,
with reported temporality and monotonicity in its receipt and detail. It retains
the same local controls, exact clock, unit omission and unavailable-value
presentation as the gauge path. There is no connected timeline or derived rate.
The focused shared-owner module independently passes **22 tests, zero skips**
(130.7 ms), including strict profile admission, stock and unknown metadata,
shared retention and cross-profile late-response retirement.

Independently viewed genuine runtime frames at **1440, 390 and 320 px** under
`docs/journeys/screenshots/441-sum-observations-*.png`. The compact desktop table
and mobile stacked rows remain readable, with visible column labels and reachable
Inspect/Close controls. The reported Cumulative · Monotonic tuple, source-read
clock and fixed window remain visible before the rows. Narrow exact timestamps
wrap rather than causing horizontal overflow. Missing instance identity stays
distinct from a reported label, and flagged raw zero remains unavailable.

The known no-recorded-value diagnostic now displays **Unavailable · not
recorded**, avoiding a requirement to interpret the raw flags. The raw scalar
and typed value state remain intact. Independent final focused tests pass all
**22 tests, zero skips** (335.8 ms). The final replacement frames were independently rechecked at all three widths:
the direct not-recorded label is readable, including its two-line 320 px form,
and the surrounding layout remains within the viewport.

No blocking visual or accessibility-source finding remains. Parent-owned final
browser acceptance records actual keyboard and source/lifecycle results;
this review does not claim a physical screen-reader or cross-engine test.
