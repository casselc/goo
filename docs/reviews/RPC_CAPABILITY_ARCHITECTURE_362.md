# Current RPC capability parity — architecture contract 362

Consolidation selected by checkpoint 359 after retained-page breadth/depth 360–361.
This is a bounded metadata/planner alignment, not a new convention registry or
source engine. It reuses the [357 official pin](../CURRENT_RPC_CONVENTION_357.md):
OTel semantic-conventions v1.45.0, commit
`fb30ca17193b551473b1f43c315482918ab6987c`. Implementation/runtime status must be
recorded separately; this review performs no heavy checks or new network reads.

## Concrete gap

`exploration/dashboard.go::dashboardNativeRPCCurrent` recognizes
`rpc.server.call.duration` with unit seconds and delegates to the existing native
histogram/scope binder. Service defaults label it as independent current RPC
native evidence. `measurements/instrument_capabilities.go::PlanInstrument` has
curated HTTP, pool/runtime and messaging branches but no RPC branch, so metric
discovery presents the same instrument as generic. `observedCoverage` likewise
omits RPC's meaningful conditional/required attribute coverage. This is a real
cross-surface description gap, not a reason to broaden metric eligibility.

## Small shared facts, separate authorities

Introduce or reuse a small ordinary Go helper for the exact current metric shape:
name `rpc.server.call.duration`, `histogram`, unit `s`. Both the curated capability
branch and service native recognizer consume that fact. The helper must not depend
on `CapabilityPlan`, observed stream previews or service policy; doing so would
create a planning/admission cycle. Constants for the pin/name/unit can live beside
this helper if they avoid drift. No public descriptor DSL, generated validator,
global registry, advanced user type API or legacy alias normalization is needed.

The existing service `dashboardNativeDuration` remains authoritative for supported
binding, matching definition kind/name/unit/scope/version/schema, exact service/
namespace/environment and supported OTLP temporality 1/2. Do not replace those
checks with the shared shape helper. The RPC convention itself does not require
one temporality; this remains a Goo transport/evaluator capability. No new schema
URL requirement is introduced: a schema URL is part of exact instrument identity,
not proof that a producer meets every convention requirement.

## Discovery description and coverage

Add a curated current server-call-duration capability using the existing
`InstrumentCapabilityPlan` fields: candidate title, pinned rule set, shape match,
verification `not-run`, reasons, hints and observed coverage. The instrument
profile describes native metric metadata, not the span-record grain; do not imply
a native response proves the current RPC server-span population is observed.

Keep `ShapeMatches` separate from generic `Eligible`. A current metric name with a
wrong unit/type may still have a valid generic package binding under existing
kernels, but must not become a current RPC default or be silently converted. Keep
custom metrics and the distinct legacy `rpc.server.duration` milliseconds behavior;
no compatibility floor requires aliases, and no reason exists to delete useful
older observations.

Watch bounded current attributes such as `rpc.system.name`, `rpc.method`,
`rpc.status_code` and `error.type` through existing `DimensionCoverage`/
`DimensionPresence`. Preserve missing, empty and reported values independently.
System name is required by the convention; method and status are conditional.
Missing method must not reject eligibility, remove a stream or become synthetic
`_OTHER`. A dimensions list alone cannot validate actual values, and a bounded
preview cannot certify all exported streams. Error/status labels do not become a
generic protocol-success or conformance policy.

Hints should make exact stream identity, independent source/read clocks and
interval/reset/coverage evaluation clear. Seconds are not legacy milliseconds;
per-stream histograms are not span counts or fleet percentiles. Name/type/unit
recognition is a candidate, not measured values, complete producers or a request
volume/health statement.

## Source and API limits

`measurements/discovery.go` already calls `PlanInstrument` on discovered metadata
and observations; its existing DTO is sufficient. Do not add a source read merely
to populate a capability description or alter sampling/preview budgets. Service
planning remains source-free, and selecting/running the exact measurement stays
explicit. Current/legacy auto priority and same-row positive server-span witnesses
from 357 remain unchanged. UI raw-row witnesses stay independently checked rather
than generated from the native capability helper.

## Acceptance

Test exact current and legacy names, histogram/gauge/counter kinds, seconds/
milliseconds/custom units, supported/unsupported temporality and definition
mismatches. Show that only the current correct shape is recognized by both
surfaces while generic bindings remain available where previously eligible.
Verify raw observation coverage for current system present/empty/missing and
optional method/status/error dimensions, including omitted previews. Test exact
service/namespace/environment/scope/version/schema admission at the service
boundary; the shape helper must not bypass it. Preserve custom metric behavior,
legacy profile identity, original question/stream bytes, source budgets and native
per-stream evaluation. Native discovery/service frames should show consistent
candidate meaning and the original independent identity, not a new health score.

## Candidate source acceptance

The frozen candidate implements `measurements.CurrentRPCDurationShape` with only
exact name, histogram kind and seconds. Service recognition delegates to it and
then retains the full existing native duration binder. `PlanInstrument` adds the
`rpc-server-call-duration` capability, pinned rule set and candidate title without
changing generic `Eligible`; wrong shapes receive explicit hints rather than unit
conversion. Coverage watches the four current keys and preserves conditional
method/status/error semantics. No schema gate or source read is added.

The actual discovery tests exercise `discoverRows` for delta/cumulative seconds
and milliseconds, raw method present/empty/`_OTHER`/missing coverage, custom schema
identity and unchanged generated bindings with or without preview observations.
Shape-only recognition is separately shown not to grant eligibility, and wrong
legacy/client/custom names or kinds/units are refused by the shape helper. The
small `slices.Sort` cleanup preserves existing deterministic ordering.

Read-only source acceptance is closed with no blocker. The parent reports the
first full Go suite passing; final focused cleanup rerun and generated/native/
model/frame gates are recorded separately rather than presumed here.

## Execution acceptance

The parent reports canonical native acceptance in 47.995 seconds (runner 49.729),
including all six actual read boundaries, metadata-only planning and unchanged
healthy widget DOM. Capturing explicitly performed its own fresh read; refused-
source reopen then held source calls at 11 and SQL at 40, adding one store GET
while preserving full payload, clocks and bytes. Required observed validation used
the actual discovery→inspect `instrument_presence` lifecycle, passed in 1.50 seconds
with zero skips and six Go corruption controls; it did not invent a six-step model.
The positive actor passed 605 snapshots and nine stages in 35.110 seconds against
actual rendered data. Source and execution acceptance are closed. The fixture
catalogue-scope correction changed the oracle to the actual environment-scoped
catalogue, not product source behavior.
