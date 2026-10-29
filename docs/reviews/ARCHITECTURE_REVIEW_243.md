# Architecture and reliability review — 243

Reviewed 2026-10-29 against `65ac193`, after the single-Go runtime and the first
saved-capture search delivery. This is an independent source review of the
architecture, persistence and lifetime contracts. It also uses explicitly
attributed reproduction results from the simultaneous Go and root task reviews.
It is not a production load, security or filesystem-crash certification.

Only this report was added. No runtime implementation, dependency or live data
was changed. The one local probe used disposable sparse files and removed them.
No Docker rebuild or broad test suite was run by this reviewer.

## Recommendation

Keep the current module/runtime structure. There is no reason to reopen the
completed namespace/reference cleanup, add a query engine, introduce a universal
store, or replace the passive frontend component model. The largest architectural
risk now is inconsistent admission and recovery across a writer and its catalog.
The new search index has a concrete instance of that problem, alongside the
already known native-definition capacity mismatch.

Before adding more revision-producing features, make an accepted write compatible
with its discovery path. Preserve exact-ID reopen and already committed bytes;
do not turn an index failure after commit into a false failed-save response.
Keep the trace-sheet parent-return contract and the direct-Go malformed-string
finding in the finite queue. Normal Close trace → Return does work; interception
of an occluded background button alone is not a pointer dead end. Environment adoption and exact preservation remain
product contracts to settle, rather than reasons to start a generalized storage
or selector rewrite.

## Findings and evidence

| Finding | Confidence and evidence | Priority |
| --- | --- | --- |
| A valid capture can exceed the derived-entry bound, commit successfully and disable the complete catalog; rebuild cannot repair it | Source independently checked; fresh real-handler counterexample from the Go reviewer, details below | High, new 242 correction |
| Native definition publication can make its own catalog unavailable | Source independently checked; root's isolated actual-handler run acknowledges 257 distinct publications, then catalog returns 400; duplicate retry still returns 201 | High/medium, finish I2 before revision expansion |
| Index repair requires a whole-store rebuild that refuses above 256 MiB of payloads, even when metadata search can otherwise admit the store | Source-verified recovery gap; disposable sparse-file preflight corroborates the ceiling, not valid-capture publication | Medium, bounded operational follow-up |
| Explanation promotion holds the bank-wide mutex through payload and derived-index filesystem work | Source-verified serialization; no slow-filesystem or production stall was reproduced | Medium reliability cut |
| API effect documentation classifies legacy capture GET as the mixed source/publication path | Source-verified spec-generator defect; the handler is source-free | Small documentation/conformance fix |
| Metadata reads have per-request limits, but no separate server-wide I/O admission | Source-verified policy difference; no OOM or load failure was reproduced | Selected deployment/performance follow-up, not an immediate framework rewrite |

### AC1 — Align capture admission with indexability

The authoritative/derived split is correct. The failure is the incompatible size
contracts around it:

- `exploration/capture_store.go:373–385` admits a capture up to 24 MiB, publishes
  and verifies it, then intentionally ignores the result of `indexCapture`.
- `exploration/capture_catalog_index.go:21–42` serializes the **complete receipt,
  including its full typed question**, into an envelope capped at 64 KiB.
- `exploration/capture_search.go` refuses the entire search when any authoritative
  ID has no metadata. This is deliberate fail-closed behavior, not page filtering.
- `RebuildCatalog` calls the same bounded entry writer, so a permanent size mismatch
  is not repaired by the recovery command the error recommends.

The Go reviewer supplied this fresh actual-handler result from an owned temporary
store and source fixture:

```text
POST /captures                          201
raw JSON request bytes                  12,728
serialized receipt bytes                65,557
GET /captures/search                    503 (metadata catalog unavailable)
GET /captures/{new ID}                  200
rebuild indexed                         0
rebuild error                           capture search scan budget exceeded
```

