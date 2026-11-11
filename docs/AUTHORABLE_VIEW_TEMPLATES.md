# Author reusable views

The category-bound arm below is complemented by [ungrouped record/summary
templates](UNGROUPED_SUMMARY_TEMPLATES.md) (404), with an explicit `group: none`
binding and no aggregate or baseline. Both arms require page-zero roots.

Use the existing view composition as the canvas: choose bars or a table, its
measure, record columns, an optional [fixed timeline](TIMELINE_VIEW_TEMPLATES.md),
and a [declared facet](DECLARATIVE_FACETS.md).
The records child can use Table or
[List](RECORD_LIST.md) over the same supplied page. Under Preserve, name the view, enable **Reusable template**, choose
its **Data profile**, and Save. Templates use the same immutable
saved-view catalog; they do not create another dashboard or package store.

```json
{"template":{"version":"goo.template/1","group":"destination","profile":"otel.db-client"}}
```

This optional declaration belongs to an ordinary single-query aggregate-and-records
view, optionally retaining its fixed timeline and either a literal facet or a
[comparison with an explicit destination baseline](COMPARISON_VIEW_TEMPLATES.md). The title names the template, and its composition owns presentation. The
saved query is a validation specimen, never a default service, environment, time
window or predicate for another investigation.

## Preview in another scope

Open an ordinary grouped population in the same signal. In Saved exploration views,
choose **Inspect template**. Preview identifies the concrete destination grouping,
profile requirements, kinds, measure and record columns. **Apply template** applies
configuration locally, keeping the target question and held data. Run remains an
explicit independent read, and Save creates a new immutable configuration.

The category binding is explicit: it uses the destination's existing qualified
single grouping, even when that field differs from the specimen's field. It never
adds a grouping, weakens a predicate, computes a new clock or borrows source rows.
An omitted aggregate display title follows the current grouping and measure;
an explicit title remains literal, including text that matches a shipped preset.
The renderer always shows the actual qualified grouping and measure separately.
The template/catalog title names the reusable definition, not its destination field.
A facet has a different binding: its qualified field, title, applied-scope basis
and visibility stay literal. Destination validation checks that field but never
rewrites it to the category. Loaded values and selections are not copied.

Authored expansion refuses nonzero delivery offsets; it does not rewrite them.
Applying in the UI requires a page-zero
configured root; ordinary record-page inspection keeps that root and remains
supported. Ordinary saved-composition reuse
continues to require its existing exact grouping compatibility.

## Profiles and limits

| Profile | Checked applicability |
| --- | --- |
| Ordinary records | Same signal, one ordinary nonnumeric grouping, supported count/duration display |
| OTel database client spans | Spans with exact nonempty service, reported Client kind and nonempty current `db.system.name` |
| OTel HTTP server spans | Spans with exact nonempty service, reported Server kind and nonempty current `http.request.method` |
| OTel current RPC server spans | Spans with exact nonempty service, reported Server kind and nonempty current `rpc.system.name`; `rpc.method` is conditional |

Mandatory profile conditions must be explicit conjuncts; appearances beneath OR
cannot establish them. Selecting a profile validates the current question and does
not add filters. Recorded spans do not establish database/HTTP/RPC health or complete
upstream telemetry. Fixed record fields retain their qualified keys; a declaration
does not prove those attributes were observed.

The original 296 template supports aggregate and records; 299 adds the existing
fixed timeline with an explicitly matching destination interval width and supported
count/duration measure. Preview names both required and current widths; correction
uses the existing interval controls and explicit Run. Hidden timelines retain the
same compatibility requirements. 300 adds the existing count-comparison child
with an explicit destination-baseline binding/input; its stored evidence stays
unverified until existing preparation checks it. 311 adds literal facets beside
aggregate and records, including in the shipped HTTP methods/route view. Facet and
comparison children remain mutually exclusive; hidden controls retain all
validation requirements. Numeric/histogram, reported-edge
and grouped-time profiles, tails and nonempty local group Find are refused. Nothing is silently removed. A duration
display also requires that measure in the destination question. Shared kinds retain
their own units, support and unknown-data semantics.

The shipped **Database client activity** preset starts from the observed database
operations widget, grouped by `db.operation.name`. It shares profile validation and
expansion with the HTTP preset and authored templates. A supported materialized
HTTP view with numeric bindings may still open, but cannot opt into this first
ordinary reusable-template profile.

The [Current RPC server methods](CURRENT_RPC_VIEW_TEMPLATE.md) preset uses the
existing current-RPC dimension and a method facet. Its `otel.rpc-server-current`
binding preserves missing, empty and `_OTHER` method groups and requires the
current `rpc.system.name` witness. Authored reuse binds presentation to the full
destination scope; legacy-only RPC metadata or Client spans are incompatible.

## Configuration-only expansion

The existing endpoint accepts one exclusive request arm:

```text
POST /api/view-templates/expand
{ "source": <authored ExplorationView>, "query": <explicit target Query> }
```

Comparison templates additionally require an explicit `baselineCapture` request
input and `template.baseline: "destination"`; noncomparison and shipped arms forbid
that input. Only the chosen baseline reference is substituted, with no capture
fetch or source fallback. See [the binding contract](COMPARISON_VIEW_TEMPLATES.md).

The shipped arm remains `{template, version, widget}`. Expansion compiles the
question, checks source/template and destination compatibility, and returns detached
configuration. Equivalent Go wire defaults (omitted false flags or empty predicate
values) do not change question identity; equality-to-empty remains distinct from
missing. Presentation and binding checks remain exact. Expansion performs no
telemetry read or publication. A supplied source ID
must match its content; an unsaved draft can receive a content digest without
being published. Neither is source evidence or execution authority.

```mermaid
flowchart LR
  C[Existing composition canvas] --> T[Explicit template binding and profile]
  T --> S[Save in existing view catalog]
  S --> P[Inspect against current target question]
  Q[Destination grouping and scope] --> P
  P -->|Compatible: explicit Apply| A[Local target composition]
  A -->|Run| R[New independent telemetry read]
  A -->|Save| V[Reusable fixed target view]
```

See the [executable storyboard](../journeys/296-authored-template-storyboard.md),
[architecture review](reviews/AUTHORABLE_TEMPLATE_ARCHITECTURE_296.md) and
[design review](reviews/AUTHORABLE_TEMPLATE_DESIGN_296.md). General expressions,
team access control, new contributed kind implementations and WYSIWYG layout are
separate work; this is a bounded declarative template over existing kinds.
