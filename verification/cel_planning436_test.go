package verification

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/querycel"
	"hegel.dev/go/hegel"
)

// Exercise the actual shared planner against independent parent AND CEL membership.
func TestHegelCELPlanning436Conservation(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		signal := hegel.Draw(tc, hegel.SampledFrom([]string{"logs", "spans"}))
		literal := hegel.Draw(tc, hegel.SampledFrom([]string{"needle' OR 1=1 --", "needle雪_%", "needle\nline"}))
		threshold := hegel.Draw(tc, hegel.SampledFrom([]uint64{0, 17, 255, 9007199254740993, ^uint64(0)}))
		revision := hegel.Draw(tc, hegel.SampledFrom([]string{"9007199254740993", "18446744073709551615", "draft-雪-3"}))
		field, numeric, variable, numvar := "body", "severity_number", "log.body", "log.severity_number"
		if signal == "spans" {
			field, numeric, variable, numvar = "name", "duration_ns", "span.name", "span.duration_ns"
		}
		expression := variable + ".contains(" + celLiteral(literal) + ") && " + numvar + " >= " + strconv.FormatUint(threshold, 10) + "u"
		parent := exploration.Query{Signal: signal, Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: hegel.Draw(tc, hegel.Integers(1, 50)), Offset: hegel.Draw(tc, hegel.Integers(0, 100)), GroupBy: []exploration.Field{{Kind: "column", Key: "service"}}, Where: &exploration.Predicate{Any: []exploration.Predicate{{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "api"}, {Field: &exploration.Field{Kind: "resource", Key: "region"}, Op: "missing"}}}}
		request, err := querycel.NewPlanRequest(parent, expression, "condition-1", "mount-1", revision)
		if err != nil {
			tc.Fatal(err)
		}
		result, err := querycel.PlanRecord(request)
		if err != nil {
			tc.Fatal(err)
		}
		if result.Binding != request.Binding || result.Expression != expression {
			tc.Fatal("planning lost exact binding or source")
		}
		wire, err := json.Marshal(result)
		if err != nil {
			tc.Fatal(err)
		}
		var reopened querycel.RecordPlan
		if err = json.Unmarshal(wire, &reopened); err != nil || reopened.Binding != request.Binding {
			tc.Fatal("wire rounded or changed binding", err)
		}
		actualPolicy, wantPolicy := result.Question, parent
		actualPolicy.Where = nil
		wantPolicy.Where = nil
		if !reflect.DeepEqual(actualPolicy, wantPolicy) || !reflect.DeepEqual(result.Question.Where.All[0], *parent.Where) {
			tc.Fatal("planning changed parent policy or OR")
		}
		refinement436Detached(tc, result.Question, request.Base)
		refinement436Detached(tc, request.Base, parent)
		program, err := querycel.Compile(signal, expression)
		if err != nil {
			tc.Fatal(err)
		}
		for _, service := range []string{"api", "other"} {
			for _, region := range []string{"missing", "present"} {
				for _, body := range []string{literal, "ordinary"} {
					for _, number := range []uint64{0, 17, 255, 9007199254740993, ^uint64(0)} {
						record := celRecord(number, body, service)
						record[exploration.Field{Kind: "column", Key: field}] = body
						record[exploration.Field{Kind: "column", Key: numeric}] = strconv.FormatUint(number, 10)
						if region == "present" {
							record[exploration.Field{Kind: "resource", Key: "region"}] = "west"
						}
						condition, err := program.Eval(record)
						if err != nil {
							tc.Fatal(err)
						}
						if got, want := exploration.Match(result.Question.Where, record), exploration.Match(parent.Where, record) && condition; got != want {
							tc.Fatal("planned membership changed original parent")
						}
					}
				}
			}
		}
		sql := result.Plan.Rows + result.Plan.Count + result.Plan.Groups + result.Plan.Fields + result.Plan.TimeSeries
		if strings.Contains(sql, literal) || strings.Contains(sql, strconv.FormatUint(^uint64(0), 10)) {
			tc.Fatal("literal interpolated into target SQL")
		}
		literalBound, numberBound := false, false
		for _, value := range result.Plan.Parameters {
			literalBound = literalBound || value == literal
			numberBound = numberBound || value == strconv.FormatUint(threshold, 10)
		}
		if !literalBound || !numberBound {
			tc.Fatal("exact literal omitted from typed parameters")
		}
		changed := request
		changed.Base.Environment = "staging"
		rejected, err := querycel.PlanRecord(changed)
		var diagnostic *querycel.Diagnostic
		if rejected != nil || !errors.As(err, &diagnostic) || diagnostic.Path != "binding.baseIdentity" {
			tc.Fatal("changed base accepted stale binding", err)
		}
		for _, unsupported := range []string{`false && log.body.matches(".*")`, `true || span.name.matches(".*")`} {
			invalid := request
			invalid.Expression = unsupported
			rejected, err = querycel.PlanRecord(invalid)
			if rejected != nil || !errors.As(err, &diagnostic) {
				tc.Fatal("dead unsupported branch returned partial admitted result", err)
			}
		}
	}, options(120)...)
}

func TestCELPlanning436ExactUint64(t *testing.T) {
	for _, signal := range []string{"logs", "spans"} {
		for _, number := range []uint64{9007199254740993, ^uint64(0)} {
			field := "log.severity_number"
			if signal == "spans" {
				field = "span.duration_ns"
			}
			expression := field + " == " + strconv.FormatUint(number, 10) + "u"
			request, err := querycel.NewPlanRequest(exploration.Query{Signal: signal, Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 1}, expression, "slot", "mount", "18446744073709551615")
			if err != nil {
				t.Fatal(err)
			}
			result, err := querycel.PlanRecord(request)
			if err != nil {
				t.Fatal(err)
			}
			exact := strconv.FormatUint(number, 10)
			found := false
			for _, value := range result.Plan.Parameters {
				found = found || value == exact
			}
			if !found || strings.Contains(result.Plan.Rows, exact) {
				t.Fatal("uint64 precision or parameterization lost", number)
			}
		}
	}
}
