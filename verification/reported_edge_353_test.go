package verification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
)

type reportedSpan353 struct {
	At       int64             `json:"timestamp_ms"`
	Service  string            `json:"ServiceName"`
	Trace    string            `json:"TraceId"`
	ID       string            `json:"SpanId"`
	Parent   string            `json:"ParentSpanId"`
	Resource map[string]string `json:"ResourceAttributes"`
	Attrs    map[string]string `json:"SpanAttributes"`
	Tag      string            `json:"tag"`
}
type reportedFixture353 struct {
	Mode, Duplicates  int
	Environment       string
	Parent, Child     string
	ParentNS, ChildNS *identity.Namespace
	Scope             exploration.InventoryQuestion
	Rows              []reportedSpan353
	Where             *exploration.Predicate
}
type reportedOwner353 struct {
	Present bool
	Name    string
	Service string
}

func reportedResource353(environment map[string]string, namespace *identity.Namespace) map[string]string {
	r := map[string]string{}
	for key, value := range environment {
		r[key] = value
	}
	if namespace.Mode == "value" {
		r["service.namespace"] = namespace.Value
	}
	return r
}
func reportedID353(value string, width int) bool {
	if len(value) != width || value == strings.Repeat("0", width) {
		return false
	}
	for _, r := range value {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') {
			return false
		}
	}
	return true
}

// Ownership is resolved from every parent row in the checked discovery domain
// before comparing a target owner or applying child predicates. A set of owner
// tuples deliberately collapses duplicate deliveries from one logical owner.
func reportedOracle353(f reportedFixture353) []reportedSpan353 {
	owners := map[[2]string]map[reportedOwner353]bool{}
	for _, row := range f.Rows {
		if row.At < f.Scope.FromMS || row.At >= f.Scope.ToMS || !continuationEnvironment352(f.Mode, f.Environment, row.Resource) || !continuationNamespace352(f.Scope.Namespace, row.Resource) {
			continue
		}
		key := [2]string{row.Trace, row.ID}
		if owners[key] == nil {
			owners[key] = map[reportedOwner353]bool{}
		}
		value, present := row.Resource["service.namespace"]
		owners[key][reportedOwner353{present, value, row.Service}] = true
	}
	target := reportedOwner353{f.ParentNS.Mode == "value", f.ParentNS.Value, f.Parent}
	selected := []reportedSpan353{}
	for _, row := range f.Rows {
		if row.At < f.Scope.FromMS || row.At >= f.Scope.ToMS || row.Service != f.Child || !continuationEnvironment352(f.Mode, f.Environment, row.Resource) || !continuationNamespace352(f.ChildNS, row.Resource) || !reportedID353(row.Trace, 32) || !reportedID353(row.Parent, 16) || f.Where != nil && row.Attrs["release"] != f.Where.Value {
			continue
		}
		ownerSet := owners[[2]string{row.Trace, row.Parent}]
		if len(ownerSet) == 1 && ownerSet[target] {
			selected = append(selected, row)
		}
	}
	return selected
}

