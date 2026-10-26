# 200 · Reported Go runtime observations

Generated ordinary schema-1.37 OTLP fixture telemetry, not measurements emitted by
a running Go application. No positive source or UI responses were mocked.
`go.memory.used`, `go.memory.gc.goal` and optional `go.goroutine.count` are
cumulative, nonmonotonic stocks. GC goal is a reported heap target, not GC pause,
activity, pressure, memory capacity, or a utilization denominator.

The accepted functional source is `a98b9a9`; the separate label-only display
revision is `1357004`. [Deployment receipt](../artifacts/runtime-observations/deployment.json)
records runtime binary SHA256
`34cd5aa470ebbaa8f4bea570362f4b84452eb502f0ec41765858b7bc8684ed5c`,
actual asset HTTP/MIME/SHA checks, server and source/prebuilt copies. Only the
existing packages/query containers were restarted. Traffic stayed running.

## Actual journey

1. Discover `jobs / runtime-worker`, with 120 matching span records and ordinary
   native metadata. Default native selection is exactly two pinned full streams:
   one reported memory component and a GC goal from the same process identity.
   The accepted fixture selects worker-a's `go.memory.type=other`, **16,384 By**,
   and GC heap goal **262,144 By**. Other memory categories are not summed.
2. Inspect the native evidence disclosure, stage **Inspect exact runtime
   measurement**, and use **Back to service**. Both actions are source-free; the
   retained dashboard HTML, values and read clocks remain unchanged.
3. Stage that individual measurement again and explicitly Run. Capture reads that
   one exact source question afresh and stores an ordinary immutable measurement
   capture. Public reopen and later full UI reload use frozen evidence without a
   telemetry read. There is no paired runtime/GC capture or new store.
4. Return to the retained Services view after capture/public reopen, clear native
   selection and choose a single exact goroutine option. Editing reads nothing.
   Explicit **Apply and reread overview** performs one record aggregate read and
   one goroutine measurement read. Exactly one native widget remains, with **7
   goroutines** and raw unit `{goroutine}`; memory/GC defaults do not silently
   reappear. Full original capture bytes remain unchanged afterward.

## Actual frames

| Frame | Evidence |
| --- | --- |
| [200a desktop dashboard](../docs/journeys/screenshots/200a-runtime-observations.png) | Fresh display fixture's two independent default stocks after the label-only fix |
| [200b mobile dashboard](../docs/journeys/screenshots/200b-mobile-runtime-observations.png) | Same fresh display read; native values, instances, UTC clocks and plots contained |
| [200c workspace context](../docs/journeys/screenshots/200c-runtime-in-context.png) | Actual Services surface around the fresh display fixture's selected service |

The runtime cards have 90px plots, raw-value tooltips and UTC/ns timestamps. Their
visible important contents pass X/Y containment. The larger inherited dashboard
still scrolls on mobile and the generic operation table can pan internally; this
milestone does not claim to repair catalog density or all product layout.

## Actual receipts

- Independent observed validation **0.329s** (main 0.25s / owner 0.06s),
  **thirteen changed controls rejected**, with no observer failures:
  [native log](../artifacts/runtime-observations/native-validation.txt). Full raw
  identities, selected values/epochs, source effects and complete capture content
  ID/bytes are independently checked.
- Main **17.3s**, ownership **5.2s**, combined invocation **25.6s**:
  [main ledger](../artifacts/runtime-observations/runtime-observed-220f90ed2a/ledger.json),
  [owner ledger](../artifacts/runtime-observations/runtime-observed-99f75b3052/ownership.json),
  [log](../artifacts/runtime-observations/main-owner-second.log).
- Actor **32.0s / 38.6s invocation**, eight phases and 284 snapshots:
  [summary](../artifacts/runtime-observations/runtime-observed-b665fa0d36/bombadil/acceptance-summary.json).
  The separate ordinary fixture ends with baseline 8 / source 9 and completed
  native evidence restored to the retained service view. Its summary counts
  telemetry operations; detailed pure API effects are retained by the main
  observer, not a separate actor planner counter.
