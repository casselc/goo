# Captured-record review — strict readiness contract 365

> Full feature delivered after the readiness snapshot below. Final source review
> accepted the production store/API/UI contracts; native fifteen-fact replay,
> positive actor and genuine frames pass. See [current contract](../CAPTURE_RECORD_REVIEW.md)
> and [executed validation](../../VALIDATION.md). The gate-only introduction below
> records the earlier stage; it is not the current implementation status.

**Readiness gate: ACCEPTED for explicit selected-revision branches.** The bounded
contract and resolver/file-publication demonstrations are accepted. This is
architecture/foundation readiness, **not delivery of the reviewed labeling feature**.
Only the capture resolver is production code; the task/decision/manifest example
remains test-only. No review UI, HTTP family or production export is implemented.
This review ran no heavy checks; execution results below are reported by root.


## Actual foundations and gaps

`exploration.CaptureStore.Load` bounds bytes (24 MiB), checks immutable content ID
and validates at most 1,000 stored inspection rows. Resolve a zero-based absolute
position in `SavedCapture.Explore.Rows`; it is not a browser page position,
reported span ID, timestamp or caller-supplied record. `json.RawMessage` preserves
numeric tokens and exact stored row content. Decoding through `map[string]any`
would corrupt integers beyond 2^53 and is prohibited for evidence resolution.

`workbench.MeaningTarget` has no capture-member target and `CreateMeaning` rejects
evidence targets without the optional evidence store. This is not an existing
file-tag feature with a new form. Reuse file publication primitives for separate
bounded review artifacts rather than overloading notes or requiring XTDB.
`Store.publish`/`immutablefile.Publish` provide immutable filename publication,
not a current-head CAS, global transaction or consensus service. Load/validate the
winner and recompute its content ID; a successful filename collision alone is not
proof of valid bytes.

## Minimal wire and server ownership

Use ordinary versioned Go structs and strict JSON fields:

- **Task definition:** version, name/title and one categorical vocabulary of
  distinct bounded `{key,title}` entries. Its immutable content ID identifies the
  exact vocabulary version. A negative class must be explicitly declared; no
  complement or absent-label inference is permitted. One task version per manifest.
- **Record input target:** `{captureID, absoluteIndex}` only. The server loads the
  verified capture, refuses non-exploration/out-of-range targets, clones the exact
  raw message and computes its SHA-256. Server-owned resolved metadata includes
  receipt, original question/basis/read/custody and row hash. Browser record fields,
  IDs, number conversions and submitted digests are not evidence authority.
- **Decision revision:** task ID, resolved target/hash, bounded parent revision
  IDs, judgment `{kind:'label',value:key}` or `{kind:'uncertain'}` or
  `{kind:'abstain'}`, review `unreviewed|reviewed|disputed`, state
  `active|withdrawn`, and bounded note. Server assigns timestamp/content ID and
  `local-user-unverified` origin. Parents must have the same task and resolved
  target. Withdrawal requires a parent; no prior artifact is overwritten.
- **Manifest:** version and explicit mode `pinned-review-branches`, one task ID,
  explicit selected decision IDs, bounded pinned ancestry/observed-head IDs and
  resolved capture-member/hash references. Server supplies publication clock/ID.
  It contains no telemetry query, latest pointer, automatic membership rule or
  generic model format. Raw evidence remains resolvable by immutable references.

The axes remain separate: no decision is unlabeled, not abstention; abstain and
uncertain are deliberate judgments, not negative labels. Reviewed means a local
recorded review decision, not authenticated reviewer identity or ground truth.
Withdrawn/disputed/unreviewed/uncertain/abstaining decisions do not contribute a
training label. A historical reviewed revision can be pinned explicitly, but is
never presented as the current consensus.

## Closed concurrency and manifest semantics

Allow append-only forks. Never choose the newest timestamp or silently replace a
winner. A manifest is an **explicit selected decision branch snapshot**, not all
current judgments, latest truth or a globally consistent cut. Pin its selected
revision and known bounded ancestry so later correction, withdrawal or another
branch cannot rewrite the already published manifest. Existing manifests remain
reopenable and report exactly their original pinned judgments.

An explicit disputed revision or multiple heads in the supplied pinned known-head
set makes that target ineligible until an explicit resolution revision references
those heads. Parent/head references are validated against the same target/task.
Unknown concurrent or unselected branches are not claimed absent. Display that
limitation before publication. No atomic “all current heads” guarantee is needed
or offered by this contract. If the desired feature instead requires automatic
latest/consensus export, this gate is **DEFER**: the current primitives do not
provide that stronger contract and it must not be smuggled into the pilot.

