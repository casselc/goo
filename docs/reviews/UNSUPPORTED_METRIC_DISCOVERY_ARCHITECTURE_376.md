# Unsupported metric discovery — architecture 376

Status: delivered; source review, full Go/API checks, generated transport
properties and native browser/actor acceptance pass. This is discovery-only exponential histogram visibility, not execution.

## Existing gap and reusable boundary

`measurements/discovery.go` enumerates gauge/sum/explicit histogram metadata with
one global 10,001-row sentinel limit. Its instrument DTO already distinguishes
supported execution, an optional Definition and unsupported reasons. 374 verified
13 actual exponential points but those observations are invisible to this union.
The right extension is a distinct metadata family, not conversion to explicit
histograms or a new evaluator.

## Optional table availability is explicit

Use a fixed `EXISTS TABLE otel_metrics_exp_histogram` probe against the configured
source connection/database. Admit exactly one expected row containing a canonical
0 or 1 result. Authentication, timeout, malformed/extra/empty rows and other errors
fail the complete discovery request; none is interpreted as table absence.

A valid 0 allows existing families to be discovered and adds the sole explicit
missing-table diagnostic below. It means the optional table is unavailable to
this connection, not that no application exports exponential metrics. A valid 1
includes that family in the same globally bounded UNION. A later table-race or
query failure propagates; never silently retry without the family. Present table
plus zero matching rows does not imply ingestion completeness or health.

The result adds `diagnostics: []DiscoveryDiagnostic`, always an array:

```
{kind: "exponential-histogram", reason: "optional-table-missing"}
```

The array is empty when the optional table is included. It does not become a
registry, source-health state machine or evidence of complete capabilities.
Pre-query invalid requests remain separate from failed source operations.

## One bounded metadata population

All included families share **one** final LIMIT 10001 sentinel and 10,000-row
admission; no per-family limits, top-N catalogue or independently truncated page
may hide overflow. Probe rows are not discovery observation records. Preserve the
cloned fixed input question, checked typed environment policy, service, namespace
presence and from/to bounds across every branch. Metadata selection does not add
or broaden execution scope. Table name and SQL structure are fixed, not user input.

Exponential instruments use distinct kind `exponential-histogram`, preserving
reported scope/version/schema/unit, temporality, dimensions and full stream
identity. They always have supported=false and no generated Definition. The
capability envelope is descriptive/ineligible; familiar duration names or seconds
units cannot turn them into explicit-histogram candidates. Typed-policy fallback
reasons must not mask the specific unsupported representation reason.

The UI may display these rows and their checked metadata using current presenters,
but Preview/Run must remain unavailable. Records count source point rows, not the
sum of instrument observation counts, unique requests or invocations. No bucket
geometry, percentile, interval or exemplar relationship is inferred from this read.

## Source fidelity remains independent

374's pinned lab exporter loses zero threshold and optional Sum/Min/Max presence;
general representation also needs exact irrational boundaries, negative closure
and zero-region handling. Discovery must not invent those values or reopen the
executor gate. Conversely, a metadata read against an arbitrary connection does
not prove it uses the pinned exporter: unsupported text should identify Goo's
execution/representation limit and avoid universally asserting that every table
has the lab's exact losses. Table presence alone establishes no conformance.

## Required verification

Test the actual source boundary for probe 0/1, malformed/multiple/empty results,
authorization/timeout/query/race failure, one combined cap and unchanged supported
families. Independently check whole typed question/service/namespace/time
membership, distinct metadata identity and unsupported/no-Definition invariants.
Use the retained genuine 374 observations for discovery and UI evidence; compare
raw point counts without treating large observation totals as metadata rows.
Missing-table UI must show the diagnostic rather than a falsely complete empty
catalogue. Verify disabled and saved stale Preview actions cannot execute.

Schema must reflect Diagnostics and the expanded discovery-only kind while leaving
measurement Definition kinds, native view guards and executable plans unchanged.
A single ordinary read still uses the existing source owner/cancellation boundary;
no new registry, backend, query engine or compatibility alias is justified.

## Candidate source review

The candidate clones/validates before probing, propagates probe/query errors
without fallback, and includes the optional branch under one global sentinel cap.
Discovery and pure capability planning both keep exponential metadata ineligible;
no Definition is created, and the typed-policy generic reason cannot replace its
specific unsupported-representation reason. The fixed reason describes the
current adapter/kernel limit rather than claiming all source versions lose the
same fields. Diagnostics is an explicit array, not inferred from row count.

One small canonical-probe correction was requested: decoding into an integer
accepts JSON `-0` as zero. The agreed exact 0/1 protocol should compare the trimmed
raw token against those two literal tokens, rejecting numeric aliases/null/strings.
No broader JSON codec or registry is needed. Subject to that correction, source
review finds no material authority or execution blocker; runtime and test gates
remain separate.

## Canonical probe correction closed

The final source compares the trimmed raw token against literal `0` and `1`.
Focused adapter and generated transport checks reject numeric aliases and malformed
answers. Final full Go and API checks pass; native UI acceptance remains separate.
No duplicate-key codec or broader source-capability registry was introduced.
