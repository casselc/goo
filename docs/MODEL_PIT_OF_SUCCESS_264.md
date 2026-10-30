# Model hardening: a pit of success

The native scoped-view work is useful and its Go producers are conservative, but
extension authors currently have too many invariants to assemble by hand. The
next foundation work should reduce those choices before adding renderers or a
visual authoring system. This review is about correctness and usability boundaries,
not a claim that current source adapters emit the bad examples below.

## Findings and safer defaults

| Priority | Actual rough edge | Pit-of-success default | Status |
| --- | --- | --- | --- |
| P0 | Copying a complete ranked snapshot and slicing its rows can retain a valid-looking whole-query guarantee. Structural validation cannot establish where the data came from. | One host adoption/normalization boundary; immutable handles. Projection APIs preserve, narrow or remove guarantees automatically. Authors request requirements and cannot construct attestations. | Required hardening gate |
| P0 | `scope.generation` is a content basis, not a particular active execution. Current safe staging additionally relies on per-screen pointer/phase checks. | Distinguish source/workspace, logical scope, immutable observation/content revision, data role and host execution epoch. One resolver validates actions against the active handle. | Required hardening gate |
| P0 | Hand-built objects have redundant values, states, support and completeness flags. Omitted counts, undeclared nullability and misspelled display requirements could pass. | Shape-specific constructors derive metadata; closed authored schemas reject typos. Null, missing source value, unavailable measure, projected-out and unmaterialized are distinct. | Smaller validation issues fixed in264; constructors remain |
| P1 | Exact decimal strings used float plotability as a data-validity check; equivalent rational strings behaved differently. | Decode exact scalars once. Exact tables retain valid values; renderer compatibility checks range/precision separately. | Separation fixed in264 |
| P1 | Selection capability was spread across shape checks: an in-progress summary table could offer Select although summaries deliberately have no subgroup membership. | Registry declares capabilities; controls derive available actions, rather than failing after a click. | Central capability flags and passive summary table fixed in264 |
| P1 | Mutable component properties could change data without a revision or repaint. | Immutable normalized handles, shared once per host. Defensive property boundaries until that ownership API exists. | Defensive copies added; central adoption remains |
| P1 | “Complete”, valid measurement, exact arithmetic, sampling, read coherence and freshness are easy to conflate. `freshness: fixed/live` is really delivery mode. | Independent guarantee axes; unknown by default. Rename/refine delivery, clock, coverage, coherence, correction and continuity semantics in the normalized model. | Partial axes present; full model remains |
| P1 | Field key/unit equality is weaker than measure compatibility. Numeric coefficients, human units, quantile methods and denominators can diverge. | Stable measure descriptors carry binding, source domain, parse/scale/unit, method and support semantics. Operators declare legal combinations; quantiles cannot acquire a mean-style merge. | Exact264 comparisons validate full bindings/methods; general descriptor/operation registry remains |
| P1 | Comparison provenance could flatten independent sides or treat caller-supplied source basis as proof. | Named baseline/selection roles retain independent questions, clocks and support. Hash actual held content independently; preserve claimed source basis separately. Held data never gains source completeness. |264 implementation |
| P2 | Application screens repeat lifecycle, baseline, selection and action ownership. Future custom views could accidentally introduce another owner. | One context owns named data demands, subscriptions, draft/applied scope, history and action resolution. Components stay passive. | Prerequisite for authored whole-screen views |

The HTTP comparison request still reflects grouped-cohort implementation details
(such as numeric opt-in and categorical accounting). Normalized summary handles
should become the direct comparison inputs, without authors assembling duplicate
payloads or introducing an arbitrary grouping just to compare population measures.
This is part of construction/operation cleanup, not a reason to add more flags.

## Shared model does not mean one flat shape

Use a shared envelope and typed variants: records, intervals/series, groups,
distributions, spans/relationships, population summaries and comparison roles.
Do not force them into a generic row bag that erases temporality, support, interval
closure, relationship certainty or source identity. The current ranked/time/summary
variants are a beginning, not an exhaustive universal model.

Likewise, a valid schema is not source authority. `origin: source-read` is descriptive
metadata, not authentication. The host classifies data at an actual trusted read
boundary; imports/browser-held data enter conservatively. Declarative templates
bind a host data role and may request guarantees. They must not provide their own
origin/completeness/provenance facts to satisfy those requirements.

A safe API should let ordinary authors supply role, field references and kind:

