# Verification and process review — 243

Reviewed `65ac193` on 2026-10-29. This is an independent source/contract audit plus
one small synthetic model replay, not a rerun of every suite, a new deployment,
or a production load/accessibility certification. Only this report is changed.

Read AGENTS, current roadmap/backlog and validation through 242, CI, the optional
runner/tool pins, browser verdict helper, catalog Go/Playwright/Quint/Bombadil
checks, and the pinned gotogether pilot/follow-up. The parent supplied its new
combined 243 task results and isolated endpoint counterexamples; those are
identified below rather than claimed as reproductions performed by this reviewer.
No live capture, annotation, package or workload was published by this review.

## Recommendation

Preserve the fast gates and the explicit limits established in 237–242. Before
more feature breadth, repair accepted-writer/reader consistency at the capture
index and native package catalog. Then choose one complete pointer/keyboard user
task as a bounded required browser tier. Keep the larger historical campaigns
optional; adding test counts or replacing all tools would not solve these gaps.

## Findings and bounded acceptance

### V1 — High: accepted capture inputs can exceed the derived index contract

**Team endpoint reproduction; source mechanism independently confirmed.** The
parent reports a valid 12,728-byte `POST /captures` returning 201 with a 65,557-byte
receipt: the 64 KiB metadata entry cannot hold it, search returns 503, exact-ID
reopen returns 200, and rebuild returns zero indexed/error. This is permanent
indexability failure for an accepted shape, not a transient missing-file incident.

`CaptureStore.publish` commits and verifies the immutable capture, then discards
`indexCapture`'s error. The index serializes the complete `CaptureReceipt`,
including its normalized question, under a separate64KiB limit. Rebuild invokes
the same serializer. Thus existing missing-index/rebuild tests do not establish
that every accepted writer input is discoverable or recoverable.

**Next cut:** prefer a bounded catalog summary projection rather than duplicating
arbitrary complete questions in every index entry. Preserve full questions in the
immutable capture and fully verify them on reopen. Keep operational metadata-write
failure separate: do not report an already committed capture as an uncommitted
save, but do not admit a deterministic unindexable shape as routine success.

Acceptance must cross the actual HTTP writer, stored normalized receipt, search
and rebuild. Include the small-body/large-receipt witness and a property over
accepted bounded question shapes: every successful ordinary publication is
indexable/rebuildable, or a deterministic limit is refused before source work and
authoritative publication. Injected storage failures have a distinct recovery
contract. One failing entry must not silently turn unrelated catalog results into
an apparently empty population. No existing payload/ID may be rewritten to fix it.

Source: [publication](../../exploration/capture_store.go),
[index serialization/rebuild](../../exploration/capture_catalog_index.go),
[HTTP owner](../../internal/capturehttp/routes.go),
[current contract](../CAPTURE_CATALOG_SEARCH.md).

### V2 — Medium/high: native writer/catalog consistency needs an endpoint regression

**Parent's isolated actual-handler reproduction:** 257 distinct package POSTs all
return 201, catalog GET returns 400, and the duplicate retry still returns201.
This upgrades I2 beyond the previous source-only finding. It does not establish
payload loss: the symptom is a successful publication making discovery unavailable.

Add a producer/consumer contract test before choosing the implementation. Cover
builtins plus saved definitions, physical versus logical limits, duplicate retries
at/above capacity, corrupt winners, concurrent new publications, and preexisting
over-limit stores. State whether the supported outcome is admission refusal,
pagination/direct-ID lookup or another bounded policy; do not simply raise one
constant and retain divergent contracts. Keep authoritative files and IDs intact.

Source: [native package owner](../../cmd/goo/measurements_http.go).

### V3 — High process priority: no required browser tier exercises the product task

**Verified CI scope.** The sole workflow executes root/display/API tests,
independently versioned application modules and refusal controls. It does not
execute Playwright, Bombadil or the selected nested offline/woven suite. The
strict optional runner is useful, but an operator's successful local run and a
green hosted fast job are different claims.

The new combined 243 review task provides useful positive composition evidence:
three tail reads, two explicit trace/linked-log reads, one fixed analysis and one
capture; parent return/find/reopen introduce no new telemetry reads. Its initial
mobile direct parent-return attempt was intercepted by the full-screen trace
sheet. The visible Close trace inspector → Return path passes. Treat that as a
close/parent-return interaction contract decision, not proof that return is
impossible. Do not freeze a mistaken test expectation by forcing the pointer click.

**Next cut after that interaction is decided:** a selected required browser task
covering ordinary OTel → service scope → signal switch → inspect → supporting
trace/logs → preserve the declared result → navigate away → find/reopen. Run real
pointer hit/scroll checks at mobile width and a practical keyboard path. Assert
actual requested predicates/windows, which context reads broaden scope, held-state
return, immutable result IDs and source-free reopen. Use an owned temporary store;
choose a small explicit lab fixture/runner rather than making all old journeys a
mandatory suite. Required setup/collection failures must fail, never skip.

Source: [CI](../../.github/workflows/api-contract.yml),
[optional runner](../../scripts/verify_optional.py),
[verification entry point](../../verification/README.md).

