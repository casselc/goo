# Captured record rates (406)

`record.rate` is exact captured count × 1000 / actual elapsed milliseconds. It is
not an estimate of unsampled traffic or a bounded percentage. A zero with known
exposure remains derived; an unreported interval has null rate and no support facts.

The Hegel test drives public source projection through a deterministic transport,
then preserves complete summary/timeline/store/template carriers across Go and
JavaScript. UInt64 extrema and clipped buckets are synthetic arithmetic inputs,
not additional source observations. The API harness checks wire shape only.

The actual journey report is checked against all 24 raw ClickHouse records,
qualified predicates/environment policies, query windows and exact bucket bounds.
It verifies destination scope, clock/basis, source plan and local-operation budgets.
Quint replays 24 observed boundaries; a planted source read during local selection
must fail. Bombadil's separate actor is matched to the same raw source bank.

Run serially with the project toolchain and infrastructure:

```sh
(cd verification && go test . -run 'Test(HegelRecordRate406|RecordRate406SupportProof)$' -count=1)
python3 tests/record_rate_406_api.py
GOO_GENERIC_BINARY=/path/to/goo npx playwright test journeys/406-record-rate.spec.js --workers=1
(cd verification && GOO_RECORD_RATE_406_REPORT=$PWD/../artifacts/record-rate406/summary.json go test . -run TestQuintRecordRateObserved406 -count=1)
GOO_GENERIC_BINARY=/path/to/goo npx playwright test journeys/406-record-rate-actor.spec.js --workers=1
(cd verification && GOO_RECORD_RATE_406_REPORT=$PWD/../artifacts/record-rate406/summary.json GOO_RECORD_RATE_406_ACTOR_REPORT=$PWD/../artifacts/record-rate406/actor.json go test . -run TestRecordRateActorObserved406 -count=1)
```

Missing source or actor reports skip their opt-in tests. A successful skipped test
is not observed coverage. Unknown interval behavior is verified by semantic
properties; the finite actual source journey exhaustively reports known zeros.
