# 405: shared ClickHouse parameter encoding

Run heavy checks serially with Go 1.27.3 and Node 24. Unit and HTTP transport
checks live in `internal/sourcehttp`, `exploration` and `measurements`.

From `verification/`, with a writable `XDG_CACHE_HOME` for Hegel:

```sh
go test -run 'Test(HegelParameterEncoding405|ParameterCanonicalDecoderControls405|HegelSourceParameterIntent393|TypedSourceParameterRefusals393)' -count=1 -v
```

For actual source acceptance, generate a fresh instance of the existing ordinary
OTel bank from the repository root. This ingests unique records; it resets no
telemetry or stores:

```sh
NO_PROXY=127.0.0.1,localhost node journeys/support/literal-transport393-bank.js artifacts/parameter-encoding405/bank.json
```

Then from `verification/`:

```sh
GOO_SOURCE_PARAMETER_393_BANK=../artifacts/parameter-encoding405/bank.json GOO_SOURCE_PARAMETER_405_BANK=../artifacts/parameter-encoding405/bank.json NO_PROXY=127.0.0.1,localhost go test -run 'TestActual(SourceParameterRoundTrips393|ParameterEncodingDifferential405)' -count=1 -v
```

The reused 393 audit admits fresh physical hex/identity evidence and actual public
record/metric parameter paths, including near-value and missing controls. The 405
differential compares frozen old metric encoding and shared encoding through real
ClickHouse String decoding. It reports changed wire bytes separately from wrong
semantic values; the old writer is test-only, never a compatibility fallback.
Artifacts are `artifacts/verification/source-parameter393-audit.json` and
`parameter-encoding405-differential.json`.

The pure helper accepts bytes; public Unicode, NUL and size admission remain
family rules. Canonical-decoder negative controls concern the helper's output
grammar, not every alternate spelling accepted by ClickHouse. No UI, API type or
ownership state changed, so this cut adds no browser/actor/model campaign. Existing
404 storyboards and frames remain current. [Validation](../VALIDATION.md) records
executed results and optional skipped checks.
