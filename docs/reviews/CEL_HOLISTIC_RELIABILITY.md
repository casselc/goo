# CEL integration: reliability and workflow acceptance

Review scope: source inspection of the delivered stateless pilot and proposed app authoring bridge. This document proposes acceptance; it does not report executing new tests or shipping a UI. Read alongside [pilot](../CEL_PREDICATE_PILOT.md), [current roadmap](../ROADMAP_2026_11_14_CHECKPOINT_430.md), top of [validation](../../VALIDATION.md), and [known issues](../KNOWN_ISSUES.md).

## Three early decisions

1. **Give compile results an owner before adding asynchronous authoring.** `ui/question-session.mjs` already protects execution publication with monotonically increasing generation and semantic requested-question identity. It does not own CEL text compilation. A future compile needs an editor revision, captured parent revision, signal, library version and route/mount lease. Abort is cleanup, not correctness. An A→B→A edit must retire the first A result even when text and semantic question match again. `ui/stream-entry.mjs` already uses scope epochs for analogous retained observation callbacks; reuse this principle rather than identifying owners solely by serialized queries.
2. **Choose additive versus replacement intent explicitly.** `querycel/program.go:Apply` clones and validates the complete parent and adds its whole predicate under AND. That is implemented and valuable. Reapplying an edited expression to an already refined question would accumulate old clauses. Proposed initial UX: “Add condition to this question”, with a captured immutable parent. A replace mode needs an owned clause identifier/revision and an explicit base; never infer removable ownership from syntactic AST resemblance. Neither mode may silently remove environment policy, fixed bounds, grouping, projection, pagination or an existing OR branch.
3. **Put new verification effort at the app ownership boundary.** Existing `verification/querycel_test.go` exercises actual CEL versus Goo evaluation, parent immutability and source parity; `querycel/program_test.go` and `cmd/goo/plan_cel_test.go` own compiler/CLI refusals. The next risk is an apparently harmless editor gesture causing a telemetry read, write, stale staging or loss of held evidence. Observe actual outgoing requests, not only mocked callbacks or disabled-button attributes.

## Current implementation and reuse boundary

`querycel/compile.go` uses canonical pinned CEL, typechecks and admits resolved overloads, then lowers to the ordinary predicate grammar. Unsupported operations are refused even in dead branches. `Program.Eval` validates every referenced fixed column before evaluation: missing and malformed inputs are errors; present empty strings are values. `Program.Apply` returns detached data and revalidates the expanded parent. `cmd/goo/plan_cel.go` emits a source-free planning receipt only after compile, refinement and source planning succeed. There is no CEL editor, planning HTTP endpoint, attribute-map lowering, visual-to-text serializer or new evidence format in this slice.

The actual installed manifest declares Playwright 1.56.1, Bombadil 0.7.9 and Quint 0.32.0 (`package.json`). The separate verification module pins Hegel 0.9.8 and replaces quintgo with sibling `../../observability/quintgo` (`verification/go.mod`). These are declared dependencies, not a claim that every binary is available or runnable on this host. `verification/README.md` documents gotogether pilots explicitly as isolated work without a project runtime or dependency upgrade. Do not invent a gotogether integration or upgrade Hegel merely for CEL.

Existing reuse points are `question-session.mjs` for semantic question identity/execution ownership, `filter-editor.mjs` for pure validation and guarded mutation, `dom-visibility.mjs` for invocation readiness, and the application owners for reads/navigation. Keep display components passive. Existing Activity search434 and severity435 journeys/actors already demonstrate source-free staging, full-parent preservation, fixed-window custody and held return. Adapt their request ledger and observed replay patterns rather than copy a full parallel harness.

## Proposed state contract

Represent draft state separately from applied evidence. A minimal source-free authoring state is `{ownerEpoch, draftRevision, parentRevision, parentQuestion, text, signal, libraryVersion, phase, compiledRevision, compiledPredicate, refinedQuestion, diagnostics}`. Phases can be editing, compiling, valid, invalid and staged; existing execution/session state remains authoritative for running/applied/failed evidence. Clones cross boundaries. Semantic identity checks scope; revisions establish temporal ownership.

