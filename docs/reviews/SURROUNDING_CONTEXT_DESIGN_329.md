# Qualified surrounding context — design review 329

Status: final source and all eight maintained actual frames accepted.

## User task and shared scope

A responder asks for logs near an observed event. Every surrounding-log entry
point must use the same checked service identity. The existing
`Discovery.surroundingQuestion` owns admission and query construction; the kept
adapter supplies its original member question and delegates without an overlay.
This removes a service-name-only interpretation from the older record inspector.

Require a nonempty reported string service and a known plain resource map.
Resource `service.namespace` has three admitted meanings: an absent own key,
a present empty string, or a literal string value. Missing telemetry is distinct
from unknown reporting: absent/invalid resource maps and nonstring namespace
values refuse before a source read. Log attributes never supply resource identity.
There is no all-namespaces fallback or compatibility switch.

The scope line reuses `ServiceIdentity.label`: **exact service NAME**, followed by
**namespace missing**, **namespace present empty**, or **namespace "literal"**.
The qualified resource predicate, rather than a decorative label, defines the read.
Unknown reporting has an explicit unavailable reason and no successful scope label.
A failed new anchor must clear any previous successful scope so a stale service
cannot appear to describe the refused question.

## Bounded context and acquisition owners

Keep the default ±30 seconds and accepted integer radius range 1–300 seconds.
Clip to the original half-open observation window, include the anchor's whole
millisecond, preserve offset/page behavior and retain the checked typed environment.
Neither the present clock nor custody time supplies replacement bounds.
Original message, severity, trace, package and other record predicates are omitted;
the scope explains this expansion. Time proximity is context, not causality,
proof of one request, or a complete sequence. A qualified service may have several
instances, and a fresh source can no longer contain the original observation.

The older record inspector's **Surrounding logs** click is its explicit immediate
acquisition. Existing radius, Refresh and Next controls retain their ownership.
The kept action stages its records view without telemetry/custody/catalog reads;
**Run** remains its separate explicit acquisition. Existing recipe/retention
configuration GETs remain separately counted metadata, not source reads.
These paths share a pure planner, not a navigation or execution owner.

Keep one compact scope/status surface and existing native disclosures. Reuse the
records display, existing actions and shared label helper; no wizard, new radius
controls, repeated explanation cards, new endpoint or preservation format.
The trace-ID question intentionally expands across services and remains unchanged.
Its validity and readiness stay independent from surrounding-log admission.

## Return and evidence

The old inspector retains Close/Escape focus restoration, request abortion and
late-result guards. Refusal clears rows and exact-scope text without a source read.
Kept continuation retains its original inspector, opener focus, observed scroll,
immutable receipt and original member window. Its Origin scope continues to show
that original window separately from the clipped child question.

Completed kept Back reads nothing. Pending Back may issue the existing exact
query-cancel control request, then refuses late adoption. This does not prove
physical cancellation after a source read has completed. Fresh result clocks,
coverage uncertainty and ordinary pagination remain separate from kept custody
and the original observed basis. Save/share preserves the question, not the
in-tab return lease.

## Final source and actual evidence

Source review accepts the three frozen UI assets: one canonical qualified
predicate, kept delegation, shared labels and cleared stale scope. Verification
reports named/empty/missing membership, radius/clipping, page offset, Close focus
and cancellation passing. Full source checks and release build preceded the sole
browser run; no product changes followed. The combined run passed both journeys
(329 12.5 seconds; refreshed 328 with actor 48.4 seconds).

Viewed all eight final frames. `329-qualified-surrounding.png` and refreshed
`30b-independent-surrounding-logs.png` show the same real legacy panel: exact
service and quoted namespace "team329", typed current/legacy environment,
exclusive UTC window and omitted-filter/proximity limit. Existing radius/Refresh/
Next controls remain compact; 103 captured / 100 on this page and enabled Next
make the page limit visible. The INFO anchor and WARN neighbor are qualified
identically. This bounded panel crop does not show all 100 rows or prove exclusion.

The three refreshed 328 frames show independent trace refusal, enabled surrounding
action, shared note/disclosure, frozen original window, explicit staged Run and a
two-row named-namespace INFO/WARN table with fresh clock/unknown coverage.
The three refreshed 327 frames retain both actions for a present-empty namespace,
explicit original-window staging and a two-service Waterfall. Exact predicates,
receipt equality, Back, focus and pending cancellation are native assertions,
not inferences from these crops. No visual or semantic blocker remains; no new
controls, cards or broader feature are requested by this review.
