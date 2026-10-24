# Whole-product architecture review — 2026-10-24

Reviewed baseline: `86e35a5`, with concurrent implementation intentionally excluded from this assessment. This is a source and documentation review, not new live-stack acceptance. Recommendations feed the revised roadmap and backlog; historical review documents are evidence, not parallel work queues.

## Architectural direction

Goo is an interactive observability workspace over standard telemetry, with a Go semantic core that makes questions, populations, measurements and retained evidence reusable. Its strongest organizing flow is live Activity → scoped population → explicit comparison or capture → durable investigation. Keep that flow visible while retaining accurate distinctions between transient observations, fixed queries and immutable captures.

The code already has useful layers. Preserve them rather than replacing the project with a universal query runtime or universal workspace owner:

- `exploration/` owns bounded record questions and population semantics. A displayed page is not a complete population.
- `metrics/` owns native stream, reset, temporality and distribution semantics; `measurements/` composes supported native and lifecycle measurements.
- `measurement/` compiles portable prepared-fact definitions. Its boundary is different from arbitrary ClickHouse SQL or native metric execution.
- `annotations/` describes instrumentation and operation planning; authored investigation meaning belongs to the workbench domain.
- Immutable files, expiring process-local capabilities, browser retry intent and optional XTDB custody have different authority and failure modes. One store abstraction must not erase these differences.
- Passive custom elements receive data; application adapters own source requests, navigation, polling and publication.

Standard OTel vocabulary is a discovery aid, not proof of compatible type, unit, scope, stream identity or runtime coverage. Keep physical adapters explicit and generated Databricks semantics restricted to supported, validated translations.

## Reconcile the existing plans first

Several earlier recommendations are delivered: bounded shared source transport (`internal/sourcehttp`), startup source dependencies, shared immutable-file mechanics, the OpenAPI contract and conformance fixtures, source-free saved presenters, grouping controls, and exact Activity retention. Do not reopen them merely because an earlier review says “next.” In particular, `docs/PRODUCT_IMPLEMENTATION_PROPOSAL.md` contains older tail-retention and publisher priorities superseded by later implementation and validation entries.

Maintain one current milestone/status index with delivered, active, deferred and blocked states. Preserve historical acceptance records and link them from the current plan. Review completion against code and `VALIDATION.md`, not the age or repetition of a proposal.

## Ranked implementation candidates

### 1. Make immutable publication success truthful

Evidence: `cmd/goo/measurements_http.go`, `cmd/goo/native_storage_contract_test.go`, `internal/immutablefile/publish.go`.

The native package POST currently accepts an already-existing hash-named file without validating its contents. A corrupt winner yields 201 success, while reopening the registry fails. The storage contract test explicitly preserves that historical behavior. This is a real reliability issue, not a reason to silently change the common publisher.

Bounded change: validate the existing native package winner before reporting success. Preserve valid bytes and refuse a corrupt or mismatched winner without overwriting it. Make the HTTP contract change deliberate. Keep shared publication mechanics agnostic about domain identity. `PublishFileOnly` also intentionally omits directory sync; do not claim power-loss durability beyond that contract.

Acceptance: fresh publication, valid duplicate, corrupt duplicate, wrong-identity valid YAML, concurrent duplicate and fresh-registry reopen cases; no orphan scratch files or overwrite. Update frozen contract and API documentation as applicable. No source read is necessary.

### 2. Recover ordinary investigation drafts safely

Evidence: `ui/documents.js` reports unsaved drafts retained only in the page; `ui/investigation-trail.js` retains ordinary entries in memory. Native retention retry journals solve a different problem.

Bounded change: persist versioned authored draft intent with explicit recover/discard after reload. Preserve pinned references and display settings; reopen paused, without telemetry requests or automatic publication. Bound size and handle malformed storage, quota denial and multi-tab conflicts. Choose and document a conflict policy before shipping; local recovery is not collaboration.

Acceptance: reload recovery, explicit discard, corrupt/version-unknown entry, unavailable storage, size limit and competing-tab behavior. A recovery never substitutes local rows for durable evidence, resumes a source subscription, or silently saves a revision. Add a real browser journey and Bombadil actions/invariants.

### 3. Expose trace-anchor continuity without claiming completeness

Evidence: `exploration/trace_examples.go`, `ui/trace-journey.js`. Example selection is bounded and trace finding captures are later explicit reads. The interface correctly warns that anchor retention and trace completeness remain unverified.

Bounded change: derive selected-anchor membership from the actual retained capture pages when available: present, not found in retained page, or unavailable. Preserve the original observation and later capture bases. Do not infer deletion, exhaustive trace coverage or historical membership from a page miss, and do not add hidden rereads.

