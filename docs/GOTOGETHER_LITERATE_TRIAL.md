# Gotogether literate acquisition contract

Goo's isolated verification module now pins gotogether
`d4c558eec9b71d4f3b544a8fa519ffb317fdeb75`. The acquisition binding names
[the authored Markdown contract](../verification/gotogether/model/acquisition.md).
Upstream generates its sibling `.qnt`, including document hash and line map.
The model body is unchanged; Goo has no private tangler.

## Workflow

From `verification/gotogether`, with Go 1.27.3 and Quint 0.33.1 available:

```sh
go tool gotogether tangle model/acquisition.md
go tool gotogether tangle -check model/acquisition.md
```

The CLI is a pinned Go tool in this verification module, so installation of an
unrelated global `gotogether` is unnecessary. From the repository root:

```sh
npm run test:gotogether-controls
GOO_GENERIC_BINARY=/absolute/path/to/goo npm run test:gotogether-shadow
```

CI checks generated drift **before** loading the binding: loading intentionally
regenerates the derived file. Edit Markdown, regenerate, and commit both files.
Even prose changes update the document hash and source map. Do not edit `.qnt`
by hand. Normal tangling refuses a hand-written sibling (exit 2); `-check`
refuses missing or stale output (exit 1).

The [frozen CW1 handoff](../verification/gotogether/testdata/cw1/contract.md)
remains byte-for-byte unchanged. Its captured model body and replay basis are
checked against the new authoring location; future behavioral changes require
reviewing that basis rather than silently rewriting capture provenance.

## Acceptance

- The original acquisition log and named scenario retain source/compiled verdict
  parity, including pending calls, missing final display, stale bank, hidden
  display, lost tail and stale compiled-artifact controls.
- A deliberately broken invariant reports `acquisition.md:92`. An unknown model
  symbol maps the frontend error to the same authored line. The test computes
  the expected line from the declaration, so harmless prose moves remain valid.
- The upstream CLI checks fresh/stale/missing output and preserves a hand-written
  sibling when refusing to overwrite it.
- The live OTLP/ClickHouse journey still runs its independent native gate before
  the shadow: 24 source SQL assertions, nine separate actor SQL calls, runtime
  and owned-database cleanup. Static fixture replay does not rerun that gate.
- Compiled positive replay runs without Node/Quint on PATH. The explicit drain
  bridge remains; EOF, browser closure and telemetry completeness are distinct.

Independent review confirmed unchanged model semantics, isolated dependencies,
intact frozen provenance and no weakening of the native gate. No product UI or
source-read implementation changed. The existing acquisition frames were
refreshed by the live journey.

## Other recent upstream work

| Delivered in this pin | Use for Goo |
| --- | --- |
| Structured counterexamples (#224) | Prefer typed steps, picks, states and replay when adding failure presentation; avoid parsing detail prose. No new viewer is introduced here. |
| Literate models (#225) | Adopted for acquisition: one authored model beside its explanation, source-mapped diagnostics and drift checks. |
| Graph memory accounting (#226) | Useful for future bounded graph jobs and investigating memory pressure. Periodic Go-runtime mapped-memory checks are not a hard cgroup/RSS limit; truncated exploration is not proof. |
| Package-local Hegel layout guidance (#227) | A useful pattern when testing private code, but no reason to move Goo's isolated verification dependencies into product modules. |
| Document design matrices and per-variant bounds/early stopping (#228–229) | Consider for genuine alternatives in paired-definition design; do not manufacture variants merely to repackage this fixed captured scenario. |

Request-log settlement events, generated typed alignment, causal
request/effect/adoption/display links and operational contract monitoring remain
planned. These additions do not replace independent DOM observations or the
physical-source oracle. The next product breadth task remains reusable paired
Activity definitions, with paused recall and explicit acquisition.

[Upstream issue #220](https://github.com/chucklehead-dev/gotogether/issues/220)
tracks the handoff and consumer feedback. Successful local adoption does not by
itself establish that every upstream CI job has completed.
