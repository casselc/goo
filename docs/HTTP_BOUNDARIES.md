# HTTP effects and dependency boundaries — C2 active

This inventory describes the current Go server routes, not a new API. Browser
`/api` proxy mapping remains in Python. `serve-http` loads binding/inventory and
operation packages at startup, expands plans, builds one 15-second HTTP client,
registers route families and wraps the mux in `newQueryBudget(..., 4)`. Server
header timeout is five seconds; write timeout is 90 seconds. Pure planning, source
execution, filesystem publication and XTDB transactions retain different owners.

## Endpoint effects

Limits below are request JSON byte ceilings, not result or measurement limits.
Validation may reject smaller requests. Route-specific codes/messages and source
family limits remain authoritative in code/tests; this table groups related routes.

| Routes / owner | Effect | Request ceiling / execution lifetime |
| --- | --- | --- |
| GET `/http-package`, `/operations`, `/counter-rates`, `/histograms` in `serve_http.go` | Execute ClickHouse reads; operation views can issue multiple reads | Query parameters; shared client 15s per HTTP request, incoming context |
| GET `/operation-packages` | Return startup declarations; no telemetry | Query validation, no body |
| GET `/measurement-packages`, `/investigation-workflows` | Read definition/workflow files; no telemetry | Catalog/file bounds, no body |
| POST `/measurement-packages/plan` | Validate/compile native definition; no source or publication | 64 KiB |
| POST `/measurement-packages` | Publish canonical YAML; preserve existing duplicate policy | 64 KiB; filesystem |
| POST `/metric-discovery` | Query native instrument inventory | 64 KiB; 15s context |
| POST `/measurements` | Execute native measurement; lifecycle may create first-observation custody files | 64 KiB; 15s context |
| POST `/inventory`, `/discovery`, `/explore` | Query ClickHouse; Explore also freezes a bounded process-local result-page capability when possible | 16 KiB; 15s / 30s / 55s contexts |
| POST `/compare` | Compare supplied summaries; no telemetry query | 1 MiB |
| POST `/tail` | SSE polling reads; own four-slot budget, bounded stream lifetime and write deadlines | 16 KiB; 60s context, 2s polling, 5s per-write deadline |
| GET `/investigations/{id}` | Reopen immutable operation capture; no telemetry | File bounds |
| POST `/investigations` | Read operation/rate/histogram evidence and publish a capture; refresh checks parent question/package revision | 16 KiB; incoming context and shared HTTP client |
| GET `/captures`, `/captures/{id}` and suffixes `followups`, `compare`, `windows`, `receipt` | Read/page/recompute from saved evidence; no telemetry reread | File/catalog/cursor bounds |
| POST `/captures` | Execute source family and publish immutable evidence; refresh retains original question | 16 KiB; 55s context |
| POST `/captures/live` | Read fresh evidence against a saved baseline and hold expiring drafts, or freeze a held draft into immutable capture files without a reread | 16 KiB; read branch 55s context; draft budget separate |
| GET `/recipes`, POST `/recipes` | Return templates / plan supplied question; no telemetry | POST 16 KiB |
| POST `/dashboards/plan` | Plan from supplied inventory/declarations; no telemetry | 256 KiB |
| POST `/telemetry-packages/draft`, `/numeric-plan`, `/plan` | Build/validate/compile reusable definitions; no telemetry | 64 KiB |
| POST `/telemetry-packages`; GET collection or `/{revision}` | Publish or read reusable definition files | POST 64 KiB; store bounds |
| POST `/population/plan` | Plan typed population transformation; no telemetry | 128 KiB workbench decoder |
| GET/POST `/annotations`, `/exploration-views`, `/deployments`, `/documents`, plus GET `/{id}` | Read/publish authored context; document creation checks references and can verify pinned XTDB review citations | POST 128 KiB; citation checks 30s each |
| POST `/captures/comparisons`; GET `/captures/comparisons/{id}/receipt` | Recompute/save comparison context from saved captures; no telemetry reread | POST 128 KiB |
| GET `/retained-evidence/status` | Report optional XTDB configuration; no database probe | No body |
| GET retained collection, manifest, annotations/history or pinned revision | Read XTDB metadata/custody; await-token lower bounds remain separate | 30s contexts |
| POST `/retained-evidence` and `/{id}/annotations` | Freeze selected saved/server-held rows or correct review metadata in XTDB; request recovery may precede publication; no ClickHouse reread | 16 KiB; 30s contexts |
| POST `/queries/{id}/cancel` in budget wrapper | Cancel/mark a query capability; no source publication | Wrapper validates UUID and capability budgets |

