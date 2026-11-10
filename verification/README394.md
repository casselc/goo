# 394: direct distribution customization

Accepted ordinary Count Bar/Table distributions can stage the existing population
workspace directly. Configuration carries the complete source question; staging
carries no chart evidence. Explicit Run supplies child data. Held Back restores
the original read, Find text and presentation without telemetry.

Run actual jobs serially against an existing collector/ClickHouse and the reviewed
Go binary. From the repository root, with Node 24 and the installed browser tools:

```sh
GOO_GENERIC_BINARY=/tmp/goo-394-reviewed-goo NO_PROXY=127.0.0.1,localhost npx playwright test journeys/394-distribution-customize.spec.js --workers=1
GOO_GENERIC_BINARY=/tmp/goo-394-reviewed-goo NO_PROXY=127.0.0.1,localhost npx playwright test journeys/394-distribution-customize-actor.spec.js --workers=1
```

The companion uses the main run's fresh ordinary OTLP bank and writes a separate
`actor.json`. It creates a uniquely named runtime/campaign directory. Run both
with loopback access in the cloud outer execution environment.

From the `verification` Go module, with Node 24 and Quint on `PATH`:

```sh
XDG_CACHE_HOME=/workspace/.cache go test . -run '^TestHegelDistributionCustomization394$' -count=1 -v
GOO_DISTRIBUTION_CUSTOMIZE_394_REPORT=/workspace/goo/artifacts/distribution-customize394/summary.json \
  GOO_DISTRIBUTION_CUSTOMIZE_394_ACTOR=/workspace/goo/artifacts/distribution-customize394/actor.json \
  go test . -run '^TestQuintDistributionCustomizationObserved394$' -count=1 -v
```

The generated check passes 32 shared JavaScript/Go handoffs and store round trips.
It covers exact query preservation and excluded profiles. Observed admission and
required Quint replay pass in 4.64s (package 4.651s), zero skips. Six semantic
mutants and an unauthorized staging-read control are rejected.

Actual acceptance:

- Eighteen raw logs and eighteen raw spans admit source/child counts of eight and
  destination counts of ten, with the full typed questions and nondefault limit.
- The main journey passes in 20.0s (runner 21.7s): 35 measured boundaries, fourteen
  source requests, four configuration-object reads, two browser configuration
  publications and fifty SQL statements. Two destination definition publications
  are explicit setup, separate from browser counters.
- Real hidden, inert, pending and superseded producer callbacks cannot stage a
  continuation. Actual empty, offset and failed reads expose no usable action.
  Changed-read refusal remains generated/helper coverage.
- The actor passes in 37.9s (runner 40.1s): 608 snapshots, eight phases, two actual
  request/response pairs and eight SQL statements after its source-frame prelude.
  The prelude is one source request/four SQL statements. Main and actor budgets
  remain separate; no resource-timing or actor pixel claim is used.

Four genuine Playwright frames and the user journey are in the
[storyboard](../journeys/394-distribution-customize-storyboard.md). The initial
continuation excludes numeric, summary/status, reported-edge, specialized pattern,
histogram, grouped-time and nonzero-offset queries. Complete captured groups do
not imply complete upstream ingestion.
