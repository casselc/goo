# Independent roadmap and process reconciliation — 235

Reviewed 2026-10-28 against `5d12ce9`. This is planning input from the current
independent reviews, not another execution queue or a claim that proposed fixes
are shipped. The user's standing implementation authorization is unchanged.
Only this report was written; no runtime, shared roadmap, fixtures, dependencies
or stored data were modified. No tests or live probes were repeated by this
planning review.

## Inputs and judgment

Read AGENTS, current roadmap/backlog/validation/issues, the post-234 review,
Oct25 planning/product/verification reviews, the older consolidated implementation
proposal, and the Go/frontend/API-usability library reviews. Reconciled all five
new independent reports after they were available:

- [Product/usability](PRODUCT_USABILITY_REVIEW_235.md)
- [Architecture](ARCHITECTURE_REVIEW_235.md)
- [Frontend](FRONTEND_REVIEW_235.md)
- [Go/semantic libraries](GO_REVIEW_235.md)
- [Verification/process](VERIFICATION_REVIEW_235.md)

The refreshed P1–P4 user outcomes are useful, but the initial order overweights
visible density relative to continuity of scope and inspection. The new reviews
found specific failures in the current task, not a need for more chart types or a
universal architecture. Add a short correctness/reliability gate, then continue the
existing product loop: observe → isolate → inspect supporting signals → preserve
and reuse. Keep completed 217–234 deliveries closed at their documented scope.

## Evidence-based triage

| Finding | Evidence classification | Disposition |
| --- | --- | --- |
| Incomplete JSON upload retains admission after cancel | Architecture reproduced actual Budget/Decode path: cancel 202, context cancelled, next request 429 until uploader closes, then 200; Go independently confirmed the I/O mechanism. Configured 60-second expiry was not waited out. | Accept an early bounded ingress fix. Do not describe it as a demonstrated Internet exploit or erase normal browser-abort/body-consumption acceptance. |
| Logs → Spans drops a valid service predicate and reads broader data | Product live request capture; frontend source corroboration. The shared predicate is not an incompatible log-only field. | Accept as core interaction correctness ahead of density. Preserve or explicitly stage predicates; no silent broadening. |
| Inspect opens details below the mobile viewport without a focus handoff | Frontend keyboard probe at 390×844: inspector y=901.09–1354.39, opener retains focus; source identifies app/presenter handoff. | Accept as part of the immediate Activity continuity gate. A smaller button is insufficient if its outcome is not brought into view. |
| Browser-to-Go ledger helper accepts no matching test | Verification reproduced Go exit zero for a nonexistent name and inspected helper's exit-only check. Its scan found no missing current literal test declarations. | Accept a small required-verdict fix. Do not claim previous ledgers were shown to skip. |
| Different malformed prepared string labels collapse into one group | Go public API experiment accepted 0xff and 0xfe labels with distinct valid subjects, yielding one group/count 2. | Accept a bounded literal-domain correction. It is not residual hash cleanup; likelihood is conditional on malformed Go input. |
| Native publication can cross its catalog's 256-YAML reader bound | Architecture source-confirmed producer/reader mismatch; no endpoint-capacity reproduction was performed. | Accept a contract defect and isolated regression task before further revision-producing features. Do not call it a live reproduced failure. |
| Exact occurrence bookmarking requires optional retention plus a tag/note | Product inspected a configured live selection; unconfigured behavior and reread Capture distinction are source/document verified. | Treat as a product/capability decision, not automatically a bug or a mandate for XTDB. |
| Ordinary OTel without deployment.environment.name is inaccessible to the exact environment selector | Product source-verified adoption precondition; no unscoped live read was attempted. | Resolve explicit onboarding/selector policy. Do not silently widen isolation or conflate missing, empty and named values. |
| Promotion holds a bank-wide lock across filesystem publication | Go source-verified serialization risk; no production stall benchmark. | Keep a named reliability follow-up with a keyed in-flight/reservation design, after the immediate gate. |
| Source-family ClickHouse cancellation/memory settings differ | Architecture source-confirmed difference; continued execution is unverified. | Require a targeted real source conformance probe before deciding a settings/helper refactor. |

