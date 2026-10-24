# Shared staged filter editor

Activity's **Filters**, the single-question **Filters** disclosure and fixed
population **Edit logs/spans filters**
use `GooFilterEditor.mount`. Start with a known field even when no matching record
has appeared. Choose Column, Attribute or Resource, enter its key, choose an
operator and value. Column keys have signal-specific browser suggestions;
attribute/resource keys remain explicit, including OTel convention keys.

- Groups support AND/OR, nested within the existing Go budget of 32 nodes and
  depth four. Removing the last child removes its group. Clear filters removes
  the selector, preserving environment and time ownership.
- Equals with a blank value means **present empty string**. Missing checks absence;
  not-equal does not include missing keys. Add an OR missing condition explicitly
  when that is the intended population.
- String fields support equals/not-equals/contains/presence. Native duration and
  timestamp columns additionally support unsigned decimal ordering; values remain
  strings through transport, preserving uint64 precision. Duration/timestamp
  predicate literals use nanoseconds, explicitly labeled in the editor. This cut
  does not introduce numeric attribute conversion or transforms.
- Editing pauses Activity and preserves its held observations. Apply scope starts
  a new recent read. Invalid drafts block Apply, Resume, Save and fixed analysis.
  Saved Activity configurations restore the complete tree and remain paused until
  explicit Start. Changing environment preserves editor input and validation.
- Fixed editing preserves the configured parent. Open edited population stages a
  child with that signal's selector replaced; companion selectors, grouping,
  measures, environment and UTC bounds remain. Run selected population performs
  the reads. The parent stays paused. Fixed view publication saves configured
  queries; editor drafts belong to their disclosed edit controls until opened.

```mermaid
flowchart LR
  Config[Saved or current selector] --> Editor[Shared supplied AST editor]
  Editor --> Validate[Typed draft validation]
  Validate --> Activity[Activity draft / held rows paused]
  Validate --> Single[Single question draft / fixed bounds]
  Single -->|explicit Run| Query
  Validate --> Child[Fixed child configuration / parent paused]
  Activity -->|explicit Apply or Start| Tail[Existing tail owner]
  Child -->|explicit Run| Query[Existing fixed query owner]
  Tail --> Source[ClickHouse OTel records]
  Query --> Source
```

The helper owns DOM/input state and emits detached normalized predicates or a
validation error. It performs no fetch, navigation, subscription or publication.
Application adapters own execution, drafts and custody. Native Go validation
remains authoritative. Empty `value` is omitted exactly as Go JSON transport does;
its operator retains empty-versus-missing meaning. Existing saved-view hashes and
query API schemas are unchanged. The single-question builder now uses the same editor. Service/message/duration
shortcuts contribute visible predicates to the tree; editing that tree consumes
the shortcuts into its full AST and clears their inputs. Later shortcut entries
add explicit AND conditions. Refining an OR population adds a conjunct around the
whole existing tree, preserving its Boolean meaning. Environment edits preserve
filters and fixed bounds; signal changes reset incompatible predicates. Invalid
drafts block Run, sharing and source actions without discarding partial input.
Shortcut inputs pause during invalid tree edits so correcting the tree cannot
overwrite a second, contradictory shortcut draft. Environment remains editable.
Source-supplied field suggestions are passed through `setFields` without fetching;
refreshing suggestions preserves input and does not emit an edit.

Executable storyboard: journey 148 sends three ordinary OTel logs (missing,
empty, canary), verifies ingestion independently in ClickHouse, edits a saved
selector without a source read, checks invalid/keyboard/mobile behavior, applies
it, saves/reopens the exact AST, then edits a fixed continuation before explicit
Run. Bombadil explores staged group changes. Observed Quint replay checks the
source-free staged configuration, including existing rejecting controls. Hegel
checks missing/empty membership through sampled nested JSON round trips.

