# Supplied-span waterfall — architecture review 301

Review begins at checkpoint `84e1132`. The contract below is accepted scope;
implementation review and executed validation are identified separately.

## Decision and ownership

Add `record-span-waterfall`, version 1, as a spans-only kind of the existing
records role. It consumes the same admitted `goo.record-page/1`, configured core
columns and up to eight qualified fields as Table/List. Required UTC and summary
remain. A hidden role still obeys the signal/kind declaration contract. Saved
views, authored templates and local kind changes retain the same query, page,
source token, basis, read clock and pager. There is no new source endpoint, query
owner, capture shape or measurement total.

The public component is passive: `.page`/`.display` and equivalent JSON attributes
support server HTML and detached SPA construction. Local Details, layout and
keyboard behavior do not fetch, navigate or publish. Application adapters mount
the existing shared record controller into explicit local slots. Serialized data
never grants record or trace authority.

Keep explicit Expand trace in its existing independent context owner. That action
may acquire spans outside the service/predicate scope of the held page. A local
waterfall must never secretly make that acquisition, imply complete traces,
compute a critical path, infer invocation counts or promote references into causal
edges.

## Geometry and relationships

Reuse the existing waterfall layout/row presentation rather than building another
chart implementation. An explicit `trace-relative` origin option can coexist with
the existing global-origin default for current inspector consumers. Parent safety
improvements apply to both modes; any broader visual default change needs its own
review of current consumers.

For each valid reported trace, subtract its earliest **supplied timed span** in
exact nanoseconds. Use one common duration scale across groups: the maximum
supplied trace-fragment extent. Independent per-trace autoscale would make a 5 ms
and a 500 ms span appear equivalent; one global multi-minute page extent would
turn both into unreadable slivers. Label the chosen origin/scale and retain exact
absolute UTC and duration tooltips. Neither origin nor extent is the real trace
start, complete trace latency or service-response denominator.

Resolve parent candidates through a multimap of valid same-trace span IDs. An edge
requires exactly one reported candidate occurrence; duplicate candidates are
ambiguous, not an invitation to choose the last row. Unknown/invalid trace IDs
must not be combined into an implied trace or share parent edges. Invalid,
unreported, zero/root, missing, ambiguous and cyclic parent states remain explicit.
Reported async links remain references and do not add parent edges.

Malformed timing does not become zero and does not remove an observation. Retain
its original page index, structured fields and Details/Inspect capability, show
Timing unavailable, and exclude it from numeric origins/extents/scale. A group
with no valid timed observations has unknown timing, not zero extent. Valid
zero-duration spans are known-zero observations and remain inspectable; a numeric
denominator safeguard must not change their reported extent.

Subtract BigInts before approximate Number coordinates. Use a bounded scalar
projection for layout/ancestry so rendering a rich page does not clone its entire
payload once per row. Rich full records remain in the admitted page for explicit
inspection. A bounded presentation is not evidence of complete upstream telemetry.

## Occurrence identity and mounting

Parent-first visual order differs from source order. `.recordItems()` returns
frozen handles indexed by original page position, even if their elements occur in
a different DOM order. Keep, selection and inspector keys use this occurrence
position. Two identical records or repeated object references remain separate
observations; no `rows.indexOf(record)` or object-key deduplication is allowed.

Each handle supplies the same action/detail mounts used by List. Native waterfall
row buttons must have adjacent checkbox/action hosts and inspector content, never
nested interactive elements. A timing button may be supplied as an additional
handle for the application to attach the same Inspect behavior. Its default local
behavior remains passive.

Page/display replacement creates fresh hosts; detached construction and reconnect
must preserve explicitly mounted current controls as List does. Bind application
authority to exact page/display/render handles plus the private target owner,
generation, page serial, question/basis and selection revision. Reattaching old
DOM or restoring old component children must not revive source/Keep actions.
Pending selections and publication workflows retain existing refusal/paused rules;
this milestone does not add another selection controller or intent-transfer system.

## Acceptance

Use real OTel spans containing multiple traces, out-of-order source rows, duplicate
parent IDs, missing parents, cycles, zero durations and a cross-service context
span. Prove source-free Table/List/waterfall changes over the exact same page,
readable common-scale geometry, original-index Keep after visual reordering,
qualified missing/empty/null fields, explicit independent Expand trace, paging
with held aggregate/timeline provenance, save/template/reopen and logs refusal.

Generated properties should exercise exact timing, domain/scale bounds,
occurrence permutation and ambiguous/cyclic ancestry. Public SSR/detached and
invalid replacement tests must show no action capability or stale metadata.
Current-render and reattached-handle negative controls, native keyboard behavior,
actual compact frames and a completed external-driver campaign complement source
review. Executed results belong in VALIDATION.md.

## Implementation source review

Reviewed the Go declaration/schema admission and the supplied component/layout and
shared application surface as implementation became available. No architecture
blocker remains in the reviewed boundary:

- Go and JS admit the pinned kind only for spans; schema adds the same signal
  conditional. Existing field budgets and required columns remain. Hidden roles,
  four-role compositions, destination template grouping/baseline/width checks and
  ordinary exact-root composition reuse retain their prior validation.
- Layout uses valid-ID parent candidate arrays rather than single-value overwrite.
  Parent matching includes all candidate occurrences independently of whether their
  timing is available. Invalid trace occurrences get independent group keys and no
  parent links. Iterative traversal retains every original index once, including
  cycles and repeated/duplicate data. Global inspector geometry remains the default;
  the new component explicitly requests trace-relative origins/common scale.
- Invalid timing yields null coordinates and does not enter group origin/extent
  calculations. Unknown-only groups keep null origins/extents. Exact zero remains
  separate from a tiny positive coordinate that rounds during approximate layout.
  The displayed origin is labelled earliest **timed** supplied span, since a row
  can retain a known timestamp but lack a usable duration.
- The public component retains original rich records for explicit Details while
  projecting bounded scalar geometry/IDs. It validates and freezes the page through
  the existing admission path; source scope/token authority is not reconstructed
  from those serialized fields. Service visibility and qualified missing/empty/null
  fields remain separate from geometry. Intrinsic trace-group keys remain visible
  when the optional ancestry disclosure is hidden.
- The shared application surface indexes supplied host handles by their explicit
  original page positions, rejects duplicate/out-of-range/missing positions, and
  binds currentness to the exact component page, display and handle array plus its
  existing private owner. Parent-first DOM order cannot change Keep membership.
- The native timing button becomes the existing Inspect action in the app; local
  passive Details is hidden there. Checkbox/trace/actions mount beside that button,
  not inside it. The default standalone button opens native local Details and has
  no transport/capture/navigation callback. The existing application inspector,
  publication and independent trace context owner are reused.

This is source review rather than an executed browser, keyboard, model or
performance claim. Final outcomes and any fixes are recorded in VALIDATION.md.

Reviewed focused test coverage as well: exact-ns translation/common-scale geometry,
ambiguous/missing/cyclic/invalid identities, unavailable versus known-zero timing,
original-index mounting despite DOM reordering, page custody without totals,
qualified-field state, detached/SSR/reconnect/replacement and service visibility.
Legacy visualization fixtures now use valid OTel-sized IDs while retaining their
prior parent/cycle assertions; invalid IDs are not re-enabled merely to preserve
shorthand fixtures. Positive sub-microsecond duration labels remain distinct from
reported zero while retaining exact raw tooltips.