These findings have different likelihood, evidence and dependencies. They should
not become ten undifferentiated P0s or an indefinite backend cleanup program.

## Proposed execution order

### Gate 0 — Small required reliability and verdict corrections

1. **Inbound receive boundedness.** Define a body receive deadline/cancellation
   policy separately from source deadlines. Prove it with a real partial TCP body
   and a complete JSON value whose body never closes: cancel/timeout releases
   admission while the peer stays open, starts no source work/publication, and
   permits a next valid request. Preserve strict size/trailing-data refusal,
   cancellation tombstones and SSE response lifetime. Cover direct Go first;
   inspect/probe Python ingress separately rather than assuming parity. A context
   timeout or synchronous Body.Close alone is not demonstrated to fix the defect.
2. **Required browser-ledger verdict.** Give `verifyWithGo` the named JSON-event
   pass/skip/fail contract already present in the optional runner. Reject a missing
   name, unset ledger skip, skipped child, malformed/torn output and failed model;
   one real existing ledger must still pass. Share the verdict contract/fixtures
   rather than invent a second testing platform. This can proceed independently
   of the transport fix; neither requires all historical campaigns.

Finish these as separate reviewable cuts. They are finite trust boundaries, not
prerequisites for a production authorization/tenancy system.

### Gate 1 — Activity intent and inspect continuity

Before the cosmetic P1 cut, execute one operator task: scope logs to service A,
switch to spans, inspect one observed record, return, and resume deliberately.

- Preserve valid shared service/resource/namespace predicates. Make incompatible
  signal fields explicit and stage any removal before executing a broader query.
  Include nested AND/OR, missing/present-empty namespace, saved drafts and late
  old-signal bodies. Assert actual API selectors and source counts, not chips alone.
- Add an application-owned inspect reveal/focus/announcement/close handoff. Keep
  the reusable inspector passive; property updates must not all become modal or
  steal focus. Keyboard and pointer behavior should be deliberate. Close returns
  to a valid opener or fallback after withdrawal/eviction.
- Test 320/390/700px and desktop, including visible outcome and retained selection.
  Keep existing pause/resume/cancellation ownership. Add one adverse scheduling
  case where it catches a distinct bug; do not add a new formal model per button.

This expands P1's acceptance unit from row geometry to a complete human task.

### Small integrity fixes — bounded, not a renewed cleanup track

- Reject malformed prepared literal strings before they participate in filtering,
  grouping or subject identity; review authored predicates/enums for the same text
  domain. Preserve valid Unicode, U+FFFD, empty values and supported outputs. Test
  direct Go inputs and excluded facts. No normalization or universal digest API.
- Define native catalog capacity admission or bounded pagination before publishing
  another revision can make the complete reader fail. Test built-ins plus saved
  physical files, duplicates, distinct saves at capacity and concurrent publishers.
  A racy count-before-publish is insufficient. Preserve existing bytes/IDs/winner
  checks; do not delete an existing over-limit store to make the test pass.

Both are accepted bounded work. They need not block the Activity scope/inspect
fixes. Resolve catalog policy before further revision-producing features, and the
text-domain defect before expanding portable compiler/evaluator promises.

### Product queue after the gates

| Existing item | Revised acceptance/priority |
| --- | --- |
| **P1: common scoping and density** | Follow scope/inspect correctness with service/namespace refinement discoverability, repeated row/field command consolidation and toolbar chrome. Measure the same supplied records. Treat keyboard traversal, meaningful selection and trace access as part of density; do not merely hide buttons or strip keyboard access. The 340-control probe included disabled controls and is not an exact tabbable count or WCAG finding. |
| **P2: find saved work** | Retain as the next broad product feature. Define bounded metadata I/O/allocation, integrity timing, query-bound cursors and publication/rebuild policy first. Search/order must cover the admitted catalog. Never sort one page as global recency, infer service from the first predicate, or load/validate every full capture body on every query. Direct-ID reopen and content verification remain explicit. |
| **P3: comparison support** | Show exact side counts/denominators and unavailable coverage first. Thresholds are explicit options; rare first failures should not disappear by default. Totals/receipts stay unchanged, and no significance/confidence claim is added. |
| **P4: recurring trace question** | Keep behind the named slow-database-plus-error task and versioned saved-question boundary. Preserve different-span and same-observation witnesses, condition-1 parent scope and source-free reopen. No parallel trace store or witness-page/full-trace promotion. |

