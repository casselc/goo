# 183 — Compare log patterns across completed reads

The operator has a full-population pattern question and wants to compare its observed shapes with a preceding window or a saved scope. Counts, duration-normalized rates and population shares answer different questions. The [comparison contract](../docs/LOG_PATTERN_TIME_COMPARISONS.md) keeps their arithmetic and evidence scopes explicit.

## Journey

1. **Read the current scope.** Open the fixed 60-second pattern question and Run. Its 400 matching logs retain the namespace, two-service and team filters. Choose the previous adjacent window. Staging is source-free; Run independently reads both complete scopes.
2. **Compare the aligned union.** The previous/current pair contains six patterns, including empty and oversized unclassified bodies. Switching count, rate and share is local. Current-only and baseline-only shapes say “Not observed” in the other completed read, without claiming historical novelty or resolution.
3. **Inspect one side.** Open the baseline-only shape. Its exact identity, original baseline window and all predicates enter a staged question. Run reads 160 matching records. Back restores the comparison without a read.
4. **Load saved scope, then reread it.** The original saved 120-second capture groups by service and contains 400 logs. Twelve ordinary late records are confirmed visible in that historical scope before loading its metadata. Loading stages no source read. Comparison Run measures 412 baseline logs and 400 current logs, using the versioned pattern grouping and each side's original window and predicates. The original source capture remains byte-identical.
5. **Interpret unequal durations.** The common shape has 320 baseline records over 120 seconds and 160 current records over 60 seconds. Both rates are exactly `8/3 logs/s`: the actual rate bars are equally 246px wide. In count mode, the same data produces 246px and 123px bars under a shared 320-log scale. Shares still differ because each denominator is the whole population.
6. **Preserve evidence and intent.** Save the configured comparison view without a read. **Capture comparison (fresh reads)** explicitly rereads both scopes into a new immutable comparison; its underlying source receipts have later read clocks. Reloading that capture reads storage only. Reopening the configured view stages its queries; Run rereads them.
7. **Distinguish zero from unavailable.** A successful empty baseline has zero log rate and undefined share. Share ranking says differences are unavailable with stable pattern ordering. An actual baseline response whose browser delivery fails supplies neither a zero nor an absence label. A held response delivered after navigation cannot publish an obsolete comparison; explicit recovery Run succeeds. A stale side link cannot navigate or stage while its owner is hidden.

## Ordinary OTel fixture

Both allowed services are present, with additional excluded namespace/service/team records. The immutable `body_pattern_v1` profile is unchanged.

| Category | Previous 60s | Current 60s | Saved baseline 120s |
| --- | ---: | ---: | ---: |
| Common shape | 160 | 160 | 320 |
| Rising shape | 40 | 160 | 40 |
| Current-only shape | 0 | 40 | 0 |
| Baseline-only shape | 160 | 0 | 20 |
| Empty body, unclassified | 20 | 20 | 10 |
| Oversized body, unclassified | 20 | 20 | 10 |
| Total before late arrival | 400 | 400 | 400 |

Twelve late baseline-only records raise the saved source scope to 412. An actual ready probe confirms visibility before comparison Run. Independent BigInt/rational oracles derive every union member, count, rate, share, difference and bounded observation label from the raw emitted facts and exact half-open windows.

## Actual frames

| Frame | Evidence |
| --- | --- |
| [Unequal-window rate comparison](../docs/journeys/screenshots/183a-unequal-window-pattern-rates.png) | Actual 412/400 source populations with visible 120s/60s durations and a common shared scale |
| [Frozen comparison on mobile](../docs/journeys/screenshots/183b-mobile-pattern-time-comparison.png) | Source-free reopening of the fresh comparison capture, with title, measure chooser, paired values and complete rows unobscured |
| [Empty-baseline shares](../docs/journeys/screenshots/183c-empty-baseline-share-unavailable.png) | A real successful zero baseline, undefined shares and the stable-order explanation |