### V4 — Medium: read-effect documentation can be wrong while schema gates pass

**Independently verified generated output and source.** `GET /captures` currently
has `x-effects: source-read-and-capture-or-held-explanation-publication`, identical
to POST. Its handler only calls `store.List`. `effect(method,path)` selects this
label by path before checking the method. The new search GET is correctly labelled
as a saved-data read. Route/schema/fixture validation does not catch this semantic
mismatch.

Add a method-specific effect fixture for the mixed capture family. GET list/reopen
must be saved-data reads; POST retains its genuine source-versus-held publication
branches. Use these explicit effects when classifying source-triggering requests
in a selected task; a POST-only counter is valid for 242's actions, not a universal
source-work oracle because routes such as GET `/operations` do read telemetry.
No universal handler registry or framework migration is required for this fix.

Source: [effect generator](../../scripts/build_openapi.py),
[generated contract](../../api/openapi.json),
[capture routes](../../internal/capturehttp/routes.go).

### V5 — Medium: keep model coverage separate from user-task completion

The catalog model observes only `reads` and a supplied `coherent` boolean. The
browser projection sets coherence from absence of page errors; separate UI/Go
assertions carry scope, ordering, IDs, index failure and integrity checks. The
validation record correctly limits the model's claim to read ownership.

A bounded synthetic probe here supplied only:

```json
[{"Action":"save","Reads":2,"Coherent":true}]
```

`TestQuintCaptureCatalogLedger` passed, including its three rejecting mutations,
in 6.18s. This confirms it accepts a conforming saved prefix without browsing or
reopening. That is expected for this safety abstraction, not evidence that 242's
Playwright journey falsely passed. It cannot independently certify task completion,
query identity or filesystem correctness.

For the chosen complete task, project a small stable vocabulary: source owner,
selector/window identity, selected/saved/reopened artifact identity and an
independently observed terminal checkpoint. Reject wrong predicate/window,
wrong reopened ID, an uncompleted action and extra source work at a local return.
Do not model every DOM detail. The gotogether follow-up already demonstrates
invoke/completion plus settled/eventual distinctions; reuse one such adapter if
it helps this task, rather than migrating all suites. A click is not completion,
and uncollected/lost tails must not become passing eventual claims.

Source: [model](../../verification/model/capture_catalog.qnt),
[binding](../../verification/model/capture_catalog.yaml),
[validator](../../verification/capture_catalog_test.go),
[browser assertions](../../journeys/242-capture-search.spec.js),
[pinned gotogether follow-up](../GOTOGETHER_FOLLOWUP.md).

### V6 — Medium enabler: automate code-under-test identity before broader claims

`startRuntime` records a binary hash, and recent milestones manually compared served
JS/CSS/OpenAPI bytes with checkout. The thin deployed API checker matches the spec.
These are improvements. An optional prebuilt `GOO_GENERIC_BINARY`, however, is not
itself proven to contain the expected source tree, and a matching spec cannot
distinguish two Go/UI builds with identical wire shapes.

A selected required/deployed tier should check one explicit build/input manifest
before executing: binary/image identity plus relevant embedded asset/spec hashes.
A known stale binary/asset with an unchanged API schema must be a rejecting control.
Keep this a small preflight shared by the selected runners, not a release platform
or a new cache. No stale deployed-code failure was reproduced in this review.

Source: [isolated runtime helper](../../journeys/191-generic-runtime.js),
[deployed checker](../../scripts/check_deployed_api.py).

## Preserve what works; keep maintenance finite

- The required verdict boundary now requires exact named Go run/pass and package
  pass, rejecting skips, empty matches, malformed/torn output and failures. Keep
  its shared JS/Python negative fixtures; do not reopen the fixed 237 defect.
- Hegel state/swarm tests over actual implementations and deliberate negative
  controls remain valuable. Add the I1 malformed-Go-byte boundary generator and
  writer/indexability properties; valid string generators alone cannot establish
  rejection of malformed UTF-8. Swarm rule subsets are not concurrent execution.
- The 242 metadata/content distinction is explicit and tested: checksum/stat checks
  are not arbitrary payload authentication. Preserve those limits; do not weaken
  corruption assertions to obtain a passing index migration.
- Keep model generation/MBT, observed-trace conformance, UI exploration and executable
  storyboards as distinct tools. One passing safety model is not a replacement for
  pointer hit testing, backend invariants or a settled collector.
- The gotogether integration remains a pinned pilot with retained restoration
  recipes, not a dependency of the root runtime or an automatically required tier.
  This review did not fetch its current HEAD or validate newer APIs.
- Passing campaign screenshots/full copied stores should not accumulate by quota.
  Keep current documentation frames, concise receipts/manifests and small useful
  replay ledgers; retain minimized failing reproductions until their fixes land.
  No broad new campaign or evidence-retention subsystem is recommended.

## Scope and evidence limits

No browser, collector, filesystem durability or race campaign was rerun here. The
243 complete task and endpoint probes are parent/team observations; source/CI/
generated-contract inspection and the single-event model replay are independent
checks performed for this report. Temporary synthetic replay files were removed.
The report proposes no runtime changes and does not certify production readiness.
