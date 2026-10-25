package exploration

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"unicode/utf8"
)

const RecordAggregatesVersion = "goo/record-aggregates/v1"

type RecordAggregatesReads struct {
	Count      CountedValuesRead  `json:"count"`
	Groups     CountedValuesRead  `json:"groups"`
	TimeSeries *CountedValuesRead `json:"timeSeries,omitempty"`
}

// Changed reads retain diagnostics, never an actionable conserved snapshot.
type RecordAggregatesResult struct {
	Version    string                `json:"version"`
	Question   Query                 `json:"question"`
	CapturedAt string                `json:"capturedAt"`
	Basis      string                `json:"basis"`
	Count      string                `json:"count"`
	State      string                `json:"state"`
	Groups     []json.RawMessage     `json:"groups"`
	Duration   *DurationSummary      `json:"duration,omitempty"`
	SpanStatus *SpanStatusSummary    `json:"spanStatus,omitempty"`
	TimeSeries []TimePoint           `json:"timeSeries,omitempty"`
	Warnings   []string              `json:"warnings"`
	Reads      RecordAggregatesReads `json:"reads"`
}

func ValidateRecordAggregatesQuery(q Query) error {
	if e := q.Validate(); e != nil {
		return e
	}
	if len(q.GroupBy) != 1 || q.Numeric != nil || q.NumericHistogram != nil || q.GroupTimeSeries || q.Offset != 0 || q.Summary != "" && q.Summary != "span-duration" {
		return fmt.Errorf("record aggregates requires one string group, offset zero and supported duration/status/time projections only")
	}
	typ, e := q.GroupBy[0].Type(q.Signal)
	if e != nil {
		return e
	}
	if typ != "string" {
		return fmt.Errorf("record aggregates requires a string group")
	}
	return nil
}

// FetchRecordAggregates executes only full-population aggregate reads. It never
// inspects records or discovers fields, and cannot substitute for saved evidence.
func (s Source) FetchRecordAggregates(ctx context.Context, q Query) (*RecordAggregatesResult, error) {
	if e := ValidateRecordAggregatesQuery(q); e != nil {
		return nil, e
	}
	encoded, e := json.Marshal(q)
	if e != nil {
		return nil, e
	}
	var question Query
	if e = json.Unmarshal(encoded, &question); e != nil {
		return nil, e
	}
	plan, e := Compile(question)
	if e != nil {
		return nil, e
	}
	read := func(sql string) ([]json.RawMessage, CountedValuesRead, error) {
		if e := ctx.Err(); e != nil {
			return nil, CountedValuesRead{}, e
		}
		raw, e := s.read(ctx, sql, plan.Parameters)
		if e != nil {
			return nil, CountedValuesRead{}, e
		}
		return raw, CountedValuesRead{ReadAt: time.Now().UTC().Format(time.RFC3339Nano), Basis: countedValuesHash(raw)}, nil
	}
	raw, countRead, e := read(plan.Count)
	if e != nil {
		return nil, e
	}
	var total struct {
		Count *string `json:"count"`
	}
	if len(raw) != 1 || json.Unmarshal(raw[0], &total) != nil || total.Count == nil {
		return nil, fmt.Errorf("invalid aggregate count result")
	}
	count, e := countedValuesUint(*total.Count, false)
	if e != nil {
		return nil, e
	}
	r := &RecordAggregatesResult{Version: RecordAggregatesVersion, Question: question, Count: *total.Count, State: "observed", Groups: []json.RawMessage{}, Warnings: []string{"Captured records only; sampling, ingestion completeness and clock accuracy are unverified.", "Count, groups and time series are independent aggregate reads, not a transactional snapshot. No records or field discovery were read."}, Reads: RecordAggregatesReads{Count: countRead}}
	if count == 0 {
		r.State = "no_data"
	}
	if q.Summary != "" {
		r.Duration, e = DecodeDurationSummary(raw[0])
		if e != nil {
			return nil, e
		}
	}
	if q.SpanStatus {
		r.SpanStatus, e = DecodeSpanStatus(raw[0])
		if e != nil {
			return nil, e
		}
	}
	groups, groupRead, e := read(plan.Groups)
	if e != nil {
		return nil, e
	}
	r.Reads.Groups = groupRead
	if len(groups) > 1000 {
		return nil, fmt.Errorf("more than 1000 groups; narrow query; no partial grouping totals")
	}
	for _, raw := range groups {
		var g map[string]json.RawMessage
		if !utf8.Valid(raw) || json.Unmarshal(raw, &g) != nil {
			return nil, fmt.Errorf("invalid aggregate group")
		}
		var n string
		if json.Unmarshal(g["count"], &n) != nil {
			return nil, fmt.Errorf("invalid aggregate group count")
		}
		if _, e = countedValuesUint(n, true); e != nil {
			return nil, e
		}
	}
	var changed bool
	r.Groups, changed, e = decodeAggregateGroups(question, groups, r.Count, r.Duration, r.SpanStatus)
	if e != nil {
		return nil, e
	}
	if changed {
		r.State = "changed_between_reads"
	}
	if q.BucketMS != 0 {
		series, receipt, e := read(plan.TimeSeries)
		if e != nil {
			return nil, e
		}
		r.Reads.TimeSeries = &receipt
		r.TimeSeries, e = DecodeTimeSeries(question, series)
		if e != nil {
			return nil, e
		}
		if !aggregateSeriesAgree(r.Count, r.SpanStatus, r.Duration, r.TimeSeries) {
			r.State = "changed_between_reads"
		}
	}
	if r.State == "changed_between_reads" {
		r.Warnings = append(r.Warnings, "Independent aggregate reads disagree; retained values are diagnostics, not an actionable population snapshot.")
	}
	r.CapturedAt = time.Now().UTC().Format(time.RFC3339Nano)
	basis := *r
	basis.CapturedAt = ""
	basis.Reads.Count.ReadAt = ""
	basis.Reads.Groups.ReadAt = ""
	if basis.Reads.TimeSeries != nil {
		receipt := *basis.Reads.TimeSeries
		receipt.ReadAt = ""
		basis.Reads.TimeSeries = &receipt
	}
	r.Basis = countedValuesHash(basis)
	return r, nil
}
