# 192 — From a native interval to coincident records

Status: live acceptance passes. Core/UI source `75e76ea` and retained-origin provenance `15b8b5c` are deployed; main, bounded cases, ownership and actor pass. The native entry frame is a separately recorded fresh read of the accepted fixed question.

The engineer selects one supplied counter or histogram export interval, stages a fixed log/span question, explicitly reads the coincident records, and returns to the retained measurement without rereading it. Shared time/resource identity is contextual evidence, not an exemplar, contributor, causal edge or service-health claim.

## Scope contract

The record question uses the interval's exact half-open nanosecond bounds. Its coarse storage window is floor(fromNS) to ceil(toNS) in milliseconds, with `timestamp_ns >= fromNS` and `timestamp_ns <= toNS-1` preserving exact membership. Only reported environment, service, resource namespace (missing/empty/value) and instance transfer. Metric attributes, metric name, instrumentation scope/version and schema are not record filters.

Existing trace expansion deliberately broadens the record selection: it reads the reported trace ID inside the coarse millisecond window and can include other instances/namespaces or observations just beyond the native nanosecond edges. The exact record question remains the trace origin; the context disclosure explains the broader trace window.

## Generated ordinary OTel fixture

The fixture is synthetic telemetry sent through the ordinary OTLP collector, without mocked positive source/UI results or a manifest. It contains counter/histogram streams for namespace value, empty and missing states, metric-only labels that collide with resource field names, an unknown flagged export, a gap and a later observed-empty interval.

There are 123 supplied spans and correlated logs. Exact primary interval membership is 115, beyond the first 100-row page. F−1ns, F, F+1ns, T−1ns, T and T+1ns exercise boundaries inside the same coarse envelopes. Other instance/namespace records remain outside the exact record selection. A chosen reported trace has four observations in the intentionally broader coarse window, including instance and nanosecond-edge controls; a fifth observation outside that coarse window remains excluded.

## Executed user journey

1. Open a reusable measurement question and Run explicitly. Select the native interval locally.
2. Stage Spans in interval without reading. Run returns the full 115-record population; inspect a reported trace through the existing fresh trace/log owner and verify its broader four-record context.
3. Back to measurement restores the original question, basis, read clock, interval selection, retained DOM and opener focus without a source read. Repeat with logs.
4. Exercise histogram/empty namespace, counter/missing namespace, unknown export and a known-empty native interval. Unknown rates remain unavailable even when coincident records exist.
5. Hold an actual record response, return to native context, then release it. It cannot publish into the canceled owner. Retained hidden/closed actions are inert; actual browser Back also retains native evidence.
6. Run a short Bombadil campaign through native Run → local selection → source-free stage → explicit records Run → retained Back, checking all phases and source effects.

## Reproduction

```sh
npx playwright test journeys/native-record-context.spec.js
GOO_NATIVE_CONTEXT_BOMBADIL=1 npx playwright test journeys/native-record-context-bombadil.spec.js
GOO_NATIVE_CONTEXT_LEDGER=/absolute/path/to/ledger.json npx playwright test journeys/native-record-context-display.spec.js
```

Run only against the frozen deployed implementation. Artifacts are retained under `artifacts/native-record-context/<environment>/`; native API responses include their actual raw JSON text, and before/after snapshots preserve native question/plan/basis/readAt text independently of record results.

## Actual frames

- [192c — Select a supplied native interval](../docs/journeys/screenshots/192c-selected-native-interval.png): original accepted question/telemetry, one separately labeled fresh native read. Shows counter interval geometry, UTC bounds, local Logs/Spans actions and the context caveat. The attribute viewer is scrollable; this frame does not claim to display every resource field.
- [192a — Run the exact span context](../docs/journeys/screenshots/192a-native-interval-record-context.png): 115 matching span records, exact timestamp predicates and a retained measurement return action. The source declaration and broader trace context are explicit.
- [192b — Inspect exact log context on mobile](../docs/journeys/screenshots/192b-mobile-exact-interval-logs.png): the same accepted main's 115-log interval question, not a new native read or a different window.

Frames 192a/b belong to the accepted main's same generated environment. Frame 192c rereads that exact retained native question; its fresh source clock is documented in `display.json`. All frames are actual running-app browser captures without image editing or mocked data.

## Acceptance receipts

- Main **9.2 seconds**, 11.6-second invocation: `artifacts/native-record-context/native-context-a57c48baf4/ledger.json`; log `main-second.log`. Five actual source POSTs: one native measurement, one span context, the existing two trace/log reads, and one log context. Local selection, both stages and both returns add no reads. Original native question/plan/basis/readAt text, interval selection and opener focus survive Back unchanged.
- Four bounded cases **14.7 seconds**, first pass: `native-context-1a693076d7/cases.json`, within the first 56.4-second three-test invocation. Eight source reads, one explicit native and one explicit record read per case. Results are histogram/empty namespace 1, counter/missing namespace 1, flagged unknown interval 115 with no supplied rate, and known-empty interval 0.
- Ownership **5.6 seconds**, first pass: `native-context-1480119d7a/ownership.json`. Two source reads plus one explicit cancellation-control POST. The historical observer counter is three because it counts `/api/queries/<id>/cancel`; that control is not a third telemetry read. A held real response is released only after return; the canceled record owner stays unpublished. Hidden/closed anchors and actual browser Back cannot cause a fresh read or replace native evidence.
- Bombadil **25.8 seconds**, 28.2-second invocation: `native-context-9e531b8bdf/bombadil/acceptance-summary.json`; log `actor-first.log`. All eight phases, 418 snapshots, exactly two explicit reads and a source-free focused return are verified.
- Native-frame replay **3.6 seconds**, 5.6-second invocation: accepted main directory `display.json`; log `display-first.log`. It accepts an explicit retained ledger, performs no ingestion, compares the full native counter payload with the archived actual result, makes one explicit native read and verifies both contextual actions without further source reads.

The setup fixture ingestion/readiness requests are recorded separately from browser effects. No captures are created or republished in this slice. Actual native response raw JSON, immutable question, basis and read clock are retained in the ledgers; the measurement parent is retained in memory rather than reconstructed from a saved-capture format.

## Diagnostic and deployment

The initial main stopped after 27.0 seconds because its harness asked for the obsolete accessible button name `Expand`; the existing owner exposes `Expand trace <id>`. The 115-record scope and provenance assertions had already passed. `first-failure-diagnostics/` preserves trace, snapshot and screenshot; only the selector was corrected. The accepted cases and ownership tests were not repeated, and no application behavior or exact-boundary guard was weakened.

`artifacts/native-record-context/deployment.json` retains actual HTTP200/MIME/byte hashes for the index and three native modules, server hash and runtime binary SHA-256 `0d7ddbe24f646d668d4345efb32b6fa009f54755a31fe4eb5faa5de768bcc554`. The first static verification was interrupted by the explicitly approved provenance patch; the two changed assets were refreshed and all final hashes verified before browser execution. Only packages/query were restarted; existing lab arguments, ClickHouse and data volumes were preserved. The temporary build binary was removed.

The first adapter is live Measurements raw counter/histogram export intervals. Saved measurement displays, gauge points and fixed histogram time buckets do not acquire these actions. Existing fixed-editor density and inspector scroll layout are unchanged; this milestone does not claim to solve the wider UI backlog.
