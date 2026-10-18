package exploration

import (
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
)

const MaxTimeSeriesGroups = 8
const MaxTimeSeriesCells = 1000

type TimeSeriesGroup struct {
	Values []*string   `json:"values"`
	Points []TimePoint `json:"points"`
}

func groupKey(values []*string) string { b, _ := json.Marshal(values); return string(b) }
func groupValues(q Query, raw json.RawMessage) ([]*string, error) {
	var row map[string]json.RawMessage
	if json.Unmarshal(raw, &row) != nil {
		return nil, fmt.Errorf("invalid series group")
	}
	values := []*string{}
	for i := range q.GroupBy {
		b, ok := row[fmt.Sprintf("g%d", i)]
		if !ok {
			return nil, fmt.Errorf("missing series dimension")
		}
		var v *string
		if json.Unmarshal(b, &v) != nil {
			return nil, fmt.Errorf("invalid series dimension")
		}
		values = append(values, v)
	}
	return values, nil
}
func DecodeGroupedTimeSeries(q Query, raw []json.RawMessage) ([]TimeSeriesGroup, error) {
	if e := q.Validate(); e != nil {
		return nil, e
	}
	if !q.GroupTimeSeries {
		return nil, fmt.Errorf("split chart not requested")
	}
	if len(raw) > MaxTimeSeriesCells {
		return nil, fmt.Errorf("split chart exceeds 1000 observed cells; narrow scope or increase interval")
	}
	rows := map[string][]json.RawMessage{}
	values := map[string][]*string{}
	for _, r := range raw {
		v, e := groupValues(q, r)
		if e != nil {
			return nil, e
		}
		key := groupKey(v)
		values[key] = v
		rows[key] = append(rows[key], r)
		if len(rows) > MaxTimeSeriesGroups {
			return nil, fmt.Errorf("split chart exceeds 8 series; narrow scope or choose another breakdown")
		}
	}
	keys := []string{}
	for key := range rows {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := []TimeSeriesGroup{}
	for _, key := range keys {
		points, e := DecodeTimeSeries(q, rows[key])
		if e != nil {
			return nil, e
		}
		out = append(out, TimeSeriesGroup{Values: values[key], Points: points})
	}
	return out, ValidateGroupedTimeSeries(q, out)
}
func ValidateGroupedTimeSeries(q Query, groups []TimeSeriesGroup) error {
	if !q.GroupTimeSeries {
		if len(groups) != 0 {
			return fmt.Errorf("unexpected split series")
		}
		return nil
	}
	if e := q.Validate(); e != nil {
		return e
	}
	if len(groups) > MaxTimeSeriesGroups {
		return fmt.Errorf("too many split series")
	}
	previous := ""
	observed := 0
	for _, g := range groups {
		if len(g.Values) != len(q.GroupBy) {
			return fmt.Errorf("series dimension count mismatch")
		}
		key := groupKey(g.Values)
		if previous != "" && key <= previous {
			return fmt.Errorf("duplicate or unordered series")
		}
		previous = key
		if e := ValidateTimeSeries(q, g.Points); e != nil {
			return e
		}
		nonempty := false
		for _, p := range g.Points {
			if p.Count != "0" {
				observed++
				nonempty = true
			}
		}
		if !nonempty {
			return fmt.Errorf("invented unobserved series")
		}
	}
	if observed > MaxTimeSeriesCells {
		return fmt.Errorf("too many observed series cells")
	}
	return nil
}

// Compare independent reads using additive states, never scalar percentiles.
// Agreement cannot establish a transaction, completeness or identical raw rows.
func groupedSeriesAgree(r *Result) bool {
	q := r.Question
	if !q.GroupTimeSeries {
		return true
	}
	totals := make([]big.Int, len(r.TimeSeries))
	sums := make([]big.Int, len(r.TimeSeries))
	statuses := make([]SpanStatusSummary, len(r.TimeSeries))
	for i := range statuses {
		statuses[i] = emptySpanStatus()
	}
	type total struct {
		count, sum string
		status     *SpanStatusSummary
	}
	expected := map[string]total{}
	for _, raw := range r.Groups {
		values, e := groupValues(q, raw)
		if e != nil {
			return false
		}
		var v struct {
			Count    string             `json:"count"`
			Duration *DurationSummary   `json:"duration"`
			Status   *SpanStatusSummary `json:"spanStatus"`
		}
		if json.Unmarshal(raw, &v) != nil {
			return false
		}
		sum := ""
		if v.Duration != nil {
			sum = v.Duration.SumNS
		}
		expected[groupKey(values)] = total{v.Count, sum, v.Status}
	}
	if len(expected) != len(r.GroupedTimeSeries) {
		return false
	}
	for _, g := range r.GroupedTimeSeries {
		count, sum := new(big.Int), new(big.Int)
		status := emptySpanStatus()
		for i, p := range g.Points {
			n, _ := new(big.Int).SetString(p.Count, 10)
			totals[i].Add(&totals[i], n)
			count.Add(count, n)
			if p.Duration != nil {
				n, _ = new(big.Int).SetString(p.Duration.SumNS, 10)
				sums[i].Add(&sums[i], n)
				sum.Add(sum, n)
			}
			if p.SpanStatus != nil {
				statuses[i].add(*p.SpanStatus)
				status.add(*p.SpanStatus)
			}
		}
		want, ok := expected[groupKey(g.Values)]
		if !ok || want.count != count.String() {
			return false
		}
		if q.Summary != "" && want.sum != sum.String() {
			return false
		}
		if q.SpanStatus && (want.status == nil || *want.status != status) {
			return false
		}
	}
	for i, p := range r.TimeSeries {
		if totals[i].String() != p.Count {
			return false
		}
		if p.Duration != nil && sums[i].String() != p.Duration.SumNS {
			return false
		}
		if p.SpanStatus != nil && statuses[i] != *p.SpanStatus {
			return false
		}
	}
	return true
}
