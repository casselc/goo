# Explicit counted field values

**Load values for this scope** runs a separate, explicit breakdown read for a
chosen field. It is available beside an observed field and an editor condition.
Apply/Run the main scope first. Browsing suggestions, searching loaded values and
selecting a counted value never execute the main question automatically.

Shown-record samples and counted results stay separate. Samples cover at most
100 shown records; a counted lookup covers all matching records in its fixed
query, within the existing grouping limits. Missing keys, present-empty strings
and literal values have separate entries and exact decimal counts. The count list
is ordered by count, with local search and bounded scrolling. Long literals are
abbreviated visually; values outside predicate budgets cannot be staged.

The application owner preserves environment, signal, selector, reported edge and
UTC bounds, and requests one grouping field through `POST /api/explore` with a
one-row inspection page. Presentation grouping, offsets and aggregate measures
are omitted. This is an ordinary record-count breakdown, not a metric conversion
or a new query language. No new backend route or saved format is added.

- Activity pauses its actual subscription before binding the latest completed
  read window. That window remains fixed for the lookup. Resume/replacement,
  mode changes and scope edits invalidate counted results.
- Single questions use their successfully applied scope. Existing page evidence,
  charts and capture identity remain separate from the lookup's read.
- Fixed populations own one lookup per signal/editor. Opening a child or leaving
  the view cancels pending lookups; returning retains working controls. Editing
  stages a child before Run. Saved configuration reloads contain no counted data.
- Selecting a value adds an AND condition around the complete current selector.
  It stages `missing` or exact `eq` (including present-empty); it does not select
  physical rows, establish membership proof or publish retained evidence.

Go's existing grouping implementation refuses more than 1,000 groups and applies
its bounded source responses. The UI never presents the first 1,000 as a complete
result, invents a residual bucket or uses the one-row page as a denominator.
Independent count/group reads must agree; changed/incomplete reads are refused.
Counts retain uint64 precision and their source read clock/basis. Matching counts
are not a transactional snapshot or a claim of complete telemetry coverage.
Failures clear counted rows and allow explicit retry. Scope/generation checks
reject late responses, including failures from a departed scope. Nothing is
fetched by the display element, disclosure or search input.

```mermaid
sequenceDiagram
  participant User
  participant Editor as Shared editor
  participant Owner as Applied-scope lookup owner
  participant Go as Existing Go exploration
  participant CH as ClickHouse
  User->>Editor: Load values for this scope
  Editor->>Owner: Supplied qualified field
  Owner->>Owner: Pause live work and bind fixed scope
  Owner->>Go: POST explore with groupBy field / limit 1
  Go->>CH: Existing count, bounded row, group and field reads
  CH-->>Go: Typed nullable groups and exact counts
  Go-->>Owner: Validated result or grouping refusal
  Owner->>Owner: Check scope, generation and count conservation
  Owner-->>Editor: Supplied counted state / read basis
  User->>Editor: Search or choose a value
  Editor-->>Owner: Stage predicate; invalidate lookup
  Note over User,Owner: Apply/Run remains a separate explicit action
```

Journey 151 sends 103 ordinary OTel logs and verifies that a lookup finds values
outside the first 100-row sample. It covers exact groups, keyboard/mobile use,
local search, staged selection, late-response refusal, transport failure/retry,
saved fixed children and paused Activity windows. Journey 152 ingests 1,001 unique
values and verifies a real source refusal without partial totals. The actor
campaign explicitly runs and loads once, then stages without another query.

![Counted values beyond a page sample](journeys/screenshots/151a-counted-field-values.png)
![Mobile counted values](journeys/screenshots/151b-mobile-counted-values.png)
![Grouping limit refusal](journeys/screenshots/152-counted-value-refusal.png)
