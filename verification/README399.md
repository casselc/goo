# 399: passive display helper consolidation

The structural change removes an application import from display formatting and
coordinates. Existing source ownership, telemetry queries, scope transformations
and publication contracts remain in their application adapters.

Run heavy checks serially with Node 24, a reviewed Go binary, the retained lab
collector/ClickHouse and loopback/browser access:

```sh
GOO_GENERIC_BINARY=/tmp/goo-399-reviewed-goo NO_PROXY=127.0.0.1,localhost npx playwright test journeys/399-display-helpers.spec.js --workers=1
GOO_GENERIC_BINARY=/tmp/goo-399-reviewed-goo NO_PROXY=127.0.0.1,localhost npx playwright test journeys/398-grouped-mean.spec.js --workers=1
```

The 399 journey ingests a unique twelve-span ordinary OTLP bank and checks the
persisted timestamp, duration, reported status and numeric/group attribute
presence. An independent exact arithmetic oracle checks per-cell mean, p95,
status fraction and numeric magnitude geometry. All local measure switches,
Find, staged fixed questions and held returns have separate source/physical-SQL counters.
A second actual question covers duration-bucket rectangle dispatch through the
application-owned scope transform. It stages `[5,50) ms` with exact inclusive
nanosecond predicates, without another source read.

The full report is `artifacts/display-helpers399/summary.json`. Desktop/narrow
frames and a rectangle frame are captured by the actual app. No mock positive
source response or performance improvement is claimed. Source-free baseline
vectors and dependency/effect checks complement these application observations;
they do not replace real source membership. The actual 399 main passes in 22.0 s (25.9s runner): 12 physical spans, eight metrics,
16 measured boundaries, two source requests and 12 physical SQL queries. The
shared 398 mean journey also passes on the reviewed 399 binary (21.2s/23.7s runner),
as does its independent Bombadil companion (36.5s/37.6s runner; 323 snapshots,
nine phases, 12 physical SQL queries). Design reviewed all three new 399 frames.


Run local regressions using the repository Node module-mock flag:

```sh
node --experimental-test-module-mocks --test tests/distribution-values-399.test.mjs tests/display-modules.test.mjs
npm run test:visualizations
```

The focused helper/migrated-consumer check passes 34 tests; the full frontend
passes 1,452 tests with no skips. The frozen `011580e3` fixture is test-only old
production formatting code, not a product compatibility layer. The helper graph
is exactly four pure modules; the public passive display graph is 59 modules and
still refuses application/source dependencies.

Required 398 observed/model revalidation uses the fresh 398 main/actor reports on
the 399 binary, with commands in [README398](README398.md). It passes in 7.90 s; all
seven semantic mutants and the staging-read model control fail. Hegel passes 32
local staging/reuse/store vectors in 1.63 s. These checks cover the existing shared
grouped contract; the separate 399 legacy-renderer journey supplies its own raw
record/geometry oracle. No new 399 model or schema is required by this refactor.
