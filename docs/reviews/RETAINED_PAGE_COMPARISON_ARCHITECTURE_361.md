# Retained span-page comparison — architecture contract 361

**Implementation, execution and final design-frame review: accepted.**
Stored-only admission, frozen pair versus current candidate, verified opaque side
locations and independent raw-label provenance have passed source, full UI,
generated, native, observed-model and focused actor checks. Genuine-frame design
acceptance is being recorded separately.

Depth cut after captured span-page display 360, following checkpoint 359. This
contract reuses the existing comparison model and stored-page owner; implementation
and validation remain separate. No telemetry acquisition, persistent comparison
artifact, capture format migration or global reference manager is added.

## Explicit operand grain

Compare the **whole verified current stored span page** on each side. Do not
silently filter a trace, merge stored pages, expand a trace or turn a selected row
into an inferred complete trace. A capture page may contain one, multiple or no
valid reported trace IDs. The existing model supports these cases explicitly.
A later trace-subset feature would need an explicit selection and original-index
mapping; it is not a hidden prerequisite of this cut.

The existing Go `CompareTracePages` and JS `trace-page-comparison.compare` consume
two bounded raw row arrays and known truncation flags. Reuse
`goo-trace-comparison` with `{baseline:{rows,truncated},selection:{rows,truncated}}`.
Its page origin is each whole page's earliest valid start, with a common duration
scale. This differs from 360's per-trace relative origins and must be disclosed.
Mixed/unreported trace pages remain useful group observations, but the model
withholds unique row pairing with `mixed-or-unreported-trace` rather than pretending
their aligned labels identify invocations.

## Verified immutable inputs and independent provenance

Only the existing SavedRecords `showSaved` admission path may supply operands:
`SourceSnapshots.adoptStoredExplore` verifies the immutable receipt/question and
provides its immutable supplied RecordPage. Keep each operand's capture ID,
original receipt/question, page-adjusted question, exact environment policy,
service/namespace/predicates, basis/source read clock, custody clock, offset/limit,
returned-row count, retained-record count and retention-limit status independently.
Metadata is detached/frozen when retained; reuse the already immutable RecordPage
rather than copying or mutating its rich rows on every render.

The full query count is not this page's row count. Different signals are refused;
logs are not a span comparison input. Different fixed windows or checked policies
may be compared as supplied observations, but neither operand may acquire the
other's scope. Explicit roles remain baseline/reference and selection/current;
choosing a reference is not an automatic query execution or baseline capture.

Known partial-page state includes offset greater than zero, a reported next page,
or original capture retention limits. A false known-truncation flag never claims
complete source coverage or a complete trace. Keep no-data, invalid timing,
changed-between-reads and omitted records visible as existing model diagnostics.

## Reference lifecycle and actions

Keep one explicit held baseline and one explicit selected current page in the
existing SavedRecords instance/tab. Replace or clear through current visible
controls; navigation may retain the chosen reference but must not silently change
its bytes, page or clocks. No global singleton, durable reference store, automatic
next-page selection or automatic source refresh is required.

Every callback captures its operand selection and current view/render lease.
Page changes, new capture responses, suspension, disposal, clear/replace and late
metadata must prevent retired callbacks from staging or opening a different page.
The component is display-only; events and mounting handles do not confer capture
or source authority. Keep receipt identity independent of row IDs, labels and
visual sorting. Duplicate IDs are different original occurrences.

For row inspection, reuse exact stored occurrence inspection if already available;
otherwise offer explicit **Open baseline/selection captured page** through the
existing capture owner. The CapturePage DTO exposes previous/next cursors but not
the current cursor. Capture that location from the verified root `captureVisit`
when retaining an operand, then delegate `{id,cursor}` to the existing open path.
Do not invent a cursor from an offset, open page one while claiming the original
page, or derive source authority from a passive component's row index. The local
pair-attribute inspector remains available only for model-certified unique label
matches; this is not identity of an invocation.

## Quantities and meaning

Groups align reported service/operation labels. Repeated labels are ambiguous;
unreported labels and mismatched span kinds do not become unique pairs. One-sided
groups say “not in supplied page,” not zero work. Inclusive duration sums can
overlap. Outside-observed-children subtotals retain unknown coverage; they are not
exclusive CPU/self time, causal attribution or critical path. No fleet percentile,
request deduplication, aggregate denominator or new completeness witness is added.