func reportedFixtureFor353(seed, mode, nsMode, duplicates int, scoped bool) reportedFixture353 {
	f := reportedFixture353{Mode: mode, Duplicates: duplicates, Environment: fmt.Sprintf("edge-env-'\\雪-%d", seed), Parent: fmt.Sprintf("parent-'\\雪-%d", seed), Child: fmt.Sprintf("child-'\\雪-%d", seed)}
	namespace := func(mode int) *identity.Namespace {
		if mode == 0 {
			return &identity.Namespace{Mode: "missing"}
		}
		value := ""
		if mode == 2 {
			value = fmt.Sprintf("team-'\\雪-%d", seed)
		}
		return &identity.Namespace{Mode: "value", Value: value}
	}
	f.ParentNS, f.ChildNS = namespace(nsMode), namespace((nsMode+1)%3)
	if scoped {
		f.ChildNS = f.ParentNS.Clone()
	} else if seed%3 == 0 {
		f.Child = f.Parent // Equal names in different namespaces are distinct owners.
	}
	from := int64(1800000000000 + seed*1000)
	f.Scope = exploration.InventoryQuestion{Environment: f.Environment, FromMS: from, ToMS: from + 60000}
	if mode > 0 {
		f.Scope.Environment = ""
		f.Scope.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: f.Environment}}
		if mode == 1 {
			f.Scope.EnvironmentSelector.Policy = "current-only"
		} else if mode == 3 {
			f.Scope.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "empty"}
		} else if mode == 4 {
			f.Scope.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "unreported"}
		}
	}
	if scoped {
		f.Scope.Namespace = f.ParentNS.Clone()
	}
	f.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "release"}, Op: "eq", Value: "selected-'\\雪"}
	named := map[string]string{"deployment.environment.name": f.Environment}
	legacy := map[string]string{"deployment.environment": f.Environment}
	conflict := map[string]string{"deployment.environment.name": "foreign", "deployment.environment": f.Environment}
	empty := map[string]string{"deployment.environment.name": "", "deployment.environment": f.Environment}
	base := named
	if mode == 3 {
		base = empty
	} else if mode == 4 {
		base = map[string]string{}
	}
	ordinal := 0
	pair := func(tag string, parentEnv, childEnv map[string]string, parentAt, childAt int64) (reportedSpan353, reportedSpan353) {
		ordinal++
		trace, parentID := fmt.Sprintf("%032x", seed*100+ordinal), fmt.Sprintf("%016x", seed*100+ordinal*2)
		p := reportedSpan353{At: parentAt, Service: f.Parent, Trace: trace, ID: parentID, Resource: reportedResource353(parentEnv, f.ParentNS), Attrs: map[string]string{"release": "parent-only"}, Tag: tag + "-parent"}
		c := reportedSpan353{At: childAt, Service: f.Child, Trace: trace, ID: fmt.Sprintf("%016x", seed*100+ordinal*2+1), Parent: parentID, Resource: reportedResource353(childEnv, f.ChildNS), Attrs: map[string]string{"release": f.Where.Value}, Tag: tag + "-child"}
		f.Rows = append(f.Rows, p, c)
		return p, c
	}
	p, child := pair("valid", base, base, from+10, from+20)
	for i := 0; i < duplicates; i++ {
		f.Rows = append(f.Rows, p)
	}
	f.Rows = append(f.Rows, child) // Child delivery duplicates remain records.
	pair("legacy", legacy, legacy, from+10, from+20)
	pair("parent-conflict", conflict, base, from+10, from+20)
	pair("child-conflict", base, conflict, from+10, from+20)
	pair("empty-shadow", empty, base, from+10, from+20)
	pair("parent-outside", base, base, from-1, from+20)
	pair("child-outside", base, base, from+10, from+60000)
	p, _ = pair("ambiguous-service", base, base, from+10, from+20)
	p.Service = "foreign-owner"
	f.Rows = append(f.Rows, p)
	p, _ = pair("ambiguous-namespace", base, base, from+10, from+20)
	p.Resource = reportedResource353(base, namespace((nsMode+1)%3))
	f.Rows = append(f.Rows, p)
	p, _ = pair("foreign-environment-duplicate", base, base, from+10, from+20)
	p.Service, p.Resource = "foreign-owner", reportedResource353(conflict, f.ParentNS)
	f.Rows = append(f.Rows, p)
	p, child = pair("invalid-trace", base, base, from+10, from+20)
	f.Rows = f.Rows[:len(f.Rows)-2]
	p.Trace, child.Trace = "A"+strings.Repeat("0", 31), "A"+strings.Repeat("0", 31)
	f.Rows = append(f.Rows, p, child)
	_, child = pair("invalid-parent-id", base, base, from+10, from+20)
	f.Rows[len(f.Rows)-1].Parent = strings.Repeat("0", 16)
	_, child = pair("unresolved", base, base, from+10, from+20)
	f.Rows[len(f.Rows)-1].Parent = "ffffffffffffffff"
	_, child = pair("child-filter", base, base, from+10, from+20)
	f.Rows[len(f.Rows)-1].Attrs = map[string]string{"release": "foreign-release"}
	return f
}

