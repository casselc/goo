# Metrics-only entry — 432a acceptance

432a delivers bounded identity bootstrap, not the remaining sparse/delayed-event
recovery in 432b. See the [contract](../METRIC_SCOPE_DISCOVERY.md),
[storyboard](../journeys/metric-scopes.md) and [validation](../../VALIDATION.md).

Independent architecture review accepted the fixed-adapter aggregation, current-key
presence precedence, namespace presence, complete group limit, exact UInt128
counts and explicit read/local staging boundary. Follow-up fixes reject mismatched
source diagnostics in the browser and clarify empty/omitted service names.

The first visual review rejected mobile entry: oversized controls, setup before
discovery, duplicate window framing, uppercase identifiers and unusable rows first.
The final independent review accepts regenerated 1440/390/320 frames and the
instrument, staged and evaluated views. Discovery precedes package setup; the
shared mobile control layout no longer stretches inputs vertically. Usable choices
appear first without mutating source order. Exclusions remain in native details.
Actual identity case and missing/empty distinctions remain visible.

Measured mobile inputs are 32px high. The first choice ends at 579px on a 390px
screen and 614px on a 320px screen, inside the tested 900px viewport. Caption and
page containment checks pass. This is a bounded Chromium review, not an assistive
technology certification.

Remaining follow-up: existing instrument coverage warnings are verbose on mobile;
move detailed qualifications into existing support/read details while retaining
concise point/instrument counts. More importantly, planning must explain the
current reported-instance evaluation policy before Run. OTLP does not universally
require `service.instance.id`; generic reported-stream evaluation needs its own
architecture decision. These are tracked in [433](../ROADMAP_2026_11_14_CHECKPOINT_430.md).