## Acceptance

Use genuine independently captured single-trace pages for unique-pair inspection,
then mixed-trace/repeated/one-sided and partial-page cases. Keep scalar and typed
policy, namespace missing/empty, unequal windows/extents, duplicate IDs, invalid/
zero/large timing and no-data cases in independent model/property fixtures. Prove
exact operand rows/order/location/receipt/question/clocks, source-free comparison,
local inspection, metadata-only side opening and Back. Late response, clear,
replacement, route/page retirement and disposed-owner negative controls must not
revive old operands. Real desktop/narrow frames must make roles, current page
grain and independent provenance understandable. Do not manufacture a new model
lifecycle for events the browser did not actually observe.

## Closed root decisions and candidate review

The applied pair remains frozen independently of the current candidate through
capture/page navigation and explicit side opening. Each new verified page updates
only the candidate. Compare is explicit; Replace baseline/Clear retire the prior
pair. App-control leases rotate on update/suspend/resume without rerendering the
passive pair, preserving its internal local inspection state. Root side-opening
uses the actual loaded visit's opaque `{id,cursor,offset}`, verifies the returned
visit/page before focus and preserves independently moved focus. POST-created
page-zero receipts establish a loaded page-zero visit before operand selection.

Group identity is exactly the existing nullable raw `(ServiceName, SpanName)`
key, not a namespace-qualified OTel service identity. Namespace and other scopes
remain independently disclosed operand metadata; neither raw rows nor grouping
keys are transformed. Matching labels remain descriptive even across different
namespaces or policies, and do not become invocation identity.

Source review found the route integration sound but reported two bounded closure
items: make the raw label grain clear in the comparison note, and distinguish
stored admission from generic admitted record pages in the exported operand
helper. `getRecordPage` alone accepts both fresh and stored adapters; the actual
SavedRecords path already adopts stored pages, but a fresh admitted wrapper plus
receipt-shaped metadata should not satisfy a helper advertised as stored-only.
Final source acceptance awaits root's bounded resolution; runtime gates remain
separate.

The stored-only association is now implemented as
`SourceSnapshots.getStoredRecordPage(captured)`. It returns a supplied page only
for the exact wrapper returned by stored admission, with the same admitted
exploration response, detached deeply frozen receipt, count, retained count and
retention flag. Generic fresh adapters, copied claimed wrappers and changed
metadata fail closed. The association requires the actual canonical uint64 count,
bounded retained count and Boolean retention flag; it changes no public stored
format. Caller receipt/row data are not frozen or mutated. Focused source/page
checks and new direct negatives passed; the UI consumes this narrower API.


Final integration review confirms the controller consumes `getStoredRecordPage`,
with fresh/copied-wrapper negatives. Shared comparison labels consistently say
Baseline/Selected; the app explicitly states raw service/operation grouping is
not namespaced service or invocation identity. Observation-state facts disclose
changed original reads or no data separately from supplied-row comparison.
No source authority is inferred from the passive component. Current candidate
changes and stored-side opening retain the existing frozen pair and its row/data
identity; replace/clear and app-control leases retire old actions. Root verified
cursor/focus/page-zero save integration and focused UI/source checks are reported
passing. Source acceptance is closed with no remaining blocker; execution gates
remain separate and are not presumed by this review.


## Final execution acceptance

The parent reports full UI **1,234 tests passing**, generated Hegel **40 cases**,
initial Quint lifecycle checks with two negative controls, and the canonical native
journey's **18 real-browser facts** passing. Native local work issued zero source
queries/SQL and ten evidence-store operations; immutable capture bytes were
unchanged. Required observed-model validation passed in 15.22 seconds with zero
skips and six corruption negatives. The positive actor passed 574 snapshots and
12 stages in 35.689 seconds, with zero source calls/writes and three evidence-store
GETs. These are observed effects, not synthetic completeness claims.

A final bounded recovery correction is also accepted: failed stored-side opening
marks the target visit failed, clears `aria-busy`, then resumes an eligible
previously held exploration page. Existing current receipt/hash/visibility guards
prevent restoration into another route; the failed visit supplies no candidate
location. Held comparison actions recover without adopting the failed target or
issuing telemetry. Focused tests retain the original receipt/hash and assert null
failed location and resume ordering. Final genuine-frame design review remains a
separate gate being completed by the design owner.
