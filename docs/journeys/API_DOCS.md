# API contract discovery journey

Implemented documentation surface; no telemetry or publication is executed.
Run `node scripts/check_api_docs.cjs` after installing the project's npm tools and
Chromium. It builds and starts the checkout's embedded Go runtime from an empty directory
on an ephemeral loopback port, then closes it, leaving the running lab untouched.

1. Open `/api-docs`: see 88 Go operations and choose browser-prefixed versus native paths.
2. Search `/tail`: expand the operation and its contract, inspect SSE media type,
   effect, request limits and event schemas. No Execute button or network source read.
3. Search `/measurements`: see the source/custody effect and 65,536-byte body limits.
4. At 390 px width, search `/compare`: distinguish supplied-summary comparison from
   saved comparison without horizontal page overflow. Browser errors must be absent.

![Actual API docs showing tail contract](screenshots/api-docs-tail.png)

The screenshot comes from the executable journey, not a mock. API behavior is
verified separately by the [OpenAPI contract checks](../API.md#updating-and-checking)
and Go wire/SSE tests. This documentation journey performs read-only checks and never queries telemetry.
It also opens the typed origin-aware retained-content contract.

To exercise the rebuilt deployed server instead, run
`GOO_API_DOCS_URL=http://127.0.0.1:8088/api-docs node scripts/check_api_docs.cjs`.
The same filtering, schema/limit disclosure and desktop/mobile checks apply.
See [API acceptance](../API_ACCEPTANCE.md) for the separate real telemetry and
saved/retained-evidence journeys.