The global query budget currently controls only POST metric discovery, native
measurements, inventory, discovery, Explore, captures and live comparisons. It
adds a 60-second outer context and handles optional query IDs. GET executions,
operation investigations, tailing and retention are not all covered by that
wrapper; tailing retains its own slots. This extraction does not broaden budgets.
A GET method therefore does not establish source-free behavior, and a source read
does not establish absence of filesystem custody effects.

## Decoder, response and stream policies

Most JSON route families disallow unknown fields, cap the body and refuse a second
JSON value. Trailing-value messages differ (`one JSON object required`, `expected
one JSON object`, `require one JSON object`); helpers must retain these choices.
Malformed/body-limit errors normally return 400 rather than universally using 413.
Live snapshot size limits can return 413; capacity can return 429. Saved-reference,
refresh, backend and optional-store failures retain per-route 404/409/410/502/503
policies. The current native success/failure and YAML duplicate response fixtures
are in `testdata/storage` and the native route tests.

JSON/no-store is common, not universal: early tail errors currently set JSON but
not no-store; successful SSE sets `text/event-stream`, no-store and
`X-Accel-Buffering: no`. Some inline GET handlers use `http.Error`, whose content
type/body differ from JSON encoder failures. Default mux 404/405 responses retain
Go's behavior. SSE emits reset/stream events and must preserve flush, cancellation,
continuity caveats and disconnect handling. Do not route streaming through a
buffered JSON responder or silently standardize these policies in a refactor.

## Startup dependencies: native and exploration/tail adoption

`nativeMeasurementSourceFromEnvironment` resolves native ClickHouse endpoint,
credentials and the HTTP client at startup. Empty endpoint retains
`http://localhost:8123`. `registerMeasurements` receives an existing
`measurements.Source` value; handlers no longer read ClickHouse environment
variables. Its custody directory remains derived from the supplied server data
root. Registration copies source strings; the HTTP client remains a shared pointer.
Process environment mutation is not a configuration reload mechanism for these
adopted families. `explorationSourceFromEnvironment` also resolves a typed
`exploration.Source` at startup. `registerExploration` passes that configured value
to inventory, service discovery, Explore and tail registration. Record query
timeouts and the separate tail slots/context/poll/write lifetime are unchanged.

```mermaid
flowchart LR
  Env[Startup environment] --> Wiring[Native source wiring]
  Client[Shared HTTP client] --> Wiring
  Wiring --> Source[Configured measurements.Source value]
  Root[Server data root] --> Register[Register native routes]
  Source --> Register
  Register --> Pure[Definition plan / registry routes]
  Register --> Run[Native discovery / execution]
  Run --> CH[ClickHouse adapter]
  Run --> Custody[Lifecycle first-observation files]
  Pure --> Files[Definition files where requested]
  Env --> RecordWiring[Record source wiring]
  Client --> RecordWiring
  RecordWiring --> RecordSource[Configured exploration.Source value]
  RecordSource --> RecordRegister[Register exploration and tail routes]
  RecordRegister --> Queries[Inventory / discovery / Explore]
  RecordRegister --> SSE[Tail: own slots and streaming lifetime]
  Queries --> CH
  SSE --> CH
```

