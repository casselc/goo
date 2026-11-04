# Loaded-match selection: measured recurring work (326)

This benchmark mounts the actual 325 `Tail.mount` in standalone Chromium. It
serves a frozen copy of UI assets, feeds deterministic synthetic SSE rows, pauses
the stream, and injects hidden selected references through the real `addMany`
method. Those references model selection retained after eviction; this is a
performance fixture, not a claim that the benchmark obtained server custody.
No product endpoint, Docker service or journey wrapper is involved.

Baseline Tail SHA-256:
`1057b765c0bbfffaadb3e896a1547a0357fdc7aafe2ecf471df106e864a20bed`.
Engine: Chromium 154.0.8037.57. Times below are **Chromium**, not Node timings.
The timer wrapper measures the complete actual one-second `refreshSelection`
callback. A `structuredClone` spy counts calls; a separate `planMany` wrapper
attributes its work. Setup and rendering are excluded.

| Paused fixture | Loaded bytes / rows | Hidden selected bytes / refs | Match state | Clones per idle tick | Three complete idle ticks (ms) |
|---|---:|---:|---|---:|---|
| Short success | 77,985 / 100 | 785 / 1 | 2 matches; union 3 | 106 | 13.1, 15.4, 8.3 |
| Large overflow | 1,068,385 / 100 | 213,761 / 20 | 1 match; union 21 refused | 124 | 17.5, 7.0, 11.8 |
| Same large fixture, no Find | 1,068,385 / 100 | 213,761 / 20 | Control hidden | 0 | 1.6, 1.3, 1.5 |

Each nonempty-Find idle tick invoked `planMany` once; the no-Find control invoked
it zero times. All fixtures remained within the actual 100-row/2-MiB tail bounds
and the 20-reference selection bound. The overflow reason named the actual
21-item union. There were no page exceptions or active-phase source reads/writes.

The enabled short-case action took 11.2 ms, made 219 clone calls and three
planner calls, and selected the expected union of three. The overflow and
no-Find controls were disabled and did no invocation work. Three timer samples
are enough to expose deterministic clone counts, not to establish a latency
percentile; the wall times vary with host scheduling.

The actual host repeatedly called `columnControl.layout()` inside the per-row
matcher, adding roughly 100 small layout clones per tick. The measured large
planner itself was much smaller than the whole callback (the initial separate
three-tick run measured 1.3, 3.7 and 1.9 ms). This supports taking one layout
snapshot per operation. It does **not** establish a need for advisory caching,
clock-based validity caching or skipped proof preflight.

Reproduce (installed Chromium and Playwright required):

```sh
node scripts/benchmark_batch_selection.mjs /tmp/batch-selection.json all
```

`CHROMIUM_PATH` can override `/usr/bin/chromium`. To reproduce the baseline with
the current measurement script, extract the complete prior UI into a temporary
directory and point `BENCH_UI_DIR` to it:

```sh
mkdir -p /tmp/goo-325-baseline
git archive 4cfffc7 ui | tar -x -C /tmp/goo-325-baseline
BENCH_UI_DIR=/tmp/goo-325-baseline/ui node scripts/benchmark_batch_selection.mjs /tmp/batch-before.json all
```

`BENCH_UI_DIR` selects an immutable UI copy for paired measurements. The script removes its temporary
server/assets on both success and startup failure. Raw baseline results for this
session are `/tmp/goo-batch-selection-baseline-all.json` and
`/tmp/goo-batch-selection-after-all.json`. The frozen UI copy was removed after
paired review; temporary server copies are removed by the script.

## Paired result after the bounded refactor

The refactor changes only per-operation layout snapshots and removes the unused
preflight plan from captured button state. Proof preflight still runs on every
applicable idle tick and again on activation. No validity cache was added.
After Tail SHA-256:
`eebac39154de4388a53da7ff3e265f1f39a2406f5dc98b89b23aa87818a1a5be`.
The engine, all fixture counts/bytes, visible match counts, hidden-selection
counts and enabled/refused outcomes matched exactly. No page exceptions,
active-phase source reads or writes occurred.

| Fixture | Before → after clone calls per tick | After three complete ticks (ms) | Planner calls per tick |
|---|---:|---|---:|
| Short success | 106 → 6 | 22.0, 62.0, 15.5 | 1 |
| Large overflow | 124 → 24 | 18.7, 24.3, 5.9 | 1 |
| Same large fixture, no Find | 0 → 0 | 7.3, 2.3, 1.3 | 0 |

The successful click dropped from 219 to 19 clone calls and still invoked the
planner three times and selected three observations. Its elapsed time was
17.7 ms after versus 11.2 ms before. Both disabled controls still did no
invocation work.

These results establish the deterministic reduction in redundant clone calls,
**not a latency improvement**: the after elapsed samples were noisier and often
slower. Three samples do not justify percentile, speedup or memory claims. The paired run followed the full regression checks.
