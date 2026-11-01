# Portable view artifacts: architecture review — 305

Status: agreed bounded contract; implementation review pending. This advances the
260 service/team contribution direction through the current saved-view AST rather
than a registry, evaluator, renderer package system or new authoring owner.

## Transport contract

Use a YAML envelope with `version: goo.view-artifact/1`, bounded `namespace`, `name`
and `revision` labels, and `view` containing the existing `ExplorationView` shape
without a store ID. Reuse existing validation for flat or canonical compositions,
queries, kinds, fields and optional supported template bindings. Labels identify a
reviewable repository artifact; they do not establish trusted code, renderer
loading authority, registry lookup or immutable publication provenance.

Encoding and decoding are pure API transformations. The proposed `/view-artifacts/encode`
and `/view-artifacts/decode` return the validated artifact and canonical YAML. An
explicit subsequent import uses existing `/exploration-views` creation and its
ordinary catalog. No second persistence owner, package registry or source read is
needed. Imported views use normal content/store IDs; repeated label/revision does
not overwrite another saved view. Metadata may remain transport-only: exporting
asks for labels again, and import confirmation explains that it creates an ordinary
saved copy rather than installing a resolvable package.

## Strict YAML and typed AST

The project already depends on `yaml.v3`, with bounded node checks in measurement
and exploration package parsers. Existing view/query structs use JSON tags, not
uniform YAML tags: decoding YAML directly into them risks naming/type divergence.
Use a strict bounded YAML-node to JSON-compatible representation and the existing
strict JSON decoders, then normal view validation. Reject duplicate/non-string
mapping keys, aliases/merges, custom tags, multiple documents, unknown properties,
wrong scalar types and explicit size/depth/node budget violations. Preserve exact
large integer/string semantics; do not round-trip through untyped floating point.
Canonical YAML must decode to the same validated AST.

## Portable intent and local references

Questions and declared fixed comparison queries are configuration and may travel
when the existing validator admits them. Rows, datasets, source basis, result tokens,
store IDs, capture receipts and dynamic execution state must not. Reject an artifact
with a store-local capture reference, including a captured comparison's baseline or
a tail baseline. Do not silently remove a comparison, fabricate a portable baseline
or read a source capture while encoding. An explicit future destination-reference
binding may extend this contract; it is not required for this first outcome.

An exported fixed question remains that fixed question, including time bounds,
selector and environment policy. Importing it does not rebase to now or another
service. Opening it stages the original question; Run is explicit. Applying its
presentation uses the existing ordinary composition/template reuse path and retains
the current destination question under those rules. The UI must distinguish those
actions rather than call both Import or Apply.

## Product and owner boundaries

Place import/export in the existing saved-view/preserve workflow. Current draft
export must use the same `settings()` admission as Save; unapplied source/grouping
or pending record workflows must not export a different hidden configuration.
Decode/preview cannot change the current question or publish a saved object. Only
explicit import saves, and the existing catalog then owns Open/Apply.

Retire asynchronous previews on source edits, owner changes and replacement
requests. A delayed encode/decode result must not overwrite a newer draft, download
an artifact for a stale view or enable Import for different text. Use the current
owner/generation patterns rather than a new generalized capability framework.

## Acceptance

Author a real service view through the 304 flow, export and review its YAML, decode
without a read, import a saved copy and use existing catalog Apply for another
service. Demonstrate fixed-question Open separately, requiring Run. Include raw
Waterfall/List and an admitted aggregate template round trip, typed environment
policy and qualified fields, malformed YAML/property controls, store-local reference
refusal, and stale decode/import controls. Prove no source requests occur during
transport/preview/save and no evidence authority is copied. Report executed checks
separately from this source contract.

## Initial implementation review findings

The backend uses pure encode/decode routes, a YAML-node bridge with strict JSON
admission, and existing view validation. Capture-reference rejection and ordinary
saved-view persistence follow the contract. Numeric scalar overflow must reject
rather than silently enter a floating-point fallback before typed admission.

Frontend review identified three concrete lifecycle/representation gaps for
resolution: encode responses must preserve the complete admitted view (not just
labels and questions); closing a pending disclosure must retire/reset that request
so reopening permits a new attempt; and an already-created download must be
retired when its source draft changes. Existing generation/input/serial fencing
should cover these corrections without adding a new owner.

## Final source review

The backend now explicitly rejects YAML integer fallback outside the exact 64-bit
range. Its detached encoder, strict decode/admission, reference refusal and pure
HTTP/schema paths have no remaining architecture blocker.

The frontend now compares the full normalized validated view on encode responses,
allowing known omitted-default representation differences rather than permitting
presentation substitution. Mutation tests cover title/kind/options/columns/template
and scope changes. Export/import track pending contexts; disclosure changes, visit
retirement, rerender and definition edits retire stale work through the existing
serial/context hooks. Download URLs are revoked and hidden when their definition
changes. First-class transport remains separate from explicit ordinary Save.

Second source review accepts these corrections. No new query, persistence,
renderer or evidence owner was added. Architecture review has no remaining
blockers; actual native/Go/schema/property/model execution belongs in validation.

Follow-up source checks strengthened two edge contracts. YAML may tag a huge
integer literal directly as float before scalar decoding; the parser now rejects
that lexical integer overflow as well as integer-decoder fallback. Artifact
controls also account for ancestor disclosure geometry/toggle events, so closing
Preserve or Saved views retires work just like closing the inner transport panel.
Empty view IDs follow existing omission semantics; nonempty local IDs remain
refused.

A final interoperability investigation checked handwritten YAML without `limit`.
The current Go query validator rejects limit below one before compilation, so an
omitted limit is refused rather than emitted as invalid zero. An initial review
assumption about an older compiler default was incorrect. Keep explicit valid
limits in documentation/examples; if omission convenience is adopted, materialize
it only from absent YAML fields before typed validation, while explicit zero and
typed zero requests remain invalid. Do not loosen browser page admission or mutate
submitted views. Actual Go-to-browser fixtures remain the parity authority.

## Recommended next outcome: repository-to-catalog workflow (306)

The next bounded step should make the portable artifact usable in normal service
repository CI and an explicit deployment/release action. The existing Go CLI owns
offline grammar validation/compilation, but 305 exposes view codecs only as Go
helpers, pure HTTP transforms and manual UI import. That leaves code-managed users
writing glue around an otherwise complete checked artifact.

Add an offline `validate-view -file` command using the same `DecodeViewArtifact`
function, with a nonzero exit on invalid input and canonical output or a concise
validated identity result. Add a distinct explicit `import-view -file -url` command
that validates locally, then posts only `artifact.view` to the existing
`/api/exploration-views` route. Require a destination URL instead of choosing an
implicit server; return the ordinary saved ID and descriptive artifact label.
Do not open/run queries, preview source data, write directly into a server's store
directory, watch files or introduce a second importer/registry. Endpoint failure
must produce a failing command, not a successful validation message that obscures
an unsuccessful import.

Demonstrate a service repository artifact checked offline in CI, an explicit import
into the runnable lab, and the ordinary UI catalog Open/Apply distinction. Existing
fixed bounds remain fixed; this is checked configuration delivery, not live
dashboard deployment or automatic current-time rebasing. Tests should exercise the
actual CLI binary, zero network access during validation, exact single existing
save request during import, invalid input before request, and clear HTTP/receipt
failure handling. This closes a real code-managed workflow before new renderer,
parameter or provenance machinery is considered.
