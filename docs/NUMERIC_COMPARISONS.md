# Numeric measure comparisons

Implemented and independently reviewed. This extends the existing record comparison flow to
explicit integer attribute measures. It uses existing source aggregates and
passive comparison components; it adds no query language or automatic source read.

## Compatibility and evidence

Both sides must have identical numeric field namespace/key, integer type, unit
and canonical parsing policy. Selectors and fixed windows may differ, and remain
visible as independently read scopes. Histogram presentation may differ without
changing numeric meaning. No type inference, unit conversion or duration/rate
normalization is implied. Sums over different population sizes or window lengths
must be read with their corresponding counts and coverage.

The new `POST /compare` request explicitly opts into numeric semantics with
`numeric: true`. Legacy
count-only requests, including old numeric questions whose comparison payloads
omitted numeric summaries, retain their bytes and count-only behavior. New
numeric requests require complete totals and group summaries: inspection rows
cannot supply missing aggregates. Immutable captures are never rewritten.

## Exact arithmetic and state

Go validates canonical per-side coverage and integer bounds, group-to-total
accounting and extrema. It computes selected minus baseline for sum/min/max as
integers and mean as a reduced exact rational, using valid-value count as the
denominator. Rational values are strings rather than JSON numbers. No decimal
rounding becomes arithmetic authority.

Missing, present-empty and invalid values remain visible in coverage. An absent
group or a side without valid values has unavailable numeric arithmetic; it does
not become a measured zero. A real zero remains valid. Changed source reads
withhold deltas as unknown. Coverage counts retain their separate semantics.

## Product flow

With one breakdown field configured, pin a numeric baseline, change the selected
population and compare, or explicitly
read the previous fixed window. The existing range-selection tray also offers a
numeric comparison against the whole rest, alongside the existing count-only
comparison. Go derives the same selected/rest predicates and one shared breakdown
(first declared field, otherwise service name); neither UI nor chart evaluates
records. Staging remains source-free and Run is explicit.

Four compact rows show mean, sum, minimum and maximum. Each row uses one signed
scale shared by both sides with a visible zero, so a large sum does not hide mean
or extrema. Exact values/deltas accompany the bars, with valid/missing/empty/
invalid counts beside them. Ranking uses exact changes, including negative values;
one-sided groups remain unavailable for arithmetic.

Stored capture comparisons also support ungrouped numeric totals. Live pinning
and previous-window comparison retain the existing one-breakdown requirement;
choosing that field stages configuration without a source read.

Saved comparisons retain both summaries and the exact numeric intent. Reopening
uses stored evidence. Reusable questions preserve the original numeric binding;
opening a definition does not silently pin a new baseline.

```mermaid
flowchart LR
  A[Two explicit fixed questions] --> B[Existing source aggregates]
  B --> C[Full numeric summaries and coverage]
  C --> D[Go compatibility and accounting checks]
  D --> E[Exact integer and rational differences]
  E --> F[Passive signed paired display]
  C --> G[Immutable capture evidence]
  G --> D
  H[Bin selection intent] --> I[Pure selected/rest planner]
  I --> A
```

## Delivered review gates

NC1: Go validation, exact comparison, legacy compatibility and saved evidence;
independent semantic review plus property/source parity checks.

NC2: Existing live/stored comparison owners, signed passive display, baseline and
range continuation; component/protocol tests and independent UI review.

NC3: Ordinary OTel live journey with capture/reopen, explicit source-read counts,
negative/zero/nonvalid and large-integer cases, screenshots, native model replay
and Bombadil; final review, commit and push.

## Storyboard 175 — compare populations without losing numeric meaning

Executable [journey](../journeys/numeric-comparisons.spec.js) emits ordinary OTel
spans and logs with a unique environment, shared and one-sided service groups,
negative/zero/large integer values and missing/empty/invalid inputs. No goo
annotations are needed.

1. Open a fixed numeric question with a service breakdown. Run reads the complete
   population independently of its 100-row inspection page. Pin captures the
   supplied aggregate as an in-memory baseline; filter edits remain source-free.
2. Run the selected population. Mean/sum/min/max show exact selected-minus-baseline
   changes, including a fractional mean delta, with coverage beside the values.
   Changing the ranking is local and keeps exact group arithmetic. One-sided and
   unparsed groups show unavailable arithmetic rather than invented zeroes.
3. Compare previous window explicitly reads the adjacent fixed window using the
   same binding. Both source windows/read details remain visible.
4. Capture the comparison and reopen it without telemetry or comparison requests.
   Mobile metric rows place labelled baseline and comparison values together,
   preserving exact tooltips, shared signed scales and visible zero points.
5. Brush an integer histogram and stage numeric comparison with the rest. Go
   derives the exact same partitions as the count action, retaining the numeric
   binding. Run reads the two populations; a saved exploration view preserves the
   baseline and ranking but requires an explicit Run when reopened.
6. Direct actual-source cases cover logs and uint64 extremes; strict compatibility
   refusals, empty/all-invalid and changed-read cases verify unavailable/unknown
   behavior. A separate case delays a real comparison response, edits the draft
   and confirms that the old response cannot revive it.

![Signed numeric populations and exact deltas](journeys/screenshots/175a-signed-numeric-populations.png)

![Frozen comparison with both sides visible on mobile](journeys/screenshots/175b-mobile-frozen-numeric-comparison.png)

![Numeric range compared with the whole remainder](journeys/screenshots/175c-numeric-range-versus-rest.png)

Validation: NC1 `8c67196`, NC2 `dcb62e4`; real journey/model/Bombadil and final
mobile evidence are recorded in [VALIDATION.md](../VALIDATION.md). Legacy frozen
comparisons keep their original count-only result rather than acquiring numeric
arithmetic on reopen.

[Aligned distribution comparison](DISTRIBUTION_COMPARISONS.md) adds explicit
paired histograms over the same bindings, with all-record population shares and
side-specific range inspection. Older numeric comparisons keep their original
meaning unless distribution intent is explicitly requested.
