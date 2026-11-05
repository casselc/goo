# Narrow both fixed comparison windows with a categorical facet — 337

Executable: [337-paired-facet.spec.js](../../journeys/337-paired-facet.spec.js).
Actor: [paired-facet.ts](../../verification/bombadil/paired-facet.ts).

The executable now follows the 338 paired-value contract: each **Read values**
reads both original windows, with four category lookups in its terminal ledger
and two additional lookups in the supplemental check. The original 337 run below
is historical; current 338 regression validation is recorded separately.

This journey emits ordinary OTLP logs and server spans with `service.name`,
`service.namespace`, `http.request.method` and `http.route`. The API service uses
the current environment key; the worker uses the legacy key. Activity discovers
their effective named environment and stages its existing previous-window count
comparison. The UI adds the existing categorical facet, saves the definition and
later exports/reopens it. No view, capture, receipt or telemetry response is seeded.

1. Open View settings → Facet and choose `attribute:http.request.method` with the
   title HTTP methods. Adding it and saving the original comparison read no
   telemetry. Close settings and Run explicitly.
2. The selected window has 107 records, although its inspection page contains
   only 100. The previous window has 13. Read both windows explicitly; selected values are:
   GET 75, POST 27, present-empty 2 and missing 3. These are full applied-scope
   counts, not counts synthesized from inspection rows. The ordinary record-list
   inspector also refuses a route-field gesture while the declared facet is the
   method field, without a misleading staged message, invalid tray or source read.
   [Pair and selected values](screenshots/337-pair-and-selected-values.png).
3. Exclude GET. The held pair stays unchanged while the tray explains **Both
   windows**. Parent Run is disabled; directly invoking its retained handler also
   leaves the pair and pending choices unchanged, without reading the source.
   Clear staged re-enables Run. Choose **Filter both windows** to stage a child, without a source
   read. Run the child explicitly: the independently read populations are 32 and
   6. Each request keeps its original window/policy and appends the same bounded
   categorical exclusion, including missing values.
   [Common filter intent](screenshots/337-filter-both-windows.png).
4. Read the selected child's values explicitly, then Include POST. Another common
   child stages without reads; Run independently observes 27 selected and 4
   previous records. Both predicates retain the earlier exclusion.
5. Follow the previous-window API service group. Its records-only child stages
   the exact previous operand plus the service predicate; one explicit page read
   observes three actual POST records in that exact interval. Back restores the
   supplied pair, Find text, presentation and group-action focus without reads.
6. Save the narrowed comparison and export YAML. Both fixed questions and the
   categorical facet survive; pending choices and observations are not the saved
   definition. Reopen stages configuration without telemetry; Run reads 27/4 anew.
   [Reopened common scope](screenshots/337-reopened-common-scope.png).
7. A supplemental logs check reads both current operand values once, retains
   facet Find/counts through another staged child and Back, and reattaches an old
   Exclude control. The retired control must remain inert; subsequent local
   gestures read nothing. These lookups are outside the terminal model ledger.
8. Repeat the scope/count/record checks for OTel spans and check the 390px viewport.
   [Narrow comparison](screenshots/337-paired-facet-narrow.png).

## Current executable request budget (338)

The observed model ledger covers the first logs task from the saved original
definition through reopened explicit Run: nine population/page reads, four value
lookups, four comparisons, one final Save and one YAML encode. Readiness polling,
Activity adoption, original Save, the supplemental lookups and the separate spans
task are outside that relative ledger. The model checks effects/completion, not
physical SQL query counts, cancellation or private source authority; Go/JavaScript
semantic properties and the actual
requests/returned records provide the independent scope checks.

The bounded actor opens the genuine UI-saved original logs definition, explicitly
reads the full 107/13 pair and both value partitions, stages Exclude GET with zero reads,
then explicitly obtains 32/6. Admission requires four independently verified
transitions and a positive terminal snapshot, six source reads (two value
lookups), two comparisons, zero writes and unchanged retained parent facts.

Current 338 regression passes in 2.9 minutes (3.0 total), including the same
logs/spans counts, three supporting records, retired-control refusal and four
refreshed frames. Its observed 17-step model replay passes in 22.884 seconds:
nine population/page reads plus four category lookups, thirteen source requests.
The supplemental check adds two explicit operand lookups and zero local-gesture
reads. The adapted 337 actor was not rerun this round; the separate
[338 campaign](338-category-pair.md) exercises the current paired-value path.

## Historical 337 validation

The historical 337 deployed-source journey passed in 2.3 minutes for both logs and spans,
including the disabled-Run/retained-handler checks and four refreshed viewport
frames. Its 17-step observed Quint ledger passes in 20.331 seconds. The 35-second
actor campaign passes with 448 snapshots across phases 0–7 and the four verified
transitions above. That original actor used one selected-value lookup and five
source reads; the current paired-value actor budget is two lookups/six reads.
No browser errors or prohibited capture/document writes occur;
the wrapper restores demo traffic and retains all 51 pre-existing stale pending
stream entries.

Actor screenshots are not used as visual proof: this campaign records 448 JPEGs
with four distinct hashes, of which 441 repeat a child loading state although
the final DOM/API snapshot is applied. An earlier campaign repeated one initial
loading frame. The current Playwright viewports provide the reviewed visuals;
this driver-capture observation is not attributed to a product rendering cause.

The original 337 cut did not provide values from both windows or baseline-only
category discovery; 338 adds those capabilities. Upstream completeness, rates,
frozen-capture filtering and selected/rest set reinterpretation remain outside
this task.
