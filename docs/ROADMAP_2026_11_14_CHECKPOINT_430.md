# Roadmap after the ad-hoc telemetry entry pass

The [four-expert review](reviews/ADHOC_OTLP_REVIEW_2026_11_14.md) prioritizes reaching
useful unfamiliar telemetry before more preservation depth. This checkpoint
adopts that direction without renumbering the older pending 427–429 contracts.

Status: **430 delivered and validated**; 431–433 are planned.

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

Reuse 431's capabilities for bounded environment/service/instrument discovery.
Remove the generic Measurements `local` default and avoid requiring a known
service before discovering that one exists. Preserve namespace, environment
policy, unit/schema/temporality and existing instrument support reasons.

From an empty live stream, offer a wider fixed-window search that retains its
selector and held parent. Event-time windows must be visible; newer read time is
not proof of ingestion freshness. Keep clock skew, source retention, data absence
and source failure distinct. Do not introduce arrival-order guarantees.

Acceptance: no manifests, unknown labels, metric-only data, sparse/late data,
source refusals, explicit reads and keyboard/mobile completion.

## 433 — consolidate discovery and recovery

Use the exercised paths to unify scope selection, per-source state/refusal,
retry and action ownership across Activity, Services and Measurements. Consolidate
only matching contracts. Preserve local staging versus source reads, current-key
versus effective environment policy, and held results versus active actions.

Run the finite first-use matrix from the review, refresh affected storyboards,
and perform the next whole-product review. Measure before claiming performance
improvement; no compatibility shims are required for this prerelease.

## Retained follow-on work

[Exact saved-view recall, paired-capture adoption and preservation consolidation](ROADMAP_2026_11_13_CHECKPOINT_426.md)
remain useful pending work. Their prior acceptance contracts remain available;
none is replaced by an improvised discovery-specific view or capture format.