| Event | Required effect | Forbidden effect |
| --- | --- | --- |
| Type/paste/visual edit | Advance draft revision; retire compile/stage eligibility immediately | Telemetry/metadata discovery, capture/save, mutation of applied evidence |
| Compile/validate | Capture full owner tuple; produce bounded diagnostics or detached result | Execute a telemetry plan; treat prior successful compile as current |
| Parent/signal change | Advance parent/owner revision, invalidate result, preserve or explicitly discard text | Apply a logs result to spans; silently rebase onto changed scope |
| Compile completion | Publish only when full tuple remains current and owner is active | Late success/error overwrites current text/status; ABA resurrection |
| Stage | Revalidate complete combined AST and owner, stage ordinary question | Read, save, capture, start stream or overwrite held evidence |
| Explicit Run/Apply | Existing app owner checks current scope, dispatches existing typed question | Execute CEL source as alternate payload or bypass Session publication guards |
| Hide/route/dispose | Retire callback leases; cancel pending work; define focus fallback | Hidden/detached controls stage or read when invoked directly |
| Return/reopen | Retain original fixed bounds and held evidence; require current intent | Recompile/reexecute/capture automatically or use wall-clock “now” |

Do not require a new global coordinator. A small pure reducer/lease abstraction shared by authoring adapters is sufficient if more than one surface needs it. Do not merge Activity subscription lifecycle with fixed-question execution: their read gates differ. Editing a parent during compile should leave a recoverable “Question changed; validate again” state, not an invisible pending result.

## Prioritized acceptance

**P0 before app bridge acceptance:** add and replace semantics are specified; zero source/store writes and zero telemetry reads on typing, validation, mode switch and stage; stale success and stale failure are ignored; dirty-parent, hidden, disconnected, wrong-route and retired callbacks cannot stage; full-parent OR remains under AND; fixed bounds and typed environment policy survive unchanged. Actual run payload must pass current question admission and source planning. Failed validation must leave held evidence and text intact and produce no partial valid receipt. Compilation transport, if introduced, is explicitly source-free and counted separately from source reads.

**P0 semantic contract:** text→checked CEL→ordinary AST→existing compiler remains the only execution route. Compare semantic identity rather than object key order or exact pretty-print text. Visual/text round trips support only a declared intersection: current visual predicates exceed the admitted CEL subset. Unsupported visual ASTs must stay lossless/readable with an explanation, never drop a clause or silently approximate missing/empty semantics. Same text under another library version is not assumed equivalent.

**P0 budgets/data quality:** enforce UTF-8 byte limits at every ingress (browser character counts are insufficient), 4,096 expression bytes, 65,536 bytes per referenced string and 1,000,000 CEL work units. Preserve parser/AST limits and expanded ordinary 32-node/depth-4 budget including the parent. Test boundary and one-over values, Unicode, uint64 above JavaScript precision, reversed comparisons and SQL-looking literal parameters. Costs are not general memory/wall-time guarantees; future network compile needs a bounded request and cancellation contract. Missing referenced fields, noncanonical uint64 and invalid UTF-8 remain explicit evaluation errors even behind short-circuit branches. Reference failure is not evidence that the source matched zero rows. Attribute maps remain refused until presence/empty/null/error semantics and source lowering have a separate contract; never translate absent keys to empty strings.

**P0 time and custody:** freeze parent event-time bounds once; compile and stage cannot consult time or shift rolling windows. Keep event time, read time, frozen evaluation time and capture custody time distinct. Existing `ui/saved-capture-protocol.mjs` verifies immutable receipt scope and refresh predecessor/question; route CEL-authored ordinary questions through it. Do not mutate old receipts when text/library/definition changes. A planning receipt is not a capture or proof of source membership.

**P1 when saving authoring metadata:** preserve expression bytes, CEL library/profile version, lowered canonical AST and parent/definition revision as distinct facts if the product promises reauthoring. Specify content identity and migration before persisting those additions. Existing captures need no speculative metadata retrofit. Reopening a versioned definition must not compile historical text with the latest library and relabel prior evidence. A refreshed capture keeps original question and predecessor while advancing evidence basis through existing protocols.

