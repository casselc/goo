# Investigate recurring exception signatures

Executable: [exception-groups.spec.js](../../journeys/exception-groups.spec.js).
The journey uses ordinary OTLP logs and spans, an isolated Go application, real
ClickHouse reads, and an independent raw-row oracle. Validation records actual
commands and outcomes; this document describes the executable contract.

An operator finds repeated exception evidence in Activity and wants to identify
a useful population, inspect its exact inputs, and open its records without
rebuilding the source scope. Fingerprints identify exact reported type/stack
pairs under the pinned v1 framing. They are not issues, failures, root causes or
cross-deployment identity claims.

| Step | Interaction | Observable check |
| --- | --- | --- |
| Observe | Start an explicitly scoped saved Activity view, then pause. | A bounded 100-row buffer is independent from the later full population count. |
| Choose | Investigate → Errors & exceptions → Exception groups. | Opening stages a fixed window; it performs no new source read. |
| Measure | Read exception groups. | One request and two SQL reads return conserved classified and unclassified groups. The oracle checks exact fingerprints, counts, first/last record times, and affected qualified contexts. |
| Understand | Find, reorder, or open Signature details. | Local actions use supplied data. Exact type/stack inputs, identity and bounded-context meaning stay inspectable. |
| Drill | Stage group records. | The existing fixed editor receives the original environment/window/filter plus the exact fingerprint predicate, with no records borrowed and no source read. |
| Execute | Run the staged question. | The source returns 144 matching records, with a 100-row inspection page. Timestamp/window membership and fingerprint membership are independently checked. |
| Return | Return to paused activity. | The same accepted group facts return without a source read. |
| Recover | Refresh after a source refusal. | Failure withdraws prior actions and receipts; recovery is explicit. |
| Depart | Close triage while a read is held. | The attempted count read and the source-free retirement are recorded separately. A released late response cannot reopen the retired view. |

The banks include exact duplicate span-event tuples, conflicting crossed pairs,
complete plus incomplete events, unrelated event names, missing/empty fields,
UTF-8 inputs, framing delimiter collisions, oversize fields and event budgets.
Repeated identical exception events count once per span record. First/last uses
the matching record time, not an exception-event clock. Qualified service,
namespace and version presence remain distinct; a classified population with 36
contexts exercises the first-32 bound and explicit limited state.

Current-key/legacy environment precedence and out-of-window records are explicit
controls. A similarly named or partial event cannot fabricate a type/stack pair.
The raw-row oracle implements framing and classification independently of both
Go and the browser module.

The executable captures actual 1440px, 390px and 320px frames at
`docs/journeys/screenshots/exception-groups-{1440,390,320}.png`. Exact timestamps
remain available as UTC detail/tooltips; operator-facing cells do not expose raw
nanosecond arithmetic.

With `GOO_EXCEPTION_GROUPS_BOMBADIL=1`, a separate fresh bank drives a bounded
45-second Activity → grouping → explicit record read → held-return campaign.
The actor checks completion, the full 144-record count, bounded 100-row result,
source-free staging/return, no publication, and layout containment. Its run and
trace stay beside the main ignored `artifacts/exception-groups416/report.json`.
The test stops its runtime in `finally`, preserving the existing telemetry stack
and stored evidence.

## Recorded acceptance

The final real-source journey passes for both signals: **185 logs in 13 groups**
and **191 spans in 15 groups**. Each exact signature drill returns the full
**144-record** count with a separately bounded **100-row** inspection page.
All **24** recorded action facts preserve their declared source budgets. The
classified population's **36 qualified contexts** returns the exact sorted first
32 with an explicit limited flag. Held return restores the original row's focus.

The 45-second actor completes with **615 snapshots**, zero page errors and no
forbidden publication. The accompanying native model/property suite passes with
no skips; commands and its execution times are recorded in
[validation](../../VALIDATION.md). Final frames also preserve literal type casing
in table row headers, including narrow layouts.

The presentation refinement keeps Population options closed initially. The journey
opens and closes it with the keyboard, checks control visibility and zero source
reads, checks that a folded breakdown action cannot stage a question, and checks a native
share meter before opening exact signature details.
The compact header, toolbar and mobile cards use the same captured frames above.
