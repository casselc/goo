# Independent Go and semantic-library review — 235

Reviewed 2026-10-28 against `5d12ce9`. Scope: root semantic/compiler/source/storage/
HTTP packages, the independent instrumentation module, example applications and
nested verification/tooling boundaries. This is source review plus two small Go
experiments; no broad root/race campaign, live telemetry mutation, code fix or
performance benchmark was performed. Only this report was added.

## Recommendation

Keep the finite package/identity cleanup closed. The new public checked APIs and
explicit runtime ownership are improvements. Put a bounded inbound-body deadline
fix ahead of product P1, and fix malformed prepared-string identity handling as a
small semantic correction. Preserve P1/P2's user outcomes rather than opening a
new generics/framework cleanup program. Promotion lock scope deserves a separate
reliability cut; P2 already needs a careful metadata/read-budget contract.

## Findings, ordered by impact

### G1 — High: admission/cancellation does not bound an incomplete request body

**Verified mechanism; actual-budget corroboration in the architecture review.**
`internal/queryhttp/budget.go:97–112` creates a 60-second context and releases its
active slot only when the handler returns. `internal/httpjson/json.go:17–22`
blocks on JSON body reads. The configured server at `cmd/goo/serve_http.go:122`
sets `ReadHeaderTimeout` and `WriteTimeout`, but no request-body read deadline.
A context deadline does not itself interrupt an inbound `net/http` Body.Read.

My narrow Go 1.27 experiment used a real local HTTP server, Content-Length 100,
one supplied body byte and a 50ms request context. After cancellation and a
further 150ms observation the body read was still blocked; closing the peer
released it. This confirms the mechanism, not a production load measurement.
The independent architecture reviewer exercised the actual NewBudget/Decode
path with capacity 1: explicit cancel returned 202 and cancelled the context, but
another valid request got 429 until the incomplete uploader disconnected, then
200. See [that review](ARCHITECTURE_REVIEW_235.md) for its exact reproducer.

Consequence: cancellation/admission ceilings are incomplete before decoding
finishes. With the configured shared limit of four, incomplete controlled requests
can occupy capacity despite cancellation. This is relevant to the direct Go HTTP
server; it does not establish that every Python-proxied request follows the same
streaming upload path, nor require assuming a public hostile deployment.

**Bounded fix/acceptance:** define inbound body time/cancellation policy before
source execution. Use actual network read deadlines (for example ResponseController
where supported, or a reviewed server read policy), not just a context or an
unproven goroutine calling Body.Close. Preserve long response/SSE and immutable
publication semantics. Real partial TCP-body tests must prove cancellation/timeout
releases capacity without requiring peer close, no source read starts, retry works,
and ordinary body limits/statuses still match. This should precede cosmetic P1.

### G2 — Medium: malformed UTF-8 prepared strings silently collapse groups

**Reproduced public-library semantic defect on malformed Go input.**
`subjectfacts/evaluate.go:63–68` checks that declared string fields are Go strings,
but not that they are well-formed UTF-8. Group identity at `:93–103` uses
`json.Marshal(groupValues)`. encoding/json replaces malformed UTF-8 with U+FFFD,
so different accepted strings can acquire the same group key.

A direct public API experiment constructed a valid definition with fields
`evaluation_id`, `id`, `label`, subject key `id`, dimension `label` and a count.
Two facts used distinct valid subject IDs and label values
`string([]byte{0xff})` and `string([]byte{0xfe})`. Evaluate returned
`error=<nil> groups=1 count=2`. This is not a claim that normal browser JSON can
carry arbitrary invalid Go bytes; it is a real unsupported-input hole in the Go
library's strict preparation/grouping contract. ClickHouse String grouping can
preserve such distinct byte values, so silently accepting them also undermines a
portable semantic comparison.

The exact experiment was equivalent to:

