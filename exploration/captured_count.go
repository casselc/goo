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
	if len(q.GroupBy) != 1 || q.Numeric != nil || q.NumericHistogram != nil || q.ReportedEdge != nil || q.SpanStatus || q.GroupTimeSeries || q.GroupBy[0].Kind == "column" && (q.GroupBy[0].Key == "body_pattern_v1" || q.GroupBy[0].Key == "duration_bucket") {
		return errors.New("captured-count comparison requires an ordinary single-field record query without numeric, status, split-time or reported-edge profiles")
	}
	return nil
}

// CompatibleCapturedCountQueries disregards only fixed windows and documented
// delivery paging. Environment equality is typed and checked before removing its
// representation; predicates, grouping and every measure directive remain exact.
func CompatibleCapturedCountQueries(a, b Query) error {
	if err := ValidateCapturedCountQuery(a); err != nil {
		return err
	}
	if err := ValidateCapturedCountQuery(b); err != nil {
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
	if !reflect.DeepEqual(x, y) {
		return errors.New("captured-count comparison requires the same root query definition apart from windows and inspection paging")
	}
	return nil
}
