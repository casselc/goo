# Service handoff source audit — 316

Source review, not a new discovery API or measured performance claim.

## Supported scope boundary

| Owner | Environment input | Meaning |
| --- | --- | --- |
| Record Query | legacy string or checked EnvironmentSelector | Legacy string means current-only named; typed selector also supports effective policy |
| InventoryQuestion / inventory and service discovery | legacy string | Explicit CurrentEnvironment; current `deployment.environment.name` only |
| DashboardRequest.Scope / generated record widgets | InventoryQuestion | Same current-only named scope |
| measurements.DiscoveryQuestion / native metric definitions | legacy string | Current resource environment key equality; no legacy-key fallback |

Evidence: `exploration/inventory.go` (`InventoryQuestion.Query`, `CompileInventory`),
`exploration/service_discovery.go` (`CompileServiceDiscovery`),
`exploration/dashboard.go` (`DashboardRequest`, `DashboardRecordQuestion`),
`measurements/discovery.go` (`Discover`), `measurements/source.go` (`Compile`).
The service browser starts four independent supporting requests before dashboard
planning: discovery, spans, metric discovery, inventory (`ui/service-browser.mjs`).
This is four HTTP requests, not a count of SQL statements or transactions.

A checked current-only **named nonempty** selection can be encoded as the existing
service scope string without changing meaning. For nonempty values, native metric
map lookup equality also implies presence. Effective named selections cannot be
encoded this way: fallback to `deployment.environment` would be lost. Empty and
unreported selectors must likewise refuse, with a reason. Do not guess a named
scope from a display label. Retain the original typed selector in the handoff.

## Identity and window boundaries

`identity.Namespace` distinguishes wildcard nil, missing, value-empty, and a
literal value. `FromValue(nil)` means missing; pointer-to-empty means value-empty.
All preserve presence through SQL and record predicates. Do not use a truthy
string fallback to choose namespace identity.

Service discovery requires a present string service dimension but allows the
reported empty string. Record equality and dashboard planning permit this
identity. Native metric discovery rejects empty service through Binding.Services
validation (`measurements/model.go`). A bounded first handoff can honestly refuse
empty service instead of presenting a partially unsupported metric overview.
Missing service is never an alias for a reported empty name.

Record and native windows are fixed, half-open, nonnegative and at most 120
minutes. Native validation requires `toMs < 4102444800000`, while records allow
that exact ceiling. Use the common stricter bound for the all-signal destination;
never replace the original read window with Date.now().

## Staged destination contract

A pure checked intent identifies exact reported service and namespace, original
checked environment and fixed read window. It explicitly states that Activity
predicates are not carried to the service overview. A service overview spans
multiple signal families, so copying arbitrary log/span predicates would invent
cross-signal meaning.

Staging performs no discovery, inventory, metric, dashboard or profile reads.
One visible **Read service overview** action enters the existing service browser
pipeline. Back returns to paused Activity before or after reading without a new
source read. Replacement/navigation retires pending work; stale, detached or
inert callbacks cannot read. The existing observed profile selection remains;
no profile is inferred from the source row alone.

## Targeted check

`TestServiceHandoffCurrentNamedPlanParity316` compares complete executable record
plans and parameters for typed current-only versus legacy named scope across both
signals and wildcard/missing/empty/literal namespace selectors. It also asserts
effective policy distinction and the empty-service metric boundary. Go 1.27.2 execution passes (`go test ./exploration -run
TestServiceHandoffCurrentNamedPlanParity316 -count=1`, 0.011 seconds). Six focused
destination ownership tests also pass, including zero-I/O staging/return, exact
scope requests, hidden/inert/stale refusals and cancellation retaining Back.
These checks do not replace the separate native full-service journey.
