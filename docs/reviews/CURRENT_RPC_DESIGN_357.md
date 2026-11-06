# Current RPC observation profile — design review 357

Status: pinned source and all three current actual-frame reviews accepted.
Execution results remain with native/validation owners. 358 remains the selected
presentation consolidation; no new component/product queue is added.

## Keep two interpretations explicit

The existing `rpc-server` overview is a deliberately legacy lens using reported
Server kind, nonempty `rpc.system` and `rpc.method`, with separate native legacy
millisecond duration. Current conventions must have their own checked profile and
ruleset rather than treating old/new aliases as interchangeable.

Prefer **Incoming RPC** for the new current profile and **Incoming RPC (legacy)**
for the older lens, or explicit current/legacy qualifiers together if needed to
avoid ambiguity. Existing current All spans, HTTP and other service views remain
recognizable. The legacy profile is useful existing behavior, not a compatibility
alias for the new interpretation.

## Candidate, observation and explicit reading

A flat attribute-key inventory may make a profile selectable but does not prove
that required current attributes occurred on a Server record. Auto selection needs
the agreed positive same-row scoped witness or an eligible exact native instrument
under the pinned rule. The architecture contract decides priority when current,
legacy and HTTP evidence coexist; the UI must not silently alter that priority.
A bounded held page without a witness means unknown, not absent RPC work.

Changing profile stages intent; existing Apply owns its independent reads. An
empty current profile stays no-data with an explicit All spans fallback. It must
not silently broaden or reuse legacy rows to make a dashboard look populated.
A current native-only candidate can show its native data while span observations
remain not observed. Discovery hints do not certify an instrumentation schema.

## Population and labels

Every record widget for the current profile uses the same exact service/namespace,
typed environment policy, fixed UTC window, reported role and pinned current-key
presence/value rules. Missing/empty/legacy-only contrasts remain visible in tests.
Method grouping is a literal dimension within this parent population, not a claim
that a method name uniquely identifies an RPC operation across systems/services.

Use **RPC server-span records**, **Recorded span Error** and **Mean duration**
where appropriate. Span Error is not an inferred protocol/gRPC status policy,
business failure or request uniqueness. Raw span-record rate is not deduplicated
RPC call throughput. Durations are observed span durations, not end-to-end request
latency or complete traces.

Current native duration uses the pinned seconds instrument and keeps its exact
stream/library/schema/unit/temporality identity. Legacy millisecond instruments
remain separate. Do not convert units to claim profile eligibility, merge fleet
percentiles or blend native counts with span counts. Human display formatting can
choose readable time units while source unit/evidence remains inspectable.

## Follow-up and ownership

A method/operation continuation adds its exact literal predicate to the whole
current parent question and stages before Run. It keeps role/current conventions,
service/namespace/policy/window and actual operand meaning. Customize/templates
only appear for profiles their existing checked contract admits; the new RPC lens
must not imply an HTTP/DB template is now compatible.

Profile/range/service changes retire departed widget work under the existing
354 owner. Retry stays local to the failed current question while successful
siblings retain their clocks. Unsupported prerequisites remain explanatory, not
retryable source outages. Back restores the held service with current focus and
no hidden reads.

## Acceptance

Pin official conventions before source changes. Exercise current-only, legacy-only,
mixed, client-only, missing/empty-current-key and native-only cases, including
current versus legacy native seconds/milliseconds. Verify candidate versus positive
same-row hint and exact automatic precedence. Include explicit current selection
with a zero result and manual All spans fallback.

Native UI should demonstrate scoped service entry, current/legacy labels, explicit
Apply, exact native stream, method continuation and retained Back. Actual desktop/
narrow frames establish readability and no overflow; source/generated tests prove
population membership, field-version/unit eligibility and scope. No claim of full
RPC RED, protocol outcome inference or OTel conformance follows from this slice.

## Closed pin, labels and frozen source acceptance

Architecture pins semantic-conventions v1.45.0 at commit
`fb30ca17193b551473b1f43c315482918ab6987c` in the official convention research.
The closed current population requires exact Server kind and nonempty
`rpc.system.name`; `rpc.method` is conditional and is grouped without excluding
missing or empty values. `_OTHER` remains a reported literal, never synthesized.
The native instrument is exactly `rpc.server.call.duration`, histogram in seconds;
legacy `rpc.server.duration` milliseconds remains independent.

Final agreed visible labels are **Current RPC** and **Legacy RPC**, replacing the
initial Incoming-label suggestion above. Auto precedence is HTTP, current RPC,
legacy RPC, then existing lower-priority profiles. Flat key metadata is candidate
only; the current same-row witness checks original question/service/namespace/
policy/window plus role and current system. Native instrument recognition does not
claim a span witness or full convention compliance. Source review accepts these
bounded semantics and existing explicit Apply/fallback/retry ownership.

## Actual desktop/narrow acceptance

Viewed `357-current-refusal.png`, `357-current-rpc.png` and
`357-methods-390.png` at original detail. Current RPC and Legacy RPC are visibly
distinct and All spans remains available. Failed current span acquisition stays
current/unavailable with concise refusal and Retry, while independent method
observations remain visible. Recovery shows five server-span records, recorded
Error 1/5 and mean 3 ms without implying protocol outcomes. Method rows preserve
missing key, present empty string, `_OTHER` and literal names as separate groups.

The narrow methods surface remains in its existing bounded horizontal table
viewport; not all statistics/actions are visible at once, but the document does
not overflow and labels remain readable. This is a known existing presentation
limit, not a 357 blocker or an unrequested layout expansion. The native seconds
instrument title is explicit even where its own read is unavailable. All three
current frames pass visual acceptance. Candidate/native-only/no-data/precedence
and exact source membership are generated/native execution assertions, not facts
inferred from these task-positioned frames. Execution gates remain separate.

## Canonical final refreshed frames

Re-viewed the canonical refreshed `357-current-rpc.png` and
`357-methods-390.png` at original detail after the native fixture's exact streams
were corrected. The overview now shows successful separate
`rpc.server.call.duration [s]` interval series, each labelled observations per
interval, alongside the five current server-span records. Overall text correctly
returns to independent reads and unknown coverage/health rather than retaining a
failed-native summary. Narrow still keeps the seconds-instrument identity clear.
The prior unavailable-native frame observation describes the earlier fixture, not
these final source results. Final visual acceptance remains passed.

Canonical native/actor/model results and exact delta/cumulative interval counts are
recorded in [validation](../../VALIDATION.md). Maintained Playwright captures are
the visual authority; repeated stale actor JPEGs do not establish current or terminal
appearance. No production layout change was required for the fixture correction.
