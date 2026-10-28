# Independent review reconciliation — 236

Six delegated reviews examine `5d12ce9`: five independent domain assessments and
one planning reconciliation. These reports challenge the parent review's proposed
order. This integration records confidence and decisions; it implements no fix.

## Reports

- [Product and usability](PRODUCT_USABILITY_REVIEW_235.md): core tasks, shared
  signal scope, preservation/adoption choices and recurring question reuse.
- [Frontend](FRONTEND_REVIEW_235.md): live inspection handoff, keyboard workload,
  passive component/update contracts and measured-versus-hypothetical performance.
- [Architecture](ARCHITECTURE_REVIEW_235.md): reproduced ingress boundedness,
  producer/catalog limits, persistence boundaries and source-policy questions.
- [Go and semantic libraries](GO_REVIEW_235.md): reproduced malformed-text grouping,
  ingress corroboration, locking/aliasing risks and simple public APIs.
- [Verification/process](VERIFICATION_REVIEW_235.md): required verdicts, test-tier
  coverage, deployment identity and targeted acceptance rather than test quotas.
- [Planning reconciliation](ROADMAP_RECONCILIATION_235.md): execution order,
  dependencies and stale-authority corrections after reading all five reports.

## Findings that change the queue

| Finding | Evidence/confidence | Decision |
| --- | --- | --- |
| Incomplete JSON upload retains admission after explicit cancellation | Architecture reproduced actual budget/decoder behavior over a local socket: cancel 202, next request 429 until peer close, then 200. Go independently reproduced the blocking-read mechanism. Body.Close also remained blocked. The full 60-second ceiling was not waited out in the architecture probe. | Early bounded ingress fix with real network regression; context cancellation alone is insufficient. Preserve SSE and publication lifetimes. |
| Switching Logs → Spans drops a valid service filter and immediately reads broader scope | Product captured both real tail request bodies; frontend cross-checked source. | Fix applied-scope continuity before cosmetic density. Preserve compatible predicates; surface incompatible ones before broadening. |
| Mobile Inspect opens details off-screen without moving focus | Frontend keyboard probe at 390×844: inspector starts at y901.09; focus remains Inspect, no opener controls/expanded relationship. | First Activity usability gate includes visible/accessibly announced inspection and reliable close/return. |
| Browser ledger helper accepts zero matching Go tests | Verification deliberately selected a nonexistent name and observed exit 0; helper checks only process success. Its static scan found no currently missing literal validator declarations. | Require named pass/no required skips. This is a verified guard defect, not evidence all previous runs were false green. |
| Malformed prepared Go strings merge distinct groups | Go public API probe with label bytes `0xff`/`0xfe` returned nil error, one group,count 2. Normal browser JSON cannot carry arbitrary invalid Go bytes. | Bounded UTF-8 contract correction, preserving valid Unicode, empty values and historical visibility; not generic hash cleanup. |
| Native package publication can exceed the catalog's 256-file bound | Architecture inspected producer and reader contracts; no endpoint/lab publication reproduction was performed. | Resolve capacity semantics before expanding revision-producing features; preserve duplicate retries, concurrency, existing IDs and corruption refusal. |

## Product decisions, not accidental defaults

- **Missing environment:** ordinary OTel need not contain
  `deployment.environment.name`, but current record/native scope requires a named
  exact environment. Decide explicit enrichment/diagnostics or a checked selector
  policy. Never turn missing into a silent all-environment wildcard or imply an
  empty scoped result proves the source is empty.
- **Keep observed items:** exact live occurrence preservation is currently optional
  retention plus an annotation; general Capture honestly rereads fixed scope.
  Decide whether core bookmarking should require that capability/interpretation.
  Preserve server-attested membership and immutable bytes. Do not call a reread
  the original selection or make XTDB mandatory by default.
- **Preservation vocabulary:** questions, fixed captures, exact selections and
  investigation documents have different effects. Clarify outcomes/destinations
  without collapsing them into an ambiguous generic Save operation.
- **Comparison support:** show support first and keep rare failures visible by
  default. Threshold suppression is optional and descriptive, not significance.

These decisions should be resolved with named journeys alongside bounded fixes;
they need not block capture discovery behind a generalized selector/store rewrite.

## Accepted constraints and deferred proposals

Capture search remains the next broad product feature. Its metadata path must not
load every complete 24 MiB capture in a 10,000-entry catalog. Decide list versus
reopen/rebuild integrity guarantees explicitly; metadata-only listing cannot also
promise to detect arbitrary later payload corruption without reading it.

Source-verified follow-ups include snapshot-bank locking across filesystem
publication, mutable input aliasing in compiled plan metadata, plot keyboard
navigation and property-update focus/disclosure stability. These have useful
bounded acceptance in the reports; no production stall/OOM/full accessibility
failure was reproduced for them. Source-policy differences deserve a targeted
conformance probe before claiming continued ClickHouse work after cancellation.

The reviewers agree on preserving typed OTel/numeric semantics, passive web
components, concrete source owners, immutable/replay contracts and separate tool
modules. They do not recommend a framework migration, universal store/history/query
layer or reopening the completed package/reference cleanup. They differ on density
versus capture-discovery ordering; the active roadmap makes that tradeoff explicit.

## Integration and limits

The [active roadmap](../ROADMAP_2026_10_28.md) owns execution order and
[backlog](../BACKLOG_2026_10_28.md) owns bounded acceptance. Existing 217–234 delivery
records remain valid for their stated scopes; these new counterexamples add gates
at uncovered seams. Default unit counts do not imply full browser/native/model
coverage. Keep fresh-lab and deployed-asset identity limitations explicit.

Review probes used synthetic local sockets/Go inputs or read-only existing local
telemetry. No workload, capture, annotation or document was published. No runtime
code or dependencies changed in this review round. This is expert review plus
selected reproductions, not a production load/security/accessibility certification.
