# Investigation preservation ownership review — 426

Status: source changes, focused owner tests and the final bounded UX/accessibility
frame follow-up are accepted. Actual browser, model and whole-product execution
results are recorded separately in [validation](../../VALIDATION.md).

The reviewer inspected current preservation, saved-definition reopening and the
425 retained counted-value ownership boundary. This review found a concrete gap:
old preservation controls checked the surrounding disclosure's rectangle, so a
closed disclosure with a visible summary could still permit a cached callback to
save, stage a question, open a capture form or navigate to a saved view.

## Consolidated control boundary

`ui/population-workspace.mjs` now uses the existing native-control readiness
helpers at the actual input/button/link, while retaining render-owner, query,
draft-revision, saved-ID and serial checks. Query generation remains a strict
pending-response guard; it is not substituted for retained DOM ownership. No new query, store, permission
or persisted format was introduced.

- Save, title editing, question sharing/staging, capture handoff, saved-view
  attachment, portable import/export and saved-library actions refuse closed,
  hidden, inert, disabled or otherwise withdrawn controls.
- A per-render preservation owner survives an unchanged held child and Return;
  replacing or invalidating the render revokes it. Saved-definition links retain
  normal native navigation when current. Refused
  link activation prevents the browser's default hash navigation, so a stale or
  closed anchor cannot start a store read or replace the active investigation.
- Pending save and portable-artifact responses use completion readiness. Their
  own deliberately disabled opener does not reject a valid completion; disabled
  ancestors, withdrawal and changed source/draft identity do reject publication.
- Request cleanup uses matching closure-local serial/input revision separately
  from source generation, current native visibility and publication permission. Withdrawing an export/preview/import drops its output/admission or
  receipt and releases its own pending state. Returning can explicitly retry;
  controls cannot remain permanently disabled merely because a reply was dropped
  or an unchanged held parent advanced its child-query generation.

The selected/rest comparison's ordinary Question and Evidence actions now name
the selected population explicitly. Save exploration view still preserves both
fixed queries and presentation. Capture selected query opens the existing form
for a new independent read of that query; it does not claim to save either held
operand or the complete comparison. Existing pair-capture facilities elsewhere
remain a separate adoption opportunity, not a missing global storage primitive.

## Metadata-only entry before a source read

Native browser verification exposed a pre-existing entry gap: the saved-view
library lived below a result-context ancestor hidden before any result existed.
Geometry-based callbacks had obscured that gap. The root exploration owner now
supplies its existing query-options row as an optional library host. The same
native saved-view/import details node lives there while no population is open,
moves into existing View settings for an active population, and returns home on
invalidation. Disposal removes it even from the external host. Child workspaces
keep their default placement. There is no duplicate form, new API or automatic
source read; this does not implement the later Recent/Pinned recall milestone.

## Verification and limits

`tests/preservation-owner-426.test.mjs` executes the actual production handler
sections with the real preservation, question, composition, artifact and native
readiness helpers. Its small DOM intentionally reports rectangles for content of
closed disclosures to demonstrate why geometry alone is insufficient.

Five focused checks pass: withdrawn actions cause no writes, staging, capture or
title changes; a valid save retains both comparison queries; selected question
and capture handoffs retain only their stated query; native saved links prevent
withdrawn/stale navigation; pending publication and portable cleanup/retry obey
separate custody and permission checks. The tests include a valid own-disabled
pending save and closed, inert, changed-revision, changed-generation, replaced-render-owner and disabled
ancestor refusals. A positive child-generation advance confirms that unchanged
held Save and saved-link controls remain usable after Return. Portable cleanup
and retry also pass after both inert withdrawal and a child-generation advance.

These are source/owner checks, not an independent browser or assistive-technology
certification. The executable 426 journey and portable-view regression provide
actual browser coverage; root records their outcomes separately.

## Roadmap implications

Recent/Pinned recall should retain an existing exact saved-view reference rather
than flatten a fixed comparison to its selected query. Further reuse should adopt
existing comparison/capture/template primitives before adding another format.
Normalization, retained custody and active permission have different purposes;
consolidation must preserve those distinctions rather than introduce a universal
identity that erases authored clause paths or evidence clocks.


## Independent usability/accessibility follow-up

Reviewed current 426 preservation, reopened-definition, applied-comparison and
returned-Activity frames at 1440/390/320 (plus returned 390), and the composing
preservation source. The reviewer did not author the production controls or run
the browser/model suites. This is bounded source/frame review, not a screen-reader
certification.

Accepted: **Share selected question**, **Add selected question to investigation**
and **Capture selected query** identify the selected-only scope directly. The
existing Question / Evidence / View definition groups remain useful; no extra
wizard or preservation format was needed. The capture note states a new independent
read; the view note states both fixed questions and no saved observations. Saved
configuration links, titles and native actions wrap at narrow widths without a
new clipped control. The applied 259-rest / 2-selected comparison retains shared
scale, exact change and scope/read disclosures. Returned Activity visibly retains
search and two hidden selections; root's actual journey establishes focus/source
behavior rather than this image review.

The initial staged frame still displayed the generic **Independent reads.**
status despite no operands loaded. Agreed refinement: replace that line when
opening a saved definition with **Configuration only · Run reads both fixed
questions.** (generic telemetry wording for other views). This makes state clear
beside Run and replaces generic prose rather than adding another warning. The
final rendered status-line follow-up is accepted below.

This journey verifies configuration preservation. It makes no capture POST and
claims no immutable pair receipt. Existing comparison-capture infrastructure is
an adoption opportunity for the next depth item; it is not globally absent.

### Initial library host lifecycle check

The existing saved-view route-owner fixture also exercises the real empty
population `stage` path with an external library host. All 11 focused route checks
pass. The new case verifies the same catalog node starts in the visible metadata
home, moves into existing View settings on local stage, returns home on
invalidation and is removed on disposal. Every local move performs zero fetches.
The actual empty-workspace portable-import browser flow remains separately
validated by root.


### Final configuration-state and metadata-entry follow-up

Reviewed the final `investigation-reuse-staged-{1440,390,320}.png` frames after
the status refinement. Accepted: **Configuration only · Run reads both fixed
questions.** is now visible beside Run; the configured selector, fixed UTC bounds,
native Run focus and Return to paused activity remain legible. No observations
or successful source read are implied by the reopened definition. No outstanding
finding remains from this bounded source/frame UX/accessibility review.

Also reviewed `305-import-preview-{390,320}.png`. The same **Saved exploration
views** disclosure appears in the existing advanced query options before any
population read, with native Preview YAML and Import view controls and a concise
configuration-only preview meaning. The metadata summary and exact scope wrap;
the YAML textarea remains a native editing surface. No new whole-page overflow or
clipped action was established. This reuses the existing library rather than
adding another import wizard or the later Recent/Pinned saved-reference feature.

Root separately reports the empty-store portable-import browser/native-model
regression and the complete 426 journey/45-second actor passing. Those execution
outcomes establish no-read budgets and ownership behavior; this reviewer did not
rerun them and does not infer them from the captures.
