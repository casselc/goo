# Record-intent admission review — 417

Status: accepted for this bounded consolidation. All final reported acceptance
gates pass; no blocking source findings remain. This review is independent of the
production JavaScript and Go decoder authors. Commands actually executed belong
in [validation](../../VALIDATION.md).

## Accepted contract

Ordinary record questions share one pure semantic admission function across
question constructors, links, session staging/execution and scope history.
Predicate and field validation use the same closed supplied-data helpers. The
contract follows Go execution limits: 32 predicate nodes, depth four, 128-byte
qualified field keys, 1,024-byte string values, page limits 1–100, offsets 0–1,000,
and fixed windows up to 120 minutes ending no later than the Go 2100 bound.
UTF-8 byte counts, signal-specific fields/operators, typed numeric thresholds,
duplicate grouping, chart interval limits and reported-edge namespace ownership
are checked before intent is adopted.

Go remains authoritative for execution. Its query decoder now refuses unknown
or case-aliased fields, explicit null arms, malformed Unicode and mixed predicate
branches before typed zero values can erase supplied meaning. Duplicate object
members follow the browser's last-value semantics rather than merging nested
structs. Semantic validation remains a separate explicit step after decoding.
There is no legacy-format repair or new query/store abstraction.

History retains a valid nonzero offset and page limit. Recall no longer silently
turns a continuation into page zero; the picker identifies the starting record.
Actions that deliberately select a new population may still explicitly construct
page-zero questions. Link/history storage budgets and editor affordances remain
separate from whether an ordinary question is semantically valid.

One final source correction aligns identity comparison with Go's optional offset:
for ordinary log/span questions only, an omitted offset compares as zero. The
comparison uses a clone and does not insert defaults into authored configuration,
links or stored intent. Explicit nonzero offsets remain distinct. This is a known
Go wire-default comparison, not a query rewrite or a compatibility format.
The same comparison clone removes an empty reported namespace `value` member
when Go omits it. The explicit namespace mode is retained, so present empty and
missing namespaces remain distinct. The final unit check covers that distinction
and leaves the authored question unchanged.

## Review findings and remediation

- Nested environment objects originally used a weaker local shape guard. They
  now share descriptor, prototype and symbol checks with the rest of the query.
- History serialized before validation and could execute getters or `toJSON`.
  It now closes the envelope, selector, layout arrays and fields before copying
  or serializing admitted intent. Its view default applies only to omission;
  explicit null/empty views are not silently repaired.
- The execution owner revoked held actions before complete query admission.
  `run` now checks the shared contract first; `Session.start` and `Session.stage`
  also refuse invalid questions before changing lifecycle state or emitting an
  event. Invalid link restoration reports refusal while retaining held results.
- `recordStep` and the link constructor originally cloned before admission.
  Both now validate before environment projection or cloning. Supplied accessors
  are refused without evaluation, including through the history boundary.

These blocking findings are resolved in the source inspected for acceptance.
The new unit tests explicitly check accessor non-evaluation and unchanged applied
session state/events after refused starts and stages. No remaining source blocker
was found in the reviewed ordinary-record path.

## Final acceptance evidence

The parent reports the final full frontend suite passing **1,575 tests with no
skips**, all four affected Go packages passing, and API validation covering
**114 operations, 759 schemas, 501 fixture responses, 202 admitted requests and
29 mux refusals**. This reviewer prepared the independent legality corpus and
inspected the production remediation; heavyweight checks were run by the parent,
not repeated here.

The final served real-source journey and **35-second Bombadil campaign pass in
45.5 seconds**. They retain exact Unicode-boundary intent through links and
durable history, explicitly stage, then Run once. The source reports **140
matching records with a separately bounded 100-row page**. Eleven refused
handoffs preserve those held rows without an extra POST, and a later valid stage
withdraws them without acquiring a source. The final primary question deliberately
omits offset, exercising adoption of Go's explicit zero rather than only testing
that comparison in isolation.

The full native model/property package passes in **15.450 seconds with no skips**.
It includes the independent 32-seed Go/JavaScript admission differential, the
actual five-fact DOM phase/held/API ledger and rejecting implicit-read and
lost-held-state mutants. The root supplied the actual completed browser report;
this acceptance does not rely on a skipped observed-ledger check.

The reviewer independently inspected the final 390px refusal screenshot. The
applied-question count, existing rows, valid rerun/paging controls and explicit
refusal message coexist clearly, without discarding or replacing the held result.
This is a bounded visual check of that surface, not whole-product accessibility
or cross-browser certification.

## Scope limits

This is ordinary-record configuration consolidation. It does not certify every
measurement/trace-search constructor or every application persistence format.
A valid query is not a claim that telemetry is complete or healthy, and a
source-free stage is not execution. URL-size, browser-storage and endpoint body
budgets may still refuse an otherwise valid question; refusal must remain
explicit. The reviewed code introduces no schema migration, compatibility floor,
source write, background read or issue lifecycle.
