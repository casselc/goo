package exploration

import (
	"encoding/json"
	"testing"
)

func TestPopulationPreservesPortableFiltersAndWithholdsIncompatibleCompanion(t *testing.T) {
	q := basic()
	q.Signal = "spans"
	q.Where = &Predicate{Any: []Predicate{{Field: &Field{Kind: "attribute", Key: "region"}, Op: "missing"}, {Field: &Field{Kind: "attribute", Key: "region"}, Op: "eq", Value: ""}}}
	before, _ := json.Marshal(q)
	p, e := PlanPopulation(PopulationRequest{Question: q})
	if e != nil || len(p.Queries) != 2 {
		t.Fatal(p, e)
	}
	for _, v := range p.Queries {
		a, _ := json.Marshal(v.Where)
		b, _ := json.Marshal(q.Where)
		if string(a) != string(b) || v.FromMS != q.FromMS || v.ToMS != q.ToMS || v.Environment != q.Environment || v.BucketMS == 0 || v.Offset != 0 {
			t.Fatal("scope changed", v)
		}
	}
	after, _ := json.Marshal(q)
	if string(before) != string(after) {
		t.Fatal("mutated input")
	}
	q.Where = &Predicate{Field: &Field{Kind: "column", Key: "status"}, Op: "eq", Value: "Error"}
	p, e = PlanPopulation(PopulationRequest{Question: q})
	if e != nil || len(p.Queries) != 1 || len(p.Notices) < 2 {
		t.Fatal("span-only filter silently dropped", p, e)
	}
	other := basic()
	other.Signal = "logs"
	other.Where = nil
	p, e = PlanPopulation(PopulationRequest{Question: q, Companion: &other})
	if e != nil || len(p.Queries) != 2 || p.Queries[1].Where != nil {
		t.Fatal(p, e)
	}
	other.Environment = "foreign"
	if _, e = PlanPopulation(PopulationRequest{Question: q, Companion: &other}); e == nil {
		t.Fatal("foreign companion")
	}
}
