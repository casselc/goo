# Counted selection integration review — 425

Accepted for the bounded architecture/source review below. Actual browser,
exploratory, property and model execution results belong in
[validation](../../VALIDATION.md). This review did not run those checks.

The reviewer authored the separate pure planner and planner tests, then reviewed
root's composing changes in `ui/filter-editor.mjs`, `ui/filter-values.mjs`, and
`ui/stream-entry.mjs`. The shared editor tests described below use the existing
422 fixture. This is not independent acceptance of the reviewer's own planner.

## Intent and source ownership

The shared editor's optional Compare action returns a detached original counted
question, exact qualified field, all selected categorical values and the actual
opener. Values hidden by local search remain selected. Comparison does not
mutate the editor's predicate, clear its selection, invoke `onChange`, or turn
counted totals into comparison operands. No selection disables Compare. A
comparison-only AST/profile refusal preserves an otherwise legal Stage action.

Native readiness, current owner, generation, mounted controls and original
predicate identity are rechecked at invocation. The application adapter also
requires the current applied Activity scope and reconstructs its field lookup
question from fresh `Tail.scope()` metadata. A changed completed window,
environment policy or selector cannot be substituted for the counted question.
The adapter creates the existing fixed query-comparison view and stages it using
the established `onOpenFixed` owner. There is no new source path or grammar role.

The fixed baseline is the rest; selection is the chosen values. Both retain the
same parent filters, qualified grouping, fixed window and environment policy.
The declaration title follows that operand order. The generated Boolean
partition describes query intent; independent source reads do not attest a
simultaneous partition of a changing data source. Run remains explicit.

## Held return and authority

`FilterValues.mount` distinguishes descriptive retention from active read and
publication authority. During a held comparison visit, `retainedCurrent` keeps
the same ready result DOM, selection and opener in custody. Its ready callback
still requires both retained and active question identities. Hidden/suspended
controls cannot acquire or publish merely because the original result is held.

A suspended Load callback returns before pausing, invalidating or replacing the
held result. Pending responses require a current generation and active identity
before publishing either ready or error UI; a response during suspension is
dropped. Replacement invalidates the generation, preventing an identical later
question from reviving a retired action. Returning uses the existing parent
anchor and actual opener, with fresh active guards and no implicit source read.

The new code uses metadata-only scope snapshots rather than copying retained
rows for permission checks. Counted data remains descriptive; source reads,
navigation, parent custody and comparison evaluation remain application-owned.

## Focused checks and limits

Three focused tests extend `tests/counted-value-editor-422.test.mjs`:

- Detached handoff includes missing and empty values hidden by search, the exact
  original question and actual opener; selection and draft remain unchanged.
- No-selection, stale owner, mutable-owner refusal, closed disclosure, hidden,
  inert, disabled, aria-disabled, hidden document and retired callbacks refuse.
- A genuine rest AST-budget overflow disables Compare while Stage remains legal;
  the current comparison reason is checked again when invoked.

The focused file passes under `node --test`. Existing owner tests separately
exercise retained ready custody and late hidden response refusal. Real keyboard,
responsive frame, source-budget and retained-return acceptance remains part of
root's actual journey checks rather than being inferred from these source tests.

No unresolved blocker was found in this source review.

## Independent UX/accessibility source and frame acceptance

A second reviewer inspected the composing source and the actual
`counted-selection-controls-{1440,390,320}.png` and
`counted-selection-comparison-{1440,390,320}.png` frames. This reviewer authored
the descriptive-retention source-owner change and its focused tests, so that
part is not independent acceptance of the reviewer's own implementation. The
comparison UI, application handoff and generated fixed comparison are reviewed
independently. No browser or model suite was run by this reviewer.

Accepted within these bounded limits:

- Missing, present-empty and literal lookalike values have distinct native
  checkbox labels; selection counts and unknown coverage remain visible. Search
  uses the available width, and Stage/Clear/Compare controls fit at narrow widths.
- Compare has its own native control and remains separate from predicate staging.
  Its accessible description links to the selected-count status. A selected but
  unsupported comparison exposes its refusal in that visible status, rather than
  relying only on a disabled button's tooltip.
- The shared count comparison preserves readable paired values, a shared scale
  and a distinct signed change. The title follows operand order: rest on the
  left (259), selected values on the right (2), change −257. Generic fixed-window
  role labels remain operationally correct, while exact selectors, clocks and
  independent read facts remain in the existing native disclosure.
- Unknown upstream coverage stays visible. Run has a clear focus outline in the
  narrow frame; the application uses the existing actual-opener parent anchor for
  Return to paused activity. Root's executed journey, rather than this frame
  review, establishes keyboard behavior, exact source budgets and retained focus.

No remaining UX/accessibility blocker was found. The initial desktop frame showed
a full-list-width Compare control despite its short label. Root refined the shared
style to an intrinsic-width desktop control while retaining the full-width narrow
target. The final rendered follow-up is accepted below.
The 426 pass can consolidate action-row composition further without introducing
another interaction shell.

### Final responsive-control follow-up

Reviewed the regenerated `counted-selection-controls-{1440,390,320}.png` frames
(22:21 candidate) after the intrinsic-width desktop refinement. Accepted: Compare
now has a compact natural-width desktop target; the 390/320 layouts retain a
clear full-width target beneath the existing selection actions. All count labels,
missing/empty/literal choices, native checkbox targets and scope/coverage notes
remain legible. The CSS changes presentation only. No outstanding UX/accessibility
finding remains from this bounded review. Root separately reports the final real
journey and 35-second actor passing; remaining regression/model execution belongs
in validation and is not inferred from this frame inspection.
