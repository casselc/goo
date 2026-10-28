# Independent architecture review — after 235

Reviewed 2026-10-28 at `5d12ce9`. This is an independent architecture/source review,
not an endorsement of the current backlog. No runtime source was edited, live
source queried, stored data changed or production deployment attempted. One
isolated loopback Go harness exercised the actual admission and JSON-decoding
packages. The review used local source/tooling and did not refresh external documentation.

## Conclusion and priority challenge

The product's main architectural direction is sound: typed semantics over ordinary
OTel, bounded source adapters, explicit effects, immutable captures and passive
framework-neutral display elements. Keep those boundaries. The main architectural
risk is now interaction between otherwise well-bounded layers, rather than lack of
a general query engine or storage abstraction.

I recommend a small ingress reliability cut before the larger catalog/reuse
features, alongside the mobile Activity work. I also recommend resolving package
catalog capacity admission before increasing revision-producing features. P2 saved
capture discovery is a strong user outcome, but its integrity and metadata-cost
contract needs an explicit decision before an index is implemented.

| Finding | Classification | Priority |
| --- | --- | --- |
| Cancelled incomplete JSON uploads retain query admission until the peer closes | Reproduced defect | Early bounded reliability cut |
| Native package publication can cross the reader's 256-file catalog bound | Source-confirmed producer/reader contract defect; no endpoint reproduction performed | Before further package/reuse expansion |
| Capture metadata search/order currently requires decisions about full-content verification cost | Verified architectural constraint, not a newly reproduced runtime failure | Design prerequisite to P2 |
| Native/prepared ClickHouse cancellation and resource settings differ from record-source settings | Source-confirmed policy difference; continued source execution is an unverified hypothesis | Targeted source conformance probe, then decide |
| Trace question reuse crosses a distinct semantic/persistence boundary | Verified integration constraint; proposed work | P4 after a selected recurring task |

## 1. Request context cancellation does not interrupt request-body decoding

`internal/queryhttp/budget.go:97-112` starts a 60-second context, admits the request,
and releases admission only after the downstream handler returns. Its cancellation
route calls the cancel function. `internal/httpjson/json.go:18-27` synchronously
reads `r.Body` through a byte-limited JSON decoder. A byte limit is not a read-time
limit. `cmd/goo/serve_http.go:122` sets a five-second header timeout and 90-second
write timeout, but no request-body read deadline.

The actual packages were exercised through an isolated `httptest` server with a
one-slot budget, a `/read` handler using `httpjson.Decode`, and a valid query ID.
The client sent headers declaring 100 bytes but only wrote `{`:

```text
cancel status 202
request context cancelled
next valid request while original body incomplete 429
Body.Close still blocked 150ms after context cancellation
next valid request after original connection closed 200
```

The cancellation acknowledgement and request context were observed immediately;
the incomplete read did not return. The `Body.Close` variant still blocked for an
additional 150ms. After closing the original client TCP connection and waiting
20ms, a valid request succeeded. The configured budget's 60-second deadline was
**not waited out**; this observation specifically proves the explicit-cancellation
case and explains why a context ceiling alone is not an I/O read deadline. Normal
browser abort, which closes its connection, is a different case and remains
covered by existing tests.

A slow/incomplete request can occupy an admission slot despite successful manual
cancellation. The default four-slot server limits the immediate exposure; local
loopback defaults and the existing lack of shared-deployment authentication mean
this is not a claim of an Internet-exposed exploit. It is still a reliability
contract defect in the present cancellation mechanism.

**Bounded response:** define transport body-read deadlines/cancellation separately
from post-validation source deadlines. Cover direct Go and Python proxy ingress;
`server.py:187-193` also performs a blocking `rfile.read(length)` before forwarding,
and its `ThreadingHTTPServer` has no corresponding body-time policy. The Python
case is source inspection, not separately reproduced here. Do not assume a
synchronous `Body.Close` fixes the Go case. Avoid applying ordinary fixed-query
write/read limits indiscriminately to SSE.

**Acceptance:** hold an incomplete body; cancel or expire its permitted receive
time; assert prompt admission release, zero downstream source calls, no leaked
handler, and a subsequent valid request succeeds before the original peer closes.
Also check cancellation-before-admission tombstones, exact size refusal, normal
large supported requests, HTTP/1 behavior, any supported HTTP/2 behavior, proxy
forwarding and unaffected SSE lifetime. Use a real network request, not only a
pre-cancelled in-memory reader. Existing `internal/queryhttp/budget_test.go` covers
response-header cancellation; add this distinct incoming-body failure class.

## 2. Package writer and catalog reader do not share a capacity contract

The native package reader at `cmd/goo/measurements_http.go:68-125` reads at most
257 directory entries per directory and fails once total YAML entries exceed 256.
The publisher at `:179-217` validates the definition, publishes immutable bytes and
returns 201 without considering catalog capacity. Therefore publication can create
a valid new revision which the following complete catalog request cannot list.
The source paths establish that mismatch; I did not fill a lab directory or call
the live publisher to reproduce it.

