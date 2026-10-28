# 217 — Leave a read without leaving work behind

An engineer inspects declared application operations, then moves to Services before
the response finishes. Leaving must stop that read; later rates/histogram requests
must not start in the hidden workspace. This uses existing ordinary sample
telemetry and real backend responses, delayed at the browser transport boundary.

| User journey | Executable expectation |
| --- | --- |
| Inspect operations, leave before the first response | Its query ID is cancelled; the held response cannot render or launch native follow-on reads. |
| Leave while rates or histogram evidence is pending | The active request is cancelled, existing displayed evidence stays unchanged, and no later pipeline stage starts. |
| Evaluate the prepared HTTP package, then leave Definitions | The independent HTTP-preview owner stops; late JSON cannot populate the hidden preview. |
| Change the lookback while reading | The old scope is cancelled; the new lookback is staged, with an explicit read required. |
| Hide the document, then show it again | A synthetic browser visibility transition cancels pending work; showing the page does not silently restart it. |
| Return and inspect again | A fresh query ID reads/render normally and preserves the existing package/revision guards. |

Run the [Playwright journey](217-legacy-query-ownership.spec.js):

```sh
GOO_LEGACY_OWNER_BOMBADIL=1 npx playwright test journeys/217-legacy-query-ownership.spec.js --reporter=line
```

The opt-in [Bombadil actor](../verification/bombadil/legacy-query-ownership.ts)
exercises source reads, workspace changes and scope edits without saving or
changing the workload. It requires an actual source read, records attempted GETs
at their invocation, and checks that they have query IDs and the correct visible
owner. Other invariants reject a reading status in a hidden owner and uncaught
errors. Its probe is test-only, not a product subscription or query registry.

Python proxy tests cover exact GET header forwarding and declaration exclusion.
Go admission tests cover cancellation/released capacity; transport tests include
JSON-body cancellation and an independent best-effort cancellation control.
Saved publication and SSE have separate lifetimes. An interrupted response is
not a rollback or proof that any publication failed. No new screenshot is needed
for this behavior-only owner change; the product layout is unchanged.