For the first implementation, keep a small declared ceiling (proposed: 16 labels,
32 explicitly selected members and bounded 64 revision references per manifest,
8 parents per decision). Apply the existing 128 KiB artifact limit and explicit
capture-read/ancestry limits; refuse excess rather than truncate. A one-capture
manifest is an acceptable initial restriction and bounds raw evidence acquisition
to one existing 24 MiB capture. No automatic global head scan, taxonomy, reviewer
assignment or multi-tenant authorization is required. Exact numerical limits must
be finalized in code/tests before public feature work.

## API and storage placement

The root demonstration should expose a plain resolver such as
`CaptureStore.ResolveRecord({CaptureID, AbsoluteIndex})`, returning detached
`json.RawMessage` and server-computed immutable identity. A small review store or
workbench methods can accept that resolver as an explicit dependency, analogous
to existing package-reference resolution; avoid a global resolver registry.
Store task/decision/manifest families in separate trusted workbench subdirectories
using existing digest/publish/load validation. The Go server's existing API module
owns bounded decode and calls these methods. No source query occurs when
resolving, judging or publishing a pinned manifest; capture file reads are separate
counted effects.

## Required readiness demonstrations

1. Resolve a real saved row beyond page zero and duplicate span IDs by absolute
   occurrence; reject foreign/malformed/out-of-range references. Show raw JSON
   integers above 2^53, decimal tokens, missing/empty/null and Unicode are retained
   without float conversion. Mutating a returned raw byte slice cannot mutate
   stored bytes or another resolution.
2. Publish/load an immutable task, append decision revisions and pin a bounded
   manifest through existing file primitives, verifying content hashes and size
   refusal. Caller inputs remain detached; reopening checks actual referenced
   capture bytes/hash rather than trusting browser fields.
3. Demonstrate explicit disputed/forked pinned ancestry refusal, explicit
   resolution, uncertain/abstain/withdrawn exclusion and old-manifest stability
   after a later correction. Concurrent different revisions may both exist; no
   latest/consensus claim is made. Publication errors must not imply nothing was
   committed; retry/load verifies an existing winner.

These demonstrations close a useful small review/export contract. They do not
implement a label suggestion engine, classifier, XTDB migration, collaborative
review service or a general dataset platform.


## Concrete bounded wire example

Illustrative names, with IDs replaced by actual server content IDs at publication:

```json
{
  "version": 1,
  "taskID": "<immutable-task-content-id>",
  "target": {"captureID": "<immutable-capture-id>", "absoluteIndex": 123},
  "parents": ["<prior-decision-id>"],
  "judgment": {"kind": "label", "value": "timeout"},
  "review": "reviewed",
  "state": "active",
  "note": "Reviewed this exact retained observation."
}
```

The server resolves/adds the row hash and original evidence identity; the request
cannot replace the record. `{"kind":"abstain"}` and
`{"kind":"uncertain"}` omit `value`. A `disputed` review or `withdrawn` state
retains history but contributes no active label. A manifest explicitly identifies
its `pinned-review-branches` mode and lists selected revision IDs; it does not ask
for “latest reviewed records.” Its known-head/ancestry set is bounded and declared,
not a proof that no unselected concurrent branch exists. Later withdrawal does
not change an earlier manifest. Selecting that older revision for another manifest
must be an explicit historical pin with the same no-current-consensus limitation.

Root implementation at this gate is limited to the production capture resolver
and a test-only file-publication prototype. The reviewed UI, HTTP review family,
manifest product flow and full metadata artifact family remain unimplemented.

## Resolver source review

The root's production foundation uses plain `CaptureRecordRef` with Go fields
`CaptureID` and `Index` (wire `captureId`/`index`); `Index` is the zero-based absolute
retained-row occurrence described above. `CaptureStore.ResolveRecord` bounds the
index before file access, calls the existing verified loader, requires logs/spans
exploration and checks actual retained length. It returns caller-owned
`json.RawMessage` via `bytes.Clone`, a SHA-256 of those stored row bytes and the
loaded receipt/source clock/basis/state/full-count/retention metadata. This reads
no telemetry and grants no judgment/review state.

Source review accepts this foundation. The new test source covers identical values
at positions zero and 100, page-local-to-absolute mapping, large integer/time
tokens, returned-byte/receipt mutation isolation, invalid indices/path/missing
files and corrupted capture integrity. Execution and the test-only publication
prototype result are recorded below; no full-feature delivery is implied.


## Final demonstration acceptance and remaining work