This is not a request to remove the bound. Define whether the supported behavior
is capacity admission/refusal, a paged bounded metadata catalog, or another explicit
contract. Account for built-ins plus saved files, file-scan limits versus unique
revision counts, duplicate idempotent publication, and concurrent publishers. A
check-then-publish count without serialization/reservation can race. Publishing a
built-in revision into the saved directory can consume a physical entry without
adding a unique catalog revision, so a simplistic unique-count check is incomplete.

**Acceptance:** a new distinct save at capacity cannot acknowledge success and
break the existing catalog; duplicate retries retain their existing semantics;
concurrent saves cannot bypass the selected capacity policy. Existing valid IDs,
bytes, winner verification and corruption refusal remain intact. Inventory other
definition stores rather than assuming they share this precise native route.

## 3. P2 capture discovery needs an integrity/cost decision, not just a sort

`exploration/capture_store.go:492-547` scans at most 10,001 directory entries,
selects 20 IDs in lexical order, and calls `Load` for each selected capture.
`Load` at `:381-416` reads up to 24 MiB, decodes complete content, verifies its digest
and validates the saved result. This is stronger than reading a small receipt.
`CaptureReceipt` at `:135-142` is a useful existing metadata shape: ID, parent,
title, captured time and typed question.

A whole-catalog recent-first search cannot be implemented correctly by sorting one
20-item page. Conversely, expanding current full-result validation to all 10,000
entries on every search would dramatically increase possible read/validation work.
That is a scaling constraint derived from bounds, not a measured production
performance result.

Prefer a capture-family metadata projection/index with explicit lifecycle and
limits over a universal store rewrite. But state the integrity promise clearly:
a publication-time verified metadata index cannot detect later payload corruption
without rechecking payload content. Retaining today's immediate catalog corruption
refusal and avoiding all content reads are competing requirements. Decide which
checks occur on list versus reopen/rebuild, what a stale/unavailable index means,
and how publication ambiguity/recovery works. Do not silently weaken integrity
under an apparently presentation-only change. Keep direct-ID reopen independent.

Service scope extraction must describe the actual typed question. A predicate can
cover several services or nested conditions; choosing the first service literal
would manufacture entity identity. Reuse the existing scope-description boundary
where applicable and expose unavailable/compound scope honestly.

## 4. Source-policy parity deserves one focused conformance check

`exploration/source.go:71-78` explicitly sends ClickHouse read-only cancellation on
client close, a time ceiling, result bounds, memory bound and bytes-to-read bound.
`measurements/source.go:166-168` sets readonly/time/rows/bytes but not that explicit
client-close setting or memory setting. `prebuilt/clickhouse_http.go:135` onward
has its own historical policy. These differences are verified; their actual effect
with the deployed ClickHouse profiles was not tested here.

Do not report that every native query necessarily leaks after cancellation. The
server may have its own policy and the source still has execution limits. Instead,
run a real long query/cancel conformance check per source family, observe actual
query termination and admission release, then make only the justified settings
consistent. Share reviewed transport-policy helpers if that removes demonstrated
drift, while keeping family row limits/semantics distinct. This is a better seam
check than another generic HTTP framework extraction.

## 5. Preserve the existing semantic and persistence layers

- `subjectfacts/compile.go:10-20` produces inspectable prepared-fact artifacts with
  explicit requirements. This is the offline ClickHouse/Databricks compilation
  path, not a deployed planner or proof of live warehouse parity.
- `measurements/source.go:257-268` makes supplied native evaluation storage-free;
  lifecycle custody is an explicit different operation. It still constructs the
  native result/plan contract. Do not call it a backend-independent optimizer or
  force DataFusion into this path merely to remove an SQL artifact.
- Native `Question.Validate` preserves canonical exact nanosecond bounds with
  matching enclosing millisecond bounds (`measurements/model.go:271-295`). Keep
  event time, fixed evaluation time and custody/read clocks separate.
- Typed source/OTel profiles correctly distinguish histogram/counter/state/stock
  semantics, reset/gap/unknown and literal namespace/map identity. New convention
  aliases and role profiles require explicit version/unit/source contracts.
- `internal/immutablefile/publish.go:10-35` explicitly distinguishes link success,
  collision and directory-sync ambiguity. Preserve winner checks and ambiguous
  commit recovery. Do not replace this with a generic Save interface that promises
  rollback or atomic batches the implementation does not provide.
- `ResultSnapshots` and tail membership capabilities have explicit process/byte/
  time bounds. They are not authenticated user identity or durable collaboration
  state. Shared deployment needs a tenancy/authentication decision before these
  capabilities are treated as cross-user guarantees.
- UI source ownership stays in application adapters; display custom elements remain
  supplied-data renderers. Resource-return focus/ancestry being app-owned is sound.
  Do not move query/store subscriptions into component lifecycle for reuse.

## 6. P4 is useful but is a distinct question family

