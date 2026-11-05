# 330 — Find text inside kept observations

The executable storyboard is `journeys/330-kept-content-search.spec.js`. It sends ordinary OTLP logs and spans and keeps one of each through Activity's real selection and Keep controls. Search starts only after the user chooses **Find kept observations** with **Record text** filled.

1. Search `Literal330`. Both kept selections appear with matching-record counts and bounded witnesses naming the original record and field. The witness text is rendered as text, including HTML-looking content.
2. Search `%_`. These characters are literal text and find only the log Body. Lowercase `literal330`, an absent phrase, and an arbitrary resource attribute value find nothing. ServiceName, TraceId and SpanId searches verify their corresponding witness fields. A single space remains a real search value.
3. Select Spans and find again. Only the span keep appears. Open the result to verify the exact original receipt and inspect its original record.
4. Change text while a genuine server response is held in transit. The earlier response cannot repopulate the changed search. Editing alone does not read custody.

After the two Keeps, browser request accounting requires zero telemetry reads and zero writes during discovery and Open. Each exact receipt is unchanged afterward. Content results expose counts and safe bounded witnesses; Open remains the separate original-record verification action.

The optional `GOO_KEPT_CONTENT_BOMBADIL=1` campaign runs for 30 seconds over the genuine UI-created keeps. It must progress through positive search, input invalidation, negative search, signal filtering, exact Open and inspection. Invariants prohibit source rereads, writes, implicit discovery, and more than three explicit searches. A completion snapshot is required, so an idle actor cannot pass.

Validated on 2026-11-05: the native journey and configured 30-second actor passed (42.5 seconds; wrapper 44.4 seconds). The actor recorded 563 state snapshots across ten phases, three verified actual search responses and zero source reads or writes. Exact Open currently causes two reads of the same ID through the existing navigation lifecycle; both responses are compared with the unchanged original receipt. Milestone 331 removes this duplicate through route ownership; the maintained regression and actor now require one custody GET per Open intent. The figures in this paragraph describe the historical 330 run.

Generated frames: `screenshots/330-kept-content-witnesses.png`, `screenshots/330-kept-span-search.png`, and `screenshots/330-kept-content-narrow.png`. The 390-pixel capture verifies stacked results and safe witness wrapping; a 256-byte input also stays within the viewport. Its targeted capture copied only the successful run's two genuine UI keeps into an isolated runtime. Copying changes file timestamps, so an explicit setup-only metadata repair was necessary before the read-only browser capture. Receipt contents were unchanged. The maintained journey captures a successful narrow result before testing long-input invalidation.

Compact evidence lives under `artifacts/kept-content330/`: `summary.json`, `campaign-summary.json`, and `narrow/summary.json`. Initial attempts exposed environment PATH selection (GNU Go instead of the configured Go toolchain), then a harness assumption of one Open read instead of the existing two; these were corrected without timeout changes. Large continuation cases are covered by Go property verification rather than bulk browser fixtures.

Root independently reviewed the main and actor raw evidence and the narrow summary; design accepted all three frames. Owned raw browser/actor traces, copied capture stores and runtime diagnostics were then removed. Compact summaries and maintained frames remain. Docker volumes and existing stream entries were preserved, and the wrapper restored demo traffic.

Milestone 331 regression validation on 2026-11-05: the updated native journey and configured 30-second actor passed in 40.0 seconds; the actor recorded 548 snapshots across ten phases, three verified searches, exactly one custody GET, and no source reads or writes. All three maintained search frames were refreshed from that run.
