# Goo operation and event annotations

The first annotation slice is implemented with custom Orchestrion aspects, a
small isolated OTel runtime module, a source-contract manifest generator and a
lint-before-test `goo weave` command. Aggregate policy remains in goo packages.

## Supported declarations

```go
//goo:operation name:checkout.reserve role:server
func Reserve(ctx context.Context, fail bool) (result *Decision, err error)

//goo:event name:reservation.accepted subject:result.ID when:ok
func Accept(ctx context.Context, fail bool) (result *Decision, err error)

//goo:event name:reservation.terminal subject:decision.ID outcome:decision.Outcome when:ok
func Complete(ctx context.Context, decision *Decision) (err error)
```

See the executable [example](../verification/weaving/example/checkout.go).
Operation roles are explicitly `internal`, `server`, `client`, `producer` or
`consumer`. This selects span kind; it does not synthesize HTTP/RPC/messaging
semantic-convention attributes or assert a network protocol.

One goo directive per function is supported. Functions need exactly one named
`context.Context` parameter. A returned builtin `error` must be the single named
error result; the aspect observes its value on return. Subject/outcome expressions
are restricted to parameters, named results and field selectors. Calls, indexing,
unknown names, duplicate/unknown fields and unsupported directives fail lint.
Go compilation verifies selector result types: subject/outcome must be strings.
Use an explicit named builtin `error` result rather than custom error-return
conventions; this first slice does not resolve error policies through aliases.

## Emitted contract

Scope: `github.com/casselc/goo/instrumentation`, version `0.1.0`, schema
`goo://instrumentation/v1`. These are custom instruments, not standard HTTP
request metrics.

| Signal | Kind / unit | Dimensions and meaning |
| --- | --- | --- |
| `goo.operation.calls` | Counter / `{invocation}` | Operation name, role, outcome (`ok`, `error`, `panic`); incremented on completion/unwind |
| `goo.operation.duration` | Histogram / `s` | Same dimensions; elapsed monotonic time, with explicit second-based bucket boundaries |
| `goo.operation.active` | UpDownCounter / `{invocation}` | Operation name/role; +1 on entry, −1 on completion/unwind |
| `goo.instrumentation.dropped_events` | Counter / `{event}` | Declared event name and bounded drop reason |

An operation creates an OTel span and passes its context into the function body.
The finalizer is idempotent. Error results and panics produce their respective
outcome and span error status. Panics are rethrown; return values/error values
are preserved. Raw error messages, panic values and subject IDs are not added to
metric dimensions. Error type is recorded without calling the error's message.
The panic path is tested under the default modern Go panic/recover behavior;
legacy `GODEBUG=panicnil=1` and `runtime.Goexit` are not classified by this pilot.

An event is added to the current recording span after normal return with nil
error. Its attributes include `goo.event.name`, `goo.subject.id`, and optional
`goo.subject.outcome`. There is no synthetic subject count or business success
ratio. A missing/oversized subject, invalid selector/outcome, or absent recording
parent yields a dropped-event counter observation rather than an invented event.
Selector access is guarded, so a nil result field does not introduce a new
application panic. The runtime does not install providers or exporters: configure
standard OTel providers before first use. Sampling can remove span events even
when operation metrics are emitted.

`when:ok` means the function returned without an error. The application author
owns what the function promises. This does not establish transaction commit,
durable acceptance or terminal business completion. Subject IDs are opaque until
a package explicitly binds tenant/environment/business-key scope. Retry attempts
must not silently become a subject-level denominator.

## Manifest and opt-in weaving

```sh
go run ./cmd/goo manifest -dir verification/weaving/example

go run ./cmd/goo weave \
  -module verification/weaving -dir example -test ./scenario \
  -manifest artifacts/annotation-manifest.json
```

Use the configured Go toolchain; keep `node_modules/.bin` on PATH for the existing
Quint check in the scenario suite. Optional retained evidence:

