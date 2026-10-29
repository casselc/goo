# Record selector API review — 250

Architecture approved a coordinated server slice with explicit unsupported
family/persistence gates and browser adoption deferred. It required checked
compatibility at all pair/member boundaries, detached DTOs, policy-explicit
summaries and no conversion of a scope value into a legacy name. View baselines,
reported-parent ownership, native discovery and XTDB stay unsupported. Pure exact
capture IDs remain references, not new named-entity definitions.

Go review found two defects: record comparison checked semantic equality and then
incorrectly compared the two raw environment strings; immutable capture validation
did not recheck new-selector row membership. Both are corrected, with regression
checks for equivalent current-only wires and correctly rehashed foreign rows.
Context/followup propagation, capability binding and refusal gates were otherwise
sound in the reviewed cut.

Architecture accepted the optional v2 catalog extension because omitted fields
preserve old entry bytes/digests. Tests freeze those bytes, exercise mixed
pagination/rebuild and exact reopen, and ensure legacy filters do not conflate new
selector entries. Rollback documentation states that older readers may reject the
entire mixed catalog, not merely omit new captures.

Frontend found no reachable read/action bypass at the browser gate and requested
an actionable API fallback plus a central display guard. The shared display
boundary now refuses before mounting old owners, with an exact captured JSON
link. Actual browser checks assert no telemetry POST on that path. Linked scope
labels remain passive. Browser authoring is still not implemented.

Reviews were read-only. Executed suites, bounded property/model checks and actual
source journeys are recorded in VALIDATION, without a full accessibility,
production performance or complete A0 claim.