**P1 product completion:** label whether users are adding a condition or replacing an owned condition; show parent scope and staged/applied status. Diagnostics identify parse/type/unsupported/budget/scope-change categories, provide a recovery action and avoid SQL/internal exception leakage. Associate errors with the editor, announce status changes without every-keystroke live-region chatter, retain text and selection, support keyboard stage/run and return focus. Verify desktop, 390px and 320px without horizontal overflow. Physical assistive-technology and cross-engine results must be separately reported.

## Small real storyboard and test matrix

Use one fixed logs question with environment policy, `(service=api OR service=worker)`, fixed UTC bounds and a retained result. Seed actual OTLP cases through the existing lab journey infrastructure: matching Error+timeout, Error without timeout, Info carrying misleading FATAL text, wrong environment/service, literal Unicode/punctuation and a record outside the window. Include more source matches than the retained page so a preview cannot masquerade as a full population. Independent source assertions count membership; screenshots explain the user path.

| Frame/action | Observed acceptance | Primary tool |
| --- | --- | --- |
| Held parent → open editor → type valid expression | Held row keys/basis unchanged, zero telemetry reads/writes | Playwright request ledger |
| Validate → stage Error band + literal contains | Current revision, complete parent AND refinement, identical bounds; still zero reads | Pure reducer tests + Playwright |
| Run explicitly | Existing typed API payload and parameterized plan; source membership agrees with independent oracle | Existing real OTLP/ClickHouse journey |
| Invalid edit and repair | Text/evidence retained; categorized accessible diagnostic; no partial stage | Node/controller + keyboard Playwright |
| Delayed A, B, A results; parent dirty; navigate away/back | Old successes and failures rejected, retired direct handler invocations harmless | Deterministic deferred-promise/controller tests |
| Return to held parent | Original keys/basis/window retained; zero rereads | Existing held-return journey helpers |

Add one short Bombadil actor only after the real authoring surface exists: bounded type/validate/stage/reopen actions, stale/hidden controls and explicit-run gates, with positive completion requiring a successfully run current revision. Keep independent actor reads distinct from main storyboard counters. No actor is warranted for the current stateless CLI alone.

A small proposed Quint model should include owner epoch, draft/parent revision, compile token, active/visible status, staged identity, source-read and write counters. Observe request and publication events from the actual browser trace; assert no-read local macrosteps and exact run ownership. Model sampling does not prove arbitrary scheduling or accessibility. Avoid adding a second abstract CEL evaluator: Hegel already compares actual CEL/Goo paths. Add generated draft-event sequences only when the reducer exists.

Required negative controls mutate a completed admitted trace or result: increment a source-read counter during stage; admit first-A after A→B→A; drop one OR-parent branch; alter end bound/environment policy; accept stale error over repaired text; replace numeric Error with SeverityText; substitute retained-page count for source total; alter captured question/library provenance. Each must fail the relevant admission/property. Rejection must be attributable to the intended invariant, not merely malformed JSON. The fixture must also admit its unmodified positive control.

## Sequence and evidence discipline

Before UI: keep focused root compiler/CLI tests, current Hegel CEL/Goo/parent properties and real read-only ClickHouse parity as the semantic gate; specify round-trip intersection and owner/add/replace contract. For the bridge: pure deterministic lifecycle tests and one real storyboard with request counters are essential. Add observed Quint replay and the bounded actor once state is observable; preserve existing adjacent search/severity/filter regressions instead of a second CEL-specific app stack. Persistence tests become essential only when authoring persistence ships.

Retain one compact admitted report with question identities, event revisions, measured request/publication counters, independent source facts and actual versions; use the normal journey screenshots for a few meaningful frames. Do not archive every AST edit, whole telemetry stores or duplicate screenshots to create apparent assurance. Record exact commands, skips and environment limitations in validation when implementation occurs. No new checks were executed for this review; existing validation claims remain historical evidence, not fresh reruns.