Independent source review accepts the production resolver, including explicit
inventory-family refusal, and the private workbench publication demonstration.
The prototype uses real existing file publication/read primitives, verifies
content IDs/winners, preserves two concurrently appended corrections, refuses
ineligible judgments and unresolved supplied heads, and publishes an explicit
known-branch resolution. Reopening the pinned manifest after later withdrawal
retains its original bytes and selected revision.

The final prototype export now reloads the pinned task and decision by manifest
IDs, verifies their content digests and task/target/record-digest binding, checks
label/review/state eligibility and resolves the raw captured row again. JSONL uses
the **loaded decision's label**, not the original draft variable. Foreign task,
target and record-digest joins are rejected. Large JSON numeric tokens remain
exact, oversize publication refuses the 128 KiB bound, and an existing forged
winner is not trusted. Root reports the final focused publication test passing in
0.013 seconds and the full main Go suite passing with the production resolver and
inventory refusal before the final test-only join addition.

The readiness result means these primitives support a bounded implementation
without a new storage/consensus platform. It does not mean the private prototype
is a production review engine. Remaining feature work includes final production
versioned structs/strict decode and bounds, every parent/head task-target-DAG
validation, server clocks/origin and publication/load validation, resolver-wired
HTTP routes, current/stale/retry/withdrawal UI ownership, explicit historical pin
and unknown-branch disclosure, production manifest/export selection and privacy/
size/error behavior. Any list/index or latest/consensus semantics would need its
own contract; none is promised. No optional evidence-store dependency or
MeaningTarget alias was introduced.

The reviewed UI/HTTP/schema lifecycle and dataset feature remain **not delivered**
in this milestone. Advancing to 366 is consistent with closing this readiness
proof; later implementation must not present this gate as completed labeling.

## Full implementation follow-up: source review, runtime pending

The earlier readiness-only status above describes the gate, not the subsequently
authorized implementation. Production `workbench/capture_review.go` now provides
bounded immutable task, decision and historical manifest publication/loading and
raw-record JSONL export. Shared parent-load caches and visited graph walks bound
DAG work to 64 revisions; selected/known-head joins remain task, absolute target
and record-digest specific. Export re-resolves immutable capture members and
preserves raw numeric tokens, with a 4 MiB complete-output limit.

`capture_review_intent.go` fixes a server timestamp through a durable request
intent before publication. The three `CreateCaptureReview*Once` APIs recover the
same revision after a lost response; changed kind/input returns the exported
intent-conflict sentinel. An internal self-digest detects altered intent fields,
including a syntactically valid replacement clock. This is corruption detection,
not authentication or a transaction across arbitrary artifacts.

The reviewed HTTP integration has no telemetry dependency. Its bounded strict
JSON boundary checks Unicode, unknown fields, explicit absolute index presence,
and omission of value for nonlabel judgments. POST requires a request ID;
conflicting reuse returns 409. Canonical record paths and verified artifact IDs
precede resolver access/export naming. Publication uses existing immutable-file
winner verification. No latest-head lookup, global consensus, authenticated
reviewer identity, task registry or optional XTDB dependency is implied.

Production backend source is accepted. The focused Go 1.27.3
`TestCaptureReview` rerun, including intent clock-corruption rejection, passed in
0.136 seconds. Final independent source review also accepts the integrated
HTTP and frontend boundaries: explicit judgments have no default label; changed
judgment/note must be saved as a correction before review; decisions bind the
resolved absolute occurrence/digest/task; known forks require explicit merging;
POST retries retain the original request ID and frozen body. Disclosure and
owner leases retire old actions and late adoption. Historical manifest opening
remains task-scoped rather than falsely limited to the inspected occurrence.

The actual eight-handler source-free HTTP test passes in 0.028 seconds and emits
13 successful wire observations, including raw numeric JSONL and stable export
after withdrawal. Strict refusal and no-store checks cover the same boundary.
Root reports full Go, 1,249 UI checks, OpenAPI 108 operations and Hegel/Quint smoke
gates passing. No material source blocker remains. Genuine runtime acceptance
is still pending and is not inferred from those code gates.


## Full-feature runtime acceptance

Root reports final Go/UI/API/generated gates passing, plus the final genuine
Playwright run (59.8s), fifteen recorded facts admitted to Quint (7.05s) and
Bombadil positive completion across 594 snapshots. Exact capture occurrence,
immutable retry identity, known-branch resolution and historical export survive
zero-telemetry operation. Native bytes and transport report objects have distinct
validation authorities; report pretty formatting is never hashed as source bytes.
No additional storage layer, latest-head scan or consensus claim was introduced.
