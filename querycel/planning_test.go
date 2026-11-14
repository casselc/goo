package querycel

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
)

func planningBase() exploration.Query {
	return exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 20, Where: &exploration.Predicate{Any: []exploration.Predicate{
		{Field: &exploration.Field{Kind: "attribute", Key: "team"}, Op: "eq", Value: "edge"},
		{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "api"},
	}}}
}
func TestRecordPlanningBoundPureReceipt(t *testing.T) {
	base := planningBase()
	request, err := NewPlanRequest(base, `log.body == "x' OR 1=1 --"`, "condition", "mount-2", "9007199254740993")
	if err != nil {
		t.Fatal(err)
	}
	original, _ := exploration.CloneQuery(base)
	base.Where.Any[0].Value = "changed"
	result, err := PlanRecord(request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Binding != request.Binding || !reflect.DeepEqual(result.Question.Where.All[0], *original.Where) {
		t.Fatal("binding or complete parent changed")
	}
	if strings.Contains(result.Plan.Rows, "x' OR 1=1 --") {
		t.Fatal("literal interpolated into SQL")
	}
	result.Question.Where.All[0].Any[0].Value = "changed again"
	if request.Base.Where.Any[0].Value != "edge" {
		t.Fatal("result aliases request")
	}
}
func TestRecordPlanningRejectsUnboundAndInvalidInputs(t *testing.T) {
	request, err := NewPlanRequest(planningBase(), `log.body == "ok"`, "condition", "mount", "1")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		edit func(*PlanRequest)
	}{
		{"base", func(r *PlanRequest) { r.Base.Limit++ }},
		{"operation", func(r *PlanRequest) { r.Binding.Operation = "execute" }},
		{"signal", func(r *PlanRequest) { r.Binding.Signal = "spans" }},
		{"library", func(r *PlanRequest) { r.Binding.Library = "other" }},
		{"compiler", func(r *PlanRequest) { r.Binding.Compiler = "other" }},
		{"target", func(r *PlanRequest) { r.Binding.Target = "databricks" }},
		{"profile", func(r *PlanRequest) { r.Binding.Profile = "native" }},
		{"slot", func(r *PlanRequest) { r.Binding.Slot = "" }},
		{"epoch", func(r *PlanRequest) { r.Binding.Epoch = "bad\x00epoch" }},
		{"revision", func(r *PlanRequest) { r.Binding.Revision = strings.Repeat("x", 129) }},
		{"expression", func(r *PlanRequest) { r.Expression = `false && log.body.matches(".*")` }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := request
			tc.edit(&r)
			out, err := PlanRecord(r)
			var diagnostic *Diagnostic
			if out != nil || !errors.As(err, &diagnostic) {
				t.Fatalf("partial receipt or untyped refusal: %v %v", out, err)
			}
		})
	}
	bad := planningBase()
	bad.Where = &exploration.Predicate{All: []exploration.Predicate{}}
	if _, err := BaseIdentity(bad); err == nil {
		t.Fatal("invalid base hashed")
	}
}
