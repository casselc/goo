package serviceversions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

type projection struct {
	Question    exploration.Query
	Count, Mean workbench.ExplorationView
}

func leaf(kind, key, op, value string) map[string]any {
	p := map[string]any{"field": map[string]string{"kind": kind, "key": key}, "op": op}
	if op != "missing" {
		p["value"] = value
	}
	return p
}
func fixture(seed, mode int) (map[string]any, exploration.Query, any, any) {
	service := fmt.Sprintf("worker-'雪-%d", seed)
	scope := map[string]any{"mode": []string{"named", "empty", "unreported"}[mode%3]}
	if mode%3 == 0 {
		scope["value"] = fmt.Sprintf("prod-%d", seed)
	}
	policy := "effective"
	if mode == 3 {
		policy = "current-only"
	}
	env := map[string]any{"version": "goo/environment-selector/v1", "policy": policy, "scope": scope}
	ns := map[string]any{"mode": "missing"}
	np := leaf("resource", "service.namespace", "missing", "")
	if mode%2 == 1 {
		ns = map[string]any{"mode": "value", "value": ""}
		np = leaf("resource", "service.namespace", "eq", "")
	}
	if mode == 4 {
		ns = map[string]any{"mode": "value", "value": "team-雪"}
		np = leaf("resource", "service.namespace", "eq", "team-雪")
	}
	group := map[string]string{"kind": "column", "key": "name"}
	if mode%2 == 1 {
		group = map[string]string{"kind": "attribute", "key": "http.route"}
	}
	from := int64(1800000000000 + seed*1000)
	to := from + 60000
	parent := map[string]any{"signal": "spans", "environment": "", "environmentSelector": env, "fromMs": from, "toMs": to, "limit": 100, "offset": 0, "groupBy": []any{group}, "where": map[string]any{"all": []any{leaf("column", "service", "eq", service), np, leaf("column", "kind", "eq", "Server"), leaf("column", "name", "eq", "POST /checkout")}}}
	raw, _ := json.Marshal(parent)
	var q exploration.Query
	json.Unmarshal(raw, &q)
	discoveryQ := map[string]any{}
	for k, v := range parent {
		discoveryQ[k] = v
	}
	discoveryQ["limit"] = 1
	discoveryQ["groupBy"] = []any{map[string]string{"kind": "resource", "key": "service.version"}}
	labels := []any{nil, "", fmt.Sprintf("version-'雪-%d", seed), "  v1  "}
	groups := []any{}
	total := new(big.Int)
	for i, label := range labels {
		n := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 53), big.NewInt(int64(seed+i+1)))
		total.Add(total, n)
		groups = append(groups, map[string]any{"g0": label, "count": n.String()})
	}
	receipt := map[string]any{"readAt": "2026-11-12T12:00:00Z", "basis": fmt.Sprintf("%064x", seed+1)}
	dto := map[string]any{"version": "goo/record-aggregates/v1", "question": discoveryQ, "capturedAt": "2026-11-12T12:00:00Z", "basis": fmt.Sprintf("%064x", seed+2), "count": total.String(), "state": "observed", "groups": groups, "warnings": []string{}, "reads": map[string]any{"count": receipt, "groups": receipt}}
	a, b := labels[mode%4], labels[(mode+1)%4]
	context := map[string]any{"service": service, "scope": map[string]any{"environment": "", "environmentSelector": env, "fromMs": from, "toMs": to, "namespace": ns}}
	original := map[string]any{"question": parent, "state": "observed"}
	return map[string]any{"parent": parent, "parentData": original, "context": context, "discovery": dto, "baseline": a, "selection": b}, q, a, b
}
func side(parent exploration.Query, value any, mean bool) exploration.Query {
	q := parent
	q.Offset = 0
	q.Limit = 100
	q.BucketMS = 0
	q.SpanStatus = false
	q.Summary = ""
	if mean {
		q.Summary = "span-duration"
	}
	predicate := exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: "service.version"}, Op: "missing"}
	if value != nil {
		predicate.Op = "eq"
		predicate.Value = value.(string)
	}
	q.Where = &exploration.Predicate{All: []exploration.Predicate{*parent.Where, predicate}}
	return q
}
func TestVersionComparisonPreservesPopulationAndLiteralIdentity(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		cases := []any{}
		parents := []exploration.Query{}
		values := [][2]any{}
		for mode := range 6 {
			c, q, a, b := fixture(seed, mode)
			cases = append(cases, c)
			parents = append(parents, q)
			values = append(values, [2]any{a, b})
		}
		input, _ := json.Marshal(cases)
		cmd := exec.CommandContext(t.Context(), "node", "bridge.mjs")
		cmd.Stdin = bytes.NewReader(input)
		raw, err := cmd.CombinedOutput()
		if err != nil {
			tc.Fatalf("actual JS builder: %v %s", err, raw)
		}
		var outputs []projection
		if err = json.Unmarshal(raw, &outputs); err != nil {
			tc.Fatalf("projection: %v %s", err, raw)
		}
		if len(outputs) != len(cases) {
			tc.Fatal("missing generated projection")
		}
		for i, p := range outputs {
			want := parents[i]
			want.Limit = 1
			want.GroupBy = []exploration.Field{{Kind: "resource", Key: "service.version"}}
			if !reflect.DeepEqual(p.Question, want) {
				tc.Fatalf("discovery changed fixed parent: got %+v want %+v", p.Question, want)
			}
			for j, view := range []workbench.ExplorationView{p.Count, p.Mean} {
				if err = view.Validate(); err != nil {
					tc.Fatalf("comparison grammar refused: %v", err)
				}
				selection, baseline := side(parents[i], values[i][1], j == 1), side(parents[i], values[i][0], j == 1)
				if len(view.Queries) != 1 || !reflect.DeepEqual(view.Queries[0], selection) {
					tc.Fatalf("selection changed scope or original operation grouping: got %+v want %+v", view.Queries, selection)
				}
				var actual *exploration.Query
				for _, child := range view.Composition.Children {
					if child.Comparison != nil {
						actual = child.Comparison.BaselineQuery
					}
				}
				if actual == nil || !reflect.DeepEqual(*actual, baseline) {
					tc.Fatal("baseline changed parent/typed missing-empty predicate")
				}
				for _, q := range []exploration.Query{p.Question, selection, baseline} {
					if _, err = exploration.Compile(q); err != nil {
						tc.Fatalf("Go compiler refused: %v", err)
					}
				}
			}
		}
	}, hegel.WithTestCases(32), hegel.WithSeed(415), hegel.WithDatabase(filepath.Join(t.TempDir(), "hegel")))
}
