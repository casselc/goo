package verification

import (
	"bytes"
	"encoding/json"
	"math/big"
	"os/exec"
	"reflect"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
)

// Exercise the actual Go range planner, browser whole-question protocol and Go
// admission. The membership oracle uses independently constructed integer edges.
func TestHegelTypedDistributionContinuation372(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		mode := hegel.Draw(tc, hegel.Integers(0, 3))
		b := exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "payload.size"}, Type: "int64", Unit: "items", Parse: "canonical-decimal"}
		parent := histogramParent(b)
		parent.Environment = ""
		parent.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "production-" + strconv.Itoa(seed)}}
		if mode == 0 {
			parent.EnvironmentSelector.Policy = "current-only"
		} else if mode == 2 {
			parent.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "empty"}
		} else if mode == 3 {
			parent.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "unreported"}
		}
		ns := &exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "eq", Value: ""}
		if seed%3 == 0 {
			ns.Op = "missing"
		} else if seed%3 == 2 {
			ns.Value = "team-雪"
		}
		parent.Where = &exploration.Predicate{All: []exploration.Predicate{*parent.Where, *ns}}
		parent.Offset = seed % 100
		parent.Limit = 1 + seed%100
		parent.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}}
		parent.BucketMS = 1000
		from := hegel.Draw(tc, hegel.Integers(0, 127))
		to := hegel.Draw(tc, hegel.Integers(from+1, 128))
		plan, err := exploration.PlanNumericHistogramSelection(parent, from, to)
		if err != nil {
			tc.Fatal(err)
		}
		data, err := json.Marshal(plan)
		if err != nil {
			tc.Fatal(err)
		}
		cmd := exec.CommandContext(t.Context(), "node", "typed_distribution_372_fixture.mjs")
		cmd.Stdin = bytes.NewReader(data)
		out, err := cmd.CombinedOutput()
		if err != nil {
			tc.Fatalf("actual JS continuation: %v\n%s", err, out)
		}
		var actual struct {
			Selected, Rest exploration.Query
			Checks         int
		}
		if err = json.Unmarshal(out, &actual); err != nil || actual.Checks != 6 || !reflect.DeepEqual(actual.Selected, plan.Selected) || !reflect.DeepEqual(actual.Rest, plan.Rest) {
			tc.Fatalf("whole question changed: %v\n%s", err, out)
		}
		for _, q := range []exploration.Query{actual.Selected, actual.Rest} {
			if _, err = exploration.Compile(q); err != nil {
				tc.Fatal(err)
			}
			if !reflect.DeepEqual(q.EnvironmentSelector, parent.EnvironmentSelector) || q.FromMS != parent.FromMS || q.ToMS != parent.ToMS || q.Offset != 0 || q.Limit != parent.Limit || !reflect.DeepEqual(q.Numeric, parent.Numeric) || !reflect.DeepEqual(q.NumericHistogram, parent.NumericHistogram) || !reflect.DeepEqual(q.GroupBy, parent.GroupBy) || q.BucketMS != parent.BucketMS {
				tc.Fatal("typed continuation lost original definition")
			}
		}
		edges := histogramOracleEdges("int64")
		lo, _ := new(big.Int).SetString(edges[from], 10)
		hi, _ := new(big.Int).SetString(edges[to], 10)
		for _, raw := range histogramBoundaryValues("int64") {
			for _, inParent := range []bool{false, true} {
				row := predicateValues(b, raw)
				row[exploration.Field{Kind: "column", Key: "service"}] = "allowed"
				if !inParent {
					row[exploration.Field{Kind: "column", Key: "service"}] = "other"
				}
				if ns.Op != "missing" {
					row[*ns.Field] = ns.Value
				}
				valid := predicateState(raw, "int64") == "valid"
				inside := false
				if valid {
					v, _ := new(big.Int).SetString(*raw, 10)
					inside = v.Cmp(lo) >= 0 && v.Cmp(hi) < 0
				}
				selected := exploration.Match(actual.Selected.Where, row)
				rest := exploration.Match(actual.Rest.Where, row)
				if selected != (inParent && inside) || rest != (inParent && !inside) {
					tc.Fatal("exact typed partition changed", seed, from, to, raw, selected, rest)
				}
			}
		}
	}, options(40)...)
}
