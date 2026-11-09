# Roadmap after the 390 checkpoint

Adopted November 9, 2026 after the 390 whole-product checkpoint. This replaces the
completed [387 checkpoint queue](ROADMAP_2026_11_09_CHECKPOINT_387.md).
The [current capability index](CAPABILITY_GUIDE.md) summarizes delivered behavior;
[validation](../VALIDATION.md) distinguishes checks actually run. This is one
finite next queue, not parallel implementation plans.

Inputs: the independent [whole-product design](reviews/WHOLE_PRODUCT_DESIGN_390.md)
and [architecture](reviews/WHOLE_PRODUCT_ARCHITECTURE_390.md) reviews, the
[current capability index](CAPABILITY_GUIDE.md), recent feature contracts,
source inspection and [actual validation](../VALIDATION.md). Older reviews are
reconciled inputs; their historical “next” labels do not activate additional work.

## Direction and lessons

Keep the core loop **current logs/spans → isolate a population → follow evidence
→ preserve or reuse the question**. Ordinary OpenTelemetry remains sufficient;
manifests improve authoring but are not admission tickets. Existing Activity,
typed filters/facets, comparisons, log patterns, trace search/waterfalls, service
and resource dashboards, captures, annotations and authorable views are delivered
capabilities, not missing categories to recreate.

The outgoing sequence closes three concrete native gaps: ordinary gauge
observation-to-record context (388), counter display reuse (389), and data-first
[native composition](NATIVE_DATA_FIRST.md) (390). The next breadth item returns to
ordinary logs and the shared grammar; do not let successive native-adoption
follow-ups become the whole roadmap.

Recent work establishes acceptance rules for the next cuts:

- Accepted evidence, pending requests, authored configuration and execution
  authority remain separate. Source-free staging from held evidence must not
  reactivate source/publication leases.
- Preserve the complete destination question on reuse. Metadata compatibility
  must agree with the loaded definition; family, units, domain and field presence
  are real boundaries, not cosmetic labels.
- Unknown calculated values can have recorded source exports. Distinguish
  calculation unavailability, optional values not supplied and recorded zero.
- A model pass cannot establish truthful geometry. Check actual rendered data
  items, clipping ancestors and real scroll containers. Test-driven scrolling is
  not evidence of a useful entry layout.
- Component chrome belongs to component styles. Keep passive density/disclosure
  presentation separate from source reads, navigation and session retirement.
- Collector acceptance is not source settlement. Independently admit the raw
  fixture bank; distinguish browser operations, physical SQL and setup artifact
  publications. Observe completed outcomes, not clicks or actor phases alone.
- Put Go/JavaScript/API default, omission, null and foreign-field vectors into
  each new finite contract's acceptance. The 389 empty-domain schema correction
  is a reason for that discipline, not another validator or profile engine.

This prerelease has no legacy route, format or API floor requiring aliases or
migration machinery. Use the current Go 1.27.3 toolchain and standard-library
facilities where they simplify a touched path. No speculative runtime upgrade or
advanced caller-side type machinery is scheduled.

## 391 — breadth: reusable log-pattern count displays

User task: “Keep this useful view of log shapes and apply it to another scoped log
population without rebuilding its presentation.”

[Log-pattern exploration](LOG_PATTERNS.md) already provides deterministic v1
normalization, exact counts/shares, unclassified groups, record continuation,
overall timelines, time comparisons and captures. The concrete gap is portable
display admission: `workbench/view_declarations.go` and `ui/view-declaration.mjs`
explicitly exclude `column:body_pattern_v1` from the ordinary Bar/Table grammar.

- Add a checked, count-only log-pattern display profile through the existing
  declaration, supplied-data renderer and immutable exploration-view owner.
  Keep applicability finite: logs grouped by this exact versioned field. Do not
  merely remove the exclusion and treat encoded identities as ordinary labels.
- Canonical `lp1` identifiers remain selection/action/filter keys. Readable labels
  are escaped presentation of inferred text shapes, not canonical identity,
  incidents or causes. Empty, oversized and invalid-body groups retain their
  explicit unclassified meanings and full-population counts.
- Support existing Bar/Table and count/population-share presentation with local
  Find/order/kind changes. Preserve full denominators and coverage; a filtered or
  bounded display must not become the population. No top-K truncation is admitted
  as a complete partition. Changed-between-read evidence must not acquire valid
  share/drill authority merely because the declaration is compatible; complete
  groups do not establish complete upstream ingestion.
- Save/reopen/reuse uses the existing configuration workflow. The destination
  retains all filters, environment policy and fixed bounds; declaration reuse
  performs no telemetry read. Run remains explicit. Existing pattern record and
  comparison continuations retain their own checked scope and held return.
- Before coding, finalize the smallest explicit display/data-profile contract
  across Go, JavaScript and OpenAPI. Keep it distinct from ordinary string
  categories and from numeric/duration measures that this cut does not admit.

Acceptance: shared positive/refused contract vectors; canonical ID versus label
collision/control-character/unclassified cases; full count/share oracles; actual
ordinary OTel logs → pattern analysis → choose/save display → different scoped
destination → inspect/apply → explicit Run. Verify zero source reads on local
changes and configuration operations, exact destination identity, unchanged
captured bytes and current-owner guards. Refresh storyboards, frames, Bombadil
actions and measured Quint checks with rejecting controls.

Boundary: no new normalization version, clustering engine, arbitrary text
classifier, template-expansion arm or grouped-pattern time-series grammar. The
existing overall timeline and comparison functionality are not missing features.

## 392 — depth: native outgoing HTTP duration observations

User task: “This service reports HTTP-client duration histograms; show their
measured request-duration distribution and let me investigate the observations.”

