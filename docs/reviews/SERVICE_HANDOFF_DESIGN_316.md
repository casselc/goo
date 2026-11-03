# Activity → service context handoff — design contract 316

Status: bounded implementation source and all three actual native frames accepted.
The environment-policy boundary below remains explicit; unsupported policies are refused.
Read with [post-source priorities](POST_SOURCE_PRODUCT_314.md).

## User outcome and existing entry

A responder sees an interesting service in Activity and opens that service's
observed context without manually re-entering its identity/environment and
accidentally changing the time window. The Activity remains paused and retained
for return. This is a context handoff, not proof that a log predicate and HTTP
server-span population are the same query.

Add **Open service** to the existing row service menu beside **Only this service**.
Rename that menu's accessible summary from Filter service to **Service actions**,
because navigation is no longer a filtering action. Keep the filter gesture and
its existing staged behavior. Do not add a global toolbar, new wizard or gallery.

## Exact carried context, explicit limits

The admitted handoff carries the observed nonempty reported service name, known
namespace identity, supported checked environment policy and that observation's
fixed half-open read window. Use `item.readWindow`, not the latest global tail
window or a fresh discovery lookback. Show full UTC bounds in the existing scope
presentation and preserve the original Activity return identity separately.

Namespace missing, present-empty and named are distinct. A present resource
attribute map without a namespace key can establish known missing; an omitted or
unavailable map cannot. Do not treat unknown namespace as all namespaces or
missing. Non-string namespace values and absent/empty/ambiguous service names
must not become convenient wildcards.

Only enable policies the actual service owner/API can honor. Current service
inventory/discovery/dashboard boundaries use named string environments; an
accepted record selector does not automatically make every service workflow
selector-aware. A first supported named current-only context can be useful, but
effective, empty-label, unreported or other unsupported policies must refuse
rather than be weakened into a name string. Any equivalent representation
projection requires explicit checked architecture admission, not a UI guess.

## Staged service surface and one acquisition action

Opening stages the service context in the existing Services/resource owner,
without discovery/overview requests. Show a compact **From paused Activity**
context and **Read service overview** as the single deliberate acquisition action.
That button may execute the existing bounded service plan/reads; it does not make
them atomic or reuse Activity rows as overview data. Existing source/quality
facts remain visible after acquisition. Do not merge Activity and discovery query
owners or invent another service query executor.

Use a concise visible fact: **Service context only; Activity filters are not
carried.** The scope identifies what does carry: service, known namespace,
checked environment and fixed window. A details location can show omitted
Activity clauses if useful. A log-body filter, trace predicate or other
signal-specific condition must not silently enter or disappear from a destination
labelled Exact Activity scope. Existing service defaults/profile selection remain
their own checked process after the explicit read.

Once read, ordinary default/Customize/contributed-view operations continue through
the existing service workflow. No new prebuilt chooser or extra capture step is
needed. Changing staged context or leaving invalidates pending reads through the
existing generation/cancellation owner.

## Unsupported cases and manual discovery

For a known unsupported identity/policy, expose an actionable reason rather than
a working-looking action that drops context. Prefer a disabled Open service with
adjacent/disclosed text accessible to keyboard/touch, or a clear refusal status
before navigation. A disabled button title alone is insufficient. Examples should
state the actual requirement, such as **Service overview cannot use this
environment policy**, not Missing service when the name is present.

The existing manual Services & resources entry remains available. It starts a
separate user-chosen discovery scope and must not be described as an automatic
fallback preserving the Activity query. Do not silently navigate there with a
blank environment, all namespaces or new window after a refused handoff.

## Named return and lifecycle

Use the existing named paused-Activity/resource return boundary. **Back to paused
Activity** restores the same held buffer, filters, selection, scroll and visible
service-menu opener without a read, resume, new history visit or durable write.
Do not focus a detached/hidden row action; use its current retained equivalent or
the established Activity fallback. Service-source reads retire on leave through
the existing owner; return does not restart them.

The action belongs to the current observed occurrence, selector and read-window
lease. A stale row, hidden menu, changed signal/environment or retired visit must
not open a service under another observation's context. No selected-record Keep
intent or capture provenance is transferred. Resource return presents ancestry;
it is not a combined query/session store.

## Gate and acceptance

Before product code, settle: eligible environment representations, known namespace
availability, the source-free stage seam in Services, explicit read ownership and
the exact Activity return/focus lease. If current Services can only auto-read or
weaken selector semantics, keep this contract proposed until that bounded seam
exists; do not ship a misleading button to satisfy the navigation request.

Then verify a real incident row → service menu → staged scope → explicit overview
read → Customize/view → return. Stage performs no request; read uses the exact
admitted identity/environment/window. Compare current-only named, namespace
missing/empty/named, omitted resource identity and unsupported environment cases.
Show carried versus omitted filters, with no implicit broadening/fresh time.
Return preserves held Activity and focus without IO. Update real compact frames
and executable storyboard; static screenshots do not establish lifecycle/effect
correctness. No production-store writes are needed merely to review this handoff.


## Gate resolution and final source acceptance

Architecture admitted one bounded service encoding: a named current-only record
environment is projected into the existing named Services scope only after
checked environment/SQL semantic equivalence. Effective, empty-label and
unreported policies remain refused with a reason/manual path; there is no broad
selector migration or silent fallback. Known resource namespace and per-occurrence
completed read window remain part of the checked handoff. Unknown namespace does
not become missing, and conflicting log attributes do not replace resource identity.

The source adds Open service inside Service actions, stages through the existing
service owner with one explicit Read service overview, and exposes the omitted
Activity predicate under **Activity filters are not carried** with origin signal.
The reason for unsupported cases stays in that existing menu rather than adding
prose to every row. The shared action-visibility correction makes a closed menu's
summary usable while preventing hidden descendant commands. Existing named
paused-Activity return and read cancellation retain their owners. No source UX
blocker remains.

## Actual 316 frame acceptance

All three final actual frames are accepted. The
[Activity action](../journeys/screenshots/316-activity-service-action.png) shows
Open service beside the existing service/namespace filters in the incident row's
menu; the paused buffer and log-message filter remain visibly intact. The
[staged service context](../journeys/screenshots/316-staged-service-context.png)
shows the resource namespace/service, named current-key-only environment and the
original fixed interval **16:35:22.932–16:36:22.932 UTC**, with a clear Back to
paused Activity, omission disclosure and single Read service overview action.
It does not claim the log-message predicate is applied to HTTP spans.

The [narrow acquired overview](../journeys/screenshots/316-read-service-overview.png)
retains that checked context/omission fact and shows 107 observed HTTP server-span
records after explicit acquisition. Incoming HTTP versus All spans remains the
existing service profile choice; recorded counts and unknown coverage/health are
not converted into verified requests or success. No new wizard or catalog appears.

The native verification owner reports green in **51.3 seconds** (**57.5** including
runner), proving source-free staging, the earlier occurrence's window rather than
the latest window, seven explicit overview source requests and Back before/after
read or cancellation restoring the paused 13-row buffer, exact 69/689 scroll
positions and row-summary focus without a tail restart. Unsupported policy and
identity counterexamples are unit/Hegel coverage, not a claimed native policy
matrix. Optional campaign completion is reported separately; it is not assumed
from the screenshots or this review.

No visual blocker remains for this finite round. No additional product features,
source edits or durable writes were requested by the final design review.
