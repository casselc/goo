# Side-specific group staging architecture review — 292

2026-10-31. Read-only review of the delivered291 group profile, existing population
child ownership and the292 roadmap. This review runs no browser or semantic tests;
executed checks belong in [VALIDATION](../../VALIDATION.md).

## Bounded direction

An explicit side-specific staging intent fits the existing retained-child owner.
The passive group display may identify a requested side and category, but it must
not execute a query, navigate, select a baseline or manufacture source authority.
The application checks its admitted prepared pair and derives the chosen fixed
question. Staging performs no source or storage read. The child requires explicit
Run, and the original parent remains held and returnable.

Do not adapt the existing ranked `groupIntent` to a summary Dataset by weakening
its shape checks. Group-count comparison summaries describe independently held
whole counts, while the aligned rows live in their own closed profile. A bounded
side/key/context intent is appropriate, using the existing
`datasetId`, `revision`, `scopeId`, `generation` vocabulary. Here `generation` is
the snapshot scope generation; the private owner/read epoch is a distinct value. It needs no query-bearing chart child, new API, stored object or universal action registry.

Existing `scope-actions.mjs` deliberately refuses the baseline role and summary
slot for generic refinement. Preserve those rules. The new bounded adapter stages
a baseline question intent; it does not turn baseline summary data into a generic
H1 source-refinement capability.

## Required authority checks

The component's capability is opt-in and supplied by a connected application
owner. Standalone canonical data remains display-only. A valid-looking snapshot,
context tuple or DOM event is insufficient by itself. Before deriving a question,
the host must check all of the following:

- The exact currently admitted prepared pair and active held parent result are
  still current. Baseline ID, side dataset identity/revision/scope generation,
  original source basis, chosen root question and group category must agree.
- The group comparison and selected side are coherent. The selected row has
  present membership and a known positive count. Diagnostic, unknown and aligned
  absent-zero cells supply no staging action in this first profile.
- The comparison role, group view, mount and requested control are connected and
  visible. Hidden or detached old controls cannot be activated programmatically.
- The parent has no child, is below the existing depth bound, and is not reading,
  preparing a baseline, paging or holding pending record/facet/query/source work.
  Existing draft and record-selection guards remain authoritative.
- The event is an exact closed intent with a known side and canonical key from the
  supplied comparison; the host derives the value/field from its own admitted
  pair, not an event-supplied query or predicate.

An explicit positive present count on one side is not proof of upstream telemetry
completeness. These checks establish a current scope operation, not source
attestation or a claim that the resulting query will return the same records.

## Query derivation

Validate the original selected-side query first. Clone it, reset only inspection
offset to zero and combine the existing predicate with exactly one group clause.
For null category values, use `missing`; for the present empty string, use an
explicit `eq` with empty value. Preserve the original signal, typed environment,
half-open window, grouping, measures, bucket width, limits and all other definition
fields. Existing OR structure must remain intact under a new conjunction. Validate
the final query against predicate depth/node/size limits; refusal is preferable
to dropping or weakening the original expression to make room.

The baseline side stages its original fixed question, not a filtered retained
record bank. Label the child as an independently runnable question: a later read
may differ because of arrivals, retention or source changes. Do not promise exact
captured rows or silently replace the baseline's window with the current one.

## Parent return and retirement

`openChild` currently pauses live/source owners, increments generation and hides
the parent. Existing source callbacks must stay retired after that boundary.
The group element is retained across ordinary parent renders to preserve local
Find/page state. Therefore action ownership cannot simply be attached once to
that cached element and dynamically accept whatever generation exists later.

Opening the child must revoke the existing group action owner. Back may establish
new current controls/lease for the same held pair, while preserving passive local
presentation. Old captured buttons and intents must remain refused even after
return; new current buttons may stage another child without rereading the parent.
Local display-only state may survive independently of this action lifetime.

Each rendered control must capture the action binding that created it. Do not let
an old button read a newly assigned `element.actions` callback at invocation.
Rebinding should create fresh controls; old callbacks retain their revoked owner.
A generic forged DOM event with a matching public context must not be the authority
bridge. Property-only callbacks/opaque runtime owner handles are suitable; JSON
attributes and serialized canonical data cannot grant staging capability.

