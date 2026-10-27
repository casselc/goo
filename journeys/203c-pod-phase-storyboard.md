# 203c · Reported pod phase observations

An operator asks: “What did this pod report during the selected window?”

Generated ordinary OTLP fixtures exercise Collector Contrib 0.137.0 conventions;
no actual k8sclusterreceiver, Kubernetes API, RBAC connection or physical-pod
identity is established. No positive source or UI response is mocked. Reported
phase is categorical evidence, not pod health or readiness.

## Actual flow and exact scope

1. Explicit Kubernetes discovery reads the full resource/schema inventory. The
   fixture has 22 metric points and 11 log/span records. Select a known-empty
   resource-schema variant with a reported UID and no service/instance labels.
2. A mixed resource keeps both existing memory defaults. Phase appears as an
   explicit candidate; Preview and retained Back stage it without a source read.
   A phase-only resource defaults to one exact categorical measurement.
3. The exact receiver scope is
   `github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver`,
   version `0.137.0`, empty instrumentation schema and empty unit. The physical
   descriptor remains a gauge; its definition is `state` with
   `stateProfile: k8scluster-pod-phase/0.137.0`. Full raw resource/schema/attributes
   remain pinned. Same UID with another resource schema/map does not broaden it.
4. Actual samples show Pending then Running. Seven additional resource reads
   prove observed code 5 **Unknown** separately from telemetry quality unknown;
   unsupported code 7, fractional 1.5, stale supported/unsupported, conflict and
   no-recorded-value observations preserve independent domain and quality.
   Incorrect receiver version/scope/schema/unit/kind produce no approved widget.
5. Inspect stages the ordinary native owner without a telemetry read. Back
   restores the held panel and opener focus. Explicit Run reads fresh source;
   individual Capture also performs a fresh source read and creates the existing
   immutable measurement capture. No composite pod store or phase arithmetic is
   introduced.
6. The real stored capture reopens with no attempted API POST under a guard that
   would abort unexpected source calls. This proves the source-free path, not an
   actual ClickHouse outage. Full before/after stored bytes and content ID remain
   unchanged. Phase has no counter/histogram interval-to-record action.

## Accepted receipts and source effects

[Main ledger](../artifacts/pod-phase/pod-phase-0313f9b522/ledger.json) passed
**24.6s**. The complete first five-case invocation passed **1.2m**:
phase ownership **14.6s**, prior memory ownership **17.3s**, and explicitly
synthetic generic component cases **2.5s/8.8s**. Those gallery/SSR/detached cases
are supplied examples, separate from actual OTLP evidence.

The main's **27 logical source operations** consist of one inventory, fourteen
metric metadata lookups, eleven native measurement requests and one fresh-read
capture. **Seventeen pure plans** are separate. Native Preview/staging, retained
Back, local disclosures and frozen reopen add no telemetry reads. Setup OTLP
posts/readiness probes and direct store GETs are outside the page journal;
logical API operations are not ClickHouse statement or scan counts.

[Phase owner ledger](../artifacts/pod-phase/pod-phase-a272d10e81/ownership.json)
retains actual fetched responses held across filter, close, another row, tab and
scope changes. Releasing them cannot publish into or replace the current owner.
The accepted previous memory owner is
[its own receipt](../artifacts/pod-memory/pod-memory-6835e1dac1/ownership.json).
No broader historical browser campaign was repeated.

Capture `5c9a9c2a15d4bb0e4c11b37db29d60c6f163411b049868052f354e43c7f7981f`
has **6,145 bytes**, SHA256
`ea2851f4c30cb122fcb825e68a4d4de936aa1964fd93b659233f0247838dd9c2`.
The full before/after files are adjacent to the main ledger.

## Actual frames

All three frames use the accepted main fixture and its original read clocks.
The component and whole workspace retain existing surrounding navigation;
this slice does not claim to solve overall editor/catalog density.

- [Desktop phase panel](../docs/journeys/screenshots/203ca-pod-phase.png).
- [Mobile phase panel](../docs/journeys/screenshots/203cb-mobile-pod-phase.png).
- [Whole workspace context](../docs/journeys/screenshots/203cc-pod-phase-in-context.png).

