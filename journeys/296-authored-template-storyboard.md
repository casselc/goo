# Author one database template; keep the destination question — 296

Executable: [296-authored-template.spec.js](296-authored-template.spec.js).
The generic Go runtime starts without saved definitions. Six standard OTLP Client
spans are ingested through the collector into ClickHouse for two services and
namespaces. They are fixtures representing database-client telemetry, not proof
of a running database server or complete upstream instrumentation.

1. Discover the catalog service and open its **Database client activity** preset.
   **Customize database client view** expands the checked operation-grouped query
   without reading again. Run acquires three matching client-span records.
2. A local Find literal makes the reusable-template declaration unavailable. Save
   refuses without a POST; clearing Find repairs it without reading. Save a regular
   working view, then dispatch an adversarial change from the disabled Data profile
   control: it cannot clear the saved identity or create a draft/publication.
3. Change the actual aggregate to **Table** and **duration p95**. Enable **Reusable
   template**, choose **OTel database client spans**, name and Save the definition.
   These edits use native controls; no source JSON or second catalog is needed.
   [Author controls](../docs/journeys/screenshots/296-author-template.png).
4. Open a different service/namespace. Customize the same DB preset, explicitly
   replace its grouping with `attribute:db.namespace`, then Run. The source groups
   are `SELECT`/`INSERT`; the target groups are `ledger`/`archive`. The services,
   namespace selectors and fixed windows also differ.
5. In Saved exploration views, **Inspect template** previews the concrete target
   field, profile requirements, table/p95 and record columns. A deliberately delayed
   expansion cannot become applicable after the current title/draft changes; it
   remains disabled with an Inspect-again reason. A fresh preview succeeds.
   [Checked preview](../docs/journeys/screenshots/296-template-preview.png).
6. **Apply template** changes presentation only. The target query, current held
   data and request/write budget remain unchanged. The existing target rows render
   as the authored table; an explicit Run then sends exactly the target query.
   [Applied composition](../docs/journeys/screenshots/296-applied-template.png).
   The unrelated sticky investigation trail is hidden only for this capture's
   framing, using Playwright's screenshot style; the UI/data are not mocked.
7. Save the destination definition and reopen it. Opening stages configuration
   without a source read. Run again reads the destination query and displays the
   `ledger` p95 of 9 ms. The original source template remains byte-equivalent and
   independently saved.
   [Reopened table](../docs/journeys/screenshots/296-reopened-template.png).

The observed sixteen-step Quint ledger distinguishes two service-overview Explore
requests and four explicit population Runs from three explicit view saves. It
rejects six corrupt observations and an unfinished prefix. The main scenario makes
four pure template-expansion calls (two presets, one retired preview and one current
preview), no capture/document publication. This is HTTP accounting, not physical
ClickHouse query counts; fixture ingestion/readiness polling is separate.

Hegel runs 50 successful bindings and 50 refusals through the actual Go expansion,
then validates successful output with the actual browser request/response module.
It varies signal/profile, environment policy, Boolean predicates, qualified groups,
windows, paging, duration support and detached composition. Explicit empty equality
values/false optional flags are checked against Go's canonical wire forms, while
changed service/group qualification still refuses. Source Find, OR-weakened required
conditions, fake source identity and unsupported shapes are negative cases.

The independent 30-second Bombadil actor opens the saved destination, reads it,
changes its kind to Bars, discovers the source template through the actual catalog,
previews and applies it. Completion requires the real Table/p95/target-row outcome.
Its fetch observer delegates requests unchanged and counts POST writes by method,
so catalog GETs are not mislabeled as writes. The passing campaign records 497
snapshots across six phases, one Explore request, one expansion, zero writes and
no viewport overflow.

Validation: the final question-codec main journey and observed Quint pass in 36.5
seconds. The independent Bombadil campaign above passes. Existing HTTP preset properties,
HTTP browser/Quint/Bombadil and contextual-comparison regression also pass. Native
control interaction is covered; no screen-reader, mobile or production-load study
is claimed. The p95 table's existing wide exact-value layout is not redesigned here.
