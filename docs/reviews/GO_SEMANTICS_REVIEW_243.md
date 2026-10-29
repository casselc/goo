# Independent Go and semantic-library review — 243

Reviewed 2026-10-29 at `65ac193`. This is an independent source review and bounded
pure-Go/temporary-store HTTP experiments. It covers semantic/compiler ownership,
native/source boundaries, capture persistence, HTTP lifetime and instrumentation.
No broad suite, load benchmark, live telemetry/evidence mutation, runtime fix or
new dependency was introduced. Only this report is a deliverable.

## Recommendation

Preserve the completed semantic layers and single Go runtime. Fix the newly
reproduced accepted-capture/index mismatch before treating saved search as a
closed reliability gate. Keep I1's malformed-text correction small and explicit.
Detach compiled-plan metadata when that compiler is next touched; isolate snapshot
publication I/O separately. None of these requires a universal planner/store,
framework migration or advanced type parameters in user-facing APIs.

## G1 — Accepted capture can make the entire metadata catalog unavailable

**New, reproduced through actual HTTP handlers; high confidence.**

The capture reader permits 32 predicate nodes and string values of up to 1,024
bytes. The capture POST accepts 16 KiB. Derived metadata instead stores the entire
receipt/question in a 64 KiB envelope. Go JSON encoding expands `<` into `\u003c`.
These limits are not aligned:

- A valid log question with 31 AND leaves, each `column:service eq
  strings.Repeat("<",340)`, has a 12,728-byte request when encoded without HTML
  escaping, as browser `JSON.stringify` does.
- `POST /captures` returns 201. The receipt alone marshals to 65,557 bytes, before
  the additional metadata envelope, exceeding 64 KiB.
- `publish` has already committed and verified the capture, then deliberately
  ignores `indexCapture`'s error. No metadata entry is written.
- Search refuses any missing entry across the admitted catalog. Consequently
  `GET /captures/search` returns 503 even for an unrelated/no-match filter.
- Exact-ID reopen returns 200. Rebuild indexes 0 and returns the same search-budget
  error, so its recommended recovery cannot repair this valid capture.

Observed result from a fresh owned store and a loopback source returning no rows:

```json
{"captureStatus":201,"rawRequestBytes":12728,"receiptBytes":65557,
 "searchStatus":503,"reopenStatus":200,"rebuildIndexed":0,
 "rebuildError":"capture search scan budget exceeded; open a capture by ID instead"}
```

The HTTP experiment used a Go build overlay to run the existing
`internal/capturehttp.Register` handlers, not a mock publication implementation.
The fake source supplies an ordinary zero-count result; no live source/store was
used. A separate public `CaptureStore.Create` experiment has the same outcome.

Relevant source: [predicate bounds](../../exploration/model.go),
[HTTP bound](../../internal/capturehttp/policy.go),
[full-receipt metadata and cap](../../exploration/capture_catalog_index.go),
[post-commit best-effort metadata](../../exploration/capture_store.go),
[whole-catalog missing-entry refusal](../../exploration/capture_search.go).
This is an availability/contract defect, not evidence of payload corruption or a
security incident. The small input is synthetic but valid under the public API.

**Preferred bounded response:** project a dedicated, size-bounded catalog summary
from a verified receipt: immutable ID, title, captured time, parent, kind/subtype
and the existing declared scope. Search/render need not retain/revalidate every
complete query/definition. Derive scope through `CaptureQuestion.Scope`; do not
invent a second OTel inference engine. Version/rebuild only the derived index and
update the explicit search wire model if needed; immutable capture bytes and IDs
must stay unchanged. Keep the existing metadata-versus-content integrity boundary.

Alternatively, prove indexability before source work and immutable publication.
Do not merely return a failed-save error after committing, silently skip the
capture, increase caps without an aggregate budget, or switch search back to full
payload scans. Existing accepted captures still need a bounded repair path.

Acceptance: reproduce the 12,728-byte case through HTTP; successful capture remains
findable alongside earlier captures, and rebuild restores it. Cover escaped Unicode
and HTML, long comparison/native definitions, failed metadata I/O, duplicate retries,
concurrent indexing and stable immutable payload hashes. Keep query, per-entry and
retained-memory bounds explicit. The ordinary small-store 242 tests remain valid;
they did not cover this relationship between accepted input and derived output.

## G2 — I1 remains open across facts and authored literals

**Independently reproduced public Go API defect; high confidence.**

At [Evaluate](../../subjectfacts/evaluate.go), string fields are type-checked but
not UTF-8-checked before subject identity and group keys use `json.Marshal`.
Two selected facts with valid distinct subject IDs and label bytes `0xff`/`0xfe`
return nil error, one group and count 2. Different malformed strings normalize to
U+FFFD in the identity encoding. This is not a claim that normal browser JSON can
carry arbitrary malformed Go bytes.

The authored boundary has the same hole: a `Population` string literal containing
`0xff` passes `Definition.Validate`, as do enum values containing distinct `0xff`
and `0xfe` strings. See [predicate and enum checks](../../subjectfacts/model.go).

**Bounded response:** reject malformed text, rather than replacing it, at accepted
prepared-fact and authored-literal boundaries. Include the execution evaluation ID
and selected facts before population filtering. Preserve existing visibility:
validly identified facts in a different evaluation are not evaluated; an invalid
selected row cannot be concealed by a population predicate. Audit subject keys,
dimensions, predicates and enums consistently.

