# Raw explicit histogram acceptance — 2026-11-15

Independent architecture/OTel and UX review accepted the separate raw-point
profile and actual 1440/390/320 layouts. The new family reuses the raw observation
owner, held table and inspector. Fact/detail helpers remain passive.

The review confirmed exact UInt64/BigInt bucket arithmetic, legal omitted
vectors versus all-values buckets, and separation of recording state, structural
validity and completeness. Unknown flags stay unknown. Flag 1 suppresses normalized
payload but keeps source facts inspectable. Wrong tuple/window/identity refuses
the whole response; malformed payload remains diagnostic.

The source exporter loses optional Sum/Min/Max presence. Every statistic therefore
keeps presence unavailable; absent or incompatible physical columns project NULL
without adding a source query. No source zero is promoted into a reported minimum,
maximum, sum, evaluated distribution or health result.

Review and live testing improved two accessibility details: bucket-page ranges
announce through a status element, and the native tables themselves have explicit
accessible names. A 151-bucket regression exercises paging, disabled/inactive
controls, retained detail and zero source work.

Actual OTLP points, an independent source oracle, declared missing-column and
late-response faults, retained Return, separate units/temporalities and a
Bombadil journey cover the operator flow. Hegel and Go/JS banks challenge exact
facts; observed validation reuses the existing population interaction model.
Cancellation is modeled as retirement of an owned pending read, distinct from a
new read or write. Negative controls retain that boundary.

The minor follow-up is consolidation: compact repeated inspector provenance and
verbose clocks through shared timestamp/tooltip components. No blocking visual or
semantic findings remain. See [executed checks](../../VALIDATION.md) and
[storyboard](../journeys/raw-histogram-observations.md).
