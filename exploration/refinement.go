package exploration

import (
	"fmt"
	"slices"
)

// ValidatePredicate admits an ordinary predicate independently of a question.
// A nil predicate means no restriction. Traversal stops at the ordinary depth
// and node limits, including for cyclic input.
func ValidatePredicate(p *Predicate, signal string) error {
	if signal != "logs" && signal != "spans" {
		return fmt.Errorf("signal must be spans or logs")
	}
	count := 0
	return validatePredicate(p, signal, 0, &count)
}

// ClonePredicate returns a detached copy after bounded admission.
func ClonePredicate(p *Predicate, signal string) (*Predicate, error) {
	if err := ValidatePredicate(p, signal); err != nil {
		return nil, err
	}
	return clonePredicate(p), nil
}

// clonePredicate only accepts predicates already admitted by ValidatePredicate.
func clonePredicate(p *Predicate) *Predicate {
	if p == nil {
		return nil
	}
	out := *p
	if p.Field != nil {
		field := *p.Field
		out.Field = &field
	}
	if p.Numeric != nil {
		numeric := *p.Numeric
		out.Numeric = &numeric
	}
	if p.All != nil {
		out.All = make([]Predicate, len(p.All))
		for i := range p.All {
			out.All[i] = *clonePredicate(&p.All[i])
		}
	}
	if p.Any != nil {
		out.Any = make([]Predicate, len(p.Any))
		for i := range p.Any {
			out.Any[i] = *clonePredicate(&p.Any[i])
		}
	}
	return &out
}

// CloneQuery returns a fully detached copy of an admitted question.
func CloneQuery(q Query) (Query, error) {
	if err := q.Validate(); err != nil {
		return Query{}, err
	}
	return cloneQuery(q), nil
}

func cloneQuery(q Query) Query {
	out := q
	out.Where = clonePredicate(q.Where)
	out.GroupBy = slices.Clone(q.GroupBy)
	if q.EnvironmentSelector != nil {
		selector := *q.EnvironmentSelector
		out.EnvironmentSelector = &selector
	}
	if q.Numeric != nil {
		numeric := *q.Numeric
		out.Numeric = &numeric
	}
	if q.NumericHistogram != nil {
		histogram := *q.NumericHistogram
		out.NumericHistogram = &histogram
	}
	if q.ReportedEdge != nil {
		edge := *q.ReportedEdge
		if edge.Parent.Namespace != nil {
			ns := *edge.Parent.Namespace
			edge.Parent.Namespace = &ns
		}
		if edge.Child.Namespace != nil {
			ns := *edge.Child.Namespace
			edge.Child.Namespace = &ns
		}
		if edge.DiscoveryNamespace != nil {
			ns := *edge.DiscoveryNamespace
			edge.DiscoveryNamespace = &ns
		}
		out.ReportedEdge = &edge
	}
	return out
}

// Refine adds a predicate under AND with the complete parent restriction.
// It preserves environment, window, measurements, grouping and paging. Callers
// own any policy such as resetting the page. The combined budget is admitted
// before cloning either input; failures return no partial question.
func Refine(parent Query, predicate *Predicate) (Query, error) {
	if err := parent.Validate(); err != nil {
		return Query{}, fmt.Errorf("invalid parent question: %w", err)
	}
	if err := ValidatePredicate(predicate, parent.Signal); err != nil {
		return Query{}, fmt.Errorf("invalid refinement predicate: %w", err)
	}
	combined := predicate
	if parent.Where != nil {
		combined = parent.Where
		if predicate != nil {
			combined = &Predicate{All: []Predicate{*parent.Where, *predicate}}
		}
	}
	if err := ValidatePredicate(combined, parent.Signal); err != nil {
		return Query{}, fmt.Errorf("combined predicate: %w", err)
	}
	out := parent
	out.Where = combined
	return cloneQuery(out), nil
}
