package verification

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

type summaryTemplateVector404 struct {
	Source      workbench.ExplorationView       `json:"source"`
	Destination exploration.Query               `json:"destination"`
	Expansion   workbench.ViewTemplateExpansion `json:"expansion"`
}

func summaryTemplateSpecimen404(seed int, spans bool) (workbench.ExplorationView, exploration.Query) {
	signal := "logs"
	fields := []string{"count"}
	if spans {
		signal = "spans"
		fields = []string{"count", "duration.mean", "duration.p95", "status.error", "status.error_fraction"}
	}
	count := 1 + seed%min(6, len(fields))
	selected := []string{}
	for i := 0; i < count; i++ {
		selected = append(selected, fields[(seed+i)%len(fields)])
	}
	query := map[string]any{"signal": signal, "environment": "source404", "fromMs": 1000 + seed, "toMs": 10000 + seed, "limit": 3, "offset": 0, "where": map[string]any{"all": []any{map[string]any{"field": map[string]string{"kind": "resource", "key": "service.namespace"}, "op": "eq", "value": ""}, map[string]any{"any": []any{map[string]any{"field": map[string]string{"kind": "attribute", "key": "team"}, "op": "eq", "value": "雪"}, map[string]any{"field": map[string]string{"kind": "attribute", "key": "team"}, "op": "missing"}}}}}}
	if seed%5 != 0 {
		scope := map[string]any{"mode": "named", "value": "source404"}
		if seed%5 == 3 {
			scope = map[string]any{"mode": "empty"}
		}
		if seed%5 == 4 {
			scope = map[string]any{"mode": "unreported"}
		}
		query["environment"] = ""
		policy := "effective"
		if seed%5 == 1 {
			policy = "current-only"
		}
		query["environmentSelector"] = map[string]any{"version": "goo/environment-selector/v1", "policy": policy, "scope": scope}
	}
	if spans {
		query["summary"] = "span-duration"
		query["spanStatus"] = true
	}
	summary := map[string]any{"summary": map[string]any{"version": "goo.view/1", "scope": map[string]string{"kind": "record-query", "signal": signal}, "kind": map[string]any{"name": "measure-summary", "version": 1}, "display": map[string]any{"fields": selected, "title": "Relevant facts 雪"}}}
	records := map[string]any{"records": map[string]any{"version": "goo.view/1", "scope": map[string]string{"kind": "record-query", "signal": signal}, "kind": map[string]any{"name": []string{"record-table", "record-list"}[seed%2], "version": 1}, "display": map[string]any{"columns": []string{"utc", "summary", "service"}, "fieldColumns": []any{map[string]string{"kind": "attribute", "key": "team"}}}}}
	children := []any{summary, records}
	if seed%2 == 1 {
		children = []any{records, summary}
	}
	wire, _ := json.Marshal(map[string]any{"version": 1, "title": "Reusable facts " + strconv.Itoa(seed), "queries": []any{query}, "composition": map[string]any{"version": "goo.composition/1", "kind": map[string]any{"name": "stack", "version": 1}, "children": children}, "template": map[string]string{"version": "goo.template/1", "group": "none", "profile": "ordinary-summary"}})
	var source workbench.ExplorationView
	if err := json.Unmarshal(wire, &source); err != nil {
		panic(err)
	}
	destination := source.Queries[0]
	if destination.EnvironmentSelector == nil {
		destination.Environment = "destination404"
	}
	destination.FromMS += 100000
	destination.ToMS += 100000
	destination.Limit = 5
	destination.Where = &exploration.Predicate{All: []exploration.Predicate{*destination.Where, {Field: &exploration.Field{Kind: "attribute", Key: "phase"}, Op: "eq", Value: "destination"}}}
	return source, destination
}
func TestHegelSummaryTemplate404(t *testing.T) {
	vectors := []summaryTemplateVector404{}
	store := workbench.Store{Dir: t.TempDir()}
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		for _, spans := range []bool{false, true} {
			source, destination := summaryTemplateSpecimen404(seed, spans)
			before, _ := json.Marshal(source)
			saved, err := store.CreateExplorationView(source)
			if err != nil {
				tc.Fatal(err)
			}
			loaded, err := store.LoadExplorationView(saved.ID)
			if err != nil || !reflect.DeepEqual(saved, loaded) {
				tc.Fatal("immutable source round trip", err)
			}
			expanded, err := workbench.ExpandViewTemplate(workbench.ViewTemplateExpansionRequest{Source: saved, Query: &destination})
			if err != nil {
				tc.Fatal(err)
			}
			after, _ := json.Marshal(source)
			if !bytes.Equal(before, after) || !reflect.DeepEqual(expanded.View.Queries, []exploration.Query{destination}) || !reflect.DeepEqual(expanded.View.Composition, source.Composition) || !reflect.DeepEqual(expanded.View.Template, source.Template) {
				tc.Fatal("template changed source, destination scope, role/order or presentation")
			}
			accepted, err := store.CreateExplorationView(expanded.View)
			if err != nil {
				tc.Fatal(err)
			}
			again, err := store.LoadExplorationView(accepted.ID)
			if err != nil || !reflect.DeepEqual(accepted, again) {
				tc.Fatal("immutable destination round trip", err)
			}
			vectors = append(vectors, summaryTemplateVector404{*saved, destination, expanded})
		}
	}, options(24)...)
	controls := 0
	for _, v := range vectors {
		for _, mode := range []string{"empty", "unreported"} {
			raw, _ := json.Marshal(v.Source)
			var bad map[string]any
			json.Unmarshal(raw, &bad)
			q := bad["queries"].([]any)[0].(map[string]any)
			q["environment"] = ""
			q["environmentSelector"] = map[string]any{"version": "goo/environment-selector/v1", "policy": "current-only", "scope": map[string]string{"mode": mode}}
			raw, _ = json.Marshal(bad)
			var invalid workbench.ExplorationView
			err := json.Unmarshal(raw, &invalid)
			if err == nil {
				err = invalid.Validate()
			}
			if err == nil {
				t.Fatal("Go accepted current-only nonnamed environment")
			}
			controls++
		}
		for _, mutate := range []func(map[string]any){func(x map[string]any) { x["snapshot"] = map[string]any{} }, func(x map[string]any) { x["basis"] = "foreign" }, func(x map[string]any) { x["template"].(map[string]any)["group"] = "destination" }, func(x map[string]any) { x["template"].(map[string]any)["profile"] = "ordinary" }, func(x map[string]any) { x["template"].(map[string]any)["baseline"] = "destination" }, func(x map[string]any) { x["queries"].([]any)[0].(map[string]any)["offset"] = 1 }, func(x map[string]any) {
			x["queries"].([]any)[0].(map[string]any)["groupBy"] = []any{map[string]string{"kind": "column", "key": "service"}}
		}} {
			raw, _ := json.Marshal(v.Source)
			var x map[string]any
			json.Unmarshal(raw, &x)
			mutate(x)
			raw, _ = json.Marshal(x)
			var bad workbench.ExplorationView
			err := json.Unmarshal(raw, &bad)
			if err == nil {
				err = bad.Validate()
			}
			if err == nil {
				t.Fatal("Go accepted foreign template carrier")
			}
			controls++
		}
		for _, mutate := range []func(*exploration.Query){func(q *exploration.Query) { q.Offset = 1 }, func(q *exploration.Query) { q.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}} }, func(q *exploration.Query) {
			if q.Signal == "logs" {
				q.Signal = "spans"
			} else {
				q.Signal = "logs"
			}
		}} {
			raw, _ := json.Marshal(v.Destination)
			var bad exploration.Query
			json.Unmarshal(raw, &bad)
			mutate(&bad)
			if _, err := workbench.ExpandViewTemplate(workbench.ViewTemplateExpansionRequest{Source: &v.Source, Query: &bad}); err == nil {
				t.Fatal("Go accepted incompatible destination")
			}
			controls++
		}
	}
	t.Logf("%d Go enclosing carrier/destination refusal controls", controls)
	input, _ := json.Marshal(vectors)
	cmd := exec.CommandContext(t.Context(), "node", "verification/summary_template_404_fixture.mjs")
	cmd.Dir = ".."
	cmd.Stdin = bytes.NewReader(input)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Go/JS summary template carriers: %v %s", err, out)
	} else {
		t.Log(strings.TrimSpace(string(out)))
	}
	if err := os.MkdirAll("../artifacts/verification", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("../artifacts/verification/summary-template404-carriers.json", input, 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("%d complete Go/JS record-summary template/expansion/store carriers", len(vectors))
}
