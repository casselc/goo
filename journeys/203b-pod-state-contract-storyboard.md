# 203b — native pod phase contract

Status: API-level source/semantic/custody acceptance. Product discovery and phase
presentation are planned, not demonstrated here. Existing 202 screenshots remain
accurate for the unchanged memory UI.

An operator with an exact reported pod resource variant can author a native
`state` definition using `k8scluster-pod-phase/0.137.0`. Its resource/schema,
receiver identity, metric, unit and point attributes describe one full stream.
This is not a broad Kubernetes health question.

| Step | Implemented outcome | Executable check |
| --- | --- | --- |
| Submit the exact native question | Existing measurement endpoint reads the gauge table, preserving raw service-less resource identity | `TestPodPhaseHTTPExecuteCaptureAndOfflineReopen` |
| Inspect a freshly reported code 5 | Result is observed, label Unknown; no healthy/missing conversion | Same handler test and `TestPodPhaseDefinitionSourceAndFrozenReplay` |
| Capture the observation | One fresh source read creates an individual immutable capture | Same handler test |
| Reopen after replacing the handler with source unavailable | No new source read; full stored bytes and normalized measurement result unchanged | Same handler test |
| Attempt a generic comparison | Explicit unsupported result; no phase subtraction or numeric difference | Same handler test |
| Submit an unsupported profile | Rejected before source access | Same handler test |
| Alter a captured label to Healthy | Replay rejects it | `TestHegelPodStateNativeReplayRejectsLabelMutation` |

The handler uses a controlled dependency transport and the real routing,
compilation, evaluation, filesystem custody and reopen code. It is not ordinary
OTLP ingestion, a running k8scluster receiver or a browser phase journey.

Run from the repository root:

```sh
go test ./metrics ./measurements ./exploration ./cmd/goo
cd verification
go test -run TestHegelPodState -count=1 .
```

Hegel requires the established writable cache configuration; source transport
regressions need local-loopback access. API observation output is optional via
`GOO_POD_STATE_API_OBSERVATIONS`; validate it using
`scripts/check_openapi.py --pod-state-observations <file>` in the API-tools venv.

The next journey must use actual ordinary OTLP through Collector/ClickHouse,
then metadata discovery, native UI inspection, individual capture/Back and a
source-free frozen reopen. It needs actual desktop/mobile state frames, ownership
withdrawal checks and updated Bombadil actions/invariants before UI acceptance.
