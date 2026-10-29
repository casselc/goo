# Independent review reconciliation — 243

Reviewed 2026-10-29 at `65ac193`. Five domain agents and a planning agent each
prepared their own report. The parent reconciled them with an actual combined
user task and temporary-store probes. This is review/planning work, not a runtime
fix, production load test, full accessibility audit or new vendor survey.

## Reports

- [Product/usability](PRODUCT_USABILITY_REVIEW_243.md)
- [Frontend/components](FRONTEND_REVIEW_243.md)
- [Architecture/reliability](ARCHITECTURE_REVIEW_243.md)
- [Go/semantic libraries](GO_SEMANTICS_REVIEW_243.md)
- [Verification/process](VERIFICATION_PROCESS_REVIEW_243.md)
- [Planning reconciliation](PLANNING_RECONCILIATION_243.md)

All six were consulted; the planning report incorporates all five domain reports.
They agree on the core purpose and the first correctness gate. Their priorities
otherwise reflect different domains; the [new roadmap](../ROADMAP_2026_10_29.md)
and [acceptance backlog](../BACKLOG_2026_10_29.md) are the single execution queue.

## Findings that change the queue

| Finding | Evidence/confidence | Integrated decision |
|---|---|---|
| A valid capture commits but cannot fit its metadata entry, disabling search; rebuild cannot repair it | Go reviewer reproduced through public Go and actual HTTP handlers; architecture independently checked the mechanism. Request12,728 bytes →201; receipt65,557 bytes; search503; exact-ID200; rebuild0/error. | **C0 first.** Prefer a bounded catalog-summary projection from the canonical declared scope. Preserve full immutable questions/IDs, successful commits and honest content verification on reopen. |
| Malformed prepared/authored strings collapse identity/grouping | Independent Go and parent probes:0xff/0xfe labels →nil error,one group,count2; valid U+FFFD/lambda →two groups. Go also reproduced invalid authored predicates/enums. | **C1.** Reject malformed text at the selected-fact/authored boundary; no replacement/normalization or generic hash migration. |
| Native publisher can make its own catalog unavailable | Parent's owned generic runtime:257 distinct POSTs all201; GET catalog400 scan limit; duplicate retry201. Architecture confirms source. | **C2.** One writer/reader capacity contract, with duplicates, built-ins, concurrency and over-limit recovery. No data deletion or silent truncation. |
| Mobile trace sheet covers the global parent-return button | Actual pointer hit at390×844 is a SUMMARY inside the full-screen trace sheet. **Normal visible Close trace → Return works.** Frontend/product reconcile this nuance. | Small parent-aware close/focus contract, not a declaration that all return is broken and not a global z-index patch. |
| Every missing metadata entry needs a whole-store rebuild capped at256MiB | Architecture source + sparse-file preflight probe; valid large-store failure not independently published. | Bounded per-ID or resumable repair. Keep query I/O small; do not repair with hidden full reads during search. |
| Dense component updates lose interaction state | Frontend's supplied-data gallery probe:256 timeline button tab stops; identical data replaces focused mark; identical latency summary resets disclosure/focus. Zero API reads. | Shared keyed update/keyboard conventions before expanding live widgets, not a framework rewrite or inferred WCAG verdict. |
| Fast CI does not run the complete browser task | Verification audited hosted workflow; strict Go verdict helper is correctly improved. | Select one real required task, after deciding its close/return contract. Other campaigns remain opt-in. |
| Generated GET /captures effects describe mixed source/publication | Architecture/verification source-confirmed path-only classifier. Handler is source-free. | Small method-specific documentation/conformance correction. |

The accepted metadata-only integrity boundary is **not** the C0 defect. A listed
entry still cannot prove arbitrary payload integrity without a read; exact reopen
and rebuild do verify contents. The failure is admitting a deterministic,
unindexable question and recommending recovery that cannot handle it. Missing
metadata due to an injected operational failure remains a different case.

## Integrated task evidence

The parent used an empty-directory Go runtime and an owned temporary store,
connected to the existing lab ClickHouse. Namespaced OTLP fixtures represented
ordinary HTTP server spans, a database-client child span and linked logs.

The task is Logs → service filter → Spans with the same predicate → Inspect →
cross-service trace/linked logs → close trace → return held rows → explicit fixed
analysis → fresh fixed-result capture → leave → title search → reopen unchanged
receipt. The initial direct header-return attempt was occluded on mobile. The
normal close-then-return path completed in 9s before adding surface diagnostics.
The final run including service, Definitions and Operations walkthroughs passed
in 17.0s (19.7s including runner setup). Close returned focus to `exploreRun`;
Shift+Tab from the trace close button reached outside the sheet. These support
the parent-aware close and keyboard ownership findings. The service frame also
shows repeated scope/evidence disclosures that the compactness work should unify.

| Action | Cumulative tail requests | Explicit explore/context requests | Capture requests |
|---|---:|---:|---:|
| Initial Activity |1|0|0|
| Apply service scope |2|0|0|
| Switch signal |3|0|0|
| Inspect/close |3|0|0|
| Trace and linked logs |3|2|0|
| Close sheet/return held Activity |3|2|0|
| Run fixed analysis |3|3|0|
| Capture fixed question |3|3|1|
| Leave/find/reopen |3|3|1|

These are browser source-triggering requests, not counts of internal SQL queries.
The trace explicitly states it broadens to all services within the original
bounds/environment. The saved artifact contains the original scoped fixed result;
visited supporting context is **not** implicitly archived with it. A separate
finding/document action is needed for that understanding. No selected row is
misrepresented as the fresh capture.

The [review probe](../../journeys/243-whole-product-loop.spec.js) is opt-in. It can
reproduce direct occlusion or continue through the visible close action. It is
not promoted to required CI or a completed pointer-return contract. Its capacity
mode intentionally observes the current defect in an owned store; it is not a
product contract endorsing257 publications. Future fixes need normal regression
assertions for their chosen policy.

Root appended only namespaced test telemetry and temporary captures/definitions;
existing live evidence, definitions, annotations, volumes and payloads were not
modified/deleted. Domain reviewers' pure/supplied-data/temp experiments have
separate limits in their reports. No broad suites or new runtime deployment were
needed for this review-only cut.

## Product and architecture judgment

Preserve single Go serving, semantic package layers, explicit instrumentation
runtime, checked references, passive web components, exact values/units and
concrete source owners. The239–242 improvements remain real deliveries; C0 reopens
one producer/discovery contract, not the entire saved-work or component design.

After C0–C2, settle two general-adoption decisions: explicit support for existing
telemetry without the optional environment label, and whether exact observed
bookmarks are baseline with optional interpretation. Do not use a magic environment
string, wildcard, client-asserted evidence, mandatory XTDB, or a reread disguised as
an original observation. Collector enrichment alone does not expose historical
unlabelled data.

Then consolidate reopened-evidence mastheads and child-return ownership; use the
existing support/context primitives to refine comparisons before adding thresholds.
Keep low-support/one-sided failures visible. Recurring multi-condition trace reuse
remains useful after these foundations. Per-ID repair, promotion-lock isolation,
detached compiler metadata and bounded plot update/navigation have explicit small
cuts in the backlog. No universal store/query/history abstraction, display framework
migration, arbitrary dashboard builder, new query engine or broad profile expansion
is justified by these findings.
