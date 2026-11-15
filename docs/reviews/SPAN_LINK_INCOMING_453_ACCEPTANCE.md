# 453 scoped acceptance

453 adds inverse discovery from a selected span: which span observations report
a link to this exact trace/span pair? It reuses the ordinary query model and the
452 staged lookup/held-return owner. 454 remains the consolidation step.

## Architecture and data semantics

Independent review accepted the `linksTo` selector rather than representing
correlated link arrays as independent string predicates. The compiler checks
same-index membership and aligned trace/span/attribute arrays without expanding
span records. Repeated link entries cannot multiply the matching record count;
separately stored observations remain separate records.

The existing spans source profile checks nonnullable link-column types before
population reads. It adds two metadata SQL calls to a successful acquisition.
Unsupported schemas refuse explicitly. Every returned row independently passes
membership, environment and fixed-window admission.

Fixed questions, clones, captures, packages and supported complete-query displays
retain the selector. Unsupported specialized transformations and rolling tails
refuse before acquisition. Main-editor Run and counted-value requests preserve
it. Review found no remaining architecture blocker.

## UX, accessibility and web standards

Independent source review requested human page ranges instead of raw offsets,
fresh-read wording and a directional disclosure name. The panel now shows, for
example, **105 matching span observations · 101–105 shown**, explains that pages
may change, and names the disclosure **Question and reference target**. Exact
offsets remain in the query/provenance. A failed page retries that requested page.

Follow-up review accepted actual 1440/390/320 screenshots. The shared responsive
record layout remains readable; Previous/Next states match the result. Stage
focuses Run. Enter on Run or Next retains visible Back focus while the source
request is pending and after rendering; native dialog containment and exact
held scroll/focus return remain intact.

An initial screenshot helper scrolled the entire 100-row panel under its sticky
header. The helper was corrected to frame the active control. Separate real
interaction probes, before/after reads and after layout settles, check that Back
is below the sticky heading and passes a center hit test. These passed on the
unchanged product binary; capture-helper behavior is not reported as a product fix.
No concrete scoped UX or accessibility/web-standards blocker remains. This is not
a broad WCAG, production authorization or cross-browser audit.

## Next consolidation

Adding a selector exposed manual query reconstruction across several owners.
454 should consolidate a small number of identical pure transformations and add
a shared selector-conservation case bank: preserve the complete population or
refuse before IO; make deliberate widening explicit. Keep specialized profile
rules local. Do not introduce a universal controller, policy registry, readiness
cache or another carrier. Retain raw metric visuals as the following cycle.

Executed source, browser, actor and model checks are in the
[validation record](../../VALIDATION.md) and
[storyboard](../journeys/span-link-incoming.md).