The mobile frame shows the freshly captured pair reopened from storage. Its counts and fractions match the earlier live comparison; its capture/read clocks remain distinct. It uses the real mobile-width DOM and a tall viewport so the entire component fits clear of sticky content. No styles or pixels are edited.

## Acceptance evidence

The first frozen application run passes all three cases in one minute: main 35.2s, empty baseline 12.6s and held-delivery ownership 12.1s.

- Main: `artifacts/log-pattern-time/log-pattern-time-d9bcc1b3fe/ledger.json`. Full model JSON and exact DOM attributes are retained for all three measures. The main ends at nine logical source operations, three comparison protocols and one configured-view publication.
- Empty: `artifacts/log-pattern-time/log-pattern-time-568c93a96c/empty.json`. The baseline is an actual empty gap between populated fixed windows. Rates are known zero, while its shares and share differences are unavailable.
- Ownership/failure: `artifacts/log-pattern-time/log-pattern-time-519626dc3b/ownership.json`. Actual upstream responses are retained before a delivery abort or delayed fulfillment. No positive telemetry response is fabricated. The failed baseline publishes no zero/absence conclusion; hidden responses and stale links publish no obsolete question.
- Bombadil: `artifacts/log-pattern-time/log-pattern-time-3bedfcfee4/bombadil/acceptance-summary.json`. The 55-second campaign passes 13 phases and 860 state snapshots, with four logical source operations and one comparison protocol. Its terminal state includes both bounded absence labels, exact common rates, shared scales and the 160-record baseline-side inspection.

Logical source operations count application requests. The comparison capture is one explicit API operation that rereads two source scopes; these counts do not describe the number of underlying ClickHouse queries.

The original 400-record source capture has adjacent `scope-capture.json` and `scope-capture.after.json` with identical 191,983 bytes and SHA-256 `0786c7f3ec619ec37aa87e0488bdbab4e0db854f8ba4b6058d4c4d60860ce826`. The before copy precedes late ingestion. Public pages alone do not establish full stored-byte custody. Independent native/model acceptance is a separate gate, recorded in VALIDATION.md and verification/README.md.

## Reproduce

```sh
npx playwright test journeys/log-pattern-time-comparisons.spec.js --grep '183 (real|successful)'
GOO_PATTERN_TIME_BOMBADIL=1 npx playwright test journeys/log-pattern-time-comparisons.spec.js --grep '183 Bombadil'
```

The checked-in visual replay accepts explicit retained ledgers and explicit modes:

```sh
node journeys/183-visual-replay.js --ledger artifacts/log-pattern-time/log-pattern-time-d9bcc1b3fe/ledger.json --mode mobile
node journeys/183-visual-replay.js --ledger artifacts/log-pattern-time/log-pattern-time-568c93a96c/empty.json --mode empty
```

`mobile` performs zero source operations and verifies frozen rows, paired scales, unobscured heading/chooser, full-component fit and original custody. `empty` performs exactly three Explore operations, validates the successful zero/undefined-share case, and records actual entry-control geometry. `both` accepts the main ledger plus `--empty-ledger <empty.json>` and runs the modes sequentially. `GOO_QUERY_URL`, `CHROMIUM_PATH` and `GOO_QUERY_CONTAINER` select the existing lab endpoints and tools.

The completed replay log is `artifacts/log-pattern-time-visual-replay-complete.log`, with adjacent `mobile-screenshot-replay.json` and `empty-screenshot-replay.json`. Two preceding source-free framing attempts failed the strict whole-component fit guard because an unconditional offset exceeded document-bottom scroll space; their logs are retained separately. The corrected replay offsets only an actual sticky overlap. It keeps every fit assertion and performs no application/style mutation.

The accepted main used the earlier custody helper. The checked-in future helper and executed replay now explicitly select `unix:///var/run/docker.sock`, clear Docker host/context/TLS/certificate selectors, and preserve Docker configuration. The hardened replay independently verifies the same actual stored bytes; application query behavior and accepted arithmetic are unchanged.
