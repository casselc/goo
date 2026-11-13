# Guided service-version comparisons

Milestone 415 is delivered. See [validation](../VALIDATION.md), the
[storyboard](journeys/service-versions.md) and
[independent review](reviews/SERVICE_VERSION_COMPARISON_REVIEW_415.md).

From a service's ordinary span overview, open **Compare versions**, explicitly
**Discover versions**, select a baseline and comparison version, then **Stage
version comparison**. The destination uses the existing population workspace;
**Run** acquires fresh evidence. **Back to service** restores the held overview.

Discovery stays inside the original fixed window, exact qualified service and
typed environment policy. Existing role/profile and operation predicates are
preserved. The version dimension is the reported OTel resource `service.version`;
a similarly named span attribute or a deployment receipt is not substituted.
Missing, present empty and literal reported values remain distinct.

The first slice requests one count plus complete version groups. Selection is
available only after a coherent, validated response with conserved group counts.
There is no default pair. The two populations must be distinct and explicitly
selected. Choosing versions and staging are local configuration operations.
They do not acquire telemetry or save evidence.

Count, supported mean duration, recorded Error count and recorded Error fraction
use the existing checked comparison model.
Comparison preserves the original widget's group dimension so corresponding
operations remain comparable. Each population has its own captured count and
source-read basis. Captured records are not request throughput, sampling coverage
or a deployment's full traffic. Mean duration is not an averaged percentile;
Recorded Error is the reported span status, not an HTTP failure classifier or
deployment attribution.

```mermaid
flowchart LR
    S[Exact service overview and fixed span question] --> D[Explicit version discovery]
    D --> A[Validated complete version counts]
    A --> P[Explicit version pair and measure]
    P -->|Stage only| V[Existing population comparison]
    V -->|Explicit Run| Q[Two fresh source populations]
    Q --> C[Exact counts, duration or Error context]
    V -->|Back without source reads| S
```

Owners must cancel departed discovery, refuse late responses and retire obsolete
controls. A disclosure is not a source subscription. Shared display components
receive supplied facts; they do not own query scope, source reads or navigation.
The flow adds no query engine, issue store or compatibility format.

Discovery has its own Read details disclosure: acquisition time, independent
count/group read clocks and bases, exact question, environment and fixed window.
The concise status shows missing and present-empty counts even after two ordinary
versions are selected. Read details opens with native keyboard behavior and does
not query the source. Unrepresentable values remain counted and visibly
unselectable; a failed discovery withdraws old choices and receipts.

## Recorded Error context (419)

Choose **Recorded Error count** or **Recorded Error fraction** in the existing
version measure selector. Both request the typed span-status partition for each
fresh population. Error, explicit OK, unset and unknown counts must conserve the
whole captured population. The shared comparison shows Error/count plus unset
and unknown context. A zero Error numerator is known zero; an empty population
has an unavailable fraction. Fraction change is an exact signed difference,
formatted as percentage points rather than relative percent change.

The selected measurement determines summary directives: Error choices request
`spanStatus: true`, mean duration requests `summary: span-duration`. Discovery
remains count-only. This preserves the original predicates, operation grouping,
qualified service, role, environment policy and window without claiming every
measurement directive is unchanged. Each side keeps independent read facts;
recorded Error does not establish success, health, sampling completeness or an SLO.

Status comparisons are summary-only, use fresh fixed queries, and retain the
existing same/adjacent equal-duration window rules. Stored-baseline status and
per-operation status comparisons remain outside this slice. Diagnostic changes
are withheld. Display/disclosure/Back do not acquire telemetry.

[Source review](reviews/RECORDED_ERROR_VERSION_REVIEW_419.md) and
[executable storyboard](journeys/service-versions.md) document acceptance limits.
