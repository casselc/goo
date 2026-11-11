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
	"github.com/casselc/goo/scopedview"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

type histogramVector407 struct {
	Dataset          *scopedview.Dataset             `json:"dataset"`
	Source           workbench.ExplorationView       `json:"source"`
	Destination      exploration.Query               `json:"destination"`
	Expansion        workbench.ViewTemplateExpansion `json:"expansion"`
	Rows             []map[string]string             `json:"rows"`
	Support          map[string]string               `json:"support"`
	ExpectedComplete bool                            `json:"expectedComplete"`
}

func histogramView407(t testing.TB, q exploration.Query, seed int) workbench.ExplorationView {
	t.Helper()
	leaf := map[string]any{"version": "goo.view/1", "scope": map[string]string{"kind": "record-query", "signal": q.Signal}, "kind": map[string]any{"name": []string{"histogram", "table"}[seed%2], "version": 1}, "display": map[string]any{"category": "from", "value": "count", "title": "Value distribution 雪", "requiresComplete": seed%3 == 0}}
	records := map[string]any{"version": "goo.view/1", "scope": map[string]string{"kind": "record-query", "signal": q.Signal}, "kind": map[string]any{"name": []string{"record-table", "record-list"}[seed%2], "version": 1}, "display": map[string]any{"columns": []string{"utc", "summary", "service"}}}
	children := []any{map[string]any{"histogram": leaf}, map[string]any{"records": records}}
	if seed%2 == 1 {
		children[0], children[1] = children[1], children[0]
	}
	wire, _ := json.Marshal(map[string]any{"version": 1, "title": "Histogram " + strconv.Itoa(seed), "queries": []exploration.Query{q}, "composition": map[string]any{"version": "goo.composition/1", "kind": map[string]any{"name": "stack", "version": 1}, "children": children}, "template": map[string]string{"version": "goo.template/1", "group": "none", "profile": "ordinary-histogram"}})
	var view workbench.ExplorationView
	if e := json.Unmarshal(wire, &view); e != nil {
		t.Fatal(e)
	}
	return view
}

func TestHegelHistogramView407(t *testing.T) {
	vectors := []histogramVector407{}
	store := workbench.Store{Dir: t.TempDir()}
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		for _, typ := range []string{"int64", "uint64", "decimal64"} {
			for mode := 0; mode < 3; mode++ {
				d, q, rows, support, e := histogramContractFixture(typ, 1+seed%6, seed, mode)
				if e != nil {
					tc.Fatal(e)
				}
				view := histogramView407(t, q, seed)
				if e = view.Validate(); e != nil {
					tc.Fatal(e)
				}
				before, _ := json.Marshal(view)
				saved, e := store.CreateExplorationView(view)
				if e != nil {
					tc.Fatal(e)
				}
				loaded, e := store.LoadExplorationView(saved.ID)
				if e != nil || !reflect.DeepEqual(saved, loaded) {
					tc.Fatal("histogram store carrier changed", e)
				}
				dest := q
				dest.Environment = "destination407"
				dest.FromMS += 100000
				dest.ToMS += 100137
				dest.Limit = 3
				dest.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "eq", Value: ""}
				expanded, e := workbench.ExpandViewTemplate(workbench.ViewTemplateExpansionRequest{Source: &view, Query: &dest})
				if e != nil {
					tc.Fatal(e)
				}
				if !reflect.DeepEqual(expanded.View.Queries, []exploration.Query{dest}) || !sameTransport384(expanded.View.Composition, view.Composition) {
					tc.Fatal("histogram reuse changed destination or presentation")
				}
				for _, mutate := range []func(*exploration.Query){
					func(q *exploration.Query) { q.Offset = 1 },
					func(q *exploration.Query) { q.Numeric.Unit += " " },
					func(q *exploration.Query) { q.Numeric.Field.Kind = "resource" },
					func(q *exploration.Query) { q.Numeric.Field.Key += ".other" },
					func(q *exploration.Query) { q.Numeric.Type = "decimal64"; q.Numeric.Scale = 7 },
					func(q *exploration.Query) { q.Numeric.Parse = "coerce" },
					func(q *exploration.Query) { q.NumericHistogram.Binning = "linear" },
					func(q *exploration.Query) { q.NumericHistogram = nil },
					func(q *exploration.Query) { q.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}} },
				} {
					raw, _ := json.Marshal(dest)
					var bad exploration.Query
					_ = json.Unmarshal(raw, &bad)
					mutate(&bad)
					if _, e := workbench.ExpandViewTemplate(workbench.ViewTemplateExpansionRequest{Source: &view, Query: &bad}); e == nil {
						tc.Fatal("incompatible histogram destination accepted")
					}
				}
				after, _ := json.Marshal(view)
				if !bytes.Equal(before, after) {
					tc.Fatal("donor mutated")
				}
				vectors = append(vectors, histogramVector407{Dataset: d, Source: view, Destination: dest, Expansion: expanded, Rows: rows, Support: support, ExpectedComplete: mode == 2})
			}
		}
	}, options(6)...)
	raw, _ := json.Marshal(vectors)
	cmd := exec.CommandContext(t.Context(), "node", "verification/histogram_view_407_fixture.mjs")
	cmd.Dir = ".."
	cmd.Stdin = bytes.NewReader(raw)
	out, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("histogram Go/JS bridge: %v %s", e, out)
	}
	t.Log(strings.TrimSpace(string(out)))
	if e = os.MkdirAll("../artifacts/verification", 0755); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile("../artifacts/verification/histogram407-carriers.json", raw, 0644); e != nil {
		t.Fatal(e)
	}
}
