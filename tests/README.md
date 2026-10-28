# Frontend checks

Use Node 24.19.0 (also pinned in CI) and `npm ci`, then:

```sh
npm run test:visualizations
```

The script enables Node's native module-mocking flag for fake-DOM owner fixtures.
`support/module-dependencies.cjs` maps those test fixtures to native imports and
keeps real fallbacks in an isolated module URL graph. These are test-only seams;
production modules publish no `Goo*` namespaces and require no test flags or loader.
For a single owner fixture, use
`node --experimental-test-module-mocks --test tests/host-dashboard.test.js`.
Pure module/helper tests can run with `node --test` directly.

Application import tests check that owners can load without a DOM or source read,
and that the HTML has one native bootstrap. Retention transport tests create
isolated clients with fake storage/HTTP boundaries; the application keeps its
normal default client. No runtime dependency injection is required by embedders.

Browser journeys require the running lab and Chromium. `npm run test:journeys`
uses the existing traffic/drain wrapper. Representative module checks are
`journeys/display-modules.spec.js`, `journeys/stream-entry.spec.js`, and
`journeys/activity-views.spec.js`. Actors use explicit native imports in their
async custom actions; any cached validation helpers live only on the actor's test
fixture DOM, not in the application. See [native properties](../verification/README.md)
and [Bombadil workflows](../docs/BOMBADIL_TESTING.md) for scoped campaigns.
