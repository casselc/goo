package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

// Actual production JavaScript builds requests and guards returned side plans.
// The Go planner and an independent expected whole-question transformation check
// preservation, projection and literal filter authority without any source IO.
func TestHegelFreshPairExplanation348(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		profile := hegel.Draw(tc, hegel.Integers(0, 2))
		scope := hegel.Draw(tc, hegel.Integers(0, 4))
		literal := fmt.Sprintf("literal-'\\雪-Ω-%d", seed)
		q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1800000000000 + int64(seed)*1000, ToMS: 1800000060000 + int64(seed)*1000, Limit: 1 + seed%100, Offset: 0, GroupBy: []exploration.Field{{Kind: "column", Key: "service"}}, BucketMS: 1000, Where: &exploration.Predicate{All: []exploration.Predicate{
			{Field: &exploration.Field{Kind: "attribute", Key: "tenant"}, Op: "eq", Value: literal},
			{Any: []exploration.Predicate{{Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "missing"}, {Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "eq", Value: ""}}},
		}}}
		measure := "count"
		if profile == 0 {
			q.Signal = "logs"
		} else if profile == 2 {
			q.Summary = "span-duration"
			measure = "duration.mean"
		}
		if scope > 0 {
			q.Environment = ""
			selector := &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod-'雪"}}
			switch scope {
			case 1:
				selector.Policy = "current-only"
			case 3:
				selector.Scope = identity.EnvironmentScope{Mode: "empty"}
			case 4:
				selector.Scope = identity.EnvironmentScope{Mode: "unreported"}
			}
			q.EnvironmentSelector = selector
		}
		clone := func(x exploration.Query) exploration.Query {
			raw, _ := json.Marshal(x)
			var out exploration.Query
			json.Unmarshal(raw, &out)
			return out
		}
		baseline := clone(q)
		baseline.FromMS -= 60000
		baseline.ToMS -= 60000
		if seed%2 == 0 {
			baseline.FromMS = q.FromMS
			baseline.ToMS = q.ToMS
		}
		baseline.Limit = 1 + (seed+7)%100
		baseline.Where.All[0].Value = "independent-baseline-'雪"
		fields := []exploration.Field{{Kind: "attribute", Key: "cohort"}, {Kind: "resource", Key: "service.version"}}
		declaration := workbench.CountComparisonViewDeclaration{Version: "goo.view/1", Scope: workbench.ViewDeclarationScope{Kind: "record-query", Signal: q.Signal}, Kind: workbench.CountComparisonViewKind{Name: "query-count-comparison", Version: 1}, BaselineQuery: &baseline, Display: workbench.CountComparisonViewDisplay{Value: measure}}
		if err := declaration.Validate(q); err != nil {
			tc.Fatal("generated invalid fresh pair", err)
			return
		}
		payload := struct {
			Selected   exploration.Query                        `json:"selected"`
			Comparison workbench.CountComparisonViewDeclaration `json:"comparison"`
			Fields     []exploration.Field                      `json:"fields"`
			Literal    string                                   `json:"literal"`
		}{q, declaration, fields, literal}
		raw, err := json.Marshal(payload)
		if err != nil {
			tc.Fatal(err)
			return
		}
		cmd := exec.CommandContext(t.Context(), "node", "--no-warnings", "fixtures/fresh-pair-explanation-348.mjs")
		cmd.Stdin = bytes.NewReader(raw)
		out, err := cmd.CombinedOutput()
		if err != nil {
			tc.Fatal(err, string(out))
			return
		}
		var result struct {
			Request, AfterReturnMutation exploration.ExplanationRequest
			Rejects                      []bool
			Filters                      []struct {
				Wire struct {
					Request    exploration.ExplanationRequest
					Field      exploration.Field
					Value      *string
					Side, Mode string
				}
				Question exploration.Query
			}
		}
		if err = json.Unmarshal(out, &result); err != nil {
			tc.Fatal(err, string(out))
			return
		}
		expected := exploration.ExplanationRequest{Version: exploration.ExplanationVersion, Baseline: baseline, Selection: q, Fields: fields}
		if !reflect.DeepEqual(result.Request, expected) || !reflect.DeepEqual(result.AfterReturnMutation, expected) {
			tc.Fatal("UI request weakened original definitions or aliased caller data", result.Request, expected)
			return
		}
		if len(result.Rejects) != 6 || len(result.Filters) != 12 {
			tc.Fatal("unfinished generated request schedule")
			return
		}
		for _, rejected := range result.Rejects {
			if !rejected {
				tc.Fatal("unsupported comparison or request admitted")
				return
			}
		}
		plan, err := exploration.PlanExplanation(result.Request)
		if err != nil {
			tc.Fatal(err)
			return
		}
		if !reflect.DeepEqual(plan.Request, expected) {
			tc.Fatal("Go plan changed original request")
			return
		}
		for i, actual := range []exploration.Query{plan.Baseline, plan.Selection} {
			original := []exploration.Query{baseline, q}[i]
			wanted := clone(original)
			wanted.Numeric = nil
			wanted.NumericHistogram = nil
			wanted.Summary = ""
			wanted.SpanStatus = false
			wanted.BucketMS = 0
			wanted.GroupTimeSeries = false
			wanted.GroupBy = nil
			wanted.Offset = 0
			wanted.Limit = 100
			if !reflect.DeepEqual(actual, wanted) {
				tc.Fatal("derived count acquired wrong scope/profile", actual, wanted)
				return
			}
			sql, err := exploration.Compile(actual)
			if err != nil {
				tc.Fatal(err)
				return
			}
			if sql.Parameters["from"] != strconv.FormatInt(original.FromMS, 10) || sql.Parameters["to"] != strconv.FormatInt(original.ToMS, 10) || strings.Contains(sql.Count, literal) {
				tc.Fatal("compiled count lost fixed bounds or interpolated literal")
				return
			}
		}
		for _, item := range result.Filters {
			wire := item.Wire
			if !reflect.DeepEqual(wire.Request, expected) || wire.Field != fields[0] {
				tc.Fatal("side planning lost original request or qualified field")
				return
			}
			actual, err := exploration.PlanExplanationFilter(wire.Request, wire.Field, wire.Value, wire.Side, wire.Mode)
			if err != nil {
				tc.Fatal(err)
				return
			}
			wanted := clone(q)
			if wire.Side == "baseline" {
				wanted = clone(baseline)
			}
			f := wire.Field
			predicate := exploration.Predicate{Field: &f, Op: "missing"}
			if wire.Value != nil {
				predicate.Op = "eq"
				predicate.Value = *wire.Value
			}
			if wire.Mode == "exclude" {
				if wire.Value == nil {
					predicate.Op = "exists"
				} else {
					predicate = exploration.Predicate{Any: []exploration.Predicate{{Field: &f, Op: "missing"}, {Field: &f, Op: "ne", Value: *wire.Value}}}
				}
			}
			wanted.Where = &exploration.Predicate{All: []exploration.Predicate{*wanted.Where, predicate}}
			wanted.Offset = 0
			if !reflect.DeepEqual(actual, wanted) || !reflect.DeepEqual(item.Question, wanted) {
				tc.Fatal("side result guard or Go filter changed full original profile/policy/predicate", wire.Side, wire.Mode, actual, item.Question, wanted)
				return
			}
		}
		// Planner outputs must not retain aliases into the original two questions.
		plan.Request.Baseline.Where.All[0].Value = "mutated plan"
		if !reflect.DeepEqual(result.Request, expected) {
			tc.Fatal("Go planner aliases original request")
			return
		}
	}, options(40)...)
}
