# Runtime/source understandability review

Reviewed 2026-11-16 against the current code; recommendations are proposals, not
implemented refactors. Scope: CLI/runtime composition, HTTP families, record-source
execution, transport helpers and the boundary to workbench storage. Domain
measurement/evaluator internals and store formats belong to their own reviews.

[Walkthrough](../../walkthroughs/runtime.md).

## Overall assessment

The runtime can be understood in an afternoon **if the reader follows one request
first**. Most individual boundaries are already explicit: one server, injected
source values, standard `http.ServeMux`, typed route helpers, family-owned effects,
no hidden startup query, and a deliberately small source-transport package.
Understanding all supported record projections still takes substantially longer;
that is partly real product breadth, partly the acquisition coordinator's shape.

The main simplification opportunity is to make the executed work apparent without
mentally evaluating many feature flags. Removing bounds or uncertainty handling
would make the code shorter while making its meaning less reliable.

## Findings, ordered by value

### R1 — Make the record acquisition sequence explicit

**Evidence:** [`Source.fetchProfile`](../../../exploration/source.go) handles
question freezing, page-only query rewriting/recompilation, count/row reads,
optional numeric/status/duration decoding, grouping, discovery, timelines,
changed-between-read detection, and display dataset construction. `Result` in the
same file has many optional presentation payloads. A reader cannot see a small
execution outline without following each conditional through the coordinator.

**Proposal:** Extract a handful of named stages inside `exploration`, keeping a
plain coordinator visible: freeze/compile question; acquire population and rows;
acquire requested supporting results; establish read consistency; build datasets.
Start with the page-only versus complete-result branch. Give each stage typed
inputs/outputs; do not replace this with reflection, a plugin registry, or a
universal executor. Keep independent reads and their failure behavior explicit.

**Acceptance:** Same query/parameter bytes and source-read order; unchanged exact
counts, page-only avoidance of unnecessary reads, refusal-before-source tests,
changed-between-read outcomes, and source-budget failures. Test the source request
count as well as returned JSON. This can be done without changing the public API.

### R2 — Name and expose the layered execution budgets

**Evidence:** [`configuredHTTPServer`](../../../cmd/goo/serve_http.go) chooses a
15-second shared HTTP client, four generic query slots, server receive/write
bounds and shutdown grace. [`recordhttp.PolicyFor`](../../../internal/recordhttp/policy.go)
chooses route deadlines up to 55 seconds. [`Source.read`](../../../exploration/source.go)
sets the physical ClickHouse execution, byte, row and memory settings.
[`registerTail`](../../../cmd/goo/tail_http.go) separately chooses four stream slots,
connection lifetime, polling interval and frame-write bounds. These distinctions
are legitimate, but their relationship is not available in one small code view.

**Proposal:** Introduce named policy values at the existing owners and a short
budget table beside the runtime walkthrough. Share a value only when its meaning
is actually identical. Keep tail and ordinary-query admission separate and label
the combined process exposure rather than pretending one limit covers both.
Do not expose every tuning knob in the UI or build a generic configuration DSL.

**Acceptance:** Existing deadline/cancellation/admission tests remain exact; a
fixture demonstrates nested deadlines cannot extend the parent. Preserve bounded
SSE writes and client-disconnect cancellation. Any actual budget increase should
be a separate measured change, not incidental to renaming constants.

### R3 — Keep route effects easy to audit as families grow

**Evidence:** [`newQueryBudget`](../../../cmd/goo/query_budget.go) lists family
policy lookups; each family declares exact method/path policies and separately
registers routes. [`policy_contract_test.go`](../../../internal/queryhttp/policy_contract_test.go)
compares OpenAPI body limits/admission to those declarations. Generic
[`ReadHandler`](../../../internal/queryhttp/handlers.go) is concise and infers its
type arguments, while mixed publication routes retain explicit ownership.

**Proposal:** Keep this design. Add a concise effect legend/link to family route
files and group registration visibly by pure planning, source read and publication.
If registration/policy drift becomes common, use one small typed declaration per
family; do not introduce a second global route registry merely to remove a short
list. The existing contract test is more valuable than cosmetic DRYness.

**Acceptance:** Every new published source route is covered by both policy and
OpenAPI checks. Pure routes never consume source capacity. Mixed publication
lifetimes are not silently inherited from read-only wrappers.

### R4 — Reduce adapter-local HTTP boilerplate where behavior is identical

**Evidence:** [`registerWorkbench`](../../../cmd/goo/workbench_http.go) still defines
local `reply`/`fail`/`decode` helpers while neighboring extracted families reuse
[`httpjson`](../../../internal/httpjson/json.go). The local writer currently has
the same JSON/no-store shape, but status mapping and store effects remain local.

**Proposal:** On the next touched workbench route family, reuse `httpjson.Write`
for demonstrably identical writers and extract a named registration unit around
one document family. Do not migrate all routes in one sweep or homogenize source,
publication, SSE and download responses. Favor ordinary typed functions over
higher-order options that obscure the concrete HTTP contract.

**Acceptance:** Wire tests verify status, headers, body limits and error envelopes;
publication/conflict/uncertain-outcome tests remain intact. No source or store
work moves into transport helpers.

## Dead or unnecessary work

No production runtime file was proven unused by this review. Existing route
registrations, CLI dispatch and tests show active entry points. The API-only
root paths and browser `/api` prefix intentionally share the same handler; that
is not evidence of a duplicate network gateway. Generic and lab profiles have
real different configuration requirements.

The word `Legacy` remains in some contract-test names, for example dashboard
omitted-profile tests. That wording alone does not prove a legacy behavior can be
deleted: first decide whether the omitted profile is still the default product
contract. Rename misleading tests or remove an unused contract only after checking
current callers. There is no prerelease compatibility obligation to preserve a
truly obsolete format, but this review did not establish one here.

## Performance and reliability conclusions

- The ordinary fetch path deliberately makes several sequential reads. There is
  no evidence here that parallelizing them improves real latency within source
  budgets. Measure first; preserving read-consistency semantics matters.
- `Source.fetchProfile` freezes a query via JSON round-trip. That is real copying,
  but queries are bounded and no profile established it as a bottleneck. A typed
  deep-copy helper is a candidate only if it is clearer and covered against caller
  aliasing; do not remove the freeze as an optimization.
- Cancellation tombstones, bounded JSONEachRow parsing, physical source limits,
  source-specific error mapping and shutdown draining are useful mechanisms.
  They should survive any simplification.

## Checks

Source inspection and reference searches performed. Focused checks passed on Go 1.27.3:

```sh
go test -p 2 ./internal/queryhttp ./internal/recordhttp ./internal/sourcehttp \
  ./internal/sourcejson ./internal/httpjson -count=1
```

All five packages pass. These exercise local HTTP fixtures, not a production
source or browser campaign. No implementation was changed for this review.
