# Review after parameter-encoding consolidation

Architecture/Go and design/product agents independently reviewed the current
capabilities, source boundaries and recent implementation lessons. The adopted
finite cadence is 406 breadth → 407 depth → 408 consolidation, then review.

| Candidate | Decision and boundary |
| --- | --- |
| Ordinary captured-record activity rate | Adopt. Service dashboards already supply a fixed-window record rate, but generic summary/timeline declarations omit it. Generalize the typed data contract, not renderer-only arithmetic or inferred application throughput. |
| Authored numeric histogram presentation | Adopt. Numeric histogram execution, snapshots, rendering, range planning and question save/capture already exist. The gap is a finite authored histogram-plus-records composition and reuse arm. |
| Passive grammar entry and examples | Adopt. The public display entry omits ScopeView and MeasureSummary, and the standalone reference lacks exact goo-chart/measure-summary examples. This is discoverability/embedding coverage, not a demonstrated registration failure. |
| Grouped P95 | Defer. Scalar percentile reuse still requires independent raw-cell/population quantile support. Mean and summary support do not establish it. |
| Host sample context | Defer. Full-window matching-resource record links already ship; narrower sample context requires a concrete user need, unique host identity and separate sample gates. |
| Exponential histograms | Defer. Exporter loss and mathematical/source representation remain explicit gates. Discovery metadata does not grant evaluation. |
| Broader platform work | Tenancy, alerts, executable kinds, new stores and broad vendor parity remain separate design/platform work. |

The source → supplied snapshot → passive element → application owner → immutable
configuration/evidence layers remain useful. No measured evidence justifies a
broad clone, source-concurrency or runtime rewrite. Go 1.27.3 remains current.

Carry forward whole-carrier admission, physical SQL accounting, permanently retired
draft capabilities and actual visible-control journeys. A valid leaf does not
prove its enclosing saved/configuration carrier. Hidden controls must be fixed in
the product or opened normally, never forced by an acceptance test. Wire-format
differences are not semantic corruption without a physical-value proof.
