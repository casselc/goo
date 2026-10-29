package exploration

import (
	"fmt"
	"strconv"
)

// SurroundingQuestion deliberately replaces identity/message/package predicates
// with time proximity on the anchor's service. It never expands the original window.
func SurroundingQuestion(origin Query, anchorNS, service string, radiusMS int64, offset int) (Query, error) {
	if err := origin.Validate(); err != nil {
		return Query{}, err
	}
	if origin.Signal != "logs" || radiusMS < 1000 || radiusMS > 300000 {
		return Query{}, fmt.Errorf("log origin and radius 1..300 seconds required")
	}
	ns, err := strconv.ParseUint(anchorNS, 10, 64)
	if err != nil {
		return Query{}, fmt.Errorf("invalid anchor timestamp")
	}
	center := int64(ns / 1_000_000)
	if center < origin.FromMS || center >= origin.ToMS {
		return Query{}, fmt.Errorf("anchor outside original window")
	}
	from, to := max(origin.FromMS, center-radiusMS), min(origin.ToMS, center+radiusMS+1)
	f := Field{Kind: "column", Key: "service"}
	q := Query{Signal: "logs", Environment: origin.Environment, FromMS: from, ToMS: to, Where: &Predicate{Field: &f, Op: "eq", Value: service}, Limit: 100, Offset: offset}
	q = q.WithEnvironmentFrom(origin)
	return q, q.Validate()
}
