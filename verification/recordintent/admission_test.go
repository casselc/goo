package recordintent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

type admissionCase struct {
	Name  string         `json:"name"`
	Query map[string]any `json:"query"`
	Valid bool           `json:"valid"`
}

func ordinary(signal string, seed int) map[string]any {
	return map[string]any{"signal": signal, "environment": fmt.Sprintf("env-'雪-%d", seed), "fromMs": int64(1800000000000 + seed*1000), "toMs": int64(1800000060000 + seed*1000), "limit": 100, "offset": 0}
}
func field(kind, key string) map[string]any { return map[string]any{"kind": kind, "key": key} }
func leaf(kind, key, op, value string) map[string]any {
	return map[string]any{"field": field(kind, key), "op": op, "value": value}
}
func clone(q map[string]any) map[string]any {
	b, _ := json.Marshal(q)
	var v map[string]any
	_ = json.Unmarshal(b, &v)
	return v
}
func corpus(seed int) []admissionCase {
	var cases []admissionCase
	for _, signal := range []string{"logs", "spans"} {
		base := ordinary(signal, seed)
		add := func(name string, valid bool, change func(map[string]any)) {
			q := clone(base)
			change(q)
			cases = append(cases, admissionCase{signal + "/" + name, q, valid})
		}
		add("ordinary", true, func(q map[string]any) {})
		add("offset-omitted", true, func(q map[string]any) { delete(q, "offset") })
		for _, policy := range []string{"current-only", "effective"} {
			for _, mode := range []string{"named", "empty", "unreported"} {
				if policy == "current-only" && mode != "named" {
					continue
				}
				add("environment/"+policy+"/"+mode, true, func(q map[string]any) {
					delete(q, "environment")
					scope := map[string]any{"mode": mode}
					if mode == "named" {
						scope["value"] = fmt.Sprintf("ops-'雪-%d", seed)
					}
					q["environmentSelector"] = map[string]any{"version": "goo/environment-selector/v1", "policy": policy, "scope": scope}
				})
			}
		}
		for _, n := range []int{1, 100, 0, 101, 200} {
			add(fmt.Sprintf("limit-%d", n), n >= 1 && n <= 100, func(q map[string]any) { q["limit"] = n })
		}
		add("limit-missing", false, func(q map[string]any) { delete(q, "limit") })
		for _, key := range []string{"offset", "limit", "environment", "environmentSelector", "where", "groupBy", "numeric", "numericHistogram", "reportedEdge", "spanStatus", "bucketMs", "summary"} {
			add("explicit-null-"+key, false, func(q map[string]any) { q[key] = nil })
		}
		add("case-alias-query", false, func(q map[string]any) { q["Signal"] = signal })
		add("case-alias-field", false, func(q map[string]any) {
			q["where"] = map[string]any{"field": map[string]any{"kind": "column", "key": "service", "Key": "name"}, "op": "eq", "value": "api"}
		})
		add("empty-branch", false, func(q map[string]any) { q["where"] = map[string]any{"all": []any{}} })
		add("both-branches", false, func(q map[string]any) {
			q["where"] = map[string]any{"all": []any{leaf("column", "service", "eq", "api")}, "any": []any{leaf("column", "service", "eq", "worker")}}
		})

		for _, n := range []int{0, 1000, -1, 1001} {
			add(fmt.Sprintf("offset-%d", n), n >= 0 && n <= 1000, func(q map[string]any) { q["offset"] = n })
		}
		for _, end := range []int64{4102444800000, 4102444800001} {
			add(fmt.Sprintf("end-%d", end), end == 4102444800000, func(q map[string]any) { q["fromMs"] = end - 60000; q["toMs"] = end })
		}
		for _, width := range []int64{7200000, 7200001} {
			add(fmt.Sprintf("width-%d", width), width == 7200000, func(q map[string]any) { q["toMs"] = q["fromMs"].(float64) + float64(width) })
		}
		add("empty-window", false, func(q map[string]any) { q["toMs"] = q["fromMs"] })
		add("key128-utf8", true, func(q map[string]any) { q["where"] = leaf("attribute", strings.Repeat("雪", 42)+"ab", "eq", "") })
		add("key129-utf8", false, func(q map[string]any) { q["where"] = leaf("attribute", strings.Repeat("雪", 42)+"abc", "eq", "") })
		add("value1024-utf8", true, func(q map[string]any) { q["where"] = leaf("resource", "tag", "eq", strings.Repeat("雪", 341)+"a") })
		add("value1025-utf8", false, func(q map[string]any) { q["where"] = leaf("resource", "tag", "eq", strings.Repeat("雪", 341)+"ab") })
		add("literal-nul", false, func(q map[string]any) { q["where"] = leaf("attribute", "tag", "eq", "a\x00b") })
		add("missing-distinct-empty", true, func(q map[string]any) {
			q["where"] = map[string]any{"any": []any{leaf("resource", "service.namespace", "missing", ""), leaf("resource", "service.namespace", "eq", "")}}
		})
		for _, n := range []int{31, 32} {
			add(fmt.Sprintf("root-plus-%d-leaves", n), n == 31, func(q map[string]any) {
				children := make([]any, n)
				for i := range children {
					children[i] = leaf("attribute", fmt.Sprintf("tag%d", i), "eq", fmt.Sprint(seed))
				}
				q["where"] = map[string]any{"all": children}
			})
		}
		for _, depth := range []int{4, 5} {
			add(fmt.Sprintf("depth-%d", depth), depth == 4, func(q map[string]any) {
				var p any = leaf("column", "service", "eq", "api")
				for range depth {
					p = map[string]any{"all": []any{p}}
				}
				q["where"] = p
			})
		}
		for _, key := range []string{"future", "basis", "resultToken"} {
			add("unknown-query-"+key, false, func(q map[string]any) { q[key] = "do-not-drop" })
		}
		add("unknown-field", false, func(q map[string]any) {
			p := leaf("column", "service", "eq", "api")
			p["field"].(map[string]any)["unit"] = "s"
			q["where"] = p
		})
		add("unknown-leaf", false, func(q map[string]any) {
			p := leaf("column", "service", "eq", "api")
			p["token"] = "do-not-drop"
			q["where"] = p
		})
		add("unknown-group", false, func(q map[string]any) {
			q["where"] = map[string]any{"all": []any{leaf("column", "service", "eq", "api")}, "token": "do-not-drop"}
		})
		add("mixed-group-leaf", false, func(q map[string]any) {
			q["where"] = map[string]any{"all": []any{leaf("column", "service", "eq", "api")}, "field": field("column", "service"), "op": "eq", "value": "api"}
		})
		add("unknown-column", false, func(q map[string]any) { q["where"] = leaf("column", "unrecognized", "exists", "") })
		add("fingerprint-classified", true, func(q map[string]any) {
			q["where"] = leaf("column", "exception_fingerprint_v1", "eq", "ef1:c:"+strings.Repeat("a", 64))
		})
		add("fingerprint-unclassified", true, func(q map[string]any) {
			q["where"] = leaf("column", "exception_fingerprint_v1", "eq", "ef1:u:missing-type")
		})
		add("fingerprint-malformed", false, func(q map[string]any) { q["where"] = leaf("column", "exception_fingerprint_v1", "eq", "ef1:c:abc") })
		add("fingerprint-contains", false, func(q map[string]any) {
			q["where"] = leaf("column", "exception_fingerprint_v1", "contains", "ef1:u:missing-type")
		})
		add("body-pattern", signal == "logs", func(q map[string]any) { q["where"] = leaf("column", "body_pattern_v1", "eq", "lp1:p:61n62") })
		add("body-pattern-contains", false, func(q map[string]any) { q["where"] = leaf("column", "body_pattern_v1", "contains", "lp1:e") })

		add("wrong-signal-column", false, func(q map[string]any) {
			k := "duration_ns"
			if signal == "spans" {
				k = "body"
			}
			q["where"] = leaf("column", k, "eq", "0")
		})
		add("unordered-string", false, func(q map[string]any) { q["where"] = leaf("attribute", "tag", "gte", "3") })
		add("uint-max", true, func(q map[string]any) { q["where"] = leaf("column", "timestamp_ns", "gte", "18446744073709551615") })
		add("uint-overflow", false, func(q map[string]any) { q["where"] = leaf("column", "timestamp_ns", "gte", "18446744073709551616") })
		add("uint-contains", false, func(q map[string]any) { q["where"] = leaf("column", "timestamp_ns", "contains", "1") })
		add("group-two", true, func(q map[string]any) {
			q["groupBy"] = []any{field("column", "service"), field("resource", "service.namespace")}
		})
		add("duplicate-group", false, func(q map[string]any) { q["groupBy"] = []any{field("column", "service"), field("column", "service")} })
		add("group-unknown-field", false, func(q map[string]any) {
			q["groupBy"] = []any{map[string]any{"kind": "column", "key": "service", "hidden": true}}
		})
		add("group-three", false, func(q map[string]any) {
			q["groupBy"] = []any{field("column", "service"), field("resource", "one"), field("resource", "two")}
		})
		add("chart240", true, func(q map[string]any) { q["toMs"] = q["fromMs"].(float64) + 240000; q["bucketMs"] = 1000 })
		add("chart241", false, func(q map[string]any) { q["toMs"] = q["fromMs"].(float64) + 240001; q["bucketMs"] = 1000 })
		add("split-missing-groups", false, func(q map[string]any) { q["groupTimeSeries"] = true; q["bucketMs"] = 1000 })
		add("split-missing-bucket", false, func(q map[string]any) { q["groupTimeSeries"] = true; q["groupBy"] = []any{field("column", "service")} })
		add("split", true, func(q map[string]any) {
			q["groupTimeSeries"] = true
			q["bucketMs"] = 1000
			q["groupBy"] = []any{field("column", "service")}
		})
		add("duration-summary", signal == "spans", func(q map[string]any) { q["summary"] = "span-duration" })
		add("span-status", signal == "spans", func(q map[string]any) { q["spanStatus"] = true })
		binding := func(typ string) map[string]any {
			return map[string]any{"field": field("attribute", "quantity"), "type": typ, "unit": "1", "parse": "canonical-decimal"}
		}
		for _, typ := range []string{"int64", "uint64", "decimal64"} {
			add("numeric-"+typ, true, func(q map[string]any) {
				b := binding(typ)
				v := "18446744073709551615"
				if typ == "int64" {
					v = "-9223372036854775808"
				}
				if typ == "decimal64" {
					b["scale"] = 3
					v = "-9223372036854775.808"
				}
				q["where"] = map[string]any{"numeric": b, "op": "gte", "value": v}
			})
		}
		add("numeric-overflow", false, func(q map[string]any) {
			q["where"] = map[string]any{"numeric": binding("int64"), "op": "eq", "value": "9223372036854775808"}
		})
		add("numeric-noncanonical", false, func(q map[string]any) {
			q["where"] = map[string]any{"numeric": binding("int64"), "op": "eq", "value": "01"}
		})
		add("numeric-state", true, func(q map[string]any) { q["where"] = map[string]any{"numeric": binding("uint64"), "op": "invalid"} })
		add("numeric-unit-spaces", false, func(q map[string]any) { b := binding("uint64"); b["unit"] = "bytes per second"; q["numeric"] = b })
		add("numeric-unknown-binding", false, func(q map[string]any) { b := binding("uint64"); b["hidden"] = true; q["numeric"] = b })
		add("numeric-histogram", true, func(q map[string]any) {
			q["numeric"] = binding("uint64")
			q["numericHistogram"] = map[string]any{"binning": "powers-of-two"}
		})
		add("numeric-histogram-unbound", false, func(q map[string]any) { q["numericHistogram"] = map[string]any{"binning": "powers-of-two"} })
		edge := func() map[string]any {
			return map[string]any{"parent": map[string]any{"service": "api", "namespace": map[string]any{"mode": "missing"}}, "child": map[string]any{"service": "worker", "namespace": map[string]any{"mode": "value", "value": ""}}}
		}
		add("reported-edge", signal == "spans", func(q map[string]any) { q["reportedEdge"] = edge() })
		add("edge-same-owner", false, func(q map[string]any) {
			e := edge()
			e["child"] = clone(e["parent"].(map[string]any))
			q["reportedEdge"] = e
		})
		add("edge-namespace-conflict", false, func(q map[string]any) {
			e := edge()
			e["discoveryNamespace"] = map[string]any{"mode": "missing"}
			q["reportedEdge"] = e
		})
		add("edge-unknown-service-field", false, func(q map[string]any) {
			e := edge()
			e["parent"].(map[string]any)["version"] = "v1"
			q["reportedEdge"] = e
		})
	}
	return cases
}
func goAdmission(raw []byte) error {
	var q exploration.Query
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&q); err != nil {
		return err
	}
	return q.Validate()
}
func TestGeneratedRecordIntentAdmissionAgreesWithStrictGo(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		cases := corpus(seed)
		for _, c := range cases {
			raw, _ := json.Marshal(c.Query)
			err := goAdmission(raw)
			if (err == nil) != c.Valid {
				tc.Fatalf("independent Go expectation %s: valid=%v err=%v", c.Name, c.Valid, err)
			}
		}
		raw, _ := json.Marshal(cases)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "node", "bridge.mjs")
		cmd.Stdin = bytes.NewReader(raw)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		output, err := cmd.Output()
		if err != nil {
			tc.Fatalf("actual JS admission: %v\n%s", err, stderr.String())
		}
		var results []struct {
			Name  string `json:"name"`
			Valid bool   `json:"valid"`
			Error string `json:"error"`
		}
		if err = json.Unmarshal(output, &results); err != nil {
			tc.Fatalf("bridge JSON: %v", err)
		}
		if len(results) != len(cases) {
			tc.Fatalf("bridge lost cases: %d vs %d", len(results), len(cases))
		}
		for i, r := range results {
			if r.Name != cases[i].Name || r.Valid != cases[i].Valid {
				tc.Fatalf("JS/Go admission differs %s: expected%v got%v (%s)", cases[i].Name, cases[i].Valid, r.Valid, r.Error)
			}
		}
	}, hegel.WithTestCases(32), hegel.WithSeed(417), hegel.WithDatabase(filepath.Join(t.TempDir(), "hegel")))
}
