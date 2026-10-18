# Scoped investigation recipes and shared examples

Implemented locally. Explore → Reusable questions provides three built-in question
transforms: span duration by operation name, recorded span errors by service, and
log counts by exact severity. They are compiled by the existing Go OTel adapter;
there is no second evaluator, SQL authoring surface or new recipe language.

Start with an applied Explore question. Preview preserves its signal, environment,
half-open fixed time bounds, full predicate tree and chart interval. It replaces
breakdown and measure settings and resets inspection to offset 0 / limit 100.
Preview is source-free; Run this question explicitly performs a fresh source read.
Wrong-signal recipes are rejected rather than translating filters across signals.
The shared [scope/transform contract](QUESTION_SCOPE.md) now derives keeps/changes
from actual question fields and checks same-population claims. Previews display
population grain and disclose unbound convention/schema metadata explicitly.
Edits, a new execution or navigation invalidate pending previews. The UI requires
another applied question after draft changes; stale responses cannot reopen a plan.

The prerequisites are declared, not automatically verified instrumentation
coverage. Recipe execution retains normal source no-data, disagreement, unknown,
group cardinality and query-budget behavior. Operation names do not distinguish
span kinds automatically; add an explicit kind/service filter when required.
Recorded Error / all matching captured spans includes retries, Unset and unknown
statuses. It is not a business failure rate. Severity is case-sensitive exporter
text. Scalar percentiles are never averaged across groups.

## From question to evidence

Add hypothesis to notes creates only a plain-text hypothesis describing the
question and its limits. It attaches no evidence and asserts no findings. Run the
question, use Capture investigation to server-reread/freeze its exact population,
then add that stored capture to the same notes. The capture can contain later
arrivals than the live result. Create package saves reusable filters/measures in
the existing versioned telemetry-package schema; running a saved package resolves
a fresh window, whereas capture refresh preserves fixed bounds. To retain the hypothesis’s original scope,
capture it before reusing a package. A stored capture’s Create package action can
derive the reusable definition without changing that evidence.

GET `/recipes` lists the three compiled-in contracts. POST `/recipes` accepts
`{id, question}` and returns recipe metadata, transformed question and parameterized
source plan, without a ClickHouse request. The Python UI proxy exposes these under
`/api`. These are bounded built-ins, not user-defined multistep workflow storage.

## Examples use the product

The primary Examples gallery now loads six ordinary YAML telemetry packages from
[examples/telemetry](../examples/telemetry): spans and logs for the reservation
pipeline, woven Go checkout and woven Go reports. Each has a five-minute lookback,
explicit local environment/service population and normal adapter bindings. The
reservation selector is an OR across web, gRPC and consumer services. The gallery
itself only reads local definitions; a button plans and runs a fresh question.
The package source is visible and can be refined, saved, exported, captured and
cited through the same product workflow as arbitrary OTel applications.

The [shared measurement packages](MEASUREMENT_PACKAGES.md) now add gauges,
native intervals and live lifecycle contracts alongside these captured-record
packages. Both Go examples emit real OTLP logs. The original reservation
lifecycle/resource lab remains an explicit collapsed compatibility section and
only starts querying when opened. Its historical preview assumptions are retained;
the primary measurement path uses explicit local custody and unknown absence.
Go application operation manifests and native metric contracts remain available
under Operations. The new packages do not require goo annotations and also work
over the Python sample. A sample with no exported logs correctly shows no data;
a package cannot create missing telemetry.

[Storyboard 37](journeys/README.md#37-shared-examples-and-scoped-investigation-recipes)
executes real report work, independently discovers its ClickHouse trace, opens a
sample package, narrows it, previews/runs a recipe,
freezes/cites evidence, then derives and reuses a definition from that capture. Hegel state/swarm checks preserve population across recipe,
signal, scope and predicate changes. Quint observed-trace checks cover preview →
run → fixed capture → narrative → fresh package reuse, with rejecting controls. This is bounded conformance,
not an exhaustive proof of browser interleavings. Bombadil additionally explores
preview guards without an applied population or source writes.

Remaining: capability discovery for richer recipes, UI-authored ordered steps,
explicit successor comparison in narratives,
authenticated collaboration and production storage/retention.
