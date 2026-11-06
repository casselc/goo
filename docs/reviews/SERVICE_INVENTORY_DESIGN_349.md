# Selected-service operation inventory — design review 349

Status: implemented source and all three regenerated actual-frame reviews
accepted. Execution/custody results are recorded separately in validation/journey.

A user opens an observed service to see its operations. Unrelated services should
not exhaust that service's discovery budget. Keep the current service overview
and inventory table; change the scope of its source acquisition rather than add
another dashboard, cap increase or discovery engine.

## Source finding

`ui/service-browser.mjs` currently requests inventory with the environment,
namespace and fixed window, then filters rows by selected service in the browser.
Its capture copy accurately warns that the reread includes other service names.
`exploration/inventory.go` groups service/operation names before enforcing the
1,000-pair bound. Thus a small selected service can lose its inventory because
unrelated pairs exhaust that whole-scope bound.

The intended repair is exact selected-service filtering before grouping/admission,
with matching response proof and capture scope. Whole-environment inventory remains
available in its existing intentional path; this cut does not silently redefine it.

## Identity and visible meaning

Preserve the exact service name, namespace identity, typed environment policy and
original fixed UTC window. A selected empty service name is an explicit identity,
not an omitted all-services selector. Use **(empty service name)** consistently in
headings, table captions, accessible names and capture title. Do not infer a
service from display text or use truthiness to drop empty scope. Namespace missing,
empty and literal values remain distinct under existing admission rules.

The inventory heading should name the chosen service, with existing scope evidence
available. Counts mean captured span records for operations; recorded Error is not
request uniqueness, service health or application outcomes. Duration buckets keep
their current unit/range meaning and do not become fleet percentiles.

## Capture and continuation

Use the existing capture gesture with a concise service-specific label such as
**Capture service operation inventory**. Its copy says it independently rereads
operations for this service in the selected namespace/environment/fixed window.
It does not archive the currently displayed catalog, and it does not include other
service names. The capture preview and reopened receipt must show that exact same
service scope, including explicit empty identity.

Opening an operation preserves exact service, operation name, namespace, typed
policy and fixed bounds. Stage first; the existing explicit Run reads telemetry.
No baseline or inventory receipt is silently substituted for the child read.
Back returns to the held service overview with local Find and current usable
controls, focusing a connected visible semantic opener. No Back/disclosure/search
replacement reads are introduced.

## Honest states and recovery

A successful zero-operation read says no operations were observed in this scope;
upstream coverage remains unknown. A refused/unavailable source is not a zero-count
service. Keep successful independent service spans/metrics visible with their own
receipts when operation inventory fails. Reuse shared concise refusal/details and
existing explicit recovery boundaries where admitted; do not broaden this cut into
an overhaul of every service-dashboard source owner.

## Acceptance

Use a genuine ordinary OTel service in an environment containing more than 1,000
unrelated service/operation pairs. Its small exact inventory must still succeed;
whole-environment inventory may retain the existing honest refusal. Include named
and explicitly empty services, namespace distinctions and typed policy conflicts.
Prove source SQL admission scopes before grouping, response/capture scope equality,
full counts and no unrelated records in the selected service capture.

Follow one operation into supporting records, Run, Back and capture/reopen through
the real UI. Check original bounds, policies, exact names, independent read clocks,
Find/disclosure state and source budgets. Actual desktop/narrow frames establish
readable service identity, operation counts, capture meaning and return focus;
semantic/observed tests separately establish source membership and custody.

## Frozen source acceptance

The backend now distinguishes omitted all-services inventory from a present service
string, including exact empty string, and adds that predicate before grouping.
Rows are checked against the requested service. The UI builds and validates the
complete returned inventory question rather than filtering an environment result
into apparent service authority. Operation continuation refuses a row outside the
selected service.

Service overview captions and accessible action names use the explicit empty-name
label. Capture is now Capture service operation inventory, with truthful
selected-service-only independent-reread copy and a bounded valid service title.
Local Find reports no local matches separately from a successful no-observation
read, retaining source totals; current-generation/visible button guards prevent
departed controls from acting. Source review accepts the bounded two-file UI path.
Actual named/empty service, capture/reopen and return frames remain pending.

## First actual desktop frame

Reviewed `349-selected-service.png`. The selected service's three operation names
and seven captured spans are readable, including literal/Unicode names. Local Find
is visibly focused, recorded-error counts and duration-bucket disclosures remain
separate, and Capture service operation inventory is followed by truthful
service-only independent-reread copy. No metrics observed remains explicitly
coverage-unknown. Desktop task-view appearance passes. The selected policy/window
context is above this scrolled viewport, so the frame does not establish fresh-entry
scope placement or independently prove source membership; native checks do that.
Narrow and captured-receipt frame acceptance remain pending.

## First narrow/custody review

Reviewed `349-selected-service-390.png` and
`349-stored-service-inventory-390.png`. The stored capture title, seven captured
spans, checked environment policy and original fixed window are readable; custody
frame appearance passes. Tall stacked operation cards are expected, and fitting
all three above the fold is not an acceptance requirement.

The selected-service narrow frame exposes a real caption sizing issue: the mobile
inventory table becomes `display:block` but its caption retains native table-caption
layout, wrapping even a short selected-service identifier into roughly six-character
lines despite the wide card below. This is not explained by task scrolling. A
bounded owner-scoped mobile caption display/width fix, following existing service
observation/saved-record caption patterns, is recommended. Final narrow acceptance
awaits the corrected actual frame; no new control or layout redesign is needed.

The existing `goo-table[mobile-layout=stacked]` opt-in is the preferred repair:
its shared style already makes captions block-sized and retains the two-column
metadata layout. Inventory's existing first/duration/action cell spans still apply.
Source review approves setting that attribute on the service-inventory viewport
rather than duplicating mobile CSS. Actual regenerated narrow verification remains
necessary; this source judgment does not claim the frame is already corrected.

## Final actual-frame acceptance

Reviewed all three regenerated final frames after opting the inventory viewport
into the existing shared stacked-table mechanism. The desktop table remains clear
and unchanged in meaning. In `349-selected-service-390.png`, the full caption now
uses the card width and the selected literal service is readable on one line;
operation names, 3/2/2 span counts, recorded errors, duration disclosures and
Explore spans remain legible. The capture action is visible and its service-only
reread note wraps within the viewport. The caption shrink-wrap defect is resolved
without new CSS or a new mobile control surface.

The regenerated stored frame retains the service-specific title, seven captured
spans, typed environment policy and fixed window. Final visual acceptance passes.
These are task-scrolled inventory/custody views, not fresh-entry scope placement
claims. Exact service membership, empty-name semantics, unrelated-cardinality
isolation, source counts and reopen-without-telemetry are independently verified
by source/generated/native checks rather than inferred from the named-service
screenshots.
