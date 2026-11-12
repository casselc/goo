# Recent and pinned scope review — 414

Status: accepted for milestone 414. This report is
an independent product/ownership/accessibility review of milestone 414, not a
whole-product accessibility certification.

## Scope and findings

Reviewed the actual configuration journal, shared picker, Activity staging,
fixed-question integration and retained-visit return. Inspected generated
1440/390/320px frames and the final real-source journey report. The passing
47.8-second browser run records 18 facts, fixed counts of 624 and 156 with 100
displayed rows each, and no page errors. Two explicit fixed executions cost six
SQL statements; explicit stream Start adds one. Staging, pinning, dirty-scope
refusal, held return, configuration-only reload, clearing and disabling
persistence add no telemetry reads in their recorded checks. The completed
30-second actor campaign records 534 snapshots while API POST/SQL counters remain
3/7 before and after.

An earlier journey expected zero visits after reload, although the fixed URL
creates one newly staged visit. The corrected check distinguishes that new view
from retained result rows; it does not weaken configuration-only persistence.

The implementation keeps configuration separate from supplied results. Activity
remains a rolling selector; fixed intent retains exact bounds. The journal starts
in memory, persistence is opt-in and path-scoped, and browser-profile sharing is
explicit. Recent entries expire after 30 days; pins remain until removal, with
the twenty-first pin refused. No server personal-history or authentication claim
is introduced.

Review identified and follow-up source inspection confirmed fixes for:

- Held return from Activity must stop the stream and foreground the fixed owner,
  including the current trail entry detached behind Activity.
- Pin current must choose the foreground owner; a dirty Activity scope must not
  fall back to an unrelated hidden fixed question. Measurements hides the picker.
- Recalling Activity must refuse to overwrite a dirty scope/title. Explicit Start
  remains separate from staging.
- Recalled fixed display selection is forwarded rather than silently discarded.
- Re-rendered picker callbacks bind a revision. Keyboard row actions recover focus
  to the disclosure summary when their button is removed. Row controls have
  contextual descriptions.
- Persistence and individual configuration limits use UTF-8 bytes. Nested
  unsupported fields are refused before storage.
- Held entries distinguish their visit ordinal, fixed window and selector;
  supplied results and a merely staged view have different labels.

The optional disclosure fits narrow screens without horizontal overflow. Its
bounded inner scroll keeps a loaded list from becoming an unbounded page. Exact
scope text is necessarily denser than the ordinary Activity stream; it appears
only while the picker is open. A future refinement can abbreviate repeated scope
metadata further, but must preserve distinguishable destinations.

## Verification limits

The isolated Hegel package passes 64 cases in 8.163 seconds, exercising the actual
JS Journal through a source-denied bridge against an independent Go state oracle.
Each case forces capacity/expiry boundaries then generates 100 state commands.
It covers ordered deduplication,
page-zero intent, pin capacity, expiry, clear/reset, opt-in persistence and reload,
plus root/nested poisoned-field controls. It does not establish browser storage
security, authenticated isolation, UI lifetime correctness or exhaustive product
model coverage.

The browser journey supplies real OTLP records and independently checks source
membership and SQL budgets. Final frames include the held-label correction:
visit IDs, selectors and held-result versus staged-view meaning are visible.
The 390/320px layouts remain contained; the optional list scrolls internally.
No physical screen-reader, cross-engine or full WCAG conformance claim is made.

Actual commands and final outcomes belong in [validation](../../VALIDATION.md).
