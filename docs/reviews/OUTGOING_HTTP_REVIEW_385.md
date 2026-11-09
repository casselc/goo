# Outgoing HTTP review — 385

Read-only architecture/usability review of the milestone 385 source candidate,
November 9, 2026. No heavy tests, source queries or browser campaign were run by
this reviewer. Actual checks belong in [validation](../../VALIDATION.md).

## Scope and population

`exploration/dashboard.go` adds an explicit `http-client` family. Current
`attribute:http.request.method` metadata suggests a candidate; it does not prove
Client records. Every executable client widget applies exact reported `Client`
kind and a present nonempty current method to the established service, namespace,
environment-policy and fixed-window question. `_OTHER`, case and custom nonempty
values remain reported strings; this is an observation policy, not semantic
convention certification.

The five questions cover span-name activity and qualified destination, method,
port and response-status groups. Optional destination/context fields do not
restrict admission, so missing/empty groups remain possible. Full counts retain
record grain and duplicate observations; they do not establish unique requests,
physical attempts, callee-service identity or remote health. Recorded span status
is separate from HTTP response status. The official reference is the pinned
[OTel HTTP client convention](../HTTP_CLIENT_CONVENTION_385.md).

Existing auto-profile precedence is unchanged. The client family does not enter
that ordering or grant a native HTTP-client metric credential. Generic native
instrument availability remains independent per-stream evidence. The shared
Server/Client predicate helper is a bounded internal reuse; a new registry or
execution engine is unnecessary.

## Presentation and interaction

`ui/service-dashboard.mjs` gives population and destinations primary placement,
with method/port/response cards in one native context disclosure. Pending,
unavailable and changed-read counts remain visible in its summary. The existing
two-wide explicit overview queue acquires results; disclosure changes read no
telemetry. Current-owner/visibility guards stop hidden, inert or retired group
callbacks before the global investigation trail can follow them.

Literal port/status labels use `Distributions.valueLabel`, which quotes reported
strings rather than parsing or normalizing them. Values such as `443`, `0443`,
`443 `, invalid text and out-of-range text stay different from each other and from
missing/present-empty values. `diagnosisQuestion` preserves the exact qualified
value or missing predicate in the full parent question.

## Display grammar and evidence

The ordinary dataset builder in `exploration/view_dataset.go` retains the full
Query in `Scope.Selector`; it has no family mapping that could turn a client
population into a server population. The rich-record page boundary likewise
retains and validates the whole question. Generic views, question cards and
captures therefore preserve the exact client selector without a new authored
HTTP-client template category.

`ui/view-template-handoff.mjs` admits shipped templates by exact family and
qualified grouping. A shared method key alone cannot offer the HTTP-server route
template to the client lens. Ordinary explicit configuration reuse remains
available under its existing destination-authority rules.

## Required trace-expansion clarification

Review found and closed one user-visible scope gap before build. The established
`ui/record-results.mjs` whole-trace helper deliberately drops service, namespace
and other population/group predicates, keeping environment, bounds and reported
trace ID. Its former “across services”/fixed-bounds text did not name namespace
omission.

The inspector now has a persistent visible `exploreTraceScope` paragraph populated
before the read starts. It shows the reported ID, checked environment, exact UTC
bounds/end exclusion, omitted service/namespace/population predicates and the
limits on invocation/completeness/causality claims. Loading, refusal, retry and
successful result messages update separate status nodes, so they cannot overwrite
this scope statement. The mounted cancellation/refusal test checks persistence
through those transitions. The feature contract now names namespace expansion
explicitly. This adds no query policy or navigation manager.

No remaining source-review blocker was identified in the final candidate.

Runtime acceptance remains pending at this review checkpoint. Required proof is
independent raw Client/Server membership, missing/empty and literal dimensions,
explicit empty/refused client-profile retention, bounded full counts, compact
service-to-trace frames, exact parent/group versus broader trace requests,
immutable save/capture and no-read held Back. Source review alone does not prove
those outcomes.


## Actual-runtime follow-up

Reviewed the final `artifacts/http-client385/summary.json`, browser driver and
journey/feature documentation after the parent reported the browser run passed.
The report contains 44 API calls, no page errors, 89 captured SQL calls and an
Bombadil campaign with 554 observed states and a successful process exit.
Its recorded completed phase is not sufficient acceptance: the parent found its
final frame still showed Explore, so independently visible Services completion
requires a strengthened actor gate and rerun. The exact retained trace notice names the omitted
namespace/population filters. The browser driver independently checks the 117
matching Client records against physical source rows, a 100-row inspection page,
117 retained capture records and unchanged stored capture bytes on source-refused
reopen. These are bounded local-workload outcomes, not general ingestion or trace
completeness guarantees.

The actual return gesture is **Services** workspace navigation after closing the
trace inspector, not the dedicated contextual **Back to service** action. The
runtime driver checks retained widget DOM/evidence and zero additional telemetry
reads on this return, with focus on `workspaceTitle`. Both the feature document
and storyboard now state this distinction; no contextual-Back result is inferred
from this journey. Consolidating those return affordances remains a usability
opportunity. Screenshots deliberately scroll the destination table into view;
no above-the-fold fit or broader viewport claim is made by this review.

The parent separately reports 44 response and 24 request schema validations.
Those counts are not fields in the browser summary and are not independently
re-run here. The verification owner subsequently reports the required-Quint main-browser
ledger passed with zero skips: 13 facts, 23 measured model steps, ten semantic
mutants and the passive-read negative model control rejected. This is attributed
to that separate verification run; this reviewer did not rerun it. The actor is
deliberately excluded from that ledger and is not covered by its pass. No further documentation mismatch was found. Actor completion acceptance remains
pending the stronger visible-workspace gate and fresh campaign.


### Strengthened actor acceptance

The initial 554-state campaign's premature completion was rejected, not accepted
as proof of return. The strengthened full journey subsequently passed. Its fresh
report records 552 actor states, a distinct `returning` phase and completion only
with `returned=true`, `preserved=true`, `activeWorkspace=services`, `sourceSafe=true`
and `returnSafe=true`. The parent also personally inspected the final actor frame
and confirmed the actual Outgoing HTTP dashboard was painted, showing 117 records
and nine recorded errors. This resolves the actor visible-return acceptance gap.
The review did not independently inspect that image or rerun the browser.

The prior ledger pass applies to the previous main-browser report. A ledger
recheck against this refreshed report was still pending at this final checkpoint;
the stronger actor result does not substitute for that separate check.

Final verification follow-up: the verification owner reran the required Quint
ledger against the fresh strengthened-campaign report: **3.668 s**, zero skips,
13 facts/23 steps, all ten semantic mutants and the passive-read model control
rejected. This closes the final-report replay gate; it adds no actor booleans to
the measured ledger.
