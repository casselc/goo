package exploration

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type TimePoint struct {
	Numeric    *NumericSummary    `json:"numeric,omitempty"`
	SpanStatus *SpanStatusSummary `json:"spanStatus,omitempty"`
	FromMS     int64              `json:"fromMs"`
	ToMS       int64              `json:"toMs"`
	Count      string             `json:"count"`
	State      string             `json:"state"`
	Duration   *DurationSummary   `json:"duration,omitempty"`
}

// Empty intervals mean no captured records in this aggregate read, not health.
func DecodeTimeSeries(q Query, raw []json.RawMessage) ([]TimePoint, error) {
	if e := q.Validate(); e != nil {
		return nil, e
	}
	if q.BucketMS == 0 {
		return nil, fmt.Errorf("time bucket required")
	}
	if len(raw) > 240 {
		return nil, fmt.Errorf("time series exceeds interval budget")
	}
	points := []TimePoint{}
	indices := map[int64]int{}
	for start := q.FromMS; start < q.ToMS; start += q.BucketMS {
		end := start + q.BucketMS
		if end > q.ToMS {
			end = q.ToMS
		}
		indices[start] = len(points)
		p := TimePoint{FromMS: start, ToMS: end, Count: "0", State: "no_data"}
		if q.Summary != "" {
			d := ReferenceDurationSummary(nil)
			p.Duration = &d
		}
		if q.SpanStatus {
			d := emptySpanStatus()
			p.SpanStatus = &d
		}
		if q.Numeric != nil {
			d, _ := EvaluateNumericValues(*q.Numeric, nil)
			p.Numeric = &d
		}
		points = append(points, p)
	}
	seen := map[int64]bool{}
	for _, r := range raw {
		var v struct {
			From  string `json:"from_ms"`
			Count string `json:"count"`
		}
		if json.Unmarshal(r, &v) != nil {
			return nil, fmt.Errorf("invalid time series row")
		}
		start, e := strconv.ParseInt(v.From, 10, 64)
		if e != nil || strconv.FormatInt(start, 10) != v.From {
			return nil, fmt.Errorf("invalid interval start")
		}
		i, ok := indices[start]
		if !ok || seen[start] {
			return nil, fmt.Errorf("unexpected or duplicate interval")
		}
		seen[start] = true
		n, e := durationDecimal(v.Count, 64)
		if e != nil || n.Sign() == 0 {
			return nil, fmt.Errorf("invalid observed interval count")
		}
		if q.SpanStatus {
			d, e := DecodeSpanStatus(r)
			if e != nil {
				return nil, e
			}
			points[i].SpanStatus = d
		}
		if q.Numeric != nil {
			d, e := decodeNumeric(r, *q.Numeric, v.Count)
			if e != nil {
				return nil, e
			}
			points[i].Numeric = d
		}
		points[i].Count = v.Count
		points[i].State = "observed"
		if q.Summary != "" {
			d, e := DecodeDurationSummary(r)
			if e != nil {
				return nil, e
			}
			points[i].Duration = d
		}
	}
	return points, nil
}

func ValidateTimeSeries(q Query, points []TimePoint) error {
	if q.BucketMS == 0 {
		if len(points) != 0 {
			return fmt.Errorf("unexpected time series")
		}
		return nil
	}
	if e := q.Validate(); e != nil {
		return e
	}
	expected := int((q.ToMS - q.FromMS + q.BucketMS - 1) / q.BucketMS)
	if len(points) != expected {
		return fmt.Errorf("incomplete time intervals")
	}
	for i, p := range points {
		start := q.FromMS + int64(i)*q.BucketMS
		end := min(start+q.BucketMS, q.ToMS)
		if p.FromMS != start || p.ToMS != end {
			return fmt.Errorf("invalid interval bounds")
		}
		n, e := durationDecimal(p.Count, 64)
		if e != nil {
			return e
		}
		state := "observed"
		if n.Sign() == 0 {
			state = "no_data"
		}
		if q.SpanStatus {
			if p.SpanStatus == nil || p.SpanStatus.Count != p.Count || p.SpanStatus.Validate() != nil {
				return fmt.Errorf("invalid interval span status")
			}
		} else if p.SpanStatus != nil {
			return fmt.Errorf("unexpected interval span status")
		}
		if err := validateOptionalNumeric(q.Numeric, p.Numeric, p.Count); err != nil {
			return err
		}
		if p.State != state {
			return fmt.Errorf("invalid interval state")
		}
		if q.Summary != "" {
			if p.Duration == nil || p.Duration.Count != p.Count || p.Duration.Validate() != nil {
				return fmt.Errorf("invalid interval durations")
			}
		} else if p.Duration != nil {
			return fmt.Errorf("unexpected interval durations")
		}
	}
	return nil
}
