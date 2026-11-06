# Exact-service operation inventory — 349

Status: closed source-grounded contract after 348. Extend the existing inventory
question/compiler and service browser; no new inventory engine, endpoint, index,
cap increase or client top-K is selected.

## Why source scope changes

The service browser currently requests `/api/inventory` for the whole applied
window/environment/namespace, then filters returned rows by service in the browser.
Inventory SQL groups all ServiceName/SpanName pairs and refuses more than 1,000.
Unrelated services can therefore make a small selected service unavailable. Its
capture action correctly warns that the source read includes other services.
349 makes the selected-service task a real source question and changes that
capture wording/authority together.

## Closed selector and scope contract

Add optional `service` to the existing inventory question:

- Omitted means all service names, preserving intentional global inventory.
- A present string means exact equality, including `""` as exact reported empty
  ServiceName. Do not use truthiness or collapse empty into omitted.
- Explicit null is invalid, not omission. Values are well-formed UTF-8, at most
  128 bytes and contain no NUL, matching existing service-dashboard identity bounds.
- Unknown fields and invalid selector combinations continue to refuse before IO.

Preserve the existing environment union: scalar current-only named or a checked
current/effective selector with named/empty/unreported scope. Preserve namespace
omission/all, missing, present-empty and named distinctions and the original fixed
window. Service is an additional independent exact predicate, not a namespace
alias, guessed entity owner or implicit environment fallback.

Clone the service pointer and other caller-owned scope before acquisition. Query
construction conjoins exact service with the namespace predicate using the existing
AST/compiler. The parameterized filter must precede GROUP BY and its 1,001-row
sentinel. Do not interpolate service values, filter a returned global population,
raise the budget or silently truncate. Exact service inventory still refuses if
that selected service itself exceeds 1,000 pairs.

## Shared-type boundary: do not partially narrow discovery

InventoryQuestion is also embedded in discovery and service-dashboard requests.
Adding its field must not accidentally broaden this feature into all those APIs.
`CompileServiceDiscovery` uses q.Query for span/log reads and parent ownership,
but builds metric-family SQL separately. Blindly adding service to q.Query would
narrow only some signals and alter cross-service parent resolution.

Scope 349 to operation inventory and its capture. Reject the optional field on
`/api/discovery` through a dedicated validation-before-source boundary and compiler
check. Dashboard scope also rejects it because DashboardRequest already has its
own Service field; do not accept contradictory or ignored service selectors.
No second public question type/engine is required just to preserve these current
boundaries. Any future service-scoped discovery needs its own all-signal/graph
semantics, not an accidental effect of this field.

## Result and custody authority

The source result echoes the complete detached inventory question. Every returned
row must have exactly that service when selected, in addition to existing required
dimensions, unique service/name pair, canonical count/error bounds and exact
duration-bucket partition checks. A source returning another service is rejected,
not filtered away. An empty successful selected inventory is no_data with zero
observed spans; it does not assert health or complete instrumentation.

Stored inventory validation must check selected-service membership as well as
question equality and existing receipt/hash constraints. Capture normalization
clones and retains the pointer: present-empty must remain present-empty in JSON,
receipt identity and reopened content. The capture form submits the actual
selected inventory question and performs its existing explicit fresh source read.
Custody time and source CapturedAt remain independent. Reopen reads stored custody,
not telemetry; no alias, migration layer or automatic latest-service capture is
introduced.

The browser admits the complete echoed question and every raw row before rendering
or offering capture/continuation. Local filtering may implement Find only; it is
not proof that an environment-wide response was service-scoped. A poisoned row or
changed policy/window/namespace/service echo must refuse before staging actions.
For a valid operation row, follow-up spans retain the exact checked service,
namespace, environment and UTC bounds, adding only the literal span-name predicate.
Reported empty service and span names remain exact values. Counts/errors/buckets
are captured span records, not unique requests, business errors or root latency.

## Browser ownership and natural flow

Only the inventory request receives the service field in the current browser.
Service discovery/global dashboard scope keeps its separate input contract.
Capture title, button/context and explanation must say selected service rather
than the old whole-environment read. The intentional global inventory path remains
explicit and truthful about its cap and capture scope.

Reuse service-detail generation and cancellation, retained Activity return and
shared query/source boundaries. Changing service/policy/namespace retires pending
old work; late responses cannot render or capture under a newer service. Local
Find and retained return perform zero source reads. A refused inventory does not
invent zero operation counts; independently successful signal/dashboard sections
remain their own observations. Do not bundle new service-health/status inference
or native metric scope generalization into this cut.

## Acceptance

1. Pure Go/UI/planner cases: omitted service, literal UTF-8/quotes/backslashes,
   exact empty, null/overlong/NUL/invalid input; typed environment policies and all
   namespace modes. Source compilation binds exact service before grouping; direct
   caller mutation cannot alter the echoed scope. Invalid shared-consumer use
   refuses before source work.
2. Genuine ordinary OTel population with more than 1,000 unrelated global pairs
   and a small selected service: global refusal remains truthful, selected read
   succeeds with exact independent raw counts/errors/buckets. Namespace collision
   and current/legacy environment conflicts prove policy rather than name matching.
