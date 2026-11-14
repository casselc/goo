package exploration

import (
	"errors"
	"reflect"
)

// ValidateCapturedCountQuery defines the bounded ordinary count-comparison
// profile. Count is source records, never a rate, unique requests or health.
func ValidateCapturedCountQuery(q Query) error {
	if err := q.Validate(); err != nil {
		return err
	}
	if len(q.GroupBy) != 1 || q.Numeric != nil || q.NumericHistogram != nil || q.ReportedEdge != nil || q.SpanStatus || q.LogSeverity || q.GroupTimeSeries || q.GroupBy[0].Kind == "column" && (q.GroupBy[0].Key == "body_pattern_v1" || q.GroupBy[0].Key == "duration_bucket") {
		return errors.New("captured-count comparison requires an ordinary single-field record query without numeric, status, split-time or reported-edge profiles")
	}
	return nil
}

// ValidateCountCohortQuery extends fresh ordinary comparisons with the typed
// span-status summary. It does not widen stored-capture/template reuse or
// authorize per-group status comparisons.
func ValidateCountCohortQuery(q Query) error {
	if q.LogSeverity {
		return errors.New("numeric log severity requires its own summary projection")
	}
	if err := q.Validate(); err != nil {
		return err
	}
	q.SpanStatus = false
	return ValidateCapturedCountQuery(q)
}

// CompatibleCapturedCountQueries disregards only fixed windows and documented
// delivery paging. Environment equality is typed and checked before removing its
// representation; predicates, grouping and every measure directive remain exact.
func CompatibleCapturedCountQueries(a, b Query) error {
	return compatibleCountQueries(a, b, false)
}

// CompatibleCountCohortQueries aligns ordinary supplied count cohorts with
// independently checked predicates. It never authorizes capture/template reuse.
// Only predicates, windows, delivery paging and equivalent environment wire
// representations may differ; grouping and measure directives remain exact.
func CompatibleCountCohortQueries(a, b Query) error {
	return compatibleCountQueries(a, b, true)
}

func compatibleCountQueries(a, b Query, independentPredicates bool) error {
	validate := ValidateCapturedCountQuery
	if independentPredicates {
		validate = ValidateCountCohortQuery
	}
	if err := validate(a); err != nil {
		return err
	}
	if err := validate(b); err != nil {
		return err
	}
	if !a.SameEnvironment(b) {
		return errors.New("captured-count comparison requires the same checked environment policy and scope")
	}
	x := *(CaptureQuestion{Kind: "explore", Explore: &a}).Normalized().Explore
	y := *(CaptureQuestion{Kind: "explore", Explore: &b}).Normalized().Explore
	x.Environment = ""
	y.Environment = ""
	x.EnvironmentSelector = nil
	y.EnvironmentSelector = nil
	x.FromMS = 0
	x.ToMS = 0
	y.FromMS = 0
	y.ToMS = 0
	if independentPredicates {
		x.Where = nil
		y.Where = nil
	}
	if !reflect.DeepEqual(x, y) {
		if independentPredicates {
			return errors.New("count cohorts require the same query profile apart from predicates, windows and inspection paging")
		}
		return errors.New("captured-count comparison requires the same root query definition apart from windows and inspection paging")
	}
	return nil
}
