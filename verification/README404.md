# 404: ungrouped record/summary templates

Use Go 1.27.3, Node 24, Quint 0.32, a reviewed Go binary and the preserved lab.
Run heavy checks serially. From the repository root:

```sh
GOO_GENERIC_BINARY=/tmp/goo-404-reviewed-goo NO_PROXY=127.0.0.1,localhost npx playwright test journeys/404-summary-template.spec.js --workers=1
GOO_GENERIC_BINARY=/tmp/goo-404-reviewed-goo NO_PROXY=127.0.0.1,localhost npx playwright test journeys/404-summary-template-actor.spec.js --workers=1
```

The main ingests a unique raw bank and writes `artifacts/summary-template404/summary.json`.
The actor uses that same source bank, creates its own two configuration fixtures
as explicit setup, then inspects/applies and reads once. It writes `actor.json`.
Neither resets telemetry or stores. Local configuration steps remain separate
from telemetry reads and publication counts.

From `verification/`:

```sh
go test -run '^TestHegelSummaryTemplate404$' -count=1 -v
GOO_SUMMARY_TEMPLATE_404_REPORT=../artifacts/summary-template404/summary.json GOO_SUMMARY_TEMPLATE_404_ACTOR=../artifacts/summary-template404/actor.json go test -run 'Test(Quint|Bombadil)SummaryTemplateObserved404' -count=1 -v
```

Use a writable `XDG_CACHE_HOME` for Hegel's native runtime. The generated corpus
is `artifacts/verification/summary-template404-carriers.json`. From the root,
run `python tests/summary_template_404_api.py` in the API virtual environment.
Input/output carrier shape checks do not replace runtime field/profile/scope
admission. The observed verifier independently checks raw values, support, units,
source plans/binds, content identities and owner/read budgets. Its semantic
mutants and local-read model control must fail.

[Validation](../VALIDATION.md) records executed results. The
[storyboard](../journeys/404-summary-template-storyboard.md) links actual frames.
