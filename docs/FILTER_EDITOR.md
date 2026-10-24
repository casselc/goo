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