```js
// Illustrative future API, not implemented public methods.
const current = await host.run(recipe);
view.bind('current', current.data);
view.add({kind: 'bars', data: 'current', category: 'service', value: 'duration.p95'});
```

The compiler resolves field references to checked descriptors. Constructors derive
units, cell availability, support and guarantee lineage; authors do not manually
synchronize `values`, `cells`, `quality` and attestations. Advanced renderers consume
the same normalized facts and cannot bypass requirements or source ownership.

Backend planning must also preserve the declared method and numeric range. An
exact ClickHouse quantile must not silently become an approximate percentile on
another platform; unsupported precision/methods need a plan diagnostic or an
explicitly different measure contract.

## Identity, transformations and actions

Separate logical scope from a particular read and from immutable contents. A
same-content rerun may retain the same content revision but must have a new active
execution epoch. A baseline and current selection are different named roles even
when their queries or contents coincide. Source identity must be part of the host
context before multiple sources/workspaces can share a composition.

Provide a small transform vocabulary with declared grain and guarantee rules:

- Sorting/relabeling can retain population membership and exact values.
- Filtering supplied rows is a local subset, not a complete parent query.
- Authoritative population filtering is a new checked recipe/read.
- Grouping a row page cannot become a full-population aggregate.
- Combining means needs sufficient statistics, compatible units and a defensible
  population relationship. Combining scalar quantiles is unavailable.

Selections are tagged capabilities. A selected group is a predicate refinement;
a selected interval is a bounded time refinement; neither is an attested list of
all member records. Only verified record references authorize keeping those
records. A central resolver validates role, epoch, revision, stable key and current
scope before executing the captured intent. Components never re-read a global
“current scope” at completion time to retarget an older action.

## Implementation order

1. **Normalize and seal.** One internal snapshot construction/adoption path;
   discriminated shapes, immutable data, exact scalar decoding, closed options and
   explicit availability. Use shared schema/golden cases across Go, JS and OpenAPI
   instead of independently evolving the wire rules in several places.
2. **Centralize ownership and transformations.** Host execution epochs and data-role
   handles; one action resolver. Safe projection APIs with guarantee propagation.
   Keep current source/measurement checks while replacing repeated screen guards.
Normalize once and share immutable handles across components, with bounded local
projections, whole-snapshot byte/node budgets and batched updates. Defensive copies at each widget boundary are a
temporary safety measure, not the desired live-update performance architecture.

3. **Compile contributed views.** One canonical ViewDefinition/recipe model with
   typed field handles, kind capabilities and declared requirements. Source/YAML
   and visual editors share one AST. Renderer/vendor expansion follows these gates.

The bounded264 summary/comparison integration uses conservative producers and
includes the smaller fixes above. It also removes an actual refresh path that
could silently discard declared numeric measures to preserve an older profile. The next milestone is this hardening work,
not another gallery of widgets, a Plotly shortcut or an editor over raw snapshots.

## Acceptance tests

- Identical-content rerun rejects the old action despite equal content revision.
- Filtering supplied groups loses the complete-parent guarantee; exact member
  selection never turns aggregate groups into record references.
- A misspelled requirement refuses visibly; malformed/null/omitted fields fail at
  the construction boundary with actionable diagnostics.
- A valid exact number outside float range remains readable in the table.
- Exhaustive groups with invalid-only values remain complete in extent but have
  unavailable measures; zero is not substituted.
- Replacing a baseline does not retarget a current-role selection or an in-flight
  durable action. Different sources cannot alias through an identical query hash.
- Units, parse/scale, method and support changes require explicit compatibility;
  count differences do not masquerade as normalized rates or causal effects.


## H0 delivery265

[Immutable snapshot construction](SNAPSHOT_CONSTRUCTION.md) implements normalized,
bounded admission and shared frozen handles. Source-adoption witnesses are private
runtime facts; imported JSON and supplied-row subsets cannot recreate whole-query
proof. Typed source-question/profile matching uses the existing question protocol.
This replaces the temporary per-widget snapshot-copy approach for the scoped pilot.
266 adds [bounded H1 execution/role ownership](SCOPED_ACTION_OWNERSHIP.md) for
canonical fixed Explore group/facet and timeline staging. Same-content reruns and
suspended visits revoke old mounts. Other families still need explicit migration.
Neither a source label, source adapter nor action context authenticates producers
or proves a human click.
