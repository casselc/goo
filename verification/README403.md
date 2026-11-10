# 403: trace-search continuation

Run heavy checks serially using Go 1.27.3, Node24, Quint0.32 and the preserved lab.
From `verification/`:

```sh
go test -run 'Test(HegelTraceContinuation403|TracePositionHashGolden403)' -count=1 -v
GOO_TRACE_CONTINUATION_403_REPORT=../artifacts/trace-continuation403/summary.json GOO_TRACE_CONTINUATION_403_ACTOR=../artifacts/trace-continuation403/actor.json go test -run 'Test(Quint|Bombadil)TraceContinuationObserved403' -count=1 -v
```

The property test partitions fixed generated populations into complete page
carriers, including terminal empty pages with a positive full total. The existing
independent set-intersection/witness oracle checks their complete source bank.
Go/JavaScript hash vectors include Unicode, HTML characters, U+2028/U+2029,
empty literals and exact numeric text. Runtime admission mutations are distinct
from API-shape proofs. From the repository root, after generating the corpus and
current OpenAPI, run `python tests/trace_continuation_403_api.py` in the API venv.

Actual source acceptance uses:

```sh
GOO_GENERIC_BINARY=/tmp/goo-403-reviewed-goo NO_PROXY=127.0.0.1,localhost npx playwright test journeys/403-trace-continuation.spec.js --workers=1
GOO_GENERIC_BINARY=/tmp/goo-403-reviewed-goo NO_PROXY=127.0.0.1,localhost npx playwright test journeys/403-trace-continuation-actor.spec.js --workers=1
```

The observed verifier checks the complete raw fixture bank, lexical page IDs,
qualified witnesses, totals before positioning, exact source SQL/bindings, immutable
positioned capture bytes, source-free reopen, authored root intent and held return.
The Quint replay checks ownership/read/publication boundaries and a planted local
read; ten independent semantic mutants check source, scope, custody and state
admission. The actor performs its own actual first → next → first reads against
the same source bank; it does not establish a stable snapshot across reads.

Executed results belong in [VALIDATION](../VALIDATION.md). Optional observed tests
skip without their report environment variables; required acceptance supplies both
reports and records zero skips. No existing telemetry or store needs resetting.
