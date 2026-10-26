# Isolated lifecycle shrinking pilot

Core and qhegel are remotely resolved at `v0.0.0-20261026141556-65093ec1c98a` (commit `65093ec1c98abd095f6d399fae038e466555c231`), with go.sum verification enabled. The only active replace is the actual Goo module at `/workspace/goo`; Goo's dependencies and source are untouched. `pinned/` is an unused source archive for reading pinned skills, not a module replacement. Research checkout HEAD remained `2992868097ec5cb1eecf230097739bf090d91573`.

The copied existing `model/lifecycle.qnt`, binding and `world_test.go` retain Goo's bounded one-subject custody model: acceptance60, terminal70 delivered95, basis80/95/96/100, at most3 frozen snapshots. The existing production `lifecycle.Policy.Prepare` executes; recording uses the new toolkit qobs through qhegel. No lifecycle algorithm is reimplemented. Six explicit faithful schedules cover cutoff equality, strict-after visibility, duplicate delivery, terminal-before-accept and snapshot retention.

`overlay.json` substitutes only a copied `prepare.go` at build time, changing delivery exclusion `received >= cutoff` to `received > cutoff`. No production edit or copied implementation is selected in the faithful build. `FindViolation` varies zero to six redundant duplicate deliveries, seed3, at most8 generated cases. A valid result must be a model **state mismatch after advance95**, not a compile/tool failure or merely an invariant violation. The smallest trace must be exactly terminal→advance95. Both proper one-action subsequences are checked directly against the original bounded projection; removing either action eliminates the mismatch. This establishes minimality within this constrained two-required-action/duplicate-noise search, not a globally minimal arbitrary lifecycle program.

Commands (run sequentially):

```sh
export PATH=/workspace/.toolchains/go/bin:$PATH
export GOPATH=/workspace/.go GOMODCACHE=/tmp/goo-module-cache
export GOCACHE=/workspace/.cache/go-build XDG_CACHE_HOME=/workspace/.cache
export GOTOOLCHAIN=local GOMAXPROCS=2
go mod tidy
go test -run '^TestFaithfulBounded$' -v . > faithful.log 2>&1
PILOT_MUTANT=1 go test -overlay overlay.json -run '^TestMutantShrinks$' -v . > mutant.log 2>&1
```

The validator explicitly selects `/workspace/goo/node_modules/.bin/quint`0.32.0, TypeScript backend, `Eval:"quint"`; missing Quint fails rather than skips. No Rust/Apalache, live telemetry, concurrency scheduler or durable-storage completeness claim follows. No CI-green claim is made: remote CI API was unavailable/Forbidden in the surrounding research.

Setup diagnostics are retained separately: the first local preparation hit a restricted checksum-proxy socket; subsequent remotely downloaded pinned modules resolved with normal checksums enabled. Initial test compilation used nonexistent `hegel.TestCase.Fatalf` (corrected to Errorf); the next invocation required a writable XDG cache instead of the read-only default home cache. These were setup/harness failures, not observed semantic counterexamples.

Actual results: faithful six scenarios PASS13.201s (`faithful.log`); overlay mutant expected mismatch/shrinking PASS33.953s (`mutant.log`), seven Quint checks and seven cache hits. The saved `smallest-mutant-trace.json` contains exactly terminal→advance95, with unknown=true observed at equality although the model keeps the terminal invisible. Both proper single-action subsequences matched. `receipt.json` records source/module/model/overlay hashes; `overlay.diff` shows the single intentional comparison change. Actual production prepare.go SHA256 remained `8b85766e87a34d0d324a73fe039930a8f94402c7920641c1c9b38679e344a3ca`. No further campaign was run after success.