![Desktop staged filter editor](journeys/screenshots/148a-filter-editor.png)
![Mobile staged filter editor](journeys/screenshots/148b-mobile-filters.png)

Journey 149 covers the single-question adoption: nested missing/empty edits,
invalid field preservation across environment edits, source-free shared-link
reload, keyboard Run, exact observed records, supplied field suggestions and
fixed-window refinement. Its Quint replay rejects hidden reads, mismatched
predicates, changed windows and changed environments. Bombadil explores supplied
groups and invalid fields without running a source query.

![Single question editor](journeys/screenshots/149a-single-question-filters.png)
![Mobile single question editor](journeys/screenshots/149b-mobile-single-filters.png)


## Observed fields and values

Open **Observed fields and values** to search qualified keys in the applied scope.
Each field offers **Is present**, **Missing key**, **Present empty string** and
recorded literal values where supplied. Choosing one appends a condition with AND
around the existing selector, then stages it. The field/value datalists also use
supplied observations; manual entry stays available.

No suggestion query is added. Activity derives observations from its held buffer;
single questions and fixed populations reuse their successful result's field
metadata and row page. Saved configurations reopen without observations. Start
or Run explicitly to supply them. The label states environment, signal and fixed
UTC bounds, or the held buffer's latest read window. A live buffer can contain
observations from multiple read windows; its label does not bind all rows to the
latest window. Suggestions are withdrawn on scope edits and source replacement; failed fixed
reads leave them empty. Activity may retain its explicitly labeled held-buffer
observations while paused or disconnected. Existing generation guards reject obsolete responses before presentation.

`GooAttributes.suggestionData` scans at most 100 supplied records, 64 qualified
fields and 12 distinct values per field. Limits are disclosed. Values are shown
samples, without population counts, ranking, cardinality or completeness claims.
An absent key is suggested only when the relevant attribute/resource map is
actually present. An unavailable map or a present null/non-string value is not
coerced into missing or a string. Column types still follow the Go grammar.

The editor accepts detached `setSuggestions` data and never fetches. Updating
observations refreshes datalists without changing the typed tree or emitting an
edit. Unchanged field/value lists retain their disclosures and focus during live
updates. Search is local. Adapters clear scoped observations rather than carrying
suggestions from another environment, selector, signal or saved revision.

```mermaid
flowchart LR
  Tail[Existing Activity read] --> Held[Held records / applied selector]
  Fixed[Existing fixed read] --> Page[Successful page / scoped fields]
  Held --> Bound[Bounded observation adapter]
  Page --> Bound
  Bound --> Browse[Shared observed fields and values]
  Browse -->|choose literal / empty / missing| Draft[Draft predicate]
  Draft -->|explicit Apply or Run| Read[Existing source owner]
  Edit[Scope edit / failed read / reload] --> Clear[Withdraw scoped suggestions]
```

Journey 150 covers actual OTel/ClickHouse populations, keyboard/mobile selection,
source-free search/selection, saved reopen, fixed child staging, obsolete-response
refusal and a negative transport fault. Its actor campaign explicitly reads once,
then explores and stages without executing another query.

![Observed Activity fields](journeys/screenshots/150a-observed-activity-fields.png)
![Mobile observed values](journeys/screenshots/150b-mobile-observed-fields.png)

Observed sample fields and typed conditions now offer [explicit counted value lookup](FILTER_VALUES.md). This separate read covers the applied fixed scope; choosing a counted entry stages a predicate, while sample browsing and typing remain source-free.

[OTel field hints](OTEL_FIELD_HINTS.md) now enrich supplied observations, typed-field suggestions and counted headings with readable names and local meaning-based search. Exact keys, legacy spellings, staged actions, scope invalidation and sample limits are preserved.

[Observed-field grouping](OBSERVED_GROUPING.md) now stages exact one-field breakdowns through owner callbacks. Activity freezes its completed applied window; fixed questions stage a draft; population views stage children with companions intact. Choosing a field does not execute a query.
