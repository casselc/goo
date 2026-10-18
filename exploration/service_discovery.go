package exploration

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type ServiceSignal struct {
	Service string `json:"service"`
	Signal  string `json:"signal"`
	Count   string `json:"count"`
	FirstMS int64  `json:"firstMs"`
	LastMS  int64  `json:"lastMs"`
}
type Dependency struct {
	Parent string `json:"parent"`
	Child  string `json:"child"`
	Count  string `json:"count"`
	Kind   string `json:"kind"`
}
type ServiceDiscovery struct {
	Question     InventoryQuestion `json:"question"`
	CapturedAt   string            `json:"capturedAt"`
	Basis        string            `json:"basis"`
	Signals      []ServiceSignal   `json:"signals"`
	Dependencies []Dependency      `json:"dependencies"`
	Plan         Plan              `json:"plan"`
	Warnings     []string          `json:"warnings"`
}

func CompileServiceDiscovery(q InventoryQuestion) (Plan, error) {
	if e := q.Query().Validate(); e != nil {
		return Plan{}, e
	}
	c := compiler{signal: "spans", params: map[string]string{"environment": q.Environment, "from": strconv.FormatInt(q.FromMS, 10), "to": strconv.FormatInt(q.ToMS, 10)}}
	spanBase, _ := c.source(q.Query())
	logs := q.Query()
	logs.Signal = "logs"
	logBase, _ := c.source(logs)
	aggregate := func(signal, base string) string {
		return "SELECT ServiceName AS service, '" + signal + "' AS signal, toString(count()) AS count, toString(min(toUnixTimestamp64Milli(Timestamp))) AS first_ms, toString(max(toUnixTimestamp64Milli(Timestamp))) AS last_ms" + base + " GROUP BY service"
	}
	signals := "SELECT * FROM (" + aggregate("spans", spanBase) + " UNION ALL " + aggregate("logs", logBase) + ") ORDER BY service, signal LIMIT 1001 FORMAT JSONEachRow"
	// Parent identity is exact reported trace/span identity, scoped to this environment
	// and window. Ambiguous service ownership never becomes an arbitrary graph edge.
	edges := "WITH captured AS (SELECT TraceId, SpanId, ParentSpanId, ServiceName" + spanBase + "), parents AS (SELECT TraceId, SpanId, groupUniqArray(ServiceName) AS services FROM captured GROUP BY TraceId, SpanId) SELECT c.ServiceName AS child, if(length(p.services)=1, p.services[1], '') AS parent, multiIf(NOT match(c.TraceId, '^[a-f0-9]{32}$') OR c.TraceId=repeat('0',32) OR NOT match(c.ParentSpanId, '^[a-f0-9]{16}$'), 'invalid', length(p.services)=0, 'unresolved', length(p.services)>1, 'ambiguous', c.ServiceName=p.services[1], 'internal', 'edge') AS kind, toString(count()) AS count FROM captured AS c LEFT JOIN parents AS p ON c.TraceId=p.TraceId AND c.ParentSpanId=p.SpanId WHERE c.ParentSpanId NOT IN ('', '0000000000000000') GROUP BY child, parent, kind ORDER BY child, parent, kind LIMIT 1001 FORMAT JSONEachRow"
	return Plan{Count: signals, Groups: edges, Parameters: c.params}, nil
}
func (s Source) DiscoverServices(ctx context.Context, q InventoryQuestion) (*ServiceDiscovery, error) {
	p, e := CompileServiceDiscovery(q)
	if e != nil {
		return nil, e
	}
	raw, e := s.read(ctx, p.Count, p.Parameters)
	if e != nil {
		return nil, e
	}
	if len(raw) > 1000 {
		return nil, fmt.Errorf("service/signal cardinality exceeds 1000; narrow window")
	}
	r := &ServiceDiscovery{Question: q, CapturedAt: time.Now().UTC().Format(time.RFC3339Nano), Signals: []ServiceSignal{}, Dependencies: []Dependency{}, Plan: p, Warnings: []string{"Event timestamps and counts describe captured spans/logs, not ingestion lag, telemetry completeness, service health or declared membership.", "Parent links use reported trace/parent-span IDs within the environment and fixed window. Counts are child span records, not unique requests. Unresolved/ambiguous/invalid parents are diagnostics, not inferred edges.", "Signal presence and parent links are separate aggregate reads; no transactional consistency or complete topology is claimed. Native metrics, asynchronous span links and external peer inference are not included."}}
	seen := map[[2]string]bool{}
	for _, v := range raw {
		var x struct {
			Service *string `json:"service"`
			Signal  string  `json:"signal"`
			Count   string  `json:"count"`
			First   string  `json:"first_ms"`
			Last    string  `json:"last_ms"`
		}
		if json.Unmarshal(v, &x) != nil || x.Service == nil || x.Signal != "spans" && x.Signal != "logs" {
			return nil, fmt.Errorf("invalid service signal")
		}
		n, e := inventoryNumber(x.Count)
		if e != nil || n.Sign() == 0 {
			return nil, fmt.Errorf("invalid signal count")
		}
		first, e := strconv.ParseInt(x.First, 10, 64)
		if e != nil {
			return nil, e
		}
		last, e := strconv.ParseInt(x.Last, 10, 64)
		if e != nil || first < q.FromMS || last >= q.ToMS || last < first {
			return nil, fmt.Errorf("invalid signal event bounds")
		}
		key := [2]string{*x.Service, x.Signal}
		if seen[key] {
			return nil, fmt.Errorf("duplicate signal")
		}
		seen[key] = true
		r.Signals = append(r.Signals, ServiceSignal{*x.Service, x.Signal, x.Count, first, last})
	}
	raw, e = s.read(ctx, p.Groups, p.Parameters)
	if e != nil {
		return nil, e
	}
	if len(raw) > 1000 {
		return nil, fmt.Errorf("dependency cardinality exceeds 1000; narrow window")
	}
	keys := map[[3]string]bool{}
	for _, v := range raw {
		var dimensions map[string]json.RawMessage
		if json.Unmarshal(v, &dimensions) != nil {
			return nil, fmt.Errorf("invalid relationship row")
		}
		for _, key := range []string{"child", "parent"} {
			var value *string
			if json.Unmarshal(dimensions[key], &value) != nil || value == nil {
				return nil, fmt.Errorf("missing relationship dimension")
			}
		}
		var d Dependency
		if json.Unmarshal(v, &d) != nil {
			return nil, fmt.Errorf("invalid dependency")
		}
		n, e := inventoryNumber(d.Count)
		if e != nil || n.Sign() == 0 {
			return nil, fmt.Errorf("invalid dependency count")
		}
		if d.Kind != "edge" && d.Kind != "internal" && d.Kind != "ambiguous" && d.Kind != "unresolved" && d.Kind != "invalid" {
			return nil, fmt.Errorf("invalid relationship kind")
		}
		if d.Kind == "edge" && d.Parent == d.Child || d.Kind == "internal" && d.Parent != d.Child {
			return nil, fmt.Errorf("invalid relationship identity")
		}
		key := [3]string{d.Child, d.Parent, d.Kind}
		if keys[key] {
			return nil, fmt.Errorf("duplicate dependency")
		}
		keys[key] = true
		r.Dependencies = append(r.Dependencies, d)
	}
	r.CapturedAt = time.Now().UTC().Format(time.RFC3339Nano)
	basis := *r
	basis.CapturedAt = ""
	b, e := json.Marshal(basis)
	if e != nil {
		return nil, e
	}
	h := sha256.Sum256(b)
	r.Basis = hex.EncodeToString(h[:])
	return r, nil
}