Do not require every medium reliability recommendation to finish before P2 design
can start. Keep each implementation's dependencies explicit rather than letting
review volume determine the queue's length.

## Product decisions to resolve early, without inventing a backend dependency

**Environment adoption:** choose a supported, explicit path for useful OTel that
lacks the optional deployment environment resource attribute: documented checked
enrichment, a missing-environment selector, or reviewed source/environment
discovery. At minimum expose the current prerequisite and useful empty-state
next steps. A non-demo fixture lacking the attribute must exercise the chosen
policy. No arbitrary environment string, wildcard fallback or silent aliasing.

**Bookmark versus annotate:** choose whether keeping exact observed occurrences
is baseline or an optional capability, and communicate availability before a user
invests in selection. Distinguish bookmarking from asserting a tag/classifier
label. A baseline fallback, if selected, must preserve server-attested membership,
immutable bytes, expiry and retry guarantees; Capture's honest reread is not the
same object. This is not approval to make XTDB mandatory or to let client-supplied
rows claim source custody.

Resolve these product contracts while the early fixes proceed. Do not make the
whole capture-discovery task wait for a universal selector/store redesign.

## Architectural and process dispositions

**Keep:** typed source/OTel and numeric semantics, explicit runtime/binding and
provider ownership, source-free supplied evaluation, concrete checked reference
APIs, passive framework-neutral elements, application query/navigation owners,
immutable publication ambiguity/winner rules and exact clocks. Existing useful
profiles are delivered, not automatically the next dashboard backlog.

**Named later work:** bank promotion lock isolation with preserved in-flight
identity; prepared Plan metadata detachment when that API is next touched; shared
plot keyboard navigation and compatible-update focus policy for a named live
widget; measured row-projection cost; source-policy conformance; a small required
core browser/native tier with explicit deployment identity and disposable stores.
The served OpenAPI matching the checkout does not prove the running binary and
copied UI came from that commit. Existing candid deployment records remain valid.

**Stop/defer:** blanket generic-hash rewrites, framework/runtime migrations,
universal query/store/history layers, automatic fleet health, broad new convention
profiles and mandatory bitemporal storage. Duplicate retention status calls are a
small measured enabler, not a higher-value main milestone than scope/inspect.
Current module/DOM counts alone do not establish a performance defect.

**Clarify authority cheaply:** `PRODUCT_IMPLEMENTATION_PROPOSAL.md:3–19,77–82`
still calls its older table the single active queue and a tail publisher “next.”
`IMPLEMENTATION_PLAN.md:3–5` points to the current roadmap, but `:15–18` immediately
calls Oct25 current. Add explicit historical/current pointers to these sections;
retain delivery history. Do not turn this into rewriting 3,000 lines of historical
issues/proposals or another parallel queue. The current roadmap owns order and a
single bounded backlog owns acceptance; specialist reviews provide evidence.

**Verification policy:** require named outcomes where a tier is required, retain
passed/failed/skipped distinctions for opt-ins, and make fast/unit versus real
browser/source/model claims explicit. Introduce a small selected core spine with
skip refusal and `--forbid-only` when its deployment/store prerequisites are ready;
do not run all 259 historical specs or triple every campaign across engines.
Future Firefox/WebKit support claims need representative acceptance, not a forced
framework/fallback rewrite today. Preserve minimized failures/tool/model versions;
large passing diagnostic bundles are not a product prerequisite.

## Next whole-product gate and stopping rules

After Activity continuity and the first capture-discovery slice, re-run one full
ordinary-OTel task through scope → signal switch → record inspection → supporting
trace/context → preserve → navigate away → find → reopen → deliberate fresh Run.
Include configured/unconfigured capability states and one cancelled/late body.
Assess whether the user can explain which population, clock, grain and preservation
method each result represents. An entrance-only screenshot pass is useful visual
input, but insufficient acceptance for that complete loop.

Advance when the bounded defects are fixed and the named task is coherent; do not
wait for all architecture/style proposals. Conversely, do not declare the core
loop complete from a large green unit count or compact entry frame alone.
