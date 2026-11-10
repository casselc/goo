# 401: gauge context in Services and Saved

Run heavy checks serially with Go 1.27.3, Node 24, a reviewed binary, the retained
lab collector/ClickHouse and loopback/browser access:

```sh
GOO_GENERIC_BINARY=/tmp/goo-401-reviewed-goo NO_PROXY=127.0.0.1,localhost npx playwright test journeys/401-retained-gauge-context.spec.js --workers=1
GOO_GENERIC_BINARY=/tmp/goo-401-reviewed-goo NO_PROXY=127.0.0.1,localhost npx playwright test journeys/401-retained-gauge-context-actor.spec.js --workers=1
```

The main ingests a unique ordinary OTel gauge/log/span bank. Six cases cover
Services and immutable Saved evidence in named, missing and present-empty
namespaces; current-only/effective environment policies; real zero and unavailable
gauge samples; exact lower/upper clipping and raw record membership. Setup capture
publications and their source SQL are explicitly separate from observed browser
operations. Service discovery/dashboard setup reads are recorded as setup facts;
Stage, Return, retired callbacks and local selection must add no source/store IO.

Main observations are written to `artifacts/retained-gauge401/summary.json`.
Physical raw metric/record rows, exact emitted questions, parent DOM identities,
selected points, receipt bytes, actual request/response pairs and proxy SQL remain
available to the independent verifier. One Saved record response is deliberately
held until after Return; its aborted late delivery cannot alter the parent.

The actor consumes the main's same-day source bank and makes its own real gauge
capture as explicit setup. It selects an unavailable Saved sample, stages nearby
logs, explicitly reads them and returns to the unchanged selected gauge/receipt.
It records separate setup and actor budgets in `actor.json` and a unique Bombadil
directory. No positive source result is mocked. Telemetry and existing stores are
not reset. Executed results are recorded in [VALIDATION](../VALIDATION.md); the storyboard
links the actual desktop and narrow frames.

Run the local carrier and required observed checks serially from `verification/`:

```sh
go test -run '^TestHegelRetainedGaugeContext401$' -count=1 -v
GOO_RETAINED_GAUGE_401_REPORT=../artifacts/retained-gauge401/summary.json GOO_RETAINED_GAUGE_401_ACTOR=../artifacts/retained-gauge401/actor.json go test -run 'Test(Quint|Bombadil)RetainedGaugeObserved401' -count=1 -v
```

From the repository root, run `python tests/retained_gauge_401_api.py` in the API
virtual environment after generating the carrier corpus. Each explicit child
executes the existing count, rows and field-discovery SQL; this is three physical
statements, distinct from dashboard/setup reads.
