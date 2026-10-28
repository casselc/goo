# Independent verification and process review — 235

Reviewed 2026-10-28 against `5d12ce9`. Scope: default CI, optional verification,
browser/Bombadil integration, Hegel/Quint/MBT, API/storage contracts, setup,
review evidence and proposed P1–P4 acceptance. This is an independent source/process
review. I did not rerun broad suites, mutate runtime code, publish data or inspect
credentials. The only execution probe selected a deliberately nonexistent Go test
name; it ran no semantic tests.

## Main conclusion

The project has substantial useful verification, particularly independent source
oracles, exact identity/storage fixtures, negative model controls and recent
source/focus/viewport journeys. The refreshed product queue is sensible, but two
small verification/reliability cuts should precede or accompany P1: require a real
verdict from browser-to-Go ledger validation, and add the incomplete incoming-body
cancellation failure class confirmed by the architecture review. Neither calls
for rebuilding the test platform or rerunning every historical campaign.

## Ranked findings

### V0 — High: browser ledger validation can accept no validation

**Verified guard defect.** `journeys/support/cross-app.js:146–157` invokes
`go test -run '^<name>$'` and requires only exit code zero. It does not require the
named test to execute/pass or reject `SKIP`; it has no required test-verdict
boundary beyond process success.
An actual targeted probe in `verification/` returned exit zero:

```text
go test -mod=readonly -count=1 -run '^TestThisNameMustNotExistForReview235$' .
ok github.com/casselc/goo/verification 0.009s [no tests to run]
```

This helper would accept that output. A missing ledger variable can also lead to
an ordinary successful Go skip; for example
`verification/measurements_live_test.go:25–29` explicitly skips without its ledger.
The current callers usually supply the variable correctly. A static scan found
106 literal browser-validator references and **no missing current declarations**;
this finding is not evidence that previous reported ledgers silently skipped.

The new `scripts/verify_optional.py:55–86` already solves the corresponding problem
for its own named suites by consuming Go JSON events, requiring named passes and
rejecting skipped subtests. Browser ledgers bypass that guard.

**Bounded response:** apply the same required-verdict contract to `verifyWithGo`:
Go JSON events, exact named pass, no skipped required test/subtest, nonzero/error
handling and bounded timeout. Keep model test names/version/tool context explicit.
Share verdict fixtures or a small common contract rather than growing unrelated
parsers. Add deliberate nonexistent-name, unset-ledger, skipped-child and failed-
model controls; one existing real ledger should still pass. No full browser suite
is needed to prove the guard itself.

### V1 — High: incoming-body cancellation is a missing failure class

**Confirmed by the independent architecture reviewer; test/source gap inspected
here, probe not duplicated.** `internal/queryhttp/budget.go` releases admission
only when the downstream handler returns; `internal/httpjson/json.go:16–27` can
block reading an incomplete request body. Cancelling a context does not itself
interrupt that socket read. The architecture review reproduced cancel returning
202, context cancellation, a subsequent 429, and capacity releasing only after
peer close; see its separate review/reproducer for exact evidence.

Current tests are valuable but exercise different phases:
`internal/queryhttp/budget_test.go:12` starts at response headers;
`cmd/goo/operation_budget_test.go:42–109` waits until a source transport starts
before cancelling. They do not establish bounded admission lifetime while a peer
is still uploading JSON or withholding the EOF needed by strict trailing-data
validation.

**Required regression:** real local HTTP connection, one admission slot, partial
JSON (also a complete value with an unclosed body), cancellation/deadline, no
source execution/publication, bounded slot release while the peer stays open,
and a succeeding next request. Cover response status where the peer can receive
it, cancellation-before-admission and normal strict JSON refusal. A bounded network
failure test is more useful here than another whole UI or numeric property run.
Do not infer shared-production readiness from context deadline assertions alone.

### V2 — Medium: the browser tier has no required-suite verdict boundary