3. Poisoned returned service or changed question is rejected by source, stored
   capture admission and browser action admission; no client-side filtering hides
   the mistake. Above-2^53 count semantics remain canonical exact strings.
4. Actual service entry → operation Find → supporting spans → retained return →
   capture selected inventory → reopen → exact-name continuation. Capture alone
   rereads its scoped question; reopen/Find/Back do not. Record source/custody clocks,
   exact queries and requests independently of visible labels.
5. Current desktop/narrow frames, unavailable/empty distinction, keyboard focus
   and late-owner refusal. Reuse existing service-handoff/parent lifecycle models
   where the projection fits; independent SQL/membership/capture oracles prove the
   new selector. A new model merely restating a string field is unnecessary.

Close this depth cut before 350 typed-scope package authoring. Preserve the finite
347 sequence and current Go/shared-source architecture.

## Admission review: preserve raw identity before JSON repair

Exact service membership must be checked before lossy JSON string decoding can
change identity. Go's ordinary decoder replaces invalid UTF-8 and unpaired
surrogate escapes with U+FFFD. A raw `utf8.Valid` check alone misses ASCII-encoded
unpaired escapes. The existing checked environment-selector JSON validator already
handles both forms while admitting valid surrogate pairs and literal U+FFFD;
reuse that mechanism rather than add widget/parser variants.

Apply the same rule to source and stored inventory row identities. Capture loading
decodes before canonical digest verification, so a poisoned raw row repaired into
the selected literal U+FFFD could otherwise retain the canonical digest. A missing
or null stored row service must also not decode into an exact empty name. The
source already requires non-null service/name dimensions; a shared inventory row
codec can retain that raw presence/Unicode proof for fresh and stored rows without
changing unrelated capture/log formats. Post-decoding string validation alone is
insufficient. These review findings require focused admission tests before final
source acceptance.

## Frozen backend/source admission acceptance

The final backend uses the existing Unicode guard as `identity.WellFormedJSON`;
its algorithm is unchanged and callers still check JSON grammar separately.
InventoryQuestion applies it before decoding the raw service selector. The new
InventoryRow codec requires raw non-null string service/name dimensions and
well-formed JSON text before ordinary decoding. The same codec runs for fresh
source rows and stored inventory rows, closing the canonical-hash repair path.
The duplicate source-only dimension decoding was removed rather than maintained
as a second admission implementation.

Focused cases cover literal U+FFFD and valid surrogate pairs, invalid UTF-8,
unpaired high/low surrogate source dimensions, and stored exact-empty/U+FFFD rows
mutated to missing, null or unpaired service text. Question/service-pointer
isolation, foreign row rejection, selected/global cap behavior and capture
round-trip checks remain in the same suite. Discovery validates before execution;
its compiler and dashboard planning reject unsupported shared-type service use.
PopulationRequest continues to use Query/Companion rather than InventoryQuestion.

Architecture accepts the final backend/source boundary and test design; whitespace
validation passes. Coordinator-reported root Go/API gates are separate checks.
Subsequent genuine ClickHouse/OTel checks pass independently: 1,050 unrelated
pairs coexist with the selected service's seven effective/five current-only spans;
empty-name ingestion and exact selection, typed budget/unavailable refusal,
operation continuation, retained Activity Back and selected capture/reopen hold.
Actual desktop/narrow/custody frames pass review, including a measured full-width
caption through the existing stacked-table option. These native results are
recorded separately in [validation](../../VALIDATION.md), not inferred from the
synthetic source tests above.

## Next 350 readiness handoff (proposal only)

Start typed-scope package work in `exploration/package.go`: PackageSource is
scalar-only, Question copies only Source.Environment, and DraftTelemetryPackage
explicitly refuses EnvironmentSelector (including a positional PackageSource
literal). Inspect both DecodePackageJSON and ReadTelemetryPackage, then
`package_store.go` save/load/revision and `cmd/goo/telemetry_packages_http.go`
draft/plan/save. Preserve selector presence, null refusal, Unicode identity and
canonical definition hashing consistently across JSON and YAML; do not merely
add a DTO field that one decoder or Question silently drops.

UI consumers to inspect first are `telemetry-packages.mjs` and the
`exploration.mjs` Create-package guards, library description/search and
`goo:apply-telemetry-package` handler. The latter restores p.question but still
fills the scalar input from definition.source.environment and immediately runs
the question. Close the distinction between staging/opening and an explicitly
named Run action before claiming source-free reuse. Validation/build/open must
not acquire hidden telemetry; any retained explicit run behavior needs truthful
labeling and source-count acceptance.

350 remains scope parity, not new measure implementation: numeric/histogram,
duration/status, grouped-time and interval profiles already exist in the package
model. Test typed current/effective named, empty and unreported scope through
admitted draft → JSON/YAML → plan → save/load → current UI question, retaining
profile/predicates and relative lookback versus fixed preview clocks. Invalid
unions and unsupported reported-edge combinations refuse before acquisition.
Every consumer must preserve the selector or explicitly reject it, borrowing
349's shared-type lesson; scalar fallback is not an acceptable interpretation.
This handoff selects no production changes before 349 acceptance.