The valid question has 31 AND predicates over `column:service`, each comparing a
340-character `<` string. JSON encoded without HTML escaping is within the capture
handler's 16 KiB input bound, as with `JSON.stringify` in the browser. Go metadata
marshalling expands those characters; the receipt alone exceeds the entry cap
before its envelope is added. A no-match or unusual query is still a valid query;
accepting it must not silently make unrelated captures undiscoverable.

**Bounded response:** prefer a small catalog projection of the fields actually
needed for discovery and display: immutable ID/parent, title, capture time,
question family/display kind and canonical declared scope. Derive scope through
the existing `CaptureQuestion.Scope`; do not infer the first service predicate.
Keep the full question in the authoritative capture and exact-ID response. The
new search API is prerelease, so its response can be made explicit rather than
retaining a large receipt solely for a table that uses a few fields. This does
not require another database or a shared universal store.

An alternative is explicit, correctly encoded indexability admission **before**
immutable publication, with an actionable refusal. Merely propagating the ignored
error after publication is wrong: the capture is already committed. Raising the
entry limit alone moves the mismatch and increases scan/memory costs.

**Acceptance:** the 12,728-byte request must either be rejected before any final
capture is published, or return 201 with a working catalog containing it. Existing
captures remain discoverable. Test direct Go and actual-handler input, metadata
encoding expansion, duplicate retries and repair of already committed captures
from the failing state. Preserve their immutable bytes and IDs.

### AC2 — Give native definition writer and reader one capacity contract

`cmd/goo/measurements_http.go:70–142` caps per-directory physical scans and total
YAML entries at 256 before returning its complete revision list. The POST path
at `:179–219` validates, publishes and returns 201 without capacity admission.
The root's isolated generic-runtime probe now reproduces the old source finding:
257 distinct POSTs return 201, GET returns 400 with `package catalog scan limit
exceeded`, and a duplicate of the last revision returns 201. The runtime has no
built-in directory, so this does not depend on lab configuration. See the root's
`artifacts/review-243/62a3d74e/review.json` while that temporary review evidence is
retained; the review's code/source description remains the durable record.

Resolve whether the supported contract is bounded publication admission, paged
catalog plus exact-revision lookup, or another explicit scheme. Count built-ins,
saved files, temporary/unrelated physical entries and unique revisions separately.
A built-in revision copied into the saved directory consumes a physical file but
not a new logical revision. A plain count-then-publish without reservation can race.

**Acceptance:** an accepted new revision does not break discovery. Duplicate
success still verifies its existing winner; corrupt winners are not replaced.
Concurrent distinct saves cannot bypass the chosen policy. Existing over-limit
stores retain their bytes and known revisions; do not solve admission by deleting
old definitions or silently truncating a purported complete catalog.

### AC3 — Make metadata repair possible independently of total payload size

`CaptureSearchMaxBytes` bounds metadata input to 16 MiB, while
`RebuildCatalog` preflights **all payload sizes** against 256 MiB before repairing
any entry (`capture_catalog_index.go:139–160`). There is no ID/range/batch option.
Thus a catalog can have small, admissible metadata yet lose all search because of
one missing entry and have no supported repair command once payload volume grows
past the rebuild ceiling. Ordinary committed captures need not approach 10,000
entries to exceed 256 MiB.

A disposable probe created eleven regular sparse files of 24 MiB and called the
public `RebuildCatalog`; it returned `indexed=0` and `capture rebuild exceeds
256 MiB; direct IDs remain reopenable` at 276,824,064 logical bytes. Those files
were intentionally not valid captures: this only corroborates the preflight
policy. The valid-store consequence follows from the independent per-capture,
metadata and whole-store bounds; it is not a claim of a production outage.

**Bounded response:** offer exact-ID repair or deterministic resumable bounded
batches, preserving full verification for each repaired payload and refusing
partial search until missing entries are resolved. Name repair versus whole-store
verification distinctly. Expose which entry needs repair through a bounded
metadata diagnostic; the current generic missing-index message omits its ID.
Do not replace explicit repair with an unbounded hidden UI body scan.

### AC4 — Move promotion I/O outside the bank-wide mutex

