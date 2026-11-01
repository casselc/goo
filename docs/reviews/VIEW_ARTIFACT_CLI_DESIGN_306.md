# View artifact CLI usability and whole-flow review — 306

Status: command source and all three actual native workflow frames accepted.
Read with [portable view design](PORTABLE_VIEWS_DESIGN_305.md) and
[the current portable artifact contract](../PORTABLE_VIEWS.md).

## User outcome

A service team validates its checked-in Goo view in CI without a Goo instance,
then explicitly imports that same file into a chosen installation's existing
catalog. This completes repository → CI → catalog. Neither command is a registry,
deployment controller or telemetry query.

## Two plain commands

`goo validate-view -file service-view.yaml` validates offline using the same
artifact decoder and existing definition admission as browser import. It requires
no network, server, credentials or store. Successful stdout is a stable JSON
receipt containing the checked artifact identity and admitted view; failures
produce a concise actionable stderr error and nonzero exit. Do not imply that
validation proves a service has matching telemetry or that an installation can
execute every configured query successfully.

`goo import-view -file service-view.yaml -url https://chosen-goo.example` first
validates locally, then explicitly publishes the admitted view through the
existing saved-view API. Both file and destination are required. A missing URL
must not silently target localhost or a remembered installation. Unsupported
URL forms and invalid files fail before network publication. The command does
not Open, Apply, Run, capture evidence or initiate a subscription.

Successful stdout is a stable JSON receipt with artifact namespace/name/revision
and the ordinary saved-view ID. These transport labels do not become registry
keys, unique releases or updates. Existing content identity governs repeated
imports; the command must not invent overwrite/update semantics. Keep diagnostics
on stderr so CI can consume stdout without trimming human progress messages.

Errors should identify the failed step and relevant file/destination without
echoing the full artifact or secrets. A failure after request submission can be
an uncertain publication outcome: do not claim no saved copy exists. Explain
that retrying the same configuration uses the existing content identity, rather
than adding another recovery protocol or a force flag. Keep network time bounded
and avoid redirecting a publication to another destination.

## The whole contribution flow

Use one real authored YAML specimen checked into a service repository. Validate
it from an environment with no Goo server. Verify invalid syntax/reference input
fails before any network. Import it explicitly into a clean running installation,
then find its ordinary entry in Saved exploration views.

Open its original fixed question to see that configuration stages without reading.
Separately Apply its presentation to a different compatible current scope; the
destination's environment, predicates, bounds and held observations remain its
own. Run explicitly where required. These existing UI actions complete the CLI
workflow; no separate imported-package page or command-specific Open button is
needed. Browser/model verification must distinguish publication from acquisition
and terminal completion from an unfinished prefix.

## Next substantive candidate — 307 cohort comparison continuity

After the code-managed contribution loop, return to the primary observability
question: which population differs, and how? Current source exposes a concrete
cohesion gap in `ui/population-workspace.mjs`: the selection tray's Compare and
Compare with rest require non-authored views. The new saved/composed breakdowns
use `renderAuthored`, whose interaction currently stages an included facet. Thus
users who adopt the canonical authoring path cannot necessarily perform the same
selected-population versus parent/rest comparison available in the older native
breakdown flow. The 304 end-to-end task uses captured baseline comparison, which
answers a different question.

Prefer one bounded **selected cohort versus rest of its fixed parent** continuation
for an eligible authored count breakdown, reusing existing checked partitioning,
query owner and comparison displays. Preserve the original predicates, checked
environment and fixed bounds; retain missing/empty distinctions and independent
read semantics. A supplied record page must never define the rest population.
The action should live with existing group-selection/refinement actions rather
than become another top-level comparison tool.

Before selecting this cut, confirm with architecture which existing captured or
query-baseline composition can express it without a duplicate executor. If the
existing operation is already reachable through another ordinary authored path,
reconcile that discoverability instead of implementing it twice. Success is an
investigator selecting a visibly anomalous group in their saved/team view and
comparing it to the complementary scoped population, then returning to/reusing
the same view. This advances core investigation rather than adding another CLI
flag or specialized chart profile.


## Implementation and actual workflow review

The final implementation matches the command contract: required `-file` and
explicit import `-url`, regular bounded file admission before networking,
`goo.view-validation/1` stdout with admitted view, and `goo.view-import/1` stdout
with the existing saved ID. Import performs one bounded, nonredirecting publication
and validates the returned definition/content ID. Failures after submission state
that the destination may have saved the view and that no retry was attempted.
A receipt-output failure after successful save identifies the saved ID. No source
UX blocker remains; no additional CLI flags are needed for this cut.

The native verification owner reports the real CLI/browser journey green in
51.0 seconds. The actual [catalog frame](../journeys/screenshots/306-imported-catalog.png)
shows one ordinary template entry with the existing Open original question and
Inspect template actions; no command-specific registry or navigation appears.
The actual [executed imported view](../journeys/screenshots/306-imported-view.png)
shows seven spans split five gateway/two worker, alongside a Table with UTC,
duration and qualified `attribute:http.route`. The actual
[destination reuse frame](../journeys/screenshots/306-reused-view.png) shows that
scope's three worker spans and matching Table/route values rather than reusing
the source count. Source and destination full/page counts remain explicit.

All three actual frames are accepted with no visual blocker. Network/write absence,
offline validation and terminal completion are established by the native ledger
and verification report, not inferred from screenshots. See the
[306 executable storyboard](../journeys/306-view-cli.md) for the complete task.
The next 307 cohort-continuity recommendation above remains the substantive
product direction after this completed repository/CI/catalog loop.
