package exploration

import (
	"errors"
	"testing"
)

func TestPredicateBudgetIdentity(t *testing.T) {
	p := &Predicate{Field: &Field{Kind: "column", Key: "body"}, Op: "eq", Value: "x"}
	for range 6 {
		p = &Predicate{All: []Predicate{*p}}
	}
	if err := ValidatePredicate(p, "logs"); !errors.Is(err, ErrPredicateBudget) {
		t.Fatalf("expected budget identity, got %v", err)
	}
	if _, err := Refine(Query{}, nil); errors.Is(err, ErrPredicateBudget) {
		t.Fatalf("unrelated invalid query reported budget: %v", err)
	}
}
