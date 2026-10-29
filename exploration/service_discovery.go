package exploration

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type ServiceSignal struct {
	Namespace *string `json:"namespace,omitempty"`
	Service   string  `json:"service"`
	Signal    string  `json:"signal"`
	Count     string  `json:"count"`
	FirstMS   int64   `json:"firstMs"`
	LastMS    int64   `json:"lastMs"`
}
type Dependency struct {
	ParentNamespace *string `json:"parentNamespace,omitempty"`
	ChildNamespace  *string `json:"childNamespace,omitempty"`
	Parent          string  `json:"parent"`
	Child           string  `json:"child"`
	Count           string  `json:"count"`
	Kind            string  `json:"kind"`
}
type ServiceDiscovery struct {
	Identity     string            `json:"identity,omitempty"`
	Question     InventoryQuestion `json:"question"`
	CapturedAt   string            `json:"capturedAt"`
	Basis        string            `json:"basis"`
	Signals      []ServiceSignal   `json:"signals"`
	Dependencies []Dependency      `json:"dependencies"`
	Plan         Plan              `json:"plan"`
	Warnings     []string          `json:"warnings"`
}

func CompileServiceDiscovery(q InventoryQuestion) (Plan, error) {
	if e := q.Validate(); e != nil {
		return Plan{}, e
	}
	c := compiler{signal: "spans", params: map[string]string{"environment": q.Environment, "from": strconv.FormatInt(q.FromMS, 10), "to": strconv.FormatInt(q.ToMS, 10)}}
	spanBase, _, sourceErr := c.source(q.Query())
	if sourceErr != nil {
		return Plan{}, sourceErr
	}
	logs := q.Query()
	logs.Signal = "logs"
	logBase, _, logSourceErr := c.source(logs)
	if logSourceErr != nil {
		return Plan{}, logSourceErr
	}
	aggregate := func(signal, base string) string {
		return "SELECT " + namespaceColumn("ResourceAttributes") + " AS namespace, ServiceName AS service, '" + signal + "' AS signal, toString(count()) AS count, toString(min(toUnixTimestamp64Milli(Timestamp))) AS first_ms, toString(max(toUnixTimestamp64Milli(Timestamp))) AS last_ms" + base + " GROUP BY namespace, service"
	}
	// Count exporter data-point records, never metric values or histogram observations.
	// Each family is aggregated before the union to bound intermediate cardinality.
	metricParts := []string{}
	for _, family := range []string{"gauge", "sum", "histogram"} {
		metricParts = append(metricParts, "SELECT "+namespaceColumn("ResourceAttributes")+" AS namespace, ServiceName AS service, count() AS n, min(toUnixTimestamp64Milli(TimeUnix)) AS first_ms, max(toUnixTimestamp64Milli(TimeUnix)) AS last_ms FROM otel_metrics_"+family+" WHERE TimeUnix>=fromUnixTimestamp64Milli({from:Int64}) AND TimeUnix<fromUnixTimestamp64Milli({to:Int64}) AND ResourceAttributes['deployment.environment.name']={environment:String} "+namespaceFilter(q.Namespace, c.params)+" GROUP BY namespace, service")
	}
	metrics := "SELECT namespace, service, 'metrics' AS signal, toString(sum(n)) AS count, toString(min(first_ms)) AS first_ms, toString(max(last_ms)) AS last_ms FROM (" + strings.Join(metricParts, " UNION ALL ") + ") GROUP BY namespace, service"
	signals := "SELECT * FROM (" + aggregate("spans", spanBase) + " UNION ALL " + aggregate("logs", logBase) + " UNION ALL " + metrics + ") ORDER BY service, namespace, signal LIMIT 1001 FORMAT JSONEachRow"
	// Parent identity is exact reported trace/span identity, scoped to this environment
	// and window. Ambiguous service ownership never becomes an arbitrary graph edge.
	owner := "tuple(mapContains(ResourceAttributes,'service.namespace'),ResourceAttributes['service.namespace'],ServiceName)"
	edges := "WITH captured AS (SELECT TraceId, SpanId, ParentSpanId, ServiceName, " + owner + " AS owner" + spanBase + "), parents AS (SELECT TraceId, SpanId, groupUniqArray(owner) AS services FROM captured GROUP BY TraceId, SpanId) SELECT c.ServiceName AS child, if(c.owner.1,toNullable(c.owner.2),NULL) AS child_namespace, if(length(p.services)=1,p.services[1].3,'') AS parent, if(length(p.services)=1 AND p.services[1].1,toNullable(p.services[1].2),NULL) AS parent_namespace, multiIf(NOT match(c.TraceId, '^[a-f0-9]{32}$') OR c.TraceId=repeat('0',32) OR NOT match(c.ParentSpanId, '^[a-f0-9]{16}$'), 'invalid', length(p.services)=0, 'unresolved', length(p.services)>1, 'ambiguous', c.owner=p.services[1], 'internal', 'edge') AS kind, toString(count()) AS count FROM captured AS c LEFT JOIN parents AS p ON c.TraceId=p.TraceId AND c.ParentSpanId=p.SpanId WHERE c.ParentSpanId NOT IN ('', '0000000000000000') GROUP BY child, child_namespace, parent, parent_namespace, kind ORDER BY child, child_namespace, parent, parent_namespace, kind LIMIT 1001 FORMAT JSONEachRow"

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
	r := &ServiceDiscovery{Identity: "otel-service-namespace/v1", Question: q, CapturedAt: time.Now().UTC().Format(time.RFC3339Nano), Signals: []ServiceSignal{}, Dependencies: []Dependency{}, Plan: p, Warnings: []string{"Event timestamps and counts describe captured spans, logs and native metric data-point records, not ingestion lag, telemetry completeness, service health or declared membership. Metric counts are records across gauge, sum and explicit-histogram tables, not values, unique streams or histogram observation counts.", "Parent links use reported trace/parent-span IDs within the environment and fixed window. Counts are child span records, not unique requests. Unresolved/ambiguous/invalid parents are diagnostics, not inferred edges.", "Signal presence and parent links are separate aggregate reads; no transactional consistency or complete topology is claimed. Metric-only services create no parent links. Summary/exponential-histogram metric families, asynchronous span links and external peer inference are not included."}}
	seen := map[[3]string]bool{}
	for _, v := range raw {
		var x struct {
			Service   *string         `json:"service"`
			Namespace json.RawMessage `json:"namespace"`
			Signal    string          `json:"signal"`
			Count     string          `json:"count"`
			First     string          `json:"first_ms"`
			Last      string          `json:"last_ms"`
		}
		if json.Unmarshal(v, &x) != nil || x.Service == nil || x.Signal != "spans" && x.Signal != "logs" && x.Signal != "metrics" {
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
		nspace, err := decodeNamespace(x.Namespace)
		if err != nil {
			return nil, err
		}
		key := [3]string{*x.Service, x.Signal, namespaceKey(nspace)}
		if seen[key] {
			return nil, fmt.Errorf("duplicate signal")
		}
		seen[key] = true
		r.Signals = append(r.Signals, ServiceSignal{Namespace: nspace, Service: *x.Service, Signal: x.Signal, Count: x.Count, FirstMS: first, LastMS: last})
	}
	raw, e = s.read(ctx, p.Groups, p.Parameters)
	if e != nil {
		return nil, e
	}
	if len(raw) > 1000 {
		return nil, fmt.Errorf("dependency cardinality exceeds 1000; narrow window")
	}
	keys := map[[5]string]bool{}
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
		pn, err := decodeNamespace(dimensions["parent_namespace"])
		if err != nil {
			return nil, err
		}
		cn, err := decodeNamespace(dimensions["child_namespace"])
		if err != nil {
			return nil, err
		}
		d.ParentNamespace, d.ChildNamespace = pn, cn
		n, e := inventoryNumber(d.Count)
		if e != nil || n.Sign() == 0 {
			return nil, fmt.Errorf("invalid dependency count")
		}
		if d.Kind != "edge" && d.Kind != "internal" && d.Kind != "ambiguous" && d.Kind != "unresolved" && d.Kind != "invalid" {
			return nil, fmt.Errorf("invalid relationship kind")
		}
		same := d.Parent == d.Child && namespaceKey(d.ParentNamespace) == namespaceKey(d.ChildNamespace)
		if d.Kind == "edge" && same || d.Kind == "internal" && !same {
			return nil, fmt.Errorf("invalid relationship identity")
		}
		key := [5]string{d.Child, d.Parent, d.Kind, namespaceKey(cn), namespaceKey(pn)}
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