Inline operation GET handlers, captures/live comparisons and investigations
still resolve ClickHouse environment at request time. Retention
snapshots `GOO_XTDB_URL` on registration, while document citation checks currently
read it at request time. Definitions/workflow paths still depend on working
directory; result snapshots remain process-global and live comparison drafts remain
registration-local. These are explicit remaining boundaries, not unified services.

## Checks and next cuts

Existing native registry/preview/YAML fixtures pass before and after injection.
Two injected backends return different gauge values with distinct credentials;
an ambient trap source receives zero requests. Tests prove injected client use,
source-free planning, native/discovery routing, invalid-input rejection before reads,
request cancellation and absence of gauge/discovery lifecycle files. Startup tests
pin default endpoint and configuration snapshot behavior. Record dependency tests
keep two routers/backends registered simultaneously, proving inventory, discovery,
Explore and tail use their own endpoint, credentials and client. Caller source
mutation does not redirect a registration; an ambient trap receives no reads.
Invalid and cancelled record queries avoid backend reads. Actual SSE reset/batch
frames, no-store/no-buffering headers and disconnect are exercised with injected
sources. The full HTTP race suite
retains budget, SSE, capture and optional-store tests.

C2 remains active. The pure planning adoption below is delivered. Next migrate
capture/live-comparison source dependencies, then operation readers and remaining
optional-store configuration. Bounded ClickHouse
transport extraction follows one family at a time, preserving parameter escaping,
read limits, basis/identity and non-2xx/oversize/cancellation behavior. Python proxy
removal, changing custody/duplicate policy and universal endpoint budgets are not
part of these cuts.

## Pure planning JSON adoption

Recipes and dashboard planning now use `internal/httpjson.Decode` and `Write`.
The decoder takes an explicit byte limit and trailing-input message. It retains
`encoding/json` typed decoding and unknown-field refusal without imposing content
media-type checks, null rejection or duplicate-key rules. The route validates the
decoded domain object after successful decoding; targets can be partly populated
on error and must not be executed/published in that state. Writer behavior is the
existing JSON content type, no-store header and newline-terminated JSON encoder;
encoding/write errors remain ignored as in the original routes. It is not SSE or
a universal error/status mapper.

```mermaid
flowchart LR
  Recipe[Recipes: 16 KiB / require one JSON object] --> Decode[Shared bounded typed JSON decoder]
  Dashboard[Dashboard: 256 KiB / expected one JSON object] --> Decode
  Decode --> Valid[Route-specific semantic validation / pure plan]
  Decode --> Error[Route-specific error response]
  Valid --> Write[Shared JSON / no-store writer]
  Error --> Write
```

`testdata/http/planning.json` pins 32 selected wire cases across the two POST
routes and mux method refusal. It includes valid plans, empty/null/array/truncated
input, unknown top-level/nested fields, second values, malformed tails, trailing
whitespace, exact and over-limit bodies, ignored request content type, empty
objects and 405 behavior. Large bodies are reconstructed from compact fill/padding
fields. Status, exact response bytes and selected headers are compared.

Both fixture sets passed on the old handlers before extraction. Recipes retain
“require one JSON object”; dashboards retain “expected one JSON object”. Exceeding
the body budget during the first value retains the body-too-large error; doing so
while checking trailing whitespace retains the caller's trailing-input message.
Both continue to use 400. Null still reaches domain validation rather than a new
generic “must be object” guard. Default mux text/Allow headers remain outside the
JSON writer.

The Hegel campaign generates Unicode/escaped labels, exact typed int64 values above
JS safe range, whitespace and exact/insufficient byte budgets. It checks roundtrip,
unknown fields, second values and the JSON/no-store response. Existing HTTP race
checks retain pure planning/registry declaration behavior and unrelated native,
SSE, budget, retention and capture contracts. No source or persistence dependency
was added to these planning routes. Other JSON route families are not yet adopted.