## Verification gates

Use Go/browser property parity for exact query derivation over both sides,
windows, typed environments, null/empty/ordinary values and nested predicates.
Check no mutation and full-definition preservation. Browser journey must observe
staged child/no reads, explicit Run with exact chosen scope, held parent Back,
then refusal of retained old controls and success of newly mounted controls.
Cover altered side/key/context, hidden views, baseline replacement, source refresh,
record selection, pending page/preparation and diagnostic/absent states. A clicked
button or conforming trace prefix is not a completed journey.

## Disposition

The bounded direction is sound subject to the final contract matching these
ownership and derivation constraints. Proposal and implementation review remain
pending. This document makes no delivery or test-pass claim.

## Final proposal review

Reviewed [the complete proposal](../GROUP_SIDE_STAGING_PROPOSAL.md). Approved for
the bounded implementation: closed `scope-group-side` intent, explicit side and
summary context, positive present/coherent eligibility, exact checked query
conjunction, no implicit reads and fresh post-Back action controls. No general
H1 baseline or summary authority is broadened.

The proposed property-only component interaction binding is suitable if each
rendered handler captures its binding, comparison and private action epoch.
Rebind, redraw and disconnect must revoke prior controls. The application still
owns the actual eligibility/current-state check and query derivation; emitting an
observational event is never sufficient authority. Final source and actual-frame
review remain pending, with executed checks delivery-owned.

## Go and pure browser planner source review

The implemented Go `StageGroupSide` reconstructs the already validated canonical
comparison through `CompareCohorts`, requires the exact selected summary tuple,
refuses diagnostic/absent/nonpositive membership and decodes a detached selected
selector before adding the conjunction and validating the final query. The closed
JSON intent decoder rejects unknown fields. The pure browser planner follows the
same selection and final-validation boundary. Neither performs IO or grants
application action authority. No source-level query derivation blocker found.

The existing empty-string wire default is retained: Go may omit an empty `value`
field for `eq`, while the browser can retain `value: ""`. The operator distinguishes
that predicate from `missing`; no global predicate marshaler or semantic filter
normalization was widened for this feature. Generated parity checks are owned by
the verification agent.

## Application and component source disposition

Reviewed the component and host integration. The element defaults to no staging
interaction; its opt-in callbacks are property-only. Each native control captures
its comparison, callback binding and monotonically changing action epoch. Redraw,
rebind and disconnect invalidate old controls. Strict intent construction runs on
activation rather than validating the entire comparison repeatedly for each
rendered button.

The application checks a private active/generation lease, exact current pair and
held result, baseline reference, original element comparison handle, selected-side
question/source basis/read clock/count, connected visibility and all pending/depth
conditions. It derives from its own admitted pair, not event-supplied query data.
The element stores an immutable clone of canonical input, so the host correctly
retains that initial validated element handle instead of comparing it by reference
to the separately admitted original. Rebinding never blesses arbitrary replacement
element data.

Summary switching, child opening and owner suspension revoke the old binding.
Back installs fresh controls over the held pair, preserving passive Find/page
state. A side/key focus descriptor targets the new current control; the detached
old opener is not reactivated merely to restore keyboard focus. Generic H1
baseline/summary restrictions remain unchanged. No remaining source-level
architecture blocker found; browser behavior, generated parity and actual-frame
results remain delivery-owned.

## Property-test identity-frame finding

The initial cross-language test passed a raw Go canonical intent to the browser's
conservatively admitted comparison. `importSnapshot` deliberately replaces raw
Dataset IDs and revisions with process-local `goo.local/...` identities, while
retaining scope ID and scope generation. Rejecting the raw intent in that imported
frame is correct and must not be fixed by relaxing context checks.

Verification should compare Go/browser planning directly in the same raw frame,
then generate equivalent intents in the actual admitted browser frame and assert
the same semantic question. Equivalent corruptions must still reject, preserving
the original-side tuple before flipping its side. A raw-frame intent must explicitly
fail against the admitted frame. These contexts identify the supplied snapshot
instance, not a portable authorization token; no planner HTTP bridge or durable
intent serialization is introduced by292. No production correction was needed.