```go
package main
import ("fmt"; "github.com/casselc/goo/subjectfacts")
func main() {
    d := subjectfacts.Definition{APIVersion:"goo/v1alpha1", Name:"bytes",
        Source:subjectfacts.Source{Kind:"subject-facts",
            Fields:map[string]string{"evaluation_id":"string", "id":"string", "label":"string"},
            SubjectKey:[]string{"id"}, Bindings:map[string]string{"clickhouse":"facts"}},
        Dimensions:map[string]string{"label":"label"},
        Measures:map[string]subjectfacts.Measure{"rows":{Op:"count"}}}
    rows, err := d.Evaluate([]subjectfacts.Fact{
        {"evaluation_id":"e", "id":"a", "label":string([]byte{0xff})},
        {"evaluation_id":"e", "id":"b", "label":string([]byte{0xfe})},
    }, "e", []string{"label"})
    fmt.Printf("error=%v groups=%d count=%d\n", err, len(rows), *rows[0].Measures["rows"].Count)
}
// Observed: error=<nil> groups=1 count=2
```

**Bounded fix/acceptance:** explicitly reject malformed strings in selected-
evaluation facts before filtering/grouping/subject-key construction, consistent
with native identity validation. Review authored string predicates and enum values
at `subjectfacts/model.go:91–110,117–146` for the same text contract. Preserve valid
Unicode, literal empty values and existing IDs/output. Reject, do not normalize,
malformed strings; test distinct bad bytes, valid U+FFFD and excluded malformed
facts. Add the direct Go API case alongside existing source/compiler parity tests.
This is a semantic correction, not another generic-hash migration.

### G3 — Medium: one promotion's filesystem work serializes the whole snapshot bank

**Source-verified serialization risk; no production stall benchmark claimed.**
`exploration/result_snapshots.go:138–172` takes the single bank mutex and retains
it across `store.Load` and the explanation publisher, which reaches immutable file
write/sync/publication. Freeze, Load, other promotions and expiry pruning need the
same mutex. Slow storage for one save can therefore delay unrelated result-page
operations; current token/byte bounds do not bound filesystem latency.

Move publication outside the bank-wide critical section only with a keyed in-flight
reservation. Preserve the already-reserved CapturedAt, retry candidate identity,
ambiguous-commit behavior and bounded lifecycle/expiry of in-flight entries. A
barrier publisher test should hold save A and show that unrelated Freeze/Load B
still completes while duplicate save A remains idempotent. Do not introduce a
universal background job/store layer or discard custody reservation to shorten
the lock. Schedule separately from the initial body-deadline correction.

### G4 — P2 constraint: catalog ordering/search cannot scan full capture bodies naively

**Verified current read path; prospective performance risk, not a measured OOM.**
`exploration/capture_store.go:492–540` sorts IDs and calls full `Load` for each of
up to 20 selected entries. `Load:385–400` reads up to 24MiB, converts the whole byte
slice to a string for decoding, then validates the saved content/digest. A page
can therefore consume substantial cumulative read/decode/allocation work even
though its response contains only receipts. Twenty maximum-sized files mean
roughly 480MiB of cumulative input, not 480MiB of necessarily simultaneous resident
memory. The string conversion is another avoidable full-buffer allocation.

The current P2 backlog correctly forbids page-local fake recency and requires a
metadata/index contract. Strengthen acceptance with bounded metadata I/O and
allocation tests; do not implement global sorting/search by repeatedly loading
all 10,000 possible complete bodies. Keep content integrity and direct-ID reopen
separate from metadata lookup, with explicit rebuild/stale/corruption policy.
`bytes.NewReader(b)` can remove the redundant string allocation as a small
mechanical change, but does not solve full-catalog metadata access. No new database
is implied by this finding.

### G5 — Low/medium API contract gap: compiled plan metadata aliases the authoring input

**Source-verified alias; no existing user failure reproduced.**
`subjectfacts/compile.go:65–66` retains `Plan.Definition = d`, while SQL/metric-view
text is generated immediately. An ordinary caller editing the definition after
Compile changes reported plan metadata without regenerating Query/MetricYAML.
The API does not clearly state that this is a live view requiring caller-owned
immutability. This differs from the detached native evaluation/result/reference
contracts established elsewhere.

For a prerelease API, prefer a detached definition snapshot in the compiled Plan,
or document an intentionally borrowed view and expose a separate immutable report.
A detached snapshot is simpler for ordinary callers and avoids asking them to
understand aliasing across nested maps. Accept with an external consumer test:
compile, mutate original maps/predicates/bindings, and prove plan text and metadata
remain aligned. Do not use a generic reflection clone API just for this type.
This can follow user-facing work unless the compiler API is being expanded now.

