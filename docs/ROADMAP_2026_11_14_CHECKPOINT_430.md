# Roadmap after the ad-hoc telemetry entry pass

The [four-expert review](reviews/ADHOC_OTLP_REVIEW_2026_11_14.md) prioritizes reaching
useful unfamiliar telemetry before more preservation depth. This checkpoint
adopts that direction without renumbering the older pending 427–429 contracts.

Status: **430–435 and 436.1–436.4 delivered and validated**. See
[shared refinement and held handoff](REFINEMENT_CONSOLIDATION.md). Next:
**436.5 revision-owned preservation and reuse**, under the
adopted [whole-product CEL integration design](CEL_PRODUCT_INTEGRATION.md).

The user-directed [CEL predicate pilot](CEL_PREDICATE_PILOT.md) is also delivered:
a real versioned CEL extension, checked overload lowering, `plan-cel`, immutable
parent refinement and CEL/Goo/ClickHouse parity. The specialist review now sequences
domain ownership, diagnostics and honest representability before app authoring.
An owned refinement slot preserves unsupported parent predicates; repeated edits
replace the slot rather than accumulate AND clauses. Attribute maps still need an
explicit presence contract. No new executor or competing platform queue is adopted.

## 430 — first-use scope entry and interaction reliability

The existing Find environments chooser is adjacent to Environment, highlighted
for unscoped generic entry and readable at narrow widths. Preview/staging retains
typed environment policy, Start remains explicit, named controls are accessible,
and keyboard Pause/Resume preserves focus. Unchanged status announcements are
suppressed. See the [contract](UNKNOWN_DATA_ENTRY.md),
[storyboard](journeys/unknown-data-entry.md) and [validation](../VALIDATION.md).

This is the bounded first tranche of unknown-data entry, not completion of the
entire review. No source capability API, metrics bootstrap or new ingestion path
is claimed delivered.

## 431 — breadth: discover available signal families

Implemented contract: [source capabilities](SOURCE_CAPABILITIES.md), with
[executable partial-installation journey](journeys/source-capabilities.md).
Final executed results are recorded in [validation](../VALIDATION.md).

Outcome: a user can find which supported telemetry families are available without
guessing the signal or assuming every exporter table exists.

Introduce the smallest fixed-adapter capability/readiness result reusable by
Activity, Services and Measurements. Keep discovery explicit and bounded. Check
supported tables and required columns; distinguish absent, unreadable,
incompatible, successful-empty and observed data. Metadata access denial must not
be interpreted as table absence. A missing unrelated family must not prevent
inventory over available sources. Keep count units per family explicit.

Acceptance: isolated logs-only, spans-only and metrics-family schemas, denied
metadata/read access, unavailable source and empty supported tables. Choosing a
supported signal stages existing scope discovery; it must not silently scan
unsupported tables or rewrite source configuration. Use existing source adapters,
not browser SQL or a second collector.

## 432 — depth: metrics-only and sparse-data first use

**432a implemented:** [metric scope discovery](METRIC_SCOPE_DISCOVERY.md) and its
[executable journey](journeys/metric-scopes.md) reuse 431's capabilities for bounded environment/service/instrument discovery.
Remove the generic Measurements `local` default and avoid requiring a known
service before discovering that one exists. Preserve namespace, environment
policy, unit/schema/temporality and existing instrument support reasons.

**432b delivered:** [Activity recovery](ACTIVITY_SPARSE_RECOVERY.md) and its
[executed journey](journeys/sparse-data-recovery.md) stage a wider fixed-window search
when a successful unlimited read leaves no loaded observations. It retains its
selector and held parent. Event-time windows must be visible; newer read time is
not proof of ingestion freshness. Keep clock skew, source retention, data absence
and source failure distinct. Do not introduce arrival-order guarantees.

Acceptance: no manifests, unknown labels, metric-only data, sparse/late data,
source refusals, explicit reads and keyboard/mobile completion.

## 433 — consolidate discovery and recovery

