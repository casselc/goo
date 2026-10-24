# Fixed-scale decimal attributes

Implemented and independently reviewed, with real OTel browser and native acceptance.
This contract extends the existing
[numeric binding](NUMERIC_ATTRIBUTE_PLANNING.md), not the meaning of ordinary
string attributes or native OTel metric streams.

An explicit `decimal64` binding declares a `scale` from 1 through 9 and a unit:

```yaml
numeric:
  field:
    kind: attribute
    key: example.cost
  type: decimal64
  scale: 2
  unit: USD
  parse: canonical-decimal
numericHistogram:
  binning: powers-of-two
```

The unit is a label, not currency conversion or dimensional validation. No type,
unit or scale is inferred from an OTel convention, key name or observed values.

At scale 2, `12.50`, `0.00` and `-0.01` are valid; `12.5`, `12`, `+12.50`,
`012.50`, `12.500`, `1.25e1`, whitespace and `-0.00` are invalid. A literal must
have exactly the declared number of fractional digits. There is no rounding,
truncation or implicit conversion. Typed predicate thresholds use the same rule.
Missing keys, present empty strings, invalid values and valid values remain
separate coverage partitions. Numeric comparisons select valid values only.

The signed coefficient must fit int64. At scale 2 the inclusive range is
`-92233720368547758.08` through `92233720368547758.07`. Sums use a wider exact
integer accumulator; mean uses the valid count, never all records.

## Exact transport and display

Numeric summary `sum`, `min` and `max` are **integer coefficient strings**;
the summary includes its declared `scale`. A sum of `1250` at scale 2 means
`12.50` in the declared unit. This is deliberate exact transport, not a floating
point JSON number. Mean in input units is `sum / (valid × 10^scale)`.
Comparison differences and reduced rational mean differences are also in
coefficient units; their comparison binding supplies the scale. Histogram
boundaries are coefficient strings, interpreted with the question binding.
Display adapters apply scale to labels, values and tooltips. Sum/extrema changes
show fixed decimal values. Mean changes retain exact rational strings; displayed
means round to `max(3, scale)` fractional digits. Shared-range captions use an
exact terminating decimal when possible and a reduced rational otherwise.

Existing powers-of-two histograms partition the coefficient domain. At scale 2,
coefficient bounds `[64,128)` therefore display as `[0.64,1.28)`. The exclusive
maximum sentinel is geometry, not a valid predicate threshold. Range selection
retains exact original filters and uses the existing valid/missing/empty/invalid
partition for selection versus rest.

Bindings with different fields, types, scales, units or parse policies are
different measurements. Comparisons withhold arithmetic across incompatible
bindings; they never rescale or rebin implicitly. Integer bindings omit scale
and retain their existing transport. An integer binding with nonzero scale is
invalid.

## User flow and ownership

The existing numeric measure and typed filter controls own declarations. The
same question flows through totals, groups, timelines, saved views, YAML packages,
captures and comparisons. Display elements stay passive: formatting and local
selection do not fetch or subscribe. Opening a range or saved definition stages
a question; an explicit Run owns the subsequent source read.

Captured observations remain bounded evidence of independent source reads, not
a transactional snapshot, original OTLP scalar-type proof or a completeness
claim. Arbitrary floating point parsing, multiple numeric measures and unit
conversion remain outside this contract.

The coefficient-and-scale representation gives a future platform adapter a
precise lowering target: preserve input grammar and coverage before mapping
coefficients to a sufficiently wide decimal expression. A backend decimal cast
alone is insufficient if it accepts extra precision or rounds malformed input.
This milestone does not add or validate Databricks execution.

## Accepted journeys and verification

[Journey 180](../journeys/180-decimal-numerics-storyboard.md) exercises real
ordinary OTLP spans/logs: author filters, inspect exact full-population summaries,
brush decimal bins, compare selected versus rest, save/reopen and reuse YAML.
Both signals pass with 400 source records and a 100-record inspection page.
[Actual frames](journeys/README.md) include compact range and mobile comparison
views; the separate SSR/SPA component frame uses explicitly supplied synthetic data.

Acceptance includes 300 Hegel examples, 2,100 state steps, 30 boundary cases,
native raw-source/capture validation with 76 rejecting controls, finite Quint/qobs
trace replay, a 55-second ten-phase Bombadil run and a passive-component test with
zero API requests. [VALIDATION.md](../VALIDATION.md) records commands/evidence,
actual source parity, and the initial transient timeout and harness corrections.

Live Databricks and production-load performance remain separate acceptance gates.
The observed histogram SQL expansion adds overhead; current local accepted reads
retain the existing source budgets and fail closed under source errors.
