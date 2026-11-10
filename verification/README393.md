# 393: literal source parameters

The finite audit compares returned bytes and membership through public source
paths, not equality of encoder output. The existing record and metric encoders
remain unchanged. Native malformed-input validation now runs before JSON cloning
in both `Source.Run` and `Evaluate`.

Run these jobs serially against the existing collector and ClickHouse. The bank
tool emits a fresh ordinary OTLP cohort and leaves other lab data intact. From
the repository root:

```sh
NO_PROXY=127.0.0.1,localhost node journeys/support/literal-transport393-bank.js artifacts/literal-transport393/bank.json
```

The bank has 50 legal target literals, 50 near misses and one missing-key row per
signal. Its eight bounded readonly queries verified 101 logs and 101 gauges,
including original UTF-8 hex, IDs and metric identity. NUL and malformed UTF-8
are never seeded as legal values. The manifest format is
`goo.literalTransport393Bank/1`.

From the `verification` module:

```sh
GOO_SOURCE_PARAMETER_393_BANK=/workspace/goo/artifacts/literal-transport393/bank.json \
  XDG_CACHE_HOME=/workspace/.cache \
  go test -run '^(TestHegelSourceParameterIntent393|TestTypedSourceParameterRefusals393|TestActualSourceParameterRoundTrips393)$' -count=1 -v .
```

Run the actual audit with loopback access in the cloud outer execution
environment. A socket-denied attempt supplies no codec verdict.

`GOO_SOURCE_PARAMETER_393_ENDPOINT` overrides the source URL; otherwise the audit
uses `GOO_TEST_CLICKHOUSE_URL`, then the local lab endpoint. Configured
`GOO_TEST_CLICKHOUSE_USER` and `GOO_TEST_CLICKHOUSE_PASSWORD` override local lab
credentials. The actual audit skips if the bank variable is absent; a skipped
run is not round-trip evidence.

From the repository root, check native validation before cloning:

```sh
go test ./measurements -run '^TestEvaluateRejectsMalformedLiteralBeforeClone393$' -count=1 -v
```

Actual acceptance: 150 source arms, twelve changed-literal/decode-intent selections and 147
distinct server-decoded `String` echoes passed in 4.10s (package 4.123s), zero
skips. Direct record and metric attributes and exact-stream JSON are separate
arms. Changed-literal controls return distinct known IDs that the original
target-ID oracle rejects. The transport recorder checks actual SQL/typed parameters, readonly and
budget settings. The emitted `source-parameter393-audit.json` uses
`goo.sourceParameter393/1`. The combined post-fix run passes: generated intent 0.01s, typed refusals 0.00s,
actual round trips 4.10s; zero skips.

There is no speedup claim, UI change or new screenshot/model proof. See the
[scope and validation boundaries](../docs/LITERAL_PARAMETER_TRANSPORT.md).
