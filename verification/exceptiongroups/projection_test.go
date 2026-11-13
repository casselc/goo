package exceptiongroups

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
	"hegel.dev/go/hegel"
)

func leaf(kind, key, op, value string) map[string]any {
	p := map[string]any{"field": map[string]string{"kind": kind, "key": key}, "op": op}
	if op != "missing" {
		p["value"] = value
	}
	return p
}
func jsonQuery(tc interface{ Fatalf(string, ...any) }, v any) exploration.Query {
	b, _ := json.Marshal(v)
	var q exploration.Query
	if e := json.Unmarshal(b, &q); e != nil {
		tc.Fatalf("query decode: %v", e)
	}
	return q
}
func queryFixture(seed, mode int) (map[string]any, exploration.Query, exploration.Query) {
	scope := map[string]any{"mode": []string{"named", "empty", "unreported"}[mode%3]}
	if mode%3 == 0 {
		scope["value"] = fmt.Sprintf("env-'雪-%d", seed)
	}
	policy := "effective"
	if mode == 3 {
		policy = "current-only"
	}
	signal := "logs"
	if mode%2 == 1 {
		signal = "spans"
	}
	env := map[string]any{"version": "goo/environment-selector/v1", "policy": policy, "scope": scope}
	from := int64(1800000000000 + seed*1000)
	to := from + 60000
	ns := leaf("resource", "service.namespace", "missing", "")
	if mode%2 == 1 {
		ns = leaf("resource", "service.namespace", "eq", "")
	}
	where := map[string]any{"all": []any{leaf("column", "service", "eq", fmt.Sprintf("worker-'雪-%d", seed)), ns, map[string]any{"any": []any{leaf("attribute", "team", "eq", "platform"), leaf("resource", "release.channel", "eq", "canary")}}}}
	parent := map[string]any{"signal": signal, "environment": "", "environmentSelector": env, "fromMs": from, "toMs": to, "offset": 0, "limit": 100, "where": where, "groupBy": []any{map[string]string{"kind": "column", "key": "service"}}, "bucketMs": 10000}
	if signal == "spans" {
		parent["summary"] = "span-duration"
		parent["spanStatus"] = true
	}
	projected := map[string]any{}
	for k, v := range parent {
		projected[k] = v
	}
	for _, k := range []string{"groupBy", "bucketMs", "groupTimeSeries", "summary", "spanStatus"} {
		delete(projected, k)
	}
	typ := fmt.Sprintf("Failure-'雪-%d", seed)
	stack := "at request.go:42\n at worker.go:12\t(literal 123)"
	id := exactID(typ, stack)
	n := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 53), big.NewInt(int64(seed+1)))
	total := new(big.Int).Add(new(big.Int).Set(n), big.NewInt(2))
	first, last := fmt.Sprint(from*1000000), fmt.Sprint(to*1000000-1)
	groups := []any{map[string]any{"fingerprint": id, "state": "classified", "reason": "", "count": n.String(), "firstNS": first, "lastNS": last, "inputs": map[string]string{"type": typ, "stacktrace": stack}, "affected": []any{map[string]any{"service": "worker", "namespace": nil, "version": ""}}, "affectedCount": "1", "affectedLimited": false}, map[string]any{"fingerprint": "ef1:u:missing-type", "state": "unclassified", "reason": "missing-type", "count": "2", "firstNS": first, "lastNS": last, "inputs": nil, "affected": []any{map[string]any{"service": "worker", "namespace": "", "version": nil}, map[string]any{"service": "worker", "namespace": nil, "version": ""}}, "affectedCount": "2", "affectedLimited": false}}
	receipt := map[string]any{"readAt": "2026-11-13T00:00:00.000000001Z", "basis": fmt.Sprintf("%064x", seed)}
	dto := map[string]any{"version": "goo/exception-groups/v1", "question": projected, "count": total.String(), "state": "observed", "capturedAt": "2026-11-13T00:00:00.000000002Z", "basis": fmt.Sprintf("%064x", seed+1), "reads": map[string]any{"count": receipt, "groups": receipt}, "groups": groups, "warnings": []string{}}
	chosen := id
	if mode%3 == 1 {
		chosen = "ef1:u:missing-type"
	}
	staged := map[string]any{}
	for k, v := range projected {
		staged[k] = v
	}
	staged["where"] = map[string]any{"all": []any{where, leaf("column", "exception_fingerprint_v1", "eq", chosen)}}
	return map[string]any{"parent": parent, "dto": dto, "fingerprint": chosen}, jsonQuery(nilFatal{}, projected), jsonQuery(nilFatal{}, staged)
}

type nilFatal struct{}

func (nilFatal) Fatalf(format string, args ...any) { panic(fmt.Sprintf(format, args...)) }
func TestExceptionUIProjectionPreservesFixedScopeAndUnclassifiedGroups(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		input := []any{}
		expected := [][2]exploration.Query{}
		for mode := range 6 {
			v, q, stage := queryFixture(seed, mode)
			input = append(input, v)
			expected = append(expected, [2]exploration.Query{q, stage})
		}
		body, _ := json.Marshal(input)
		cmd := exec.Command("node", "bridge.mjs")
		cmd.Stdin = bytes.NewReader(body)
		var errOutput bytes.Buffer
		cmd.Stderr = &errOutput
		raw, err := cmd.Output()
		if err != nil {
			tc.Fatalf("actual JS projection: %v\n%s", err, errOutput.String())
		}
		var actual []struct{ Question, Staged exploration.Query }
		if err = json.Unmarshal(raw, &actual); err != nil {
			tc.Fatalf("bridge output: %v", err)
		}
		if len(actual) != len(expected) {
			tc.Fatalf("projection count differs")
		}
		for i, result := range actual {
			for j, q := range []exploration.Query{result.Question, result.Staged} {
				if !reflect.DeepEqual(q, expected[i][j]) {
					tc.Fatalf("scope changed case%dside%d: got%+v want%+v", i, j, q, expected[i][j])
				}
				if _, err := exploration.Compile(q); err != nil {
					tc.Fatalf("Go ordinary compiler refused projected scope: %v", err)
				}
			}
			plan, err := exploration.CompileExceptionGroups(result.Question)
			if err != nil || plan.Count == "" || plan.Groups == "" {
				tc.Fatalf("metadata compiler refused projected scope: %v", err)
			}
		}
	}, hegel.WithTestCases(32), hegel.WithSeed(416), hegel.WithDatabase(filepath.Join(t.TempDir(), "hegel")))
}
