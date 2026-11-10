# 400: author compact population summaries

The optional typed Summary child consumes the existing source summary dataset.
Records remain required; no grouping, additional summary SQL or new evaluator is
introduced. Chosen fields and their order are local presentation, including
mixed-family count, numeric, duration and recorded-status facts.

Run heavy checks serially with Node 24, a reviewed Go binary, the retained lab
collector/ClickHouse and loopback/browser access:

```sh
GOO_GENERIC_BINARY=/tmp/goo-400-reviewed-goo NO_PROXY=127.0.0.1,localhost npx playwright test journeys/400-population-summary.spec.js --workers=1
GOO_GENERIC_BINARY=/tmp/goo-400-reviewed-goo NO_PROXY=127.0.0.1,localhost npx playwright test journeys/400-population-summary-actor.spec.js --workers=1
```

The main ingests 36 uniquely scoped physical ordinary logs/spans. Its three
source/destination cases retain nested predicates, two environment policies,
exact windows and nondefault page limits without grouping. Physical raw presence,
values, nanosecond timestamps/durations and reported status feed independent
sum/count/reduced-mean/support checks. Additional source questions cover empty
input, no valid numeric values and a valid numeric zero.

The report is `artifacts/population-summary400/summary.json`. It separates source
requests/physical SQL, stored-definition GETs and browser configuration
publications. Destination setup uses three additional explicit fixture
publications. Forty measured boundaries and thirteen explicit source requests
were observed; exact outcomes and limits are recorded in validation. Current-definition
publication, source-free staging/field edits, save/reopen/reuse, held Back and
retired hidden/inert/pending/superseded callbacks are observed separately.

The actor consumes the main's same-day source bank; rerun main if its telemetry
expired. It writes `actor.json` and a unique Bombadil artifact directory. It
independently runs the source and child population, chooses numeric sum, moves it
before count and mean, and returns to the exact held source summary. Its DOM
observations are not a pixel oracle. No lab telemetry, pending streams or stores
are reset by either test.

Actual execution, frame review and generated/model checks are pending root's
serialized acceptance; syntax checks alone do not establish a passing journey.


Executed results: the main passes in about one minute (1.1 minutes runner), with
40 boundaries, 13 source requests, nine configuration GETs, three browser
publications and 39 physical SQL queries. Three destination setup publications
are counted separately. Bombadil passes in 36.3s (37.7s runner), recording 400
snapshots across nine phases and two reads/six physical queries. All three frames
were reviewed. Neither campaign uses a pixel oracle for source correctness.

From the verification module with Go 1.27.3, Node 24 and Quint on PATH:

```sh
GOO_SUMMARY_VIEW_400_REPORT=/workspace/goo/artifacts/population-summary400/summary.json \
GOO_SUMMARY_VIEW_400_ACTOR=/workspace/goo/artifacts/population-summary400/actor.json \
go test -run 'Test(HegelSummaryView400|QuintSummaryViewObserved400)$' -count=1 -v
go test -run '^TestSummaryView400WireParity$' -count=1 -v
```

Observed Quint passes in 6.49s, rejecting eight semantic mutations and the
staging-read control. Hegel passes 48 generated local authoring/reuse/store cases
in 1.83s. These are distinct from actual source reads; neither check skipped.
The 89 leaf and 30 carrier cases pass actual Go/JS admission/store and API
input/output shape checks (`tests/summary_view_400_api.py` using the API venv).