[Outgoing HTTP Client-span observations](OUTGOING_HTTP.md) are delivered in 385.
Native `http.client.request.duration` profile execution is separate: current
instrument capability/dashboard binding code recognizes server duration, and
385 tests explicitly prevent a generic client histogram from acquiring outgoing
profile authority. Extend the established histogram path rather than adding an
executor or relabelling incoming HTTP data.

- Pin and review the relevant OTel semantic convention before enabling execution.
  Define exact metric name, explicit-histogram source representation, supported
  schema/unit/temporality and required-versus-optional dimensions. The current
  standard metric uses seconds; do not inherit another unit or alias policy
  merely because a generic histogram kernel supports it. Generic duration
  execution may already be available: this cut adds justified discovered-profile
  eligibility/defaults, not a second evaluator.
- Add a finite discovered native binding using existing per-stream evaluation,
  source/capture contracts and supplied-data histogram/table displays. Preserve
  service, instance, namespace, environment, full stream identity, reset/gap and
  export-coverage semantics. No fleet percentile averaging or scalar-percentile
  substitution.
- Keep native-metric evidence independent from Client-span population evidence.
  A supported histogram does not prove Client spans, a nonempty current span
  method field, a unique request count, retries, callee identity or remote health.
  Metrics-only discovery must remain useful without silently enabling a record
  lens whose mandatory evidence is absent.
- Literal destination/method/status context may be exposed only when supported
  by the admitted source/profile. Do not infer default ports, aliases, destinations
  or a new failure policy from names. Existing record lenses keep their predicates
  and independent source clocks.
- Use current native export/interval and capture/reuse mechanisms. Any new
  reported-exemplar continuation is separate unless its existing authority
  contract can be adopted without widening this acceptance cut.

Acceptance: convention/source matrix; supported/refused source metadata; actual
metrics-only, spans-only and combined cases; independent raw bucket/count/range
oracles; units and cumulative/delta/reset/gap/unknown controls; exact source and
destination question identity; explicit reads and no-read local/held operations.
Preserve incoming HTTP, RPC, messaging and generic discovery regressions. Add
actual journey/actor/model outcomes and truthful data-first frames.

Boundary: no HTTP health score, attempt deduplication, remote-service topology,
automatic retry inference or general exponential-histogram execution.

## 393 — consolidate: prove literal source-parameter transport

User task: “A literal filter or metric binding should select the same bytes in
either source path, including unusual but valid attribute text.”

The previous architecture review identified differing encoders:
`exploration/source.go` hex-escapes all ASCII controls and DEL, while
`measurements/source.go` replaces backslash, LF, tab, CR and NUL. Source inspection
alone does not prove a defect or a performance benefit. Resolve that uncertainty
with a finite actual round-trip audit before consolidating implementation.

- Build legal parameter vectors covering backslashes, ASCII controls/DEL and
  Unicode, bounded by each public field's existing validation. Keep malformed or
  prohibited input as explicit refusal cases rather than calling it valid data.
- Exercise record and metric source paths against a settled ClickHouse bank with
  known exact values. Independently check returned identity/membership and the
  actual typed parameters, rather than asserting equality of encoded strings.
- If a correctness difference is demonstrated, share one small pure Escaped-text
  encoder at the source boundary and prove the previously failing round trip.
  Keep URL policy, readonly settings, limits, family error classes and mathematical
  evaluation in their existing owners. Do not create a universal source client.
- If both paths round-trip the supported values correctly, finish with the
  documented parity result and regression vectors. Do not force a refactor to
  satisfy a planned code-change quota or claim an unmeasured speedup.

Acceptance includes negative controls that alter a literal or decoding boundary,
unchanged ordinary/native source questions, no widened row/byte/time budgets,
Go semantic properties and actual source membership. A benchmark supports a
performance claim only if performance work is actually undertaken.

This is one bounded consolidation task with an explicit no-change outcome, not a
parallel transport rewrite. Reassess the whole product after 393.

## Reconciled deferred work

| Earlier candidate | Disposition after this checkpoint |
| --- | --- |
| Gauge observation context | Ordinary Measurements delivered in 388. Saved/service adoption remains useful but needs explicit retained-sample admission; it does not displace log breadth in this queue. Stocks and resource contexts stay separate. |
| Counter display reuse | Delivered in 389, including exact domain/unit compatibility. Do not build another catalog or evaluator. Resource-bound counters and other grains are separate work. |
| Native data-first entry and component styles | Delivered in 390 with actual natural-entry geometry. Do not turn the completed cut into another broad stylesheet/framework rewrite. |
| Contextual review-task chooser → verified kept-occurrence review | Still a coherent later pair. Existing captured-record review, labels and export are delivered; do not create a fresh capture to impersonate an original kept observation. |
| Service/Saved exemplar adoption and wider native contexts | Separate retained-evidence/citation admission; existing histogram execution is not the missing category. |
| Log-pattern normalization/timelines/comparisons | Delivered execution; 391 closes display portability only. |
| Shared native default/presence conformance corpus | Include relevant vectors in each new arm's acceptance. Not another native-only milestone or universal runtime validator. |
| General exponential histograms | Retain the [source/mathematical gate](EXPONENTIAL_HISTOGRAM_READINESS.md); observed metadata is not execution authority. |
| WYSIWYG/contributed executable kinds, production tenancy/presence, alert delivery, broad DB/broker server health, warehouse lowering/new stores | Wider design/validation initiatives, not implicit prerequisites or parallel queues. |

Before beginning each item, recheck current source and acceptance results. Revise
this one queue if actual use exposes a more important core gap; do not append
unbounded plans to historical backlogs.