func reportedUI353(f reportedFixture353) (exploration.Query, exploration.Query, error) {
	edge := map[string]any{"parent": f.Parent, "child": f.Child}
	if f.ParentNS.Mode == "value" {
		edge["parentNamespace"] = f.ParentNS.Value
	}
	if f.ChildNS.Mode == "value" {
		edge["childNamespace"] = f.ChildNS.Value
	}
	input, _ := json.Marshal(map[string]any{"scope": f.Scope, "edge": edge, "where": f.Where})
	cmd := exec.Command("node", "reported_edge_353_fixture.mjs")
	cmd.Stdin = bytes.NewReader(input)
	raw, err := cmd.CombinedOutput()
	if err != nil {
		return exploration.Query{}, exploration.Query{}, fmt.Errorf("actual UI reported-edge helpers: %v %s", err, raw)
	}
	var got struct {
		Exact, Filtered, Counted             exploration.Query
		Unchanged, Detached, CountedDetached bool
	}
	if json.Unmarshal(raw, &got) != nil || !got.Unchanged || !got.Detached || !got.CountedDetached {
		return exploration.Query{}, exploration.Query{}, fmt.Errorf("edge helper mutation isolation failed: %s", raw)
	}
	want := exploration.Query{Signal: "spans", Environment: f.Scope.Environment, EnvironmentSelector: f.Scope.EnvironmentSelector, FromMS: f.Scope.FromMS, ToMS: f.Scope.ToMS, Limit: 100, Summary: "span-duration", SpanStatus: true, GroupBy: []exploration.Field{{Kind: "column", Key: "name"}}, ReportedEdge: &exploration.ReportedEdge{Parent: exploration.EdgeService{Service: f.Parent, Namespace: f.ParentNS.Clone()}, Child: exploration.EdgeService{Service: f.Child, Namespace: f.ChildNS.Clone()}, DiscoveryNamespace: f.Scope.Namespace.Clone()}}
	if continuationJSON352(got.Exact) != continuationJSON352(want) {
		return exploration.Query{}, exploration.Query{}, fmt.Errorf("exact UI helper changed its qualified scope/profile: %s", raw)
	}
	want.Where = f.Where
	if continuationJSON352(got.Filtered) != continuationJSON352(want) || got.Filtered.Validate() != nil || got.Counted.Validate() != nil || continuationJSON352(got.Counted.ReportedEdge) != continuationJSON352(want.ReportedEdge) || continuationJSON352(got.Counted.EnvironmentSelector) != continuationJSON352(want.EnvironmentSelector) {
		return exploration.Query{}, exploration.Query{}, fmt.Errorf("child filter/counted values lost exact edge scope: %s", raw)
	}
	return got.Filtered, got.Counted, nil
}

type reportedTransport353 func(*http.Request) (*http.Response, error)

