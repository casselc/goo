# Exception-group review — 416

Status: accepted for this bounded first slice. This review is independent of the
Go source and application UI authors. No blocking findings remain.

## Source and identity contract

The versioned profile identifies exact **reported type + stack text**, with byte
length framing and an explicit SHA-256 domain. It does not normalize frames,
addresses, goroutine identifiers, messages embedded in stack text or line numbers.
Identical call sites can therefore fragment, and matching signatures are not
portable issue identity across languages/builds. A future normalization profile
requires source examples and a separately versioned contract.

The span adapter keeps type and stack from the same named `exception` event. It
classifies one input record only when all named events contain one identical,
complete tuple. Repeated identical events still contribute one span record;
crossed/incomplete/conflicting tuples remain unclassified. Missing and present
empty log values are distinct. Malformed source-array alignment refuses the read.
Bounds and invalid UTF-8 have explicit states; newline/tab/NUL text stays literal.

First/last refer to matching **record timestamps in the searched window** (span
start/log timestamp), not exception-event time, lifetime first seen or a complete
upstream occurrence history. Qualified contexts preserve namespace/version
presence and empty values. The deterministic bounded prefix has an exact distinct
count and limitation flag; it is not inferred from retained examples.

The new carrier uses two source reads with independent receipts. Counts conserve
classified and unclassified groups. Disagreement leaves diagnostic data and
withdraws ranking/drill authority. More than 1,000 groups or the response byte
budget refuses success. Grouped type/stack minima/maxima are checked against the
fingerprint before inputs are admitted. The source response and UI share an
8 MiB bound.

## Interaction review

The existing Errors & exceptions owner stages the fixed population; the new
application adapter owns explicit metadata reads and local presentation. Standard
`exception.type`/`exception.stacktrace` inputs do not silently inherit a custom
triage type field. Rule/parent/close changes retire callbacks and pending reads.
Stage must retain the full selector, typed environment and exact window; a
predicate-budget refusal must never weaken those constraints.

Two usability refinements were requested during review: Find must match displayed
human reason labels, and same-type/different-stack groups need a visible literal
stack cue. These are presentation details, not normalization or identity changes.
Exact signature inputs and source clocks belong in native details, with source
uncertainty still visible. No issue lifecycle/store or source-write operation is
introduced.

The requested refinements are implemented. Regenerated 1440, 390 and 320 pixel
frames were inspected: literal type/stack case is preserved, same-type groups
have distinct stack/hash cues, UTC timestamps remain readable, bounded context
counts are explicit, and mobile cards contain the content. Native disclosure
focus remains visible. Existing triage setup text is verbose, but does not block
this slice; broader control consolidation belongs to a separate product pass.

## Verification scope

The isolated `verification/exceptiongroups` package includes an independent
record-classification/hash oracle, actual JavaScript scope/DTO/stage projection,
Go query validation, a readonly actual-ClickHouse SQL differential, and native
Quint protocol/actual-ledger replay with rejecting controls. The differential
rewrites only the compiler table reference to constant synthetic rows, executing
production SQL and its strict decoder. No source table is created or modified.

Final commands/results and actual screenshots belong in [validation](../../VALIDATION.md).
The actual browser report passed both log and span flows with 24 ledger facts and
no page errors, including failed/pending read retirement and exact group drill.
The accompanying 45 second Bombadil campaign passed. The parent executed the
isolated native suite successfully in 19.680 seconds with no skips: 20 actual
ClickHouse cases/40 reads, 32 identity-oracle seeds, 32 seeds across six JavaScript
and Go scope projections, and Quint protocol/observed-ledger replay with negative
controls. The observed fixture demonstrates 144 matching group records versus
100 retained drill rows and 36 qualified contexts versus the first 32 displayed.

A passing native model/property family does not establish full UI accessibility
or upstream telemetry completeness. Browser review is bounded to exercised
Chromium surfaces; screen-reader/cross-engine certification is not claimed.
