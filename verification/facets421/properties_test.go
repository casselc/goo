package facets421

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"hegel.dev/go/hegel"
)

type sample struct {
	Scope    map[string]any
	Field    map[string]any
	Window   map[string]int64
	Expected map[string]any
	Response map[string]any
}

func TestGeneratedFacetScopesAndExactCounts(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 65535))
		cases := []sample{}
		for i := range 12 {
			signal := "logs"
			if i%2 == 1 {
				signal = "spans"
			}
			kind := []string{"attribute", "resource", "column"}[i%3]
			key := "release.kind"
			if kind == "column" {
				key = "service"
			}
			field := map[string]any{"kind": kind, "key": key}
			from := int64(121000 + seed*1000 + i*60000)
			to := from + 60000
			selector := map[string]any{"version": "goo/environment-selector/v1", "policy": "effective", "scope": map[string]any{"mode": "named", "value": fmt.Sprintf("prod-%d", seed)}}
			if i%4 == 0 {
				selector["scope"] = map[string]any{"mode": "empty"}
			}
			if i%4 == 1 {
				selector["scope"] = map[string]any{"mode": "unreported"}
			}
			if i%4 == 2 {
				selector["policy"] = "current-only"
			}
			where := map[string]any{"all": []any{map[string]any{"field": map[string]any{"kind": "resource", "key": "service.namespace"}, "op": "eq"}, map[string]any{"field": map[string]any{"kind": "attribute", "key": "tenant"}, "op": "eq", "value": fmt.Sprintf("tenant-λ-%d", seed)}, map[string]any{"field": map[string]any{"kind": "attribute", "key": "optional"}, "op": "missing"}}}
			scope := map[string]any{"signal": signal, "environmentSelector": selector, "fromMs": from, "toMs": to, "where": where, "limit": 100, "offset": 100, "groupBy": []any{map[string]any{"kind": "column", "key": "service"}}, "bucketMs": 1000, "numeric": map[string]any{"field": map[string]any{"kind": "attribute", "key": "amount"}, "type": "int64"}}
			expected := map[string]any{"signal": signal, "environmentSelector": selector, "fromMs": from, "toMs": to, "where": where, "limit": 1, "offset": 0, "groupBy": []any{field}}
			base := uint64(seed + 1)
			if i%3 == 0 {
				base += 9007199254740993
			}
			groups := []any{map[string]any{"g0": nil, "count": "1"}, map[string]any{"g0": "", "count": "2"}, map[string]any{"g0": "Missing key", "count": strconv.FormatUint(base, 10)}}
			total := strconv.FormatUint(base+3, 10)
			state := "observed"
			if i == 11 {
				groups = []any{}
				total = "0"
				state = "no_data"
			}
			response := map[string]any{"version": "goo/counted-values/v1", "question": expected, "state": state, "count": total, "groups": groups, "capturedAt": "2026-11-13T00:02:03.123456789Z", "basis": "source-counted-facts", "reads": map[string]any{"count": map[string]any{"readAt": "2026-11-13T00:02:02Z", "basis": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}, "groups": map[string]any{"readAt": "2026-11-13T00:02:03Z", "basis": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}}}
			cases = append(cases, sample{scope, field, map[string]int64{"fromMs": from - 60000, "toMs": from}, expected, response})
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
			tc.Fatalf("actual facet adapters: %v %s", err, stderr.String())
		}
		var count int
		if err = json.Unmarshal(out, &count); err != nil || count != len(cases) {
			tc.Fatalf("bridge lost cases: %s %v", out, err)
		}
	}, hegel.WithTestCases(32), hegel.WithSeed(421), hegel.WithDatabase(filepath.Join(t.TempDir(), "hegel")))
}