func (f reportedTransport353) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func reportedSource353(rows []map[string]any) exploration.Source {
	return exploration.Source{Endpoint: "http://fixture.invalid", Client: &http.Client{Transport: reportedTransport353(func(r *http.Request) (*http.Response, error) {
		sql, _ := io.ReadAll(r.Body)
		body := ""
		if strings.HasPrefix(string(sql), "SELECT toString(count()) AS count") {
			body = fmt.Sprintf("{\"count\":\"%d\"}", len(rows))
		} else if strings.Contains(string(sql), "AS timestamp_ns") {
			for _, row := range rows {
				body += continuationJSON352(row) + "\n"
			}
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
	})}}
}

func reportedRows353(rows []reportedSpan353) []map[string]any {
	out := []map[string]any{}
	for _, row := range rows {
		out = append(out, map[string]any{"timestamp_ns": strconv.FormatInt(row.At*1000000, 10), "duration_ns": "1000000", "SpanName": "reported child", "ServiceName": row.Service, "TraceId": row.Trace, "SpanId": row.ID, "ParentSpanId": row.Parent, "ResourceAttributes": row.Resource, "SpanAttributes": row.Attrs})
	}
	return out
}

func TestHegelReportedEdgeScope353(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		mode := hegel.Draw(tc, hegel.Integers(0, 4))
		nsMode := hegel.Draw(tc, hegel.Integers(0, 2))
		scoped := hegel.Draw(tc, hegel.Booleans())
		duplicates := hegel.Draw(tc, hegel.Integers(1, 3))
		f := reportedFixtureFor353(seed, mode, nsMode, duplicates, scoped)
		q, counted, err := reportedUI353(f)
		if err != nil {
			tc.Fatal(err)
			return
		}
		selected := reportedOracle353(f)
		tags := map[string]int{}
		for _, row := range selected {
			tags[row.Tag]++
		}
		if tags["valid-child"] != 2 || tags["foreign-environment-duplicate-child"] != 1 || tags["ambiguous-service-child"] != 0 || (tags["ambiguous-namespace-child"] == 1) != scoped || (tags["legacy-child"] == 1) != (mode == 2) || (tags["empty-shadow-child"] == 1) != (mode == 3) {
			tc.Fatal("independent owner oracle lost duplicate/ambiguity/presence distinctions", tags)
			return
		}
		for _, label := range []string{"parent-outside-child", "child-outside-child", "invalid-trace-child", "invalid-parent-id-child", "unresolved-child", "child-filter-child"} {
			if tags[label] != 0 {
				tc.Fatal("invalid child admitted by oracle", label)
				return
			}
		}
		plan, err := exploration.Compile(q)
		if err != nil {
			tc.Fatal(err)
			return
		}
		if _, err = exploration.Compile(counted); err != nil {
			tc.Fatal("typed counted edge query rejected", err)
			return
		}
		for _, sql := range []string{plan.Count, plan.Rows, plan.Groups, plan.Fields} {
			if sql == "" {
				continue
			}
			start, group := strings.Index(sql, "SELECT TraceId, SpanId FROM otel_traces WHERE "), strings.Index(sql, " GROUP BY TraceId, SpanId HAVING groupUniqArray")
			if start < 0 || group <= start || strings.Contains(sql[start:group], "ServiceName=") || strings.Contains(sql[start:group], "SpanAttributes") || strings.Contains(sql, f.Parent) || strings.Contains(sql, f.Child) || strings.Count(sql, "Timestamp >= fromUnixTimestamp64Milli({from:Int64})") < 2 {
				tc.Fatal("parent ownership narrowed before grouping or scope interpolated", sql)
				return
			}
		}
		if mode <= 1 {
			scalar, typed := q, q
			scalar.Environment, scalar.EnvironmentSelector = f.Environment, nil
			typed.Environment, typed.EnvironmentSelector = "", &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "current-only", Scope: identity.EnvironmentScope{Mode: "named", Value: f.Environment}}
			a, e1 := exploration.Compile(scalar)
			b, e2 := exploration.Compile(typed)
			if e1 != nil || e2 != nil || continuationJSON352(a) != continuationJSON352(b) {
				tc.Fatal("scalar/current-only compiler inequivalence", e1, e2)
				return
			}
		}
		plain := q
		plain.Summary, plain.SpanStatus, plain.GroupBy = "", false, nil
		rows := reportedRows353(selected)
		if result, err := reportedSource353(rows).Fetch(context.Background(), plain); err != nil || len(result.Rows) != len(rows) {
			tc.Fatal("valid selected child rows rejected", err)
			return
		}
		for _, poison := range []func(map[string]any){
			func(row map[string]any) { row["ServiceName"] = "foreign-child" },
			func(row map[string]any) { row["timestamp_ns"] = strconv.FormatInt((q.FromMS-1)*1000000, 10) },
			func(row map[string]any) { row["TraceId"] = "A" + strings.Repeat("0", 31) },
			func(row map[string]any) { row["ParentSpanId"] = strings.Repeat("0", 16) },
			func(row map[string]any) {
				row["ResourceAttributes"] = map[string]string{"deployment.environment.name": "foreign"}
			},
			func(row map[string]any) {
				resource := row["ResourceAttributes"].(map[string]any)
				if _, present := resource["service.namespace"]; present {
					delete(resource, "service.namespace")
				} else {
					resource["service.namespace"] = ""
				}
			},
		} {
			var bad map[string]any
			json.Unmarshal([]byte(continuationJSON352(rows[0])), &bad)
			poison(bad)
			if _, err := reportedSource353([]map[string]any{bad}).Fetch(context.Background(), plain); err == nil {
				tc.Fatal("invalid child row admitted by source")
				return
			}
		}
		ids := []string{}
		for _, row := range selected {
			ids = append(ids, row.ID)
		}
		slices.Sort(ids)
		if len(ids) < 2 || ids[0] == "" {
			tc.Fatal("missing bounded child records")
		}
	}, options(40)...)
}
