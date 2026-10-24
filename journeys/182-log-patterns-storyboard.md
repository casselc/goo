# 182 — Investigate reported log patterns

An operator wants to find common or unusual message shapes in the current log scope. **Log patterns** counts the full matching population, retains the raw messages, and lets the operator explicitly investigate one shape or compare its service distribution with the rest. Frequency is descriptive; no historical novelty, anomaly, health or causal claim is inferred.

## Journey

1. **Freeze the applied scope.** Start the saved Activity view for namespace `commerce`, either `checkout-api` or `queue-worker`, and team `core`. Choose **Log patterns** from Investigate. Activity pauses and stages its completed fixed window without an aggregate read.
2. **Read complete counts.** Run the analysis. The actual window contains 382 logs in 27 full-population patterns, while the inspection page contains 100 records. The time chart shows overall matching log volume; it does not claim a separate trend for every pattern row. Local most/least sorting does not read telemetry.
3. **Inspect exact meaning.** Pattern identity uses immutable `body_pattern_v1`: only maximal ASCII digit runs are replaced. Signs, decimal punctuation, IP separators, UUID letters, Unicode and control bytes remain literal. Empty, oversized and invalid UTF-8 bodies are explicit unclassified states. Inspect a recorded message containing HTML-like text: its original body is unchanged, and the label executes no markup.
4. **Follow one shape.** The most frequent observed shape, `request id=⟪digits⟫ completed in ⟪digits⟫ms`, has 128 records. **Open pattern logs** stages an exact-ID filter in the original window; Run explicitly reads its trend. Back restores the parent without a read.
5. **Compare services with the rest.** Stage the selected pattern versus its whole-parent remainder, then Run. Exact selected/rest predicates retain all original filters and partition all 382 records. Unclassified bodies remain in the 254-record remainder.
6. **Preserve and reuse.** Capture the observed parent and reopen its immutable evidence. Save the same explicit pattern column through existing YAML question facilities; Use performs a named fresh read. Return to paused Activity restores the Investigate opener and the original selector.
7. **Reject misleading or obsolete results.** A separate actual population has 1,100 distinct alphabetic patterns. Run refuses the group budget and publishes no top list. A retained native pattern link cannot stage a question or read telemetry while its owner is hidden or its scope is dirty.

The actual service comparison is useful because both services are present in the allowed scope:

| Service | Selected shape | Rest of parent |
| --- | ---: | ---: |
| checkout-api | 116 | 12 |
| queue-worker | 12 | 242 |
| Total | 128 | 254 |

These are observed counts in one fixed scope, not an explanation of the difference.

## Actual frames

| Frame | What it establishes |
| --- | --- |
| [Full-population patterns](../docs/journeys/screenshots/182a-full-population-log-patterns.png) | Actual complete counts, shares including unclassified bodies, escaped labels and descriptive local sorting |
| [Saved patterns on mobile](../docs/journeys/screenshots/182b-saved-log-patterns-mobile.png) | The same immutable evidence, retained labels and full counts without page overflow |
| [Cardinality refusal](../docs/journeys/screenshots/182c-pattern-cardinality-refusal.png) | Actual 1,100-pattern source population is refused without partial grouping totals |

## Acceptance evidence

The first frozen browser invocation passes all three cases in 41.6s: main 28.9s, stale-link ownership 4.5s and cardinality refusal 2.4s. Main execution ends at seven logical source operations, one comparison protocol, one publication and one separately tracked Activity SSE stream. These are application-level operations, not counts of the underlying ClickHouse queries. Sorting, inspection, staging, Back, saved reopening and YAML publication add no source operation.

- Main: `artifacts/log-patterns/log-patterns-7e40a8e9f4/ledger.json`. The fixture emits 400 matching records and 30 excluded namespace/service/team controls. Its observed fixed window contains 382 matches; raw facts independently determine every group and time interval.
- Ownership: `artifacts/log-patterns/log-patterns-7e3f5ef9fc/ownership.json`. The obsolete link remains connected while its workspace is hidden. Programmatic bubbling clicks change neither the URL nor staged-question count and add no read. A dirty owner also rejects its retained link.
- Cardinality: `artifacts/log-patterns/log-patterns-bccde52ccd/cardinality.json`. An actual ready probe counts 1,100 matching logs and the independent byte scanner derives 1,100 identities. The grouped read returns the explicit budget refusal, with no group payload or visible pattern rows.
- Bombadil: `artifacts/log-patterns/log-patterns-cf4a74ceb7/bombadil/acceptance-summary.json`. The 55-second campaign passes all 13 phases with 907 state snapshots, four logical source operations and one comparison protocol. It exercises sorting, raw inspection, exact trend staging/Run/Back and the applied selected/rest pair.

The main has adjacent `scope-capture.json` and `capture-custody.json`: 222,507 actual stored bytes, SHA-256 `808ea28381d0d788991e3f5c820d970d100c76c7c39f380bae5d57af11c1033f`. Public 100-row pages alone do not establish complete retained-record or byte custody. Independent native/model verification is a separate gate; its final receipts belong in VALIDATION.md and verification/README.md.

The corpus covers digit widths, signs and decimals, IP/UUID punctuation, full-width non-ASCII digits, distinct composed/decomposed Unicode, valid U+FFFD, escaped HTML, NUL/newline/tab text, empty bodies, exactly 256-byte bodies and oversized bodies. OTLP JSON cannot carry malformed physical UTF-8 bytes: `lp1:u` is covered by the native/backend byte corpus, not a fabricated live response. Oversized bodies remain intact in raw evidence; the bound applies to classification.

## Run

```sh
npx playwright test journeys/log-patterns.spec.js --grep '182 (real|stale|more)'
GOO_LOG_PATTERNS_BOMBADIL=1 npx playwright test journeys/log-patterns.spec.js --grep '182 Bombadil'
```

These journeys require the deployed lab and ordinary OTel ingestion. Run browser campaigns serially and preserve diagnostic traces outside the disposable Playwright output before another invocation.