## Structure, standard library and public ergonomics

- **Preserve the semantic layers.** `subjectfacts` is the small portable prepared-
  fact compiler/evaluator; `lifecycle` prepares business observations; `metrics`
  owns numeric kernels; `measurements` binds native OTel and evaluates supplied
  rows. `measurements/source.go:141–154,260–268` separates source execution from
  supplied-data evaluation and explicit lifecycle custody. Do not flatten these
  into a universal planner or let UI convenience definitions alter metric grain.
- **Recent APIs are well chosen.** `Definition.Revision()` and `Question.Key()`
  validate concrete values and return errors, while generated authoring seeds stay
  private. Compatibility fixtures justify their storage promises. No clever type
  parameters are required by users. Leave audited result/custody identities alone
  unless a real consumer requires change.
- **Use modern Go internally where it reduces work.** The current runtime candidate
  path uses `slices.Sorted(maps.Keys(...))` and reuses final revisions. This is a
  good application of standard generic collections. `subjectfacts.names` already
  sorts, so `subjectfacts/evaluate.go:112–114` sorting its output again is harmless
  redundancy, not a performance project. Consolidate while touching that code.
- **Concurrency primitives are purposeful.** Runtime `atomic.Pointer`, per-key
  `sync.Once` and finalizer `sync.Once` express one-time binding and concurrent
  completion directly (`instrumentation/runtime.go:79–90,112–160`). Replacing them
  with generic registries or newer syntax without benefit would obscure ownership.
  The finite initialization map bounds bookkeeping, not all emitted metric label
  cardinality; callers must still use finite names/roles as documented.
- **Snapshot ownership should become explicit, not generic.** New code should
  detach only the concrete inputs it retains. Current JSON clone patterns in
  runtime discovery are understandable but should have checked errors or a typed
  clone where identity/serialization is not the intended operation. Do not rewrite
  every map/slice with reflection or user-facing type machinery.
- **Filesystem improvements are conditional.** Immutable publication correctly
  documents ambiguous link/directory-sync outcomes (`internal/immutablefile/
  publish.go:11–28`). Preserve these distinctions. `os.Root` could anchor a future
  P2 metadata implementation, but is not justification to change trusted directory
  or symlink policy silently. It is not a substitute for auth or content integrity.
- **Modules remain useful.** Root, instrumentation, two examples and two verification
  modules keep SDK/native/tool graphs separable. The root still includes optional
  retention/pgx code; moving it to another module is only justified by an actual
  lightweight-consumer/deployment requirement. A metrics-only import does not
  automatically link the PG driver merely because root go.mod lists it.
- **Examples and tests should keep teaching ownership.** App-owned provider shutdown,
  bounded server drain and direct Runtime construction are preferable to implicit
  process-global SDK installation. Verify a woven binary through Orchestrion;
  manifest emission or plain compilation is not proof of instrumentation.
- **Verification is now more honest.** Nested modules still use a sibling replace
  (`verification/go.mod:35`), but the opt-in runner pins/checks that sibling and
  rejects missing named passes/skips. Root CI is not evidence every optional
  Hegel/Quint/woven/XTDB/browser path ran. Keep those tiers proportionate.

## Revised order and validation limits

1. Correct inbound body cancellation/deadlines; add real TCP/admission regression.
2. Correct prepared malformed-string identity and document the supported text domain.
3. Continue mobile Activity P1 and capture discovery P2, carrying the metadata budget
   acceptance above; promotion lock isolation can be an independent reliability cut.
4. Retain comparison-support P3. Validate denominator/control semantics against exact
   counts and incomplete coverage before adding inferential language.
5. Reusable trace-query P4 needs one concrete cross-service contract and storage/
   protocol compatibility, not a general DataFusion/CEL execution-engine rewrite.

No current evidence supports a Go/Rust rewrite, production query optimizer, store
unification, exported generic endpoint framework, broad module split or replacing
stable semantic code solely to use newer stdlib/runtime features. No runtime code,
fixtures or dependencies were modified by this review. Both small experimental
programs were temporary and removed; their observed results are recorded above.
