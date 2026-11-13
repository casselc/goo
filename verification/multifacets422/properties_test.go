package multifacets422

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"hegel.dev/go/hegel"
)

type transition struct {
	Index    int
	Selected bool
	Clear    bool
}
type sample struct {
	Signal       string
	Field, Scope map[string]any
	Values       []any
	Steps        []transition
	Count        string
}

func TestGeneratedSelectionStateAndPredicateMeaning(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 65535))
		rng := rand.New(rand.NewPCG(uint64(seed), 422))
		cases := []sample{}
		for i := range 12 {
			signal := "logs"
			if i%2 == 1 {
				signal = "spans"
			}
			kind := []string{"attribute", "resource", "column"}[i%3]
			key := "cohort"
			if kind == "column" {
				key = "service"
			}
			field := map[string]any{"kind": kind, "key": key}
			tenant := map[string]any{"field": map[string]any{"kind": "attribute", "key": "tenant"}, "op": "eq", "value": fmt.Sprintf("tenant-λ-%d", seed)}
			ns := map[string]any{"field": map[string]any{"kind": "resource", "key": "service.namespace"}, "op": "eq", "value": ""}
			where := map[string]any{"all": []any{tenant, ns}}
			if i%4 == 0 {
				where["all"] = append(where["all"].([]any), map[string]any{"field": field, "op": "exists"})
			}
			scope := map[string]any{"signal": signal, "environmentSelector": map[string]any{"version": "goo/environment-selector/v1", "policy": "effective", "scope": map[string]any{"mode": "named", "value": fmt.Sprintf("prod-%d", seed)}}, "fromMs": 1000 + seed*1000, "toMs": 61000 + seed*1000, "where": where}
			values := []any{nil, "", "Missing key", "canary", fmt.Sprintf("λ-%d", seed)}
			steps := []transition{{Index: 0, Selected: true}, {Index: 1, Selected: true}, {Index: 1, Selected: true}, {Index: 0, Selected: false}}
			for range 32 {
				steps = append(steps, transition{Index: rng.IntN(len(values)), Selected: rng.IntN(2) == 1, Clear: rng.IntN(11) == 0})
			}
			cases = append(cases, sample{signal, field, scope, values, steps, fmt.Sprintf("%d", uint64(9007199254740993)+uint64(seed))})
		}
		raw, err := json.Marshal(cases)
		if err != nil {
			tc.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "node", "bridge.mjs")
		cmd.Stdin = bytes.NewReader(raw)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		out, err := cmd.Output()
		if err != nil {
			tc.Fatalf("actual multi-value predicates: %v %s", err, stderr.String())
		}
		var count int
		if err = json.Unmarshal(out, &count); err != nil || count != len(cases) {
			tc.Fatalf("lost generated cases: %s %v", out, err)
		}
	}, hegel.WithTestCases(32), hegel.WithSeed(422), hegel.WithDatabase(filepath.Join(t.TempDir(), "hegel")))
}