`exploration/trace_search.go:18-27` represents an environment/window plus an array
of span conditions; `:77-80` validates each condition as its own span query.
`ui/question-protocol.mjs:37-38` currently distinguishes record and measurement
steps. Trace search is neither one record predicate nor a native measurement.
Preserve its witness semantics when extending saved questions: conditions may
match different services or the same observation; they do not establish topology,
ordering, complete traces or causation. A versioned question family can reuse the
existing registry/document surfaces without pretending all definitions are the
same AST. This remains behind the selected recurring task and capture discovery.

## Recommended sequence and stopping rules

1. Add the small incoming-body reliability cut; mobile scanning can proceed in
   parallel as a product/display change.
2. Resolve native package catalog publication/capacity, then implement P2 against
   a reviewed bounded metadata/integrity/cursor contract.
3. Add descriptive comparison support context; keep exact counts and uncertainty,
   and do not introduce unsupported statistical significance.
4. Extend saved trace questions for one recurring task with a versioned semantic
   contract and old-question compatibility controls.
5. Use the targeted source cancellation conformance check to decide whether any
   transport-policy deduplication is justified.

Stop the generic hash cleanup, broad universal-query/store/history refactors and
mandatory optional-storage integrations. Prerelease compatibility is negotiable,
but revisions, native exact bounds and existing durable captures are real user
contracts; change them only for a demonstrated benefit with a migration or explicit
preservation rule. Preserve production authentication/isolation/retention/restore as
separate selected deployment work, not an inferred requirement to ship a local UI
polish cut.

## Reproducer for finding 1

Use Go 1.27.2. In a temporary directory create `go.mod` with the following content
(adjust the local replacement path), then `main.go` below. This imports the actual
current admission/decoder implementations; it does not start Goo or contact the lab.

```go
module github.com/casselc/goo/architectureprobe

go 1.27.0
require github.com/casselc/goo v0.0.0
replace github.com/casselc/goo => /workspace/goo
```

Run `go run .` for context cancellation alone and `go run . close` to also attempt
`Body.Close`. The 20ms post-close wait is the observed harness interval, not a
promised scheduling SLA. All connections and the temporary server are closed.

```go
package main

import (
	"fmt"
	"github.com/casselc/goo/internal/httpjson"
	"github.com/casselc/goo/internal/queryhttp"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"
)

func main() {
	mode := "plain"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	closed := make(chan struct{}, 1)
	entered := make(chan struct{}, 1)
	cancelled := make(chan struct{}, 1)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Goo-Query-ID") != "" {
			entered <- struct{}{}
			go func() {
				<-r.Context().Done()
				select {
				case cancelled <- struct{}{}:
				default:
				}
				if mode == "close" {
					r.Body.Close()
					closed <- struct{}{}
				}
			}()
		}
		var v map[string]string
		if e := httpjson.Decode(w, r, &v, 1024, "trailing"); e != nil {
			http.Error(w, e.Error(), 400)
			return
		}
		w.Write([]byte("ok"))
	})
	s := httptest.NewServer(queryhttp.NewBudget(handler, 1, func(r *http.Request) bool { return r.URL.Path == "/read" }))
	defer s.Close()
	client := s.Client()
	client.Timeout = 2 * time.Second
	c, e := net.Dial("tcp", strings.TrimPrefix(s.URL, "http://"))
	if e != nil {
		panic(e)
	}
	defer c.Close()
	id := "12345678-1234-1234-1234-123456789abc"
	fmt.Fprintf(c, "POST /read HTTP/1.1\r\nHost: localhost\r\nContent-Length: 100\r\nX-Goo-Query-ID: %s\r\n\r\n{", id)
	<-entered
	r, e := client.Post(s.URL+"/queries/"+id+"/cancel", "application/json", nil)
	if e != nil {
		panic(e)
	}
	io.Copy(io.Discard, r.Body)
	r.Body.Close()
	fmt.Println("cancel status", r.StatusCode)
	select {
	case <-cancelled:
		fmt.Println("request context cancelled")
	case <-time.After(time.Second):
		panic("not cancelled")
	}
	r, e = client.Post(s.URL+"/read", "application/json", strings.NewReader("{}"))
	if e != nil {
		panic(e)
	}
	io.Copy(io.Discard, r.Body)
	r.Body.Close()
	fmt.Println("next valid request while original body incomplete", r.StatusCode)
	if mode == "close" {
		select {
		case <-closed:
			fmt.Println("Body.Close completed before peer close")
		case <-time.After(150 * time.Millisecond):
			fmt.Println("Body.Close still blocked 150ms after context cancellation")
		}
	}
	c.Close()
	time.Sleep(20 * time.Millisecond)
	r, e = client.Post(s.URL+"/read", "application/json", strings.NewReader("{}"))
	if e != nil {
		panic(e)
	}
	io.Copy(io.Discard, r.Body)
	r.Body.Close()
	fmt.Println("next valid request after original connection closed", r.StatusCode)
}
```
