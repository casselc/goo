package servicetriage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

type group struct {
	G0       string   `json:"g0"`
	G1       *string  `json:"g1"`
	Count    string   `json:"count"`
	Duration duration `json:"duration"`
	Status   status   `json:"spanStatus"`
}
type duration struct {
	State string `json:"state"`
	Count string `json:"count"`
	Sum   string `json:"sumNs"`
	Min   string `json:"minNs"`
	P50   string `json:"p50Ns"`
	P95   string `json:"p95Ns"`
	P99   string `json:"p99Ns"`
	Max   string `json:"maxNs"`
}
type status struct {
	Count   string `json:"count"`
	Error   string `json:"error"`
	OK      string `json:"ok"`
	Unset   string `json:"unset"`
	Unknown string `json:"unknown"`
}
type fraction struct {
	N string `json:"n"`
	D string `json:"d"`
}
type outputRow struct {
	Key, Service              string
	Namespace                 map[string]string
	Current, Baseline         *group
	CountDelta                *string
	ErrorDelta, DurationDelta *fraction
	Question                  exploration.Query
}
type projection struct {
	Plan struct {
		Scope             map[string]any
		Current, Baseline exploration.Query
	}
	Rows   []outputRow
	Orders map[string][]string
}

func integer(s string) *big.Int {
	n, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic("noninteger fixture")
	}
	return n
}
func key(service string, ns *string) string {
	mode, value := "missing", ""
	if ns != nil {
		mode, value = "value", *ns
	}
	b, _ := json.Marshal([]string{service, mode, value})
	return string(b)
}
func groups(seed int, prior bool) []group {
	out := []group{}
	empty, named := "", "team-'雪"
	namespaces := []*string{nil, &empty, &named}
	for i, ns := range namespaces {
		n := new(big.Int).Add(integer("9007199254740993"), big.NewInt(int64(seed+i)))
		if prior {
			n.Add(n, big.NewInt(int64((i-1)*seed)))
		}
		d := int64(1000001 + i*1000000)
		sum := new(big.Int).Mul(n, big.NewInt(d))
		errors := new(big.Int).Add(new(big.Int).Div(new(big.Int).Set(n), big.NewInt(2)), big.NewInt(int64(i)))
		ok := new(big.Int).Sub(new(big.Int).Set(n), errors)
		v := fmt.Sprint(d)
		out = append(out, group{G0: "api-'雪", G1: ns, Count: n.String(), Duration: duration{"observed", n.String(), sum.String(), v, v, v, v, v}, Status: status{n.String(), errors.String(), ok.String(), "0", "0"}})
	}
	return out
}
func dto(gs []group) map[string]any {
	n, sum, errors, ok := new(big.Int), new(big.Int), new(big.Int), new(big.Int)
	for _, g := range gs {
		n.Add(n, integer(g.Count))
		sum.Add(sum, integer(g.Duration.Sum))
		errors.Add(errors, integer(g.Status.Error))
		ok.Add(ok, integer(g.Status.OK))
	}
	quantile := func(percent int64) string {
		rank := new(big.Int).Div(new(big.Int).Mul(new(big.Int).Set(n), big.NewInt(percent)), big.NewInt(100))
		seen := new(big.Int)
		for _, g := range gs {
			seen.Add(seen, integer(g.Count))
			if rank.Cmp(seen) < 0 {
				return g.Duration.Min
			}
		}
		return gs[len(gs)-1].Duration.Max
	}
	read := map[string]any{"readAt": "2026-11-12T12:00:00Z", "basis": strings.Repeat("a", 64)}
	return map[string]any{"version": "goo/record-aggregates/v1", "capturedAt": "2026-11-12T12:00:00Z", "basis": strings.Repeat("b", 64), "state": "observed", "count": n.String(), "groups": gs, "duration": duration{"observed", n.String(), sum.String(), gs[0].Duration.Min, quantile(50), quantile(95), quantile(99), gs[len(gs)-1].Duration.Max}, "spanStatus": status{n.String(), errors.String(), ok.String(), "0", "0"}, "warnings": []string{}, "reads": map[string]any{"count": read, "groups": read}}
}
func TestQualifiedTriageProjectionAndExactRanking(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		width := int64(1000 + seed%60000)
		scope := map[string]any{"environment": "prod-'雪", "fromMs": 3*width + int64(seed), "toMs": 4*width + int64(seed)}
		if seed%2 == 0 {
			scope = map[string]any{"environment": "", "environmentSelector": map[string]any{"version": "goo/environment-selector/v1", "policy": "effective", "scope": map[string]any{"mode": "unreported"}}, "fromMs": 3*width + int64(seed), "toMs": 4*width + int64(seed)}
		}
		current, prior := groups(seed, false), groups(seed, true)
		metric := map[string]any{"service": "metrics-only", "namespace": map[string]string{"mode": "missing"}, "observed": true, "signals": map[string]any{"metrics": map[string]string{"count": "9"}}, "declarations": []any{}}
		input, _ := json.Marshal([]any{map[string]any{"scope": scope, "current": dto(current), "baseline": dto(prior), "catalog": []any{metric}}})
		cmd := exec.CommandContext(t.Context(), "node", "bridge.mjs")
		cmd.Stdin = bytes.NewReader(input)
		raw, err := cmd.CombinedOutput()
		if err != nil {
			tc.Fatalf("actual JS model: %v %s", err, raw)
		}
		var got []projection
		if err = json.Unmarshal(raw, &got); err != nil || len(got) != 1 {
			tc.Fatalf("model bridge: %v %s", err, raw)
		}
		p := got[0]
		if p.Plan.Current.ToMS-p.Plan.Current.FromMS != width || p.Plan.Baseline.ToMS != p.Plan.Current.FromMS || p.Plan.Baseline.FromMS != p.Plan.Current.FromMS-width {
			tc.Fatal("comparison windows changed")
		}
		for _, q := range []exploration.Query{p.Plan.Current, p.Plan.Baseline} {
			if err = exploration.ValidateRecordAggregatesQuery(q); err != nil {
				tc.Fatal(err)
			}
			compiled, e := exploration.Compile(q)
			if e != nil {
				tc.Fatal(e)
			}
			if strings.Contains(compiled.Groups, "prod-'雪") {
				tc.Fatal("scope interpolated into SQL")
			}
			if !reflect.DeepEqual(q.GroupBy, []exploration.Field{{Kind: "column", Key: "service"}, {Kind: "resource", Key: "service.namespace"}}) || !q.SpanStatus || q.Summary != "span-duration" {
				tc.Fatal("qualified triage projection changed")
			}
		}
		if len(p.Rows) != 4 {
			tc.Fatal("three qualified rows plus metric-only row required")
		}
		byKey := map[string]outputRow{}
		for _, r := range p.Rows {
			if _, exists := byKey[r.Key]; exists {
				tc.Fatal("namespace identities collapsed")
			}
			byKey[r.Key] = r
		}
		for i, g := range current {
			r, exists := byKey[key(g.G0, g.G1)]
			if !exists || r.Current == nil || r.Baseline == nil {
				tc.Fatal("qualified service lost")
			}
			b := prior[i]
			delta := new(big.Int).Sub(integer(g.Count), integer(b.Count))
			if r.CountDelta == nil || *r.CountDelta != delta.String() {
				tc.Fatal("rounded count delta")
			}
			checkDelta(tc, r.ErrorDelta, g.Status.Error, g.Count, b.Status.Error, b.Count)
			checkDelta(tc, r.DurationDelta, g.Duration.Sum, g.Count, b.Duration.Sum, b.Count)
			vals := map[exploration.Field]string{{Kind: "column", Key: "service"}: g.G0, {Kind: "column", Key: "kind"}: "Server"}
			if g.G1 != nil {
				vals[exploration.Field{Kind: "resource", Key: "service.namespace"}] = *g.G1
			}
			match := exploration.Match(r.Question.Where, vals)
			if !match {
				tc.Fatal("drilldown membership changed")
			}
			vals[exploration.Field{Kind: "column", Key: "kind"}] = "Client"
			match = exploration.Match(r.Question.Where, vals)
			if match {
				tc.Fatal("client span admitted into Server drilldown")
			}
		}
		metricRow := byKey[key("metrics-only", nil)]
		if metricRow.Current != nil || metricRow.Baseline != nil || metricRow.CountDelta != nil {
			tc.Fatal("metric data became server evidence")
		}
		for _, metric := range []string{"count", "error-rate", "mean", "p95", "count-change"} {
			want := slices.Clone(p.Rows)
			slices.SortStableFunc(want, func(a, b outputRow) int {
				aN, aD, aOK := measure(a, metric)
				bN, bD, bOK := measure(b, metric)
				if aOK != bOK {
					if aOK {
						return -1
					}
					return 1
				}
				if aOK {
					x := new(big.Int).Mul(aN, bD)
					y := new(big.Int).Mul(bN, aD)
					if c := y.Cmp(x); c != 0 {
						return c
					}
				}
				return strings.Compare(a.Key, b.Key)
			})
			keys := make([]string, len(want))
			for i, r := range want {
				keys[i] = r.Key
			}
			if !reflect.DeepEqual(p.Orders[metric], keys) {
				tc.Fatalf("%s ranking rounded or unstable: got%v want%v", metric, p.Orders[metric], keys)
			}
		}
	}, hegel.WithTestCases(64), hegel.WithSeed(412), hegel.WithDatabase(filepath.Join(t.TempDir(), "hegel")))
}
func checkDelta(t *hegel.T, got *fraction, a, b, c, d string) {
	if got == nil {
		t.Fatal("missing comparable delta")
	}
	wantN := new(big.Int).Sub(new(big.Int).Mul(integer(a), integer(d)), new(big.Int).Mul(integer(c), integer(b)))
	wantD := new(big.Int).Mul(integer(b), integer(d))
	if new(big.Int).Mul(integer(got.N), wantD).Cmp(new(big.Int).Mul(wantN, integer(got.D))) != 0 || integer(got.D).Sign() <= 0 {
		t.Fatal("rational delta changed")
	}
}
func measure(r outputRow, metric string) (*big.Int, *big.Int, bool) {
	one := big.NewInt(1)
	if metric == "count-change" {
		if r.CountDelta == nil {
			return nil, nil, false
		}
		return integer(*r.CountDelta), one, true
	}
	if r.Current == nil {
		return nil, nil, false
	}
	g := r.Current
	switch metric {
	case "count":
		return integer(g.Count), one, true
	case "error-rate":
		return integer(g.Status.Error), integer(g.Count), true
	case "mean":
		return integer(g.Duration.Sum), integer(g.Count), true
	case "p95":
		return integer(g.Duration.P95), one, true
	}
	panic(metric)
}