- Default overview is **8 logical source operations / 1 pure plan**. Native
  staging and retained Back stay at 8. Explicit native Run advances to 9 and
  individual capture to 10. Local goroutine selection stays at 10; Apply adds one
  aggregate and one native read, ending at **12 source operations / 4 pure
  plans**. These are application operations, not a count of physical SQL scans.
- Owner holds a real memory source response after `route.fetch`, leaves the
  service, then releases the actual body. No late runtime card, navigation or
  additional source operation may publish. Cancellation controls are recorded
  separately from telemetry reads.
- Individual capture ID
  `12678adfe6198c9c43e69c9001637f6c1077b049a01b0ee564326dfee3bf7f00`,
  **6,132 bytes**, SHA256 before/after
  `860494a978114b26c7e75e4c79be8169974ebcfa0e65aacf004ccb9cde8a8771`.
  Full actual stored files are adjacent to the main ledger.

```sh
npx playwright test journeys/200-runtime.spec.js journeys/200-runtime-ownership.spec.js --workers=1
GOO_RUNTIME_BOMBADIL=1 npx playwright test journeys/200-runtime-bombadil.spec.js --workers=1
```

## Diagnostic history and boundaries

The first main/owner invocation failed a local bookkeeping adapter: shared 187's
observer exposes numeric getters, but the 200 specs called `source()`/`plans()`.
The local support adapter now provides functions over those live getters. A
socket-free adapter transition check and syntax check passed before the corrected
invocation. Full initial frames/traces remain under
`artifacts/runtime-observations/first-failure-diagnostics`. No application source
or semantic assertions changed for that correction.

The raw fixture retains stack/other/custom/missing memory types, two instances,
resource variants, epoch changes and wrong schema/unit/kind/temporality controls.
Their actual metadata is distinct from the two selected default streams and the
explicit goroutine read; no claim is made that every metadata control received a
separate browser measurement. Exact namespace, resource, schema, scope/version and
metric dimensions are pinned in each selected question. Independent reads are not
an atomic snapshot, and rising values do not establish a leak or health state.

## Separate label-only display replay

The initial cards were readable, but an empty GC-goal metric-attribute list left
a trailing separator after its instance label. The reviewed `1357004` module-only
fix removes it and explicitly labels an unreported memory type. The missing-type
branch has focused component-unit coverage; no additional browser measurement
was invented for that branch.

The fresh fixture `runtime-observed-47d2f0174c` passed display acceptance
**8.6s / 11.1s invocation**:
[receipt](../artifacts/runtime-observations/runtime-observed-47d2f0174c/display.json),
[asset deployment](../artifacts/runtime-observations/display-deployment.json),
[frame hashes](../artifacts/runtime-observations/display-frame-manifest.json).
It performs the ordinary eight overview source operations plus one pure plan.
Viewport resizing, labels and screenshots add no reads. Both widths retain 90px
plots, exact raw-value/UTC/ns tooltips and contained visible native contents.
The actual final desktop/mobile PNGs were inspected. All three final frames are
from this new fixture/window/read clocks, separately from main/capture evidence.

The original capture was copied from its actual store again after display; its
6,132 bytes and SHA256 remain unchanged. The Go binary was unchanged; only the
reviewed module was copied, with actual HTTP/MIME/SHA checks and no restart.
No accepted main, owner or actor campaign was repeated for this label fix.

```sh
GOO_RUNTIME_DISPLAY=1 npx playwright test journeys/200-runtime-display.spec.js --workers=1
```

For a retained compatible main ledger, the display check also accepts
`GOO_RUNTIME_LEDGER=/absolute/path/to/ledger.json` for immutable-capture comparison.
Normal diagnostic retention preserves compact evidence and initial failures;
exact task build/deploy helpers are removed and traffic remains running.
