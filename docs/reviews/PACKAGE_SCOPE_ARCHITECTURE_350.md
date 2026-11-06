# Typed scopes in manifest-free packages — 350

Status: implemented; final source review closed without blockers. The initial
contract below records the gap after the 347 roadmap and 349 inventory handoff;
implementation review and closure follow it. This is scope parity for the existing
telemetry-package compiler/editor/store. No new grammar engine, native metric model, migration
layer or compatibility alias is selected.

## Source gap and consumer inventory

`exploration/package.go` currently stores scalar PackageSource.Environment and
DraftTelemetryPackage rejects a query's EnvironmentSelector. Explore disables
Create package for those ordinary typed questions. Package summaries, both
libraries and the apply event use source.environment directly. Thus current
Activity/Services scope semantics cannot pass through a central Goo authoring
workflow without losing policy.

The relevant paths are DraftTelemetryPackage, Question, Validate/Revision,
ReadTelemetryPackage/DecodePackageJSON, PlanTelemetryPackage, PackageStore
Save/Load/List and the telemetry HTTP draft/plan/save routes; the browser editor,
preview, library and Explore apply owner consume their results. Preserve existing
numeric/histogram, duration/status, grouping/split and reported-edge profile
contracts. Reported-edge combinations remain governed by Query admission.

The optional annotation handoff generates a function marker, not a source binding.
The source types in the annotations/native measurement packages are separate
contracts despite similar names; do not generalize their exact stream identity or
current-key environment semantics as a side effect. Generated/workflow package
handoffs must pass through the same package plan/DTO checks as imported packages.

## Closed source union and parsing

PackageSource admits adapter/signal and exactly one wire environment arm:

- scalar `environment`: existing explicit current-key named semantics;
- `environmentSelector`: the shared version/policy/scope DTO, including current-only
  named or effective named/empty/unreported.

JSON/YAML presence matters. Reject explicit null arms, both arms (including an
explicit empty scalar beside a selector), missing arms, unknown keys and malformed
nested fields. Do not turn empty/unreported into a wildcard or project effective
named into scalar current-only. In programmatic Go values the unused string may
be empty; canonical serialization omits it beside the typed selector. Retain one
canonical valid representation per admitted definition and preserve the complete
policy in hashing and export.

Use shared checked selection semantics. A custom PackageSource JSON decoder must
not bypass outer DisallowUnknownFields; parse its exact keys explicitly and keep
the shared strict selector decoder. A custom YAML decoder likewise bypasses some
KnownFields behavior: explicitly validate source and nested selector/scope keys,
node kinds, nulls and duplicate keys rather than let node.Decode discard them.
Existing one-document, aliases/merges/depth, 32 KiB package and 64 KiB authoring
limits remain. JSON duplicate handling must match the existing parsed-map
last-member semantics, not struct-object merging; YAML duplicates refuse.

Reject malformed Unicode before a decoder/JSON clone repairs it into a different
selector/hash. Valid non-ASCII, surrogate pairs and literal U+FFFD remain valid.
Validate direct Go inputs before detachment; YAML and JSON source admission must
agree. The new strict decoder must not silently open a weaker arm on malformed
input. No legacy field aliases or silent format conversion are needed.

## Detachment, revision and clocks

Question must produce a detached full query, including selector, predicate,
grouping and every optional profile pointer. Plan must detach its definition and
question before returning canonical YAML/revision; caller mutation after planning
cannot make displayed definition/question disagree with the revision. Saving
serializes a validated detached definition, and Load verifies its content digest
and limits. Revising policy or selector changes the definition identity; changing
only fixed preview bounds does not.

A package is reusable configuration: selector/profile plus lookback, not an
observation. Draft starts from the applied full question, not unexecuted inputs or
loaded rows. Original from/to, page limit/offset and source basis are preview
context and do not enter reusable configuration. Validation/build/plan/save/open
perform no telemetry acquisition. Preview explicitly reads its fixed question and
gets independent evidence.

For saved relative reuse, an explicit **Stage in Explore** action requests a pure
plan with server-resolved now-minus-lookback bounds. That returned question is
then fixed: subsequent Run reads those exact bounds until an explicit fresh-window
change. An original draft or workflow with `previewWindow != null` uses **Stage fixed
question** and keeps those given preview bounds. Saved/imported configuration
with no previewWindow resolves relative bounds at the explicit Stage action; a
metadata preview must not silently advance its held window. All package handoffs stage; direct Explore-library Run labels are renamed
to Stage rather than retain an ambiguous event that sometimes reads. There is no
automatic telemetry read on opening/importing/reusing a definition.

## Shared browser admission and presentation

Use one bounded package source/plan helper across editor, preview, libraries,
workflow handoff and Explore adoption. It must validate the closed source union,
shared selector, question protocol, canonical definition/question scope/profile
agreement and expected revision where an immutable revision was selected. Do not
solve this only by removing the Create package disabled guard. A mismatched source
policy/profile or malformed plan must refuse before preview/read or stage.

Use shared RecordEnvironment labels/fields for summaries, search text and editor
scope; do not concatenate undefined source.environment. Typed policy remains
visible; typing a replacement scalar environment is an explicit scope change and
detaches the package binding according to existing semantics. A stage label is not
permission to invoke Run. Preserve all existing measure controls when applying a
package and keep package revision distinct from query/source basis.

Metadata/editor/library generations still own adoption. A delayed plan/import or
library-open response after Close, a new draft or workspace departure cannot stage
old configuration or initiate a read. Save remains explicit immutable publication;
its acknowledgment and display ownership do not authorize automatic replay.

## Optional Go annotation boundary (closed decision)

Normal planning with no annotation target succeeds for every supported typed
scope. Only the existing opt-in function template is restricted:

