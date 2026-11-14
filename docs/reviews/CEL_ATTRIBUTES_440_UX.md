# Attribute conditions: UX and observability review

## Contract

The accepted vocabulary uses two explicit maps: `resource.attributes` and
`record.attributes`. The latter resolves to the selected signal's log or span
attributes. Keys are literal strings, including dotted semantic-convention keys;
there is no implied nesting, convention fallback or cross-map search.

The bounded map-member helpers `.eq(key, value)`, `.ne(key, value)` and
`.contains(key, fragment)` all require a reported key. Standard literal
membership and negated membership express present and missing. A reported empty
string remains distinct from an absent key. Missing-inclusive comparisons require
an explicit missing branch. General negation and bare indexing refuse rather
than silently changing the present-only visual predicate semantics.

The standard guarded-index alternative was considered. Absorbing a presence
guard into a comparison would introduce a structural normalization rule solely
to make the source expression fit the existing AST. The accepted helpers instead
name that present-only comparison directly and permit exact inverse printing.
Their ordinary map receiver and literal arguments avoid a user-facing optional
value/type protocol.

Exporter maps are strings. No numeric parsing or original SDK type inference is
introduced. Existing typed numeric conditions remain a separate explicit
contract; semantic-convention naming does not prove a value's physical type.

## Interaction expectations

Reuse the existing contextual textarea, Check action, located diagnostic and
Apply/Run owner. Do not introduce a second query editor, execution button or
persistent toolbar before the stream. Examples should be short, copyable and
available on demand. One concise missing-versus-empty explanation is sufficient;
source/readiness and compiler implementation details do not belong in the flow.

A successful source-free check stages the ordinary additional condition. Invalid
text remains recoverable. Mode switching and parent changes follow the existing
revision owner, not a new attribute-specific state machine. Saved definitions
retain checked source through their existing provenance owner.

## Acceptance status

Source review accepts the qualified map-member overloads, literal presence and
negated-presence lowering, reference map omission for missing keys, and direct
attribute-leaf printing. The shared editor adds only a native collapsed
**Attribute examples** disclosure with two selectable code snippets and one short
missing-versus-empty sentence. Component-scoped wrapping keeps the snippets
within the existing layout; no additional toolbar, execution or copy action is
introduced. All **14 focused editor tests** pass, including local disclosure and
preserved diagnostic association.

Independently viewed the actual application frames at **1440, 390 and 320 px**
under `docs/journeys/screenshots/440-attribute-condition-*.png`. The native
Examples disclosure stays within the existing editor, has a visible keyboard
focus outline and keeps selectable snippets readable without horizontal
overflow. Both mobile widths preserve the textarea label, existing Check/Clear
actions and staged-condition feedback. Examples remain optional rather than
occupying a permanent toolbar.

No blocking UX or accessibility-source finding remains for this bounded change.
The parent-owned real log/span journey and native verification report source
execution and reference semantics separately; this review does not claim a
physical screen-reader or cross-engine test.
