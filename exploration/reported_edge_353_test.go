package exploration

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/casselc/goo/identity"
)

func edgeQueries353() []Query {
	queries := []Query{edgeQuery()}
	selectors := []identity.EnvironmentSelector{{Version: identity.EnvironmentSelectorVersion, Policy: "current-only", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}}}
	for _, scope := range []identity.EnvironmentScope{{Mode: "named", Value: "prod"}, {Mode: "empty"}, {Mode: "unreported"}} {
		selection, _ := identity.EffectiveEnvironment(scope)
		wire, _ := selection.Selector()
		selectors = append(selectors, wire)
	}
	for _, selector := range selectors {
		q := edgeQuery()
		q.Environment = ""
		q.EnvironmentSelector = new(selector)
		queries = append(queries, q)
	}
	return queries
}

func TestReportedEdgeSharedOwnershipScope353(t *testing.T) {
	for _, q := range edgeQueries353() {
		q.Where = &Predicate{Field: &Field{Kind: "attribute", Key: "child-only"}, Op: "eq", Value: "selected-child"}
		selection, _ := q.EnvironmentSelection()
		predicate, bindings, _ := environmentSelectionPredicate(selection)
		plan, err := Compile(q)
		if err != nil {
			t.Fatal(err)
		}
		for _, sql := range []string{plan.Count, plan.Rows, plan.Fields} {
			if strings.Count(sql, predicate) != 2 {
				t.Fatalf("parent/child environment scope diverged: %s", sql)
			}
			start := strings.Index(sql, "SELECT TraceId, SpanId FROM otel_traces WHERE ")
			end := strings.Index(sql[start:], " GROUP BY TraceId, SpanId HAVING")
			parent := sql[start : start+end]
			if strings.Contains(parent, "ServiceName=") || strings.Contains(parent, "SpanAttributes") || strings.Contains(parent, "child-only") {
				t.Fatalf("ownership prefiltered by selected owner/child: %s", parent)
			}
			if !strings.Contains(sql, "groupUniqArray(tuple(mapContains(ResourceAttributes,'service.namespace'),ResourceAttributes['service.namespace'],ServiceName))=[") {
				t.Fatal("lost unique namespace-aware ownership")
			}
			if !strings.Contains(parent, "Timestamp < fromUnixTimestamp64Milli({to:Int64})") {
				t.Fatal("parent lost half-open window")
			}
		}
		for key, value := range bindings {
			if plan.Parameters[key] != value {
				t.Fatal("policy binding changed")
			}
		}
		if selection.Scope().Mode != "named" {
			if _, ok := plan.Parameters["environment"]; ok {
				t.Fatal("nameless policy inherited named binding")
			}
		}
	}
}

func TestReportedEdgeLocalRowAdmission353(t *testing.T) {
	for _, q := range edgeQueries353() {
		selection, _ := q.EnvironmentSelection()
		resource := map[string]string{}
		if selection.CurrentOnly() {
			resource[identity.EnvironmentKey] = "prod"
		} else {
			switch selection.Scope().Mode {
			case "named":
				resource[identity.LegacyEnvironmentKey] = "prod"
			case "empty":
				resource[identity.EnvironmentKey] = ""
				resource[identity.LegacyEnvironmentKey] = "prod"
			}
		}
		row := map[string]any{"timestamp_ns": "1000000000", "SpanName": "work", "ServiceName": "worker", "TraceId": strings.Repeat("a", 32), "ParentSpanId": strings.Repeat("b", 16), "ResourceAttributes": resource}
		encode := func() json.RawMessage {
			b, err := json.Marshal(row)
			if err != nil {
				t.Fatal(err)
			}
			return b
		}
		if err := validateSelectedRows(q, []json.RawMessage{encode()}); err != nil {
			t.Fatal(q, err)
		}
		snapshots := &ResultSnapshots{}
		if _, err := snapshots.Freeze(&Result{Question: q, Rows: []json.RawMessage{encode()}}); err != nil {
			t.Fatal("valid edge snapshot refused", err)
		}
		for _, change := range []struct {
			key   string
			value any
		}{{"ServiceName", "other"}, {"TraceId", strings.Repeat("0", 32)}, {"ParentSpanId", strings.Repeat("B", 16)}, {"timestamp_ns", "61000000000"}, {"ResourceAttributes", map[string]string{identity.NamespaceKey: ""}}, {"ResourceAttributes", map[string]any{identity.NamespaceKey: nil}}} {
			old := row[change.key]
			row[change.key] = change.value
			if validateSelectedRows(q, []json.RawMessage{encode()}) == nil {
				t.Fatalf("accepted foreign/malformed %s for %+v", change.key, q)
			}
			if _, err := snapshots.Freeze(&Result{Question: q, Rows: []json.RawMessage{encode()}}); err == nil {
				t.Fatalf("foreign/malformed %s snapshot issued", change.key)
			}
			row[change.key] = old
		}
		raw := strings.Replace(string(encode()), `"worker"`, `"\ud800"`, 1)
		if validateSelectedRows(q, []json.RawMessage{json.RawMessage(raw)}) == nil {
			t.Fatal("repaired unpaired surrogate")
		}
		q.ReportedEdge.Child.Namespace = &identity.Namespace{Mode: "value"}
		resource[identity.NamespaceKey] = ""
		if err := validateSelectedRows(q, []json.RawMessage{encode()}); err != nil {
			t.Fatal("present-empty namespace refused", err)
		}
		delete(resource, identity.NamespaceKey)
		if validateSelectedRows(q, []json.RawMessage{encode()}) == nil {
			t.Fatal("missing namespace became present empty")
		}
	}
}

func TestReportedEdgeRawIdentityDecode353(t *testing.T) {
	for _, literal := range []string{`"\ud800"`, `"\udfff"`, `"bad\ud800tail"`} {
		for _, dimension := range []string{"service", "namespace"} {
			service, namespace := `"parent"`, `"payments"`
			if dimension == "service" {
				service = literal
			} else {
				namespace = literal
			}
			raw := `{"signal":"spans","environment":"prod","fromMs":1000,"toMs":61000,"limit":100,"reportedEdge":{"parent":{"service":` + service + `,"namespace":{"mode":"value","value":` + namespace + `}},"child":{"service":"worker","namespace":{"mode":"missing"}}}}`
			var q Query
			if err := json.Unmarshal([]byte(raw), &q); err == nil {
				t.Fatalf("repaired malformed %s identity: %s", dimension, raw)
			}
		}
	}
	for _, literal := range []string{`"\ufffd"`, `"�"`, `"\ud83d\ude80"`} {
		raw := `{"signal":"spans","environment":"prod","fromMs":1000,"toMs":61000,"limit":100,"reportedEdge":{"parent":{"service":` + literal + `,"namespace":{"mode":"value","value":` + literal + `}},"child":{"service":"worker","namespace":{"mode":"missing"}}}}`
		var q Query
		if err := json.Unmarshal([]byte(raw), &q); err != nil {
			t.Fatal("valid literal rejected", err)
		}
		if err := q.Validate(); err != nil {
			t.Fatal("valid identity rejected", err)
		}
	}
	for _, raw := range []string{
		`{"parent":{"service":"api","namespace":{"mode":"missing"},"extra":true},"child":{"service":"worker","namespace":{"mode":"missing"}}}`,
		`{"parent":{"service":"api","namespace":{"mode":"missing","extra":true}},"child":{"service":"worker","namespace":{"mode":"missing"}}}`,
	} {
		var edge ReportedEdge
		if json.Unmarshal([]byte(raw), &edge) == nil {
			t.Fatal("unknown identity field accepted")
		}
	}
}
