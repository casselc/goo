# View artifact CLI architecture review — 306

Status: accepted implementation contract; source review pending. This completes
the repository → CI → explicit catalog workflow recommended by 305. Existing
artifact codecs, view validation, store identity and catalog remain authoritative.

## Commands and effects

`validate-view -file` reads at most the artifact budget plus one byte and uses
`workbench.DecodeViewArtifact`. It performs no network call, publication, source
query or file mutation. Invalid input exits nonzero. Its checked JSON result may
include versioned validation receipt, artifact labels and canonical view, making
the admitted fixed question inspectable in CI without inventing another AST.

`import-view -file -url` requires both file and destination. It validates the file
locally before any request, then makes exactly one POST containing `artifact.view`
to the existing saved-view route. Success returns a versioned receipt containing
descriptive labels and ordinary saved ID. Import does not Open, Apply, Run, fetch
telemetry, create a package registry or update a definition by artifact label.

## Explicit destination and bounded transport

Accept an absolute HTTP(S) UI-root/base URL with optional deployment prefix, and
append `/api/exploration-views`. For example `https://host/goo` targets
`https://host/goo/api/exploration-views`. Do not guess whether a supplied `/api`
segment is already an API root. Reject userinfo, query and fragment; require a
host and valid parse. The caller chooses destination explicitly; no default local
or remote installation is inferred.

Use a small command-local HTTP client with a finite deadline (15 seconds), normal
environment proxy behavior and bounded response body. Refuse redirects rather
than changing the destination or replaying a POST. No retry is performed. A timeout,
connection failure or invalid response after POST may follow successful persistence;
report failure and its possible saved effect honestly, without automatically
issuing a second request. Do not add a generic transport framework or new store
filesystem writer for this command.

## Receipt and identity

The current route returns 201 with the saved `ExplorationView`. Require that status,
one strict JSON object and a validated ID-bearing view. Verify full submitted
content correspondence after detaching the returned ID and verify the saved ID
against the existing content-digest algorithm. A syntactically valid unrelated
64-character ID is insufficient. Preserve artifact transport labels only in the
CLI receipt; the existing store ID continues to identify the saved configuration.

## Acceptance and review

Run the real CLI against the checked repository fixture. Prove offline validation
without networking, invalid file refusal before requests, bounded reads, prefix
URL semantics, exactly one existing save request and truthful import receipt.
Exercise redirects, malformed/trailing/oversized response, unrelated valid-looking
ID/content, error status and timeout. Assert these failures do not retry or execute
source reads. The existing catalog Open/Apply journey remains the post-import UI;
there is no new browser entry or automatic query execution.

## Initial source review

The command implementation reuses the bounded artifact decoder before constructing
an HTTP request. It uses a command-local client with environment-proxy defaults,
15-second timeout, refused redirects and one explicit save POST. The bounded strict
201 JSON response is checked against the submitted ID-free view and existing
SHA-256 store identity. No source execution or automatic retry is present.

One URL-prefix finding was sent for correction: appending to decoded `URL.Path`
then clearing `RawPath` changes an escaped separator in a supplied prefix into an
actual path separator. Preserve escaped deployment prefixes when appending or
refuse ambiguous encoded separators explicitly; do not silently choose a different
destination prefix. A concrete destination test should cover that distinction.

## Final source review

The URL finding is resolved by appending to the escaped prefix, deriving decoded
Path from that final escaped value and retaining matching RawPath. Tests include
both interior and trailing encoded separators, so URL serialization cannot silently
select a different deployment path. The POST body is explicitly non-replayable and
redirects remain refused.

Catalog identity is now owned by the small `workbench.ViewContentID` helper shared
by Create, Load and CLI receipt verification. It requires ID-free input and leaves
admission to existing validation, avoiding a second hash definition in transport.
Final architecture source review has no remaining blockers. `git diff --check`
passed; executed CLI/package/native outcomes remain in the validation record.
