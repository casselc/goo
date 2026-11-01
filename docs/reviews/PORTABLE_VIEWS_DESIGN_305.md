# Portable service/team view artifacts — design review 305

Status: implemented source and three actual browser frames accepted.
Read with [304 whole-flow review](AUTHORING_DESIGN_304.md) and the contribution
intent in [the grammar proposal](../VIEW_GRAMMAR_PROPOSAL_260.md).

## User outcome

A team turns a useful authored view into a reviewable YAML file beside its service
code, then loads that definition in another Goo installation and uses it on that
installation's telemetry. The file contains configuration, not evidence or source
access. This completes contribution of existing views; it does not introduce a
second view language, marketplace or editor.

The current Save exploration view flow already admits checked fixed questions and
presentation. The existing catalog already distinguishes Open original question
from Inspect display → Apply composition. Reuse both. The portable artifact wraps
the supported existing definition with explicit namespace, name and version,
without inventing another composition schema.

## Export in the existing authoring workflow

Offer **Export YAML** within the existing Preserve/View definition area for a
supported checked definition. Keep identity controls together: Namespace, Name,
Revision. Title remains the human-readable view title; namespace/name/revision is
its code-managed identity. Example values should look like `payments`,
`database-errors`, `1.0.0`, subject to the backend's checked identity rules.

Export the admitted current configuration, not a mixture of unapplied source,
filter or breakdown drafts. Finish or discard those drafts first using existing
rules. If export requires a saved definition, state Save first and use that one
immutable definition; do not silently save or publish as part of download.

The resulting artifact can include its original fixed query as a reproducible
configuration specimen. A concise line must say that opening that question keeps
its environment/window while applying its presentation keeps the destination's
scope. Neither action acquires telemetry until the existing explicit Run.

A downloaded filename may use namespace/name/revision, but the artifact's declared
identity is authoritative. The portable envelope contains no store document ID,
rows, result token, receipts, source basis or completeness witness. Never claim
that exporting a view exports the telemetry or its capture.

## Import within Saved exploration views

Add a single **Import view** disclosure to the existing Saved exploration views
library. Use a bounded YAML input/file mechanism and the existing action/status
conventions. Do not add a top-level navigation destination or a parallel source
editor. The library remains usable before any population is open.

The user explicitly submits **Import view**. Parsing and semantic admission must
complete before any definition is stored. Failure stays in this disclosure with
an actionable error and preserves their input; it must not replace the current
view. Success reports imported identity/title and makes the ordinary catalog
entry available. Import is a configuration write, not Run, Apply or Capture.
Do not navigate or start telemetry automatically.

Show namespace/name/revision with the import success receipt and ordinary saved ID.
These are descriptive transport metadata, not a registry address or uniqueness
claim. The imported result is an ordinary saved copy; existing content-addressed
identity and deduplication apply. Same label/version with different content must
not overwrite or update a previous saved view. There is no need to expand the
persistent view schema to retain a package label in this first cut.

Export asks for the envelope identity explicitly each time, including re-export
of an imported or subsequently edited definition. It must not silently claim an
unchanged upstream revision. The label is not a source execution or lookup
capability.

## Two existing ways to use the result

**Open original question** stages the exported fixed query and its original bounds
and environment. It is useful for revisiting that configured question, but may
show no data in another installation. The UI must not silently substitute local
service names, a recent window or another environment.

**Inspect display → Apply composition** checks compatibility with a current
population and applies presentation intent while retaining its query, environment,
fixed bounds, paging and held data. A bounded category template uses its existing
binding rules. Import alone does not establish compatibility or promise that a
raw view is a general parameterized template.

Keep these choices in the existing catalog, with its current availability/status
rules. An imported definition does not need a separate Open team view command.

## References and failure behavior

Store-local captured baseline IDs cannot identify equivalent evidence in another
installation. Reject nonportable comparison/tail/capture references with a clear
message identifying the unsupported reference. Do not drop the comparison,
choose another baseline, fetch remote evidence, or rewrite it as an ordinary
count view. A plain noncomparison authored view is a valid first portable cut.

Unknown kinds, unsupported schema versions, invalid qualified fields, ambiguous
YAML and over-budget inputs must fail before publication. No declaration can
supply arbitrary HTML, executable JavaScript, arbitrary endpoints or source
credentials. These are admission rules, not extra user-facing safety controls.

## Acceptance as a real contribution workflow

Author a useful effective-environment service view from Activity through the
connected 304 breakdown/records flow, with qualified fields and a meaningful title.
Save and export through native controls; inspect the YAML's identity and absence
of store-local custody. Import in a clean store through the same library, with no
telemetry reads or automatic navigation.

Open the imported original question and verify staged original scope. Separately
apply its composition to a different compatible service scope and verify the
destination query/count/records remain its own. Run explicitly where acquisition
is required, then save using the existing workflow. Include an edited-invalid
artifact and a captured-baseline artifact as rejected controls, proving no partial
publication or silent stripping. Keyboard progression must reach the status and
next catalog action without returning to a hidden or detached control.

Actual compact frames should show export identity/configuration, successful
import in the existing library, and the presentation on destination telemetry.
Update the executable storyboard and diagrams without claiming broader template
parameters, repository synchronization or new custom kinds.


## Implementation and actual frame review

The implementation stays in the existing Preserve and Saved exploration views
surfaces. Export accepts a valid current definition without silently saving it;
Namespace/Name/Revision and definition changes retire a prior generated download.
Import separates Preview YAML from the explicit configuration write, shows fixed
scope before publication and retires the preview after edits or visit changes.
Store-local capture references fail instead of being stripped. Import success
uses the ordinary saved catalog and existing content ID. Source review has no
remaining design blocker.

The actual [preview frame](../journeys/screenshots/305-import-preview.png) shows
`platform/task-mix @ v1`, the fixed-question title and checked environment/time/
service scope with separate Preview and Import controls. It states that import
neither opens, applies nor runs telemetry. The long scope remains detailed rather
than being confused with the destination's current question.

The actual [imported view](../journeys/screenshots/305-imported-view.png) shows
seven gateway records split five fast/two slow, alongside the retained List,
qualified record/resource zone values and UTC. The actual
[reuse frame](../journeys/screenshots/305-reused-view.png) shows the destination's
three worker records and its own count, with the same chart/List presentation.
These frames make the separation between imported configuration and destination
observations visible. All three are accepted with no UX blocker. Browser execution
results remain owned by the native verification report; static frames alone do
not establish absence of network or publication effects.

## Recommended next outcome — 306 repository and CI workflow

Complete the code-managed contribution task with the existing Go CLI, rather than
adding another kind or catalog. Provide offline `validate-view -file` over the
same artifact decoder/admission, with an admitted identity/title summary or
canonical artifact output and nonzero exit on failure. This command must work
without a Goo instance, credentials, telemetry or a publication write, making it
useful in a service repository's CI.

Provide a separate explicit `import-view -file -url` that validates locally and
publishes through the existing saved-view API. Require a deliberate destination;
return the artifact label and ordinary saved ID. Do not Open, Apply, Run, follow a
watcher or install a new registry as part of import. Errors must remain distinct
from successful publication, and an uncertain response must not be represented
as proof no copy was created; existing content identity supports safe retry.

Document one checked-in team view validated offline, imported to a clean test
installation and reused through the same existing UI. That is the missing
repo → CI → explicit installation loop. It advances the original grammar purpose
without promising automatic reconciliation, release resolution or new template
parameters.
