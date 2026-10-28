# 219 — Load a workspace's catalogs when they become useful

A fresh Activity entry should show observations without constructing the hidden
sample galleries or reading every optional definition catalog. First opening a
workspace loads its own declarations; returning keeps that local state. Catalog
loading is not a telemetry query.

| User journey | Executable expectation |
| --- | --- |
| Open fresh Activity | Native/record definition workbenches remain available, but Examples and legacy declaration catalogs have not loaded. The live tail remains the only automatic telemetry stream. |
| Open Definitions | Legacy package/prebuilt metadata and operation declarations load once. |
| Open Examples | Galleries/workflows mount on first use. One example-question response supplies independent snapshots to both consumers. |
| Choose a workflow, leave and return | The chosen workflow/window stays held; navigation does not duplicate catalog reads or perform telemetry queries. |
| Preview its telemetry step | The shared catalog actually supplies the definition; pure planning succeeds without reading telemetry. |
| Refresh native packages explicitly | A fresh catalog request still occurs. This is not a global response cache. |
| Enter Examples, Definitions or Operations by URL | Only that destination's deferred catalogs activate; no initial telemetry read is introduced. |

Run:

```sh
GOO_STARTUP_BOMBADIL=1 npx playwright test journeys/219-startup-catalogs.spec.js --reporter=line
```

The workspace Bombadil actor now visits Definitions and configured Examples before
ordinary navigation and rejects duplicate loads of these one-time catalogs. It
retains the existing navigation/saved/live-owner and uncaught-error invariants.

One local before/after recording at 1440×1000, after about 2.2 seconds, had:

| Observation | Before | After |
| --- | ---: | ---: |
| Startup API requests, including the live tail | 14 | 7 |
| Native modules loaded | 136 | 136 |
| DOM nodes with 90 supplied live rows in each run | 3,252 | 2,936 |

The removed startup work is now owned by a later explicit workspace visit, not
silently discarded. Separate retention-status consumers still issue two startup
requests. These local counts are not a latency, throughput or load benchmark;
there is no claim of a module-loading speedup, new bundler, universal cache or
complete hidden-work elimination. No new visual layout or screenshot is required.
The isolated generic-startup journey also remains valid without sample files.
