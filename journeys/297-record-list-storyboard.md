# Inspect supplied records as a reusable list — 297

Executable: [297-record-list.spec.js](297-record-list.spec.js).
The main journey starts a generic Go runtime without seeded definitions and uses
real OTLP database-client spans from two services. Ingestion/readiness polling is
separate from browser request accounting.

1. Open the discovered source service and Customize its database view. Run the
   exact fixed query. A selected table row blocks switching the records kind;
   a retained disabled control cannot discard that selection.
2. Choose qualified resource/attribute fields, then switch **Records view → List**.
   Switching is local. The new kind consumes the same bounded supplied page;
   **Use Table for selection and source actions** makes the first cut's scope clear.
   **Details** opens lazily and inspects only supplied records; it does not fetch.
   [Supplied list](../docs/journeys/screenshots/297-supplied-list.png).
3. Select p95 and Table for the aggregate. Its category and selected measure are
   the first exact columns, with the p95 header and data cell inside the initial
   viewport. Other fields remain available, and the source snapshot is unchanged.
   [Selected measure](../docs/journeys/screenshots/297-selected-measure.png).
4. Save the list as a reusable template. Open another service, explicitly choose
   a different grouping, inspect that template and Apply. The target query and
   supplied values remain its own; the records kind and qualified field choices
   travel with the saved composition.
5. After a late arrival, page the List explicitly. The page reports its fresh
   count and rows while the aggregate retains its earlier basis/read clock.
   New page details start closed. Retired kind/detail controls cannot act on it.
   [Independent page](../docs/journeys/screenshots/297-paged-list.png).
6. Switch back to Table and List locally, Save the target and reopen it. Reopen
   stages configuration; only explicit Run reads the current fixed question.

A separate actual-OTLP log fixture seeds only a checked display configuration to
exercise two supplied logs, qualified resource/attribute values, present-empty vs
missing fields, lazy details, serialized HTML attributes, detached property updates
and refusal/withdrawal of forged page metadata. This fixture is not represented as
a second unseeded authoring journey.

Generated properties exercise the actual Go declaration/template validator and JS
record declaration/public-page model, including null/rich values, omitted unknown
continuation and malformed/accessor inputs. The observed Quint lifecycle must
reject six corrupted ledgers and an unfinished prefix. The independent Bombadil
campaign observes local Details and kind switching, then one explicit page read
with the exact expected rows and held aggregate; click completion alone is insufficient.

Validation passed: both browser fixtures (1.7 minutes), 100 Hegel cases, the
21-step observed Quint lifecycle, six corrupt-ledger negative controls and an
unfinished prefix. Bombadil exercised 482 snapshots and seven completed phases:
one Explore, one explicit page and zero writes. The main browser budget was six
Explore requests (two service overviews plus four explicit Runs), one page, three
template expansions and two view saves, with no captures, documents or Keep writes.
The 296 template and 289 independent-page regressions also passed.

Frames are generated from the running app. Screenshot capture hides the unrelated
sticky investigation trail; the selected-measure frame captures the table viewport
at its original horizontal position.