```sh
export QUINTGO_BACKEND=typescript
export QUINTGO_KEEP="$PWD/artifacts/weaving"
export GOO_ANNOTATION_ARTIFACTS="$PWD/artifacts/annotations"
```

The generated manifest records source digests, declaration identities, observation
units, custom signal names/types/units/dimensions and runtime scope/schema. Its
content revision is deterministic and independent of the checkout's absolute
path. It describes **one source package**, not the actual linked binary: build
constraints, module/dependency versions and enabled instrumentation across all
compiled dependencies are not resolved. Manifest generation neither executes the
package nor establishes runtime completeness.

The initial `goo weave` command is **test-only**. It requires an explicitly opted-in
module with `orchestrion.tool.go`, validates the selected package, writes the
manifest and executes Orchestrion via Go without editing module dependencies.
It does not install a tool, build a deployment artifact, or verify that every
other compiled dependency was linted. Direct Orchestrion users must run goo lint
before weaving too; the aspect templates assume the validated dialect.

The example module opts in with:

```go
//go:build tools

package tools

import (
    _ "github.com/DataDog/orchestrion"
    _ "github.com/casselc/goo/instrumentation"
)
```

Its extra quintgo import enables the separate model-observation pilot. Goo's
aspects do not enable Datadog tracing integrations. The `instrumentation/` module
contains only OTel API dependencies; the root compiler/query module still has no
OTel dependency. The optional weaving module pins the already-validated
Orchestrion/x-tools combination documented in [INSTRUMENTATION.md](INSTRUMENTATION.md).
Plain Go builds ignore the annotation comments.

## Executable developer storyboard

| Step | Actual check |
| --- | --- |
| Describe source observations | `goo manifest` validates operation/event declarations and emits a versioned contract |
| Weave the example | `goo weave` runs custom Orchestrion aspects on real Go functions |
| Exercise success | The original decision returns unchanged; accepted and terminal events land on the recording operation span |
| Exercise failure | The original error returns unchanged; no success event is emitted from a nil result |
| Exercise panic | The original panic value propagates; metrics record panic and release the active operation |
| Exercise missing event evidence | Nil/empty subject and no recording parent produce three diagnosed drops |
| Compare emitted OTel with the manifest | Real OTel SDK in-memory export verifies scope/version, units, dimensions, metric counts and span events |
| Explore operation lifecycles | Hegel state/swarm sequences start, finish, repeat finalization, error and panic outcomes; active count equals unfinished invocations after every step |

The SDK capture is genuine woven telemetry retained at
`artifacts/annotations/otel.json` when requested. It is not a collector/ClickHouse
integration journey; no screenshots or production auto-instrumentation coverage
are claimed. Existing real browser journeys are unchanged. The existing woven
HTTP preparer → Quint retention model also remains passing, including its
negative control. A formal operation-lifecycle/Goexit/concurrency model is not
claimed by the state/swarm check.

## Aggregation and next steps

Invocation count, outcome dimensions and duration observations can support a
reusable operation package. Event observations can support a separately scoped
business-event contract. The manifest does not define targets, windows,
notification policies, histogram merging or counter rates. Those operators and
source bindings belong in the package/compiler model, with temporality and
coverage checks. Never sum exported cumulative points or average scalar p95s.

Next: bind manifest declarations to observed inventory, including resource scope
and sampling assumptions, then expand an operation package into shared views.
A true build manifest should pin dependencies, integrations, exclusions and binary
identity and be checked against runtime conformance. This slice provides source
contracts and executable emission checks, not that full build pipeline.

The next implemented layer is [observed operation conformance](OPERATION_CONFORMANCE.md):
match normalized completed spans to the source contract and inspect invocation
counts, errors, panics and exact duration samples without inferring completeness.

`instrumentation.WithSubjectScope(ctx, namespace, tenant)` now attaches explicit
scope to annotation span events without subject metric dimensions. The live
source adapter and UI support [scoped subject navigation](SUBJECT_NAVIGATION.md);
this does not establish durable events, retries or business completion.
