# Roadmap after the 384 product checkpoint

Adopted November 9, 2026 after the 384 runtime and product checkpoint. This
replaces the completed [381 sequence](ROADMAP_2026_11_08_CHECKPOINT_381.md). The
[current capability/gap index](CAPABILITY_GUIDE.md) summarizes delivered behavior;
[validation](../VALIDATION.md) records actual checks.

Inputs: independent [design](reviews/WHOLE_PRODUCT_DESIGN_384.md) and
[architecture](reviews/WHOLE_PRODUCT_ARCHITECTURE_384.md) reviews, current source,
and the measured fixed-result layout baseline. The reviews are inputs to this
finite queue, not competing execution plans.

## Product direction and lessons

The primary loop remains current telemetry → isolate a population → follow its
evidence → preserve or reuse the question. OTel data is useful without a Goo
manifest. Existing comparisons, explanations, log patterns, service/resource
observations, review labels and scoped display grammar already cover broad tasks;
they are not missing categories to recreate.

382–383 joined cross-service search intent to immutable evidence and retained
follow-through. 384 consolidates the evidence surface. Natural-top geometry in
the actual scroll container is the usability contract; a screenshot after forced
scroll is insufficient. One passive summary can serve live and stored views while
question sessions, reads and custody remain in their existing owners.

Generated Go/JavaScript admission found a real disguised-family capture hole in
383. New disclosure composition also exposes hidden-action and transient-status
requirements. Keep semantic generation, actual-fact replay and positive visible
browser completion; avoid new frameworks or tests that merely mirror DOM moves.

The adopted cadence is **385 breadth (delivered) → 386 facet depth (delivered)
→ 387 consolidation (next)**.
Complete each acceptance cut before the next. This prerelease has no compatibility
floor requiring aliases, migration adapters or old public formats. Current Go
1.27.3 and standard libraries remain the baseline; use newer facilities when they
simplify the work, without imposing type machinery on callers.

## 385 — breadth: investigate outgoing HTTP dependencies

Delivered: [contract](OUTGOING_HTTP.md), [storyboard](journeys/385-outgoing-http.md)
and [actual checks](../VALIDATION.md). The remaining queue starts at 387.

User task: “Which outgoing web dependency is slow or has recorded errors?”
Current service profiles cover incoming HTTP/RPC, DB clients and messaging, but
have no outgoing HTTP-client lens. Reuse service inventory, DashboardRequest/Plan,
ordinary typed record questions, captures and retained trace follow-through.

- Pin and review the current OTel HTTP client convention before implementation.
  Select exact Client-kind observations with a nonempty `http.request.method`;
  preserve service, exact namespace, environment policy and fixed bounds.
- Expose reported destination/method dimensions, starting with qualified
  `server.address` and supported `server.port` context. Missing and present-empty
  values stay distinct. A remote address does not identify an instrumented callee
  service or verified topology.
- Preserve full matching Client-span record counts, duration and recorded span
  status. Additional reported retry spans remain observations; do not infer
  unreported attempts, unique requests, remote server health,
  causality or a universal HTTP failure policy. Keep response-status evidence
  separate from recorded span Error and unreported status.
- Make the outgoing lens an explicit discoverable choice. Metadata candidacy
  cannot weaken the actual read predicate or silently replace the current lens
  after an empty/failed read. Preserve existing supported profile behavior.
- Follow a reported trace through the established explicit fresh-read/held-Back
  flow, save its checked question and capture compatible evidence.

Acceptance includes independent raw Client/Server and method/destination evidence,
Hegel boundary cases, unchanged totals under bounded lists, exact scoped requests,
actual desktop/narrow service-to-trace journeys and measured no-read Back. Native
`http.client.request.duration` profile adoption is separate future depth; generic
instrument discovery is not a claim of that profile's execution semantics.

## 386 — depth: reuse native gauge display configuration

Delivered: [contract](NATIVE_GAUGE_VIEW_REUSE.md),
[storyboard](../journeys/386-native-gauge-view-storyboard.md) and
[actual checks](../VALIDATION.md). The remaining queue is 387.

User task: “Use this useful gauge view on another compatible gauge.” Histogram
configuration reuse already exists. Extend the typed native display contract with
a bounded gauge arm using existing gauge sample/staleness rendering and scoped
view machinery; do not add another registry or measurement executor.