**Verified structure; readiness risk rather than a reproduced current failure.**
There are 259 `journeys/*.spec.js` files and 18 unconditional historical skip
statements. Other campaigns intentionally skip unless their environment flag is
set. `scripts/run_journeys.sh:49–55` ultimately delegates to Playwright without
requiring a selected set of named outcomes. `playwright.config.js:4–19` supplies
one Chromium worker, failure recordings and timeouts, but no required smoke
manifest. Thus “the browser command exited zero” and “the required journey ran”
are different claims. The current documentation often distinguishes them correctly.

Keep historical specs as archives or explicitly excluded history, not an apparent
current coverage set. Introduce a small named core selection, with required test
IDs/outcomes and skip refusal for that selection; intentional opt-in campaigns
remain optional outside it. `--forbid-only` is a cheap CI guard. Include one actual
cross-surface task and explicit ownership/failure checks, not all 259 scenarios.
Continue to report skipped optional campaigns as skipped. Do not put a full native/
model/Bombadil/three-browser matrix on every small UI edit.

### V3 — Medium: deployed-asset identity and disposable-lab setup remain implicit

**Verified boundary; no current stale-deployment failure claimed.**
`check_deployed_api.py:14–17` proves the served spec matches the checkout. It does
not prove the running Go binary and copied UI modules came from that commit.
The recent validation record correctly says individual assets were copied into
an existing lab and that images/binaries were not rebuilt. A matching OpenAPI
shape is compatible with an older implementation having the same wire format.

The repository pins tool versions and sibling source well for the optional
entry point (`verification/README.md:5–47`). Browser commands still depend on a
preexisting Chromium path (`playwright.config.js:15`), running ports and an existing
lab. The thin `--suite lab` check explicitly does not provision/restart or prove
collector ingestion (`verification/README.md:49–65`); keep that distinction.

**Bounded enabling task:** define one clean-lab/selected-core entry point with
explicit prerequisites and a small commit/asset/binary identity record. It should
refuse a mismatched deployment or clearly label that it is reviewing one, without
forcing rebuilds for every browser check. Use a disposable test store/profile for
publication scenarios. No automatic reset of a user's existing named volumes.
A small manifest is enough; a permanent evidence lake is unnecessary.

### V4 — Medium: aggregate green counts can obscure the test tier

**Verified current limits, not a demand to make every tier mandatory.** The sole
workflow `.github/workflows/api-contract.yml:21–48` runs root Go/Node, independent
instrumentation/example modules, API regeneration and direct/proxy observations.
It does not run the separate Hegel/Quint/woven module or any live browser campaign.
That is a reasonable fast tier, and current docs are candid. A “588 frontend tests”
claim is unit/module coverage, not 588 full-browser or independent-source checks.
Some Node tests use deliberate dependency/DOM fixtures through
`tests/support/module-dependencies.cjs:10–25`; recent 234 viewport/focus journeys
show why those fixtures need real-browser complements.

Keep a concise tier summary in validation results: required fast checks, required
selected journey/model verdicts, skipped opt-ins, deployment identity and material
limits. Add a manual/scheduled pinned native/model/core-browser job when its clean
prerequisites are available; do not silently skip it when required. Chromium-only
coverage should remain explicit. A targeted Firefox/WebKit component smoke is a
later platform gate, not a reason to triple all historical journeys immediately.

### V5 — Low/medium: subprocess and evidence boundaries can be tighter

**Source-confirmed risks, not observed hangs/disk exhaustion in this review.**
`verify_optional.py:20–24` captures subprocess output with no outer timeout.
Its `go test -timeout=5m` bounds test execution but not Go/Orchestrion build time or
preflight executable hangs. The browser helper has a 60-second child timeout
(`cross-app.js:150–151`); Bombadil has its own process timeout and active marker
(`cross-app.js:105–135`). Add explicit per-phase outer timeouts/process cleanup
where required, without treating a slow compile as a model failure.

