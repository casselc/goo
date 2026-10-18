package exploration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestRecipesPreservePopulationAndCompile(t *testing.T) {
	for _, r := range Recipes() {
		t.Run(r.ID, func(t *testing.T) {
			f := Field{Kind: "attribute", Key: "tenant.id"}
			q := Query{Signal: r.Signal, Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 3, Offset: 100, BucketMS: 1000, Where: &Predicate{Any: []Predicate{{Field: &f, Op: "eq", Value: "x' OR 1=1"}, {Field: &f, Op: "missing"}}}}
			before, _ := json.Marshal(q)
			p, e := PlanRecipe(r.ID, q)
			if e != nil {
				t.Fatal(e)
			}
			if p.Question.Signal != q.Signal || p.Question.Environment != q.Environment || p.Question.FromMS != q.FromMS || p.Question.ToMS != q.ToMS || !reflect.DeepEqual(p.Question.Where, q.Where) || p.Question.BucketMS != q.BucketMS || p.Question.Offset != 0 || p.Question.Limit != 100 {
				t.Fatal("population or page contract changed")
			}
			if strings.Contains(p.Plan.Count, "x' OR 1=1") {
				t.Fatal("unparameterized selector")
			}
			p.Question.Where.Any[0].Value = "changed"
			after, _ := json.Marshal(q)
			if string(before) != string(after) {
				t.Fatal("input aliased")
			}
			if _, e = PlanRecipe("bogus", q); e == nil {
				t.Fatal("unknown accepted")
			}
			if q.Signal == "logs" {
				q.Signal = "spans"
			} else {
				q.Signal = "logs"
			}
			q.Where = nil
			if _, e = PlanRecipe(r.ID, q); e == nil {
				t.Fatal("implicit signal translation")
			}
		})
	}
}

func TestExamplePackagesUseSharedContracts(t *testing.T) {
	files, e := filepath.Glob("../examples/telemetry/*.yaml")
	if e != nil || len(files) != 6 {
		t.Fatal("require six example packages", e)
	}
	for _, name := range files {
		data, e := os.ReadFile(name)
		if e != nil {
			t.Fatal(e)
		}
		p, e := ReadTelemetryPackage(string(data))
		if e != nil {
			t.Fatal(name, e)
		}
		q := p.Question(0, 300000)
		for _, r := range Recipes() {
			if r.Signal == q.Signal {
				plan, e := PlanRecipe(r.ID, q)
				if e != nil || !reflect.DeepEqual(plan.Question.Where, q.Where) {
					t.Fatal(name, e)
				}
			}
		}
	}
}
