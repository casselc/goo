# Trace-wide search inside typed OTel environments — 340

Executable: [340-typed-trace-search.spec.js](../../journeys/340-typed-trace-search.spec.js).
Actor: [typed-trace-search.ts](../../verification/bombadil/typed-trace-search.ts).
Fixture: [ordinary OTLP data](../../journeys/340-typed-trace-search-fixture.js).

Status: deployed-source native, observed-model and bounded-actor acceptance passed.

This journey emits ordinary spans and correlated logs with standard resource
attributes. It creates no Goo manifest, saved view, capture, document or receipt.
The API, database and worker span conditions can match different reported spans
of one trace. Matching does not imply parent-child causality or a complete trace.

1. In Activity, select spans and discover environments. Preview the new named
   environment, stage it and explicitly start the stream. Inspect the reported
   database operation and use **Service actions → Open service**. Read its service
   overview explicitly, then Customize the database client view and Run the staged
   population. The effective typed environment travels through Activity, Services
   and the existing template handoff. Empty/unreported cases instead use Inspect
   → **Analyze this observation window**, then **Run analysis** in the staged root
   fixed-question surface.
2. Inspect that operation and choose **Find traces with this operation**. Add a
   separate `status_class = error` span condition. Staging reads nothing. Run
   trace search explicitly. Fifteen distinct trace IDs match; the default twenty-item cap lists all fifteen.
   Parent population filters apply only to the first condition.
3. Inspect the first legacy-only cross-service result. Its database and error
   witnesses have different services and span IDs. Witness fields include both
   environment resource keys: missing is distinct from an explicitly empty value.
   Local witness disclosure performs no source request. Independent full-list
   and limit-two probes check all fifteen matches, a same-span witness case, a
   disconnected worker witness (no parent edge required), and
   exact `hasMore` without substituting the presentation limit for the total.
4. Choose **Open trace (fresh read)**. The span and correlated-log reads retain
   the exact environment policy, fixed interval and selected trace ID, and do not
   inherit the search conditions. A legacy-only result returns three actual spans
   and one log; replacing its effective selector with a scalar current-key name
   would fail this physical-data check. Close returns to the held search without
   rereading it.
5. Repeat from a current-only named fixed question, then Activity's empty and
   unreported environment scopes. Current-only has thirteen matches, empty has
   two, unreported has one. Present current attributes override legacy attributes,
   including present-empty; missing both keys belongs only to unreported.
6. Hold an actual search response, close the owner, then release it. It must not
   resurrect results. Reopen and return an otherwise real response with a weaker
   echoed policy: admission must reject it before displaying trace actions.

The four per-profile observed ledgers have six events each: parent Run, local
staging, explicit search, witness disclosure, exact trace opening, and retained
return. Their source budgets are respectively 1, 1, 2, 2, 4, 4. Direct ingestion
readiness/full-list/bounded-list probes, Activity environment adoption and the late/echo
supplement are outside that relative ledger. Counts are HTTP source operations,
not physical SQL statements. The existing trace-search Quint model checks these
observed effects; an independent resource/predicate oracle checks actual records.

The bounded actor uses the effective typed fixed question over the same actual
OTLP fixture, with explicit parent Run, operation search, error condition, bounded
results, witness disclosure, exact trace/log opening and retained return. A
positive terminal state must verify four outcomes and policy preservation with
four source reads. Actor captures are not accepted as visual proof without an
independent audit; current Playwright viewport frames supply the visual record.

The final-source native run passed in 3.1 minutes (observed model 31.021 seconds),
covering all four profiles, with no browser errors and
three spans plus one correlated log in every exact-trace continuation. All four
retained returns and explicit search closes restored an unobstructed focused
control, checked both in the viewport and by pointer hit testing. This caught and
fixed two real return-navigation defects: an offscreen opener and an opener
covered by the sticky trail. Local Close performed no source request.

The 40-second actor recorded 435 snapshots covering all nine phases, with four
verified outcomes, policy preservation and four source reads. Its screenshot
stream was independently audited: 435 images contained only four unique images,
and 424 repeated an intermediate search-result frame, including the terminal
image. Those captures do not prove terminal appearance. The maintained viewport
frames below are the visual record. The browser/model were subsequently rerun
on the identical deployed source (2.4 minutes; observed model 19.66 seconds);
the compact receipt explicitly distinguishes that fixture from the actor run.
The maintained frames and latest observed ledger use fixture `af6a0d3863`
(search read at `2026-11-05T23:29:07.993764328Z`); the 435-snapshot actor and
independently audited capture use fixture `237a435bba`.

- [Effective search](screenshots/340-effective-trace-search.png)
- [Empty scope, narrow viewport](screenshots/340-empty-scope-narrow.png)
- [Unreported witnesses](screenshots/340-unreported-witnesses.png)
- [Checked exact trace](screenshots/340-unreported-checked-trace.png)

The fixture refreshes Activity data only after its previous version has left the
actual 60-second observation window, retaining both versions in the independent
oracle. It does not change the product window or replace real data with mocks.
The journey does not add arbitrary trace joins, a causal path query, upstream
completeness or automatic background searches.