`retain_diagnostics.py:14–16,83–98` already bounds raw recordings by outcome/age and
budget. Structured ledgers and published immutable captures are different data.
Repeated real publication journeys can grow a shared store toward the 10,000-entry
catalog bound (`exploration/capture_store.go:29,505–513`); raw-recording pruning
cannot solve that. Isolate disposable test publications instead of deleting user
captures or weakening scan/integrity refusal. Preserve a minimized failing input,
model/binding version and seed when needed; routine passing runs need a small
result summary and maintained documentation frames, not every raw artifact.

## Strengths worth retaining

- API gates check generated shape drift and actual handler/proxy observations;
  frozen storage identities and corruption/symlink/publication tests protect more
  than a happy-path schema. Preserve them when P2 changes metadata/query cursors.
- `verification/model_test.go:33,92` combines generated traces driving real
  preparation with observed-trace checking and a rejecting mutation. It does not
  merely run a model in isolation. Missing Quint already fails in
  `model_test.go:17–29`; avoid replacing that with unconditional optional skips.
- Hegel/state/swarm uses actual implementations and independent oracles in many
  semantic families. Keep exact-numeric, reset/coverage, namespace and historical
  time properties; avoid duplicating implementation formulas as their own oracle.
- Representative Bombadil actors assert terminal phases, source budgets and
  independently observed DOM outcomes, e.g. `200-runtime-bombadil.spec.js:11–15`.
  Many are bounded scripted workflows, not broad random UI exploration; name the
  coverage accurately. They complement rather than replace adversarial scheduling.
- Recent 233/234 journeys check retained node identity, native selection, actual
  focused element and viewport geometry alongside source counts. These are useful
  distinct failure classes, unlike screenshots alone.
- The imported handoff is marked historical and optional gotogether work is
  identified as an isolated pilot. I found no active repo-owned SKILL.md bundle
  to audit as a production dependency; do not claim the pilot is an integrated
  testing layer or silently upgrade it to replace quintgo.

## Acceptance additions for the refreshed product backlog

| Task | Add or preserve the following gate |
| --- | --- |
| P1 Activity scanning | Use the same fixed short/long records for before/after geometry. Include keyboard order, accessible names/states and target hit areas, selection versus inspect, source-free disclosure, append/eviction/pause/resume and withdrawn actions. Real traces use the current source owner; classify admitted GET and POST reads correctly rather than treating every POST as telemetry. No numerical Hegel campaign for a CSS-only change. |
| P2 Capture discovery | First define query-bound cursor/version/index publication rules. Exercise later-page matches, duplicate titles, literal-empty versus missing metadata, equal-time ID ties, concurrent publication/deletion/rebuild boundaries, corruption and scan limits. A trusted metadata index cannot silently hide a corrupt immutable capture. Whole-catalog search, API fixtures and source-free reopen need distinct checks. |
| P3 Comparison support | Independent exact-count/ranking oracle; permutation and threshold boundary cases; zero/unavailable side counts; unchanged full totals and immutable captures. Name threshold semantics before tests. State tests matter if saved settings/async ownership changes; do not invent a significance model merely to supply a new Quint file. |
| P4 Trace-question reuse | Preserve different-span witnesses and the permitted same-observation witness. Reject absent conditions and unsupported ordering/topology assumptions; share/reopen must remain source-free. Versioned definition identity and old question compatibility need frozen fixtures. A witness page is not complete trace custody. Require a real model/ledger verdict through the repaired guard. |

## Proposed execution order

1. Repair browser-to-Go required verdicts and add their deliberate false-green
   controls. This is a small enabling cut, not a new testing framework.
2. Fix the architecture review's stalled-upload admission defect with a real
   bounded network regression. Keep it ahead of new feature confidence claims.
3. Proceed with product P1, then P2. In parallel only when useful, define the small
   clean-lab/deployment-identity tier so future browser evidence is reproducible.
4. P3/P4 as their contracts settle; add only relevant properties/models/journeys.

This changes the refreshed queue's enabling gates, not its product purpose. No
basis was found to reopen completed package/identity refactors or to make large
passing evidence bundles a prerequisite for ordinary development.