- checked current-only named selection supports it, whether scalar or typed;
- effective named/empty/unreported selection explicitly refuses a requested Go
  target because exact effective resource-key/absence policy is not a code-local
  function marker.

UI disables only Generate Go for that unsupported target with a concise reason.
Save, Preview and Stage remain available. Even where allowed, preserve the full
typed selector and retain the existing non-equivalence warning: the snippet does
not encode environment/namespace/predicate membership or establish equivalent
instrumentation. Do not coerce a typed current-only source into scalar merely to
satisfy template generation. No actual source file is patched.

## Acceptance

- JSON/YAML/direct-Go round trips for scalar, typed current-only named and effective
  named/empty/unreported; policy/current-legacy conflict witnesses; strict null,
  both-arm, unknown, duplicate, Unicode and size/depth/alias negatives. Hash changes
  for policy changes, not preview time; mutation tests cover every returned nested
  selector/predicate/profile and store/plan boundaries.
- Actual generated/browser DTOs through Go plan/compiler: whole predicate and
  optional numeric/histogram/duration/status/split/group profile survive. Corrupt
  source/question/revision echoes refuse before stage/read. Existing native and
  operation annotation contracts remain unchanged.
- Genuine ordinary OTel typed-scope investigation → Build/Validate → YAML/save/
  library reopen → Stage → explicit Run/Preview. Count source reads separately
  from metadata/planning/custody; Build/import/open/Stage cause zero telemetry reads.
  A saved relative Stage resolves a fresh window; held fixed workflow Stage does
  not advance it. Each successful read has its own basis/clock.
- Generate target behavior follows checked semantics, not wire-format presence;
  normal typed package use remains possible after unsupported target refusal.
- Current labels, exact scope and readable UTC/focus at desktop/narrow; source
  refusal/cancel/late metadata retain existing ownership. Reuse existing model
  effects where appropriate rather than add a model merely for another selector.

Close source and actual runtime acceptance before selecting the next product cut.

## Initial backend review findings (before final acceptance)

The candidate implements detached full questions/plans and hashes/publication from
one validated snapshot; typed annotation admission uses checked CurrentOnly
semantics. YAML source conversion through a checked map into the shared strict
JSON selector path preserves unknown/null/node-type rejection.

Two JSON-boundary issues were returned for correction before final acceptance:
PackageSource must decode its parsed last-member map rather than revisit every
original duplicate object, and full TelemetryPackage JSON admission must reject
malformed Unicode before predicate/title decoding repairs it. Guarding stored
DecodePackageJSON alone does not cover an HTTP `definition` pointer. Keep these
checks local to package DTOs instead of changing shared transport semantics.

### Corrected backend acceptance

Final backend source review accepts both corrections. PackageSource normalizes
its exact-key parsed map before strict wire decoding, including invalid-first /
valid-last and the reverse duplicate-selector cases. TelemetryPackage performs a
whole-definition Unicode precheck, and this HTTP module bounds a raw JSON value
and checks it before decoding its original strict request DTO. This covers draft,
nested definition, numeric plan and JSON-wrapped YAML requests without changing
shared transport behavior. Source YAML rejects binary/custom scalar tags before
map-to-JSON conversion can repair strings.

Question's manual clone preserves invalid caller strings for subsequent refusal
rather than repairing them, and detaches selector, predicate, grouping and optional
profile pointers. Planning and Store.Save hash/publish one validated frozen byte
sequence; returned definition/question/measure data remain independent. Annotation
admission follows checked current-only named semantics while ordinary effective
plans remain usable. No blocking backend source issue remains; whitespace checks
pass. This review ran no heavy tests and does not replace the coordinator's actual
Go/HTTP/semantic/native gates. Browser consumer/staging review remains separate.

## Initial browser review findings (before final acceptance)

The candidate centralizes source/question/revision/relative-width checks and
stages all handoffs, with fixed preview versus fresh relative planning explicitly
separated. Review returned two bounded admission corrections: validate consumed
view/handoff shapes before stage (otherwise a malformed views value can throw
only after openQuestion has already staged), and compare a returned immutable
selected definition with the expected original definition, not only its claimed
revision and reconstructed query. Lookback is configuration outside the query
projection: changing it together with returned bounds can satisfy a width check
against the changed definition. An expected-definition comparison needs no browser
hashing or second compiler; authored YAML edits intentionally have no old immutable
definition to match. Final browser acceptance follows those corrections.

### Browser correction review

The frozen helper now compares a known immutable definition with sorted canonical
JSON, separately from its opaque revision and projected query. Saved library and
relative reuse callers supply that original definition. Consumed view and handoff
shapes are checked before adoption. Authoring a newly edited definition does not
pretend to match an old immutable version.

Library Stage/Edit actions check the current selected revision, connected/visible
control and current workspace before metadata work; editor/library responses keep
those fences after awaits. All handoffs stage, and the shared Explore owner no
longer invokes Run implicitly. Preview uses the checked exact question; fixed
preview staging is source-free and relative staging obtains only a fresh pure
plan. Saved-record Create package now carries its immutable page/revision owner
instead of a typed-scope prohibition. Review requested one final teardown guard
so disposal, like suspension/page replacement, retires that callback before DOM
removal. No broad source or runtime changes are required.

### Final source closure

Backend and browser source review is closed with no remaining blocker. The final
SavedRecords disposal correction advances its revision before clearing DOM, so
saved package callbacks short-circuit after page replacement, suspension and
teardown. This does not add a source operation. Shared plan admission, all staged
handoffs, checked optional annotation capability and exact preview/profile echoes
remain as reviewed above. `git diff --check` passes. Actual Go/API/generated/UI
checks and native typed-scope/capture continuation results are owned by their
separate validation gates; this review ran no heavy or browser checks.
