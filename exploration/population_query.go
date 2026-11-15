package exploration

import "fmt"

// PopulationQuery removes presentation and aggregate projections from an admitted
// record question. It retains the signal, environment representation, fixed bounds
// and complete population restrictions, including typed relationships. Delivery
// paging is explicit. The result is detached and never authorizes a source read.
//
// This is not a widening operation: changing signal, replacing population filters
// or selecting another reference requires a separate, explicit planner.
func PopulationQuery(parent Query, limit, offset int) (Query, error) {
	out, err := CloneQuery(parent)
	if err != nil {
		return Query{}, fmt.Errorf("invalid parent population: %w", err)
	}
	out.GroupBy = nil
	out.Numeric = nil
	out.NumericHistogram = nil
	out.Summary = ""
	out.SpanStatus = false
	out.LogSeverity = false
	out.BucketMS = 0
	out.GroupTimeSeries = false
	out.Limit, out.Offset = limit, offset
	if err := out.Validate(); err != nil {
		return Query{}, fmt.Errorf("invalid population projection: %w", err)
	}
	return out, nil
}