- Save only display intent and its applicability contract. Apply to an exact
  compatible per-stream gauge destination; keep destination source, unit, grain,
  question, read clock and coverage authoritative.
- Stage configuration without telemetry reads. Run remains explicit. Donor
  streams, resource identity, values, completeness or custody cannot transfer.
- Preserve missing/stale samples and gaps. Do not fill absent values with zero,
  change units implicitly, aggregate incompatible streams or infer health.
- Reject incompatible family/unit/grain and unsupported kinds before applying.
  Begin with gauge; counter reset/rate policy reuse is a separate later cut.

Acceptance includes independent declaration round trips, generated compatible and
refused applications, supplied-data rendering, exact destination request equality,
immutable saved configuration and real save→apply→Run journeys with current frames.
Choose the smallest public API justified by the existing grammar.

385 reconnaissance confirms that gauge reuse must not merely relax the histogram
arm: explicit bins, complete export intervals and the histogram seconds/milliseconds
unit whitelist do not describe gauge observations. Prefer a small discriminated
native-display envelope over duplicated save/apply managers. Reuse the existing
gauge timeline/table renderer, keep exact destination units and sample/staleness
semantics, and distinguish a retained sample bound from continuous coverage.
The [386 design review](reviews/NATIVE_GAUGE_VIEW_REVIEW_386.md) finalized the typed
arm before implementation. The adopted gauge
arm uses the existing compiler's nonempty exact-unit policy; this cut does not
broaden native measurement admission.

## 387 — consolidate: close current capture-family admission

User task: “A saved item always means the evidence its receipt says it contains.”
383 fixed one concrete arm leak; current family shortcuts still merit a bounded
whole-envelope pass. Consolidate the six existing full capture families, then
retain their pure family-specific semantic validators.

- Check receipt kind, question arm, result arm and page metadata coherently before
  family-specific shortcuts. Refuse disguised, mixed, null question/result arms and mismatched payloads.
  Preserve legitimate nullable in-family data, such as a missing qualified witness
  value; an absent value is not a forbidden payload arm.
- Preserve each family's request, source/custody clocks, totals, bounds, warnings
  and pagination semantics. Historical plans remain evidence; retain each family's existing refresh and
  compiler-drift semantics rather than imposing one gate on every family.
- Distinguish full capture pages from metadata-only investigation references.
  Metadata cannot acquire result authority or require a telemetry read.
- Use generated Go/JavaScript cross-language mutants over the finite current
  union. No universal schema engine, generic registry or compatibility layer is
  needed. Remove duplicated checks only where their replacement is clearer.

Acceptance covers all current valid families, rehashed invalid payloads, direct
HTTP and UI refusal, source-free reopen, immutable bytes and actual retained
follow-through. Query execution and persisted evidence meaning remain unchanged.

## Deferred alternatives and checkpoint

Contextual review-task discovery → verified kept-occurrence review remains a
coherent later pair. Log-pattern display grammar admission and native counter
configuration reuse are valuable later depth. Neither is a prerequisite for this
queue. Service exemplar adoption and general exponential evaluation keep their
separate source/mathematical gates.

Broad DB/broker server health, alert delivery, authenticated tenancy/presence,
contributed executable kinds/WYSIWYG authoring, warehouse lowering and new storage
engines remain wider initiatives. Do not turn this list into a parallel backlog.
385 also exposed an existing navigation-affordance mismatch: ordinary widget/group
links return through Services, while Customize/native actions offer contextual
Back. Both paths can retain evidence without reads; unify delegation in a later
bounded usability consolidation rather than adding a second navigation owner.

386 also reinforced that shared chart alignment must be checked in physical
coordinates under real application styles, and collector acceptance is not proof
of source settlement. Keep raw-source and geometry gates independent of click
completion or percentage attributes. Bound disposable builder cache; preserve
source stores when recovering lab disk pressure.

Actual 386 narrow frames also show that native stream identity, diagnostic prose
and save controls can push the primary visual low in the viewport. Reassess
native-display density at the next whole-product checkpoint; 384's verified
first-item contract covered fixed record surfaces, not every native screen.

Reassess the whole product after 387 or earlier if actual usage invalidates the
chosen sequence. Update this one queue from evidence rather than appending plans.