**433a delivered:** [instrument preflight](INSTRUMENT_PREFLIGHT.md) checks all
supplied observations for the existing service evaluator’s reported nonempty
instance requirement. Measurements and Services share supplied-plan readiness and
actual callback guards. Resource-bound host/container policies remain separate;
nil observations remain unknown. Long metadata is disclosed, specific refusal
reasons remain visible, and partial-source catalogs retain a visible cue. Activity
draft comparison now uses existing semantic query identity rather than raw key order.

The [independent checkpoint review](reviews/CONSOLIDATION_REVIEW_433.md) accepts this
reuse boundary and keeps discovery controllers separate: their completeness,
preview, source-presence and stream-custody contracts differ.

**433b delivered:** [Activity source recovery](ACTIVITY_SOURCE_RECOVERY.md) preserves
typed refusal state and held observations, distinguishes transient reconnect from
stable configuration/access/schema refusal, and offers explicit Retry. Failed
reads cannot become empty-data recovery or automatic wider queries. The
[acceptance review](reviews/TAIL_SOURCE_RECOVERY_433.md) covers strict protocol,
virtual-time rollover, safe causes, held selection, and keyboard/mobile recovery.

## Next operator cycle

**434 breadth delivered:** [Activity message search](ACTIVITY_MESSAGE_SEARCH.md)
keeps local Find unchanged and stages case-sensitive literal Body search over the
latest completed 60-second event-time window. The entire applied selector stays
inside an AND; only Run reads. Real OTLP cases, an independent Bombadil campaign
and native verification cover literal/case/ID differences, no-read staging/return
and held custody. [Storyboard](journeys/activity-search.md).

**435 depth delivered:** [Numeric severity](ACTIVITY_NUMERIC_SEVERITY.md) extends
the existing volume strip with all eight exact numeric categories and supplied
time context. Local band choice keeps the all-log scale; Filter stages an exact
numeric predicate for Apply. SeverityText remains independent. Real OTLP includes
unspecified/invalid values, more records than the retained buffer and banded
explanation with actual held numbers. [Storyboard](journeys/activity-severity.md).

- **436 consolidate — aggregate selection and held return.** Align log-volume and
  span-latency gestures only where read-gate and passive intent contracts match;
  refresh narrow/desktop flows and measure before claiming speedups.
  The [reconciled plan](reviews/CEL_HOLISTIC_PLAN.md) divides this into six small
  slices: 436.1 ordinary refinement/cloning and matching aggregate return; 436.2
  pure checked planning/positioned diagnostics; 436.3 representation coverage and
  revision-safe slots; 436.4 shared contextual editor/compact chrome; 436.5 one
  existing preservation owner; 436.6 observed whole-product acceptance. 436.1–436.4 are delivered; remaining slices are
  proposed changes, not delivered UI. Do not require a universal model, general
  extension registry or full attribute vocabulary before the bounded bridge.
  Carry forward 435's closed refinement lesson: validate the original selector AND
  a declared band/range, rather than weakening exact-parent guards. Keep physical
  row fields, readiness and aggregate projections aligned. Do not generalize
  controllers with different acquisition or custody contracts. Shared UTC-bound
  formatting can improve the narrow date wrapping identified in 434.

Existing filters, charts, dashboards and waterfalls are delivered capabilities,
not additional missing primitives. The following boundaries remain in force.

Generic resource/scope/instrument/attribute stream evaluation stays a separate
architecture decision for gauges, counters, histograms and epochs. No synthetic
instance IDs, merged producers or silent kernel guard removal. Measure metadata
probe latency before optimizing; prefer bounded request-local reuse over durable
readiness caching if measurements justify it.

## Retained follow-on work

[Exact saved-view recall, paired-capture adoption and preservation consolidation](ROADMAP_2026_11_13_CHECKPOINT_426.md)
remain useful pending work. Their prior acceptance contracts remain available;
none is replaced by an improvised discovery-specific view or capture format.
