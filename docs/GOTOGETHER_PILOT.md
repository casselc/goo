# Gotogether outside-agent pilot

The approved onboarding and shadow-validation pilot ran in
`/tmp/goo-gotogether-pilot`, pinned to
[`50157b805ae455de165f37bf39da78690958763c`](https://github.com/chucklehead-dev/gotogether/tree/50157b805ae455de165f37bf39da78690958763c).
Its remote module version is `v0.0.0-20261026025747-50157b805ae4`.
Goo's production and verification dependencies, skills and current checkers were
not changed. This report records a consumer experiment, not adoption or migration.

## Onboarding results

Fresh remote CLI/core installation succeeded without local replacement of the
new library. `gotogether skills install` installed 28 matching skill/reference
files into each scratch `.agents/skills` and `.claude/skills` target. The agent
followed installed `gotogether-start`, workflow and traces skills.

Go 1.27.2 and existing Quint 0.32.0 worked. TypeScript was selected explicitly;
Rust and qweave were not installed. For the published hello example, only scratch
module files changed: local development replacements were removed and both core
and optional tutorial qhegel were pinned to the same commit. Tutorial source,
tests, model and binding were unchanged. This did not upgrade Goo's Hegel.

| Tutorial check | Actual result |
| --- | --- |
| Faithful faulted runs | Pass, 5.26s |
| Deliberate no-deduplication mutant | Pass: 23/30 faulted runs rejected; nominal runs conform, 37.10s |
| Bounded exhaustion | Pass, 21.05s |
| Hegel shrinking | Pass, shrinks to one fault, 36.23s |
| Optional woven mutation test | Explicit documented skip; qweave not installed |

The package completed in 99.675s. There were no missing-Quint or unexpected skips.
Independent CLI replay accepted a faithful run (exit 0) and rejected a mutant run
(exit 1). Module/build caches were reused; this is not a cold-install or speed
benchmark. Two initial working-directory/replay-path mistakes were pilot harness
errors, not library defects.

[Pin/skill/status manifest](evidence/gotogether-pilot/onboarding-manifest.json)
and [dependency-only tutorial patch](evidence/gotogether-pilot/hello-dependency.patch.txt)
retain compact proof; full command logs and actual exploration data remain in
scratch. The default-target skill install was tested, not `-target all`.

## Shadowing one actual Goo capture workflow

A separate scratch module imports the existing sibling quintgo and the remotely
pinned gotogether core together. It copies the existing `captures.qnt` and binding
unchanged and uses the original `TestQuintCaptureLiveLedger` observations against
an archived actual ClickHouse/browser ledger, not a new mocked positive UI.

Original ledger: `artifacts/captures/capture-968a089abf/ledger.json`, 455,612 bytes,
SHA256 `d7b47942244035a4d3e8be51c5e990a41b709c0e7432a4cb347d7e8798cd4307`.
Its bytes remain unchanged. Both checkers read the same serialized native step
log and use Quint 0.32/TypeScript; gotogether explicitly uses `Eval: "quint"`.
This isolates API/wire/model compatibility, not a benchmark of its default Go
evaluator. No live source, capture refresh or custody write occurs.

| Case | Existing checker | Gotogether |
| --- | --- | --- |
| Actual accepted save/page/refresh/reopen journey | Accept | Accept |
| Changed reopened count | Reject | Reject |
| Changed page basis | Reject | Reject |
| Changed refresh scope | Reject | Reject |
| Incorrect cross-revision status | Reject | Reject |
| Missing middle step | Reject | Reject |
| Disabled first action | Reject | Reject |
| Well-formed complete prefix | Accept | Accept |

All eight agree on conformance, failure kind and offending transition. The
comparison passed in 42.459s with zero skips, with per-check time/artifact sizes
recorded in [the shadow summary](evidence/gotogether-pilot/shadow-summary.json).
Single-run timings are observations, not performance guarantees.

The prefix result is deliberate: this safety-only model permits valid finite
prefixes. It does not prove that the whole save/page/refresh/reopen workflow
finished. Independent receipt checks and required workflow steps remain necessary.
For a physically cut-off final JSON line, the old reader errors and gotogether
returns three retained steps plus `TornTail`. The pilot propagates that incomplete
status instead of validating the surviving prefix as a full success; see
[incomplete receipt](evidence/gotogether-pilot/incomplete.json).

[Actual scratch test source](evidence/gotogether-pilot/shadow_test.go.txt) retains
the precise oracle and controls. Its module uses local replacements only for
Goo/the existing sibling checker, as Goo verification already does; the new
library comes from the pinned remote module. Reproduce with both repositories
available, the copied captures model/binding, and a separate scratch module:

```sh
QUINT_BIN=/absolute/path/to/quint \
GOO_SHADOW_LEDGER=/absolute/path/to/goo/artifacts/captures/capture-968a089abf/ledger.json \
go test -json -run '^TestShadowCaptureWorkflow$' -count=1 -timeout=180s
```

A missing Quint path fails the test rather than skipping.

## Reproduced library defect and issue draft

A separate adversarial parser check fails at the pinned commit: a newline-
terminated native log line containing one valid JSON object followed by another
object or trailing garbage returns one step and no error. `parseLine` decodes
once without checking EOF. This violates the documented one-object-per-line
contract and silently drops input. It is distinct from the supported torn-tail
case above. Source comparison confirms the same parser code in `2992868`, but
only the approved pinned commit was executed.

The [minimal regression](evidence/gotogether-pilot/malformed_test.go.txt) and
[actual failing output](evidence/gotogether-pilot/malformed-jsonl.txt) prove parser
data loss, not a separately demonstrated downstream false-green model verdict.
Independent review confirmed the finding. No library patch or weakened oracle
was used to make the check pass.

The user authorized repository feedback. GitHub API reads and issue creation
returned `Forbidden`; no issue was created. The complete, independently reviewed
[ready-to-file issue](evidence/gotogether-pilot/issue-jsonl-trailing-data.md)
contains versions, reproduction, expected/actual results and a bounded fix
suggestion. This is an environment/API-access limitation, not a gotogether defect.

## Adoption recommendation

The isolated compatibility pilot is encouraging. Keep the existing checkers
and independent raw-record/custody oracles authoritative. Resolve the malformed-
log reader behavior before routing external logs through it as trusted evidence.
A next bounded experiment could use model-judged shrinking around the real
lifecycle preparer and a deliberate mutant; that adoption is not part of this
pilot. Do not migrate all 105 import sites or change production's Go requirement
from this single workflow result.

Real filesystem and XTDB checks remain necessary: qenv's simulated disk does not
establish link/directory-fsync durability, nor does a simulated store establish
XTDB semantics. Browser exploration, actual OTLP/ClickHouse checks and screenshots
remain with the existing Playwright/Bombadil/source verification tiers.