Actual desktop/mobile PNGs were inspected. Geometry checks cover X/Y containment,
full labels/actions, independent observation quality and exact UTC tooltips.
History is a categorical list; phase codes are never plotted on a numeric axis.

## Reproduce

```sh
npx playwright test journeys/203c-pod-phase.spec.js journeys/203c-pod-phase-ownership.spec.js journeys/202-pod-memory-ownership.spec.js journeys/state-observations-components.spec.js
GOO_POD_PHASE_BOMBADIL=1 npx playwright test journeys/203c-pod-phase-bombadil.spec.js
```

The first actor stopped at phase 3 because its readiness selector expected
obsolete status text. The actual panel was loaded with source 3/pure plan 1;
no app exception or extra read occurred. Its trace/screenshots are retained in
`artifacts/pod-phase/first-actor-diagnostics/` and fixture `pod-phase-aa47c977d9`.
Only the selector was corrected; accepted main/owner/component cases were not
repeated. Final actor and independent observer receipts are recorded below.

## Deployment and disk handling

Functional source `8d785ee` runs with binary SHA256
`24ca91df0f0a130f4bc1aa8d55d638a3b0c52e399f1bbaf3d3693281707190cb`.
[Deployment receipt](../artifacts/pod-phase/deployment.json) verifies twenty
changed files in both runtime roots and ten actual served asset/spec hashes,
MIME types and metadata readiness, including `state-measurement.js`.
Only existing packages/query containers were restarted; traffic and volumes
were preserved. Dependencies were relocated intact to `/tmp` via a symlink.
Six rotated diagnostic logs (594,850,797 bytes) were checksum-archived before
removing only their originals; active logs and telemetry tables stayed intact.

## Actor and independent raw/custody validation

Corrected actor passes in **31.0s** (32.2s invocation), nine phases 0–8,
379 snapshots, final source **4**, pure plans **2**, terminal true. Its actual
reported Unknown category, explicit native Run and retained Back are checked in
`artifacts/pod-phase/pod-phase-f636229d27/bombadil/acceptance-summary.json`.
The actor does not repeat the main journey's separate capture checks.

Independent native validation passes in **0.627s** (main 0.54s, owner 0.07s),
rejecting **twelve main and four owner mutations**. It reconstructs categorical
labels and observation quality from raw fixture records, binds exact resource/
receiver/schema/attribute membership, checks returned rows/clocks/bases against
actual responses, validates all seven boundaries and five refusals, and verifies
full capture bytes/content identity plus real held-response ownership.

The initial main observer expected 21 events, while the completed journey has
22 because explicit phase-candidate Preview/Back was added after memory defaults.
That step leaves source at four and adds one pure plan. The exact expected
sequence was corrected; raw, metadata, presentation, custody and owner predicates
were retained. The first owner passed. Initial diagnostics remain in
`artifacts/pod-phase/native-validation.initial-observer.txt` and
`native-observer-diagnostic.txt`. No application/fixture/browser correction or
repeat was needed for this observer-schema fix.

| Main interaction reached | Cumulative source operations | Cumulative pure plans |
| --- | ---: | ---: |
| Inventory | 1 | 0 |
| Mixed memory defaults | 4 | 1 |
| Explicit phase Preview and Back | 4 | 2 |
| Phase-only default | 6 | 3 |
| Native staging and retained Back | 6 | 4 |
| Explicit native Run | 7 | 5 |
| Fresh individual capture and retained return | 8 | 5 |
| Seven explicitly opened boundary panels | 22 | 12 |
| Five explicitly opened refusal panels | 27 | 17 |
| Frozen capture reopen | 27 | 17 |

The general display component preserves history disclosure-open state on data
updates; descendant focus across rerender is not guaranteed. App-owned retained
Back focus is separately verified. No automatic live refresh is introduced.

Independent source and final release reviews approve this bounded slice. Compact
committed [receipt](../docs/evidence/pod-phase-203c.json) records source/binary,
source effects, capture and actual frame hashes; full session diagnostics remain
under the normal retention policy. No new Quint model is claimed for the unchanged
explicit read/retained-owner algebra. No actual Kubernetes receiver deployment,
API/RBAC, complete pod inventory, health verdict or ClickHouse outage is claimed.