Acceptance: distinct bad bytes fail; valid U+FFFD, non-ASCII strings and permitted
empty values retain their existing semantics. Add direct public API controls,
including selected/excluded evaluations and population-excluded bad rows. Keep
existing definition outputs, compiler fixtures and valid identities unchanged.
This remains a finite correctness cut, not a hashing/identity redesign.

## G3 — Compiled plan metadata still borrows the mutable authoring definition

**Reproduced public API alias; high confidence, ordinary usage risk.**

[Compile](../../subjectfacts/compile.go) assigns `Plan.Definition = d`, while
query/metric-view text is generated immediately. I compiled a definition bound to
`facts`, then changed its original binding to `other_facts` and dimension mapping
from `label` to `id`. The plan's SQL still uses the original source and mapping,
but its metadata reports the edited values; the Definition pointer is identical.

No deployed serving failure was reproduced. The compiler explicitly returns a
prepared-source artifact, so do not inflate this into a current authorization
exploit. It is nevertheless a poor default for a reusable library whose result
is described as an inspectable plan.

**Bounded response:** retain a detached definition snapshot in the compiled plan.
Use concrete copying: `maps.Clone` for flat maps, `slices.Clone` for slice values,
and explicit copies of predicates/measure fields and nested enum slices. A shallow
struct or top-level map copy is insufficient. No reflection-based generic clone,
caller-facing ownership type parameter or mandatory builder is necessary.

Acceptance: compile for ClickHouse and Databricks; mutate original bindings,
field/dimension maps, predicates, enum and unknown-condition slices; plan text and
metadata remain aligned. Document result ownership. Existing checked native
revision/question-reference APIs should not be reopened for this fix.

## G4 — Snapshot-bank mutex still covers filesystem publication

**Source-verified serialization risk; no production stall/latency claim.**

[PromoteExplanation](../../exploration/result_snapshots.go) holds the bank-wide
mutex across existing-ID `Load` and the publisher. The publisher now also performs
capture verification and derived metadata write/sync/rename/directory-sync. Thus
one slow filesystem operation can delay unrelated Freeze/Load/promote operations.
The new index does not cause the original problem, but adds work under that lock.

Use a keyed in-flight reservation and publish outside the bank-wide critical
section. Preserve reserved custody time, retry candidate identity, uncertain-commit
semantics and bounded entry/promotion lifecycle. Expiry must not discard an active
reservation or make a duplicate publish another custody candidate. Avoid a new
general job system or blanket asynchronous publication.

Acceptance: a barrier publisher holds save A while unrelated Freeze/Load B completes;
duplicate A remains idempotent and receives the same candidate/outcome. Cover
failure after immutable publication, expiry and cancellation at the defined mixed
publication boundary. This should be its own reliability cut.

## What to preserve and what to simplify

| Boundary | Assessment / targeted follow-up |
|---|---|
| `subjectfacts` → prepared facts; `lifecycle` → preparation; `metrics` → numeric kernels; `measurements` → native binding/evaluation | Keep these distinct. A shared SQL planner is not a replacement for their different grains, temporality, coverage and custody contracts. |
| Native storage-free evaluation and explicit lifecycle observation | Preserve concrete checked definitions/questions and explicit publication. Do not make ordinary users manipulate reflection or type machinery. |
| Instrumentation `Runtime`, `atomic.Pointer`, `sync.Once` | Purposeful ownership and one-time completion. Keep explicit providers/binding/flush and finite names/roles. The bounded initialization map is not a full metric-cardinality guarantee. |
| HTTP family policies and single Go composition | The actual transport read deadline, joined `context.AfterFunc`, `ResponseController` and shutdown/error-join mechanisms are appropriate modern APIs. Earlier ingress/proxy findings are delivered/retired, not current blockers. |
| Metadata query kernel and typed source adapters | Reusing declared scope and `slices.SortFunc` is good. Preserve simple query/page types and exact clocks; fix G1 through a bounded projection rather than a new database abstraction. |
| Byte decoding | `capture_store.go` and `capture_catalog_index.go` still create a full string copy before JSON decoding. `bytes.NewReader(b)` removes that copy mechanically when touching these loaders. No benchmark here establishes overall memory/latency benefit. |
| Bounded metadata rebuild | Its 256 MiB whole-store preflight is honestly documented but not incremental. A valid larger store cannot repair a missing index with the only provided command. Plan a resumable bounded batch/range workflow before claiming broader catalog scalability; do not remove the limit or delete captures. |
| Native package capacity | The prior producer/reader 256 mismatch remains visible in `cmd/goo/measurements_http.go`. Coordinate its semantics with architecture review before more revision-producing features; preserve existing files and duplicate retries. This review did not separately publish 257 native packages. |

The appropriate recent-Go work is small and internal: concrete map/slice ownership,
zero-copy readers where possible, named deterministic helpers, and tested
concurrency boundaries. Replacing clear goroutine/channel lifetimes with
`WaitGroup.Go`, introducing generic registries or changing public APIs merely
because the Go baseline allows it would not solve these findings.

## Scope and limits

Reviewed current storage/search contracts, source and native evaluation boundaries,
prepared model/compiler/evaluator, result snapshots, HTTP receive/shutdown and
instrumentation ownership; reconciled earlier 235/236 findings and 242 delivery.
Small probes establish G1–G3 above. No broad root/race/native suite, collector lab,
performance benchmark, security audit or new Databricks-instance execution was run.
Only owned temporary directories, loopback fixture servers and build overlays were
used; all probe directories were removed. Runtime/library files remain unchanged.
