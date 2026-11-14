# Acquisition consumer fixture for gotogether CW1

A user acquires an OTLP-backed view, inspects locally, leaves and returns, then cancels a reread. A late successful response must not replace the retained view. The final visible view must preserve its bank, window, question and local inspection state.

This is a frozen handoff, not an implemented Markdown compiler. The source model remains `../../model/acquisition.qnt`. CW1 should extract the fenced model below, map a deliberately broken invariant back to its document line, reject stale generated output, and replay this log. Those extraction/source-map features belong upstream and are not claimed here.

Files: [model](model/acquisition.qnt), [binding](acquisition.binding.yaml), [scenario and bridge](scenario.go.txt), [recorded log](requests.jsonl), [manifest](manifest.json).

The protocol is the upstream [qdriver request contract](https://github.com/chucklehead-dev/gotogether/tree/a1a90fe5653db07cbf6cc467021bda18e4b27d79/qdriver). Producer: [application adapter](../../../../journeys/support/gotogether-driver.js) and its unchanged vendored driver (see [provenance](../../../../journeys/support/gotogether-driver-PROVENANCE.md)); the manifest records the source commit. The source commit precedes this fixture-only handoff and dependency pin update.

## Scope and replay

`requests.jsonl` is a verbatim real browser capture using `gotogether.requests/1`. It contains synthetic fixture identifiers and bounded DOM observations, not the 240 raw OTLP records. The native source gate passed during capture; static replay does not rerun its SQL assertions or establish global telemetry completeness. HTTP completion alone does not prove DOM adoption.

`scenario.go.txt` snapshots both the exact `Scenario()` expectation and the drain bridge. The named scenario requires read 1 and the ordered explicit-read, local-inspect, retained-return, cancel-pending-reread, late-success-ignored and final observations. The YAML simulator's five-show possible outline is a different, weaker reachability check. Only an explicit final `drained` record becomes settled; EOF does not. Lost tails remain unjudged.

From the repository root, with Quint 0.33.1 and Go on PATH:

```sh
cd verification/gotogether
GOTOGETHER_STRICT=1 QUINT_BACKEND=typescript go test -mod=readonly -count=1 -run '^TestCW1Fixture$' .
```

The test checks snapshot hashes and drift, then runs the same positive log, negative controls and stale compiled-artifact refusal used by the live pilot. It checks both TypeScript source and compiled replay without Node/Quint on PATH. `npm run test:gotogether-shadow` remains the separate live native-source gate.

## Model snapshot

<a id="acquisition-retained-ownership"></a>

```quint
// Visible DOM ownership projection. Read IDs identify completed transports only;
// their bodies never establish a displayed bank. Navigation epochs are checked
// for monotonic progression; retained DOM identity rejects stale adoption.
// Source/query conservation remains the independent native gate's job.
module acquisition {
  var held: {bank: str, clock: str, question: int}
  var local: {page: int, find: str, detail: int}
  var acquired: bool
  var returned: bool
  var late: bool
  var finished: bool
  var navigationEpoch: int
  var reads: Set[int]
  var safe: bool
  action init = all {
    held' = {bank: "", clock: "", question: 0},
    local' = {page: 0, find: "", detail: 0},
    acquired' = false, returned' = false, late' = false,
    finished' = false, safe' = true, navigationEpoch' = 0, reads' = Set(),
  }
  action read(read: int, epoch: int, code: int): bool = all {
    read > 0, epoch >= 0, code >= 0,
    not(reads.contains(read)), epoch <= navigationEpoch,
    reads' = reads.union(Set(read)), navigationEpoch' = navigationEpoch,
    held' = held, local' = local, acquired' = acquired,
    returned' = returned, late' = late, finished' = finished, safe' = safe,
  }
  action show(phase: str, bank: str, clock: str, question: int, page: int, find: str, detail: int, epoch: int, visible: bool): bool = {
    val display = {bank: bank, clock: clock, question: question}
    val view = {page: page, find: find, detail: detail}
    val adopt = phase == "explicit-read" or phase == "explicit-reread"
    val retain = Set("retained-return", "cancel-pending-reread", "pending-departure", "late-success-ignored", "final").contains(phase)
    all {
      epoch >= navigationEpoch, epoch <= navigationEpoch + 1,
      not(adopt) or reads.size() > 0,
      navigationEpoch' = epoch, reads' = reads,
      held' = if (adopt) display else held,
      local' = if (phase == "local-inspect" or phase == "explicit-reread") view else local,
      acquired' = (acquired or phase == "explicit-read"),
      returned' = (returned or phase == "retained-return"),
      late' = (late or phase == "late-success-ignored"),
      finished' = (phase == "final" and acquired and returned and late),
      safe' = (safe and (not(acquired or adopt) or visible) and (not(adopt) or bank != "") and
        (not(acquired) or adopt or display == held) and
        (not(retain) or view == local)),
    }
  }
  action step = any {
    read(1, 0, 200),
    show("explicit-read", "bank", "clock", 1, 1, "", 1, 0, true),
    show("local-inspect", "bank", "clock", 1, 1, "", 1, 0, true),
    show("retained-return", "bank", "clock", 1, 1, "", 1, 0, true),
    show("cancel-pending-reread", "bank", "clock", 1, 1, "", 1, 0, true),
    show("late-success-ignored", "bank", "clock", 1, 1, "", 1, 0, true),
    show("final", "bank", "clock", 1, 1, "", 1, 0, true),
  }
  val retainedOwnership = safe
  val journeyFinished = finished
}
```

## Binding snapshot

```yaml
spec: model/acquisition.qnt
module: acquisition
# Opaque DOM text identities retain raw strings in the log; equality only.
intern:
  question: actor
  detail: actor
actions:
  read: {}
  show: {}
invariants: [retainedOwnership]
eventually: [journeyFinished]
complete: true
simulate:
  allows: [retainedOwnership]
  possible:
    - name: acquisition and retained return
      steps: [show, show, show, show, show]
```
