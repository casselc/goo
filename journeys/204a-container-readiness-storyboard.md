# 204a — container authority and native readiness

Status: API prerequisite. The operator-facing discovery/UI journey is planned;
existing 203c screenshots continue to show the unchanged running product.

An operator with an exact reported container resource wants to preserve its
reported readiness during a fixed window, without confusing a replacement
container, a missing sample or pod health with that observation.

| Step | Implemented behavior | Executable check |
| --- | --- | --- |
| Author a pinned container readiness definition | Actual reported ID plus pod UID and container-name context, full exact stream, per-stream grain; no service fabrication | `TestContainerReadinessCompileReplayAndAuthoring` |
| Execute through the existing measurement endpoint | Exact gauge read returns state code 0 as observed Not ready | `TestContainerReadinessHTTPExecuteCaptureAndOfflineReopen` |
| Preserve code 1, unsupported values and observation quality | Ready/Not ready remain separate from stale/missing/conflicting/unsupported data | `TestContainerReadinessDomainAndQuality` |
| Observe a replacement container ID | Same pod/name cannot make a new ID match the original question | Hegel authority/quality property and raw refusal tests |
| Record an extremely small nonzero decimal | Rounding cannot manufacture Not ready | Precision rejecting cases and Hegel property |
| Capture then reopen with a fresh offline handler | No telemetry reread; complete stored bytes unchanged | Same HTTP execute/capture/reopen test |
| Attempt comparison or change the saved label to Healthy pod | Categorical arithmetic is unsupported; modified evidence fails replay | Handler test and Hegel mutation |

These use the real compiler, evaluator, HTTP handlers and temporary filesystem
custody with a controlled dependency transport. They are not ordinary OTLP,
a running receiver, browser integration or an actual ClickHouse outage test.
The previous accepted pod-phase raw/custody oracle is replayed from its existing
session ledger to check compatibility after sharing the private state kernel;
no browser campaign or source read is repeated.

Next: discover reported container variants, expose exact readiness candidates
through native UI/capture views, and exercise generated ordinary OTLP through
Collector/ClickHouse with real browser, ownership, Bombadil and independent
raw/custody checks before deploying the combined feature.