Acceptance: retained anchor, later page miss, source failure, partial capture recovery, context change during capture and distinct logs/spans captures. Show clear provenance in the finding cards; keep immutable old bytes unchanged. Test exact identities and truncated pages, not only the positive journey.

### 4. Extract the pure question/link protocol

Evidence: `ui/workspace.js` combines navigation mounting with question construction, evidence difference, record/native step validation, hash link parsing and fixed-window equivalence. `docs/QUESTION_PROTOCOL.md` and existing cross-layer fixtures already describe a domain boundary.

Bounded change: move pure versioned question/step/link/evidence transforms to a DOM-free module, leaving compatible exports at existing callers. Preserve old links, byte limits, defaults, exact timestamps and malformed-input rejection. Do not combine this with schema tightening, digest changes or query rewrites.

Acceptance: existing protocol fixtures pass unchanged, old URL corpus round-trips, malformed links remain rejected, import has no DOM/source side effects. Browser traversal and bookmark reopening remain compatible.

### 5. Consolidate field choice without merging field semantics

Evidence: `ui/numeric-measures.js`, existing grouping controls and numeric planning documents. Grouping and numeric measure authoring both select qualified fields, but grouping modes/two-field limits do not belong to numeric bindings.

Bounded change: reuse qualified-field parsing, advisory metadata and a narrow supplied-data field-choice primitive. Preserve numeric type/unit/parse decisions as explicit authoring inputs. Keep options scoped to the selected namespace and current inventory; stale selections must be visible and non-executable until resolved.

Acceptance: absent field metadata, resource/attribute same-key ambiguity, namespace switching, previously saved fields, keyboard interaction, server-rendered/detached component use and no component-owned network request. This is the active implementation direction, not a request to reopen delivered grouping work.

## Demand-led semantic work after those cuts

Typed numeric cohort predicates are a useful general observability primitive. Today `exploration/model.go` limits ordered predicates to supported uint64 fields; arbitrary attribute/resource filters retain string semantics. Numeric aggregate bindings do not make arbitrary numeric ranges available. Implement Go validation, typed missing/empty/invalid handling, overflow/order behavior, compiler/reference parity and semantic properties before introducing chart brushing or numeric sliders over these fields. Start with explicit int64/uint64; float, decimal, inferred units and conversion need their own decisions.

Native metrics currently require nonempty `service.instance.id` in supported kernels (`docs/E1_ACCEPTANCE.md`). Do not unblock generality by inventing an empty shared instance: that can merge distinct streams. Improve eligibility diagnostics first; any relaxation requires a stream-identity policy, full resource/scope identity and reset/temporality property tests.

Keep record trace navigation, native metrics and prepared-fact portability distinct until a real use case establishes a common semantic operation. DataFusion, an external bitemporal store or a general query planner would not remove these contracts. Existing optional custody integration should remain optional; local annotation intent and native measurement evaluation do not require a new database.

## Operational and deployment reliability

`docs/DIAGNOSTIC_RETENTION.md` distinguishes raw diagnostic recordings from retained evidence. Its cleanup is a soft budget and has no scheduler; direct test invocation can bypass wrapper cleanup. ClickHouse console rotation in Compose does not itself bound internal system log tables or server file logs. Current disk pressure makes an explicit diagnostic-storage budget and read-only size reporting worthwhile. A later configuration change should target diagnostics individually and prove telemetry, packages, captures and custody are untouched. Do not use evidence deletion as build hygiene.

Static/native deployments and image builds can diverge in a constrained environment. Record the actual served code/assets and verification target for every milestone; a successful source test alone does not prove the browser is running it. Broader deployment-status instrumentation should follow an inventory of existing health/build reporting rather than adding another overlapping status surface.

Browser proxy and direct Go endpoints remain separate transport boundaries. Choose concrete discrepancies with observed evidence and update both routes plus contract tests; do not use “unify API” as an unbounded rewrite. Shared-environment authentication, tenancy and production rollout remain a separate deployment gate, not an implied authorization in this review.

## Review and rollout criteria

Each cut should have a narrow diff, domain-owner review and explicit negative case. Semantic changes need Go reference/property parity and Hegel or Quint observed-trace checks where they describe a state machine; display-only extraction does not justify inventing a new formal model. User-visible changes need executable browser stories, affected screenshots and current Bombadil actions/invariants. A source-free component must remain source-free under detached updates and remounts.

Record actual commands and failures/skips in `VALIDATION.md`; do not claim a fresh full-stack check from prior evidence. Keep current roadmap status, README/design diagrams and storyboards aligned with delivered behavior. Commit reviewed milestones individually and avoid broad container builds or concurrent browser suites under current disk constraints. No external service instance or production deployment is required for the proposed first cuts.