`exploration/result_snapshots.go:135–182` holds one `ResultSnapshots.mu` while
loading a previously promoted capture or publishing a new explanation. The latter
now also verifies the payload and performs derived-index write/sync/rename/sync.
Freeze/load/pruning of unrelated observations need that same mutex. This remains
a source-confirmed serialization seam; this review does not claim measured stalls.

Keep the existing pre-I/O custody reservation and ambiguous-commit identity. Use a
keyed in-flight promotion reservation while releasing the **bank-wide** lock for
filesystem work. A barrier publisher should hold A while unrelated Freeze/Load B
completes; duplicate A must still produce one candidate, including on uncertain
commit/retry. Bound reservation expiry and count. No universal job/history layer
is needed for this fix.

### AC5 — Make effect descriptions method-aware

`scripts/build_openapi.py:259` special-cases `/captures` without checking its
method, reporting `source-read-and-capture-or-held-explanation-publication` for
legacy GET as well as POST. The GET handler in `internal/capturehttp/routes.go`
only calls `store.List`; it does not use the configured telemetry source. The new
`/captures/search` effect is correctly separate. Fix the method-specific mapping
and test source-free GET versus mixed POST. Over time, consume family-owned effect
metadata where practical so route, admission and documentation declarations cannot
drift, without generating semantic handlers or imposing one lifetime on all paths.

### AC6 — Treat metadata admission as a separate measured concern

`newQueryBudget(mux,4)` controls selected source/pure-plan effects. The saved-capture
GET handlers bypass that budget, and Go serves concurrent requests. A per-search
16 MiB/8 MiB input/receipt budget is not a server-wide concurrency or RSS budget.
The contract already says so, which is good. If shared deployment or actual
interactive contention is selected, measure metadata work separately and consider
a small independent admission bound. Do not put source-free catalog reads behind
long-lived SSE slots or claim an OOM from maximum-size arithmetic alone. An owned
barrier test can verify partitioning and release without a large load campaign.

## Boundaries to preserve

- One Go process owns browser/native routes, embedded assets, admission and
  ephemeral capabilities. The removed Python gateway must stay removed.
- `subjectfacts`, `lifecycle`, `metrics`, `measurements` and exploration keep their
  distinct semantic roles. Native evaluation can remain storage-free while an
  explicit observation path owns custody. A catalog issue does not justify
  introducing DataFusion, CEL/KCL execution or mandatory XTDB.
- Family-owned source policies and shared low-level transport mechanics are the
  right split. Preserve source/SSE/mixed-publication lifetimes and explicit SDK
  ownership. The 237 receive fix and 238 runtime consolidation close earlier
  findings; they should not be reopened as unfinished proxy work.
- Derived metadata is replaceable; immutable payloads and custody remain
  authoritative. The documented stat/checksum versus reopen verification split
  is honest and was not found to be a defect. Do not imply that a metadata listing
  authenticates full payloads.
- Ko's generic image and the configured lab image currently have different
  filesystem declaration inputs and ownership requirements. Those boundaries
  are documented and preferable to a premature packaging unification. Requiring
  configured input bundles before claiming equivalent ko lab deployment is sound.

## Suggested order

1. Close AC1's accepted-capture/catalog mismatch, which directly breaks the
   save → find loop. Give the trace sheet an explicit visible parent-return
   contract as a bounded coherence task; do not treat normal background-button
   occlusion as proof that the working Close trace → Return path is unavailable.
2. Finish the small malformed prepared-string correction and AC2 with explicit
   native capacity semantics. The Go reviewer owns the UTF-8 details.
3. Settle adoption for unreported environment and the baseline exact-observation
   preservation contract with named tasks; keep absence distinct from wildcard.
4. Add AC3's local repair path and AC4's promotion isolation as bounded follow-ups.
5. Proceed to comparison support and recurring trace question reuse, then decide
   whether measured deployment needs AC6. AC5 is a small useful conformance fix.

This order is a proposal for the planner, not a new independent backlog. Preserve
completed 239–242 outcomes and fold the agreed tasks into the single active roadmap.
