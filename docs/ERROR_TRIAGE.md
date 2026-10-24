# Recorded error and exception triage

Implemented and independently reviewed after actual source/UI/model acceptance. The Activity entrance stages
ordinary log/span questions in the existing fixed exploration workflow. Counts
measure matching records, not unique requests, exception-event counts, failures,
issues or instrumentation completeness. No manifest or application-specific
annotation is required.

## Explicit evidence rules

The default span rule matches reported status class `error`, a named `exception`
event, or a nonempty span `error.type` attribute. An exception can be handled while
its span reports `Ok`; that record still contains exception evidence. The default
log rule matches either OTel severity number 17–24, one of the exact canonical
`ERROR`, `ERROR2`, `ERROR3`, `ERROR4`, `FATAL`, `FATAL2`, `FATAL3`, `FATAL4` text labels,
or nonempty structured `error.type`/`exception.type`/`exception.message`/
`exception.stacktrace`. Numeric severity and text remain separate reported signals;
conflicts are not reconciled into an inferred application outcome. Arbitrary
message text, casing and prefixes are not used to guess errors.

The rule selector exposes narrower status/severity, type-data and span-event
choices. Exact predicates are inspectable and editable before explicit Run.
Original applied selectors and the last completed Activity window are preserved;
opening, choosing a rule or staging the fixed triage question does not query
telemetry. The stream pauses, and returning restores the same held Activity.

## Span event projection

The physical adapter already supplies parallel event-name, attribute and timestamp
arrays. New spans-only fields preserve the distinction from span attributes:

| Field | Meaning |
| --- | --- |
| `column:exception_event` | `yes`/`no` for an exactly named `exception` event |
| `column:exception_type` | Present only when those events report exactly one distinct `exception.type`; present empty remains distinct |
| `column:exception_type_state` | `unreported`, `empty`, `single` or `multiple` reported type values |

Only matching event names contribute types. Repeated occurrences of the same type
still represent one span record and one reported type; a mixture of empty and
nonempty types is `multiple`. Typed plus untyped events have one reported type,
without claiming every event was typed. There is no arbitrary first-event choice
or array expansion. Type/state projection refuses mismatched name/attribute arrays.

The nullable type group is labeled **No unique reported exception type**, covering
unreported and multiple cases. The state field distinguishes them. It is not an
invented exception type or an issue fingerprint. Nonempty `attribute:error.type`
is an independent producer-reported classification, available as a type-source
preset; it is not silently merged with event types. Logs retain structured
attribute-based exception fields and gain the intrinsic `column:severity_number`.
Derived row values are projected only when requested, preserving old row SQL and
stored capture representations. Existing predicates, grouping, counted lookups,
comparisons and explanations share the new field/presence rules. Caller-held
`/compare` groups with null values for nonnullable fixed columns (for example
`column:service`) are now rejected with HTTP 400; earlier validation admitted
those malformed groups. Missing map keys and nullable derived exception types
remain supported. Valid legacy source captures and their hashes are unchanged.

## One investigation flow

Activity's **Investigate → Errors & exceptions** freezes the completed applied
window. A compact staging panel offers the evidence rule and reported type source,
with custom qualified fields under a disclosure. **Open fixed triage** stages the
ordinary question; its existing Run performs the source read.

The initial full breakdown is service plus selected type, alongside the existing
count timeline and record page. **Inspect group** stages exact service/type
membership. **Compare within service** retains the original error rule and exact
service, grouping only by type so existing baseline/previous-window comparison
controls can compare populations. Neither continuation performs a hidden read.
Source group/cardinality limits and changed-between-read states remain explicit;
no truncated top-N is called complete.

A passive **Exception details** disclosure serves Activity, fixed, captured and
trace record inspectors. It shows supplied named span-event or structured log
exception evidence, UTC event times, original casing, message and expandable
stacktrace. Missing or misaligned arrays remain unavailable/partial; details are
not paired across mismatched arrays. Display limits are disclosed and more supplied
occurrences can be revealed locally. Inspecting these details never reads sources.

The existing record/trace inspector, reusable question packages, immutable captures
and investigation notes preserve the result and explanation. Annotations attach
to exact captured evidence or occurrences; no stable cross-window issue identity,
automatic deduplication, stack fingerprint or resolution lifecycle is inferred.

## Portability and acceptance

The new fields are physical-source projections. Offline Databricks compilation
continues to require declared prepared facts; this does not add raw nested-event
lowering. A future prepared-fact adapter must preserve record grain, exact event
names, alignment refusal and nullable unique-type states before aggregation.

ET1 delivered in `982b9e9`: Go/reference/ClickHouse field parity, intrinsic severity and nullable presence,
legacy row/capture compatibility, protocol/editor/hint support and bounded errors.
ET2 delivered in `0c4ab38`: compact Activity handoff, reusable group continuations and passive exception
details, with no parallel workbench or automatic reads. ET3 delivered after actual acceptance: real ordinary OTel
journeys, full aggregate oracles, malformed/multiple/no-data cases, saved reuse,
annotations, Hegel properties, finite Quint replay/rejecting controls, Bombadil,
reviewed screenshots and synchronized documentation. Review and commit validated
milestones, then push the completed round.

## Storyboard 179

The [executable journey](../journeys/error-triage.spec.js) ingests ordinary OTel
spans with named exception events and logs with numeric/text/structured evidence.
It verifies full record counts beyond the page, exact groups and time intervals,
handled exceptions, group/service staged continuations, trace details, comparison,
reusable questions, captures and investigation notes. Separate supplied-data
[component cases](../journeys/exception-details-components.spec.js) check passive
SSR and SPA composition; they are not simulated OTLP integration evidence.

| Frame | Evidence |
| --- | --- |
| [Span service/type groups](journeys/screenshots/179a-spans-error-evidence-groups.png) | Empty and no-unique types remain distinct; counts measure spans once |
| [Log service/type groups](journeys/screenshots/179a-logs-error-evidence-groups.png) | Explicit severity/structured evidence over log records |
| [Reported span exception](journeys/screenshots/179b-spans-reported-exception-detail.png) | Supplied event type, message and UTC time; stacktrace remains disclosed locally |
| [Reported log exception](journeys/screenshots/179b-logs-reported-exception-detail.png) | Structured attribute evidence through the same passive element |
| [Standalone supplied-data component](journeys/screenshots/179c-passive-exception-details-components.png) | Explicit synthetic SSR/SPA fixture with actual component styles; separate from the real-source frames |

Both signals, source-free stale-owner cases, passive component cases and the
55-second Bombadil campaign pass. Independent Hegel/raw-fact/Quint checks reject
80 changed controls across the two main and ownership ledgers. The existing
Activity explanation mains also pass. See [VALIDATION.md](../VALIDATION.md) for
exact results, corrected supplemental probe history, model scope and publication.
