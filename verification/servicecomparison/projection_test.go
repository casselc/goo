package servicecomparison

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

// These properties check the new source-free entry against the existing Go
// comparison grammar. They do not claim additional source observations.
func TestCountProjectionAndPreviousWindow(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		questions := []exploration.Query{}
		for _, signal := range []string{"logs", "spans"} {
			for _, mode := range []string{"named", "empty", "unreported"} {
				width := int64(1+seed%100) * 1000
				scope := map[string]any{"mode": mode}
				if mode == "named" {
					scope["value"] = fmt.Sprintf("prod-'雪-%d", seed)
				}
				policy := "effective"
				if mode == "named" && seed%2 == 0 {
					policy = "current-only"
				}
				ns := map[string]any{"field": map[string]string{"kind": "resource", "key": "service.namespace"}, "op": "eq", "value": ""}
				if seed%2 == 0 {
					ns["op"] = "missing"
					delete(ns, "value")
				}
				numeric := map[string]any{"field": map[string]string{"kind": "attribute", "key": "payload.size"}, "type": "uint64", "unit": "By", "parse": "canonical-decimal"}
				wire := map[string]any{"signal": signal, "environment": "", "environmentSelector": map[string]any{"version": "goo/environment-selector/v1", "policy": policy, "scope": scope}, "fromMs": width*2 + int64(seed)*1000, "toMs": width*3 + int64(seed)*1000, "limit": 1 + seed%100, "offset": 0, "groupBy": []any{map[string]string{"kind": "attribute", "key": "http.route"}}, "where": map[string]any{"all": []any{map[string]any{"field": map[string]string{"kind": "column", "key": "service"}, "op": "eq", "value": "worker-'雪"}, ns, map[string]any{"numeric": numeric, "op": "gte", "value": "9007199254740993"}}}}
				if signal == "spans" {
					wire["spanStatus"] = true
					wire["summary"] = "span-duration"
				}
				if seed%2 == 0 {
					wire["bucketMs"] = 1000
				}
				raw, _ := json.Marshal(wire)
				var q exploration.Query
				if err := json.Unmarshal(raw, &q); err != nil {
					tc.Fatal(err)
				}
				if err := q.Validate(); err != nil {
					tc.Fatal(err)
				}
				questions = append(questions, q)
			}
		}
		input, _ := json.Marshal(questions)
		cmd := exec.CommandContext(t.Context(), "node", "bridge.mjs")
		cmd.Stdin = bytes.NewReader(input)
		output, err := cmd.CombinedOutput()
		if err != nil {
			tc.Fatalf("JS projection: %v %s", err, output)
		}
		var views []workbench.ExplorationView
		if err = json.Unmarshal(output, &views); err != nil {
			tc.Fatal(err)
		}
		if len(views) != len(questions) {
			tc.Fatal("one whole view required per question")
		}
		store := workbench.Store{Dir: t.TempDir()}
		for i, view := range views {
			if err = view.Validate(); err != nil {
				tc.Fatal(err)
			}
			expected := questions[i]
			expected.Summary = ""
			expected.SpanStatus = false
			if len(view.Queries) != 1 || !reflect.DeepEqual(view.Queries[0], expected) {
				tc.Fatal("count projection changed population, page or other query directives")
			}
			var baseline *exploration.Query
			for _, child := range view.Composition.Children {
				if child.Comparison != nil {
					baseline = child.Comparison.BaselineQuery
				}
			}
			prior := expected
			prior.FromMS = expected.FromMS - (expected.ToMS - expected.FromMS)
			prior.ToMS = expected.FromMS
			if baseline == nil || !reflect.DeepEqual(*baseline, prior) {
				tc.Fatal("previous window changed definition or exact half-open bounds")
			}
			for _, q := range []exploration.Query{expected, prior} {
				if _, err := exploration.Compile(q); err != nil {
					tc.Fatal(err)
				}
			}
			saved, err := store.CreateExplorationView(view)
			if err != nil {
				tc.Fatal(err)
			}
			loaded, err := store.LoadExplorationView(saved.ID)
			if err != nil || !reflect.DeepEqual(saved, loaded) {
				tc.Fatal("saved comparison changed", err)
			}
		}
	}, hegel.WithTestCases(8), hegel.WithSeed(409), hegel.WithDatabase(filepath.Join(t.TempDir(), "hegel")))
}
