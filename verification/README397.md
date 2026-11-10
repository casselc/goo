# 397: ordinary quantitative customization

Run heavy checks serially. The main journey ingests unique ordinary OTLP logs and
spans into the retained lab and starts a fresh reviewed Go runtime. It does not
reset existing telemetry or stores. Node 24, loopback access, ClickHouse and OTLP
collector are required.

```sh
GOO_GENERIC_BINARY=/tmp/goo-397-reviewed-goo NO_PROXY=127.0.0.1,localhost npx playwright test journeys/397-quantitative-customize.spec.js --workers=1
GOO_GENERIC_BINARY=/tmp/goo-397-reviewed-goo NO_PROXY=127.0.0.1,localhost npx playwright test journeys/397-quantitative-customize-actor.spec.js --workers=1
```

The actor consumes the same-day main report at
`artifacts/distribution-customize397/summary.json`; rerun main if the bank expired.
Actual source work, physical proxy SQL, configuration writes, raw records and
independent aggregates are retained there. Actor observations and its unique
campaign directory are recorded separately in `actor.json`.

The main covers numeric logs, numeric spans and mean span duration. Exact
canonical integer arithmetic checks valid/missing/empty/invalid support, signed
and zero values, and separate per-group sums and duration sums/counts. Staging,
local changes and held Back are measured separately from explicit source reads
and configuration publication. Neither configuration storage nor artifact GETs
are counted as telemetry reads. The actual main passes in 58.6 seconds (about one minute runner), with 55 measured
facts and 36 physical ordinary records. There are 21 source requests, nine
configuration GETs, three browser configuration publications (plus three explicit
destination-setup publications) and 75 physical SQL attempts
(including three deliberately failed source Counts). The independent Bombadil
campaign passes in 41.3 seconds (43.7 seconds runner), with 295 snapshots across
eight phases and eight campaign SQL reads after four setup reads. It is not a
pixel-verification claim. Current desktop and narrow frames are reviewed.


From the verification module, with Go 1.27.3, Node 24 and Quint on PATH:

```sh
GOO_DISTRIBUTION_CUSTOMIZE_397_REPORT=/workspace/goo/artifacts/distribution-customize397/summary.json \
GOO_DISTRIBUTION_CUSTOMIZE_397_ACTOR=/workspace/goo/artifacts/distribution-customize397/actor.json \
go test -run 'Test(HegelQuantitativeCustomization397|QuintQuantitativeCustomizationObserved397)$' -count=1 -v
go test -run '^TestQuantitativeView397WireParity$' -count=1 -v
```

The shared 58-vector corpus runs through actual Go/JavaScript validators and
accepted immutable-store round trips. `tests/quantitative_view_397_api.py` runs
its schema-shape expectations separately; query-supported measures and UTF-8
byte/control/Unicode semantics remain runtime checks. Generated local properties
are not additional source observations. The observed ledger independently checks
raw records, aggregates, exact means, source plans and the actor before model
replay; semantic mutants and a staging-read control must fail.

Executed property/model results: 48 generated local vectors passed in 0.53s;
observed Quint passed in 13.11s, with all seven semantic mutants and the
staging-read control rejected. Neither check skipped.
